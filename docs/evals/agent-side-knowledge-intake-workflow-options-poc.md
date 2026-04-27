# Agent-Side Knowledge Intake Workflow Options POC

## Status

Planned targeted POC/eval contract for `oc-tw5`.

This document does not add runner actions, schemas, storage migrations, skill
behavior, public API, or release-blocking production gates. It compares
skill-level workflows for routine "document this" intake using only existing
`openclerk document` and `openclerk retrieval` JSON runner actions.

The governing ADR is
[`../architecture/agent-side-knowledge-intake-autofiling-adr.md`](../architecture/agent-side-knowledge-intake-autofiling-adr.md).

## Purpose

This POC should determine how the OpenClerk skill can handle natural intake
requests while still producing strict runner JSON. The target pressure is user
intent such as "document this", "save this note", "capture these links", or
"update the existing synthesis" where the request may omit path, title, body,
source hints, duplicate target, or freshness context.

The POC compares three agent-side workflow options:

- **Ask for missing fields:** answer once with no tools when required fields are
  missing, name the missing fields, and wait for user input.
- **Propose before create:** derive a candidate path, title, body, or update
  target from explicit user-provided content, then ask before writing.
- **Limited create then report:** write immediately only when path, title, body,
  and target semantics are explicit enough to form strict runner JSON without
  guessing durable identity.

The POC is evidence for the follow-up document-this eval. It is not a promoted
intake behavior by itself.

## AgentOps Contract

Candidate workflows must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Allowed runner actions are the current public actions: `validate`,
`create_document`, `list_documents`, `get_document`, `append_document`,
`replace_section`, `ingest_source_url`, retrieval `search`,
`provenance_events`, and `projection_states` where applicable.

Routine execution must not use broad repo search, direct SQLite, direct vault
inspection, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, direct file edits, or ad
hoc runtime programs.

Scenario answers and reduced reports must preserve citations, source refs,
provenance, projection freshness, metadata authority, and repo-relative paths
or neutral placeholders such as `<run-root>`.

## Workflow Comparison

| Workflow | User friction | Duplicate risk | Misfile risk | Metadata authority | No-tools compatibility | Freshness/provenance coverage | Strict-runner compatibility |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Ask for missing fields | Medium-high | Low | Low | Strong | Strong | Strong after fields are supplied | Native current workflow |
| Propose before create | Medium | Low-medium | Low-medium | Strong if proposed metadata is explicit | Preserved before write | Strong if proposal includes required checks | Uses existing workflow after approval |
| Limited create then report | Low | Medium | Medium | Strong only when fields are explicit | Not used after write | Requires post-write inspection where relevant | Uses existing workflow only when JSON is complete |

The default safe outcome remains ask-for-missing-fields. Proposal and limited
create-then-report are candidates only when the user has supplied enough
content and target semantics to avoid inventing durable identity.

## Example Prompt And Runner Sequences

### Explicit Document Creation

Prompt:

```text
Create notes/projects/intake-policy.md titled Intake Policy with this body:
# Intake Policy

## Summary
Use strict runner JSON for OpenClerk knowledge intake.
```

Workflow: limited create then report. The user provided path, title, and body.

Runner JSON:

```json
{"action":"validate","document":{"path":"notes/projects/intake-policy.md","title":"Intake Policy","body":"# Intake Policy\n\n## Summary\nUse strict runner JSON for OpenClerk knowledge intake.\n"}}
```

```json
{"action":"create_document","document":{"path":"notes/projects/intake-policy.md","title":"Intake Policy","body":"# Intake Policy\n\n## Summary\nUse strict runner JSON for OpenClerk knowledge intake.\n"}}
```

Expected final answer: report the created vault-relative path and any runner
JSON identifiers returned by the create result.

### Ambiguous "Document This"

Prompt:

```text
Document this: use strict runner JSON for OpenClerk knowledge intake.
```

Workflow: ask for missing fields. The request omits required `document.path`,
`document.title`, and an explicit document body shape.

No runner JSON is allowed. The agent answers once without tools:

```text
I can't create the document yet because document.path, document.title, and
document.body are missing. Provide those fields and I can run the OpenClerk
document runner.
```

This preserves strict binary validation by not asking the runner to infer
missing fields.

### Source URL With Required Hints

Prompt:

```text
Ingest https://example.test/intake.pdf as sources/intake.md with asset
assets/sources/intake.pdf and title Intake Source.
```

Workflow: limited create then report. The user provided source URL, source path
hint, asset path hint, and optional title.

Runner JSON:

```json
{"action":"ingest_source_url","source":{"url":"https://example.test/intake.pdf","path_hint":"sources/intake.md","asset_path_hint":"assets/sources/intake.pdf","title":"Intake Source"}}
```

Expected final answer: report `ingestion.source_path`,
`ingestion.asset_path`, citations, hash, page count, and capture metadata from
the JSON result.

### Source URL Without Required Hints

Prompt:

```text
Document this source: https://example.test/intake.pdf
```

Workflow: ask for missing fields. Source URL create mode is missing
`source.path_hint` and `source.asset_path_hint`.

No runner JSON is allowed. The agent answers once without tools and asks for
the missing source and asset path hints.

### Update Existing Document After Candidate Lookup

Prompt:

```text
Update the existing intake policy note under notes/projects/ with this section:
## Decisions
Use strict runner JSON for document-this intake.
```

