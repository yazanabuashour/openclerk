# Path, Title, And Autofiling UX Audit

## Status

Implemented for `oc-4rxs`.

This audit is documentation and eval-design work only. It does not change
runner behavior, JSON schemas, storage, public APIs, `skills/openclerk/SKILL.md`,
or the eval harness. The current shipped ceiling remains the accepted
propose-before-create skill policy recorded in
`docs/architecture/agent-chosen-document-artifact-candidate-generation-adr.md`.

## Evidence Map

| Evidence | What it proves | UX read |
| --- | --- | --- |
| `docs/architecture/agent-chosen-vault-path-selection-adr.md` and `docs/evals/results/ockp-path-title-autonomy-pressure.md` | Path/title autonomy did not expose a runner capability gap across selected URL-only, artifact-hint, duplicate, override, and metadata-authority pressure. | Capability passed, but the explicit-field workflow can still feel heavier than normal capture intent. |
| `docs/architecture/agent-side-knowledge-intake-autofiling-adr.md` and `docs/evals/results/ockp-document-this-intake-pressure.md` | Current document/retrieval primitives handled document-this intake without promoting runner autofiling or direct create. | Successful rows can still be ceremonial when a user asks for capture rather than schema-shaped JSON. |
| `docs/architecture/agent-chosen-document-artifact-candidate-generation-adr.md` and `docs/evals/results/ockp-document-artifact-candidate-generation.md` | Candidate path/title/body generation before write is safe when it is derived from explicit user content and validated before presentation. | This is the current best smoother surface for missing document fields. |
| `docs/evals/results/ockp-document-artifact-candidate-ergonomics.md` | Natural-intent and scripted candidate rows passed after `oc-9k3`, including duplicate-risk and low-confidence behavior. | Propose-before-create is acceptable for covered rows, but future low-risk capture may still need ceremony pressure. |
| `docs/architecture/knowledge-configuration-v1-adr.md` and `docs/evals/results/ockp-web-url-intake-pressure.md` | Public web URL ingestion belongs under `ingest_source_url`; a public URL is enough permission to fetch, while durable writes still need complete fields. | `oc-v1ed` is the taste baseline: distinguish read/fetch permission from durable-write approval. |

## Infer, Propose, Or Ask Boundaries

| Intake field or flow | Current boundary | Classification | Required safety clarification | Likely taste debt to pressure-test |
| --- | --- | --- | --- | --- |
| `document.body` omitted | Do not synthesize body content from intent alone. A candidate body may use only explicit user-supplied content. | Ask, unless explicit content supports a faithful proposal. | Ask when body content is missing or the candidate would require fetching or inventing facts. | Natural "save this" prompts with obvious pasted body should stay proposal-shaped, not missing-field shaped. |
| `document.path` omitted | Preserve explicit user paths. Otherwise choose a path only inside a candidate proposal, commonly under `notes/candidates/<slug-from-title>.md` for note-like content. | Propose. | Ask when path choice depends on unknown artifact type, source set, authority, or low confidence. | Low-risk notes may make approval-before-write feel ceremonial even when validation passes. |
| `document.title` omitted | Preserve explicit titles. Otherwise choose a title only inside a faithful candidate proposal from headings or supplied subject text. | Propose. | Ask when no stable title can be formed without adding meaning. | Title-only friction is likely taste debt when body and intent are clear. |
| Public web URL with `source.path_hint` | `ingest_source_url` may fetch public HTML/web content through the runner without separate pre-fetch approval. | Infer fetch permission; create through runner. | Durable write still requires complete runner fields and supported public acquisition. | None for fetch approval; this is the baseline taste improvement from `oc-v1ed`. |
| Public web URL missing `source.path_hint` | Current shipped behavior asks for `source.path_hint`; `source.asset_path_hint` is not used for web source create mode. | Ask. | Ask because the durable canonical source path is missing. | A future eval should test whether proposing a source path before write is better for natural "document these links" requests. |
| PDF source URL missing hints | PDF create mode still requires `source.path_hint` and `source.asset_path_hint`. | Ask. | Ask because both canonical source and asset locations affect durable authority and repairability. | A source-path proposal may be worth evaluating, but asset placement has higher durable-write risk. |
| Duplicate candidate visible | Use runner-visible `search`, `list_documents`, and `get_document` only when the workflow is otherwise valid; then ask whether to update the visible document or create a new confirmed path. | Ask after lookup. | Ask before writing because duplicate avoidance and update intent are authority-sensitive. | The ask can be smoother if it names the visible target and the alternative path clearly. |
| Explicit path, title, type, or naming instructions | Explicit user values win unless validation fails or runner-visible authority conflicts. | Infer only by preserving explicit input. | Ask or report validation failure when explicit values are invalid or conflict with authority. | Any future smoother autofiling flow must prove it never silently overrides explicit naming. |
| Ambiguous source vs note vs synthesis intent | Do not treat path conventions as document type authority. Metadata, source refs, provenance, and projection freshness remain authoritative. | Ask or propose, depending on supplied content and confidence. | Ask when durable artifact kind or authority model is unclear. | Multi-link capture may need a better proposal shape that distinguishes source creation from synthesis creation. |

## Audit Findings

The prior capability decisions remain sound. `oc-iat` and `oc-99z` did not
authorize runner autofiling, autonomous path/title selection, direct create,
schema changes, storage changes, or public API changes. Current primitives can
technically express the selected workflows while preserving canonical markdown
authority, duplicate handling, runner-only access, provenance, freshness, and
approval before durable writes.

The taste review changes how future defer/reference outcomes should be read.
When a normal user says "document these links" or "save this note", they are
often asking for capture, not for an exact vault path exercise. A completed eval
row can still be UX debt if it depends on exact prompt choreography, long step
count, or a clarification turn for a naming decision OpenClerk could safely
propose.

The strongest current boundary is:

- infer public fetch permission from a user-provided public URL
- infer path/title only inside a non-durable proposal from explicit supplied
  content
- ask before durable writes when fields are incomplete, duplicate/update intent
  is ambiguous, authority is unclear, or confidence is low

This keeps `oc-v1ed`'s fetch/write distinction without weakening existing
invariants. A public URL can authorize the runner to read or inspect public
content. It does not authorize durable knowledge writes without complete runner
fields or an approved candidate workflow.

## Eval-Design Follow-Ups

The following follow-up Beads were filed from this audit:

| Bead | Flow | Eval-design question |
| --- | --- | --- |
| `oc-zf3o` | Document these links | Should OpenClerk propose source or synthesis paths for public links before write instead of asking for `source.path_hint` immediately? |
| `oc-qjhm` | Save this note | Does the current candidate path/title/body proposal handle natural note capture without unnecessary missing-field friction? |
| `oc-k6eb` | Duplicate candidate update vs new path | Can duplicate-risk wording be smoother while preserving runner-visible lookup and no duplicate writes? |
| `oc-mjpz` | Explicit overrides in smoother autofiling | Do future smoother flows preserve explicit path/title/type/body instructions and fail validation instead of silently rewriting them? |
| `oc-18oo` | Low-risk capture ceremony | Is approval-before-create still the right UX for low-risk capture, or is it ceremonial despite passing safety checks? |

These follow-ups are eval/design scoped. They do not authorize implementation.
Any future behavior change still needs targeted eval evidence and an explicit
promotion decision naming the public surface, safety gates, and compatibility
expectations.
