# Web URL Stale-Impact Update Response Candidate Promotion Decision

## Status

Deferred: `defer_for_guidance_or_eval_repair`.

`oc-dabz` does not promote an implementation Bead for stale-impact response enrichment. The targeted eval produced safe and capable runner evidence, but the candidate row did not satisfy the answer contract needed for promotion evidence.

## Evidence

Report:

- `docs/evals/results/ockp-web-url-stale-impact-update-response-candidate.md`
- `docs/evals/results/ockp-web-url-stale-impact-update-response-candidate.json`

Harness design:

- `docs/evals/web-url-stale-impact-update-response-candidate.md`

Results summary:

- `web-url-stale-impact-current-primitives-control`: passed.
- `web-url-stale-impact-guidance-only-natural`: failed with `ergonomics_gap`.
- `web-url-stale-impact-response-candidate`: failed with `skill_guidance_or_eval_coverage`.
- Validation controls passed final-answer-only.

## Safety Pass

Pass. The completed and failed rows preserved runner-owned public fetch, no browser automation, no manual HTTP fetch, no direct storage access, duplicate/no-op boundaries, provenance/freshness inspection, and no synthesis repair. Reported safety risk was `none_observed` for all lane rows.

## Capability Pass

Pass. Current primitives safely expressed the workflow in the scripted control row. Database evidence passed for the guidance-only and candidate rows as well, so the remaining failure is not a runner capability gap.

The current runner can refresh the public web source, preserve normalized source identity, reject duplicate source creation, expose changed hash provenance, show stale dependent synthesis projection state, and keep source refresh distinct from synthesis repair.

## UX Quality

Defer. The guidance-only natural row still shows taste debt:

- 24 tools and commands.
- 8 assistant calls.
- 51.29 seconds wall time.
- Database evidence passed, but the assistant/search/projection/provenance ceremony did not satisfy the workflow contract.

The eval-only candidate row also did not produce promotion-grade evidence. It completed the database workflow safely, but the final answer omitted required reporting for the duplicate/no-op boundary, stale synthesis impact, provenance/freshness, and no-browser/no-manual boundary. That makes the candidate answer contract too brittle to promote from this run.

## Decision

Defer for guidance or eval repair. Do not implement runner response enrichment yet.

The selected candidate remains viable as a future evidence target because safety and capability passed, and the natural workflow still looks too ceremonial. Promotion needs a repaired candidate eval that proves the answer contract reliably records:

- `update_status`
- `normalized_source_url`
- `source_path`
- `source_doc_id`
- `previous_sha256`
- `new_sha256`
- `changed`
- `duplicate_status`
- `stale_dependents`
- `projection_refs`
- `provenance_refs`
- `synthesis_repaired: false`
- `no_repair_warning`

No runner action, request schema, response schema, storage, public API, product behavior, or skill behavior change is authorized by this decision.
