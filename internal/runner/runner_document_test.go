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

func TestDocumentTaskCompileSynthesisCreatesAndUpdatesOneTarget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/synthesis-a.md", "Synthesis source A", "# Synthesis source A\n\n## Summary\nCurrent synthesis workflow evidence A.\n")
	createDocument(t, ctx, config, "sources/synthesis-b.md", "Synthesis source B", "# Synthesis source B\n\n## Summary\nCurrent synthesis workflow evidence B.\n")

	body := strings.TrimSpace(`# Workflow Synthesis

## Summary
Initial workflow synthesis.

## Sources
- sources/synthesis-a.md
- sources/synthesis-b.md

## Freshness
Checked current source evidence.
`) + "\n"
	created, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:       "synthesis/workflow.md",
			Title:      "Workflow Synthesis",
			SourceRefs: []string{"sources/synthesis-a.md", "sources/synthesis-b.md"},
			Body:       body,
			Mode:       "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis create: %v", err)
	}
	if created.Rejected || created.CompileSynthesis == nil {
		t.Fatalf("compile synthesis create result = %+v", created)
	}
	if created.CompileSynthesis.WriteStatus != "created" ||
		created.CompileSynthesis.DuplicateStatus != "no_duplicate_created" ||
		len(created.CompileSynthesis.SourceEvidence) != 2 ||
		len(created.CompileSynthesis.ProjectionFreshness) == 0 ||
		created.CompileSynthesis.AgentHandoff == nil ||
		!strings.Contains(created.CompileSynthesis.AgentHandoff.AnswerSummary, "compile_synthesis created synthesis/workflow.md") ||
		!strings.Contains(created.CompileSynthesis.AgentHandoff.FollowUpPrimitiveInspection, "not required") ||
		!strings.Contains(created.CompileSynthesis.ValidationBoundaries, "no broad repo search") {
		t.Fatalf("compile synthesis create report = %+v", created.CompileSynthesis)
	}

	updateBody := strings.Replace(body, "Initial workflow synthesis.", "Updated workflow synthesis.", 1)
	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:       "synthesis/workflow.md",
			Title:      "Workflow Synthesis",
			SourceRefs: []string{"sources/synthesis-a.md", "sources/synthesis-b.md"},
			Body:       updateBody,
			Mode:       "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis update: %v", err)
	}
	if updated.CompileSynthesis == nil ||
		!updated.CompileSynthesis.ExistingCandidate ||
		updated.CompileSynthesis.WriteStatus != "updated" ||
		updated.CompileSynthesis.DocumentID != created.CompileSynthesis.DocumentID {
		t.Fatalf("compile synthesis update report = %+v", updated.CompileSynthesis)
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "synthesis/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list synthesis docs: %v", err)
	}
	if len(list.Documents) != 1 {
		t.Fatalf("synthesis docs = %+v, want one target", list.Documents)
	}
	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  created.CompileSynthesis.DocumentID,
	})
	if err != nil {
		t.Fatalf("get synthesis: %v", err)
	}
	if get.Document == nil ||
		!strings.Contains(get.Document.Body, "source_refs: sources/synthesis-a.md, sources/synthesis-b.md") ||
		!strings.Contains(get.Document.Body, "Updated workflow synthesis.") {
		t.Fatalf("compiled synthesis body = %q", get.Document.Body)
	}
}

func TestDocumentTaskCompileSynthesisBuildsBodyFromFacts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/fact-current.md", "Fact source current", "# Fact source current\n\n## Summary\nCurrent fact source.\n")
	createDocument(t, ctx, config, "sources/fact-old.md", "Fact source old", "# Fact source old\n\n## Summary\nSuperseded fact source.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:          "synthesis/fact-built.md",
			Title:         "Fact Built",
			SourceRefs:    []string{"sources/fact-current.md", "sources/fact-old.md"},
			BodyFacts:     []string{"Current source: sources/fact-current.md", "Superseded source: sources/fact-old.md"},
			FreshnessNote: "Checked through runner-owned body assembly.",
			Mode:          "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis body facts: %v", err)
	}
	if result.Rejected || result.CompileSynthesis == nil || result.CompileSynthesis.AgentHandoff == nil {
		t.Fatalf("compile synthesis body facts result = %+v", result)
	}
	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  result.CompileSynthesis.DocumentID,
	})
	if err != nil {
		t.Fatalf("get fact-built synthesis: %v", err)
	}
	for _, want := range []string{
		"source_refs: sources/fact-current.md, sources/fact-old.md",
		"## Summary",
		"- Current source: sources/fact-current.md",
		"## Sources",
		"- sources/fact-current.md",
		"## Freshness",
		"Checked through runner-owned body assembly.",
	} {
		if get.Document == nil || !strings.Contains(get.Document.Body, want) {
			t.Fatalf("fact-built body missing %q:\n%s", want, get.Document.Body)
		}
	}
}

