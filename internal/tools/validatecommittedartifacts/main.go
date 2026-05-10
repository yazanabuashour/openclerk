package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	machineUnixPathPattern    = regexp.MustCompile(`/(Users|home)/[^/\s"'\\)]+`)
	machineWindowsPathPattern = regexp.MustCompile(`(?i)\b[A-Z]:\\Users\\[^\\\s"']+`)
	privateNotesPathPattern   = regexp.MustCompile(`(^|[\s"'(])~?/notes(/|\b)`)
	realVaultDocPathPattern   = regexp.MustCompile(`\b(sources|synthesis|notes|transcripts|articles|meetings|blogs|receipts|invoices|legal|contracts)/[A-Za-z0-9._/-]+\.md\b`)
	realVaultDocIDPattern     = regexp.MustCompile(`\bdoc_[A-Za-z0-9][A-Za-z0-9_-]*\b`)
	realVaultChunkIDPattern   = regexp.MustCompile(`\bchunk_[A-Za-z0-9][A-Za-z0-9_-]*\b`)
	slugTokenPattern          = regexp.MustCompile(`\b[a-z0-9][a-z0-9-]{20,}\b`)
)

var privateResearchNameHashes = map[string]struct{}{
	"de7b169e2dab278337cf5d127e06e6430caaa3673d52252f28b0e39c61764f8e": {},
	"6d8e52e24eda4d45426f8a6ea87d3dbafd97f9fb096aebc6222c936af6da4395": {},
	"22362e88567b0f8783335d08d5548f56c52c86c2e8f5fd4becb9f2bbefbf513c": {},
	"d33e740cde66125c2fba0ca7363cacbbec6dfce65714295f3bf24b001f43d69e": {},
	"866fffc2a236eab6dd849b4b820853f8c883899c2d1153e402d0af76af9cae97": {},
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 0 {
		return errors.New("usage: scripts/validate-committed-artifacts.sh")
	}
	if err := validateCommittedArtifacts("."); err != nil {
		return err
	}
	_, err := fmt.Fprintln(stdout, "validated committed artifacts")
	return err
}

func validateCommittedArtifacts(root string) error {
	files, err := trackedFiles(root)
	if err != nil {
		return err
	}
	if err := validateArtifactFiles(root, files); err != nil {
		return err
	}
	if err := validateRealVaultReport(root); err != nil {
		return err
	}
	if err := validateRealVaultRoutineUXReport(root); err != nil {
		return err
	}
	if err := validateNoRealVaultJSONReports(files); err != nil {
		return err
	}
	if err := validatePublicVaultReports(root); err != nil {
		return err
	}
	if err := validateModuleDocumentation(root, files); err != nil {
		return err
	}
	if err := validateLiveInstallSmokeReport(root); err != nil {
		return err
	}
	return nil
}

func trackedFiles(root string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", "-z")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list tracked files: %w", err)
	}
	parts := bytes.Split(output, []byte{0})
	files := []string{}
	for _, part := range parts {
		name := strings.TrimSpace(string(part))
		if name != "" {
			files = append(files, filepath.ToSlash(name))
		}
	}
	return files, nil
}

