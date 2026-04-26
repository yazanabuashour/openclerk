# OpenClerk Agent Eval

- Decision Bead: `oc-gfg`
- Lane: synthesis maintenance ergonomics
- Release blocking: `false`
- Raw logs: existing evidence only; no new eval run

## POC Synthesis Maintenance Ergonomics Decision

`oc-gfg` asked whether repeated source-linked synthesis maintenance is too
brittle through the current `openclerk document` and `openclerk retrieval`
workflows under populated-vault pressure.

Decision: defer product/API promotion and keep the current `openclerk document`
and `openclerk retrieval` actions as the sufficient public workflow for now.

This decision does not add a public runner action, schema, migration, storage
API, product behavior, or public OpenClerk interface. A future ergonomics action
must be justified by repeated targeted failures that show the existing document
and retrieval workflows are structurally insufficient despite correct use.

Evidence base:

| Evidence | Result | Relevance |
| --- | --- | --- |
| `docs/evals/results/ockp-populated-vault-targeted.md` | The heterogeneous retrieval miss was classified as skill guidance / eval coverage, while freshness-conflict and synthesis-update scenarios passed. | Populated-vault pressure did not show missed candidate discovery, duplicate synthesis, dropped source refs, or skipped freshness inspection caused by the runner surface. |
| `docs/evals/results/ockp-populated-vault-guidance-hardening.md` | The focused rerun passed after guidance hardening and did not add a runner action. | Supports the original classification as guidance/eval hardening rather than a product/API promotion trigger. |
| `docs/evals/results/ockp-synthesis-compiler-pressure.md` | Source-linked synthesis scenarios completed through existing document/retrieval actions, including stale repair, candidate pressure, source-set creation, and multi-turn drift repair. | Repeated synthesis maintenance pressure remained workable without a dedicated synthesis maintenance command. |

## Promotion Gate

Do not promote a new synthesis maintenance runner surface unless future targeted
eval evidence repeatedly shows at least one of these failures after correct use
of the documented AgentOps workflow:

- missed existing synthesis candidates
- duplicate synthesis creation
- dropped or malformed `source_refs`
- skipped `projection_states` or `provenance_events` inspection when freshness
  or derivation history matters
- inability to preserve `## Sources` and `## Freshness` through
  `replace_section` or `append_document`

Failures caused by skill wording, eval prompt ambiguity, report interpretation,
or final-answer handling should continue to be handled as guidance or eval
coverage work, not as immediate product/API promotion.

