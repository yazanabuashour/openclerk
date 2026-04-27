package sqlite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

func TestCreateDocumentRejectsDuplicatePath(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	first, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Widget One",
		Body:  "# Widget One\n\nfirst body",
	})
	if err != nil {
		t.Fatalf("create first document: %v", err)
	}

	_, err = store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Widget Two",
		Body:  "# Widget Two\n\nsecond body",
	})
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 409 {
		t.Fatalf("duplicate create error = %v, want already exists 409", err)
	}

	got, err := store.GetDocument(context.Background(), first.DocID)
	if err != nil {
		t.Fatalf("get original document: %v", err)
	}
	if got.Title != "Widget One" || !strings.Contains(got.Body, "first body") {
		t.Fatalf("original document was overwritten: %+v", got)
	}
}

func TestIngestSourceURLKeepsAssetWhenSourceNotePersistsBeforeError(t *testing.T) {
	pdfBytes := minimalStorePDF("Partial ingest PDF", "OpenClerk Test", "Partial ingest text")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(pdfBytes)
	}))
	t.Cleanup(server.Close)

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	oldWriteFile := osWriteFile
	osWriteFile = func(name string, data string) error {
		if strings.HasSuffix(name, filepath.Join("sources", "partial-ingest.md")) {
			if err := os.WriteFile(name, []byte(data), 0o644); err != nil {
				return err
			}
			return errors.New("forced source note sync failure")
		}
		return oldWriteFile(name, data)
	}
	t.Cleanup(func() {
		osWriteFile = oldWriteFile
	})

	_, err := store.IngestSourceURL(context.Background(), domain.SourceURLInput{
		URL:           server.URL + "/partial.pdf",
		PathHint:      "sources/partial-ingest.md",
		AssetPathHint: "assets/sources/partial-ingest.pdf",
	})
	if err == nil {
		t.Fatalf("ingest error = nil, want forced source note failure")
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "sources", "partial-ingest.md")); err != nil {
		t.Fatalf("source note stat: %v", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "assets", "sources", "partial-ingest.pdf")); err != nil {
		t.Fatalf("asset stat: %v", err)
	}
}

func TestIngestSourceURLEvalFixtureURL(t *testing.T) {
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "artifacts", "vendor-security-paper.pdf")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, minimalStorePDF("Eval fixture PDF", "OpenClerk Test", "Eval fixture evidence"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	t.Setenv(evalSourceFixtureRootEnv, fixtureRoot)

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	const sourceURL = "http://openclerk-eval.local/artifacts/vendor-security-paper.pdf"
	ingestion, err := store.IngestSourceURL(context.Background(), domain.SourceURLInput{
		URL:           sourceURL,
		PathHint:      "sources/eval-fixture.md",
		AssetPathHint: "assets/sources/eval-fixture.pdf",
		Title:         "Eval Fixture",
	})
	if err != nil {
		t.Fatalf("ingest eval fixture: %v", err)
	}
	if ingestion.SourcePath != "sources/eval-fixture.md" || ingestion.AssetPath != "assets/sources/eval-fixture.pdf" || len(ingestion.Citations) == 0 {
		t.Fatalf("ingestion = %+v", ingestion)
	}
	doc, err := store.GetDocument(context.Background(), ingestion.DocID)
	if err != nil {
		t.Fatalf("get ingested eval fixture source: %v", err)
	}
	if doc.Metadata["source_url"] != sourceURL || doc.Metadata["source_type"] != "pdf" {
		t.Fatalf("metadata = %+v", doc.Metadata)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "assets", "sources", "eval-fixture.pdf")); err != nil {
		t.Fatalf("asset stat: %v", err)
	}
}

func TestEvalFixtureURLNotInterceptedWithoutEnv(t *testing.T) {
	t.Setenv(evalSourceFixtureRootEnv, "")

	_, ok, err := resolveEvalSourceFixturePath("http://openclerk-eval.local/artifacts/vendor-security-paper.pdf")
	if err != nil {
		t.Fatalf("resolve eval fixture without env: %v", err)
	}
	if ok {
		t.Fatal("eval fixture URL was intercepted without fixture env")
	}
}

