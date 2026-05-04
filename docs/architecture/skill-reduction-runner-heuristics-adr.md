---
decision_id: adr-skill-reduction-runner-heuristics
decision_status: accepted
decision_scope: skill-reduction-runner-heuristics
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/agent-knowledge-plane.md, docs/architecture/deferred-capability-promotion-gates.md, docs/architecture/thin-skill-workflow-surface-comparison-decision.md, docs/evals/skill-reduction-runner-heuristics-poc.md, docs/evals/results/ockp-skill-reduction-runner-heuristics.md
---

# Skill Reduction Into Runner Heuristics ADR

## Context

The `oc-uj2y.5` track reduces `skills/openclerk/SKILL.md` toward a compact
activation, routing, and safety map. Workflow recipes that require exact JSON,
exact command ordering, or prompt choreography belong in runner-owned help,
workflow actions, `agent_handoff`, tests, or maintainer/eval docs.

Required reference URLs:

- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Candidate Options

| Candidate | Safety | Capability | UX quality | Decision |
| --- | --- | --- | --- | --- |
| Keep current skill text only | Safe but close to the line-count budget. | Works for current tests. | Weak because routine surface choice still depends on skill prose. | Not enough. |
| Compact `SKILL.md` plus help text only | Safe. | Good for users who already know which help command to run. | Better, but still leaves surface selection outside the runner. | Keep as base. |
| `workflow_guide_report` | Pass. Read-only and no storage inspection. | Maps routine intent to promoted runner surfaces and returns `agent_handoff`. | Pass. Moves surface selection into runner-owned JSON. | Promote. |
| Add one new action per workflow | Safe only when each workflow has separate evidence. | Strong for mature workflows. | Too much product surface for routing guidance alone. | Use only after separate promotion. |

## Decision

Promote `workflow_guide_report` as the read-only runner-owned workflow surface
selection report, and shrink `SKILL.md` below a stricter budget.

The skill remains the activation and no-tools boundary. The runner now owns a
compact intent-to-surface guide that points agents to promoted workflow actions
or current primitives without teaching long recipes in the skill.

## Non-Goals

- No autonomous router that executes the selected action.
- No storage inspection, document lookup, source fetch, or durable write.
- No replacement for source-sensitive evidence from the selected action.
- No hidden policy that can override runner rejections or approval boundaries.

## Promotion And Kill Criteria

Future skill growth should fail unless it is compact safety/routing guidance
or is paired with a runner-owned surface comparison. Kill any guide surface
that executes actions automatically, hides runner validation boundaries, or
encourages routine agents to bypass installed runner JSON.
