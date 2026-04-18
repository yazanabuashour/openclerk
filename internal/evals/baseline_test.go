package evals_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yazanabuashour/openclerk/agentops"
	"github.com/yazanabuashour/openclerk/client/local"
)

func TestUnifiedOpenClerkAgentOpsBaseline(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := local.Config{DataDir: filepath.Join(t.TempDir(), "data")}
	source := createDocument(t, ctx, config, "notes/architecture/knowledge-plane.md", "Knowledge plane", "---\ntype: note\nstatus: active\n---\n# Knowledge plane\n\n## Summary\nCanonical architecture note.\n")
	target := createDocument(t, ctx, config, "notes/projects/openclerk-roadmap.md", "Roadmap", "---\ntype: project\nstatus: active\n---\n# Roadmap\n\n## Summary\nSee the [knowledge plane](../architecture/knowledge-plane.md).\n")
	createDocument(t, ctx, config, "records/assets/transmission-solenoid.md", "Transmission solenoid", "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\ntype: record\nstatus: active\n---\n# Transmission solenoid\n\n## Facts\n- sku: SOL-1\n")

	search, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action: agentops.RetrievalTaskActionSearch,
		Search: agentops.SearchOptions{Text: "roadmap", Limit: 10},
	})
	if err != nil {
		t.Fatalf("search task: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || search.Search.Hits[0].DocID != target.DocID {
		t.Fatalf("search result = %+v", search.Search)
	}

	links, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action: agentops.RetrievalTaskActionDocumentLinks,
		DocID:  target.DocID,
	})
	if err != nil {
		t.Fatalf("links task: %v", err)
	}
	if links.Links == nil || len(links.Links.Outgoing) != 1 || links.Links.Outgoing[0].DocID != source.DocID {
		t.Fatalf("links result = %+v", links.Links)
	}

	records, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action:  agentops.RetrievalTaskActionRecordsLookup,
		Records: agentops.RecordLookupOptions{Text: "solenoid", Limit: 10},
	})
	if err != nil {
		t.Fatalf("records task: %v", err)
	}
	if records.Records == nil || len(records.Records.Entities) != 1 {
		t.Fatalf("records result = %+v", records.Records)
	}

	events, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action: agentops.RetrievalTaskActionProvenanceEvents,
		Provenance: agentops.ProvenanceEventOptions{
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

func createDocument(t *testing.T, ctx context.Context, config local.Config, path string, title string, body string) agentops.Document {
	t.Helper()
	result, err := agentops.RunDocumentTask(ctx, config, agentops.DocumentTaskRequest{
		Action: agentops.DocumentTaskActionCreate,
		Document: agentops.DocumentInput{
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
