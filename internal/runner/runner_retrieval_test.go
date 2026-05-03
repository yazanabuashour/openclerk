package runner_test

import (
	"context"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRetrievalTaskSynthesisFreshnessProjectionAndProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nInitial source guidance.\n")
	synthesis := createDocument(t, ctx, config, "synthesis/runner.md", "Runner Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner.md
---
# Runner Synthesis

## Summary
Initial source guidance.

## Sources
- sources/runner.md

## Freshness
Checked source refs.
`)+"\n")

	time.Sleep(time.Millisecond)
	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   source.DocID,
		Heading: "Summary",
		Content: "Updated source guidance.",
	})
	if err != nil {
		t.Fatalf("update source: %v", err)
	}
	if updated.Document == nil {
		t.Fatalf("update result = %+v", updated)
	}

	projections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      synthesis.DocID,
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("synthesis projection task: %v", err)
	}
	if projections.Projections == nil ||
		len(projections.Projections.Projections) != 1 ||
		projections.Projections.Projections[0].Freshness != "stale" ||
		projections.Projections.Projections[0].Details["stale_source_refs"] != "sources/runner.md" {
		t.Fatalf("synthesis projections result = %+v", projections)
	}

	sourceEvents, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "source",
			RefID:   source.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("source provenance task: %v", err)
	}
	if sourceEvents.Provenance == nil ||
		!runnerEventTypesInclude(sourceEvents.Provenance.Events, "source_created") ||
		!runnerEventTypesInclude(sourceEvents.Provenance.Events, "source_updated") {
		t.Fatalf("source provenance result = %+v", sourceEvents)
	}

	synthesisEvents, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "synthesis:" + synthesis.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("synthesis provenance task: %v", err)
	}
	if synthesisEvents.Provenance == nil || !runnerEventTypesInclude(synthesisEvents.Provenance.Events, "projection_invalidated") {
		t.Fatalf("synthesis provenance result = %+v", synthesisEvents)
	}
}

func TestRetrievalTaskAuditContradictionsPlansAndRepairsExistingSynthesis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	oldSource := createDocument(t, ctx, config, "sources/audit-runner-old.md", "Old audit runner source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-runner-current.md
---
# Old audit runner source

## Summary
Older source-sensitive audit runner repair evidence said agents should prefer a legacy command-path workaround.
`)+"\n")
	currentSource := createDocument(t, ctx, config, "sources/audit-runner-current.md", "Current audit runner source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-runner-old.md
---
# Current audit runner source

## Summary
Current source-sensitive audit runner repair evidence says use the installed openclerk JSON runner.
`)+"\n")
	synthesis := createDocument(t, ctx, config, "synthesis/audit-runner-routing.md", "Audit runner routing", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/audit-runner-current.md, sources/audit-runner-old.md
---
# Audit runner routing

## Summary
Stale audit claim: agents should prefer a legacy command-path workaround.

## Sources
- sources/audit-runner-current.md
- sources/audit-runner-old.md

## Freshness
Checked source refs.
`)+"\n")
	createDocument(t, ctx, config, "synthesis/audit-runner-routing-decoy.md", "Audit runner routing decoy", "# Audit runner routing decoy\n\n## Summary\nDecoy.\n")
	createDocument(t, ctx, config, "sources/audit-conflict-alpha.md", "Audit conflict alpha", strings.TrimSpace(`---
type: source
status: active
---
# Audit conflict alpha

## Summary
Source sensitive audit conflict runner retention is seven days.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-conflict-bravo.md", "Audit conflict bravo", strings.TrimSpace(`---
type: source
status: active
---
# Audit conflict bravo

## Summary
Source sensitive audit conflict runner retention is thirty days.
`)+"\n")

	time.Sleep(time.Millisecond)
	_, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   currentSource.DocID,
		Heading: "Summary",
		Content: "Current source-sensitive audit runner repair evidence says use the installed openclerk JSON runner for audit repairs.",
	})
	if err != nil {
		t.Fatalf("update current source: %v", err)
	}

	plan, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit runner repair evidence",
			TargetPath:    "synthesis/audit-runner-routing.md",
			Mode:          "plan_only",
			ConflictQuery: "source sensitive audit conflict runner retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit plan: %v", err)
	}
	if plan.Audit == nil ||
		plan.Audit.RepairStatus != "planned" ||
		plan.Audit.RepairApplied ||
		plan.Audit.SelectedTargetPath != "synthesis/audit-runner-routing.md" ||
		!containsString(plan.Audit.CandidateSynthesisPaths, "synthesis/audit-runner-routing-decoy.md") ||
		!containsString(plan.Audit.CurrentSourcePaths, currentSource.Path) ||
		!containsString(plan.Audit.SupersededSourcePaths, oldSource.Path) ||
		len(plan.Audit.ProjectionFreshnessBefore) == 0 ||
		len(plan.Audit.ProjectionFreshnessAfter) == 0 ||
		len(plan.Audit.UnresolvedConflictGroups) != 1 ||
		plan.Audit.UnresolvedConflictGroups[0].Status != "unresolved" {
		t.Fatalf("audit plan result = %+v", plan.Audit)
	}
	if !auditInspectedPath(plan.Audit.ProvenanceInspected, "sources/audit-conflict-alpha.md") ||
		!auditInspectedPath(plan.Audit.ProvenanceInspected, "sources/audit-conflict-bravo.md") {
		t.Fatalf("audit plan did not inspect conflict provenance: %+v", plan.Audit.ProvenanceInspected)
	}
	unchanged, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get unchanged synthesis: %v", err)
	}
	if !strings.Contains(unchanged.Document.Body, "legacy command-path workaround") {
		t.Fatalf("plan_only changed synthesis body = %q", unchanged.Document.Body)
	}

	sourceAudit, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSourceAuditReport,
		SourceAudit: runner.SourceAuditReportOptions{
			Query:         "source-sensitive audit runner repair evidence",
			TargetPath:    "synthesis/audit-runner-routing.md",
			Mode:          "explain",
			ConflictQuery: "source sensitive audit conflict runner retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("source audit explain: %v", err)
	}
	if sourceAudit.SourceAudit == nil ||
		sourceAudit.SourceAudit.Mode != "explain" ||
		sourceAudit.SourceAudit.RepairStatus != "planned" ||
		sourceAudit.SourceAudit.RepairApplied ||
		sourceAudit.SourceAudit.SelectedTargetPath != "synthesis/audit-runner-routing.md" ||
		sourceAudit.SourceAudit.AgentHandoff == nil ||
		!strings.Contains(sourceAudit.SourceAudit.AgentHandoff.AnswerSummary, "source_audit_report explain") ||
		!strings.Contains(sourceAudit.SourceAudit.AgentHandoff.FollowUpPrimitiveInspection, "not required") ||
		!strings.Contains(sourceAudit.SourceAudit.ValidationBoundaries, "explain mode is read-only") ||
		!strings.Contains(sourceAudit.SourceAudit.AuthorityLimits, "unresolved current-source conflicts") {
		t.Fatalf("source audit explain result = %+v", sourceAudit.SourceAudit)
	}
	stillUnchanged, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get source audit unchanged synthesis: %v", err)
	}
	if !strings.Contains(stillUnchanged.Document.Body, "legacy command-path workaround") {
		t.Fatalf("source_audit explain changed synthesis body = %q", stillUnchanged.Document.Body)
	}

	repaired, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit runner repair evidence",
			TargetPath:    "synthesis/audit-runner-routing.md",
			Mode:          "repair_existing",
			ConflictQuery: "source sensitive audit conflict runner retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit repair: %v", err)
	}
	if repaired.Audit == nil ||
		repaired.Audit.RepairStatus != "applied" ||
		!repaired.Audit.RepairApplied ||
		repaired.Audit.DuplicatePrevention != "existing_target_selected_no_duplicate_created" ||
		repaired.Audit.FailureClassification != "none" {
		t.Fatalf("audit repair result = %+v", repaired.Audit)
	}
	if len(repaired.Audit.ProjectionFreshnessAfter) == 0 || repaired.Audit.ProjectionFreshnessAfter[0].Freshness != "fresh" {
		t.Fatalf("projection after repair = %+v", repaired.Audit.ProjectionFreshnessAfter)
	}
	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get repaired synthesis: %v", err)
	}
	for _, want := range []string{
		"source_refs: sources/audit-runner-current.md, sources/audit-runner-old.md",
		"Current audit guidance: use the installed openclerk JSON runner.",
		"Current source: sources/audit-runner-current.md.",
		"Superseded source: sources/audit-runner-old.md.",
		"## Sources",
		"## Freshness",
	} {
		if !strings.Contains(updated.Document.Body, want) {
			t.Fatalf("repaired body missing %q:\n%s", want, updated.Document.Body)
		}
	}
}

func TestRetrievalTaskSourceAuditReportRepairsOnlyExistingSynthesis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/source-audit-old.md", "Old source audit source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/source-audit-current.md
---
# Old source audit source

## Summary
Old source-sensitive audit workflow says use direct vault inspection.
`)+"\n")
	createDocument(t, ctx, config, "sources/source-audit-current.md", "Current source audit source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/source-audit-old.md
---
# Current source audit source

## Summary
Current source-sensitive audit workflow says use source_audit_report.
`)+"\n")
	synthesis := createDocument(t, ctx, config, "synthesis/source-audit.md", "Source audit synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/source-audit-current.md, sources/source-audit-old.md
---
# Source audit synthesis

## Summary
Stale audit workflow says use direct vault inspection.

## Sources
- sources/source-audit-current.md
- sources/source-audit-old.md

## Freshness
Checked source refs.
`)+"\n")

	repaired, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSourceAuditReport,
		SourceAudit: runner.SourceAuditReportOptions{
			Query:      "source-sensitive audit workflow",
			TargetPath: "synthesis/source-audit.md",
			Mode:       "repair_existing",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("source audit repair: %v", err)
	}
	if repaired.SourceAudit == nil ||
		repaired.SourceAudit.Mode != "repair_existing" ||
		!repaired.SourceAudit.RepairApplied ||
		repaired.SourceAudit.RepairStatus != "applied" ||
		repaired.SourceAudit.DuplicatePrevention != "existing_target_selected_no_duplicate_created" {
		t.Fatalf("source audit repair result = %+v", repaired.SourceAudit)
	}

	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  synthesis.DocID,
	})
	if err != nil {
		t.Fatalf("get repaired source audit synthesis: %v", err)
	}
	if strings.Contains(updated.Document.Body, "direct vault inspection") ||
		!strings.Contains(updated.Document.Body, "installed openclerk JSON runner") {
		t.Fatalf("source audit repaired body = %q", updated.Document.Body)
	}

	missing, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSourceAuditReport,
		SourceAudit: runner.SourceAuditReportOptions{
			Query:      "source-sensitive audit workflow",
			TargetPath: "synthesis/missing-source-audit.md",
			Mode:       "repair_existing",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("source audit missing target: %v", err)
	}
	if missing.SourceAudit == nil ||
		missing.SourceAudit.RepairApplied ||
		missing.SourceAudit.FailureClassification != "target_not_found" {
		t.Fatalf("source audit missing target result = %+v", missing.SourceAudit)
	}
}

