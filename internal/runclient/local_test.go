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
