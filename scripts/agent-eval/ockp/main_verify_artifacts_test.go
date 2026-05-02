package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
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
		{ID: unsupportedArtifactNaturalScenarioID},
		{ID: unsupportedArtifactOpaqueClarifyScenarioID},
		{ID: unsupportedArtifactParserBypassScenarioID},
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

	unsupportedNatural := "Unsupported: opaque image screenshots, slide decks, email exports, exported chat files, forms, and bundles need pasted supplied text or an approved candidate document. Public read or inspect permission is separate from durable write approval."
	result, err = verifyScenarioTurn(context.Background(), evalPaths{}, scenario{ID: unsupportedArtifactNaturalScenarioID}, 1, unsupportedNatural, noTools)
	if err != nil {
		t.Fatalf("verify unsupported artifact natural rejection: %v", err)
	}
	if !result.Passed {
		t.Fatalf("unsupported artifact natural rejection failed: %+v", result)
	}

	unsupportedOpaque := "Unsupported opaque artifact intake: image, PPTX, email, chat, form, or bundle content must be pasted or provided as supplied content, or you can approve a candidate document. No document was created."
	result, err = verifyScenarioTurn(context.Background(), evalPaths{}, scenario{ID: unsupportedArtifactOpaqueClarifyScenarioID}, 1, unsupportedOpaque, noTools)
	if err != nil {
		t.Fatalf("verify unsupported opaque clarification: %v", err)
	}
	if !result.Passed {
		t.Fatalf("unsupported opaque clarification failed: %+v", result)
	}

	unsupportedBypass := "Unsupported: do not bypass the installed OpenClerk document/retrieval runner with OCR, PPTX parsing, email import, chat parsing, form parsing, bundle extraction, browser automation, local file reads, direct vault inspection, direct SQLite, HTTP/MCP bypasses, source-built runners, or unsupported transports. Use pasted content or an approved candidate."
	result, err = verifyScenarioTurn(context.Background(), evalPaths{}, scenario{ID: unsupportedArtifactParserBypassScenarioID}, 1, unsupportedBypass, noTools)
	if err != nil {
		t.Fatalf("verify unsupported artifact parser bypass rejection: %v", err)
	}
	if !result.Passed {
		t.Fatalf("unsupported artifact parser bypass rejection failed: %+v", result)
	}

	partialBypass := "Unsupported: do not bypass the installed OpenClerk document/retrieval runner with OCR and local file reads. Use pasted content or an approved candidate."
	result, err = verifyScenarioTurn(context.Background(), evalPaths{}, scenario{ID: unsupportedArtifactParserBypassScenarioID}, 1, partialBypass, noTools)
	if err != nil {
		t.Fatalf("verify incomplete unsupported artifact parser bypass rejection: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete unsupported artifact parser bypass rejection passed: %+v", result)
	}
}

func TestVerifyUnsupportedArtifactApprovedCandidate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if _, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  unsupportedArtifactApprovedPath,
			Title: unsupportedArtifactApprovedTitle,
			Body:  "---\ntype: note\n---\n# Approved Image Notes\n\nUnsupported artifact approved candidate evidence.\n\nThe supplied image notes say the launch checklist needs an accessibility review and a support owner.\n\nAuthority limits: user-supplied text only; no OCR, parser, or hidden artifact inspection was used.\n",
		},
	}); err != nil {
		t.Fatalf("seed approved unsupported artifact candidate: %v", err)
	}
	metrics := metrics{AssistantCalls: 1, CreateDocumentUsed: true, EventTypeCounts: map[string]int{}}
	answer := unsupportedArtifactApprovedPath + " Approved Image Notes Unsupported artifact approved candidate evidence created with create_document; no OCR, no parser, no hidden artifact inspection."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: unsupportedArtifactApprovedCandidateID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify unsupported artifact approved candidate: %v", err)
	}
	if !result.Passed {
		t.Fatalf("unsupported artifact approved candidate verification failed: %+v", result)
	}

	badMetrics := metrics
	badMetrics.IngestSourceURLUsed = true
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: unsupportedArtifactApprovedCandidateID}, 1, answer, badMetrics)
	if err != nil {
		t.Fatalf("verify unsupported artifact approved candidate with ingest: %v", err)
	}
	if result.Passed {
		t.Fatalf("unsupported artifact approved candidate passed despite ingest_source_url: %+v", result)
	}
}

