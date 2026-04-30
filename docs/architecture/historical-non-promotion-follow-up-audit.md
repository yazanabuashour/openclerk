# Historical Non-Promotion Follow-Up Audit

## Status

Implemented for `oc-b2wr`.

This audit backfills the non-promotion follow-up loop across closed historical
OpenClerk decision and eval Beads. It is documentation and tracker hygiene
only. It does not authorize runner actions, schemas, storage changes, public
APIs, skill behavior, eval harness changes, or implementation follow-up.

The audit source of truth is closed Beads first, then the accepted architecture
and eval records they cite. Open Beads are used only as coverage targets to
avoid duplicate follow-up comparison epics.

## Method

The reviewed historical set includes closed decision, eval, and process Beads
with defer, keep-as-reference, reference-pressure, no-promotion, or partial
promotion outcomes. The key closed Beads reviewed were:

- `oc-iat`, `oc-99z`, and `oc-60s` for path/title, candidate intake, and
  propose-before-create policy
- `oc-no2` and `oc-oot` for artifact and video/source ingestion boundaries
- `oc-q7c`, `oc-vq9`, `oc-nqf`, `oc-7gp`, and `oc-bh4` for deferred
  high-touch or reference capabilities
- `oc-xh72` for explicit-overrides capture ceremony
- `oc-4rxs`, `oc-fbqy`, `oc-l6su`, and `oc-n959` for the completed taste
  process and re-audit epics
- `oc-o9p` for the broad contradiction/audit track, because it promoted a
  narrow surface while explicitly rejecting a broad semantic engine

Classifications use the current follow-up loop:

- `already covered`: an existing open or closed comparison/eval epic covers the
  remaining need
- `needs comparison epic`: a real unresolved need remains and no Bead covers
  candidate-surface comparison
- `no valid remaining need`: the deferred shape failed and no durable product,
  safety, workflow, ergonomics, or auditability need remains
- `superseded by later work`: later shipped behavior, accepted decisions, or
  completed Beads resolved or replaced the old concern

## Audit Findings

