---
decision_id: decision-semantic-recall-hybrid-vector
decision_status: deferred
decision_scope: semantic-recall-hybrid-vector
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/semantic-recall-hybrid-vector-prototype.md, docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md, docs/architecture/hybrid-retrieval-adr.md
---

# Semantic Recall Hybrid Vector Decision

## Decision

Defer product promotion, but keep the capability need alive.

The `oc-rlg7` and `oc-ye6w` evidence shows that current lexical FTS can miss
normal paraphrase, synonym, concept-recall, and indirect-source questions. A
real Gemini-backed vector prototype recovered the expected documents with
stable path, chunk, heading, and line-span citations on the reduced committed
doc corpus.

Do not promote durable embedding storage, provider configuration, background
indexing, hosted vector stores, or default hybrid ranking from this result.

## Safety Pass

Pass for evidence. The POC used isolated temp storage, committed docs, an
installed lexical runner surface, temporary embedding cache, no production
writes, no durable index, no default-ranking change, and no committed raw logs
or credentials.

Not a product safety pass. The tested embedding path required network access
and sent committed chunk/query text to an external provider. Routine OpenClerk
retrieval remains local-first and citation-bearing lexical FTS.

## Capability Pass

Pass for identifying a real retrieval-quality gap. The final chunk-level run
measured:

- lexical FTS: 0/8 hit@3, 0.000 MRR
- vector-only: 8/8 hit@3, 0.938 MRR
- hybrid: 8/8 hit@3, 0.938 MRR

Partial pass for implementation readiness. The POC does not settle the durable
store shape, local/offline embedding model, privacy disclosure, stale-index
repair path, large-corpus cost, or production duplicate-ranking behavior.

## UX Quality

The user need remains valid. A normal OpenClerk user should ask one
source-grounded question through `search` and receive cited evidence; they
should not decide between FTS, local embeddings, provider embeddings, vector
stores, and memory engines.

The evaluated shape is still too ceremonial and provider-dependent for normal
use. The next design pass should compare implementation candidates that hide
index mechanics behind the natural retrieval surface.

## Follow-up Beads

Searches performed before close:

- `bd search "local-first hybrid retrieval implementation" --status all`: no
  existing bead found.
- `bd search "citation-preserving vector retrieval" --status all`: no existing
  bead found.
- `bd search "offline embedding retrieval" --status all`: no existing bead
  found.

Created and deferred:

- `oc-9ijx`: compare local-first hybrid retrieval implementation candidates.

## Closure

`oc-rlg7` closes with outcome `promote follow-up candidate comparison`, not
product implementation. `oc-ye6w` closes as completed POC evidence. The
remaining implementation decision is represented by deferred bead `oc-9ijx`.
