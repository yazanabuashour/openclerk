package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

const embeddingDimensions = 16

var (
	headingPattern = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)
	linkPattern    = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	wordPattern    = regexp.MustCompile(`[A-Za-z0-9_]+`)
)

type Config struct {
	Backend           domain.BackendKind
	DatabasePath      string
	VaultRoot         string
	EmbeddingProvider string
}

type Store struct {
	db                *sql.DB
	backend           domain.BackendKind
	vaultRoot         string
	embeddingProvider string
	now               func() time.Time
}

type section struct {
	Heading   string
	Level     int
	Content   string
	LineStart int
	LineEnd   int
}

func New(ctx context.Context, cfg Config) (*Store, error) {
	if cfg.DatabasePath == "" {
		return nil, domain.ValidationError("database path is required", nil)
	}
	if cfg.VaultRoot == "" {
		return nil, domain.ValidationError("vault root is required", nil)
	}
	if err := ensureDir(cfg.VaultRoot); err != nil {
		return nil, domain.InternalError("create vault root", err)
	}
	if err := ensureDir(filepath.Dir(cfg.DatabasePath)); err != nil {
		return nil, domain.InternalError("create database directory", err)
	}

	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, domain.InternalError("open sqlite database", err)
	}
	db.SetMaxOpenConns(1)

	store := &Store{
		db:                db,
		backend:           cfg.Backend,
		vaultRoot:         cfg.VaultRoot,
		embeddingProvider: strings.TrimSpace(cfg.EmbeddingProvider),
		now:               time.Now,
	}
	if err := store.initSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.syncVault(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Capabilities(_ context.Context) (domain.Capabilities, error) {
	capabilities := domain.Capabilities{
		Backend:     s.backend,
		AuthMode:    "none",
		SearchModes: []string{"lexical"},
	}
	if s.backend == domain.BackendHybrid && s.embeddingProvider != "" {
		capabilities.SearchModes = []string{"lexical", "vector", "hybrid"}
	}
	switch s.backend {
	case domain.BackendGraph:
		capabilities.Extensions = []string{"graph"}
	case domain.BackendRecords:
		capabilities.Extensions = []string{"records"}
	default:
		capabilities.Extensions = []string{}
	}
	return capabilities, nil
}

func (s *Store) Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResult, error) {
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
	if s.backend == domain.BackendHybrid && s.embeddingProvider != "" {
		return s.hybridSearch(ctx, query.Text, limit, decodeCursor(query.Cursor))
	}
	return s.lexicalSearch(ctx, query.Text, limit, decodeCursor(query.Cursor))
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
SELECT doc_id, path, title, body, headings_json, created_at, updated_at
FROM documents
WHERE doc_id = ?`
	var (
		document     domain.Document
		headingsJSON string
		createdAt    string
		updatedAt    string
	)
	err := s.db.QueryRowContext(ctx, query, docID).Scan(
		&document.DocID,
		&document.Path,
		&document.Title,
		&document.Body,
		&headingsJSON,
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
	document.CreatedAt = mustParseTime(createdAt)
	document.UpdatedAt = mustParseTime(updatedAt)
	return document, nil
}

func (s *Store) AppendDocument(ctx context.Context, docID string, input domain.AppendDocumentInput) (domain.Document, error) {
	if strings.TrimSpace(input.Content) == "" {
		return domain.Document{}, domain.ValidationError("content is required", nil)
	}
	doc, err := s.GetDocument(ctx, docID)
	if err != nil {
		return domain.Document{}, err
	}
	body := strings.TrimRight(doc.Body, "\n")
	body = body + "\n\n" + strings.TrimSpace(input.Content) + "\n"
	if err := osWriteFile(filepath.Join(s.vaultRoot, filepath.FromSlash(doc.Path)), body); err != nil {
		return domain.Document{}, domain.InternalError("append document content", err)
	}
	if err := s.syncDocumentFromDisk(ctx, doc.Path, ""); err != nil {
		return domain.Document{}, err
	}
	return s.GetDocument(ctx, docID)
}

func (s *Store) ReplaceDocumentSection(ctx context.Context, docID string, input domain.ReplaceSectionInput) (domain.Document, error) {
	if strings.TrimSpace(input.Heading) == "" {
		return domain.Document{}, domain.ValidationError("heading is required", nil)
	}
	doc, err := s.GetDocument(ctx, docID)
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

func (s *Store) GraphNeighborhood(ctx context.Context, input domain.GraphNeighborhoodInput) (domain.GraphNeighborhood, error) {
	if s.backend != domain.BackendGraph {
		return domain.GraphNeighborhood{}, domain.UnsupportedError("graph extension", s.backend)
	}
	nodeID := strings.TrimSpace(input.NodeID)
	if nodeID == "" {
		switch {
		case input.DocID != "":
			nodeID = "doc:" + input.DocID
		case input.ChunkID != "":
			nodeID = "chunk:" + input.ChunkID
		default:
			return domain.GraphNeighborhood{}, domain.ValidationError("docId, chunkId, or nodeId is required", nil)
		}
	}
	limit := input.Limit
	if limit == 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT edge_id, from_node_id, to_node_id, kind, evidence_doc_id, evidence_chunk_id, evidence_path, evidence_heading, evidence_line_start, evidence_line_end
FROM graph_edges
WHERE from_node_id = ? OR to_node_id = ?
ORDER BY edge_id
LIMIT ?`, nodeID, nodeID, limit)
	if err != nil {
		return domain.GraphNeighborhood{}, domain.InternalError("query graph edges", err)
	}
	defer rows.Close()

	edges := make([]domain.GraphEdge, 0, limit)
	nodeSet := map[string]struct{}{nodeID: {}}
	for rows.Next() {
		var (
			edge       domain.GraphEdge
			citation   domain.Citation
			headingRaw sql.NullString
		)
		if err := rows.Scan(
			&edge.EdgeID,
			&edge.FromNodeID,
			&edge.ToNodeID,
			&edge.Kind,
			&citation.DocID,
			&citation.ChunkID,
			&citation.Path,
			&headingRaw,
			&citation.LineStart,
			&citation.LineEnd,
		); err != nil {
			return domain.GraphNeighborhood{}, domain.InternalError("scan graph edge", err)
		}
		citation.Heading = headingRaw.String
		edge.Citations = []domain.Citation{citation}
		edges = append(edges, edge)
		nodeSet[edge.FromNodeID] = struct{}{}
		nodeSet[edge.ToNodeID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return domain.GraphNeighborhood{}, domain.InternalError("iterate graph edges", err)
	}

	nodeIDs := make([]string, 0, len(nodeSet))
	for id := range nodeSet {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)
	nodes := make([]domain.GraphNode, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		var (
			node       domain.GraphNode
			citation   domain.Citation
			headingRaw sql.NullString
		)
		err := s.db.QueryRowContext(ctx, `
SELECT node_id, type, label, evidence_doc_id, evidence_chunk_id, evidence_path, evidence_heading, evidence_line_start, evidence_line_end
FROM graph_nodes
WHERE node_id = ?`, id).Scan(
			&node.NodeID,
			&node.Type,
			&node.Label,
			&citation.DocID,
			&citation.ChunkID,
			&citation.Path,
			&headingRaw,
			&citation.LineStart,
			&citation.LineEnd,
		)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return domain.GraphNeighborhood{}, domain.InternalError("query graph node", err)
		}
		citation.Heading = headingRaw.String
		node.Citations = []domain.Citation{citation}
		nodes = append(nodes, node)
	}

	return domain.GraphNeighborhood{Nodes: nodes, Edges: edges}, nil
}

