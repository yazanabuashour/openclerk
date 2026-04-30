# Post-`oc-v1ed` URL And Artifact Intake Audit

## Status

Audit for `oc-fbqy`. This note records evidence and future eval-design
backlog only. It does not authorize runner actions, storage changes, schema
changes, public APIs, parser pipelines, or skill behavior changes.

## Baseline

`oc-v1ed` is the current baseline correction for URL intake. Public
HTML/web-page source URLs belong under the existing `openclerk document`
`ingest_source_url` surface, not under an adjacent unsupported class.

The accepted boundary is:

- a user-provided public URL is enough permission for runner-owned read, fetch,
  and inspect work
- durable writes require complete runner fields or an approved candidate
  workflow
- duplicate handling, provenance, citations, freshness, and canonical markdown
  authority remain required
- private URLs, auth, account state, captcha, paywall access, purchase or cart
  actions, browser automation, direct vault writes, manual downloads, direct
  SQLite, HTTP/MCP bypasses, and source-built runners remain unsupported

Baseline evidence:

- [`knowledge-configuration-v1-adr.md`](knowledge-configuration-v1-adr.md)
- [`../evals/web-url-intake-pressure.md`](../evals/web-url-intake-pressure.md)
- [`../evals/results/ockp-web-url-intake-pressure.md`](../evals/results/ockp-web-url-intake-pressure.md)
- [`deferred-capability-promotion-gates.md`](deferred-capability-promotion-gates.md)
- [`openclerk-taste-review-backlog.md`](openclerk-taste-review-backlog.md)

## Audit Findings

| Area | Current safety pass | Current capability pass | UX quality / taste pass |
| --- | --- | --- | --- |
| Public HTML/web-page URLs | Pass. `ingest_source_url` preserves runner ownership, public-only fetch policy, duplicate rejection, provenance, citations, and freshness. | Pass. `oc-v1ed` promoted public web ingestion through the existing URL action. | Pass as the preferred baseline. Do not reopen the old unsupported-public-URL boundary. |
| PDF source URLs and update mode | Pass. Existing PDF create and update behavior preserves asset paths, source URL identity, duplicate rejection, conflict handling, provenance events, search refresh, and stale synthesis visibility. | Pass. Existing `ingest_source_url` create/update covers the PDF source workflow. | Mixed. Scripted control is acceptable, but natural PDF intent remains a taste flag when it needs many runner calls or exact prompt choreography. |
| Document-this and candidate intake | Pass. Strict runner writes and propose-before-create policy keep approval at durable writes and avoid invented body content. | Pass for explicit content, approved candidates, duplicate checks, and existing-document updates. | Mixed. Missing body, path, title, or source hints should still clarify, but future evals can test whether common "save this note" and "document these links" flows are too ceremonial. |
| Generalized artifact ingestion | Pass. The deferred decision avoids second truth surfaces, parser authority, hidden provenance, duplicate truth, and bypass paths. | Pass for the evaluated rows: current document/retrieval workflows express the targeted PDF, transcript, invoice, receipt, and mixed-artifact cases without runner capability gaps. | Taste debt remains for high-step natural artifact rows, especially where a normal user may expect smoother local file, multi-artifact, or unsupported-kind intake. |
| Video and YouTube source intake | Pass. `ingest_video_url` is limited to supplied transcripts, and native media acquisition stays outside routine AgentOps. | Pass for supplied transcript text; no pass for native acquisition because acquisition is intentionally deferred. | Defer. A user may expect a video URL to work like a public source URL, but media download, transcript acquisition, privacy policy, dependency choice, and citation mapping need separate design/eval evidence. |
| Product-page public web intake | Pass for public visible HTML. The runner must not automate login, account state, captcha, paywall access, cart state, checkout, purchase actions, or private-network access. | Pass for public HTML product-page style URLs under `source_type: web`. | Mixed. Rich product pages can still expose edge cases around dynamic content, blocked public HTML, tracking parameters, variant text, and visible text fidelity. These are eval-design gaps, not implementation approval. |

## Remaining Gaps

The remaining gaps are eval/design backlog, not implementation approval:

- **Local file artifact intake:** clarify whether local files should stay as
  user-supplied content/candidate documents, require explicit asset policy, or
  need future source ingestion evidence.
- **Native media acquisition:** evaluate video/audio transcript acquisition
  options, privacy policy, dependency policy, transcript provenance, citation
  mapping, and update/freshness behavior before any promoted surface.
- **Richer public product-page cases:** test public HTML pages with tracking
  parameters, blocked or non-HTML responses, dynamic content omissions, and
  visible-text fidelity without browser automation or purchase flows.
- **Unsupported artifact kinds:** design eval pressure for images, slide decks,
  emails, exported chats, forms, and other artifacts to decide whether existing
  candidate/document workflows are acceptable or overly ceremonial.

These gaps should produce future eval-design beads only. Implementation beads
remain inappropriate until targeted evidence and a promotion decision name an
exact surface, request/response shape, compatibility policy, failure modes, and
gates.

## Decision

Keep `oc-v1ed` as the accepted intake correction. Preserve public URL fetch and
inspect through `ingest_source_url`, keep durable-write approval at complete
runner requests or approved candidate workflows, and keep unsupported
acquisition boundaries intact.

Close the `oc-fbqy` audit after filing future eval-design beads for the
remaining gaps. Do not file implementation work from this audit.
