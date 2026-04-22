# openclerk

OpenClerk is a local-first, agent-facing knowledge plane for notes, documents,
promoted records, source-linked synthesis, and provenance-backed retrieval.

The supported product interface is the AgentOps pattern: a shipped `openclerk`
runner plus the single-file skill at `skills/openclerk/SKILL.md`. There is no
public importable Go API, hosted service, remote HTTP API, or daemon in the
supported product path.

OpenClerk is infrastructure for persistent agent-maintained knowledge: useful
synthesis should become cited, inspectable markdown rather than being
rediscovered from scratch on every query or lost in chat history.

## Quickstart

### Agent Install

Tell your agent:

```text
Install https://github.com/yazanabuashour/openclerk
```

The repository publishes an Agent Skills-compatible skill at `skills/openclerk`
and an `openclerk` runner binary. Agents should use their native skill
installer or skill directory; this repository does not assume a specific agent
vendor or skill path.

### Manual Install, Latest Release

```bash
curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh | sh
```

The installer installs only the `openclerk` runner binary. It prints the skill
source URL so you can install `skills/openclerk` with your agent's native skill
installer or skill directory.

### Manual Install, Pinned Version

```bash
curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/v0.1.0/install.sh | sh
```

Use this for reproducible setup.

## AgentOps Architecture

OpenClerk's agent-facing path is AgentOps: the skill gives the agent task
policy, and the local runner performs stateful knowledge-plane operations
through structured JSON. This keeps product rules close to the agent, avoids
broad repo search and ad hoc lower-level workflows, and leaves storage local
without requiring a hosted service.

The runner/skill pair is the competitive interface for agents. MCP or other
adapters may be evaluated later only if they wrap equivalent runner semantics
and improve measured agent behavior without weakening validation, provenance,
or source-authority guarantees.

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

Service-centric retrieval can use the first typed promoted-domain projection:

```bash
printf '%s\n' '{"action":"services_lookup","services":{"text":"OpenClerk runner","interface":"JSON runner","limit":10}}' |
  openclerk retrieval
```

Configured knowledge layout is explained through runner-derived JSON, not a
committed manifest:

```bash
printf '%s\n' '{"action":"inspect_layout"}' | openclerk document
```

Validation rejections are JSON results with `rejected: true`. Runtime failures
exit non-zero and write errors to stderr.

## Local Storage

By default, the runner stores data under:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk
```

That directory contains:

- `openclerk.sqlite` for the SQLite database
- `vault/` for canonical markdown documents

Override storage with `OPENCLERK_DATA_DIR`, `OPENCLERK_DATABASE_PATH`, or
`OPENCLERK_VAULT_ROOT`. Runner flags `--data-dir`, `--db`, `--vault-root`, and
`--embedding-provider` are for explicit datasets, tests, or manual debugging.

## Architecture Notes

- Canonical docs stay markdown-backed and inspectable.
- Source-linked synthesis can live in markdown when it carries citations and
  provenance back to canonical sources.
- Graph traversal is a derived docs capability, not a second truth system.
- The service registry is the first typed promoted-domain prototype; promoted
  records remain selective structured layers for domains that fail as plain
  docs.
- Provenance and projection-state reads make derivation and freshness
  inspectable, including source-linked synthesis freshness through the
  `synthesis` projection.
- Memory and autonomous routing are intentionally out of scope for this release.

See `docs/architecture/agent-knowledge-plane.md` for the in-repo design
summary, `docs/evals/baseline-scenarios.md` for the eval task set, and
`docs/evals/agent-production.md` for production agent workflow eval guidance.

## Eval Evidence

Production evals gate the shipped AgentOps surface against correctness and
hygiene requirements: no direct SQLite access, no broad repo search, no module
cache inspection, no source-built runner bypass, and final-answer-only rejection
for rule-covered invalid requests.

## Contributing and Maintainer Setup

Repository development uses the full local toolchain:

```bash
mise install
test -z "$(gofmt -l $(git ls-files '*.go'))"
go test ./...
mise exec -- golangci-lint run
```

## Release Contract

The `0.1.0` release deliverables are:

- platform archives for the `openclerk` binary
- the Agent Skills-compatible `openclerk` skill archive
- the release installer script
- the Go module import path rooted at `github.com/yazanabuashour/openclerk`

The release workflow is built around semantic version tags in the `v0.y.z`
range. Each tagged GitHub Release publishes binary archives, the skill archive,
a release installer, a canonical source archive, SHA256 checksums, an SBOM, and
GitHub attestations for release verification.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests.
Beads is maintainer-only workflow tooling and is not required for community
contributions.

See `CONTRIBUTING.md` for contribution expectations, `SECURITY.md` for
vulnerability reporting, and `skills/openclerk/SKILL.md` for the agent-facing
usage guide.
