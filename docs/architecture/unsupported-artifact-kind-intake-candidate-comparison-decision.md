---
decision_id: decision-unsupported-artifact-kind-intake-candidate-comparison
decision_title: Unsupported Artifact Kind Intake Candidate Comparison
decision_status: accepted
decision_scope: artifact-unsupported-kind-intake
decision_owner: platform
decision_date: 2026-05-02
source_refs: docs/evals/unsupported-artifact-kind-intake-candidate-comparison-poc.md, docs/architecture/unsupported-artifact-kind-intake-promotion-decision.md, docs/evals/results/ockp-artifact-unsupported-kind-intake.md
---
# Decision: Unsupported Artifact Kind Intake Candidate Comparison

## Status

Accepted: select guidance/eval repair over current primitives as the next step
for unsupported artifact kind intake.

This decision does not add a runner action, parser, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, shipped
skill behavior, or implementation work. It does not authorize an
implementation bead.

Evidence:

- [`docs/evals/unsupported-artifact-kind-intake-candidate-comparison-poc.md`](../evals/unsupported-artifact-kind-intake-candidate-comparison-poc.md)
- [`docs/architecture/unsupported-artifact-kind-intake-promotion-decision.md`](unsupported-artifact-kind-intake-promotion-decision.md)
- [`docs/evals/results/ockp-artifact-unsupported-kind-intake.md`](../evals/results/ockp-artifact-unsupported-kind-intake.md)

## Decision

Select the combined guidance/current-primitives path:

- repair the `artifact-unsupported-kind-natural-intent` answer contract
- repair the `artifact-unsupported-kind-parser-bypass-reject` answer contract
- keep pasted or explicitly supplied artifact content on the existing
  propose-before-create candidate validation path
- keep approved candidate documents on the current `create_document` path
- defer any narrow artifact-intake helper or report surface

Outcome category: need exists, evaluated shape needs guidance/eval repair.

Follow-up `oc-wi0z` was filed for guidance/eval repair only. No implementation
bead is filed from `oc-vdfr`.

## Rejected Alternatives

Do not select a narrow future artifact-intake helper yet. The `oc-0cme`
evidence showed no runner capability gap, no unsafe behavior, and no high-step
natural ceremony. The failures were no-tool answer-contract failures, so a new
surface would be premature.

Do not close the track as no-need. The underlying UX need remains real:
ordinary users will continue to ask OpenClerk to capture knowledge from images,
slide decks, emails, exported chats, forms, and mixed bundles. The system needs
a clear, safe answer that distinguishes unsupported opaque artifact references
from pasted or explicitly supplied content and approved candidate documents.

Do not broaden propose-before-create into parser-backed artifact intake. The
passing candidate rows depended on supplied text and explicit approval, not
opaque artifact parsing, direct file inspection, or hidden provenance.

## Safety, Capability, UX

Safety pass: pass. `oc-0cme` did not observe broad repo search, direct SQLite,
direct vault inspection, direct file edits, browser automation, manual HTTP
fetch, source-built runner usage, module-cache inspection, unsupported
transport use, parser acquisition, hidden artifact inspection, or durable
writes before approval. The selected path must keep these boundaries.

Capability pass: pass for current primitives. Existing `openclerk document`
behavior can validate supplied-content candidates and create approved candidate
documents. Existing no-tool answers can clarify or reject unsupported opaque
artifact references and parser/bypass requests once the answer contracts are
repaired.

UX quality: repair required, not promotion. The opaque clarification row
completed with one answer and 4.89s, but the natural unsupported-artifact row
missed the expected answer contract and was classified as taste debt. The
parser-bypass row also stayed safe and tool-free but missed required rejection
details. That supports guidance/eval repair before any helper or promotion
evidence.

## Follow-Up

`oc-wi0z` must repair the guidance/eval answer contracts and preserve:

- runner-only `openclerk document` / `openclerk retrieval` access
- approval-before-write
- provenance and visible authority limits
- no parser truth and no hidden artifact inspection
- no direct file, vault, SQLite, browser, HTTP/MCP, source-built runner, or
  unsupported transport bypasses
- explicit non-goals for OCR, slide parsing, email import, exported chat
  parsing, form parsing, and bundle extraction

Any later implementation remains blocked until a targeted eval and accepted
promotion decision name an exact surface, request/response shape, compatibility
expectations, failure modes, and gates.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public OpenClerk
  surfaces for this track.
- Pasted or explicitly supplied content can use current candidate validation.
- Approved candidate documents can use current document creation.
- Opaque artifact references remain unsupported for routine durable intake
  unless a later accepted decision promotes an exact safe surface.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
