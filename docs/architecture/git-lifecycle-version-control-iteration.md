# Iteration: Git Lifecycle Version Control

## Status

Iteration recorded after promoting and implementing `git_lifecycle_report`.

Required references:

- [`agent-knowledge-plane.md`](agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Implemented Shape

`openclerk document` now accepts `git_lifecycle_report`:

- `status`: read-only local Git dirty-path metadata.
- `history`: read-only local Git commit metadata.
- `checkpoint`: explicit local checkpoint for caller-specified paths when
  `--git-checkpoints` or `OPENCLERK_GIT_CHECKPOINTS=1` is enabled.

The implementation updates runner JSON types, CLI help, README action index,
OpenClerk skill guidance, workflow-guide routing, and tests.

## Post-Implementation Boundaries

The surface deliberately does not:

- push, pull, fetch, merge, rebase, switch branches, checkout, reset, restore,
  or rollback
- expose raw diffs or file bodies
- treat Git commits as citations, provenance, freshness, or approval evidence
- automatically checkpoint ordinary OpenClerk writes
- inspect SQLite or bypass the installed runner

## Follow-Up Trigger

Do not iterate into restore-plan or review-queue behavior from this evidence
alone. Consider a new candidate comparison only after real use shows that
status/history/checkpoint reporting is safe but still leaves repeated
restore-plan or review-queue ceremony. That future comparison must include at
least:

- current primitives plus `git_lifecycle_report`
- a read-only restore-plan report
- a semantic OpenClerk review queue

The future decision must choose, defer, kill, or record `none viable yet` and
must preserve authority, citations/source refs, provenance, freshness,
approval-before-write, privacy-safe summaries, local-first operation, and
no-bypass boundaries.
