# Scale-Ladder Validation

## Purpose

This lane measures OpenClerk's current SQLite FTS and projection behavior at
deterministic synthetic corpus sizes before any retrieval/indexing promotion
decision. It validates 10 MB and 100 MB first. The 1 GB tier is optional and
requires explicit justification from smaller-tier reports.

The harness writes generated markdown and SQLite artifacts under `<run-root>`
only. Committed artifacts are reduced JSON/Markdown reports under
`docs/evals/results/`.

## Corpus Rules

The scale harness generates deterministic markdown from a seed. The default
seed is `53`.

Generated documents include:

- `sources/` source documents
- `synthesis/` synthesis documents with `source_refs` and freshness metadata
- decision-like markdown with decision frontmatter
- duplicate-marked documents
- stale/superseded documents
- tagged notes for filter pressure

The generator intentionally uses synthetic text only. It does not copy private
vault content, raw logs, or machine-local paths into committed reports.

## Run

Run the 10 MB tier:

```bash
mise exec -- go run ./scripts/agent-eval/ockp maturity scale-ladder \
  --tier 10mb \
  --seed 53 \
  --run-root <run-root> \
  --report-name ockp-scale-ladder-10mb
```

Run the 100 MB tier after the 10 MB report is reviewed:

```bash
mise exec -- go run ./scripts/agent-eval/ockp maturity scale-ladder \
  --tier 100mb \
  --seed 53 \
  --run-root <run-root> \
  --report-name ockp-scale-ladder-100mb
```

Run the 1 GB tier only if the 10 MB and 100 MB reports show the trial will
produce meaningful new evidence:

```bash
mise exec -- go run ./scripts/agent-eval/ockp maturity scale-ladder \
  --tier 1gb \
  --allow-1gb \
  --seed 53 \
  --run-root <run-root> \
  --report-name ockp-scale-ladder-1gb
```

For quick calibration tests, use `--target-bytes <bytes>` with any tier label.
Do not commit calibration reports unless they are explicitly part of an eval
decision.

## Metrics

Each reduced report records:

- generated corpus bytes and SQLite storage bytes
- document, source, synthesis, decision, duplicate-marked, stale-marked, and
  tagged-document counts
- generation time
- import/sync time
- reopen/rebuild time unless `--skip-reopen` is used
- reduced sync diagnostics for scan, document writes, FTS writes, projection
  rebuilds, and no-op reopen behavior
- FTS search timing and hit counts
- `list_documents` and `get_document` timing
- synthesis projection-state sample timing and count
- provenance sample timing and count

Reports intentionally omit document paths, titles, snippets, document ids,
chunk ids, generated vault roots, database paths, and machine-absolute paths.

## Decision Criteria

Use 10 MB and 100 MB results to decide whether the current lexical SQLite FTS
path is adequate.

Valid outcomes:

- keep lexical SQLite FTS
- tune current indexes or query behavior
- create hybrid/vector candidate-comparison Beads
- defer more scale evidence
- kill the scale track if it does not represent a valid OpenClerk need

Do not promote hybrid/vector retrieval solely because the corpus is large.
Promotion pressure requires observed lexical relevance, latency, workflow,
auditability, or UX failures that current tuning cannot reasonably address.

## Initial `oc-oa53` Evidence

The first 10 MB reduced report is
[`docs/evals/results/ockp-scale-ladder-10mb.md`](results/ockp-scale-ladder-10mb.md).
It completed with reduced-report safety checks passing and no raw generated
corpus, SQLite database, raw logs, or machine-absolute paths committed.

The first 100 MB attempts did not complete in-session. The reduced timeout
report is
[`docs/evals/results/ockp-scale-ladder-100mb-timeout.md`](results/ockp-scale-ladder-100mb-timeout.md).
Follow-up `oc-oa53.12` owns the 100 MB performance investigation and must
either produce a completed reduced 100 MB report or promote the timeout/stall
finding into a decision-quality performance diagnosis.

The first-pass decision is
[`docs/architecture/openclerk-next-phase-maturity-validation-decision.md`](../architecture/openclerk-next-phase-maturity-validation-decision.md):
do not run 1 GB yet, and diagnose or tune the current SQLite FTS/projection
path before considering hybrid/vector retrieval.

## `oc-oa53.12` Follow-Up Evidence

`oc-oa53.12` diagnosed the initial 100 MB timeout as a sync/projection
implementation issue rather than a retrieval-mode issue. Full-vault sync now
imports documents in batches, rebuilds projections once after import, skips
chunk/FTS writes for unchanged documents during reopen, and writes reduced
progress diagnostics during sync so interrupted runs still leave local
phase-level evidence.

The tuned 100 MB report is
[`docs/evals/results/ockp-scale-ladder-100mb.md`](results/ockp-scale-ladder-100mb.md).
It completed with import/sync 19.38s, reopen/no-op sync 0.39s, and FTS probes
0.28s total.

The 1 GB report is
[`docs/evals/results/ockp-scale-ladder-1gb.md`](results/ockp-scale-ladder-1gb.md).
It completed with import/sync 1657.81s, reopen/no-op sync 4.91s, and FTS probes
8.88s total. The 1 GB result should not be judged by a fixed 10-minute cutoff:
the completed reduced report shows the current path is functionally capable but
import-bound at this scale.
