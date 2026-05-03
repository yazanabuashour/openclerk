# Deferred Capability Promotion Gates

## Status

Accepted as the decision method for deferred OpenClerk capabilities.

This document defines how OpenClerk decides whether to promote capabilities
that are intentionally outside the v1 AgentOps runner slice. It is a gate
contract, not a new public API.

## Scope

These gates apply to:

- Mem0 or a memory API
- autonomous router behavior
- semantic graph layers as truth
- broad contradiction engines
- document history and review controls
- new public runner actions

The default decision is to keep each capability as reference or deferred, but
that default is no longer allowed to hide repeated prompt choreography in
`SKILL.md`. When routine work needs exact JSON, literal shell commands,
"skip setup discovery" instructions, or workflow-specific skill recipes, treat
that as product UX evidence. The next step is normally a candidate-surface
comparison before adding more durable skill procedure.

Promotion requires targeted AgentOps evidence through one of two paths:

- **Capability gap:** the existing `openclerk document` and
  `openclerk retrieval` actions cannot safely express the workflow.
- **Ergonomics gap:** the existing actions can technically express the
  workflow, but repeated evidence shows the workflow is too slow, too many
  steps, too scripted, too error-prone, or too guidance-dependent for routine
  AgentOps use.

## Skill Budget And Workflow Action Promotion

`skills/openclerk/SKILL.md` is a router and safety contract, not the durable
home for long product workflows. Its budget is:

- core guardrails and no-tools boundaries
- a compact index of promoted `openclerk document` and `openclerk retrieval`
  actions
- short runner-use hints needed to avoid unsafe transports or durable writes

Detailed multi-step recipes should move into runner workflow actions or
maintainer/user docs. Routine repeated behavior belongs in runner-owned
workflow actions. Adding substantial skill content is acceptable only when it
documents an already-promoted runner surface in compact index form, bridges a
temporary safety gap, or is paired with an explicit candidate comparison
against a narrow runner workflow action. Long request examples, field catalogs,
and workflow recipes should not live in the skill after the corresponding
runner action is promoted.

Treat skill growth as UX debt when the new text teaches agents to perform a
routine workflow by memorizing exact command order, exact JSON, exact shell
phrases, or scenario-specific prompts. A passing eval that relies on that
skill content proves capability or safety only; it does not prove UX quality
unless a natural prompt also passes with acceptable step count, assistant-turn
count, and retry risk. If the natural row passes but still exceeds the
low-ceremony threshold, classify it as taste debt rather than acceptable UX.

The preferred long-term shape for repeated workflows is narrow, runner-owned,
local-first, JSON-only action support that preserves source authority,
citations, provenance, freshness, duplicate handling, runner-only access, and
approval-before-write. Existing primitives should remain for advanced or
manual cases.

## Shared Rubric

Use the same decision rubric for every deferred capability:

- **Promote via capability gap** when repeated targeted AgentOps eval failures
  show the existing document/retrieval workflow is structurally insufficient,
  not merely underspecified, missing data, or missing eval coverage.
- **Promote via ergonomics gap** when the current workflow is expressible but
  repeated targeted evidence shows unacceptable step count, latency, prompt
  brittleness, retry risk, or guidance dependence, and a proposed surface
  reduces that cost without weakening authority, citations, provenance,
  freshness, or local-first operation.
- **Defer** when current runner actions pass with acceptable ergonomics,
  failures are data hygiene, ordinary skill-guidance gaps, or eval coverage
  gaps, or the evidence is too narrow to justify a production surface. Treat
  skill guidance as ordinary only when clarifying existing runner usage is
  enough; repeated need for workflow-specific prompt choreography or skill
  intervention is ergonomics-gap evidence. Defer is not acceptable when the
  only passing shape requires exact runner recipes and a natural prompt has
  not shown acceptable UX.
- **Kill** when the capability mostly duplicates docs retrieval, weakens source
  authority, hides provenance or freshness, increases duplicate/conflicting
  truth, or encourages routine bypasses.
- **Keep as reference** when the capability is useful benchmark pressure but
  does not justify implementation.

