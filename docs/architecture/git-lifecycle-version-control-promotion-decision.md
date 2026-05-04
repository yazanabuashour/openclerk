---
decision_id: decision-git-lifecycle-version-control-promotion
decision_title: Git Lifecycle Version Control Promotion
decision_status: accepted
decision_scope: git-lifecycle-version-control
decision_owner: platform
---
# Decision: Git Lifecycle Version Control Promotion

## Status

Accepted: promote `openclerk document` `git_lifecycle_report` with read-only
`status` and `history` modes plus explicit config-gated local `checkpoint`
mode. Do not promote restore, rollback, branch switching, remote push, raw diff,
automatic checkpointing, or Git-backed canonical truth.

Required references:

- [`agent-knowledge-plane.md`](agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

Evidence:

- [`git-lifecycle-version-control-adr.md`](git-lifecycle-version-control-adr.md)
- [`../evals/git-lifecycle-version-control-poc.md`](../evals/git-lifecycle-version-control-poc.md)
- [`../evals/results/ockp-git-lifecycle-version-control.md`](../evals/results/ockp-git-lifecycle-version-control.md)

## Promoted Surface

```json
{"action":"git_lifecycle_report","git_lifecycle":{"mode":"status","paths":["synthesis/example.md"],"limit":10}}
```

```json
{"action":"git_lifecycle_report","git_lifecycle":{"mode":"checkpoint","paths":["synthesis/example.md"],"message":"openclerk: update synthesis example"}}
```

Defaults and config:

- `mode` defaults to `status`.
- `status` and `history` are read-only.
- `checkpoint` is disabled by default.
- `checkpoint` requires `--git-checkpoints` or
  `OPENCLERK_GIT_CHECKPOINTS=1`.
- `checkpoint` requires explicit vault-relative `paths` and a one-line
  `message`.

Write boundaries:

- only local `git add -- <paths>` and `git commit -m <message> -- <paths>`
- no remote operation
- no branch operation
- no checkout/reset/restore
- no raw diff output
- no automatic checkpoint after ordinary writes

## Decision

Promote the candidate because it passes safety, capability, and UX quality for
the targeted status/history/checkpoint pressure. It is a natural extension of
the existing document lifecycle surface: the user asks about durable document
state, and the runner can answer without requiring routine agents to leave
OpenClerk for manual Git commands.

Git remains storage-level history. It can say that local bytes changed or that
a local commit exists. It cannot say the knowledge is true, cited, fresh,
approved, or semantically restored. Product evidence remains canonical
markdown, citations/source refs, provenance events, projection freshness, and
the OpenClerk write result.

## Restore And Review Queue

Restore, rollback, review queue, and semantic lifecycle history are not
promoted by this decision. The evaluated need was local storage status and
checkpointing around approved writes. Destructive restore would require a
separate candidate-surface comparison that proves source authority, rollback
target accuracy, privacy-safe diff handling, provenance, projection freshness,
operator approval, and no-bypass behavior.

## Compatibility

Existing `openclerk document` and `openclerk retrieval` behavior remains
unchanged unless the caller explicitly asks for `git_lifecycle_report`.
Existing installs do not create checkpoints unless the caller enables the
config gate and asks for checkpoint mode.

## Follow-Up Search

The non-promoted restore/review-queue portion remains a future lifecycle need,
but this track promotes a complete checkpoint/reporting surface. Before
closing the decision, existing follow-up work was checked by Beads search:

- `bd search "git lifecycle restore" --status all`
- `bd search "git checkpoint restore plan" --status all`
- `bd search "document lifecycle restore queue" --status all`

No implementation follow-up is authorized here for restore. If future evidence
shows checkpoint/reporting is safe and useful but users still need restore-plan
or review-queue UX, open a candidate-surface comparison before implementing
restore behavior.
