package runclient

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSemanticModuleRuntimeConfigRedactsAndVerifiesManifest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
	commandPath := writeRunclientExecutable(t, semanticModuleCommand)
	installed, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
		Provider:     "gemini",
		ManifestPath: manifestPath,
		Command:      commandPath,
		ProviderConfig: map[string]string{
			"embedding_model": "gemini-embedding-001",
			"api_key":         "must-not-be-stored",
		},
	})
	if err != nil {
		t.Fatalf("install module: %v", err)
	}
	if installed.ProviderConfig["credential_ref"] != "runtime_config:GEMINI_API_KEY" ||
		installed.ProviderConfig["api_key"] != "" ||
		installed.VerificationStatus != "verified" {
		t.Fatalf("installed config = %+v", installed)
	}

	read, err := ReadSemanticModuleConfig(ctx, config, "gemini")
	if err != nil {
		t.Fatalf("read module: %v", err)
	}
	if !read.Enabled || read.ManifestSHA256 == "" || read.ProviderConfig["api_key"] != "" {
		t.Fatalf("read config = %+v", read)
	}

	if err := os.WriteFile(manifestPath, []byte(strings.ReplaceAll(string(mustReadFile(t, manifestPath)), "0.1.0", "0.1.1")), 0o600); err != nil {
		t.Fatalf("mutate manifest: %v", err)
	}
	changed, err := ReadSemanticModuleConfig(ctx, config, "gemini")
	if err == nil || !strings.Contains(err.Error(), "digest changed") || changed.VerificationStatus != "verification_failed" {
		t.Fatalf("changed config=%+v err=%v", changed, err)
	}
}

func TestSemanticModuleProviderConfigRejectsUnsafeEndpointConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("remote ollama install", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
		_, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderOllama,
			ManifestPath: manifestPath,
			ProviderConfig: map[string]string{
				"ollama_url": "https://embeddings.example.test",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "module.provider_config.ollama_url must be a loopback HTTP URL") {
			t.Fatalf("install error = %v", err)
		}
	})

	t.Run("remote gemini install", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
		_, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderGemini,
			ManifestPath: manifestPath,
			ProviderConfig: map[string]string{
				"gemini_api_base": "http://127.0.0.1:9999",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "module.provider_config.gemini_api_base must be https://generativelanguage.googleapis.com/v1beta") {
			t.Fatalf("install error = %v", err)
		}
	})

	t.Run("remote gemini configure", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
		if _, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderGemini,
			ManifestPath: manifestPath,
			Command:      writeRunclientExecutable(t, semanticModuleCommand),
		}); err != nil {
			t.Fatalf("install module: %v", err)
		}
		_, err := ConfigureSemanticModule(ctx, config, SemanticModuleConfigureInput{
			Provider: SemanticModuleProviderGemini,
			ProviderConfig: map[string]string{
				"gemini_api_base": "https://generativelanguage.googleapis.com/v1beta?key=inline",
			},
		})
		if err == nil || !strings.Contains(err.Error(), "module.provider_config.gemini_api_base must be https://generativelanguage.googleapis.com/v1beta") {
			t.Fatalf("configure error = %v", err)
		}
	})

	t.Run("disable legacy unsafe config", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
		installed, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderOllama,
			ManifestPath: manifestPath,
			Command:      writeRunclientExecutable(t, semanticModuleCommand),
		})
		if err != nil {
			t.Fatalf("install module: %v", err)
		}
		installed.ProviderConfig["ollama_url"] = "https://embeddings.example.test"
		if err := writeSemanticModuleConfig(ctx, config, installed); err != nil {
			t.Fatalf("poison legacy config: %v", err)
		}

		disabled := false
		configured, err := ConfigureSemanticModule(ctx, config, SemanticModuleConfigureInput{
			Provider: SemanticModuleProviderOllama,
			Enabled:  &disabled,
		})
		if err != nil {
			t.Fatalf("disable legacy unsafe config: %v", err)
		}
		if configured.Enabled {
			t.Fatalf("configured module still enabled: %+v", configured)
		}

		enabled := true
		_, err = ConfigureSemanticModule(ctx, config, SemanticModuleConfigureInput{
			Provider: SemanticModuleProviderOllama,
			Enabled:  &enabled,
		})
		if err == nil || !strings.Contains(err.Error(), "module.provider_config.ollama_url must be a loopback HTTP URL") {
			t.Fatalf("reenable error = %v", err)
		}
	})
}

