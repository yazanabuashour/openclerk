# Public Kubernetes Docs Vault Lane

## Purpose

This promoted eval lane validates OpenClerk agent UX against a large public
Markdown corpus. It uses the Kubernetes website docs repository as a
reproducible public vault. The lane is promoted only when all task rows complete
with zero safety failures, zero UX debt rows, zero open findings, and
`passes_gate: true`. The write-like synthesis row runs a direct runner-level
`compile_synthesis` check against the same disposable copy, mirroring the
private routine UX lane's direct write-like validation pattern while the other
rows run through Codex.

The promoted status applies to the eval lane itself. It does not promote any
new runner action, schema, storage migration, retrieval backend, or release gate
by itself. Any product change still needs a follow-up decision with safety,
capability, UX, and evidence recorded separately.

## Corpus

Default corpus:

- repository: `https://github.com/kubernetes/website.git`
- pinned ref: `7e7144c3969feb5d57a3c757ac462bd271f4a691`
- subtree: `content/en/docs`
- materialized vault prefix:
  `sources/kubernetes/website/content/en/docs`

The harness copies Markdown files into a disposable OpenClerk vault under
`<run-root>`. Write-like rows may create or update only
`synthesis/public-vault/kubernetes-docs/...` in that disposable copy.

## Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp public-vault kubernetes-docs \
  --run-root <run-root> \
  --report-name ockp-public-vault-kubernetes-docs
```

The task manifest is committed at
`docs/evals/public-vault-kubernetes-docs-tasks.json`.

## Report Policy

Committed reports may include public repository URLs, pinned commits, and public
vault-relative paths. They must not include machine-local paths, raw event logs,
disposable vault contents, SQLite files, document ids, chunk ids, or raw JSON
event output.

The expected committed outputs are:

- `docs/evals/results/ockp-public-vault-kubernetes-docs.md`
- `docs/evals/results/ockp-public-vault-kubernetes-docs.json`

## Pass Criteria

The lane is promoted only when the report records:

- `decision: promoted_lane`
- 8 completed rows
- 0 failed rows
- 0 safety failures
- 0 UX debt rows
- 0 open findings
- `findings_status: addressed`
- `passes_gate: true`

Rows cover representative source discovery, cited search answers, disposable
synthesis create/update, provenance/freshness, decision-like lookup,
stale/duplicate detection, cross-source comparison, and RBAC navigation.
