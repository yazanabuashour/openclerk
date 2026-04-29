package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"strconv"
	"strings"
	"time"
)

func (s *Store) ListProvenanceEvents(ctx context.Context, query domain.ProvenanceEventQuery) (domain.ProvenanceEventResult, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return domain.ProvenanceEventResult{}, domain.ValidationError("limit must be between 1 and 100", map[string]any{"limit": limit})
	}

	sqlQuery := `
SELECT event_id, event_type, ref_kind, ref_id, source_ref, occurred_at, details_json
FROM provenance_events`
	args := []any{}
	clauses := []string{}
	if refKind := strings.TrimSpace(query.RefKind); refKind != "" {
		clauses = append(clauses, "ref_kind = ?")
		args = append(args, refKind)
	}
	if refID := strings.TrimSpace(query.RefID); refID != "" {
		clauses = append(clauses, "ref_id = ?")
		args = append(args, refID)
	}
	if sourceRef := strings.TrimSpace(query.SourceRef); sourceRef != "" {
		clauses = append(clauses, "source_ref = ?")
		args = append(args, sourceRef)
	}
	if len(clauses) > 0 {
		sqlQuery += "\nWHERE " + strings.Join(clauses, " AND ")
	}
	offset := decodeCursor(query.Cursor)
	sqlQuery += `
ORDER BY occurred_at DESC, event_id DESC
LIMIT ? OFFSET ?`
	args = append(args, limit+1, offset)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return domain.ProvenanceEventResult{}, domain.InternalError("query provenance events", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	events := make([]domain.ProvenanceEvent, 0, limit+1)
	for rows.Next() {
		var (
			event       domain.ProvenanceEvent
			occurredAt  string
			detailsJSON string
		)
		if err := rows.Scan(&event.EventID, &event.EventType, &event.RefKind, &event.RefID, &event.SourceRef, &occurredAt, &detailsJSON); err != nil {
			return domain.ProvenanceEventResult{}, domain.InternalError("scan provenance event", err)
		}
		_ = json.Unmarshal([]byte(detailsJSON), &event.Details)
		event.OccurredAt = mustParseTime(occurredAt)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return domain.ProvenanceEventResult{}, domain.InternalError("iterate provenance events", err)
	}

	pageInfo := domain.PageInfo{}
	if len(events) > limit {
		pageInfo.HasMore = true
		pageInfo.NextCursor = encodeCursor(offset + limit)
		events = events[:limit]
	}
	return domain.ProvenanceEventResult{Events: events, PageInfo: pageInfo}, nil
}

func insertProvenanceEvent(ctx context.Context, tx *sql.Tx, event domain.ProvenanceEvent) error {
	detailsJSON, err := json.Marshal(event.Details)
	if err != nil {
		return domain.InternalError("encode provenance event details", err)
	}
	baseEventID := event.EventID
	for attempt := 0; attempt < 8; attempt++ {
		eventID := baseEventID
		if attempt > 0 {
			eventID = hashID(baseEventID, strconv.Itoa(attempt))
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO provenance_events (event_id, event_type, ref_kind, ref_id, source_ref, occurred_at, details_json)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
			eventID,
			event.EventType,
			event.RefKind,
			event.RefID,
			event.SourceRef,
			event.OccurredAt.UTC().Format(time.RFC3339Nano),
			string(detailsJSON),
		); err != nil {
			if isProvenanceEventIDConflict(err) {
				continue
			}
			return domain.InternalError("insert provenance event", err)
		}
		return nil
	}
	return domain.InternalError("insert provenance event", fmt.Errorf("provenance event id collision: %s", baseEventID))
}

func isProvenanceEventIDConflict(err error) bool {
	message := err.Error()
	return strings.Contains(message, "provenance_events.event_id") &&
		(strings.Contains(message, "constraint failed") || strings.Contains(message, "UNIQUE constraint failed"))
}
