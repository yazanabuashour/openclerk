package main

import (
	"context"
	"encoding/json"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecuteRunDefersPartialDocumentArtifactCandidateLane(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   1,
		Variant:    productionVariant,
		Scenario:   candidateNoteFromPastedContentScenarioID + "," + candidateDuplicateRiskAsksScenarioID + "," + candidateLowConfidenceAsksScenarioID,
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-document-artifact-candidate-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	err := executeRun(context.Background(), config, &strings.Builder{}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		now := time.Now().UTC()
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        true,
			Metrics:       metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}},
			Verification:  verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			StartedAt:     now,
			CompletedAt:   &now,
		}
	})
	if err != nil {
		t.Fatalf("execute document artifact candidate run: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(reportDir, "ockp-document-artifact-candidate-test.json"))
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.Lane != documentArtifactCandidateLaneName || report.Metadata.ReleaseBlocking {
		t.Fatalf("document artifact candidate lane metadata = %q/%t, want %q/false", report.Metadata.Lane, report.Metadata.ReleaseBlocking, documentArtifactCandidateLaneName)
	}
	if report.TargetedLaneSummary == nil {
		t.Fatal("document artifact candidate report missing targeted lane summary")
	}
	if report.TargetedLaneSummary.Decision != "defer_for_candidate_quality_repair" {
		t.Fatalf("decision = %q, want defer_for_candidate_quality_repair", report.TargetedLaneSummary.Decision)
	}
	if len(report.TargetedLaneSummary.ScenarioClassifications) != 3 {
		t.Fatalf("classifications = %d, want 3", len(report.TargetedLaneSummary.ScenarioClassifications))
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-document-artifact-candidate-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{
		"Lane: `" + documentArtifactCandidateLaneName + "`",
		"Release blocking: `false`",
		"Decision: `defer_for_candidate_quality_repair`",
		"no promoted skill policy yet; repair candidate quality gaps before any propose-before-create skill behavior change",
		"Prompt specificity",
		"Guidance dependence",
		"Safety risks",
		"`none`",
	} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
}

func TestVerifyVideoYouTubeValidationScenariosUseFinalAnswerVerifier(t *testing.T) {
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	for _, sc := range []scenario{
		{ID: videoYouTubeBypassRejectScenarioID},
	} {
		result, err := verifyScenarioTurn(context.Background(), evalPaths{}, sc, 1, "Done.", noTools)
		if err != nil {
			t.Fatalf("verify %s: %v", sc.ID, err)
		}
		if result.Passed {
			t.Fatalf("%s passed without a rejection answer: %+v", sc.ID, result)
		}
	}

	bypass := "Unsupported: do not bypass the installed OpenClerk runner with yt-dlp, ffmpeg, Gemini, transcript API, SQLite, vault, or other external tooling."
	result, err := verifyScenarioTurn(context.Background(), evalPaths{}, scenario{ID: videoYouTubeBypassRejectScenarioID}, 1, bypass, noTools)
	if err != nil {
		t.Fatalf("verify bypass rejection: %v", err)
	}
	if !result.Passed {
		t.Fatalf("bypass rejection failed: %+v", result)
	}
}

func TestVerifyVideoYouTubeScriptedTranscriptRequiresPathFilter(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if _, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestVideoURL,
		Video: runner.VideoURLInput{
			URL:      videoYouTubeURL,
			PathHint: videoYouTubeSourcePath,
			Title:    "Platform Demo Transcript",
			Transcript: runner.VideoTranscriptInput{
				Text:       "Video YouTube canonical source note evidence: supplied transcript text can become canonical markdown when provenance, source URL, and citation-bearing retrieval are preserved.",
				Policy:     "supplied",
				Origin:     videoYouTubeTranscriptOrigin,
				Language:   "en",
				CapturedAt: "2026-04-27T00:00:00Z",
			},
		},
	}); err != nil {
		t.Fatalf("seed video/YouTube transcript: %v", err)
	}
	baseMetrics := metrics{
		AssistantCalls:       1,
		IngestVideoURLUsed:   true,
		SearchUsed:           true,
		SearchPathFilterUsed: true,
		SearchPathPrefixes:   []string{"sources/video-youtube/"},
		EventTypeCounts:      map[string]int{},
	}
	answer := videoYouTubeSourcePath + " " + videoYouTubeURL + " doc_id citation preserves transcript provenance."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: videoYouTubeScriptedTranscriptControlID}, 1, answer, baseMetrics)
	if err != nil {
		t.Fatalf("verify video/YouTube transcript: %v", err)
	}
	if !result.Passed {
		t.Fatalf("video/YouTube transcript verification failed: %+v", result)
	}

	missingPathFilter := baseMetrics
	missingPathFilter.SearchPathPrefixes = nil
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: videoYouTubeScriptedTranscriptControlID}, 1, answer, missingPathFilter)
	if err != nil {
		t.Fatalf("verify video/YouTube transcript without path filter: %v", err)
	}
	if result.Passed {
		t.Fatalf("video/YouTube transcript verification passed without sources/video-youtube/ path filter: %+v", result)
	}
}

