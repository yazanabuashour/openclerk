# Skill Reduction Into Runner Heuristics Eval

Date: 2026-05-04

## Scenario

Evaluate whether OpenClerk can reduce `SKILL.md` workflow detail by promoting
a runner-owned intent-to-surface guide while preserving the no-tools and
runner-only boundaries.

## Result

Promote `workflow_guide_report` and shrink `SKILL.md` from 250 lines to 216
lines. Tighten the skill budget test from 250 lines to 225 lines.

## Safety Pass

Pass.

The selected report is read-only and does not inspect documents, query
storage, fetch URLs, create candidates, create or update documents, repair
synthesis, or execute the selected action. The skill still owns bootstrap
no-tools rejections and lower-level bypass rejection.

## Capability Pass

Pass.

The report maps routine intent to promoted runner-owned surfaces, including
duplicate-candidate checks, public URL placement, source-linked synthesis,
source audit, evidence bundles, memory/router recall, structured-store
decision support, hybrid retrieval decision support, and current primitives.
It returns request-shape guidance and `agent_handoff`.

## UX Quality

Pass.

The runner now owns the surface-selection heuristic that would otherwise live
as repeated skill prose. Agents can use one JSON report before running the
selected action, and the skill remains a compact activation, routing, and
safety contract.

## Evidence Posture

The proof is committed code, tests, help text, README guidance, skill guidance,
and docs. It does not include raw logs, private data, storage snapshots, or
machine-specific artifact paths.

Relevant tests:

- `TestRetrievalTaskWorkflowGuideReportRoutesIntentWithoutStore`
- `TestRetrievalTaskWorkflowGuideReportRejectsMissingIntent`
- `TestRetrievalTaskWorkflowGuideReportPrioritizesSpecificDecisionSurfaces`
- `TestRunnerDocumentAndRetrievalJSONRoundTrip`
- `TestReadOnlyActionsDoNotTakeRunnerWriteLock`
- `TestOpenClerkSkillStaysWithinThinRouterBudget`

## Decision

Select `workflow_guide_report` as the promoted surface. Keep future workflow
detail out of `SKILL.md` unless it is compact safety/routing guidance or is
paired with a candidate-surface comparison.
