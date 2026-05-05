package runclient

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
)

type SemanticModuleConfig struct {
	Provider           string            `json:"provider"`
	ModuleName         string            `json:"module_name"`
	Enabled            bool              `json:"enabled"`
	Command            string            `json:"command,omitempty"`
	CommandArgs        []string          `json:"command_args,omitempty"`
	ManifestPath       string            `json:"manifest_path,omitempty"`
	ManifestSHA256     string            `json:"manifest_sha256,omitempty"`
	ProviderConfig     map[string]string `json:"provider_config,omitempty"`
	VerificationStatus string            `json:"verification_status"`
	RedactionStatus    string            `json:"redaction_status"`
}

type SemanticModuleInstallInput struct {
	Provider       string            `json:"provider"`
	ModuleName     string            `json:"module_name,omitempty"`
	ManifestPath   string            `json:"manifest_path"`
	Command        string            `json:"command"`
	CommandArgs    []string          `json:"command_args,omitempty"`
	ProviderConfig map[string]string `json:"provider_config,omitempty"`
	Enabled        *bool             `json:"enabled,omitempty"`
}

type SemanticModuleConfigureInput struct {
	Provider       string            `json:"provider"`
	ProviderConfig map[string]string `json:"provider_config,omitempty"`
	Enabled        *bool             `json:"enabled,omitempty"`
}

