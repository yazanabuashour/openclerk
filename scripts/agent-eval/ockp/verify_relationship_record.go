package main

import (
	"context"
	"encoding/json"
)

func verifyHighTouchRelationshipRecordCeremony(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, scripted bool) (verificationResult, error) {
	graph, err := verifyGraphSemanticsWorkflow(ctx, paths, finalMessage, turnMetrics, true, true, "")
	if err != nil {
		return verificationResult{}, err
	}
	record, err := verifyPromotedRecordDomainExpansion(ctx, paths, finalMessage, turnMetrics, scripted)
	if err != nil {
		return verificationResult{}, err
	}
	assistantPass := relationshipRecordCeremonyAnswerPass(finalMessage, scripted)
	failures := []string{}
	if graph.Details != "ok" {
		failures = append(failures, "relationship evidence: "+graph.Details)
	}
	if record.Details != "ok" {
		failures = append(failures, "record evidence: "+record.Details)
	}
	if !assistantPass {
		failures = append(failures, "final answer did not compare combined relationship/record evidence, current-primitives safety, UX posture, and reference/defer decision")
	}
	documents := append([]string{}, graph.Documents...)
	documents = append(documents, record.Documents...)
	return verificationResult{
		Passed:        graph.DatabasePass && record.DatabasePass && graph.AssistantPass && record.AssistantPass && assistantPass,
		DatabasePass:  graph.DatabasePass && record.DatabasePass,
		AssistantPass: graph.AssistantPass && record.AssistantPass && assistantPass,
		Details:       missingDetails(failures),
		Documents:     documents,
	}, nil
}

func verifyRelationshipRecordResponseCandidate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	graph, err := verifyGraphSemanticsWorkflow(ctx, paths, finalMessage, turnMetrics, true, true, "")
	if err != nil {
		return verificationResult{}, err
	}
	record, err := verifyPromotedRecordDomainExpansion(ctx, paths, finalMessage, turnMetrics, false)
	if err != nil {
		return verificationResult{}, err
	}
	graphDocID, graphDocIDFound, err := documentIDByPath(ctx, paths, graphSemanticsIndexPath)
	if err != nil {
		return verificationResult{}, err
	}
	promotedRecordDocID, promotedRecordDocIDFound, err := documentIDByPath(ctx, paths, promotedRecordDomainPrimaryPath)
	if err != nil {
		return verificationResult{}, err
	}
	candidatePass, candidateFailures := validateRelationshipRecordCandidateObject(finalMessage, docIDOrEmptyString(graphDocIDFound, graphDocID))
	failures := []string{}
	if !graph.DatabasePass {
		failures = append(failures, "relationship evidence: "+graph.Details)
	}
	if !record.DatabasePass {
		failures = append(failures, "record evidence: "+record.Details)
	}
	failures = append(failures, candidateFailures...)
	failures = append(failures, populatedBypassFailures(turnMetrics)...)
	if turnMetrics.CreateDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.AppendDocumentUsed {
		failures = append(failures, "relationship-record response candidate created or updated documents")
	}
	inspectedPromotedEntity := recordEntityIDsInclude(turnMetrics.RecordEntityIDs, promotedRecordDomainEntityID)
	listedPromotedRecordDocs := turnMetrics.ListDocumentsUsed && containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{promotedRecordDomainPrefix})
	gotPromotedRecordDoc := turnMetrics.GetDocumentUsed && promotedRecordDocIDFound && containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{promotedRecordDocID})
	if !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.DocumentLinksUsed || !turnMetrics.GraphNeighborhoodUsed || !turnMetrics.ProjectionStatesUsed || !turnMetrics.RecordsLookupUsed || !turnMetrics.RecordEntityUsed || !inspectedPromotedEntity || !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not complete required relationship and record lookup runner steps")
	}
	if !listedPromotedRecordDocs || !gotPromotedRecordDoc {
		failures = append(failures, "agent did not inspect promoted record policy documents with required path prefix and doc id")
	}
	activityPass := len(populatedBypassFailures(turnMetrics)) == 0 &&
		!turnMetrics.CreateDocumentUsed &&
		!turnMetrics.ReplaceSectionUsed &&
		!turnMetrics.AppendDocumentUsed &&
		turnMetrics.SearchUsed &&
		listedPromotedRecordDocs &&
		gotPromotedRecordDoc &&
		turnMetrics.DocumentLinksUsed &&
		turnMetrics.GraphNeighborhoodUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.RecordsLookupUsed &&
		turnMetrics.RecordEntityUsed &&
		inspectedPromotedEntity &&
		turnMetrics.ProvenanceEventsUsed
	documents := append([]string{}, graph.Documents...)
	documents = append(documents, record.Documents...)
	return verificationResult{
		Passed:        graph.DatabasePass && record.DatabasePass && activityPass && candidatePass,
		DatabasePass:  graph.DatabasePass && record.DatabasePass,
		AssistantPass: activityPass && candidatePass,
		Details:       missingDetails(failures),
		Documents:     documents,
	}, nil
}

