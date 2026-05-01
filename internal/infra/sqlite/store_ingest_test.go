package sqlite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

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
	requireNoOpSourceUpdateImpact(t, same, created.SourceURL, created.DocID, created.SHA256)
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
	requireChangedSourceUpdateImpact(t, changed, created.SourceURL, created.DocID, created.SHA256, changed.SHA256, "synthesis/runner.md", "sources/runner-ingest.md")
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

func TestIngestSourceURLWebCreateAndUpdateMode(t *testing.T) {
	initialHTML := `<!doctype html><html><head><title>Runner Web Title</title><script>hidden()</script></head><body><nav><a href="/login">Login</a></nav><h1>Runner Web Title</h1><p>Initial web evidence for OpenClerk.</p><button>Add to cart</button></body></html>`
	updatedHTML := `<!doctype html><html><head><title>Runner Web Title Updated</title></head><body><h1>Runner Web Title Updated</h1><p>Updated web evidence for OpenClerk.</p></body></html>`
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "runner-product.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(initialHTML), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv(evalSourceFixtureRootEnv, fixtureRoot)
	sourceURL := "http://openclerk-eval.local/web/runner-product.html?tag=tracking"

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	created, err := store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:      sourceURL + "#fragment",
		PathHint: "sources/web/runner-product.md",
	})
	if err != nil {
		t.Fatalf("create web source URL: %v", err)
	}
	if created.SourceType != "web" ||
		created.SourceURL != sourceURL ||
		created.AssetPath != "" ||
		created.SourcePath != "sources/web/runner-product.md" ||
		created.DerivedPath != "sources/web/runner-product.md" ||
		created.MIMEType != "text/html" ||
		created.PageCount != 0 ||
		len(created.SHA256) != 64 ||
		len(created.Citations) == 0 {
		t.Fatalf("created web ingestion = %+v", created)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "assets")); !os.IsNotExist(err) {
		t.Fatalf("web ingestion created asset directory: %v", err)
	}
	doc, err := store.GetDocument(ctx, created.DocID)
	if err != nil {
		t.Fatalf("get web source: %v", err)
	}
	if doc.Metadata["source_type"] != "web" ||
		doc.Metadata["source_url"] != created.SourceURL ||
		doc.Metadata["asset_path"] != "" ||
		!strings.Contains(doc.Body, "Initial web evidence for OpenClerk.") ||
		strings.Contains(doc.Body, "hidden()") {
		t.Fatalf("web source document = %+v", doc)
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "synthesis/web-runner.md",
		Title: "Web Runner Synthesis",
		Body:  synthesisBody("sources/web/runner-product.md", "Initial web evidence for OpenClerk."),
	}); err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	var appErr *domain.Error
	_, err = store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:      created.SourceURL,
		PathHint: "sources/web/runner-product-copy.md",
	})
	if !errors.As(err, &appErr) || appErr.Status != 409 || appErr.Code != "already_exists" {
		t.Fatalf("duplicate web create error = %v, want already_exists 409", err)
	}
	duplicateDocs, err := store.ListDocuments(ctx, domain.DocumentListQuery{PathPrefix: "sources/web/runner-product-copy.md", Limit: 10})
	if err != nil {
		t.Fatalf("list duplicate copy path: %v", err)
	}
	if len(duplicateDocs.Documents) != 0 {
		t.Fatalf("duplicate create wrote copy path: %+v", duplicateDocs.Documents)
	}
	_, err = store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:        created.SourceURL,
		SourceType: "pdf",
		Mode:       "update",
	})
	if !errors.As(err, &appErr) || appErr.Status != 409 || appErr.Code != "conflict" {
		t.Fatalf("mismatched source type update error = %v, want conflict 409", err)
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
		URL:  created.SourceURL,
		Mode: "update",
	})
	if err != nil {
		t.Fatalf("same web update: %v", err)
	}
	if same.DocID != created.DocID || same.SourcePath != created.SourcePath || same.AssetPath != "" || same.SHA256 != created.SHA256 || same.SourceType != "web" {
		t.Fatalf("same web update result = %+v, want existing ingestion result", same)
	}
	requireNoOpSourceUpdateImpact(t, same, created.SourceURL, created.DocID, created.SHA256)
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
		t.Fatalf("same web update created stale-state churn: source=%+v projection=%+v", afterSourceEvents.Events, afterProjectionEvents.Events)
	}

	if err := os.WriteFile(fixturePath, []byte(updatedHTML), 0o644); err != nil {
		t.Fatalf("write updated web fixture: %v", err)
	}
	changed, err := store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:        created.SourceURL,
		PathHint:   "sources/web/runner-product.md",
		SourceType: "web",
		Mode:       "update",
	})
	if err != nil {
		t.Fatalf("changed web update: %v", err)
	}
	if changed.DocID != created.DocID || changed.SourcePath != created.SourcePath || changed.AssetPath != "" || changed.SHA256 == created.SHA256 {
		t.Fatalf("changed web update = %+v, created = %+v", changed, created)
	}
	requireChangedSourceUpdateImpact(t, changed, created.SourceURL, created.DocID, created.SHA256, changed.SHA256, "synthesis/web-runner.md", "sources/web/runner-product.md")
	updatedDoc, err := store.GetDocument(ctx, created.DocID)
	if err != nil {
		t.Fatalf("get updated web source: %v", err)
	}
	if updatedDoc.Metadata["sha256"] != changed.SHA256 || !strings.Contains(updatedDoc.Body, "Updated web evidence for OpenClerk.") {
		t.Fatalf("updated web source document = %+v", updatedDoc)
	}
	search, err := store.Search(ctx, domain.SearchQuery{Text: "Updated web evidence", PathPrefix: "sources/", Limit: 10})
	if err != nil {
		t.Fatalf("search updated web source: %v", err)
	}
	if len(search.Hits) == 0 || search.Hits[0].Citations[0].DocID != created.DocID {
		t.Fatalf("search updated web source = %+v", search)
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
			event.Details["source_type"] == "web" &&
			event.Details["source_url"] == created.SourceURL {
			foundUpdate = true
		}
	}
	if !foundUpdate {
		t.Fatalf("source update provenance = %+v", sourceEvents.Events)
	}
	projection := requireSynthesisProjection(t, ctx, store, docIDForPath("synthesis/web-runner.md"))
	if projection.Freshness != "stale" || projection.Details["stale_source_refs"] != "sources/web/runner-product.md" {
		t.Fatalf("synthesis projection after web source update = %+v", projection)
	}
}

