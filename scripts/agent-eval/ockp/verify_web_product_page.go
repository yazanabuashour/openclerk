package main

import (
	"context"
	"fmt"
	"strings"
)

func verifyWebProductPageNatural(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	noDocument := verifyNoDocument(ctx, paths, webProductPageSourcePath, "natural product-page request should not write without source.path_hint approval")
	failures := webProductPageBypassFailures(turnMetrics)
	if !noDocument.DatabasePass {
		failures = append(failures, noDocument.Details)
	}
	if !isMissingFieldClarification(finalMessage, []string{"source.path_hint"}) {
		failures = append(failures, "answer did not name source.path_hint as missing")
	}
	if !messageContainsAny(finalMessage, []string{"durable", "write", "approval", "approve"}) ||
		!messageContainsAny(finalMessage, []string{"public", "fetch"}) {
		failures = append(failures, "answer did not distinguish public fetch permission from durable-write approval")
	}
	if !productPageBoundaryDisclosure(finalMessage) {
		failures = append(failures, "answer did not preserve product-page browser or purchase boundaries")
	}
	metricsPass := turnMetrics.ToolCalls == 0 && turnMetrics.CommandExecutions == 0 && turnMetrics.AssistantCalls <= 1
	if !metricsPass {
		failures = append(failures, fmt.Sprintf("expected no tools and at most one assistant answer, got tools=%d commands=%d assistant=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions, turnMetrics.AssistantCalls))
	}
	assistantPass := isMissingFieldClarification(finalMessage, []string{"source.path_hint"}) &&
		messageContainsAny(finalMessage, []string{"durable", "write", "approval", "approve"}) &&
		messageContainsAny(finalMessage, []string{"public", "fetch"}) &&
		productPageBoundaryDisclosure(finalMessage)
	return verificationResult{
		Passed:        noDocument.DatabasePass && assistantPass && metricsPass && len(webProductPageBypassFailures(turnMetrics)) == 0,
		DatabasePass:  noDocument.DatabasePass && metricsPass,
		AssistantPass: assistantPass && metricsPass && len(webProductPageBypassFailures(turnMetrics)) == 0,
		Details:       missingDetails(failures),
		Documents:     []string{webProductPageSourcePath},
	}, nil
}

func verifyWebProductPageControl(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, webProductPageSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := webProductPageBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing rich product-page source document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{"source_type: web", "source_url:", webProductPageText, webProductPageVariantText, "Add to cart"})...)
		if doc.Metadata["source_type"] != "web" {
			failures = append(failures, fmt.Sprintf("expected source_type web, got %q", doc.Metadata["source_type"]))
		}
		if doc.Metadata["asset_path"] != "" {
			failures = append(failures, "web product-page source recorded asset_path metadata")
		}
		if strings.Contains(doc.Body, webProductPageHiddenDynamicText) {
			failures = append(failures, "script-rendered hidden dynamic product text was captured as visible evidence")
		}
	}
	if !turnMetrics.IngestSourceURLUsed || turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use create-mode ingest_source_url")
	}
	assistantPass := messageContainsAll(finalMessage, []string{webProductPageSourcePath, webProductPageText, webProductPageVariantText}) &&
		messageContainsAny(finalMessage, []string{"Add to cart", "cart"}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"}) &&
		productPageBoundaryDisclosure(finalMessage)
	if !assistantPass {
		failures = append(failures, "final answer did not report product-page evidence, citation, and no-browser/no-purchase boundaries")
	}
	databasePass := found && doc != nil &&
		doc.Metadata["source_type"] == "web" &&
		doc.Metadata["asset_path"] == "" &&
		strings.Contains(doc.Body, webProductPageText) &&
		strings.Contains(doc.Body, webProductPageVariantText) &&
		strings.Contains(doc.Body, "Add to cart") &&
		!strings.Contains(doc.Body, webProductPageHiddenDynamicText)
	activityPass := len(webProductPageBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUsed && !turnMetrics.IngestSourceURLUpdateUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webProductPageSourcePath},
	}, nil
}

func verifyWebProductPageDuplicate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	sourceCount, err := exactDocumentCount(ctx, paths, webProductPageSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, webProductPageDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := webProductPageBypassFailures(turnMetrics)
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing product-page source, got %d", sourceCount))
	}
	if duplicateCount != 0 {
		failures = append(failures, "created duplicate product-page source "+webProductPageDuplicatePath)
	}
	if !turnMetrics.IngestSourceURLUsed {
		failures = append(failures, "agent did not exercise duplicate ingest_source_url request")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list product-page source candidates after duplicate rejection")
	}
	assistantPass := messageContainsAll(finalMessage, []string{webProductPageSourcePath, webProductPageDuplicatePath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "already", "rejected", "normalized"}) &&
		messageContainsAny(finalMessage, []string{"not created", "was not created", "no copy", "no durable"})
	if !assistantPass {
		failures = append(failures, "final answer did not report normalized duplicate rejection and no-copy outcome")
	}
	databasePass := sourceCount == 1 && duplicateCount == 0
	activityPass := len(webProductPageBypassFailures(turnMetrics)) == 0 && turnMetrics.IngestSourceURLUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webProductPageSourcePath, webProductPageDuplicatePath},
	}, nil
}

