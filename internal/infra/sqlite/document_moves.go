package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

const (
	moveStableIDPresent = "frontmatter_id_present"
	moveStableIDAdded   = "frontmatter_id_will_be_added"

	moveDuplicateNone       = "none"
	moveDuplicateSamePath   = "same_path"
	moveDuplicateTargetDoc  = "target_document_exists"
	moveDuplicateTargetFile = "target_file_exists"
)

func (s *Store) PlanMoveDocument(ctx context.Context, input domain.MoveDocumentInput) (domain.DocumentMovePlan, error) {
	return s.planMoveDocument(ctx, input)
}

func (s *Store) MoveDocument(ctx context.Context, input domain.MoveDocumentInput) (domain.DocumentMoveResult, error) {
	plan, err := s.planMoveDocument(ctx, input)
	if err != nil {
		return domain.DocumentMoveResult{}, err
	}
	switch plan.DuplicateRisk {
	case moveDuplicateSamePath:
		return domain.DocumentMoveResult{}, domain.ValidationError("target_path must be different from the current document path", map[string]any{
			"path":        plan.SourcePath,
			"target_path": plan.TargetPath,
		})
	case moveDuplicateTargetDoc, moveDuplicateTargetFile:
		return domain.DocumentMoveResult{}, domain.ConflictError("target_path already exists; move_document does not overwrite or delete existing targets", map[string]any{
			"target_path":    plan.TargetPath,
			"duplicate_risk": plan.DuplicateRisk,
		})
	}

	sourceDoc, err := s.GetDocument(ctx, plan.DocID)
	if err != nil {
		return domain.DocumentMoveResult{}, err
	}
	if sourceDoc.Path != plan.SourcePath {
		return domain.DocumentMoveResult{}, domain.ConflictError("document path changed after move plan", map[string]any{
			"planned_path": plan.SourcePath,
			"current_path": sourceDoc.Path,
		})
	}

	sourceAbs, err := s.vaultExistingAbsPath(plan.SourcePath, "validate source document path")
	if err != nil {
		return domain.DocumentMoveResult{}, err
	}
	targetAbs, err := s.vaultCreateAbsPath(plan.TargetPath, "validate target document path")
	if err != nil {
		return domain.DocumentMoveResult{}, err
	}
	if err := ensureDir(filepath.Dir(targetAbs)); err != nil {
		return domain.DocumentMoveResult{}, domain.InternalError("create target document directory", err)
	}
	if _, err := osLstat(targetAbs); err == nil {
		return domain.DocumentMoveResult{}, domain.ConflictError("target_path already exists; choose an empty target path", map[string]any{
			"target_path": plan.TargetPath,
		})
	} else if !errors.Is(err, fs.ErrNotExist) {
		return domain.DocumentMoveResult{}, domain.InternalError("stat target document path", err)
	}

	movedBody := sourceDoc.Body
	if plan.StableIDStatus == moveStableIDAdded {
		movedBody = ensureMarkdownFrontmatterID(sourceDoc.Body, sourceDoc.DocID)
	}
	linkUpdatesApplied := []domain.DocumentLinkUpdate{}
	indexUpdatesApplied := []domain.DocumentIndexUpdate{}
	for _, update := range plan.LinkUpdates {
		if update.Path != plan.SourcePath {
			continue
		}
		var occurrences int
		movedBody, occurrences = rewriteMarkdownLinksForPlannedUpdate(movedBody, plan.SourcePath, update.OldTarget, update.NewTarget)
		if occurrences == 0 {
			continue
		}
		applied := update
		applied.Path = plan.TargetPath
		applied.Occurrences = occurrences
		linkUpdatesApplied = append(linkUpdatesApplied, applied)
		if applied.IndexCandidate {
			indexUpdatesApplied = append(indexUpdatesApplied, domain.DocumentIndexUpdate{
				Path:   plan.TargetPath,
				Status: "updated",
				Reason: "updated markdown link target from old path to new path",
			})
		}
	}
	if err := osRename(sourceAbs, targetAbs); err != nil {
		return domain.DocumentMoveResult{}, domain.InternalError("move document file", err)
	}
	if movedBody != sourceDoc.Body {
		if err := osWriteFile(targetAbs, movedBody); err != nil {
			_ = osRename(targetAbs, sourceAbs)
			return domain.DocumentMoveResult{}, domain.InternalError("write stable document id frontmatter", err)
		}
	}
	if err := s.syncDocumentFromDisk(ctx, plan.TargetPath, sourceDoc.Title); err != nil {
		return domain.DocumentMoveResult{}, err
	}

	for _, update := range plan.LinkUpdates {
		if update.Path == plan.SourcePath {
			continue
		}
		updatePath := update.Path
		absPath, err := s.vaultExistingAbsPath(updatePath, "validate link update document path")
		if err != nil {
			return domain.DocumentMoveResult{}, err
		}
		bodyBytes, err := osReadFile(absPath)
		if err != nil {
			return domain.DocumentMoveResult{}, domain.InternalError("read document for move link update", err)
		}
		updatedBody, occurrences := rewriteMarkdownLinksForPlannedUpdate(string(bodyBytes), updatePath, update.OldTarget, update.NewTarget)
		if occurrences == 0 {
			continue
		}
		if err := osWriteFile(absPath, updatedBody); err != nil {
			return domain.DocumentMoveResult{}, domain.InternalError("write document move link update", err)
		}
		if err := s.syncDocumentFromDisk(ctx, updatePath, ""); err != nil {
			return domain.DocumentMoveResult{}, err
		}
		applied := update
		applied.Path = updatePath
		applied.Occurrences = occurrences
		linkUpdatesApplied = append(linkUpdatesApplied, applied)
		if applied.IndexCandidate {
			indexUpdatesApplied = append(indexUpdatesApplied, domain.DocumentIndexUpdate{
				Path:   updatePath,
				Status: "updated",
				Reason: "updated markdown link target from old path to new path",
			})
		}
	}

	provenanceRefs, err := s.recordDocumentMoveProvenance(ctx, plan, len(linkUpdatesApplied), movedBody != sourceDoc.Body)
	if err != nil {
		return domain.DocumentMoveResult{}, err
	}
	projections, err := s.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		RefKind: "document",
		RefID:   plan.DocID,
		Limit:   20,
	})
	if err != nil {
		return domain.DocumentMoveResult{}, err
	}
	document, err := s.GetDocument(ctx, plan.DocID)
	if err != nil {
		return domain.DocumentMoveResult{}, err
	}
	plan.WriteStatus = "applied"
	return domain.DocumentMoveResult{
		Plan:                plan,
		Document:            document,
		LinkUpdatesApplied:  linkUpdatesApplied,
		IndexUpdatesApplied: indexUpdatesApplied,
		ProvenanceRefs:      provenanceRefs,
		ProjectionFreshness: projections.Projections,
		WriteStatus:         "applied",
	}, nil
}

