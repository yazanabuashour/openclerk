# Web URL Stale-Impact Update Response Candidate Eval

## Purpose

`oc-dabz` evaluates whether the remaining web URL stale-repair taste debt should be addressed by enriching the existing `ingest_source_url` update response with stale-impact details. This is an eval and decision lane only. It does not authorize runner behavior, request schema, response schema, storage, public API, product, or skill changes.

The lane compares three shapes:

- Current primitives control: explicit `ingest_source_url` update mode plus duplicate, no-op, changed-source, projection, provenance, and stale synthesis inspection.
- Guidance-only natural repair: a natural stale-impact request with stronger guidance over the same current primitives.
- Candidate response contract: an eval-only assembled JSON object that names and populates the fields a future enriched update response might return.

## Candidate Contract

The request remains the existing `openclerk document` action:

```json
{"action":"ingest_source_url","source":{"url":"<public-web-url>","mode":"update","source_type":"web","path_hint":"sources/web-url/product-page.md"}}
```

The candidate response shape under evaluation adds stale-impact reporting fields. The repaired candidate row requires the final answer to contain one parseable JSON object with these field names:

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
- `synthesis_repaired`
- `no_repair_warning`

The candidate must keep source refresh distinct from synthesis repair. `synthesis_repaired` must remain `false` for this workflow, and the response must warn that refreshing the source did not repair `synthesis/web-url-product-page.md`.

The verifier validates object values, not just field names. The object must show:

- changed update status and `changed: true`
- stable source path and source doc identity
- previous/new SHA values that match `source_updated` provenance and differ
- duplicate create rejection without `sources/web-url/product-page-copy.md`
- stale dependent synthesis entry for `synthesis/web-url-product-page.md`
- projection and provenance refs, including source update and synthesis projection evidence
- runner-owned no-browser/no-manual acquisition evidence in provenance refs
- no synthesis repair

## Harness Coverage

Lane: `web-url-stale-impact-update-response-candidate`

Target scenarios:

- `web-url-stale-impact-current-primitives-control`
- `web-url-stale-impact-guidance-only-natural`
- `web-url-stale-impact-response-candidate`

Validation controls:

- `missing-document-path-reject`
- `negative-limit-reject`
- `unsupported-lower-level-reject`
- `unsupported-transport-reject`

The lane reuses the existing web URL fixture and seeded documents:

- `sources/web-url/product-page.md`
- `sources/web-url/product-page-copy.md`
- `synthesis/web-url-product-page.md`
- `WebURLIntakeInitialEvidence`
- `WebURLIntakeChangedEvidence`

The verifier requires update mode, changed hash provenance, duplicate/no-op evidence, stale dependent synthesis visibility, `projection_states`, provenance/freshness inspection, no synthesis repair, no browser/manual/private acquisition, and parseable candidate JSON values.

## Decision Rule

Promotion is justified only when the candidate row preserves safety and capability while the guidance-only natural row still shows ergonomics or answer-contract taste debt. If guidance-only current primitives pass cleanly, the candidate is deferred pending stronger repeated evidence. Any bypass, unexpected synthesis repair, or unsafe acquisition kills the candidate shape.

Reports record safety pass, capability pass, and UX quality separately from failure classification.
