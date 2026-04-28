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

This ADR defines when an agent may infer `document.path`, `document.title`, and
`document.body` from explicit user-provided content, and when it must ask first.
It does not change the `openclerk document` or `openclerk retrieval` JSON
schemas, the installed runner, storage behavior, public API, or shipped skill
behavior.

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
requests can contain enough content to form a document body, but may omit the
durable location, title, document kind, source hints, or update target. The
agent policy must preserve a useful distinction:

- infer from explicit user-provided content when the runner request is already
  complete and unambiguous
- ask before creating durable knowledge when required fields, target identity,
  placement, or body content are missing
- reject lower-level bypasses and invalid values without tools

The policy is intentionally agent-side. It tells agents how to interact before
calling the strict runner; it does not make the runner guess missing fields.

## Decision

OpenClerk accepts a strict agent-side intake policy.

Agents may use `openclerk document` or `openclerk retrieval` only when the
required request fields are explicit or directly derivable from explicit
user-provided content. Direct derivation means the agent is formatting,
normalizing, or wrapping content the user supplied, not inventing durable
placement, title, document kind, source identity, or body substance.

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
the runner. URL-only intake remains missing required fields for create mode.

**Existing-document updates:** The user identifies the existing document or
provides a runner-visible way to find it, and the requested update operation is
clear. The agent may use runner list/search/get/update flows to target that
document. If multiple plausible targets exist, the agent asks instead of
choosing one silently.

**Retrieval-only lookups:** The user asks to list, search, inspect, or answer
from existing OpenClerk knowledge with valid retrieval fields. The agent may
use `openclerk retrieval` or document list/get flows and answer from JSON
results.

**Ambiguous "document this" requests:** The user provides content or links but
omits a required path, title, body, source hint, asset hint, document kind, or
update target. The agent must ask for the missing required fields before using
the runner. It may not turn ambiguity into autonomous autofiling.

## Interaction Modes

**Use runner:** When all required fields are present and the requested workflow
is supported, the agent uses installed runner JSON and answers only from the
JSON result.

**No-tools clarification:** Before using any runner or inspection tool, the
agent gives one assistant response and no tools when a routine request is
missing required fields. The response names the missing fields and asks the
user to provide them. This applies to missing `document.path`,
`document.title`, `document.body`, `source.url`, `source.path_hint`,
`source.asset_path_hint`, retrieval fields, or an explicit update target.

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
- update `skills/openclerk/SKILL.md`
- add storage migrations, background placement heuristics, or new indexes
- let agents infer source identity, document kind, or durable authority from
  filenames alone
- permit direct SQLite, direct vault inspection, broad repo search, HTTP/MCP
  bypasses, source-built runner paths, backend variants, module-cache
  inspection, unsupported transports, or ad hoc runtime programs for routine
  OpenClerk knowledge work

## Promotion Gates

Any future relaxation, autonomous autofiling behavior, path/title/body
recommendation surface, or new public runner action requires a separate
implementation Bead and targeted AgentOps eval evidence. The evidence must show
repeated `runner_capability_gap` failures where existing document and retrieval
workflows are structurally insufficient, not merely awkward, underspecified,
missing examples, missing fixture data, or missing skill guidance.

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

Supported behavior remains unchanged:

- ask once with no tools when required `document.path`, `document.title`,
  `document.body`, source hints, retrieval fields, or update targets are
  missing and no faithful propose-before-create candidate can be formed
- use runner JSON only when strict required fields and targets are explicit or
  directly derivable from explicit user-provided content
- let explicit user path, title, body, source hints, and update targets
  override defaults unless they conflict with runner-visible authority
- perform duplicate, freshness, and provenance checks through existing
  runner-visible list/search/get, `projection_states`, and
  `provenance_events` actions where the workflow is already valid
- keep metadata, frontmatter, canonical markdown, promoted records, and
  runner-visible registry state authoritative over inferred filenames or
  placement
- never call `create_document`, `append_document`, or `replace_section` for a
  proposed candidate until the user approves the target and write

Because no `runner_capability_gap` was found, `oc-umk` is not authorized to add
runner, schema, storage, migration, public API, or direct-create work. The
existing propose-before-create skill policy in `skills/openclerk/SKILL.md`
is the complete promoted ergonomics surface after `oc-60s`. Any future
direct-create, autofiling, or runner relaxation requires a separate decision
with targeted eval evidence and exact implementation gates.

The later corrected candidate-generation track is recorded in
[`agent-chosen-document-artifact-candidate-generation-adr.md`](agent-chosen-document-artifact-candidate-generation-adr.md).
It narrows this `oc-99z` outcome to runner, schema, storage, public API, direct
create, and strict-runner behavior. It evaluates candidate quality and
no-write-before-approval behavior using a quality rubric instead of a runner
capability-gap rubric. The refreshed `oc-9k3` ergonomics scorecard now passes
for the existing propose-before-create skill policy, which is the only
ergonomics surface promoted by the amended `oc-99z` decision.
