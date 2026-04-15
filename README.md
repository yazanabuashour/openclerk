# openclerk

openclerk is a local-first, agent-facing document store with one authored OpenAPI contract and four committed generated Go clients over SQLite-backed backend variants.

## Backend variants

| Backend | Generated package | Architecture |
| --- | --- | --- |
| `fts` | [`client/fts`](client/fts) | Canonical documents + deterministic chunks + SQLite FTS5/BM25 |
| `hybrid` | [`client/hybrid`](client/hybrid) | FTS core plus persisted local embeddings and reciprocal-rank fusion |
| `graph` | [`client/graph`](client/graph) | FTS core plus evidence-linked graph projections |
| `records` | [`client/records`](client/records) | FTS core plus promoted entities and provenance-backed record lookup |

The contract source of truth is [`openapi/v1/openclerk.yaml`](openapi/v1/openclerk.yaml). Generated Go client code is committed and is the public SDK surface for this repository.

## API surface

Stable core methods:

- `GET /v1/capabilities`
- `POST /v1/search/query`
- `GET /v1/documents/{docId}`
- `GET /v1/chunks/{chunkId}`
- `POST /v1/documents`
- `POST /v1/documents/{docId}:append`
- `POST /v1/documents/{docId}:replace-section`

Extension methods:

- `POST /v1/extensions/graph/neighborhood`
- `POST /v1/extensions/records/lookup`
- `GET /v1/extensions/records/entities/{entityId}`

v1 is local-first and intentionally simple:

- Auth mode is `none`.
- There is no users table.
- SQLite is the only persisted database.
- Canonical markdown documents are stored under the configured vault root and indexes are derived from them.

## Quick start

Start the daemon against a local SQLite file and vault root:

```bash
go run ./cmd/openclerkd serve \
  --backend=fts \
  --db ./var/openclerk.sqlite \
  --vault-root ./vault \
  --addr 127.0.0.1:8080
```

To enable hybrid vector scoring without an external vector database, start the hybrid backend with the local embedding provider:

```bash
go run ./cmd/openclerkd serve \
  --backend=hybrid \
  --embedding-provider local \
  --db ./var/openclerk.sqlite \
  --vault-root ./vault \
  --addr 127.0.0.1:8080
```

## Install generated clients

Install the package that matches the backend you want to benchmark or ship:

```bash
go get github.com/yazanabuashour/openclerk/client/fts@v0.x.y-alpha.n
go get github.com/yazanabuashour/openclerk/client/hybrid@v0.x.y-alpha.n
go get github.com/yazanabuashour/openclerk/client/graph@v0.x.y-alpha.n
go get github.com/yazanabuashour/openclerk/client/records@v0.x.y-alpha.n
```

Each package exposes the generated typed client directly via `NewClientWithResponses`.

## Example programs

These examples are runnable against an empty local daemon. Each one creates sample content through the generated client before reading it back.

- [`examples/fts-client/main.go`](examples/fts-client/main.go)
- [`examples/hybrid-client/main.go`](examples/hybrid-client/main.go)
- [`examples/graph-client/main.go`](examples/graph-client/main.go)
- [`examples/records-client/main.go`](examples/records-client/main.go)

Run one with:

```bash
OPENCLERK_SERVER=http://127.0.0.1:8080 go run ./examples/fts-client
```

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

Run formatting and tests:

```bash
gofmt -w $(git ls-files '*.go')
go test ./...
golangci-lint run
```

## Repository contents

- [`cmd/openclerkd`](cmd/openclerkd) contains the HTTP daemon entrypoint.
- [`cmd/openclerk`](cmd/openclerk) contains the bootstrap CLI entrypoint.
- [`internal/app`](internal/app) contains application services and use cases.
- [`internal/domain`](internal/domain) defines the strict domain contracts and shared types.
- [`internal/infra/sqlite`](internal/infra/sqlite) contains the SQLite-backed implementations and derived projections.
- [`client`](client) contains committed generated Go clients.
- [`docs/maintainers.md`](docs/maintainers.md) explains Beads-based maintainer workflow and repo administration notes.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for contribution expectations, [`SECURITY.md`](SECURITY.md) for vulnerability reporting, and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) for community standards.
