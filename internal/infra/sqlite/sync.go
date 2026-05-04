package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type documentSyncOptions struct {
	RebuildProjections bool
	DeferFTS           bool
	Diagnostics        *SyncDiagnostics
	Tx                 *sql.Tx
}

type documentSyncResult struct {
	Created        bool
	Updated        bool
	Unchanged      bool
	BytesRead      int64
	ChunksWritten  int
	FTSRowsWritten int
}

const (
	syncVaultDocumentBatchSize = 200

	ftsStrategyPending          = "pending"
	ftsStrategyBulkRebuild      = "bulk_rebuild"
	ftsStrategyIncrementalRows  = "incremental_rows"
	ftsStrategySkippedNoChanges = "skipped_no_changes"
)

func (s *Store) syncVault(ctx context.Context) error {
	totalStart := time.Now()
	diagnostics := newSyncDiagnostics()

	paths := make([]string, 0, 32)
	scanStart := time.Now()
	err := filepath.WalkDir(s.vaultRoot, func(absPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(absPath) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(s.vaultRoot, absPath)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return domain.InternalError("scan vault root", err)
	}
	sort.Strings(paths)
	diagnostics.ScanSeconds = syncSecondsSince(scanStart)
	diagnostics.PathsScanned = len(paths)
	diagnostics.LastPhase = "scan_complete"
	diagnostics.TotalSeconds = syncSecondsSince(totalStart)
	if err := writeSyncDiagnostics(s.syncDiagnosticsPath, diagnostics); err != nil {
		return err
	}

	pruneStart := time.Now()
	diagnostics.LastPhase = "prune"
	pruned, err := s.pruneMissingDocuments(ctx, paths)
	diagnostics.PruneSeconds = syncSecondsSince(pruneStart)
	diagnostics.DocumentsPruned = pruned
	if err != nil {
		return err
	}
	diagnostics.LastPhase = "document_import"
	diagnostics.TotalSeconds = syncSecondsSince(totalStart)
	if err := writeSyncDiagnostics(s.syncDiagnosticsPath, diagnostics); err != nil {
		return err
	}
	for start := 0; start < len(paths); start += syncVaultDocumentBatchSize {
		end := start + syncVaultDocumentBatchSize
		if end > len(paths) {
			end = len(paths)
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return domain.InternalError("begin vault sync batch", err)
		}
		for _, relPath := range paths[start:end] {
			if _, err := s.syncDocumentFromDiskWithOptions(ctx, relPath, "", documentSyncOptions{
				RebuildProjections: false,
				DeferFTS:           true,
				Diagnostics:        &diagnostics,
				Tx:                 tx,
			}); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return domain.InternalError("commit vault sync batch", err)
		}
		diagnostics.TotalSeconds = syncSecondsSince(totalStart)
		if err := writeSyncDiagnostics(s.syncDiagnosticsPath, diagnostics); err != nil {
			return err
		}
	}

	diagnostics.LastPhase = "fts_rebuild_check"
	diagnostics.TotalSeconds = syncSecondsSince(totalStart)
	if err := writeSyncDiagnostics(s.syncDiagnosticsPath, diagnostics); err != nil {
		return err
	}
	if err := s.rebuildChunkFTSIfNeeded(ctx, &diagnostics, totalStart); err != nil {
		return err
	}

	shouldRebuild := diagnostics.changedDocuments() > 0
	if !shouldRebuild {
		bootstrap, err := s.needsProjectionBootstrap(ctx)
		if err != nil {
			return err
		}
		diagnostics.ProjectionBootstrap = bootstrap
		shouldRebuild = bootstrap
	}
	if shouldRebuild {
		diagnostics.LastPhase = "projection_rebuild"
		diagnostics.TotalSeconds = syncSecondsSince(totalStart)
		if err := writeSyncDiagnostics(s.syncDiagnosticsPath, diagnostics); err != nil {
			return err
		}
		if err := s.rebuildAllProjections(ctx, &diagnostics); err != nil {
			return err
		}
		if err := s.clearProjectionRebuildPending(ctx); err != nil {
			return err
		}
	} else {
		diagnostics.ProjectionRebuildSkipped = true
	}

	diagnostics.Status = "completed"
	diagnostics.LastPhase = "completed"
	diagnostics.TotalSeconds = syncSecondsSince(totalStart)
	return writeSyncDiagnostics(s.syncDiagnosticsPath, diagnostics)
}

func (s *Store) pruneMissingDocuments(ctx context.Context, livePaths []string) (int, error) {
	live := make(map[string]struct{}, len(livePaths))
	for _, relPath := range livePaths {
		live[relPath] = struct{}{}
	}

	rows, err := s.db.QueryContext(ctx, `SELECT doc_id, path FROM documents`)
	if err != nil {
		return 0, domain.InternalError("query existing documents for pruning", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	staleDocIDs := make([]string, 0, 8)
	for rows.Next() {
		var (
			docID string
			path  string
		)
		if err := rows.Scan(&docID, &path); err != nil {
			return 0, domain.InternalError("scan existing document for pruning", err)
		}
		if _, ok := live[path]; ok {
			continue
		}
		staleDocIDs = append(staleDocIDs, docID)
	}
	if err := rows.Err(); err != nil {
		return 0, domain.InternalError("iterate existing documents for pruning", err)
	}
	if len(staleDocIDs) == 0 {
		return 0, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, domain.InternalError("begin prune missing documents", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := s.markProjectionRebuildPending(ctx, tx, s.now().UTC()); err != nil {
		return 0, err
	}
	if err := s.markFTSRebuildPending(ctx, tx, s.now().UTC()); err != nil {
		return 0, err
	}
	for _, docID := range staleDocIDs {
		if _, err := tx.ExecContext(ctx, `DELETE FROM documents WHERE doc_id = ?`, docID); err != nil {
			return 0, domain.InternalError("delete missing document", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, domain.InternalError("commit prune missing documents", err)
	}
	return len(staleDocIDs), nil
}

func (s *Store) syncDocumentFromDisk(ctx context.Context, relPath string, preferredTitle string) error {
	_, err := s.syncDocumentFromDiskWithOptions(ctx, relPath, preferredTitle, documentSyncOptions{
		RebuildProjections: true,
	})
	return err
}

func (s *Store) syncDocumentFromDiskWithOptions(ctx context.Context, relPath string, preferredTitle string, options documentSyncOptions) (documentSyncResult, error) {
	readParseStart := time.Now()
	bodyBytes, err := osReadFile(filepath.Join(s.vaultRoot, filepath.FromSlash(relPath)))
	if err != nil {
		return documentSyncResult{}, domain.InternalError("read document from disk", err)
	}
	body := string(bodyBytes)
	headings, sections, frontmatter := parseMarkdown(body, relPath)
	docID := docIDForPath(relPath)
	now := s.now().UTC()
	headingsJSON, _ := json.Marshal(headings)
	metadataJSON, _ := json.Marshal(frontmatter)
	result := documentSyncResult{
		BytesRead: int64(len(bodyBytes)),
	}
	if options.Diagnostics != nil {
		options.Diagnostics.BytesRead += result.BytesRead
		options.Diagnostics.DocumentReadParseSeconds += syncSecondsSince(readParseStart)
	}

	tx := options.Tx
	ownTx := tx == nil
	if ownTx {
		var err error
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return documentSyncResult{}, domain.InternalError("begin transaction", err)
		}
	}
	defer func() {
		if ownTx {
			_ = tx.Rollback()
		}
	}()

	createdAt := now.Format(time.RFC3339Nano)
	updatedAt := now.Format(time.RFC3339Nano)
	var existingTitle string
	var existingBody string
	var existingHeadingsJSON string
	var existingMetadataJSON string
	var createdAtExisting string
	var updatedAtExisting string
	eventType := "document_created"
	err = tx.QueryRowContext(ctx, `
SELECT created_at, updated_at, title, body, headings_json, metadata_json
FROM documents
WHERE doc_id = ?`, docID).Scan(
		&createdAtExisting,
		&updatedAtExisting,
		&existingTitle,
		&existingBody,
		&existingHeadingsJSON,
		&existingMetadataJSON,
	)
	if err == nil {
		createdAt = createdAtExisting
		eventType = "document_updated"
	} else if !errors.Is(err, sql.ErrNoRows) {
		return documentSyncResult{}, domain.InternalError("query existing document timestamp", err)
	}
	previousFrontmatter := map[string]string{}
	if eventType == "document_updated" {
		_ = json.Unmarshal([]byte(existingMetadataJSON), &previousFrontmatter)
	}

	title := resolvedDocumentTitle(relPath, body, headings, frontmatter, preferredTitle, existingTitle)
	contentChanged := eventType == "document_created" ||
		existingTitle != title ||
		existingBody != body ||
		existingHeadingsJSON != string(headingsJSON) ||
		existingMetadataJSON != string(metadataJSON)
	if !contentChanged {
		result.Unchanged = true
		if options.Diagnostics != nil {
			options.Diagnostics.DocumentsUnchanged++
		}
		return result, nil
	}
	writeStart := time.Now()
	recordWriteStart := time.Now()
	if err := s.markProjectionRebuildPending(ctx, tx, now); err != nil {
		return documentSyncResult{}, err
	}
	if options.DeferFTS {
		if err := s.markFTSRebuildPending(ctx, tx, now); err != nil {
			return documentSyncResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
	INSERT INTO documents (doc_id, path, title, body, headings_json, metadata_json, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(doc_id) DO UPDATE SET
	path = excluded.path,
	title = excluded.title,
	body = excluded.body,
	headings_json = excluded.headings_json,
	metadata_json = excluded.metadata_json,
	updated_at = excluded.updated_at`,
		docID,
		relPath,
		title,
		body,
		string(headingsJSON),
		string(metadataJSON),
		createdAt,
		updatedAt,
	); err != nil {
		return documentSyncResult{}, domain.InternalError("upsert document", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM document_metadata WHERE doc_id = ?`, docID); err != nil {
		return documentSyncResult{}, domain.InternalError("delete document metadata", err)
	}
	for key, value := range frontmatter {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO document_metadata (doc_id, key_name, value_text)
VALUES (?, ?, ?)`,
			docID,
			strings.ToLower(strings.TrimSpace(key)),
			strings.TrimSpace(value),
		); err != nil {
			return documentSyncResult{}, domain.InternalError("insert document metadata", err)
		}
	}
	if options.Diagnostics != nil {
		options.Diagnostics.DocumentRecordWriteSeconds += syncSecondsSince(recordWriteStart)
	}

	if !options.DeferFTS {
		ftsWriteStart := time.Now()
		if _, err := tx.ExecContext(ctx, `DELETE FROM chunk_fts WHERE doc_id = ?`, docID); err != nil {
			return documentSyncResult{}, domain.InternalError("delete indexed chunks", err)
		}
		if options.Diagnostics != nil {
			options.Diagnostics.FTSStrategy = ftsStrategyIncrementalRows
			options.Diagnostics.IncrementalFTSWriteSeconds += syncSecondsSince(ftsWriteStart)
		}
	}

	chunkWriteStart := time.Now()
	if _, err := tx.ExecContext(ctx, `DELETE FROM chunks WHERE doc_id = ?`, docID); err != nil {
		return documentSyncResult{}, domain.InternalError("delete chunks", err)
	}
	for _, sec := range sections {
		chunkID := chunkIDForSection(docID, sec)
		if _, err := tx.ExecContext(ctx, `
INSERT INTO chunks (chunk_id, doc_id, path, heading, content, line_start, line_end)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
			chunkID,
			docID,
			relPath,
			sec.Heading,
			sec.Content,
			sec.LineStart,
			sec.LineEnd,
		); err != nil {
			return documentSyncResult{}, domain.InternalError("insert chunk", err)
		}
		result.ChunksWritten++
	}
	if options.Diagnostics != nil {
		options.Diagnostics.ChunkWriteSeconds += syncSecondsSince(chunkWriteStart)
	}
	if !options.DeferFTS {
		ftsWriteStart := time.Now()
		for _, sec := range sections {
			chunkID := chunkIDForSection(docID, sec)
			if _, err := tx.ExecContext(ctx, `
	INSERT INTO chunk_fts (chunk_id, doc_id, path, heading, content)
	VALUES (?, ?, ?, ?, ?)`,
				chunkID,
				docID,
				relPath,
				sec.Heading,
				sec.Content,
			); err != nil {
				return documentSyncResult{}, domain.InternalError("insert chunk index", err)
			}
			result.FTSRowsWritten++
		}
		if options.Diagnostics != nil {
			options.Diagnostics.FTSStrategy = ftsStrategyIncrementalRows
			options.Diagnostics.IncrementalFTSWriteSeconds += syncSecondsSince(ftsWriteStart)
		}
	}

	provenanceWriteStart := time.Now()
	contentVersion := hashID("document-version", relPath, title, body, string(headingsJSON), string(metadataJSON))
	if contentChanged {
		if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
			EventID:    hashID("event", eventType, relPath, now.Format(time.RFC3339Nano), contentVersion),
			EventType:  eventType,
			RefKind:    "document",
			RefID:      docID,
			SourceRef:  "doc:" + docID,
			OccurredAt: now,
			Details: map[string]string{
				"path": relPath,
			},
		}); err != nil {
			return documentSyncResult{}, domain.InternalError("record provenance event", err)
		}
		if !isSynthesisDocument(relPath, frontmatter) {
			sourceEventType := "source_created"
			if eventType == "document_updated" {
				sourceEventType = "source_updated"
			}
			sourceDetails := sourceProvenanceDetails(relPath, frontmatter)
			if sourceEventType == "source_updated" {
				sourceDetails = sourceUpdateProvenanceDetails(relPath, frontmatter, previousFrontmatter)
			}
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", sourceEventType, relPath, now.Format(time.RFC3339Nano), contentVersion),
				EventType:  sourceEventType,
				RefKind:    "source",
				RefID:      docID,
				SourceRef:  "doc:" + docID,
				OccurredAt: now,
				Details:    sourceDetails,
			}); err != nil {
				return documentSyncResult{}, domain.InternalError("record source provenance event", err)
			}
		}
	}
	if contentChanged && supportsGraph(s.backend) {
		if err := upsertProjectionState(ctx, tx, domain.ProjectionState{
			Projection:        "graph",
			RefKind:           "document",
			RefID:             docID,
			SourceRef:         "doc:" + docID,
			Freshness:         "stale",
			ProjectionVersion: hashID("graph", docID, "stale", now.Format(time.RFC3339Nano)),
			UpdatedAt:         now,
			Details: map[string]string{
				"path": relPath,
			},
		}); err != nil {
			return documentSyncResult{}, domain.InternalError("mark graph projection stale", err)
		}
		if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
			EventID:    hashID("event", "projection_invalidated", "graph", docID, now.Format(time.RFC3339Nano), contentVersion),
			EventType:  "projection_invalidated",
			RefKind:    "projection",
			RefID:      "graph:" + docID,
			SourceRef:  "doc:" + docID,
			OccurredAt: now,
			Details: map[string]string{
				"projection": "graph",
				"path":       relPath,
			},
		}); err != nil {
			return documentSyncResult{}, domain.InternalError("record graph invalidation event", err)
		}
	}
	if contentChanged && supportsRecords(s.backend) {
		_, _, projectsRecords := extractRecordProjection(body)
		_, _, projectedRecords := extractRecordProjection(existingBody)
		if projectsRecords || projectedRecords {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "projection_invalidated", "records", docID, now.Format(time.RFC3339Nano), contentVersion),
				EventType:  "projection_invalidated",
				RefKind:    "projection",
				RefID:      "records-source:" + docID,
				SourceRef:  "doc:" + docID,
				OccurredAt: now,
				Details: map[string]string{
					"projection": "records",
					"path":       relPath,
				},
			}); err != nil {
				return documentSyncResult{}, domain.InternalError("record records invalidation event", err)
			}
		}
	}
	if contentChanged && supportsServices(s.backend) {
		_, projectsServices := extractServiceProjection(body)
		_, projectedServices := extractServiceProjection(existingBody)
		if projectsServices || projectedServices {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "projection_invalidated", "services", docID, now.Format(time.RFC3339Nano), contentVersion),
				EventType:  "projection_invalidated",
				RefKind:    "projection",
				RefID:      "services-source:" + docID,
				SourceRef:  "doc:" + docID,
				OccurredAt: now,
				Details: map[string]string{
					"projection": "services",
					"path":       relPath,
				},
			}); err != nil {
				return documentSyncResult{}, domain.InternalError("record services invalidation event", err)
			}
		}
	}
	if contentChanged && supportsDecisions(s.backend) {
		_, projectsDecisions := extractDecisionProjection(body)
		_, projectedDecisions := extractDecisionProjection(existingBody)
		if projectsDecisions || projectedDecisions {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "projection_invalidated", "decisions", docID, now.Format(time.RFC3339Nano), contentVersion),
				EventType:  "projection_invalidated",
				RefKind:    "projection",
				RefID:      "decisions-source:" + docID,
				SourceRef:  "doc:" + docID,
				OccurredAt: now,
				Details: map[string]string{
					"projection": "decisions",
					"path":       relPath,
				},
			}); err != nil {
				return documentSyncResult{}, domain.InternalError("record decisions invalidation event", err)
			}
		}
	}
	if options.Diagnostics != nil {
		options.Diagnostics.ProvenanceWriteSeconds += syncSecondsSince(provenanceWriteStart)
	}

	if eventType == "document_created" {
		result.Created = true
		if options.Diagnostics != nil {
			options.Diagnostics.DocumentsCreated++
		}
	} else {
		result.Updated = true
		if options.Diagnostics != nil {
			options.Diagnostics.DocumentsUpdated++
		}
	}
	if options.Diagnostics != nil {
		options.Diagnostics.ChunksWritten += result.ChunksWritten
		options.Diagnostics.FTSRowsWritten += result.FTSRowsWritten
		options.Diagnostics.DocumentWriteSeconds += syncSecondsSince(writeStart)
	}
	if ownTx {
		if err := tx.Commit(); err != nil {
			return documentSyncResult{}, domain.InternalError("commit document sync", err)
		}
	}
	if ownTx && options.RebuildProjections {
		if err := s.rebuildAllProjections(ctx, options.Diagnostics); err != nil {
			return documentSyncResult{}, err
		}
		if err := s.clearProjectionRebuildPending(ctx); err != nil {
			return documentSyncResult{}, err
		}
	}
	return result, nil
}

