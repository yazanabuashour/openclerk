# Baseline Scenarios

These scenarios define regression coverage for the installed OpenClerk AgentOps
runner/skill surface and the knowledge-model behavior behind it. They also
define proof obligations for the AgentOps-only knowledge-plane direction in
`docs/architecture/eval-backed-knowledge-plane-adr.md`.

The benchmark archetypes and comparison axes behind these scenarios are defined
in `docs/evals/knowledge-plane-archetype-matrix.md`.

Deferred capability promotion gates are defined in
`docs/architecture/deferred-capability-promotion-gates.md`. Mem0 or memory
APIs, autonomous routing, semantic graph truth, broad contradiction engines,
and new public runner actions require targeted AgentOps eval evidence and an
explicit promote/defer/kill/reference decision before any implementation issue
is filed. Promotion evidence can be a capability gap or an ergonomics gap; an
expressible workflow may still justify promotion if targeted evidence shows it
is too slow, too many steps, too scripted, too brittle, or too dependent on
skill guidance for routine use. Defer/reference decisions must still run the
taste review: distinguish read/fetch/inspect permission from durable-write
approval, prefer natural extensions of existing runner actions when the input
belongs there, and record completed-but-ceremonial passes as possible taste
debt.

## Ergonomics Scorecard

Targeted deferred-capability lanes should report an ergonomics scorecard
alongside correctness:

- tool or command count
- assistant calls
- wall time
- non-cached input tokens when available
- prompt specificity required to make the workflow pass
- whether a natural user-intent prompt passes without scripted runner steps
- retry or brittleness signs, such as duplicate creation, wrong target
  selection, skipped freshness inspection, dropped citations, or missing source
  refs
- authority, provenance, freshness, privacy, and bypass risks for any proposed
  surface
- safety pass, capability pass, and UX quality as separate report conclusions

Each lane should include a natural-user-intent scenario and a scripted-control
scenario. The natural scenario measures UX; the scripted control verifies what
the current primitives can do when the agent is given exact instructions.
Passing both scenarios does not erase taste debt. If the workflow succeeds only
through ceremony, high latency, high step count, exact prompt choreography, or
surprising clarification turns, record UX quality separately from the safety
and capability passes.

## Source-Grounded Retrieval

- Create canonical notes with stable headings and exact terms.
- Verify search returns the correct `doc_id`, `chunk_id`, and citations.
- Verify path-prefix and metadata filters reduce scope correctly.
- Verify the `rag-retrieval-baseline` scenario can answer accurately from
  retrieval-only search while preserving source path, `doc_id`, `chunk_id`, and
  line citations.
- Verify repeated retrieval-only answers rerun search and do not create
  `synthesis/` durable synthesis unless explicitly filed.

## Source Ingest And Synthesis

- Create a source-shaped canonical doc and a source-linked synthesis page that
  cites it.
- Add a second source that updates or challenges the synthesis.
- Verify the synthesis is updated rather than duplicated.
- Verify synthesis pages live under `synthesis/`, include `type:
  synthesis`, `status: active`, `freshness: fresh`, `source_refs`, a
  `Sources` section, and a `Freshness` section.
- Verify source-sensitive claims preserve citation paths, chunk ids, or explicit
  source refs.
- Verify useful answer material can be filed back into durable markdown instead
  of remaining only in chat history.
- Verify LLM Wiki-style synthesis stays subordinate to canonical sources and
  does not create a second authority layer.

## Contradiction And Stale Synthesis

- Create an initial source and synthesis page with a cited claim.
- Add a later source that supersedes or contradicts the claim.
- Verify the agent finds the existing synthesis page before writing a new one.
- Verify the synthesis is updated with the newer evidence, contradiction note,
  or explicit stale-state language.
- Verify the agent retrieves the existing synthesis document before replacing
  or appending sections.
- Verify the final answer identifies which source is current when the prompt is
  source-sensitive.
- Verify source-sensitive audit repair distinguishes current sources,
  superseded sources, stale synthesis, and decoy synthesis candidates before
  updating durable markdown.
- Verify unresolved conflicting current sources are explained with both source
  paths and no winner is chosen when runner-visible source authority or
  supersession metadata is absent.
- Treat arbitrary semantic contradiction detection as a non-goal unless a later
  promoted primitive is justified by eval failures.

## Synthesis Compiler Pressure

- Seed current and superseded sources, an existing stale synthesis page, and a
  decoy synthesis candidate; verify the agent selects and repairs the target
  synthesis without creating a duplicate.
- Seed a multi-source evidence set; verify a new synthesis preserves every
  source ref as single-line comma-separated frontmatter and includes
  `## Sources` and `## Freshness`.
- Run a resumed multi-turn drift repair where the first turn creates synthesis
  and the second turn updates newer source evidence before repairing the
  synthesis.
- Verify pressure scenarios still use only `search`, `list_documents`,
  `get_document`, `replace_section` or `append_document`, and
  `projection_states` where freshness is relevant.
