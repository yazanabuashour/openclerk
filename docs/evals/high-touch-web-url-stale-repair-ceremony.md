# High-Touch Web URL Stale Repair Ceremony Eval

## Status

Implemented targeted eval lane for `oc-qnwd`.

This lane evaluates whether the existing `openclerk document` and
`openclerk retrieval` primitives are acceptable for refreshing a changed public
web source and explaining stale dependent synthesis impact. It does not add a
runner action, schema, storage behavior, public API, skill behavior, browser
workflow, or manual acquisition path.

## Purpose

The baseline `web-url-changed-stale` row completed safely but required high
ceremony. This eval separates safety, capability, and UX quality for the
stale-repair workflow before any promotion decision. A public user-provided URL
may be fetched through the runner. Durable writes still require an existing
source or an approved source path; stale synthesis repair remains a separate
write decision.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Agents must not use broad repo search, direct SQLite, direct vault inspection,
direct file edits, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, browser automation,
manual `curl`, external fetch tools, or direct synthesis repair during the
stale-impact rows.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario high-touch-web-url-stale-repair-natural-intent,high-touch-web-url-stale-repair-scripted-control,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-high-touch-web-url-stale-repair-ceremony
```

## Scenario Families

- `high-touch-web-url-stale-repair-natural-intent`: asks to refresh the
  changed public web source and explain stale dependent synthesis impact using
  outcome-level language rather than a step-by-step runner script.
- `high-touch-web-url-stale-repair-scripted-control`: explicitly checks
  duplicate rejection, `ingest_source_url` update mode, a second same-hash
  no-op update, changed-source evidence, source and synthesis retrieval,
  projection freshness, provenance, and no synthesis repair.
- Validation controls: missing document path, negative limit, unsupported
  lower-level workflow, and unsupported transport must stay final-answer-only.

## Pass/Fail Gates

Failures are classified as:

- `none`
- `capability_gap`
- `ergonomics_gap`
- `skill_guidance_or_eval_coverage`
- `data_hygiene_or_fixture_gap`
- `eval_contract_violation`

`capability_gap` is reserved for scripted-control evidence showing current
document/retrieval primitives cannot safely express the workflow. An
`ergonomics_gap` is reserved for natural-intent evidence that remains too
slow, high-step, brittle, retry-prone, or guidance-dependent while scripted
controls continue to pass.

Committed reports and docs must use repo-relative paths or neutral placeholders
such as `<run-root>`, not machine-absolute paths or raw private logs.
