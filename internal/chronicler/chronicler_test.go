package chronicler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestRunOnceEmptyIsReadOnlyReport(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	result, err := RunOnce(context.Background(), runclient.Config{DatabasePath: dbPath}, RunRequest{})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if result.SchemaVersion != SchemaVersion || result.Action != ActionRun {
		t.Fatalf("identity = %+v", result)
	}
	if result.Result.Mode != "once" || !result.Result.PlannedNoWrite || result.Result.WritesPerformed != 0 {
		t.Fatalf("write posture = %+v", result.Result)
	}
	if len(result.Result.InboxCandidates) != 0 ||
		len(result.Result.ContextPacks) != 0 ||
		len(result.Result.Blockers) != 0 {
		t.Fatalf("empty result = %+v", result.Result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("empty run created storage directory: %v", err)
	}
}

func TestRunOnceRejectsInvalidInboxPathWithoutStorage(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	result, err := RunOnce(context.Background(), runclient.Config{DatabasePath: dbPath}, RunRequest{
		InboxPaths: []string{filepath.Join(t.TempDir(), "missing.md")},
	})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(result.Result.Blockers) != 1 ||
		!strings.Contains(result.Result.Blockers[0], "local_inbox:missing.md") ||
		!strings.Contains(result.Result.Blockers[0], "not readable") {
		t.Fatalf("blockers = %+v", result.Result.Blockers)
	}
	if len(result.Result.InboxCandidates) != 0 || result.Result.WritesPerformed != 0 {
		t.Fatalf("invalid inbox result = %+v", result.Result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("invalid inbox created storage directory: %v", err)
	}
}

func TestRunOncePlansExplicitInboxNonRecursive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	createDocument(t, ctx, config, "notes/existing-renewal.md", "Existing Renewal", "# Existing Renewal\n\nRenewal marker already captured.\n")

	inboxRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(inboxRoot, "candidate.md"), []byte("# Renewal Candidate\n\nRenewal marker already captured with new details.\n"), 0o644); err != nil {
		t.Fatalf("write candidate: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inboxRoot, "ignored.bin"), []byte{0, 1, 2}, 0o644); err != nil {
		t.Fatalf("write ignored binary: %v", err)
	}
	nestedRoot := filepath.Join(inboxRoot, "nested")
	if err := os.MkdirAll(nestedRoot, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedRoot, "nested.md"), []byte("# Nested\n\nMust not be scanned.\n"), 0o644); err != nil {
		t.Fatalf("write nested: %v", err)
	}

	result, err := RunOnce(ctx, config, RunRequest{InboxPaths: []string{inboxRoot}, Limit: 5})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(result.Result.Blockers) != 0 {
		t.Fatalf("blockers = %+v", result.Result.Blockers)
	}
	if len(result.Result.InboxCandidates) != 1 {
		t.Fatalf("inbox candidates = %+v", result.Result.InboxCandidates)
	}
	candidate := result.Result.InboxCandidates[0]
	if candidate.SourceFile != "candidate.md" ||
		candidate.SourceRef != "local_inbox:candidate.md" ||
		candidate.WriteStatus != "planned_no_write" ||
		candidate.ProposedPath == "" ||
		!strings.Contains(strings.Join(candidate.SourceRefs, " "), "sha256:") {
		t.Fatalf("candidate = %+v", candidate)
	}
	if strings.Contains(candidate.Summary, "Nested") {
		t.Fatalf("candidate summary includes nested file: %+v", candidate)
	}
	if len(result.Result.PendingReview) != 1 || result.Result.WritesPerformed != 0 {
		t.Fatalf("review/write posture = %+v", result.Result)
	}
}

