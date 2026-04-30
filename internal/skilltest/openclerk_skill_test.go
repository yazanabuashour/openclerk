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
		"one no-tools assistant answer",
		"synthesis/",
		"source_refs",
		"## Sources",
		"## Freshness",
		"provenance_events",
		"projection_states",
		"audit_contradictions",
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
		"name the missing fields",
		"ask the user to provide them",
		"Do not open this skill file",
		"run commands",
		"use tools",
		"call the runner",
		"limit -3",
		"bypass the runner",
		"SQLite",
		"HTTP",
		"MCP",
		"legacy or source-built paths",
		"unsupported transports",
		"this description is complete",
		"respond with exactly one no-tools assistant answer",
		"openclerk document",
		"openclerk retrieval",
		"rg --files",
		"find",
		"ls",
		"direct vault inspection",
		"repo search",
	} {
		if !strings.Contains(description, want) {
			t.Fatalf("SKILL.md description missing %q: %s", want, description)
		}
	}
}

func TestOpenClerkSkillRejectsPollutedPopulatedVaultEvidence(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	normalized := strings.Join(strings.Fields(text), " ")
	for _, want := range []string{
		"messy populated-vault retrieval",
		"Metadata-filtered authority results",
		"stale, draft, archived, duplicate, or candidate",
		"explicitly reject that hit as not authority",
		"do not repeat its false claim text as a valid answer",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("SKILL.md missing polluted-evidence guidance %q", want)
		}
	}
	for _, want := range []string{"`status: polluted`", "`populated_role: decoy`"} {
		if !strings.Contains(text, want) {
			t.Fatalf("SKILL.md missing polluted-evidence guidance %q", want)
		}
	}
	if !strings.Contains(normalized, "active canonical sources") {
		t.Fatal("SKILL.md missing polluted-evidence guidance for active canonical sources")
	}
}

func TestOpenClerkSkillGuidesDuplicateCandidateClarification(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	proposalSection := markdownSection(text, "## Propose-Before-Create Candidate Documents", "## Document Tasks")
	if proposalSection == "" {
		t.Fatal("missing propose-before-create section")
	}
	normalized := strings.Join(strings.Fields(proposalSection), " ")
	for _, want := range []string{
		"duplicate risk is requested or plausible",
		"valid runner-backed capture work",
		"retrieval `search`",
		"document `list_documents`",
		"include that `path_prefix` in the retrieval search",
		"use the same prefix for `list_documents`",
		"`get_document`",
		"likely target path and title",
		"search/list/get evidence",
		"no document was created or updated",
		"update the existing target or create a new document at a confirmed path",
		"Do not call `validate`, `create_document`, `append_document`, or `replace_section`",
		"update-versus-new-path intent is unresolved",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("propose-before-create guidance missing %q", want)
		}
	}
	for _, forbidden := range []string{
		"new runner action",
		"runner action",
		"schema",
		"public API",
		"storage migration",
		"direct-create shortcut",
	} {
		if strings.Contains(strings.ToLower(proposalSection), forbidden) {
			t.Fatalf("duplicate-candidate guidance contains promotion language %q", forbidden)
		}
	}
}

func TestOpenClerkSkillGuidesSaveThisNotePolicy(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	proposalSection := markdownSection(text, "## Propose-Before-Create Candidate Documents", "## Document Tasks")
	if proposalSection == "" {
		t.Fatal("missing propose-before-create section")
	}
	normalized := strings.Join(strings.Fields(proposalSection), " ")
	for _, want := range []string{
		`"save this note" requests with explicit note content but no path or title`,
		"derive a faithful note candidate from the supplied content",
		"validate it",
		"show the candidate",
		"state that no document was created",
		"ask for approval before creating anything",
		"bare prior-context requests",
		"save this note from what we discussed last week",
		"use the no-tools rule",
		"ask for the actual note content plus any path, title, or placement preferences",
		"do not invent a path, title, or body",
		"notes/candidates/<slug-from-title>.md",
		"Final answers for proposals show `Path:`, `Title:`, and `Body preview:`",
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("save-this-note guidance missing %q", want)
		}
	}
	for _, forbidden := range []string{
		"runner action",
		"schema",
		"public API",
		"storage migration",
		"direct-create shortcut",
	} {
		if strings.Contains(strings.ToLower(proposalSection), forbidden) {
			t.Fatalf("save-this-note guidance contains promotion language %q", forbidden)
		}
	}
}

func TestOpenClerkSkillDescriptionDoesNotSuppressDuplicateRiskChecks(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(openClerkSkillPath(t))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	description := frontmatterDescription(string(content))
	for _, want := range []string{
		"faithful propose-before-create candidate or duplicate-risk check",
		"explicit user content",
		"this description is complete",
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