func TestRetrievalTaskAuditContradictionsFindsTargetAfterFirstSynthesisPage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/audit-page-old.md", "Old audit page source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-page-current.md
---
# Old audit page source

## Summary
Older source-sensitive audit page evidence said to use a legacy command path.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-page-current.md", "Current audit page source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-page-old.md
---
# Current audit page source

## Summary
Current source-sensitive audit page evidence says use the installed openclerk JSON runner.
`)+"\n")
	for i := 0; i < 101; i++ {
		createDocument(t, ctx, config, fmt.Sprintf("synthesis/aa-audit-decoy-%03d.md", i), fmt.Sprintf("Audit decoy %03d", i), fmt.Sprintf("# Audit decoy %03d\n\n## Summary\nDecoy.\n", i))
	}
	target := createDocument(t, ctx, config, "synthesis/zz-audit-page-target.md", "Audit page target", strings.TrimSpace(`---
type: synthesis
status: active
source_refs: sources/audit-page-current.md, sources/audit-page-old.md
---
# Audit page target

## Summary
Stale audit page claim.

## Sources
- sources/audit-page-current.md
- sources/audit-page-old.md

## Freshness
Checked source refs.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:      "source-sensitive audit page evidence",
			TargetPath: target.Path,
			Mode:       "plan_only",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("audit paginated target: %v", err)
	}
	if result.Audit == nil ||
		result.Audit.SelectedTargetPath != target.Path ||
		result.Audit.RepairStatus != "planned" ||
		result.Audit.DuplicatePrevention != "existing_target_selected_no_duplicate_created" {
		t.Fatalf("audit paginated target result = %+v", result.Audit)
	}
}

