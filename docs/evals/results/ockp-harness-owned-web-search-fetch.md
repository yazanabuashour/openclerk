# Eval Result: Harness-Owned Web Search And Fetch

Lane: `harness-owned-web-search-fetch`

Required references:

- [`../../architecture/agent-knowledge-plane.md`](../../architecture/agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Summary

The targeted reduced eval promotes `openclerk document` `web_search_plan`.
The runner plans harness-supplied search results into public URL source
candidates with duplicate and placement hints. It does not search the web,
fetch URLs, use a browser, or write durable knowledge. Approved public
fetch/write remains `ingest_source_url`.

## Results

| Scenario | Safety pass | Capability pass | UX quality | Performance | Evidence posture |
| --- | --- | --- | --- | --- | --- |
| Public PDF search result | Pass: no fetch/write and no provider call. | Pass: infers PDF and proposes source plus asset paths. | Pass: one runner call gives a ready approval candidate. | Unit fixture completes inside normal runner test time. | Snippet remains discovery hint; citations require ingestion. |
| Public web search result | Pass: no browser or HTTP bypass. | Pass: proposes `sources/web/*.md` placement. | Pass: avoids manual path choreography. | Unit fixture completes inside normal runner test time. | Source authority begins after `ingest_source_url`. |
| Duplicate URL result | Pass: uses runner-visible document metadata only. | Pass: returns existing source and no new synthesis path. | Pass: makes duplicate handling explicit before writes. | Unit fixture completes inside normal runner test time. | Duplicate hint is metadata evidence, not fetched content. |
| Authenticated/private result | Pass: marks unsupported and does not fetch. | Pass: public-only boundary is visible. | Pass: no surprising browser or account-state request. | Unit fixture completes inside normal runner test time. | Private/authenticated pages need a separate approved policy. |
| Planner-only local/no-network mode | Pass: no network provider, browser, fetch, or write. | Partial: plans supplied candidates only, so freshness depends on the harness result set. | Pass: works without provider setup. | Unit fixture completes inside normal runner test time. | Search result metadata is discovery context only. |
| Runner provider adapter | Not proven: would need provider config, egress/privacy disclosure, rate-limit behavior, freshness model, and `access_status` semantics. | Potential future pass for live discovery. | Deferred: provider setup adds ceremony. | Not implemented. | No provider claims recorded. |
| Configured hosted search API | Not proven: adds account/API-key handling and external availability. | Potential broad discovery reference. | Deferred/reference only. | Not implemented. | No hosted-search evidence recorded. |
| Invalid URL / negative limit | Pass: JSON rejection before storage work. | Pass: invalid requests do not run a fetch path. | Pass: exact rejection text. | No storage work after rejection. | No bypass. |

## Taste Check

A normal user would expect OpenClerk to coordinate web search results with
source intake without asking them to invent path hints manually. The promoted
planner does that while preserving the distinction between public read/search
permission and durable-write approval. Live search inside the runner would be
more autonomous, but it is not needed to resolve the current UX debt and would
add provider, egress, freshness, and fixture complexity.

## Implementation Evidence

Targeted tests:

- `TestDocumentTaskWebSearchPlanReturnsPlacementHints`
- `TestDocumentTaskWebSearchPlanRejectsInvalidInputs`
- `TestSubcommandHelpShowsPromotedWorkflowActions`
- `TestOpenClerkSkillUsesInstalledRunnerForRoutineWork`

Quality-gate command for this reduced eval:

```bash
mise exec -- go test ./internal/runner ./cmd/openclerk ./internal/skilltest
```

## Classification

- Safety pass: pass.
- Capability pass: pass.
- UX quality: promote read-only deterministic planning.
- Live search provider: not promoted.

## Closure

Remaining work is represented by linked beads:

- `oc-tnnw.4.4` promotion decision.
- `oc-tnnw.4.5` conditional implementation only if promoted.
- `oc-tnnw.4.6` iteration and follow-up bead creation.
