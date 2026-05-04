package runner

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"rsc.io/pdf"
)

const (
	artifactPlanValidationBoundaries = "read-only artifact candidate planning; local file inspection only when artifact.local_path is explicitly supplied, limited to UTF-8 text/markdown/text-bearing PDF; no OCR, no opaque file parsing, no browser automation, no HTTP fetch, no durable document write, no direct vault inspection, no direct SQLite, no source-built runner, and no unsupported transport"
	artifactPlanAuthorityLimits      = "candidate path, title, tags, fields, and body preview are planning hints from explicit content or approved runner handoff context only; canonical markdown, citations, provenance, freshness, and projections become authority only after approved runner writes"
)

func runArtifactCandidatePlan(ctx context.Context, client *runclient.Client, options ArtifactPlanOptions) (ArtifactCandidatePlan, error) {
	localArtifact, extractedContent, err := inspectLocalArtifact(options.LocalPath)
	if err != nil {
		return ArtifactCandidatePlan{}, err
	}
	if strings.TrimSpace(options.Content) == "" && strings.TrimSpace(options.Body) == "" && extractedContent != "" {
		options.Content = extractedContent
		if options.SourceType == "" {
			options.SourceType = "local_artifact"
		}
	}
	content := strings.TrimSpace(options.Content)
	explicitBody := strings.TrimSpace(options.Body)
	artifactKind := inferArtifactKind(options.ArtifactKind, content, explicitBody)
	sourceType := artifactSourceType(options)
	sourceURL := normalizeArtifactSourceURL(options.SourceURL)

	title, titleReason := artifactTitle(options.Title, content, explicitBody, artifactKind, sourceURL)
	candidatePath, pathReason := artifactCandidatePath(options.Path, artifactKind, title, sourceURL)
	tags := artifactTags(options.Tags, artifactKind)
	metadataFields := artifactMetadataFields(options.Fields, tags, artifactKind, sourceType, sourceURL)
	bodyPreview := artifactBodyPreview(title, explicitBody, content, metadataFields)

	duplicateSearch, likelyDuplicate, err := artifactDuplicateEvidence(ctx, client, options, title, bodyPreview, candidatePath)
	if err != nil {
		return ArtifactCandidatePlan{}, err
	}
	existingSource, err := artifactExistingSource(ctx, client, sourceURL, options.SourceURL)
	if err != nil {
		return ArtifactCandidatePlan{}, err
	}

	duplicateStatus := artifactDuplicateStatus(duplicateSearch, likelyDuplicate, existingSource)
	confidence, confidenceReasons := artifactConfidence(options, artifactKind, bodyPreview, titleReason, pathReason, duplicateStatus)
	nextCreate := artifactNextCreateRequest(candidatePath, title, bodyPreview, confidence, likelyDuplicate, existingSource)
	nextIngest := artifactNextIngestRequest(sourceURL, sourceType, title, existingSource)

	plan := ArtifactCandidatePlan{
		ArtifactKind:            artifactKind,
		SourceType:              sourceType,
		SourceURL:               sourceURL,
		LocalArtifact:           localArtifact,
		CandidatePath:           candidatePath,
		CandidateTitle:          title,
		BodyPreview:             bodyPreview,
		Tags:                    tags,
		MetadataFields:          metadataFields,
		DuplicateSearch:         duplicateSearch,
		LikelyDuplicate:         likelyDuplicate,
		ExistingSource:          existingSource,
		DuplicateStatus:         duplicateStatus,
		Confidence:              confidence,
		ConfidenceReasons:       confidenceReasons,
		FetchStatus:             "planned_no_fetch",
		WriteStatus:             "planned_no_write",
		ApprovalBoundary:        "candidate planning, public read, public fetch, or duplicate inspection is not durable-write approval; approve create_document, ingest_source_url, or update-versus-new choice before mutating",
		ValidationBoundaries:    artifactPlanValidationBoundaries,
		AuthorityLimits:         artifactPlanAuthorityLimits,
		NextCreateRequest:       nextCreate,
		NextIngestSourceRequest: nextIngest,
	}
	plan.AgentHandoff = artifactCandidatePlanHandoff(plan)
	return plan, nil
}

