package sqlite

import (
	"context"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestParallelRuntimeConfigInitialization(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	defaultVault := filepath.Join(filepath.Dir(dbPath), "vault")

	var wg sync.WaitGroup
	errs := make(chan error, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			config, err := ResolveRuntimeConfig(ctx, dbPath, defaultVault)
			if err != nil {
				errs <- err
				return
			}
			if config.VaultRoot != defaultVault || config.LayoutConventionVersion != defaultLayoutConventionVersion {
				errs <- fmt.Errorf("runtime config = %+v", config)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "sqlite") ||
				strings.Contains(strings.ToLower(err.Error()), "runtime config") ||
				strings.Contains(strings.ToLower(err.Error()), "upsert") {
				t.Fatalf("raw runtime config failure leaked: %v", err)
			}
			t.Fatalf("runtime config initialization failed: %v", err)
		}
	}
}

func TestParallelReadOnlyStoreStartup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	_, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "docs/parallel.md",
		Title: "Parallel",
		Body:  "# Parallel\n\n## Summary\nSafe read-only store startup.\n",
	})
	if err != nil {
		t.Fatalf("seed document: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close seed store: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			readStore, err := NewReadOnly(ctx, Config{
				Backend:      domain.BackendOpenClerk,
				DatabasePath: dbPath,
				VaultRoot:    vaultRoot,
			})
			if err != nil {
				errs <- err
				return
			}
			defer func() {
				_ = readStore.Close()
			}()
			result, err := readStore.ListDocuments(ctx, domain.DocumentListQuery{PathPrefix: "docs/", Limit: 10})
			if err != nil {
				errs <- err
				return
			}
			if len(result.Documents) != 1 || result.Documents[0].Path != "docs/parallel.md" {
				errs <- fmt.Errorf("documents = %+v", result.Documents)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("parallel read-only startup failed: %v", err)
		}
	}
}

func TestReadOnlyStoreRepairsPartialSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	db, err := openSQLiteDatabase(ctx, dbPath)
	if err != nil {
		t.Fatalf("open partial database: %v", err)
	}
	for _, statement := range []string{
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
		`CREATE TABLE IF NOT EXISTS chunks (
			chunk_id TEXT PRIMARY KEY,
			doc_id TEXT NOT NULL,
			path TEXT NOT NULL,
			heading TEXT NOT NULL,
			content TEXT NOT NULL,
			line_start INTEGER NOT NULL,
			line_end INTEGER NOT NULL
		);`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS chunk_fts USING fts5(
			chunk_id UNINDEXED,
			doc_id UNINDEXED,
			path UNINDEXED,
			heading,
			content,
			tokenize = 'unicode61'
		);`,
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("create partial schema: %v", err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close partial database: %v", err)
	}

	store, err := NewReadOnly(ctx, Config{
		Backend:      domain.BackendOpenClerk,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("open read-only store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()
	if _, err := store.ListProjectionStates(ctx, domain.ProjectionStateQuery{Limit: 10}); err != nil {
		t.Fatalf("list projection states after read-only startup: %v", err)
	}
}
