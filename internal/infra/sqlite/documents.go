package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"io/fs"
	"math"
	"path/filepath"
	"strings"
)

func (s *Store) Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResult, error) {
	normalized, err := normalizeSearchQuery(query)
	if err != nil {
		return domain.SearchResult{}, err
	}
	query = normalized
	if strings.TrimSpace(query.Text) == "" {
		return domain.SearchResult{}, domain.ValidationError("search text is required", nil)
	}
	limit := query.Limit
	if limit == 0 {
		limit = 10
	}
	if limit < 1 || limit > 100 {
		return domain.SearchResult{}, domain.ValidationError("limit must be between 1 and 100", map[string]any{"limit": limit})
	}
	if (query.MetadataKey == "") != (query.MetadataValue == "") {
		return domain.SearchResult{}, domain.ValidationError("metadataKey and metadataValue must be provided together", nil)
	}
	return s.lexicalSearch(ctx, query, limit, decodeCursor(query.Cursor))
}

func (s *Store) ListDocuments(ctx context.Context, query domain.DocumentListQuery) (domain.DocumentListResult, error) {
	normalized, err := normalizeDocumentListQuery(query)
	if err != nil {
		return domain.DocumentListResult{}, err
	}
	query = normalized
	if (query.MetadataKey == "") != (query.MetadataValue == "") {
		return domain.DocumentListResult{}, domain.ValidationError("metadataKey and metadataValue must be provided together", nil)
	}
	limit := query.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return domain.DocumentListResult{}, domain.ValidationError("limit must be between 1 and 100", map[string]any{"limit": limit})
	}

	sqlQuery := `
SELECT d.doc_id, d.path, d.title, d.metadata_json, d.updated_at
FROM documents d`
	args := []any{}
	clauses := []string{}
	if prefix := strings.TrimSpace(query.PathPrefix); prefix != "" {
		clauses = append(clauses, "d.path LIKE ?")
		args = append(args, prefix+"%")
	}
	if query.MetadataKey != "" {
		sqlQuery += `
JOIN document_metadata dm ON dm.doc_id = d.doc_id`
		clauses = append(clauses, "dm.key_name = ? AND dm.value_text = ?")
		args = append(args, strings.ToLower(strings.TrimSpace(query.MetadataKey)), strings.TrimSpace(query.MetadataValue))
	}
	if len(clauses) > 0 {
		sqlQuery += "\nWHERE " + strings.Join(clauses, " AND ")
	}
	sqlQuery += `
ORDER BY d.path
LIMIT ? OFFSET ?`
	args = append(args, limit+1, decodeCursor(query.Cursor))

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return domain.DocumentListResult{}, domain.InternalError("query document registry", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	documents := make([]domain.DocumentSummary, 0, limit+1)
	for rows.Next() {
		var (
			document     domain.DocumentSummary
			metadataJSON string
			updatedAt    string
		)
		if err := rows.Scan(&document.DocID, &document.Path, &document.Title, &metadataJSON, &updatedAt); err != nil {
			return domain.DocumentListResult{}, domain.InternalError("scan document registry row", err)
		}
		_ = json.Unmarshal([]byte(metadataJSON), &document.Metadata)
		document.UpdatedAt = mustParseTime(updatedAt)
		documents = append(documents, document)
	}
	if err := rows.Err(); err != nil {
		return domain.DocumentListResult{}, domain.InternalError("iterate document registry rows", err)
	}

	pageInfo := domain.PageInfo{}
	offset := decodeCursor(query.Cursor)
	if len(documents) > limit {
		pageInfo.HasMore = true
		pageInfo.NextCursor = encodeCursor(offset + limit)
		documents = documents[:limit]
	}
	return domain.DocumentListResult{Documents: documents, PageInfo: pageInfo}, nil
}

func normalizeSearchQuery(query domain.SearchQuery) (domain.SearchQuery, error) {
	tag := strings.TrimSpace(query.Tag)
	if tag == "" {
		if query.Tag != "" {
			return domain.SearchQuery{}, domain.ValidationError("tag must be non-empty", nil)
		}
		return query, nil
	}
	if strings.TrimSpace(query.MetadataKey) != "" || strings.TrimSpace(query.MetadataValue) != "" {
		return domain.SearchQuery{}, domain.ValidationError("tag cannot be combined with metadataKey or metadataValue", nil)
	}
	query.Tag = tag
	query.MetadataKey = "tag"
	query.MetadataValue = tag
	return query, nil
}

