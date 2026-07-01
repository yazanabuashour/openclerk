# Commands

```bash
demo_dir="$(mktemp -d)"
cp -R examples/knowledge-packs/stale-runbook/vault "$demo_dir/vault"
openclerk init --db "$demo_dir/openclerk.sqlite" \
  --vault-root "$demo_dir/vault"
openclerk inspect --db "$demo_dir/openclerk.sqlite"
sleep 1
touch "$demo_dir/vault/sources/deploy-source.md"
printf '%s\n' '{"action":"maintenance_report","maintenance":{"query":"deploy window","path_prefix":"synthesis/","limit":10}}' \
  | openclerk retrieval --db "$demo_dir/openclerk.sqlite"
```

Expected high-level outcome: the copied source is newer than the copied
synthesis page, so the report stays read-only and points the agent to review
stale synthesis before submitting an approved `compile_synthesis` request.