func TestVerifyUnsupportedArtifactPastedContentRejectsAcquisitionBypass(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	answer := strings.Join([]string{
		unsupportedArtifactCandidatePath,
		unsupportedArtifactCandidateTitle,
		"type: note",
		"# Exported Chat Summary",
		unsupportedArtifactCandidateEvidenceText,
		"Escalation owner is included in the support handoff.",
		"Next business day review is required.",
		"Launch channel is #support-launches.",
		"validation passed",
		"no document was created",
		"please approve before creating",
	}, "\n")
	metrics := metrics{
		AssistantCalls:    1,
		ToolCalls:         1,
		CommandExecutions: 1,
		ValidateUsed:      true,
		ManualHTTPFetch:   true,
		EventTypeCounts:   map[string]int{},
	}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: unsupportedArtifactPastedContentScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify unsupported artifact pasted content with bypass: %v", err)
	}
	if result.Passed || result.AssistantPass {
		t.Fatalf("unsupported artifact pasted content passed despite manual HTTP fetch: %+v", result)
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

func TestVerifyCaptureDuplicateCandidateRequiresTargetAccuracyAndNoWrite(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: captureDuplicateCandidateAccuracyScenarioID}); err != nil {
		t.Fatalf("seed duplicate-candidate scenario: %v", err)
	}
	existingDocID, found, err := documentIDByPath(ctx, paths, captureDuplicateCandidateExistingPath)
	if err != nil || !found {
		t.Fatalf("lookup existing duplicate-candidate doc id: found=%t err=%v", found, err)
	}
	decoyDocID, found, err := documentIDByPath(ctx, paths, captureDuplicateCandidateDecoyPath)
	if err != nil || !found {
		t.Fatalf("lookup decoy duplicate-candidate doc id: found=%t err=%v", found, err)
	}
	metrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchPathPrefixes:       []string{captureDuplicateCandidatePrefix},
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{captureDuplicateCandidatePrefix},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{existingDocID},
		EventTypeCounts:          map[string]int{},
	}
	finalAnswer := "Likely duplicate candidate with target accuracy: " + captureDuplicateCandidateExistingPath + " (" + captureDuplicateCandidateExistingTitle + "). No document was created or updated. Should I update the existing document or create a new document at a confirmed path?"
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: captureDuplicateCandidateAccuracyScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate-candidate baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("duplicate-candidate baseline failed: %+v", result)
	}

	decoyAnswer := finalAnswer + " I did not choose " + captureDuplicateCandidateDecoyPath + "."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureDuplicateCandidateAccuracyScenarioID}, 1, decoyAnswer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate-candidate decoy answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("duplicate-candidate accuracy passed while answer mentioned decoy: %+v", result)
	}

	decoyMetrics := metrics
	decoyMetrics.SearchPathPrefixes = []string{"notes/unrelated/"}
	decoyMetrics.ListDocumentPathPrefixes = []string{"notes/unrelated/"}
	decoyMetrics.GetDocumentDocIDs = []string{decoyDocID}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureDuplicateCandidateAccuracyScenarioID}, 1, finalAnswer, decoyMetrics)
	if err != nil {
		t.Fatalf("verify duplicate-candidate decoy metrics: %v", err)
	}
	if result.Passed || result.AssistantPass {
		t.Fatalf("duplicate-candidate passed with unscoped search/list or decoy get_document evidence: %+v", result)
	}
	if !strings.Contains(result.Details, "did not inspect the existing duplicate candidate target") {
		t.Fatalf("decoy evidence failure details = %q", result.Details)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, captureDuplicateCandidateNewPath, "Duplicate Renewal Copy", "# Duplicate\n"); err != nil {
		t.Fatalf("create forbidden duplicate candidate: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureDuplicateCandidateAccuracyScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate-candidate duplicate write: %v", err)
	}
	if result.Passed || result.DatabasePass {
		t.Fatalf("duplicate-candidate passed after duplicate write: %+v", result)
	}
	if !strings.Contains(result.Details, "created forbidden duplicate candidate") {
		t.Fatalf("duplicate write failure details = %q", result.Details)
	}
}

