package sqlite

import (
	"context"
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateDocumentRejectsDuplicatePath(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	first, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Widget One",
		Body:  "# Widget One\n\nfirst body",
	})
	if err != nil {
		t.Fatalf("create first document: %v", err)
	}

	_, err = store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Widget Two",
		Body:  "# Widget Two\n\nsecond body",
	})
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Status != 409 {
		t.Fatalf("duplicate create error = %v, want already exists 409", err)
	}

	got, err := store.GetDocument(context.Background(), first.DocID)
	if err != nil {
		t.Fatalf("get original document: %v", err)
	}
	if got.Title != "Widget One" || !strings.Contains(got.Body, "first body") {
		t.Fatalf("original document was overwritten: %+v", got)
	}
}

func TestMoveDocumentPreservesStableIDAndUpdatesLinks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	source, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "technology/projects.md",
		Title: "Projects",
		Body:  "# Projects\n\n## Summary\nTrack project ideas and see [Ideas](ideas.md).\n",
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "technology/ideas.md",
		Title: "Ideas",
		Body:  "# Ideas\n\n## Summary\nArchive older ideas.\n",
	}); err != nil {
		t.Fatalf("create outgoing target: %v", err)
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "technology/_index.md",
		Title: "Technology Index",
		Body:  "# Technology\n\n- [Projects](projects.md)\n",
	}); err != nil {
		t.Fatalf("create index: %v", err)
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "projects/idea-backlog.md",
		Title: "Idea Backlog",
		Body:  "# Idea Backlog\n\nSee [technology projects](../technology/projects.md).\n",
	}); err != nil {
		t.Fatalf("create backlink: %v", err)
	}

	plan, err := store.PlanMoveDocument(ctx, domain.MoveDocumentInput{
		Path:          "technology/projects.md",
		TargetPath:    "technology/project-ideas.md",
		UpdateLinks:   true,
		UpdateIndexes: true,
	})
	if err != nil {
		t.Fatalf("plan move: %v", err)
	}
	if plan.DocID != source.DocID ||
		plan.StableIDStatus != "frontmatter_id_will_be_added" ||
		plan.DuplicateRisk != "none" ||
		!movePlanIncludesLinkUpdate(plan.LinkUpdates, "projects/idea-backlog.md", "../technology/project-ideas.md") ||
		!movePlanIncludesLinkUpdate(plan.LinkUpdates, "technology/_index.md", "project-ideas.md") ||
		!movePlanIncludesIndexStatus(plan.IndexUpdates, "technology/_index.md", "link_update_planned") ||
		len(plan.OutgoingLinks) == 0 ||
		len(plan.IncomingLinks) == 0 {
		t.Fatalf("move plan = %+v", plan)
	}

	moved, err := store.MoveDocument(ctx, domain.MoveDocumentInput{
		Path:          "technology/projects.md",
		TargetPath:    "technology/project-ideas.md",
		UpdateLinks:   true,
		UpdateIndexes: true,
	})
	if err != nil {
		t.Fatalf("move document: %v", err)
	}
	if moved.Document.DocID != source.DocID || moved.Document.Path != "technology/project-ideas.md" {
		t.Fatalf("moved document identity/path = %+v", moved.Document)
	}
	if !strings.Contains(moved.Document.Body, "id: \""+source.DocID+"\"") {
		t.Fatalf("moved document body missing stable id:\n%s", moved.Document.Body)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "technology", "projects.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old path stat err = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "technology", "project-ideas.md")); err != nil {
		t.Fatalf("target path stat: %v", err)
	}
	backlog, err := store.GetDocument(ctx, docIDForPath("projects/idea-backlog.md"))
	if err != nil {
		t.Fatalf("get updated backlink: %v", err)
	}
	if !strings.Contains(backlog.Body, "../technology/project-ideas.md") || strings.Contains(backlog.Body, "../technology/projects.md") {
		t.Fatalf("backlink body = %q", backlog.Body)
	}
	index, err := store.GetDocument(ctx, docIDForPath("technology/_index.md"))
	if err != nil {
		t.Fatalf("get updated index: %v", err)
	}
	if !strings.Contains(index.Body, "](project-ideas.md)") || strings.Contains(index.Body, "](projects.md)") {
		t.Fatalf("index body = %q", index.Body)
	}
	events, err := store.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "document", RefID: source.DocID, Limit: 20})
	if err != nil {
		t.Fatalf("list move provenance: %v", err)
	}
	if !hasEventType(events.Events, "document_moved") {
		t.Fatalf("move provenance events = %+v", events.Events)
	}
	links, err := store.GetDocumentLinks(ctx, source.DocID)
	if err != nil {
		t.Fatalf("get moved links: %v", err)
	}
	if !documentLinksIncludePath(links.Incoming, "projects/idea-backlog.md") ||
		!documentLinksIncludePath(links.Incoming, "technology/_index.md") {
		t.Fatalf("incoming links after move = %+v", links.Incoming)
	}
}

