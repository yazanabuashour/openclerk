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
