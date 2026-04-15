# Contributing

Outside contributors do not need Beads to contribute to this repository.

## Current project shape

The shipped surface is the embedded Go module exposed through [`client/local`](client/local) plus the generated backend clients in [`client`](client). The repository still contains an HTTP adapter in [`cmd/openclerkd`](cmd/openclerkd), but contributors should treat that as internal compatibility and contract-testing infrastructure rather than the primary product path.

## Local setup

Maintainers prefer:

```bash
mise install
```

The current local validation commands are:

```bash
go generate ./...
git diff --exit-code
test -z "$(gofmt -l $(git ls-files '*.go'))"
go test ./...
golangci-lint run
XDG_DATA_HOME="$(mktemp -d)" go run ./examples/records-client
```

Outside contributors may use their own local tooling if they can satisfy the repository checks.

Beads and Dolt are maintainer-only tools. They are optional for outside contributors and are not required to open, review, or merge pull requests.

## Pull request expectations

- Keep changes reviewable without access to Beads state.
- Update repository docs when the public contract or storage behavior changes.
- Do not commit credentials, private infrastructure details, or sensitive sample data.
- Route security issues through the private process in [SECURITY.md](SECURITY.md), not through public issues or pull requests.

## Checks and review rules

Current pull request checks validate:

- required repository policy files
- machine-path hygiene in committed docs
- generated client drift
- Go formatting
- Go tests
- `golangci-lint`
- dependency-review safety

Pull requests that touch the embedded runtime, generated clients, or examples should leave the repository in a public-safe, policy-consistent, and runnable state without requiring a local daemon.

## Support and compatibility

Before `1.0.0`, compatibility is best effort and may change between releases. The current support target is Go `1.26.x` on current Linux and macOS environments using the embedded local runtime.

Maintainer workflow notes live in [docs/maintainers.md](docs/maintainers.md).
