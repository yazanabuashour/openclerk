package main

import (
	"context"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func TestVerifyTaggingLookupRequiresForbiddenFixture(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, taggingDisambiguationTargetPath, "Customer Risk", taggedSeedBody("Customer Risk", taggingCustomerRiskTag, "Tagging exact customer risk evidence belongs to the active customer-risk tag.")); err != nil {
		t.Fatal(err)
	}

	result, err := verifyTaggingLookup(ctx, paths, "Found notes/tagging/customer-risk.md with tag customer-risk; exact tag disambiguation excluded adjacent matches and no durable write occurred.", metrics{
		SearchMetadataFilters: []string{"tag=" + taggingCustomerRiskTag},
		ListMetadataFilters:   []string{"tag=" + taggingCustomerRiskTag},
	}, taggingLookupExpectation{
		TargetPath:       taggingDisambiguationTargetPath,
		TargetTag:        taggingCustomerRiskTag,
		ForbiddenPath:    taggingDisambiguationDecoyPath,
		ForbiddenTag:     taggingCustomerRiskArchiveTag,
		RequireExactText: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || result.DatabasePass {
		t.Fatalf("tagging lookup passed without forbidden fixture: %+v", result)
	}
}

func TestTaggingBypassFailuresCatchInspectionFetchAndBrowser(t *testing.T) {
	failures := taggingBypassFailures(metrics{
		FileInspectionCommands: 1,
		ManualHTTPFetch:        true,
		BrowserAutomation:      true,
	})
	for _, want := range []string{
		"used direct file inspection for tagging workflow",
		"used manual HTTP fetch for tagging workflow",
		"used browser automation for tagging workflow",
	} {
		if !containsAnyString(failures, []string{want}) {
			t.Fatalf("missing %q in failures: %v", want, failures)
		}
	}
}
