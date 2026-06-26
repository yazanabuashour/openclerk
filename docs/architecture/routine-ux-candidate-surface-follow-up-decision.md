---
decision_id: decision-routine-ux-candidate-surface-follow-up
decision_title: Routine UX Candidate Surface Follow-Up
decision_status: evidence_only
decision_scope: real-vault-routine-ux-follow-up
decision_owner: platform
decision_date: 2026-05-10
source_refs: docs/evals/results/ockp-real-vault-routine-ux.md
---
# Decision: Routine UX Candidate Surface Follow-Up

## Status

Evidence only. The sanitized real-vault routine UX telemetry report remains
private-boundary evidence and does not promote a public runner action, schema,
storage migration, skill behavior, retrieval backend, or release gate.

The report showed no safety failures, but it did show capability and UX debt:
representative source discovery and decision-like record lookup failed
verification, while validation synthesis create/update completed only after too
much primitive choreography and answer repair. A normal user would expect these
to be natural runner-level workflows rather than brittle prompt sequences.

## Taste Review

Read, fetch, and inspect permission is distinct from durable-write approval.
The candidate surfaces below may inspect runner-visible evidence, return
sanitized summaries, and operate on disposable validation copies where needed.
They must not write to a live private vault, expose private prompts or content,
or treat telemetry success as release evidence.

Record safety, capability, and UX quality separately for every candidate. A
safe primitive sequence that still needs high step count, exact prompt
choreography, or repair turns is capability evidence, not UX acceptance.

## Follow-Up Work

| Need | Candidate surfaces to compare | Required evidence | Follow-up |
| --- | --- | --- | --- |
| Representative source discovery | A: keep search/list/get primitives with clearer handoff. B: extend the natural retrieval search action with a discovery/report mode that returns sanitized source-category summaries. C: add a narrow read-only `source_discovery_report` workflow action. | Safety: no raw private prompts, snippets, paths, ids, or direct vault inspection. Capability: finds representative runner-visible sources and explains category coverage. UX: natural prompt succeeds without exact action choreography. | `oc-h9u1` |
| Decision-like record lookup | A: keep `decisions_lookup`, `records_lookup`, provenance, projection, and evidence primitives. B: extend `evidence_bundle_report` to gracefully include decision-like records across formal decisions, promoted records, provenance, and source evidence. C: add a narrow read-only decision-context lookup action. | Safety: read-only, citations/provenance preserved, no hidden authority ranking. Capability: handles formal decisions and decision-like records that live outside one current surface. UX: natural lookup succeeds without lookup/provenance/projection choreography. | `oc-idol` |
| Validation synthesis create/update | A: keep `compile_synthesis` plus existing primitives and improve handoff text. B: extend `compile_synthesis` with a validation mode for disposable-copy workflows that wraps source refs, create/replace, duplicate checks, provenance, and freshness. C: add a narrow validation-synthesis action that is explicit about disposable targets and approval boundaries. | Safety: live private vault is not the mutation target; durable-write approval stays explicit. Capability: creates or updates the intended validation synthesis, preserves source refs, provenance, freshness, and duplicate handling. UX: natural validation request completes with low ceremony and no repair loop. | `oc-tg24` |

## Decision

Do not promote behavior from the current telemetry lane. Keep
`docs/evals/results/ockp-real-vault-routine-ux.md` as evidence only and use the
follow-up work to compare candidate surfaces with separate safety,
capability, and UX-quality pass criteria.

The evaluated shape failed while the underlying needs remain valid. The next
work should select, combine, defer, kill, or record `none viable yet` for each
need after candidate evidence exists.

## Compatibility

- Committed artifacts stay sanitized and repo-relative.
- Private manifests, raw logs, raw JSON, disposable vault copies, private
  prompts, private paths, titles, snippets, citations, document ids, and chunk
  ids remain local-only.
- Existing `openclerk document` and `openclerk retrieval` primitives remain the
  supported public surfaces until a later decision promotes something narrower.
- Skills should not be expanded with long routine recipes to compensate for the
  telemetry failures.