func TestDocumentTaskCompileSynthesisRejectsMissingFields(t *testing.T) {
	t.Parallel()

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:  "synthesis/missing.md",
			Title: "Missing",
			Body:  "# Missing\n\n## Summary\nMissing source refs.\n",
			Mode:  "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "synthesis.source_refs is required" {
		t.Fatalf("compile synthesis rejection = %+v", result)
	}
}

func TestDocumentTaskIngestSourceURLPlanModeIsReadOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        "https://Example.Test/product/page.html#section",
			Mode:       "plan",
			SourceType: "web",
			Title:      "Runner Product Page",
		},
	})
	if err != nil {
		t.Fatalf("source placement plan: %v", err)
	}
	if result.Rejected || result.SourcePlacement == nil {
		t.Fatalf("placement plan result = %+v", result)
	}
	plan := result.SourcePlacement
	if plan.SourceURL != "https://example.test/product/page.html" ||
		plan.SourceType != "web" ||
		plan.DuplicateStatus != "no_existing_source_url_found" ||
		plan.FetchStatus != "planned_no_fetch" ||
		plan.WriteStatus != "planned_no_write" ||
		plan.AgentHandoff == nil ||
		!strings.Contains(plan.ApprovalBoundary, "durable-write approval") ||
		!containsString(plan.CandidateSourcePaths, "sources/web/runner-product-page.md") ||
		plan.CandidateSynthesisPath != "synthesis/runner-product-page.md" ||
		!strings.Contains(plan.ValidationBoundaries, "no fetch") {
		t.Fatalf("placement plan = %+v", plan)
	}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "sources/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list sources after plan: %v", err)
	}
	if len(list.Documents) != 0 {
		t.Fatalf("plan mode wrote source documents: %+v", list.Documents)
	}
}

func TestDocumentTaskIngestSourceURLPlanModeReportsExistingSource(t *testing.T) {
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "existing.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Existing Source</title></head><body><h1>Existing Source</h1><p>Existing source evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	sourceURL := "http://openclerk-eval.local/web/existing.html"
	created, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			PathHint:   "sources/web/existing.md",
			SourceType: "web",
		},
	})
	if err != nil {
		t.Fatalf("create web source: %v", err)
	}
	plan, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			Mode:       "plan",
			SourceType: "web",
		},
	})
	if err != nil {
		t.Fatalf("source placement duplicate plan: %v", err)
	}
	if plan.SourcePlacement == nil ||
		plan.SourcePlacement.ExistingSource == nil ||
		plan.SourcePlacement.ExistingSource.DocID != created.Ingestion.DocID ||
		plan.SourcePlacement.DuplicateStatus != "existing_source_url_found_no_fetch_no_write" ||
		plan.SourcePlacement.CandidateSynthesisPath != "" ||
		!strings.Contains(plan.SourcePlacement.AgentHandoff.AnswerSummary, "no fetch or write occurred") {
		t.Fatalf("duplicate placement plan = %+v", plan.SourcePlacement)
	}
}

func TestDocumentTaskCompileSynthesisAssemblesPlainBodyAndAliases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/plain-current.md", "Plain source", "# Plain source\n\n## Summary\nPlain body source.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:        runner.DocumentTaskActionCompileSynthesis,
		Document:      runner.DocumentInput{Path: "synthesis/plain-body", Title: "Plain Body"},
		Body:          "Plain synthesis summary from the user.",
		SourceRefs:    []string{"sources/plain-current.md"},
		FreshnessNote: "Plain body wrapped by compile_synthesis.",
	})
	if err != nil {
		t.Fatalf("compile synthesis plain body aliases: %v", err)
	}
	if result.Rejected || result.CompileSynthesis == nil {
		t.Fatalf("compile synthesis plain body result = %+v", result)
	}
	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  result.CompileSynthesis.DocumentID,
	})
	if err != nil {
		t.Fatalf("get plain-body synthesis: %v", err)
	}
	for _, want := range []string{
		"source_refs: sources/plain-current.md",
		"# Plain Body",
		"## Summary",
		"Plain synthesis summary from the user.",
		"## Sources",
		"- sources/plain-current.md",
		"## Freshness",
		"Plain body wrapped by compile_synthesis.",
	} {
		if get.Document == nil || !strings.Contains(get.Document.Body, want) {
			t.Fatalf("plain-body result missing %q:\n%s", want, get.Document.Body)
		}
	}
}

