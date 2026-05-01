package main

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"strings"
)

func verifyDocumentHistoryInspection(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	docID, found, err := documentIDByPath(ctx, paths, documentHistoryPolicyPath)
	if err != nil {
		return verificationResult{}, err
	}
	doc, _, err := documentByPath(ctx, paths, documentHistoryPolicyPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "document", RefID: docID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			RefKind: "document",
			RefID:   docID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasUpdatedBody := doc != nil && strings.Contains(doc.Body, "Current state: lifecycle inspection uses list_documents")
	hasProvenance := provenance.Provenance != nil &&
		eventTypesInclude(provenance.Provenance.Events, "document_created") &&
		eventTypesInclude(provenance.Provenance.Events, "document_updated")
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness != ""
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentHistoryPolicyPath)
	}
	if !hasUpdatedBody {
		failures = append(failures, "history inspection fixture did not expose updated lifecycle text")
	}
	if !hasProvenance {
		failures = append(failures, "document provenance missing created and updated events")
	}
	if !hasProjection {
		failures = append(failures, "document projection state missing or not fresh")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryPolicyPath}) &&
		messageContainsAny(finalMessage, []string{"provenance", "document_updated", "updated"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness", "fresh"}) &&
		messageContainsAny(finalMessage, []string{"existing", "current", "document and retrieval", "runner"})
	if !assistantPass {
		failures = append(failures, "final answer did not report history inspection, provenance, projection freshness, and existing runner workflow")
	}
	databasePass := found && hasUpdatedBody && hasProvenance && hasProjection
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryPolicyPath},
	}, nil
}

func verifyDocumentHistoryDiffReview(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	previous, previousFound, err := documentByPath(ctx, paths, documentHistoryDiffPreviousPath)
	if err != nil {
		return verificationResult{}, err
	}
	current, currentFound, err := documentByPath(ctx, paths, documentHistoryDiffCurrentPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !previousFound || previous == nil {
		failures = append(failures, "missing "+documentHistoryDiffPreviousPath)
	}
	if !currentFound || current == nil {
		failures = append(failures, "missing "+documentHistoryDiffCurrentPath)
	}
	if previous == nil || !strings.Contains(previous.Body, "optional review") {
		failures = append(failures, "previous evidence missing optional review text")
	}
	if current == nil || !strings.Contains(current.Body, "required review") {
		failures = append(failures, "current evidence missing required review text")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance")...)
	pathFailures := invalidRunnerPathFailures("list_documents path_prefix", turnMetrics.ListDocumentPathPrefixes)
	pathFailures = append(pathFailures, exactRunnerPathFailures("list_documents path_prefix", turnMetrics.ListDocumentPathPrefixes, documentHistoryDiffListPrefix)...)
	finalAnswerPathFailures := invalidRunnerPathTextFailures("final answer", finalMessage)
	failures = append(failures, pathFailures...)
	failures = append(failures, finalAnswerPathFailures...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryDiffPreviousPath, documentHistoryDiffCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"optional"}) &&
		messageContainsAny(finalMessage, []string{"required"}) &&
		messageContainsAny(finalMessage, []string{"citation", "cited", "source ref", "source_refs", "source"}) &&
		messageContainsAny(finalMessage, []string{"semantic", "summary"}) &&
		messageContainsAny(finalMessage, []string{"raw diff", "private diff", "do not expose raw", "no raw"}) &&
		len(finalAnswerPathFailures) == 0
	if !assistantPass {
		failures = append(failures, "final answer did not preserve cited semantic diff summary and raw-diff privacy handling")
	}
	databasePass := previousFound && currentFound &&
		previous != nil && current != nil &&
		strings.Contains(previous.Body, "optional review") &&
		strings.Contains(current.Body, "required review")
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 &&
		len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance")) == 0 &&
		len(pathFailures) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryDiffPreviousPath, documentHistoryDiffCurrentPath},
	}, nil
}

func verifyDocumentHistoryRestore(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	target, targetFound, err := documentByPath(ctx, paths, documentHistoryRestoreTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	targetID, _, err := documentIDByPath(ctx, paths, documentHistoryRestoreTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "document", RefID: targetID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			RefKind: "document",
			RefID:   targetID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	if target != nil {
		body = target.Body
	}
	restored := strings.Contains(body, "Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits.") &&
		!strings.Contains(body, "may bypass review")
	hasProvenance := provenance.Provenance != nil && eventTypesInclude(provenance.Provenance.Events, "document_updated")
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness != ""
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !targetFound {
		failures = append(failures, "missing "+documentHistoryRestoreTargetPath)
	}
	if !restored {
		failures = append(failures, "restore target was not restored to accepted lifecycle policy")
	}
	if !hasProvenance {
		failures = append(failures, "restore target provenance missing document update")
	}
	if !hasProjection {
		failures = append(failures, "restore target projection missing or not fresh")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryRestoreTargetPath, documentHistoryRestoreSourcePath}) &&
		messageContainsAny(finalMessage, []string{"restored", "restore", "rollback"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"source", "evidence", "citation"})
	if !assistantPass {
		failures = append(failures, "final answer did not report restore evidence, source, provenance, and projection freshness")
	}
	databasePass := targetFound && restored && hasProvenance && hasProjection
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryRestoreSourcePath, documentHistoryRestoreTargetPath},
	}, nil
}