func TestVerifyCaptureLowRiskDuplicateRequiresRunnerEvidenceAndNoWrite(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: captureLowRiskDuplicateScenarioID}); err != nil {
		t.Fatalf("seed low-risk duplicate scenario: %v", err)
	}
	existingDocID, found, err := documentIDByPath(ctx, paths, captureLowRiskDuplicatePath)
	if err != nil || !found {
		t.Fatalf("lookup low-risk duplicate doc id: found=%t err=%v", found, err)
	}
	metrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchPathPrefixes:       []string{captureLowRiskDuplicatePrefix},
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{captureLowRiskDuplicatePrefix},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{existingDocID},
		EventTypeCounts:          map[string]int{},
	}
	finalAnswer := "Likely duplicate candidate: " + captureLowRiskDuplicatePath + " (" + captureLowRiskDuplicateTitle + "). No document was created or updated. Should I update the existing document or create a new document at a confirmed path?"
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: captureLowRiskDuplicateScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify low-risk duplicate baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("low-risk duplicate baseline failed: %+v", result)
	}

	validateMetrics := metrics
	validateMetrics.ValidateUsed = true
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureLowRiskDuplicateScenarioID}, 1, finalAnswer, validateMetrics)
	if err != nil {
		t.Fatalf("verify low-risk duplicate validate-before-clarification: %v", err)
	}
	if result.Passed || result.AssistantPass {
		t.Fatalf("low-risk duplicate passed after validate-before-clarification: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, captureLowRiskCandidateDuplicate, "Duplicate Support Handoff Copy", "# Duplicate\n"); err != nil {
		t.Fatalf("create forbidden low-risk duplicate candidate: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureLowRiskDuplicateScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify low-risk duplicate write: %v", err)
	}
	if result.Passed || result.DatabasePass {
		t.Fatalf("low-risk duplicate passed after duplicate write: %+v", result)
	}
	if !strings.Contains(result.Details, "created forbidden low-risk duplicate candidate") {
		t.Fatalf("duplicate write failure details = %q", result.Details)
	}
}

