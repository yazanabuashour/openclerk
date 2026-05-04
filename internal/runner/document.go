package runner

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
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
		plan, err := runArtifactCandidatePlan(ctx, client, normalized.Artifact)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		return DocumentTaskResult{
			ArtifactPlan: &plan,
			Summary:      artifactCandidatePlanSummary(plan),
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
		DocumentTaskActionAppend,
		DocumentTaskActionReplaceSection:
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
	default:
		return DocumentTaskResult{}, fmt.Errorf("unsupported mutating document task action %q", normalized.Action)
	}
}

type normalizedDocumentTaskRequest struct {
	Action       string
	Document     DocumentInput
	Source       SourceURLInput
	Video        VideoURLInput
	Synthesis    CompileSynthesisInput
	GitLifecycle GitLifecycleOptions
	WebSearch    WebSearchPlanOptions
	Artifact     ArtifactPlanOptions
	DocID        string
	Content      string
	Heading      string
	List         DocumentListOptions
}

func normalizeDocumentTaskRequest(request DocumentTaskRequest) (normalizedDocumentTaskRequest, string) {
	action := strings.TrimSpace(request.Action)
	if action == "" {
		action = DocumentTaskActionValidate
	}
	normalized := normalizedDocumentTaskRequest{
		Action:       action,
		Document:     request.Document,
		Source:       trimSourceURLInput(request.Source),
		Video:        trimVideoURLInput(request.Video),
		Synthesis:    trimCompileSynthesisInput(compileSynthesisInputFromRequest(request)),
		GitLifecycle: trimGitLifecycleOptions(request.GitLifecycle),
		WebSearch:    trimWebSearchPlanOptions(request.WebSearch),
		Artifact:     trimArtifactPlanOptions(request.Artifact),
		DocID:        strings.TrimSpace(request.DocID),
		Content:      request.Content,
		Heading:      strings.TrimSpace(request.Heading),
		List:         request.List,
	}

	if request.List.Limit < 0 {
		return normalizedDocumentTaskRequest{}, "limit must be greater than or equal to 0"
	}
	if request.GitLifecycle.Limit < 0 {
		return normalizedDocumentTaskRequest{}, "limit must be greater than or equal to 0"
	}
	if request.WebSearch.Limit < 0 {
		return normalizedDocumentTaskRequest{}, "limit must be greater than or equal to 0"
	}
	if request.Artifact.Limit < 0 {
		return normalizedDocumentTaskRequest{}, "limit must be greater than or equal to 0"
	}

	switch action {
	case DocumentTaskActionValidate:
		if request.Document != (DocumentInput{}) {
			return normalized, validateDocumentInput(request.Document)
		}
		return normalized, ""
	case DocumentTaskActionCreate:
		if rejection := validateDocumentInput(request.Document); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionIngestSourceURL:
		if rejection := validateSourceURLInput(normalized.Source); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionIngestVideoURL:
		if rejection := validateVideoURLInput(normalized.Video); rejection != "" {
			return normalizedDocumentTaskRequest{}, rejection
		}
		return normalized, ""
	case DocumentTaskActionCompileSynthesis:
		if rejection := validateCompileSynthesisInput(normalized.Synthesis); rejection != "" {
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
		if normalized.DocID == "" {
			return normalizedDocumentTaskRequest{}, "doc_id is required"
		}
		if strings.TrimSpace(normalized.Content) == "" {
			return normalizedDocumentTaskRequest{}, "content is required"
		}
		return normalized, ""
	case DocumentTaskActionReplaceSection:
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
		for _, path := range normalized.GitLifecycle.Paths {
			if path == "" || filepath.IsAbs(path) || strings.HasPrefix(path, "/") || path == ".." || strings.HasPrefix(path, "../") {
				return normalizedDocumentTaskRequest{}, "git_lifecycle.paths entries must stay inside the vault root"
			}
			if isUnsafeGitLifecyclePath(path) {
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
	default:
		return normalizedDocumentTaskRequest{}, fmt.Sprintf("unsupported document task action %q", action)
	}
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

func normalizeDocumentListTagFilter(list *DocumentListOptions) string {
	return normalizeTagFilter("list", list.Tag, list.tagProvided, &list.MetadataKey, &list.MetadataValue, &list.Tag)
}

func trimSourceURLInput(input SourceURLInput) SourceURLInput {
	return SourceURLInput{
		URL:           strings.TrimSpace(input.URL),
		PathHint:      strings.TrimSpace(input.PathHint),
		AssetPathHint: strings.TrimSpace(input.AssetPathHint),
		Title:         strings.TrimSpace(input.Title),
		Mode:          strings.TrimSpace(input.Mode),
		SourceType:    strings.TrimSpace(input.SourceType),
	}
}

func trimVideoURLInput(input VideoURLInput) VideoURLInput {
	return VideoURLInput{
		URL:           strings.TrimSpace(input.URL),
		PathHint:      strings.TrimSpace(input.PathHint),
		AssetPathHint: strings.TrimSpace(input.AssetPathHint),
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
		SourceURL:      strings.TrimSpace(input.SourceURL),
		SourceType:     strings.TrimSpace(input.SourceType),
		ArtifactKind:   strings.TrimSpace(input.ArtifactKind),
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

func normalizeVaultRelativePath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "/") {
		return trimmed
	}
	clean := path.Clean(filepath.ToSlash(trimmed))
	if clean == "." {
		return ""
	}
	return clean
}

func normalizeVaultRelativePrefix(raw string) string {
	trimmed := strings.TrimSpace(filepath.ToSlash(raw))
	if trimmed == "" || filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "/") {
		return trimmed
	}
	trailingSlash := strings.HasSuffix(trimmed, "/")
	clean := path.Clean(trimmed)
	if clean == "." {
		return ""
	}
	if trailingSlash && clean != "." && clean != ".." && !strings.HasSuffix(clean, "/") {
		clean += "/"
	}
	return clean
}

func validateWebSearchResultInput(input WebSearchResultInput) string {
	if input.URL == "" {
		return "web_search.results.url is required"
	}
	parsed, err := url.Parse(input.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "web_search.results.url must be a valid http or https URL"
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "web_search.results.url must use http or https"
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
	if strings.TrimSpace(input.Content) == "" && strings.TrimSpace(input.Body) == "" && input.SourceURL == "" {
		return "artifact.content, artifact.body, or artifact.source_url is required"
	}
	if input.SourceURL != "" {
		parsed, err := url.Parse(input.SourceURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return "artifact.source_url must be a valid http or https URL"
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return "artifact.source_url must use http or https"
		}
	}
	if input.SourceType != "" {
		switch input.SourceType {
		case "explicit_content", "public_url", "web", "pdf":
		default:
			return "artifact.source_type must be explicit_content, public_url, web, or pdf"
		}
	}
	if input.ArtifactKind != "" {
		switch input.ArtifactKind {
		case "note", "invoice", "receipt", "legal_document", "transcript", "source_summary", "unknown":
		default:
			return "artifact.artifact_kind must be note, invoice, receipt, legal_document, transcript, source_summary, or unknown"
		}
	}
	if input.Path != "" {
		if filepath.IsAbs(input.Path) || strings.HasPrefix(input.Path, "/") || input.Path == "." || input.Path == ".." || strings.HasPrefix(input.Path, "../") {
			return "artifact.path must stay inside the vault root"
		}
		if path.Ext(input.Path) != ".md" {
			return "artifact.path must end with .md"
		}
	}
	if input.PathPrefix != "" {
		if filepath.IsAbs(input.PathPrefix) || strings.HasPrefix(input.PathPrefix, "/") || input.PathPrefix == "." || input.PathPrefix == ".." || strings.HasPrefix(input.PathPrefix, "../") {
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
	if raw == "" || filepath.IsAbs(raw) || strings.HasPrefix(raw, "/") {
		return raw
	}
	clean := path.Clean(filepath.ToSlash(raw))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return clean
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
	if filepath.IsAbs(input.Path) || strings.HasPrefix(input.Path, "/") || input.Path == "." || input.Path == ".." || strings.HasPrefix(input.Path, "../") {
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
		if filepath.IsAbs(ref) || strings.HasPrefix(ref, "/") || ref == "." || ref == ".." || strings.HasPrefix(ref, "../") {
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

func validateSourceURLInput(input SourceURLInput) string {
	if input.URL == "" {
		return "source.url is required"
	}
	parsed, err := url.Parse(input.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "source.url must be a valid http or https URL"
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "source.url must use http or https"
	}
	mode := input.Mode
	if mode == "" {
		mode = "create"
	}
	if mode != "create" && mode != "update" && mode != "plan" {
		return "source.mode must be create, update, or plan"
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
	if input.URL == "" {
		return "video.url is required"
	}
	parsed, err := url.Parse(input.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "video.url must be a valid http or https URL"
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "video.url must use http or https"
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

func validateSourcePathHint(pathHint string) string {
	if strings.TrimSpace(pathHint) == "" {
		return "source.path_hint is required"
	}
	if filepath.IsAbs(pathHint) || strings.HasPrefix(strings.TrimSpace(pathHint), "/") {
		return "source.path_hint must be relative to the vault root"
	}
	clean := path.Clean(filepath.ToSlash(strings.TrimSpace(pathHint)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "source.path_hint must stay inside the vault root"
	}
	if !strings.HasPrefix(clean, "sources/") || path.Ext(clean) != ".md" {
		return "source.path_hint must be a vault-relative sources/*.md path"
	}
	return ""
}

func validateVideoPathHint(pathHint string) string {
	if strings.TrimSpace(pathHint) == "" {
		return "video.path_hint is required"
	}
	if filepath.IsAbs(pathHint) || strings.HasPrefix(strings.TrimSpace(pathHint), "/") {
		return "video.path_hint must be relative to the vault root"
	}
	clean := path.Clean(filepath.ToSlash(strings.TrimSpace(pathHint)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "video.path_hint must stay inside the vault root"
	}
	if !strings.HasPrefix(clean, "sources/") || path.Ext(clean) != ".md" {
		return "video.path_hint must be a vault-relative sources/*.md path"
	}
	return ""
}

func validateAssetPathHint(pathHint string) string {
	if strings.TrimSpace(pathHint) == "" {
		return "source.asset_path_hint is required"
	}
	if filepath.IsAbs(pathHint) || strings.HasPrefix(strings.TrimSpace(pathHint), "/") {
		return "source.asset_path_hint must be relative to the vault root"
	}
	clean := path.Clean(filepath.ToSlash(strings.TrimSpace(pathHint)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "source.asset_path_hint must stay inside the vault root"
	}
	if !strings.HasPrefix(clean, "assets/") || path.Ext(clean) != ".pdf" {
		return "source.asset_path_hint must be a vault-relative assets/**/*.pdf path"
	}
	return ""
}

func validateVideoAssetPathHint(pathHint string) string {
	if filepath.IsAbs(pathHint) || strings.HasPrefix(strings.TrimSpace(pathHint), "/") {
		return "video.asset_path_hint must be relative to the vault root"
	}
	clean := path.Clean(filepath.ToSlash(strings.TrimSpace(pathHint)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "video.asset_path_hint must stay inside the vault root"
	}
	if !strings.HasPrefix(clean, "assets/") || path.Ext(clean) != ".json" {
		return "video.asset_path_hint must be a vault-relative assets/**/*.json path"
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
