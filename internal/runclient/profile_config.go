package runclient

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

const defaultProfileRuntimeConfigPrefix = "profile.default."

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
