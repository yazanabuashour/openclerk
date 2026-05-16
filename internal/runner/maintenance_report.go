package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	maintenanceValidationBoundaries = "read-only maintenance report; packages existing runner checks only and does not create cron jobs, background jobs, autonomous repair, document writes, source fetches, direct vault inspection, direct SQLite reads, raw diffs, HTTP/MCP calls, or unsupported transports"
	maintenanceAuthorityLimits      = "maintenance findings are operational evidence and handoff guidance only; canonical markdown, citations, provenance, projection freshness, and approved runner writes remain authority"
)

func runMaintenanceReport(ctx context.Context, client *runclient.Client, config runclient.Config, options MaintenanceReportOptions) (MaintenanceReport, error) {
	limit := cappedRunnerLimit(options.Limit, 20, 100)
	layout, err := inspectKnowledgeLayout(ctx, client)
	if err != nil {
		return MaintenanceReport{}, err
	}
	projections, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{Limit: limit})
	if err != nil {
		return MaintenanceReport{}, err
	}
	convertedProjections := toProjectionStateList(projections)
	modulePostures, err := maintenanceModulePostures(ctx, config)
	if err != nil {
		return MaintenanceReport{}, err
	}
	gitLifecycle, err := runGitLifecycleReport(ctx, client.Paths().VaultRoot, GitLifecycleOptions{Mode: gitLifecycleModeStatus, Limit: limit}, config)
	if err != nil {
		return MaintenanceReport{}, err
	}

	report := MaintenanceReport{
		Query:                options.Query,
		DuplicateQuery:       options.DuplicateQuery,
		Path:                 options.Path,
		DocID:                options.DocID,
		PathPrefix:           options.PathPrefix,
		Layout:               &layout,
		Projections:          &convertedProjections,
		ModulePostures:       modulePostures,
		GitLifecycle:         &gitLifecycle,
		WriteStatus:          "read_only_no_repair",
		ValidationBoundaries: maintenanceValidationBoundaries,
		AuthorityLimits:      maintenanceAuthorityLimits,
	}

	relationshipSelector := GraphRelationshipOptions{
		DocID:      options.DocID,
		Path:       options.Path,
		Query:      options.Query,
		PathPrefix: options.PathPrefix,
		Limit:      limit,
	}
	if relationshipSelector.DocID != "" || relationshipSelector.Path != "" || relationshipSelector.Query != "" {
		relationship, err := runGraphRelationshipReport(ctx, client, relationshipSelector)
		if err != nil {
			return MaintenanceReport{}, err
		}
		report.RelationshipContext = &relationship
	}

	duplicateQuery := firstNonEmpty(options.DuplicateQuery, options.Query)
	if duplicateQuery != "" {
		duplicate, err := runDuplicateCandidateReport(ctx, client, DuplicateCandidateOptions{
			Query:      duplicateQuery,
			PathPrefix: options.PathPrefix,
			Limit:      limit,
		})
		if err != nil {
			return MaintenanceReport{}, err
		}
		report.DuplicateCandidate = &duplicate
		report.DuplicateQuery = duplicateQuery
	}

	report.Findings = maintenanceFindings(report)
	report.Recommendation = maintenanceRecommendation(report.Findings)
	report.AgentHandoff = &AgentHandoff{
		AnswerSummary:               report.Recommendation,
		Evidence:                    maintenanceHandoffEvidence(report),
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "use inspect_layout, projection_states, graph_relationship_report, duplicate_candidate_report, config inspect_config, or git_lifecycle_report for drill-down; use approved document actions for any repair",
	}
	return report, nil
}

func maintenanceModulePostures(ctx context.Context, config runclient.Config) ([]MaintenanceModulePosture, error) {
	modules, err := runclient.ListConfiguredModules(ctx, config)
	if err != nil {
		return nil, err
	}
	postures := make([]MaintenanceModulePosture, 0, len(modules))
	for _, module := range modules {
		posture := "configured"
		if !module.Enabled {
			posture = "disabled"
		} else if module.VerificationStatus == "verified" {
			posture = "enabled_verified"
		} else if module.VerificationStatus != "" {
			posture = module.VerificationStatus
		}
		postures = append(postures, MaintenanceModulePosture{
			Kind:               module.Kind,
			Provider:           module.Provider,
			ModuleName:         module.ModuleName,
			Enabled:            module.Enabled,
			VerificationStatus: module.VerificationStatus,
			Posture:            posture,
		})
	}
	return postures, nil
}

