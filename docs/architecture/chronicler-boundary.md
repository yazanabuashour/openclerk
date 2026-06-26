# Chronicler Lite Boundary

## Status

Accepted scope correction for `openclerk clerk`.

Chronicler Lite remains a concrete OpenClerk capability. Ambitious
autonomous, dreaming, always-on Chronicler is shelved as product scope and kept
only as reference pressure for later candidate comparison.

## Position

OpenClerk Core remains the trusted local-first knowledge-plane runtime.
Chronicler Lite is a first-party optional orchestration layer over Core, not a
new authority system and not a second store.

The sharper stack framing is:

| Surface | Job |
| --- | --- |
| AgentSpace / Workcell | Where the work happens. |
| OpenClerk | What the agent should know before work starts. |
| Chronicler Lite | What gets recorded after work happens. |
| Sentinel | What operational evidence explains why work is needed. |

Chronicler Lite's job is narrow: turn a completed workspace session, inbox
note, or handoff artifact into durable repo-knowledge candidates that can be
reviewed and approved through OpenClerk document lifecycle APIs. It does not
decide on its own what the repo should know, rewrite documentation in the
background, or maintain an open-ended memory system.

The Lite surface ships in the existing `openclerk` binary under:

```bash
openclerk clerk run --once
openclerk clerk inbox_scan
openclerk clerk context_pack
```

`run --once` emits the combined planning report:

```json
{
  "schema_version": "openclerk-clerk.v1",
  "action": "clerk_run",
  "result": {
    "mode": "once",
    "planned_no_write": true,
    "writes_performed": 0,
    "inbox_candidates": [],
    "context_packs": [],
    "stale_synthesis": [],
    "duplicate_risks": [],
    "pending_review": [],
    "blockers": []
  }
}
```

`inbox_scan` emits the inbox-only report:

```json
{
  "schema_version": "openclerk-clerk.v1",
  "action": "inbox_scan",
  "result": {
    "mode": "inbox_scan",
    "planned_no_write": true,
    "writes_performed": 0,
    "inbox_candidates": [],
    "context_packs": [],
    "duplicate_risks": [],
    "pending_review": [],
    "blockers": []
  }
}
```

`context_pack` emits the context-only report:

```json
{
  "schema_version": "openclerk-clerk.v1",
  "action": "context_pack",
  "result": {
    "mode": "context_pack",
    "planned_no_write": true,
    "writes_performed": 0,
    "inbox_candidates": [],
    "context_packs": [],
    "stale_synthesis": [],
    "blockers": []
  }
}
```

Additional fields may provide authority limits, approval boundaries, and
deferred capability labels, but Chronicler Lite always reports
`planned_no_write: true` and `writes_performed: 0`.

When a planning path would inspect Core evidence, existing OpenClerk storage is
required. Chronicler Lite returns a blocker rather than initializing SQLite
from a read-only command.

## Core Authority

Core owns canonical knowledge behavior:

- canonical markdown remains the human-readable authority
- citations, provenance, and projection freshness remain Core evidence
- indexes and projections remain derived recall layers
- approved document writes remain under existing `openclerk document`
  lifecycle APIs
- retrieval and report evidence remains under existing `openclerk retrieval`
  APIs

Chronicler Lite may compose Core read-only actions. It must not directly edit
canonical markdown, directly mutate SQLite, invent hidden memory, or route
around the installed runner/service boundary.

## Lite Behavior

`openclerk clerk run --once` performs one combined read-only planning pass.
`openclerk clerk inbox_scan` runs only the inbox-candidate part, and
`openclerk clerk context_pack` runs only the task-context part.

Supported inputs:

- `--inbox-path <path>` for an explicit local markdown/text workspace-session
  note, handoff, or inbox file
- `--inbox-path <path>` for an explicit local directory, scanned
  non-recursively for markdown/text files only
- `--task <text>` for a task context pack
- `--query <text>` to override the context-pack retrieval query
- `--path-prefix <prefix>` to narrow context-pack retrieval
- `--limit <n>` to cap planner and retrieval results

Inbox planning treats local files as candidate evidence, not canonical truth.
It reuses Core `artifact_candidate_plan` behavior and returns proposed
title/path/type/tags/summary/source refs/duplicate risk/recommended action.
It does not create canonical markdown, silently ingest files, recursively scan
directories, or write to the vault.

Context packs reuse existing retrieval behavior. They return compact
must-read documents, relevant decisions where runner-visible evidence exists,
stale-or-missing context notes, open questions, and citations. Path-prefixed
context packs keep decision citations inside the requested prefix. They are
supporting task context only; source-sensitive answers still depend on Core
citations, provenance, projection freshness, and authority limits.

## Shelved Scope

The following shapes are explicitly shelved for now:

- autonomous dreaming or self-directed improvement loops
- always-on daemon, watch, cron, or background documentation repair
- agent-decided documentation rewrites
- broad repo-wide knowledge gardening
- open-ended memory evolution
- complex knowledge graph or semantic ontology of the codebase
- multi-agent historian
- autonomous researcher
- repo wiki maintainer as an independent product
- runbook updater or architecture-documentation bot without an explicit
  session artifact and approval path
- direct durable writes from Chronicler

These are interesting reference pressures, but they lack a crisp early demo and
would blur read/fetch/inspect permission with durable-write approval.

## Decision Review

Safety pass: Chronicler Lite preserves runner-only access, no direct SQLite,
no direct canonical markdown mutation, no hidden memory, no autonomous routing,
no background worker, and approval-before-write through existing document
lifecycle APIs.

Capability pass: Current Lite commands can serve the workspace project by
planning repo-knowledge candidates from explicit completed-session notes and
by packaging task context before follow-up work. They do not yet complete the
full "session to committed repo knowledge" loop because durable writes remain
approval-gated and outside `openclerk clerk`.

UX quality: The narrower job is understandable to a normal user: give
Chronicler Lite the completed session or handoff, get reviewed repo-knowledge
candidates and context back. The ambitious product remains too ambiguous
because it could mean agent memory, repo wiki maintenance, runbook updates,
autonomous research, incident history, architecture documentation, session
summaries, or knowledge graph construction.

Decision: keep Chronicler Lite as the shipped concrete capability; shelve
autonomous/dreaming/always-on Chronicler as a product track; preserve the
underlying post-work recording idea through follow-up candidate comparison.

## Follow-Up

The underlying need remains valid: workcell sessions should leave durable,
reviewable repo knowledge when they create decisions, runbooks, architecture
context, incident history, or reusable handoff material.

`oc-dcy2` should compare post-Lite Chronicler candidate surfaces before any
promotion:

| Candidate | Safety | Capability | UX quality |
| --- | --- | --- | --- |
| Current `clerk run --once` plus explicit session notes | Strongest current boundary: read-only planning, no writes, no daemon. | Can propose document candidates and context from explicit artifacts. | Simple enough, but may still require manual prompt choreography and handoff discipline. |
| Dedicated `session_record_report` under `openclerk clerk` | Keeps runner-only read/fetch/inspect and planned-no-write posture. | Could package session summary, candidate doc updates, stale docs, open decisions, and next approved requests in one report. | Best near-term candidate if evidence shows the current two-command shape is too ceremonial. |
| Approval-gated review queue over document lifecycle APIs | Preserves durable-write approval, duplicate checks, provenance, and audit behavior. | Could complete session-to-repo-knowledge handoff after review without granting Chronicler autonomous authority. | Useful later, but heavier than the first concrete Lite demo. |

The follow-up must choose the best candidate, combine useful behavior if
appropriate, defer or kill the track, or record `none viable yet`.
