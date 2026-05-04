---
decision_id: decision-post-oc-uj2y-other-deferred-capability-triage
decision_status: accepted
decision_scope: post-oc-uj2y-other-deferred-capability-triage
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/deferred-capability-promotion-gates.md, docs/architecture/historical-non-promotion-follow-up-audit.md, docs/architecture/openclerk-next-phase-maturity-validation-decision.md
---

# Post-oc-uj2y Other Deferred Capability Triage

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

This triage covers deferred areas outside the named `oc-tnnw` tracks:
autonomous router behavior, semantic graph layers as truth, broad contradiction
engines, lifecycle review controls, and remaining public runner action
candidates. It does not authorize implementation.

## Candidate Areas

| Area | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Autonomous router behavior | Fails if routing hides canonical authority or stale state. | Not needed after read-only memory/report surfaces. | Too surprising for routine AgentOps. | Kill as product behavior; keep routing rationale in reports. |
| Semantic graph layer as truth | Fails because graph state is derived and cannot outrank markdown or promoted records. | Useful for navigation only. | Risky if presented as memory/truth. | Kill as truth layer. |
| Broad contradiction engine | Broad engine remains unsafe as second truth; narrow source-linked audit repair is already promoted elsewhere. | Narrow audit repair is covered; broad semantic detection is not needed. | Broad engine would be hard to audit. | Superseded by narrow audit surfaces. |
| Lifecycle review controls | Git status/history/checkpoint is covered; destructive restore remains outside this track. | Current document primitives and `git_lifecycle_report` cover the post-`oc-uj2y` need. | Further review UX can be revisited only with new evidence. | No new bead. |
| Remaining public runner action candidates | New actions require capability or ergonomics evidence under the promotion gates. | Existing current needs are represented by linked beads. | Avoid skill or API growth without evidence. | Link existing follow-ups only. |

## Decision

Record `none viable yet` for broad autonomous router, semantic graph truth, and
broad contradiction-engine product behavior. The valid remaining deferred needs
are already represented by concrete follow-up beads:

- `oc-9ijx`: local-first hybrid retrieval implementation comparison after
  real lexical versus hybrid/vector semantic recall evidence.
- `oc-w7xa`: parser-backed local artifact and OCR ingestion comparison,
  deferred until 2026-06-04.
- `oc-rcfv`: memory write transport comparison, deferred until 2026-06-04.

No additional candidate-comparison beads are created by this triage.

## Safety Pass

Pass. The triage rejects second-truth systems, hidden ranking, autonomous
durable writes, destructive restore, direct storage access, and non-runner
bypasses. It preserves canonical markdown/promoted-record authority,
citations, provenance, freshness, duplicate handling, local-first behavior, and
approval-before-write.

## Capability Pass

Pass for triage. Existing promoted surfaces and linked follow-up beads cover
the remaining concrete post-`oc-uj2y` questions. No additional product behavior
is required from this track.

## UX Quality

Pass. The result avoids reopening broad abstract capabilities and leaves future
work as concrete comparison beads a normal maintainer can evaluate.

## Follow-up Beads

Created: none.

Linked existing:

- `oc-9ijx`
- `oc-w7xa`
- `oc-rcfv`

None required beyond those linked beads because the other deferred areas are
either killed as unsafe truth layers, superseded by promoted narrow surfaces,
or lack a valid current OpenClerk workflow need.
