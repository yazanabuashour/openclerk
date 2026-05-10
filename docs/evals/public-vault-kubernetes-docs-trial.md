# Public Corpus Vault Autonomy Lanes

## Purpose

These promoted eval lanes validate OpenClerk autonomy modes against public
corpora. The first lane uses the Kubernetes website docs repository as the
promoted public-vault baseline, the second uses the Go website docs as another
large technical corpus, and the third uses a public-domain Moby-Dick corpus as
a non-technical corpus. A lane is promoted only when all task rows complete with
zero safety failures, zero UX debt rows, zero open findings, and
`passes_gate: true`. The write-like synthesis row runs a direct runner-level
`compile_synthesis` check against the same disposable copy, mirroring the
private routine UX lane's direct write-like validation pattern while the other
rows run through Codex.

The promoted status applies to the eval lane itself. It does not promote any
new runner action, schema, storage migration, retrieval backend, or release gate
by itself. Any product change still needs a follow-up decision with safety,
capability, UX, and evidence recorded separately.

## Autonomy Modes

The harness accepts these explicit modes and records them in committed reports:

- `approval_mode`: `propose_only`, `approve_write`,
  `autonomous_disposable`, `autonomous_trusted`
- `drafting_mode`: `require_explicit_fields`, `suggest_fields`,
  `autonomous_fields`
- `write_target_mode`: `existing_only`, `create_or_update`,
  `create_allowed`
- `citation_mode`: `strict`, `balanced`, `lightweight`
- `privacy_mode`: `private_summary_only`, `allow_paths`, `allow_titles`,
  `allow_snippets`
- `audience_mode`: `technical`, `plain_language`, `executive_summary`

The promoted public-corpus lanes default to
`approval_mode=autonomous_disposable`,
`drafting_mode=autonomous_fields`,
`write_target_mode=create_or_update`, `citation_mode=balanced`,
`privacy_mode=allow_paths`, and `audience_mode=plain_language`.

## Corpora

Kubernetes baseline:

- repository: `https://github.com/kubernetes/website.git`
- pinned ref: `7e7144c3969feb5d57a3c757ac462bd271f4a691`
- subtree: `content/en/docs`
- materialized vault prefix:
  `sources/kubernetes/website/content/en/docs`

The harness copies Markdown files into a disposable OpenClerk vault under
`<run-root>`. Write-like rows may create or update only
`synthesis/public-vault/kubernetes-docs/...` in that disposable copy.

Go docs technical corpus:

- repository: `https://github.com/golang/website.git`
- pinned ref: `31fb202f84245709e774bf7c85d13430925d45e5`
- subtree: `_content`
- materialized vault prefix: `sources/golang/website/_content`

Moby-Dick non-technical corpus:

- repository:
  `https://github.com/GITenberg/Moby-Dick--Or-The-Whale_2701.git`
- pinned ref: `bdf1948e6cd00963730971e5624e764a35f238c3`
- subtree: `.`
- materialized vault prefix: `sources/gitenberg/moby-dick`
- source extensions: `.txt`, `.asciidoc`, and `.rst`, converted into
  disposable Markdown documents under `<run-root>`

## Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp public-vault kubernetes-docs \
  --run-root <run-root> \
  --report-name ockp-public-vault-kubernetes-docs

mise exec -- go run ./scripts/agent-eval/ockp public-vault go-docs \
  --run-root <run-root> \
  --report-name ockp-public-vault-go-docs

mise exec -- go run ./scripts/agent-eval/ockp public-vault moby-dick \
  --run-root <run-root> \
  --report-name ockp-public-vault-moby-dick
```

The task manifests are committed at
`docs/evals/public-vault-kubernetes-docs-tasks.json`,
`docs/evals/public-vault-go-docs-tasks.json`, and
`docs/evals/public-vault-moby-dick-tasks.json`.

## Report Policy

Committed reports may include public repository URLs, pinned commits, and public
vault-relative paths. They must not include machine-local paths, raw event logs,
disposable vault contents, SQLite files, document ids, chunk ids, or raw JSON
event output.

The expected committed outputs are:

- `docs/evals/results/ockp-public-vault-kubernetes-docs.md`
- `docs/evals/results/ockp-public-vault-kubernetes-docs.json`
- `docs/evals/results/ockp-public-vault-go-docs.md`
- `docs/evals/results/ockp-public-vault-go-docs.json`
- `docs/evals/results/ockp-public-vault-moby-dick.md`
- `docs/evals/results/ockp-public-vault-moby-dick.json`

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
