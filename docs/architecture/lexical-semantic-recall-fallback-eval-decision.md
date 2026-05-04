---
decision_id: decision-lexical-semantic-recall-fallback-eval
decision_title: Lexical Semantic-Recall Fallback Eval Decision
decision_status: accepted
decision_scope: semantic-recall-lexical-fallback
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/results/ockp-semantic-recall-lexical-fallback.md, docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md, docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md
---
# Decision: Lexical Semantic-Recall Fallback Eval

## Status

Accepted as eval evidence only: lexical fallback is worth a production-safe
design pass, but no default ranking change is authorized.

The `oc-o2r8` run added the eval-only `ockp semantic-recall` harness and
recorded reduced evidence in
[`docs/evals/results/ockp-semantic-recall-lexical-fallback.md`](../evals/results/ockp-semantic-recall-lexical-fallback.md).

Existing public behavior remains unchanged:

- `openclerk retrieval search` remains current SQLite FTS.
- `hybrid_retrieval_report` remains read-only decision support.
- No runner JSON schema, storage schema, public API, or default ranking changed.

## Decision

Select lexical fallback for a follow-up production design/regression pass.

Current lexical FTS still reproduced the prior semantic-recall gap: 0/8 hit@3
and 0.000 MRR on the reduced pressure set. The eval-only token-overlap
fallback reached 7/8 hit@3 and 0.833 MRR. The eval-only alias-assisted fallback
reached 8/8 hit@3 and 0.938 MRR, matching the prior provider-vector MRR on
this reduced corpus.

Do not promote either fallback directly. The token and alias paths were
maintainer-harness candidates, not production `search` behavior. The high raw
duplicate pressure and curated alias row both require source-sensitive
regression evidence before any production ranking change.

## Safety, Capability, UX

Safety pass: pass for eval evidence. The run used copied committed docs under
`<run-root>`, produced reduced reports, made no provider calls, created no
vectors, changed no production documents, and did not modify default search.

Capability pass: partial. The no-vector candidates recovered semantic-recall
rows that current FTS missed, but they have not been tested against exact
source lookup, path-prefix behavior, metadata/tag filters, duplicate
candidate workflows, larger-corpus scale, or source-sensitive audit tasks.

UX quality: pass for the direction, not for promotion. A no-vector fallback
would preserve the simple `search` surface if invisible and citation-bearing.
It would fail taste review if users had to manage modes, aliases, or ranking
knobs.

## Follow-Up

Searches performed before closing `oc-o2r8`:

- `bd search "lexical semantic recall production fallback" --status all`: no
  existing issue found.
- `bd search "semantic recall fallback regression" --status all`: no existing
  issue found.

Created follow-up Bead:

- `oc-1amj`: design a production-safe lexical semantic-recall fallback or
  reject it after regression evidence.

`oc-o2r8` can close as eval completed with outcome `candidate selected for
follow-up design`, not production promotion.

## Compatibility

Committed reports and docs use repo-relative paths only. The eval-only
fallback can guide future implementation, but canonical markdown and promoted
records remain authoritative, and default retrieval ranking remains current
SQLite FTS until a later promotion decision says otherwise.
