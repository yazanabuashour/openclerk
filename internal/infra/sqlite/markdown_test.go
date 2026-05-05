package sqlite

import "testing"

func TestNormalizeReferencePathUsesVaultPathPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "normalizes backslash separators",
			raw:  `sources\alpha`,
			want: "sources/alpha.md",
		},
		{
			name: "preserves doc refs",
			raw:  "doc:abc123",
			want: "doc:abc123",
		},
		{
			name: "drops traversal",
			raw:  `..\alpha.md`,
			want: "",
		},
		{
			name: "drops absolute",
			raw:  `/alpha.md`,
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeReferencePath(tt.raw); got != tt.want {
				t.Fatalf("normalizeReferencePath(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
