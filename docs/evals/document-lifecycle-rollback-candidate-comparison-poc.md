# Document Lifecycle Rollback Candidate Comparison POC

## Status

Implemented candidate-comparison framing for `oc-awo6`.

This document compares candidate shapes for reducing document lifecycle
review and rollback ceremony after `oc-k8ba`. It does not add runner actions,
schemas, migrations, storage behavior, public API behavior, product behavior,
or shipped skill behavior.

Governing evidence:

- [`docs/evals/results/ockp-high-touch-document-lifecycle-ceremony.md`](results/ockp-high-touch-document-lifecycle-ceremony.md)
- [`docs/architecture/document-lifecycle-ceremony-promotion-decision.md`](../architecture/document-lifecycle-ceremony-promotion-decision.md)
- [`docs/evals/high-touch-document-lifecycle-ceremony.md`](high-touch-document-lifecycle-ceremony.md)
- [`docs/architecture/document-lifecycle-promotion-decision.md`](../architecture/document-lifecycle-promotion-decision.md)

## Candidate Workflows

| Candidate | Shape | Strength | Risk |
| --- | --- | --- | --- |
| Guidance-only repair | Keep existing `openclerk document` and `openclerk retrieval` calls; repair skill or prompt guidance for lifecycle review and rollback. | No API or response change; preserves all current safety boundaries. | The `oc-k8ba` natural row completed safely but still required 40 tool/command calls and 10 assistant calls, so guidance alone may preserve the high-touch ceremony. |
| Narrow lifecycle review/rollback candidate | Evaluate a future helper or report surface that packages target identity, source evidence, before/after summary, restore reason, provenance refs, projection freshness, privacy/no-diff boundaries, no-bypass boundaries, and write status. | Directly targets the natural user expectation for routine lifecycle repair while keeping safety evidence visible. | Must not hide authority, provenance, freshness, rollback target accuracy, privacy boundaries, or approval-before-write behind a convenient restore result. |
| No new surface after prompt or harness repair | Treat `oc-k8ba` as acceptable reference pressure because natural and scripted rows completed with classification `none`. | Avoids over-promoting from high command count alone. | Leaves a real UX need unresolved: normal users should not need a 40-step ceremony for routine lifecycle review and rollback. |

## Selected Candidate

Select the narrow lifecycle review/rollback candidate for future targeted
evidence, not implementation.

The future candidate should evaluate an explicit, evidence-visible workflow
shape such as:

```json
{
  "action": "review_lifecycle_rollback",
  "lifecycle": {
    "target_path": "notes/history-review/restore-target.md",
    "source_refs": ["sources/history-review/restore-authority.md"],
    "restore_section": "Summary",
    "restore_reason": "unsafe accepted lifecycle summary",
    "mode": "review_then_restore"
  }
}
```

The future response candidate should make the safety-critical evidence visible
without requiring a separate scripted retrieval sequence:

- selected target path and stable document identity
- source evidence and citation/source-ref paths
- before/after summary without raw private diffs
- restore reason and whether a durable write occurred
- provenance refs for the target document
- projection freshness for the target document after restore
- privacy/no-diff and no-bypass boundaries
- write status: restored, unchanged, rejected, ambiguous, or needs approval

The response must not make storage-level history a new OpenClerk authority.
Canonical markdown remains authoritative. The candidate must not authorize
direct file edits, direct SQLite, direct vault inspection, broad repo search,
source-built runners, HTTP/MCP bypasses, unsupported transports, backend
variants, module-cache inspection, raw private diff leakage, or automatic
approval of source-sensitive durable writes.

## Evidence Scorecard

| Evidence | Safety | Capability | UX quality |
| --- | --- | --- | --- |
| `oc-k8ba` natural row | Passed no-bypass controls, privacy-safe summary handling, source evidence, provenance, and projection freshness. | Passed with `none`: current primitives completed lifecycle review and rollback. | Completed but high-touch: 40 tools/commands, 10 assistant calls, and 42.29s. |
| `oc-k8ba` scripted control | Passed validation and no-bypass boundaries while preserving rollback target accuracy. | Passed with `none`: current primitives express search, list, get, targeted restore, provenance inspection, and projection freshness. | Still ceremonial: 22 tools/commands, 7 assistant calls, and 59.79s for a scripted control. |
| Prior document lifecycle decision | Kept source refs/citations, provenance, freshness, privacy, local-first operation, and no-bypass invariants. | Current primitives remained sufficient for history inspection, semantic diff review, restore/rollback, pending review, stale synthesis inspection, and validation handling. | Natural lifecycle row remained high-touch at 40 tools/commands, 6 assistant calls, and 76.40s. |

## Conclusion

Do not file an implementation bead from this comparison. File targeted
eval/promotion evidence for the selected narrow lifecycle review/rollback
candidate.

The future eval should compare the selected candidate against current
primitives and guidance-only repair. Promotion remains blocked until evidence
shows the candidate reduces ceremony while preserving canonical authority,
source refs or citations, provenance, projection freshness, rollback target
accuracy, privacy-safe summaries, no raw private diff leakage, local-first
runner-only access, approval-before-write, and validation controls.
