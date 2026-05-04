---
decision_id: decision-skill-reduction-runner-heuristics
decision_status: accepted
decision_scope: skill-reduction-runner-heuristics
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/skill-reduction-runner-heuristics-adr.md, docs/evals/skill-reduction-runner-heuristics-poc.md, docs/evals/results/ockp-skill-reduction-runner-heuristics.md, docs/architecture/thin-skill-workflow-surface-comparison-decision.md
---

# Skill Reduction Into Runner Heuristics Promotion Decision

## Decision

Accept `workflow_guide_report` as the promoted read-only surface for
`oc-uj2y.5`.

The report moves routine workflow surface selection into runner JSON while
keeping `skills/openclerk/SKILL.md` as a compact activation, routing, and
safety contract.

## Safety Pass

Pass. The report is read-only and forbids storage inspection, document
inspection, URL fetches, candidate creation, durable writes, source repair,
selected-action execution, direct SQLite, direct vault inspection, HTTP/MCP
bypasses, source-built runners, unsupported transports, and hidden authority
ranking.

## Capability Pass

Pass. The report returns:

- `recommended_surface`
- `runner_domain`
- `request_shape`
- `use_when`
- `do_not_use_for`
- `candidate_surfaces`
- `validation_boundaries`
- `authority_limits`
- `agent_handoff`

## UX Quality

Pass. The skill line count drops from 250 to 216 lines, and the tested budget
tightens to 225 lines. Routine surface selection no longer requires adding
more durable skill prose.

## Conditional Implementation

Implemented in this epic:

- runner JSON action `workflow_guide_report`
- request object `workflow_guide`
- response object `workflow_guide`
- heuristic surface selection for promoted workflow actions and primitives
- `agent_handoff`
- CLI help text
- README promoted-action guidance
- compact skill action index entry
- stricter skill budget test
- unit tests for routing and validation
- runtime-config locking for fresh-start read path resolution without taking
  the runner write lock

No storage, projection, schema, migration, durable write path, fetch behavior,
or autonomous action execution is added.

## Iteration Gate

Future workflow guidance should compare these candidates before skill growth:

- keep current primitives and shrink the skill
- extend an existing natural runner action
- add a narrow workflow action with `agent_handoff`
- update `workflow_guide_report` only for routing and safety language

If none of those shapes can preserve no-tools boundaries, runner-only access,
citations, provenance, freshness, duplicate handling, and
approval-before-write, record `none viable yet`.
