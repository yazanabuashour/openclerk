package skilltest_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestModuleSkillsStayCompactAndBoundaryFocused(t *testing.T) {
	t.Parallel()

	modules := moduleSkillManifests(t)
	if len(modules) == 0 {
		t.Fatal("no module skills discovered from module manifests")
	}
	for _, module := range modules {
		module := module
		t.Run(module.name, func(t *testing.T) {
			t.Parallel()

			contentBytes, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(module.skillPath)))
			if err != nil {
				t.Fatalf("read module skill %s: %v", module.skillPath, err)
			}
			content := string(contentBytes)
			lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
			if len(lines) > 80 {
				t.Fatalf("%s line count = %d, want <= 80", module.skillPath, len(lines))
			}
			metadata := scalarFrontmatter(t, module.skillPath, content)
			if utf8.RuneCountInString(metadata["description"]) > 300 {
				t.Fatalf("%s description length = %d, want <= 300", module.skillPath, utf8.RuneCountInString(metadata["description"]))
			}
			if utf8.RuneCountInString(metadata["compatibility"]) > 300 {
				t.Fatalf("%s compatibility length = %d, want <= 300", module.skillPath, utf8.RuneCountInString(metadata["compatibility"]))
			}

			normalized := strings.Join(strings.Fields(content), " ")
			for _, want := range []string{
				"optional",
				"module",
				"installed and enabled",
				"production OpenClerk skill",
				"direct SQLite",
				"module-cache inspection",
				"repo-relative paths or neutral placeholders",
			} {
				if !strings.Contains(normalized, want) {
					t.Fatalf("%s missing module boundary language %q", module.skillPath, want)
				}
			}
			switch module.kind {
			case "embedding_provider":
				for _, want := range []string{
					"openclerk retrieval",
					"semantic_search",
					"citation",
					"does not replace the production OpenClerk skill",
					"default semantic ranking",
				} {
					if !strings.Contains(normalized, want) {
						t.Fatalf("%s missing embedding boundary language %q", module.skillPath, want)
					}
				}
			case "ocr_provider":
				for _, want := range []string{
					"openclerk document runner",
					"artifact_candidate_plan",
					"read-only",
					"Durable writes still require explicit approval",
				} {
					if !strings.Contains(normalized, want) {
						t.Fatalf("%s missing OCR boundary language %q", module.skillPath, want)
					}
				}
			default:
				t.Fatalf("%s has unsupported module kind %q", module.skillPath, module.kind)
			}
		})
	}
}

type moduleSkillManifest struct {
	name      string
	kind      string
	skillPath string
}

func moduleSkillManifests(t *testing.T) []moduleSkillManifest {
	t.Helper()
	root := repoRoot(t)
	out := []moduleSkillManifest{}
	if err := filepath.WalkDir(filepath.Join(root, "modules"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "module.json" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var manifest struct {
			Module struct {
				Name string `json:"name"`
				Kind string `json:"kind"`
			} `json:"module"`
			Provides []struct {
				Type string `json:"type"`
				Path string `json:"path"`
			} `json:"provides"`
		}
		if err := json.Unmarshal(content, &manifest); err != nil {
			return err
		}
		skillCount := 0
		for _, provided := range manifest.Provides {
			if provided.Type == "skill" && strings.TrimSpace(provided.Path) != "" {
				skillCount++
				out = append(out, moduleSkillManifest{
					name:      manifest.Module.Name,
					kind:      manifest.Module.Kind,
					skillPath: filepath.ToSlash(provided.Path),
				})
			}
		}
		if skillCount == 0 && manifest.Module.Kind != "retrieval_adapter" {
			return fmt.Errorf("%s module kind %q must provide a skill path", filepath.ToSlash(path), manifest.Module.Kind)
		}
		return nil
	}); err != nil {
		t.Fatalf("discover module skill manifests: %v", err)
	}
	return out
}

func scalarFrontmatter(t *testing.T, path string, content string) map[string]string {
	t.Helper()
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		t.Fatalf("%s must start with frontmatter", path)
	}
	metadata := map[string]string{}
	for _, line := range lines[1:] {
		if line == "---" {
			return metadata
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("%s frontmatter line must be key/value: %s", path, line)
		}
		metadata[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	t.Fatalf("%s missing closing frontmatter delimiter", path)
	return nil
}
