package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

type sourceLinkedSynthesisExpectations struct {
	SourceRefs                 []string
	RequireSearch              bool
	RequireList                bool
	RequireGet                 bool
	RequireRecordsLookup       bool
	RequireProvenanceEvents    bool
	RequireProjectionStates    bool
	Metrics                    metrics
	FinalAnswerPath            bool
	AdditionalDocs             []string
	AdditionalBodyRequirements []string
}

func verifySourceLinkedSynthesis(ctx context.Context, paths evalPaths, docPath string, finalMessage string, expectations sourceLinkedSynthesisExpectations) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	documents := append([]string{}, expectations.AdditionalDocs...)
	documents = append(documents, docPath)
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"## Sources",
		"## Freshness",
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, expectations.SourceRefs)...)
	failures = append(failures, missingRequiredFold(body, expectations.AdditionalBodyRequirements)...)
	if expectations.FinalAnswerPath && !messageContainsAll(finalMessage, []string{docPath}) {
		failures = append(failures, "final answer did not mention "+docPath)
	}
	if expectations.RequireSearch && !expectations.Metrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if expectations.RequireList && !expectations.Metrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if expectations.RequireGet && !expectations.Metrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if expectations.RequireRecordsLookup && !expectations.Metrics.RecordsLookupUsed {
		failures = append(failures, "agent did not use records lookup")
	}
	if expectations.RequireProvenanceEvents && !expectations.Metrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if expectations.RequireProjectionStates && !expectations.Metrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	databaseFailures := missingRequired(body, required)
	databaseFailures = append(databaseFailures, sourceRefsFrontmatterFailures(body, expectations.SourceRefs)...)
	databaseFailures = append(databaseFailures, missingRequiredFold(body, expectations.AdditionalBodyRequirements)...)
	databasePass := found && len(databaseFailures) == 0
	assistantPass := strings.TrimSpace(finalMessage) != ""
	if expectations.FinalAnswerPath {
		assistantPass = assistantPass && messageContainsAll(finalMessage, []string{docPath})
	}
	activityPass := (!expectations.RequireSearch || expectations.Metrics.SearchUsed) &&
		(!expectations.RequireList || expectations.Metrics.ListDocumentsUsed) &&
		(!expectations.RequireGet || expectations.Metrics.GetDocumentUsed) &&
		(!expectations.RequireRecordsLookup || expectations.Metrics.RecordsLookupUsed) &&
		(!expectations.RequireProvenanceEvents || expectations.Metrics.ProvenanceEventsUsed) &&
		(!expectations.RequireProjectionStates || expectations.Metrics.ProjectionStatesUsed)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     documents,
	}, nil
}
func verifySourceURLUpdateDuplicateCreate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := sourceURLUpdateBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing original source URL document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "source_url:", "asset_path:", sourceURLUpdateAssetPath})...)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one original source document, got %d", sourceCount))
	}
	if duplicateCount != 0 {
		failures = append(failures, "duplicate create wrote "+sourceURLUpdateDuplicatePath)
	}
	if !turnMetrics.IngestSourceURLUsed || turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not attempt default create-mode source URL ingestion")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source documents after duplicate rejection")
	}
	assistantPass := messageContainsAll(finalMessage, []string{sourceURLUpdateSourcePath, sourceURLUpdateDuplicatePath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "already exists", "rejected"}) &&
		messageContainsAny(finalMessage, []string{"not created", "was not created", "no copy"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate rejection and no-write outcome")
	}
	databasePass := found && sourceCount == 1 && duplicateCount == 0 && doc != nil &&
		len(missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "source_url:", "asset_path:", sourceURLUpdateAssetPath})) == 0
	activityPass := len(sourceURLUpdateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUsed && !turnMetrics.IngestSourceURLUpdateUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceURLUpdateSourcePath},
	}, nil
}
func verifySourceURLUpdateSameSHA(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := sourceURLUpdateSourceEvents(ctx, paths, docIDOrEmpty(doc))
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDoc, synthesisFound, err := documentByPath(ctx, paths, sourceURLUpdateSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docIDOrEmpty(synthesisDoc))
	if err != nil {
		return verificationResult{}, err
	}
	projectionEvents, err := sourceURLUpdateProjectionEvents(ctx, paths, docIDOrEmpty(synthesisDoc))
	if err != nil {
		return verificationResult{}, err
	}
	search, err := sourceURLUpdateSearch(ctx, paths, sourceURLUpdateInitialText)
	if err != nil {
		return verificationResult{}, err
	}
	failures := sourceURLUpdateBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing source URL document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "asset_path:", sourceURLUpdateAssetPath})...)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one source document, got %d", sourceCount))
	}
	if eventTypesInclude(sourceEvents, "source_updated") {
		failures = append(failures, "same-SHA update emitted source_updated provenance")
	}
	if !synthesisFound || projection == nil || projection.Freshness != "fresh" {
		failures = append(failures, "same-SHA update did not leave dependent synthesis fresh")
	}
	if eventTypesInclude(projectionEvents, "projection_invalidated") {
		failures = append(failures, "same-SHA update invalidated dependent synthesis")
	}
	if !searchContainsPath(search, sourceURLUpdateSourcePath) || !searchResultHasCitations(search) {
		failures = append(failures, "same-SHA source evidence was not searchable with citations")
	}
	if !turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use source.mode update")
	}
	if !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProvenanceEventsUsed || !turnMetrics.SearchUsed || !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect source document, provenance, search evidence, and synthesis projection")
	}
	assistantPass := messageContainsAll(finalMessage, []string{sourceURLUpdateSourcePath, sourceURLUpdateSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"same-sha", "same sha", "no-op", "unchanged"}) &&
		messageContainsAny(finalMessage, []string{"citation", "source evidence", "preserved"}) &&
		messageContainsAny(finalMessage, []string{"fresh"}) &&
		messageContainsAny(finalMessage, []string{"no changed", "not changed", "no refresh", "not needed"})
	if !assistantPass {
		failures = append(failures, "final answer did not report same-SHA no-op with preserved evidence")
	}
	databasePass := found && doc != nil && sourceCount == 1 &&
		len(missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "asset_path:", sourceURLUpdateAssetPath})) == 0 &&
		!eventTypesInclude(sourceEvents, "source_updated") &&
		synthesisFound &&
		projection != nil &&
		projection.Freshness == "fresh" &&
		!eventTypesInclude(projectionEvents, "projection_invalidated") &&
		searchContainsPath(search, sourceURLUpdateSourcePath) &&
		searchResultHasCitations(search)
	activityPass := len(sourceURLUpdateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUpdateUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.SearchUsed &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceURLUpdateSourcePath, sourceURLUpdateSynthesisPath},
	}, nil
}
func verifySourceURLUpdateChangedPDF(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDoc, synthesisFound, err := documentByPath(ctx, paths, sourceURLUpdateSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := sourceURLUpdateSourceEvents(ctx, paths, docIDOrEmpty(doc))
	if err != nil {
		return verificationResult{}, err
	}
	changedSearch, err := sourceURLUpdateSearch(ctx, paths, sourceURLUpdateChangedText)
	if err != nil {
		return verificationResult{}, err
	}
	oldSearch, err := sourceURLUpdateSearch(ctx, paths, sourceURLUpdateInitialText)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docIDOrEmpty(synthesisDoc))
	if err != nil {
		return verificationResult{}, err
	}
	projectionEvents, err := sourceURLUpdateProjectionEvents(ctx, paths, docIDOrEmpty(synthesisDoc))
	if err != nil {
		return verificationResult{}, err
	}
	updateEventOK := sourceURLUpdateEventHasSHAChange(sourceEvents)
	hasStaleProjection := projection != nil &&
		projection.Freshness == "stale" &&
		projectionDetailContains(projection.Details, "stale_source_refs", sourceURLUpdateSourcePath)
	failures := sourceURLUpdateBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing updated source URL document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{sourceURLUpdateChangedText, "asset_path:", sourceURLUpdateAssetPath})...)
		failures = append(failures, presentForbidden(doc.Body, []string{sourceURLUpdateInitialText})...)
	}
	if !synthesisFound || synthesisDoc == nil {
		failures = append(failures, "missing dependent synthesis")
	} else if !strings.Contains(synthesisDoc.Body, sourceURLUpdateInitialText) {
		failures = append(failures, "dependent synthesis was repaired or no longer contains initial stale claim")
	}
	if !searchContainsPath(changedSearch, sourceURLUpdateSourcePath) || !searchResultHasCitations(changedSearch) {
		failures = append(failures, "changed source evidence was not searchable with citations")
	}
	if searchContainsPath(oldSearch, sourceURLUpdateSourcePath) {
		failures = append(failures, "old source evidence remained indexed for the source path")
	}
	if !updateEventOK {
		failures = append(failures, "source update provenance missing previous/new SHA details")
	}
	if !hasStaleProjection {
		failures = append(failures, "dependent synthesis projection is not visibly stale")
	}
	if !eventTypesInclude(projectionEvents, "projection_invalidated") {
		failures = append(failures, "synthesis projection invalidation event missing")
	}
	if !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProjectionStatesUsed || !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not use search, source/synthesis listing/get, projection, and provenance workflow")
	}
	assistantPass := messageContainsAll(finalMessage, []string{sourceURLUpdateSourcePath, sourceURLUpdateSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"changed-pdf", "changed pdf", "updated pdf", "changed"}) &&
		messageContainsAny(finalMessage, []string{"stale"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source_updated", "source update"}) &&
		messageContainsAny(finalMessage, []string{"citation", "evidence"})
	if !assistantPass {
		failures = append(failures, "final answer did not report changed update, stale projection, provenance, and citations")
	}
	databasePass := found && doc != nil && synthesisFound && synthesisDoc != nil &&
		len(missingRequired(doc.Body, []string{sourceURLUpdateChangedText, "asset_path:", sourceURLUpdateAssetPath})) == 0 &&
		len(presentForbidden(doc.Body, []string{sourceURLUpdateInitialText})) == 0 &&
		strings.Contains(synthesisDoc.Body, sourceURLUpdateInitialText) &&
		searchContainsPath(changedSearch, sourceURLUpdateSourcePath) &&
		searchResultHasCitations(changedSearch) &&
		!searchContainsPath(oldSearch, sourceURLUpdateSourcePath) &&
		updateEventOK &&
		hasStaleProjection &&
		eventTypesInclude(projectionEvents, "projection_invalidated")
	activityPass := len(sourceURLUpdateBypassFailures(turnMetrics)) == 0 &&
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
		Documents:     []string{sourceURLUpdateSourcePath, sourceURLUpdateSynthesisPath},
	}, nil
}
func verifySourceURLUpdateConflict(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	conflictCount, err := exactDocumentCount(ctx, paths, sourceURLUpdateConflictPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := sourceURLUpdateSourceEvents(ctx, paths, docIDOrEmpty(doc))
	if err != nil {
		return verificationResult{}, err
	}
	failures := sourceURLUpdateBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing original source URL document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "asset_path:", sourceURLUpdateAssetPath})...)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one original source document, got %d", sourceCount))
	}
	if conflictCount != 0 {
		failures = append(failures, "conflict update wrote "+sourceURLUpdateConflictPath)
	}
	if eventTypesInclude(sourceEvents, "source_updated") {
		failures = append(failures, "conflict update emitted source_updated provenance")
	}
	if !turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use source.mode update")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source documents after conflict")
	}
	assistantPass := messageContainsAll(finalMessage, []string{sourceURLUpdateSourcePath}) &&
		messageContainsAny(finalMessage, []string{sourceURLUpdateConflictPath, "source-url-update-conflict.md"}) &&
		messageContainsAny(finalMessage, []string{"conflict", "mismatch", "path hint", "path-hint"}) &&
		messageContainsAny(finalMessage, []string{"not created", "was not created", "no write", "without writing"})
	if !assistantPass {
		failures = append(failures, "final answer did not report path-hint conflict and no-write outcome")
	}
	databasePass := found && doc != nil && sourceCount == 1 && conflictCount == 0 &&
		len(missingRequired(doc.Body, []string{sourceURLUpdateInitialText, "asset_path:", sourceURLUpdateAssetPath})) == 0 &&
		!eventTypesInclude(sourceEvents, "source_updated")
	activityPass := len(sourceURLUpdateBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUpdateUsed &&
		turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{sourceURLUpdateSourcePath},
	}, nil
}
func sourceURLUpdateBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func docIDOrEmpty(doc *runner.Document) string {
	if doc == nil {
		return ""
	}
	return doc.DocID
}
func sourceURLUpdateSourceEvents(ctx context.Context, paths evalPaths, docID string) ([]runner.ProvenanceEvent, error) {
	if strings.TrimSpace(docID) == "" {
		return nil, nil
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "source", RefID: docID, Limit: 20},
	})
	if err != nil || result.Provenance == nil {
		return nil, err
	}
	return result.Provenance.Events, nil
}
func sourceURLUpdateProjectionEvents(ctx context.Context, paths evalPaths, synthesisDocID string) ([]runner.ProvenanceEvent, error) {
	if strings.TrimSpace(synthesisDocID) == "" {
		return nil, nil
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "projection", RefID: "synthesis:" + synthesisDocID, Limit: 20},
	})
	if err != nil || result.Provenance == nil {
		return nil, err
	}
	return result.Provenance.Events, nil
}
func sourceURLUpdateSearch(ctx context.Context, paths evalPaths, text string) (runner.RetrievalTaskResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	return runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       text,
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
}
func sourceURLUpdateEventHasSHAChange(events []runner.ProvenanceEvent) bool {
	for _, event := range events {
		if event.EventType != "source_updated" {
			continue
		}
		previous := strings.TrimSpace(event.Details["previous_sha256"])
		next := strings.TrimSpace(event.Details["new_sha256"])
		if previous != "" && next != "" && previous != next &&
			event.Details["asset_path"] == sourceURLUpdateAssetPath {
			return true
		}
	}
	return false
}
func verifyStaleSynthesisUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	docPath := "synthesis/runner-routing.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	createdCurrent, err := exactDocumentCount(ctx, paths, "synthesis/runner-routing-current.md")
	if err != nil {
		return verificationResult{}, err
	}
	createdUpdated, err := exactDocumentCount(ctx, paths, "synthesis/runner-routing-updated.md")
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", docPath, exactCount))
	}
	if createdCurrent != 0 || createdUpdated != 0 {
		failures = append(failures, "created duplicate synthesis path")
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current guidance: routine agents must use openclerk JSON runner",
		"Current source: sources/runner-current-runner.md",
		"Supersedes: sources/runner-old-workaround.md",
		"## Sources",
		"## Freshness",
	}
	sourceRefs := []string{"sources/runner-current-runner.md", "sources/runner-old-workaround.md"}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	failures = append(failures, presentForbidden(body, []string{"may bypass OpenClerk runner through a temporary command-path workaround"})...)
	if !containsAny(strings.ToLower(body), []string{"stale", "supersedes", "superseded", "contradiction", "current guidance"}) {
		failures = append(failures, "missing stale or supersession language")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	assistantPass := messageContainsAll(finalMessage, []string{docPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "current", "supersedes", "stale"})
	if !assistantPass {
		failures = append(failures, "final answer did not describe the synthesis update")
	}
	databasePass := found && exactCount == 1 && createdCurrent == 0 && createdUpdated == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		len(presentForbidden(body, []string{"may bypass OpenClerk runner through a temporary command-path workaround"})) == 0 &&
		containsAny(strings.ToLower(body), []string{"stale", "supersedes", "superseded", "contradiction", "current guidance"})
	activityPass := turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}