func inspectLocalArtifact(localPath string) (*LocalArtifact, string, error) {
	if strings.TrimSpace(localPath) == "" {
		return nil, "", nil
	}
	info, err := os.Stat(localPath)
	if err != nil {
		return nil, "", domain.ValidationError("artifact.local_path is not readable", nil)
	}
	if info.IsDir() {
		return nil, "", domain.ValidationError("artifact.local_path must be a file", nil)
	}
	if !info.Mode().IsRegular() {
		return nil, "", domain.ValidationError("artifact.local_path must be a regular file", nil)
	}
	const maxLocalArtifactBytes = 10 * 1024 * 1024
	if info.Size() > maxLocalArtifactBytes {
		return nil, "", domain.ValidationError("artifact.local_path exceeds 10MB read-only planning limit", nil)
	}
	data, err := os.ReadFile(localPath)
	if err != nil {
		return nil, "", domain.ValidationError("artifact.local_path is not readable", nil)
	}
	sum := sha256.Sum256(data)
	local := &LocalArtifact{
		SourceRef: "user_supplied_local_artifact",
		FileName:  filepath.Base(localPath),
		SizeBytes: info.Size(),
		SHA256:    fmt.Sprintf("%x", sum[:]),
	}
	ext := strings.ToLower(filepath.Ext(localPath))
	switch ext {
	case ".md", ".markdown", ".txt", ".csv", ".tsv", ".log":
		if !utf8.Valid(data) {
			local.MIMEType = "application/octet-stream"
			local.Parser = "none"
			local.TextStatus = "unsupported_binary"
			local.Confidence = "none"
			local.UnsupportedReason = "local artifact is not UTF-8 text"
			return nil, "", domain.ValidationError("artifact.local_path must be UTF-8 text, markdown, or text-bearing PDF", nil)
		}
		local.MIMEType = localArtifactMIME(ext, data)
		local.Parser = "utf8_text"
		local.TextStatus = "extracted"
		local.Confidence = "high"
		return local, string(data), nil
	case ".pdf":
		text, pages, err := extractLocalArtifactPDF(localPath)
		if err != nil {
			return nil, "", err
		}
		local.MIMEType = "application/pdf"
		local.Parser = "rsc.io/pdf text extraction"
		local.TextStatus = "extracted"
		local.Confidence = "medium"
		local.PageCount = pages
		return local, text, nil
	default:
		local.MIMEType = localArtifactMIME(ext, data)
		local.Parser = "none"
		local.TextStatus = "unsupported"
		local.Confidence = "none"
		local.UnsupportedReason = "only UTF-8 text, markdown, and text-bearing PDF local artifacts are supported"
		return nil, "", domain.ValidationError("artifact.local_path must be text, markdown, or text-bearing PDF; OCR/image parsing is unsupported", nil)
	}
}

func localArtifactMIME(ext string, data []byte) string {
	if value := mime.TypeByExtension(ext); value != "" {
		return strings.Split(value, ";")[0]
	}
	return http.DetectContentType(data)
}

func extractLocalArtifactPDF(localPath string) (string, int, error) {
	reader, err := pdf.Open(localPath)
	if err != nil {
		return "", 0, domain.ValidationError("artifact.local_path PDF is not readable", nil)
	}
	pages := reader.NumPage()
	if pages == 0 {
		return "", 0, domain.ValidationError("artifact.local_path PDF has no pages", nil)
	}
	var builder strings.Builder
	for pageNum := 1; pageNum <= pages; pageNum++ {
		texts := append([]pdf.Text(nil), reader.Page(pageNum).Content().Text...)
		sort.SliceStable(texts, func(i, j int) bool {
			if texts[i].Y == texts[j].Y {
				return texts[i].X < texts[j].X
			}
			return texts[i].Y > texts[j].Y
		})
		var previous *pdf.Text
		for idx := range texts {
			text := texts[idx]
			if text.S == "" {
				continue
			}
			if previous != nil {
				if text.Y < previous.Y-previous.FontSize*0.5 {
					builder.WriteByte('\n')
				} else if gap := text.X - (previous.X + previous.W); gap > max(previous.FontSize*0.2, 1) {
					builder.WriteByte(' ')
				}
			}
			builder.WriteString(text.S)
			previous = &texts[idx]
		}
	}
	extracted := strings.TrimSpace(builder.String())
	if extracted == "" {
		return "", pages, domain.ValidationError("artifact.local_path PDF has no extractable text; OCR is unsupported", nil)
	}
	return extracted + "\n", pages, nil
}