func (s *Store) planMoveDocument(ctx context.Context, input domain.MoveDocumentInput) (domain.DocumentMovePlan, error) {
	sourceDoc, err := s.resolveMoveSourceDocument(ctx, input)
	if err != nil {
		return domain.DocumentMovePlan{}, err
	}
	targetPath, err := normalizePath(input.TargetPath)
	if err != nil {
		return domain.DocumentMovePlan{}, err
	}
	documents, err := s.loadAllDocuments(ctx)
	if err != nil {
		return domain.DocumentMovePlan{}, err
	}
	documentByPath := documentsByPath(documents)
	frontmatterID := strings.TrimSpace(sourceDoc.Metadata["id"])
	stableIDStatus := moveStableIDPresent
	if frontmatterID == "" {
		frontmatterID = sourceDoc.DocID
		stableIDStatus = moveStableIDAdded
	}

	links, err := s.GetDocumentLinks(ctx, sourceDoc.DocID)
	if err != nil {
		links = domain.DocumentLinks{DocID: sourceDoc.DocID}
	}
	duplicateRisk := moveDuplicateNone
	var existingTarget *domain.DocumentSummary
	if sourceDoc.Path == targetPath {
		duplicateRisk = moveDuplicateSamePath
	} else if targetDoc, ok := documentByPath[targetPath]; ok {
		duplicateRisk = moveDuplicateTargetDoc
		summary := documentSummaryFromDocument(targetDoc)
		existingTarget = &summary
	} else if targetAbs, err := s.vaultCreateAbsPath(targetPath, "validate target document path"); err != nil {
		return domain.DocumentMovePlan{}, err
	} else if _, statErr := osLstat(targetAbs); statErr == nil {
		duplicateRisk = moveDuplicateTargetFile
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return domain.DocumentMovePlan{}, domain.InternalError("stat target document path", statErr)
	}

	linkUpdates := []domain.DocumentLinkUpdate{}
	if input.UpdateLinks || input.UpdateIndexes {
		for _, doc := range documents {
			indexCandidate := path.Base(doc.Path) == "_index.md"
			if !input.UpdateLinks && (!input.UpdateIndexes || !indexCandidate) {
				continue
			}
			occurrences := countMarkdownLinksToPath(doc.Body, doc.Path, sourceDoc.Path)
			if occurrences == 0 {
				continue
			}
			linkDocPath := doc.Path
			if linkDocPath == sourceDoc.Path {
				linkDocPath = targetPath
			}
			linkUpdates = append(linkUpdates, domain.DocumentLinkUpdate{
				DocID:          doc.DocID,
				Path:           doc.Path,
				OldTarget:      sourceDoc.Path,
				NewTarget:      relativeMarkdownTarget(linkDocPath, targetPath, ""),
				Occurrences:    occurrences,
				IndexCandidate: indexCandidate,
			})
		}
	}
	if input.UpdateLinks {
		linkUpdates = append(linkUpdates, outgoingLinkUpdatesForMove(sourceDoc, targetPath)...)
	}
	sort.Slice(linkUpdates, func(i, j int) bool {
		if linkUpdates[i].Path != linkUpdates[j].Path {
			return linkUpdates[i].Path < linkUpdates[j].Path
		}
		return linkUpdates[i].OldTarget < linkUpdates[j].OldTarget
	})

	indexUpdates := moveIndexUpdateCandidates(documents, sourceDoc.Path, targetPath, linkUpdates, input.UpdateIndexes)
	projectionRefresh := moveProjectionRefresh(planRefreshDocIDs(sourceDoc.DocID, linkUpdates))
	warnings := moveValidationWarnings(sourceDoc, targetPath, duplicateRisk, input, linkUpdates, indexUpdates)
	return domain.DocumentMovePlan{
		DocID:                sourceDoc.DocID,
		SourcePath:           sourceDoc.Path,
		TargetPath:           targetPath,
		Title:                sourceDoc.Title,
		FrontmatterID:        frontmatterID,
		StableIDStatus:       stableIDStatus,
		DuplicateRisk:        duplicateRisk,
		ExistingTarget:       existingTarget,
		OutgoingLinks:        links.Outgoing,
		IncomingLinks:        links.Incoming,
		LinkUpdates:          linkUpdates,
		IndexUpdates:         indexUpdates,
		ProjectionRefresh:    projectionRefresh,
		ValidationWarnings:   warnings,
		WriteStatus:          "no_write",
		ApprovalBoundary:     "plan_move_document is read-only; move_document, rename_document, and promote_candidate require an approved non-propose-only document request and never overwrite existing targets",
		ValidationBoundaries: "Only the move file path, reported frontmatter id addition, reported markdown links, reported outgoing relative links, and reported index link targets are eligible for durable edits.",
	}, nil
}

