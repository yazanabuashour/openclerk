package sqlite

import (
	"bytes"
	"context"
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
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
