package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"sort"
	"strings"
	"time"
)

type decisionProjection struct {
	DecisionID   string
	Title        string
	Status       string
	Scope        string
	Owner        string
	Date         string
	Supersedes   []string
	SupersededBy []string
	SourceRefs   []string
}

func (s *Store) DecisionsLookup(ctx context.Context, input domain.DecisionLookupInput) (domain.DecisionLookupResult, error) {
	if !supportsDecisions(s.backend) {
		return domain.DecisionLookupResult{}, domain.UnsupportedError("decisions extension", s.backend)
	}
	limit := input.Limit
	if limit == 0 {
		limit = 10
	}
	if limit < 1 || limit > 100 {
		return domain.DecisionLookupResult{}, domain.ValidationError("limit must be between 1 and 100", map[string]any{"limit": limit})
	}
	offset := decodeCursor(input.Cursor)

	args := []any{}
	clauses := []string{}
	if text := strings.ToLower(strings.TrimSpace(input.Text)); text != "" {
		clauses = append(clauses, "(LOWER(decision_id) LIKE ? OR LOWER(title) LIKE ? OR LOWER(status) LIKE ? OR LOWER(scope) LIKE ? OR LOWER(owner) LIKE ? OR LOWER(summary) LIKE ? OR LOWER(supersedes) LIKE ? OR LOWER(superseded_by) LIKE ? OR LOWER(source_refs) LIKE ?)")
		pattern := "%" + text + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		clauses = append(clauses, "LOWER(status) = ?")
		args = append(args, strings.ToLower(status))
	}
	if scope := strings.TrimSpace(input.Scope); scope != "" {
		clauses = append(clauses, "LOWER(scope) = ?")
		args = append(args, strings.ToLower(scope))
	}
	if owner := strings.TrimSpace(input.Owner); owner != "" {
		clauses = append(clauses, "LOWER(owner) = ?")
		args = append(args, strings.ToLower(owner))
	}

	query := `
SELECT decision_id, title, status, scope, owner, decision_date, summary, supersedes, superseded_by, source_refs, updated_at
FROM decision_records`
	if len(clauses) > 0 {
		query += "\nWHERE " + strings.Join(clauses, " AND ")
	}
	query += `
ORDER BY title
LIMIT ? OFFSET ?`
	args = append(args, limit+1, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.DecisionLookupResult{}, domain.InternalError("query decision records", err)
	}
	decisions := make([]domain.DecisionRecord, 0, limit+1)
	for rows.Next() {
		var decision domain.DecisionRecord
		var supersedesRaw, supersededByRaw, sourceRefsRaw, updatedAt string
		if err := rows.Scan(
			&decision.DecisionID,
			&decision.Title,
			&decision.Status,
			&decision.Scope,
			&decision.Owner,
			&decision.Date,
			&decision.Summary,
			&supersedesRaw,
			&supersededByRaw,
			&sourceRefsRaw,
			&updatedAt,
		); err != nil {
			return domain.DecisionLookupResult{}, domain.InternalError("scan decision record", err)
		}
		decision.Supersedes = splitCSVList(supersedesRaw)
		decision.SupersededBy = splitCSVList(supersededByRaw)
		decision.SourceRefs = splitPathList(sourceRefsRaw)
		decision.UpdatedAt = mustParseTime(updatedAt)
		decisions = append(decisions, decision)
	}
	if err := rows.Err(); err != nil {
		return domain.DecisionLookupResult{}, domain.InternalError("iterate decision records", err)
	}
	if err := rows.Close(); err != nil {
		return domain.DecisionLookupResult{}, domain.InternalError("close decision record rows", err)
	}
	for idx := range decisions {
		loaded, err := s.loadDecisionRecordDetails(ctx, decisions[idx])
		if err != nil {
			return domain.DecisionLookupResult{}, err
		}
		decisions[idx] = loaded
	}
	pageInfo := domain.PageInfo{}
	if len(decisions) > limit {
		pageInfo.HasMore = true
		pageInfo.NextCursor = encodeCursor(offset + limit)
		decisions = decisions[:limit]
	}
	return domain.DecisionLookupResult{Decisions: decisions, PageInfo: pageInfo}, nil
}

