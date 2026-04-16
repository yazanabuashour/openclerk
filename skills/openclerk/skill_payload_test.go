package openclerkskill_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

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

func TestSkillUsesCodeFirstFacadeForRoutineWork(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("SKILL.md")
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "local.OpenClient") {
		t.Fatal("SKILL.md must point routine work at local.OpenClient")
	}
	if strings.Contains(text, "WithResponse") {
		t.Fatal("SKILL.md should not steer routine tasks to generated WithResponse methods")
	}
}

func shouldSkipLinkTarget(target string) bool {
	return target == "" ||
		strings.HasPrefix(target, "#") ||
		strings.Contains(target, "://") ||
		filepath.IsAbs(target)
}
