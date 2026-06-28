package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func RunDocumentTask(ctx context.Context, config runclient.Config, request DocumentTaskRequest) (DocumentTaskResult, error) {
	normalized, rejection := normalizeDocumentTaskRequest(request)
	if rejection != "" {
		return DocumentTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			ExampleRequest:  documentTaskExampleRequest(request.Action),
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == DocumentTaskActionValidate {
		return runDocumentValidationTask(ctx, config, request, normalized)
	}

	autonomy, err := applyProfileDefaults(ctx, config, request.Autonomy)
	if err != nil {
		return DocumentTaskResult{}, err
	}
	request.Autonomy = autonomy
	normalized, rejection = normalizeDocumentTaskRequest(request)
	if rejection != "" {
		return DocumentTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			ExampleRequest:  documentTaskExampleRequest(request.Action),
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == DocumentTaskActionResolvePaths {
		paths, err := runclient.ResolvePaths(config)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toPaths(paths)
		return DocumentTaskResult{
			Paths:   &converted,
			Summary: "resolved OpenClerk paths",
		}, nil
	}

	if normalized.Action == DocumentTaskActionGitLifecycle && normalized.GitLifecycle.Mode == gitLifecycleModeCheckpoint {
		var result DocumentTaskResult
		err := runclient.WithWriteLock(ctx, config, func() error {
			client, err := runclient.OpenForWrite(config)
			if err != nil {
				return err
			}
			defer func() {
				_ = client.Close()
			}()
			report, err := runGitLifecycleReport(ctx, client.Paths().VaultRoot, normalized.GitLifecycle, config)
			if err != nil {
				return err
			}
			result = DocumentTaskResult{
				GitLifecycle: &report,
				Summary:      gitLifecycleSummary(report),
			}
			return nil
		})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return result, nil
	}

	if isMutatingDocumentAction(normalized) {
		var result DocumentTaskResult
		err := runclient.WithWriteLock(ctx, config, func() error {
			client, err := runclient.OpenForWrite(config)
			if err != nil {
				return err
			}
			defer func() {
				_ = client.Close()
			}()
			result, err = runMutatingDocumentTask(ctx, client, normalized)
			return err
		})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return result, nil
	}

	client, err := runclient.OpenReadOnly(config)
	if err != nil {
		return DocumentTaskResult{}, err
	}
	defer func() {
		_ = client.Close()
	}()

	switch normalized.Action {
	case DocumentTaskActionIngestSourceURL:
		if normalized.Source.Mode == "inspect" {
			plan, err := runSourceURLIntakePlan(ctx, client, normalized.Source)
			if err != nil {
				return DocumentTaskResult{}, err
			}
			return DocumentTaskResult{
				SourceIntakePlan: &plan,
				Summary:          sourceURLIntakePlanSummary(plan),
			}, nil
		}
		plan, err := runSourcePlacementPlan(ctx, client, normalized.Source)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			SourcePlacement: &plan,
			Summary:         "planned source URL placement",
		}, nil
	case DocumentTaskActionList:
		documents, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix:    normalized.List.PathPrefix,
			MetadataKey:   normalized.List.MetadataKey,
			MetadataValue: normalized.List.MetadataValue,
			Tag:           normalized.List.Tag,
			Limit:         normalized.List.Limit,
			Cursor:        normalized.List.Cursor,
		})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			Documents: toDocumentSummaries(documents.Documents),
			PageInfo:  toPageInfo(documents.PageInfo),
			Summary:   fmt.Sprintf("returned %d documents", len(documents.Documents)),
		}, nil
	case DocumentTaskActionGet:
		document, err := getDocumentByNormalizedTarget(ctx, client, normalized)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		return DocumentTaskResult{
			Document: &converted,
			Summary:  fmt.Sprintf("returned document %s", converted.DocID),
		}, nil
	case DocumentTaskActionReplaceSection:
		return runDocumentSectionReplacement(ctx, client, normalized, "dry_run")
	case DocumentTaskActionReplaceDocument:
		return runDocumentReplacement(ctx, client, normalized, "dry_run")
	case DocumentTaskActionGitLifecycle:
		report, err := runGitLifecycleReport(ctx, client.Paths().VaultRoot, normalized.GitLifecycle, config)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			GitLifecycle: &report,
			Summary:      gitLifecycleSummary(report),
		}, nil
	case DocumentTaskActionWebSearchPlan:
		plan, err := runWebSearchPlan(ctx, client, normalized.WebSearch)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			WebSearchPlan: &plan,
			Summary:       webSearchPlanSummary(plan),
		}, nil
	case DocumentTaskActionArtifactPlan:
		plan, err := runArtifactCandidatePlan(ctx, client, config, normalized.Artifact)
		if err != nil {
			var domainErr *domain.Error
			if errors.As(err, &domainErr) && domainErr.Code == "validation_error" {
				return DocumentTaskResult{
					Rejected:        true,
					RejectionReason: domainErr.Message,
					Summary:         domainErr.Message,
				}, nil
			}
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			ArtifactPlan: &plan,
			Summary:      artifactCandidatePlanSummary(plan),
		}, nil
	case DocumentTaskActionPlanMoveDocument:
		plan, err := client.PlanMoveDocument(ctx, toDomainMoveDocumentInput(normalized.Move))
		if err != nil {
			return DocumentTaskResult{}, err
		}
		if rejection := rejectMovePlanForAction(normalized.Action, plan); rejection != "" {
			return DocumentTaskResult{
				Rejected:        true,
				RejectionReason: rejection,
				MovePlan:        ptr(toDocumentMovePlan(plan)),
				Summary:         rejection,
			}, nil
		}
		converted := toDocumentMovePlan(plan)
		return DocumentTaskResult{
			MovePlan: &converted,
			Summary:  fmt.Sprintf("planned move from %s to %s", converted.SourcePath, converted.TargetPath),
		}, nil
	case DocumentTaskActionPlanPathCleanup:
		plan, err := runPathCleanupPlan(ctx, client, normalized.PathCleanup, normalized.Autonomy)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			PathCleanup: &plan,
			Summary:     fmt.Sprintf("planned path cleanup for %d documents", len(plan.Candidates)),
		}, nil
	case DocumentTaskActionInspectLayout:
		layout, err := inspectKnowledgeLayout(ctx, client)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		status := "valid"
		if !layout.Valid {
			status = "invalid"
		}
		return DocumentTaskResult{
			Layout:  &layout,
			Summary: fmt.Sprintf("inspected %s OpenClerk knowledge layout", status),
		}, nil
	default:
		return DocumentTaskResult{}, fmt.Errorf("unsupported document task action %q", normalized.Action)
	}
}

func isMutatingDocumentAction(normalized normalizedDocumentTaskRequest) bool {
	switch normalized.Action {
	case DocumentTaskActionCreate,
		DocumentTaskActionIngestVideoURL,
		DocumentTaskActionCompileSynthesis,
		DocumentTaskActionValidationSynthesis,
		DocumentTaskActionAppend,
		DocumentTaskActionMoveDocument,
		DocumentTaskActionRenameDocument,
		DocumentTaskActionPromoteCandidate:
		return true
	case DocumentTaskActionReplaceDocument, DocumentTaskActionReplaceSection:
		return !normalized.DryRun
	case DocumentTaskActionPlanPathCleanup:
		return normalized.PathCleanup.Mode == pathCleanupModeApply
	case DocumentTaskActionGitLifecycle:
		return normalized.GitLifecycle.Mode == gitLifecycleModeCheckpoint
	case DocumentTaskActionIngestSourceURL:
		return normalized.Source.Mode != "plan" && normalized.Source.Mode != "inspect"
	default:
		return false
	}
}

