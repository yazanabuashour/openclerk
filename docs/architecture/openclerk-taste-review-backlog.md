# OpenClerk Taste Review Backlog

## Status

Planning backlog created after `oc-v1ed`.

This note records a process refinement after `oc-v1ed`, not a new public API.
It keeps the successful ADR, POC, eval, decision, and implementation workflow,
while adding a clearer taste review for cases where OpenClerk is technically
safe but unnecessarily awkward.

## Baseline Lesson

`oc-v1ed` is the new baseline for URL intake UX. Public HTML and web-page URLs
are now handled by the existing `openclerk document` `ingest_source_url`
surface instead of being treated as an adjacent unsupported input class.

The important distinction is the approval boundary:

- a user-provided public URL is enough permission for the runner to fetch and
  inspect that URL
- durable writes still require complete runner fields or an approved candidate
  workflow
- private access, account state, captcha, paywall access, purchase flows,
  browser automation, direct vault writes, and lower-level acquisition bypasses
  remain unsupported

Evidence:

- `docs/architecture/knowledge-configuration-v1-adr.md`
- `docs/evals/web-url-intake-pressure.md`
- `docs/evals/results/ockp-web-url-intake-pressure.md`
- `skills/openclerk/SKILL.md`

## Taste Review Lens

Future deferral or reference decisions should ask one more question after the
safety and capability checks: would a normal user reasonably expect a simpler
OpenClerk surface here?

Useful signals include:

- the workflow passes but needs many runner calls, assistant turns, or exact
  prompt choreography
- the user intent fits the natural scope of an existing action, but the current
  policy declares it unsupported
- the agent asks for approval before a read, fetch, or inspect step when the
  real approval boundary is a durable write, external egress, credentialed
  access, purchase, or irreversible mutation
- the result is safe but ceremonial, surprising, or hard to explain to a
  routine user

This lens does not weaken OpenClerk invariants. Authority, citations,
provenance, projection freshness, local-first operation, duplicate handling,
runner-only access, and approval-before-write still decide whether a smoother
surface is acceptable. Taste review can identify audit, design, or eval
backlog, but it does not authorize runner actions, storage migrations, schema
changes, public APIs, or skill behavior changes without targeted eval evidence
and an explicit promotion decision.

## Non-Promotion Follow-Up Loop

Taste debt and non-promotion outcomes are not dead ends when the user need is
real. A defer, keep-as-reference, or other non-promotion decision should state
whether there is no remaining need, or whether the evaluated shape failed while
the underlying OpenClerk need remains valid.

When the need remains valid, create or propose a Beads comparison epic before
handoff. The epic should normally include ADR, POC, Eval, and Decision children
and compare 2-3 plausible candidate surfaces unless the decision documents why
only one shape is viable. The Decision child must choose the best candidate,
combine useful behaviors where appropriate, defer or kill the track, or record
`none viable yet`.

This loop creates audit, design, and eval backlog only. It does not authorize
runner actions, storage migrations, schema changes, public APIs, skill behavior
changes, or durable writes. Candidate comparison must preserve authority,
citations, provenance, freshness, local-first operation, duplicate handling,
runner-only access, approval-before-write, public-source and synthetic-fixture
boundaries, and rejection of lower-level bypasses.

## Tracker Backlog

The following Beads epics track the revisit work:

- `oc-fbqy`: Re-audit post-oc-v1ed URL and artifact intake UX
- `oc-4rxs`: Re-audit path, title, and autofiling UX
- `oc-l6su`: Re-audit high-touch successful workflows
  ([audit](high-touch-successful-workflows-ux-audit.md))
- `oc-n959`: Update OpenClerk decision process for taste
- `oc-b2wr`: Audit closed historical non-promotion decisions
  ([audit](historical-non-promotion-follow-up-audit.md))

These epics are docs and evaluation-design backlog only. They do not authorize
runner actions, schema changes, storage migrations, skill behavior changes, or
implementation follow-up. Any future implementation still needs a targeted
eval report and an explicit promotion decision naming the exact surface and
gates.

## Initial Audit Targets

Re-audit URL and artifact intake after `oc-v1ed`. Treat public web URL intake
through `ingest_source_url` as the good baseline, then compare older artifact,
video, PDF source, and generalized ingestion decisions for remaining places
where adjacent user intent may have been held outside the natural runner
surface.

Re-audit path, title, and autofiling UX around `oc-iat`, `oc-99z`, `oc-9k3`,
document-this pressure, document artifact candidate generation, and the new web
`source.path_hint` behavior. The question is when OpenClerk should infer,
propose, or ask for path, title, body, and source hints.

Re-audit high-touch successful workflows where natural rows passed but stayed
expensive or ceremonial. Initial candidates include synthesis maintenance,
document lifecycle review and rollback, graph semantics, memory/router
revisit, promoted record lookup, document-this synthesis freshness, and the
`web-url-changed-stale` scenario. The `oc-l6su` audit is recorded in
[`high-touch-successful-workflows-ux-audit.md`](high-touch-successful-workflows-ux-audit.md).

Update process docs so future eval reports can separately record:

- safety pass: the workflow preserved invariants and rejected bypasses
- capability pass: current primitives can technically express the workflow
- UX quality: the workflow is or is not acceptable for routine use

Reports may mark safety pass and capability pass while still recording UX
quality as taste debt. Committed reports should keep repo-relative paths and
neutral placeholders such as `<run-root>`; raw logs and machine-local paths
remain out of committed artifacts.

The closed historical decision backfill is recorded in
[`historical-non-promotion-follow-up-audit.md`](historical-non-promotion-follow-up-audit.md).
It classifies past defer/reference/non-promotion outcomes against existing
coverage and found no new comparison epics were needed.
