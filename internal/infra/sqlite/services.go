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

type serviceProjection struct {
	ServiceID string
	Name      string
	Status    string
	Owner     string
	Interface string
	Facts     []domain.ServiceFact
}

func (s *Store) ServicesLookup(ctx context.Context, input domain.ServiceLookupInput) (domain.ServiceLookupResult, error) {
	if !supportsServices(s.backend) {
		return domain.ServiceLookupResult{}, domain.UnsupportedError("services extension", s.backend)
	}
	limit := input.Limit
	if limit == 0 {
		limit = 10
	}
	if limit < 1 || limit > 100 {
		return domain.ServiceLookupResult{}, domain.ValidationError("limit must be between 1 and 100", map[string]any{"limit": limit})
	}
	offset := decodeCursor(input.Cursor)

	args := []any{}
	clauses := []string{}
	if text := strings.ToLower(strings.TrimSpace(input.Text)); text != "" {
		clauses = append(clauses, "(LOWER(service_id) LIKE ? OR LOWER(name) LIKE ? OR LOWER(summary) LIKE ?)")
		pattern := "%" + text + "%"
		args = append(args, pattern, pattern, pattern)
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		clauses = append(clauses, "LOWER(status) = ?")
		args = append(args, strings.ToLower(status))
	}
	if owner := strings.TrimSpace(input.Owner); owner != "" {
		clauses = append(clauses, "LOWER(owner) = ?")
		args = append(args, strings.ToLower(owner))
	}
	if serviceInterface := strings.TrimSpace(input.Interface); serviceInterface != "" {
		clauses = append(clauses, "LOWER(service_interface) = ?")
		args = append(args, strings.ToLower(serviceInterface))
	}

	query := `
SELECT service_id, name, status, owner, service_interface, summary, updated_at
FROM service_records`
	if len(clauses) > 0 {
		query += "\nWHERE " + strings.Join(clauses, " AND ")
	}
	query += `
ORDER BY name
LIMIT ? OFFSET ?`
	args = append(args, limit+1, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.ServiceLookupResult{}, domain.InternalError("query service records", err)
	}
	services := make([]domain.ServiceRecord, 0, limit+1)
	for rows.Next() {
		var service domain.ServiceRecord
		var updatedAt string
		if err := rows.Scan(&service.ServiceID, &service.Name, &service.Status, &service.Owner, &service.Interface, &service.Summary, &updatedAt); err != nil {
			return domain.ServiceLookupResult{}, domain.InternalError("scan service record", err)
		}
		service.UpdatedAt = mustParseTime(updatedAt)
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return domain.ServiceLookupResult{}, domain.InternalError("iterate service records", err)
	}
	if err := rows.Close(); err != nil {
		return domain.ServiceLookupResult{}, domain.InternalError("close service record rows", err)
	}
	for idx := range services {
		loaded, err := s.loadServiceRecordDetails(ctx, services[idx])
		if err != nil {
			return domain.ServiceLookupResult{}, err
		}
		services[idx] = loaded
	}
	pageInfo := domain.PageInfo{}
	if len(services) > limit {
		pageInfo.HasMore = true
		pageInfo.NextCursor = encodeCursor(offset + limit)
		services = services[:limit]
	}
	return domain.ServiceLookupResult{Services: services, PageInfo: pageInfo}, nil
}

func (s *Store) GetServiceRecord(ctx context.Context, serviceID string) (domain.ServiceRecord, error) {
	if !supportsServices(s.backend) {
		return domain.ServiceRecord{}, domain.UnsupportedError("services extension", s.backend)
	}
	var service domain.ServiceRecord
	var updatedAt string
	err := s.db.QueryRowContext(ctx, `
SELECT service_id, name, status, owner, service_interface, summary, updated_at
FROM service_records
WHERE service_id = ?`, serviceID).Scan(
		&service.ServiceID,
		&service.Name,
		&service.Status,
		&service.Owner,
		&service.Interface,
		&service.Summary,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ServiceRecord{}, domain.NotFoundError("service", serviceID)
	}
	if err != nil {
		return domain.ServiceRecord{}, domain.InternalError("query service record", err)
	}
	service.UpdatedAt = mustParseTime(updatedAt)
	return s.loadServiceRecordDetails(ctx, service)
}

