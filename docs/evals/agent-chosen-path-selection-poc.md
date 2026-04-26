# Agent-Chosen Path Selection POC

## Status

Implemented targeted POC/eval lane for post-v1 path-selection evidence. The
current reduced report is
[`results/ockp-agent-chosen-path-selection-poc.md`](results/ockp-agent-chosen-path-selection-poc.md).

This document does not add runner actions, schemas, storage migrations, skill
behavior, public API, or release-blocking production gates. The lane uses
harness-generated throwaway fixtures only.

The design decision is recorded in
[`../architecture/agent-chosen-vault-path-selection-adr.md`](../architecture/agent-chosen-vault-path-selection-adr.md).

## Purpose

This POC should determine whether explicit user-provided document paths are
structurally insufficient for routine OpenClerk knowledge work. The target
pressure is user intent that names material or desired knowledge but omits the
durable vault-relative path.

The POC follows the deferred-capability pattern: start with current
`openclerk document` and `openclerk retrieval` workflows, add targeted pressure
only where path choice may be the suspected failure mode, classify failures,
and end with a promote, defer, kill, or keep-as-reference decision.

## AgentOps Contract

Executable scenarios must use only installed OpenClerk runner JSON:

- `openclerk document`
- `openclerk retrieval`

Routine execution must not use broad repo search, direct SQLite, direct vault
inspection, source-built runner paths, HTTP/MCP bypasses, unsupported
transports, backend variants, module-cache inspection, or ad hoc runtime
programs.

Scenario answers and reduced reports must preserve citations, source refs,
provenance, projection freshness, metadata authority, and repo-relative paths
or neutral placeholders such as `<run-root>`.

Run the targeted lane from the repository root with pinned tools:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario explicit-fields-path-title-type,missing-path-title-type-reject,url-only-documentation-path-proposal,url-only-documentation-autonomous-placement,multi-source-synthesis-path-selection,ambiguous-document-type-path-selection,user-path-instructions-win,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-agent-chosen-path-selection-poc
```

The reduced reports are written under `docs/evals/results/` by default. This
targeted lane is not a release-blocking production gate replacement.

Current decision: keep as reference. The refreshed policy comparison records
mixed behavior across explicit fields, missing-field clarification,
proposal-before-create wording, autonomous source placement, multi-source
synthesis path selection, ambiguous metadata-authority placement, explicit user
path precedence, invalid-limit rejection, and bypass rejection. The latest
evidence does not prove a runner capability gap or justify a public surface
change; failures remain classified as guidance/eval or durable-evidence
pressure rather than promotion evidence. No runner action, schema, migration,
storage API, product behavior, public interface, or missing-field policy change
is promoted from this evidence.

## Naming/Path/Title Policy Under Test

The policy under evaluation is:

- user-provided paths or naming instructions always win
- otherwise the agent chooses a clear, stable, vault-relative slug under the
  best conventional home
- the agent chooses a title from user instructions, source metadata, or concise
  human-readable subject text
- the agent reports the chosen path and title
- metadata, not filename, determines document type and identity
- filenames and directories remain conventions only

The POC compares four interaction shapes:

- **Explicit fields required:** create only after path, title, and document type
  are provided.
- **Ask for missing fields:** name missing path, title, and document type fields
  without tools, then wait for user input.
- **Propose before create:** propose the chosen path/title before writing.
- **Create then report:** create at the chosen path/title, then report it.

## Risk Matrix

| Interaction shape | Duplicate risk | Misfile risk | User friction | Metadata authority | Provenance/freshness | No-tools validation | Runner compatibility |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Explicit fields required | Low | Low | High | Strong | Strong | Strong | Native current workflow |
| Ask for missing fields | Low | Low | Medium-high | Strong | Strong | Strong | Native current workflow |
| Propose before create | Low-medium | Low-medium | Medium | Strong if metadata is explicit | Strong if citations/source refs are proposed | Preserved before write | Uses existing workflow after approval |
| Create then report | Medium-high | Medium-high | Low | Requires verifier pressure | Requires post-write inspection | Not applicable after write | Uses existing workflow, but autonomy risk is highest |

## Scenario Families

- `explicit-fields-path-title-type`: provide path, title, and document type up
  front, create exactly that document, and verify no autonomous `sources/` or
  `synthesis/` placement occurs.
- `missing-path-title-type-reject`: omit path, title, and document type, then
  verify the agent names the missing fields without tools.
- `url-only-documentation-path-proposal`: use the required two-URL prompt,
  derive `sources/openai-harness-and-prompt-guidance.md`, and ask before
  creating. The scenario verifies that no document is written and no unsupported
  runner action is implied.
- `url-only-documentation-autonomous-placement`: use the same URL-only
  documentation pressure, create through existing document workflow at a clear
  chosen `sources/` path, and report that path. The scenario measures duplicate,
  misfile, and missing-citation risk without network fetching.
- `multi-source-synthesis-path-selection`: create or update source-linked
  synthesis from several canonical sources while selecting a stable
  `synthesis/` path only when no user path is provided. The scenario must
  preserve `type: synthesis`, `status: active`, `freshness: fresh`,
  single-line `source_refs`, `## Sources`, `## Freshness`, and synthesis
  projection inspection.
- `ambiguous-document-type-path-selection`: pressure cases where the same user
  intent could be a source note, synthesis page, decision, service, or generic
  record-shaped document. The scenario must prove metadata/frontmatter, not
  filename, determines the decision identity.
- `user-path-instructions-win`: provide explicit path or naming instructions
  and verify the agent does not override them with autonomous conventions.
- `bypass-and-validation-pressure`: reject direct SQLite, direct vault,
  HTTP/MCP, source-built runner, unsupported transport, invalid-limit, and
  lower-level workflow requests under the existing no-tools and
  final-answer-only policy.

## Pass/Fail Gates

The POC supplies promotion evidence only if repeated targeted failures show
that explicit user-provided paths are structurally insufficient for routine
knowledge work.

Failures must be classified as:

- data hygiene
- skill guidance
- eval coverage
- runner capability gap

Promotion is not justified by awkward but successful multi-step workflows,
missing instructions, weak fixture data, or evaluator pressure that bypasses
the AgentOps contract.

The candidate should be deferred or killed if it weakens metadata authority,
creates duplicate/conflicting durable knowledge, hides provenance or freshness,
drops source refs or citations, makes path conventions a second type system, or
requires direct SQLite, direct vault inspection, HTTP/MCP, source-built runner
paths, backend variants, module-cache inspection, unsupported actions, or ad
hoc runtime programs.

## Expected Decision Output

A completed targeted report should record:

- the selected scenario set and control prompts
- which runner-visible evidence was used
- how the four interaction shapes compared
- whether failures were capability gaps or non-product gaps
- the decision: promote, defer, kill, or keep as reference
- the exact follow-up implementation surface only if promotion is justified

The current report keeps agent-chosen path selection as reference evidence. The
refreshed policy comparison records mixed behavior across all four interaction
shapes and does not expose a path/title runner capability gap.
