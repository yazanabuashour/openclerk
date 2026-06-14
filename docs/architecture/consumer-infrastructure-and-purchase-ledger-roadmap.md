---
decision_id: roadmap-consumer-infrastructure-and-purchase-ledger
decision_title: Consumer Infrastructure And Purchase Ledger Roadmap
decision_status: accepted
decision_scope: consumer-infrastructure-and-receipts
decision_owner: platform
decision_date: 2026-06-14
source_refs: docs/architecture/agent-knowledge-plane.md, docs/architecture/chronicler-boundary.md, docs/architecture/structured-data-canonical-stores-adr.md, docs/architecture/generalized-artifact-ingestion-adr.md, docs/architecture/artifact-intake-autofiling-tags-fields-adr.md, docs/architecture/ocr-module-final-decision.md
---
# Roadmap: Consumer Infrastructure And Purchase Ledger

## Status

Accepted as roadmap documentation only.

This document does not add a runner action, module, schema, storage migration,
background worker, hosted service, connector, skill behavior, parser, or public
API. It records future direction and follow-up work.

## Existing Documentation

The infrastructure direction is partially documented already:

- [`agent-knowledge-plane.md`](agent-knowledge-plane.md) positions OpenClerk as
  the installed JSON runner plus thin skill guidance for agent workflows. It
  also states that OpenClerk is infrastructure for persistent
  agent-maintained knowledge.
- [`chronicler-boundary.md`](chronicler-boundary.md) records Chronicler as a
  first-party optional orchestration layer over Core, initially read-only and
  not a second authority system.
- [`structured-data-canonical-stores-adr.md`](structured-data-canonical-stores-adr.md)
  records the gate for future dense or correction-heavy structured domains.
- [`generalized-artifact-ingestion-adr.md`](generalized-artifact-ingestion-adr.md),
  [`artifact-intake-autofiling-tags-fields-adr.md`](artifact-intake-autofiling-tags-fields-adr.md),
  and [`ocr-module-final-decision.md`](ocr-module-final-decision.md) record the
  current artifact, OCR, invoice, and receipt boundaries.

What was missing was one explicit statement that OpenClerk is expected to be
consumed by automated and user-facing systems as infrastructure, rather than
only by a human driving one-off prompts. This document records that direction.

## Consumer Infrastructure Direction

OpenClerk Core should remain the local-first knowledge-plane runtime that
other surfaces consume through stable JSON request/response contracts. The
consumer may be an agent, a scheduled job, a desktop app, a CLI wrapper, a
human-facing workflow, or a first-party orchestration layer such as Chronicler.
The consumer should not become a privileged bypass around Core.

The durable authority model stays the same:

- canonical markdown and promoted canonical records remain the source of truth
- indexes, graphs, reports, extracted text, and module output remain derived or
  candidate evidence
- routine consumers invoke installed runner surfaces rather than inspecting
  SQLite, raw vault files, implementation internals, module caches, source-built
  runners, HTTP/MCP bypasses, or ad hoc scripts
- read/fetch/inspect permission is distinct from durable-write approval
- durable writes require the existing approved document or future promoted
  domain lifecycle, with citations, provenance, duplicate handling, projection
  freshness, and rollback/audit posture intact

Automation is appropriate for read-only planning and inspection: context
packs, inbox scans, duplicate reports, stale projection reports, source
authority checks, candidate extraction, and recommended next requests. Durable
mutation remains explicit and auditable. Human review can be one consumer of
that audit path, but a human should not have to be the main step-by-step driver
for every read-only planning pass.

The near-term integration contract is local process invocation of the installed
runner, not a hosted service. Future consumer work should compare exact
surfaces before implementation:

| Candidate | Fit | Boundary |
| --- | --- | --- |
| Direct installed runner invocation | Best current contract for agents, scripts, and local apps. | Keep JSON in/out, no direct storage access, no hidden writes. |
| First-party worker or Chronicler extension | Useful for scheduled read-only planning, inbox scans, context packs, and review queues. | Must compose Core runner actions and keep writes on approved document/domain APIs. |
| Integration envelope or event contract | Useful if user-facing systems need stable job ids, proposed writes, audit events, and retryable handoffs. | Should wrap runner results rather than create a second product authority. |
| Hosted HTTP server or multi-user service | Not the current direction. | Revisit only after local authority, review, and lifecycle contracts are mature. |

