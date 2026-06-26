---
decision_id: decision-local-ocr-review-contract-dependency-policy
decision_title: Local OCR Review Contract And Dependency Policy
decision_status: accepted
decision_scope: artifact-candidate-plan-ocr-review-contract
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/local-ocr-review-contract-dependency-policy.md, docs/evals/results/ockp-local-ocr-review-contract-dependency-policy.md, docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md, docs/architecture/local-ocr-artifact-candidate-planning-decision.md
---
# Decision: Local OCR Review Contract And Dependency Policy

## Status

Accepted as a non-implementation decision for `oc-i8yk`.

Supersession note: this historical non-implementation decision was superseded
on 2026-05-05 by [`ocr-module-final-decision.md`](ocr-module-final-decision.md),
which implements the local OCR engine comparison as the optional
`modules/tesseract-ocr/module.json` module. The contract and dependency gates
below remain the evidence basis for that promoted shape.

Do not add runtime OCR, multimodal model calls, scanned-PDF OCR fallback,
image parsing, domain-specific receipt or invoice OCR actions, parser
pipelines, storage migrations, public APIs, shipped skill behavior, or product
implementation work from this decision.

Evidence:

- [`docs/evals/local-ocr-review-contract-dependency-policy.md`](../evals/local-ocr-review-contract-dependency-policy.md)
- [`docs/evals/results/ockp-local-ocr-review-contract-dependency-policy.md`](../evals/results/ockp-local-ocr-review-contract-dependency-policy.md)
- [`docs/evals/results/ockp-local-ocr-artifact-fixture-evidence.md`](../evals/results/ockp-local-ocr-artifact-fixture-evidence.md)
- [`docs/architecture/local-ocr-artifact-candidate-planning-decision.md`](local-ocr-artifact-candidate-planning-decision.md)

## Decision

Select `artifact_candidate_plan` OCR review as the future contract, with
model-assisted extraction as the simplest first evidence candidate and local
OCR engines as the required local-first comparison.

The future request shape remains:

```json
{"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","text_extraction":"ocr_review","limit":5}}
```

That shape is not promoted here. It may be promoted only after targeted
evidence proves extractor identity, dependency policy, page/image refs,
confidence, uncertainty, correction workflow, duplicate suppression,
unsupported-file behavior, and approval-before-write.

## Candidate Comparison

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Model-assisted OCR review inside `artifact_candidate_plan` | Plausible if runner-owned, explicit, no hidden egress, and no write before review. | Best first candidate to test because the eval harness already records multimodal-capable agent models such as `gpt-5.4-mini`, so configured model OCR may avoid a separate stack. | Best simplicity fit for normal users. | Select for future evidence. |
| Local OCR engine or library | Strongest local-first posture if dependency identity, version, language data, and errors are visible. | Plausible with Tesseract for images and OCRmyPDF/Tesseract/Poppler for scanned PDFs. | Good offline fit, higher setup burden. | Required comparison. |
| Scanned-PDF OCR fallback | Plausible as a narrow extension of current PDF planning. | Strongest current gap after text-bearing PDF extraction. | Strong expectation. | Fold into OCR review evidence. |
| Domain-specific receipt/invoice OCR | Risky before generic review because field extraction can overstate authority. | Useful later. | Premature. | Defer. |
| `none viable yet` | Passes for current production safety. | Does not recover OCR text. | Acceptable only as temporary boundary. | Keep current behavior. |

## Safety, Capability, UX

Safety pass: pass for non-implementation. Current production still rejects
OCR/image parsing, scanned-PDF OCR fallback, opaque parsing, hidden file
inspection, direct vault or SQLite access, unsupported transports, and durable
writes before approval.

Capability pass: partial. The contract and dependency policy are now concrete
enough for targeted evidence, but extraction quality, confidence calibration,
correction behavior, duplicate suppression, and unsupported-file behavior are
not proven.

UX quality: candidate selected for future evidence. A normal user would expect
a simpler surface than rejecting scanned receipts and PDFs forever. The best
taste fit is extending existing `artifact_candidate_plan`, not adding a new
artifact ingestion action.

## Promotion Gates

A later promotion decision must name:

- exact request and response fields for OCR review
- configured extraction mode and visible extractor identity
- model id or local runtime/version, plus egress posture
- page/image refs, text spans, confidence, uncertainty, and correction status
- unsupported-file, missing-dependency, timeout, size, and page-limit behavior
- duplicate checks before any next-create request
- approval-before-write through existing durable actions only
- targeted eval rows for model-assisted OCR, local OCR, scanned PDF,
  low-confidence correction, duplicate risk, and bypass rejection

## Follow-Up

No product implementation work item should be filed from `oc-i8yk`.

The remaining need is valid and now has a selected contract for future
promotion evidence. The next follow-up should evaluate `artifact_candidate_plan`
OCR review with model-assisted extraction against a local OCR dependency
comparison before any implementation is authorized.

Created follow-up:

- `oc-s3wg`: evaluate OCR review extraction candidates for
  `artifact_candidate_plan`.

`oc-s3wg` completed that extraction-candidate comparison in
[`docs/architecture/local-ocr-review-extraction-candidate-decision.md`](local-ocr-review-extraction-candidate-decision.md).
It selected durable non-promotion for this OCR/scanned-PDF path and filed no
product implementation work item.

## Compatibility

Existing behavior remains unchanged:

- `artifact_candidate_plan` remains read-only and supports only explicit
  UTF-8 text, markdown, and text-bearing PDF local artifacts.
- OCR/image parsing, scanned-PDF OCR fallback, and opaque binary parsing remain
  unsupported.
- Durable writes still require approved `create_document` or
  `ingest_source_url`.
- Committed docs and reports use repo-relative paths or neutral placeholders
  such as `<explicit-user-local-file>`.
