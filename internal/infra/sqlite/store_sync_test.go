package sqlite

import (
	"context"
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncVaultPrunesDeletedDocuments(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	docPath := filepath.Join(vaultRoot, "docs", "widget.md")
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := os.WriteFile(docPath, []byte("# Widget\n\nalpha signal\n"), 0o644); err != nil {
		t.Fatalf("write vault doc: %v", err)
	}

	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	search, err := store.Search(context.Background(), domain.SearchQuery{Text: "alpha", Limit: 10})
	if err != nil {
		t.Fatalf("search before delete: %v", err)
	}
	if len(search.Hits) != 1 {
		t.Fatalf("search before delete hit count = %d, want 1", len(search.Hits))
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	if err := os.Remove(docPath); err != nil {
		t.Fatalf("remove vault doc: %v", err)
	}

	reopened := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = reopened.Close()
	}()

	search, err = reopened.Search(context.Background(), domain.SearchQuery{Text: "alpha", Limit: 10})
	if err != nil {
		t.Fatalf("search after delete: %v", err)
	}
	if len(search.Hits) != 0 {
		t.Fatalf("search after delete hit count = %d, want 0", len(search.Hits))
	}

	_, err = reopened.GetDocument(context.Background(), docIDForPath("docs/widget.md"))
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 404 {
		t.Fatalf("get deleted document error = %v, want not found 404", err)
	}
}

func TestSyncVaultPrunesDeletedServiceProjection(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	if _, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "records/services/openclerk-runner.md",
		Title: "OpenClerk runner",
		Body: strings.TrimSpace(`---
service_id: openclerk-runner
service_name: OpenClerk runner
service_status: active
service_owner: runner
service_interface: JSON runner
---
# OpenClerk runner

## Summary
Production service.
`) + "\n",
	}); err != nil {
		t.Fatalf("create service document: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	if err := os.Remove(filepath.Join(vaultRoot, "records", "services", "openclerk-runner.md")); err != nil {
		t.Fatalf("remove service doc: %v", err)
	}

	reopened := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = reopened.Close()
	}()

	services, err := reopened.ServicesLookup(context.Background(), domain.ServiceLookupInput{Text: "OpenClerk runner", Limit: 10})
	if err != nil {
		t.Fatalf("services lookup after delete: %v", err)
	}
	if len(services.Services) != 0 {
		t.Fatalf("services after delete = %+v, want none", services.Services)
	}
}
