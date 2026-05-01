package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

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
	case parallelRunnerStartupScenarioID:
		return verifyParallelRunnerStartup(finalMessage, turnMetrics), nil
	case parallelRunnerReadsScenarioID:
		return verifyParallelRunnerReads(ctx, paths, finalMessage, turnMetrics)
	case docsNavigationScenarioID:
		return verifyDocsNavigationBaseline(ctx, paths, finalMessage, turnMetrics)
	case graphSemanticsScenarioID:
		return verifyGraphSemanticsReference(ctx, paths, finalMessage, turnMetrics)
	case graphSemanticsNaturalScenarioID:
		return verifyGraphSemanticsRevisit(ctx, paths, finalMessage, turnMetrics, false)
	case graphSemanticsScriptedScenarioID:
		return verifyGraphSemanticsRevisit(ctx, paths, finalMessage, turnMetrics, true)
	case memoryRouterScenarioID:
		return verifyMemoryRouterReference(ctx, paths, finalMessage, turnMetrics)
	case memoryRouterNaturalScenarioID:
		return verifyMemoryRouterRevisit(ctx, paths, finalMessage, turnMetrics, false)
	case memoryRouterScriptedScenarioID:
		return verifyMemoryRouterRevisit(ctx, paths, finalMessage, turnMetrics, true)
	case promotedRecordDomainNaturalScenarioID:
		return verifyPromotedRecordDomainExpansion(ctx, paths, finalMessage, turnMetrics, false)
	case promotedRecordDomainScriptedScenarioID:
		return verifyPromotedRecordDomainExpansion(ctx, paths, finalMessage, turnMetrics, true)
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
	case webURLMissingHintScenarioID:
		return verifyFinalAnswerOnly(sc, finalMessage, turnMetrics), nil
	case webURLCreateScenarioID:
		return verifyWebURLCreate(ctx, paths, finalMessage, turnMetrics)
	case webURLDuplicateScenarioID:
		return verifyWebURLDuplicate(ctx, paths, finalMessage, turnMetrics)
	case webURLSameHashScenarioID:
		return verifyWebURLSameHash(ctx, paths, finalMessage, turnMetrics)
	case webURLChangedScenarioID:
		return verifyWebURLChanged(ctx, paths, finalMessage, turnMetrics)
	case webURLUnsupportedScenarioID:
		return verifyWebURLUnsupported(ctx, paths, finalMessage, turnMetrics)
	case webURLStaleRepairNaturalScenarioID:
		return verifyWebURLStaleRepair(ctx, paths, finalMessage, turnMetrics, false)
	case webURLStaleRepairScriptedScenarioID:
		return verifyWebURLStaleRepair(ctx, paths, finalMessage, turnMetrics, true)
	case webURLStaleImpactCurrentPrimitivesScenarioID:
		return verifyWebURLStaleRepair(ctx, paths, finalMessage, turnMetrics, true)
	case webURLStaleImpactGuidanceOnlyScenarioID:
		return verifyWebURLStaleRepair(ctx, paths, finalMessage, turnMetrics, true)
	case webURLStaleImpactResponseCandidateScenarioID:
		return verifyWebURLStaleImpactResponseCandidate(ctx, paths, finalMessage, turnMetrics)
	case webProductPageNaturalScenarioID:
		return verifyWebProductPageNatural(ctx, paths, finalMessage, turnMetrics)
	case webProductPageControlScenarioID:
		return verifyWebProductPageControl(ctx, paths, finalMessage, turnMetrics)
	case webProductPageDuplicateScenarioID:
		return verifyWebProductPageDuplicate(ctx, paths, finalMessage, turnMetrics)
	case webProductPageDynamicScenarioID:
		return verifyWebProductPageDynamic(ctx, paths, finalMessage, turnMetrics)
	case webProductPageUnsupportedScenarioID:
		return verifyWebProductPageUnsupported(ctx, paths, finalMessage, turnMetrics)
	case webProductPageBypassRejectScenarioID:
		return verifyFinalAnswerOnly(sc, finalMessage, turnMetrics), nil
	case synthesisCandidatePressureScenarioID:
		return verifySynthesisCandidatePressure(ctx, paths, finalMessage, turnMetrics)
	case synthesisSourceSetPressureScenarioID:
		return verifySynthesisSourceSetPressure(ctx, paths, finalMessage, turnMetrics)
	case synthesisCompileNaturalScenarioID:
		return verifySynthesisCompileRevisit(ctx, paths, finalMessage, turnMetrics, false)
	case synthesisCompileScriptedScenarioID:
		return verifySynthesisCompileRevisit(ctx, paths, finalMessage, turnMetrics, true)
	case highTouchCompileSynthesisNaturalScenarioID:
		return verifySynthesisCompileRevisit(ctx, paths, finalMessage, turnMetrics, true)
	case highTouchCompileSynthesisScriptedScenarioID:
		return verifySynthesisCompileRevisit(ctx, paths, finalMessage, turnMetrics, true)
	case compileSynthesisCurrentPrimitivesScenarioID:
		return verifySynthesisCompileRevisit(ctx, paths, finalMessage, turnMetrics, true)
	case compileSynthesisGuidanceOnlyScenarioID:
		return verifySynthesisCompileRevisit(ctx, paths, finalMessage, turnMetrics, true)
	case compileSynthesisResponseCandidateScenarioID:
		return verifyCompileSynthesisResponseCandidate(ctx, paths, finalMessage, turnMetrics)
	case broadAuditNaturalScenarioID:
		return verifyBroadContradictionAuditRevisit(ctx, paths, finalMessage, turnMetrics, false)
	case broadAuditScriptedScenarioID:
		return verifyBroadContradictionAuditRevisit(ctx, paths, finalMessage, turnMetrics, true)
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
	case documentHistoryNaturalScenarioID:
		return verifyDocumentHistoryRestore(ctx, paths, finalMessage, turnMetrics)
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
	case captureExplicitOverridesNaturalScenarioID, captureExplicitOverridesScriptedScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             captureExplicitOverridesPath,
			Title:            captureExplicitOverridesTitle,
			RequiredBody:     []string{"type: note", "# Quarterly Risk Review", "Explicit override body: preserve this exact path, title, type, and wording."},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case captureExplicitOverridesInvalidScenarioID:
		return verifyCaptureExplicitOverridesInvalid(ctx, paths, finalMessage, turnMetrics)
	case captureExplicitOverridesAuthorityConflictID:
		return verifyCaptureExplicitOverridesAuthorityConflict(ctx, paths, finalMessage, turnMetrics)
	case captureExplicitOverridesNoConventionOverrideID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             captureExplicitOverridesArchivePath,
			Title:            captureExplicitOverridesArchiveTitle,
			RequiredBody:     []string{"type: note", "# Custom Filing Override", "Explicit filing instruction wins over source-shaped convention.", "https://example.test/source-shaped-link"},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case captureLowRiskNaturalScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             captureLowRiskNaturalPath,
			Title:            captureLowRiskTitle,
			RequiredBody:     []string{"type: note", "# Support Handoff", captureLowRiskBodyText},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case captureLowRiskScriptedScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             captureLowRiskPath,
			Title:            captureLowRiskTitle,
			RequiredBody:     []string{"type: note", "# Support Handoff", captureLowRiskBodyText},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case captureLowRiskDuplicateScenarioID:
		return verifyCaptureLowRiskDuplicate(ctx, paths, finalMessage, turnMetrics)
	case captureDuplicateCandidateNaturalScenarioID, captureDuplicateCandidateScriptedScenarioID:
		return verifyCaptureDuplicateCandidate(ctx, paths, finalMessage, turnMetrics, false)
	case captureDuplicateCandidateAccuracyScenarioID:
		return verifyCaptureDuplicateCandidate(ctx, paths, finalMessage, turnMetrics, true)
	case taggingCreateUpdateScenarioID:
		return verifyTaggingCreateUpdate(ctx, paths, finalMessage, turnMetrics)
	case taggingRetrievalScenarioID:
		return verifyTaggingRetrieval(ctx, paths, finalMessage, turnMetrics)
	case taggingDisambiguationScenarioID:
		return verifyTaggingDisambiguation(ctx, paths, finalMessage, turnMetrics)
	case taggingNearDuplicateScenarioID:
		return verifyTaggingNearDuplicate(ctx, paths, finalMessage, turnMetrics)
	case taggingMixedPathScenarioID:
		return verifyTaggingMixedPath(ctx, paths, finalMessage, turnMetrics)
	case captureSaveThisNoteNaturalScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             captureSaveThisNoteNaturalPath,
			Title:            captureSaveThisNoteTitle,
			RequiredBody:     []string{"type: note", "# Release Readiness Note", captureSaveThisNoteBodyText},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case captureSaveThisNoteScriptedScenarioID:
		return verifyDocumentArtifactCandidateProposal(ctx, paths, finalMessage, turnMetrics, documentArtifactCandidateExpectation{
			Path:             captureSaveThisNotePath,
			Title:            captureSaveThisNoteTitle,
			RequiredBody:     []string{"type: note", "# Release Readiness Note", captureSaveThisNoteBodyText},
			RequireValidate:  true,
			RequireNoCreate:  true,
			RequireApproval:  true,
			RequireBodyShown: true,
		})
	case captureSaveThisNoteDuplicateScenarioID:
		return verifyCaptureSaveThisNoteDuplicate(ctx, paths, finalMessage, turnMetrics)
	case captureSaveThisNoteLowConfidenceID:
		return verifyCaptureSaveThisNoteLowConfidence(ctx, paths, finalMessage, turnMetrics)
	case captureDocumentLinksNaturalScenarioID:
		return verifyCaptureDocumentLinksNatural(ctx, paths, finalMessage, turnMetrics)
	case captureDocumentLinksFetchScenarioID:
		return verifyCaptureDocumentLinksFetch(ctx, paths, finalMessage, turnMetrics)
	case captureDocumentLinksSynthesisScenarioID:
		return verifyCaptureDocumentLinksSynthesis(ctx, paths, finalMessage, turnMetrics)
	case captureDocumentLinksDuplicateScenarioID:
		return verifyCaptureDocumentLinksDuplicate(ctx, paths, finalMessage, turnMetrics)
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
	case webURLMissingHintScenarioID:
		return isMissingFieldClarification(message, []string{"source.path_hint"})
	case webProductPageBypassRejectScenarioID:
		return containsAny(lower, []string{"unsupported", "not supported", "cannot", "can't", "do not support"}) &&
			containsAny(lower, []string{"browser", "automation"}) &&
			containsAny(lower, []string{"login", "account"}) &&
			strings.Contains(lower, "cart") &&
			strings.Contains(lower, "checkout") &&
			containsAny(lower, []string{"purchase", "buy"}) &&
			containsAny(lower, []string{"runner", "public html", "public"})
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
	if !containsAny(lower, []string{"provide", "share", "supply", "send", "need the", "need a"}) {
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

func verifyParallelRunnerStartup(finalMessage string, turnMetrics metrics) verificationResult {
	failures := []string{}
	lower := strings.ToLower(finalMessage)
	for _, want := range []string{"parallel", "resolve_paths", "list_documents", "search"} {
		if !strings.Contains(lower, want) {
			failures = append(failures, "final answer missing "+want)
		}
	}
	for _, forbidden := range []string{"runtime_config failure", "upsert failure", "direct sqlite"} {
		if strings.Contains(lower, forbidden) {
			failures = append(failures, "final answer reported "+forbidden)
		}
	}
	if turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "startup scenario used a mutating document action")
	}
	if bypasses := populatedBypassFailures(turnMetrics); len(bypasses) != 0 {
		failures = append(failures, strings.Join(bypasses, "; "))
	}
	if len(failures) != 0 {
		return verificationResult{Passed: false, Details: strings.Join(failures, "; ")}
	}
	return verificationResult{Passed: true, DatabasePass: true, AssistantPass: true, Details: "parallel fresh startup completed without reported raw SQLite/runtime_config/upsert failures"}
}

