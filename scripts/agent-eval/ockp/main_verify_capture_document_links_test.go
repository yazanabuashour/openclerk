package main

import (
	"context"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func TestVerifyCaptureDocumentLinksFetchRejectsExtraWrites(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	body := strings.TrimSpace(`---
type: source
source_type: web
source_url: "https://example.test/openclerk-runner-guidance"
---
# Runner Guidance Link

## Summary
source_type: web source_url: https://example.test/openclerk-runner-guidance `+webURLInitialText+`
`) + "\n"
	if err := createSeedDocument(ctx, cfg, captureDocumentLinksSourcePath, captureDocumentLinksSourceTitle, body); err != nil {
		t.Fatalf("seed document-these-links source: %v", err)
	}
	answer := strings.Join([]string{
		captureDocumentLinksSourcePath,
		"source_type web",
		"doc_id",
		"source.path_hint approved",
	}, " ")
	baseMetrics := metrics{
		AssistantCalls:      1,
		IngestSourceURLUsed: true,
		CommandExecutions:   1,
		ToolCalls:           1,
		EventTypeCounts:     map[string]int{},
	}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: captureDocumentLinksFetchScenarioID}, 1, answer, baseMetrics)
	if err != nil {
		t.Fatalf("verify document-these-links fetch: %v", err)
	}
	if !result.Passed {
		t.Fatalf("document-these-links fetch verification failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics){
		"create_document":  func(m *metrics) { m.CreateDocumentUsed = true },
		"append_document":  func(m *metrics) { m.AppendDocumentUsed = true },
		"replace_section":  func(m *metrics) { m.ReplaceSectionUsed = true },
		"ingest_video_url": func(m *metrics) { m.IngestVideoURLUsed = true },
	} {
		mutatingMetrics := baseMetrics
		mutate(&mutatingMetrics)
		result, err = verifyScenarioTurn(ctx, paths, scenario{ID: captureDocumentLinksFetchScenarioID}, 1, answer, mutatingMetrics)
		if err != nil {
			t.Fatalf("verify document-these-links fetch with %s: %v", name, err)
		}
		if result.Passed {
			t.Fatalf("document-these-links fetch passed with extra %s: %+v", name, result)
		}
	}
}
