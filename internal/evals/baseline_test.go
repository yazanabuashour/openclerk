package evals_test

import (
	"context"
	"path/filepath"
	"testing"

	ftsclient "github.com/yazanabuashour/openclerk/client/fts"
	graphclient "github.com/yazanabuashour/openclerk/client/graph"
	localclient "github.com/yazanabuashour/openclerk/client/local"
	openclerkclient "github.com/yazanabuashour/openclerk/client/openclerk"
	recordsclient "github.com/yazanabuashour/openclerk/client/records"
)

func TestImplementationVariantBaselines(t *testing.T) {
	t.Run("fts-source-grounded-search", func(t *testing.T) {
		client, runtime, err := localclient.OpenFTS(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
		if err != nil {
			t.Fatalf("open fts client: %v", err)
		}
		t.Cleanup(func() { _ = runtime.Close() })

		create, err := client.CreateDocumentWithResponse(context.Background(), ftsclient.CreateDocumentRequest{
			Path:  "notes/architecture/knowledge-plane.md",
			Title: "Knowledge plane",
			Body:  "# Knowledge plane\n\n## Summary\nCanonical architecture note.\n",
		})
		if err != nil || create.JSON201 == nil {
			t.Fatalf("create document failed: %v %s", err, string(create.Body))
		}
		search, err := client.SearchQueryWithResponse(context.Background(), ftsclient.SearchQuery{Text: "architecture"})
		if err != nil || search.JSON200 == nil || len(search.JSON200.Hits) != 1 {
			t.Fatalf("search response = %#v err=%v", search.JSON200, err)
		}
	})

	t.Run("hybrid-capability-baseline", func(t *testing.T) {
		client, runtime, err := localclient.OpenHybrid(localclient.Config{
			DataDir:           filepath.Join(t.TempDir(), "data"),
			EmbeddingProvider: "local",
		})
		if err != nil {
			t.Fatalf("open hybrid client: %v", err)
		}
		t.Cleanup(func() { _ = runtime.Close() })

		capabilities, err := client.GetCapabilitiesWithResponse(context.Background())
		if err != nil || capabilities.JSON200 == nil {
			t.Fatalf("get capabilities failed: %v %s", err, string(capabilities.Body))
		}
		if len(capabilities.JSON200.SearchModes) != 3 {
			t.Fatalf("search modes = %v", capabilities.JSON200.SearchModes)
		}
	})

	t.Run("graph-link-neighborhood", func(t *testing.T) {
		client, runtime, err := localclient.OpenGraph(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
		if err != nil {
			t.Fatalf("open graph client: %v", err)
		}
		t.Cleanup(func() { _ = runtime.Close() })

		target, err := client.CreateDocumentWithResponse(context.Background(), graphclient.CreateDocumentRequest{
			Path:  "notes/reference.md",
			Title: "Reference",
			Body:  "# Reference\n\n## Summary\nCanonical supporting note.\n",
		})
		if err != nil || target.JSON201 == nil {
			t.Fatalf("create target failed: %v %s", err, string(target.Body))
		}
		source, err := client.CreateDocumentWithResponse(context.Background(), graphclient.CreateDocumentRequest{
			Path:  "notes/guide.md",
			Title: "Guide",
			Body:  "# Guide\n\n## Summary\nSee the [reference](reference.md).\n",
		})
		if err != nil || source.JSON201 == nil {
			t.Fatalf("create source failed: %v %s", err, string(source.Body))
		}
		links, err := client.GraphNeighborhoodWithResponse(context.Background(), graphclient.GraphNeighborhoodRequest{DocId: &source.JSON201.DocId})
		if err != nil || links.JSON200 == nil || len(links.JSON200.Edges) == 0 {
			t.Fatalf("graph response = %#v err=%v", links.JSON200, err)
		}
	})

	t.Run("records-promoted-domain", func(t *testing.T) {
		client, runtime, err := localclient.OpenRecords(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
		if err != nil {
			t.Fatalf("open records client: %v", err)
		}
		t.Cleanup(func() { _ = runtime.Close() })

		create, err := client.CreateDocumentWithResponse(context.Background(), recordsclient.CreateDocumentRequest{
			Path:  "records/assets/transmission-solenoid.md",
			Title: "Transmission solenoid",
			Body:  "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\n---\n# Transmission solenoid\n\n## Summary\nCanonical promoted-domain baseline.\n\n## Facts\n- sku: SOL-1\n",
		})
		if err != nil || create.JSON201 == nil {
			t.Fatalf("create record failed: %v %s", err, string(create.Body))
		}
		lookup, err := client.RecordsLookupWithResponse(context.Background(), recordsclient.RecordsLookupRequest{Text: "solenoid"})
		if err != nil || lookup.JSON200 == nil || len(lookup.JSON200.Entities) != 1 {
			t.Fatalf("records lookup response = %#v err=%v", lookup.JSON200, err)
		}
	})

	t.Run("openclerk-unified-surface", func(t *testing.T) {
		client, runtime, err := localclient.Open(localclient.Config{DataDir: filepath.Join(t.TempDir(), "data")})
		if err != nil {
			t.Fatalf("open openclerk client: %v", err)
		}
		t.Cleanup(func() { _ = runtime.Close() })

		source, err := client.CreateDocumentWithResponse(context.Background(), openclerkclient.CreateDocumentRequest{
			Path:  "notes/architecture/knowledge-plane.md",
			Title: "Knowledge plane",
			Body:  "---\ntype: note\nstatus: active\n---\n# Knowledge plane\n\n## Summary\nCanonical architecture note.\n",
		})
		if err != nil || source.JSON201 == nil {
			t.Fatalf("create source failed: %v %s", err, string(source.Body))
		}
		target, err := client.CreateDocumentWithResponse(context.Background(), openclerkclient.CreateDocumentRequest{
			Path:  "notes/projects/openclerk-roadmap.md",
			Title: "Roadmap",
			Body:  "---\ntype: project\nstatus: active\n---\n# Roadmap\n\n## Summary\nSee the [knowledge plane](../architecture/knowledge-plane.md).\n",
		})
		if err != nil || target.JSON201 == nil {
			t.Fatalf("create target failed: %v %s", err, string(target.Body))
		}
		_, err = client.CreateDocumentWithResponse(context.Background(), openclerkclient.CreateDocumentRequest{
			Path:  "records/assets/transmission-solenoid.md",
			Title: "Transmission solenoid",
			Body:  "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\ntype: record\nstatus: active\n---\n# Transmission solenoid\n\n## Summary\nCanonical promoted-domain baseline.\n\n## Facts\n- sku: SOL-1\n",
		})
		if err != nil {
			t.Fatalf("create promoted record failed: %v", err)
		}

		pathPrefix := "notes/"
		list, err := client.ListDocumentsWithResponse(context.Background(), &openclerkclient.ListDocumentsParams{PathPrefix: &pathPrefix})
		if err != nil || list.JSON200 == nil || len(list.JSON200.Documents) != 2 {
			t.Fatalf("list documents response = %#v err=%v", list.JSON200, err)
		}
		links, err := client.GetDocumentLinksWithResponse(context.Background(), target.JSON201.DocId)
		if err != nil || links.JSON200 == nil || len(links.JSON200.Outgoing) != 1 || links.JSON200.Outgoing[0].DocId != source.JSON201.DocId {
			t.Fatalf("document links response = %#v err=%v", links.JSON200, err)
		}
		events, err := client.ListProvenanceEventsWithResponse(context.Background(), nil)
		if err != nil || events.JSON200 == nil || len(events.JSON200.Events) == 0 {
			t.Fatalf("provenance events response = %#v err=%v", events.JSON200, err)
		}
	})
}
