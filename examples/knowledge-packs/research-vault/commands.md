# Commands

```bash
demo_dir="$(mktemp -d)"
openclerk init --db "$demo_dir/openclerk.sqlite" \
  --vault-root examples/knowledge-packs/research-vault/vault
printf '%s\n' '{"action":"search","search":{"text":"citation freshness retrieval","limit":5}}' \
  | openclerk retrieval --db "$demo_dir/openclerk.sqlite"
printf '%s\n' '{"action":"projection_states","projection":{"projection":"synthesis","limit":10}}' \
  | openclerk retrieval --db "$demo_dir/openclerk.sqlite"
```

Expected high-level outcome: search results include citations to `sources/`
documents, and synthesis projection state names whether source refs are fresh.
