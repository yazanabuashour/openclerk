package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	local "github.com/yazanabuashour/openclerk/client/local"
	openclerk "github.com/yazanabuashour/openclerk/client/openclerk"
)

type options struct {
	dataDir       string
	databasePath  string
	vaultRoot     string
	query         string
	pathPrefix    string
	metadataKey   string
	metadataValue string
	docID         string
	records       bool
	provenance    bool
	limit         int
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	cfg := local.Config{
		DataDir:      opts.dataDir,
		DatabasePath: opts.databasePath,
		VaultRoot:    opts.vaultRoot,
	}
	paths, err := local.ResolvePaths(cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "dataDir=%s\n", paths.DataDir)
	fmt.Fprintf(out, "db=%s\n", paths.DatabasePath)
	fmt.Fprintf(out, "vault=%s\n", paths.VaultRoot)

	client, runtime, err := local.Open(cfg)
	if err != nil {
		return err
	}
	defer runtime.Close()

	if opts.docID != "" {
		if err := printLinks(ctx, out, client, opts.docID); err != nil {
			return err
		}
		if opts.provenance {
			if err := printProvenance(ctx, out, client, opts.docID, opts.limit); err != nil {
				return err
			}
		}
		if opts.query == "" {
			return nil
		}
	}
	if opts.query != "" {
		if err := printSearch(ctx, out, client, opts); err != nil {
			return err
		}
		if opts.records {
			if err := printRecords(ctx, out, client, opts.query, opts.limit); err != nil {
				return err
			}
		}
		return nil
	}
	return printDocuments(ctx, out, client, opts)
}

func parseOptions(args []string) (options, error) {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	opts := options{limit: 20}
	fs := flag.NewFlagSet("openclerk-query", flag.ContinueOnError)
	fs.StringVar(&opts.dataDir, "data-dir", "", "OpenClerk data directory")
	fs.StringVar(&opts.databasePath, "db", "", "OpenClerk SQLite database path")
	fs.StringVar(&opts.vaultRoot, "vault", "", "OpenClerk vault root")
	fs.StringVar(&opts.query, "q", "", "search text")
	fs.StringVar(&opts.pathPrefix, "path-prefix", "", "document path prefix filter")
	fs.StringVar(&opts.metadataKey, "metadata-key", "", "document metadata key filter")
	fs.StringVar(&opts.metadataValue, "metadata-value", "", "document metadata value filter")
	fs.StringVar(&opts.docID, "doc-id", "", "document id for link and provenance inspection")
	fs.BoolVar(&opts.records, "records", false, "include promoted record lookup for -q")
	fs.BoolVar(&opts.provenance, "provenance", false, "include provenance and projection summaries")
	fs.IntVar(&opts.limit, "limit", 20, "maximum results to print")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	if opts.limit < 1 {
		opts.limit = 1
	}
	if opts.limit > 100 {
		opts.limit = 100
	}
	return opts, nil
}

