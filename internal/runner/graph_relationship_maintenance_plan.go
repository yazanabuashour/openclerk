package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	graphRelationshipMaintenanceValidationBoundaries = "read-only maintenance plan; uses graph_relationship_report evidence through installed OpenClerk retrieval JSON only; no durable document write, no direct vault inspection, no direct SQLite, no source-built runners, no HTTP/MCP bypasses, no unsupported transports, no semantic-label graph truth, no hidden authority ranking, no graph memory, no durable semantic graph storage, and no automatic repair"
	graphRelationshipMaintenanceAuthorityLimits      = "canonical markdown remains relationship authority; candidate maintenance text and typed relationship annotations are review-required suggestions from source citations, cited markdown wording, links, backlinks, graph projection freshness, and provenance, not durable facts until an explicit approved document write"
	graphRelationshipMaintenanceApprovalBoundary     = "read/fetch/inspect planning is not durable-write approval; approve the exact replace_section or append_document request before mutating canonical markdown"
)

func runGraphRelationshipMaintenancePlan(ctx context.Context, client *runclient.Client, options GraphRelationshipMaintenanceOptions) (GraphRelationshipMaintenancePlan, error) {
	report, err := runGraphRelationshipReport(ctx, client, GraphRelationshipOptions(options))
	if err != nil {
		return GraphRelationshipMaintenancePlan{}, err
	}

	actions := graphRelationshipMaintenanceActions(report)
	content := graphRelationshipMaintenanceSectionContent(report)
	nextReplace := graphRelationshipNextReplaceSectionRequest(report, "Relationships", content)
	nextAppend := graphRelationshipNextAppendDocumentRequest(report, content)
	evidenceInspected := append([]string{}, report.EvidenceInspected...)
	evidenceInspected = append(evidenceInspected,
		fmt.Sprintf("maintenance_actions:%d", len(actions)),
		"next_replace_section_request:planned_no_write",
		"next_append_document_request:planned_no_write",
	)

	plan := GraphRelationshipMaintenancePlan{
		Query:                     options.Query,
		Path:                      options.Path,
		DocID:                     options.DocID,
		PathPrefix:                options.PathPrefix,
		SourceDocument:            report.SourceDocument,
		SourceSelection:           report.SourceSelection,
		ProposedActions:           actions,
		CandidateSectionHeading:   "Relationships",
		CandidateSectionContent:   content,
		NextReplaceSectionRequest: nextReplace,
		NextAppendDocumentRequest: nextAppend,
		WriteStatus:               "planned_no_write",
		ApprovalBoundary:          graphRelationshipMaintenanceApprovalBoundary,
		RollbackAuditPath:         "after an approved write, use git_lifecycle_report for local storage checkpoint/history, provenance_events for document/projection provenance, and projection_states for graph freshness; rollback remains an explicit document/version-control workflow, not automatic",
		DuplicateHandling:         "single source document selected by doc_id, exact path, or visible query candidate; maintenance edits target that existing document and do not create duplicate relationship pages",
		FailureModes: []string{
			"source selector is missing, ambiguous, or resolves to the wrong visible query candidate",
			"target document lacks a Relationships heading, so append_document may be safer than replace_section",
			"canonical relationship text is absent or too weak for typed annotation without human review",
			"graph projection freshness is stale or unknown after an approved write and needs projection repair before relying on graph evidence",
			"duplicate or conflicting relationship notes require explicit source-sensitive audit rather than automatic graph repair",
		},
		GraphProjection:      report.GraphProjection,
		ProvenanceRefs:       report.ProvenanceRefs,
		CandidateSurfaces:    graphRelationshipMaintenanceCandidates(),
		Recommendation:       "promote graph_relationship_maintenance_plan as the approval-before-write plan surface for canonical markdown relationship annotation and maintenance; keep graph_relationship_report as the read-only report and document writes as explicit approval steps",
		SafetyPass:           "passes: plan is read-only, local-first, runner-only, citation-bearing, and returns explicit next write requests without executing them",
		CapabilityPass:       "passes: combines relationship report evidence with candidate section content, approval boundary, duplicate handling, rollback/audit path, freshness/provenance posture, and failure modes",
		UXQuality:            "improves maintenance tasks by replacing a ceremonial report-plus-manual-write-plan sequence with one plan response while preserving explicit durable-write approval",
		EvidencePosture:      "evidence comes from graph_relationship_report direct markdown evidence, relationship paths, typed candidates, limited audit findings, graph projection freshness, and provenance refs",
		ValidationBoundaries: graphRelationshipMaintenanceValidationBoundaries,
		AuthorityLimits:      graphRelationshipMaintenanceAuthorityLimits,
		EvidenceInspected:    evidenceInspected,
	}
	plan.AgentHandoff = graphRelationshipMaintenanceHandoff(plan)
	return plan, nil
}