func TestOCRModuleRuntimeConfigVerifiesManifest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	manifestPath := writeRunclientOCRModuleManifest(t, t.TempDir())
	tesseractPath := writeRunclientExecutable(t, ocrModuleTesseractCommand)
	ocrmypdfPath := writeRunclientExecutable(t, ocrModuleOCRmyPDFCommand)
	installed, err := InstallOCRModule(ctx, config, SemanticModuleInstallInput{
		Kind:         ModuleKindOCRProvider,
		Provider:     OCRModuleProviderTesseract,
		ManifestPath: manifestPath,
		Command:      tesseractPath,
		ProviderConfig: map[string]string{
			"ocrmypdf_command": ocrmypdfPath,
			"language":         "eng",
		},
	})
	if err != nil {
		t.Fatalf("install OCR module: %v", err)
	}
	if installed.Kind != ModuleKindOCRProvider ||
		installed.Provider != OCRModuleProviderTesseract ||
		installed.ProviderConfig["language"] != "eng" ||
		installed.VerificationStatus != "verified" {
		t.Fatalf("installed OCR config = %+v", installed)
	}

	read, err := ReadOCRModuleConfig(ctx, config, OCRModuleProviderTesseract)
	if err != nil {
		t.Fatalf("read OCR module: %v", err)
	}
	if !read.Enabled || read.ManifestSHA256 == "" || read.Kind != ModuleKindOCRProvider {
		t.Fatalf("read OCR config = %+v", read)
	}

	if err := os.WriteFile(manifestPath, []byte(strings.ReplaceAll(string(mustReadFile(t, manifestPath)), "0.1.0", "0.1.1")), 0o600); err != nil {
		t.Fatalf("mutate manifest: %v", err)
	}
	changed, err := ReadOCRModuleConfig(ctx, config, OCRModuleProviderTesseract)
	if err == nil || !strings.Contains(err.Error(), "digest changed") || changed.VerificationStatus != "verification_failed" {
		t.Fatalf("changed OCR config=%+v err=%v", changed, err)
	}
}

