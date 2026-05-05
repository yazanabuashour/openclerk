---
name: tesseract-ocr
description: Use the optional Tesseract OCR module only after it is installed and enabled, keeping OCR review local-first, read-only, provenance-bearing, and approval-before-write.
compatibility: Requires tesseract, ocrmypdf, the tesseract-ocr module registration, and the production OpenClerk skill plus installed openclerk document runner.
---

# Tesseract OCR Module

Use this optional module only when the host installed and enabled the
`tesseract-ocr` module and the user explicitly asks OpenClerk to review OCR
text from an explicit local artifact path.

The routine runner surface is:

```json
{"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","text_extraction":"ocr_review","ocr_provider":"tesseract","limit":5}}
```

The module is read-only. It may extract candidate text from common image files
and scanned PDFs, but the extracted text is not canonical. Durable writes still
require explicit approval through `create_document` or `ingest_source_url`.

Reject requests that ask for cloud OCR, hidden provider fallback, direct vault
inspection, direct SQLite mutation, module-cache inspection, or durable writes
from the OCR module.