func TestDocumentTaskCompileSynthesisRejectsMissingBodyAndFactsWithoutWrite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:       "synthesis/missing-body.md",
			Title:      "Missing Body",
			SourceRefs: []string{"sources/a.md"},
			Mode:       "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis missing body: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "synthesis.body or synthesis.body_facts is required" {
		t.Fatalf("compile synthesis missing body rejection = %+v", result)
	}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after missing body reject: %v", err)
	}
	if len(list.Documents) != 0 {
		t.Fatalf("missing body rejection wrote documents: %+v", list.Documents)
	}
}

func TestDocumentTaskCompileSynthesisRejectsCleanedPathsOutsideNamespaces(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	body := strings.TrimSpace(`# Invalid

## Summary
Invalid synthesis.

## Sources
- sources/source.md

## Freshness
Checked.
`) + "\n"

	for _, tt := range []struct {
		name      string
		path      string
		sourceRef string
		want      string
	}{
		{
			name:      "target traversal",
			path:      "synthesis/../notes/escaped.md",
			sourceRef: "sources/source.md",
			want:      "synthesis.path must be under synthesis/",
		},
		{
			name:      "source ref traversal",
			path:      "synthesis/valid.md",
			sourceRef: "sources/../notes/source.md",
			want:      "synthesis.source_refs entries must be under sources/",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionCompileSynthesis,
				Synthesis: runner.CompileSynthesisInput{
					Path:       tt.path,
					Title:      "Invalid",
					SourceRefs: []string{tt.sourceRef},
					Body:       body,
					Mode:       "create_or_update",
				},
			})
			if err != nil {
				t.Fatalf("compile synthesis reject: %v", err)
			}
			if !result.Rejected || result.RejectionReason != tt.want {
				t.Fatalf("compile synthesis rejection = %+v, want %q", result, tt.want)
			}
		})
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after rejected compile synthesis: %v", err)
	}
	if len(list.Documents) != 0 {
		t.Fatalf("rejected compile synthesis wrote documents: %+v", list.Documents)
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

func TestDocumentTaskIngestSourceURLUpdateStaleImpactResponse(t *testing.T) {
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "runner-product.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Runner Web Title</title></head><body><h1>Runner Web Title</h1><p>Initial runner evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	sourceURL := "http://openclerk-eval.local/web/runner-product.html"
	created, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:      sourceURL,
			PathHint: "sources/web/runner-product.md",
		},
	})
	if err != nil {
		t.Fatalf("create web source: %v", err)
	}
	if created.Ingestion == nil || created.Ingestion.UpdateStatus != "" {
		t.Fatalf("create ingestion = %+v", created.Ingestion)
	}
	createJSON, err := json.Marshal(created.Ingestion)
	if err != nil {
		t.Fatalf("marshal create ingestion: %v", err)
	}
	if strings.Contains(string(createJSON), "update_status") {
		t.Fatalf("create ingestion leaked update fields: %s", createJSON)
	}

	if _, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "synthesis/web-runner.md",
			Title: "Web Runner Synthesis",
			Body:  "---\ntype: synthesis\nsource_refs: sources/web/runner-product.md\n---\n# Web Runner Synthesis\n\n## Summary\nInitial runner evidence.\n",
		},
	}); err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	same, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:  sourceURL,
			Mode: "update",
		},
	})
	if err != nil {
		t.Fatalf("same web update: %v", err)
	}
	if same.Ingestion == nil ||
		same.Ingestion.UpdateStatus != "no_op" ||
		same.Ingestion.NormalizedSourceURL != sourceURL ||
		same.Ingestion.SourceDocID != created.Ingestion.DocID ||
		same.Ingestion.PreviousSHA256 != created.Ingestion.SHA256 ||
		same.Ingestion.NewSHA256 != created.Ingestion.SHA256 ||
		same.Ingestion.Changed == nil || *same.Ingestion.Changed ||
		same.Ingestion.SynthesisRepaired == nil || *same.Ingestion.SynthesisRepaired ||
		same.Ingestion.StaleDependents == nil || len(*same.Ingestion.StaleDependents) != 0 ||
		same.Ingestion.ProjectionRefs == nil || len(*same.Ingestion.ProjectionRefs) != 0 ||
		same.Ingestion.ProvenanceRefs == nil || len(*same.Ingestion.ProvenanceRefs) != 0 {
		t.Fatalf("same update ingestion = %+v", same.Ingestion)
	}

	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Runner Web Title Updated</title></head><body><h1>Runner Web Title Updated</h1><p>Updated runner evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write updated web fixture: %v", err)
	}
	changed, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			PathHint:   "sources/web/runner-product.md",
			SourceType: "web",
			Mode:       "update",
		},
	})
	if err != nil {
		t.Fatalf("changed web update: %v", err)
	}
	if changed.Ingestion == nil ||
		changed.Ingestion.UpdateStatus != "changed" ||
		changed.Ingestion.NormalizedSourceURL != sourceURL ||
		changed.Ingestion.SourceDocID != created.Ingestion.DocID ||
		changed.Ingestion.PreviousSHA256 != created.Ingestion.SHA256 ||
		changed.Ingestion.NewSHA256 == created.Ingestion.SHA256 ||
		changed.Ingestion.Changed == nil || !*changed.Ingestion.Changed ||
		changed.Ingestion.SynthesisRepaired == nil || *changed.Ingestion.SynthesisRepaired ||
		changed.Ingestion.StaleDependents == nil || len(*changed.Ingestion.StaleDependents) != 1 ||
		changed.Ingestion.ProjectionRefs == nil || len(*changed.Ingestion.ProjectionRefs) == 0 ||
		changed.Ingestion.ProvenanceRefs == nil || !runnerSourceProvenanceRefsInclude(*changed.Ingestion.ProvenanceRefs, "source_updated") ||
		!strings.Contains(changed.Ingestion.NoRepairWarning, "synthesis/web-runner.md") {
		t.Fatalf("changed update ingestion = %+v", changed.Ingestion)
	}
	updateJSON, err := json.Marshal(changed.Ingestion)
	if err != nil {
		t.Fatalf("marshal changed ingestion: %v", err)
	}
	for _, want := range []string{`"update_status":"changed"`, `"normalized_source_url":"` + sourceURL + `"`, `"source_doc_id":"`, `"previous_sha256":"`, `"new_sha256":"`, `"changed":true`, `"stale_dependents":[`, `"projection_refs":[`, `"provenance_refs":[`, `"synthesis_repaired":false`, `"no_repair_warning":"`} {
		if !strings.Contains(string(updateJSON), want) {
			t.Fatalf("changed ingestion JSON missing %s: %s", want, updateJSON)
		}
	}
}