func (s *Store) resolveMoveSourceDocument(ctx context.Context, input domain.MoveDocumentInput) (domain.Document, error) {
	if strings.TrimSpace(input.DocID) != "" && strings.TrimSpace(input.Path) != "" {
		return domain.Document{}, domain.ValidationError("provide only one of doc_id or path", nil)
	}
	if docID := strings.TrimSpace(input.DocID); docID != "" {
		return s.GetDocument(ctx, docID)
	}
	sourcePath, err := normalizePath(input.Path)
	if err != nil {
		return domain.Document{}, domain.ValidationError("doc_id or path is required", nil)
	}
	return s.getDocumentByPath(ctx, sourcePath)
}

func (s *Store) getDocumentByPath(ctx context.Context, relPath string) (domain.Document, error) {
	const query = `
SELECT doc_id, path, title, body, headings_json, metadata_json, created_at, updated_at
FROM documents
WHERE path = ?`
	var (
		document     domain.Document
		headingsJSON string
		metadataJSON string
		createdAt    string
		updatedAt    string
	)
	err := s.db.QueryRowContext(ctx, query, relPath).Scan(
		&document.DocID,
		&document.Path,
		&document.Title,
		&document.Body,
		&headingsJSON,
		&metadataJSON,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Document{}, domain.NotFoundError("document path", relPath)
	}
	if err != nil {
		return domain.Document{}, domain.InternalError("query document by path", err)
	}
	_ = json.Unmarshal([]byte(headingsJSON), &document.Headings)
	_ = json.Unmarshal([]byte(metadataJSON), &document.Metadata)
	document.CreatedAt = mustParseTime(createdAt)
	document.UpdatedAt = mustParseTime(updatedAt)
	return document, nil
}

func documentSummaryFromDocument(document domain.Document) domain.DocumentSummary {
	return domain.DocumentSummary{
		DocID:     document.DocID,
		Path:      document.Path,
		Title:     document.Title,
		Metadata:  document.Metadata,
		UpdatedAt: document.UpdatedAt,
	}
}

