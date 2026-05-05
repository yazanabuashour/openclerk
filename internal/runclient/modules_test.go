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

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
