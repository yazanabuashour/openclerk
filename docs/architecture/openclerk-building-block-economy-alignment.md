---
decision_id: decision-openclerk-building-block-economy-alignment
decision_title: OpenClerk Building Block Economy Alignment
decision_status: accepted
decision_scope: public-runner-surface
decision_owner: agentops
decision_date: 2026-05-05
source_refs: README.md, skills/openclerk/SKILL.md, cmd/openclerk/main.go, https://mitchellh.com/writing/building-block-economy
---
# Decision: OpenClerk Building Block Economy Alignment

## Status

Accepted: OpenClerk should present itself as a narrow local-first runner plus
composable document, retrieval, and module building blocks.

## Decision

OpenClerk exposes a stable `openclerk capabilities` discovery command. The
command emits an `openclerk-capabilities.v1` JSON manifest that lists:

- document, retrieval, and module runner domains
- primitive actions for manual or advanced assembly
- promoted workflow actions for repeated high-touch flows
- optional module extension points
- local-first, citation, provenance, freshness, approval, and bypass
  boundaries

This makes the building-block inventory directly inspectable by agents without
turning `SKILL.md` into a long recipe catalog or relying on ad hoc scraping of
help output.

## Rationale

Mitchell Hashimoto's building-block framing argues for high-quality, robust,
well-documented components that other users and agents can assemble. OpenClerk
already has many such components, but the overall inventory was implicit across
runner help, the skill, module docs, and architecture decisions.

The capabilities manifest turns those scattered signals into one small,
machine-readable surface. It also preserves the mainline product posture:
default lexical search, canonical markdown authority, runner-only access,
approval-before-write, and optional provider modules instead of hidden default
provider behavior.

## Safety, Capability, UX

Safety pass: pass. The manifest is static discovery metadata only; it does not
inspect vault contents, read SQLite, fetch URLs, create documents, install
modules, or answer source-sensitive user questions.

Capability pass: pass. Agents can discover the current blocks and route to the
natural runner action, while actual task evidence still comes from document,
retrieval, or module JSON results.

UX quality: pass. A normal agent or maintainer can ask for the block inventory
once, then assemble or delegate through explicit runner actions. This reduces
prompt choreography and keeps detailed workflows in runner surfaces, docs, and
evals rather than in the activation skill.

## Non-Goals

This decision does not promote:

- semantic retrieval as the default search mode
- hidden remote provider fallback
- direct SQLite or raw vault access
- durable writes without approval
- source-sensitive answers from static capability metadata alone
- long scenario-specific recipes in `skills/openclerk/SKILL.md`