func (s *Store) loadServiceRecordDetails(ctx context.Context, service domain.ServiceRecord) (domain.ServiceRecord, error) {
	factRows, err := s.db.QueryContext(ctx, `
SELECT key_name, value_text, observed_at
FROM service_facts
WHERE service_id = ?
ORDER BY key_name`, service.ServiceID)
	if err != nil {
		return domain.ServiceRecord{}, domain.InternalError("query service facts", err)
	}
	defer func() {
		_ = factRows.Close()
	}()
	for factRows.Next() {
		var (
			fact        domain.ServiceFact
			observedRaw sql.NullString
		)
		if err := factRows.Scan(&fact.Key, &fact.Value, &observedRaw); err != nil {
			return domain.ServiceRecord{}, domain.InternalError("scan service fact", err)
		}
		if observedRaw.Valid {
			observed := mustParseTime(observedRaw.String)
			fact.ObservedAt = &observed
		}
		service.Facts = append(service.Facts, fact)
	}
	if err := factRows.Err(); err != nil {
		return domain.ServiceRecord{}, domain.InternalError("iterate service facts", err)
	}

	citationRows, err := s.db.QueryContext(ctx, `
SELECT source_doc_id, source_chunk_id, source_path, source_heading, source_line_start, source_line_end
FROM service_citations
WHERE service_id = ?
ORDER BY source_doc_id, source_chunk_id`, service.ServiceID)
	if err != nil {
		return domain.ServiceRecord{}, domain.InternalError("query service citations", err)
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
			return domain.ServiceRecord{}, domain.InternalError("scan service citation", err)
		}
		citation.Heading = headingRaw.String
		service.Citations = append(service.Citations, citation)
	}
	if err := citationRows.Err(); err != nil {
		return domain.ServiceRecord{}, domain.InternalError("iterate service citations", err)
	}
	return service, nil
}

func (s *Store) rebuildServices(ctx context.Context) error {
	documents, err := s.loadAllDocuments(ctx)
	if err != nil {
		return err
	}
	chunksByDoc, err := s.loadChunksByDoc(ctx)
	if err != nil {
		return err
	}
	previousStates, err := s.loadProjectionStateSnapshots(ctx, "services")
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin services rebuild", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, stmt := range []string{
		`DELETE FROM service_citations;`,
		`DELETE FROM service_facts;`,
		`DELETE FROM service_records;`,
		`DELETE FROM projection_states WHERE projection_name = 'services';`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return domain.InternalError("reset services projection", err)
		}
	}

	now := s.now().UTC()
	for _, doc := range documents {
		projected, ok := extractServiceProjection(doc.Body)
		if !ok {
			continue
		}
		summary := firstSummaryParagraph(doc.Body)
		versionInputs := []string{
			"service:" + projected.ServiceID,
			"name:" + projected.Name,
			"status:" + projected.Status,
			"owner:" + projected.Owner,
			"interface:" + projected.Interface,
			"updated:" + doc.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		for _, fact := range projected.Facts {
			versionInputs = append(versionInputs, "fact:"+fact.Key+"="+fact.Value)
		}
		sort.Strings(versionInputs)
		version := hashID("services", strings.Join(versionInputs, "|"))
		serviceUpdatedAt := now
		serviceChanged := true
		if previous, ok := previousStates[projected.ServiceID]; ok && previous.ProjectionVersion == version {
			serviceUpdatedAt = previous.UpdatedAt
			serviceChanged = false
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO service_records (service_id, name, status, owner, service_interface, summary, source_doc_id, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			projected.ServiceID,
			projected.Name,
			projected.Status,
			projected.Owner,
			projected.Interface,
			summary,
			doc.DocID,
			serviceUpdatedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return domain.InternalError("insert service record", err)
		}
		for _, fact := range projected.Facts {
			var observedAt *string
			if fact.ObservedAt != nil {
				value := fact.ObservedAt.UTC().Format(time.RFC3339Nano)
				observedAt = &value
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO service_facts (service_id, key_name, value_text, observed_at)
VALUES (?, ?, ?, ?)`,
				projected.ServiceID,
				fact.Key,
				fact.Value,
				observedAt,
			); err != nil {
				return domain.InternalError("insert service fact", err)
			}
		}
		citation := documentCitation(doc, chunksByDoc[doc.DocID])
		if _, err := tx.ExecContext(ctx, `
INSERT INTO service_citations (service_id, source_doc_id, source_chunk_id, source_path, source_heading, source_line_start, source_line_end)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
			projected.ServiceID,
			citation.DocID,
			citation.ChunkID,
			citation.Path,
			nullIfEmpty(citation.Heading),
			citation.LineStart,
			citation.LineEnd,
		); err != nil {
			return domain.InternalError("insert service citation", err)
		}
		if serviceChanged {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "service", projected.ServiceID, now.Format(time.RFC3339Nano)),
				EventType:  "service_extracted_from_doc",
				RefKind:    "service",
				RefID:      projected.ServiceID,
				SourceRef:  "doc:" + doc.DocID,
				OccurredAt: now,
				Details: map[string]string{
					"service_name": projected.Name,
					"path":         doc.Path,
				},
			}); err != nil {
				return domain.InternalError("record services provenance event", err)
			}
		}
		if err := upsertProjectionState(ctx, tx, domain.ProjectionState{
			Projection:        "services",
			RefKind:           "service",
			RefID:             projected.ServiceID,
			SourceRef:         "doc:" + doc.DocID,
			Freshness:         "fresh",
			ProjectionVersion: version,
			UpdatedAt:         serviceUpdatedAt,
			Details: map[string]string{
				"path":      doc.Path,
				"status":    projected.Status,
				"owner":     projected.Owner,
				"interface": projected.Interface,
			},
		}); err != nil {
			return err
		}
		if serviceChanged {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "projection_refreshed", "services", projected.ServiceID, version, now.Format(time.RFC3339Nano)),
				EventType:  "projection_refreshed",
				RefKind:    "projection",
				RefID:      "services:" + projected.ServiceID,
				SourceRef:  "doc:" + doc.DocID,
				OccurredAt: now,
				Details: map[string]string{
					"projection": "services",
					"service_id": projected.ServiceID,
					"version":    version,
				},
			}); err != nil {
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit services rebuild", err)
	}
	return nil
}

