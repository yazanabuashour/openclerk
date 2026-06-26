# Real-Vault Routine UX Telemetry

## Purpose

This lane measures routine AgentOps UX against a maintainer-supplied private
vault while preserving the real-vault evidence boundary. It is telemetry only:
it does not promote a runner action, schema, storage migration, skill behavior,
retrieval backend, or release gate.

Use this lane after the reduced real-vault maturity report when the next
question is agent ergonomics rather than runtime timing. The committed output is
a sanitized Markdown summary only.

## Inputs

The maintainer supplies the private vault at run time:

```bash
mise exec -- go run ./scripts/agent-eval/ockp routine-ux real-vault \
  --vault-root "<private-vault>" \
  --task-manifest "<private-task-manifest>" \
  --run-root "<run-root>" \
  --report-name ockp-real-vault-routine-ux
```

`<private-vault>` may point at the maintainer's real notes vault. Do not commit
that path, an expanded home path, the private task manifest, raw JSON, event
logs, disposable vault copies, SQLite files, private prompts, private paths,
titles, snippets, document ids, chunk ids, or raw runner output.

## Private Task Manifest

The task manifest is local-only JSON with schema
`openclerk-real-vault-routine-ux-tasks.v1`. It must contain exactly one task for
each class:

- `source_discovery`
- `cited_search_answer`
- `synthesis_create_update`
- `provenance_freshness`
- `decision_record_lookup`
- `stale_duplicate_detection`

Each task has a private natural-language prompt and optional expected or
forbidden runner action classes. The harness never writes the prompt text to
the committed Markdown report.

Example shape, with placeholder prompts only:

```json
{
  "schema_version": "openclerk-real-vault-routine-ux-tasks.v1",
  "tasks": [
    {
      "class": "source_discovery",
      "prompt": "<private prompt>",
      "expected_runner_actions": ["list_documents", "get_document", "search"]
    },
    {
      "class": "cited_search_answer",
      "prompt": "<private prompt>",
      "expected_runner_actions": ["search"]
    },
    {
      "class": "synthesis_create_update",
      "prompt": "<private prompt>",
      "expected_runner_actions": ["create_document", "replace_section"]
    },
    {
      "class": "provenance_freshness",
      "prompt": "<private prompt>",
      "expected_runner_actions": ["provenance_events", "projection_states"]
    },
    {
      "class": "decision_record_lookup",
      "prompt": "<private prompt>",
      "expected_runner_actions": ["decisions_lookup", "decision_record"]
    },
    {
      "class": "stale_duplicate_detection",
      "prompt": "<private prompt>",
      "expected_runner_actions": ["search", "projection_states"]
    }
  ]
}
```

## Safety Model

The harness copies the private vault into `<run-root>` and points an isolated
OpenClerk database at that disposable copy. Write-like rows can create or update
validation content in the copy, but the live private vault is not the mutation
target.

Each Codex row is wrapped with routine AgentOps boundaries:

- use installed `openclerk document` and `openclerk retrieval`
- no direct SQLite
- no direct vault inspection
- no broad repo search
- no source-built runner paths
- no HTTP/MCP bypasses
- no browser automation
- no unsupported transports
- no module-cache inspection

## Committed Report

The committed report is `docs/evals/results/ockp-real-vault-routine-ux.md`.
It records only aggregate row metrics:

- task reference such as `private-task-1`
- task class
- status and failure classification
- tool calls, command executions, assistant calls, wall time, retries, and
  final-answer repair turns
- runner action classes used
- safety pass, capability pass, UX quality, and safety risks

Raw JSON is written only under `<run-root>`. Raw event logs remain under
`<run-root>`. Do not commit either.

## Decision Use

This lane can identify taste debt and justify follow-up work for candidate
comparison. It cannot by itself promote a public OpenClerk surface. Any future
implementation still needs targeted candidate evidence and a decision that names
the exact surface, compatibility expectations, failure modes, and safety gates.