func TestModuleRuntimeConfigConfigureRemoveAndList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	semanticManifest := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
	ocrManifest := writeRunclientOCRModuleManifest(t, t.TempDir())
	semanticCommand := writeRunclientExecutable(t, semanticModuleCommand)
	tesseractPath := writeRunclientExecutable(t, ocrModuleTesseractCommand)
	ocrmypdfPath := writeRunclientExecutable(t, ocrModuleOCRmyPDFCommand)
	disabled := false

	if _, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
		Provider:     SemanticModuleProviderOllama,
		ManifestPath: semanticManifest,
		Command:      semanticCommand,
		ProviderConfig: map[string]string{
			"embedding_model": "embeddinggemma",
		},
	}); err != nil {
		t.Fatalf("install semantic module: %v", err)
	}
	if _, err := InstallOCRModule(ctx, config, SemanticModuleInstallInput{
		Kind:         ModuleKindOCRProvider,
		Provider:     OCRModuleProviderTesseract,
		ManifestPath: ocrManifest,
		Command:      tesseractPath,
		ProviderConfig: map[string]string{
			"ocrmypdf_command": ocrmypdfPath,
		},
	}); err != nil {
		t.Fatalf("install OCR module: %v", err)
	}
	configured, err := ConfigureSemanticModule(ctx, config, SemanticModuleConfigureInput{
		Provider:       SemanticModuleProviderOllama,
		Enabled:        &disabled,
		ProviderConfig: map[string]string{"ollama_url": "http://localhost:11434"},
	})
	if err != nil {
		t.Fatalf("configure semantic module: %v", err)
	}
	if configured.Enabled || configured.ProviderConfig["embedding_model"] != "embeddinggemma" || configured.ProviderConfig["ollama_url"] != "http://localhost:11434" {
		t.Fatalf("configured semantic module = %+v", configured)
	}
	list, err := ListSemanticModules(ctx, config)
	if err != nil {
		t.Fatalf("list modules: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("module list length = %d, want 2; modules=%+v", len(list), list)
	}
	removed, err := RemoveOCRModule(ctx, config, OCRModuleProviderTesseract)
	if err != nil {
		t.Fatalf("remove OCR module: %v", err)
	}
	if removed.Enabled || removed.VerificationStatus != "removed" || removed.Kind != ModuleKindOCRProvider {
		t.Fatalf("removed OCR module = %+v", removed)
	}
	list, err = ListSemanticModules(ctx, config)
	if err != nil {
		t.Fatalf("list modules after remove: %v", err)
	}
	if len(list) != 1 || list[0].Provider != SemanticModuleProviderOllama {
		t.Fatalf("module list after remove = %+v", list)
	}
}

func TestModuleManifestVerificationSurvivesCWDChange(t *testing.T) {
	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	releaseRoot := t.TempDir()
	semanticManifest := writeRunclientSemanticModuleManifest(t, filepath.Join(releaseRoot, "modules", "ollama-embeddings"), "ollama")
	ocrManifest := writeRunclientOCRModuleManifest(t, filepath.Join(releaseRoot, "modules", "tesseract-ocr"))
	semanticCommand := writeRunclientExecutable(t, semanticModuleCommand)
	tesseractPath := writeRunclientExecutable(t, ocrModuleTesseractCommand)
	ocrmypdfPath := writeRunclientExecutable(t, ocrModuleOCRmyPDFCommand)

	originalCWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(releaseRoot); err != nil {
		t.Fatalf("chdir release root: %v", err)
	}
	semanticRel, err := filepath.Rel(releaseRoot, semanticManifest)
	if err != nil {
		t.Fatalf("semantic relative path: %v", err)
	}
	ocrRel, err := filepath.Rel(releaseRoot, ocrManifest)
	if err != nil {
		t.Fatalf("OCR relative path: %v", err)
	}
	if _, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
		Provider:     SemanticModuleProviderOllama,
		ManifestPath: filepath.ToSlash(semanticRel),
		Command:      semanticCommand,
	}); err != nil {
		_ = os.Chdir(originalCWD)
		t.Fatalf("install semantic module: %v", err)
	}
	if _, err := InstallOCRModule(ctx, config, SemanticModuleInstallInput{
		Kind:         ModuleKindOCRProvider,
		Provider:     OCRModuleProviderTesseract,
		ManifestPath: filepath.ToSlash(ocrRel),
		Command:      tesseractPath,
		ProviderConfig: map[string]string{
			"ocrmypdf_command": ocrmypdfPath,
		},
	}); err != nil {
		_ = os.Chdir(originalCWD)
		t.Fatalf("install OCR module: %v", err)
	}
	otherCWD := t.TempDir()
	if err := os.Chdir(otherCWD); err != nil {
		_ = os.Chdir(originalCWD)
		t.Fatalf("chdir other cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalCWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	modules, err := ListSemanticModules(ctx, config)
	if err != nil {
		t.Fatalf("list modules from other cwd: %v", err)
	}
	if len(modules) != 2 {
		t.Fatalf("modules = %+v, want semantic and OCR modules", modules)
	}
	for _, module := range modules {
		data, err := json.Marshal(module)
		if err != nil {
			t.Fatalf("marshal module: %v", err)
		}
		if strings.Contains(string(data), "manifest_resolved_path") || strings.Contains(string(data), releaseRoot) {
			t.Fatalf("public module JSON exposed private resolved path: %s", string(data))
		}
	}
}

func TestModuleManifestRootInstallInputResolvesRelativeManifests(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	releaseRoot := t.TempDir()
	manifestPath := writeRunclientSemanticModuleManifest(t, filepath.Join(releaseRoot, "modules", "gemini-embeddings"), "gemini")
	commandPath := writeRunclientExecutable(t, semanticModuleCommand)
	manifestRel, err := filepath.Rel(releaseRoot, manifestPath)
	if err != nil {
		t.Fatalf("relative manifest path: %v", err)
	}

	installed, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
		Provider:     SemanticModuleProviderGemini,
		ManifestPath: filepath.ToSlash(manifestRel),
		ManifestRoot: releaseRoot,
		Command:      commandPath,
		ProviderConfig: map[string]string{
			"embedding_model": "gemini-embedding-001",
			"gemini_api_base": canonicalGeminiAPIBase,
		},
	})
	if err != nil {
		t.Fatalf("install semantic module with manifest_root: %v", err)
	}
	if installed.ManifestPath != filepath.ToSlash(manifestRel) || strings.TrimSpace(installed.ManifestResolvedPath) == "" {
		t.Fatalf("installed config = %+v, want public relative path plus private resolved path", installed)
	}
	data, err := json.Marshal(installed)
	if err != nil {
		t.Fatalf("marshal installed module: %v", err)
	}
	if strings.Contains(string(data), "manifest_resolved_path") || strings.Contains(string(data), releaseRoot) {
		t.Fatalf("installed module JSON exposed private resolved path: %s", string(data))
	}
	read, err := ReadSemanticModuleConfig(ctx, config, SemanticModuleProviderGemini)
	if err != nil {
		t.Fatalf("read semantic module: %v", err)
	}
	if read.VerificationStatus != "verified" || strings.TrimSpace(read.ManifestResolvedPath) == "" {
		t.Fatalf("read config = %+v, want verified with private resolved path", read)
	}
}

