package sqlite

import (
	"slices"
	"testing"
)

func TestNormalizeDocumentSourceURLCanonicalizesGitHubMarkdownSources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "repository landing page",
			raw:  "https://github.com/mvanhorn/last30days-skill#readme",
			want: "https://raw.githubusercontent.com/mvanhorn/last30days-skill/HEAD/README.md",
		},
		{
			name: "markdown blob",
			raw:  "https://github.com/mvanhorn/last30days-skill/blob/main/skills/last30days/SKILL.md?plain=1",
			want: "https://raw.githubusercontent.com/mvanhorn/last30days-skill/main/skills/last30days/SKILL.md",
		},
		{
			name: "non markdown blob stays on original URL",
			raw:  "https://github.com/mvanhorn/last30days-skill/blob/main/package.json",
			want: "https://github.com/mvanhorn/last30days-skill/blob/main/package.json",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeDocumentSourceURL(tt.raw)
			if err != nil {
				t.Fatalf("normalize document source URL: %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeDocumentSourceURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNormalizeDocumentSourceURLWithAliasesIncludesGitHubOriginals(t *testing.T) {
	t.Parallel()

	sourceURL, lookupURLs, err := normalizeDocumentSourceURLWithAliases("https://raw.githubusercontent.com/mvanhorn/last30days-skill/HEAD/README.md")
	if err != nil {
		t.Fatalf("normalize document source URL aliases: %v", err)
	}
	if sourceURL != "https://raw.githubusercontent.com/mvanhorn/last30days-skill/HEAD/README.md" ||
		!slices.Contains(lookupURLs, sourceURL) ||
		!slices.Contains(lookupURLs, "https://github.com/mvanhorn/last30days-skill") ||
		!slices.Contains(lookupURLs, "https://github.com/mvanhorn/last30days-skill/") ||
		!slices.Contains(lookupURLs, "https://github.com/mvanhorn/last30days-skill/blob/HEAD/README.md") {
		t.Fatalf("sourceURL=%q lookupURLs=%+v", sourceURL, lookupURLs)
	}
}
