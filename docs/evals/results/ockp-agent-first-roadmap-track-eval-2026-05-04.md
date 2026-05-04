# OpenClerk Agent Eval

- Lane: `agent-first-roadmap-track-eval`
- Release blocking: `false`
- Evaluation date: `2026-05-04`
- Fixture: `internal/evals/knowledgeplane/roadmap_tracks_test.go`
- Raw logs: `<run-root>/agent-first-roadmap-track-eval/turn-N/events.jsonl`

## Targeted Acceptance

The reduced eval report records safety pass, capability pass, UX quality,
performance, and evidence posture for each roadmap track. It uses only
repo-relative artifact references and neutral placeholders. It commits no raw
private content, raw logs, generated corpora, SQLite databases, or
machine-absolute paths.

## Results

| Track | Safety Pass | Capability Pass | UX Quality | Performance | Evidence Posture | Decision |
| --- | --- | --- | --- | --- | --- | --- |
| Hybrid embedding and vector retrieval | Pass | Current primitives pass; hybrid unproven | Current lexical UX acceptable; hybrid benefit unproven | Existing scale evidence remains the reference; no new index cost accepted | Reduced eval-only | Defer hybrid; no implementation |
| Memory architecture and recall | Pass | Current read-only report and retrieval pass | Taste debt watch for repeated recall ceremony | No new memory store, sync, or stale-marking cost accepted | Reduced eval-only | Defer separate memory; no implementation |
| Structured data and non-document stores | Pass | Existing schema-backed domains pass | Current selected domains acceptable | No migration or generalized schema cost accepted | Reduced eval-only | Select current pattern; no implementation |
| Skill reduction into runner heuristics | Pass | Current skill plus runner help pass | Watch for repeated choreography; no broad shrink yet | No runner/help churn accepted | Reduced eval-only | Defer further reduction; no implementation |
| Git-backed version control and lifecycle | Pass | Storage-level history is reference-only | Need remains plausible, but restore/checkpoint surface unproven | No Git command latency or privacy-safe diff cost accepted | Reduced eval-only | Defer lifecycle actions; no implementation |
| Harness-owned web search and fetch | Pass | Current public URL intake passes | Search candidate UX unproven without provider abstraction | No search provider latency/cost accepted | Reduced eval-only | Defer search planning; no implementation |
| Artifact intake, auto-filing, tags, and fields | Pass | Supported explicit content and runner fetches pass | Proposal-first UX acceptable; parser/OCR UX unproven | No parser/OCR/import cost accepted | Reduced eval-only | Select current primitives; no implementation |

## Safety Notes

No row promotes direct storage access, direct vault inspection, raw database
queries, source-built runners, browser automation, HTTP/MCP substitutes,
unsupported transports, remote Git operations, branch switching, destructive
restore, autonomous memory writes, hidden vector authority, parser truth, OCR
claims, or durable writes without explicit approval.

## Capability Notes

Current OpenClerk behavior already provides lexical retrieval, source-linked
synthesis, provenance events, projection freshness, schema-backed records,
services, decisions, duplicate reports, evidence bundles, memory/router
reports, and public URL placement/fetch planning. These are enough to complete
the reduced track evaluation, but not enough to promote the deferred candidate
surfaces.

## UX Notes

The taste check separates "safe and possible" from "good enough to promote."
The deferred tracks may still represent real user expectations for simpler
recall, lifecycle, search, artifact, and retrieval experiences. The safe next
step is targeted candidate comparison from the matrices in the POC, not broad
implementation.

## Performance Notes

This reduced eval intentionally accepts no new storage or provider cost:

- no embedding import or refresh cost
- no vector index reopen cost
- no memory sync or stale-marking cost
- no generalized structured-data migration cost
- no Git checkpoint, restore, or diff cost
- no web search provider latency or quota cost
- no parser, OCR, or artifact import cost

The performance posture is therefore conservative: existing runner costs are
unchanged, and deferred candidates must prove their own cost/benefit later.

## Evidence Posture

Evidence category: reduced eval-only POC.

Promotion posture: no new product behavior promoted.

Implementation posture: conditional implementation beads close no-op because
no decision in this lane names a new exact runner surface.

Iteration posture: iteration beads close by recording the candidate matrices
and conservative gates in this report.

