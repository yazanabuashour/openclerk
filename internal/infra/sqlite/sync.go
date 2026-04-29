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

func (s *Store) syncVault(ctx context.Context) error {
	paths := make([]string, 0, 32)
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
	if err := s.pruneMissingDocuments(ctx, paths); err != nil {
		return err
	}
	for _, relPath := range paths {
		if err := s.syncDocumentFromDisk(ctx, relPath, ""); err != nil {
			return err
		}
	}
	if err := s.rebuildGraph(ctx); err != nil {
		return err
	}
	if err := s.rebuildRecords(ctx); err != nil {
		return err
	}
	if err := s.rebuildServices(ctx); err != nil {
		return err
	}
	if err := s.rebuildDecisions(ctx); err != nil {
		return err
	}
	return s.rebuildSynthesis(ctx)
}

func (s *Store) pruneMissingDocuments(ctx context.Context, livePaths []string) error {
	live := make(map[string]struct{}, len(livePaths))
	for _, relPath := range livePaths {
		live[relPath] = struct{}{}
	}

	rows, err := s.db.QueryContext(ctx, `SELECT doc_id, path FROM documents`)
	if err != nil {
		return domain.InternalError("query existing documents for pruning", err)
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
			return domain.InternalError("scan existing document for pruning", err)
		}
		if _, ok := live[path]; ok {
			continue
		}
		staleDocIDs = append(staleDocIDs, docID)
	}
	if err := rows.Err(); err != nil {
		return domain.InternalError("iterate existing documents for pruning", err)
	}
	if len(staleDocIDs) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin prune missing documents", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, docID := range staleDocIDs {
		if _, err := tx.ExecContext(ctx, `DELETE FROM chunk_fts WHERE doc_id = ?`, docID); err != nil {
			return domain.InternalError("delete missing document chunks from index", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM documents WHERE doc_id = ?`, docID); err != nil {
			return domain.InternalError("delete missing document", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit prune missing documents", err)
	}
	return nil
}

func (s *Store) syncDocumentFromDisk(ctx context.Context, relPath string, preferredTitle string) error {
	bodyBytes, err := osReadFile(filepath.Join(s.vaultRoot, filepath.FromSlash(relPath)))
	if err != nil {
		return domain.InternalError("read document from disk", err)
	}
	body := string(bodyBytes)
	headings, sections, frontmatter := parseMarkdown(body, relPath)
	docID := docIDForPath(relPath)
	now := s.now().UTC()
	headingsJSON, _ := json.Marshal(headings)
	metadataJSON, _ := json.Marshal(frontmatter)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin transaction", err)
	}
	defer func() {
		_ = tx.Rollback()
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
		return domain.InternalError("query existing document timestamp", err)
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
		updatedAt = updatedAtExisting
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
		return domain.InternalError("upsert document", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM document_metadata WHERE doc_id = ?`, docID); err != nil {
		return domain.InternalError("delete document metadata", err)
	}
	for key, value := range frontmatter {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO document_metadata (doc_id, key_name, value_text)
VALUES (?, ?, ?)`,
			docID,
			strings.ToLower(strings.TrimSpace(key)),
			strings.TrimSpace(value),
		); err != nil {
			return domain.InternalError("insert document metadata", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM chunk_fts WHERE doc_id = ?`, docID); err != nil {
		return domain.InternalError("delete indexed chunks", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM chunks WHERE doc_id = ?`, docID); err != nil {
		return domain.InternalError("delete chunks", err)
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
			return domain.InternalError("insert chunk", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO chunk_fts (chunk_id, doc_id, path, heading, content)
VALUES (?, ?, ?, ?, ?)`,
			chunkID,
			docID,
			relPath,
			sec.Heading,
			sec.Content,
		); err != nil {
			return domain.InternalError("insert chunk index", err)
		}
	}

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
			return domain.InternalError("record provenance event", err)
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
				return domain.InternalError("record source provenance event", err)
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
			return domain.InternalError("mark graph projection stale", err)
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
			return domain.InternalError("record graph invalidation event", err)
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
				return domain.InternalError("record records invalidation event", err)
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
				return domain.InternalError("record services invalidation event", err)
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
				return domain.InternalError("record decisions invalidation event", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit document sync", err)
	}
	if err := s.rebuildGraph(ctx); err != nil {
		return err
	}
	if err := s.rebuildRecords(ctx); err != nil {
		return err
	}
	if err := s.rebuildServices(ctx); err != nil {
		return err
	}
	if err := s.rebuildDecisions(ctx); err != nil {
		return err
	}
	return s.rebuildSynthesis(ctx)
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
