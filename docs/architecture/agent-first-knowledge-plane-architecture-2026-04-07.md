---
id: "architecture/agent-first-knowledge-plane"
title: "Research: Agent-First Knowledge Plane Architecture (2026-04-07)"
type: "research"
status: "active"
modality: "markdown"
updated_at: "2026-05-04T00:00:00Z"
tags: ["research","architecture","vault","mem0","knowledge-plane","agentops","provenance","truth-maintenance","llm-wiki","source-linked-synthesis"]
aliases: ["Agent-first knowledge architecture","Knowledge plane architecture","Docs records memory router","LLM Wiki alignment"]
---
# Research: Agent-First Knowledge Plane Architecture (2026-04-07)

This note documents the target architecture for an agent-first personal
knowledge system built around an operator-managed knowledge vault and the
OpenClerk AgentOps JSON runner.

It exists to answer one specific concern:

- how to improve the agent-facing knowledge stack
- without building an elaborate system that ends up no better than simple RAG over documents plus a few hand-written indexes

This note treats that concern as a hard design constraint, not a footnote.

Related repo references:

- [Agent Knowledge Plane](agent-knowledge-plane.md)
- [AgentOps-Only Knowledge Plane Direction](eval-backed-knowledge-plane-adr.md)
- [OpenClerk Next-Phase Maturity Validation](openclerk-next-phase-maturity-validation-decision.md)
- [Deferred Capability Promotion Gates](deferred-capability-promotion-gates.md)

## May 4, 2026 Roadmap Tracker Update

Beads epic `oc-uj2y` tracks the next agent-first knowledge-plane roadmap. The
roadmap is intentionally evidence-gated: each capability track must proceed
through ADR, POC, eval, promotion decision, conditional implementation, and
iteration before product behavior changes.

The active tracks are:

- hybrid embedding and vector retrieval
- memory architecture and recall
- structured data and non-document canonical stores
- skill reduction into runner heuristics
- Git-backed version control and lifecycle
- harness-owned web search and fetch
- artifact intake, auto-filing, tags, and fields

This tracker does not promote any new runner action, storage layer, parser, web
search behavior, memory system, vector store, or durable-write surface by
itself. The production boundary remains the installed OpenClerk JSON runner;
routine agents must not bypass it through direct SQLite, source-built binaries,
HTTP/MCP variants, private implementation paths, or raw vault/file inspection.

For every non-promotion outcome, the decision bead must separate safety pass,
capability pass, and UX quality. If a real workflow need remains, the decision
must run `bd search` for existing follow-up work and create or link
candidate-surface comparison Beads before closing or handing off.

## April 2026 Direction Update

OpenClerk is now on the AgentOps JSON runner path. The production agent
interface is the installed `openclerk` runner plus its skill guidance. The
runner is the substrate for document, retrieval, synthesis, records,
provenance, and freshness work.

The active question is how to structure an agent-first knowledge plane behind
AgentOps: canonical docs, source-linked synthesis, selective promoted records,
provenance/truth sync, and later memory and routing.

The next build target is the source-linked synthesis lifecycle. The POC should
turn `synthesis/` into durable agent-maintained wiki pages that search
canonical sources first, update existing synthesis rather than duplicate it,
preserve source refs and freshness, handle stale or contradictory claims, and
file reusable answers back into markdown.

## April 22, 2026 Validation Update

The first OpenClerk architecture slice is now eval-backed, but the full future
architecture is not "proven" in one jump.

Validated for v1:

- Phase 0 is complete: the installed `openclerk` JSON runner is the production
  AgentOps surface for routine document and retrieval workflows.
- The docs/synthesis/provenance/records slice is proven enough for v1:
  canonical docs, source-linked synthesis, promoted records, provenance events,
  projection freshness, and final-answer-only rejection gates are covered by
  current production and targeted AgentOps evidence.
- Knowledge configuration v1 is accepted through `oc-za6`: convention-first
  layout plus runner-visible `inspect_layout`, not a committed manifest.
- Decision records were the promoted follow-up domain and were hardened through
  `oc-j0a`.

Still deferred:

- Mem0 or a memory API
- autonomous router
- semantic graph layer as truth
- broad contradiction engine
- new public runner actions

Those deferred capabilities should not be promoted until targeted evals show
the existing `openclerk document` and `openclerk retrieval` actions are
structurally insufficient while preserving citations, source refs, provenance,
freshness, and bypass prevention.

The next validation step is not more architecture layering. It is a real-vault
trial against representative operator-vault workflows plus a separate
promotion-gates rubric for deferred capabilities.

## April 23, 2026 Post-v0.1.0 Document Lifecycle Update

The v0.1.0 slice is sufficient to ship the first OpenClerk production surface:
`openclerk document`, `openclerk retrieval`, source-linked synthesis,
provenance, projection freshness, decision records, and real-vault validation
all have eval-backed evidence.

