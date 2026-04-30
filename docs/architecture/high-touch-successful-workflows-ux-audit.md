# High-Touch Successful Workflows UX Audit

## Status

Implemented for `oc-l6su`.

This audit is documentation and eval-design work only. It does not authorize
runner actions, schemas, storage changes, public APIs, skill behavior, eval
harness changes, or implementation follow-up. It re-reads successful OpenClerk
eval rows where safety and capability passed, but the completed workflow still
looked expensive, ceremonial, or dependent on exact prompt shape.

## Taste Flags

These flags supplement, but do not replace, the deferred-capability gates:

- **High step count:** the row completes only after many runner calls or command
  executions.
- **Long latency:** the row completes but takes enough wall time that routine
  use would feel slow.
- **Repeated assistant turns:** the row needs many assistant calls to finish an
  otherwise natural user request.
- **Exact prompt choreography:** scripted controls pass because the prompt
  spells out the runner sequence or answer contract.
- **Surprising ceremony:** the user intent is natural, but the path to complete
  it requires several inspect, freshness, provenance, or duplicate checks that a
  normal user would not expect to name.
- **Freshness/provenance boilerplate:** the workflow repeatedly succeeds by
  restating projection freshness, source refs, citations, provenance, and
  no-bypass boundaries rather than through a simpler OpenClerk surface.

These flags do not weaken OpenClerk invariants. Authority, citations,
provenance, projection freshness, duplicate handling, local-first operation,
runner-only access, and approval-before-write still decide whether a smoother
surface can be considered. A successful but high-touch row can justify more
eval/design work, not implementation.

## Audit Findings

| Workflow | Evidence report | Scenario | Tools / commands | Assistant calls | Wall time | Safety pass | Capability pass | UX quality | Taste-debt classification |
| --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- |
| Compile synthesis | [`docs/evals/results/ockp-synthesis-compile-revisit-pressure.md`](../evals/results/ockp-synthesis-compile-revisit-pressure.md) | `synthesis-compile-natural-intent` | 34 / 34 | 12 | 105.24s | Pass. Source authority, source refs, freshness, provenance posture, duplicate prevention, and no-bypass boundaries were preserved. | Pass. Existing `openclerk document` and `openclerk retrieval` primitives completed the workflow. | Taste debt. Natural intent completed, but the row was slow and high-step with repeated assistant turns. | Strong candidate for eval/design revisit. |
| Document lifecycle review/rollback | [`docs/evals/results/ockp-document-lifecycle-pressure.md`](../evals/results/ockp-document-lifecycle-pressure.md) | `document-lifecycle-natural-intent` | 40 / 40 | 6 | 76.40s | Pass. Canonical authority, source refs, provenance, projection freshness, privacy, local-first operation, and bypass boundaries were preserved. | Pass. Current document/retrieval primitives expressed lifecycle review and rollback pressure. | Taste debt. The current path is safe but highly procedural for a natural lifecycle request. | Strong candidate for eval/design revisit. |
| Relationship-shaped graph semantics | [`docs/evals/results/ockp-graph-semantics-revisit-pressure.md`](../evals/results/ockp-graph-semantics-revisit-pressure.md) | `graph-semantics-revisit-natural-intent` | 28 / 28 | 5 | 99.11s | Pass. Canonical markdown relationship authority, citations, graph projection freshness, and bypass boundaries were preserved. | Pass. Search, document links, backlinks, and graph neighborhood evidence expressed the relationship workflow. | Taste debt. The row completed, but relationship lookup stayed slow and ceremonially evidence-heavy. | Combine with promoted-record lookup for a relationship/record eval-design revisit. |
| Memory/router-style recall | [`docs/evals/results/ockp-memory-router-revisit-pressure.md`](../evals/results/ockp-memory-router-revisit-pressure.md) | `memory-router-revisit-natural-intent` | 26 / 26 | 5 | 66.91s | Pass. Canonical memory/router authority, source refs, provenance, synthesis freshness, and bypass boundaries were preserved. | Pass. Current primitives answered the recall/routing comparison without memory transports or autonomous router APIs. | Taste debt. The user-facing recall question is natural, while the safe current path is still high-touch. | Strong candidate for eval/design revisit. |
| Promoted record domain lookup | [`docs/evals/results/ockp-promoted-record-domain-expansion-pressure.md`](../evals/results/ockp-promoted-record-domain-expansion-pressure.md) | `promoted-record-domain-expansion-natural-intent` | 36 / 36 | 5 | 114.40s | Pass. Canonical record authority, citations, provenance, records projection freshness, and bypass boundaries were preserved. | Pass. Generic records plus document/retrieval primitives expressed the policy-like lookup. | Taste debt. The workflow completed but was the slowest natural row in this audit and required many evidence checks. | Combine with graph semantics for a relationship/record eval-design revisit. |
| Web URL changed-source stale repair | [`docs/evals/results/ockp-web-url-intake-pressure.md`](../evals/results/ockp-web-url-intake-pressure.md) | `web-url-changed-stale` | 56 / 56 | 11 | 73.38s | Pass. Runner-owned public URL ingestion, update behavior, duplicate handling, stale synthesis visibility, provenance, and freshness boundaries were preserved. | Pass. Existing `ingest_source_url` update behavior expressed the changed-source workflow. | Taste debt. The safe update and stale-derived-state path is extremely high-step for routine URL refresh intent. | Strong candidate for eval/design revisit. |
| Document-this synthesis freshness | [`docs/evals/results/ockp-document-this-intake-pressure.md`](../evals/results/ockp-document-this-intake-pressure.md) | `document-this-synthesis-freshness` | 50 / 50 | 9 | 119.86s | Pass. The row completed with `none` failure classification under current document/retrieval behavior. | Pass for the selected pressure lane. | Taste debt. The row completed, but it was the highest-step and slowest workflow in this audit; the older report lacks the full ergonomics scorecard needed to rank it against the stronger candidates. | Monitor through future document-this or capture ergonomics work rather than filing a dedicated revisit from this audit. |

