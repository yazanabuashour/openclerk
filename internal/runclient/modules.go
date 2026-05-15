package runclient

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
	_ "modernc.org/sqlite"
)

const (
	SemanticModuleProviderOllama = "ollama"
	SemanticModuleProviderGemini = "gemini"
	OCRModuleProviderTesseract   = "tesseract"

	ModuleKindEmbeddingProvider = "embedding_provider"
	ModuleKindOCRProvider       = "ocr_provider"

	semanticModuleCommand       = "semantic-retrieval-adapter"
	semanticModuleSearchCommand = "semantic-retrieval-adapter search"
	canonicalGeminiAPIBase      = "https://generativelanguage.googleapis.com/v1beta"
)

type SemanticModuleConfig struct {
	Kind                 string            `json:"kind,omitempty"`
	Provider             string            `json:"provider"`
	ModuleName           string            `json:"module_name"`
	Enabled              bool              `json:"enabled"`
	Command              string            `json:"command,omitempty"`
	CommandArgs          []string          `json:"command_args,omitempty"`
	ManifestPath         string            `json:"manifest_path,omitempty"`
	ManifestSHA256       string            `json:"manifest_sha256,omitempty"`
	ManifestResolvedPath string            `json:"-"`
	ProviderConfig       map[string]string `json:"provider_config,omitempty"`
	VerificationStatus   string            `json:"verification_status"`
	RedactionStatus      string            `json:"redaction_status"`
}

type SemanticModuleInstallInput struct {
	Kind           string            `json:"kind,omitempty"`
	Provider       string            `json:"provider"`
	ModuleName     string            `json:"module_name,omitempty"`
	ManifestPath   string            `json:"manifest_path"`
	ManifestRoot   string            `json:"manifest_root,omitempty"`
	Command        string            `json:"command"`
	CommandArgs    []string          `json:"command_args,omitempty"`
	ProviderConfig map[string]string `json:"provider_config,omitempty"`
	Enabled        *bool             `json:"enabled,omitempty"`
}

type SemanticModuleConfigureInput struct {
	Kind           string            `json:"kind,omitempty"`
	Provider       string            `json:"provider"`
	ProviderConfig map[string]string `json:"provider_config,omitempty"`
	Enabled        *bool             `json:"enabled,omitempty"`
}

