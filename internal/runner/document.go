package runner

import (
	"context"
	"errors"
	"fmt"
	"path"
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
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == DocumentTaskActionValidate {
		return DocumentTaskResult{Summary: "valid"}, nil
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
		document, err := client.GetDocument(ctx, normalized.DocID)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		return DocumentTaskResult{
			Document: &converted,
			Summary:  fmt.Sprintf("returned document %s", converted.DocID),
		}, nil
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
		DocumentTaskActionReplaceSection,
		DocumentTaskActionMoveDocument,
		DocumentTaskActionRenameDocument,
		DocumentTaskActionPromoteCandidate:
		return true
	case DocumentTaskActionIngestSourceURL:
		return normalized.Source.Mode != "plan"
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
		document, err := client.AppendDocument(ctx, normalized.DocID, domain.AppendDocumentInput{Content: normalized.Content})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		return DocumentTaskResult{
			Document: &converted,
			Summary:  fmt.Sprintf("appended document %s", converted.DocID),
		}, nil
	case DocumentTaskActionReplaceSection:
		document, err := client.ReplaceSection(ctx, normalized.DocID, domain.ReplaceSectionInput{
			Heading: normalized.Heading,
			Content: normalized.Content,
		})
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		return DocumentTaskResult{
			Document: &converted,
			Summary:  fmt.Sprintf("replaced section in document %s", converted.DocID),
		}, nil
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
	default:
		return DocumentTaskResult{}, fmt.Errorf("unsupported mutating document task action %q", normalized.Action)
	}
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
	DocID               string
	Content             string
	Heading             string
	List                DocumentListOptions
}

func normalizeDocumentTaskRequest(request DocumentTaskRequest) (normalizedDocumentTaskRequest, string) {
	action := strings.TrimSpace(request.Action)
	if action == "" {
		action = DocumentTaskActionValidate
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
		DocID:               strings.TrimSpace(request.DocID),
		Content:             request.Content,
		Heading:             strings.TrimSpace(request.Heading),
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
	); rejection != "" {
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
		if normalized.DocID == "" {
			return normalizedDocumentTaskRequest{}, "doc_id is required"
		}
		return normalized, ""
	case DocumentTaskActionAppend:
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if normalized.DocID == "" {
			return normalizedDocumentTaskRequest{}, "doc_id is required"
		}
		if strings.TrimSpace(normalized.Content) == "" {
			return normalizedDocumentTaskRequest{}, "content is required"
		}
		return normalized, ""
	case DocumentTaskActionReplaceSection:
		if rejection := rejectDocumentAutonomyWrite(normalized); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		if normalized.DocID == "" {
			return normalizedDocumentTaskRequest{}, "doc_id is required"
		}
		if normalized.Heading == "" {
			return normalizedDocumentTaskRequest{}, "heading is required"
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
	if mode != "create" && mode != "update" && mode != "plan" {
		return "source.mode must be create, update, or plan"
	}
	if mode == "plan" {
		if rejection := validateRunnerPublicURLHost(parsed, "source.url"); rejection != "" {
			return rejection
		}
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
