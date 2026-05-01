package main

import (
	"context"
	"strings"
	"testing"
)

func TestVerifyRelationshipRecordResponseCandidateRequiresContractAndWorkflow(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: relationshipRecordResponseCandidateScenarioID}); err != nil {
		t.Fatalf("seed relationship-record candidate scenario: %v", err)
	}
	graphDocID, found, err := documentIDByPath(ctx, paths, graphSemanticsIndexPath)
	if err != nil {
		t.Fatalf("lookup graph doc id: %v", err)
	}
	if !found {
		t.Fatalf("missing graph doc id for %s", graphSemanticsIndexPath)
	}
	promotedRecordDocID, found, err := documentIDByPath(ctx, paths, promotedRecordDomainPrimaryPath)
	if err != nil {
		t.Fatalf("lookup promoted record doc id: %v", err)
	}
	if !found {
		t.Fatalf("missing promoted record doc id for %s", promotedRecordDomainPrimaryPath)
	}
	workflowMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{graphSemanticsPrefix, promotedRecordDomainPrefix},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{graphDocID, promotedRecordDocID},
		DocumentLinksUsed:        true,
		GraphNeighborhoodUsed:    true,
		ProjectionStatesUsed:     true,
		RecordsLookupUsed:        true,
		RecordEntityUsed:         true,
		RecordEntityIDs:          []string{promotedRecordDomainEntityID},
		ProvenanceEventsUsed:     true,
		EventTypeCounts:          map[string]int{},
		CommandExecutions:        10,
		ToolCalls:                10,
	}
	candidateAnswer := "```json\n{\"query_summary\":\"relationship-record lookup for graph semantics relationships plus AgentOps Escalation Policy record evidence\",\"relationship_evidence\":\"" + graphSemanticsIndexPath + " canonical markdown says requires, supersedes, related to, and operationalizes; graph projections are derived evidence, not independent authority\",\"link_evidence\":\"document_links for " + graphDocID + " include outgoing links to " + graphSemanticsRoutingPath + ", " + graphSemanticsFreshnessPath + ", " + graphSemanticsOperationsPath + " and incoming backlinks from linked graph semantics docs\",\"graph_freshness\":\"fresh graph projection for " + graphSemanticsIndexPath + "\",\"record_lookup_evidence\":\"records_lookup found entity_id " + promotedRecordDomainEntityID + " for " + promotedRecordDomainEntityName + " with citation evidence from " + promotedRecordDomainPrimaryPath + "\",\"record_entity_evidence\":\"record_entity " + promotedRecordDomainEntityID + " reports policy owner platform, status active, review cadence monthly\",\"citation_refs\":[\"" + graphSemanticsIndexPath + "\",\"" + promotedRecordDomainPrimaryPath + "\"],\"provenance_refs\":[\"entity:" + promotedRecordDomainEntityID + "\",\"" + promotedRecordDomainEntityID + "\",\"runner-owned no-bypass\"],\"records_freshness\":\"fresh records projection for entity " + promotedRecordDomainEntityID + "\",\"validation_boundaries\":\"no direct SQLite, no direct vault inspection, no direct file edits, no broad repo search, no source-built runner, no unsupported transports or actions; read-only current openclerk document and retrieval JSON only\",\"authority_limits\":\"canonical markdown remains authority; graph and records projections are derived evidence with citations, provenance, and freshness; this eval-only response does not implement a relationship-record lookup action\"}\n```"
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: relationshipRecordResponseCandidateScenarioID}, 1, candidateAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify relationship-record candidate: %v", err)
	}
	if !result.Passed {
		t.Fatalf("relationship-record candidate failed: %+v", result)
	}
	proseWrappedAnswer := "Candidate:\n" + candidateAnswer
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: relationshipRecordResponseCandidateScenarioID}, 1, proseWrappedAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify prose-wrapped relationship-record candidate: %v", err)
	}
	if result.Passed {
		t.Fatalf("relationship-record candidate passed with prose outside JSON fence: %+v", result)
	}
	missingAuthorityAnswer := strings.Replace(candidateAnswer, "canonical markdown remains authority", "graph and records are authoritative", 1)
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: relationshipRecordResponseCandidateScenarioID}, 1, missingAuthorityAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify relationship-record candidate missing authority limit: %v", err)
	}
	if result.Passed {
		t.Fatalf("relationship-record candidate passed without canonical authority limit: %+v", result)
	}
	extraFieldAnswer := strings.Replace(candidateAnswer, "}\n```", ",\"unexpected\":\"field\"}\n```", 1)
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: relationshipRecordResponseCandidateScenarioID}, 1, extraFieldAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify relationship-record candidate with extra field: %v", err)
	}
	if result.Passed {
		t.Fatalf("relationship-record candidate passed with an extra response field: %+v", result)
	}
	missingGraphMetrics := workflowMetrics
	missingGraphMetrics.GraphNeighborhoodUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: relationshipRecordResponseCandidateScenarioID}, 1, candidateAnswer, missingGraphMetrics)
	if err != nil {
		t.Fatalf("verify relationship-record candidate missing graph metrics: %v", err)
	}
	if result.Passed {
		t.Fatalf("relationship-record candidate passed without graph_neighborhood metrics: %+v", result)
	}
	missingPromotedRecordDocsMetrics := workflowMetrics
	missingPromotedRecordDocsMetrics.ListDocumentPathPrefixes = []string{graphSemanticsPrefix}
	missingPromotedRecordDocsMetrics.GetDocumentDocIDs = []string{graphDocID}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: relationshipRecordResponseCandidateScenarioID}, 1, candidateAnswer, missingPromotedRecordDocsMetrics)
	if err != nil {
		t.Fatalf("verify relationship-record candidate missing promoted record document metrics: %v", err)
	}
	if result.Passed {
		t.Fatalf("relationship-record candidate passed without promoted record document metrics: %+v", result)
	}
}
