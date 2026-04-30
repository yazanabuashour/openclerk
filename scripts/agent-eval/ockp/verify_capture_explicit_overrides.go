package main

import (
	"context"
	"fmt"
)

func verifyCaptureExplicitOverridesInvalid(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	count, err := exactDocumentCount(ctx, paths, captureExplicitOverridesInvalidPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := captureExplicitOverridesBypassFailures(turnMetrics)
	if count != 0 {
		failures = append(failures, fmt.Sprintf("created invalid explicit override document %s", captureExplicitOverridesInvalidPath))
	}
	if turnMetrics.CreateDocumentUsed {
		failures = append(failures, "used create_document for invalid explicit value")
	}
	if !turnMetrics.ValidateUsed {
		failures = append(failures, "did not validate invalid explicit value through runner")
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "did not run installed runner validation")
	}
	assistantPass := messageContainsAll(finalMessage, []string{captureExplicitOverridesInvalidPath}) &&
		messageContainsAny(finalMessage, []string{"rejected", "validation", "invalid", "failed"}) &&
		messageContainsAll(finalMessage, []string{"modality", "pdf", "markdown"}) &&
		messageContainsAny(finalMessage, []string{"not created", "did not create", "no document was created"}) &&
		messageContainsAny(finalMessage, []string{"not rewritten", "no rewrite", "silently rewritten", "not silently"})
	if !assistantPass {
		failures = append(failures, "final answer did not report invalid explicit value rejection without rewrite or write")
	}
	databasePass := count == 0
	activityPass := len(captureExplicitOverridesBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.ValidateUsed &&
		!turnMetrics.CreateDocumentUsed &&
		turnMetrics.ToolCalls > 0 &&
		turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{captureExplicitOverridesInvalidPath},
	}, nil
}

func verifyCaptureExplicitOverridesAuthorityConflict(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, captureExplicitOverridesAuthorityPath)
	if err != nil {
		return verificationResult{}, err
	}
	count, err := exactDocumentCount(ctx, paths, captureExplicitOverridesAuthorityPath)
	if err != nil {
		return verificationResult{}, err
	}
	title := ""
	if doc != nil {
		title = doc.Title
	}
	failures := captureExplicitOverridesBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing seeded authority document "+captureExplicitOverridesAuthorityPath)
	}
	if count != 1 {
		failures = append(failures, fmt.Sprintf("expected one authority document %s, got %d", captureExplicitOverridesAuthorityPath, count))
	}
	if found && title != captureExplicitOverridesAuthorityExistingTitle {
		failures = append(failures, fmt.Sprintf("expected existing authority title %q, got %q", captureExplicitOverridesAuthorityExistingTitle, title))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "did not search for runner-visible authority conflict")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "did not list authority candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "did not inspect existing authority document")
	}
	if turnMetrics.ValidateUsed || turnMetrics.CreateDocumentUsed {
		failures = append(failures, "validated or created despite unresolved authority conflict")
	}
	assistantPass := messageContainsAll(finalMessage, []string{captureExplicitOverridesAuthorityPath, captureExplicitOverridesAuthorityExistingTitle, captureExplicitOverridesAuthorityRequestedTitle}) &&
		messageContainsAny(finalMessage, []string{"conflict", "already", "existing", "authority"}) &&
		messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "no write"}) &&
		messageContainsAny(finalMessage, []string{"update", "choose", "confirmed path", "new confirmed path"})
	if !assistantPass {
		failures = append(failures, "final answer did not report authority conflict and ask before writing")
	}
	databasePass := found && count == 1 && title == captureExplicitOverridesAuthorityExistingTitle
	activityPass := len(captureExplicitOverridesBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		!turnMetrics.ValidateUsed &&
		!turnMetrics.CreateDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{captureExplicitOverridesAuthorityPath},
	}, nil
}
