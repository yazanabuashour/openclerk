package domain

import "testing"

func TestNormalizeVaultRelativePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		want      string
		wantIssue VaultPathIssue
	}{
		{
			name: "cleans forward slash path",
			raw:  "sources/../sources/runner.md",
			want: "sources/runner.md",
		},
		{
			name: "normalizes backslash separators",
			raw:  `sources\runner.md`,
			want: "sources/runner.md",
		},
		{
			name:      "rejects backslash traversal",
			raw:       `..\runner.md`,
			wantIssue: VaultPathEscapesRoot,
		},
		{
			name:      "rejects leading slash after normalization",
			raw:       `\tmp\runner.md`,
			wantIssue: VaultPathAbsolute,
		},
		{
			name:      "rejects windows drive paths on unix hosts",
			raw:       `C:\tmp\runner.md`,
			wantIssue: VaultPathAbsolute,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, issue := NormalizeVaultRelativePath(tt.raw)
			if issue != tt.wantIssue {
				t.Fatalf("issue = %q, want %q", issue, tt.wantIssue)
			}
			if got != tt.want {
				t.Fatalf("path = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeOptionalVaultRelativePrefixPreservesTrailingSlash(t *testing.T) {
	t.Parallel()

	got, issue := NormalizeOptionalVaultRelativePrefix(`sources\reports\`)
	if issue != VaultPathOK {
		t.Fatalf("issue = %q, want ok", issue)
	}
	if got != "sources/reports/" {
		t.Fatalf("prefix = %q, want sources/reports/", got)
	}
}

func TestVaultIgnoreMatcherUsesDefaultsAndConfigurablePaths(t *testing.T) {
	t.Parallel()

	rules, err := EffectiveVaultIgnorePaths([]string{`scratch\private`, ".git"})
	if err != nil {
		t.Fatalf("effective ignore paths: %v", err)
	}
	wantRules := []string{".stversions", ".git", ".openclerk", ".backups", "scratch/private"}
	if len(rules) != len(wantRules) {
		t.Fatalf("rules = %+v, want %+v", rules, wantRules)
	}
	for i, want := range wantRules {
		if rules[i] != want {
			t.Fatalf("rules[%d] = %q, want %q; rules=%+v", i, rules[i], want, rules)
		}
	}

	matcher, err := NewVaultIgnoreMatcher(rules)
	if err != nil {
		t.Fatalf("new matcher: %v", err)
	}
	for _, path := range []string{".git", ".git/objects/pack", "scratch/private", "scratch/private/note.md"} {
		if !matcher.Matches(path) {
			t.Fatalf("matcher did not ignore %q", path)
		}
	}
	for _, path := range []string{"scratch/private-notes.md", "sources/.git-notes.md", "sources/live.md"} {
		if matcher.Matches(path) {
			t.Fatalf("matcher unexpectedly ignored %q", path)
		}
	}
}

func TestVaultIgnorePathsRejectAbsoluteOrEscapingPaths(t *testing.T) {
	t.Parallel()

	if _, err := NormalizeVaultIgnorePaths([]string{"/tmp/vault"}); err == nil {
		t.Fatalf("absolute ignore path accepted")
	}
	if _, err := NormalizeVaultIgnorePaths([]string{"../outside"}); err == nil {
		t.Fatalf("escaping ignore path accepted")
	}
}