func supportsServices(backend domain.BackendKind) bool {
	return backend == domain.BackendOpenClerk
}

func extractServiceProjection(body string) (serviceProjection, bool) {
	lines := strings.Split(body, "\n")
	frontmatter, contentStart := parseFrontmatter(lines)
	recordFacts := extractRecordFacts(lines, contentStart)
	facts := make([]domain.ServiceFact, 0, len(recordFacts))
	for _, fact := range recordFacts {
		facts = append(facts, domain.ServiceFact(fact))
	}

	projected := serviceProjection{
		ServiceID: strings.TrimSpace(frontmatter["service_id"]),
		Name:      strings.TrimSpace(frontmatter["service_name"]),
		Status:    strings.TrimSpace(frontmatter["service_status"]),
		Owner:     strings.TrimSpace(frontmatter["service_owner"]),
		Interface: strings.TrimSpace(frontmatter["service_interface"]),
		Facts:     facts,
	}
	if projected.ServiceID == "" && strings.EqualFold(strings.TrimSpace(frontmatter["entity_type"]), "service") {
		projected.ServiceID = strings.TrimSpace(frontmatter["entity_id"])
		projected.Name = strings.TrimSpace(frontmatter["entity_name"])
	}
	if projected.ServiceID == "" || projected.Name == "" {
		return serviceProjection{}, false
	}
	if projected.Status == "" {
		projected.Status = serviceFactValue(facts, "status")
	}
	if projected.Owner == "" {
		projected.Owner = serviceFactValue(facts, "owner")
	}
	if projected.Interface == "" {
		projected.Interface = serviceFactValue(facts, "interface")
	}
	return projected, true
}

func serviceFactValue(facts []domain.ServiceFact, key string) string {
	for _, fact := range facts {
		if strings.EqualFold(strings.TrimSpace(fact.Key), key) {
			return strings.TrimSpace(fact.Value)
		}
	}
	return ""
}

func firstSummaryParagraph(body string) string {
	lines := strings.Split(body, "\n")
	_, contentStart := parseFrontmatter(lines)
	summaryLines := []string{}
	inSummary := false
	summaryLevel := 0
	for idx := contentStart; idx < len(lines); idx++ {
		line := strings.TrimSpace(lines[idx])
		if matches := headingPattern.FindStringSubmatch(line); len(matches) > 0 {
			level := len(matches[1])
			if inSummary && level <= summaryLevel {
				break
			}
			inSummary = strings.EqualFold(strings.TrimSpace(matches[2]), "Summary")
			if inSummary {
				summaryLevel = level
				summaryLines = summaryLines[:0]
			}
			continue
		}
		if !inSummary {
			continue
		}
		if line == "" {
			if len(summaryLines) > 0 {
				break
			}
			continue
		}
		summaryLines = append(summaryLines, line)
	}
	if len(summaryLines) > 0 {
		return firstNRunes(strings.Join(summaryLines, "\n"), 240)
	}
	for _, block := range strings.Split(body, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" || strings.HasPrefix(block, "---") || strings.HasPrefix(block, "#") {
			continue
		}
		return firstNRunes(block, 240)
	}
	return ""
}
