package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	graphRelationshipValidationBoundaries = "read-only report; uses installed OpenClerk document/retrieval JSON evidence through the runner only; no writes, no direct vault inspection, no direct SQLite, no source-built runners, no HTTP/MCP bypasses, no unsupported transports, no semantic-label graph truth, no hidden authority ranking, and no graph memory"
	graphRelationshipAuthorityLimits      = "canonical markdown remains relationship authority; typed relationship candidates are labels suggested from cited markdown wording, and derived graph edges/backlinks are navigation evidence, not independent semantic truth, contradiction proof, or authority ranking"
)

func runGraphRelationshipReport(ctx context.Context, client *runclient.Client, options GraphRelationshipOptions) (GraphRelationshipReport, error) {
	contextReport, err := runGraphContextReport(ctx, client, GraphContextOptions(options))
	if err != nil {
		return GraphRelationshipReport{}, err
	}
	relationshipPaths := graphRelationshipPaths(contextReport)
	directRelationships := graphDirectRelationshipEvidence(contextReport)
	derivedRelationships := graphDerivedRelationshipEvidence(contextReport)
	typedCandidates := graphTypedRelationshipCandidates(contextReport.CanonicalRelationshipText)
	auditFindings := graphRelationshipAuditFindings(contextReport)
	evidenceInspected := append([]string{}, contextReport.EvidenceInspected...)
	evidenceInspected = append(evidenceInspected,
		fmt.Sprintf("relationship_paths:%d", len(relationshipPaths)),
		fmt.Sprintf("direct_relationships:%d", len(directRelationships)),
		fmt.Sprintf("derived_relationships:%d", len(derivedRelationships)),
		fmt.Sprintf("typed_relationship_candidates:%d", len(typedCandidates)),
	)

	report := GraphRelationshipReport{
		Query:                       options.Query,
		Path:                        options.Path,
		DocID:                       options.DocID,
		PathPrefix:                  options.PathPrefix,
		SourceDocument:              contextReport.SourceDocument,
		SourceSelection:             contextReport.SourceSelection,
		SourceCandidates:            contextReport.SourceCandidates,
		RelationshipPaths:           relationshipPaths,
		DirectRelationships:         directRelationships,
		DerivedRelationships:        derivedRelationships,
		TypedRelationshipCandidates: typedCandidates,
		AuditFindings:               auditFindings,
		GraphProjection:             contextReport.GraphProjection,
		Provenance:                  contextReport.Provenance,
		ProvenanceRefs:              contextReport.ProvenanceRefs,
		CandidateSurfaces:           graphRelationshipCandidates(),
		Recommendation:              "promote graph_relationship_report as the narrow read-only follow-up to graph_context_report for path finding, direct-vs-derived relationship reporting, typed candidates from canonical markdown, and limited stale/orphaned/contradiction audits",
		SafetyPass:                  "passes: report is read-only, local-first, runner-only, and keeps relationship meaning in cited canonical markdown while labeling graph evidence as derived navigation context",
		CapabilityPass:              "passes: combines graph_context_report evidence with one-hop relationship paths, direct markdown evidence, derived graph/backlink evidence, typed candidates, graph projection freshness, provenance refs, and limited audit findings",
		UXQuality:                   "improves deferred graph tasks by replacing a multi-step graph_context_report plus primitive-drilldown workflow with one relationship-focused retrieval action",
		EvidencePosture:             "evidence comes from canonical relationship text, markdown links/backlinks, structural graph edges with citations, graph projection freshness, and provenance; no semantic graph storage or hidden inference is introduced",
		ValidationBoundaries:        graphRelationshipValidationBoundaries,
		AuthorityLimits:             graphRelationshipAuthorityLimits,
		EvidenceInspected:           evidenceInspected,
	}
	report.AgentHandoff = graphRelationshipHandoff(report)
	return report, nil
}

func graphRelationshipPaths(report GraphContextReport) []GraphRelationshipPath {
	paths := []GraphRelationshipPath{}
	if report.Links == nil {
		return paths
	}
	for _, link := range report.Links.Outgoing {
		paths = append(paths, GraphRelationshipPath{
			Direction: "outgoing",
			DocID:     link.DocID,
			Path:      link.Path,
			Title:     link.Title,
			Evidence:  "direct markdown link from source document",
			Citations: link.Citations,
		})
	}
	for _, link := range report.Links.Incoming {
		paths = append(paths, GraphRelationshipPath{
			Direction: "incoming",
			DocID:     link.DocID,
			Path:      link.Path,
			Title:     link.Title,
			Evidence:  "incoming markdown backlink to source document",
			Citations: link.Citations,
		})
	}
	return paths
}