func TestMoveDocumentRejectsExistingTarget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t, domain.BackendOpenClerk, filepath.Join(t.TempDir(), "openclerk.sqlite"), t.TempDir())
	defer func() {
		_ = store.Close()
	}()
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{Path: "notes/source.md", Title: "Source", Body: "# Source\n"}); err != nil {
		t.Fatalf("create source: %v", err)
	}
	if _, err := store.CreateDocument(ctx, domain.CreateDocumentInput{Path: "notes/target.md", Title: "Target", Body: "# Target\n"}); err != nil {
		t.Fatalf("create target: %v", err)
	}
	plan, err := store.PlanMoveDocument(ctx, domain.MoveDocumentInput{Path: "notes/source.md", TargetPath: "notes/target.md", UpdateLinks: true})
	if err != nil {
		t.Fatalf("plan duplicate move: %v", err)
	}
	if plan.DuplicateRisk != "target_document_exists" || plan.ExistingTarget == nil {
		t.Fatalf("duplicate plan = %+v", plan)
	}
	_, err = store.MoveDocument(ctx, domain.MoveDocumentInput{Path: "notes/source.md", TargetPath: "notes/target.md", UpdateLinks: true})
	var appErr *domain.Error
	if !errors.As(err, &appErr) || appErr.Code != "conflict" {
		t.Fatalf("move duplicate error = %v, want conflict", err)
	}
}

func TestMoveDocumentPreservesOutgoingRelativeLinksAcrossDirectories(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t, domain.BackendOpenClerk, filepath.Join(t.TempDir(), "openclerk.sqlite"), t.TempDir())
	defer func() {
		_ = store.Close()
	}()
	source, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/a.md",
		Title: "A",
		Body:  "# A\n\nSee [B](b.md).\n",
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	target, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "notes/b.md",
		Title: "B",
		Body:  "# B\n",
	})
	if err != nil {
		t.Fatalf("create outgoing target: %v", err)
	}

	plan, err := store.PlanMoveDocument(ctx, domain.MoveDocumentInput{
		Path:        "notes/a.md",
		TargetPath:  "archive/a.md",
		UpdateLinks: true,
	})
	if err != nil {
		t.Fatalf("plan cross-directory move: %v", err)
	}
	if !movePlanIncludesResolvedLinkUpdate(plan.LinkUpdates, "notes/a.md", "notes/b.md", "../notes/b.md") {
		t.Fatalf("cross-directory plan link updates = %+v", plan.LinkUpdates)
	}

	moved, err := store.MoveDocument(ctx, domain.MoveDocumentInput{
		Path:        "notes/a.md",
		TargetPath:  "archive/a.md",
		UpdateLinks: true,
	})
	if err != nil {
		t.Fatalf("move cross-directory document: %v", err)
	}
	if !strings.Contains(moved.Document.Body, "](../notes/b.md)") || strings.Contains(moved.Document.Body, "](b.md)") {
		t.Fatalf("moved body = %q", moved.Document.Body)
	}
	links, err := store.GetDocumentLinks(ctx, source.DocID)
	if err != nil {
		t.Fatalf("get moved outgoing links: %v", err)
	}
	if !documentLinksIncludePath(links.Outgoing, target.Path) {
		t.Fatalf("outgoing links after move = %+v", links.Outgoing)
	}
}