func (s *Store) RecordsLookup(ctx context.Context, input domain.RecordLookupInput) (domain.RecordLookupResult, error) {
	if s.backend != domain.BackendRecords {
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
	if s.backend != domain.BackendRecords {
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
	defer factRows.Close()
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
	defer citationRows.Close()
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

func (s *Store) initSchema(ctx context.Context) error {
	statements := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS documents (
			doc_id TEXT PRIMARY KEY,
			path TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			body TEXT NOT NULL,
			headings_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS chunks (
			chunk_id TEXT PRIMARY KEY,
			doc_id TEXT NOT NULL,
			path TEXT NOT NULL,
			heading TEXT NOT NULL,
			content TEXT NOT NULL,
			line_start INTEGER NOT NULL,
			line_end INTEGER NOT NULL,
			FOREIGN KEY (doc_id) REFERENCES documents(doc_id) ON DELETE CASCADE
		);`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS chunk_fts USING fts5(
			chunk_id UNINDEXED,
			doc_id UNINDEXED,
			path UNINDEXED,
			heading,
			content,
			tokenize = 'unicode61'
		);`,
		`CREATE TABLE IF NOT EXISTS embeddings (
			chunk_id TEXT PRIMARY KEY,
			vector_json TEXT NOT NULL,
			FOREIGN KEY (chunk_id) REFERENCES chunks(chunk_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS graph_nodes (
			node_id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			label TEXT NOT NULL,
			evidence_doc_id TEXT NOT NULL,
			evidence_chunk_id TEXT NOT NULL,
			evidence_path TEXT NOT NULL,
			evidence_heading TEXT,
			evidence_line_start INTEGER NOT NULL,
			evidence_line_end INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS graph_edges (
			edge_id TEXT PRIMARY KEY,
			from_node_id TEXT NOT NULL,
			to_node_id TEXT NOT NULL,
			kind TEXT NOT NULL,
			evidence_doc_id TEXT NOT NULL,
			evidence_chunk_id TEXT NOT NULL,
			evidence_path TEXT NOT NULL,
			evidence_heading TEXT,
			evidence_line_start INTEGER NOT NULL,
			evidence_line_end INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS record_entities (
			entity_id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			name TEXT NOT NULL,
			summary TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS record_facts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id TEXT NOT NULL,
			key_name TEXT NOT NULL,
			value_text TEXT NOT NULL,
			observed_at TEXT,
			FOREIGN KEY (entity_id) REFERENCES record_entities(entity_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS record_citations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			source_chunk_id TEXT NOT NULL,
			source_path TEXT NOT NULL,
			source_heading TEXT,
			source_line_start INTEGER NOT NULL,
			source_line_end INTEGER NOT NULL,
			FOREIGN KEY (entity_id) REFERENCES record_entities(entity_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS provenance_events (
			event_id TEXT PRIMARY KEY,
			event_type TEXT NOT NULL,
			ref_kind TEXT NOT NULL,
			ref_id TEXT NOT NULL,
			source_ref TEXT NOT NULL,
			occurred_at TEXT NOT NULL,
			details_json TEXT NOT NULL
		);`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return domain.InternalError("initialize sqlite schema", err)
		}
	}
	return nil
}

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
	switch s.backend {
	case domain.BackendGraph:
		return s.rebuildGraph(ctx)
	case domain.BackendRecords:
		return s.rebuildRecords(ctx)
	default:
		return nil
	}
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
	defer rows.Close()

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
	defer tx.Rollback()

	for _, docID := range staleDocIDs {
		if _, err := tx.ExecContext(ctx, `DELETE FROM chunk_fts WHERE doc_id = ?`, docID); err != nil {
			return domain.InternalError("delete missing document chunks from index", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM documents WHERE doc_id = ?`, docID); err != nil {
			return domain.InternalError("delete missing document", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM embeddings WHERE chunk_id NOT IN (SELECT chunk_id FROM chunks)`); err != nil {
		return domain.InternalError("prune embeddings after document deletion", err)
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin transaction", err)
	}
	defer tx.Rollback()

	createdAt := now.Format(time.RFC3339Nano)
	var existingTitle string
	var createdAtExisting string
	err = tx.QueryRowContext(ctx, `SELECT created_at, title FROM documents WHERE doc_id = ?`, docID).Scan(&createdAtExisting, &existingTitle)
	if err == nil {
		createdAt = createdAtExisting
	} else if !errors.Is(err, sql.ErrNoRows) {
		return domain.InternalError("query existing document timestamp", err)
	}

	title := resolvedDocumentTitle(relPath, body, headings, frontmatter, preferredTitle, existingTitle)
	if _, err := tx.ExecContext(ctx, `
INSERT INTO documents (doc_id, path, title, body, headings_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(doc_id) DO UPDATE SET
	path = excluded.path,
	title = excluded.title,
	body = excluded.body,
	headings_json = excluded.headings_json,
	updated_at = excluded.updated_at`,
		docID,
		relPath,
		title,
		body,
		string(headingsJSON),
		createdAt,
		now.Format(time.RFC3339Nano),
	); err != nil {
		return domain.InternalError("upsert document", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM chunk_fts WHERE doc_id = ?`, docID); err != nil {
		return domain.InternalError("delete indexed chunks", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM chunks WHERE doc_id = ?`, docID); err != nil {
		return domain.InternalError("delete chunks", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM embeddings WHERE chunk_id NOT IN (SELECT chunk_id FROM chunks)`); err != nil {
		return domain.InternalError("prune embeddings", err)
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
		if s.embeddingProvider != "" {
			vectorJSON, err := json.Marshal(embedText(sec.Content))
			if err != nil {
				return domain.InternalError("encode embedding", err)
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO embeddings (chunk_id, vector_json)
VALUES (?, ?)`, chunkID, string(vectorJSON)); err != nil {
				return domain.InternalError("insert embedding", err)
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO provenance_events (event_id, event_type, ref_kind, ref_id, source_ref, occurred_at, details_json)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		hashID("event", relPath, now.Format(time.RFC3339Nano)),
		"document_synced",
		"document",
		docID,
		"doc:"+docID,
		now.Format(time.RFC3339Nano),
		fmt.Sprintf(`{"path":%q}`, relPath),
	); err != nil {
		return domain.InternalError("record provenance event", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit document sync", err)
	}
	switch s.backend {
	case domain.BackendGraph:
		return s.rebuildGraph(ctx)
	case domain.BackendRecords:
		return s.rebuildRecords(ctx)
	default:
		return nil
	}
}

func (s *Store) lexicalSearch(ctx context.Context, text string, limit int, offset int) (domain.SearchResult, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT c.chunk_id, c.doc_id, d.title, c.path, c.heading, c.content, c.line_start, c.line_end, bm25(chunk_fts)
FROM chunk_fts
JOIN chunks c ON c.chunk_id = chunk_fts.chunk_id
JOIN documents d ON d.doc_id = c.doc_id
WHERE chunk_fts MATCH ?
ORDER BY bm25(chunk_fts), c.chunk_id
LIMIT ? OFFSET ?`, ftsExpression(text), limit+1, offset)
	if err != nil {
		return domain.SearchResult{}, domain.InternalError("run lexical search", err)
	}
	defer rows.Close()
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
		hit.Snippet = snippetForSearch(content, text)
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

func (s *Store) hybridSearch(ctx context.Context, text string, limit int, offset int) (domain.SearchResult, error) {
	lexical, err := s.lexicalSearch(ctx, text, max(limit*2, 20), 0)
	if err != nil {
		return domain.SearchResult{}, err
	}
	queryVector := embedText(text)
	rows, err := s.db.QueryContext(ctx, `
SELECT c.chunk_id, c.doc_id, d.title, c.path, c.heading, c.content, c.line_start, c.line_end, e.vector_json
FROM chunks c
JOIN documents d ON d.doc_id = c.doc_id
JOIN embeddings e ON e.chunk_id = c.chunk_id`)
	if err != nil {
		return domain.SearchResult{}, domain.InternalError("query embeddings", err)
	}
	defer rows.Close()

	type candidate struct {
		hit domain.SearchHit
	}
	vectorCandidates := make([]candidate, 0, 32)
	for rows.Next() {
		var (
			hit        domain.SearchHit
			pathValue  string
			heading    string
			content    string
			lineStart  int
			lineEnd    int
			vectorJSON string
		)
		if err := rows.Scan(&hit.ChunkID, &hit.DocID, &hit.Title, &pathValue, &heading, &content, &lineStart, &lineEnd, &vectorJSON); err != nil {
			return domain.SearchResult{}, domain.InternalError("scan embedding row", err)
		}
		var vector []float64
		if err := json.Unmarshal([]byte(vectorJSON), &vector); err != nil {
			return domain.SearchResult{}, domain.InternalError("decode embedding", err)
		}
		hit.Score = cosineSimilarity(queryVector, vector)
		hit.Snippet = snippetForSearch(content, text)
		hit.Citations = []domain.Citation{{
			DocID:     hit.DocID,
			ChunkID:   hit.ChunkID,
			Path:      pathValue,
			Heading:   heading,
			LineStart: lineStart,
			LineEnd:   lineEnd,
		}}
		vectorCandidates = append(vectorCandidates, candidate{hit: hit})
	}
	if err := rows.Err(); err != nil {
		return domain.SearchResult{}, domain.InternalError("iterate embeddings", err)
	}
	sort.Slice(vectorCandidates, func(i, j int) bool {
		if vectorCandidates[i].hit.Score == vectorCandidates[j].hit.Score {
			return vectorCandidates[i].hit.ChunkID < vectorCandidates[j].hit.ChunkID
		}
		return vectorCandidates[i].hit.Score > vectorCandidates[j].hit.Score
	})

	fused := map[string]*domain.SearchHit{}
	for rank, hit := range lexical.Hits {
		copyHit := hit
		copyHit.Score = 1.0 / float64(60+rank+1)
		fused[hit.ChunkID] = &copyHit
	}
	for rank, candidate := range vectorCandidates {
		if rank >= max(limit*4, 50) {
			break
		}
		score := 1.0 / float64(60+rank+1)
		if existing, ok := fused[candidate.hit.ChunkID]; ok {
			existing.Score += score
			continue
		}
		copyHit := candidate.hit
		copyHit.Score = score
		fused[candidate.hit.ChunkID] = &copyHit
	}

	hits := make([]domain.SearchHit, 0, len(fused))
	for _, hit := range fused {
		hits = append(hits, *hit)
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].ChunkID < hits[j].ChunkID
		}
		return hits[i].Score > hits[j].Score
	})
	if offset > len(hits) {
		offset = len(hits)
	}
	return paginateSearchResults(hits[offset:], limit, offset), nil
}

func (s *Store) rebuildGraph(ctx context.Context) error {
	documents, err := s.loadAllDocuments(ctx)
	if err != nil {
		return err
	}
	chunksByDoc, err := s.loadChunksByDoc(ctx)
	if err != nil {
		return err
	}
	documentIndex := documentsByPath(documents)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin graph rebuild", err)
	}
	defer tx.Rollback()
	for _, stmt := range []string{
		`DELETE FROM graph_edges;`,
		`DELETE FROM graph_nodes;`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return domain.InternalError("reset graph projection", err)
		}
	}

	for _, doc := range documents {
		nodeID := "doc:" + doc.DocID
		citation := documentCitation(doc, chunksByDoc[doc.DocID])
		if err := insertGraphNode(ctx, tx, nodeID, "document", doc.Title, citation); err != nil {
			return err
		}
		for _, chunk := range chunksByDoc[doc.DocID] {
			chunkNodeID := "chunk:" + chunk.ChunkID
			if err := insertGraphNode(ctx, tx, chunkNodeID, "chunk", chunk.Heading, domain.Citation{
				DocID:     chunk.DocID,
				ChunkID:   chunk.ChunkID,
				Path:      chunk.Path,
				Heading:   chunk.Heading,
				LineStart: chunk.LineStart,
				LineEnd:   chunk.LineEnd,
			}); err != nil {
				return err
			}
			if err := insertGraphEdge(ctx, tx, hashID("edge", nodeID, chunkNodeID), nodeID, chunkNodeID, "mentions", domain.Citation{
				DocID:     chunk.DocID,
				ChunkID:   chunk.ChunkID,
				Path:      chunk.Path,
				Heading:   chunk.Heading,
				LineStart: chunk.LineStart,
				LineEnd:   chunk.LineEnd,
			}); err != nil {
				return err
			}
			for _, link := range extractMarkdownLinks(chunk.Content) {
				targetPath := resolveLinkPath(doc.Path, link)
				targetDoc, ok := documentIndex[targetPath]
				if !ok {
					continue
				}
				citation := domain.Citation{
					DocID:     chunk.DocID,
					ChunkID:   chunk.ChunkID,
					Path:      chunk.Path,
					Heading:   chunk.Heading,
					LineStart: chunk.LineStart,
					LineEnd:   chunk.LineEnd,
				}
				if err := insertGraphEdge(ctx, tx, hashID("edge", nodeID, targetDoc.DocID, link, chunk.ChunkID), nodeID, "doc:"+targetDoc.DocID, "links_to", citation); err != nil {
					return err
				}
				if err := insertGraphEdge(ctx, tx, hashID("edge", chunkNodeID, targetDoc.DocID, link), chunkNodeID, "doc:"+targetDoc.DocID, "links_to", citation); err != nil {
					return err
				}
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit graph rebuild", err)
	}
	return nil
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin records rebuild", err)
	}
	defer tx.Rollback()
	for _, stmt := range []string{
		`DELETE FROM record_citations;`,
		`DELETE FROM record_facts;`,
		`DELETE FROM record_entities;`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return domain.InternalError("reset records projection", err)
		}
	}

	now := s.now().UTC().Format(time.RFC3339Nano)
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
		if _, err := tx.ExecContext(ctx, `
INSERT INTO record_entities (entity_id, entity_type, name, summary, source_doc_id, updated_at)
VALUES (?, ?, ?, ?, ?, ?)`,
			entityID,
			entityType,
			name,
			summary,
			doc.DocID,
			now,
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
		if _, err := tx.ExecContext(ctx, `
INSERT INTO provenance_events (event_id, event_type, ref_kind, ref_id, source_ref, occurred_at, details_json)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
			hashID("event", "record", entityID, now),
			"record_extracted_from_doc",
			"entity",
			entityID,
			"doc:"+doc.DocID,
			now,
			fmt.Sprintf(`{"entity_type":%q,"entity_name":%q}`, entityType, name),
		); err != nil {
			return domain.InternalError("record records provenance event", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit records rebuild", err)
	}
	return nil
}

func (s *Store) loadAllDocuments(ctx context.Context) ([]domain.Document, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT doc_id, path, title, body, headings_json, created_at, updated_at
FROM documents
ORDER BY path`)
	if err != nil {
		return nil, domain.InternalError("query documents", err)
	}
	defer rows.Close()
	docs := []domain.Document{}
	for rows.Next() {
		var (
			doc          domain.Document
			headingsJSON string
			createdAt    string
			updatedAt    string
		)
		if err := rows.Scan(&doc.DocID, &doc.Path, &doc.Title, &doc.Body, &headingsJSON, &createdAt, &updatedAt); err != nil {
			return nil, domain.InternalError("scan document", err)
		}
		_ = json.Unmarshal([]byte(headingsJSON), &doc.Headings)
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
	defer rows.Close()
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

func normalizePath(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", domain.ValidationError("path is required", nil)
	}
	if filepath.IsAbs(raw) {
		return "", domain.ValidationError("path must be repo-relative to the vault root", map[string]any{"path": raw})
	}
	clean := path.Clean(filepath.ToSlash(raw))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", domain.ValidationError("path must stay inside the vault root", map[string]any{"path": raw})
	}
	if ext := path.Ext(clean); ext == "" {
		clean += ".md"
	} else if ext != ".md" {
		return "", domain.ValidationError("path must end with .md", map[string]any{"path": raw})
	}
	return clean, nil
}

func parseMarkdown(body string, relPath string) ([]string, []section, map[string]string) {
	lines := strings.Split(body, "\n")
	frontmatter, contentStart := parseFrontmatter(lines)
	headings := []string{}
	sections := []section{}
	type headingInfo struct {
		index int
		level int
		title string
	}
	infos := []headingInfo{}
	for idx := contentStart; idx < len(lines); idx++ {
		matches := headingPattern.FindStringSubmatch(lines[idx])
		if len(matches) == 0 {
			continue
		}
		title := strings.TrimSpace(matches[2])
		headings = append(headings, title)
		infos = append(infos, headingInfo{
			index: idx,
			level: len(matches[1]),
			title: title,
		})
	}
	if len(infos) == 0 {
		title := documentTitle(relPath, body, nil, frontmatter)
		return []string{}, []section{{
			Heading:   title,
			Level:     1,
			Content:   strings.TrimSpace(body),
			LineStart: contentStart + 1,
			LineEnd:   len(lines),
		}}, frontmatter
	}
	if infos[0].index > contentStart {
		preamble := strings.TrimSpace(strings.Join(lines[contentStart:infos[0].index], "\n"))
		if preamble != "" {
			sections = append(sections, section{
				Heading:   documentTitle(relPath, body, headings, frontmatter),
				Level:     1,
				Content:   preamble,
				LineStart: contentStart + 1,
				LineEnd:   infos[0].index,
			})
		}
	}
	for i, info := range infos {
		end := len(lines)
		if i+1 < len(infos) {
			end = infos[i+1].index
		}
		sections = append(sections, section{
			Heading:   info.title,
			Level:     info.level,
			Content:   strings.TrimSpace(strings.Join(lines[info.index:end], "\n")),
			LineStart: info.index + 1,
			LineEnd:   end,
		})
	}
	return headings, sections, frontmatter
}

func parseFrontmatter(lines []string) (map[string]string, int) {
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return map[string]string{}, 0
	}
	frontmatter := map[string]string{}
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			return frontmatter, idx + 1
		}
		key, value, ok := strings.Cut(lines[idx], ":")
		if !ok {
			continue
		}
		frontmatter[strings.TrimSpace(strings.ToLower(key))] = strings.TrimSpace(value)
	}
	return map[string]string{}, 0
}

func documentTitle(relPath string, body string, headings []string, frontmatter map[string]string) string {
	return resolvedDocumentTitle(relPath, body, headings, frontmatter, "", "")
}

func resolvedDocumentTitle(relPath string, body string, headings []string, frontmatter map[string]string, preferredTitle string, existingTitle string) string {
	if title := strings.TrimSpace(preferredTitle); title != "" {
		return title
	}
	if title := strings.TrimSpace(frontmatter["title"]); title != "" {
		return title
	}
	if len(headings) > 0 {
		return headings[0]
	}
	if title := strings.TrimSpace(existingTitle); title != "" {
		return title
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "---" || strings.Contains(line, ":") {
			continue
		}
		return strings.TrimPrefix(line, "# ")
	}
	return strings.TrimSuffix(path.Base(relPath), path.Ext(relPath))
}

func replaceSection(body string, targetHeading string, content string) (string, error) {
	lines := strings.Split(body, "\n")
	_, contentStart := parseFrontmatter(lines)
	targetHeading = strings.TrimSpace(targetHeading)
	for idx := contentStart; idx < len(lines); idx++ {
		matches := headingPattern.FindStringSubmatch(lines[idx])
		if len(matches) == 0 {
			continue
		}
		if strings.TrimSpace(matches[2]) != targetHeading {
			continue
		}
		level := len(matches[1])
		end := len(lines)
		for next := idx + 1; next < len(lines); next++ {
			nextMatches := headingPattern.FindStringSubmatch(lines[next])
			if len(nextMatches) == 0 {
				continue
			}
			if len(nextMatches[1]) <= level {
				end = next
				break
			}
		}
		replacement := []string{lines[idx]}
		replacement = append(replacement, strings.Split(strings.TrimSpace(content), "\n")...)
		updated := append([]string{}, lines[:idx]...)
		updated = append(updated, replacement...)
		updated = append(updated, lines[end:]...)
		return strings.TrimRight(strings.Join(updated, "\n"), "\n") + "\n", nil
	}
	return "", domain.NotFoundError("heading", targetHeading)
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

func docIDForPath(relPath string) string {
	return hashID("doc", relPath)
}

func chunkIDForSection(docID string, sec section) string {
	return hashID("chunk", docID, sec.Heading, sec.Content)
}

func hashID(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:8])
}

func encodeCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeCursor(cursor string) int {
	if cursor == "" {
		return 0
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0
	}
	value, err := strconv.Atoi(string(decoded))
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func ftsExpression(text string) string {
	parts := wordPattern.FindAllString(strings.ToLower(text), -1)
	if len(parts) == 0 {
		return `"` + strings.ReplaceAll(strings.TrimSpace(text), `"`, `""`) + `"`
	}
	return strings.Join(parts, " ")
}

func embedText(text string) []float64 {
	vector := make([]float64, embeddingDimensions)
	for _, token := range wordPattern.FindAllString(strings.ToLower(text), -1) {
		sum := sha256.Sum256([]byte(token))
		index := int(sum[0]) % embeddingDimensions
		sign := 1.0
		if sum[1]%2 == 1 {
			sign = -1.0
		}
		vector[index] += sign
	}
	var norm float64
	for _, value := range vector {
		norm += value * value
	}
	if norm == 0 {
		return vector
	}
	norm = math.Sqrt(norm)
	for idx := range vector {
		vector[idx] /= norm
	}
	return vector
}

func cosineSimilarity(left []float64, right []float64) float64 {
	if len(left) != len(right) || len(left) == 0 {
		return 0
	}
	var score float64
	for idx := range left {
		score += left[idx] * right[idx]
	}
	return score
}

func extractMarkdownLinks(content string) []string {
	matches := linkPattern.FindAllStringSubmatch(content, -1)
	links := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		links = append(links, match[1])
	}
	return links
}

func resolveLinkPath(docPath string, target string) string {
	target = strings.TrimSpace(target)
	if target == "" || strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return ""
	}
	target = strings.Split(target, "#")[0]
	if target == "" {
		return ""
	}
	resolved := path.Clean(path.Join(path.Dir(docPath), target))
	if path.Ext(resolved) == "" {
		resolved += ".md"
	}
	return resolved
}

func documentsByPath(documents []domain.Document) map[string]domain.Document {
	result := make(map[string]domain.Document, len(documents))
	for _, doc := range documents {
		result[doc.Path] = doc
	}
	return result
}

func insertGraphNode(ctx context.Context, tx *sql.Tx, nodeID, nodeType, label string, citation domain.Citation) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO graph_nodes (node_id, type, label, evidence_doc_id, evidence_chunk_id, evidence_path, evidence_heading, evidence_line_start, evidence_line_end)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		nodeID,
		nodeType,
		label,
		citation.DocID,
		citation.ChunkID,
		citation.Path,
		nullIfEmpty(citation.Heading),
		citation.LineStart,
		citation.LineEnd,
	)
	if err != nil {
		return domain.InternalError("insert graph node", err)
	}
	return nil
}

func insertGraphEdge(ctx context.Context, tx *sql.Tx, edgeID, fromNodeID, toNodeID, kind string, citation domain.Citation) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO graph_edges (edge_id, from_node_id, to_node_id, kind, evidence_doc_id, evidence_chunk_id, evidence_path, evidence_heading, evidence_line_start, evidence_line_end)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		edgeID,
		fromNodeID,
		toNodeID,
		kind,
		citation.DocID,
		citation.ChunkID,
		citation.Path,
		nullIfEmpty(citation.Heading),
		citation.LineStart,
		citation.LineEnd,
	)
	if err != nil {
		return domain.InternalError("insert graph edge", err)
	}
	return nil
}

func extractRecordProjection(body string) (map[string]string, []domain.RecordFact, bool) {
	lines := strings.Split(body, "\n")
	frontmatter, contentStart := parseFrontmatter(lines)
	if frontmatter["entity_type"] == "" && frontmatter["entity_name"] == "" {
		return nil, nil, false
	}
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
	return frontmatter, facts, true
}

func firstSummaryParagraph(body string) string {
	for _, block := range strings.Split(body, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" || strings.HasPrefix(block, "---") || strings.HasPrefix(block, "#") {
			continue
		}
		return firstNRunes(block, 240)
	}
	return ""
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

func firstNRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func ensureDir(dir string) error {
	return osMkdirAll(dir, 0o755)
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func mustParseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