func (s *Store) rebuildChunkFTSIfNeeded(ctx context.Context, diagnostics *SyncDiagnostics, totalStart time.Time) error {
	pending, err := s.ftsRebuildPending(ctx)
	if err != nil {
		return err
	}
	chunkRows, ftsRows, err := s.chunkFTSCounts(ctx)
	if err != nil {
		return err
	}
	countMismatch := chunkRows != ftsRows
	changedDocuments := 0
	if diagnostics != nil {
		changedDocuments = diagnostics.changedDocuments()
		diagnostics.FTSRebuildPending = pending
		diagnostics.FTSBootstrap = pending || countMismatch
	}
	if changedDocuments == 0 && !pending && !countMismatch {
		if diagnostics != nil {
			diagnostics.FTSStrategy = ftsStrategySkippedNoChanges
			diagnostics.FTSRebuildSkipped = true
		}
		return nil
	}
	if diagnostics != nil {
		diagnostics.FTSStrategy = ftsStrategyBulkRebuild
		diagnostics.FTSRebuildSkipped = false
		diagnostics.LastPhase = "fts_rebuild"
		diagnostics.TotalSeconds = syncSecondsSince(totalStart)
		if err := writeSyncDiagnostics(s.syncDiagnosticsPath, *diagnostics); err != nil {
			return err
		}
	}
	return s.rebuildChunkFTS(ctx, diagnostics)
}

