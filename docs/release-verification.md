# Release Verification

Tagged OpenClerk releases publish:

- `openclerk_<version>_<os>_<arch>.tar.gz`
- `openclerk_<version>_skill.tar.gz`
- `openclerk_<version>_source.tar.gz`
- `openclerk_<version>_checksums.txt`
- `openclerk_<version>_sbom.json`
- `install.sh`

The platform archives contain the `openclerk` runner. The skill archive
contains `skills/openclerk/SKILL.md`. Checksums and GitHub attestations verify
that release assets were produced by this repository's workflow.
Release assets are intended to be immutable once published. If an artifact is
wrong, ship a new patch release instead of mutating the existing release.

## Verify a Release

Download the assets from the GitHub Release page for the tag you want to verify,
then run:

```bash
shasum -a 256 -c openclerk_<version>_checksums.txt
gh attestation verify openclerk_<version>_<os>_<arch>.tar.gz --repo yazanabuashour/openclerk
gh attestation verify openclerk_<version>_skill.tar.gz --repo yazanabuashour/openclerk
gh attestation verify openclerk_<version>_source.tar.gz --repo yazanabuashour/openclerk
gh attestation verify install.sh --repo yazanabuashour/openclerk
```

For the latest release, verify GitHub's latest pointer resolves to the expected
tag:

```bash
gh release view --repo yazanabuashour/openclerk --json tagName --jq .tagName
```

## Smoke-Test an Install

Install into a temporary directory, then verify the runner version and commands:

```bash
install_dir="$(mktemp -d)"
OPENCLERK_INSTALL_DIR="$install_dir" \
  OPENCLERK_VERSION=v0.2.3 \
  sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/v0.2.3/install.sh)"

export PATH="$install_dir:$PATH"
command -v openclerk
openclerk --version
openclerk --help
printf '%s\n' '{"action":"resolve_paths"}' | openclerk document
printf '%s\n' '{"action":"inspect_layout"}' | openclerk document
```

The valid runner domains are `document` and `retrieval`. A complete install
also registers the matching `skills/openclerk/SKILL.md` with the user's agent.
Release verification should confirm installed runner and skill alignment: the
agent-facing examples must use the installed `openclerk document` and
`openclerk retrieval` commands, not source-tree binaries, direct SQLite access,
HTTP/MCP bypasses, or retired APIs.

Use `resolve_paths` and `inspect_layout` before `openclerk init` when
diagnosing an upgraded install or routine runner failure. `init --vault-root
<vault-root>` is for first-time binding or intentional rebinding of an existing
vault, not routine repair.

For v0.2.x source URL update behavior, the release notes and skill examples
should match this runner request shape:

```bash
printf '%s\n' '{"action":"ingest_source_url","source":{"url":"https://example.test/source.pdf","mode":"update"}}' | openclerk document
```

For web source URL behavior after `oc-v1ed`, the skill examples should match
this runner request shape and must not imply external browser, HTTP/MCP,
purchase, login, captcha, paywall, or direct vault acquisition:

```bash
printf '%s\n' '{"action":"ingest_source_url","source":{"url":"https://example.test/page.html","path_hint":"sources/web/example.md","source_type":"web"}}' | openclerk document
```

For supplied-transcript video/YouTube source ingestion, the release notes and
skill examples should match this runner request shape and must not imply native
video download, platform caption retrieval, local STT, transcript APIs, or
Gemini extraction:

```bash
printf '%s\n' '{"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","path_hint":"sources/video-youtube/demo.md","transcript":{"text":"Supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}' | openclerk document
```

For promoted workflow actions after `oc-e8om`, `oc-w8x0`, and `oc-lrqi`,
release notes and skill examples may claim only these narrow surfaces:

```bash
printf '%s\n' '{"action":"compile_synthesis","synthesis":{"path":"synthesis/example.md","title":"Example","source_refs":["sources/example.md"],"body":"# Example\n\n## Summary\nSource-backed synthesis.\n\n## Sources\n- sources/example.md\n\n## Freshness\nChecked with runner-visible source evidence.","mode":"create_or_update"}}' | openclerk document
printf '%s\n' '{"action":"source_audit_report","source_audit":{"query":"source-sensitive audit runner repair evidence","target_path":"synthesis/example.md","mode":"explain","limit":10}}' | openclerk retrieval
printf '%s\n' '{"action":"evidence_bundle_report","evidence_bundle":{"query":"AgentOps Escalation Policy","entity_id":"agentops-escalation-policy","projection":"records","limit":10}}' | openclerk retrieval
```

