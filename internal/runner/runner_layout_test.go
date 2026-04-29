package runner_test

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentTaskInspectLayout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nCanonical source guidance.\n")
	createDocument(t, ctx, config, "synthesis/runner.md", "Runner Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner.md
---
# Runner Synthesis

## Summary
Canonical source guidance.

## Sources
- sources/runner.md

## Freshness
Checked source refs.
`)+"\n")
	createDocument(t, ctx, config, "records/assets/openclerk-runner.md", "OpenClerk runner record", "---\nentity_type: tool\nentity_name: OpenClerk runner\n---\n# OpenClerk runner record\n\n## Facts\n- status: active\n")
	createDocument(t, ctx, config, "records/services/openclerk-runner.md", "OpenClerk runner", "---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service.\n")
	createDocument(t, ctx, config, "docs/architecture/runner-decision.md", "Runner decision", "---\ndecision_id: adr-runner\ndecision_title: Use JSON runner\ndecision_status: accepted\ndecision_scope: agentops\ndecision_owner: platform\n---\n# Runner decision\n\n## Summary\nUse the JSON runner for AgentOps.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect layout: %v", err)
	}
	if result.Layout == nil || !result.Layout.Valid {
		t.Fatalf("layout = %+v", result.Layout)
	}
	if result.Layout.Mode != "convention_first" || result.Layout.ConfigArtifactRequired {
		t.Fatalf("layout configuration = %+v", result.Layout)
	}
	if !layoutChecksInclude(result.Layout.Checks, "synthesis_source_refs_resolve", "pass") ||
		!layoutChecksInclude(result.Layout.Checks, "service_identity_metadata", "pass") ||
		!layoutChecksInclude(result.Layout.Checks, "decision_identity_metadata", "pass") ||
		!layoutChecksInclude(result.Layout.Checks, "record_identity_metadata", "pass") {
		t.Fatalf("layout checks = %+v", result.Layout.Checks)
	}
}

func TestDocumentTaskInspectLayoutUsesRootRelativeSourceAndSynthesisPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nCanonical source guidance.\n")
	createDocument(t, ctx, config, "synthesis/runner.md", "Runner Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: "sources/runner.md"
---
# Runner Synthesis

## Summary
Canonical source guidance.

## Sources
- sources/runner.md

## Freshness
Checked source refs.
`)+"\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect layout: %v", err)
	}
	if result.Layout == nil || !result.Layout.Valid {
		t.Fatalf("layout = %+v", result.Layout)
	}
	if !layoutPathConventionsInclude(result.Layout.ConventionalPaths, "sources/") ||
		!layoutPathConventionsInclude(result.Layout.ConventionalPaths, "synthesis/") {
		t.Fatalf("layout conventions = %+v", result.Layout.ConventionalPaths)
	}
	if !layoutChecksInclude(result.Layout.Checks, "synthesis_source_refs_resolve", "pass") {
		t.Fatalf("layout checks = %+v", result.Layout.Checks)
	}
	for _, check := range result.Layout.Checks {
		if check.ID != "optional_conventional_prefixes" {
			continue
		}
		prefixes := check.Details["path_prefixes"]
		if strings.Contains(prefixes, "sources/") || strings.Contains(prefixes, "synthesis/") {
			t.Fatalf("optional prefix warning reports populated root paths: %+v", check)
		}
	}
}

func TestDocumentTaskInspectLayoutDoesNotTreatNestedNotesSynthesisPathAsConvention(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/synthesis/legacy.md", "Legacy Synthesis", "# Legacy Synthesis\n\n## Summary\nLegacy nested path.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect layout: %v", err)
	}
	if result.Layout == nil || !result.Layout.Valid {
		t.Fatalf("layout = %+v", result.Layout)
	}
	for _, id := range []string{"synthesis_source_refs", "synthesis_sources_section", "synthesis_freshness_section"} {
		if layoutChecksInclude(result.Layout.Checks, id, "fail") {
			t.Fatalf("layout checks treated nested notes path as synthesis convention: %+v", result.Layout.Checks)
		}
	}
}

func TestDocumentTaskInspectLayoutReportsInvalidConventions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "synthesis/incomplete.md", "Incomplete Synthesis", "# Incomplete Synthesis\n\n## Summary\nMissing evidence conventions.\n")
	createDocument(t, ctx, config, "records/services/incomplete.md", "Incomplete Service", "---\nservice_id: incomplete\n---\n# Incomplete Service\n\n## Summary\nMissing service name.\n")
	createDocument(t, ctx, config, "records/decisions/incomplete.md", "Incomplete Decision", "---\ndecision_id: incomplete\n---\n# Incomplete Decision\n\n## Summary\nMissing decision title and status.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect invalid layout: %v", err)
	}
	if result.Layout == nil || result.Layout.Valid {
		t.Fatalf("layout = %+v, want invalid", result.Layout)
	}
	for _, id := range []string{
		"synthesis_source_refs",
		"synthesis_sources_section",
		"synthesis_freshness_section",
		"service_identity_metadata",
		"decision_identity_metadata",
	} {
		if !layoutChecksInclude(result.Layout.Checks, id, "fail") {
			t.Fatalf("layout checks missing failing %s: %+v", id, result.Layout.Checks)
		}
	}
}

func TestDocumentTaskInspectLayoutRequiresLevelTwoSynthesisSections(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/runner.md", "Runner Source", "# Runner Source\n\n## Summary\nCanonical source guidance.\n")
	createDocument(t, ctx, config, "synthesis/wrong-levels.md", "Wrong Level Synthesis", strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner.md
---
# Wrong Level Synthesis

## Summary
Canonical source guidance.

# Sources
- sources/runner.md

### Freshness
Checked source refs.
`)+"\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionInspectLayout,
	})
	if err != nil {
		t.Fatalf("inspect wrong-level layout: %v", err)
	}
	if result.Layout == nil || result.Layout.Valid {
		t.Fatalf("layout = %+v, want invalid", result.Layout)
	}
	for _, id := range []string{
		"synthesis_sources_section",
		"synthesis_freshness_section",
	} {
		if !layoutChecksInclude(result.Layout.Checks, id, "fail") {
			t.Fatalf("layout checks missing failing %s: %+v", id, result.Layout.Checks)
		}
	}
}

func layoutChecksInclude(checks []runner.KnowledgeLayoutCheck, id string, status string) bool {
	for _, check := range checks {
		if check.ID == id && check.Status == status {
			return true
		}
	}
	return false
}

func layoutPathConventionsInclude(conventions []runner.LayoutPathConvention, pathPrefix string) bool {
	for _, convention := range conventions {
		if convention.PathPrefix == pathPrefix {
			return true
		}
	}
	return false
}