Workflow: propose before create or update. The user provided a path prefix and
update content, but not an exact document id. The agent may use runner-visible
candidate lookup because retrieval fields are valid.

Runner JSON:

```json
{"action":"list_documents","list":{"path_prefix":"notes/projects/","limit":20}}
```

If one clear candidate is returned:

```json
{"action":"get_document","doc_id":"DOC_ID_FROM_LIST"}
```

Then the agent proposes the target and update:

```text
I found notes/projects/intake-policy.md as the matching intake policy note.
I can append the Decisions section there.
```

After user approval:

```json
{"action":"append_document","doc_id":"DOC_ID_FROM_LIST","content":"## Decisions\nUse strict runner JSON for document-this intake.\n"}
```

If multiple plausible candidates are returned, the agent must ask which target
to update instead of choosing one silently.

### Duplicate-Risk Handling Before Create

Prompt:

```text
Create a new source note for Intake Policy under sources/ using this body:
# Intake Policy

## Summary
Strict runner JSON keeps intake compatible with OpenClerk.
```

Workflow: propose before create. The user supplied enough content to write, but
the request is nontrivial durable knowledge under a broad prefix. The agent
checks for duplicate risk through runner-visible lookup before writing.

Runner JSON:

```json
{"action":"list_documents","list":{"path_prefix":"sources/","limit":50}}
```

```json
{"action":"search","search":{"text":"Intake Policy strict runner JSON","path_prefix":"sources/","limit":10}}
```

If no duplicate is visible, the agent proposes an exact path and title before
creating. If a likely duplicate exists, the agent asks whether to update the
existing document or create a new document at a user-provided path.

### Synthesis-Style Freshness And Provenance Checks

Prompt:

```text
Update the existing synthesis page about intake policy using the current
sources.
```

Workflow: propose before update. The user named a synthesis intent but not an
exact target. The agent must discover the target and freshness state through
runner JSON before proposing an update.

Runner JSON:

```json
{"action":"search","search":{"text":"intake policy","path_prefix":"sources/","limit":10}}
```

```json
{"action":"list_documents","list":{"path_prefix":"synthesis/","limit":20}}
```

```json
{"action":"get_document","doc_id":"SYNTHESIS_DOC_ID_FROM_LIST"}
```

```json
{"action":"projection_states","projection":{"projection":"synthesis","ref_kind":"document","ref_id":"SYNTHESIS_DOC_ID_FROM_LIST","limit":5}}
```

```json
{"action":"provenance_events","provenance":{"ref_kind":"document","ref_id":"SYNTHESIS_DOC_ID_FROM_LIST","limit":20}}
```

The agent proposes the target, source refs, freshness state, and intended
section replacement. After approval, it may use `replace_section` or
`append_document`. It must not create duplicate synthesis when an existing
synthesis target is visible.

## Candidate Skill Guidance

Skill guidance for document-this intake should remain compatible with the
existing no-tools rule:

- If required fields are missing, ask once without tools and name the missing
  fields.
- If path, title, and body are explicit, use `validate` or `create_document`
  through `openclerk document`.
- If source URL create mode lacks path or asset hints, ask for them; do not
  invent hints.
- If the request names an existing target loosely, use runner list/search/get
  only when retrieval fields are valid, then ask before writing if the target
  is ambiguous.
- Before creating nontrivial durable knowledge, check duplicate risk through
  runner-visible list/search/get when the workflow is already valid.
- For synthesis-style updates, inspect source evidence, existing synthesis,
  projection freshness, and provenance before repair.
- Never use direct vault inspection, direct SQLite, broad repo search, direct
  file edits, source-built runners, HTTP/MCP bypasses, unsupported transports,
  backend variants, module-cache inspection, or ad hoc runtime programs for
  routine intake work.

This guidance is POC reference material. It does not update
`skills/openclerk/SKILL.md`.

## Failure Modes And Classification

Failures must be classified as:

- `none`: the workflow completed with existing runner actions and preserved the
  AgentOps contract.
- `skill_guidance_or_eval_coverage`: the workflow failed because instructions,
  wording, or verifier coverage were too thin, but the runner surface was
  sufficient.
- `data_hygiene_or_fixture_gap`: the workflow failed because fixture documents,
  source refs, metadata, or seeded provenance/freshness state were missing or
  inconsistent.
- `eval_contract_violation`: the agent bypassed the runner contract, used
  prohibited tools, omitted required no-tools handling, or wrote unsupported
  JSON.
- `runner_capability_gap`: existing document and retrieval actions cannot
  express the required intake behavior while preserving strict validation,
  duplicate avoidance, metadata authority, citations or source refs,
  provenance, and freshness.

Promotion is justified only by repeated targeted `runner_capability_gap`
failures. Awkward but successful multi-step workflows, missing examples,
ambiguous prompts, or thin fixture data are not promotion evidence.

## Expected Eval Follow-Up

The follow-up `oc-u9l` eval should pressure-test:

- no-tools clarification for ambiguous "document this" prompts
- explicit path/title/body creation
- source URL intake with and without required hints
- proposal-before-create for likely duplicate documents
- existing-document update with one candidate, multiple candidates, and no
  candidates
- synthesis-style update with source refs, projection freshness, and
  provenance inspection
- final-answer-only rejection for invalid limits and bypass requests

The eval should end with a decision to promote, defer, kill, or keep as
reference. Any promoted implementation must name an exact public surface and
remain separate from this POC.