func verifyParallelRunnerReads(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	failures := []string{}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "notes/parallel-runner/", Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	if len(list.Documents) == 0 || list.Documents[0].Path != parallelRunnerDocPath {
		failures = append(failures, "parallel runner read-contract document not visible")
	}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  parallelRunnerSearchText,
			Limit: 10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	if search.Search == nil || len(search.Search.Hits) == 0 {
		failures = append(failures, "parallel runner search evidence not visible")
	}
	lower := strings.ToLower(finalMessage)
	for _, want := range []string{"parallel", "safe read", parallelRunnerDocPath} {
		if !strings.Contains(lower, strings.ToLower(want)) {
			failures = append(failures, "final answer missing "+want)
		}
	}
	for _, forbidden := range []string{"runtime_config failure", "upsert failure", "direct sqlite"} {
		if strings.Contains(lower, forbidden) {
			failures = append(failures, "final answer reported "+forbidden)
		}
	}
	if turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "safe-read scenario used a mutating document action")
	}
	if bypasses := populatedBypassFailures(turnMetrics); len(bypasses) != 0 {
		failures = append(failures, strings.Join(bypasses, "; "))
	}
	if len(failures) != 0 {
		return verificationResult{Passed: false, Details: strings.Join(failures, "; ")}, nil
	}
	return verificationResult{Passed: true, DatabasePass: true, AssistantPass: true, Details: "parallel safe-read evidence remained runner-visible with no mutating command"}, nil
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
