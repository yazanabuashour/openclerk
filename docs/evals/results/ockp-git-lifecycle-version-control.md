# Eval Result: Git Lifecycle Version Control

Lane: `git-lifecycle-version-control`

Required references:

- [`../../architecture/agent-knowledge-plane.md`](../../architecture/agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Summary

The targeted reduced eval promotes local Git status/history reporting and an
explicit config-gated checkpoint mode for OpenClerk-authored durable writes.
Restore/rollback remains outside the promoted surface.

## Results

| Scenario | Safety pass | Capability pass | UX quality | Performance | Evidence posture |
| --- | --- | --- | --- | --- | --- |
| Status after approved write | Pass: local-only Git metadata, no raw diff, no write. | Pass: returns dirty vault-relative paths and `write_status: no_write`. | Pass: one runner call replaces manual Git status. | Unit fixture completes inside normal runner test time. | Git status is storage metadata; OpenClerk provenance/freshness remains product evidence. |
| Checkpoint without config | Pass: durable Git write is rejected by default. | Pass: reports `checkpoint_status: disabled` and `write_status: rejected`. | Pass: boundary is visible in JSON. | Unit fixture completes inside normal runner test time. | Approval-before-write preserved. |
| Checkpoint with config | Pass: only local `git add`/`git commit` for supplied paths; no push, branch switch, restore, or raw diff. | Pass: creates one local checkpoint and returns commit id. | Pass: one runner call after approved write. | Unit fixture completes inside normal runner test time. | Commit id is storage history, not canonical authority. |
| SQLite default-enabled checkpoint config | Fail for this track: persisted enablement could make later checkpoint writes surprising. | Partial: would reduce invocation setup after opt-in. | Fail: normal users should see durable checkpoint approval at the call boundary. | Not implemented. | Rejected in favor of `--git-checkpoints` or `OPENCLERK_GIT_CHECKPOINTS=1`. |
| History after checkpoint | Pass: commit metadata only, no raw diff. | Pass: returns path-scoped local commit metadata. | Pass: one runner call replaces manual Git log. | Unit fixture completes inside normal runner test time. | Product claims still require citations, provenance, and projection freshness. |
| Invalid mode/path | Pass: JSON rejection before storage work. | Pass: unsupported restore-like modes are not accepted. | Pass: exact rejection text. | No storage open. | No bypass or destructive operation. |

## Taste Check

A normal user would expect a simpler OpenClerk surface for local status and
checkpoints around approved durable writes. The promoted surface meets that
expectation without turning Git into semantic truth or adding restore behavior.
Automatic checkpointing would reduce steps further, but it hides write timing
and risks surprising commits. The explicit config-gated checkpoint mode is the
best fit for this evidence.

## Implementation Evidence

Targeted tests:

- `TestDocumentTaskGitLifecycleStatusAndHistory`
- `TestDocumentTaskGitLifecycleCheckpointRequiresConfig`
- `TestSubcommandHelpShowsPromotedWorkflowActions`
- `TestOpenClerkSkillUsesInstalledRunnerForRoutineWork`
- `TestOpenClerkSkillKeepsWorkflowPoliciesCompact`

Quality-gate command for this reduced eval:

```bash
mise exec -- go test ./internal/runner ./cmd/openclerk ./internal/skilltest
```

## Classification

- Safety pass: pass.
- Capability pass: pass.
- UX quality: promote the explicit report/checkpoint surface.
- Restore/rollback: not promoted; destructive restore remains outside this
  track.

## Closure

Remaining work is represented by linked beads:

- `oc-tnnw.3.4` promotion decision.
- `oc-tnnw.3.5` conditional implementation only if promoted.
- `oc-tnnw.3.6` iteration and follow-up bead creation.