func (s *Store) rebuildChunkFTS(ctx context.Context, diagnostics *SyncDiagnostics) error {
	var chunkRows int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chunks`).Scan(&chunkRows); err != nil {
		return domain.InternalError("count chunks for FTS rebuild", err)
	}
	start := time.Now()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin FTS rebuild", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.ExecContext(ctx, `DELETE FROM chunk_fts`); err != nil {
		return domain.InternalError("delete chunk FTS rows", err)
	}
	if _, err := tx.ExecContext(ctx, `
	INSERT INTO chunk_fts (chunk_id, doc_id, path, heading, content)
	SELECT chunk_id, doc_id, path, heading, content
	FROM chunks`); err != nil {
		return domain.InternalError("bulk rebuild chunk FTS", err)
	}
	if err := s.clearFTSRebuildPendingTx(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit FTS rebuild", err)
	}
	if diagnostics != nil {
		diagnostics.FTSRowsWritten += chunkRows
		diagnostics.BulkFTSRebuildSeconds += syncSecondsSince(start)
		diagnostics.FTSRebuildPending = false
	}
	return nil
}

func (s *Store) chunkFTSCounts(ctx context.Context) (int, int, error) {
	var chunkRows int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chunks`).Scan(&chunkRows); err != nil {
		return 0, 0, domain.InternalError("count chunks for FTS bootstrap", err)
	}
	var ftsRows int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chunk_fts`).Scan(&ftsRows); err != nil {
		return 0, 0, domain.InternalError("count FTS rows for bootstrap", err)
	}
	return chunkRows, ftsRows, nil
}

func (s *Store) rebuildAllProjections(ctx context.Context, diagnostics *SyncDiagnostics) error {
	rebuilders := []struct {
		name      string
		supported bool
		rebuild   func(context.Context) error
	}{
		{name: "graph", supported: supportsGraph(s.backend), rebuild: s.rebuildGraph},
		{name: "records", supported: supportsRecords(s.backend), rebuild: s.rebuildRecords},
		{name: "services", supported: supportsServices(s.backend), rebuild: s.rebuildServices},
		{name: "decisions", supported: supportsDecisions(s.backend), rebuild: s.rebuildDecisions},
		{name: "synthesis", supported: true, rebuild: s.rebuildSynthesis},
	}
	for _, rebuilder := range rebuilders {
		if !rebuilder.supported {
			continue
		}
		start := time.Now()
		if err := rebuilder.rebuild(ctx); err != nil {
			return err
		}
		seconds := syncSecondsSince(start)
		if diagnostics != nil {
			diagnostics.ProjectionRebuildSeconds += seconds
			diagnostics.ProjectionRebuilds = append(diagnostics.ProjectionRebuilds, ProjectionRebuildDiagnostics{
				Projection: rebuilder.name,
				Seconds:    seconds,
			})
		}
	}
	return nil
}

func (s *Store) needsProjectionBootstrap(ctx context.Context) (bool, error) {
	pending, err := s.projectionRebuildPending(ctx)
	if err != nil {
		return false, err
	}
	if pending {
		return true, nil
	}
	var documentCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&documentCount); err != nil {
		return false, domain.InternalError("count documents for projection bootstrap", err)
	}
	if documentCount == 0 {
		return false, nil
	}
	var projectionCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projection_states`).Scan(&projectionCount); err != nil {
		return false, domain.InternalError("count projection states for bootstrap", err)
	}
	if projectionCount == 0 {
		return true, nil
	}
	if supportsGraph(s.backend) {
		graphNeedsBootstrap, err := s.graphProjectionNeedsBootstrap(ctx, documentCount)
		if err != nil {
			return false, err
		}
		if graphNeedsBootstrap {
			return true, nil
		}
	}
	synthesisNeedsBootstrap, err := s.synthesisProjectionNeedsBootstrap(ctx)
	if err != nil {
		return false, err
	}
	return synthesisNeedsBootstrap, nil
}

