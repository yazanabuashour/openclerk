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

type storedProjectionState struct {
	ProjectionVersion string
	UpdatedAt         time.Time
}

type section struct {
	Heading   string
	Level     int
	Content   string
	LineStart int
	LineEnd   int
}

type serviceProjection struct {
	ServiceID string
	Name      string
	Status    string
	Owner     string
	Interface string
	Facts     []domain.ServiceFact
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
		Extensions:  []string{"provenance"},
	}
	if supportsHybridSearch(s.backend) && s.embeddingProvider != "" {
		capabilities.SearchModes = []string{"lexical", "vector", "hybrid"}
	}
	if supportsGraph(s.backend) {
		capabilities.Extensions = append(capabilities.Extensions, "graph")
	}
	if supportsRecords(s.backend) {
		capabilities.Extensions = append(capabilities.Extensions, "records")
	}
	if supportsServices(s.backend) {
		capabilities.Extensions = append(capabilities.Extensions, "services")
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
	if (query.MetadataKey == "") != (query.MetadataValue == "") {
		return domain.SearchResult{}, domain.ValidationError("metadataKey and metadataValue must be provided together", nil)
	}
	if supportsHybridSearch(s.backend) && s.embeddingProvider != "" {
		return s.hybridSearch(ctx, query, limit, decodeCursor(query.Cursor))
	}
	return s.lexicalSearch(ctx, query, limit, decodeCursor(query.Cursor))
}