func runnerSourceProvenanceRefsInclude(refs []runner.SourceProvenanceRef, eventType string) bool {
	for _, ref := range refs {
		if ref.EventType == eventType {
			return true
		}
	}
	return false
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

func TestDocumentTaskListTagFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/tagging/support-handoff.md", "Support Handoff", strings.TrimSpace(`---
tag: support-handoff
---
# Support Handoff

## Summary
Support handoff tag list evidence belongs under active notes.
`)+"\n")
	createDocument(t, ctx, config, "archive/tagging/support-handoff.md", "Archived Support Handoff", strings.TrimSpace(`---
tag: support-handoff
---
# Archived Support Handoff

## Summary
Archived support handoff tag list evidence must be excluded by path prefix.
`)+"\n")
	createDocument(t, ctx, config, "notes/tagging/support-handoffs.md", "Support Handoffs", strings.TrimSpace(`---
tag: support-handoffs
---
# Support Handoffs

## Summary
Plural support handoffs tag list evidence must not match singular support-handoff.
`)+"\n")

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			PathPrefix: "notes/tagging/",
			Tag:        " support-handoff ",
			Limit:      20,
		},
	})
	if err != nil {
		t.Fatalf("list tag: %v", err)
	}
	if len(list.Documents) != 1 || list.Documents[0].Path != "notes/tagging/support-handoff.md" {
		t.Fatalf("list tag result = %+v", list.Documents)
	}

	backCompat, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			MetadataKey:   "tag",
			MetadataValue: "support-handoff",
			Limit:         20,
		},
	})
	if err != nil {
		t.Fatalf("list metadata tag: %v", err)
	}
	if len(backCompat.Documents) != 2 {
		t.Fatalf("backward-compatible metadata tag result = %+v", backCompat.Documents)
	}

	mixed, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			Tag:           "support-handoff",
			MetadataKey:   "tag",
			MetadataValue: "support-handoff",
		},
	})
	if err != nil {
		t.Fatalf("mixed tag list validation: %v", err)
	}
	if !mixed.Rejected || mixed.RejectionReason != "list.tag cannot be combined with metadata_key or metadata_value" {
		t.Fatalf("mixed list result = %+v", mixed)
	}

	empty, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			Tag: " ",
		},
	})
	if err != nil {
		t.Fatalf("empty tag list validation: %v", err)
	}
	if !empty.Rejected || empty.RejectionReason != "list.tag must be non-empty" {
		t.Fatalf("empty list result = %+v", empty)
	}
}
