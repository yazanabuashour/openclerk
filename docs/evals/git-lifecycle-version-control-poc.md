# POC: Git Lifecycle Version Control

## Purpose

This POC compares local Git-backed lifecycle surfaces for OpenClerk-authored
durable knowledge. It is intentionally local-only and eval-facing: no remote
push, branch switch, destructive restore, raw private diff, direct SQLite,
source-built runner, HTTP/MCP, or raw vault inspection is allowed.

Required references:

- [`../architecture/agent-knowledge-plane.md`](../architecture/agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Candidate Surface

Promoted candidate:

```json
{"action":"git_lifecycle_report","git_lifecycle":{"mode":"status","paths":["notes/git-lifecycle.md"],"limit":10}}
```

Checkpoint candidate:

```json
{"action":"git_lifecycle_report","git_lifecycle":{"mode":"checkpoint","paths":["notes/git-lifecycle.md"],"message":"openclerk: checkpoint git lifecycle note"}}
```

The checkpoint candidate is evaluated only when runner config explicitly
enables local checkpoints. It is a local storage checkpoint for caller-named
paths, not approval to write knowledge content, restore content, or replace
OpenClerk provenance.

Rejected checkpoint-default candidate:

```json
{"action":"git_lifecycle_report","git_lifecycle":{"mode":"checkpoint","paths":["notes/git-lifecycle.md"],"message":"openclerk: checkpoint git lifecycle note"}}
```

This shape would become default-enabled through persisted SQLite config instead
of the `--git-checkpoints` or `OPENCLERK_GIT_CHECKPOINTS=1` invocation gate.
The POC rejects it because checkpoint commits are durable local writes. A
persisted default can make a later runner invocation commit unexpectedly, while
the explicit gate keeps the write boundary visible.

Restore/rollback candidates are intentionally represented only as rejected
validation pressure. The POC does not create a restore plan or write restored
bytes because that would be destructive storage mutation outside the promoted
surface.

## Fixture Shape

The deterministic unit fixture creates a local vault bound to an OpenClerk
database, initializes the vault as a Git repository, commits a `.gitkeep`
baseline, writes an OpenClerk document through `create_document`, reports Git
status, creates a checkpoint with explicit config, and reads history for the
same vault-relative path.

Committed reports use only repo-relative fixture descriptions and
vault-relative paths. No generated vaults, SQLite databases, raw logs, or
machine-absolute paths are committed.

## Scenarios

| Scenario | Expected behavior |
| --- | --- |
| Status after approved write | Report changed vault-relative paths, no raw diff, no write. |
| Checkpoint without config | Return `checkpoint_status: disabled` and `write_status: rejected`; no Git commit. |
| Checkpoint with config | Create one local commit for caller-specified paths and return the short commit id. |
| History after checkpoint | Return local commit metadata for the path without raw diff content. |
| Invalid mode or path | Return a JSON validation rejection before storage opens. |

## Acceptance

The POC passes only if safety pass, capability pass, and UX quality are recorded
separately in the eval result and the promotion decision names exact defaults,
config gates, write boundaries, and restore non-goals.

Remaining work is represented by linked beads:

- `oc-tnnw.3.3` eval for safety, capability, and UX quality.
- `oc-tnnw.3.4` promotion decision.
- `oc-tnnw.3.5` conditional implementation only if promoted.
- `oc-tnnw.3.6` iteration and follow-up bead creation.
