# AgentOps Knowledge-Plane POC Decision Record

## Status

Accepted as the final POC recommendation for `oc-54s`.

## Decision

OpenClerk should keep AgentOps as the only production agent surface and build
the source-linked docs/provenance synthesis path first.

The production contract remains:

- `openclerk document`
- `openclerk retrieval`
- the Agent Skills-compatible guidance in `skills/openclerk/SKILL.md`

No new public runner action, synthesis-specific transport, direct SQLite path,
HTTP path, MCP path, or source-built runtime path is part of this decision.

The next production slice is to harden the runner/skill validation failures
shown by the production eval report, then continue the source-linked
docs/provenance synthesis slice through the existing document and retrieval
actions. This keeps durable compiled knowledge in inspectable markdown while
preserving canonical source authority, provenance, citations, and projection
freshness.

## Evidence

The completed POC tracks support the chosen path:

- `oc-d2v` defined the benchmark matrix in
  `docs/evals/knowledge-plane-archetype-matrix.md`, including canonical docs,
  RAG retrieval, source-linked synthesis, graph navigation, provenance,
  promoted records, and graph/vector memory as a reference archetype.
- `oc-85c` showed canonical markdown path and link navigation are useful and
  repairable, but relationship-heavy tasks need AgentOps-backed link and graph
  inspection rather than directory listing alone.
- `oc-7qg` showed RAG-style retrieval can provide source discovery,
  citation-bearing answers, metadata filters, path filters, and repeated query
  behavior, but retrieval alone does not prove durable compounding, conflict
  repair, or structured-domain precision.
- `oc-rsj` showed the source-linked synthesis lifecycle can run through the
  existing document and retrieval actions: search sources first, create or
  update synthesis, file durable answers, repair stale claims, and avoid
  duplicates.
- `oc-etv` showed provenance and projection-state inspection can explain and
  repair source-linked synthesis freshness from canonical sources.
- `oc-vn2` decided a dedicated synthesis runner action is not needed for this
  slice because the existing document and retrieval workflows were sufficient.

The reduced production eval report in
`docs/evals/results/ockp-agentops-production.md` is the main cross-POC evidence
artifact. It records 14/18 production scenarios passing, including retrieval,
canonical docs navigation, source-linked synthesis, freshness repair,
append/replace, promoted records, service lookup comparison, mixed
synthesis/records, and multi-turn synthesis scenarios. The failed scenarios are
validation and contract-enforcement gaps: missing required fields, negative
limits, unsupported lower-level requests, unsupported transport requests,
no-tools invalid-request handling, and one lower-level bypass attempt that
still used broad repo search and direct SQLite. Those failures should be fixed before
release, but they do not change the selected knowledge-model path.

## Next Build Slice

The next build slice should be:

- harden no-tools invalid-request handling for OpenClerk knowledge requests:
  missing document fields should trigger one clarification response naming the
  missing fields, while invalid limits, unsupported lower-level workflows, and
  unsupported transports remain explicit rejects.
- keep source-linked synthesis behind `openclerk document` and
  `openclerk retrieval`; do not add a dedicated synthesis action unless future
  eval evidence shows repeated brittle multi-step behavior.
- keep synthesis markdown under `notes/synthesis/` with `type: synthesis`,
  `status: active`, `freshness: fresh`, stable `source_refs`, `## Sources`,
  and `## Freshness`.
- preserve the workflow that searches canonical sources first, lists existing
  synthesis candidates, retrieves existing synthesis before updating, repairs
  stale or contradictory claims, and updates rather than duplicates.
- continue using production AgentOps evals as the release gate for source
  authority, citation correctness, provenance, freshness, and bypass
  prevention.

## Later POCs

Later POCs should stay behind the same AgentOps contract and should only move
forward when eval evidence shows they add value over the docs/provenance path:

- graph/vector memory reference behavior, including ontology grounding,
  temporal recall, session promotion, and feedback weighting.
- broader promoted-record domains beyond the service registry when typed lookup
  improves precision, update safety, or repeatable lookup without weakening
  citations.
- autonomous routing across docs, records, graph, and future memory only after
  docs, synthesis, provenance, and truth-sync behavior are reliable.
- graph navigation improvements that remain derived from canonical markdown and
  keep projection freshness inspectable.

## Kill Criteria

Keep a layer optional or remove it if it:

- behaves mainly like a more complicated way to do docs retrieval.
- obscures canonical source authority or lets derived knowledge outrank
  canonical docs and promoted canonical records.
- increases duplicate or conflicting truths.
- cannot explain provenance, citations, source refs, or freshness for
  source-sensitive answers.
- encourages routine agents to bypass the OpenClerk AgentOps runner for direct
  SQLite, HTTP, MCP, backend variants, module-cache inspection, source-built
  command paths, or ad hoc runtime programs.
- improves one workflow class while regressing source-grounded retrieval,
  citation correctness, synthesis lifecycle reliability, projection freshness,
  validation rejection, or bypass prevention.

## Non-Goals

- Do not add a dedicated synthesis runner action for the current slice.
- Do not introduce Cognee or another graph/vector memory dependency.
- Do not expose memory-first `remember` or `recall` semantics as OpenClerk's
  product surface.
- Do not make HTTP, MCP, direct SQLite, backend variants, module-cache
  inspection, source-built command paths, or ad hoc runtime programs routine
  production agent paths.
- Do not let graph, records, synthesis, memory, search indexes, feedback
  weighting, or routing outrank canonical docs and promoted canonical records.
- Do not treat the current validation failures as permission to bypass
  AgentOps; they are release-blocking contract enforcement work.