func TestModuleManifestRootInstallInputIsAuthoritative(t *testing.T) {
	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	currentRoot := t.TempDir()
	manifestPath := writeRunclientSemanticModuleManifest(t, filepath.Join(currentRoot, "modules", "ollama-embeddings"), "ollama")
	commandPath := writeRunclientExecutable(t, semanticModuleCommand)
	manifestRel, err := filepath.Rel(currentRoot, manifestPath)
	if err != nil {
		t.Fatalf("relative manifest path: %v", err)
	}
	originalCWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(currentRoot); err != nil {
		t.Fatalf("chdir current root: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalCWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	_, err = InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
		Provider:     SemanticModuleProviderOllama,
		ManifestPath: filepath.ToSlash(manifestRel),
		ManifestRoot: filepath.Join(t.TempDir(), "missing-release-root"),
		Command:      commandPath,
	})
	if err == nil || !strings.Contains(err.Error(), "semantic module manifest could not be resolved") {
		t.Fatalf("install err = %v, want authoritative manifest_root resolution failure", err)
	}
}

func TestLegacyModuleManifestConfigFallsBackToPublicPathResolution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	releaseRoot := t.TempDir()
	manifestPath := writeRunclientSemanticModuleManifest(t, filepath.Join(releaseRoot, "modules", "ollama-embeddings"), "ollama")
	manifestRel, err := filepath.Rel(releaseRoot, manifestPath)
	if err != nil {
		t.Fatalf("relative manifest path: %v", err)
	}
	config := Config{
		DatabasePath:       filepath.Join(t.TempDir(), "data", "openclerk.sqlite"),
		ModuleManifestRoot: releaseRoot,
	}
	commandPath := writeRunclientExecutable(t, semanticModuleCommand)
	manifest, sha, err := verifySemanticModuleManifest(manifestPath, SemanticModuleProviderOllama, "")
	if err != nil {
		t.Fatalf("verify fixture manifest: %v", err)
	}
	command, commandSHA, err := resolveBoundModuleExecutable(commandPath, semanticModuleCommand, manifest, semanticModuleManifestPolicy)
	if err != nil {
		t.Fatalf("resolve fixture command: %v", err)
	}
	writeLegacyModuleRuntimeConfig(t, ctx, config, semanticModuleRuntimeConfig, SemanticModuleConfig{
		Kind:               ModuleKindEmbeddingProvider,
		Provider:           SemanticModuleProviderOllama,
		ModuleName:         "ollama-embeddings",
		Enabled:            true,
		Command:            command,
		CommandSHA256:      commandSHA,
		ManifestPath:       filepath.ToSlash(manifestRel),
		ManifestSHA256:     sha,
		ProviderConfig:     map[string]string{},
		VerificationStatus: "verified",
		RedactionStatus:    "redacted",
	})

	read, err := ReadSemanticModuleConfig(ctx, config, SemanticModuleProviderOllama)
	if err != nil {
		t.Fatalf("read legacy module config: %v", err)
	}
	if read.VerificationStatus != "verified" || read.ManifestPath != filepath.ToSlash(manifestRel) {
		t.Fatalf("legacy read = %+v, want verified public path", read)
	}
}

