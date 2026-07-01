# Codebase Decisions Knowledge Pack

This pack shows how an agent asks for cited task context before changing a
subsystem with prior decisions.

Contents:

- `vault/sources/` has subsystem notes.
- `vault/decisions/` has one accepted decision note.
- `commands.md` shows the inspect and context-pack flow.

Expected outcome: `openclerk inspect` reports a bound vault, and
`openclerk clerk context_pack` returns the auth callback source and decision as
must-read context before implementation work.
