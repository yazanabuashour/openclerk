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
- reduced sync diagnostics for scan, document writes, FTS strategy/write/rebuild
  timings, projection rebuilds, and no-op reopen behavior
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

## `oc-oa53.12.1` FTS Write-Path Evidence

`oc-oa53.12.1` kept the lexical SQLite FTS path and tuned the full-vault import
write shape. Full-vault sync now defers row-by-row `chunk_fts` writes, writes
documents and chunks first, then bulk-rebuilds `chunk_fts` once from `chunks`.
Single-document writes still use incremental FTS rows for immediate search
consistency, and a durable FTS rebuild-pending flag recovers interrupted full
imports.

The FTS-write-tuned 100 MB report is
[`docs/evals/results/ockp-scale-ladder-100mb-fts-write-tuned.md`](results/ockp-scale-ladder-100mb-fts-write-tuned.md).
It completed with import/sync 4.06s, reopen/no-op sync 0.30s, and FTS probes
0.25s total. Import diagnostics reported document writes 1.18s, bulk FTS rebuild
1.31s, projection rebuild 0.82s, and incremental FTS write time 0.00s.

The FTS-write-tuned 1 GB report is
[`docs/evals/results/ockp-scale-ladder-1gb-fts-write-tuned.md`](results/ockp-scale-ladder-1gb-fts-write-tuned.md).
It completed with import/sync 68.06s, reopen/no-op sync 5.97s, and FTS probes
4.40s total. Import diagnostics reported document writes 8.26s, bulk FTS rebuild
17.61s, projection rebuild 31.70s, and incremental FTS write time 0.00s.

Before/after posture:

| Area | Result |
| --- | --- |
| Safety | Pass: only reduced JSON/Markdown reports are committed; generated corpora, SQLite databases, raw logs, private content, and machine-absolute paths stay out of artifacts. |
| Capability | Pass: deterministic 100 MB and 1 GB corpora synced, reopened, searched, sampled projections, and sampled provenance through the embedded runtime. |
| UX quality | Not direct agent UX evidence; the 1 GB cold import is now below the 600s guardrail used for routine usability pressure. |
| Performance | Pass for slope evidence: 1 GB import/sync improved from 1657.81s to 68.06s, and the tuned 1 GB/100 MB import ratio is about 16.8x for about 10.2x more generated bytes, or about 1.6x byte-linear. |
| Evidence posture | Synthetic scale evidence only; reports expose reduced counters and timings without paths, titles, snippets, document ids, chunk ids, raw roots, or logs. |

This evidence does not justify a hybrid/vector track. The bottleneck was the
current lexical FTS write path, and current-path tuning reduced it enough for
continued SQLite FTS maturity work.
