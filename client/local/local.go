package local

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	ftsclient "github.com/yazanabuashour/openclerk/client/fts"
	graphclient "github.com/yazanabuashour/openclerk/client/graph"
	hybridclient "github.com/yazanabuashour/openclerk/client/hybrid"
	openclerkclient "github.com/yazanabuashour/openclerk/client/openclerk"
	recordsclient "github.com/yazanabuashour/openclerk/client/records"
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
	paths := Paths{DataDir: filepath.Clean(dataDir)}
	if strings.TrimSpace(cfg.DatabasePath) != "" {
		paths.DatabasePath = filepath.Clean(cfg.DatabasePath)
	} else {
		paths.DatabasePath = filepath.Join(paths.DataDir, defaultDBFile)
	}
	if strings.TrimSpace(cfg.VaultRoot) != "" {
		paths.VaultRoot = filepath.Clean(cfg.VaultRoot)
	} else {
		paths.VaultRoot = filepath.Join(paths.DataDir, defaultVaultDir)
	}
	return paths, nil
}

// OpenFTS creates an embedded FTS client without binding a local port.
func OpenFTS(cfg Config) (*ftsclient.ClientWithResponses, *Runtime, error) {
	runtime, err := newRuntime(domain.BackendFTS, cfg)
	if err != nil {
		return nil, nil, err
	}
	client, err := ftsclient.NewClientWithResponses(inProcessBaseURL, ftsclient.WithHTTPClient(handlerDoer{handler: runtime.handler}))
	if err != nil {
		_ = runtime.Close()
		return nil, nil, err
	}
	return client, runtime, nil
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

// OpenHybrid creates an embedded hybrid client without binding a local port.
func OpenHybrid(cfg Config) (*hybridclient.ClientWithResponses, *Runtime, error) {
	runtime, err := newRuntime(domain.BackendHybrid, cfg)
	if err != nil {
		return nil, nil, err
	}
	client, err := hybridclient.NewClientWithResponses(inProcessBaseURL, hybridclient.WithHTTPClient(handlerDoer{handler: runtime.handler}))
	if err != nil {
		_ = runtime.Close()
		return nil, nil, err
	}
	return client, runtime, nil
}

// OpenGraph creates an embedded graph client without binding a local port.
func OpenGraph(cfg Config) (*graphclient.ClientWithResponses, *Runtime, error) {
	runtime, err := newRuntime(domain.BackendGraph, cfg)
	if err != nil {
		return nil, nil, err
	}
	client, err := graphclient.NewClientWithResponses(inProcessBaseURL, graphclient.WithHTTPClient(handlerDoer{handler: runtime.handler}))
	if err != nil {
		_ = runtime.Close()
		return nil, nil, err
	}
	return client, runtime, nil
}

// OpenRecords creates an embedded records client without binding a local port.
func OpenRecords(cfg Config) (*recordsclient.ClientWithResponses, *Runtime, error) {
	runtime, err := newRuntime(domain.BackendRecords, cfg)
	if err != nil {
		return nil, nil, err
	}
	client, err := recordsclient.NewClientWithResponses(inProcessBaseURL, recordsclient.WithHTTPClient(handlerDoer{handler: runtime.handler}))
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
