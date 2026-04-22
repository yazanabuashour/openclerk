# Canonical Docs Link Navigation POC

`oc-85c` implements the canonical docs and link navigation baseline as the
`canonical-docs-navigation-baseline` AgentOps eval scenario. The goal is to
prove what a markdown wiki can do before source-linked synthesis, promoted
records, or memory-style layers are needed.

## Fixture

The scenario seeds a small linked wiki:

- `notes/wiki/agentops/index.md`
- `notes/wiki/agentops/runner-policy.md`
- `notes/wiki/architecture/knowledge-plane.md`
- `notes/wiki/ops/runner-playbook.md`

The `notes/wiki/agentops/` directory is intentionally narrower than the full
relationship set. It contains the local index and runner policy, while the
architecture and operations notes live outside that prefix and connect through
markdown links.

## Runner Actions

The scenario uses only existing OpenClerk runner actions:

- `openclerk document` with `list_documents` and `path_prefix:
  "notes/wiki/agentops/"` to inspect directory-shaped scope.
- `openclerk document` with `get_document` for
  `notes/wiki/agentops/index.md` to inspect stable headings and body content.
- `openclerk retrieval` with `document_links` for the index `doc_id` to inspect
  outgoing links and incoming backlinks.
- `openclerk retrieval` with `graph_neighborhood` for the index `doc_id` to
  inspect derived relationship context with citations.
- `openclerk retrieval` with `projection_states` for `projection: "graph"` to
  confirm the graph projection is fresh and tied to the index document.

No new public runner action is part of this POC.

## Baseline Findings

Directory and heading navigation is sufficient when the task is local and
path-shaped: listing `notes/wiki/agentops/` identifies the local index and
runner policy, and `get_document` exposes headings that make the index
inspectable and repairable.

Plain folder and markdown-link navigation fails when the task depends on
relationship context outside the directory. A path-prefix list does not reveal
that architecture and operations notes link back to the index. Reading the
index also does not by itself distinguish outgoing links from incoming
backlinks or explain the broader graph neighborhood.

AgentOps-backed retrieval adds the missing relationship layer while preserving
canonical markdown authority. `document_links` exposes outgoing and incoming
relationships with citation paths, `graph_neighborhood` shows nearby document
and chunk nodes with cited edges, and graph `projection_states` makes freshness
inspectable. The derived graph remains refreshable from canonical docs rather
than becoming an independent truth source.

## Pass Criteria

The scenario passes only when the agent:

- scopes directory navigation through `list_documents` rather than direct vault
  inspection.
- retrieves the index document and uses returned headings.
- inspects both outgoing links and incoming backlinks through `document_links`.
- inspects cited graph nodes and edges through `graph_neighborhood`.
- checks graph projection freshness through `projection_states`.
- explains where directory/link navigation is sufficient, where it fails, and
  what AgentOps-backed graph behavior adds.