func runMutatingDocumentTask(ctx context.Context, client *runclient.Client, normalized normalizedDocumentTaskRequest) (DocumentTaskResult, error) {
	switch normalized.Action {
	case DocumentTaskActionCreate:
		document, err := client.CreateDocument(ctx, domain.CreateDocumentInput{
			Path:  normalized.Document.Path,
			Title: normalized.Document.Title,
			Body:  normalized.Document.Body,
		})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		return DocumentTaskResult{
			Document: &converted,
			Summary:  fmt.Sprintf("created document %s", converted.DocID),
		}, nil
	case DocumentTaskActionIngestSourceURL:
		ingestion, err := client.IngestSourceURL(ctx, domain.SourceURLInput{
			URL:           normalized.Source.URL,
			PathHint:      normalized.Source.PathHint,
			AssetPathHint: normalized.Source.AssetPathHint,
			Title:         normalized.Source.Title,
			Mode:          normalized.Source.Mode,
			SourceType:    normalized.Source.SourceType,
		})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toSourceIngestionResult(ingestion)
		return DocumentTaskResult{
			Ingestion: &converted,
			Summary:   fmt.Sprintf("ingested source URL into %s", converted.SourcePath),
		}, nil
	case DocumentTaskActionIngestVideoURL:
		ingestion, err := client.IngestVideoURL(ctx, domain.VideoURLInput{
			URL:           normalized.Video.URL,
			PathHint:      normalized.Video.PathHint,
			AssetPathHint: normalized.Video.AssetPathHint,
			Title:         normalized.Video.Title,
			Mode:          normalized.Video.Mode,
			Transcript: domain.VideoTranscriptInput{
				Text:       normalized.Video.Transcript.Text,
				Policy:     normalized.Video.Transcript.Policy,
				Origin:     normalized.Video.Transcript.Origin,
				Language:   normalized.Video.Transcript.Language,
				CapturedAt: normalized.Video.Transcript.CapturedAt,
				Tool:       normalized.Video.Transcript.Tool,
				Model:      normalized.Video.Transcript.Model,
				SHA256:     normalized.Video.Transcript.SHA256,
			},
		})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toVideoIngestionResult(ingestion)
		return DocumentTaskResult{
			VideoIngestion: &converted,
			Summary:        fmt.Sprintf("ingested video URL into %s", converted.SourcePath),
		}, nil
	case DocumentTaskActionCompileSynthesis:
		compiled, err := runCompileSynthesis(ctx, client, normalized.Synthesis)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			CompileSynthesis: &compiled,
			Summary:          fmt.Sprintf("compiled synthesis %s", compiled.SelectedPath),
		}, nil
	case DocumentTaskActionValidationSynthesis:
		compiled, err := runValidationSynthesis(ctx, client, normalized.ValidationSynthesis)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			ValidationSynthesis: &compiled,
			Summary:             fmt.Sprintf("validated synthesis %s", compiled.SelectedPath),
		}, nil
	case DocumentTaskActionAppend:
		before, err := client.GetDocument(ctx, normalized.DocID)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		document, err := client.AppendDocument(ctx, normalized.DocID, domain.AppendDocumentInput{Content: normalized.Content})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		update := documentUpdateResult(DocumentTaskActionAppend, before, document, normalized, "applied")
		return DocumentTaskResult{
			Document:       &converted,
			DocumentUpdate: &update,
			Summary:        fmt.Sprintf("appended document %s", converted.DocID),
		}, nil
	case DocumentTaskActionReplaceDocument:
		return runDocumentReplacement(ctx, client, normalized, "applied")
	case DocumentTaskActionReplaceSection:
		return runDocumentSectionReplacement(ctx, client, normalized, "applied")
	case DocumentTaskActionMoveDocument, DocumentTaskActionRenameDocument, DocumentTaskActionPromoteCandidate:
		plan, err := client.PlanMoveDocument(ctx, toDomainMoveDocumentInput(normalized.Move))
		if err != nil {
			return DocumentTaskResult{}, err
		}
		if rejection := rejectMovePlanForAction(normalized.Action, plan); rejection != "" {
			return DocumentTaskResult{
				Rejected:        true,
				RejectionReason: rejection,
				MovePlan:        ptr(toDocumentMovePlan(plan)),
				Summary:         rejection,
			}, nil
		}
		moved, err := client.MoveDocument(ctx, toDomainMoveDocumentInput(normalized.Move))
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocumentMoveResult(moved)
		return DocumentTaskResult{
			MoveResult: &converted,
			Summary:    fmt.Sprintf("moved document from %s to %s", converted.Plan.SourcePath, converted.Plan.TargetPath),
		}, nil
	case DocumentTaskActionPlanPathCleanup:
		plan, err := runPathCleanupPlan(ctx, client, normalized.PathCleanup, normalized.Autonomy)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			PathCleanup: &plan,
			Summary:     fmt.Sprintf("applied path cleanup to %d documents", plan.AppliedCount),
		}, nil
	default:
		return DocumentTaskResult{}, fmt.Errorf("unsupported mutating document task action %q", normalized.Action)
	}
}

func runDocumentValidationTask(ctx context.Context, config runclient.Config, request DocumentTaskRequest, normalized normalizedDocumentTaskRequest) (DocumentTaskResult, error) {
	if !looksLikeDocumentUpdateValidation(request, normalized) {
		return DocumentTaskResult{Summary: "valid"}, nil
	}
	client, err := runclient.OpenReadOnly(config)
	if err != nil {
		return DocumentTaskResult{}, err
	}
	defer func() {
		_ = client.Close()
	}()
	if strings.TrimSpace(normalized.Body) != "" {
		normalized.Action = DocumentTaskActionReplaceDocument
		normalized.DryRun = true
		return runDocumentReplacement(ctx, client, normalized, "validated_no_write")
	}
	if normalized.DocID == "" && normalized.Path != "" {
		document, err := client.GetDocumentByPath(ctx, normalized.Path)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		normalized.DocID = document.DocID
	}
	normalized.Action = DocumentTaskActionReplaceSection
	normalized.DryRun = true
	return runDocumentSectionReplacement(ctx, client, normalized, "validated_no_write")
}

func looksLikeDocumentUpdateValidation(request DocumentTaskRequest, normalized normalizedDocumentTaskRequest) bool {
	hasTarget := normalized.DocID != "" || normalized.Path != ""
	if !hasTarget {
		return false
	}
	if strings.TrimSpace(normalized.Body) != "" {
		return true
	}
	if normalized.Heading == "" {
		return false
	}
	if request.jsonDecoded {
		return request.fieldProvided("content")
	}
	return true
}

func runDocumentReplacement(ctx context.Context, client *runclient.Client, normalized normalizedDocumentTaskRequest, writeStatus string) (DocumentTaskResult, error) {
	before, err := resolveDocumentUpdateTarget(ctx, client, normalized)
	if err != nil {
		if rejected, ok := rejectedDocumentResultFromError(err, DocumentTaskActionReplaceDocument); ok {
			return rejected, nil
		}
		return DocumentTaskResult{}, err
	}
	dryRun := writeStatus != "applied" || normalized.DryRun
	after, err := client.ReplaceDocument(ctx, before.DocID, domain.ReplaceDocumentInput{
		Path:             normalized.Path,
		Title:            normalized.Title,
		Body:             normalized.Body,
		Metadata:         normalized.Metadata,
		AllowDocIDChange: normalized.AllowDocIDChange,
		DryRun:           dryRun,
	})
	if err != nil {
		if rejected, ok := rejectedDocumentResultFromError(err, DocumentTaskActionReplaceDocument); ok {
			return rejected, nil
		}
		return DocumentTaskResult{}, err
	}
	converted := toDocument(after)
	update := documentUpdateResult(DocumentTaskActionReplaceDocument, before, after, normalized, writeStatus)
	return DocumentTaskResult{
		Document:       &converted,
		DocumentUpdate: &update,
		AgentHandoff:   update.AgentHandoff,
		Summary:        documentUpdateSummary("document replacement", after.DocID, writeStatus),
	}, nil
}

