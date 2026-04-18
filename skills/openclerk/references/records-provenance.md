# Records, Graph, And Provenance Recipes

OpenClerk keeps canonical Markdown as the source of truth. Links, graph
neighborhoods, promoted records, provenance events, and projection states are
derived views over those documents.

## Links And Graph

```bash
printf '%s\n' '{"action":"document_links","doc_id":"doc_id_from_json"}' |
  go run ./cmd/openclerk-agentops retrieval

printf '%s\n' '{"action":"graph_neighborhood","doc_id":"doc_id_from_json","limit":20}' |
  go run ./cmd/openclerk-agentops retrieval
```

Use links and graph results to navigate canonical documents. Do not treat graph
state as a separate source of truth.

## Promoted Records

```bash
printf '%s\n' '{"action":"records_lookup","records":{"text":"solenoid","limit":10}}' |
  go run ./cmd/openclerk-agentops retrieval

printf '%s\n' '{"action":"record_entity","entity_id":"entity_id_from_json"}' |
  go run ./cmd/openclerk-agentops retrieval
```

Use records lookup only for promoted-domain questions. The current projection is
derived from record-shaped Markdown with `entity_*` frontmatter and a `Facts`
section.

## Provenance And Freshness

```bash
printf '%s\n' '{"action":"provenance_events","provenance":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}' |
  go run ./cmd/openclerk-agentops retrieval

printf '%s\n' '{"action":"projection_states","projection":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}' |
  go run ./cmd/openclerk-agentops retrieval
```

Use provenance and projection freshness when maintaining source-linked
synthesis. A synthesis page should not hide whether it came from canonical docs,
promoted records, or a derived projection that may need refresh.
