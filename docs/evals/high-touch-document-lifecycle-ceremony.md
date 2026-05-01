# High-Touch Document Lifecycle Ceremony Eval

## Status

Implemented targeted eval lane for `oc-k8ba`.

This lane evaluates whether the existing `openclerk document` and
`openclerk retrieval` primitives remain acceptable for lifecycle review and
rollback of an unsafe accepted summary under natural lifecycle intent. It does
not add a runner action, schema, storage behavior, public API, skill behavior,
or product implementation.

## Purpose

The baseline `document-lifecycle-natural-intent` row completed safely but
required high ceremony. This eval separates safety, capability, and UX quality
for lifecycle rollback before any promotion decision. Durable lifecycle writes
must preserve canonical authority, source refs or citations, provenance,
projection freshness, rollback target accuracy, privacy-safe summaries,
local-first runner-only access, and no-bypass controls.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Agents must not use broad repo search, direct SQLite, direct vault inspection,
direct file edits, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, raw private diffs, or
storage-root paths in final answers or committed artifacts.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario high-touch-document-lifecycle-natural-intent,high-touch-document-lifecycle-scripted-control,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-high-touch-document-lifecycle-ceremony
```

## Scenario Families

- `high-touch-document-lifecycle-natural-intent`: asks for lifecycle review and
  rollback of an unsafe accepted summary using outcome-level language rather
  than a step-by-step runner script.
- `high-touch-document-lifecycle-scripted-control`: explicitly searches
  restore authority evidence, lists the lifecycle target, retrieves the target
  before editing, restores the exact accepted summary with `replace_section`,
  and inspects provenance plus projection freshness.
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
document/retrieval primitives cannot safely express lifecycle rollback. An
`ergonomics_gap` is reserved for natural-intent evidence that remains too
slow, high-step, brittle, retry-prone, or guidance-dependent while scripted
controls continue to pass.

Committed reports and docs must use repo-relative paths or neutral placeholders
such as `<run-root>`, not machine-absolute paths, raw private logs, or raw
private diffs.