func runDocumentSectionReplacement(ctx context.Context, client *runclient.Client, normalized normalizedDocumentTaskRequest, writeStatus string) (DocumentTaskResult, error) {
	before, err := client.GetDocument(ctx, normalized.DocID)
	if err != nil {
		return DocumentTaskResult{}, err
	}
	includeSubsections := normalized.IncludeSubsections
	dryRun := writeStatus != "applied" || normalized.DryRun
	after, err := client.ReplaceSection(ctx, normalized.DocID, domain.ReplaceSectionInput{
		Heading:            normalized.Heading,
		Content:            normalized.Content,
		IncludeHeading:     normalized.IncludeHeading,
		IncludeSubsections: &includeSubsections,
		DryRun:             dryRun,
	})
	if err != nil {
		if rejected, ok := rejectedDocumentResultFromError(err, DocumentTaskActionReplaceSection); ok {
			return rejected, nil
		}
		return DocumentTaskResult{}, err
	}
	converted := toDocument(after)
	update := documentUpdateResult(DocumentTaskActionReplaceSection, before, after, normalized, writeStatus)
	headingPreserved := !normalized.IncludeHeading
	update.Heading = normalized.Heading
	update.HeadingPreserved = &headingPreserved
	update.IncludeHeading = normalized.IncludeHeading
	update.IncludeSubsections = normalized.IncludeSubsections
	return DocumentTaskResult{
		Document:       &converted,
		DocumentUpdate: &update,
		AgentHandoff:   update.AgentHandoff,
		Summary:        documentUpdateSummary("section replacement", after.DocID, writeStatus),
	}, nil
}

func getDocumentByNormalizedTarget(ctx context.Context, client *runclient.Client, normalized normalizedDocumentTaskRequest) (domain.Document, error) {
	if normalized.DocID != "" {
		return client.GetDocument(ctx, normalized.DocID)
	}
	return client.GetDocumentByPath(ctx, normalized.Path)
}

func resolveDocumentUpdateTarget(ctx context.Context, client *runclient.Client, normalized normalizedDocumentTaskRequest) (domain.Document, error) {
	if normalized.DocID == "" {
		return client.GetDocumentByPath(ctx, normalized.Path)
	}
	document, err := client.GetDocument(ctx, normalized.DocID)
	if err != nil {
		return domain.Document{}, err
	}
	if normalized.Path != "" && normalized.Path != document.Path {
		return domain.Document{}, &runclient.Error{
			Code:    "validation_error",
			Message: "replacement path does not match doc_id",
			Status:  400,
			Details: map[string]any{
				"doc_id":        normalized.DocID,
				"current_path":  document.Path,
				"request_path":  normalized.Path,
				"expected_path": document.Path,
			},
		}
	}
	return document, nil
}

func documentUpdateResult(action string, before domain.Document, after domain.Document, normalized normalizedDocumentTaskRequest, writeStatus string) DocumentUpdateResult {
	dryRun := writeStatus != "applied" || normalized.DryRun
	result := DocumentUpdateResult{
		Action:               action,
		DocID:                after.DocID,
		WriteStatus:          writeStatus,
		DryRun:               dryRun,
		Before:               documentUpdateSnapshot(before),
		After:                documentUpdateSnapshot(after),
		Diff:                 compactMarkdownDiff(before.Body, after.Body),
		PreimageSHA256:       sha256Hex(before.Body),
		ApprovalBoundary:     documentUpdateApprovalBoundary(action),
		ValidationBoundaries: documentUpdateValidationBoundaries(action),
		AuthorityLimits:      "OpenClerk updates only the targeted runner-visible markdown document; citations, provenance, and projection freshness remain separate evidence surfaces.",
	}
	if action == DocumentTaskActionReplaceDocument || action == DocumentTaskActionReplaceSection {
		result.NextWriteRequest = nextDocumentUpdateRequest(action, before.DocID, before.Path, normalized)
	}
	if writeStatus == "applied" {
		result.NextWriteRequest = ""
		result.RollbackRequest = rollbackReplaceDocumentRequest(after.DocID, after.Path, before.Body)
	}
	result.AgentHandoff = documentUpdateHandoff(result)
	return result
}

func documentUpdateSnapshot(document domain.Document) DocumentUpdateSnapshot {
	return DocumentUpdateSnapshot{
		Title:    document.Title,
		Path:     document.Path,
		Headings: append([]string(nil), document.Headings...),
	}
}

func documentUpdateSummary(kind string, docID string, writeStatus string) string {
	switch writeStatus {
	case "applied":
		return fmt.Sprintf("applied %s for document %s", kind, docID)
	case "validated_no_write":
		return fmt.Sprintf("validated %s for document %s without writing", kind, docID)
	default:
		return fmt.Sprintf("previewed %s for document %s without writing", kind, docID)
	}
}

func documentUpdateApprovalBoundary(action string) string {
	switch action {
	case DocumentTaskActionReplaceDocument:
		return "validate or dry_run is not durable-write approval; approve and run the returned replace_document next_write_request before replacing the full note"
	case DocumentTaskActionReplaceSection:
		return "validate or dry_run is not durable-write approval; approve and run the returned replace_section next_write_request before replacing the section"
	default:
		return "approved durable writes return a preimage hash and rollback request for OpenClerk-based repair"
	}
}

func documentUpdateValidationBoundaries(action string) string {
	switch action {
	case DocumentTaskActionReplaceDocument:
		return "replace_document validates doc_id/path agreement, parses frontmatter before writing, preserves the current stable id unless allow_doc_id_change=true, and returns before/after title, path, headings, and a compact diff"
	case DocumentTaskActionReplaceSection:
		return "replace_section preserves the matched heading unless include_heading=true, rejects same-level replacement headings when preserving the heading, and can replace nested subsections through include_subsections=true"
	default:
		return "document write receipt reports the targeted document preimage and resulting title, path, and headings"
	}
}

func documentUpdateHandoff(update DocumentUpdateResult) *AgentHandoff {
	followUp := "Use the returned rollback_request with openclerk document if the applied edit is undesired."
	if update.WriteStatus != "applied" {
		followUp = "Approve and run next_write_request for the durable write; do not substitute another mutation primitive."
		if update.Action == DocumentTaskActionReplaceDocument {
			followUp = "Approve and run next_write_request for the durable full-note replacement; do not use replace_section for this full-note update."
		}
		if update.Action == DocumentTaskActionReplaceSection && update.HeadingPreserved != nil {
			followUp = fmt.Sprintf("Approve and run next_write_request for the durable section update; heading_preserved=%t.", *update.HeadingPreserved)
		}
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf("%s %s for %s -> %s", update.Action, update.WriteStatus, update.Before.Path, update.After.Path),
		Evidence: []string{
			"before_title:" + update.Before.Title,
			"before_path:" + update.Before.Path,
			"after_title:" + update.After.Title,
			"after_path:" + update.After.Path,
			"preimage_sha256:" + update.PreimageSHA256,
		},
		ValidationBoundaries:        update.ValidationBoundaries,
		AuthorityLimits:             update.AuthorityLimits,
		FollowUpPrimitiveInspection: followUp,
	}
}

func nextDocumentUpdateRequest(action string, docID string, docPath string, normalized normalizedDocumentTaskRequest) string {
	switch action {
	case DocumentTaskActionReplaceDocument:
		request := map[string]any{
			"action": DocumentTaskActionReplaceDocument,
			"doc_id": docID,
			"body":   normalized.Body,
		}
		if docPath != "" {
			request["path"] = docPath
		}
		if normalized.Title != "" {
			request["title"] = normalized.Title
		}
		if len(normalized.Metadata) > 0 {
			request["metadata"] = normalized.Metadata
		}
		if normalized.AllowDocIDChange {
			request["allow_doc_id_change"] = true
		}
		return marshalRequestString(request)
	case DocumentTaskActionReplaceSection:
		request := map[string]any{
			"action":              DocumentTaskActionReplaceSection,
			"doc_id":              docID,
			"heading":             normalized.Heading,
			"content":             normalized.Content,
			"include_subsections": normalized.IncludeSubsections,
		}
		if normalized.IncludeHeading {
			request["include_heading"] = true
		}
		return marshalRequestString(request)
	default:
		return ""
	}
}

func rollbackReplaceDocumentRequest(docID string, docPath string, body string) string {
	return marshalRequestString(map[string]any{
		"action":              DocumentTaskActionReplaceDocument,
		"doc_id":              docID,
		"path":                docPath,
		"body":                body,
		"allow_doc_id_change": true,
	})
}

