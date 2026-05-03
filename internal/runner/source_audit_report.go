package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func runSourceAuditReport(ctx context.Context, client *runclient.Client, options SourceAuditReportOptions) (SourceAuditReport, error) {
	auditMode := auditModePlanOnly
	if options.Mode == "repair_existing" {
		auditMode = auditModeRepairExisting
	}
	audit, err := runAuditContradictions(ctx, client, AuditContradictionsOptions{
		Query:         options.Query,
		TargetPath:    options.TargetPath,
		Mode:          auditMode,
		ConflictQuery: options.ConflictQuery,
		Limit:         options.Limit,
	})
	if err != nil {
		return SourceAuditReport{}, err
	}
	validationBoundaries := sourceAuditValidationBoundaries()
	authorityLimits := sourceAuditAuthorityLimits()
	report := SourceAuditReport{
		Query:                     audit.Query,
		TargetPath:                audit.TargetPath,
		Mode:                      options.Mode,
		SelectedTargetPath:        audit.SelectedTargetPath,
		CandidateSynthesisPaths:   audit.CandidateSynthesisPaths,
		SourcePaths:               audit.SourcePaths,
		Citations:                 audit.Citations,
		CurrentSourcePaths:        audit.CurrentSourcePaths,
		SupersededSourcePaths:     audit.SupersededSourcePaths,
		ProvenanceInspected:       audit.ProvenanceInspected,
		ProjectionFreshnessBefore: audit.ProjectionFreshnessBefore,
		ProjectionFreshnessAfter:  audit.ProjectionFreshnessAfter,
		RepairStatus:              audit.RepairStatus,
		RepairApplied:             audit.RepairApplied,
		DuplicatePrevention:       audit.DuplicatePrevention,
		UnresolvedConflictGroups:  audit.UnresolvedConflictGroups,
		FailureClassification:     audit.FailureClassification,
		ValidationBoundaries:      validationBoundaries,
		AuthorityLimits:           authorityLimits,
	}
	report.AgentHandoff = sourceAuditHandoff(report)
	return report, nil
}

func sourceAuditReportSummary(report SourceAuditReport) string {
	return fmt.Sprintf("source_audit_report %s for %s; repair %s; %d unresolved conflict groups",
		report.Mode,
		report.TargetPath,
		report.RepairStatus,
		len(report.UnresolvedConflictGroups),
	)
}

func sourceAuditValidationBoundaries() string {
	return "runner-owned source-sensitive audit workflow; explain mode is read-only; repair_existing may update only an existing synthesis target; no broad repo search, direct vault inspection, direct file edits, direct SQLite, source-built runners, HTTP/MCP bypasses, unsupported transports, broad contradiction engine, or duplicate synthesis creation"
}

func sourceAuditAuthorityLimits() string {
	return "canonical source documents and runner-visible supersession/freshness evidence remain authority; unresolved current-source conflicts are explained with source paths and are not forced to a winner"
}

func sourceAuditHandoff(report SourceAuditReport) *AgentHandoff {
	projectionSummary := projectionFreshnessSummary(report.ProjectionFreshnessAfter)
	if len(report.ProjectionFreshnessAfter) == 0 {
		projectionSummary = projectionFreshnessSummary(report.ProjectionFreshnessBefore)
	}
	conflictStatus := fmt.Sprintf("%d unresolved conflict groups", len(report.UnresolvedConflictGroups))
	conflictSourcePaths, conflictClaims := sourceAuditConflictEvidence(report.UnresolvedConflictGroups)
	evidence := []string{
		"selected_target_path=" + report.SelectedTargetPath,
		"source_paths=" + strings.Join(report.SourcePaths, ", "),
		"current_source_paths=" + strings.Join(report.CurrentSourcePaths, ", "),
		"superseded_source_paths=" + strings.Join(report.SupersededSourcePaths, ", "),
		"conflict_source_paths=" + strings.Join(conflictSourcePaths, ", "),
		"conflict_claims=" + strings.Join(conflictClaims, " | "),
		"duplicate_prevention=" + report.DuplicatePrevention,
		"repair_status=" + report.RepairStatus,
		"projection_freshness=" + projectionSummary,
		"conflict_status=" + conflictStatus,
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"source_audit_report %s for %s; repair=%s; %s; %s",
			report.Mode,
			report.TargetPath,
			report.RepairStatus,
			conflictStatus,
			projectionSummary,
		),
		Evidence:                    evidence,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "not required for routine answer; use primitives only for explicit follow-up inspection, alternate target selection, or runner rejection repair",
	}
}

func sourceAuditConflictEvidence(groups []AuditConflictGroup) ([]string, []string) {
	sourcePaths := []string{}
	claims := []string{}
	for _, group := range groups {
		for _, sourcePath := range group.SourcePaths {
			sourcePaths = appendUniqueString(sourcePaths, sourcePath)
		}
		for _, claim := range group.Claims {
			claims = appendUniqueString(claims, claim)
		}
	}
	return sourcePaths, claims
}