No promoted implementation work should be filed from this document alone. A
separate follow-up Bead is allowed only after a targeted eval report and
decision note identify the exact promoted surface and its gates.

## Taste Review Checkpoint

Before any defer or keep-as-reference outcome is accepted, run a qualitative
taste review in addition to the safety and capability checks. The review asks
whether a normal user would reasonably expect a simpler OpenClerk surface and
whether the current workflow is safe but surprisingly indirect. A decision note
or targeted report should record the result even when the final outcome stays
defer or reference.

Use `oc-v1ed` as the reference correction: public HTML/web-page URLs became
part of the existing `ingest_source_url` runner-owned intake surface, and a
user-provided public URL became sufficient permission for read, fetch, and
inspect through the runner. The approval boundary stayed at durable writes,
credentialed or private access, external egress beyond the configured runner
policy, purchase or account actions, and irreversible mutation.

The taste review should flag possible UX debt when:

- an adjacent input naturally belongs under an existing runner action, but the
  plan declares it unsupported instead of extending that action
- an eval row completes with `none` classification but still needs high step
  count, high latency, exact prompt choreography, or repeated assistant turns
- a scripted or exact-command row passes, but the matching natural prompt row
  fails or needs repair
- the proposed fix is substantial `SKILL.md` growth instead of a candidate
  workflow action comparison
- the agent asks for approval before a read, fetch, or inspect step when the
  real safety boundary is a durable write or privileged action
- a workflow is technically expressible only through ceremony that routine
  users would not expect

Taste debt does not automatically authorize implementation. It creates
follow-up audit, design, or eval backlog unless the targeted evidence and a
promotion decision name the exact smoother surface and show that authority,
citations, provenance, freshness, local-first operation, duplicate handling,
runner-only access, and approval-before-write remain intact. In particular, a
successful but ceremonial eval pass can justify more evaluation or design work,
but not a runner action, schema, storage, or skill behavior change by itself.

## Non-Promotion Follow-Up Loop

After any defer, keep-as-reference, or other non-promotion outcome, state which
follow-up category applies:

- **No need:** the current OpenClerk surface is sufficient, and the evaluated
  capability does not represent a valid unresolved user need.
- **Need exists, evaluated shape is wrong:** the user need remains real, but
  the tested surface should not advance.
- **Need exists, candidate comparison required:** the user need remains real,
  and the next step is a Beads comparison epic with ADR, POC, Eval, and
  Decision children.
- **Candidate selected for future promotion evidence:** candidate comparison
  has selected a shape for later targeted promotion evidence, but has not
  authorized implementation.

When a real capability, ergonomics, safety, auditability, or workflow gap
remains, create or propose the comparison epic before handoff. The comparison
should normally evaluate 2-3 plausible candidate surfaces unless the decision
documents why only one shape is viable. Its Decision child must choose the best
candidate, combine useful behaviors where appropriate, defer or kill the track,
or record `none viable yet`.

This loop creates audit, design, and eval backlog only. Implementation remains
blocked until a later accepted decision names the exact OpenClerk runner
action, skill behavior, schema, storage behavior, or public interface and its
gates. Candidate comparison must preserve authority, citations, provenance,
freshness, local-first operation, duplicate handling, runner-only access,
approval-before-write, public-source and synthetic-fixture boundaries, and
rejection of lower-level bypasses.

## Intake Ladder

Use the post-`oc-v1ed` intake ladder when evaluating URL, source, and artifact
workflows. The ladder keeps fetch permission separate from durable-write
approval:

1. Public HTTP/HTTPS source URLs may be fetched and inspected only through the
   installed runner surface, currently `openclerk document`
   `ingest_source_url` for PDF and public HTML/web-page sources.
2. The runner owns normalization, duplicate source URL checks, canonical source
   note creation, citation/search evidence, provenance, and freshness
   visibility.
3. Durable writes require a complete runner request or an approved candidate
   workflow. Missing path hints, asset hints, body content, or target identity
   still clarify before writing.
4. Private URLs, authenticated access, account state, captcha, paywall access,
   cart state, checkout, purchases, browser automation, direct acquisition,
   direct vault writes, direct SQLite, source-built runners, HTTP/MCP bypasses,
   and unsupported transports reject or defer.