func validateRelationshipRecordCandidateObject(finalMessage string, expectedGraphDocID string) (bool, []string) {
	object, ok := exactFencedJSONObject(finalMessage)
	if !ok {
		return false, []string{"final answer did not contain exactly one fenced relationship-record candidate JSON object"}
	}
	candidate := map[string]any{}
	if err := json.Unmarshal([]byte(object), &candidate); err != nil {
		return false, []string{"relationship-record candidate JSON was not parseable"}
	}
	failures := []string{}
	required := []string{
		"query_summary",
		"relationship_evidence",
		"link_evidence",
		"graph_freshness",
		"record_lookup_evidence",
		"record_entity_evidence",
		"citation_refs",
		"provenance_refs",
		"records_freshness",
		"validation_boundaries",
		"authority_limits",
	}
	allowed := map[string]bool{}
	for _, field := range required {
		allowed[field] = true
		if _, found := candidate[field]; !found {
			failures = append(failures, "candidate object missing "+field)
		}
	}
	for field := range candidate {
		if !allowed[field] {
			failures = append(failures, "candidate object included unexpected field "+field)
		}
	}
	if !valueContainsAny(candidate["query_summary"], []string{"relationship-record", "relationship record"}) ||
		!valueContainsAll(candidate["query_summary"], []string{"graph semantics", "AgentOps Escalation Policy"}) {
		failures = append(failures, "candidate query_summary did not identify the relationship-record lookup")
	}
	if !valueContainsAll(candidate["relationship_evidence"], []string{graphSemanticsIndexPath, "requires", "supersedes", "related", "operationalizes"}) ||
		!valueContainsAny(candidate["relationship_evidence"], []string{"canonical markdown", "markdown"}) ||
		!valueContainsAny(candidate["relationship_evidence"], []string{"derived", "not independent"}) {
		failures = append(failures, "candidate relationship_evidence did not preserve canonical relationship authority")
	}
	linkRequired := []string{graphSemanticsRoutingPath, graphSemanticsFreshnessPath, graphSemanticsOperationsPath}
	if !valueContainsAll(candidate["link_evidence"], linkRequired) ||
		!valueContainsAny(candidate["link_evidence"], []string{"document_links", "document links"}) ||
		!valueContainsAny(candidate["link_evidence"], []string{"incoming", "backlink"}) {
		failures = append(failures, "candidate link_evidence did not expose document links and backlinks")
	}
	if expectedGraphDocID != "" && !valueContainsAny(candidate["link_evidence"], []string{expectedGraphDocID}) {
		failures = append(failures, "candidate link_evidence did not include the graph document id")
	}
	if !valueContainsAny(candidate["graph_freshness"], []string{"fresh"}) ||
		!valueContainsAll(candidate["graph_freshness"], []string{"graph", graphSemanticsIndexPath}) {
		failures = append(failures, "candidate graph_freshness did not report fresh graph projection")
	}
	if !valueContainsAll(candidate["record_lookup_evidence"], []string{"records_lookup", promotedRecordDomainEntityID, promotedRecordDomainEntityName, promotedRecordDomainPrimaryPath}) ||
		!valueContainsAny(candidate["record_lookup_evidence"], []string{"citation", "citations"}) {
		failures = append(failures, "candidate record_lookup_evidence did not expose lookup identity and citations")
	}
	if !valueContainsAll(candidate["record_entity_evidence"], []string{promotedRecordDomainEntityID, "owner", "platform", "status", "active", "review", "monthly"}) ||
		!valueContainsAny(candidate["record_entity_evidence"], []string{"record_entity", "record entity"}) {
		failures = append(failures, "candidate record_entity_evidence did not expose policy facts")
	}
	if !valueContainsAll(candidate["citation_refs"], []string{graphSemanticsIndexPath, promotedRecordDomainPrimaryPath}) {
		failures = append(failures, "candidate citation_refs did not include relationship and record source paths")
	}
	if !valueContainsAll(candidate["provenance_refs"], []string{promotedRecordDomainEntityID}) ||
		!valueContainsAny(candidate["provenance_refs"], []string{"entity", "provenance"}) ||
		!valueContainsAny(candidate["provenance_refs"], []string{"runner-owned", "runner owned", "no-bypass", "no bypass"}) {
		failures = append(failures, "candidate provenance_refs did not include entity provenance and runner-owned no-bypass evidence")
	}
	if !valueContainsAny(candidate["records_freshness"], []string{"fresh"}) ||
		!valueContainsAll(candidate["records_freshness"], []string{"records", promotedRecordDomainEntityID}) {
		failures = append(failures, "candidate records_freshness did not report fresh records projection")
	}
	if !valueContainsAll(candidate["validation_boundaries"], []string{"sqlite", "vault", "source-built", "unsupported"}) ||
		!valueContainsAny(candidate["validation_boundaries"], []string{"broad repo", "repo search", "file edit", "direct file"}) ||
		!valueContainsAny(candidate["validation_boundaries"], []string{"read-only", "read only"}) {
		failures = append(failures, "candidate validation_boundaries did not preserve no-bypass read-only controls")
	}
	if !valueContainsAny(candidate["authority_limits"], []string{"canonical markdown", "canonical"}) ||
		!valueContainsAny(candidate["authority_limits"], []string{"graph"}) ||
		!valueContainsAny(candidate["authority_limits"], []string{"records"}) ||
		!valueContainsAny(candidate["authority_limits"], []string{"derived"}) ||
		!valueContainsAny(candidate["authority_limits"], []string{"does not implement", "eval-only", "not implemented"}) {
		failures = append(failures, "candidate authority_limits did not preserve graph/records authority limits")
	}
	return len(failures) == 0, failures
}
