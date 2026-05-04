# Hybrid Retrieval Candidate Comparison Eval

Date: 2026-05-04

## Scenario

Evaluate `oc-uj2y.2` candidate surfaces for hybrid embedding/vector retrieval:
current lexical FTS, mode-flag-only search, durable local vector index,
external/hosted vector stores, and a read-only runner report.

Required references:

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Result

Promote `hybrid_retrieval_report` as a read-only decision-support runner
surface. Keep lexical FTS as the default retrieval behavior.

## Safety Pass

Pass.

The promoted report uses installed `openclerk retrieval` JSON, performs one
current lexical search, returns citation-bearing evidence, and does not write
documents, create embeddings, build vector indexes, call external APIs, use
direct SQLite, inspect raw vault files, invoke HTTP/MCP bypasses, or change
default ranking.

## Capability Pass

Partial pass for the selected surface.

The report satisfies the capability needed for this stage: it packages current
baseline evidence and candidate comparison in one runner action. It does not
claim semantic/vector recall; durable vector retrieval still needs a separate
index POC with freshness, provenance, citation, and scale checks.

## UX Quality

Pass.

Before this surface, agents had to run search and restate architecture policy
manually. The report gives a normal maintainer one action with `agent_handoff`
and clear boundaries. A user asking ordinary retrieval questions still uses
plain `search`.

## Performance

The selected report is bounded by one lexical search with caller-provided
`limit`. It adds no import job, no reopen/rebuild cost, no generated corpus,
and no remote embedding latency.

Durable vector retrieval remains gated on existing scale evidence under
`docs/evals/results/ockp-scale-ladder-100mb-fts-write-tuned.md` and
`docs/evals/results/ockp-scale-ladder-1gb-fts-write-tuned.md`, plus future
hybrid-specific recall and citation-regression data.

## Evidence Posture

The eval is reduced and deterministic. It records design evidence, runner
contract tests, and public reference posture. It intentionally omits raw logs,
private corpus examples, generated corpora, SQLite databases, and
machine-absolute paths.

## Decision

Promote `openclerk retrieval` `hybrid_retrieval_report`.

Do not promote a durable embedding store, vector DB, OpenAI vector-store
integration, hosted retrieval path, or default hybrid ranking yet.

## Closure

Remaining work is represented by linked beads:

- `oc-tnnw.1.4` promotion decision.
- `oc-tnnw.1.5` conditional implementation only if promoted.
- `oc-tnnw.1.6` iteration and follow-up bead creation.