func maintenanceFindings(report MaintenanceReport) []MaintenanceFinding {
	findings := []MaintenanceFinding{}
	if report.Layout != nil {
		findings = append(findings, maintenanceLayoutFinding(*report.Layout))
	}
	if report.Projections != nil {
		findings = append(findings, maintenanceProjectionFinding(*report.Projections))
	}
	if report.RelationshipContext != nil {
		findings = append(findings, maintenanceRelationshipFinding(*report.RelationshipContext))
	} else {
		findings = append(findings, MaintenanceFinding{
			Area:     "relationship_context",
			Status:   "not_evaluated",
			Summary:  "No doc_id, path, or query was supplied for relationship context inspection.",
			NextStep: `rerun maintenance_report with maintenance.doc_id, maintenance.path, or maintenance.query, or run graph_relationship_report directly`,
		})
	}
	if report.DuplicateCandidate != nil {
		findings = append(findings, maintenanceDuplicateFinding(*report.DuplicateCandidate))
	} else {
		findings = append(findings, MaintenanceFinding{
			Area:     "duplicate_risk",
			Status:   "not_evaluated",
			Summary:  "No query or duplicate_query was supplied for duplicate-risk inspection.",
			NextStep: `rerun maintenance_report with maintenance.query or maintenance.duplicate_query, or run duplicate_candidate_report directly`,
		})
	}
	findings = append(findings, maintenanceModuleFinding(report.ModulePostures))
	if report.GitLifecycle != nil {
		findings = append(findings, maintenanceGitFinding(*report.GitLifecycle))
	}
	return findings
}

func maintenanceLayoutFinding(layout KnowledgeLayout) MaintenanceFinding {
	failures, warnings := 0, 0
	evidence := []string{fmt.Sprintf("layout_valid=%t", layout.Valid)}
	for _, check := range layout.Checks {
		switch check.Status {
		case layoutCheckFail:
			failures++
		case layoutCheckWarn:
			warnings++
		}
	}
	evidence = append(evidence, fmt.Sprintf("layout_failures=%d", failures), fmt.Sprintf("layout_warnings=%d", warnings))
	if failures > 0 {
		return MaintenanceFinding{Area: "layout", Status: "attention", Summary: "Layout checks found failing convention or source-reference requirements.", Evidence: evidence, NextStep: "inspect layout checks and repair through approved document actions"}
	}
	if warnings > 0 {
		return MaintenanceFinding{Area: "layout", Status: "warn", Summary: "Layout checks found warnings but no failures.", Evidence: evidence, NextStep: "review optional-prefix and production-completeness warnings"}
	}
	return MaintenanceFinding{Area: "layout", Status: "clear", Summary: "Layout checks are valid.", Evidence: evidence}
}

func maintenanceProjectionFinding(projections ProjectionStateList) MaintenanceFinding {
	stale, unknown := 0, 0
	evidence := []string{fmt.Sprintf("projection_count=%d", len(projections.Projections))}
	for _, projection := range projections.Projections {
		switch projection.Freshness {
		case "fresh":
		case "stale":
			stale++
		default:
			unknown++
		}
	}
	evidence = append(evidence, fmt.Sprintf("stale=%d", stale), fmt.Sprintf("unknown=%d", unknown))
	if stale > 0 || unknown > 0 {
		return MaintenanceFinding{Area: "projection_freshness", Status: "attention", Summary: "Projection freshness includes stale or unknown state.", Evidence: evidence, NextStep: "inspect projection_states and use approved repair actions where appropriate"}
	}
	return MaintenanceFinding{Area: "projection_freshness", Status: "clear", Summary: "Projection freshness has no stale state in the returned window.", Evidence: evidence}
}

