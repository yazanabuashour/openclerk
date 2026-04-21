# ADR: AgentOps-Only Knowledge Plane Direction

## Status

Accepted as the current architecture direction.

## Context

OpenClerk is being shaped as a local-first, agent-facing knowledge plane. The
agent-facing substrate is AgentOps: the installed `openclerk` JSON runner plus
`skills/openclerk/SKILL.md`. Routine agents use task-shaped JSON through that
surface for document and retrieval work.

The remaining design question is how to let useful knowledge compound over time
without turning the vault into an opaque memory system or a multi-store system
with no truth-maintenance model.

Karpathy's LLM Wiki pattern validates an important part of the direction:
agents should maintain durable markdown synthesis, links, contradiction notes,
and filed answers instead of rediscovering everything from raw sources on every
query. OpenClerk should not clone that pattern literally. It should implement
the useful part as source-linked synthesis inside a provenance-backed docs
layer.

## Direction Considered

The research path considered several knowledge-model patterns for an
agent-first vault:

- **Current vault baseline:** markdown notes plus human navigation and current
  retrieval behavior.
- **Literal LLM Wiki:** raw sources, an LLM-owned markdown wiki, and an
  instruction file. This is useful as a workflow pattern, but too loose as the
  authority model for OpenClerk.
- **Docs/provenance synthesis:** canonical markdown docs plus source-linked
  synthesis, citations, provenance events, projection freshness, search, and
  graph navigation.
- **Cognee-style graph/vector memory engine:** graph and vector retrieval,
  ontology grounding, temporal search, session memory, feedback weighting, and
  memory-style agent integrations. This is useful as a reference architecture
  and benchmark input, but too memory-first and multi-surface to become
  OpenClerk's product contract.
- **Full docs, records, memory, and router:** the long-term shape for selected
  future domains, after the docs/provenance path is solid.

## Decision

OpenClerk uses AgentOps as the only production agent interface:

- routine agents use the installed `openclerk` runner with task-shaped JSON
- document and retrieval actions are the machine-facing product contract
- direct SQLite, backend variants, module-cache spelunking, ad hoc runtime
  programs, source-built command paths, HTTP server calls, and unsupported
  transports are outside routine production-agent work

OpenClerk uses the docs/provenance synthesis architecture as the first
knowledge-model build slice behind AgentOps:

- canonical docs remain markdown-backed and inspectable
- source-linked synthesis lives inside the docs layer
- synthesis must preserve source refs, citations, or equivalent stable evidence
  in `notes/synthesis/` markdown with `type: synthesis`, `status: active`,
  `freshness: fresh`, `source_refs`, `## Sources`, and `## Freshness`
- synthesis lifecycle behavior must search canonical sources first, find and
  update existing synthesis before creating duplicates, preserve source
  authority, repair stale or contradictory claims, and file reusable answers
  back into durable markdown
- provenance and projection freshness are required before broader memory or
  routing work
- graph/search are derived docs capabilities, not independent truth systems
- service registry is the first typed promoted-domain prototype
- memory and autonomous routing remain deferred until the docs, synthesis, and
  truth-sync layers are reliable through AgentOps

Cognee reinforces the need to evaluate graph/vector memory capabilities later,
especially ontology/entity grounding, temporal retrieval, feedback-weighted
ranking, and session-to-durable promotion. It does not change the accepted
AgentOps-only interface or make memory-first `remember`/`recall` semantics the
OpenClerk product surface.

## Invariants

- Canonical docs and promoted records outrank synthesis and memory.
- Source-linked synthesis is durable compiled knowledge, not a higher authority
  than the sources it cites.
- Every source-sensitive synthesis result must retain source refs, citations, or
  stable identifiers that let an agent inspect the evidence.
- Derived graph, records, search indexes, and future memory entries must expose
  freshness or provenance sufficient to explain their relationship to canonical
  docs or records.
- Graph/vector memory outputs must not outrank canonical docs or promoted
  records.
- Session-derived memory cannot become durable truth without canonicalization
  and provenance.
- Feedback weighting cannot hide stale or weakly sourced evidence.
- Routine agent tasks must use the OpenClerk AgentOps surface.
- New public runner actions are added only when repeated AgentOps workflows show
  that existing document and retrieval actions force brittle behavior.
- `oc-rsj` verified the current AgentOps document/retrieval runner actions are
  sufficient for source-linked synthesis lifecycle maintenance; no dedicated
  synthesis action or new public API is part of this slice.

## Regression Gates

Evals validate AgentOps behavior and knowledge-model quality.

- production OpenClerk AgentOps passes selected knowledge-plane scenarios
- source-linked synthesis is updated rather than duplicated
- source-sensitive answers preserve citations, chunk ids, paths, or explicit
  source refs
- provenance and projection-state reads can explain freshness
- promoted records, including the service registry, preserve citation
  correctness and improve structured lookup behavior for their target domain
- no production scenario requires direct SQLite, backend variants, module-cache
  inspection, broad repo search, stale surface inspection, or routine
  lower-level runtime work
- bypass requests for lower-level routine workflows are rejected final-answer-only

## Kill Criteria

Keep a layer optional or remove it if it:

- behaves mainly like a more complicated way to do docs retrieval
- obscures canonical source authority
- increases duplicate or conflicting truths
- cannot explain provenance or freshness
- encourages routine agents to bypass OpenClerk runner for lower-level APIs
- improves one workflow class while regressing core source-grounded retrieval,
  citation correctness, or synthesis lifecycle reliability
- creates opaque graph truth, bypasses runner tasks, weakens citation
  correctness, or increases stale-memory and conflicting-truth failures when
  borrowing Cognee-inspired behavior

## Beads Ownership

- `oc-sg6` owns this architecture decision.
- `oc-0em` owns the source-linked synthesis prototype slice.
- `oc-mjd` owns the AgentOps-only direction cleanup and synthesis lifecycle
  next-step alignment.
- `oc-0cm` owns the first promoted structured-domain prototype.
