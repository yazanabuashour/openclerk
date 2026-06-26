---
decision_id: adr-agent-side-knowledge-intake-autofiling
decision_title: Agent-Side Knowledge Intake And Autofiling
decision_status: accepted
decision_scope: agent-interaction-policy
decision_owner: platform
---
# ADR: Agent-Side Knowledge Intake And Autofiling

## Status

Accepted as an agent interaction policy for routine OpenClerk knowledge intake.

This ADR defines when an agent may infer or propose `document.path`,
`document.title`, `document.body`, tags, and metadata fields from explicit
user-provided content, and when it must ask first. It does not change the
`openclerk document` or `openclerk retrieval` JSON schemas, the installed
runner, storage behavior, or public API.

`oc-wm04` updates the shipped skill policy to make proposal-first intake the
default for supported content: the agent/OpenClerk may choose candidate path,
title, body preview, tags, fields, and next approved request shape, while
durable writes still require approval.

Prior path/title autonomy evidence is reference-only. The `oc-iat` decision
kept explicit/no-tools behavior and did not promote a constrained autonomy
policy, runner action, schema, storage migration, skill behavior, or public
OpenClerk interface. The related evidence is recorded in
[`agent-chosen-vault-path-selection-adr.md`](agent-chosen-vault-path-selection-adr.md)
and
[`../evals/results/ockp-path-title-autonomy-pressure.md`](../evals/results/ockp-path-title-autonomy-pressure.md).

## Context

OpenClerk's production agent surface is AgentOps: routine agents use the
installed `openclerk` JSON runner through `openclerk document`,
`openclerk retrieval`, and `skills/openclerk/SKILL.md`. The runner remains
strict. Document creation requires explicit `document.path`, `document.title`,
and `document.body`. Source URL create mode requires explicit `source.url`,
`source.path_hint`, and `source.asset_path_hint`.

Users still issue natural intake requests such as "document this", "save this
note", "capture these links", or "update the existing synthesis". Those
requests can contain enough content to form a document body, tags, fields, or
source placement plan, but may omit the durable location, title, document kind,
source hints, or update target. The agent policy must preserve a useful
distinction:

- propose defaults from explicit user-provided content or runner-supported
  public-source context when a faithful no-write plan can be formed
- ask before creating durable knowledge when required fields, target identity,
  placement, or body content are missing
- reject lower-level bypasses and invalid values without tools

The policy is intentionally agent-side and planning-oriented. It tells agents
how to interact before calling strict write actions; it does not make the
runner guess missing fields for durable writes.

## Decision

OpenClerk accepts a proposal-first strict agent-side intake policy.

Agents may use `openclerk document` or `openclerk retrieval` when the required
request fields are explicit, directly derivable from explicit user-provided
content, or valid for a read-only planning action such as
`artifact_candidate_plan` or `ingest_source_url` plan mode. Direct derivation
means the agent is formatting, normalizing, or wrapping content the user
supplied. Planning may choose candidate placement, title, tags, fields, and body
preview; durable write actions may not invent body substance, source identity,
or update targets.

Explicit user instructions override defaults. If the user supplies a
vault-relative path, title, document body, source hint, update target, or
instruction about document type, the agent must honor it unless it conflicts
with the runner contract or existing runner-visible state. When explicit
instructions conflict, the agent must surface the conflict instead of silently
choosing a winner.

Metadata, frontmatter, citations, source refs, provenance, projection freshness,
and runner-visible registry state remain authoritative. Filenames and
directories are conventions, not document identity. Canonical markdown and
promoted records outrank inferred placement, synthesis, memory, graph state, or
agent naming preferences.

## Supported Input Classes

**Explicit runner-shaped document requests:** The user provides enough
information to fill `document.path`, `document.title`, and `document.body`.
The agent may validate, create, append, or replace sections through
`openclerk document` and answer only from the JSON result.

**User-supplied text with explicit path, title, and body:** The user provides
natural-language content plus an explicit durable path and title. The agent may
derive the body by preserving or lightly formatting the supplied text. It may
not add unsupported facts or source-sensitive claims without citations or
source refs.

