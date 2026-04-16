---
name: openclerk
description: Use OpenClerk when an agent needs a local-first knowledge plane for canonical notes, documents, promoted records, and provenance-backed retrieval in a Go workspace through the code-first local SDK.
license: MIT
---

# OpenClerk Agent SDK

Use OpenClerk when the task needs local notes, documents, promoted records, or provenance-backed retrieval without running a daemon or depending on an external service.

## Default Path

- Install the current development line with `go get github.com/yazanabuashour/openclerk/client/local@main` until the first release tag is published.
- Import `github.com/yazanabuashour/openclerk/client/local`.
- Open live local state with `local.OpenClient(local.Config{})`.
- Use the code-first methods on `*local.Client` for routine work: `CreateDocument`, `ListDocuments`, `Search`, `AppendDocument`, `ReplaceSection`, `GetDocumentLinks`, `GraphNeighborhood`, `LookupRecords`, `GetRecordEntity`, `ListProvenanceEvents`, and `ListProjectionStates`.
- Use `local.Config{DataDir: "..."}`, `local.Config{DatabasePath: "..."}`, or `local.Config{VaultRoot: "..."}` only when the user names a specific dataset or you are using an isolated test database.

Do not inspect generated clients, generated server code, large dependency directories, or the Go module cache for routine document/search/records/provenance tasks. Use targeted repo searches only when the local SDK facade does not cover the user's ask.

## Quick Start

```go
package main

import (
	"context"
	"log"

	"github.com/yazanabuashour/openclerk/client/local"
)

func main() {
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
		log.Printf("%s %s", hit.DocID, hit.Snippet)
	}
}
```

## Common Tasks

- Create, list, read, append, and replace canonical Markdown documents with the snippets in [references/documents.md](references/documents.md).
- Search what OpenClerk knows with path and metadata filters using [references/search.md](references/search.md).
- Inspect derived graph links, promoted records, provenance events, and projection freshness with [references/records-provenance.md](references/records-provenance.md).
- Use `local.ResolvePaths(local.Config{})` when the user asks where OpenClerk data lives. It reports `DataDir`, `DatabasePath`, and `VaultRoot` without opening the runtime.

## Generated Client Fallback

The generated OpenAPI client remains available through `client.Generated()` and the legacy `local.Open(...)` return value for raw API-contract work, HTTP debugging, or endpoints not yet covered by the local SDK facade. Do not start there for routine local agent tasks.
