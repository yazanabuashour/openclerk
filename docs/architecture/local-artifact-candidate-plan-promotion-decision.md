---
decision_id: decision-local-artifact-candidate-plan-promotion
decision_title: Local Artifact Candidate Plan Promotion Decision
decision_status: accepted
decision_scope: artifact-candidate-plan-local-file
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/results/ockp-artifact-local-path-candidate-plan-promotion.md, docs/architecture/parser-ocr-artifact-ingestion-candidate-comparison-decision.md, docs/architecture/generalized-artifact-ingestion-promotion-decision.md
---
# Decision: Local Artifact Candidate Plan Promotion

## Status

Accepted for `oc-omg3`: extend read-only `artifact_candidate_plan` with
explicit local artifact planning for UTF-8 text, markdown, and text-bearing
PDF files.

## Decision

Promote the plan-before-write shape:

- input: `artifact.local_path` supplied explicitly by the user
- parser scope: UTF-8 text, markdown, and text-bearing PDF
- output: extracted candidate text/provenance/confidence/duplicate evidence
- durable writes: still require approval through existing `create_document` or
  `ingest_source_url` flows

Reject image OCR, scanned-PDF OCR, opaque binary parsing, home-relative path
shortcuts, directories, and oversized local artifacts in this implementation.

## Safety, Capability, UX

Safety pass: pass. The runner reads only the explicitly supplied file, returns
candidate evidence, commits no artifact content in eval reports, and does not
write a document until the user approves an existing write action.

Capability pass: pass for text and text-bearing PDF candidate planning.
Parser/OCR extraction remains a valid but separate need because this pass does
not prove image text confidence, scanned-PDF recovery, review ergonomics, or
correction workflows.

UX quality: pass. The user gets a simpler OpenClerk surface for supplied local
files without bypassing approval-before-write.

## Compatibility

This adds fields to the read-only candidate plan response but does not change
approved durable write schemas. It does not add OCR, a local artifact registry,
generalized `ingest_artifact`, background file scanning, or opaque parsing.

## Follow-Up

Search performed before closing `oc-omg3`:

- `bd search "OCR scanned PDF artifact candidate plan" --status all`: no
  existing issue found.

Created follow-up:

- `oc-d1pm`: evaluate local OCR artifact candidate planning candidates.
