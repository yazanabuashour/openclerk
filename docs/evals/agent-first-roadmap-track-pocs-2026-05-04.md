# Agent-First Roadmap Track POCs

## Status

Implemented as a reduced, eval-only candidate comparison for the seven
agent-first knowledge-plane roadmap tracks. The executable fixture is
`internal/evals/knowledgeplane/roadmap_tracks_test.go`.

This POC does not add runner actions, schemas, migrations, vector indexes,
memory stores, Git operations, web search providers, parsers, OCR, skill
behavior, public API behavior, or durable writes.

## Source References

- Active local agent-first knowledge-plane architecture note
- Karpathy, LLM Wiki gist, 2026-04-04
- Mitchell Hashimoto, "The Building Block Economy"
- OpenAI API prompt guidance
- OpenAI, "Harness engineering: leveraging Codex in an agent-first world"
- OpenAI API embeddings guide
- OpenAI API retrieval guide
- Mem0 OSS overview

The OpenAI docs MCP was registered but not callable in this session; official
OpenAI web pages were used as fallback.

## POC Contract

The POC uses deterministic candidate matrices to test these invariants:

- every track has at least two candidate shapes
- every acceptable baseline is both safe and capable
- every track keeps product behavior unchanged unless a promotion decision
  names an exact surface
- rejected candidates cannot bypass runner-only, local-first, source
  authority, provenance, freshness, no-bypass, or approval-before-write
  boundaries

## Candidate Matrices

| Track | Baseline Candidate | Deferred or Rejected Candidates | POC Outcome |
| --- | --- | --- | --- |
| Hybrid retrieval | Existing lexical search over runner-visible chunks. | Eval-only hybrid fusion, external vector stores, OpenAI vector stores. | Keep lexical default; hybrid remains reference. |
| Memory recall | Existing read-only memory/router report and canonical evidence retrieval. | Mem0 recall layer, internal memory projection, autonomous memory writes. | Keep read-only report; defer separate memory. |
| Structured stores | Existing schema-backed records, services, and decisions. | Dynamic schemas, broad untyped facts, docs-only for all structured facts. | Select existing schema-backed pattern; defer new domains. |
| Skill reduction | Current thin skill plus runner help and handoffs. | Move more policy into help, remove no-tools policy, long skill recipes. | Keep current skill; defer additional shrink. |
| Git lifecycle | Storage-level Git history outside OpenClerk authority. | Privacy-safe status report, checkpoint action, restore, branch switch, remote push. | Defer runner-owned lifecycle. |
| Web search/fetch | Existing public URL placement and approved runner fetch. | Search planning provider, browser automation, HTTP bypass. | Keep runner fetch; defer search planning. |
| Artifact intake | Proposal-first validation, source placement, duplicate checks, explicit content. | Tag/field autofill planner, parser/OCR claims, opaque artifact inspection. | Keep proposal-first current primitives; defer parser-backed claims. |

## Reduced Fixture Scenarios

The fixture rows intentionally use neutral data and do not commit raw logs,
generated corpora, database files, or private content.

| Scenario | Safety | Capability | UX | Reason |
| --- | --- | --- | --- | --- |
| Current primitives baseline | Pass | Pass | Mixed to acceptable | Existing runner surfaces preserve authority and auditability. |
| Narrow read-only candidate | Pass | Partial to pass | Potentially better | Needs targeted evidence before promotion. |
| Authority-expanding candidate | Fail | Often capable | Not acceptable | Adds hidden storage, parser, memory, vector, web, or Git authority. |

## POC Conclusion

The reduced POC supports closing the ADR, POC, and eval gates for the open
roadmap epics as no-product-change work. It does not justify implementing a
new product surface in this session.

