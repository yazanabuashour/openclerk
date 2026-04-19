# openclerk

OpenClerk is a local-first, agent-facing knowledge plane for notes, documents,
promoted records, source-linked synthesis, and provenance-backed retrieval.

The production agent surface is the installed `openclerk` JSON runner. The Go
developer surface is the direct local SDK in `client/local`. There is no hosted
service, remote HTTP API, or daemon in the `0.1.0` path.

OpenClerk is also infrastructure for persistent agent-maintained knowledge:
useful synthesis should become cited, inspectable markdown rather than being
rediscovered from scratch on every query or lost in chat history.

## Quickstart

### Agent Install

Tell your agent:

```text
Install https://github.com/yazanabuashour/openclerk
```

The repository publishes an Agent Skills-compatible skill at
`skills/openclerk` and an `openclerk` runner binary. Agents should use their
native skill installer or skill directory to install the skill; this repository
does not assume a specific agent vendor or skill path.

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

## Runner Interface

The skill calls these runner domains:

```bash
openclerk document
openclerk retrieval
```

The runner reads structured JSON from stdin, validates and normalizes the
request, performs the local knowledge-plane operation, and writes structured
JSON to stdout.

Example:

```bash
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":10}}' |
  openclerk retrieval
```

Validation rejections are JSON results with `rejected: true`. Runtime failures
exit non-zero and write errors to stderr.

## Local Go SDK

Go developers can embed the same local runtime directly:

```bash
go get github.com/yazanabuashour/openclerk/client/local@v0.1.0
```

Minimal usage from Go:

```go
package main

import (
	"context"
	"fmt"
	"log"

	local "github.com/yazanabuashour/openclerk/client/local"
)

func main() {
	client, err := local.OpenClient(local.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	document, err := client.CreateDocument(context.Background(), local.DocumentInput{
		Path:  "notes/hello.md",
		Title: "Hello",
		Body:  "# Hello\n\n## Summary\nEmbedded OpenClerk runtime.\n",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("doc=%s dataDir=%s\n", document.DocID, client.Paths().DataDir)
}
```

`local.OpenClient(...)` opens SQLite locally, syncs the vault, and calls the
same local service used by the runner.

## Local Storage

By default, the local runtime stores data under:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk
```

That directory contains:

- `openclerk.sqlite` for the SQLite database
- `vault/` for canonical markdown documents

Override storage with `client/local.Config`, `OPENCLERK_DATA_DIR`,
`OPENCLERK_DATABASE_PATH`, or `OPENCLERK_VAULT_ROOT`. Explicit config fields and
runner flags take precedence over environment variables.

## Architecture Notes

- Canonical docs stay markdown-backed and inspectable.
- Source-linked synthesis can live in markdown when it carries citations and
  provenance back to canonical sources.
- Graph traversal is a derived docs capability, not a second truth system.
- Promoted records are a selective structured layer for domains that fail as
  plain docs.
- Provenance and projection-state reads make derivation and freshness
  inspectable.
- Memory and routing are intentionally out of scope for this release.

See `docs/architecture/agent-knowledge-plane.md` for the in-repo design
summary, `docs/evals/baseline-scenarios.md` for the eval task set, and
`docs/evals/agent-production.md` for production agent workflow eval guidance.

## Eval Evidence

The runner-backed production skill beat the SDK-oriented baseline in the latest
full proof-obligation eval report:
`docs/evals/results/ockp-adr-proof-obligations.md`.

## Contributing and Maintainer Setup

Repository development uses the full local toolchain:

```bash
mise install
OPENCLERK_DATA_DIR="$(mktemp -d)" go run ./examples/openclerk-client
test -z "$(gofmt -l $(git ls-files '*.go'))"
go test ./...
golangci-lint run
```

## Release Contract

The `0.1.0` release deliverables are:

- platform archives for the `openclerk` binary
- the Agent Skills-compatible `openclerk` skill archive
- the release installer script
- the Go module import path rooted at `github.com/yazanabuashour/openclerk`
- the direct-local Go package at `github.com/yazanabuashour/openclerk/client/local`

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