func normalizeDocumentListQuery(query domain.DocumentListQuery) (domain.DocumentListQuery, error) {
	tag := strings.TrimSpace(query.Tag)
	if tag == "" {
		if query.Tag != "" {
			return domain.DocumentListQuery{}, domain.ValidationError("tag must be non-empty", nil)
		}
		return query, nil
	}
	if strings.TrimSpace(query.MetadataKey) != "" || strings.TrimSpace(query.MetadataValue) != "" {
		return domain.DocumentListQuery{}, domain.ValidationError("tag cannot be combined with metadataKey or metadataValue", nil)
	}
	query.Tag = tag
	query.MetadataKey = "tag"
	query.MetadataValue = tag
	return query, nil
}

func (s *Store) CreateDocument(ctx context.Context, input domain.CreateDocumentInput) (domain.Document, error) {
	relPath, err := normalizePath(input.Path)
	if err != nil {
		return domain.Document{}, err
	}
	if strings.TrimSpace(input.Title) == "" {
		return domain.Document{}, domain.ValidationError("title is required", nil)
	}
	if strings.TrimSpace(input.Body) == "" {
		return domain.Document{}, domain.ValidationError("body is required", nil)
	}
	absPath := filepath.Join(s.vaultRoot, filepath.FromSlash(relPath))
	if err := ensureDir(filepath.Dir(absPath)); err != nil {
		return domain.Document{}, domain.InternalError("create document directory", err)
	}
	if _, err := osStat(absPath); err == nil {
		return domain.Document{}, domain.AlreadyExistsError("document path", relPath)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return domain.Document{}, domain.InternalError("stat document path", err)
	}
	if err := osWriteFile(absPath, input.Body); err != nil {
		return domain.Document{}, domain.InternalError("write document", err)
	}
	if err := s.syncDocumentFromDisk(ctx, relPath, input.Title); err != nil {
		return domain.Document{}, err
	}
	return s.GetDocument(ctx, docIDForPath(relPath))
}

func (s *Store) GetDocument(ctx context.Context, docID string) (domain.Document, error) {
	const query = `
SELECT doc_id, path, title, body, headings_json, metadata_json, created_at, updated_at
FROM documents
WHERE doc_id = ?`
	var (
		document     domain.Document
		headingsJSON string
		metadataJSON string
		createdAt    string
		updatedAt    string
	)
	err := s.db.QueryRowContext(ctx, query, docID).Scan(
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
		return domain.Document{}, domain.NotFoundError("document", docID)
	}
	if err != nil {
		return domain.Document{}, domain.InternalError("query document", err)
	}
	_ = json.Unmarshal([]byte(headingsJSON), &document.Headings)
	_ = json.Unmarshal([]byte(metadataJSON), &document.Metadata)
	document.CreatedAt = mustParseTime(createdAt)
	document.UpdatedAt = mustParseTime(updatedAt)
	return document, nil
}

func (s *Store) AppendDocument(ctx context.Context, docID string, input domain.AppendDocumentInput) (domain.Document, error) {
	if strings.TrimSpace(input.Content) == "" {
		return domain.Document{}, domain.ValidationError("content is required", nil)
	}
	doc, err := s.refreshDocumentFromDisk(ctx, docID)
	if err != nil {
		return domain.Document{}, err
	}
	body := strings.TrimRight(doc.Body, "\n")
	body = body + "\n\n" + strings.TrimSpace(input.Content) + "\n"
	if err := osWriteFile(filepath.Join(s.vaultRoot, filepath.FromSlash(doc.Path)), body); err != nil {
		return domain.Document{}, domain.InternalError("append document content", err)
	}
	if err := s.syncDocumentFromDisk(ctx, doc.Path, doc.Title); err != nil {
		return domain.Document{}, err
	}
	return s.GetDocument(ctx, docID)
}

func (s *Store) ReplaceDocumentSection(ctx context.Context, docID string, input domain.ReplaceSectionInput) (domain.Document, error) {
	if strings.TrimSpace(input.Heading) == "" {
		return domain.Document{}, domain.ValidationError("heading is required", nil)
	}
	doc, err := s.refreshDocumentFromDisk(ctx, docID)
	if err != nil {
		return domain.Document{}, err
	}
	body, err := replaceSection(doc.Body, input.Heading, input.Content)
	if err != nil {
		return domain.Document{}, err
	}
	if err := osWriteFile(filepath.Join(s.vaultRoot, filepath.FromSlash(doc.Path)), body); err != nil {
		return domain.Document{}, domain.InternalError("replace document section", err)
	}
	if err := s.syncDocumentFromDisk(ctx, doc.Path, ""); err != nil {
		return domain.Document{}, err
	}
	return s.GetDocument(ctx, docID)
}

func (s *Store) refreshDocumentFromDisk(ctx context.Context, docID string) (domain.Document, error) {
	doc, err := s.GetDocument(ctx, docID)
	if err != nil {
		return domain.Document{}, err
	}
	if err := s.syncDocumentFromDisk(ctx, doc.Path, ""); err != nil {
		return domain.Document{}, err
	}
	return s.GetDocument(ctx, docID)
}

