package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func seedDocumentArtifactCandidateDuplicate(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: note
status: active
---
# Existing Pricing Model Note

## Summary
Candidate generation duplicate pricing model marker.
The pricing model note already captures packaging tiers and renewal notes.
`) + "\n"
	return createSeedDocument(ctx, cfg, candidateDuplicateExistingPath, "Existing Pricing Model Note", body)
}
func seedArtifactTranscript(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: transcript
status: active
artifact_kind: transcript
---
# Vendor Demo Transcript

## Summary
Artifact transcript canonical markdown evidence: vendor demo transcript says agents may store transcripts as canonical markdown when the transcript text is already supplied.

## Excerpt
Speaker A: Keep transcript artifacts citeable through document search.
Speaker B: Do not require native audio or video ingestion for pasted transcript text.
`) + "\n"
	return createSeedDocument(ctx, cfg, artifactTranscriptPath, "Vendor Demo Transcript", body)
}
func seedArtifactInvoiceReceipt(ctx context.Context, cfg runclient.Config) error {
	invoiceBody := strings.TrimSpace(`---
type: invoice
status: active
artifact_kind: invoice
vendor: Atlas Platform
total_usd: "1250.00"
---
# Atlas Platform April Invoice

## Summary
Artifact invoice receipt authority evidence: Atlas Platform invoice total is USD 1250.00 and requires approval above USD 500.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, artifactInvoicePath, "Atlas Platform April Invoice", invoiceBody); err != nil {
		return err
	}
	receiptBody := strings.TrimSpace(`---
type: receipt
status: active
artifact_kind: receipt
vendor: Nebula Office
total_usd: "86.40"
---
# Nebula USB-C Hub Receipt

## Summary
Artifact invoice receipt authority evidence: Nebula USB-C Hub receipt total is USD 86.40.
`) + "\n"
	return createSeedDocument(ctx, cfg, artifactReceiptPath, "Nebula USB-C Hub Receipt", receiptBody)
}
func seedArtifactMixedSynthesis(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/artifacts/mixed-current.md
artifact_kind: mixed
---
# Mixed Artifact Old Source

## Summary
Older mixed artifact source said artifact ingestion should prefer duplicate synthesis pages.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, artifactMixedSynthesisOldPath, "Mixed Artifact Old Source", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
type: source
status: active
supersedes: sources/artifacts/mixed-old.md
artifact_kind: mixed
---
# Mixed Artifact Current Source

## Summary
Artifact mixed synthesis freshness evidence: current mixed artifacts should update existing source-linked synthesis and preserve citations, provenance, and freshness.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, artifactMixedSynthesisCurrentPath, "Mixed Artifact Current Source", currentBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/artifacts/mixed-old.md
---
# Artifact Ingestion Pressure

## Summary
Stale mixed artifact synthesis says duplicate synthesis pages are acceptable.

## Sources
- sources/artifacts/mixed-old.md

## Freshness
Fresh before heterogeneous artifact ingestion pressure checks.
`) + "\n"
	return createSeedDocument(ctx, cfg, artifactMixedSynthesisPath, "Artifact Ingestion Pressure", synthesisBody)
}
func seedVideoYouTubeSynthesisFreshness(ctx context.Context, cfg runclient.Config) error {
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestVideoURL,
		Video: runner.VideoURLInput{
			URL:      videoYouTubeURL,
			PathHint: videoYouTubeCurrentSourcePath,
			Title:    "Platform Demo Current Transcript",
			Transcript: runner.VideoTranscriptInput{
				Text:       videoYouTubeSynthesisCurrentEvidenceText + ": current transcript source notes must preserve transcript provenance, citations, and freshness before source-linked synthesis is trusted.",
				Policy:     "supplied",
				Origin:     videoYouTubeTranscriptOrigin,
				Language:   "en",
				CapturedAt: "2026-04-27T00:00:00Z",
			},
		},
	})
	if err != nil {
		return err
	}
	if result.Rejected || result.VideoIngestion == nil {
		return fmt.Errorf("seed video source ingestion failed: %+v", result)
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/video-youtube/platform-demo-current.md
---
# Video YouTube Ingestion Pressure

## Summary
Fresh video synthesis cites the current transcript before update pressure.

## Sources
- sources/video-youtube/platform-demo-current.md

## Freshness
Fresh before video/YouTube ingestion pressure checks.
`) + "\n"
	return createSeedDocument(ctx, cfg, videoYouTubeSynthesisPath, "Video YouTube Ingestion Pressure", synthesisBody)
}
