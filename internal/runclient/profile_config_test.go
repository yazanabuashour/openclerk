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

func TestVaultIgnorePathRuntimeConfigRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	initial, err := ReadVaultIgnorePathConfig(ctx, config)
	if err != nil {
		t.Fatalf("read initial vault ignore paths: %v", err)
	}
	if len(initial) != 0 {
		t.Fatalf("initial vault ignore paths = %+v", initial)
	}

	written, err := WriteVaultIgnorePathConfig(ctx, config, []string{"scratch/", `private\drafts`, "scratch"})
	if err != nil {
		t.Fatalf("write vault ignore paths: %v", err)
	}
	want := []string{"scratch", "private/drafts"}
	if len(written) != len(want) {
		t.Fatalf("written vault ignore paths = %+v, want %+v", written, want)
	}
	for i, path := range want {
		if written[i] != path {
			t.Fatalf("written[%d] = %q, want %q; written=%+v", i, written[i], path, written)
		}
	}

	read, err := ReadVaultIgnorePathConfig(ctx, config)
	if err != nil {
		t.Fatalf("read vault ignore paths: %v", err)
	}
	for i, path := range want {
		if read[i] != path {
			t.Fatalf("read[%d] = %q, want %q; read=%+v", i, read[i], path, read)
		}
	}

	if err := ClearVaultIgnorePathConfig(ctx, config); err != nil {
		t.Fatalf("clear vault ignore paths: %v", err)
	}
	cleared, err := ReadVaultIgnorePathConfig(ctx, config)
	if err != nil {
		t.Fatalf("read cleared vault ignore paths: %v", err)
	}
	if len(cleared) != 0 {
		t.Fatalf("cleared vault ignore paths = %+v", cleared)
	}
}
