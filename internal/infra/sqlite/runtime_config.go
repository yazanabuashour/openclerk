package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
	_ "modernc.org/sqlite"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Backend      domain.BackendKind
	DatabasePath string
	VaultRoot    string
}

type RuntimeConfig struct {
	VaultRoot               string
	LayoutConventionVersion string
}

func ResolveRuntimeConfig(ctx context.Context, databasePath string, defaultVaultRoot string) (RuntimeConfig, error) {
	return configureRuntime(ctx, databasePath, "", defaultVaultRoot)
}

func InitializeRuntimeConfig(ctx context.Context, databasePath string, vaultRoot string, defaultVaultRoot string) (RuntimeConfig, error) {
	return configureRuntime(ctx, databasePath, vaultRoot, defaultVaultRoot)
}

func configureRuntime(ctx context.Context, databasePath string, vaultRoot string, defaultVaultRoot string) (RuntimeConfig, error) {
	if strings.TrimSpace(databasePath) == "" {
		return RuntimeConfig{}, domain.ValidationError("database path is required", nil)
	}
	if strings.TrimSpace(defaultVaultRoot) == "" {
		return RuntimeConfig{}, domain.ValidationError("default vault root is required", nil)
	}
	if err := ensureDir(filepath.Dir(databasePath)); err != nil {
		return RuntimeConfig{}, domain.InternalError("create database directory", err)
	}
	db, err := openSQLiteDatabase(ctx, databasePath)
	if err != nil {
		return RuntimeConfig{}, err
	}
	defer func() {
		_ = db.Close()
	}()
	if err := initRuntimeConfigSchema(ctx, db); err != nil {
		return RuntimeConfig{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	resolvedVaultRoot := filepath.Clean(defaultVaultRoot)
	explicitVaultRoot := strings.TrimSpace(vaultRoot) != ""
	if explicitVaultRoot {
		resolvedVaultRoot = filepath.Clean(vaultRoot)
	} else if stored, err := runtimeConfigValue(ctx, db, configKeyVaultRoot); err != nil {
		return RuntimeConfig{}, err
	} else if stored != "" {
		resolvedVaultRoot = filepath.Clean(stored)
	}
	if err := ensureDir(resolvedVaultRoot); err != nil {
		return RuntimeConfig{}, domain.InternalError("create vault root", err)
	}
	if explicitVaultRoot {
		if err := upsertRuntimeConfigValue(ctx, db, configKeyVaultRoot, resolvedVaultRoot, now); err != nil {
			return RuntimeConfig{}, err
		}
	} else if err := insertRuntimeConfigValueIfAbsent(ctx, db, configKeyVaultRoot, resolvedVaultRoot, now); err != nil {
		return RuntimeConfig{}, err
	}

	if err := insertRuntimeConfigValueIfAbsent(ctx, db, configKeyLayoutConventionVersion, defaultLayoutConventionVersion, now); err != nil {
		return RuntimeConfig{}, err
	}
	layoutVersion, err := runtimeConfigValue(ctx, db, configKeyLayoutConventionVersion)
	if err != nil {
		return RuntimeConfig{}, err
	}
	if strings.TrimSpace(layoutVersion) == "" {
		layoutVersion = defaultLayoutConventionVersion
	}
	return RuntimeConfig{
		VaultRoot:               resolvedVaultRoot,
		LayoutConventionVersion: layoutVersion,
	}, nil
}

func openSQLiteDatabase(ctx context.Context, databasePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, domain.InternalError("open sqlite database", err)
	}
	db.SetMaxOpenConns(1)
	if err := configureSQLiteConnection(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func openSQLiteDatabaseReadMostly(ctx context.Context, databasePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, domain.InternalError("open sqlite database", err)
	}
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA foreign_keys = ON;`,
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			_ = db.Close()
			return nil, domain.InternalError("configure sqlite connection", err)
		}
	}
	return db, nil
}

func configureSQLiteConnection(ctx context.Context, db *sql.DB) error {
	for _, statement := range []string{
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA journal_mode = WAL;`,
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return domain.InternalError("configure sqlite connection", err)
		}
	}
	return nil
}

func initRuntimeConfigSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS runtime_config (
		key_name TEXT PRIMARY KEY,
		value_text TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);`); err != nil {
		return domain.InternalError("initialize runtime config schema", err)
	}
	return nil
}

func runtimeConfigValue(ctx context.Context, db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRowContext(ctx, `SELECT value_text FROM runtime_config WHERE key_name = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", domain.InternalError("query runtime config", err)
	}
	return value, nil
}

func upsertRuntimeConfigValue(ctx context.Context, db *sql.DB, key string, value string, updatedAt string) error {
	if _, err := db.ExecContext(ctx, `
INSERT INTO runtime_config (key_name, value_text, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key_name) DO UPDATE SET
	value_text = excluded.value_text,
	updated_at = excluded.updated_at`, key, value, updatedAt); err != nil {
		return domain.InternalError("upsert runtime config", err)
	}
	return nil
}

func insertRuntimeConfigValueIfAbsent(ctx context.Context, db *sql.DB, key string, value string, updatedAt string) error {
	if _, err := db.ExecContext(ctx, `
INSERT OR IGNORE INTO runtime_config (key_name, value_text, updated_at)
VALUES (?, ?, ?)`, key, value, updatedAt); err != nil {
		return domain.InternalError("initialize runtime config", err)
	}
	return nil
}
