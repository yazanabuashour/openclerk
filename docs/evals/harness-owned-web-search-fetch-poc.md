# POC: Harness-Owned Web Search And Fetch

## Purpose

This POC evaluates a deterministic read-only planning surface for
harness-supplied web search results. It does not authorize a live search
provider, browser automation, HTTP fetch, durable write, source ingestion, or
synthesis creation.

Required references:

- [`../architecture/agent-knowledge-plane.md`](../architecture/agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Candidate Surface

```json
{"action":"web_search_plan","web_search":{"query":"public source planning evidence","results":[{"url":"https://example.test/new-report.pdf","title":"New Report","snippet":"Public report search snippet."}],"limit":10}}
```

The candidate accepts search results from the harness, then returns:

- query
- rank
- URL and normalized URL
- title and snippet as discovery hints
- inferred source type
- duplicate source hints
- candidate source path hints
- candidate asset path hints for PDFs
- candidate synthesis path
- blocked/private/public candidate status
- next `ingest_source_url` request shape for approved fetch/write
- `agent_handoff`

## Scenarios

| Scenario | Expected behavior |
| --- | --- |
| Public PDF search result | Infer `pdf`, propose `sources/*.md` and `assets/sources/*.pdf`, no fetch/write. |
| Public web search result | Infer `web`, propose `sources/web/*.md`, no fetch/write. |
| Duplicate URL result | Return existing runner-visible source document and omit synthesis placement. |
| Authenticated/private result | Mark unsupported and do not fetch. |
| Invalid URL or negative limit | Return JSON validation rejection before storage work. |

## Deterministic Fixture

The unit fixture seeds an existing source document with `source_url` metadata,
then calls `web_search_plan` with one new PDF result, one duplicate public web
result, and one authenticated result. The fixture commits no generated
corpus, database, raw logs, private content, or machine-absolute paths.

## Acceptance

The eval report must record safety pass, capability pass, UX quality,
performance posture, and evidence posture separately. Promotion must keep
approved fetch/write in `ingest_source_url`.