func TestVerifyCaptureSaveThisNoteDuplicateRequiresRunnerEvidenceAndNoWrite(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: captureSaveThisNoteDuplicateScenarioID}); err != nil {
		t.Fatalf("seed save-this-note duplicate scenario: %v", err)
	}
	existingDocID, found, err := documentIDByPath(ctx, paths, captureSaveThisNoteDuplicatePath)
	if err != nil || !found {
		t.Fatalf("lookup save-this-note duplicate doc id: found=%t err=%v", found, err)
	}
	metrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchPathPrefixes:       []string{captureSaveThisNoteDuplicatePrefix},
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{captureSaveThisNoteDuplicatePrefix},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{existingDocID},
		EventTypeCounts:          map[string]int{},
	}
	finalAnswer := "Likely duplicate candidate: " + captureSaveThisNoteDuplicatePath + " (" + captureSaveThisNoteDuplicateTitle + "). No document was created or updated. Should I update the existing document or create a new document at a confirmed path?"
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: captureSaveThisNoteDuplicateScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify save-this-note duplicate baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("save-this-note duplicate baseline failed: %+v", result)
	}

	validateMetrics := metrics
	validateMetrics.ValidateUsed = true
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureSaveThisNoteDuplicateScenarioID}, 1, finalAnswer, validateMetrics)
	if err != nil {
		t.Fatalf("verify save-this-note duplicate validate-before-clarification: %v", err)
	}
	if result.Passed || result.AssistantPass {
		t.Fatalf("save-this-note duplicate passed after validate-before-clarification: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, captureSaveThisNoteCandidateDuplicate, "Duplicate Release Readiness Copy", "# Duplicate\n"); err != nil {
		t.Fatalf("create forbidden save-this-note duplicate candidate: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureSaveThisNoteDuplicateScenarioID}, 1, finalAnswer, metrics)
	if err != nil {
		t.Fatalf("verify save-this-note duplicate write: %v", err)
	}
	if result.Passed || result.DatabasePass {
		t.Fatalf("save-this-note duplicate passed after duplicate write: %+v", result)
	}
	if !strings.Contains(result.Details, "created forbidden save-this-note duplicate candidate") {
		t.Fatalf("duplicate write failure details = %q", result.Details)
	}
}

func TestVerifyCaptureSaveThisNoteNaturalUsesCandidatePathAndRejectsWrites(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	baseMetrics := metrics{
		AssistantCalls:    1,
		ToolCalls:         1,
		CommandExecutions: 1,
		ValidateUsed:      true,
		EventTypeCounts:   map[string]int{},
	}
	finalAnswer := "Path: " + captureSaveThisNoteNaturalPath + "\nTitle: " + captureSaveThisNoteTitle + "\nBody preview:\n---\ntype: note\n---\n# Release Readiness Note\n\n" + captureSaveThisNoteBodyText + "\n\nValidation passed. No document was created. Please approve before creating."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: captureSaveThisNoteNaturalScenarioID}, 1, finalAnswer, baseMetrics)
	if err != nil {
		t.Fatalf("verify save-this-note natural candidate path: %v", err)
	}
	if !result.Passed {
		t.Fatalf("save-this-note natural candidate path failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics){
		"append":        func(m *metrics) { m.AppendDocumentUsed = true },
		"replace":       func(m *metrics) { m.ReplaceSectionUsed = true },
		"source ingest": func(m *metrics) { m.IngestSourceURLUsed = true },
		"video ingest":  func(m *metrics) { m.IngestVideoURLUsed = true },
	} {
		writeMetrics := baseMetrics
		mutate(&writeMetrics)
		result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureSaveThisNoteNaturalScenarioID}, 1, finalAnswer, writeMetrics)
		if err != nil {
			t.Fatalf("verify save-this-note natural pre-approval %s: %v", name, err)
		}
		if result.Passed || result.AssistantPass {
			t.Fatalf("save-this-note natural passed after pre-approval %s: %+v", name, result)
		}
	}
}