func inferArtifactKind(explicit string, content string, body string) string {
	if explicit != "" {
		return explicit
	}
	text := strings.ToLower(content + "\n" + body)
	switch {
	case strings.Contains(text, "invoice") || strings.Contains(text, "amount due"):
		return "invoice"
	case strings.Contains(text, "receipt") || strings.Contains(text, "total paid"):
		return "receipt"
	case strings.Contains(text, "agreement") || strings.Contains(text, "contract") || strings.Contains(text, "legal"):
		return "legal_document"
	case strings.Contains(text, "transcript") || strings.Contains(text, "speaker:"):
		return "transcript"
	case strings.TrimSpace(text) != "":
		return "note"
	default:
		return "unknown"
	}
}

func artifactSourceType(options ArtifactPlanOptions) string {
	if options.SourceURL != "" {
		if options.SourceType == "web" || options.SourceType == "pdf" {
			return options.SourceType
		}
		input := SourceURLInput{URL: options.SourceURL}
		return sourcePlacementType(input)
	}
	if options.SourceType == "web" || options.SourceType == "pdf" || options.SourceType == "explicit_content" || options.SourceType == "public_url" || options.SourceType == "local_artifact" {
		return options.SourceType
	}
	return "explicit_content"
}

func normalizeArtifactSourceURL(raw string) string {
	if raw == "" {
		return ""
	}
	return normalizeSourcePlacementURL(raw)
}

func artifactTitle(explicit string, content string, body string, artifactKind string, sourceURL string) (string, string) {
	if explicit != "" {
		return explicit, "explicit_title"
	}
	for _, text := range []string{body, content} {
		if heading := firstMarkdownHeading(text); heading != "" {
			return heading, "heading_title"
		}
	}
	if sourceURL != "" {
		slug := sourcePlacementSlug(SourceURLInput{URL: sourceURL}, sourceURL)
		return titleFromSlug(slug), "source_url_title"
	}
	if title := conciseContentTitle(content); title != "" {
		return title, "content_title"
	}
	return titleFromSlug(artifactKind + " artifact"), "fallback_title"
}

func firstMarkdownHeading(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}
	return ""
}

func conciseContentTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(strings.Trim(line, "#-*` "))
		if line == "" || strings.Contains(line, ":") && len(line) < 24 {
			continue
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		if len(words) > 6 {
			words = words[:6]
		}
		return titleFromSlug(slugifyPlacementLabel(strings.Join(words, " ")))
	}
	return ""
}