func (s *Store) markProjectionRebuildPending(ctx context.Context, tx *sql.Tx, now time.Time) error {
	return upsertRuntimeConfigValueTx(ctx, tx, configKeyProjectionRebuildPending, "true", now.UTC().Format(time.RFC3339Nano))
}

func (s *Store) markFTSRebuildPending(ctx context.Context, tx *sql.Tx, now time.Time) error {
	return upsertRuntimeConfigValueTx(ctx, tx, configKeyFTSRebuildPending, "true", now.UTC().Format(time.RFC3339Nano))
}

func (s *Store) clearProjectionRebuildPending(ctx context.Context) error {
	return upsertRuntimeConfigValue(ctx, s.db, configKeyProjectionRebuildPending, "false", s.now().UTC().Format(time.RFC3339Nano))
}

func (s *Store) clearFTSRebuildPendingTx(ctx context.Context, tx *sql.Tx) error {
	return upsertRuntimeConfigValueTx(ctx, tx, configKeyFTSRebuildPending, "false", s.now().UTC().Format(time.RFC3339Nano))
}

func (s *Store) projectionRebuildPending(ctx context.Context) (bool, error) {
	value, err := runtimeConfigValue(ctx, s.db, configKeyProjectionRebuildPending)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(value), "true"), nil
}

