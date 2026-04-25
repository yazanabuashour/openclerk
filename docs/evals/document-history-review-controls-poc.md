# Document History And Review Controls POC

## Status

Planned targeted POC/eval contract for post-v0.1.0 document lifecycle
evidence.

This document does not add fixture data, runner actions, schemas, storage
migrations, public API, or release-blocking production gates.

## Purpose

This POC should determine whether OpenClerk needs semantic document history and
review controls beyond the current v1 AgentOps document and retrieval surface.
The goal is to test real lifecycle pressure from agent-authored durable edits
without promoting a new runner surface by default.

The POC follows the v1 pattern: start with current `openclerk document` and
`openclerk retrieval` workflows, add targeted pressure prompts only where the
current workflow may be structurally insufficient, classify failures, and end
with a promote, defer, kill, or keep-as-reference decision.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Routine execution must not use broad repo search, direct SQLite, direct vault
inspection, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, or ad hoc runtime
programs.

Scenario answers and reduced reports must preserve source refs, citations,
provenance, projection freshness, and repo-relative paths or neutral
placeholders such as `<run-root>`. Public artifacts must not include raw
private document diffs.

## Scenario Families

- **History inspection control:** use existing document retrieval,
  provenance-events, and projection-states workflows to explain recent
  OpenClerk-managed edits for a registered document before assuming a new
  history action is needed.
- **Diff review pressure:** compare current content with prior
  runner-visible references or evidence summaries while preserving citations
  and avoiding raw private diff leakage in committed reports.
- **Restore and rollback pressure:** evaluate whether an unsafe
  OpenClerk-authored edit can be identified, explained, and restored or
  prepared for restoration with explicit evidence and operator-visible state.
- **Pending-change review pressure:** evaluate whether agent-authored changes
  that should not become accepted knowledge immediately can be surfaced for
  human review without hidden autonomous rewrites.
- **Stale synthesis after revision:** verify that a canonical document change
  exposes stale derived synthesis or projections through provenance and
  projection freshness before any repair or accepted rollback.
- **Bypass and validation pressure:** reject direct SQLite, direct vault,
  HTTP/MCP, source-built runner, unsupported transport, invalid-limit, and
  lower-level workflow requests under the existing no-tools and
  final-answer-only policy.

## Pass/Fail Gates

The POC passes as evidence for promotion only when repeated targeted failures
show that existing document, retrieval, provenance, and projection freshness
workflows are structurally insufficient for semantic document lifecycle
control.

Failures must be classified as:

- data hygiene
- skill guidance
- eval coverage
- runner capability gap

Promotion is not justified by awkward but successful multi-step workflows,
missing instructions, missing fixture data, or evaluator pressure that bypasses
the AgentOps contract.

The POC fails or kills the candidate if the proposed behavior duplicates Git or
sync history, weakens canonical markdown authority, drops source refs or
citations, hides provenance or freshness, creates hidden autonomous rewrites,
or requires direct SQLite, direct vault inspection, HTTP/MCP, source-built
runner paths, backend variants, module-cache inspection, or ad hoc runtime
programs.

## Expected Decision Output

A completed targeted report should record:

- the selected scenario set and control prompts
- which runner-visible evidence was used
- whether failures were capability gaps or non-product gaps
- privacy handling for raw diffs and document bodies
- the decision: promote, defer, kill, or keep as reference
- the exact follow-up implementation surface only if promotion is justified

Until that report exists, document history and review controls remain deferred.