func countMarkdownLinksToPath(body string, docPath string, targetPath string) int {
	count := 0
	matches := linkPattern.FindAllStringSubmatchIndex(body, -1)
	for _, match := range matches {
		if len(match) < 4 || match[2] < 0 || match[3] < 0 {
			continue
		}
		target := body[match[2]:match[3]]
		if resolveLinkPath(docPath, target) == targetPath {
			count++
		}
	}
	return count
}

func rewriteMarkdownLinksForPlannedUpdate(body string, docPath string, oldResolvedPath string, newTarget string) (string, int) {
	matches := linkPattern.FindAllStringSubmatchIndex(body, -1)
	if len(matches) == 0 {
		return body, 0
	}
	var builder strings.Builder
	last := 0
	count := 0
	for _, match := range matches {
		if len(match) < 4 || match[2] < 0 || match[3] < 0 {
			continue
		}
		target := body[match[2]:match[3]]
		if resolveLinkPath(docPath, target) != oldResolvedPath {
			continue
		}
		builder.WriteString(body[last:match[2]])
		builder.WriteString(newTarget + markdownLinkFragment(target))
		last = match[3]
		count++
	}
	if count == 0 {
		return body, 0
	}
	builder.WriteString(body[last:])
	return builder.String(), count
}

func outgoingLinkUpdatesForMove(sourceDoc domain.Document, targetPath string) []domain.DocumentLinkUpdate {
	if path.Dir(sourceDoc.Path) == path.Dir(targetPath) {
		return nil
	}
	type updateTarget struct {
		newTarget   string
		occurrences int
	}
	targets := map[string]updateTarget{}
	matches := linkPattern.FindAllStringSubmatchIndex(sourceDoc.Body, -1)
	for _, match := range matches {
		if len(match) < 4 || match[2] < 0 || match[3] < 0 {
			continue
		}
		rawTarget := sourceDoc.Body[match[2]:match[3]]
		resolved := resolveLinkPath(sourceDoc.Path, rawTarget)
		if resolved == "" || resolved == sourceDoc.Path {
			continue
		}
		current := targets[resolved]
		current.newTarget = relativeMarkdownTarget(targetPath, resolved, "")
		current.occurrences++
		targets[resolved] = current
	}
	if len(targets) == 0 {
		return nil
	}
	resolvedPaths := make([]string, 0, len(targets))
	for resolved := range targets {
		resolvedPaths = append(resolvedPaths, resolved)
	}
	sort.Strings(resolvedPaths)
	updates := make([]domain.DocumentLinkUpdate, 0, len(resolvedPaths))
	for _, resolved := range resolvedPaths {
		target := targets[resolved]
		updates = append(updates, domain.DocumentLinkUpdate{
			DocID:       sourceDoc.DocID,
			Path:        sourceDoc.Path,
			OldTarget:   resolved,
			NewTarget:   target.newTarget,
			Occurrences: target.occurrences,
		})
	}
	return updates
}

func markdownLinkFragment(target string) string {
	if _, fragment, ok := strings.Cut(target, "#"); ok {
		return "#" + fragment
	}
	return ""
}

func relativeMarkdownTarget(fromDocPath string, targetPath string, fragment string) string {
	fromDir := path.Dir(fromDocPath)
	if fromDir == "." {
		fromDir = "."
	}
	rel, err := filepath.Rel(filepath.FromSlash(fromDir), filepath.FromSlash(targetPath))
	if err != nil {
		rel = targetPath
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		rel = path.Base(targetPath)
	}
	return rel + fragment
}

func ensureMarkdownFrontmatterID(body string, docID string) string {
	lines := strings.Split(body, "\n")
	if len(lines) >= 3 && strings.TrimSpace(lines[0]) == "---" {
		for idx := 1; idx < len(lines); idx++ {
			if strings.TrimSpace(lines[idx]) == "---" {
				updated := append([]string{}, lines[:idx]...)
				updated = append(updated, "id: "+frontmatterScalar(docID))
				updated = append(updated, lines[idx:]...)
				return strings.TrimRight(strings.Join(updated, "\n"), "\n") + "\n"
			}
		}
	}
	return "---\nid: " + frontmatterScalar(docID) + "\n---\n" + strings.TrimLeft(body, "\n")
}

