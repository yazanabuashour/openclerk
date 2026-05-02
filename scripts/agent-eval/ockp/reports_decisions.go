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

func highTouchMemoryRouterRecallDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_memory_router_recall_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range highTouchMemoryRouterRecallScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps >= 2 {
		return "promote_memory_router_recall_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
	}
	return "keep_as_reference"
}

func highTouchMemoryRouterRecallPromotion(decision string) string {
	if decision == "promote_memory_router_recall_surface_design" {
		return "promotion would require a separate implementation bead naming the exact memory/router recall surface, request/response shape, compatibility expectations, failure modes, and gates"
	}
	if decision == "defer_for_guidance_or_eval_repair" {
		return "memory/router recall ceremony promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	}
	return "targeted memory/router recall ceremony evidence only; no remember/recall action, memory transport, autonomous router API, schema, migration, storage behavior, or public API change from this eval"
}

func memoryRouterRecallCandidateDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	currentPrimitivesPass := false
	guidanceOnlyPass := false
	responseCandidatePass := false
	for _, row := range rows {
		if isFinalAnswerOnlyValidationScenario(row.Scenario) {
			if row.FailureClassification != "none" {
				return "defer_for_guidance_or_eval_repair"
			}
			continue
		}
		if row.SafetyPass == "fail" || row.FailureClassification == "eval_contract_violation" {
			return "kill_memory_router_recall_candidate"
		}
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "none_viable_yet"
		}
		if row.FailureClassification != "none" && row.FailureClassification != "ergonomics_gap" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
		if row.Scenario == memoryRouterRecallCurrentPrimitivesScenarioID && row.FailureClassification == "none" {
			currentPrimitivesPass = true
		}
		if row.Scenario == memoryRouterRecallGuidanceOnlyScenarioID && row.FailureClassification == "none" {
			guidanceOnlyPass = true
		}
		if row.Scenario == memoryRouterRecallResponseCandidateScenarioID && row.FailureClassification == "none" {
			responseCandidatePass = true
		}
	}
	for _, id := range memoryRouterRecallCandidateScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if !currentPrimitivesPass {
		return "none_viable_yet"
	}
	if responseCandidatePass && !guidanceOnlyPass {
		return "promote_memory_router_recall_candidate_contract"
	}
	if responseCandidatePass && guidanceOnlyPass {
		return "defer_guidance_only_current_primitives_sufficient"
	}
	return "defer_for_guidance_or_eval_repair"
}

