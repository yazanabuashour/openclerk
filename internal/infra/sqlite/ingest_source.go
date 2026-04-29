package sqlite

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"rsc.io/pdf"
	"sort"
	"strings"
	"time"
)

type pdfExtraction struct {
	Metadata domain.SourcePDFMetadata
	Text     string
	Pages    int
}

func (s *Store) IngestSourceURL(ctx context.Context, input domain.SourceURLInput) (domain.SourceIngestionResult, error) {
	sourceURL, err := normalizeSourceURL(input.URL)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	mode, err := normalizeSourceURLMode(input.Mode)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	switch mode {
	case sourceURLModeCreate:
		return s.createSourceURL(ctx, input, sourceURL)
	case sourceURLModeUpdate:
		return s.updateSourceURL(ctx, input, sourceURL)
	default:
		return domain.SourceIngestionResult{}, domain.ValidationError("source mode must be create or update", map[string]any{"mode": input.Mode})
	}
}

func (s *Store) createSourceURL(ctx context.Context, input domain.SourceURLInput, sourceURL string) (domain.SourceIngestionResult, error) {
	sourcePath, err := normalizeSourceDocumentPath(input.PathHint)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	assetPath, err := normalizeSourceAssetPath(input.AssetPathHint)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	if exists, err := s.sourceURLExists(ctx, sourceURL); err != nil {
		return domain.SourceIngestionResult{}, err
	} else if exists {
		return domain.SourceIngestionResult{}, domain.AlreadyExistsError("source URL", sourceURL)
	}
	sourceAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(sourcePath))
	if _, err := osStat(sourceAbsPath); err == nil {
		return domain.SourceIngestionResult{}, domain.AlreadyExistsError("document path", sourcePath)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return domain.SourceIngestionResult{}, domain.InternalError("stat source document path", err)
	}
	assetAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(assetPath))
	if _, err := osStat(assetAbsPath); err == nil {
		return domain.SourceIngestionResult{}, domain.AlreadyExistsError("asset path", assetPath)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return domain.SourceIngestionResult{}, domain.InternalError("stat source asset path", err)
	}

	pdfBytes, mimeType, err := downloadSourcePDF(ctx, sourceURL)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	if err := ensureDir(filepath.Dir(assetAbsPath)); err != nil {
		return domain.SourceIngestionResult{}, domain.InternalError("create source asset directory", err)
	}
	if err := osWriteBytes(assetAbsPath, pdfBytes); err != nil {
		return domain.SourceIngestionResult{}, domain.InternalError("write source asset", err)
	}
	assetWritten := true
	defer func() {
		if assetWritten {
			if _, err := osStat(sourceAbsPath); errors.Is(err, fs.ErrNotExist) {
				_ = osRemove(assetAbsPath)
			}
		}
	}()

	extracted, err := extractPDF(assetAbsPath)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	capturedAt := s.now().UTC()
	sha := sha256.Sum256(pdfBytes)
	shaHex := hex.EncodeToString(sha[:])
	title := resolvedSourceTitle(input.Title, extracted.Metadata.Title, sourcePath)
	body := buildSourceNoteBody(sourceURL, sourcePath, assetPath, title, shaHex, int64(len(pdfBytes)), mimeType, capturedAt, extracted)
	document, err := s.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  sourcePath,
		Title: title,
		Body:  body,
	})
	if err != nil {
		if _, statErr := osStat(sourceAbsPath); statErr == nil {
			assetWritten = false
		}
		return domain.SourceIngestionResult{}, err
	}
	assetWritten = false
	citations, err := s.sourceDocumentCitations(ctx, document)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	if len(citations) == 0 {
		return domain.SourceIngestionResult{}, domain.InternalError("validate source ingestion citations", errors.New("created source has no indexed citations"))
	}
	if err := validateIngestedSource(document, sourceURL, sourcePath, assetPath, assetAbsPath); err != nil {
		return domain.SourceIngestionResult{}, err
	}
	return domain.SourceIngestionResult{
		DocID:       document.DocID,
		SourcePath:  sourcePath,
		AssetPath:   assetPath,
		DerivedPath: sourcePath,
		Citations:   citations,
		SHA256:      shaHex,
		SizeBytes:   int64(len(pdfBytes)),
		MIMEType:    mimeType,
		PageCount:   extracted.Pages,
		CapturedAt:  capturedAt,
		PDFMetadata: extracted.Metadata,
	}, nil
}

