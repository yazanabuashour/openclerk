# Iteration: Harness-Owned Web Search And Fetch

## Status

Iteration recorded after promoting and implementing `web_search_plan`.

Required references:

- [`agent-knowledge-plane.md`](agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Implemented Shape

`openclerk document` now accepts `web_search_plan` with:

- `query`
- harness-supplied `results`
- optional `limit`
- per-result `url`, `title`, `snippet`, optional `source_type`, and optional
  `access_status`

The response returns ranked candidates, duplicate hints, placement hints,
blocked/private candidate status, no-fetch/no-write status, approval boundary,
authority limits, and `agent_handoff`.

## Post-Implementation Boundaries

The surface deliberately does not:

- search the web from inside the runner
- fetch URLs
- use browser automation
- create or update source documents
- create or update synthesis
- treat search snippets as citations
- handle private/authenticated pages as routine source intake

## Follow-Up Trigger

Only consider provider abstraction or live search after real use shows that
harness-supplied search-result planning is safe but leaves repeated discovery
ceremony. A future comparison should include:

- current harness-supplied `web_search_plan`
- a provider abstraction with deterministic fixtures
- no live search provider

The future decision must choose, defer, kill, or record `none viable yet` and
must preserve public-only boundaries, runner-only fetch/write, durable-write
approval, duplicate handling, citations/source refs, provenance, freshness,
local-first behavior, and no-bypass constraints.
