package main

import (
	"context"
	"fmt"
	"strings"
)

func verifyCaptureDocumentLinksNatural(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	counts := map[string]int{}
	for _, path := range []string{
		captureDocumentLinksSourcePath,
		captureDocumentLinksSecondSourcePath,
		captureDocumentLinksSynthesisPath,
	} {
		count, err := exactDocumentCount(ctx, paths, path)
		if err != nil {
			return verificationResult{}, err
		}
		counts[path] = count
	}
	failures := captureDocumentLinksBypassFailures(turnMetrics)
	created := []string{}
	for path, count := range counts {
		if count != 0 {
			created = append(created, path)
		}
	}
	if len(created) != 0 {
		failures = append(failures, "created durable document-these-links candidates: "+strings.Join(created, ", "))
	}
	if turnMetrics.ValidateUsed || turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "validated, wrote, or ingested before source paths and synthesis placement were approved")
	}
	assistantPass := messageContainsAll(finalMessage, []string{
		captureDocumentLinksSourcePath,
		captureDocumentLinksSecondSourcePath,
		captureDocumentLinksSynthesisPath,
	}) &&
		messageContainsAny(finalMessage, []string{"no source", "no document", "not created", "did not create", "no durable", "no write"}) &&
		messageContainsAny(finalMessage, []string{"approval", "approve", "confirm"}) &&
		messageContainsAny(finalMessage, []string{"source.path_hint", "source path", "source paths"})
	if !assistantPass {
		failures = append(failures, "final answer did not propose source paths and synthesis placement with no-write approval boundary")
	}
	activityPass := len(captureDocumentLinksBypassFailures(turnMetrics)) == 0 &&
		!turnMetrics.ValidateUsed &&
		!turnMetrics.CreateDocumentUsed &&
		!turnMetrics.AppendDocumentUsed &&
		!turnMetrics.ReplaceSectionUsed &&
		!turnMetrics.IngestSourceURLUsed &&
		!turnMetrics.IngestVideoURLUsed
	databasePass := len(created) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{captureDocumentLinksSourcePath, captureDocumentLinksSecondSourcePath, captureDocumentLinksSynthesisPath},
	}, nil
}

func verifyCaptureDocumentLinksFetch(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, captureDocumentLinksSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := captureDocumentLinksBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing document-these-links source document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{"source_type: web", "source_url:", webURLInitialText})...)
		if doc.Metadata["source_type"] != "web" {
			failures = append(failures, fmt.Sprintf("expected source_type web, got %q", doc.Metadata["source_type"]))
		}
		if doc.Metadata["asset_path"] != "" {
			failures = append(failures, "web source recorded asset_path metadata")
		}
	}
	if !turnMetrics.IngestSourceURLUsed || turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use create-mode ingest_source_url")
	}
	if turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "agent used an unsupported durable write action")
	}
	assistantPass := messageContainsAll(finalMessage, []string{captureDocumentLinksSourcePath}) &&
		messageContainsAny(finalMessage, []string{"source_type web", "source_type: web", "web"}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"}) &&
		messageContainsAny(finalMessage, []string{"path_hint", "approved", "source.path_hint"})
	if !assistantPass {
		failures = append(failures, "final answer did not report approved source.path_hint web ingestion with citation evidence")
	}
	databasePass := found && doc != nil &&
		doc.Metadata["source_type"] == "web" &&
		doc.Metadata["asset_path"] == "" &&
		len(missingRequired(doc.Body, []string{"source_type: web", "source_url:", webURLInitialText})) == 0
	activityPass := len(captureDocumentLinksBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUsed &&
		!turnMetrics.IngestSourceURLUpdateUsed &&
		!turnMetrics.CreateDocumentUsed &&
		!turnMetrics.AppendDocumentUsed &&
		!turnMetrics.ReplaceSectionUsed &&
		!turnMetrics.IngestVideoURLUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{captureDocumentLinksSourcePath},
	}, nil
}