func validateArtifactFiles(root string, files []string) error {
	for _, rel := range files {
		if !isPublicArtifactPath(rel) {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		if err := validatePublicArtifactText(rel, strings.ReplaceAll(string(content), "\r\n", "\n")); err != nil {
			return err
		}
	}
	return nil
}

func isPublicArtifactPath(rel string) bool {
	rel = filepath.ToSlash(rel)
	ext := filepath.Ext(rel)
	switch ext {
	case ".md", ".json", ".yml", ".yaml", ".toml", ".txt":
	default:
		return false
	}
	if strings.HasPrefix(rel, "docs/") ||
		strings.HasPrefix(rel, "modules/docs/") ||
		strings.HasPrefix(rel, "skills/") ||
		strings.HasPrefix(rel, ".github/") {
		return true
	}
	switch rel {
	case "AGENTS.md", "README.md", "CHANGELOG.md", "CONTRIBUTING.md", "SECURITY.md", "CODE_OF_CONDUCT.md":
		return true
	default:
		return false
	}
}

func validatePublicArtifactText(rel string, text string) error {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lineNumber := i + 1
		if match := machineUnixPathPattern.FindString(line); match != "" {
			return fmt.Errorf("%s:%d contains machine-absolute path %q", rel, lineNumber, match)
		}
		if match := machineWindowsPathPattern.FindString(line); match != "" {
			return fmt.Errorf("%s:%d contains machine-absolute path %q", rel, lineNumber, match)
		}
		if match := privateNotesPathPattern.FindString(line); match != "" {
			return fmt.Errorf("%s:%d contains private notes path %q", rel, lineNumber, strings.TrimSpace(match))
		}
		if containsPrivateResearchName(line) {
			return fmt.Errorf("%s:%d contains private research note reference", rel, lineNumber)
		}
		if strings.Contains(line, "events.jsonl") && !containsRunRootPlaceholder(line) {
			return fmt.Errorf("%s:%d references raw eval logs without <run-root> placeholder", rel, lineNumber)
		}
		if strings.Contains(line, `"raw_logs_committed": true`) {
			return fmt.Errorf("%s:%d marks raw eval logs as committed", rel, lineNumber)
		}
	}
	return nil
}

func containsPrivateResearchName(line string) bool {
	for _, token := range slugTokenPattern.FindAllString(strings.ToLower(line), -1) {
		sum := sha256.Sum256([]byte(token))
		if _, ok := privateResearchNameHashes[hex.EncodeToString(sum[:])]; ok {
			return true
		}
	}
	return false
}

func validateModuleDocumentation(root string, files []string) error {
	read := func(rel string) (string, error) {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("%s not found", rel)
			}
			return "", fmt.Errorf("read %s: %w", rel, err)
		}
		return strings.ReplaceAll(string(content), "\r\n", "\n"), nil
	}
	readme, err := read("README.md")
	if err != nil {
		return err
	}
	if err := validateReadmeModuleSection(readme); err != nil {
		return err
	}
	if err := validateReadmeAgentPrompts(readme); err != nil {
		return err
	}
	moduleInstallDoc, err := read("modules/docs/install.md")
	if err != nil {
		return err
	}
	if err := validateModuleInstallDoc(moduleInstallDoc); err != nil {
		return err
	}
	modules, err := documentedModules(root, files)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		return errors.New("module docs validation found no module manifests")
	}
	for _, module := range modules {
		for _, target := range []struct {
			name string
			text string
		}{
			{name: "README.md", text: readme},
			{name: "modules/docs/install.md", text: moduleInstallDoc},
		} {
			if !strings.Contains(target.text, module.ManifestPath) {
				return fmt.Errorf("%s must reference module manifest %s", target.name, module.ManifestPath)
			}
			if !strings.Contains(target.text, module.SkillPath) {
				return fmt.Errorf("%s must reference module skill %s", target.name, module.SkillPath)
			}
		}
	}
	return nil
}