func TestIngestSourceURLUpdateMode(t *testing.T) {
	var (
		mu         sync.Mutex
		currentPDF = minimalStorePDF("Runner Intake PDF Title", "OpenClerk Test", "Initial PDF evidence")
		updatedPDF = minimalStorePDF("Runner Intake PDF Title Updated", "OpenClerk Test", "Updated PDF evidence")
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		body := append([]byte(nil), currentPDF...)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	created, err := store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:           server.URL + "/runner.pdf",
		PathHint:      "sources/runner-ingest.md",
		AssetPathHint: "assets/sources/runner-ingest.pdf",
		Title:         "Runner Ingest Override",
	})
	if err != nil {
		t.Fatalf("create source URL: %v", err)
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "synthesis/runner.md",
		Title: "Runner Synthesis",
		Body:  synthesisBody("sources/runner-ingest.md", "Initial PDF evidence."),
	}); err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	_, err = store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:  server.URL + "/missing.pdf",
		Mode: "update",
	})
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 404 {
		t.Fatalf("missing update error = %v, want not found 404", err)
	}

	_, err = store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:      server.URL + "/runner.pdf",
		PathHint: "sources/other.md",
		Mode:     "update",
	})
	if !errors.As(err, &appErr) || appErr.Status != 409 || appErr.Code != "conflict" {
		t.Fatalf("mismatched path update error = %v, want conflict 409", err)
	}
	_, err = store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:           server.URL + "/runner.pdf",
		AssetPathHint: "assets/sources/other.pdf",
		Mode:          "update",
	})
	if !errors.As(err, &appErr) || appErr.Status != 409 || appErr.Code != "conflict" {
		t.Fatalf("mismatched asset update error = %v, want conflict 409", err)
	}

	beforeSourceEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "source", RefID: created.DocID, Limit: 20})
	if err != nil {
		t.Fatalf("source events before no-op: %v", err)
	}
	beforeProjectionEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "projection", Limit: 50})
	if err != nil {
		t.Fatalf("projection events before no-op: %v", err)
	}
	same, err := store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:  server.URL + "/runner.pdf",
		Mode: "update",
	})
	if err != nil {
		t.Fatalf("same PDF update: %v", err)
	}
	if same.DocID != created.DocID || same.SourcePath != created.SourcePath || same.AssetPath != created.AssetPath || same.SHA256 != created.SHA256 || len(same.Citations) == 0 {
		t.Fatalf("same PDF update result = %+v, want existing ingestion result", same)
	}
	afterSourceEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "source", RefID: created.DocID, Limit: 20})
	if err != nil {
		t.Fatalf("source events after no-op: %v", err)
	}
	afterProjectionEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "projection", Limit: 50})
	if err != nil {
		t.Fatalf("projection events after no-op: %v", err)
	}
	if countEventType(afterSourceEvents.Events, "source_updated") != countEventType(beforeSourceEvents.Events, "source_updated") ||
		countEventType(afterProjectionEvents.Events, "projection_invalidated") != countEventType(beforeProjectionEvents.Events, "projection_invalidated") {
		t.Fatalf("same PDF update created stale-state churn: source=%+v projection=%+v", afterSourceEvents.Events, afterProjectionEvents.Events)
	}

	mu.Lock()
	currentPDF = updatedPDF
	mu.Unlock()
	changed, err := store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:           server.URL + "/runner.pdf",
		PathHint:      "sources/runner-ingest.md",
		AssetPathHint: "assets/sources/runner-ingest.pdf",
		Mode:          "update",
	})
	if err != nil {
		t.Fatalf("changed PDF update: %v", err)
	}
	if changed.DocID != created.DocID || changed.SourcePath != created.SourcePath || changed.AssetPath != created.AssetPath || changed.SHA256 == created.SHA256 {
		t.Fatalf("changed update = %+v, created = %+v", changed, created)
	}
	updatedDoc, err := store.GetDocument(ctx, created.DocID)
	if err != nil {
		t.Fatalf("get updated source: %v", err)
	}
	if updatedDoc.Metadata["sha256"] != changed.SHA256 || !strings.Contains(updatedDoc.Body, "UpdatedPDFevidence") {
		t.Fatalf("updated source document = %+v", updatedDoc)
	}
	search, err := store.Search(ctx, domain.SearchQuery{Text: "UpdatedPDFevidence", PathPrefix: "sources/", Limit: 10})
	if err != nil {
		t.Fatalf("search updated source: %v", err)
	}
	if len(search.Hits) == 0 || search.Hits[0].Citations[0].DocID != created.DocID {
		t.Fatalf("search updated source = %+v", search)
	}
	sourceEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "source", RefID: created.DocID, Limit: 20})
	if err != nil {
		t.Fatalf("source update events: %v", err)
	}
	var foundUpdate bool
	for _, event := range sourceEvents.Events {
		if event.EventType == "source_updated" &&
			event.Details["previous_sha256"] == created.SHA256 &&
			event.Details["new_sha256"] == changed.SHA256 &&
			event.Details["asset_path"] == created.AssetPath &&
			event.Details["source_url"] == server.URL+"/runner.pdf" {
			foundUpdate = true
		}
	}
	if !foundUpdate {
		t.Fatalf("source update provenance = %+v", sourceEvents.Events)
	}
	projection := requireSynthesisProjection(t, ctx, store, docIDForPath("synthesis/runner.md"))
	if projection.Freshness != "stale" || projection.Details["stale_source_refs"] != "sources/runner-ingest.md" {
		t.Fatalf("synthesis projection after source update = %+v", projection)
	}
}

