# Contributing

Outside contributors do not need Beads to contribute to this repository.

## Current Project Shape

The shipped product surface is the installed `openclerk` JSON runner plus the
single-file skill at `skills/openclerk/SKILL.md`. There is no public importable
Go API, hosted service, remote HTTP API, or daemon in the supported product
path.

## Public Install Contract

For agents, install the repository:

```text
Install https://github.com/yazanabuashour/openclerk
```

The repository publishes an Agent Skills-compatible skill at `skills/openclerk`
and an `openclerk` runner binary. Agents should use their native skill installer
or skill directory; docs should not assume a specific agent vendor or fixed
skill path.

## Local Setup

Maintainers prefer:

```bash
mise install
```

The current local validation commands are:

```bash
test -z "$(gofmt -l $(git ls-files '*.go'))"
go test ./...
mise exec -- golangci-lint run
```

Outside contributors may use their own local tooling if they can satisfy the
repository checks.

Beads and Dolt are maintainer-only tools. They are optional for outside
contributors and are not required to open, review, or merge pull requests.

## Pull Request Expectations

- Keep changes reviewable without access to Beads state.
- Update repository docs when the public contract or storage behavior changes.
- Do not commit credentials, private infrastructure details, or sensitive sample data.
- Route security issues through the private process in `SECURITY.md`, not through public issues or pull requests.

## Checks And Review Rules

Current pull request checks validate:

- required repository policy files
- machine-path hygiene in committed docs
- Agent Skills package validity
- Go formatting
- Go tests
- `golangci-lint`
- dependency-review safety

Pull requests that touch the internal runtime, runner, or skill should
leave the repository in a public-safe, policy-consistent, and runnable state
without requiring a local daemon.

If a change affects the public product story, keep the docs aligned with the
single-surface agent knowledge plane framing in `README.md` and
`docs/architecture/agent-knowledge-plane.md`.

## Support And Compatibility

Before `1.0.0`, compatibility is best effort and may change between releases.
The current support target is Go `1.26.x` on current Linux and macOS
environments using the installed local runner.

Maintainer workflow notes live in `docs/maintainers.md`.
