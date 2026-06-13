package runner

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	graphContextValidationBoundaries = "read-only report; uses installed OpenClerk document/retrieval JSON evidence through the runner only; no writes, no direct vault inspection, no direct SQLite, no source-built runners, no HTTP/MCP bypasses, no unsupported transports, no semantic-label graph truth, no hidden authority ranking, and no graph memory"
	graphContextAuthorityLimits      = "canonical markdown remains semantic relationship authority; graph edges, links, backlinks, provenance, and projection freshness are derived navigation evidence with citations and freshness, not independent truth or authority ranking"
)

func runGraphContextReport(ctx context.Context, client *runclient.Client, options GraphContextOptions) (GraphContextReport, error) {
	limit, err := boundedRunnerLimit(options.Limit, 10, 50, "graph_context")
	if err != nil {
		return GraphContextReport{}, err
	}

	source, sourceSelection, sourceCandidates, evidenceInspected, err := resolveGraphContextSource(ctx, client, options, limit)
	if err != nil {
		return GraphContextReport{}, err
	}

	links, err := client.GetDocumentLinks(ctx, source.DocID, limit)
	if err != nil {
		return GraphContextReport{}, err
	}
	convertedLinks := toDocumentLinksResult(links)
	evidenceInspected = append(evidenceInspected,
		fmt.Sprintf("document_links:outgoing:%d", len(convertedLinks.Outgoing)),
		fmt.Sprintf("document_links:incoming:%d", len(convertedLinks.Incoming)),
	)

	graph, err := client.GraphNeighborhood(ctx, domain.GraphNeighborhoodInput{DocID: source.DocID, Limit: limit})
	if err != nil {
		return GraphContextReport{}, err
	}
	convertedGraph := toGraphNeighborhood(graph)
	evidenceInspected = append(evidenceInspected, fmt.Sprintf("graph_neighborhood:nodes:%d", len(convertedGraph.Nodes)), fmt.Sprintf("graph_neighborhood:edges:%d", len(convertedGraph.Edges)))

	projections, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: "graph",
		RefKind:    "document",
		RefID:      source.DocID,
		Limit:      limit,
	})
	if err != nil {
		return GraphContextReport{}, err
	}
	convertedProjections := toProjectionStateList(projections)
	evidenceInspected = append(evidenceInspected, "projection:graph", "projection_freshness:"+projectionListFreshnessSummary(&convertedProjections))

	provenance, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "document",
		RefID:   source.DocID,
		Limit:   limit,
	})
	if err != nil {
		return GraphContextReport{}, err
	}
	convertedProvenance := toProvenanceEventList(provenance)
	evidenceInspected = append(evidenceInspected, fmt.Sprintf("provenance_events:%d", len(convertedProvenance.Events)))

	relationshipText := canonicalRelationshipText(source, limit)
	evidenceInspected = append(evidenceInspected, fmt.Sprintf("canonical_relationship_text:%d", len(relationshipText)))

	report := GraphContextReport{
		Query:                     options.Query,
		Path:                      options.Path,
		DocID:                     options.DocID,
		PathPrefix:                options.PathPrefix,
		SourceDocument:            graphContextSourceSummary(source),
		SourceSelection:           sourceSelection,
		SourceCandidates:          sourceCandidates,
		CanonicalRelationshipText: relationshipText,
		Links:                     &convertedLinks,
		Graph:                     &convertedGraph,
		GraphProjection:           &convertedProjections,
		Provenance:                &convertedProvenance,
		CandidateSurfaces:         graphContextCandidates(),
		Recommendation:            "promote graph_context_report for routine read-only relationship graph inspection; keep document_links, graph_neighborhood, get_document, provenance_events, and projection_states for explicit drill-down",
		SafetyPass:                "passes: report is read-only, local-first, runner-only, and keeps relationship meaning in cited canonical markdown while exposing derived graph evidence as navigation context",
		CapabilityPass:            "passes: packages source identity, canonical relationship text, links/backlinks, nearby structural graph context, graph projection freshness, and document provenance refs",
		UXQuality:                 "improves routine relationship inspection by replacing multi-step search/list/get/links/graph/projection/provenance choreography with one retrieval action when the source doc_id, path, or query is known",
		EvidencePosture:           "evidence comes from canonical markdown, citation-bearing links and graph edges, graph projection freshness, and document provenance; no semantic graph labels or hidden authority rank are claimed",
		ValidationBoundaries:      graphContextValidationBoundaries,
		AuthorityLimits:           graphContextAuthorityLimits,
		EvidenceInspected:         evidenceInspected,
	}
	report.ProvenanceRefs = graphContextProvenanceRefs(report)
	report.AgentHandoff = graphContextHandoff(report)
	return report, nil
}

