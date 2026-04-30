package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func verifyTaggingCreateUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, taggingCreatePath)
	if err != nil {
		return verificationResult{}, err
	}
	count, err := exactDocumentCount(ctx, paths, taggingCreatePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := taggingBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing created tagged document "+taggingCreatePath)
	}
	if count != 1 {
		failures = append(failures, fmt.Sprintf("expected one created tagged document %s, got %d", taggingCreatePath, count))
	}
	if doc != nil {
		if got := strings.TrimSpace(doc.Metadata["tag"]); got != taggingLaunchRiskTag {
			failures = append(failures, fmt.Sprintf("expected tag metadata %q, got %q", taggingLaunchRiskTag, got))
		}
		if !strings.Contains(doc.Body, "Launch readiness tag update evidence remains on the same tagged document.") {
			failures = append(failures, "created tagged document was not updated")
		}
	}
	if !turnMetrics.CreateDocumentUsed {
		failures = append(failures, "did not create tagged document")
	}
	if !turnMetrics.AppendDocumentUsed {
		failures = append(failures, "did not update tagged document")
	}
	if !taggingSearchMetadataFilterUsed(turnMetrics, "tag", taggingLaunchRiskTag) {
		failures = append(failures, "did not verify tag through backward-compatible search metadata filter")
	}
	if !taggingListMetadataFilterUsed(turnMetrics, "tag", taggingLaunchRiskTag) {
		failures = append(failures, "did not verify tag through backward-compatible list_documents metadata filter")
	}
	if !containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{taggingPrefix}) {
		failures = append(failures, "list_documents did not constrain tagged create/update verification to "+taggingPrefix)
	}
	assistantPass := messageContainsAll(finalMessage, []string{taggingCreatePath, taggingLaunchRiskTag}) &&
		messageContainsAny(finalMessage, []string{"frontmatter", "canonical markdown"}) &&
		messageContainsAny(finalMessage, []string{"same document", "updated"}) &&
		messageContainsAny(finalMessage, []string{"metadata_key", "metadata_value", "metadata"})
	if !assistantPass {
		failures = append(failures, "final answer did not report tag, canonical authority, update, and metadata filter ceremony")
	}
	databasePass := found && count == 1 && doc != nil &&
		strings.TrimSpace(doc.Metadata["tag"]) == taggingLaunchRiskTag &&
		strings.Contains(doc.Body, "Launch readiness tag update evidence remains on the same tagged document.")
	activityPass := len(taggingBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.CreateDocumentUsed &&
		turnMetrics.AppendDocumentUsed &&
		taggingSearchMetadataFilterUsed(turnMetrics, "tag", taggingLaunchRiskTag) &&
		taggingListMetadataFilterUsed(turnMetrics, "tag", taggingLaunchRiskTag)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{taggingCreatePath},
	}, nil
}

func verifyTaggingRetrieval(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	return verifyTaggingLookup(ctx, paths, finalMessage, turnMetrics, taggingLookupExpectation{
		TargetPath:       taggingRetrievalPath,
		TargetTag:        taggingAccountRenewalTag,
		ForbiddenPath:    taggingRetrievalDecoyPath,
		RequirePathScope: false,
		RequireCeremony:  true,
		RequireTagField:  true,
		Posture:          "tag retrieval used promoted tag filter while preserving canonical markdown/frontmatter authority",
	})
}

func verifyTaggingDisambiguation(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	return verifyTaggingLookup(ctx, paths, finalMessage, turnMetrics, taggingLookupExpectation{
		TargetPath:       taggingDisambiguationTargetPath,
		TargetTag:        taggingCustomerRiskTag,
		ForbiddenPath:    taggingDisambiguationDecoyPath,
		ForbiddenTag:     taggingCustomerRiskArchiveTag,
		RequirePathScope: false,
		RequireExactText: true,
		RequireTagField:  true,
		Posture:          "tag disambiguation used exact promoted tag filtering and excluded adjacent tag authority",
	})
}

func verifyTaggingNearDuplicate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	return verifyTaggingLookup(ctx, paths, finalMessage, turnMetrics, taggingLookupExpectation{
		TargetPath:       taggingNearDuplicateTargetPath,
		TargetTag:        taggingOpsReviewTag,
		ForbiddenPath:    taggingNearDuplicateDecoyPath,
		ForbiddenTag:     taggingOpsReviewsTag,
		RequirePathScope: false,
		RequireExactText: true,
		RequireTagField:  true,
		Posture:          "near-duplicate tag names used exact promoted tag filtering without merging distinct tags",
	})
}

func verifyTaggingMixedPath(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	return verifyTaggingLookup(ctx, paths, finalMessage, turnMetrics, taggingLookupExpectation{
		TargetPath:       taggingMixedPathTargetPath,
		TargetTag:        taggingSupportHandoffTag,
		ForbiddenPath:    taggingMixedPathArchivePath,
		RequirePathScope: true,
		RequireTagField:  true,
		Posture:          "mixed path plus tag query used both path_prefix and promoted tag filtering",
	})
}

