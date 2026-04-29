package runner_test

import (
	"context"
	"encoding/json"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		{
			name: "invalid source type",
			source: runner.SourceURLInput{
				URL:        "https://example.test/uploaded.html",
				PathHint:   "sources/uploaded-web.md",
				SourceType: "html",
			},
			wantErr: "source.source_type",
		},
		{
			name: "web source asset path",
			source: runner.SourceURLInput{
				URL:           "https://example.test/uploaded.html",
				PathHint:      "sources/uploaded-web.md",
				AssetPathHint: "assets/sources/uploaded-web.pdf",
				SourceType:    "web",
			},
			wantErr: "source.asset_path_hint",
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

func TestDocumentTaskIngestSourceURLWeb(t *testing.T) {
	htmlBody := `<!doctype html>
<html>
<head><title>Runner Web Product</title><style>.hidden{display:none}</style></head>
<body>
<h1>Runner Web Product</h1>
<p>Visible web source evidence for OpenClerk.</p>
<script>doNotIndex()</script>
</body>
</html>`
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "runner-product.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(htmlBody), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)
	sourceURL := "http://openclerk-eval.local/web/runner-product.html?ref=tracker"

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			PathHint:   "sources/web/runner-product.md",
			SourceType: "web",
		},
	})
	if err != nil {
		t.Fatalf("ingest web source URL: %v", err)
	}
	if result.Rejected || result.Ingestion == nil {
		t.Fatalf("ingest result = %+v", result)
	}
	ingestion := result.Ingestion
	if ingestion.DocID == "" ||
		ingestion.SourcePath != "sources/web/runner-product.md" ||
		ingestion.SourceURL != sourceURL ||
		ingestion.SourceType != "web" ||
		ingestion.AssetPath != "" ||
		ingestion.DerivedPath != "sources/web/runner-product.md" ||
		ingestion.PageCount != 0 ||
		ingestion.MIMEType != "text/html" ||
		len(ingestion.Citations) == 0 ||
		len(ingestion.SHA256) != 64 {
		t.Fatalf("ingestion = %+v", ingestion)
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  ingestion.DocID,
	})
	if err != nil {
		t.Fatalf("get ingested web document: %v", err)
	}
	if get.Document == nil ||
		get.Document.Metadata["source_url"] != sourceURL ||
		get.Document.Metadata["source_type"] != "web" ||
		get.Document.Metadata["asset_path"] != "" ||
		get.Document.Metadata["mime_type"] != "text/html" ||
		!strings.Contains(get.Document.Body, "Runner Web Product") ||
		!strings.Contains(get.Document.Body, "Visible web source evidence for OpenClerk.") ||
		strings.Contains(get.Document.Body, "doNotIndex") {
		t.Fatalf("ingested web document = %+v", get.Document)
	}

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "Visible web source evidence",
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("search ingested web text: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || search.Search.Hits[0].Citations[0].Path != "sources/web/runner-product.md" {
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
	if details["source_url"] != sourceURL || details["source_type"] != "web" || details["asset_path"] != "" {
		t.Fatalf("source provenance details = %+v", details)
	}

	duplicate, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:      sourceURL,
			PathHint: "sources/web/runner-product-copy.md",
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
			SourceType:    "pdf",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "PDF") {
		t.Fatalf("err = %v, want non-PDF rejection", err)
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
