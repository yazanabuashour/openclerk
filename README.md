# openclerk

openclerk is a local-first, agent-facing knowledge plane for notes, documents, promoted records, and provenance-backed retrieval.

The public surface is one authored OpenAPI contract, one generated Go client, and one embedded SQLite-backed runtime that does not require a daemon or bound port. Canonical docs remain markdown in the vault; graph traversal and promoted-domain lookup stay derived from those canonical sources.

## Public surface

- [`client/openclerk`](client/openclerk) is the primary generated SDK.
- [`client/local`](client/local) opens the embedded runtime in process.
- [`openapi/v1/openclerk.yaml`](openapi/v1/openclerk.yaml) is the contract source of truth.

The legacy `fts`, `hybrid`, `graph`, and `records` packages remain in the repo as implementation-variant fixtures for evals and internal comparison. They are not the preferred product entrypoint.

## Install the embedded client

```bash
go get github.com/yazanabuashour/openclerk/client/local@latest
go get github.com/yazanabuashour/openclerk/client/openclerk@latest
```

For reproducible installs, pin the module to a release tag such as `@v0.y.z`.

## Quick start

The primary usage flow is one embedded client over one agent-facing surface:

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
	ctx := context.Background()
	client, runtime, err := local.Open(local.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer runtime.Close()

	architecture, err := client.CreateDocumentWithResponse(ctx, openclerk.CreateDocumentRequest{
		Path:  "notes/architecture/knowledge-plane.md",
		Title: "Knowledge plane",
		Body:  "---\ntype: note\nstatus: active\n---\n# Knowledge plane\n\n## Summary\nCanonical architecture note.\n",
	})
	if err != nil {
		log.Fatal(err)
	}
	if architecture.JSON201 == nil {
		log.Fatalf("create document failed: %s", string(architecture.Body))
	}

	roadmap, err := client.CreateDocumentWithResponse(ctx, openclerk.CreateDocumentRequest{
		Path:  "notes/projects/openclerk-roadmap.md",
		Title: "Roadmap",
		Body:  "---\ntype: project\nstatus: active\n---\n# Roadmap\n\n## Summary\nSee the [knowledge plane](../architecture/knowledge-plane.md).\n",
	})
	if err != nil {
		log.Fatal(err)
	}
	if roadmap.JSON201 == nil {
		log.Fatalf("create linked document failed: %s", string(roadmap.Body))
	}

	record, err := client.CreateDocumentWithResponse(ctx, openclerk.CreateDocumentRequest{
		Path:  "records/assets/transmission-solenoid.md",
		Title: "Transmission solenoid",
		Body:  "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\ntype: record\nstatus: active\n---\n# Transmission solenoid\n\n## Summary\nCanonical promoted-domain baseline.\n\n## Facts\n- sku: SOL-1\n- vendor: OpenClerk Motors\n",
	})
	if err != nil {
		log.Fatal(err)
	}
	if record.JSON201 == nil {
		log.Fatalf("create record failed: %s", string(record.Body))
	}

	pathPrefix := "notes/"
	docs, err := client.ListDocumentsWithResponse(ctx, &openclerk.ListDocumentsParams{PathPrefix: &pathPrefix})
	if err != nil || docs.JSON200 == nil {
		log.Fatal("list documents failed")
	}

	links, err := client.GetDocumentLinksWithResponse(ctx, roadmap.JSON201.DocId)
	if err != nil || links.JSON200 == nil {
		log.Fatal("get document links failed")
	}

	lookup, err := client.RecordsLookupWithResponse(ctx, openclerk.RecordsLookupRequest{Text: "solenoid"})
	if err != nil || lookup.JSON200 == nil {
		log.Fatal("records lookup failed")
	}

	refKind := "document"
	events, err := client.ListProvenanceEventsWithResponse(ctx, &openclerk.ListProvenanceEventsParams{
		RefKind: &refKind,
		RefId:   &roadmap.JSON201.DocId,
	})
	if err != nil || events.JSON200 == nil {
		log.Fatal("list provenance events failed")
	}

	fmt.Printf("docs=%d links=%d entity=%s events=%d dataDir=%s\n",
		len(docs.JSON200.Documents),
		len(links.JSON200.Outgoing),
		lookup.JSON200.Entities[0].EntityId,
		len(events.JSON200.Events),
		runtime.Paths().DataDir,
	)
}
```

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
- Sigstore-backed provenance and SBOM attestation bundles

To verify a tagged release:

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
