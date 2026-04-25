# Agent-Chosen Path Selection POC

## Status

Planned targeted POC/eval contract for post-v1 path-selection evidence.

This document does not add fixture data, runner actions, schemas, storage
migrations, skill behavior, public API, or release-blocking production gates.

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

## Naming/Path Policy Under Test

The policy under evaluation is:

- user-provided paths or naming instructions always win
- otherwise the agent chooses a clear, stable, vault-relative slug under the
  best conventional home
- the agent reports the chosen path
- metadata, not filename, determines document type and identity
- filenames and directories remain conventions only

The POC must compare at least two interaction shapes:

- **Path proposal before create:** propose the chosen path before writing.
- **Autonomous create then report:** create at the chosen path, then report it.

## Scenario Families

- `url-only-documentation-path-proposal`: use the required two-URL prompt,
  derive candidate vault-relative placement, and ask before creating. The
  scenario should verify that the agent explains why the proposed conventional
  home fits and that no unsupported runner action is implied.
- `url-only-documentation-autonomous-placement`: use the same URL-only
  documentation pressure, create through existing document workflow at a clear
  chosen path, and report that path. The scenario should measure duplicate,
  misfile, and missing-citation risk.
- `multi-source-synthesis-path-selection`: create or update source-linked
  synthesis from several canonical sources while selecting a stable
  `synthesis/` path only when no user path is provided. The scenario must
  preserve `type: synthesis`, `status: active`, `freshness: fresh`,
  single-line `source_refs`, `## Sources`, and `## Freshness`.
- `ambiguous-document-type-path-selection`: pressure cases where the same user
  intent could be a source note, synthesis page, decision, service, or generic
  record-shaped document. The scenario must prove metadata, not filename,
  determines the promoted identity.
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
- how the two interaction shapes compared
- whether failures were capability gaps or non-product gaps
- the decision: promote, defer, kill, or keep as reference
- the exact follow-up implementation surface only if promotion is justified

Until that report exists, agent-chosen path selection remains deferred.
