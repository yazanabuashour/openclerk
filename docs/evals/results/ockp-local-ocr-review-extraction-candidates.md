# OpenClerk Local OCR Review Extraction Candidates

## Summary

`oc-s3wg` compared model-assisted OCR review, local OCR engine options,
scanned-PDF fallback, domain-specific receipt/invoice OCR, and current
unsupported behavior for `artifact_candidate_plan`.

The result is durable non-promotion. Runtime OCR remains unsupported. No
product implementation work item is filed.

## Targeted Evidence

| Row | Scenario | Status | Safety pass | Capability pass | UX quality | Evidence posture |
| --- | --- | --- | --- | --- | --- | --- |
| current-schema-control | Draft `text_extraction: "ocr_review"` request | rejected | pass | fail for OCR | acceptable boundary | Installed runner rejects the future field as unknown; no accidental OCR surface exists. |
| image-unsupported-control | Explicit local PNG artifact | rejected | pass | fail for OCR | acceptable boundary | Installed runner rejects image/OCR parsing and writes nothing. |
| dependency-preflight | Tesseract, OCRmyPDF, Poppler availability | partial | partial | setup burden | Tesseract and OCRmyPDF were unavailable; Poppler helper was present but is not OCR. |
| model-assisted-candidate | Multimodal model OCR review | not promoted | fail for production today | partial | best simplicity fit | Needs explicit egress/provider policy, model identity, confidence, correction, duplicate, and review evidence. |
| local-ocr-candidate | Tesseract or OCRmyPDF/Tesseract/Poppler | not promoted | partial | partial | heavy setup | Needs runtime/version/language data, page rendering, timeout, missing-dependency, fixture, and duplicate evidence. |
| scanned-pdf-fallback | OCR after PDF text extraction has no text | not promoted | partial | partial | strong expectation | Depends on unresolved OCR review and local dependency policy. |
| domain-ocr-fields | Receipt/invoice-specific OCR fields | not promoted | fail | partial | premature | Field extraction should not precede generic reviewed OCR text. |
| current-boundary | Reject OCR and use supplied text/current text-PDF planning | selected | pass | partial | acceptable durable boundary | Preserves runner-only access, local-first behavior, unsupported-file rejection, duplicate handling for supplied text, and approval-before-write. |

## Candidate Comparison

| Candidate | Decision | Reason |
| --- | --- | --- |
| Model-assisted OCR review | Keep as reference pressure only. | Simplest user experience, but remote or harness model use is not local-first without a promoted provider/egress policy. |
| Local Tesseract OCR | Do not promote. | Local-first candidate is unavailable in the current environment and lacks dependency/version/language-pack policy. |
| OCRmyPDF/Tesseract/Poppler scanned-PDF path | Do not promote. | Most complete local scanned-PDF stack, but multiple external dependencies and page provenance remain unproven. |
| Poppler-only fallback | Kill as OCR candidate. | Useful PDF helper, but not OCR. |
| Go-native OCR binding | Do not promote. | Does not remove native dependency and platform-risk questions. |
| Domain-specific receipt/invoice OCR | Keep future-only outside this path. | Requires generic review, correction, and confidence behavior first. |
| Current unsupported boundary | Select. | Safest final outcome for this path. |

## Safety, Capability, UX

Safety pass: pass for selected durable non-promotion. The path preserves
runner-only access, local-first behavior, no hidden OCR/model/parser truth, no
direct vault or SQLite access, unsupported-file rejection, duplicate handling
for supplied or extracted text currently supported by the runner, and
approval-before-write.

Capability pass: partial. OpenClerk still does not recover text from scanned
receipts or scanned PDFs. Current behavior covers explicit supplied content,
UTF-8 text, markdown, and text-bearing PDF local artifacts.

UX quality: acceptable as a durable boundary for this path, not as a feature.
Model-assisted OCR remains the best taste reference, but it cannot become
product behavior without a broader runner-owned provider/egress policy.

## Decision

Close the local OCR/scanned-PDF artifact-candidate path with durable
non-promotion.

No implementation work item is filed. Broader prerequisite policy work is tracked
in `oc-hbu5`. Future OCR should start as a new product decision track only
after one of these prerequisites exists:

- accepted runner-owned model/provider egress policy for local artifacts
- accepted local OCR runtime/dependency policy with fixtures
- accepted platform media/acquisition policy that covers local artifact
  extraction provenance, review, and approval boundaries
