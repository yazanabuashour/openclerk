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
