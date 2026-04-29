package runner_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestDocumentTaskCreateListGetAndUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	create, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "notes/projects/roadmap.md",
			Title: "Roadmap",
			Body:  "# Roadmap\n\n## Summary\nCanonical project note.\n",
		},
	})
	if err != nil {
		t.Fatalf("create document task: %v", err)
	}
	if create.Rejected || create.Document == nil || create.Document.DocID == "" {
		t.Fatalf("create result = %+v", create)
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "notes/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list document task: %v", err)
	}
	if len(list.Documents) != 1 || list.Documents[0].Path != "notes/projects/roadmap.md" {
		t.Fatalf("list result = %+v", list)
	}

	appendResult, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionAppend,
		DocID:   create.Document.DocID,
		Content: "## Decisions\nUse the OpenClerk runner.\n",
	})
	if err != nil {
		t.Fatalf("append document task: %v", err)
	}
	if appendResult.Document == nil || !strings.Contains(appendResult.Document.Body, "OpenClerk runner") {
		t.Fatalf("append result = %+v", appendResult)
	}

	replace, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   create.Document.DocID,
		Heading: "Decisions",
		Content: "Use `openclerk` for routine agent work.",
	})
	if err != nil {
		t.Fatalf("replace document task: %v", err)
	}
	if replace.Document == nil ||
		!strings.Contains(replace.Document.Body, "openclerk") ||
		strings.Contains(replace.Document.Body, "OpenClerk runner") {
		t.Fatalf("replace result body = %q", replace.Document.Body)
	}

	cleared, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   create.Document.DocID,
		Heading: "Decisions",
		Content: "",
	})
	if err != nil {
		t.Fatalf("clear section task: %v", err)
	}
	if cleared.Document == nil ||
		!strings.Contains(cleared.Document.Body, "## Decisions") ||
		strings.Contains(cleared.Document.Body, "openclerk") {
		t.Fatalf("cleared section body = %q", cleared.Document.Body)
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  create.Document.DocID,
	})
	if err != nil {
		t.Fatalf("get document task: %v", err)
	}
	if get.Document == nil || get.Document.Path != create.Document.Path {
		t.Fatalf("get result = %+v", get)
	}

	if _, err := json.Marshal(get); err != nil {
		t.Fatalf("marshal document task result: %v", err)
	}
}

func TestDocumentTaskRejectsInvalidCreateFrontmatterBeforeRuntimeFiles(t *testing.T) {
	t.Parallel()

	dataDir := filepath.Join(t.TempDir(), "data")
	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(dataDir, "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "sources/uploaded-pdf.md",
			Title: "Uploaded PDF",
			Body: strings.TrimSpace(`---
type: source
modality: pdf
---
# Uploaded PDF

## Summary
Extracted note.
`) + "\n",
		},
	})
	if err != nil {
		t.Fatalf("document task: %v", err)
	}
	if !result.Rejected || !strings.Contains(result.RejectionReason, "modality") || !strings.Contains(result.RejectionReason, "markdown") {
		t.Fatalf("result = %+v, want modality rejection", result)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Fatalf("data dir exists after validation rejection: %v", err)
	}
}

func TestDocumentTaskAllowsMarkdownSourceWithPDFSourceType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "sources/uploaded-pdf.md",
			Title: "Uploaded PDF",
			Body: strings.TrimSpace(`---
type: source
source_type: pdf
modality: markdown
---
# Uploaded PDF

## Summary
Markdown notes extracted from a PDF source.
`) + "\n",
		},
	})
	if err != nil {
		t.Fatalf("document task: %v", err)
	}
	if result.Rejected || result.Document == nil {
		t.Fatalf("result = %+v, want created source document", result)
	}
	if result.Document.Metadata["source_type"] != "pdf" || result.Document.Metadata["modality"] != "markdown" {
		t.Fatalf("metadata = %+v", result.Document.Metadata)
	}
}

func TestDocumentTaskRejectsInvalidSourceURLIngestBeforeRuntimeFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  runner.SourceURLInput
		wantErr string
	}{
		{
			name: "missing url",
			source: runner.SourceURLInput{
				PathHint:      "sources/uploaded-pdf.md",
				AssetPathHint: "assets/sources/uploaded-pdf.pdf",
			},
			wantErr: "source.url is required",
		},
		{
			name: "invalid scheme",
			source: runner.SourceURLInput{
				URL:           "file:///tmp/uploaded-pdf.pdf",
				PathHint:      "sources/uploaded-pdf.md",
				AssetPathHint: "assets/sources/uploaded-pdf.pdf",
			},
			wantErr: "http or https",
		},
		{
			name: "unsafe source path",
			source: runner.SourceURLInput{
				URL:           "https://example.test/uploaded-pdf.pdf",
				PathHint:      "../uploaded-pdf.md",
				AssetPathHint: "assets/sources/uploaded-pdf.pdf",
			},
			wantErr: "source.path_hint",
		},
		{
			name: "unsafe asset path",
			source: runner.SourceURLInput{
				URL:           "https://example.test/uploaded-pdf.pdf",
				PathHint:      "sources/uploaded-pdf.md",
				AssetPathHint: "../uploaded-pdf.pdf",
			},
			wantErr: "source.asset_path_hint",
		},
		{
			name: "invalid mode",
			source: runner.SourceURLInput{
				URL:           "https://example.test/uploaded-pdf.pdf",
				PathHint:      "sources/uploaded-pdf.md",
				AssetPathHint: "assets/sources/uploaded-pdf.pdf",
				Mode:          "replace",
			},
			wantErr: "source.mode",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dataDir := filepath.Join(t.TempDir(), "data")
			result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(dataDir, "openclerk.sqlite")}, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionIngestSourceURL,
				Source: tt.source,
			})
			if err != nil {
				t.Fatalf("document task: %v", err)
			}
			if !result.Rejected || !strings.Contains(result.RejectionReason, tt.wantErr) {
				t.Fatalf("result = %+v, want rejection containing %q", result, tt.wantErr)
			}
			if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
				t.Fatalf("data dir exists after validation rejection: %v", err)
			}
		})
	}
}

