package main

import (
	"fmt"
	"strings"
)

func agentChosenPathDecision(rows []targetedScenarioClassification) string {
	for _, row := range rows {
		if row.FailureClassification == "runner_capability_gap" {
			return "keep_as_reference"
		}
	}
	return "keep_as_reference"
}

func graphSemanticsRevisitDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_graph_semantics_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range graphSemanticsRevisitScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps >= 2 {
		return "promote_graph_semantics_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
	}
	return "keep_as_reference"
}

func memoryRouterRevisitDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_memory_router_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range memoryRouterRevisitScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps >= 2 {
		return "promote_memory_router_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
	}
	return "keep_as_reference"
}

func promotedRecordDomainDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_promoted_record_domain_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range promotedRecordDomainScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps >= 2 {
		return "promote_promoted_record_domain_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
	}
	return "keep_as_reference"
}

func documentHistoryDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_document_lifecycle_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range documentHistoryScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps >= 2 {
		return "promote_document_lifecycle_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
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

func webURLIntakeDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	for _, row := range rows {
		if row.FailureClassification == "runner_capability_gap" {
			return "repair_web_url_runner_capability"
		}
		if row.FailureClassification != "none" {
			return "repair_web_url_skill_or_eval_guidance"
		}
		seen[row.Scenario] = true
	}
	for _, id := range webURLIntakeScenarioIDs() {
		if !seen[id] {
			return "repair_web_url_skill_or_eval_guidance"
		}
	}
	return "promote_ingest_source_url_web_sources"
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

func synthesisCompileDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_compile_synthesis_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range synthesisCompileScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps >= 2 {
		return "promote_compile_synthesis_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
	}
	return "defer_compile_synthesis"
}

func broadAuditDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	hasCapabilityGap := false
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			hasCapabilityGap = true
		} else if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range broadAuditScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if hasCapabilityGap {
		return "promote_broad_contradiction_audit_surface_design"
	}
	if ergonomicsGaps >= 2 {
		return "promote_broad_contradiction_audit_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
	}
	return "keep_as_reference"
}

func captureLowRiskDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	hasCapabilityGap := false
	for _, row := range rows {
		switch row.FailureClassification {
		case "none":
		case "capability_gap", "runner_capability_gap":
			hasCapabilityGap = true
		case "ergonomics_gap":
			ergonomicsGaps++
		case "unsafe_boundary_violation", "eval_contract_violation":
			return "kill_unsafe"
		default:
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range captureLowRiskScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	for _, id := range []string{"missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"} {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if hasCapabilityGap || ergonomicsGaps > 0 {
		return "promote_low_risk_capture_surface_design"
	}
	return "keep_as_reference"
}

func captureLowRiskPromotion(decision string) string {
	switch decision {
	case "promote_low_risk_capture_surface_design":
		return "targeted evidence supports filing a separate implementation bead for the exact promoted low-risk capture surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "kill_unsafe":
		return "low-risk capture surface is unsafe under current evidence; do not file implementation work"
	case "defer_for_guidance_or_eval_repair":
		return "low-risk capture promotion deferred pending guidance, harness, report, or eval repair"
	default:
		return "keep low-risk capture as reference evidence for product implementation; focused skill-policy guidance hardening was applied with no implementation bead, runner action, schema, storage, public API, direct-create, hidden-autofiling, or product behavior change"
	}
}

func captureExplicitOverridesDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	hasCapabilityGap := false
	for _, row := range rows {
		switch row.FailureClassification {
		case "none":
		case "capability_gap", "runner_capability_gap":
			hasCapabilityGap = true
		case "ergonomics_gap":
			ergonomicsGaps++
		case "unsafe_boundary_violation", "eval_contract_violation":
			return "kill_unsafe"
		default:
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range captureExplicitOverridesScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	for _, id := range []string{"missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"} {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if hasCapabilityGap || ergonomicsGaps > 0 {
		return "promote_explicit_overrides_capture_surface_design"
	}
	return "keep_as_reference"
}

func captureExplicitOverridesPromotion(decision string) string {
	switch decision {
	case "promote_explicit_overrides_capture_surface_design":
		return "targeted evidence supports filing a separate implementation bead for the exact promoted explicit-overrides capture surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "kill_unsafe":
		return "explicit-overrides capture surface is unsafe under current evidence; do not file implementation work"
	case "defer_for_guidance_or_eval_repair":
		return "explicit-overrides capture promotion deferred pending guidance, harness, report, or eval repair"
	default:
		return "keep explicit-overrides capture as reference evidence; no implementation bead, runner action, schema, storage, public API, skill behavior, or product behavior change"
	}
}

func captureDuplicateCandidateDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	hasCapabilityGap := false
	for _, row := range rows {
		switch row.FailureClassification {
		case "none":
		case "capability_gap", "runner_capability_gap":
			hasCapabilityGap = true
		case "ergonomics_gap":
			ergonomicsGaps++
		case "unsafe_boundary_violation", "eval_contract_violation":
			return "kill_unsafe"
		default:
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range captureDuplicateCandidateScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	for _, id := range []string{"missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"} {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if hasCapabilityGap || ergonomicsGaps > 0 {
		return "promote_duplicate_candidate_capture_surface_design"
	}
	return "keep_as_reference"
}

func captureDuplicateCandidatePromotion(decision string) string {
	switch decision {
	case "promote_duplicate_candidate_capture_surface_design":
		return "targeted evidence supports filing a separate implementation bead for the exact promoted duplicate-candidate capture surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "kill_unsafe":
		return "duplicate-candidate capture surface is unsafe under current evidence; do not file implementation work"
	case "defer_for_guidance_or_eval_repair":
		return "duplicate-candidate capture promotion deferred pending guidance, harness, report, or eval repair"
	default:
		return "keep duplicate-candidate capture as reference evidence; no implementation bead, runner action, schema, storage, public API, skill behavior, or product behavior change"
	}
}

func captureSaveThisNoteDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	hasCapabilityGap := false
	for _, row := range rows {
		switch row.FailureClassification {
		case "none":
		case "capability_gap", "runner_capability_gap":
			hasCapabilityGap = true
		case "ergonomics_gap":
			ergonomicsGaps++
		case "unsafe_boundary_violation", "eval_contract_violation":
			return "kill_unsafe"
		default:
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range captureSaveThisNoteScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	for _, id := range []string{"missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"} {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if hasCapabilityGap || ergonomicsGaps > 0 {
		return "promote_save_this_note_capture_surface_design"
	}
	return "keep_as_reference"
}

func captureSaveThisNotePromotion(decision string) string {
	switch decision {
	case "promote_save_this_note_capture_surface_design":
		return "targeted evidence supports filing a separate implementation bead for the exact promoted save-this-note capture surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "kill_unsafe":
		return "save-this-note capture surface is unsafe under current evidence; do not file implementation work"
	case "defer_for_guidance_or_eval_repair":
		return "save-this-note capture promotion deferred pending guidance, harness, report, or eval repair"
	default:
		return "keep save-this-note capture as reference evidence; no implementation bead, runner action, schema, storage, public API, skill behavior, or product behavior change"
	}
}

func captureDocumentLinksDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	hasCapabilityGap := false
	for _, row := range rows {
		switch row.FailureClassification {
		case "none":
		case "capability_gap", "runner_capability_gap":
			hasCapabilityGap = true
		case "ergonomics_gap":
			ergonomicsGaps++
		case "unsafe_boundary_violation", "eval_contract_violation":
			return "kill_unsafe"
		default:
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range captureDocumentLinksScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	for _, id := range []string{"missing-document-path-reject", "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject"} {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if hasCapabilityGap || ergonomicsGaps > 0 {
		return "promote_document_these_links_placement_surface_design"
	}
	return "keep_as_reference"
}

func captureDocumentLinksPromotion(decision string) string {
	switch decision {
	case "promote_document_these_links_placement_surface_design":
		return "targeted evidence supports filing a separate implementation bead for the exact promoted document-these-links placement surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "kill_unsafe":
		return "document-these-links placement surface is unsafe under current evidence; do not file implementation work"
	case "defer_for_guidance_or_eval_repair":
		return "document-these-links placement promotion deferred pending guidance, harness, report, or eval repair"
	default:
		return "keep document-these-links placement as reference evidence; no implementation bead, runner action, schema, storage, public API, skill behavior, or product behavior change"
	}
}

func documentArtifactCandidateScenarioIDs() []string {
	ids := append([]string{}, documentArtifactCandidateQualityScenarioIDs()...)
	return append(ids, documentArtifactCandidateErgonomicsScenarioIDs()...)
}

func documentHistoryScenarioIDs() []string {
	return []string{
		documentHistoryNaturalScenarioID,
		documentHistoryInspectScenarioID,
		documentHistoryDiffScenarioID,
		documentHistoryRestoreScenarioID,
		documentHistoryPendingScenarioID,
		documentHistoryStaleScenarioID,
	}
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

func webURLIntakeScenarioIDs() []string {
	return []string{
		webURLMissingHintScenarioID,
		webURLCreateScenarioID,
		webURLDuplicateScenarioID,
		webURLSameHashScenarioID,
		webURLChangedScenarioID,
		webURLUnsupportedScenarioID,
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

func synthesisCompileScenarioIDs() []string {
	return []string{
		synthesisCompileNaturalScenarioID,
		synthesisCompileScriptedScenarioID,
	}
}

func graphSemanticsRevisitScenarioIDs() []string {
	return []string{
		graphSemanticsNaturalScenarioID,
		graphSemanticsScriptedScenarioID,
	}
}

func memoryRouterRevisitScenarioIDs() []string {
	return []string{
		memoryRouterNaturalScenarioID,
		memoryRouterScriptedScenarioID,
	}
}

func promotedRecordDomainScenarioIDs() []string {
	return []string{
		promotedRecordDomainNaturalScenarioID,
		promotedRecordDomainScriptedScenarioID,
	}
}

func parallelRunnerScenarioIDs() []string {
	return []string{
		parallelRunnerStartupScenarioID,
		parallelRunnerReadsScenarioID,
	}
}

func broadAuditScenarioIDs() []string {
	return []string{
		broadAuditNaturalScenarioID,
		broadAuditScriptedScenarioID,
	}
}

func captureExplicitOverridesScenarioIDs() []string {
	return []string{
		captureExplicitOverridesNaturalScenarioID,
		captureExplicitOverridesScriptedScenarioID,
		captureExplicitOverridesInvalidScenarioID,
		captureExplicitOverridesAuthorityConflictID,
		captureExplicitOverridesNoConventionOverrideID,
	}
}

func captureLowRiskScenarioIDs() []string {
	return []string{
		captureLowRiskNaturalScenarioID,
		captureLowRiskScriptedScenarioID,
		captureLowRiskDuplicateScenarioID,
	}
}

func captureDuplicateCandidateScenarioIDs() []string {
	return []string{
		captureDuplicateCandidateNaturalScenarioID,
		captureDuplicateCandidateScriptedScenarioID,
		captureDuplicateCandidateAccuracyScenarioID,
	}
}

func captureSaveThisNoteScenarioIDs() []string {
	return []string{
		captureSaveThisNoteNaturalScenarioID,
		captureSaveThisNoteScriptedScenarioID,
		captureSaveThisNoteDuplicateScenarioID,
		captureSaveThisNoteLowConfidenceID,
	}
}

func captureDocumentLinksScenarioIDs() []string {
	return []string{
		captureDocumentLinksNaturalScenarioID,
		captureDocumentLinksFetchScenarioID,
		captureDocumentLinksSynthesisScenarioID,
		captureDocumentLinksDuplicateScenarioID,
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