func validateReadmeModuleSection(readme string) error {
	modulesIndex := strings.Index(readme, "\n## Modules\n")
	agentIndex := strings.Index(readme, "\n### Agent Module Instructions\n")
	availableIndex := strings.Index(readme, "\nAvailable installable modules:\n")
	linkIndex := strings.Index(readme, "`modules/docs/install.md`")
	if modulesIndex < 0 {
		return errors.New("README.md must include ## Modules")
	}
	if agentIndex < 0 {
		return errors.New("README.md module section must include ### Agent Module Instructions")
	}
	if availableIndex < 0 {
		return errors.New("README.md module section must include available installable modules")
	}
	if modulesIndex >= agentIndex || agentIndex >= availableIndex {
		return errors.New("README.md must put Agent Module Instructions at the beginning of the Modules section")
	}
	if linkIndex < 0 {
		return errors.New("README.md module section must link to modules/docs/install.md")
	}
	moduleSection := readme[modulesIndex:]
	required := []string{
		"Install prompt:",
		"Upgrade prompt:",
		"<module-provider>",
		"<module-manifest-path>",
		"<module-command>",
		"<module-skill-path>",
		"<module-name>",
		"<module-version-or-latest>",
	}
	for _, want := range required {
		if !strings.Contains(moduleSection, want) {
			return fmt.Errorf("README.md module section missing concise agent module instruction placeholder %q", want)
		}
	}
	for _, forbidden := range []string{
		"Tell your agent:",
		"Module commands are available from the current source checkout",
		"Released runners through",
		"mise exec -- go build -o \"$HOME/.local/bin/openclerk\" ./cmd/openclerk",
		"Install Ollama embeddings:",
		"Install Gemini embeddings:",
		"Configure a module:",
		"Remove a module:",
		`{"action":"install_module"`,
	} {
		if strings.Contains(moduleSection, forbidden) {
			return fmt.Errorf("README.md module section must not inline module implementation detail %q", forbidden)
		}
	}
	return nil
}

func validateModuleInstallDoc(text string) error {
	for _, want := range []string{
		"## Install a Module Release",
		"## Upgrade a Module Release",
		"## Register or Refresh Module Registration",
		"scripts/build-module-release-bundle.sh",
		"scripts/install-module.sh",
	} {
		if !strings.Contains(text, want) {
			return fmt.Errorf("modules/docs/install.md missing module install/upgrade guidance %q", want)
		}
	}
	return nil
}

func validateReadmeAgentPrompts(readme string) error {
	moduleSectionIndex := strings.Index(readme, "\n### Agent Module Instructions\n")
	if moduleSectionIndex < 0 {
		return errors.New("README.md must include Agent Module Instructions before validating prompts")
	}
	moduleSection := readme[moduleSectionIndex:]
	prompts := []struct {
		name     string
		text     string
		maxRunes int
		required []string
	}{
		{
			name:     "core install prompt",
			text:     fencedTextAfter(readme, "**Or tell your agent:**", 0),
			maxRunes: 360,
			required: []string{
				"https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh",
				"$HOME/.local/bin",
				"skills/openclerk/SKILL.md",
				"command -v openclerk",
				"openclerk --version",
				"skill path",
				"runner and skill",
			},
		},
		{
			name:     "core upgrade prompt",
			text:     fencedTextAfter(readme, "**Upgrade prompt:**", strings.Index(readme, "**Or tell your agent:**")),
			maxRunes: 340,
			required: []string{
				"Upgrade OpenClerk",
				"https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh",
				"Re-register",
				"skills/openclerk/SKILL.md",
				"command -v openclerk",
				"openclerk --version",
				"skill path",
				"runner and skill",
			},
		},
		{
			name:     "module install prompt",
			text:     fencedTextAfter(moduleSection, "Install prompt:", 0),
			maxRunes: 260,
			required: []string{
				"<module-provider>",
				"<module-manifest-path>",
				"<module-command>",
				"<module-skill-path>",
				"openclerk module",
				"list_modules",
				"command_args",
				"SQLite",
			},
		},
		{
			name:     "module upgrade prompt",
			text:     fencedTextAfter(moduleSection, "Upgrade prompt:", 0),
			maxRunes: 240,
			required: []string{
				"<module-name>",
				"<module-version-or-latest>",
				"openclerk module",
				"preserve existing provider config",
				"list_modules",
				"SQLite",
			},
		},
	}
	for _, prompt := range prompts {
		if strings.TrimSpace(prompt.text) == "" {
			return fmt.Errorf("README.md missing %s", prompt.name)
		}
		if got := utf8.RuneCountInString(prompt.text); got > prompt.maxRunes {
			return fmt.Errorf("README.md %s length = %d, want <= %d", prompt.name, got, prompt.maxRunes)
		}
		for _, want := range prompt.required {
			if !strings.Contains(prompt.text, want) {
				return fmt.Errorf("README.md %s missing %q", prompt.name, want)
			}
		}
		for _, forbidden := range []string{
			"go build",
			"go run",
			"source checkout",
			"source-built",
			"cmd/openclerk",
			"Install Ollama embeddings:",
			"Install Gemini embeddings:",
			"Configure a module:",
			`{"action"`,
		} {
			if strings.Contains(prompt.text, forbidden) {
				return fmt.Errorf("README.md %s must not include recipe/source-build detail %q", prompt.name, forbidden)
			}
		}
	}
	return nil
}

