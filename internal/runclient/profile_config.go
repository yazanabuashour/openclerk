package runclient

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

const defaultProfileRuntimeConfigPrefix = "profile.default."
const vaultIgnorePathsRuntimeConfigKey = "vault_ignore_paths"

func ReadDefaultProfileConfig(ctx context.Context, cfg Config) (map[string]string, error) {
	values := map[string]string{}
	err := withRuntimeConfigDB(ctx, cfg, false, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT key_name, value_text FROM runtime_config WHERE key_name LIKE ?`, defaultProfileRuntimeConfigPrefix+"%")
		if err != nil {
			return domain.InternalError("read profile runtime config", err)
		}
		defer func() {
			_ = rows.Close()
		}()
		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err != nil {
				return domain.InternalError("scan profile runtime config", err)
			}
			values[strings.TrimPrefix(key, defaultProfileRuntimeConfigPrefix)] = value
		}
		if err := rows.Err(); err != nil {
			return domain.InternalError("iterate profile runtime config", err)
		}
		return nil
	})
	return values, err
}

func WriteDefaultProfileConfig(ctx context.Context, cfg Config, values map[string]string) error {
	return withRuntimeConfigDB(ctx, cfg, true, func(db *sql.DB) error {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		for key, value := range values {
			if _, err := db.ExecContext(ctx, `
INSERT INTO runtime_config (key_name, value_text, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key_name) DO UPDATE SET
	value_text = excluded.value_text,
	updated_at = excluded.updated_at`, defaultProfileRuntimeConfigPrefix+key, value, now); err != nil {
				return domain.InternalError("write profile runtime config", err)
			}
		}
		return nil
	})
}

func ClearDefaultProfileConfig(ctx context.Context, cfg Config) error {
	return withRuntimeConfigDB(ctx, cfg, true, func(db *sql.DB) error {
		if _, err := db.ExecContext(ctx, `DELETE FROM runtime_config WHERE key_name LIKE ?`, defaultProfileRuntimeConfigPrefix+"%"); err != nil {
			return domain.InternalError("clear profile runtime config", err)
		}
		return nil
	})
}

func ReadVaultIgnorePathConfig(ctx context.Context, cfg Config) ([]string, error) {
	var raw string
	err := withRuntimeConfigDB(ctx, cfg, false, func(db *sql.DB) error {
		err := db.QueryRowContext(ctx, `SELECT value_text FROM runtime_config WHERE key_name = ?`, vaultIgnorePathsRuntimeConfigKey).Scan(&raw)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return domain.InternalError("read vault ignore paths config", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var paths []string
	if err := json.Unmarshal([]byte(raw), &paths); err != nil {
		return nil, domain.InternalError("decode vault ignore paths config", err)
	}
	return domain.NormalizeVaultIgnorePaths(paths)
}

func WriteVaultIgnorePathConfig(ctx context.Context, cfg Config, paths []string) ([]string, error) {
	normalized, err := domain.NormalizeVaultIgnorePaths(paths)
	if err != nil {
		return nil, err
	}
	content, err := json.Marshal(normalized)
	if err != nil {
		return nil, domain.InternalError("encode vault ignore paths config", err)
	}
	err = withRuntimeConfigDB(ctx, cfg, true, func(db *sql.DB) error {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := db.ExecContext(ctx, `
INSERT INTO runtime_config (key_name, value_text, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key_name) DO UPDATE SET
	value_text = excluded.value_text,
	updated_at = excluded.updated_at`, vaultIgnorePathsRuntimeConfigKey, string(content), now); err != nil {
			return domain.InternalError("write vault ignore paths config", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return normalized, nil
}
