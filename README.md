# OpenClerk

OpenClerk is a local-first knowledge-plane runtime for agents. The supported
agent path is a small `openclerk` runner plus a single-file skill.

## Install

Tell your agent:

```text
Install OpenClerk from https://github.com/yazanabuashour/openclerk.
Complete both required steps before reporting success:
1. Install and verify the openclerk runner binary with `openclerk --version`.
2. Register the OpenClerk skill from skills/openclerk/SKILL.md using your native skill system.
```

For the latest release:

```bash
sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"
```

For a pinned release:

```bash
OPENCLERK_VERSION=v0.1.0 sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/v0.1.0/install.sh)"
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
1. Upgrade and verify the openclerk runner binary with `openclerk --version`.
2. Re-register the OpenClerk skill from skills/openclerk/SKILL.md using your native skill system.
```

Or upgrade the runner manually:

```bash
sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"
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

Validation rejections are JSON results with `rejected: true`. Runtime failures
exit non-zero and write errors to stderr.

## Local Storage

The default data directory is
`${XDG_DATA_HOME:-~/.local/share}/openclerk`. It contains `openclerk.sqlite`
and the markdown `vault/`. Override storage with `OPENCLERK_DATA_DIR`,
`OPENCLERK_DATABASE_PATH`, or `OPENCLERK_VAULT_ROOT`.

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
  OPENCLERK_DATA_DIR="$(mktemp -d)" mise exec -- go run ./cmd/openclerk document
test -z "$(gofmt -l $(git ls-files '*.go'))"
mise exec -- golangci-lint run
mise exec -- go test ./...
mise exec -- ./scripts/validate-agent-skill.sh skills/openclerk
mise exec -- ./scripts/validate-release-docs.sh v0.1.0
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
