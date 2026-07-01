# Commands

```bash
demo_dir="$(mktemp -d)"
openclerk init --db "$demo_dir/openclerk.sqlite" \
  --vault-root examples/knowledge-packs/agent-session-to-docs/vault
openclerk clerk run --once --db "$demo_dir/openclerk.sqlite" \
  --inbox-path examples/knowledge-packs/agent-session-to-docs/handoffs/session.md \
  --task "summarize completed auth callback work into repo knowledge" \
  --limit 5
```

Expected high-level outcome: context is cited before planning, and the
after-work report remains read-only with approval-required candidate writes.