func TestRetrievalTaskAuditContradictionsDoesNotReportMatchingCurrentSourcesAsConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/audit-agree-old.md", "Old audit agree source", strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/audit-agree-current.md
---
# Old audit agree source

## Summary
Older source-sensitive audit agreement evidence said to use a legacy command path.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-agree-current.md", "Current audit agree source", strings.TrimSpace(`---
type: source
status: active
supersedes: sources/audit-agree-old.md
---
# Current audit agree source

## Summary
Current source-sensitive audit agreement evidence says use the installed openclerk JSON runner.
`)+"\n")
	createDocument(t, ctx, config, "synthesis/audit-agree-target.md", "Audit agree target", strings.TrimSpace(`---
type: synthesis
status: active
source_refs: sources/audit-agree-current.md, sources/audit-agree-old.md
---
# Audit agree target

## Summary
Stale audit agreement claim.

## Sources
- sources/audit-agree-current.md
- sources/audit-agree-old.md

## Freshness
Checked source refs.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-retention-alpha.md", "Audit retention alpha", strings.TrimSpace(`---
type: source
status: active
---
# Audit retention alpha

## Summary
Source sensitive audit matching retention is seven days.
`)+"\n")
	createDocument(t, ctx, config, "sources/audit-retention-bravo.md", "Audit retention bravo", strings.TrimSpace(`---
type: source
status: active
---
# Audit retention bravo

