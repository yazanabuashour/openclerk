package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runner"
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
	} else if strings.Contains(synthesis.Body, webURLChangedText) {
		failures = append(failures, "synthesis was repaired during stale-impact inspection")
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
		!strings.Contains(synthesis.Body, webURLChangedText) &&
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
		!strings.Contains(synthesis.Body, webURLChangedText) &&
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

func verifyWebURLStaleImpactResponseCandidate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	baseEvidenceMessage := "Duplicate normalized source URL was rejected and " + webURLDuplicatePath + " was not created. Changed web update refreshed " + webURLSourcePath + " with " + webURLChangedText + "; the second same-hash update was a no-op. " + webURLSynthesisPath + " now has stale synthesis projection freshness with provenance evidence. No browser or manual acquisition was used."
	base, err := verifyWebURLStaleRepair(ctx, paths, baseEvidenceMessage, turnMetrics, true)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if base.Details != "" && base.Details != "ok" {
		failures = append(failures, base.Details)
	}
	sourceDocID, sourceDocIDFound, err := documentIDByPath(ctx, paths, webURLSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDocID, synthesisDocIDFound, err := documentIDByPath(ctx, paths, webURLSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	var sourceEvents []runner.ProvenanceEvent
	if sourceDocIDFound {
		sourceEvents, err = sourceURLUpdateSourceEvents(ctx, paths, sourceDocID)
		if err != nil {
			return verificationResult{}, err
		}
	}
	previousSourceSHA, newSourceSHA, shaChange := webURLSourceUpdatedSHAChange(sourceEvents)
	if !shaChange {
		failures = append(failures, "source update provenance missing previous/new SHA details")
	}
	candidatePass, candidateFailures := validateWebURLStaleImpactCandidateObject(finalMessage, sourceDocID, docIDOrEmptyString(sourceDocIDFound, sourceDocID), docIDOrEmptyString(synthesisDocIDFound, synthesisDocID), previousSourceSHA, newSourceSHA)
	failures = append(failures, candidateFailures...)
	databasePass := base.DatabasePass && shaChange
	assistantPass := base.AssistantPass && candidatePass
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{webURLSourcePath, webURLDuplicatePath, webURLSynthesisPath},
	}, nil
}

func docIDOrEmptyString(found bool, docID string) string {
	if !found {
		return ""
	}
	return docID
}

func validateWebURLStaleImpactCandidateObject(finalMessage string, sourceDocID string, expectedSourceDocID string, expectedSynthesisDocID string, expectedPreviousSHA string, expectedNewSHA string) (bool, []string) {
	candidate, ok := extractWebURLStaleImpactCandidateObject(finalMessage)
	if !ok {
		return false, []string{"final answer did not contain exactly one fenced stale-impact candidate JSON object"}
	}
	failures := []string{}
	required := []string{
		"update_status",
		"normalized_source_url",
		"source_path",
		"source_doc_id",
		"previous_sha256",
		"new_sha256",
		"changed",
		"duplicate_status",
		"stale_dependents",
		"projection_refs",
		"provenance_refs",
		"synthesis_repaired",
		"no_repair_warning",
	}
	for _, field := range required {
		if _, found := candidate[field]; !found {
			failures = append(failures, "candidate object missing "+field)
		}
	}
	if !containsAny(strings.ToLower(stringValue(candidate["update_status"])), []string{"changed", "updated", "refreshed"}) {
		failures = append(failures, "candidate update_status did not report a changed update")
	}
	if !truthyValue(candidate["changed"]) {
		failures = append(failures, "candidate changed was not true")
	}
	if stringValue(candidate["source_path"]) != webURLSourcePath {
		failures = append(failures, "candidate source_path did not match "+webURLSourcePath)
	}
	if expectedSourceDocID != "" && stringValue(candidate["source_doc_id"]) != expectedSourceDocID {
		failures = append(failures, "candidate source_doc_id did not match runner source doc_id")
	}
	normalizedURL := strings.ToLower(stringValue(candidate["normalized_source_url"]))
	if normalizedURL == "" || !strings.Contains(normalizedURL, "product-page") {
		failures = append(failures, "candidate normalized_source_url did not identify the product page")
	}
	previousSHA := strings.TrimSpace(stringValue(candidate["previous_sha256"]))
	newSHA := strings.TrimSpace(stringValue(candidate["new_sha256"]))
	if previousSHA == "" || newSHA == "" || previousSHA == newSHA {
		failures = append(failures, "candidate previous_sha256/new_sha256 did not prove a changed hash")
	}
	if expectedPreviousSHA != "" && previousSHA != expectedPreviousSHA {
		failures = append(failures, "candidate previous_sha256 did not match source_updated provenance")
	}
	if expectedNewSHA != "" && newSHA != expectedNewSHA {
		failures = append(failures, "candidate new_sha256 did not match source_updated provenance")
	}
	if !valueContainsAll(candidate["duplicate_status"], []string{"reject"}) ||
		!valueContainsAny(candidate["duplicate_status"], []string{"not created", "no copy", "no duplicate", "rejected_no_copy"}) {
		failures = append(failures, "candidate duplicate_status did not prove duplicate rejection without copy")
	}
	if !valueContainsAll(candidate["stale_dependents"], []string{webURLSynthesisPath}) ||
		!valueContainsAny(candidate["stale_dependents"], []string{"stale", "stale_source_refs", "projection"}) {
		failures = append(failures, "candidate stale_dependents did not report stale dependent synthesis")
	}
	projectionRefOptions := []string{"synthesis:", "synthesis", "projection", webURLSynthesisPath}
	if expectedSynthesisDocID != "" {
		projectionRefOptions = append(projectionRefOptions, expectedSynthesisDocID)
	}
	if !valueContainsAny(candidate["projection_refs"], projectionRefOptions) {
		failures = append(failures, "candidate projection_refs did not include synthesis projection evidence")
	}
	provenanceRefs := candidate["provenance_refs"]
	sourceProvenanceOptions := []string{"source_updated"}
	if sourceDocID != "" {
		sourceProvenanceOptions = append(sourceProvenanceOptions, sourceDocID)
	}
	if !valueContainsAny(provenanceRefs, sourceProvenanceOptions) ||
		!valueContainsAny(provenanceRefs, []string{"synthesis:", "projection"}) ||
		!valueContainsAny(provenanceRefs, []string{"no_browser", "no browser", "no_manual", "no manual", "runner-owned", "runner owned"}) {
		failures = append(failures, "candidate provenance_refs did not include source update, projection, and runner-owned no-browser/no-manual evidence")
	}
	if !falseValue(candidate["synthesis_repaired"]) {
		failures = append(failures, "candidate synthesis_repaired was not false")
	}
	if !valueContainsAll(candidate["no_repair_warning"], []string{webURLSynthesisPath}) ||
		!valueContainsAny(candidate["no_repair_warning"], []string{"did not repair", "not repaired", "no repair"}) {
		failures = append(failures, "candidate no_repair_warning did not warn that synthesis remained unrepaired")
	}
	return len(failures) == 0, failures
}

func extractWebURLStaleImpactCandidateObject(message string) (map[string]any, bool) {
	object, ok := exactFencedJSONObject(message)
	if !ok {
		return nil, false
	}
	candidate := map[string]any{}
	if err := json.Unmarshal([]byte(object), &candidate); err != nil {
		return nil, false
	}
	if _, found := candidate["update_status"]; !found {
		return nil, false
	}
	return candidate, true
}

func exactFencedJSONObject(message string) (string, bool) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(message, "\r\n", "\n"))
	if !strings.HasPrefix(trimmed, "```json\n") || !strings.HasSuffix(trimmed, "\n```") {
		return "", false
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "```json\n"), "\n```"))
	if strings.Contains(inner, "```") {
		return "", false
	}
	objects := jsonObjectsInText(inner)
	if len(objects) != 1 || strings.TrimSpace(objects[0]) != inner {
		return "", false
	}
	return inner, true
}