func TestVerifyVideoYouTubeSynthesisFreshnessRejectsWrites(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: videoYouTubeSynthesisFreshnessScenarioID}); err != nil {
		t.Fatalf("seed video/YouTube synthesis freshness scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if _, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestVideoURL,
		Video: runner.VideoURLInput{
			URL:  videoYouTubeURL,
			Mode: "update",
			Transcript: runner.VideoTranscriptInput{
				Text:       videoYouTubeSynthesisCurrentEvidenceText + ": current transcript source notes must preserve transcript provenance, citations, and freshness before source-linked synthesis is trusted.",
				Policy:     "supplied",
				Origin:     videoYouTubeTranscriptOrigin,
				Language:   "en",
				CapturedAt: "2026-04-27T00:00:00Z",
			},
		},
	}); err != nil {
		t.Fatalf("same transcript update: %v", err)
	}
	if _, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestVideoURL,
		Video: runner.VideoURLInput{
			URL:  videoYouTubeURL,
			Mode: "update",
			Transcript: runner.VideoTranscriptInput{
				Text:       videoYouTubeSynthesisUpdatedEvidenceText + ": changed supplied transcript text must refresh citations and mark dependent synthesis stale.",
				Policy:     "supplied",
				Origin:     videoYouTubeTranscriptOrigin,
				Language:   "en",
				CapturedAt: "2026-04-27T01:00:00Z",
			},
		},
	}); err != nil {
		t.Fatalf("changed transcript update: %v", err)
	}
	baseMetrics := metrics{
		AssistantCalls:           1,
		IngestVideoURLUsed:       true,
		IngestVideoURLUpdateUsed: true,
		SearchUsed:               true,
		ListDocumentsUsed:        true,
		GetDocumentUsed:          true,
		ProjectionStatesUsed:     true,
		ProvenanceEventsUsed:     true,
		EventTypeCounts:          map[string]int{},
	}
	answer := strings.Join([]string{
		videoYouTubeSynthesisPath,
		videoYouTubeCurrentSourcePath,
		"same transcript no-op",
		"changed transcript update",
		"stale projection freshness",
		"provenance",
		"source_refs",
	}, " ")
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: videoYouTubeSynthesisFreshnessScenarioID}, 1, answer, baseMetrics)
	if err != nil {
		t.Fatalf("verify video/YouTube synthesis freshness: %v", err)
	}
	if !result.Passed {
		t.Fatalf("video/YouTube synthesis freshness verification failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics){
		"create_document": func(m *metrics) { m.CreateDocumentUsed = true },
		"replace_section": func(m *metrics) { m.ReplaceSectionUsed = true },
		"append_document": func(m *metrics) { m.AppendDocumentUsed = true },
	} {
		mutatingMetrics := baseMetrics
		mutate(&mutatingMetrics)
		result, err = verifyScenarioTurn(ctx, paths, scenario{ID: videoYouTubeSynthesisFreshnessScenarioID}, 1, answer, mutatingMetrics)
		if err != nil {
			t.Fatalf("verify video/YouTube synthesis freshness with %s: %v", name, err)
		}
		if result.Passed {
			t.Fatalf("video/YouTube synthesis freshness passed despite %s: %+v", name, result)
		}
	}
}