func TestDocumentTaskRejectsInvalidVideoURLIngestBeforeRuntimeFiles(t *testing.T) {
	t.Parallel()

	validTranscript := runner.VideoTranscriptInput{Text: "Supplied transcript evidence.", Policy: "supplied"}
	tests := []struct {
		name    string
		video   runner.VideoURLInput
		wantErr string
	}{
		{
			name: "missing url",
			video: runner.VideoURLInput{
				PathHint:   "sources/video-youtube/uploaded.md",
				Transcript: validTranscript,
			},
			wantErr: "video.url is required",
		},
		{
			name: "unsafe source path",
			video: runner.VideoURLInput{
				URL:        "https://www.youtube.com/watch?v=openclerk",
				PathHint:   "../uploaded.md",
				Transcript: validTranscript,
			},
			wantErr: "video.path_hint",
		},
		{
			name: "non source markdown path",
			video: runner.VideoURLInput{
				URL:        "https://www.youtube.com/watch?v=openclerk",
				PathHint:   "notes/uploaded.md",
				Transcript: validTranscript,
			},
			wantErr: "sources/*.md",
		},
		{
			name: "invalid mode",
			video: runner.VideoURLInput{
				URL:        "https://www.youtube.com/watch?v=openclerk",
				PathHint:   "sources/video-youtube/uploaded.md",
				Mode:       "replace",
				Transcript: validTranscript,
			},
			wantErr: "video.mode",
		},
		{
			name: "missing transcript",
			video: runner.VideoURLInput{
				URL:      "https://www.youtube.com/watch?v=openclerk",
				PathHint: "sources/video-youtube/uploaded.md",
			},
			wantErr: "video.transcript.text",
		},
		{
			name: "unsupported policy",
			video: runner.VideoURLInput{
				URL:      "https://www.youtube.com/watch?v=openclerk",
				PathHint: "sources/video-youtube/uploaded.md",
				Transcript: runner.VideoTranscriptInput{
					Text:   "Supplied transcript evidence.",
					Policy: "platform_caption",
				},
			},
			wantErr: "video.transcript.policy",
		},
		{
			name: "unsafe asset path",
			video: runner.VideoURLInput{
				URL:           "https://www.youtube.com/watch?v=openclerk",
				PathHint:      "sources/video-youtube/uploaded.md",
				AssetPathHint: "../uploaded.json",
				Transcript:    validTranscript,
			},
			wantErr: "video.asset_path_hint",
		},
		{
			name: "non json asset path",
			video: runner.VideoURLInput{
				URL:           "https://www.youtube.com/watch?v=openclerk",
				PathHint:      "sources/video-youtube/uploaded.md",
				AssetPathHint: "assets/video-youtube/uploaded.txt",
				Transcript:    validTranscript,
			},
			wantErr: "assets/**/*.json",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dataDir := filepath.Join(t.TempDir(), "data")
			result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(dataDir, "openclerk.sqlite")}, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionIngestVideoURL,
				Video:  tt.video,
			})
			if err != nil {
				t.Fatalf("document task: %v", err)
			}
			if !result.Rejected || !strings.Contains(result.RejectionReason, tt.wantErr) {
				t.Fatalf("result = %+v, want rejection containing %q", result, tt.wantErr)
			}
			if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
				t.Fatalf("data dir exists after validation rejection: %v", err)
			}
		})
	}
}