- Promote a dedicated synthesis/compiler action only if repeated failures show
  the existing runner workflow is structurally insufficient or ergonomically
  unacceptable under the scorecard above.

## Docs Navigation

- Run the `canonical-docs-navigation-baseline` scenario from `oc-85c`.
- Create a linked wiki fixture under `notes/wiki/` with an AgentOps index,
  runner policy, architecture note, and operations playbook.
- Verify `list_documents` with `path_prefix: "notes/wiki/agentops/"` returns
  the directory-shaped index and policy docs without depending on direct vault
  inspection.
- Verify `get_document` returns stable headings for the index document.
- Verify `document_links` returns outgoing markdown links and incoming
  backlinks with citations.
- Verify `graph_neighborhood` returns source-linked graph nodes and edges.
- Verify graph `projection_states` exposes fresh derived graph state.
- Require the final answer to document where directory/link navigation is
  sufficient, where it fails, and what AgentOps-backed graph behavior adds.
- Run the `graph-semantics-reference-poc` scenario from `oc-za6.6`.
- Compare semantic relationship words in canonical markdown against search,
  `document_links`, backlinks, `graph_neighborhood`, and graph freshness.
- Verify richer relationship labels remain canonical markdown evidence rather
  than promoted graph edge authority.

## Graph/Memory Reference Archetype

These are future benchmark categories inspired by graph/vector memory systems
such as Cognee. They should remain reference comparisons unless the current
AgentOps eval harness supports them cleanly.

- Compare derived graph navigation against docs search and link expansion.
- Verify graph answers preserve source citations and projection freshness.
- Keep semantic graph labeling as a reference/deferred pattern unless eval
  evidence shows it adds value over markdown text, links, backlinks, graph
  neighborhoods, and search.
- Run the `memory-router-reference-poc` scenario from `oc-za6.7` as a
  reference-only benchmark for temporal recall, session promotion, feedback
  weighting, and routing.
- Verify temporal retrieval distinguishes current, observed, effective, and
  superseded facts.
- Verify session-derived or feedback-weighted recall can explain source refs
  and freshness before being trusted.
- Verify memory/router answers remain subordinate to canonical docs and
  promoted records, expose provenance/freshness, and do not promote
  `remember`/`recall`, autonomous routing, or a new production interface.

## Promoted-Domain Lookup

- Create a canonical record-shaped doc with `entity_*` frontmatter and `Facts`.
- Verify promoted lookup returns the expected entity and citations.
- Create a canonical service-shaped doc and verify typed `services_lookup`
  returns service id, owner, status, interface, facts, and citations.
- Update the canonical source and verify the derived projection refreshes.
- Compare the service registry path against plain docs retrieval for the same
  service-centric task.
- Compare decision records against plain docs retrieval for decision-centric
  tasks requiring status, scope, owner, repeatable lookup, citations, and
  supersession freshness.
- Accept the promoted-domain path only when it improves precision, update
  safety, or structured lookup behavior without weakening citation correctness.

## Provenance And Freshness

- Verify document create/update events are emitted.
- Verify source create/update events are emitted for canonical source docs.
- Verify projection invalidation and projection refresh events are visible.
- Verify projection-state reads expose current freshness and version markers.
- Verify synthesis projection-state reads explain whether a synthesis page is
  fresh or stale and report current, superseded, missing, and stale source refs.
- Verify synthesis pages can be traced back to the canonical docs or records
  they summarize and can be repaired after source updates.
- Verify promoted-record synthesis inspects `records_lookup`,
  `provenance_events`, and `projection_states` before writing durable
  synthesis.
- Verify decision supersession projection states expose stale superseded
  decisions, fresh replacements, provenance, and citation paths.
- Verify source-sensitive audit workflows inspect `provenance_events` and
  `projection_states` before repairing stale synthesis and preserve current,
  superseded, and unresolved conflict source refs in final answers.

## Knowledge Layout Inspection

- Verify `inspect_layout` explains the effective convention-first layout
  through runner JSON only.
- Verify layout JSON reports `config_artifact_required: false`, conventional
  prefixes such as `sources/` and `synthesis/`, first-class
  document kinds, and pass/warn/fail checks.
- Verify invalid or incomplete synthesis, missing `source_refs`, missing
  `## Sources` or `## Freshness`, missing source paths, and partial
  record/service identity metadata are visible as failed layout checks.
- Verify agents do not inspect the vault, SQLite, source files, or alternate
  transports to explain routine layout validity.

## AgentOps Contract Enforcement

- Verify production tasks use `openclerk` JSON runner requests rather than
  direct SQLite, backend variants, stale API paths, ad hoc runtime programs, or
  source-built command paths.
- Verify requests missing required document or retrieval fields produce one
  no-tools clarification response that names the missing fields and asks the
  user to provide them.
- Verify routine attempts to bypass the OpenClerk runner through lower-level or
  alternate transports are rejected final-answer-only without tools.
- Verify the selected knowledge-model workflows stay expressible through
  documented document and retrieval runner actions.