func TestRunSessionRecordReportPackagesInboxAndContextWithoutWrites(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	doc := createDocument(t, ctx, config, "docs/session-record.md", "Session Record", "# Session Record\n\nSession report marker evidence.\n")
	beforeBody := getDocumentBody(t, ctx, config, doc.DocID)
	beforeDocs := listDocumentCount(t, ctx, config)
	beforeEvents := provenanceEventCount(t, ctx, config)
	beforeStorage := storageSnapshot(t, filepath.Dir(dbPath))

	inboxPath := filepath.Join(t.TempDir(), "session.md")
	if err := os.WriteFile(inboxPath, []byte("# Completed Session\n\nSession report marker candidate update.\n"), 0o644); err != nil {
		t.Fatalf("write session note: %v", err)
	}
	result, err := RunSessionRecordReport(ctx, config, RunRequest{
		InboxPaths: []string{inboxPath},
		Task:       "Session report marker",
		Limit:      5,
	})
	if err != nil {
		t.Fatalf("session record report: %v", err)
	}
	if result.SchemaVersion != SchemaVersion ||
		result.Action != ActionSessionRecordReport ||
		result.Result.Mode != ActionSessionRecordReport ||
		!result.Result.PlannedNoWrite ||
		result.Result.WritesPerformed != 0 {
		t.Fatalf("identity/write posture = %+v", result)
	}
	if len(result.Result.InboxCandidates) != 1 ||
		len(result.Result.ContextPacks) != 1 ||
		len(result.Result.PendingReview) != 1 {
		t.Fatalf("report result = %+v", result.Result)
	}
	candidate := result.Result.InboxCandidates[0]
	if candidate.WriteStatus != "planned_no_write" ||
		candidate.NextCreateRequest == "" ||
		candidate.ApprovalBoundary == "" {
		t.Fatalf("candidate = %+v", candidate)
	}
	if len(result.Result.ContextPacks[0].MustRead) == 0 ||
		result.Result.ContextPacks[0].WriteStatus != "read_only_no_write" {
		t.Fatalf("context pack = %+v", result.Result.ContextPacks[0])
	}
	assertStorageUnchanged(t, beforeStorage, storageSnapshot(t, filepath.Dir(dbPath)))
	if got := getDocumentBody(t, ctx, config, doc.DocID); got != beforeBody {
		t.Fatalf("document body changed:\n%s", got)
	}
	if got := listDocumentCount(t, ctx, config); got != beforeDocs {
		t.Fatalf("document count = %d, want %d", got, beforeDocs)
	}
	if got := provenanceEventCount(t, ctx, config); got != beforeEvents {
		t.Fatalf("provenance event count = %d, want %d", got, beforeEvents)
	}
}

func TestRunInboxScanPlansExplicitInboxOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	createDocument(t, ctx, config, "notes/inbox-scan-seed.md", "Inbox Scan Seed", "# Inbox Scan Seed\n\nExisting Core storage seed.\n")
	inboxPath := filepath.Join(t.TempDir(), "candidate.txt")
	if err := os.WriteFile(inboxPath, []byte("Inbox scan standalone marker."), 0o644); err != nil {
		t.Fatalf("write inbox: %v", err)
	}

	result, err := RunInboxScan(ctx, config, RunRequest{
		InboxPaths: []string{inboxPath},
		Task:       "must not build context pack",
		Limit:      5,
	})
	if err != nil {
		t.Fatalf("inbox scan: %v", err)
	}
	if result.SchemaVersion != SchemaVersion ||
		result.Action != ActionInboxScan ||
		result.Result.Mode != ActionInboxScan ||
		!result.Result.PlannedNoWrite ||
		result.Result.WritesPerformed != 0 {
		t.Fatalf("identity/write posture = %+v", result)
	}
	if len(result.Result.InboxCandidates) != 1 || len(result.Result.ContextPacks) != 0 {
		t.Fatalf("scan result = %+v", result.Result)
	}
	if result.Result.InboxCandidates[0].WriteStatus != "planned_no_write" {
		t.Fatalf("candidate = %+v", result.Result.InboxCandidates[0])
	}
}

