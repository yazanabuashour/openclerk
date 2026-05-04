# Representative Real-Vault Dogfood

## Purpose

This lane validates OpenClerk against a private or local representative vault
without committing private content. It should run before any 1 GB scale trial
or retrieval/indexing promotion decision.

The lane produces a reduced report only: counts, timings, probe outcomes,
safety/capability/UX placeholders, performance posture, and neutral artifact
references. It must not commit raw logs, private document paths, titles,
snippets, document ids, chunk ids, private query text, database paths, vault
roots, or machine-absolute paths.

## Representative Vault Dimensions

Choose a private/local vault that includes, or intentionally documents gaps for:

- source documents under source-like collections
- source-linked synthesis documents
- stale or superseded documents
- duplicate or duplicate-looking documents
- tags or metadata used for filtering
- promoted-record source material, including records and decisions where
  available
- provenance and projection freshness visible after runner sync
- public/private boundary cases that should remain local and reduced in reports

The committed report should describe only aggregate counts and behavior. Use
`<private-vault>` and `<run-root>` placeholders for artifact references.

## Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp maturity real-vault \
  --vault-root <private-vault> \
  --run-root <run-root> \
  --report-name ockp-real-vault-dogfood
```

Optional read probes can be supplied as comma-separated private queries. The
reduced report records them as `private-query-N`, not as query text:

```bash
mise exec -- go run ./scripts/agent-eval/ockp maturity real-vault \
  --vault-root <private-vault> \
  --run-root <run-root> \
  --query "<private-query-1>,<private-query-2>" \
  --report-name ockp-real-vault-dogfood
```

## Success Criteria

The report is useful only if it records:

- document, source, synthesis, decision, duplicate-marked, stale-marked, and
  tagged-document counts
- SQLite storage bytes
- initial import/sync time and optional reopen/rebuild time
- FTS search, list, get, synthesis projection, and provenance sample timings
- reduced-report safety: no private content, no raw logs, no machine-absolute
  artifact refs
- explicit evidence posture separating runtime behavior from routine-agent UX

This maintainer harness does not prove routine-agent no-bypass behavior by
itself because it does not run Codex event-log scenarios. If a decision claims
routine-agent safety, pair this report with an agent eval row or manual
event-log review that checks no direct SQLite, direct vault inspection,
source-built runner, HTTP/MCP bypass, broad repo search, or unsupported
transport was used.

## Decision Use

Use the real-vault report to decide whether current v1 surfaces are sufficient
for representative workflows. If the report shows capability, safety,
auditability, ergonomics, or workflow gaps that remain valid after taste
review, create candidate-comparison Beads before proposing any new public
runner action, schema, storage behavior, skill behavior, or retrieval backend.

The first-pass `oc-oa53` decision is
[`docs/architecture/openclerk-next-phase-maturity-validation-decision.md`](../architecture/openclerk-next-phase-maturity-validation-decision.md).
It uses the existing sanitized real-vault trial at
[`docs/evals/results/ockp-real-vault-agentops-trial.md`](results/ockp-real-vault-agentops-trial.md)
and does not promote any new real-vault workflow surface from that evidence.
