package main

import (
	"context"
	"fmt"
	"strings"
)

func verifyWebURLCreate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, webURLSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := webURLBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing web URL source document")
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
	assistantPass := messageContainsAll(finalMessage, []string{webURLSourcePath}) &&
		messageContainsAny(finalMessage, []string{"source_type web", "source_type: web", "web"}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"})
	if !assistantPass {
		failures = append(failures, "final answer did not report web source ingestion with citation evidence")
	}
	databasePass := found && doc != nil &&
		doc.Metadata["source_type"] == "web" &&
		doc.Metadata["asset_path"] == "" &&
		len(missingRequired(doc.Body, []string{"source_type: web", "source_url:", webURLInitialText})) == 0
	activityPass := len(webURLBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUsed && !turnMetrics.IngestSourceURLUpdateUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webURLSourcePath},
	}, nil
}

func verifyWebURLDuplicate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	sourceCount, err := exactDocumentCount(ctx, paths, webURLSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, webURLDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := webURLBypassFailures(turnMetrics)
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing web source, got %d", sourceCount))
	}
	if duplicateCount != 0 {
		failures = append(failures, "created duplicate web source "+webURLDuplicatePath)
	}
	if !turnMetrics.IngestSourceURLUsed {
		failures = append(failures, "agent did not exercise duplicate ingest_source_url request")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list web URL source candidates after duplicate rejection")
	}
	assistantPass := messageContainsAll(finalMessage, []string{webURLSourcePath, webURLDuplicatePath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "already", "rejected"}) &&
		messageContainsAny(finalMessage, []string{"not created", "was not created", "no copy", "no durable"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate rejection and no-copy outcome")
	}
	databasePass := sourceCount == 1 && duplicateCount == 0
	activityPass := len(webURLBypassFailures(turnMetrics)) == 0 && turnMetrics.IngestSourceURLUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webURLSourcePath, webURLDuplicatePath},
	}, nil
}

func verifyWebURLSameHash(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, webURLSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := webURLBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing web URL source document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{webURLInitialText})...)
		if strings.Contains(doc.Body, webURLChangedText) {
			failures = append(failures, "same-hash update unexpectedly changed source body")
		}
	}
	if !turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use source.mode update")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search preserved web evidence")
	}
	assistantPass := messageContainsAll(finalMessage, []string{webURLSourcePath}) &&
		messageContainsAny(finalMessage, []string{"same-hash", "no-op", "no changed", "preserved"}) &&
		messageContainsAny(finalMessage, []string{"citation", "doc_id", "chunk_id"})
	if !assistantPass {
		failures = append(failures, "final answer did not report same-hash no-op and preserved citations")
	}
	databasePass := found && doc != nil &&
		strings.Contains(doc.Body, webURLInitialText) &&
		!strings.Contains(doc.Body, webURLChangedText)
	activityPass := len(webURLBypassFailures(turnMetrics)) == 0 && turnMetrics.IngestSourceURLUpdateUsed && turnMetrics.SearchUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webURLSourcePath},
	}, nil
}

func verifyWebURLChanged(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, webURLSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesis, synthesisFound, err := documentByPath(ctx, paths, webURLSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docIDOrEmpty(synthesis))
	if err != nil {
		return verificationResult{}, err
	}
	failures := webURLBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing web URL source document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{webURLChangedText})...)
	}
	if !synthesisFound || synthesis == nil {
		failures = append(failures, "missing web URL synthesis document")
	}
	if projection == nil || projection.Freshness != "stale" || !strings.Contains(projection.Details["stale_source_refs"], webURLSourcePath) {
		failures = append(failures, fmt.Sprintf("expected stale synthesis projection for %s, got %+v", webURLSourcePath, projection))
	}
	if !turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use source.mode update")
	}
	if !turnMetrics.SearchUsed || !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not search changed evidence and inspect projection_states")
	}
	assistantPass := messageContainsAll(finalMessage, []string{webURLSourcePath, webURLSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"changed", "refreshed"}) &&
		messageContainsAny(finalMessage, []string{"stale", "projection"})
	if !assistantPass {
		failures = append(failures, "final answer did not report changed web update and stale synthesis projection")
	}
	databasePass := found && doc != nil && synthesisFound && synthesis != nil &&
		strings.Contains(doc.Body, webURLChangedText) &&
		projection != nil && projection.Freshness == "stale" &&
		strings.Contains(projection.Details["stale_source_refs"], webURLSourcePath)
	activityPass := len(webURLBypassFailures(turnMetrics)) == 0 && turnMetrics.IngestSourceURLUpdateUsed && turnMetrics.SearchUsed && turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webURLSourcePath, webURLSynthesisPath},
	}, nil
}

