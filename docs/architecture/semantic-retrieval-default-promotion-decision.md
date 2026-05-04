---
decision_id: decision-semantic-retrieval-default-promotion
decision_title: Semantic Retrieval Default Promotion Decision
decision_status: accepted
decision_scope: semantic-retrieval-default-search
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/results/ockp-semantic-retrieval-promotion-comparison.md, docs/evals/results/ockp-semantic-recall-local-embeddinggemma-promotion-rerun.md, docs/evals/results/ockp-semantic-recall-gemini-promotion-benchmark.md, docs/evals/results/ockp-semantic-recall-lexical-fallback-promotion-rerun.md, modules/semantic-retrieval-adapter/module.json
---
# Decision: Semantic Retrieval Default Promotion

## Status

Accepted for `oc-by5n`: keep `modules/semantic-retrieval-adapter` as an
optional module. Do not promote semantic/hybrid ranking into default
`openclerk retrieval search`, and do not add an explicit core semantic mode in
this Bead.

Evidence:

- [`docs/evals/results/ockp-semantic-retrieval-promotion-comparison.md`](../evals/results/ockp-semantic-retrieval-promotion-comparison.md)
- [`docs/evals/results/ockp-semantic-recall-local-embeddinggemma-promotion-rerun.md`](../evals/results/ockp-semantic-recall-local-embeddinggemma-promotion-rerun.md)
- [`docs/evals/results/ockp-semantic-recall-gemini-promotion-benchmark.md`](../evals/results/ockp-semantic-recall-gemini-promotion-benchmark.md)
- [`docs/evals/results/ockp-semantic-recall-lexical-fallback-promotion-rerun.md`](../evals/results/ockp-semantic-recall-lexical-fallback-promotion-rerun.md)

## Decision

Select candidate 1: keep the optional module.

Do not select candidate 2, default local hybrid ranking. Local Ollama
`embeddinggemma` remains at 7/8 hit@3 with 0.906 MRR, missing the explicit 8/8
promotion threshold and the known semantic retrieval gap.

Do not select candidate 3, explicit core semantic mode. The adapter still does
not expose tag or metadata filters, and normal-user core semantics would need a
durable local cache/index lifecycle and broader source-sensitive regression
evidence.

Do not select `none viable yet`. The optional adapter is viable for
maintainer/agent composition and continued evidence gathering.

## Candidate Comparison

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Keep optional module | Pass. Read-only, explicit provider behavior, external cache, no core schema or ranking change. | Partial but useful: local hybrid 7/8 and Gemini benchmark 8/8. | Acceptable for maintainers and agents. | Selected. |
| Promote local hybrid default search | Not ready: ranking authority would change and filter/cache lifecycle is incomplete. | Fails threshold: local hybrid remains 7/8. | Too surprising while users must manage local model/cache behavior. | Reject. |
| Add explicit core semantic mode | Safer than default promotion but still needs tag/metadata filters, cache lifecycle, and public contract design. | Not ready from current adapter shape. | Better than default promotion later, but premature now. | Defer. |
| None viable yet | Too conservative because optional module evidence is valid. | Would discard useful local/offline building block. | Poor for agent workflows. | Reject. |

## Safety, Capability, UX

Safety pass: pass for non-promotion. `openclerk retrieval search` remains
lexical plus zero-hit lexical fallback. No provider key, committed embedding
cache, provider config write, durable vector store, or hidden provider fallback
is introduced.

Capability pass: partial. Local/offline semantic retrieval is real but does not
yet clear default-search thresholds. Gemini confirms the vector/hybrid
mechanics can recover 8/8, but provider-backed evidence cannot justify a local
default.

UX quality: partial. Normal users deserve a simpler future surface than a
separate optional module, but promoting it now would expose model/cache/filter
differences before the evidence is strong enough.

## Follow-Up

Searches performed before closing `oc-by5n`:

- `bd search "semantic retrieval promotion evidence" --status all`: no
  existing issue found.
- `bd search "default semantic search local hybrid" --status all`: no existing
  issue found.
- `bd search "semantic retrieval cache index lifecycle" --status all`: no
  existing issue found.

Created deferred follow-up:

- `oc-sloi`: re-run semantic retrieval promotion after stronger local evidence.

## Compatibility

Core behavior is unchanged:

- `openclerk retrieval search` remains lexical plus zero-hit lexical fallback.
- `modules/semantic-retrieval-adapter` remains explicit optional tooling.
- Gemini remains explicit benchmark/fallback evidence only.
- Any future default-search or explicit core semantic mode needs a new Bead.

