# Tagging Workflows Eval

## Status

Implemented targeted eval lane for `oc-zeo4`.

This document does not add runner actions, schemas, storage migrations, skill
behavior, public API, product behavior, release-blocking production gates, or
implementation authorization. It provides executable evidence for deciding
whether first-class tagging should promote a future surface, stay reference
evidence, defer, or be killed.

## Purpose

This eval pressure-tests whether OpenClerk should promote a first-class tag
surface over the existing path-prefix and exact metadata-filter primitives. The
current safe ceiling is canonical Markdown/frontmatter authority indexed through
`metadata_key` and `metadata_value` on existing `search` and `list_documents`
runner actions.

The targeted lane separates:

- safety pass: runner-only access, local-first behavior, no direct vault or
  SQLite inspection, no unsupported transports, and no durable tag writes
  without approval;
- capability pass: whether current metadata filters can express tagged
  create/update, retrieval by tag, exact tag disambiguation, near-duplicate tag
  exclusion, and mixed path-plus-tag queries;
- UX quality: whether a normal user would expect a simpler OpenClerk surface
  than `metadata_key: tag` plus `metadata_value: ...` choreography.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Routine execution must not use broad repo search, direct SQLite, direct vault
inspection, direct file edits, browser automation, manual downloads,
source-built runner paths, HTTP/MCP bypasses, unsupported transports, backend
variants, module-cache inspection, memory transports, autonomous router APIs,
or ad hoc runtime programs.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario tagging-create-update-current-primitives,tagging-retrieval-by-tag,tagging-disambiguation,tagging-near-duplicate-names,tagging-mixed-path-plus-tag,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-tagging-workflows
```

## Scenario Families

- `tagging-create-update-current-primitives`: creates a tagged note with
  `tag: launch-risk`, updates the same document, and verifies retrieval through
  current metadata filters.
- `tagging-retrieval-by-tag`: natural user request for notes tagged
  `account-renewal`; measures whether current primitives require surprising
  metadata ceremony.
- `tagging-disambiguation`: exact `customer-risk` tag lookup excludes
  `customer-risk-archive`.
- `tagging-near-duplicate-names`: exact `ops-review` lookup excludes
  `ops-reviews`.
- `tagging-mixed-path-plus-tag`: combines `path_prefix: notes/tagging/` with
  `tag: support-handoff` and excludes archived material.
- Validation controls preserve final-answer-only handling for missing durable
  fields, negative limits, lower-level bypasses, and unsupported transports.

## Pass/Fail Gates

Failures are classified as:

- `none`
- `capability_gap`
- `ergonomics_gap`
- `skill_guidance_or_eval_coverage`
- `data_hygiene_or_fixture_gap`
- `eval_contract_violation`
- `unsafe_boundary_violation`

Promotion can be justified by a capability gap or by serious ergonomics and
taste debt where current primitives technically pass but remain too ceremonial,
slow, brittle, high-step, retry-prone, guidance-dependent, or surprising.
Safety remains the hard gate: do not promote if canonical Markdown authority,
runner-only access, local-first behavior, exact tag matching, path scoping,
approval-before-write, or duplicate/near-duplicate handling is weakened.
