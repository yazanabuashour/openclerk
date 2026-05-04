# Skill Reduction Into Runner Heuristics POC

## Scope

This POC compares surfaces for reducing `SKILL.md` workflow detail without
weakening safety, local-first behavior, runner-only access, citations,
provenance, freshness, duplicate handling, or approval-before-write.

## Candidate Shapes

| Shape | What It Proves | What It Does Not Prove |
| --- | --- | --- |
| Keep current skill text | Current skill can still guide agents. | Lower ceremony or durable line-count headroom. |
| Compact skill plus help text | Runner help owns request shapes. | Intent-level surface selection before help is consulted. |
| `workflow_guide_report` | Runner JSON can recommend the next surface and return `agent_handoff` without storage access. | That it answered the user's substantive source-sensitive question. |
| Per-workflow action expansion | Mature workflows can get dedicated reports. | That routing guidance alone needs many new actions. |

## Selected POC Surface

`workflow_guide_report`:

```json
{"action":"workflow_guide_report","workflow_guide":{"intent":"should I update an existing note or create a new one?"}}
```

The report returns:

- `recommended_surface`
- `runner_domain`
- `request_shape`
- `use_when`
- `do_not_use_for`
- `candidate_surfaces`
- `validation_boundaries`
- `authority_limits`
- `agent_handoff`

## Taste Review

A normal agent should not have to read a long workflow manual just to decide
whether the next safe step is `duplicate_candidate_report`,
`ingest_source_url` plan mode, `compile_synthesis`, or current primitives.
The selected report gives runner-owned routing guidance while keeping final
answers and durable writes on the selected action.
