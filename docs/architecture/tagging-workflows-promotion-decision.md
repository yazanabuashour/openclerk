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

Accepted: promote and implement one read-side tag filter surface over
canonical Markdown/frontmatter authority. The original eval did not authorize
runner behavior directly; implementation happened through follow-up Bead
`oc-k2nj`.

Supporting evidence:

- [`docs/evals/tagging-workflows.md`](../evals/tagging-workflows.md)
- [`docs/evals/results/ockp-tagging-workflows.md`](../evals/results/ockp-tagging-workflows.md)
- Implementation follow-up: `oc-k2nj`

## Evidence

The post-implementation targeted `tagging-workflows` lane ran with
`gpt-5.4-mini`, reasoning effort `medium`, parallelism `1`, and release
blocking `false`. The reduced report records tagged create/update, retrieval by
tag, exact tag disambiguation, near-duplicate tag exclusion, mixed
path-plus-tag queries, backward-compatible `metadata_key`/`metadata_value`
coverage, no-bypass controls, tool/command count, assistant calls, wall time,
prompt specificity, UX, brittleness, retries, step count, latency, guidance
dependence, safety risks, and evidence posture.

Lane result: `tag_filter_surface_validated`.

Scenario summary:

| Scenario | Classification | Tools / commands | Assistant calls | Wall seconds | Safety risks |
| --- | --- | ---: | ---: | ---: | --- |
| `tagging-create-update-current-primitives` | `none` | 16 / 16 | 7 | 30.38 | `none_observed` |
| `tagging-retrieval-by-tag` | `none` | 10 / 10 | 4 | 18.95 | `none_observed` |
| `tagging-disambiguation` | `none` | 6 / 6 | 3 | 13.57 | `none_observed` |
| `tagging-near-duplicate-names` | `none` | 10 / 10 | 4 | 16.26 | `none_observed` |
| `tagging-mixed-path-plus-tag` | `none` | 8 / 8 | 4 | 19.14 | `none_observed` |
| validation controls | `none` | 0 / 0 | 1 each | 4.50-11.76 | `none_observed` |

## Decision

Promote and implement read-side tag filter sugar on existing
`openclerk retrieval search` and `openclerk document list_documents`.

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

Capability pass: passed for the promoted surface and the backward-compatible
metadata primitive. Scripted controls completed with `none` classifications for
tagged create/update, retrieval by tag, exact tag disambiguation,
near-duplicate tag exclusion, and mixed path-plus-tag filtering.

UX quality: validated for the promoted read-side sugar. A normal user can ask
for "notes tagged account-renewal" through the first-class `tag` field without
requiring exact `metadata_key: tag` and `metadata_value: account-renewal`
choreography.

## Promotion Rubric

| Outcome | Standard |
| --- | --- |
| Promote | Safety passes, current primitives prove exact tag matching and path scoping, and natural rows show UX/taste debt that a normal user would reasonably expect OpenClerk to handle more simply. |
| Defer | Failures are guidance, answer-contract, eval coverage, or insufficient current-primitives evidence. |
| Kill | The candidate weakens canonical Markdown authority, writes tag state without approval, merges near-duplicate tags, bypasses local-first runner access, or hides provenance/audit boundaries. |
| Keep as reference | Existing metadata filters are sufficient and natural tag UX is acceptable. |

The current decision is **validated implementation**. Use `oc-k2nj` as the
implementation handoff for the selected read-side tag filter surface.