func TestDocumentTaskIngestSourceURLPDF(t *testing.T) {
	t.Parallel()

	pdfBytes := minimalPDF("Runner Intake PDF Title", "OpenClerk Test", "Runner intake unique text")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(pdfBytes)
	}))
	t.Cleanup(server.Close)

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           server.URL + "/runner.pdf",
			PathHint:      "sources/runner-ingest.md",
			AssetPathHint: "assets/sources/runner-ingest.pdf",
			Title:         "Runner Ingest Override",
		},
	})
	if err != nil {
		t.Fatalf("ingest source URL: %v", err)
	}
	if result.Rejected || result.Ingestion == nil {
		t.Fatalf("ingest result = %+v", result)
	}
	ingestion := result.Ingestion
	if ingestion.DocID == "" ||
		ingestion.SourcePath != "sources/runner-ingest.md" ||
		ingestion.AssetPath != "assets/sources/runner-ingest.pdf" ||
		ingestion.DerivedPath != "sources/runner-ingest.md" ||
		ingestion.PageCount != 1 ||
		ingestion.SizeBytes != int64(len(pdfBytes)) ||
		ingestion.MIMEType != "application/pdf" ||
		len(ingestion.Citations) == 0 ||
		len(ingestion.SHA256) != 64 {
		t.Fatalf("ingestion = %+v", ingestion)
	}
	if ingestion.PDFMetadata.Title != "Runner Intake PDF Title" || ingestion.PDFMetadata.Author != "OpenClerk Test" {
		t.Fatalf("pdf metadata = %+v", ingestion.PDFMetadata)
	}
	assetPath := filepath.Join(filepath.Dir(dbPath), "vault", "assets", "sources", "runner-ingest.pdf")
	if _, err := os.Stat(assetPath); err != nil {
		t.Fatalf("asset stat: %v", err)
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  ingestion.DocID,
	})
	if err != nil {
		t.Fatalf("get ingested document: %v", err)
	}
	if get.Document == nil ||
		get.Document.Metadata["source_url"] != server.URL+"/runner.pdf" ||
		get.Document.Metadata["asset_path"] != "assets/sources/runner-ingest.pdf" ||
		get.Document.Metadata["source_type"] != "pdf" ||
		get.Document.Metadata["mime_type"] != "application/pdf" ||
		!strings.Contains(get.Document.Body, "Runner Ingest Override") ||
		!strings.Contains(get.Document.Body, "Runner Intake PDF Title") ||
		!strings.Contains(get.Document.Body, "Runnerintakeuniquetext") {
		t.Fatalf("ingested document = %+v", get.Document)
	}

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "Runner Intake PDF Title",
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("search ingested text: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || search.Search.Hits[0].Citations[0].Path != "sources/runner-ingest.md" {
		t.Fatalf("search = %+v", search.Search)
	}

	provenance, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "source",
			RefID:   ingestion.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("source provenance: %v", err)
	}
	if provenance.Provenance == nil || len(provenance.Provenance.Events) == 0 {
		t.Fatalf("provenance = %+v", provenance.Provenance)
	}
	if got := provenance.Provenance.Events[0].Details["source_url"]; got != server.URL+"/runner.pdf" {
		t.Fatalf("source provenance details = %+v", provenance.Provenance.Events[0].Details)
	}

	sameUpdate, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:  server.URL + "/runner.pdf",
			Mode: "update",
		},
	})
	if err != nil {
		t.Fatalf("same source URL update: %v", err)
	}
	if sameUpdate.Rejected || sameUpdate.Ingestion == nil ||
		sameUpdate.Ingestion.DocID != ingestion.DocID ||
		sameUpdate.Ingestion.SourcePath != ingestion.SourcePath ||
		sameUpdate.Ingestion.AssetPath != ingestion.AssetPath ||
		sameUpdate.Ingestion.SHA256 != ingestion.SHA256 {
		t.Fatalf("same source URL update = %+v, want existing ingestion", sameUpdate)
	}

	duplicate, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           server.URL + "/runner.pdf",
			PathHint:      "sources/runner-duplicate.md",
			AssetPathHint: "assets/sources/runner-duplicate.pdf",
		},
	})
	if err == nil {
		t.Fatalf("duplicate result = %+v, want error", duplicate)
	}
	if !strings.Contains(err.Error(), "source URL") {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestDocumentTaskIngestVideoURLSuppliedTranscript(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	videoURL := "https://www.youtube.com/watch?v=openclerk-demo"
	transcript := "OpenClerk video canonical transcript unique evidence."
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestVideoURL,
		Video: runner.VideoURLInput{
			URL:           videoURL,
			PathHint:      "sources/video-youtube/runner-demo.md",
			AssetPathHint: "assets/video-youtube/runner-demo.json",
			Title:         "Runner Video Demo",
			Transcript: runner.VideoTranscriptInput{
				Text:       transcript,
				Policy:     "supplied",
				Origin:     "user_fixture",
				Language:   "en",
				CapturedAt: "2026-04-27T10:00:00Z",
				Tool:       "manual",
				Model:      "none",
			},
		},
	})
	if err != nil {
		t.Fatalf("ingest video URL: %v", err)
	}
	if result.Rejected || result.VideoIngestion == nil {
		t.Fatalf("ingest result = %+v", result)
	}
	ingestion := result.VideoIngestion
	if ingestion.DocID == "" ||
		ingestion.SourcePath != "sources/video-youtube/runner-demo.md" ||
		ingestion.SourceURL != videoURL ||
		ingestion.AssetPath != "assets/video-youtube/runner-demo.json" ||
		ingestion.TranscriptPolicy != "supplied" ||
		ingestion.TranscriptOrigin != "user_fixture" ||
		ingestion.Language != "en" ||
		ingestion.Tool != "manual" ||
		ingestion.Model != "none" ||
		len(ingestion.TranscriptSHA256) != 64 ||
		len(ingestion.Citations) == 0 {
		t.Fatalf("ingestion = %+v", ingestion)
	}
	assetPath := filepath.Join(filepath.Dir(dbPath), "vault", "assets", "video-youtube", "runner-demo.json")
	assetBytes, err := os.ReadFile(assetPath)
	if err != nil {
		t.Fatalf("asset read: %v", err)
	}
	if strings.Contains(string(assetBytes), transcript) || !strings.Contains(string(assetBytes), ingestion.TranscriptSHA256) {
		t.Fatalf("metadata asset = %s", string(assetBytes))
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  ingestion.DocID,
	})
	if err != nil {
		t.Fatalf("get ingested video document: %v", err)
	}
	if get.Document == nil ||
		get.Document.Metadata["source_type"] != "video_transcript" ||
		get.Document.Metadata["source_url"] != videoURL ||
		get.Document.Metadata["transcript_sha256"] != ingestion.TranscriptSHA256 ||
		get.Document.Metadata["asset_path"] != "assets/video-youtube/runner-demo.json" ||
		!strings.Contains(get.Document.Body, "Runner Video Demo") ||
		!strings.Contains(get.Document.Body, transcript) {
		t.Fatalf("ingested document = %+v", get.Document)
	}

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "canonical transcript unique evidence",
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("search ingested transcript: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || search.Search.Hits[0].Citations[0].Path != "sources/video-youtube/runner-demo.md" {
		t.Fatalf("search = %+v", search.Search)
	}

	provenance, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "source",
			RefID:   ingestion.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("source provenance: %v", err)
	}
	if provenance.Provenance == nil || len(provenance.Provenance.Events) == 0 {
		t.Fatalf("provenance = %+v", provenance.Provenance)
	}
	details := provenance.Provenance.Events[0].Details
	if details["source_url"] != videoURL || details["transcript_sha256"] != ingestion.TranscriptSHA256 {
		t.Fatalf("source provenance details = %+v", details)
	}
}

func TestDocumentTaskRejectsNonPDFSourceURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("not a PDF"))
	}))
	t.Cleanup(server.Close)

	_, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           server.URL + "/not-pdf.txt",
			PathHint:      "sources/not-pdf.md",
			AssetPathHint: "assets/sources/not-pdf.pdf",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "PDF") {
		t.Fatalf("err = %v, want non-PDF rejection", err)
	}
}

