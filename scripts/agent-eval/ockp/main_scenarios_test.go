package main

import (
	"strings"
	"testing"
)

func TestScenarioIDsIncludeADRProofObligations(t *testing.T) {
	ids := map[string]bool{}
	for _, id := range scenarioIDs() {
		ids[id] = true
	}
	for _, want := range []string{"answer-filing", ragRetrievalScenarioID, parallelRunnerStartupScenarioID, parallelRunnerReadsScenarioID, docsNavigationScenarioID, graphSemanticsScenarioID, graphSemanticsNaturalScenarioID, graphSemanticsScriptedScenarioID, memoryRouterNaturalScenarioID, memoryRouterScriptedScenarioID, promotedRecordDomainNaturalScenarioID, promotedRecordDomainScriptedScenarioID, broadAuditNaturalScenarioID, broadAuditScriptedScenarioID, memoryRouterScenarioID, configuredLayoutScenarioID, invalidLayoutScenarioID, sourceURLUpdateDuplicateScenarioID, sourceURLUpdateSameSHAScenarioID, sourceURLUpdateChangedScenarioID, sourceURLUpdateConflictScenarioID, synthesisCandidatePressureScenarioID, synthesisSourceSetPressureScenarioID, synthesisCompileNaturalScenarioID, synthesisCompileScriptedScenarioID, decisionRecordVsDocsScenarioID, decisionSupersessionScenarioID, sourceAuditRepairScenarioID, sourceAuditConflictScenarioID, documentHistoryNaturalScenarioID, documentHistoryInspectScenarioID, documentHistoryDiffScenarioID, documentHistoryRestoreScenarioID, documentHistoryPendingScenarioID, documentHistoryStaleScenarioID, populatedHeterogeneousScenarioID, populatedFreshnessConflictScenarioID, populatedSynthesisUpdateScenarioID, agentChosenExplicitScenarioID, agentChosenMissingFieldsScenarioID, agentChosenPathProposalScenarioID, agentChosenAutonomousScenarioID, agentChosenSynthesisScenarioID, agentChosenAmbiguousScenarioID, agentChosenUserPathScenarioID, pathTitleURLOnlyScenarioID, pathTitleArtifactMissingHintsScenarioID, pathTitleMultiSourceDuplicateScenarioID, pathTitleExplicitOverridesScenarioID, pathTitleDuplicateRiskScenarioID, pathTitleMetadataAuthorityScenarioID, captureLowRiskNaturalScenarioID, captureLowRiskScriptedScenarioID, captureLowRiskDuplicateScenarioID, captureExplicitOverridesNaturalScenarioID, captureExplicitOverridesScriptedScenarioID, captureExplicitOverridesInvalidScenarioID, captureExplicitOverridesAuthorityConflictID, captureExplicitOverridesNoConventionOverrideID, captureDuplicateCandidateNaturalScenarioID, captureDuplicateCandidateScriptedScenarioID, captureDuplicateCandidateAccuracyScenarioID, captureSaveThisNoteNaturalScenarioID, captureSaveThisNoteScriptedScenarioID, captureSaveThisNoteDuplicateScenarioID, captureSaveThisNoteLowConfidenceID, captureDocumentLinksNaturalScenarioID, captureDocumentLinksFetchScenarioID, captureDocumentLinksSynthesisScenarioID, captureDocumentLinksDuplicateScenarioID, documentThisMissingFieldsScenarioID, documentThisExplicitCreateScenarioID, documentThisSourceURLMissingHintsScenarioID, documentThisExplicitOverridesScenarioID, documentThisDuplicateCandidateScenarioID, documentThisExistingUpdateScenarioID, documentThisSynthesisFreshnessScenarioID, candidateNoteFromPastedContentScenarioID, candidateTitleAndPathFromHeadingScenarioID, candidateMixedSourceSummaryScenarioID, candidateExplicitOverridesWinScenarioID, candidateDuplicateRiskAsksScenarioID, candidateLowConfidenceAsksScenarioID, candidateBodyFaithfulnessScenarioID, artifactPDFSourceURLScenarioID, artifactPDFNaturalIntentScenarioID, artifactTranscriptScenarioID, artifactInvoiceReceiptScenarioID, artifactMixedSynthesisScenarioID, artifactSourceMissingHintsScenarioID, artifactUnsupportedVideoScenarioID, artifactBypassScenarioID, videoYouTubeNaturalIntentScenarioID, videoYouTubeScriptedTranscriptControlID, videoYouTubeSynthesisFreshnessScenarioID, videoYouTubeBypassRejectScenarioID, mtSynthesisDriftPressureScenarioID, "stale-synthesis-update", "promoted-record-vs-docs", "unsupported-transport-reject"} {
		if !ids[want] {
			t.Fatalf("scenarioIDs missing %q in %v", want, scenarioIDs())
		}
	}
	for _, want := range webProductPageScenarioIDs() {
		if !ids[want] {
			t.Fatalf("scenarioIDs missing web product-page scenario %q in %v", want, scenarioIDs())
		}
	}
	for _, want := range webURLStaleRepairScenarioIDs() {
		if !ids[want] {
			t.Fatalf("scenarioIDs missing web URL stale repair scenario %q in %v", want, scenarioIDs())
		}
	}
	for _, want := range highTouchCompileSynthesisScenarioIDs() {
		if !ids[want] {
			t.Fatalf("scenarioIDs missing high-touch compile synthesis scenario %q in %v", want, scenarioIDs())
		}
	}
	for _, want := range webURLStaleImpactScenarioIDs() {
		if !ids[want] {
			t.Fatalf("scenarioIDs missing web URL stale impact scenario %q in %v", want, scenarioIDs())
		}
	}
}