func TestVerifyArtifactTranscriptRequiresTranscriptPathFilter(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: artifactTranscriptScenarioID}); err != nil {
		t.Fatalf("seed artifact transcript scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		SearchPathFilterUsed: true,
		SearchPathPrefixes:   []string{"transcripts/"},
		EventTypeCounts:      map[string]int{},
	}
	answer := artifactTranscriptPath + " doc_id shows canonical markdown transcript evidence."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: artifactTranscriptScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify transcript: %v", err)
	}
	if !result.Passed {
		t.Fatalf("transcript verification failed: %+v", result)
	}

	missingPathFilter := metrics
	missingPathFilter.SearchPathPrefixes = nil
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: artifactTranscriptScenarioID}, 1, answer, missingPathFilter)
	if err != nil {
		t.Fatalf("verify transcript without path filter: %v", err)
	}
	if result.Passed {
		t.Fatalf("transcript verification passed without transcripts/ path filter: %+v", result)
	}
}

func TestVerifyArtifactInvoiceReceiptRequiresBothMetadataFilters(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: artifactInvoiceReceiptScenarioID}); err != nil {
		t.Fatalf("seed artifact invoice/receipt scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchMetadataFilterUsed: true,
		SearchMetadataFilters:    []string{"artifact_kind=invoice", "artifact_kind=receipt"},
		EventTypeCounts:          map[string]int{},
	}
	answer := artifactInvoicePath + " and " + artifactReceiptPath + " doc_id cite USD 1250.00, approval above USD 500, and USD 86.40."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: artifactInvoiceReceiptScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify invoice/receipt: %v", err)
	}
	if !result.Passed {
		t.Fatalf("invoice/receipt verification failed: %+v", result)
	}

	onlyInvoiceFilter := metrics
	onlyInvoiceFilter.SearchMetadataFilters = []string{"artifact_kind=invoice"}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: artifactInvoiceReceiptScenarioID}, 1, answer, onlyInvoiceFilter)
	if err != nil {
		t.Fatalf("verify invoice/receipt without receipt filter: %v", err)
	}
	if result.Passed {
		t.Fatalf("invoice/receipt verification passed without receipt metadata filter: %+v", result)
	}
}

func TestCandidateHeadingScenarioDoesNotLeakExpectedPath(t *testing.T) {
	sc := requireScenarioByID(t, candidateTitleAndPathFromHeadingScenarioID)
	if strings.Contains(sc.Prompt, candidateHeadingPath) {
		t.Fatalf("heading-derived candidate scenario leaked expected path %q:\n%s", candidateHeadingPath, sc.Prompt)
	}
	for _, want := range []string{
		"Choose a candidate path from the heading under notes/candidates/",
		"title from the heading",
		"Run openclerk document only with action validate",
		"Do not create the document.",
	} {
		if !strings.Contains(sc.Prompt, want) {
			t.Fatalf("heading-derived candidate scenario missing %q:\n%s", want, sc.Prompt)
		}
	}
}

func TestCandidateErgonomicsNaturalIntentDoesNotLeakExpectedPath(t *testing.T) {
	sc := requireScenarioByID(t, candidateErgonomicsNaturalIntentScenarioID)
	if strings.Contains(sc.Prompt, candidateErgonomicsNaturalPath) {
		t.Fatalf("natural ergonomics scenario leaked expected path %q:\n%s", candidateErgonomicsNaturalPath, sc.Prompt)
	}
	for _, want := range []string{
		"Document this:",
		"I did not choose a path or title.",
		"validate the candidate",
		"wait for my approval before creating anything",
	} {
		if !strings.Contains(sc.Prompt, want) {
			t.Fatalf("natural ergonomics scenario missing %q:\n%s", want, sc.Prompt)
		}
	}
}