func TestIngestSourceURLUpdateImpactPagesSynthesisProjections(t *testing.T) {
	initialHTML := `<!doctype html><html><head><title>Paged Source</title></head><body><h1>Paged Source</h1><p>Initial paged source evidence.</p></body></html>`
	updatedHTML := `<!doctype html><html><head><title>Paged Source Updated</title></head><body><h1>Paged Source Updated</h1><p>Updated paged source evidence.</p></body></html>`
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "paged-source.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(initialHTML), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv(evalSourceFixtureRootEnv, fixtureRoot)

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	clock := testClock()
	store.now = func() time.Time {
		return clock
	}

	sourceURL := "http://openclerk-eval.local/web/paged-source.html"
	created, err := store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:      sourceURL,
		PathHint: "sources/web/paged-source.md",
	})
	if err != nil {
		t.Fatalf("create web source URL: %v", err)
	}
	clock = clock.Add(time.Minute)
	const synthesisCount = 105
	for i := 0; i < synthesisCount; i++ {
		if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
			Path:  fmt.Sprintf("synthesis/paged-%03d.md", i),
			Title: fmt.Sprintf("Paged Synthesis %03d", i),
			Body:  synthesisBody("sources/web/paged-source.md", "Initial paged source evidence."),
		}); err != nil {
			t.Fatalf("create synthesis %03d: %v", i, err)
		}
	}

	if err := os.WriteFile(fixturePath, []byte(updatedHTML), 0o644); err != nil {
		t.Fatalf("write updated web fixture: %v", err)
	}
	clock = clock.Add(time.Minute)
	changed, err := store.IngestSourceURL(ctx, domain.SourceURLInput{
		URL:        sourceURL,
		SourceType: "web",
		Mode:       "update",
	})
	if err != nil {
		t.Fatalf("changed web update: %v", err)
	}
	if changed.UpdateStatus != "changed" || changed.DocID != created.DocID {
		t.Fatalf("changed update = %+v, created = %+v", changed, created)
	}
	if len(changed.StaleDependents) != synthesisCount || len(changed.ProjectionRefs) != synthesisCount {
		t.Fatalf("paged stale impact counts: dependents=%d projections=%d, want %d", len(changed.StaleDependents), len(changed.ProjectionRefs), synthesisCount)
	}
	if len(changed.ProvenanceRefs) <= synthesisCount {
		t.Fatalf("paged provenance refs = %d, want source update plus %d projection invalidations: %+v", len(changed.ProvenanceRefs), synthesisCount, changed.ProvenanceRefs)
	}
	if !sourceImpactHasStaleDependent(changed.StaleDependents, "synthesis/paged-104.md", "sources/web/paged-source.md") {
		t.Fatalf("stale dependents missing final paged synthesis: %+v", changed.StaleDependents)
	}
}

