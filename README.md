# openclerk

openclerk is a local-first, agent-facing knowledge plane for notes, documents, promoted records, source-linked synthesis, and provenance-backed retrieval.

The production agent surface is the `cmd/openclerk-agentops` JSON runner backed by the `agentops` package. The developer surface is one code-first local SDK, one generated Go client for fallback contract work, one authored OpenAPI contract, and one embedded SQLite-backed runtime that does not require a daemon or bound port. Canonical docs remain markdown in the vault; graph traversal and promoted-domain lookup stay derived from those canonical sources.

OpenClerk is also infrastructure for persistent agent-maintained knowledge. It is meant to help useful synthesis compound over time as cited, inspectable markdown rather than being rediscovered from scratch on every query or lost in chat history.

## Public surface

- [`agentops`](agentops) exposes task-shaped document and retrieval facades for coding agents.
- [`cmd/openclerk-agentops`](cmd/openclerk-agentops) is the machine-facing JSON runner agents should use for routine local knowledge tasks.
- [`client/local`](client/local) opens the embedded runtime in process and provides the preferred code-first SDK facade.
- [`client/openclerk`](client/openclerk) provides generated request and response types from the same module for raw OpenAPI fallback work.
- [`openapi/v1/openclerk.yaml`](openapi/v1/openclerk.yaml) is the contract source of truth.

Backend-specific public clients have been removed. Graph, records, and provenance are OpenClerk capabilities, not separate public backends.

## Install in your Go project

```bash
go get github.com/yazanabuashour/openclerk/client/local@main
```

Import `client/local` to open the embedded runtime and use routine OpenClerk workflows without generated response wrappers. Do not document or require a second `go get` for `client/openclerk`; it is part of the same module when raw OpenAPI types are needed.

For agent workflows inside this repository, use the JSON runner:

```bash
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":10}}' |
  go run ./cmd/openclerk-agentops retrieval
```

## Quick start

The normal user path is one embedded runtime opened directly inside the caller's process:

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
		Body:  "---\ntype: note\nstatus: active\n---\n# Hello\n\n## Summary\nEmbedded OpenClerk runtime.\n",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("doc=%s dataDir=%s\n", document.DocID, client.Paths().DataDir)
}
```

`local.OpenClient(...)` opens SQLite locally, syncs the vault, and calls the in-process service directly. `local.Open(...)`, `Client.Generated()`, and `cmd/openclerkd serve` remain available for intentional HTTP debugging, compatibility work, or raw OpenAPI response handling.

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

Override any of these locations through [`client/local.Config`](client/local/local.go), `OPENCLERK_DATA_DIR`, `OPENCLERK_DATABASE_PATH`, or `OPENCLERK_VAULT_ROOT`. Explicit config fields and runner flags take precedence over environment variables.

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

## Code-first local SDK

Prefer `local.OpenClient(...)` for Go application and developer workflows:

```go
client, err := local.OpenClient(local.Config{})
if err != nil {
	log.Fatal(err)
}
defer client.Close()

results, err := client.Search(context.Background(), local.SearchOptions{
	Text:  "architecture",
	Limit: 10,
})
if err != nil {
	log.Fatal(err)
}
for _, hit := range results.Hits {
	fmt.Printf("%s %s\n", hit.DocID, hit.Snippet)
}
```

The facade covers document create/list/get, search, append, replace-section, links, graph neighborhood, records lookup, record entity reads, provenance events, and projection states. Use generated methods only for endpoints or raw response details not yet covered by the facade.

## Architecture notes

- Canonical docs stay markdown-backed and inspectable.
- Source-linked synthesis can live in markdown when it carries citations and provenance back to canonical sources.
- Graph traversal is a derived docs capability, not a second truth system.
- Promoted records are a selective structured layer for domains that fail as plain docs.
- Provenance and projection-state APIs make derivation and freshness inspectable.
- Memory and routing are intentionally out of scope for this rewrite.

See [`docs/architecture/agent-knowledge-plane.md`](docs/architecture/agent-knowledge-plane.md) for the in-repo design summary, [`docs/evals/baseline-scenarios.md`](docs/evals/baseline-scenarios.md) for the eval task set, and [`docs/evals/agent-production.md`](docs/evals/agent-production.md) for production agent workflow eval guidance.

### LLM-maintained synthesis

Karpathy's LLM Wiki pattern is related to the OpenClerk vision: both reject pure query-time RAG as the whole answer and favor durable markdown knowledge that compounds through summaries, links, contradiction checks, and filed answers.

OpenClerk should support that workflow through its existing docs, search, graph, records, and provenance surface before adding new public APIs. The OpenClerk version keeps raw sources and accepted canonical notes inspectable, treats synthesis as source-linked markdown, and uses provenance/freshness state so agent-authored synthesis does not become an opaque second truth system.

## AgentOps Runner

Agents should use `cmd/openclerk-agentops` for routine document and retrieval tasks instead of creating temporary Go programs or inspecting generated clients.

```bash
printf '%s\n' '{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}' |
  go run ./cmd/openclerk-agentops document
```

Validation rejections are JSON results with `rejected: true`. Runtime failures exit non-zero and write errors to stderr.

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

- [`client/local`](client/local) contains the embedded runtime entrypoint and code-first SDK facade.
- [`client/openclerk`](client/openclerk) contains the generated Go client for raw OpenAPI fallback work.
- [`agentops`](agentops) contains the task-shaped agent facade used by the JSON runner.
- [`openapi/v1/openclerk.yaml`](openapi/v1/openclerk.yaml) contains the contract source of truth.
- [`internal/infra/sqlite`](internal/infra/sqlite) contains the SQLite-backed implementation and derived projections.
- [`cmd/openclerk-agentops`](cmd/openclerk-agentops) contains the machine-facing AgentOps JSON runner.
- [`cmd/openclerkd`](cmd/openclerkd) remains available for adapter work and contract testing, but it is not the primary runtime path.
- [`docs/maintainers.md`](docs/maintainers.md) explains Beads-based maintainer workflow and release administration notes.

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for contribution expectations, [`SECURITY.md`](SECURITY.md) for vulnerability reporting, and [`skills/openclerk/SKILL.md`](skills/openclerk/SKILL.md) for the agent-facing usage guide.
