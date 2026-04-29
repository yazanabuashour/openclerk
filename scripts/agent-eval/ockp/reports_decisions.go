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