func TestModuleManifestVerificationUsesSharedResolver(t *testing.T) {
	t.Parallel()

	source := string(mustReadFile(t, "modules.go"))
	if strings.Contains(source, "resolveSemanticModuleManifestPath") {
		t.Fatalf("modules.go still contains legacy semantic-only manifest resolver")
	}
	if got := strings.Count(source, "verifyInstalledModuleManifest(cfg, config,"); got != 2 {
		t.Fatalf("installed module verification shared resolver call count = %d, want 2", got)
	}
	if got := strings.Count(source, "resolveModuleManifestPathForInstall(cfg, input.ManifestRoot, manifestPath,"); got != 2 {
		t.Fatalf("install manifest resolver call count = %d, want 2", got)
	}
}

func TestInstallSemanticModuleLocksCommandSurface(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects noncanonical command", func(t *testing.T) {
		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
		_, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderOllama,
			ManifestPath: manifestPath,
			Command:      "/bin/sh",
		})
		if err == nil || !strings.Contains(err.Error(), "semantic module command must be semantic-retrieval-adapter") {
			t.Fatalf("err = %v, want semantic command rejection", err)
		}
	})

	t.Run("rejects command args", func(t *testing.T) {
		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
		_, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderOllama,
			ManifestPath: manifestPath,
			Command:      semanticModuleCommand,
			CommandArgs:  []string{"-c", "echo pwned"},
		})
		if err == nil || !strings.Contains(err.Error(), "module.command_args are unsupported for semantic modules") {
			t.Fatalf("err = %v, want semantic command_args rejection", err)
		}
	})

	t.Run("resolves canonical command when omitted", func(t *testing.T) {
		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
		commandPath := writeRunclientExecutable(t, semanticModuleCommand)
		t.Setenv("PATH", filepath.Dir(commandPath)+string(os.PathListSeparator)+os.Getenv("PATH"))
		installed, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderOllama,
			ManifestPath: manifestPath,
		})
		if err != nil {
			t.Fatalf("install semantic module: %v", err)
		}
		if installed.Command != commandPath || installed.CommandSHA256 == "" || len(installed.CommandArgs) != 0 {
			t.Fatalf("installed command surface = %q sha=%q args=%v, want resolved command without args", installed.Command, installed.CommandSHA256, installed.CommandArgs)
		}
	})

	t.Run("resolves canonical command when provided", func(t *testing.T) {
		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
		commandPath := writeRunclientExecutable(t, semanticModuleCommand)
		installed, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderGemini,
			ManifestPath: manifestPath,
			Command:      commandPath,
		})
		if err != nil {
			t.Fatalf("install semantic module: %v", err)
		}
		if installed.Command != commandPath || installed.CommandSHA256 == "" || len(installed.CommandArgs) != 0 {
			t.Fatalf("installed command surface = %q sha=%q args=%v, want resolved command without args", installed.Command, installed.CommandSHA256, installed.CommandArgs)
		}
	})
}