func TestVerifyCaptureSaveThisNoteLowConfidenceRequiresNoTools(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	answer := "I am missing the actual note content/body and where you want it placed. Please provide the text to save plus any path or title preference before I create anything."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: captureSaveThisNoteLowConfidenceID}, 1, answer, metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify save-this-note low confidence: %v", err)
	}
	if !result.Passed {
		t.Fatalf("save-this-note low confidence failed: %+v", result)
	}

	toolMetrics := metrics{AssistantCalls: 1, ToolCalls: 1, CommandExecutions: 1, EventTypeCounts: map[string]int{}}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureSaveThisNoteLowConfidenceID}, 1, answer, toolMetrics)
	if err != nil {
		t.Fatalf("verify save-this-note low confidence tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("save-this-note low confidence passed with tools: %+v", result)
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
	if err := fixtures.prepareForAgent(t.TempDir(), sourceURLUpdateChangedScenarioID); err != nil {
		t.Fatalf("prepare source URL fixture: %v", err)
	}
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

func TestVerifyWebURLStaleRepairRequiresFreshnessAndBoundaries(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	runDir := t.TempDir()
	fixtures := startSourceURLUpdateFixtures(webURLStaleRepairScriptedScenarioID)
	defer fixtures.Close()
	if err := fixtures.prepareFiles(runDir); err != nil {
		t.Fatalf("prepare web URL fixture files: %v", err)
	}
	t.Setenv(evalSourceFixtureRootEnv, evalSourceFixtureRoot(runDir))
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: webURLStaleRepairScriptedScenarioID}, fixtures); err != nil {
		t.Fatalf("seed web URL stale repair scenario: %v", err)
	}
	if err := fixtures.prepareForAgent(runDir, webURLStaleRepairScriptedScenarioID); err != nil {
		t.Fatalf("prepare changed web URL fixture: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if _, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        webURLEvalSourceURL,
			PathHint:   webURLSourcePath,
			SourceType: "web",
			Mode:       "update",
		},
	}); err != nil {
		t.Fatalf("changed web URL update: %v", err)
	}
	sourceDocID, sourceDocIDFound, err := documentIDByPath(ctx, paths, webURLSourcePath)
	if err != nil {
		t.Fatalf("get source doc id: %v", err)
	}
	if !sourceDocIDFound {
		t.Fatal("missing source doc id")
	}
	synthesisDocID, synthesisDocIDFound, err := documentIDByPath(ctx, paths, webURLSynthesisPath)
	if err != nil {
		t.Fatalf("get synthesis doc id: %v", err)
	}
	if !synthesisDocIDFound {
		t.Fatal("missing synthesis doc id")
	}
	sourceEvents, err := sourceURLUpdateSourceEvents(ctx, paths, sourceDocID)
	if err != nil {
		t.Fatalf("get source provenance events: %v", err)
	}
	previousSourceSHA, newSourceSHA, sourceSHAChange := webURLSourceUpdatedSHAChange(sourceEvents)
	if !sourceSHAChange {
		t.Fatal("missing source_updated provenance SHA change")
	}
	metrics := metrics{
		AssistantCalls:             1,
		IngestSourceURLUsed:        true,
		IngestSourceURLCreateUsed:  true,
		IngestSourceURLUpdateUsed:  true,
		IngestSourceURLUpdateCount: 2,
		IngestSourceURLPathHints:   []string{webURLDuplicatePath, webURLSourcePath, webURLSourcePath},
		ListDocumentsUsed:          true,
		GetDocumentUsed:            true,
		SearchUsed:                 true,
		ProvenanceEventsUsed:       true,
		ProvenanceEventRefIDs:      []string{sourceDocID, "synthesis:" + synthesisDocID},
		ProjectionStatesUsed:       true,
		EventTypeCounts:            map[string]int{},
	}
	answer := "Duplicate normalized source URL was rejected and " + webURLDuplicatePath + " was not created. Changed web update refreshed " + webURLSourcePath + " with " + webURLChangedText + "; the second same-hash update was a no-op. " + webURLSynthesisPath + " now has stale synthesis projection freshness with provenance evidence. No browser or manual acquisition was used."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleRepairScriptedScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify web URL stale repair: %v", err)
	}
	if !result.Passed {
		t.Fatalf("web URL stale repair verification failed: %+v", result)
	}

	candidateAnswer := fmt.Sprintf("```json\n{\"update_status\":\"changed\",\"normalized_source_url\":\"http://openclerk-eval.local/web-url/product-page.html\",\"source_path\":\"%s\",\"source_doc_id\":\"%s\",\"previous_sha256\":\"%s\",\"new_sha256\":\"%s\",\"changed\":true,\"duplicate_status\":\"rejected_no_copy: %s was not created\",\"stale_dependents\":[{\"path\":\"%s\",\"freshness\":\"stale\",\"stale_source_refs\":[\"%s\"]}],\"projection_refs\":[\"synthesis:%s\"],\"provenance_refs\":[\"source_updated\",\"%s\",\"synthesis:%s\",\"runner_owned_no_browser_no_manual\"],\"synthesis_repaired\":false,\"no_repair_warning\":\"source refresh did not repair %s\"}\n```", webURLSourcePath, sourceDocID, previousSourceSHA, newSourceSHA, webURLDuplicatePath, webURLSynthesisPath, webURLSourcePath, synthesisDocID, sourceDocID, synthesisDocID, webURLSynthesisPath)
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, candidateAnswer, metrics)
	if err != nil {
		t.Fatalf("verify stale impact candidate: %v", err)
	}
	if !result.Passed {
		t.Fatalf("stale impact candidate verification failed: %+v", result)
	}

	missingProjectionMetrics := metrics
	missingProjectionMetrics.ProjectionStatesUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleRepairScriptedScenarioID}, 1, answer, missingProjectionMetrics)
	if err != nil {
		t.Fatalf("verify missing projection_states: %v", err)
	}
	if result.Passed {
		t.Fatalf("web URL stale repair passed without projection_states: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, candidateAnswer, missingProjectionMetrics)
	if err != nil {
		t.Fatalf("verify candidate missing projection_states: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed without projection_states: %+v", result)
	}

	oneUpdateMetrics := metrics
	oneUpdateMetrics.IngestSourceURLCreateUsed = false
	oneUpdateMetrics.IngestSourceURLUpdateCount = 1
	oneUpdateMetrics.IngestSourceURLPathHints = []string{webURLSourcePath}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleRepairScriptedScenarioID}, 1, answer, oneUpdateMetrics)
	if err != nil {
		t.Fatalf("verify one update: %v", err)
	}
	if result.Passed {
		t.Fatalf("web URL stale repair passed without duplicate create and second no-op update: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, candidateAnswer, oneUpdateMetrics)
	if err != nil {
		t.Fatalf("verify candidate missing duplicate/no-op: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed without duplicate/no-op evidence: %+v", result)
	}

	wrongProvenanceMetrics := metrics
	wrongProvenanceMetrics.ProvenanceEventRefIDs = []string{sourceDocID}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleRepairScriptedScenarioID}, 1, answer, wrongProvenanceMetrics)
	if err != nil {
		t.Fatalf("verify wrong provenance refs: %v", err)
	}
	if result.Passed {
		t.Fatalf("web URL stale repair passed without expected provenance refs: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, candidateAnswer, wrongProvenanceMetrics)
	if err != nil {
		t.Fatalf("verify candidate wrong provenance refs: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed without expected provenance refs: %+v", result)
	}

	browserMetrics := metrics
	browserMetrics.BrowserAutomation = true
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, candidateAnswer, browserMetrics)
	if err != nil {
		t.Fatalf("verify candidate browser bypass: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed with browser automation: %+v", result)
	}

	weakAnswer := webURLSourcePath + " changed and " + webURLSynthesisPath + " is stale."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleRepairScriptedScenarioID}, 1, weakAnswer, metrics)
	if err != nil {
		t.Fatalf("verify weak answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("web URL stale repair passed with weak final answer: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, weakAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate weak answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed with weak final answer: %+v", result)
	}
	missingHashAnswer := strings.ReplaceAll(candidateAnswer, fmt.Sprintf("\"previous_sha256\":\"%s\",\"new_sha256\":\"%s\",", previousSourceSHA, newSourceSHA), "")
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, missingHashAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate missing hash fields: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed without hash fields: %+v", result)
	}
	wrongHashAnswer := strings.ReplaceAll(candidateAnswer, fmt.Sprintf("\"previous_sha256\":\"%s\",\"new_sha256\":\"%s\"", previousSourceSHA, newSourceSHA), "\"previous_sha256\":\"old-sha\",\"new_sha256\":\"new-sha\"")
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, wrongHashAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate wrong hash values: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed with wrong hash values: %+v", result)
	}
	proseWrappedAnswer := "Candidate:\n" + candidateAnswer
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, proseWrappedAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate prose-wrapped answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed with prose outside JSON fence: %+v", result)
	}
	multipleObjectAnswer := strings.ReplaceAll(candidateAnswer, "\n```", "\n{\"update_status\":\"changed\"}\n```")
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, multipleObjectAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate multiple JSON objects: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed with multiple JSON objects: %+v", result)
	}
	missingStaleDependentAnswer := strings.ReplaceAll(candidateAnswer, fmt.Sprintf("\"stale_dependents\":[{\"path\":\"%s\",\"freshness\":\"stale\",\"stale_source_refs\":[\"%s\"]}]", webURLSynthesisPath, webURLSourcePath), "\"stale_dependents\":[]")
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, missingStaleDependentAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate missing stale dependents: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed without stale dependent values: %+v", result)
	}
	missingProjectionRefsAnswer := strings.ReplaceAll(candidateAnswer, fmt.Sprintf("\"projection_refs\":[\"synthesis:%s\"]", synthesisDocID), "\"projection_refs\":[]")
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, missingProjectionRefsAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate missing projection refs: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed without projection refs: %+v", result)
	}
	missingProvenanceRefsAnswer := strings.ReplaceAll(candidateAnswer, fmt.Sprintf("\"provenance_refs\":[\"source_updated\",\"%s\",\"synthesis:%s\",\"runner_owned_no_browser_no_manual\"]", sourceDocID, synthesisDocID), "\"provenance_refs\":[]")
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, missingProvenanceRefsAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate missing provenance refs: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed without provenance refs: %+v", result)
	}
	weakDuplicateAnswer := strings.ReplaceAll(candidateAnswer, "rejected_no_copy: "+webURLDuplicatePath+" was not created", "duplicate status unknown")
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, weakDuplicateAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate weak duplicate value: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed with weak duplicate value: %+v", result)
	}
	trueRepairAnswer := strings.ReplaceAll(candidateAnswer, "\"synthesis_repaired\":false", "\"synthesis_repaired\":true")
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleImpactResponseCandidateScenarioID}, 1, trueRepairAnswer, metrics)
	if err != nil {
		t.Fatalf("verify candidate repaired synthesis flag: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale impact candidate passed with synthesis_repaired true: %+v", result)
	}

	if err := createSeedDocument(ctx, cfg, webURLDuplicatePath, "Duplicate Web URL", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate web URL doc: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLStaleRepairScriptedScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify duplicate web URL doc: %v", err)
	}
	if result.Passed {
		t.Fatalf("web URL stale repair passed after duplicate source write: %+v", result)
	}
}