func jsonObjectsInText(text string) []string {
	objects := []string{}
	start := -1
	depth := 0
	inString := false
	escaped := false
	for i, r := range text {
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == '"' {
				inString = false
			}
			continue
		}
		if r == '"' {
			inString = true
			continue
		}
		if r == '{' {
			if depth == 0 {
				start = i
			}
			depth++
			continue
		}
		if r == '}' && depth > 0 {
			depth--
			if depth == 0 && start >= 0 {
				objects = append(objects, text[start:i+1])
				start = -1
			}
		}
	}
	return objects
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%.0f", typed)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func valueText(value any) string {
	encoded, err := json.Marshal(value)
	if err == nil {
		return strings.ToLower(string(encoded))
	}
	return strings.ToLower(stringValue(value))
}

func valueContainsAll(value any, required []string) bool {
	text := valueText(value)
	for _, item := range required {
		if !strings.Contains(text, strings.ToLower(item)) {
			return false
		}
	}
	return true
}

func valueContainsAny(value any, options []string) bool {
	text := valueText(value)
	for _, item := range options {
		if strings.Contains(text, strings.ToLower(item)) {
			return true
		}
	}
	return false
}

func truthyValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func falseValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return !typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "false")
	default:
		return false
	}
}

func webURLSourceUpdatedSHAChange(events []runner.ProvenanceEvent) (string, string, bool) {
	for _, event := range events {
		if event.EventType != "source_updated" {
			continue
		}
		previous := strings.TrimSpace(event.Details["previous_sha256"])
		next := strings.TrimSpace(event.Details["new_sha256"])
		if previous != "" && next != "" && previous != next {
			return previous, next, true
		}
	}
	return "", "", false
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
	failures := artifactIngestionBypassFailures(turnMetrics)
	if turnMetrics.ManualHTTPFetch {
		failures = append(failures, "agent used manual HTTP fetch")
	}
	if turnMetrics.BrowserAutomation {
		failures = append(failures, "agent used browser automation")
	}
	return failures
}
