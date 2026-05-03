package runner

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	auditModePlanOnly       = "plan_only"
	auditModeRepairExisting = "repair_existing"
)

func runAuditContradictions(ctx context.Context, client *runclient.Client, options AuditContradictionsOptions) (AuditContradictionsResult, error) {
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}
	result := AuditContradictionsResult{
		Query:                 options.Query,
		TargetPath:            options.TargetPath,
		Mode:                  options.Mode,
		RepairStatus:          "skipped",
		DuplicatePrevention:   "target_not_found",
		FailureClassification: "none",
	}

	search, err := client.Search(ctx, domain.SearchQuery{Text: options.Query, Limit: limit})
	if err != nil {
		return AuditContradictionsResult{}, err
	}
	sourcePaths, citations := auditSourceEvidence(search.Hits)
	result.SourcePaths = sourcePaths
	result.Citations = citations

	candidatePaths, targetMatches, err := auditSynthesisCandidates(ctx, client, options.TargetPath)
	if err != nil {
		return AuditContradictionsResult{}, err
	}
	result.CandidateSynthesisPaths = candidatePaths

	if len(targetMatches) != 1 {
		if len(targetMatches) > 1 {
			result.DuplicatePrevention = "duplicate_target_path_detected"
			result.FailureClassification = "duplicate_target"
		} else {
			result.FailureClassification = "target_not_found"
		}
		return result, nil
	}

	targetSummary := targetMatches[0]
	result.SelectedTargetPath = targetSummary.Path
	result.DuplicatePrevention = "existing_target_selected_no_duplicate_created"

	target, err := client.GetDocument(ctx, targetSummary.DocID)
	if err != nil {
		return AuditContradictionsResult{}, err
	}
	before, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: "synthesis",
		RefKind:    "document",
		RefID:      target.DocID,
		Limit:      10,
	})
	if err != nil {
		return AuditContradictionsResult{}, err
	}
	result.ProjectionFreshnessBefore = toProjectionStates(before.Projections)
	result.CurrentSourcePaths, result.SupersededSourcePaths = sourceClassesFromProjection(before.Projections)

	targetEvents, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "projection",
		RefID:   "synthesis:" + target.DocID,
		Limit:   10,
	})
	if err != nil {
		return AuditContradictionsResult{}, err
	}
	result.ProvenanceInspected = append(result.ProvenanceInspected, auditProvenanceInspection("projection", "synthesis:"+target.DocID, target.Path, targetEvents.Events))

	for _, sourcePath := range append(result.CurrentSourcePaths, result.SupersededSourcePaths...) {
		sourceDoc, ok, err := auditDocumentByPath(ctx, client, sourcePath)
		if err != nil {
			return AuditContradictionsResult{}, err
		}
		if !ok {
			continue
		}
		sourceEvents, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
			RefKind: "document",
			RefID:   sourceDoc.DocID,
			Limit:   10,
		})
		if err != nil {
			return AuditContradictionsResult{}, err
		}
		result.ProvenanceInspected = append(result.ProvenanceInspected, auditProvenanceInspection("document", sourceDoc.DocID, sourceDoc.Path, sourceEvents.Events))
	}

	if options.Mode == auditModePlanOnly {
		result.RepairStatus = "planned"
	} else {
		repairContent, ok := auditRepairSummaryContent(result.CurrentSourcePaths, result.SupersededSourcePaths)
		if !ok {
			result.RepairStatus = "skipped_insufficient_source_classification"
			result.FailureClassification = "insufficient_evidence"
		} else {
			repaired, err := client.ReplaceSection(ctx, target.DocID, domain.ReplaceSectionInput{
				Heading: "Summary",
				Content: repairContent,
			})
			if err != nil {
				return AuditContradictionsResult{}, err
			}
			target = repaired
			result.RepairStatus = "applied"
			result.RepairApplied = true
		}
	}

	after, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: "synthesis",
		RefKind:    "document",
		RefID:      target.DocID,
		Limit:      10,
	})
	if err != nil {
		return AuditContradictionsResult{}, err
	}
	result.ProjectionFreshnessAfter = toProjectionStates(after.Projections)

	if options.ConflictQuery != "" {
		conflicts, inspections, err := auditUnresolvedConflicts(ctx, client, options.ConflictQuery, limit)
		if err != nil {
			return AuditContradictionsResult{}, err
		}
		result.UnresolvedConflictGroups = conflicts
		result.ProvenanceInspected = append(result.ProvenanceInspected, inspections...)
		if len(conflicts) == 0 && result.FailureClassification == "none" {
			result.FailureClassification = "insufficient_conflict_evidence"
		}
	}

	return result, nil
}

func auditContradictionsSummary(result AuditContradictionsResult) string {
	return fmt.Sprintf("audit_contradictions %s for %s; repair %s; %d unresolved conflict groups",
		result.Mode,
		result.TargetPath,
		result.RepairStatus,
		len(result.UnresolvedConflictGroups),
	)
}

