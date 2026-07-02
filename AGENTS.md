- For all committed docs, reports, and artifact references, use repo-relative paths or neutral repo-relative placeholders. Never use machine-absolute filesystem paths.
- For repo-pinned developer tools declared in `mise.toml`, run commands through `mise exec -- ...` so agents use the same tool versions as local docs and CI.

## ADR/POC/Eval Decision Taste Review

When doing OpenClerk ADR, POC, eval, promotion, or deferred-capability decision work, keep the existing evidence discipline but add a taste check before accepting a defer/reference outcome:

- Ask whether a normal user would expect a simpler OpenClerk surface than the one being preserved.
- Distinguish read/fetch/inspect permission from durable-write approval. A public user-provided URL can be enough to fetch through the runner; approval belongs at durable writes, privileged access, purchases/actions, or other irreversible changes.
- Prefer extending the natural existing runner action when the input clearly belongs there, instead of declaring the adjacent UX unsupported.
- Treat "completed but ceremonial" eval passes as possible taste debt when they require high step count, long latency, exact prompt choreography, or surprising clarification turns.
- Record safety pass, capability pass, and UX quality separately when a report or decision needs to justify defer/reference.
- When taste debt, defer, keep-as-reference, or another non-promotion outcome still leaves a real capability, ergonomics, safety, auditability, or workflow need, identify whether the evaluated shape failed while the need remains valid. If it does, continue iterating on the candidate-surface comparison before handoff when feasible, normally with 2-3 plausible shapes unless the decision documents why only one is viable. The follow-up must compare candidates, choose the best, combine useful behaviors if appropriate, defer or kill the track, or record `none viable yet`.
- Before closing any ADR, POC, eval, promotion, or deferred-capability decision epic with outcome `keep-as-reference`, `defer`, `more evidence`, `candidate selected`, `none viable yet`, or another non-promotion result, check existing follow-up work in public docs. If none exists and the work cannot be resolved in the current checkpoint, record the remaining comparison need in the decision or handoff instead of opening a new issue unless the maintainer explicitly asks.
- Do not use taste review to bypass safety or evidence discipline: authority, citations, provenance, freshness, local-first behavior, duplicate handling, runner-only access, approval-before-write, and ADR/POC/eval/promotion decisions still apply.

## Agent Orchestration

For non-trivial work, proactively choose the smallest effective orchestration pattern. Use extra agents only when they reduce risk, improve independent coverage, or save enough wall-clock time to justify added tokens and integration overhead.

Prefer:
- the main session for architecture, task decomposition, integration, final verification, review sequencing, and commits;
- read-only subagents for scoped investigation, source/API verification, test-gap analysis, fixture ideas, security/correctness/complexity review, and other work that can return evidence without owning design;
- isolated worktree sessions for parallel implementations, dependency experiments, large migrations, risky edits, or work that would disturb the active checkout;
- many-agent or batch workflows only for broad audits, large mechanical migrations, cross-checked research, or file-partitioned work with an explicit merge and review plan.

Use orchestration without waiting for a second prompt when the task is clearly decomposable and the harness supports it. Good triggers include unfamiliar codebase exploration, broad review, multi-file migrations, competing implementation approaches, dependency/toolchain experiments, difficult bug hunts, performance/correctness investigations, and work that benefits from independent reviewer perspectives.

Do not use extra agents for tightly coupled edits, small one-file fixes, vague exploratory churn, or tasks where merge/integration overhead would dominate.

Do not delegate away architectural ownership. Parallel agents must report scope, files inspected or changed, commands run, result, risks, and next recommended action. The main session integrates all work under the normal review and commit workflow.

Prefer cheaper/faster models for read-only scans and supporting subagents. Use the strongest available coding model with higher reasoning effort for architecture, complex implementation, high-risk review, or final integration. Put exact model IDs in harness config, agent definitions, skills, or scripts rather than here.

Use goals only for long-running work with a verifiable stop condition, such as passing tests/builds/evals, a benchmark result, a scan report, or a committed artifact. Open-ended goals must include a turn, time, or attempt bound.

## Completion Contract

For each completed work item or independent review checkpoint:

1. Record remaining follow-up work as issues or in the repo-local backlog.
2. Run the relevant quality gates for changed code.
3. Run `scripts/codex-review-sequence.sh` exactly once before commit.
4. Address review findings without rerunning the same checkpoint review unless explicitly requested.
5. Commit the intended files locally after gates and review are complete.
6. Push only when explicitly requested by the maintainer or task.
7. Hand off with changed files, gates run, review result, commit hash, and remaining risks.

Review sequence defaults and optional extras:
- The standard checkpoint command is `scripts/codex-review-sequence.sh`; it runs the built-in uncommitted review and avoidable-complexity review.
- Add focused reviewers only when the change warrants them, using `CODEX_REVIEW_EXTRA=test-gaps`, `security`, `api-compat`, or `concurrency`.
- Use `test-gaps` for behavior changes, bug fixes, migrations, or weak validation risk.
- Use `security` for auth, permissions, shell/filesystem/network/browser/URL handling, secrets, or dependency-sensitive changes.
- Use `api-compat` for public CLI/API/config/env/schema/docs contract changes.
- Use `concurrency` for async, lifecycle, retry, cache/state, transactionality, or parallelism changes.
- Combine extras with commas, for example `CODEX_REVIEW_EXTRA=security,test-gaps scripts/codex-review-sequence.sh`.
- Agents may override `CODEX_REVIEW_MODEL` and `CODEX_REVIEW_EFFORT` for the whole run when needed. Exact default model IDs belong in the script, not here.

Never push without explicit request. Never commit before quality gates and review.