func verifyHighTouchDocumentLifecycleScripted(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	result, err := verifyDocumentHistoryRestore(ctx, paths, finalMessage, turnMetrics)
	if err != nil {
		return verificationResult{}, err
	}
	targetID, targetFound, err := documentIDByPath(ctx, paths, documentHistoryRestoreTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !targetFound {
		failures = append(failures, "missing "+documentHistoryRestoreTargetPath)
	}
	failures = append(failures, invalidRunnerPathFailures("list_documents path_prefix", turnMetrics.ListDocumentPathPrefixes)...)
	failures = append(failures, exactRunnerPathFailures("list_documents path_prefix", turnMetrics.ListDocumentPathPrefixes, documentHistoryDiffListPrefix)...)
	if targetFound && !containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{targetID}) {
		failures = append(failures, "agent did not get restore target before editing")
	}
	if targetFound && !documentActionBefore(turnMetrics.DocumentActionEvents, "get_document:"+targetID, "replace_section:"+targetID) {
		failures = append(failures, "get_document for restore target did not precede replace_section")
	}
	if len(failures) == 0 {
		return result, nil
	}
	if result.Details != "" && result.Details != "ok" {
		failures = append([]string{result.Details}, failures...)
	}
	result.Passed = false
	result.AssistantPass = false
	result.Details = missingDetails(failures)
	return result, nil
}

func documentActionBefore(events []string, before string, after string) bool {
	beforeIndex := -1
	for index, event := range events {
		if event == before && beforeIndex < 0 {
			beforeIndex = index
		}
		if event == after {
			return beforeIndex >= 0 && beforeIndex < index
		}
	}
	return false
}

func verifyDocumentHistoryPendingReview(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	targetBody, targetFound, err := documentBodyByPath(ctx, paths, documentHistoryPendingTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	proposalBody, proposalFound, err := documentBodyByPath(ctx, paths, documentHistoryPendingProposalPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !targetFound {
		failures = append(failures, "missing "+documentHistoryPendingTargetPath)
	}
	if !strings.Contains(targetBody, "Accepted lifecycle policy: source-sensitive durable edits require human review before acceptance.") ||
		strings.Contains(targetBody, "Auto-accept pending change") {
		failures = append(failures, "accepted target changed instead of remaining under review")
	}
	requiredProposal := []string{
		"type: review",
		"status: pending",
		"Review state: pending human review.",
		"Proposed change: Auto-accept pending change only after operator approval.",
		"Target document: notes/history-review/pending-target.md",
	}
	if !proposalFound {
		failures = append(failures, "missing "+documentHistoryPendingProposalPath)
	}
	failures = append(failures, missingRequired(proposalBody, requiredProposal)...)
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryPendingTargetPath, documentHistoryPendingProposalPath}) &&
		messageContainsAny(finalMessage, []string{"pending", "review"}) &&
		messageContainsAny(finalMessage, []string{"not accepted", "not become accepted", "did not change", "no accepted change"}) &&
		messageContainsAny(finalMessage, []string{"human", "operator"})
	if !assistantPass {
		failures = append(failures, "final answer did not report pending review state and unchanged accepted target")
	}
	databasePass := targetFound && proposalFound &&
		strings.Contains(targetBody, "Accepted lifecycle policy: source-sensitive durable edits require human review before acceptance.") &&
		!strings.Contains(targetBody, "Auto-accept pending change") &&
		len(missingRequired(proposalBody, requiredProposal)) == 0
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryPendingTargetPath, documentHistoryPendingProposalPath},
	}, nil
}

