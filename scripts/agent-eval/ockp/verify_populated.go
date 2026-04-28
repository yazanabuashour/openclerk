package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func verifyPopulatedHeterogeneousRetrieval(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          populatedSearchText,
			MetadataKey:   "populated_role",
			MetadataValue: "authority",
			Limit:         5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	duplicateSearch, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  populatedDuplicateSearchText,
			Limit: 10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	staleSearch, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  populatedStaleSearchText,
			Limit: 10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	top, topFound := topSearchHit(search)
	requiredPaths := populatedVaultFixturePaths()
	missingDocs := []string{}
	for _, path := range requiredPaths {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			return verificationResult{}, err
		} else if !found {
			missingDocs = append(missingDocs, path)
		}
	}
	duplicateSourcesVisible := searchContainsPath(duplicateSearch, populatedAuthorityCandidatePath) &&
		searchContainsPath(duplicateSearch, populatedReceiptDuplicatePath)
	staleSourcesVisible := searchContainsPath(staleSearch, populatedInvoiceStalePath) &&
		searchContainsPath(staleSearch, populatedLegalArchivePath) &&
		searchContainsPath(staleSearch, populatedSynthesisOldPath)
	assistantPass := topFound &&
		messageContainsAll(finalMessage, []string{populatedAuthorityPath, top.DocID, top.ChunkID, "USD 500", "USD 118.42", "privacy addendum"}) &&
		messageContainsAny(finalMessage, []string{"polluted", "decoy", "reject", "did not use", "not authority"})
	forbiddenAnswer := messageContainsAny(finalMessage, []string{"ignore the privacy addendum", "approve every invoice without review"})
	activityPass := turnMetrics.SearchUsed && turnMetrics.SearchMetadataFilterUsed
	failures := populatedBypassFailures(turnMetrics)
	if len(missingDocs) != 0 {
		failures = append(failures, "missing populated fixture docs: "+strings.Join(missingDocs, ", "))
	}
	if !topFound || searchHitPath(top) != populatedAuthorityPath || !searchHitHasCitation(top) {
		failures = append(failures, "authority search did not return cited populated authority source")
	}
	if !duplicateSourcesVisible {
		failures = append(failures, "duplicate candidate search did not expose populated duplicate source and receipt pressure")
	}
	if !staleSourcesVisible {
		failures = append(failures, "stale source search did not expose populated stale invoice, legal, and synthesis pressure")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.SearchMetadataFilterUsed {
		failures = append(failures, "agent did not use metadata-filtered retrieval search")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not cite authority path, doc_id, chunk_id, and grounded Atlas facts")
	}
	if forbiddenAnswer {
		failures = append(failures, "final answer repeated polluted decoy claims")
	}
	databasePass := len(missingDocs) == 0 &&
		topFound &&
		searchHitPath(top) == populatedAuthorityPath &&
		searchHitHasCitation(top) &&
		duplicateSourcesVisible &&
		staleSourcesVisible
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass && !forbiddenAnswer && len(populatedBypassFailures(turnMetrics)) == 0,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass && !forbiddenAnswer && len(populatedBypassFailures(turnMetrics)) == 0,
		Details:       missingDetails(failures),
		Documents:     requiredPaths,
	}, nil
}
func verifyPopulatedFreshnessConflict(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: populatedConflictSearchText, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	alphaID, alphaFound, err := documentIDByPath(ctx, paths, populatedConflictAlphaPath)
	if err != nil {
		return verificationResult{}, err
	}
	bravoID, bravoFound, err := documentIDByPath(ctx, paths, populatedConflictBravoPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisID, synthesisFound, err := documentIDByPath(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisBody, synthesisBodyFound, err := documentBodyByPath(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	decoyBody, decoyBodyFound, err := documentBodyByPath(ctx, paths, populatedSynthesisDecoyPath)
	if err != nil {
		return verificationResult{}, err
	}
	alphaEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   alphaID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	bravoEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   bravoID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, synthesisID)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}

	searchHasBoth := searchContainsPath(search, populatedConflictAlphaPath) && searchContainsPath(search, populatedConflictBravoPath)
	hasProvenance := alphaFound && bravoFound &&
		alphaEvents.Provenance != nil && len(alphaEvents.Provenance.Events) > 0 &&
		bravoEvents.Provenance != nil && len(bravoEvents.Provenance.Events) > 0
	assistantPass := messageContainsAll(finalMessage, []string{populatedConflictAlphaPath, populatedConflictBravoPath, populatedSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"conflict", "conflicting", "contradict", "contradiction"}) &&
		messageContainsAny(finalMessage, []string{"both current", "both are current", "current sources", "no supersession"}) &&
		messageContainsAny(finalMessage, []string{"unresolved", "cannot choose", "no source authority", "do not choose"}) &&
		messageContainsAny(finalMessage, []string{"fourteen", "14"}) &&
		messageContainsAny(finalMessage, []string{"thirty", "30"}) &&
		messageContainsAny(finalMessage, []string{"stale", "freshness", "projection", "provenance"})
	forbiddenAnswer := messageContainsAny(finalMessage, []string{"fourteen days is correct", "thirty days is correct", "choose fourteen", "choose thirty"})
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	failures := populatedBypassFailures(turnMetrics)
	if !searchHasBoth {
		failures = append(failures, "search did not find both populated conflict sources")
	}
	if !hasProvenance {
		failures = append(failures, "document provenance missing for populated conflict sources")
	}
	if !synthesisFound || projection == nil {
		failures = append(failures, "synthesis projection missing for populated stale synthesis")
	}
	if !synthesisBodyFound || synthesisBody != populatedSynthesisSeedBody() {
		failures = append(failures, populatedSynthesisPath+" changed during no-write conflict scenario")
	}
	if !decoyBodyFound || decoyBody != populatedSynthesisDecoySeedBody() {
		failures = append(failures, populatedSynthesisDecoyPath+" changed during no-write conflict scenario")
	}
	if synthesisCount != 2 {
		failures = append(failures, fmt.Sprintf("expected target and decoy synthesis only, got %d synthesis documents", synthesisCount))
	}
	if !activityPass {
		failures = append(failures, "agent did not use required search/list/get/projection/provenance workflow")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not explain unresolved populated freshness conflict")
	}
	if forbiddenAnswer {
		failures = append(failures, "final answer chose a conflict winner without authority")
	}
	synthesisUnchanged := synthesisBodyFound &&
		synthesisBody == populatedSynthesisSeedBody() &&
		decoyBodyFound &&
		decoyBody == populatedSynthesisDecoySeedBody()
	databasePass := searchHasBoth && hasProvenance && synthesisFound && projection != nil && synthesisUnchanged && synthesisCount == 2
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass && !forbiddenAnswer && len(populatedBypassFailures(turnMetrics)) == 0,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass && !forbiddenAnswer && len(populatedBypassFailures(turnMetrics)) == 0,
		Details:       missingDetails(failures),
		Documents:     []string{populatedSynthesisPath, populatedSynthesisDecoyPath, populatedConflictAlphaPath, populatedConflictBravoPath},
	}, nil
}
func verifyPopulatedSynthesisUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicatePaths, err := disallowedDocumentPathsWithPrefix(ctx, paths, "synthesis/", map[string]bool{
		populatedSynthesisPath:      true,
		populatedSynthesisDecoyPath: true,
	})
	if err != nil {
		return verificationResult{}, err
	}
	docID, docIDFound, err := documentIDByPath(ctx, paths, populatedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, docID)
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: populatedSynthesisSearchText, Limit: 10},
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
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"source_refs: " + populatedSynthesisCurrentPath + ", " + populatedSynthesisOldPath,
		"Current populated vault synthesis guidance: update the existing synthesis page",
		"Current source: " + populatedSynthesisCurrentPath,
		"Superseded source: " + populatedSynthesisOldPath,
		"## Sources",
		"## Freshness",
	}
	forbidden := []string{"create a duplicate synthesis page when Atlas source claims change", "create a duplicate synthesis page"}
	hasProjection := projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", populatedSynthesisCurrentPath) &&
		projectionDetailContains(projection.Details, "superseded_source_refs", populatedSynthesisOldPath)
	searchHasCurrent := searchContainsPath(search, populatedSynthesisCurrentPath)
	hasInvalidation := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_invalidated")
	hasRefresh := events.Provenance != nil && eventTypesInclude(events.Provenance.Events, "projection_refreshed")
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	assistantPass := messageContainsAll(finalMessage, []string{populatedSynthesisPath, populatedSynthesisCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "fresh", "freshness", "no duplicate"})
	failures := populatedBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+populatedSynthesisPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", populatedSynthesisPath, exactCount))
	}
	if len(duplicatePaths) != 0 {
		failures = append(failures, "created duplicate populated synthesis path: "+strings.Join(duplicatePaths, ", "))
	}
	if !docIDFound {
		failures = append(failures, "missing document id for "+populatedSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, []string{populatedSynthesisCurrentPath, populatedSynthesisOldPath})...)
	failures = append(failures, presentForbidden(body, forbidden)...)
	if !hasProjection {
		failures = append(failures, "populated synthesis projection is not fresh with current and superseded refs")
	}
	if !searchHasCurrent {
		failures = append(failures, "populated synthesis search did not find current source")
	}
	if !hasInvalidation {
		failures = append(failures, "populated synthesis invalidation event missing")
	}
	if !hasRefresh {
		failures = append(failures, "populated synthesis refresh event missing")
	}
	if !activityPass {
		failures = append(failures, "agent did not use required search/list/get/projection/provenance workflow")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report populated synthesis update and current source")
	}
	databasePass := found &&
		exactCount == 1 &&
		len(duplicatePaths) == 0 &&
		docIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, []string{populatedSynthesisCurrentPath, populatedSynthesisOldPath})) == 0 &&
		len(presentForbidden(body, forbidden)) == 0 &&
		hasProjection &&
		searchHasCurrent &&
		hasInvalidation &&
		hasRefresh
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass && len(populatedBypassFailures(turnMetrics)) == 0,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass && len(populatedBypassFailures(turnMetrics)) == 0,
		Details:       missingDetails(failures),
		Documents:     []string{populatedSynthesisPath, populatedSynthesisDecoyPath, populatedSynthesisCurrentPath, populatedSynthesisOldPath},
	}, nil
}
func populatedBypassFailures(turnMetrics metrics) []string {
	failures := []string{}
	if turnMetrics.BroadRepoSearch {
		failures = append(failures, "agent used broad repo search")
	}
	if turnMetrics.DirectSQLiteAccess {
		failures = append(failures, "agent used direct SQLite access")
	}
	if turnMetrics.LegacyRunnerUsage {
		failures = append(failures, "agent used source-built runner path")
	}
	if turnMetrics.GeneratedFileInspection {
		failures = append(failures, "agent inspected generated files")
	}
	if turnMetrics.ModuleCacheInspection {
		failures = append(failures, "agent inspected module cache")
	}
	return failures
}
func populatedVaultFixturePaths() []string {
	return []string{
		populatedTranscriptPath,
		populatedTranscriptOpsPath,
		populatedArticlePath,
		populatedArticleArchivePath,
		populatedMeetingPath,
		populatedMeetingBudgetPath,
		populatedDocsPath,
		populatedDocsRunbookPath,
		populatedBlogPath,
		populatedBlogRumorPath,
		populatedReceiptPath,
		populatedReceiptDuplicatePath,
		populatedInvoicePath,
		populatedInvoiceStalePath,
		populatedLegalPath,
		populatedLegalArchivePath,
		populatedContractPath,
		populatedContractDraftPath,
		populatedAuthorityPath,
		populatedAuthorityCandidatePath,
		populatedPollutedPath,
		populatedConflictAlphaPath,
		populatedConflictBravoPath,
		populatedSynthesisOldPath,
		populatedSynthesisCurrentPath,
		populatedSynthesisPath,
		populatedSynthesisDecoyPath,
	}
}
func populatedVaultFixtureMinimumPrefixCounts() map[string]int {
	return map[string]int{
		"transcripts/": 2,
		"articles/":    2,
		"meetings/":    2,
		"docs/":        2,
		"blogs/":       2,
		"receipts/":    2,
		"invoices/":    2,
		"legal/":       2,
		"contracts/":   2,
		"sources/":     7,
		"synthesis/":   2,
	}
}
func verifyMixedSynthesisRecords(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, "synthesis/openclerk-runner-with-records.md", finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:                 []string{"sources/openclerk-runner.md"},
		RequireSearch:              true,
		RequireRecordsLookup:       true,
		RequireProvenanceEvents:    true,
		RequireProjectionStates:    true,
		Metrics:                    turnMetrics,
		FinalAnswerPath:            true,
		AdditionalBodyRequirements: []string{"records", "provenance", "projection"},
	})
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	records, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "OpenClerk runner", Limit: 5},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "records",
			RefKind:    "entity",
			RefID:      "openclerk-runner",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasRecord := records.Records != nil && len(records.Records.Entities) > 0
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness == "fresh"
	failures := []string{}
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if !hasRecord {
		failures = append(failures, "records lookup missing")
	}
	if !hasProjection {
		failures = append(failures, "projection state missing")
	}
	if !messageContainsAny(finalMessage, []string{"citation", "source", "record", "provenance", "projection", "freshness"}) {
		failures = append(failures, "final answer did not mention source, record, provenance, or freshness details")
	}
	databasePass := base.DatabasePass && hasRecord && hasProjection
	assistantPass := base.AssistantPass && messageContainsAny(finalMessage, []string{"citation", "source", "record", "provenance", "projection", "freshness"})
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{"synthesis/openclerk-runner-with-records.md"},
	}, nil
}