func graphDirectRelationshipEvidence(report GraphContextReport) []GraphRelationshipEvidence {
	evidence := []GraphRelationshipEvidence{}
	for _, text := range report.CanonicalRelationshipText {
		types := graphRelationshipTypes(text.Text)
		if len(types) == 0 {
			types = []string{"canonical_relationship_text"}
		}
		for _, relationshipType := range types {
			evidence = append(evidence, GraphRelationshipEvidence{
				Kind:             "canonical_markdown_relationship_text",
				RelationshipType: relationshipType,
				Source:           "canonical_markdown",
				EvidenceText:     text.Text,
				Citations:        []Citation{text.Citation},
				Authority:        "direct_canonical_markdown",
			})
		}
	}
	if report.Links != nil {
		for _, link := range report.Links.Outgoing {
			evidence = append(evidence, GraphRelationshipEvidence{
				Kind:      "direct_outgoing_markdown_link",
				Direction: "outgoing",
				Source:    link.Path,
				Citations: link.Citations,
				Authority: "direct_markdown_link",
			})
		}
		for _, link := range report.Links.Incoming {
			evidence = append(evidence, GraphRelationshipEvidence{
				Kind:      "direct_incoming_markdown_backlink",
				Direction: "incoming",
				Source:    link.Path,
				Citations: link.Citations,
				Authority: "direct_markdown_backlink",
			})
		}
	}
	return evidence
}

func graphDerivedRelationshipEvidence(report GraphContextReport) []GraphRelationshipEvidence {
	evidence := []GraphRelationshipEvidence{}
	if report.Graph == nil {
		return evidence
	}
	for _, edge := range report.Graph.Edges {
		if !isRelationshipGraphEdge(edge) {
			continue
		}
		evidence = append(evidence, GraphRelationshipEvidence{
			Kind:              "derived_graph_edge",
			RelationshipType:  edge.Kind,
			Source:            edge.EdgeID,
			Citations:         edge.Citations,
			Authority:         "derived_navigation_evidence",
			InferenceBoundary: "structural graph edge is derived from indexed markdown/link evidence and does not create semantic-label graph truth",
		})
	}
	return evidence
}

func isRelationshipGraphEdge(edge GraphEdge) bool {
	return edge.Kind == "links_to"
}

func graphTypedRelationshipCandidates(text []CanonicalRelationshipText) []GraphRelationshipTypeCandidate {
	candidates := []GraphRelationshipTypeCandidate{}
	for _, entry := range text {
		for _, relationshipType := range graphRelationshipTypes(entry.Text) {
			candidates = append(candidates, GraphRelationshipTypeCandidate{
				RelationshipType: relationshipType,
				EvidenceText:     entry.Text,
				Citation:         entry.Citation,
				Status:           "candidate_from_cited_markdown_text",
				Authority:        "label suggestion only; canonical markdown text remains authority",
			})
		}
	}
	return candidates
}

func graphRelationshipTypes(text string) []string {
	lower := strings.ToLower(text)
	candidates := []struct {
		marker string
		label  string
	}{
		{"superseded by", "superseded_by"},
		{"supersedes", "supersedes"},
		{"requires", "requires"},
		{"depends on", "depends_on"},
		{"dependency", "depends_on"},
		{"blocked by", "blocked_by"},
		{"blocks", "blocks"},
		{"related to", "related_to"},
		{"operationalizes", "operationalizes"},
		{"references", "references"},
		{"source_refs", "source_refs"},
		{"](", "markdown_link"},
	}
	seen := map[string]struct{}{}
	types := []string{}
	for _, candidate := range candidates {
		if !strings.Contains(lower, candidate.marker) {
			continue
		}
		if _, ok := seen[candidate.label]; ok {
			continue
		}
		seen[candidate.label] = struct{}{}
		types = append(types, candidate.label)
	}
	return types
}

func graphRelationshipAuditFindings(report GraphContextReport) []GraphRelationshipAuditFinding {
	findings := []GraphRelationshipAuditFinding{
		graphProjectionFreshnessFinding(report),
		graphOrphanFinding(report),
		graphContradictionFinding(report),
	}
	return findings
}

