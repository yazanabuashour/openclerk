# Contributing

Outside contributors do not need Beads to contribute to this repository.

## Current project shape

This repository now includes a bootstrap Go CLI in [cmd/openclerk](cmd/openclerk). There is still no published package or deployed service, so changes should keep the runnable surface intentionally small and documented.

## Local setup

Maintainers prefer:

```bash
mise install
```

The current local validation commands are:

```bash
go run ./cmd/openclerk
go test ./...
golangci-lint run
```

Outside contributors may use their own local tooling if they can satisfy the repository checks.

Beads and Dolt are maintainer-only tools. They are optional for outside contributors and are not required to open, review, or merge pull requests.

## Pull request expectations

- Keep changes reviewable without access to Beads state.
- Update repository docs when the public contract changes.
- Do not commit credentials, private infrastructure details, or sensitive sample data.
- Route security issues through the private process in [SECURITY.md](SECURITY.md), not through public issues or pull requests.

## Checks and review rules

Current pull request checks validate repository policy, dependency-review safety, Go formatting, and Go tests. Pull requests that touch the bootstrap CLI should leave the repository in a public-safe, policy-consistent, and runnable state.

## Support and compatibility

Before `0.1.0`, compatibility is best effort and may change between releases. The current support target is Go `1.26.x` for local development on current macOS and Linux environments.

Maintainer workflow notes live in [docs/maintainers.md](docs/maintainers.md).