func printDocuments(ctx context.Context, out io.Writer, client *openclerk.ClientWithResponses, opts options) error {
	params := openclerk.ListDocumentsParams{Limit: &opts.limit}
	if opts.pathPrefix != "" {
		params.PathPrefix = &opts.pathPrefix
	}
	if opts.metadataKey != "" {
		params.MetadataKey = &opts.metadataKey
	}
	if opts.metadataValue != "" {
		params.MetadataValue = &opts.metadataValue
	}
	response, err := client.ListDocumentsWithResponse(ctx, &params)
	if err != nil {
		return err
	}
	if response.JSON200 == nil {
		return fmt.Errorf("list documents failed: %s", string(response.Body))
	}
	fmt.Fprintf(out, "documents=%d hasMore=%t\n", len(response.JSON200.Documents), response.JSON200.PageInfo.HasMore)
	for _, document := range response.JSON200.Documents {
		fmt.Fprintf(out, "doc %s path=%s title=%q updated=%s", document.DocId, document.Path, document.Title, document.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
		if len(document.Metadata) > 0 {
			fmt.Fprintf(out, " metadata=%s", formatMap(document.Metadata))
		}
		fmt.Fprintln(out)
	}
	return nil
}

func printSearch(ctx context.Context, out io.Writer, client *openclerk.ClientWithResponses, opts options) error {
	query := openclerk.SearchQuery{Text: opts.query, Limit: &opts.limit}
	if opts.pathPrefix != "" {
		query.PathPrefix = &opts.pathPrefix
	}
	if opts.metadataKey != "" {
		query.MetadataKey = &opts.metadataKey
	}
	if opts.metadataValue != "" {
		query.MetadataValue = &opts.metadataValue
	}
	response, err := client.SearchQueryWithResponse(ctx, query)
	if err != nil {
		return err
	}
	if response.JSON200 == nil {
		return fmt.Errorf("search failed: %s", string(response.Body))
	}
	fmt.Fprintf(out, "searchHits=%d hasMore=%t\n", len(response.JSON200.Hits), response.JSON200.PageInfo.HasMore)
	for _, hit := range response.JSON200.Hits {
		path := firstCitationPath(hit.Citations)
		fmt.Fprintf(out, "hit rank=%d score=%.3f doc=%s chunk=%s path=%s title=%q snippet=%q\n", hit.Rank, hit.Score, hit.DocId, hit.ChunkId, path, hit.Title, compact(hit.Snippet))
	}
	return nil
}

func printRecords(ctx context.Context, out io.Writer, client *openclerk.ClientWithResponses, text string, limit int) error {
	response, err := client.RecordsLookupWithResponse(ctx, openclerk.RecordsLookupRequest{Text: text, Limit: &limit})
	if err != nil {
		return err
	}
	if response.JSON200 == nil {
		return fmt.Errorf("records lookup failed: %s", string(response.Body))
	}
	fmt.Fprintf(out, "records=%d hasMore=%t\n", len(response.JSON200.Entities), response.JSON200.PageInfo.HasMore)
	for _, entity := range response.JSON200.Entities {
		fmt.Fprintf(out, "record %s type=%s name=%q facts=%d summary=%q\n", entity.EntityId, entity.EntityType, entity.Name, len(entity.Facts), compact(entity.Summary))
	}
	return nil
}

func printLinks(ctx context.Context, out io.Writer, client *openclerk.ClientWithResponses, docID string) error {
	response, err := client.GetDocumentLinksWithResponse(ctx, docID)
	if err != nil {
		return err
	}
	if response.JSON200 == nil {
		return fmt.Errorf("links failed: %s", string(response.Body))
	}
	fmt.Fprintf(out, "links doc=%s outgoing=%d incoming=%d\n", response.JSON200.DocId, len(response.JSON200.Outgoing), len(response.JSON200.Incoming))
	for _, link := range response.JSON200.Outgoing {
		fmt.Fprintf(out, "outgoing %s path=%s title=%q citations=%d\n", link.DocId, link.Path, link.Title, len(link.Citations))
	}
	for _, link := range response.JSON200.Incoming {
		fmt.Fprintf(out, "incoming %s path=%s title=%q citations=%d\n", link.DocId, link.Path, link.Title, len(link.Citations))
	}
	return nil
}

func printProvenance(ctx context.Context, out io.Writer, client *openclerk.ClientWithResponses, docID string, limit int) error {
	refKind := "document"
	events, err := client.ListProvenanceEventsWithResponse(ctx, &openclerk.ListProvenanceEventsParams{
		RefKind: &refKind,
		RefId:   &docID,
		Limit:   &limit,
	})
	if err != nil {
		return err
	}
	if events.JSON200 == nil {
		return fmt.Errorf("provenance failed: %s", string(events.Body))
	}
	fmt.Fprintf(out, "provenanceEvents=%d hasMore=%t\n", len(events.JSON200.Events), events.JSON200.PageInfo.HasMore)
	for _, event := range events.JSON200.Events {
		fmt.Fprintf(out, "event %s type=%s source=%s occurred=%s details=%s\n", event.EventId, event.EventType, event.SourceRef, event.OccurredAt.Format("2006-01-02T15:04:05Z07:00"), formatMap(event.Details))
	}

	projections, err := client.ListProjectionStatesWithResponse(ctx, &openclerk.ListProjectionStatesParams{
		RefKind: &refKind,
		RefId:   &docID,
		Limit:   &limit,
	})
	if err != nil {
		return err
	}
	if projections.JSON200 == nil {
		return fmt.Errorf("projections failed: %s", string(projections.Body))
	}
	fmt.Fprintf(out, "projections=%d hasMore=%t\n", len(projections.JSON200.Projections), projections.JSON200.PageInfo.HasMore)
	for _, projection := range projections.JSON200.Projections {
		fmt.Fprintf(out, "projection %s freshness=%s source=%s updated=%s details=%s\n", projection.Projection, projection.Freshness, projection.SourceRef, projection.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), formatMap(projection.Details))
	}
	return nil
}

func firstCitationPath(citations []openclerk.Citation) string {
	if len(citations) == 0 {
		return ""
	}
	return citations[0].Path
}

func compact(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= 120 {
		return value
	}
	return value[:117] + "..."
}

func formatMap(values map[string]string) string {
	if len(values) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(values))
	for key, value := range values {
		parts = append(parts, fmt.Sprintf("%s=%q", key, value))
	}
	sort.Strings(parts)
	return "{" + strings.Join(parts, ",") + "}"
}
