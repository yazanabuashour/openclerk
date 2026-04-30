package main

import (
	"context"
	"fmt"
)

func verifyCaptureLowRiskDuplicate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingDoc, existingFound, err := documentByPath(ctx, paths, captureLowRiskDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	existingCount, err := exactDocumentCount(ctx, paths, captureLowRiskDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, captureLowRiskCandidateDuplicate)
	if err != nil {
		return verificationResult{}, err
	}
	prefixCount, err := documentCountWithPrefix(ctx, paths, captureLowRiskDuplicatePrefix)
	if err != nil {
		return verificationResult{}, err
	}

	failures := captureLowRiskBypassFailures(turnMetrics)
	if !existingFound {
		failures = append(failures, "missing seeded low-risk duplicate target "+captureLowRiskDuplicatePath)
	}
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one low-risk duplicate target %s, got %d", captureLowRiskDuplicatePath, existingCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created forbidden low-risk duplicate candidate "+captureLowRiskCandidateDuplicate)
	}
	if prefixCount != 1 {
		failures = append(failures, fmt.Sprintf("expected only seeded low-risk duplicate target under %s, got %d", captureLowRiskDuplicatePrefix, prefixCount))
	}
	if existingDoc != nil && existingDoc.Title != captureLowRiskDuplicateTitle {
		failures = append(failures, fmt.Sprintf("expected existing target title %q, got %q", captureLowRiskDuplicateTitle, existingDoc.Title))
	}
	searchedPrefix := containsAllStrings(turnMetrics.SearchPathPrefixes, []string{captureLowRiskDuplicatePrefix})
	listedPrefix := containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{captureLowRiskDuplicatePrefix})
	inspectedTarget := existingDoc != nil && containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{existingDoc.DocID})
	if !turnMetrics.SearchUsed {
		failures = append(failures, "did not search for runner-visible low-risk duplicate evidence")
	} else if !searchedPrefix {
		failures = append(failures, "search did not use low-risk path prefix "+captureLowRiskDuplicatePrefix)
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "did not list low-risk candidate paths")
	} else if !listedPrefix {
		failures = append(failures, "list_documents did not use low-risk path prefix "+captureLowRiskDuplicatePrefix)
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "did not inspect the runner-visible low-risk duplicate target")
	} else if !inspectedTarget {
		failures = append(failures, "did not inspect the existing low-risk duplicate target")
	}
	if turnMetrics.ValidateUsed || turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "validated, wrote, or ingested while update versus new path was unresolved")
	}

	assistantPass := messageContainsAll(finalMessage, []string{captureLowRiskDuplicatePath, captureLowRiskDuplicateTitle}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already", "likely"}) &&
		messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "no write", "no document was created or updated"}) &&
		messageContainsAny(finalMessage, []string{"not updated", "did not update", "no document was created or updated", "no write"}) &&
		messageContainsAny(finalMessage, []string{"update", "new document", "confirmed path", "create a new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report the low-risk duplicate target, no-write outcome, and update-versus-new clarification")
	}

	databasePass := existingFound &&
		existingCount == 1 &&
		candidateCount == 0 &&
		prefixCount == 1 &&
		existingDoc != nil &&
		existingDoc.Title == captureLowRiskDuplicateTitle
	activityPass := len(captureLowRiskBypassFailures(turnMetrics)) == 0 &&
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
		Documents:     []string{captureLowRiskDuplicatePath, captureLowRiskCandidateDuplicate},
	}, nil
}
