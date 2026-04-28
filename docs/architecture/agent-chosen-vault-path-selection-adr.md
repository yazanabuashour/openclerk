---
decision_id: adr-agent-chosen-vault-path-selection
decision_title: Agent-Chosen Vault Path Selection
decision_status: deferred
decision_scope: document-path-selection
decision_owner: platform
---
# ADR: Agent-Chosen Vault Path Selection

## Status

Deferred after v1 and kept as reference after the targeted POC, post-`oc-6fr`
source URL update work, and the `oc-iat` path/title autonomy decision. This ADR
records the naming/path/title policy OpenClerk evaluated for agent-chosen
vault-relative paths, filenames, and titles, but it does not add a public runner
action, JSON schema, storage migration, skill behavior, public interface, or
implementation commitment.

The targeted POC/eval contract is recorded in
[`../evals/agent-chosen-path-selection-poc.md`](../evals/agent-chosen-path-selection-poc.md).
The current reduced report is
[`../evals/results/ockp-agent-chosen-path-selection-poc.md`](../evals/results/ockp-agent-chosen-path-selection-poc.md).
The follow-up pressure report for `oc-iat` is
[`../evals/results/ockp-path-title-autonomy-pressure.md`](../evals/results/ockp-path-title-autonomy-pressure.md).

## Context

OpenClerk v1 keeps AgentOps as the production agent surface: routine agents use
the installed `openclerk` JSON runner plus `skills/openclerk/SKILL.md`. The v1
document runner requires create requests to name an explicit vault-relative
document path, title, and body. Source URL ingestion in default `create` mode
requires explicit `source.path_hint` and `source.asset_path_hint`. Those
constraints are intentionally conservative because path and title selection
affect durable knowledge organization, duplicate risk, later retrieval, and
operator repairability.

Routine knowledge requests are not always path-shaped. A user may ask an agent
to document URLs, synthesize several sources, or capture an ambiguous note
without knowing whether the durable home should be `sources/`, `synthesis/`,
`records/`, `docs/`, or another conventional prefix. That creates legitimate
post-v1 pressure: explicit user-provided paths may be structurally insufficient
when the user intent is clear but the best durable location is a knowledge-plane
organization choice.

The required design/eval example is:

```text
let's document:

https://openai.com/index/harness-engineering/
https://developers.openai.com/api/docs/guides/prompt-guidance
```

This prompt supplies source material but no target path, filename, title,
document type, artifact path, or synthesis policy. A future capability must
prove whether the current explicit-field workflow is enough, whether skill
guidance is enough, or whether a promoted product behavior is justified.

The `oc-6fr` source URL update mode does not weaken that default. Missing
`source.mode` means `create`, and create requests still require explicit source
and asset path hints. `source.mode: "update"` targets an existing source by
normalized `source.url`; it does not create a missing source, it preserves the
stored source and asset paths, and any supplied path hints must match existing
metadata or conflict without writing.

## Decision

Keep agent-chosen vault path selection deferred/reference. The targeted POC and
the follow-up `oc-iat` pressure lane did not prove that explicit-path workflows
or existing document/retrieval runner actions are structurally insufficient.

The candidate naming/path/title policy to evaluate is:

- user-provided paths or naming instructions always win
- otherwise the agent chooses a clear, stable, vault-relative slug under the
  best conventional home
- the agent chooses a title from user instructions, source metadata, or concise
  human-readable subject text
- the agent reports the chosen path and title to the user
- filenames and directories are conventions only
- metadata, not filename, remains authoritative for document type and identity
- source URL create/update semantics keep source and asset paths explicit or
  storage-backed; title/path autonomy must not invent a second source identity

`oc-iat` confirms no promotion of this policy into product behavior. The
current production contract still requires explicit `document.path`,
`document.title`, and `document.body` for document creation, and explicit
`source.path_hint` and `source.asset_path_hint` for source URL create mode. A
URL-only, artifact-only, or ambiguous path/title request remains a no-tools
clarification when required fields are missing, unless the request can be
completed through an already valid existing `openclerk document` or
`openclerk retrieval` workflow.

The policy must preserve the v1 model that canonical markdown and promoted
records are source authority. A path such as `sources/openai-harness-engineering.md`
may be a useful convention, but it must not determine whether the document is a
source, synthesis page, service, decision, promoted record, or durable answer.
Frontmatter metadata, runner-visible registry state, citations, source refs,
provenance, and projection freshness remain the authoritative signals.

## Interaction Shapes

Four interaction shapes should be compared before any promotion decision.

**Explicit fields required:** the agent asks the user for required document
path, title, body, source path hint, asset path hint, or other artifact fields
before using the runner. This is the current conservative default for document
creation and source URL create mode. It preserves no-tools validation and
avoids guessed durable placement, but it can block otherwise clear capture or
documentation requests on naming decisions the user may not care about.

**Ask for missing fields:** the agent treats missing path/title/artifact fields
as a clarification turn, names the missing fields, and waits for user-supplied
values. This keeps explicit user intent authoritative and works with the
existing no-tools rule for required fields. It improves clarity over a generic
rejection, but it still adds user friction and does not test whether
OpenClerk-controlled conventions would be safe for low-risk captures.

