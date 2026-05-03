package runner

import (
	"context"
	"fmt"

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
	return SourceAuditReport{
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
		ValidationBoundaries:      sourceAuditValidationBoundaries(),
		AuthorityLimits:           sourceAuditAuthorityLimits(),
	}, nil
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