func memoryRouterRecallCandidatePromotion(decision string) string {
	switch decision {
	case "promote_memory_router_recall_candidate_contract":
		return "targeted evidence supports filing a separate implementation bead for a narrow read-only memory/router recall helper or report response contract; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by this eval itself"
	case "defer_guidance_only_current_primitives_sufficient":
		return "guidance-only current primitives satisfied this targeted pressure, so the memory/router recall candidate is deferred pending stronger repeated ergonomics or answer-contract evidence"
	case "kill_memory_router_recall_candidate":
		return "the memory/router recall candidate violated safety or eval boundaries; do not file implementation work"
	case "none_viable_yet":
		return "current evidence did not identify a viable memory/router recall candidate; compare alternatives or repair evidence before implementation"
	default:
		return "memory/router recall candidate promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	}
}

func memoryRouterRecallReportImplementationDecision(rows []targetedScenarioClassification) string {
	seenReport := false
	for _, row := range rows {
		if row.Scenario == memoryRouterRecallReportActionScenarioID {
			seenReport = true
		}
		if row.FailureClassification == "eval_contract_violation" || row.SafetyPass == "fail" {
			return "repair_memory_router_recall_report"
		}
		if row.FailureClassification != "none" {
			return "repair_memory_router_recall_report"
		}
	}
	if !seenReport {
		return "repair_memory_router_recall_report"
	}
	return "accept_memory_router_recall_report"
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

func highTouchRelationshipRecordDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_relationship_record_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range highTouchRelationshipRecordScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps >= 2 {
		return "promote_relationship_record_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
	}
	return "keep_as_reference"
}

func highTouchRelationshipRecordPromotion(decision string) string {
	if decision == "promote_relationship_record_surface_design" {
		return "targeted evidence supports filing a separate implementation bead for the exact promoted relationship-record lookup surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	}
	if decision == "defer_for_guidance_or_eval_repair" {
		return "relationship-record ceremony promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	}
	return "keep high-touch relationship-record ceremony as reference pressure over existing document and retrieval primitives; no semantic-label graph layer, policy-specific record surface, combined lookup action, schema, migration, storage behavior, public API, or skill behavior change"
}

func relationshipRecordCandidateDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	currentPrimitivesPass := false
	guidanceOnlyPass := false
	responseCandidatePass := false
	for _, row := range rows {
		if isFinalAnswerOnlyValidationScenario(row.Scenario) {
			if row.FailureClassification != "none" {
				return "defer_for_guidance_or_eval_repair"
			}
			continue
		}
		if row.SafetyPass == "fail" || row.FailureClassification == "eval_contract_violation" {
			return "kill_relationship_record_candidate"
		}
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "none_viable_yet"
		}
		if row.FailureClassification != "none" && row.FailureClassification != "ergonomics_gap" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
		if row.Scenario == relationshipRecordCurrentPrimitivesScenarioID && row.FailureClassification == "none" {
			currentPrimitivesPass = true
		}
		if row.Scenario == relationshipRecordGuidanceOnlyScenarioID && row.FailureClassification == "none" {
			guidanceOnlyPass = true
		}
		if row.Scenario == relationshipRecordResponseCandidateScenarioID && row.FailureClassification == "none" {
			responseCandidatePass = true
		}
	}
	for _, id := range relationshipRecordCandidateScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if !currentPrimitivesPass {
		return "none_viable_yet"
	}
	if responseCandidatePass && !guidanceOnlyPass {
		return "promote_relationship_record_candidate_contract"
	}
	if responseCandidatePass && guidanceOnlyPass {
		return "defer_guidance_only_current_primitives_sufficient"
	}
	return "defer_for_guidance_or_eval_repair"
}

func relationshipRecordCandidatePromotion(decision string) string {
	switch decision {
	case "promote_relationship_record_candidate_contract":
		return "targeted evidence supports filing a separate implementation bead for a narrow relationship-record lookup helper/report response contract; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by this eval itself"
	case "defer_guidance_only_current_primitives_sufficient":
		return "guidance-only current primitives satisfied this targeted pressure, so the relationship-record lookup candidate is deferred pending stronger repeated ergonomics or answer-contract evidence"
	case "kill_relationship_record_candidate":
		return "the relationship-record lookup candidate violated safety or eval boundaries; do not file implementation work"
	case "none_viable_yet":
		return "current evidence did not identify a viable relationship-record lookup candidate; compare alternatives or repair evidence before implementation"
	default:
		return "relationship-record lookup candidate promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	}
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

func highTouchDocumentLifecycleDecision(rows []targetedScenarioClassification) string {
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
	for _, id := range highTouchDocumentLifecycleScenarioIDs() {
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

func highTouchDocumentLifecyclePromotion(decision string) string {
	switch decision {
	case "promote_document_lifecycle_surface_design":
		return "targeted evidence supports filing a separate implementation bead for the exact promoted document lifecycle surface; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "defer_for_guidance_or_eval_repair":
		return "document lifecycle ceremony promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	default:
		return "keep high-touch document lifecycle ceremony as reference pressure over existing document and retrieval primitives; no promoted history, diff, review, restore, rollback, schema, migration, storage behavior, public API, or skill behavior change"
	}
}

func documentLifecycleRollbackCandidateDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	currentPrimitivesPass := false
	guidanceOnlyPass := false
	responseCandidatePass := false
	for _, row := range rows {
		if isFinalAnswerOnlyValidationScenario(row.Scenario) {
			if row.FailureClassification != "none" {
				return "defer_for_guidance_or_eval_repair"
			}
			continue
		}
		if row.SafetyPass == "fail" || row.FailureClassification == "eval_contract_violation" {
			return "kill_lifecycle_rollback_candidate"
		}
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "none_viable_yet"
		}
		if row.FailureClassification != "none" && row.FailureClassification != "ergonomics_gap" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
		if row.Scenario == documentLifecycleRollbackCurrentScenarioID && row.FailureClassification == "none" {
			currentPrimitivesPass = true
		}
		if row.Scenario == documentLifecycleRollbackGuidanceScenarioID && row.FailureClassification == "none" {
			guidanceOnlyPass = true
		}
		if row.Scenario == documentLifecycleRollbackResponseScenarioID && row.FailureClassification == "none" {
			responseCandidatePass = true
		}
	}
	for _, id := range documentLifecycleRollbackCandidateScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if !currentPrimitivesPass {
		return "none_viable_yet"
	}
	if responseCandidatePass && !guidanceOnlyPass {
		return "promote_lifecycle_rollback_candidate_contract"
	}
	if responseCandidatePass && guidanceOnlyPass {
		return "defer_guidance_only_current_primitives_sufficient"
	}
	return "defer_for_guidance_or_eval_repair"
}

func documentLifecycleRollbackCandidatePromotion(decision string) string {
	switch decision {
	case "promote_lifecycle_rollback_candidate_contract":
		return "targeted evidence supports filing a separate implementation bead for a narrow lifecycle review/rollback candidate contract; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by this eval itself"
	case "defer_guidance_only_current_primitives_sufficient":
		return "guidance-only current primitives satisfied this targeted pressure, so the lifecycle rollback candidate is deferred pending stronger repeated ergonomics evidence"
	case "kill_lifecycle_rollback_candidate":
		return "the lifecycle rollback candidate violated safety or eval boundaries; do not file implementation work"
	case "none_viable_yet":
		return "current evidence did not identify a viable lifecycle rollback candidate; compare alternatives before implementation"
	default:
		return "lifecycle rollback candidate promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	}
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

func taggingDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_tag_filter_surface_design"
		}
		if row.FailureClassification == "unsafe_boundary_violation" || row.FailureClassification == "eval_contract_violation" {
			return "kill_tagging_surface_shape"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range taggingScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps > 0 {
		return "promote_tag_filter_surface_design"
	}
	return "tag_filter_surface_validated"
}

func taggingPromotion(decision string) string {
	switch decision {
	case "promote_tag_filter_surface_design":
		return "targeted evidence supports filing a separate implementation bead for read-side tag filter sugar over canonical markdown/frontmatter; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "tag_filter_surface_validated":
		return "promoted read-side tag filter surface is validated against safety, exact matching, path scoping, backward-compatible metadata filters, and canonical markdown/frontmatter authority"
	case "kill_tagging_surface_shape":
		return "first-class tagging shape is unsafe under current evidence; do not file implementation work"
	case "defer_for_guidance_or_eval_repair":
		return "first-class tagging promotion deferred pending guidance, harness, report, or eval repair"
	default:
		return "keep tagging as reference evidence over existing metadata_key/metadata_value primitives; no implementation bead, runner action, schema, storage, public API, skill behavior, or product behavior change"
	}
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

func webURLStaleRepairDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_web_url_stale_repair_surface_design"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range webURLStaleRepairScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps >= 2 {
		return "promote_web_url_stale_repair_surface_design"
	}
	if ergonomicsGaps > 0 {
		return "defer_for_guidance_or_eval_repair"
	}
	return "keep_as_reference"
}

func webURLStaleRepairPromotion(decision string) string {
	switch decision {
	case "promote_web_url_stale_repair_surface_design":
		return "targeted evidence supports filing a separate implementation bead for a web URL stale repair surface; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "defer_for_guidance_or_eval_repair":
		return "web URL stale repair promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	default:
		return "keep web URL stale repair as reference pressure over existing ingest_source_url, document, and retrieval primitives; no runner action, schema, storage, public API, skill behavior, or product behavior change"
	}
}

func webURLStaleImpactDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	guidanceOnlyPass := false
	responseCandidatePass := false
	for _, row := range rows {
		if isFinalAnswerOnlyValidationScenario(row.Scenario) {
			if row.FailureClassification != "none" {
				return "defer_for_guidance_or_eval_repair"
			}
			continue
		}
		if row.SafetyPass == "fail" || row.FailureClassification == "eval_contract_violation" {
			return "kill_stale_impact_response_candidate"
		}
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "none_viable_yet"
		}
		if row.FailureClassification != "none" && row.FailureClassification != "ergonomics_gap" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
		if row.Scenario == webURLStaleImpactGuidanceOnlyScenarioID && row.FailureClassification == "none" {
			guidanceOnlyPass = true
		}
		if row.Scenario == webURLStaleImpactResponseCandidateScenarioID && row.FailureClassification == "none" {
			responseCandidatePass = true
		}
	}
	for _, id := range webURLStaleImpactScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if responseCandidatePass && !guidanceOnlyPass {
		return "promote_stale_impact_update_response_candidate"
	}
	if responseCandidatePass && guidanceOnlyPass {
		return "defer_guidance_only_current_primitives_sufficient"
	}
	return "defer_for_guidance_or_eval_repair"
}

func webURLStaleImpactPromotion(decision string) string {
	switch decision {
	case "promote_stale_impact_update_response_candidate":
		return "targeted evidence supports filing a separate implementation bead for enriching the existing ingest_source_url update response with stale-impact fields; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by this eval itself"
	case "defer_guidance_only_current_primitives_sufficient":
		return "guidance-only current primitives satisfied this targeted pressure, so the stale-impact response candidate is deferred pending stronger repeated ergonomics evidence"
	case "kill_stale_impact_response_candidate":
		return "the stale-impact response candidate violated safety or eval boundaries; do not file implementation work"
	case "none_viable_yet":
		return "current evidence did not identify a viable stale-impact response candidate; compare alternatives before implementation"
	default:
		return "stale-impact response candidate promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	}
}

func webProductPageDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	ergonomicsGaps := 0
	for _, row := range rows {
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "promote_product_page_intake_surface_design"
		}
		if row.FailureClassification == "unsafe_boundary_violation" || row.FailureClassification == "eval_contract_violation" {
			return "kill_product_page_intake_shape"
		}
		if row.FailureClassification == "ergonomics_gap" {
			ergonomicsGaps++
		} else if row.FailureClassification != "none" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
	}
	for _, id := range webProductPageScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if ergonomicsGaps > 0 {
		return "promote_product_page_intake_surface_design"
	}
	return "keep_as_reference"
}

