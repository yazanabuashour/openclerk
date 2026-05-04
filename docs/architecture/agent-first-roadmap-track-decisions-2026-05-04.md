---
decision_id: decision-agent-first-roadmap-track-decisions-2026-05-04
decision_title: Agent-First Knowledge Plane Roadmap Track Decisions
decision_status: accepted
decision_scope: agent-first-roadmap
decision_owner: platform
decision_date: 2026-05-04
source_refs: docs/evals/agent-first-roadmap-track-pocs-2026-05-04.md, docs/evals/results/ockp-agent-first-roadmap-track-eval-2026-05-04.md
---
# Decision: Agent-First Knowledge Plane Roadmap Track Decisions

## Status

Accepted for roadmap closure: complete the seven evidence-gated tracks as
ADR, POC, eval, decision, conditional implementation, and iteration work
without promoting new product behavior.

The governing local reference is the active agent-first knowledge-plane
architecture note. External references reviewed for this decision:

- Karpathy, LLM Wiki gist, 2026-04-04
- Mitchell Hashimoto, "The Building Block Economy"
- OpenAI API prompt guidance
- OpenAI, "Harness engineering: leveraging Codex in an agent-first world"
- OpenAI API embeddings guide
- OpenAI API retrieval guide
- Mem0 OSS overview

OpenAI docs MCP tooling was not callable in this session after registering the
official MCP server, so the OpenAI references were reviewed from official
OpenAI web pages as fallback.

Evidence:

- [POC: Agent-First Roadmap Track POCs](../evals/agent-first-roadmap-track-pocs-2026-05-04.md)
- [Eval: Agent-First Roadmap Track Eval](../evals/results/ockp-agent-first-roadmap-track-eval-2026-05-04.md)
- Eval-only fixture: `internal/evals/knowledgeplane/roadmap_tracks_test.go`

## Shared ADR

All tracks preserve the same baseline architecture:

- The installed OpenClerk JSON runner remains the production agent boundary.
- Canonical markdown, source-linked synthesis, selected structured records,
  provenance, and projection freshness remain the truth and audit substrate.
- Embeddings, vector stores, memory, graph projections, web search, parser
  outputs, autofill, and Git state are auxiliary infrastructure unless a later
  accepted decision names a narrower authoritative surface.
- Public read, fetch, or inspect permission is separate from durable-write
  approval.
- No routine agent path may bypass the runner through direct storage, direct
  vault inspection, raw database access, browser automation, HTTP/MCP
  substitutes, source-built runners, backend variants, or unsupported
  transports.

## Track Decisions

| Bead | Track | Decision | Product Behavior |
| --- | --- | --- | --- |
| `oc-uj2y.2` | Hybrid embedding and vector retrieval | Defer hybrid/vector promotion and keep lexical retrieval as default. The eval-only hybrid fixture remains reference evidence. | No runner, schema, vector store, embedding index, storage, or skill change. |
| `oc-uj2y.3` | Memory architecture and recall | Defer a separate memory layer. Keep the existing read-only memory/router report path as the only acceptable memory-adjacent public surface. | No Mem0 integration, memory transport, remember/recall write action, autonomous routing, or storage change. |
| `oc-uj2y.4` | Structured data and non-document stores | Select current schema-backed promoted domains as the viable pattern. Defer new domains until a specific schema and migration policy beat docs-only storage. | No generalized schema, dynamic records, metrics store, or new domain implementation. |
| `oc-uj2y.5` | Skill reduction into runner heuristics | Defer additional skill shrink. The skill is already compact enough to protect safety, and further reduction needs workflow-specific evidence. | No skill, runner help, or handoff behavior change. |
| `oc-uj2y.6` | Git-backed version control and lifecycle | Defer runner-owned Git lifecycle behavior. Treat Git as storage-level history only. | No checkpoint action, branch switching, remote push, restore, rollback, raw diff, or review queue. |
| `oc-uj2y.7` | Harness-owned web search and fetch | Defer web search planning. Keep public URL placement/fetch under existing runner-owned intake. | No web search provider, browser fetch, external HTTP bypass, or new search action. |
| `oc-uj2y.8` | Artifact intake, auto-filing, tags, and fields | Select proposal-first current primitives for supported explicit content and runner-supported public fetches. Defer parser/OCR and opaque artifact claims. | No parser, OCR, file-inspection, tag store, metadata schema, or new artifact action. |

## Safety, Capability, UX

Safety pass: pass for all tracks. The reduced POC and eval evidence preserves
runner-only local-first access, source authority, citations/source refs,
provenance, projection freshness, no-bypass boundaries, and
approval-before-write.

Capability pass: pass for the current primitives and selected existing
surfaces. Hybrid retrieval, separate memory, new structured domains, more skill
shrink, Git lifecycle actions, web search planning, and parser-backed artifact
autofill remain unproven as promoted product behavior.

UX quality: mixed. Existing lexical retrieval, schema-backed records, runner
handoffs, source placement, duplicate checks, and memory/router reports cover
many workflows. The deferred shapes still carry possible UX value, but the
current evidence does not justify widening authority, storage, or durable-write
behavior.

## Taste Review

Normal users would reasonably expect OpenClerk to become simpler over time:
fewer manual evidence steps, clearer placement suggestions, less skill
choreography, and safer recall or lifecycle summaries. That expectation is
valid. The evaluated shapes do not all fail because the needs are invalid;
they fail because promotion would currently add authority, storage, provider,
parser, or workflow risk before the exact safe surface is proven.

Each track therefore records capability need separately from product
promotion. The matching iteration beads close as no-op gates with this report
as the candidate-surface comparison. Future work should start from the
candidate matrices here instead of reopening broad architecture questions.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the supported public
  runner domains.
- Existing lexical search remains the default retrieval path.
- Existing read-only reports and handoffs remain read-only.
- Existing schema-backed record, service, and decision projections remain the
  only promoted structured domains.
- Existing public URL intake remains runner-owned and approval-gated.
- Opaque artifact parsing, OCR, local file inspection, autonomous memory
  writes, remote Git actions, browser or HTTP bypasses, and external vector
  authority remain unsupported.

