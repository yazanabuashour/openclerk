package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"io/fs"
	"strings"
)

func sqliteStoreInitialized(ctx context.Context, databasePath string) (bool, error) {
	if _, err := osStat(databasePath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, domain.InternalError("stat sqlite database", err)
	}
	db, err := openSQLiteDatabaseReadMostly(ctx, databasePath)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = db.Close()
	}()
	for _, table := range sqliteSchemaTables() {
		var name string
		err := db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE name = ?`, table).Scan(&name)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		if err != nil {
			return false, domain.InternalError("inspect sqlite schema", err)
		}
	}
	return true, nil
}

func sqliteSchemaTables() []string {
	return []string{
		"runtime_config",
		"documents",
		"document_metadata",
		"chunks",
		"chunk_fts",
		"graph_nodes",
		"graph_edges",
		"record_entities",
		"record_facts",
		"record_citations",
		"service_records",
		"service_facts",
		"service_citations",
		"decision_records",
		"decision_citations",
		"provenance_events",
		"projection_states",
	}
}

func (s *Store) initSchema(ctx context.Context) error {
	statements := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS runtime_config (
			key_name TEXT PRIMARY KEY,
			value_text TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS documents (
			doc_id TEXT PRIMARY KEY,
			path TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			body TEXT NOT NULL,
			headings_json TEXT NOT NULL,
			metadata_json TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS document_metadata (
			doc_id TEXT NOT NULL,
			key_name TEXT NOT NULL,
			value_text TEXT NOT NULL,
			PRIMARY KEY (doc_id, key_name),
			FOREIGN KEY (doc_id) REFERENCES documents(doc_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS chunks (
			chunk_id TEXT PRIMARY KEY,
			doc_id TEXT NOT NULL,
			path TEXT NOT NULL,
			heading TEXT NOT NULL,
			content TEXT NOT NULL,
			line_start INTEGER NOT NULL,
			line_end INTEGER NOT NULL,
			FOREIGN KEY (doc_id) REFERENCES documents(doc_id) ON DELETE CASCADE
		);`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS chunk_fts USING fts5(
			chunk_id UNINDEXED,
			doc_id UNINDEXED,
			path UNINDEXED,
			heading,
			content,
			tokenize = 'unicode61'
		);`,
		`CREATE TABLE IF NOT EXISTS graph_nodes (
			node_id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			label TEXT NOT NULL,
			evidence_doc_id TEXT NOT NULL,
			evidence_chunk_id TEXT NOT NULL,
			evidence_path TEXT NOT NULL,
			evidence_heading TEXT,
			evidence_line_start INTEGER NOT NULL,
			evidence_line_end INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS graph_edges (
			edge_id TEXT PRIMARY KEY,
			from_node_id TEXT NOT NULL,
			to_node_id TEXT NOT NULL,
			kind TEXT NOT NULL,
			evidence_doc_id TEXT NOT NULL,
			evidence_chunk_id TEXT NOT NULL,
			evidence_path TEXT NOT NULL,
			evidence_heading TEXT,
			evidence_line_start INTEGER NOT NULL,
			evidence_line_end INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS record_entities (
			entity_id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			name TEXT NOT NULL,
			summary TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS record_facts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id TEXT NOT NULL,
			key_name TEXT NOT NULL,
			value_text TEXT NOT NULL,
			observed_at TEXT,
			FOREIGN KEY (entity_id) REFERENCES record_entities(entity_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS record_citations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			source_chunk_id TEXT NOT NULL,
			source_path TEXT NOT NULL,
			source_heading TEXT,
			source_line_start INTEGER NOT NULL,
			source_line_end INTEGER NOT NULL,
			FOREIGN KEY (entity_id) REFERENCES record_entities(entity_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS service_records (
			service_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			owner TEXT NOT NULL,
			service_interface TEXT NOT NULL,
			summary TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS service_facts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			service_id TEXT NOT NULL,
			key_name TEXT NOT NULL,
			value_text TEXT NOT NULL,
			observed_at TEXT,
			FOREIGN KEY (service_id) REFERENCES service_records(service_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS service_citations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			service_id TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			source_chunk_id TEXT NOT NULL,
			source_path TEXT NOT NULL,
			source_heading TEXT,
			source_line_start INTEGER NOT NULL,
			source_line_end INTEGER NOT NULL,
			FOREIGN KEY (service_id) REFERENCES service_records(service_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS decision_records (
			decision_id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			scope TEXT NOT NULL,
			owner TEXT NOT NULL,
			decision_date TEXT NOT NULL,
			summary TEXT NOT NULL,
			supersedes TEXT NOT NULL,
			superseded_by TEXT NOT NULL,
			source_refs TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS decision_citations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			decision_id TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			source_chunk_id TEXT NOT NULL,
			source_path TEXT NOT NULL,
			source_heading TEXT,
			source_line_start INTEGER NOT NULL,
			source_line_end INTEGER NOT NULL,
			FOREIGN KEY (decision_id) REFERENCES decision_records(decision_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS provenance_events (
			event_id TEXT PRIMARY KEY,
			event_type TEXT NOT NULL,
			ref_kind TEXT NOT NULL,
			ref_id TEXT NOT NULL,
			source_ref TEXT NOT NULL,
			occurred_at TEXT NOT NULL,
			details_json TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS projection_states (
			projection_name TEXT NOT NULL,
			ref_kind TEXT NOT NULL,
			ref_id TEXT NOT NULL,
			source_ref TEXT NOT NULL,
			freshness TEXT NOT NULL,
			projection_version TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			details_json TEXT NOT NULL,
			PRIMARY KEY (projection_name, ref_kind, ref_id)
		);`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return domain.InternalError("initialize sqlite schema", err)
		}
	}
	if err := ensureColumn(ctx, s.db, "documents", "metadata_json", "TEXT NOT NULL DEFAULT '{}'"); err != nil {
		return err
	}
	return nil
}

func ensureColumn(ctx context.Context, db *sql.DB, table string, column string, definition string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return domain.InternalError("inspect sqlite table", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var (
			cid        int
			name       string
			typ        string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultVal, &primaryKey); err != nil {
			return domain.InternalError("scan sqlite table info", err)
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return domain.InternalError("iterate sqlite table info", err)
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return nil
		}
		return domain.InternalError("alter sqlite table", err)
	}
	return nil
}
