# Web URL Stale Repair Candidate Comparison POC

## Status

Implemented candidate-comparison framing for `oc-81vp`.

This document compares candidate shapes for reducing web URL stale-repair
ceremony after `oc-qnwd`. It does not add runner actions, schemas, migrations,
storage behavior, public API behavior, product behavior, or shipped skill
behavior.

Governing evidence:

- [`docs/evals/results/ockp-high-touch-web-url-stale-repair-ceremony.md`](results/ockp-high-touch-web-url-stale-repair-ceremony.md)
- [`docs/architecture/web-url-stale-repair-ceremony-promotion-decision.md`](../architecture/web-url-stale-repair-ceremony-promotion-decision.md)
- [`docs/evals/results/ockp-web-url-intake-pressure.md`](results/ockp-web-url-intake-pressure.md)

## Candidate Workflows

| Candidate | Shape | Strength | Risk |
| --- | --- | --- | --- |
| Guidance-only repair | Keep existing `ingest_source_url`, document, and retrieval calls; repair skill or prompt guidance for stale-impact reporting. | No API or response change; preserves all current safety boundaries. | The `oc-qnwd` natural row already had correct durable evidence but failed answer/search ceremony, so guidance alone may keep the high-touch workflow. |
| Stale-impact response on update | Keep `openclerk document` `ingest_source_url` with `source.mode: "update"` and enrich the future update response with stale-impact details. | Keeps the natural existing action, reduces separate retrieval ceremony, and exposes stale dependent synthesis impact without auto-repairing synthesis. | Must not hide provenance/freshness behind a summary or silently authorize synthesis writes. |
| No new surface after prompt/harness repair | Treat `oc-qnwd` as one narrow answer-contract failure and defer until repeated evidence exists. | Avoids over-promoting from one natural ergonomics gap. | Leaves a real stale-repair UX need unresolved for normal users. |

## Selected Candidate

Select the stale-impact response candidate for future targeted evidence, not
implementation.

The candidate keeps the existing request surface:

```json
{"action":"ingest_source_url","source":{"url":"<public-web-url>","mode":"update","source_type":"web"}}
```

The candidate response should be evaluated for fields that make stale repair
impact visible without requiring a separate scripted retrieval sequence:

- update status: changed, unchanged/no-op, duplicate/conflict, or rejected
- normalized source URL and existing source document identity
- previous and new hash evidence when content changes
- dependent stale synthesis references, if any
- projection/provenance references sufficient for later inspection
- explicit no-repair warning that source refresh did not update synthesis

The response must not repair synthesis automatically. A synthesis repair
remains a separate durable write requiring current document/retrieval evidence
and approval where applicable.

## Evidence Scorecard

| Evidence | Safety | Capability | UX quality |
| --- | --- | --- | --- |
| `oc-qnwd` natural row | No bypass observed; runner-owned update path used. | Database evidence passed. | Failed with `ergonomics_gap`: 24 tools/commands, 6 assistant calls, 65.21s, and missing answer/search evidence. |
| `oc-qnwd` scripted control | Passed no-browser/no-manual and validation boundaries. | Passed with `none`: current primitives express duplicate/no-op, changed update, stale synthesis visibility, provenance, and freshness. | Still high ceremony: 26 tools/commands and 5 assistant calls for a routine stale-impact explanation. |
| Existing web URL intake pressure | Preserved public runner fetch, duplicate handling, update/no-op behavior, and unsupported acquisition rejection. | `ingest_source_url` already owns public web source create/update. | Changed-source stale impact remains the most ceremonial part of the workflow. |

## Conclusion

Do not file an implementation bead from this comparison. File targeted
eval/promotion evidence for the selected stale-impact response candidate.

The future eval should compare the selected response-enrichment candidate
against current primitives and guidance-only repair. Promotion remains blocked
until evidence shows the candidate reduces ceremony while preserving
runner-owned public fetch, normalized URL identity, duplicate/no-op handling,
stale synthesis visibility, provenance/freshness, local-first runner-only
access, approval-before-write, and rejection of browser/manual/private
acquisition paths.
