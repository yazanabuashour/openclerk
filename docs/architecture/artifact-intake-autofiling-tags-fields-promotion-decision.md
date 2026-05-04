---
decision_id: decision-artifact-intake-autofiling-tags-fields-promotion
status: accepted
decision_scope: artifact-intake
source_refs: docs/architecture/artifact-intake-autofiling-tags-fields-adr.md, docs/evals/artifact-intake-autofiling-tags-fields-poc.md, docs/evals/results/ockp-artifact-intake-autofiling-tags-fields.md
---

# Promotion Decision: Artifact Intake, Auto-Filing, Tags, and Fields

## Required References

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Decision

Promote and implement `openclerk document` action
`artifact_candidate_plan`.

Approved writes remain conditional:

- explicit document creation uses `create_document`
- public URL fetch/write uses `ingest_source_url`
- duplicate cases require update-versus-new approval before any write

## Safety Pass

Pass. The promoted action is read-only and runner-owned. It does not use OCR,
opaque file parsing, local file reads, browser automation, HTTP fetch, direct
vault inspection, direct SQLite, source-built runners, or unsupported
transports. It returns no durable write result and labels public source context
as handoff only.

## Capability Pass

Pass. The action covers candidate path, title, body preview, tags, metadata
fields, duplicate evidence, source URL duplicate status, confidence, and
approved create/ingest handoff.

## UX Quality

Pass. Existing primitives were safe but too ceremonial for normal artifact
intake. The promoted action removes repeated prompt choreography while keeping
explicit overrides, duplicate handling, and approval-before-write visible.

## Implementation Requirements

- Add `DocumentTaskActionArtifactPlan`.
- Add strict JSON request/response types.
- Validate negative limits, source URLs, artifact kinds, source types, and
  vault-relative markdown paths.
- Search runner-visible duplicate evidence only through the installed runner
  client.
- Return `agent_handoff` with evidence, validation boundaries, authority limits,
  and next-step guidance.
- Update runner help, README, skill guidance, tests, and decision docs.

## Non-Promotion Follow-Up

Not required. This is a promotion outcome, not `keep-as-reference`, `defer`,
`more evidence`, `none viable yet`, or another non-promotion result.
