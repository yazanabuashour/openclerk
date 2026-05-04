---
decision_id: decision-parser-ocr-artifact-ingestion-candidate-comparison
decision_title: Parser OCR Artifact Ingestion Candidate Comparison
decision_status: accepted
decision_scope: parser-ocr-artifact-ingestion-candidates
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/generalized-artifact-ingestion-promotion-decision.md, docs/architecture/unsupported-artifact-kind-intake-candidate-comparison-decision.md, docs/evals/results/ockp-artifact-local-file-intake-ladder.json, docs/evals/results/ockp-artifact-unsupported-kind-intake-guidance-repair.json
---
# Decision: Parser OCR Artifact Ingestion Candidate Comparison

## Status

Accepted as a non-promotion decision for `oc-w7xa`.

Do not add parser-backed local artifact ingestion, OCR extraction, generalized
`ingest_artifact`, domain-specific artifact actions, local file reads, asset
registry writes, parser pipelines, storage migrations, or new public APIs.

## Decision

Select the current combined shape:

- explicit user-provided content can continue through the existing
  propose-before-create and approved `create_document` flow
- vault-relative asset placement can be recorded only when explicitly supplied
  and approved by the user
- opaque local artifacts, OCR, parser extraction, and generalized artifact
  ingestion remain unsupported until repeated promotion evidence exists

Candidate comparison:

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Explicit supplied content | Pass with approval-before-write and visible authority limits. | Covers current safe artifact capture. | Good enough for routine supplied text. | Keep. |
| Local artifact registry | Not proven without asset policy and provenance lifecycle. | Could help later. | Adds file-management ceremony. | Defer. |
| Parser/OCR candidate extraction | Not proven for confidence, provenance, unsupported files, or corrections. | Valid need remains. | Useful only if extraction is reviewed before write. | Defer. |
| Generalized `ingest_artifact` | Too broad without exact kinds and failure modes. | Not justified by current evidence. | Risky hidden parser truth. | Do not promote. |
| Domain-specific artifact actions | Could be safer later for high-value formats. | Needs repeated targeted gaps first. | Better than generalized parsing if promoted. | Defer. |

## Safety, Capability, UX

Safety pass: pass. The selected outcome preserves installed-runner access,
local-first behavior, visible provenance, duplicate handling,
unsupported-file behavior, asset policy, and approval before durable records
are written.

Capability pass: partial. Current primitives cover explicit supplied content
and approved candidate documents. Parser/OCR extraction remains a real but
unproven capability need.

UX quality: partial. Supplied-content capture is safe but can be high ceremony;
parser/OCR would improve UX only if confidence, provenance, review, and
approval boundaries are explicit.

## Follow-Up

Search performed before closing `oc-w7xa`:

- `bd search "parser OCR local artifact ingestion" --status all`: no existing issue found.

Created follow-up:

- `oc-omg3`: gather parser OCR artifact ingestion promotion evidence.

## Compatibility

Existing behavior remains unchanged. No parser, OCR engine, local file read,
asset registry, runner schema, storage schema, public API, or durable write
path is added.