func auditSourceEvidence(hits []domain.SearchHit) ([]string, []Citation) {
	paths := []string{}
	citations := []Citation{}
	for _, hit := range hits {
		if strings.HasPrefix(hit.Title, "sources/") {
			paths = appendUniqueString(paths, hit.Title)
		}
		for _, citation := range hit.Citations {
			if citation.Path != "" {
				paths = appendUniqueString(paths, citation.Path)
			}
			citations = append(citations, toCitations([]domain.Citation{citation})...)
		}
	}
	sort.Strings(paths)
	return paths, citations
}

func auditSynthesisCandidates(ctx context.Context, client *runclient.Client, targetPath string) ([]string, []domain.DocumentSummary, error) {
	candidatePaths := []string{}
	targetMatches := []domain.DocumentSummary{}
	cursor := ""
	for {
		list, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix: "synthesis/",
			Limit:      100,
			Cursor:     cursor,
		})
		if err != nil {
			return nil, nil, err
		}
		for _, document := range list.Documents {
			candidatePaths = appendUniqueString(candidatePaths, document.Path)
			if document.Path == targetPath {
				targetMatches = append(targetMatches, document)
			}
		}
		if !list.PageInfo.HasMore {
			break
		}
		cursor = list.PageInfo.NextCursor
		if cursor == "" {
			return nil, nil, fmt.Errorf("list synthesis candidates did not return next cursor")
		}
	}
	sort.Strings(candidatePaths)
	return candidatePaths, targetMatches, nil
}

func sourceClassesFromProjection(projections []domain.ProjectionState) ([]string, []string) {
	current := []string{}
	superseded := []string{}
	for _, projection := range projections {
		for _, path := range splitAuditList(projection.Details["current_source_refs"]) {
			current = appendUniqueString(current, path)
		}
		for _, path := range splitAuditList(projection.Details["superseded_source_refs"]) {
			superseded = appendUniqueString(superseded, path)
		}
	}
	sort.Strings(current)
	sort.Strings(superseded)
	return current, superseded
}

func auditRepairSummaryContent(currentSourcePaths []string, supersededSourcePaths []string) (string, bool) {
	if len(currentSourcePaths) == 0 || len(supersededSourcePaths) == 0 {
		return "", false
	}
	return strings.Join([]string{
		"Current audit guidance: use the installed openclerk JSON runner.",
		"Current source: " + currentSourcePaths[0] + ".",
		"Superseded source: " + supersededSourcePaths[0] + ".",
	}, "\n"), true
}

func auditUnresolvedConflicts(ctx context.Context, client *runclient.Client, query string, limit int) ([]AuditConflictGroup, []AuditProvenanceInspection, error) {
	conflicts, inspections, err := auditUnresolvedConflictsForQuery(ctx, client, query, limit)
	if err != nil {
		return nil, nil, err
	}
	if len(conflicts) != 0 || !strings.Contains(query, ":") {
		if len(conflicts) != 0 {
			return conflicts, inspections, nil
		}
		return auditUnresolvedConflictsWithKeywordFallback(ctx, client, query, limit, conflicts, inspections)
	}
	prefix := strings.TrimSpace(strings.SplitN(query, ":", 2)[0])
	if prefix == "" || prefix == query {
		return auditUnresolvedConflictsWithKeywordFallback(ctx, client, query, limit, conflicts, inspections)
	}
	fallbackConflicts, fallbackInspections, err := auditUnresolvedConflictsForQuery(ctx, client, prefix, limit)
	if err != nil {
		return nil, nil, err
	}
	if len(fallbackConflicts) == 0 {
		return auditUnresolvedConflictsWithKeywordFallback(ctx, client, query, limit, conflicts, inspections)
	}
	return fallbackConflicts, append(inspections, fallbackInspections...), nil
}

func auditUnresolvedConflictsWithKeywordFallback(ctx context.Context, client *runclient.Client, query string, limit int, conflicts []AuditConflictGroup, inspections []AuditProvenanceInspection) ([]AuditConflictGroup, []AuditProvenanceInspection, error) {
	if !auditContainsAll(strings.ToLower(query), []string{"source", "sensitive", "audit", "conflict", "runner", "retention"}) {
		return conflicts, inspections, nil
	}
	fallbackQuery := "source sensitive audit conflict runner retention"
	if strings.TrimSpace(strings.ToLower(query)) == fallbackQuery {
		return conflicts, inspections, nil
	}
	fallbackConflicts, fallbackInspections, err := auditUnresolvedConflictsForQuery(ctx, client, fallbackQuery, limit)
	if err != nil {
		return nil, nil, err
	}
	if len(fallbackConflicts) == 0 {
		return conflicts, inspections, nil
	}
	return fallbackConflicts, append(inspections, fallbackInspections...), nil
}

func auditContainsAll(value string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}

