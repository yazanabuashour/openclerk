package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAcceptsValidReleaseDocs(t *testing.T) {
	root := writeReleaseRepo(t, "v0.1.0", validReleaseNotes("v0.1.0"), changelogFor("v0.1.0"))
	var stdout bytes.Buffer
	withWorkingDir(t, root, func() {
		if err := run([]string{"v0.1.0"}, &stdout); err != nil {
			t.Fatalf("run validator: %v", err)
		}
	})
	if !strings.Contains(stdout.String(), "validated release docs for v0.1.0") {
		t.Fatalf("stdout = %q, want validated message", stdout.String())
	}
}

func TestValidateReleaseDocsRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tag       string
		notes     *string
		changelog string
		wantErr   string
	}{
		{
			name:      "invalid tag",
			tag:       "0.1.0",
			notes:     strPtr(validReleaseNotes("v0.1.0")),
			changelog: changelogFor("v0.1.0"),
			wantErr:   "tag must match",
		},
		{
			name:      "missing release notes",
			tag:       "v0.1.0",
			notes:     nil,
			changelog: changelogFor("v0.1.0"),
			wantErr:   "docs/release-notes/v0.1.0.md not found",
		},
		{
			name:      "wrong title",
			tag:       "v0.1.0",
			notes:     strPtr(strings.Replace(validReleaseNotes("v0.1.0"), "# OpenClerk v0.1.0", "# OpenClerk 0.1.0", 1)),
			changelog: changelogFor("v0.1.0"),
			wantErr:   `must start with "# OpenClerk v0.1.0"`,
		},
		{
			name:      "missing changed section",
			tag:       "v0.1.0",
			notes:     strPtr(strings.Replace(validReleaseNotes("v0.1.0"), "## Changed", "## Updates", 1)),
			changelog: changelogFor("v0.1.0"),
			wantErr:   "must include ## Changed",
		},
		{
			name:      "missing verification section",
			tag:       "v0.1.0",
			notes:     strPtr(strings.Replace(validReleaseNotes("v0.1.0"), "## Verification", "## Tests", 1)),
			changelog: changelogFor("v0.1.0"),
			wantErr:   "must include ## Verification",
		},
		{
			name:      "missing changelog link",
			tag:       "v0.1.0",
			notes:     strPtr(validReleaseNotes("v0.1.0")),
			changelog: "# Changelog\n\nNo matching release link.\n",
			wantErr:   "CHANGELOG.md must link",
		},
		{
			name: "hard wrapped prose",
			tag:  "v0.1.0",
			notes: strPtr(`# OpenClerk v0.1.0

This paragraph was manually wrapped before the end of the prose sentence
and should be rejected by the release-doc validator.

## Changed

- Added a thing.

## Verification

- Checked a thing.
`),
			changelog: changelogFor("v0.1.0"),
			wantErr:   "appears to hard-wrap prose",
		},
		{
			name: "hard wrapped bullet",
			tag:  "v0.1.0",
			notes: strPtr(`# OpenClerk v0.1.0

## Changed

- Added release notes validation that should reject manually wrapped
  list item continuation text.

## Verification

- Checked a thing.
`),
			changelog: changelogFor("v0.1.0"),
			wantErr:   "appears to hard-wrap list item",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := writeReleaseRepo(t, tt.tag, derefString(tt.notes), tt.changelog)
			if tt.notes == nil {
				notesPath := filepath.Join(root, "docs", "release-notes", tt.tag+".md")
				if err := os.Remove(notesPath); err != nil && !os.IsNotExist(err) {
					t.Fatalf("remove notes: %v", err)
				}
			}
			err := validateReleaseDocs(root, tt.tag)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateReleaseDocs error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateReleaseNotesAllowsMarkdownStructure(t *testing.T) {
	t.Parallel()

	notes := `# OpenClerk v0.1.0

Short standalone prose is okay.

## Changed

- Bullets may wrap naturally in rendered Markdown without counting as prose.
1. Ordered list items are allowed.

` + "```" + `
Code fences can have consecutive plain-looking lines.
They are not release prose.
` + "```" + `

## Verification

- Verification bullets are okay.
`
	if err := validateReleaseNotes("docs/release-notes/v0.1.0.md", notes, "v0.1.0"); err != nil {
		t.Fatalf("validateReleaseNotes: %v", err)
	}
}

func TestValidateReleaseNotesAcceptsV020SourceURLUpdateCoverage(t *testing.T) {
	t.Parallel()

	if err := validateReleaseNotes("docs/release-notes/v0.2.0.md", validV020ReleaseNotes(), "v0.2.0"); err != nil {
		t.Fatalf("validateReleaseNotes: %v", err)
	}
}

func TestValidateReleaseNotesRejectsIncompleteV020SourceURLUpdateCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		replace string
		with    string
		wantErr string
	}{
		{
			name:    "missing update mode semantics",
			replace: "`source.mode: \"update\"` re-ingests an existing `source.url`",
			with:    "`source.mode` supports updates",
			wantErr: "source URL update mode defaults",
		},
		{
			name:    "missing conflict no-write behavior",
			replace: "conflict without writing extra docs or assets",
			with:    "conflict cleanly",
			wantErr: "duplicate and path-hint conflicts",
		},
		{
			name:    "missing same SHA no-op behavior",
			replace: "same-SHA updates are no-ops",
			with:    "same-SHA updates are detected",
			wantErr: "same-SHA no-op behavior",
		},
		{
			name:    "missing stale synthesis visibility",
			replace: "`projection_states`",
			with:    "projection state output",
			wantErr: "changed-PDF stale synthesis visibility",
		},
		{
			name:    "missing targeted evidence pointer",
			replace: "docs/evals/results/ockp-source-url-update-mode.md",
			with:    "docs/evals/results/ockp-latest.md",
			wantErr: "targeted source URL update evidence",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			notes := strings.Replace(validV020ReleaseNotes(), tt.replace, tt.with, 1)
			if notes == validV020ReleaseNotes() {
				t.Fatalf("test replacement %q did not change fixture", tt.replace)
			}
			err := validateReleaseNotes("docs/release-notes/v0.2.0.md", notes, "v0.2.0")
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateReleaseNotes error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func validReleaseNotes(tag string) string {
	return "# OpenClerk " + tag + `

This release adds the local-first runner and keeps release prose on one source line so GitHub Releases can wrap it naturally.

## Changed

- Added the OpenClerk JSON runner and Agent Skills-compatible skill.

## Verification

- Production eval passed all selected scenarios.
`
}

func validV020ReleaseNotes() string {
	return `# OpenClerk v0.2.0

This release tightens the installed OpenClerk runner and skill contract after the first public release, with clearer routine knowledge-task policy, database-anchored vault configuration, and release-gate evidence for the current AgentOps workflow.

## Changed

- Documented ` + "`ingest_source_url`" + ` update mode: missing ` + "`source.mode`" + ` defaults to ` + "`create`" + `, duplicate creates reject, ` + "`source.mode: \"update\"`" + ` re-ingests an existing ` + "`source.url`" + `, and mismatched ` + "`path_hint`" + ` or ` + "`asset_path_hint`" + ` values conflict without writing extra docs or assets.
- Verified source URL update freshness behavior: same-SHA updates are no-ops without ` + "`source_updated`" + ` provenance or projection invalidation churn, while changed-PDF updates preserve source and asset paths, refresh searchable extracted text and citations, emit previous/new SHA provenance, and expose stale dependent synthesis through ` + "`projection_states`" + `.

## Verification

- Targeted source URL update evidence is committed at ` + "`docs/evals/results/ockp-source-url-update-mode.md`" + `, covering duplicate create rejection, same-SHA no-op behavior, changed-PDF stale synthesis visibility, and path-hint conflict no-write behavior for the shipped runner and skill.
`
}

func changelogFor(tag string) string {
	return "# Changelog\n\n- [" + tag + "](https://github.com/yazanabuashour/openclerk/releases/tag/" + tag + ") adds release docs validation.\n"
}

func writeReleaseRepo(t *testing.T, tag string, notes string, changelog string) string {
	t.Helper()

	root := t.TempDir()
	if notes != "" {
		notesPath := filepath.Join(root, "docs", "release-notes", tag+".md")
		if err := os.MkdirAll(filepath.Dir(notesPath), 0o755); err != nil {
			t.Fatalf("mkdir release notes dir: %v", err)
		}
		if err := os.WriteFile(notesPath, []byte(notes), 0o644); err != nil {
			t.Fatalf("write release notes: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "CHANGELOG.md"), []byte(changelog), 0o644); err != nil {
		t.Fatalf("write changelog: %v", err)
	}
	return root
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("restore dir %s: %v", oldDir, err)
		}
	}()
	fn()
}

func strPtr(value string) *string {
	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