func TestRunInboxScanRequiresExistingStorageWithoutCreatingIt(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	inboxPath := filepath.Join(t.TempDir(), "candidate.md")
	if err := os.WriteFile(inboxPath, []byte("# Candidate\n\nStorage must already exist.\n"), 0o644); err != nil {
		t.Fatalf("write inbox: %v", err)
	}
	result, err := RunInboxScan(context.Background(), runclient.Config{DatabasePath: dbPath}, RunRequest{
		InboxPaths: []string{inboxPath},
	})
	if err != nil {
		t.Fatalf("inbox scan: %v", err)
	}
	if result.Action != ActionInboxScan ||
		len(result.Result.Blockers) != 1 ||
		result.Result.Blockers[0] != storageMissingBlocker ||
		len(result.Result.InboxCandidates) != 0 ||
		result.Result.WritesPerformed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("inbox scan created storage directory: %v", err)
	}
}

func TestRunOnceDoesNotReportCleanInboxCandidateAsDuplicateRisk(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	createDocument(t, ctx, config, "notes/clean-candidate-seed.md", "Clean Candidate Seed", "# Clean Candidate Seed\n\nExisting Core storage seed.\n")
	inboxPath := filepath.Join(t.TempDir(), "unique-note.md")
	if err := os.WriteFile(inboxPath, []byte("# Unique Note\n\nUnique chronicler marker qwerty.\n"), 0o644); err != nil {
		t.Fatalf("write inbox: %v", err)
	}

	result, err := RunOnce(ctx, config, RunRequest{InboxPaths: []string{inboxPath}, Limit: 5})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(result.Result.InboxCandidates) != 1 {
		t.Fatalf("inbox candidates = %+v", result.Result.InboxCandidates)
	}
	if result.Result.InboxCandidates[0].DuplicateRisk != "no_duplicate_found" {
		t.Fatalf("candidate duplicate status = %+v", result.Result.InboxCandidates[0])
	}
	if len(result.Result.DuplicateRisks) != 0 {
		t.Fatalf("duplicate risks = %+v", result.Result.DuplicateRisks)
	}
	if result.Result.InboxCandidates[0].RecommendedAction != "review_then_approve_create_document" {
		t.Fatalf("candidate action = %+v", result.Result.InboxCandidates[0])
	}
}

func TestRunContextPackBuildsContextOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	createDocument(t, ctx, config, "docs/context/context-pack-test.md", "Context Pack Test", "# Context Pack Test\n\nContext pack standalone marker evidence.\n")

	result, err := RunContextPack(ctx, config, RunRequest{
		Task:  "Context pack standalone marker",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("context pack: %v", err)
	}
	if result.SchemaVersion != SchemaVersion ||
		result.Action != ActionContextPack ||
		result.Result.Mode != ActionContextPack ||
		!result.Result.PlannedNoWrite ||
		result.Result.WritesPerformed != 0 {
		t.Fatalf("identity/write posture = %+v", result)
	}
	if len(result.Result.InboxCandidates) != 0 || len(result.Result.ContextPacks) != 1 {
		t.Fatalf("context result = %+v", result.Result)
	}
	if result.Result.ContextPacks[0].WriteStatus != "read_only_no_write" ||
		len(result.Result.ContextPacks[0].MustRead) == 0 {
		t.Fatalf("context pack = %+v", result.Result.ContextPacks[0])
	}
}

func TestRunContextPackRequiresTaskOrQueryWithoutStorage(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	result, err := RunContextPack(context.Background(), runclient.Config{DatabasePath: dbPath}, RunRequest{})
	if err != nil {
		t.Fatalf("context pack: %v", err)
	}
	if result.Action != ActionContextPack ||
		result.Result.Mode != ActionContextPack ||
		len(result.Result.Blockers) != 1 ||
		result.Result.Blockers[0] != "task or query is required for context_pack" {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Result.ContextPacks) != 0 || result.Result.WritesPerformed != 0 {
		t.Fatalf("context result = %+v", result.Result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("missing context query created storage directory: %v", err)
	}
}

func TestRunContextPackRequiresExistingStorageWithoutCreatingIt(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	result, err := RunContextPack(context.Background(), runclient.Config{DatabasePath: dbPath}, RunRequest{
		Task: "Storage must already exist",
	})
	if err != nil {
		t.Fatalf("context pack: %v", err)
	}
	if result.Action != ActionContextPack ||
		len(result.Result.Blockers) != 1 ||
		result.Result.Blockers[0] != storageMissingBlocker ||
		len(result.Result.ContextPacks) != 0 ||
		result.Result.WritesPerformed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("context pack created storage directory: %v", err)
	}
}

