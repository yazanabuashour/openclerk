package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAcceptsValidSkillWithFrontmatter(t *testing.T) {
	t.Parallel()

	skillDir := writeSkill(t, "openclerk", `---
name: openclerk
description: Use OpenClerk for local-first knowledge-plane tasks through the installed openclerk JSON runner.
compatibility: Requires local filesystem access and an installed openclerk binary on PATH.
license: MIT
---

# OpenClerk

Use:

- openclerk document
- openclerk retrieval
`)

	var stdout bytes.Buffer
	if err := run([]string{skillDir}, &stdout); err != nil {
		t.Fatalf("run validator: %v", err)
	}
	if !strings.Contains(stdout.String(), "validated ") {
		t.Fatalf("stdout = %q, want validated message", stdout.String())
	}
}

func TestRunRejectsInvalidSkillPayloads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   map[string]string
		wantErr string
	}{
		{
			name: "extra file",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openclerk"),
				"notes.md": "extra",
			},
			wantErr: "must contain only SKILL.md",
		},
		{
			name: "missing opening delimiter",
			files: map[string]string{
				"SKILL.md": "name: openclerk\n",
			},
			wantErr: "must start with YAML frontmatter",
		},
		{
			name: "missing closing delimiter",
			files: map[string]string{
				"SKILL.md": "---\nname: openclerk\n",
			},
			wantErr: "must include a closing ---",
		},
		{
			name: "non scalar frontmatter",
			files: map[string]string{
				"SKILL.md": "---\nname: openclerk\ndescription\n---\n",
			},
			wantErr: "must be a scalar key-value pair",
		},
		{
			name: "missing required name",
			files: map[string]string{
				"SKILL.md": "---\ndescription: Use OpenClerk locally.\n---\n",
			},
			wantErr: "must define a non-empty name",
		},
		{
			name: "description too long",
			files: map[string]string{
				"SKILL.md": "---\nname: openclerk\ndescription: " + strings.Repeat("a", 1025) + "\n---\n",
			},
			wantErr: "description must be 1024 characters or fewer",
		},
		{
			name: "missing referenced file",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openclerk") + "\n[Reference](references/foo.md)\n",
			},
			wantErr: "is not installed with the skill",
		},
		{
			name: "referenced file outside skill directory",
			files: map[string]string{
				"SKILL.md":     validSkillMarkdown("openclerk") + "\n[Reference](../README.md)\n",
				"../README.md": "outside the packaged skill",
			},
			wantErr: "escapes skill directory",
		},
		{
			name: "retired runner binary name",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openclerk") + "\nRun `openclerk-agentops document`.\n",
			},
			wantErr: "retired product binary name",
		},
		{
			name: "retired transport",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openclerk") + "\nUse the generated client.\n",
			},
			wantErr: "retired transport guidance",
		},
		{
			name: "source built path",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openclerk") + "\nRun `go run ./cmd/openclerk document`.\n",
			},
			wantErr: "retired product guidance",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			skillDir := writeSkillFiles(t, "openclerk", tt.files)
			var stdout bytes.Buffer
			err := run([]string{skillDir}, &stdout)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("run error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func validSkillMarkdown(name string) string {
	return `---
name: ` + name + `
description: Use OpenClerk locally.
compatibility: Requires local filesystem access and an installed openclerk binary on PATH.
---

# OpenClerk

Run ` + "`openclerk document`" + `.
`
}

func writeSkill(t *testing.T, name string, content string) string {
	t.Helper()
	return writeSkillFiles(t, name, map[string]string{"SKILL.md": content})
}

func writeSkillFiles(t *testing.T, name string, files map[string]string) string {
	t.Helper()

	skillDir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	for fileName, content := range files {
		path := filepath.Join(skillDir, fileName)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	return skillDir
}
