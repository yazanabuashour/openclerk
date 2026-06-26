package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io/fs"
	"math"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
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
	limit, err := normalizePageLimit(query.Limit, 10)
	if err != nil {
		return domain.SearchResult{}, err
	}
	if (query.MetadataKey == "") != (query.MetadataValue == "") {
		return domain.SearchResult{}, domain.ValidationError("metadataKey and metadataValue must be provided together", nil)
	}
	offset := decodeCursor(query.Cursor)
	result, err := s.lexicalSearch(ctx, query, limit, offset)
	if err != nil {
		return domain.SearchResult{}, err
	}
	if len(result.Hits) > 0 || offset > 0 {
		return result, nil
	}
	return s.lexicalTokenFallbackSearch(ctx, query, limit, offset)
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
	limit, err := normalizePageLimit(query.Limit, 20)
	if err != nil {
		return domain.DocumentListResult{}, err
	}

	sqlQuery := `
SELECT d.doc_id, d.path, d.title, d.metadata_json, d.updated_at
FROM documents d`
	args := []any{}
	clauses := []string{}
	if prefix := strings.TrimSpace(query.PathPrefix); prefix != "" {
		clauses = append(clauses, "d.path LIKE ? ESCAPE '\\'")
		args = append(args, pathPrefixLikePattern(prefix))
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

	offset := decodeCursor(query.Cursor)
	documents, pageInfo := paginateSlice(documents, limit, offset)
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
	if _, err := s.vaultCreateAbsPath(relPath, "validate document path"); err != nil {
		return domain.Document{}, err
	}
	if _, err := s.lstatVaultPath(relPath, "stat document path"); err == nil {
		return domain.Document{}, domain.AlreadyExistsError("document path", relPath)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return domain.Document{}, domain.InternalError("stat document path", err)
	}
	if err := s.writeNewVaultFile(relPath, []byte(input.Body), "write document"); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return domain.Document{}, domain.AlreadyExistsError("document path", relPath)
		}
		return domain.Document{}, err
	}
	if err := s.syncDocumentFromDisk(ctx, relPath, input.Title); err != nil {
		return domain.Document{}, err
	}
	return s.getDocumentByPath(ctx, relPath)
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
	if _, err := s.vaultExistingAbsPath(doc.Path, "validate document path"); err != nil {
		return domain.Document{}, err
	}
	if err := s.writeExistingVaultFile(doc.Path, []byte(body), "append document content"); err != nil {
		return domain.Document{}, err
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
	if _, err := s.vaultExistingAbsPath(doc.Path, "validate document path"); err != nil {
		return domain.Document{}, err
	}
	if err := s.writeExistingVaultFile(doc.Path, []byte(body), "replace document section"); err != nil {
		return domain.Document{}, err
	}
	if err := s.syncDocumentFromDisk(ctx, doc.Path, ""); err != nil {
		return domain.Document{}, err
	}
	return s.GetDocument(ctx, docID)
}

