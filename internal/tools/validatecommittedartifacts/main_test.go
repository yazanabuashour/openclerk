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