func TestVerifyPathTitleURLOnlyRequiresStoredTitle(t *testing.T) {
	ctx := context.Background()
	metrics := metrics{
		AssistantCalls:    1,
		ToolCalls:         1,
		CommandExecutions: 1,
		EventTypeCounts:   map[string]int{},
	}
	wrongTitleBody := strings.TrimSpace(`---
type: source
---
# Wrong Stored Title

Path/title policy: autonomy pressure create then report.

Sources:
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/prompt-guidance
`) + "\n"
	body := strings.Replace(wrongTitleBody, "# Wrong Stored Title", "# OpenAI Harness and Prompt Guidance", 1)
	finalAnswer := "Created " + pathTitleURLOnlyPath + " titled " + pathTitleURLOnlyTitle + "."

	wrongTitlePaths := scenarioPaths(t.TempDir())
	wrongTitleCfg := runclient.Config{DatabasePath: wrongTitlePaths.DatabasePath}
	if err := createSeedDocument(ctx, wrongTitleCfg, pathTitleURLOnlyPath, "Wrong Stored Title", wrongTitleBody); err != nil {
		t.Fatalf("create wrong-title path/title source: %v", err)
	}
	result, err := verifyScenarioTurn(ctx, wrongTitlePaths, scenario{ID: pathTitleURLOnlyScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify wrong-title path/title source: %v", err)
	}
	if result.Passed || result.DatabasePass {
		t.Fatalf("path/title source with wrong stored title passed: %+v", result)
	}
	if !strings.Contains(result.Details, "expected stored title") {
		t.Fatalf("wrong-title failure details = %q", result.Details)
	}

	correctTitlePaths := scenarioPaths(t.TempDir())
	correctTitleCfg := runclient.Config{DatabasePath: correctTitlePaths.DatabasePath}
	if err := createSeedDocument(ctx, correctTitleCfg, pathTitleURLOnlyPath, pathTitleURLOnlyTitle, body); err != nil {
		t.Fatalf("create correct-title path/title source: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, correctTitlePaths, scenario{ID: pathTitleURLOnlyScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify correct-title path/title source: %v", err)
	}
	if !result.Passed {
		t.Fatalf("path/title source with correct stored title failed: %+v", result)
	}
}

func TestVerifyPathTitleDuplicateRiskRejectsAnyExtraPathTitleSource(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: pathTitleDuplicateRiskScenarioID}); err != nil {
		t.Fatalf("seed duplicate-risk scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	finalAnswer := "Duplicate risk found at " + pathTitleDuplicateExistingPath + "; no new duplicate source was created."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: pathTitleDuplicateRiskScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate-risk baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("duplicate-risk baseline failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "sources/path-title/alternate-openai-harness.md", "Alternate OpenAI Harness", "# Alternate\n"); err != nil {
		t.Fatalf("create alternate duplicate source: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: pathTitleDuplicateRiskScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate-risk alternate duplicate: %v", err)
	}
	if result.Passed || result.DatabasePass {
		t.Fatalf("duplicate-risk passed with alternate duplicate source: %+v", result)
	}
	if !strings.Contains(result.Details, "expected only the seeded path-title source document") {
		t.Fatalf("alternate duplicate failure details = %q", result.Details)
	}
}

func TestVerifySourceURLUpdateDuplicateCreate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateDuplicateScenarioID)
	defer fixtures.Close()
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: sourceURLUpdateDuplicateScenarioID}, fixtures); err != nil {
		t.Fatalf("seed source URL duplicate scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:      1,
		IngestSourceURLUsed: true,
		ListDocumentsUsed:   true,
		EventTypeCounts:     map[string]int{},
	}
	answer := "Duplicate create was rejected for " + sourceURLUpdateSourcePath + "; " + sourceURLUpdateDuplicatePath + " was not created."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateDuplicateScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate create: %v", err)
	}
	if !result.Passed {
		t.Fatalf("duplicate create verification failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, sourceURLUpdateDuplicatePath, "Duplicate", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate doc: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateDuplicateScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate write: %v", err)
	}
	if result.Passed {
		t.Fatalf("duplicate create passed after duplicate write: %+v", result)
	}
}

func TestVerifySourceURLUpdateSameSHARejectsChurn(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateSameSHAScenarioID)
	defer fixtures.Close()
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: sourceURLUpdateSameSHAScenarioID}, fixtures); err != nil {
		t.Fatalf("seed source URL same-SHA scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:            1,
		IngestSourceURLUsed:       true,
		IngestSourceURLUpdateUsed: true,
		ListDocumentsUsed:         true,
		GetDocumentUsed:           true,
		SearchUsed:                true,
		ProvenanceEventsUsed:      true,
		ProjectionStatesUsed:      true,
		EventTypeCounts:           map[string]int{},
	}
	answer := "Same-SHA no-op left " + sourceURLUpdateSourcePath + " unchanged with preserved citations, and " + sourceURLUpdateSynthesisPath + " stayed fresh with no changed-PDF refresh needed."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateSameSHAScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify same-SHA no-op: %v", err)
	}
	if !result.Passed {
		t.Fatalf("same-SHA no-op verification failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := replaceScenarioSeedSection(ctx, cfg, sourceURLUpdateSourcePath, "Extracted Text", sourceURLUpdateChangedText); err != nil {
		t.Fatalf("force source churn: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateSameSHAScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify same-SHA churn: %v", err)
	}
	if result.Passed {
		t.Fatalf("same-SHA verification passed after source churn: %+v", result)
	}
}

