package main

import "context"

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
