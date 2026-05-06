package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func TestVerifyInstallOrUpgradeInstructions(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	writeTestFile(t, repoRoot, ".agents/skills/openclerk/SKILL.md", "# OpenClerk\n")
	paths := evalPaths{DatabasePath: filepath.Join(repoRoot, ".openclerk-eval", "openclerk.db")}
	evalMetrics := metrics{
		OpenClerkPathCheckUsed:    true,
		OpenClerkVersionCheckUsed: true,
		OpenClerkSkillCheckUsed:   true,
		EventTypeCounts:           map[string]int{},
	}
	answer := "Install verified with command -v openclerk, openclerk --version, and skills/openclerk/SKILL.md registered."
	result := verifyInstallOrUpgradeInstructions(paths, answer, evalMetrics, false)
	if !result.Passed {
		t.Fatalf("install verification failed: %+v", result)
	}

	result = verifyInstallOrUpgradeInstructions(paths, answer, metrics{OpenClerkPathCheckUsed: true, EventTypeCounts: map[string]int{}}, false)
	if result.Passed {
		t.Fatalf("install verification passed without version check: %+v", result)
	}
}

func writeTestFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestVerifyModuleAgentInstall(t *testing.T) {
	_, dbPath := seedModuleVerificationFixture(t, moduleAgentInstallEmbeddingModel)
	evalMetrics := metrics{
		ModuleInstallUsed: true,
		ModuleListUsed:    true,
		EventTypeCounts:   map[string]int{},
	}
	answer := "Module-agent install verified for ollama using modules/ollama-embeddings/module.json and modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md. list_modules returned verified redacted state; no direct SQLite or provider semantic_search was used."
	result, err := verifyModuleAgentInstall(context.Background(), evalPaths{DatabasePath: dbPath}, answer, evalMetrics)
	if err != nil {
		t.Fatalf("verify module install: %v", err)
	}
	if !result.Passed {
		t.Fatalf("module verification failed: %+v", result)
	}

	result, err = verifyModuleAgentInstall(context.Background(), evalPaths{DatabasePath: dbPath}, answer, metrics{ModuleInstallUsed: true, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify module install missing list: %v", err)
	}
	if result.Passed {
		t.Fatalf("module verification passed without list_modules: %+v", result)
	}

	result, err = verifyModuleAgentInstall(context.Background(), evalPaths{DatabasePath: dbPath}, answer, metrics{ModuleInstallUsed: true, ModuleListUsed: true, SemanticSearchUsed: true, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify module install semantic search: %v", err)
	}
	if result.Passed {
		t.Fatalf("module verification passed with semantic_search: %+v", result)
	}
}

func TestVerifyModuleAgentUpgrade(t *testing.T) {
	_, dbPath := seedModuleVerificationFixture(t, moduleAgentUpgradeEmbeddingModel)
	evalMetrics := metrics{
		ModuleInstallUsed: true,
		ModuleListUsed:    true,
		EventTypeCounts:   map[string]int{},
	}
	answer := "Module-agent upgrade verified for ollama using modules/ollama-embeddings/module.json and modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md. list_modules preserved existing provider config nomic-embed-text with verified redacted state; no direct SQLite or provider semantic_search was used."
	result, err := verifyModuleAgentUpgrade(context.Background(), evalPaths{DatabasePath: dbPath}, answer, evalMetrics)
	if err != nil {
		t.Fatalf("verify module upgrade: %v", err)
	}
	if !result.Passed {
		t.Fatalf("module upgrade verification failed: %+v", result)
	}

	result, err = verifyModuleAgentUpgrade(context.Background(), evalPaths{DatabasePath: dbPath}, answer, metrics{ModuleInstallUsed: true, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify module upgrade missing list: %v", err)
	}
	if result.Passed {
		t.Fatalf("module upgrade verification passed without list_modules: %+v", result)
	}
}

func seedModuleVerificationFixture(t *testing.T, embeddingModel string) (string, string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	sourceRepoRoot, err := filepath.Abs(filepath.Join(wd, "..", "..", ".."))
	if err != nil {
		t.Fatalf("source repo root: %v", err)
	}
	sourceManifest, err := os.ReadFile(filepath.Join(sourceRepoRoot, filepath.FromSlash(moduleAgentInstallManifestPath)))
	if err != nil {
		t.Fatalf("read source module manifest: %v", err)
	}
	repoRoot := t.TempDir()
	writeTestFile(t, repoRoot, moduleAgentInstallManifestPath, string(sourceManifest))

	dbPath := filepath.Join(repoRoot, ".openclerk-eval", "openclerk.db")
	_, err = runclient.InstallSemanticModule(context.Background(), runclient.Config{
		DatabasePath:       dbPath,
		ModuleManifestRoot: repoRoot,
	}, runclient.SemanticModuleInstallInput{
		Provider:     moduleAgentInstallProvider,
		ManifestPath: moduleAgentInstallManifestPath,
		Command:      moduleAgentInstallCommand,
		ProviderConfig: map[string]string{
			"embedding_model": embeddingModel,
			"ollama_url":      moduleAgentInstallOllamaURL,
		},
	})
	if err != nil {
		t.Fatalf("install module fixture: %v", err)
	}
	return repoRoot, dbPath
}
