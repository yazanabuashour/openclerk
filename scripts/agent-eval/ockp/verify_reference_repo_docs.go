package main

import (
	"context"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

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

func verifyRepoDocsReleaseReadiness(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       repoDocsReleaseSearchText,
			PathPrefix: "docs/",
			Limit:      10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			Tag:   repoDocsReleaseTag,
			Limit: 20,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	failures := repoDocsBypassFailures(turnMetrics)
	hasReleaseVerification := searchContainsPath(search, repoDocsReleaseVerification) || documentSummariesContainPath(list.Documents, repoDocsReleaseVerification)
	hasMaintainers := searchContainsPath(search, repoDocsMaintainersPath) || documentSummariesContainPath(list.Documents, repoDocsMaintainersPath)
	if !hasReleaseVerification {
		failures = append(failures, "release-verification doc missing from repo-docs release readiness evidence")
	}
	if !hasMaintainers {
		failures = append(failures, "maintainers doc missing from repo-docs release readiness evidence")
	}
	if !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"docs/"}) {
		failures = append(failures, "agent did not use docs/ path-prefix search")
	}
	if !turnMetrics.ListTagFilterUsed || !stringValuesInclude(turnMetrics.ListTagFilters, repoDocsReleaseTag) {
		failures = append(failures, "agent did not list repo-release-docs tag")
	}
	assistantPass := messageContainsAll(finalMessage, []string{repoDocsReleaseVerification, repoDocsMaintainersPath, repoDocsReleaseTag}) &&
		messageContainsAny(finalMessage, []string{"dogfood", "dogfooding"}) &&
		messageContainsAny(finalMessage, []string{"mandatory", "required", "before tagging"}) &&
		messageContainsAny(finalMessage, []string{"validate-release-docs", "validate release docs"}) &&
		messageContainsAny(finalMessage, []string{"AgentOps", "production gate"})
	if !assistantPass {
		failures = append(failures, "final answer did not report mandatory dogfood release readiness evidence")
	}
	databasePass := hasReleaseVerification && hasMaintainers
	activityPass := len(repoDocsBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListTagFilterUsed &&
		stringValuesInclude(turnMetrics.ListTagFilters, repoDocsReleaseTag)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{repoDocsReleaseVerification, repoDocsMaintainersPath},
	}, nil
}

func verifyRepoDocsTagFilter(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  repoDocsReleaseSearchText,
			Tag:   repoDocsReleaseTag,
			Limit: 10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Tag: repoDocsReleaseTag, Limit: 20},
	})
	if err != nil {
		return verificationResult{}, err
	}
	failures := repoDocsBypassFailures(turnMetrics)
	searchHasTaggedReleaseEvidence := search.Search != nil && len(search.Search.Hits) > 0
	listHasReleaseDoc := documentSummariesContainPath(list.Documents, repoDocsReleaseVerification)
	listHasMaintainers := documentSummariesContainPath(list.Documents, repoDocsMaintainersPath)
	if !searchHasTaggedReleaseEvidence {
		failures = append(failures, "tag-filtered search did not return tagged release evidence")
	}
	if !listHasReleaseDoc || !listHasMaintainers {
		failures = append(failures, "tag-filtered list did not return expected release docs")
	}
	if !turnMetrics.SearchTagFilterUsed || !stringValuesInclude(turnMetrics.SearchTagFilters, repoDocsReleaseTag) {
		failures = append(failures, "agent did not use search tag filter")
	}
	if !turnMetrics.ListTagFilterUsed || !stringValuesInclude(turnMetrics.ListTagFilters, repoDocsReleaseTag) {
		failures = append(failures, "agent did not use list tag filter")
	}
	assistantPass := messageContainsAll(finalMessage, []string{repoDocsReleaseVerification, repoDocsMaintainersPath, repoDocsReleaseTag}) &&
		messageContainsAny(finalMessage, []string{"tag", "tag filter", "tag-filter"}) &&
		messageContainsAny(finalMessage, []string{"canonical markdown", "canonical", "authority"})
	if !assistantPass {
		failures = append(failures, "final answer did not report tag-filter evidence and canonical authority boundary")
	}
	databasePass := searchHasTaggedReleaseEvidence && listHasReleaseDoc && listHasMaintainers
	activityPass := len(repoDocsBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchTagFilterUsed && stringValuesInclude(turnMetrics.SearchTagFilters, repoDocsReleaseTag) &&
		turnMetrics.ListTagFilterUsed && stringValuesInclude(turnMetrics.ListTagFilters, repoDocsReleaseTag)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{repoDocsReleaseVerification, repoDocsMaintainersPath},
	}, nil
}