func TestIngestSourceURLWebRejectsPrivateNetworkHost(t *testing.T) {
	_, err := downloadSource(context.Background(), "http://127.0.0.1/private-page", sourceTypeWeb)
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 400 || !strings.Contains(appErr.Message, "publicly fetchable") {
		t.Fatalf("private web URL error = %v, want publicly fetchable validation", err)
	}
}

func TestIngestSourceURLWebRejectsPrivateRedirectTarget(t *testing.T) {
	oldClient := sourceHTTPClient
	defer func() {
		sourceHTTPClient = oldClient
	}()
	var requestedPrivate bool
	sourceHTTPClient = func(checkRedirect func(*http.Request, []*http.Request) error) *http.Client {
		return &http.Client{
			CheckRedirect: checkRedirect,
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.URL.Host == "127.0.0.1" {
					requestedPrivate = true
					return htmlTestResponse(request, "private"), nil
				}
				return &http.Response{
					StatusCode: http.StatusFound,
					Header:     http.Header{"Location": []string{"http://127.0.0.1/private-page"}},
					Body:       io.NopCloser(strings.NewReader("")),
					Request:    request,
				}, nil
			}),
		}
	}

	_, err := downloadSource(context.Background(), "http://93.184.216.34/web-page", sourceTypeWeb)
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 400 || !strings.Contains(appErr.Message, "publicly fetchable") {
		t.Fatalf("private redirect error = %v, want publicly fetchable validation", err)
	}
	if requestedPrivate {
		t.Fatal("followed private redirect target")
	}
}

