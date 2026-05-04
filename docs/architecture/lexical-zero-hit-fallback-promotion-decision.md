---
decision_id: decision-lexical-zero-hit-fallback-promotion
decision_title: Lexical Zero-Hit Fallback Promotion Decision
decision_status: accepted
decision_scope: lexical-semantic-recall-production-fallback
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/results/ockp-semantic-recall-lexical-fallback-regression-gated.md, docs/architecture/lexical-semantic-recall-fallback-production-design-decision.md
---
# Decision: Lexical Zero-Hit Fallback Promotion

## Status

Accepted for `oc-1pxu`: promote a conservative lexical fallback inside
production `openclerk retrieval search` only when SQLite FTS returns zero hits.

## Decision

Add token-overlap scoring with title, path, heading, and body weighting after a
zero-hit FTS result. Preserve the existing Search JSON schema, path-prefix
filters, metadata filters, tag filters, citation fields, snippets, pagination,
and one strongest chunk per document.

Alias expansion remains eval-only. It scored higher on the reduced
semantic-recall set, but curated domain aliases are too surprising for default
ranking without a separate authority and maintenance model.

Regression-gated evidence:

| Method | hit@3 | MRR | Notes |
| --- | ---: | ---: | --- |
| current `Search` after zero-hit fallback | 7/8 | 0.900 | Production runner path; no schema change |
| eval token-overlap fallback | 7/8 | 0.833 | Same conservative family as production fallback |
| eval alias-overlap fallback | 8/8 | 0.938 | Kept eval-only |

## Safety, Capability, UX

Safety pass: pass. Exact FTS matches keep their existing ranking path because
fallback runs only after zero hits. Source filters, citations, and document
collapse are preserved.

Capability pass: pass for a narrow production fallback. The fallback improves
semantic-recall pressure without embeddings, provider calls, migrations, or
public schema changes. It does not replace vector/hybrid retrieval for deeper
semantic recall.

UX quality: pass. Normal users keep using `openclerk retrieval search`; there
are no ranking knobs, provider prompts, or separate query modes.

## Compatibility

No JSON schema, public command, provider configuration, durable cache, or
embedding store changed. The behavior change is limited to zero-hit lexical
searches.