These claims do not imply a broad contradiction engine, embeddings/vector DB,
memory transport, autonomous router API, browser acquisition, direct local-file
intake, or lower-level storage access.

For `oc-nj5h`, the targeted workflow-action reports must also show
`ux_quality: workflow_action_acceptable` on natural rows. The installed runner
may expose compact request-shape help through `openclerk document --help` and
`openclerk retrieval --help`; do not move long workflow recipes into
`skills/openclerk/SKILL.md` to satisfy these lanes.

The current full production OpenClerk AgentOps gate remains
`docs/evals/results/ockp-agentops-production.md`. Source URL update mode is
covered by targeted AgentOps evidence at
`docs/evals/results/ockp-source-url-update-mode.md`; that targeted lane proves
duplicate create rejection, same-SHA no-op updates, changed-PDF stale synthesis
visibility, and path-hint conflict no-write behavior, but does not replace the
release-blocking production gate. Supplied-transcript video/YouTube ingestion
is covered by targeted AgentOps evidence at
`docs/evals/results/ockp-video-youtube-canonical-source-note.md`; that lane
proves `ingest_video_url` create/update behavior, transcript provenance,
citation-bearing search, same-hash no-op behavior, changed-transcript stale
synthesis visibility, and external-tool bypass rejection.

Future timestamp-span citations, platform caption retrieval, local STT, and
remote transcript API policies are design-only in
`docs/architecture/video-transcript-acquisition-design.md`. Release notes,
skills, and smoke tests must not imply those acquisition paths are available
until a later promoted implementation ships them.

## Pre-Release Dogfood

Before tagging any release, refresh both the full AgentOps production gate and
the mandatory repo-docs dogfood lane:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run --report-name ockp-agentops-production
mise exec -- go run ./scripts/agent-eval/ockp run --parallel 4 --scenario repo-docs-agentops-retrieval,repo-docs-synthesis-maintenance,repo-docs-decision-records,repo-docs-release-readiness,repo-docs-tag-filter,repo-docs-memory-router-recall-report,repo-docs-release-synthesis-freshness --report-name ockp-repo-docs-dogfood
mise exec -- go run ./scripts/agent-eval/ockp run --scenario compile-synthesis-workflow-action-natural --report-name ockp-compile-synthesis-workflow-action
mise exec -- go run ./scripts/agent-eval/ockp run --scenario source-audit-workflow-action-natural --report-name ockp-source-audit-workflow-action
mise exec -- go run ./scripts/agent-eval/ockp run --scenario evidence-bundle-workflow-action-natural --report-name ockp-evidence-bundle-workflow-action
```

The dogfood lane imports only committed public markdown into an isolated
OpenClerk eval vault and exercises installed `openclerk document` and
`openclerk retrieval` JSON results. It covers repo-doc retrieval,
source-linked synthesis maintenance, decision-record explainability,
release-readiness answers, read-side tag filters, the read-only
`memory_router_recall_report` action, and release synthesis freshness. Commit
the reduced `docs/evals/results/ockp-repo-docs-dogfood.md` and `.json`
reports; never commit raw logs or machine-absolute artifact paths.

Pre-release evidence must separate safety, capability, and UX. Scripted or
exact-command rows prove capability and safety only; they do not satisfy UX
acceptance unless matching natural prompts pass with acceptable command count,
assistant turns, prompt specificity, and retry rate. Treat
`workflow_choreography_gap`, `skill_bloat_risk`, and
`ergonomics_gap_despite_capability_pass` as tag-blocking taste debt unless a
decision note explicitly classifies the failure as fixture/reporting-only.

Committed reports and docs must use repo-relative artifact paths. Raw eval log
references, when included in reduced reports, must use neutral placeholders
such as `<run-root>/<variant>/<scenario>/turn-N/events.jsonl` rather than
machine-absolute paths. Raw logs are not committed.

## SBOM

The JSON SBOM asset is intended for audit tooling and manual inspection:

```bash
jq '.components | length' openclerk_<version>_sbom.json
```

The SBOM is generated from the tagged source contents during the release
workflow and attached to the same GitHub Release as the binary, skill, and
source archives.