type taggingLookupExpectation struct {
	TargetPath       string
	TargetTag        string
	ForbiddenPath    string
	ForbiddenTag     string
	RequirePathScope bool
	RequireCeremony  bool
	RequireExactText bool
	RequireTagField  bool
	Posture          string
}

func verifyTaggingLookup(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, expectation taggingLookupExpectation) (verificationResult, error) {
	target, targetFound, err := documentByPath(ctx, paths, expectation.TargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	forbiddenCount, err := exactDocumentCount(ctx, paths, expectation.ForbiddenPath)
	if err != nil {
		return verificationResult{}, err
	}
	matches, err := taggedDocuments(ctx, paths, expectation.TargetTag, expectation.pathPrefix())
	if err != nil {
		return verificationResult{}, err
	}

	failures := taggingBypassFailures(turnMetrics)
	if !targetFound {
		failures = append(failures, "missing tagged target "+expectation.TargetPath)
	}
	if forbiddenCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one forbidden/decoy tagged document %s, got %d", expectation.ForbiddenPath, forbiddenCount))
	}
	if target != nil && strings.TrimSpace(target.Metadata["tag"]) != expectation.TargetTag {
		failures = append(failures, fmt.Sprintf("expected target tag %q, got %q", expectation.TargetTag, target.Metadata["tag"]))
	}
	if !containsAllStrings(matches, []string{expectation.TargetPath}) {
		failures = append(failures, "tag-filtered lookup did not include target "+expectation.TargetPath)
	}
	if containsAnyString(matches, []string{expectation.ForbiddenPath}) {
		failures = append(failures, "tag-filtered lookup included forbidden path "+expectation.ForbiddenPath)
	}
	if !taggingSearchFilterUsed(turnMetrics, "tag", expectation.TargetTag) {
		failures = append(failures, "search did not use expected tag filter")
	}
	if !taggingListFilterUsed(turnMetrics, "tag", expectation.TargetTag) {
		failures = append(failures, "list_documents did not use expected tag filter")
	}
	if expectation.RequireTagField && !taggingSearchTagFilterUsed(turnMetrics, expectation.TargetTag) {
		failures = append(failures, "search did not use promoted tag filter")
	}
	if expectation.RequireTagField && !taggingListTagFilterUsed(turnMetrics, expectation.TargetTag) {
		failures = append(failures, "list_documents did not use promoted tag filter")
	}
	if expectation.RequirePathScope {
		if !containsAllStrings(turnMetrics.SearchPathPrefixes, []string{taggingPrefix}) {
			failures = append(failures, "search did not combine tag filter with "+taggingPrefix)
		}
		if !containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{taggingPrefix}) {
			failures = append(failures, "list_documents did not combine tag filter with "+taggingPrefix)
		}
	}
	assistantPass := messageContainsAll(finalMessage, []string{expectation.TargetPath, expectation.TargetTag}) &&
		messageContainsAny(finalMessage, []string{"no durable write", "no document was created", "did not create", "no write", "no durable write occurred"})
	if expectation.RequireCeremony {
		ceremonyReported := messageContainsAny(finalMessage, []string{"metadata_key", "metadata_value", "ceremony", "metadata", "first-class tag", "tag filter"})
		tagFieldAvoidedMetadata := expectation.RequireTagField &&
			taggingSearchTagFilterUsed(turnMetrics, expectation.TargetTag) &&
			taggingListTagFilterUsed(turnMetrics, expectation.TargetTag) &&
			!turnMetrics.SearchMetadataFilterUsed &&
			!turnMetrics.ListMetadataFilterUsed
		assistantPass = assistantPass && (ceremonyReported || tagFieldAvoidedMetadata)
	}
	if expectation.RequireExactText {
		assistantPass = assistantPass && messageContainsAny(finalMessage, []string{"exact", "excluded", "disambiguation"})
	}
	if expectation.ForbiddenPath != "" && (expectation.RequireExactText || expectation.RequirePathScope) {
		assistantPass = assistantPass && messageContainsAny(finalMessage, []string{"excluded", "not return", "not include"})
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report expected tag lookup, exclusions, no-write outcome, or ceremony")
	}

	databasePass := targetFound &&
		target != nil &&
		forbiddenCount == 1 &&
		strings.TrimSpace(target.Metadata["tag"]) == expectation.TargetTag &&
		containsAllStrings(matches, []string{expectation.TargetPath}) &&
		!containsAnyString(matches, []string{expectation.ForbiddenPath})
	activityPass := len(taggingBypassFailures(turnMetrics)) == 0 &&
		taggingSearchFilterUsed(turnMetrics, "tag", expectation.TargetTag) &&
		taggingListFilterUsed(turnMetrics, "tag", expectation.TargetTag)
	if expectation.RequireTagField {
		activityPass = activityPass &&
			taggingSearchTagFilterUsed(turnMetrics, expectation.TargetTag) &&
			taggingListTagFilterUsed(turnMetrics, expectation.TargetTag)
	}
	if expectation.RequirePathScope {
		activityPass = activityPass &&
			containsAllStrings(turnMetrics.SearchPathPrefixes, []string{taggingPrefix}) &&
			containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{taggingPrefix})
	}
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{expectation.TargetPath, expectation.ForbiddenPath},
	}, nil
}

