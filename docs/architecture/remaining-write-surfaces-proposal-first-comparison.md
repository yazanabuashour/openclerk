---
decision_id: decision-remaining-write-surfaces-proposal-first-comparison
decision_title: Remaining Write Surfaces Proposal-First Comparison
decision_status: accepted
decision_scope: agent-interaction-policy
decision_owner: platform
source_refs: docs/architecture/path-title-autofiling-ux-audit.md, docs/architecture/native-media-transcript-acquisition-candidate-comparison-decision.md, docs/architecture/synthesis-compile-revisit-promotion-decision.md, docs/architecture/git-lifecycle-version-control-promotion-decision.md, docs/architecture/document-lifecycle-ceremony-promotion-decision.md
---
# Decision: Remaining Write Surfaces Proposal-First Comparison

## Status

Accepted for `oc-11yz`: use the current-surface routing matrix for remaining
proposal-first write surfaces.

This decision does not add a runner action, schema, migration, storage behavior,
public API, direct-create behavior, automatic checkpoint, native transcript
acquisition, restore/rollback surface, or autonomous write. It records how the
proposal-first OpenClerk skill policy applies to remaining write-adjacent
surfaces after `oc-wm04`.

## Candidate Surfaces

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Current-surface routing matrix | Pass: keeps writes on existing approved actions and planning on read-only reports. | Pass: covers supplied transcripts, synthesis, update targeting, and checkpoint guidance with existing runner evidence. | Good enough: simpler than asking for every field, without hiding authority or approval. | Select and document. |
| Unified `intake_defaults_plan` runner action | Risky: too broad across media, synthesis, updates, and Git state. | Partial: would centralize defaults but blur domain-specific provenance and failure modes. | Convenient, but likely surprising. | Reject. |
| Per-surface planning helpers | Potentially safe if narrowly scoped. | Future-only; each would need targeted evidence and exact request/response contracts. | Useful only if repeated ceremony returns. | Defer until evidence recurs. |

## Selected Matrix

| Surface | Proposal-first behavior | Durable-write boundary |
| --- | --- | --- |
| Video transcript intake | Preserve supplied transcript text and provenance; infer filing only from supplied transcript content or explicit hints. | `ingest_video_url` remains supplied-transcript-only. Missing transcript text, native acquisition, downloader, STT, transcript APIs, and remote extraction stay unsupported. |
| Source-linked synthesis | Use runner-visible source refs, placement plans, provenance, and projection freshness to propose synthesis path/title/body facts when confidence is high. | Durable synthesis writes still use approved `compile_synthesis`, `append_document`, or `replace_section` flows with source refs and freshness preserved. |
| Existing-document updates | Use `duplicate_candidate_report`, search/list/get, provenance, and projection evidence to name likely targets and alternatives. | `append_document` and `replace_section` require approved target, heading/content, and update-vs-new choice. |
| Checkpoint guidance | Use `git_lifecycle_report` status/history for read-only storage context around approved writes. | Checkpoint mode remains explicit, local-only, disabled by default, path/message-gated, and never automatic. |

## Decision

Select the current-surface routing matrix and close the `oc-11yz` path.

The evaluated need is real: these surfaces still benefit from agent/OpenClerk
defaults and optional overrides. The right implementation is not a new broad
planner. The current OpenClerk runner already has the needed safe split:
read-only planning and evidence actions can propose candidates, while durable
actions require explicit approval and domain-specific authority.

Explicit user values still win unless invalid or in conflict with runner-visible
authority. Planning output remains non-authoritative until an approved runner
write creates or updates canonical markdown, source notes, synthesis, or local
checkpoint commits.

## Follow-Up Check

Before closing `oc-11yz`, work items search checked for existing active work:

- `follow-up list --status=open`: no open issues found
- `follow-up search "video"`: no issues found
- `follow-up search "synthesis"`: no issues found
- `follow-up search "document lifecycle ceremony"`: no issues found
- `follow-up search "checkpoint git lifecycle"`: no issues found

No follow-up work item is created from this comparison. The selected current-surface
matrix is implemented by the shipped skill policy and this decision record. A
future work item should be opened only if new targeted evidence shows repeated
capability or ergonomics gaps for one named surface while preserving
approval-before-write, runner-only access, explicit override precedence,
citations/source refs, provenance, and freshness.
