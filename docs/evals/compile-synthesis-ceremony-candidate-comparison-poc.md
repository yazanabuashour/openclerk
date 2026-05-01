# Compile Synthesis Ceremony Candidate Comparison POC

## Status

Implemented candidate-comparison framing for `oc-zu6y`.

This document compares candidate shapes for reducing compile synthesis
ceremony after `oc-7feg`. It does not add runner actions, schemas, migrations,
storage behavior, public API behavior, product behavior, or shipped skill
behavior.

Governing evidence:

- [`docs/evals/results/ockp-high-touch-compile-synthesis-ceremony.md`](results/ockp-high-touch-compile-synthesis-ceremony.md)
- [`docs/architecture/compile-synthesis-ceremony-promotion-decision.md`](../architecture/compile-synthesis-ceremony-promotion-decision.md)
- [`docs/evals/results/ockp-synthesis-compile-revisit-pressure.md`](results/ockp-synthesis-compile-revisit-pressure.md)
- [`docs/evals/synthesis-compile-revisit-comparison-poc.md`](synthesis-compile-revisit-comparison-poc.md)

## Candidate Workflows

| Candidate | Shape | Strength | Risk |
| --- | --- | --- | --- |
| Guidance-only repair | Keep existing `openclerk document` and `openclerk retrieval` calls; repair skill or prompt guidance for synthesis maintenance. | No API or response change; preserves all current safety boundaries. | The `oc-7feg` natural row completed but still required 42 tool/command calls, so guidance alone may preserve the high-touch ceremony. |
| Narrow `compile_synthesis` candidate | Evaluate a future helper/report surface that packages candidate selection, source refs, provenance, projection freshness, duplicate behavior, update mode, and write status. | Directly targets the natural user expectation for source-backed synthesis maintenance while making safety evidence visible. | Must not hide source authority, provenance, freshness, duplicate checks, or approval-before-write behind a convenient write result. |
| No new surface after prompt or harness repair | Treat `oc-7feg` as acceptable reference pressure because both natural and scripted rows completed with classification `none`. | Avoids over-promoting from high command count alone. | Leaves a real UX need unresolved: normal users should not need a 42-step ceremony for routine synthesis maintenance. |

## Selected Candidate

Select the narrow `compile_synthesis` candidate for future targeted evidence,
not implementation.

The candidate should evaluate the existing deferred request shape:

```json
{
  "action": "compile_synthesis",
  "synthesis": {
    "path": "synthesis/example.md",
    "title": "Example",
    "source_refs": ["sources/source-a.md", "sources/source-b.md"],
    "body": "# Example\n\n## Summary\n...\n\n## Sources\n...\n\n## Freshness\n...",
    "mode": "create_or_update"
  }
}
```

The future response candidate should make the safety-critical evidence visible
without requiring a separate scripted retrieval sequence:

- selected synthesis path and whether it updated an existing candidate
- source evidence and normalized single-line `source_refs`
- candidate or duplicate status, including decoy avoidance
- provenance references and projection freshness
- write status: created, updated, appended, unchanged, rejected, or ambiguous
- no-bypass and no-automatic-authority escalation boundaries

The response must not make synthesis higher authority than canonical source
docs or promoted records. It also must not authorize lower-level storage access,
direct file edits, direct SQLite, broad repo search, HTTP/MCP bypasses,
unsupported transports, source-built runners, or unsupported write actions.

## Evidence Scorecard

| Evidence | Safety | Capability | UX quality |
| --- | --- | --- | --- |
| `oc-7feg` natural row | Passed no-bypass controls and preserved source authority, source refs, provenance/freshness, and duplicate prevention. | Passed with `none`: current primitives completed source-backed synthesis maintenance. | Completed but high-touch: 42 tools/commands, 8 assistant calls, and 53.08s. |
| `oc-7feg` scripted control | Passed validation and no-bypass boundaries. | Passed with `none`: current primitives express search, candidate list, get, projection/provenance inspection, and update. | Still ceremonial: 18 tools/commands and 6 assistant calls for a scripted control. |
| Prior synthesis revisit pressure | Preserved candidate selection, single-line `source_refs`, current/superseded source status, `## Sources`, `## Freshness`, and no-bypass boundaries. | Passed with `none` for natural and scripted rows. | Natural row remained high-touch: 34 tools/commands, 12 assistant calls, and 105.24s. |

## Conclusion

Do not file an implementation bead from this comparison. File targeted
eval/promotion evidence for the selected narrow `compile_synthesis` candidate.

The future eval should compare the selected candidate against current
primitives and guidance-only repair. Promotion remains blocked until evidence
shows the candidate reduces ceremony while preserving source authority,
citations or source paths, single-line `source_refs`, provenance, projection
freshness, duplicate prevention, local-first runner-only access,
approval-before-write, and validation controls.