func (s *Store) updateSourceURL(ctx context.Context, input domain.SourceURLInput, sourceURL string) (domain.SourceIngestionResult, error) {
	document, err := s.sourceDocumentByURL(ctx, sourceURL)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	sourcePath := document.Path
	assetPath := strings.TrimSpace(document.Metadata["asset_path"])
	if assetPath == "" {
		return domain.SourceIngestionResult{}, domain.InternalError("source document is missing asset path", fmt.Errorf("source_url %q", sourceURL))
	}
	if strings.TrimSpace(input.PathHint) != "" {
		requestPath, err := normalizeSourceDocumentPath(input.PathHint)
		if err != nil {
			return domain.SourceIngestionResult{}, err
		}
		if requestPath != sourcePath {
			return domain.SourceIngestionResult{}, domain.ConflictError("source path hint does not match existing source", map[string]any{"source_url": sourceURL, "path_hint": requestPath, "existing_path": sourcePath})
		}
	}
	if strings.TrimSpace(input.AssetPathHint) != "" {
		requestAssetPath, err := normalizeSourceAssetPath(input.AssetPathHint)
		if err != nil {
			return domain.SourceIngestionResult{}, err
		}
		if requestAssetPath != assetPath {
			return domain.SourceIngestionResult{}, domain.ConflictError("source asset path hint does not match existing source", map[string]any{"source_url": sourceURL, "asset_path_hint": requestAssetPath, "existing_asset_path": assetPath})
		}
	}
	assetAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(assetPath))
	if _, err := osStat(assetAbsPath); err != nil {
		return domain.SourceIngestionResult{}, domain.InternalError("validate existing source asset", err)
	}

	pdfBytes, mimeType, err := downloadSourcePDF(ctx, sourceURL)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	sha := sha256.Sum256(pdfBytes)
	shaHex := hex.EncodeToString(sha[:])
	if strings.TrimSpace(document.Metadata["sha256"]) == shaHex {
		return s.sourceIngestionResultFromDocument(ctx, document)
	}

	tempAssetPath := assetAbsPath + ".openclerk-update-" + hashID("asset", sourceURL, shaHex)
	if err := osWriteBytes(tempAssetPath, pdfBytes); err != nil {
		return domain.SourceIngestionResult{}, domain.InternalError("write source asset staging file", err)
	}
	defer func() {
		_ = osRemove(tempAssetPath)
	}()
	extracted, err := extractPDF(tempAssetPath)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}

	sourceAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(sourcePath))
	oldBody := document.Body
	oldAssetBytes, err := osReadFile(assetAbsPath)
	if err != nil {
		return domain.SourceIngestionResult{}, domain.InternalError("read existing source asset", err)
	}
	capturedAt := s.now().UTC()
	title := resolvedSourceTitle(input.Title, extracted.Metadata.Title, sourcePath)
	body := buildSourceNoteBody(sourceURL, sourcePath, assetPath, title, shaHex, int64(len(pdfBytes)), mimeType, capturedAt, extracted)
	if err := s.replaceSourceAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, pdfBytes, body, title); err != nil {
		return domain.SourceIngestionResult{}, err
	}
	updated, err := s.GetDocument(ctx, document.DocID)
	if err != nil {
		return domain.SourceIngestionResult{}, s.restoreSourceAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err)
	}
	citations, err := s.sourceDocumentCitations(ctx, updated)
	if err != nil {
		return domain.SourceIngestionResult{}, s.restoreSourceAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err)
	}
	if len(citations) == 0 {
		return domain.SourceIngestionResult{}, s.restoreSourceAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, domain.InternalError("validate source update citations", errors.New("updated source has no indexed citations")))
	}
	if err := validateIngestedSource(updated, sourceURL, sourcePath, assetPath, assetAbsPath); err != nil {
		return domain.SourceIngestionResult{}, s.restoreSourceAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err)
	}
	return domain.SourceIngestionResult{
		DocID:       updated.DocID,
		SourcePath:  sourcePath,
		AssetPath:   assetPath,
		DerivedPath: sourcePath,
		Citations:   citations,
		SHA256:      shaHex,
		SizeBytes:   int64(len(pdfBytes)),
		MIMEType:    mimeType,
		PageCount:   extracted.Pages,
		CapturedAt:  capturedAt,
		PDFMetadata: extracted.Metadata,
	}, nil
}

