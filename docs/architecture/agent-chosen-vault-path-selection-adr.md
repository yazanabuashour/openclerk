---
decision_id: adr-agent-chosen-vault-path-selection
decision_title: Agent-Chosen Vault Path Selection
decision_status: deferred
decision_scope: document-path-selection
decision_owner: platform
---
# ADR: Agent-Chosen Vault Path Selection

## Status

Deferred after v1 and kept as reference after the targeted POC. This ADR
records the naming/path policy OpenClerk should evaluate for agent-chosen
vault-relative paths, but it does not add a public runner action, JSON schema,
storage migration, or implementation commitment.

The targeted POC/eval contract is recorded in
[`../evals/agent-chosen-path-selection-poc.md`](../evals/agent-chosen-path-selection-poc.md).
The current reduced report is
[`../evals/results/ockp-agent-chosen-path-selection-poc.md`](../evals/results/ockp-agent-chosen-path-selection-poc.md).

## Context

OpenClerk v1 keeps AgentOps as the production agent surface: routine agents use
the installed `openclerk` JSON runner plus `skills/openclerk/SKILL.md`. The v1
document runner requires create requests to name an explicit vault-relative
document path. That constraint is intentionally conservative because path
selection affects durable knowledge organization, duplicate risk, later
retrieval, and operator repairability.

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

This prompt supplies source material but no target path, title, document type,
or synthesis policy. A future capability must prove whether the current
explicit-path workflow is enough, whether skill guidance is enough, or whether
a promoted product behavior is justified.

## Decision

Keep agent-chosen vault path selection deferred/reference. The targeted POC did
not prove that explicit-path workflows or existing document/retrieval runner
actions are structurally insufficient.

The candidate naming/path policy to evaluate is:

- user-provided paths or naming instructions always win
- otherwise the agent chooses a clear, stable, vault-relative slug under the
  best conventional home
- the agent reports the chosen path to the user
- filenames and directories are conventions only
- metadata, not filename, remains authoritative for document type and identity

The policy must preserve the v1 model that canonical markdown and promoted
records are source authority. A path such as `sources/openai-harness-engineering.md`
may be a useful convention, but it must not determine whether the document is a
source, synthesis page, service, decision, promoted record, or durable answer.
Frontmatter metadata, runner-visible registry state, citations, source refs,
provenance, and projection freshness remain the authoritative signals.

## Interaction Shapes

Two interaction shapes should be compared before any promotion decision.

**Propose before create:** the agent derives a candidate path, title, document
kind, and source/synthesis intent, then asks for confirmation before writing.
This shape is safer for ambiguous placement and high-value durable knowledge,
but it adds a turn and may make simple capture workflows feel unnecessarily
ceremonial.

**Create then report:** the agent chooses the best conventional path, writes
the document through the existing runner workflow, and reports the chosen path.
This shape fits low-risk routine capture, but it can create duplicate or
misfiled durable knowledge if the agent guesses the home, document kind, or
source set incorrectly.

Both shapes must honor explicit user naming instructions. Neither shape implies
that OpenClerk should add an autonomous placement runner action in v1.

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

If promoted later, a separate implementation Bead must name the exact public
surface, request and response shape if any, backward compatibility
expectations, failure modes, and targeted eval gate. This ADR alone must not be
used to add a runner action or product capability.

## POC Result

The `agent-chosen-path-selection-poc` lane exercised proposal-before-create,
autonomous placement, multi-source synthesis path selection, ambiguous
metadata-authority placement, explicit user path precedence, and validation
pressure. Autonomous placement, synthesis path selection, explicit user path
precedence, missing-path clarification, and bypass rejection completed through
the existing `openclerk document` and `openclerk retrieval` public surface.

The lane did not justify promotion. Remaining validation failures were
classified as skill guidance or eval coverage around assistant answer wording
and final-answer-only handling, not as path-selection runner capability gaps.
Missing-path clarification remains the default until separate evidence proves a
product change is needed.
