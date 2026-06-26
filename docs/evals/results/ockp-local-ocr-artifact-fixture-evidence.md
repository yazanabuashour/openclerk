# OpenClerk Local OCR Artifact Fixture Evidence

## Summary

`oc-osmc` gathered synthetic fixture evidence for local OCR and scanned-PDF
artifact candidate planning after `oc-d1pm`. The evidence does not promote
runtime OCR. It confirms that the strongest future shape is still an explicit
review-mode extension of `artifact_candidate_plan`, while the current evidence
remains insufficient for implementation.

This report is evidence/design only. It does not add runner actions, OCR
dependencies, parser pipelines, storage changes, product behavior, public APIs,
or shipped skill behavior.

## Fixture Set

| Fixture | Input placeholder | Expected pressure | Current outcome |
| --- | --- | --- | --- |
| Scanned receipt image | `<explicit-user-local-file>` | User expects a receipt image to become reviewable candidate text. | Unsupported; no safe OCR confidence or correction contract exists. |
| Scanned PDF receipt | `<explicit-user-local-file>` | User expects scanned-PDF fallback after normal PDF text extraction has no text. | Unsupported; text-bearing PDF is supported, OCR fallback is not. |
| Low-confidence OCR page | `<explicit-user-local-file>` | OCR text should require correction before any next-create handoff. | Contract not proven; no response field captures disputed text spans or correction state. |
| Duplicate OCR receipt | `<explicit-user-local-file>` | Duplicate search should suppress next-create until update-versus-new is approved. | Duplicate behavior is proven for text candidates, not OCR-derived candidates. |
| Opaque/private artifact | `<explicit-user-local-file>` | Unsupported or private acquisition should reject without bypasses. | Current unsupported-file boundary remains safest. |

## Candidate Comparison

| Candidate | Safety pass | Capability pass | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| `artifact_candidate_plan` with `text_extraction: "ocr_review"` | Partial. The read-only/no-write boundary fits existing behavior, but extractor identity, local dependency policy, page/image refs, confidence calibration, and low-confidence correction are not specified enough. | Partial. It could cover image and scanned-PDF candidate text, but no fixture proves reliable extraction or duplicate suppression from OCR output. | Best future fit. A normal user would expect OCR review to live beside existing local artifact planning. | Defer. |
| Scanned-PDF OCR fallback after text extraction fails | Partial. It narrows the surface to PDFs already adjacent to the promoted text-bearing PDF path, but page-level provenance and OCR fallback rules are missing. | Partial. It addresses the clearest capability gap, but does not cover images and does not prove confidence or correction behavior. | Strong, because scanned PDFs are a natural extension of current PDF planning. | Defer. |
| Domain-specific receipt/invoice OCR candidate planning | Fail for promotion. Field extraction can overstate parser authority before a generic OCR review envelope is safe. | Partial. It could help high-value artifact workflows later, but only after OCR review, confidence, and correction are proven. | Potentially good later, but too narrow and authority-heavy now. | Defer. |
| `none viable yet` | Pass. Preserves current no-OCR, no-hidden-parser, no-write-before-approval behavior. | Partial. Keeps safe rejection but does not solve OCR/scanned-PDF candidate recovery. | Acceptable as a temporary decision; valid user pressure remains. | Select. |

## Required Contract Before Promotion

A future promotion decision must prove all of these through targeted evals or
runner tests before any product implementation work item is filed:

- explicit opt-in request field, such as `text_extraction: "ocr_review"`
- local-first OCR dependency policy and visible extractor identity
- supported file types, size/page limits, and unsupported-file errors
- page/image refs for extracted text provenance
- confidence and uncertainty fields that do not imply canonical truth
- correction workflow for low-confidence or disputed OCR text
- duplicate search that blocks next-create on likely duplicates
- `planned_no_fetch` and `planned_no_write` for all candidate planning rows
- approved durable write handoff only through existing runner actions

## Decision

Select `none viable yet` for implementation promotion. Do not create a product
implementation work item from `oc-osmc`.

The remaining need is real, but this fixture evidence shows the next useful
work would be another promotion-quality eval only after there is a concrete
OCR dependency and correction contract to test. Filing another evidence work item
without a narrower contract would duplicate `oc-osmc`; filing implementation
work would bypass `oc-d1pm`.

## Path Status

Created deferred follow-up:

- `oc-i8yk`: define local OCR review contract and dependency policy.

No ready work items remain on this OCR/scanned-PDF artifact-candidate path. The path
should resume at `oc-i8yk` when there is a concrete OCR review contract,
dependency policy, and correction workflow to test against the required gates
above.
