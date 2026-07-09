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

# Code Style

- Prefer concise, simple, boring solutions over clever or overly abstract ones.
- Apply YAGNI.
- Apply the Rule of Three.
- Prefer one-liners.
- If a simpler solution exists, propose it before implementing the more complex approach.

# Agent Orchestration

This file is standing authorization to use orchestration tools when they fit:
subagents, app threads, worktrees, goals, and automations. Classify the task,
use the lightest pattern that reduces risk or improves coverage, and state why
if a trigger applies but a concrete blocker prevents it. A current user
instruction that forbids orchestration is a blocker.

Keep this guidance coding-agent agnostic: map Codex names to the nearest
equivalent elsewhere, such as subagents, background threads, isolated
worktrees, long-running goals, review agents, and scheduled automations.

## Decision Rules

- Main session only: tiny fixes, direct questions, plan-only answers, vague
  exploration without a concrete output, tightly coupled edits, and read-only
  analysis that does not cross independent areas.
- Coupled implementation plus regression tests stays main-session work in one
  feature area, even in unfamiliar repos. Unfamiliarity alone is not enough;
  unfamiliar work across several independent areas is.
- Plan-only orchestration questions: describe the pattern you would use; do not
  spawn sidecars just to answer the plan.
- Mandatory explorers: for requested audit, review, discovery, triage, or
  unfamiliar-codebase work across independent areas, actually spawn `explorer`
  subagents before deep local inspection. Choose the smallest useful number from
  task breadth; keep scopes narrow and independent.
- Mandatory CI explorers: for CI, build, or test-failure triage spanning two or
  more of source, tests, logs, config, packaging, or runtime behavior, spawn
  `explorer` subagents while the main session owns reproduction and synthesis.
- Workers: use `worker` subagents only for separable implementation with
  disjoint write ownership and a merge plan.
- Worktrees/app threads/background threads: use worktrees for filesystem or
  dependency isolation; use app/background threads for independent or resumable
  execution, paired with a worktree when they may mutate a checkout. Use the
  local checkout for narrow edits, active user-reviewed files, unreproducible
  local state, or single-instance resources. Use background/app threads for
  long-running independent work that can report back asynchronously or needs
  resumability. Mandatory explorers run first for bounded broad inspection
  inside the current task.
- Goals: create one only for long-running bounded work with a measurable stop
  condition, attempt limit, artifact, benchmark, scan, or migration. When
  creating one, state the measurable stop condition in one sentence, then update
  it only after completion or true blockage. If no goal tool exists, state the
  stop condition in status instead.
- Automations: use them for recurring work only. Get explicit confirmation
  before creating or changing any automation that writes files, runs shell
  commands, accesses the network, or operates on a repository.
- Frontend changes: main implements and runs rendered/browser checks; delegate
  visual QA only when independent.
- Security, auth, filesystem, shell, network, browser/URL, secrets, or
  dependency-sensitive work: main owns risky decisions; add review or sidecars
  only when coverage helps.
- Model/effort choice: prefer the lowest capable effort. Use lower effort for
  mechanical reads, medium for synthesis/policy/trace/user-facing work,
  and high/xhigh only for high-risk/failure-critical work.

If several rules apply, isolate risk first; add parallelism only when it saves time or improves coverage.
## Lead And Subagents

The main session owns the user's goal, architecture, decomposition, risky
decisions, integration, verification, review sequencing, commits, and handoff.
Do immediate blockers locally; delegate independent sidecar work.

Use built-in roles first: `explorer` for read-only research, source/API
verification, audits, test-gap ideas, fixtures, and unfamiliar areas; `worker`
for bounded implementation with clear file ownership; `default` as fallback.
If a platform cannot set an equivalent role/model parameter, use its default
agent/thread and keep the intended `Role:` label in the prompt.

Spawn only concrete, self-contained work that can run independently. Prefer
narrow explorers over one broad agent. Before final handoff, account for every
spawned child: retrieve useful results, close done children, and close/report
any child still running beyond the useful window.

