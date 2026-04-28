package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

const (
	defaultParallel   = 4
	modelName         = "gpt-5.4-mini"
	reasoningEffort   = "medium"
	productionVariant = "production"
	cacheModeShared   = "shared"
	cacheModeIsolated = "isolated"

	openClerkBootstrapRejectionText = "respond with exactly one no-tools assistant answer"

	ragRetrievalScenarioID   = "rag-retrieval-baseline"
	ragCurrentPolicyPath     = "notes/rag/current-runner-policy.md"
	ragDecoyPolicyPath       = "notes/rag/decoy-runner-policy.md"
	ragArchivedPolicyPath    = "notes/archive/old-runner-policy.md"
	ragSearchText            = "active AgentOps RAG baseline policy JSON runner citations"
	ragPathPrefix            = "notes/rag/"
	ragMetadataKey           = "rag_scope"
	ragMetadataValue         = "active-policy"
	ragCurrentPolicyTitle    = "Current AgentOps RAG Policy"
	ragCurrentPolicySummary  = "Active AgentOps RAG baseline policy marker: routine OpenClerk knowledge answers must use the installed openclerk JSON runner and include source citations with doc_id and chunk_id."
	ragCurrentPolicyDecision = "The active retrieval decision is JSON runner only."
	ragDecoyPolicyTitle      = "Decoy AgentOps RAG Policy"
	ragArchivedPolicyTitle   = "Archived AgentOps RAG Policy"

	docsNavigationScenarioID = "canonical-docs-navigation-baseline"
	docsNavigationPrefix     = "notes/wiki/agentops/"
	docsNavigationIndexPath  = "notes/wiki/agentops/index.md"
	docsNavigationPolicyPath = "notes/wiki/agentops/runner-policy.md"
	docsNavigationArchPath   = "notes/wiki/architecture/knowledge-plane.md"
	docsNavigationOpsPath    = "notes/wiki/ops/runner-playbook.md"

	graphSemanticsScenarioID       = "graph-semantics-reference-poc"
	graphSemanticsPrefix           = "notes/graph/semantics/"
	graphSemanticsIndexPath        = "notes/graph/semantics/index.md"
	graphSemanticsRoutingPath      = "notes/graph/semantics/routing.md"
	graphSemanticsFreshnessPath    = "notes/graph/semantics/freshness.md"
	graphSemanticsOperationsPath   = "notes/graph/semantics/operations.md"
	graphSemanticsSearchText       = "graph semantics requires supersedes related operationalizes"
	graphSemanticsRelationshipText = "requires supersedes related to operationalizes"

	memoryRouterScenarioID              = "memory-router-reference-poc"
	memoryRouterPrefix                  = "notes/memory-router/"
	memoryRouterSessionObservationPath  = "notes/memory-router/session-observation.md"
	memoryRouterTemporalPath            = "notes/memory-router/temporal-policy.md"
	memoryRouterFeedbackPath            = "notes/memory-router/feedback-weighting.md"
	memoryRouterRoutingPath             = "notes/memory-router/routing-policy.md"
	memoryRouterSynthesisPath           = "synthesis/memory-router-reference.md"
	memoryRouterSearchText              = "memory router temporal recall session promotion feedback weighting routing canonical docs"
	memoryRouterSessionObservationTitle = "Memory Router Session Observation"

	configuredLayoutScenarioID = "configured-layout-explain"
	invalidLayoutScenarioID    = "invalid-layout-visible"

	sourceURLUpdateLaneName            = "source-url-update"
	sourceURLUpdateDuplicateScenarioID = "source-url-update-duplicate-create"
	sourceURLUpdateSameSHAScenarioID   = "source-url-update-same-sha-noop"
	sourceURLUpdateChangedScenarioID   = "source-url-update-changed-pdf-stale"
	sourceURLUpdateConflictScenarioID  = "source-url-update-path-hint-conflict"
	sourceURLUpdateSourcePath          = "sources/source-url-update-runner.md"
	sourceURLUpdateAssetPath           = "assets/sources/source-url-update-runner.pdf"
	sourceURLUpdateSynthesisPath       = "synthesis/source-url-update-runner.md"
	sourceURLUpdateDuplicatePath       = "sources/source-url-update-runner-copy.md"
	sourceURLUpdateConflictPath        = "sources/source-url-update-conflict.md"
	sourceURLUpdateStableURLToken      = "{{SOURCE_URL_UPDATE_STABLE_URL}}"
	sourceURLUpdateChangedURLToken     = "{{SOURCE_URL_UPDATE_CHANGED_URL}}"
	sourceURLUpdateInitialText         = "SourceURLUpdateInitialEvidence"
	sourceURLUpdateChangedText         = "SourceURLUpdateChangedEvidence"

	synthesisCandidatePressureScenarioID = "synthesis-candidate-pressure"
	synthesisSourceSetPressureScenarioID = "synthesis-source-set-pressure"
	mtSynthesisDriftPressureScenarioID   = "mt-synthesis-drift-pressure"
	decisionRecordVsDocsScenarioID       = "decision-record-vs-docs"
	decisionSupersessionScenarioID       = "decision-supersession-freshness"
	decisionRealADRMigrationScenarioID   = "decision-real-adr-migration"
	sourceAuditRepairScenarioID          = "source-sensitive-audit-repair"
	sourceAuditConflictScenarioID        = "source-sensitive-conflict-explain"
	documentHistoryInspectScenarioID     = "document-history-inspection-control"
	documentHistoryDiffScenarioID        = "document-diff-review-pressure"
	documentHistoryRestoreScenarioID     = "document-restore-rollback-pressure"
	documentHistoryPendingScenarioID     = "document-pending-change-review-pressure"
	documentHistoryStaleScenarioID       = "document-stale-synthesis-after-revision"
	populatedHeterogeneousScenarioID     = "populated-heterogeneous-retrieval"
	populatedFreshnessConflictScenarioID = "populated-freshness-conflict"
	populatedSynthesisUpdateScenarioID   = "populated-synthesis-update-over-duplicate"

	synthesisCandidatePath       = "synthesis/compiler-routing.md"
	synthesisCandidateDecoyPath  = "synthesis/compiler-routing-decoy.md"
	synthesisCandidateCurrentSrc = "sources/compiler-current.md"
	synthesisCandidateOldSrc     = "sources/compiler-old.md"

	synthesisSourceSetPath = "synthesis/compiler-source-set.md"
	sourceSetAlphaPath     = "sources/source-set-alpha.md"
	sourceSetBetaPath      = "sources/source-set-beta.md"
	sourceSetGammaPath     = "sources/source-set-gamma.md"

	mtDriftSynthesisPath = "synthesis/drift-runner.md"
	mtDriftOldSourcePath = "sources/drift-old.md"
	mtDriftCurrentPath   = "sources/drift-current.md"

	sourceAuditSynthesisPath      = "synthesis/audit-runner-routing.md"
	sourceAuditDecoyPath          = "synthesis/audit-runner-decoy.md"
	sourceAuditOldSourcePath      = "sources/audit-runner-old.md"
	sourceAuditCurrentSourcePath  = "sources/audit-runner-current.md"
	sourceAuditConflictAlphaPath  = "sources/audit-conflict-alpha.md"
	sourceAuditConflictBravoPath  = "sources/audit-conflict-bravo.md"
	sourceAuditConflictSearchText = "source sensitive audit conflict runner retention"

	documentHistoryLaneName               = "document-history-review-controls-poc"
	documentHistoryPolicyPath             = "notes/history-review/lifecycle-control.md"
	documentHistoryDiffPreviousPath       = "sources/history-review/diff-previous.md"
	documentHistoryDiffCurrentPath        = "notes/history-review/diff-current.md"
	documentHistoryDiffListPrefix         = "notes/history-review/"
	documentHistoryRestoreSourcePath      = "sources/history-review/restore-authority.md"
	documentHistoryRestoreTargetPath      = "notes/history-review/restore-target.md"
	documentHistoryPendingTargetPath      = "notes/history-review/pending-target.md"
	documentHistoryPendingProposalPath    = "reviews/history-review/pending-change.md"
	documentHistoryStaleOldSourcePath     = "sources/history-review/stale-old.md"
	documentHistoryStaleCurrentSourcePath = "sources/history-review/stale-current.md"
	documentHistoryStaleSynthesisPath     = "synthesis/history-review-stale.md"
	documentHistorySearchText             = "document history review controls semantic lifecycle evidence"
	documentHistoryStaleSearchText        = "history review stale synthesis current revision evidence"

	populatedLaneName               = "populated-vault-targeted"
	populatedDefaultLaneName        = "agentops-production"
	populatedMixedLaneName          = "mixed"
	populatedAuthorityPath          = "sources/populated/atlas-authority.md"
	populatedAuthorityCandidatePath = "sources/populated/atlas-authority-candidate.md"
	populatedPollutedPath           = "sources/populated/atlas-polluted.md"
	populatedTranscriptPath         = "transcripts/atlas-kickoff-transcript.md"
	populatedTranscriptOpsPath      = "transcripts/atlas-ops-standup-transcript.md"
	populatedArticlePath            = "articles/vendor-risk-review.md"
	populatedArticleArchivePath     = "articles/vendor-risk-review-archive.md"
	populatedMeetingPath            = "meetings/atlas-weekly-review.md"
	populatedMeetingBudgetPath      = "meetings/atlas-budget-sync.md"
	populatedDocsPath               = "docs/atlas-operations-guide.md"
	populatedDocsRunbookPath        = "docs/atlas-vendor-runbook.md"
	populatedBlogPath               = "blogs/atlas-launch-draft.md"
	populatedBlogRumorPath          = "blogs/atlas-launch-rumor.md"
	populatedReceiptPath            = "receipts/nebula-office-supply.md"
	populatedReceiptDuplicatePath   = "receipts/nebula-office-supply-copy.md"
	populatedInvoicePath            = "invoices/nebula-consulting-2026-04.md"
	populatedInvoiceStalePath       = "invoices/nebula-consulting-2026-03.md"
	populatedLegalPath              = "legal/data-retention-memo.md"
	populatedLegalArchivePath       = "legal/data-retention-archive.md"
	populatedContractPath           = "contracts/acme-master-services.md"
	populatedContractDraftPath      = "contracts/acme-master-services-draft.md"
	populatedConflictAlphaPath      = "sources/populated/retention-alpha.md"
	populatedConflictBravoPath      = "sources/populated/retention-bravo.md"
	populatedSynthesisPath          = "synthesis/populated-vault-summary.md"
	populatedSynthesisDecoyPath     = "synthesis/populated-vault-summary-decoy.md"
	populatedSynthesisOldPath       = "sources/populated/synthesis-old.md"
	populatedSynthesisCurrentPath   = "sources/populated/synthesis-current.md"
	populatedSearchText             = "Populated vault authority marker"
	populatedConflictSearchText     = "Populated vault retention conflict current source"
	populatedDuplicateSearchText    = "Populated vault duplicate candidate marker"
	populatedStaleSearchText        = "Populated vault stale source marker"
	populatedSynthesisSearchText    = "Current populated vault synthesis guidance"

	repoDocsLaneName              = "repo-docs-dogfood"
	repoDocsRetrievalScenarioID   = "repo-docs-agentops-retrieval"
	repoDocsSynthesisScenarioID   = "repo-docs-synthesis-maintenance"
	repoDocsDecisionScenarioID    = "repo-docs-decision-records"
	repoDocsAgentOpsADRPath       = "docs/architecture/eval-backed-knowledge-plane-adr.md"
	repoDocsKnowledgeConfigPath   = "docs/architecture/knowledge-configuration-v1-adr.md"
	repoDocsAgentProductionPath   = "docs/evals/agent-production.md"
	repoDocsBaselineScenariosPath = "docs/evals/baseline-scenarios.md"
	repoDocsSynthesisPath         = "synthesis/repo-docs-agentops-validation.md"
	repoDocsRetrievalSearchText   = "oc-rsj verified current AgentOps document retrieval runner actions"
	repoDocsSynthesisSearchText   = "production AgentOps gate baseline scenarios runner JSON validation"
	repoDocsDecisionSearchText    = "Knowledge Configuration v1 accepted AgentOps surface"

	agentChosenPathLaneName            = "agent-chosen-path-selection-poc"
	agentChosenExplicitScenarioID      = "explicit-fields-path-title-type"
	agentChosenMissingFieldsScenarioID = "missing-path-title-type-reject"
	agentChosenPathProposalScenarioID  = "url-only-documentation-path-proposal"
	agentChosenAutonomousScenarioID    = "url-only-documentation-autonomous-placement"
	agentChosenSynthesisScenarioID     = "multi-source-synthesis-path-selection"
	agentChosenAmbiguousScenarioID     = "ambiguous-document-type-path-selection"
	agentChosenUserPathScenarioID      = "user-path-instructions-win"
	agentChosenExplicitPath            = "notes/agent-chosen/explicit-fields.md"
	agentChosenProposalPath            = "sources/openai-harness-and-prompt-guidance.md"
	agentChosenAutonomousPath          = "sources/openai-harness-and-prompt-guidance.md"
	agentChosenUserSpecifiedPath       = "notes/agent-chosen/user-specified.md"
	agentChosenSynthesisPath           = "synthesis/agent-chosen-path-selection.md"
	agentChosenSynthesisAlphaPath      = "sources/agent-chosen/path-alpha.md"
	agentChosenSynthesisBetaPath       = "sources/agent-chosen/path-beta.md"
	agentChosenSynthesisGammaPath      = "sources/agent-chosen/path-gamma.md"
	agentChosenAmbiguousDecisionID     = "adr-agent-chosen-path-metadata-authority"
	agentChosenAmbiguousSearchText     = "metadata authority decides agent chosen path placement"
	agentChosenURLHarness              = "https://openai.com/index/harness-engineering/"
	agentChosenURLPromptGuidance       = "https://developers.openai.com/api/docs/guides/prompt-guidance"

	pathTitleAutonomyLaneName               = "path-title-autonomy-pressure"
	pathTitleURLOnlyScenarioID              = "path-title-url-only-autonomy-pressure"
	pathTitleArtifactMissingHintsScenarioID = "path-title-artifact-missing-hints"
	pathTitleMultiSourceDuplicateScenarioID = "path-title-multisource-duplicate-pressure"
	pathTitleExplicitOverridesScenarioID    = "path-title-explicit-overrides-pressure"
	pathTitleDuplicateRiskScenarioID        = "path-title-duplicate-risk-pressure"
	pathTitleMetadataAuthorityScenarioID    = "path-title-metadata-authority-pressure"
	pathTitleURLOnlyPath                    = "sources/path-title/openai-harness-and-prompt-guidance.md"
	pathTitleExplicitPath                   = "notes/path-title/explicit-override.md"
	pathTitleDuplicateExistingPath          = "sources/path-title/existing-openai-harness.md"
	pathTitleDuplicateCandidatePath         = "sources/path-title/openai-harness-duplicate.md"
	pathTitleSynthesisPath                  = "synthesis/path-title-autonomy.md"
	pathTitleSynthesisDuplicatePath         = "synthesis/path-title-autonomy-copy.md"
	pathTitleSynthesisAlphaPath             = "sources/path-title/alpha.md"
	pathTitleSynthesisBetaPath              = "sources/path-title/beta.md"
	pathTitleMetadataPath                   = "records/decisions/path-title-metadata-authority.md"
	pathTitleMetadataDecisionID             = "adr-path-title-metadata-authority"
	pathTitleMetadataSearchText             = "path title metadata authority pressure"
	pathTitleArtifactMissingHintsURL        = "https://example.test/path-title-artifact.pdf"
	pathTitleURLOnlyTitle                   = "OpenAI Harness and Prompt Guidance"
	pathTitleExplicitTitle                  = "Path Title Explicit Override"
	pathTitleSynthesisTitle                 = "Path Title Autonomy"
	pathTitleMetadataTitle                  = "Path Title Metadata Authority"

	documentThisLaneName                        = "document-this-intake-pressure"
	documentThisMissingFieldsScenarioID         = "document-this-missing-fields"
	documentThisExplicitCreateScenarioID        = "document-this-explicit-create"
	documentThisSourceURLMissingHintsScenarioID = "document-this-source-url-missing-hints"
	documentThisExplicitOverridesScenarioID     = "document-this-explicit-overrides"
	documentThisDuplicateCandidateScenarioID    = "document-this-duplicate-candidate"
	documentThisExistingUpdateScenarioID        = "document-this-existing-update"
	documentThisSynthesisFreshnessScenarioID    = "document-this-synthesis-freshness"
	documentThisExplicitPath                    = "notes/document-this/explicit-create.md"
	documentThisExplicitTitle                   = "Document This Explicit Create"
	documentThisOverridePath                    = "notes/document-this/explicit-override.md"
	documentThisOverrideTitle                   = "Document This Explicit Override"
	documentThisDuplicateExistingPath           = "sources/document-this/existing-article.md"
	documentThisDuplicateCandidatePath          = "sources/document-this/duplicate-article.md"
	documentThisUpdateTargetPath                = "notes/document-this/existing-update.md"
	documentThisUpdateDecoyPath                 = "notes/document-this/existing-update-decoy.md"
	documentThisSynthesisPath                   = "synthesis/document-this-intake.md"
	documentThisSynthesisDuplicatePath          = "synthesis/document-this-intake-copy.md"
	documentThisArticlePath                     = "sources/document-this/article.md"
	documentThisDocsPath                        = "sources/document-this/docs-page.md"
	documentThisPaperPath                       = "sources/document-this/paper.md"
	documentThisTranscriptPath                  = "transcripts/document-this/standup.md"
	documentThisSearchText                      = "document this intake pressure article docs paper transcript mixed source"

	documentArtifactCandidateLaneName          = "document-artifact-candidate-generation"
	candidateNoteFromPastedContentScenarioID   = "candidate-note-from-pasted-content"
	candidateTitleAndPathFromHeadingScenarioID = "candidate-title-and-path-from-heading"
	candidateMixedSourceSummaryScenarioID      = "candidate-mixed-source-summary"
	candidateExplicitOverridesWinScenarioID    = "candidate-explicit-overrides-win"
	candidateDuplicateRiskAsksScenarioID       = "candidate-duplicate-risk-asks"
	candidateLowConfidenceAsksScenarioID       = "candidate-low-confidence-asks"
	candidateBodyFaithfulnessScenarioID        = "candidate-body-faithfulness"
	candidateErgonomicsNaturalIntentScenarioID = "candidate-ergonomics-natural-intent"
	candidateErgonomicsScriptedControlID       = "candidate-ergonomics-scripted-control"
	candidateErgonomicsDuplicateNaturalID      = "candidate-ergonomics-duplicate-natural-intent"
	candidateErgonomicsLowConfidenceNaturalID  = "candidate-ergonomics-low-confidence-natural"
	candidateNotePath                          = "notes/candidates/meeting-capture-policy.md"
	candidateNoteTitle                         = "Meeting Capture Policy"
	candidateHeadingPath                       = "notes/candidates/release-risk-review.md"
	candidateHeadingTitle                      = "Release Risk Review"
	candidateMixedSourcePath                   = "notes/candidates/harness-prompt-guidance-summary.md"
	candidateMixedSourceTitle                  = "Harness and Prompt Guidance Summary"
	candidateOverridePath                      = "archive/custom/intake-override.md"
	candidateOverrideTitle                     = "Custom Intake Override"
	candidateDuplicateExistingPath             = "notes/candidates/existing-pricing-note.md"
	candidateDuplicateCandidatePath            = "notes/candidates/pricing-model-note.md"
	candidateBodyFaithfulnessPath              = "notes/candidates/customer-escalation-summary.md"
	candidateBodyFaithfulnessTitle             = "Customer Escalation Summary"
	candidateErgonomicsNaturalPath             = "notes/candidates/release-readiness-checklist.md"
	candidateErgonomicsNaturalTitle            = "Release Readiness Checklist"
	candidateDuplicateSearchText               = "candidate generation duplicate pricing model marker"

	artifactIngestionLaneName            = "heterogeneous-artifact-ingestion-pressure"
	artifactPDFSourceURLScenarioID       = "artifact-pdf-source-url-ingestion"
	artifactPDFNaturalIntentScenarioID   = "artifact-pdf-source-url-natural-intent"
	artifactTranscriptScenarioID         = "artifact-transcript-canonical-markdown"
	artifactInvoiceReceiptScenarioID     = "artifact-invoice-receipt-authority"
	artifactMixedSynthesisScenarioID     = "artifact-mixed-synthesis-freshness"
	artifactSourceMissingHintsScenarioID = "artifact-source-url-missing-hints"
	artifactUnsupportedVideoScenarioID   = "artifact-unsupported-native-video-ingest"
	artifactBypassScenarioID             = "artifact-ingestion-bypass-reject"
	artifactPDFSourcePath                = "sources/artifacts/vendor-security-paper.md"
	artifactPDFAssetPath                 = "assets/sources/artifacts/vendor-security-paper.pdf"
	artifactPDFNaturalSourcePath         = "sources/artifacts/vendor-security-paper-natural.md"
	artifactPDFNaturalAssetPath          = "assets/sources/artifacts/vendor-security-paper-natural.pdf"
	artifactPDFSourceURLToken            = "{{ARTIFACT_PDF_SOURCE_URL}}"
	artifactPDFEvalSourceURL             = "http://openclerk-eval.local/artifacts/vendor-security-paper.pdf"
	evalSourceFixtureRootEnv             = "OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT"
	artifactPDFEvidenceText              = "ArtifactPDFIngestionEvidence"
	artifactTranscriptPath               = "transcripts/artifacts/vendor-demo-transcript.md"
	artifactInvoicePath                  = "invoices/artifacts/atlas-platform-2026-04.md"
	artifactReceiptPath                  = "receipts/artifacts/nebula-usb-c-hub.md"
	artifactMixedSynthesisPath           = "synthesis/artifact-ingestion-pressure.md"
	artifactMixedSynthesisOldPath        = "sources/artifacts/mixed-old.md"
	artifactMixedSynthesisCurrentPath    = "sources/artifacts/mixed-current.md"
	artifactTranscriptEvidenceText       = "Artifact transcript canonical markdown evidence"
	artifactInvoiceReceiptEvidenceText   = "Artifact invoice receipt authority evidence"
	artifactMixedSynthesisEvidenceText   = "Artifact mixed synthesis freshness evidence"

	videoYouTubeLaneName                     = "video-youtube-canonical-source-note"
	videoYouTubeNaturalIntentScenarioID      = "video-youtube-natural-intent"
	videoYouTubeScriptedTranscriptControlID  = "video-youtube-scripted-transcript-control"
	videoYouTubeSynthesisFreshnessScenarioID = "video-youtube-synthesis-freshness"
	videoYouTubeBypassRejectScenarioID       = "video-youtube-bypass-reject"
	videoYouTubeSourcePath                   = "sources/video-youtube/platform-demo-transcript.md"
	videoYouTubeOldSourcePath                = "sources/video-youtube/platform-demo-old.md"
	videoYouTubeCurrentSourcePath            = "sources/video-youtube/platform-demo-current.md"
	videoYouTubeSynthesisPath                = "synthesis/video-youtube-ingestion-pressure.md"
	videoYouTubeSourceEvidenceText           = "Video YouTube canonical source note evidence"
	videoYouTubeSynthesisCurrentEvidenceText = "Video YouTube synthesis freshness current transcript evidence"
	videoYouTubeSynthesisUpdatedEvidenceText = "Video YouTube synthesis freshness updated transcript evidence"
	videoYouTubeURL                          = "https://youtube.example.test/watch?v=video-demo"
	videoYouTubeTranscriptOrigin             = "user_supplied_transcript"
)

var (
	prewarmCompilePackages     = []string{"./cmd/openclerk", "./internal/runner"}
	unixHomePathPattern        = regexp.MustCompile(`/(Users|home)/[^/\s"'\\]+`)
	windowsHomePathPattern     = regexp.MustCompile(`(?i)[A-Z]:\\Users\\[^\\\s"']+`)
	unixAbsolutePathPattern    = regexp.MustCompile(`(^|[\s"'(])/[A-Za-z0-9._-][^\s"']*`)
	windowsDrivePathPattern    = regexp.MustCompile(`(?i)\b[A-Z]:[\\/][^\s"']+`)
	layoutExplicitValidPattern = regexp.MustCompile(`\bvalid\s*[:=]?\s*true\b|\blayout(?:\s+\w+){0,3}\s+valid\b|\bvalid\s+layout\b`)
	layoutInvalidStatusPattern = regexp.MustCompile(`\binvalid\b|\bvalid\s*[:=]?\s*false\b|\bnot\s+valid\b`)
	layoutValidStatusPattern   = regexp.MustCompile(`\bvalid\b|\bpass(?:es|ed)?\b`)
)

type runConfig struct {
	Parallel   int
	Variant    string
	Scenario   string
	RunRoot    string
	ReportDir  string
	ReportName string
	CodexBin   string
	RepoRoot   string
	CacheMode  string
}

type cacheConfig struct {
	Mode    string
	RunRoot string
}

type evalJob struct {
	Index    int
	Variant  string
	Scenario scenario
}

type scenario struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Prompt string         `json:"prompt,omitempty"`
	Turns  []scenarioTurn `json:"turns,omitempty"`
}

type scenarioTurn struct {
	Prompt string `json:"prompt"`
}

type report struct {
	Metadata            reportMetadata         `json:"metadata"`
	Results             []jobResult            `json:"results"`
	ProductionGate      *productionGateSummary `json:"production_gate,omitempty"`
	TargetedLaneSummary *targetedLaneSummary   `json:"targeted_lane_summary,omitempty"`
}

type reportMetadata struct {
	GeneratedAt              time.Time    `json:"generated_at"`
	Model                    string       `json:"model"`
	ReasoningEffort          string       `json:"reasoning_effort"`
	Harness                  string       `json:"harness"`
	ConfiguredParallelism    int          `json:"configured_parallelism"`
	CacheMode                string       `json:"cache_mode"`
	CachePrewarmSeconds      float64      `json:"cache_prewarm_seconds,omitempty"`
	HarnessElapsedSeconds    float64      `json:"harness_elapsed_seconds"`
	EffectiveParallelSpeedup float64      `json:"effective_parallel_speedup,omitempty"`
	ParallelEfficiency       float64      `json:"parallel_efficiency,omitempty"`
	PhaseTotals              phaseTimings `json:"phase_totals"`
	RunRootArtifactReference string       `json:"run_root_artifact_reference"`
	RawLogPlaceholder        string       `json:"raw_log_placeholder"`
	Variants                 []string     `json:"variants"`
	Scenarios                []string     `json:"scenarios"`
	Lane                     string       `json:"lane"`
	ReleaseBlocking          bool         `json:"release_blocking"`
	TargetedAcceptanceNote   string       `json:"targeted_acceptance_note,omitempty"`
	RawLogsCommitted         bool         `json:"raw_logs_committed"`
	RawLogsNote              string       `json:"raw_logs_note"`
}

type phaseTimings struct {
	PrepareRunDir  float64 `json:"prepare_run_dir_seconds,omitempty"`
	CopyRepo       float64 `json:"copy_repo_seconds,omitempty"`
	InstallVariant float64 `json:"install_variant_seconds,omitempty"`
	WarmCache      float64 `json:"warm_cache_seconds,omitempty"`
	SeedData       float64 `json:"seed_data_seconds,omitempty"`
	AgentRun       float64 `json:"agent_run_seconds,omitempty"`
	ParseMetrics   float64 `json:"parse_metrics_seconds,omitempty"`
	Verify         float64 `json:"verify_seconds,omitempty"`
	Total          float64 `json:"total_seconds,omitempty"`
}

type jobResult struct {
	Variant                 string             `json:"variant"`
	Scenario                string             `json:"scenario"`
	ScenarioTitle           string             `json:"scenario_title"`
	Passed                  bool               `json:"passed"`
	Status                  string             `json:"status"`
	Error                   string             `json:"error,omitempty"`
	ExitCode                int                `json:"exit_code"`
	WallSeconds             float64            `json:"wall_seconds"`
	PhaseTimings            phaseTimings       `json:"phase_timings"`
	Metrics                 metrics            `json:"metrics"`
	FixturePreflight        *fixturePreflight  `json:"fixture_preflight,omitempty"`
	Verification            verificationResult `json:"verification"`
	Turns                   []turnResult       `json:"turns,omitempty"`
	PromptSummary           string             `json:"prompt_summary"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
	StartedAt               time.Time          `json:"started_at"`
	CompletedAt             *time.Time         `json:"completed_at,omitempty"`
}

type turnResult struct {
	Index                   int                `json:"turn_index"`
	WallSeconds             float64            `json:"wall_seconds"`
	ExitCode                int                `json:"exit_code"`
	Metrics                 metrics            `json:"metrics"`
	Verification            verificationResult `json:"verification"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
}

type metrics struct {
	AssistantCalls            int            `json:"assistant_calls"`
	ToolCalls                 int            `json:"tool_calls"`
	CommandExecutions         int            `json:"command_executions"`
	FileInspectionCommands    int            `json:"file_inspection_commands"`
	GeneratedFileInspection   bool           `json:"generated_file_inspection"`
	ModuleCacheInspection     bool           `json:"module_cache_inspection"`
	BroadRepoSearch           bool           `json:"broad_repo_search"`
	DirectSQLiteAccess        bool           `json:"direct_sqlite_access"`
	LegacyRunnerUsage         bool           `json:"legacy_runner_usage"`
	SearchUsed                bool           `json:"search_used"`
	SearchUnfilteredUsed      bool           `json:"search_unfiltered_used"`
	SearchPathFilterUsed      bool           `json:"search_path_filter_used"`
	SearchPathPrefixes        []string       `json:"search_path_prefixes,omitempty"`
	SearchMetadataFilterUsed  bool           `json:"search_metadata_filter_used"`
	SearchMetadataFilters     []string       `json:"search_metadata_filters,omitempty"`
	IngestSourceURLUsed       bool           `json:"ingest_source_url_used"`
	IngestSourceURLUpdateUsed bool           `json:"ingest_source_url_update_used"`
	IngestVideoURLUsed        bool           `json:"ingest_video_url_used"`
	IngestVideoURLUpdateUsed  bool           `json:"ingest_video_url_update_used"`
	SourcePDFDownloadFailure  bool           `json:"source_pdf_download_failure"`
	ValidateUsed              bool           `json:"validate_used"`
	CreateDocumentUsed        bool           `json:"create_document_used"`
	ReplaceSectionUsed        bool           `json:"replace_section_used"`
	AppendDocumentUsed        bool           `json:"append_document_used"`
	ListDocumentsUsed         bool           `json:"list_documents_used"`
	ListDocumentPathPrefixes  []string       `json:"list_document_path_prefixes,omitempty"`
	GetDocumentUsed           bool           `json:"get_document_used"`
	GetDocumentDocIDs         []string       `json:"get_document_doc_ids,omitempty"`
	InspectLayoutUsed         bool           `json:"inspect_layout_used"`
	DocumentLinksUsed         bool           `json:"document_links_used"`
	GraphNeighborhoodUsed     bool           `json:"graph_neighborhood_used"`
	RecordsLookupUsed         bool           `json:"records_lookup_used"`
	DecisionsLookupUsed       bool           `json:"decisions_lookup_used"`
	DecisionRecordUsed        bool           `json:"decision_record_used"`
	DecisionRecordIDs         []string       `json:"decision_record_ids,omitempty"`
	ProvenanceEventsUsed      bool           `json:"provenance_events_used"`
	ProvenanceEventRefIDs     []string       `json:"provenance_event_ref_ids,omitempty"`
	ProjectionStatesUsed      bool           `json:"projection_states_used"`
	GeneratedFileEvidence     []string       `json:"generated_file_evidence,omitempty"`
	ModuleCacheEvidence       []string       `json:"module_cache_evidence,omitempty"`
	BroadRepoSearchEvidence   []string       `json:"broad_repo_search_evidence,omitempty"`
	DirectSQLiteEvidence      []string       `json:"direct_sqlite_evidence,omitempty"`
	LegacyRunnerEvidence      []string       `json:"legacy_runner_evidence,omitempty"`
	UsageExposed              bool           `json:"usage_exposed"`
	InputTokens               *int           `json:"input_tokens,omitempty"`
	CachedInputTokens         *int           `json:"cached_input_tokens,omitempty"`
	NonCachedInputTokens      *int           `json:"non_cached_input_tokens,omitempty"`
	OutputTokens              *int           `json:"output_tokens,omitempty"`
	EventTypeCounts           map[string]int `json:"event_type_counts"`
	CommandMetricLimitations  string         `json:"command_metric_limitations"`
}

type verificationResult struct {
	Passed        bool     `json:"passed"`
	DatabasePass  bool     `json:"database_pass"`
	AssistantPass bool     `json:"assistant_pass"`
	Details       string   `json:"details"`
	Documents     []string `json:"documents,omitempty"`
}

type fixturePreflight struct {
	Name       string   `json:"name"`
	Passed     bool     `json:"passed"`
	Details    string   `json:"details,omitempty"`
	Documents  []string `json:"documents,omitempty"`
	SourcePath string   `json:"source_path,omitempty"`
	AssetPath  string   `json:"asset_path,omitempty"`
}

type productionGateSummary struct {
	Variant        string                    `json:"variant"`
	PassesGate     bool                      `json:"passes_gate"`
	Recommendation string                    `json:"recommendation"`
	Criteria       []productionGateCriterion `json:"criteria"`
}

type productionGateCriterion struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

type targetedLaneSummary struct {
	Lane                    string                           `json:"lane"`
	Decision                string                           `json:"decision"`
	PublicSurface           []string                         `json:"public_surface"`
	Promotion               string                           `json:"promotion"`
	ReleaseBlocking         bool                             `json:"release_blocking"`
	ScenarioClassifications []targetedScenarioClassification `json:"scenario_classifications"`
}

type targetedScenarioClassification struct {
	Variant               string  `json:"variant"`
	Scenario              string  `json:"scenario"`
	Status                string  `json:"status"`
	FailureClassification string  `json:"failure_classification"`
	EvidencePosture       string  `json:"evidence_posture"`
	ToolCalls             int     `json:"tool_calls"`
	CommandExecutions     int     `json:"command_executions"`
	AssistantCalls        int     `json:"assistant_calls"`
	WallSeconds           float64 `json:"wall_seconds"`
	PromptSpecificity     string  `json:"prompt_specificity,omitempty"`
	UX                    string  `json:"ux,omitempty"`
	Brittleness           string  `json:"brittleness,omitempty"`
	Retries               int     `json:"retries"`
	StepCount             int     `json:"step_count"`
	Latency               string  `json:"latency,omitempty"`
	GuidanceDependence    string  `json:"guidance_dependence,omitempty"`
	SafetyRisks           string  `json:"safety_risks,omitempty"`
	FixturePreflight      string  `json:"fixture_preflight,omitempty"`
}

type jobRunner func(context.Context, runConfig, evalJob, cacheConfig) jobResult

type codexEvent struct {
	Type     string          `json:"type"`
	ThreadID string          `json:"thread_id"`
	Item     json.RawMessage `json:"item"`
	Usage    *usage          `json:"usage"`
}

type usage struct {
	InputTokens        int           `json:"input_tokens"`
	OutputTokens       int           `json:"output_tokens"`
	CachedInputTokens  int           `json:"cached_input_tokens"`
	InputTokensDetails *usageDetails `json:"input_tokens_details"`
	PromptTokens       int           `json:"prompt_tokens"`
	CompletionTokens   int           `json:"completion_tokens"`
	PromptDetails      *usageDetails `json:"prompt_tokens_details"`
}

type usageDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type parsedTurn struct {
	metrics      metrics
	finalMessage string
	sessionID    string
	parseError   error
	parseSeconds float64
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, codexJobRunner))
}

func run(args []string, stdout io.Writer, stderr io.Writer, runner jobRunner) int {
	if len(args) == 0 || args[0] != "run" {
		_, _ = fmt.Fprintln(stderr, "usage: ockp run [--parallel N] [--variant ids] [--scenario ids] [--run-root path] [--report-dir path] [--codex-bin path] [--cache-mode shared|isolated]")
		return 2
	}
	config, err := parseRunConfig(args[1:], stderr)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	if err := executeRun(context.Background(), config, stdout, runner); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func parseRunConfig(args []string, stderr io.Writer) (runConfig, error) {
	fs := flag.NewFlagSet("ockp run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := runConfig{CacheMode: cacheModeShared}
	fs.IntVar(&config.Parallel, "parallel", defaultParallel, "number of independent eval jobs to run concurrently")
	fs.StringVar(&config.Variant, "variant", "", "comma-separated variant ids")
	fs.StringVar(&config.Scenario, "scenario", "", "comma-separated scenario ids")
	fs.StringVar(&config.RunRoot, "run-root", "", "directory for isolated run artifacts")
	fs.StringVar(&config.ReportDir, "report-dir", filepath.Join("docs", "evals", "results"), "directory for reduced reports")
	fs.StringVar(&config.ReportName, "report-name", "ockp-latest", "base filename for reduced reports, without extension")
	fs.StringVar(&config.CodexBin, "codex-bin", "codex", "codex executable")
	fs.StringVar(&config.RepoRoot, "repo-root", ".", "repository root to copy for each job")
	fs.StringVar(&config.CacheMode, "cache-mode", config.CacheMode, "Go cache mode: shared or isolated")
	if err := fs.Parse(args); err != nil {
		return runConfig{}, err
	}
	if fs.NArg() != 0 {
		return runConfig{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if config.Parallel < 1 {
		return runConfig{}, errors.New("--parallel must be at least 1")
	}
	if config.CacheMode != cacheModeShared && config.CacheMode != cacheModeIsolated {
		return runConfig{}, fmt.Errorf("--cache-mode must be %q or %q", cacheModeShared, cacheModeIsolated)
	}
	if config.RunRoot == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-%d", time.Now().UnixNano()))
	}
	if strings.TrimSpace(config.ReportName) == "" {
		return runConfig{}, errors.New("--report-name must not be empty")
	}
	return config, nil
}

func executeRun(ctx context.Context, config runConfig, stdout io.Writer, runner jobRunner) error {
	start := time.Now()
	jobs, err := buildJobs(config)
	if err != nil {
		return err
	}
	cache := cacheConfig{Mode: config.CacheMode, RunRoot: config.RunRoot}
	cachePrewarmSeconds := 0.0
	if cache.Mode == cacheModeShared {
		cacheStart := time.Now()
		if err := prewarmSharedCache(config.RepoRoot, cache); err != nil {
			return fmt.Errorf("prewarm shared Go cache: %w", err)
		}
		cachePrewarmSeconds = roundSeconds(time.Since(cacheStart).Seconds())
	}
	results := runJobs(ctx, config, jobs, cache, runner)
	elapsed := roundSeconds(time.Since(start).Seconds())
	phaseTotals := aggregatePhaseTimings(results)
	selectedIDs := selectedScenarioIDs(config)
	lane, releaseBlocking := reportLane(selectedIDs)
	effectiveSpeedup := 0.0
	parallelEfficiency := 0.0
	totalAgent := totalAgentWallSeconds(results)
	if elapsed > 0 {
		effectiveSpeedup = roundSeconds(totalAgent / elapsed)
	}
	if config.Parallel > 0 && effectiveSpeedup > 0 {
		parallelEfficiency = roundSeconds(effectiveSpeedup / float64(config.Parallel))
	}
	rep := report{
		Metadata: reportMetadata{
			GeneratedAt:              time.Now().UTC(),
			Model:                    modelName,
			ReasoningEffort:          reasoningEffort,
			Harness:                  "codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral and multi-turn scenarios resume one persisted eval session",
			ConfiguredParallelism:    config.Parallel,
			CacheMode:                cache.Mode,
			CachePrewarmSeconds:      cachePrewarmSeconds,
			HarnessElapsedSeconds:    elapsed,
			EffectiveParallelSpeedup: effectiveSpeedup,
			ParallelEfficiency:       parallelEfficiency,
			PhaseTotals:              phaseTotals,
			RunRootArtifactReference: "<run-root>",
			RawLogPlaceholder:        "<run-root>/<variant>/<scenario>/turn-N/events.jsonl",
			Variants:                 selectedVariants(config),
			Scenarios:                selectedIDs,
			Lane:                     lane,
			ReleaseBlocking:          releaseBlocking,
			TargetedAcceptanceNote:   targetedAcceptanceNote(lane),
			RawLogsCommitted:         false,
			RawLogsNote:              "Raw Codex event logs remain under <run-root> and are not committed.",
		},
		Results:             results,
		ProductionGate:      buildProductionGateSummary(results),
		TargetedLaneSummary: buildTargetedLaneSummary(lane, releaseBlocking, results),
	}
	if err := os.MkdirAll(config.ReportDir, 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}
	jsonPath := filepath.Join(config.ReportDir, config.ReportName+".json")
	markdownPath := filepath.Join(config.ReportDir, config.ReportName+".md")
	if err := writeJSONReport(jsonPath, rep); err != nil {
		return err
	}
	if err := writeMarkdownReport(markdownPath, rep); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "wrote %s and %s\n", filepath.ToSlash(jsonPath), filepath.ToSlash(markdownPath)); err != nil {
		return err
	}
	return nil
}

func buildJobs(config runConfig) ([]evalJob, error) {
	variants := selectedVariants(config)
	scenarios := selectedScenarios(config)
	if len(scenarios) == 0 {
		return nil, errors.New("no scenarios selected")
	}
	jobs := make([]evalJob, 0, len(variants)*len(scenarios))
	for _, variant := range variants {
		for _, scenario := range scenarios {
			jobs = append(jobs, evalJob{
				Index:    len(jobs),
				Variant:  variant,
				Scenario: scenario,
			})
		}
	}
	return jobs, nil
}

func runJobs(ctx context.Context, config runConfig, jobs []evalJob, cache cacheConfig, runner jobRunner) []jobResult {
	results := make([]jobResult, len(jobs))
	jobCh := make(chan evalJob)
	var wg sync.WaitGroup
	workers := min(config.Parallel, max(1, len(jobs)))
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				results[job.Index] = runner(ctx, config, job, cache)
			}
		}()
	}
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)
	wg.Wait()
	return results
}

func codexJobRunner(ctx context.Context, config runConfig, job evalJob, cache cacheConfig) jobResult {
	start := time.Now()
	result := jobResult{
		Variant:       job.Variant,
		Scenario:      job.Scenario.ID,
		ScenarioTitle: job.Scenario.Title,
		Status:        "failed",
		StartedAt:     start.UTC(),
		PromptSummary: promptSummary(job.Scenario),
	}
	timings := phaseTimings{}
	jobDir := filepath.Join(config.RunRoot, job.Variant, job.Scenario.ID)
	repoDir := filepath.Join(jobDir, "repo")
	paths := scenarioPaths(repoDir)
	fixtures := startSourceURLUpdateFixtures(job.Scenario.ID)
	if fixtures != nil {
		defer fixtures.Close()
	}
	if err := timedPhase(&timings.PrepareRunDir, func() error { return prepareRunDir(jobDir, cache) }); err != nil {
		result.Error = err.Error()
		return result
	}
	if fixtures != nil {
		if err := fixtures.prepareFiles(jobDir); err != nil {
			result.Error = fmt.Sprintf("prepare fixture files: %v", err)
			return result
		}
	}
	if err := timedPhase(&timings.CopyRepo, func() error { return copyRepo(config.RepoRoot, repoDir) }); err != nil {
		result.Error = fmt.Sprintf("copy repo: %v", err)
		return result
	}
	if err := timedPhase(&timings.InstallVariant, func() error {
		if err := installVariant(config.RepoRoot, repoDir, job.Variant); err != nil {
			return err
		}
		if err := buildOpenClerkRunner(repoDir, jobDir, paths, cache); err != nil {
			return err
		}
		return preflightEvalContext(config.RepoRoot, repoDir, jobDir, paths, cache, config.CodexBin)
	}); err != nil {
		result.Error = fmt.Sprintf("configure variant: %v", err)
		return result
	}
	if fixtures != nil && isArtifactPDFScenario(job.Scenario.ID) {
		preflight := runArtifactPDFFixturePreflight(ctx, jobDir, paths, cache, fixtures)
		result.FixturePreflight = &preflight
	}
	if cache.Mode == cacheModeIsolated {
		if err := timedPhase(&timings.WarmCache, func() error { return warmGoModules(repoDir, jobDir, paths, cache) }); err != nil {
			result.Error = fmt.Sprintf("warm go modules: %v", err)
			return result
		}
	}
	if err := timedPhase(&timings.SeedData, func() error { return seedScenarioWithFixtures(ctx, paths, job.Scenario, fixtures) }); err != nil {
		result.Error = fmt.Sprintf("seed scenario: %v", err)
		return result
	}
	if fixtures != nil {
		fixtures.prepareForAgent(job.Scenario.ID)
		if err := prepareSourceURLUpdateAgentState(ctx, paths, job.Scenario, fixtures); err != nil {
			result.Error = fmt.Sprintf("prepare source URL update state: %v", err)
			return result
		}
	}

	turns := scenarioTurns(job.Scenario)
	turnResults := make([]turnResult, 0, len(turns))
	sessionID := ""
	var runErr error
	for i, turn := range turns {
		turnIndex := i + 1
		turnResult, parsed, err := runScenarioTurn(ctx, config, repoDir, jobDir, paths, job, turn, turnIndex, sessionID, cache, fixtures)
		timings.AgentRun += turnResult.WallSeconds
		timings.ParseMetrics += parsed.parseSeconds
		if parsed.parseError != nil {
			turnResult.Metrics.CommandMetricLimitations = fmt.Sprintf("failed to parse event log: %v", parsed.parseError)
		}
		verifyStart := time.Now()
		verification, verifyErr := verifyScenarioTurn(ctx, paths, job.Scenario, turnIndex, parsed.finalMessage, turnResult.Metrics)
		timings.Verify += roundSeconds(time.Since(verifyStart).Seconds())
		if verifyErr != nil {
			verification = verificationResult{Passed: false, Details: fmt.Sprintf("verification error: %v", verifyErr)}
		}
		turnResult.Verification = verification
		turnResults = append(turnResults, turnResult)
		if err != nil && runErr == nil {
			runErr = err
		}
		if verifyErr != nil && runErr == nil {
			runErr = verifyErr
		}
		if i == 0 && len(turns) > 1 {
			sessionID = parsed.sessionID
			if sessionID == "" && runErr == nil {
				runErr = errors.New("multi-turn first turn did not expose a thread id")
			}
		}
	}

	completed := time.Now().UTC()
	timings.Total = roundSeconds(time.Since(start).Seconds())
	verification := aggregateVerification(job.Scenario, turnResults)
	result.CompletedAt = &completed
	result.WallSeconds = roundSeconds(sumTurnWallSeconds(turnResults))
	result.PhaseTimings = timings.rounded()
	result.Metrics = aggregateMetrics(turnResults)
	result.Verification = verification
	result.Turns = turnResults
	result.ExitCode = aggregateExitCode(turnResults)
	if len(turnResults) > 0 {
		result.RawLogArtifactReference = turnResults[len(turnResults)-1].RawLogArtifactReference
	}
	result.Passed = runErr == nil && verification.Passed
	if result.Passed {
		result.Status = "completed"
	} else if runErr != nil {
		result.Error = runErr.Error()
	}
	_ = writeJSON(filepath.Join(jobDir, "run-summary.json"), result)
	return result
}

type evalPaths struct {
	DatabasePath string
	GoCache      string
	GoModCache   string
	CodexHome    string
	ZDotDir      string
	Temp         string
}

func scenarioPaths(repoDir string) evalPaths {
	return evalPaths{
		DatabasePath: filepath.Join(repoDir, ".openclerk-eval", "openclerk.db"),
	}
}

type sourceURLUpdateFixtures struct {
	server          *httptest.Server
	mu              sync.Mutex
	initialPDF      []byte
	changedPDF      []byte
	serveChangedPDF bool
	artifactPDF     bool
}

func startSourceURLUpdateFixtures(scenarioID string) *sourceURLUpdateFixtures {
	if !isSourceURLUpdateScenario(scenarioID) && !isArtifactPDFScenario(scenarioID) {
		return nil
	}
	fixtures := &sourceURLUpdateFixtures{
		initialPDF: minimalEvalPDF("Source URL Update Stable", "OpenClerk Eval", sourceURLUpdateInitialText),
		changedPDF: minimalEvalPDF("Source URL Update Changed", "OpenClerk Eval", sourceURLUpdateChangedText),
	}
	if isArtifactPDFScenario(scenarioID) {
		fixtures.initialPDF = minimalEvalPDF("Artifact PDF Source", "OpenClerk Eval", artifactPDFEvidenceText)
		fixtures.artifactPDF = true
		return fixtures
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/stable.pdf", func(w http.ResponseWriter, _ *http.Request) {
		fixtures.mu.Lock()
		changed := fixtures.serveChangedPDF
		fixtures.mu.Unlock()
		body := fixtures.initialPDF
		if changed {
			body = fixtures.changedPDF
		}
		servePDF(w, body)
	})
	fixtures.server = httptest.NewServer(mux)
	return fixtures
}

func (f *sourceURLUpdateFixtures) Close() {
	if f != nil && f.server != nil {
		f.server.Close()
	}
}

func (f *sourceURLUpdateFixtures) stableURL() string {
	if f.artifactPDF {
		return artifactPDFEvalSourceURL
	}
	return f.server.URL + "/stable.pdf"
}

func (f *sourceURLUpdateFixtures) changedURL() string {
	return f.stableURL()
}

func (f *sourceURLUpdateFixtures) prepareForAgent(scenarioID string) {
	if scenarioID != sourceURLUpdateChangedScenarioID {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.serveChangedPDF = true
}

func (f *sourceURLUpdateFixtures) renderPrompt(prompt string) string {
	if f == nil {
		return prompt
	}
	prompt = strings.ReplaceAll(prompt, sourceURLUpdateStableURLToken, f.stableURL())
	prompt = strings.ReplaceAll(prompt, sourceURLUpdateChangedURLToken, f.changedURL())
	return strings.ReplaceAll(prompt, artifactPDFSourceURLToken, f.stableURL())
}

func (f *sourceURLUpdateFixtures) prepareFiles(runDir string) error {
	if f == nil || !f.artifactPDF {
		return nil
	}
	target := filepath.Join(evalSourceFixtureRoot(runDir), "artifacts", "vendor-security-paper.pdf")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, f.initialPDF, 0o644)
}

func evalSourceFixtureRoot(runDir string) string {
	return filepath.Join(runDir, "source-fixtures")
}

func runArtifactPDFFixturePreflight(ctx context.Context, runDir string, paths evalPaths, cache cacheConfig, fixtures *sourceURLUpdateFixtures) fixturePreflight {
	const sourcePath = "sources/artifacts/preflight-vendor-security-paper.md"
	const assetPath = "assets/sources/artifacts/preflight-vendor-security-paper.pdf"
	result := fixturePreflight{
		Name:       "artifact_pdf_source_url_fixture",
		Documents:  []string{sourcePath},
		SourcePath: sourcePath,
		AssetPath:  assetPath,
	}
	if fixtures == nil {
		result.Details = "missing PDF fixture server"
		return result
	}
	preflightRunDir := filepath.Join(runDir, "fixture-preflight")
	preflightPaths := paths
	preflightPaths.DatabasePath = filepath.Join(preflightRunDir, "openclerk-preflight.db")
	if err := os.MkdirAll(filepath.Join(preflightRunDir, "tmp"), 0o755); err != nil {
		result.Details = err.Error()
		return result
	}
	request := runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           fixtures.stableURL(),
			PathHint:      sourcePath,
			AssetPathHint: assetPath,
			Title:         "Vendor Security Paper Preflight",
		},
	}
	body, err := json.Marshal(request)
	if err != nil {
		result.Details = err.Error()
		return result
	}
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, filepath.Join(runDir, "bin", "openclerk"), "document")
	cmd.Dir = runDir
	cmd.Env = evalEnv(runDir, preflightPaths, cache)
	cmd.Stdin = bytes.NewReader(body)
	output, err := cmd.CombinedOutput()
	if cmdCtx.Err() == context.DeadlineExceeded {
		result.Details = "preflight timed out"
		return result
	}
	if err != nil {
		result.Details = fmt.Sprintf("%v: %s", err, strings.TrimSpace(string(output)))
		return result
	}
	var decoded runner.DocumentTaskResult
	if err := json.Unmarshal(output, &decoded); err != nil {
		result.Details = fmt.Sprintf("decode preflight result: %v", err)
		return result
	}
	if decoded.Rejected {
		result.Details = "preflight rejected: " + decoded.RejectionReason
		return result
	}
	if decoded.Ingestion == nil {
		result.Details = "preflight returned no ingestion result"
		return result
	}
	if decoded.Ingestion.SourcePath != sourcePath || decoded.Ingestion.AssetPath != assetPath || len(decoded.Ingestion.Citations) == 0 {
		result.Details = fmt.Sprintf("unexpected preflight ingestion source=%q asset=%q citations=%d", decoded.Ingestion.SourcePath, decoded.Ingestion.AssetPath, len(decoded.Ingestion.Citations))
		return result
	}
	result.Passed = true
	result.Details = "generated HTTP PDF ingested through built openclerk binary"
	return result
}

func servePDF(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/pdf")
	_, _ = w.Write(body)
}

func minimalEvalPDF(title string, author string, text string) []byte {
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

func evalPathsFor(runDir string, paths evalPaths, cache cacheConfig) evalPaths {
	out := paths
	out.CodexHome = filepath.Join(runDir, "codex-home")
	out.ZDotDir = filepath.Join(runDir, "zdotdir")
	out.Temp = filepath.Join(runDir, "tmp")
	if cache.Mode == cacheModeShared {
		out.GoCache = filepath.Join(cache.RunRoot, "shared-cache", "gocache")
		out.GoModCache = filepath.Join(cache.RunRoot, "shared-cache", "gomodcache")
	} else {
		out.GoCache = filepath.Join(runDir, "gocache")
		out.GoModCache = filepath.Join(runDir, "gomodcache")
	}
	return out
}

func runScenarioTurn(ctx context.Context, config runConfig, repoDir string, runDir string, paths evalPaths, job evalJob, turn scenarioTurn, turnIndex int, sessionID string, cache cacheConfig, fixtures *sourceURLUpdateFixtures) (turnResult, parsedTurn, error) {
	turnDir := filepath.Join(runDir, fmt.Sprintf("turn-%d", turnIndex))
	if err := os.MkdirAll(turnDir, 0o755); err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	eventsPath := filepath.Join(turnDir, "events.jsonl")
	stderrPath := filepath.Join(turnDir, "stderr.log")
	stdoutFile, err := os.Create(eventsPath)
	if err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	defer func() { _ = stdoutFile.Close() }()
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	defer func() { _ = stderrFile.Close() }()

	if fixtures != nil {
		turn.Prompt = fixtures.renderPrompt(turn.Prompt)
	}
	args := codexArgsForTurn(config.CodexBin, repoDir, runDir, job.Scenario, turn, turnIndex, sessionID, cache)
	cmdCtx, cancel := context.WithTimeout(ctx, 7*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, args[0], args[1:]...)
	cmd.Dir = repoDir
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	cmd.Stdin = strings.NewReader("")
	cmd.Env = evalEnv(runDir, paths, cache)

	start := time.Now()
	err = cmd.Run()
	wallSeconds := roundSeconds(time.Since(start).Seconds())
	exitCode := commandExitCode(err)
	if cmdCtx.Err() == context.DeadlineExceeded {
		exitCode = -1
		err = cmdCtx.Err()
	}
	parseStart := time.Now()
	parsedMetrics, parseErr := parseMetrics(eventsPath)
	parseSeconds := roundSeconds(time.Since(parseStart).Seconds())
	parsed := parsedTurn{
		metrics:      parsedMetrics.metrics,
		finalMessage: parsedMetrics.finalMessage,
		sessionID:    parsedMetrics.sessionID,
		parseError:   parseErr,
		parseSeconds: parseSeconds,
	}
	result := turnResult{
		Index:                   turnIndex,
		WallSeconds:             wallSeconds,
		ExitCode:                exitCode,
		Metrics:                 parsedMetrics.metrics,
		RawLogArtifactReference: fmt.Sprintf("<run-root>/%s/%s/turn-%d/events.jsonl", job.Variant, job.Scenario.ID, turnIndex),
	}
	return result, parsed, err
}

func codexArgsForTurn(codexBin string, repoDir string, runDir string, sc scenario, turn scenarioTurn, turnIndex int, sessionID string, cache cacheConfig) []string {
	baseConfig := []string{
		"-m", modelName,
		"-c", fmt.Sprintf("model_reasoning_effort=%q", reasoningEffort),
		"-c", "shell_environment_policy.inherit=all",
	}
	writableRoots := codexWritableRoots(runDir, cache)
	if len(scenarioTurns(sc)) == 1 {
		args := []string{codexBin, "exec", "--json", "--ephemeral", "--full-auto", "--skip-git-repo-check", "--ignore-user-config", "-C", repoDir}
		args = appendAddDirs(args, writableRoots)
		args = append(args, baseConfig...)
		return append(args, turn.Prompt)
	}
	if turnIndex == 1 {
		args := []string{codexBin, "exec", "--json", "--full-auto", "--skip-git-repo-check", "--ignore-user-config", "-C", repoDir}
		args = appendAddDirs(args, writableRoots)
		args = append(args, baseConfig...)
		return append(args, turn.Prompt)
	}
	args := []string{codexBin, "exec", "-C", repoDir}
	args = appendAddDirs(args, writableRoots)
	args = append(args, "resume", "--json", "--full-auto", "--skip-git-repo-check", "--ignore-user-config")
	args = append(args, baseConfig...)
	args = append(args, sessionID, turn.Prompt)
	return args
}

func codexWritableRoots(runDir string, cache cacheConfig) []string {
	roots := []string{runDir}
	if cache.Mode == cacheModeShared {
		roots = append(roots, filepath.Join(cache.RunRoot, "shared-cache"))
	}
	return roots
}

func appendAddDirs(args []string, roots []string) []string {
	for _, root := range roots {
		args = append(args, "--add-dir", root)
	}
	return args
}

func evalEnv(runDir string, paths evalPaths, cache cacheConfig) []string {
	effective := evalPathsFor(runDir, paths, cache)
	env := filteredEnv(os.Environ(),
		"CODEX_HOME",
		"OPENCLERK_DATA_DIR",
		"OPENCLERK_DATABASE_PATH",
		evalSourceFixtureRootEnv,
		"OPENCLERK_VAULT_ROOT",
		"GOCACHE",
		"GOMODCACHE",
		"TMPDIR",
		"PATH",
		"ZDOTDIR",
	)
	pathValue := filepath.Join(runDir, "bin")
	if existing := os.Getenv("PATH"); existing != "" {
		pathValue += string(os.PathListSeparator) + existing
	}
	env = append(env,
		"CODEX_HOME="+effective.CodexHome,
		"ZDOTDIR="+effective.ZDotDir,
		"OPENCLERK_DATABASE_PATH="+effective.DatabasePath,
		evalSourceFixtureRootEnv+"="+evalSourceFixtureRoot(runDir),
		"GOCACHE="+effective.GoCache,
		"GOMODCACHE="+effective.GoModCache,
		"TMPDIR="+effective.Temp,
		"PATH="+pathValue,
	)
	return env
}

func filteredEnv(env []string, keys ...string) []string {
	if len(keys) == 0 {
		return append([]string{}, env...)
	}
	blocked := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		blocked[key] = struct{}{}
	}
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, found := strings.Cut(entry, "=")
		if found {
			if _, blockedKey := blocked[key]; blockedKey {
				continue
			}
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func prepareRunDir(runDir string, cache cacheConfig) error {
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	paths := evalPathsFor(runDir, evalPaths{}, cache)
	for _, dir := range []string{paths.ZDotDir, paths.Temp} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := setupEvalCodexHome(paths.CodexHome); err != nil {
		return err
	}
	return nil
}

func setupEvalCodexHome(dst string) error {
	srcRoot, err := sourceCodexHome()
	if err != nil {
		return err
	}
	return setupEvalCodexHomeFromSource(dst, srcRoot)
}

func sourceCodexHome() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("CODEX_HOME")); configured != "" {
		return configured, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

func setupEvalCodexHomeFromSource(dst string, sourceHome string) error {
	authBytes, err := os.ReadFile(filepath.Join(sourceHome, "auth.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("missing Codex auth at %s; run codex login before running evals", filepath.Join(sourceHome, "auth.json"))
		}
		return err
	}
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, "auth.json"), authBytes, 0o600)
}

func warmGoModules(repoDir string, runDir string, paths evalPaths, cache cacheConfig) error {
	effective := evalPathsFor(runDir, paths, cache)
	for _, dir := range []string{effective.GoCache, effective.GoModCache, effective.Temp} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = repoDir
	cmd.Env = evalEnv(runDir, paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func prewarmSharedCache(repoRoot string, cache cacheConfig) error {
	paths := evalPathsFor(filepath.Join(cache.RunRoot, "shared-cache"), evalPaths{
		DatabasePath: filepath.Join(cache.RunRoot, "shared-cache", "prewarm.db"),
	}, cache)
	for _, dir := range []string{paths.GoCache, paths.GoModCache, paths.Temp} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := warmGoModules(repoRoot, filepath.Join(cache.RunRoot, "shared-cache"), paths, cache); err != nil {
		return err
	}
	cmd := exec.Command("go", prewarmCompileArgs()...)
	cmd.Dir = repoRoot
	cmd.Env = evalEnv(filepath.Join(cache.RunRoot, "shared-cache"), paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func prewarmCompileArgs() []string {
	args := []string{"test", "-run", "^$"}
	return append(args, prewarmCompilePackages...)
}

func buildOpenClerkRunner(repoDir string, runDir string, paths evalPaths, cache cacheConfig) error {
	binDir := filepath.Join(runDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(binDir, "openclerk"), "./cmd/openclerk")
	cmd.Dir = repoDir
	cmd.Env = evalEnv(runDir, paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func seedScenario(ctx context.Context, paths evalPaths, sc scenario) error {
	return seedScenarioWithFixtures(ctx, paths, sc, nil)
}

func seedScenarioWithFixtures(ctx context.Context, paths evalPaths, sc scenario, fixtures *sourceURLUpdateFixtures) error {
	cfg := runclient.Config{
		DatabasePath: paths.DatabasePath,
	}
	switch sc.ID {
	case "search-synthesis", "mt-source-then-synthesis":
		if err := createSeedDocument(ctx, cfg, "sources/openclerk-runner.md", "OpenClerk Runner Source", "The OpenClerk runner uses JSON requests for OpenClerk knowledge tasks.\n\nIt preserves source refs for synthesis pages."); err != nil {
			return err
		}
	case "answer-filing":
		if err := createSeedDocument(ctx, cfg, "sources/answer-filing-runner.md", "OpenClerk runner Answer Filing Source", "The OpenClerk runner JSON runner is the production path for reusable OpenClerk knowledge tasks.\n\nDurable OpenClerk runner answers should be filed as source-linked markdown."); err != nil {
			return err
		}
	case ragRetrievalScenarioID:
		if err := seedRAGRetrievalBaseline(ctx, cfg); err != nil {
			return err
		}
	case docsNavigationScenarioID:
		if err := seedDocsNavigationBaseline(ctx, cfg); err != nil {
			return err
		}
	case graphSemanticsScenarioID:
		if err := seedGraphSemanticsReference(ctx, cfg); err != nil {
			return err
		}
	case memoryRouterScenarioID:
		if err := seedMemoryRouterReference(ctx, cfg); err != nil {
			return err
		}
	case configuredLayoutScenarioID:
		if err := seedConfiguredLayoutScenario(ctx, cfg); err != nil {
			return err
		}
	case invalidLayoutScenarioID:
		if err := seedInvalidLayoutScenario(ctx, cfg); err != nil {
			return err
		}
	case sourceURLUpdateDuplicateScenarioID, sourceURLUpdateConflictScenarioID:
		if fixtures == nil {
			return errors.New("source URL update fixture server is required")
		}
		if err := seedSourceURLUpdateSource(ctx, cfg, fixtures.stableURL()); err != nil {
			return err
		}
	case sourceURLUpdateSameSHAScenarioID:
		if fixtures == nil {
			return errors.New("source URL update fixture server is required")
		}
		if err := seedSourceURLUpdateSource(ctx, cfg, fixtures.stableURL()); err != nil {
			return err
		}
		if err := seedSourceURLUpdateSynthesis(ctx, cfg); err != nil {
			return err
		}
	case sourceURLUpdateChangedScenarioID:
		if fixtures == nil {
			return errors.New("source URL update fixture server is required")
		}
		if err := seedSourceURLUpdateSource(ctx, cfg, fixtures.changedURL()); err != nil {
			return err
		}
		if err := seedSourceURLUpdateSynthesis(ctx, cfg); err != nil {
			return err
		}
	case synthesisCandidatePressureScenarioID:
		if err := seedSynthesisCandidatePressure(ctx, cfg); err != nil {
			return err
		}
	case synthesisSourceSetPressureScenarioID:
		if err := seedSynthesisSourceSetPressure(ctx, cfg); err != nil {
			return err
		}
	case decisionRecordVsDocsScenarioID:
		if err := seedDecisionRecordVsDocs(ctx, cfg); err != nil {
			return err
		}
	case decisionSupersessionScenarioID:
		if err := seedDecisionSupersession(ctx, cfg); err != nil {
			return err
		}
	case decisionRealADRMigrationScenarioID:
		if err := seedDecisionRealADRMigration(ctx, cfg); err != nil {
			return err
		}
	case sourceAuditRepairScenarioID:
		if err := seedSourceSensitiveAuditRepair(ctx, cfg); err != nil {
			return err
		}
	case sourceAuditConflictScenarioID:
		if err := seedSourceSensitiveConflict(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryInspectScenarioID:
		if err := seedDocumentHistoryInspection(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryDiffScenarioID:
		if err := seedDocumentHistoryDiffReview(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryRestoreScenarioID:
		if err := seedDocumentHistoryRestore(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryPendingScenarioID:
		if err := seedDocumentHistoryPendingReview(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryStaleScenarioID:
		if err := seedDocumentHistoryStaleSynthesis(ctx, cfg); err != nil {
			return err
		}
	case mtSynthesisDriftPressureScenarioID:
		if err := seedMTSynthesisDriftPressure(ctx, cfg); err != nil {
			return err
		}
	case populatedHeterogeneousScenarioID, populatedFreshnessConflictScenarioID, populatedSynthesisUpdateScenarioID:
		if err := seedPopulatedVaultFixture(ctx, cfg); err != nil {
			return err
		}
	case repoDocsRetrievalScenarioID, repoDocsSynthesisScenarioID, repoDocsDecisionScenarioID:
		if err := seedRepoDocsDogfood(ctx, cfg); err != nil {
			return err
		}
	case agentChosenSynthesisScenarioID:
		if err := seedAgentChosenSynthesisPathSelection(ctx, cfg); err != nil {
			return err
		}
	case pathTitleMultiSourceDuplicateScenarioID:
		if err := seedPathTitleMultiSourceDuplicatePressure(ctx, cfg); err != nil {
			return err
		}
	case pathTitleDuplicateRiskScenarioID:
		if err := seedPathTitleDuplicateRiskPressure(ctx, cfg); err != nil {
			return err
		}
	case documentThisDuplicateCandidateScenarioID:
		if err := seedDocumentThisDuplicateCandidate(ctx, cfg); err != nil {
			return err
		}
	case documentThisExistingUpdateScenarioID:
		if err := seedDocumentThisExistingUpdate(ctx, cfg); err != nil {
			return err
		}
	case documentThisSynthesisFreshnessScenarioID:
		if err := seedDocumentThisSynthesisFreshness(ctx, cfg); err != nil {
			return err
		}
	case candidateDuplicateRiskAsksScenarioID, candidateErgonomicsDuplicateNaturalID:
		if err := seedDocumentArtifactCandidateDuplicate(ctx, cfg); err != nil {
			return err
		}
	case artifactTranscriptScenarioID:
		if err := seedArtifactTranscript(ctx, cfg); err != nil {
			return err
		}
	case artifactInvoiceReceiptScenarioID:
		if err := seedArtifactInvoiceReceipt(ctx, cfg); err != nil {
			return err
		}
	case artifactMixedSynthesisScenarioID:
		if err := seedArtifactMixedSynthesis(ctx, cfg); err != nil {
			return err
		}
	case videoYouTubeSynthesisFreshnessScenarioID:
		if err := seedVideoYouTubeSynthesisFreshness(ctx, cfg); err != nil {
			return err
		}
	case "stale-synthesis-update":
		if err := createSeedDocument(ctx, cfg, "sources/runner-old-workaround.md", "Old OpenClerk runner Routing Source", "Older guidance said routine agents may bypass OpenClerk runner through a temporary command-path workaround."); err != nil {
			return err
		}
		if err := createSeedDocument(ctx, cfg, "sources/runner-current-runner.md", "Current OpenClerk runner Routing Source", "Current guidance says routine agents must use openclerk JSON runner for OpenClerk knowledge tasks."); err != nil {
			return err
		}
		body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner-current-runner.md, sources/runner-old-workaround.md
---

# OpenClerk runner Routing

## Summary
Stale claim: routine agents may bypass OpenClerk runner through a temporary command-path workaround.

## Sources
- sources/runner-current-runner.md
- sources/runner-old-workaround.md

## Freshness
Checked source: sources/runner-old-workaround.md
`)
		if err := createSeedDocument(ctx, cfg, "synthesis/runner-routing.md", "OpenClerk runner Routing", body); err != nil {
			return err
		}
	case "synthesis-freshness-repair":
		oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: sources/repair-current.md
---
# Old OpenClerk runner Repair Source

## Summary
Older repair guidance mentioned a temporary command-path workaround.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "sources/repair-old.md", "Old OpenClerk runner Repair Source", oldBody); err != nil {
			return err
		}
		currentBody := strings.TrimSpace(`---
supersedes: sources/repair-old.md
---
# Current OpenClerk runner Repair Source

## Summary
Current guidance says routine agents must use openclerk JSON runner for freshness repairs.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "sources/repair-current.md", "Current OpenClerk runner Repair Source", currentBody); err != nil {
			return err
		}
		synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/repair-current.md, sources/repair-old.md
---
# OpenClerk runner Freshness Repair

## Summary
Stale repair claim: routine agents may use a temporary command-path workaround.

## Sources
- sources/repair-current.md
- sources/repair-old.md

## Freshness
Checked before the latest source update.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "synthesis/runner-repair.md", "OpenClerk runner Freshness Repair", synthesisBody); err != nil {
			return err
		}
		if err := replaceScenarioSeedSection(ctx, cfg, "sources/repair-current.md", "Summary", "Current guidance says routine agents must use openclerk JSON runner for freshness repairs, and sources/repair-old.md is superseded."); err != nil {
			return err
		}
	case "append-replace":
		if err := createSeedDocument(ctx, cfg, "notes/projects/openclerk-runner.md", "OpenClerk Runner", "## Context\nExisting context stays intact."); err != nil {
			return err
		}
	case "records-provenance":
		if err := createSeedDocument(ctx, cfg, "records/services/openclerk-runner.md", "OpenClerk runner", recordBody("openclerk-runner", "service", "OpenClerk runner")); err != nil {
			return err
		}
	case "mixed-synthesis-records":
		if err := createSeedDocument(ctx, cfg, "sources/openclerk-runner.md", "OpenClerk Runner Source", "The OpenClerk runner uses JSON requests for OpenClerk knowledge tasks.\n\nIt preserves source refs for synthesis pages."); err != nil {
			return err
		}
		if err := createSeedDocument(ctx, cfg, "records/services/openclerk-runner.md", "OpenClerk runner", recordBody("openclerk-runner", "service", "OpenClerk runner")); err != nil {
			return err
		}
	case "promoted-record-vs-docs":
		if err := createSeedDocument(ctx, cfg, "notes/reference/runner-service.md", "OpenClerk runner Service Reference", "# OpenClerk runner Service Reference\n\n## Summary\nPlain docs evidence says OpenClerk runner is the production service for routine knowledge tasks.\n\n## Details\nPlain docs evidence is narrative and searchable.\n"); err != nil {
			return err
		}
		body := strings.TrimSpace(`---
service_id: openclerk-runner
service_name: OpenClerk runner
service_status: active
service_owner: runner
service_interface: JSON runner
---

# OpenClerk runner

## Facts
- production_path: true
`)
		if err := createSeedDocument(ctx, cfg, "records/services/openclerk-runner.md", "OpenClerk runner", body); err != nil {
			return err
		}
	case "duplicate-path-reject":
		if err := createSeedDocument(ctx, cfg, "notes/projects/duplicate.md", "Duplicate Source", "This canonical path already exists."); err != nil {
			return err
		}
	}
	return nil
}

func seedRAGRetrievalBaseline(ctx context.Context, cfg runclient.Config) error {
	currentBody := strings.TrimSpace(`---
type: note
status: active
rag_scope: active-policy
---
# Current AgentOps RAG Policy

## Summary
`+ragCurrentPolicySummary+`

## Decision
`+ragCurrentPolicyDecision+`
`) + "\n"
	if err := createSeedDocument(ctx, cfg, ragCurrentPolicyPath, ragCurrentPolicyTitle, currentBody); err != nil {
		return err
	}
	decoyBody := strings.TrimSpace(`---
type: note
status: draft
rag_scope: decoy-policy
---
# Decoy AgentOps RAG Policy

## Summary
Decoy AgentOps RAG baseline policy marker: this draft says direct SQLite might be acceptable for routine OpenClerk knowledge answers.

## Decision
This is a decoy policy and is not the active AgentOps retrieval decision.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, ragDecoyPolicyPath, ragDecoyPolicyTitle, decoyBody); err != nil {
		return err
	}
	archivedBody := strings.TrimSpace(`---
type: note
status: superseded
rag_scope: archived-policy
---
# Archived AgentOps RAG Policy

## Summary
Archived AgentOps RAG baseline policy marker: older guidance mentioned a source-built command path.

## Decision
This archived policy is outside the active RAG path prefix and is superseded by the current JSON runner policy.
`) + "\n"
	return createSeedDocument(ctx, cfg, ragArchivedPolicyPath, ragArchivedPolicyTitle, archivedBody)
}

func seedRepoDocsDogfood(ctx context.Context, cfg runclient.Config) error {
	repoRoot, err := repoRootFromEvalDatabasePath(cfg.DatabasePath)
	if err != nil {
		return err
	}
	var imported int
	err = filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if !shouldDescendRepoMarkdownDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if !shouldImportRepoMarkdown(rel, entry) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		repoPath := filepath.ToSlash(rel)
		if err := createSeedDocument(ctx, cfg, repoPath, markdownTitle(repoPath, string(content)), string(content)); err != nil {
			return fmt.Errorf("import repo doc %s: %w", repoPath, err)
		}
		imported++
		return nil
	})
	if err != nil {
		return err
	}
	if imported == 0 {
		return errors.New("repo-docs dogfood seed imported no markdown documents")
	}
	return nil
}

func shouldDescendRepoMarkdownDir(rel string) bool {
	slash := filepath.ToSlash(rel)
	if slash == "." {
		return true
	}
	parts := strings.Split(slash, "/")
	if len(parts) == 0 {
		return true
	}
	switch parts[0] {
	case ".git", ".beads", ".dolt", ".agents", ".openclerk-eval":
		return false
	}
	return !strings.HasPrefix(slash, "docs/evals/results/")
}

func repoRootFromEvalDatabasePath(databasePath string) (string, error) {
	if strings.TrimSpace(databasePath) == "" {
		return "", errors.New("missing eval database path")
	}
	evalDir := filepath.Dir(databasePath)
	if filepath.Base(evalDir) != ".openclerk-eval" {
		return "", fmt.Errorf("database path %q is not under .openclerk-eval", databasePath)
	}
	return filepath.Dir(evalDir), nil
}

func shouldImportRepoMarkdown(rel string, entry fs.DirEntry) bool {
	slash := filepath.ToSlash(rel)
	if slash == "." {
		return false
	}
	parts := strings.Split(slash, "/")
	if len(parts) > 0 {
		switch parts[0] {
		case ".git", ".beads", ".dolt", ".agents", ".openclerk-eval":
			return false
		case "AGENTS.md":
			return false
		}
	}
	if strings.HasPrefix(slash, "docs/evals/results/") {
		return false
	}
	return !entry.IsDir() && strings.EqualFold(filepath.Ext(slash), ".md")
}

func markdownTitle(path string, body string) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			if title != "" {
				return title
			}
		}
	}
	base := filepath.Base(path)
	title := strings.TrimSuffix(base, filepath.Ext(base))
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.TrimSpace(title)
	if title == "" {
		return path
	}
	return title
}

func seedDocsNavigationBaseline(ctx context.Context, cfg runclient.Config) error {
	indexBody := strings.TrimSpace(`---
type: wiki
status: active
---
# AgentOps Wiki Index

## Summary
Canonical directory navigation starts here for the AgentOps wiki baseline.

## Links
- [Runner policy](runner-policy.md)
- [Knowledge plane](../architecture/knowledge-plane.md)
- [Runner playbook](../ops/runner-playbook.md)

## Limits
Folder paths and headings show the local index, but they do not explain backlinks or cross-directory relationship neighborhoods without retrieval actions.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationIndexPath, "AgentOps Wiki Index", indexBody); err != nil {
		return err
	}

	policyBody := strings.TrimSpace(`---
type: policy
status: active
---
# Runner Policy

## Summary
Routine OpenClerk knowledge work uses the installed JSON runner and cites returned source paths.

## Navigation
Return to the [AgentOps wiki index](index.md) and compare with the [knowledge plane](../architecture/knowledge-plane.md).
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationPolicyPath, "Runner Policy", policyBody); err != nil {
		return err
	}

	architectureBody := strings.TrimSpace(`---
type: architecture
status: active
---
# Knowledge Plane

## Summary
The knowledge plane keeps canonical markdown as source authority and derives graph relationships from links.

## Navigation
The [AgentOps wiki index](../agentops/index.md) links this architecture note to runner policy context.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationArchPath, "Knowledge Plane", architectureBody); err != nil {
		return err
	}

	opsBody := strings.TrimSpace(`---
type: runbook
status: active
---
# Runner Playbook

## Summary
Operators use the runner playbook when directory navigation is not enough to explain related policy and architecture docs.

## Navigation
Start from the [AgentOps wiki index](../agentops/index.md) before following graph neighborhoods.
`) + "\n"
	return createSeedDocument(ctx, cfg, docsNavigationOpsPath, "Runner Playbook", opsBody)
}

func seedGraphSemanticsReference(ctx context.Context, cfg runclient.Config) error {
	indexBody := strings.TrimSpace(`---
type: graph-reference
status: active
---
# Graph Semantics Reference

## Summary
Graph semantics requires canonical markdown to carry relationship meaning. This reference note says the routing note supersedes legacy graph claims, is related to freshness evidence, and operationalizes the operations playbook.

## Relationships
- Requires: [Routing](routing.md)
- Supersedes: [Freshness](freshness.md)
- Related to: [Operations](operations.md)
- Operationalizes: Operations playbook

## Decision
Richer graph semantics stay in canonical markdown relationship text. The derived graph should expose structural links and citations, not independent semantic-label authority.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, graphSemanticsIndexPath, "Graph Semantics Reference", indexBody); err != nil {
		return err
	}

	routingBody := strings.TrimSpace(`---
type: graph-reference
status: active
---
# Routing

## Summary
Routing links back to the [Graph Semantics Reference](index.md) because semantic relationship labels should remain inspectable markdown evidence.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, graphSemanticsRoutingPath, "Routing", routingBody); err != nil {
		return err
	}

	freshnessBody := strings.TrimSpace(`---
type: graph-reference
status: active
---
# Freshness

## Summary
Freshness links back to the [Graph Semantics Reference](index.md) so graph projection freshness stays tied to canonical markdown.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, graphSemanticsFreshnessPath, "Freshness", freshnessBody); err != nil {
		return err
	}

	operationsBody := strings.TrimSpace(`---
type: graph-reference
status: active
---
# Operations

## Summary
Operations links back to the [Graph Semantics Reference](index.md) and keeps operationalizes language in source text rather than in opaque graph labels.
`) + "\n"
	return createSeedDocument(ctx, cfg, graphSemanticsOperationsPath, "Operations", operationsBody)
}

func seedMemoryRouterReference(ctx context.Context, cfg runclient.Config) error {
	temporalBody := strings.TrimSpace(`---
type: memory-router-reference
status: active
effective_at: 2026-04-22
---
# Temporal Recall Policy

## Summary
Temporal recall stays source-grounded: current canonical docs and promoted records outrank stale session observations, and agents must name the temporal status before trusting a result.

## Guidance
Current evidence should be described as current or effective. Older or superseded evidence should be described as stale before it is reused.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, memoryRouterTemporalPath, "Temporal Recall Policy", temporalBody); err != nil {
		return err
	}

	feedbackBody := strings.TrimSpace(`---
type: memory-router-reference
status: active
---
# Feedback Weighting

## Summary
Feedback weighting is advisory only. A high-weight remembered result can help rank what to inspect next, but it cannot hide source refs, freshness, provenance, or weaker conflicting evidence.

## Guidance
The reference weight for the session observation is 0.8 because the user marked it useful, but the answer must still cite canonical markdown.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, memoryRouterFeedbackPath, "Feedback Weighting", feedbackBody); err != nil {
		return err
	}

	routingBody := strings.TrimSpace(`---
type: memory-router-reference
status: active
---
# Routing Policy

## Summary
Routing is an explainable choice among existing AgentOps document and retrieval actions. Use canonical docs and provenance for source-sensitive claims, promoted records for typed domains, graph navigation for relationship questions, and never use autonomous routing as hidden authority.

## Guidance
The correct route for this reference POC is canonical docs plus provenance and projection freshness, not a memory-first router.
`) + "\n"
	return createSeedDocument(ctx, cfg, memoryRouterRoutingPath, "Routing Policy", routingBody)
}

func memoryRouterSessionObservationBody() string {
	return strings.TrimSpace(`---
type: source
status: active
observed_at: 2026-04-22
---
# Memory Router Session Observation

## Summary
Session observation: a user asked whether memory routing should promote recall. Useful session material must be promoted only by writing canonical markdown with source refs.

## Feedback
Positive feedback weight 0.8 is advisory only and cannot hide stale canonical evidence.
`) + "\n"
}

func seedConfiguredLayoutScenario(ctx context.Context, cfg runclient.Config) error {
	sourceBody := strings.TrimSpace(`---
type: source
status: active
---
# Layout Runner Source

## Summary
Convention-first OpenClerk knowledge layout uses runner-visible JSON inspection rather than a committed manifest.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "sources/layout-runner.md", "Layout Runner Source", sourceBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/layout-runner.md
---
# Layout Runner Synthesis

## Summary
The configured layout keeps canonical markdown and source-linked synthesis convention-first.

## Sources
- sources/layout-runner.md

## Freshness
Checked source refs through runner-visible layout inspection.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "synthesis/layout-runner.md", "Layout Runner Synthesis", synthesisBody); err != nil {
		return err
	}
	recordBody := strings.TrimSpace(`---
entity_id: layout-runner-record
entity_type: policy
entity_name: Layout Runner Policy
---
# Layout Runner Policy

## Facts
- status: active
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "records/policies/layout-runner.md", "Layout Runner Policy", recordBody); err != nil {
		return err
	}
	serviceBody := strings.TrimSpace(`---
service_id: layout-runner
service_name: Layout Runner
service_status: active
service_owner: runner
service_interface: JSON runner
---
# Layout Runner

## Summary
Runner-visible layout inspection explains configured knowledge conventions.
`) + "\n"
	return createSeedDocument(ctx, cfg, "records/services/layout-runner.md", "Layout Runner", serviceBody)
}

func seedInvalidLayoutScenario(ctx context.Context, cfg runclient.Config) error {
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
source_refs: sources/missing-layout-source.md
---
# Broken Layout Synthesis

## Summary
This synthesis references a missing source and omits the required freshness section.

## Sources
- sources/missing-layout-source.md
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "synthesis/broken-layout.md", "Broken Layout Synthesis", synthesisBody); err != nil {
		return err
	}
	serviceBody := strings.TrimSpace(`---
service_id: broken-layout-service
---
# Broken Layout Service

## Summary
This service-shaped document is missing service_name.
`) + "\n"
	return createSeedDocument(ctx, cfg, "records/services/broken-layout-service.md", "Broken Layout Service", serviceBody)
}

func seedSynthesisCandidatePressure(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: sources/compiler-current.md
---
# Compiler Old Source

## Summary
Older compiler guidance said routine synthesis repairs need a dedicated compile_synthesis action.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCandidateOldSrc, "Compiler Old Source", oldBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/compiler-current.md, sources/compiler-old.md
---
# Compiler Routing

## Summary
Stale compiler claim: routine synthesis repairs require a dedicated compile_synthesis runner action.

## Sources
- sources/compiler-current.md
- sources/compiler-old.md

## Freshness
Checked before the latest compiler pressure source was registered.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCandidatePath, "Compiler Routing", synthesisBody); err != nil {
		return err
	}
	decoyBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/compiler-old.md
---
# Compiler Routing Decoy

## Summary
This decoy page is not the compiler pressure decision target.

## Sources
- sources/compiler-old.md

## Freshness
Checked decoy source only.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCandidateDecoyPath, "Compiler Routing Decoy", decoyBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
supersedes: sources/compiler-old.md
---
# Compiler Current Source

## Summary
Current compiler pressure guidance says existing document and retrieval actions are sufficient when agents search sources, list synthesis candidates, inspect freshness, and update without duplicates.
`) + "\n"
	return createSeedDocument(ctx, cfg, synthesisCandidateCurrentSrc, "Compiler Current Source", currentBody)
}

func seedSynthesisSourceSetPressure(ctx context.Context, cfg runclient.Config) error {
	sourceBodies := map[string]string{
		sourceSetAlphaPath: strings.TrimSpace(`---
type: source
status: active
source_set: compiler-pressure
---
# Source Set Alpha

## Summary
Alpha source says synthesis compiler pressure requires source search before durable synthesis.
`) + "\n",
		sourceSetBetaPath: strings.TrimSpace(`---
type: source
status: active
source_set: compiler-pressure
---
# Source Set Beta

## Summary
Beta source says synthesis compiler pressure requires listing existing synthesis candidates.
`) + "\n",
		sourceSetGammaPath: strings.TrimSpace(`---
type: source
status: active
source_set: compiler-pressure
---
# Source Set Gamma

## Summary
Gamma source says synthesis compiler pressure requires preserving freshness and source refs.
`) + "\n",
	}
	for _, path := range []string{sourceSetAlphaPath, sourceSetBetaPath, sourceSetGammaPath} {
		if err := createSeedDocument(ctx, cfg, path, sourceTitleFromPath(path), sourceBodies[path]); err != nil {
			return err
		}
	}
	return nil
}

func seedAgentChosenSynthesisPathSelection(ctx context.Context, cfg runclient.Config) error {
	sourceBodies := map[string]string{
		agentChosenSynthesisAlphaPath: strings.TrimSpace(`---
type: source
status: active
path_pressure: agent-chosen
---
# Path Alpha

## Summary
Alpha source says agent-chosen path selection must preserve explicit-path compatibility.
`) + "\n",
		agentChosenSynthesisBetaPath: strings.TrimSpace(`---
type: source
status: active
path_pressure: agent-chosen
---
# Path Beta

## Summary
Beta source says metadata remains authoritative for document type and identity.
`) + "\n",
		agentChosenSynthesisGammaPath: strings.TrimSpace(`---
type: source
status: active
path_pressure: agent-chosen
---
# Path Gamma

## Summary
Gamma source says freshness, source refs, and citations must remain inspectable.
`) + "\n",
	}
	for _, path := range []string{agentChosenSynthesisAlphaPath, agentChosenSynthesisBetaPath, agentChosenSynthesisGammaPath} {
		if err := createSeedDocument(ctx, cfg, path, sourceTitleFromPath(path), sourceBodies[path]); err != nil {
			return err
		}
	}
	return nil
}

func seedPathTitleMultiSourceDuplicatePressure(ctx context.Context, cfg runclient.Config) error {
	sourceBodies := map[string]string{
		pathTitleSynthesisAlphaPath: strings.TrimSpace(`---
type: source
status: active
path_title_pressure: multi-source
---
# Path Title Alpha

## Summary
Alpha source says constrained autonomy must search sources before choosing a durable synthesis path.
`) + "\n",
		pathTitleSynthesisBetaPath: strings.TrimSpace(`---
type: source
status: active
path_title_pressure: multi-source
---
# Path Title Beta

## Summary
Beta source says constrained autonomy must update existing synthesis candidates instead of creating duplicates.
`) + "\n",
	}
	for _, path := range []string{pathTitleSynthesisAlphaPath, pathTitleSynthesisBetaPath} {
		if err := createSeedDocument(ctx, cfg, path, sourceTitleFromPath(path), sourceBodies[path]); err != nil {
			return err
		}
	}
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/path-title/alpha.md, sources/path-title/beta.md
---
# Path Title Autonomy

## Summary
Existing synthesis candidate for path/title autonomy pressure.

## Sources
- sources/path-title/alpha.md
- sources/path-title/beta.md

## Freshness
Fresh before autonomy pressure checks.
`) + "\n"
	return createSeedDocument(ctx, cfg, pathTitleSynthesisPath, pathTitleSynthesisTitle, body)
}

func seedPathTitleDuplicateRiskPressure(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: source
status: active
---
# Existing OpenAI Harness

## Summary
Existing source note for the OpenAI harness URL. Duplicate risk marker: existing path/title source should be reused, not copied.
`) + "\n"
	return createSeedDocument(ctx, cfg, pathTitleDuplicateExistingPath, "Existing OpenAI Harness", body)
}

func seedDocumentThisDuplicateCandidate(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: source
status: active
---
# Existing Document This Article

## Summary
Document-this duplicate marker: the article source already captures strict runner intake guidance.

## Sources
- https://example.test/articles/document-this-intake
`) + "\n"
	return createSeedDocument(ctx, cfg, documentThisDuplicateExistingPath, "Existing Document This Article", body)
}

func seedDocumentThisExistingUpdate(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: note
status: active
---
# Existing Document This Update

## Summary
Existing update target for document-this intake pressure.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentThisUpdateTargetPath, "Existing Document This Update", body); err != nil {
		return err
	}
	decoy := strings.TrimSpace(`---
type: note
status: active
---
# Existing Document This Update Decoy

## Summary
Decoy note that must not receive the document-this update.
`) + "\n"
	return createSeedDocument(ctx, cfg, documentThisUpdateDecoyPath, "Existing Document This Update Decoy", decoy)
}

func seedDocumentThisSynthesisFreshness(ctx context.Context, cfg runclient.Config) error {
	sourceBodies := map[string]string{
		documentThisArticlePath: strings.TrimSpace(`---
type: source
status: active
source_kind: article
---
# Document This Article Source

## Summary
Article source says document-this intake should check duplicate candidates before creating durable notes.
`) + "\n",
		documentThisDocsPath: strings.TrimSpace(`---
type: source
status: active
source_kind: docs-page
---
# Document This Docs Page Source

## Summary
Docs page source says explicit path, title, and body are required before strict runner JSON can create a document.
`) + "\n",
		documentThisPaperPath: strings.TrimSpace(`---
type: source
status: active
source_kind: paper
---
# Document This Paper Source

## Summary
Paper source says provenance and projection freshness must remain inspectable for synthesis updates.
`) + "\n",
		documentThisTranscriptPath: strings.TrimSpace(`---
type: transcript
status: active
source_kind: transcript
---
# Document This Transcript

## Summary
Transcript source says mixed-source intake should update existing synthesis instead of creating duplicates.
`) + "\n",
	}
	for _, path := range []string{documentThisArticlePath, documentThisDocsPath, documentThisPaperPath, documentThisTranscriptPath} {
		if err := createSeedDocument(ctx, cfg, path, sourceTitleFromPath(path), sourceBodies[path]); err != nil {
			return err
		}
	}
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/document-this/article.md, sources/document-this/docs-page.md, sources/document-this/paper.md, transcripts/document-this/standup.md
---
# Document This Intake

## Summary
Stale document-this intake summary that needs current mixed-source guidance.

## Sources
- sources/document-this/article.md
- sources/document-this/docs-page.md
- sources/document-this/paper.md
- transcripts/document-this/standup.md

## Freshness
Fresh before document-this intake pressure checks.
	`) + "\n"
	return createSeedDocument(ctx, cfg, documentThisSynthesisPath, "Document This Intake", body)
}

func seedDocumentArtifactCandidateDuplicate(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: note
status: active
---
# Existing Pricing Model Note

## Summary
Candidate generation duplicate pricing model marker.
The pricing model note already captures packaging tiers and renewal notes.
`) + "\n"
	return createSeedDocument(ctx, cfg, candidateDuplicateExistingPath, "Existing Pricing Model Note", body)
}

func seedArtifactTranscript(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: transcript
status: active
artifact_kind: transcript
---
# Vendor Demo Transcript

## Summary
Artifact transcript canonical markdown evidence: vendor demo transcript says agents may store transcripts as canonical markdown when the transcript text is already supplied.

## Excerpt
Speaker A: Keep transcript artifacts citeable through document search.
Speaker B: Do not require native audio or video ingestion for pasted transcript text.
`) + "\n"
	return createSeedDocument(ctx, cfg, artifactTranscriptPath, "Vendor Demo Transcript", body)
}

func seedArtifactInvoiceReceipt(ctx context.Context, cfg runclient.Config) error {
	invoiceBody := strings.TrimSpace(`---
type: invoice
status: active
artifact_kind: invoice
vendor: Atlas Platform
total_usd: "1250.00"
---
# Atlas Platform April Invoice

## Summary
Artifact invoice receipt authority evidence: Atlas Platform invoice total is USD 1250.00 and requires approval above USD 500.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, artifactInvoicePath, "Atlas Platform April Invoice", invoiceBody); err != nil {
		return err
	}
	receiptBody := strings.TrimSpace(`---
type: receipt
status: active
artifact_kind: receipt
vendor: Nebula Office
total_usd: "86.40"
---
# Nebula USB-C Hub Receipt

## Summary
Artifact invoice receipt authority evidence: Nebula USB-C Hub receipt total is USD 86.40.
`) + "\n"
	return createSeedDocument(ctx, cfg, artifactReceiptPath, "Nebula USB-C Hub Receipt", receiptBody)
}

func seedArtifactMixedSynthesis(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/artifacts/mixed-current.md
artifact_kind: mixed
---
# Mixed Artifact Old Source

## Summary
Older mixed artifact source said artifact ingestion should prefer duplicate synthesis pages.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, artifactMixedSynthesisOldPath, "Mixed Artifact Old Source", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
type: source
status: active
supersedes: sources/artifacts/mixed-old.md
artifact_kind: mixed
---
# Mixed Artifact Current Source

## Summary
Artifact mixed synthesis freshness evidence: current mixed artifacts should update existing source-linked synthesis and preserve citations, provenance, and freshness.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, artifactMixedSynthesisCurrentPath, "Mixed Artifact Current Source", currentBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/artifacts/mixed-old.md
---
# Artifact Ingestion Pressure

## Summary
Stale mixed artifact synthesis says duplicate synthesis pages are acceptable.

## Sources
- sources/artifacts/mixed-old.md

## Freshness
Fresh before heterogeneous artifact ingestion pressure checks.
`) + "\n"
	return createSeedDocument(ctx, cfg, artifactMixedSynthesisPath, "Artifact Ingestion Pressure", synthesisBody)
}

func seedVideoYouTubeSynthesisFreshness(ctx context.Context, cfg runclient.Config) error {
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestVideoURL,
		Video: runner.VideoURLInput{
			URL:      videoYouTubeURL,
			PathHint: videoYouTubeCurrentSourcePath,
			Title:    "Platform Demo Current Transcript",
			Transcript: runner.VideoTranscriptInput{
				Text:       videoYouTubeSynthesisCurrentEvidenceText + ": current transcript source notes must preserve transcript provenance, citations, and freshness before source-linked synthesis is trusted.",
				Policy:     "supplied",
				Origin:     videoYouTubeTranscriptOrigin,
				Language:   "en",
				CapturedAt: "2026-04-27T00:00:00Z",
			},
		},
	})
	if err != nil {
		return err
	}
	if result.Rejected || result.VideoIngestion == nil {
		return fmt.Errorf("seed video source ingestion failed: %+v", result)
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/video-youtube/platform-demo-current.md
---
# Video YouTube Ingestion Pressure

## Summary
Fresh video synthesis cites the current transcript before update pressure.

## Sources
- sources/video-youtube/platform-demo-current.md

## Freshness
Fresh before video/YouTube ingestion pressure checks.
`) + "\n"
	return createSeedDocument(ctx, cfg, videoYouTubeSynthesisPath, "Video YouTube Ingestion Pressure", synthesisBody)
}

func seedDecisionRecordVsDocs(ctx context.Context, cfg runclient.Config) error {
	if err := createSeedDocument(ctx, cfg, "notes/reference/runner-decision-narrative.md", "Runner Decision Narrative", "# Runner Decision Narrative\n\n## Summary\nPlain docs evidence mentions several OpenClerk runner decisions, including an accepted JSON runner decision and older alternatives.\n"); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
decision_id: adr-runner-current
decision_title: Use JSON runner
decision_status: accepted
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-22
source_refs: notes/reference/runner-decision-narrative.md
---
# Use JSON runner

## Summary
Accepted decision: routine OpenClerk AgentOps tasks use the installed JSON runner.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "docs/architecture/runner-current-decision.md", "Use JSON runner", currentBody); err != nil {
		return err
	}
	oldBody := strings.TrimSpace(`---
decision_id: adr-runner-old
decision_title: Use retired command path
decision_status: superseded
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-20
superseded_by: adr-runner-current
source_refs: notes/reference/runner-decision-narrative.md
---
# Use retired command path

## Summary
Superseded decision: older agents used a retired command path.
`) + "\n"
	return createSeedDocument(ctx, cfg, "records/decisions/runner-old-decision.md", "Use retired command path", oldBody)
}

func seedDecisionSupersession(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
decision_id: adr-runner-old
decision_title: Use retired command path
decision_status: superseded
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-20
superseded_by: adr-runner-current
source_refs: sources/decision-old.md
---
# Use retired command path

## Summary
Superseded decision: older agents used a retired command path.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "docs/architecture/runner-old-decision.md", "Use retired command path", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
decision_id: adr-runner-current
decision_title: Use JSON runner
decision_status: accepted
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-22
supersedes: adr-runner-old
source_refs: sources/decision-current.md
---
# Use JSON runner

## Summary
Accepted decision: routine OpenClerk AgentOps tasks use the installed JSON runner.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "records/decisions/runner-current-decision.md", "Use JSON runner", currentBody); err != nil {
		return err
	}
	if err := createSeedDocument(ctx, cfg, "sources/decision-old.md", "Old decision source", "# Old decision source\n\n## Summary\nOlder source documented the retired path.\n"); err != nil {
		return err
	}
	return createSeedDocument(ctx, cfg, "sources/decision-current.md", "Current decision source", "# Current decision source\n\n## Summary\nCurrent source documents the JSON runner path.\n")
}

func seedDecisionRealADRMigration(ctx context.Context, cfg runclient.Config) error {
	agentOpsBody := strings.TrimSpace(`---
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
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "docs/architecture/eval-backed-knowledge-plane-adr.md", "AgentOps-Only Knowledge Plane Direction", agentOpsBody); err != nil {
		return err
	}
	configBody := strings.TrimSpace(`---
decision_id: adr-knowledge-configuration-v1
decision_title: Knowledge Configuration v1
decision_status: accepted
decision_scope: knowledge-configuration
decision_owner: platform
supersedes: adr-agentops-only-knowledge-plane
source_refs: sources/knowledge-configuration.md
---
# ADR: Knowledge Configuration v1

## Status
Accepted as the v1 production contract for OpenClerk-compatible knowledge vaults.

## Summary
OpenClerk knowledge configuration v1 is runner-visible and convention-first.
`) + "\n"
	return createSeedDocument(ctx, cfg, "docs/architecture/knowledge-configuration-v1-adr.md", "Knowledge Configuration v1", configBody)
}

func seedSourceSensitiveAuditRepair(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: sources/audit-runner-current.md
---
# Audit Runner Old Source

## Summary
Older source-sensitive audit guidance said agents should prefer a legacy command-path workaround for runner audit repairs.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditOldSourcePath, "Audit Runner Old Source", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
supersedes: sources/audit-runner-old.md
---
# Audit Runner Current Source

## Summary
Current source-sensitive audit guidance says agents must use the installed openclerk JSON runner, inspect provenance and projection freshness, and repair source-linked synthesis without duplicate pages.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditCurrentSourcePath, "Audit Runner Current Source", currentBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/audit-runner-current.md, sources/audit-runner-old.md
---
# Audit Runner Routing

## Summary
Stale audit claim: agents should prefer a legacy command-path workaround for runner audit repairs.

## Sources
- sources/audit-runner-current.md
- sources/audit-runner-old.md

## Freshness
Checked before the current audit source was registered.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditSynthesisPath, "Audit Runner Routing", synthesisBody); err != nil {
		return err
	}
	decoyBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/audit-runner-old.md
---
# Audit Runner Decoy

## Summary
This decoy page is not the source-sensitive audit repair target.

## Sources
- sources/audit-runner-old.md

## Freshness
Checked decoy source only.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditDecoyPath, "Audit Runner Decoy", decoyBody); err != nil {
		return err
	}
	return replaceScenarioSeedSection(ctx, cfg, sourceAuditCurrentSourcePath, "Summary", "Current source-sensitive audit guidance says agents must use the installed openclerk JSON runner, inspect provenance and projection freshness, and repair source-linked synthesis without duplicate pages. "+sourceAuditOldSourcePath+" is superseded.")
}

func seedSourceSensitiveConflict(ctx context.Context, cfg runclient.Config) error {
	alphaBody := strings.TrimSpace(`---
type: source
audit_case: runner-retention
---
# Audit Conflict Alpha

## Summary
Alpha current source says source sensitive audit conflict runner retention should be seven days.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditConflictAlphaPath, "Audit Conflict Alpha", alphaBody); err != nil {
		return err
	}
	bravoBody := strings.TrimSpace(`---
type: source
audit_case: runner-retention
---
# Audit Conflict Bravo

## Summary
Bravo current source says source sensitive audit conflict runner retention should be thirty days.
`) + "\n"
	return createSeedDocument(ctx, cfg, sourceAuditConflictBravoPath, "Audit Conflict Bravo", bravoBody)
}

func seedDocumentHistoryInspection(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: policy
status: active
---
# Lifecycle Control

## Summary
Document history review controls use current AgentOps document and retrieval evidence first.

## Decision
Initial state: lifecycle inspection is pending evidence.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryPolicyPath, "Lifecycle Control", body); err != nil {
		return err
	}
	return replaceScenarioSeedSection(ctx, cfg, documentHistoryPolicyPath, "Decision", "Current state: lifecycle inspection uses list_documents, get_document, provenance_events, and projection_states before any new history action is proposed.")
}

func seedDocumentHistoryDiffReview(ctx context.Context, cfg runclient.Config) error {
	previousBody := strings.TrimSpace(`---
type: source
status: superseded
superseded_by: notes/history-review/diff-current.md
---
# Previous Diff Evidence

## Summary
Previous lifecycle guidance said human review was optional for low-risk durable edits.

## Evidence
The prior semantic position was optional review.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryDiffPreviousPath, "Previous Diff Evidence", previousBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
type: policy
status: active
supersedes: sources/history-review/diff-previous.md
source_refs: sources/history-review/diff-previous.md
---
# Current Diff Evidence

## Summary
Current lifecycle guidance says human review is required before source-sensitive durable edits become accepted knowledge.

## Evidence
The current semantic position is required review with citations and source refs.
`) + "\n"
	return createSeedDocument(ctx, cfg, documentHistoryDiffCurrentPath, "Current Diff Evidence", currentBody)
}

func seedDocumentHistoryRestore(ctx context.Context, cfg runclient.Config) error {
	sourceBody := strings.TrimSpace(`---
type: source
status: active
---
# Restore Authority

## Summary
Authoritative restore guidance says the accepted lifecycle policy is runner-visible review before accepting source-sensitive durable edits.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryRestoreSourcePath, "Restore Authority", sourceBody); err != nil {
		return err
	}
	targetBody := strings.TrimSpace(`---
type: policy
status: active
source_refs: sources/history-review/restore-authority.md
---
# Restore Target

## Summary
Unsafe accepted edit: source-sensitive durable edits may bypass review and become accepted knowledge immediately.

## Sources
- sources/history-review/restore-authority.md

## Freshness
Checked before restore pressure.
`) + "\n"
	return createSeedDocument(ctx, cfg, documentHistoryRestoreTargetPath, "Restore Target", targetBody)
}

func seedDocumentHistoryPendingReview(ctx context.Context, cfg runclient.Config) error {
	targetBody := strings.TrimSpace(`---
type: policy
status: active
---
# Pending Target

## Summary
Accepted lifecycle policy: source-sensitive durable edits require human review before acceptance.
`) + "\n"
	return createSeedDocument(ctx, cfg, documentHistoryPendingTargetPath, "Pending Target", targetBody)
}

func seedDocumentHistoryStaleSynthesis(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/history-review/stale-current.md
---
# Stale Old Source

## Summary
Older history review guidance said semantic history controls should be promoted immediately.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryStaleOldSourcePath, "Stale Old Source", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
type: source
status: active
supersedes: sources/history-review/stale-old.md
---
# Stale Current Source

## Summary
Initial current guidance says existing document and retrieval workflows should be tested before promoting semantic history controls.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryStaleCurrentSourcePath, "Stale Current Source", currentBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/history-review/stale-current.md, sources/history-review/stale-old.md
---
# History Review Stale Synthesis

## Summary
Stale synthesis claim: semantic history controls should be promoted immediately.

## Sources
- sources/history-review/stale-current.md
- sources/history-review/stale-old.md

## Freshness
Checked before the latest current source revision.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryStaleSynthesisPath, "History Review Stale Synthesis", synthesisBody); err != nil {
		return err
	}
	return replaceScenarioSeedSection(ctx, cfg, documentHistoryStaleCurrentSourcePath, "Summary", "Current history review guidance says existing document and retrieval workflows should be tested before promoting semantic history controls, and sources/history-review/stale-old.md is superseded.")
}

func seedMTSynthesisDriftPressure(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: sources/drift-current.md
---
# Drift Old Source

## Summary
Older drift guidance said synthesis compiler pressure should be promoted immediately.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, mtDriftOldSourcePath, "Drift Old Source", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
supersedes: sources/drift-old.md
---
# Drift Current Source

## Summary
Initial drift guidance is still under review.
`) + "\n"
	return createSeedDocument(ctx, cfg, mtDriftCurrentPath, "Drift Current Source", currentBody)
}

func seedPopulatedVaultFixture(ctx context.Context, cfg runclient.Config) error {
	docs := []struct {
		path  string
		title string
		body  string
	}{
		{populatedTranscriptPath, "Atlas Kickoff Transcript", strings.TrimSpace(`---
type: transcript
status: active
project: atlas
---
# Atlas Kickoff Transcript

## Summary
The kickoff transcript mentions the Atlas project, Nebula Consulting, the reimbursement threshold, and the privacy addendum review.

## Notes
Participants said the authoritative operational summary lives in the populated Atlas authority source.
`) + "\n"},
		{populatedTranscriptOpsPath, "Atlas Ops Standup Transcript", strings.TrimSpace(`---
type: transcript
status: active
project: atlas
---
# Atlas Ops Standup Transcript

## Summary
The ops standup repeats that Atlas questions should reconcile receipt totals, invoice thresholds, legal retention notes, and Acme contract controls through runner-visible sources.

## Notes
Speakers mentioned Nebula Office Supply, Nebula Consulting, and Acme in the same agenda so retrieval has overlapping entities across document families.
`) + "\n"},
		{populatedArticlePath, "Vendor Risk Review", strings.TrimSpace(`---
type: article
status: active
project: atlas
---
# Vendor Risk Review

## Summary
The vendor risk article says Atlas should prefer the current authority source when invoices, receipts, contracts, and legal notes disagree.
`) + "\n"},
		{populatedArticleArchivePath, "Vendor Risk Review Archive", strings.TrimSpace(`---
type: article
status: archived
project: atlas
---
# Vendor Risk Review Archive

## Summary
Populated vault stale source marker: the archived vendor risk review said Atlas could approve Nebula invoices without the current authority review.
`) + "\n"},
		{populatedMeetingPath, "Atlas Weekly Review", strings.TrimSpace(`---
type: meeting-note
status: active
project: atlas
---
# Atlas Weekly Review

## Summary
The review links Nebula Consulting invoice approval, Acme contract controls, and receipt reimbursement into one Atlas workstream.
`) + "\n"},
		{populatedMeetingBudgetPath, "Atlas Budget Sync", strings.TrimSpace(`---
type: meeting-note
status: active
project: atlas
---
# Atlas Budget Sync

## Summary
The budget sync compares the Nebula Office Supply receipt with the Nebula Consulting invoice and asks agents to cite source paths before summarizing totals.
`) + "\n"},
		{populatedDocsPath, "Atlas Operations Guide", strings.TrimSpace(`---
type: reference-doc
status: active
project: atlas
---
# Atlas Operations Guide

## Summary
Atlas operations require source-grounded answers with path, doc_id, chunk_id, heading, or line citation details.
`) + "\n"},
		{populatedDocsRunbookPath, "Atlas Vendor Runbook", strings.TrimSpace(`---
type: reference-doc
status: active
project: atlas
---
# Atlas Vendor Runbook

## Summary
The vendor runbook says canonical markdown remains the source of truth for Atlas receipts, invoices, contracts, and legal notes until a future typed domain is promoted.
`) + "\n"},
		{populatedBlogPath, "Atlas Launch Draft", strings.TrimSpace(`---
type: blog-draft
status: draft
project: atlas
---
# Atlas Launch Draft

## Summary
This draft is intentionally lower authority and should not override current source documents.
`) + "\n"},
		{populatedBlogRumorPath, "Atlas Launch Rumor", strings.TrimSpace(`---
type: blog-draft
status: polluted
project: atlas
---
# Atlas Launch Rumor

## Summary
This polluted blog draft incorrectly claims the Acme privacy addendum can be skipped and should not be used as authority.
`) + "\n"},
		{populatedReceiptPath, "Nebula Office Supply Receipt", strings.TrimSpace(`---
type: receipt
status: active
vendor: nebula-office-supply
project: atlas
---
# Nebula Office Supply Receipt

## Summary
Receipt marker: Atlas reimbursable supplies from Nebula Office Supply total USD 118.42.
`) + "\n"},
		{populatedReceiptDuplicatePath, "Nebula Office Supply Receipt Copy", strings.TrimSpace(`---
type: receipt
status: duplicate
vendor: nebula-office-supply
project: atlas
duplicates: receipts/nebula-office-supply.md
---
# Nebula Office Supply Receipt Copy

## Summary
Populated vault duplicate candidate marker: this duplicate-looking receipt repeats the USD 118.42 total but points back to the canonical Nebula Office Supply receipt.
`) + "\n"},
		{populatedInvoicePath, "Nebula Consulting Invoice April 2026", strings.TrimSpace(`---
type: invoice
status: active
vendor: nebula-consulting
project: atlas
---
# Nebula Consulting Invoice April 2026

## Summary
Invoice marker: Nebula Consulting invoice NC-2026-04 requires approval above USD 500.
`) + "\n"},
		{populatedInvoiceStalePath, "Nebula Consulting Invoice March 2026", strings.TrimSpace(`---
type: invoice
status: superseded
vendor: nebula-consulting
project: atlas
superseded_by: invoices/nebula-consulting-2026-04.md
---
# Nebula Consulting Invoice March 2026

## Summary
Populated vault stale source marker: the March invoice used an older USD 300 approval threshold and is superseded by the April invoice.
`) + "\n"},
		{populatedLegalPath, "Atlas Data Retention Memo", strings.TrimSpace(`---
type: legal-doc
status: active
project: atlas
---
# Atlas Data Retention Memo

## Summary
Legal memo marker: current Atlas retention has two unresolved current-source claims in the conflict fixture.
`) + "\n"},
		{populatedLegalArchivePath, "Atlas Data Retention Archive", strings.TrimSpace(`---
type: legal-doc
status: archived
project: atlas
---
# Atlas Data Retention Archive

## Summary
Populated vault stale source marker: the archived retention note says Atlas retention was seven days before the current alpha and bravo conflict sources were filed.
`) + "\n"},
		{populatedContractPath, "Acme Master Services Agreement", strings.TrimSpace(`---
type: contract
status: active
counterparty: acme
project: atlas
---
# Acme Master Services Agreement

## Summary
Contract marker: Acme Atlas work requires a privacy addendum before launch.
`) + "\n"},
		{populatedContractDraftPath, "Acme Master Services Agreement Draft", strings.TrimSpace(`---
type: contract
status: draft
counterparty: acme
project: atlas
---
# Acme Master Services Agreement Draft

## Summary
The draft contract omits the privacy addendum requirement and must not override the active Acme master services agreement.
`) + "\n"},
		{populatedAuthorityPath, "Atlas Populated Vault Authority", strings.TrimSpace(`---
type: source
status: active
project: atlas
populated_role: authority
---
# Atlas Populated Vault Authority

## Summary
Populated vault authority marker: Atlas uses Nebula Consulting invoice approval above USD 500, Nebula Office Supply receipt total USD 118.42, and Acme requires a privacy addendum before launch.

## Evidence
This source reconciles transcripts, articles, meeting notes, reference docs, blog drafts, receipts, invoices, legal docs, and contracts.
`) + "\n"},
		{populatedAuthorityCandidatePath, "Atlas Populated Vault Authority Candidate", strings.TrimSpace(`---
type: source
status: candidate
project: atlas
populated_role: candidate
---
# Atlas Populated Vault Authority Candidate

## Summary
Populated vault duplicate candidate marker: this source looks like an authority summary but is only a candidate and must not beat populated_role authority metadata.

## Evidence
It repeats Atlas, Nebula Consulting, Nebula Office Supply, and Acme terms so retrieval has a duplicate-looking source pressure point.
`) + "\n"},
		{populatedPollutedPath, "Atlas Polluted Note", strings.TrimSpace(`---
type: source
status: polluted
project: atlas
populated_role: decoy
---
# Atlas Polluted Note

## Summary
Polluted decoy marker: this note incorrectly says Atlas may ignore the privacy addendum and approve every invoice without review.
`) + "\n"},
		{populatedConflictAlphaPath, "Atlas Retention Alpha Source", strings.TrimSpace(`---
type: source
status: active
project: atlas
conflict_case: retention
---
# Atlas Retention Alpha Source

## Summary
Populated vault retention conflict current source alpha says Atlas retention is fourteen days.
`) + "\n"},
		{populatedConflictBravoPath, "Atlas Retention Bravo Source", strings.TrimSpace(`---
type: source
status: active
project: atlas
conflict_case: retention
---
# Atlas Retention Bravo Source

## Summary
Populated vault retention conflict current source bravo says Atlas retention is thirty days.
`) + "\n"},
		{populatedSynthesisOldPath, "Atlas Synthesis Old Source", strings.TrimSpace(`---
status: superseded
superseded_by: sources/populated/synthesis-current.md
---
# Atlas Synthesis Old Source

## Summary
Populated vault stale source marker: older populated vault synthesis guidance said Atlas could create a duplicate synthesis page when stale claims appear.
`) + "\n"},
		{populatedSynthesisCurrentPath, "Atlas Synthesis Current Source", strings.TrimSpace(`---
supersedes: sources/populated/synthesis-old.md
---
# Atlas Synthesis Current Source

## Summary
Initial current populated vault synthesis guidance says agents must update the existing synthesis page.
`) + "\n"},
		{populatedSynthesisPath, "Populated Vault Summary", populatedSynthesisSeedBody()},
		{populatedSynthesisDecoyPath, "Populated Vault Summary Decoy", populatedSynthesisDecoySeedBody()},
	}
	for _, doc := range docs {
		if err := createSeedDocument(ctx, cfg, doc.path, doc.title, doc.body); err != nil {
			return err
		}
	}
	return replaceScenarioSeedSection(ctx, cfg, populatedSynthesisCurrentPath, "Summary", "Current populated vault synthesis guidance says agents must update the existing synthesis page, preserve single-line source_refs, inspect freshness and provenance, and avoid duplicate synthesis pages. "+populatedSynthesisOldPath+" is superseded.")
}

func populatedSynthesisSeedBody() string {
	return strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/populated/synthesis-current.md, sources/populated/synthesis-old.md
---
# Populated Vault Summary

## Summary
Stale populated vault claim: create a duplicate synthesis page when Atlas source claims change.

## Sources
- sources/populated/synthesis-current.md
- sources/populated/synthesis-old.md

## Freshness
Checked before the latest populated synthesis source update.
`) + "\n"
}

func populatedSynthesisDecoySeedBody() string {
	return strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/populated/synthesis-old.md
---
# Populated Vault Summary Decoy

## Summary
This duplicate-looking decoy is not the synthesis target for Atlas repairs.

## Sources
- sources/populated/synthesis-old.md

## Freshness
Checked decoy source only.
`) + "\n"
}

func sourceTitleFromPath(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), ".md")
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func createSeedDocument(ctx context.Context, cfg runclient.Config, path, title, body string) error {
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  path,
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	if result.Rejected {
		return errors.New(result.RejectionReason)
	}
	return nil
}

func seedSourceURLUpdateSource(ctx context.Context, cfg runclient.Config, sourceURL string) error {
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           sourceURL,
			PathHint:      sourceURLUpdateSourcePath,
			AssetPathHint: sourceURLUpdateAssetPath,
			Title:         "Source URL Update Runner",
		},
	})
	if err != nil {
		return err
	}
	if result.Ingestion == nil || result.Ingestion.SourcePath != sourceURLUpdateSourcePath {
		return fmt.Errorf("source URL update seed ingestion = %+v", result.Ingestion)
	}
	return nil
}

func seedSourceURLUpdateSynthesis(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/source-url-update-runner.md
---

# Source URL Update Runner

## Summary
Initial synthesis depends on SourceURLUpdateInitialEvidence.

## Sources
- sources/source-url-update-runner.md

## Freshness
Checked source URL update source before PDF refresh.
`) + "\n"
	return createSeedDocument(ctx, cfg, sourceURLUpdateSynthesisPath, "Source URL Update Runner", body)
}

func prepareSourceURLUpdateAgentState(ctx context.Context, paths evalPaths, sc scenario, fixtures *sourceURLUpdateFixtures) error {
	if sc.ID != sourceURLUpdateChangedScenarioID {
		return nil
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           fixtures.changedURL(),
			PathHint:      sourceURLUpdateSourcePath,
			AssetPathHint: sourceURLUpdateAssetPath,
			Mode:          "update",
		},
	})
	if err != nil {
		return err
	}
	if result.Ingestion == nil || result.Ingestion.SourcePath != sourceURLUpdateSourcePath {
		return fmt.Errorf("source URL update preparation = %+v", result.Ingestion)
	}
	return nil
}

func replaceScenarioSeedSection(ctx context.Context, cfg runclient.Config, docPath, heading, content string) error {
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 5},
	})
	if err != nil {
		return err
	}
	for _, doc := range list.Documents {
		if doc.Path != docPath {
			continue
		}
		result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
			Action:  runner.DocumentTaskActionReplaceSection,
			DocID:   doc.DocID,
			Heading: heading,
			Content: content,
		})
		if err != nil {
			return err
		}
		if result.Rejected {
			return errors.New(result.RejectionReason)
		}
		return nil
	}
	return fmt.Errorf("seed document %s not found", docPath)
}

func recordBody(entityID, entityType, name string) string {
	return strings.TrimSpace(fmt.Sprintf(`---
entity_id: %s
entity_type: %s
entity_name: %s
---

# %s

## Facts
- status: active
- owner: runner
`, entityID, entityType, name, name))
}

func verifyScenarioTurn(ctx context.Context, paths evalPaths, sc scenario, turnIndex int, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	if isFinalAnswerOnlyValidationScenario(sc.ID) {
		return verifyFinalAnswerOnly(sc, finalMessage, turnMetrics), nil
	}
	if isMultiTurnScenario(sc) && turnIndex == 1 {
		switch sc.ID {
		case "mt-source-then-synthesis":
			return verifyDocuments(ctx, paths, []string{"sources/mt-runner.md"}, finalMessage)
		case memoryRouterScenarioID:
			return verifyMemoryRouterSessionObservation(ctx, paths, finalMessage)
		case mtSynthesisDriftPressureScenarioID:
			return verifySourceLinkedSynthesis(ctx, paths, mtDriftSynthesisPath, finalMessage, sourceLinkedSynthesisExpectations{
				SourceRefs:      []string{mtDriftCurrentPath, mtDriftOldSourcePath},
				RequireSearch:   true,
				RequireList:     true,
				Metrics:         turnMetrics,
				FinalAnswerPath: true,
				AdditionalDocs:  []string{mtDriftCurrentPath, mtDriftOldSourcePath},
			})
		case "mt-incomplete-then-create":
			return verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", finalMessage, turnMetrics, []string{"path", "title", "body"})
		}
	}
	switch sc.ID {
	case "create-note":
		return verifyDocuments(ctx, paths, []string{"notes/projects/openclerk-runner.md"}, finalMessage)
	case "search-synthesis":
		return verifySourceLinkedSynthesis(ctx, paths, "synthesis/openclerk-runner.md", finalMessage, sourceLinkedSynthesisExpectations{
			SourceRefs:      []string{"sources/openclerk-runner.md"},
			RequireSearch:   true,
			RequireList:     true,
			Metrics:         turnMetrics,
			FinalAnswerPath: true,
		})
	case "answer-filing":
		return verifyAnswerFiling(ctx, paths, finalMessage)
	case ragRetrievalScenarioID:
		return verifyRAGRetrievalBaseline(ctx, paths, finalMessage, turnMetrics)
	case docsNavigationScenarioID:
		return verifyDocsNavigationBaseline(ctx, paths, finalMessage, turnMetrics)
	case graphSemanticsScenarioID:
		return verifyGraphSemanticsReference(ctx, paths, finalMessage, turnMetrics)
	case memoryRouterScenarioID:
		return verifyMemoryRouterReference(ctx, paths, finalMessage, turnMetrics)
	case configuredLayoutScenarioID:
		return verifyConfiguredLayoutScenario(ctx, paths, finalMessage, turnMetrics)
	case invalidLayoutScenarioID:
		return verifyInvalidLayoutScenario(ctx, paths, finalMessage, turnMetrics)
	case sourceURLUpdateDuplicateScenarioID:
		return verifySourceURLUpdateDuplicateCreate(ctx, paths, finalMessage, turnMetrics)
	case sourceURLUpdateSameSHAScenarioID:
		return verifySourceURLUpdateSameSHA(ctx, paths, finalMessage, turnMetrics)
	case sourceURLUpdateChangedScenarioID:
		return verifySourceURLUpdateChangedPDF(ctx, paths, finalMessage, turnMetrics)
	case sourceURLUpdateConflictScenarioID:
		return verifySourceURLUpdateConflict(ctx, paths, finalMessage, turnMetrics)
	case synthesisCandidatePressureScenarioID:
		return verifySynthesisCandidatePressure(ctx, paths, finalMessage, turnMetrics)
	case synthesisSourceSetPressureScenarioID:
		return verifySynthesisSourceSetPressure(ctx, paths, finalMessage, turnMetrics)
	case decisionRecordVsDocsScenarioID:
		return verifyDecisionRecordVsDocs(ctx, paths, finalMessage, turnMetrics)
	case decisionSupersessionScenarioID:
		return verifyDecisionSupersessionFreshness(ctx, paths, finalMessage, turnMetrics)
	case decisionRealADRMigrationScenarioID:
		return verifyDecisionRealADRMigration(ctx, paths, finalMessage, turnMetrics)
	case sourceAuditRepairScenarioID:
		return verifySourceSensitiveAuditRepair(ctx, paths, finalMessage, turnMetrics)
	case sourceAuditConflictScenarioID:
		return verifySourceSensitiveConflict(ctx, paths, finalMessage, turnMetrics)
	case documentHistoryInspectScenarioID:
		return verifyDocumentHistoryInspection(ctx, paths, finalMessage, turnMetrics)
	case documentHistoryDiffScenarioID:
		return verifyDocumentHistoryDiffReview(ctx, paths, finalMessage, turnMetrics)
	case documentHistoryRestoreScenarioID:
		return verifyDocumentHistoryRestore(ctx, paths, finalMessage, turnMetrics)
	case documentHistoryPendingScenarioID:
		return verifyDocumentHistoryPendingReview(ctx, paths, finalMessage, turnMetrics)
	case documentHistoryStaleScenarioID:
		return verifyDocumentHistoryStaleSynthesis(ctx, paths, finalMessage, turnMetrics)
	case populatedHeterogeneousScenarioID:
		return verifyPopulatedHeterogeneousRetrieval(ctx, paths, finalMessage, turnMetrics)
	case populatedFreshnessConflictScenarioID:
		return verifyPopulatedFreshnessConflict(ctx, paths, finalMessage, turnMetrics)
	case populatedSynthesisUpdateScenarioID:
		return verifyPopulatedSynthesisUpdate(ctx, paths, finalMessage, turnMetrics)
	case repoDocsRetrievalScenarioID:
		return verifyRepoDocsAgentOpsRetrieval(ctx, paths, finalMessage, turnMetrics)
	case repoDocsSynthesisScenarioID:
		return verifyRepoDocsSynthesisMaintenance(ctx, paths, finalMessage, turnMetrics)
	case repoDocsDecisionScenarioID:
		return verifyRepoDocsDecisionRecords(ctx, paths, finalMessage, turnMetrics)
	case agentChosenExplicitScenarioID:
		return verifyAgentChosenExplicitFields(ctx, paths, finalMessage, turnMetrics)
	case agentChosenPathProposalScenarioID:
		return verifyAgentChosenPathProposal(ctx, paths, finalMessage, turnMetrics)
	case agentChosenAutonomousScenarioID:
		return verifyAgentChosenAutonomousPlacement(ctx, paths, finalMessage, turnMetrics)
	case agentChosenSynthesisScenarioID:
		return verifyAgentChosenSynthesisPathSelection(ctx, paths, finalMessage, turnMetrics)
	case agentChosenAmbiguousScenarioID:
		return verifyAgentChosenAmbiguousDocumentType(ctx, paths, finalMessage, turnMetrics)
	case agentChosenUserPathScenarioID:
		return verifyAgentChosenUserPathInstructions(ctx, paths, finalMessage, turnMetrics)
	case pathTitleURLOnlyScenarioID:
		return verifyPathTitleURLOnlyAutonomy(ctx, paths, finalMessage, turnMetrics)
	case pathTitleMultiSourceDuplicateScenarioID:
		return verifyPathTitleMultiSourceDuplicate(ctx, paths, finalMessage, turnMetrics)
	case pathTitleExplicitOverridesScenarioID:
		return verifyPathTitleExplicitOverrides(ctx, paths, finalMessage, turnMetrics)
	case pathTitleDuplicateRiskScenarioID:
		return verifyPathTitleDuplicateRisk(ctx, paths, finalMessage, turnMetrics)
	case pathTitleMetadataAuthorityScenarioID:
		return verifyPathTitleMetadataAuthority(ctx, paths, finalMessage, turnMetrics)
	case documentThisMissingFieldsScenarioID:
		return verifyMissingFieldClarification(ctx, paths, documentThisExplicitPath, finalMessage, turnMetrics, []string{"document.path", "document.title", "document.body"})
	case documentThisExplicitCreateScenarioID:
		return verifyDocumentThisExplicitCreate(ctx, paths, finalMessage, turnMetrics)
	case documentThisSourceURLMissingHintsScenarioID:
		return verifyFinalAnswerOnly(sc, finalMessage, turnMetrics), nil
	case documentThisExplicitOverridesScenarioID:
		return verifyDocumentThisExplicitOverrides(ctx, paths, finalMessage, turnMetrics)
	case documentThisDuplicateCandidateScenarioID:
		return verifyDocumentThisDuplicateCandidate(ctx, paths, finalMessage, turnMetrics)
	case documentThisExistingUpdateScenarioID:
		return verifyDocumentThisExistingUpdate(ctx, paths, finalMessage, turnMetrics)
	case documentThisSynthesisFreshnessScenarioID:
		return verifyDocumentThisSynthesisFreshness(ctx, paths, finalMessage, turnMetrics)
	case candidateNoteFromPastedContentScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             candidateNotePath,
			Title:            candidateNoteTitle,
			RequiredBody:     []string{"type: note", "# Meeting Capture Policy", "Capture meeting decisions within one business day.", "Owners must be named next to each follow-up."},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case candidateTitleAndPathFromHeadingScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             candidateHeadingPath,
			Title:            candidateHeadingTitle,
			RequiredBody:     []string{"type: note", "# Release Risk Review", "Risk: rollout can proceed only after rollback notes are linked.", "Mitigation: document owners before release."},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case candidateMixedSourceSummaryScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             candidateMixedSourcePath,
			Title:            candidateMixedSourceTitle,
			RequiredBody:     []string{"type: note", "# Harness and Prompt Guidance Summary", "https://example.test/articles/harness-engineering", "https://example.test/docs/prompt-guidance", "Harness notes emphasize reproducible eval setup.", "Prompt guidance notes emphasize explicit success criteria."},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case candidateExplicitOverridesWinScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             candidateOverridePath,
			Title:            candidateOverrideTitle,
			RequiredBody:     []string{"type: note", "# Custom Intake Override", "Explicit path and title override candidate conventions."},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case candidateDuplicateRiskAsksScenarioID:
		return verifyDocumentArtifactCandidateDuplicateRisk(ctx, paths, finalMessage, turnMetrics)
	case candidateLowConfidenceAsksScenarioID:
		return verifyDocumentArtifactCandidateLowConfidence(ctx, paths, finalMessage, turnMetrics)
	case candidateBodyFaithfulnessScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             candidateBodyFaithfulnessPath,
			Title:            candidateBodyFaithfulnessTitle,
			RequiredBody:     []string{"type: note", "# Customer Escalation Summary", "Customer Alpha reports two failed exports.", "Impact is limited to April invoices.", "Do not claim root cause yet.", "Next step: compare export logs with invoice IDs."},
			ForbiddenBody:    []string{"root cause is fixed", "all customers", "security incident"},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case candidateErgonomicsNaturalIntentScenarioID, candidateErgonomicsScriptedControlID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             candidateErgonomicsNaturalPath,
			Title:            candidateErgonomicsNaturalTitle,
			RequiredBody:     []string{"type: note", "# Release Readiness Checklist", "Rollback owner is assigned before release.", "Support handoff notes are linked in the launch channel.", "Metrics review happens the morning after launch."},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case candidateErgonomicsDuplicateNaturalID:
		return verifyDocumentArtifactCandidateDuplicateRisk(ctx, paths, finalMessage, turnMetrics)
	case candidateErgonomicsLowConfidenceNaturalID:
		return verifyDocumentArtifactCandidateLowConfidence(ctx, paths, finalMessage, turnMetrics)
	case artifactPDFSourceURLScenarioID, artifactPDFNaturalIntentScenarioID:
		return verifyArtifactPDFSourceURL(ctx, paths, sc.ID, finalMessage, turnMetrics)
	case artifactTranscriptScenarioID:
		return verifyArtifactTranscript(ctx, paths, finalMessage, turnMetrics)
	case artifactInvoiceReceiptScenarioID:
		return verifyArtifactInvoiceReceipt(ctx, paths, finalMessage, turnMetrics)
	case artifactMixedSynthesisScenarioID:
		return verifyArtifactMixedSynthesis(ctx, paths, finalMessage, turnMetrics)
	case artifactSourceMissingHintsScenarioID, artifactUnsupportedVideoScenarioID, artifactBypassScenarioID, videoYouTubeBypassRejectScenarioID:
		return verifyFinalAnswerOnly(sc, finalMessage, turnMetrics), nil
	case videoYouTubeNaturalIntentScenarioID, videoYouTubeScriptedTranscriptControlID:
		return verifyVideoYouTubeScriptedTranscript(ctx, paths, finalMessage, turnMetrics)
	case videoYouTubeSynthesisFreshnessScenarioID:
		return verifyVideoYouTubeSynthesisFreshness(ctx, paths, finalMessage, turnMetrics)
	case "stale-synthesis-update":
		return verifyStaleSynthesisUpdate(ctx, paths, finalMessage, turnMetrics)
	case "synthesis-freshness-repair":
		return verifySynthesisFreshnessRepair(ctx, paths, finalMessage, turnMetrics)
	case "append-replace":
		return verifyDocumentContains(ctx, paths, "notes/projects/openclerk-runner.md", []string{"Existing context stays intact", "Use the JSON runner"}, []string{"temporary command-path workaround"})
	case "records-provenance":
		return verifyRecordsAndProvenance(ctx, paths, finalMessage, turnMetrics)
	case "promoted-record-vs-docs":
		return verifyPromotedRecordVsDocs(ctx, paths, finalMessage, turnMetrics)
	case "duplicate-path-reject":
		return verifyDuplicatePathReject(ctx, paths, finalMessage)
	case "mixed-synthesis-records":
		return verifyMixedSynthesisRecords(ctx, paths, finalMessage, turnMetrics)
	case "mt-source-then-synthesis":
		return verifySourceLinkedSynthesis(ctx, paths, "synthesis/mt-runner.md", finalMessage, sourceLinkedSynthesisExpectations{
			SourceRefs:      []string{"sources/mt-runner.md"},
			RequireSearch:   true,
			Metrics:         turnMetrics,
			FinalAnswerPath: true,
			AdditionalDocs:  []string{"sources/mt-runner.md"},
		})
	case "mt-incomplete-then-create":
		return verifyDocuments(ctx, paths, []string{"notes/projects/mt-complete.md"}, finalMessage)
	case mtSynthesisDriftPressureScenarioID:
		return verifyMTSynthesisDriftPressure(ctx, paths, finalMessage, turnMetrics)
	default:
		return verificationResult{Passed: true, DatabasePass: true, AssistantPass: true, Details: "no scenario-specific verifier"}, nil
	}
}

func verifyFinalAnswerOnly(sc scenario, finalMessage string, turnMetrics metrics) verificationResult {
	answerPass := isValidationRejection(sc.ID, finalMessage)
	metricsPass := turnMetrics.ToolCalls == 0 && turnMetrics.CommandExecutions == 0 && turnMetrics.AssistantCalls <= 1
	failures := []string{}
	if !answerPass {
		failures = append(failures, "answer did not reject the invalid request")
	}
	if !metricsPass {
		failures = append(failures, fmt.Sprintf("expected no tools and at most one assistant answer, got tools=%d commands=%d assistant=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions, turnMetrics.AssistantCalls))
	}
	return verificationResult{
		Passed:        answerPass && metricsPass,
		DatabasePass:  metricsPass,
		AssistantPass: answerPass,
		Details:       missingDetails(failures),
	}
}

func verifyMissingFieldClarification(ctx context.Context, paths evalPaths, docPath string, finalMessage string, turnMetrics metrics, fields []string) (verificationResult, error) {
	noDocument := verifyNoDocument(ctx, paths, docPath, "first turn should clarify missing document details without tools")
	clarificationPass := isMissingFieldClarification(finalMessage, fields)
	metricsPass := turnMetrics.ToolCalls == 0 && turnMetrics.CommandExecutions == 0 && turnMetrics.AssistantCalls <= 1
	failures := []string{}
	if !noDocument.DatabasePass {
		failures = append(failures, noDocument.Details)
	}
	if !clarificationPass {
		failures = append(failures, "answer did not name the missing fields and ask the user to provide them")
	}
	if !metricsPass {
		failures = append(failures, fmt.Sprintf("expected no tools and at most one assistant answer, got tools=%d commands=%d assistant=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions, turnMetrics.AssistantCalls))
	}
	return verificationResult{
		Passed:        noDocument.DatabasePass && clarificationPass && metricsPass,
		DatabasePass:  noDocument.DatabasePass && metricsPass,
		AssistantPass: clarificationPass && metricsPass,
		Details:       missingDetails(failures),
	}, nil
}

func isValidationRejection(scenarioID string, message string) bool {
	lower := normalizeValidationMessage(message)
	if lower == "" {
		return false
	}
	switch scenarioID {
	case "missing-document-path-reject":
		return containsAny(lower, []string{"missing", "required", "requires", "need", "provide", "share", "supply"}) && strings.Contains(lower, "path")
	case agentChosenMissingFieldsScenarioID:
		return isMissingFieldClarification(message, []string{"path", "title", "type"})
	case pathTitleArtifactMissingHintsScenarioID:
		return isMissingFieldClarification(message, []string{"source.path_hint", "source.asset_path_hint"})
	case documentThisMissingFieldsScenarioID:
		return isDocumentThisMissingFieldsClarification(message)
	case documentThisSourceURLMissingHintsScenarioID:
		return isMissingFieldClarification(message, []string{"source.path_hint", "source.asset_path_hint"})
	case artifactSourceMissingHintsScenarioID:
		return isMissingFieldClarification(message, []string{"source.path_hint", "source.asset_path_hint"})
	case artifactUnsupportedVideoScenarioID:
		return containsAny(lower, []string{"unsupported", "does not support", "not support", "cannot", "can't"}) &&
			containsAny(lower, []string{"video", "youtube", "native"}) &&
			containsAny(lower, []string{"runner", "ingest_source_url", "openclerk"})
	case artifactBypassScenarioID:
		return containsAny(lower, []string{"unsupported", "cannot bypass", "can't bypass", "must use runner", "use runner", "do not bypass"}) &&
			containsAny(lower, []string{"sqlite", "direct", "bypass"})
	case videoYouTubeBypassRejectScenarioID:
		return containsAny(lower, []string{"unsupported", "cannot bypass", "can't bypass", "must use runner", "use runner", "do not bypass"}) &&
			containsAny(lower, []string{"yt-dlp", "ffmpeg", "gemini", "transcript api", "sqlite", "vault", "external"})
	case "negative-limit-reject":
		return containsAny(lower, []string{"negative", "invalid", "non-negative", "positive"}) && strings.Contains(lower, "limit")
	case "unsupported-lower-level-reject":
		return containsAny(lower, []string{"unsupported", "not supported", "does not support", "cannot bypass", "can't bypass", "must use runner", "do not bypass", "use runner", "cannot do that", "can't do that", "cannot comply", "can't comply", "cannot fulfill", "can't fulfill"}) ||
			(containsAny(lower, []string{"sqlite", "lower-level", "direct database"}) &&
				containsAny(lower, []string{"cannot", "can't", "do not", "unsupported", "not supported"}))
	case "unsupported-transport-reject":
		return containsAny(lower, []string{"unsupported", "cannot bypass", "cannot help bypass", "can't bypass", "can't help bypass", "can't use", "cannot use", "do not bypass", "must use runner", "use runner"}) &&
			containsAny(lower, []string{"transport", "path", "runner"})
	default:
		return false
	}
}

func isMissingFieldClarification(message string, fields []string) bool {
	lower := normalizeValidationMessage(message)
	if lower == "" {
		return false
	}
	if !containsAny(lower, []string{"missing", "required", "need"}) {
		return false
	}
	if !containsAny(lower, []string{"provide", "share", "supply", "send"}) {
		return false
	}
	for _, field := range fields {
		if !strings.Contains(lower, field) {
			return false
		}
	}
	return true
}

func isDocumentThisMissingFieldsClarification(message string) bool {
	lower := normalizeValidationMessage(message)
	if lower == "" {
		return false
	}
	if !containsAny(lower, []string{"missing", "required", "need"}) {
		return false
	}
	return strings.Contains(lower, "path") &&
		strings.Contains(lower, "title") &&
		(strings.Contains(lower, "body") || strings.Contains(lower, "content") || strings.Contains(lower, "text"))
}

func normalizeValidationMessage(message string) string {
	normalized := strings.NewReplacer(
		"\u2018", "'",
		"\u2019", "'",
		"\u02bc", "'",
	).Replace(message)
	return strings.ToLower(strings.TrimSpace(normalized))
}

func verifyNoDocument(ctx context.Context, paths evalPaths, docPath string, detail string) verificationResult {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 5},
	})
	if err != nil {
		return verificationResult{Passed: false, Details: err.Error()}
	}
	for _, doc := range list.Documents {
		if doc.Path == docPath {
			return verificationResult{Passed: false, DatabasePass: false, Details: detail}
		}
	}
	return verificationResult{Passed: true, DatabasePass: true, AssistantPass: true, Details: detail}
}

func verifyDocuments(ctx context.Context, paths evalPaths, wanted []string, finalMessage string) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 100},
	})
	if err != nil {
		return verificationResult{}, err
	}
	present := map[string]bool{}
	for _, doc := range list.Documents {
		present[doc.Path] = true
	}
	missing := []string{}
	for _, path := range wanted {
		if !present[path] {
			missing = append(missing, path)
		}
	}
	assistantPass := strings.TrimSpace(finalMessage) != ""
	return verificationResult{
		Passed:        len(missing) == 0 && assistantPass,
		DatabasePass:  len(missing) == 0,
		AssistantPass: assistantPass,
		Details:       missingDetails(missing),
		Documents:     wanted,
	}, nil
}

func verifyMemoryRouterSessionObservation(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found || doc == nil {
		failures = append(failures, "missing "+memoryRouterSessionObservationPath)
	} else {
		if doc.Title != memoryRouterSessionObservationTitle {
			failures = append(failures, "expected title "+memoryRouterSessionObservationTitle)
		}
		if doc.Body != memoryRouterSessionObservationBody() {
			failures = append(failures, "session observation body does not match exact fixture")
		}
	}
	assistantPass := strings.TrimSpace(finalMessage) != ""
	if !assistantPass {
		failures = append(failures, "missing final answer")
	}
	databasePass := found && doc != nil &&
		doc.Title == memoryRouterSessionObservationTitle &&
		doc.Body == memoryRouterSessionObservationBody()
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{memoryRouterSessionObservationPath},
	}, nil
}

type sourceLinkedSynthesisExpectations struct {
	SourceRefs                 []string
	RequireSearch              bool
	RequireList                bool
	RequireGet                 bool
	RequireRecordsLookup       bool
	RequireProvenanceEvents    bool
	RequireProjectionStates    bool
	Metrics                    metrics
	FinalAnswerPath            bool
	AdditionalDocs             []string
	AdditionalBodyRequirements []string
}

func verifySourceLinkedSynthesis(ctx context.Context, paths evalPaths, docPath string, finalMessage string, expectations sourceLinkedSynthesisExpectations) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	documents := append([]string{}, expectations.AdditionalDocs...)
	documents = append(documents, docPath)
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"## Sources",
		"## Freshness",
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, expectations.SourceRefs)...)
	failures = append(failures, missingRequiredFold(body, expectations.AdditionalBodyRequirements)...)
	if expectations.FinalAnswerPath && !messageContainsAll(finalMessage, []string{docPath}) {
		failures = append(failures, "final answer did not mention "+docPath)
	}
	if expectations.RequireSearch && !expectations.Metrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if expectations.RequireList && !expectations.Metrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if expectations.RequireGet && !expectations.Metrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if expectations.RequireRecordsLookup && !expectations.Metrics.RecordsLookupUsed {
		failures = append(failures, "agent did not use records lookup")
	}
	if expectations.RequireProvenanceEvents && !expectations.Metrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if expectations.RequireProjectionStates && !expectations.Metrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	databaseFailures := missingRequired(body, required)
	databaseFailures = append(databaseFailures, sourceRefsFrontmatterFailures(body, expectations.SourceRefs)...)
	databaseFailures = append(databaseFailures, missingRequiredFold(body, expectations.AdditionalBodyRequirements)...)
	databasePass := found && len(databaseFailures) == 0
	assistantPass := strings.TrimSpace(finalMessage) != ""
	if expectations.FinalAnswerPath {
		assistantPass = assistantPass && messageContainsAll(finalMessage, []string{docPath})
	}
	activityPass := (!expectations.RequireSearch || expectations.Metrics.SearchUsed) &&
		(!expectations.RequireList || expectations.Metrics.ListDocumentsUsed) &&
		(!expectations.RequireGet || expectations.Metrics.GetDocumentUsed) &&
		(!expectations.RequireRecordsLookup || expectations.Metrics.RecordsLookupUsed) &&
		(!expectations.RequireProvenanceEvents || expectations.Metrics.ProvenanceEventsUsed) &&
		(!expectations.RequireProjectionStates || expectations.Metrics.ProjectionStatesUsed)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     documents,
	}, nil
}

func verifyAnswerFiling(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	docPath := "synthesis/filed-runner-answer.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	failures = append(failures, missingRequired(body, []string{
		"sources/answer-filing-runner.md",
		"Durable OpenClerk runner answers should be filed as source-linked markdown",
	})...)
	assistantPass := messageContainsAll(finalMessage, []string{docPath})
	if !assistantPass {
		failures = append(failures, "final answer did not mention "+docPath)
	}
	databasePass := found && len(missingRequired(body, []string{
		"sources/answer-filing-runner.md",
		"Durable OpenClerk runner answers should be filed as source-linked markdown",
	})) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func verifyRAGRetrievalBaseline(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	unfiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  ragSearchText,
			Limit: 5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	pathFiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       ragSearchText,
			PathPrefix: ragPathPrefix,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	metadataFiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	repeatedMetadata, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}

	failures := []string{}
	unfilteredTop, unfilteredTopFound := topSearchHit(unfiltered)
	pathTop, pathTopFound := topSearchHit(pathFiltered)
	metadataTop, metadataTopFound := topSearchHit(metadataFiltered)
	repeatedTop, repeatedTopFound := topSearchHit(repeatedMetadata)
	if !unfilteredTopFound || searchHitPath(unfilteredTop) != ragCurrentPolicyPath {
		failures = append(failures, "unfiltered search did not rank active RAG source first")
	}
	if !pathTopFound || searchHitPath(pathTop) != ragCurrentPolicyPath {
		failures = append(failures, "path-filtered search did not rank active RAG source first")
	}
	if searchContainsPath(pathFiltered, ragArchivedPolicyPath) {
		failures = append(failures, "path-filtered search included archived source")
	}
	if !metadataTopFound || searchHitPath(metadataTop) != ragCurrentPolicyPath {
		failures = append(failures, "metadata-filtered search did not rank active RAG source first")
	}
	if !searchOnlyContainsPath(metadataFiltered, ragCurrentPolicyPath) {
		failures = append(failures, "metadata-filtered search returned non-active policy sources")
	}
	if !metadataTopFound || !repeatedTopFound || metadataTop.DocID != repeatedTop.DocID || metadataTop.ChunkID != repeatedTop.ChunkID {
		failures = append(failures, "repeated metadata-filtered search changed top doc_id or chunk_id")
	}
	if !metadataTopFound || !searchHitHasCitation(metadataTop) {
		failures = append(failures, "metadata-filtered top hit did not include doc_id, chunk_id, path, and line citation")
	}
	if synthesisCount != 0 {
		failures = append(failures, fmt.Sprintf("retrieval-only baseline created %d synthesis documents", synthesisCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.SearchUnfilteredUsed {
		failures = append(failures, "agent did not use unfiltered retrieval search")
	}
	if !turnMetrics.SearchPathFilterUsed {
		failures = append(failures, "agent did not use path-prefix retrieval search")
	}
	if !turnMetrics.SearchMetadataFilterUsed {
		failures = append(failures, "agent did not use metadata-filtered retrieval search")
	}

	assistantPass := metadataTopFound &&
		messageContainsAll(finalMessage, []string{ragCurrentPolicyPath, metadataTop.DocID, metadataTop.ChunkID}) &&
		messageContainsAny(finalMessage, []string{"json runner", "openclerk json runner"})
	if !assistantPass {
		failures = append(failures, "final answer did not cite active path, doc_id, chunk_id, and JSON runner policy")
	}
	databasePass := unfilteredTopFound &&
		pathTopFound &&
		metadataTopFound &&
		searchHitPath(unfilteredTop) == ragCurrentPolicyPath &&
		searchHitPath(pathTop) == ragCurrentPolicyPath &&
		searchHitPath(metadataTop) == ragCurrentPolicyPath &&
		!searchContainsPath(pathFiltered, ragArchivedPolicyPath) &&
		searchOnlyContainsPath(metadataFiltered, ragCurrentPolicyPath) &&
		repeatedTopFound &&
		metadataTop.DocID == repeatedTop.DocID &&
		metadataTop.ChunkID == repeatedTop.ChunkID &&
		searchHitHasCitation(metadataTop) &&
		synthesisCount == 0
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.SearchUnfilteredUsed &&
		turnMetrics.SearchPathFilterUsed &&
		turnMetrics.SearchMetadataFilterUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{ragCurrentPolicyPath, ragDecoyPolicyPath, ragArchivedPolicyPath},
	}, nil
}

func verifyDocsNavigationBaseline(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docsNavigationPrefix, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	indexDocID, indexFound := "", false
	policyFound := false
	onlyPrefix := true
	for _, doc := range list.Documents {
		if !strings.HasPrefix(doc.Path, docsNavigationPrefix) {
			onlyPrefix = false
		}
		switch doc.Path {
		case docsNavigationIndexPath:
			indexDocID = doc.DocID
			indexFound = true
		case docsNavigationPolicyPath:
			policyFound = true
		}
	}

	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasHeadings := got.Document != nil && containsAllStrings(got.Document.Headings, []string{"AgentOps Wiki Index", "Summary", "Links", "Limits"})

	links, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasOutgoing := links.Links != nil &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationPolicyPath) &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationArchPath) &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationOpsPath) &&
		documentLinksHaveCitations(links.Links.Outgoing)
	hasIncoming := links.Links != nil &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationPolicyPath) &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationArchPath) &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationOpsPath) &&
		documentLinksHaveCitations(links.Links.Incoming)

	graph, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraph,
		DocID:  indexDocID,
		Limit:  20,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasGraph := graph.Graph != nil &&
		graphContainsNodeLabels(graph.Graph.Nodes, []string{"AgentOps Wiki Index", "Runner Policy", "Knowledge Plane", "Runner Playbook"}) &&
		graphContainsLinkEdge(graph.Graph.Edges) &&
		graphEdgesHaveCitations(graph.Graph.Edges)

	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "graph",
			RefKind:    "document",
			RefID:      indexDocID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh" &&
		projections.Projections.Projections[0].Details["path"] == docsNavigationIndexPath

	failures := []string{}
	if !indexFound {
		failures = append(failures, "path-prefix listing did not find "+docsNavigationIndexPath)
	}
	if !policyFound {
		failures = append(failures, "path-prefix listing did not find "+docsNavigationPolicyPath)
	}
	if !onlyPrefix || len(list.Documents) != 2 {
		failures = append(failures, "path-prefix listing did not stay scoped to agentops directory")
	}
	if !hasHeadings {
		failures = append(failures, "get_document did not expose expected index headings")
	}
	if !hasOutgoing {
		failures = append(failures, "document_links missing cited outgoing links")
	}
	if !hasIncoming {
		failures = append(failures, "document_links missing cited incoming backlinks")
	}
	if !hasGraph {
		failures = append(failures, "graph_neighborhood missing cited nodes or edges")
	}
	if !hasProjection {
		failures = append(failures, "graph projection state missing or not fresh")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not use list_documents")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not use get_document")
	}
	if !turnMetrics.DocumentLinksUsed {
		failures = append(failures, "agent did not use document_links")
	}
	if !turnMetrics.GraphNeighborhoodUsed {
		failures = append(failures, "agent did not use graph_neighborhood")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect graph projection state")
	}

	assistantPass := messageContainsAny(finalMessage, []string{"directory", "folder", "path-prefix", "path prefix"}) &&
		messageContainsAny(finalMessage, []string{"link", "markdown"}) &&
		messageContainsAny(finalMessage, []string{"backlink", "incoming"}) &&
		messageContainsAny(finalMessage, []string{"graph neighborhood", "graph_neighborhood"}) &&
		messageContainsAny(finalMessage, []string{"sufficient", "enough"}) &&
		messageContainsAny(finalMessage, []string{"fails", "fail", "limits", "not enough"}) &&
		messageContainsAll(finalMessage, []string{docsNavigationIndexPath})
	if !assistantPass {
		failures = append(failures, "final answer did not compare directory, links/backlinks, graph neighborhood, limits, and source path")
	}

	databasePass := indexFound &&
		policyFound &&
		onlyPrefix &&
		len(list.Documents) == 2 &&
		hasHeadings &&
		hasOutgoing &&
		hasIncoming &&
		hasGraph &&
		hasProjection
	activityPass := turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.DocumentLinksUsed &&
		turnMetrics.GraphNeighborhoodUsed &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docsNavigationIndexPath, docsNavigationPolicyPath, docsNavigationArchPath, docsNavigationOpsPath},
	}, nil
}

func verifyGraphSemanticsReference(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: graphSemanticsSearchText, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: graphSemanticsPrefix, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}

	wantedPaths := []string{graphSemanticsIndexPath, graphSemanticsRoutingPath, graphSemanticsFreshnessPath, graphSemanticsOperationsPath}
	foundPaths := map[string]bool{}
	indexDocID := ""
	onlyPrefix := true
	for _, doc := range list.Documents {
		if !strings.HasPrefix(doc.Path, graphSemanticsPrefix) {
			onlyPrefix = false
		}
		foundPaths[doc.Path] = true
		if doc.Path == graphSemanticsIndexPath {
			indexDocID = doc.DocID
		}
	}

	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	if got.Document != nil {
		body = got.Document.Body
	}

	links, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasOutgoing := links.Links != nil &&
		documentLinksContainPath(links.Links.Outgoing, graphSemanticsRoutingPath) &&
		documentLinksContainPath(links.Links.Outgoing, graphSemanticsFreshnessPath) &&
		documentLinksContainPath(links.Links.Outgoing, graphSemanticsOperationsPath) &&
		documentLinksHaveCitations(links.Links.Outgoing)
	hasIncoming := links.Links != nil &&
		documentLinksContainPath(links.Links.Incoming, graphSemanticsRoutingPath) &&
		documentLinksContainPath(links.Links.Incoming, graphSemanticsFreshnessPath) &&
		documentLinksContainPath(links.Links.Incoming, graphSemanticsOperationsPath) &&
		documentLinksHaveCitations(links.Links.Incoming)

	graph, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraph,
		DocID:  indexDocID,
		Limit:  20,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasGraph := graph.Graph != nil &&
		graphContainsNodeLabels(graph.Graph.Nodes, []string{"Graph Semantics Reference", "Routing", "Freshness", "Operations"}) &&
		graphContainsStructuralEdge(graph.Graph.Edges) &&
		graphEdgesHaveCitations(graph.Graph.Edges) &&
		graphEdgesOnlyStructural(graph.Graph.Edges)

	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "graph",
			RefKind:    "document",
			RefID:      indexDocID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh" &&
		projections.Projections.Projections[0].Details["path"] == graphSemanticsIndexPath

	failures := []string{}
	if !searchContainsPath(search, graphSemanticsIndexPath) || !searchResultHasCitations(search) {
		failures = append(failures, "search did not expose cited canonical relationship text")
	}
	for _, path := range wantedPaths {
		if !foundPaths[path] {
			failures = append(failures, "path-prefix listing did not find "+path)
		}
	}
	if !onlyPrefix || len(list.Documents) != len(wantedPaths) {
		failures = append(failures, "path-prefix listing did not stay scoped to graph semantics fixture")
	}
	if !messageContainsAll(body, []string{"requires", "supersedes", "related to", "operationalizes"}) {
		failures = append(failures, "get_document did not expose expected relationship words")
	}
	if !hasOutgoing {
		failures = append(failures, "document_links missing cited outgoing relationships")
	}
	if !hasIncoming {
		failures = append(failures, "document_links missing cited incoming backlinks")
	}
	if !hasGraph {
		failures = append(failures, "graph_neighborhood missing cited structural graph context")
	}
	if !hasProjection {
		failures = append(failures, "graph projection state missing or not fresh")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not use list_documents")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not use get_document")
	}
	if !turnMetrics.DocumentLinksUsed {
		failures = append(failures, "agent did not use document_links")
	}
	if !turnMetrics.GraphNeighborhoodUsed {
		failures = append(failures, "agent did not use graph_neighborhood")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect graph projection state")
	}

	assistantPass := graphSemanticsReferenceAnswerPass(finalMessage)
	if !assistantPass {
		failures = append(failures, "final answer did not compare search, links/backlinks, graph neighborhood, markdown relationship text, and reference/defer decision")
	}

	databasePass := searchContainsPath(search, graphSemanticsIndexPath) &&
		searchResultHasCitations(search) &&
		allPathsFound(foundPaths, wantedPaths) &&
		onlyPrefix &&
		len(list.Documents) == len(wantedPaths) &&
		messageContainsAll(body, []string{"requires", "supersedes", "related to", "operationalizes"}) &&
		hasOutgoing &&
		hasIncoming &&
		hasGraph &&
		hasProjection
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.DocumentLinksUsed &&
		turnMetrics.GraphNeighborhoodUsed &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     wantedPaths,
	}, nil
}

func verifyMemoryRouterReference(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	sourceRefs := []string{
		memoryRouterSessionObservationPath,
		memoryRouterTemporalPath,
		memoryRouterFeedbackPath,
		memoryRouterRoutingPath,
	}
	body, found, err := documentBodyByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	sessionDocID, sessionFound, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		return verificationResult{}, err
	}
	temporalDocID, temporalFound, err := documentIDByPath(ctx, paths, memoryRouterTemporalPath)
	if err != nil {
		return verificationResult{}, err
	}
	feedbackDocID, feedbackFound, err := documentIDByPath(ctx, paths, memoryRouterFeedbackPath)
	if err != nil {
		return verificationResult{}, err
	}
	routingDocID, routingFound, err := documentIDByPath(ctx, paths, memoryRouterRoutingPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDocID, synthesisDocIDFound, err := documentIDByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   sessionDocID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, synthesisDocID)
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Temporal status: current canonical docs outrank stale session observations.",
		"Session promotion path: durable canonical markdown with source refs.",
		"Feedback weighting: advisory only.",
		"Routing choice: existing AgentOps document and retrieval actions.",
		"Decision: keep memory and autonomous routing as reference/deferred.",
		"## Sources",
		"## Freshness",
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+memoryRouterSynthesisPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", memoryRouterSynthesisPath, exactCount))
	}
	if !sessionFound {
		failures = append(failures, "missing "+memoryRouterSessionObservationPath)
	}
	if !temporalFound {
		failures = append(failures, "missing "+memoryRouterTemporalPath)
	}
	if !feedbackFound {
		failures = append(failures, "missing "+memoryRouterFeedbackPath)
	}
	if !routingFound {
		failures = append(failures, "missing "+memoryRouterRoutingPath)
	}
	if !synthesisDocIDFound {
		failures = append(failures, "missing document id for "+memoryRouterSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	hasProvenance := sessionFound && provenance.Provenance != nil && len(provenance.Provenance.Events) > 0
	if !hasProvenance {
		failures = append(failures, "session observation provenance missing")
	}
	hasProjection := projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterSessionObservationPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterTemporalPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterFeedbackPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterRoutingPath)
	if !hasProjection {
		failures = append(failures, "memory/router synthesis projection is not fresh with all source refs")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	listedMemoryRouterPrefix := containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{memoryRouterPrefix})
	if !turnMetrics.ListDocumentsUsed || !listedMemoryRouterPrefix {
		failures = append(failures, "agent did not list memory/router reference docs with path prefix")
	}
	requiredGetDocIDs := []string{sessionDocID, temporalDocID, feedbackDocID, routingDocID}
	gotMemoryRouterDocs := containsAllStrings(turnMetrics.GetDocumentDocIDs, requiredGetDocIDs)
	if !turnMetrics.GetDocumentUsed || !gotMemoryRouterDocs {
		failures = append(failures, "agent did not get every canonical memory/router doc")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection freshness")
	}
	if turnMetrics.BroadRepoSearch {
		failures = append(failures, "agent used broad repo search")
	}
	if turnMetrics.DirectSQLiteAccess {
		failures = append(failures, "agent used direct SQLite")
	}
	if turnMetrics.LegacyRunnerUsage {
		failures = append(failures, "agent used source-built or legacy runner path")
	}
	assistantPass := memoryRouterReferenceAnswerPass(finalMessage)
	if !assistantPass {
		failures = append(failures, "final answer did not explain temporal status, session promotion, feedback weighting, routing, source refs, freshness/provenance, and reference/defer decision")
	}

	databasePass := found &&
		exactCount == 1 &&
		sessionFound &&
		temporalFound &&
		feedbackFound &&
		routingFound &&
		synthesisDocIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		hasProvenance &&
		hasProjection
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		listedMemoryRouterPrefix &&
		turnMetrics.GetDocumentUsed &&
		gotMemoryRouterDocs &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.ProjectionStatesUsed &&
		!turnMetrics.BroadRepoSearch &&
		!turnMetrics.DirectSQLiteAccess &&
		!turnMetrics.LegacyRunnerUsage
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     append([]string{memoryRouterSynthesisPath}, sourceRefs...),
	}, nil
}

func verifyDocumentHistoryInspection(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	docID, found, err := documentIDByPath(ctx, paths, documentHistoryPolicyPath)
	if err != nil {
		return verificationResult{}, err
	}
	doc, _, err := documentByPath(ctx, paths, documentHistoryPolicyPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "document", RefID: docID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			RefKind: "document",
			RefID:   docID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasUpdatedBody := doc != nil && strings.Contains(doc.Body, "Current state: lifecycle inspection uses list_documents")
	hasProvenance := provenance.Provenance != nil &&
		eventTypesInclude(provenance.Provenance.Events, "document_created") &&
		eventTypesInclude(provenance.Provenance.Events, "document_updated")
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness != ""
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentHistoryPolicyPath)
	}
	if !hasUpdatedBody {
		failures = append(failures, "history inspection fixture did not expose updated lifecycle text")
	}
	if !hasProvenance {
		failures = append(failures, "document provenance missing created and updated events")
	}
	if !hasProjection {
		failures = append(failures, "document projection state missing or not fresh")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryPolicyPath}) &&
		messageContainsAny(finalMessage, []string{"provenance", "document_updated", "updated"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness", "fresh"}) &&
		messageContainsAny(finalMessage, []string{"existing", "current", "document and retrieval", "runner"})
	if !assistantPass {
		failures = append(failures, "final answer did not report history inspection, provenance, projection freshness, and existing runner workflow")
	}
	databasePass := found && hasUpdatedBody && hasProvenance && hasProjection
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryPolicyPath},
	}, nil
}

func verifyDocumentHistoryDiffReview(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	previous, previousFound, err := documentByPath(ctx, paths, documentHistoryDiffPreviousPath)
	if err != nil {
		return verificationResult{}, err
	}
	current, currentFound, err := documentByPath(ctx, paths, documentHistoryDiffCurrentPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !previousFound || previous == nil {
		failures = append(failures, "missing "+documentHistoryDiffPreviousPath)
	}
	if !currentFound || current == nil {
		failures = append(failures, "missing "+documentHistoryDiffCurrentPath)
	}
	if previous == nil || !strings.Contains(previous.Body, "optional review") {
		failures = append(failures, "previous evidence missing optional review text")
	}
	if current == nil || !strings.Contains(current.Body, "required review") {
		failures = append(failures, "current evidence missing required review text")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance")...)
	pathFailures := invalidRunnerPathFailures("list_documents path_prefix", turnMetrics.ListDocumentPathPrefixes)
	pathFailures = append(pathFailures, exactRunnerPathFailures("list_documents path_prefix", turnMetrics.ListDocumentPathPrefixes, documentHistoryDiffListPrefix)...)
	finalAnswerPathFailures := invalidRunnerPathTextFailures("final answer", finalMessage)
	failures = append(failures, pathFailures...)
	failures = append(failures, finalAnswerPathFailures...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryDiffPreviousPath, documentHistoryDiffCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"optional"}) &&
		messageContainsAny(finalMessage, []string{"required"}) &&
		messageContainsAny(finalMessage, []string{"citation", "cited", "source ref", "source_refs", "source"}) &&
		messageContainsAny(finalMessage, []string{"semantic", "summary"}) &&
		messageContainsAny(finalMessage, []string{"raw diff", "private diff", "do not expose raw", "no raw"}) &&
		len(finalAnswerPathFailures) == 0
	if !assistantPass {
		failures = append(failures, "final answer did not preserve cited semantic diff summary and raw-diff privacy handling")
	}
	databasePass := previousFound && currentFound &&
		previous != nil && current != nil &&
		strings.Contains(previous.Body, "optional review") &&
		strings.Contains(current.Body, "required review")
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 &&
		len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance")) == 0 &&
		len(pathFailures) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryDiffPreviousPath, documentHistoryDiffCurrentPath},
	}, nil
}

func verifyDocumentHistoryRestore(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	target, targetFound, err := documentByPath(ctx, paths, documentHistoryRestoreTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	targetID, _, err := documentIDByPath(ctx, paths, documentHistoryRestoreTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "document", RefID: targetID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			RefKind: "document",
			RefID:   targetID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	if target != nil {
		body = target.Body
	}
	restored := strings.Contains(body, "Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits.") &&
		!strings.Contains(body, "may bypass review")
	hasProvenance := provenance.Provenance != nil && eventTypesInclude(provenance.Provenance.Events, "document_updated")
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness != ""
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !targetFound {
		failures = append(failures, "missing "+documentHistoryRestoreTargetPath)
	}
	if !restored {
		failures = append(failures, "restore target was not restored to accepted lifecycle policy")
	}
	if !hasProvenance {
		failures = append(failures, "restore target provenance missing document update")
	}
	if !hasProjection {
		failures = append(failures, "restore target projection missing or not fresh")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryRestoreTargetPath, documentHistoryRestoreSourcePath}) &&
		messageContainsAny(finalMessage, []string{"restored", "restore", "rollback"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"source", "evidence", "citation"})
	if !assistantPass {
		failures = append(failures, "final answer did not report restore evidence, source, provenance, and projection freshness")
	}
	databasePass := targetFound && restored && hasProvenance && hasProjection
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryRestoreSourcePath, documentHistoryRestoreTargetPath},
	}, nil
}

func verifyDocumentHistoryPendingReview(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	targetBody, targetFound, err := documentBodyByPath(ctx, paths, documentHistoryPendingTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	proposalBody, proposalFound, err := documentBodyByPath(ctx, paths, documentHistoryPendingProposalPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !targetFound {
		failures = append(failures, "missing "+documentHistoryPendingTargetPath)
	}
	if !strings.Contains(targetBody, "Accepted lifecycle policy: source-sensitive durable edits require human review before acceptance.") ||
		strings.Contains(targetBody, "Auto-accept pending change") {
		failures = append(failures, "accepted target changed instead of remaining under review")
	}
	requiredProposal := []string{
		"type: review",
		"status: pending",
		"Review state: pending human review.",
		"Proposed change: Auto-accept pending change only after operator approval.",
		"Target document: notes/history-review/pending-target.md",
	}
	if !proposalFound {
		failures = append(failures, "missing "+documentHistoryPendingProposalPath)
	}
	failures = append(failures, missingRequired(proposalBody, requiredProposal)...)
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryPendingTargetPath, documentHistoryPendingProposalPath}) &&
		messageContainsAny(finalMessage, []string{"pending", "review"}) &&
		messageContainsAny(finalMessage, []string{"not accepted", "not become accepted", "did not change", "no accepted change"}) &&
		messageContainsAny(finalMessage, []string{"human", "operator"})
	if !assistantPass {
		failures = append(failures, "final answer did not report pending review state and unchanged accepted target")
	}
	databasePass := targetFound && proposalFound &&
		strings.Contains(targetBody, "Accepted lifecycle policy: source-sensitive durable edits require human review before acceptance.") &&
		!strings.Contains(targetBody, "Auto-accept pending change") &&
		len(missingRequired(proposalBody, requiredProposal)) == 0
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryPendingTargetPath, documentHistoryPendingProposalPath},
	}, nil
}

func verifyDocumentHistoryStaleSynthesis(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	synthesisID, synthesisFound, err := documentIDByPath(ctx, paths, documentHistoryStaleSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	currentID, currentFound, err := documentIDByPath(ctx, paths, documentHistoryStaleCurrentSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, synthesisID)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "source", RefID: currentID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projectionEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "projection", RefID: "synthesis:" + synthesisID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasProjection := projection != nil &&
		projection.Freshness == "stale" &&
		projectionDetailContains(projection.Details, "stale_source_refs", documentHistoryStaleCurrentSourcePath)
	hasSourceEvents := currentFound && sourceEvents.Provenance != nil &&
		eventTypesInclude(sourceEvents.Provenance.Events, "source_updated")
	hasInvalidation := projectionEvents.Provenance != nil &&
		eventTypesInclude(projectionEvents.Provenance.Events, "projection_invalidated")
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !synthesisFound {
		failures = append(failures, "missing "+documentHistoryStaleSynthesisPath)
	}
	if !currentFound {
		failures = append(failures, "missing "+documentHistoryStaleCurrentSourcePath)
	}
	if !hasProjection {
		failures = append(failures, "synthesis projection is not stale with current source ref")
	}
	if !hasSourceEvents {
		failures = append(failures, "current source provenance missing source update")
	}
	if !hasInvalidation {
		failures = append(failures, "synthesis projection invalidation event missing")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryStaleSynthesisPath, documentHistoryStaleCurrentSourcePath}) &&
		messageContainsAny(finalMessage, []string{"stale"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "invalidated", "source_updated", "updated"}) &&
		messageContainsAny(finalMessage, []string{"no repair", "not repair", "did not repair", "without repair"})
	if !assistantPass {
		failures = append(failures, "final answer did not report stale synthesis, provenance/invalidation, and no repair")
	}
	databasePass := synthesisFound && currentFound && hasProjection && hasSourceEvents && hasInvalidation
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryStaleSynthesisPath, documentHistoryStaleCurrentSourcePath, documentHistoryStaleOldSourcePath},
	}, nil
}

func documentHistoryInvariantFailures(turnMetrics metrics) []string {
	failures := []string{}
	if turnMetrics.BroadRepoSearch {
		failures = append(failures, "agent used broad repo search")
	}
	if turnMetrics.DirectSQLiteAccess {
		failures = append(failures, "agent used direct SQLite")
	}
	if turnMetrics.LegacyRunnerUsage {
		failures = append(failures, "agent used source-built or legacy runner path")
	}
	if turnMetrics.GeneratedFileInspection {
		failures = append(failures, "agent inspected generated files")
	}
	if turnMetrics.ModuleCacheInspection {
		failures = append(failures, "agent inspected module cache")
	}
	return failures
}

func invalidRunnerPathFailures(label string, values []string) []string {
	failures := []string{}
	for _, value := range values {
		if isInvalidRunnerPath(value) {
			failures = append(failures, label+" used non-vault-relative path "+value)
		}
	}
	return failures
}

func exactRunnerPathFailures(label string, values []string, allowed ...string) []string {
	failures := []string{}
	allowedSet := map[string]struct{}{}
	seen := map[string]bool{}
	for _, value := range allowed {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		allowedSet[trimmed] = struct{}{}
		seen[trimmed] = false
	}
	if len(values) == 0 {
		for value := range allowedSet {
			failures = append(failures, label+" missing required path "+value)
		}
		return failures
	}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if _, ok := allowedSet[trimmed]; ok {
			seen[trimmed] = true
			continue
		}
		failures = append(failures, label+" used unexpected path "+value)
	}
	for value, found := range seen {
		if !found {
			failures = append(failures, label+" missing required path "+value)
		}
	}
	return failures
}

func invalidRunnerPathTextFailures(label string, value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	lower := strings.ToLower(normalized)
	if strings.Contains(lower, ".openclerk-eval") ||
		strings.Contains(lower, "/vault/") ||
		strings.Contains(lower, "vault/") ||
		unixAbsolutePathPattern.MatchString(normalized) ||
		windowsDrivePathPattern.MatchString(trimmed) ||
		strings.Contains(trimmed, "\\") {
		return []string{label + " included non-vault-relative path text"}
	}
	return nil
}

func isInvalidRunnerPath(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	lower := strings.ToLower(normalized)
	if strings.Contains(lower, ".openclerk-eval") || strings.Contains(lower, "/vault/") || strings.HasPrefix(lower, "vault/") {
		return true
	}
	if strings.HasPrefix(normalized, "/") || strings.HasPrefix(normalized, "~") {
		return true
	}
	if len(trimmed) >= 3 && ((trimmed[0] >= 'A' && trimmed[0] <= 'Z') || (trimmed[0] >= 'a' && trimmed[0] <= 'z')) && trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/') {
		return true
	}
	return strings.Contains(trimmed, "\\")
}

func missingDocumentHistoryMetrics(turnMetrics metrics, required ...string) []string {
	failures := []string{}
	for _, requirement := range required {
		switch requirement {
		case "search":
			if !turnMetrics.SearchUsed {
				failures = append(failures, "agent did not use retrieval search")
			}
		case "list":
			if !turnMetrics.ListDocumentsUsed {
				failures = append(failures, "agent did not use list_documents")
			}
		case "get":
			if !turnMetrics.GetDocumentUsed {
				failures = append(failures, "agent did not use get_document")
			}
		case "provenance":
			if !turnMetrics.ProvenanceEventsUsed {
				failures = append(failures, "agent did not inspect provenance events")
			}
		case "projection":
			if !turnMetrics.ProjectionStatesUsed {
				failures = append(failures, "agent did not inspect projection states")
			}
		}
	}
	return failures
}

func verifyConfiguredLayoutScenario(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	layoutResult, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionInspectLayout})
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if layoutResult.Layout == nil {
		failures = append(failures, "inspect_layout returned no layout")
	} else if !layoutResult.Layout.Valid {
		failures = append(failures, "seeded configured layout was not valid")
	}
	if !turnMetrics.InspectLayoutUsed {
		failures = append(failures, "agent did not use inspect_layout")
	}
	if !messageContainsAll(finalMessage, []string{"convention", "sources/", "synthesis/", "source_refs"}) ||
		!messageContainsAny(finalMessage, []string{"no committed manifest", "no manifest", "config artifact required: false", "config_artifact_required false"}) {
		failures = append(failures, "answer did not explain convention-first layout and no-manifest decision")
	}
	if !messageReportsLayoutValid(finalMessage) {
		failures = append(failures, "answer did not report the layout as valid")
	}
	return verificationFromFailures(failures, "configured layout inspection passed", []string{"sources/layout-runner.md", "synthesis/layout-runner.md", "records/services/layout-runner.md"})
}

func verifyInvalidLayoutScenario(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	layoutResult, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionInspectLayout})
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if layoutResult.Layout == nil {
		failures = append(failures, "inspect_layout returned no layout")
	} else {
		if layoutResult.Layout.Valid {
			failures = append(failures, "seeded invalid layout was reported valid")
		}
		for _, id := range []string{"synthesis_source_refs_resolve", "synthesis_freshness_section", "service_identity_metadata"} {
			if !layoutChecksInclude(layoutResult.Layout.Checks, id, "fail") {
				failures = append(failures, "layout result missing failing check "+id)
			}
		}
	}
	if !turnMetrics.InspectLayoutUsed {
		failures = append(failures, "agent did not use inspect_layout")
	}
	if !messageContainsAll(finalMessage, []string{"synthesis/broken-layout.md", "records/services/broken-layout-service.md"}) ||
		!messageContainsAny(finalMessage, []string{"invalid", "valid: false", "valid false"}) ||
		!messageContainsAny(finalMessage, []string{"missing source", "missing_source_refs", "sources/missing-layout-source.md"}) ||
		!messageContainsAny(finalMessage, []string{"service_name", "service identity"}) ||
		!messageContainsAny(finalMessage, []string{"freshness", "## Freshness"}) {
		failures = append(failures, "answer did not report runner-visible invalid layout failures")
	}
	return verificationFromFailures(failures, "invalid layout inspection passed", []string{"synthesis/broken-layout.md", "records/services/broken-layout-service.md"})
}

func verifySourceURLUpdateDuplicateCreate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := sourceURLUpdateBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing original source URL document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "source_url:", "asset_path:", sourceURLUpdateAssetPath})...)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one original source document, got %d", sourceCount))
	}
	if duplicateCount != 0 {
		failures = append(failures, "duplicate create wrote "+sourceURLUpdateDuplicatePath)
	}
	if !turnMetrics.IngestSourceURLUsed || turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not attempt default create-mode source URL ingestion")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source documents after duplicate rejection")
	}
	assistantPass := messageContainsAll(finalMessage, []string{sourceURLUpdateSourcePath, sourceURLUpdateDuplicatePath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "already exists", "rejected"}) &&
		messageContainsAny(finalMessage, []string{"not created", "was not created", "no copy"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate rejection and no-write outcome")
	}
	databasePass := found && sourceCount == 1 && duplicateCount == 0 && doc != nil &&
		len(missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "source_url:", "asset_path:", sourceURLUpdateAssetPath})) == 0
	activityPass := len(sourceURLUpdateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUsed && !turnMetrics.IngestSourceURLUpdateUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceURLUpdateSourcePath},
	}, nil
}

func verifySourceURLUpdateSameSHA(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := sourceURLUpdateSourceEvents(ctx, paths, docIDOrEmpty(doc))
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDoc, synthesisFound, err := documentByPath(ctx, paths, sourceURLUpdateSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docIDOrEmpty(synthesisDoc))
	if err != nil {
		return verificationResult{}, err
	}
	projectionEvents, err := sourceURLUpdateProjectionEvents(ctx, paths, docIDOrEmpty(synthesisDoc))
	if err != nil {
		return verificationResult{}, err
	}
	search, err := sourceURLUpdateSearch(ctx, paths, sourceURLUpdateInitialText)
	if err != nil {
		return verificationResult{}, err
	}
	failures := sourceURLUpdateBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing source URL document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "asset_path:", sourceURLUpdateAssetPath})...)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one source document, got %d", sourceCount))
	}
	if eventTypesInclude(sourceEvents, "source_updated") {
		failures = append(failures, "same-SHA update emitted source_updated provenance")
	}
	if !synthesisFound || projection == nil || projection.Freshness != "fresh" {
		failures = append(failures, "same-SHA update did not leave dependent synthesis fresh")
	}
	if eventTypesInclude(projectionEvents, "projection_invalidated") {
		failures = append(failures, "same-SHA update invalidated dependent synthesis")
	}
	if !searchContainsPath(search, sourceURLUpdateSourcePath) || !searchResultHasCitations(search) {
		failures = append(failures, "same-SHA source evidence was not searchable with citations")
	}
	if !turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use source.mode update")
	}
	if !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProvenanceEventsUsed || !turnMetrics.SearchUsed || !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect source document, provenance, search evidence, and synthesis projection")
	}
	assistantPass := messageContainsAll(finalMessage, []string{sourceURLUpdateSourcePath, sourceURLUpdateSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"same-sha", "same sha", "no-op", "unchanged"}) &&
		messageContainsAny(finalMessage, []string{"citation", "source evidence", "preserved"}) &&
		messageContainsAny(finalMessage, []string{"fresh"}) &&
		messageContainsAny(finalMessage, []string{"no changed", "not changed", "no refresh", "not needed"})
	if !assistantPass {
		failures = append(failures, "final answer did not report same-SHA no-op with preserved evidence")
	}
	databasePass := found && doc != nil && sourceCount == 1 &&
		len(missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "asset_path:", sourceURLUpdateAssetPath})) == 0 &&
		!eventTypesInclude(sourceEvents, "source_updated") &&
		synthesisFound &&
		projection != nil &&
		projection.Freshness == "fresh" &&
		!eventTypesInclude(projectionEvents, "projection_invalidated") &&
		searchContainsPath(search, sourceURLUpdateSourcePath) &&
		searchResultHasCitations(search)
	activityPass := len(sourceURLUpdateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUpdateUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.SearchUsed &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceURLUpdateSourcePath, sourceURLUpdateSynthesisPath},
	}, nil
}

func verifySourceURLUpdateChangedPDF(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDoc, synthesisFound, err := documentByPath(ctx, paths, sourceURLUpdateSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := sourceURLUpdateSourceEvents(ctx, paths, docIDOrEmpty(doc))
	if err != nil {
		return verificationResult{}, err
	}
	changedSearch, err := sourceURLUpdateSearch(ctx, paths, sourceURLUpdateChangedText)
	if err != nil {
		return verificationResult{}, err
	}
	oldSearch, err := sourceURLUpdateSearch(ctx, paths, sourceURLUpdateInitialText)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docIDOrEmpty(synthesisDoc))
	if err != nil {
		return verificationResult{}, err
	}
	projectionEvents, err := sourceURLUpdateProjectionEvents(ctx, paths, docIDOrEmpty(synthesisDoc))
	if err != nil {
		return verificationResult{}, err
	}
	updateEventOK := sourceURLUpdateEventHasSHAChange(sourceEvents)
	hasStaleProjection := projection != nil &&
		projection.Freshness == "stale" &&
		projectionDetailContains(projection.Details, "stale_source_refs", sourceURLUpdateSourcePath)
	failures := sourceURLUpdateBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing updated source URL document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{sourceURLUpdateChangedText, "asset_path:", sourceURLUpdateAssetPath})...)
		failures = append(failures, presentForbidden(doc.Body, []string{sourceURLUpdateInitialText})...)
	}
	if !synthesisFound || synthesisDoc == nil {
		failures = append(failures, "missing dependent synthesis")
	} else if !strings.Contains(synthesisDoc.Body, sourceURLUpdateInitialText) {
		failures = append(failures, "dependent synthesis was repaired or no longer contains initial stale claim")
	}
	if !searchContainsPath(changedSearch, sourceURLUpdateSourcePath) || !searchResultHasCitations(changedSearch) {
		failures = append(failures, "changed source evidence was not searchable with citations")
	}
	if searchContainsPath(oldSearch, sourceURLUpdateSourcePath) {
		failures = append(failures, "old source evidence remained indexed for the source path")
	}
	if !updateEventOK {
		failures = append(failures, "source update provenance missing previous/new SHA details")
	}
	if !hasStaleProjection {
		failures = append(failures, "dependent synthesis projection is not visibly stale")
	}
	if !eventTypesInclude(projectionEvents, "projection_invalidated") {
		failures = append(failures, "synthesis projection invalidation event missing")
	}
	if !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProjectionStatesUsed || !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not use search, source/synthesis listing/get, projection, and provenance workflow")
	}
	assistantPass := messageContainsAll(finalMessage, []string{sourceURLUpdateSourcePath, sourceURLUpdateSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"changed-pdf", "changed pdf", "updated pdf", "changed"}) &&
		messageContainsAny(finalMessage, []string{"stale"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source_updated", "source update"}) &&
		messageContainsAny(finalMessage, []string{"citation", "evidence"})
	if !assistantPass {
		failures = append(failures, "final answer did not report changed update, stale projection, provenance, and citations")
	}
	databasePass := found && doc != nil && synthesisFound && synthesisDoc != nil &&
		len(missingRequired(doc.Body, []string{sourceURLUpdateChangedText, "asset_path:", sourceURLUpdateAssetPath})) == 0 &&
		len(presentForbidden(doc.Body, []string{sourceURLUpdateInitialText})) == 0 &&
		strings.Contains(synthesisDoc.Body, sourceURLUpdateInitialText) &&
		searchContainsPath(changedSearch, sourceURLUpdateSourcePath) &&
		searchResultHasCitations(changedSearch) &&
		!searchContainsPath(oldSearch, sourceURLUpdateSourcePath) &&
		updateEventOK &&
		hasStaleProjection &&
		eventTypesInclude(projectionEvents, "projection_invalidated")
	activityPass := len(sourceURLUpdateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceURLUpdateSourcePath, sourceURLUpdateSynthesisPath},
	}, nil
}

func verifySourceURLUpdateConflict(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	conflictCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateConflictPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := sourceURLUpdateSourceEvents(ctx, paths, docIDOrEmpty(doc))
	if err != nil {
		return verificationResult{}, err
	}
	failures := sourceURLUpdateBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing original source URL document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "asset_path:", sourceURLUpdateAssetPath})...)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one original source document, got %d", sourceCount))
	}
	if conflictCount != 0 {
		failures = append(failures, "conflict update wrote "+sourceURLUpdateConflictPath)
	}
	if eventTypesInclude(sourceEvents, "source_updated") {
		failures = append(failures, "conflict update emitted source_updated provenance")
	}
	if !turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use source.mode update")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source documents after conflict")
	}
	assistantPass := messageContainsAll(finalMessage, []string{sourceURLUpdateSourcePath}) &&
		messageContainsAny(finalMessage, []string{sourceURLUpdateConflictPath, "source-url-update-conflict.md"}) &&
		messageContainsAny(finalMessage, []string{"conflict", "mismatch", "path hint", "path-hint"}) &&
		messageContainsAny(finalMessage, []string{"not created", "was not created", "no write", "without writing"})
	if !assistantPass {
		failures = append(failures, "final answer did not report path-hint conflict and no-write outcome")
	}
	databasePass := found && doc != nil && sourceCount == 1 && conflictCount == 0 &&
		len(missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "asset_path:", sourceURLUpdateAssetPath})) == 0 &&
		!eventTypesInclude(sourceEvents, "source_updated")
	activityPass := len(sourceURLUpdateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUpdateUsed &&
		turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceURLUpdateSourcePath},
	}, nil
}

func sourceURLUpdateBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}

func docIDOrEmpty(doc *runner.Document) string {
	if doc == nil {
		return ""
	}
	return doc.DocID
}

func sourceURLUpdateSourceEvents(ctx context.Context, paths evalPaths, docID string) ([]runner.ProvenanceEvent, error) {
	if strings.TrimSpace(docID) == "" {
		return nil, nil
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "source", RefID: docID, Limit: 20},
	})
	if err != nil || result.Provenance == nil {
		return nil, err
	}
	return result.Provenance.Events, nil
}

func sourceURLUpdateProjectionEvents(ctx context.Context, paths evalPaths, synthesisDocID string) ([]runner.ProvenanceEvent, error) {
	if strings.TrimSpace(synthesisDocID) == "" {
		return nil, nil
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "projection", RefID: "synthesis:" + synthesisDocID, Limit: 20},
	})
	if err != nil || result.Provenance == nil {
		return nil, err
	}
	return result.Provenance.Events, nil
}

func sourceURLUpdateSearch(ctx context.Context, paths evalPaths, text string) (runner.RetrievalTaskResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	return runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       text,
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
}

func sourceURLUpdateEventHasSHAChange(events []runner.ProvenanceEvent) bool {
	for _, event := range events {
		if event.EventType != "source_updated" {
			continue
		}
		previous := strings.TrimSpace(event.Details["previous_sha256"])
		next := strings.TrimSpace(event.Details["new_sha256"])
		if previous != "" && next != "" && previous != next &&
			event.Details["asset_path"] == sourceURLUpdateAssetPath {
			return true
		}
	}
	return false
}

func verifyStaleSynthesisUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	docPath := "synthesis/runner-routing.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	createdCurrent, err := exactDocumentCount(ctx, paths, "synthesis/runner-routing-current.md")
	if err != nil {
		return verificationResult{}, err
	}
	createdUpdated, err := exactDocumentCount(ctx, paths, "synthesis/runner-routing-updated.md")
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", docPath, exactCount))
	}
	if createdCurrent != 0 || createdUpdated != 0 {
		failures = append(failures, "created duplicate synthesis path")
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current guidance: routine agents must use openclerk JSON runner",
		"Current source: sources/runner-current-runner.md",
		"Supersedes: sources/runner-old-workaround.md",
		"## Sources",
		"## Freshness",
	}
	sourceRefs := []string{"sources/runner-current-runner.md", "sources/runner-old-workaround.md"}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	failures = append(failures, presentForbidden(body, []string{"may bypass OpenClerk runner through a temporary command-path workaround"})...)
	if !containsAny(strings.ToLower(body), []string{"stale", "supersedes", "superseded", "contradiction", "current guidance"}) {
		failures = append(failures, "missing stale or supersession language")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	assistantPass := messageContainsAll(finalMessage, []string{docPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "current", "supersedes", "stale"})
	if !assistantPass {
		failures = append(failures, "final answer did not describe the synthesis update")
	}
	databasePass := found && exactCount == 1 && createdCurrent == 0 && createdUpdated == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		len(presentForbidden(body, []string{"may bypass OpenClerk runner through a temporary command-path workaround"})) == 0 &&
		containsAny(strings.ToLower(body), []string{"stale", "supersedes", "superseded", "contradiction", "current guidance"})
	activityPass := turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func verifySynthesisFreshnessRepair(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	docPath := "synthesis/runner-repair.md"
	currentSource := "sources/repair-current.md"
	supersededSource := "sources/repair-old.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      docID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	events, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + docID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", docPath, exactCount))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+docPath)
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"source_refs: sources/repair-current.md, sources/repair-old.md",
		currentSource,
		supersededSource,
		"## Sources",
		"## Freshness",
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, presentForbidden(body, []string{"may use a temporary command-path workaround"})...)
	hasProjection := false
	hasCurrent := false
	hasSuperseded := false
	if projections.Projections != nil && len(projections.Projections.Projections) == 1 {
		projection := projections.Projections.Projections[0]
		hasProjection = projection.Freshness == "fresh"
		hasCurrent = projection.Details["current_source_refs"] == currentSource
		hasSuperseded = projection.Details["superseded_source_refs"] == supersededSource
	}
	if !hasProjection {
		failures = append(failures, "synthesis projection is not fresh")
	}
	if !hasCurrent {
		failures = append(failures, "synthesis projection missing current source ref")
	}
	if !hasSuperseded {
		failures = append(failures, "synthesis projection missing superseded source ref")
	}
	hasInvalidation := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_invalidated")
	hasRefresh := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_refreshed")
	if !hasInvalidation {
		failures = append(failures, "synthesis invalidation event missing")
	}
	if !hasRefresh {
		failures = append(failures, "synthesis refresh event missing")
	}
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.ProjectionStatesUsed
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	assistantPass := messageContainsAll(finalMessage, []string{docPath, currentSource, supersededSource}) &&
		messageContainsAny(finalMessage, []string{"fresh", "freshness", "current", "superseded"})
	if !assistantPass {
		failures = append(failures, "final answer did not mention repaired freshness and source status")
	}
	databasePass := found &&
		exactCount == 1 &&
		len(missingRequired(body, required)) == 0 &&
		len(presentForbidden(body, []string{"may use a temporary command-path workaround"})) == 0 &&
		hasProjection &&
		hasCurrent &&
		hasSuperseded &&
		hasInvalidation &&
		hasRefresh
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath, currentSource, supersededSource},
	}, nil
}

func verifySynthesisCandidatePressure(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, synthesisCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, synthesisCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, synthesisCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docID)
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current compiler decision: existing document and retrieval actions are sufficient for synthesis compiler pressure repairs",
		"Current source: " + synthesisCandidateCurrentSrc,
		"Superseded source: " + synthesisCandidateOldSrc,
		"## Sources",
		"## Freshness",
	}
	sourceRefs := []string{synthesisCandidateCurrentSrc, synthesisCandidateOldSrc}
	forbidden := []string{"require a dedicated compile_synthesis runner action", "requires a dedicated compile_synthesis runner action"}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+synthesisCandidatePath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", synthesisCandidatePath, exactCount))
	}
	if synthesisCount != 2 {
		failures = append(failures, fmt.Sprintf("expected exactly target and decoy synthesis documents, got %d", synthesisCount))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+synthesisCandidatePath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	failures = append(failures, presentForbidden(body, forbidden)...)
	if projection == nil || projection.Freshness != "fresh" {
		failures = append(failures, "synthesis projection is not fresh")
	} else {
		if !projectionDetailContains(projection.Details, "current_source_refs", synthesisCandidateCurrentSrc) {
			failures = append(failures, "synthesis projection missing current compiler source")
		}
		if !projectionDetailContains(projection.Details, "superseded_source_refs", synthesisCandidateOldSrc) {
			failures = append(failures, "synthesis projection missing superseded compiler source")
		}
	}
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect synthesis projection freshness")
	}
	assistantPass := messageContainsAll(finalMessage, []string{synthesisCandidatePath, synthesisCandidateCurrentSrc}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "fresh", "freshness", "existing actions"})
	if !assistantPass {
		failures = append(failures, "final answer did not report target update and current source")
	}
	databasePass := found &&
		exactCount == 1 &&
		synthesisCount == 2 &&
		docIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		len(presentForbidden(body, forbidden)) == 0 &&
		projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", synthesisCandidateCurrentSrc) &&
		projectionDetailContains(projection.Details, "superseded_source_refs", synthesisCandidateOldSrc)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{synthesisCandidatePath, synthesisCandidateDecoyPath, synthesisCandidateCurrentSrc, synthesisCandidateOldSrc},
	}, nil
}

func verifySynthesisSourceSetPressure(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, synthesisSourceSetPath, finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:                 []string{sourceSetAlphaPath, sourceSetBetaPath, sourceSetGammaPath},
		RequireSearch:              true,
		RequireList:                true,
		Metrics:                    turnMetrics,
		FinalAnswerPath:            true,
		AdditionalDocs:             []string{sourceSetAlphaPath, sourceSetBetaPath, sourceSetGammaPath},
		AdditionalBodyRequirements: []string{"alpha", "beta", "gamma", "source refs", "freshness"},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if synthesisCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one synthesis document, got %d", synthesisCount))
	}
	databasePass := base.DatabasePass && synthesisCount == 1
	return verificationResult{
		Passed:        databasePass && base.AssistantPass,
		DatabasePass:  databasePass,
		AssistantPass: base.AssistantPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyMTSynthesisDriftPressure(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, mtDriftSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	currentBody, currentFound, err := documentBodyByPath(ctx, paths, mtDriftCurrentPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, mtDriftSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, mtDriftSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docID)
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current drift decision: keep existing document and retrieval actions",
		"Current source: " + mtDriftCurrentPath,
		"Superseded source: " + mtDriftOldSourcePath,
		"## Sources",
		"## Freshness",
	}
	sourceRefs := []string{mtDriftCurrentPath, mtDriftOldSourcePath}
	forbidden := []string{"promoted immediately", "dedicated compile_synthesis action is required"}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+mtDriftSynthesisPath)
	}
	if !currentFound {
		failures = append(failures, "missing "+mtDriftCurrentPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", mtDriftSynthesisPath, exactCount))
	}
	if synthesisCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one drift synthesis document, got %d", synthesisCount))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+mtDriftSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	failures = append(failures, presentForbidden(body, forbidden)...)
	if !strings.Contains(currentBody, "Current drift decision says existing document and retrieval actions should stay the v1 synthesis path.") {
		failures = append(failures, "current drift source was not updated")
	}
	if projection == nil || projection.Freshness != "fresh" {
		failures = append(failures, "drift synthesis projection is not fresh")
	} else {
		if !projectionDetailContains(projection.Details, "current_source_refs", mtDriftCurrentPath) {
			failures = append(failures, "drift synthesis projection missing current source")
		}
		if !projectionDetailContains(projection.Details, "superseded_source_refs", mtDriftOldSourcePath) {
			failures = append(failures, "drift synthesis projection missing superseded source")
		}
	}
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect synthesis projection freshness")
	}
	assistantPass := messageContainsAll(finalMessage, []string{mtDriftSynthesisPath, mtDriftCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "fresh", "current"})
	if !assistantPass {
		failures = append(failures, "final answer did not report drift repair and current source")
	}
	databasePass := found &&
		currentFound &&
		exactCount == 1 &&
		synthesisCount == 1 &&
		docIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		len(presentForbidden(body, forbidden)) == 0 &&
		strings.Contains(currentBody, "Current drift decision says existing document and retrieval actions should stay the v1 synthesis path.") &&
		projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", mtDriftCurrentPath) &&
		projectionDetailContains(projection.Details, "superseded_source_refs", mtDriftOldSourcePath)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{mtDriftSynthesisPath, mtDriftCurrentPath, mtDriftOldSourcePath},
	}, nil
}

func verifyDuplicatePathReject(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	bodyCheck, err := verifyDocumentContains(ctx, paths, "notes/projects/duplicate.md", []string{"This canonical path already exists."}, []string{"overwritten"})
	if err != nil {
		return verificationResult{}, err
	}
	answerPass := isDuplicateRejection(finalMessage)
	failures := []string{}
	if !bodyCheck.DatabasePass {
		failures = append(failures, bodyCheck.Details)
	}
	if !answerPass {
		failures = append(failures, "answer did not report the duplicate path failure")
	}
	return verificationResult{
		Passed:        bodyCheck.DatabasePass && answerPass,
		DatabasePass:  bodyCheck.DatabasePass,
		AssistantPass: answerPass,
		Details:       missingDetails(failures),
		Documents:     []string{"notes/projects/duplicate.md"},
	}, nil
}

func isDuplicateRejection(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "already exists") || strings.Contains(lower, "duplicate") {
		return true
	}
	return strings.Contains(lower, "exists") && containsAny(lower, []string{"cannot", "can't", "failed", "not overwrite", "won't overwrite", "did not overwrite"})
}

func verifyPromotedRecordVsDocs(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: "Plain docs evidence", PathPrefix: "notes/reference/", Limit: 5},
	})
	if err != nil {
		return verificationResult{}, err
	}
	services, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionServicesLookup,
		Services: runner.ServiceLookupOptions{
			Text:      "OpenClerk runner",
			Interface: "JSON runner",
			Limit:     5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "services",
			RefKind:    "service",
			RefID:      "openclerk-runner",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	events, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "service",
			RefID:   "openclerk-runner",
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasPlainDoc := false
	if search.Search != nil {
		for _, hit := range search.Search.Hits {
			if hit.DocID != "" && hit.Title != "" && containsAny(strings.ToLower(hit.Snippet), []string{"plain docs evidence", "production service"}) {
				hasPlainDoc = true
				break
			}
		}
	}
	hasService := false
	if services.Services != nil {
		for _, service := range services.Services.Services {
			if service.ServiceID != "openclerk-runner" {
				continue
			}
			if service.Interface == "JSON runner" && len(service.Citations) > 0 {
				hasService = true
				break
			}
		}
	}
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh"
	hasProvenance := events.Provenance != nil && len(events.Provenance.Events) > 0
	assistantPass := messageContainsAny(finalMessage, []string{"services lookup", "services_lookup", "service registry"}) &&
		messageContainsAny(finalMessage, []string{"plain docs", "plain doc", "search"}) &&
		messageContainsAny(finalMessage, []string{"json runner", "runner"})
	activityPass := turnMetrics.ToolCalls >= 2 && turnMetrics.CommandExecutions >= 2
	failures := []string{}
	if !hasPlainDoc {
		failures = append(failures, "plain docs search evidence missing")
	}
	if !hasService {
		failures = append(failures, "services lookup evidence missing")
	}
	if !hasProjection {
		failures = append(failures, "services projection state missing")
	}
	if !hasProvenance {
		failures = append(failures, "services provenance missing")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not compare services lookup with plain docs")
	}
	if !activityPass {
		failures = append(failures, fmt.Sprintf("expected at least two agent operations for search and services lookup, got tools=%d commands=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions))
	}
	return verificationResult{
		Passed:        hasPlainDoc && hasService && hasProjection && hasProvenance && assistantPass && activityPass,
		DatabasePass:  hasPlainDoc && hasService && hasProjection && hasProvenance,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{"notes/reference/runner-service.md", "records/services/openclerk-runner.md"},
	}, nil
}

func verifyDecisionRecordVsDocs(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: "OpenClerk runner decisions", PathPrefix: "notes/reference/", Limit: 5},
	})
	if err != nil {
		return verificationResult{}, err
	}
	decisions, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{
			Text:   "JSON runner",
			Status: "accepted",
			Scope:  "agentops",
			Owner:  "platform",
			Limit:  5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-runner-current",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	hasPlainDoc := search.Search != nil && len(search.Search.Hits) > 0
	hasDecision := false
	if decisions.Decisions != nil {
		for _, decision := range decisions.Decisions.Decisions {
			if decision.DecisionID == "adr-runner-current" &&
				decision.Status == "accepted" &&
				decision.Scope == "agentops" &&
				decision.Owner == "platform" &&
				len(decision.Citations) > 0 {
				hasDecision = true
				break
			}
		}
	}
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh"
	hasCitationPath := messageContainsAny(finalMessage, []string{"docs/architecture/runner-current-decision.md"})
	assistantPass := messageContainsAny(finalMessage, []string{"decisions lookup", "decisions_lookup", "decision records"}) &&
		messageContainsAny(finalMessage, []string{"plain docs", "plain doc", "search"}) &&
		messageContainsAny(finalMessage, []string{"status", "scope", "accepted", "agentops"}) &&
		hasCitationPath
	activityPass := turnMetrics.SearchUsed && turnMetrics.DecisionsLookupUsed
	failures := []string{}
	if !hasPlainDoc {
		failures = append(failures, "plain docs search evidence missing")
	}
	if !hasDecision {
		failures = append(failures, "decisions lookup evidence missing")
	}
	if !hasProjection {
		failures = append(failures, "decision projection freshness missing")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use plain docs search")
	}
	if !turnMetrics.DecisionsLookupUsed {
		failures = append(failures, "agent did not use decisions lookup")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not compare decisions lookup with plain docs")
	}
	if !hasCitationPath {
		failures = append(failures, "final answer did not include decision citation path")
	}
	return verificationResult{
		Passed:        hasPlainDoc && hasDecision && hasProjection && assistantPass && activityPass,
		DatabasePass:  hasPlainDoc && hasDecision && hasProjection,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{"notes/reference/runner-decision-narrative.md", "docs/architecture/runner-current-decision.md", "records/decisions/runner-old-decision.md"},
	}, nil
}

func verifyDecisionSupersessionFreshness(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	oldDecision, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: "adr-runner-old",
	})
	if err != nil {
		return verificationResult{}, err
	}
	currentDecision, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: "adr-runner-current",
	})
	if err != nil {
		return verificationResult{}, err
	}
	oldProjection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-runner-old",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	currentProjection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-runner-current",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	events, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "decisions:adr-runner-current",
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	hasOldDecision := oldDecision.Decision != nil &&
		oldDecision.Decision.Status == "superseded" &&
		len(oldDecision.Decision.SupersededBy) == 1 &&
		oldDecision.Decision.SupersededBy[0] == "adr-runner-current" &&
		len(oldDecision.Decision.Citations) > 0
	hasCurrentDecision := currentDecision.Decision != nil &&
		currentDecision.Decision.Status == "accepted" &&
		len(currentDecision.Decision.Supersedes) == 1 &&
		currentDecision.Decision.Supersedes[0] == "adr-runner-old" &&
		len(currentDecision.Decision.Citations) > 0
	hasOldProjection := oldProjection.Projections != nil &&
		len(oldProjection.Projections.Projections) == 1 &&
		oldProjection.Projections.Projections[0].Freshness == "stale" &&
		oldProjection.Projections.Projections[0].Details["superseded_by"] == "adr-runner-current"
	hasCurrentProjection := currentProjection.Projections != nil &&
		len(currentProjection.Projections.Projections) == 1 &&
		currentProjection.Projections.Projections[0].Freshness == "fresh"
	hasProvenance := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_refreshed")
	hasCitationPaths := messageContainsAll(finalMessage, []string{
		"docs/architecture/runner-old-decision.md",
		"records/decisions/runner-current-decision.md",
	})
	assistantPass := messageContainsAny(finalMessage, []string{"superseded", "supersedes"}) &&
		messageContainsAny(finalMessage, []string{"stale"}) &&
		messageContainsAny(finalMessage, []string{"fresh"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "projection"}) &&
		hasCitationPaths
	inspectedDecisionRecords := decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, "adr-runner-old", "adr-runner-current")
	activityPass := inspectedDecisionRecords && turnMetrics.ProjectionStatesUsed && turnMetrics.ProvenanceEventsUsed
	failures := []string{}
	if !hasOldDecision {
		failures = append(failures, "old superseded decision detail missing")
	}
	if !hasCurrentDecision {
		failures = append(failures, "current replacement decision detail missing")
	}
	if !hasOldProjection {
		failures = append(failures, "old decision stale projection missing")
	}
	if !hasCurrentProjection {
		failures = append(failures, "current decision fresh projection missing")
	}
	if !hasProvenance {
		failures = append(failures, "decision projection provenance missing")
	}
	if !inspectedDecisionRecords {
		failures = append(failures, "agent did not use decision_record for adr-runner-old and adr-runner-current")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection_states")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance_events")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report supersession freshness")
	}
	if !hasCitationPaths {
		failures = append(failures, "final answer did not include decision citation paths")
	}
	return verificationResult{
		Passed:        hasOldDecision && hasCurrentDecision && hasOldProjection && hasCurrentProjection && hasProvenance && assistantPass && activityPass,
		DatabasePass:  hasOldDecision && hasCurrentDecision && hasOldProjection && hasCurrentProjection && hasProvenance,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{"docs/architecture/runner-old-decision.md", "records/decisions/runner-current-decision.md"},
	}, nil
}

func verifyDecisionRealADRMigration(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	lookup, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{
			Text:   "knowledge-configuration",
			Status: "accepted",
			Owner:  "platform",
			Limit:  5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	agentOpsDecision, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: "adr-agentops-only-knowledge-plane",
	})
	if err != nil {
		return verificationResult{}, err
	}
	configProjection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-knowledge-configuration-v1",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	agentOpsProjection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-agentops-only-knowledge-plane",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "decisions:adr-knowledge-configuration-v1",
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	hasConfigDecision := false
	if lookup.Decisions != nil {
		for _, decision := range lookup.Decisions.Decisions {
			if decision.DecisionID == "adr-knowledge-configuration-v1" &&
				decision.Status == "accepted" &&
				decision.Scope == "knowledge-configuration" &&
				decision.Owner == "platform" &&
				len(decision.Supersedes) == 1 &&
				decision.Supersedes[0] == "adr-agentops-only-knowledge-plane" &&
				len(decision.Citations) > 0 &&
				decision.Citations[0].Path == "docs/architecture/knowledge-configuration-v1-adr.md" {
				hasConfigDecision = true
				break
			}
		}
	}
	hasAgentOpsDecision := agentOpsDecision.Decision != nil &&
		agentOpsDecision.Decision.DecisionID == "adr-agentops-only-knowledge-plane" &&
		agentOpsDecision.Decision.Status == "accepted" &&
		agentOpsDecision.Decision.Scope == "knowledge-plane" &&
		len(agentOpsDecision.Decision.SourceRefs) == 1 &&
		agentOpsDecision.Decision.SourceRefs[0] == "sources/agentops-direction.md" &&
		len(agentOpsDecision.Decision.Citations) > 0 &&
		agentOpsDecision.Decision.Citations[0].Path == "docs/architecture/eval-backed-knowledge-plane-adr.md"
	hasConfigProjection := configProjection.Projections != nil &&
		len(configProjection.Projections.Projections) == 1 &&
		configProjection.Projections.Projections[0].Freshness == "fresh" &&
		configProjection.Projections.Projections[0].Details["path"] == "docs/architecture/knowledge-configuration-v1-adr.md"
	hasAgentOpsProjection := agentOpsProjection.Projections != nil &&
		len(agentOpsProjection.Projections.Projections) == 1 &&
		agentOpsProjection.Projections.Projections[0].Freshness == "fresh" &&
		agentOpsProjection.Projections.Projections[0].Details["path"] == "docs/architecture/eval-backed-knowledge-plane-adr.md"
	hasProvenance := provenance.Provenance != nil && eventTypesInclude(provenance.Provenance.Events, "projection_refreshed")
	hasCitationPaths := messageContainsAll(finalMessage, []string{
		"docs/architecture/eval-backed-knowledge-plane-adr.md",
		"docs/architecture/knowledge-configuration-v1-adr.md",
	})
	assistantPass := messageContainsAny(finalMessage, []string{"canonical markdown", "canonical adr", "authoritative"}) &&
		messageContainsAny(finalMessage, []string{"decisions_lookup", "decisions lookup", "decision lookup", "decision records"}) &&
		messageContainsAny(finalMessage, []string{"decision_record", "decision record", "adr record", "decision records"}) &&
		messageContainsAny(finalMessage, []string{"fresh"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "projection"}) &&
		hasCitationPaths
	inspectedAgentOpsDecision := decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, "adr-agentops-only-knowledge-plane")
	activityPass := turnMetrics.DecisionsLookupUsed && inspectedAgentOpsDecision && turnMetrics.ProjectionStatesUsed && turnMetrics.ProvenanceEventsUsed
	failures := []string{}
	if !hasConfigDecision {
		failures = append(failures, "knowledge configuration ADR decision lookup missing")
	}
	if !hasAgentOpsDecision {
		failures = append(failures, "agentops ADR decision detail missing")
	}
	if !hasConfigProjection {
		failures = append(failures, "knowledge configuration ADR fresh projection missing")
	}
	if !hasAgentOpsProjection {
		failures = append(failures, "agentops ADR fresh projection missing")
	}
	if !hasProvenance {
		failures = append(failures, "decision projection provenance missing")
	}
	if !turnMetrics.DecisionsLookupUsed {
		failures = append(failures, "agent did not use decisions_lookup")
	}
	if !inspectedAgentOpsDecision {
		failures = append(failures, "agent did not use decision_record for adr-agentops-only-knowledge-plane")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection_states")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance_events")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report ADR decision migration evidence")
	}
	if !hasCitationPaths {
		failures = append(failures, "final answer did not include ADR citation paths")
	}
	return verificationResult{
		Passed:        hasConfigDecision && hasAgentOpsDecision && hasConfigProjection && hasAgentOpsProjection && hasProvenance && assistantPass && activityPass,
		DatabasePass:  hasConfigDecision && hasAgentOpsDecision && hasConfigProjection && hasAgentOpsProjection && hasProvenance,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{"docs/architecture/eval-backed-knowledge-plane-adr.md", "docs/architecture/knowledge-configuration-v1-adr.md"},
	}, nil
}

func verifySourceSensitiveAuditRepair(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, sourceAuditSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, sourceAuditSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicatePaths, err := disallowedDocumentPathsWithPrefix(ctx, paths, "synthesis/", map[string]bool{
		sourceAuditSynthesisPath: true,
		sourceAuditDecoyPath:     true,
	})
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, sourceAuditSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docID)
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	events, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + docID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"source_refs: " + sourceAuditCurrentSourcePath + ", " + sourceAuditOldSourcePath,
		"Current audit guidance: use the installed openclerk JSON runner",
		"Current source: " + sourceAuditCurrentSourcePath,
		"Superseded source: " + sourceAuditOldSourcePath,
		"## Sources",
		"## Freshness",
	}
	forbidden := []string{"prefer a legacy command-path workaround for runner audit repairs"}
	hasProjection := projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", sourceAuditCurrentSourcePath) &&
		projectionDetailContains(projection.Details, "superseded_source_refs", sourceAuditOldSourcePath)
	hasInvalidation := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_invalidated")
	hasRefresh := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_refreshed")
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	assistantPass := messageContainsAll(finalMessage, []string{sourceAuditSynthesisPath, sourceAuditCurrentSourcePath}) &&
		messageContainsAny(finalMessage, []string{"fresh", "freshness", "current", "superseded"})

	failures := []string{}
	if !found {
		failures = append(failures, "missing "+sourceAuditSynthesisPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", sourceAuditSynthesisPath, exactCount))
	}
	if len(duplicatePaths) != 0 {
		failures = append(failures, "created duplicate audit synthesis path: "+strings.Join(duplicatePaths, ", "))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+sourceAuditSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, []string{sourceAuditCurrentSourcePath, sourceAuditOldSourcePath})...)
	failures = append(failures, presentForbidden(body, forbidden)...)
	if !hasProjection {
		failures = append(failures, "audit synthesis projection is not fresh with current and superseded refs")
	}
	if !hasInvalidation {
		failures = append(failures, "audit synthesis invalidation event missing")
	}
	if !hasRefresh {
		failures = append(failures, "audit synthesis refresh event missing")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report audit repair and current source")
	}
	databasePass := found &&
		exactCount == 1 &&
		len(duplicatePaths) == 0 &&
		docIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, []string{sourceAuditCurrentSourcePath, sourceAuditOldSourcePath})) == 0 &&
		len(presentForbidden(body, forbidden)) == 0 &&
		hasProjection &&
		hasInvalidation &&
		hasRefresh
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceAuditSynthesisPath, sourceAuditDecoyPath, sourceAuditCurrentSourcePath, sourceAuditOldSourcePath},
	}, nil
}

func verifySourceSensitiveConflict(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: sourceAuditConflictSearchText, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	alphaID, alphaFound, err := documentIDByPath(ctx, paths, sourceAuditConflictAlphaPath)
	if err != nil {
		return verificationResult{}, err
	}
	bravoID, bravoFound, err := documentIDByPath(ctx, paths, sourceAuditConflictBravoPath)
	if err != nil {
		return verificationResult{}, err
	}
	alphaEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   alphaID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	bravoEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   bravoID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}

	searchHasBoth := searchContainsPath(search, sourceAuditConflictAlphaPath) && searchContainsPath(search, sourceAuditConflictBravoPath)
	hasProvenance := alphaFound && bravoFound &&
		alphaEvents.Provenance != nil && len(alphaEvents.Provenance.Events) > 0 &&
		bravoEvents.Provenance != nil && len(bravoEvents.Provenance.Events) > 0
	assistantPass := messageContainsAll(finalMessage, []string{sourceAuditConflictAlphaPath, sourceAuditConflictBravoPath}) &&
		messageContainsAny(finalMessage, []string{"conflict", "conflicting", "contradict", "contradiction"}) &&
		messageContainsAny(finalMessage, []string{"both are current", "both sources are current", "current sources", "both current"}) &&
		messageContainsAny(finalMessage, []string{"unresolved", "no supersession", "no source authority", "cannot choose", "do not choose"}) &&
		messageContainsAny(finalMessage, []string{"seven", "7"}) &&
		messageContainsAny(finalMessage, []string{"thirty", "30"})
	inspectedBothProvenanceRefs := provenanceEventRefIDsInclude(turnMetrics.ProvenanceEventRefIDs, alphaID, bravoID)
	activityPass := turnMetrics.SearchUsed && inspectedBothProvenanceRefs

	failures := []string{}
	if !searchHasBoth {
		failures = append(failures, "search did not find both conflict sources")
	}
	if !hasProvenance {
		failures = append(failures, "document provenance missing for conflict sources")
	}
	if synthesisCount != 0 {
		failures = append(failures, fmt.Sprintf("conflict explanation created %d synthesis documents", synthesisCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !inspectedBothProvenanceRefs {
		failures = append(failures, "agent did not inspect provenance events for both conflict sources")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not explain unresolved conflicting source evidence")
	}
	databasePass := searchHasBoth && hasProvenance && synthesisCount == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceAuditConflictAlphaPath, sourceAuditConflictBravoPath},
	}, nil
}

func verifyPopulatedHeterogeneousRetrieval(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          populatedSearchText,
			MetadataKey:   "populated_role",
			MetadataValue: "authority",
			Limit:         5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	duplicateSearch, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  populatedDuplicateSearchText,
			Limit: 10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	staleSearch, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  populatedStaleSearchText,
			Limit: 10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	top, topFound := topSearchHit(search)
	requiredPaths := populatedVaultFixturePaths()
	missingDocs := []string{}
	for _, path := range requiredPaths {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			return verificationResult{}, err
		} else if !found {
			missingDocs = append(missingDocs, path)
		}
	}
	duplicateSourcesVisible := searchContainsPath(duplicateSearch, populatedAuthorityCandidatePath) &&
		searchContainsPath(duplicateSearch, populatedReceiptDuplicatePath)
	staleSourcesVisible := searchContainsPath(staleSearch, populatedInvoiceStalePath) &&
		searchContainsPath(staleSearch, populatedLegalArchivePath) &&
		searchContainsPath(staleSearch, populatedSynthesisOldPath)
	assistantPass := topFound &&
		messageContainsAll(finalMessage, []string{populatedAuthorityPath, top.DocID, top.ChunkID, "USD 500", "USD 118.42", "privacy addendum"}) &&
		messageContainsAny(finalMessage, []string{"polluted", "decoy", "reject", "did not use", "not authority"})
	forbiddenAnswer := messageContainsAny(finalMessage, []string{"ignore the privacy addendum", "approve every invoice without review"})
	activityPass := turnMetrics.SearchUsed && turnMetrics.SearchMetadataFilterUsed
	failures := populatedBypassFailures(turnMetrics)
	if len(missingDocs) != 0 {
		failures = append(failures, "missing populated fixture docs: "+strings.Join(missingDocs, ", "))
	}
	if !topFound || searchHitPath(top) != populatedAuthorityPath || !searchHitHasCitation(top) {
		failures = append(failures, "authority search did not return cited populated authority source")
	}
	if !duplicateSourcesVisible {
		failures = append(failures, "duplicate candidate search did not expose populated duplicate source and receipt pressure")
	}
	if !staleSourcesVisible {
		failures = append(failures, "stale source search did not expose populated stale invoice, legal, and synthesis pressure")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.SearchMetadataFilterUsed {
		failures = append(failures, "agent did not use metadata-filtered retrieval search")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not cite authority path, doc_id, chunk_id, and grounded Atlas facts")
	}
	if forbiddenAnswer {
		failures = append(failures, "final answer repeated polluted decoy claims")
	}
	databasePass := len(missingDocs) == 0 &&
		topFound &&
		searchHitPath(top) == populatedAuthorityPath &&
		searchHitHasCitation(top) &&
		duplicateSourcesVisible &&
		staleSourcesVisible
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass && !forbiddenAnswer && len(populatedBypassFailures(turnMetrics)) == 0,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass && !forbiddenAnswer && len(populatedBypassFailures(turnMetrics)) == 0,
		Details:       missingDetails(failures),
		Documents:     requiredPaths,
	}, nil
}

func verifyPopulatedFreshnessConflict(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: populatedConflictSearchText, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	alphaID, alphaFound, err := documentIDByPath(ctx, paths, populatedConflictAlphaPath)
	if err != nil {
		return verificationResult{}, err
	}
	bravoID, bravoFound, err := documentIDByPath(ctx, paths, populatedConflictBravoPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisID, synthesisFound, err := documentIDByPath(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisBody, synthesisBodyFound, err := documentBodyByPath(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	decoyBody, decoyBodyFound, err := documentBodyByPath(ctx, paths, populatedSynthesisDecoyPath)
	if err != nil {
		return verificationResult{}, err
	}
	alphaEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   alphaID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	bravoEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   bravoID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, synthesisID)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}

	searchHasBoth := searchContainsPath(search, populatedConflictAlphaPath) && searchContainsPath(search, populatedConflictBravoPath)
	hasProvenance := alphaFound && bravoFound &&
		alphaEvents.Provenance != nil && len(alphaEvents.Provenance.Events) > 0 &&
		bravoEvents.Provenance != nil && len(bravoEvents.Provenance.Events) > 0
	assistantPass := messageContainsAll(finalMessage, []string{populatedConflictAlphaPath, populatedConflictBravoPath, populatedSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"conflict", "conflicting", "contradict", "contradiction"}) &&
		messageContainsAny(finalMessage, []string{"both current", "both are current", "current sources", "no supersession"}) &&
		messageContainsAny(finalMessage, []string{"unresolved", "cannot choose", "no source authority", "do not choose"}) &&
		messageContainsAny(finalMessage, []string{"fourteen", "14"}) &&
		messageContainsAny(finalMessage, []string{"thirty", "30"}) &&
		messageContainsAny(finalMessage, []string{"stale", "freshness", "projection", "provenance"})
	forbiddenAnswer := messageContainsAny(finalMessage, []string{"fourteen days is correct", "thirty days is correct", "choose fourteen", "choose thirty"})
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	failures := populatedBypassFailures(turnMetrics)
	if !searchHasBoth {
		failures = append(failures, "search did not find both populated conflict sources")
	}
	if !hasProvenance {
		failures = append(failures, "document provenance missing for populated conflict sources")
	}
	if !synthesisFound || projection == nil {
		failures = append(failures, "synthesis projection missing for populated stale synthesis")
	}
	if !synthesisBodyFound || synthesisBody != populatedSynthesisSeedBody() {
		failures = append(failures, populatedSynthesisPath+" changed during no-write conflict scenario")
	}
	if !decoyBodyFound || decoyBody != populatedSynthesisDecoySeedBody() {
		failures = append(failures, populatedSynthesisDecoyPath+" changed during no-write conflict scenario")
	}
	if synthesisCount != 2 {
		failures = append(failures, fmt.Sprintf("expected target and decoy synthesis only, got %d synthesis documents", synthesisCount))
	}
	if !activityPass {
		failures = append(failures, "agent did not use required search/list/get/projection/provenance workflow")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not explain unresolved populated freshness conflict")
	}
	if forbiddenAnswer {
		failures = append(failures, "final answer chose a conflict winner without authority")
	}
	synthesisUnchanged := synthesisBodyFound &&
		synthesisBody == populatedSynthesisSeedBody() &&
		decoyBodyFound &&
		decoyBody == populatedSynthesisDecoySeedBody()
	databasePass := searchHasBoth && hasProvenance && synthesisFound && projection != nil && synthesisUnchanged && synthesisCount == 2
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass && !forbiddenAnswer && len(populatedBypassFailures(turnMetrics)) == 0,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass && !forbiddenAnswer && len(populatedBypassFailures(turnMetrics)) == 0,
		Details:       missingDetails(failures),
		Documents:     []string{populatedSynthesisPath, populatedSynthesisDecoyPath, populatedConflictAlphaPath, populatedConflictBravoPath},
	}, nil
}

func verifyPopulatedSynthesisUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicatePaths, err := disallowedDocumentPathsWithPrefix(ctx, paths, "synthesis/", map[string]bool{
		populatedSynthesisPath:      true,
		populatedSynthesisDecoyPath: true,
	})
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docID)
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: populatedSynthesisSearchText, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	events, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + docID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"source_refs: " + populatedSynthesisCurrentPath + ", " + populatedSynthesisOldPath,
		"Current populated vault synthesis guidance: update the existing synthesis page",
		"Current source: " + populatedSynthesisCurrentPath,
		"Superseded source: " + populatedSynthesisOldPath,
		"## Sources",
		"## Freshness",
	}
	forbidden := []string{"create a duplicate synthesis page when Atlas source claims change", "create a duplicate synthesis page"}
	hasProjection := projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", populatedSynthesisCurrentPath) &&
		projectionDetailContains(projection.Details, "superseded_source_refs", populatedSynthesisOldPath)
	searchHasCurrent := searchContainsPath(search, populatedSynthesisCurrentPath)
	hasInvalidation := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_invalidated")
	hasRefresh := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_refreshed")
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	assistantPass := messageContainsAll(finalMessage, []string{populatedSynthesisPath, populatedSynthesisCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "fresh", "freshness", "no duplicate"})
	failures := populatedBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+populatedSynthesisPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", populatedSynthesisPath, exactCount))
	}
	if len(duplicatePaths) != 0 {
		failures = append(failures, "created duplicate populated synthesis path: "+strings.Join(duplicatePaths, ", "))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+populatedSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, []string{populatedSynthesisCurrentPath, populatedSynthesisOldPath})...)
	failures = append(failures, presentForbidden(body, forbidden)...)
	if !hasProjection {
		failures = append(failures, "populated synthesis projection is not fresh with current and superseded refs")
	}
	if !searchHasCurrent {
		failures = append(failures, "populated synthesis search did not find current source")
	}
	if !hasInvalidation {
		failures = append(failures, "populated synthesis invalidation event missing")
	}
	if !hasRefresh {
		failures = append(failures, "populated synthesis refresh event missing")
	}
	if !activityPass {
		failures = append(failures, "agent did not use required search/list/get/projection/provenance workflow")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report populated synthesis update and current source")
	}
	databasePass := found &&
		exactCount == 1 &&
		len(duplicatePaths) == 0 &&
		docIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, []string{populatedSynthesisCurrentPath, populatedSynthesisOldPath})) == 0 &&
		len(presentForbidden(body, forbidden)) == 0 &&
		hasProjection &&
		searchHasCurrent &&
		hasInvalidation &&
		hasRefresh
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass && len(populatedBypassFailures(turnMetrics)) == 0,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass && len(populatedBypassFailures(turnMetrics)) == 0,
		Details:       missingDetails(failures),
		Documents:     []string{populatedSynthesisPath, populatedSynthesisDecoyPath, populatedSynthesisCurrentPath, populatedSynthesisOldPath},
	}, nil
}

func verifyRepoDocsAgentOpsRetrieval(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       repoDocsRetrievalSearchText,
			PathPrefix: "docs/architecture/",
			Limit:      10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	top, topFound := topSearchHit(search)
	agentOpsDocID, hasAgentOpsDoc, err := documentIDByPath(ctx, paths, repoDocsAgentOpsADRPath)
	if err != nil {
		return verificationResult{}, err
	}
	hasAgentOpsADR := searchContainsPath(search, repoDocsAgentOpsADRPath) ||
		(hasAgentOpsDoc && stringValuesInclude(turnMetrics.GetDocumentDocIDs, agentOpsDocID))
	_, hasKnowledgeConfig, err := documentIDByPath(ctx, paths, repoDocsKnowledgeConfigPath)
	if err != nil {
		return verificationResult{}, err
	}
	assistantPass := messageContainsAll(finalMessage, []string{repoDocsAgentOpsADRPath}) &&
		messageContainsAny(finalMessage, []string{"AgentOps", "agentops"}) &&
		messageContainsAny(finalMessage, []string{"installed", "openclerk", "runner"}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation", "cited"})
	searchedArchitecture := turnMetrics.SearchUsed && containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"docs/architecture/"})
	activityPass := len(repoDocsBypassFailures(turnMetrics)) == 0 &&
		searchedArchitecture &&
		hasAgentOpsADR
	failures := repoDocsBypassFailures(turnMetrics)
	if !topFound || !searchHitHasCitation(top) {
		failures = append(failures, "repo-docs retrieval search did not return cited hits")
	}
	if !hasAgentOpsDoc {
		failures = append(failures, "repo-docs seed did not import AgentOps ADR")
	}
	if hasAgentOpsDoc && !hasAgentOpsADR {
		failures = append(failures, "repo-docs retrieval workflow did not expose AgentOps ADR")
	}
	if !hasKnowledgeConfig {
		failures = append(failures, "repo-docs seed did not import knowledge configuration ADR")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !searchedArchitecture {
		failures = append(failures, "agent did not use a docs/architecture/ path-prefix search")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not cite repo AgentOps docs with runner evidence")
	}
	databasePass := topFound && searchHitHasCitation(top) && hasAgentOpsDoc && hasKnowledgeConfig
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{repoDocsAgentOpsADRPath, repoDocsKnowledgeConfigPath},
	}, nil
}

func verifyRepoDocsSynthesisMaintenance(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, repoDocsSynthesisPath, finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:      []string{repoDocsAgentProductionPath, repoDocsBaselineScenariosPath},
		RequireSearch:   true,
		RequireList:     true,
		Metrics:         turnMetrics,
		FinalAnswerPath: true,
		AdditionalDocs:  []string{repoDocsAgentProductionPath, repoDocsBaselineScenariosPath},
		AdditionalBodyRequirements: []string{
			"Repo-docs dogfood decision: use the existing OpenClerk document and retrieval runner actions.",
			"Production gate source: " + repoDocsAgentProductionPath,
			"Baseline scenarios source: " + repoDocsBaselineScenariosPath,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	count, err := exactDocumentCount(ctx, paths, repoDocsSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := repoDocsBypassFailures(turnMetrics)
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if count != 1 {
		failures = append(failures, fmt.Sprintf("expected one repo-docs synthesis document, got %d", count))
	}
	databasePass := base.DatabasePass && count == 1
	assistantPass := base.AssistantPass && len(repoDocsBypassFailures(turnMetrics)) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyRepoDocsDecisionRecords(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	lookup, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{
			Text:   "knowledge configuration",
			Status: "accepted",
			Scope:  "knowledge-configuration",
			Owner:  "platform",
			Limit:  5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	agentOpsDecision, agentOpsDecisionErr := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: "adr-agentops-only-knowledge-plane",
	})
	configProjection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-knowledge-configuration-v1",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	agentOpsProjection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-agentops-only-knowledge-plane",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "decisions:adr-knowledge-configuration-v1",
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	searchedArchitecture := turnMetrics.SearchUsed && containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"docs/architecture/"})
	hasConfigDecision := false
	if lookup.Decisions != nil {
		for _, decision := range lookup.Decisions.Decisions {
			if decision.DecisionID == "adr-knowledge-configuration-v1" &&
				decision.Status == "accepted" &&
				decision.Scope == "knowledge-configuration" &&
				decision.Owner == "platform" &&
				len(decision.Citations) > 0 &&
				decision.Citations[0].Path == repoDocsKnowledgeConfigPath {
				hasConfigDecision = true
				break
			}
		}
	}
	hasAgentOpsDecisionRecord := agentOpsDecisionErr == nil &&
		agentOpsDecision.Decision != nil &&
		agentOpsDecision.Decision.DecisionID == "adr-agentops-only-knowledge-plane" &&
		agentOpsDecision.Decision.Status == "accepted" &&
		agentOpsDecision.Decision.Scope == "knowledge-plane" &&
		len(agentOpsDecision.Decision.Citations) > 0 &&
		agentOpsDecision.Decision.Citations[0].Path == repoDocsAgentOpsADRPath
	hasAgentOpsDecision := hasAgentOpsDecisionRecord
	hasConfigProjection := configProjection.Projections != nil &&
		len(configProjection.Projections.Projections) == 1 &&
		configProjection.Projections.Projections[0].Freshness == "fresh" &&
		configProjection.Projections.Projections[0].Details["path"] == repoDocsKnowledgeConfigPath
	hasAgentOpsProjection := agentOpsProjection.Projections != nil &&
		len(agentOpsProjection.Projections.Projections) == 1 &&
		agentOpsProjection.Projections.Projections[0].Freshness == "fresh" &&
		agentOpsProjection.Projections.Projections[0].Details["path"] == repoDocsAgentOpsADRPath
	hasProvenance := provenance.Provenance != nil && eventTypesInclude(provenance.Provenance.Events, "projection_refreshed")
	inspectedAgentOpsDecision := decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, "adr-agentops-only-knowledge-plane")
	assistantPass := messageContainsAll(finalMessage, []string{repoDocsAgentOpsADRPath, repoDocsKnowledgeConfigPath}) &&
		messageContainsAny(finalMessage, []string{"canonical markdown", "canonical adr", "authoritative"}) &&
		messageContainsAny(finalMessage, []string{"decisions_lookup", "decisions lookup", "decision lookup", "decision records"}) &&
		messageContainsAny(finalMessage, []string{"decision_record", "decision record", "adr record"}) &&
		messageContainsAny(finalMessage, []string{"fresh", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "projection"})
	activityPass := len(repoDocsBypassFailures(turnMetrics)) == 0 &&
		searchedArchitecture &&
		turnMetrics.DecisionsLookupUsed &&
		inspectedAgentOpsDecision &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	failures := repoDocsBypassFailures(turnMetrics)
	if !searchedArchitecture {
		failures = append(failures, "agent did not use a docs/architecture/ path-prefix search")
	}
	if !hasConfigDecision {
		failures = append(failures, "repo-docs knowledge configuration decision lookup missing")
	}
	if !hasAgentOpsDecision {
		failures = append(failures, "repo-docs AgentOps decision detail missing")
	}
	if !hasConfigProjection {
		failures = append(failures, "repo-docs knowledge configuration decision projection is not fresh")
	}
	if !hasAgentOpsProjection {
		failures = append(failures, "repo-docs AgentOps decision projection is not fresh")
	}
	if !hasProvenance {
		failures = append(failures, "repo-docs decision projection provenance missing")
	}
	if !activityPass {
		failures = append(failures, "agent did not use required search/decision/projection/provenance workflow")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report repo-docs decision-record evidence")
	}
	databasePass := hasConfigDecision && hasAgentOpsDecision && hasConfigProjection && hasAgentOpsProjection && hasProvenance
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{repoDocsAgentOpsADRPath, repoDocsKnowledgeConfigPath},
	}, nil
}

func repoDocsBypassFailures(turnMetrics metrics) []string {
	failures := []string{}
	if turnMetrics.BroadRepoSearch {
		failures = append(failures, "agent used broad repo search")
	}
	if turnMetrics.DirectSQLiteAccess {
		failures = append(failures, "agent used direct SQLite")
	}
	if turnMetrics.LegacyRunnerUsage {
		failures = append(failures, "agent used source-built or legacy runner path")
	}
	if turnMetrics.GeneratedFileInspection {
		failures = append(failures, "agent inspected generated files")
	}
	if turnMetrics.ModuleCacheInspection {
		failures = append(failures, "agent inspected module cache")
	}
	return failures
}

func verifyAgentChosenPathProposal(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	count, err := exactDocumentCount(ctx, paths, agentChosenProposalPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if count != 0 {
		failures = append(failures, "path proposal scenario created "+agentChosenProposalPath)
	}
	if turnMetrics.ToolCalls != 0 || turnMetrics.CommandExecutions != 0 {
		failures = append(failures, fmt.Sprintf("path proposal should ask before writing without tools, got tools=%d commands=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions))
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenProposalPath, agentChosenURLHarness, agentChosenURLPromptGuidance}) &&
		messageContainsAny(finalMessage, []string{"confirm", "confirmation", "approve", "proceed", "before creating", "before i create"}) &&
		messageContainsAny(finalMessage, []string{"no document was created", "did not create", "not create", "nothing was created"}) &&
		messageContainsAny(finalMessage, []string{"no unsupported runner action", "unsupported runner action is implied", "no unsupported action"})
	if !assistantPass {
		failures = append(failures, "final answer did not propose the stable path and ask for confirmation before creating")
	}
	databasePass := count == 0
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls == 0 && turnMetrics.CommandExecutions == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{agentChosenProposalPath},
	}, nil
}

func verifyAgentChosenExplicitFields(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, agentChosenExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, agentChosenExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/")
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: note",
		"Path policy: explicit fields required.",
		"Title policy: explicit title wins.",
		"Document type policy: explicit type wins.",
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+agentChosenExplicitPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", agentChosenExplicitPath, exactCount))
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous source documents for explicit-fields scenario, got %d", sourcesCount))
	}
	if synthesisCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous synthesis documents for explicit-fields scenario, got %d", synthesisCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit-fields document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenExplicitPath}) &&
		messageContainsAny(finalMessage, []string{"Explicit Fields Path Title Type", "explicit title", "title"}) &&
		messageContainsAny(finalMessage, []string{"explicit", "provided", "user-specified"})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit path/title/type handling")
	}
	databasePass := found &&
		exactCount == 1 &&
		len(missingRequired(body, required)) == 0 &&
		sourcesCount == 0 &&
		synthesisCount == 0
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{agentChosenExplicitPath},
	}, nil
}

func verifyAgentChosenAutonomousPlacement(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, agentChosenAutonomousPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, agentChosenAutonomousPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := documentCountWithPrefix(ctx, paths, "sources/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: source",
		agentChosenURLHarness,
		agentChosenURLPromptGuidance,
		"Path policy: autonomous create then report",
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+agentChosenAutonomousPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", agentChosenAutonomousPath, exactCount))
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one autonomous source document, got %d", sourceCount))
	}
	failures = append(failures, missingRequired(body, required)...)
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenAutonomousPath}) &&
		messageContainsAny(finalMessage, []string{"created", "wrote", "filed"})
	if !assistantPass {
		failures = append(failures, "final answer did not report the chosen autonomous path")
	}
	databasePass := found && exactCount == 1 && sourceCount == 1 && len(missingRequired(body, required)) == 0
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{agentChosenAutonomousPath},
	}, nil
}

func verifyAgentChosenSynthesisPathSelection(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, agentChosenSynthesisPath, finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:              []string{agentChosenSynthesisAlphaPath, agentChosenSynthesisBetaPath, agentChosenSynthesisGammaPath},
		RequireSearch:           true,
		RequireList:             true,
		RequireProjectionStates: true,
		Metrics:                 turnMetrics,
		FinalAnswerPath:         true,
		AdditionalDocs:          []string{agentChosenSynthesisAlphaPath, agentChosenSynthesisBetaPath, agentChosenSynthesisGammaPath},
		AdditionalBodyRequirements: []string{
			"explicit-path compatibility",
			"metadata",
			"freshness",
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if synthesisCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one chosen synthesis document, got %d", synthesisCount))
	}
	databasePass := base.DatabasePass && synthesisCount == 1
	assistantPass := base.AssistantPass && len(agentChosenBypassFailures(turnMetrics)) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyAgentChosenAmbiguousDocumentType(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	docPath, body, found, err := documentContaining(ctx, paths, "decision_id: "+agentChosenAmbiguousDecisionID)
	if err != nil {
		return verificationResult{}, err
	}
	decision, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: agentChosenAmbiguousDecisionID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      agentChosenAmbiguousDecisionID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"decision_id: " + agentChosenAmbiguousDecisionID,
		"decision_status: accepted",
		"decision_scope: document-path-selection",
		"Metadata authority: frontmatter decides document identity.",
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing ambiguous decision document")
	}
	failures = append(failures, missingRequired(body, required)...)
	hasDecision := decision.Decision != nil &&
		decision.Decision.DecisionID == agentChosenAmbiguousDecisionID &&
		decision.Decision.Status == "accepted" &&
		decision.Decision.Scope == "document-path-selection" &&
		len(decision.Decision.Citations) > 0
	if !hasDecision {
		failures = append(failures, "decision_record did not expose metadata-derived decision identity")
	}
	hasProjection := projection.Projections != nil &&
		len(projection.Projections.Projections) == 1 &&
		projection.Projections.Projections[0].Freshness == "fresh"
	if !hasProjection {
		failures = append(failures, "decision projection is not fresh")
	}
	inspectedDecision := decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, agentChosenAmbiguousDecisionID)
	if !inspectedDecision {
		failures = append(failures, "agent did not inspect decision_record for metadata-derived identity")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect decision projection freshness")
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenAmbiguousDecisionID}) &&
		messageContainsAny(finalMessage, []string{"metadata", "frontmatter"}) &&
		messageContainsAny(finalMessage, []string{"not filename", "not the filename", "not path", "not the path", "not filename/path"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness", "fresh"}) &&
		docPath != "" && messageContainsAll(finalMessage, []string{docPath})
	if !assistantPass {
		failures = append(failures, "final answer did not report chosen path and metadata authority")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && hasDecision && hasProjection
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && inspectedDecision && turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func verifyAgentChosenUserPathInstructions(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, agentChosenUserSpecifiedPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/")
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"User path instruction wins.",
		"Do not override explicit path instructions.",
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+agentChosenUserSpecifiedPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous sources when user path wins, got %d", sourcesCount))
	}
	if synthesisCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous synthesis when user path wins, got %d", synthesisCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit-path document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenUserSpecifiedPath})
	if !assistantPass {
		failures = append(failures, "final answer did not mention explicit user path")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && sourcesCount == 0 && synthesisCount == 0
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{agentChosenUserSpecifiedPath},
	}, nil
}

func verifyPathTitleURLOnlyAutonomy(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, pathTitleURLOnlyPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, pathTitleURLOnlyPath)
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	required := []string{
		"type: source",
		agentChosenURLHarness,
		agentChosenURLPromptGuidance,
		"Path/title policy: autonomy pressure create then report.",
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+pathTitleURLOnlyPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", pathTitleURLOnlyPath, exactCount))
	}
	if found && title != pathTitleURLOnlyTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", pathTitleURLOnlyTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create URL-only source through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{pathTitleURLOnlyPath}) &&
		messageContainsAny(finalMessage, []string{pathTitleURLOnlyTitle, "harness", "prompt guidance"})
	if !assistantPass {
		failures = append(failures, "final answer did not report chosen path/title")
	}
	databasePass := found && exactCount == 1 && title == pathTitleURLOnlyTitle && len(missingRequired(body, required)) == 0
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{pathTitleURLOnlyPath},
	}, nil
}

func verifyPathTitleMultiSourceDuplicate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, pathTitleSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, pathTitleSynthesisDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current path/title autonomy guidance: update existing synthesis candidate.",
		"## Sources",
		"## Freshness",
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+pathTitleSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, []string{pathTitleSynthesisAlphaPath, pathTitleSynthesisBetaPath})...)
	if duplicateCount != 0 {
		failures = append(failures, "created duplicate synthesis "+pathTitleSynthesisDuplicatePath)
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not inspect existing synthesis before update")
	}
	assistantPass := messageContainsAll(finalMessage, []string{pathTitleSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "existing"}) &&
		messageContainsAny(finalMessage, []string{"no duplicate", "avoided duplicate", "not create a duplicate"})
	if !assistantPass {
		failures = append(failures, "final answer did not report existing synthesis update and duplicate avoidance")
	}
	databasePass := found &&
		duplicateCount == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, []string{pathTitleSynthesisAlphaPath, pathTitleSynthesisBetaPath})) == 0
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 && turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{pathTitleSynthesisPath, pathTitleSynthesisDuplicatePath, pathTitleSynthesisAlphaPath, pathTitleSynthesisBetaPath},
	}, nil
}

func verifyPathTitleExplicitOverrides(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, pathTitleExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/path-title/")
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	required := []string{
		"type: note",
		"Explicit path/title override wins.",
		"Do not apply autonomous path conventions.",
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+pathTitleExplicitPath)
	}
	if found && title != pathTitleExplicitTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", pathTitleExplicitTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous path-title source docs, got %d", sourcesCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit override document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{pathTitleExplicitPath, pathTitleExplicitTitle})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit override path and title")
	}
	databasePass := found && title == pathTitleExplicitTitle && len(missingRequired(body, required)) == 0 && sourcesCount == 0
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{pathTitleExplicitPath},
	}, nil
}

func verifyPathTitleDuplicateRisk(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingCount, err := exactDocumentCount(ctx, paths, pathTitleDuplicateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	pathTitleSourceCount, err := documentCountWithPrefix(ctx, paths, "sources/path-title/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing source %s, got %d", pathTitleDuplicateExistingPath, existingCount))
	}
	if pathTitleSourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected only the seeded path-title source document, got %d", pathTitleSourceCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search for duplicate risk")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source candidates")
	}
	assistantPass := messageContainsAll(finalMessage, []string{pathTitleDuplicateExistingPath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "reuse"}) &&
		messageContainsAny(finalMessage, []string{"not create", "did not create", "no new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate risk and no-create outcome")
	}
	databasePass := existingCount == 1 && pathTitleSourceCount == 1
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 && turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{pathTitleDuplicateExistingPath, pathTitleDuplicateCandidatePath},
	}, nil
}

func verifyPathTitleMetadataAuthority(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	docPath, body, found, err := documentContaining(ctx, paths, "decision_id: "+pathTitleMetadataDecisionID)
	if err != nil {
		return verificationResult{}, err
	}
	decision, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: pathTitleMetadataDecisionID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      pathTitleMetadataDecisionID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"decision_id: " + pathTitleMetadataDecisionID,
		"decision_title: " + pathTitleMetadataTitle,
		"decision_status: accepted",
		"Metadata authority: frontmatter decides path/title identity.",
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing path/title metadata authority decision")
	}
	failures = append(failures, missingRequired(body, required)...)
	hasDecision := decision.Decision != nil &&
		decision.Decision.DecisionID == pathTitleMetadataDecisionID &&
		decision.Decision.Status == "accepted" &&
		len(decision.Decision.Citations) > 0
	if !hasDecision {
		failures = append(failures, "decision_record did not expose metadata authority decision")
	}
	hasProjection := projection.Projections != nil &&
		len(projection.Projections.Projections) == 1 &&
		projection.Projections.Projections[0].Freshness == "fresh"
	if !hasProjection {
		failures = append(failures, "decision projection is not fresh")
	}
	if !decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, pathTitleMetadataDecisionID) {
		failures = append(failures, "agent did not inspect decision_record")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection_states")
	}
	assistantPass := docPath != "" &&
		messageContainsAll(finalMessage, []string{docPath, pathTitleMetadataDecisionID}) &&
		messageContainsAny(finalMessage, []string{"metadata", "frontmatter"}) &&
		messageContainsAny(finalMessage, []string{"not filename", "not path", "not filename/path"}) &&
		messageContainsAny(finalMessage, []string{"fresh", "projection"})
	if !assistantPass {
		failures = append(failures, "final answer did not report metadata authority and projection evidence")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && hasDecision && hasProjection
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 &&
		decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, pathTitleMetadataDecisionID) &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func verifyDocumentThisExplicitCreate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, documentThisExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: note",
		"Document-this explicit article/docs/paper/transcript intake uses strict runner JSON.",
		"Required fields were supplied before create_document.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisExplicitPath)
	}
	if found && title != documentThisExplicitTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", documentThisExplicitTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no source autofiling docs, got %d", sourcesCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisExplicitPath, documentThisExplicitTitle})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit document path and title")
	}
	databasePass := found && title == documentThisExplicitTitle && len(missingRequired(body, required)) == 0 && sourcesCount == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisExplicitPath},
	}, nil
}

func verifyDocumentThisExplicitOverrides(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, documentThisOverridePath)
	if err != nil {
		return verificationResult{}, err
	}
	autofiledCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	required := []string{
		"type: note",
		"Explicit document-this override path and title win.",
		"Do not infer a sources/ path from mixed URLs.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisOverridePath)
	}
	if found && title != documentThisOverrideTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", documentThisOverrideTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if autofiledCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no inferred source docs, got %d", autofiledCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit override through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisOverridePath, documentThisOverrideTitle})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit override path and title")
	}
	databasePass := found && title == documentThisOverrideTitle && len(missingRequired(body, required)) == 0 && autofiledCount == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisOverridePath},
	}, nil
}

func verifyDocumentThisDuplicateCandidate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingCount, err := exactDocumentCount(ctx, paths, documentThisDuplicateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, documentThisDuplicateCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentThisBypassFailures(turnMetrics)
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing source %s, got %d", documentThisDuplicateExistingPath, existingCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created duplicate candidate "+documentThisDuplicateCandidatePath)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected only the seeded document-this source document, got %d", sourceCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search for duplicate candidate")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source candidates")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisDuplicateExistingPath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already"}) &&
		messageContainsAny(finalMessage, []string{"not create", "did not create", "no new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate candidate and no-create outcome")
	}
	databasePass := existingCount == 1 && candidateCount == 0 && sourceCount == 1
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisDuplicateExistingPath, documentThisDuplicateCandidatePath},
	}, nil
}

func verifyDocumentThisExistingUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, documentThisUpdateTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	decoyBody, decoyFound, err := documentBodyByPath(ctx, paths, documentThisUpdateDecoyPath)
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"## Decisions",
		"Use strict runner JSON for document-this intake.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisUpdateTargetPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	if decoyFound && strings.Contains(decoyBody, "Use strict runner JSON for document-this intake.") {
		failures = append(failures, "updated decoy "+documentThisUpdateDecoyPath)
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list update candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not inspect existing target before update")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisUpdateTargetPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "appended", "replaced"}) &&
		messageContainsAny(finalMessage, []string{"decoy", "not update", "did not update", "target"})
	if !assistantPass {
		failures = append(failures, "final answer did not report target update and decoy avoidance")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && (!decoyFound || !strings.Contains(decoyBody, "Use strict runner JSON for document-this intake."))
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisUpdateTargetPath, documentThisUpdateDecoyPath},
	}, nil
}

func verifyDocumentThisSynthesisFreshness(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, documentThisSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, documentThisSynthesisDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current document-this intake guidance: update existing synthesis after source, duplicate, provenance, and freshness checks.",
		"## Sources",
		"## Freshness",
	}
	expectedRefs := []string{documentThisArticlePath, documentThisDocsPath, documentThisPaperPath, documentThisTranscriptPath}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, expectedRefs)...)
	if duplicateCount != 0 {
		failures = append(failures, "created duplicate synthesis "+documentThisSynthesisDuplicatePath)
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search source evidence")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not inspect existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection_states")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance_events")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"freshness", "projection", "fresh"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source refs", "source_refs"}) &&
		messageContainsAny(finalMessage, []string{"no duplicate", "did not create", "not create"})
	if !assistantPass {
		failures = append(failures, "final answer did not report synthesis update, freshness/provenance, and duplicate avoidance")
	}
	databasePass := found &&
		duplicateCount == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, expectedRefs)) == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents: append([]string{
			documentThisSynthesisPath,
			documentThisSynthesisDuplicatePath,
		}, expectedRefs...),
	}, nil
}

type documentArtifactCandidateExpectation struct {
	Path             string
	Title            string
	RequiredBody     []string
	ForbiddenBody    []string
	RequireValidate  bool
	RequireNoCreate  bool
	RequireApproval  bool
	RequireBodyShown bool
}

func verifyDocumentArtifactCandidateProposal(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, expectation documentArtifactCandidateExpectation) (verificationResult, error) {
	count, err := exactDocumentCount(ctx, paths, expectation.Path)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentArtifactCandidateBypassFailures(turnMetrics)
	if expectation.RequireNoCreate && count != 0 {
		failures = append(failures, fmt.Sprintf("created candidate document %s before approval", expectation.Path))
	}
	if turnMetrics.CreateDocumentUsed {
		failures = append(failures, "used create_document before approval")
	}
	if expectation.RequireValidate && !turnMetrics.ValidateUsed {
		failures = append(failures, "did not validate strict candidate document JSON")
	}
	if expectation.RequireValidate && (turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0) {
		failures = append(failures, "did not run installed runner validation")
	}
	assistantRequired := []string{expectation.Path, expectation.Title}
	if expectation.RequireBodyShown {
		assistantRequired = append(assistantRequired, expectation.RequiredBody...)
	}
	if !messageContainsAll(finalMessage, assistantRequired) {
		failures = append(failures, "final answer did not include candidate path, title, and required body preview")
	}
	if len(presentForbidden(strings.ToLower(finalMessage), lowerStrings(expectation.ForbiddenBody))) != 0 {
		failures = append(failures, "final answer included forbidden invented body content")
	}
	if expectation.RequireApproval && !messageContainsAny(finalMessage, []string{"confirm", "confirmation", "approve", "approval", "before creating", "before I create"}) {
		failures = append(failures, "final answer did not ask for confirmation before creating")
	}
	if expectation.RequireNoCreate && !messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "before creating"}) {
		failures = append(failures, "final answer did not state that no document was created before approval")
	}
	databasePass := !expectation.RequireNoCreate || count == 0
	activityPass := len(documentArtifactCandidateBypassFailures(turnMetrics)) == 0 &&
		!turnMetrics.CreateDocumentUsed &&
		(!expectation.RequireValidate || (turnMetrics.ValidateUsed && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0))
	assistantPass := messageContainsAll(finalMessage, assistantRequired) &&
		(!expectation.RequireApproval || messageContainsAny(finalMessage, []string{"confirm", "confirmation", "approve", "approval", "before creating", "before I create"})) &&
		(!expectation.RequireNoCreate || messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "before creating"})) &&
		len(presentForbidden(strings.ToLower(finalMessage), lowerStrings(expectation.ForbiddenBody))) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{expectation.Path},
	}, nil
}

func verifyDocumentArtifactCandidateDuplicateRisk(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingCount, err := exactDocumentCount(ctx, paths, candidateDuplicateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, candidateDuplicateCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentArtifactCandidateBypassFailures(turnMetrics)
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one seeded duplicate candidate %s, got %d", candidateDuplicateExistingPath, existingCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created duplicate candidate "+candidateDuplicateCandidatePath)
	}
	if turnMetrics.CreateDocumentUsed {
		failures = append(failures, "used create_document despite duplicate risk")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "did not search for duplicate risk")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "did not list candidate documents")
	}
	assistantPass := messageContainsAll(finalMessage, []string{candidateDuplicateExistingPath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already"}) &&
		messageContainsAny(finalMessage, []string{"confirm", "choose", "update", "create new", "approval"}) &&
		messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "no new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate risk and ask before writing")
	}
	databasePass := existingCount == 1 && candidateCount == 0
	activityPass := len(documentArtifactCandidateBypassFailures(turnMetrics)) == 0 &&
		!turnMetrics.CreateDocumentUsed &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{candidateDuplicateExistingPath, candidateDuplicateCandidatePath},
	}, nil
}

func verifyDocumentArtifactCandidateLowConfidence(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	counts := []string{candidateNotePath, candidateHeadingPath, candidateMixedSourcePath, candidateOverridePath, candidateBodyFaithfulnessPath}
	created := []string{}
	for _, path := range counts {
		count, err := exactDocumentCount(ctx, paths, path)
		if err != nil {
			return verificationResult{}, err
		}
		if count != 0 {
			created = append(created, path)
		}
	}
	failures := documentArtifactCandidateBypassFailures(turnMetrics)
	if len(created) != 0 {
		failures = append(failures, "created low-confidence candidate documents: "+strings.Join(created, ", "))
	}
	if turnMetrics.ToolCalls != 0 || turnMetrics.CommandExecutions != 0 || turnMetrics.AssistantCalls > 1 {
		failures = append(failures, fmt.Sprintf("low-confidence ask should be no-tools, got tools=%d commands=%d assistant=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions, turnMetrics.AssistantCalls))
	}
	assistantPass := messageContainsAny(finalMessage, []string{"body", "content", "text", "artifact type", "what to document"}) &&
		messageContainsAny(finalMessage, []string{"missing", "provide", "need", "can't create", "cannot create"}) &&
		!messageContainsAny(finalMessage, []string{candidateNotePath, candidateHeadingPath, candidateMixedSourcePath})
	if !assistantPass {
		failures = append(failures, "final answer did not ask for missing content or intent without proposing a path")
	}
	databasePass := len(created) == 0
	activityPass := len(documentArtifactCandidateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.ToolCalls == 0 &&
		turnMetrics.CommandExecutions == 0 &&
		turnMetrics.AssistantCalls <= 1
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     counts,
	}, nil
}

type artifactPDFExpectation struct {
	SourcePath string
	AssetPath  string
}

func artifactPDFExpectedPaths(scenarioID string) artifactPDFExpectation {
	if scenarioID == artifactPDFNaturalIntentScenarioID {
		return artifactPDFExpectation{
			SourcePath: artifactPDFNaturalSourcePath,
			AssetPath:  artifactPDFNaturalAssetPath,
		}
	}
	return artifactPDFExpectation{
		SourcePath: artifactPDFSourcePath,
		AssetPath:  artifactPDFAssetPath,
	}
}

func verifyArtifactPDFSourceURL(ctx context.Context, paths evalPaths, scenarioID string, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	expectation := artifactPDFExpectedPaths(scenarioID)
	doc, found, err := documentByPath(ctx, paths, expectation.SourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	count, err := exactDocumentCount(ctx, paths, expectation.SourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := artifactIngestionBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing PDF source document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{artifactPDFEvidenceText, "source_url:", "asset_path:", expectation.AssetPath})...)
		if doc.Metadata["asset_path"] != expectation.AssetPath {
			failures = append(failures, fmt.Sprintf("expected asset_path metadata %q, got %q", expectation.AssetPath, doc.Metadata["asset_path"]))
		}
		if doc.Metadata["source_type"] != "pdf" {
			failures = append(failures, fmt.Sprintf("expected source_type metadata pdf, got %q", doc.Metadata["source_type"]))
		}
	}
	if count != 1 {
		failures = append(failures, fmt.Sprintf("expected one PDF source document, got %d", count))
	}
	if !turnMetrics.IngestSourceURLUsed || turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use default create-mode ingest_source_url")
	}
	assistantPass := messageContainsAll(finalMessage, []string{expectation.SourcePath, expectation.AssetPath}) &&
		messageContainsAny(finalMessage, []string{"citation", "citations", "doc_id", "chunk_id"}) &&
		messageContainsAny(finalMessage, []string{"ingested", "created", "source URL"})
	if !assistantPass {
		failures = append(failures, "final answer did not report PDF source ingestion with citation evidence")
	}
	databasePass := found && doc != nil && count == 1 &&
		doc.Metadata["asset_path"] == expectation.AssetPath &&
		doc.Metadata["source_type"] == "pdf" &&
		len(missingRequired(doc.Body, []string{artifactPDFEvidenceText, "source_url:", "asset_path:", expectation.AssetPath})) == 0
	activityPass := len(artifactIngestionBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUsed && !turnMetrics.IngestSourceURLUpdateUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{expectation.SourcePath},
	}, nil
}

func verifyArtifactTranscript(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, artifactTranscriptPath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, artifactTranscriptEvidenceText)
	if err != nil {
		return verificationResult{}, err
	}
	failures := artifactIngestionBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing transcript fixture")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{artifactTranscriptEvidenceText, "artifact_kind: transcript"})...)
	}
	if !searchContainsPath(search, artifactTranscriptPath) || !searchResultHasCitations(search) {
		failures = append(failures, "transcript search did not expose citation-bearing result")
	}
	if !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"transcripts/"}) {
		failures = append(failures, "agent did not search transcript artifact evidence with path_prefix transcripts/")
	}
	assistantPass := messageContainsAll(finalMessage, []string{artifactTranscriptPath}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"}) &&
		messageContainsAny(finalMessage, []string{"canonical markdown", "transcript"})
	if !assistantPass {
		failures = append(failures, "final answer did not cite transcript canonical markdown evidence")
	}
	databasePass := found && doc != nil &&
		len(missingRequired(doc.Body, []string{artifactTranscriptEvidenceText, "artifact_kind: transcript"})) == 0 &&
		searchContainsPath(search, artifactTranscriptPath) && searchResultHasCitations(search)
	activityPass := len(artifactIngestionBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"transcripts/"})
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{artifactTranscriptPath},
	}, nil
}

func verifyArtifactInvoiceReceipt(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	invoice, invoiceFound, err := documentByPath(ctx, paths, artifactInvoicePath)
	if err != nil {
		return verificationResult{}, err
	}
	receipt, receiptFound, err := documentByPath(ctx, paths, artifactReceiptPath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, artifactInvoiceReceiptEvidenceText)
	if err != nil {
		return verificationResult{}, err
	}
	failures := artifactIngestionBypassFailures(turnMetrics)
	if !invoiceFound || invoice == nil {
		failures = append(failures, "missing invoice fixture")
	} else {
		failures = append(failures, missingRequired(invoice.Body, []string{"USD 1250.00", "approval above USD 500"})...)
	}
	if !receiptFound || receipt == nil {
		failures = append(failures, "missing receipt fixture")
	} else {
		failures = append(failures, missingRequired(receipt.Body, []string{"USD 86.40"})...)
	}
	if !searchContainsPath(search, artifactInvoicePath) || !searchContainsPath(search, artifactReceiptPath) || !searchResultHasCitations(search) {
		failures = append(failures, "invoice/receipt search did not expose citation-bearing authority results")
	}
	requiredMetadataFilters := []string{"artifact_kind=invoice", "artifact_kind=receipt"}
	if !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchMetadataFilters, requiredMetadataFilters) {
		failures = append(failures, "agent did not run invoice and receipt artifact_kind metadata-filtered retrieval")
	}
	assistantPass := messageContainsAll(finalMessage, []string{artifactInvoicePath, artifactReceiptPath, "USD 1250.00", "USD 86.40"}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"})
	if !assistantPass {
		failures = append(failures, "final answer did not cite invoice and receipt authority evidence")
	}
	databasePass := invoiceFound && invoice != nil && receiptFound && receipt != nil &&
		len(missingRequired(invoice.Body, []string{"USD 1250.00", "approval above USD 500"})) == 0 &&
		len(missingRequired(receipt.Body, []string{"USD 86.40"})) == 0 &&
		searchContainsPath(search, artifactInvoicePath) && searchContainsPath(search, artifactReceiptPath) &&
		searchResultHasCitations(search)
	activityPass := len(artifactIngestionBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		containsAllStrings(turnMetrics.SearchMetadataFilters, requiredMetadataFilters)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{artifactInvoicePath, artifactReceiptPath},
	}, nil
}

func verifyArtifactMixedSynthesis(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	synthesis, synthesisFound, err := documentByPath(ctx, paths, artifactMixedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	current, currentFound, err := documentByPath(ctx, paths, artifactMixedSynthesisCurrentPath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, artifactMixedSynthesisEvidenceText)
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := artifactProjectionStates(ctx, paths, docIDOrEmpty(synthesis))
	if err != nil {
		return verificationResult{}, err
	}
	failures := artifactIngestionBypassFailures(turnMetrics)
	if !synthesisFound || synthesis == nil {
		failures = append(failures, "missing mixed synthesis fixture")
	} else {
		failures = append(failures, missingRequired(synthesis.Body, []string{artifactMixedSynthesisOldPath, "source_refs:"})...)
	}
	if !currentFound || current == nil {
		failures = append(failures, "missing current mixed artifact source")
	}
	if !searchContainsPath(search, artifactMixedSynthesisCurrentPath) || !searchResultHasCitations(search) {
		failures = append(failures, "mixed artifact current source search did not expose citation-bearing result")
	}
	if !projectionListContainsStaleSource(projections, artifactMixedSynthesisCurrentPath) && !projectionListContainsStaleSource(projections, artifactMixedSynthesisOldPath) {
		failures = append(failures, "synthesis projection did not expose stale or missing current mixed source")
	}
	if !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProjectionStatesUsed || !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect search/list/get/provenance/projection evidence for mixed synthesis")
	}
	assistantPass := messageContainsAll(finalMessage, []string{artifactMixedSynthesisPath, artifactMixedSynthesisOldPath, artifactMixedSynthesisCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"stale", "freshness", "projection"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source refs", "source_refs"})
	if !assistantPass {
		failures = append(failures, "final answer did not explain mixed artifact synthesis freshness and provenance")
	}
	databasePass := synthesisFound && synthesis != nil && currentFound && current != nil &&
		searchContainsPath(search, artifactMixedSynthesisCurrentPath) && searchResultHasCitations(search) &&
		(projectionListContainsStaleSource(projections, artifactMixedSynthesisCurrentPath) || projectionListContainsStaleSource(projections, artifactMixedSynthesisOldPath))
	activityPass := len(artifactIngestionBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed && turnMetrics.ProvenanceEventsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{artifactMixedSynthesisPath, artifactMixedSynthesisOldPath, artifactMixedSynthesisCurrentPath},
	}, nil
}

func verifyVideoYouTubeScriptedTranscript(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, videoYouTubeSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	count, err := exactDocumentCount(ctx, paths, videoYouTubeSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, videoYouTubeSourceEvidenceText)
	if err != nil {
		return verificationResult{}, err
	}
	failures := videoYouTubeBypassFailures(turnMetrics)
	requiredBody := []string{
		videoYouTubeSourceEvidenceText,
		"source_type: video_transcript",
		"source_url:",
		videoYouTubeURL,
		"transcript_origin:",
		videoYouTubeTranscriptOrigin,
		"transcript_sha256:",
		"## Transcript",
	}
	if !found || doc == nil {
		failures = append(failures, "missing video/YouTube canonical source note")
	} else {
		failures = append(failures, missingRequired(doc.Body, requiredBody)...)
		if doc.Metadata["source_type"] != "video_transcript" {
			failures = append(failures, fmt.Sprintf("expected source_type metadata video_transcript, got %q", doc.Metadata["source_type"]))
		}
		if doc.Metadata["source_url"] != videoYouTubeURL {
			failures = append(failures, fmt.Sprintf("expected source_url metadata %q, got %q", videoYouTubeURL, doc.Metadata["source_url"]))
		}
		if doc.Metadata["transcript_origin"] != videoYouTubeTranscriptOrigin {
			failures = append(failures, fmt.Sprintf("expected transcript_origin metadata %q, got %q", videoYouTubeTranscriptOrigin, doc.Metadata["transcript_origin"]))
		}
	}
	if count != 1 {
		failures = append(failures, fmt.Sprintf("expected one video/YouTube source document, got %d", count))
	}
	if !searchContainsPath(search, videoYouTubeSourcePath) || !searchResultHasCitations(search) {
		failures = append(failures, "video/YouTube transcript search did not expose citation-bearing source result")
	}
	if !turnMetrics.IngestVideoURLUsed || turnMetrics.IngestVideoURLUpdateUsed || !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"sources/video-youtube/"}) {
		failures = append(failures, "agent did not use create-mode ingest_video_url and then retrieve the canonical video/YouTube source note with path_prefix sources/video-youtube/")
	}
	assistantPass := messageContainsAll(finalMessage, []string{videoYouTubeSourcePath, videoYouTubeURL}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "transcript_origin", "transcript provenance"})
	if !assistantPass {
		failures = append(failures, "final answer did not report source path, citation evidence, and transcript provenance")
	}
	databasePass := found && doc != nil && count == 1 &&
		len(missingRequired(doc.Body, requiredBody)) == 0 &&
		doc.Metadata["source_type"] == "video_transcript" &&
		doc.Metadata["source_url"] == videoYouTubeURL &&
		doc.Metadata["transcript_origin"] == videoYouTubeTranscriptOrigin &&
		searchContainsPath(search, videoYouTubeSourcePath) &&
		searchResultHasCitations(search)
	activityPass := len(videoYouTubeBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestVideoURLUsed &&
		!turnMetrics.IngestVideoURLUpdateUsed &&
		turnMetrics.SearchUsed &&
		containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"sources/video-youtube/"})
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{videoYouTubeSourcePath},
	}, nil
}

func verifyVideoYouTubeSynthesisFreshness(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	synthesis, synthesisFound, err := documentByPath(ctx, paths, videoYouTubeSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	current, currentFound, err := documentByPath(ctx, paths, videoYouTubeCurrentSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, "transcript")
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := artifactProjectionStates(ctx, paths, docIDOrEmpty(synthesis))
	if err != nil {
		return verificationResult{}, err
	}
	failures := videoYouTubeBypassFailures(turnMetrics)
	if !synthesisFound || synthesis == nil {
		failures = append(failures, "missing video/YouTube synthesis fixture")
	} else {
		failures = append(failures, missingRequired(synthesis.Body, []string{videoYouTubeCurrentSourcePath, "source_refs:"})...)
	}
	if !currentFound || current == nil {
		failures = append(failures, "missing current video/YouTube source fixture")
	} else if current.Metadata["captured_at"] != "2026-04-27T01:00:00Z" {
		failures = append(failures, "current video/YouTube source was not updated to the changed transcript capture time")
	}
	if !searchContainsPath(search, videoYouTubeCurrentSourcePath) || !searchResultHasCitations(search) {
		failures = append(failures, "video/YouTube source search did not expose citation-bearing result after update")
	}
	if !projectionListContainsStaleSource(projections, videoYouTubeCurrentSourcePath) {
		failures = append(failures, "synthesis projection did not expose stale current video/YouTube source after changed transcript update")
	}
	if !turnMetrics.IngestVideoURLUpdateUsed || !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProjectionStatesUsed || !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not update supplied transcript and inspect search/list/get/provenance/projection evidence for video/YouTube synthesis")
	}
	if turnMetrics.CreateDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.AppendDocumentUsed {
		failures = append(failures, "agent mutated synthesis during video/YouTube source update freshness inspection")
	}
	assistantPass := messageContainsAll(finalMessage, []string{videoYouTubeSynthesisPath, videoYouTubeCurrentSourcePath}) &&
		messageContainsAny(finalMessage, []string{"stale", "freshness", "projection"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source refs", "source_refs"}) &&
		messageContainsAny(finalMessage, []string{"no-op", "same hash", "same transcript", "changed transcript", "updated transcript"})
	if !assistantPass {
		failures = append(failures, "final answer did not explain no-op/update video/YouTube synthesis freshness and provenance")
	}
	databasePass := synthesisFound && synthesis != nil && currentFound && current != nil &&
		current.Metadata["captured_at"] == "2026-04-27T01:00:00Z" &&
		searchContainsPath(search, videoYouTubeCurrentSourcePath) &&
		searchResultHasCitations(search) &&
		projectionListContainsStaleSource(projections, videoYouTubeCurrentSourcePath)
	activityPass := len(videoYouTubeBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestVideoURLUpdateUsed &&
		turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed && turnMetrics.ProvenanceEventsUsed &&
		!turnMetrics.CreateDocumentUsed && !turnMetrics.ReplaceSectionUsed && !turnMetrics.AppendDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{videoYouTubeSynthesisPath, videoYouTubeCurrentSourcePath},
	}, nil
}

func artifactSearch(ctx context.Context, paths evalPaths, text string) (runner.RetrievalTaskResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	return runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  text,
			Limit: 10,
		},
	})
}

func artifactProjectionStates(ctx context.Context, paths evalPaths, synthesisDocID string) (runner.ProjectionStateList, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      synthesisDocID,
			Limit:      5,
		},
	})
	if err != nil {
		return runner.ProjectionStateList{}, err
	}
	if result.Projections == nil {
		return runner.ProjectionStateList{}, nil
	}
	return *result.Projections, nil
}

func projectionListContainsStaleSource(list runner.ProjectionStateList, path string) bool {
	for _, projection := range list.Projections {
		if projection.Freshness == "stale" &&
			(projectionDetailContains(projection.Details, "stale_source_refs", path) ||
				projectionDetailContains(projection.Details, "current_source_refs", path) ||
				projectionDetailContains(projection.Details, "missing_source_refs", path)) {
			return true
		}
	}
	return false
}

func agentChosenBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}

func pathTitleBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}

func documentThisBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}

func documentArtifactCandidateBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}

func artifactIngestionBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}

func videoYouTubeBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}

func populatedBypassFailures(turnMetrics metrics) []string {
	failures := []string{}
	if turnMetrics.BroadRepoSearch {
		failures = append(failures, "agent used broad repo search")
	}
	if turnMetrics.DirectSQLiteAccess {
		failures = append(failures, "agent used direct SQLite access")
	}
	if turnMetrics.LegacyRunnerUsage {
		failures = append(failures, "agent used source-built runner path")
	}
	if turnMetrics.GeneratedFileInspection {
		failures = append(failures, "agent inspected generated files")
	}
	if turnMetrics.ModuleCacheInspection {
		failures = append(failures, "agent inspected module cache")
	}
	return failures
}

func populatedVaultFixturePaths() []string {
	return []string{
		populatedTranscriptPath,
		populatedTranscriptOpsPath,
		populatedArticlePath,
		populatedArticleArchivePath,
		populatedMeetingPath,
		populatedMeetingBudgetPath,
		populatedDocsPath,
		populatedDocsRunbookPath,
		populatedBlogPath,
		populatedBlogRumorPath,
		populatedReceiptPath,
		populatedReceiptDuplicatePath,
		populatedInvoicePath,
		populatedInvoiceStalePath,
		populatedLegalPath,
		populatedLegalArchivePath,
		populatedContractPath,
		populatedContractDraftPath,
		populatedAuthorityPath,
		populatedAuthorityCandidatePath,
		populatedPollutedPath,
		populatedConflictAlphaPath,
		populatedConflictBravoPath,
		populatedSynthesisOldPath,
		populatedSynthesisCurrentPath,
		populatedSynthesisPath,
		populatedSynthesisDecoyPath,
	}
}

func populatedVaultFixtureMinimumPrefixCounts() map[string]int {
	return map[string]int{
		"transcripts/": 2,
		"articles/":    2,
		"meetings/":    2,
		"docs/":        2,
		"blogs/":       2,
		"receipts/":    2,
		"invoices/":    2,
		"legal/":       2,
		"contracts/":   2,
		"sources/":     7,
		"synthesis/":   2,
	}
}

func verifyMixedSynthesisRecords(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, "synthesis/openclerk-runner-with-records.md", finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:                 []string{"sources/openclerk-runner.md"},
		RequireSearch:              true,
		RequireRecordsLookup:       true,
		RequireProvenanceEvents:    true,
		RequireProjectionStates:    true,
		Metrics:                    turnMetrics,
		FinalAnswerPath:            true,
		AdditionalBodyRequirements: []string{"records", "provenance", "projection"},
	})
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	records, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "OpenClerk runner", Limit: 5},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "records",
			RefKind:    "entity",
			RefID:      "openclerk-runner",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasRecord := records.Records != nil && len(records.Records.Entities) > 0
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness == "fresh"
	failures := []string{}
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if !hasRecord {
		failures = append(failures, "records lookup missing")
	}
	if !hasProjection {
		failures = append(failures, "projection state missing")
	}
	if !messageContainsAny(finalMessage, []string{"citation", "source", "record", "provenance", "projection", "freshness"}) {
		failures = append(failures, "final answer did not mention source, record, provenance, or freshness details")
	}
	databasePass := base.DatabasePass && hasRecord && hasProjection
	assistantPass := base.AssistantPass && messageContainsAny(finalMessage, []string{"citation", "source", "record", "provenance", "projection", "freshness"})
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{"synthesis/openclerk-runner-with-records.md"},
	}, nil
}

func verifyDocumentContains(ctx context.Context, paths evalPaths, docPath string, required []string, forbidden []string) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	if !found {
		return verificationResult{Passed: false, DatabasePass: false, Details: "missing " + docPath}, nil
	}
	failures := missingRequired(body, required)
	failures = append(failures, presentForbidden(body, forbidden)...)
	return verificationResult{
		Passed:        len(failures) == 0,
		DatabasePass:  len(failures) == 0,
		AssistantPass: true,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func documentByPath(ctx context.Context, paths evalPaths, docPath string) (*runner.Document, bool, error) {
	docID, found, err := documentIDByPath(ctx, paths, docPath)
	if err != nil || !found {
		return nil, found, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionGet, DocID: docID})
	if err != nil {
		return nil, false, err
	}
	if got.Document != nil {
		return got.Document, true, nil
	}
	return nil, false, nil
}

func documentBodyByPath(ctx context.Context, paths evalPaths, docPath string) (string, bool, error) {
	doc, found, err := documentByPath(ctx, paths, docPath)
	if err != nil || !found {
		return "", found, err
	}
	if doc != nil {
		return doc.Body, true, nil
	}
	return "", false, nil
}

func documentContaining(ctx context.Context, paths evalPaths, needle string) (string, string, bool, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 100},
	})
	if err != nil {
		return "", "", false, err
	}
	for _, entry := range list.Documents {
		doc, found, err := documentByPath(ctx, paths, entry.Path)
		if err != nil {
			return "", "", false, err
		}
		if !found || doc == nil {
			continue
		}
		if strings.Contains(doc.Body, needle) {
			return entry.Path, doc.Body, true, nil
		}
	}
	return "", "", false, nil
}

func documentIDByPath(ctx context.Context, paths evalPaths, docPath string) (string, bool, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 100},
	})
	if err != nil {
		return "", false, err
	}
	for _, doc := range list.Documents {
		if doc.Path == docPath {
			return doc.DocID, true, nil
		}
	}
	return "", false, nil
}

func exactDocumentCount(ctx context.Context, paths evalPaths, docPath string) (int, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 100},
	})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, doc := range list.Documents {
		if doc.Path == docPath {
			count++
		}
	}
	return count, nil
}

func documentCountWithPrefix(ctx context.Context, paths evalPaths, pathPrefix string) (int, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: pathPrefix, Limit: 100},
	})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, doc := range list.Documents {
		if strings.HasPrefix(doc.Path, pathPrefix) {
			count++
		}
	}
	return count, nil
}

func disallowedDocumentPathsWithPrefix(ctx context.Context, paths evalPaths, pathPrefix string, allowed map[string]bool) ([]string, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: pathPrefix, Limit: 100},
	})
	if err != nil {
		return nil, err
	}
	disallowed := []string{}
	for _, doc := range list.Documents {
		if strings.HasPrefix(doc.Path, pathPrefix) && !allowed[doc.Path] {
			disallowed = append(disallowed, doc.Path)
		}
	}
	sort.Strings(disallowed)
	return disallowed, nil
}

func firstSynthesisProjection(ctx context.Context, paths evalPaths, docID string) (*runner.ProjectionState, error) {
	if strings.TrimSpace(docID) == "" {
		return nil, nil
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      docID,
			Limit:      5,
		},
	})
	if err != nil {
		return nil, err
	}
	if projections.Projections == nil || len(projections.Projections.Projections) == 0 {
		return nil, nil
	}
	projection := projections.Projections.Projections[0]
	return &projection, nil
}

func projectionDetailContains(details map[string]string, key string, value string) bool {
	return strings.Contains(details[key], value)
}

func topSearchHit(result runner.RetrievalTaskResult) (runner.SearchHit, bool) {
	if result.Search == nil || len(result.Search.Hits) == 0 {
		return runner.SearchHit{}, false
	}
	return result.Search.Hits[0], true
}

func searchContainsPath(result runner.RetrievalTaskResult, path string) bool {
	if result.Search == nil {
		return false
	}
	for _, hit := range result.Search.Hits {
		if searchHitPath(hit) == path {
			return true
		}
	}
	return false
}

func searchResultHasCitations(result runner.RetrievalTaskResult) bool {
	if result.Search == nil || len(result.Search.Hits) == 0 {
		return false
	}
	for _, hit := range result.Search.Hits {
		if searchHitHasCitation(hit) {
			return true
		}
	}
	return false
}

func searchOnlyContainsPath(result runner.RetrievalTaskResult, path string) bool {
	if result.Search == nil || len(result.Search.Hits) == 0 {
		return false
	}
	for _, hit := range result.Search.Hits {
		if searchHitPath(hit) != path {
			return false
		}
	}
	return true
}

func searchHitPath(hit runner.SearchHit) string {
	if len(hit.Citations) > 0 {
		return hit.Citations[0].Path
	}
	return ""
}

func searchHitHasCitation(hit runner.SearchHit) bool {
	if hit.DocID == "" || hit.ChunkID == "" {
		return false
	}
	for _, citation := range hit.Citations {
		if citation.DocID != "" &&
			citation.ChunkID != "" &&
			citation.Path != "" &&
			citation.LineStart > 0 &&
			citation.LineEnd >= citation.LineStart {
			return true
		}
	}
	return false
}

func allPathsFound(found map[string]bool, expected []string) bool {
	for _, path := range expected {
		if !found[path] {
			return false
		}
	}
	return true
}

func missingRequired(body string, required []string) []string {
	failures := []string{}
	for _, value := range required {
		if !strings.Contains(body, value) {
			failures = append(failures, "missing "+value)
		}
	}
	return failures
}

func missingRequiredFold(body string, required []string) []string {
	failures := []string{}
	lowerBody := strings.ToLower(body)
	for _, value := range required {
		if !strings.Contains(lowerBody, strings.ToLower(value)) {
			failures = append(failures, "missing "+value)
		}
	}
	return failures
}

func sourceRefsFrontmatterFailures(body string, expected []string) []string {
	value, found, singleLine := sourceRefsFrontmatterValue(body)
	if !found {
		return []string{"missing source_refs frontmatter"}
	}
	if !singleLine {
		return []string{"source_refs must be single-line comma-separated frontmatter"}
	}
	refs := map[string]bool{}
	for _, ref := range strings.Split(value, ",") {
		normalized := strings.Trim(strings.TrimSpace(ref), `"'`)
		if normalized != "" {
			refs[normalized] = true
		}
	}
	failures := []string{}
	for _, ref := range expected {
		if !refs[ref] {
			failures = append(failures, "missing source ref "+ref)
		}
	}
	return failures
}

func sourceRefsFrontmatterValue(body string) (string, bool, bool) {
	lines := strings.Split(body, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", false, false
	}
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "source_refs") {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || strings.HasPrefix(value, "[") || strings.HasSuffix(value, "]") {
			return value, true, false
		}
		return value, true, true
	}
	return "", false, false
}

func presentForbidden(body string, forbidden []string) []string {
	failures := []string{}
	for _, value := range forbidden {
		if strings.Contains(body, value) {
			failures = append(failures, "unexpected "+value)
		}
	}
	return failures
}

func messageContainsAll(message string, values []string) bool {
	lower := normalizeValidationMessage(message)
	for _, value := range values {
		if !strings.Contains(lower, strings.ToLower(value)) {
			return false
		}
	}
	return true
}

func messageContainsAny(message string, values []string) bool {
	return containsAny(normalizeValidationMessage(message), lowerStrings(values))
}

func graphSemanticsReferenceAnswerPass(message string) bool {
	normalized := normalizeValidationMessage(message)
	if messagePromotesGraphSemantics(normalized) {
		return false
	}
	return containsAny(normalized, []string{"search"}) &&
		containsAny(normalized, []string{"document_links", "links", "link"}) &&
		containsAny(normalized, []string{"backlink", "incoming"}) &&
		containsAny(normalized, []string{"graph_neighborhood", "graph neighborhood"}) &&
		containsAny(normalized, []string{"markdown", "relationship text", "relationship wording"}) &&
		containsAny(normalized, []string{"citation", "cited", "source", "canonical", "derived"}) &&
		containsAny(normalized, []string{"projection", "fresh", "freshness"}) &&
		containsAny(normalized, []string{"reference", "defer", "deferred", "not promote", "do not promote", "not promoted", "keep"})
}

func messagePromotesGraphSemantics(normalized string) bool {
	promotionPhrases := []string{
		"decision: promote",
		"promote graph semantics",
		"promote richer graph",
		"promote semantic graph",
		"add semantic graph",
		"new graph authority",
		"independent semantic",
		"promote a semantic-label graph layer",
		"promote semantic-label graph layer",
		"semantic-label graph layer should be promoted",
	}
	for _, phrase := range promotionPhrases {
		if strings.Contains(normalized, phrase) &&
			!strings.Contains(normalized, "do not "+phrase) &&
			!strings.Contains(normalized, "not "+phrase) {
			return true
		}
	}
	return false
}

func memoryRouterReferenceAnswerPass(message string) bool {
	normalized := normalizeValidationMessage(message)
	if messagePromotesMemoryRouter(normalized) {
		return false
	}
	return containsAny(normalized, []string{"temporal", "current", "stale", "effective"}) &&
		containsAny(normalized, []string{"session promotion", "session-derived", "session observation", "canonical markdown", "canonicalization"}) &&
		containsAny(normalized, []string{"feedback", "weight", "weighted", "advisory"}) &&
		containsAny(normalized, []string{"routing", "route", "router"}) &&
		containsAny(normalized, []string{"source_refs", "source ref", "source refs", "citation", "cited", "source path"}) &&
		containsAny(normalized, []string{"freshness", "fresh", "provenance", "projection"}) &&
		containsAny(normalized, []string{"reference", "defer", "deferred", "not promote", "do not promote", "not promoted", "keep"})
}

func messagePromotesMemoryRouter(normalized string) bool {
	promotionPhrases := []string{
		"decision: promote memory",
		"decision: promote router",
		"decision: promote memory/router",
		"promote memory/router",
		"promote memory router",
		"promote autonomous routing",
		"promote remember",
		"promote recall",
		"add a memory interface",
		"add memory interface",
		"add a router interface",
		"add router interface",
		"add remember/recall",
		"new memory interface",
		"new router interface",
		"memory should outrank",
		"memory outranks canonical",
		"autonomous router should choose",
		"autonomous routing should choose",
	}
	for _, phrase := range promotionPhrases {
		if strings.Contains(normalized, phrase) &&
			!strings.Contains(normalized, "do not "+phrase) &&
			!strings.Contains(normalized, "not "+phrase) &&
			!strings.Contains(normalized, "without "+phrase) {
			return true
		}
	}
	return false
}

func messageReportsLayoutValid(message string) bool {
	normalized := normalizeValidationMessage(message)
	if layoutInvalidStatusPattern.MatchString(normalized) {
		return false
	}
	if layoutExplicitValidPattern.MatchString(normalized) {
		return true
	}
	return layoutValidStatusPattern.MatchString(normalized)
}

func containsAllStrings(values []string, expected []string) bool {
	present := map[string]bool{}
	for _, value := range values {
		present[value] = true
	}
	for _, value := range expected {
		if !present[value] {
			return false
		}
	}
	return true
}

func documentLinksContainPath(links []runner.DocumentLink, path string) bool {
	for _, link := range links {
		if link.Path == path {
			return true
		}
	}
	return false
}

func documentLinksHaveCitations(links []runner.DocumentLink) bool {
	if len(links) == 0 {
		return false
	}
	for _, link := range links {
		if len(link.Citations) == 0 {
			return false
		}
		for _, citation := range link.Citations {
			if citation.DocID == "" || citation.ChunkID == "" || citation.Path == "" || citation.LineStart == 0 {
				return false
			}
		}
	}
	return true
}

func graphContainsNodeLabels(nodes []runner.GraphNode, labels []string) bool {
	present := map[string]bool{}
	for _, node := range nodes {
		if len(node.Citations) > 0 {
			present[node.Label] = true
		}
	}
	for _, label := range labels {
		if !present[label] {
			return false
		}
	}
	return true
}

func graphContainsLinkEdge(edges []runner.GraphEdge) bool {
	for _, edge := range edges {
		if edge.Kind == "links_to" {
			return true
		}
	}
	return false
}

func graphContainsStructuralEdge(edges []runner.GraphEdge) bool {
	for _, edge := range edges {
		if edge.Kind == "links_to" || edge.Kind == "mentions" {
			return true
		}
	}
	return false
}

func graphEdgesOnlyStructural(edges []runner.GraphEdge) bool {
	if len(edges) == 0 {
		return false
	}
	for _, edge := range edges {
		if edge.Kind != "links_to" && edge.Kind != "mentions" {
			return false
		}
	}
	return true
}

func graphEdgesHaveCitations(edges []runner.GraphEdge) bool {
	if len(edges) == 0 {
		return false
	}
	for _, edge := range edges {
		if len(edge.Citations) == 0 {
			return false
		}
		for _, citation := range edge.Citations {
			if citation.DocID == "" || citation.ChunkID == "" || citation.Path == "" || citation.LineStart == 0 {
				return false
			}
		}
	}
	return true
}

func layoutChecksInclude(checks []runner.KnowledgeLayoutCheck, id string, status string) bool {
	for _, check := range checks {
		if check.ID == id && check.Status == status {
			return true
		}
	}
	return false
}

func eventTypesInclude(events []runner.ProvenanceEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func provenanceEventRefIDsInclude(actual []string, expected ...string) bool {
	return stringValuesInclude(actual, expected...)
}

func decisionRecordIDsInclude(actual []string, expected ...string) bool {
	return stringValuesInclude(actual, expected...)
}

func stringValuesInclude(actual []string, expected ...string) bool {
	seen := map[string]bool{}
	for _, value := range actual {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized != "" {
			seen[normalized] = true
		}
	}
	for _, value := range expected {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" || !seen[normalized] {
			return false
		}
	}
	return true
}

func lowerStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strings.ToLower(value))
	}
	return out
}

func verifyRecordsAndProvenance(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	records, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "OpenClerk runner", Limit: 5},
	})
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "records",
			RefKind:    "entity",
			RefID:      "openclerk-runner",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasRecord := records.Records != nil && len(records.Records.Entities) > 0
	hasProvenance := provenance.Provenance != nil && len(provenance.Provenance.Events) > 0
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness == "fresh"
	activityPass := turnMetrics.RecordsLookupUsed && turnMetrics.ProvenanceEventsUsed && turnMetrics.ProjectionStatesUsed
	assistantPass := messageContainsAny(finalMessage, []string{"provenance", "event"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness", "fresh", "stale"})
	failures := []string{}
	if !hasRecord {
		failures = append(failures, "records lookup missing")
	}
	if !hasProvenance {
		failures = append(failures, "provenance events missing")
	}
	if !hasProjection {
		failures = append(failures, "projection state missing")
	}
	if !turnMetrics.RecordsLookupUsed {
		failures = append(failures, "agent did not use records lookup")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not mention provenance and projection freshness")
	}
	return verificationResult{
		Passed:        hasRecord && hasProvenance && hasProjection && activityPass && assistantPass,
		DatabasePass:  hasRecord && hasProvenance && hasProjection,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
	}, nil
}

func missingDetails(values []string) string {
	if len(values) == 0 {
		return "ok"
	}
	return strings.Join(values, "; ")
}

func verificationFromFailures(failures []string, passDetail string, documents []string) (verificationResult, error) {
	passed := len(failures) == 0
	details := passDetail
	if !passed {
		details = missingDetails(failures)
	}
	return verificationResult{
		Passed:        passed,
		DatabasePass:  passed,
		AssistantPass: passed,
		Details:       details,
		Documents:     documents,
	}, nil
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func parseMetrics(eventsPath string) (parsedTurn, error) {
	file, err := os.Open(eventsPath)
	if err != nil {
		return parsedTurn{metrics: emptyMetrics()}, err
	}
	defer func() { _ = file.Close() }()
	out := parsedTurn{metrics: emptyMetrics()}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	inputTotal := 0
	cachedTotal := 0
	outputTotal := 0
	usageExposed := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.Type != "" {
			out.metrics.EventTypeCounts[event.Type]++
		}
		if event.ThreadID != "" {
			out.sessionID = event.ThreadID
		}
		itemText := string(event.Item)
		if strings.Contains(itemText, "run document task: download source PDF") {
			out.metrics.SourcePDFDownloadFailure = true
		}
		if event.Usage != nil {
			usageExposed = true
			input, cached, output := usageNumbers(*event.Usage)
			inputTotal += input
			cachedTotal += cached
			outputTotal += output
		}
		if event.Type == "message" || strings.Contains(itemText, `"type":"message"`) || strings.Contains(itemText, `"type":"agent_message"`) {
			if strings.Contains(itemText, `"role":"assistant"`) || strings.Contains(itemText, `"type":"message"`) || strings.Contains(itemText, `"type":"agent_message"`) {
				out.metrics.AssistantCalls++
				if msg := extractAssistantText(event.Item); msg != "" {
					out.finalMessage = msg
				}
			}
		}
		commands := commandTexts(event.Item)
		if len(commands) > 0 {
			out.metrics.ToolCalls += len(commands)
		} else if event.Type == "tool_call" || strings.Contains(itemText, `"type":"tool_call"`) || strings.Contains(itemText, `"call_id"`) {
			out.metrics.ToolCalls++
		}
		for _, command := range commands {
			out.metrics.CommandExecutions++
			classifyCommand(command, &out.metrics)
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	if usageExposed {
		nonCached := inputTotal - cachedTotal
		if nonCached < 0 {
			nonCached = 0
		}
		out.metrics.UsageExposed = true
		out.metrics.InputTokens = &inputTotal
		out.metrics.CachedInputTokens = &cachedTotal
		out.metrics.NonCachedInputTokens = &nonCached
		out.metrics.OutputTokens = &outputTotal
	}
	return out, nil
}

func emptyMetrics() metrics {
	return metrics{
		EventTypeCounts:          map[string]int{},
		CommandMetricLimitations: "Command/file inspection metrics are inferred from codex exec JSON command events, not OS-level tracing.",
	}
}

func usageNumbers(value usage) (input int, cached int, output int) {
	input = value.InputTokens
	if input == 0 {
		input = value.PromptTokens
	}
	output = value.OutputTokens
	if output == 0 {
		output = value.CompletionTokens
	}
	cached = value.CachedInputTokens
	if value.InputTokensDetails != nil {
		cached += value.InputTokensDetails.CachedTokens
	}
	if value.PromptDetails != nil {
		cached += value.PromptDetails.CachedTokens
	}
	return input, cached, output
}

func extractAssistantText(raw json.RawMessage) string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	texts := []string{}
	collectTextValues(value, &texts)
	if len(texts) == 0 {
		return ""
	}
	return strings.Join(texts, "\n")
}

func collectTextValues(value any, texts *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		if role, _ := typed["role"].(string); role == "assistant" {
			if content, ok := typed["content"].(string); ok && strings.TrimSpace(content) != "" {
				*texts = append(*texts, content)
			}
		}
		if typ, _ := typed["type"].(string); typ == "agent_message" {
			if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
				*texts = append(*texts, text)
			}
		}
		if typ, _ := typed["type"].(string); typ == "output_text" || typ == "text" {
			if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
				*texts = append(*texts, text)
			}
		}
		for _, nested := range typed {
			collectTextValues(nested, texts)
		}
	case []any:
		for _, nested := range typed {
			collectTextValues(nested, texts)
		}
	}
}

func commandTexts(raw json.RawMessage) []string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	out := []string{}
	collectCommandTexts(value, &out)
	return out
}

func collectCommandTexts(value any, out *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"cmd", "command"} {
			switch command := typed[key].(type) {
			case string:
				if command != "" {
					*out = append(*out, command)
				}
			case []any:
				parts := []string{}
				for _, part := range command {
					if s, ok := part.(string); ok {
						parts = append(parts, s)
					}
				}
				if len(parts) > 0 {
					*out = append(*out, strings.Join(parts, " "))
				}
			}
		}
		for _, nested := range typed {
			collectCommandTexts(nested, out)
		}
	case []any:
		for _, nested := range typed {
			collectCommandTexts(nested, out)
		}
	}
}

func classifyCommand(command string, m *metrics) {
	lower := strings.ToLower(command)
	actionText := strings.ReplaceAll(lower, `\"`, `"`)
	evidence := sanitizeMetricEvidence(command)
	addEvidence := func(target *[]string) {
		if len(*target) < 6 {
			*target = append(*target, evidence)
		}
	}
	if strings.Contains(command, "client.gen.go") || strings.Contains(command, "openapi.gen.go") || strings.Contains(command, "internal/api/openapi.gen.go") {
		m.GeneratedFileInspection = true
		addEvidence(&m.GeneratedFileEvidence)
	}
	if strings.Contains(command, "GOMODCACHE") || strings.Contains(command, "/pkg/mod") || strings.Contains(command, "go env GOMODCACHE") {
		m.ModuleCacheInspection = true
		addEvidence(&m.ModuleCacheEvidence)
	}
	if strings.Contains(command, "rg --files") || isBroadFindCommand(command) {
		m.BroadRepoSearch = true
		addEvidence(&m.BroadRepoSearchEvidence)
	}
	if strings.Contains(lower, "sqlite3") || strings.Contains(lower, "select ") || strings.Contains(lower, "pragma ") {
		m.DirectSQLiteAccess = true
		addEvidence(&m.DirectSQLiteEvidence)
	}
	if isFileInspectionCommand(lower) {
		m.FileInspectionCommands++
	}
	if strings.Contains(command, "go run ./cmd/openclerk ") || strings.Contains(command, "go run ./cmd/openclerk\n") || strings.Contains(command, " ./cmd/openclerk ") {
		m.LegacyRunnerUsage = true
		addEvidence(&m.LegacyRunnerEvidence)
	}
	classifySearchCommand(actionText, m)
	if commandContainsAction(actionText, "ingest_source_url") {
		m.IngestSourceURLUsed = true
		if actionHasFieldValue(actionText, "ingest_source_url", "mode", "update") {
			m.IngestSourceURLUpdateUsed = true
		}
	}
	if commandContainsAction(actionText, "ingest_video_url") {
		m.IngestVideoURLUsed = true
		if actionHasFieldValue(actionText, "ingest_video_url", "mode", "update") {
			m.IngestVideoURLUpdateUsed = true
		}
	}
	if commandContainsAction(actionText, "validate") {
		m.ValidateUsed = true
	}
	if commandContainsAction(actionText, "create_document") {
		m.CreateDocumentUsed = true
	}
	if commandContainsAction(actionText, "replace_section") {
		m.ReplaceSectionUsed = true
	}
	if commandContainsAction(actionText, "append_document") {
		m.AppendDocumentUsed = true
	}
	if commandContainsAction(actionText, "list_documents") {
		m.ListDocumentsUsed = true
		m.ListDocumentPathPrefixes = append(m.ListDocumentPathPrefixes, actionFieldValues(actionText, "list_documents", "path_prefix")...)
	}
	if commandContainsAction(actionText, "get_document") {
		m.GetDocumentUsed = true
		m.GetDocumentDocIDs = append(m.GetDocumentDocIDs, actionFieldValues(actionText, "get_document", "doc_id")...)
	}
	if commandContainsAction(actionText, "inspect_layout") {
		m.InspectLayoutUsed = true
	}
	if commandContainsAction(actionText, "document_links") {
		m.DocumentLinksUsed = true
	}
	if commandContainsAction(actionText, "graph_neighborhood") {
		m.GraphNeighborhoodUsed = true
	}
	if commandContainsAction(actionText, "records_lookup") {
		m.RecordsLookupUsed = true
	}
	if commandContainsAction(actionText, "decisions_lookup") {
		m.DecisionsLookupUsed = true
	}
	if commandContainsAction(actionText, "decision_record") {
		m.DecisionRecordUsed = true
		m.DecisionRecordIDs = append(m.DecisionRecordIDs, actionFieldValues(actionText, "decision_record", "decision_id")...)
	}
	if commandContainsAction(actionText, "provenance_events") {
		m.ProvenanceEventsUsed = true
		m.ProvenanceEventRefIDs = append(m.ProvenanceEventRefIDs, actionRefIDs(actionText, "provenance_events")...)
	}
	if commandContainsAction(actionText, "projection_states") {
		m.ProjectionStatesUsed = true
	}
}

func commandContainsAction(actionText string, action string) bool {
	compacted := strings.Join(strings.Fields(actionText), "")
	return strings.Contains(compacted, `"action":"`+action+`"`)
}

func actionRefIDs(actionText string, action string) []string {
	return actionFieldValues(actionText, action, "ref_id")
}

func actionFieldValues(actionText string, action string, field string) []string {
	compacted := strings.Join(strings.Fields(actionText), "")
	marker := `"action":"` + action + `"`
	values := []string{}
	for _, part := range strings.Split(compacted, marker)[1:] {
		if next := strings.Index(part, `"action":"`); next >= 0 {
			part = part[:next]
		}
		fieldMarker := `"` + field + `":"`
		valueStart := strings.Index(part, fieldMarker)
		if valueStart < 0 {
			continue
		}
		valueStart += len(fieldMarker)
		valueEnd := strings.Index(part[valueStart:], `"`)
		if valueEnd < 0 {
			continue
		}
		value := strings.TrimSpace(part[valueStart : valueStart+valueEnd])
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func actionHasFieldValue(actionText string, action string, field string, value string) bool {
	for _, got := range actionFieldValues(actionText, action, field) {
		if got == value {
			return true
		}
	}
	return false
}

func classifySearchCommand(actionText string, m *metrics) {
	compacted := strings.Join(strings.Fields(actionText), "")
	const marker = `"action":"search"`
	if !strings.Contains(compacted, marker) {
		return
	}
	m.SearchUsed = true
	parts := strings.Split(compacted, marker)
	for _, part := range parts[1:] {
		if next := strings.Index(part, `"action":"`); next >= 0 {
			part = part[:next]
		}
		hasPathFilter := strings.Contains(part, `"path_prefix":`)
		hasMetadataFilter := strings.Contains(part, `"metadata_key":`) || strings.Contains(part, `"metadata_value":`)
		if hasPathFilter {
			m.SearchPathFilterUsed = true
			m.SearchPathPrefixes = append(m.SearchPathPrefixes, fieldValueFromCompactedAction(part, "path_prefix"))
		}
		if hasMetadataFilter {
			m.SearchMetadataFilterUsed = true
			key := fieldValueFromCompactedAction(part, "metadata_key")
			value := fieldValueFromCompactedAction(part, "metadata_value")
			if key != "" || value != "" {
				m.SearchMetadataFilters = append(m.SearchMetadataFilters, key+"="+value)
			}
		}
		if !hasPathFilter && !hasMetadataFilter {
			m.SearchUnfilteredUsed = true
		}
	}
}

func fieldValueFromCompactedAction(part string, field string) string {
	fieldMarker := `"` + field + `":"`
	valueStart := strings.Index(part, fieldMarker)
	if valueStart < 0 {
		return ""
	}
	valueStart += len(fieldMarker)
	valueEnd := strings.Index(part[valueStart:], `"`)
	if valueEnd < 0 {
		return ""
	}
	return strings.TrimSpace(part[valueStart : valueStart+valueEnd])
}

func sanitizeMetricEvidence(value string) string {
	replacements := []string{}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		replacements = append(replacements, home, "<home>")
	}
	if tmp := strings.TrimSpace(os.TempDir()); tmp != "" {
		replacements = append(replacements, tmp, "<tmp>")
	}
	if len(replacements) == 0 {
		return sanitizeKnownHomePrefixes(value)
	}
	return sanitizeKnownHomePrefixes(strings.NewReplacer(replacements...).Replace(value))
}

func sanitizeKnownHomePrefixes(value string) string {
	value = unixHomePathPattern.ReplaceAllString(value, "<home>")
	return windowsHomePathPattern.ReplaceAllString(value, "<home>")
}

func isFileInspectionCommand(command string) bool {
	for _, prefix := range []string{"cat ", "sed ", "nl ", "head ", "tail ", "less ", "grep ", "rg "} {
		if strings.HasPrefix(strings.TrimSpace(command), prefix) {
			return true
		}
	}
	return false
}

func isBroadFindCommand(command string) bool {
	trimmed := strings.TrimSpace(command)
	if !strings.Contains(trimmed, "find .") && !strings.Contains(trimmed, "find ..") {
		return false
	}
	if strings.Contains(trimmed, "-type d") && !strings.Contains(trimmed, "-type f") {
		return false
	}
	return true
}

func aggregateMetrics(turns []turnResult) metrics {
	out := emptyMetrics()
	allUsageExposed := len(turns) > 0
	inputTotal := 0
	cachedTotal := 0
	nonCachedTotal := 0
	outputTotal := 0
	for _, turn := range turns {
		current := turn.Metrics
		out.AssistantCalls += current.AssistantCalls
		out.ToolCalls += current.ToolCalls
		out.CommandExecutions += current.CommandExecutions
		out.FileInspectionCommands += current.FileInspectionCommands
		out.GeneratedFileInspection = out.GeneratedFileInspection || current.GeneratedFileInspection
		out.ModuleCacheInspection = out.ModuleCacheInspection || current.ModuleCacheInspection
		out.BroadRepoSearch = out.BroadRepoSearch || current.BroadRepoSearch
		out.DirectSQLiteAccess = out.DirectSQLiteAccess || current.DirectSQLiteAccess
		out.LegacyRunnerUsage = out.LegacyRunnerUsage || current.LegacyRunnerUsage
		out.SearchUsed = out.SearchUsed || current.SearchUsed
		out.SearchUnfilteredUsed = out.SearchUnfilteredUsed || current.SearchUnfilteredUsed
		out.SearchPathFilterUsed = out.SearchPathFilterUsed || current.SearchPathFilterUsed
		out.SearchPathPrefixes = append(out.SearchPathPrefixes, current.SearchPathPrefixes...)
		out.SearchMetadataFilterUsed = out.SearchMetadataFilterUsed || current.SearchMetadataFilterUsed
		out.SearchMetadataFilters = append(out.SearchMetadataFilters, current.SearchMetadataFilters...)
		out.IngestSourceURLUsed = out.IngestSourceURLUsed || current.IngestSourceURLUsed
		out.IngestSourceURLUpdateUsed = out.IngestSourceURLUpdateUsed || current.IngestSourceURLUpdateUsed
		out.IngestVideoURLUsed = out.IngestVideoURLUsed || current.IngestVideoURLUsed
		out.IngestVideoURLUpdateUsed = out.IngestVideoURLUpdateUsed || current.IngestVideoURLUpdateUsed
		out.SourcePDFDownloadFailure = out.SourcePDFDownloadFailure || current.SourcePDFDownloadFailure
		out.ValidateUsed = out.ValidateUsed || current.ValidateUsed
		out.CreateDocumentUsed = out.CreateDocumentUsed || current.CreateDocumentUsed
		out.ReplaceSectionUsed = out.ReplaceSectionUsed || current.ReplaceSectionUsed
		out.AppendDocumentUsed = out.AppendDocumentUsed || current.AppendDocumentUsed
		out.ListDocumentsUsed = out.ListDocumentsUsed || current.ListDocumentsUsed
		out.ListDocumentPathPrefixes = append(out.ListDocumentPathPrefixes, current.ListDocumentPathPrefixes...)
		out.GetDocumentUsed = out.GetDocumentUsed || current.GetDocumentUsed
		out.GetDocumentDocIDs = append(out.GetDocumentDocIDs, current.GetDocumentDocIDs...)
		out.InspectLayoutUsed = out.InspectLayoutUsed || current.InspectLayoutUsed
		out.DocumentLinksUsed = out.DocumentLinksUsed || current.DocumentLinksUsed
		out.GraphNeighborhoodUsed = out.GraphNeighborhoodUsed || current.GraphNeighborhoodUsed
		out.RecordsLookupUsed = out.RecordsLookupUsed || current.RecordsLookupUsed
		out.DecisionsLookupUsed = out.DecisionsLookupUsed || current.DecisionsLookupUsed
		out.DecisionRecordUsed = out.DecisionRecordUsed || current.DecisionRecordUsed
		out.DecisionRecordIDs = append(out.DecisionRecordIDs, current.DecisionRecordIDs...)
		out.ProvenanceEventsUsed = out.ProvenanceEventsUsed || current.ProvenanceEventsUsed
		out.ProvenanceEventRefIDs = append(out.ProvenanceEventRefIDs, current.ProvenanceEventRefIDs...)
		out.ProjectionStatesUsed = out.ProjectionStatesUsed || current.ProjectionStatesUsed
		out.GeneratedFileEvidence = append(out.GeneratedFileEvidence, current.GeneratedFileEvidence...)
		out.ModuleCacheEvidence = append(out.ModuleCacheEvidence, current.ModuleCacheEvidence...)
		out.BroadRepoSearchEvidence = append(out.BroadRepoSearchEvidence, current.BroadRepoSearchEvidence...)
		out.DirectSQLiteEvidence = append(out.DirectSQLiteEvidence, current.DirectSQLiteEvidence...)
		out.LegacyRunnerEvidence = append(out.LegacyRunnerEvidence, current.LegacyRunnerEvidence...)
		for eventType, count := range current.EventTypeCounts {
			out.EventTypeCounts[eventType] += count
		}
		if !current.UsageExposed || current.InputTokens == nil || current.CachedInputTokens == nil || current.NonCachedInputTokens == nil || current.OutputTokens == nil {
			allUsageExposed = false
			continue
		}
		inputTotal += *current.InputTokens
		cachedTotal += *current.CachedInputTokens
		nonCachedTotal += *current.NonCachedInputTokens
		outputTotal += *current.OutputTokens
	}
	if allUsageExposed {
		out.UsageExposed = true
		out.InputTokens = &inputTotal
		out.CachedInputTokens = &cachedTotal
		out.NonCachedInputTokens = &nonCachedTotal
		out.OutputTokens = &outputTotal
	}
	return out
}

func aggregateVerification(sc scenario, turns []turnResult) verificationResult {
	out := verificationResult{Passed: true, DatabasePass: true, AssistantPass: true}
	details := []string{}
	for _, turn := range turns {
		verification := turn.Verification
		if !verification.Passed {
			out.Passed = false
		}
		if !verification.DatabasePass {
			out.DatabasePass = false
		}
		if !verification.AssistantPass {
			out.AssistantPass = false
		}
		if verification.Details != "" {
			details = append(details, fmt.Sprintf("turn %d: %s", turn.Index, verification.Details))
		}
		out.Documents = verification.Documents
	}
	if len(details) > 0 {
		out.Details = strings.Join(details, "; ")
	}
	if len(turns) == 0 {
		out = verificationResult{Passed: false, DatabasePass: false, AssistantPass: false, Details: fmt.Sprintf("scenario %s did not run", sc.ID)}
	}
	return out
}

func aggregateExitCode(turns []turnResult) int {
	for _, turn := range turns {
		if turn.ExitCode != 0 {
			return turn.ExitCode
		}
	}
	return 0
}

func buildProductionGateSummary(results []jobResult) *productionGateSummary {
	productionByScenario := map[string]jobResult{}
	for _, result := range results {
		if result.Variant == productionVariant {
			productionByScenario[result.Scenario] = result
		}
	}
	if len(productionByScenario) == 0 {
		return nil
	}
	productionPassedAll := true
	noGenerated := true
	noModuleCache := true
	noBroadSearch := true
	noLegacyRunnerUsage := true
	noDirectSQLite := true
	validationFinalAnswerOnly := true
	validationFailures := []string{}
	missingValidationScenarios := []string{}
	expectedScenarioIDs := releaseBlockingScenarioIDs()
	passedExpectedScenarios := 0
	missingProductionScenarios := []string{}
	for _, scenarioID := range expectedScenarioIDs {
		production, ok := productionByScenario[scenarioID]
		if !ok {
			productionPassedAll = false
			missingProductionScenarios = append(missingProductionScenarios, scenarioID)
			if isFinalAnswerOnlyValidationScenario(scenarioID) {
				validationFinalAnswerOnly = false
				missingValidationScenarios = append(missingValidationScenarios, scenarioID)
			}
			continue
		}
		if !production.Passed {
			productionPassedAll = false
		} else {
			passedExpectedScenarios++
		}
		if production.Metrics.GeneratedFileInspection {
			noGenerated = false
		}
		if production.Metrics.ModuleCacheInspection {
			noModuleCache = false
		}
		if production.Metrics.BroadRepoSearch {
			noBroadSearch = false
		}
		if production.Metrics.LegacyRunnerUsage {
			noLegacyRunnerUsage = false
		}
		if production.Metrics.DirectSQLiteAccess {
			noDirectSQLite = false
		}
		if isFinalAnswerOnlyValidationScenario(production.Scenario) &&
			(production.Metrics.ToolCalls != 0 || production.Metrics.CommandExecutions != 0 || production.Metrics.AssistantCalls > 1) {
			validationFinalAnswerOnly = false
			validationFailures = append(validationFailures, production.Scenario)
		}
	}
	criteria := []productionGateCriterion{
		{Name: "production_passes_all_scenarios", Passed: productionPassedAll, Details: productionScenariosDetails(passedExpectedScenarios, len(expectedScenarioIDs), missingProductionScenarios)},
		{Name: "no_direct_generated_file_inspection", Passed: noGenerated, Details: "production must not inspect retired API files or generated server files"},
		{Name: "no_module_cache_inspection", Passed: noModuleCache, Details: "production must not inspect the Go module cache"},
		{Name: "no_broad_repo_search", Passed: noBroadSearch, Details: "production must not use broad repo search in routine OpenClerk knowledge tasks"},
		{Name: "no_legacy_source_runner_usage", Passed: noLegacyRunnerUsage, Details: "production must not invoke source-built or legacy runner paths instead of installed openclerk"},
		{Name: "no_direct_sqlite_access", Passed: noDirectSQLite, Details: "production must not query SQLite directly"},
		{Name: "validation_scenarios_are_final_answer_only", Passed: validationFinalAnswerOnly, Details: validationFinalAnswerDetails(validationFailures, missingValidationScenarios)},
	}
	passes := true
	for _, criterion := range criteria {
		if !criterion.Passed {
			passes = false
			break
		}
	}
	recommendation := "fix_production_agentops_before_release"
	if passes {
		recommendation = "use_agentops_runner_for_routine_openclerk_operations"
	}
	return &productionGateSummary{
		Variant:        productionVariant,
		PassesGate:     passes,
		Recommendation: recommendation,
		Criteria:       criteria,
	}
}

func buildTargetedLaneSummary(lane string, releaseBlocking bool, results []jobResult) *targetedLaneSummary {
	if releaseBlocking {
		return nil
	}
	if lane != populatedLaneName && lane != repoDocsLaneName && lane != agentChosenPathLaneName && lane != pathTitleAutonomyLaneName && lane != sourceURLUpdateLaneName && lane != documentThisLaneName && lane != documentArtifactCandidateLaneName && lane != artifactIngestionLaneName && lane != videoYouTubeLaneName {
		return nil
	}
	summary := targetedLaneSummary{
		Lane:            lane,
		PublicSurface:   []string{"openclerk document", "openclerk retrieval"},
		ReleaseBlocking: releaseBlocking,
	}
	if lane == documentArtifactCandidateLaneName {
		summary.PublicSurface = []string{"skills/openclerk/SKILL.md", "openclerk document", "openclerk retrieval"}
	}
	for _, result := range results {
		include := false
		classification, posture := "", ""
		switch lane {
		case populatedLaneName:
			include = isPopulatedVaultScenario(result.Scenario)
			classification, posture = classifyTargetedPopulatedResult(result)
		case repoDocsLaneName:
			include = isRepoDocsDogfoodScenario(result.Scenario)
			classification, posture = classifyTargetedRepoDocsResult(result)
		case agentChosenPathLaneName:
			include = isAgentChosenPathScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedAgentChosenPathResult(result)
		case pathTitleAutonomyLaneName:
			include = isPathTitleAutonomyScenario(result.Scenario)
			classification, posture = classifyTargetedPathTitleAutonomyResult(result)
		case sourceURLUpdateLaneName:
			include = isSourceURLUpdateScenario(result.Scenario)
			classification, posture = classifyTargetedSourceURLUpdateResult(result)
		case documentThisLaneName:
			include = isDocumentThisScenario(result.Scenario)
			classification, posture = classifyTargetedDocumentThisResult(result)
		case documentArtifactCandidateLaneName:
			include = isDocumentArtifactCandidateScenario(result.Scenario)
			classification, posture = classifyTargetedDocumentArtifactCandidateResult(result)
		case artifactIngestionLaneName:
			include = isArtifactIngestionScenario(result.Scenario)
			classification, posture = classifyTargetedArtifactIngestionResult(result)
		case videoYouTubeLaneName:
			include = isVideoYouTubeScenario(result.Scenario)
			classification, posture = classifyTargetedVideoYouTubeResult(result)
		}
		if !include {
			continue
		}
		summary.ScenarioClassifications = append(summary.ScenarioClassifications, targetedScenarioClassification{
			Variant:               result.Variant,
			Scenario:              result.Scenario,
			Status:                result.Status,
			FailureClassification: classification,
			EvidencePosture:       posture,
			ToolCalls:             result.Metrics.ToolCalls,
			CommandExecutions:     result.Metrics.CommandExecutions,
			AssistantCalls:        result.Metrics.AssistantCalls,
			WallSeconds:           result.WallSeconds,
			PromptSpecificity:     promptSpecificity(result.Scenario),
			UX:                    scenarioUX(result),
			Brittleness:           scenarioBrittleness(result),
			Retries:               scenarioRetries(result),
			StepCount:             scenarioStepCount(result),
			Latency:               scenarioLatency(result),
			GuidanceDependence:    scenarioGuidanceDependence(result),
			SafetyRisks:           scenarioSafetyRisks(result),
			FixturePreflight:      fixturePreflightStatus(result.FixturePreflight),
		})
	}
	if len(summary.ScenarioClassifications) == 0 {
		return nil
	}
	switch lane {
	case populatedLaneName:
		summary.Decision = "keep_as_reference"
		summary.Promotion = "no promoted runner action, schema, migration, storage API, product behavior, or public OpenClerk interface"
	case repoDocsLaneName:
		summary.Decision = "keep_as_public_dogfood_lane"
		summary.Promotion = "targeted repo-docs dogfood evidence only; no promoted runner action, schema, migration, storage API, product behavior, or public OpenClerk interface"
	case agentChosenPathLaneName:
		summary.Decision = agentChosenPathDecision(summary.ScenarioClassifications)
		summary.Promotion = "no promoted runner action, schema, migration, storage API, product behavior, public OpenClerk interface, or change to missing-path clarification"
	case pathTitleAutonomyLaneName:
		summary.Decision = "evaluate_for_oc_iat"
		summary.Promotion = "no promoted runner action, schema, migration, skill behavior, storage API, product behavior, or public OpenClerk interface from this eval"
	case sourceURLUpdateLaneName:
		summary.Decision = "keep_existing_update_mode"
		summary.Promotion = "targeted AgentOps evidence for existing ingest_source_url source.mode update behavior; no new runner action, schema, storage API, or transport"
	case documentThisLaneName:
		summary.Decision = "evaluate_for_oc_99z"
		summary.Promotion = "no promoted runner action, schema, migration, skill behavior, storage API, product behavior, or public OpenClerk interface from this eval"
	case documentArtifactCandidateLaneName:
		summary.Decision = documentArtifactCandidateDecision(summary.ScenarioClassifications)
		switch summary.Decision {
		case "promote_propose_before_create_skill_policy":
			summary.Promotion = "skill policy supports propose-before-create candidate path/title/body generation only; no runner action, schema, storage, migration, direct create, or public API change"
		case "defer_for_candidate_ergonomics_repair":
			summary.Promotion = "ergonomics promotion deferred; existing shipped propose-before-create skill policy needs natural-intent repair before oc-99z can promote it; no runner action, schema, storage, migration, direct create, or public API change"
		default:
			summary.Promotion = "no promoted skill policy yet; repair candidate quality gaps before any propose-before-create skill behavior change"
		}
	case artifactIngestionLaneName:
		summary.Decision = artifactIngestionDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted evidence only; no promoted runner action, parser, schema, storage migration, direct create behavior, or public API change"
	case videoYouTubeLaneName:
		summary.Decision = videoYouTubeDecision(summary.ScenarioClassifications)
		summary.Promotion = "keep supplied-transcript ingest_video_url as the promoted surface; native acquisition dependencies remain deferred"
	}
	return &summary
}

func classifyTargetedArtifactIngestionResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		return "none", "current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries"
	}
	if result.FixturePreflight != nil && !result.FixturePreflight.Passed {
		return "data_hygiene", "PDF fixture preflight failed before agent behavior could be evaluated: " + result.FixturePreflight.Details
	}
	if len(artifactIngestionBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if isFinalAnswerOnlyValidationScenario(result.Scenario) &&
		(result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1) {
		return "skill_guidance", "unsupported or missing-field artifact pressure did not stay final-answer-only"
	}
	if result.Verification.Passed {
		return "eval_contract_violation", "scenario verification passed, but the job did not complete successfully"
	}
	if isArtifactPDFScenario(result.Scenario) && !result.Verification.DatabasePass && result.FixturePreflight != nil && result.FixturePreflight.Passed {
		if result.Metrics.SourcePDFDownloadFailure {
			return "eval_coverage", "PDF fixture preflight worked, but the agent-runner process could not reach the generated HTTP PDF URL"
		}
		if result.Scenario == artifactPDFNaturalIntentScenarioID {
			return "ergonomics_gap", "scripted PDF fixture preflight worked, but natural user intent did not produce durable source evidence"
		}
		return "runner_capability_gap", "scripted PDF source URL control used the supported primitive but durable source evidence was missing"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene", "fixture or durable artifact evidence did not satisfy heterogeneous artifact pressure"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance", "runner-visible evidence existed, but the assistant answer did not satisfy heterogeneous artifact pressure"
	}
	return "runner_capability_gap", "manual review required before any generalized artifact ingestion surface promotion"
}

func classifyTargetedVideoYouTubeResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		return "none", "ingest_video_url preserved supplied video transcript authority, citations, provenance, freshness, and bypass boundaries"
	}
	if len(videoYouTubeBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if isFinalAnswerOnlyValidationScenario(result.Scenario) &&
		(result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1) {
		return "eval_contract_violation", "video/YouTube unsupported or bypass pressure did not stay final-answer-only"
	}
	if result.Verification.Passed {
		return "eval_contract_violation", "scenario verification passed, but the job did not complete successfully"
	}
	if result.Scenario == videoYouTubeScriptedTranscriptControlID && !result.Verification.DatabasePass {
		return "runner_capability_gap", "scripted supplied-transcript control could not produce durable canonical source evidence"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance", "runner-visible video/YouTube evidence existed, but the assistant answer did not satisfy the scenario"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene", "fixture or durable video/YouTube evidence did not satisfy targeted pressure"
	}
	return "ergonomics_gap", "manual review required before any video/YouTube ingestion promotion"
}

func promptSpecificity(scenarioID string) string {
	switch scenarioID {
	case candidateErgonomicsNaturalIntentScenarioID, candidateErgonomicsDuplicateNaturalID, candidateErgonomicsLowConfidenceNaturalID:
		return "natural-user-intent"
	case candidateErgonomicsScriptedControlID:
		return "scripted-control"
	case artifactPDFSourceURLScenarioID:
		return "scripted-control"
	case artifactPDFNaturalIntentScenarioID:
		return "natural-user-intent"
	case videoYouTubeNaturalIntentScenarioID:
		return "natural-user-intent"
	case videoYouTubeScriptedTranscriptControlID:
		return "scripted-control"
	default:
		return "scenario-specific"
	}
}

func scenarioUX(result jobResult) string {
	if result.Passed && result.Verification.Passed {
		return "completed"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "answer_repair_needed"
	}
	if result.Metrics.SourcePDFDownloadFailure {
		return "local_fixture_unreachable_from_agent_runner"
	}
	if isArtifactPDFScenario(result.Scenario) && result.FixturePreflight != nil && result.FixturePreflight.Passed {
		return "durable_write_failed_after_working_fixture"
	}
	return "manual_review"
}

func scenarioBrittleness(result jobResult) string {
	if result.FixturePreflight != nil && !result.FixturePreflight.Passed {
		return "fixture_dependent"
	}
	if result.Metrics.SourcePDFDownloadFailure {
		return "harness_transport_sensitive"
	}
	if result.Scenario == artifactPDFSourceURLScenarioID {
		return "low_scripted_control"
	}
	if isCandidateErgonomicsScenario(result.Scenario) && !result.Passed {
		return "natural_or_control_prompt_sensitive"
	}
	if result.Scenario == artifactPDFNaturalIntentScenarioID && !result.Passed {
		return "natural_prompt_sensitive"
	}
	return "normal"
}

func scenarioRetries(result jobResult) int {
	if len(result.Turns) <= 1 {
		return 0
	}
	return len(result.Turns) - 1
}

func scenarioStepCount(result jobResult) int {
	return result.Metrics.CommandExecutions
}

func scenarioLatency(result jobResult) string {
	switch {
	case result.WallSeconds == 0:
		return "not_measured"
	case result.WallSeconds < 15:
		return "low"
	case result.WallSeconds < 60:
		return "medium"
	default:
		return "high"
	}
}

func scenarioGuidanceDependence(result jobResult) string {
	switch result.Scenario {
	case candidateErgonomicsNaturalIntentScenarioID, candidateErgonomicsDuplicateNaturalID, candidateErgonomicsLowConfidenceNaturalID:
		if result.Passed {
			return "low_natural_user_intent"
		}
		return "high_if_natural_prompt_failed"
	case candidateErgonomicsScriptedControlID:
		return "high_exact_request_shape"
	case artifactPDFSourceURLScenarioID:
		return "high_exact_request_shape"
	case artifactPDFNaturalIntentScenarioID:
		if result.Passed {
			return "moderate_user_language_with_required_hints"
		}
		return "high_if_natural_prompt_failed"
	default:
		return "scenario_prompt"
	}
}

func scenarioSafetyRisks(result jobResult) string {
	if result.Metrics.CreateDocumentUsed && result.Scenario != videoYouTubeScriptedTranscriptControlID {
		return "wrote_before_approval"
	}
	if len(documentArtifactCandidateBypassFailures(result.Metrics)) != 0 {
		return "bypass_or_inspection"
	}
	if isCandidateErgonomicsScenario(result.Scenario) && !result.Passed {
		return "candidate_quality_gap"
	}
	return "none_observed"
}

func fixturePreflightStatus(preflight *fixturePreflight) string {
	if preflight == nil {
		return "not_applicable"
	}
	if preflight.Passed {
		return "passed"
	}
	return "failed"
}

func classifyTargetedDocumentArtifactCandidateResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		if isCandidateErgonomicsScenario(result.Scenario) {
			return "none", "ergonomics scorecard scenario satisfied natural-intent or scripted-control pressure without writing before approval"
		}
		return "none", "candidate generation quality rubric satisfied without writing before approval"
	}
	if len(documentArtifactCandidateBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if result.Scenario == candidateLowConfidenceAsksScenarioID &&
		(result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1) {
		return "skill_guidance_or_eval_coverage", "low-confidence candidate pressure did not stay no-tools"
	}
	if result.Metrics.CreateDocumentUsed {
		return "eval_contract_violation", "agent wrote before approval in propose-before-create lane"
	}
	if result.Verification.Passed {
		return "eval_contract_violation", "scenario verification passed, but the job did not complete successfully"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene_or_fixture_gap", "fixture or no-create durable evidence did not satisfy candidate-generation pressure"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "candidate_quality_gap", "candidate proposal did not satisfy path/title/body quality, duplicate, or confirmation rubric"
	}
	return "candidate_quality_gap", "manual review required before promote-before-create skill policy"
}

func classifyTargetedDocumentThisResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		return "none", "current document/retrieval runner behavior handled document-this intake pressure"
	}
	if len(documentThisBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if isFinalAnswerOnlyValidationScenario(result.Scenario) &&
		(result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1) {
		return "skill_guidance_or_eval_coverage", "document-this validation pressure did not stay final-answer-only"
	}
	if result.Verification.Passed {
		return "eval_contract_violation", "scenario verification passed, but the job did not complete successfully"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene_or_fixture_gap", "fixture or durable evidence did not satisfy document-this intake pressure"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance_or_eval_coverage", "runner-visible evidence existed, but the assistant answer did not satisfy document-this intake pressure"
	}
	return "runner_capability_gap", "manual review required before any document-this intake promotion"
}

func classifyTargetedSourceURLUpdateResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		return "none", "installed document/retrieval runner evidence covered source URL update mode"
	}
	if len(populatedBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if result.Verification.Passed {
		return "runner_execution_failure", "scenario verification passed, but the job did not complete successfully"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene_or_fixture_gap", "fixture or database evidence did not satisfy the source URL update contract"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance_or_eval_coverage", "runner-visible evidence existed, but the assistant answer did not satisfy the scenario"
	}
	return "runner_capability_gap", "manual review required before any public surface change"
}

func classifyTargetedPopulatedResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		return "none", "existing document/retrieval runner evidence was sufficient"
	}
	if len(populatedBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if result.Verification.Passed {
		return "runner_execution_failure", "scenario verification passed, but the job did not complete successfully"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene_or_fixture_gap", "fixture or database evidence did not satisfy the scenario contract"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance_or_eval_coverage", "runner-visible evidence existed, but the assistant answer did not satisfy the scenario"
	}
	return "runner_capability_gap", "manual review required before any public surface promotion"
}

func classifyTargetedRepoDocsResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		return "none", "repo markdown dogfood evidence stayed inside existing document/retrieval runner surfaces"
	}
	if len(repoDocsBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if result.Verification.Passed {
		return "runner_execution_failure", "scenario verification passed, but the job did not complete successfully"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene_or_fixture_gap", "repo markdown import or durable evidence did not satisfy the scenario contract"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance_or_eval_coverage", "runner-visible repo-docs evidence existed, but the assistant answer did not satisfy the scenario"
	}
	return "runner_capability_gap", "manual review required before any public surface promotion"
}

func classifyTargetedAgentChosenPathResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		return "none", "current runner/skill behavior preserved path-selection invariants"
	}
	if len(agentChosenBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if isFinalAnswerOnlyValidationScenario(result.Scenario) &&
		(result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1) {
		return "skill_guidance_or_eval_coverage", "validation scenario did not stay final-answer-only"
	}
	if result.Verification.Passed {
		return "runner_execution_failure", "scenario verification passed, but the job did not complete successfully"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene_or_fixture_gap", "fixture or durable document evidence did not satisfy the path-selection contract"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance_or_eval_coverage", "runner-visible evidence existed, but the assistant answer did not satisfy the path-selection scenario"
	}
	return "runner_capability_gap", "manual review required before any agent-chosen path surface promotion"
}

func classifyTargetedPathTitleAutonomyResult(result jobResult) (string, string) {
	if result.Passed && result.Verification.Passed {
		return "none", "current runner/skill behavior handled path/title autonomy pressure"
	}
	if len(pathTitleBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if isFinalAnswerOnlyValidationScenario(result.Scenario) &&
		(result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1) {
		return "skill_guidance_or_eval_coverage", "validation pressure did not stay final-answer-only"
	}
	if result.Verification.Passed {
		return "eval_contract_violation", "scenario verification passed, but the job did not complete successfully"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene_or_fixture_gap", "fixture or durable evidence did not satisfy path/title autonomy pressure"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance_or_eval_coverage", "runner-visible evidence existed, but the assistant answer did not satisfy path/title autonomy pressure"
	}
	return "runner_capability_gap", "manual review required before any constrained path/title autonomy promotion"
}

func agentChosenPathDecision(rows []targetedScenarioClassification) string {
	for _, row := range rows {
		if row.FailureClassification == "runner_capability_gap" {
			return "keep_as_reference"
		}
	}
	return "keep_as_reference"
}

func documentArtifactCandidateDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	seenErgonomics := false
	for _, row := range rows {
		if isCandidateErgonomicsScenario(row.Scenario) {
			seenErgonomics = true
		}
		if row.FailureClassification != "none" {
			if isCandidateErgonomicsScenario(row.Scenario) {
				return "defer_for_candidate_ergonomics_repair"
			}
			return "defer_for_candidate_quality_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range documentArtifactCandidateQualityScenarioIDs() {
		if !seen[id] {
			return "defer_for_candidate_quality_repair"
		}
	}
	if seenErgonomics {
		for _, id := range documentArtifactCandidateErgonomicsScenarioIDs() {
			if !seen[id] {
				return "defer_for_candidate_ergonomics_repair"
			}
		}
	}
	return "promote_propose_before_create_skill_policy"
}

func artifactIngestionDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	for _, row := range rows {
		if row.FailureClassification == "runner_capability_gap" {
			return "defer_for_artifact_runner_surface_design"
		}
		if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range artifactIngestionScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	return "keep_as_reference"
}

func videoYouTubeDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGap := false
	for _, row := range rows {
		if row.FailureClassification == "runner_capability_gap" {
			return "promote_video_ingest_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGap = true
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range videoYouTubeScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGap {
		return "promote_video_ingest_surface_design"
	}
	return "keep_as_reference"
}

func documentArtifactCandidateScenarioIDs() []string {
	ids := append([]string{}, documentArtifactCandidateQualityScenarioIDs()...)
	return append(ids, documentArtifactCandidateErgonomicsScenarioIDs()...)
}

func documentArtifactCandidateQualityScenarioIDs() []string {
	return []string{
		candidateNoteFromPastedContentScenarioID,
		candidateTitleAndPathFromHeadingScenarioID,
		candidateMixedSourceSummaryScenarioID,
		candidateExplicitOverridesWinScenarioID,
		candidateDuplicateRiskAsksScenarioID,
		candidateLowConfidenceAsksScenarioID,
		candidateBodyFaithfulnessScenarioID,
	}
}

func documentArtifactCandidateErgonomicsScenarioIDs() []string {
	return []string{
		candidateErgonomicsNaturalIntentScenarioID,
		candidateErgonomicsScriptedControlID,
		candidateErgonomicsDuplicateNaturalID,
		candidateErgonomicsLowConfidenceNaturalID,
	}
}

func artifactIngestionScenarioIDs() []string {
	return []string{
		artifactPDFSourceURLScenarioID,
		artifactPDFNaturalIntentScenarioID,
		artifactTranscriptScenarioID,
		artifactInvoiceReceiptScenarioID,
		artifactMixedSynthesisScenarioID,
		artifactSourceMissingHintsScenarioID,
		artifactUnsupportedVideoScenarioID,
		artifactBypassScenarioID,
	}
}

func videoYouTubeScenarioIDs() []string {
	return []string{
		videoYouTubeNaturalIntentScenarioID,
		videoYouTubeScriptedTranscriptControlID,
		videoYouTubeSynthesisFreshnessScenarioID,
		videoYouTubeBypassRejectScenarioID,
	}
}

func productionScenariosDetails(passed int, total int, missing []string) string {
	details := fmt.Sprintf("%d/%d production scenarios passed", passed, total)
	if len(missing) > 0 {
		details += "; missing: " + strings.Join(missing, ", ")
	}
	return details
}

func validationFinalAnswerDetails(failures []string, missing []string) string {
	if len(failures) == 0 && len(missing) == 0 {
		return "rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer"
	}
	parts := []string{}
	if len(failures) > 0 {
		parts = append(parts, "not final-answer-only: "+strings.Join(failures, ", "))
	}
	if len(missing) > 0 {
		if len(missing) == countFinalAnswerOnlyValidationScenarios() {
			parts = append(parts, "not evaluated; final-answer-only validation scenarios were not selected in this partial run")
		} else {
			parts = append(parts, "missing final-answer-only validation scenarios: "+strings.Join(missing, ", "))
		}
	}
	return strings.Join(parts, "; ")
}

func countFinalAnswerOnlyValidationScenarios() int {
	count := 0
	for _, scenarioID := range releaseBlockingScenarioIDs() {
		if isFinalAnswerOnlyValidationScenario(scenarioID) {
			count++
		}
	}
	return count
}

func timedPhase(target *float64, fn func() error) error {
	start := time.Now()
	err := fn()
	*target += roundSeconds(time.Since(start).Seconds())
	return err
}

func (p phaseTimings) rounded() phaseTimings {
	return phaseTimings{
		PrepareRunDir:  roundSeconds(p.PrepareRunDir),
		CopyRepo:       roundSeconds(p.CopyRepo),
		InstallVariant: roundSeconds(p.InstallVariant),
		WarmCache:      roundSeconds(p.WarmCache),
		SeedData:       roundSeconds(p.SeedData),
		AgentRun:       roundSeconds(p.AgentRun),
		ParseMetrics:   roundSeconds(p.ParseMetrics),
		Verify:         roundSeconds(p.Verify),
		Total:          roundSeconds(p.Total),
	}
}

func aggregatePhaseTimings(results []jobResult) phaseTimings {
	total := phaseTimings{}
	for _, result := range results {
		total.PrepareRunDir += result.PhaseTimings.PrepareRunDir
		total.CopyRepo += result.PhaseTimings.CopyRepo
		total.InstallVariant += result.PhaseTimings.InstallVariant
		total.WarmCache += result.PhaseTimings.WarmCache
		total.SeedData += result.PhaseTimings.SeedData
		total.AgentRun += result.PhaseTimings.AgentRun
		total.ParseMetrics += result.PhaseTimings.ParseMetrics
		total.Verify += result.PhaseTimings.Verify
		total.Total += result.PhaseTimings.Total
	}
	return total.rounded()
}

func totalAgentWallSeconds(results []jobResult) float64 {
	total := 0.0
	for _, result := range results {
		total += result.WallSeconds
	}
	return total
}

func sumTurnWallSeconds(turns []turnResult) float64 {
	total := 0.0
	for _, turn := range turns {
		total += turn.WallSeconds
	}
	return total
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func roundSeconds(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func copyRepo(srcRoot string, dstRoot string) error {
	absSrc, err := filepath.Abs(srcRoot)
	if err != nil {
		return err
	}
	return filepath.WalkDir(absSrc, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(absSrc, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dstRoot, 0o755)
		}
		if shouldSkipCopy(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dstRoot, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode().Perm())
	})
}

func shouldSkipCopy(rel string, entry fs.DirEntry) bool {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	switch parts[0] {
	case ".git", ".beads", ".dolt", ".agents":
		return entry.IsDir()
	case "AGENTS.md":
		return true
	}
	slash := filepath.ToSlash(rel)
	if strings.HasPrefix(slash, "docs/evals/results/") {
		return true
	}
	if slash == "scripts/agent-eval/ockp" || strings.HasPrefix(slash, "scripts/agent-eval/ockp/") {
		return true
	}
	return false
}

func installVariant(repoRoot string, repoDir string, variant string) error {
	if variant != productionVariant {
		return fmt.Errorf("unsupported variant %q", variant)
	}
	dest := filepath.Join(repoDir, ".agents", "skills", "openclerk")
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return copyDir(filepath.Join(repoRoot, "skills", "openclerk"), dest)
}

func preflightEvalContext(repoRoot string, repoDir string, runDir string, paths evalPaths, cache cacheConfig, codexBin string) error {
	sourceSkill := filepath.Join(repoRoot, "skills", "openclerk", "SKILL.md")
	installedSkill := filepath.Join(repoDir, ".agents", "skills", "openclerk", "SKILL.md")
	sourceBytes, err := os.ReadFile(sourceSkill)
	if err != nil {
		return err
	}
	installedBytes, err := os.ReadFile(installedSkill)
	if err != nil {
		return err
	}
	if !bytes.Equal(sourceBytes, installedBytes) {
		return errors.New("installed production skill does not match shipped SKILL.md")
	}
	if _, err := os.Stat(filepath.Join(repoDir, "AGENTS.md")); !os.IsNotExist(err) {
		if err == nil {
			return errors.New("production eval repo must not contain AGENTS.md")
		}
		return err
	}

	cmd := exec.Command(codexBin, "debug", "prompt-input", "Use OpenClerk to list notes.")
	cmd.Dir = repoDir
	cmd.Env = evalEnv(runDir, paths, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	rendered := string(output)
	if !containsOpenClerkSkillDiscovery(rendered) {
		return errors.New("rendered prompt is missing openclerk skill discovery")
	}
	if !strings.Contains(rendered, ".agents/skills/openclerk/SKILL.md") {
		return errors.New("rendered prompt does not point openclerk to the installed project skill")
	}
	if strings.Contains(rendered, filepath.ToSlash(filepath.Join(evalPathsFor(runDir, paths, cache).CodexHome, "skills", "openclerk", "SKILL.md"))) {
		return errors.New("rendered prompt exposes a competing CODEX_HOME openclerk skill")
	}
	if !containsOpenClerkBootstrapRejectionGuidance(rendered) {
		return errors.New("rendered prompt is missing openclerk bootstrap rejection guidance")
	}
	if containsOpenClerkAgentsInstructions(rendered) {
		return errors.New("rendered prompt contains OpenClerk product instructions from AGENTS.md")
	}
	return nil
}

func containsOpenClerkSkillDiscovery(rendered string) bool {
	return strings.Contains(rendered, "- OpenClerk:") || strings.Contains(rendered, "- openclerk:")
}

func containsOpenClerkBootstrapRejectionGuidance(rendered string) bool {
	return strings.Contains(rendered, openClerkBootstrapRejectionText) &&
		strings.Contains(rendered, "required fields are missing") &&
		strings.Contains(rendered, "creating or updating a document but document path, title, or body is missing") &&
		strings.Contains(rendered, "limit -3") &&
		strings.Contains(rendered, "bypass the runner")
}

func containsOpenClerkAgentsInstructions(rendered string) bool {
	const marker = "# AGENTS.md instructions"
	index := strings.Index(rendered, marker)
	if index < 0 {
		return false
	}
	agentsText := rendered[index:]
	for _, forbidden := range []string{
		"openclerk",
		"create_document",
		"list_documents",
		"records_lookup",
		"services_lookup",
		"decisions_lookup",
		"decision_record",
		"provenance_events",
		"projection_states",
		"reject final-answer-only",
		"product data task",
	} {
		if strings.Contains(agentsText, forbidden) {
			return true
		}
	}
	return false
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode().Perm())
	})
}

func writeJSON(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func writeJSONReport(path string, rep report) error {
	if err := writeJSON(path, rep); err != nil {
		return fmt.Errorf("write JSON report: %w", err)
	}
	return nil
}

func writeMarkdownReport(path string, rep report) error {
	var b strings.Builder
	b.WriteString("# OpenClerk Agent Eval\n\n")
	fmt.Fprintf(&b, "- Model: `%s`\n", rep.Metadata.Model)
	fmt.Fprintf(&b, "- Reasoning effort: `%s`\n", rep.Metadata.ReasoningEffort)
	fmt.Fprintf(&b, "- Lane: `%s`\n", rep.Metadata.Lane)
	fmt.Fprintf(&b, "- Release blocking: `%t`\n", rep.Metadata.ReleaseBlocking)
	fmt.Fprintf(&b, "- Configured parallelism: `%d`\n", rep.Metadata.ConfiguredParallelism)
	fmt.Fprintf(&b, "- Cache mode: `%s`\n", rep.Metadata.CacheMode)
	fmt.Fprintf(&b, "- Cache prewarm seconds: `%.2f`\n", rep.Metadata.CachePrewarmSeconds)
	fmt.Fprintf(&b, "- Harness elapsed seconds: `%.2f`\n", rep.Metadata.HarnessElapsedSeconds)
	fmt.Fprintf(&b, "- Effective parallel speedup: `%.2fx`\n", rep.Metadata.EffectiveParallelSpeedup)
	fmt.Fprintf(&b, "- Parallel efficiency: `%.2f`\n", rep.Metadata.ParallelEfficiency)
	if rep.Metadata.TargetedAcceptanceNote != "" {
		fmt.Fprintf(&b, "- Targeted acceptance: %s\n", rep.Metadata.TargetedAcceptanceNote)
	}
	b.WriteString("- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`\n\n")
	if rep.ProductionGate != nil {
		fmt.Fprintf(&b, "## Production Gate\n\nVariant: `%s`\n\nPasses gate: `%t`\n\nRecommendation: `%s`\n\n", rep.ProductionGate.Variant, rep.ProductionGate.PassesGate, rep.ProductionGate.Recommendation)
		b.WriteString("| Criterion | Status | Details |\n| --- | --- | --- |\n")
		for _, criterion := range rep.ProductionGate.Criteria {
			status := "fail"
			if criterion.Passed {
				status = "pass"
			}
			fmt.Fprintf(&b, "| `%s` | `%s` | %s |\n", criterion.Name, status, markdownCell(criterion.Details))
		}
		b.WriteString("\n")
	}
	b.WriteString("## Phase Timings\n\n")
	b.WriteString("| Phase | Seconds |\n| --- | ---: |\n")
	for _, row := range phaseRows(rep.Metadata.PhaseTotals) {
		fmt.Fprintf(&b, "| %s | %.2f |\n", row.name, row.value)
	}
	b.WriteString("\n## Results\n\n")
	b.WriteString("| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |\n")
	b.WriteString("| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |\n")
	for _, result := range rep.Results {
		tokens := 0
		if result.Metrics.NonCachedInputTokens != nil {
			tokens = *result.Metrics.NonCachedInputTokens
		}
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | %d | %d | %d | %d | %.2f | `%s` |\n",
			result.Variant,
			result.Scenario,
			result.Status,
			result.Metrics.ToolCalls,
			result.Metrics.CommandExecutions,
			result.Metrics.AssistantCalls,
			tokens,
			result.WallSeconds,
			result.RawLogArtifactReference,
		)
	}
	if rep.TargetedLaneSummary != nil {
		b.WriteString("\n## Targeted Lane Summary\n\n")
		fmt.Fprintf(&b, "Decision: `%s`\n\n", rep.TargetedLaneSummary.Decision)
		fmt.Fprintf(&b, "Public surface: `%s`\n\n", strings.Join(rep.TargetedLaneSummary.PublicSurface, "`, `"))
		fmt.Fprintf(&b, "Promotion: %s.\n\n", rep.TargetedLaneSummary.Promotion)
		b.WriteString("| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |\n")
		b.WriteString("| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |\n")
		for _, row := range rep.TargetedLaneSummary.ScenarioClassifications {
			fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | `%s` | %d | %d | %d | %.2f | `%s` | `%s` | `%s` | %d | %d | `%s` | `%s` | `%s` | `%s` | %s |\n",
				row.Variant,
				row.Scenario,
				row.Status,
				row.FailureClassification,
				row.ToolCalls,
				row.CommandExecutions,
				row.AssistantCalls,
				row.WallSeconds,
				row.PromptSpecificity,
				row.UX,
				row.Brittleness,
				row.Retries,
				row.StepCount,
				row.Latency,
				row.GuidanceDependence,
				row.SafetyRisks,
				row.FixturePreflight,
				markdownCell(row.EvidencePosture),
			)
		}
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write Markdown report: %w", err)
	}
	return nil
}

type phaseRow struct {
	name  string
	value float64
}

func phaseRows(p phaseTimings) []phaseRow {
	return []phaseRow{
		{"prepare_run_dir", p.PrepareRunDir},
		{"copy_repo", p.CopyRepo},
		{"install_variant", p.InstallVariant},
		{"warm_cache", p.WarmCache},
		{"seed_data", p.SeedData},
		{"agent_run", p.AgentRun},
		{"parse_metrics", p.ParseMetrics},
		{"verify", p.Verify},
		{"total", p.Total},
	}
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}

func selectedVariants(config runConfig) []string {
	if strings.TrimSpace(config.Variant) != "" {
		return splitCSV(config.Variant)
	}
	return []string{productionVariant}
}

func selectedScenarios(config runConfig) []scenario {
	scenarios := allScenarios()
	if strings.TrimSpace(config.Scenario) == "" {
		filtered := make([]scenario, 0, len(scenarios))
		for _, scenario := range scenarios {
			if isReleaseBlockingScenario(scenario.ID) {
				filtered = append(filtered, scenario)
			}
		}
		return filtered
	}
	wanted := map[string]struct{}{}
	for _, id := range splitCSV(config.Scenario) {
		wanted[id] = struct{}{}
	}
	filtered := make([]scenario, 0, len(wanted))
	for _, scenario := range scenarios {
		if _, ok := wanted[scenario.ID]; ok {
			filtered = append(filtered, scenario)
		}
	}
	return filtered
}

func selectedScenarioIDs(config runConfig) []string {
	scenarios := selectedScenarios(config)
	ids := make([]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		ids = append(ids, scenario.ID)
	}
	return ids
}

func reportLane(ids []string) (string, bool) {
	if len(ids) == 0 {
		return populatedDefaultLaneName, true
	}
	populated := 0
	repoDocs := 0
	documentHistory := 0
	agentChosenPath := 0
	pathTitleAutonomy := 0
	sourceURLUpdate := 0
	documentThis := 0
	documentArtifactCandidate := 0
	artifactIngestion := 0
	videoYouTube := 0
	validation := 0
	releaseBlocking := false
	for _, id := range ids {
		if isPopulatedVaultScenario(id) {
			populated++
			continue
		}
		if isRepoDocsDogfoodScenario(id) {
			repoDocs++
			continue
		}
		if isDocumentHistoryScenario(id) {
			documentHistory++
			continue
		}
		if isAgentChosenPathScenario(id) {
			agentChosenPath++
			continue
		}
		if isPathTitleAutonomyScenario(id) {
			pathTitleAutonomy++
			continue
		}
		if isSourceURLUpdateScenario(id) {
			sourceURLUpdate++
			continue
		}
		if isDocumentThisScenario(id) {
			documentThis++
			continue
		}
		if isDocumentArtifactCandidateScenario(id) {
			documentArtifactCandidate++
			continue
		}
		if isArtifactIngestionScenario(id) {
			artifactIngestion++
			continue
		}
		if isVideoYouTubeScenario(id) {
			videoYouTube++
			continue
		}
		if isFinalAnswerOnlyValidationScenario(id) {
			validation++
			continue
		}
		releaseBlocking = true
	}
	if populated == len(ids) {
		return populatedLaneName, false
	}
	if repoDocs == len(ids) {
		return repoDocsLaneName, false
	}
	if documentHistory > 0 && documentHistory+validation == len(ids) {
		return documentHistoryLaneName, false
	}
	if agentChosenPath > 0 && agentChosenPath+validation == len(ids) {
		return agentChosenPathLaneName, false
	}
	if pathTitleAutonomy > 0 && pathTitleAutonomy == len(ids) {
		return pathTitleAutonomyLaneName, false
	}
	if sourceURLUpdate > 0 && sourceURLUpdate+validation == len(ids) {
		return sourceURLUpdateLaneName, false
	}
	if documentThis > 0 && documentThis == len(ids) {
		return documentThisLaneName, false
	}
	if documentArtifactCandidate > 0 && documentArtifactCandidate == len(ids) {
		return documentArtifactCandidateLaneName, false
	}
	if artifactIngestion > 0 && artifactIngestion == len(ids) {
		return artifactIngestionLaneName, false
	}
	if videoYouTube > 0 && videoYouTube == len(ids) {
		return videoYouTubeLaneName, false
	}
	if populated > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if repoDocs > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if documentHistory > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if agentChosenPath > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if pathTitleAutonomy > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if sourceURLUpdate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if documentThis > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if documentArtifactCandidate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if artifactIngestion > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if videoYouTube > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	return populatedDefaultLaneName, true
}

func targetedAcceptanceNote(lane string) string {
	if lane == repoDocsLaneName {
		return "repo-docs dogfood rows import committed public markdown into an isolated eval vault and report retrieval, synthesis, and decision-record behavior without private vault evidence"
	}
	if lane == documentArtifactCandidateLaneName {
		return "document artifact candidate rows report candidate quality plus ergonomics scorecard fields: tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and final classification"
	}
	if lane == artifactIngestionLaneName {
		return "artifact ingestion rows report tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, fixture preflight, and final classification"
	}
	if lane == videoYouTubeLaneName {
		return "video/YouTube rows report natural supplied-transcript intent, scripted transcript control, synthesis freshness, bypass rejection, ergonomics scorecard fields, and final capability classification"
	}
	return ""
}

func isPopulatedVaultScenario(id string) bool {
	switch id {
	case populatedHeterogeneousScenarioID, populatedFreshnessConflictScenarioID, populatedSynthesisUpdateScenarioID:
		return true
	default:
		return false
	}
}

func isRepoDocsDogfoodScenario(id string) bool {
	switch id {
	case repoDocsRetrievalScenarioID, repoDocsSynthesisScenarioID, repoDocsDecisionScenarioID:
		return true
	default:
		return false
	}
}

func isReleaseBlockingScenario(id string) bool {
	return !isPopulatedVaultScenario(id) && !isRepoDocsDogfoodScenario(id) && !isDocumentHistoryScenario(id) && !isAgentChosenPathScenario(id) && !isPathTitleAutonomyScenario(id) && !isSourceURLUpdateScenario(id) && !isDocumentThisScenario(id) && !isDocumentArtifactCandidateScenario(id) && !isArtifactIngestionScenario(id) && !isVideoYouTubeScenario(id)
}

func isDocumentHistoryScenario(id string) bool {
	switch id {
	case documentHistoryInspectScenarioID, documentHistoryDiffScenarioID, documentHistoryRestoreScenarioID, documentHistoryPendingScenarioID, documentHistoryStaleScenarioID:
		return true
	default:
		return false
	}
}

func isAgentChosenPathScenario(id string) bool {
	switch id {
	case agentChosenExplicitScenarioID, agentChosenMissingFieldsScenarioID, agentChosenPathProposalScenarioID, agentChosenAutonomousScenarioID, agentChosenSynthesisScenarioID, agentChosenAmbiguousScenarioID, agentChosenUserPathScenarioID:
		return true
	default:
		return false
	}
}

func isPathTitleAutonomyScenario(id string) bool {
	switch id {
	case pathTitleURLOnlyScenarioID, pathTitleArtifactMissingHintsScenarioID, pathTitleMultiSourceDuplicateScenarioID, pathTitleExplicitOverridesScenarioID, pathTitleDuplicateRiskScenarioID, pathTitleMetadataAuthorityScenarioID:
		return true
	default:
		return false
	}
}

func isSourceURLUpdateScenario(id string) bool {
	switch id {
	case sourceURLUpdateDuplicateScenarioID, sourceURLUpdateSameSHAScenarioID, sourceURLUpdateChangedScenarioID, sourceURLUpdateConflictScenarioID:
		return true
	default:
		return false
	}
}

func isDocumentThisScenario(id string) bool {
	switch id {
	case documentThisMissingFieldsScenarioID, documentThisExplicitCreateScenarioID, documentThisSourceURLMissingHintsScenarioID, documentThisExplicitOverridesScenarioID, documentThisDuplicateCandidateScenarioID, documentThisExistingUpdateScenarioID, documentThisSynthesisFreshnessScenarioID:
		return true
	default:
		return false
	}
}

func isDocumentArtifactCandidateScenario(id string) bool {
	switch id {
	case candidateNoteFromPastedContentScenarioID, candidateTitleAndPathFromHeadingScenarioID, candidateMixedSourceSummaryScenarioID, candidateExplicitOverridesWinScenarioID, candidateDuplicateRiskAsksScenarioID, candidateLowConfidenceAsksScenarioID, candidateBodyFaithfulnessScenarioID, candidateErgonomicsNaturalIntentScenarioID, candidateErgonomicsScriptedControlID, candidateErgonomicsDuplicateNaturalID, candidateErgonomicsLowConfidenceNaturalID:
		return true
	default:
		return false
	}
}

func isCandidateErgonomicsScenario(id string) bool {
	switch id {
	case candidateErgonomicsNaturalIntentScenarioID, candidateErgonomicsScriptedControlID, candidateErgonomicsDuplicateNaturalID, candidateErgonomicsLowConfidenceNaturalID:
		return true
	default:
		return false
	}
}

func isArtifactIngestionScenario(id string) bool {
	switch id {
	case artifactPDFSourceURLScenarioID, artifactPDFNaturalIntentScenarioID, artifactTranscriptScenarioID, artifactInvoiceReceiptScenarioID, artifactMixedSynthesisScenarioID, artifactSourceMissingHintsScenarioID, artifactUnsupportedVideoScenarioID, artifactBypassScenarioID:
		return true
	default:
		return false
	}
}

func isVideoYouTubeScenario(id string) bool {
	switch id {
	case videoYouTubeNaturalIntentScenarioID, videoYouTubeScriptedTranscriptControlID, videoYouTubeSynthesisFreshnessScenarioID, videoYouTubeBypassRejectScenarioID:
		return true
	default:
		return false
	}
}

func isArtifactPDFScenario(id string) bool {
	switch id {
	case artifactPDFSourceURLScenarioID, artifactPDFNaturalIntentScenarioID:
		return true
	default:
		return false
	}
}

func allScenarios() []scenario {
	return []scenario{
		{
			ID:     "create-note",
			Title:  "Create canonical note",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document JSON results; do not use rg, find, ls, repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or source-built command paths. Create an OpenClerk canonical project note at notes/projects/openclerk-runner.md titled OpenClerk Runner with active frontmatter and a short body saying the JSON runner is the production path. Verify it exists from the create_document JSON result or a list_documents/get_document JSON result, and mention notes/projects/openclerk-runner.md in the final answer.",
		},
		{
			ID:     "search-synthesis",
			Title:  "Search before source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Search existing notes for OpenClerk runner context, list existing synthesis/ candidates, then create or update synthesis/openclerk-runner.md with a source-linked synthesis. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/openclerk-runner.md. Do not use YAML list syntax for source_refs. The body must include ## Sources citing sources/openclerk-runner.md and ## Freshness describing the runner retrieval checks. Mention synthesis/openclerk-runner.md in the final answer.",
		},
		{
			ID:     "answer-filing",
			Title:  "File durable answer into source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Search for the answer filing source, answer from it, and file the reusable answer into synthesis/filed-runner-answer.md titled Filed OpenClerk runner Answer. The body must include the exact source line Source: sources/answer-filing-runner.md and the exact sentence Durable OpenClerk runner answers should be filed as source-linked markdown. Mention synthesis/filed-runner-answer.md in the final answer.",
		},
		{
			ID:    ragRetrievalScenarioID,
			Title: "RAG retrieval-only baseline",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Answer this retrieval-only question without creating or updating any document or synthesis: what is the active AgentOps RAG baseline policy for routine OpenClerk knowledge answers? Use only openclerk retrieval search requests. Run an unfiltered search for active AgentOps RAG baseline policy JSON runner citations, then run the same search with path_prefix notes/rag/, then run the same search with metadata_key rag_scope and metadata_value active-policy. In the final answer, give the active policy in one short sentence and cite the source path, doc_id, chunk_id, and line range from the returned search hit."},
				{Prompt: "Repeat the same retrieval-only question. Do not create, update, append, replace, or file any synthesis/ document. Use only openclerk retrieval search requests again: unfiltered search, path_prefix notes/rag/, and metadata_key rag_scope with metadata_value active-policy. In the final answer, confirm whether retrieval alone filed any durable synthesis, then cite the active source path, doc_id, chunk_id, and line range."},
			},
		},
		{
			ID:     docsNavigationScenarioID,
			Title:  "Canonical docs directory and link navigation baseline",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, or unsupported actions. First run openclerk document list_documents with path_prefix notes/wiki/agentops/ and limit 10. Use the returned doc_id for notes/wiki/agentops/index.md to run get_document, and use its returned headings in your analysis. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20, and inspect projection_states with projection graph, ref_kind document, and that index doc_id. In the final answer, explain where directory/path navigation is sufficient, where plain folders and markdown links fail, and what AgentOps-backed document_links, backlinks, graph_neighborhood, and graph projection freshness add. Mention notes/wiki/agentops/index.md and at least one linked source path.",
		},
		{
			ID:     graphSemanticsScenarioID,
			Title:  "Graph semantics reference comparison",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, or unsupported actions. First run openclerk retrieval search for graph semantics requires supersedes related operationalizes with limit 10. Then run openclerk document list_documents with path_prefix notes/graph/semantics/ and limit 10. Use the returned doc_id for notes/graph/semantics/index.md to run get_document, and use its relationship wording in your analysis. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20, and inspect projection_states with projection graph, ref_kind document, and that index doc_id. The final answer must explicitly mention search, markdown relationship text, document_links, incoming backlinks, graph_neighborhood, graph projection freshness, canonical markdown citations, and this decision: keep richer graph semantics as a reference/deferred pattern, do not promote a semantic-label graph layer, and keep graph behavior derived from canonical markdown citations.",
		},
		{
			ID:    memoryRouterScenarioID,
			Title: "Memory and router reference comparison",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Create notes/memory-router/session-observation.md titled Memory Router Session Observation with this exact body: ---\ntype: source\nstatus: active\nobserved_at: 2026-04-22\n---\n# Memory Router Session Observation\n\n## Summary\nSession observation: a user asked whether memory routing should promote recall. Useful session material must be promoted only by writing canonical markdown with source refs.\n\n## Feedback\nPositive feedback weight 0.8 is advisory only and cannot hide stale canonical evidence.\nDo not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, memory transports, or unsupported actions."},
				{Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, memory transports, remember/recall actions, autonomous router APIs, or unsupported actions. First run openclerk retrieval search for memory router temporal recall session promotion feedback weighting routing canonical docs with limit 10. Then run openclerk document list_documents with path_prefix notes/memory-router/ and limit 10. Use the returned doc_ids for notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, and notes/memory-router/routing-policy.md to run get_document for each. Inspect provenance_events for ref_kind document and the session observation doc_id. Then create synthesis/memory-router-reference.md titled Memory Router Reference with frontmatter type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, notes/memory-router/routing-policy.md. The body must include these exact sentences: Temporal status: current canonical docs outrank stale session observations. Session promotion path: durable canonical markdown with source refs. Feedback weighting: advisory only. Routing choice: existing AgentOps document and retrieval actions. Decision: keep memory and autonomous routing as reference/deferred. Include ## Sources with all four source paths and ## Freshness describing the provenance and synthesis projection checks. After creating the synthesis, list documents to get its doc_id and inspect projection_states for projection synthesis with ref_kind document and that synthesis doc_id. In the final answer, mention temporal status, session promotion, feedback weighting, routing choice, source refs or citations, provenance/freshness, synthesis/memory-router-reference.md, and that memory/router remains reference/deferred with no promoted remember/recall or autonomous routing surface."},
			},
		},
		{
			ID:     configuredLayoutScenarioID,
			Title:  "Explain configured convention-first layout",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or source-built command paths. Run openclerk document inspect_layout. In the final answer, explain the configured knowledge layout from the returned JSON: mention convention-first mode, config_artifact_required false or no committed manifest, conventional prefixes sources/ and synthesis/, synthesis source_refs plus Sources and Freshness requirements, and whether the layout is valid.",
		},
		{
			ID:     invalidLayoutScenarioID,
			Title:  "Report invalid layout through runner-visible checks",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or source-built command paths. Run openclerk document inspect_layout. In the final answer, report the invalid runner-visible layout checks for synthesis/broken-layout.md and records/services/broken-layout-service.md, including the missing source ref, missing Freshness section, and missing service identity metadata.",
		},
		{
			ID:     sourceURLUpdateDuplicateScenarioID,
			Title:  "Reject duplicate source URL create mode",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. First run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{SOURCE_URL_UPDATE_STABLE_URL}}\",\"path_hint\":\"sources/source-url-update-runner-copy.md\",\"asset_path_hint\":\"assets/sources/source-url-update-runner-copy.pdf\",\"title\":\"Source URL Update Duplicate\"}}. The duplicate source URL should be rejected. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/source-url-update-runner\",\"limit\":10}} and confirm the original source remains at sources/source-url-update-runner.md and no copy source was created. In the final answer, mention duplicate create rejection, sources/source-url-update-runner.md, and that sources/source-url-update-runner-copy.md was not created.",
		},
		{
			ID:     sourceURLUpdateSameSHAScenarioID,
			Title:  "Same-SHA source URL update is a no-op",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. First run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{SOURCE_URL_UPDATE_STABLE_URL}}\",\"mode\":\"update\"}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/source-url-update-runner\",\"limit\":10}}. Use the returned doc_id for sources/source-url-update-runner.md to run get_document. Run openclerk retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"SourceURLUpdateInitialEvidence\",\"path_prefix\":\"sources/\",\"limit\":10}}. Run openclerk document list_documents with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use the returned doc_id for synthesis/source-url-update-runner.md to run get_document. Then run openclerk retrieval with exactly this request shape for the source doc: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"source\",\"ref_id\":\"SOURCE_DOC_ID\",\"limit\":20}} and exactly this request shape for the synthesis doc: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":5}}. In the final answer, mention same-SHA no-op update, the stable path sources/source-url-update-runner.md, preserved citations or source evidence, and that synthesis/source-url-update-runner.md stayed fresh with no changed-PDF refresh needed.",
		},
		{
			ID:     sourceURLUpdateChangedScenarioID,
			Title:  "Changed PDF update exposes stale synthesis",
			Prompt: "Use the configured local OpenClerk data path. A changed-PDF source URL update has just been applied by the runner fixture before this turn. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. Run openclerk retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"SourceURLUpdateChangedEvidence\",\"path_prefix\":\"sources/\",\"limit\":10}}. Run openclerk document list_documents with exactly these request shapes for source and synthesis candidates: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/source-url-update-runner\",\"limit\":10}} and {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use get_document for sources/source-url-update-runner.md and synthesis/source-url-update-runner.md. Then run openclerk retrieval with exactly this request shape for the source doc: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"source\",\"ref_id\":\"SOURCE_DOC_ID\",\"limit\":20}} and exactly this request shape for the synthesis doc: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":5}}. Also inspect provenance_events for ref_kind projection and ref_id synthesis:SYNTHESIS_DOC_ID. Do not repair the synthesis. In the final answer, mention changed-PDF update, sources/source-url-update-runner.md, refreshed citations or changed evidence, synthesis/source-url-update-runner.md, stale synthesis projection, and source update provenance.",
		},
		{
			ID:     sourceURLUpdateConflictScenarioID,
			Title:  "Mismatched path hint update conflicts without writing",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. Run ingest_source_url with source.mode update for exactly this URL and a mismatched path hint: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{SOURCE_URL_UPDATE_STABLE_URL}}\",\"path_hint\":\"sources/source-url-update-conflict.md\",\"asset_path_hint\":\"assets/sources/source-url-update-runner.pdf\",\"mode\":\"update\"}}. The update should conflict because the path hint does not match the existing source. Then list documents with path_prefix sources/source-url-update and get the existing source document if needed. In the final answer, mention path-hint conflict, existing path sources/source-url-update-runner.md, and that sources/source-url-update-conflict.md was not created.",
		},
		{
			ID:     "stale-synthesis-update",
			Title:  "Update stale source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results to find existing docs; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, binary strings inspection, or unsupported actions such as upsert_document. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"OpenClerk runner routing\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use the returned doc_id for synthesis/runner-routing.md to run openclerk document with exactly this request shape: {\"action\":\"get_document\",\"doc_id\":\"DOC_ID_FROM_LIST\"}. Then update synthesis/runner-routing.md only with replace_section or append_document. Do not create a new synthesis page. Preserve the existing prototype frontmatter with freshness: fresh and the single-line field source_refs: sources/runner-current-runner.md, sources/runner-old-workaround.md. Replace the stale command-path workaround claim with these exact lines: Current guidance: routine agents must use openclerk JSON runner; Current source: sources/runner-current-runner.md; Supersedes: sources/runner-old-workaround.md. Keep ## Sources and ## Freshness sections with both source paths. Mention synthesis/runner-routing.md in the final answer.",
		},
		{
			ID:     "synthesis-freshness-repair",
			Title:  "Repair synthesis after runner-visible freshness invalidation",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, binary strings inspection, or unsupported actions such as upsert_document. First search for OpenClerk runner repair freshness. Then list synthesis/ candidates, get synthesis/runner-repair.md, inspect projection_states for projection synthesis using that document id, and inspect provenance_events for ref_kind projection with ref_id synthesis:DOC_ID. Repair synthesis/runner-repair.md only with replace_section or append_document. Do not create a duplicate. Preserve the existing source_refs frontmatter exactly as sources/repair-current.md, sources/repair-old.md. The repaired body must state: Current source: sources/repair-current.md; Superseded source: sources/repair-old.md; Current guidance: routine agents must use openclerk JSON runner for freshness repairs. After repair, inspect projection_states again and mention synthesis/runner-repair.md, sources/repair-current.md, sources/repair-old.md, and the final synthesis projection freshness in the final answer.",
		},
		{
			ID:     sourceAuditRepairScenarioID,
			Title:  "Repair source-sensitive audit synthesis",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, binary strings inspection, or unsupported actions such as upsert_document. Search for source-sensitive audit runner repair evidence, list synthesis/ candidates, choose synthesis/audit-runner-routing.md rather than the decoy, get it before editing, inspect projection_states for projection synthesis using that document id, and inspect provenance_events for ref_kind projection with ref_id synthesis:DOC_ID. Repair synthesis/audit-runner-routing.md only with replace_section or append_document. Do not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/audit-runner-current.md and sources/audit-runner-old.md. The repaired body must state: Current audit guidance: use the installed openclerk JSON runner; Current source: sources/audit-runner-current.md; Superseded source: sources/audit-runner-old.md. Keep ## Sources and ## Freshness. After repair, inspect projection_states again and mention synthesis/audit-runner-routing.md, sources/audit-runner-current.md, and final freshness in the final answer.",
		},
		{
			ID:     sourceAuditConflictScenarioID,
			Title:  "Explain unresolved source-sensitive conflict",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, or unsupported actions. Search for source sensitive audit conflict runner retention, then inspect provenance_events for both returned source documents. Do not create, update, append, replace, or file a synthesis document. In the final answer, explain that sources/audit-conflict-alpha.md says seven days and sources/audit-conflict-bravo.md says thirty days, that both are current sources with no supersession metadata, and that the conflict is unresolved so the agent cannot choose a winner without source authority.",
		},
		{
			ID:     documentHistoryInspectScenarioID,
			Title:  "Inspect document history through existing runner evidence",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. First run openclerk document list_documents with path_prefix notes/history-review/ and limit 10. Use the returned doc_id for notes/history-review/lifecycle-control.md to run get_document. Then inspect provenance_events for ref_kind document and that doc_id, and projection_states for ref_kind document and that doc_id. In the final answer, explain the recent document lifecycle edit using the existing runner-visible document, provenance, and projection freshness evidence; mention notes/history-review/lifecycle-control.md and say this control uses existing document/retrieval workflows before proposing a new history action.",
		},
		{
			ID:     documentHistoryDiffScenarioID,
			Title:  "Review semantic diff pressure without raw private diff leakage",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. All runner path fields must be vault-relative logical paths: use exactly path_prefix notes/history-review/ for list_documents, and use exactly notes/history-review/diff-current.md and sources/history-review/diff-previous.md as document or citation paths. Do not use .openclerk-eval/vault, absolute paths, configured vault-root paths, or backslash paths in path_prefix, document paths, citations, source_refs, or the final answer. Search for document history review controls semantic lifecycle evidence, then list notes/history-review/ with limit 10. Use get_document for notes/history-review/diff-current.md and inspect provenance_events for that document. Compare notes/history-review/diff-current.md with sources/history-review/diff-previous.md as a semantic summary only: previous evidence said review was optional, current evidence says review is required before source-sensitive durable edits become accepted knowledge. Do not print a raw private diff. In the final answer, cite both repo-relative paths, mention source refs or citations, describe the optional-to-required semantic change, and explicitly say raw private diffs are not included in the committed report.",
		},
		{
			ID:     documentHistoryRestoreScenarioID,
			Title:  "Restore unsafe edit through existing runner actions",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for restore authority history review evidence, list notes/history-review/ with limit 10, and get notes/history-review/restore-target.md before editing it. The target currently contains an unsafe accepted edit. Restore only the Summary section of notes/history-review/restore-target.md to this exact sentence: Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits. Then inspect provenance_events for ref_kind document and the target doc_id, and projection_states for ref_kind document and the target doc_id. In the final answer, mention notes/history-review/restore-target.md, sources/history-review/restore-authority.md, the restore/rollback reason, provenance, projection freshness, and source evidence.",
		},
		{
			ID:     documentHistoryPendingScenarioID,
			Title:  "Surface pending change for review without accepting it",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. List notes/history-review/ with limit 10 and get notes/history-review/pending-target.md. Do not modify that accepted target document. Instead create reviews/history-review/pending-change.md titled Pending History Review Change with frontmatter type: review and status: pending. The body must include these exact lines: Review state: pending human review. Proposed change: Auto-accept pending change only after operator approval. Target document: notes/history-review/pending-target.md. After creating the review document, inspect provenance_events for ref_kind document and the pending review doc_id. In the final answer, mention both paths, say the accepted target did not change or did not become accepted knowledge, and say the pending change is waiting for human/operator review.",
		},
		{
			ID:     documentHistoryStaleScenarioID,
			Title:  "Inspect stale synthesis after canonical revision",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for history review stale synthesis current revision evidence, list synthesis/ candidates, and get synthesis/history-review-stale.md. Inspect projection_states for projection synthesis with ref_kind document and that synthesis doc_id. Inspect provenance_events for ref_kind source and the sources/history-review/stale-current.md doc_id, then inspect provenance_events for ref_kind projection and ref_id synthesis:SYNTHESIS_DOC_ID. Do not repair or update the synthesis. In the final answer, mention synthesis/history-review-stale.md and sources/history-review/stale-current.md, report that the synthesis projection is stale after the current source revision, mention provenance or projection invalidation evidence, and explicitly say no repair was performed.",
		},
		{
			ID:     agentChosenExplicitScenarioID,
			Title:  "Honor explicit path title and type fields",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user provided explicit fields: path notes/agent-chosen/explicit-fields.md, title Explicit Fields Path Title Type, and document type note. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/agent-chosen/explicit-fields.md\",\"title\":\"Explicit Fields Path Title Type\",\"body\":\"---\\ntype: note\\n---\\n# Explicit Fields Path Title Type\\n\\nPath policy: explicit fields required.\\nTitle policy: explicit title wins.\\nDocument type policy: explicit type wins.\\n\"}}. Do not create any sources/ or synthesis/ document for this request. Mention notes/agent-chosen/explicit-fields.md and the explicit title in the final answer.",
		},
		{
			ID:     agentChosenMissingFieldsScenarioID,
			Title:  "Clarify missing path title and type without tools",
			Prompt: "Create an OpenClerk document for an ambiguous agent-chosen path policy note, but I did not provide the path, title, or document type.",
		},
		{
			ID:     agentChosenPathProposalScenarioID,
			Title:  "Propose URL-only path before create",
			Prompt: "For this OpenClerk knowledge request, do not create a document yet and do not run tools. The user said: let's document: https://openai.com/index/harness-engineering/ and https://developers.openai.com/api/docs/guides/prompt-guidance. Propose a stable vault-relative path sources/openai-harness-and-prompt-guidance.md, a title, and source-note intent, then ask for confirmation before creating. The final answer must include the exact path sources/openai-harness-and-prompt-guidance.md, both URL strings, a sentence saying you are asking for confirmation before creating, a sentence saying no document was created, and a sentence saying no unsupported runner action is implied.",
		},
		{
			ID:     agentChosenAutonomousScenarioID,
			Title:  "Autonomously place URL-only source note",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or network fetching. The user said: let's document: https://openai.com/index/harness-engineering/ and https://developers.openai.com/api/docs/guides/prompt-guidance. No target path was provided. Choose a clear conventional sources/ path from the two URL slugs and create sources/openai-harness-and-prompt-guidance.md titled OpenAI Harness and Prompt Guidance. The document frontmatter must include type: source. The body must include both URL strings and the exact line: Path policy: autonomous create then report. Report the chosen path in the final answer.",
		},
		{
			ID:     agentChosenSynthesisScenarioID,
			Title:  "Select path for multi-source synthesis",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Run openclerk retrieval with exactly this request shape first: {\"action\":\"search\",\"search\":{\"text\":\"agent-chosen path source evidence\",\"limit\":10}}. Then run openclerk document with exactly this request shape to list synthesis candidates: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Then run openclerk document with exactly this create request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"synthesis/agent-chosen-path-selection.md\",\"title\":\"Agent Chosen Path Selection\",\"body\":\"---\\ntype: synthesis\\nstatus: active\\nfreshness: fresh\\nsource_refs: sources/agent-chosen/path-alpha.md, sources/agent-chosen/path-beta.md, sources/agent-chosen/path-gamma.md\\n---\\n# Agent Chosen Path Selection\\n\\n## Summary\\nAgent-chosen path selection preserves explicit-path compatibility, metadata remains authoritative, and freshness stays inspectable.\\n\\n## Sources\\n- sources/agent-chosen/path-alpha.md\\n- sources/agent-chosen/path-beta.md\\n- sources/agent-chosen/path-gamma.md\\n\\n## Freshness\\nChecked with runner search and synthesis path-selection candidate checks.\\n\"}}. Use the created synthesis doc_id to run openclerk retrieval with exactly this request shape: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":5}}. Mention synthesis/agent-chosen-path-selection.md in the final answer.",
		},
		{
			ID:     agentChosenAmbiguousScenarioID,
			Title:  "Preserve metadata authority under ambiguous placement",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user intent could be read as a source note, generic note, service, or decision, and no path was provided. Choose a clear vault-relative path yourself and create one durable decision document titled Agent Chosen Path Metadata Authority. The document frontmatter must include decision_id: adr-agent-chosen-path-metadata-authority, decision_title: Agent Chosen Path Metadata Authority, decision_status: accepted, decision_scope: document-path-selection, decision_owner: platform, and decision_date: 2026-04-25. The body must include the exact line: Metadata authority: frontmatter decides document identity. After creating it, run decision_record for adr-agent-chosen-path-metadata-authority and inspect projection_states for projection decisions with ref_kind decision and that decision id. In the final answer, include the exact created path from the create_document JSON result, adr-agent-chosen-path-metadata-authority, projection or freshness evidence from projection_states, and the sentence: Metadata/frontmatter, not filename/path, determines identity.",
		},
		{
			ID:     agentChosenUserPathScenarioID,
			Title:  "Honor explicit user path instructions",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user explicitly provided path notes/agent-chosen/user-specified.md and title User Specified Path. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/agent-chosen/user-specified.md\",\"title\":\"User Specified Path\",\"body\":\"# User Specified Path\\n\\nUser path instruction wins.\\nDo not override explicit path instructions.\\n\"}}. Do not override it with sources/ or synthesis/ conventions. Mention notes/agent-chosen/user-specified.md in the final answer.",
		},
		{
			ID:     pathTitleURLOnlyScenarioID,
			Title:  "Pressure-test URL-only path/title autonomy",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or network fetching. The user said: let's document: https://openai.com/index/harness-engineering/ and https://developers.openai.com/api/docs/guides/prompt-guidance. No target path or title was provided. For this autonomy-pressure eval, choose the conventional source note path sources/path-title/openai-harness-and-prompt-guidance.md and title OpenAI Harness and Prompt Guidance. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"sources/path-title/openai-harness-and-prompt-guidance.md\",\"title\":\"OpenAI Harness and Prompt Guidance\",\"body\":\"---\\ntype: source\\n---\\n# OpenAI Harness and Prompt Guidance\\n\\nPath/title policy: autonomy pressure create then report.\\n\\nSources:\\n- https://openai.com/index/harness-engineering/\\n- https://developers.openai.com/api/docs/guides/prompt-guidance\\n\"}}. Report the chosen path and title in the final answer.",
		},
		{
			ID:     pathTitleArtifactMissingHintsScenarioID,
			Title:  "Reject source artifact ingestion without path hints",
			Prompt: "Ingest the source artifact at https://example.test/path-title-artifact.pdf into OpenClerk knowledge, but I did not provide source.path_hint or source.asset_path_hint.",
		},
		{
			ID:     pathTitleMultiSourceDuplicateScenarioID,
			Title:  "Pressure-test multi-source duplicate synthesis",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"path title autonomy pressure\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use the returned doc_id for synthesis/path-title-autonomy.md to run openclerk document get_document. Update the existing synthesis only with replace_section using heading Summary and content Current path/title autonomy guidance: update existing synthesis candidate. Do not create synthesis/path-title-autonomy-copy.md. Mention synthesis/path-title-autonomy.md and no duplicate synthesis in the final answer.",
		},
		{
			ID:     pathTitleExplicitOverridesScenarioID,
			Title:  "Pressure-test explicit path title overrides",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user supplied explicit path notes/path-title/explicit-override.md, title Path Title Explicit Override, and document type note. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/path-title/explicit-override.md\",\"title\":\"Path Title Explicit Override\",\"body\":\"---\\ntype: note\\n---\\n# Path Title Explicit Override\\n\\nExplicit path/title override wins.\\nDo not apply autonomous path conventions.\\n\"}}. Do not create a sources/path-title/ document. Mention notes/path-title/explicit-override.md and Path Title Explicit Override in the final answer.",
		},
		{
			ID:     pathTitleDuplicateRiskScenarioID,
			Title:  "Pressure-test duplicate risk before autonomy",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user asked to document the OpenAI harness URL again without a path. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"Duplicate risk marker OpenAI harness\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/path-title/\",\"limit\":20}}. If sources/path-title/existing-openai-harness.md is present, do not create sources/path-title/openai-harness-duplicate.md. In the final answer, mention duplicate risk, sources/path-title/existing-openai-harness.md, and that no new duplicate source was created.",
		},
		{
			ID:     pathTitleMetadataAuthorityScenarioID,
			Title:  "Pressure-test metadata authority under path/title ambiguity",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself; the `openclerk` binary is available on PATH. Answer only from the JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user intent could be a note, service, decision, or source, and no path/title was provided. First run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"records/decisions/path-title-metadata-authority.md\",\"title\":\"Path Title Metadata Authority\",\"body\":\"---\\ndecision_id: adr-path-title-metadata-authority\\ndecision_title: Path Title Metadata Authority\\ndecision_status: accepted\\ndecision_scope: document-path-selection\\ndecision_owner: platform\\ndecision_date: 2026-04-26\\n---\\n# Path Title Metadata Authority\\n\\n## Summary\\nMetadata authority: frontmatter decides path/title identity.\\n\"}}. Then run openclerk retrieval with exactly this request shape: {\"action\":\"decision_record\",\"decision_id\":\"adr-path-title-metadata-authority\"}. Then run openclerk retrieval with exactly this request shape: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"decisions\",\"ref_kind\":\"decision\",\"ref_id\":\"adr-path-title-metadata-authority\",\"limit\":5}}. In the final answer, mention records/decisions/path-title-metadata-authority.md, adr-path-title-metadata-authority, projection freshness, and that metadata/frontmatter, not filename/path, determines identity.",
		},
		{
			ID:     documentThisMissingFieldsScenarioID,
			Title:  "Document-this missing fields clarify without tools",
			Prompt: "Document this mixed article/docs/paper/transcript intake note for OpenClerk, but I did not provide document.path, document.title, or document.body.",
		},
		{
			ID:     documentThisExplicitCreateScenarioID,
			Title:  "Document-this explicit create uses strict JSON",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user supplied explicit path notes/document-this/explicit-create.md, title Document This Explicit Create, and body content. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/document-this/explicit-create.md\",\"title\":\"Document This Explicit Create\",\"body\":\"---\\ntype: note\\nstatus: active\\n---\\n# Document This Explicit Create\\n\\n## Summary\\nDocument-this explicit article/docs/paper/transcript intake uses strict runner JSON.\\nRequired fields were supplied before create_document.\\n\"}}. Do not create any sources/document-this/ document. Mention notes/document-this/explicit-create.md and Document This Explicit Create in the final answer.",
		},
		{
			ID:     documentThisSourceURLMissingHintsScenarioID,
			Title:  "Document-this source URL missing hints clarify without tools",
			Prompt: "Ingest the source artifact at https://example.test/document-this-paper.pdf into OpenClerk knowledge, but I did not provide source.path_hint or source.asset_path_hint.",
		},
		{
			ID:     documentThisExplicitOverridesScenarioID,
			Title:  "Document-this explicit overrides win",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user supplied explicit path notes/document-this/explicit-override.md and title Document This Explicit Override for mixed URLs that might otherwise look source-shaped. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/document-this/explicit-override.md\",\"title\":\"Document This Explicit Override\",\"body\":\"---\\ntype: note\\nstatus: active\\n---\\n# Document This Explicit Override\\n\\n## Summary\\nExplicit document-this override path and title win.\\nDo not infer a sources/ path from mixed URLs.\\n\"}}. Do not create any sources/document-this/ document. Mention notes/document-this/explicit-override.md and Document This Explicit Override in the final answer.",
		},
		{
			ID:     documentThisDuplicateCandidateScenarioID,
			Title:  "Document-this duplicate candidate avoids create",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user asked: document this article again: https://example.test/articles/document-this-intake. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"Document-this duplicate marker strict runner intake\",\"path_prefix\":\"sources/document-this/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/document-this/\",\"limit\":20}}. If sources/document-this/existing-article.md is present, do not create sources/document-this/duplicate-article.md. In the final answer, mention duplicate candidate, sources/document-this/existing-article.md, and that no new duplicate source was created.",
		},
		{
			ID:     documentThisExistingUpdateScenarioID,
			Title:  "Document-this existing update chooses target",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user supplied the update target path notes/document-this/existing-update.md, title Existing Document This Update, and this body section to append: ## Decisions\\nUse strict runner JSON for document-this intake. First run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/document-this/\",\"limit\":20}}. Use the returned doc_id for notes/document-this/existing-update.md to run get_document. Then append exactly this content to that document only: ## Decisions\\nUse strict runner JSON for document-this intake. Do not update notes/document-this/existing-update-decoy.md. In the final answer, mention notes/document-this/existing-update.md was updated and the decoy was not updated.",
		},
		{
			ID:     documentThisSynthesisFreshnessScenarioID,
			Title:  "Document-this synthesis freshness over duplicate",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user asked to document mixed article, docs page, paper, and transcript guidance into existing synthesis. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"document this intake pressure article docs paper transcript mixed source\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use the returned doc_id for synthesis/document-this-intake.md to run get_document. Inspect projection_states for projection synthesis with ref_kind document and that synthesis doc_id. Inspect provenance_events for ref_kind document and that synthesis doc_id. Update synthesis/document-this-intake.md only with replace_section using heading Summary and content Current document-this intake guidance: update existing synthesis after source, duplicate, provenance, and freshness checks. Keep the existing source_refs frontmatter and keep ## Sources and ## Freshness sections. Do not create synthesis/document-this-intake-copy.md. In the final answer, mention synthesis/document-this-intake.md, no duplicate synthesis, source refs or source_refs, projection freshness, and provenance.",
		},
		{
			ID:    candidateNoteFromPastedContentScenarioID,
			Title: "Candidate note from pasted content",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this note:
# Meeting Capture Policy

Capture meeting decisions within one business day.
Owners must be named next to each follow-up.

Choose a candidate strict document JSON using path notes/candidates/meeting-capture-policy.md, title Meeting Capture Policy, and this faithful body:
---
type: note
---
# Meeting Capture Policy

Capture meeting decisions within one business day.
Owners must be named next to each follow-up.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/candidates/meeting-capture-policy.md
- include the candidate title Meeting Capture Policy
- include the complete body preview exactly enough to show type: note, # Meeting Capture Policy, Capture meeting decisions within one business day., and Owners must be named next to each follow-up.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateTitleAndPathFromHeadingScenarioID,
			Title: "Candidate title and path from heading",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this:
# Release Risk Review

Risk: rollout can proceed only after rollback notes are linked.
Mitigation: document owners before release.

Choose a candidate path from the heading under notes/candidates/ and title from the heading. Build a faithful candidate body with type: note frontmatter, the supplied heading, and only the supplied facts.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the derived candidate path
- include the derived candidate title
- include the complete body preview exactly enough to show type: note, the supplied heading, Risk: rollout can proceed only after rollback notes are linked., and Mitigation: document owners before release.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateMixedSourceSummaryScenarioID,
			Title: "Candidate mixed-source summary",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, network fetching, or create_document.

The user said:
Document this mixed-source summary:
- https://example.test/articles/harness-engineering says harness notes emphasize reproducible eval setup.
- https://example.test/docs/prompt-guidance says prompt guidance notes emphasize explicit success criteria.

Choose a candidate note path notes/candidates/harness-prompt-guidance-summary.md and title Harness and Prompt Guidance Summary from the supplied text only. Use this faithful body:
---
type: note
---
# Harness and Prompt Guidance Summary

## Summary
- https://example.test/articles/harness-engineering: Harness notes emphasize reproducible eval setup.
- https://example.test/docs/prompt-guidance: Prompt guidance notes emphasize explicit success criteria.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/candidates/harness-prompt-guidance-summary.md
- include the candidate title Harness and Prompt Guidance Summary
- include the complete body preview exactly enough to show type: note, # Harness and Prompt Guidance Summary, both URLs, Harness notes emphasize reproducible eval setup., and Prompt guidance notes emphasize explicit success criteria.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateExplicitOverridesWinScenarioID,
			Title: "Candidate explicit overrides win",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this at archive/custom/intake-override.md titled Custom Intake Override:
Explicit path and title override candidate conventions.

Honor the explicit user path and title. Use path archive/custom/intake-override.md, title Custom Intake Override, and this faithful body:
---
type: note
---
# Custom Intake Override

Explicit path and title override candidate conventions.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path archive/custom/intake-override.md
- include the candidate title Custom Intake Override
- include the complete body preview exactly enough to show type: note, # Custom Intake Override, and Explicit path and title override candidate conventions.
- state that explicit user path and title win
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:     candidateDuplicateRiskAsksScenarioID,
			Title:  "Candidate duplicate risk asks before write",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document. The user said: document this pricing model note:\nPackaging tiers and renewal notes for the pricing model.\nBefore proposing a new write, run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"candidate generation duplicate pricing model marker\",\"path_prefix\":\"notes/candidates/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/candidates/\",\"limit\":20}}. If notes/candidates/existing-pricing-note.md is visible, do not create notes/candidates/pricing-model-note.md and do not validate a duplicate create. In the final answer, mention the duplicate risk, notes/candidates/existing-pricing-note.md, that no document was created, and ask whether to update the existing note or create a new one at a user-confirmed path.",
		},
		{
			ID:    candidateLowConfidenceAsksScenarioID,
			Title: "Candidate low confidence asks without tools",
			Prompt: `Document this artifact from the links we discussed last week, but I did not provide the actual content, artifact type, path, title, or body.

For this low-confidence candidate-generation request, do not use tools or commands and answer once only. Do not propose a path, title, or body. Say the request is missing actual content/body and artifact type, and ask me to provide the content, artifact type, path/title preferences if any, and the body or source text to document.`,
		},
		{
			ID:    candidateBodyFaithfulnessScenarioID,
			Title: "Candidate body faithfulness",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this escalation summary:
Customer Alpha reports two failed exports.
Impact is limited to April invoices.
Do not claim root cause yet.
Next step: compare export logs with invoice IDs.

Choose path notes/candidates/customer-escalation-summary.md and title Customer Escalation Summary. Use this faithful body:
---
type: note
---
# Customer Escalation Summary

Customer Alpha reports two failed exports.
Impact is limited to April invoices.
Do not claim root cause yet.
Next step: compare export logs with invoice IDs.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/candidates/customer-escalation-summary.md
- include the candidate title Customer Escalation Summary
- include the complete body preview exactly enough to show type: note, # Customer Escalation Summary, Customer Alpha reports two failed exports., Impact is limited to April invoices., Do not claim root cause yet., and Next step: compare export logs with invoice IDs.
- do not add root cause, all-customer, or security-incident claims
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateErgonomicsNaturalIntentScenarioID,
			Title: "Candidate ergonomics natural intent",
			Prompt: `Use the configured local OpenClerk data path and the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

Document this:
# Release Readiness Checklist

Rollback owner is assigned before release.
Support handoff notes are linked in the launch channel.
Metrics review happens the morning after launch.

I did not choose a path or title. Please propose a faithful OpenClerk document candidate from the supplied content, validate the candidate, show me the proposed path, title, and body preview, and wait for my approval before creating anything.`,
		},
		{
			ID:    candidateErgonomicsScriptedControlID,
			Title: "Candidate ergonomics scripted control",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this:
# Release Readiness Checklist

Rollback owner is assigned before release.
Support handoff notes are linked in the launch channel.
Metrics review happens the morning after launch.

Use path notes/candidates/release-readiness-checklist.md, title Release Readiness Checklist, and this faithful body:
---
type: note
---
# Release Readiness Checklist

Rollback owner is assigned before release.
Support handoff notes are linked in the launch channel.
Metrics review happens the morning after launch.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/candidates/release-readiness-checklist.md
- include the candidate title Release Readiness Checklist
- include the complete body preview exactly enough to show type: note, # Release Readiness Checklist, Rollback owner is assigned before release., Support handoff notes are linked in the launch channel., and Metrics review happens the morning after launch.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateErgonomicsDuplicateNaturalID,
			Title: "Candidate ergonomics duplicate natural intent",
			Prompt: `Use the configured local OpenClerk data path and the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

Document this pricing model note:
Packaging tiers and renewal notes for the pricing model.

Please check whether a similar note already exists before proposing a new write. If one exists, do not create or validate a duplicate; tell me the existing candidate and ask whether to update it or create a new document at a confirmed path.`,
		},
		{
			ID:     candidateErgonomicsLowConfidenceNaturalID,
			Title:  "Candidate ergonomics low confidence natural intent",
			Prompt: `Document this artifact from the links we discussed last week.`,
		},
		{
			ID:     artifactPDFSourceURLScenarioID,
			Title:  "Artifact PDF source URL ingestion",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. Run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{ARTIFACT_PDF_SOURCE_URL}}\",\"path_hint\":\"sources/artifacts/vendor-security-paper.md\",\"asset_path_hint\":\"assets/sources/artifacts/vendor-security-paper.pdf\",\"title\":\"Vendor Security Paper\"}}. In the final answer, mention sources/artifacts/vendor-security-paper.md, assets/sources/artifacts/vendor-security-paper.pdf, citation evidence such as doc_id or chunk_id, and that the PDF source URL was ingested through ingest_source_url.",
		},
		{
			ID:    artifactPDFNaturalIntentScenarioID,
			Title: "Artifact PDF source URL natural intent",
			Prompt: `Use the configured local OpenClerk data path. I need this vendor security paper PDF captured as an OpenClerk source from {{ARTIFACT_PDF_SOURCE_URL}}.

Use the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads.

Please store the source note at sources/artifacts/vendor-security-paper-natural.md, store the PDF asset at assets/sources/artifacts/vendor-security-paper-natural.pdf, title it Vendor Security Paper Natural, and answer from the runner JSON result. In your final answer, mention the source note path, asset path, citation evidence such as doc_id or chunk_id, and that the PDF source URL was ingested through ingest_source_url.`,
		},
		{
			ID:     artifactTranscriptScenarioID,
			Title:  "Artifact transcript as canonical markdown",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or native audio/video ingestion. Search for Artifact transcript canonical markdown evidence with path_prefix transcripts/ and limit 10. In the final answer, cite transcripts/artifacts/vendor-demo-transcript.md with doc_id or chunk_id, explain that supplied transcript text is already canonical markdown, and state that no native audio/video parser was needed.",
		},
		{
			ID:     artifactInvoiceReceiptScenarioID,
			Title:  "Artifact invoice and receipt authority retrieval",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for Artifact invoice receipt authority evidence with metadata_key artifact_kind and metadata_value invoice, limit 10. Also search for Artifact invoice receipt authority evidence with metadata_key artifact_kind and metadata_value receipt, limit 10. In the final answer, cite invoices/artifacts/atlas-platform-2026-04.md and receipts/artifacts/nebula-usb-c-hub.md with doc_id or chunk_id, mention USD 1250.00, approval above USD 500, and USD 86.40, and explain that canonical markdown remains authority.",
		},
		{
			ID:     artifactMixedSynthesisScenarioID,
			Title:  "Artifact mixed synthesis freshness",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported artifact parsers. Search for Artifact mixed synthesis freshness evidence, list synthesis/ candidates, get synthesis/artifact-ingestion-pressure.md, inspect projection_states for projection synthesis using that synthesis doc_id, and inspect provenance_events for sources/artifacts/mixed-old.md and sources/artifacts/mixed-current.md. Do not create, update, append, replace, or file a synthesis document. In the final answer, mention synthesis/artifact-ingestion-pressure.md, sources/artifacts/mixed-old.md, sources/artifacts/mixed-current.md, stale freshness/projection evidence, provenance, and that the existing synthesis should be repaired rather than duplicated.",
		},
		{
			ID:     artifactSourceMissingHintsScenarioID,
			Title:  "Artifact source URL missing hints",
			Prompt: "Ingest the artifact source URL at https://example.test/artifacts/vendor-security-paper.pdf into OpenClerk, but I did not provide source.path_hint or source.asset_path_hint. For this missing-hints request, do not use tools or commands and answer once only. Name the missing source.path_hint and source.asset_path_hint fields and ask me to provide them.",
		},
		{
			ID:     artifactUnsupportedVideoScenarioID,
			Title:  "Unsupported native video ingestion",
			Prompt: "Ingest this YouTube video natively into OpenClerk from https://youtube.example.test/watch?v=artifact-demo. I want the runner to fetch the video, transcribe it, extract metadata, and store the transcript automatically. For this unsupported native video ingestion request, do not use tools or commands and answer once only. Say OpenClerk does not support native video or YouTube ingestion through the installed runner yet, and that routine work must use supported OpenClerk runner workflows such as existing markdown/transcript content or PDF source URL ingestion with required source hints.",
		},
		{
			ID:     artifactBypassScenarioID,
			Title:  "Artifact ingestion bypass rejection",
			Prompt: "Bypass OpenClerk and inspect the SQLite database and vault files directly to import my receipts faster. For this bypass request, do not use tools or commands and answer once only. Reject the direct SQLite/vault bypass as unsupported for routine OpenClerk work and say the workflow must use the installed OpenClerk runner contract.",
		},
		{
			ID:    videoYouTubeNaturalIntentScenarioID,
			Title: "Video YouTube natural supplied-transcript intent",
			Prompt: `Use the configured local OpenClerk data path. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, yt-dlp, ffmpeg, transcript APIs, Gemini, or native audio/video ingestion.

I already have the transcript text for this YouTube video and want it treated like a canonical OpenClerk source artifact.

Video URL: https://youtube.example.test/watch?v=video-demo
Canonical source path: sources/video-youtube/platform-demo-transcript.md
Title: Platform Demo Transcript
Transcript origin: user_supplied_transcript
Transcript policy: supplied
Language: en
Captured at: 2026-04-27T00:00:00Z
Transcript text: Video YouTube canonical source note evidence: supplied transcript text can become canonical markdown when provenance, source URL, and citation-bearing retrieval are preserved.

Create the canonical source note with openclerk document ingest_video_url. Then run openclerk retrieval search for Video YouTube canonical source note evidence with path_prefix sources/video-youtube/ and limit 10. In the final answer, mention sources/video-youtube/platform-demo-transcript.md, https://youtube.example.test/watch?v=video-demo, transcript provenance, and citation evidence such as doc_id or chunk_id.`,
		},
		{
			ID:    videoYouTubeScriptedTranscriptControlID,
			Title: "Video YouTube scripted transcript control",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, yt-dlp, ffmpeg, transcript APIs, Gemini, or native audio/video ingestion.

Run openclerk document ingest_video_url with exactly these video fields: url https://youtube.example.test/watch?v=video-demo, path_hint sources/video-youtube/platform-demo-transcript.md, title Platform Demo Transcript, transcript.text "Video YouTube canonical source note evidence: supplied transcript text can become canonical markdown when provenance, source URL, and citation-bearing retrieval are preserved. 00:00 Speaker A: Keep video transcripts citeable as canonical source notes. 00:15 Speaker B: Preserve transcript provenance, source URL, and freshness checks before synthesis.", transcript.policy supplied, transcript.origin user_supplied_transcript, transcript.language en, transcript.captured_at 2026-04-27T00:00:00Z.

After ingest_video_url succeeds, run openclerk retrieval search for Video YouTube canonical source note evidence with path_prefix sources/video-youtube/ and limit 10. In the final answer, mention sources/video-youtube/platform-demo-transcript.md, https://youtube.example.test/watch?v=video-demo, transcript provenance, and citation evidence such as doc_id or chunk_id.`,
		},
		{
			ID:    videoYouTubeSynthesisFreshnessScenarioID,
			Title: "Video YouTube synthesis freshness",
			Prompt: `Use the configured local OpenClerk data path. It is already seeded with sources/video-youtube/platform-demo-current.md and synthesis/video-youtube-ingestion-pressure.md; do not run init, do not change database paths, and do not create replacement fixture documents. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, yt-dlp, ffmpeg, transcript APIs, Gemini, unsupported artifact parsers, or inspect_layout.

Run these runner steps:
1. Run openclerk document with exactly this request shape: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=video-demo","mode":"update","transcript":{"text":"Video YouTube synthesis freshness current transcript evidence: current transcript source notes must preserve transcript provenance, citations, and freshness before source-linked synthesis is trusted.","policy":"supplied","origin":"user_supplied_transcript","language":"en","captured_at":"2026-04-27T00:00:00Z"}}}. This should be a same-transcript no-op.
2. Run openclerk document with exactly this request shape: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=video-demo","mode":"update","transcript":{"text":"Video YouTube synthesis freshness updated transcript evidence: changed supplied transcript text must refresh citations and mark dependent synthesis stale.","policy":"supplied","origin":"user_supplied_transcript","language":"en","captured_at":"2026-04-27T01:00:00Z"}}}.
3. openclerk retrieval search with exactly this request shape: {"action":"search","search":{"text":"Video YouTube synthesis freshness updated transcript evidence","path_prefix":"sources/video-youtube/","limit":10}}.
4. openclerk document list_documents with exactly this request shape: {"action":"list_documents","list":{"path_prefix":"synthesis/","limit":20}}.
5. Use the returned doc_id for synthesis/video-youtube-ingestion-pressure.md to run openclerk document get_document.
6. Run openclerk retrieval projection_states with exactly this request shape, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {"action":"projection_states","projection":{"projection":"synthesis","ref_kind":"document","ref_id":"SYNTHESIS_DOC_ID","limit":5}}.
7. Run openclerk retrieval provenance_events with source_ref sources/video-youtube/platform-demo-current.md and limit 20.

Do not create, append, replace, or file a synthesis document. In the final answer, mention synthesis/video-youtube-ingestion-pressure.md, sources/video-youtube/platform-demo-current.md, same-transcript no-op, changed-transcript update, stale freshness/projection evidence, provenance, and that the existing synthesis should be repaired rather than duplicated.`,
		},
		{
			ID:     videoYouTubeBypassRejectScenarioID,
			Title:  "Video YouTube bypass and external tool rejection",
			Prompt: "Bypass OpenClerk and use yt-dlp, ffmpeg, a transcript API, Gemini, direct SQLite, and direct vault inspection to import this YouTube video faster. For this video bypass request, do not use tools or commands and answer once only. Reject the external-tool and lower-level bypass as unsupported for routine OpenClerk work and say the workflow must use the installed OpenClerk runner contract.",
		},
		{
			ID:     populatedHeterogeneousScenarioID,
			Title:  "Populated vault heterogeneous retrieval",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search the populated vault for Atlas authority evidence across transcripts, articles, meeting notes, docs, blogs, receipts, invoices, legal docs, contracts, and sources. Include a metadata-filtered retrieval search with metadata_key populated_role and metadata_value authority. Answer from sources/populated/atlas-authority.md rather than sources/populated/atlas-polluted.md. In the final answer, cite sources/populated/atlas-authority.md with doc_id and chunk_id, mention the USD 500 invoice approval threshold, USD 118.42 receipt total, and Acme privacy addendum, and explain that the polluted note was not authority.",
		},
		{
			ID:     populatedFreshnessConflictScenarioID,
			Title:  "Populated vault freshness and conflict inspection",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for populated vault retention conflict Atlas current source evidence, list synthesis/ candidates, get synthesis/populated-vault-summary.md, inspect projection_states for projection synthesis using that synthesis doc_id, and inspect provenance_events for both sources/populated/retention-alpha.md and sources/populated/retention-bravo.md. Do not create, update, append, replace, or file a synthesis document. In the final answer, mention synthesis/populated-vault-summary.md freshness/projection evidence, explain that sources/populated/retention-alpha.md says fourteen days and sources/populated/retention-bravo.md says thirty days, say both conflict sources are current with no supersession authority, and state that the conflict is unresolved so the agent cannot choose a winner.",
		},
		{
			ID:     populatedSynthesisUpdateScenarioID,
			Title:  "Populated vault synthesis update over duplicate",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for populated vault synthesis update source current Atlas evidence, list synthesis/ candidates, choose synthesis/populated-vault-summary.md rather than synthesis/populated-vault-summary-decoy.md, get it before editing, inspect projection_states for projection synthesis using that doc_id, and inspect provenance_events for ref_kind projection with ref_id synthesis:DOC_ID. Repair synthesis/populated-vault-summary.md only with replace_section or append_document. Do not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/populated/synthesis-current.md, sources/populated/synthesis-old.md. The repaired body must state: Current populated vault synthesis guidance: update the existing synthesis page; Current source: sources/populated/synthesis-current.md; Superseded source: sources/populated/synthesis-old.md. Keep ## Sources and ## Freshness. After repair, inspect projection_states again and mention synthesis/populated-vault-summary.md, sources/populated/synthesis-current.md, no duplicate synthesis, and final freshness in the final answer.",
		},
		{
			ID:     repoDocsRetrievalScenarioID,
			Title:  "Repo docs AgentOps retrieval dogfood",
			Prompt: "Use the configured local OpenClerk data path. The vault has been seeded from this repository's committed public markdown docs. Use only installed openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for oc-rsj verified current AgentOps document retrieval runner actions with path_prefix docs/architecture/ and limit 10. Answer this question from the repo docs only: what is OpenClerk's current production agent surface? In the final answer, cite docs/architecture/eval-backed-knowledge-plane-adr.md and include citation evidence such as doc_id and chunk_id.",
		},
		{
			ID:     repoDocsSynthesisScenarioID,
			Title:  "Repo docs synthesis maintenance dogfood",
			Prompt: "Use the configured local OpenClerk data path. The vault has been seeded from this repository's committed public markdown docs. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for production AgentOps gate baseline scenarios runner JSON validation with path_prefix docs/evals/ and limit 10. List synthesis/ candidates before writing. Then create synthesis/repo-docs-agentops-validation.md titled Repo Docs AgentOps Validation with frontmatter type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: docs/evals/agent-production.md, docs/evals/baseline-scenarios.md. Do not use YAML list syntax for source_refs. The body must include these exact lines: Repo-docs dogfood decision: use the existing OpenClerk document and retrieval runner actions.; Production gate source: docs/evals/agent-production.md; Baseline scenarios source: docs/evals/baseline-scenarios.md. Include ## Sources with both source paths and ## Freshness describing the runner search and synthesis-candidate checks. Mention synthesis/repo-docs-agentops-validation.md in the final answer.",
		},
		{
			ID:     repoDocsDecisionScenarioID,
			Title:  "Repo docs decision-record dogfood",
			Prompt: "Use the configured local OpenClerk data path. The vault has been seeded from this repository's committed public markdown docs. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. First search for Knowledge Configuration v1 accepted AgentOps surface with path_prefix docs/architecture/ and limit 10. Then use decisions_lookup for the accepted platform knowledge-configuration decision. Then use decision_record for adr-agentops-only-knowledge-plane. Inspect projection_states for projection decisions for both adr-knowledge-configuration-v1 and adr-agentops-only-knowledge-plane. Inspect provenance_events for ref_kind projection and ref_id decisions:adr-knowledge-configuration-v1. In the final answer, explain that canonical markdown ADRs remain authoritative while decision records are derived, report fresh projection/provenance evidence, and include citation paths docs/architecture/eval-backed-knowledge-plane-adr.md and docs/architecture/knowledge-configuration-v1-adr.md.",
		},
		{
			ID:     synthesisCandidatePressureScenarioID,
			Title:  "Pressure-test synthesis candidate selection",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or unsupported actions such as upsert_document. Search for synthesis compiler pressure evidence, list synthesis/ candidates, choose the existing compiler pressure synthesis rather than the decoy, get it before editing, inspect its synthesis projection freshness, and repair it only with replace_section or append_document. Do not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/compiler-current.md and sources/compiler-old.md. The repaired body must state: Current compiler decision: existing document and retrieval actions are sufficient for synthesis compiler pressure repairs; Current source: sources/compiler-current.md; Superseded source: sources/compiler-old.md. Keep ## Sources and ## Freshness. Mention synthesis/compiler-routing.md and the final freshness in the final answer.",
		},
		{
			ID:     synthesisSourceSetPressureScenarioID,
			Title:  "Pressure-test multi-source synthesis creation",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or unsupported actions such as upsert_document. Search for synthesis compiler pressure source set evidence, list synthesis/ candidates, then create synthesis/compiler-source-set.md as a new source-linked synthesis. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/source-set-alpha.md, sources/source-set-beta.md, sources/source-set-gamma.md. Do not use YAML list syntax for source_refs. The body must mention alpha, beta, and gamma source evidence, include ## Sources with all three source paths, and include ## Freshness describing the runner search and synthesis-candidate checks. Mention synthesis/compiler-source-set.md in the final answer.",
		},
		{
			ID:     "append-replace",
			Title:  "Append and replace sections",
			Prompt: "Use the configured local OpenClerk data path. Append a Decisions section to notes/projects/openclerk-runner.md, then replace only that Decisions section with: Use the JSON runner for routine AgentOps knowledge tasks. Do not remove the existing Context section.",
		},
		{
			ID:     "records-provenance",
			Title:  "Records and provenance inspection",
			Prompt: "Use the configured local OpenClerk data path. Inspect the promoted-record-shaped OpenClerk runner document through records_lookup, provenance_events, and projection_states. Report the records lookup result plus provenance event and projection freshness details.",
		},
		{
			ID:     "promoted-record-vs-docs",
			Title:  "Compare promoted records against plain docs",
			Prompt: "Use the configured local OpenClerk data path. Search plain docs for OpenClerk runner evidence, then run services lookup for OpenClerk runner. Compare plain docs/search against services lookup for this service-centric question: what is the production interface? The final answer must mention plain docs or search, services lookup or service registry, and JSON runner.",
		},
		{
			ID:     decisionRecordVsDocsScenarioID,
			Title:  "Compare decision records against plain docs",
			Prompt: "Use the configured local OpenClerk data path. Search plain docs for OpenClerk runner decision evidence, then run decisions_lookup for the accepted AgentOps JSON runner decision with status accepted, scope agentops, and owner platform. Compare plain docs/search against decisions_lookup for this decision-centric question: what is the current accepted runner decision? Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct SQLite, openclerk --help, or source-built command paths. The final answer must mention plain docs or search, decisions lookup or decision records, status/scope filtering, JSON runner, and citation details from the decision result.",
		},
		{
			ID:     decisionSupersessionScenarioID,
			Title:  "Inspect decision supersession and freshness",
			Prompt: "Use the configured local OpenClerk data path. Run decision_record for adr-runner-old and run decision_record for adr-runner-current; do not substitute decisions_lookup for those two detail requests. Then inspect projection_states for projection decisions for both decision ids, and inspect provenance_events for the current decision projection. Use only OpenClerk runner retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct SQLite, openclerk --help, or source-built command paths. In the final answer, report that adr-runner-old is superseded/stale, adr-runner-current supersedes it and is fresh, and mention provenance/projection evidence plus citation paths.",
		},
		{
			ID:     decisionRealADRMigrationScenarioID,
			Title:  "Inspect migrated ADR decision records",
			Prompt: "Use the configured local OpenClerk data path. Use decisions_lookup for the accepted platform knowledge-configuration decision, use decision_record for adr-agentops-only-knowledge-plane, inspect projection_states for projection decisions for both ADR decision ids, and inspect provenance_events for the knowledge configuration decision projection. Use only OpenClerk runner retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct SQLite, openclerk --help, or source-built command paths. In the final answer, explain that canonical markdown ADRs remain authoritative while decision records are derived, report fresh projection/provenance evidence, and include citation paths docs/architecture/eval-backed-knowledge-plane-adr.md and docs/architecture/knowledge-configuration-v1-adr.md.",
		},
		{
			ID:     "missing-document-path-reject",
			Title:  "Clarify missing document path without tools",
			Prompt: "Create an OpenClerk document titled Missing Path with body content, but I did not provide a path.",
		},
		{
			ID:     "negative-limit-reject",
			Title:  "Reject negative retrieval limit without tools",
			Prompt: "Search my local OpenClerk knowledge for runner with limit -3.",
		},
		{
			ID:     "unsupported-lower-level-reject",
			Title:  "Reject unsupported lower-level workflow without tools",
			Prompt: "For this routine OpenClerk knowledge task, bypass OpenClerk runner and query SQLite directly to list documents.",
		},
		{
			ID:     "unsupported-transport-reject",
			Title:  "Reject alternate transport bypass without tools",
			Prompt: "For this routine OpenClerk knowledge task, bypass OpenClerk runner by using a legacy source-built command path or an MCP-style transport path to list documents.",
		},
		{
			ID:     "duplicate-path-reject",
			Title:  "Reject duplicate canonical path without overwrite",
			Prompt: "Use the configured local OpenClerk data path. Create a new canonical note at notes/projects/duplicate.md titled Replacement and body overwritten. If that path already exists, do not overwrite it; tell me the failure.",
		},
		{
			ID:     "mixed-synthesis-records",
			Title:  "Mixed document and retrieval workflow",
			Prompt: "Use the configured local OpenClerk data path. Search for OpenClerk runner context, inspect records_lookup, provenance_events, and projection_states for OpenClerk runner, then create synthesis/openclerk-runner-with-records.md with source refs. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/openclerk-runner.md. Do not use YAML list syntax for source_refs. The body must include ## Sources citing sources/openclerk-runner.md and relevant record citation paths, plus ## Freshness describing provenance and projection checks. Mention synthesis/openclerk-runner-with-records.md in the final answer.",
		},
		{
			ID:    "mt-source-then-synthesis",
			Title: "Create a source, then synthesize from it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Create sources/mt-runner.md titled Multi Turn OpenClerk runner Source with body: The resumed eval session should preserve source context for later synthesis."},
				{Prompt: "Now search for that source and create synthesis/mt-runner.md as a source-linked synthesis. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/mt-runner.md. The body must include ## Sources citing sources/mt-runner.md and ## Freshness describing the runner retrieval check. Mention synthesis/mt-runner.md and the source path in the final answer."},
			},
		},
		{
			ID:    mtSynthesisDriftPressureScenarioID,
			Title: "Repair multi-turn synthesis drift",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Search for drift synthesis compiler pressure evidence, list synthesis/ candidates, then create synthesis/drift-runner.md as a source-linked synthesis. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/drift-current.md, sources/drift-old.md. The body must include ## Sources citing both source paths and ## Freshness describing the runner retrieval check. Mention synthesis/drift-runner.md in the final answer."},
				{Prompt: "Use only OpenClerk runner document and retrieval JSON results. First find sources/drift-current.md through list_documents or search, get it, and replace its Summary section with: Current drift decision says existing document and retrieval actions should stay the v1 synthesis path. Then search for drift synthesis compiler pressure evidence, list synthesis/ candidates, get synthesis/drift-runner.md, inspect projection_states for projection synthesis using that document id, and repair synthesis/drift-runner.md only with replace_section or append_document. Do not create a duplicate. Preserve the existing single-line source_refs for sources/drift-current.md and sources/drift-old.md. The repaired body must state: Current drift decision: keep existing document and retrieval actions; Current source: sources/drift-current.md; Superseded source: sources/drift-old.md. Mention synthesis/drift-runner.md, sources/drift-current.md, and final freshness in the final answer."},
			},
		},
		{
			ID:    "mt-incomplete-then-create",
			Title: "Clarify incomplete request, then complete it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Create an OpenClerk canonical project note, but I have not provided the path, title, or body yet."},
				{Prompt: "Use path notes/projects/mt-complete.md, title Multi Turn Complete, and body: Multi-turn completion should use the OpenClerk runner after required fields are provided."},
			},
		},
	}
}

func scenarioIDs() []string {
	scenarios := allScenarios()
	ids := make([]string, 0, len(scenarios))
	for _, sc := range scenarios {
		ids = append(ids, sc.ID)
	}
	return ids
}

func releaseBlockingScenarioIDs() []string {
	ids := []string{}
	for _, id := range scenarioIDs() {
		if isReleaseBlockingScenario(id) {
			ids = append(ids, id)
		}
	}
	return ids
}

func scenarioTurns(sc scenario) []scenarioTurn {
	if len(sc.Turns) > 0 {
		return sc.Turns
	}
	return []scenarioTurn{{Prompt: sc.Prompt}}
}

func isMultiTurnScenario(sc scenario) bool {
	return len(scenarioTurns(sc)) > 1
}

func isFinalAnswerOnlyValidationScenario(id string) bool {
	switch id {
	case "missing-document-path-reject", agentChosenMissingFieldsScenarioID, pathTitleArtifactMissingHintsScenarioID, documentThisMissingFieldsScenarioID, documentThisSourceURLMissingHintsScenarioID, artifactSourceMissingHintsScenarioID, artifactUnsupportedVideoScenarioID, artifactBypassScenarioID, videoYouTubeBypassRejectScenarioID, "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject":
		return true
	default:
		return false
	}
}

func promptSummary(sc scenario) string {
	if len(sc.Turns) == 0 {
		return sc.Prompt
	}
	parts := make([]string, 0, len(sc.Turns))
	for i, turn := range sc.Turns {
		parts = append(parts, fmt.Sprintf("turn %d: %s", i+1, turn.Prompt))
	}
	return strings.Join(parts, " | ")
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func containsArgPair(args []string, key string, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}
