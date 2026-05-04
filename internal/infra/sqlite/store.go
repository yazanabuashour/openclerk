package sqlite

import (
	"context"
	"database/sql"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"path/filepath"
	"time"
)

type Store struct {
	db                  *sql.DB
	backend             domain.BackendKind
	vaultRoot           string
	syncDiagnosticsPath string
	now                 func() time.Time
}

func New(ctx context.Context, cfg Config) (*Store, error) {
	return newStore(ctx, cfg, true)
}

func NewUnsynced(ctx context.Context, cfg Config) (*Store, error) {
	return newStore(ctx, cfg, false)
}

func NewReadOnly(ctx context.Context, cfg Config) (*Store, error) {
	return newReadOnlyStore(ctx, cfg)
}

func newStore(ctx context.Context, cfg Config, syncVault bool) (*Store, error) {
	if cfg.DatabasePath == "" {
		return nil, domain.ValidationError("database path is required", nil)
	}
	if cfg.VaultRoot == "" {
		return nil, domain.ValidationError("vault root is required", nil)
	}
	if err := ensureDir(cfg.VaultRoot); err != nil {
		return nil, domain.InternalError("create vault root", err)
	}
	if err := ensureDir(filepath.Dir(cfg.DatabasePath)); err != nil {
		return nil, domain.InternalError("create database directory", err)
	}

	db, err := openSQLiteDatabase(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	store := &Store{
		db:                  db,
		backend:             cfg.Backend,
		vaultRoot:           cfg.VaultRoot,
		syncDiagnosticsPath: cfg.SyncDiagnosticsPath,
		now:                 time.Now,
	}
	if err := store.initSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := insertRuntimeConfigValueIfAbsent(ctx, db, configKeyVaultRoot, filepath.Clean(cfg.VaultRoot), now); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := insertRuntimeConfigValueIfAbsent(ctx, db, configKeyLayoutConventionVersion, defaultLayoutConventionVersion, now); err != nil {
		_ = db.Close()
		return nil, err
	}
	if syncVault {
		if err := store.syncVault(ctx); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	return store, nil
}

func newReadOnlyStore(ctx context.Context, cfg Config) (*Store, error) {
	if cfg.DatabasePath == "" {
		return nil, domain.ValidationError("database path is required", nil)
	}
	if cfg.VaultRoot == "" {
		return nil, domain.ValidationError("vault root is required", nil)
	}
	if initialized, err := sqliteStoreInitialized(ctx, cfg.DatabasePath); err != nil {
		return nil, err
	} else if !initialized {
		return newStore(ctx, cfg, false)
	}
	db, err := openSQLiteDatabaseReadMostly(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	return &Store{
		db:        db,
		backend:   cfg.Backend,
		vaultRoot: cfg.VaultRoot,
		now:       time.Now,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Capabilities(_ context.Context) (domain.Capabilities, error) {
	capabilities := domain.Capabilities{
		Backend:     s.backend,
		AuthMode:    "none",
		SearchModes: []string{"lexical"},
		Extensions:  []string{"provenance"},
	}
	if supportsGraph(s.backend) {
		capabilities.Extensions = append(capabilities.Extensions, "graph")
	}
	if supportsRecords(s.backend) {
		capabilities.Extensions = append(capabilities.Extensions, "records")
	}
	if supportsServices(s.backend) {
		capabilities.Extensions = append(capabilities.Extensions, "services")
	}
	if supportsDecisions(s.backend) {
		capabilities.Extensions = append(capabilities.Extensions, "decisions")
	}
	return capabilities, nil
}
