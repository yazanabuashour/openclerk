package local_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	localclient "github.com/yazanabuashour/openclerk/client/local"
	openclerkclient "github.com/yazanabuashour/openclerk/client/openclerk"
)

func TestResolvePathsDefaultAndEnvStorage(t *testing.T) {
	xdgDataHome := filepath.Join(t.TempDir(), "xdg")
	t.Setenv("XDG_DATA_HOME", xdgDataHome)

	paths, err := localclient.ResolvePaths(localclient.Config{})
	if err != nil {
		t.Fatalf("resolve default paths: %v", err)
	}
	wantDataDir := filepath.Join(xdgDataHome, "openclerk")
	if paths.DataDir != wantDataDir {
		t.Fatalf("data dir = %q, want %q", paths.DataDir, wantDataDir)
	}
	if paths.DatabasePath != filepath.Join(wantDataDir, "openclerk.sqlite") {
		t.Fatalf("database path = %q", paths.DatabasePath)
	}
	if paths.VaultRoot != filepath.Join(wantDataDir, "vault") {
		t.Fatalf("vault root = %q", paths.VaultRoot)
	}

	t.Setenv("OPENCLERK_DATA_DIR", filepath.Join(t.TempDir(), "env-data"))
	t.Setenv("OPENCLERK_DATABASE_PATH", filepath.Join(t.TempDir(), "env-db", "openclerk.sqlite"))
	t.Setenv("OPENCLERK_VAULT_ROOT", filepath.Join(t.TempDir(), "env-vault"))
	paths, err = localclient.ResolvePaths(localclient.Config{})
	if err != nil {
		t.Fatalf("resolve env paths: %v", err)
	}
	if paths.DataDir != os.Getenv("OPENCLERK_DATA_DIR") ||
		paths.DatabasePath != os.Getenv("OPENCLERK_DATABASE_PATH") ||
		paths.VaultRoot != os.Getenv("OPENCLERK_VAULT_ROOT") {
		t.Fatalf("env paths = %+v", paths)
	}

	explicitDataDir := filepath.Join(t.TempDir(), "explicit-data")
	paths, err = localclient.ResolvePaths(localclient.Config{DataDir: explicitDataDir})
	if err != nil {
		t.Fatalf("resolve explicit data dir paths: %v", err)
	}
	if paths.DataDir != explicitDataDir ||
		paths.DatabasePath != filepath.Join(explicitDataDir, "openclerk.sqlite") ||
		paths.VaultRoot != filepath.Join(explicitDataDir, "vault") {
		t.Fatalf("explicit data dir paths = %+v", paths)
	}
}