Follow-up work:

- `oc-ix56`: compare OpenClerk consumer integration surfaces.
- `oc-dcy2`: compare post-MVP Chronicler surfaces.

## Receipt And Invoice Purchase Tracking Direction

The desired capability is long-horizon tracking of many purchased items across
receipts, invoices, warranties, renewals, refunds, vendors, taxes, categories,
and source artifacts. The current product can store receipt and invoice notes
as canonical markdown, plan artifact candidates from explicit content, and use
the optional Tesseract OCR module for reviewed candidate text. It does not
claim structured receipt fields, invoice line-item authority, purchase-ledger
queries, or parser truth.

The best direction is a split design:

- **Extraction and normalization should start as optional modules.** Receipt
  and invoice extraction depends on OCR quality, layout models, vendor formats,
  currencies, taxes, line items, discounts, refunds, and confidence handling.
  Those provider and parser dependencies belong behind manifest-verified,
  pluggable modules that emit candidate evidence only.
- **The durable purchase ledger should not be module-only if promoted.** A
  long-lived purchase history needs stable identity, duplicate handling,
  correction and delete lifecycle, provenance, projection freshness, local-first
  storage posture, and auditable approved writes. If evidence proves markdown
  and derived records are insufficient, the ledger should become a promoted
  OpenClerk domain in Core rather than hidden state owned by an extractor
  module.

The likely promoted shape, if evidence supports it, is therefore:

1. Current artifact/OCR planning produces reviewed receipt or invoice text.
2. Optional receipt/invoice extractor modules propose structured purchase
   candidates with source refs, confidence, warnings, and no durable writes.
3. A Core purchase-ledger domain accepts approved candidates through exact
   runner actions with validation, correction, duplicate handling, provenance,
   freshness, and local-first storage.

Candidate comparison frame:

| Candidate | Safety | Capability | UX quality | Roadmap posture |
| --- | --- | --- | --- | --- |
| Current markdown plus OCR candidate planning | Strong current boundary. | Works for low-volume notes and reviewed text. | Too manual for large purchase history. | Keep as baseline. |
| Optional receipt/invoice extractor module only | Good for parser churn if no durable writes occur. | Can recover vendors, dates, totals, line items, and confidence from artifacts. | Useful but incomplete for long-term tracking. | Use for candidate extraction, not durable ledger authority. |
| Core promoted purchase-ledger domain plus optional extractor modules | Strongest if approval, validation, correction, provenance, and freshness are exact. | Best fit for high-volume, long-horizon item tracking. | Likely simplest normal-user surface after promotion evidence. | Preferred future candidate to compare. |
| External finance or commerce connector | Unproven. | May help imports later. | Adds credentials, sync, provider policy, and privacy ceremony. | Defer until local-first ledger semantics are proven. |

Follow-up work:

- `oc-63pl`: compare receipt and invoice purchase-ledger surfaces.

## Promotion Gates

Receipt and invoice work should not be implemented before targeted evidence
separates safety, capability, and UX quality.

Safety pass requires:

- runner-only routine access
- no purchase, checkout, cart, account-state, login, paywall, captcha, or
  irreversible commerce actions
- local-first default behavior and explicit provider/egress policy for any
  non-local module
- candidate extraction provenance back to receipt, invoice, page, image, or
  reviewed text evidence
- approval before durable ledger writes
- correction, delete, duplicate, and audit lifecycle
- no parser, OCR, or module output becoming canonical without approval

Capability pass requires evidence that current canonical markdown,
`artifact_candidate_plan`, OCR review, generic records, and reports cannot
handle the purchase-history workflow at useful scale.

UX quality pass requires a normal user or consuming system to get a simpler
surface than manual prompt choreography while still seeing confidence,
warnings, source refs, duplicate risks, and next approved write requests.

## Non-Goals

This roadmap does not promote:

- automatic purchase actions or commerce automation
- private email, bank, store, or accounting imports
- hidden cloud OCR, hidden model egress, or unreviewed remote providers
- receipt/invoice field authority from OCR or parser output alone
- independent module-owned durable purchase state
- hosted OpenClerk service behavior
- direct SQLite, direct vault mutation, broad file inspection, source-built
  runner paths, HTTP/MCP bypasses, or ad hoc lower-level transports