The audit applying this ladder to URL and artifact intake is recorded in
[`post-oc-v1ed-url-artifact-intake-audit.md`](post-oc-v1ed-url-artifact-intake-audit.md).

## Required Invariants

Every candidate must preserve the current AgentOps invariants:

- citations, source refs, or stable source identifiers remain attached to
  source-sensitive claims
- provenance and projection freshness stay inspectable
- canonical docs and promoted canonical records outrank synthesis, memory,
  graph state, and routing choices
- routine agents do not use direct SQLite, broad repo search, source-built
  runner paths, HTTP/MCP bypasses, backend variants, module-cache inspection,
  or ad hoc runtime programs
- invalid routine requests still preserve the OpenClerk no-tools contract:
  missing required fields clarify, while invalid limits and bypass requests
  reject

If a candidate cannot preserve these invariants, kill or defer it.

## Ergonomics Evidence

Every targeted POC and eval for a deferred capability should report both
technical expressibility and UX acceptability. The minimum ergonomics scorecard
is:

- tool or command count for the current workflow and candidate surface
- assistant calls and wall time
- amount of prompt specificity required to make the workflow pass
- whether a natural user-intent prompt passes without scripting every runner
  step
- retry or brittleness indicators, including duplicate creation, skipped
  freshness inspection, dropped citations, or wrong target selection
- failure classification separated into data hygiene, ordinary skill guidance,
  eval coverage, capability gap, ergonomics gap, workflow choreography gap,
  skill bloat risk, ergonomics gap despite capability pass, or contract
  violation
- authority, provenance, freshness, privacy, and bypass risks introduced by
  any proposed surface
- `safety_pass`, `capability_pass`, and `ux_pass` or `ux_quality` as separate
  conclusions, so a technically passing workflow can still be recorded as
  taste debt

Use these three conclusions consistently:

- **Safety pass:** the workflow preserved authority, citations, provenance,
  freshness, local-first operation, duplicate handling, runner-only access, and
  approval-before-write.
- **Capability pass:** current primitives can technically express the workflow
  without a new public runner action, schema, storage behavior, or transport.
- **UX quality:** the workflow is acceptable for routine use, or it is recorded
  as taste debt because it is too ceremonial, high-touch, slow, brittle, or
  guidance-dependent.

Use these additional classifications when applicable:

- `workflow_choreography_gap`: exact JSON, literal shell commands, exact
  command ordering, "skip setup discovery" instructions, or other prompt
  choreography are required for routine success.
- `skill_bloat_risk`: the likely fix is a larger `SKILL.md` workflow recipe,
  and that recipe has not been compared against a narrow runner action.
- `ergonomics_gap_despite_capability_pass`: the runner can technically express
  the workflow and safety passed, but the observed UX is too ceremonial for
  routine users.

Ergonomics promotion is not a shortcut around safety. A smoother surface should
be killed or deferred if it creates a second truth system, hides provenance,
drops citations, hides freshness, or normalizes lower-level bypasses.

## Revisit Triggers

Apply this budget to prior non-promotion decisions before closing release prep.
At minimum, source-linked synthesis maintenance, source-sensitive audit repair,
multi-source synthesis creation, stale/fresh synthesis inspection, and
records/provenance/decision evidence bundles need candidate-surface comparison
Beads when current evidence shows agents repeatedly compose lookup,
provenance, projection, and repair steps by prompt choreography.

Each comparison should evaluate:

- Candidate A: keep current primitives and shrink `SKILL.md` to router and
  safety contract
- Candidate B: promote one narrow workflow action
- Candidate C: combine one narrow action with existing primitives for
  advanced/manual cases

## Capability-Specific Proof Obligations

### Mem0 Or Memory API

Promotion requires evidence that repeated recall improves real workflows after
canonicalization. Memory must remain recall, not authority. The candidate must
expose source refs, promotion path, temporal status, and stale or superseded
state before memory-derived output is trusted.

Kill or defer the candidate if it introduces memory-first truth, hides stale
canonical evidence behind ranking, cannot cite canonical docs or records, or
requires routine agents to use memory transports outside AgentOps.

