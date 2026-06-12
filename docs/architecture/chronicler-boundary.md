# Chronicler Boundary

## Status

Initial MVP boundary for `openclerk clerk`.

## Position

OpenClerk Core remains the trusted local-first knowledge-plane runtime.
Chronicler is a first-party optional orchestration layer over Core, not a new
authority system and not a second store.

The MVP ships in the existing `openclerk` binary under:

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
deferred capability labels, but the MVP always reports `planned_no_write: true`
and `writes_performed: 0`.

When a planning path would inspect Core evidence, existing OpenClerk storage is
required. Chronicler returns a blocker rather than initializing SQLite from a
read-only command.

## Core Authority

Core owns canonical knowledge behavior:

- canonical markdown remains the human-readable authority
- citations, provenance, and projection freshness remain Core evidence
- indexes and projections remain derived recall layers
- approved document writes remain under existing `openclerk document`
  lifecycle APIs
- retrieval and report evidence remains under existing `openclerk retrieval`
  APIs

Chronicler may compose Core read-only actions. It must not directly edit
canonical markdown, directly mutate SQLite, invent hidden memory, or route
around the installed runner/service boundary.

## MVP Behavior

`openclerk clerk run --once` performs one combined read-only planning pass.
`openclerk clerk inbox_scan` runs only the inbox-candidate part, and
`openclerk clerk context_pack` runs only the task-context part.

Supported inputs:

- `--inbox-path <path>` for an explicit local markdown/text file
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

## Non-Goals

The MVP does not implement:

- daemon or watch mode
- review approval queues
- auto-filing or autonomous durable writes
- autonomous browsing or recursive crawling
- autonomous routing
- broad vector memory or hidden memory
- direct SQLite access from Chronicler
- direct canonical markdown mutation from Chronicler

Future write or review-queue behavior must go through approved document
lifecycle APIs that preserve provenance, duplicate checks, rollback/audit
behavior, and review policy.