The next maturity gap is not memory, autonomous routing, semantic graph truth,
or a broad contradiction engine. The more important post-v0.1.0 product
direction is document lifecycle control for agent-authored durable knowledge.

Version and history control are not human-world holdovers. They become more
important when agents write durable knowledge because the system needs to
answer what changed, why it changed, what source evidence justified the change,
which previous content was replaced, and how an unsafe edit can be reviewed or
restored.

OpenClerk should treat Git, sync providers, and filesystem snapshots as
storage-level history. It should add semantic document history only when
dogfooding and targeted AgentOps evals justify it:

- revision records for OpenClerk-authored document changes
- content hashes for before and after states
- edit summaries, actor/source metadata, and source refs
- diff inspection that does not expose raw vault content in public artifacts
- restore or rollback semantics for OpenClerk-managed changes
- review queues for agent-authored changes before they become accepted
  knowledge

This should remain a post-v0.1.0 capability, not a reason to delay the first
release. The repo-side product vision is recorded in
`docs/architecture/openclerk-document-post-v0.1.0.md`.

## Executive Summary

The architecture should not be:

- markdown everywhere
- or Mem0 everywhere
- or "embed everything and hope vector search is enough"
- or a multi-store system with no clear truth-maintenance model

The architecture should be:

0. **AgentOps JSON runner**
   the single production agent substrate for document and retrieval workflows
1. **canonical docs**
   narrative, research, project context, runbooks, and source-linked synthesis remain in markdown
2. **canonical records**
   schema-first domains move into structured storage only when they clearly benefit from it
3. **memory**
   Mem0 stores distilled recall, not canonical truth
4. **truth sync / provenance**
   a thin but explicit subsystem keeps cross-store truth coherent over time
5. **router**
   a thin, auditable layer decides which system to read from or write to
6. **quality gates**
   every new capability behind AgentOps must preserve source authority,
   citation correctness, and freshness before it becomes durable product

The key idea is:

- agents should interact with a **knowledge plane**
- through AgentOps JSON tasks rather than folders and files as their primary interface
- and useful synthesis should compound as source-linked durable knowledge instead of disappearing into chat history

But the underlying canonical truth still matters. The agent-facing interface becomes runner-first. The source-of-truth layer does not need to become memory-native.

Karpathy's LLM Wiki pattern is related to this direction. It reinforces the value of a persistent, agent-maintained markdown synthesis layer, but OpenClerk should adopt that idea as a provenance-backed docs-layer workflow rather than as a separate authority system.

The missing problem that must be solved explicitly is:

- not just where different kinds of knowledge live
- but how truth stays coherent across docs, records, and memory as those stores evolve over time

If that is left implicit, the likely failure mode is predictable:

- docs say one thing
- records say a newer thing
- memory recalls an older abstraction
- the router returns whichever answer is fastest
- operator trust collapses

## 1. Problem To Solve

The historical vault design solved a human problem:

- how to write
- organize
- link
- browse
- and retrieve knowledge manually

That produced a human-first substrate:

- markdown notes
- folder hierarchy
- indexes
- links
- formatting conventions

Now the primary interface is increasingly an agent. That changes the retrieval surface, but it does **not** eliminate the need for canonical truth, provenance, inspectability, or structured state.

The real problem is now:

- how should canonical knowledge, structured records, and agent memory coexist
- so that the agent can answer and act correctly
- without exploding complexity

The danger case is obvious:

- building a multi-layer "agent-native" architecture
- that still ends up functioning like basic chunked RAG over docs
- but with more maintenance, more failure modes, and more persistence systems

A second danger case is equally important:

- building cleanly separated docs, records, and memory stores
- but leaving cross-store truth maintenance undefined
- so the system drifts into conflicting representations over time

This architecture is designed to avoid both failures.

## 2. Design Constraints

These are the hard constraints that shape the design.

### Constraint A: canonical truth must remain inspectable

When the system is wrong, an operator needs to inspect what the agent relied on, repair it, and understand the chain of truth.

That requires:

- source documents
- provenance
- stable identifiers
- deterministic storage
- safe update paths
- inspectable derivation history

### Constraint B: not all knowledge is the same shape

Some things are naturally documents:

- research notes
- project context
- reasoning
- runbooks
- decisions with rationale

Some things are naturally records:

- asset, part, and maintenance records
- lab, measurement, and observation results
- contact-style entities
- recurring preferences with stable fields
- assets with metadata

Some things are naturally memory:

- stable user preferences
- repeated facts
- cross-session decisions
- distilled recall

One storage model should not be forced to do all three jobs.

### Constraint C: complexity must justify itself empirically

The default answer is not "build the elaborate system."

The default answer is:

- keep the AgentOps contract small
- add one knowledge-model capability only when it improves reliability without
  weakening source authority or freshness

### Constraint D: the first router must be deterministic

The initial system should not depend on an autonomous classifier making opaque storage decisions.

The first version should use explicit policies and audit logs.

### Constraint E: cross-store truth must have an explicit lifecycle

If a fact can appear in more than one place, the system needs a defined way to answer:

