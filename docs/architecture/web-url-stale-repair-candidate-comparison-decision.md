---
decision_id: decision-web-url-stale-repair-candidate-comparison
decision_title: Web URL Stale Repair Candidate Comparison
decision_status: accepted
decision_scope: web-url-stale-repair
decision_owner: platform
---
# Decision: Web URL Stale Repair Candidate Comparison

## Status

Accepted: select a future stale-impact response candidate for targeted
promotion evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, or shipped
skill behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/web-url-stale-repair-candidate-comparison-poc.md`](../evals/web-url-stale-repair-candidate-comparison-poc.md)
- [`docs/architecture/web-url-stale-repair-ceremony-promotion-decision.md`](web-url-stale-repair-ceremony-promotion-decision.md)
- [`docs/evals/results/ockp-high-touch-web-url-stale-repair-ceremony.md`](../evals/results/ockp-high-touch-web-url-stale-repair-ceremony.md)
- [`docs/evals/results/ockp-web-url-intake-pressure.md`](../evals/results/ockp-web-url-intake-pressure.md)

## Decision

Select the candidate shape: enrich future `openclerk document`
`ingest_source_url` update responses with stale-impact details. Do not add a
new command and do not implement the candidate yet.

The selected future candidate keeps the request shape on the natural existing
surface:

```json
{"action":"ingest_source_url","source":{"url":"<public-web-url>","mode":"update","source_type":"web"}}
```

The future response candidate should expose update/no-op status, normalized
source URL identity, existing source document identity, previous and new hash
evidence when content changes, dependent stale synthesis refs, projection and
provenance refs, and an explicit warning that source refresh did not repair
synthesis.

Rejected alternatives:

- Guidance-only repair is too weak as the next step because the `oc-qnwd`
  natural row already produced correct durable evidence but failed the answer
  and search ceremony.
- No new surface is premature because the stale-repair need remains real and
  normal users would reasonably expect OpenClerk to explain dependent stale
  synthesis impact more directly after a source refresh.

## Safety, Capability, UX

Safety pass: pass. Existing evidence preserved runner-owned public fetch,
duplicate/no-op handling, provenance/freshness, no browser or manual
acquisition, local-first runner-only access, and approval-before-write. The
selected candidate must keep private URLs, account state, captcha, paywalls,
cart or purchase actions, browser automation, manual fetches, direct vault
inspection, direct SQLite, HTTP/MCP bypasses, source-built runners, and
unsupported transports outside the routine workflow.

Capability pass: pass for current primitives. The `oc-qnwd` scripted control
completed with classification `none`, and current `ingest_source_url`,
document, and retrieval primitives can express the workflow safely.

UX quality: not acceptable enough to stop at reference pressure. The natural
row failed with classification `ergonomics_gap` despite correct database
evidence. It used 24 tools/commands, 6 assistant calls, and 65.21 wall seconds,
and missed required search plus final-answer evidence for changed update,
stale synthesis impact, provenance/freshness, and no-browser/no-manual
boundaries. The scripted control still required 26 tools/commands and 5
assistant calls.

## Follow-Up

File one follow-up Bead for targeted eval and promotion evidence for the
selected stale-impact response candidate. Do not file an implementation Bead.

The follow-up must compare the candidate against current primitives and
guidance-only repair, then either promote an exact response contract, defer,
kill, or record `none viable yet`. Any later promotion decision must name the
exact response fields, compatibility expectations, failure modes, and gates.

## Compatibility

Existing behavior remains unchanged:

- `ingest_source_url` remains the public web source create/update primitive.
- Source refresh remains distinct from dependent synthesis repair.
- Dependent synthesis repair remains a separate durable write workflow.
- Existing duplicate, conflict, no-op, unsupported content, and validation
  behavior remains valid.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