func (s *Store) sourceDocumentByURL(ctx context.Context, sourceURL string) (domain.Document, error) {
	var docID string
	err := s.db.QueryRowContext(ctx, `
SELECT doc_id
FROM document_metadata
WHERE key_name = 'source_url' AND value_text = ?
ORDER BY doc_id
LIMIT 1`, sourceURL).Scan(&docID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Document{}, domain.NotFoundError("source URL", sourceURL)
	}
	if err != nil {
		return domain.Document{}, domain.InternalError("query source URL document", err)
	}
	return s.GetDocument(ctx, docID)
}

func (s *Store) sourceURLExists(ctx context.Context, sourceURL string) (bool, error) {
	var found string
	err := s.db.QueryRowContext(ctx, `
SELECT doc_id
FROM document_metadata
WHERE key_name = 'source_url' AND value_text = ?
LIMIT 1`, sourceURL).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, domain.InternalError("query duplicate source URL", err)
	}
	return true, nil
}

func (s *Store) sourceIngestionResultFromDocument(ctx context.Context, document domain.Document) (domain.SourceIngestionResult, error) {
	sourceURL := strings.TrimSpace(document.Metadata["source_url"])
	assetPath := strings.TrimSpace(document.Metadata["asset_path"])
	derivedPath := strings.TrimSpace(document.Metadata["derived_path"])
	if derivedPath == "" {
		derivedPath = document.Path
	}
	if sourceURL == "" || assetPath == "" {
		return domain.SourceIngestionResult{}, domain.InternalError("source document is missing ingestion metadata", fmt.Errorf("path %q", document.Path))
	}
	assetAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(assetPath))
	if err := validateIngestedSource(document, sourceURL, document.Path, assetPath, assetAbsPath); err != nil {
		return domain.SourceIngestionResult{}, err
	}
	citations, err := s.sourceDocumentCitations(ctx, document)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	return domain.SourceIngestionResult{
		DocID:       document.DocID,
		SourcePath:  document.Path,
		AssetPath:   assetPath,
		DerivedPath: derivedPath,
		Citations:   citations,
		SHA256:      strings.TrimSpace(document.Metadata["sha256"]),
		SizeBytes:   parseInt64Metadata(document.Metadata["size_bytes"]),
		MIMEType:    strings.TrimSpace(document.Metadata["mime_type"]),
		PageCount:   parseIntMetadata(document.Metadata["page_count"]),
		CapturedAt:  parseTimeMetadata(document.Metadata["captured_at"]),
		PDFMetadata: domain.SourcePDFMetadata{
			Title:         strings.TrimSpace(document.Metadata["pdf_title"]),
			Author:        strings.TrimSpace(document.Metadata["pdf_author"]),
			PublishedDate: strings.TrimSpace(document.Metadata["pdf_published_date"]),
		},
	}, nil
}

func (s *Store) replaceSourceAssetAndNote(ctx context.Context, sourcePath string, sourceAbsPath string, assetAbsPath string, oldBody string, oldAssetBytes []byte, newAssetBytes []byte, newBody string, title string) error {
	if err := osWriteBytes(assetAbsPath, newAssetBytes); err != nil {
		return domain.InternalError("write source asset", err)
	}
	if err := osWriteFile(sourceAbsPath, newBody); err != nil {
		return s.restoreSourceAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, domain.InternalError("write source note", err))
	}
	if err := s.syncDocumentFromDisk(ctx, sourcePath, title); err != nil {
		return s.restoreSourceAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err)
	}
	return nil
}

