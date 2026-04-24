package runclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/app"
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
	DatabasePath      string
	EmbeddingProvider string
}

// Paths describes the resolved runtime locations on disk.
type Paths struct {
	DataDir      string
	DatabasePath string
	VaultRoot    string
}

// Runtime owns the in-process store used by the runner runtime.
type Runtime struct {
	paths   Paths
	service *app.Service
}

// Close releases the underlying SQLite-backed runtime.
func (r *Runtime) Close() error {
	if r == nil || r.service == nil {
		return nil
	}
	return r.service.Close()
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
	databasePath, err := resolveDatabasePath(cfg)
	if err != nil {
		return Paths{}, err
	}
	databasePath = filepath.Clean(databasePath)
	dataDir := filepath.Dir(databasePath)
	runtimeConfig, err := sqlite.ResolveRuntimeConfig(context.Background(), databasePath, filepath.Join(dataDir, defaultVaultDir))
	if err != nil {
		return Paths{}, err
	}
	return Paths{DataDir: dataDir, DatabasePath: databasePath, VaultRoot: runtimeConfig.VaultRoot}, nil
}

func InitializePaths(cfg Config, vaultRoot string) (Paths, error) {
	databasePath, err := resolveDatabasePath(cfg)
	if err != nil {
		return Paths{}, err
	}
	databasePath = filepath.Clean(databasePath)
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
		return Paths{}, err
	}
	return Paths{DataDir: dataDir, DatabasePath: databasePath, VaultRoot: runtimeConfig.VaultRoot}, nil
}

func newRuntime(backend domain.BackendKind, cfg Config) (*Runtime, error) {
	paths, err := ResolvePaths(cfg)
	if err != nil {
		return nil, err
	}
	store, err := sqlite.New(context.Background(), sqlite.Config{
		Backend:           backend,
		DatabasePath:      paths.DatabasePath,
		VaultRoot:         paths.VaultRoot,
		EmbeddingProvider: cfg.EmbeddingProvider,
	})
	if err != nil {
		return nil, err
	}
	service := app.New(store)
	return &Runtime{
		paths:   paths,
		service: service,
	}, nil
}

func withDefaultEmbeddingProvider(cfg Config) Config {
	if strings.TrimSpace(cfg.EmbeddingProvider) == "" {
		cfg.EmbeddingProvider = "local"
	}
	return cfg
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