func TestReadSemanticModuleConfigRejectsPoisonedStoredCommand(t *testing.T) {
	t.Parallel()

	t.Run("command", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		installed := installRunclientSemanticModuleForPoisonTest(t, config)
		installed.Command = "/bin/sh"
		if err := writeSemanticModuleConfig(context.Background(), config, installed); err != nil {
			t.Fatalf("write poisoned config: %v", err)
		}

		read, err := ReadSemanticModuleConfig(context.Background(), config, SemanticModuleProviderOllama)
		if err == nil || !strings.Contains(err.Error(), "semantic module command must be semantic-retrieval-adapter") {
			t.Fatalf("read=%+v err=%v, want poisoned command rejection", read, err)
		}
		if read.VerificationStatus != "verification_failed" {
			t.Fatalf("verification status = %q, want verification_failed", read.VerificationStatus)
		}
	})

	t.Run("command args", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		installed := installRunclientSemanticModuleForPoisonTest(t, config)
		installed.CommandArgs = []string{"-c", "echo pwned"}
		if err := writeSemanticModuleConfig(context.Background(), config, installed); err != nil {
			t.Fatalf("write poisoned config: %v", err)
		}

		read, err := ReadSemanticModuleConfig(context.Background(), config, SemanticModuleProviderOllama)
		if err == nil || !strings.Contains(err.Error(), "semantic module command_args are unsupported") {
			t.Fatalf("read=%+v err=%v, want poisoned command_args rejection", read, err)
		}
		if read.VerificationStatus != "verification_failed" {
			t.Fatalf("verification status = %q, want verification_failed", read.VerificationStatus)
		}
	})

	t.Run("legacy PATH command", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		installed := installRunclientSemanticModuleForPoisonTest(t, config)
		installed.Command = semanticModuleCommand
		if err := writeSemanticModuleConfig(context.Background(), config, installed); err != nil {
			t.Fatalf("write poisoned config: %v", err)
		}

		read, err := ReadSemanticModuleConfig(context.Background(), config, SemanticModuleProviderOllama)
		if err == nil || !strings.Contains(err.Error(), "command must be an absolute path") {
			t.Fatalf("read=%+v err=%v, want legacy command rejection", read, err)
		}
		if read.VerificationStatus != "verification_failed" {
			t.Fatalf("verification status = %q, want verification_failed", read.VerificationStatus)
		}
	})
}

func TestReadOCRModuleConfigRejectsUnboundCommands(t *testing.T) {
	t.Parallel()

	t.Run("primary command", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		installed := installRunclientOCRModuleForPoisonTest(t, config)
		installed.Command = "/bin/sh"
		if err := writeOCRModuleConfig(context.Background(), config, installed); err != nil {
			t.Fatalf("write poisoned OCR config: %v", err)
		}

		read, err := ReadOCRModuleConfig(context.Background(), config, OCRModuleProviderTesseract)
		if err == nil || !strings.Contains(err.Error(), "OCR module command must be tesseract") {
			t.Fatalf("read=%+v err=%v, want poisoned OCR command rejection", read, err)
		}
		if read.VerificationStatus != "verification_failed" {
			t.Fatalf("verification status = %q, want verification_failed", read.VerificationStatus)
		}
	})

	t.Run("PDF command", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		installed := installRunclientOCRModuleForPoisonTest(t, config)
		installed.ProviderConfig["ocrmypdf_command"] = "/bin/sh"
		if err := writeOCRModuleConfig(context.Background(), config, installed); err != nil {
			t.Fatalf("write poisoned OCR config: %v", err)
		}

		read, err := ReadOCRModuleConfig(context.Background(), config, OCRModuleProviderTesseract)
		if err == nil || !strings.Contains(err.Error(), "OCR module command must be ocrmypdf") {
			t.Fatalf("read=%+v err=%v, want poisoned OCR PDF command rejection", read, err)
		}
		if read.VerificationStatus != "verification_failed" {
			t.Fatalf("verification status = %q, want verification_failed", read.VerificationStatus)
		}
	})
}