func verifyRepoDocsMemoryRouterRecallReport(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Query: repoDocsMemorySearchText,
			Limit: 10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	failures := repoDocsBypassFailures(turnMetrics)
	report := result.MemoryRouterRecall
	databasePass := report != nil &&
		len(report.CanonicalEvidenceRefs) > 0 &&
		len(report.ProvenanceRefs) > 0 &&
		report.SynthesisFreshness != "" &&
		report.ValidationBoundaries != "" &&
		report.AuthorityLimits != ""
	if !databasePass {
		failures = append(failures, "memory_router_recall_report missing required evidence fields")
	}
	assistantPass := messageContainsAny(finalMessage, []string{"memory_router_recall_report", "memory router recall report"}) &&
		messageContainsAny(finalMessage, []string{"canonical evidence", "canonical_evidence_refs"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "provenance_refs"}) &&
		messageContainsAny(finalMessage, []string{"synthesis freshness", "synthesis_freshness"}) &&
		messageContainsAny(finalMessage, []string{"validation", "boundaries"}) &&
		messageContainsAny(finalMessage, []string{"authority", "authority limits"}) &&
		messageContainsAny(finalMessage, []string{"read-only", "read only", "read-only dogfood"})
	if !assistantPass {
		failures = append(failures, "final answer did not summarize memory-router recall report fields and non-promotion boundary")
	}
	activityPass := len(repoDocsBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.MemoryRouterRecallReportUsed
	if !turnMetrics.MemoryRouterRecallReportUsed {
		failures = append(failures, "agent did not use memory_router_recall_report")
	}
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{memoryRouterSynthesisPath, memoryRouterSessionObservationPath},
	}, nil
}

func verifyRepoDocsReleaseSynthesisFreshness(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	doc, found, err := documentByPath(ctx, paths, repoDocsReleaseSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := repoDocsBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing repo-docs release synthesis")
		return verificationFromFailures(failures, "repo-docs release synthesis missing", []string{repoDocsReleaseSynthesisPath})
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      doc.DocID,
			Limit:      10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + doc.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasStaleProjection := false
	if projections.Projections != nil {
		for _, projection := range projections.Projections.Projections {
			if projection.RefID == doc.DocID && projection.Freshness == "stale" {
				hasStaleProjection = true
				break
			}
		}
	}
	hasProvenance := provenance.Provenance != nil && len(provenance.Provenance.Events) > 0
	if !hasStaleProjection {
		failures = append(failures, "repo-docs release synthesis projection is not stale after source update")
	}
	if !hasProvenance {
		failures = append(failures, "repo-docs release synthesis projection provenance missing")
	}
	if !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProjectionStatesUsed || !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not use required search/list/get/projection/provenance workflow")
	}
	assistantPass := messageContainsAll(finalMessage, []string{repoDocsReleaseSynthesisPath, repoDocsReleaseVerification, repoDocsMaintainersPath}) &&
		messageContainsAny(finalMessage, []string{"stale", "freshness", "projection"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "projection_refreshed", "event"}) &&
		messageContainsAny(finalMessage, []string{"no repair", "not repaired", "no update", "did not perform"})
	if !assistantPass {
		failures = append(failures, "final answer did not report release synthesis freshness/provenance and no-repair boundary")
	}
	databasePass := hasStaleProjection && hasProvenance
	activityPass := len(repoDocsBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed && turnMetrics.ProvenanceEventsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{repoDocsReleaseSynthesisPath, repoDocsReleaseVerification, repoDocsMaintainersPath},
	}, nil
}

func documentSummariesContainPath(documents []runner.DocumentSummary, path string) bool {
	for _, document := range documents {
		if document.Path == path {
			return true
		}
	}
	return false
}
