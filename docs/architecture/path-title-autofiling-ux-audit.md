# Path, Title, And Autofiling UX Audit

## Status

Implemented for `oc-4rxs`.

This audit started as documentation and eval-design work for `oc-4rxs`.
`oc-wm04` implements the resulting taste direction in `skills/openclerk/SKILL.md`
without changing runner behavior, JSON schemas, storage, public APIs, or the
eval harness. The shipped ceiling is proposal-first planning: agent/OpenClerk
defaults may choose candidate path, title, body preview, tags, and fields, but
durable writes remain approval-gated.

## Evidence Map

| Evidence | What it proves | UX read |
| --- | --- | --- |
| `docs/architecture/agent-chosen-vault-path-selection-adr.md` and `docs/evals/results/ockp-path-title-autonomy-pressure.md` | Path/title autonomy did not expose a runner capability gap across selected URL-only, artifact-hint, duplicate, override, and metadata-authority pressure. | Capability passed, but the explicit-field workflow can still feel heavier than normal capture intent. |
| `docs/architecture/agent-side-knowledge-intake-autofiling-adr.md` and `docs/evals/results/ockp-document-this-intake-pressure.md` | Current document/retrieval primitives handled document-this intake without promoting runner autofiling or direct create. | Successful rows can still be ceremonial when a user asks for capture rather than schema-shaped JSON. |
| `docs/architecture/agent-chosen-document-artifact-candidate-generation-adr.md` and `docs/evals/results/ockp-document-artifact-candidate-generation.md` | Candidate path/title/body generation before write is safe when it is derived from explicit user content and validated before presentation. | This is the smoother surface for missing document fields. |
| `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` and `docs/evals/artifact-intake-autofiling-tags-fields-poc.md` | `artifact_candidate_plan` can package candidate path, title, body preview, tags, fields, duplicates, confidence, and next approved requests. | This is the default planning surface when omitted tags, fields, or source handoff matter. |
| `docs/evals/results/ockp-document-artifact-candidate-ergonomics.md` | Natural-intent and scripted candidate rows passed after `oc-9k3`, including duplicate-risk and low-confidence behavior. | Propose-before-create is acceptable for covered rows, but future low-risk capture may still need ceremony pressure. |
| `docs/architecture/knowledge-configuration-v1-adr.md` and `docs/evals/results/ockp-web-url-intake-pressure.md` | Public web URL ingestion belongs under `ingest_source_url`; a public URL is enough permission to fetch, while durable writes still need complete fields. | `oc-v1ed` is the taste baseline: distinguish read/fetch permission from durable-write approval. |

## Infer, Propose, Or Ask Boundaries

| Intake field or flow | Current boundary | Classification | Required safety clarification | Likely taste debt to pressure-test |
| --- | --- | --- | --- | --- |
| `document.body` omitted | Do not synthesize body content from intent alone. A candidate body may use only explicit user-supplied content. | Ask, unless explicit content supports a faithful proposal. | Ask when body content is missing or the candidate would require fetching or inventing facts. | Natural "save this" prompts with obvious pasted body should stay proposal-shaped, not missing-field shaped. |
| `document.path` omitted | Preserve explicit user paths. Otherwise choose a path only inside a candidate proposal, commonly under `notes/candidates/<slug-from-title>.md` for note-like content. | Propose. | Ask when path choice depends on unknown artifact type, source set, authority, or low confidence. | Low-risk notes may make approval-before-write feel ceremonial even when validation passes. |
| `document.title` omitted | Preserve explicit titles. Otherwise choose a title only inside a faithful candidate proposal from headings or supplied subject text. | Propose. | Ask when no stable title can be formed without adding meaning. | Title-only friction is likely taste debt when body and intent are clear. |
| Tags or metadata fields omitted | Preserve explicit tags and fields first. Otherwise let `artifact_candidate_plan` infer visible planning tags and fields from explicit content, artifact kind, and source context. | Propose. | Ask when metadata would assert unsupported authority or conflict with runner-visible state. | Normal users expect organization defaults and occasional explicit tag overrides, not a required tagging schema. |
| Public web URL with `source.path_hint` | `ingest_source_url` may fetch public HTML/web content through the runner without separate pre-fetch approval. | Infer fetch permission; create through runner. | Durable write still requires complete runner fields and supported public acquisition. | None for fetch approval; this is the baseline taste improvement from `oc-v1ed`. |
| Public web URL missing `source.path_hint` | Direct create still needs `source.path_hint`, but `ingest_source_url` plan mode or `artifact_candidate_plan` may propose source and synthesis placement before write. | Propose. | Ask only when placement confidence, source type, private access, or authority is unclear. | Natural "document these links" should use planning before demanding path policy from the user. |
| PDF source URL missing hints | PDF create mode still requires `source.path_hint` and `source.asset_path_hint`; planning may propose both before durable fetch/write. | Propose with approval. | Ask when asset placement, source type, or acquisition boundary is unclear. | Asset placement remains higher-risk, so approval wording must stay explicit. |
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
- infer path/title/tags/fields only inside a non-durable proposal from explicit
  supplied content or runner-supported public-source context
