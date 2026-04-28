package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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
	if lane != populatedLaneName && lane != repoDocsLaneName && lane != documentHistoryLaneName && lane != agentChosenPathLaneName && lane != pathTitleAutonomyLaneName && lane != sourceURLUpdateLaneName && lane != documentThisLaneName && lane != documentArtifactCandidateLaneName && lane != artifactIngestionLaneName && lane != videoYouTubeLaneName && lane != synthesisCompileLaneName {
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
		case documentHistoryLaneName:
			include = isDocumentHistoryScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedDocumentHistoryResult(result)
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
		case synthesisCompileLaneName:
			include = isSynthesisCompileScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedSynthesisCompileResult(result)
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
	case documentHistoryLaneName:
		summary.Decision = documentHistoryDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted document lifecycle evidence only; no promoted history, diff, review, restore, rollback, schema, migration, storage behavior, or public API change from this eval"
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
	case synthesisCompileLaneName:
		summary.Decision = synthesisCompileDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted evidence only; no compile_synthesis runner action, schema, migration, storage behavior, direct vault behavior, or public API change from this eval"
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
func classifyTargetedSynthesisCompileResult(result jobResult) (string, string) {
	if isFinalAnswerOnlyValidationScenario(result.Scenario) {
		if result.Passed && result.Verification.Passed {
			return "none", "validation control stayed final-answer-only"
		}
		if result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1 {
			return "skill_guidance_or_eval_coverage", "validation pressure did not stay final-answer-only"
		}
		return "skill_guidance_or_eval_coverage", "validation answer did not satisfy the rejection contract"
	}
	if result.Passed && result.Verification.Passed {
		return "none", "current document/retrieval workflow preserved synthesis authority, source refs, provenance/freshness checks, duplicate prevention, and bypass boundaries"
	}
	if len(populatedBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if result.Verification.Passed {
		return "eval_contract_violation", "scenario verification passed, but the job did not complete successfully"
	}
	if result.Scenario == synthesisCompileScriptedScenarioID && !result.Verification.DatabasePass {
		return "capability_gap", "scripted current-primitives control could not safely repair source-linked synthesis"
	}
	if result.Scenario == synthesisCompileNaturalScenarioID && !result.Verification.Passed {
		return "ergonomics_gap", "natural compile_synthesis revisit intent did not complete the safe current-primitives workflow"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene_or_fixture_gap", "fixture or durable synthesis evidence did not satisfy compile_synthesis revisit pressure"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance_or_eval_coverage", "runner-visible synthesis evidence existed, but the assistant answer or required runner steps did not satisfy the scenario"
	}
	return "ergonomics_gap", "manual review required before any compile_synthesis promotion"
}
func promptSpecificity(scenarioID string) string {
	switch scenarioID {
	case documentHistoryNaturalScenarioID:
		return "natural-user-intent"
	case documentHistoryInspectScenarioID, documentHistoryDiffScenarioID, documentHistoryRestoreScenarioID, documentHistoryPendingScenarioID, documentHistoryStaleScenarioID:
		return "scripted-control"
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
	case synthesisCompileNaturalScenarioID:
		return "natural-user-intent"
	case synthesisCompileScriptedScenarioID:
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
	if result.Scenario == synthesisCompileNaturalScenarioID && !result.Passed {
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
	case documentHistoryNaturalScenarioID:
		if result.Passed {
			return "low_natural_user_intent"
		}
		return "high_if_natural_prompt_failed"
	case documentHistoryInspectScenarioID, documentHistoryDiffScenarioID, documentHistoryRestoreScenarioID, documentHistoryPendingScenarioID, documentHistoryStaleScenarioID:
		return "high_exact_runner_workflow"
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
	case synthesisCompileNaturalScenarioID:
		if result.Passed {
			return "low_natural_user_intent"
		}
		return "high_if_natural_prompt_failed"
	case synthesisCompileScriptedScenarioID:
		return "high_exact_request_shape"
	default:
		return "scenario_prompt"
	}
}
func scenarioSafetyRisks(result jobResult) string {
	if isSynthesisCompileScenario(result.Scenario) && result.Metrics.CreateDocumentUsed {
		return "duplicate_or_unexpected_create"
	}
	if result.Metrics.CreateDocumentUsed && result.Scenario != videoYouTubeScriptedTranscriptControlID && result.Scenario != documentHistoryPendingScenarioID {
		return "wrote_before_approval"
	}
	if isDocumentHistoryScenario(result.Scenario) && len(documentHistoryInvariantFailures(result.Metrics)) != 0 {
		return "bypass_or_private_artifact_risk"
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
func classifyTargetedDocumentHistoryResult(result jobResult) (string, string) {
	if isFinalAnswerOnlyValidationScenario(result.Scenario) {
		if result.Passed && result.Verification.Passed {
			return "none", "validation pressure stayed final-answer-only without bypassing the installed runner contract"
		}
		if result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1 {
			return "skill_guidance", "validation pressure did not stay final-answer-only"
		}
		return "skill_guidance", "validation answer did not satisfy the document lifecycle no-tools contract"
	}
	if result.Passed && result.Verification.Passed {
		if result.Scenario == documentHistoryNaturalScenarioID {
			return "none", "natural document lifecycle intent completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries"
		}
		return "none", "scripted document lifecycle control completed through existing document/retrieval runner evidence while preserving provenance, freshness, privacy, and bypass boundaries"
	}
	if len(documentHistoryInvariantFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if result.Verification.Passed {
		return "eval_contract_violation", "scenario verification passed, but the job did not complete successfully"
	}
	if result.Scenario == documentHistoryNaturalScenarioID {
		return "ergonomics_gap", "natural document lifecycle intent did not complete the safe current-primitives workflow"
	}
	if !result.Verification.DatabasePass {
		return "data_hygiene", "fixture or durable evidence did not satisfy document lifecycle pressure"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance", "runner-visible evidence existed, but the assistant answer did not satisfy document lifecycle pressure"
	}
	return "runner_capability_gap", "manual review required before any document lifecycle surface promotion"
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
	synthesisCompile := 0
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
		if isSynthesisCompileScenario(id) {
			synthesisCompile++
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
	if synthesisCompile > 0 && synthesisCompile+validation == len(ids) {
		return synthesisCompileLaneName, false
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
	if synthesisCompile > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	return populatedDefaultLaneName, true
}
func targetedAcceptanceNote(lane string) string {
	if lane == repoDocsLaneName {
		return "repo-docs dogfood rows import committed public markdown into an isolated eval vault and report retrieval, synthesis, and decision-record behavior without private vault evidence"
	}
	if lane == documentHistoryLaneName {
		return "document lifecycle rows report natural intent, scripted current-primitives controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, privacy handling, and capability/ergonomics classification"
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
	if lane == synthesisCompileLaneName {
		return "synthesis compile revisit rows report natural compile intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	return ""
}
