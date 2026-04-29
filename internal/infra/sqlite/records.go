package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"strings"
	"time"
)

func (s *Store) RecordsLookup(ctx context.Context, input domain.RecordLookupInput) (domain.RecordLookupResult, error) {
	if !supportsRecords(s.backend) {
		return domain.RecordLookupResult{}, domain.UnsupportedError("records extension", s.backend)
	}
	if strings.TrimSpace(input.Text) == "" {
		return domain.RecordLookupResult{}, domain.ValidationError("lookup text is required", nil)
	}
	limit := input.Limit
	if limit == 0 {
		limit = 10
	}
	offset := decodeCursor(input.Cursor)

	args := []any{"%" + strings.ToLower(strings.TrimSpace(input.Text)) + "%"}
	condition := "WHERE LOWER(name) LIKE ? OR LOWER(summary) LIKE ?"
	args = append(args, args[0])
	if input.EntityType != "" {
		condition = "WHERE (LOWER(name) LIKE ? OR LOWER(summary) LIKE ?) AND entity_type = ?"
		args = append(args, input.EntityType)
	}

	query := fmt.Sprintf(`
SELECT entity_id, entity_type, name, summary, updated_at
FROM record_entities
%s
ORDER BY name
LIMIT ? OFFSET ?`, condition)
	args = append(args, limit+1, offset)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.RecordLookupResult{}, domain.InternalError("query record entities", err)
	}

	entities := make([]domain.RecordEntity, 0, limit+1)
	for rows.Next() {
		var entity domain.RecordEntity
		var updatedAt string
		if err := rows.Scan(&entity.EntityID, &entity.EntityType, &entity.Name, &entity.Summary, &updatedAt); err != nil {
			return domain.RecordLookupResult{}, domain.InternalError("scan record entity", err)
		}
		entity.UpdatedAt = mustParseTime(updatedAt)
		entities = append(entities, entity)
	}
	if err := rows.Err(); err != nil {
		return domain.RecordLookupResult{}, domain.InternalError("iterate record entities", err)
	}
	if err := rows.Close(); err != nil {
		return domain.RecordLookupResult{}, domain.InternalError("close record entity rows", err)
	}
	for idx := range entities {
		loaded, err := s.loadRecordEntityDetails(ctx, entities[idx])
		if err != nil {
			return domain.RecordLookupResult{}, err
		}
		entities[idx] = loaded
	}
	pageInfo := domain.PageInfo{}
	if len(entities) > limit {
		pageInfo.HasMore = true
		pageInfo.NextCursor = encodeCursor(offset + limit)
		entities = entities[:limit]
	}
	return domain.RecordLookupResult{Entities: entities, PageInfo: pageInfo}, nil
}

func (s *Store) GetRecordEntity(ctx context.Context, entityID string) (domain.RecordEntity, error) {
	if !supportsRecords(s.backend) {
		return domain.RecordEntity{}, domain.UnsupportedError("records extension", s.backend)
	}
	var entity domain.RecordEntity
	var updatedAt string
	err := s.db.QueryRowContext(ctx, `
SELECT entity_id, entity_type, name, summary, updated_at
FROM record_entities
WHERE entity_id = ?`, entityID).Scan(
		&entity.EntityID,
		&entity.EntityType,
		&entity.Name,
		&entity.Summary,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RecordEntity{}, domain.NotFoundError("entity", entityID)
	}
	if err != nil {
		return domain.RecordEntity{}, domain.InternalError("query record entity", err)
	}
	entity.UpdatedAt = mustParseTime(updatedAt)
	return s.loadRecordEntityDetails(ctx, entity)
}