func (s *Store) ListDocuments(ctx context.Context, query domain.DocumentListQuery) (domain.DocumentListResult, error) {
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

func (s *Store) GetDocumentLinks(ctx context.Context, docID string) (domain.DocumentLinks, error) {
	if !supportsGraph(s.backend) {
		return domain.DocumentLinks{}, domain.UnsupportedError("document links", s.backend)
	}
	if _, err := s.GetDocument(ctx, docID); err != nil {
		return domain.DocumentLinks{}, err
	}

	outgoing, err := s.loadDocumentLinks(ctx, `
SELECT d.doc_id, d.path, d.title, ge.evidence_doc_id, ge.evidence_chunk_id, ge.evidence_path, ge.evidence_heading, ge.evidence_line_start, ge.evidence_line_end
FROM graph_edges ge
JOIN documents d ON d.doc_id = SUBSTR(ge.to_node_id, 5)
WHERE ge.kind = 'links_to' AND ge.from_node_id = ? AND ge.to_node_id LIKE 'doc:%'
ORDER BY d.path`, "doc:"+docID)
	if err != nil {
		return domain.DocumentLinks{}, err
	}
	incoming, err := s.loadDocumentLinks(ctx, `
SELECT d.doc_id, d.path, d.title, ge.evidence_doc_id, ge.evidence_chunk_id, ge.evidence_path, ge.evidence_heading, ge.evidence_line_start, ge.evidence_line_end
FROM graph_edges ge
JOIN documents d ON d.doc_id = SUBSTR(ge.from_node_id, 5)
WHERE ge.kind = 'links_to' AND ge.to_node_id = ? AND ge.from_node_id LIKE 'doc:%'
ORDER BY d.path`, "doc:"+docID)
	if err != nil {
		return domain.DocumentLinks{}, err
	}
	return domain.DocumentLinks{DocID: docID, Outgoing: outgoing, Incoming: incoming}, nil
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
	if !supportsGraph(s.backend) {
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
	defer func() {
		_ = rows.Close()
	}()

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

func (s *Store) ListProjectionStates(ctx context.Context, query domain.ProjectionStateQuery) (domain.ProjectionStateResult, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return domain.ProjectionStateResult{}, domain.ValidationError("limit must be between 1 and 100", map[string]any{"limit": limit})
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

	pageInfo := domain.PageInfo{}
	if len(projections) > limit {
		pageInfo.HasMore = true
		pageInfo.NextCursor = encodeCursor(offset + limit)
		projections = projections[:limit]
	}
	return domain.ProjectionStateResult{Projections: projections, PageInfo: pageInfo}, nil
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
			metadata_json TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS document_metadata (
			doc_id TEXT NOT NULL,
			key_name TEXT NOT NULL,
			value_text TEXT NOT NULL,
			PRIMARY KEY (doc_id, key_name),
			FOREIGN KEY (doc_id) REFERENCES documents(doc_id) ON DELETE CASCADE
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
		`CREATE TABLE IF NOT EXISTS service_records (
			service_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			owner TEXT NOT NULL,
			service_interface TEXT NOT NULL,
			summary TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS service_facts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			service_id TEXT NOT NULL,
			key_name TEXT NOT NULL,
			value_text TEXT NOT NULL,
			observed_at TEXT,
			FOREIGN KEY (service_id) REFERENCES service_records(service_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS service_citations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			service_id TEXT NOT NULL,
			source_doc_id TEXT NOT NULL,
			source_chunk_id TEXT NOT NULL,
			source_path TEXT NOT NULL,
			source_heading TEXT,
			source_line_start INTEGER NOT NULL,
			source_line_end INTEGER NOT NULL,
			FOREIGN KEY (service_id) REFERENCES service_records(service_id) ON DELETE CASCADE
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
		`CREATE TABLE IF NOT EXISTS projection_states (
			projection_name TEXT NOT NULL,
			ref_kind TEXT NOT NULL,
			ref_id TEXT NOT NULL,
			source_ref TEXT NOT NULL,
			freshness TEXT NOT NULL,
			projection_version TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			details_json TEXT NOT NULL,
			PRIMARY KEY (projection_name, ref_kind, ref_id)
		);`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return domain.InternalError("initialize sqlite schema", err)
		}
	}
	if err := ensureColumn(ctx, s.db, "documents", "metadata_json", "TEXT NOT NULL DEFAULT '{}'"); err != nil {
		return err
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
	if err := s.rebuildGraph(ctx); err != nil {
		return err
	}
	if err := s.rebuildRecords(ctx); err != nil {
		return err
	}
	return s.rebuildServices(ctx)
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

	if contentChanged {
		if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
			EventID:    hashID("event", eventType, relPath, now.Format(time.RFC3339Nano)),
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
			EventID:    hashID("event", "projection_invalidated", "graph", docID, now.Format(time.RFC3339Nano)),
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
				EventID:    hashID("event", "projection_invalidated", "records", docID, now.Format(time.RFC3339Nano)),
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
				EventID:    hashID("event", "projection_invalidated", "services", docID, now.Format(time.RFC3339Nano)),
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

	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit document sync", err)
	}
	if err := s.rebuildGraph(ctx); err != nil {
		return err
	}
	if err := s.rebuildRecords(ctx); err != nil {
		return err
	}
	return s.rebuildServices(ctx)
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

func (s *Store) hybridSearch(ctx context.Context, query domain.SearchQuery, limit int, offset int) (domain.SearchResult, error) {
	lexical, err := s.lexicalSearch(ctx, query, max(limit*2, 20), 0)
	if err != nil {
		return domain.SearchResult{}, err
	}
	queryVector := embedText(query.Text)
	baseQuery := `
SELECT c.chunk_id, c.doc_id, d.title, c.path, c.heading, c.content, c.line_start, c.line_end, e.vector_json
FROM chunks c
JOIN documents d ON d.doc_id = c.doc_id
JOIN embeddings e ON e.chunk_id = c.chunk_id`
	whereClause, args := filteredDocumentClauses(query)
	sqlQuery := baseQuery
	if whereClause != "" {
		sqlQuery += "\nWHERE " + whereClause
	}
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return domain.SearchResult{}, domain.InternalError("query embeddings", err)
	}
	defer func() {
		_ = rows.Close()
	}()

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
		hit.Snippet = snippetForSearch(content, query.Text)
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
	previousStates, err := s.loadProjectionStateSnapshots(ctx, "graph")
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin graph rebuild", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for _, stmt := range []string{
		`DELETE FROM graph_edges;`,
		`DELETE FROM graph_nodes;`,
		`DELETE FROM projection_states WHERE projection_name = 'graph';`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return domain.InternalError("reset graph projection", err)
		}
	}

	now := s.now().UTC()
	versionInputs := make(map[string][]string, len(documents))
	for _, doc := range documents {
		versionInputs[doc.DocID] = append(versionInputs[doc.DocID],
			"doc:"+doc.DocID,
			"path:"+doc.Path,
			"updated:"+doc.UpdatedAt.UTC().Format(time.RFC3339Nano),
		)
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
				versionInputs[doc.DocID] = append(versionInputs[doc.DocID],
					fmt.Sprintf("out:%s:%s:%s:%d:%d", targetDoc.DocID, citation.ChunkID, citation.Path, citation.LineStart, citation.LineEnd),
				)
				versionInputs[targetDoc.DocID] = append(versionInputs[targetDoc.DocID],
					fmt.Sprintf("in:%s:%s:%s:%d:%d", doc.DocID, citation.ChunkID, citation.Path, citation.LineStart, citation.LineEnd),
				)
			}
		}
	}
	for _, doc := range documents {
		markers := append([]string(nil), versionInputs[doc.DocID]...)
		sort.Strings(markers)
		version := hashID("graph", doc.DocID, strings.Join(markers, "|"))
		stateUpdatedAt := now
		if previous, ok := previousStates[doc.DocID]; ok && previous.ProjectionVersion == version {
			stateUpdatedAt = previous.UpdatedAt
		}
		if err := upsertProjectionState(ctx, tx, domain.ProjectionState{
			Projection:        "graph",
			RefKind:           "document",
			RefID:             doc.DocID,
			SourceRef:         "doc:" + doc.DocID,
			Freshness:         "fresh",
			ProjectionVersion: version,
			UpdatedAt:         stateUpdatedAt,
			Details: map[string]string{
				"path": doc.Path,
			},
		}); err != nil {
			return err
		}
		if previous, ok := previousStates[doc.DocID]; !ok || previous.ProjectionVersion != version {
			if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
				EventID:    hashID("event", "projection_refreshed", "graph", doc.DocID, now.Format(time.RFC3339Nano)),
				EventType:  "projection_refreshed",
				RefKind:    "projection",
				RefID:      "graph:" + doc.DocID,
				SourceRef:  "doc:" + doc.DocID,
				OccurredAt: now,
				Details: map[string]string{
					"projection": "graph",
					"path":       doc.Path,
					"version":    version,
				},
			}); err != nil {
				return err
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
				EventID:    hashID("event", "projection_refreshed", "records", entityID, now.Format(time.RFC3339Nano)),
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
				EventID:    hashID("event", "projection_refreshed", "services", projected.ServiceID, now.Format(time.RFC3339Nano)),
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

func (s *Store) loadDocumentLinks(ctx context.Context, query string, nodeID string) ([]domain.DocumentLink, error) {
	rows, err := s.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, domain.InternalError("query document links", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	links := []domain.DocumentLink{}
	for rows.Next() {
		var (
			link       domain.DocumentLink
			citation   domain.Citation
			headingRaw sql.NullString
		)
		if err := rows.Scan(
			&link.DocID,
			&link.Path,
			&link.Title,
			&citation.DocID,
			&citation.ChunkID,
			&citation.Path,
			&headingRaw,
			&citation.LineStart,
			&citation.LineEnd,
		); err != nil {
			return nil, domain.InternalError("scan document link", err)
		}
		citation.Heading = headingRaw.String
		link.Citations = []domain.Citation{citation}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.InternalError("iterate document links", err)
	}
	return links, nil
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

func ensureColumn(ctx context.Context, db *sql.DB, table string, column string, definition string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return domain.InternalError("inspect sqlite table", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var (
			cid        int
			name       string
			typ        string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultVal, &primaryKey); err != nil {
			return domain.InternalError("scan sqlite table info", err)
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return domain.InternalError("iterate sqlite table info", err)
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)); err != nil {
		return domain.InternalError("alter sqlite table", err)
	}
	return nil
}

func supportsHybridSearch(backend domain.BackendKind) bool {
	return backend == domain.BackendOpenClerk
}

func supportsGraph(backend domain.BackendKind) bool {
	return backend == domain.BackendOpenClerk
}

func supportsRecords(backend domain.BackendKind) bool {
	return backend == domain.BackendOpenClerk
}

func supportsServices(backend domain.BackendKind) bool {
	return backend == domain.BackendOpenClerk
}

func insertProvenanceEvent(ctx context.Context, tx *sql.Tx, event domain.ProvenanceEvent) error {
	detailsJSON, err := json.Marshal(event.Details)
	if err != nil {
		return domain.InternalError("encode provenance event details", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO provenance_events (event_id, event_type, ref_kind, ref_id, source_ref, occurred_at, details_json)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.EventID,
		event.EventType,
		event.RefKind,
		event.RefID,
		event.SourceRef,
		event.OccurredAt.UTC().Format(time.RFC3339Nano),
		string(detailsJSON),
	); err != nil {
		return domain.InternalError("insert provenance event", err)
	}
	return nil
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
	return frontmatter, extractRecordFacts(lines, contentStart), true
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
