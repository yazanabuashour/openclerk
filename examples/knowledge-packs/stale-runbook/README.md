# Stale Runbook Knowledge Pack

This pack demonstrates stale synthesis detection without automatic repair.

Contents:

- `vault/sources/deploy-source.md` is the current source note.
- `vault/synthesis/deploy-runbook.md` is a durable runbook that cites the
  source but contains older guidance.
- `commands.md` shows the read-only inspection and repair-request path.

Expected outcome: projection freshness can report the runbook as stale after
the source is newer than the synthesis. The follow-up is an approved
`compile_synthesis` request, not automatic repair.
