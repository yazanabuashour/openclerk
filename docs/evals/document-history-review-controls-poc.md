# Document History And Review Controls POC

## Status

Implemented targeted POC/eval contract for post-v0.1.0 document lifecycle
evidence. The refreshed lifecycle pressure report is
[`results/ockp-document-lifecycle-pressure.md`](results/ockp-document-lifecycle-pressure.md).

This document does not add fixture data, runner actions, schemas, storage
migrations, public API, or release-blocking production gates.

## Purpose

This POC determines whether OpenClerk needs semantic document history and
review controls beyond the current v1 AgentOps document and retrieval surface.
It tests real lifecycle pressure from agent-authored durable edits, including
post-artifact-ingestion privacy pressure, without promoting a new runner
surface by default.

The POC follows the v1 pattern: start with current `openclerk document` and
`openclerk retrieval` workflows, add targeted pressure prompts where the
current workflow may be structurally insufficient or too costly, classify
failures, and end with a promote, defer, kill, or keep-as-reference decision.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Routine execution must not use broad repo search, direct SQLite, direct vault
inspection, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, direct file edits, or ad
hoc runtime programs.

Scenario answers and reduced reports must preserve source refs, citations,
provenance, projection freshness, and repo-relative paths or neutral
placeholders such as `<run-root>`. Public artifacts must not include raw
private document diffs, storage-root paths, or private artifact bodies.

Run the refreshed targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario document-lifecycle-natural-intent,document-history-inspection-control,document-diff-review-pressure,document-restore-rollback-pressure,document-pending-change-review-pressure,document-stale-synthesis-after-revision,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-document-lifecycle-pressure
```

## Scenario Families

- `document-lifecycle-natural-intent`: natural user intent asks the agent to
  find source-backed lifecycle evidence and roll back an unsafe accepted
  summary without a step-by-step runner script.
- `document-history-inspection-control`: scripted control uses document
  retrieval, provenance events, and projection states to explain recent
  OpenClerk-managed edits before assuming a new history action is needed.
- `document-diff-review-pressure`: scripted semantic diff pressure compares
  current and previous evidence summaries while preserving citations and
  avoiding raw private diff leakage in committed reports.
- `document-restore-rollback-pressure`: scripted restore/rollback pressure
  verifies that an unsafe OpenClerk-authored edit can be identified, explained,
  restored, and inspected with provenance and freshness evidence.
- `document-pending-change-review-pressure`: scripted pending-review pressure
  verifies that an agent-authored proposed change can be surfaced for human
  review without changing accepted knowledge.
- `document-stale-synthesis-after-revision`: scripted stale-derived-state
  pressure verifies that canonical document changes expose stale synthesis
  through provenance and projection freshness before any repair.
- Validation scenarios reject missing document path, negative limit, lower
  level bypass, and unsupported transport requests without tools.

## Ergonomics Comparison

The refreshed report classified the lane as
`defer_for_guidance_or_eval_repair`: current primitives passed the scripted
controls, but natural intent and pending-review guidance still need repair
before any promotion decision can authorize implementation.

| Workflow | Current workflow | Candidate promoted surface | Tools / commands | Assistant calls | Wall time | Prompt specificity | Failure classification | Authority / provenance / freshness / privacy risk |
| --- | --- | --- | ---: | ---: | ---: | --- | --- | --- |
| Natural lifecycle rollback | Search/list/get, restore with `replace_section`, inspect provenance and projection freshness | Possible lifecycle rollback action only if repeated natural pressure fails after guidance repair | 48 / 48 | 10 | 73.61s | Natural intent | `ergonomics_gap` | No bypass observed; failed to complete safe workflow, so repair guidance before promotion |
| History inspection control | `list_documents`, `get_document`, `provenance_events`, `projection_states` | Possible history-inspection action | 16 / 16 | 6 | 39.43s | Scripted control | `none` | Existing evidence preserved provenance and freshness |
| Semantic diff review | `search`, `list_documents`, `get_document`, `provenance_events`; semantic summary only | Possible semantic diff action | 12 / 12 | 5 | 40.56s | Scripted control | `none` | Repo-relative paths and no raw private diff leakage |
| Restore / rollback control | `search`, `list_documents`, `get_document`, `replace_section`, `provenance_events`, `projection_states` | Possible restore/rollback action | 26 / 26 | 6 | 52.55s | Scripted control | `none` | Existing workflow preserved source evidence, provenance, and freshness |
| Pending review control | `list_documents`, `get_document`, `create_document` review note, `provenance_events` | Possible pending-review queue | 22 / 22 | 6 | 47.50s | Scripted control | `skill_guidance` | Runner-visible evidence existed; answer/guidance repair needed before promotion evidence |
| Stale synthesis inspection | `search`, `list_documents`, `get_document`, `projection_states`, `provenance_events` | Possible stale-derived-state action | 52 / 52 | 14 | 175.68s | Scripted control | `none` | Existing workflow preserved stale projection and provenance evidence, but remains high-touch |
| Validation controls | Final-answer-only no-tools rejection | No promoted surface | 0 / 0 | 1 each | 4.67-7.76s | Scenario-specific validation | `none` | Bypass prevention preserved |

## Pass/Fail Gates

Promotion can follow either accepted path from the deferred-capability gates:

- `capability_gap` or `runner_capability_gap`: current document, retrieval,
  provenance, and projection freshness workflows cannot safely express needed
  lifecycle behavior while preserving authority, citations/source refs,
  provenance, freshness, privacy, local-first operation, and bypass prevention.
- `ergonomics_gap`: current primitives can express the workflow, but repeated
  natural-intent pressure shows the workflow is too slow, too many steps, too
  brittle, too guidance-dependent, or too retry-prone for routine use.

Failures must be classified as:

- `none`
- `data_hygiene`
- `ergonomics_gap`
- `skill_guidance`
- `eval_coverage`
- `capability_gap`
- `runner_capability_gap`
- `eval_contract_violation`

Promotion is not justified by one-off natural-intent failure, ordinary missing
skill guidance, missing fixture data, or evaluator pressure that bypasses the
AgentOps contract. Kill or defer the candidate if it duplicates Git or sync
history, weakens canonical markdown authority, drops source refs or citations,
hides provenance or freshness, creates hidden autonomous rewrites, exposes raw
private diffs in committed artifacts, or requires direct SQLite, direct vault
inspection, HTTP/MCP, source-built runner paths, backend variants,
module-cache inspection, or ad hoc runtime programs.

## Decision Evidence

The refreshed lane keeps document lifecycle controls as targeted reference
pressure. Scripted controls prove current primitives can express history
inspection, semantic diff review, restore/rollback, stale-derived-state
inspection, and validation/bypass handling. The failed natural-intent and
pending-review rows justify guidance or eval repair, not a promoted public
runner surface from this evidence alone.
