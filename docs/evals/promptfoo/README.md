# Optional Promptfoo Smoke Suite

This directory contains a tiny Promptfoo smoke suite for interoperability and
review ergonomics. It is optional and non-authoritative: OCKP remains the
production and release authority.

Run from the repository root:

```bash
promptfoo validate -c docs/evals/promptfoo/ockp-smoke.promptfooconfig.yaml
promptfoo eval -c docs/evals/promptfoo/ockp-smoke.promptfooconfig.yaml
```

If the Promptfoo CLI is not installed locally, the same smoke suite can be run
ephemerally:

```bash
npx --yes promptfoo@latest validate -c docs/evals/promptfoo/ockp-smoke.promptfooconfig.yaml
npx --yes promptfoo@latest eval -c docs/evals/promptfoo/ockp-smoke.promptfooconfig.yaml --no-cache --no-write --no-table
```

The suite uses Promptfoo's `exec:` provider to call
[`ockp_smoke_provider.sh`](ockp_smoke_provider.sh), which targets the installed
`openclerk` runner by default. Override the runner command only for local
debugging:

```bash
OCKP_PROMPTFOO_OPENCLERK_CMD=openclerk promptfoo eval -c docs/evals/promptfoo/ockp-smoke.promptfooconfig.yaml
```

## Scope

- Fast smoke checks for installed runner capabilities and validation behavior.
- Deterministic assertions only; no LLM rubric or model-only prompt is used.
- Synthetic paths and public example URLs only; no secrets, customer data, or
  sensitive personal data.
- No durable writes; all runner calls are `capabilities` or `validate`.

Passing this suite does not bless a release, and failing this suite does not
block a release by itself. Treat failures as review prompts, then confirm any
real behavior question through OCKP.
