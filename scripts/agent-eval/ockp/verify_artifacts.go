package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

type documentArtifactCandidateExpectation struct {
	Path             string
	Title            string
	RequiredBody     []string
	ForbiddenBody    []string
	RequireValidate  bool
	RequireNoCreate  bool
	RequireApproval  bool
	RequireBodyShown bool
}

func verifyDocumentArtifactCandidateProposal(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, expectation documentArtifactCandidateExpectation) (verificationResult, error) {
	count, err := exactDocumentCount(ctx, paths, expectation.Path)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentArtifactCandidateBypassFailures(turnMetrics)
	if expectation.RequireNoCreate && count != 0 {
		failures = append(failures, fmt.Sprintf("created candidate document %s before approval", expectation.Path))
	}
	preApprovalWriteUsed := turnMetrics.CreateDocumentUsed || turnMetrics.AppendDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.IngestSourceURLUsed || turnMetrics.IngestVideoURLUsed
	if expectation.RequireNoCreate && preApprovalWriteUsed {
		failures = append(failures, "used create_document, append_document, replace_section, ingest_source_url, or ingest_video_url before approval")
	} else if turnMetrics.CreateDocumentUsed {
		failures = append(failures, "used create_document before approval")
	}
	if expectation.RequireValidate && !turnMetrics.ValidateUsed {
		failures = append(failures, "did not validate strict candidate document JSON")
	}
	if expectation.RequireValidate && (turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0) {
		failures = append(failures, "did not run installed runner validation")
	}
	assistantRequired := []string{expectation.Path, expectation.Title}
	if expectation.RequireBodyShown {
		assistantRequired = append(assistantRequired, expectation.RequiredBody...)
	}
	if !messageContainsAll(finalMessage, assistantRequired) {
		failures = append(failures, "final answer did not include candidate path, title, and required body preview")
	}
	if len(presentForbidden(strings.ToLower(finalMessage), lowerStrings(expectation.ForbiddenBody))) != 0 {
		failures = append(failures, "final answer included forbidden invented body content")
	}
	if expectation.RequireApproval && !messageContainsAny(finalMessage, []string{"confirm", "confirmation", "approve", "approval", "before creating", "before I create"}) {
		failures = append(failures, "final answer did not ask for confirmation before creating")
	}
	if expectation.RequireNoCreate && !messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "before creating"}) {
		failures = append(failures, "final answer did not state that no document was created before approval")
	}
	databasePass := !expectation.RequireNoCreate || count == 0
	activityPass := len(documentArtifactCandidateBypassFailures(turnMetrics)) == 0 &&
		!preApprovalWriteUsed &&
		(!expectation.RequireValidate || (turnMetrics.ValidateUsed && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0))
	assistantPass := messageContainsAll(finalMessage, assistantRequired) &&
		(!expectation.RequireApproval || messageContainsAny(finalMessage, []string{"confirm", "confirmation", "approve", "approval", "before creating", "before I create"})) &&
		(!expectation.RequireNoCreate || messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "before creating"})) &&
		len(presentForbidden(strings.ToLower(finalMessage), lowerStrings(expectation.ForbiddenBody))) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{expectation.Path},
	}, nil
}
func verifyDocumentArtifactCandidateDuplicateRisk(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingCount, err := exactDocumentCount(ctx, paths, candidateDuplicateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, candidateDuplicateCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentArtifactCandidateBypassFailures(turnMetrics)
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one seeded duplicate candidate %s, got %d", candidateDuplicateExistingPath, existingCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created duplicate candidate "+candidateDuplicateCandidatePath)
	}
	if turnMetrics.CreateDocumentUsed {
		failures = append(failures, "used create_document despite duplicate risk")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "did not search for duplicate risk")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "did not list candidate documents")
	}
	assistantPass := messageContainsAll(finalMessage, []string{candidateDuplicateExistingPath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already"}) &&
		messageContainsAny(finalMessage, []string{"confirm", "choose", "update", "create new", "approval"}) &&
		messageContainsAny(finalMessage, []string{"no document was created", "not created", "did not create", "no new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate risk and ask before writing")
	}
	databasePass := existingCount == 1 && candidateCount == 0
	activityPass := len(documentArtifactCandidateBypassFailures(turnMetrics)) == 0 &&
		!turnMetrics.CreateDocumentUsed &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{candidateDuplicateExistingPath, candidateDuplicateCandidatePath},
	}, nil
}
func verifyDocumentArtifactCandidateLowConfidence(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	counts := []string{candidateNotePath, candidateHeadingPath, candidateMixedSourcePath, candidateOverridePath, candidateBodyFaithfulnessPath}
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
	failures := documentArtifactCandidateBypassFailures(turnMetrics)
	if len(created) != 0 {
		failures = append(failures, "created low-confidence candidate documents: "+strings.Join(created, ", "))
	}
	if turnMetrics.ToolCalls != 0 || turnMetrics.CommandExecutions != 0 || turnMetrics.AssistantCalls > 1 {
		failures = append(failures, fmt.Sprintf("low-confidence ask should be no-tools, got tools=%d commands=%d assistant=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions, turnMetrics.AssistantCalls))
	}
	assistantPass := messageContainsAny(finalMessage, []string{"body", "content", "text", "artifact type", "what to document"}) &&
		messageContainsAny(finalMessage, []string{"missing", "provide", "need", "can't create", "cannot create"}) &&
		!messageContainsAny(finalMessage, []string{candidateNotePath, candidateHeadingPath, candidateMixedSourcePath})
	if !assistantPass {
		failures = append(failures, "final answer did not ask for missing content or intent without proposing a path")
	}
	databasePass := len(created) == 0
	activityPass := len(documentArtifactCandidateBypassFailures(turnMetrics)) == 0 &&
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

type artifactPDFExpectation struct {
	SourcePath string
	AssetPath  string
}

func artifactPDFExpectedPaths(scenarioID string) artifactPDFExpectation {
	if scenarioID == artifactPDFNaturalIntentScenarioID {
		return artifactPDFExpectation{
			SourcePath: artifactPDFNaturalSourcePath,
			AssetPath:  artifactPDFNaturalAssetPath,
		}
	}
	return artifactPDFExpectation{
		SourcePath: artifactPDFSourcePath,
		AssetPath:  artifactPDFAssetPath,
	}
}
func verifyArtifactPDFSourceURL(ctx context.Context, paths evalPaths, scenarioID string, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	expectation := artifactPDFExpectedPaths(scenarioID)
	doc, found, err := documentByPath(ctx, paths, expectation.SourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	count, err := exactDocumentCount(ctx, paths, expectation.SourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := artifactIngestionBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing PDF source document")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{artifactPDFEvidenceText, "source_url:", "asset_path:", expectation.AssetPath})...)
		if doc.Metadata["asset_path"] != expectation.AssetPath {
			failures = append(failures, fmt.Sprintf("expected asset_path metadata %q, got %q", expectation.AssetPath, doc.Metadata["asset_path"]))
		}
		if doc.Metadata["source_type"] != "pdf" {
			failures = append(failures, fmt.Sprintf("expected source_type metadata pdf, got %q", doc.Metadata["source_type"]))
		}
	}
	if count != 1 {
		failures = append(failures, fmt.Sprintf("expected one PDF source document, got %d", count))
	}
	if !turnMetrics.IngestSourceURLUsed || turnMetrics.IngestSourceURLUpdateUsed {
		failures = append(failures, "agent did not use default create-mode ingest_source_url")
	}
	assistantPass := messageContainsAll(finalMessage, []string{expectation.SourcePath, expectation.AssetPath}) &&
		messageContainsAny(finalMessage, []string{"citation", "citations", "doc_id", "chunk_id"}) &&
		messageContainsAny(finalMessage, []string{"ingested", "created", "source URL"})
	if !assistantPass {
		failures = append(failures, "final answer did not report PDF source ingestion with citation evidence")
	}
	databasePass := found && doc != nil && count == 1 &&
		doc.Metadata["asset_path"] == expectation.AssetPath &&
		doc.Metadata["source_type"] == "pdf" &&
		len(missingRequired(doc.Body, []string{artifactPDFEvidenceText, "source_url:", "asset_path:", expectation.AssetPath})) == 0
	activityPass := len(artifactIngestionBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestSourceURLUsed && !turnMetrics.IngestSourceURLUpdateUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{expectation.SourcePath},
	}, nil
}
func verifyArtifactTranscript(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, artifactTranscriptPath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, artifactTranscriptEvidenceText)
	if err != nil {
		return verificationResult{}, err
	}
	failures := artifactIngestionBypassFailures(turnMetrics)
	if !found || doc == nil {
		failures = append(failures, "missing transcript fixture")
	} else {
		failures = append(failures, missingRequired(doc.Body, []string{artifactTranscriptEvidenceText, "artifact_kind: transcript"})...)
	}
	if !searchContainsPath(search, artifactTranscriptPath) || !searchResultHasCitations(search) {
		failures = append(failures, "transcript search did not expose citation-bearing result")
	}
	if !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"transcripts/"}) {
		failures = append(failures, "agent did not search transcript artifact evidence with path_prefix transcripts/")
	}
	assistantPass := messageContainsAll(finalMessage, []string{artifactTranscriptPath}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"}) &&
		messageContainsAny(finalMessage, []string{"canonical markdown", "transcript"})
	if !assistantPass {
		failures = append(failures, "final answer did not cite transcript canonical markdown evidence")
	}
	databasePass := found && doc != nil &&
		len(missingRequired(doc.Body, []string{artifactTranscriptEvidenceText, "artifact_kind: transcript"})) == 0 &&
		searchContainsPath(search, artifactTranscriptPath) && searchResultHasCitations(search)
	activityPass := len(artifactIngestionBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"transcripts/"})
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{artifactTranscriptPath},
	}, nil
}
func verifyArtifactInvoiceReceipt(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	invoice, invoiceFound, err := documentByPath(ctx, paths, artifactInvoicePath)
	if err != nil {
		return verificationResult{}, err
	}
	receipt, receiptFound, err := documentByPath(ctx, paths, artifactReceiptPath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, artifactInvoiceReceiptEvidenceText)
	if err != nil {
		return verificationResult{}, err
	}
	failures := artifactIngestionBypassFailures(turnMetrics)
	if !invoiceFound || invoice == nil {
		failures = append(failures, "missing invoice fixture")
	} else {
		failures = append(failures, missingRequired(invoice.Body, []string{"USD 1250.00", "approval above USD 500"})...)
	}
	if !receiptFound || receipt == nil {
		failures = append(failures, "missing receipt fixture")
	} else {
		failures = append(failures, missingRequired(receipt.Body, []string{"USD 86.40"})...)
	}
	if !searchContainsPath(search, artifactInvoicePath) || !searchContainsPath(search, artifactReceiptPath) || !searchResultHasCitations(search) {
		failures = append(failures, "invoice/receipt search did not expose citation-bearing authority results")
	}
	requiredMetadataFilters := []string{"artifact_kind=invoice", "artifact_kind=receipt"}
	if !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchMetadataFilters, requiredMetadataFilters) {
		failures = append(failures, "agent did not run invoice and receipt artifact_kind metadata-filtered retrieval")
	}
	assistantPass := messageContainsAll(finalMessage, []string{artifactInvoicePath, artifactReceiptPath, "USD 1250.00", "USD 86.40"}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"})
	if !assistantPass {
		failures = append(failures, "final answer did not cite invoice and receipt authority evidence")
	}
	databasePass := invoiceFound && invoice != nil && receiptFound && receipt != nil &&
		len(missingRequired(invoice.Body, []string{"USD 1250.00", "approval above USD 500"})) == 0 &&
		len(missingRequired(receipt.Body, []string{"USD 86.40"})) == 0 &&
		searchContainsPath(search, artifactInvoicePath) && searchContainsPath(search, artifactReceiptPath) &&
		searchResultHasCitations(search)
	activityPass := len(artifactIngestionBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		containsAllStrings(turnMetrics.SearchMetadataFilters, requiredMetadataFilters)
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{artifactInvoicePath, artifactReceiptPath},
	}, nil
}
func verifyArtifactMixedSynthesis(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	synthesis, synthesisFound, err := documentByPath(ctx, paths, artifactMixedSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	current, currentFound, err := documentByPath(ctx, paths, artifactMixedSynthesisCurrentPath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, artifactMixedSynthesisEvidenceText)
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := artifactProjectionStates(ctx, paths, docIDOrEmpty(synthesis))
	if err != nil {
		return verificationResult{}, err
	}
	failures := artifactIngestionBypassFailures(turnMetrics)
	if !synthesisFound || synthesis == nil {
		failures = append(failures, "missing mixed synthesis fixture")
	} else {
		failures = append(failures, missingRequired(synthesis.Body, []string{artifactMixedSynthesisOldPath, "source_refs:"})...)
	}
	if !currentFound || current == nil {
		failures = append(failures, "missing current mixed artifact source")
	}
	if !searchContainsPath(search, artifactMixedSynthesisCurrentPath) || !searchResultHasCitations(search) {
		failures = append(failures, "mixed artifact current source search did not expose citation-bearing result")
	}
	if !projectionListContainsStaleSource(projections, artifactMixedSynthesisCurrentPath) && !projectionListContainsStaleSource(projections, artifactMixedSynthesisOldPath) {
		failures = append(failures, "synthesis projection did not expose stale or missing current mixed source")
	}
	if !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProjectionStatesUsed || !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect search/list/get/provenance/projection evidence for mixed synthesis")
	}
	assistantPass := messageContainsAll(finalMessage, []string{artifactMixedSynthesisPath, artifactMixedSynthesisOldPath, artifactMixedSynthesisCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"stale", "freshness", "projection"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source refs", "source_refs"})
	if !assistantPass {
		failures = append(failures, "final answer did not explain mixed artifact synthesis freshness and provenance")
	}
	databasePass := synthesisFound && synthesis != nil && currentFound && current != nil &&
		searchContainsPath(search, artifactMixedSynthesisCurrentPath) && searchResultHasCitations(search) &&
		(projectionListContainsStaleSource(projections, artifactMixedSynthesisCurrentPath) || projectionListContainsStaleSource(projections, artifactMixedSynthesisOldPath))
	activityPass := len(artifactIngestionBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed && turnMetrics.ProvenanceEventsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{artifactMixedSynthesisPath, artifactMixedSynthesisOldPath, artifactMixedSynthesisCurrentPath},
	}, nil
}
func verifyVideoYouTubeScriptedTranscript(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, videoYouTubeSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	count, err := exactDocumentCount(ctx, paths, videoYouTubeSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, videoYouTubeSourceEvidenceText)
	if err != nil {
		return verificationResult{}, err
	}
	failures := videoYouTubeBypassFailures(turnMetrics)
	requiredBody := []string{
		videoYouTubeSourceEvidenceText,
		"source_type: video_transcript",
		"source_url:",
		videoYouTubeURL,
		"transcript_origin:",
		videoYouTubeTranscriptOrigin,
		"transcript_sha256:",
		"## Transcript",
	}
	if !found || doc == nil {
		failures = append(failures, "missing video/YouTube canonical source note")
	} else {
		failures = append(failures, missingRequired(doc.Body, requiredBody)...)
		if doc.Metadata["source_type"] != "video_transcript" {
			failures = append(failures, fmt.Sprintf("expected source_type metadata video_transcript, got %q", doc.Metadata["source_type"]))
		}
		if doc.Metadata["source_url"] != videoYouTubeURL {
			failures = append(failures, fmt.Sprintf("expected source_url metadata %q, got %q", videoYouTubeURL, doc.Metadata["source_url"]))
		}
		if doc.Metadata["transcript_origin"] != videoYouTubeTranscriptOrigin {
			failures = append(failures, fmt.Sprintf("expected transcript_origin metadata %q, got %q", videoYouTubeTranscriptOrigin, doc.Metadata["transcript_origin"]))
		}
	}
	if count != 1 {
		failures = append(failures, fmt.Sprintf("expected one video/YouTube source document, got %d", count))
	}
	if !searchContainsPath(search, videoYouTubeSourcePath) || !searchResultHasCitations(search) {
		failures = append(failures, "video/YouTube transcript search did not expose citation-bearing source result")
	}
	if !turnMetrics.IngestVideoURLUsed || turnMetrics.IngestVideoURLUpdateUsed || !turnMetrics.SearchUsed || !containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"sources/video-youtube/"}) {
		failures = append(failures, "agent did not use create-mode ingest_video_url and then retrieve the canonical video/YouTube source note with path_prefix sources/video-youtube/")
	}
	assistantPass := messageContainsAll(finalMessage, []string{videoYouTubeSourcePath, videoYouTubeURL}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "transcript_origin", "transcript provenance"})
	if !assistantPass {
		failures = append(failures, "final answer did not report source path, citation evidence, and transcript provenance")
	}
	databasePass := found && doc != nil && count == 1 &&
		len(missingRequired(doc.Body, requiredBody)) == 0 &&
		doc.Metadata["source_type"] == "video_transcript" &&
		doc.Metadata["source_url"] == videoYouTubeURL &&
		doc.Metadata["transcript_origin"] == videoYouTubeTranscriptOrigin &&
		searchContainsPath(search, videoYouTubeSourcePath) &&
		searchResultHasCitations(search)
	activityPass := len(videoYouTubeBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestVideoURLUsed &&
		!turnMetrics.IngestVideoURLUpdateUsed &&
		turnMetrics.SearchUsed &&
		containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"sources/video-youtube/"})
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{videoYouTubeSourcePath},
	}, nil
}
func verifyVideoYouTubeSynthesisFreshness(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	synthesis, synthesisFound, err := documentByPath(ctx, paths, videoYouTubeSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	current, currentFound, err := documentByPath(ctx, paths, videoYouTubeCurrentSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	search, err := artifactSearch(ctx, paths, "transcript")
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := artifactProjectionStates(ctx, paths, docIDOrEmpty(synthesis))
	if err != nil {
		return verificationResult{}, err
	}
	failures := videoYouTubeBypassFailures(turnMetrics)
	if !synthesisFound || synthesis == nil {
		failures = append(failures, "missing video/YouTube synthesis fixture")
	} else {
		failures = append(failures, missingRequired(synthesis.Body, []string{videoYouTubeCurrentSourcePath, "source_refs:"})...)
	}
	if !currentFound || current == nil {
		failures = append(failures, "missing current video/YouTube source fixture")
	} else if current.Metadata["captured_at"] != "2026-04-27T01:00:00Z" {
		failures = append(failures, "current video/YouTube source was not updated to the changed transcript capture time")
	}
	if !searchContainsPath(search, videoYouTubeCurrentSourcePath) || !searchResultHasCitations(search) {
		failures = append(failures, "video/YouTube source search did not expose citation-bearing result after update")
	}
	if !projectionListContainsStaleSource(projections, videoYouTubeCurrentSourcePath) {
		failures = append(failures, "synthesis projection did not expose stale current video/YouTube source after changed transcript update")
	}
	if !turnMetrics.IngestVideoURLUpdateUsed || !turnMetrics.SearchUsed || !turnMetrics.ListDocumentsUsed || !turnMetrics.GetDocumentUsed || !turnMetrics.ProjectionStatesUsed || !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not update supplied transcript and inspect search/list/get/provenance/projection evidence for video/YouTube synthesis")
	}
	if turnMetrics.CreateDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.AppendDocumentUsed {
		failures = append(failures, "agent mutated synthesis during video/YouTube source update freshness inspection")
	}
	assistantPass := messageContainsAll(finalMessage, []string{videoYouTubeSynthesisPath, videoYouTubeCurrentSourcePath}) &&
		messageContainsAny(finalMessage, []string{"stale", "freshness", "projection"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source refs", "source_refs"}) &&
		messageContainsAny(finalMessage, []string{"no-op", "same hash", "same transcript", "changed transcript", "updated transcript"})
	if !assistantPass {
		failures = append(failures, "final answer did not explain no-op/update video/YouTube synthesis freshness and provenance")
	}
	databasePass := synthesisFound && synthesis != nil && currentFound && current != nil &&
		current.Metadata["captured_at"] == "2026-04-27T01:00:00Z" &&
		searchContainsPath(search, videoYouTubeCurrentSourcePath) &&
		searchResultHasCitations(search) &&
		projectionListContainsStaleSource(projections, videoYouTubeCurrentSourcePath)
	activityPass := len(videoYouTubeBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.IngestVideoURLUpdateUsed &&
		turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed && turnMetrics.ProvenanceEventsUsed &&
		!turnMetrics.CreateDocumentUsed && !turnMetrics.ReplaceSectionUsed && !turnMetrics.AppendDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{videoYouTubeSynthesisPath, videoYouTubeCurrentSourcePath},
	}, nil
}
func artifactSearch(ctx context.Context, paths evalPaths, text string) (runner.RetrievalTaskResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	return runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  text,
			Limit: 10,
		},
	})
}
func artifactProjectionStates(ctx context.Context, paths evalPaths, synthesisDocID string) (runner.ProjectionStateList, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      synthesisDocID,
			Limit:      5,
		},
	})
	if err != nil {
		return runner.ProjectionStateList{}, err
	}
	if result.Projections == nil {
		return runner.ProjectionStateList{}, nil
	}
	return *result.Projections, nil
}
func projectionListContainsStaleSource(list runner.ProjectionStateList, path string) bool {
	for _, projection := range list.Projections {
		if projection.Freshness == "stale" &&
			(projectionDetailContains(projection.Details, "stale_source_refs", path) ||
				projectionDetailContains(projection.Details, "current_source_refs", path) ||
				projectionDetailContains(projection.Details, "missing_source_refs", path)) {
			return true
		}
	}
	return false
}
func agentChosenBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func pathTitleBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func captureLowRiskBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func captureExplicitOverridesBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func captureDuplicateCandidateBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func captureSaveThisNoteBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func captureDocumentLinksBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func documentThisBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func documentArtifactCandidateBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func artifactIngestionBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
func videoYouTubeBypassFailures(turnMetrics metrics) []string {
	return populatedBypassFailures(turnMetrics)
}
