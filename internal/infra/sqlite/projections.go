package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

type storedProjectionState struct {
	ProjectionVersion string
	UpdatedAt         time.Time
}

type duplicateProjectionSource struct {
	DocID string
	Path  string
}

func (s *Store) ListProjectionStates(ctx context.Context, query domain.ProjectionStateQuery) (domain.ProjectionStateResult, error) {
	limit, err := normalizePageLimit(query.Limit, 20)
	if err != nil {
		return domain.ProjectionStateResult{}, err
	}

	sqlQuery := `
SELECT projection_name, ref_kind, ref_id, source_ref, freshness, projection_version, updated_at, details_json
FROM projection_states`
	args := []any{}
	clauses := []string{}
	if projection := strings.TrimSpace(query.Projection); projection != "" {
		clauses = append(clauses, "projection_name = ?")
		args = append(args, projection)
	}
	if refKind := strings.TrimSpace(query.RefKind); refKind != "" {
		clauses = append(clauses, "ref_kind = ?")
		args = append(args, refKind)
	}
	if refID := strings.TrimSpace(query.RefID); refID != "" {
		clauses = append(clauses, "ref_id = ?")
		args = append(args, refID)
	}
	if len(clauses) > 0 {
		sqlQuery += "\nWHERE " + strings.Join(clauses, " AND ")
	}
	offset := decodeCursor(query.Cursor)
	sqlQuery += `
ORDER BY projection_name, ref_kind, ref_id
LIMIT ? OFFSET ?`
	args = append(args, limit+1, offset)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return domain.ProjectionStateResult{}, domain.InternalError("query projection states", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	projections := make([]domain.ProjectionState, 0, limit+1)
	for rows.Next() {
		var (
			projection  domain.ProjectionState
			updatedAt   string
			detailsJSON string
		)
		if err := rows.Scan(&projection.Projection, &projection.RefKind, &projection.RefID, &projection.SourceRef, &projection.Freshness, &projection.ProjectionVersion, &updatedAt, &detailsJSON); err != nil {
			return domain.ProjectionStateResult{}, domain.InternalError("scan projection state", err)
		}
		_ = json.Unmarshal([]byte(detailsJSON), &projection.Details)
		projection.UpdatedAt = mustParseTime(updatedAt)
		projections = append(projections, projection)
	}
	if err := rows.Err(); err != nil {
		return domain.ProjectionStateResult{}, domain.InternalError("iterate projection states", err)
	}

	projections, pageInfo := paginateSlice(projections, limit, offset)
	return domain.ProjectionStateResult{Projections: projections, PageInfo: pageInfo}, nil
}

func (s *Store) loadProjectionStateSnapshots(ctx context.Context, projection string) (map[string]storedProjectionState, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT ref_id, projection_version, updated_at
FROM projection_states
WHERE projection_name = ?`, projection)
	if err != nil {
		return nil, domain.InternalError("query projection state snapshots", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	snapshots := map[string]storedProjectionState{}
	for rows.Next() {
		var (
			refID             string
			projectionVersion string
			updatedAt         string
		)
		if err := rows.Scan(&refID, &projectionVersion, &updatedAt); err != nil {
			return nil, domain.InternalError("scan projection state snapshot", err)
		}
		snapshots[refID] = storedProjectionState{
			ProjectionVersion: projectionVersion,
			UpdatedAt:         mustParseTime(updatedAt),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError("iterate projection state snapshots", err)
	}
	return snapshots, nil
}

func upsertProjectionState(ctx context.Context, tx *sql.Tx, projection domain.ProjectionState) error {
	detailsJSON, err := json.Marshal(projection.Details)
	if err != nil {
		return domain.InternalError("encode projection state details", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO projection_states (projection_name, ref_kind, ref_id, source_ref, freshness, projection_version, updated_at, details_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(projection_name, ref_kind, ref_id) DO UPDATE SET
	source_ref = excluded.source_ref,
	freshness = excluded.freshness,
	projection_version = excluded.projection_version,
	updated_at = excluded.updated_at,
	details_json = excluded.details_json`,
		projection.Projection,
		projection.RefKind,
		projection.RefID,
		projection.SourceRef,
		projection.Freshness,
		projection.ProjectionVersion,
		projection.UpdatedAt.UTC().Format(time.RFC3339Nano),
		string(detailsJSON),
	); err != nil {
		return domain.InternalError("upsert projection state", err)
	}
	return nil
}

func duplicateProjectionState(projection string, refKind string, refID string, sources []duplicateProjectionSource, previousStates map[string]storedProjectionState, now time.Time) domain.ProjectionState {
	details := duplicateProjectionDetails(refID, sources)
	versionInputs := []string{
		"projection:" + projection,
		"ref_kind:" + refKind,
		"ref_id:" + refID,
	}
	for _, source := range sortedDuplicateProjectionSources(sources) {
		versionInputs = append(versionInputs, "source:"+source.DocID+":"+source.Path)
	}
	sort.Strings(versionInputs)
	version := hashID(projection, "duplicate", strings.Join(versionInputs, "|"))
	updatedAt := now
	if previous, ok := previousStates[refID]; ok && previous.ProjectionVersion == version {
		updatedAt = previous.UpdatedAt
	}
	return domain.ProjectionState{
		Projection:        projection,
		RefKind:           refKind,
		RefID:             refID,
		SourceRef:         duplicateProjectionSourceRef(sources),
		Freshness:         "stale",
		ProjectionVersion: version,
		UpdatedAt:         updatedAt,
		Details:           details,
	}
}

func duplicateProjectionDetails(refID string, sources []duplicateProjectionSource) map[string]string {
	paths := make([]string, 0, len(sources))
	docIDs := make([]string, 0, len(sources))
	for _, source := range sortedDuplicateProjectionSources(sources) {
		paths = append(paths, source.Path)
		docIDs = append(docIDs, source.DocID)
	}
	return map[string]string{
		"duplicate_ref_id":  refID,
		"duplicate_paths":   strings.Join(paths, ", "),
		"duplicate_doc_ids": strings.Join(docIDs, ", "),
		"freshness_reason":  "duplicate projected id; no trusted record materialized",
	}
}

func duplicateProjectionSourceRef(sources []duplicateProjectionSource) string {
	refs := make([]string, 0, len(sources))
	for _, source := range sortedDuplicateProjectionSources(sources) {
		refs = append(refs, "doc:"+source.DocID)
	}
	return strings.Join(refs, ", ")
}

func insertDuplicateProjectionEvent(ctx context.Context, tx *sql.Tx, eventType string, refKind string, refID string, projection domain.ProjectionState, now time.Time) error {
	return insertProvenanceEventIfAbsent(ctx, tx, domain.ProvenanceEvent{
		EventID:    hashID("event", eventType, refKind, refID, projection.ProjectionVersion),
		EventType:  eventType,
		RefKind:    refKind,
		RefID:      refID,
		SourceRef:  projection.SourceRef,
		OccurredAt: now,
		Details:    projection.Details,
	})
}

func sortedDuplicateProjectionSources(sources []duplicateProjectionSource) []duplicateProjectionSource {
	sorted := append([]duplicateProjectionSource(nil), sources...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Path != sorted[j].Path {
			return sorted[i].Path < sorted[j].Path
		}
		return sorted[i].DocID < sorted[j].DocID
	})
	return sorted
}
