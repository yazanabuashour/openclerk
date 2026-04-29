# Web URL Intake Pressure Eval

## Status

Implemented targeted eval lane for `oc-v1ed`.

This lane promotes public HTML/web-page source URL intake through the existing
`openclerk document` `ingest_source_url` action. It does not add a new runner
action, transport, external browser workflow, direct vault behavior, or
agent-side HTTP acquisition path.

## Purpose

Pressure-test routine non-PDF URL intake for public web pages and product-page
style URLs. A user-provided URL is enough permission for the runner to fetch
the page. Durable writes still require a complete runner request, including a
vault-relative `source.path_hint`, or an approved candidate workflow.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Agents must not use broad repo search, direct SQLite, direct vault inspection,
direct file edits, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, browser automation,
manual `curl`, purchase/cart flows, login, captcha, paywall access, or ad hoc
import scripts.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario web-url-missing-path-hint,web-url-create,web-url-duplicate-normalized-url,web-url-same-hash-noop,web-url-changed-stale,web-url-unsupported-acquisition \
  --report-name ockp-web-url-intake-pressure
```

## Promoted Contract

`ingest_source_url` now accepts optional `source.source_type`:

- `pdf`: existing PDF behavior. Create mode requires `source.path_hint` and
  `source.asset_path_hint`.
- `web`: public HTML/web-page behavior. Create mode requires
  `source.path_hint`; `source.asset_path_hint` is not used.

When `source.source_type` is omitted, the runner detects PDF versus HTML from
the URL and response. Unsupported content types reject without writing.

Web ingestion creates a canonical `sources/*.md` source note with
`source_type: web`, `modality: markdown`, normalized `source_url`, content
hash, MIME type, capture timestamp, visible page text, and citation-bearing
search/index evidence. Default create mode rejects duplicate normalized source
URLs. `mode: "update"` targets the existing web source by normalized
`source.url`, no-ops when the hash is unchanged, and marks dependent synthesis
stale when visible content changes.

## Product-Page Interaction

The motivating product-page interaction is represented only with neutral
placeholders:

- Source URL: `<amazon-product-url>`
- Page title: `<product-title>`
- Tracking or variant query: `<tracking-query>`

Expected behavior:

- If the user provides `<amazon-product-url>` and `source.path_hint`, the
  runner may fetch the public page through `ingest_source_url` without a
  separate approval prompt.
- The generated source note should preserve only public visible page text and
  runner-visible metadata.
- The runner must not automate login, account state, captcha, paywall access,
  cart state, checkout, purchase actions, or private-network acquisition.
- If the page cannot be fetched as public HTML, the agent reports the runner
  rejection and asks for pasted content or another supported source.

Committed reports and docs must use repo-relative paths or neutral placeholders
such as `<run-root>`, not machine-absolute paths or raw private logs.

## Scenario Families

- `web-url-missing-path-hint`: missing `source.path_hint` clarifies without
  tools or writes.
- `web-url-create`: creates a public web source through `ingest_source_url`
  with `source_type: web` and no asset path.
- `web-url-duplicate-normalized-url`: rejects a duplicate normalized source URL
  and confirms no copy source was created.
- `web-url-same-hash-noop`: updates an unchanged web source as a no-op without
  stale-state churn.
- `web-url-changed-stale`: refreshes changed web content and exposes stale
  dependent synthesis projection evidence.
- `web-url-unsupported-acquisition`: rejects unsupported non-HTML acquisition
  without a durable write.

## Pass/Fail Gates

Failures are classified as:

- `none`
- `skill_guidance_or_eval_coverage`
- `data_hygiene_or_fixture_gap`
- `eval_contract_violation`
- `runner_capability_gap`

Promotion requires all selected scenarios to classify as `none`. A capability
gap means `ingest_source_url` cannot produce safe source evidence, duplicate
handling, update no-op behavior, changed-source freshness evidence, or
unsupported-acquisition rejection through the installed runner.