func marshalRequestString(request map[string]any) string {
	encoded, err := json.Marshal(request)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func compactMarkdownDiff(before string, after string) string {
	if before == after {
		return "no content changes"
	}
	beforeLines := splitDiffLines(before)
	afterLines := splitDiffLines(after)
	prefix := 0
	for prefix < len(beforeLines) && prefix < len(afterLines) && beforeLines[prefix] == afterLines[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(beforeLines)-prefix && suffix < len(afterLines)-prefix &&
		beforeLines[len(beforeLines)-1-suffix] == afterLines[len(afterLines)-1-suffix] {
		suffix++
	}
	removed := beforeLines[prefix : len(beforeLines)-suffix]
	added := afterLines[prefix : len(afterLines)-suffix]
	var builder strings.Builder
	_, _ = fmt.Fprintf(&builder, "@@ line %d; -%d +%d @@\n", prefix+1, len(removed), len(added))
	appendDiffLines(&builder, "-", removed)
	appendDiffLines(&builder, "+", added)
	return strings.TrimRight(builder.String(), "\n")
}

func splitDiffLines(value string) []string {
	trimmed := strings.TrimRight(value, "\n")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "\n")
}

func appendDiffLines(builder *strings.Builder, prefix string, lines []string) {
	const maxDiffLines = 8
	for idx, line := range lines {
		if idx >= maxDiffLines {
			_, _ = fmt.Fprintf(builder, "%s ... (%d more lines)\n", prefix, len(lines)-maxDiffLines)
			return
		}
		_, _ = fmt.Fprintf(builder, "%s %s\n", prefix, line)
	}
}

func rejectedDocumentResultFromError(err error, action string) (DocumentTaskResult, bool) {
	var runErr *runclient.Error
	if !errors.As(err, &runErr) || runErr.Code != "validation_error" {
		return DocumentTaskResult{}, false
	}
	return DocumentTaskResult{
		Rejected:        true,
		RejectionReason: runErr.Message,
		ExampleRequest:  documentTaskExampleRequest(action),
		Summary:         runErr.Message,
	}, true
}

func toDomainMoveDocumentInput(options MoveDocumentOptions) domain.MoveDocumentInput {
	updateLinks := true
	if options.UpdateLinks != nil {
		updateLinks = *options.UpdateLinks
	}
	return domain.MoveDocumentInput{
		DocID:         options.DocID,
		Path:          options.Path,
		TargetPath:    options.TargetPath,
		UpdateLinks:   updateLinks,
		UpdateIndexes: options.UpdateIndexes,
	}
}

func rejectMovePlanForAction(action string, plan domain.DocumentMovePlan) string {
	switch action {
	case DocumentTaskActionRenameDocument:
		if path.Dir(plan.SourcePath) != path.Dir(plan.TargetPath) {
			return "rename_document target_path must stay in the same directory as the current path"
		}
	case DocumentTaskActionPromoteCandidate:
		if !strings.HasPrefix(plan.SourcePath, "notes/candidates/") {
			return "promote_candidate source path must be under notes/candidates/"
		}
		if strings.HasPrefix(plan.TargetPath, "notes/candidates/") {
			return "promote_candidate target_path must leave notes/candidates/"
		}
	}
	return ""
}

func ptr[T any](value T) *T {
	return &value
}

type normalizedDocumentTaskRequest struct {
	Action              string
	Autonomy            AutonomyModes
	Document            DocumentInput
	Source              SourceURLInput
	Video               VideoURLInput
	Synthesis           CompileSynthesisInput
	ValidationSynthesis ValidationSynthesisInput
	GitLifecycle        GitLifecycleOptions
	WebSearch           WebSearchPlanOptions
	Artifact            ArtifactPlanOptions
	Move                MoveDocumentOptions
	PathCleanup         PathCleanupOptions
	DocID               string
	Path                string
	Title               string
	Body                string
	Metadata            map[string]string
	Content             string
	Heading             string
	IncludeHeading      bool
	IncludeSubsections  bool
	DryRun              bool
	Diff                bool
	AllowDocIDChange    bool
	List                DocumentListOptions
}

func normalizeDocumentTaskRequest(request DocumentTaskRequest) (normalizedDocumentTaskRequest, string) {
	action := strings.TrimSpace(request.Action)
	if action == "" {
		action = DocumentTaskActionValidate
	}
	includeSubsections := true
	if request.IncludeSubsections != nil {
		includeSubsections = *request.IncludeSubsections
	}
	normalized := normalizedDocumentTaskRequest{
		Action:              action,
		Autonomy:            request.Autonomy,
		Document:            request.Document,
		Source:              trimSourceURLInput(request.Source),
		Video:               trimVideoURLInput(request.Video),
		Synthesis:           trimCompileSynthesisInput(compileSynthesisInputFromRequest(request)),
		ValidationSynthesis: trimValidationSynthesisInput(request.ValidationSynthesis),
		GitLifecycle:        trimGitLifecycleOptions(request.GitLifecycle),
		WebSearch:           trimWebSearchPlanOptions(request.WebSearch),
		Artifact:            trimArtifactPlanOptions(request.Artifact),
		Move:                trimMoveDocumentOptions(moveDocumentOptionsFromRequest(request)),
		PathCleanup:         trimPathCleanupOptions(pathCleanupOptionsFromRequest(request)),
		DocID:               strings.TrimSpace(request.DocID),
		Path:                normalizeVaultRelativePath(firstNonEmpty(request.Path, request.Document.Path)),
		Title:               strings.TrimSpace(firstNonEmpty(request.Title, request.Document.Title)),
		Body:                firstNonEmpty(request.Body, request.Document.Body),
		Metadata:            replacementMetadataFromRequest(request),
		Content:             request.Content,
		Heading:             strings.TrimSpace(request.Heading),
		IncludeHeading:      request.IncludeHeading,
		IncludeSubsections:  includeSubsections,
		DryRun:              request.DryRun || request.Diff,
		Diff:                request.Diff,
		AllowDocIDChange:    request.AllowDocIDChange,
		List:                request.List,
	}

	autonomy, rejection := normalizeAutonomyModes(request.Autonomy)
	if rejection != "" {
		return normalizedDocumentTaskRequest{}, rejection
	}
	normalized.Autonomy = autonomy

	if rejection := rejectNegativeRunnerLimits(
		request.List.Limit,
		request.GitLifecycle.Limit,
		request.WebSearch.Limit,
		request.Artifact.Limit,
		request.PathCleanup.Limit,
	); rejection != "" {
		return normalizedDocumentTaskRequest{}, rejection
	}
	if rejection := rejectUnsupportedDocumentPreviewFields(request, action); rejection != "" {
		return normalizedDocumentTaskRequest{}, rejection
	}

	switch action {
	case DocumentTaskActionValidate:
		if request.Document != (DocumentInput{}) {
			return normalized, validateDocumentInput(request.Document)
		}
		return normalized, ""
	case DocumentTaskActionCreate:
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := validateDocumentInput(request.Document); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionIngestSourceURL:
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := validateSourceURLInput(normalized.Source); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionIngestVideoURL:
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := validateVideoURLInput(normalized.Video); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionCompileSynthesis:
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := validateCompileSynthesisInput(normalized.Synthesis); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionValidationSynthesis:
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := validateValidationSynthesisInput(normalized.ValidationSynthesis); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionList:
		if rejection := normalizeDocumentListTagFilter(&normalized.List); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionGet:
		if rejection := rejectIgnoredDocumentFields(request, map[string]bool{
			"action": true, "doc_id": true, "path": true,
		}); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if normalized.DocID == "" && normalized.Path == "" {
			return normalizedDocumentTaskRequest{}, "doc_id or path is required"
		}
		if normalized.DocID != "" && normalized.Path != "" {
			return normalizedDocumentTaskRequest{}, "get_document accepts only one of doc_id or path"
		}
		return normalized, ""
	case DocumentTaskActionAppend:
		if rejection := rejectIgnoredDocumentFields(request, map[string]bool{
			"action": true, "autonomy": true, "doc_id": true, "content": true,
		}); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if normalized.DocID == "" {
			return normalizedDocumentTaskRequest{}, "doc_id is required"
		}
		if request.jsonDecoded && !request.fieldProvided("content") {
			return normalizedDocumentTaskRequest{}, "content is required"
		}
		if strings.TrimSpace(normalized.Content) == "" {
			return normalizedDocumentTaskRequest{}, "content is required"
		}
		return normalized, ""
	case DocumentTaskActionReplaceDocument:
		if rejection := rejectIgnoredDocumentFields(request, map[string]bool{
			"action": true, "autonomy": true, "doc_id": true, "path": true, "title": true, "body": true,
			"metadata": true, "document": true, "allow_doc_id_change": true, "dry_run": true, "diff": true,
		}); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if normalized.DocID == "" {
			return normalizedDocumentTaskRequest{}, "doc_id is required"
		}
		if request.jsonDecoded && !request.fieldProvided("body") && !request.fieldProvided("document") {
			return normalizedDocumentTaskRequest{}, "body or document.body is required"
		}
		if strings.TrimSpace(normalized.Body) == "" {
			return normalizedDocumentTaskRequest{}, "body is required"
		}
		if normalized.Path != "" {
			if _, rejection := validateRequiredVaultPath(normalized.Path, "path is required", "path must be relative to the vault root", "path must stay inside the vault root"); rejection != "" {
				return normalizedDocumentTaskRequest{}, rejection
			}
			if path.Ext(normalized.Path) != ".md" {
				return normalizedDocumentTaskRequest{}, "path must be a vault-relative markdown path"
			}
		}
		return normalized, ""
	case DocumentTaskActionReplaceSection:
		if rejection := rejectIgnoredDocumentFields(request, map[string]bool{
			"action": true, "autonomy": true, "doc_id": true, "heading": true, "content": true,
			"include_heading": true, "include_subsections": true,
			"dry_run": true, "diff": true,
		}); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if normalized.DocID == "" {
			return normalizedDocumentTaskRequest{}, "doc_id is required"
		}
		if normalized.Heading == "" {
			return normalizedDocumentTaskRequest{}, "heading is required"
		}
		if request.jsonDecoded && !request.fieldProvided("content") {
			return normalizedDocumentTaskRequest{}, "content is required"
		}
		return normalized, ""
	case DocumentTaskActionResolvePaths:
		return normalized, ""
	case DocumentTaskActionInspectLayout:
		return normalized, ""
	case DocumentTaskActionGitLifecycle:
		if normalized.GitLifecycle.Mode == "" {
			normalized.GitLifecycle.Mode = gitLifecycleModeStatus
		}
		for _, vaultPath := range normalized.GitLifecycle.Paths {
			_, issue := domain.NormalizeVaultRelativePath(vaultPath)
			if issue != domain.VaultPathOK {
				return normalizedDocumentTaskRequest{}, "git_lifecycle.paths entries must stay inside the vault root"
			}
			if isUnsafeGitLifecyclePath(vaultPath) {
				return normalizedDocumentTaskRequest{}, "git_lifecycle.paths entries must be literal vault-relative paths"
			}
		}
		switch normalized.GitLifecycle.Mode {
		case gitLifecycleModeStatus, gitLifecycleModeHistory:
			return normalized, ""
		case gitLifecycleModeCheckpoint:
			if len(normalized.GitLifecycle.Paths) == 0 {
				return normalizedDocumentTaskRequest{}, "git_lifecycle.paths is required for checkpoint mode"
			}
			if strings.TrimSpace(normalized.GitLifecycle.Message) == "" {
				return normalizedDocumentTaskRequest{}, "git_lifecycle.message is required for checkpoint mode"
			}
			if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
				return normalizedDocumentTaskRequest{}, rejection
			}
			return normalized, ""
		default:
			return normalizedDocumentTaskRequest{}, "git_lifecycle.mode must be status, history, or checkpoint"
		}
	case DocumentTaskActionWebSearchPlan:
		if normalized.WebSearch.Query == "" {
			return normalizedDocumentTaskRequest{}, "web_search.query is required"
		}
		if len(normalized.WebSearch.Results) == 0 {
			return normalizedDocumentTaskRequest{}, "web_search.results is required"
		}
		for _, result := range normalized.WebSearch.Results {
			if rejection := validateWebSearchResultInput(result); rejection != "" {
				return normalizedDocumentTaskRequest{}, rejection
			}
		}
		return normalized, ""
	case DocumentTaskActionArtifactPlan:
		if rejection := validateArtifactPlanOptions(normalized.Artifact); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionPlanMoveDocument:
		if rejection := validateMoveDocumentOptions(normalized.Move); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionPlanPathCleanup:
		if rejection := validatePathCleanupOptions(normalized.PathCleanup); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if normalized.PathCleanup.Mode == pathCleanupModeApply {
			if request.Autonomy == (AutonomyModes{}) {
				return normalizedDocumentTaskRequest{}, "path_cleanup.mode apply requires explicit autonomy.approval_mode autonomous_trusted or autonomous_disposable"
			}
			if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
				return normalizedDocumentTaskRequest{}, rejection
			}
			if normalized.Autonomy.ApprovalMode != ApprovalModeAutonomousTrusted &&
				normalized.Autonomy.ApprovalMode != ApprovalModeAutonomousDisposable {
				return normalizedDocumentTaskRequest{}, "path_cleanup.mode apply requires autonomy.approval_mode autonomous_trusted or autonomous_disposable"
			}
		}
		return normalized, ""
	case DocumentTaskActionMoveDocument, DocumentTaskActionRenameDocument, DocumentTaskActionPromoteCandidate:
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if rejection := validateMoveDocumentOptions(normalized.Move); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	default:
		return normalizedDocumentTaskRequest{}, fmt.Sprintf("unsupported document task action %q", action)
	}
}

