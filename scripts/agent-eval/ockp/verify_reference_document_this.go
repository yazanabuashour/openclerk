package main

import (
	"context"
	"fmt"
	"strings"
)

func verifyDocumentThisExplicitCreate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, documentThisExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: note",
		"Document-this explicit article/docs/paper/transcript intake uses strict runner JSON.",
		"Required fields were supplied before create_document.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisExplicitPath)
	}
	if found && title != documentThisExplicitTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", documentThisExplicitTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no source autofiling docs, got %d", sourcesCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisExplicitPath, documentThisExplicitTitle})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit document path and title")
	}
	databasePass := found && title == documentThisExplicitTitle && len(missingRequired(body, required)) == 0 && sourcesCount == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisExplicitPath},
	}, nil
}

func verifyDocumentThisExplicitOverrides(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, documentThisOverridePath)
	if err != nil {
		return verificationResult{}, err
	}
	autofiledCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	required := []string{
		"type: note",
		"Explicit document-this override path and title win.",
		"Do not infer a sources/ path from mixed URLs.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisOverridePath)
	}
	if found && title != documentThisOverrideTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", documentThisOverrideTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if autofiledCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no inferred source docs, got %d", autofiledCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit override through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisOverridePath, documentThisOverrideTitle})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit override path and title")
	}
	databasePass := found && title == documentThisOverrideTitle && len(missingRequired(body, required)) == 0 && autofiledCount == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisOverridePath},
	}, nil
}

func verifyDocumentThisDuplicateCandidate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingCount, err := exactDocumentCount(ctx, paths, documentThisDuplicateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, documentThisDuplicateCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentThisBypassFailures(turnMetrics)
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing source %s, got %d", documentThisDuplicateExistingPath, existingCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created duplicate candidate "+documentThisDuplicateCandidatePath)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected only the seeded document-this source document, got %d", sourceCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search for duplicate candidate")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source candidates")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisDuplicateExistingPath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already"}) &&
		messageContainsAny(finalMessage, []string{"not create", "did not create", "no new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate candidate and no-create outcome")
	}
	databasePass := existingCount == 1 && candidateCount == 0 && sourceCount == 1
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisDuplicateExistingPath, documentThisDuplicateCandidatePath},
	}, nil
}

func verifyDocumentThisExistingUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, documentThisUpdateTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	decoyBody, decoyFound, err := documentBodyByPath(ctx, paths, documentThisUpdateDecoyPath)
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"## Decisions",
		"Use strict runner JSON for document-this intake.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisUpdateTargetPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	if decoyFound && strings.Contains(decoyBody, "Use strict runner JSON for document-this intake.") {
		failures = append(failures, "updated decoy "+documentThisUpdateDecoyPath)
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list update candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not inspect existing target before update")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisUpdateTargetPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "appended", "replaced"}) &&
		messageContainsAny(finalMessage, []string{"decoy", "not update", "did not update", "target"})
	if !assistantPass {
		failures = append(failures, "final answer did not report target update and decoy avoidance")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && (!decoyFound || !strings.Contains(decoyBody, "Use strict runner JSON for document-this intake."))
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisUpdateTargetPath, documentThisUpdateDecoyPath},
	}, nil
}

func verifyDocumentThisSynthesisFreshness(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, documentThisSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, documentThisSynthesisDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current document-this intake guidance: update existing synthesis after source, duplicate, provenance, and freshness checks.",
		"## Sources",
		"## Freshness",
	}
	expectedRefs := []string{documentThisArticlePath, documentThisDocsPath, documentThisPaperPath, documentThisTranscriptPath}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, expectedRefs)...)
	if duplicateCount != 0 {
		failures = append(failures, "created duplicate synthesis "+documentThisSynthesisDuplicatePath)
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search source evidence")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not inspect existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection_states")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance_events")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"freshness", "projection", "fresh"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source refs", "source_refs"}) &&
		messageContainsAny(finalMessage, []string{"no duplicate", "did not create", "not create"})
	if !assistantPass {
		failures = append(failures, "final answer did not report synthesis update, freshness/provenance, and duplicate avoidance")
	}
	databasePass := found &&
		duplicateCount == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, expectedRefs)) == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents: append([]string{
			documentThisSynthesisPath,
			documentThisSynthesisDuplicatePath,
		}, expectedRefs...),
	}, nil
}