func fencedTextAfter(text string, marker string, start int) string {
	if start < 0 || start >= len(text) {
		return ""
	}
	markerIndex := strings.Index(text[start:], marker)
	if markerIndex < 0 {
		return ""
	}
	afterMarker := text[start+markerIndex+len(marker):]
	fenceStart := strings.Index(afterMarker, "```text")
	if fenceStart < 0 {
		return ""
	}
	afterFence := afterMarker[fenceStart+len("```text"):]
	afterFence = strings.TrimPrefix(afterFence, "\r\n")
	afterFence = strings.TrimPrefix(afterFence, "\n")
	fenceEnd := strings.Index(afterFence, "```")
	if fenceEnd < 0 {
		return ""
	}
	return strings.TrimSpace(afterFence[:fenceEnd])
}

type documentedModule struct {
	ManifestPath string
	SkillPath    string
}

func documentedModules(root string, files []string) ([]documentedModule, error) {
	out := []documentedModule{}
	for _, rel := range files {
		rel = filepath.ToSlash(rel)
		if !strings.HasPrefix(rel, "modules/") || !strings.HasSuffix(rel, "/module.json") {
			continue
		}
		module, err := readDocumentedModule(root, rel)
		if err != nil {
			return nil, err
		}
		if module.ManifestPath != "" {
			out = append(out, module)
		}
	}
	return out, nil
}

