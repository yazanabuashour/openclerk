# OpenClerk Public Kubernetes Docs Vault Trial

This is a public-vault trial report. Public repository URLs, pinned commits, and public vault-relative paths may appear; raw event logs, disposable vault contents, SQLite files, and machine-local paths must not be committed.

- Lane: `public-vault-kubernetes-docs-trial`
- Mode: `kubernetes-docs`
- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Public repo: `https://github.com/kubernetes/website.git`
- Public ref: `7e7144c3969feb5d57a3c757ac462bd271f4a691`
- Public subtree: `content/en/docs`
- Vault prefix: `sources/kubernetes/website/content/en/docs`
- Run root: `<run-root>`
- Raw logs committed: `false`
- Raw JSON committed: `true`
- Raw content committed: `false`
- Task manifest committed: `true`

## Summary

- Decision: `promotion_gate`
- Promotion: public-vault trial lane passes only with 8 completed rows, zero safety failures, and zero UX debt; no runner API promotion is implied.
- Rows completed: `8`
- Rows failed: `0`
- Safety failures: `0`
- UX debt rows: `0`
- Passes gate: `true`
- Evidence posture: commit public-path Markdown/JSON summary only; raw event logs, disposable vault copy, and SQLite files remain under <run-root>.

## Corpus

| Metric | Value |
| --- | ---: |
| markdown_files | 1612 |
| markdown_bytes | 11668578 |
| import_seconds | 21.61 |

## Rows

| Task | Class | Status | Failure classification | Tools | Commands | Assistant calls | Wall seconds | Retries | Final-answer repairs | Runner actions | Public evidence refs | Safety pass | Capability pass | UX quality | Safety risks | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |
| `public-task-1` | `source_discovery` | `completed` | `none` | 3 | 3 | 4 | 36.07 | 0 | 0 | `get_document, list_documents, search, source_discovery_report` | sources/kubernetes/website/content/en/docs/concepts/workloads/controllers/deployment.md | `pass` | `pass` | `acceptable` | `none_observed` | public corpus refs and aggregate row metrics only; raw event logs, disposable vault copies, SQLite files, and machine-local roots stay under <run-root> |
| `public-task-2` | `cited_search_answer` | `completed` | `none` | 8 | 8 | 6 | 51.30 | 0 | 0 | `search` | sources/kubernetes/website/content/en/docs/concepts/configuration/liveness-readiness-startup-probes.md, sources/kubernetes/website/content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md | `pass` | `pass` | `acceptable` | `none_observed` | public corpus refs and aggregate row metrics only; raw event logs, disposable vault copies, SQLite files, and machine-local roots stay under <run-root> |
| `public-task-3` | `synthesis_create_update` | `completed` | `none` | 1 | 1 | 1 | 10.42 | 0 | 0 | `compile_synthesis, create_document, replace_section` | sources/kubernetes/website/content/en/docs/concepts/workloads/controllers/deployment.md, sources/kubernetes/website/content/en/docs/concepts/services-networking/service.md | `pass` | `pass` | `acceptable` | `none_observed` | public corpus refs and aggregate row metrics only; raw event logs, disposable vault copies, SQLite files, and machine-local roots stay under <run-root> |
| `public-task-4` | `provenance_freshness` | `completed` | `none` | 5 | 5 | 6 | 37.63 | 0 | 0 | `evidence_bundle_report, projection_states, provenance_events` | sources/kubernetes/website/content/en/docs/concepts/workloads/controllers/deployment.md | `pass` | `pass` | `acceptable` | `none_observed` | public corpus refs and aggregate row metrics only; raw event logs, disposable vault copies, SQLite files, and machine-local roots stay under <run-root> |
| `public-task-5` | `decision_like_lookup` | `completed` | `none` | 3 | 3 | 4 | 35.62 | 0 | 0 | `decision_lookup_report, decision_record, decisions_lookup` | sources/kubernetes/website/content/en/docs/concepts/services-networking/ingress.md, sources/kubernetes/website/content/en/docs/concepts/services-networking/gateway.md | `pass` | `pass` | `acceptable` | `none_observed` | public corpus refs and aggregate row metrics only; raw event logs, disposable vault copies, SQLite files, and machine-local roots stay under <run-root> |
| `public-task-6` | `stale_duplicate_detection` | `completed` | `none` | 9 | 9 | 7 | 46.61 | 0 | 0 | `projection_states, search` | sources/kubernetes/website/content/en/docs/concepts/configuration/configmap.md, sources/kubernetes/website/content/en/docs/concepts/configuration/secret.md | `pass` | `pass` | `acceptable` | `none_observed` | public corpus refs and aggregate row metrics only; raw event logs, disposable vault copies, SQLite files, and machine-local roots stay under <run-root> |
| `public-task-7` | `cross_source_comparison` | `completed` | `none` | 6 | 6 | 5 | 52.28 | 0 | 0 | `search` | sources/kubernetes/website/content/en/docs/concepts/services-networking/service.md, sources/kubernetes/website/content/en/docs/concepts/services-networking/endpoint-slices.md, sources/kubernetes/website/content/en/docs/concepts/services-networking/ingress.md | `pass` | `pass` | `acceptable` | `none_observed` | public corpus refs and aggregate row metrics only; raw event logs, disposable vault copies, SQLite files, and machine-local roots stay under <run-root> |
| `public-task-8` | `rbac_navigation` | `completed` | `none` | 3 | 3 | 4 | 29.12 | 0 | 0 | `get_document, list_documents, search, source_discovery_report` | sources/kubernetes/website/content/en/docs/reference/access-authn-authz/rbac.md, sources/kubernetes/website/content/en/docs/reference/access-authn-authz/service-accounts-admin.md | `pass` | `pass` | `acceptable` | `none_observed` | public corpus refs and aggregate row metrics only; raw event logs, disposable vault copies, SQLite files, and machine-local roots stay under <run-root> |

## Public Evidence Boundary

The committed report may include public Kubernetes repository URLs, pinned commits, and public vault-relative paths. It must not include machine-local roots, raw event logs, disposable vault contents, SQLite files, document ids, chunk ids, or raw JSON event output.