- what is the authoritative source
- what derived representations came from it
- what becomes stale when the source changes
- how supersession and retraction work
- how time-sensitive facts are interpreted

If these rules are missing, the architecture becomes elegant-looking multi-store drift.

## 3. Non-Goals

This architecture is explicitly **not** trying to do these things:

- replace all markdown with memory entries
- replace canonical notes with vector chunks
- convert every domain into a database schema up front
- build a knowledge graph because knowledge graphs are fashionable
- make the agent the sole owner of truth
- create a many-database system before a benchmark shows a need
- maintain more hand-written navigation than necessary
- rely on implicit latest-write-wins semantics for all factual state
- treat memory as authoritative when source-sensitive answers require canonical evidence
- clone an LLM-maintained wiki pattern literally without stable ids, citations, freshness, and evals
- let agent-authored synthesis outrank canonical sources or promoted records

## 4. Architecture Overview

The target architecture is a **knowledge plane** with six logical components.

| Layer | Role | Canonical? | Primary object |
| --- | --- | --- | --- |
| Docs | Canonical narrative knowledge and source-linked synthesis | Yes | document / section / chunk |
| Records | Canonical structured state | Yes | entity / row / relation |
| Memory | Distilled long-term recall | No | memory item |
| Truth Sync / Provenance | Cross-store coherence and lifecycle | No | source ref / event / validity state |
| Router | Read/write orchestration | No | request / action |
| Evals | Measurement and gates | No | benchmark run |

The important distinction is:

- **docs and records are truth systems**
- **memory is a recall system**
- **truth sync keeps linked representations coherent over time**

The architecture is not complete unless all three claims are true.

### LLM Wiki As A Related Pattern

Karpathy's LLM Wiki gist describes a personal knowledge pattern with three layers:

- immutable raw sources
- an LLM-maintained markdown wiki
- a schema or instruction file that tells the LLM how to maintain the wiki

The core insight is directly relevant:

- pure query-time RAG keeps rediscovering knowledge from scratch
- a persistent wiki lets synthesis, cross-links, contradictions, and summaries compound over time
- useful answers can be filed back into the knowledge base instead of being lost in chat history
- the human curates sources and directs analysis while the LLM handles maintenance work

The closest OpenClerk mapping is:

| LLM Wiki concept | OpenClerk interpretation |
| --- | --- |
| Raw sources | canonical source docs and assets with stable ids and provenance |
| Wiki | accepted canonical notes plus source-linked synthesis, topic, entity, and comparison pages |
| Schema | repo docs, agent instructions, and the OpenClerk skill payload |
| `index.md` | search, metadata filters, graph neighborhoods, and optional generated index pages |
| `log.md` | provenance events, projection states, and optional human-readable activity notes |

The similarities are important:

- both reject pure query-time RAG as the complete answer
- both favor durable markdown knowledge that compounds over time
- both put the LLM in charge of summaries, links, filing, contradiction checks, and maintenance
- both rely on explicit operating instructions so the agent behaves like a disciplined maintainer

The differences are just as important:

- Karpathy's pattern is an abstract workflow; OpenClerk is a local-first SDK, API contract, and runtime
- the LLM Wiki layer is largely LLM-owned; OpenClerk should distinguish raw sources, accepted canonical notes, source-linked synthesis, promoted records, and derived projections
- `index.md` and `log.md` can work at moderate scale; OpenClerk should expose those roles through search, provenance, projection state, and evals
- Karpathy does not strongly separate records and memory; OpenClerk should keep records canonical for selected structured domains and reserve memory for future recall-only promotion

The right adoption path is therefore:

- support LLM-maintained synthesis inside the docs layer
- require source refs, citations, and freshness state for synthesis that answers source-sensitive questions
- treat synthesis as durable compiled knowledge, not as a new authority above canonical sources
- benchmark whether the workflow improves real agent tasks before adding new public API surface

### Cognee As Reference Architecture

Cognee is a useful reference architecture for graph/vector AI memory systems,
but it points in a different product direction than OpenClerk. It validates
that serious agent memory systems need more than vector chunks:

- graph retrieval
- vector retrieval
- relational metadata
- session memory
- feedback weighting
- ontology and entity grounding
- temporal retrieval modes

It also demonstrates the drift risk this architecture is designed to avoid.
A memory-first interface can make derived graph or session state feel
canonical unless provenance, source authority, and freshness are explicit.

OpenClerk should borrow these ideas as benchmark categories and later internal
design inputs:

- retriever taxonomy across chunks, graph, summaries, temporal retrieval, and
  exact lexical lookup
- ontology or entity-disambiguation techniques for duplicate and alias-heavy
  domains
- temporal retrieval as an eval category for current, observed, effective, and
  superseded facts
- feedback weighting as a possible later ranking signal
- session-to-durable promotion as a future workflow pattern

OpenClerk should reject these Cognee-style moves for now:

- memory-first product language as the primary agent interface
- broad public surfaces beyond AgentOps
- graph as an independent truth system
- automatic durable session memory without canonicalization

