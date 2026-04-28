---
decision_id: adr-broad-contradiction-audit-engine-revisit
decision_title: Broad Contradiction/Audit Engine Revisit
decision_status: accepted
decision_scope: broad-contradiction-audit
decision_owner: platform
---
# ADR: Broad Contradiction/Audit Engine Revisit

## Status

Accepted as a deferred-capability revisit track.

This ADR frames broad contradiction/audit as evidence gathering only. It does
not add runner actions, schemas, storage behavior, migrations, skill behavior,
or public APIs.

## Context

OpenClerk already keeps source-sensitive audit and contradiction-like workflows
as a reference pattern in
[`knowledge-configuration-v1-adr.md`](knowledge-configuration-v1-adr.md) and
[`../evals/results/ockp-source-sensitive-audit-poc.md`](../evals/results/ockp-source-sensitive-audit-poc.md).
That prior evidence showed agents can search canonical sources, inspect
provenance and projection freshness, repair stale source-linked synthesis, and
explain unresolved conflicts when current sources disagree without source
authority.

The revisit asks whether that reference posture still holds under the
deferred-capability promotion rubric in
[`deferred-capability-promotion-gates.md`](deferred-capability-promotion-gates.md).
The track tests both:

- `capability_gap`: whether existing `openclerk document` and
  `openclerk retrieval` actions are structurally insufficient for broad
  contradiction/audit workflows.
- `ergonomics_gap`: whether existing actions can express the workflow but are
  too slow, too scripted, too brittle, too guidance-dependent, or too costly
  for routine AgentOps use.

## Decision

Use targeted ADR, POC, eval, and decision artifacts before any implementation
work. The default outcome is to keep broad contradiction/audit as
reference/deferred unless repeated targeted evidence proves a capability gap or
ergonomics gap.

The current public surface remains:

- `openclerk document`
- `openclerk retrieval`

Canonical markdown and promoted canonical records remain the source authority.
A broad audit workflow may summarize conflicts, stale derived synthesis, and
source supersession, but it must not invent semantic contradiction truth or
choose a winner when current sources conflict without runner-visible authority.

## Options

| Option | Description | Promotion posture |
| --- | --- | --- |
| Keep current primitives | Use search, `list_documents`, `get_document`, `provenance_events`, `projection_states`, and document edits over canonical markdown. | Default/reference if natural and scripted pressure pass with acceptable ergonomics. |
| Strengthen guidance/evals | Keep the runner unchanged while improving skill wording or targeted scenarios. | Use when failures are ordinary guidance, eval coverage, or fixture gaps. |
| Add a narrow audit surface | Add a promoted runner action that packages source search, candidate selection, conflict explanation, provenance, freshness, and safe repair. | Consider only if repeated natural rows show unacceptable UX while scripted controls prove current primitives are technically sufficient. |
| Add a broad semantic contradiction engine | Infer contradictions as a first-class truth layer independent of source authority. | Kill unless it can preserve canonical authority, citations, provenance, freshness, and no-bypass invariants. |

## Invariants

Any future promoted surface must preserve:

- AgentOps-only routine operation through installed runner JSON.
- Canonical markdown and promoted records as source authority.
- Citations, source refs, or stable source identifiers for source-sensitive
  claims.
- Inspectable provenance and projection freshness.
- Local-first operation.
- No broad repo search, direct vault inspection, direct SQLite, source-built
  runner paths, HTTP/MCP bypasses, backend variants, module-cache inspection,
  or ad hoc lower-level transports for routine tasks.
- Final-answer-only handling for invalid no-tools requests.
- Unresolved conflict behavior when current sources disagree and no
  runner-visible authority chooses a winner.

## Non-Goals

This ADR does not:

- define a promoted contradiction or audit runner action
- add semantic contradiction storage
- add a graph, memory, router, or audit database
- add migrations, background jobs, indexes, or parser pipelines
- make synthesis more authoritative than canonical sources or promoted records
- relax citation, source ref, provenance, freshness, duplicate-prevention, or
  validation requirements

## Promotion Gates

Promotion via `capability_gap` requires repeated scripted-control failures that
show current document/retrieval primitives cannot safely express broad
contradiction/audit workflows while preserving authority, citations,
provenance, freshness, duplicate prevention, and unresolved-conflict behavior.

Promotion via `ergonomics_gap` requires repeated natural-intent failures or
unacceptable UX cost where the scripted control still passes. Evidence must
show high step count, latency, prompt brittleness, retries, wrong target
selection, duplicate creation, skipped freshness/provenance checks, or
workflow-specific guidance dependence that a proposed surface would reduce
without weakening the invariants above.

Defer or keep as reference when current primitives pass with acceptable
ergonomics, when failures are data hygiene, ordinary skill guidance, eval
coverage, or partial evidence, or when the proposed surface would mostly be a
more complicated way to do source-linked retrieval and synthesis repair.

Kill the candidate if it makes arbitrary semantic contradiction claims, drops
source paths, hides supersession or freshness evidence, hides provenance,
forces winners for unresolved current-source conflicts, creates a second truth
system, or encourages bypasses.

## Evidence Plan

The POC comparison is
[`../evals/broad-contradiction-audit-engine-revisit-comparison-poc.md`](../evals/broad-contradiction-audit-engine-revisit-comparison-poc.md).
The targeted reduced report is
[`../evals/results/ockp-broad-contradiction-audit-revisit-pressure.md`](../evals/results/ockp-broad-contradiction-audit-revisit-pressure.md).
The final promotion decision is
[`broad-contradiction-audit-engine-revisit-promotion-decision.md`](broad-contradiction-audit-engine-revisit-promotion-decision.md).
