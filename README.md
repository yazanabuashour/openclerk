# openclerk

openclerk is a local-first, agent-facing knowledge plane for notes, documents, promoted records, and provenance-backed retrieval.

The public surface is one authored OpenAPI contract, one generated Go client, and one embedded SQLite-backed runtime that does not require a daemon or bound port. Canonical docs remain markdown in the vault; graph traversal and promoted-domain lookup stay derived from those canonical sources.

## Public surface

- [`client/local`](client/local) opens the embedded runtime in process.
- [`client/openclerk`](client/openclerk) provides the generated request and response types from the same module.
- [`openapi/v1/openclerk.yaml`](openapi/v1/openclerk.yaml) is the contract source of truth.

The legacy `fts`, `hybrid`, `graph`, and `records` packages remain in the repo as implementation-variant fixtures for evals and internal comparison. They are not the preferred product entrypoint.

## Install in your Go project

```bash
go get github.com/yazanabuashour/openclerk/client/local@v0.1.0
```

Import `client/local` to open the embedded runtime and `client/openclerk` from the same module for generated request and response types. Do not document or require a second `go get` for `client/openclerk`.

## Quick start

The normal user path is one embedded runtime opened directly inside the caller's process:

```go
package main

import (
	"context"
	"fmt"
	"log"

	local "github.com/yazanabuashour/openclerk/client/local"
	openclerk "github.com/yazanabuashour/openclerk/client/openclerk"
)

func main() {
	client, runtime, err := local.Open(local.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer runtime.Close()

	create, err := client.CreateDocumentWithResponse(context.Background(), openclerk.CreateDocumentRequest{
		Path:  "notes/hello.md",
		Title: "Hello",
		Body:  "---\ntype: note\nstatus: active\n---\n# Hello\n\n## Summary\nEmbedded OpenClerk runtime.\n",
	})
	if err != nil {
		log.Fatal(err)
	}
	if create.JSON201 == nil {
		log.Fatalf("create document failed: %s", string(create.Body))
	}

	fmt.Printf("doc=%s dataDir=%s\n", create.JSON201.DocId, runtime.Paths().DataDir)
}
```

`cmd/openclerkd serve` remains available for intentional HTTP debugging and compatibility work. When you start that adapter manually, create a remote client explicitly with `openclerk.NewClientWithResponses(baseURL)`.

Runnable example:

```bash
OPENCLERK_DATA_DIR="$(mktemp -d)" go run ./examples/openclerk-client
```

## Default storage

By default, the embedded runtime stores data under:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk
```

That directory contains:

- `openclerk.sqlite` for the SQLite database
- `vault/` for canonical markdown documents

Override any of these locations through [`client/local.Config`](client/local/local.go).

## API surface

Core docs and retrieval:

- `GET /v1/capabilities`
- `POST /v1/search/query`
- `GET /v1/documents`
- `POST /v1/documents`
- `GET /v1/documents/{docId}`
- `GET /v1/documents/{docId}/links`
- `POST /v1/documents/{docId}:append`
- `POST /v1/documents/{docId}:replace-section`
- `GET /v1/chunks/{chunkId}`

Derived capabilities:

- `POST /v1/extensions/graph/neighborhood`
- `POST /v1/extensions/records/lookup`
- `GET /v1/extensions/records/entities/{entityId}`
- `GET /v1/provenance/events`
- `GET /v1/provenance/projections`

The OpenAPI contract remains the single definition of operations, schemas, and generated request and response types even when the runtime is embedded.

## Architecture notes

- Canonical docs stay markdown-backed and inspectable.
- Graph traversal is a derived docs capability, not a second truth system.
- Promoted records are a selective structured layer for domains that fail as plain docs.
- Provenance and projection-state APIs make derivation and freshness inspectable.
- Memory and routing are intentionally out of scope for this rewrite.

See [`docs/architecture/agent-knowledge-plane.md`](docs/architecture/agent-knowledge-plane.md) for the in-repo design summary and [`docs/evals/baseline-scenarios.md`](docs/evals/baseline-scenarios.md) for the eval task set used to compare implementation variants.

## Implementation variants

The repo still contains implementation-specific fixtures and examples used for eval work:

- [`client/fts`](client/fts)
- [`client/hybrid`](client/hybrid)
- [`client/graph`](client/graph)
- [`client/records`](client/records)
- [`examples/fts-client/main.go`](examples/fts-client/main.go)
- [`examples/hybrid-client/main.go`](examples/hybrid-client/main.go)
- [`examples/graph-client/main.go`](examples/graph-client/main.go)
- [`examples/records-client/main.go`](examples/records-client/main.go)

These help benchmark storage and retrieval approaches. They are not the preferred application-facing SDK surface.

## Verify a release

Tagged releases publish:

- a source archive
- a SHA-256 checksum file
- a CycloneDX SBOM
- a Sigstore-backed provenance bundle
- a Sigstore-backed SBOM bundle

The source-only asset contract is the same for `v0.1.0` and later tags. To verify a tagged release:

```bash
shasum -a 256 -c openclerk-v0.y.z.tar.gz.sha256
gh attestation verify openclerk-v0.y.z.tar.gz --repo yazanabuashour/openclerk
```

The release assets and attestation bundles are generated by [`.github/workflows/release.yml`](.github/workflows/release.yml).

## Local development

Install pinned tooling with:

```bash
mise install
```

Regenerate clients and verify there is no drift:

```bash
go generate ./...
git diff --exit-code
```

Run formatting, tests, and lint:

```bash
test -z "$(gofmt -l $(git ls-files '*.go'))"
go test ./...
golangci-lint run
OPENCLERK_DATA_DIR="$(mktemp -d)" go run ./examples/openclerk-client
```

## Repository contents

- [`client/local`](client/local) contains the embedded runtime entrypoint for the public client.
- [`client/openclerk`](client/openclerk) contains the primary generated Go client.
- [`client`](client) also contains internal variant clients used for evals.
- [`openapi/v1/openclerk.yaml`](openapi/v1/openclerk.yaml) contains the contract source of truth.
- [`internal/infra/sqlite`](internal/infra/sqlite) contains the SQLite-backed implementation and derived projections.
- [`cmd/openclerkd`](cmd/openclerkd) remains available for adapter work and contract testing, but it is not the primary runtime path.
- [`docs/maintainers.md`](docs/maintainers.md) explains Beads-based maintainer workflow and release administration notes.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for contribution expectations, [`SECURITY.md`](SECURITY.md) for vulnerability reporting, and [`SKILL.md`](SKILL.md) for the agent-facing usage guide.
