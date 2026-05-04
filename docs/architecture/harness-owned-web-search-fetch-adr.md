---
decision_id: adr-harness-owned-web-search-fetch
decision_title: Harness-Owned Web Search And Fetch Coordination
decision_status: accepted
decision_scope: harness-owned-web-search-fetch
decision_owner: platform
---
# ADR: Harness-Owned Web Search And Fetch Coordination

## Status

Accepted for targeted POC/eval. Product behavior is authorized only by the
promotion decision in
[`harness-owned-web-search-fetch-promotion-decision.md`](harness-owned-web-search-fetch-promotion-decision.md).

Required references:

- [`agent-knowledge-plane.md`](agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Context

OpenClerk already has runner-owned public URL fetch/write through
`ingest_source_url`, including read-only `plan` mode. The missing surface is
not another fetcher. It is coordination between harness-owned web search
results and the existing source-ingestion workflow.

The user expectation is natural: "find sources about this and document the
best one" should not require exact prompt choreography or manual path
construction. The safety boundary is equally important: public search/read
permission is not durable-write approval, search snippets are not source
evidence, and approved fetch/write must still flow through `ingest_source_url`.

## Options

| Option | Safety | Capability | UX quality | Decision posture |
| --- | --- | --- | --- | --- |
| Current primitives only | Safe, because fetch/write already stays in `ingest_source_url`. | Can work if the agent separately handles search results and placement. | Too ceremonial for routine search-to-source capture. | Reference only. |
| Runner live web search provider | Needs provider abstraction, deterministic fixtures, privacy and egress policy, freshness model, and rate/error handling. | Could find candidates itself. | Good later, too broad now. | Defer. |
| Browser automation or HTTP bypass | Violates runner-owned fetch and public/private boundaries. | Can inspect pages, but outside OpenClerk's installed runner contract. | Unsafe for routine work. | Kill. |
| Read-only search-result planner | Safe if search results are harness-supplied and the runner only ranks, normalizes, deduplicates, and proposes placement. | Coordinates public URL candidates with `ingest_source_url`. | Good; removes prompt choreography while preserving approval boundaries. | Promote. |

## Promoted Candidate

```json
{"action":"web_search_plan","web_search":{"query":"example source","results":[{"url":"https://example.test/page.html","title":"Example","snippet":"Search snippet"}],"limit":10}}
```

The runner does not call a live search provider. It accepts search results from
the harness, validates public HTTP/HTTPS URLs, infers `web` or `pdf`, checks
runner-visible duplicate source URLs, proposes source/asset/synthesis
placement hints, and returns an `agent_handoff`.

## Safety Constraints

- No live search provider call inside the runner.
- No browser automation, HTTP fetch, direct filesystem fetch, or MCP/HTTP
  bypass.
- No durable source write or synthesis write.
- No private/authenticated page handling beyond marking it unsupported.
- Search snippets are discovery hints only, not citations or source evidence.
- Durable fetch/write remains `ingest_source_url` after approval.
- Source claims require citations/source refs, provenance, and projection
  freshness after approved ingestion.

## Promotion And Kill Criteria

Promote if deterministic fixtures show the planner can preserve public-only
boundaries, duplicate URL hints, placement quality, no-fetch/no-write status,
and approval-before-write while improving routine search-to-source UX.

Kill any shape that fetches outside `ingest_source_url`, treats snippets as
canonical evidence, writes before approval, handles private/authenticated
content as routine input, hides duplicate status, or requires direct SQLite,
raw vault inspection, source-built runners, browser automation, HTTP/MCP
bypasses, or unsupported transports.

## Non-Goals

- no live search provider integration
- no browser automation
- no public web fetch outside `ingest_source_url`
- no private/authenticated page acquisition
- no durable writes from search results
- no citations from search snippets