func TestRetrievalTaskSynthesisFreshnessProjectionAndProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nInitial source guidance.\n")
	synthesis := createDocument(t, ctx, config, "synthesis/runner.md", "Runner Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner.md
---
# Runner Synthesis

## Summary
Initial source guidance.

## Sources
- sources/runner.md

## Freshness
Checked source refs.
`)+"\n")

	time.Sleep(time.Millisecond)
	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   source.DocID,
		Heading: "Summary",
		Content: "Updated source guidance.",
	})
	if err != nil {
		t.Fatalf("update source: %v", err)
	}
	if updated.Document == nil {
		t.Fatalf("update result = %+v", updated)
	}

	projections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      synthesis.DocID,
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("synthesis projection task: %v", err)
	}
	if projections.Projections == nil ||
		len(projections.Projections.Projections) != 1 ||
		projections.Projections.Projections[0].Freshness != "stale" ||
		projections.Projections.Projections[0].Details["stale_source_refs"] != "sources/runner.md" {
		t.Fatalf("synthesis projections result = %+v", projections)
	}

	sourceEvents, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "source",
			RefID:   source.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("source provenance task: %v", err)
	}
	if sourceEvents.Provenance == nil ||
		!runnerEventTypesInclude(sourceEvents.Provenance.Events, "source_created") ||
		!runnerEventTypesInclude(sourceEvents.Provenance.Events, "source_updated") {
		t.Fatalf("source provenance result = %+v", sourceEvents)
	}

	synthesisEvents, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + synthesis.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("synthesis provenance task: %v", err)
	}
	if synthesisEvents.Provenance == nil || !runnerEventTypesInclude(synthesisEvents.Provenance.Events, "projection_invalidated") {
		t.Fatalf("synthesis provenance result = %+v", synthesisEvents)
	}
}

func TestRetrievalTaskAuditContradictionsPlansAndRepairsExistingSynthesis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	oldSource := createDocument(t, ctx, config, "sources/audit-runner-old.md", "Old audit runner source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-runner-current.md
---
# Old audit runner source

## Summary
Older source-sensitive audit runner repair evidence said agents should prefer a legacy command-path workaround.
`)+"\n")
	currentSource := createDocument(t, ctx, config, "sources/audit-runner-current.md", "Current audit runner source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-runner-old.md
---
# Current audit runner source

## Summary
Current source-sensitive audit runner repair evidence says use the installed openclerk JSON runner.
`)+"\n")
	synthesis := createDocument(t, ctx, config, "synthesis/audit-runner-routing.md", "Audit runner routing", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/audit-runner-current.md, sources/audit-runner-old.md
---
# Audit runner routing

## Summary
Stale audit claim: agents should prefer a legacy command-path workaround.

## Sources
- sources/audit-runner-current.md
- sources/audit-runner-old.md

## Freshness
Checked source refs.
`)+"\n")
	createDocument(t, ctx, config, "synthesis/audit-runner-routing-decoy.md", "Audit runner routing decoy", "# Audit runner routing decoy\n\n## Summary\nDecoy.\n")
	createDocument(t, ctx, config, "sources/audit-conflict-alpha.md", "Audit conflict alpha", strings.TrimSpace(`---
type: source
status: active
---
# Audit conflict alpha

## Summary
Source sensitive audit conflict runner retention is seven days.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-conflict-bravo.md", "Audit conflict bravo", strings.TrimSpace(`---
type: source
status: active
---
# Audit conflict bravo

## Summary
Source sensitive audit conflict runner retention is thirty days.
`)+"\n")

	time.Sleep(time.Millisecond)
	_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   currentSource.DocID,
		Heading: "Summary",
		Content: "Current source-sensitive audit runner repair evidence says use the installed openclerk JSON runner for audit repairs.",
	})
	if err != nil {
		t.Fatalf("update current source: %v", err)
	}

	plan, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit runner repair evidence",
			TargetPath:    "synthesis/audit-runner-routing.md",
			Mode:          "plan_only",
			ConflictQuery: "source sensitive audit conflict runner retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit plan: %v", err)
	}
	if plan.Audit == nil ||
		plan.Audit.RepairStatus != "planned" ||
		plan.Audit.RepairApplied ||
		plan.Audit.SelectedTargetPath != "synthesis/audit-runner-routing.md" ||
		!containsString(plan.Audit.CandidateSynthesisPaths, "synthesis/audit-runner-routing-decoy.md") ||
		!containsString(plan.Audit.CurrentSourcePaths, currentSource.Path) ||
		!containsString(plan.Audit.SupersededSourcePaths, oldSource.Path) ||
		len(plan.Audit.ProjectionFreshnessBefore) == 0 ||
		len(plan.Audit.ProjectionFreshnessAfter) == 0 ||
		len(plan.Audit.UnresolvedConflictGroups) != 1 ||
		plan.Audit.UnresolvedConflictGroups[0].Status != "unresolved" {
		t.Fatalf("audit plan result = %+v", plan.Audit)
	}
	if !auditInspectedPath(plan.Audit.ProvenanceInspected, "sources/audit-conflict-alpha.md") ||
		!auditInspectedPath(plan.Audit.ProvenanceInspected, "sources/audit-conflict-bravo.md") {
		t.Fatalf("audit plan did not inspect conflict provenance: %+v", plan.Audit.ProvenanceInspected)
	}
	unchanged, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get unchanged synthesis: %v", err)
	}
	if !strings.Contains(unchanged.Document.Body, "legacy command-path workaround") {
		t.Fatalf("plan_only changed synthesis body = %q", unchanged.Document.Body)
	}

	repaired, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit runner repair evidence",
			TargetPath:    "synthesis/audit-runner-routing.md",
			Mode:          "repair_existing",
			ConflictQuery: "source sensitive audit conflict runner retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit repair: %v", err)
	}
	if repaired.Audit == nil ||
		repaired.Audit.RepairStatus != "applied" ||
		!repaired.Audit.RepairApplied ||
		repaired.Audit.DuplicatePrevention != "existing_target_selected_no_duplicate_created" ||
		repaired.Audit.FailureClassification != "none" {
		t.Fatalf("audit repair result = %+v", repaired.Audit)
	}
	if len(repaired.Audit.ProjectionFreshnessAfter) == 0 || repaired.Audit.ProjectionFreshnessAfter[0].Freshness != "fresh" {
		t.Fatalf("projection after repair = %+v", repaired.Audit.ProjectionFreshnessAfter)
	}
	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get repaired synthesis: %v", err)
	}
	for _, want := range []string{
		"source_refs: sources/audit-runner-current.md, sources/audit-runner-old.md",
		"Current audit guidance: use the installed openclerk JSON runner.",
		"Current source: sources/audit-runner-current.md.",
		"Superseded source: sources/audit-runner-old.md.",
		"## Sources",
		"## Freshness",
	} {
		if !strings.Contains(updated.Document.Body, want) {
			t.Fatalf("repaired body missing %q:\n%s", want, updated.Document.Body)
		}
	}
}

func TestRetrievalTaskAuditContradictionsFindsTargetAfterFirstSynthesisPage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/audit-page-old.md", "Old audit page source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-page-current.md
---
# Old audit page source

## Summary
Older source-sensitive audit page evidence said to use a legacy command path.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-page-current.md", "Current audit page source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-page-old.md
---
# Current audit page source

## Summary
Current source-sensitive audit page evidence says use the installed openclerk JSON runner.
`)+"\n")
	for i := 0; i < 101; i++ {
		createDocument(t, ctx, config, fmt.Sprintf("synthesis/aa-audit-decoy-%03d.md", i), fmt.Sprintf("Audit decoy %03d", i), fmt.Sprintf("# Audit decoy %03d\n\n## Summary\nDecoy.\n", i))
	}
	target := createDocument(t, ctx, config, "synthesis/zz-audit-page-target.md", "Audit page target", strings.TrimSpace(`---
type: synthesis
status: active
source_refs: sources/audit-page-current.md, sources/audit-page-old.md
---
# Audit page target

## Summary
Stale audit page claim.

## Sources
- sources/audit-page-current.md
- sources/audit-page-old.md

## Freshness
Checked source refs.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:      "source-sensitive audit page evidence",
			TargetPath: target.Path,
			Mode:       "plan_only",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("audit paginated target: %v", err)
	}
	if result.Audit == nil ||
		result.Audit.SelectedTargetPath != target.Path ||
		result.Audit.RepairStatus != "planned" ||
		result.Audit.DuplicatePrevention != "existing_target_selected_no_duplicate_created" {
		t.Fatalf("audit paginated target result = %+v", result.Audit)
	}
}

func TestRetrievalTaskAuditContradictionsDoesNotReportMatchingCurrentSourcesAsConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/audit-agree-old.md", "Old audit agree source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-agree-current.md
---
# Old audit agree source

## Summary
Older source-sensitive audit agreement evidence said to use a legacy command path.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-agree-current.md", "Current audit agree source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-agree-old.md
---
# Current audit agree source

## Summary
Current source-sensitive audit agreement evidence says use the installed openclerk JSON runner.
`)+"\n")
	createDocument(t, ctx, config, "synthesis/audit-agree-target.md", "Audit agree target", strings.TrimSpace(`---
type: synthesis
status: active
source_refs: sources/audit-agree-current.md, sources/audit-agree-old.md
---
# Audit agree target

## Summary
Stale audit agreement claim.

## Sources
- sources/audit-agree-current.md
- sources/audit-agree-old.md

## Freshness
Checked source refs.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-retention-alpha.md", "Audit retention alpha", strings.TrimSpace(`---
type: source
status: active
---
# Audit retention alpha

## Summary
Source sensitive audit matching retention is seven days.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-retention-bravo.md", "Audit retention bravo", strings.TrimSpace(`---
type: source
status: active
---
# Audit retention bravo

## Summary
Source sensitive audit matching retention is seven days.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit agreement evidence",
			TargetPath:    "synthesis/audit-agree-target.md",
			Mode:          "plan_only",
			ConflictQuery: "source sensitive audit matching retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit matching conflict: %v", err)
	}
	if result.Audit == nil || len(result.Audit.UnresolvedConflictGroups) != 0 {
		t.Fatalf("matching current sources reported as conflict = %+v", result.Audit)
	}
	if !auditInspectedPath(result.Audit.ProvenanceInspected, "sources/audit-retention-alpha.md") ||
		!auditInspectedPath(result.Audit.ProvenanceInspected, "sources/audit-retention-bravo.md") {
		t.Fatalf("audit matching conflict did not inspect provenance: %+v", result.Audit.ProvenanceInspected)
	}
}