- ask before durable writes when fields are incomplete, duplicate/update intent
  is ambiguous, authority is unclear, or confidence is low

This keeps `oc-v1ed`'s fetch/write distinction without weakening existing
invariants. A public URL can authorize the runner to read or inspect public
content. It does not authorize durable knowledge writes without complete runner
fields or an approved candidate workflow.

## Explicit Requirement Inventory

| Area | Explicit user requirement today | Proposal-first default | Still ask when |
| --- | --- | --- | --- |
| Document capture | `document.path`, `document.title`, `document.body` for durable create | Plan candidate path, title, body preview, tags, and fields from explicit content; explicit overrides win | Body/content is missing, confidence is low, or authority conflicts |
| Artifact intake | optional path/title/body/tags/fields plus duplicate scope | Prefer `artifact_candidate_plan` for omitted filing defaults, metadata, confidence, duplicate evidence, and next approved request | Artifact is opaque or unsupported, OCR review is not explicitly requested, or parser truth would be invented |
| Public source URL | `source.url`; direct create also needs source path and PDF asset path hints | Use `ingest_source_url` plan mode or `artifact_candidate_plan` to propose source, asset, and synthesis placement | URL is private/authenticated, unsupported, ambiguous, or needs a durable write without approval |
| Video source | `video.url`, transcript text/provenance, and create/update path policy | Preserve supplied transcript and explicit hints; candidate filing can be proposed only from supplied transcript content | Transcript text/provenance is missing or native acquisition/transcription would be required |
| Synthesis | `synthesis.path`, title, source refs, and body/body facts for durable compile | Agent may propose path/title/body facts from runner-visible source refs before approved compile | Source refs, source authority, body facts, or duplicate target are unclear |
| Existing updates | target `doc_id`, section/append content, or update-vs-new choice | Use runner-visible duplicate/search/list/get evidence to name likely targets and alternatives | Multiple targets, conflicting authority, missing content, or update intent is unclear |
| Git lifecycle | checkpoint paths, message, and explicit checkpoint gate | Status/history remain read-only planning evidence | Checkpoint config, paths, or message are missing |

## Eval-Design Follow-Ups

The following follow-up Beads were filed from this audit. Their consolidated
future eval-design framing is
[`../evals/path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md).

| Bead | Flow | Eval-design question | Design |
| --- | --- | --- | --- |
| `oc-zf3o` | Document these links | Should OpenClerk propose source or synthesis paths for public links before write instead of asking for `source.path_hint` immediately? | [`path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) |
| `oc-qjhm` | Save this note | Does the current candidate path/title/body proposal handle natural note capture without unnecessary missing-field friction? | [`path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) |
| `oc-k6eb` | Duplicate candidate update vs new path | Can duplicate-risk wording be smoother while preserving runner-visible lookup and no duplicate writes? | [`path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) |
| `oc-mjpz` | Explicit overrides in smoother autofiling | Do future smoother flows preserve explicit path/title/type/body instructions and fail validation instead of silently rewriting them? | [`path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) |
| `oc-18oo` | Low-risk capture ceremony | Is approval-before-create still the right UX for low-risk capture, or is it ceremonial despite passing safety checks? | [`path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) |
| `oc-11yz` | Remaining write surfaces | Which proposal-first shapes are viable for video transcript intake, source-linked synthesis, existing-document update targeting, and checkpoint guidance? | Follow-up candidate-surface comparison |

These follow-ups are eval/design scoped. They do not authorize implementation.
Any future behavior change still needs targeted eval evidence and an explicit
promotion decision naming the public surface, safety gates, and compatibility
expectations.
