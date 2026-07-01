# Commands

```bash
demo_dir="$(mktemp -d)"
openclerk init --db "$demo_dir/openclerk.sqlite" \
  --vault-root examples/knowledge-packs/incident-handoff/vault
openclerk clerk run --once --db "$demo_dir/openclerk.sqlite" \
  --inbox-path examples/knowledge-packs/incident-handoff/handoffs/session.md \
  --task "turn incident handoff into durable repo knowledge" \
  --limit 5
```

Expected high-level outcome: the report stays planned-no-write, includes a
candidate note from the handoff, and surfaces context from `incidents/`.