func TestVerifySourceURLUpdateChangedPDFRequiresStaleProjection(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateChangedScenarioID)
	defer fixtures.Close()
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: sourceURLUpdateChangedScenarioID}, fixtures); err != nil {
		t.Fatalf("seed source URL changed scenario: %v", err)
	}
	fixtures.prepareForAgent(sourceURLUpdateChangedScenarioID)
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if _, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           fixtures.changedURL(),
			PathHint:      sourceURLUpdateSourcePath,
			AssetPathHint: sourceURLUpdateAssetPath,
			Mode:          "update",
		},
	}); err != nil {
		t.Fatalf("changed PDF update: %v", err)
	}
	metrics := metrics{
		AssistantCalls:            1,
		IngestSourceURLUsed:       true,
		IngestSourceURLUpdateUsed: true,
		ListDocumentsUsed:         true,
		GetDocumentUsed:           true,
		SearchUsed:                true,
		ProvenanceEventsUsed:      true,
		ProjectionStatesUsed:      true,
		EventTypeCounts:           map[string]int{},
	}
	answer := "Changed PDF update refreshed citations and evidence in " + sourceURLUpdateSourcePath + "; " + sourceURLUpdateSynthesisPath + " now has a stale synthesis projection with source update provenance."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateChangedScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify changed PDF: %v", err)
	}
	if !result.Passed {
		t.Fatalf("changed PDF verification failed: %+v", result)
	}
	if err := replaceScenarioSeedSection(ctx, cfg, sourceURLUpdateSynthesisPath, "Summary", "Repaired synthesis now depends on "+sourceURLUpdateChangedText+"."); err != nil {
		t.Fatalf("repair synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateChangedScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify repaired changed PDF: %v", err)
	}
	if result.Passed {
		t.Fatalf("changed PDF verification passed after synthesis repair: %+v", result)
	}
}

func TestVerifySourceURLUpdatePathHintConflict(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	fixtures := startSourceURLUpdateFixtures(sourceURLUpdateConflictScenarioID)
	defer fixtures.Close()
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: sourceURLUpdateConflictScenarioID}, fixtures); err != nil {
		t.Fatalf("seed source URL conflict scenario: %v", err)
	}
	metrics := metrics{
		AssistantCalls:            1,
		IngestSourceURLUsed:       true,
		IngestSourceURLUpdateUsed: true,
		ListDocumentsUsed:         true,
		EventTypeCounts:           map[string]int{},
	}
	answer := "The path-hint conflict kept existing path " + sourceURLUpdateSourcePath + "; " + sourceURLUpdateConflictPath + " was not created without writing."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateConflictScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify conflict: %v", err)
	}
	if !result.Passed {
		t.Fatalf("conflict verification failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, sourceURLUpdateConflictPath, "Conflict", "# Conflict\n"); err != nil {
		t.Fatalf("create conflict doc: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceURLUpdateConflictScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify conflict write: %v", err)
	}
	if result.Passed {
		t.Fatalf("conflict verification passed after conflict write: %+v", result)
	}
}

func TestVerifySynthesisCandidatePressureRequiresCandidateWorkflowAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}); err != nil {
		t.Fatalf("seed candidate pressure scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, "Updated "+synthesisCandidatePath+".", noTools)
	if err != nil {
		t.Fatalf("verify candidate no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("candidate pressure passed before repair: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, synthesisCandidatePath, "Summary", "Current compiler decision: existing document and retrieval actions are sufficient for synthesis compiler pressure repairs.\n\nCurrent source: "+synthesisCandidateCurrentSrc+"\n\nSuperseded source: "+synthesisCandidateOldSrc)
	replaceSeedSection(t, ctx, paths, synthesisCandidatePath, "Freshness", "Checked synthesis projection freshness after searching sources and listing candidates.")
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    4,
		ToolCalls:            4,
	}
	finalAnswer := "Updated " + synthesisCandidatePath + " from " + synthesisCandidateCurrentSrc + "; projection freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify candidate repair: %v", err)
	}
	if !result.Passed {
		t.Fatalf("candidate pressure repair failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/compiler-routing-copy.md", "Compiler Routing Copy", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate candidate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify candidate duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("candidate pressure passed with duplicate synthesis: %+v", result)
	}
}