func verifyCaptureDocumentLinksSynthesis(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	firstDoc, firstFound, err := documentByPath(ctx, paths, captureDocumentLinksSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	secondDoc, secondFound, err := documentByPath(ctx, paths, captureDocumentLinksSecondSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := exactDocumentCount(ctx, paths, captureDocumentLinksSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := captureDocumentLinksBypassFailures(turnMetrics)
	if !firstFound || firstDoc == nil {
		failures = append(failures, "missing seeded source "+captureDocumentLinksSourcePath)
	}
	if !secondFound || secondDoc == nil {
		failures = append(failures, "missing seeded source "+captureDocumentLinksSecondSourcePath)
	}
	if synthesisCount != 0 {
		failures = append(failures, "created synthesis before approval")
	}
	if !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchPathPrefixes, []string{captureDocumentLinksSourcePrefix}) {
		failures = append(failures, "did not search runner-visible document-these-links source evidence")
	}
	if !turnMetrics.ListDocumentsUsed || !containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{captureDocumentLinksSourcePrefix}) {
		failures = append(failures, "did not list document-these-links source candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "did not inspect source documents before proposing synthesis placement")
	}
	if firstDoc != nil && !containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{firstDoc.DocID}) {
		failures = append(failures, "did not inspect first source document")
	}
	if secondDoc != nil && !containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{secondDoc.DocID}) {
		failures = append(failures, "did not inspect second source document")
	}
	if !turnMetrics.ValidateUsed || turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "did not validate synthesis proposal or wrote/ingested before approval")
	}
	assistantPass := messageContainsAll(finalMessage, []string{
		captureDocumentLinksSourcePath,
		captureDocumentLinksSecondSourcePath,
		captureDocumentLinksSynthesisPath,
	}) &&
		messageContainsAny(finalMessage, []string{"validation passed", "validated"}) &&
		messageContainsAny(finalMessage, []string{"no synthesis document was created", "no document was created", "not created", "did not create"}) &&
		messageContainsAny(finalMessage, []string{"approval", "approve", "confirm"})
	if !assistantPass {
		failures = append(failures, "final answer did not report synthesis placement proposal, validation, no-write outcome, and approval boundary")
	}
	databasePass := firstFound && firstDoc != nil && secondFound && secondDoc != nil && synthesisCount == 0
	activityPass := len(captureDocumentLinksBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ValidateUsed &&
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
		Documents:     []string{captureDocumentLinksSourcePath, captureDocumentLinksSecondSourcePath, captureDocumentLinksSynthesisPath},
	}, nil
}

func verifyCaptureDocumentLinksDuplicate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	sourceDoc, sourceFound, err := documentByPath(ctx, paths, captureDocumentLinksDuplicateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDoc, synthesisFound, err := documentByPath(ctx, paths, captureDocumentLinksSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := exactDocumentCount(ctx, paths, captureDocumentLinksDuplicateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, captureDocumentLinksDuplicateCandidate)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := exactDocumentCount(ctx, paths, captureDocumentLinksSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateSynthesisCount, err := exactDocumentCount(ctx, paths, captureDocumentLinksDuplicateSynthesis)
	if err != nil {
		return verificationResult{}, err
	}
	failures := captureDocumentLinksBypassFailures(turnMetrics)
	if !sourceFound || sourceDoc == nil || sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing source %s, got %d", captureDocumentLinksDuplicateSourcePath, sourceCount))
	}
	if !synthesisFound || synthesisDoc == nil || synthesisCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing synthesis %s, got %d", captureDocumentLinksSynthesisPath, synthesisCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created forbidden duplicate source "+captureDocumentLinksDuplicateCandidate)
	}
	if duplicateSynthesisCount != 0 {
		failures = append(failures, "created forbidden duplicate synthesis "+captureDocumentLinksDuplicateSynthesis)
	}
	if !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchPathPrefixes, []string{captureDocumentLinksSourcePrefix}) {
		failures = append(failures, "did not search runner-visible duplicate source evidence")
	}
	if !turnMetrics.ListDocumentsUsed || !containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{captureDocumentLinksSourcePrefix, "synthesis/"}) {
		failures = append(failures, "did not list source and synthesis placement candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "did not inspect existing source and synthesis candidates")
	}
	if sourceDoc != nil && !containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{sourceDoc.DocID}) {
		failures = append(failures, "did not inspect existing duplicate source target")
	}
	if synthesisDoc != nil && !containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{synthesisDoc.DocID}) {
		failures = append(failures, "did not inspect existing synthesis target")
	}
	if turnMetrics.ValidateUsed || turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "validated, wrote, or ingested while update versus new placement was unresolved")
	}
	assistantPass := messageContainsAll(finalMessage, []string{
		captureDocumentLinksDuplicateSourcePath,
		captureDocumentLinksSynthesisPath,
	}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already", "likely"}) &&
		messageContainsAny(finalMessage, []string{"no source", "no synthesis", "no document", "not created", "did not create", "no write", "not updated"}) &&
		messageContainsAny(finalMessage, []string{"update", "new confirmed", "confirmed path", "create new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate source/synthesis targets, no-write outcome, and update-versus-new placement clarification")
	}
	databasePass := sourceFound && sourceDoc != nil && sourceCount == 1 &&
		synthesisFound && synthesisDoc != nil && synthesisCount == 1 &&
		candidateCount == 0 &&
		duplicateSynthesisCount == 0
	activityPass := len(captureDocumentLinksBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
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
		Documents:     []string{captureDocumentLinksDuplicateSourcePath, captureDocumentLinksDuplicateCandidate, captureDocumentLinksSynthesisPath, captureDocumentLinksDuplicateSynthesis},
	}, nil
}
