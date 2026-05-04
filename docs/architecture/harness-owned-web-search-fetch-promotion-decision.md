---
decision_id: decision-harness-owned-web-search-fetch-promotion
decision_title: Harness-Owned Web Search And Fetch Promotion
decision_status: accepted
decision_scope: harness-owned-web-search-fetch
decision_owner: platform
---
# Decision: Harness-Owned Web Search And Fetch Promotion

## Status

Accepted: promote `openclerk document` `web_search_plan` as a read-only
planning action for harness-supplied web search results. Do not promote live
search provider calls, browser automation, HTTP fetch outside
`ingest_source_url`, private/authenticated page acquisition, or durable writes
from search results.

Required references:

- [`agent-knowledge-plane.md`](agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

Evidence:

- [`harness-owned-web-search-fetch-adr.md`](harness-owned-web-search-fetch-adr.md)
- [`../evals/harness-owned-web-search-fetch-poc.md`](../evals/harness-owned-web-search-fetch-poc.md)
- [`../evals/results/ockp-harness-owned-web-search-fetch.md`](../evals/results/ockp-harness-owned-web-search-fetch.md)

## Promoted Surface

```json
{"action":"web_search_plan","web_search":{"query":"example source","results":[{"url":"https://example.test/page.html","title":"Example","snippet":"Search snippet"}],"limit":10}}
```

The runner:

- accepts deterministic harness-supplied search results
- validates HTTP/HTTPS public URL candidates
- ranks candidates by supplied order
- infers `web` or `pdf`
- checks duplicate source URL metadata through the installed runner
- proposes source, asset, and synthesis placement hints
- returns a next `ingest_source_url` request shape for approved fetch/write
- returns `agent_handoff`

Defaults:

- `limit` defaults to 10 and caps at 20.
- `access_status` defaults to `public`.
- only `web` and `pdf` source types are accepted.

## Decision

Promote because the candidate passes safety, capability, and UX quality for
search-to-source planning. It extends the natural existing `openclerk document`
source-intake surface without changing the durable fetch/write path.

Search results are discovery hints only. They do not create citations, source
authority, provenance, or projection freshness. Those begin only after the user
approves `ingest_source_url` and the installed runner creates or updates a
source document.

## Non-Promoted Surfaces

Live search provider integration is not promoted. It may be useful later, but
it needs deterministic fixtures or a provider abstraction, egress/privacy
policy, freshness/error modeling, rate-limit behavior, and separate evidence.

Browser automation and non-runner HTTP fetch are killed for routine OpenClerk
source intake. They bypass the installed runner and blur public/private,
blocked, authenticated, and durable-write boundaries.

## Compatibility

Existing `ingest_source_url` behavior remains unchanged. Existing public URL
fetch/write still requires approval and must use `ingest_source_url`.
`web_search_plan` is read-only and safe to call before approval.

## Follow-Up

Do not implement live search or provider configuration from this decision.
Future iteration may compare provider abstraction candidates after the
read-only planner has real use evidence. That comparison must choose, defer,
kill, or record `none viable yet` and must preserve runner-only access,
public-only boundaries, approval-before-write, citations/source refs,
provenance, freshness, duplicate handling, and local-first operation.
