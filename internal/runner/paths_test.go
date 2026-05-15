package runner

import "testing"

func TestRunnerVaultRelativePathNormalization(t *testing.T) {
	t.Parallel()

	if got := normalizeVaultRelativePath(" ./sources/../sources/example.md "); got != "sources/example.md" {
		t.Fatalf("normalizeVaultRelativePath = %q, want sources/example.md", got)
	}
	if got := normalizeVaultRelativePath("../escape.md"); got != "../escape.md" {
		t.Fatalf("unsafe normalizeVaultRelativePath = %q, want original trimmed path", got)
	}
	if got := normalizeVaultRelativePrefix(" sources//web "); got != "sources/web" {
		t.Fatalf("normalizeVaultRelativePrefix = %q, want sources/web", got)
	}
}

func TestRunnerPathHintValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		validate func(string) string
		pathHint string
		want     string
	}{
		{
			name:     "source path hint is required",
			validate: validateSourcePathHint,
			want:     "source.path_hint is required",
		},
		{
			name:     "source path hint must be relative",
			validate: validateSourcePathHint,
			pathHint: "/tmp/source.md",
			want:     "source.path_hint must be relative to the vault root",
		},
		{
			name:     "source path hint must stay inside root",
			validate: validateSourcePathHint,
			pathHint: "../source.md",
			want:     "source.path_hint must stay inside the vault root",
		},
		{
			name:     "source path hint must use sources markdown shape",
			validate: validateSourcePathHint,
			pathHint: "notes/source.md",
			want:     "source.path_hint must be a vault-relative sources/*.md path",
		},
		{
			name:     "source path hint accepts source markdown",
			validate: validateSourcePathHint,
			pathHint: "sources/example.md",
		},
		{
			name:     "video path hint accepts source markdown",
			validate: validateVideoPathHint,
			pathHint: "sources/video-youtube/example.md",
		},
		{
			name:     "source asset path hint accepts pdf asset",
			validate: validateAssetPathHint,
			pathHint: "assets/sources/example.pdf",
		},
		{
			name:     "video asset path hint accepts json asset",
			validate: validateVideoAssetPathHint,
			pathHint: "assets/video-youtube/example.json",
		},
		{
			name:     "video asset path hint requires json asset",
			validate: validateVideoAssetPathHint,
			pathHint: "assets/video-youtube/example.txt",
			want:     "video.asset_path_hint must be a vault-relative assets/**/*.json path",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.validate(tt.pathHint); got != tt.want {
				t.Fatalf("rejection = %q, want %q", got, tt.want)
			}
		})
	}
}
