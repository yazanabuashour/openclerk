# Document-This Intake Pressure Eval

## Status

Implemented targeted eval lane for `oc-u9l`. The reduced report is
[`results/ockp-document-this-intake-pressure.md`](results/ockp-document-this-intake-pressure.md).

This document does not add runner actions, schemas, storage migrations, skill
behavior, public API, or release-blocking production gates. The lane provides
evidence for the promotion decision tracked by `oc-99z`.

`oc-99z` kept current strict behavior and did not promote autonomous
autofiling, path/title/body inference, runner changes, schema changes, storage
changes, or skill behavior changes. The selected scenarios in the reduced
report are classified `none`.

`oc-99z` has been reopened for an ergonomics-gate refresh. The refreshed
decision should reconcile this strict-intake evidence with the later
propose-before-create candidate-generation policy and should report both
technical expressibility and routine UX acceptability.

## Purpose

This eval pressure-tests user prompts like "document this" across article,
docs page, paper, transcript, mixed-source, explicit override, duplicate
candidate, and synthesis-freshness cases. It verifies that the OpenClerk skill
can preserve strict runner JSON without relaxing binary validation.

The controlling POC is
[`agent-side-knowledge-intake-workflow-options-poc.md`](agent-side-knowledge-intake-workflow-options-poc.md).

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Routine execution must not use broad repo search, direct SQLite, direct vault
inspection, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, direct file edits, or ad
hoc runtime programs.

Scenario answers and reduced reports must preserve citations, source refs,
provenance, projection freshness, metadata authority, and repo-relative paths
or neutral placeholders such as `<run-root>`.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario document-this-missing-fields,document-this-explicit-create,document-this-source-url-missing-hints,document-this-explicit-overrides,document-this-duplicate-candidate,document-this-existing-update,document-this-synthesis-freshness \
  --report-name ockp-document-this-intake-pressure
```

## Scenario Families

- `document-this-missing-fields`: ambiguous mixed-source "document this"
  request missing `document.path`, `document.title`, and `document.body`;
  verifies one no-tools clarification.
- `document-this-explicit-create`: explicit path, title, and body are supplied;
  verifies strict `create_document` JSON and no source-path autofiling.
- `document-this-source-url-missing-hints`: source URL without
  `source.path_hint` or `source.asset_path_hint`; verifies one no-tools
  clarification.
- `document-this-explicit-overrides`: mixed URLs with explicit path/title;
  verifies explicit user instructions win over inferred source placement.
- `document-this-duplicate-candidate`: seeded source candidate for an article;
  verifies runner-visible duplicate search/list and no duplicate create.
- `document-this-existing-update`: seeded target and decoy notes; verifies
  runner-visible candidate lookup, target inspection, and updating only the
  intended target.
- `document-this-synthesis-freshness`: article, docs page, paper, and
  transcript sources plus existing synthesis; verifies search, synthesis
  candidate lookup, source refs, projection freshness, provenance inspection,
  and no duplicate synthesis.

## Pass/Fail Gates

Failures are classified as:

- `none`
- `skill_guidance_or_eval_coverage`
- `data_hygiene_or_fixture_gap`
- `eval_contract_violation`
- `runner_capability_gap`

Promotion evidence requires repeated `runner_capability_gap` failures showing
that existing document/retrieval workflows cannot express document-this intake
while preserving strict validation, duplicate avoidance, metadata authority,
citations or source refs, provenance, and freshness.

This lane itself does not promote product behavior. `oc-99z` decided to keep
the candidate behavior as current behavior and reference evidence, with no
promoted runner or skill changes.
