---
decision_id: adr-artifact-intake-autofiling-tags-fields
status: accepted
decision_scope: artifact-intake
source_refs: docs/architecture/agent-knowledge-plane.md, docs/evals/artifact-intake-autofiling-tags-fields-poc.md, docs/evals/results/ockp-artifact-intake-autofiling-tags-fields.md
---

# ADR: Artifact Intake, Auto-Filing, Tags, and Fields

## Required References

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Context

Previous candidate-generation work proved a safe propose-before-create policy
for path, title, and body, but left too much routine artifact intake ceremony in
skill prose. A normal OpenClerk user expects "document this invoice/receipt/legal
note/transcript" to produce a candidate path, title, body preview, tags, fields,
duplicate posture, and approval boundary without exact command choreography.

The natural surface is not a generalized parser-backed `ingest_artifact`. It is
a read-only planning action that uses explicit user content or runner-supported
public-source handoff context, then leaves durable writes to existing approved
write actions.

The source-control boundary is the same: planning may suggest checkpoint
context, but Git status/history/checkpoint behavior belongs to
`git_lifecycle_report` and checkpoint commits require the explicit
`--git-checkpoints` or `OPENCLERK_GIT_CHECKPOINTS=1` gate. This ADR does not
authorize automatic commits, branch operations, restore, or remote operations.

## Candidate Options

| Candidate | Safety | Capability | UX quality | Decision |
| --- | --- | --- | --- | --- |
| Extend existing natural runner actions only | Pass, because current create/fetch actions preserve explicit approval. | Can work when the caller supplies exact path, title, tags, fields, and body. | Too ceremonial for routine artifact organization. | Keep as write path. |
| Dedicated read-only planning surface | Pass if it returns confidence, provenance, duplicate evidence, and no-write next requests. | Handles path, title, tags, fields, source handoff, and approval boundaries together. | Good: normal users get an inspectable candidate before durable writes. | Promote `artifact_candidate_plan`. |
| Dedicated durable organizing action | Not proven; could move, retag, or rewrite canonical docs unexpectedly. | Could reduce steps if approvals and diffs were exact. | Risky before review and rollback surfaces mature. | Do not promote. |
| Automatic source-control checkpointing | Not safe here; source-control writes must stay explicit and gated. | Could preserve local storage history. | Surprising if coupled to organization planning. | Use `git_lifecycle_report` only. |

## Decision

Promote a read-only `openclerk document` action named
`artifact_candidate_plan`.

The action plans:

- candidate path
- candidate title
- body preview
- artifact kind and source type
- tags
- metadata fields
- duplicate evidence and likely target
- confidence and confidence reasons
- approval, validation, and authority boundaries
- next approved `create_document` or `ingest_source_url` request shape when safe

No durable write, URL fetch, OCR, opaque file parse, local file read, browser
automation, direct vault inspection, direct SQLite access, or source-built runner
path is permitted.

Public read/fetch/inspect permission is not durable organization approval.
Plans may inspect explicit user content, public URL handoff context, and
runner-visible duplicate metadata, but path creation, metadata/tag writes,
source ingestion, and checkpoint commits remain separate approved actions.

## Supported Inputs

Supported:

- pasted or otherwise explicit text content
- explicit body markdown supplied in JSON
- user-provided public HTTP/HTTPS URL as handoff context, without runner fetch in
  this action
- explicit artifact kind, path, title, body, tags, and fields
- duplicate query and path prefix for runner-visible duplicate evidence

Unsupported in this action:

- opaque PDFs, images, screenshots, slide decks, email exports, chat archives,
  bundles, or local paths without pasted text
- OCR or parser truth claims
- private/authenticated acquisition
- durable create/update/fetch operations

## Override Precedence

Explicit user values win.

- `artifact.path` overrides generated path when it is vault-relative and ends in
  `.md`.
- `artifact.title` overrides heading/content/source-derived title.
- `artifact.body` overrides runner-assembled body preview.
- `artifact.tags` are preserved first; inferred tags may be appended.
- `artifact.fields` override inferred metadata keys with the same name.

Configurable defaults may influence generated candidates only when they remain
visible in the plan. They must not override explicit user values, and they must
not create durable organization changes without approval.

## Confidence Policy

High confidence requires explicit path, title, and body. Medium confidence is
allowed when the runner generated a faithful candidate from explicit content.
Low confidence is returned when no explicit body/content exists, the artifact
kind is unknown, or the result should only hand off to source URL ingestion.

Low confidence is not a failure; it is a no-write planning result that asks for
missing content, artifact type, or update-versus-new approval.

## Taste Check

The old shape was safe but too ceremonial: agents had to infer path/title/body,
then validate, then run duplicate checks, then restate approval boundaries. This
made successful flows depend on prompt choreography. The promoted surface keeps
safety and local-first authority intact while giving normal users the simpler
proposal they expect.

## Non-Goals

This decision does not promote:

- generalized `ingest_artifact`
- OCR/media parsing
- invoice, receipt, legal, or transcript semantic extraction beyond explicit
  content and explicit fields
- durable tag or metadata authority outside canonical markdown frontmatter
- autonomous writes
- vector memory, embedding stores, or hidden ranking authority

## Closure

Safety, capability, and UX quality remain separate gates:

- Safety pass requires explicit override precedence, visible defaults,
  confidence/provenance in returned plans, duplicate evidence, and approval
  before durable organization changes.
- Capability pass requires better path/title/tag/field planning than manual
  prompt choreography while keeping existing write actions authoritative.
- UX quality pass requires a normal user to receive a useful candidate without
  learning path policy or source-control gates.

Remaining work is represented by linked beads:

- `oc-tnnw.7.2` POC for naming/tagging/organizing/source-control evidence.
- `oc-tnnw.7.3` eval for safety, capability, and UX quality.
- `oc-tnnw.7.4` promotion decision.
- `oc-tnnw.7.5` conditional implementation only if promoted.
- `oc-tnnw.7.6` iteration and follow-up bead creation.