func (s *Store) restoreSourceAssetAndNote(ctx context.Context, sourcePath string, sourceAbsPath string, assetAbsPath string, oldBody string, oldAssetBytes []byte, cause error) error {
	if restoreErr := osWriteBytes(assetAbsPath, oldAssetBytes); restoreErr != nil {
		return domain.InternalError("restore source asset after failed update", errors.Join(cause, restoreErr))
	}
	if restoreErr := osWriteFile(sourceAbsPath, oldBody); restoreErr != nil {
		return domain.InternalError("restore source note after failed update", errors.Join(cause, restoreErr))
	}
	if restoreErr := s.syncDocumentFromDisk(ctx, sourcePath, ""); restoreErr != nil {
		return domain.InternalError("restore indexed source after failed update", errors.Join(cause, restoreErr))
	}
	return cause
}

func downloadSourcePDF(ctx context.Context, sourceURL string) ([]byte, string, error) {
	if fixturePath, ok, err := resolveEvalSourceFixturePath(sourceURL); err != nil {
		return nil, "", err
	} else if ok {
		body, err := osReadFile(fixturePath)
		if err != nil {
			return nil, "", domain.InternalError("read eval source PDF fixture", err)
		}
		if len(body) > maxSourceDownloadBytes {
			return nil, "", domain.ValidationError("source PDF exceeds maximum supported size", map[string]any{"max_bytes": maxSourceDownloadBytes})
		}
		if !looksLikePDF(body) {
			return nil, "", domain.ValidationError("source URL did not return a PDF", nil)
		}
		return body, "application/pdf", nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, "", domain.ValidationError("source url must be fetchable", map[string]any{"url": sourceURL})
	}
	client := http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, "", domain.InternalError("download source PDF", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", domain.ValidationError("source URL returned a non-success status", map[string]any{"status": response.StatusCode})
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxSourceDownloadBytes+1))
	if err != nil {
		return nil, "", domain.InternalError("read source PDF response", err)
	}
	if len(body) > maxSourceDownloadBytes {
		return nil, "", domain.ValidationError("source PDF exceeds maximum supported size", map[string]any{"max_bytes": maxSourceDownloadBytes})
	}
	if !looksLikePDF(body) {
		return nil, "", domain.ValidationError("source URL did not return a PDF", nil)
	}
	mimeType := ""
	if parsed, _, err := mime.ParseMediaType(response.Header.Get("Content-Type")); err == nil && parsed == "application/pdf" {
		mimeType = parsed
	} else {
		mimeType = http.DetectContentType(body)
	}
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = "application/pdf"
	}
	return body, mimeType, nil
}

func resolveEvalSourceFixturePath(sourceURL string) (string, bool, error) {
	root := strings.TrimSpace(os.Getenv(evalSourceFixtureRootEnv))
	if root == "" {
		return "", false, nil
	}
	parsed, err := url.Parse(sourceURL)
	if err != nil || parsed.Hostname() != evalSourceFixtureHost {
		return "", false, nil
	}
	clean := path.Clean("/" + strings.TrimPrefix(parsed.EscapedPath(), "/"))
	if clean == "/" || strings.Contains(clean, "%2f") || strings.Contains(clean, "%2F") {
		return "", false, domain.ValidationError("eval source fixture path is invalid", map[string]any{"path": parsed.Path})
	}
	rel := strings.TrimPrefix(clean, "/")
	target := filepath.Join(root, filepath.FromSlash(rel))
	rootClean, err := filepath.Abs(root)
	if err != nil {
		return "", false, domain.InternalError("resolve eval source fixture root", err)
	}
	targetClean, err := filepath.Abs(target)
	if err != nil {
		return "", false, domain.InternalError("resolve eval source fixture path", err)
	}
	if targetClean != rootClean && !strings.HasPrefix(targetClean, rootClean+string(os.PathSeparator)) {
		return "", false, domain.ValidationError("eval source fixture path escapes root", map[string]any{"path": parsed.Path})
	}
	return targetClean, true, nil
}

func looksLikePDF(body []byte) bool {
	prefixLen := min(len(body), 1024)
	return bytes.Contains(body[:prefixLen], []byte("%PDF-"))
}