func TestOpenKnowledgePlaneSurface(t *testing.T) {
	t.Parallel()

	client, runtime, err := localclient.Open(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
	if err != nil {
		t.Fatalf("open knowledge-plane client: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

	capabilities, err := client.GetCapabilitiesWithResponse(context.Background())
	if err != nil {
		t.Fatalf("get capabilities: %v", err)
	}
	if capabilities.JSON200 == nil {
		t.Fatalf("capabilities error: %s", string(capabilities.Body))
	}
	if capabilities.JSON200.Backend != openclerkclient.Openclerk {
		t.Fatalf("backend = %q", capabilities.JSON200.Backend)
	}
	if !slices.Equal(enumStrings(capabilities.JSON200.Extensions), []string{"provenance", "graph", "records"}) {
		t.Fatalf("extensions = %v", capabilities.JSON200.Extensions)
	}

	architecture := createDocument(t, client, "notes/architecture/knowledge-plane.md", "Knowledge plane", strings.TrimSpace(`
---
type: note
status: active
---
# Knowledge plane

## Summary
Canonical agent-facing architecture note.
`))
	roadmap := createDocument(t, client, "notes/projects/openclerk-roadmap.md", "Roadmap", strings.TrimSpace(`
---
type: project
status: active
---
# Roadmap

## Summary
See the [knowledge plane](../architecture/knowledge-plane.md) architecture note.
`))
	createDocument(t, client, "records/assets/transmission-solenoid.md", "Transmission solenoid", strings.TrimSpace(`
---
entity_type: part
entity_name: Transmission solenoid
entity_id: transmission-solenoid
type: record
status: active
---
# Transmission solenoid

## Summary
Canonical promoted-domain baseline.

## Facts
- sku: SOL-1
- vendor: OpenClerk Motors
`))

	pathPrefix := "notes/"
	list, err := client.ListDocumentsWithResponse(context.Background(), &openclerkclient.ListDocumentsParams{PathPrefix: &pathPrefix})
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	if list.JSON200 == nil || len(list.JSON200.Documents) != 2 {
		t.Fatalf("list documents response = %#v", list.JSON200)
	}

	searchLimit := 5
	projectType := "type"
	projectValue := "project"
	search, err := client.SearchQueryWithResponse(context.Background(), openclerkclient.SearchQuery{
		Text:          "roadmap",
		Limit:         &searchLimit,
		MetadataKey:   &projectType,
		MetadataValue: &projectValue,
	})
	if err != nil {
		t.Fatalf("search query: %v", err)
	}
	if search.JSON200 == nil || len(search.JSON200.Hits) == 0 || search.JSON200.Hits[0].DocId != roadmap.docID {
		t.Fatalf("search response = %#v", search.JSON200)
	}

	links, err := client.GetDocumentLinksWithResponse(context.Background(), roadmap.docID)
	if err != nil {
		t.Fatalf("document links: %v", err)
	}
	if links.JSON200 == nil || len(links.JSON200.Outgoing) != 1 || links.JSON200.Outgoing[0].DocId != architecture.docID {
		t.Fatalf("document links response = %#v", links.JSON200)
	}

	lookup, err := client.RecordsLookupWithResponse(context.Background(), openclerkclient.RecordsLookupRequest{Text: "solenoid"})
	if err != nil {
		t.Fatalf("records lookup: %v", err)
	}
	if lookup.JSON200 == nil || len(lookup.JSON200.Entities) != 1 || lookup.JSON200.Entities[0].EntityId != "transmission-solenoid" {
		t.Fatalf("records lookup response = %#v", lookup.JSON200)
	}

	events := provenanceEvents(t, client, "document", roadmap.docID)
	if len(events) == 0 {
		t.Fatal("expected roadmap provenance events")
	}

	projection := "graph"
	projections, err := client.ListProjectionStatesWithResponse(context.Background(), &openclerkclient.ListProjectionStatesParams{
		Projection: &projection,
		RefId:      &roadmap.docID,
	})
	if err != nil {
		t.Fatalf("list projection states: %v", err)
	}
	if projections.JSON200 == nil || len(projections.JSON200.Projections) != 1 || projections.JSON200.Projections[0].Freshness != openclerkclient.Fresh {
		t.Fatalf("projection states response = %#v", projections.JSON200)
	}
}

func TestOpenReopenPreservesDocumentState(t *testing.T) {
	t.Parallel()

	dataDir := filepath.Join(t.TempDir(), "data")
	client, runtime, err := localclient.Open(localclient.Config{DataDir: dataDir})
	if err != nil {
		t.Fatalf("open knowledge-plane client: %v", err)
	}
	create := createDocument(t, client, "notes/ops/runbook.md", "Runbook", "# Runbook\n\n## Summary\nCanonical operating notes.\n")
	initialDocument := getDocument(t, client, create.docID)
	initialEvents := provenanceEvents(t, client, "document", create.docID)
	if len(initialEvents) != 1 || initialEvents[0].EventType != "document_created" {
		t.Fatalf("initial document events = %#v", initialEvents)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close first runtime: %v", err)
	}

	client, runtime, err = localclient.Open(localclient.Config{DataDir: dataDir})
	if err != nil {
		t.Fatalf("reopen knowledge-plane client: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

	reopenedDocument := getDocument(t, client, create.docID)
	if !reopenedDocument.UpdatedAt.Equal(initialDocument.UpdatedAt) {
		t.Fatalf("updatedAt changed across reopen: got %s want %s", reopenedDocument.UpdatedAt, initialDocument.UpdatedAt)
	}
	reopenedEvents := provenanceEvents(t, client, "document", create.docID)
	if len(reopenedEvents) != 1 || reopenedEvents[0].EventType != "document_created" {
		t.Fatalf("document events after reopen = %#v", reopenedEvents)
	}
}

func TestOpenPlainNotesDoNotInvalidateRecords(t *testing.T) {
	t.Parallel()

	client, runtime, err := localclient.Open(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
	if err != nil {
		t.Fatalf("open knowledge-plane client: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

	create := createDocument(t, client, "notes/team/briefing.md", "Briefing", "# Briefing\n\n## Summary\nAgent-facing team notes.\n")
	recordEvents := provenanceEvents(t, client, "projection", "records-source:"+create.docID)
	if len(recordEvents) != 0 {
		t.Fatalf("records invalidation events for plain note = %#v", recordEvents)
	}
}

func TestOpenGraphRefreshesOnlyAffectedDocuments(t *testing.T) {
	t.Parallel()

	client, runtime, err := localclient.Open(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
	if err != nil {
		t.Fatalf("open knowledge-plane client: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

	unrelated := createDocument(t, client, "notes/archive/unrelated.md", "Unrelated", "# Unrelated\n\n## Summary\nIndependent workspace context.\n")
	target := createDocument(t, client, "notes/reference/reference.md", "Reference", "# Reference\n\n## Summary\nLinked supporting note.\n")
	source := createDocument(t, client, "notes/reference/source.md", "Source", "# Source\n\n## Summary\nSee the [reference](reference.md).\n")

	replaceSection(t, client, source.docID, "Summary", "Updated summary without the link.")
	unrelatedEvents := provenanceEvents(t, client, "projection", "graph:"+unrelated.docID)
	if countEventType(unrelatedEvents, "projection_refreshed") != 1 {
		t.Fatalf("unrelated graph events = %#v", unrelatedEvents)
	}
	targetEvents := provenanceEvents(t, client, "projection", "graph:"+target.docID)
	if countEventType(targetEvents, "projection_refreshed") < 2 {
		t.Fatalf("target graph events = %#v", targetEvents)
	}
}

func TestOpenRecordsRefreshOnlyAffectedEntities(t *testing.T) {
	t.Parallel()

	client, runtime, err := localclient.Open(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
	if err != nil {
		t.Fatalf("open knowledge-plane client: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

	createDocument(t, client, "records/assets/transmission-solenoid.md", "Transmission solenoid", "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\n---\n# Transmission solenoid\n\n## Facts\n- sku: SOL-1\n")
	affected := createDocument(t, client, "records/assets/diagnostic-scanner.md", "Diagnostic scanner", "---\nentity_type: tool\nentity_name: Diagnostic scanner\nentity_id: diagnostic-scanner\n---\n# Diagnostic scanner\n\n## Facts\n- sku: TOOL-1\n")

	replaceSection(t, client, affected.docID, "Facts", "- sku: TOOL-2")
	unrelatedEvents := provenanceEvents(t, client, "projection", "records:transmission-solenoid")
	if len(unrelatedEvents) != 1 {
		t.Fatalf("unrelated record events = %#v", unrelatedEvents)
	}
	affectedEvents := provenanceEvents(t, client, "projection", "records:diagnostic-scanner")
	if len(affectedEvents) < 2 {
		t.Fatalf("affected record events = %#v", affectedEvents)
	}
}

type documentInfo struct {
	docID string
	path  string
}

func createDocument(t *testing.T, client *openclerkclient.ClientWithResponses, path, title, body string) documentInfo {
	t.Helper()
	response, err := client.CreateDocumentWithResponse(context.Background(), openclerkclient.CreateDocumentRequest{
		Path:  path,
		Title: title,
		Body:  body + "\n",
	})
	if err != nil {
		t.Fatalf("create document %s: %v", path, err)
	}
	if response.JSON201 == nil {
		t.Fatalf("create document %s error: %s", path, string(response.Body))
	}
	return documentInfo{docID: response.JSON201.DocId, path: response.JSON201.Path}
}

func getDocument(t *testing.T, client *openclerkclient.ClientWithResponses, docID string) openclerkclient.Document {
	t.Helper()
	response, err := client.GetDocumentWithResponse(context.Background(), docID)
	if err != nil {
		t.Fatalf("get document: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("get document error: %s", string(response.Body))
	}
	return *response.JSON200
}

func replaceSection(t *testing.T, client *openclerkclient.ClientWithResponses, docID string, heading string, content string) {
	t.Helper()
	response, err := client.ReplaceDocumentSectionWithResponse(context.Background(), docID, openclerkclient.ReplaceSectionRequest{
		Heading: heading,
		Content: content,
	})
	if err != nil {
		t.Fatalf("replace section: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("replace section error: %s", string(response.Body))
	}
}

func provenanceEvents(t *testing.T, client *openclerkclient.ClientWithResponses, refKind string, refID string) []openclerkclient.ProvenanceEvent {
	t.Helper()
	response, err := client.ListProvenanceEventsWithResponse(context.Background(), &openclerkclient.ListProvenanceEventsParams{
		RefKind: &refKind,
		RefId:   &refID,
	})
	if err != nil {
		t.Fatalf("list provenance events: %v", err)
	}
	if response.JSON200 == nil {
		t.Fatalf("list provenance events error: %s", string(response.Body))
	}
	return response.JSON200.Events
}

func countEventType(events []openclerkclient.ProvenanceEvent, eventType string) int {
	count := 0
	for _, event := range events {
		if event.EventType == eventType {
			count++
		}
	}
	return count
}

func enumStrings[T ~string](values []T) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}
