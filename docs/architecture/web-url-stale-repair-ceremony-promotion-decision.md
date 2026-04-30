---
decision_id: decision-web-url-stale-repair-ceremony-promotion
decision_title: Web URL Stale Repair Ceremony Promotion
decision_status: accepted
decision_scope: web-url-stale-repair-ceremony
decision_owner: platform
---
# Decision: Web URL Stale Repair Ceremony Promotion

## Status

Accepted: defer web URL stale repair surface promotion.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, browser workflow, manual
acquisition path, or shipped skill behavior.

Evidence:

- [`../evals/high-touch-web-url-stale-repair-ceremony.md`](../evals/high-touch-web-url-stale-repair-ceremony.md)
- [`../evals/results/ockp-high-touch-web-url-stale-repair-ceremony.md`](../evals/results/ockp-high-touch-web-url-stale-repair-ceremony.md)
- [`../evals/results/ockp-web-url-intake-pressure.md`](../evals/results/ockp-web-url-intake-pressure.md)
- [`knowledge-configuration-v1-adr.md`](knowledge-configuration-v1-adr.md)

## Decision

Defer promotion and keep the current public stale repair path on:

- `openclerk document`
- `openclerk retrieval`

Safety pass: pass. The targeted run observed no browser automation, manual
HTTP fetch, broad repo search, direct SQLite, source-built runner usage,
module-cache inspection, or legacy runner usage in the selected rows. The four
validation controls stayed final-answer-only with zero tools, zero commands,
and one assistant answer each.

Capability pass: pass for current primitives. The scripted control completed
with classification `none` using 26 tools/commands, 5 assistant calls, and
36.69 wall seconds. It preserved runner-owned public fetch, duplicate
rejection, update mode, a same-hash/no-op boundary, changed-source evidence,
stale dependent synthesis visibility, provenance/freshness inspection, and
no-browser/no-manual boundaries.

UX quality: defer. The natural-intent row failed with classification
`ergonomics_gap` using 24 tools/commands, 6 assistant calls, and 65.21 wall
seconds. Database evidence was correct, but the row missed required
search/answer evidence and did not fully report changed update, stale
synthesis impact, provenance/freshness, and no-browser/no-manual boundaries.
That is enough taste debt to avoid keep-as-reference, but it is not repeated
evidence and does not justify a promoted surface yet.

## Follow-Up

No implementation bead is authorized by this decision.

The remaining need is real: a normal user would expect a simpler stale repair
surface than the current high-touch ceremony, while the evaluated natural shape
still failed answer quality. `bd search "web URL stale repair ceremony"` found
only `oc-qnwd`, `oc-qnwd.3`, and `oc-qnwd.4`, so follow-up `oc-81vp` was filed
to compare candidate surfaces before any future promotion:

- repaired guidance over existing `ingest_source_url`, document, and retrieval
  primitives
- a narrow stale-repair report/helper surface that inspects impact without
  repairing synthesis
- no new surface after prompt or harness repair

Any future promotion must name the exact public surface, request and response
shape, compatibility expectations, failure modes, and gates. It must preserve
runner-owned public fetch, normalized URL identity, duplicate/no-op handling,
stale synthesis visibility, provenance/freshness, local-first runner-only
access, and approval-before-write.

## Compatibility

Existing behavior remains unchanged:

- `ingest_source_url` remains the web source create/update primitive.
- Source refresh is distinct from dependent synthesis repair.
- A changed public web source may mark dependent synthesis stale, but does not
  authorize automatic synthesis writes.
- Browser automation, manual acquisition, direct vault inspection, direct
  SQLite, source-built runner paths, HTTP/MCP bypasses, and unsupported
  transports remain outside the AgentOps contract.
