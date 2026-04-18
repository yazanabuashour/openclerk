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

## Source-Linked Synthesis Workflow

OpenClerk can support LLM Wiki-style maintenance when the user wants durable knowledge that compounds over time.

- Search existing notes and synthesis before creating a new document.
- Use `CreateDocument` for new canonical source notes or new synthesis pages.
- Use `AppendDocument` or `ReplaceSection` to update existing synthesis instead of duplicating nearby pages.
- Preserve source-sensitive claims with citations, source refs, or provenance references in the document body/frontmatter.
- Use `ListProvenanceEvents` and `ListProjectionStates` when the user asks where knowledge came from, whether derived views are fresh, or whether a synthesis page is stale.
- Treat promoted records as selective structured domains. Do not use records lookup as the default wiki mechanism.
- File a useful answer back into OpenClerk only when it is reusable beyond the current chat and can point back to source evidence.

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
- Maintain source-linked synthesis by searching first, updating existing pages when possible, and preserving citations/provenance for source-sensitive claims.
- Use `local.ResolvePaths(local.Config{})` when the user asks where OpenClerk data lives. It reports `DataDir`, `DatabasePath`, and `VaultRoot` without opening the runtime.

## Generated Client Fallback

The generated OpenAPI client remains available through `client.Generated()` and the legacy `local.Open(...)` return value for raw API-contract work, HTTP debugging, or endpoints not yet covered by the local SDK facade. Do not start there for routine local agent tasks.
