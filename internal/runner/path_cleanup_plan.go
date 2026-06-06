package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	pathCleanupModePlan  = "plan"
	pathCleanupModeApply = "apply"

	pathCleanupKindAuto               = "auto"
	pathCleanupKindRename             = "rename"
	pathCleanupKindCandidatePromotion = "candidate_promotion"
)

type pathCleanupSource struct {
	DocID    string
	Path     string
	Title    string
	Metadata map[string]string
}

func runPathCleanupPlan(ctx context.Context, client *runclient.Client, options PathCleanupOptions, autonomy AutonomyModes) (PathCleanupPlan, error) {
	sources, scope, err := pathCleanupSources(ctx, client, options)
	if err != nil {
		return PathCleanupPlan{}, err
	}
	candidates := make([]PathCleanupCandidate, 0, len(sources))
	for idx, source := range sources {
		candidate, err := pathCleanupCandidate(ctx, client, options, source, idx+1)
		if err != nil {
			return PathCleanupPlan{}, err
		}
		candidates = append(candidates, candidate)
	}
	markPathCleanupTargetCollisions(candidates)

	applied := 0
	if options.Mode == pathCleanupModeApply {
		for idx := range candidates {
			if candidates[idx].NextRequest == "" || candidates[idx].DuplicateRisk != "none" || candidates[idx].Confidence == "low" {
				candidates[idx].WriteStatus = "skipped_not_low_risk"
				continue
			}
			moved, err := client.MoveDocument(ctx, domain.MoveDocumentInput{
				Path:          candidates[idx].CurrentPath,
				TargetPath:    candidates[idx].ProposedTargetPath,
				UpdateLinks:   true,
				UpdateIndexes: options.UpdateIndexes,
			})
			if err != nil {
				return PathCleanupPlan{}, err
			}
			converted := toDocumentMoveResult(moved)
			candidates[idx].MoveResult = &converted
			candidates[idx].WriteStatus = "applied"
			applied++
		}
	}

	writeStatus := "planned_no_write"
	if options.Mode == pathCleanupModeApply {
		writeStatus = "applied"
		if applied == 0 {
			writeStatus = "no_low_risk_candidates_applied"
		}
	}
	plan := PathCleanupPlan{
		Mode:                 options.Mode,
		CleanupKind:          options.CleanupKind,
		Scope:                scope,
		Candidates:           candidates,
		AppliedCount:         applied,
		WriteStatus:          writeStatus,
		ApprovalBoundary:     pathCleanupApprovalBoundary(options, autonomy),
		ValidationBoundaries: "Path cleanup proposes target paths from runner-visible title, metadata, and current path only; durable apply uses move_document semantics and refuses existing targets.",
		AuthorityLimits:      "No semantic taxonomy inference, hidden vault reads, raw filesystem moves, overwrite, or delete is performed.",
	}
	plan.AgentHandoff = pathCleanupHandoff(plan)
	return plan, nil
}

func pathCleanupSources(ctx context.Context, client *runclient.Client, options PathCleanupOptions) ([]pathCleanupSource, string, error) {
	limit := options.Limit
	if limit <= 0 {
		limit = 10
	}
	switch {
	case options.DocID != "":
		document, err := client.GetDocument(ctx, options.DocID)
		if err != nil {
			return nil, "", err
		}
		return []pathCleanupSource{pathCleanupSourceFromDocument(document)}, "doc_id:" + options.DocID, nil
	case options.Path != "":
		documents, err := client.ListDocuments(ctx, domain.DocumentListQuery{PathPrefix: options.Path, Limit: limit})
		if err != nil {
			return nil, "", err
		}
		for _, document := range documents.Documents {
			if document.Path == options.Path {
				return []pathCleanupSource{pathCleanupSourceFromSummary(document)}, "path:" + options.Path, nil
			}
		}
		return nil, "", domain.NotFoundError("document path", options.Path)
	case options.PathPrefix != "":
		documents, err := client.ListDocuments(ctx, domain.DocumentListQuery{PathPrefix: options.PathPrefix, Limit: limit})
		if err != nil {
			return nil, "", err
		}
		sources := make([]pathCleanupSource, 0, len(documents.Documents))
		for _, document := range documents.Documents {
			sources = append(sources, pathCleanupSourceFromSummary(document))
		}
		return sources, "path_prefix:" + options.PathPrefix, nil
	default:
		search, err := client.Search(ctx, domain.SearchQuery{Text: options.Query, Limit: limit})
		if err != nil {
			return nil, "", err
		}
		seen := map[string]struct{}{}
		sources := []pathCleanupSource{}
		for _, hit := range search.Hits {
			if _, ok := seen[hit.DocID]; ok {
				continue
			}
			seen[hit.DocID] = struct{}{}
			document, err := client.GetDocument(ctx, hit.DocID)
			if err != nil {
				return nil, "", err
			}
			sources = append(sources, pathCleanupSourceFromDocument(document))
		}
		return sources, "query:" + options.Query, nil
	}
}