func graphProjectionFreshnessFinding(report GraphContextReport) GraphRelationshipAuditFinding {
	if report.GraphProjection == nil || len(report.GraphProjection.Projections) == 0 {
		return GraphRelationshipAuditFinding{
			Kind:     "stale_graph_projection",
			Status:   "unknown",
			Evidence: "no graph projection freshness state was returned for the source document",
		}
	}
	for _, projection := range report.GraphProjection.Projections {
		if projection.Freshness != "fresh" {
			return GraphRelationshipAuditFinding{
				Kind:     "stale_graph_projection",
				Status:   "attention",
				Evidence: "graph projection freshness is " + projection.Freshness,
			}
		}
	}
	return GraphRelationshipAuditFinding{
		Kind:     "stale_graph_projection",
		Status:   "clear",
		Evidence: "graph projection freshness is fresh",
	}
}

func graphOrphanFinding(report GraphContextReport) GraphRelationshipAuditFinding {
	outgoing, incoming, edges := 0, 0, 0
	if report.Links != nil {
		outgoing = len(report.Links.Outgoing)
		incoming = len(report.Links.Incoming)
	}
	if report.Graph != nil {
		for _, edge := range report.Graph.Edges {
			if isRelationshipGraphEdge(edge) {
				edges++
			}
		}
	}
	if outgoing == 0 && incoming == 0 && edges == 0 {
		return GraphRelationshipAuditFinding{
			Kind:     "orphaned_graph_context",
			Status:   "attention",
			Evidence: "source document has no outgoing links, incoming backlinks, or graph edges in the returned evidence",
		}
	}
	return GraphRelationshipAuditFinding{
		Kind:     "orphaned_graph_context",
		Status:   "clear",
		Evidence: fmt.Sprintf("source document has %d outgoing links, %d incoming backlinks, and %d graph edges", outgoing, incoming, edges),
	}
}

func graphContradictionFinding(report GraphContextReport) GraphRelationshipAuditFinding {
	for _, text := range report.CanonicalRelationshipText {
		lower := strings.ToLower(text.Text)
		if strings.Contains(lower, "supersedes") && strings.Contains(lower, "superseded by") {
			return GraphRelationshipAuditFinding{
				Kind:      "contradictory_relationship_text",
				Status:    "attention",
				Evidence:  "same canonical relationship line contains both supersedes and superseded by markers",
				Citations: []Citation{text.Citation},
			}
		}
	}
	return GraphRelationshipAuditFinding{
		Kind:     "contradictory_relationship_text",
		Status:   "clear_limited",
		Evidence: "no simple contradictory supersedes/superseded-by marker was found in returned canonical relationship text; broad semantic contradiction detection remains out of scope",
	}
}

