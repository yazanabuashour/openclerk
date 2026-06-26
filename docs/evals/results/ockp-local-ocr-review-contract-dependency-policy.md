# OpenClerk Local OCR Review Contract And Dependency Policy

## Summary

`oc-i8yk` defined a future OCR review contract and dependency policy for
`artifact_candidate_plan`. The evidence does not promote runtime OCR. It
selects model-assisted OCR review as the simplest next candidate to test,
while requiring a local OCR dependency comparison for strict local-first
operation.

This report is evidence/design only. It does not add runner actions, model
calls, OCR binaries, parser pipelines, storage changes, product behavior,
public APIs, or shipped skill behavior.

## Candidate Comparison

| Candidate | Safety pass | Capability pass | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| `artifact_candidate_plan` OCR review with model-assisted extraction | Partial. The no-write review boundary fits, but hosted model use must expose egress, model identity, confidence, uncertainty, and review status. | Best candidate to test first because the current eval harness already records multimodal-capable agent models such as `gpt-5.4-mini`, so a later pass can test configured model OCR before adding a separate stack. | Best simplicity fit if the runner owns the call and keeps review before write. | Select for future evidence only. |
| `artifact_candidate_plan` OCR review with local OCR engine | Partial. Strong local-first posture, but dependency/version/language-pack/page-rendering policy is still unproven. | Plausible for receipts and scanned PDFs, especially with Tesseract or OCRmyPDF. | Good for offline users, but higher setup burden. | Keep as required comparison. |
| Scanned-PDF OCR fallback | Partial. Natural extension of text-bearing PDF planning, but still requires page refs, renderer policy, confidence, and correction workflow. | Strongest narrow gap after current PDF text extraction. | Strong user expectation. | Fold into OCR review evidence. |
| Domain-specific receipt/invoice OCR | Not ready. Field extraction can overstate authority before generic OCR review is safe. | Useful later after generic review passes. | Potentially good, but premature. | Defer. |
| `none viable yet` | Pass for production today because current unsupported boundaries stay intact. | Partial because it does not recover OCR text. | Temporary only; valid need remains. | Keep current behavior. |

## Contract Gates

A future promotion decision must prove these fields and behaviors before
implementation:

- explicit opt-in request field: `artifact.text_extraction: "ocr_review"`
- visible extraction mode, extractor identity, version or model id, and egress
  posture
- supported file types, size/page limits, timeouts, and deterministic
  unsupported or missing-dependency errors
- page/image refs and text-span provenance for extracted candidate text
- confidence, uncertainty notes, and correction status that do not imply
  canonical truth
- duplicate search from reviewed candidate text, with next-create suppressed on
  likely duplicates or unreviewed low-confidence OCR
- `planned_no_fetch`, `planned_no_write`, approval-before-write, validation
  boundaries, authority limits, and `agent_handoff`

## Dependency Policy

Model-assisted OCR is the simplest first evidence candidate, not the production
default. It is acceptable only as an explicit configured runner mode with
visible model identity and egress posture. The runner must not silently send a
private local artifact to a remote model.

Local OCR remains the strict local-first comparison. Tesseract is the primary
local image OCR candidate; OCRmyPDF is the primary scanned-PDF wrapper
candidate; Poppler is useful for PDF rendering or text helpers but is not OCR
by itself; Go-native bindings are acceptable only if they do not hide native
runtime and language-pack requirements.

## Safety, Capability, UX

Safety pass: pass for the selected non-implementation outcome. Existing
production behavior still rejects OCR/image parsing, scanned-PDF OCR fallback,
hidden file inspection, direct vault or SQLite access, lower-level bypasses,
and durable writes before approval.

Capability pass: partial. The contract identifies a plausible future surface
and dependency policy, but no fixture proves extraction quality, confidence
calibration, correction behavior, duplicate suppression, or unsupported-file
handling for OCR-derived text.

UX quality: candidate selected for future evidence. A normal user would expect
scanned receipts and PDFs to become reviewable candidates through the same
artifact planning action. Model-assisted OCR is the simplest candidate to test,
but UX does not pass for production until review/correction and dependency
failure behavior are proven.

## Decision

Select `artifact_candidate_plan` OCR review with model-assisted extraction as
the next evidence candidate. Keep local OCR engines as a required comparison
for local-first policy. Keep runtime OCR unsupported in current production.

Do not file a product implementation work item from `oc-i8yk`. The next valid work
is targeted promotion evidence for the selected contract, not implementation.

Created follow-up:

- `oc-s3wg`: evaluate OCR review extraction candidates for
  `artifact_candidate_plan`.
