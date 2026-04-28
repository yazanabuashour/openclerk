---
decision_id: decision-broad-contradiction-audit-engine-revisit-promotion
decision_title: Broad Contradiction/Audit Engine Revisit Promotion
decision_status: accepted
decision_scope: broad-contradiction-audit
decision_owner: platform
---
# Decision: Broad Contradiction/Audit Engine Revisit Promotion

## Status

Accepted: promote a narrow broad contradiction/audit surface design follow-up.

This decision does not itself add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, or shipped skill behavior.
It authorizes the existing implementation placeholder `oc-nw7` to design and
implement the promoted surface under the gates below.

Evidence:

- [`broad-contradiction-audit-engine-revisit-adr.md`](broad-contradiction-audit-engine-revisit-adr.md)
- [`../evals/broad-contradiction-audit-engine-revisit-comparison-poc.md`](../evals/broad-contradiction-audit-engine-revisit-comparison-poc.md)
- [`../evals/results/ockp-broad-contradiction-audit-revisit-pressure.md`](../evals/results/ockp-broad-contradiction-audit-revisit-pressure.md)
- [`../evals/results/ockp-source-sensitive-audit-poc.md`](../evals/results/ockp-source-sensitive-audit-poc.md)

## Decision

Promote a narrow `audit_contradictions` runner surface design for source-linked
audit repair and unresolved-conflict explanation. Do not promote a broad
semantic contradiction truth engine.

Capability path: promote. The scripted-control row in
`docs/evals/results/ockp-broad-contradiction-audit-revisit-pressure.md` failed
as `capability_gap`. With exact current-primitives instructions, the agent made
224 tool/command calls over 420.01s, hit `context deadline exceeded`, left the
legacy audit claim unrepaired, did not refresh the synthesis projection, and
did not produce the required conflict and promotion decision answer.

Ergonomics path: promote. The natural-intent row failed as `ergonomics_gap`
after 42 tool/command calls, 8 assistant calls, and 112.06s. It did not
complete the safe current workflow, missed required repair content, and did
not inspect provenance for both conflict sources.

Validation controls passed final-answer-only for missing document path,
negative limit, unsupported lower-level workflow, and unsupported transport.
No bypass risk was observed in the selected rows.

Current primitives are therefore not safe enough for this combined workflow:
the scripted control did not complete even with exact instructions. Current UX
is also not acceptable enough to keep without promotion: natural intent failed
before reaching the required repair, provenance, and conflict evidence.

## Promoted Surface Gates

The implementation follow-up must use the action name `audit_contradictions`
unless a mechanical naming conflict is found. It must stay inside installed
OpenClerk runner JSON and expose this minimum request shape:

```json
{
  "action": "audit_contradictions",
  "audit": {
    "query": "source-sensitive audit runner repair evidence",
    "target_path": "synthesis/audit-runner-routing.md",
    "mode": "plan_only|repair_existing",
    "conflict_query": "source sensitive audit conflict runner retention",
    "limit": 10
  }
}
```

The response must expose:

- selected target path and candidate synthesis paths
- source paths, source refs, or citations used for every source-sensitive claim
- current and superseded source classification when runner-visible authority
  exists
- unresolved conflict groups when current sources disagree without
  supersession or source authority
- provenance event references inspected for target and conflicting sources
- projection freshness before and after any repair
- whether a repair was applied, skipped, or rejected
- duplicate-prevention outcome
- failure classification when evidence is insufficient

`mode: plan_only` must never write. `mode: repair_existing` may update only an
existing target document and must not create duplicate synthesis pages. When
current sources conflict and no runner-visible authority chooses a winner, the
action must report the conflict as unresolved rather than forcing a semantic
winner.

## Compatibility And Tests

Existing behavior remains unchanged until `oc-nw7` lands:

- `openclerk document` and `openclerk retrieval` remain valid public workflows.
- Canonical markdown sources and promoted records outrank synthesis and audit
  output.
- Source-sensitive audit output must preserve source refs, citations or source
  paths, provenance, freshness, and no-bypass invariants.
- Missing-field and invalid-request handling must continue to preserve the
  final-answer-only validation contract.

`oc-nw7` must add targeted tests and an eval lane that prove:

- stale source-linked audit synthesis can be repaired without duplicates
- unresolved current-source conflicts remain unresolved
- provenance and projection freshness are inspectable in the response
- broad repo search, direct vault inspection, direct SQLite, source-built
  runner paths, HTTP/MCP bypasses, unsupported transports, backend variants,
  module-cache inspection, and ad hoc lower-level programs remain prohibited
- existing document/retrieval workflows remain backward compatible

No broader contradiction engine, second truth system, semantic contradiction
store, graph layer, memory/router behavior, or hidden authority ranking is
authorized by this decision.