func TestIngestSourceURLUpdateRollbackRestoresPreviousState(t *testing.T) {
	var (
		mu         sync.Mutex
		currentPDF = minimalStorePDF("Rollback PDF", "OpenClerk Test", "Rollback old evidence")
		newPDF     = minimalStorePDF("Rollback PDF New", "OpenClerk Test", "Rollback new evidence")
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		body := append([]byte(nil), currentPDF...)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	created, err := store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:           server.URL + "/rollback.pdf",
		PathHint:      "sources/rollback.md",
		AssetPathHint: "assets/sources/rollback.pdf",
	})
	if err != nil {
		t.Fatalf("create source URL: %v", err)
	}
	oldDoc, err := store.GetDocument(ctx, created.DocID)
	if err != nil {
		t.Fatalf("get old document: %v", err)
	}
	oldAsset, err := os.ReadFile(filepath.Join(vaultRoot, "assets", "sources", "rollback.pdf"))
	if err != nil {
		t.Fatalf("read old asset: %v", err)
	}

	oldWriteFile := osWriteFile
	osWriteFile = func(name string, data string) error {
		if strings.HasSuffix(name, filepath.Join("sources", "rollback.md")) {
			if err := os.WriteFile(name, []byte(data), 0o644); err != nil {
				return err
			}
			return errors.New("forced source update note failure")
		}
		return oldWriteFile(name, data)
	}
	t.Cleanup(func() {
		osWriteFile = oldWriteFile
	})

	mu.Lock()
	currentPDF = newPDF
	mu.Unlock()
	_, err = store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:  server.URL + "/rollback.pdf",
		Mode: "update",
	})
	if err == nil {
		t.Fatalf("update error = nil, want forced failure")
	}
	gotDoc, err := store.GetDocument(ctx, created.DocID)
	if err != nil {
		t.Fatalf("get document after rollback: %v", err)
	}
	if gotDoc.Metadata["sha256"] != oldDoc.Metadata["sha256"] || gotDoc.Body != oldDoc.Body {
		t.Fatalf("document after rollback = %+v, want old metadata/body", gotDoc)
	}
	gotAsset, err := os.ReadFile(filepath.Join(vaultRoot, "assets", "sources", "rollback.pdf"))
	if err != nil {
		t.Fatalf("read asset after rollback: %v", err)
	}
	if !bytes.Equal(gotAsset, oldAsset) {
		t.Fatalf("asset after rollback changed")
	}
	search, err := store.Search(ctx, domain.SearchQuery{Text: "Rollbackoldevidence", PathPrefix: "sources/", Limit: 10})
	if err != nil {
		t.Fatalf("search old evidence after rollback: %v", err)
	}
	if len(search.Hits) == 0 {
		t.Fatalf("old evidence missing after rollback")
	}
	search, err = store.Search(ctx, domain.SearchQuery{Text: "Rollbacknewevidence", PathPrefix: "sources/", Limit: 10})
	if err != nil {
		t.Fatalf("search new evidence after rollback: %v", err)
	}
	if len(search.Hits) != 0 {
		t.Fatalf("new evidence indexed after rollback: %+v", search.Hits)
	}
}

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

