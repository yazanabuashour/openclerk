package runner

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

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

	client, err := runclient.Open(config)
	if err != nil {
		return DocumentTaskResult{}, err
	}
	defer func() {
		_ = client.Close()
	}()

	switch normalized.Action {
	case DocumentTaskActionCreate:
		document, err := client.CreateDocument(ctx, runclient.DocumentInput(normalized.Document))
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		return DocumentTaskResult{
			Document: &converted,
			Summary:  fmt.Sprintf("created document %s", converted.DocID),
		}, nil
	case DocumentTaskActionIngestSourceURL:
		ingestion, err := client.IngestSourceURL(ctx, runclient.SourceURLInput(normalized.Source))
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toSourceIngestionResult(ingestion)
		return DocumentTaskResult{
			Ingestion: &converted,
			Summary:   fmt.Sprintf("ingested source URL into %s", converted.SourcePath),
		}, nil
	case DocumentTaskActionList:
		documents, err := client.ListDocuments(ctx, runclient.DocumentListOptions(normalized.List))
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
	case DocumentTaskActionAppend:
		document, err := client.AppendDocument(ctx, normalized.DocID, normalized.Content)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		return DocumentTaskResult{
			Document: &converted,
			Summary:  fmt.Sprintf("appended document %s", converted.DocID),
		}, nil
	case DocumentTaskActionReplaceSection:
		document, err := client.ReplaceSection(ctx, normalized.DocID, normalized.Heading, normalized.Content)
		if err != nil {
			return DocumentTaskResult{}, err
		}
		converted := toDocument(document)
		return DocumentTaskResult{
			Document: &converted,
			Summary:  fmt.Sprintf("replaced section in document %s", converted.DocID),
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

type normalizedDocumentTaskRequest struct {
	Action   string
	Document DocumentInput
	Source   SourceURLInput
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
	if mode == "create" || input.PathHint != "" {
		if rejection := validateSourcePathHint(input.PathHint); rejection != "" {
			return rejection
		}
	}
	if mode == "create" || input.AssetPathHint != "" {
		if rejection := validateAssetPathHint(input.AssetPathHint); rejection != "" {
			return rejection
		}
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