func moveIndexUpdateCandidates(documents []domain.Document, sourcePath string, targetPath string, linkUpdates []domain.DocumentLinkUpdate, updateIndexes bool) []domain.DocumentIndexUpdate {
	documentByPath := documentsByPath(documents)
	linkUpdateByPath := map[string]struct{}{}
	for _, update := range linkUpdates {
		if update.IndexCandidate {
			linkUpdateByPath[update.Path] = struct{}{}
		}
	}
	candidates := []string{
		path.Join(path.Dir(sourcePath), "_index.md"),
		path.Join(path.Dir(targetPath), "_index.md"),
	}
	seen := map[string]struct{}{}
	result := []domain.DocumentIndexUpdate{}
	for _, candidate := range candidates {
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		status := "not_found"
		reason := "no _index.md document exists for this move directory"
		if _, ok := documentByPath[candidate]; ok {
			status = "candidate_only"
			reason = "index candidate reported; no index edit will be made unless it has a reported markdown link update"
			if _, ok := linkUpdateByPath[candidate]; ok {
				status = "link_update_planned"
				reason = "index contains a markdown link to the source path"
				if !updateIndexes {
					status = "link_update_planned_by_update_links"
					reason = "index link update is included in regular markdown link updates"
				}
			}
		}
		result = append(result, domain.DocumentIndexUpdate{Path: candidate, Status: status, Reason: reason})
	}
	return result
}

func planRefreshDocIDs(sourceDocID string, linkUpdates []domain.DocumentLinkUpdate) []string {
	seen := map[string]struct{}{sourceDocID: {}}
	for _, update := range linkUpdates {
		seen[update.DocID] = struct{}{}
	}
	docIDs := make([]string, 0, len(seen))
	for docID := range seen {
		docIDs = append(docIDs, docID)
	}
	sort.Strings(docIDs)
	return docIDs
}

func moveProjectionRefresh(docIDs []string) []domain.DocumentProjectionRefresh {
	refresh := make([]domain.DocumentProjectionRefresh, 0, len(docIDs))
	for _, docID := range docIDs {
		refresh = append(refresh, domain.DocumentProjectionRefresh{
			Projection: "graph",
			RefKind:    "document",
			RefID:      docID,
			Status:     "will_refresh_after_approved_move",
		})
	}
	return refresh
}

func moveValidationWarnings(sourceDoc domain.Document, targetPath string, duplicateRisk string, input domain.MoveDocumentInput, linkUpdates []domain.DocumentLinkUpdate, indexUpdates []domain.DocumentIndexUpdate) []string {
	warnings := []string{}
	if strings.TrimSpace(sourceDoc.Metadata["id"]) == "" {
		warnings = append(warnings, "source document has no id frontmatter; approved move will add id to preserve stable doc_id")
	}
	if duplicateRisk != moveDuplicateNone {
		warnings = append(warnings, "target path is not empty: "+duplicateRisk)
	}
	if !input.UpdateLinks {
		warnings = append(warnings, "update_links is false; inbound markdown links will not be rewritten")
	}
	if input.UpdateLinks && len(linkUpdates) == 0 {
		warnings = append(warnings, "no inbound markdown links to the source path were found")
	}
	if input.UpdateIndexes {
		hasIndexPlan := false
		for _, update := range indexUpdates {
			if strings.HasPrefix(update.Status, "link_update_planned") {
				hasIndexPlan = true
			}
		}
		if !hasIndexPlan {
			warnings = append(warnings, "update_indexes requested but no index markdown link update was found")
		}
	}
	if path.Ext(targetPath) != ".md" {
		warnings = append(warnings, "target path was normalized to markdown")
	}
	return warnings
}

func (s *Store) recordDocumentMoveProvenance(ctx context.Context, plan domain.DocumentMovePlan, linkUpdateCount int, frontmatterIDAdded bool) ([]string, error) {
	now := s.now().UTC()
	eventID := hashID("event", "document_moved", plan.DocID, plan.SourcePath, plan.TargetPath, now.Format(time.RFC3339Nano))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, domain.InternalError("begin document move provenance", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	details := map[string]string{
		"old_path":             plan.SourcePath,
		"new_path":             plan.TargetPath,
		"frontmatter_id":       plan.FrontmatterID,
		"frontmatter_id_added": strconv.FormatBool(frontmatterIDAdded),
		"link_updates_applied": strconv.Itoa(linkUpdateCount),
		"duplicate_risk":       plan.DuplicateRisk,
	}
	if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
		EventID:    eventID,
		EventType:  "document_moved",
		RefKind:    "document",
		RefID:      plan.DocID,
		SourceRef:  "doc:" + plan.DocID,
		OccurredAt: now,
		Details:    details,
	}); err != nil {
		return nil, domain.InternalError("record document move provenance event", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, domain.InternalError("commit document move provenance", err)
	}
	return []string{fmt.Sprintf("document:%s:document_moved", plan.DocID)}, nil
}