func resolveGraphContextSource(ctx context.Context, client *runclient.Client, options GraphContextOptions, limit int) (domain.Document, string, []SearchHit, []string, error) {
	if options.DocID != "" {
		document, err := client.GetDocument(ctx, options.DocID)
		if err != nil {
			return domain.Document{}, "", nil, nil, err
		}
		return document, "doc_id_exact_match", nil, []string{"get_document:" + options.DocID}, nil
	}

	if options.Path != "" {
		if !isGraphContextRepoRelativeMarkdownPath(options.Path) {
			return domain.Document{}, "", nil, nil, domain.ValidationError("graph_context.path must be a repo-relative markdown path", map[string]any{"path": options.Path})
		}
		documents, err := client.ListDocuments(ctx, domain.DocumentListQuery{PathPrefix: options.Path, Limit: 100})
		if err != nil {
			return domain.Document{}, "", nil, nil, err
		}
		for _, candidate := range documents.Documents {
			if candidate.Path != options.Path {
				continue
			}
			document, err := client.GetDocument(ctx, candidate.DocID)
			if err != nil {
				return domain.Document{}, "", nil, nil, err
			}
			return document, "path_exact_match", nil, []string{"list_documents:" + options.Path, "get_document:" + options.Path}, nil
		}
		return domain.Document{}, "", nil, nil, domain.NotFoundError("document path", options.Path)
	}

	search, err := client.Search(ctx, domain.SearchQuery{Text: options.Query, PathPrefix: options.PathPrefix, Limit: limit})
	if err != nil {
		return domain.Document{}, "", nil, nil, err
	}
	convertedSearch := toSearchResult(search)
	if len(search.Hits) == 0 {
		return domain.Document{}, "", nil, nil, domain.NotFoundError("document query", options.Query)
	}
	document, err := client.GetDocument(ctx, search.Hits[0].DocID)
	if err != nil {
		return domain.Document{}, "", nil, nil, err
	}
	evidence := []string{"search:" + options.Query, fmt.Sprintf("source_candidates:%d", len(search.Hits)), "get_document:" + document.Path}
	if options.PathPrefix != "" {
		evidence = append(evidence, "path_prefix:"+options.PathPrefix)
	}
	return document, "query_first_cited_search_hit_visible_candidate_not_authority_ranking", convertedSearch.Hits, evidence, nil
}

func isGraphContextRepoRelativeMarkdownPath(value string) bool {
	clean := path.Clean(value)
	return value == clean &&
		!strings.HasPrefix(value, "/") &&
		value != "." &&
		value != ".." &&
		!strings.HasPrefix(value, "../") &&
		!strings.Contains(value, "/../") &&
		strings.HasSuffix(value, ".md")
}

func graphContextSourceSummary(document domain.Document) *DocumentSummary {
	return &DocumentSummary{
		DocID:     document.DocID,
		Path:      document.Path,
		Title:     document.Title,
		Metadata:  cloneStringMap(document.Metadata),
		UpdatedAt: document.UpdatedAt,
	}
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func canonicalRelationshipText(document domain.Document, limit int) []CanonicalRelationshipText {
	lines := strings.Split(document.Body, "\n")
	text := make([]CanonicalRelationshipText, 0, min(limit, 12))
	currentHeading := ""
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			currentHeading = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			continue
		}
		if !isCanonicalRelationshipLine(trimmed) {
			continue
		}
		text = append(text, CanonicalRelationshipText{
			Text: trimmed,
			Citation: Citation{
				DocID:     document.DocID,
				Path:      document.Path,
				Heading:   currentHeading,
				LineStart: index + 1,
				LineEnd:   index + 1,
			},
		})
		if len(text) >= limit {
			break
		}
	}
	return text
}