Any Cognee-inspired capability should have strict gates:

- graph projection must improve navigation without reducing citation
  correctness
- ontology grounding must reduce entity duplicates without inventing authority
- temporal retrieval must surface source time semantics
- feedback weighting must remain auditable and source-linked
- memory or session promotion must mark stale and superseded state when
  canonical truth changes

## 5. The Docs Layer

### Role

The docs layer remains the home for:

- raw or source documents
- research
- project notes
- runbooks
- rationale
- narrative reference notes
- source-linked synthesis pages
- topic, entity, comparison, and overview pages when they remain source-grounded
- assets and their derived text

### Canonical form

The existing vault model stays largely valid:

- markdown remains canonical for narrative knowledge
- assets remain in place
- derived text/manifests remain machine-generated
- accepted synthesis pages are markdown docs that carry source references and freshness metadata

Source-linked synthesis is the OpenClerk version of the LLM Wiki's generated wiki layer.

It should be allowed to accumulate:

- topic summaries
- entity pages
- comparisons
- evolving theses
- filed answers from useful queries
- contradiction and uncertainty notes

But it must not become a free-floating authority layer.

For source-sensitive claims, synthesis pages should point back to:

- canonical source docs
- chunks or heading locators
- promoted record ids where a structured domain is authoritative
- provenance events that explain when the synthesis was created or refreshed

### Best agent-facing interface

The agent should not navigate raw markdown files as its primary interface.

The best agent-facing interface for docs is a **docs service** over the vault with:

- hybrid lexical + vector search
- chunk retrieval with citations
- heading-aware retrieval
- link/backlink expansion
- metadata filtering
- recency and quality signals
- explicit read and write APIs

That means:

- markdown stays canonical
- navigation becomes derived

### Minimum docs service capabilities

1. document registry keyed by stable note `id` and vault-relative path
2. chunk store with deterministic chunk ids
3. lexical index
4. embedding index
5. link graph derived from markdown links
6. source lineage from raw asset to derived text
7. citation-bearing retrieval results
8. safe write operations:
   - create canonical note
   - append capture to existing note
   - update frontmatter-aware sections
   - create or update source-linked synthesis from cited evidence

### Required identity model

The docs layer must expose stable references that other layers can point to:

- `doc_id`
- `chunk_id`
- canonical vault-relative path
- section or heading locator where needed

If memory or records cannot point back to these identifiers, provenance becomes hand-wavy.

### Why this is not "just vector DB"

The docs layer exists to solve problems that pure vector retrieval does not solve well:

- exact-term lookup
- source citations
- stable document identity
- safe writes
- asset lineage
- link-context expansion
- provenance and repair

If the final implementation were only "pipe markdown through embeddings into a vector DB," then this architecture would have failed its own standard.

### Retrieval efficiency is an implementation concern, not a memory-model change

Recent work on vector compression is relevant here, but only at the implementation layer.

For example, Google Research's **TurboQuant** work targets two bottlenecks that matter for agent-facing docs systems:

- compressing high-dimensional vector indexes used for semantic retrieval
- compressing key-value cache state used for long-context model serving

This matters if the vault grows into:

- larger multimodal corpora
- heavier PDF/manual indexing
- large local or hosted vector indexes
- longer-context source-grounded retrieval flows

But it does **not** change the logical architecture in this note.

TurboQuant does not alter:

- the distinction between canonical docs, canonical records, and promoted memory
- the need for provenance and explicit source authority
- the requirement for hybrid retrieval rather than vector-only retrieval
- the need to gate every extra layer on source authority, citation correctness,
  and freshness

The right interpretation is narrower:

- embedding-model choice changes semantic quality, modality coverage, and vector shape
- vector-compression techniques such as TurboQuant change storage footprint, indexing cost, and retrieval/runtime efficiency

Embedding-model choice should be handled as a separate benchmark input rather
than as part of the canonical knowledge-model decision.

If adopted later, this kind of compression should remain optional and benchmark-gated. A compressed embedding index is only useful if it reduces memory and latency without creating unacceptable regression in:

- retrieval recall
- citation correctness
- provenance traceability
- operator trust in source-grounded answers

### Optional derived graph projection

The docs layer may also expose a **derived graph projection** over canonical documents and their derived artifacts.

This graph is not a separate truth system.

It exists to support agent-facing operations that are awkward in plain chunk retrieval alone, such as:

- path queries between concepts, decisions, and assets
- neighborhood expansion around an entity or note
- bridge-node detection across otherwise separate topics
- community summaries over a large corpus
- structure-first navigation and question suggestion

The graph should remain explicitly **derived** from canonical docs and assets.

That means:

- docs remain authoritative for narrative truth and evidence
- the graph never becomes the sole source of truth
- every graph node and edge must retain `evidence_ref` values back to canonical `doc_id`, `chunk_id`, heading locators, or asset lineage records
- graph artifacts must be refreshable and invalidatable when canonical sources change