func InstallSemanticModule(ctx context.Context, cfg Config, input SemanticModuleInstallInput) (SemanticModuleConfig, error) {
	provider, err := normalizeSemanticModuleProvider(input.Provider)
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
	manifest, manifestSHA, err := verifySemanticModuleManifest(resolveSemanticModuleManifestPath(cfg, manifestPath), provider, input.ModuleName)
	if err != nil {
		return SemanticModuleConfig{}, err
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	config := SemanticModuleConfig{
		Provider:           provider,
		ModuleName:         manifest.Module.Name,
		Enabled:            enabled,
		Command:            strings.TrimSpace(input.Command),
		CommandArgs:        sanitizedArgs(input.CommandArgs),
		ManifestPath:       manifestPath,
		ManifestSHA256:     manifestSHA,
		ProviderConfig:     redactedProviderConfig(provider, input.ProviderConfig),
		VerificationStatus: "verified",
		RedactionStatus:    "redacted",
	}
	if err := writeSemanticModuleConfig(ctx, cfg, config); err != nil {
		return SemanticModuleConfig{}, err
	}
	return config, nil
}

func ConfigureSemanticModule(ctx context.Context, cfg Config, input SemanticModuleConfigureInput) (SemanticModuleConfig, error) {
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
	for key, value := range redactedProviderConfig(provider, input.ProviderConfig) {
		merged[key] = value
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
	config := SemanticModuleConfig{
		Provider:           normalized,
		ModuleName:         values["module_name"],
		Enabled:            values["enabled"] == "true",
		Command:            values["command"],
		ManifestPath:       values["manifest_path"],
		ManifestSHA256:     values["manifest_sha256"],
		ProviderConfig:     map[string]string{},
		VerificationStatus: "not_installed",
		RedactionStatus:    "redacted",
	}
	_ = json.Unmarshal([]byte(values["command_args_json"]), &config.CommandArgs)
	_ = json.Unmarshal([]byte(values["provider_config_json"]), &config.ProviderConfig)
	if config.ProviderConfig == nil {
		config.ProviderConfig = map[string]string{}
	}
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

func verifyInstalledSemanticModule(cfg Config, config SemanticModuleConfig) error {
	if strings.TrimSpace(config.Command) == "" {
		return domain.ValidationError("semantic module command is missing", map[string]any{"provider": config.Provider})
	}
	if strings.TrimSpace(config.ManifestPath) == "" || strings.TrimSpace(config.ManifestSHA256) == "" {
		return domain.ValidationError("semantic module manifest verification is missing", map[string]any{"provider": config.Provider})
	}
	manifest, sha, err := verifySemanticModuleManifest(resolveSemanticModuleManifestPath(cfg, config.ManifestPath), config.Provider, config.ModuleName)
	if err != nil {
		return err
	}
	if sha != config.ManifestSHA256 {
		return domain.ValidationError("semantic module manifest digest changed", map[string]any{"provider": config.Provider, "module": manifest.Module.Name})
	}
	return nil
}

func resolveSemanticModuleManifestPath(cfg Config, manifestPath string) string {
	if cfg.ModuleManifestRoot == "" || filepath.IsAbs(manifestPath) {
		return manifestPath
	}
	return filepath.Join(cfg.ModuleManifestRoot, manifestPath)
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

func verifySemanticModuleManifest(path string, provider string, expectedName string) (semanticModuleManifest, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return semanticModuleManifest{}, "", domain.InternalError("read semantic module manifest", err)
	}
	sum := sha256.Sum256(data)
	sha := hex.EncodeToString(sum[:])
	var manifest semanticModuleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return semanticModuleManifest{}, "", domain.ValidationError("decode semantic module manifest", map[string]any{"error": err.Error()})
	}
	if manifest.SchemaVersion != "openclerk-module.v1" {
		return semanticModuleManifest{}, "", domain.ValidationError("semantic module manifest schema_version must be openclerk-module.v1", nil)
	}
	if strings.TrimSpace(manifest.Module.Name) == "" {
		return semanticModuleManifest{}, "", domain.ValidationError("semantic module manifest module.name is required", nil)
	}
	if strings.TrimSpace(expectedName) != "" && strings.TrimSpace(expectedName) != manifest.Module.Name {
		return semanticModuleManifest{}, "", domain.ValidationError("semantic module manifest module.name mismatch", map[string]any{"expected": expectedName, "actual": manifest.Module.Name})
	}
	wantSuffix := "-" + provider
	if !strings.Contains(manifest.Module.Name, provider) && !strings.HasSuffix(manifest.Module.Name, wantSuffix) {
		return semanticModuleManifest{}, "", domain.ValidationError("semantic module manifest module.name must identify the provider", map[string]any{"provider": provider, "module": manifest.Module.Name})
	}
	if manifest.Module.Kind != "embedding_provider" {
		return semanticModuleManifest{}, "", domain.ValidationError("semantic module manifest module.kind must be embedding_provider", nil)
	}
	if manifest.Authority.Default != "read_only" || manifest.Authority.DurableWrites != "forbidden" {
		return semanticModuleManifest{}, "", domain.ValidationError("semantic module manifest must be read-only and forbid durable writes", nil)
	}
	if manifest.Release.Status != "supported_optional_module" {
		return semanticModuleManifest{}, "", domain.ValidationError("semantic module manifest release.status must be supported_optional_module", nil)
	}
	hasSearch := false
	for _, provided := range manifest.Provides {
		if provided.Type == "command" && strings.Contains(provided.Name, " search") {
			hasSearch = true
			break
		}
	}
	if !hasSearch {
		return semanticModuleManifest{}, "", domain.ValidationError("semantic module manifest must provide a search command", nil)
	}
	return manifest, sha, nil
}

func writeSemanticModuleConfig(ctx context.Context, cfg Config, config SemanticModuleConfig) error {
	values := map[string]string{
		"enabled":              fmt.Sprint(config.Enabled),
		"module_name":          config.ModuleName,
		"command":              config.Command,
		"manifest_path":        config.ManifestPath,
		"manifest_sha256":      config.ManifestSHA256,
		"command_args_json":    mustMarshalString(config.CommandArgs),
		"provider_config_json": mustMarshalString(config.ProviderConfig),
	}
	return withRuntimeConfigDB(ctx, cfg, true, func(db *sql.DB) error {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		for key, value := range values {
			if _, err := db.ExecContext(ctx, `
INSERT INTO runtime_config (key_name, value_text, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key_name) DO UPDATE SET
	value_text = excluded.value_text,
	updated_at = excluded.updated_at`, semanticModuleKey(config.Provider, key), value, now); err != nil {
				return domain.InternalError("write semantic module runtime config", err)
			}
		}
		return nil
	})
}

func deleteSemanticModuleConfig(ctx context.Context, cfg Config, provider string) error {
	return withRuntimeConfigDB(ctx, cfg, true, func(db *sql.DB) error {
		prefix := semanticModuleKey(provider, "")
		_, err := db.ExecContext(ctx, `DELETE FROM runtime_config WHERE key_name LIKE ?`, prefix+"%")
		if err != nil {
			return domain.InternalError("remove semantic module runtime config", err)
		}
		return nil
	})
}

func readSemanticModuleValues(ctx context.Context, cfg Config, provider string) (map[string]string, error) {
	values := map[string]string{}
	err := withRuntimeConfigDB(ctx, cfg, false, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT key_name, value_text FROM runtime_config WHERE key_name LIKE ?`, semanticModuleKey(provider, "")+"%")
		if err != nil {
			return domain.InternalError("read semantic module runtime config", err)
		}
		defer func() {
			_ = rows.Close()
		}()
		prefix := semanticModuleKey(provider, "")
		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err != nil {
				return domain.InternalError("scan semantic module runtime config", err)
			}
			values[strings.TrimPrefix(key, prefix)] = value
		}
		if err := rows.Err(); err != nil {
			return domain.InternalError("iterate semantic module runtime config", err)
		}
		return nil
	})
	return values, err
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
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func IsSemanticModuleNotConfigured(err error) bool {
	var domainErr *domain.Error
	return errors.As(err, &domainErr) && domainErr.Code == "validation_error" && strings.Contains(err.Error(), "not installed")
}