func extractPDF(assetPath string) (pdfExtraction, error) {
	reader, err := pdf.Open(assetPath)
	if err != nil {
		return pdfExtraction{}, domain.ValidationError("source asset is not a readable PDF", nil)
	}
	pages := reader.NumPage()
	if pages < 1 {
		return pdfExtraction{}, domain.ValidationError("source PDF must contain at least one page", nil)
	}
	info := reader.Trailer().Key("Info")
	metadata := domain.SourcePDFMetadata{
		Title:         pdfText(info.Key("Title")),
		Author:        pdfText(info.Key("Author")),
		PublishedDate: firstNonEmpty(pdfText(info.Key("CreationDate")), pdfText(info.Key("ModDate"))),
	}
	pageTexts := make([]string, 0, pages)
	for pageNum := 1; pageNum <= pages; pageNum++ {
		texts := append([]pdf.Text(nil), reader.Page(pageNum).Content().Text...)
		sort.SliceStable(texts, func(i, j int) bool {
			if texts[i].Y == texts[j].Y {
				return texts[i].X < texts[j].X
			}
			return texts[i].Y > texts[j].Y
		})
		var pageText strings.Builder
		var previous *pdf.Text
		for idx := range texts {
			text := texts[idx]
			if text.S != "" {
				if previous != nil {
					if text.Y < previous.Y-previous.FontSize*0.5 {
						pageText.WriteString("\n")
					} else if gap := text.X - (previous.X + previous.W); gap > max(previous.FontSize*0.2, 1) {
						pageText.WriteString(" ")
					}
				}
				pageText.WriteString(text.S)
				previous = &texts[idx]
			}
		}
		if value := strings.TrimSpace(pageText.String()); value != "" {
			pageTexts = append(pageTexts, fmt.Sprintf("Page %d\n%s", pageNum, value))
		}
	}
	return pdfExtraction{
		Metadata: metadata,
		Text:     strings.TrimSpace(strings.Join(pageTexts, "\n\n")),
		Pages:    pages,
	}, nil
}

func pdfText(value pdf.Value) string {
	text := strings.TrimSpace(value.Text())
	if text != "" {
		return text
	}
	text = strings.TrimSpace(value.TextFromUTF16())
	if text != "" {
		return text
	}
	return strings.TrimSpace(value.RawString())
}

func resolvedSourceTitle(requestTitle string, pdfTitle string, sourcePath string) string {
	if title := strings.TrimSpace(requestTitle); title != "" {
		return title
	}
	if title := strings.TrimSpace(pdfTitle); title != "" {
		return title
	}
	return strings.TrimSuffix(path.Base(sourcePath), path.Ext(sourcePath))
}

func buildSourceNoteBody(sourceURL string, sourcePath string, assetPath string, title string, sha string, sizeBytes int64, mimeType string, capturedAt time.Time, extracted pdfExtraction) string {
	var body strings.Builder
	body.WriteString("---\n")
	body.WriteString("type: source\n")
	body.WriteString("source_type: pdf\n")
	body.WriteString("modality: markdown\n")
	body.WriteString("source_url: " + frontmatterScalar(sourceURL) + "\n")
	body.WriteString("asset_path: " + frontmatterScalar(assetPath) + "\n")
	body.WriteString("derived_path: " + frontmatterScalar(sourcePath) + "\n")
	body.WriteString("sha256: " + frontmatterScalar(sha) + "\n")
	_, _ = fmt.Fprintf(&body, "size_bytes: %d\n", sizeBytes)
	body.WriteString("mime_type: " + frontmatterScalar(mimeType) + "\n")
	_, _ = fmt.Fprintf(&body, "page_count: %d\n", extracted.Pages)
	body.WriteString("captured_at: " + frontmatterScalar(capturedAt.Format(time.RFC3339Nano)) + "\n")
	if extracted.Metadata.Title != "" {
		body.WriteString("pdf_title: " + frontmatterScalar(extracted.Metadata.Title) + "\n")
	}
	if extracted.Metadata.Author != "" {
		body.WriteString("pdf_author: " + frontmatterScalar(extracted.Metadata.Author) + "\n")
	}
	if extracted.Metadata.PublishedDate != "" {
		body.WriteString("pdf_published_date: " + frontmatterScalar(extracted.Metadata.PublishedDate) + "\n")
	}
	body.WriteString("---\n")
	body.WriteString("# " + markdownLine(title) + "\n\n")
	body.WriteString("## Summary\n")
	body.WriteString("PDF source ingested from " + sourceURL + ".\n\n")
	body.WriteString("## Source Asset\n")
	body.WriteString("- Source URL: " + sourceURL + "\n")
	body.WriteString("- Asset path: " + assetPath + "\n")
	body.WriteString("- SHA256: " + sha + "\n")
	_, _ = fmt.Fprintf(&body, "- Size bytes: %d\n", sizeBytes)
	_, _ = fmt.Fprintf(&body, "- Page count: %d\n\n", extracted.Pages)
	if extracted.Metadata.Title != "" || extracted.Metadata.Author != "" || extracted.Metadata.PublishedDate != "" {
		body.WriteString("## PDF Metadata\n")
		if extracted.Metadata.Title != "" {
			body.WriteString("- Title: " + extracted.Metadata.Title + "\n")
		}
		if extracted.Metadata.Author != "" {
			body.WriteString("- Author: " + extracted.Metadata.Author + "\n")
		}
		if extracted.Metadata.PublishedDate != "" {
			body.WriteString("- Published date: " + extracted.Metadata.PublishedDate + "\n")
		}
		body.WriteString("\n")
	}
	body.WriteString("## Extracted Text\n")
	if extracted.Text != "" {
		body.WriteString(extracted.Text + "\n")
	} else {
		body.WriteString(sourceMetadataFallbackText(extracted.Metadata, title) + "\n")
	}
	return body.String()
}

