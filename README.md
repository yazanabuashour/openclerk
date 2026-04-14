# openclerk

## Repository contents

- [CONTRIBUTING.md](CONTRIBUTING.md) explains how outside contributors should propose changes.
- [SECURITY.md](SECURITY.md) explains how to report vulnerabilities privately and what response timing to expect.
- [docs/maintainers.md](docs/maintainers.md) documents Beads-based maintainer workflow and repo administration notes.
- [cmd/openclerk](cmd/openclerk) contains the bootstrap Go CLI entrypoint.
- [LICENSE](LICENSE) defines the project license.

## Release contract

The initial release surface is GitHub Releases with semantic version tags in the `0.y.z` range. Release notes are generated from protected tags. This repository does not currently publish packages or downloadable build artifacts.

## Local development

Install pinned tooling with:

```bash
mise install
```

Run the bootstrap CLI with:

```bash
go run ./cmd/openclerk
```

Validate the Go module with:

```bash
gofmt -w cmd internal
go test ./...
golangci-lint run
```

## Compatibility

The current runnable surface is a bootstrap Go CLI. Maintainers validate against Go `1.26.x` and do not yet promise support beyond best-effort local development on current macOS and Linux environments.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.
