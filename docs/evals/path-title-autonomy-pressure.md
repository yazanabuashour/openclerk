# Path/Title Autonomy Pressure Eval

## Status

Implemented targeted eval lane for the post-POC path/title autonomy decision.
The reduced report is
[`results/ockp-path-title-autonomy-pressure.md`](results/ockp-path-title-autonomy-pressure.md).

This document does not add runner actions, schemas, storage migrations, skill
behavior, public API, or release-blocking production gates. The lane provides
evidence for the follow-up policy decision tracked by `oc-iat`.

`oc-iat` decided to keep current explicit/no-tools behavior and keep this lane
as reference evidence only. The reduced report found no `runner_capability_gap`
failures across the selected path/title pressure scenarios, so no constrained
autonomy policy, runner action, schema, skill behavior, storage migration, or
public interface is promoted from this evidence.

`oc-iat` has been reopened for an ergonomics-gate refresh. Future path/title
pressure should measure both whether current primitives can express the
workflow and whether the current UX is acceptable under natural prompts, step
count, latency, prompt specificity, retry risk, and guidance dependence.

## Purpose

This eval pressure-tests whether current explicit/no-tools behavior remains
sufficient for routine path/title work, or whether evidence justifies a
constrained autonomy policy. It focuses on URL-only inputs, artifact ingestion
without hints, multi-source synthesis, duplicate risk, explicit overrides, and
metadata authority.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Routine execution must not use broad repo search, direct SQLite, direct vault
inspection, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, or ad hoc runtime
programs.

Scenario answers and reduced reports must preserve citations, source refs,
provenance, projection freshness, metadata authority, and repo-relative paths
or neutral placeholders such as `<run-root>`.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario path-title-url-only-autonomy-pressure,path-title-artifact-missing-hints,path-title-multisource-duplicate-pressure,path-title-explicit-overrides-pressure,path-title-duplicate-risk-pressure,path-title-metadata-authority-pressure \
  --report-name ockp-path-title-autonomy-pressure
```

## Scenario Families

- `path-title-url-only-autonomy-pressure`: URL-only documentation request with
  no path/title; measures whether an autonomous source placement can stay
  runner-only and report the selected path/title.
- `path-title-artifact-missing-hints`: source artifact ingestion without
  `source.path_hint` or `source.asset_path_hint`; verifies no-tools
  missing-field handling.
- `path-title-multisource-duplicate-pressure`: seeded sources plus an existing
  synthesis candidate; verifies candidate discovery and duplicate avoidance.
- `path-title-explicit-overrides-pressure`: explicit path/title/type are
  supplied; verifies explicit user instructions override conventions.
- `path-title-duplicate-risk-pressure`: seeded near-duplicate source candidate;
  verifies duplicate risk is surfaced without creating a conflicting document.
- `path-title-metadata-authority-pressure`: ambiguous document-type request;
  verifies frontmatter/metadata authority and projection evidence, not
  filename/path identity.

## Pass/Fail Gates

Failures must be classified as:

- `none`
- `skill_guidance_or_eval_coverage`
- `data_hygiene_or_fixture_gap`
- `eval_contract_violation`
- `runner_capability_gap`

Promotion evidence requires repeated `runner_capability_gap` failures showing
that current document/retrieval workflows cannot express safe constrained
path/title autonomy while preserving explicit instructions, duplicate
avoidance, metadata authority, provenance, freshness, and no-tools validation.

This lane itself does not promote product behavior. The completed `oc-iat`
decision keeps explicit/no-tools behavior, keeps path/title autonomy as
reference evidence, and does not create implementation gates for product or
skill behavior changes.
