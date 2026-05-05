package main

import (
	"strings"
	"testing"
)

func TestValidatePublicArtifactTextRejectsPrivateAndMachinePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "machine path",
			content: "stored at /Users/example/notes/private.md\n",
			wantErr: "machine-absolute path",
		},
		{
			name:    "private notes path",
			content: "use ~/notes/research/private.md\n",
			wantErr: "private notes path",
		},
		{
			name:    "private research note",
			content: "see agent-first-knowledge-plane-architecture-2026-04-07.md\n",
			wantErr: "private research note reference",
		},
		{
			name:    "raw log without placeholder",
			content: "raw log: /tmp/run/production/create/turn-1/events.jsonl\n",
			wantErr: "without <run-root>",
		},
		{
			name:    "raw logs committed",
			content: `"raw_logs_committed": true`,
			wantErr: "raw eval logs",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validatePublicArtifactText("docs/example.md", tt.content)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validatePublicArtifactText error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePublicArtifactTextAllowsSanitizedPlaceholders(t *testing.T) {
	t.Parallel()

	content := "Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`\n" +
		"The private report uses `<private-vault>` and `<run-root>` placeholders.\n" +
		`"raw_logs_committed": false`
	if err := validatePublicArtifactText("docs/evals/results/example.md", content); err != nil {
		t.Fatalf("validatePublicArtifactText: %v", err)
	}
}

func TestValidateRealVaultReportPolicy(t *testing.T) {
	t.Parallel()

	valid := `# OpenClerk Real-Vault AgentOps Trial

This is a sanitized real-vault trial report.

Raw logs remain local-only.

- The report omits private paths, titles, snippets, citations, and document identifiers.
- The report omits validation document paths and raw JSON.
`
	root := t.TempDir()
	writeTestFile(t, root, "docs/evals/results/ockp-real-vault-agentops-trial.md", valid)
	if err := validateRealVaultReport(root); err != nil {
		t.Fatalf("validateRealVaultReport: %v", err)
	}

	leaky := valid + "Private path: sources/private/topic.md with doc_private and chunk_private.\n"
	root = t.TempDir()
	writeTestFile(t, root, "docs/evals/results/ockp-real-vault-agentops-trial.md", leaky)
	err := validateRealVaultReport(root)
	if err == nil || !strings.Contains(err.Error(), "vault-relative markdown path") {
		t.Fatalf("validateRealVaultReport error = %v, want vault path rejection", err)
	}
}

func TestValidateNoRealVaultJSONReports(t *testing.T) {
	t.Parallel()

	err := validateNoRealVaultJSONReports([]string{
		"docs/evals/results/ockp-real-vault-agentops-trial.md",
		"docs/evals/results/ockp-real-vault-agentops-trial.json",
	})
	if err == nil || !strings.Contains(err.Error(), "local-only") {
		t.Fatalf("validateNoRealVaultJSONReports error = %v, want local-only rejection", err)
	}
}

func TestValidateReadmeModuleSectionRequiresAgentInstructionsFirst(t *testing.T) {
	t.Parallel()

	valid := "# OpenClerk\n\n## Modules\n\n### Agent Module Instructions\n\nTell an agent.\n\nAvailable installable modules:\n\nExact module commands live in `modules/docs/install.md`.\n"
	if err := validateReadmeModuleSection(valid); err != nil {
		t.Fatalf("validateReadmeModuleSection valid: %v", err)
	}

	lateAgent := "# OpenClerk\n\n## Modules\n\nAvailable installable modules:\n\n### Agent Module Instructions\n\nExact module commands live in `modules/docs/install.md`.\n"
	err := validateReadmeModuleSection(lateAgent)
	if err == nil || !strings.Contains(err.Error(), "beginning") {
		t.Fatalf("validateReadmeModuleSection late agent error = %v, want beginning rejection", err)
	}

	inlineCommand := valid + "Install Ollama embeddings:\n"
	err = validateReadmeModuleSection(inlineCommand)
	if err == nil || !strings.Contains(err.Error(), "must not inline") {
		t.Fatalf("validateReadmeModuleSection inline command error = %v, want inline rejection", err)
	}
}

func TestValidateModuleDocumentationReferencesEmbeddingModules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	readme := `# OpenClerk

## Modules

### Agent Module Instructions

Tell an agent.

Available installable modules:

| Module | Skill |
| --- | --- |
| ` + "`modules/example-embeddings/module.json`" + ` | ` + "`modules/example-embeddings/skill/example/SKILL.md`" + ` |

Exact module commands live in ` + "`modules/docs/install.md`" + `.
`
	moduleDoc := `# OpenClerk Module Install

| Module | Skill |
| --- | --- |
| ` + "`modules/example-embeddings/module.json`" + ` | ` + "`modules/example-embeddings/skill/example/SKILL.md`" + ` |
`
	manifest := `{
  "module": {"kind": "embedding_provider"},
  "provides": [{"type": "skill", "path": "modules/example-embeddings/skill/example/SKILL.md"}]
}`
	writeTestFile(t, root, "README.md", readme)
	writeTestFile(t, root, "modules/docs/install.md", moduleDoc)
	writeTestFile(t, root, "modules/example-embeddings/module.json", manifest)
	files := []string{
		"README.md",
		"modules/docs/install.md",
		"modules/example-embeddings/module.json",
	}
	if err := validateModuleDocumentation(root, files); err != nil {
		t.Fatalf("validateModuleDocumentation: %v", err)
	}

	writeTestFile(t, root, "modules/docs/install.md", "# OpenClerk Module Install\n")
	err := validateModuleDocumentation(root, files)
	if err == nil || !strings.Contains(err.Error(), "modules/docs/install.md must reference module manifest") {
		t.Fatalf("validateModuleDocumentation missing module error = %v, want manifest reference rejection", err)
	}
}
