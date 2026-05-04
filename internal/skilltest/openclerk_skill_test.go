package skilltest_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestOpenClerkSkillPayloadContainsOnlySkillMarkdown(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(openClerkSkillDir(t))
	if err != nil {
		t.Fatalf("read skill dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 1 || names[0] != "SKILL.md" {
		t.Fatalf("skill payload files = %v, want exactly SKILL.md", names)
	}
}

func TestOpenClerkSkillMarkdownLinksReferenceInstalledFiles(t *testing.T) {
	t.Parallel()

	skillDir := openClerkSkillDir(t)
	markdownFiles := []string{}
	if err := filepath.WalkDir(skillDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		markdownFiles = append(markdownFiles, path)
		return nil
	}); err != nil {
		t.Fatalf("walk markdown files: %v", err)
	}

	linkPattern := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	for _, path := range markdownFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, match := range linkPattern.FindAllStringSubmatch(string(content), -1) {
			target := match[1]
			if shouldSkipLinkTarget(target) {
				continue
			}
			targetPath := filepath.Clean(filepath.Join(filepath.Dir(path), target))
			if _, err := os.Stat(targetPath); err != nil {
				t.Fatalf("%s link target %q is not installed with the skill: %v", path, target, err)
			}
		}
	}
}

func TestOpenClerkSkillUsesInstalledRunnerForRoutineWork(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"name: OpenClerk",
		"openclerk document",
		"openclerk retrieval",
		"AgentOps",
		"Agent",
		"Skills-compatible",
		"Do not inspect source files",
		"synthesis/",
		"source_refs",
		"## Sources",
		"## Freshness",
		"provenance_events",
		"projection_states",
		"audit_contradictions",
		"compile_synthesis",
		"web_search_plan",
		"artifact_candidate_plan",
		"git_lifecycle_report",
		"source_audit_report",
		"evidence_bundle_report",
		"workflow_guide_report",
		"memory_router_recall_report",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("SKILL.md missing %q", want)
		}
	}
	for _, stale := range []string{
		"agentops",
		"cmd/openclerk-agentops",
		"client/" + "local",
		"go run",
		"local." + "Open" + "Client",
		"Open" + "Client",
		"S" + "DK",
		"WithResponse",
		"generated-" + "client",
		"client/" + "openclerk",
		"openapi",
		"cmd/openclerkd",
		"client/" + "fts",
		"client/" + "hybrid",
		"client/" + "graph",
		"client/" + "records",
		".agents/skills",
		".claude/skills",
		".openclaw/skills",
		"~/.codex",
		"Codex",
		"Claude",
		"OpenClaw",
		"Hermes",
		"temporary Go module",
		"mktemp",
	} {
		if strings.Contains(text, stale) {
			t.Fatalf("SKILL.md contains stale routine guidance %q", stale)
		}
	}
}

func TestOpenClerkSkillStaysWithinThinRouterBudget(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) > 225 {
		t.Fatalf("SKILL.md line count = %d, want <= 225", len(lines))
	}

	text := strings.Join(strings.Fields(string(content)), " ")
	for _, want := range []string{
		"activation, routing, and safety contract",
		"not the durable home for long workflow recipes",
		"use agent autonomy with runner JSON results",
		"workflow-action comparison",
		"openclerk document --help",
		"openclerk retrieval --help",
		"Detailed versions of these workflows belong in runner actions",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("SKILL.md missing thin-router contract %q", want)
		}
	}
}