func (s *Store) loadRecordEntityDetails(ctx context.Context, entity domain.RecordEntity) (domain.RecordEntity, error) {
	factRows, err := s.db.QueryContext(ctx, `
SELECT key_name, value_text, observed_at
FROM record_facts
WHERE entity_id = ?
ORDER BY key_name`, entity.EntityID)
	if err != nil {
		return domain.RecordEntity{}, domain.InternalError("query record facts", err)
	}
	defer func() {
		_ = factRows.Close()
	}()
	for factRows.Next() {
		var (
			fact        domain.RecordFact
			observedRaw sql.NullString
		)
		if err := factRows.Scan(&fact.Key, &fact.Value, &observedRaw); err != nil {
			return domain.RecordEntity{}, domain.InternalError("scan record fact", err)
		}
		if observedRaw.Valid {
			observed := mustParseTime(observedRaw.String)
			fact.ObservedAt = &observed
		}
		entity.Facts = append(entity.Facts, fact)
	}
	if err := factRows.Err(); err != nil {
		return domain.RecordEntity{}, domain.InternalError("iterate record facts", err)
	}

	citationRows, err := s.db.QueryContext(ctx, `
SELECT source_doc_id, source_chunk_id, source_path, source_heading, source_line_start, source_line_end
FROM record_citations
WHERE entity_id = ?
ORDER BY source_doc_id, source_chunk_id`, entity.EntityID)
	if err != nil {
		return domain.RecordEntity{}, domain.InternalError("query record citations", err)
	}
	defer func() {
		_ = citationRows.Close()
	}()
	for citationRows.Next() {
		var (
			citation   domain.Citation
			headingRaw sql.NullString
		)
		if err := citationRows.Scan(
			&citation.DocID,
			&citation.ChunkID,
			&citation.Path,
			&headingRaw,
			&citation.LineStart,
			&citation.LineEnd,
		); err != nil {
			return domain.RecordEntity{}, domain.InternalError("scan record citation", err)
		}
		citation.Heading = headingRaw.String
		entity.Citations = append(entity.Citations, citation)
	}
	if err := citationRows.Err(); err != nil {
		return domain.RecordEntity{}, domain.InternalError("iterate record citations", err)
	}
	return entity, nil
}