func TestIngestSourceURLAutodetectedWebRejectsPrivateFinalURL(t *testing.T) {
	oldClient := sourceHTTPClient
	defer func() {
		sourceHTTPClient = oldClient
	}()
	sourceHTTPClient = func(checkRedirect func(*http.Request, []*http.Request) error) *http.Client {
		return &http.Client{
			CheckRedirect: checkRedirect,
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return htmlTestResponse(request, "private html from pdf-looking URL"), nil
			}),
		}
	}

	_, err := downloadSource(context.Background(), "http://127.0.0.1/private.pdf", "")
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 400 || !strings.Contains(appErr.Message, "publicly fetchable") {
		t.Fatalf("private autodetected web error = %v, want publicly fetchable validation", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func htmlTestResponse(request *http.Request, text string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       io.NopCloser(strings.NewReader("<!doctype html><html><body>" + text + "</body></html>")),
		Request:    request,
	}
}

func TestIngestVideoURLSuppliedTranscriptUpdateMode(t *testing.T) {
	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	videoURL := "https://www.youtube.com/watch?v=openclerk-demo"
	created, err := store.IngestVideoURL(ctx, domain.VideoURLInput{
		URL:           videoURL,
		PathHint:      "sources/video-youtube/runner-demo.md",
		AssetPathHint: "assets/video-youtube/runner-demo.json",
		Title:         "Runner Video Demo",
		Transcript: domain.VideoTranscriptInput{
			Text:       "Initial video transcript evidence.",
			Policy:     "supplied",
			Origin:     "manual_fixture",
			Language:   "en",
			CapturedAt: "2026-04-27T10:00:00Z",
			Tool:       "manual",
			Model:      "none",
		},
	})
	if err != nil {
		t.Fatalf("create video URL: %v", err)
	}
	if created.SourcePath != "sources/video-youtube/runner-demo.md" ||
		created.SourceURL != videoURL ||
		created.AssetPath != "assets/video-youtube/runner-demo.json" ||
		created.TranscriptPolicy != "supplied" ||
		created.TranscriptOrigin != "manual_fixture" ||
		len(created.TranscriptSHA256) != 64 ||
		len(created.Citations) == 0 {
		t.Fatalf("created video ingestion = %+v", created)
	}
	assetBytes, err := os.ReadFile(filepath.Join(vaultRoot, "assets", "video-youtube", "runner-demo.json"))
	if err != nil {
		t.Fatalf("read metadata asset: %v", err)
	}
	if strings.Contains(string(assetBytes), "Initial video transcript evidence.") ||
		!strings.Contains(string(assetBytes), created.TranscriptSHA256) {
		t.Fatalf("metadata asset = %s", string(assetBytes))
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "synthesis/video-runner.md",
		Title: "Video Runner Synthesis",
		Body:  synthesisBody("sources/video-youtube/runner-demo.md", "Initial video transcript evidence."),
	}); err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	_, err = store.IngestVideoURL(ctx, domain.VideoURLInput{
		URL:        "https://www.youtube.com/watch?v=missing",
		Mode:       "update",
		Transcript: domain.VideoTranscriptInput{Text: "Missing update transcript."},
	})
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 404 {
		t.Fatalf("missing update error = %v, want not found 404", err)
	}
	_, err = store.IngestVideoURL(ctx, domain.VideoURLInput{
		URL:      videoURL,
		PathHint: "sources/video-youtube/other.md",
		Mode:     "update",
		Transcript: domain.VideoTranscriptInput{
			Text: "Changed transcript evidence.",
		},
	})
	if !errors.As(err, &appErr) || appErr.Status != 409 || appErr.Code != "conflict" {
		t.Fatalf("mismatched path update error = %v, want conflict 409", err)
	}
	_, err = store.IngestVideoURL(ctx, domain.VideoURLInput{
		URL:           videoURL,
		AssetPathHint: "assets/video-youtube/other.json",
		Mode:          "update",
		Transcript: domain.VideoTranscriptInput{
			Text: "Changed transcript evidence.",
		},
	})
	if !errors.As(err, &appErr) || appErr.Status != 409 || appErr.Code != "conflict" {
		t.Fatalf("mismatched asset update error = %v, want conflict 409", err)
	}
	_, err = store.IngestVideoURL(ctx, domain.VideoURLInput{
		URL:      videoURL,
		PathHint: "sources/video-youtube/duplicate.md",
		Transcript: domain.VideoTranscriptInput{
			Text: "Duplicate transcript.",
		},
	})
	if !errors.As(err, &appErr) || appErr.Status != 409 || appErr.Code != "already_exists" {
		t.Fatalf("duplicate create error = %v, want already exists 409", err)
	}

	beforeSourceEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "source", RefID: created.DocID, Limit: 20})
	if err != nil {
		t.Fatalf("source events before no-op: %v", err)
	}
	beforeProjectionEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "projection", Limit: 50})
	if err != nil {
		t.Fatalf("projection events before no-op: %v", err)
	}
	same, err := store.IngestVideoURL(ctx, domain.VideoURLInput{
		URL:  videoURL,
		Mode: "update",
		Transcript: domain.VideoTranscriptInput{
			Text: "Initial video transcript evidence.",
		},
	})
	if err != nil {
		t.Fatalf("same transcript update: %v", err)
	}
	if same.DocID != created.DocID ||
		same.SourcePath != created.SourcePath ||
		same.AssetPath != created.AssetPath ||
		same.TranscriptSHA256 != created.TranscriptSHA256 ||
		same.PreviousTranscriptSHA256 != "" ||
		same.NewTranscriptSHA256 != "" ||
		len(same.Citations) == 0 {
		t.Fatalf("same transcript update result = %+v, want existing ingestion result", same)
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
		t.Fatalf("same transcript update created stale-state churn: source=%+v projection=%+v", afterSourceEvents.Events, afterProjectionEvents.Events)
	}

	time.Sleep(time.Millisecond)
	changed, err := store.IngestVideoURL(ctx, domain.VideoURLInput{
		URL:           videoURL,
		PathHint:      "sources/video-youtube/runner-demo.md",
		AssetPathHint: "assets/video-youtube/runner-demo.json",
		Mode:          "update",
		Transcript: domain.VideoTranscriptInput{
			Text:       "Updated video transcript evidence.",
			Policy:     "local_first",
			Origin:     "reviewed_local_fixture",
			Language:   "en",
			CapturedAt: "2026-04-27T11:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("changed transcript update: %v", err)
	}
	if changed.DocID != created.DocID ||
		changed.SourcePath != created.SourcePath ||
		changed.AssetPath != created.AssetPath ||
		changed.TranscriptSHA256 == created.TranscriptSHA256 ||
		changed.PreviousTranscriptSHA256 != created.TranscriptSHA256 ||
		changed.NewTranscriptSHA256 != changed.TranscriptSHA256 ||
		changed.TranscriptPolicy != "local_first" ||
		changed.TranscriptOrigin != "reviewed_local_fixture" {
		t.Fatalf("changed update = %+v, created = %+v", changed, created)
	}
	updatedDoc, err := store.GetDocument(ctx, created.DocID)
	if err != nil {
		t.Fatalf("get updated video source: %v", err)
	}
	if updatedDoc.Metadata["transcript_sha256"] != changed.TranscriptSHA256 ||
		updatedDoc.Metadata["transcript_policy"] != "local_first" ||
		!strings.Contains(updatedDoc.Body, "Updated video transcript evidence.") {
		t.Fatalf("updated video source document = %+v", updatedDoc)
	}
	search, err := store.Search(ctx, domain.SearchQuery{Text: "Updated video transcript evidence", PathPrefix: "sources/", Limit: 10})
	if err != nil {
		t.Fatalf("search updated transcript: %v", err)
	}
	if len(search.Hits) == 0 || search.Hits[0].Citations[0].DocID != created.DocID {
		t.Fatalf("search updated video source = %+v", search)
	}
	sourceEvents, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "source", RefID: created.DocID, Limit: 20})
	if err != nil {
		t.Fatalf("source update events: %v", err)
	}
	var foundUpdate bool
	for _, event := range sourceEvents.Events {
		if event.EventType == "source_updated" &&
			event.Details["previous_transcript_sha256"] == created.TranscriptSHA256 &&
			event.Details["new_transcript_sha256"] == changed.TranscriptSHA256 &&
			event.Details["asset_path"] == created.AssetPath &&
			event.Details["source_url"] == videoURL {
			foundUpdate = true
		}
	}
	if !foundUpdate {
		t.Fatalf("source update provenance = %+v", sourceEvents.Events)
	}
	projection := requireSynthesisProjection(t, ctx, store, docIDForPath("synthesis/video-runner.md"))
	if projection.Freshness != "stale" || projection.Details["stale_source_refs"] != "sources/video-youtube/runner-demo.md" {
		t.Fatalf("synthesis projection after video source update = %+v", projection)
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

func requireNoOpSourceUpdateImpact(t *testing.T, result domain.SourceIngestionResult, sourceURL string, sourceDocID string, sha string) {
	t.Helper()
	if result.UpdateStatus != "no_op" ||
		result.NormalizedSourceURL != sourceURL ||
		result.SourceDocID != sourceDocID ||
		result.PreviousSHA256 != sha ||
		result.NewSHA256 != sha ||
		result.Changed ||
		result.SynthesisRepaired {
		t.Fatalf("no-op update impact = %+v", result)
	}
	if result.DuplicateStatus != "existing_source_matched_no_duplicate_created" {
		t.Fatalf("duplicate status = %q", result.DuplicateStatus)
	}
	if len(result.StaleDependents) != 0 || len(result.ProjectionRefs) != 0 || len(result.ProvenanceRefs) != 0 {
		t.Fatalf("no-op update reported stale-impact churn: dependents=%+v projections=%+v provenance=%+v", result.StaleDependents, result.ProjectionRefs, result.ProvenanceRefs)
	}
	if !strings.Contains(result.NoRepairWarning, "no content change") || !strings.Contains(result.NoRepairWarning, "no synthesis repair") {
		t.Fatalf("no-op warning = %q", result.NoRepairWarning)
	}
}

func requireChangedSourceUpdateImpact(t *testing.T, result domain.SourceIngestionResult, sourceURL string, sourceDocID string, previousSHA string, newSHA string, synthesisPath string, sourcePath string) {
	t.Helper()
	if result.UpdateStatus != "changed" ||
		result.NormalizedSourceURL != sourceURL ||
		result.SourceDocID != sourceDocID ||
		result.PreviousSHA256 != previousSHA ||
		result.NewSHA256 != newSHA ||
		!result.Changed ||
		result.SynthesisRepaired {
		t.Fatalf("changed update impact = %+v", result)
	}
	if result.DuplicateStatus != "existing_source_matched_no_duplicate_created" {
		t.Fatalf("duplicate status = %q", result.DuplicateStatus)
	}
	if !sourceImpactHasStaleDependent(result.StaleDependents, synthesisPath, sourcePath) {
		t.Fatalf("stale dependents = %+v, want %s stale on %s", result.StaleDependents, synthesisPath, sourcePath)
	}
	if !sourceImpactHasProjectionRef(result.ProjectionRefs, "synthesis", docIDForPath(synthesisPath)) {
		t.Fatalf("projection refs = %+v, want synthesis %s", result.ProjectionRefs, docIDForPath(synthesisPath))
	}
	if !sourceImpactHasProvenanceRef(result.ProvenanceRefs, "source_updated", sourceDocID) ||
		!sourceImpactHasProvenanceRef(result.ProvenanceRefs, "projection_invalidated", "synthesis:"+docIDForPath(synthesisPath)) {
		t.Fatalf("provenance refs = %+v", result.ProvenanceRefs)
	}
	if !strings.Contains(result.NoRepairWarning, synthesisPath) || !strings.Contains(result.NoRepairWarning, "did not repair") {
		t.Fatalf("no-repair warning = %q", result.NoRepairWarning)
	}
}

func sourceImpactHasStaleDependent(dependents []domain.SourceStaleDependent, synthesisPath string, sourcePath string) bool {
	for _, dependent := range dependents {
		if dependent.Path == synthesisPath &&
			dependent.DocID == docIDForPath(synthesisPath) &&
			dependent.Projection == "synthesis" &&
			dependent.Freshness == "stale" &&
			sourcePathListContains(dependent.StaleSourceRefs, sourcePath) {
			return true
		}
	}
	return false
}

func sourceImpactHasProjectionRef(refs []domain.SourceProjectionRef, projection string, refID string) bool {
	for _, ref := range refs {
		if ref.Projection == projection && ref.RefID == refID && ref.Freshness == "stale" {
			return true
		}
	}
	return false
}

func sourceImpactHasProvenanceRef(refs []domain.SourceProvenanceRef, eventType string, refID string) bool {
	for _, ref := range refs {
		if ref.EventType == eventType && ref.RefID == refID {
			return true
		}
	}
	return false
}