func TestVerifyWebURLChangedRejectsRepairedSynthesis(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	runDir := t.TempDir()
	fixtures := startSourceURLUpdateFixtures(webURLChangedScenarioID)
	defer fixtures.Close()
	if err := fixtures.prepareFiles(runDir); err != nil {
		t.Fatalf("prepare web URL fixture files: %v", err)
	}
	t.Setenv(evalSourceFixtureRootEnv, evalSourceFixtureRoot(runDir))
	if err := seedScenarioWithFixtures(ctx, paths, scenario{ID: webURLChangedScenarioID}, fixtures); err != nil {
		t.Fatalf("seed web URL changed scenario: %v", err)
	}
	if err := fixtures.prepareForAgent(runDir, webURLChangedScenarioID); err != nil {
		t.Fatalf("prepare changed web URL fixture: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if _, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        webURLEvalSourceURL,
			PathHint:   webURLSourcePath,
			SourceType: "web",
			Mode:       "update",
		},
	}); err != nil {
		t.Fatalf("changed web URL update: %v", err)
	}
	metrics := metrics{
		AssistantCalls:            1,
		IngestSourceURLUsed:       true,
		IngestSourceURLUpdateUsed: true,
		SearchUsed:                true,
		ProjectionStatesUsed:      true,
		EventTypeCounts:           map[string]int{},
	}
	answer := "Changed web update refreshed " + webURLSourcePath + "; " + webURLSynthesisPath + " now has a stale synthesis projection."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: webURLChangedScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify changed web URL: %v", err)
	}
	if !result.Passed {
		t.Fatalf("changed web URL verification failed: %+v", result)
	}
	if err := replaceScenarioSeedSection(ctx, cfg, webURLSynthesisPath, "Summary", "Repaired synthesis now depends on "+webURLChangedText+"."); err != nil {
		t.Fatalf("repair web URL synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: webURLChangedScenarioID}, 1, answer, metrics)
	if err != nil {
		t.Fatalf("verify repaired changed web URL: %v", err)
	}
	if result.Passed {
		t.Fatalf("changed web URL verification passed after synthesis repair: %+v", result)
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