A useful initial graph model is:

- node types:
  - `document`
  - `chunk`
  - `heading`
  - `entity`
  - `rationale`
  - `asset`
  - optional `record_ref`
  - optional `memory_ref`
- edge types:
  - `links_to`
  - `mentions`
  - `cites`
  - `derived_from`
  - `rationale_for`
  - `semantically_related_to`
  - `participates_in`

The graph model should also allow **hyperedges** or equivalent n-ary group relations for cases where pairwise edges lose meaning, such as:

- one design decision tying together several constraints
- one workflow involving several steps or components
- one rationale fragment explaining several related concepts

Every non-trivial derived relationship should carry explicit epistemic labeling:

- `EXTRACTED`
  directly supported by source structure or text
- `INFERRED`
  derived by rules or model reasoning
- `AMBIGUOUS`
  uncertain enough that operator review may be needed

`INFERRED` and `AMBIGUOUS` relationships should carry `confidence_score`.

This projection complements lexical retrieval, embedding retrieval, and link expansion. It does not replace them.

## 6. The Records Layer

### Role

The records layer exists for domains where documents are a poor canonical shape.

Representative knowledge domains that may qualify:

- assets, parts, and repairs
- measurement or lab records
- contact-style entities
- recurring purchasing data
- possibly preferences that have stable fields and update semantics

### When to promote a domain to records

A domain should move out of pure markdown only when at least one of these is true:

1. the data has stable fields and constraints
2. updates are frequent and should be merged, not appended narratively
3. the agent needs deterministic filters and joins
4. duplicate/conflicting entries are a real problem
5. queries are naturally entity-centric rather than document-centric

### Storage model

Mem0 is **not** this layer.

Mem0 provides memory storage and retrieval plumbing, but not an OpenClerk
domain schema. OpenClerk still needs its own tables, ids, constraints, and
relationships for canonical structured data.

Good starting options:

- SQLite for local-first and operational simplicity
- Postgres if stronger relational semantics, concurrent access, or growth room
  are required

### Initial principle

Do not create a records system for everything.

Only promote domains that fail as documents.

### Required identity and version model

Records must expose stable identifiers and update history semantics:

- `entity_id`
- row or relation identifiers as needed
- timestamps for observed/effective changes
- supersession or version markers where the domain requires history

If records are treated as mutable rows with no version semantics, they will be hard to reconcile against docs and memory.

## 7. The Memory Layer

### Role

Mem0 fits here.

Use it for:

- stable user preferences
- repeated facts worth recalling quickly
- decisions
- distilled conclusions
- long-lived constraints
- cross-session personalization

### What Mem0 is not

Mem0 is not:

- canonical notes
- canonical records
- source-grounded document storage
- a substitute for OpenClerk-owned schema

### Canonical write rule

If a fact has a canonical home in docs or records, write it there first or at least ensure it exists there.

Then promote it to Mem0 when recall value is high.

That means:

- canonical first
- memory second

### Promotion rule

Only promote into Mem0 when one of these is true:

1. the fact is likely to be asked again
2. the fact should shape future agent behavior
3. the fact is annoying to rediscover from documents every time
4. the fact is stable enough to survive abstraction

### Required provenance model

Memory items must not float free.

Each promoted memory item should retain at least:

- `memory_id`
- one or more `source_ref` values pointing to canonical docs, records, or events
- promotion timestamp
- freshness state
- optional derivation summary of why it was promoted

### Memory freshness states

At minimum, a memory item should support:

- `fresh`
- `stale`
- `superseded`
- `revalidated`

If the system cannot mark memory as stale when canonical truth changes, memory will slowly become agent-maintained folklore.

## 8. The Truth Sync / Provenance Layer

### Role

This is the missing subsystem that makes the rest of the architecture credible.

Its job is to keep cross-store truth coherent and explainable over time.

It answers questions like:

- what canonical source a memory item came from
- what record was extracted from which document
- what downstream representations became stale after an update
- whether a recalled fact is still valid now
- which version of a fact was true at a given time

### Why this must be a distinct logical component

The router decides where to read and write.

That is not the same job as maintaining lifecycle, provenance, supersession, and invalidation across stores.

If those responsibilities are left implicit inside routing logic, they will be inconsistently applied and hard to audit.

### Minimum capabilities

1. stable cross-store references
2. append-only event log for knowledge lifecycle changes
3. source authority and precedence rules
4. invalidation and refresh rules for derived memory and indexes
5. temporal semantics for facts that change over time
6. auditability for operator inspection and repair

### Stable cross-store reference model

The system should standardize a small set of reference types:

- `source_ref`
- `doc_id`
- `chunk_id`
- `entity_id`
- `memory_id`
- `event_id`

Not every object needs every field, but cross-store links must use a shared reference vocabulary.

### Event model

The first implementation should be an append-only ledger, not an autonomous reconciliation engine.

A useful minimum event family is:

- `created`
- `updated`
- `superseded`
- `retracted`
- `promoted_to_memory`
- `memory_marked_stale`
- `memory_refreshed`
- `record_extracted_from_doc`

This does not require a complex event-sourced application architecture. It does require an inspectable audit spine.

### Concrete defaults

Likely first implementation defaults:

- SQLite for the first provenance ledger if the system remains local-first
- Postgres only if concurrency or growth justifies it
- append-only event rows plus reference tables
- no autonomous schema invention
- no opaque cross-store reconciliation agent in the first version

### Source authority rules

The system needs explicit precedence rules.

Recommended defaults:

- for record-owned factual state in promoted domains, records are authoritative
- for rationale, narrative context, and source-grounded evidence, docs are authoritative
- memory is a recall accelerator, not an authority source, for source-sensitive questions
- when memory conflicts with canonical stores, canonical stores win and memory should be marked stale or superseded

### Invalidation and refresh

When canonical truth changes, derived representations must not be assumed valid.

Required behavior:

1. canonical update emits an event
2. linked memory items become `stale` unless revalidated
3. derived chunks/indexes are refreshed from canonical sources
4. retrieval either prefers refreshed memory or falls back to canonical retrieval
5. stale memory usage is visible in logs and evaluable

### Derived graph invalidation rules

Because the docs layer may expose a derived graph projection, invalidation rules must cover graph artifacts as well as chunks and indexes.

Required behavior:

1. canonical doc or asset update emits an event
2. dependent graph nodes, edges, and hyperedges become stale until rebuilt or revalidated
3. graph artifacts preserve `projection_version` or equivalent derivation metadata
4. retrieval can still use stale graph artifacts only if their status is visible and canonical evidence remains accessible
5. graph-derived answers must surface whether the relevant structure was `EXTRACTED`, `INFERRED`, or `AMBIGUOUS`

Authority remains unchanged:

- graph structure is a derived navigation and reasoning aid
- canonical docs and records still determine truth
- graph outputs may suggest where to look, but they do not outrank canonical evidence

### Temporal semantics

The architecture must distinguish between at least these time concepts where relevant:

- `observed_at`
- `effective_at`
- `superseded_at`
- retrieval-time `current` truth

Without this, queries like these become unreliable:

- what are the latest measurement results
- what part am I currently using
- what did we decide at the time
- what is the current preference

The wrong default is silent latest-write-wins.

### What this layer should not do at first

- autonomous reconciliation across stores without auditability
- invent schemas or entity models dynamically
- allow uncontrolled dual-write fanout
- replace the canonical systems themselves
- become a heavyweight distributed data synchronization platform

## 9. The Router

### Role

OpenClerk needs a routing layer if it separates docs, records, and memory.

But the first router should be deliberately thin.

The router depends on the truth sync / provenance rules. It should not own them implicitly.

### Read routing

The router classifies requests into four practical shapes:

1. **document-grounded**
   "show me the note / source / evidence"
2. **record lookup**
   "what part was purchased?" or "what are the latest measurement results?"
3. **memory recall**
   "what do I prefer?" or "what did we decide?"
4. **mixed**
   recall first, then verify from docs or records

### Write routing

The first write router should be policy-based:

- research, rationale, runbooks, project context -> docs
- structured domain updates -> records
- durable preferences and distilled recall -> Mem0
- ambiguous/raw capture -> inbox or staging

### Promotion path

The router may do dual writes only in controlled ways:

1. write canonical entry
2. create provenance references
3. optionally promote memory
4. log why promotion happened

### What the router should not do at first

- silent autonomous schema invention
- opaque multi-store fanout
- uncontrolled duplication across stores
- unlogged memory promotion
- invent its own precedence rules separate from the truth sync subsystem

## 10. Query Flow

Recommended query flow:

1. classify intent
2. choose primary source:
   - docs
   - records
   - memory
3. if memory is used, check freshness and source refs
4. retrieve from primary source
5. if needed, verify or enrich from a second canonical source
6. return answer with provenance when the question is source-sensitive
7. when the answer is durable and useful, file it back as source-linked synthesis only with explicit source refs

Examples:

- "what part was needed for this maintenance issue?"
  - records first if that domain exists there
  - otherwise docs first
  - memory can serve as a fast path only if it is still fresh and source-linked

- "what was decided about browser-control policy?"
  - memory first for fast recall
  - docs second for citation-quality confirmation

- "show me the source note about the maintenance issue"
  - docs only

- "what are the latest measurement results?"
  - records first if measurements are a promoted record domain
  - if memory recalls a previous result, it must still defer to current canonical state

## 11. Write Flow

Recommended write flow:

1. capture incoming content
2. classify storage target
3. canonicalize into docs or records
4. emit provenance/event entries
5. optionally promote summary/fact into memory
6. mark downstream validity state as needed
7. emit audit event

For LLM Wiki-style workflows, the write flow also needs a synthesis path:

