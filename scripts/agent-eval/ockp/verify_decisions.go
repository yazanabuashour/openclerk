package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

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
	assistantPass := messageContainsAny(finalMessage, []string{"decisions lookup", "decision lookup", "decisions_lookup", "decision records"}) &&
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
	auditReportUsed := turnMetrics.AuditContradictionsUsed || turnMetrics.SourceAuditReportUsed
	activityPass := auditReportUsed || turnMetrics.SearchUsed &&
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
	if !auditReportUsed {
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
	explainsCurrentSourceEvidence := messageContainsAny(finalMessage, []string{"both are current", "both sources are current", "current sources", "both current"}) ||
		(messageContainsAny(finalMessage, []string{"current source says", "current-source says"}) &&
			messageContainsAny(finalMessage, []string{"seven", "7"}) &&
			messageContainsAny(finalMessage, []string{"thirty", "30"})) ||
		(messageContainsAny(finalMessage, []string{"only `document_created`", "only document_created", "document_created events", "document_created event"}) &&
			messageContainsAny(finalMessage, []string{"no supersession", "no source authority"}))
	assistantPass := messageContainsAll(finalMessage, []string{sourceAuditConflictAlphaPath, sourceAuditConflictBravoPath}) &&
		messageContainsAny(finalMessage, []string{"conflict", "conflicting", "contradict", "contradiction"}) &&
		explainsCurrentSourceEvidence &&
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
func verifyBroadContradictionAuditRevisit(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, scripted bool) (verificationResult, error) {
	repair, err := verifySourceSensitiveAuditRepair(ctx, paths, finalMessage, turnMetrics)
	if err != nil {
		return verificationResult{}, err
	}
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
	hasConflictProvenance := alphaFound && bravoFound &&
		alphaEvents.Provenance != nil && len(alphaEvents.Provenance.Events) > 0 &&
		bravoEvents.Provenance != nil && len(bravoEvents.Provenance.Events) > 0
	inspectedBothProvenanceRefs := provenanceEventRefIDsInclude(turnMetrics.ProvenanceEventRefIDs, alphaID, bravoID)
	auditReportUsed := turnMetrics.AuditContradictionsUsed || turnMetrics.SourceAuditReportUsed
	conflictActivityPass := auditReportUsed || turnMetrics.SearchUsed && inspectedBothProvenanceRefs
	conflictAnswerPass := messageContainsAll(finalMessage, []string{sourceAuditConflictAlphaPath, sourceAuditConflictBravoPath}) &&
		messageContainsAny(finalMessage, []string{"conflict", "conflicting", "contradict", "contradiction"}) &&
		(auditReportUsed || messageContainsAny(finalMessage, []string{"both are current", "both sources are current", "current sources", "both current"})) &&
		messageContainsAny(finalMessage, []string{"unresolved", "no supersession", "no source authority", "cannot choose", "do not choose"}) &&
		messageContainsAny(finalMessage, []string{"seven", "7"}) &&
		messageContainsAny(finalMessage, []string{"thirty", "30"})
	decisionAnswerPass := broadContradictionAuditAnswerPass(finalMessage, scripted)

	failures := populatedBypassFailures(turnMetrics)
	if !repair.Passed {
		failures = append(failures, "audit repair failed: "+repair.Details)
	}
	if turnMetrics.CreateDocumentUsed {
		failures = append(failures, "agent created a document instead of updating existing synthesis and explaining conflict")
	}
	if !searchHasBoth {
		failures = append(failures, "search did not find both conflict sources")
	}
	if !hasConflictProvenance {
		failures = append(failures, "document provenance missing for conflict sources")
	}
	if synthesisCount != 2 {
		failures = append(failures, fmt.Sprintf("expected target and decoy synthesis documents only, got %d", synthesisCount))
	}
	if !auditReportUsed && !inspectedBothProvenanceRefs {
		failures = append(failures, "agent did not use source audit report or inspect provenance events for both conflict sources")
	}
	if !conflictAnswerPass {
		failures = append(failures, "final answer did not explain unresolved conflicting source evidence")
	}
	if !decisionAnswerPass {
		failures = append(failures, "final answer did not classify capability/ergonomics posture and reference/defer decision")
	}
	databasePass := repair.DatabasePass && searchHasBoth && hasConflictProvenance && synthesisCount == 2
	assistantPass := len(populatedBypassFailures(turnMetrics)) == 0 && !turnMetrics.CreateDocumentUsed && repair.AssistantPass && conflictAnswerPass && decisionAnswerPass && conflictActivityPass
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceAuditSynthesisPath, sourceAuditDecoyPath, sourceAuditCurrentSourcePath, sourceAuditOldSourcePath, sourceAuditConflictAlphaPath, sourceAuditConflictBravoPath},
	}, nil
}

func verifySourceAuditWorkflowAction(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	repair, err := verifySourceSensitiveAuditRepair(ctx, paths, finalMessage, turnMetrics)
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: sourceAuditConflictSearchText, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	searchHasBoth := searchContainsPath(search, sourceAuditConflictAlphaPath) && searchContainsPath(search, sourceAuditConflictBravoPath)
	conflictAnswerPass := messageContainsAll(finalMessage, []string{sourceAuditConflictAlphaPath, sourceAuditConflictBravoPath}) &&
		messageContainsAny(finalMessage, []string{"conflict", "conflicting", "contradict", "contradiction"}) &&
		messageContainsAny(finalMessage, []string{"unresolved", "no supersession", "no source authority", "cannot choose", "do not choose"}) &&
		messageContainsAny(finalMessage, []string{"seven", "7"}) &&
		messageContainsAny(finalMessage, []string{"thirty", "30"})
	actionAnswerPass := messageContainsAll(finalMessage, []string{"source_audit_report", sourceAuditSynthesisPath, sourceAuditCurrentSourcePath}) &&
		messageContainsAll(finalMessage, []string{"provenance", "projection", "freshness", "duplicate", "validation", "authority"})
	failures := populatedBypassFailures(turnMetrics)
	if !repair.Passed {
		failures = append(failures, "source audit repair failed: "+repair.Details)
	}
	if !turnMetrics.SourceAuditReportUsed {
		failures = append(failures, "agent did not use source_audit_report")
	}
	if turnMetrics.CreateDocumentUsed {
		failures = append(failures, "agent created a document instead of repairing only existing synthesis")
	}
	if !searchHasBoth {
		failures = append(failures, "search did not find both conflict sources")
	}
	if synthesisCount != 2 {
		failures = append(failures, fmt.Sprintf("expected target and decoy synthesis documents only, got %d", synthesisCount))
	}
	if !conflictAnswerPass {
		failures = append(failures, "final answer did not explain unresolved conflicting source evidence")
	}
	if !actionAnswerPass {
		failures = append(failures, "final answer did not report source_audit_report evidence, validation boundaries, and authority limits")
	}
	databasePass := repair.DatabasePass && searchHasBoth && synthesisCount == 2
	assistantPass := len(populatedBypassFailures(turnMetrics)) == 0 &&
		!turnMetrics.CreateDocumentUsed &&
		turnMetrics.SourceAuditReportUsed &&
		repair.AssistantPass &&
		conflictAnswerPass &&
		actionAnswerPass
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceAuditSynthesisPath, sourceAuditDecoyPath, sourceAuditCurrentSourcePath, sourceAuditOldSourcePath, sourceAuditConflictAlphaPath, sourceAuditConflictBravoPath},
	}, nil
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