func TestRetrievalTaskAuditContradictionsValidation(t *testing.T) {
	t.Parallel()

	missing, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			TargetPath: "synthesis/audit-runner-routing.md",
		},
	})
	if err != nil {
		t.Fatalf("missing audit query validation: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "audit.query is required" {
		t.Fatalf("missing result = %+v", missing)
	}

	invalidMode, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:      "source-sensitive audit runner repair evidence",
			TargetPath: "synthesis/audit-runner-routing.md",
			Mode:       "create_new",
		},
	})
	if err != nil {
		t.Fatalf("invalid audit mode validation: %v", err)
	}
	if !invalidMode.Rejected || invalidMode.RejectionReason != "audit.mode must be plan_only or repair_existing" {
		t.Fatalf("invalid mode result = %+v", invalidMode)
	}
}

func TestDocumentTaskInspectLayout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nCanonical source guidance.\n")
	createDocument(t, ctx, config, "synthesis/runner.md", "Runner Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner.md
---
# Runner Synthesis

## Summary
Canonical source guidance.

## Sources
- sources/runner.md

## Freshness
Checked source refs.
`)+"\n")
	createDocument(t, ctx, config, "records/assets/openclerk-runner.md", "OpenClerk runner record", "---\nentity_type: tool\nentity_name: OpenClerk runner\n---\n# OpenClerk runner record\n\n## Facts\n- status: active\n")
	createDocument(t, ctx, config, "records/services/openclerk-runner.md", "OpenClerk runner", "---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service.\n")
	createDocument(t, ctx, config, "docs/architecture/runner-decision.md", "Runner decision", "---\ndecision_id: adr-runner\ndecision_title: Use JSON runner\ndecision_status: accepted\ndecision_scope: agentops\ndecision_owner: platform\n---\n# Runner decision\n\n## Summary\nUse the JSON runner for AgentOps.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect layout: %v", err)
	}
	if result.Layout == nil || !result.Layout.Valid {
		t.Fatalf("layout = %+v", result.Layout)
	}
	if result.Layout.Mode != "convention_first" || result.Layout.ConfigArtifactRequired {
		t.Fatalf("layout configuration = %+v", result.Layout)
	}
	if !layoutChecksInclude(result.Layout.Checks, "synthesis_source_refs_resolve", "pass") ||
		!layoutChecksInclude(result.Layout.Checks, "service_identity_metadata", "pass") ||
		!layoutChecksInclude(result.Layout.Checks, "decision_identity_metadata", "pass") ||
		!layoutChecksInclude(result.Layout.Checks, "record_identity_metadata", "pass") {
		t.Fatalf("layout checks = %+v", result.Layout.Checks)
	}
}

func TestDocumentTaskInspectLayoutUsesRootRelativeSourceAndSynthesisPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nCanonical source guidance.\n")
	createDocument(t, ctx, config, "synthesis/runner.md", "Runner Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: "sources/runner.md"
---
# Runner Synthesis

## Summary
Canonical source guidance.

## Sources
- sources/runner.md

## Freshness
Checked source refs.
`)+"\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect layout: %v", err)
	}
	if result.Layout == nil || !result.Layout.Valid {
		t.Fatalf("layout = %+v", result.Layout)
	}
	if !layoutPathConventionsInclude(result.Layout.ConventionalPaths, "sources/") ||
		!layoutPathConventionsInclude(result.Layout.ConventionalPaths, "synthesis/") {
		t.Fatalf("layout conventions = %+v", result.Layout.ConventionalPaths)
	}
	if !layoutChecksInclude(result.Layout.Checks, "synthesis_source_refs_resolve", "pass") {
		t.Fatalf("layout checks = %+v", result.Layout.Checks)
	}
	for _, check := range result.Layout.Checks {
		if check.ID != "optional_conventional_prefixes" {
			continue
		}
		prefixes := check.Details["path_prefixes"]
		if strings.Contains(prefixes, "sources/") || strings.Contains(prefixes, "synthesis/") {
			t.Fatalf("optional prefix warning reports populated root paths: %+v", check)
		}
	}
}

