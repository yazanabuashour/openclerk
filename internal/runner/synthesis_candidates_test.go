package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func TestInspectSynthesisCandidatesPaginatesAndMatchesTarget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client, err := runclient.OpenForWrite(runclient.Config{
		DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite"),
	})
	if err != nil {
		t.Fatalf("open client: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	for index := 0; index < 105; index++ {
		path := fmt.Sprintf("synthesis/candidate-%03d.md", index)
		_, err := client.CreateDocument(ctx, domain.CreateDocumentInput{
			Path:  path,
			Title: fmt.Sprintf("Candidate %03d", index),
			Body:  fmt.Sprintf("# Candidate %03d\n\n## Summary\nCandidate synthesis document.\n", index),
		})
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
	}

	inspection, err := inspectSynthesisCandidates(ctx, client, "synthesis/candidate-042.md")
	if err != nil {
		t.Fatalf("inspect synthesis candidates: %v", err)
	}
	if len(inspection.Paths) != 105 ||
		!slices.IsSorted(inspection.Paths) ||
		!slices.Contains(inspection.Paths, "synthesis/candidate-000.md") ||
		!slices.Contains(inspection.Paths, "synthesis/candidate-104.md") {
		t.Fatalf("paths = %+v", inspection.Paths)
	}
	if len(inspection.TargetMatches) != 1 ||
		inspection.TargetMatches[0].Path != "synthesis/candidate-042.md" {
		t.Fatalf("target matches = %+v", inspection.TargetMatches)
	}
}