func (s *Store) ftsRebuildPending(ctx context.Context) (bool, error) {
	value, err := runtimeConfigValue(ctx, s.db, configKeyFTSRebuildPending)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(value), "true"), nil
}

func (s *Store) graphProjectionNeedsBootstrap(ctx context.Context, documentCount int) (bool, error) {
	var graphCount int
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM projection_states
WHERE projection_name = 'graph'
	AND ref_kind = 'document'`).Scan(&graphCount); err != nil {
		return false, domain.InternalError("count graph projection states for bootstrap", err)
	}
	if graphCount != documentCount {
		return true, nil
	}
	var staleGraphCount int
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM projection_states
WHERE projection_name = 'graph'
	AND freshness = 'stale'`).Scan(&staleGraphCount); err != nil {
		return false, domain.InternalError("count stale graph projection states for bootstrap", err)
	}
	return staleGraphCount > 0, nil
}

func (s *Store) synthesisProjectionNeedsBootstrap(ctx context.Context) (bool, error) {
	var synthesisDocuments int
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM documents AS d
JOIN document_metadata AS m ON m.doc_id = d.doc_id
WHERE d.path LIKE 'synthesis/%'
	AND m.key_name = 'type'
	AND lower(m.value_text) = 'synthesis'`).Scan(&synthesisDocuments); err != nil {
		return false, domain.InternalError("count synthesis documents for projection bootstrap", err)
	}
	var synthesisProjections int
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM projection_states
WHERE projection_name = 'synthesis'`).Scan(&synthesisProjections); err != nil {
		return false, domain.InternalError("count synthesis projection states for bootstrap", err)
	}
	return synthesisDocuments != synthesisProjections, nil
}