func TestDocumentTaskInspectLayoutDoesNotTreatNestedNotesSynthesisPathAsConvention(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/synthesis/legacy.md", "Legacy Synthesis", "# Legacy Synthesis\n\n## Summary\nLegacy nested path.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect layout: %v", err)
	}
	if result.Layout == nil || !result.Layout.Valid {
		t.Fatalf("layout = %+v", result.Layout)
	}
	for _, id := range []string{"synthesis_source_refs", "synthesis_sources_section", "synthesis_freshness_section"} {
		if layoutChecksInclude(result.Layout.Checks, id, "fail") {
			t.Fatalf("layout checks treated nested notes path as synthesis convention: %+v", result.Layout.Checks)
		}
	}
}

func TestDocumentTaskInspectLayoutReportsInvalidConventions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "synthesis/incomplete.md", "Incomplete Synthesis", "# Incomplete Synthesis\n\n## Summary\nMissing evidence conventions.\n")
	createDocument(t, ctx, config, "records/services/incomplete.md", "Incomplete Service", "---\nservice_id: incomplete\n---\n# Incomplete Service\n\n## Summary\nMissing service name.\n")
	createDocument(t, ctx, config, "records/decisions/incomplete.md", "Incomplete Decision", "---\ndecision_id: incomplete\n---\n# Incomplete Decision\n\n## Summary\nMissing decision title and status.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect invalid layout: %v", err)
	}
	if result.Layout == nil || result.Layout.Valid {
		t.Fatalf("layout = %+v, want invalid", result.Layout)
	}
	for _, id := range []string{
		"synthesis_source_refs",
		"synthesis_sources_section",
		"synthesis_freshness_section",
		"service_identity_metadata",
		"decision_identity_metadata",
	} {
		if !layoutChecksInclude(result.Layout.Checks, id, "fail") {
			t.Fatalf("layout checks missing failing %s: %+v", id, result.Layout.Checks)
		}
	}
}

