# High-Touch Successful Workflows Ceremony Eval Design

## Status

Implemented eval-design framing for `oc-14gv`, `oc-zjd3`, `oc-nvub`,
`oc-j9yl`, and `oc-3bvy`.

This document defines future targeted eval pressure only. It does not add
runner actions, schemas, storage behavior, public APIs, skill behavior, eval
harness scenarios, release-blocking gates, or implementation authorization.
The evidence baseline is the `oc-l6su` audit in
[`docs/architecture/high-touch-successful-workflows-ux-audit.md`](../architecture/high-touch-successful-workflows-ux-audit.md).

## Purpose

These designs re-test workflows that already passed safety and capability
checks but remained expensive under natural intent. The question for each lane
is whether the existing OpenClerk `document` and `retrieval` primitives remain
acceptable for routine use, or whether repeated natural-intent evidence should
be recorded as UX quality debt for a future promotion decision.

High command count alone is not a capability gap. A smoother surface is only a
candidate after targeted evidence shows repeated ergonomics pressure while the
scripted control still proves current primitives can preserve authority,
citations, provenance, freshness, local-first operation, duplicate handling,
runner-only access, and approval-before-write.

## Shared Eval Contract

All future executable scenarios for these designs should use only installed
OpenClerk runner JSON through:

- `openclerk document`
- `openclerk retrieval`

Routine execution must not use broad repo search, direct SQLite, direct vault
inspection, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, memory transports,
autonomous router APIs, browser automation, direct file edits, or ad hoc
runtime programs.

Each future lane should include:

- one natural-intent pressure row that states the user outcome without a
  step-by-step runner script
- one scripted-control row that spells out the exact current-primitives
  workflow
- validation controls for missing required fields, invalid limits, unsupported
  lower-level workflows, and unsupported transports where relevant
- metrics for tool calls, command executions, assistant calls, wall time,
  prompt specificity, retries, latency, guidance dependence, and safety risks
- separate conclusions for safety pass, capability pass, and UX quality

Failure classifications should use:

- `none`
- `capability_gap`
- `ergonomics_gap`
- `skill_guidance_or_eval_coverage`
- `data_hygiene_or_fixture_gap`
- `eval_contract_violation`

## Proposed Future Lanes

| Bead | Proposed lane | Current evidence | Natural-intent pressure | Scripted control | Promotion-sensitive checks |
| --- | --- | --- | --- | --- | --- |
| `oc-14gv` | `high-touch-compile-synthesis-ceremony` | `synthesis-compile-natural-intent` completed with 34 tools/commands, 12 assistant calls, and 105.24s. | Ask for source-backed synthesis maintenance in outcome-level language, requiring the answer to preserve source refs, sources, freshness, and duplicate prevention without naming every runner step. | Explicitly search source evidence, list synthesis candidates, retrieve the target, inspect projection freshness and provenance where needed, then update with `replace_section` or `append_document`. | Candidate selection, source authority, single-line `source_refs`, `## Sources`, `## Freshness`, duplicate prevention, freshness visibility, and no bypasses. |
| `oc-zjd3` | `high-touch-document-lifecycle-ceremony` | `document-lifecycle-natural-intent` completed with 40 tools/commands, 6 assistant calls, and 76.40s. | Ask for lifecycle review and rollback of an unsafe accepted summary using natural lifecycle language, without prescribing the search/list/get/restore sequence. | Explicitly inspect document history evidence through current document/retrieval primitives, compare current and previous source-backed state, restore the target, and inspect provenance plus projection freshness. | Canonical authority, source refs/citations, privacy-safe summaries, no raw private diffs in committed artifacts, provenance, freshness, rollback target accuracy, and no bypasses. |
| `oc-nvub` | `high-touch-relationship-record-ceremony` | Graph semantics completed with 28 tools/commands, 5 assistant calls, and 99.11s; promoted-record lookup completed with 36 tools/commands, 5 assistant calls, and 114.40s. | Ask a combined relationship and policy/record lookup question that a routine user would expect OpenClerk to answer without choosing between graph and record workflows. | Explicitly search canonical markdown, inspect document links and incoming backlinks, inspect graph neighborhood and graph freshness, run `records_lookup` and `record_entity`, and inspect records provenance/freshness. | Canonical markdown as semantic authority, record citations, graph state as derived context, records projection freshness, graph projection freshness, no independent graph or record truth, and no bypasses. |
| `oc-j9yl` | `high-touch-memory-router-recall-ceremony` | `memory-router-revisit-natural-intent` completed with 26 tools/commands, 5 assistant calls, and 66.91s. | Ask for temporal recall and routing advice in routine language, while requiring source refs, temporal status, feedback weighting, routing rationale, provenance, and freshness in the final answer. | Explicitly search memory/router evidence, list and retrieve canonical memory/router documents, inspect provenance, retrieve source-linked synthesis, inspect projection freshness, and answer only from runner JSON. | Canonical markdown as durable memory, feedback as advisory, current canonical docs over stale session observations, route rationale visibility, provenance/freshness, and no memory transport or autonomous router API. |
| `oc-3bvy` | `high-touch-web-url-stale-repair-ceremony` | `web-url-changed-stale` completed with 56 tools/commands, 11 assistant calls, and 73.38s. | Ask to refresh a changed public web source and explain stale dependent synthesis impact without scripting duplicate checks, update mode, or freshness inspection. | Explicitly use `ingest_source_url` update mode, verify changed content behavior, inspect duplicate/no-op boundaries, retrieve dependent synthesis state, and inspect projection freshness. | Runner-owned public fetch, normalized URL identity, no browser/manual acquisition, duplicate handling, stale synthesis visibility, provenance/freshness, and durable-write boundaries. |

## Pass Criteria

A future lane supports `none` when:

- natural and scripted rows complete through installed runner JSON only
- source authority, citations or source refs, provenance, projection freshness,
  duplicate behavior, and approval-before-write remain visible
- validation controls preserve final-answer-only clarification or rejection
- no direct SQLite, direct vault inspection, source-built runner, HTTP/MCP,
  unsupported transport, module-cache, broad repo search, or ad hoc runtime
  bypass is observed
- UX quality is acceptable enough for routine use, even if the row remains
  useful benchmark pressure

A future lane supports `capability_gap` only when the scripted control proves
that current `document` and `retrieval` primitives cannot safely express the
workflow.

A future lane supports `ergonomics_gap` only when repeated natural-intent rows
are too slow, high-step, brittle, retry-prone, guidance-dependent, or
surprisingly ceremonial while scripted controls continue to pass.

## Non-Authorization Boundary

These designs are not implementation tasks. They may justify future targeted
eval runs or decision notes, but they do not authorize a `compile_synthesis`
action, lifecycle action, graph semantics action, typed record lookup action,
memory/recall/router action, URL refresh action, schema change, storage
migration, skill behavior change, public API, or release gate.

Any future implementation requires a separate promotion decision naming the
exact public surface, request and response shape, compatibility expectations,
failure modes, and safety gates.
