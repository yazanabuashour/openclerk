package main

import (
	"context"
	"fmt"
	"strings"
)

func verifyCaptureSaveThisNoteDuplicate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingDoc, existingFound, err := documentByPath(ctx, paths, captureSaveThisNoteDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	existingCount, err := exactDocumentCount(ctx, paths, captureSaveThisNoteDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, captureSaveThisNoteCandidateDuplicate)
	if err != nil {
		return verificationResult{}, err
	}
	prefixCount, err := documentCountWithPrefix(ctx, paths, captureSaveThisNoteDuplicatePrefix)
	if err != nil {
		return verificationResult{}, err
	}

	failures := captureSaveThisNoteBypassFailures(turnMetrics)
	if !existingFound {
		failures = append(failures, "missing seeded save-this-note duplicate target "+captureSaveThisNoteDuplicatePath)
	}
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one save-this-note duplicate target %s, got %d", captureSaveThisNoteDuplicatePath, existingCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created forbidden save-this-note duplicate candidate "+captureSaveThisNoteCandidateDuplicate)
	}
	if prefixCount != 1 {
		failures = append(failures, fmt.Sprintf("expected only seeded save-this-note duplicate target under %s, got %d", captureSaveThisNoteDuplicatePrefix, prefixCount))
	}
	if existingDoc != nil && existingDoc.Title != captureSaveThisNoteDuplicateTitle {
		failures = append(failures, fmt.Sprintf("expected existing target title %q, got %q", captureSaveThisNoteDuplicateTitle, existingDoc.Title))
	}
	searchedPrefix := containsAllStrings(turnMetrics.SearchPathPrefixes, []string{captureSaveThisNoteDuplicatePrefix})
	listedPrefix := containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{captureSaveThisNoteDuplicatePrefix})
	inspectedTarget := existingDoc != nil && containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{existingDoc.DocID})
	if !turnMetrics.SearchUsed {
		failures = append(failures, "did not search for runner-visible save-this-note duplicate evidence")
	} else if !searchedPrefix {
		failures = append(failures, "search did not use save-this-note path prefix "+captureSaveThisNoteDuplicatePrefix)
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "did not list save-this-note candidate paths")
	} else if !listedPrefix {
		failures = append(failures, "list_documents did not use save-this-note path prefix "+captureSaveThisNoteDuplicatePrefix)
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "did not inspect the runner-visible save-this-note duplicate target")
	} else if !inspectedTarget {
		failures = append(failures, "did not inspect the existing save-this-note duplicate target")
	}
	if turnMetrics.ValidateUsed || turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "validated, wrote, or ingested while update versus new path was unresolved")
	}

	assistantPass := messageContainsAll(finalMessage, []string{captureSaveThisNoteDuplicatePath, captureSaveThisNoteDuplicateTitle}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already", "likely"}) &&
		messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "no write", "no document was created or updated"}) &&
		messageContainsAny(finalMessage, []string{"not updated", "did not update", "no document was created or updated", "no write"}) &&
		messageContainsAny(finalMessage, []string{"update", "new document", "confirmed path", "create a new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report the duplicate target, no-write outcome, and update-versus-new clarification")
	}

	databasePass := existingFound &&
		existingCount == 1 &&
		candidateCount == 0 &&
		prefixCount == 1 &&
		existingDoc != nil &&
		existingDoc.Title == captureSaveThisNoteDuplicateTitle
	activityPass := len(captureSaveThisNoteBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		searchedPrefix &&
		turnMetrics.ListDocumentsUsed &&
		listedPrefix &&
		turnMetrics.GetDocumentUsed &&
		inspectedTarget &&
		!turnMetrics.ValidateUsed &&
		!turnMetrics.CreateDocumentUsed &&
		!turnMetrics.AppendDocumentUsed &&
		!turnMetrics.ReplaceSectionUsed &&
		!turnMetrics.IngestSourceURLUsed &&
		!turnMetrics.IngestVideoURLUsed

	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{captureSaveThisNoteDuplicatePath, captureSaveThisNoteCandidateDuplicate},
	}, nil
}

func verifyCaptureSaveThisNoteLowConfidence(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	counts := []string{captureSaveThisNoteNaturalPath, captureSaveThisNotePath, captureSaveThisNoteDuplicatePath, captureSaveThisNoteCandidateDuplicate}
	created := []string{}
	for _, path := range counts {
		count, err := exactDocumentCount(ctx, paths, path)
		if err != nil {
			return verificationResult{}, err
		}
		if count != 0 {
			created = append(created, path)
		}
	}
	failures := captureSaveThisNoteBypassFailures(turnMetrics)
	if len(created) != 0 {
		failures = append(failures, "created low-confidence save-this-note documents: "+strings.Join(created, ", "))
	}
	if turnMetrics.ToolCalls != 0 || turnMetrics.CommandExecutions != 0 || turnMetrics.AssistantCalls > 1 {
		failures = append(failures, fmt.Sprintf("low-confidence save-this-note ask should be no-tools, got tools=%d commands=%d assistant=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions, turnMetrics.AssistantCalls))
	}
	assistantPass := messageContainsAny(finalMessage, []string{"body", "content", "text", "actual note", "what to save"}) &&
		messageContainsAny(finalMessage, []string{"missing", "provide", "need", "can't create", "cannot create"}) &&
		messageContainsAny(finalMessage, []string{"path", "title", "placement", "where"}) &&
		!messageContainsAny(finalMessage, []string{captureSaveThisNotePath, captureSaveThisNoteDuplicatePath, captureSaveThisNoteCandidateDuplicate})
	if !assistantPass {
		failures = append(failures, "final answer did not ask for missing note content and durable placement without proposing a path")
	}
	databasePass := len(created) == 0
	activityPass := len(captureSaveThisNoteBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.ToolCalls == 0 &&
		turnMetrics.CommandExecutions == 0 &&
		turnMetrics.AssistantCalls <= 1
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     counts,
	}, nil
}