func titleFromSlug(slug string) string {
	words := strings.Fields(strings.ReplaceAll(slug, "-", " "))
	for i, word := range words {
		if word == "" {
			continue
		}
		runes := []rune(word)
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func artifactCandidatePath(explicit string, artifactKind string, title string, sourceURL string) (string, string) {
	if explicit != "" {
		return explicit, "explicit_path"
	}
	slug := slugifyPlacementLabel(title)
	if sourceURL != "" && slug == "source" {
		slug = sourcePlacementSlug(SourceURLInput{URL: sourceURL, Title: title}, sourceURL)
	}
	prefix := "notes/candidates/"
	switch artifactKind {
	case "invoice":
		prefix = "artifacts/invoices/"
	case "receipt":
		prefix = "artifacts/receipts/"
	case "legal_document":
		prefix = "artifacts/legal/"
	case "transcript":
		prefix = "artifacts/transcripts/"
	case "source_summary":
		prefix = "sources/candidates/"
	}
	return prefix + slug + ".md", "generated_path"
}

func artifactTags(explicit []string, artifactKind string) []string {
	tags := []string{}
	for _, tag := range explicit {
		tags = appendUniqueString(tags, tag)
	}
	tags = appendUniqueString(tags, "artifact-intake")
	if artifactKind != "" && artifactKind != "unknown" {
		tags = appendUniqueString(tags, strings.ReplaceAll(artifactKind, "_", "-"))
	}
	return tags
}

func artifactMetadataFields(explicit map[string]string, tags []string, artifactKind string, sourceType string, sourceURL string) map[string]string {
	fields := map[string]string{
		"type":          artifactDocumentType(artifactKind),
		"artifact_kind": artifactKind,
		"source_type":   sourceType,
	}
	if sourceURL != "" {
		fields["source_url"] = sourceURL
	}
	if len(tags) > 0 {
		fields["tag"] = tags[0]
	}
	for key, value := range explicit {
		fields[key] = value
	}
	return fields
}

func artifactDocumentType(artifactKind string) string {
	switch artifactKind {
	case "note":
		return "note"
	case "source_summary":
		return "source"
	case "unknown":
		return "artifact-candidate"
	default:
		return "artifact"
	}
}

func artifactBodyPreview(title string, explicitBody string, content string, fields map[string]string) string {
	if explicitBody != "" {
		return ensureTrailingNewline(explicitBody)
	}
	if strings.TrimSpace(content) == "" {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("---\n")
	for _, key := range orderedArtifactFieldKeys(fields) {
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(artifactFrontmatterValue(fields[key]))
		builder.WriteString("\n")
	}
	builder.WriteString("---\n")
	builder.WriteString("# ")
	builder.WriteString(title)
	builder.WriteString("\n\n")
	builder.WriteString(strings.TrimSpace(content))
	builder.WriteString("\n")
	return ensureTrailingNewline(builder.String())
}

func orderedArtifactFieldKeys(fields map[string]string) []string {
	keys := make([]string, 0, len(fields))
	seen := map[string]bool{}
	for _, key := range []string{"type", "artifact_kind", "source_type", "source_url", "tag"} {
		if _, ok := fields[key]; ok {
			keys = append(keys, key)
			seen[key] = true
		}
	}
	extra := make([]string, 0, len(fields))
	for key := range fields {
		if !seen[key] {
			extra = append(extra, key)
		}
	}
	sort.Strings(extra)
	return append(keys, extra...)
}

func artifactFrontmatterValue(value string) string {
	return strconv.Quote(strings.TrimSpace(value))
}

func ensureTrailingNewline(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return value + "\n"
}

func artifactDuplicateEvidence(ctx context.Context, client *runclient.Client, options ArtifactPlanOptions, title string, bodyPreview string, candidatePath string) (*SearchResult, *SearchHit, error) {
	query := options.DuplicateQuery
	if query == "" {
		query = artifactDuplicateQuery(title, bodyPreview)
	}
	if query == "" {
		return nil, nil, nil
	}
	limit := options.Limit
	if limit == 0 {
		limit = 5
	}
	if limit > 10 {
		limit = 10
	}
	pathPrefix := options.PathPrefix
	if pathPrefix == "" && candidatePath != "" {
		pathPrefix = path.Dir(candidatePath)
		if pathPrefix != "." {
			pathPrefix += "/"
		}
	}
	result, err := client.Search(ctx, domain.SearchQuery{
		Text:       query,
		PathPrefix: pathPrefix,
		Limit:      limit,
	})
	if err != nil {
		return nil, nil, err
	}
	converted := toSearchResult(result)
	var likely *SearchHit
	if len(converted.Hits) > 0 {
		hit := converted.Hits[0]
		likely = &hit
	}
	return &converted, likely, nil
}

func artifactDuplicateQuery(title string, bodyPreview string) string {
	text := strings.TrimSpace(title)
	body := strings.TrimSpace(stripFrontmatterAndHeading(bodyPreview))
	if body != "" {
		words := strings.Fields(body)
		if len(words) > 24 {
			words = words[:24]
		}
		text = strings.TrimSpace(text + " " + strings.Join(words, " "))
	}
	return text
}

func stripFrontmatterAndHeading(body string) string {
	lines := strings.Split(body, "\n")
	start := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for idx := 1; idx < len(lines); idx++ {
			if strings.TrimSpace(lines[idx]) == "---" {
				start = idx + 1
				break
			}
		}
	}
	kept := make([]string, 0, len(lines)-start)
	for _, line := range lines[start:] {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func artifactExistingSource(ctx context.Context, client *runclient.Client, sourceURL string, rawSourceURL string) (*DocumentSummary, error) {
	if sourceURL == "" {
		return nil, nil
	}
	return sourcePlacementExistingSource(ctx, client, sourceURL, rawSourceURL)
}

func artifactDuplicateStatus(search *SearchResult, likely *SearchHit, existingSource *DocumentSummary) string {
	if existingSource != nil {
		return "existing_source_url_found_no_write"
	}
	if likely != nil {
		return "likely_duplicate_candidate_no_write"
	}
	if search != nil {
		return "no_duplicate_found"
	}
	return "no_duplicate_search_performed"
}

func artifactConfidence(options ArtifactPlanOptions, artifactKind string, bodyPreview string, titleReason string, pathReason string, duplicateStatus string) (string, []string) {
	reasons := []string{titleReason, pathReason}
	if duplicateStatus != "" {
		reasons = append(reasons, duplicateStatus)
	}
	if bodyPreview == "" {
		return "low", append(reasons, "no_explicit_body_or_content")
	}
	if artifactKind == "unknown" {
		return "low", append(reasons, "unknown_artifact_kind")
	}
	if options.Title != "" && options.Path != "" && options.Body != "" {
		return "high", append(reasons, "explicit_path_title_body")
	}
	return "medium", append(reasons, "generated_candidate_from_explicit_content")
}

func artifactNextCreateRequest(candidatePath string, title string, body string, confidence string, likelyDuplicate *SearchHit, existingSource *DocumentSummary) string {
	if candidatePath == "" || title == "" || body == "" || confidence == "low" || likelyDuplicate != nil || existingSource != nil {
		return ""
	}
	payload := map[string]any{
		"action": DocumentTaskActionCreate,
		"document": map[string]string{
			"path":  candidatePath,
			"title": title,
			"body":  body,
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func artifactNextIngestRequest(sourceURL string, sourceType string, title string, existingSource *DocumentSummary) string {
	if sourceURL == "" {
		return ""
	}
	if existingSource != nil {
		return webSearchNextIngestRequest(sourceURL, sourceType, "", true, "public_candidate_requires_ingest_source_url_approval")
	}
	slug := sourcePlacementSlug(SourceURLInput{URL: sourceURL, Title: title, SourceType: sourceType}, sourceURL)
	return webSearchNextIngestRequest(sourceURL, sourceType, slug, false, "public_candidate_requires_ingest_source_url_approval")
}

func artifactCandidatePlanHandoff(plan ArtifactCandidatePlan) *AgentHandoff {
	evidence := []string{
		"artifact_kind=" + plan.ArtifactKind,
		"candidate_path=" + plan.CandidatePath,
		"candidate_title=" + plan.CandidateTitle,
		"duplicate_status=" + plan.DuplicateStatus,
		"confidence=" + plan.Confidence,
		"fetch_status=" + plan.FetchStatus,
		"write_status=" + plan.WriteStatus,
	}
	if plan.SourceURL != "" {
		evidence = append(evidence, "source_url="+plan.SourceURL)
	}
	if plan.LikelyDuplicate != nil {
		evidence = append(evidence, "likely_duplicate_doc_id="+plan.LikelyDuplicate.DocID)
	}
	if plan.ExistingSource != nil {
		evidence = append(evidence, "existing_source="+plan.ExistingSource.Path)
	}
	followUp := "after approval, call create_document for the candidate or ingest_source_url for a public source URL; if duplicate_status is not no_duplicate_found, ask update-versus-new before writing"
	if plan.NextCreateRequest == "" && plan.NextIngestSourceRequest == "" {
		followUp = "ask for missing content, artifact type, or update-versus-new approval before any durable write"
	}
	return &AgentHandoff{
		AnswerSummary: fmt.Sprintf(
			"artifact_candidate_plan proposed %s at %s with %s confidence; no fetch or write occurred",
			plan.ArtifactKind,
			plan.CandidatePath,
			plan.Confidence,
		),
		Evidence:                    evidence,
		ValidationBoundaries:        plan.ValidationBoundaries,
		AuthorityLimits:             plan.AuthorityLimits,
		FollowUpPrimitiveInspection: followUp,
	}
}

func artifactCandidatePlanSummary(plan ArtifactCandidatePlan) string {
	return fmt.Sprintf("planned %s artifact candidate with %s confidence", plan.ArtifactKind, plan.Confidence)
}