**Source URL ingestion with all hints:** The user supplies `source.url`,
`source.path_hint`, and `source.asset_path_hint`, or supplies an exact runner
JSON shape containing those fields. The agent may use source ingestion through
the runner. URL-only intake may use `ingest_source_url` plan mode or
`artifact_candidate_plan` to propose placement before any durable fetch/write;
direct create mode still requires the runner's explicit source path and asset
fields.

**Existing-document updates:** The user identifies the existing document or
provides a runner-visible way to find it, and the requested update operation is
clear. The agent may use runner list/search/get/update flows to target that
document. If multiple plausible targets exist, the agent asks instead of
choosing one silently.

**Retrieval-only lookups:** The user asks to list, search, inspect, or answer
from existing OpenClerk knowledge with valid retrieval fields. The agent may
use `openclerk retrieval` or document list/get flows and answer from JSON
results.

**Proposal-first "document this" requests:** The user provides enough explicit
content or runner-supported public-source context to form a faithful candidate
but omits path, title, tags, fields, or final body formatting. The agent should
use a read-only proposal surface, show the candidate, and ask before writing.

**Ambiguous or low-confidence intake:** The user omits body content, source
evidence, transcript text, durable artifact type, or update target, or multiple
runner-visible targets conflict. The agent must ask for the missing fields or
target before using write actions. It may not turn ambiguity into autonomous
autofiling.

## Interaction Modes

**Use runner:** When all required fields are present and the requested workflow
is supported, the agent uses installed runner JSON and answers only from the
JSON result.

**No-tools clarification:** Before using any runner or inspection tool, the
agent gives one assistant response and no tools when a routine request is
missing required content, source/video fields, retrieval fields, or an explicit
update target and no faithful proposal can be formed. The response names the
missing fields and asks the user to provide them.

**Final-answer-only rejection:** Before using tools, the agent rejects invalid
limits and requests to bypass the runner through routine lower-level runtime,
HTTP, SQLite, MCP, legacy or source-built command paths, unsupported
transports, backend variants, module-cache inspection, direct vault inspection,
or broad repository search.

**Conflict clarification:** If user instructions, runner-visible metadata, or
existing documents conflict, the agent reports the conflict and asks for the
intended target or value. It must not silently rewrite metadata authority,
retarget an update, or create a duplicate to avoid the conflict.

## Duplicate Checks

Before creating durable knowledge from nontrivial supplied content, the agent
should use runner-visible duplicate checks when the workflow is already valid.
For example, it may use document list/search/get or retrieval search to find an
existing document under an explicit path prefix, title, source URL, decision id,
or source ref named by the user.

Duplicate checking must stay inside AgentOps. Routine agents must not inspect
the vault directly, query SQLite, enumerate repository files, or use source-built
helpers to find duplicates.

When a likely existing target is found, the agent should update that target only
if the user made the target or update operation explicit. If the target is not
explicit, or multiple plausible targets exist, the agent asks before writing.
Creating a near-duplicate is allowed only when the user explicitly requests a
new document and provides the required path, title, and body.

## Non-Goals

This ADR does not:

- add an autonomous autofiling runner action
- change `openclerk document` or `openclerk retrieval` request or response
  schemas
- relax required document, retrieval, or source-ingestion fields
- relax approval-before-write for durable creates, updates, ingests, or repairs
- add storage migrations, background placement heuristics, or new indexes
- let agents infer source identity, document kind, or durable authority from
  filenames alone
- permit direct SQLite, direct vault inspection, broad repo search, HTTP/MCP
  bypasses, source-built runner paths, backend variants, module-cache
  inspection, unsupported transports, or ad hoc runtime programs for routine
  OpenClerk knowledge work

## Promotion Gates

Any future autonomous write behavior, broader autofiling behavior, or new public
runner action requires a separate implementation work item and targeted AgentOps eval
evidence. The evidence must show repeated `runner_capability_gap` failures
where existing document and retrieval workflows are structurally insufficient,
not merely awkward, underspecified, missing examples, missing fixture data, or
missing skill guidance.

Promotion must preserve explicit user instruction precedence, no-tools
validation, duplicate avoidance, metadata authority, citations or source refs,
provenance, projection freshness, and operator-visible repairability. If a
candidate cannot preserve those invariants, it remains deferred or reference
only.

