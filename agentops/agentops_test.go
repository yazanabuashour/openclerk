package agentops_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/agentops"
	"github.com/yazanabuashour/openclerk/client/local"
)

func TestDocumentTaskCreateListGetAndUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := local.Config{DataDir: filepath.Join(t.TempDir(), "data")}
	create, err := agentops.RunDocumentTask(ctx, config, agentops.DocumentTaskRequest{
		Action: agentops.DocumentTaskActionCreate,
		Document: agentops.DocumentInput{
			Path:  "notes/projects/roadmap.md",
			Title: "Roadmap",
			Body:  "# Roadmap\n\n## Summary\nCanonical project note.\n",
		},
	})
	if err != nil {
		t.Fatalf("create document task: %v", err)
	}
	if create.Rejected || create.Document == nil || create.Document.DocID == "" {
		t.Fatalf("create result = %+v", create)
	}

	list, err := agentops.RunDocumentTask(ctx, config, agentops.DocumentTaskRequest{
		Action: agentops.DocumentTaskActionList,
		List:   agentops.DocumentListOptions{PathPrefix: "notes/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list document task: %v", err)
	}
	if len(list.Documents) != 1 || list.Documents[0].Path != "notes/projects/roadmap.md" {
		t.Fatalf("list result = %+v", list)
	}

	appendResult, err := agentops.RunDocumentTask(ctx, config, agentops.DocumentTaskRequest{
		Action:  agentops.DocumentTaskActionAppend,
		DocID:   create.Document.DocID,
		Content: "## Decisions\nUse the AgentOps runner.\n",
	})
	if err != nil {
		t.Fatalf("append document task: %v", err)
	}
	if appendResult.Document == nil || !strings.Contains(appendResult.Document.Body, "AgentOps runner") {
		t.Fatalf("append result = %+v", appendResult)
	}

	replace, err := agentops.RunDocumentTask(ctx, config, agentops.DocumentTaskRequest{
		Action:  agentops.DocumentTaskActionReplaceSection,
		DocID:   create.Document.DocID,
		Heading: "Decisions",
		Content: "Use `cmd/openclerk-agentops` for routine agent work.",
	})
	if err != nil {
		t.Fatalf("replace document task: %v", err)
	}
	if replace.Document == nil ||
		!strings.Contains(replace.Document.Body, "cmd/openclerk-agentops") ||
		strings.Contains(replace.Document.Body, "AgentOps runner") {
		t.Fatalf("replace result body = %q", replace.Document.Body)
	}

	cleared, err := agentops.RunDocumentTask(ctx, config, agentops.DocumentTaskRequest{
		Action:  agentops.DocumentTaskActionReplaceSection,
		DocID:   create.Document.DocID,
		Heading: "Decisions",
		Content: "",
	})
	if err != nil {
		t.Fatalf("clear section task: %v", err)
	}
	if cleared.Document == nil ||
		!strings.Contains(cleared.Document.Body, "## Decisions") ||
		strings.Contains(cleared.Document.Body, "cmd/openclerk-agentops") {
		t.Fatalf("cleared section body = %q", cleared.Document.Body)
	}

	get, err := agentops.RunDocumentTask(ctx, config, agentops.DocumentTaskRequest{
		Action: agentops.DocumentTaskActionGet,
		DocID:  create.Document.DocID,
	})
	if err != nil {
		t.Fatalf("get document task: %v", err)
	}
	if get.Document == nil || get.Document.Path != create.Document.Path {
		t.Fatalf("get result = %+v", get)
	}

	if _, err := json.Marshal(get); err != nil {
		t.Fatalf("marshal document task result: %v", err)
	}
}

