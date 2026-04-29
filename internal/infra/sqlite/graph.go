package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"path"
	"sort"
	"strings"
	"time"
)

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
				EventID:    hashID("event", "projection_refreshed", "graph", doc.DocID, version, now.Format(time.RFC3339Nano)),
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

func supportsGraph(backend domain.BackendKind) bool {
	return backend == domain.BackendOpenClerk
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
INSERT OR IGNORE INTO graph_edges (edge_id, from_node_id, to_node_id, kind, evidence_doc_id, evidence_chunk_id, evidence_path, evidence_heading, evidence_line_start, evidence_line_end)
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
