package main

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeedRepoDocsDogfoodImportsPublicMarkdownOnly(t *testing.T) {
	ctx := context.Background()
	repoDir := t.TempDir()
	for path, body := range map[string]string{
		"README.md":                          "# OpenClerk\n\nPublic project overview.\n",
		"docs/architecture/public-adr.md":    "# Public ADR\n\nPublic architecture note.\n",
		"docs/evals/results/private-run.md":  "# Result\n\nSkipped reduced report fixture.\n",
		"AGENTS.md":                          "# Instructions\n\nSkipped agent instructions.\n",
		".openclerk-eval/generated-note.md":  "# Generated\n\nSkipped eval storage.\n",
		"skills/openclerk/SKILL.md":          "# OpenClerk Skill\n\nPublic skill markdown.\n",
		"docs/architecture/not-markdown.txt": "not markdown\n",
	} {
		target := filepath.Join(repoDir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	paths := scenarioPaths(repoDir)
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := seedRepoDocsDogfood(ctx, cfg); err != nil {
		t.Fatalf("seed repo docs: %v", err)
	}
	for _, path := range []string{"README.md", "docs/architecture/public-adr.md", "skills/openclerk/SKILL.md"} {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("lookup %s: %v", path, err)
		} else if !found {
			t.Fatalf("expected imported repo doc %s", path)
		}
	}
	for _, path := range []string{"docs/evals/results/private-run.md", "AGENTS.md", ".openclerk-eval/generated-note.md", "docs/architecture/not-markdown.txt"} {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("lookup skipped %s: %v", path, err)
		} else if found {
			t.Fatalf("unexpected imported repo doc %s", path)
		}
	}
}

func seedMinimalRepoDocsDogfood(t *testing.T, ctx context.Context, cfg runclient.Config) {
	t.Helper()
	agentOps := strings.TrimSpace(`---
decision_id: adr-agentops-only-knowledge-plane
decision_title: AgentOps-Only Knowledge Plane Direction
decision_status: accepted
decision_scope: knowledge-plane
decision_owner: platform
---
# ADR: AgentOps-Only Knowledge Plane Direction

## Decision
oc-rsj verified current AgentOps document retrieval runner actions keep the installed openclerk runner as the production agent surface.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, repoDocsAgentOpsADRPath, "AgentOps-Only Knowledge Plane Direction", agentOps); err != nil {
		t.Fatalf("seed AgentOps repo doc: %v", err)
	}
	knowledgeConfig := strings.TrimSpace(`---
decision_id: adr-knowledge-configuration-v1
decision_title: Knowledge Configuration v1
decision_status: accepted
decision_scope: knowledge-configuration
decision_owner: platform
---
# ADR: Knowledge Configuration v1

## Decision
Knowledge Configuration v1 accepted AgentOps surface keeps canonical markdown docs authoritative and exposes derived records through runner JSON.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, repoDocsKnowledgeConfigPath, "Knowledge Configuration v1", knowledgeConfig); err != nil {
		t.Fatalf("seed knowledge configuration repo doc: %v", err)
	}
}

func TestSourceURLUpdateFixturePromptRendering(t *testing.T) {
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateChangedScenarioID)
	if fixtures == nil {
		t.Fatal("fixture server not started")
	}
	defer fixtures.Close()

	rendered := fixtures.renderPrompt(sourceURLUpdateStableURLToken + " " + sourceURLUpdateChangedURLToken + " " + artifactPDFSourceURLToken)
	if strings.Contains(rendered, sourceURLUpdateStableURLToken) || strings.Contains(rendered, sourceURLUpdateChangedURLToken) || strings.Contains(rendered, artifactPDFSourceURLToken) {
		t.Fatalf("prompt still contains fixture token: %s", rendered)
	}
	if !strings.Contains(rendered, fixtures.stableURL()) || !strings.Contains(rendered, fixtures.changedURL()) {
		t.Fatalf("prompt missing fixture URLs: %s", rendered)
	}
}

func TestArtifactPDFFixturePromptRenderingUsesEvalURL(t *testing.T) {
	fixtures := startSourceURLUpdateFixtures(artifactPDFSourceURLScenarioID)
	if fixtures == nil {
		t.Fatal("artifact PDF fixture not started")
	}
	defer fixtures.Close()

	runDir := t.TempDir()
	if err := fixtures.prepareFiles(runDir); err != nil {
		t.Fatalf("prepare artifact PDF fixture: %v", err)
	}
	rendered := fixtures.renderPrompt(artifactPDFSourceURLToken)
	if rendered != artifactPDFEvalSourceURL {
		t.Fatalf("artifact PDF URL = %q, want %q", rendered, artifactPDFEvalSourceURL)
	}
	if strings.Contains(rendered, "127.0.0.1") || strings.Contains(rendered, "localhost") {
		t.Fatalf("artifact PDF URL still uses loopback: %s", rendered)
	}
	if _, err := os.Stat(filepath.Join(evalSourceFixtureRoot(runDir), "artifacts", "vendor-security-paper.pdf")); err != nil {
		t.Fatalf("artifact PDF fixture file stat: %v", err)
	}
}

func TestSeedRAGRetrievalBaselineCreatesFilteredFixture(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	for _, path := range []string{ragCurrentPolicyPath, ragDecoyPolicyPath, ragArchivedPolicyPath} {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("find %s: %v", path, err)
		} else if !found {
			t.Fatalf("missing seeded document %s", path)
		}
	}
	body, found, err := documentBodyByPath(ctx, paths, ragCurrentPolicyPath)
	if err != nil {
		t.Fatalf("get current policy: %v", err)
	}
	if !found || !strings.Contains(body, "rag_scope: active-policy") || !strings.Contains(body, ragCurrentPolicyDecision) {
		t.Fatalf("current policy body = %q", body)
	}
	count, err := documentCountWithPrefix(ctx, paths, ragPathPrefix)
	if err != nil {
		t.Fatalf("count RAG docs: %v", err)
	}
	if count != 2 {
		t.Fatalf("RAG path count = %d, want 2", count)
	}
}

func TestSeedPopulatedVaultFixtureCreatesMixedDocumentFamilies(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}); err != nil {
		t.Fatalf("seed populated scenario: %v", err)
	}
	for _, path := range populatedVaultFixturePaths() {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("find %s: %v", path, err)
		} else if !found {
			t.Fatalf("missing seeded populated document %s", path)
		}
	}
	for prefix, minimum := range populatedVaultFixtureMinimumPrefixCounts() {
		count, err := documentCountWithPrefix(ctx, paths, prefix)
		if err != nil {
			t.Fatalf("count %s: %v", prefix, err)
		}
		if count < minimum {
			t.Fatalf("expected at least %d populated fixture docs under %s, got %d", minimum, prefix, count)
		}
	}
}

func replaceSeedSection(t *testing.T, ctx context.Context, paths evalPaths, docPath string, heading string, content string) {
	t.Helper()
	docID, found, err := documentIDByPath(ctx, paths, docPath)
	if err != nil {
		t.Fatalf("find %s: %v", docPath, err)
	}
	if !found {
		t.Fatalf("missing %s", docPath)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   docID,
		Heading: heading,
		Content: content,
	})
	if err != nil {
		t.Fatalf("replace %s section %s: %v", docPath, heading, err)
	}
	if result.Rejected {
		t.Fatalf("replace %s rejected: %s", docPath, result.RejectionReason)
	}
}
