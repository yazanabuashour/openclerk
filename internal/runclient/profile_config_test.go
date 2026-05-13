package runclient

import (
	"context"
	"path/filepath"
	"testing"
)

func TestDefaultProfileRuntimeConfigRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	initial, err := ReadDefaultProfileConfig(ctx, config)
	if err != nil {
		t.Fatalf("read initial profile: %v", err)
	}
	if len(initial) != 0 {
		t.Fatalf("initial profile = %+v", initial)
	}

	values := map[string]string{
		"approval_mode":     "propose_only",
		"drafting_mode":     "require_explicit_fields",
		"write_target_mode": "existing_only",
		"citation_mode":     "strict",
		"privacy_mode":      "private_summary_only",
		"audience_mode":     "executive_summary",
	}
	if err := WriteDefaultProfileConfig(ctx, config, values); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	read, err := ReadDefaultProfileConfig(ctx, config)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	for key, want := range values {
		if read[key] != want {
			t.Fatalf("profile %s = %q, want %q; profile=%+v", key, read[key], want, read)
		}
	}

	if err := ClearDefaultProfileConfig(ctx, config); err != nil {
		t.Fatalf("clear profile: %v", err)
	}
	cleared, err := ReadDefaultProfileConfig(ctx, config)
	if err != nil {
		t.Fatalf("read cleared profile: %v", err)
	}
	if len(cleared) != 0 {
		t.Fatalf("cleared profile = %+v", cleared)
	}
}