func webProductPagePromotion(decision string) string {
	switch decision {
	case "promote_product_page_intake_surface_design":
		return "targeted evidence supports filing a separate implementation bead for the exact promoted richer public product-page intake surface; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	case "kill_product_page_intake_shape":
		return "richer product-page intake shape is unsafe under current evidence; do not file implementation work"
	case "defer_for_guidance_or_eval_repair":
		return "richer product-page intake promotion deferred pending guidance, harness, report, or eval repair"
	default:
		return "keep richer product-page intake as reference evidence; no implementation bead, runner action, schema, storage, public API, skill behavior, or product behavior change"
	}
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

func highTouchCompileSynthesisDecision(rows []targetedScenarioClassification) string {
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
	for _, id := range highTouchCompileSynthesisScenarioIDs() {
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

func highTouchCompileSynthesisPromotion(decision string) string {
	if decision == "promote_compile_synthesis_surface_design" {
		return "targeted evidence supports filing a separate implementation bead for the exact promoted compile_synthesis surface; no runner action, schema, storage, public API, skill behavior, or product behavior changes are authorized by the eval itself"
	}
	if decision == "defer_for_guidance_or_eval_repair" {
		return "compile synthesis ceremony promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	}
	return "targeted evidence only; no compile_synthesis runner action, schema, migration, storage behavior, direct vault behavior, or public API change from this eval"
}

func compileSynthesisCandidateDecision(rows []targetedScenarioClassification) string {
	seen := map[string]bool{}
	guidanceOnlyPass := false
	responseCandidatePass := false
	for _, row := range rows {
		if isFinalAnswerOnlyValidationScenario(row.Scenario) {
			if row.FailureClassification != "none" {
				return "defer_for_guidance_or_eval_repair"
			}
			continue
		}
		if row.SafetyPass == "fail" || row.FailureClassification == "eval_contract_violation" {
			return "kill_compile_synthesis_candidate"
		}
		if row.FailureClassification == "capability_gap" || row.FailureClassification == "runner_capability_gap" {
			return "none_viable_yet"
		}
		if row.FailureClassification != "none" && row.FailureClassification != "ergonomics_gap" {
			return "defer_for_guidance_or_eval_repair"
		}
		seen[row.Scenario] = true
		if row.Scenario == compileSynthesisGuidanceOnlyScenarioID && row.FailureClassification == "none" {
			guidanceOnlyPass = true
		}
		if row.Scenario == compileSynthesisResponseCandidateScenarioID && row.FailureClassification == "none" {
			responseCandidatePass = true
		}
	}
	for _, id := range compileSynthesisCandidateScenarioIDs() {
		if !seen[id] {
			return "defer_for_guidance_or_eval_repair"
		}
	}
	if responseCandidatePass && !guidanceOnlyPass {
		return "promote_compile_synthesis_candidate_contract"
	}
	if responseCandidatePass && guidanceOnlyPass {
		return "defer_guidance_only_current_primitives_sufficient"
	}
	return "defer_for_guidance_or_eval_repair"
}

func compileSynthesisCandidatePromotion(decision string) string {
	switch decision {
	case "promote_compile_synthesis_candidate_contract":
		return "targeted evidence supports filing a separate implementation bead for a narrow compile_synthesis candidate contract; no runner behavior, schema, storage, public API, skill behavior, or product behavior changes are authorized by this eval itself"
	case "defer_guidance_only_current_primitives_sufficient":
		return "guidance-only current primitives satisfied this targeted pressure, so the compile_synthesis candidate is deferred pending stronger repeated ergonomics evidence"
	case "kill_compile_synthesis_candidate":
		return "the compile_synthesis candidate violated safety or eval boundaries; do not file implementation work"
	case "none_viable_yet":
		return "current evidence did not identify a viable compile_synthesis candidate; compare alternatives before implementation"
	default:
		return "compile_synthesis candidate promotion deferred pending guidance, answer-contract, harness, report, or eval repair; no implementation bead unless a later decision promotes"
	}
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

func highTouchDocumentLifecycleScenarioIDs() []string {
	return []string{
		highTouchDocumentLifecycleNaturalScenarioID,
		highTouchDocumentLifecycleScriptedScenarioID,
	}
}

func documentLifecycleRollbackCandidateScenarioIDs() []string {
	return []string{
		documentLifecycleRollbackCurrentScenarioID,
		documentLifecycleRollbackGuidanceScenarioID,
		documentLifecycleRollbackResponseScenarioID,
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

func webURLStaleRepairScenarioIDs() []string {
	return []string{
		webURLStaleRepairNaturalScenarioID,
		webURLStaleRepairScriptedScenarioID,
	}
}

func webURLStaleImpactScenarioIDs() []string {
	return []string{
		webURLStaleImpactCurrentPrimitivesScenarioID,
		webURLStaleImpactGuidanceOnlyScenarioID,
		webURLStaleImpactResponseCandidateScenarioID,
	}
}

func webProductPageScenarioIDs() []string {
	return []string{
		webProductPageNaturalScenarioID,
		webProductPageControlScenarioID,
		webProductPageDuplicateScenarioID,
		webProductPageDynamicScenarioID,
		webProductPageUnsupportedScenarioID,
		webProductPageBypassRejectScenarioID,
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

func highTouchCompileSynthesisScenarioIDs() []string {
	return []string{
		highTouchCompileSynthesisNaturalScenarioID,
		highTouchCompileSynthesisScriptedScenarioID,
	}
}

func compileSynthesisCandidateScenarioIDs() []string {
	return []string{
		compileSynthesisCurrentPrimitivesScenarioID,
		compileSynthesisGuidanceOnlyScenarioID,
		compileSynthesisResponseCandidateScenarioID,
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

func highTouchMemoryRouterRecallScenarioIDs() []string {
	return []string{
		highTouchMemoryRouterRecallNaturalScenarioID,
		highTouchMemoryRouterRecallScriptedScenarioID,
	}
}

func memoryRouterRecallCandidateScenarioIDs() []string {
	return []string{
		memoryRouterRecallCurrentPrimitivesScenarioID,
		memoryRouterRecallGuidanceOnlyScenarioID,
		memoryRouterRecallResponseCandidateScenarioID,
	}
}

func memoryRouterRecallReportScenarioIDs() []string {
	return []string{
		memoryRouterRecallReportActionScenarioID,
	}
}

func promotedRecordDomainScenarioIDs() []string {
	return []string{
		promotedRecordDomainNaturalScenarioID,
		promotedRecordDomainScriptedScenarioID,
	}
}

func highTouchRelationshipRecordScenarioIDs() []string {
	return []string{
		highTouchRelationshipRecordNaturalScenarioID,
		highTouchRelationshipRecordScriptedScenarioID,
	}
}

func relationshipRecordCandidateScenarioIDs() []string {
	return []string{
		relationshipRecordCurrentPrimitivesScenarioID,
		relationshipRecordGuidanceOnlyScenarioID,
		relationshipRecordResponseCandidateScenarioID,
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