## `oc-99z` Decision

Final amended decision: keep strict runner behavior. The capability path remains
accepted with no runner, schema, storage, migration, public API, autofiling, or
direct-create promotion. The ergonomics path is amended after `oc-9k3`: promote
only the already implemented propose-before-create skill policy in
`skills/openclerk/SKILL.md`. Do not promote autonomous autofiling, direct
path/title/body inference into a durable write, a runner action, a JSON schema
change, a storage change, public API behavior, or any direct-create behavior
from this intake lane.

The targeted `oc-u9l` eval report
[`../evals/results/ockp-document-this-intake-pressure.md`](../evals/results/ockp-document-this-intake-pressure.md)
completed the selected document-this scenarios with classification `none`.
Those scenarios covered missing-field clarification, explicit creation, source
URL missing hints, explicit overrides, duplicate candidates, existing-document
updates, and synthesis freshness. The evidence shows existing
`openclerk document` and `openclerk retrieval` workflows can express the
pressure-tested behavior without a runner capability gap.

Capability path: no promotion. Current `openclerk document` and
`openclerk retrieval` primitives safely express the tested workflows while
preserving strict validation, duplicate checks, metadata authority, source refs
or citations, provenance, and freshness.

Ergonomics path: promoted only for the existing propose-before-create skill
policy. The only candidate surface reviewed was the narrow skill policy
accepted in
[`agent-chosen-document-artifact-candidate-generation-adr.md`](agent-chosen-document-artifact-candidate-generation-adr.md).
For supported inputs, the agent may propose a candidate `document.path`,
`document.title`, and `document.body` from explicit user-provided content,
optionally validate the candidate or check duplicate risk through existing
runner actions, show the candidate path, title, and body preview, state that no
document was created, and ask for approval before any durable write. The
existing candidate-generation lane proves candidate quality and safety
boundaries, and the refreshed ergonomics scorecard
[`../evals/results/ockp-document-artifact-candidate-ergonomics.md`](../evals/results/ockp-document-artifact-candidate-ergonomics.md)
reports `promote_propose_before_create_skill_policy` with every selected
quality and ergonomics row classified as `none`. Natural-intent proposal,
scripted-control, duplicate-risk, and low-confidence rows pass while preserving
approval-before-write and strict runner compatibility.

Supported behavior after `oc-wm04`:

- use proposal-first runner JSON when explicit content or runner-supported
  public-source context can produce candidate path, title, body preview, tags,
  fields, confidence, duplicate posture, and next approved request shape
- ask once with no tools when required content, source/video fields, retrieval
  fields, or update targets are missing and no faithful proposal can be formed
- use durable-write runner JSON only when strict required fields and targets are
  explicit, approved, and compatible with runner-visible authority
- let explicit user path, title, body, tags, fields, source hints, and targets
  override defaults unless they conflict with runner-visible authority
- perform duplicate, freshness, and provenance checks through existing
  runner-visible list/search/get, `projection_states`, and
  `provenance_events` actions where the workflow is already valid
- keep metadata, frontmatter, canonical markdown, promoted records, and
  runner-visible registry state authoritative over inferred filenames or
  placement
- never call `create_document`, `append_document`, or `replace_section` for a
  proposed candidate until the user approves the target and write

Because no `runner_capability_gap` was found, this policy does not authorize
runner, schema, storage, migration, public API, or direct-create work. The
proposal-first skill policy in `skills/openclerk/SKILL.md` is the promoted
ergonomics surface. Any future direct-create, autonomous write, or runner
relaxation requires a separate decision with targeted eval evidence and exact
implementation gates.

The later corrected candidate-generation track is recorded in
[`agent-chosen-document-artifact-candidate-generation-adr.md`](agent-chosen-document-artifact-candidate-generation-adr.md).
It narrows this `oc-99z` outcome to runner, schema, storage, public API, direct
create, and strict-runner behavior. It evaluates candidate quality and
no-write-before-approval behavior using a quality rubric instead of a runner
capability-gap rubric. The refreshed `oc-9k3` ergonomics scorecard now passes
for the existing propose-before-create skill policy, which is the only
ergonomics surface promoted by the amended `oc-99z` decision.
