# Knowledge-Plane Archetype Benchmark Matrix

This matrix defines benchmark coverage for AgentOps-native knowledge-plane
archetypes. It is a planning and evaluation contract for OpenClerk POCs, not a
new public runner API or alternate transport.

All executable benchmark work should preserve the current production surface:
agents use the installed `openclerk` JSON runner and the shipped
`skills/openclerk/SKILL.md` instructions. Direct SQLite access, HTTP calls,
backend variants, source-built command paths, MCP-style bypasses, and opaque
memory transports are not valid production paths for these benchmarks.

## Benchmark Axes

The matrix compares archetypes on these axes:

- **Graph utility:** whether graph traversal adds useful answers beyond search
  and explicit markdown links.
- **Ontology/entity grounding:** whether named entities, services, and typed
  records are grounded in canonical docs rather than inferred as free-floating
  facts.
- **Temporal retrieval:** whether current, observed, effective, stale, and
  superseded claims can be distinguished.
- **Feedback weighting:** whether learned or weighted recall can explain why a
  result is trusted without hiding weaker evidence.
- **Session memory promotion:** whether useful chat/session material becomes
  durable, source-linked markdown only after canonicalization.
- **Provenance explainability:** whether an operator can inspect the source,
  projection state, and freshness of a result.
- **Citation correctness:** whether answers preserve paths, `doc_id`,
  `chunk_id`, headings, or equivalent stable source references.
- **Duplicate/conflict rate:** whether workflows create duplicate synthesis or
  conflicting facts when updating existing knowledge.
- **Operator repairability:** whether stale, wrong, or ambiguous results can be
  repaired by editing canonical docs or rerunning runner-backed workflows.

## Archetype Matrix

| Archetype | Task Categories | Expected Outputs | Pass/Fail Gates | What It Proves |
| --- | --- | --- | --- | --- |
| Canonical docs and link navigation | Create and list canonical notes; retrieve exact docs; follow outgoing and incoming markdown links; inspect directory-shaped path prefixes. | Stable vault-relative paths, `doc_id` values, headings, metadata, link citations, and concise answers grounded in canonical markdown. | Passes only if results come through documented document/retrieval actions, links resolve to canonical docs, path-prefix filters scope results correctly, and final answers do not depend on direct vault inspection. Fails if the agent bypasses AgentOps, invents links, or loses source paths. | Proves the baseline markdown knowledge plane is inspectable, repairable, and usable before adding derived layers. |
| Basic RAG retrieval | Search by terms, headings, metadata, and path prefix; answer source-grounded questions from search hits; compare filtered and unfiltered retrieval. `oc-7qg` implements this as the `rag-retrieval-baseline` AgentOps eval scenario. | Ranked hits with snippets, citations, `doc_id`, `chunk_id`, source paths, and final answers that cite or name the relevant source. | Passes only if the correct source appears in top results, citations remain attached to claims, filters reduce scope, and unsupported negative limits or missing fields are rejected. Fails if answers cite the wrong source, omit source references, or use repo-wide search for routine knowledge work. | Proves semantic or lexical retrieval is useful for source discovery but does not by itself prove durable synthesis, conflict repair, or structured domain precision. |
| Source-linked synthesis lifecycle | Create synthesis from canonical sources; update existing synthesis; file reusable answers; repair stale or contradictory synthesis; avoid duplicates. | `notes/synthesis/` markdown with `type: synthesis`, `status: active`, `freshness: fresh`, single-line `source_refs`, `## Sources`, `## Freshness`, and source-sensitive claims tied to paths or chunk citations. | Passes only if the agent searches sources first, lists synthesis candidates, retrieves existing synthesis before updating, preserves source refs, and updates rather than duplicates. Fails if synthesis outranks canonical sources, uses YAML-list `source_refs`, drops citations, or writes unsupported actions. | Proves durable compiled knowledge can compound without becoming a second authority layer. |
| Graph navigation | Expand document links and graph neighborhoods; compare graph results with search and link expansion; answer relationship questions from derived graph state. | Source-linked graph nodes and edges with labels, relationship kinds, and citations back to canonical docs or chunks. | Passes only if graph results remain derived from canonical docs, preserve citations, and improve relationship discovery over plain search for graph-shaped questions. Fails if graph state becomes opaque truth, lacks source links, or contradicts canonical docs. | Proves graph traversal is valuable only when it adds relationship utility while staying refreshable and source-grounded. |
| Provenance and freshness repair | Inspect document events, projection invalidation, projection refresh, stale synthesis, superseded claims, and projection states before answering or repairing. | Answers and synthesis updates that name current sources, stale or superseded sources, relevant provenance events, and projection freshness state. | Passes only if source-sensitive repairs inspect `provenance_events` or `projection_states` when required, identify current evidence, and preserve repairable markdown state. Fails if freshness is asserted without inspection or stale claims remain unmarked. | Proves truth-sync behavior is operator-inspectable and that stale derived knowledge can be repaired through AgentOps workflows. |
| Promoted records | Create record-shaped and service-shaped canonical docs; query records and services; compare typed lookup with plain docs retrieval; refresh projections after source updates. | Typed record or service results with ids, names, status, owners, interfaces, facts, citations, provenance, and projection freshness. | Passes only if promoted lookup improves precision or structure without weakening citation correctness, and canonical markdown remains the source of truth. Fails if records become independent truth, stale projections are hidden, or typed lookup performs worse than docs retrieval for its target task. | Proves selective structured domains are justified only where they improve precision, update safety, or repeatable lookup behavior. |
| Graph/vector memory reference archetype | Compare future graph/vector memory behavior against docs, RAG, synthesis, graph, provenance, and records; evaluate ontology grounding, temporal recall, session promotion, and feedback weighting. | Reference results that explain source refs, freshness, temporal status, promotion path, and any feedback-derived ranking signals before they are trusted. | Passes only as a reference benchmark if memory outputs remain subordinate to canonical docs and promoted records, use AgentOps-compatible evidence, and expose provenance. Fails if it introduces `remember`/`recall` as a production surface, requires Cognee or another dependency, bypasses the runner, or hides stale/conflicting evidence behind ranking. | Proves which memory-engine capabilities may be worth future internal design pressure; it does not prove a new public interface or dependency should be adopted. |

## Coverage Expectations

Each implemented POC should identify which archetype it exercises and which
axes it is expected to improve. A useful benchmark result should include:

- The task prompt or scenario category.
- The runner actions or observable AgentOps behavior used.
- The expected durable output, if any.
- The required citations, source refs, provenance, or projection-state evidence.
- The pass/fail gate and the failure mode it is intended to catch.

Archetypes are cumulative rather than interchangeable. Basic RAG should be
measured against canonical docs, synthesis should be measured against both RAG
and canonical source authority, promoted records should be measured against
plain docs retrieval, and memory-style behavior should be measured against all
source, provenance, and repairability requirements before it is considered for
implementation.

## Non-Goals

- Do not add a new public runner action for this matrix alone.
- Do not introduce Cognee, graph/vector memory dependencies, or memory-first
  `remember`/`recall` semantics as OpenClerk product surface.
- Do not add hidden evaluator-only instructions to make scenarios pass.
- Do not treat graph, records, synthesis, or memory as higher authority than
  canonical docs and promoted canonical records.