func TestModuleManifestValidationRejectsSharedPolicyViolations(t *testing.T) {
	t.Parallel()

	t.Run("semantic expected name mismatch", func(t *testing.T) {
		t.Parallel()

		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
		_, _, err := verifySemanticModuleManifest(manifestPath, "gemini", "gemini-search")
		if err == nil || !strings.Contains(err.Error(), "semantic module manifest module.name mismatch") {
			t.Fatalf("err = %v, want semantic name mismatch", err)
		}
	})

	t.Run("semantic missing command", func(t *testing.T) {
		t.Parallel()

		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
		replaceManifestFile(t, manifestPath, "semantic-retrieval-adapter search", "semantic-retrieval-adapter index")
		_, _, err := verifySemanticModuleManifest(manifestPath, "gemini", "")
		if err == nil || !strings.Contains(err.Error(), "semantic module manifest must provide a search command") {
			t.Fatalf("err = %v, want semantic command validation", err)
		}
	})

	t.Run("semantic shell-looking command", func(t *testing.T) {
		t.Parallel()

		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
		replaceManifestFile(t, manifestPath, "semantic-retrieval-adapter search", "/bin/sh -c payload search")
		_, _, err := verifySemanticModuleManifest(manifestPath, "gemini", "")
		if err == nil || !strings.Contains(err.Error(), "semantic module manifest must provide a search command") {
			t.Fatalf("err = %v, want exact semantic command validation", err)
		}
	})

	t.Run("semantic search command with extra flag", func(t *testing.T) {
		t.Parallel()

		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
		replaceManifestFile(t, manifestPath, "semantic-retrieval-adapter search", "semantic-retrieval-adapter search --flag")
		_, _, err := verifySemanticModuleManifest(manifestPath, "gemini", "")
		if err == nil || !strings.Contains(err.Error(), "semantic module manifest must provide a search command") {
			t.Fatalf("err = %v, want exact semantic command validation", err)
		}
	})

	t.Run("ocr durable writes", func(t *testing.T) {
		t.Parallel()

		manifestPath := writeRunclientOCRModuleManifest(t, t.TempDir())
		replaceManifestFile(t, manifestPath, `"durable_writes":"forbidden"`, `"durable_writes":"allowed"`)
		_, _, err := verifyOCRModuleManifest(manifestPath, OCRModuleProviderTesseract, "")
		if err == nil || !strings.Contains(err.Error(), "OCR module manifest must be read-only and forbid durable writes") {
			t.Fatalf("err = %v, want OCR authority validation", err)
		}
	})
}