func isCanonicalRelationshipLine(line string) bool {
	if line == "" {
		return false
	}
	lower := strings.ToLower(line)
	if strings.Contains(lower, "](") {
		return true
	}
	for _, marker := range []string{
		"requires",
		"supersedes",
		"superseded by",
		"related to",
		"operationalizes",
		"depends on",
		"dependency",
		"blocks",
		"blocked by",
		"links to",
		"link to",
		"references",
		"source_refs",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func graphContextProvenanceRefs(report GraphContextReport) []string {
	refs := []string{}
	if report.SourceDocument != nil {
		refs = append(refs, "document:"+report.SourceDocument.DocID)
	}
	if report.GraphProjection != nil {
		for _, projection := range report.GraphProjection.Projections {
			refs = append(refs, "projection:"+projection.Projection+":"+projection.RefKind+":"+projection.RefID+":"+projection.Freshness)
		}
	}
	if report.Provenance != nil {
		for _, event := range report.Provenance.Events {
			refs = append(refs, "provenance:"+event.EventType+":"+event.RefKind+":"+event.RefID)
		}
	}
	return refs
}

func graphContextHandoff(report GraphContextReport) *AgentHandoff {
	sourcePath := ""
	sourceDocID := ""
	if report.SourceDocument != nil {
		sourcePath = report.SourceDocument.Path
		sourceDocID = report.SourceDocument.DocID
	}
	outgoing := 0
	incoming := 0
	if report.Links != nil {
		outgoing = len(report.Links.Outgoing)
		incoming = len(report.Links.Incoming)
	}
	graphNodes := 0
	graphEdges := 0
	if report.Graph != nil {
		graphNodes = len(report.Graph.Nodes)
		graphEdges = len(report.Graph.Edges)
	}
	provenanceCount := 0
	if report.Provenance != nil {
		provenanceCount = len(report.Provenance.Events)
	}
	evidence := []string{
		"source_path=" + sourcePath,
		"doc_id=" + sourceDocID,
		fmt.Sprintf("canonical_relationship_text=%d", len(report.CanonicalRelationshipText)),
		fmt.Sprintf("outgoing_links=%d", outgoing),
		fmt.Sprintf("incoming_backlinks=%d", incoming),
		fmt.Sprintf("graph_nodes=%d", graphNodes),
		fmt.Sprintf("graph_edges=%d", graphEdges),
		"graph_projection_freshness=" + projectionListFreshnessSummary(report.GraphProjection),
		fmt.Sprintf("provenance_events=%d", provenanceCount),
		"read_only=true",
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"graph_context_report returned source %s with %d canonical relationship text refs, %d outgoing links, %d incoming backlinks, %d graph nodes, %d graph edges, %s, %d provenance events, read-only behavior, validation boundaries, and authority limits",
			sourcePath,
			len(report.CanonicalRelationshipText),
			outgoing,
			incoming,
			graphNodes,
			graphEdges,
			projectionListFreshnessSummary(report.GraphProjection),
			provenanceCount,
		),
		Evidence:                    evidence,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine relationship graph context; use get_document, document_links, graph_neighborhood, provenance_events, and projection_states directly only for explicit drill-down or runner rejection repair",
	}
}

func graphContextCandidates() []GraphContextCandidate {
	return []GraphContextCandidate{
		{
			Surface:    "current_primitives_plus_help",
			Status:     "available_reference",
			Safety:     "passes when agents use search/list/get, document_links, graph_neighborhood, provenance_events, and projection_states through the installed runner",
			Capability: "passes for relationship inspection because canonical markdown, links/backlinks, structural graph context, provenance, and graph freshness are visible",
			UXQuality:  "ceremonial for routine use because agents must stitch many primitive calls and policy boundaries before answering",
			Implementation: []string{
				"existing openclerk retrieval/document primitives",
				"retrieval help and skill guidance",
			},
		},
		{
			Surface:    "graph_context_report",
			Status:     "promoted_read_only",
			Safety:     "passes because it packages only current runner evidence and declares graph state derived navigation evidence",
			Capability: "passes for source identity, canonical relationship text, links/backlinks, graph neighborhood, graph freshness, and provenance refs",
			UXQuality:  "one retrieval action replaces the high-step relationship inspection workflow for routine graph context answers",
			Implementation: []string{
				"runner JSON action under openclerk retrieval",
				"agent_handoff for final-answer evidence",
				"primitive actions remain available for drill-down",
			},
		},
		{
			Surface:    "no_new_surface",
			Status:     "not_selected",
			Safety:     "passes by avoiding API growth",
			Capability: "technically sufficient but leaves relationship-context packaging to the assistant",
			UXQuality:  "does not solve the observed ceremony and latency pressure for normal users asking for relationship context",
			Implementation: []string{
				"no runner change",
				"keep graph semantics as reference pressure only",
			},
		},
	}
}