func TestOpenClerkSkillDescriptionContainsBootstrapNoToolsGuard(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	description := frontmatterDescription(string(content))
	if description == "" {
		t.Fatal("SKILL.md frontmatter description is empty")
	}
	for _, want := range []string{
		"Bootstrap no-tools rule",
		"required fields are missing",
		"document path, title, or body is missing",
		"explicit user content for a faithful candidate, duplicate-risk check, or public-link placement proposal",
		"numeric limit is negative",
		"bypass the runner",
		"SQLite",
		"raw vault/file/repo inspection",
		"HTTP/MCP",
		"legacy/source-built paths",
		"unsupported transports",
		"backend variants",
		"module-cache inspection",
		"rg",
		"find",
		"ls",
		"OCR",
		"browser automation",
		"local file reads",
		"opaque artifact parsing",
		"this description is complete",
		"For those invalid cases only",
		"With explicit user content, validate a faithful candidate through the runner but do not write before approval",
		"Do not open this skill file",
		"run commands",
		"use tools",
		"call the runner",
		"respond with exactly one no-tools assistant answer naming the missing/invalid fields or unsupported workflow",
		"openclerk document",
		"openclerk retrieval",
	} {
		if !strings.Contains(description, want) {
			t.Fatalf("SKILL.md description missing %q: %s", want, description)
		}
	}
	if len([]rune(description)) > 1024 {
		t.Fatalf("description length = %d, want <= 1024", len([]rune(description)))
	}
}

func TestOpenClerkSkillKeepsWorkflowPoliciesCompact(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	policySection := markdownSection(text, "## Workflow Policies", "## Document Tasks")
	if policySection == "" {
		t.Fatal("missing workflow policies section")
	}
	normalized := strings.Join(strings.Fields(policySection), " ")
	for _, want := range []string{
		"Candidate documents",
		"preserve explicit user path/title/body/type/naming instructions",
		"validate with `openclerk document` before presenting a candidate",
		"`notes/candidates/<slug-from-title>.md`",
		"concise singular noun phrase",
		"Include `type: note` frontmatter",
		"`# <Title>` heading",
		"`Path:`, `Title:`, and `Body preview:`",
		"Duplicate checks",
		"duplicate risk is requested or plausible",
		"runner-visible evidence",
		"no document was created or updated",
		"update the existing target or create a confirmed new path",
		"Public URL/source intake",
		"`ingest_source_url`",
		"Do not fetch URLs with browser, HTTP, filesystem, or other non-runner tools",
		"Document lifecycle review, rollback, restore, and semantic diff",
		"Use `git_lifecycle_report` only for local Git status/history/checkpoints",
		"There is no public raw diff, restore, or rollback action",
		"Messy populated-vault retrieval",
		"metadata-filtered authority results",
		"polluted, decoy, stale, draft, archived, duplicate, or candidate documents as non-authority",
		"Synthesis maintenance",
		"prefer `compile_synthesis`",
		"Detailed versions of these workflows belong in runner actions, compact runner help, maintainer/eval docs, or follow-up candidate-surface comparisons",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("workflow policy missing %q", want)
		}
	}
}

func TestOpenClerkSkillRejectsRecipeCreep(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	for _, forbidden := range []string{
		"## Propose-Before-Create Candidate Documents",
		"## Lifecycle Quick Rules",
		"Unsupported opaque artifact rules",
		"Parser and acquisition bypass rules",
		"sources/candidates/<slug-from-label-or-url>.md",
		"synthesis/<shared-topic-or-url-set>.md",
		"save this note from what we discussed last week",
		"Do not answer with only validation status",
		"workflow is incomplete even when validation passed",
		"Do not call `validate`, `create_document`, `append_document`, or `replace_section` while duplicate",
		"Common request shapes:",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("SKILL.md contains recipe creep %q", forbidden)
		}
	}
}

func openClerkSkillPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(openClerkSkillDir(t), "SKILL.md")
}

func openClerkSkillDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "skills", "openclerk")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate current test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}

func frontmatterDescription(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return ""
	}
	for _, line := range lines[1:] {
		if line == "---" {
			return ""
		}
		if strings.HasPrefix(line, "description: ") {
			return strings.TrimPrefix(line, "description: ")
		}
	}
	return ""
}

func shouldSkipLinkTarget(target string) bool {
	return target == "" ||
		strings.HasPrefix(target, "#") ||
		strings.Contains(target, "://") ||
		filepath.IsAbs(target)
}

func markdownSection(content string, startHeading string, endHeading string) string {
	start := strings.Index(content, startHeading)
	if start == -1 {
		return ""
	}
	rest := content[start:]
	end := strings.Index(rest, endHeading)
	if end == -1 {
		return rest
	}
	return rest[:end]
}
