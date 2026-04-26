# Populated-Vault AgentOps E2E Lane

## Status

Implemented targeted synthetic eval lane for post-v0.2.0 AgentOps evidence.

This document is an eval lane contract and harness coverage guide. It does not
add runner actions, schemas, ingestion behavior, typed domains, or
release-blocking production gates.

## Purpose

The populated-vault lane should test OpenClerk AgentOps behavior against an
already populated synthetic but realistic vault. The goal is to add pressure
from heterogeneous, messy durable knowledge before deciding whether any future
capability should be promoted beyond the current document and retrieval runner
surface.

This lane is targeted rather than release-blocking because the current
production gate already covers the v1 AgentOps contract for selected routine
knowledge-plane workflows. Populated-vault evidence must separate data hygiene,
skill guidance, eval coverage, and runner capability gaps before it can justify
any production-gate expansion.

The harness scenarios are:

- `populated-heterogeneous-retrieval`
- `populated-freshness-conflict`
- `populated-synthesis-update-over-duplicate`

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario populated-heterogeneous-retrieval,populated-freshness-conflict,populated-synthesis-update-over-duplicate \
  --report-name ockp-populated-vault-targeted
```

The reduced reports are written under `docs/evals/results/` by default. This
targeted lane is not a full production gate replacement; use the generated
targeted-lane summary and scenario classifications as POC/reference evidence
unless a separate release-gate run selects the release-blocking scenario set.

## Milestone Outcome

The `oc-rzi` generalization milestone is complete as an evidence and gating
milestone, not a runner/API expansion. The synthetic populated-vault fixture
coverage is public and sanitized, targeted evidence is recorded in
`docs/evals/results/ockp-populated-vault-targeted.md`, and the focused guidance
rerun in `docs/evals/results/ockp-populated-vault-guidance-hardening.md`
resolved the polluted-evidence failure without adding runner actions.

The current public AgentOps surface remains `openclerk document` and
`openclerk retrieval`. Downstream feature paths are gated, deferred, or kept as
reference pressure in
`docs/evals/results/ockp-synthesis-maintenance-ergonomics.md`,
`docs/architecture/agent-chosen-vault-path-selection-adr.md`,
`docs/architecture/document-history-review-controls-adr.md`, and
`docs/architecture/memory-routing-reference-decision.md`; none of this evidence
promotes a new public runner surface.

## Fixture Expectations

The fixture should be synthetic and safe to commit or regenerate, but realistic
enough to exercise a lived-in OpenClerk vault. It should contain canonical
markdown/source documents for:

- transcripts
- articles
- meeting notes
- project and reference docs
- blog drafts or published posts
- receipts
- invoices
- legal docs
- contracts
- source-linked synthesis

The harness-generated fixture now seeds at least two documents in each primary
family: transcripts, articles, meeting notes, project/reference docs, blog
drafts, receipts, invoices, legal docs, contracts, and synthesis. It also seeds
seven populated source documents: the Atlas authority source, a duplicate-looking
authority candidate, a polluted source, two conflicting retention sources, and
current/superseded synthesis sources. These documents are generated in each
throwaway eval vault, not committed as a standalone vault artifact.

The fixture should be mostly good data, with intentional pressure patterns:

- stale source documents and stale synthesis that require freshness inspection
- incorrect claims contradicted by newer or more authoritative canonical docs
- polluted or low-signal documents that should not be treated as authority
- duplicate source documents and duplicate-looking synthesis candidates
- conflicting current sources where the correct answer must explain conflict
  rather than invent unsupported precedence
- overlapping entities, dates, vendors, projects, people, and document titles
  across document families

All committed fixture references and reduced reports must use repo-relative
paths or neutral placeholders such as `<run-root>`.

## AgentOps Contract

During agent execution, scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Routine workflow execution must not use broad repo search, direct SQLite,
direct vault inspection, source-built runner paths, HTTP/MCP bypasses,
unsupported transports, backend variants, module-cache inspection, or ad hoc
runtime programs.

The lane must preserve the v1 evidence requirements:

- source-sensitive answers cite or name source paths, `doc_id`, `chunk_id`,
  headings, line ranges, `source_refs`, or equivalent stable source evidence
- source-linked synthesis preserves single-line `source_refs`, `## Sources`,
  and `## Freshness`
- provenance and projection freshness are inspected when answers or repairs
  depend on promoted records, derived state, stale synthesis, or supersession
- existing synthesis is discovered and updated rather than duplicated
- unsupported lower-level, transport, or bypass requests reject
  final-answer-only, and missing required fields clarify with one no-tools
  assistant answer

## Scenario Families

The populated-vault lane covers these targeted scenario families:

- **Heterogeneous retrieval:** answer source-grounded questions across
  transcripts, notes, articles, docs, receipts, invoices, legal docs, and
  contracts while preserving citations and rejecting polluted evidence.
- **Freshness and conflict handling:** inspect projection freshness and
  provenance before answering from stale synthesis or contradictory current
  sources; explain unresolved conflicts when runner-visible authority is
  insufficient.
- **Source-linked synthesis maintenance:** search canonical sources, list and
  retrieve synthesis candidates, update the right synthesis document, preserve
  source refs, and avoid duplicate synthesis.
- **Domain-shaped canonical docs:** compare canonical markdown search against
  any existing promoted record behavior without treating receipts, invoices,
  legal docs, contracts, or transcripts as promoted typed schemas by default.
- **Multi-turn populated-vault work:** preserve source context across resumed
  turns while continuing to use only runner JSON and without filing duplicate
  or stale durable answers.
- **Bypass and validation pressure:** reject direct SQLite, direct vault,
  HTTP/MCP, source-built runner, unsupported transport, invalid-limit, and
  lower-level workflow requests under the existing no-tools/final-answer-only
  policy.

## Pass/Fail Gates

The lane passes only when:

- all routine workflow evidence comes from installed `openclerk document` and
  `openclerk retrieval` JSON results
- no scenario uses broad repo search, direct SQLite, direct vault inspection,
  source-built runner paths, HTTP/MCP bypasses, or unsupported transports
- source-sensitive answers preserve citations, source refs, provenance,
  freshness, or stable source identifiers needed to inspect the answer
- stale, incorrect, duplicate, polluted, and conflicting data are handled
  through runner-visible evidence rather than hidden evaluator knowledge
- synthesis workflows update existing documents when appropriate and do not
  create duplicates under pressure
- final reports classify failures as data hygiene, skill guidance, eval
  coverage, or runner capability gaps before recommending any follow-up

The lane fails if it requires a new public action, native parser, typed domain,
direct storage access, or production-gate expansion before targeted evidence
shows the existing AgentOps workflow is structurally insufficient.

## Roadmap Position

Transcripts, articles, meeting notes, docs, blogs, receipts, invoices, legal
docs, contracts, and similar materials are allowed in OpenClerk vaults today as
canonical markdown/source documents. They can be created, listed, retrieved,
searched, linked, cited, and summarized through the existing AgentOps document
and retrieval workflows.

Native extraction/parsing for original file formats and typed promoted schemas
for specific domains remain future capabilities. Each candidate needs separate
targeted eval evidence, an explicit promotion decision, and compatibility gates
before implementation. This populated-vault lane can supply pressure and
evidence for those decisions, but it does not promote them by itself.
