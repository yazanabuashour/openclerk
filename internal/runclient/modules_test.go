package runclient

import (
	"context"
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
	installed, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
		Provider:     "gemini",
		ManifestPath: manifestPath,
		Command:      "semantic-retrieval-adapter",
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

func TestOCRModuleRuntimeConfigVerifiesManifest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	manifestPath := writeRunclientOCRModuleManifest(t, t.TempDir())
	installed, err := InstallOCRModule(ctx, config, SemanticModuleInstallInput{
		Kind:         ModuleKindOCRProvider,
		Provider:     OCRModuleProviderTesseract,
		ManifestPath: manifestPath,
		Command:      "tesseract",
		ProviderConfig: map[string]string{
			"ocrmypdf_command": "ocrmypdf",
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
	disabled := false

	if _, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
		Provider:     SemanticModuleProviderOllama,
		ManifestPath: semanticManifest,
		Command:      "semantic-retrieval-adapter",
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
		Command:      "tesseract",
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

func TestInstallSemanticModuleLocksCommandSurface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("rejects noncanonical command", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
		_, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderOllama,
			ManifestPath: manifestPath,
			Command:      "/bin/sh",
		})
		if err == nil || !strings.Contains(err.Error(), "module.command must be semantic-retrieval-adapter") {
			t.Fatalf("err = %v, want semantic command rejection", err)
		}
	})

	t.Run("rejects command args", func(t *testing.T) {
		t.Parallel()

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

	t.Run("persists canonical command when omitted", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "ollama")
		installed, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderOllama,
			ManifestPath: manifestPath,
		})
		if err != nil {
			t.Fatalf("install semantic module: %v", err)
		}
		if installed.Command != semanticModuleCommand || len(installed.CommandArgs) != 0 {
			t.Fatalf("installed command surface = %q %v, want canonical command without args", installed.Command, installed.CommandArgs)
		}
	})

	t.Run("persists canonical command when provided", func(t *testing.T) {
		t.Parallel()

		config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
		manifestPath := writeRunclientSemanticModuleManifest(t, t.TempDir(), "gemini")
		installed, err := InstallSemanticModule(ctx, config, SemanticModuleInstallInput{
			Provider:     SemanticModuleProviderGemini,
			ManifestPath: manifestPath,
			Command:      semanticModuleCommand,
		})
		if err != nil {
			t.Fatalf("install semantic module: %v", err)
		}
		if installed.Command != semanticModuleCommand || len(installed.CommandArgs) != 0 {
			t.Fatalf("installed command surface = %q %v, want canonical command without args", installed.Command, installed.CommandArgs)
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
		installed.CommandArgs = []string{"-c", "echo pwned"}
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
	installed, err := InstallSemanticModule(context.Background(), config, SemanticModuleInstallInput{
		Provider:     SemanticModuleProviderOllama,
		ManifestPath: manifestPath,
		Command:      semanticModuleCommand,
	})
	if err != nil {
		t.Fatalf("install semantic module: %v", err)
	}
	return installed
}

func writeRunclientOCRModuleManifest(t *testing.T, dir string) string {
	t.Helper()
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
		}},
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
