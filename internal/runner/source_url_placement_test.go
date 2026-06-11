package runner

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func TestSourceURLPlacementPlansPDFSurface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openReadOnlyPlacementClient(t)
	defer func() {
		_ = client.Close()
	}()

	placement, err := planSourceURLPlacement(ctx, client, sourceURLPlacementInput{
		URL:   "https://Example.test/reports/latest.PDF#page=2",
		Title: "Latest Report",
	})
	if err != nil {
		t.Fatalf("source URL placement: %v", err)
	}
	if placement.SourceURL != "https://example.test/reports/latest.PDF" ||
		placement.SourceType != "pdf" ||
		placement.Slug != "latest-report" ||
		placement.DuplicateStatus != "no_existing_source_url_found" ||
		placement.CandidateSynthesisPath != "synthesis/latest-report.md" ||
		!slices.Contains(placement.CandidateSourcePaths, "sources/latest-report.md") ||
		!slices.Contains(placement.CandidateAssetPaths, "assets/sources/latest-report.pdf") {
		t.Fatalf("placement = %+v", placement)
	}
	next := placement.NextIngestSourceRequest("public_candidate_requires_ingest_source_url_approval")
	if !strings.Contains(next, `"url":"https://example.test/reports/latest.PDF"`) ||
		!strings.Contains(next, `"asset_path_hint":"assets/sources/latest-report.pdf"`) {
		t.Fatalf("next ingest request = %s", next)
	}
	if blocked := placement.NextIngestSourceRequest("blocked_no_fetch"); blocked != "" {
		t.Fatalf("blocked next ingest request = %s", blocked)
	}
}

func TestSourceURLPlacementCanonicalizesGitHubMarkdownSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openReadOnlyPlacementClient(t)
	defer func() {
		_ = client.Close()
	}()

	placement, err := planSourceURLPlacement(ctx, client, sourceURLPlacementInput{
		URL: "https://github.com/mvanhorn/last30days-skill",
	})
	if err != nil {
		t.Fatalf("source URL placement: %v", err)
	}
	if placement.SourceURL != "https://raw.githubusercontent.com/mvanhorn/last30days-skill/HEAD/README.md" ||
		placement.SourceType != "web" ||
		placement.Slug != "mvanhorn-last30days-skill" ||
		!slices.Contains(placement.CandidateSourcePaths, "sources/web/mvanhorn-last30days-skill.md") {
		t.Fatalf("repository placement = %+v", placement)
	}
	next := placement.NextIngestSourceRequest("public_candidate_requires_ingest_source_url_approval")
	if !strings.Contains(next, `"url":"https://raw.githubusercontent.com/mvanhorn/last30days-skill/HEAD/README.md"`) ||
		!strings.Contains(next, `"path_hint":"sources/web/mvanhorn-last30days-skill.md"`) {
		t.Fatalf("repository next ingest request = %s", next)
	}

	blob, err := planSourceURLPlacement(ctx, client, sourceURLPlacementInput{
		URL: "https://github.com/mvanhorn/last30days-skill/blob/main/skills/last30days/SKILL.md",
	})
	if err != nil {
		t.Fatalf("source URL placement for blob: %v", err)
	}
	if blob.SourceURL != "https://raw.githubusercontent.com/mvanhorn/last30days-skill/main/skills/last30days/SKILL.md" ||
		blob.Slug != "mvanhorn-last30days-skill-skill" ||
		!slices.Contains(blob.CandidateSourcePaths, "sources/web/mvanhorn-last30days-skill-skill.md") {
		t.Fatalf("blob placement = %+v", blob)
	}
}

func TestSourceURLPlacementFindsExistingRawSourceURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	created, err := RunDocumentTask(ctx, config, DocumentTaskRequest{
		Action: DocumentTaskActionCreate,
		Document: DocumentInput{
			Path:  "sources/web/existing-artifact-source.md",
			Title: "Existing Artifact Source",
			Body: strings.TrimSpace(`---
type: source
source_url: https://Example.test/artifact#section
source_type: web
---
# Existing Artifact Source

Existing source evidence.
`) + "\n",
		},
	})
	if err != nil {
		t.Fatalf("create existing source: %v", err)
	}

	client, err := runclient.OpenReadOnly(config)
	if err != nil {
		t.Fatalf("open client: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	placement, err := planSourceURLPlacement(ctx, client, sourceURLPlacementInput{
		URL:   "https://Example.test/artifact#section",
		Title: "Existing Artifact Source",
	})
	if err != nil {
		t.Fatalf("source URL placement: %v", err)
	}
	if placement.SourceURL != "https://example.test/artifact" ||
		placement.ExistingSource == nil ||
		placement.ExistingSource.DocID != created.Document.DocID ||
		placement.DuplicateStatus != "existing_source_url_found_no_fetch_no_write" ||
		placement.CandidateSynthesisPath != "" {
		t.Fatalf("placement = %+v", placement)
	}
	next := placement.NextIngestSourceRequest("public_candidate_requires_ingest_source_url_approval")
	if !strings.Contains(next, `"mode":"update"`) ||
		!strings.Contains(next, `"url":"https://example.test/artifact"`) {
		t.Fatalf("next ingest request = %s", next)
	}
}

func openReadOnlyPlacementClient(t *testing.T) *runclient.Client {
	t.Helper()

	client, err := runclient.OpenReadOnly(runclient.Config{
		DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite"),
	})
	if err != nil {
		t.Fatalf("open client: %v", err)
	}
	return client
}
