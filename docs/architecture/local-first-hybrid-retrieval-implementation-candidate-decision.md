---
decision_id: decision-local-first-hybrid-retrieval-implementation-candidates
decision_title: Local-First Hybrid Retrieval Implementation Candidate Decision
decision_status: accepted
decision_scope: semantic-recall-hybrid-retrieval
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/local-first-hybrid-retrieval-implementation-candidates.md, docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md, docs/architecture/semantic-recall-hybrid-vector-decision.md, docs/architecture/hybrid-retrieval-promotion-decision.md
---
# Decision: Local-First Hybrid Retrieval Implementation Candidates

## Status

Accepted: select a future local/offline citation-preserving hybrid retrieval
index as the best product candidate to test next. Pair it with a lexical
no-vector fallback eval. Keep explicit opt-in provider embeddings as
reference/prototype evidence only.

This decision does not add a runner action, schema, migration, embedding
store, provider configuration, public API, public OpenClerk interface,
background indexer, or default ranking change. It does not authorize durable
implementation work.

Evidence:

- [`docs/evals/local-first-hybrid-retrieval-implementation-candidates.md`](../evals/local-first-hybrid-retrieval-implementation-candidates.md)
- [`docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md`](../evals/results/ockp-semantic-recall-hybrid-vector-prototype.md)
- [`docs/architecture/semantic-recall-hybrid-vector-decision.md`](semantic-recall-hybrid-vector-decision.md)
- [`docs/architecture/hybrid-retrieval-promotion-decision.md`](hybrid-retrieval-promotion-decision.md)

## Decision

Select candidate 1 for the next evidence pass: a local/offline embedding model
with a citation-preserving local index, explicit embedding provenance, and
stale-index invalidation. This is the only candidate that can plausibly close
the measured semantic-recall gap while preserving routine OpenClerk
local-first expectations.

Combine it with candidate 3 as a companion fallback: lexical tuning should be
tested because it has the smallest safety and operational footprint, even if
the prior semantic-recall result suggests lexical FTS alone is unlikely to
recover the whole paraphrase/synonym/concept gap.

Do not select candidate 2 as a production default. Explicit opt-in provider
embeddings remain useful as benchmark/reference evidence, but provider text
transfer, credentials, rate limits, privacy disclosure, and approval ceremony
make them a poor default for routine local OpenClerk retrieval.

## Candidate Comparison

| Candidate | Safety | Capability | UX quality | Decision |
| --- | --- | --- | --- | --- |
| Local/offline embedding model plus local hybrid index | Potentially viable. Must prove runner-only access, local/offline operation, citation preservation, stale-index invalidation, duplicate collapse, rebuild visibility, and no hidden authority ranking. | Best fit for the measured gap if local model quality is sufficient. Prior provider-vector evidence showed semantic recall can recover documents lexical FTS missed, but local model quality is unproven. | Best product taste if hidden behind plain `search`; poor if users must manage model/index choices. | Select for next POC in `oc-bq8c`. |
| Explicit opt-in provider embeddings with rebuildable local cache | Safe only with clear opt-in, privacy disclosure, no credential leaks, bounded retries, cache rebuildability, and approval before durable writes. Not local-first during embedding generation. | Strong benchmark signal from the Gemini prototype: vector-only and hybrid recovered 8/8 expected docs after citation collapse. | Too much setup and disclosure ceremony for the default user surface; acceptable as advanced opt-in/reference only. | Keep as reference; do not promote as default. |
| Lexical tuning/no-vector fallback | Strongest default safety posture because it stays local, citation-bearing, and schema-light. Must avoid surprising ranking regressions or exact/source lookup damage. | Unproven for the semantic-recall pressure set; may improve recall with query normalization, OR/phrase fallback, title/heading weighting, or aliases, but may not close the full gap. | Best if improvements are invisible inside `search`; no new user-facing infrastructure. | Evaluate as fallback/companion in `oc-o2r8`. |

## Safety, Capability, UX

Safety pass: pass for this decision, partial for implementation readiness.
Current behavior remains unchanged. The selected local/offline candidate still
needs proof for embedding provenance, local model packaging, stale-index
invalidation, citation correctness, duplicate handling, rebuild/reopen cost,
and failure behavior. Provider embeddings do not pass as a default because
they send corpus/query text to an external service during embedding creation.

Capability pass: pass for preserving the valid need and selecting a plausible
next candidate. The prior prototype measured lexical FTS at 0/8 hit@3 and
vector-only plus hybrid at 8/8 hit@3 on the reduced semantic-recall set. That
proves a real gap, not that any local/offline model or durable index is ready
for production.

UX quality: pass for the selected direction. A normal user should keep asking
source-grounded questions through `search`; model choice, index rebuilds,
provider credentials, and vector-store mechanics should not become routine
user-facing decisions. The provider-opt-in shape fails as a default taste
choice even though it remains useful evidence.

## Follow-Up

Searches performed before closing `oc-9ijx`:

- `bd search "offline embedding retrieval" --status all`: no existing issue
  found.
- `bd search "local hybrid index" --status all`: no existing issue found.
- `bd search "lexical semantic recall tuning" --status all`: no existing
  issue found.
- `bd search "opt-in provider embeddings" --status all`: no existing issue
  found.

Created follow-up Beads:

- `oc-bq8c`: prototype a local/offline citation-preserving hybrid retrieval
  index and record model, index, freshness, duplicate, citation, privacy,
  rebuild, and scale evidence.
- `oc-o2r8`: evaluate lexical semantic-recall fallback tuning against the
  existing pressure set before any no-vector default-ranking change.

## Compatibility

Existing behavior remains unchanged:

- `openclerk retrieval search` remains lexical and citation-bearing.
- `openclerk retrieval hybrid_retrieval_report` remains read-only decision
  support and must not claim vector-ranked retrieval.
- Canonical markdown and promoted records remain authoritative.
- Embeddings, vector indexes, provider caches, and default hybrid ranking
  require later promotion evidence and explicit approval.
- Committed docs and reports must continue to use repo-relative paths or
  neutral placeholders such as `<run-root>`.
