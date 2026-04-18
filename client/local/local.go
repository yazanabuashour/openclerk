package local

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	openclerkclient "github.com/yazanabuashour/openclerk/client/openclerk"
	"github.com/yazanabuashour/openclerk/internal/api"
	"github.com/yazanabuashour/openclerk/internal/app"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/infra/sqlite"
)

const (
	defaultAppDir    = "openclerk"
	defaultDBFile    = "openclerk.sqlite"
	defaultVaultDir  = "vault"
	inProcessBaseURL = "http://openclerk.local"
)

// Config controls where the embedded runtime stores SQLite and canonical markdown data.
type Config struct {
	DataDir           string
	DatabasePath      string
	VaultRoot         string
	EmbeddingProvider string
}

// Paths describes the resolved runtime locations on disk.
type Paths struct {
	DataDir      string
	DatabasePath string
	VaultRoot    string
}

// Runtime owns the in-process store used by an embedded client.
type Runtime struct {
	paths   Paths
	service *app.Service
	handler http.Handler
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

// ResolvePaths returns the effective storage layout for an embedded runtime.
func ResolvePaths(cfg Config) (Paths, error) {
	dataDir, err := resolveDataDir(cfg)
	if err != nil {
		return Paths{}, err
	}
	explicitDataDir := strings.TrimSpace(cfg.DataDir) != ""
	paths := Paths{DataDir: filepath.Clean(dataDir)}
	if strings.TrimSpace(cfg.DatabasePath) != "" {
		paths.DatabasePath = filepath.Clean(cfg.DatabasePath)
	} else if envDatabasePath := strings.TrimSpace(os.Getenv("OPENCLERK_DATABASE_PATH")); !explicitDataDir && envDatabasePath != "" {
		paths.DatabasePath = filepath.Clean(envDatabasePath)
	} else {
		paths.DatabasePath = filepath.Join(paths.DataDir, defaultDBFile)
	}
	if strings.TrimSpace(cfg.VaultRoot) != "" {
		paths.VaultRoot = filepath.Clean(cfg.VaultRoot)
	} else if envVaultRoot := strings.TrimSpace(os.Getenv("OPENCLERK_VAULT_ROOT")); !explicitDataDir && envVaultRoot != "" {
		paths.VaultRoot = filepath.Clean(envVaultRoot)
	} else {
		paths.VaultRoot = filepath.Join(paths.DataDir, defaultVaultDir)
	}
	return paths, nil
}

// Open creates the primary embedded OpenClerk client without binding a local port.
func Open(cfg Config) (*openclerkclient.ClientWithResponses, *Runtime, error) {
	runtime, err := newRuntime(domain.BackendOpenClerk, withDefaultEmbeddingProvider(cfg))
	if err != nil {
		return nil, nil, err
	}
	client, err := openclerkclient.NewClientWithResponses(inProcessBaseURL, openclerkclient.WithHTTPClient(handlerDoer{handler: runtime.handler}))
	if err != nil {
		_ = runtime.Close()
		return nil, nil, err
	}
	return client, runtime, nil
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
		handler: api.NewHandler(service),
	}, nil
}

func withDefaultEmbeddingProvider(cfg Config) Config {
	if strings.TrimSpace(cfg.EmbeddingProvider) == "" {
		cfg.EmbeddingProvider = "local"
	}
	return cfg
}

func resolveDataDir(cfg Config) (string, error) {
	switch {
	case strings.TrimSpace(cfg.DataDir) != "":
		return cfg.DataDir, nil
	case strings.TrimSpace(cfg.DatabasePath) != "":
		return filepath.Dir(cfg.DatabasePath), nil
	case strings.TrimSpace(cfg.VaultRoot) != "":
		return filepath.Dir(cfg.VaultRoot), nil
	case strings.TrimSpace(os.Getenv("OPENCLERK_DATA_DIR")) != "":
		return os.Getenv("OPENCLERK_DATA_DIR"), nil
	case strings.TrimSpace(os.Getenv("OPENCLERK_DATABASE_PATH")) != "":
		return filepath.Dir(os.Getenv("OPENCLERK_DATABASE_PATH")), nil
	case strings.TrimSpace(os.Getenv("OPENCLERK_VAULT_ROOT")) != "":
		return filepath.Dir(os.Getenv("OPENCLERK_VAULT_ROOT")), nil
	default:
		return defaultDataDir()
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

type handlerDoer struct {
	handler http.Handler
}

func (d handlerDoer) Do(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	d.handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}