func maintenanceRelationshipFinding(report GraphRelationshipReport) MaintenanceFinding {
	evidence := []string{fmt.Sprintf("audit_findings=%d", len(report.AuditFindings))}
	for _, finding := range report.AuditFindings {
		evidence = append(evidence, finding.Kind+"="+finding.Status)
		if finding.Kind == "orphaned_graph_context" && finding.Status == "attention" {
			return MaintenanceFinding{
				Area:     "relationship_context",
				Status:   "attention",
				Summary:  "Relationship context appears orphaned in the inspected graph evidence.",
				Evidence: evidence,
				NextStep: "use graph_relationship_maintenance_plan for approval-gated canonical markdown relationship candidates",
			}
		}
		if finding.Status == "attention" {
			return MaintenanceFinding{
				Area:     "relationship_context",
				Status:   "attention",
				Summary:  "Relationship audit reported attention for " + finding.Kind + ".",
				Evidence: evidence,
				NextStep: "inspect graph_relationship_report evidence and use approved maintenance actions where appropriate",
			}
		}
	}
	return MaintenanceFinding{Area: "relationship_context", Status: "clear", Summary: "Inspected relationship context has no orphan finding.", Evidence: evidence}
}

func maintenanceDuplicateFinding(report DuplicateCandidateReport) MaintenanceFinding {
	evidence := []string{"duplicate_status=" + report.DuplicateStatus}
	if report.LikelyTarget != nil {
		evidence = append(evidence, "likely_target="+report.LikelyTarget.Path)
		return MaintenanceFinding{
			Area:     "duplicate_risk",
			Status:   "attention",
			Summary:  "Duplicate candidate report found a likely existing target.",
			Evidence: evidence,
			NextStep: "ask whether to update the existing target or create a confirmed new path before any durable write",
		}
	}
	return MaintenanceFinding{Area: "duplicate_risk", Status: "clear", Summary: "Duplicate candidate report did not find a runner-visible duplicate.", Evidence: evidence}
}

func maintenanceModuleFinding(postures []MaintenanceModulePosture) MaintenanceFinding {
	if len(postures) == 0 {
		return MaintenanceFinding{
			Area:     "module_config",
			Status:   "info",
			Summary:  "No optional modules are configured; core lexical search still works.",
			Evidence: []string{"configured_modules=0"},
			NextStep: "install optional semantic or OCR modules only when the workflow explicitly needs them",
		}
	}
	attention := 0
	evidence := []string{fmt.Sprintf("configured_modules=%d", len(postures))}
	for _, posture := range postures {
		evidence = append(evidence, posture.Provider+"="+posture.Posture)
		if posture.Enabled && posture.VerificationStatus != "verified" {
			attention++
		}
	}
	if attention > 0 {
		return MaintenanceFinding{Area: "module_config", Status: "attention", Summary: "One or more enabled modules are not verified.", Evidence: evidence, NextStep: "inspect config/module state and reinstall or disable unverified modules"}
	}
	return MaintenanceFinding{Area: "module_config", Status: "clear", Summary: "Configured modules have no enabled unverified state.", Evidence: evidence}
}

func maintenanceGitFinding(report GitLifecycleReport) MaintenanceFinding {
	evidence := []string{"git_status=" + report.GitStatus, fmt.Sprintf("dirty_paths=%d", len(report.DirtyPaths))}
	switch report.GitStatus {
	case "dirty":
		return MaintenanceFinding{Area: "git_lifecycle", Status: "attention", Summary: "Vault Git lifecycle has dirty paths.", Evidence: evidence, NextStep: "review local storage status before durable maintenance writes"}
	case "unavailable":
		return MaintenanceFinding{Area: "git_lifecycle", Status: "warn", Summary: "Vault Git lifecycle metadata is unavailable.", Evidence: evidence, NextStep: "use git_lifecycle_report only when the vault root is a Git worktree"}
	default:
		return MaintenanceFinding{Area: "git_lifecycle", Status: "clear", Summary: "Vault Git lifecycle status is clean or available.", Evidence: evidence}
	}
}

func maintenanceRecommendation(findings []MaintenanceFinding) string {
	attention := []string{}
	for _, finding := range findings {
		if finding.Status == "attention" {
			attention = append(attention, finding.Area)
		}
	}
	if len(attention) == 0 {
		return "maintenance_report found no attention findings in evaluated checks; use listed not_evaluated follow-ups only when that context matters"
	}
	return "maintenance_report found attention findings in " + strings.Join(attention, ", ") + "; use the named runner follow-ups and approved document actions for repairs"
}

func maintenanceHandoffEvidence(report MaintenanceReport) []string {
	evidence := []string{"write_status=" + report.WriteStatus}
	for _, finding := range report.Findings {
		evidence = append(evidence, finding.Area+"="+finding.Status)
	}
	return evidence
}