func verifyDocumentHistoryStaleSynthesis(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	synthesisID, synthesisFound, err := documentIDByPath(ctx, paths, documentHistoryStaleSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	currentID, currentFound, err := documentIDByPath(ctx, paths, documentHistoryStaleCurrentSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, synthesisID)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "source", RefID: currentID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projectionEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "projection", RefID: "synthesis:" + synthesisID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasProjection := projection != nil &&
		projection.Freshness == "stale" &&
		projectionDetailContains(projection.Details, "stale_source_refs", documentHistoryStaleCurrentSourcePath)
	hasSourceEvents := currentFound && sourceEvents.Provenance != nil &&
		eventTypesInclude(sourceEvents.Provenance.Events, "source_updated")
	hasInvalidation := projectionEvents.Provenance != nil &&
		eventTypesInclude(projectionEvents.Provenance.Events, "projection_invalidated")
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !synthesisFound {
		failures = append(failures, "missing "+documentHistoryStaleSynthesisPath)
	}
	if !currentFound {
		failures = append(failures, "missing "+documentHistoryStaleCurrentSourcePath)
	}
	if !hasProjection {
		failures = append(failures, "synthesis projection is not stale with current source ref")
	}
	if !hasSourceEvents {
		failures = append(failures, "current source provenance missing source update")
	}
	if !hasInvalidation {
		failures = append(failures, "synthesis projection invalidation event missing")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryStaleSynthesisPath, documentHistoryStaleCurrentSourcePath}) &&
		messageContainsAny(finalMessage, []string{"stale"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "invalidated", "source_updated", "updated"}) &&
		messageContainsAny(finalMessage, []string{"no repair", "not repair", "did not repair", "without repair"})
	if !assistantPass {
		failures = append(failures, "final answer did not report stale synthesis, provenance/invalidation, and no repair")
	}
	databasePass := synthesisFound && currentFound && hasProjection && hasSourceEvents && hasInvalidation
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryStaleSynthesisPath, documentHistoryStaleCurrentSourcePath, documentHistoryStaleOldSourcePath},
	}, nil
}

func invalidRunnerPathFailures(label string, values []string) []string {
	failures := []string{}
	for _, value := range values {
		if isInvalidRunnerPath(value) {
			failures = append(failures, label+" used non-vault-relative path "+value)
		}
	}
	return failures
}

func exactRunnerPathFailures(label string, values []string, allowed ...string) []string {
	failures := []string{}
	allowedSet := map[string]struct{}{}
	seen := map[string]bool{}
	for _, value := range allowed {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		allowedSet[trimmed] = struct{}{}
		seen[trimmed] = false
	}
	if len(values) == 0 {
		for value := range allowedSet {
			failures = append(failures, label+" missing required path "+value)
		}
		return failures
	}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if _, ok := allowedSet[trimmed]; ok {
			seen[trimmed] = true
			continue
		}
		failures = append(failures, label+" used unexpected path "+value)
	}
	for value, found := range seen {
		if !found {
			failures = append(failures, label+" missing required path "+value)
		}
	}
	return failures
}

func invalidRunnerPathTextFailures(label string, value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	lower := strings.ToLower(normalized)
	if strings.Contains(lower, ".openclerk-eval") ||
		strings.Contains(lower, "/vault/") ||
		strings.Contains(lower, "vault/") ||
		unixAbsolutePathPattern.MatchString(normalized) ||
		windowsDrivePathPattern.MatchString(trimmed) ||
		strings.Contains(trimmed, "\\") {
		return []string{label + " included non-vault-relative path text"}
	}
	return nil
}

func isInvalidRunnerPath(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	lower := strings.ToLower(normalized)
	if strings.Contains(lower, ".openclerk-eval") || strings.Contains(lower, "/vault/") || strings.HasPrefix(lower, "vault/") {
		return true
	}
	if strings.HasPrefix(normalized, "/") || strings.HasPrefix(normalized, "~") {
		return true
	}
	if len(trimmed) >= 3 && ((trimmed[0] >= 'A' && trimmed[0] <= 'Z') || (trimmed[0] >= 'a' && trimmed[0] <= 'z')) && trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/') {
		return true
	}
	return strings.Contains(trimmed, "\\")
}

func missingDocumentHistoryMetrics(turnMetrics metrics, required ...string) []string {
	failures := []string{}
	for _, requirement := range required {
		switch requirement {
		case "search":
			if !turnMetrics.SearchUsed {
				failures = append(failures, "agent did not use retrieval search")
			}
		case "list":
			if !turnMetrics.ListDocumentsUsed {
				failures = append(failures, "agent did not use list_documents")
			}
		case "get":
			if !turnMetrics.GetDocumentUsed {
				failures = append(failures, "agent did not use get_document")
			}
		case "provenance":
			if !turnMetrics.ProvenanceEventsUsed {
				failures = append(failures, "agent did not inspect provenance events")
			}
		case "projection":
			if !turnMetrics.ProjectionStatesUsed {
				failures = append(failures, "agent did not inspect projection states")
			}
		}
	}
	return failures
}