func writeRunclientSemanticModuleManifest(t *testing.T, dir string, provider string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create manifest dir: %v", err)
	}
	path := filepath.Join(dir, "module.json")
	manifest := map[string]any{
		"schema_version": "openclerk-module.v1",
		"module": map[string]any{
			"name":    provider + "-embeddings",
			"version": "0.1.0",
			"kind":    "embedding_provider",
		},
		"provides": []map[string]any{{
			"type": "command",
			"name": "semantic-retrieval-adapter search",
		}},
		"requires": map[string]any{
			"tools": []string{"semantic-retrieval-adapter"},
		},
		"authority": map[string]any{
			"default":        "read_only",
			"durable_writes": "forbidden",
			"forbidden":      []string{"write_documents", "change_openclerk_search_default"},
		},
		"release": map[string]any{
			"status": "supported_optional_module",
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func installRunclientSemanticModuleForPoisonTest(t *testing.T, config Config) SemanticModuleConfig {
	t.Helper()

	manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
	commandPath := writeRunclientExecutable(t, semanticModuleCommand)
	installed, err := InstallSemanticModule(context.Background(), config, SemanticModuleInstallInput{
		Provider:     SemanticModuleProviderOllama,
		ManifestPath: manifestPath,
		Command:      commandPath,
	})
	if err != nil {
		t.Fatalf("install semantic module: %v", err)
	}
	return installed
}

func writeRunclientOCRModuleManifest(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create OCR manifest dir: %v", err)
	}
	path := filepath.Join(dir, "module.json")
	manifest := map[string]any{
		"schema_version": "openclerk-module.v1",
		"module": map[string]any{
			"name":    "tesseract-ocr",
			"version": "0.1.0",
			"kind":    ModuleKindOCRProvider,
		},
		"provides": []map[string]any{{
			"type": "command",
			"name": "tesseract ocr",
		}, {
			"type": "command",
			"name": "ocrmypdf ocr",
		}},
		"requires": map[string]any{
			"tools": []string{"tesseract", "ocrmypdf"},
		},
		"authority": map[string]any{
			"default":        "read_only",
			"durable_writes": "forbidden",
			"forbidden":      []string{"write_documents", "hidden_cloud_egress"},
		},
		"release": map[string]any{
			"status": "supported_optional_module",
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal OCR manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write OCR manifest: %v", err)
	}
	return path
}

func installRunclientOCRModuleForPoisonTest(t *testing.T, config Config) SemanticModuleConfig {
	t.Helper()

	manifestPath := writeRunclientOCRModuleManifest(t, t.TempDir())
	tesseractPath := writeRunclientExecutable(t, ocrModuleTesseractCommand)
	ocrmypdfPath := writeRunclientExecutable(t, ocrModuleOCRmyPDFCommand)
	installed, err := InstallOCRModule(context.Background(), config, SemanticModuleInstallInput{
		Kind:         ModuleKindOCRProvider,
		Provider:     OCRModuleProviderTesseract,
		ManifestPath: manifestPath,
		Command:      tesseractPath,
		ProviderConfig: map[string]string{
			"ocrmypdf_command": ocrmypdfPath,
		},
	})
	if err != nil {
		t.Fatalf("install OCR module: %v", err)
	}
	return installed
}

func writeRunclientExecutable(t *testing.T, name string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '%s\\n' "+name+"\n"), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
	return path
}

func replaceManifestFile(t *testing.T, path string, old string, replacement string) {
	t.Helper()
	data := string(mustReadFile(t, path))
	if !strings.Contains(data, old) {
		t.Fatalf("manifest %s does not contain %q", path, old)
	}
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(data, old, replacement)), 0o600); err != nil {
		t.Fatalf("mutate manifest: %v", err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func writeLegacyModuleRuntimeConfig(t *testing.T, ctx context.Context, cfg Config, store moduleRuntimeConfigStore, config SemanticModuleConfig) {
	t.Helper()
	values := moduleRuntimeConfigValues(config)
	delete(values, "manifest_resolved_path")
	if err := withRuntimeConfigDB(ctx, cfg, true, func(db *sql.DB) error {
		for key, value := range values {
			if _, err := db.ExecContext(ctx, `
INSERT INTO runtime_config (key_name, value_text, updated_at)
VALUES (?, ?, 'legacy-test')
ON CONFLICT(key_name) DO UPDATE SET
	value_text = excluded.value_text,
	updated_at = excluded.updated_at`, store.key(config.Provider, key), value); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("write legacy module config: %v", err)
	}
}
