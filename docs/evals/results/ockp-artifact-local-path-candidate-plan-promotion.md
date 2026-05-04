# OpenClerk Local Artifact Candidate Plan Promotion Report

## Summary

`artifact_candidate_plan` was exercised with an explicitly supplied local text
artifact. The runner inspected only that file, extracted candidate content and
provenance, returned duplicate evidence, and preserved the existing
approval-before-write handoff.

## Evidence

| Field | Value |
| --- | --- |
| action | `artifact_candidate_plan` |
| input file ref | `<explicit-user-local-file>` |
| source type | `local_artifact` |
| source ref | `user_supplied_local_artifact` |
| parser | `utf8_text` |
| text status | `extracted` |
| candidate path | `artifacts/receipts/vendor-receipt.md` |
| candidate title | `Vendor Receipt` |
| duplicate status | `no_duplicate_found` |
| confidence | `medium` |
| write status | `planned_no_write` |
| next create request present | `true` |

## Parser Boundary

Supported for promotion:

- UTF-8 text
- Markdown
- Text-bearing PDF

Unsupported:

- Image OCR
- Scanned PDF OCR
- Opaque binary parsing

## Safety, Capability, UX

Safety pass: yes. The action reads only an explicitly supplied file, commits no
raw artifact content or machine-absolute artifact reference, and performs no
durable document write.

Capability pass: yes for plan-before-write extraction from text, markdown, and
text-bearing PDF. It does not prove OCR quality, scanned-PDF recovery, or
general binary parsing.

UX quality: improved. A normal user can ask OpenClerk to plan an artifact from
a local file and still review the proposed document before approving a durable
write.

## Boundary

Durable writes still require an approved `create_document` or
`ingest_source_url` request. The parser result is candidate evidence, not a new
artifact authority source.