## Summary
Source sensitive audit matching retention is seven days.
`)+"\n")

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:         "source-sensitive audit agreement evidence",
			TargetPath:    "synthesis/audit-agree-target.md",
			Mode:          "plan_only",
			ConflictQuery: "source sensitive audit matching retention",
			Limit:         10,
		},
	})
	if err != nil {
		t.Fatalf("audit matching conflict: %v", err)
	}
	if result.Audit == nil || len(result.Audit.UnresolvedConflictGroups) != 0 {
		t.Fatalf("matching current sources reported as conflict = %+v", result.Audit)
	}
	if !auditInspectedPath(result.Audit.ProvenanceInspected, "sources/audit-retention-alpha.md") ||
		!auditInspectedPath(result.Audit.ProvenanceInspected, "sources/audit-retention-bravo.md") {
		t.Fatalf("audit matching conflict did not inspect provenance: %+v", result.Audit.ProvenanceInspected)
	}
}

func TestRetrievalTaskAuditContradictionsValidation(t *testing.T) {
	t.Parallel()

	missing, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			TargetPath: "synthesis/audit-runner-routing.md",
		},
	})
	if err != nil {
		t.Fatalf("missing audit query validation: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "audit.query is required" {
		t.Fatalf("missing result = %+v", missing)
	}

	invalidMode, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionAuditContradictions,
		Audit: runner.AuditContradictionsOptions{
			Query:      "source-sensitive audit runner repair evidence",
			TargetPath: "synthesis/audit-runner-routing.md",
			Mode:       "create_new",
		},
	})
	if err != nil {
		t.Fatalf("invalid audit mode validation: %v", err)
	}
	if !invalidMode.Rejected || invalidMode.RejectionReason != "audit.mode must be plan_only or repair_existing" {
		t.Fatalf("invalid mode result = %+v", invalidMode)
	}
}

func TestRetrievalTaskSearchLinksRecordsAndProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	architecture := createDocument(t, ctx, config, "notes/architecture/knowledge-plane.md", "Knowledge plane", "# Knowledge plane\n\n## Summary\nCanonical architecture note.\n")
	roadmap := createDocument(t, ctx, config, "notes/projects/roadmap.md", "Roadmap", "# Roadmap\n\n## Summary\nSee the [knowledge plane](../architecture/knowledge-plane.md).\n")
	createDocument(t, ctx, config, "records/assets/transmission-solenoid.md", "Transmission solenoid", "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\n---\n# Transmission solenoid\n\n## Facts\n- sku: SOL-1\n")
	createDocument(t, ctx, config, "records/services/openclerk-runner.md", "OpenClerk runner", "---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service for routine knowledge tasks.\n\n## Facts\n- tier: production\n")
	createDocument(t, ctx, config, "docs/architecture/runner-old-decision.md", "Old runner decision", "---\ndecision_id: adr-runner-old\ndecision_title: Old runner path\ndecision_status: superseded\ndecision_scope: agentops\ndecision_owner: platform\ndecision_date: 2026-04-20\nsuperseded_by: adr-runner-current\nsource_refs: sources/runner-old.md\n---\n# Old runner path\n\n## Summary\nOld decision used a retired runner path.\n")
	createDocument(t, ctx, config, "sources/runner-old.md", "Old runner source", "# Old runner source\n\n## Summary\nRetired runner path source.\n")
	createDocument(t, ctx, config, "notes/architecture/runner-current-decision.md", "Current runner decision", "---\ndecision_id: adr-runner-current\ndecision_title: Use JSON runner\ndecision_status: accepted\ndecision_scope: agentops\ndecision_owner: platform\ndecision_date: 2026-04-22\nsupersedes: adr-runner-old\nsource_refs: sources/runner-current.md\n---\n# Use JSON runner\n\n## Summary\nAccepted decision uses the JSON runner for routine AgentOps work.\n")
	createDocument(t, ctx, config, "sources/runner-current.md", "Current runner source", "# Current runner source\n\n## Summary\nCurrent runner source.\n")

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  "roadmap",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("search task: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || len(search.Search.Hits[0].Citations) == 0 {
		t.Fatalf("search result = %+v", search)
	}

	links, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  roadmap.DocID,
	})
	if err != nil {
		t.Fatalf("links task: %v", err)
	}
	if links.Links == nil || len(links.Links.Outgoing) != 1 || links.Links.Outgoing[0].DocID != architecture.DocID {
		t.Fatalf("links result = %+v", links)
	}

	graph, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraph,
		DocID:  roadmap.DocID,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("graph task: %v", err)
	}
	if graph.Graph == nil || len(graph.Graph.Nodes) == 0 || len(graph.Graph.Edges) == 0 {
		t.Fatalf("graph result = %+v", graph)
	}

	records, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "solenoid", Limit: 10},
	})
	if err != nil {
		t.Fatalf("records task: %v", err)
	}
	if records.Records == nil || len(records.Records.Entities) != 1 {
		t.Fatalf("records result = %+v", records)
	}

	entity, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionRecordEntity,
		EntityID: records.Records.Entities[0].EntityID,
	})
	if err != nil {
		t.Fatalf("record entity task: %v", err)
	}
	if entity.Entity == nil || entity.Entity.EntityID != "transmission-solenoid" {
		t.Fatalf("entity result = %+v", entity)
	}

	services, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionServicesLookup,
		Services: runner.ServiceLookupOptions{
			Text:      "OpenClerk runner",
			Interface: "JSON runner",
			Limit:     10,
		},
	})
	if err != nil {
		t.Fatalf("services task: %v", err)
	}
	if services.Services == nil || len(services.Services.Services) != 1 || services.Services.Services[0].ServiceID != "openclerk-runner" {
		t.Fatalf("services result = %+v", services)
	}

	service, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:    runner.RetrievalTaskActionServiceRecord,
		ServiceID: "openclerk-runner",
	})
	if err != nil {
		t.Fatalf("service record task: %v", err)
	}
	if service.Service == nil ||
		service.Service.Status != "active" ||
		service.Service.Owner != "runner" ||
		service.Service.Interface != "JSON runner" ||
		len(service.Service.Citations) == 0 {
		t.Fatalf("service result = %+v", service)
	}

	decisions, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{
			Text:   "JSON runner",
			Status: "accepted",
			Scope:  "agentops",
			Owner:  "platform",
			Limit:  10,
		},
	})
	if err != nil {
		t.Fatalf("decisions task: %v", err)
	}
	if decisions.Decisions == nil ||
		len(decisions.Decisions.Decisions) != 1 ||
		decisions.Decisions.Decisions[0].DecisionID != "adr-runner-current" ||
		len(decisions.Decisions.Decisions[0].Citations) == 0 {
		t.Fatalf("decisions result = %+v", decisions)
	}

	decision, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: "adr-runner-old",
	})
	if err != nil {
		t.Fatalf("decision record task: %v", err)
	}
	if decision.Decision == nil ||
		decision.Decision.Status != "superseded" ||
		len(decision.Decision.SupersededBy) != 1 ||
		decision.Decision.SupersededBy[0] != "adr-runner-current" ||
		len(decision.Decision.Citations) == 0 {
		t.Fatalf("decision result = %+v", decision)
	}

	provenance, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   roadmap.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("provenance task: %v", err)
	}
	if provenance.Provenance == nil || len(provenance.Provenance.Events) == 0 {
		t.Fatalf("provenance result = %+v", provenance)
	}

	projections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "graph",
			RefKind:    "document",
			RefID:      roadmap.DocID,
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("projection task: %v", err)
	}
	if projections.Projections == nil || len(projections.Projections.Projections) != 1 {
		t.Fatalf("projection result = %+v", projections)
	}

	serviceProjections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "services",
			RefKind:    "service",
			RefID:      "openclerk-runner",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("service projection task: %v", err)
	}
	if serviceProjections.Projections == nil ||
		len(serviceProjections.Projections.Projections) != 1 ||
		serviceProjections.Projections.Projections[0].Freshness != "fresh" {
		t.Fatalf("service projections result = %+v", serviceProjections)
	}

	decisionProjections, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-runner-old",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("decision projection task: %v", err)
	}
	if decisionProjections.Projections == nil ||
		len(decisionProjections.Projections.Projections) != 1 ||
		decisionProjections.Projections.Projections[0].Freshness != "stale" ||
		decisionProjections.Projections.Projections[0].Details["superseded_by"] != "adr-runner-current" {
		t.Fatalf("decision projections result = %+v", decisionProjections)
	}

	evidence, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionEvidenceBundle,
		EvidenceBundle: runner.EvidenceBundleOptions{
			Query:      "JSON runner",
			EntityID:   "transmission-solenoid",
			DecisionID: "adr-runner-current",
			RefKind:    "document",
			RefID:      roadmap.DocID,
			Projection: "graph",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("evidence bundle task: %v", err)
	}
	if evidence.EvidenceBundle == nil ||
		evidence.EvidenceBundle.Search == nil ||
		evidence.EvidenceBundle.Records == nil ||
		evidence.EvidenceBundle.Entity == nil ||
		evidence.EvidenceBundle.Decisions == nil ||
		evidence.EvidenceBundle.Decision == nil ||
		evidence.EvidenceBundle.Provenance == nil ||
		evidence.EvidenceBundle.Projections == nil ||
		evidence.EvidenceBundle.AgentHandoff == nil ||
		!strings.Contains(evidence.EvidenceBundle.AgentHandoff.AnswerSummary, "evidence_bundle_report returned") ||
		!strings.Contains(evidence.EvidenceBundle.AgentHandoff.FollowUpPrimitiveInspection, "not required") ||
		len(evidence.EvidenceBundle.Citations) == 0 ||
		!strings.Contains(evidence.EvidenceBundle.ValidationBoundaries, "read-only") ||
		!strings.Contains(evidence.EvidenceBundle.AuthorityLimits, "does not create a new authority source") {
		t.Fatalf("evidence bundle result = %+v", evidence.EvidenceBundle)
	}

	entityOnlyEvidence, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionEvidenceBundle,
		EvidenceBundle: runner.EvidenceBundleOptions{
			EntityID: "transmission-solenoid",
			Limit:    10,
		},
	})
	if err != nil {
		t.Fatalf("entity-only evidence bundle task: %v", err)
	}
	if entityOnlyEvidence.EvidenceBundle == nil ||
		entityOnlyEvidence.EvidenceBundle.Records != nil ||
		entityOnlyEvidence.EvidenceBundle.Entity == nil ||
		entityOnlyEvidence.EvidenceBundle.AgentHandoff == nil ||
		!strings.Contains(entityOnlyEvidence.EvidenceBundle.AgentHandoff.AnswerSummary, "1 records") ||
		!strings.Contains(strings.Join(entityOnlyEvidence.EvidenceBundle.AgentHandoff.Evidence, "\n"), "records=1") ||
		!strings.Contains(strings.Join(entityOnlyEvidence.EvidenceBundle.AgentHandoff.Evidence, "\n"), "decisions=0") {
		t.Fatalf("entity-only evidence bundle handoff = %+v", entityOnlyEvidence.EvidenceBundle)
	}

	decisionOnlyEvidence, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionEvidenceBundle,
		EvidenceBundle: runner.EvidenceBundleOptions{
			DecisionID: "adr-runner-current",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("decision-only evidence bundle task: %v", err)
	}
	if decisionOnlyEvidence.EvidenceBundle == nil ||
		decisionOnlyEvidence.EvidenceBundle.Decisions != nil ||
		decisionOnlyEvidence.EvidenceBundle.Decision == nil ||
		decisionOnlyEvidence.EvidenceBundle.AgentHandoff == nil ||
		!strings.Contains(decisionOnlyEvidence.EvidenceBundle.AgentHandoff.AnswerSummary, "1 decisions") ||
		!strings.Contains(strings.Join(decisionOnlyEvidence.EvidenceBundle.AgentHandoff.Evidence, "\n"), "records=0") ||
		!strings.Contains(strings.Join(decisionOnlyEvidence.EvidenceBundle.AgentHandoff.Evidence, "\n"), "decisions=1") {
		t.Fatalf("decision-only evidence bundle handoff = %+v", decisionOnlyEvidence.EvidenceBundle)
	}

	afterEvidenceList, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 100},
	})
	if err != nil {
		t.Fatalf("list after evidence bundle: %v", err)
	}
	if len(afterEvidenceList.Documents) != 8 {
		t.Fatalf("evidence bundle changed document count: %+v", afterEvidenceList.Documents)
	}
}

func TestRetrievalTaskMemoryRouterRecallReport(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	seedMemoryRouterRecallDocs(t, ctx, config)

	before, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 50},
	})
	if err != nil {
		t.Fatalf("list before: %v", err)
	}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Query: "memory router temporal recall session promotion feedback weighting routing canonical docs",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("memory/router recall report: %v", err)
	}
	if result.Rejected || result.MemoryRouterRecall == nil {
		t.Fatalf("memory/router recall result = %+v", result)
	}
	report := result.MemoryRouterRecall
	for _, value := range []string{
		report.QuerySummary,
		report.TemporalStatus,
		report.StaleSessionStatus,
		report.FeedbackWeighting,
		report.RoutingRationale,
		report.SynthesisFreshness,
		report.ValidationBoundaries,
		report.AuthorityLimits,
	} {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("empty report field in %+v", report)
		}
	}
	for _, want := range []string{
		"notes/memory-router/session-observation.md",
		"notes/memory-router/temporal-policy.md",
		"notes/memory-router/feedback-weighting.md",
		"notes/memory-router/routing-policy.md",
		"synthesis/memory-router-reference.md",
	} {
		if !containsString(report.CanonicalEvidenceRefs, want) {
			t.Fatalf("canonical refs %v missing %q", report.CanonicalEvidenceRefs, want)
		}
	}
	if len(report.ProvenanceRefs) == 0 || !strings.HasPrefix(report.ProvenanceRefs[0], "document:") {
		t.Fatalf("provenance refs = %+v", report.ProvenanceRefs)
	}
	if !strings.Contains(report.SynthesisFreshness, "fresh synthesis projection") {
		t.Fatalf("synthesis freshness = %q", report.SynthesisFreshness)
	}
	if strings.Contains(report.ValidationBoundaries, "missing evidence") {
		t.Fatalf("unexpected missing evidence in validation boundaries: %q", report.ValidationBoundaries)
	}

	after, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 50},
	})
	if err != nil {
		t.Fatalf("list after: %v", err)
	}
	if len(after.Documents) != len(before.Documents) {
		t.Fatalf("memory/router recall report mutated document count: before=%d after=%d", len(before.Documents), len(after.Documents))
	}
}

func TestRetrievalTaskMemoryRouterRecallReportPagesUntilFixedEvidence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	seedMemoryRouterRecallDocs(t, ctx, config)
	for i := range 12 {
		createDocument(
			t,
			ctx,
			config,
			fmt.Sprintf("notes/memory-router/000-extra-%02d.md", i),
			fmt.Sprintf("Memory Router Extra %02d", i),
			fmt.Sprintf("# Memory Router Extra %02d\n\n## Summary\nDistractor memory/router note.\n", i),
		)
	}
	for i := range 24 {
		createDocument(
			t,
			ctx,
			config,
			fmt.Sprintf("synthesis/000-extra-%02d.md", i),
			fmt.Sprintf("Synthesis Extra %02d", i),
			fmt.Sprintf("# Synthesis Extra %02d\n\n## Summary\nDistractor synthesis.\n", i),
		)
	}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Query: "memory router temporal recall session promotion feedback weighting routing canonical docs",
			Limit: 4,
		},
	})
	if err != nil {
		t.Fatalf("memory/router recall report with paged evidence: %v", err)
	}
	if result.Rejected || result.MemoryRouterRecall == nil {
		t.Fatalf("paged evidence result = %+v", result)
	}
	report := result.MemoryRouterRecall
	for _, want := range []string{
		"notes/memory-router/session-observation.md",
		"notes/memory-router/temporal-policy.md",
		"notes/memory-router/feedback-weighting.md",
		"notes/memory-router/routing-policy.md",
		"synthesis/memory-router-reference.md",
	} {
		if !containsString(report.CanonicalEvidenceRefs, want) {
			t.Fatalf("canonical refs %v missing %q", report.CanonicalEvidenceRefs, want)
		}
	}
	if strings.Contains(report.ValidationBoundaries, "missing evidence") {
		t.Fatalf("paged lookup reported missing evidence: %q", report.ValidationBoundaries)
	}
}

func TestRetrievalTaskMemoryRouterRecallReportReportsMissingEvidence(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}

	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("memory/router recall report with missing evidence: %v", err)
	}
	if result.Rejected || result.MemoryRouterRecall == nil {
		t.Fatalf("missing evidence result = %+v", result)
	}
	report := result.MemoryRouterRecall
	if !containsString(report.CanonicalEvidenceRefs, "missing:notes/memory-router/session-observation.md") {
		t.Fatalf("missing evidence refs = %+v", report.CanonicalEvidenceRefs)
	}
	if !strings.Contains(report.ValidationBoundaries, "missing evidence") {
		t.Fatalf("validation boundaries = %q", report.ValidationBoundaries)
	}
}

func TestRetrievalTaskMemoryRouterRecallReportRejectsNegativeLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Limit: -3,
		},
	})
	if err != nil {
		t.Fatalf("negative memory/router recall limit: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative limit result = %+v", result)
	}
}

func seedMemoryRouterRecallDocs(t *testing.T, ctx context.Context, config runclient.Config) {
	t.Helper()
	createDocument(t, ctx, config, "notes/memory-router/session-observation.md", "Memory Router Session Observation", strings.TrimSpace(`---
