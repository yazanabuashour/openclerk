# Evidence Bundle Workflow Action Candidate Comparison

Bead: `oc-lrqi`

## Decision

Select Candidate C: promote read-only retrieval action `evidence_bundle_report`
and keep existing records/provenance/decision/projection primitives for
advanced or manual investigation.

## Candidates

| Candidate | Shape | Safety | Capability | UX |
| --- | --- | --- | --- | --- |
| A | Keep current primitives with smaller skill. | Pass when agents preserve citations, exact records/decisions, provenance, projection freshness, and read-only behavior. | Pass. Existing primitives can assemble the evidence. | Fails routine UX because agents repeatedly compose lookup plus provenance plus projection calls. |
| B | Promote one narrow evidence-bundle report action. | Pass if the action is read-only and exposes citations, provenance, freshness, validation boundaries, and authority limits. | Pass. It composes existing read paths without new storage. | Better UX for routine bundles, but manual analysis still needs primitives. |
| C | Narrow report action plus existing primitives. | Pass. Preserves read-only behavior and canonical markdown authority. | Pass. No schema migration, vector DB, memory transport, or hidden ranking. | Selected. Natural prompts can request an evidence bundle without exact workflow choreography. |

## Selected Surface

`evidence_bundle_report` packages records/decisions lookup, exact
record/decision evidence when IDs are supplied, provenance events, projection
freshness, citations/source refs, validation boundaries, and authority limits.
It is read-only and adds no new storage.

## Evidence Requirements

The targeted lane is `ockp-evidence-bundle-workflow-action`. Candidate A/B/C
rows must report tool/command count, assistant turns, prompt specificity,
failure/retry rate, `safety_pass`, `capability_pass`, and `ux_quality`.
Scripted primitive success proves capability only; natural workflow-action
success is required for UX.
