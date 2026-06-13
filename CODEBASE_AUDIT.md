# Codebase Audit

Date: 2026-06-13

## Scope

Audit goal: improve this repository until no high- or medium-confidence, evidence-backed, safe improvements remain under the current local validation surface.

Constraints honored:

- Current branch only.
- No skill installation or persistent skill writes.
- External skill material read as methodology only from a temporary cache outside the repository.
- Committed docs use repo-relative paths or neutral placeholders only.
- Repo-pinned tools from `mise.toml` are run through `mise exec --`.

## Repo Map

- `cmd/openclerk/`: CLI entrypoint and runner command wiring.
- `internal/runner/`: public task-shaped JSON contract for document, retrieval, config, URL planning, artifact planning, synthesis, graph, and eval report actions.
- `internal/runclient/`: local runtime path resolution, write locks, module config, and runner client facade.
- `internal/infra/sqlite/`: SQLite-backed store, vault filesystem writes, source URL/video ingestion, projections, provenance, sync, graph, and record extraction.
- `internal/domain/`: domain types, errors, and vault-relative path normalization.
- `modules/`: optional module manifests, skills, docs, and the semantic retrieval adapter implementation.
- `skills/openclerk/`: production agent skill surface.
- `scripts/`: release/install validation and OCKP eval harnesses.
- `docs/architecture/`, `docs/evals/`, `docs/security-operations.md`: evidence and decision history for runner surfaces and release/security posture.
- `.github/`: PR, release, CodeQL, dependency review, and repository policy automation.

## Validation Baseline

Commands run:

- `mise exec -- sh -c 'printf "%s\n" "{\"action\":\"resolve_paths\"}" | OPENCLERK_DATABASE_PATH="$(mktemp -d)/openclerk.sqlite" go run ./cmd/openclerk document'`
- `mise exec -- sh -c 'test -z "$(gofmt -l $(git ls-files "*.go"))"'`
- `mise exec -- golangci-lint run`
- `mise exec -- go test ./...`
- `mise exec -- ./scripts/validate-all-agent-skills.sh`
- `mise exec -- ./scripts/validate-committed-artifacts.sh`
- `mise exec -- go build ./cmd/openclerk ./modules/semantic-retrieval-adapter ./internal/tools/validateagentskill ./internal/tools/validatecommittedartifacts ./internal/tools/validatereleasedocs ./internal/tools/validateliveinstallsmoke`
- `mise exec -- go test ./... -coverprofile=<temp>/openclerk-cover.out`

Baseline result: all commands passed.

Coverage snapshot from the temporary cover profile:

- `cmd/openclerk`: 82.6%
- `internal/chronicler`: 83.6%
- `internal/domain`: 54.3%
- `internal/infra/sqlite`: 72.9%
- `internal/runclient`: 51.5%
- `internal/runner`: 82.1%
- `modules/semantic-retrieval-adapter`: 74.2%
- `scripts/agent-eval/ockp`: 57.9%

Lower coverage is concentrated in thin facade/error helpers, module/runtime tooling, and eval/support code. The primary document/retrieval runner packages have strong coverage.

## Methodology References Used

Read-only external skill references considered:

- `shadcn/improve` `SKILL.md` and audit playbook: evidence-backed findings, impact/risk/effort, and no secret reproduction.
- `mattpocock/improve-codebase-architecture` `SKILL.md`, `LANGUAGE.md`, and `DEEPENING.md`: module depth, locality, leverage, and deletion-test framing.
- Anthropic engineering `architecture`, `code-review`, `tech-debt`, and `testing-strategy`: ADR tradeoffs, correctness/security/performance review, debt scoring, and test-strategy focus.
- Vercel React best practices: considered but not applied because no React/Next.js surface was found.

Repo-local guidance used:

- `AGENTS.md`
- `README.md`
- `CONTRIBUTING.md`
- `SECURITY.md`
- `docs/maintainers.md`
- `docs/security-operations.md`
- Recent architecture/eval material under `docs/architecture/` and `docs/evals/`

## Initial Prioritized Findings

### F1. Reject URL userinfo on public/source URL boundaries

- Evidence: `internal/runner/urls.go` parses runner HTTP URLs without rejecting `parsed.User`.
- Evidence: `internal/infra/sqlite/normalize.go` normalizes source URLs without rejecting userinfo before storage metadata and duplicate matching.
- Evidence: `internal/infra/sqlite/ingest_source.go` validates fetch scheme/host but does not reject userinfo as defense in depth before HTTP fetch.
- Impact: A URL like `https://user:pass@example.test/source.md` can carry embedded credentials through runner planning, source ingestion, stored metadata, provenance-adjacent output, or outbound fetch handling. OpenClerk's public-source model treats public user-provided URLs as fetchable evidence, but embedded credentials are not public evidence and should not be preserved or sent.
- Effort: S.
- Fix risk: Low. Normal public URLs, query strings, GitHub URL normalization, and fragments remain supported; only userinfo-bearing URLs become invalid.
- Confidence: High.
- Proposed validation: add runner URL tests, source URL normalization tests, then run targeted tests plus full baseline validation.
- Status: fixed. Runner HTTP URL parsing now rejects userinfo in `internal/runner/urls.go`; SQLite source URL normalization and fetch target validation now reject userinfo in `internal/infra/sqlite/normalize.go` and `internal/infra/sqlite/ingest_source.go`. Regression tests were added in `internal/runner/urls_test.go` and `internal/infra/sqlite/source_url_normalize_test.go`.

