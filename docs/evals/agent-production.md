# Production Agent Evaluation Plan

This project evaluates agent behavior against the same surface a real OpenClerk
agent receives. The production eval must use only the installed
`skills/openclerk` payload and a fresh session. Do not add hidden evaluator
instructions that tell the agent which API to call unless those instructions are
also present in the production skill.

## Primary Production Eval

- Start from a fresh agent session with the production `openclerk` skill.
- Provide natural user prompts such as "create a project note for OpenClerk",
  "what do I know about the roadmap?", "append a decision", or "show provenance
  for this document".
- Include compounding-knowledge prompts such as "ingest this source into the
  OpenClerk synthesis", "update the existing synthesis with this new evidence",
  "file this answer for reuse", and "lint the synthesis for stale claims".
- Use the normal local Go/tool environment and default OpenClerk data path unless
  the scenario explicitly provides a temporary data directory.
- Judge success by final vault/database state, citation quality, duplicate-path
  behavior, projection freshness, tool calls, assistant calls, wall time,
  non-cache input tokens, and whether the agent read generated files or the Go
  module cache.
- The expected production path is `local.OpenClient(...)` plus code-first helper
  methods on `*local.Client`.

## Isolated Variants

Keep comparison variants outside the production skill so the real skill stays
opinionated and narrow.

- Baseline A: current or archived generated-client skill surface.
- Variant B: production code-first SDK skill surface.
- Variant C: CLI-oriented harness or alternate skill payload, if a CLI-oriented
  workflow is added.
- Variant D: implementation-fixture comparisons for `fts`, `hybrid`, `graph`,
  and `records`.

Each variant should have its own skill payload or harness instructions. Do not
combine generated-client, SDK, CLI, and implementation-fixture recipes in the
same production skill.

## Core Scenarios

- Create canonical notes with stable paths, frontmatter, headings, and body
  content, then verify they exist in the vault and document registry.
- Search before creating a synthesis page, then create or update source-linked
  synthesis with citations/source refs to canonical evidence.
- Add a newer source that contradicts or supersedes an older synthesis claim and
  verify the agent surfaces the conflict instead of blindly trusting the older
  synthesis.
- Repeat the same create request and assert duplicate paths fail clearly instead
  of silently overwriting canonical Markdown.
- Search for a user concept and verify citation paths and line ranges point back
  to the correct Markdown source.
- Ask a source-grounded question, file the reusable answer back into a synthesis
  page, and verify a later query uses the filed synthesis plus citations rather
  than rediscovering the answer from scratch.
- Append a new section and replace an existing section without losing unrelated
  document content.
- Inspect document links and graph neighborhood output for linked notes.
- Create a promoted-record-shaped document and verify records lookup returns the
  expected entity, facts, and citations.
- Inspect provenance events and projection states after document create/update
  workflows.
- Run a wiki-health pass that looks for stale synthesis, missing source refs,
  orphan pages, and missing cross-links.
