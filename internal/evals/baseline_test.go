package evals_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestUnifiedOpenClerkRunnerGate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DataDir: filepath.Join(t.TempDir(), "data")}
	source := createDocument(t, ctx, config, "notes/architecture/knowledge-plane.md", "Knowledge plane", "---\ntype: note\nstatus: active\n---\n# Knowledge plane\n\n## Summary\nCanonical architecture note.\n")
	target := createDocument(t, ctx, config, "notes/projects/openclerk-roadmap.md", "Roadmap", "---\ntype: project\nstatus: active\n---\n# Roadmap\n\n## Summary\nSee the [knowledge plane](../architecture/knowledge-plane.md).\n")
	createDocument(t, ctx, config, "records/assets/transmission-solenoid.md", "Transmission solenoid", "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\ntype: record\nstatus: active\n---\n# Transmission solenoid\n\n## Facts\n- sku: SOL-1\n")
	createDocument(t, ctx, config, "records/services/openclerk-runner.md", "OpenClerk runner", "---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service.\n")
	createDocument(t, ctx, config, "docs/architecture/runner-decision.md", "Runner decision", "---\ndecision_id: adr-runner\ndecision_title: Use JSON runner\ndecision_status: accepted\ndecision_scope: agentops\ndecision_owner: platform\n---\n# Runner decision\n\n## Summary\nUse the JSON runner for routine AgentOps knowledge tasks.\n")

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: "roadmap", Limit: 10},
	})
	if err != nil {
		t.Fatalf("search task: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || search.Search.Hits[0].DocID != target.DocID {
		t.Fatalf("search result = %+v", search.Search)
	}

	links, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  target.DocID,
	})
	if err != nil {
		t.Fatalf("links task: %v", err)
	}
	if links.Links == nil || len(links.Links.Outgoing) != 1 || links.Links.Outgoing[0].DocID != source.DocID {
		t.Fatalf("links result = %+v", links.Links)
	}

	records, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:  runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{Text: "solenoid", Limit: 10},
	})
	if err != nil {
		t.Fatalf("records task: %v", err)
	}
	if records.Records == nil || len(records.Records.Entities) != 1 {
		t.Fatalf("records result = %+v", records.Records)
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
	if services.Services == nil || len(services.Services.Services) != 1 {
		t.Fatalf("services result = %+v", services.Services)
	}

	decisions, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:    runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{Text: "JSON runner", Status: "accepted", Scope: "agentops", Limit: 10},
	})
	if err != nil {
		t.Fatalf("decisions task: %v", err)
	}
	if decisions.Decisions == nil || len(decisions.Decisions.Decisions) != 1 {
		t.Fatalf("decisions result = %+v", decisions.Decisions)
	}

	events, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   target.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("events task: %v", err)
	}
	if events.Provenance == nil || len(events.Provenance.Events) == 0 {
		t.Fatalf("events result = %+v", events.Provenance)
	}
}

func createDocument(t *testing.T, ctx context.Context, config runclient.Config, path string, title string, body string) runner.Document {
	t.Helper()
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  path,
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		t.Fatalf("create document %s: %v", path, err)
	}
	if result.Document == nil {
		t.Fatalf("create document %s result = %+v", path, result)
	}
	return *result.Document
}