func TestServicesLookupSearchesSummarySection(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

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
Production service for routine local knowledge tasks.

## Facts
- tier: production
`) + "\n",
	}); err != nil {
		t.Fatalf("create service document: %v", err)
	}

	services, err := store.ServicesLookup(context.Background(), domain.ServiceLookupInput{Text: "routine local knowledge", Limit: 10})
	if err != nil {
		t.Fatalf("services lookup: %v", err)
	}
	if len(services.Services) != 1 || services.Services[0].ServiceID != "openclerk-runner" {
		t.Fatalf("services lookup = %+v, want openclerk-runner", services)
	}
	if services.Services[0].Summary != "Production service for routine local knowledge tasks." {
		t.Fatalf("service summary = %q", services.Services[0].Summary)
	}
}

func TestDecisionProjectionLookupAndSupersessionFreshness(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	oldDecision, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "docs/architecture/old-runner-decision.md",
		Title: "Old runner decision",
		Body: strings.TrimSpace(`---
decision_id: adr-runner-old
decision_title: Old runner path
decision_status: superseded
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-20
superseded_by: adr-runner-current
source_refs: sources/runner-old.md
---
# Old runner path

## Summary
Old decision used a retired runner path.
`) + "\n",
	})
	if err != nil {
		t.Fatalf("create old decision: %v", err)
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/architecture/current-runner-decision.md",
		Title: "Current runner decision",
		Body: strings.TrimSpace(`---
decision_id: adr-runner-current
decision_title: Use JSON runner
decision_status: accepted
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-22
supersedes: adr-runner-old
source_refs: sources/runner-current.md
---
# Use JSON runner

## Summary
Accepted decision uses the JSON runner for routine AgentOps work.
`) + "\n",
	}); err != nil {
		t.Fatalf("create current decision: %v", err)
	}

	lookup, err := store.DecisionsLookup(ctx, domain.DecisionLookupInput{
		Text:   "JSON runner",
		Status: "accepted",
		Scope:  "agentops",
		Owner:  "platform",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("decision lookup: %v", err)
	}
	if len(lookup.Decisions) != 1 ||
		lookup.Decisions[0].DecisionID != "adr-runner-current" ||
		len(lookup.Decisions[0].Citations) == 0 ||
		lookup.Decisions[0].Citations[0].Path != "notes/architecture/current-runner-decision.md" {
		t.Fatalf("lookup = %+v", lookup)
	}

	detail, err := store.GetDecisionRecord(ctx, "adr-runner-old")
	if err != nil {
		t.Fatalf("decision detail: %v", err)
	}
	if detail.Status != "superseded" ||
		len(detail.SupersededBy) != 1 ||
		detail.SupersededBy[0] != "adr-runner-current" ||
		len(detail.Citations) == 0 {
		t.Fatalf("detail = %+v", detail)
	}

	projections, err := store.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: "decisions",
		RefKind:    "decision",
		RefID:      "adr-runner-old",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("decision projection: %v", err)
	}
	if len(projections.Projections) != 1 ||
		projections.Projections[0].Freshness != "stale" ||
		projections.Projections[0].Details["superseded_by"] != "adr-runner-current" {
		t.Fatalf("old projection = %+v", projections)
	}

	events, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "decision",
		RefID:   "adr-runner-current",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("decision provenance: %v", err)
	}
	if !hasEventType(events.Events, "decision_extracted_from_doc") {
		t.Fatalf("events = %+v", events)
	}

	if _, err := store.ReplaceDocumentSection(ctx, oldDecision.DocID, domain.ReplaceSectionInput{
		Heading: "Summary",
		Content: "Old decision is explicitly superseded by adr-runner-current.",
	}); err != nil {
		t.Fatalf("replace old decision summary: %v", err)
	}
	events, err = store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "projection",
		RefID:   "decisions-source:" + oldDecision.DocID,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("decision invalidation events: %v", err)
	}
	if !hasEventType(events.Events, "projection_invalidated") {
		t.Fatalf("invalidation events = %+v", events)
	}
}

func TestDecisionProjectionVersionChangesWhenReplacementAppears(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "docs/architecture/old-runner-decision.md",
		Title: "Old runner decision",
		Body: strings.TrimSpace(`---
decision_id: adr-runner-old
decision_title: Old runner path
decision_status: superseded
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-20
superseded_by: adr-runner-current
---
# Old runner path

## Summary
Old decision used a retired runner path.
`) + "\n",
	}); err != nil {
		t.Fatalf("create old decision: %v", err)
	}
	initial := requireDecisionProjection(t, ctx, store, "adr-runner-old")
	if initial.Details["missing_replacement_ids"] != "adr-runner-current" ||
		initial.Details["freshness_reason"] != "decision superseded with missing replacement" {
		t.Fatalf("initial old projection details = %+v", initial.Details)
	}

	clock = clock.Add(time.Minute)
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "records/decisions/current-runner-decision.md",
		Title: "Current runner decision",
		Body: strings.TrimSpace(`---
decision_id: adr-runner-current
decision_title: Use JSON runner
decision_status: accepted
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-22
supersedes: adr-runner-old
---
# Use JSON runner

## Summary
Accepted decision uses the JSON runner.
`) + "\n",
	}); err != nil {
		t.Fatalf("create current decision: %v", err)
	}
	updated := requireDecisionProjection(t, ctx, store, "adr-runner-old")
	if _, ok := updated.Details["missing_replacement_ids"]; ok {
		t.Fatalf("updated old projection still has missing replacement: %+v", updated.Details)
	}
	if updated.Details["freshness_reason"] != "decision superseded" {
		t.Fatalf("updated old projection details = %+v", updated.Details)
	}
	if updated.ProjectionVersion == initial.ProjectionVersion {
		t.Fatalf("projection version did not change after replacement appeared: %q", updated.ProjectionVersion)
	}
	if !updated.UpdatedAt.After(initial.UpdatedAt) {
		t.Fatalf("updated_at = %s, want after %s", updated.UpdatedAt, initial.UpdatedAt)
	}
}

func TestDecisionProjectionCoversADRMarkdownAndClassificationSearch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "docs/architecture/eval-backed-knowledge-plane-adr.md",
		Title: "AgentOps-Only Knowledge Plane Direction",
		Body: strings.TrimSpace(`---
decision_id: adr-agentops-only-knowledge-plane
decision_title: AgentOps-Only Knowledge Plane Direction
decision_status: accepted
decision_scope: knowledge-plane
decision_owner: platform
source_refs: sources/agentops-direction.md
---
# ADR: AgentOps-Only Knowledge Plane Direction

## Status
Accepted as the current architecture direction.

## Summary
OpenClerk uses AgentOps as the only production agent interface.
`) + "\n",
	}); err != nil {
		t.Fatalf("create adr decision: %v", err)
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "docs/architecture/knowledge-configuration-v1-adr.md",
		Title: "Knowledge Configuration v1",
		Body: strings.TrimSpace(`---
decision_id: adr-knowledge-configuration-v1
decision_title: Knowledge Configuration v1
decision_status: accepted
decision_scope: knowledge-configuration
decision_owner: platform
supersedes: adr-agentops-only-knowledge-plane
---
# ADR: Knowledge Configuration v1

## Status
Accepted as the v1 production contract.

## Summary
OpenClerk knowledge configuration v1 is runner-visible and convention-first.
`) + "\n",
	}); err != nil {
		t.Fatalf("create second adr decision: %v", err)
	}

	lookup, err := store.DecisionsLookup(ctx, domain.DecisionLookupInput{Text: "knowledge-plane", Limit: 10})
	if err != nil {
		t.Fatalf("decision lookup by classification text: %v", err)
	}
	if len(lookup.Decisions) != 2 {
		t.Fatalf("classification lookup = %+v, want both knowledge-plane decisions", lookup.Decisions)
	}
	sourceRefLookup, err := store.DecisionsLookup(ctx, domain.DecisionLookupInput{Text: "sources/agentops-direction.md", Limit: 10})
	if err != nil {
		t.Fatalf("decision lookup by source ref: %v", err)
	}
	if len(sourceRefLookup.Decisions) != 1 ||
		sourceRefLookup.Decisions[0].DecisionID != "adr-agentops-only-knowledge-plane" ||
		len(sourceRefLookup.Decisions[0].Citations) == 0 ||
		sourceRefLookup.Decisions[0].Citations[0].Path != "docs/architecture/eval-backed-knowledge-plane-adr.md" ||
		len(sourceRefLookup.Decisions[0].SourceRefs) != 1 ||
		sourceRefLookup.Decisions[0].SourceRefs[0] != "sources/agentops-direction.md" {
		t.Fatalf("source ref lookup = %+v", sourceRefLookup.Decisions)
	}

	projection := requireDecisionProjection(t, ctx, store, "adr-agentops-only-knowledge-plane")
	if projection.Freshness != "fresh" ||
		projection.Details["path"] != "docs/architecture/eval-backed-knowledge-plane-adr.md" ||
		projection.Details["freshness_reason"] != "decision current" {
		t.Fatalf("adr projection = %+v", projection)
	}
	events, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "decision",
		RefID:   "adr-agentops-only-knowledge-plane",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("decision provenance: %v", err)
	}
	if !hasEventType(events.Events, "decision_extracted_from_doc") {
		t.Fatalf("events = %+v", events)
	}
}

func TestDecisionProjectionRefreshesFromCanonicalADRMarkdownOnReopen(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)

	path := "docs/architecture/eval-backed-knowledge-plane-adr.md"
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  path,
		Title: "AgentOps-Only Knowledge Plane Direction",
		Body: strings.TrimSpace(`---
decision_id: adr-agentops-only-knowledge-plane
decision_title: AgentOps-Only Knowledge Plane Direction
decision_status: accepted
decision_scope: knowledge-plane
decision_owner: platform
---
# ADR: AgentOps-Only Knowledge Plane Direction

## Summary
Initial canonical decision text.
`) + "\n",
	}); err != nil {
		t.Fatalf("create adr decision: %v", err)
	}
	initial, err := store.GetDecisionRecord(ctx, "adr-agentops-only-knowledge-plane")
	if err != nil {
		t.Fatalf("initial decision detail: %v", err)
	}
	if initial.Title != "AgentOps-Only Knowledge Plane Direction" ||
		initial.Summary != "Initial canonical decision text." {
		t.Fatalf("initial decision = %+v", initial)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	updatedBody := strings.TrimSpace(`---
decision_id: adr-agentops-only-knowledge-plane
decision_title: Updated AgentOps Knowledge Plane Direction
decision_status: accepted
decision_scope: knowledge-plane
decision_owner: platform
source_refs: sources/updated-agentops-direction.md
---
# ADR: AgentOps-Only Knowledge Plane Direction

## Summary
Updated canonical decision text from markdown.
`) + "\n"
	if err := os.WriteFile(filepath.Join(vaultRoot, filepath.FromSlash(path)), []byte(updatedBody), 0o644); err != nil {
		t.Fatalf("write updated adr markdown: %v", err)
	}

	reopened := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = reopened.Close()
	}()
	updated, err := reopened.GetDecisionRecord(ctx, "adr-agentops-only-knowledge-plane")
	if err != nil {
		t.Fatalf("updated decision detail: %v", err)
	}
	if updated.Title != "Updated AgentOps Knowledge Plane Direction" ||
		updated.Summary != "Updated canonical decision text from markdown." ||
		len(updated.SourceRefs) != 1 ||
		updated.SourceRefs[0] != "sources/updated-agentops-direction.md" ||
		len(updated.Citations) == 0 ||
		updated.Citations[0].Path != path {
		t.Fatalf("updated decision = %+v", updated)
	}
	projection := requireDecisionProjection(t, ctx, reopened, "adr-agentops-only-knowledge-plane")
	if projection.Details["path"] != path ||
		projection.Details["source_refs"] != "sources/updated-agentops-direction.md" {
		t.Fatalf("projection after reopen = %+v", projection)
	}
}

func TestSynthesisProjectionIsFreshForCurrentSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	source, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "sources/current.md",
		Title: "Current Source",
		Body:  "# Current Source\n\n## Summary\nCurrent canonical evidence.\n",
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "synthesis/current.md",
		Title: "Current Synthesis",
		Body:  synthesisBody("sources/current.md", "Current canonical evidence."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	projection := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if projection.Freshness != "fresh" {
		t.Fatalf("freshness = %q, want fresh: %+v", projection.Freshness, projection)
	}
	if projection.SourceRef != "doc:"+source.DocID {
		t.Fatalf("source_ref = %q, want doc ref for source", projection.SourceRef)
	}
	if projection.Details["current_source_refs"] != "sources/current.md" ||
		projection.Details["source_refs"] != "sources/current.md" ||
		projection.Details["freshness_reason"] != "sources current" {
		t.Fatalf("projection details = %+v", projection.Details)
	}
}

func TestSynthesisProjectionSupportsRootRelativeNotesVaultPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	source, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "sources/current.md",
		Title: "Current Source",
		Body:  "# Current Source\n\n## Summary\nCurrent canonical evidence.\n",
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "synthesis/current.md",
		Title: "Current Synthesis",
		Body:  synthesisBody("\"sources/current.md\"", "Current canonical evidence."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	projection := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if projection.Freshness != "fresh" {
		t.Fatalf("freshness = %q, want fresh: %+v", projection.Freshness, projection)
	}
	if projection.SourceRef != "doc:"+source.DocID {
		t.Fatalf("source_ref = %q, want doc ref for root-relative source", projection.SourceRef)
	}
	if projection.Details["current_source_refs"] != "sources/current.md" ||
		projection.Details["source_refs"] != "sources/current.md" ||
		projection.Details["freshness_reason"] != "sources current" {
		t.Fatalf("projection details = %+v", projection.Details)
	}
}

func TestSynthesisProjectionStaleAfterSourceUpdateAndFreshAfterRepair(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	source, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "sources/runner.md",
		Title: "Runner Source",
		Body:  "# Runner Source\n\n## Summary\nInitial source guidance.\n",
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "synthesis/runner.md",
		Title: "Runner Synthesis",
		Body:  synthesisBody("sources/runner.md", "Initial source guidance."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}
	if got := requireSynthesisProjection(t, ctx, store, synthesis.DocID); got.Freshness != "fresh" {
		t.Fatalf("initial projection freshness = %q, want fresh", got.Freshness)
	}

	clock = clock.Add(time.Minute)
	if _, err := store.ReplaceDocumentSection(ctx, source.DocID, domain.ReplaceSectionInput{
		Heading: "Summary",
		Content: "Updated source guidance.",
	}); err != nil {
		t.Fatalf("update source: %v", err)
	}
	stale := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if stale.Freshness != "stale" ||
		stale.Details["stale_source_refs"] != "sources/runner.md" ||
		!strings.Contains(stale.Details["freshness_reason"], "source newer than synthesis") {
		t.Fatalf("stale projection = %+v", stale)
	}
	invalidations, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "projection",
		RefID:   "synthesis:" + synthesis.DocID,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("list invalidations: %v", err)
	}
	if !hasEventType(invalidations.Events, "projection_invalidated") {
		t.Fatalf("missing synthesis invalidation event: %+v", invalidations.Events)
	}
	sourceEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "source",
		RefID:   source.DocID,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("list source events: %v", err)
	}
	if !hasEventType(sourceEvents.Events, "source_created") || !hasEventType(sourceEvents.Events, "source_updated") {
		t.Fatalf("source events = %+v, want created and updated", sourceEvents.Events)
	}

	clock = clock.Add(time.Minute)
	if _, err := store.ReplaceDocumentSection(ctx, synthesis.DocID, domain.ReplaceSectionInput{
		Heading: "Freshness",
		Content: "Checked source: sources/runner.md after the source update.",
	}); err != nil {
		t.Fatalf("repair synthesis: %v", err)
	}
	repaired := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if repaired.Freshness != "fresh" || repaired.Details["stale_source_refs"] != "" {
		t.Fatalf("repaired projection = %+v", repaired)
	}
	events, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "projection",
		RefID:   "synthesis:" + synthesis.DocID,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("list synthesis events: %v", err)
	}
	if !hasEventType(events.Events, "projection_refreshed") {
		t.Fatalf("missing synthesis refresh event: %+v", events.Events)
	}
}

func TestSynthesisProjectionReportsMissingAndSupersededSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "sources/old.md",
		Title: "Old Source",
		Body: strings.TrimSpace(`---
status: superseded
superseded_by: sources/current.md
---
# Old Source

## Summary
Old guidance.
`) + "\n",
	}); err != nil {
		t.Fatalf("create old source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "synthesis/missing.md",
		Title: "Missing Synthesis",
		Body:  synthesisBody("sources/old.md, sources/missing.md", "Old guidance."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	projection := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if projection.Freshness != "stale" {
		t.Fatalf("freshness = %q, want stale", projection.Freshness)
	}
	if projection.Details["missing_source_refs"] != "sources/missing.md" {
		t.Fatalf("missing source refs = %q", projection.Details["missing_source_refs"])
	}
	if projection.Details["superseded_source_refs"] != "sources/old.md" {
		t.Fatalf("superseded source refs = %q", projection.Details["superseded_source_refs"])
	}
	if projection.Details["current_source_refs"] != "sources/current.md" {
		t.Fatalf("current source refs = %q", projection.Details["current_source_refs"])
	}
	if !strings.Contains(projection.Details["freshness_reason"], "current replacement missing from source refs") {
		t.Fatalf("freshness reason = %q", projection.Details["freshness_reason"])
	}
}

func TestSynthesisProjectionFreshWithSupersedesAndSupersededByMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time { return clock }

	clock = clock.Add(time.Minute)
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "sources/old.md",
		Title: "Old Source",
		Body: strings.TrimSpace(`---
status: superseded
superseded_by: sources/current.md
---
# Old Source

## Summary
Old guidance.
`) + "\n",
	}); err != nil {
		t.Fatalf("create old source: %v", err)
	}
	clock = clock.Add(time.Minute)
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "sources/current.md",
		Title: "Current Source",
		Body: strings.TrimSpace(`---
supersedes: sources/old.md
---
# Current Source

## Summary
Current guidance.
`) + "\n",
	}); err != nil {
		t.Fatalf("create current source: %v", err)
	}
	clock = clock.Add(time.Minute)
	synthesis, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "synthesis/supersession.md",
		Title: "Supersession Synthesis",
		Body:  synthesisBody("sources/current.md, sources/old.md", "Current guidance supersedes old guidance."),
	})
	if err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	projection := requireSynthesisProjection(t, ctx, store, synthesis.DocID)
	if projection.Freshness != "fresh" {
		t.Fatalf("freshness = %q, want fresh: %+v", projection.Freshness, projection)
	}
	if projection.Details["current_source_refs"] != "sources/current.md" {
		t.Fatalf("current source refs = %q", projection.Details["current_source_refs"])
	}
	if projection.Details["superseded_source_refs"] != "sources/old.md" {
		t.Fatalf("superseded source refs = %q", projection.Details["superseded_source_refs"])
	}
}

func TestCreateDocumentPreservesRequestedTitleAcrossRestart(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)

	document, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Wanted Title",
		Body:  "body only no heading",
	})
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	if document.Title != "Wanted Title" {
		t.Fatalf("created document title = %q, want %q", document.Title, "Wanted Title")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	reopened := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = reopened.Close()
	}()

	got, err := reopened.GetDocument(context.Background(), document.DocID)
	if err != nil {
		t.Fatalf("get document after restart: %v", err)
	}
	if got.Title != "Wanted Title" {
		t.Fatalf("reopened document title = %q, want %q", got.Title, "Wanted Title")
	}
}

func TestGraphNeighborhoodIncludesOutgoingLinksForChunk(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	target, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/reference.md",
		Title: "Reference",
		Body:  "# Reference\n\nCanonical supporting note.\n",
	})
	if err != nil {
		t.Fatalf("create target document: %v", err)
	}
	source, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/guide.md",
		Title: "Guide",
		Body: strings.TrimSpace(`
# Guide

## Overview
See the [reference](reference.md) for details.
`),
	})
	if err != nil {
		t.Fatalf("create source document: %v", err)
	}

	search, err := store.Search(context.Background(), domain.SearchQuery{Text: "reference", Limit: 10})
	if err != nil {
		t.Fatalf("search source chunk: %v", err)
	}
	var chunkID string
	for _, hit := range search.Hits {
		if hit.DocID == source.DocID {
			chunkID = hit.ChunkID
			break
		}
	}
	if chunkID == "" {
		t.Fatal("did not find source chunk in search results")
	}

	neighborhood, err := store.GraphNeighborhood(context.Background(), domain.GraphNeighborhoodInput{ChunkID: chunkID, Limit: 10})
	if err != nil {
		t.Fatalf("graph neighborhood by chunk: %v", err)
	}

	targetNodeID := "doc:" + target.DocID
	foundNode := false
	foundEdge := false
	for _, node := range neighborhood.Nodes {
		if node.NodeID == targetNodeID {
			foundNode = true
		}
	}
	for _, edge := range neighborhood.Edges {
		if edge.FromNodeID == "chunk:"+chunkID && edge.ToNodeID == targetNodeID && edge.Kind == "links_to" {
			foundEdge = true
		}
	}
	if !foundNode || !foundEdge {
		t.Fatalf("chunk neighborhood missing outgoing link: nodes=%v edges=%v", neighborhood.Nodes, neighborhood.Edges)
	}
}

func openTestStore(t *testing.T, backend domain.BackendKind, dbPath string, vaultRoot string) *Store {
	t.Helper()

	store, err := New(context.Background(), Config{
		Backend:      backend,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	return store
}

func testClock() time.Time {
	return time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
}

func synthesisBody(sourceRefs string, summary string) string {
	return strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: `+sourceRefs+`
---
# Synthesis

## Summary
`+summary+`

## Sources
- `+sourceRefs+`

## Freshness
Checked source refs.
`) + "\n"
}

