package main

import (
	"context"
	"fmt"
)

func verifyCaptureDuplicateCandidate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, requireTargetAccuracy bool) (verificationResult, error) {
	existingDoc, existingFound, err := documentByPath(ctx, paths, captureDuplicateCandidateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	existingCount, err := exactDocumentCount(ctx, paths, captureDuplicateCandidateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	decoyCount, err := exactDocumentCount(ctx, paths, captureDuplicateCandidateDecoyPath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, captureDuplicateCandidateNewPath)
	if err != nil {
		return verificationResult{}, err
	}
	prefixCount, err := documentCountWithPrefix(ctx, paths, captureDuplicateCandidatePrefix)
	if err != nil {
		return verificationResult{}, err
	}

	failures := captureDuplicateCandidateBypassFailures(turnMetrics)
	if !existingFound {
		failures = append(failures, "missing seeded duplicate candidate target "+captureDuplicateCandidateExistingPath)
	}
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing duplicate candidate target %s, got %d", captureDuplicateCandidateExistingPath, existingCount))
	}
	if decoyCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one target-accuracy decoy %s, got %d", captureDuplicateCandidateDecoyPath, decoyCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created forbidden duplicate candidate "+captureDuplicateCandidateNewPath)
	}
	if prefixCount != 2 {
		failures = append(failures, fmt.Sprintf("expected only seeded duplicate candidate docs under %s, got %d", captureDuplicateCandidatePrefix, prefixCount))
	}
	if existingDoc != nil && existingDoc.Title != captureDuplicateCandidateExistingTitle {
		failures = append(failures, fmt.Sprintf("expected existing target title %q, got %q", captureDuplicateCandidateExistingTitle, existingDoc.Title))
	}
	searchedDuplicatePrefix := containsAllStrings(turnMetrics.SearchPathPrefixes, []string{captureDuplicateCandidatePrefix})
	listedDuplicatePrefix := containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{captureDuplicateCandidatePrefix})
	inspectedExistingTarget := existingDoc != nil && containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{existingDoc.DocID})
	if !turnMetrics.SearchUsed {
		failures = append(failures, "did not search for runner-visible duplicate candidate evidence")
	} else if !searchedDuplicatePrefix {
		failures = append(failures, "search did not use duplicate candidate path prefix "+captureDuplicateCandidatePrefix)
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "did not list duplicate candidate paths")
	} else if !listedDuplicatePrefix {
		failures = append(failures, "list_documents did not use duplicate candidate path prefix "+captureDuplicateCandidatePrefix)
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "did not inspect the runner-visible duplicate candidate")
	} else if !inspectedExistingTarget {
		failures = append(failures, "did not inspect the existing duplicate candidate target")
	}
	if turnMetrics.ValidateUsed || turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "validated, wrote, or ingested while update versus new path was unresolved")
	}

	assistantPass := messageContainsAll(finalMessage, []string{captureDuplicateCandidateExistingPath, captureDuplicateCandidateExistingTitle}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already", "likely"}) &&
		messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "no write", "no document was created or updated"}) &&
		messageContainsAny(finalMessage, []string{"not updated", "did not update", "no document was created or updated", "no write"}) &&
		messageContainsAny(finalMessage, []string{"update", "new document", "confirmed path", "create a new"})
	if requireTargetAccuracy {
		assistantPass = assistantPass &&
			messageContainsAny(finalMessage, []string{"target accuracy", "correct target", "likely duplicate candidate"}) &&
			!messageContainsAny(finalMessage, []string{captureDuplicateCandidateDecoyPath, captureDuplicateCandidateDecoyTitle})
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report the correct duplicate candidate, no-write outcome, and update-versus-new clarification")
	}

	databasePass := existingFound &&
		existingCount == 1 &&
		decoyCount == 1 &&
		candidateCount == 0 &&
		prefixCount == 2 &&
		existingDoc != nil &&
		existingDoc.Title == captureDuplicateCandidateExistingTitle
	activityPass := len(captureDuplicateCandidateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		searchedDuplicatePrefix &&
		turnMetrics.ListDocumentsUsed &&
		listedDuplicatePrefix &&
		turnMetrics.GetDocumentUsed &&
		inspectedExistingTarget &&
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
		Documents:     []string{captureDuplicateCandidateExistingPath, captureDuplicateCandidateDecoyPath, captureDuplicateCandidateNewPath},
	}, nil
}
