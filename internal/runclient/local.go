package runclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/infra/sqlite"
)

const (
	defaultAppDir   = "openclerk"
	defaultDBFile   = "openclerk.sqlite"
	defaultVaultDir = "vault"
)

// Config controls where the internal runtime stores SQLite-backed OpenClerk data.
type Config struct {
	DatabasePath        string
	SyncDiagnosticsPath string
	GitCheckpoints      bool
}

// Paths describes the resolved runtime locations on disk.
type Paths struct {
	DatabasePath string
	VaultRoot    string
}

// Runtime owns the in-process store used by the runner runtime.
type Runtime struct {
	paths Paths
	store domain.Store
}

// Close releases the underlying SQLite-backed runtime.
func (r *Runtime) Close() error {
	if r == nil || r.store == nil {
		return nil
	}
	return r.store.Close()
}

// Paths returns the resolved storage locations for this runtime.
func (r *Runtime) Paths() Paths {
	if r == nil {
		return Paths{}
	}
	return r.paths
}

// ResolvePaths returns the effective storage layout for an internal runtime.
func ResolvePaths(cfg Config) (Paths, error) {
	return resolvePaths(cfg, true)
}

func resolvePaths(cfg Config, lock bool) (Paths, error) {
	databasePath, err := resolveDatabasePath(cfg)
	if err != nil {
		return Paths{}, err
	}
	databasePath = filepath.Clean(databasePath)
	if lock {
		unlock, err := acquireRuntimeConfigLock(context.Background(), databasePath)
		if err != nil {
			return Paths{}, err
		}
		defer unlock()
	}
	runtimeConfig, err := sqlite.ResolveRuntimeConfig(context.Background(), databasePath, filepath.Join(filepath.Dir(databasePath), defaultVaultDir))
	if err != nil {
		return Paths{}, decorateRuntimeConfigError("resolve", databasePath, err)
	}
	return Paths{DatabasePath: databasePath, VaultRoot: runtimeConfig.VaultRoot}, nil
}

func InitializePaths(cfg Config, vaultRoot string) (Paths, error) {
	databasePath, err := resolveDatabasePath(cfg)
	if err != nil {
		return Paths{}, err
	}
	databasePath = filepath.Clean(databasePath)
	unlock, err := acquireRunnerWriteLock(context.Background(), databasePath)
	if err != nil {
		return Paths{}, err
	}
	defer unlock()
	configUnlock, err := acquireRuntimeConfigLock(context.Background(), databasePath)
	if err != nil {
		return Paths{}, err
	}
	defer configUnlock()

	dataDir := filepath.Dir(databasePath)
	if strings.TrimSpace(vaultRoot) != "" {
		absVaultRoot, err := filepath.Abs(vaultRoot)
		if err != nil {
			return Paths{}, fmt.Errorf("resolve vault root: %w", err)
		}
		vaultRoot = absVaultRoot
	}
	runtimeConfig, err := sqlite.InitializeRuntimeConfig(context.Background(), databasePath, vaultRoot, filepath.Join(dataDir, defaultVaultDir))
	if err != nil {
		return Paths{}, decorateRuntimeConfigError("initialize", databasePath, err)
	}
	return Paths{DatabasePath: databasePath, VaultRoot: runtimeConfig.VaultRoot}, nil
}

func decorateRuntimeConfigError(operation string, databasePath string, err error) error {
	if err == nil {
		return nil
	}
	if strings.TrimSpace(databasePath) == "" {
		return err
	}
	message := fmt.Sprintf(
		"%s OpenClerk configuration for database %q: run {\"action\":\"resolve_paths\"} or {\"action\":\"inspect_layout\"} with openclerk document before init; use openclerk init --vault-root <vault-root> only for first-time binding or intentional rebinding",
		operation,
		databasePath,
	)
	return domain.InternalError(message, err)
}

func newRuntime(backend domain.BackendKind, cfg Config) (*Runtime, error) {
	return newRuntimeWithMode(backend, cfg, runtimeOpenSync)
}

func newWriteRuntime(backend domain.BackendKind, cfg Config) (*Runtime, error) {
	return newRuntimeWithMode(backend, cfg, runtimeOpenUnsynced)
}

func newReadOnlyRuntime(backend domain.BackendKind, cfg Config) (*Runtime, error) {
	return newRuntimeWithMode(backend, cfg, runtimeOpenReadOnly)
}

type runtimeOpenMode int

const (
	runtimeOpenSync runtimeOpenMode = iota
	runtimeOpenUnsynced
	runtimeOpenReadOnly
)

