# Contributing

Outside contributors do not need Beads to contribute to this repository.

## Project Shape

This repository exposes a production `openclerk` runner binary in
`cmd/openclerk` and a single-file OpenClerk skill in
`skills/openclerk/SKILL.md`. The supported product path is the installed runner
plus the skill; OpenClerk does not ship a public Go API, hosted service, remote
HTTP API, or daemon contract.

Changes to the runner, skill, storage behavior, or docs must keep runtime
behavior, setup docs, and CI checks aligned.

## Local Setup

Maintainers prefer:

```bash
mise install
```

Outside contributors may use their own tooling if they can satisfy the
repository checks. Beads and Dolt are maintainer-only tools and are not required
to open, review, or merge pull requests.

Contributors should be able to run:

```bash
printf '%s\n' '{"action":"resolve_paths"}' | \
  OPENCLERK_DATABASE_PATH="$(mktemp -d)/openclerk.sqlite" mise exec -- go run ./cmd/openclerk document
test -z "$(gofmt -l $(git ls-files '*.go'))"
mise exec -- golangci-lint run
mise exec -- go test ./...
mise exec -- ./scripts/validate-agent-skill.sh skills/openclerk
```

If a change touches release notes or release workflow behavior, also run:

```bash
mise exec -- ./scripts/validate-release-docs.sh v0.1.0
```

`golangci-lint` is pinned by `mise.toml`; run it through `mise exec` instead
of relying on a global binary.

## Pull Request Expectations

- Keep changes reviewable without access to Beads state.
- Update repository docs when the public contract or storage behavior changes.
- Do not commit credentials, private infrastructure details, or sensitive sample data.
- Route security issues through the private process in [SECURITY.md](SECURITY.md), not through public issues or pull requests.

## Checks and Review Rules

Pull request checks validate repository policy, Agent Skill metadata shape,
release docs, Go formatting, Go linting, unit tests, CodeQL, and
dependency-review safety.

Pull requests that touch Go code are expected to leave the repository in a
runnable, formatted, lint-clean, and test-clean state. Changes to
`skills/openclerk/SKILL.md` should also pass
`mise exec -- ./scripts/validate-agent-skill.sh skills/openclerk`.

## Support and Compatibility

Before `1.0`, compatibility is best effort and may change between releases.
The production install story is the `openclerk` runner plus the single-file
OpenClerk skill.

Go `1.26.2` is required for repository development and CI validation on
`ubuntu-latest`. Routine client-agent use should not require a Go toolchain.
OpenClerk does not promise a hosted deployment target, remote HTTP API
contract, or public Go package contract.

Maintainer workflow notes live in [docs/maintainers.md](docs/maintainers.md).
