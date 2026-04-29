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

	if isMutatingDocumentAction(normalized.Action) {
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
	case DocumentTaskActionList:
		documents, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix:    normalized.List.PathPrefix,
			MetadataKey:   normalized.List.MetadataKey,
			MetadataValue: normalized.List.MetadataValue,
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

func isMutatingDocumentAction(action string) bool {
	switch action {
	case DocumentTaskActionCreate,
		DocumentTaskActionIngestSourceURL,
		DocumentTaskActionIngestVideoURL,
		DocumentTaskActionAppend,
		DocumentTaskActionReplaceSection:
		return true
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
	Action   string
	Document DocumentInput
	Source   SourceURLInput
	Video    VideoURLInput
	DocID    string
	Content  string
	Heading  string
	List     DocumentListOptions
}

func normalizeDocumentTaskRequest(request DocumentTaskRequest) (normalizedDocumentTaskRequest, string) {
	action := strings.TrimSpace(request.Action)
	if action == "" {
		action = DocumentTaskActionValidate
	}
	normalized := normalizedDocumentTaskRequest{
		Action:   action,
		Document: request.Document,
		Source:   trimSourceURLInput(request.Source),
		Video:    trimVideoURLInput(request.Video),
		DocID:    strings.TrimSpace(request.DocID),
		Content:  request.Content,
		Heading:  strings.TrimSpace(request.Heading),
		List:     request.List,
	}

	if request.List.Limit < 0 {
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
	case DocumentTaskActionList:
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
	default:
		return normalizedDocumentTaskRequest{}, fmt.Sprintf("unsupported document task action %q", action)
	}
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
	if mode != "create" && mode != "update" {
		return "source.mode must be create or update"
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
