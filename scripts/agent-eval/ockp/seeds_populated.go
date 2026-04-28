package main

import (
	"context"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func seedPopulatedVaultFixture(ctx context.Context, cfg runclient.Config) error {
	docs := []struct {
		path  string
		title string
		body  string
	}{
		{populatedTranscriptPath, "Atlas Kickoff Transcript", strings.TrimSpace(`---
type: transcript
status: active
project: atlas
---
# Atlas Kickoff Transcript

## Summary
The kickoff transcript mentions the Atlas project, Nebula Consulting, the reimbursement threshold, and the privacy addendum review.

## Notes
Participants said the authoritative operational summary lives in the populated Atlas authority source.
`) + "\n"},
		{populatedTranscriptOpsPath, "Atlas Ops Standup Transcript", strings.TrimSpace(`---
type: transcript
status: active
project: atlas
---
# Atlas Ops Standup Transcript

## Summary
The ops standup repeats that Atlas questions should reconcile receipt totals, invoice thresholds, legal retention notes, and Acme contract controls through runner-visible sources.

## Notes
Speakers mentioned Nebula Office Supply, Nebula Consulting, and Acme in the same agenda so retrieval has overlapping entities across document families.
`) + "\n"},
		{populatedArticlePath, "Vendor Risk Review", strings.TrimSpace(`---
type: article
status: active
project: atlas
---
# Vendor Risk Review

## Summary
The vendor risk article says Atlas should prefer the current authority source when invoices, receipts, contracts, and legal notes disagree.
`) + "\n"},
		{populatedArticleArchivePath, "Vendor Risk Review Archive", strings.TrimSpace(`---
type: article
status: archived
project: atlas
---
# Vendor Risk Review Archive

## Summary
Populated vault stale source marker: the archived vendor risk review said Atlas could approve Nebula invoices without the current authority review.
`) + "\n"},
		{populatedMeetingPath, "Atlas Weekly Review", strings.TrimSpace(`---
type: meeting-note
status: active
project: atlas
---
# Atlas Weekly Review

## Summary
The review links Nebula Consulting invoice approval, Acme contract controls, and receipt reimbursement into one Atlas workstream.
`) + "\n"},
		{populatedMeetingBudgetPath, "Atlas Budget Sync", strings.TrimSpace(`---
type: meeting-note
status: active
project: atlas
---
# Atlas Budget Sync

## Summary
The budget sync compares the Nebula Office Supply receipt with the Nebula Consulting invoice and asks agents to cite source paths before summarizing totals.
`) + "\n"},
		{populatedDocsPath, "Atlas Operations Guide", strings.TrimSpace(`---
type: reference-doc
status: active
project: atlas
---
# Atlas Operations Guide

## Summary
Atlas operations require source-grounded answers with path, doc_id, chunk_id, heading, or line citation details.
`) + "\n"},
		{populatedDocsRunbookPath, "Atlas Vendor Runbook", strings.TrimSpace(`---
type: reference-doc
status: active
project: atlas
---
# Atlas Vendor Runbook

## Summary
The vendor runbook says canonical markdown remains the source of truth for Atlas receipts, invoices, contracts, and legal notes until a future typed domain is promoted.
`) + "\n"},
		{populatedBlogPath, "Atlas Launch Draft", strings.TrimSpace(`---
type: blog-draft
status: draft
project: atlas
---
# Atlas Launch Draft

## Summary
This draft is intentionally lower authority and should not override current source documents.
`) + "\n"},
		{populatedBlogRumorPath, "Atlas Launch Rumor", strings.TrimSpace(`---
type: blog-draft
status: polluted
project: atlas
---
# Atlas Launch Rumor

## Summary
This polluted blog draft incorrectly claims the Acme privacy addendum can be skipped and should not be used as authority.
`) + "\n"},
		{populatedReceiptPath, "Nebula Office Supply Receipt", strings.TrimSpace(`---
type: receipt
status: active
vendor: nebula-office-supply
project: atlas
---
# Nebula Office Supply Receipt

## Summary
Receipt marker: Atlas reimbursable supplies from Nebula Office Supply total USD 118.42.
`) + "\n"},
		{populatedReceiptDuplicatePath, "Nebula Office Supply Receipt Copy", strings.TrimSpace(`---
type: receipt
status: duplicate
vendor: nebula-office-supply
project: atlas
duplicates: receipts/nebula-office-supply.md
---
# Nebula Office Supply Receipt Copy

## Summary
Populated vault duplicate candidate marker: this duplicate-looking receipt repeats the USD 118.42 total but points back to the canonical Nebula Office Supply receipt.
`) + "\n"},
		{populatedInvoicePath, "Nebula Consulting Invoice April 2026", strings.TrimSpace(`---
type: invoice
status: active
vendor: nebula-consulting
project: atlas
---
# Nebula Consulting Invoice April 2026

## Summary
Invoice marker: Nebula Consulting invoice NC-2026-04 requires approval above USD 500.
`) + "\n"},
		{populatedInvoiceStalePath, "Nebula Consulting Invoice March 2026", strings.TrimSpace(`---
type: invoice
status: superseded
vendor: nebula-consulting
project: atlas
superseded_by: invoices/nebula-consulting-2026-04.md
---
# Nebula Consulting Invoice March 2026

## Summary
Populated vault stale source marker: the March invoice used an older USD 300 approval threshold and is superseded by the April invoice.
`) + "\n"},
		{populatedLegalPath, "Atlas Data Retention Memo", strings.TrimSpace(`---
type: legal-doc
status: active
project: atlas
---
# Atlas Data Retention Memo

## Summary
Legal memo marker: current Atlas retention has two unresolved current-source claims in the conflict fixture.
`) + "\n"},
		{populatedLegalArchivePath, "Atlas Data Retention Archive", strings.TrimSpace(`---
type: legal-doc
status: archived
project: atlas
---
# Atlas Data Retention Archive

## Summary
Populated vault stale source marker: the archived retention note says Atlas retention was seven days before the current alpha and bravo conflict sources were filed.
`) + "\n"},
		{populatedContractPath, "Acme Master Services Agreement", strings.TrimSpace(`---
type: contract
status: active
counterparty: acme
project: atlas
---
# Acme Master Services Agreement

## Summary
Contract marker: Acme Atlas work requires a privacy addendum before launch.
`) + "\n"},
		{populatedContractDraftPath, "Acme Master Services Agreement Draft", strings.TrimSpace(`---
type: contract
status: draft
counterparty: acme
project: atlas
---
# Acme Master Services Agreement Draft

## Summary
The draft contract omits the privacy addendum requirement and must not override the active Acme master services agreement.
`) + "\n"},
		{populatedAuthorityPath, "Atlas Populated Vault Authority", strings.TrimSpace(`---
type: source
status: active
project: atlas
populated_role: authority
---
# Atlas Populated Vault Authority

## Summary
Populated vault authority marker: Atlas uses Nebula Consulting invoice approval above USD 500, Nebula Office Supply receipt total USD 118.42, and Acme requires a privacy addendum before launch.

## Evidence
This source reconciles transcripts, articles, meeting notes, reference docs, blog drafts, receipts, invoices, legal docs, and contracts.
`) + "\n"},
		{populatedAuthorityCandidatePath, "Atlas Populated Vault Authority Candidate", strings.TrimSpace(`---
type: source
status: candidate
project: atlas
populated_role: candidate
---
# Atlas Populated Vault Authority Candidate

## Summary
Populated vault duplicate candidate marker: this source looks like an authority summary but is only a candidate and must not beat populated_role authority metadata.

## Evidence
It repeats Atlas, Nebula Consulting, Nebula Office Supply, and Acme terms so retrieval has a duplicate-looking source pressure point.
`) + "\n"},
		{populatedPollutedPath, "Atlas Polluted Note", strings.TrimSpace(`---
type: source
status: polluted
project: atlas
populated_role: decoy
---
# Atlas Polluted Note

## Summary
Polluted decoy marker: this note incorrectly says Atlas may ignore the privacy addendum and approve every invoice without review.
`) + "\n"},
		{populatedConflictAlphaPath, "Atlas Retention Alpha Source", strings.TrimSpace(`---
type: source
status: active
project: atlas
conflict_case: retention
---
# Atlas Retention Alpha Source

## Summary
Populated vault retention conflict current source alpha says Atlas retention is fourteen days.
`) + "\n"},
		{populatedConflictBravoPath, "Atlas Retention Bravo Source", strings.TrimSpace(`---
type: source
status: active
project: atlas
conflict_case: retention
---
# Atlas Retention Bravo Source

## Summary
Populated vault retention conflict current source bravo says Atlas retention is thirty days.
`) + "\n"},
		{populatedSynthesisOldPath, "Atlas Synthesis Old Source", strings.TrimSpace(`---
status: superseded
superseded_by: sources/populated/synthesis-current.md
---
# Atlas Synthesis Old Source

## Summary
Populated vault stale source marker: older populated vault synthesis guidance said Atlas could create a duplicate synthesis page when stale claims appear.
`) + "\n"},
		{populatedSynthesisCurrentPath, "Atlas Synthesis Current Source", strings.TrimSpace(`---
supersedes: sources/populated/synthesis-old.md
---
# Atlas Synthesis Current Source

## Summary
Initial current populated vault synthesis guidance says agents must update the existing synthesis page.
`) + "\n"},
		{populatedSynthesisPath, "Populated Vault Summary", populatedSynthesisSeedBody()},
		{populatedSynthesisDecoyPath, "Populated Vault Summary Decoy", populatedSynthesisDecoySeedBody()},
	}
	for _, doc := range docs {
		if err := createSeedDocument(ctx, cfg, doc.path, doc.title, doc.body); err != nil {
			return err
		}
	}
	return replaceScenarioSeedSection(ctx, cfg, populatedSynthesisCurrentPath, "Summary", "Current populated vault synthesis guidance says agents must update the existing synthesis page, preserve single-line source_refs, inspect freshness and provenance, and avoid duplicate synthesis pages. "+populatedSynthesisOldPath+" is superseded.")
}
func populatedSynthesisSeedBody() string {
	return strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/populated/synthesis-current.md, sources/populated/synthesis-old.md
---
# Populated Vault Summary

## Summary
Stale populated vault claim: create a duplicate synthesis page when Atlas source claims change.

## Sources
- sources/populated/synthesis-current.md
- sources/populated/synthesis-old.md

## Freshness
Checked before the latest populated synthesis source update.
`) + "\n"
}
func populatedSynthesisDecoySeedBody() string {
	return strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/populated/synthesis-old.md
---
# Populated Vault Summary Decoy

## Summary
This duplicate-looking decoy is not the synthesis target for Atlas repairs.

## Sources
- sources/populated/synthesis-old.md

## Freshness
Checked decoy source only.
`) + "\n"
}