func sourceMetadataFallbackText(metadata domain.SourcePDFMetadata, title string) string {
	parts := []string{"No extractable page text was found."}
	if strings.TrimSpace(title) != "" {
		parts = append(parts, "Title: "+title+".")
	}
	if metadata.Author != "" {
		parts = append(parts, "Author: "+metadata.Author+".")
	}
	if metadata.PublishedDate != "" {
		parts = append(parts, "Published date: "+metadata.PublishedDate+".")
	}
	return strings.Join(parts, " ")
}

func (s *Store) sourceDocumentCitations(ctx context.Context, document domain.Document) ([]domain.Citation, error) {
	chunksByDoc, err := s.loadChunksByDoc(ctx)
	if err != nil {
		return nil, err
	}
	return []domain.Citation{documentCitation(document, chunksByDoc[document.DocID])}, nil
}

func validateIngestedSource(document domain.Document, sourceURL string, sourcePath string, assetPath string, assetAbsPath string) error {
	if document.DocID == "" || document.Path != sourcePath {
		return domain.InternalError("validate ingested source document", fmt.Errorf("created document path %q does not match %q", document.Path, sourcePath))
	}
	requiredMetadata := map[string]string{
		"type":         "source",
		"source_type":  "pdf",
		"modality":     "markdown",
		"source_url":   sourceURL,
		"asset_path":   assetPath,
		"derived_path": sourcePath,
	}
	for key, want := range requiredMetadata {
		if got := strings.TrimSpace(document.Metadata[key]); got != want {
			return domain.InternalError("validate ingested source metadata", fmt.Errorf("metadata %s = %q, want %q", key, got, want))
		}
	}
	if _, err := osStat(assetAbsPath); err != nil {
		return domain.InternalError("validate ingested source asset", err)
	}
	return nil
}

func sourceProvenanceDetails(relPath string, frontmatter map[string]string) map[string]string {
	details := map[string]string{"path": relPath}
	for _, key := range []string{
		"source_type",
		"source_url",
		"asset_path",
		"derived_path",
		"sha256",
		"size_bytes",
		"mime_type",
		"page_count",
		"captured_at",
		"transcript_origin",
		"transcript_policy",
		"language",
		"transcript_sha256",
		"tool",
		"model",
	} {
		if value := strings.TrimSpace(frontmatter[key]); value != "" {
			details[key] = value
		}
	}
	return details
}

func sourceUpdateProvenanceDetails(relPath string, frontmatter map[string]string, previous map[string]string) map[string]string {
	details := sourceProvenanceDetails(relPath, frontmatter)
	for _, key := range []string{"sha256", "page_count", "captured_at", "transcript_sha256"} {
		if value := strings.TrimSpace(previous[key]); value != "" {
			details["previous_"+key] = value
		}
		if value := strings.TrimSpace(frontmatter[key]); value != "" {
			details["new_"+key] = value
		}
	}
	return details
}
