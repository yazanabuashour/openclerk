# Promoted Record Domain Expansion Comparison POC

## Status

Implemented POC framing for `oc-k47`. This document compares current
`openclerk document` and `openclerk retrieval` record workflows with a possible
promoted typed-domain record surface. It does not add runner actions, schemas,
migrations, storage behavior, public API behavior, or shipped skill behavior.

The governing ADR is
[`../architecture/promoted-record-domain-expansion-adr.md`](../architecture/promoted-record-domain-expansion-adr.md).
The targeted reduced report is
[`results/ockp-promoted-record-domain-expansion-pressure.md`](results/ockp-promoted-record-domain-expansion-pressure.md).

## Candidate Workflows

| Workflow | Existing primitives | Candidate promoted surface | Notes |
| --- | --- | --- | --- |
| Find domain records | `search`, `list_documents`, `records_lookup` with `entity_type` | `policy_records_lookup` or another typed domain lookup | Candidate could reduce choreography, but canonical markdown remains record authority. |
| Inspect one record | `get_document`, `record_entity` | Typed detail action returning facts and citations | Candidate must keep citations/source evidence visible. |
| Verify freshness | `provenance_events`, `projection_states` for projection `records` | Candidate response includes provenance and records freshness | Candidate must expose freshness instead of hiding derived record state. |
| Decide promotion | Natural-intent and scripted-control eval rows | No public surface unless final decision promotes | Candidate must prove capability or repeated ergonomics gap through targeted evidence. |

## Ergonomics Scorecard

| Workflow | Candidate promoted surface | Tool or command count | Assistant calls | Wall time | Prompt specificity required | Failure classification | Authority/provenance/freshness risk |
| --- | --- | ---: | ---: | --- | --- | --- | --- |
| Existing production record rows, `records-provenance` and `promoted-record-vs-docs` | None; current record/document workflow | Production-gate baseline | Scenario-specific | Recorded in `docs/evals/results/ockp-agentops-production.md` | Scripted-control | Prior production pressure covers generic records and service-vs-doc comparison. | Low when canonical markdown, citations, provenance, and records freshness remain inspectable. |
| New natural-intent domain expansion row | Possible typed policy/domain surface if repeated natural UX fails | 28 | 8 | 70.68s | Natural user intent | `none` | Low: current workflow preserved canonical record authority, citations, provenance, records freshness, and bypass boundaries. |
| New scripted-control domain expansion row | None; exact current primitive workflow | 16 | 4 | 33.54s | Scripted-control | `skill_guidance_or_eval_coverage` | Low for data path: runner-visible evidence existed, but the assistant answer did not satisfy the comparison contract. |

New measurements come from
[`results/ockp-promoted-record-domain-expansion-pressure.md`](results/ockp-promoted-record-domain-expansion-pressure.md).

## Technical Expressibility

Current primitives can express the seeded policy-like domain workflow at the
runner data level:

- search canonical markdown for the policy marker
- list `records/policies/` documents by path prefix
- retrieve `records/policies/agentops-escalation-policy.md`
- use `records_lookup` with `entity_type: policy`
- use `record_entity` for `agentops-escalation-policy`
- inspect entity provenance
- inspect records projection freshness
- cite canonical markdown and keep record state derived

The scripted row failed the answer contract, not the durable data contract:
runner-visible promoted-record evidence existed and the required runner steps
were observed. This is not enough to prove a structural capability gap.

## UX Acceptability

The natural-intent row completed with `none` failure classification, but it
remains high-latency pressure at 28 commands, 8 assistant calls, and 70.68s.
That is useful benchmark evidence, not repeated ergonomics-gap evidence.

The scripted-control row needs guidance or eval repair because the final answer
did not complete the capability/ergonomics comparison wording even though the
runner evidence was available. A promotion decision should not treat this as a
new public surface requirement.

## Compatibility Expectations

Any future promoted surface must:

- keep canonical markdown as record identity and fact authority
- return citation/source evidence for domain claims
- expose provenance and records projection freshness
- avoid direct SQLite, direct vault inspection, broad repo search,
  source-built runner paths, HTTP/MCP bypasses, backend variants, and
  module-cache inspection
- preserve final-answer-only invalid-request behavior
- remain backward compatible with existing `openclerk document`,
  `openclerk retrieval`, `records_lookup`, and `record_entity` workflows

## POC Conclusion

The targeted pressure lane does not justify promotion. The natural row passed,
the scripted row exposed answer-contract repair work, and validation controls
stayed final-answer-only. Current evidence supports deferring promoted record
domain expansion for guidance/eval repair rather than creating a policy-specific
record implementation follow-up.

The final decision is recorded in
[`../architecture/promoted-record-domain-expansion-promotion-decision.md`](../architecture/promoted-record-domain-expansion-promotion-decision.md).
