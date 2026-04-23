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
		"notes/synthesis/",
		"source_refs",
		"## Sources",
		"## Freshness",
		"provenance_events",
		"projection_states",
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
		"document path is missing",
		"name the missing fields",
		"ask the user to provide them",
		"limit -3",
		"bypass the runner",
		"SQLite",
		"HTTP",
		"MCP",
		"legacy or source-built paths",
		"unsupported transports",
		"this description is complete",
		"respond with exactly one no-tools assistant answer",
	} {
		if !strings.Contains(description, want) {
			t.Fatalf("SKILL.md description missing %q: %s", want, description)
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
