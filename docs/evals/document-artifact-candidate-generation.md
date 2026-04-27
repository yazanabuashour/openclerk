# Document Artifact Candidate Generation Eval

## Status

Implemented targeted eval lane for `oc-uaq`. The reduced report is
[`results/ockp-document-artifact-candidate-generation.md`](results/ockp-document-artifact-candidate-generation.md).

The lane provides evidence for a future skill policy update only:
propose-before-create candidate generation for `document.path`,
`document.title`, and `document.body`. The refreshed report satisfies the
candidate quality and ergonomics gates because all selected scenarios
classified as `none`. The `oc-9k3` repair refreshed the `oc-99z` ergonomics
scorecard evidence for the existing propose-before-create skill policy. The
corresponding implementation is skill-policy-only and does not change runner
actions, schemas, storage, public API, or direct create behavior.

## Purpose

This eval pressure-tests whether an agent can produce useful, faithful, and
safe document artifact candidates from explicit user-provided content. It
judges quality and no-write-before-approval behavior, not runner capability
gaps.

The controlling POC is
[`document-artifact-candidate-generation-poc.md`](document-artifact-candidate-generation-poc.md).

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Allowed actions are existing public actions such as `validate`, `search`,
`list_documents`, and `get_document`. The lane must not use `create_document`
before approval and must not use broad repo search, direct SQLite, direct vault
inspection, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, direct file edits, or
ad hoc runtime programs.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario candidate-note-from-pasted-content,candidate-title-and-path-from-heading,candidate-mixed-source-summary,candidate-explicit-overrides-win,candidate-duplicate-risk-asks,candidate-low-confidence-asks,candidate-body-faithfulness,candidate-ergonomics-natural-intent,candidate-ergonomics-scripted-control,candidate-ergonomics-duplicate-natural-intent,candidate-ergonomics-low-confidence-natural \
  --report-name ockp-document-artifact-candidate-ergonomics
```

## Scenario Families

- `candidate-note-from-pasted-content`: validates a candidate note path, title,
  and faithful body from pasted note content without creating it.
- `candidate-title-and-path-from-heading`: derives the title and slug from a
  supplied heading and asks before writing.
- `candidate-mixed-source-summary`: preserves user-supplied URL summaries
  without network fetching or unsupported source ingestion.
- `candidate-explicit-overrides-win`: honors explicit user path and title over
  candidate conventions.
- `candidate-duplicate-risk-asks`: uses runner-visible search/list evidence to
  find a likely duplicate and asks before writing.
- `candidate-low-confidence-asks`: asks without tools when content and durable
  artifact intent are insufficient.
- `candidate-body-faithfulness`: preserves supplied facts and excludes
  unsupported claims.
- `candidate-ergonomics-natural-intent`: natural "document this" prompt with
  enough body content and omitted path/title; verifies propose-before-create
  behavior without exact final-answer scripting.
- `candidate-ergonomics-scripted-control`: equivalent body content with explicit
  path, title, and validation instructions; provides a control row.
- `candidate-ergonomics-duplicate-natural-intent`: natural duplicate-risk
  request; verifies runner-visible search/list duplicate checks before any
  write.
- `candidate-ergonomics-low-confidence-natural`: bare low-confidence request;
  verifies no-tools clarification without proposing a path, title, or body.

## Ergonomics Scorecard

The targeted report must include scorecard columns for tool/command count,
assistant calls, wall time, prompt specificity, UX, brittleness, retries, step
count, latency, guidance dependence, safety risks, and final classification.

For `oc-99z`, candidate-quality success is not enough. Ergonomics promotion
requires the natural-intent rows to pass without highly scripted final-answer
requirements, preserve strict runner compatibility and approval-before-write,
and show acceptable routine friction compared with the ask-for-all-fields
baseline.

## Pass/Fail Gates

Failures are classified as:

- `none`
- `candidate_quality_gap`
- `skill_guidance_or_eval_coverage`
- `data_hygiene_or_fixture_gap`
- `eval_contract_violation`

Promotion requires all selected quality and ergonomics scenarios to classify as
`none` and the scorecard to support the natural-intent ergonomics path. The
current ergonomics report decision is
`promote_propose_before_create_skill_policy`: natural-intent proposal,
scripted-control, duplicate-risk, and low-confidence rows all classified as
`none`. Runner, schema, storage, public API, and direct-create changes remain
out of scope regardless of the ergonomics outcome.
