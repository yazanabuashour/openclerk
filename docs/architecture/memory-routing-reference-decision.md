# Memory And Routing Reference Decision

## Status

Kept as reference and deferred for `oc-jsg`.

## Decision

OpenClerk should keep memory and autonomous routing as benchmark pressure, not
as a production product surface, until targeted AgentOps eval evidence shows
that the existing `openclerk document` and `openclerk retrieval` workflows are
structurally insufficient.

This decision adds no public runner action, JSON schema, migration, storage API,
skill workflow, release artifact, memory-first `remember` or `recall` behavior,
autonomous router, or other public OpenClerk interface.

## Evidence

The relevant populated-vault evidence does not justify promotion:

- `docs/evals/results/ockp-populated-vault-targeted.md` classified the
  heterogeneous retrieval failure as skill guidance or eval coverage, not a
  runner capability gap. The agent used runner-visible evidence and avoided
  prohibited bypass paths, but repeated polluted decoy claims in the final
  answer.
- `docs/evals/results/ockp-populated-vault-guidance-hardening.md` reran the
  same pressure after skill guidance hardening and passed without adding runner
  actions, schemas, migrations, storage APIs, product behavior, or public
  interfaces.

This evidence supports the current AgentOps document and retrieval workflows.
It does not show that memory-first recall or autonomous routing would improve
correctness over source-grounded docs, provenance, projection freshness, and
explicit runner-action choice.

## Invariants

- Canonical docs and promoted canonical records outrank synthesis, memory,
  graph state, and routing choices.
- Memory remains recall, not authority, unless future canonicalization and
  provenance gates prove a narrower production surface.
- Routing remains deferred and explainable; it must not become a hidden
  classifier, opaque multi-store fanout, or bypass around the runner.
- Source-sensitive claims keep citations, source refs, provenance, or stable
  source identifiers attached.
- Routine agents continue to use the AgentOps surface instead of direct SQLite,
  HTTP, MCP, backend variants, module-cache inspection, source-built runner
  paths, or ad hoc runtime programs.

## Promotion Gate

Promotion requires repeated targeted AgentOps eval failures showing that the
current `openclerk document` and `openclerk retrieval` workflows are
structurally insufficient, not merely awkward, under-guided, missing data, or
thinly evaluated.

Any future promotion must first record a decision note with the exact proposed
surface, JSON request shape when applicable, backward-compatibility
expectations, failure modes, and eval gates. If the candidate makes memory
authority, bypasses provenance, or introduces autonomous routing without
explainable targeted evidence, it should be killed or kept as reference.