func TestDocumentTaskInspectLayoutRequiresLevelTwoSynthesisSections(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nCanonical source guidance.\n")
	createDocument(t, ctx, config, "synthesis/wrong-levels.md", "Wrong Level Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner.md
---
# Wrong Level Synthesis

## Summary
Canonical source guidance.

# Sources
- sources/runner.md

### Freshness
Checked source refs.
`)+"\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect wrong-level layout: %v", err)
	}
	if result.Layout == nil || result.Layout.Valid {
		t.Fatalf("layout = %+v, want invalid", result.Layout)
	}
	for _, id := range []string{
		"synthesis_sources_section",
		"synthesis_freshness_section",
	} {
		if !layoutChecksInclude(result.Layout.Checks, id, "fail") {
			t.Fatalf("layout checks missing failing %s: %+v", id, result.Layout.Checks)
		}
	}
}

func TestRetrievalTaskSearchLinksRecordsAndProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	architecture := createDocument(t, ctx, config, "notes/architecture/knowledge-plane.md", "Knowledge plane", "# Knowledge plane\n\n## Summary\nCanonical architecture note.\n")
	roadmap := createDocument(t, ctx, config, "notes/projects/roadmap.md", "Roadmap", "# Roadmap\n\n## Summary\nSee the [knowledge plane](../architecture/knowledge-plane.md).\n")
	createDocument(t, ctx, config, "records/assets/transmission-solenoid.md", "Transmission solenoid", "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\n---\n# Transmission solenoid\n\n## Facts\n- sku: SOL-1\n")
	createDocument(t, ctx, config, "records/services/openclerk-runner.md", "OpenClerk runner", "---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service for routine knowledge tasks.\n\n## Facts\n- tier: production\n")
	createDocument(t, ctx, config, "docs/architecture/runner-old-decision.md", "Old runner decision", "---\ndecision_id: adr-runner-old\ndecision_title: Old runner path\ndecision_status: superseded\ndecision_scope: agentops\ndecision_owner: platform\ndecision_date: 2026-04-20\nsuperseded_by: adr-runner-current\nsource_refs: sources/runner-old.md\n---\n# Old runner path\n\n## Summary\nOld decision used a retired runner path.\n")
	createDocument(t, ctx, config, "sources/runner-old.md", "Old runner source", "# Old runner source\n\n## Summary\nRetired runner path source.\n")
	createDocument(t, ctx, config, "notes/architecture/runner-current-decision.md", "Current runner decision", "---\ndecision_id: adr-runner-current\ndecision_title: Use JSON runner\ndecision_status: accepted\ndecision_scope: agentops\ndecision_owner: platform\ndecision_date: 2026-04-22\nsupersedes: adr-runner-old\nsource_refs: sources/runner-current.md\n---\n# Use JSON runner\n\n## Summary\nAccepted decision uses the JSON runner for routine AgentOps work.\n")
	createDocument(t, ctx, config, "sources/runner-current.md", "Current runner source", "# Current runner source\n\n## Summary\nCurrent runner source.\n")

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  "roadmap",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("search task: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || len(search.Search.Hits[0].Citations) == 0 {
		t.Fatalf("search result = %+v", search)
	}

	links, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  roadmap.DocID,
	})
	if err != nil {
		t.Fatalf("links task: %v", err)
	}
	if links.Links == nil || len(links.Links.Outgoing) != 1 || links.Links.Outgoing[0].DocID != architecture.DocID {
		t.Fatalf("links result = %+v", links)
	}

	graph, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraph,
		DocID:  roadmap.DocID,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("graph task: %v", err)
	}
	if graph.Graph == nil || len(graph.Graph.Nodes) == 0 || len(graph.Graph.Edges) == 0 {
		t.Fatalf("graph result = %+v", graph)
	}

	records, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "solenoid", Limit: 10},
	})
	if err != nil {
		t.Fatalf("records task: %v", err)
	}
	if records.Records == nil || len(records.Records.Entities) != 1 {
		t.Fatalf("records result = %+v", records)
	}

	entity, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionRecordEntity,
		EntityID: records.Records.Entities[0].EntityID,
	})
	if err != nil {
		t.Fatalf("record entity task: %v", err)
	}
	if entity.Entity == nil || entity.Entity.EntityID != "transmission-solenoid" {
		t.Fatalf("entity result = %+v", entity)
	}

	services, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionServicesLookup,
		Services: runner.ServiceLookupOptions{
			Text:      "OpenClerk runner",
			Interface: "JSON runner",
			Limit:     10,
		},
	})
	if err != nil {
		t.Fatalf("services task: %v", err)
	}
	if services.Services == nil || len(services.Services.Services) != 1 || services.Services.Services[0].ServiceID != "openclerk-runner" {
		t.Fatalf("services result = %+v", services)
	}

	service, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:    runner.RetrievalTaskActionServiceRecord,
		ServiceID: "openclerk-runner",
	})
	if err != nil {
		t.Fatalf("service record task: %v", err)
	}
	if service.Service == nil ||
		service.Service.Status != "active" ||
		service.Service.Owner != "runner" ||
		service.Service.Interface != "JSON runner" ||
		len(service.Service.Citations) == 0 {
		t.Fatalf("service result = %+v", service)
	}

	decisions, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{
			Text:   "JSON runner",
			Status: "accepted",
			Scope:  "agentops",
			Owner:  "platform",
			Limit:  10,
		},
	})
	if err != nil {
		t.Fatalf("decisions task: %v", err)
	}
	if decisions.Decisions == nil ||
		len(decisions.Decisions.Decisions) != 1 ||
		decisions.Decisions.Decisions[0].DecisionID != "adr-runner-current" ||
		len(decisions.Decisions.Decisions[0].Citations) == 0 {
		t.Fatalf("decisions result = %+v", decisions)
	}

	decision, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: "adr-runner-old",
	})
	if err != nil {
		t.Fatalf("decision record task: %v", err)
	}
	if decision.Decision == nil ||
		decision.Decision.Status != "superseded" ||
		len(decision.Decision.SupersededBy) != 1 ||
		decision.Decision.SupersededBy[0] != "adr-runner-current" ||
		len(decision.Decision.Citations) == 0 {
		t.Fatalf("decision result = %+v", decision)
	}

	provenance, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   roadmap.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("provenance task: %v", err)
	}
	if provenance.Provenance == nil || len(provenance.Provenance.Events) == 0 {
		t.Fatalf("provenance result = %+v", provenance)
	}

	projections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "graph",
			RefKind:    "document",
			RefID:      roadmap.DocID,
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("projection task: %v", err)
	}
	if projections.Projections == nil || len(projections.Projections.Projections) != 1 {
		t.Fatalf("projection result = %+v", projections)
	}

	serviceProjections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "services",
			RefKind:    "service",
			RefID:      "openclerk-runner",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("service projection task: %v", err)
	}
	if serviceProjections.Projections == nil ||
		len(serviceProjections.Projections.Projections) != 1 ||
		serviceProjections.Projections.Projections[0].Freshness != "fresh" {
		t.Fatalf("service projections result = %+v", serviceProjections)
	}

	decisionProjections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-runner-old",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("decision projection task: %v", err)
	}
	if decisionProjections.Projections == nil ||
		len(decisionProjections.Projections.Projections) != 1 ||
		decisionProjections.Projections.Projections[0].Freshness != "stale" ||
		decisionProjections.Projections.Projections[0].Details["superseded_by"] != "adr-runner-current" {
		t.Fatalf("decision projections result = %+v", decisionProjections)
	}
}

func TestRetrievalTaskTypedRecordValidation(t *testing.T) {
	t.Parallel()

	missing, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionServiceRecord,
	})
	if err != nil {
		t.Fatalf("missing service id validation: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "service_id is required" {
		t.Fatalf("missing result = %+v", missing)
	}

	negative, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionServicesLookup,
		Services: runner.ServiceLookupOptions{Limit: -1},
	})
	if err != nil {
		t.Fatalf("negative service limit validation: %v", err)
	}
	if !negative.Rejected || negative.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative result = %+v", negative)
	}

	missingDecision, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionRecord,
	})
	if err != nil {
		t.Fatalf("missing decision id validation: %v", err)
	}
	if !missingDecision.Rejected || missingDecision.RejectionReason != "decision_id is required" {
		t.Fatalf("missing decision result = %+v", missingDecision)
	}

	negativeDecision, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action:    runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{Limit: -1},
	})
	if err != nil {
		t.Fatalf("negative decision limit validation: %v", err)
	}
	if !negativeDecision.Rejected || negativeDecision.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative decision result = %+v", negativeDecision)
	}
}

func TestValidationRejectionDoesNotCreateRuntimeFiles(t *testing.T) {
	t.Parallel()

	dataDir := filepath.Join(t.TempDir(), "data")
	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(dataDir, "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Title: "Missing path",
			Body:  "# Missing path\n",
		},
	})
	if err != nil {
		t.Fatalf("document task: %v", err)
	}
	if !result.Rejected || result.RejectionReason == "" {
		t.Fatalf("result = %+v, want rejected", result)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Fatalf("data dir exists after validation rejection: %v", err)
	}
}

func TestResolvePathsUsesDatabaseAnchoredConfig(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "env-db", "openclerk.sqlite")
	t.Setenv("OPENCLERK_DATABASE_PATH", dbPath)

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionResolvePaths,
	})
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	if result.Paths == nil ||
		result.Paths.DatabasePath != dbPath ||
		result.Paths.VaultRoot != filepath.Join(filepath.Dir(dbPath), "vault") {
		t.Fatalf("paths = %+v", result.Paths)
	}

	boundVaultRoot := filepath.Join(t.TempDir(), "wiki")
	initialized, err := runclient.InitializePaths(runclient.Config{DatabasePath: dbPath}, boundVaultRoot)
	if err != nil {
		t.Fatalf("initialize paths: %v", err)
	}
	if initialized.VaultRoot != boundVaultRoot {
		t.Fatalf("initialized paths = %+v, want vault %q", initialized, boundVaultRoot)
	}

	explicit, err := runner.RunDocumentTask(context.Background(), runclient.Config{
		DatabasePath: filepath.Join(t.TempDir(), "explicit-db", "openclerk.sqlite"),
	}, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionResolvePaths})
	if err != nil {
		t.Fatalf("resolve explicit paths: %v", err)
	}
	if explicit.Paths.DatabasePath == os.Getenv("OPENCLERK_DATABASE_PATH") ||
		explicit.Paths.VaultRoot == boundVaultRoot {
		t.Fatalf("explicit config did not take precedence: %+v", explicit.Paths)
	}

	again, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: dbPath}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionResolvePaths,
	})
	if err != nil {
		t.Fatalf("resolve persisted paths: %v", err)
	}
	if again.Paths == nil || again.Paths.VaultRoot != boundVaultRoot {
		t.Fatalf("persisted paths = %+v, want vault %q", again.Paths, boundVaultRoot)
	}
}

func TestResolvePathsZeroConfigCreatesDefaultDatabaseAndVaultConfig(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "xdg"))
	t.Setenv("OPENCLERK_DATABASE_PATH", "")
	t.Setenv("OPENCLERK_DATA_DIR", filepath.Join(t.TempDir(), "retired-data"))
	t.Setenv("OPENCLERK_VAULT_ROOT", filepath.Join(t.TempDir(), "retired-vault"))

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionResolvePaths,
	})
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	wantDB := filepath.Join(os.Getenv("XDG_DATA_HOME"), "openclerk", "openclerk.sqlite")
	wantVault := filepath.Join(filepath.Dir(wantDB), "vault")
	if result.Paths == nil ||
		result.Paths.DatabasePath != wantDB ||
		result.Paths.VaultRoot != wantVault {
		t.Fatalf("paths = %+v, want db %q vault %q", result.Paths, wantDB, wantVault)
	}
	if _, err := os.Stat(wantDB); err != nil {
		t.Fatalf("default database was not created: %v", err)
	}
	if _, err := os.Stat(wantVault); err != nil {
		t.Fatalf("default vault was not created: %v", err)
	}
}

func createDocument(t *testing.T, ctx context.Context, config runclient.Config, path string, title string, body string) runner.Document {
	t.Helper()
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  path,
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	if result.Document == nil {
		t.Fatalf("create %s result = %+v", path, result)
	}
	return *result.Document
}

func runnerEventTypesInclude(events []runner.ProvenanceEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func auditInspectedPath(inspections []runner.AuditProvenanceInspection, path string) bool {
	for _, inspection := range inspections {
		if inspection.SourcePath == path && len(inspection.EventIDs) > 0 {
			return true
		}
	}
	return false
}

func minimalPDF(title string, author string, text string) []byte {
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

func layoutChecksInclude(checks []runner.KnowledgeLayoutCheck, id string, status string) bool {
	for _, check := range checks {
		if check.ID == id && check.Status == status {
			return true
		}
	}
	return false
}

func layoutPathConventionsInclude(conventions []runner.LayoutPathConvention, pathPrefix string) bool {
	for _, convention := range conventions {
		if convention.PathPrefix == pathPrefix {
			return true
		}
	}
	return false
}
