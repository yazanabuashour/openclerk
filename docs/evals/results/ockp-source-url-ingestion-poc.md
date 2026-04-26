# Source URL Ingestion POC

## Status

Decision: promote `openclerk document` action `ingest_source_url`.

Issue: `oc-jb0`

## Targeted Evidence

The POC compared the current document/retrieval runner surface against source
URL ingestion requirements for PDF sources.

Current actions can create markdown source notes, index markdown retrieval
text, inspect provenance, and reject duplicate document paths. They cannot
download a source URL, store a PDF asset under the configured vault, compute
asset hash/size/page count, extract PDF metadata/text, reject duplicate source
URLs, or return a single result tying asset provenance to the canonical source
note. Doing that externally would require routine agents to orchestrate HTTP,
filesystem asset writes, PDF parsing, note generation, duplicate detection, and
post-write validation outside the runner.

That is a structural runner capability gap rather than a skill-guidance gap.
The promoted action keeps canonical markdown authoritative: the generated
`sources/*.md` note has `modality: markdown` and `source_type: pdf`; the PDF is
registered as a vault asset path; derived retrieval text is stored in the
canonical source note and indexed through existing chunk/search behavior.

## Promoted Contract

Request shape:

```json
{
  "action": "ingest_source_url",
  "source": {
    "url": "https://example.test/source.pdf",
    "path_hint": "sources/example.md",
    "asset_path_hint": "assets/sources/example.pdf",
    "title": "Optional title"
  }
}
```

Response shape adds `ingestion` with created `doc_id`, `source_path`,
`asset_path`, `derived_path`, citations, SHA256, size, MIME type, page count,
capture timestamp, and optional PDF metadata.

The action rejects missing fields before runtime initialization, accepts only
HTTP/HTTPS PDF sources, keeps paths vault-relative, rejects duplicate
`source_url` values unless a future update mode is explicitly designed, and
validates generated metadata, asset existence, and indexed citations before
success.

## Targeted Checks

Implementation coverage is in Go tests rather than a full agent eval run:

- missing, invalid-scheme, and unsafe path requests reject before creating the
  runtime database or vault
- non-PDF responses reject
- a served PDF fixture creates the asset, source note, source metadata,
  retrieval citations, and provenance details
- duplicate source URLs reject without update behavior

Quality gates for the implementation are `mise exec -- go test ./...` and
`mise exec -- golangci-lint run`.

## Follow-Up

Update/re-ingest behavior was intentionally out of scope for `oc-jb0`. The
explicit update-mode contract was accepted in `oc-n31` and implemented by
`oc-6fr`.