**Propose before create:** the agent derives a candidate path, title, document
kind, artifact path if relevant, and source/synthesis intent, then asks for
confirmation before writing. This shape is safer for ambiguous placement,
high-value durable knowledge, and source-sensitive artifacts, but it adds a
turn and may make simple capture workflows feel unnecessarily ceremonial.

**Create then report:** the agent chooses the best conventional path and title,
writes the document through the existing runner workflow, and reports the chosen
path and title. This shape fits low-risk routine capture, but it can create
duplicate or misfiled durable knowledge if the agent guesses the home, document
kind, source set, title, or artifact path incorrectly.

All shapes must honor explicit user naming instructions. None imply that
OpenClerk should add an autonomous placement runner action in v1. The
`create then report` shape remains reference-only unless targeted follow-up
evidence, tracked by `oc-940`, proves a runner capability gap.

## Invariants

Any future path-autonomy capability must preserve these invariants:

- AgentOps remains the production surface: `openclerk document` and
  `openclerk retrieval` JSON results, plus the shipped skill guidance.
- Routine agents do not use broad repo search, direct vault inspection, direct
  SQLite, HTTP/MCP bypasses, source-built runner paths, backend variants,
  module-cache inspection, or ad hoc runtime programs.
- Source-sensitive claims retain citations, source refs, stable source paths,
  `doc_id`, `chunk_id`, headings, line ranges, or equivalent runner-visible
  evidence.
- Provenance and projection freshness remain inspectable when placement affects
  synthesis, promoted records, services, decisions, or stale derived outputs.
- Metadata, not filename, determines promoted records, services, decisions, and
  synthesis identity.
- Source URL `create` keeps requiring explicit source and asset path hints.
  Source URL `update` keeps targeting existing normalized `source.url`, with
  stable stored paths and conflict-on-mismatched-hint behavior.
- Missing required fields, invalid limits, and lower-level bypass requests keep
  the existing no-tools and final-answer-only validation behavior.

## Promotion Gate

Promotion requires targeted AgentOps eval evidence that the explicit-path
workflow is structurally insufficient for routine knowledge work. Repeated
failures must show more than awkwardness, missing examples, missing skill
guidance, missing fixture data, or thin eval coverage.

The default outcome remains defer. Promote only if the existing document and
retrieval workflow repeatedly fails to express URL-only documentation,
multi-source synthesis, or ambiguous document-type placement while preserving
citations, source refs, provenance, freshness, metadata authority, and
operator-visible repairability.

The `oc-iat` pressure lane found no `runner_capability_gap` failures. All six
selected scenarios in
[`../evals/results/ockp-path-title-autonomy-pressure.md`](../evals/results/ockp-path-title-autonomy-pressure.md)
completed with failure classification `none`, so no implementation bead should
start from that evidence.

If promoted later, a separate implementation Bead must name the exact public
surface, request and response shape if any, backward compatibility
expectations, failure modes, and targeted eval gate. That follow-up must also
state how explicit instructions, metadata authority, provenance/freshness,
duplicate avoidance, and no-tools validation are preserved. This ADR alone must
not be used to add a runner action or product capability.

## POC Result

The `agent-chosen-path-selection-poc` lane exercised proposal-before-create,
autonomous placement, multi-source synthesis path selection, ambiguous
metadata-authority placement, explicit user path precedence, and validation
pressure. Autonomous placement, synthesis path selection, explicit user path
precedence, missing-path clarification, and bypass rejection completed through
the existing `openclerk document` and `openclerk retrieval` public surface.
The later source URL update-mode work preserved explicit create-mode path hints,
stable update-mode source and asset paths, and conflict-on-mismatched-hint
behavior instead of promoting path autonomy.

The lane did not justify promotion. The refreshed guidance/eval hardening run
resolved the prior answer-wording failures for path proposal, metadata
authority, and invalid-limit rejection while preserving the existing public
surface. No path-selection runner capability gap was exposed. Missing-field
clarification remains the default until separate evidence proves a product
change is needed.

## `oc-iat` Decision

Decision: keep explicit/no-tools behavior and keep path/title autonomy as
reference evidence only. Do not promote a constrained autonomy policy, runner
action, schema, storage migration, skill behavior, public OpenClerk interface,
or missing-field policy change from the path/title pressure lane.

The refreshed `oc-iat` decision evaluated both promotion paths. On the
capability path, existing primitives expressed the selected workflows safely,
so no runner, schema, storage, migration, public API, direct-create behavior,
or autonomous path/title policy is promoted. On the ergonomics path, the only
promoted surface is the already implemented propose-before-create skill policy
recorded in
[`agent-chosen-document-artifact-candidate-generation-adr.md`](agent-chosen-document-artifact-candidate-generation-adr.md).

The follow-up lane exercised URL-only source documentation, artifact ingestion
without hints, multi-source duplicate pressure, explicit overrides, duplicate
risk, and metadata-authority pressure. Current `openclerk document` and
`openclerk retrieval` workflows handled the pressure without a classified
capability gap. Metadata and frontmatter remain authoritative over filename and
path conventions, and explicit user naming instructions still win when present.
