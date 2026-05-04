package sqlite

import (
	"context"
	"encoding/json"
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

func TestSyncVaultDiagnosticsSkipUnchangedDocumentsAndProjectionRebuild(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	writeVaultFile(t, vaultRoot, "sources/alpha.md", "# Alpha\n\nalpha signal\n")
	writeVaultFile(t, vaultRoot, "synthesis/alpha.md", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/alpha.md
---
# Alpha Synthesis

## Summary
alpha signal synthesis
`)+"\n")

	initialDiagnosticsPath := filepath.Join(t.TempDir(), "initial-sync.json")
	store, err := New(context.Background(), Config{
		Backend:             domain.BackendOpenClerk,
		DatabasePath:        dbPath,
		VaultRoot:           vaultRoot,
		SyncDiagnosticsPath: initialDiagnosticsPath,
	})
	if err != nil {
		t.Fatalf("open initial store: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}
	initialDiagnostics := readSyncDiagnosticsForTest(t, initialDiagnosticsPath)
	if initialDiagnostics.DocumentsCreated != 2 || initialDiagnostics.DocumentsUnchanged != 0 {
		t.Fatalf("initial document counts = %+v, want 2 created and 0 unchanged", initialDiagnostics)
	}
	if initialDiagnostics.ChunksWritten == 0 || initialDiagnostics.FTSRowsWritten == 0 {
		t.Fatalf("initial diagnostics did not record chunk/FTS writes: %+v", initialDiagnostics)
	}
	if initialDiagnostics.FTSStrategy != ftsStrategyBulkRebuild || initialDiagnostics.FTSRebuildSkipped {
		t.Fatalf("initial FTS diagnostics = %+v, want bulk rebuild", initialDiagnostics)
	}
	if initialDiagnostics.IncrementalFTSWriteSeconds != 0 {
		t.Fatalf("initial incremental FTS seconds = %.2f, want deferred bulk rebuild", initialDiagnostics.IncrementalFTSWriteSeconds)
	}
	if len(initialDiagnostics.ProjectionRebuilds) == 0 || initialDiagnostics.ProjectionRebuildSkipped {
		t.Fatalf("initial diagnostics did not record projection rebuilds: %+v", initialDiagnostics)
	}

	reopenDiagnosticsPath := filepath.Join(t.TempDir(), "reopen-sync.json")
	reopened, err := New(context.Background(), Config{
		Backend:             domain.BackendOpenClerk,
		DatabasePath:        dbPath,
		VaultRoot:           vaultRoot,
		SyncDiagnosticsPath: reopenDiagnosticsPath,
	})
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer func() {
		_ = reopened.Close()
	}()

	reopenDiagnostics := readSyncDiagnosticsForTest(t, reopenDiagnosticsPath)
	if reopenDiagnostics.DocumentsCreated != 0 || reopenDiagnostics.DocumentsUpdated != 0 || reopenDiagnostics.DocumentsPruned != 0 {
		t.Fatalf("reopen changed document counts = %+v, want no changes", reopenDiagnostics)
	}
	if reopenDiagnostics.DocumentsUnchanged != 2 {
		t.Fatalf("reopen unchanged documents = %d, want 2", reopenDiagnostics.DocumentsUnchanged)
	}
	if reopenDiagnostics.ChunksWritten != 0 || reopenDiagnostics.FTSRowsWritten != 0 {
		t.Fatalf("reopen wrote chunks/FTS rows: %+v", reopenDiagnostics)
	}
	if reopenDiagnostics.FTSStrategy != ftsStrategySkippedNoChanges || !reopenDiagnostics.FTSRebuildSkipped {
		t.Fatalf("reopen FTS diagnostics = %+v, want rebuild skipped", reopenDiagnostics)
	}
	if !reopenDiagnostics.ProjectionRebuildSkipped || len(reopenDiagnostics.ProjectionRebuilds) != 0 {
		t.Fatalf("reopen projection diagnostics = %+v, want rebuild skipped", reopenDiagnostics)
	}
}

func TestSyncVaultRebuildsProjectionAfterInterruptedImport(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	writeVaultFile(t, vaultRoot, "records/services/openclerk-runner.md", strings.TrimSpace(`---
service_id: openclerk-runner
service_name: OpenClerk runner
service_status: active
service_owner: runner
service_interface: JSON runner
---
# OpenClerk runner

## Summary
Production service.
`)+"\n")

	store, err := NewUnsynced(ctx, Config{
		Backend:      domain.BackendOpenClerk,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("open unsynced store: %v", err)
	}
	if _, err := store.syncDocumentFromDiskWithOptions(ctx, "records/services/openclerk-runner.md", "", documentSyncOptions{
		RebuildProjections: false,
	}); err != nil {
		t.Fatalf("sync document without projection rebuild: %v", err)
	}
	pending, err := store.projectionRebuildPending(ctx)
	if err != nil {
		t.Fatalf("check pending projection rebuild: %v", err)
	}
	if !pending {
		t.Fatalf("projection rebuild pending = false, want true after interrupted import")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close interrupted store: %v", err)
	}

	reopened, err := New(ctx, Config{
		Backend:      domain.BackendOpenClerk,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer func() {
		_ = reopened.Close()
	}()

	services, err := reopened.ServicesLookup(ctx, domain.ServiceLookupInput{Text: "OpenClerk runner", Limit: 10})
	if err != nil {
		t.Fatalf("services lookup after rebuild: %v", err)
	}
	if len(services.Services) != 1 {
		t.Fatalf("services after interrupted import recovery = %+v, want one service", services.Services)
	}
	pending, err = reopened.projectionRebuildPending(ctx)
	if err != nil {
		t.Fatalf("check pending projection rebuild after recovery: %v", err)
	}
	if pending {
		t.Fatalf("projection rebuild pending = true after successful recovery")
	}
}

func TestSyncVaultRebuildsFTSAfterInterruptedDeferredImport(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	writeVaultFile(t, vaultRoot, "sources/alpha.md", "# Alpha\n\nalpha interrupted marker\n")

	store, err := NewUnsynced(ctx, Config{
		Backend:      domain.BackendOpenClerk,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("open unsynced store: %v", err)
	}
	diagnostics := newSyncDiagnostics()
	if _, err := store.syncDocumentFromDiskWithOptions(ctx, "sources/alpha.md", "", documentSyncOptions{
		RebuildProjections: false,
		DeferFTS:           true,
		Diagnostics:        &diagnostics,
	}); err != nil {
		t.Fatalf("sync document with deferred FTS: %v", err)
	}
	if diagnostics.FTSRowsWritten != 0 || diagnostics.IncrementalFTSWriteSeconds != 0 {
		t.Fatalf("deferred FTS diagnostics = %+v, want no incremental writes", diagnostics)
	}
	pending, err := store.ftsRebuildPending(ctx)
	if err != nil {
		t.Fatalf("check pending FTS rebuild: %v", err)
	}
	if !pending {
		t.Fatalf("FTS rebuild pending = false, want true after deferred import")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close interrupted store: %v", err)
	}

	diagnosticsPath := filepath.Join(t.TempDir(), "recovered-sync.json")
	reopened, err := New(ctx, Config{
		Backend:             domain.BackendOpenClerk,
		DatabasePath:        dbPath,
		VaultRoot:           vaultRoot,
		SyncDiagnosticsPath: diagnosticsPath,
	})
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer func() {
		_ = reopened.Close()
	}()

	search, err := reopened.Search(ctx, domain.SearchQuery{Text: "interrupted marker", Limit: 10})
	if err != nil {
		t.Fatalf("search after FTS rebuild: %v", err)
	}
	if len(search.Hits) != 1 {
		t.Fatalf("search after FTS rebuild hits = %d, want 1", len(search.Hits))
	}
	pending, err = reopened.ftsRebuildPending(ctx)
	if err != nil {
		t.Fatalf("check pending FTS rebuild after recovery: %v", err)
	}
	if pending {
		t.Fatalf("FTS rebuild pending = true after successful recovery")
	}
	recoveryDiagnostics := readSyncDiagnosticsForTest(t, diagnosticsPath)
	if recoveryDiagnostics.FTSStrategy != ftsStrategyBulkRebuild || recoveryDiagnostics.FTSRebuildPending || !recoveryDiagnostics.FTSBootstrap {
		t.Fatalf("recovery FTS diagnostics = %+v, want completed bulk rebuild", recoveryDiagnostics)
	}
	if recoveryDiagnostics.FTSRowsWritten == 0 {
		t.Fatalf("recovery FTS rows written = 0, want rebuilt rows")
	}
}

func TestSyncVaultRebuildsFTSWhenCountsMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	writeVaultFile(t, vaultRoot, "sources/alpha.md", "# Alpha\n\nalpha mismatch marker\n")

	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	corrupt, err := NewUnsynced(ctx, Config{
		Backend:      domain.BackendOpenClerk,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("open unsynced store for mismatch setup: %v", err)
	}
	if _, err := corrupt.db.ExecContext(ctx, `DELETE FROM chunk_fts`); err != nil {
		t.Fatalf("delete FTS rows for mismatch setup: %v", err)
	}
	if err := corrupt.Close(); err != nil {
		t.Fatalf("close mismatch setup store: %v", err)
	}

	diagnosticsPath := filepath.Join(t.TempDir(), "mismatch-sync.json")
	reopened, err := New(ctx, Config{
		Backend:             domain.BackendOpenClerk,
		DatabasePath:        dbPath,
		VaultRoot:           vaultRoot,
		SyncDiagnosticsPath: diagnosticsPath,
	})
	if err != nil {
		t.Fatalf("reopen store after mismatch: %v", err)
	}
	defer func() {
		_ = reopened.Close()
	}()

	search, err := reopened.Search(ctx, domain.SearchQuery{Text: "mismatch marker", Limit: 10})
	if err != nil {
		t.Fatalf("search after mismatch rebuild: %v", err)
	}
	if len(search.Hits) != 1 {
		t.Fatalf("search after mismatch rebuild hits = %d, want 1", len(search.Hits))
	}
	diagnostics := readSyncDiagnosticsForTest(t, diagnosticsPath)
	if diagnostics.FTSStrategy != ftsStrategyBulkRebuild || !diagnostics.FTSBootstrap || diagnostics.FTSRebuildPending {
		t.Fatalf("mismatch FTS diagnostics = %+v, want bootstrap bulk rebuild without pending flag", diagnostics)
	}
}

func TestSyncDocumentFromDiskUsesIncrementalFTS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	writeVaultFile(t, vaultRoot, "sources/alpha.md", "# Alpha\n\nalpha incremental marker\n")

	store, err := NewUnsynced(ctx, Config{
		Backend:      domain.BackendOpenClerk,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("open unsynced store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	diagnostics := newSyncDiagnostics()
	if _, err := store.syncDocumentFromDiskWithOptions(ctx, "sources/alpha.md", "", documentSyncOptions{
		RebuildProjections: false,
		Diagnostics:        &diagnostics,
	}); err != nil {
		t.Fatalf("sync document with incremental FTS: %v", err)
	}
	if diagnostics.FTSStrategy != ftsStrategyIncrementalRows || diagnostics.FTSRowsWritten == 0 {
		t.Fatalf("incremental FTS diagnostics = %+v, want row writes", diagnostics)
	}
	pending, err := store.ftsRebuildPending(ctx)
	if err != nil {
		t.Fatalf("check pending FTS rebuild: %v", err)
	}
	if pending {
		t.Fatalf("FTS rebuild pending = true after incremental sync")
	}
	search, err := store.Search(ctx, domain.SearchQuery{Text: "incremental marker", Limit: 10})
	if err != nil {
		t.Fatalf("search after incremental sync: %v", err)
	}
	if len(search.Hits) != 1 {
		t.Fatalf("search after incremental sync hits = %d, want 1", len(search.Hits))
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

func writeVaultFile(t *testing.T, vaultRoot string, relPath string, content string) {
	t.Helper()
	absPath := filepath.Join(vaultRoot, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", relPath, err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

func readSyncDiagnosticsForTest(t *testing.T, path string) SyncDiagnostics {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read sync diagnostics: %v", err)
	}
	var diagnostics SyncDiagnostics
	if err := json.Unmarshal(content, &diagnostics); err != nil {
		t.Fatalf("decode sync diagnostics: %v", err)
	}
	return diagnostics
}
