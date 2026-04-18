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
	if _, err := fmt.Fprintf(out, "dataDir=%s\n", paths.DataDir); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "db=%s\n", paths.DatabasePath); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "vault=%s\n", paths.VaultRoot); err != nil {
		return err
	}

	client, err := local.OpenClient(cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

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

func printDocuments(ctx context.Context, out io.Writer, client *local.Client, opts options) error {
	response, err := client.ListDocuments(ctx, local.DocumentListOptions{
		PathPrefix:    opts.pathPrefix,
		MetadataKey:   opts.metadataKey,
		MetadataValue: opts.metadataValue,
		Limit:         opts.limit,
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "documents=%d hasMore=%t\n", len(response.Documents), response.PageInfo.HasMore); err != nil {
		return err
	}
	for _, document := range response.Documents {
		if _, err := fmt.Fprintf(out, "doc %s path=%s title=%q updated=%s", document.DocID, document.Path, document.Title, document.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")); err != nil {
			return err
		}
		if len(document.Metadata) > 0 {
			if _, err := fmt.Fprintf(out, " metadata=%s", formatMap(document.Metadata)); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
	}
	return nil
}

func printSearch(ctx context.Context, out io.Writer, client *local.Client, opts options) error {
	response, err := client.Search(ctx, local.SearchOptions{
		Text:          opts.query,
		PathPrefix:    opts.pathPrefix,
		MetadataKey:   opts.metadataKey,
		MetadataValue: opts.metadataValue,
		Limit:         opts.limit,
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "searchHits=%d hasMore=%t\n", len(response.Hits), response.PageInfo.HasMore); err != nil {
		return err
	}
	for _, hit := range response.Hits {
		path := firstCitationPath(hit.Citations)
		if _, err := fmt.Fprintf(out, "hit rank=%d score=%.3f doc=%s chunk=%s path=%s title=%q snippet=%q\n", hit.Rank, hit.Score, hit.DocID, hit.ChunkID, path, hit.Title, compact(hit.Snippet)); err != nil {
			return err
		}
	}
	return nil
}

func printRecords(ctx context.Context, out io.Writer, client *local.Client, text string, limit int) error {
	response, err := client.LookupRecords(ctx, local.RecordLookupOptions{Text: text, Limit: limit})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "records=%d hasMore=%t\n", len(response.Entities), response.PageInfo.HasMore); err != nil {
		return err
	}
	for _, entity := range response.Entities {
		if _, err := fmt.Fprintf(out, "record %s type=%s name=%q facts=%d summary=%q\n", entity.EntityID, entity.EntityType, entity.Name, len(entity.Facts), compact(entity.Summary)); err != nil {
			return err
		}
	}
	return nil
}

func printLinks(ctx context.Context, out io.Writer, client *local.Client, docID string) error {
	response, err := client.GetDocumentLinks(ctx, docID)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "links doc=%s outgoing=%d incoming=%d\n", response.DocID, len(response.Outgoing), len(response.Incoming)); err != nil {
		return err
	}
	for _, link := range response.Outgoing {
		if _, err := fmt.Fprintf(out, "outgoing %s path=%s title=%q citations=%d\n", link.DocID, link.Path, link.Title, len(link.Citations)); err != nil {
			return err
		}
	}
	for _, link := range response.Incoming {
		if _, err := fmt.Fprintf(out, "incoming %s path=%s title=%q citations=%d\n", link.DocID, link.Path, link.Title, len(link.Citations)); err != nil {
			return err
		}
	}
	return nil
}

func printProvenance(ctx context.Context, out io.Writer, client *local.Client, docID string, limit int) error {
	events, err := client.ListProvenanceEvents(ctx, local.ProvenanceEventOptions{
		RefKind: "document",
		RefID:   docID,
		Limit:   limit,
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "provenanceEvents=%d hasMore=%t\n", len(events.Events), events.PageInfo.HasMore); err != nil {
		return err
	}
	for _, event := range events.Events {
		if _, err := fmt.Fprintf(out, "event %s type=%s source=%s occurred=%s details=%s\n", event.EventID, event.EventType, event.SourceRef, event.OccurredAt.Format("2006-01-02T15:04:05Z07:00"), formatMap(event.Details)); err != nil {
			return err
		}
	}

	projections, err := client.ListProjectionStates(ctx, local.ProjectionStateOptions{
		RefKind: "document",
		RefID:   docID,
		Limit:   limit,
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "projections=%d hasMore=%t\n", len(projections.Projections), projections.PageInfo.HasMore); err != nil {
		return err
	}
	for _, projection := range projections.Projections {
		if _, err := fmt.Fprintf(out, "projection %s freshness=%s source=%s updated=%s details=%s\n", projection.Projection, projection.Freshness, projection.SourceRef, projection.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), formatMap(projection.Details)); err != nil {
			return err
		}
	}
	return nil
}

func firstCitationPath(citations []local.Citation) string {
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