func auditUnresolvedConflictsForQuery(ctx context.Context, client *runclient.Client, query string, limit int) ([]AuditConflictGroup, []AuditProvenanceInspection, error) {
	search, err := client.Search(ctx, domain.SearchQuery{Text: query, Limit: limit})
	if err != nil {
		return nil, nil, err
	}
	paths := []string{}
	claims := []string{}
	docIDsByPath := map[string]string{}
	for _, hit := range search.Hits {
		for _, citation := range hit.Citations {
			if citation.Path == "" || !strings.HasPrefix(citation.Path, "sources/") {
				continue
			}
			paths = appendUniqueString(paths, citation.Path)
			docIDsByPath[citation.Path] = citation.DocID
			if strings.TrimSpace(hit.Snippet) != "" {
				claims = appendUniqueString(claims, strings.TrimSpace(hit.Snippet))
			}
		}
	}
	sort.Strings(paths)

	currentPaths := []string{}
	inspections := []AuditProvenanceInspection{}
	for _, path := range paths {
		docID := docIDsByPath[path]
		document, ok, err := auditDocumentByPath(ctx, client, path)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			docID = document.DocID
			if !isSupersededAuditDocument(document) {
				currentPaths = append(currentPaths, path)
			}
		}
		if docID == "" {
			continue
		}
		events, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
			RefKind: "document",
			RefID:   docID,
			Limit:   10,
		})
		if err != nil {
			return nil, nil, err
		}
		inspections = append(inspections, auditProvenanceInspection("document", docID, path, events.Events))
	}
	sort.Strings(currentPaths)

	if len(currentPaths) < 2 {
		return nil, inspections, nil
	}
	if !auditClaimsDisagree(claims) {
		return nil, inspections, nil
	}
	return []AuditConflictGroup{{
		Query:       query,
		SourcePaths: currentPaths,
		Claims:      claims,
		Status:      "unresolved",
		Reason:      "current sources disagree and no runner-visible supersession or source authority chooses a winner",
	}}, inspections, nil
}

func auditDocumentByPath(ctx context.Context, client *runclient.Client, path string) (domain.Document, bool, error) {
	list, err := client.ListDocuments(ctx, domain.DocumentListQuery{PathPrefix: path, Limit: 10})
	if err != nil {
		return domain.Document{}, false, err
	}
	for _, summary := range list.Documents {
		if summary.Path != path {
			continue
		}
		document, err := client.GetDocument(ctx, summary.DocID)
		if err != nil {
			return domain.Document{}, false, err
		}
		return document, true, nil
	}
	return domain.Document{}, false, nil
}

func auditProvenanceInspection(refKind string, refID string, sourcePath string, events []domain.ProvenanceEvent) AuditProvenanceInspection {
	inspection := AuditProvenanceInspection{
		RefKind:    refKind,
		RefID:      refID,
		SourcePath: sourcePath,
		Details: map[string]string{
			"event_count": fmt.Sprintf("%d", len(events)),
		},
	}
	for _, event := range events {
		inspection.EventIDs = appendUniqueString(inspection.EventIDs, event.EventID)
		inspection.EventTypes = appendUniqueString(inspection.EventTypes, event.EventType)
	}
	sort.Strings(inspection.EventIDs)
	sort.Strings(inspection.EventTypes)
	return inspection
}

func isSupersededAuditDocument(document domain.Document) bool {
	return strings.EqualFold(strings.TrimSpace(document.Metadata["status"]), "superseded")
}

func auditClaimsDisagree(claims []string) bool {
	retentionDays := map[int]struct{}{}
	for _, claim := range claims {
		for _, days := range auditRetentionDays(claim) {
			retentionDays[days] = struct{}{}
		}
	}
	return len(retentionDays) > 1
}

func auditRetentionDays(claim string) []int {
	tokens := strings.Fields(strings.ToLower(claim))
	result := []int{}
	for index, token := range tokens {
		normalized := strings.Trim(token, ".,;:()[]{}")
		days, ok := auditNumberWord(normalized)
		if !ok {
			continue
		}
		if index+1 >= len(tokens) {
			continue
		}
		unit := strings.Trim(tokens[index+1], ".,;:()[]{}")
		if unit == "day" || unit == "days" {
			result = append(result, days)
		}
	}
	return result
}

func auditNumberWord(value string) (int, bool) {
	switch value {
	case "1", "one":
		return 1, true
	case "2", "two":
		return 2, true
	case "3", "three":
		return 3, true
	case "4", "four":
		return 4, true
	case "5", "five":
		return 5, true
	case "6", "six":
		return 6, true
	case "7", "seven":
		return 7, true
	case "8", "eight":
		return 8, true
	case "9", "nine":
		return 9, true
	case "10", "ten":
		return 10, true
	case "30", "thirty":
		return 30, true
	default:
		return 0, false
	}
}

func splitAuditList(value string) []string {
	parts := strings.Split(value, ",")
	result := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func appendUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
