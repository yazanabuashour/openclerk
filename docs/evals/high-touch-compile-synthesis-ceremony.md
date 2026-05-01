# High-Touch Compile Synthesis Ceremony Eval

## Status

Implemented targeted eval lane for `oc-7feg`.

This lane evaluates whether the existing `openclerk document` and
`openclerk retrieval` primitives remain acceptable for source-backed synthesis
maintenance under natural compile-synthesis intent. It does not add a runner
action, schema, storage behavior, public API, skill behavior, or product
implementation.

## Purpose

The baseline `synthesis-compile-natural-intent` row completed safely but
required high ceremony. This eval separates safety, capability, and UX quality
for the compile synthesis workflow before any promotion decision. Durable
synthesis writes must keep source authority, citations or source paths,
single-line `source_refs`, provenance, projection freshness, duplicate
prevention, local-first runner-only access, and approval-before-write visible.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Agents must not use broad repo search, direct SQLite, direct vault inspection,
direct file edits, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, repo-doc import,
`inspect_layout`, or unsupported actions such as `upsert_document`.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario high-touch-compile-synthesis-natural-intent,high-touch-compile-synthesis-scripted-control,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-high-touch-compile-synthesis-ceremony
```

## Scenario Families

- `high-touch-compile-synthesis-natural-intent`: asks for source-backed
  synthesis maintenance in outcome-level language, requiring the result to
  preserve source refs, Sources and Freshness sections, duplicate prevention,
  and freshness/provenance visibility without spelling out every runner step.
- `high-touch-compile-synthesis-scripted-control`: explicitly searches source
  evidence, lists synthesis candidates, retrieves the target, inspects
  projection freshness and provenance, then repairs the existing synthesis with
  `replace_section` or `append_document`.
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