func TestRetrievalTaskSearchLinksRecordsAndProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := local.Config{DataDir: filepath.Join(t.TempDir(), "data")}
	architecture := createDocument(t, ctx, config, "notes/architecture/knowledge-plane.md", "Knowledge plane", "# Knowledge plane\n\n## Summary\nCanonical architecture note.\n")
	roadmap := createDocument(t, ctx, config, "notes/projects/roadmap.md", "Roadmap", "# Roadmap\n\n## Summary\nSee the [knowledge plane](../architecture/knowledge-plane.md).\n")
	createDocument(t, ctx, config, "records/assets/transmission-solenoid.md", "Transmission solenoid", "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\n---\n# Transmission solenoid\n\n## Facts\n- sku: SOL-1\n")

	search, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action: agentops.RetrievalTaskActionSearch,
		Search: agentops.SearchOptions{
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

	links, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action: agentops.RetrievalTaskActionDocumentLinks,
		DocID:  roadmap.DocID,
	})
	if err != nil {
		t.Fatalf("links task: %v", err)
	}
	if links.Links == nil || len(links.Links.Outgoing) != 1 || links.Links.Outgoing[0].DocID != architecture.DocID {
		t.Fatalf("links result = %+v", links)
	}

	graph, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action: agentops.RetrievalTaskActionGraph,
		DocID:  roadmap.DocID,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("graph task: %v", err)
	}
	if graph.Graph == nil || len(graph.Graph.Nodes) == 0 || len(graph.Graph.Edges) == 0 {
		t.Fatalf("graph result = %+v", graph)
	}

	records, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action:  agentops.RetrievalTaskActionRecordsLookup,
		Records: agentops.RecordLookupOptions{Text: "solenoid", Limit: 10},
	})
	if err != nil {
		t.Fatalf("records task: %v", err)
	}
	if records.Records == nil || len(records.Records.Entities) != 1 {
		t.Fatalf("records result = %+v", records)
	}

	entity, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action:   agentops.RetrievalTaskActionRecordEntity,
		EntityID: records.Records.Entities[0].EntityID,
	})
	if err != nil {
		t.Fatalf("record entity task: %v", err)
	}
	if entity.Entity == nil || entity.Entity.EntityID != "transmission-solenoid" {
		t.Fatalf("entity result = %+v", entity)
	}

	provenance, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action: agentops.RetrievalTaskActionProvenanceEvents,
		Provenance: agentops.ProvenanceEventOptions{
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

	projections, err := agentops.RunRetrievalTask(ctx, config, agentops.RetrievalTaskRequest{
		Action: agentops.RetrievalTaskActionProjectionStates,
		Projection: agentops.ProjectionStateOptions{
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
}

func TestValidationRejectionDoesNotCreateRuntimeFiles(t *testing.T) {
	t.Parallel()

	dataDir := filepath.Join(t.TempDir(), "data")
	result, err := agentops.RunDocumentTask(context.Background(), local.Config{DataDir: dataDir}, agentops.DocumentTaskRequest{
		Action: agentops.DocumentTaskActionCreate,
		Document: agentops.DocumentInput{
			Title: "Missing path",
			Body:  "# Missing path\n",
		},
	})
	if err != nil {
		t.Fatalf("document task: %v", err)
	}
	if !result.Rejected || result.RejectionReason == "" {
		t.Fatalf("result = %+v, want rejected", result)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Fatalf("data dir exists after validation rejection: %v", err)
	}
}

func TestResolvePathsHonorsOpenClerkEnvOverrides(t *testing.T) {
	t.Setenv("OPENCLERK_DATA_DIR", filepath.Join(t.TempDir(), "env-data"))
	t.Setenv("OPENCLERK_DATABASE_PATH", filepath.Join(t.TempDir(), "env-db", "openclerk.sqlite"))
	t.Setenv("OPENCLERK_VAULT_ROOT", filepath.Join(t.TempDir(), "env-vault"))

	result, err := agentops.RunDocumentTask(context.Background(), local.Config{}, agentops.DocumentTaskRequest{
		Action: agentops.DocumentTaskActionResolvePaths,
	})
	if err != nil {
		t.Fatalf("resolve paths: %v", err)
	}
	if result.Paths == nil ||
		result.Paths.DataDir != os.Getenv("OPENCLERK_DATA_DIR") ||
		result.Paths.DatabasePath != os.Getenv("OPENCLERK_DATABASE_PATH") ||
		result.Paths.VaultRoot != os.Getenv("OPENCLERK_VAULT_ROOT") {
		t.Fatalf("paths = %+v", result.Paths)
	}

	explicit, err := agentops.RunDocumentTask(context.Background(), local.Config{
		DataDir:      filepath.Join(t.TempDir(), "explicit-data"),
		DatabasePath: filepath.Join(t.TempDir(), "explicit-db", "openclerk.sqlite"),
		VaultRoot:    filepath.Join(t.TempDir(), "explicit-vault"),
	}, agentops.DocumentTaskRequest{Action: agentops.DocumentTaskActionResolvePaths})
	if err != nil {
		t.Fatalf("resolve explicit paths: %v", err)
	}
	if explicit.Paths.DataDir == os.Getenv("OPENCLERK_DATA_DIR") ||
		explicit.Paths.DatabasePath == os.Getenv("OPENCLERK_DATABASE_PATH") ||
		explicit.Paths.VaultRoot == os.Getenv("OPENCLERK_VAULT_ROOT") {
		t.Fatalf("explicit config did not take precedence: %+v", explicit.Paths)
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
		t.Fatalf("create %s: %v", path, err)
	}
	if result.Document == nil {
		t.Fatalf("create %s result = %+v", path, result)
	}
	return *result.Document
}