func verifySynthesisFreshnessRepair(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	docPath := "synthesis/runner-repair.md"
	currentSource := "sources/repair-current.md"
	supersededSource := "sources/repair-old.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      docID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	events, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + docID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", docPath, exactCount))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+docPath)
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"source_refs: sources/repair-current.md, sources/repair-old.md",
		currentSource,
		supersededSource,
		"## Sources",
		"## Freshness",
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, presentForbidden(body, []string{"may use a temporary command-path workaround"})...)
	hasProjection := false
	hasCurrent := false
	hasSuperseded := false
	if projections.Projections != nil && len(projections.Projections.Projections) == 1 {
		projection := projections.Projections.Projections[0]
		hasProjection = projection.Freshness == "fresh"
		hasCurrent = projection.Details["current_source_refs"] == currentSource
		hasSuperseded = projection.Details["superseded_source_refs"] == supersededSource
	}
	if !hasProjection {
		failures = append(failures, "synthesis projection is not fresh")
	}
	if !hasCurrent {
		failures = append(failures, "synthesis projection missing current source ref")
	}
	if !hasSuperseded {
		failures = append(failures, "synthesis projection missing superseded source ref")
	}
	hasInvalidation := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_invalidated")
	hasRefresh := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_refreshed")
	if !hasInvalidation {
		failures = append(failures, "synthesis invalidation event missing")
	}
	if !hasRefresh {
		failures = append(failures, "synthesis refresh event missing")
	}
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.ProjectionStatesUsed
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list existing synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection states")
	}
	assistantPass := messageContainsAll(finalMessage, []string{docPath, currentSource, supersededSource}) &&
		messageContainsAny(finalMessage, []string{"fresh", "freshness", "current", "superseded"})
	if !assistantPass {
		failures = append(failures, "final answer did not mention repaired freshness and source status")
	}
	databasePass := found &&
		exactCount == 1 &&
		len(missingRequired(body, required)) == 0 &&
		len(presentForbidden(body, []string{"may use a temporary command-path workaround"})) == 0 &&
		hasProjection &&
		hasCurrent &&
		hasSuperseded &&
		hasInvalidation &&
		hasRefresh
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath, currentSource, supersededSource},
	}, nil
}
func verifySynthesisCandidatePressure(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, synthesisCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, synthesisCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, synthesisCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docID)
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current compiler decision: existing document and retrieval actions are sufficient for synthesis compiler pressure repairs",
		"Current source: " + synthesisCandidateCurrentSrc,
		"Superseded source: " + synthesisCandidateOldSrc,
		"## Sources",
		"## Freshness",
	}
	sourceRefs := []string{synthesisCandidateCurrentSrc, synthesisCandidateOldSrc}
	forbidden := []string{"require a dedicated compile_synthesis runner action", "requires a dedicated compile_synthesis runner action"}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+synthesisCandidatePath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", synthesisCandidatePath, exactCount))
	}
	if synthesisCount != 2 {
		failures = append(failures, fmt.Sprintf("expected exactly target and decoy synthesis documents, got %d", synthesisCount))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+synthesisCandidatePath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	failures = append(failures, presentForbidden(body, forbidden)...)
	if projection == nil || projection.Freshness != "fresh" {
		failures = append(failures, "synthesis projection is not fresh")
	} else {
		if !projectionDetailContains(projection.Details, "current_source_refs", synthesisCandidateCurrentSrc) {
			failures = append(failures, "synthesis projection missing current compiler source")
		}
		if !projectionDetailContains(projection.Details, "superseded_source_refs", synthesisCandidateOldSrc) {
			failures = append(failures, "synthesis projection missing superseded compiler source")
		}
	}
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect synthesis projection freshness")
	}
	assistantPass := messageContainsAll(finalMessage, []string{synthesisCandidatePath, synthesisCandidateCurrentSrc}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "fresh", "freshness", "existing actions"})
	if !assistantPass {
		failures = append(failures, "final answer did not report target update and current source")
	}
	databasePass := found &&
		exactCount == 1 &&
		synthesisCount == 2 &&
		docIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		len(presentForbidden(body, forbidden)) == 0 &&
		projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", synthesisCandidateCurrentSrc) &&
		projectionDetailContains(projection.Details, "superseded_source_refs", synthesisCandidateOldSrc)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{synthesisCandidatePath, synthesisCandidateDecoyPath, synthesisCandidateCurrentSrc, synthesisCandidateOldSrc},
	}, nil
}
func verifySynthesisSourceSetPressure(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, synthesisSourceSetPath, finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:                 []string{sourceSetAlphaPath, sourceSetBetaPath, sourceSetGammaPath},
		RequireSearch:              true,
		RequireList:                true,
		Metrics:                    turnMetrics,
		FinalAnswerPath:            true,
		AdditionalDocs:             []string{sourceSetAlphaPath, sourceSetBetaPath, sourceSetGammaPath},
		AdditionalBodyRequirements: []string{"alpha", "beta", "gamma", "source refs", "freshness"},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if synthesisCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one synthesis document, got %d", synthesisCount))
	}
	databasePass := base.DatabasePass && synthesisCount == 1
	return verificationResult{
		Passed:        databasePass && base.AssistantPass,
		DatabasePass:  databasePass,
		AssistantPass: base.AssistantPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}
