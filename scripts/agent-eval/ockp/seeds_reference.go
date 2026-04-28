package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func seedRAGRetrievalBaseline(ctx context.Context, cfg runclient.Config) error {
	currentBody := strings.TrimSpace(`---
type: note
status: active
rag_scope: active-policy
---
# Current AgentOps RAG Policy

## Summary
`+ragCurrentPolicySummary+`

## Decision
`+ragCurrentPolicyDecision+`
`) + "\n"
	if err := createSeedDocument(ctx, cfg, ragCurrentPolicyPath, ragCurrentPolicyTitle, currentBody); err != nil {
		return err
	}
	decoyBody := strings.TrimSpace(`---
type: note
status: draft
rag_scope: decoy-policy
---
# Decoy AgentOps RAG Policy

## Summary
Decoy AgentOps RAG baseline policy marker: this draft says direct SQLite might be acceptable for routine OpenClerk knowledge answers.

## Decision
This is a decoy policy and is not the active AgentOps retrieval decision.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, ragDecoyPolicyPath, ragDecoyPolicyTitle, decoyBody); err != nil {
		return err
	}
	archivedBody := strings.TrimSpace(`---
type: note
status: superseded
rag_scope: archived-policy
---
# Archived AgentOps RAG Policy

## Summary
Archived AgentOps RAG baseline policy marker: older guidance mentioned a source-built command path.

## Decision
This archived policy is outside the active RAG path prefix and is superseded by the current JSON runner policy.
`) + "\n"
	return createSeedDocument(ctx, cfg, ragArchivedPolicyPath, ragArchivedPolicyTitle, archivedBody)
}
func seedRepoDocsDogfood(ctx context.Context, cfg runclient.Config) error {
	repoRoot, err := repoRootFromEvalDatabasePath(cfg.DatabasePath)
	if err != nil {
		return err
	}
	var imported int
	err = filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if !shouldDescendRepoMarkdownDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if !shouldImportRepoMarkdown(rel, entry) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		repoPath := filepath.ToSlash(rel)
		if err := createSeedDocument(ctx, cfg, repoPath, markdownTitle(repoPath, string(content)), string(content)); err != nil {
			return fmt.Errorf("import repo doc %s: %w", repoPath, err)
		}
		imported++
		return nil
	})
	if err != nil {
		return err
	}
	if imported == 0 {
		return errors.New("repo-docs dogfood seed imported no markdown documents")
	}
	return nil
}
func shouldDescendRepoMarkdownDir(rel string) bool {
	slash := filepath.ToSlash(rel)
	if slash == "." {
		return true
	}
	parts := strings.Split(slash, "/")
	if len(parts) == 0 {
		return true
	}
	switch parts[0] {
	case ".git", ".beads", ".dolt", ".agents", ".openclerk-eval":
		return false
	}
	return !strings.HasPrefix(slash, "docs/evals/results/")
}
func repoRootFromEvalDatabasePath(databasePath string) (string, error) {
	if strings.TrimSpace(databasePath) == "" {
		return "", errors.New("missing eval database path")
	}
	evalDir := filepath.Dir(databasePath)
	if filepath.Base(evalDir) != ".openclerk-eval" {
		return "", fmt.Errorf("database path %q is not under .openclerk-eval", databasePath)
	}
	return filepath.Dir(evalDir), nil
}
func shouldImportRepoMarkdown(rel string, entry fs.DirEntry) bool {
	slash := filepath.ToSlash(rel)
	if slash == "." {
		return false
	}
	parts := strings.Split(slash, "/")
	if len(parts) > 0 {
		switch parts[0] {
		case ".git", ".beads", ".dolt", ".agents", ".openclerk-eval":
			return false
		case "AGENTS.md":
			return false
		}
	}
	if strings.HasPrefix(slash, "docs/evals/results/") {
		return false
	}
	return !entry.IsDir() && strings.EqualFold(filepath.Ext(slash), ".md")
}
func markdownTitle(path string, body string) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			if title != "" {
				return title
			}
		}
	}
	base := filepath.Base(path)
	title := strings.TrimSuffix(base, filepath.Ext(base))
	title = strings.ReplaceAll(title, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.TrimSpace(title)
	if title == "" {
		return path
	}
	return title
}
func seedDocsNavigationBaseline(ctx context.Context, cfg runclient.Config) error {
	indexBody := strings.TrimSpace(`---
type: wiki
status: active
---
# AgentOps Wiki Index

## Summary
Canonical directory navigation starts here for the AgentOps wiki baseline.

## Links
- [Runner policy](runner-policy.md)
- [Knowledge plane](../architecture/knowledge-plane.md)
- [Runner playbook](../ops/runner-playbook.md)

## Limits
Folder paths and headings show the local index, but they do not explain backlinks or cross-directory relationship neighborhoods without retrieval actions.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationIndexPath, "AgentOps Wiki Index", indexBody); err != nil {
		return err
	}

	policyBody := strings.TrimSpace(`---
type: policy
status: active
---
# Runner Policy

## Summary
Routine OpenClerk knowledge work uses the installed JSON runner and cites returned source paths.

## Navigation
Return to the [AgentOps wiki index](index.md) and compare with the [knowledge plane](../architecture/knowledge-plane.md).
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationPolicyPath, "Runner Policy", policyBody); err != nil {
		return err
	}

	architectureBody := strings.TrimSpace(`---
type: architecture
status: active
---
# Knowledge Plane

## Summary
The knowledge plane keeps canonical markdown as source authority and derives graph relationships from links.

## Navigation
The [AgentOps wiki index](../agentops/index.md) links this architecture note to runner policy context.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, docsNavigationArchPath, "Knowledge Plane", architectureBody); err != nil {
		return err
	}

	opsBody := strings.TrimSpace(`---
type: runbook
status: active
---
# Runner Playbook

## Summary
Operators use the runner playbook when directory navigation is not enough to explain related policy and architecture docs.

## Navigation
Start from the [AgentOps wiki index](../agentops/index.md) before following graph neighborhoods.
`) + "\n"
	return createSeedDocument(ctx, cfg, docsNavigationOpsPath, "Runner Playbook", opsBody)
}
func seedGraphSemanticsReference(ctx context.Context, cfg runclient.Config) error {
	indexBody := strings.TrimSpace(`---
type: graph-reference
status: active
---
# Graph Semantics Reference

## Summary
Graph semantics requires canonical markdown to carry relationship meaning. This reference note says the routing note supersedes legacy graph claims, is related to freshness evidence, and operationalizes the operations playbook.

## Relationships
- Requires: [Routing](routing.md)
- Supersedes: [Freshness](freshness.md)
- Related to: [Operations](operations.md)
- Operationalizes: Operations playbook

## Decision
Richer graph semantics stay in canonical markdown relationship text. The derived graph should expose structural links and citations, not independent semantic-label authority.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, graphSemanticsIndexPath, "Graph Semantics Reference", indexBody); err != nil {
		return err
	}

	routingBody := strings.TrimSpace(`---
type: graph-reference
status: active
---
# Routing

## Summary
Routing links back to the [Graph Semantics Reference](index.md) because semantic relationship labels should remain inspectable markdown evidence.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, graphSemanticsRoutingPath, "Routing", routingBody); err != nil {
		return err
	}

	freshnessBody := strings.TrimSpace(`---
type: graph-reference
status: active
---
# Freshness

## Summary
Freshness links back to the [Graph Semantics Reference](index.md) so graph projection freshness stays tied to canonical markdown.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, graphSemanticsFreshnessPath, "Freshness", freshnessBody); err != nil {
		return err
	}

	operationsBody := strings.TrimSpace(`---
type: graph-reference
status: active
---
# Operations

## Summary
Operations links back to the [Graph Semantics Reference](index.md) and keeps operationalizes language in source text rather than in opaque graph labels.
`) + "\n"
	return createSeedDocument(ctx, cfg, graphSemanticsOperationsPath, "Operations", operationsBody)
}
func seedMemoryRouterReference(ctx context.Context, cfg runclient.Config) error {
	temporalBody := strings.TrimSpace(`---
type: memory-router-reference
status: active
effective_at: 2026-04-22
---
# Temporal Recall Policy

## Summary
Temporal recall stays source-grounded: current canonical docs and promoted records outrank stale session observations, and agents must name the temporal status before trusting a result.

## Guidance
Current evidence should be described as current or effective. Older or superseded evidence should be described as stale before it is reused.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, memoryRouterTemporalPath, "Temporal Recall Policy", temporalBody); err != nil {
		return err
	}

	feedbackBody := strings.TrimSpace(`---
type: memory-router-reference
status: active
---
# Feedback Weighting

## Summary
Feedback weighting is advisory only. A high-weight remembered result can help rank what to inspect next, but it cannot hide source refs, freshness, provenance, or weaker conflicting evidence.

## Guidance
The reference weight for the session observation is 0.8 because the user marked it useful, but the answer must still cite canonical markdown.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, memoryRouterFeedbackPath, "Feedback Weighting", feedbackBody); err != nil {
		return err
	}

	routingBody := strings.TrimSpace(`---
type: memory-router-reference
status: active
---
# Routing Policy

## Summary
Routing is an explainable choice among existing AgentOps document and retrieval actions. Use canonical docs and provenance for source-sensitive claims, promoted records for typed domains, graph navigation for relationship questions, and never use autonomous routing as hidden authority.

## Guidance
The correct route for this reference POC is canonical docs plus provenance and projection freshness, not a memory-first router.
`) + "\n"
	return createSeedDocument(ctx, cfg, memoryRouterRoutingPath, "Routing Policy", routingBody)
}
func seedMemoryRouterRevisit(ctx context.Context, cfg runclient.Config) error {
	if err := seedMemoryRouterReference(ctx, cfg); err != nil {
		return err
	}
	if err := createSeedDocument(ctx, cfg, memoryRouterSessionObservationPath, memoryRouterSessionObservationTitle, memoryRouterSessionObservationBody()); err != nil {
		return err
	}
	return createSeedDocument(ctx, cfg, memoryRouterSynthesisPath, "Memory Router Reference", memoryRouterReferenceSynthesisBody())
}
func seedPromotedRecordDomainExpansion(ctx context.Context, cfg runclient.Config) error {
	narrativeBody := strings.TrimSpace(`---
type: note
status: active
---
# Promoted Record Domain Policy Narrative

## Summary
Plain docs evidence says the AgentOps escalation policy is important for runner review, but it does not provide typed policy filters or stable record identity.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, promotedRecordDomainNarrativePath, "Promoted Record Domain Policy Narrative", narrativeBody); err != nil {
		return err
	}
	primaryBody := strings.TrimSpace(`---
entity_id: agentops-escalation-policy
entity_type: policy
entity_name: AgentOps Escalation Policy
---
# AgentOps Escalation Policy

## Summary
Promoted record domain expansion policy marker: AgentOps escalation policy owner is platform, status active, review cadence monthly, and citations must stay with canonical markdown.

## Facts
- status: active
- owner: platform
- review_cadence: monthly
- escalation_channel: runner-review
`) + "\n"
	if err := createSeedDocument(ctx, cfg, promotedRecordDomainPrimaryPath, promotedRecordDomainEntityName, primaryBody); err != nil {
		return err
	}
	adjacentBody := strings.TrimSpace(`---
entity_id: agentops-review-policy
entity_type: policy
entity_name: AgentOps Review Policy
---
# AgentOps Review Policy

## Summary
Adjacent policy record for promoted record domain expansion pressure. It is related to runner review but is not the escalation policy.

## Facts
- status: active
- owner: operations
- review_cadence: quarterly
`) + "\n"
	return createSeedDocument(ctx, cfg, promotedRecordDomainAdjacentPath, "AgentOps Review Policy", adjacentBody)
}
func memoryRouterSessionObservationBody() string {
	return strings.TrimSpace(`---
type: source
status: active
observed_at: 2026-04-22
---
# Memory Router Session Observation

## Summary
Session observation: a user asked whether memory routing should promote recall. Useful session material must be promoted only by writing canonical markdown with source refs.

## Feedback
Positive feedback weight 0.8 is advisory only and cannot hide stale canonical evidence.
`) + "\n"
}
func memoryRouterReferenceSynthesisBody() string {
	return strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, notes/memory-router/routing-policy.md
---
# Memory Router Reference

Temporal status: current canonical docs outrank stale session observations.
Session promotion path: durable canonical markdown with source refs.
Feedback weighting: advisory only.
Routing choice: existing AgentOps document and retrieval actions.
Decision: keep memory and autonomous routing as reference/deferred.

## Sources
- notes/memory-router/session-observation.md
- notes/memory-router/temporal-policy.md
- notes/memory-router/feedback-weighting.md
- notes/memory-router/routing-policy.md

## Freshness
Provenance and synthesis projection checks are required before reuse.
`) + "\n"
}
func seedConfiguredLayoutScenario(ctx context.Context, cfg runclient.Config) error {
	sourceBody := strings.TrimSpace(`---
type: source
status: active
---
# Layout Runner Source

## Summary
Convention-first OpenClerk knowledge layout uses runner-visible JSON inspection rather than a committed manifest.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "sources/layout-runner.md", "Layout Runner Source", sourceBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/layout-runner.md
---
# Layout Runner Synthesis

## Summary
The configured layout keeps canonical markdown and source-linked synthesis convention-first.

## Sources
- sources/layout-runner.md

## Freshness
Checked source refs through runner-visible layout inspection.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "synthesis/layout-runner.md", "Layout Runner Synthesis", synthesisBody); err != nil {
		return err
	}
	recordBody := strings.TrimSpace(`---
entity_id: layout-runner-record
entity_type: policy
entity_name: Layout Runner Policy
---
# Layout Runner Policy

## Facts
- status: active
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "records/policies/layout-runner.md", "Layout Runner Policy", recordBody); err != nil {
		return err
	}
	serviceBody := strings.TrimSpace(`---
service_id: layout-runner
service_name: Layout Runner
service_status: active
service_owner: runner
service_interface: JSON runner
---
# Layout Runner

## Summary
Runner-visible layout inspection explains configured knowledge conventions.
`) + "\n"
	return createSeedDocument(ctx, cfg, "records/services/layout-runner.md", "Layout Runner", serviceBody)
}
func seedInvalidLayoutScenario(ctx context.Context, cfg runclient.Config) error {
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
source_refs: sources/missing-layout-source.md
---
# Broken Layout Synthesis

## Summary
This synthesis references a missing source and omits the required freshness section.

## Sources
- sources/missing-layout-source.md
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "synthesis/broken-layout.md", "Broken Layout Synthesis", synthesisBody); err != nil {
		return err
	}
	serviceBody := strings.TrimSpace(`---
service_id: broken-layout-service
---
# Broken Layout Service

## Summary
This service-shaped document is missing service_name.
`) + "\n"
	return createSeedDocument(ctx, cfg, "records/services/broken-layout-service.md", "Broken Layout Service", serviceBody)
}
func seedDecisionRecordVsDocs(ctx context.Context, cfg runclient.Config) error {
	if err := createSeedDocument(ctx, cfg, "notes/reference/runner-decision-narrative.md", "Runner Decision Narrative", "# Runner Decision Narrative\n\n## Summary\nPlain docs evidence mentions several OpenClerk runner decisions, including an accepted JSON runner decision and older alternatives.\n"); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
decision_id: adr-runner-current
decision_title: Use JSON runner
decision_status: accepted
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-22
source_refs: notes/reference/runner-decision-narrative.md
---
# Use JSON runner

## Summary
Accepted decision: routine OpenClerk AgentOps tasks use the installed JSON runner.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "docs/architecture/runner-current-decision.md", "Use JSON runner", currentBody); err != nil {
		return err
	}
	oldBody := strings.TrimSpace(`---
decision_id: adr-runner-old
decision_title: Use retired command path
decision_status: superseded
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-20
superseded_by: adr-runner-current
source_refs: notes/reference/runner-decision-narrative.md
---
# Use retired command path

## Summary
Superseded decision: older agents used a retired command path.
`) + "\n"
	return createSeedDocument(ctx, cfg, "records/decisions/runner-old-decision.md", "Use retired command path", oldBody)
}
func seedDecisionSupersession(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
decision_id: adr-runner-old
decision_title: Use retired command path
decision_status: superseded
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-20
superseded_by: adr-runner-current
source_refs: sources/decision-old.md
---
# Use retired command path

## Summary
Superseded decision: older agents used a retired command path.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "docs/architecture/runner-old-decision.md", "Use retired command path", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
decision_id: adr-runner-current
decision_title: Use JSON runner
decision_status: accepted
decision_scope: agentops
decision_owner: platform
decision_date: 2026-04-22
supersedes: adr-runner-old
source_refs: sources/decision-current.md
---
# Use JSON runner

## Summary
Accepted decision: routine OpenClerk AgentOps tasks use the installed JSON runner.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "records/decisions/runner-current-decision.md", "Use JSON runner", currentBody); err != nil {
		return err
	}
	if err := createSeedDocument(ctx, cfg, "sources/decision-old.md", "Old decision source", "# Old decision source\n\n## Summary\nOlder source documented the retired path.\n"); err != nil {
		return err
	}
	return createSeedDocument(ctx, cfg, "sources/decision-current.md", "Current decision source", "# Current decision source\n\n## Summary\nCurrent source documents the JSON runner path.\n")
}
func seedDecisionRealADRMigration(ctx context.Context, cfg runclient.Config) error {
	agentOpsBody := strings.TrimSpace(`---
decision_id: adr-agentops-only-knowledge-plane
decision_title: AgentOps-Only Knowledge Plane Direction
decision_status: accepted
decision_scope: knowledge-plane
decision_owner: platform
source_refs: sources/agentops-direction.md
---
# ADR: AgentOps-Only Knowledge Plane Direction

## Status
Accepted as the current architecture direction.

## Summary
OpenClerk uses AgentOps as the only production agent interface.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, "docs/architecture/eval-backed-knowledge-plane-adr.md", "AgentOps-Only Knowledge Plane Direction", agentOpsBody); err != nil {
		return err
	}
	configBody := strings.TrimSpace(`---
decision_id: adr-knowledge-configuration-v1
decision_title: Knowledge Configuration v1
decision_status: accepted
decision_scope: knowledge-configuration
decision_owner: platform
supersedes: adr-agentops-only-knowledge-plane
source_refs: sources/knowledge-configuration.md
---
# ADR: Knowledge Configuration v1

## Status
Accepted as the v1 production contract for OpenClerk-compatible knowledge vaults.

## Summary
OpenClerk knowledge configuration v1 is runner-visible and convention-first.
`) + "\n"
	return createSeedDocument(ctx, cfg, "docs/architecture/knowledge-configuration-v1-adr.md", "Knowledge Configuration v1", configBody)
}
func seedDocumentHistoryInspection(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: policy
status: active
---
# Lifecycle Control

## Summary
Document history review controls use current AgentOps document and retrieval evidence first.

## Decision
Initial state: lifecycle inspection is pending evidence.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryPolicyPath, "Lifecycle Control", body); err != nil {
		return err
	}
	return replaceScenarioSeedSection(ctx, cfg, documentHistoryPolicyPath, "Decision", "Current state: lifecycle inspection uses list_documents, get_document, provenance_events, and projection_states before any new history action is proposed.")
}
func seedDocumentHistoryDiffReview(ctx context.Context, cfg runclient.Config) error {
	previousBody := strings.TrimSpace(`---
type: source
status: superseded
superseded_by: notes/history-review/diff-current.md
---
# Previous Diff Evidence

## Summary
Previous lifecycle guidance said human review was optional for low-risk durable edits.

## Evidence
The prior semantic position was optional review.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryDiffPreviousPath, "Previous Diff Evidence", previousBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
type: policy
status: active
supersedes: sources/history-review/diff-previous.md
source_refs: sources/history-review/diff-previous.md
---
# Current Diff Evidence

## Summary
Current lifecycle guidance says human review is required before source-sensitive durable edits become accepted knowledge.

## Evidence
The current semantic position is required review with citations and source refs.
`) + "\n"
	return createSeedDocument(ctx, cfg, documentHistoryDiffCurrentPath, "Current Diff Evidence", currentBody)
}
func seedDocumentHistoryRestore(ctx context.Context, cfg runclient.Config) error {
	sourceBody := strings.TrimSpace(`---
type: source
status: active
---
# Restore Authority

## Summary
Authoritative restore guidance says the accepted lifecycle policy is runner-visible review before accepting source-sensitive durable edits.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryRestoreSourcePath, "Restore Authority", sourceBody); err != nil {
		return err
	}
	targetBody := strings.TrimSpace(`---
type: policy
status: active
source_refs: sources/history-review/restore-authority.md
---
# Restore Target

## Summary
Unsafe accepted edit: source-sensitive durable edits may bypass review and become accepted knowledge immediately.

## Sources
- sources/history-review/restore-authority.md

## Freshness
Checked before restore pressure.
`) + "\n"
	return createSeedDocument(ctx, cfg, documentHistoryRestoreTargetPath, "Restore Target", targetBody)
}
func seedDocumentHistoryPendingReview(ctx context.Context, cfg runclient.Config) error {
	targetBody := strings.TrimSpace(`---
type: policy
status: active
---
# Pending Target

## Summary
Accepted lifecycle policy: source-sensitive durable edits require human review before acceptance.
`) + "\n"
	return createSeedDocument(ctx, cfg, documentHistoryPendingTargetPath, "Pending Target", targetBody)
}
func seedDocumentHistoryStaleSynthesis(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/history-review/stale-current.md
---
# Stale Old Source

## Summary
Older history review guidance said semantic history controls should be promoted immediately.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryStaleOldSourcePath, "Stale Old Source", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
type: source
status: active
supersedes: sources/history-review/stale-old.md
---
# Stale Current Source

## Summary
Initial current guidance says existing document and retrieval workflows should be tested before promoting semantic history controls.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryStaleCurrentSourcePath, "Stale Current Source", currentBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/history-review/stale-current.md, sources/history-review/stale-old.md
---
# History Review Stale Synthesis

## Summary
Stale synthesis claim: semantic history controls should be promoted immediately.

## Sources
- sources/history-review/stale-current.md
- sources/history-review/stale-old.md

## Freshness
Checked before the latest current source revision.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentHistoryStaleSynthesisPath, "History Review Stale Synthesis", synthesisBody); err != nil {
		return err
	}
	return replaceScenarioSeedSection(ctx, cfg, documentHistoryStaleCurrentSourcePath, "Summary", "Current history review guidance says existing document and retrieval workflows should be tested before promoting semantic history controls, and sources/history-review/stale-old.md is superseded.")
}
