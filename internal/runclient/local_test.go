package runclient

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAcquireRunnerWriteLockRecoversStalePIDFile(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	lockPath := dbPath + ".runner-write.lock"
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("create lock dir: %v", err)
	}
	if err := os.WriteFile(lockPath, []byte("pid=99999999\n"), 0o644); err != nil {
		t.Fatalf("write stale lock: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock, err := acquireRunnerWriteLock(ctx, dbPath)
	if err != nil {
		t.Fatalf("acquire lock after stale pid: %v", err)
	}
	unlock()
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("lock file after unlock err = %v, want not exist", err)
	}
}

func TestResolvePathsWithSourceReportsDatabaseSource(t *testing.T) {
	explicitDB := filepath.Join(t.TempDir(), "explicit", "openclerk.sqlite")
	explicit, err := ResolvePathsWithSource(Config{DatabasePath: explicitDB})
	if err != nil {
		t.Fatalf("resolve explicit paths: %v", err)
	}
	if explicit.DatabasePath != explicitDB || explicit.DatabaseSource != "flag" {
		t.Fatalf("explicit paths = %+v, want db %q source flag", explicit, explicitDB)
	}

	envDB := filepath.Join(t.TempDir(), "env", "openclerk.sqlite")
	t.Setenv("OPENCLERK_DATABASE_PATH", envDB)
	envResolved, err := ResolvePathsWithSource(Config{})
	if err != nil {
		t.Fatalf("resolve env paths: %v", err)
	}
	if envResolved.DatabasePath != envDB || envResolved.DatabaseSource != "env" {
		t.Fatalf("env paths = %+v, want db %q source env", envResolved, envDB)
	}

	t.Setenv("OPENCLERK_DATABASE_PATH", "")
	xdgDataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdgDataHome)
	defaultResolved, err := ResolvePathsWithSource(Config{})
	if err != nil {
		t.Fatalf("resolve default paths: %v", err)
	}
	wantDefaultDB := filepath.Join(xdgDataHome, defaultAppDir, defaultDBFile)
	if defaultResolved.DatabasePath != wantDefaultDB || defaultResolved.DatabaseSource != "default" {
		t.Fatalf("default paths = %+v, want db %q source default", defaultResolved, wantDefaultDB)
	}
}