func (s *Store) GetDecisionRecord(ctx context.Context, decisionID string) (domain.DecisionRecord, error) {
	if !supportsDecisions(s.backend) {
		return domain.DecisionRecord{}, domain.UnsupportedError("decisions extension", s.backend)
	}
	var decision domain.DecisionRecord
	var supersedesRaw, supersededByRaw, sourceRefsRaw, updatedAt string
	err := s.db.QueryRowContext(ctx, `
SELECT decision_id, title, status, scope, owner, decision_date, summary, supersedes, superseded_by, source_refs, updated_at
FROM decision_records
WHERE decision_id = ?`, decisionID).Scan(
		&decision.DecisionID,
		&decision.Title,
		&decision.Status,
		&decision.Scope,
		&decision.Owner,
		&decision.Date,
		&decision.Summary,
		&supersedesRaw,
		&supersededByRaw,
		&sourceRefsRaw,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DecisionRecord{}, domain.NotFoundError("decision", decisionID)
	}
	if err != nil {
		return domain.DecisionRecord{}, domain.InternalError("query decision record", err)
	}
	decision.Supersedes = splitCSVList(supersedesRaw)
	decision.SupersededBy = splitCSVList(supersededByRaw)
	decision.SourceRefs = splitPathList(sourceRefsRaw)
	decision.UpdatedAt = mustParseTime(updatedAt)
	return s.loadDecisionRecordDetails(ctx, decision)
}

func (s *Store) loadDecisionRecordDetails(ctx context.Context, decision domain.DecisionRecord) (domain.DecisionRecord, error) {
	citationRows, err := s.db.QueryContext(ctx, `
SELECT source_doc_id, source_chunk_id, source_path, source_heading, source_line_start, source_line_end
FROM decision_citations
WHERE decision_id = ?
ORDER BY source_doc_id, source_chunk_id`, decision.DecisionID)
	if err != nil {
		return domain.DecisionRecord{}, domain.InternalError("query decision citations", err)
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
			return domain.DecisionRecord{}, domain.InternalError("scan decision citation", err)
		}
		citation.Heading = headingRaw.String
		decision.Citations = append(decision.Citations, citation)
	}
	if err := citationRows.Err(); err != nil {
		return domain.DecisionRecord{}, domain.InternalError("iterate decision citations", err)
	}
	return decision, nil
}