func graphRelationshipMaintenanceActions(report GraphRelationshipReport) []GraphRelationshipMaintenanceAction {
	actions := []GraphRelationshipMaintenanceAction{}
	for _, candidate := range report.TypedRelationshipCandidates {
		if candidate.RelationshipType == "markdown_link" {
			continue
		}
		actions = append(actions, GraphRelationshipMaintenanceAction{
			Kind:      "typed_relationship_annotation",
			Status:    "candidate_requires_approval",
			Evidence:  fmt.Sprintf("%s from cited canonical relationship text", candidate.RelationshipType),
			NextStep:  "review the candidate label, then approve a replace_section or append_document write if it should become durable canonical markdown",
			Citations: []Citation{candidate.Citation},
		})
	}
	for _, finding := range report.AuditFindings {
		if finding.Status != "attention" && finding.Status != "unknown" {
			continue
		}
		actions = append(actions, GraphRelationshipMaintenanceAction{
			Kind:      "relationship_audit_followup",
			Status:    "needs_review",
			Evidence:  finding.Kind + ": " + finding.Evidence,
			NextStep:  "inspect cited canonical markdown and graph freshness before approving any relationship maintenance write",
			Citations: finding.Citations,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, GraphRelationshipMaintenanceAction{
			Kind:     "relationship_maintenance_status",
			Status:   "no_write_needed_from_returned_evidence",
			Evidence: "returned relationship evidence did not require an automatic maintenance write",
			NextStep: "use the candidate section only if a human wants durable annotation cleanup",
		})
	}
	return actions
}

func graphRelationshipMaintenanceSectionContent(report GraphRelationshipReport) string {
	lines := []string{
		"<!-- graph_relationship_maintenance_plan: review-required candidate; canonical markdown remains authority -->",
	}
	if len(report.DirectRelationships) > 0 {
		lines = append(lines, "", "Canonical relationship evidence:")
		seen := map[string]struct{}{}
		for _, relationship := range report.DirectRelationships {
			if relationship.EvidenceText == "" {
				continue
			}
			key := relationship.RelationshipType + "|" + relationship.EvidenceText
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			label := relationship.RelationshipType
			if label == "" {
				label = relationship.Kind
			}
			lines = append(lines, "- "+label+": "+relationship.EvidenceText)
		}
	}
	if len(report.RelationshipPaths) > 0 {
		lines = append(lines, "", "Relationship paths to preserve or review:")
		for _, relationshipPath := range report.RelationshipPaths {
			target := relationshipPath.Path
			if target == "" {
				target = relationshipPath.DocID
			}
			lines = append(lines, fmt.Sprintf("- %s: %s", relationshipPath.Direction, target))
		}
	}
	if len(report.AuditFindings) > 0 {
		lines = append(lines, "", "Maintenance audit findings:")
		for _, finding := range report.AuditFindings {
			lines = append(lines, fmt.Sprintf("- %s: %s - %s", finding.Kind, finding.Status, finding.Evidence))
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func graphRelationshipNextReplaceSectionRequest(report GraphRelationshipReport, heading string, content string) string {
	if report.SourceDocument == nil {
		return ""
	}
	return marshalRunnerRequest(map[string]any{
		"action":              DocumentTaskActionReplaceSection,
		"doc_id":              report.SourceDocument.DocID,
		"heading":             heading,
		"content":             content,
		"include_subsections": true,
		"include_heading":     false,
	})
}

func graphRelationshipNextAppendDocumentRequest(report GraphRelationshipReport, content string) string {
	if report.SourceDocument == nil {
		return ""
	}
	return marshalRunnerRequest(map[string]any{
		"action":  DocumentTaskActionAppend,
		"doc_id":  report.SourceDocument.DocID,
		"content": "## Relationships\n" + strings.TrimSpace(content),
	})
}

func marshalRunnerRequest(value map[string]any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func graphRelationshipMaintenanceHandoff(plan GraphRelationshipMaintenancePlan) *AgentHandoff {
	sourcePath := ""
	sourceDocID := ""
	if plan.SourceDocument != nil {
		sourcePath = plan.SourceDocument.Path
		sourceDocID = plan.SourceDocument.DocID
	}
	citationCount := 0
	for _, action := range plan.ProposedActions {
		citationCount += len(action.Citations)
	}
	evidence := []string{
		"source_path=" + sourcePath,
		"doc_id=" + sourceDocID,
		fmt.Sprintf("proposed_actions=%d", len(plan.ProposedActions)),
		fmt.Sprintf("source_citations=%d", citationCount),
		"write_status=" + plan.WriteStatus,
		"approval_required=true",
		"graph_projection_freshness=" + projectionListFreshnessSummary(plan.GraphProjection),
		"no_durable_semantic_graph_storage=true",
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"graph_relationship_maintenance_plan returned source %s with %d proposed approval-gated maintenance actions, source citations, candidate section content, next replace/append requests, %s, provenance refs, duplicate handling, rollback/audit path, failure modes, planned_no_write status, and no durable semantic graph storage",
			sourcePath,
			len(plan.ProposedActions),
			projectionListFreshnessSummary(plan.GraphProjection),
		),
		Evidence:                    evidence,
		ValidationBoundaries:        plan.ValidationBoundaries,
		AuthorityLimits:             plan.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine relationship maintenance planning; approve and run replace_section or append_document only after reviewing exact request content. The replace_section request preserves the matched heading unless include_heading is changed. Inspect provenance_events and projection_states if freshness evidence is needed; applied document writes return rollback_request.",
	}
}

func graphRelationshipMaintenanceCandidates() []GraphRelationshipMaintenanceCandidate {
	return []GraphRelationshipMaintenanceCandidate{
		{
			Surface:    "current_primitives_plus_graph_relationship_report",
			Status:     "available_reference",
			Safety:     "passes when agents run read-only graph_relationship_report and separately draft an approved replace_section or append_document request",
			Capability: "can preserve approval and provenance, but duplicate handling, rollback/audit path, and failure modes must be assembled manually",
			UXQuality:  "ceremonial for normal users because the report-to-write-plan bridge is surprising and prompt-sensitive",
			Implementation: []string{
				"graph_relationship_report",
				"manual replace_section or append_document request after approval",
			},
		},
		{
			Surface:    "graph_relationship_maintenance_plan",
			Status:     "promoted_plan_only",
			Safety:     "passes because it plans exact document write requests without executing them and keeps canonical markdown as authority",
			Capability: "passes for candidate relationship annotations, maintenance actions, duplicate handling, approval boundary, rollback/audit path, provenance/freshness posture, and failure modes",
			UXQuality:  "best fit: one read-only action gives a normal user the maintenance plan they expect while leaving durable writes explicit",
			Implementation: []string{
				"runner JSON action under openclerk retrieval",
				"next_replace_section_request and next_append_document_request for approved document writes",
				"graph_relationship_report remains the read-only evidence source",
			},
		},
		{
			Surface:    "durable_semantic_graph_maintenance",
			Status:     "not_selected",
			Safety:     "fails current evidence because it would add durable graph authority, rollback semantics, schema/migration burden, and duplicate conflict behavior not proven by this track",
			Capability: "could eventually support richer graph maintenance, but current canonical markdown workflows do not need new graph storage",
			UXQuality:  "too large for the observed need and less inspectable than approval-gated markdown edits",
			Implementation: []string{
				"no schema migration",
				"no semantic-label graph storage",
				"no graph memory or automatic repair",
			},
		},
	}
}