func (s *Store) GetChunk(ctx context.Context, chunkID string) (domain.Chunk, error) {
	const query = `
SELECT chunk_id, doc_id, path, heading, content, line_start, line_end
FROM chunks
WHERE chunk_id = ?`
	var chunk domain.Chunk
	err := s.db.QueryRowContext(ctx, query, chunkID).Scan(
		&chunk.ChunkID,
		&chunk.DocID,
		&chunk.Path,
		&chunk.Heading,
		&chunk.Content,
		&chunk.LineStart,
		&chunk.LineEnd,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Chunk{}, domain.NotFoundError("chunk", chunkID)
	}
	if err != nil {
		return domain.Chunk{}, domain.InternalError("query chunk", err)
	}
	return chunk, nil
}

func (s *Store) lexicalSearch(ctx context.Context, query domain.SearchQuery, limit int, offset int) (domain.SearchResult, error) {
	baseQuery := `
SELECT c.chunk_id, c.doc_id, d.title, c.path, c.heading, c.content, c.line_start, c.line_end, bm25(chunk_fts)
FROM chunk_fts
JOIN chunks c ON c.chunk_id = chunk_fts.chunk_id
JOIN documents d ON d.doc_id = c.doc_id`
	whereClause, args := filteredDocumentClauses(query)
	sqlQuery := baseQuery + "\nWHERE chunk_fts MATCH ?"
	args = append([]any{ftsExpression(query.Text)}, args...)
	if whereClause != "" {
		sqlQuery += " AND " + whereClause
	}
	sqlQuery += `
ORDER BY bm25(chunk_fts), c.chunk_id
LIMIT ? OFFSET ?`
	args = append(args, limit+1, offset)
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return domain.SearchResult{}, domain.InternalError("run lexical search", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	hits := make([]domain.SearchHit, 0, limit+1)
	for rows.Next() {
		var (
			hit       domain.SearchHit
			pathValue string
			heading   string
			content   string
			lineStart int
			lineEnd   int
			bm25Score float64
		)
		if err := rows.Scan(&hit.ChunkID, &hit.DocID, &hit.Title, &pathValue, &heading, &content, &lineStart, &lineEnd, &bm25Score); err != nil {
			return domain.SearchResult{}, domain.InternalError("scan lexical result", err)
		}
		hit.Score = 1 / (1 + math.Abs(bm25Score))
		hit.Snippet = snippetForSearch(content, query.Text)
		hit.Citations = []domain.Citation{{
			DocID:     hit.DocID,
			ChunkID:   hit.ChunkID,
			Path:      pathValue,
			Heading:   heading,
			LineStart: lineStart,
			LineEnd:   lineEnd,
		}}
		hits = append(hits, hit)
	}
	if err := rows.Err(); err != nil {
		return domain.SearchResult{}, domain.InternalError("iterate lexical results", err)
	}
	return paginateSearchResults(hits, limit, offset), nil
}

func filteredDocumentClauses(query domain.SearchQuery) (string, []any) {
	clauses := []string{}
	args := []any{}
	if prefix := strings.TrimSpace(query.PathPrefix); prefix != "" {
		clauses = append(clauses, "d.path LIKE ?")
		args = append(args, prefix+"%")
	}
	if query.MetadataKey != "" && query.MetadataValue != "" {
		clauses = append(clauses, `EXISTS (
SELECT 1
FROM document_metadata dm
WHERE dm.doc_id = d.doc_id AND dm.key_name = ? AND dm.value_text = ?
)`)
		args = append(args, strings.ToLower(strings.TrimSpace(query.MetadataKey)), strings.TrimSpace(query.MetadataValue))
	}
	return strings.Join(clauses, " AND "), args
}

func snippetForSearch(content string, query string) string {
	lower := strings.ToLower(content)
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return strings.TrimSpace(content)
	}
	index := strings.Index(lower, needle)
	if index == -1 {
		return firstNRunes(strings.TrimSpace(content), 180)
	}
	start := max(index-60, 0)
	end := min(index+len(needle)+80, len(content))
	return strings.TrimSpace(content[start:end])
}

func paginateSearchResults(hits []domain.SearchHit, limit int, offset int) domain.SearchResult {
	pageInfo := domain.PageInfo{}
	if len(hits) > limit {
		pageInfo.HasMore = true
		pageInfo.NextCursor = encodeCursor(offset + limit)
		hits = hits[:limit]
	}
	for idx := range hits {
		hits[idx].Rank = offset + idx + 1
	}
	return domain.SearchResult{Hits: hits, PageInfo: pageInfo}
}