func (s *Store) ReplaceDocument(ctx context.Context, docID string, input domain.ReplaceDocumentInput) (domain.Document, error) {
	if strings.TrimSpace(input.Body) == "" {
		return domain.Document{}, domain.ValidationError("body is required", nil)
	}
	doc, err := s.refreshDocumentFromDisk(ctx, docID)
	if err != nil {
		return domain.Document{}, err
	}
	if _, err := s.vaultExistingAbsPath(doc.Path, "validate document path"); err != nil {
		return domain.Document{}, err
	}
	if err := s.writeExistingVaultFile(doc.Path, []byte(strings.TrimRight(input.Body, "\n")+"\n"), "replace document"); err != nil {
		return domain.Document{}, err
	}
	if err := s.syncDocumentFromDisk(ctx, doc.Path, strings.TrimSpace(input.Title)); err != nil {
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

func (s *Store) lexicalTokenFallbackSearch(ctx context.Context, query domain.SearchQuery, limit int, offset int) (domain.SearchResult, error) {
	tokens := searchFallbackTokens(query.Text)
	if len(tokens) == 0 {
		return domain.SearchResult{}, nil
	}
	baseQuery := `
SELECT c.chunk_id, c.doc_id, d.title, c.path, c.heading, c.content, c.line_start, c.line_end
FROM chunks c
JOIN documents d ON d.doc_id = c.doc_id`
	whereClause, args := filteredDocumentClauses(query)
	if whereClause != "" {
		baseQuery += "\nWHERE " + whereClause
	}
	baseQuery += "\nORDER BY d.path, c.line_start, c.chunk_id\nLIMIT ?"
	args = append(args, maxLexicalFallbackCandidateRows)
	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return domain.SearchResult{}, domain.InternalError("run lexical fallback search", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	hits := []domain.SearchHit{}
	for rows.Next() {
		var (
			hit       domain.SearchHit
			pathValue string
			heading   string
			content   string
			lineStart int
			lineEnd   int
		)
		if err := rows.Scan(&hit.ChunkID, &hit.DocID, &hit.Title, &pathValue, &heading, &content, &lineStart, &lineEnd); err != nil {
			return domain.SearchResult{}, domain.InternalError("scan lexical fallback result", err)
		}
		score := searchFallbackScore(tokens, hit.Title, pathValue, heading, content)
		if score <= 0 {
			continue
		}
		hit.Score = score
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
		return domain.SearchResult{}, domain.InternalError("iterate lexical fallback results", err)
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			left := ""
			right := ""
			if len(hits[i].Citations) > 0 {
				left = hits[i].Citations[0].Path + hits[i].ChunkID
			}
			if len(hits[j].Citations) > 0 {
				right = hits[j].Citations[0].Path + hits[j].ChunkID
			}
			return left < right
		}
		return hits[i].Score > hits[j].Score
	})
	return paginateSearchResults(collapseSearchHitsByDocument(hits), limit, offset), nil
}

func collapseSearchHitsByDocument(hits []domain.SearchHit) []domain.SearchHit {
	seen := map[string]struct{}{}
	collapsed := make([]domain.SearchHit, 0, len(hits))
	for _, hit := range hits {
		pathValue := ""
		if len(hit.Citations) > 0 {
			pathValue = hit.Citations[0].Path
		}
		key := pathValue
		if key == "" {
			key = hit.DocID
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		collapsed = append(collapsed, hit)
	}
	return collapsed
}

func searchFallbackScore(tokens []string, title string, pathValue string, heading string, content string) float64 {
	titleTokens := stringSet(searchFallbackTokens(title))
	pathTokens := stringSet(searchFallbackTokens(pathValue))
	headingTokens := stringSet(searchFallbackTokens(heading))
	contentTokens := stringSet(searchFallbackTokens(content))
	score := 0.0
	for _, token := range tokens {
		if _, ok := titleTokens[token]; ok {
			score += 3
		}
		if _, ok := headingTokens[token]; ok {
			score += 2
		}
		if _, ok := pathTokens[token]; ok {
			score += 1.5
		}
		if _, ok := contentTokens[token]; ok {
			score += 1
		}
	}
	return score
}

func searchFallbackTokens(text string) []string {
	raw := wordPattern.FindAllString(strings.ToLower(text), -1)
	tokens := []string{}
	for _, token := range raw {
		if _, stop := searchFallbackStopwords()[token]; stop {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func searchFallbackStopwords() map[string]struct{} {
	return map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "before": {}, "between": {}, "by": {}, "for": {}, "from": {}, "in": {}, "into": {}, "is": {}, "it": {}, "no": {}, "not": {}, "of": {}, "on": {}, "or": {}, "should": {}, "that": {}, "the": {}, "then": {}, "through": {}, "to": {}, "use": {}, "when": {}, "where": {}, "with": {}, "without": {},
	}
}

func filteredDocumentClauses(query domain.SearchQuery) (string, []any) {
	clauses := []string{}
	args := []any{}
	if prefix := strings.TrimSpace(query.PathPrefix); prefix != "" {
		clauses = append(clauses, "d.path LIKE ? ESCAPE '\\'")
		args = append(args, pathPrefixLikePattern(prefix))
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

func pathPrefixLikePattern(prefix string) string {
	return escapeLikeLiteral(prefix) + "%"
}

func escapeLikeLiteral(value string) string {
	var builder strings.Builder
	for _, char := range value {
		switch char {
		case '\\', '%', '_':
			builder.WriteRune('\\')
		}
		builder.WriteRune(char)
	}
	return builder.String()
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
	hits, pageInfo := paginateSlice(hits, limit, offset)
	for idx := range hits {
		hits[idx].Rank = offset + idx + 1
	}
	return domain.SearchResult{Hits: hits, PageInfo: pageInfo}
}