func (s *Store) rebuildDecisions(ctx context.Context) error {
	documents, err := s.loadAllDocuments(ctx)
	if err != nil {
		return err
	}
	chunksByDoc, err := s.loadChunksByDoc(ctx)
	if err != nil {
		return err
	}
	previousStates, err := s.loadProjectionStateSnapshots(ctx, "decisions")
	if err != nil {
		return err
	}

	type projectedDecision struct {
		doc       domain.Document
		decision  decisionProjection
		summary   string
		citation  domain.Citation
		freshness string
		details   map[string]string
	}
	projected := []projectedDecision{}
	decisionIDs := map[string]struct{}{}
	for _, doc := range documents {
		decision, ok := extractDecisionProjection(doc.Body)
		if !ok {
			continue
		}
		decisionIDs[decision.DecisionID] = struct{}{}
		projected = append(projected, projectedDecision{
			doc:      doc,
			decision: decision,
			summary:  firstSummaryParagraph(doc.Body),
			citation: documentCitation(doc, chunksByDoc[doc.DocID]),
		})
	}
	for idx := range projected {
		decision := projected[idx].decision
		freshness := "fresh"
		reason := "decision current"
		missingReplacements := []string{}
		if strings.EqualFold(decision.Status, "superseded") || len(decision.SupersededBy) > 0 {
			freshness = "stale"
			reason = "decision superseded"
			for _, replacementID := range decision.SupersededBy {
				if _, ok := decisionIDs[replacementID]; !ok {
					missingReplacements = appendUnique(missingReplacements, replacementID)
				}
			}
			if len(missingReplacements) > 0 {
				sort.Strings(missingReplacements)
				reason = "decision superseded with missing replacement"
			}
		}
		projected[idx].freshness = freshness
		projected[idx].details = map[string]string{
			"path":             projected[idx].doc.Path,
			"status":           decision.Status,
			"scope":            decision.Scope,
			"owner":            decision.Owner,
			"freshness_reason": reason,
		}
		if len(decision.Supersedes) > 0 {
			projected[idx].details["supersedes"] = strings.Join(decision.Supersedes, ", ")
		}
		if len(decision.SupersededBy) > 0 {
			projected[idx].details["superseded_by"] = strings.Join(decision.SupersededBy, ", ")
		}
		if len(decision.SourceRefs) > 0 {
			projected[idx].details["source_refs"] = strings.Join(decision.SourceRefs, ", ")
		}
		if len(missingReplacements) > 0 {
			projected[idx].details["missing_replacement_ids"] = strings.Join(missingReplacements, ", ")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin decisions rebuild", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, stmt := range []string{
		`DELETE FROM decision_citations;`,
		`DELETE FROM decision_records;`,
		`DELETE FROM projection_states WHERE projection_name = 'decisions';`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return domain.InternalError("reset decisions projection", err)
		}
	}

	now := s.now().UTC()
	for _, item := range projected {
		decision := item.decision
		versionInputs := []string{
			"decision:" + decision.DecisionID,
			"title:" + decision.Title,
			"status:" + decision.Status,
			"scope:" + decision.Scope,
			"owner:" + decision.Owner,
			"date:" + decision.Date,
			"supersedes:" + strings.Join(decision.Supersedes, ","),
			"superseded_by:" + strings.Join(decision.SupersededBy, ","),
			"source_refs:" + strings.Join(decision.SourceRefs, ","),
			"freshness:" + item.freshness,
			"updated:" + item.doc.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		for key, value := range item.details {
			versionInputs = append(versionInputs, "detail:"+key+"="+value)
		}
		sort.Strings(versionInputs)
		version := hashID("decisions", strings.Join(versionInputs, "|"))
		decisionUpdatedAt := now
		decisionChanged := true
		if previous, ok := previousStates[decision.DecisionID]; ok && previous.ProjectionVersion == version {
			decisionUpdatedAt = previous.UpdatedAt
			decisionChanged = false
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO decision_records (decision_id, title, status, scope, owner, decision_date, summary, supersedes, superseded_by, source_refs, source_doc_id, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			decision.DecisionID,
			decision.Title,
			decision.Status,
			decision.Scope,
			decision.Owner,
			decision.Date,
			item.summary,
			strings.Join(decision.Supersedes, ", "),
			strings.Join(decision.SupersededBy, ", "),
			strings.Join(decision.SourceRefs, ", "),
			item.doc.DocID,
			decisionUpdatedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return domain.InternalError("insert decision record", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO decision_citations (decision_id, source_doc_id, source_chunk_id, source_path, source_heading, source_line_start, source_line_end)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
			decision.DecisionID,
			item.citation.DocID,
			item.citation.ChunkID,
			item.citation.Path,
			nullIfEmpty(item.citation.Heading),
			item.citation.LineStart,
			item.citation.LineEnd,
		); err != nil {
			return domain.InternalError("insert decision citation", err)
		}
		if decisionChanged {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "decision", decision.DecisionID, now.Format(time.RFC3339Nano)),
				EventType:  "decision_extracted_from_doc",
				RefKind:    "decision",
				RefID:      decision.DecisionID,
				SourceRef:  "doc:" + item.doc.DocID,
				OccurredAt: now,
				Details: map[string]string{
					"decision_title": decision.Title,
					"path":           item.doc.Path,
				},
			}); err != nil {
				return domain.InternalError("record decisions provenance event", err)
			}
		}
		if err := upsertProjectionState(ctx, tx, domain.ProjectionState{
			Projection:        "decisions",
			RefKind:           "decision",
			RefID:             decision.DecisionID,
			SourceRef:         "doc:" + item.doc.DocID,
			Freshness:         item.freshness,
			ProjectionVersion: version,
			UpdatedAt:         decisionUpdatedAt,
			Details:           item.details,
		}); err != nil {
			return err
		}
		if decisionChanged {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "projection_refreshed", "decisions", decision.DecisionID, version, now.Format(time.RFC3339Nano)),
				EventType:  "projection_refreshed",
				RefKind:    "projection",
				RefID:      "decisions:" + decision.DecisionID,
				SourceRef:  "doc:" + item.doc.DocID,
				OccurredAt: now,
				Details: map[string]string{
					"projection":  "decisions",
					"decision_id": decision.DecisionID,
					"version":     version,
				},
			}); err != nil {
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit decisions rebuild", err)
	}
	return nil
}

func supportsDecisions(backend domain.BackendKind) bool {
	return backend == domain.BackendOpenClerk
}

func extractDecisionProjection(body string) (decisionProjection, bool) {
	lines := strings.Split(body, "\n")
	frontmatter, _ := parseFrontmatter(lines)
	projected := decisionProjection{
		DecisionID:   strings.TrimSpace(frontmatter["decision_id"]),
		Title:        strings.TrimSpace(frontmatter["decision_title"]),
		Status:       strings.TrimSpace(frontmatter["decision_status"]),
		Scope:        strings.TrimSpace(frontmatter["decision_scope"]),
		Owner:        strings.TrimSpace(frontmatter["decision_owner"]),
		Date:         strings.TrimSpace(frontmatter["decision_date"]),
		Supersedes:   splitCSVList(frontmatter["supersedes"]),
		SupersededBy: splitCSVList(frontmatter["superseded_by"]),
		SourceRefs:   splitPathList(frontmatter["source_refs"]),
	}
	if projected.DecisionID == "" || projected.Title == "" || projected.Status == "" {
		return decisionProjection{}, false
	}
	return projected, true
}
