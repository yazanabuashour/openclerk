# Commands

```bash
demo_dir="$(mktemp -d)"
openclerk init --db "$demo_dir/openclerk.sqlite" \
  --vault-root examples/knowledge-packs/codebase-decisions/vault
openclerk inspect --db "$demo_dir/openclerk.sqlite"
openclerk clerk context_pack --db "$demo_dir/openclerk.sqlite" \
  --task "change the auth callback behavior" \
  --limit 5
```

Expected high-level outcome: cited task context points to
`sources/auth-callback.md`, `sources/session-handling.md`, and the decision in
`decisions/adr-auth-callback.md`. No document is created or updated.
