# OpenClerk

OpenClerk is a local-first knowledge-plane runtime for agents. The supported
agent path is a small `openclerk` runner plus a single-file skill.

## Install

Tell your agent:

```text
Install OpenClerk from https://github.com/yazanabuashour/openclerk.
Complete both required steps before reporting success:
1. Install the openclerk runner binary into a durable user-level binary directory.
   Use `OPENCLERK_INSTALL_DIR="$HOME/.local/bin"` for a normal user-level install.
   Do not install it under `.codex/tmp`, `/tmp`, a repository checkout, or another ephemeral workspace.
   Verify with `command -v openclerk` and `openclerk --version`, and confirm `command -v openclerk` resolves to that durable install target.
   If `$HOME/.local/bin` is not on the user's future shell `PATH`, update the appropriate shell startup file and re-verify.
2. Register the OpenClerk skill from skills/openclerk/SKILL.md using your native skill system.
```

For the latest release:

```bash
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"
```

For a pinned release:

```bash
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" OPENCLERK_VERSION=v0.2.3 sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/v0.2.3/install.sh)"
```

A complete install has two parts:

- `openclerk --version` succeeds
- the matching skill is registered from `skills/openclerk/SKILL.md`,
  `https://github.com/yazanabuashour/openclerk/tree/<tag>/skills/openclerk`,
  or `openclerk_<version>_skill.tar.gz`

Use the agent's native skill manager. OpenClerk does not require a specific
skill path or agent implementation.

## Upgrade

Tell your agent:

```text
Upgrade OpenClerk from https://github.com/yazanabuashour/openclerk.
Complete both required steps before reporting success:
1. Upgrade the openclerk runner binary in its durable user-level binary directory.
   Use `OPENCLERK_INSTALL_DIR="$HOME/.local/bin"` for a normal user-level install.
   Do not upgrade it under `.codex/tmp`, `/tmp`, a repository checkout, or another ephemeral workspace.
   Verify with `command -v openclerk` and `openclerk --version`, and confirm `command -v openclerk` resolves to that durable install target.
   If `$HOME/.local/bin` is not on the user's future shell `PATH`, update the appropriate shell startup file and re-verify.
2. Re-register the OpenClerk skill from skills/openclerk/SKILL.md using your native skill system.
```

Or upgrade the runner manually:

```bash
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"
```

Then verify the runner and re-register the matching skill:

```bash
command -v openclerk
openclerk --version
```

## AgentOps Architecture

OpenClerk's agent-facing path is the AgentOps pattern: the skill gives the
agent task policy, and the local runner performs stateful knowledge-plane
operations through structured JSON. This keeps product rules close to the
agent, avoids broad repo search and lower-level runtime bypasses, and leaves
storage local instead of requiring a hosted service.

## Agent Skill Budget

`skills/openclerk/SKILL.md` is intentionally a thin activation, routing, and
safety contract. It should tell agents when to use `openclerk document` and
`openclerk retrieval`, define hard no-tools and bypass boundaries, and point
routine work at runner-owned JSON surfaces.

Long `SKILL.md` recipes are taste debt. If routine success depends on exact
JSON, command ordering, or workflow-specific prompt choreography, that evidence
should drive a candidate-surface comparison: keep primitives with a smaller
skill, extend an existing runner action, or add a narrow workflow action with
`agent_handoff`. Agents are expected to use their own autonomy with runner
help, JSON results, and runner rejections once the safe surface is clear.

## Runner Interface

The skill sends structured JSON on stdin and reads structured JSON from stdout
for these runner domains:

```bash
openclerk document
openclerk retrieval
```

Example:

```bash
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":10}}' |
  openclerk retrieval
```

Promoted narrow workflow actions cover routine source-backed workflows without
requiring agents to choreograph many primitives:

- `openclerk document --help` and `openclerk retrieval --help` expose a compact
  runner-owned action index for these promoted workflows, so routine agents do
  not need long `SKILL.md` recipes or source inspection to find the request
  shape.
- `openclerk document` `ingest_source_url` supports read-only `plan` mode for
  public-link placement before durable fetch/write. It returns candidate source
  path hints, duplicate source status, synthesis-placement guidance,
  no-fetch/no-write status, approval boundaries, and `agent_handoff`.