| Historical decision area | Closed Bead evidence | Accepted record evidence | Non-promotion outcome | Remaining need | Classification | Coverage or result |
| --- | --- | --- | --- | --- | --- | --- |
| Path/title autonomy | `oc-iat`, `oc-4rxs` | `agent-chosen-vault-path-selection-adr.md`, `path-title-autofiling-ux-audit.md`, `docs/evals/results/ockp-path-title-autonomy-pressure.md` | No constrained autonomy runner action, schema, storage migration, public API, direct-create behavior, or autonomous path/title policy. | Natural capture still needs smoother infer/propose/ask boundaries for path, title, body, and source hints. | `already covered` | Covered by closed audit `oc-4rxs`, closed capture epics `oc-hhap`, `oc-xh72`, `oc-yjuz`, `oc-xtbl`, and open `oc-3zd9`. |
| Agent-side intake and autofiling | `oc-99z`, `oc-60s`, `oc-9k3` | `agent-side-knowledge-intake-autofiling-adr.md`, `agent-chosen-document-artifact-candidate-generation-adr.md`, `docs/evals/results/ockp-document-artifact-candidate-ergonomics.md` | No runner API, schema, storage migration, public API, direct create, or autonomous autofiling; promotion is limited to the existing propose-before-create skill policy. | Propose-before-create policy became the accepted smoother surface for faithful candidate generation before write approval. | `superseded by later work` | Superseded by `oc-60s` and the implemented propose-before-create skill policy; later capture ceremony Beads cover remaining ergonomics. |
| Explicit overrides in smoother capture | `oc-xh72` | `capture-explicit-overrides-promotion-decision.md`, `docs/evals/results/ockp-capture-explicit-overrides.md` | Kept as reference pressure; no public explicit-overrides capture runner action, schema, storage, public API, skill behavior, or product behavior. | Explicit user path, title, type, and body instructions must continue to win unless validation or authority conflicts reject them. | `no valid remaining need` | The row records acceptable UX after taste review. No separate comparison epic is needed beyond preserving the explicit-override invariant in future smoother flows. |
| Generalized artifact ingestion | `oc-no2`, `oc-fbqy` | `generalized-artifact-ingestion-promotion-decision.md`, `post-oc-v1ed-url-artifact-intake-audit.md`, `docs/evals/results/ockp-heterogeneous-artifact-ingestion-pressure.md` | Keep heterogeneous artifact ingestion as reference pressure; defer generalized `ingest_artifact`, parser-backed ingestion, local file import, and unsupported-kind surfaces. | Local files, unsupported artifact kinds, richer public pages, and native media acquisition remain valid UX questions. | `already covered` | Covered by closed intake audit `oc-fbqy` and open URL/artifact epics `oc-0cme`, `oc-wqlb`, `oc-69h3`, and `oc-ijdk`. |
| Video and YouTube ingestion | `oc-oot`, `oc-fbqy` | `video-youtube-ingestion-promotion-decision.md`, `video-transcript-acquisition-design.md`, `docs/evals/results/ockp-video-youtube-canonical-source-note.md` | Promote only supplied-transcript `ingest_video_url`; defer media download, caption retrieval, local STT, transcript APIs, remote extraction, and richer timestamp-span citations. | Native transcript acquisition still needs dependency, privacy, provenance, citation, and update-policy comparison. | `already covered` | Covered by open `oc-69h3`; coordinated design remains in `video-transcript-acquisition-design.md`. |
| Document lifecycle controls | `oc-q7c`, `oc-l6su` | `document-lifecycle-promotion-decision.md`, `document-history-review-controls-adr.md`, `high-touch-successful-workflows-ux-audit.md`, `docs/evals/results/ockp-document-lifecycle-pressure.md` | Defer semantic document history, semantic diff, review queues, restore/rollback, storage migration, and new public lifecycle surfaces. | Natural lifecycle review and rollback can be safe but procedural and high-touch. | `already covered` | Covered by closed high-touch audit `oc-l6su` and open `oc-k8ba`. |
| Synthesis compile surface | `oc-vq9`, `oc-4qlx`, `oc-l6su` | `synthesis-compile-revisit-promotion-decision.md`, `synthesis-compile-revisit-adr.md`, `high-touch-successful-workflows-ux-audit.md`, `docs/evals/results/ockp-synthesis-compile-revisit-pressure.md` | Defer `compile_synthesis`; keep synthesis maintenance as current document/retrieval workflow. | Natural synthesis maintenance remains a high-step workflow. | `already covered` | Covered by closed high-touch audit `oc-l6su` and open `oc-7feg`. |
| Graph semantics | `oc-nqf`, `oc-l6su` | `graph-semantics-revisit-promotion-decision.md`, `graph-semantics-revisit-adr.md`, `high-touch-successful-workflows-ux-audit.md`, `docs/evals/results/ockp-graph-semantics-revisit-pressure.md` | Keep graph semantics as reference pressure; do not promote a semantic-label graph layer, graph semantics runner action, schema, storage behavior, or public API. | Relationship-shaped lookup remains valid UX pressure when the current path is slow or evidence-heavy. | `already covered` | Covered by closed high-touch audit `oc-l6su` and open `oc-oowv`, which combines relationship and promoted-record lookup ceremony. |
| Memory and autonomous router | `oc-7gp`, `oc-l6su` | `memory-router-revisit-promotion-decision.md`, `memory-router-revisit-adr.md`, `memory-routing-reference-decision.md`, `high-touch-successful-workflows-ux-audit.md`, `docs/evals/results/ockp-memory-router-revisit-pressure.md` | Keep as reference pressure; do not promote memory APIs, remember/recall actions, memory transports, autonomous router APIs, schemas, storage behavior, or public APIs. | Recall/routing comparison can remain too high-touch even when safe current primitives work. | `already covered` | Covered by closed high-touch audit `oc-l6su` and open `oc-nu12`. |
| Promoted record domain expansion | `oc-bh4`, `oc-l6su` | `promoted-record-domain-expansion-promotion-decision.md`, `promoted-record-domain-expansion-adr.md`, `high-touch-successful-workflows-ux-audit.md`, `docs/evals/results/ockp-promoted-record-domain-expansion-pressure.md` | Defer policy-specific or typed record-domain runner surfaces; no implementation follow-up. | Record and relationship lookup can be slow and evidence-heavy, but the evaluated typed-domain shape did not justify promotion. | `already covered` | Covered by closed high-touch audit `oc-l6su` and open `oc-oowv`. |
| Broad contradiction/audit engine | `oc-o9p`, `oc-nw7` | `broad-contradiction-audit-engine-revisit-promotion-decision.md`, `docs/evals/results/ockp-broad-contradiction-audit-revisit-pressure.md` | Do not promote a broad semantic contradiction truth engine; promote only the narrow `audit_contradictions` surface for source-linked audit repair and unresolved-conflict explanation. | The valid need was narrow audit repair, not a broad second truth system. | `superseded by later work` | Superseded by `oc-nw7`, which implemented the narrow surface and preserved the broader engine prohibition. |
| Public web URL intake boundary | `oc-v1ed`, `oc-fbqy` | `knowledge-configuration-v1-adr.md`, `post-oc-v1ed-url-artifact-intake-audit.md`, `docs/evals/results/ockp-web-url-intake-pressure.md` | The old unsupported-public-URL boundary was replaced; public HTML/web-page URLs belong under `ingest_source_url`. | Rich product pages and changed-source stale repair remain follow-up UX questions, not a reopening of the old public URL boundary. | `superseded by later work` | Superseded by `oc-v1ed`; richer pages are covered by open `oc-wqlb`, and stale repair ceremony is covered by open `oc-qnwd`. |

## Decision

No new comparison epics are required from this historical audit.

Every valid remaining need found in the closed historical decision set is
already covered by an existing closed re-audit, a current open eval/design
epic, or a later accepted implementation. Rows that are not covered represent
boundaries where the evaluated shape failed and the remaining broad need is not
valid for OpenClerk, such as autonomous hidden path/title behavior or broad
semantic truth engines.

Historical decisions remain true for their time. This audit adds classification
and coverage, but does not rewrite the accepted outcomes.

## Follow-Up

Continue the already-filed eval/design pipeline:

- path/title capture: `oc-hhap`, `oc-xh72`, `oc-yjuz`, `oc-xtbl`, `oc-3zd9`
- URL and artifact intake: `oc-0cme`, `oc-wqlb`, `oc-69h3`, `oc-ijdk`
- high-touch workflows: `oc-7feg`, `oc-k8ba`, `oc-oowv`, `oc-nu12`, `oc-qnwd`

Do not file implementation work from this audit. Future implementation still
requires targeted eval evidence and an accepted promotion decision naming the
exact OpenClerk surface and gates.