func (e taggingLookupExpectation) pathPrefix() string {
	if e.RequirePathScope {
		return taggingPrefix
	}
	return ""
}

func taggedDocuments(ctx context.Context, paths evalPaths, tag string, pathPrefix string) ([]string, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	client, err := runclient.OpenReadOnly(cfg)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = client.Close()
	}()
	result, err := client.ListDocuments(ctx, domain.DocumentListQuery{
		PathPrefix:    pathPrefix,
		MetadataKey:   "tag",
		MetadataValue: tag,
		Limit:         100,
	})
	if err != nil {
		return nil, err
	}
	pathsOut := make([]string, 0, len(result.Documents))
	for _, doc := range result.Documents {
		pathsOut = append(pathsOut, doc.Path)
	}
	return pathsOut, nil
}

func taggingSearchFilterUsed(turnMetrics metrics, key string, value string) bool {
	return taggingSearchMetadataFilterUsed(turnMetrics, key, value) || taggingSearchTagFilterUsed(turnMetrics, value)
}

func taggingListFilterUsed(turnMetrics metrics, key string, value string) bool {
	return taggingListMetadataFilterUsed(turnMetrics, key, value) || taggingListTagFilterUsed(turnMetrics, value)
}

func taggingSearchMetadataFilterUsed(turnMetrics metrics, key string, value string) bool {
	return containsAllStrings(turnMetrics.SearchMetadataFilters, []string{key + "=" + value})
}

func taggingListMetadataFilterUsed(turnMetrics metrics, key string, value string) bool {
	return containsAllStrings(turnMetrics.ListMetadataFilters, []string{key + "=" + value})
}

func taggingSearchTagFilterUsed(turnMetrics metrics, value string) bool {
	return containsAllStrings(turnMetrics.SearchTagFilters, []string{value})
}

func taggingListTagFilterUsed(turnMetrics metrics, value string) bool {
	return containsAllStrings(turnMetrics.ListTagFilters, []string{value})
}

func containsAnyString(values []string, candidates []string) bool {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		for _, value := range values {
			if value == candidate {
				return true
			}
		}
	}
	return false
}

func taggingBypassFailures(turnMetrics metrics) []string {
	failures := populatedBypassFailures(turnMetrics)
	if turnMetrics.FileInspectionCommands != 0 {
		failures = append(failures, "used direct file inspection for tagging workflow")
	}
	if turnMetrics.ManualHTTPFetch {
		failures = append(failures, "used manual HTTP fetch for tagging workflow")
	}
	if turnMetrics.BrowserAutomation {
		failures = append(failures, "used browser automation for tagging workflow")
	}
	if turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed {
		failures = append(failures, "used ingestion action for tagging workflow")
	}
	return failures
}

func classifyTargetedTaggingResult(result jobResult) (string, string) {
	if isFinalAnswerOnlyValidationScenario(result.Scenario) {
		if result.Passed && result.Verification.Passed {
			return "none", "validation control stayed final-answer-only"
		}
		if result.Metrics.ToolCalls != 0 || result.Metrics.CommandExecutions != 0 || result.Metrics.AssistantCalls > 1 {
			return "skill_guidance_or_eval_coverage", "validation pressure did not stay final-answer-only"
		}
		return "skill_guidance_or_eval_coverage", "validation answer did not satisfy the rejection contract"
	}
	if result.Passed && result.Verification.Passed {
		if result.Scenario == taggingCreateUpdateScenarioID {
			return "none", "backward-compatible metadata_key/metadata_value primitives preserved canonical markdown tag authority and runner-only boundaries"
		}
		return "none", "promoted tag filter preserved canonical markdown tag authority and runner-only boundaries"
	}
	if len(taggingBypassFailures(result.Metrics)) != 0 {
		return "eval_contract_violation", "agent used a prohibited bypass or inspection path"
	}
	if result.Verification.Passed {
		return "eval_contract_violation", "scenario verification passed, but the job did not complete successfully"
	}
	if !result.Verification.DatabasePass {
		return "capability_gap", "tag filters could not express the tagged workflow safely"
	}
	if result.Scenario == taggingRetrievalScenarioID && !result.Verification.AssistantPass {
		return "ergonomics_gap", "natural tag retrieval intent did not complete with the promoted tag filter"
	}
	if result.Verification.DatabasePass && !result.Verification.AssistantPass {
		return "skill_guidance_or_eval_coverage", "tagged evidence existed, but the assistant answer or required runner steps did not satisfy the scenario"
	}
	return "ergonomics_gap", "manual review required before accepting the promoted tag filter surface"
}
