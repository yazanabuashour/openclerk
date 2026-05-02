---
decision_id: decision-local-file-artifact-intake-candidate-comparison
decision_title: Local File Artifact Intake Candidate Comparison
decision_status: accepted
decision_scope: artifact-local-file-intake
decision_owner: platform
decision_date: 2026-05-02
source_refs: docs/evals/local-file-artifact-intake-candidate-comparison-poc.md, docs/evals/artifact-local-file-intake-ladder.md, docs/evals/results/ockp-artifact-local-file-intake-ladder.md, docs/architecture/local-file-artifact-intake-ladder-promotion-decision.md
---
# Decision: Local File Artifact Intake Candidate Comparison

## Status

Accepted: select the combined current-primitives path for local file artifact
intake candidate handling. Defer any future runner-owned local-file source
shape until duplicate/provenance repair and later promotion evidence justify an
exact surface.

This decision resolves overlapping follow-ups `oc-4leh` and `oc-vuyb`. It does
not add a runner action, parser, schema, migration, storage behavior, public
API, public OpenClerk interface, product behavior, shipped skill behavior, or
implementation work. It does not authorize an implementation bead.

Evidence:

- [`docs/evals/local-file-artifact-intake-candidate-comparison-poc.md`](../evals/local-file-artifact-intake-candidate-comparison-poc.md)
- [`docs/evals/artifact-local-file-intake-ladder.md`](../evals/artifact-local-file-intake-ladder.md)
- [`docs/evals/results/ockp-artifact-local-file-intake-ladder.md`](../evals/results/ockp-artifact-local-file-intake-ladder.md)
- [`docs/architecture/local-file-artifact-intake-ladder-promotion-decision.md`](local-file-artifact-intake-ladder-promotion-decision.md)

## Decision

Select the combined current-primitives path:

- keep pasted or explicitly supplied local-file-derived content on the existing
  candidate validation path
- keep approved candidate documents on the current `create_document` path
- allow explicit asset/source metadata only when the user supplies the content
  and explicitly approves durable source plus vault-relative asset placement
- keep future `ingest_local_file` or local-file source-ingestion requests
  unsupported unless a later accepted promotion decision names the exact
  surface
- repair duplicate/provenance guidance or eval coverage before any future
  promotion claim

Outcome category: need exists, evaluated shape needs guidance/eval repair and
candidate-surface deferral.

Follow-up `oc-ipjt` tracks duplicate/provenance answer-contract or eval repair.
No implementation bead is filed from `oc-4leh` or `oc-vuyb`.

## Rejected Alternatives

Do not promote a future runner-owned local-file source shape yet. The
`oc-ijdk` evidence showed safety pass and current-primitives capability pass,
while the remaining failed row was duplicate/provenance answer-contract or eval
coverage. A new runner surface would be premature without repaired evidence and
an exact request/response contract.

Do not treat explicit asset-path source-note policy as the normal user surface.
It completed safely, but 42 tools/commands, 5 assistant calls, and 56.48s is
too ceremonial for routine local file artifact intake.

Do not kill the track. The user need remains real: normal users can reasonably
expect a simpler OpenClerk surface for local file artifact intake than a
high-ceremony explicit asset policy workflow. The evaluated shape failed as a
promotion candidate, but the need remains valid.

## Safety, Capability, UX

Safety pass: pass. The selected path preserves runner-only access, no direct
local file reads, no hidden parser/OCR or artifact inspection, no direct vault
or SQLite access, no browser automation, no HTTP/MCP bypass, no source-built
runner usage, no unsupported transports, and no durable writes before approval.

Capability pass: pass for current primitives. Existing `openclerk document`
and `openclerk retrieval` behavior can validate supplied-content candidates,
create approved candidate documents, record explicitly approved asset/source
metadata, reject future unsupported local-file source shapes, reject bypasses,
and expose duplicate/provenance evidence. The duplicate/provenance row failed
because the answer contract or eval guidance did not require the assistant to
inspect and report the runner-visible evidence correctly.

UX quality: defer and repair. Natural clarification and validation-control
rows completed cheaply. Supplied-content and approved-candidate rows completed
through current primitives. The explicit asset-policy row is taste debt because
it required 42 tools/commands, 5 assistant calls, and 56.48s. The failed
duplicate/provenance row blocks promotion until `oc-ipjt` repairs the guidance
or eval coverage.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the only supported
  public OpenClerk surfaces for this track.
- Pasted or explicitly supplied local-file-derived content can use current
  candidate validation.
- Approved candidate documents can use current document creation.
- Explicit asset/source metadata can be recorded only from supplied content and
  explicit durable approval.
- Opaque local file references, local file reads, parser/OCR acquisition,
  hidden artifact inspection, and future local-file source-ingestion requests
  remain unsupported until a later accepted decision promotes an exact safe
  surface.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
