---
decision_id: decision-lexical-semantic-recall-fallback-production-design
decision_title: Lexical Semantic-Recall Fallback Production Design
decision_status: deferred
decision_scope: semantic-recall-lexical-production-design
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/results/ockp-semantic-recall-lexical-fallback-design.md, docs/evals/results/ockp-semantic-recall-lexical-fallback.md, docs/evals/results/ockp-semantic-recall-gemini-provider-mimic.md, docs/architecture/lexical-semantic-recall-fallback-eval-decision.md
---
# Decision: Lexical Semantic-Recall Fallback Production Design

## Status

Deferred: keep lexical fallback as candidate evidence, not production search
behavior.

The design rerun in
[`docs/evals/results/ockp-semantic-recall-lexical-fallback-design.md`](../evals/results/ockp-semantic-recall-lexical-fallback-design.md)
reproduced the earlier lexical evidence:

- current lexical FTS: 0/8 hit@3, 0.000 MRR
- token-overlap fallback: 7/8 hit@3, 0.833 MRR
- alias-assisted fallback: 8/8 hit@3, 0.938 MRR

## Decision

Do not change production `openclerk retrieval search` ranking from `oc-1amj`.

Select lexical fallback as a future regression-gated implementation candidate
or hybrid companion, not as a direct promotion. The alias-assisted result is
strong on the reduced semantic-recall set, but it depends on curated aliases
and still carries high duplicate pressure. Production ranking needs
source-sensitive regression evidence before it can safely change.

## Safety, Capability, UX

Safety pass: pass for the decision. The run used copied committed docs,
committed reduced reports only, made no provider calls, changed no runner
schema, and did not alter default search.

Capability pass: partial. The fallback candidates prove no-vector semantic
recall pressure can be reduced, but they have not passed enough exact lookup,
path-prefix, metadata/tag, duplicate, citation, freshness, performance, or
scale-decoy regression evidence for production ranking.

UX quality: pass only if invisible. A normal user should keep using plain
`search`; they should not choose token modes, alias modes, or ranking knobs.

## Follow-Up

Search performed before closing `oc-1amj`:

- `bd search "lexical semantic recall fallback production regression" --status all`: no existing issue found.

Created follow-up:

- `oc-1pxu`: regression-gate lexical semantic recall fallback implementation.

## Compatibility

Existing behavior remains unchanged:

- `openclerk retrieval search` remains current SQLite FTS.
- `hybrid_retrieval_report` remains read-only decision support.
- No public JSON schema, storage schema, durable cache, or default ranking
  changed.
