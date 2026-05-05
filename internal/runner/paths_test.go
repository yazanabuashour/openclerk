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

	if rejection := validateSourcePathHint(""); rejection != "source.path_hint is required" {
		t.Fatalf("missing source path hint rejection = %q", rejection)
	}
	if rejection := validateSourcePathHint("/tmp/source.md"); rejection != "source.path_hint must be relative to the vault root" {
		t.Fatalf("absolute source path hint rejection = %q", rejection)
	}
	if rejection := validateSourcePathHint("../source.md"); rejection != "source.path_hint must stay inside the vault root" {
		t.Fatalf("escaping source path hint rejection = %q", rejection)
	}
	if rejection := validateSourcePathHint("notes/source.md"); rejection != "source.path_hint must be a vault-relative sources/*.md path" {
		t.Fatalf("wrong source path hint rejection = %q", rejection)
	}
	if rejection := validateSourcePathHint("sources/example.md"); rejection != "" {
		t.Fatalf("valid source path hint rejection = %q", rejection)
	}
	if rejection := validateVideoPathHint("sources/video-youtube/example.md"); rejection != "" {
		t.Fatalf("valid video path hint rejection = %q", rejection)
	}
	if rejection := validateAssetPathHint("assets/sources/example.pdf"); rejection != "" {
		t.Fatalf("valid asset path hint rejection = %q", rejection)
	}
	if rejection := validateVideoAssetPathHint("assets/video-youtube/example.json"); rejection != "" {
		t.Fatalf("valid video asset path hint rejection = %q", rejection)
	}
	if rejection := validateVideoAssetPathHint("assets/video-youtube/example.txt"); rejection != "video.asset_path_hint must be a vault-relative assets/**/*.json path" {
		t.Fatalf("wrong video asset path hint rejection = %q", rejection)
	}
}
