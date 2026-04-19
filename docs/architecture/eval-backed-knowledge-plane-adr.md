# ADR: Eval-Backed Knowledge Plane Direction

## Status

Accepted as the provisional architecture direction. Binding adoption still
depends on the eval gates in this document.

## Context

OpenClerk is being shaped as a local-first, agent-facing knowledge plane. The
main design question is how to let useful knowledge compound over time without
turning the vault into an opaque memory system or a multi-store system with no
truth-maintenance model.

Karpathy's LLM Wiki pattern validates an important part of the direction:
agents should maintain durable markdown synthesis, links, contradiction notes,
and filed answers instead of rediscovering everything from raw sources on every
query. OpenClerk should not clone that pattern literally. It should implement
the useful part as source-linked synthesis inside a provenance-backed docs
layer.

The second design question is the agent-facing interface. OpenClerk now has one
task-shaped production surface: AgentOps, consisting of the installed
`openclerk` JSON runner and `skills/openclerk/SKILL.md`. That surface is easier
to evaluate and constrain than ad hoc runtime programs, direct SQLite access, or
backend-specific workflows.

## Options Considered

- **Current vault baseline:** markdown notes plus human navigation and current
  retrieval behavior. This remains the baseline that new layers must beat.
- **Literal LLM Wiki:** raw sources, an LLM-owned markdown wiki, and an
  instruction file. This is useful as a workflow pattern, but too loose as the
  authority model for OpenClerk.
- **Docs/provenance synthesis:** canonical markdown docs plus source-linked
  synthesis, citations, provenance events, projection freshness, search, and
  graph navigation. This is the first architecture slice to prove.
- **Full docs, records, memory, and router:** the target shape for selected
  future domains, but too much to adopt before docs/provenance eval evidence.
- **AgentOps runner:** task-shaped document and retrieval operations through
  the installed `openclerk` binary plus the production skill. This is the
  production agent contract.
- **Human CLI:** useful for humans and debugging, but not the routine agent
  contract.
- **MCP:** a possible adapter if it wraps the same OpenClerk runner semantics and beats
  the runner on measured agent behavior.
- **Public importable Go workflows:** removed from the product contract. Routine
  agents and developer-facing examples must not bypass the runner through a
  public lower-level package.

## Decision

OpenClerk will use the docs/provenance synthesis architecture as the first
proof slice:

- canonical docs remain markdown-backed and inspectable
- source-linked synthesis lives inside the docs layer
- synthesis must preserve source refs, citations, or equivalent stable evidence
- provenance and projection freshness are required before broad memory adoption
- graph/search are derived docs capabilities, not independent truth systems
- service registry is the first typed promoted-domain prototype, and additional
  records are promoted only for domains that beat plain docs on evals
- memory and autonomous routing remain deferred until the docs and truth-sync
  layers are benchmarked

OpenClerk will keep AgentOps as the production agent interface:

- routine agents use `openclerk` and task-shaped JSON
- CLI and MCP may be evaluated only as adapters over equivalent task shapes
- direct SQLite, backend variants, module-cache spelunking, and ad hoc runtime
  programs are not routine production-agent paths

## Invariants

- Canonical docs and promoted records outrank synthesis and memory.
- Source-linked synthesis is durable compiled knowledge, not a higher authority
  than the sources it cites.
- Every source-sensitive synthesis result must retain source refs, citations, or
  stable identifiers that let an agent inspect the evidence.
- Derived graph, records, search indexes, and future memory entries must expose
  freshness or provenance sufficient to explain their relationship to canonical
  docs or records.
- Routine agent tasks must use the OpenClerk AgentOps surface unless an evaluated
  adapter proves it can preserve the same contract with better measured behavior.
- New public API surface is added only after evals show the current surface is
  insufficient.

## Eval Gates

A layer or adapter can become permanent only when it satisfies all applicable
gates:

- production OpenClerk AgentOps passes the selected knowledge-plane scenarios
- source-linked synthesis is updated rather than duplicated
- source-sensitive answers preserve citations, chunk ids, paths, or explicit
  source refs
- provenance and projection-state reads can explain freshness
- promoted records, including the service registry, improve precision or update
  safety over plain docs for the target domain
- candidate CLI or MCP adapters match OpenClerk runner correctness and improve at least
  one measured agent-behavior metric without increasing forbidden access
- no production scenario requires direct SQLite, backend variants, module-cache
  inspection, broad repo search, stale surface inspection, or routine
  lower-level runtime work

## Kill Criteria

Keep a layer optional or remove it if it:

- behaves mainly like a more complicated way to do docs retrieval
- obscures canonical source authority
- increases duplicate or conflicting truths
- cannot explain provenance or freshness
- encourages routine agents to bypass OpenClerk runner for lower-level APIs
- improves one benchmark class while regressing core source-grounded retrieval
  or citation correctness

## Beads Ownership

- `oc-sg6` owns this architecture decision.
- `oc-0em` owns the source-linked synthesis prototype slice.
- `oc-alp` owns completion of the full eval matrix.
- `oc-0cm` remains blocked until evidence shows a promoted structured domain
  beats plain docs.