func TestRunOnceRequiresExistingStorageForPlanningWithoutCreatingIt(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	inboxPath := filepath.Join(t.TempDir(), "candidate.md")
	if err := os.WriteFile(inboxPath, []byte("# Candidate\n\nStorage must already exist.\n"), 0o644); err != nil {
		t.Fatalf("write inbox: %v", err)
	}
	result, err := RunOnce(context.Background(), runclient.Config{DatabasePath: dbPath}, RunRequest{
		InboxPaths: []string{inboxPath},
		Task:       "Storage must already exist",
	})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if result.Action != ActionRun ||
		len(result.Result.Blockers) != 1 ||
		result.Result.Blockers[0] != storageMissingBlocker ||
		len(result.Result.InboxCandidates) != 0 ||
		len(result.Result.ContextPacks) != 0 ||
		result.Result.WritesPerformed != 0 {
		t.Fatalf("result = %+v", result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("run once created storage directory: %v", err)
	}
}

func TestRunOnceBuildsContextPackFromRetrieval(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	createDocument(t, ctx, config, "docs/architecture/chronicler-test.md", "Chronicler Test", "# Chronicler Test\n\nChronicler context marker evidence for task planning.\n")
	createDocument(t, ctx, config, "docs/architecture/chronicler-decision.md", "Chronicler Decision", "---\ndecision_id: adr-chronicler-test\ndecision_title: Chronicler Lite stays read only\ndecision_status: accepted\ndecision_scope: chronicler\ndecision_owner: platform\ndecision_date: 2026-06-11\n---\n# Chronicler Decision\n\n## Summary\nChronicler context marker evidence says Chronicler Lite stays read only.\n")

	result, err := RunOnce(ctx, config, RunRequest{
		Task:       "Prepare Chronicler context marker implementation",
		Query:      "Chronicler Lite stays read only",
		PathPrefix: "docs/architecture/",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(result.Result.ContextPacks) != 1 {
		t.Fatalf("context packs = %+v", result.Result.ContextPacks)
	}
	pack := result.Result.ContextPacks[0]
	if pack.WriteStatus != "read_only_no_write" ||
		len(pack.MustRead) == 0 ||
		len(pack.RelevantDecisions) == 0 ||
		len(pack.Citations) == 0 ||
		!strings.Contains(pack.Summary, "Read-only context pack") {
		t.Fatalf("context pack = %+v", pack)
	}
	if pack.RelevantDecisions[0].DecisionID != "adr-chronicler-test" {
		t.Fatalf("decisions = %+v", pack.RelevantDecisions)
	}
}

func TestRunOnceFiltersDecisionContextByPathPrefix(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	createDocument(t, ctx, config, "notes/in-prefix.md", "In Prefix", "# In Prefix\n\nNeedle note text.\n")
	createDocument(t, ctx, config, "docs/architecture/out-of-prefix-decision.md", "Out Of Prefix Decision", "---\ndecision_id: adr-out-of-prefix\ndecision_title: Out of Prefix Decision\ndecision_status: accepted\ndecision_scope: chronicler\ndecision_owner: platform\ndecision_date: 2026-06-11\n---\n# Out Of Prefix Decision\n\nNeedle note text.\n")

	result, err := RunOnce(ctx, config, RunRequest{
		Task:       "Needle note text",
		PathPrefix: "notes/",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(result.Result.ContextPacks) != 1 {
		t.Fatalf("context packs = %+v", result.Result.ContextPacks)
	}
	pack := result.Result.ContextPacks[0]
	if len(pack.MustRead) != 1 || pack.MustRead[0].Path != "notes/in-prefix.md" {
		t.Fatalf("must read = %+v", pack.MustRead)
	}
	if len(pack.RelevantDecisions) != 0 {
		t.Fatalf("out-of-prefix decisions included = %+v", pack.RelevantDecisions)
	}
	for _, citation := range pack.Citations {
		if !strings.HasPrefix(citation.Path, "notes/") {
			t.Fatalf("out-of-prefix citation included = %+v", pack.Citations)
		}
	}
}

func TestRunOnceRejectsInvalidContextPathPrefixWithoutStorage(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	result, err := RunOnce(context.Background(), runclient.Config{DatabasePath: dbPath}, RunRequest{
		Task:       "blocked",
		PathPrefix: "../outside",
	})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(result.Result.Blockers) != 1 || result.Result.Blockers[0] != "path_prefix must be vault-relative and stay inside the vault root" {
		t.Fatalf("blockers = %+v", result.Result.Blockers)
	}
	if len(result.Result.ContextPacks) != 0 {
		t.Fatalf("context packs = %+v", result.Result.ContextPacks)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("invalid path prefix created storage directory: %v", err)
	}
}

func TestRunOnceDoesNotCreateDocumentsOrProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	doc := createDocument(t, ctx, config, "notes/stable.md", "Stable", "# Stable\n\nStable chronicler no-write marker.\n")
	beforeBody := getDocumentBody(t, ctx, config, doc.DocID)
	beforeDocs := listDocumentCount(t, ctx, config)
	beforeEvents := provenanceEventCount(t, ctx, config)
	beforeStorage := storageSnapshot(t, filepath.Dir(dbPath))

	inboxPath := filepath.Join(t.TempDir(), "new-note.txt")
	if err := os.WriteFile(inboxPath, []byte("Stable chronicler no-write marker candidate."), 0o644); err != nil {
		t.Fatalf("write inbox: %v", err)
	}
	result, err := RunOnce(ctx, config, RunRequest{
		InboxPaths: []string{inboxPath},
		Task:       "Stable chronicler no-write marker",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if !result.Result.PlannedNoWrite || result.Result.WritesPerformed != 0 {
		t.Fatalf("write posture = %+v", result.Result)
	}
	assertStorageUnchanged(t, beforeStorage, storageSnapshot(t, filepath.Dir(dbPath)))
	if got := getDocumentBody(t, ctx, config, doc.DocID); got != beforeBody {
		t.Fatalf("document body changed:\n%s", got)
	}
	if got := listDocumentCount(t, ctx, config); got != beforeDocs {
		t.Fatalf("document count = %d, want %d", got, beforeDocs)
	}
	if got := provenanceEventCount(t, ctx, config); got != beforeEvents {
		t.Fatalf("provenance event count = %d, want %d", got, beforeEvents)
	}
}

func TestRunInboxScanDoesNotCreateDocumentsOrProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	doc := createDocument(t, ctx, config, "notes/inbox-stable.md", "Inbox Stable", "# Inbox Stable\n\nStable inbox scan no-write marker.\n")
	beforeBody := getDocumentBody(t, ctx, config, doc.DocID)
	beforeDocs := listDocumentCount(t, ctx, config)
	beforeEvents := provenanceEventCount(t, ctx, config)
	beforeStorage := storageSnapshot(t, filepath.Dir(dbPath))

	inboxPath := filepath.Join(t.TempDir(), "new-note.txt")
	if err := os.WriteFile(inboxPath, []byte("Stable inbox scan no-write marker candidate."), 0o644); err != nil {
		t.Fatalf("write inbox: %v", err)
	}
	result, err := RunInboxScan(ctx, config, RunRequest{
		InboxPaths: []string{inboxPath},
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("inbox scan: %v", err)
	}
	if result.Action != ActionInboxScan || !result.Result.PlannedNoWrite || result.Result.WritesPerformed != 0 {
		t.Fatalf("write posture = %+v", result.Result)
	}
	assertStorageUnchanged(t, beforeStorage, storageSnapshot(t, filepath.Dir(dbPath)))
	if got := getDocumentBody(t, ctx, config, doc.DocID); got != beforeBody {
		t.Fatalf("document body changed:\n%s", got)
	}
	if got := listDocumentCount(t, ctx, config); got != beforeDocs {
		t.Fatalf("document count = %d, want %d", got, beforeDocs)
	}
	if got := provenanceEventCount(t, ctx, config); got != beforeEvents {
		t.Fatalf("provenance event count = %d, want %d", got, beforeEvents)
	}
}

func TestRunContextPackDoesNotCreateDocumentsOrProvenance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	doc := createDocument(t, ctx, config, "notes/context-stable.md", "Context Stable", "# Context Stable\n\nStable context pack no-write marker.\n")
	beforeBody := getDocumentBody(t, ctx, config, doc.DocID)
	beforeDocs := listDocumentCount(t, ctx, config)
	beforeEvents := provenanceEventCount(t, ctx, config)
	beforeStorage := storageSnapshot(t, filepath.Dir(dbPath))

	result, err := RunContextPack(ctx, config, RunRequest{
		Task:  "Stable context pack no-write marker",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("context pack: %v", err)
	}
	if result.Action != ActionContextPack || !result.Result.PlannedNoWrite || result.Result.WritesPerformed != 0 {
		t.Fatalf("write posture = %+v", result.Result)
	}
	assertStorageUnchanged(t, beforeStorage, storageSnapshot(t, filepath.Dir(dbPath)))
	if got := getDocumentBody(t, ctx, config, doc.DocID); got != beforeBody {
		t.Fatalf("document body changed:\n%s", got)
	}
	if got := listDocumentCount(t, ctx, config); got != beforeDocs {
		t.Fatalf("document count = %d, want %d", got, beforeDocs)
	}
	if got := provenanceEventCount(t, ctx, config); got != beforeEvents {
		t.Fatalf("provenance event count = %d, want %d", got, beforeEvents)
	}
}

func TestRunOnceNegativeLimitReturnsBlocker(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	result, err := RunOnce(context.Background(), runclient.Config{DatabasePath: dbPath}, RunRequest{
		Limit: -1,
	})
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(result.Result.Blockers) != 1 || result.Result.Blockers[0] != "limit must be greater than or equal to 0" {
		t.Fatalf("blockers = %+v", result.Result.Blockers)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("negative limit created storage directory: %v", err)
	}
}

func createDocument(t *testing.T, ctx context.Context, config runclient.Config, path string, title string, body string) runner.Document {
	t.Helper()
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  path,
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	if result.Document == nil {
		t.Fatalf("create %s result = %+v", path, result)
	}
	return *result.Document
}

func getDocumentBody(t *testing.T, ctx context.Context, config runclient.Config, docID string) string {
	t.Helper()
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  docID,
	})
	if err != nil {
		t.Fatalf("get %s: %v", docID, err)
	}
	if result.Document == nil {
		t.Fatalf("get %s result = %+v", docID, result)
	}
	return result.Document.Body
}

func listDocumentCount(t *testing.T, ctx context.Context, config runclient.Config) int {
	t.Helper()
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 100},
	})
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	return len(result.Documents)
}

func provenanceEventCount(t *testing.T, ctx context.Context, config runclient.Config) int {
	t.Helper()
	result, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{Limit: 100},
	})
	if err != nil {
		t.Fatalf("list provenance: %v", err)
	}
	if result.Provenance == nil {
		t.Fatalf("provenance result = %+v", result)
	}
	return len(result.Provenance.Events)
}

func storageSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	snapshot := map[string]string{}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return snapshot
	}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		snapshot[filepath.ToSlash(relative)] = hex.EncodeToString(sum[:])
		return nil
	}); err != nil {
		t.Fatalf("snapshot storage: %v", err)
	}
	return snapshot
}

func assertStorageUnchanged(t *testing.T, before map[string]string, after map[string]string) {
	t.Helper()
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("storage changed:\nbefore=%+v\nafter=%+v", before, after)
	}
}
