---
decision_id: decision-post-oc-uj2y-deferred-pipeline-reconciliation
decision_status: accepted
decision_scope: post-oc-uj2y-deferred-pipeline-reconciliation
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/next-eval-candidate-pipeline.md, docs/architecture/openclerk-taste-review-backlog.md, docs/architecture/historical-non-promotion-follow-up-audit.md
---

# Post-oc-uj2y Deferred Pipeline Reconciliation

## Required References

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Scope

This reconciliation audits the deferred pipeline after the `oc-tnnw` tracks
promoted or rejected the current post-`oc-uj2y` surfaces. It is tracker and
decision hygiene only. It does not authorize runner actions, schemas, storage
changes, public APIs, skill behavior, or implementation work.

Reviewed sources:

- docs/architecture/next-eval-candidate-pipeline.md
- docs/architecture/openclerk-taste-review-backlog.md
- docs/architecture/historical-non-promotion-follow-up-audit.md
- current `oc-tnnw` track decisions for hybrid retrieval, canonical stores,
  Git lifecycle, web search planning, artifact/OCR ingestion, memory write
  transports, and artifact candidate planning

## Candidate Outcomes

| Candidate | Safety | Capability | UX quality | Decision |
| --- | --- | --- | --- | --- |
| Reopen every historical deferred track | Fails tracker hygiene by duplicating superseded or promoted work. | Low; many old needs are already covered. | Noisy and hard to act on. | Reject. |
| Close the pipeline with no follow-ups | Fails follow-up discipline because real deferred needs remain. | Incomplete. | Hides future work. | Reject. |
| Link only live remaining needs | Passes: preserves tracker evidence and avoids obsolete work. | Passes: each remaining need has a concrete comparison target. | Good: next sessions can start from specific beads. | Accept. |

## Reconciliation

Promoted or completed current surfaces:

- `hybrid_retrieval_report` covers current hybrid/vector decision support but
  not real semantic recall ranking.
- `structured_store_report` covers independent canonical-store decision
  support but not new domain-specific canonical stores.
- `git_lifecycle_report` covers status/history/explicit checkpoint behavior
  and leaves restore/rollback outside the promoted surface.
- `web_search_plan` covers harness-supplied search-result planning and leaves
  live provider adapters outside the promoted surface.
- `memory_router_recall_report` covers read-only memory recall and leaves
  memory write transports outside the promoted surface.
- `artifact_candidate_plan` covers read-only naming, tagging, filing, duplicate
  posture, confidence, and approved-write handoff.

Valid remaining needs are represented by linked beads:

- `oc-9ijx`: local-first hybrid retrieval implementation candidate comparison,
  deferred until 2026-06-04 after `oc-rlg7` and `oc-ye6w` found a real
  semantic recall gap.
- `oc-w7xa`: parser-backed local artifact and OCR ingestion candidate
  comparison, deferred until 2026-06-04.
- `oc-rcfv`: memory write transport candidate comparison, deferred until
  2026-06-04.

No additional Beads are required from this reconciliation. Historical pipeline
items that are already promoted, superseded, covered by accepted decisions, or
classified as no valid remaining need should not be recreated.

## Safety Pass

Pass. This reconciliation performs no product behavior. It preserves
runner-only access, local-first behavior, canonical markdown/promoted-record
authority, citations, provenance, freshness, duplicate handling, and
approval-before-write by linking only decision or comparison work.

## Capability Pass

Pass. The current deferred pipeline can be expressed as the three linked
follow-up beads above plus the closed `oc-tnnw` decision records. No direct
implementation, migration, or source-control work is needed here.

## UX Quality

Pass. The reconciliation avoids sending future agents through obsolete
historical backlog while preserving the remaining real user-facing questions:
semantic recall quality, parser-backed artifact/OCR ingestion, and memory write
transport shape.

## Follow-up Beads

Created: none.

Linked existing:

- `oc-9ijx`
- `oc-w7xa`
- `oc-rcfv`

None required beyond those linked beads because every other reviewed deferred
pipeline item is already promoted, superseded, covered by accepted decisions,
or has no valid remaining OpenClerk need.
