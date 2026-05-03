# Source Audit Workflow Action Candidate Comparison

Bead: `oc-w8x0`

## Decision

Select Candidate C: promote narrow retrieval action `source_audit_report` and
keep existing audit/search/provenance/projection primitives for advanced or
manual cases. Broad contradiction-engine claims remain rejected.

## Candidates

| Candidate | Shape | Safety | Capability | UX |
| --- | --- | --- | --- | --- |
| A | Keep current primitives with smaller skill. | Pass when agents preserve source authority, provenance/freshness, duplicate prevention, and unresolved-current-source handling. | Pass. Existing primitives can explain and repair source-linked audit targets. | Fails routine UX when agents need exact commands and multi-step choreography. |
| B | Promote one narrow source-sensitive audit action. | Pass if repair mode can update only an existing synthesis target and never creates a new synthesis page. | Pass. Existing audit internals can be reused behind a clearer action name. | Better UX, but manual investigations still need primitives. |
| C | Narrow action plus existing primitives. | Pass. Keeps repair boundaries and leaves broad engine claims out. | Pass. No broad contradiction engine, schema migration, or hidden authority ranking. | Selected. Natural prompts can ask for a source-sensitive audit report without a long skill recipe. |

## Selected Surface

`source_audit_report` is framed as source-sensitive audit, not a broad
contradiction engine. Default `mode` is `explain` and read-only.
`repair_existing` may update only an existing synthesis target. Responses expose
source evidence, provenance/freshness, duplicate status, unresolved conflict
groups, validation boundaries, authority limits, and `agent_handoff`. The
installed `openclerk retrieval --help` output exposes the compact action shape
so routine agents do not need a long skill recipe or source inspection.

## Evidence Requirements

The targeted lane is `ockp-source-audit-workflow-action`. Candidate A/B/C rows
must report tool/command count, assistant turns, prompt specificity,
failure/retry rate, `safety_pass`, `capability_pass`, and `ux_quality`.
Scripted exact-command success proves capability only; natural workflow-action
success under the low-ceremony threshold is required for UX. `oc-nj5h`
validates UX maturity by requiring plain source-sensitive audit prompts to
complete from the narrow action and its `agent_handoff`, without broad
contradiction-engine claims or long skill recipes.
