# Local-First Hybrid Retrieval Implementation Candidates

## Scope

This eval supports `oc-9ijx`. It compares implementation candidates after
`oc-rlg7` and `oc-ye6w` proved a real semantic-recall gap but did not justify a
durable embedding store, provider configuration, background indexer, or default
ranking change.

Inputs:

- `docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md`
- `docs/evals/semantic-recall-hybrid-vector-prototype.md`
- `docs/architecture/semantic-recall-hybrid-vector-decision.md`
- `docs/architecture/hybrid-retrieval-promotion-decision.md`
- current `openclerk retrieval` `search` and `hybrid_retrieval_report`
  behavior

## Baseline Evidence

The reduced semantic-recall prototype compared current lexical FTS with a
Gemini-backed vector prototype over 12 committed architecture documents and
100 heading-section chunks. After collapsing duplicate chunk hits to one cited
hit per document:

| Method | Hit@3 | MRR | Production posture |
| --- | ---: | ---: | --- |
| Current lexical FTS | 0/8 | 0.000 | Production default remains local-first and citation-bearing. |
| Provider vector-only | 8/8 | 0.938 | Evidence only; embedding generation required network/provider text transfer. |
| Provider hybrid | 8/8 | 0.938 | Evidence only; no durable store, provider config, or ranking change was authorized. |

Current `hybrid_retrieval_report` packages lexical baseline evidence and
candidate boundaries. It does not create embeddings, call providers, build
vectors, or return vector-ranked evidence.

## Candidate Matrix

| Axis | Local/offline hybrid index | Opt-in provider embeddings | Lexical tuning/no-vector fallback |
| --- | --- | --- | --- |
| Safety | Potential pass only after proving local runner-only execution, model provenance, rebuild visibility, stale-index invalidation, duplicate collapse, citation preservation, and no hidden authority ranking. | Advanced opt-in only; requires privacy disclosure, credential handling, rate-limit controls, cache rebuildability, no committed secrets/logs, and approval before durable writes. | Strongest default safety posture; still needs regression checks for exact/source-sensitive lookup and ranking surprises. |
| Capability | Best candidate if a local model can recover paraphrase, synonym, concept, and indirect-source queries close to the provider-vector prototype. | Proven benchmark value on the reduced corpus, but capability depends on provider availability and cost/rate limits. | May improve natural query recall through normalization, fallback query expansion, title/heading weighting, or aliases, but may not close the full semantic gap. |
| UX quality | Good only if hidden behind plain `search` with automatic freshness/rebuild behavior. | Poor as default because normal users would face provider setup, disclosure, retries, and cache concepts before retrieval. | Best UX if invisible; no new setup or mode choice. |
| Citation correctness | Must preserve repo-relative path, `doc_id`, `chunk_id`, heading, and line-span citations from canonical chunks. | Prototype preserved citations by ranking chunks, but durable provider-cache behavior remains unproven. | Already citation-bearing; tuning must not remove chunk citations or snippets. |
| Duplicate handling | Must collapse duplicate chunk hits to strongest cited document evidence and avoid treating embedding similarity as truth. | Prototype had high raw duplicate chunk pressure before document collapse. | Existing chunk ranking can duplicate documents; tuning must report duplicate pressure. |
| Freshness | Requires content hashes, stale-index detection, partial rebuilds, and visible recovery after document change/delete. | Requires cache invalidation and rebuildability plus provider retry behavior. | Uses existing FTS sync/rebuild path; any tuning must preserve current stale/rebuild diagnostics. |
| Import/rebuild cost | Highest unknown; must measure import, rebuild, reopen, and query timing at reduced and larger corpus sizes. | High and rate-limit-sensitive during embedding creation; cached local search after embedding can be fast. | Lowest operational cost; should stay near current FTS import/rebuild costs. |
| Privacy/offline fit | Best target if model and index work offline. | Does not fit routine local-first default because corpus/query text leaves the machine during embedding creation. | Best fit; no external text transfer. |
| Approval boundaries | Durable index and default ranking still need later promotion approval. | Provider setup, durable cache, background embedding, and default usage require explicit opt-in and approval. | Default changes still need promotion evidence if ranking behavior changes. |

## Result

Select local/offline hybrid retrieval as the next product candidate to test,
not to implement directly in `oc-9ijx`. Combine it with a lexical fallback
eval because lexical improvements are lower risk and may reduce the amount of
vector work needed.

Keep provider embeddings as reference evidence and optional future advanced
opt-in only. They should not become the production default unless a later
decision accepts the privacy, approval, rate-limit, and disclosure burden.

## Safety Pass

Pass for the decision. No runtime behavior changes, durable writes, provider
calls, embedding stores, vector stores, raw vault inspection, direct SQLite
reads, or default ranking changes are introduced here.

Partial pass for implementation readiness. The selected local/offline path must
still prove citation correctness, stale-index invalidation, duplicate handling,
privacy/offline fit, rebuild cost, and authority boundaries before promotion.

## Capability Pass

Pass for identifying the next evidence path. The prior prototype proved a real
semantic-recall gap and showed that chunk-level vector ranking can preserve
citations in principle. It does not prove that a local/offline model has enough
recall quality or that a durable local index is operationally acceptable.

## UX Quality

Pass for the selected direction. The desired surface remains simple:
source-grounded retrieval through `search`. A normal user should not choose
between FTS, local embeddings, provider embeddings, vector stores, caches, and
memory engines. Any future hybrid implementation should hide model/index
mechanics while making freshness and authority limits visible in reports.

## Follow-Up

Created follow-up Beads:

- `oc-bq8c`: local/offline hybrid retrieval index POC.
- `oc-o2r8`: lexical semantic-recall fallback tuning eval.

`oc-9ijx` can close as candidate selected with remaining implementation and
evidence work represented by those Beads.