func InstallSemanticModule(ctx context.Context, cfg Config, input SemanticModuleInstallInput) (SemanticModuleConfig, error) {
	if moduleInstallKind(input.Kind, input.Provider) == ModuleKindOCRProvider {
		return InstallOCRModule(ctx, cfg, input)
	}
	provider, err := normalizeSemanticModuleProvider(input.Provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	if strings.TrimSpace(input.ManifestPath) == "" {
		return SemanticModuleConfig{}, domain.ValidationError("module.manifest_path is required", nil)
	}
	if command := strings.TrimSpace(input.Command); command != "" && command != semanticModuleCommand {
		return SemanticModuleConfig{}, domain.ValidationError("module.command must be semantic-retrieval-adapter for semantic modules", nil)
	}
	if len(input.CommandArgs) != 0 {
		return SemanticModuleConfig{}, domain.ValidationError("module.command_args are unsupported for semantic modules", nil)
	}
	manifestPath := filepath.Clean(input.ManifestPath)
	resolvedManifestPath, err := resolveModuleManifestPathForInstall(cfg, input.ManifestRoot, manifestPath, semanticModuleManifestPolicy)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	manifest, manifestSHA, err := verifySemanticModuleManifest(resolvedManifestPath, provider, input.ModuleName)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	providerConfig := redactedProviderConfig(provider, input.ProviderConfig)
	if err := validateSemanticProviderConfig(provider, providerConfig); err != nil {
		return SemanticModuleConfig{}, err
	}
	config := SemanticModuleConfig{
		Kind:                 ModuleKindEmbeddingProvider,
		Provider:             provider,
		ModuleName:           manifest.Module.Name,
		Enabled:              enabled,
		Command:              semanticModuleCommand,
		CommandArgs:          nil,
		ManifestPath:         manifestPath,
		ManifestSHA256:       manifestSHA,
		ManifestResolvedPath: resolvedManifestPath,
		ProviderConfig:       providerConfig,
		VerificationStatus:   "verified",
		RedactionStatus:      "redacted",
	}
	if err := writeSemanticModuleConfig(ctx, cfg, config); err != nil {
		return SemanticModuleConfig{}, err
	}
	return config, nil
}

func ConfigureSemanticModule(ctx context.Context, cfg Config, input SemanticModuleConfigureInput) (SemanticModuleConfig, error) {
	if moduleInstallKind(input.Kind, input.Provider) == ModuleKindOCRProvider {
		return ConfigureOCRModule(ctx, cfg, input)
	}
	provider, err := normalizeSemanticModuleProvider(input.Provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	current, err := ReadSemanticModuleConfig(ctx, cfg, provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	if strings.TrimSpace(current.ModuleName) == "" {
		return SemanticModuleConfig{}, domain.ValidationError("semantic module is not installed", map[string]any{"provider": provider})
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	merged := map[string]string{}
	for key, value := range current.ProviderConfig {
		merged[key] = value
	}
	providerConfigUpdates := redactedProviderConfig(provider, input.ProviderConfig)
	for key, value := range providerConfigUpdates {
		merged[key] = value
	}
	if current.Enabled || len(providerConfigUpdates) > 0 {
		if err := validateSemanticProviderConfig(provider, merged); err != nil {
			return SemanticModuleConfig{}, err
		}
	}
	current.ProviderConfig = merged
	current.VerificationStatus = "verified"
	current.RedactionStatus = "redacted"
	if err := writeSemanticModuleConfig(ctx, cfg, current); err != nil {
		return SemanticModuleConfig{}, err
	}
	return current, nil
}

func RemoveSemanticModule(ctx context.Context, cfg Config, provider string) (SemanticModuleConfig, error) {
	if strings.EqualFold(strings.TrimSpace(provider), OCRModuleProviderTesseract) {
		return RemoveOCRModule(ctx, cfg, provider)
	}
	normalized, err := normalizeSemanticModuleProvider(provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	current, err := ReadSemanticModuleConfig(ctx, cfg, normalized)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	if strings.TrimSpace(current.ModuleName) == "" {
		current.Provider = normalized
	}
	if err := deleteSemanticModuleConfig(ctx, cfg, normalized); err != nil {
		return SemanticModuleConfig{}, err
	}
	current.Enabled = false
	current.VerificationStatus = "removed"
	current.RedactionStatus = "redacted"
	return current, nil
}

func ListSemanticModules(ctx context.Context, cfg Config) ([]SemanticModuleConfig, error) {
	modules := []SemanticModuleConfig{}
	for _, provider := range []string{SemanticModuleProviderOllama, SemanticModuleProviderGemini} {
		config, err := ReadSemanticModuleConfig(ctx, cfg, provider)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(config.ModuleName) != "" {
			modules = append(modules, config)
		}
	}
	ocrConfig, err := ReadOCRModuleConfig(ctx, cfg, OCRModuleProviderTesseract)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(ocrConfig.ModuleName) != "" {
		modules = append(modules, ocrConfig)
	}
	return modules, nil
}

// ListConfiguredModules returns stored module summaries without revalidating helper commands.
func ListConfiguredModules(ctx context.Context, cfg Config) ([]SemanticModuleConfig, error) {
	modules := []SemanticModuleConfig{}
	for _, provider := range []string{SemanticModuleProviderOllama, SemanticModuleProviderGemini} {
		values, err := readSemanticModuleValues(ctx, cfg, provider)
		if err != nil {
			return nil, err
		}
		config := semanticModuleConfigFromValues(ModuleKindEmbeddingProvider, provider, values)
		if strings.TrimSpace(config.ModuleName) != "" {
			modules = append(modules, config)
		}
	}
	values, err := readOCRModuleValues(ctx, cfg, OCRModuleProviderTesseract)
	if err != nil {
		return nil, err
	}
	config := semanticModuleConfigFromValues(ModuleKindOCRProvider, OCRModuleProviderTesseract, values)
	if strings.TrimSpace(config.ModuleName) != "" {
		modules = append(modules, config)
	}
	return modules, nil
}

func ReadSemanticModuleConfig(ctx context.Context, cfg Config, provider string) (SemanticModuleConfig, error) {
	normalized, err := normalizeSemanticModuleProvider(provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	values, err := readSemanticModuleValues(ctx, cfg, normalized)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	config := semanticModuleConfigFromValues(ModuleKindEmbeddingProvider, normalized, values)
	if strings.TrimSpace(config.ModuleName) == "" {
		return config, nil
	}
	if !config.Enabled {
		config.VerificationStatus = "disabled"
		return config, nil
	}
	if err := verifyInstalledSemanticModule(cfg, config); err != nil {
		config.VerificationStatus = "verification_failed"
		return config, err
	}
	config.VerificationStatus = "verified"
	return config, nil
}

func InstallOCRModule(ctx context.Context, cfg Config, input SemanticModuleInstallInput) (SemanticModuleConfig, error) {
	provider, err := normalizeOCRModuleProvider(input.Provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	if strings.TrimSpace(input.ManifestPath) == "" {
		return SemanticModuleConfig{}, domain.ValidationError("module.manifest_path is required", nil)
	}
	if strings.TrimSpace(input.Command) == "" {
		return SemanticModuleConfig{}, domain.ValidationError("module.command is required", nil)
	}
	manifestPath := filepath.Clean(input.ManifestPath)
	resolvedManifestPath, err := resolveModuleManifestPathForInstall(cfg, input.ManifestRoot, manifestPath, ocrModuleManifestPolicy)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	manifest, manifestSHA, err := verifyOCRModuleManifest(resolvedManifestPath, provider, input.ModuleName)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	config := SemanticModuleConfig{
		Kind:                 ModuleKindOCRProvider,
		Provider:             provider,
		ModuleName:           manifest.Module.Name,
		Enabled:              enabled,
		Command:              strings.TrimSpace(input.Command),
		CommandArgs:          sanitizedArgs(input.CommandArgs),
		ManifestPath:         manifestPath,
		ManifestSHA256:       manifestSHA,
		ManifestResolvedPath: resolvedManifestPath,
		ProviderConfig:       redactedProviderConfig(provider, input.ProviderConfig),
		VerificationStatus:   "verified",
		RedactionStatus:      "redacted",
	}
	if err := writeOCRModuleConfig(ctx, cfg, config); err != nil {
		return SemanticModuleConfig{}, err
	}
	return config, nil
}

func ConfigureOCRModule(ctx context.Context, cfg Config, input SemanticModuleConfigureInput) (SemanticModuleConfig, error) {
	provider, err := normalizeOCRModuleProvider(input.Provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	current, err := ReadOCRModuleConfig(ctx, cfg, provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	if strings.TrimSpace(current.ModuleName) == "" {
		return SemanticModuleConfig{}, domain.ValidationError("OCR module is not installed", map[string]any{"provider": provider})
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	merged := map[string]string{}
	for key, value := range current.ProviderConfig {
		merged[key] = value
	}
	for key, value := range redactedProviderConfig(provider, input.ProviderConfig) {
		merged[key] = value
	}
	current.ProviderConfig = merged
	current.VerificationStatus = "verified"
	current.RedactionStatus = "redacted"
	if err := writeOCRModuleConfig(ctx, cfg, current); err != nil {
		return SemanticModuleConfig{}, err
	}
	return current, nil
}

func RemoveOCRModule(ctx context.Context, cfg Config, provider string) (SemanticModuleConfig, error) {
	normalized, err := normalizeOCRModuleProvider(provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	current, err := ReadOCRModuleConfig(ctx, cfg, normalized)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	if strings.TrimSpace(current.ModuleName) == "" {
		current.Provider = normalized
		current.Kind = ModuleKindOCRProvider
	}
	if err := deleteOCRModuleConfig(ctx, cfg, normalized); err != nil {
		return SemanticModuleConfig{}, err
	}
	current.Enabled = false
	current.VerificationStatus = "removed"
	current.RedactionStatus = "redacted"
	return current, nil
}

func ReadOCRModuleConfig(ctx context.Context, cfg Config, provider string) (SemanticModuleConfig, error) {
	normalized, err := normalizeOCRModuleProvider(provider)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	values, err := readOCRModuleValues(ctx, cfg, normalized)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	config := semanticModuleConfigFromValues(ModuleKindOCRProvider, normalized, values)
	if strings.TrimSpace(config.ModuleName) == "" {
		return config, nil
	}
	if !config.Enabled {
		config.VerificationStatus = "disabled"
		return config, nil
	}
	if err := verifyInstalledOCRModule(cfg, config); err != nil {
		config.VerificationStatus = "verification_failed"
		return config, err
	}
	config.VerificationStatus = "verified"
	return config, nil
}

func semanticModuleConfigFromValues(kind string, provider string, values map[string]string) SemanticModuleConfig {
	config := SemanticModuleConfig{
		Kind:                 kind,
		Provider:             provider,
		ModuleName:           values["module_name"],
		Enabled:              values["enabled"] == "true",
		Command:              values["command"],
		ManifestPath:         values["manifest_path"],
		ManifestSHA256:       values["manifest_sha256"],
		ManifestResolvedPath: values["manifest_resolved_path"],
		ProviderConfig:       map[string]string{},
		VerificationStatus:   "not_installed",
		RedactionStatus:      "redacted",
	}
	_ = json.Unmarshal([]byte(values["command_args_json"]), &config.CommandArgs)
	_ = json.Unmarshal([]byte(values["provider_config_json"]), &config.ProviderConfig)
	if config.ProviderConfig == nil {
		config.ProviderConfig = map[string]string{}
	}
	if strings.TrimSpace(config.ModuleName) != "" && config.Enabled {
		config.VerificationStatus = "configured"
	} else if strings.TrimSpace(config.ModuleName) != "" {
		config.VerificationStatus = "disabled"
	}
	return config
}

func normalizeSemanticModuleProvider(provider string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		normalized = SemanticModuleProviderOllama
	}
	switch normalized {
	case SemanticModuleProviderOllama, SemanticModuleProviderGemini:
		return normalized, nil
	default:
		return "", domain.ValidationError("module.provider must be ollama or gemini", nil)
	}
}

func moduleInstallKind(kind string, provider string) string {
	normalizedKind := strings.ToLower(strings.TrimSpace(kind))
	if normalizedKind != "" {
		return normalizedKind
	}
	if strings.EqualFold(strings.TrimSpace(provider), OCRModuleProviderTesseract) {
		return ModuleKindOCRProvider
	}
	return ModuleKindEmbeddingProvider
}

func normalizeOCRModuleProvider(provider string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		normalized = OCRModuleProviderTesseract
	}
	switch normalized {
	case OCRModuleProviderTesseract:
		return normalized, nil
	default:
		return "", domain.ValidationError("OCR module.provider must be tesseract", nil)
	}
}

func verifyInstalledSemanticModule(cfg Config, config SemanticModuleConfig) error {
	if strings.TrimSpace(config.Command) == "" {
		return domain.ValidationError("semantic module command is missing", map[string]any{"provider": config.Provider})
	}
	if config.Command != semanticModuleCommand {
		return domain.ValidationError("semantic module command must be semantic-retrieval-adapter", map[string]any{"provider": config.Provider})
	}
	if len(config.CommandArgs) != 0 {
		return domain.ValidationError("semantic module command_args are unsupported", map[string]any{"provider": config.Provider})
	}
	if strings.TrimSpace(config.ManifestPath) == "" || strings.TrimSpace(config.ManifestSHA256) == "" {
		return domain.ValidationError("semantic module manifest verification is missing", map[string]any{"provider": config.Provider})
	}
	manifest, sha, err := verifyInstalledModuleManifest(cfg, config, semanticModuleManifestPolicy)
	if err != nil {
		return err
	}
	if sha != config.ManifestSHA256 {
		return domain.ValidationError("semantic module manifest digest changed", map[string]any{"provider": config.Provider, "module": manifest.Module.Name})
	}
	return nil
}

func verifyInstalledOCRModule(cfg Config, config SemanticModuleConfig) error {
	if strings.TrimSpace(config.Command) == "" {
		return domain.ValidationError("OCR module command is missing", map[string]any{"provider": config.Provider})
	}
	if strings.TrimSpace(config.ManifestPath) == "" || strings.TrimSpace(config.ManifestSHA256) == "" {
		return domain.ValidationError("OCR module manifest verification is missing", map[string]any{"provider": config.Provider})
	}
	manifest, sha, err := verifyInstalledModuleManifest(cfg, config, ocrModuleManifestPolicy)
	if err != nil {
		return err
	}
	if sha != config.ManifestSHA256 {
		return domain.ValidationError("OCR module manifest digest changed", map[string]any{"provider": config.Provider, "module": manifest.Module.Name})
	}
	return nil
}

type semanticModuleManifest struct {
	SchemaVersion string `json:"schema_version"`
	Module        struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Kind    string `json:"kind"`
	} `json:"module"`
	Provides []struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"provides"`
	Authority struct {
		Default       string   `json:"default"`
		DurableWrites string   `json:"durable_writes"`
		Forbidden     []string `json:"forbidden"`
	} `json:"authority"`
	Release struct {
		Status string `json:"status"`
	} `json:"release"`
}

type moduleManifestPolicy struct {
	label                   string
	kind                    string
	providerMatches         func(moduleName string, provider string) bool
	providesRequiredCommand func(command string) bool
	missingCommandMessage   string
}

var (
	semanticModuleManifestPolicy = moduleManifestPolicy{
		label:                   "semantic module",
		kind:                    ModuleKindEmbeddingProvider,
		providerMatches:         semanticModuleProviderMatches,
		providesRequiredCommand: func(command string) bool { return strings.TrimSpace(command) == semanticModuleSearchCommand },
		missingCommandMessage:   "semantic module manifest must provide a search command",
	}
	ocrModuleManifestPolicy = moduleManifestPolicy{
		label: "OCR module",
		kind:  ModuleKindOCRProvider,
		providerMatches: func(moduleName string, provider string) bool {
			return strings.Contains(moduleName, provider)
		},
		providesRequiredCommand: func(command string) bool {
			return strings.Contains(strings.ToLower(command), "ocr")
		},
		missingCommandMessage: "OCR module manifest must provide an OCR command",
	}
)

func verifySemanticModuleManifest(path string, provider string, expectedName string) (semanticModuleManifest, string, error) {
	return verifyModuleManifest(path, provider, expectedName, semanticModuleManifestPolicy)
}

func verifyOCRModuleManifest(path string, provider string, expectedName string) (semanticModuleManifest, string, error) {
	return verifyModuleManifest(path, provider, expectedName, ocrModuleManifestPolicy)
}

func verifyInstalledModuleManifest(cfg Config, config SemanticModuleConfig, policy moduleManifestPolicy) (semanticModuleManifest, string, error) {
	path, err := resolveInstalledModuleManifestPath(cfg, config, policy)
	if err != nil {
		return semanticModuleManifest{}, "", err
	}
	return verifyModuleManifest(path, config.Provider, config.ModuleName, policy)
}

func resolveModuleManifestPathForInstall(cfg Config, manifestRoot string, manifestPath string, policy moduleManifestPolicy) (string, error) {
	return resolveModuleManifestPath(cfg, manifestPath, "", manifestRoot, policy)
}

func resolveInstalledModuleManifestPath(cfg Config, config SemanticModuleConfig, policy moduleManifestPolicy) (string, error) {
	return resolveModuleManifestPath(cfg, config.ManifestPath, config.ManifestResolvedPath, "", policy)
}

func resolveModuleManifestPath(cfg Config, manifestPath string, resolvedPath string, manifestRoot string, policy moduleManifestPolicy) (string, error) {
	for _, candidate := range moduleManifestPathCandidates(cfg, manifestPath, resolvedPath, manifestRoot) {
		if candidate == "" {
			continue
		}
		clean := filepath.Clean(candidate)
		if !filepath.IsAbs(clean) {
			abs, err := filepath.Abs(clean)
			if err != nil {
				return "", domain.InternalError("resolve "+policy.label+" manifest path", err)
			}
			clean = filepath.Clean(abs)
		}
		info, err := os.Stat(clean)
		if err == nil {
			if info.IsDir() {
				return "", domain.ValidationError(policy.label+" manifest could not be resolved", map[string]any{"manifest_path": manifestPath, "resolved_path": clean, "reason": "path is a directory"})
			}
			return clean, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", domain.InternalError("inspect "+policy.label+" manifest", err)
		}
	}
	return "", domain.ValidationError(policy.label+" manifest could not be resolved", map[string]any{"manifest_path": manifestPath, "manifest_root": strings.TrimSpace(manifestRoot), "guidance": "use module.manifest_root with relative module.manifest_path or provide an absolute module.manifest_path"})
}

func moduleManifestPathCandidates(cfg Config, manifestPath string, resolvedPath string, manifestRoot string) []string {
	manifestPath = filepath.Clean(manifestPath)
	candidates := []string{}
	if strings.TrimSpace(resolvedPath) != "" {
		return append(candidates, resolvedPath)
	}
	if filepath.IsAbs(manifestPath) {
		candidates = append(candidates, manifestPath)
		return candidates
	}
	if strings.TrimSpace(manifestRoot) != "" {
		return append(candidates, filepath.Join(manifestRoot, manifestPath))
	}
	if strings.TrimSpace(cfg.ModuleManifestRoot) != "" {
		return append(candidates, filepath.Join(cfg.ModuleManifestRoot, manifestPath))
	}
	candidates = append(candidates, manifestPath)
	return candidates
}

func verifyModuleManifest(path string, provider string, expectedName string, policy moduleManifestPolicy) (semanticModuleManifest, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return semanticModuleManifest{}, "", domain.InternalError("read "+policy.label+" manifest", err)
	}
	sum := sha256.Sum256(data)
	sha := hex.EncodeToString(sum[:])
	var manifest semanticModuleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return semanticModuleManifest{}, "", domain.ValidationError("decode "+policy.label+" manifest", map[string]any{"error": err.Error()})
	}
	if manifest.SchemaVersion != "openclerk-module.v1" {
		return semanticModuleManifest{}, "", domain.ValidationError(policy.label+" manifest schema_version must be openclerk-module.v1", nil)
	}
	if strings.TrimSpace(manifest.Module.Name) == "" {
		return semanticModuleManifest{}, "", domain.ValidationError(policy.label+" manifest module.name is required", nil)
	}
	if strings.TrimSpace(expectedName) != "" && strings.TrimSpace(expectedName) != manifest.Module.Name {
		return semanticModuleManifest{}, "", domain.ValidationError(policy.label+" manifest module.name mismatch", map[string]any{"expected": expectedName, "actual": manifest.Module.Name})
	}
	if !policy.providerMatches(manifest.Module.Name, provider) {
		return semanticModuleManifest{}, "", domain.ValidationError(policy.label+" manifest module.name must identify the provider", map[string]any{"provider": provider, "module": manifest.Module.Name})
	}
	if manifest.Module.Kind != policy.kind {
		return semanticModuleManifest{}, "", domain.ValidationError(policy.label+" manifest module.kind must be "+policy.kind, nil)
	}
	if manifest.Authority.Default != "read_only" || manifest.Authority.DurableWrites != "forbidden" {
		return semanticModuleManifest{}, "", domain.ValidationError(policy.label+" manifest must be read-only and forbid durable writes", nil)
	}
	if manifest.Release.Status != "supported_optional_module" {
		return semanticModuleManifest{}, "", domain.ValidationError(policy.label+" manifest release.status must be supported_optional_module", nil)
	}
	hasRequiredCommand := false
	for _, provided := range manifest.Provides {
		if provided.Type == "command" && policy.providesRequiredCommand(provided.Name) {
			hasRequiredCommand = true
			break
		}
	}
	if !hasRequiredCommand {
		return semanticModuleManifest{}, "", domain.ValidationError(policy.missingCommandMessage, nil)
	}
	return manifest, sha, nil
}

func semanticModuleProviderMatches(moduleName string, provider string) bool {
	return moduleName == provider+"-embeddings"
}

type moduleRuntimeConfigStore struct {
	label string
	key   func(provider string, field string) string
}

var (
	semanticModuleRuntimeConfig = moduleRuntimeConfigStore{label: "semantic module", key: semanticModuleKey}
	ocrModuleRuntimeConfig      = moduleRuntimeConfigStore{label: "OCR module", key: ocrModuleKey}
)

func (store moduleRuntimeConfigStore) write(ctx context.Context, cfg Config, config SemanticModuleConfig) error {
	values := moduleRuntimeConfigValues(config)
	return withRuntimeConfigDB(ctx, cfg, true, func(db *sql.DB) error {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		for key, value := range values {
			if _, err := db.ExecContext(ctx, `
INSERT INTO runtime_config (key_name, value_text, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key_name) DO UPDATE SET
	value_text = excluded.value_text,
	updated_at = excluded.updated_at`, store.key(config.Provider, key), value, now); err != nil {
				return domain.InternalError("write "+store.label+" runtime config", err)
			}
		}
		return nil
	})
}

func moduleRuntimeConfigValues(config SemanticModuleConfig) map[string]string {
	return map[string]string{
		"enabled":                fmt.Sprint(config.Enabled),
		"module_name":            config.ModuleName,
		"command":                config.Command,
		"manifest_path":          config.ManifestPath,
		"manifest_sha256":        config.ManifestSHA256,
		"manifest_resolved_path": config.ManifestResolvedPath,
		"command_args_json":      mustMarshalString(config.CommandArgs),
		"provider_config_json":   mustMarshalString(config.ProviderConfig),
	}
}

func (store moduleRuntimeConfigStore) delete(ctx context.Context, cfg Config, provider string) error {
	return withRuntimeConfigDB(ctx, cfg, true, func(db *sql.DB) error {
		prefix := store.key(provider, "")
		_, err := db.ExecContext(ctx, `DELETE FROM runtime_config WHERE key_name LIKE ?`, prefix+"%")
		if err != nil {
			return domain.InternalError("remove "+store.label+" runtime config", err)
		}
		return nil
	})
}

func (store moduleRuntimeConfigStore) read(ctx context.Context, cfg Config, provider string) (map[string]string, error) {
	values := map[string]string{}
	err := withRuntimeConfigDB(ctx, cfg, false, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT key_name, value_text FROM runtime_config WHERE key_name LIKE ?`, store.key(provider, "")+"%")
		if err != nil {
			return domain.InternalError("read "+store.label+" runtime config", err)
		}
		defer func() {
			_ = rows.Close()
		}()
		prefix := store.key(provider, "")
		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err != nil {
				return domain.InternalError("scan "+store.label+" runtime config", err)
			}
			values[strings.TrimPrefix(key, prefix)] = value
		}
		if err := rows.Err(); err != nil {
			return domain.InternalError("iterate "+store.label+" runtime config", err)
		}
		return nil
	})
	return values, err
}

func writeSemanticModuleConfig(ctx context.Context, cfg Config, config SemanticModuleConfig) error {
	return semanticModuleRuntimeConfig.write(ctx, cfg, config)
}

func writeOCRModuleConfig(ctx context.Context, cfg Config, config SemanticModuleConfig) error {
	return ocrModuleRuntimeConfig.write(ctx, cfg, config)
}

func deleteSemanticModuleConfig(ctx context.Context, cfg Config, provider string) error {
	return semanticModuleRuntimeConfig.delete(ctx, cfg, provider)
}

func deleteOCRModuleConfig(ctx context.Context, cfg Config, provider string) error {
	return ocrModuleRuntimeConfig.delete(ctx, cfg, provider)
}

func readSemanticModuleValues(ctx context.Context, cfg Config, provider string) (map[string]string, error) {
	return semanticModuleRuntimeConfig.read(ctx, cfg, provider)
}

func readOCRModuleValues(ctx context.Context, cfg Config, provider string) (map[string]string, error) {
	return ocrModuleRuntimeConfig.read(ctx, cfg, provider)
}

func withRuntimeConfigDB(ctx context.Context, cfg Config, write bool, fn func(*sql.DB) error) error {
	paths, err := ResolvePaths(cfg)
	if err != nil {
		return err
	}
	var unlock func()
	if write {
		unlock, err = acquireRuntimeConfigLock(ctx, paths.DatabasePath)
		if err != nil {
			return err
		}
		defer unlock()
	}
	db, err := sql.Open("sqlite", paths.DatabasePath)
	if err != nil {
		return domain.InternalError("open runtime config database", err)
	}
	defer func() {
		_ = db.Close()
	}()
	for _, statement := range []string{
		`PRAGMA busy_timeout = 5000;`,
		`CREATE TABLE IF NOT EXISTS runtime_config (
			key_name TEXT PRIMARY KEY,
			value_text TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return domain.InternalError("initialize runtime config", err)
		}
	}
	return fn(db)
}

func semanticModuleKey(provider string, field string) string {
	return "semantic_module." + provider + "." + field
}

func ocrModuleKey(provider string, field string) string {
	return "ocr_module." + provider + "." + field
}

func redactedProviderConfig(provider string, values map[string]string) map[string]string {
	config := map[string]string{}
	for key, value := range values {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if normalizedKey == "" {
			continue
		}
		trimmedValue := strings.TrimSpace(value)
		if semanticProviderConfigSecretKey(normalizedKey) {
			config[normalizedKey] = "redacted"
			continue
		}
		config[normalizedKey] = trimmedValue
	}
	if provider == SemanticModuleProviderGemini {
		config["credential_ref"] = "runtime_config:GEMINI_API_KEY"
		delete(config, "gemini_api_key")
		delete(config, "api_key")
	}
	return config
}

func validateSemanticProviderConfig(provider string, config map[string]string) error {
	switch provider {
	case SemanticModuleProviderOllama:
		if rejection := validateOptionalModuleLoopbackHTTPURL(config["ollama_url"], "module.provider_config.ollama_url"); rejection != "" {
			return domain.ValidationError(rejection, map[string]any{"provider": provider})
		}
	case SemanticModuleProviderGemini:
		if rejection := validateOptionalModuleCanonicalGeminiAPIBase(config["gemini_api_base"], "module.provider_config.gemini_api_base"); rejection != "" {
			return domain.ValidationError(rejection, map[string]any{"provider": provider})
		}
		if key := strings.TrimSpace(config["gemini_config_key"]); key != "" && key != "GEMINI_API_KEY" {
			return domain.ValidationError("module.provider_config.gemini_config_key must be GEMINI_API_KEY", map[string]any{"provider": provider})
		}
	}
	return nil
}

func validateOptionalModuleLoopbackHTTPURL(raw string, field string) string {
	parsed, rejection := validateOptionalModuleHTTPURL(raw, field)
	if rejection != "" {
		return field + " must be a loopback HTTP URL"
	}
	if parsed == nil {
		return ""
	}
	if strings.EqualFold(parsed.Hostname(), "localhost") {
		return ""
	}
	ip := net.ParseIP(parsed.Hostname())
	if ip != nil && ip.IsLoopback() {
		return ""
	}
	return field + " must be a loopback HTTP URL"
}

func validateOptionalModuleCanonicalGeminiAPIBase(raw string, field string) string {
	parsed, rejection := validateOptionalModuleHTTPURL(raw, field)
	if rejection != "" {
		return field + " must be " + canonicalGeminiAPIBase
	}
	if parsed == nil {
		return ""
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" ||
		parsed.Scheme != "https" || parsed.Host != "generativelanguage.googleapis.com" || parsed.Path != "/v1beta" {
		return field + " must be " + canonicalGeminiAPIBase
	}
	return ""
}

func validateOptionalModuleHTTPURL(raw string, field string) (*url.URL, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, field + " must be a valid http or https URL"
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, field + " must use http or https"
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return nil, field + " must not include userinfo, query, or fragment"
	}
	return parsed, ""
}

func semanticProviderConfigSecretKey(key string) bool {
	return strings.Contains(key, "secret") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "password") ||
		strings.Contains(key, "api_key")
}

func sanitizedArgs(args []string) []string {
	clean := []string{}
	for _, arg := range args {
		clean = append(clean, strings.TrimSpace(arg))
	}
	return clean
}

func mustMarshalString(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func IsSemanticModuleNotConfigured(err error) bool {
	var domainErr *domain.Error
	return errors.As(err, &domainErr) && domainErr.Code == "validation_error" && strings.Contains(err.Error(), "not installed")
}