type: source
status: active
observed_at: 2026-04-22
---
# Memory Router Session Observation

## Summary
Session observation: a user asked whether memory routing should promote recall. Useful session material must be promoted only by writing canonical markdown with source refs.

## Feedback
Positive feedback weight 0.8 is advisory only and cannot hide stale canonical evidence.
`)+"\n")
	createDocument(t, ctx, config, "notes/memory-router/temporal-policy.md", "Temporal Recall Policy", "# Temporal Recall Policy\n\n## Summary\nCurrent canonical docs outrank stale session observations for memory/router recall.\n")
	createDocument(t, ctx, config, "notes/memory-router/feedback-weighting.md", "Feedback Weighting", "# Feedback Weighting\n\n## Summary\nFeedback weighting is advisory only and cannot hide stale canonical evidence.\n")
	createDocument(t, ctx, config, "notes/memory-router/routing-policy.md", "Routing Policy", "# Routing Policy\n\n## Summary\nRouting rationale uses existing AgentOps document and retrieval actions; no autonomous router API is authority.\n")
	createDocument(t, ctx, config, "synthesis/memory-router-reference.md", "Memory Router Reference", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, notes/memory-router/routing-policy.md
---
# Memory Router Reference

## Summary
Temporal status: current canonical docs outrank stale session observations.
Session promotion path: durable canonical markdown with source refs.
Feedback weighting: advisory only.
Routing choice: existing AgentOps document and retrieval actions.
Decision: implement read-only memory/router recall report.

## Sources
- notes/memory-router/session-observation.md
- notes/memory-router/temporal-policy.md
- notes/memory-router/feedback-weighting.md
- notes/memory-router/routing-policy.md

## Freshness
Fresh synthesis projection expected for current source refs.
`)+"\n")
}

