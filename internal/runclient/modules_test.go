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

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