func (s *Store) rebuildRecords(ctx context.Context) error {
	documents, err := s.loadAllDocuments(ctx)
	if err != nil {
		return err
	}
	chunksByDoc, err := s.loadChunksByDoc(ctx)
	if err != nil {
		return err
	}
	previousStates, err := s.loadProjectionStateSnapshots(ctx, "records")
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin records rebuild", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, stmt := range []string{
		`DELETE FROM record_citations;`,
		`DELETE FROM record_facts;`,
		`DELETE FROM record_entities;`,
		`DELETE FROM projection_states WHERE projection_name = 'records';`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return domain.InternalError("reset records projection", err)
		}
	}

	now := s.now().UTC()
	for _, doc := range documents {
		frontmatter, facts, ok := extractRecordProjection(doc.Body)
		if !ok {
			continue
		}
		entityType := frontmatter["entity_type"]
		name := frontmatter["entity_name"]
		entityID := frontmatter["entity_id"]
		if entityType == "" || name == "" {
			continue
		}
		if entityID == "" {
			entityID = hashID("entity", doc.DocID, name)
		}
		summary := firstSummaryParagraph(doc.Body)
		version := hashID("records", entityID, doc.UpdatedAt.UTC().Format(time.RFC3339Nano))
		entityUpdatedAt := now
		entityChanged := true
		if previous, ok := previousStates[entityID]; ok && previous.ProjectionVersion == version {
			entityUpdatedAt = previous.UpdatedAt
			entityChanged = false
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO record_entities (entity_id, entity_type, name, summary, source_doc_id, updated_at)
VALUES (?, ?, ?, ?, ?, ?)`,
			entityID,
			entityType,
			name,
			summary,
			doc.DocID,
			entityUpdatedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return domain.InternalError("insert record entity", err)
		}
		for _, fact := range facts {
			var observedAt *string
			if fact.ObservedAt != nil {
				value := fact.ObservedAt.UTC().Format(time.RFC3339Nano)
				observedAt = &value
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO record_facts (entity_id, key_name, value_text, observed_at)
VALUES (?, ?, ?, ?)`,
				entityID,
				fact.Key,
				fact.Value,
				observedAt,
			); err != nil {
				return domain.InternalError("insert record fact", err)
			}
		}
		citation := documentCitation(doc, chunksByDoc[doc.DocID])
		if _, err := tx.ExecContext(ctx, `
INSERT INTO record_citations (entity_id, source_doc_id, source_chunk_id, source_path, source_heading, source_line_start, source_line_end)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
			entityID,
			citation.DocID,
			citation.ChunkID,
			citation.Path,
			nullIfEmpty(citation.Heading),
			citation.LineStart,
			citation.LineEnd,
		); err != nil {
			return domain.InternalError("insert record citation", err)
		}
		if entityChanged {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO provenance_events (event_id, event_type, ref_kind, ref_id, source_ref, occurred_at, details_json)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
				hashID("event", "record", entityID, now.Format(time.RFC3339Nano)),
				"record_extracted_from_doc",
				"entity",
				entityID,
				"doc:"+doc.DocID,
				now.Format(time.RFC3339Nano),
				fmt.Sprintf(`{"entity_type":%q,"entity_name":%q}`, entityType, name),
			); err != nil {
				return domain.InternalError("record records provenance event", err)
			}
		}
		if err := upsertProjectionState(ctx, tx, domain.ProjectionState{
			Projection:        "records",
			RefKind:           "entity",
			RefID:             entityID,
			SourceRef:         "doc:" + doc.DocID,
			Freshness:         "fresh",
			ProjectionVersion: version,
			UpdatedAt:         entityUpdatedAt,
			Details: map[string]string{
				"entity_type": entityType,
				"path":        doc.Path,
			},
		}); err != nil {
			return err
		}
		if entityChanged {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "projection_refreshed", "records", entityID, version, now.Format(time.RFC3339Nano)),
				EventType:  "projection_refreshed",
				RefKind:    "projection",
				RefID:      "records:" + entityID,
				SourceRef:  "doc:" + doc.DocID,
				OccurredAt: now,
				Details: map[string]string{
					"projection":  "records",
					"entity_type": entityType,
					"version":     version,
				},
			}); err != nil {
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit records rebuild", err)
	}
	return nil
}

func supportsRecords(backend domain.BackendKind) bool {
	return backend == domain.BackendOpenClerk
}

func extractRecordProjection(body string) (map[string]string, []domain.RecordFact, bool) {
	lines := strings.Split(body, "\n")
	frontmatter, contentStart := parseFrontmatter(lines)
	if frontmatter["entity_type"] == "" && frontmatter["entity_name"] == "" {
		return nil, nil, false
	}
	return frontmatter, extractRecordFacts(lines, contentStart), true
}

func extractRecordFacts(lines []string, contentStart int) []domain.RecordFact {
	facts := []domain.RecordFact{}
	inFacts := false
	for idx := contentStart; idx < len(lines); idx++ {
		line := strings.TrimSpace(lines[idx])
		if line == "" {
			continue
		}
		if matches := headingPattern.FindStringSubmatch(line); len(matches) > 0 {
			inFacts = strings.EqualFold(strings.TrimSpace(matches[2]), "Facts")
			continue
		}
		if !inFacts {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		facts = append(facts, domain.RecordFact{
			Key:   strings.TrimSpace(key),
			Value: strings.TrimSpace(value),
		})
	}
	return facts
}

func documentCitation(doc domain.Document, chunks []domain.Chunk) domain.Citation {
	if len(chunks) == 0 {
		return domain.Citation{
			DocID:     doc.DocID,
			ChunkID:   "",
			Path:      doc.Path,
			Heading:   doc.Title,
			LineStart: 1,
			LineEnd:   1,
		}
	}
	chunk := chunks[0]
	return domain.Citation{
		DocID:     chunk.DocID,
		ChunkID:   chunk.ChunkID,
		Path:      chunk.Path,
		Heading:   chunk.Heading,
		LineStart: chunk.LineStart,
		LineEnd:   chunk.LineEnd,
	}
}
