package evals_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/chronicler"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func TestRepoDocsAgentWorkflowInspectContextRecord(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/auth-callback.md", "Auth Callback Source", "# Auth Callback Source\n\n## Summary\nThe auth callback validates state before session renewal and returns retryable errors.\n")
	createDocument(t, ctx, config, "decisions/adr-auth-callback.md", "Auth Callback Decision", "---\ndecision_id: adr-auth-callback\ndecision_title: Keep callback validation before redirects\ndecision_status: accepted\ndecision_scope: auth\ndecision_owner: platform\n---\n# Auth Callback Decision\n\n## Summary\nKeep callback validation before redirects.\n")

	inspection, err := runclient.InspectExistingRuntime(ctx, config)
	if err != nil {
		t.Fatalf("inspect runtime: %v", err)
	}
	if !inspection.DatabaseExists || !inspection.DatabaseInitialized || inspection.DocumentCount < 2 {
		t.Fatalf("inspection = %+v", inspection)
	}

	contextPack, err := chronicler.RunContextPack(ctx, config, chronicler.RunRequest{
		Task:  "change the auth callback behavior",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("context pack: %v", err)
	}
	if contextPack.Action != chronicler.ActionContextPack ||
		!contextPack.Result.PlannedNoWrite ||
		contextPack.Result.WritesPerformed != 0 ||
		len(contextPack.Result.ContextPacks) != 1 ||
		len(contextPack.Result.ContextPacks[0].MustRead) == 0 ||
		len(contextPack.Result.ContextPacks[0].Citations) == 0 {
		t.Fatalf("context pack result = %+v", contextPack)
	}

	beforeDocuments := inspection.DocumentCount
	sessionPath := filepath.Join(t.TempDir(), "session.md")
	if err := os.WriteFile(sessionPath, []byte("# Completed Session\n\nChanged auth callback error wording and left a doc follow-up for callback telemetry.\n"), 0o644); err != nil {
		t.Fatalf("write session note: %v", err)
	}
	report, err := chronicler.RunSessionRecordReport(ctx, config, chronicler.RunRequest{
		InboxPaths: []string{sessionPath},
		Task:       "summarize completed auth callback work into repo knowledge",
		Limit:      5,
	})
	if err != nil {
		t.Fatalf("session record report: %v", err)
	}
	if report.Action != chronicler.ActionSessionRecordReport ||
		report.Result.Mode != chronicler.ActionSessionRecordReport ||
		!report.Result.PlannedNoWrite ||
		report.Result.WritesPerformed != 0 ||
		len(report.Result.InboxCandidates) != 1 ||
		len(report.Result.ContextPacks) != 1 ||
		report.Result.InboxCandidates[0].ApprovalBoundary == "" {
		t.Fatalf("session record report result = %+v", report)
	}
	afterInspection, err := runclient.InspectExistingRuntime(ctx, config)
	if err != nil {
		t.Fatalf("inspect runtime after session report: %v", err)
	}
	if afterInspection.DocumentCount != beforeDocuments {
		t.Fatalf("document count after session report = %d, want %d", afterInspection.DocumentCount, beforeDocuments)
	}
}