func verifyWebURLStaleRepair(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, scripted bool) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, webURLSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceDocID, sourceDocIDFound, err := documentIDByPath(ctx, paths, webURLSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := exactDocumentCount(ctx, paths, webURLSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, webURLDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesis, synthesisFound, err := documentByPath(ctx, paths, webURLSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docIDOrEmpty(synthesis))
	if err != nil {
		return verificationResult{}, err
	}

	failures := webURLBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing web URL source document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{"source_type: web", "source_url:", webURLChangedText})...)
		if strings.Contains(doc.Body, webURLInitialText) && !strings.Contains(doc.Body, webURLChangedText) {
			failures = append(failures, "web URL source still contains only initial evidence")
		}
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one stable web URL source, got %d", sourceCount))
	}
	if duplicateCount != 0 {
		failures = append(failures, "created duplicate web source "+webURLDuplicatePath)
	}
	if !sourceDocIDFound || sourceDocID == "" {
		failures = append(failures, "missing web URL source doc_id")
	}
	if !synthesisFound || synthesis == nil {
		failures = append(failures, "missing web URL synthesis document")
	}
	if projection == nil || projection.Freshness != "stale" || !strings.Contains(projection.Details["stale_source_refs"], webURLSourcePath) {
		failures = append(failures, fmt.Sprintf("expected stale synthesis projection for %s, got %+v", webURLSourcePath, projection))
	}
	duplicateCreateChecked := !scripted || turnMetrics.IngestSourceURLCreateUsed && stringValuesInclude(turnMetrics.IngestSourceURLPathHints, webURLDuplicatePath)
	secondUpdateChecked := !scripted || turnMetrics.IngestSourceURLUpdateCount >= 2
	if !turnMetrics.IngestSourceURLUsed || !turnMetrics.IngestSourceURLUpdateUsed || !duplicateCreateChecked || !secondUpdateChecked {
		failures = append(failures, "agent did not exercise ingest_source_url duplicate create, update mode, and second no-op update")
	}
	if !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not search/list/get runner-visible stale repair evidence")
	}
	expectedProvenanceRefs := []string{}
	if sourceDocIDFound && sourceDocID != "" {
		expectedProvenanceRefs = append(expectedProvenanceRefs, sourceDocID)
	}
	if synthesisFound && synthesis != nil && synthesis.DocID != "" {
		expectedProvenanceRefs = append(expectedProvenanceRefs, "synthesis:"+synthesis.DocID)
	}
	inspectedExpectedProvenanceRefs := len(expectedProvenanceRefs) > 0 && provenanceEventRefIDsInclude(turnMetrics.ProvenanceEventRefIDs, expectedProvenanceRefs...)
	if !turnMetrics.ProjectionStatesUsed || !turnMetrics.ProvenanceEventsUsed || !inspectedExpectedProvenanceRefs {
		failures = append(failures, "agent did not inspect projection_states and provenance_events for the source and synthesis projection refs")
	}
	if scripted && !messageContainsAny(finalMessage, []string{"same-hash", "no-op", "no op", "unchanged"}) {
		failures = append(failures, "scripted final answer did not report the same-hash/no-op boundary")
	}
	if scripted && !messageContainsAny(finalMessage, []string{"duplicate", "already", "rejected", "normalized"}) {
		failures = append(failures, "scripted final answer did not report duplicate rejection")
	}
	if scripted && !messageContainsAny(finalMessage, []string{"not created", "was not created", "no copy", "no duplicate"}) {
		failures = append(failures, "scripted final answer did not report no duplicate source was created")
	}

	assistantPass := messageContainsAll(finalMessage, []string{webURLSourcePath, webURLSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"changed", "refreshed", "updated"}) &&
		messageContainsAny(finalMessage, []string{webURLChangedText, "changed evidence"}) &&
		messageContainsAny(finalMessage, []string{"stale", "projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source_updated", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"no browser", "without browser", "manual acquisition", "no manual", "runner-owned", "runner owned"})
	if !assistantPass {
		failures = append(failures, "final answer did not report changed web update, stale synthesis impact, provenance/freshness, and no-browser/no-manual boundaries")
	}

	databasePass := found && doc != nil && sourceCount == 1 && duplicateCount == 0 &&
		synthesisFound && synthesis != nil &&
		strings.Contains(doc.Body, webURLChangedText) &&
		projection != nil && projection.Freshness == "stale" &&
		strings.Contains(projection.Details["stale_source_refs"], webURLSourcePath)
	activityPass := len(webURLBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUsed && turnMetrics.IngestSourceURLUpdateUsed &&
		duplicateCreateChecked && secondUpdateChecked &&
		turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed && turnMetrics.ProvenanceEventsUsed && inspectedExpectedProvenanceRefs
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webURLSourcePath, webURLDuplicatePath, webURLSynthesisPath},
	}, nil
}

func verifyWebURLUnsupported(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	count, err := exactDocumentCount(ctx, paths, "sources/web-url/unsupported.md")
	if err != nil {
		return verificationResult{}, err
	}
	failures := webURLBypassFailures(turnMetrics)
	if count != 0 {
		failures = append(failures, "created unsupported web source")
	}
	if !turnMetrics.IngestSourceURLUsed {
		failures = append(failures, "agent did not exercise unsupported ingest_source_url request")
	}
	assistantPass := messageContainsAny(finalMessage, []string{"unsupported", "non-html", "HTML", "content type", "rejected"}) &&
		messageContainsAny(finalMessage, []string{"not created", "no durable", "was not created"})
	if !assistantPass {
		failures = append(failures, "final answer did not report unsupported rejection without write")
	}
	databasePass := count == 0
	activityPass := len(webURLBypassFailures(turnMetrics)) == 0 && turnMetrics.IngestSourceURLUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{"sources/web-url/unsupported.md"},
	}, nil
}

func webURLBypassFailures(turnMetrics metrics) []string {
	return artifactIngestionBypassFailures(turnMetrics)
}