- `openclerk document` `compile_synthesis` creates or updates exactly one
  source-linked synthesis target from either explicit `body` markdown or
  runner-assembled `body_facts`, defaults the only supported
  `create_or_update` mode, builds required Sources/Freshness sections when
  needed, and returns source evidence, duplicate status, provenance refs,
  projection freshness, write status, validation boundaries, authority limits,
  and `agent_handoff`.
- `openclerk retrieval` `source_audit_report` explains source-sensitive audit
  evidence and can repair only an existing synthesis target in
  `repair_existing` mode. It returns `agent_handoff` and is not a broad
  contradiction engine.
- `openclerk retrieval` `evidence_bundle_report` is read-only and packages
  records, decisions, citations, provenance, projection freshness, validation
  boundaries, authority limits, and `agent_handoff`.
- `openclerk retrieval` `duplicate_candidate_report` is read-only and packages
  the likely duplicate target, evidence inspected, no-write status, approval
  boundary, validation boundaries, authority limits, and `agent_handoff`.

Validation rejections are JSON results with `rejected: true`. Runtime failures
exit non-zero and write errors to stderr.

## Local Storage

The default database is
`${XDG_DATA_HOME:-~/.local/share}/openclerk/openclerk.sqlite`. The database
stores the configured markdown vault root. Override the database location with
`OPENCLERK_DATABASE_PATH` or `--db`.

When troubleshooting configuration after an upgrade or runner failure, inspect
the effective database and vault paths before changing setup:

```bash
printf '%s\n' '{"action":"resolve_paths"}' | openclerk document
printf '%s\n' '{"action":"inspect_layout"}' | openclerk document
```

For an existing vault, bind it once during setup or intentionally rebind it:

```bash
openclerk init --vault-root <vault-root>
```

Do not use `init` as routine repair for document or retrieval errors; use
`resolve_paths` and `inspect_layout` first to confirm which database and
configured vault root the runner is using.

## Eval Evidence

The production runner/skill passed the current OpenClerk release gate:
[`docs/evals/results/ockp-agentops-production.md`](docs/evals/results/ockp-agentops-production.md).
The eval protocol is documented in
[`docs/evals/agent-production.md`](docs/evals/agent-production.md).

Architecture and deferred-capability decisions are preserved under
[`docs/architecture`](docs/architecture).

## Development

Use the full local toolchain for repository development:

```bash
mise install
printf '%s\n' '{"action":"resolve_paths"}' | \
  OPENCLERK_DATABASE_PATH="$(mktemp -d)/openclerk.sqlite" mise exec -- go run ./cmd/openclerk document
test -z "$(gofmt -l $(git ls-files '*.go'))"
mise exec -- golangci-lint run
mise exec -- go test ./...
mise exec -- ./scripts/validate-agent-skill.sh skills/openclerk
mise exec -- ./scripts/validate-release-docs.sh v0.2.3
mise exec -- go run ./scripts/agent-eval/ockp run --report-name ockp-agentops-production
mise exec -- go run ./scripts/agent-eval/ockp run --parallel 4 --scenario repo-docs-agentops-retrieval,repo-docs-synthesis-maintenance,repo-docs-decision-records,repo-docs-release-readiness,repo-docs-tag-filter,repo-docs-memory-router-recall-report,repo-docs-release-synthesis-freshness --report-name ockp-repo-docs-dogfood
mise exec -- go run ./scripts/agent-eval/ockp run --scenario compile-synthesis-workflow-action-natural --report-name ockp-compile-synthesis-workflow-action
mise exec -- go run ./scripts/agent-eval/ockp run --scenario source-audit-workflow-action-natural --report-name ockp-source-audit-workflow-action
mise exec -- go run ./scripts/agent-eval/ockp run --scenario evidence-bundle-workflow-action-natural --report-name ockp-evidence-bundle-workflow-action
```

`golangci-lint` is pinned by `mise.toml`; use `mise exec -- golangci-lint run`
for local checks.

## Releases

Tagged `v0.y.z` releases publish platform binary archives, the skill archive,
the installer, source archive, SHA256 checksums, an SBOM, and GitHub
attestations. Published release assets are intended to be immutable going
forward. See
[`docs/release-verification.md`](docs/release-verification.md) for verification
steps.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests.
Beads is maintainer-only workflow tooling and is not required for community
contributions.

See `CONTRIBUTING.md` for contribution expectations, `CODE_OF_CONDUCT.md` for
community standards, `SECURITY.md` for vulnerability reporting, and
`docs/maintainers.md` for maintainer-only workflow details.