## Validation After Fix

Narrow checks:

- `mise exec -- go test ./internal/runner -run 'TestRunnerHTTPURLValidation|TestRunnerLoopbackHTTPURLValidation|TestRunnerGeminiAPIBaseValidation'`
- `mise exec -- go test ./internal/infra/sqlite -run 'TestNormalizeDocumentSourceURL|TestNormalizeDocumentSourceURLWithAliasesIncludesGitHubOriginals|TestNormalizeDocumentSourceURLRejectsUserinfo|TestSourceFetchValidationRejectsUserinfo'`

Full checks:

- `mise exec -- sh -c 'printf "%s\n" "{\"action\":\"resolve_paths\"}" | OPENCLERK_DATABASE_PATH="$(mktemp -d)/openclerk.sqlite" go run ./cmd/openclerk document'`
- `mise exec -- sh -c 'test -z "$(gofmt -l $(git ls-files "*.go"))"'`
- `mise exec -- golangci-lint run`
- `mise exec -- go test ./...`
- `mise exec -- ./scripts/validate-all-agent-skills.sh`
- `mise exec -- ./scripts/validate-committed-artifacts.sh`
- `mise exec -- go build ./cmd/openclerk ./modules/semantic-retrieval-adapter ./internal/tools/validateagentskill ./internal/tools/validatecommittedartifacts ./internal/tools/validatereleasedocs ./internal/tools/validateliveinstallsmoke`
- `mise exec -- go test ./... -coverprofile=<temp>/openclerk-cover-final.out`
- `git diff --check`

Result: all checks passed.

## Final Audit Passes

Two post-fix audit passes found no remaining high- or medium-confidence safe improvements under the current validation surface.

Pass 1 perspectives:

- Correctness/security review of runner URL validation, source ingestion, artifact planning, module config, path normalization, write locks, and semantic module dispatch.
- Architecture/depth review of runner/store seams and module boundaries. Existing seams are earning their keep: runner JSON is the public interface, `runclient` owns local runtime access, SQLite owns durable store behavior, and optional modules are manifest-verified adapters.
- Testing review of critical source/vault/module paths. Primary runner packages are well covered; the selected fix added narrow regression tests at both runner and storage layers.

Pass 1 outcome: no additional high- or medium-confidence safe code changes found.

Pass 2 perspectives:

- Security/DX review of install scripts, release bundle scripts, GitHub workflows, committed skill policy, release verification docs, dynamic SQL sites, command execution sites, and ignored error/JSON fallback sites.
- Release and CI posture review confirmed pinned actions, read-only PR permissions, no `pull_request_target`, release writes behind `release`, checksum/SBOM/attestation docs, and committed artifact validation.
- Taste/UX review confirmed public URL inspect/fetch permission remains distinct from durable-write approval and that promoted workflow actions live in runner surfaces rather than expanding `skills/openclerk/SKILL.md`.

Pass 2 outcome: no additional high- or medium-confidence safe code changes found.

## Considered And Rejected

- GitHub Actions `pull_request_target` exposure: not present. PR workflow uses read-only `contents` permission and pinned actions in `.github/workflows/pull-request.yml`.
- Release workflow untrusted write exposure: release writes are tag-triggered, depend on verify job, use a protected `release` environment, and pinned actions in `.github/workflows/release.yml`.
- Source ingestion SSRF via private hosts: current code validates public hostnames/IPs before fetch and during redirects/dialing in `internal/infra/sqlite/ingest_source.go`; tests cover private URL rejection for runner plan mode.
- Vault path traversal through document writes: current path normalization and symlink checks in `internal/domain/vaultpath.go`, `internal/infra/sqlite/vault_paths.go`, and source/document tests cover parent symlink and traversal cases.
- Module command injection for semantic modules: semantic module command and args are pinned/rejected in `internal/runclient/modules.go`, with poisoning tests in `internal/runclient/modules_test.go`.
- React/Next.js performance pass: no React/Next.js files or package manifest were present.

## Change Log

- Hardened runner/source URL validation to reject embedded URL userinfo before planning, storage normalization, or fetch handling.
- Added regression tests for runner public URLs, syntax-only URLs, loopback module URLs, source URL normalization, and lower-level source fetch target validation.
- Created this audit report with repo map, validation baseline, methodology references, findings, validation, and remaining risks.

## Remaining Risks

- The audit is local and validation-backed. It does not include live Dependabot/GitHub Security tab state, private vulnerability reports, or release-environment settings beyond committed workflow evidence.
- Lower coverage remains in thin facades, support tooling, and eval harness code, but no high-risk uncovered behavior was found that had a safe, high-confidence fix in this pass.
- Deeper release/eval gates such as live install smoke and AgentOps production dogfood were not run because this change does not alter release packaging, installer behavior, skill policy, or workflow-action UX. The normal release process still requires those gates before tagging.
- No remote publication was requested or performed.
