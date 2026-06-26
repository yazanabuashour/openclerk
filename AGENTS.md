- For all committed docs, reports, and artifact references, use repo-relative paths or neutral repo-relative placeholders. Never use machine-absolute filesystem paths.
- Do work on the current branch. Do not create or switch to another branch unless explicitly instructed.
- For repo-pinned developer tools declared in `mise.toml`, run commands through `mise exec -- ...` so agents use the same tool versions as local docs and CI.

## ADR/POC/Eval Decision Taste Review

When doing OpenClerk ADR, POC, eval, promotion, or deferred-capability decision work, keep the existing evidence discipline but add a taste check before accepting a defer/reference outcome:

- Ask whether a normal user would expect a simpler OpenClerk surface than the one being preserved.
- Distinguish read/fetch/inspect permission from durable-write approval. A public user-provided URL can be enough to fetch through the runner; approval belongs at durable writes, privileged access, purchases/actions, or other irreversible changes.
- Prefer extending the natural existing runner action when the input clearly belongs there, instead of declaring the adjacent UX unsupported.
- Treat "completed but ceremonial" eval passes as possible taste debt when they require high step count, long latency, exact prompt choreography, or surprising clarification turns.
- Record safety pass, capability pass, and UX quality separately when a report or decision needs to justify defer/reference.
- When taste debt, defer, keep-as-reference, or another non-promotion outcome still leaves a real capability, ergonomics, safety, auditability, or workflow need, identify whether the evaluated shape failed while the need remains valid. If it does, create or propose follow-up work for candidate-surface comparison before handoff, normally with 2-3 plausible shapes unless the decision documents why only one is viable. The follow-up must compare candidates, choose the best, combine useful behaviors if appropriate, defer or kill the track, or record `none viable yet`.
- Before closing any ADR, POC, eval, promotion, or deferred-capability decision epic with outcome `keep-as-reference`, `defer`, `more evidence`, `candidate selected`, `none viable yet`, or another non-promotion result, check existing follow-up work in GitHub issues or public docs. If none exists, create or propose the follow-up work before closing the parent or handing off.
- Do not use taste review to bypass safety or evidence discipline: authority, citations, provenance, freshness, local-first behavior, duplicate handling, runner-only access, approval-before-write, and ADR/POC/eval/promotion decisions still apply.

## Work Item Completion

A **work item** is one logical task, story, or other coherent unit of work. **When completing each work item**, complete the workflow below through review, local commit, verification, and handoff before starting unrelated work or handing off. Push only when the maintainer or task explicitly asks for remote publication. If a single thread completes multiple independent tasks or stories, repeat this workflow once for each completed work item. If a work item contains multiple independent logical checkpoints, complete the review workflow for each checkpoint before moving to the next; a checkpoint is the smallest coherent unit whose changes can be reviewed on its own, such as one bug fix, one finding, one migration step, or one separable behavior change.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close or update the relevant public issue or project item when one exists
4. **Prepare review** - Run `git status`, summarize changed files and quality gates, and confirm no commit or push has been performed
5. **Codex review** - Run the review command once for the current work item or review checkpoint:
   ```bash
   codex --search -m gpt-5.5 -c 'model_reasoning_effort="high"' review --uncommitted
   ```
   If the review finds issues, address the findings.
6. **Commit reviewed changes** - After the review command completes, stage the intended files and create a local commit
7. **Remote publication** - Push only when explicitly requested:
   ```bash
   git pull --rebase
   git push
   git status
   ```
8. **Clean up** - Clear stashes, prune remote branches when relevant
9. **Verify** - All intended changes are committed, and pushed only when remote publication was requested
10. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Do not push to a remote unless the maintainer or task explicitly requested remote publication
- Run the Codex review command once per work item or review checkpoint; do not rerun the same checkpoint review as a workflow loop
- For multi-checkpoint work, run quality gates, Codex review, and commit after each independent checkpoint before starting the next
- Do NOT commit before quality gates and the Codex review command are complete
- After the review command completes, stage and commit the intended files; if remote publication was requested, pull/rebase, then push
- If a requested push fails, resolve and retry until it succeeds or report the blocker
