# OpenClerk Document Post-v0.1.0 Vision

## Status

Reference vision for post-v0.1.0 planning.

This document does not add a public runner action, schema, storage migration, or
API. It defines the product direction for `openclerk document` after the first
shipping release.

## Position

`openclerk document` should become the agent-first document lifecycle surface
for OpenClerk-compatible vaults.

That means OpenClerk is not a thin file wrapper. Markdown remains canonical
storage because it is inspectable, portable, editable, and durable, but routine
agents should operate documents through OpenClerk tasks rather than by
spelunking folders or editing files directly.

The v0.1.0 document surface proves the first slice:

- validate document requests
- resolve effective storage paths
- inspect convention-first layout
- create canonical markdown documents
- list and retrieve registered documents
- append durable content
- replace named sections
- keep source-linked synthesis fresh through retrieval-side provenance and
  projection inspection

The post-v0.1.0 direction is to turn that surface into a fuller document
lifecycle system for agent-authored durable knowledge.

## Why History Control Matters

Version and history control are not holdovers from human-first file management.
They become more important when agents write durable knowledge.

An agent-facing knowledge system needs to answer:

- what changed
- who or what changed it
- what source evidence justified the change
- which previous content was replaced
- whether the edit is pending review, accepted, restored, or superseded
- how to undo an unsafe or low-quality agent-authored edit

Git, sync providers, filesystem snapshots, and backups may provide storage-level
recovery. OpenClerk still needs semantic document history so agents can inspect
and reason about document lifecycle state through the runner contract.

## Post-v0.1.0 History Model

Future document history control should be an AgentOps-visible semantic layer.
The exact JSON shapes are intentionally not defined here; they require a future
implementation Bead, compatibility review, and targeted AgentOps evals.

The candidate model should cover:

- revision records for OpenClerk-authored document changes
- stable content hashes for before and after states
- edit summaries that explain why the change was made
- actor and source metadata, including whether a human, agent, import, or
  projection produced the change
- references to source evidence, provenance events, and projection state used
  during the edit
- before/after references that support diff inspection without committing raw
  private content into public artifacts
- restore or rollback semantics for OpenClerk-managed changes
- review queues for agent-authored changes that should not become accepted
  knowledge immediately

This is a semantic lifecycle layer, not a replacement for Git.

## Candidate Future Workflows

Post-v0.1.0 `openclerk document` workflows may include:

- inspect document history for a registered document
- compare the current document to a prior OpenClerk-managed revision
- explain why an agent-authored synthesis changed
- list pending document changes that need human review
- accept or reject an agent-authored draft
- restore a previous accepted revision
- annotate a rollback with the reason and source evidence
- show which synthesis pages became stale after a document revision

Each workflow must preserve the current v0.1.0 invariants:

- source-sensitive claims keep citations or source refs
- provenance and projection freshness remain inspectable
- canonical docs and promoted records outrank synthesis, memory, graph state,
  and routing choices
- routine agents do not use direct SQLite, broad repo search, source-built
  runner paths, HTTP/MCP bypasses, backend variants, module-cache inspection,
  or ad hoc runtime programs
- invalid routine requests still use no-tools handling where the skill requires
  it: missing required fields clarify, while invalid limits and bypass requests
  reject

## Path Choice And Organization

Autonomous vault path choice is aligned with the longer-term agent-first
knowledge-plane direction, but it is not part of the v0.1.0 document contract.

The current public request shape still requires explicit `document.path`.
OpenClerk should not silently guess or invent canonical paths in v1. When a
request is missing `document.path`, the correct interaction is a single
no-tools clarification response that asks the user to provide it.

Post-v0.1.0 evaluation may justify a narrower capability such as path
recommendation or constrained autonomous placement, but only if targeted
AgentOps evidence shows explicit user-provided paths are structurally
insufficient for routine knowledge work.

Any future promotion must preserve:

- vault-relative inspectability rather than hidden placement heuristics
- canonical markdown authority, source refs, provenance, and freshness
- explainable placement decisions through runner-visible evidence
- compatibility with existing explicit-path workflows
- no routine bypass to direct vault inspection or ad hoc filesystem logic

## Relationship To Git And Sync

OpenClerk should complement storage-level version control rather than replacing
it.

Git or a sync provider can answer:

- what bytes changed in a repository or folder
- how to restore a file snapshot
- how to review a commit or sync history

OpenClerk document history should answer:

- what knowledge object changed
- what agent workflow changed it
- what source evidence supported the change
- what OpenClerk revision or review state applies
- what derived projections or synthesis pages became stale
- what an agent may safely restore or ask a human to review

The layers are different. Git is storage history. OpenClerk history is
knowledge lifecycle history.

## Non-goals

Post-v0.1.0 document history control should not:

- implement a full Git clone inside OpenClerk
- make hidden autonomous rewrites routine
- promote memory as a source of authority
- introduce a new public runner action without targeted eval evidence
- let agents bypass the installed runner through direct vault edits or direct
  SQLite
- make review state invisible to the operator
- accept broad rewrite or contradiction-engine behavior without source refs,
  provenance, and freshness

## Promotion Gate

No history or review action should be implemented from this vision document
alone.

A future follow-up must first define:

- the exact workflow pressure that v0.1.0 cannot handle
- the candidate request and response shape
- compatibility behavior for existing vaults
- privacy expectations for raw diffs and private document bodies
- failure modes for stale, missing, rejected, or conflicting revisions
- whether path recommendation or autonomous placement needs a new public
  capability, and the eval pressure that justifies it
- targeted AgentOps eval scenarios
- pass/fail gates for source refs, provenance, freshness, rollback safety, and
  bypass prevention

The default decision remains defer until those gates show that document history
and review controls are needed for reliable dogfooding.
