---
decision_id: decision-hybrid-retrieval-report
decision_status: accepted
decision_scope: hybrid-retrieval
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/hybrid-retrieval-adr.md, docs/evals/hybrid-retrieval-candidate-comparison-poc.md, docs/evals/results/ockp-hybrid-retrieval-candidate-comparison.md
---

# Hybrid Retrieval Promotion Decision

## Decision

Accept `hybrid_retrieval_report` as a promoted read-only action under
`openclerk retrieval`.

The action packages current lexical baseline evidence and candidate-surface
guidance for hybrid/vector retrieval decisions. It does not promote durable
vector storage, external vector databases, OpenAI vector stores, live embedding
API calls, memory writes, or default ranking changes.

## Safety Pass

Pass. The implementation stays local-first and runner-only, performs no
durable writes, and rejects missing query input before any retrieval work.
Validation boundaries explicitly forbid direct SQLite, raw vault/file
inspection, source-built runners, HTTP/MCP bypasses, embeddings, vector stores,
and default ranking changes.

## Capability Pass

Pass for decision support. The report returns:

- `query`
- optional `path_prefix`
- `lexical_search`
- `candidate_surfaces`
- `recommendation`
- `safety_pass`
- `capability_pass`
- `ux_quality`
- `performance_posture`
- `evidence_posture`
- `validation_boundaries`
- `authority_limits`
- `evidence_inspected`
- `agent_handoff`

## UX Quality

Pass. The report replaces repeated baseline search plus architecture-policy
recitation with one natural retrieval action. Ordinary source-grounded
retrieval remains plain `search`, so the user-facing surface stays simple.

## Conditional Implementation

Implemented:

- runner JSON action `hybrid_retrieval_report`
- request object `hybrid_retrieval`
- read-only execution path
- result schema and `agent_handoff`
- CLI help text
- skill action index
- README workflow-action summary
- unit coverage for read-only behavior and missing-query rejection

No schema, projection, or storage changes were required because the selected
surface does not store embeddings or vectors.

## Iteration Gate

Before promoting durable hybrid ranking, file or execute follow-up work that
compares local vector storage, hosted vector stores, OpenAI vector stores, and
lexical-only FTS on recall, citation correctness, provenance, freshness,
import cost, reopen cost, and 100 MB/1 GB behavior.