## Decision

The accepted defer/reference decisions remain sound on safety and capability.
The rows above do not prove a runner capability gap by themselves. They do show
that several technically successful workflows remain expensive enough to treat
as UX quality debt.

The strongest revisit candidates are:

- compile synthesis natural-intent ceremony
- document lifecycle review and rollback ceremony
- relationship and promoted-record lookup ceremony
- memory/router-style recall ceremony
- web URL changed-source stale repair ceremony

Each follow-up should stay eval/design scoped until targeted evidence and an
explicit promotion decision name an exact surface, request/response shape,
compatibility policy, failure modes, and gates. No implementation issue should
be filed directly from this audit.

## Follow-Up Beads

The follow-up beads filed from this audit are eval/design backlog. They should
ask whether a simpler OpenClerk surface is warranted while preserving current
authority, citations, provenance, freshness, local-first behavior, duplicate
handling, runner-only access, and approval-before-write.

| Bead | Flow | Eval-design question |
| --- | --- | --- |
| `oc-14gv` | Compile synthesis natural-intent ceremony | Is the current high-step synthesis maintenance path acceptable UX debt, or should refreshed pressure test a simpler surface? See [`high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md). |
| `oc-zjd3` | Document lifecycle review/rollback ceremony | Is lifecycle review and rollback too procedural for natural intent despite passing current safety and capability gates? See [`high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md). |
| `oc-nvub` | Relationship and promoted-record lookup ceremony | Should relationship-shaped graph semantics and promoted-record lookup remain separate reference pressure, or be evaluated together for a simpler lookup surface? See [`high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md). |
| `oc-j9yl` | Memory/router-style recall ceremony | Is safe recall/routing comparison still too high-touch without memory transports or autonomous router APIs? See [`high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md). |
| `oc-3bvy` | Web URL changed-source stale repair ceremony | Is changed-source stale repair too ceremonial despite passing runner-owned URL update and freshness boundaries? See [`high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md). |

Do not treat these follow-ups as implementation authorization. A future
implementation still requires targeted eval evidence and a promotion decision.
