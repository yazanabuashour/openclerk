---
decision_id: adr-agent-chosen-document-artifact-candidate-generation
decision_title: Agent-Chosen Document Artifact Candidate Generation
decision_status: accepted
decision_scope: agent-interaction-policy
decision_owner: platform
---
# ADR: Agent-Chosen Document Artifact Candidate Generation

## Status

Accepted as an evidence-backed agent-side propose-before-create skill policy.
The policy is implemented in `skills/openclerk/SKILL.md`.

This ADR supersedes the `oc-99z` framing only for the product question of
whether an agent may choose a candidate `document.path`, `document.title`, and
`document.body` from explicit user-provided content. The `oc-99z` decision
remains valid for runner, schema, storage, migration, public API, and direct
create behavior: none are promoted or changed by this ADR.

For the amended `oc-99z` ergonomics path, this ADR identifies the only
candidate surface reviewed: propose-before-create skill behavior, with no
durable write before user approval. The candidate-generation evidence proves
quality and safety boundaries, but the refreshed ergonomics scorecard deferred
`oc-99z` promotion pending natural-intent repair.

Supporting evidence:

- [`../evals/document-artifact-candidate-generation-poc.md`](../evals/document-artifact-candidate-generation-poc.md)
- [`../evals/document-artifact-candidate-generation.md`](../evals/document-artifact-candidate-generation.md)
- [`../evals/results/ockp-document-artifact-candidate-generation.md`](../evals/results/ockp-document-artifact-candidate-generation.md)

## Context

The prior `oc-a8r`, `oc-tw5`, `oc-u9l`, and `oc-99z` track asked whether
existing `openclerk document` and `openclerk retrieval` runner workflows could
handle document-this intake pressure while keeping runner validation strict.
That track correctly found no runner capability gap, but it did not decide the
more useful product question: whether the agent may propose a high-quality
artifact candidate when the user supplied enough content but did not supply
the durable path, title, and final body.

This ADR evaluates that convenience behavior directly. The promotion target is
not autonomous write. It is candidate generation before write: the agent chooses
a candidate path, title, and body; validates or checks it through existing
runner actions when appropriate; reports the candidate; and asks for approval
before creating durable knowledge.

## Decision

Promote propose-before-create candidate generation as skill policy. Agents may
propose candidate `document.path`, `document.title`, and `document.body` from
explicit user-provided content when the candidate can be made
strict-runner-compatible and the final answer asks for confirmation before
creation.

The agent must not call `create_document`, `append_document`, or
`replace_section` before user approval. It may use existing runner actions such
as `validate`, `search`, `list_documents`, and `get_document` to check strict
JSON compatibility or duplicate risk. The final answer must show the proposed
path, title, and body preview clearly enough for the user to approve, revise,
or reject.

The skill policy is justified by candidate quality evidence, not by
`runner_capability_gap` evidence. A passing quality lane must show stable
conventional paths, useful titles, faithful bodies, duplicate-aware placement,
explicit override precedence, and confidence-to-ask behavior. The refreshed
targeted lane completed every selected quality scenario with classification
`none`, so the skill-policy implementation was authorized. The broader `oc-99z`
ergonomics promotion is deferred because the refreshed scorecard found
candidate-quality gaps in natural-intent behavior.

## Policy

Supported candidate inputs:

- pasted notes or excerpts with enough body content to preserve faithfully
- content with a clear heading that can become a title and slug
- user-supplied URL summaries where the user supplied the claims to preserve
- mixed-source snippets where no network fetching is required
- transcript excerpts or operational notes with clear durable note intent

No-tools clarification remains required when there is no supplied body content,
only a bare URL needing source ingestion hints, unclear durable artifact type,
invalid limits, bypass requests, or insufficient confidence to produce a
faithful candidate.

Explicit user instructions override candidate conventions. If the user supplies
a path, title, or body, the proposal must preserve those values unless they
conflict with runner validation or runner-visible authority.

Duplicate risk must be checked through existing runner-visible actions when the
workflow is already valid. If a likely duplicate is visible, the agent asks
whether to update the existing document or create a new one at a confirmed path.

## Non-Goals

This ADR does not:

- change `openclerk document` or `openclerk retrieval` schemas
- add an autofiling or proposal runner action
- relax runner validation
- authorize direct create-then-report behavior
- add storage migrations, indexes, background placement, or public API changes
- permit direct vault inspection, direct SQLite, broad repo search, HTTP/MCP
  bypasses, source-built runner paths, backend variants, module-cache
  inspection, or unsupported transports

## Skill Policy Gate

The `skills/openclerk/SKILL.md` policy must preserve the
no-create-before-approval boundary, explicit override precedence,
low-confidence clarification, duplicate checks through existing runner actions,
and strict runner JSON compatibility. Any future direct-create behavior still
requires a separate decision and eval gate.