1. search existing docs and synthesis before creating a new page
2. attach source refs to every source-sensitive claim
3. update existing topic/entity/comparison pages with append or section replacement when possible
4. mark synthesis freshness when canonical sources change
5. file useful query answers as synthesis only when they are reusable beyond the current chat

The canonicalization step matters more than the memory step.

If that step is skipped, the system eventually turns into agent-maintained folklore.

### Staging rule

Raw or ambiguous captures should not jump directly into long-term memory.

They should go to:

- inbox
- staging note
- or another explicit pre-canonical location

Only after canonicalization should they become promotion candidates.

## 12. Quality Framework

This is the most important part of the architecture.

### Principle

No new capability becomes durable product because it sounds more agent-native.

It becomes durable product only if it preserves source authority, citation
correctness, freshness, and operator repairability on real tasks.

### AgentOps contract

All quality work runs through AgentOps:

- the installed `openclerk` runner
- document and retrieval JSON requests
- skill guidance that rejects routine bypasses before tools
- stable source refs, citation paths, provenance events, and projection states

Quality gates should measure the knowledge model behind the runner, not invent a
new agent-facing path.

### Task categories

The task set should use real questions across these classes:

- source ingest and synthesis update
- source-grounded document retrieval
- exact fact lookup
- query-to-note filing
- repeated recall
- structured domain lookup
- contradiction and stale-claim linting
- orphan or missing-cross-link detection
- ambiguous routing
- write/update correctness
- cross-store freshness and invalidation
- historical vs current-state questions
- graph-native navigation and path queries
- bridge-node and community-summary usefulness
- rationale-cluster retrieval

### Metrics

Measure at least:

- answer accuracy
- citation correctness
- routing correctness
- write correctness
- latency
- index memory footprint
- index build time and rebuild cost
- retrieval recall under vector compression
- citation regression under vector compression
- duplicate/conflict rate
- maintenance overhead
- wiki health:
  - contradictions surfaced
  - stale synthesis detected
  - orphan pages detected
  - missing source refs detected
  - missing cross-links detected
  - useful chat outputs filed when they should persist
- operator trust:
  - how often operators had to inspect and repair the result
- truth drift rate
- stale-memory hit rate
- provenance completeness
- supersession/invalidation correctness
- graph path usefulness
- graph explanation faithfulness to canonical sources
- inferred-edge precision
- ambiguous-edge review burden
- hyperedge usefulness vs noise

### Synthesis lifecycle gate

The active POC is source-linked synthesis lifecycle. It should prove that
`synthesis/` can work as durable agent-maintained wiki space.

Required behavior:

- search canonical sources before writing synthesis
- list existing synthesis pages before creating a new one
- retrieve existing synthesis before updating it
- prefer section replacement or append over duplicate creation
- preserve `type: synthesis`, `status: active`, `freshness: fresh`,
  `source_refs`, `## Sources`, and `## Freshness`
- handle newer contradictory or superseding sources by updating the existing
  synthesis and naming current and superseded evidence
- inspect provenance and projection freshness when promoted records or services
  shape the synthesis

Later capabilities should have similarly concrete gates:

- Mem0 improves repeated recall without increasing truth drift
- records improve precision and update safety for entity-centric domains
- provenance logic reduces stale-memory and conflicting-truth failures
- graph projection improves structure-first navigation without obscuring
  canonical evidence

For the optional derived graph projection, the adoption gate should be strict:

- it should improve structure-first navigation, cross-note discovery, or question formulation materially over simpler docs retrieval
- it should not reduce citation correctness
- it should not obscure source authority
- it should not create a second implicit truth layer

If it behaves mainly like a more complicated way to do docs retrieval, it should remain optional or be removed.

### Kill criteria

Stop or roll back a layer if:

- it mostly duplicates existing AgentOps docs retrieval
- it introduces too much operator overhead
- it creates conflicting truths across stores
- it cannot explain its provenance
- it improves one workflow class while degrading the core use case
- it fails to reduce the drift and invalidation problems it was introduced to solve

## 13. Recommended Rollout

This should be phased and reversible.

### Phase 0: Lock the AgentOps direction

Keep the docs, ADR, skill guidance, and eval protocol aligned around:

- installed `openclerk` JSON runner as the production agent substrate
- document and retrieval task shapes as the public machine contract
- bypass rejection as product-contract enforcement
- no new public agent path for routine knowledge work

This is the anti-drift step for project direction.

### Phase 1: Deepen source-linked synthesis lifecycle

Start with the docs/synthesis layer because it is the least disruptive and most
broadly useful.

Goals:

- durable `synthesis/` pages maintained by agents
- search-before-write behavior
- update-over-duplicate behavior
- stale and contradictory claim repair
- filed answers with source refs
- freshness sections backed by runner retrieval, provenance, or projection
  checks

Use existing AgentOps document and retrieval actions first.

### Phase 2: Add the provenance / truth-sync substrate

Before broad memory adoption, add the minimum coherence substrate:

- stable reference model
- append-only event log
- authority rules
- invalidation rules
- temporal semantics

This is the anti-drift step.