### Autonomous Router

Promotion requires evidence that routing improves correctness over explicit
runner-action choice while staying explainable and audited. The candidate must
show why each source was chosen and must not invent precedence rules separate
from canonical docs, promoted records, provenance, and freshness.

Kill or defer the candidate if it becomes a hidden classifier, performs opaque
multi-store fanout, silently promotes memory, or routes around the runner.

### Semantic Graph Layer

Promotion requires evidence that richer graph semantics beat search, markdown
links, backlinks, and existing `graph_neighborhood` for relationship-shaped
tasks. Canonical markdown must remain the semantic authority, and graph output
must preserve source refs plus projection freshness.

Kill or defer the candidate if semantic edges become independent truth, lack
source evidence, hide stale graph state, or behave like a more complicated way
to do docs retrieval.

### Broad Contradiction Engine

Promotion requires evidence that a broader contradiction workflow beats the
existing source-sensitive audit path without inventing unsupported semantic
truth. Current-source conflicts with no runner-visible authority must remain
unresolved and explainable rather than forced to a winner.

Kill or defer the candidate if it makes arbitrary semantic contradiction
claims, drops source paths, hides supersession/freshness evidence, or creates
unrepairable conflict state.

### Document History And Review Controls

Promotion requires evidence that semantic document lifecycle control is needed
for agent-authored durable edits beyond storage-level Git, sync, snapshots, or
backups. The candidate must explain what changed, why it changed, which source
refs or citations justified it, which before/after content hashes or references
identify the change, which actor or source produced it, what review state
applies, and what derived projections or synthesis pages became stale.

Defer the candidate if existing document and retrieval workflows can express
the scenario through registered documents, append or replace-section,
provenance events, projection states, and operator review with acceptable
ergonomics. Defer also when failures are ordinary missing skill guidance, data
hygiene, thin dogfooding evidence, or eval coverage gaps.

Kill the candidate if it duplicates Git as byte-level history, hides review
state, enables hidden autonomous rewrites, drops source refs or citations,
hides provenance or freshness, exposes private raw diffs in public artifacts,
or requires routine direct SQLite, direct vault inspection, source-built runner
paths, HTTP/MCP bypasses, backend variants, module-cache inspection, or ad hoc
runtime programs.

### New Public Runner Actions

Promotion requires repeated failures that show existing multi-step document and
retrieval workflows cannot express the needed behavior, or repeated ergonomics
evidence that those workflows are too costly for routine use despite being
technically expressible. Any proposed action must include an exact JSON request
shape, backward compatibility expectations, failure modes, and targeted eval
gates.

Kill or defer the candidate if the existing actions pass with acceptable
ergonomics, the pressure comes from ordinary missing skill guidance, or the
proposed action would create a second authority surface.

## Prompt And Eval Pattern

Future POCs for deferred capabilities must follow this pattern:

1. Start with a control prompt that solves the workflow using only
   `openclerk document` and `openclerk retrieval`.
2. Add at least one natural user-intent prompt and one scripted-control prompt.
   The natural prompt measures UX and brittleness; the scripted control proves
   whether current primitives can still work with exact instructions.
3. Require the agent to use runner JSON evidence and preserve citations,
   source refs, provenance, and freshness where relevant.
4. Record tool/command count, assistant calls, wall time, prompt specificity,
   and retry or brittleness signals.
5. Classify failures as data hygiene, ordinary skill guidance, eval coverage,
   capability gap, ergonomics gap, or contract violation.
6. Record the taste review separately from pass/fail classification: safety
   pass, capability pass, and UX quality.
7. Record targeted evidence under `docs/evals/results/` using repo-relative
   paths and `<run-root>` placeholders.
8. End with an explicit decision: promote, defer, kill, or keep as reference.
9. For defer, keep-as-reference, or another non-promotion outcome, record the
   applicable non-promotion follow-up category and file or propose a comparison
   epic when the need remains valid.
10. If promoted, file a separate implementation Bead that names the exact
   surface and gates.

This keeps capability pressure measurable without letting interesting reference
behavior become production scope by default.