func readDocumentedModule(root string, rel string) (documentedModule, error) {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return documentedModule{}, fmt.Errorf("read %s: %w", rel, err)
	}
	var manifest struct {
		Module struct {
			Kind string `json:"kind"`
		} `json:"module"`
		Provides []struct {
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"provides"`
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		return documentedModule{}, fmt.Errorf("decode %s: %w", rel, err)
	}
	for _, provided := range manifest.Provides {
		if provided.Type == "skill" && strings.TrimSpace(provided.Path) != "" {
			return documentedModule{ManifestPath: rel, SkillPath: filepath.ToSlash(provided.Path)}, nil
		}
	}
	if manifest.Module.Kind == "retrieval_adapter" {
		return documentedModule{}, nil
	}
	if strings.TrimSpace(manifest.Module.Kind) != "" {
		return documentedModule{}, fmt.Errorf("%s module kind %q must provide a skill path", rel, manifest.Module.Kind)
	}
	return documentedModule{}, nil
}

func containsRunRootPlaceholder(line string) bool {
	return strings.Contains(line, "<run-root>/") ||
		strings.Contains(line, `\u003crun-root\u003e/`)
}

func validateRealVaultReport(root string) error {
	rel := filepath.ToSlash(filepath.Join("docs", "evals", "results", "ockp-real-vault-agentops-trial.md"))
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", rel, err)
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	required := []string{
		"sanitized real-vault trial report",
		"omits private paths, titles, snippets, citations, and document identifiers",
		"omits validation document paths and raw JSON",
		"Raw logs",
	}
	normalized := strings.Join(strings.Fields(text), " ")
	for _, want := range required {
		if !strings.Contains(normalized, want) {
			return fmt.Errorf("%s must document sanitized real-vault evidence policy: missing %q", rel, want)
		}
	}
	forbiddenPatterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{name: "vault-relative markdown path", pattern: realVaultDocPathPattern},
		{name: "document id", pattern: realVaultDocIDPattern},
		{name: "chunk id", pattern: realVaultChunkIDPattern},
		{name: "machine-absolute Unix path", pattern: machineUnixPathPattern},
		{name: "machine-absolute Windows path", pattern: machineWindowsPathPattern},
	}
	for _, forbidden := range forbiddenPatterns {
		if match := forbidden.pattern.FindString(text); match != "" {
			return fmt.Errorf("%s contains private real-vault %s %q", rel, forbidden.name, match)
		}
	}
	if strings.Contains(text, "events.jsonl") {
		return fmt.Errorf("%s must not reference private real-vault raw log files", rel)
	}
	return nil
}

func validateRealVaultRoutineUXReport(root string) error {
	rel := filepath.ToSlash(filepath.Join("docs", "evals", "results", "ockp-real-vault-routine-ux.md"))
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", rel, err)
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	required := []string{
		"sanitized real-vault routine UX telemetry report",
		"`<private-vault>`",
		"`<run-root>`",
		"Raw JSON committed: `false`",
		"Private task manifest committed: `false`",
		"omits private prompts, paths, titles, snippets, citations, document ids, chunk ids, raw JSON, event logs",
		"live private vault is never the mutation target",
	}
	normalized := strings.Join(strings.Fields(text), " ")
	for _, want := range required {
		if !strings.Contains(normalized, want) {
			return fmt.Errorf("%s must document sanitized routine UX evidence policy: missing %q", rel, want)
		}
	}
	forbiddenPatterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{name: "vault-relative markdown path", pattern: realVaultDocPathPattern},
		{name: "document id", pattern: realVaultDocIDPattern},
		{name: "chunk id", pattern: realVaultChunkIDPattern},
		{name: "machine-absolute Unix path", pattern: machineUnixPathPattern},
		{name: "machine-absolute Windows path", pattern: machineWindowsPathPattern},
	}
	for _, forbidden := range forbiddenPatterns {
		if match := forbidden.pattern.FindString(text); match != "" {
			return fmt.Errorf("%s contains private real-vault %s %q", rel, forbidden.name, match)
		}
	}
	if strings.Contains(text, "events.jsonl") && !strings.Contains(text, "<run-root>") {
		return fmt.Errorf("%s must not reference private real-vault raw log files without placeholder", rel)
	}
	return nil
}

func validateNoRealVaultJSONReports(files []string) error {
	for _, rel := range files {
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "docs/evals/results/ockp-real-vault") && strings.HasSuffix(rel, ".json") {
			return fmt.Errorf("%s must stay local-only; commit only the sanitized markdown real-vault report", rel)
		}
	}
	return nil
}

func validatePublicVaultReports(root string) error {
	reports := []struct {
		name         string
		repoURL      string
		repoRef      string
		vaultPrefix  string
		decisionText string
	}{
		{"ockp-public-vault-kubernetes-docs", "https://github.com/kubernetes/website.git", "7e7144c3969feb5d57a3c757ac462bd271f4a691", "sources/kubernetes/website/content/en/docs", "promoted public-vault lane report"},
		{"ockp-public-vault-go-docs", "https://github.com/golang/website.git", "31fb202f84245709e774bf7c85d13430925d45e5", "sources/golang/website/_content", "second large technical corpus autonomy validation"},
		{"ockp-public-vault-moby-dick", "https://github.com/GITenberg/Moby-Dick--Or-The-Whale_2701.git", "bdf1948e6cd00963730971e5624e764a35f238c3", "sources/gitenberg/moby-dick", "non-technical public-corpus autonomy validation"},
	}
	for _, report := range reports {
		if err := validatePublicVaultReport(root, report.name, report.repoURL, report.repoRef, report.vaultPrefix, report.decisionText); err != nil {
			return err
		}
	}
	return nil
}

func validatePublicVaultReport(root string, reportName string, repoURL string, repoRef string, vaultPrefix string, decisionText string) error {
	mdRel := filepath.ToSlash(filepath.Join("docs", "evals", "results", reportName+".md"))
	jsonRel := filepath.ToSlash(filepath.Join("docs", "evals", "results", reportName+".json"))
	mdContent, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(mdRel)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", mdRel, err)
	}
	mdText := strings.ReplaceAll(string(mdContent), "\r\n", "\n")
	required := []string{
		decisionText,
		repoURL,
		repoRef,
		vaultPrefix,
		"Decision: `promoted_lane`",
		"Open findings: `0`",
		"Findings status: `addressed`",
		"Passes gate: `true`",
		"must not include machine-local roots",
		"approval_mode: `autonomous_disposable`",
		"drafting_mode: `autonomous_fields`",
		"write_target_mode: `create_or_update`",
		"citation_mode: `balanced`",
		"privacy_mode: `allow_paths`",
		"audience_mode: `plain_language`",
	}
	normalized := strings.Join(strings.Fields(mdText), " ")
	for _, want := range required {
		if !strings.Contains(normalized, want) {
			return fmt.Errorf("%s must document public-vault evidence policy: missing %q", mdRel, want)
		}
	}
	for _, forbidden := range []string{"events.jsonl", "doc_", "chunk_", ".openclerk-eval"} {
		if strings.Contains(mdText, forbidden) {
			return fmt.Errorf("%s contains forbidden public-vault marker %q", mdRel, forbidden)
		}
	}
	jsonContent, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(jsonRel)))
	if err != nil {
		return fmt.Errorf("read %s: %w", jsonRel, err)
	}
	if strings.Contains(string(jsonContent), "events.jsonl") ||
		strings.Contains(string(jsonContent), "doc_") ||
		strings.Contains(string(jsonContent), "chunk_") ||
		strings.Contains(string(jsonContent), ".openclerk-eval") {
		return fmt.Errorf("%s contains raw-log or runner-internal public-vault marker", jsonRel)
	}
	var report struct {
		Summary struct {
			Decision       string `json:"decision"`
			RowsCompleted  int    `json:"rows_completed"`
			RowsFailed     int    `json:"rows_failed"`
			SafetyFailures int    `json:"safety_failures"`
			UXDebtRows     int    `json:"ux_debt_rows"`
			OpenFindings   int    `json:"open_findings"`
			FindingsStatus string `json:"findings_status"`
			PassesGate     bool   `json:"passes_gate"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(jsonContent, &report); err != nil {
		return fmt.Errorf("decode %s: %w", jsonRel, err)
	}
	if report.Summary.Decision != "promoted_lane" ||
		report.Summary.RowsCompleted != 8 ||
		report.Summary.RowsFailed != 0 ||
		report.Summary.SafetyFailures != 0 ||
		report.Summary.UXDebtRows != 0 ||
		report.Summary.OpenFindings != 0 ||
		report.Summary.FindingsStatus != "addressed" ||
		!report.Summary.PassesGate {
		return fmt.Errorf("%s must record promoted public-vault lane with 8 completed rows, zero failures/debt/findings, and addressed findings", jsonRel)
	}
	return nil
}

func validateLiveInstallSmokeReport(root string) error {
	rel := filepath.ToSlash(filepath.Join("docs", "evals", "results", "ockp-live-install-upgrade-module-smoke.json"))
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s is required; run scripts/validate-live-install-upgrade-module.sh", rel)
		}
		return fmt.Errorf("read %s: %w", rel, err)
	}
	var report struct {
		SchemaVersion string `json:"schema_version"`
		Install       struct {
			Passed              bool   `json:"passed"`
			InstallerInvocation string `json:"installer_invocation"`
			BinaryPath          string `json:"binary_path"`
			CommandPath         string `json:"command_path"`
			VersionOutput       string `json:"version_output"`
			HelpChecked         bool   `json:"help_checked"`
		} `json:"install"`
		Upgrade struct {
			Passed        bool   `json:"passed"`
			BinaryPath    string `json:"binary_path"`
			CommandPath   string `json:"command_path"`
			VersionOutput string `json:"version_output"`
			HelpChecked   bool   `json:"help_checked"`
		} `json:"upgrade"`
		Skill struct {
			Passed    bool   `json:"passed"`
			SkillPath string `json:"skill_path"`
			Source    string `json:"source"`
		} `json:"skill"`
		Module struct {
			Passed            bool              `json:"passed"`
			Provider          string            `json:"provider"`
			ManifestPath      string            `json:"manifest_path"`
			SkillPath         string            `json:"skill_path"`
			InstallPassed     bool              `json:"install_passed"`
			ConfigurePassed   bool              `json:"configure_passed"`
			UpgradePassed     bool              `json:"upgrade_passed"`
			UpgradePreserved  bool              `json:"upgrade_preserved_config"`
			ListPassed        bool              `json:"list_passed"`
			RemovePassed      bool              `json:"remove_passed"`
			FinalListEmpty    bool              `json:"final_list_empty"`
			ProviderConfig    map[string]string `json:"provider_config"`
			VerificationState string            `json:"verification_state"`
			RedactionState    string            `json:"redaction_state"`
		} `json:"module"`
		ValidationBoundaries string `json:"validation_boundaries"`
	}
	if err := json.Unmarshal(content, &report); err != nil {
		return fmt.Errorf("decode %s: %w", rel, err)
	}
	if report.SchemaVersion != "openclerk-live-install-smoke.v1" {
		return fmt.Errorf("%s schema_version = %q, want openclerk-live-install-smoke.v1", rel, report.SchemaVersion)
	}
	if !report.Install.Passed || !report.Upgrade.Passed || !report.Skill.Passed || !report.Module.Passed {
		return fmt.Errorf("%s must pass install, upgrade, skill, and module smoke checks", rel)
	}
	for _, check := range []struct {
		name        string
		binaryPath  string
		commandPath string
		version     string
		help        bool
	}{
		{name: "install", binaryPath: report.Install.BinaryPath, commandPath: report.Install.CommandPath, version: report.Install.VersionOutput, help: report.Install.HelpChecked},
		{name: "upgrade", binaryPath: report.Upgrade.BinaryPath, commandPath: report.Upgrade.CommandPath, version: report.Upgrade.VersionOutput, help: report.Upgrade.HelpChecked},
	} {
		if check.binaryPath != "$HOME/.local/bin/openclerk" || check.commandPath != "$HOME/.local/bin/openclerk" {
			return fmt.Errorf("%s %s must verify $HOME/.local/bin/openclerk binary and command path", rel, check.name)
		}
		if !strings.HasPrefix(check.version, "openclerk v") {
			return fmt.Errorf("%s %s version output must report openclerk version", rel, check.name)
		}
		if !check.help {
			return fmt.Errorf("%s %s must check openclerk --help", rel, check.name)
		}
	}
	if !strings.Contains(report.Install.InstallerInvocation, "curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh") {
		return fmt.Errorf("%s install must record the documented installer invocation", rel)
	}
	if report.Skill.SkillPath != "$CODEX_HOME/skills/openclerk/SKILL.md" || report.Skill.Source != "skills/openclerk/SKILL.md" {
		return fmt.Errorf("%s must verify the installed OpenClerk skill path and source", rel)
	}
	if report.Module.Provider != "ollama" ||
		report.Module.ManifestPath != "modules/ollama-embeddings/module.json" ||
		report.Module.SkillPath != "modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md" ||
		!report.Module.InstallPassed ||
		!report.Module.ConfigurePassed ||
		!report.Module.UpgradePassed ||
		!report.Module.UpgradePreserved ||
		!report.Module.ListPassed ||
		!report.Module.RemovePassed ||
		!report.Module.FinalListEmpty ||
		report.Module.ProviderConfig["embedding_model"] == "" ||
		report.Module.ProviderConfig["ollama_url"] != "http://localhost:11434" ||
		report.Module.VerificationState != "verified" ||
		report.Module.RedactionState != "redacted" {
		return fmt.Errorf("%s must verify ollama module install/config/upgrade/list/remove with preserved config and redacted verified state", rel)
	}
	for _, want := range []string{"temp HOME/CODEX_HOME", "no durable host install", "no network release fetch", "no direct SQLite edit"} {
		if !strings.Contains(report.ValidationBoundaries, want) {
			return fmt.Errorf("%s validation_boundaries missing %q", rel, want)
		}
	}
	return nil
}
