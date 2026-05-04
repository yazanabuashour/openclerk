---
decision_id: adr-git-lifecycle-version-control
decision_title: Local Git Lifecycle Version Control
decision_status: accepted
decision_scope: git-lifecycle-version-control
decision_owner: platform
---
# ADR: Local Git Lifecycle Version Control

## Status

Accepted for targeted POC/eval. Product behavior is authorized only by the
promotion decision in
[`git-lifecycle-version-control-promotion-decision.md`](git-lifecycle-version-control-promotion-decision.md).

Required references:

- [`agent-knowledge-plane.md`](agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Context

OpenClerk already treats markdown as the canonical durable knowledge format,
with provenance events and projection freshness as product evidence. Git is a
useful storage-history layer for local markdown vaults, but it must not become
OpenClerk truth, semantic review state, or a hidden restore mechanism.

The lifecycle pressure is practical: after an approved OpenClerk durable write,
an operator may want to know whether the local vault is dirty, see recent local
storage checkpoints for the affected paths, or create a local checkpoint. A
normal user would expect that simpler surface instead of hand-running Git, but
the surface must preserve AgentOps, local-first operation, citations/source
refs, provenance, projection freshness, approval-before-write, and no-bypass
boundaries.

## Options

| Option | Safety | Capability | UX quality | Decision posture |
| --- | --- | --- | --- | --- |
| Current primitives plus manual Git | Safe when the operator does it, but routine agents leave the runner to inspect storage. | Can answer status/history and create commits. | Too ad hoc for a runner-owned lifecycle workflow. | Reference only. |
| Read-only Git status/history report | Safe if it emits metadata only and no raw diffs. | Answers local dirty state and storage-history questions. | Good for routine inspection. | Promote as part of a narrow lifecycle report. |
| Explicit local checkpoint mode | Safe if disabled by default, path-scoped, local-only, and never pushes/switches/restores. | Creates a local storage checkpoint around approved durable writes. | Good; avoids surprising Git commands after every write. | Promote behind config. |
| Automatic checkpoint on every OpenClerk write | Risky because it can commit surprising changes or over-serialize local work. | Creates checkpoints without caller ceremony. | Convenient but too implicit. | Do not promote. |
| Restore/rollback from Git | Dangerous because it is destructive storage mutation and not semantic lifecycle repair. | Can recover bytes, but cannot prove source authority or projection freshness. | Too broad for this track. | Defer/kill for this surface. |

## Decision Model

Promote only a document-side lifecycle report:

```json
{"action":"git_lifecycle_report","git_lifecycle":{"mode":"status","paths":["synthesis/example.md"],"limit":10}}
```

The promoted modes are:

- `status`: read-only local Git dirty-path metadata for optional
  vault-relative paths.
- `history`: read-only local Git commit metadata for optional vault-relative
  paths.
- `checkpoint`: explicit local `git add` and `git commit` for caller-supplied
  vault-relative paths only, disabled unless runner config enables local
  checkpoints.

Default behavior:

- `mode` defaults to `status`.
- checkpoint writes are disabled by default.
- checkpoint writes require `--git-checkpoints` or
  `OPENCLERK_GIT_CHECKPOINTS=1`.
- checkpoint mode requires `git_lifecycle.paths` and `git_lifecycle.message`.

## Safety Constraints

- No branch creation, branch switching, checkout, reset, restore, rebase,
  remote push, pull, fetch, merge, or destructive file recovery.
- No raw private diffs or file bodies in public reports.
- Paths must be vault-relative and stay inside the vault root.
- Git metadata is storage-level evidence only.
- Canonical markdown, citations/source refs, provenance events, projection
  freshness, and OpenClerk write results remain the product evidence.
- Routine work remains through the installed `openclerk` runner; no direct
  SQLite, source-built runner, HTTP/MCP, raw vault inspection, or unsupported
  transport becomes part of the workflow.

## Promotion And Kill Criteria

Promote when targeted evals show status/history/checkpoint reporting can
reduce ceremony for OpenClerk-authored durable writes while preserving
local-only execution, path scoping, no raw diffs, no restore, and product
authority boundaries.

Kill or defer any candidate that treats Git as canonical truth, hides
provenance/freshness, commits broad unrelated paths, runs remote operations,
changes branches, restores bytes, exposes private raw diffs, or asks routine
agents to leave the installed runner.

## Non-Goals

- no semantic review queue
- no Git-backed canonical record store
- no automatic checkpointing of every write
- no restore/rollback operation
- no remote sync or collaboration workflow
- no public raw diff surface