func verifySynthesisCompileRevisit(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, requireProvenance bool) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, synthesisCompilePath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, synthesisCompilePath)
	if err != nil {
		return verificationResult{}, err
	}
	decoyCount, err := exactDocumentCount(ctx, paths, synthesisCompileDecoyPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, synthesisCompilePath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docID)
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current compile_synthesis revisit decision",
		"existing document and retrieval actions",
		"Current source: " + synthesisCompileCurrentSrc,
		"Superseded source: " + synthesisCompileOldSrc,
		"## Sources",
		"## Freshness",
	}
	sourceRefs := []string{synthesisCompileCurrentSrc, synthesisCompileOldSrc}
	forbidden := []string{"require a dedicated compile_synthesis runner action", "requires a dedicated compile_synthesis runner action"}
	failures := populatedBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+synthesisCompilePath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", synthesisCompilePath, exactCount))
	}
	if decoyCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s decoy document, got %d", synthesisCompileDecoyPath, decoyCount))
	}
	if synthesisCount != 2 {
		failures = append(failures, fmt.Sprintf("expected exactly target and decoy synthesis documents, got %d", synthesisCount))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+synthesisCompilePath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	failures = append(failures, presentForbidden(body, forbidden)...)
	if projection == nil || projection.Freshness != "fresh" {
		failures = append(failures, "synthesis projection is not fresh")
	} else {
		if !projectionDetailContains(projection.Details, "current_source_refs", synthesisCompileCurrentSrc) {
			failures = append(failures, "synthesis projection missing current compile revisit source")
		}
		if !projectionDetailContains(projection.Details, "superseded_source_refs", synthesisCompileOldSrc) {
			failures = append(failures, "synthesis projection missing superseded compile revisit source")
		}
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect synthesis projection freshness")
	}
	if requireProvenance && !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if turnMetrics.CreateDocumentUsed {
		failures = append(failures, "agent created a document instead of updating existing synthesis")
	}
	if !turnMetrics.ReplaceSectionUsed && !turnMetrics.AppendDocumentUsed {
		failures = append(failures, "agent did not update synthesis with replace_section or append_document")
	}
	assistantPass := messageContainsAll(finalMessage, []string{synthesisCompilePath, synthesisCompileCurrentSrc}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "fresh", "freshness", "existing document", "existing actions"})
	if !assistantPass {
		failures = append(failures, "final answer did not report target update, current source, and freshness")
	}

	databasePass := found &&
		exactCount == 1 &&
		decoyCount == 1 &&
		synthesisCount == 2 &&
		docIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		len(presentForbidden(body, forbidden)) == 0 &&
		projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", synthesisCompileCurrentSrc) &&
		projectionDetailContains(projection.Details, "superseded_source_refs", synthesisCompileOldSrc)
	activityPass := len(populatedBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		(!requireProvenance || turnMetrics.ProvenanceEventsUsed) &&
		!turnMetrics.CreateDocumentUsed &&
		(turnMetrics.ReplaceSectionUsed || turnMetrics.AppendDocumentUsed)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{synthesisCompilePath, synthesisCompileDecoyPath, synthesisCompileCurrentSrc, synthesisCompileOldSrc},
	}, nil
}
func verifyMTSynthesisDriftPressure(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, mtDriftSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	currentBody, currentFound, err := documentBodyByPath(ctx, paths, mtDriftCurrentPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, mtDriftSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, mtDriftSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docID)
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current drift decision: keep existing document and retrieval actions",
		"Current source: " + mtDriftCurrentPath,
		"Superseded source: " + mtDriftOldSourcePath,
		"## Sources",
		"## Freshness",
	}
	sourceRefs := []string{mtDriftCurrentPath, mtDriftOldSourcePath}
	forbidden := []string{"promoted immediately", "dedicated compile_synthesis action is required"}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+mtDriftSynthesisPath)
	}
	if !currentFound {
		failures = append(failures, "missing "+mtDriftCurrentPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", mtDriftSynthesisPath, exactCount))
	}
	if synthesisCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one drift synthesis document, got %d", synthesisCount))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+mtDriftSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	failures = append(failures, presentForbidden(body, forbidden)...)
	if !strings.Contains(currentBody, "Current drift decision says existing document and retrieval actions should stay the v1 synthesis path.") {
		failures = append(failures, "current drift source was not updated")
	}
	if projection == nil || projection.Freshness != "fresh" {
		failures = append(failures, "drift synthesis projection is not fresh")
	} else {
		if !projectionDetailContains(projection.Details, "current_source_refs", mtDriftCurrentPath) {
			failures = append(failures, "drift synthesis projection missing current source")
		}
		if !projectionDetailContains(projection.Details, "superseded_source_refs", mtDriftOldSourcePath) {
			failures = append(failures, "drift synthesis projection missing superseded source")
		}
	}
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not get existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect synthesis projection freshness")
	}
	assistantPass := messageContainsAll(finalMessage, []string{mtDriftSynthesisPath, mtDriftCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "fresh", "current"})
	if !assistantPass {
		failures = append(failures, "final answer did not report drift repair and current source")
	}
	databasePass := found &&
		currentFound &&
		exactCount == 1 &&
		synthesisCount == 1 &&
		docIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		len(presentForbidden(body, forbidden)) == 0 &&
		strings.Contains(currentBody, "Current drift decision says existing document and retrieval actions should stay the v1 synthesis path.") &&
		projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", mtDriftCurrentPath) &&
		projectionDetailContains(projection.Details, "superseded_source_refs", mtDriftOldSourcePath)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{mtDriftSynthesisPath, mtDriftCurrentPath, mtDriftOldSourcePath},
	}, nil
}
func verifyDuplicatePathReject(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	bodyCheck, err := verifyDocumentContains(ctx, paths, "notes/projects/duplicate.md", []string{"This canonical path already exists."}, []string{"overwritten"})
	if err != nil {
		return verificationResult{}, err
	}
	answerPass := isDuplicateRejection(finalMessage)
	failures := []string{}
	if !bodyCheck.DatabasePass {
		failures = append(failures, bodyCheck.Details)
	}
	if !answerPass {
		failures = append(failures, "answer did not report the duplicate path failure")
	}
	return verificationResult{
		Passed:        bodyCheck.DatabasePass && answerPass,
		DatabasePass:  bodyCheck.DatabasePass,
		AssistantPass: answerPass,
		Details:       missingDetails(failures),
		Documents:     []string{"notes/projects/duplicate.md"},
	}, nil
}
func isDuplicateRejection(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "already exists") || strings.Contains(lower, "duplicate") {
		return true
	}
	return strings.Contains(lower, "exists") && containsAny(lower, []string{"cannot", "can't", "failed", "not overwrite", "won't overwrite", "did not overwrite"})
}
