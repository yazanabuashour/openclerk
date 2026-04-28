# Broad Contradiction/Audit Engine Revisit Comparison POC

## Status

Implemented POC framing for `oc-6pc`. This document compares current
`openclerk document` and `openclerk retrieval` source-sensitive audit
workflows with a possible promoted broad contradiction/audit surface. It does
not add runner actions, schemas, migrations, storage behavior, public API
behavior, or shipped skill behavior.

The governing ADR is
[`../architecture/broad-contradiction-audit-engine-revisit-adr.md`](../architecture/broad-contradiction-audit-engine-revisit-adr.md).
The targeted reduced report is
[`results/ockp-broad-contradiction-audit-revisit-pressure.md`](results/ockp-broad-contradiction-audit-revisit-pressure.md).

## Candidate Workflows

| Workflow | Existing primitives | Candidate promoted surface | Notes |
| --- | --- | --- | --- |
| Repair stale source-linked audit synthesis | `search`, `list_documents`, `get_document`, `projection_states`, `provenance_events`, `replace_section` or `append_document` | `audit_contradictions` with freshness-aware repair guidance | Candidate must expose the source authority and freshness evidence rather than hide it behind a write result. |
| Explain unresolved current-source conflict | `search`, source `provenance_events`, final answer explanation | `audit_contradictions` returning cited unresolved conflict rows | Candidate must not choose a winner when both sources are current and no supersession/source authority exists. |
| Decide promotion | Natural-intent and scripted-control eval rows | No public surface unless final decision promotes | Candidate must prove capability or ergonomics gap through repeated targeted evidence. |

## Ergonomics Scorecard

| Workflow | Candidate promoted surface | Tool or command count | Assistant calls | Wall time | Prompt specificity required | Failure classification | Authority/provenance/freshness risk |
| --- | --- | ---: | ---: | --- | --- | --- | --- |
| Prior source-sensitive audit repair, `source-sensitive-audit-repair` | None; current document/retrieval workflow | 38 | 10 | 57.64s | Scripted-control | `none` in prior selected pressure | Low when source refs, provenance, freshness, and duplicate prevention stay visible. |
| Prior unresolved conflict explanation, `source-sensitive-conflict-explain` | None; current retrieval workflow | 12 | 5 | 31.99s | Scripted-control | `none` in prior selected pressure | Low when both source paths remain cited and the conflict stays unresolved. |
| New natural-intent revisit | Possible `audit_contradictions` if repeated natural UX fails | 42 | 8 | 112.06s | Natural user intent | `ergonomics_gap` | Medium if a promoted surface hides source authority, provenance, freshness, or unresolved-conflict status. |
| New scripted-control revisit | None; exact current primitive workflow | 224 | 12 | 420.01s | Scripted-control | `capability_gap` | Low if current primitives preserve source paths, provenance, projection freshness, and no-bypass boundaries. |

Prior measurements come from
[`results/ockp-source-sensitive-audit-poc.md`](results/ockp-source-sensitive-audit-poc.md).
New measurements are recorded in
[`results/ockp-broad-contradiction-audit-revisit-pressure.md`](results/ockp-broad-contradiction-audit-revisit-pressure.md).

## Technical Expressibility

Current primitives can express the known broad-audit ingredients when the agent
uses the documented workflow:

- search canonical source evidence
- list existing synthesis candidates
- retrieve the target synthesis before editing
- inspect projection freshness and provenance
- update the existing synthesis instead of creating a duplicate
- inspect provenance for conflicting current sources
- cite both source paths and leave conflicts unresolved when no source
  authority chooses a winner

This means promotion should not follow from high command count alone. A
`capability_gap` requires scripted-control failure: current primitives must be
unable to express the workflow safely even with exact instructions.

## UX Acceptability

The open question is whether the current workflow is acceptable under natural
routine intent. Natural prompts should not have to prescribe every request
shape, but they may name the evidence the answer must preserve: source paths,
citations or source refs, provenance, projection freshness, stale repair, and
unresolved-conflict behavior.

An `ergonomics_gap` requires repeated natural-intent evidence showing the
workflow is too brittle, too slow, too many steps, too retry-prone, or too
dependent on audit-specific prompt choreography. The scripted control must
pass to prove the pressure is UX/reliability cost rather than structural
insufficiency.

## Compatibility Expectations

Any future promoted surface must:

- keep canonical markdown and promoted records as source authority
- return source paths, citations, source refs, or stable source identifiers for
  every source-sensitive claim
- expose provenance and projection freshness
- avoid direct SQLite, direct vault inspection, broad repo search,
  source-built runner paths, HTTP/MCP bypasses, backend variants,
  module-cache inspection, and unsupported transports
- preserve final-answer-only invalid-request behavior
- preserve existing `openclerk document` and `openclerk retrieval` workflows
- keep current-source conflicts unresolved when no supersession metadata or
  other runner-visible source authority chooses a winner

## POC Conclusion

The refreshed targeted pressure lane did not pass. The natural-intent row
failed as `ergonomics_gap`: the agent did not complete the safe current
workflow, missed required repair content, and did not inspect provenance for
both conflict sources. The scripted-control row failed as `capability_gap`: the
agent exhausted the current-primitives workflow, timed out, left the stale
legacy audit claim unrepaired, did not refresh the synthesis projection, and
did not produce the required conflict and promotion decision answer.

The validation controls passed final-answer-only for missing document path,
negative limit, unsupported lower-level workflow, and unsupported transport.
No bypass risk was observed in the selected rows.

This POC supports a promotion decision for a narrow broad contradiction/audit
surface design follow-up. The evidence does not authorize a broad semantic
truth engine; any promoted design must expose source authority, citations or
source paths, provenance, projection freshness, duplicate behavior, unresolved
current-source conflicts, and final-answer-only validation.
