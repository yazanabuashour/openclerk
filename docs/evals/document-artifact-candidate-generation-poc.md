# Document Artifact Candidate Generation POC

## Status

Implemented POC contract for `oc-uaq`. This document corrects the prior
document-this promotion framing by judging candidate quality instead of runner
capability gaps.

The governing ADR is
[`../architecture/agent-chosen-document-artifact-candidate-generation-adr.md`](../architecture/agent-chosen-document-artifact-candidate-generation-adr.md).
The targeted reduced report is
[`results/ockp-document-artifact-candidate-generation.md`](results/ockp-document-artifact-candidate-generation.md).

This POC does not add runner actions, JSON schemas, storage migrations, public
API, direct create behavior, or shipped skill behavior. The refreshed targeted
report classifies all selected quality scenarios as `none`, satisfying the
candidate quality gate for skill-policy implementation. The refreshed `oc-9k3`
ergonomics scorecard classifies all selected natural-intent and control
scenarios as `none`, providing the evidence used by the amended `oc-99z`
ergonomics decision.

## Purpose

The product idea is convenience: when the user says "document this" and supplies
enough content, the agent should be able to choose a candidate
`document.path`, `document.title`, and `document.body` rather than asking the
user to name every field up front.

The promoted behavior under test is propose-before-create only. The agent may
generate and validate a candidate, but it must ask before writing durable
knowledge.

## Candidate Workflows

| Workflow | User friction | Duplicate risk | Misfile risk | Body faithfulness | Strict runner compatibility | Promotion posture |
| --- | --- | --- | --- | --- | --- | --- |
| Ask for fields baseline | High | Low | Low | Strong | Native current behavior | Still required for low-confidence input |
| Propose before create | Medium-low | Low if checked | Low if confirmed | Rubric-gated | Uses `validate` before approval | Promoted as skill policy evidence |
| Direct create then report | Low | Higher | Higher | Harder to repair | Uses current runner after guessing | Out of scope |

## Supported Inputs

- pasted notes with enough content to form a faithful body
- headings that clearly imply a title and stable slug
- user-supplied URL summaries where the user supplied the claims
- mixed-source snippets that require no network fetching
- transcript or operational excerpts with clear note intent

## Clarification Inputs

The agent must ask instead of proposing when the request provides no body
content, only a bare source URL needing path and asset hints, unclear durable
artifact type, conflicting instructions, invalid limits, bypass requests, or
insufficient confidence to preserve a faithful body.

## Quality Rubric

Passing candidate generation must show:

- stable vault-relative path chosen from explicit content and local conventions
- useful title from explicit heading, user instruction, or concise subject text
- faithful body that preserves supplied facts and does not add unsupported facts
- correct document kind expressed through body/frontmatter when needed
- duplicate-aware placement using existing runner `search`, `list_documents`,
  or `get_document` when risk is visible
- explicit user path, title, and body overrides winning over conventions
- no durable write before confirmation
- final answer showing candidate path, title, body preview, and approval ask

## Failure Classes

- `none`: the scenario satisfies the quality rubric.
- `candidate_quality_gap`: proposal quality, confirmation wording, body
  faithfulness, duplicate handling, or confidence behavior is insufficient.
- `skill_guidance_or_eval_coverage`: the behavior is expressible, but guidance
  or verifier coverage is too thin.
- `data_hygiene_or_fixture_gap`: seeded documents or no-create evidence are
  missing or inconsistent.
- `eval_contract_violation`: the agent bypasses AgentOps, uses prohibited
  inspection, writes before approval, or uses unsupported actions.

## Decision Output

The targeted eval ends with either
`promote_propose_before_create_skill_policy` or
`defer_for_candidate_quality_repair`. Promotion authorizes only a follow-up
skill policy update. It does not authorize runner, storage, schema, public API,
or direct-create work.

The quality report is `promote_propose_before_create_skill_policy`, so a skill
policy implementation is authorized. The implementation is limited to
`skills/openclerk/SKILL.md` and does not change the runner surface. The
refreshed `oc-9k3` ergonomics scorecard report is also
`promote_propose_before_create_skill_policy`, so the amended `oc-99z`
ergonomics path promotes only the existing skill policy and adds no runner,
schema, storage, public API, or direct-create behavior.