func newRuntimeWithMode(backend domain.BackendKind, cfg Config, mode runtimeOpenMode) (*Runtime, error) {
	paths, err := resolvePaths(cfg, true)
	if err != nil {
		return nil, err
	}
	sqliteConfig := sqlite.Config{
		Backend:             backend,
		DatabasePath:        paths.DatabasePath,
		VaultRoot:           paths.VaultRoot,
		SyncDiagnosticsPath: cfg.SyncDiagnosticsPath,
	}
	var store *sqlite.Store
	switch mode {
	case runtimeOpenSync:
		store, err = sqlite.New(context.Background(), sqliteConfig)
	case runtimeOpenUnsynced:
		store, err = sqlite.NewUnsynced(context.Background(), sqliteConfig)
	case runtimeOpenReadOnly:
		store, err = sqlite.NewReadOnly(context.Background(), sqliteConfig)
	default:
		return nil, domain.InternalError("open runtime", fmt.Errorf("unsupported runtime open mode %d", mode))
	}
	if err != nil {
		return nil, err
	}
	return &Runtime{
		paths: paths,
		store: store,
	}, nil
}

func WithWriteLock(ctx context.Context, cfg Config, fn func() error) error {
	databasePath, err := resolveDatabasePath(cfg)
	if err != nil {
		return err
	}
	databasePath = filepath.Clean(databasePath)
	unlock, err := acquireRunnerWriteLock(ctx, databasePath)
	if err != nil {
		return err
	}
	defer unlock()
	return fn()
}

func acquireRunnerWriteLock(ctx context.Context, databasePath string) (func(), error) {
	if err := ensureLocalDir(filepath.Dir(databasePath)); err != nil {
		return nil, err
	}
	lockPath := databasePath + ".runner-write.lock"
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = fmt.Fprintf(file, "pid=%d\n", os.Getpid())
			_ = file.Close()
			return func() {
				_ = os.Remove(lockPath)
			}, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, domain.InternalError("acquire runner write lock", err)
		}
		removed, err := removeStaleRunnerWriteLock(lockPath)
		if err != nil {
			return nil, err
		}
		if removed {
			continue
		}
		select {
		case <-ctx.Done():
			return nil, domain.ConflictError("runner write lock was not acquired before the context ended", map[string]any{"database_path": databasePath})
		case <-deadline.C:
			return nil, domain.ConflictError("runner write lock is held by another command", map[string]any{"database_path": databasePath})
		case <-ticker.C:
		}
	}
}

func acquireRuntimeConfigLock(ctx context.Context, databasePath string) (func(), error) {
	if err := ensureLocalDir(filepath.Dir(databasePath)); err != nil {
		return nil, err
	}
	lockPath := databasePath + ".runtime-config.lock"
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = fmt.Fprintf(file, "pid=%d\n", os.Getpid())
			_ = file.Close()
			return func() {
				_ = os.Remove(lockPath)
			}, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, domain.InternalError("acquire runtime config lock", err)
		}
		removed, err := removeStaleRunnerWriteLock(lockPath)
		if err != nil {
			return nil, err
		}
		if removed {
			continue
		}
		select {
		case <-ctx.Done():
			return nil, domain.ConflictError("runtime config lock was not acquired before the context ended", map[string]any{"database_path": databasePath})
		case <-deadline.C:
			return nil, domain.ConflictError("runtime config lock is held by another command", map[string]any{"database_path": databasePath})
		case <-ticker.C:
		}
	}
}

func removeStaleRunnerWriteLock(lockPath string) (bool, error) {
	pid, ok, err := runnerWriteLockPID(lockPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, domain.InternalError("inspect runner write lock", err)
	}
	if !ok || runnerProcessAlive(pid) {
		return false, nil
	}
	if err := os.Remove(lockPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, domain.InternalError("remove stale runner write lock", err)
	}
	return true, nil
}

func runnerWriteLockPID(lockPath string) (int, bool, error) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return 0, false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || key != "pid" {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || pid <= 0 {
			return 0, false, nil
		}
		return pid, true, nil
	}
	return 0, false, nil
}

func runnerProcessAlive(pid int) bool {
	if pid == os.Getpid() {
		return true
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	defer func() {
		_ = process.Release()
	}()
	err = process.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

func ensureLocalDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return domain.InternalError("create database directory", err)
	}
	return nil
}

func resolveDatabasePath(cfg Config) (string, error) {
	switch {
	case strings.TrimSpace(cfg.DatabasePath) != "":
		return cfg.DatabasePath, nil
	case strings.TrimSpace(os.Getenv("OPENCLERK_DATABASE_PATH")) != "":
		return os.Getenv("OPENCLERK_DATABASE_PATH"), nil
	default:
		dataDir, err := defaultDataDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dataDir, defaultDBFile), nil
	}
}

func defaultDataDir() (string, error) {
	xdgDataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
	if xdgDataHome != "" && filepath.IsAbs(xdgDataHome) {
		return filepath.Join(xdgDataHome, defaultAppDir), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(homeDir, ".local", "share", defaultAppDir), nil
}