func rejectDocumentAutonomyWrite(normalized normalizedDocumentTaskRequest) string {
	if !isMutatingDocumentAction(normalized) {
		return ""
	}
	if normalized.Autonomy.ApprovalMode == ApprovalModeProposeOnly {
		return "autonomy.approval_mode propose_only does not allow mutating document actions"
	}
	if normalized.Autonomy.WriteTargetMode != WriteTargetModeExistingOnly {
		return ""
	}
	switch normalized.Action {
	case DocumentTaskActionCreate, DocumentTaskActionIngestSourceURL, DocumentTaskActionIngestVideoURL:
		return "autonomy.write_target_mode existing_only does not allow create-shaped document actions"
	case DocumentTaskActionCompileSynthesis:
		return "autonomy.write_target_mode existing_only does not allow create_or_update compile_synthesis writes"
	case DocumentTaskActionValidationSynthesis:
		return "autonomy.write_target_mode existing_only does not allow create_or_update validation_synthesis writes"
	}
	return ""
}

func compileSynthesisInputFromRequest(request DocumentTaskRequest) CompileSynthesisInput {
	input := request.Synthesis
	if input.Path == "" {
		input.Path = firstNonEmpty(request.Path, request.Document.Path)
	}
	if input.Title == "" {
		input.Title = firstNonEmpty(request.Title, request.Document.Title)
	}
	if input.Body == "" {
		input.Body = firstNonEmpty(request.Body, request.Document.Body)
	}
	if len(input.BodyFacts) == 0 {
		input.BodyFacts = request.BodyFacts
	}
	if len(input.SourceRefs) == 0 {
		input.SourceRefs = request.SourceRefs
	}
	if input.FreshnessNote == "" {
		input.FreshnessNote = request.FreshnessNote
	}
	if input.Mode == "" {
		input.Mode = request.Mode
	}
	return input
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func replacementMetadataFromRequest(request DocumentTaskRequest) map[string]string {
	if len(request.Metadata) == 0 {
		return nil
	}
	metadata := make(map[string]string, len(request.Metadata))
	for key, value := range request.Metadata {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		metadata[key] = strings.TrimSpace(value)
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
}

func rejectIgnoredDocumentFields(request DocumentTaskRequest, allowed map[string]bool) string {
	if !request.jsonDecoded {
		return ""
	}
	fields := make([]string, 0, len(request.providedFields))
	for field := range request.providedFields {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	action := strings.TrimSpace(request.Action)
	if action == "" {
		action = DocumentTaskActionValidate
	}
	for _, field := range fields {
		if allowed[field] {
			continue
		}
		if action == DocumentTaskActionReplaceSection && field == "body" {
			return "replace_section uses content, not body; body is not accepted"
		}
		return fmt.Sprintf("%s does not accept field %q", action, field)
	}
	return ""
}

func rejectUnsupportedDocumentPreviewFields(request DocumentTaskRequest, action string) string {
	if !request.jsonDecoded {
		return ""
	}
	if !request.fieldProvided("dry_run") && !request.fieldProvided("diff") {
		return ""
	}
	switch action {
	case DocumentTaskActionReplaceDocument, DocumentTaskActionReplaceSection:
		return ""
	default:
		if request.fieldProvided("dry_run") {
			return fmt.Sprintf("%s does not support dry_run; use validate, plan mode, or a replacement action with dry_run", action)
		}
		return fmt.Sprintf("%s does not support diff; use replace_document or replace_section with diff=true", action)
	}
}

func documentTaskExampleRequest(action string) string {
	switch strings.TrimSpace(action) {
	case DocumentTaskActionAppend:
		return `{"action":"append_document","doc_id":"doc_id_from_json","content":"## Update\nApproved content."}`
	case DocumentTaskActionReplaceDocument:
		return `{"action":"replace_document","doc_id":"doc_id_from_json","path":"notes/example.md","body":"---\ntitle: \"Example\"\n---\n# Example\n\nApproved full replacement."}`
	case DocumentTaskActionReplaceSection:
		return `{"action":"replace_section","doc_id":"doc_id_from_json","heading":"Summary","content":"Updated summary.","include_subsections":true}`
	case DocumentTaskActionGet:
		return `{"action":"get_document","doc_id":"doc_id_from_json"} or {"action":"get_document","path":"notes/example.md"}`
	case DocumentTaskActionCreate:
		return `{"action":"create_document","document":{"path":"notes/example.md","title":"Example","body":"# Example\n\nBody."}}`
	default:
		return `{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}`
	}
}

func moveDocumentOptionsFromRequest(request DocumentTaskRequest) MoveDocumentOptions {
	options := request.Move
	if options.DocID == "" {
		options.DocID = request.DocID
	}
	if options.Path == "" {
		options.Path = request.Path
	}
	return options
}

func pathCleanupOptionsFromRequest(request DocumentTaskRequest) PathCleanupOptions {
	options := request.PathCleanup
	if options.DocID == "" {
		options.DocID = request.DocID
	}
	if options.Path == "" {
		options.Path = request.Path
	}
	return options
}

func trimMoveDocumentOptions(options MoveDocumentOptions) MoveDocumentOptions {
	trimmed := MoveDocumentOptions{
		DocID:         strings.TrimSpace(options.DocID),
		Path:          normalizeVaultRelativePath(options.Path),
		TargetPath:    normalizeVaultRelativePath(options.TargetPath),
		UpdateIndexes: options.UpdateIndexes,
	}
	if options.UpdateLinks != nil {
		updateLinks := *options.UpdateLinks
		trimmed.UpdateLinks = &updateLinks
	}
	return trimmed
}

func trimPathCleanupOptions(options PathCleanupOptions) PathCleanupOptions {
	return PathCleanupOptions{
		DocID:         strings.TrimSpace(options.DocID),
		Path:          normalizeVaultRelativePath(options.Path),
		PathPrefix:    normalizeVaultRelativePrefix(options.PathPrefix),
		Query:         strings.TrimSpace(options.Query),
		Mode:          normalizePathCleanupMode(options.Mode),
		CleanupKind:   normalizePathCleanupKind(options.CleanupKind),
		TargetPrefix:  normalizeVaultRelativePrefix(options.TargetPrefix),
		UpdateIndexes: options.UpdateIndexes,
		Limit:         options.Limit,
	}
}

func validateMoveDocumentOptions(options MoveDocumentOptions) string {
	if options.DocID == "" && options.Path == "" {
		return "move.doc_id or move.path is required"
	}
	if options.DocID != "" && options.Path != "" {
		return "move accepts only one of doc_id or path"
	}
	if options.Path != "" {
		if _, rejection := validateRequiredVaultPath(options.Path, "move.path is required", "move.path must be relative to the vault root", "move.path must stay inside the vault root"); rejection != "" {
			return rejection
		}
	}
	if options.TargetPath == "" {
		return "move.target_path is required"
	}
	if _, rejection := validateRequiredVaultPath(options.TargetPath, "move.target_path is required", "move.target_path must be relative to the vault root", "move.target_path must stay inside the vault root"); rejection != "" {
		return rejection
	}
	if path.Ext(options.TargetPath) != "" && path.Ext(options.TargetPath) != ".md" {
		return "move.target_path must be a vault-relative markdown path"
	}
	return ""
}

func validatePathCleanupOptions(options PathCleanupOptions) string {
	targets := 0
	for _, value := range []string{options.DocID, options.Path, options.PathPrefix, options.Query} {
		if strings.TrimSpace(value) != "" {
			targets++
		}
	}
	if targets == 0 {
		return "path_cleanup requires one of doc_id, path, path_prefix, or query"
	}
	if targets > 1 {
		return "path_cleanup accepts only one of doc_id, path, path_prefix, or query"
	}
	if options.Path != "" {
		if _, rejection := validateRequiredVaultPath(options.Path, "path_cleanup.path is required", "path_cleanup.path must be relative to the vault root", "path_cleanup.path must stay inside the vault root"); rejection != "" {
			return rejection
		}
		if path.Ext(options.Path) != "" && path.Ext(options.Path) != ".md" {
			return "path_cleanup.path must be a vault-relative markdown path"
		}
	}
	if options.PathPrefix != "" {
		if _, rejection := validateRequiredVaultPath(options.PathPrefix, "path_cleanup.path_prefix is required", "path_cleanup.path_prefix must be relative to the vault root", "path_cleanup.path_prefix must stay inside the vault root"); rejection != "" {
			return rejection
		}
	}
	if options.TargetPrefix != "" {
		if _, rejection := validateRequiredVaultPath(options.TargetPrefix, "path_cleanup.target_prefix is required", "path_cleanup.target_prefix must be relative to the vault root", "path_cleanup.target_prefix must stay inside the vault root"); rejection != "" {
			return rejection
		}
		if !strings.HasSuffix(options.TargetPrefix, "/") {
			return "path_cleanup.target_prefix must end with /"
		}
	}
	if options.Mode != pathCleanupModePlan && options.Mode != pathCleanupModeApply {
		return "path_cleanup.mode must be plan or apply"
	}
	if options.CleanupKind != pathCleanupKindAuto &&
		options.CleanupKind != pathCleanupKindRename &&
		options.CleanupKind != pathCleanupKindCandidatePromotion {
		return "path_cleanup.cleanup_kind must be auto, rename, or candidate_promotion"
	}
	return ""
}

func normalizeDocumentListTagFilter(list *DocumentListOptions) string {
	return normalizeTagFilter("list", list.Tag, list.tagProvided, &list.MetadataKey, &list.MetadataValue, &list.Tag)
}

func trimSourceURLInput(input SourceURLInput) SourceURLInput {
	return SourceURLInput{
		URL:           strings.TrimSpace(input.URL),
		PathHint:      normalizeVaultRelativePath(input.PathHint),
		AssetPathHint: normalizeVaultRelativePath(input.AssetPathHint),
		Title:         strings.TrimSpace(input.Title),
		Mode:          strings.TrimSpace(input.Mode),
		SourceType:    strings.TrimSpace(input.SourceType),
		Limit:         input.Limit,
	}
}

func trimVideoURLInput(input VideoURLInput) VideoURLInput {
	return VideoURLInput{
		URL:           strings.TrimSpace(input.URL),
		PathHint:      normalizeVaultRelativePath(input.PathHint),
		AssetPathHint: normalizeVaultRelativePath(input.AssetPathHint),
		Title:         strings.TrimSpace(input.Title),
		Mode:          strings.TrimSpace(input.Mode),
		Transcript: VideoTranscriptInput{
			Text:       input.Transcript.Text,
			Policy:     strings.TrimSpace(input.Transcript.Policy),
			Origin:     strings.TrimSpace(input.Transcript.Origin),
			Language:   strings.TrimSpace(input.Transcript.Language),
			CapturedAt: strings.TrimSpace(input.Transcript.CapturedAt),
			Tool:       strings.TrimSpace(input.Transcript.Tool),
			Model:      strings.TrimSpace(input.Transcript.Model),
			SHA256:     strings.TrimSpace(input.Transcript.SHA256),
		},
	}
}

func trimCompileSynthesisInput(input CompileSynthesisInput) CompileSynthesisInput {
	sourceRefs := make([]string, 0, len(input.SourceRefs))
	for _, ref := range input.SourceRefs {
		sourceRefs = append(sourceRefs, normalizeCompileSynthesisMarkdownPath(strings.TrimSpace(ref)))
	}
	bodyFacts := make([]string, 0, len(input.BodyFacts))
	for _, fact := range input.BodyFacts {
		trimmed := strings.TrimSpace(fact)
		if trimmed != "" {
			bodyFacts = append(bodyFacts, trimmed)
		}
	}
	return CompileSynthesisInput{
		Path:          normalizeCompileSynthesisMarkdownPath(strings.TrimSpace(input.Path)),
		Title:         strings.TrimSpace(input.Title),
		SourceRefs:    sourceRefs,
		Body:          input.Body,
		BodyFacts:     bodyFacts,
		FreshnessNote: strings.TrimSpace(input.FreshnessNote),
		Mode:          normalizeCompileSynthesisMode(input.Mode),
	}
}

func trimValidationSynthesisInput(input ValidationSynthesisInput) ValidationSynthesisInput {
	sourceRefs := make([]string, 0, len(input.SourceRefs))
	for _, ref := range input.SourceRefs {
		sourceRefs = append(sourceRefs, normalizeCompileSynthesisMarkdownPath(strings.TrimSpace(ref)))
	}
	bodyFacts := make([]string, 0, len(input.BodyFacts))
	for _, fact := range input.BodyFacts {
		trimmed := strings.TrimSpace(fact)
		if trimmed != "" {
			bodyFacts = append(bodyFacts, trimmed)
		}
	}
	return ValidationSynthesisInput{
		DocID:                strings.TrimSpace(input.DocID),
		Path:                 normalizeOptionalCompileSynthesisMarkdownPath(input.Path),
		Title:                strings.TrimSpace(input.Title),
		SourceRefs:           sourceRefs,
		Body:                 input.Body,
		BodyFacts:            bodyFacts,
		FreshnessNote:        strings.TrimSpace(input.FreshnessNote),
		Mode:                 strings.TrimSpace(input.Mode),
		DisposableValidation: input.DisposableValidation,
	}
}

func normalizeOptionalCompileSynthesisMarkdownPath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return normalizeCompileSynthesisMarkdownPath(trimmed)
}

func trimGitLifecycleOptions(input GitLifecycleOptions) GitLifecycleOptions {
	paths := make([]string, 0, len(input.Paths))
	for _, rawPath := range input.Paths {
		clean := normalizeVaultRelativePath(rawPath)
		if clean != "" {
			paths = append(paths, clean)
		}
	}
	return GitLifecycleOptions{
		Mode:    strings.TrimSpace(input.Mode),
		Paths:   paths,
		Message: sanitizeGitLifecycleMessage(input.Message),
		Limit:   input.Limit,
	}
}

func trimWebSearchPlanOptions(input WebSearchPlanOptions) WebSearchPlanOptions {
	results := make([]WebSearchResultInput, 0, len(input.Results))
	for _, result := range input.Results {
		results = append(results, WebSearchResultInput{
			URL:          strings.TrimSpace(result.URL),
			Title:        strings.TrimSpace(result.Title),
			Snippet:      strings.TrimSpace(result.Snippet),
			SourceType:   strings.TrimSpace(result.SourceType),
			AccessStatus: strings.TrimSpace(result.AccessStatus),
		})
	}
	return WebSearchPlanOptions{
		Query:   strings.TrimSpace(input.Query),
		Results: results,
		Limit:   input.Limit,
	}
}

func trimArtifactPlanOptions(input ArtifactPlanOptions) ArtifactPlanOptions {
	tags := make([]string, 0, len(input.Tags))
	for _, tag := range input.Tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	fields := make(map[string]string, len(input.Fields))
	for key, value := range input.Fields {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		fields[trimmedKey] = strings.TrimSpace(value)
	}
	if len(fields) == 0 {
		fields = nil
	}
	return ArtifactPlanOptions{
		Content:        input.Content,
		LocalPath:      strings.TrimSpace(input.LocalPath),
		SourceURL:      strings.TrimSpace(input.SourceURL),
		SourceType:     strings.TrimSpace(input.SourceType),
		ArtifactKind:   strings.TrimSpace(input.ArtifactKind),
		TextExtraction: strings.TrimSpace(input.TextExtraction),
		OCRProvider:    strings.TrimSpace(input.OCRProvider),
		Path:           normalizeVaultRelativePath(input.Path),
		Title:          strings.TrimSpace(input.Title),
		Body:           input.Body,
		Tags:           tags,
		Fields:         fields,
		DuplicateQuery: strings.TrimSpace(input.DuplicateQuery),
		PathPrefix:     normalizeVaultRelativePrefix(input.PathPrefix),
		Limit:          input.Limit,
	}
}

func validateWebSearchResultInput(input WebSearchResultInput) string {
	if _, rejection := validateRequiredRunnerHTTPURL(input.URL, "web_search.results.url"); rejection != "" {
		return rejection
	}
	if input.SourceType != "" && input.SourceType != "web" && input.SourceType != "pdf" {
		return "web_search.results.source_type must be web or pdf"
	}
	switch input.AccessStatus {
	case "", "public", "blocked", "authenticated", "private", "unknown":
		return ""
	default:
		return "web_search.results.access_status must be public, blocked, authenticated, private, or unknown"
	}
}

func validateArtifactPlanOptions(input ArtifactPlanOptions) string {
	if strings.TrimSpace(input.Content) == "" && strings.TrimSpace(input.Body) == "" && input.SourceURL == "" && input.LocalPath == "" {
		return "artifact.content, artifact.body, artifact.source_url, or artifact.local_path is required"
	}
	if input.LocalPath != "" {
		if strings.HasPrefix(input.LocalPath, "~") {
			return "artifact.local_path must be an explicit path, not home-relative"
		}
	}
	if _, rejection := validateOptionalRunnerHTTPURL(input.SourceURL, "artifact.source_url"); rejection != "" {
		return rejection
	}
	if input.SourceType != "" {
		switch input.SourceType {
		case "explicit_content", "public_url", "web", "pdf", "local_artifact":
		default:
			return "artifact.source_type must be explicit_content, public_url, web, pdf, or local_artifact"
		}
	}
	if input.TextExtraction != "" {
		switch input.TextExtraction {
		case "ocr_review":
			if input.LocalPath == "" {
				return "artifact.local_path is required for artifact.text_extraction ocr_review"
			}
		default:
			return "artifact.text_extraction must be ocr_review"
		}
	}
	if input.OCRProvider != "" && input.TextExtraction == "" {
		return "artifact.ocr_provider requires artifact.text_extraction ocr_review"
	}
	if input.OCRProvider != "" && input.OCRProvider != runclient.OCRModuleProviderTesseract {
		return "artifact.ocr_provider must be tesseract"
	}
	if input.ArtifactKind != "" {
		switch input.ArtifactKind {
		case "note", "invoice", "receipt", "legal_document", "transcript", "source_summary", "unknown":
		default:
			return "artifact.artifact_kind must be note, invoice, receipt, legal_document, transcript, source_summary, or unknown"
		}
	}
	if input.Path != "" {
		_, issue := domain.NormalizeVaultRelativePath(input.Path)
		if issue != domain.VaultPathOK {
			return "artifact.path must stay inside the vault root"
		}
		if path.Ext(input.Path) != ".md" {
			return "artifact.path must end with .md"
		}
	}
	if input.PathPrefix != "" {
		_, issue := domain.NormalizeVaultRelativePath(input.PathPrefix)
		if issue != domain.VaultPathOK {
			return "artifact.path_prefix must stay inside the vault root"
		}
	}
	return ""
}

func normalizeCompileSynthesisMode(raw string) string {
	mode := strings.TrimSpace(raw)
	if mode == "" {
		return "create_or_update"
	}
	return mode
}

func normalizeCompileSynthesisMarkdownPath(raw string) string {
	clean, issue := domain.NormalizeOptionalVaultRelativePath(raw)
	if issue != domain.VaultPathOK {
		return strings.TrimSpace(raw)
	}
	if path.Ext(clean) == "" {
		clean += ".md"
	}
	return clean
}

func validateCompileSynthesisInput(input CompileSynthesisInput) string {
	if input.Path == "" {
		return "synthesis.path is required"
	}
	_, issue := domain.NormalizeVaultRelativePath(input.Path)
	if issue != domain.VaultPathOK {
		return "synthesis.path must stay inside the vault root"
	}
	if !strings.HasPrefix(input.Path, "synthesis/") {
		return "synthesis.path must be under synthesis/"
	}
	if path.Ext(input.Path) != ".md" {
		return "synthesis.path must end with .md"
	}
	if input.Title == "" {
		return "synthesis.title is required"
	}
	if len(input.SourceRefs) == 0 {
		return "synthesis.source_refs is required"
	}
	for _, ref := range input.SourceRefs {
		if ref == "" {
			return "synthesis.source_refs entries must be non-empty"
		}
		if strings.ContainsAny(ref, ",\n\r\t") {
			return "synthesis.source_refs entries must be single vault-relative paths without separators"
		}
		_, issue := domain.NormalizeVaultRelativePath(ref)
		if issue != domain.VaultPathOK {
			return "synthesis.source_refs entries must stay inside the vault root"
		}
		if !strings.HasPrefix(ref, "sources/") {
			return "synthesis.source_refs entries must be under sources/"
		}
		if path.Ext(ref) != ".md" {
			return "synthesis.source_refs entries must end with .md"
		}
	}
	if strings.TrimSpace(input.Body) == "" && len(input.BodyFacts) == 0 {
		return "synthesis.body or synthesis.body_facts is required"
	}
	if input.Mode != "create_or_update" {
		return "synthesis.mode must be create_or_update"
	}
	return ""
}

func validateValidationSynthesisInput(input ValidationSynthesisInput) string {
	if !input.DisposableValidation {
		return "validation_synthesis.disposable_validation must be true for disposable validation writes"
	}
	synthesisInput := CompileSynthesisInput{
		Path:          firstNonEmpty(input.Path, "synthesis/routine-ux-validation.md"),
		Title:         firstNonEmpty(input.Title, "Routine UX Validation Synthesis"),
		SourceRefs:    input.SourceRefs,
		Body:          input.Body,
		BodyFacts:     input.BodyFacts,
		FreshnessNote: input.FreshnessNote,
		Mode:          "create_or_update",
	}
	if len(synthesisInput.SourceRefs) == 0 {
		synthesisInput.SourceRefs = []string{"sources/routine-ux-validation/source.md"}
	}
	if strings.TrimSpace(synthesisInput.Body) == "" && len(synthesisInput.BodyFacts) == 0 {
		synthesisInput.BodyFacts = []string{"validation synthesis default body fact"}
	}
	return validateCompileSynthesisInput(trimCompileSynthesisInput(synthesisInput))
}

func validateSourceURLInput(input SourceURLInput) string {
	parsed, rejection := validateRequiredRunnerHTTPURLSyntax(input.URL, "source.url")
	if rejection != "" {
		return rejection
	}
	mode := input.Mode
	if mode == "" {
		mode = "create"
	}
	if mode != "create" && mode != "update" && mode != "plan" && mode != "inspect" {
		return "source.mode must be create, update, plan, or inspect"
	}
	if mode == "plan" || mode == "inspect" {
		if rejection := validateRunnerPublicURLHost(parsed, "source.url"); rejection != "" {
			return rejection
		}
	}
	if input.Limit < 0 {
		return "source.limit must be non-negative"
	}
	sourceType := input.SourceType
	if sourceType != "" && sourceType != "pdf" && sourceType != "web" {
		return "source.source_type must be pdf or web"
	}
	if sourceType == "web" && input.AssetPathHint != "" {
		return "source.asset_path_hint is not supported for source_type web"
	}
	if mode == "create" || input.PathHint != "" {
		if rejection := validateSourcePathHint(input.PathHint); rejection != "" {
			return rejection
		}
	}
	requiresPDFAsset := sourceType == "pdf" || (sourceType == "" && strings.HasSuffix(strings.ToLower(parsed.Path), ".pdf"))
	if requiresPDFAsset && (mode == "create" || input.AssetPathHint != "") {
		if rejection := validateAssetPathHint(input.AssetPathHint); rejection != "" {
			return rejection
		}
	} else if input.AssetPathHint != "" {
		if rejection := validateAssetPathHint(input.AssetPathHint); rejection != "" {
			return rejection
		}
	}
	return ""
}

func validateVideoURLInput(input VideoURLInput) string {
	if _, rejection := validateRequiredRunnerHTTPURLSyntax(input.URL, "video.url"); rejection != "" {
		return rejection
	}
	mode := input.Mode
	if mode == "" {
		mode = "create"
	}
	if mode != "create" && mode != "update" {
		return "video.mode must be create or update"
	}
	if mode == "create" || input.PathHint != "" {
		if rejection := validateVideoPathHint(input.PathHint); rejection != "" {
			return rejection
		}
	}
	if input.AssetPathHint != "" {
		if rejection := validateVideoAssetPathHint(input.AssetPathHint); rejection != "" {
			return rejection
		}
	}
	if strings.TrimSpace(input.Transcript.Text) == "" {
		return "video.transcript.text is required; native video download, platform captions, local transcription, transcript APIs, and Gemini extraction are not supported by ingest_video_url v1"
	}
	policy := input.Transcript.Policy
	if policy == "" {
		policy = "supplied"
	}
	if policy != "supplied" && policy != "local_first" {
		return "video.transcript.policy must be supplied or local_first when transcript.text is provided"
	}
	return ""
}

func validateDocumentInput(input DocumentInput) string {
	if strings.TrimSpace(input.Path) == "" {
		return "document.path is required"
	}
	if strings.TrimSpace(input.Title) == "" {
		return "document.title is required"
	}
	if strings.TrimSpace(input.Body) == "" {
		return "document.body is required"
	}
	if rejection := validateDocumentFrontmatter(input.Body); rejection != "" {
		return rejection
	}
	return ""
}

func validateDocumentFrontmatter(body string) string {
	frontmatter := parseDocumentFrontmatter(body)
	modality := strings.TrimSpace(frontmatter["modality"])
	if modality == "" || strings.EqualFold(modality, "markdown") {
		return ""
	}
	return "document.body frontmatter modality must be markdown for runner-created Markdown documents"
}

func parseDocumentFrontmatter(body string) map[string]string {
	lines := strings.Split(body, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return map[string]string{}
	}
	frontmatter := map[string]string{}
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			return frontmatter
		}
		key, value, ok := strings.Cut(lines[idx], ":")
		if !ok {
			continue
		}
		frontmatter[strings.TrimSpace(strings.ToLower(key))] = cleanDocumentFrontmatterValue(value)
	}
	return map[string]string{}
}

func cleanDocumentFrontmatterValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 2 {
		return trimmed
	}
	if strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`) {
		unquoted, err := strconv.Unquote(trimmed)
		if err == nil {
			return strings.TrimSpace(unquoted)
		}
	}
	if strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'") {
		return strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}
	return trimmed
}