### Phase 3: Add Mem0 as the recall layer

Integrate Mem0 only for promoted facts and personalization.

Goals:

- faster repeated recall
- less repetitive rediscovery
- cross-session personalization

### Phase 4: Promote selected domains to records

Only after AgentOps-backed docs and synthesis show that a domain needs
structured state.

Likely first candidates:

- asset / parts / repairs
- measurement or lab records

### Phase 5: Add thin routing and promotion policies

Only after the individual stores and the coherence substrate work acceptably on their own.

The router should begin as deterministic policy plus audit log.

## 14. Build Behind AgentOps

This architecture does not require building every internal subsystem from
scratch, but routine agent work should stay behind AgentOps.

### Docs and synthesis layer

- use the current OpenClerk document and retrieval runner actions
- add small runner actions only when repeated workflows prove the existing
  actions are brittle
- keep markdown as canonical for narrative knowledge and source-linked
  synthesis

### Memory layer

- Mem0 remains a likely recall layer later
- memory is promoted recall, not canonical truth

### Records layer

- records remain selective and domain-shaped
- service registry is the first promoted-domain prototype

### Truth sync / provenance layer

- keep it thin and auditable
- use provenance events and projection states as the first inspection surface

### Router

- add it only after docs, synthesis, records, memory, and freshness semantics are
  clear
- begin with deterministic policy plus audit log

## 15. Best Current Recommendation

The best current architecture for OpenClerk is:

1. use AgentOps JSON runner as the production agent substrate
2. keep markdown for research, project notes, runbook-style context, and
   source-linked synthesis
3. deepen `synthesis/` as durable agent-maintained wiki space
4. keep provenance and projection freshness visible through retrieval actions
5. add Mem0 only for promoted recall after freshness rules are solid
6. move only selected domains into records
7. add routing only after the underlying stores and truth-sync semantics are
   reliable

This is intentionally conservative.

It preserves the strengths of the current vault while making the agent
interface runner-first and less file-centric.

## 16. Beads Status

The AgentOps-only direction cleanup, synthesis lifecycle POC, knowledge
configuration v1 lane, and decision-record hardening work are now represented
by closed Beads, including `oc-za6` and `oc-j0a`.

The historical follow-up themes were:

- run a real-vault AgentOps validation trial against representative
  operator-vault workflows using only the installed `openclerk document` and
  `openclerk retrieval` runners
- define prompt/eval and promotion/defer/kill gates for deferred capabilities
  before any implementation issue is filed

Broader memory and routing implementation work should still wait until the
real-vault trial or targeted evals prove the existing document and retrieval
actions are insufficient.

Reason:

- the external pattern is directly related to the OpenClerk vision
- the first changes are documentation, eval, and skill-guidance work with
  concrete acceptance criteria
- the prototype can use existing AgentOps document and retrieval actions before
  any public runner expansion
- the provenance/truth-maintenance scope remains explicit from the start

The historical tracking path included:

- a direction/docs bead for AgentOps-only alignment
- a prototype bead for source-linked synthesis lifecycle
- eval coverage for update-over-duplicate, stale claim repair, answer filing,
  source refs, and freshness checks

Broader implementation Beads for Mem0, autonomous routing, semantic graph
truth, broad contradiction detection, and new runner actions should still wait
until:

1. synthesis lifecycle is reliable through existing AgentOps actions
2. provenance and projection freshness are consistently visible
3. the next capability has concrete acceptance criteria
4. the existing runner actions prove insufficient for that capability

## 17. Highest-Signal Conclusion

The architecture problem is not "replace markdown with memory."

The real problem is:

- separate canonical docs
- canonical records
- and promoted memory
- then keep those representations coherent with explicit provenance, invalidation, and temporal rules
- then expose them through a measured, auditable AgentOps runner

The current build path is source-linked synthesis lifecycle: make durable
agent-maintained wiki pages trustworthy before adding broader memory or routing.

This architecture is only valid if it prevents cross-store drift better than ad
hoc multi-store accumulation.

## Sources

- Karpathy, "LLM Wiki": https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f
- Mitchell Hashimoto, "The Building Block Economy": https://mitchellh.com/writing/building-block-economy
- OpenAI Prompt Guidance: https://developers.openai.com/api/docs/guides/prompt-guidance
- OpenAI Harness Engineering: https://openai.com/index/harness-engineering/
- OpenAI Embeddings: https://developers.openai.com/api/docs/guides/embeddings
- OpenAI Retrieval: https://developers.openai.com/api/docs/guides/retrieval
- Mem0 OSS overview: https://docs.mem0.ai/open-source/overview
- Cognee repository: https://github.com/topoteretes/cognee
- Cognee docs: https://docs.cognee.ai/
- Google Research blog: https://research.google/blog/turboquant-redefining-ai-efficiency-with-extreme-compression/
- TurboQuant paper: https://arxiv.org/abs/2504.19874
- PolarQuant paper: https://arxiv.org/abs/2502.02617