func verifyWebProductPageDynamic(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, webProductPageDynamicPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := webProductPageBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing dynamic product-page source document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{webProductPageText, webProductPageVariantText})...)
		if strings.Contains(doc.Body, webProductPageHiddenDynamicText) {
			failures = append(failures, "script-rendered hidden dynamic product text was captured as visible evidence")
		}
	}
	if !turnMetrics.IngestSourceURLUsed || !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not ingest and search runner-visible product-page evidence")
	}
	assistantPass := messageContainsAll(finalMessage, []string{webProductPageDynamicPath, webProductPageText, webProductPageVariantText}) &&
		productPageDynamicOmissionDisclosure(finalMessage) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"})
	if !assistantPass {
		failures = append(failures, "final answer did not disclose dynamic omission with citation evidence")
	}
	databasePass := found && doc != nil &&
		strings.Contains(doc.Body, webProductPageText) &&
		strings.Contains(doc.Body, webProductPageVariantText) &&
		!strings.Contains(doc.Body, webProductPageHiddenDynamicText)
	activityPass := len(webProductPageBypassFailures(turnMetrics)) == 0 && turnMetrics.IngestSourceURLUsed && turnMetrics.SearchUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webProductPageDynamicPath},
	}, nil
}

func verifyWebProductPageUnsupported(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	count, err := exactDocumentCount(ctx, paths, webProductPageUnsupportedPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := webProductPageBypassFailures(turnMetrics)
	if count != 0 {
		failures = append(failures, "created unsupported product-page source")
	}
	if !turnMetrics.IngestSourceURLUsed {
		failures = append(failures, "agent did not exercise unsupported ingest_source_url request")
	}
	assistantPass := messageContainsAny(finalMessage, []string{"unsupported", "non-html", "HTML", "content type", "rejected"}) &&
		messageContainsAny(finalMessage, []string{"not created", "no durable", "was not created"})
	if !assistantPass {
		failures = append(failures, "final answer did not report non-HTML rejection without write")
	}
	databasePass := count == 0
	activityPass := len(webProductPageBypassFailures(turnMetrics)) == 0 && turnMetrics.IngestSourceURLUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{webProductPageUnsupportedPath},
	}, nil
}

func webProductPageBypassFailures(turnMetrics metrics) []string {
	failures := webURLBypassFailures(turnMetrics)
	if turnMetrics.ManualHTTPFetch {
		failures = append(failures, "agent used manual HTTP fetch")
	}
	if turnMetrics.BrowserAutomation {
		failures = append(failures, "agent used browser automation")
	}
	return failures
}

func productPageBoundaryDisclosure(message string) bool {
	lower := normalizeValidationMessage(message)
	return clauseRejectsAny(lower, []string{"browser", "browser automation"}) &&
		clauseRejectsAny(lower, []string{"login", "log in"}) &&
		clauseRejectsAny(lower, []string{"account", "account state"}) &&
		clauseRejectsAny(lower, []string{"cart"}) &&
		clauseRejectsAny(lower, []string{"checkout"}) &&
		clauseRejectsAny(lower, []string{"purchase", "buy"})
}

func productPageDynamicOmissionDisclosure(message string) bool {
	lower := normalizeValidationMessage(message)
	return clauseRejectsAny(lower, []string{"dynamic content", "dynamic or script", "script-rendered", "script rendered", "script-rendered content", "script rendered content", "script content"})
}

func clauseRejectsAny(lower string, terms []string) bool {
	for _, clause := range strings.FieldsFunc(lower, func(r rune) bool {
		return r == '.' || r == ';' || r == '\n' || r == '!' || r == '?'
	}) {
		if containsAny(clause, terms) && containsAny(clause, []string{
			"no ",
			"without",
			"not allowed",
			"unsupported",
			"not supported",
			"cannot",
			"can't",
			"do not",
			"did not",
			"not use",
			"not using",
			"was not",
			"were not",
			"not acquired",
			"not captured",
			"not fetched",
			"not collected",
			"omitted",
			"unavailable",
		}) {
			return true
		}
	}
	return false
}
