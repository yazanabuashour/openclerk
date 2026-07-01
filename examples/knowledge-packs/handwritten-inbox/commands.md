# Commands

```bash
demo_dir="$(mktemp -d)"
openclerk init --db "$demo_dir/openclerk.sqlite" \
  --vault-root examples/knowledge-packs/handwritten-inbox/vault
openclerk clerk inbox_scan --db "$demo_dir/openclerk.sqlite" \
  --inbox-path examples/knowledge-packs/handwritten-inbox/inbox/shorthand-note.md \
  --limit 5
```

Expected high-level outcome: the runner treats the shorthand as candidate
evidence and returns a review path rather than creating canonical knowledge.