func requireSynthesisProjection(t *testing.T, ctx context.Context, store *Store, docID string) domain.ProjectionState {
	t.Helper()

	result, err := store.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: "synthesis",
		RefKind:    "document",
		RefID:      docID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("list synthesis projection: %v", err)
	}
	if len(result.Projections) != 1 {
		t.Fatalf("projection count = %d, want 1: %+v", len(result.Projections), result.Projections)
	}
	return result.Projections[0]
}

func requireDecisionProjection(t *testing.T, ctx context.Context, store *Store, decisionID string) domain.ProjectionState {
	t.Helper()

	result, err := store.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: "decisions",
		RefKind:    "decision",
		RefID:      decisionID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("list decision projection: %v", err)
	}
	if len(result.Projections) != 1 {
		t.Fatalf("projection count = %d, want 1: %+v", len(result.Projections), result.Projections)
	}
	return result.Projections[0]
}

func hasEventType(events []domain.ProvenanceEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func countEventType(events []domain.ProvenanceEvent, eventType string) int {
	count := 0
	for _, event := range events {
		if event.EventType == eventType {
			count++
		}
	}
	return count
}

func minimalStorePDF(title string, author string, text string) []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, 6)
	writeObject := func(id int, body string) {
		offsets = append(offsets, buf.Len())
		_, _ = fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", id, body)
	}
	writeObject(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObject(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObject(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>")
	writeObject(4, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	stream := fmt.Sprintf("BT /F1 24 Tf 72 720 Td (%s) Tj ET", pdfEscape(text))
	writeObject(5, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream))
	writeObject(6, fmt.Sprintf("<< /Title (%s) /Author (%s) /CreationDate (D:20260426000000Z) >>", pdfEscape(title), pdfEscape(author)))
	xrefStart := buf.Len()
	buf.WriteString("xref\n0 7\n")
	buf.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		_, _ = fmt.Fprintf(&buf, "%010d 00000 n \n", offset)
	}
	_, _ = fmt.Fprintf(&buf, "trailer\n<< /Size 7 /Root 1 0 R /Info 6 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefStart)
	return buf.Bytes()
}

func pdfEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "(", `\(`)
	value = strings.ReplaceAll(value, ")", `\)`)
	return value
}
