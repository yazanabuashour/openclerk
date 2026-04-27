- For all committed docs, reports, and artifact references, use repo-relative paths or neutral repo-relative placeholders. Never use machine-absolute filesystem paths.
- Do work on the current branch. Do not create or switch to another branch unless explicitly instructed.
- For repo-pinned developer tools declared in `mise.toml`, run commands through `mise exec -- ...` so agents use the same tool versions as local docs and CI.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for maintainer issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- If you are acting as a maintainer or local coding agent, use `bd` for task tracking instead of ad hoc markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete steps 1-5 below, then stop for manual review before running `git add`, `git commit`, `bd dolt push`, or `git push`. The workflow is paused for manual review at step 5 with uncommitted local changes, and the work session is NOT complete until steps 6-10 are finished after review approval and `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Prepare manual review** - Run `git status`, summarize changed files and quality gates, confirm no commit or push has been performed, and leave files uncommitted for manual review
5. **Manual review** - Stop here by default with uncommitted local changes, report that the workflow is paused for manual review, and wait for explicit instruction to complete the remaining steps
6. **Commit approved changes** - After explicit review approval, stage the intended files and create a local commit
7. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
8. **Clean up** - Clear stashes, prune remote branches
9. **Verify** - All changes committed AND pushed
10. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- The workflow pauses for manual review after step 5 with uncommitted local changes, and the work session is NOT complete until `git push` succeeds
- Do NOT run `git add`, `git commit`, `bd dolt push`, or `git push` before manual review unless explicitly instructed
- Do NOT continue past `Manual review` unless explicitly instructed to complete the remaining workflow steps
- Once instructed to continue after review, stage, commit, pull/rebase, run `bd dolt push`, and `git push`; do NOT stop again with local-only changes
- NEVER say "ready to push when you are" after review approval - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