func TestRetrievalTaskTypedRecordValidation(t *testing.T) {
	t.Parallel()

	missing, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionServiceRecord,
	})
	if err != nil {
		t.Fatalf("missing service id validation: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "service_id is required" {
		t.Fatalf("missing result = %+v", missing)
	}

	negative, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionServicesLookup,
		Services: runner.ServiceLookupOptions{Limit: -1},
	})
	if err != nil {
		t.Fatalf("negative service limit validation: %v", err)
	}
	if !negative.Rejected || negative.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative result = %+v", negative)
	}

	missingDecision, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionRecord,
	})
	if err != nil {
		t.Fatalf("missing decision id validation: %v", err)
	}
	if !missingDecision.Rejected || missingDecision.RejectionReason != "decision_id is required" {
		t.Fatalf("missing decision result = %+v", missingDecision)
	}

	negativeDecision, err := runner.RunRetrievalTask(context.Background(), runclient.Config{}, runner.RetrievalTaskRequest{
		Action:    runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{Limit: -1},
	})
	if err != nil {
		t.Fatalf("negative decision limit validation: %v", err)
	}
	if !negativeDecision.Rejected || negativeDecision.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative decision result = %+v", negativeDecision)
	}
}

func TestRetrievalTaskSearchTagFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/tagging/account-renewal.md", "Account Renewal", strings.TrimSpace(`---
tag: account-renewal
---
# Account Renewal

