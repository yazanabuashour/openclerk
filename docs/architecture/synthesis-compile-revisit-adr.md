---
decision_id: adr-synthesis-compile-revisit
decision_title: Synthesis Compile Revisit
decision_status: accepted
decision_scope: synthesis-compile-revisit
decision_owner: platform
---
# ADR: Synthesis Compile Revisit

## Status

Accepted as an evidence-gathering direction only. This ADR does not promote a
new runner action, schema, storage behavior, migration, public API, or shipped
skill behavior.

Supporting evidence:

- [`knowledge-configuration-v1-adr.md`](knowledge-configuration-v1-adr.md)
- [`../evals/results/ockp-synthesis-compiler-pressure.md`](../evals/results/ockp-synthesis-compiler-pressure.md)
- [`../evals/results/ockp-synthesis-maintenance-ergonomics.md`](../evals/results/ockp-synthesis-maintenance-ergonomics.md)
- [`../evals/synthesis-compile-revisit-comparison-poc.md`](../evals/synthesis-compile-revisit-comparison-poc.md)
- [`../evals/results/ockp-synthesis-compile-revisit-pressure.md`](../evals/results/ockp-synthesis-compile-revisit-pressure.md)
- [`synthesis-compile-revisit-promotion-decision.md`](synthesis-compile-revisit-promotion-decision.md)

## Context

OpenClerk source-linked synthesis currently works through the installed
`openclerk document` and `openclerk retrieval` JSON runners. The documented
workflow searches canonical source evidence, lists `synthesis/` candidates,
retrieves an existing synthesis before editing, inspects provenance and
projection freshness when relevant, and updates with `replace_section` or
`append_document`.

`compile_synthesis` remains a deferred candidate shape from Knowledge
Configuration v1:

```json
{
  "action": "compile_synthesis",
  "synthesis": {
    "path": "synthesis/example.md",
    "title": "Example",
    "source_refs": ["sources/source-a.md", "sources/source-b.md"],
    "body": "# Example\n\n## Summary\n...\n\n## Sources\n...\n\n## Freshness\n...",
    "mode": "create_or_update"
  }
}
```

This revisit evaluates two independent promotion paths:

- **Capability gap:** whether existing document and retrieval actions cannot
  safely express source-linked synthesis compilation.
- **Ergonomics gap:** whether the actions can express the workflow but require
  too many steps, too much prompt choreography, too much latency, or too much
  guidance for routine AgentOps use.

## Decision Frame

The candidate options are:

- **Keep current document/retrieval workflow:** no new public surface; continue
  using `search`, `list_documents`, `get_document`, `projection_states`,
  `provenance_events`, `replace_section`, and `append_document`.
- **Strengthen skill and eval guidance:** keep the runner surface unchanged,
  but improve prompts, scenario coverage, and classification if failures are
  ordinary guidance or eval coverage.
- **Promote narrow `compile_synthesis`:** add a small runner action only if
  targeted evidence shows repeated capability-gap or ergonomics-gap pressure
  that the current workflow cannot absorb.
- **Keep as reference or kill:** preserve the candidate as benchmark pressure,
  or kill it if it would hide authority, citations, provenance, freshness, or
  normalize bypasses.

Prior evidence is relevant but not sufficient by itself. The
`ockp-synthesis-compiler-pressure` run showed current primitives can complete
selected synthesis workflows, but it was not framed around the newer
capability-gap and ergonomics-gap split. The maintenance ergonomics decision
deferred promotion, but the revisit must test natural user intent separately
from scripted controls before closing the promotion question.

## Invariants

- Canonical markdown source documents remain authority; synthesis is derived
  compiled knowledge.
- Source-sensitive synthesis preserves source refs, citations, source paths,
  or stable source identifiers.
- `## Sources` and `## Freshness` remain visible in synthesis markdown.
- Provenance and projection freshness remain inspectable through runner JSON.
- Existing synthesis is updated rather than duplicated when a target already
  exists.
- Routine agents use installed OpenClerk runner JSON only, not direct SQLite,
  broad repo search, direct vault inspection, source-built command paths,
  HTTP/MCP bypasses, unsupported transports, backend variants, module-cache
  inspection, or ad hoc scripts.
- Invalid routine requests keep the no-tools validation contract.
- Committed docs, reports, and raw-log references use repo-relative paths or
  neutral placeholders.

## Non-Goals

This ADR does not:

- implement `compile_synthesis`
- change document or retrieval schemas
- add storage migrations, indexes, background jobs, or parser pipelines
- make synthesis higher authority than canonical sources or promoted records
- relax source refs, citation, provenance, freshness, duplicate, or validation
  requirements
- authorize any lower-level bypass of the installed OpenClerk runner

## Promotion Gate

Use the deferred capability rubric in
[`deferred-capability-promotion-gates.md`](deferred-capability-promotion-gates.md).
Promotion is allowed only if the POC and targeted eval evidence show repeated
pressure on at least one path:

- `capability_gap`: current primitives cannot safely express the workflow while
  preserving authority, citations, provenance, freshness, and duplicate
  prevention.
- `ergonomics_gap`: current primitives can express the workflow, but natural
  AgentOps use is unacceptably brittle, slow, step-heavy, or
  guidance-dependent, and the proposed surface reduces that cost without
  weakening invariants.

A promotion decision must name the exact request and response shape,
compatibility rules, failure modes, validation behavior, and follow-up Beads.
Without that targeted evidence, the default decision is defer or keep as
reference.