func (s *Store) loadAllDocuments(ctx context.Context) ([]domain.Document, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT doc_id, path, title, body, headings_json, metadata_json, created_at, updated_at
FROM documents
ORDER BY path`)
	if err != nil {
		return nil, domain.InternalError("query documents", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	docs := []domain.Document{}
	for rows.Next() {
		var (
			doc          domain.Document
			headingsJSON string
			metadataJSON string
			createdAt    string
			updatedAt    string
		)
		if err := rows.Scan(&doc.DocID, &doc.Path, &doc.Title, &doc.Body, &headingsJSON, &metadataJSON, &createdAt, &updatedAt); err != nil {
			return nil, domain.InternalError("scan document", err)
		}
		_ = json.Unmarshal([]byte(headingsJSON), &doc.Headings)
		_ = json.Unmarshal([]byte(metadataJSON), &doc.Metadata)
		doc.CreatedAt = mustParseTime(createdAt)
		doc.UpdatedAt = mustParseTime(updatedAt)
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError("iterate documents", err)
	}
	return docs, nil
}

func (s *Store) loadChunksByDoc(ctx context.Context) (map[string][]domain.Chunk, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT chunk_id, doc_id, path, heading, content, line_start, line_end
FROM chunks
ORDER BY doc_id, line_start`)
	if err != nil {
		return nil, domain.InternalError("query chunks", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	result := map[string][]domain.Chunk{}
	for rows.Next() {
		var chunk domain.Chunk
		if err := rows.Scan(&chunk.ChunkID, &chunk.DocID, &chunk.Path, &chunk.Heading, &chunk.Content, &chunk.LineStart, &chunk.LineEnd); err != nil {
			return nil, domain.InternalError("scan chunk", err)
		}
		result[chunk.DocID] = append(result[chunk.DocID], chunk)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError("iterate chunks", err)
	}
	return result, nil
}