func TestSyncDocumentAdoptsFrontmatterIDAtSamePath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	vaultRoot := t.TempDir()
	store := openTestStore(t, domain.BackendOpenClerk, filepath.Join(t.TempDir(), "openclerk.sqlite"), vaultRoot)
	defer func() {
		_ = store.Close()
	}()
	created, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "docs/stable.md",
		Title: "Stable",
		Body:  "# Stable\n",
	})
	if err != nil {
		t.Fatalf("create path-hash document: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, "docs", "stable.md"), []byte("---\nid: \"stable-doc\"\n---\n# Stable\n\nNow stable.\n"), 0o600); err != nil {
		t.Fatalf("write stable id document: %v", err)
	}
	if err := store.syncDocumentFromDisk(ctx, "docs/stable.md", ""); err != nil {
		t.Fatalf("sync stable id document: %v", err)
	}
	stable, err := store.GetDocument(ctx, "stable-doc")
	if err != nil {
		t.Fatalf("get stable id document: %v", err)
	}
	if stable.Path != "docs/stable.md" || stable.CreatedAt != created.CreatedAt {
		t.Fatalf("stable document = %+v, created = %+v", stable, created)
	}
	if _, err := store.GetDocument(ctx, created.DocID); err == nil {
		t.Fatalf("old path-hash doc id still resolves after id adoption")
	}

	withID, err := store.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  "docs/created-with-id.md",
		Title: "Created With ID",
		Body:  "---\nid: \"created-explicit\"\n---\n# Created With ID\n",
	})
	if err != nil {
		t.Fatalf("create document with id: %v", err)
	}
	if withID.DocID != "created-explicit" || withID.Path != "docs/created-with-id.md" {
		t.Fatalf("create with id result = %+v", withID)
	}
}

func TestCreateDocumentPreservesRequestedTitleAcrossRestart(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)

	document, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/widget.md",
		Title: "Wanted Title",
		Body:  "body only no heading",
	})
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	if document.Title != "Wanted Title" {
		t.Fatalf("created document title = %q, want %q", document.Title, "Wanted Title")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	reopened := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = reopened.Close()
	}()

	got, err := reopened.GetDocument(context.Background(), document.DocID)
	if err != nil {
		t.Fatalf("get document after restart: %v", err)
	}
	if got.Title != "Wanted Title" {
		t.Fatalf("reopened document title = %q, want %q", got.Title, "Wanted Title")
	}
}

func TestCreateDocumentAllowsRepeatedIdenticalSections(t *testing.T) {
	t.Parallel()

	vaultRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "openclerk.sqlite")
	store := openTestStore(t, domain.BackendOpenClerk, dbPath, vaultRoot)
	defer func() {
		_ = store.Close()
	}()

	created, err := store.CreateDocument(context.Background(), domain.CreateDocumentInput{
		Path:  "docs/repeated-sections.md",
		Title: "Repeated Sections",
		Body: strings.TrimSpace(`
# Repeated Sections

## Example
Same example body.

## Example
Same example body.
`),
	})
	if err != nil {
		t.Fatalf("create document with repeated sections: %v", err)
	}
	search, err := store.Search(context.Background(), domain.SearchQuery{Text: "Same example body", Limit: 10})
	if err != nil {
		t.Fatalf("search repeated sections: %v", err)
	}
	matches := 0
	for _, hit := range search.Hits {
		if hit.DocID == created.DocID && len(hit.Citations) > 0 && hit.Citations[0].Heading == "Example" {
			matches++
		}
	}
	if matches != 2 {
		t.Fatalf("repeated section hits = %d, want 2; hits=%+v", matches, search.Hits)
	}
}

func movePlanIncludesLinkUpdate(updates []domain.DocumentLinkUpdate, docPath string, newTarget string) bool {
	for _, update := range updates {
		if update.Path == docPath && update.NewTarget == newTarget && update.Occurrences > 0 {
			return true
		}
	}
	return false
}

func movePlanIncludesResolvedLinkUpdate(updates []domain.DocumentLinkUpdate, docPath string, oldTarget string, newTarget string) bool {
	for _, update := range updates {
		if update.Path == docPath && update.OldTarget == oldTarget && update.NewTarget == newTarget && update.Occurrences > 0 {
			return true
		}
	}
	return false
}

func movePlanIncludesIndexStatus(updates []domain.DocumentIndexUpdate, indexPath string, status string) bool {
	for _, update := range updates {
		if update.Path == indexPath && update.Status == status {
			return true
		}
	}
	return false
}

func documentLinksIncludePath(links []domain.DocumentLink, docPath string) bool {
	for _, link := range links {
		if link.Path == docPath {
			return true
		}
	}
	return false
}