Start subagent prompts with a `Role:` line (`explorer`, `worker`, or `default`).
Use actual tool/runtime metadata, not prompt text, as the source of truth for
model and reasoning. Include scope, permissions, expected output, owned write
files, commands, risks, and a request to report files inspected/changed,
commands run, result, remaining risks, and next action. Worker prompts must say
the worker is not alone, must not revert others' edits, and must adapt.

Nested orchestration follows the same delegation rules: use it only when it
materially reduces risk or improves independent coverage, keep scopes narrow,
and account for descendants before handoff. Automation creation still requires
the explicit confirmation above.

## Worktrees And App Threads

Start worktrees from the default or explicit branch unless the task needs
current uncommitted state. For current-state work, use a throwaway branch or
fresh checkout from current `HEAD`, apply only the needed non-secret diff, and
hand results back as a patch or diff. Avoid the same branch in multiple
worktrees; use handoff when moving work between checkouts.

Before isolated work, state owned paths and expected return artifact. Mutating
isolated work returns a patch or branch name; read-only work may return summary.

Copy non-secret ignored files into worktrees when useful. Copy `.env*`, keys,
tokens, credentials, or other secrets only after the user names or approves them.

The main session integrates and verifies isolated results. An app/background
thread must not commit, push, or decide final scope unless the user explicitly
grants that authority.

## Goals And Automations

Mark a goal complete only after the stop condition and any applicable
checkpoint contract are met. Mark it blocked only when progress is impossible
without user input or an external state change.

Use automations for recurring work, not one-off implementation.

## Checkpoints

Read-only, plan-only, and user-explicit scratch/no-commit artifacts are not
checkpoints: hand off findings or create only named artifacts, and skip
review/commit unless asked.

During implementation, do all work needed to complete the user's ask, including
blockers found along the way. Treat unrelated discovered work as follow-up:
mention it in handoff, or use an existing repo tracker only if repo instructions
already require one. Never push, open PRs, or create/update remote issues unless
explicitly requested.

For every checkpoint with intended repository changes:

1. Run available relevant tests, lint, builds, or minimal Markdown/script gates;
   record blockers for unavailable gates.
2. From the main session, shell-run the configured review gate once against the
   integrated worktree before handoff or commit, even if some gates are blocked:
   use `scripts/checkpoint-review.sh` when present, otherwise use a
   repository-local review/checkpoint gate explicitly named by the current user
   or repository instructions as its replacement. Ordinary tests, lint, or build
   commands are not replacement review gates. If no review gate is configured,
   state that blocker, do not claim review coverage, and do not commit unless
   the current user explicitly waives the review gate. Do not count merely
   reading the gate or report it as run unless the main session issued that
   shell command for this checkpoint. Rerun immediately if that review command
   reports the worktree changed during review; post-review fix reruns are
   governed by step 3. Report the actual exit/error if blocked.
3. Address actionable review findings, then rerun affected gates. Do not rerun
   the full review for fixes that stay inside the already-reviewed files,
   modules, and risk categories. Rerun the review only if the review command
   reports the worktree changed during review, or if post-review fixes add
   unreviewed files, modules, or risk categories; if that distinction is
   unclear, treat it as unreviewed scope and rerun before commit. If rerun is
   blocked, do not commit and report the residual risk.
4. Commit intended files locally only after gates pass, the review command
   succeeds or the current user explicitly waived an unavailable review gate
   under step 2, and actionable review findings are addressed; if gates or
   review cannot run or pass, report the blocker and do not commit.
5. Push only when explicitly requested.
6. Hand off changed files, orchestration used, gates, review result, commit
   hash if created or the blocker/reason no commit was made, and remaining
   risks.

Review extras are opt-in, not the default. Use no extra review flags for
ordinary implementation, docs-only updates that describe existing behavior, or
single-feature regression tests outside compatibility contracts. Add
`api-compat` only for API, CLI, config/env, schema, generated output, or docs
contracts; OpenAPI/schema rendering, including oneOf, anyOf, discriminator, or
generated API-doc behavior, is `api-compat` work. Add `security`,
`concurrency`, `policy`, `test-gaps`, or custom focused prompts only for those
concrete risks.

