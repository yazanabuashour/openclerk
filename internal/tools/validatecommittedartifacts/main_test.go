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

	valid := "# OpenClerk\n\n## Modules\n\n### Agent Module Instructions\n\nInstall prompt:\n\n<module-provider> <module-manifest-path> <module-command> <module-skill-path>\n\nUpgrade prompt:\n\n<module-name> <module-version-or-latest> <module-provider>\n\nAvailable installable modules:\n\nExact module commands live in `modules/docs/install.md`.\n"
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

	theater := strings.Replace(valid, "Install prompt:", "Tell your agent:\n\nInstall prompt:", 1)
	err = validateReadmeModuleSection(theater)
	if err == nil || !strings.Contains(err.Error(), "Tell your agent") {
		t.Fatalf("validateReadmeModuleSection theater error = %v, want tell your agent rejection", err)
	}
}

func TestValidateModuleInstallDocRequiresInstallUpgradeSections(t *testing.T) {
	t.Parallel()

	valid := "# OpenClerk Module Install\n\n## Install a Module Release\n\nscripts/install-module.sh\n\n## Upgrade a Module Release\n\n## Register or Refresh Module Registration\n\nscripts/build-module-release-bundle.sh\n"
	if err := validateModuleInstallDoc(valid); err != nil {
		t.Fatalf("validateModuleInstallDoc valid: %v", err)
	}
	err := validateModuleInstallDoc(strings.Replace(valid, "## Upgrade a Module Release\n", "", 1))
	if err == nil || !strings.Contains(err.Error(), "Upgrade") {
		t.Fatalf("validateModuleInstallDoc missing upgrade error = %v, want upgrade rejection", err)
	}
}

func TestValidateModuleDocumentationReferencesEmbeddingModules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	readme := `# OpenClerk

## Modules

### Agent Module Instructions

Install prompt:

<module-provider> <module-manifest-path> <module-command> <module-skill-path>

Upgrade prompt:

<module-name> <module-version-or-latest> <module-provider>

Available installable modules:

| Module | Skill |
| --- | --- |
| ` + "`modules/example-embeddings/module.json`" + ` | ` + "`modules/example-embeddings/skill/example/SKILL.md`" + ` |

Exact module commands live in ` + "`modules/docs/install.md`" + `.
`
	moduleDoc := `# OpenClerk Module Install

## Install a Module Release

scripts/install-module.sh

## Upgrade a Module Release

## Register or Refresh Module Registration

scripts/build-module-release-bundle.sh

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

	writeTestFile(t, root, "modules/docs/install.md", "# OpenClerk Module Install\n\n## Install a Module Release\n\nscripts/install-module.sh\n\n## Upgrade a Module Release\n\n## Register or Refresh Module Registration\n\nscripts/build-module-release-bundle.sh\n")
	err := validateModuleDocumentation(root, files)
	if err == nil || !strings.Contains(err.Error(), "modules/docs/install.md must reference module manifest") {
		t.Fatalf("validateModuleDocumentation missing module error = %v, want manifest reference rejection", err)
	}
}

func TestValidateLiveInstallSmokeReport(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	report := `{
  "schema_version": "openclerk-live-install-smoke.v1",
  "install": {
    "passed": true,
    "installer_invocation": "sh -c \"$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)\"",
    "binary_path": "$HOME/.local/bin/openclerk",
    "command_path": "$HOME/.local/bin/openclerk",
    "version_output": "openclerk v0.0.0-smoke",
    "help_checked": true
  },
  "upgrade": {
    "passed": true,
    "binary_path": "$HOME/.local/bin/openclerk",
    "command_path": "$HOME/.local/bin/openclerk",
    "version_output": "openclerk v0.0.0-smoke",
    "help_checked": true
  },
  "skill": {
    "passed": true,
    "skill_path": "$CODEX_HOME/skills/openclerk/SKILL.md",
    "source": "skills/openclerk/SKILL.md"
  },
  "module": {
    "passed": true,
    "provider": "ollama",
    "manifest_path": "modules/ollama-embeddings/module.json",
    "skill_path": "modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md",
    "install_passed": true,
    "configure_passed": true,
    "list_passed": true,
    "remove_passed": true,
    "final_list_empty": true,
    "provider_config": {
      "embedding_model": "nomic-embed-text",
      "ollama_url": "http://localhost:11434"
    },
    "verification_state": "verified",
    "redaction_state": "redacted"
  },
  "validation_boundaries": "local temp HOME/CODEX_HOME only; no durable host install; no network release fetch; no direct SQLite edit"
}`
	writeTestFile(t, root, "docs/evals/results/ockp-live-install-upgrade-module-smoke.json", report)
	if err := validateLiveInstallSmokeReport(root); err != nil {
		t.Fatalf("validateLiveInstallSmokeReport: %v", err)
	}

	writeTestFile(t, root, "docs/evals/results/ockp-live-install-upgrade-module-smoke.json", strings.Replace(report, `"remove_passed": true`, `"remove_passed": false`, 1))
	err := validateLiveInstallSmokeReport(root)
	if err == nil || !strings.Contains(err.Error(), "install/config/list/remove") {
		t.Fatalf("validateLiveInstallSmokeReport missing remove error = %v, want install/config/list/remove rejection", err)
	}
}
