---
decision_id: adr-tagging-workflows
decision_title: Tagging Workflows
decision_status: accepted
decision_scope: tagging
decision_owner: platform
decision_date: 2026-04-30
source_refs: docs/evals/tagging-workflows.md, docs/evals/results/ockp-tagging-workflows.md
---
# Tagging Workflows Promotion Decision

## Status

Accepted: promote one follow-up implementation Bead for read-side tag filter
sugar over canonical Markdown/frontmatter authority. The eval itself does not
authorize runner behavior, schema, storage, public API, skill behavior, or
product behavior changes.

Supporting evidence:

- [`docs/evals/tagging-workflows.md`](../evals/tagging-workflows.md)
- [`docs/evals/results/ockp-tagging-workflows.md`](../evals/results/ockp-tagging-workflows.md)
- Implementation follow-up: `oc-k2nj`

## Evidence

The targeted `tagging-workflows` lane ran with `gpt-5.4-mini`, reasoning effort
`medium`, parallelism `1`, and release blocking `false`. The reduced report
recorded tagged create/update, retrieval by tag, exact tag disambiguation,
near-duplicate tag exclusion, mixed path-plus-tag queries,
`metadata_key`/`metadata_value` ceremony, no-bypass controls, tool/command
count, assistant calls, wall time, prompt specificity, UX, brittleness, retries,
step count, latency, guidance dependence, safety risks, and evidence posture.

Lane result: `promote_tag_filter_surface_design`.

Scenario summary:

| Scenario | Classification | Tools / commands | Assistant calls | Wall seconds | Safety risks |
| --- | --- | ---: | ---: | ---: | --- |
| `tagging-create-update-current-primitives` | `none` | 14 / 14 | 6 | 45.46 | `none_observed` |
| `tagging-retrieval-by-tag` | `ergonomics_gap` | 12 / 12 | 4 | 28.31 | `none_observed` |
| `tagging-disambiguation` | `none` | 12 / 12 | 5 | 22.60 | `none_observed` |
| `tagging-near-duplicate-names` | `none` | 6 / 6 | 3 | 16.16 | `none_observed` |
| `tagging-mixed-path-plus-tag` | `none` | 6 / 6 | 3 | 22.99 | `none_observed` |
| validation controls | `none` | 0 / 0 | 1 each | 3.79-7.57 | `none_observed` |

## Decision

Promote one implementation Bead for read-side tag filter sugar on existing
`openclerk retrieval search` and `openclerk document list_documents`. Do not
implement it as part of this decision.

Selected shape: runner JSON should accept a natural tag filter alongside
existing text, path prefix, limit, cursor, and metadata filters. The selected
surface must remain sugar over canonical Markdown/frontmatter indexed metadata,
not a separate tag database or opaque authority layer.

Candidate comparison:

| Candidate | Decision | Reason |
| --- | --- | --- |
| Plain frontmatter convention plus `metadata_key`/`metadata_value` | Keep as backward-compatible primitive | Scripted controls proved it can preserve exact matching, path scoping, local-first behavior, and canonical authority, but normal tag lookup remained too ceremonial. |
| Runner-level tag filter sugar | Promote | It addresses the natural UX gap while keeping authority in Markdown/frontmatter and reusing current read-side retrieval/list behavior. |
| Explicit tag-management runner actions | Do not promote | The eval did not show a need for separate durable tag authority or tag writes outside ordinary document creation/update approval. |

Safety pass: passed. The targeted run did not observe direct SQLite, broad repo
search, source-built runner usage, module-cache inspection, unsupported
transport, local-first bypass, hidden tag authority, or unapproved durable tag
writes. Validation controls stayed final-answer-only.

Capability pass: passed for current primitives. Scripted controls completed
with `none` classifications for tagged create/update, exact tag
disambiguation, near-duplicate tag exclusion, and mixed path-plus-tag filtering
through existing metadata filters.

UX quality: not acceptable enough to keep only as reference pressure. A normal
user would expect OpenClerk to answer "notes tagged account-renewal" without
requiring exact `metadata_key: tag` and `metadata_value: account-renewal`
choreography. The natural row failed with an `ergonomics_gap`, while successful
scripted rows still required 6-14 command executions.

## Promotion Rubric

| Outcome | Standard |
| --- | --- |
| Promote | Safety passes, current primitives prove exact tag matching and path scoping, and natural rows show UX/taste debt that a normal user would reasonably expect OpenClerk to handle more simply. |
| Defer | Failures are guidance, answer-contract, eval coverage, or insufficient current-primitives evidence. |
| Kill | The candidate weakens canonical Markdown authority, writes tag state without approval, merges near-duplicate tags, bypasses local-first runner access, or hides provenance/audit boundaries. |
| Keep as reference | Existing metadata filters are sufficient and natural tag UX is acceptable. |

The current decision is **promote**. File and use `oc-k2nj` as the exact
implementation handoff for the selected read-side tag filter surface.