func graphRelationshipHandoff(report GraphRelationshipReport) *AgentHandoff {
	sourcePath := ""
	sourceDocID := ""
	if report.SourceDocument != nil {
		sourcePath = report.SourceDocument.Path
		sourceDocID = report.SourceDocument.DocID
	}
	citationCount := graphRelationshipCitationCount(report)
	evidence := []string{
		"source_path=" + sourcePath,
		"doc_id=" + sourceDocID,
		fmt.Sprintf("relationship_paths=%d", len(report.RelationshipPaths)),
		fmt.Sprintf("direct_relationships=%d", len(report.DirectRelationships)),
		fmt.Sprintf("derived_relationships=%d", len(report.DerivedRelationships)),
		fmt.Sprintf("typed_relationship_candidates=%d", len(report.TypedRelationshipCandidates)),
		fmt.Sprintf("audit_findings=%d stale_graph_projection orphaned_graph_context contradictory_relationship_text", len(report.AuditFindings)),
		"graph_projection_freshness=" + projectionListFreshnessSummary(report.GraphProjection),
		"provenance_refs=present",
		fmt.Sprintf("source_citations=%d", citationCount),
		"safety_pass=" + report.SafetyPass,
		"capability_pass=" + report.CapabilityPass,
		"ux_quality=" + report.UXQuality,
		"authority_model=canonical markdown authority; typed candidates are cited suggestions; no semantic-label graph truth, no hidden authority ranking, no graph memory, no durable semantic graph storage",
		"provenance_freshness_posture=graph_projection freshness " + projectionListFreshnessSummary(report.GraphProjection) + " with provenance_refs",
		"workflow_impact=one graph_relationship_report action replaces current_primitives_plus_graph_context_report drilldown for relationship/path, direct-vs-derived, typed-candidate, and limited graph-audit needs",
		"candidate_comparison=current_primitives_plus_graph_context_report available_reference; graph_relationship_report promote; split_specialized_reports not_selected",
		"decision=promote graph_relationship_report",
		"follow_up_needs=no follow-up work is required for the deferred relationship/path, direct-vs-derived, typed-candidate, or limited graph-audit needs",
		"read_only=true",
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"graph_relationship_report returned source %s with %d relationship_paths, %d direct_relationships, %d derived_relationships, %d typed_relationship_candidates, %d audit_findings including stale_graph_projection, orphaned_graph_context, and contradictory_relationship_text, %s, provenance_refs, and %d source citations. Candidate comparison: current_primitives_plus_graph_context_report is the available reference, graph_relationship_report is the promoted read-only surface, and split_specialized_reports is not selected. Decision: promote graph_relationship_report; no follow-up work is required for the deferred relationship/path, direct-vs-derived, typed-candidate, or limited graph-audit needs.",
			sourcePath,
			len(report.RelationshipPaths),
			len(report.DirectRelationships),
			len(report.DerivedRelationships),
			len(report.TypedRelationshipCandidates),
			len(report.AuditFindings),
			projectionListFreshnessSummary(report.GraphProjection),
			citationCount,
		),
		Evidence:                    evidence,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine relationship paths, direct-vs-derived reporting, typed candidates, or limited graph audits; use get_document, document_links, graph_neighborhood, provenance_events, and projection_states directly only for explicit drill-down or runner rejection repair",
	}
}

func graphRelationshipCitationCount(report GraphRelationshipReport) int {
	count := 0
	for _, path := range report.RelationshipPaths {
		count += len(path.Citations)
	}
	for _, relationship := range report.DirectRelationships {
		count += len(relationship.Citations)
	}
	for _, relationship := range report.DerivedRelationships {
		count += len(relationship.Citations)
	}
	for _, candidate := range report.TypedRelationshipCandidates {
		if candidate.Citation.Path != "" || candidate.Citation.DocID != "" || candidate.Citation.ChunkID != "" {
			count++
		}
	}
	for _, finding := range report.AuditFindings {
		count += len(finding.Citations)
	}
	return count
}

func graphRelationshipCandidates() []GraphRelationshipCandidate {
	return []GraphRelationshipCandidate{
		{
			Surface:    "current_primitives_plus_graph_context_report",
			Status:     "available_reference",
			Safety:     "passes when agents use graph_context_report and explicit primitives through the installed runner",
			Capability: "can express paths, direct/derived evidence, typed candidates, freshness, provenance, and simple audits by stitching returned fields",
			UXQuality:  "ceremonial for normal users because adjacent graph questions still require manual interpretation and primitive drill-down",
			Implementation: []string{
				"graph_context_report",
				"document_links, graph_neighborhood, provenance_events, projection_states",
			},
		},
		{
			Surface:    "graph_relationship_report",
			Status:     "promoted_read_only",
			Safety:     "passes because it repackages current runner evidence without writes, semantic graph storage, hidden ranking, graph memory, or bypasses",
			Capability: "passes for relationship/path finding, direct-vs-derived reporting, typed candidates from canonical markdown, and limited stale/orphaned/contradiction audit findings",
			UXQuality:  "one retrieval action matches the deferred user-facing graph audit and relationship report needs",
			Implementation: []string{
				"runner JSON action under openclerk retrieval",
				"agent_handoff for final-answer evidence",
				"graph_context_report remains the broad context baseline",
			},
		},
		{
			Surface:    "split_specialized_reports",
			Status:     "not_selected",
			Safety:     "could pass with the same boundaries but increases public surface area",
			Capability: "would separate path, typed-candidate, and audit reports even though they share the same source evidence",
			UXQuality:  "worse than one combined read-only report because agents must choose among adjacent graph actions before seeing evidence",
			Implementation: []string{
				"defer separate relationship_path_report",
				"defer separate graph_audit_report",
				"defer separate typed_relationship_candidate_report",
			},
		},
	}
}
