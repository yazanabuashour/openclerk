- For all committed docs, reports, and artifact references, use repo-relative paths or neutral repo-relative placeholders. Never use machine-absolute filesystem paths.
- Do work on the current branch. Do not create or switch to another branch unless explicitly instructed.

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

**When ending a work session**, you MUST complete steps 1-4 below, then stop for manual review. The workflow is paused for manual review at step 4 and the work session is NOT complete until steps 5-8 are finished and `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Manual review** - Stop here by default, report that the workflow is paused for manual review, and wait for explicit instruction to complete the remaining steps
5. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
6. **Clean up** - Clear stashes, prune remote branches
7. **Verify** - All changes committed AND pushed
8. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- The workflow pauses for manual review after step 4 and the work session is NOT complete until `git push` succeeds
- Do NOT continue past `Manual review` unless explicitly instructed to complete the remaining workflow steps
- Once instructed to continue after review, complete the remaining workflow and push; do NOT stop again with local-only changes
- NEVER say "ready to push when you are" after review approval - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