## Summary
Renewal tag search evidence belongs to account renewal.
`)+"\n")
	createDocument(t, ctx, config, "notes/tagging/security-renewal.md", "Security Renewal", strings.TrimSpace(`---
tag: renewal
---
# Security Renewal

## Summary
Renewal tag search evidence belongs to security renewal.
`)+"\n")
	createDocument(t, ctx, config, "notes/tagging/ops-review.md", "Ops Review", strings.TrimSpace(`---
tag: ops-review
---
# Ops Review

## Summary
Near duplicate tag evidence belongs to singular ops review.
`)+"\n")
	createDocument(t, ctx, config, "notes/tagging/ops-reviews.md", "Ops Reviews", strings.TrimSpace(`---
tag: ops-reviews
---
# Ops Reviews

## Summary
Near duplicate tag evidence belongs to plural ops reviews.
`)+"\n")
	createDocument(t, ctx, config, "archive/tagging/account-renewal.md", "Archived Account Renewal", strings.TrimSpace(`---
tag: account-renewal
---
# Archived Account Renewal

## Summary
Archived renewal tag search evidence must be excluded by path prefix.
`)+"\n")

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  "Renewal tag search evidence",
			Tag:   " account-renewal ",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("search tag: %v", err)
	}
	if search.Search == nil || !searchResultContainsPath(search.Search.Hits, "notes/tagging/account-renewal.md") {
		t.Fatalf("search tag result = %+v", search.Search)
	}
	if searchResultContainsPath(search.Search.Hits, "notes/tagging/security-renewal.md") {
		t.Fatalf("search tag included wrong tag result: %+v", search.Search.Hits)
	}

	nearDuplicate, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  "Near duplicate tag evidence",
			Tag:   "ops-review",
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("near duplicate tag search: %v", err)
	}
	if nearDuplicate.Search == nil ||
		!searchResultContainsPath(nearDuplicate.Search.Hits, "notes/tagging/ops-review.md") ||
		searchResultContainsPath(nearDuplicate.Search.Hits, "notes/tagging/ops-reviews.md") {
		t.Fatalf("near duplicate tag result = %+v", nearDuplicate.Search)
	}

	scoped, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "renewal tag search evidence",
			PathPrefix: "notes/tagging/",
			Tag:        "account-renewal",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("scoped tag search: %v", err)
	}
	if scoped.Search == nil ||
		!searchResultContainsPath(scoped.Search.Hits, "notes/tagging/account-renewal.md") ||
		searchResultContainsPath(scoped.Search.Hits, "archive/tagging/account-renewal.md") {
		t.Fatalf("scoped tag result = %+v", scoped.Search)
	}

	mixed, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          "renewal",
			Tag:           "account-renewal",
			MetadataKey:   "tag",
			MetadataValue: "account-renewal",
		},
	})
	if err != nil {
		t.Fatalf("mixed tag search validation: %v", err)
	}
	if !mixed.Rejected || mixed.RejectionReason != "search.tag cannot be combined with metadata_key or metadata_value" {
		t.Fatalf("mixed tag result = %+v", mixed)
	}

	empty, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text: "renewal",
			Tag:  " ",
		},
	})
	if err != nil {
		t.Fatalf("empty tag search validation: %v", err)
	}
	if !empty.Rejected || empty.RejectionReason != "search.tag must be non-empty" {
		t.Fatalf("empty tag result = %+v", empty)
	}
}

func searchResultContainsPath(hits []runner.SearchHit, path string) bool {
	for _, hit := range hits {
		for _, citation := range hit.Citations {
			if citation.Path == path {
				return true
			}
		}
	}
	return false
}
