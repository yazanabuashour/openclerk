package openclerkskill_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestSkillPayloadContainsOnlySkillMarkdown(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read skill dir: %v", err)
	}
	payloadEntries := []string{}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		payloadEntries = append(payloadEntries, entry.Name())
	}
	if len(payloadEntries) != 1 || payloadEntries[0] != "SKILL.md" {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("skill payload files = %v, want exactly SKILL.md", names)
	}
}

func TestSkillMarkdownLinksReferenceInstalledFiles(t *testing.T) {
	t.Parallel()

	markdownFiles := []string{}
	if err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
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

func TestSkillUsesInstalledRunnerForRoutineWork(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("SKILL.md")
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"name: openclerk",
		"openclerk document",
		"openclerk retrieval",
		"Agent",
		"Skills-compatible",
		"Do not inspect source files",
		"reject final-answer-only",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("SKILL.md missing %q", want)
		}
	}
	for _, stale := range []string{
		"AgentOps",
		"agentops",
		"cmd/openclerk-agentops",
		"go run",
		"local.OpenClient",
		"WithResponse",
		"generated-client",
		"client/openclerk",
		"openapi",
		"cmd/openclerkd",
		"client/fts",
		"client/hybrid",
		"client/graph",
		"client/records",
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

func shouldSkipLinkTarget(target string) bool {
	return target == "" ||
		strings.HasPrefix(target, "#") ||
		strings.Contains(target, "://") ||
		filepath.IsAbs(target)
}