func TestDefaultScenarioSelectionExcludesPopulatedTargetedLane(t *testing.T) {
	defaultIDs := map[string]bool{}
	for _, scenario := range selectedScenarios(runConfig{}) {
		defaultIDs[scenario.ID] = true
	}
	for _, id := range []string{populatedHeterogeneousScenarioID, populatedFreshnessConflictScenarioID, populatedSynthesisUpdateScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted populated scenario %q", id)
		}
	}
	for _, id := range documentHistoryScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted document history scenario %q", id)
		}
	}
	for _, id := range []string{agentChosenExplicitScenarioID, agentChosenMissingFieldsScenarioID, agentChosenPathProposalScenarioID, agentChosenAutonomousScenarioID, agentChosenSynthesisScenarioID, agentChosenAmbiguousScenarioID, agentChosenUserPathScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted agent-chosen path scenario %q", id)
		}
	}
	for _, id := range []string{pathTitleURLOnlyScenarioID, pathTitleArtifactMissingHintsScenarioID, pathTitleMultiSourceDuplicateScenarioID, pathTitleExplicitOverridesScenarioID, pathTitleDuplicateRiskScenarioID, pathTitleMetadataAuthorityScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted path-title scenario %q", id)
		}
	}
	for _, id := range captureLowRiskScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted capture low-risk scenario %q", id)
		}
	}
	for _, id := range captureExplicitOverridesScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted capture explicit overrides scenario %q", id)
		}
	}
	for _, id := range captureDuplicateCandidateScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted capture duplicate candidate scenario %q", id)
		}
	}
	for _, id := range captureSaveThisNoteScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted capture save-this-note scenario %q", id)
		}
	}
	for _, id := range captureDocumentLinksScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted capture document-these-links scenario %q", id)
		}
	}
	for _, id := range []string{sourceURLUpdateDuplicateScenarioID, sourceURLUpdateSameSHAScenarioID, sourceURLUpdateChangedScenarioID, sourceURLUpdateConflictScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted source URL update scenario %q", id)
		}
	}
	for _, id := range webProductPageScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted web product-page scenario %q", id)
		}
	}
	for _, id := range webURLStaleRepairScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted web URL stale repair scenario %q", id)
		}
	}
	for _, id := range webURLStaleImpactScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted web URL stale impact scenario %q", id)
		}
	}
	for _, id := range []string{documentThisMissingFieldsScenarioID, documentThisExplicitCreateScenarioID, documentThisSourceURLMissingHintsScenarioID, documentThisExplicitOverridesScenarioID, documentThisDuplicateCandidateScenarioID, documentThisExistingUpdateScenarioID, documentThisSynthesisFreshnessScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted document-this scenario %q", id)
		}
	}
	for _, id := range []string{candidateNoteFromPastedContentScenarioID, candidateTitleAndPathFromHeadingScenarioID, candidateMixedSourceSummaryScenarioID, candidateExplicitOverridesWinScenarioID, candidateDuplicateRiskAsksScenarioID, candidateLowConfidenceAsksScenarioID, candidateBodyFaithfulnessScenarioID} {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted document artifact candidate scenario %q", id)
		}
	}
	for _, id := range artifactIngestionScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted artifact ingestion scenario %q", id)
		}
	}
	for _, id := range videoYouTubeScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted video/YouTube scenario %q", id)
		}
	}
	for _, id := range synthesisCompileScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted synthesis compile scenario %q", id)
		}
	}
	for _, id := range highTouchCompileSynthesisScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted high-touch compile synthesis scenario %q", id)
		}
	}
	for _, id := range graphSemanticsRevisitScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted graph semantics revisit scenario %q", id)
		}
	}
	for _, id := range memoryRouterRevisitScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted memory/router revisit scenario %q", id)
		}
	}
	for _, id := range promotedRecordDomainScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted promoted record domain scenario %q", id)
		}
	}
	for _, id := range parallelRunnerScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted parallel runner scenario %q", id)
		}
	}
	for _, id := range broadAuditScenarioIDs() {
		if defaultIDs[id] {
			t.Fatalf("default selected scenarios included targeted broad audit scenario %q", id)
		}
	}
	selected := selectedScenarioIDs(runConfig{Scenario: populatedHeterogeneousScenarioID + "," + populatedFreshnessConflictScenarioID + "," + populatedSynthesisUpdateScenarioID})
	lane, releaseBlocking := reportLane(selected)
	if lane != populatedLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, populatedLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: documentHistoryNaturalScenarioID + "," + documentHistoryInspectScenarioID + "," + documentHistoryDiffScenarioID + "," + documentHistoryRestoreScenarioID + "," + documentHistoryPendingScenarioID + "," + documentHistoryStaleScenarioID + ",missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject"})
	lane, releaseBlocking = reportLane(selected)
	if lane != documentHistoryLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, documentHistoryLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: agentChosenExplicitScenarioID + "," + agentChosenMissingFieldsScenarioID + "," + agentChosenPathProposalScenarioID + "," + agentChosenAutonomousScenarioID + "," + agentChosenSynthesisScenarioID + "," + agentChosenAmbiguousScenarioID + "," + agentChosenUserPathScenarioID + ",missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject"})
	lane, releaseBlocking = reportLane(selected)
	if lane != agentChosenPathLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, agentChosenPathLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: pathTitleURLOnlyScenarioID + "," + pathTitleArtifactMissingHintsScenarioID + "," + pathTitleMultiSourceDuplicateScenarioID + "," + pathTitleExplicitOverridesScenarioID + "," + pathTitleDuplicateRiskScenarioID + "," + pathTitleMetadataAuthorityScenarioID})
	lane, releaseBlocking = reportLane(selected)
	if lane != pathTitleAutonomyLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, pathTitleAutonomyLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(captureLowRiskScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != captureLowRiskLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, captureLowRiskLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(captureExplicitOverridesScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != captureExplicitOverridesLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, captureExplicitOverridesLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(captureDuplicateCandidateScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != captureDuplicateCandidateLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, captureDuplicateCandidateLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(captureSaveThisNoteScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != captureSaveThisNoteLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, captureSaveThisNoteLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(captureDocumentLinksScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != captureDocumentLinksLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, captureDocumentLinksLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: sourceURLUpdateDuplicateScenarioID + "," + sourceURLUpdateSameSHAScenarioID + "," + sourceURLUpdateChangedScenarioID + "," + sourceURLUpdateConflictScenarioID})
	lane, releaseBlocking = reportLane(selected)
	if lane != sourceURLUpdateLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, sourceURLUpdateLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(webProductPageScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != webProductPageLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, webProductPageLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: documentThisMissingFieldsScenarioID + "," + documentThisExplicitCreateScenarioID + "," + documentThisSourceURLMissingHintsScenarioID + "," + documentThisExplicitOverridesScenarioID + "," + documentThisDuplicateCandidateScenarioID + "," + documentThisExistingUpdateScenarioID + "," + documentThisSynthesisFreshnessScenarioID})
	lane, releaseBlocking = reportLane(selected)
	if lane != documentThisLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, documentThisLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(documentArtifactCandidateScenarioIDs(), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != documentArtifactCandidateLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, documentArtifactCandidateLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(artifactIngestionScenarioIDs(), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != artifactIngestionLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, artifactIngestionLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(videoYouTubeScenarioIDs(), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != videoYouTubeLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, videoYouTubeLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(webURLStaleRepairScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != webURLStaleRepairLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, webURLStaleRepairLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(webURLStaleImpactScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != webURLStaleImpactLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, webURLStaleImpactLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(synthesisCompileScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != synthesisCompileLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, synthesisCompileLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(highTouchCompileSynthesisScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != highTouchCompileSynthesisLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, highTouchCompileSynthesisLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(graphSemanticsRevisitScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != graphSemanticsRevisitLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, graphSemanticsRevisitLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(memoryRouterRevisitScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != memoryRouterRevisitLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, memoryRouterRevisitLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(promotedRecordDomainScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != promotedRecordDomainLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, promotedRecordDomainLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(parallelRunnerScenarioIDs(), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != parallelRunnerLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, parallelRunnerLaneName)
	}
	selected = selectedScenarioIDs(runConfig{Scenario: strings.Join(append(broadAuditScenarioIDs(), "missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"), ",")})
	lane, releaseBlocking = reportLane(selected)
	if lane != broadAuditLaneName || releaseBlocking {
		t.Fatalf("reportLane(%v) = %q/%t, want %q/false", selected, lane, releaseBlocking, broadAuditLaneName)
	}
}

func requireScenarioByID(t *testing.T, id string) scenario {
	t.Helper()
	for _, sc := range allScenarios() {
		if sc.ID == id {
			return sc
		}
	}
	t.Fatalf("missing scenario %q", id)
	return scenario{}
}

func containsValue(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}
