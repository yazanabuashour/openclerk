package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Backend             domain.BackendKind
	DatabasePath        string
	VaultRoot           string
	SyncDiagnosticsPath string
}

type RuntimeConfig struct {
	VaultRoot               string
	LayoutConventionVersion string
}

var runtimeConfigInitializationMu sync.Mutex

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
	runtimeConfigInitializationMu.Lock()
	defer runtimeConfigInitializationMu.Unlock()

	if err := ensureDir(filepath.Dir(databasePath)); err != nil {
		return RuntimeConfig{}, domain.InternalError("create database directory", err)
	}
	_, preexistingDBErr := os.Stat(databasePath)
	preexistingDB := preexistingDBErr == nil
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
	} else if preexistingDB {
		hasData, err := databaseHasIndexedDocuments(ctx, db)
		if err != nil {
			return RuntimeConfig{}, err
		}
		if hasData {
			return RuntimeConfig{}, domain.ValidationError("existing OpenClerk database is missing vault root binding; run openclerk init --vault-root intentionally before use", nil)
		}
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

func databaseHasIndexedDocuments(ctx context.Context, db *sql.DB) (bool, error) {
	var tableName string
	err := db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'documents'`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, domain.InternalError("inspect existing OpenClerk schema", err)
	}
	var count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents LIMIT 1`).Scan(&count)
	if err != nil {
		return false, domain.InternalError("inspect existing OpenClerk documents", err)
	}
	return count > 0, nil
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
	_ = os.Chmod(databasePath, 0o600)
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

func upsertRuntimeConfigValueTx(ctx context.Context, tx *sql.Tx, key string, value string, updatedAt string) error {
	if _, err := tx.ExecContext(ctx, `
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

func configuredVaultIgnorePaths(ctx context.Context, db *sql.DB) ([]string, error) {
	value, err := runtimeConfigValue(ctx, db, configKeyVaultIgnorePaths)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var paths []string
	if err := json.Unmarshal([]byte(value), &paths); err != nil {
		return nil, domain.InternalError("decode vault ignore paths config", err)
	}
	normalized, err := domain.NormalizeVaultIgnorePaths(paths)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}