func pathCleanupSourceFromDocument(document domain.Document) pathCleanupSource {
	return pathCleanupSource{
		DocID:    document.DocID,
		Path:     document.Path,
		Title:    document.Title,
		Metadata: document.Metadata,
	}
}

func pathCleanupSourceFromSummary(document domain.DocumentSummary) pathCleanupSource {
	return pathCleanupSource{
		DocID:    document.DocID,
		Path:     document.Path,
		Title:    document.Title,
		Metadata: document.Metadata,
	}
}

func pathCleanupCandidate(ctx context.Context, client *runclient.Client, options PathCleanupOptions, source pathCleanupSource, rank int) (PathCleanupCandidate, error) {
	targetPath, action, reason := proposedPathCleanupTarget(options, source)
	candidate := PathCleanupCandidate{
		Rank:               rank,
		DocID:              source.DocID,
		CurrentPath:        source.Path,
		Title:              source.Title,
		ProposedTargetPath: targetPath,
		RecommendedAction:  action,
		Confidence:         "low",
		Reason:             reason,
		DuplicateRisk:      "unknown",
		WriteStatus:        "no_write",
	}
	if targetPath == "" || action == "" {
		candidate.DuplicateRisk = "not_applicable"
		return candidate, nil
	}
	plan, err := client.PlanMoveDocument(ctx, domain.MoveDocumentInput{
		Path:          source.Path,
		TargetPath:    targetPath,
		UpdateLinks:   true,
		UpdateIndexes: options.UpdateIndexes,
	})
	if err != nil {
		return PathCleanupCandidate{}, err
	}
	converted := toDocumentMovePlan(plan)
	candidate.MovePlan = &converted
	candidate.DuplicateRisk = plan.DuplicateRisk
	candidate.Confidence = pathCleanupConfidence(plan, reason)
	if candidate.Confidence != "low" && plan.DuplicateRisk == "none" {
		candidate.NextRequest = pathCleanupNextRequest(action, source.Path, targetPath, options.UpdateIndexes)
	}
	return candidate, nil
}

func markPathCleanupTargetCollisions(candidates []PathCleanupCandidate) {
	targetCounts := map[string]int{}
	for _, candidate := range candidates {
		if candidate.NextRequest == "" || candidate.DuplicateRisk != "none" || candidate.ProposedTargetPath == "" {
			continue
		}
		targetCounts[candidate.ProposedTargetPath]++
	}
	for idx := range candidates {
		if targetCounts[candidates[idx].ProposedTargetPath] <= 1 {
			continue
		}
		candidates[idx].DuplicateRisk = "target_collision_in_plan"
		candidates[idx].Confidence = "low"
		candidates[idx].NextRequest = ""
		candidates[idx].Reason = appendPathCleanupReason(candidates[idx].Reason, "target_collision_in_plan")
	}
}

func appendPathCleanupReason(existing string, reason string) string {
	if existing == "" {
		return reason
	}
	if strings.Contains(existing, reason) {
		return existing
	}
	return existing + ";" + reason
}

