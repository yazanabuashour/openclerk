# Hybrid Retrieval Candidate Comparison POC

## Scope

This POC supports `oc-uj2y.2`. It compares candidate surfaces for hybrid
embedding/vector retrieval without live embedding calls, generated corpora, raw
logs, or product behavior changes before the decision.

## Candidate Shapes

| Shape | What It Proves | What It Does Not Prove |
| --- | --- | --- |
| Current lexical FTS baseline | Citation-bearing retrieval, local-first behavior, stable doc/chunk IDs, existing scale posture. | Semantic recall gains for paraphrases or concept matches. |
| Search mode flag only | JSON contract feasibility. | Retrieval quality; without a second signal it is ceremonial UX. |
| Durable local vector index | Possible future semantic recall lane. | Freshness, rebuild cost, compression quality, and citation regression until a real index POC exists. |
| Hosted or external vector store | Useful benchmark/reference shape. | Local-first routine operation or approval boundaries. |
| Read-only `hybrid_retrieval_report` | A one-action runner report for baseline evidence, candidate comparison, validation boundaries, and handoff. | Vector-ranked retrieval or embedding-store behavior. |

## POC Result

The selected POC shape is `hybrid_retrieval_report` because it improves the
decision workflow without pretending that the repo has a durable vector index.

Runner request shape:

```json
{"action":"hybrid_retrieval_report","hybrid_retrieval":{"query":"semantic recall citation quality","path_prefix":"docs/","limit":10}}
```

Expected properties:

- read-only local runner action under `openclerk retrieval`
- citation-bearing lexical baseline
- candidate-surface comparison
- explicit validation boundaries
- explicit authority limits
- `agent_handoff`
- no embeddings, vector stores, HTTP/MCP bypass, direct SQLite, direct vault
  inspection, source-built runner, or default-ranking change

## Taste Review

A normal user should not have to choose between FTS, local vectors, hosted
vectors, OpenAI vector stores, and memory stores before asking a source-grounded
question. The natural surface stays `search`. The promoted POC only helps
agents and maintainers evaluate whether the next retrieval infrastructure step
has enough evidence to justify itself.