func proposedPathCleanupTarget(options PathCleanupOptions, source pathCleanupSource) (string, string, string) {
	cleanupKind := options.CleanupKind
	if cleanupKind == pathCleanupKindAuto && strings.HasPrefix(source.Path, "notes/candidates/") {
		cleanupKind = pathCleanupKindCandidatePromotion
	}
	slug := pathCleanupSlug(source)
	if slug == "" {
		return "", "", "missing_title_or_path_slug"
	}
	switch cleanupKind {
	case pathCleanupKindCandidatePromotion:
		if !strings.HasPrefix(source.Path, "notes/candidates/") {
			return "", "", "source_not_under_notes_candidates"
		}
		prefix := options.TargetPrefix
		if prefix == "" {
			prefix = pathCleanupDefaultTargetPrefix(source)
		}
		targetPath := prefix + slug + ".md"
		if strings.HasPrefix(targetPath, "notes/candidates/") {
			return "", "", "target_prefix_still_under_notes_candidates"
		}
		return targetPath, DocumentTaskActionPromoteCandidate, "candidate_promotion_title_slug"
	default:
		targetPath := path.Join(path.Dir(source.Path), slug+".md")
		return targetPath, DocumentTaskActionRenameDocument, "same_directory_title_slug"
	}
}

func pathCleanupSlug(source pathCleanupSource) string {
	title := strings.TrimSpace(source.Metadata["title"])
	if title == "" {
		title = strings.TrimSpace(source.Title)
	}
	if title == "" {
		title = strings.TrimSuffix(path.Base(source.Path), path.Ext(source.Path))
	}
	slug := slugifyPlacementLabel(title)
	if slug == "source" && title == "" {
		return ""
	}
	return slug
}

func pathCleanupDefaultTargetPrefix(source pathCleanupSource) string {
	lower := strings.ToLower(source.Title + " " + source.Path + " " + strings.Join(metadataValues(source.Metadata), " "))
	if strings.Contains(lower, "project") || strings.Contains(lower, "idea") {
		return "notes/projects/"
	}
	return "notes/"
}

func metadataValues(metadata map[string]string) []string {
	values := make([]string, 0, len(metadata))
	for _, value := range metadata {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func pathCleanupConfidence(plan domain.DocumentMovePlan, reason string) string {
	switch {
	case plan.DuplicateRisk == "same_path":
		return "low"
	case plan.DuplicateRisk != "none":
		return "low"
	case strings.Contains(reason, "title_slug"):
		return "high"
	default:
		return "medium"
	}
}

func pathCleanupNextRequest(action string, currentPath string, targetPath string, updateIndexes bool) string {
	payload := map[string]any{
		"action": action,
		"move": map[string]any{
			"path":           currentPath,
			"target_path":    targetPath,
			"update_indexes": updateIndexes,
		},
	}
	encoded, _ := json.Marshal(payload)
	return string(encoded)
}

func pathCleanupApprovalBoundary(options PathCleanupOptions, autonomy AutonomyModes) string {
	if options.Mode == pathCleanupModeApply {
		return fmt.Sprintf("apply mode allowed by autonomy.approval_mode %s; only duplicate-risk none candidates with exact next_request are applied", autonomy.ApprovalMode)
	}
	return "plan mode is read-only; apply mode requires autonomy.approval_mode autonomous_trusted or autonomous_disposable"
}

func pathCleanupHandoff(plan PathCleanupPlan) *AgentHandoff {
	evidence := []string{
		"mode=" + plan.Mode,
		"cleanup_kind=" + plan.CleanupKind,
		"scope=" + plan.Scope,
		fmt.Sprintf("candidates=%d", len(plan.Candidates)),
		"write_status=" + plan.WriteStatus,
	}
	if plan.AppliedCount > 0 {
		evidence = append(evidence, fmt.Sprintf("applied_count=%d", plan.AppliedCount))
	}
	for _, candidate := range plan.Candidates {
		evidence = append(evidence, fmt.Sprintf("%s->%s duplicate=%s confidence=%s", candidate.CurrentPath, candidate.ProposedTargetPath, candidate.DuplicateRisk, candidate.Confidence))
	}
	return &AgentHandoff{
		AnswerSummary:               "Path cleanup planned runner-owned rename, move, or candidate promotion candidates.",
		Evidence:                    evidence,
		ValidationBoundaries:        plan.ValidationBoundaries,
		AuthorityLimits:             plan.AuthorityLimits,
		FollowUpPrimitiveInspection: "Use returned next_request only after approval or autonomous apply eligibility; inspect move_plan for link, index, duplicate, and projection details.",
	}
}

func normalizePathCleanupMode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return pathCleanupModePlan
	}
	return value
}

func normalizePathCleanupKind(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return pathCleanupKindAuto
	}
	return value
}
