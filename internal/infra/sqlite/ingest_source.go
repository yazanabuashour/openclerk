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
	"html"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

type webExtraction struct {
	Title string
	Text  string
}

type sourceDownload struct {
	Body       []byte
	MIMEType   string
	SourceType string
}

type normalizedSourceURLRequest struct {
	URL           string
	Mode          string
	RequestedType string
}

type noteMutationLabels struct {
	WriteAsset     string
	WriteNote      string
	RestoreAsset   string
	RestoreNote    string
	RestoreIndexed string
}

var sourceHTTPClient = defaultSourceHTTPClient

func defaultSourceHTTPClient(checkRedirect func(*http.Request, []*http.Request) error) *http.Client {
	return &http.Client{Timeout: 30 * time.Second, CheckRedirect: checkRedirect}
}

func (s *Store) IngestSourceURL(ctx context.Context, input domain.SourceURLInput) (domain.SourceIngestionResult, error) {
	request, err := normalizeSourceURLRequest(input)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	switch request.Mode {
	case sourceURLModeCreate:
		return s.createSourceURL(ctx, input, request.URL, request.RequestedType)
	case sourceURLModeUpdate:
		return s.updateSourceURL(ctx, input, request.URL, request.RequestedType)
	default:
		return domain.SourceIngestionResult{}, domain.ValidationError("source mode must be create or update", map[string]any{"mode": input.Mode})
	}
}

func normalizeSourceURLRequest(input domain.SourceURLInput) (normalizedSourceURLRequest, error) {
	sourceURL, err := normalizeSourceURL(input.URL)
	if err != nil {
		return normalizedSourceURLRequest{}, err
	}
	mode, err := normalizeSourceURLMode(input.Mode)
	if err != nil {
		return normalizedSourceURLRequest{}, err
	}
	requestedType, err := normalizeSourceType(input.SourceType)
	if err != nil {
		return normalizedSourceURLRequest{}, err
	}
	return normalizedSourceURLRequest{
		URL:           sourceURL,
		Mode:          mode,
		RequestedType: requestedType,
	}, nil
}

func (s *Store) createSourceURL(ctx context.Context, input domain.SourceURLInput, sourceURL string, requestedType string) (domain.SourceIngestionResult, error) {
	sourcePath, err := normalizeSourceDocumentPath(input.PathHint)
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

	download, err := downloadSource(ctx, sourceURL, requestedType)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	switch download.SourceType {
	case sourceTypePDF:
		return s.createPDFSourceURL(ctx, input, sourceURL, sourcePath, sourceAbsPath, download)
	case sourceTypeWeb:
		return s.createWebSourceURL(ctx, input, sourceURL, sourcePath, sourceAbsPath, download)
	default:
		return domain.SourceIngestionResult{}, domain.ValidationError("source URL returned an unsupported content type", map[string]any{"mime_type": download.MIMEType})
	}
}

func (s *Store) createPDFSourceURL(ctx context.Context, input domain.SourceURLInput, sourceURL string, sourcePath string, sourceAbsPath string, download sourceDownload) (domain.SourceIngestionResult, error) {
	assetPath, err := normalizeSourceAssetPath(input.AssetPathHint)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	assetAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(assetPath))
	if _, err := osStat(assetAbsPath); err == nil {
		return domain.SourceIngestionResult{}, domain.AlreadyExistsError("asset path", assetPath)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return domain.SourceIngestionResult{}, domain.InternalError("stat source asset path", err)
	}

	if err := ensureDir(filepath.Dir(assetAbsPath)); err != nil {
		return domain.SourceIngestionResult{}, domain.InternalError("create source asset directory", err)
	}
	if err := osWriteBytes(assetAbsPath, download.Body); err != nil {
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
	sha := sha256.Sum256(download.Body)
	shaHex := hex.EncodeToString(sha[:])
	title := resolvedSourceTitle(input.Title, extracted.Metadata.Title, sourcePath)
	body := buildPDFSourceNoteBody(sourceURL, sourcePath, assetPath, title, shaHex, int64(len(download.Body)), download.MIMEType, capturedAt, extracted)
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
		SourceURL:   sourceURL,
		SourceType:  sourceTypePDF,
		AssetPath:   assetPath,
		DerivedPath: sourcePath,
		Citations:   citations,
		SHA256:      shaHex,
		SizeBytes:   int64(len(download.Body)),
		MIMEType:    download.MIMEType,
		PageCount:   extracted.Pages,
		CapturedAt:  capturedAt,
		PDFMetadata: extracted.Metadata,
	}, nil
}

func (s *Store) createWebSourceURL(ctx context.Context, input domain.SourceURLInput, sourceURL string, sourcePath string, _ string, download sourceDownload) (domain.SourceIngestionResult, error) {
	if strings.TrimSpace(input.AssetPathHint) != "" {
		return domain.SourceIngestionResult{}, domain.ValidationError("source.asset_path_hint is not supported for web source ingestion", nil)
	}
	extracted, err := extractWebPage(download.Body)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	capturedAt := s.now().UTC()
	sha := sha256.Sum256(download.Body)
	shaHex := hex.EncodeToString(sha[:])
	title := resolvedSourceTitle(input.Title, extracted.Title, sourcePath)
	body := buildWebSourceNoteBody(sourceURL, sourcePath, title, shaHex, int64(len(download.Body)), download.MIMEType, capturedAt, extracted)
	document, err := s.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  sourcePath,
		Title: title,
		Body:  body,
	})
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	citations, err := s.sourceDocumentCitations(ctx, document)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	if len(citations) == 0 {
		return domain.SourceIngestionResult{}, domain.InternalError("validate web source ingestion citations", errors.New("created source has no indexed citations"))
	}
	if err := validateIngestedSource(document, sourceURL, sourcePath, "", ""); err != nil {
		return domain.SourceIngestionResult{}, err
	}
	return domain.SourceIngestionResult{
		DocID:       document.DocID,
		SourcePath:  sourcePath,
		SourceURL:   sourceURL,
		SourceType:  sourceTypeWeb,
		DerivedPath: sourcePath,
		Citations:   citations,
		SHA256:      shaHex,
		SizeBytes:   int64(len(download.Body)),
		MIMEType:    download.MIMEType,
		CapturedAt:  capturedAt,
	}, nil
}

func (s *Store) updateSourceURL(ctx context.Context, input domain.SourceURLInput, sourceURL string, requestedType string) (domain.SourceIngestionResult, error) {
	document, err := s.sourceDocumentByURL(ctx, sourceURL)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	storedType := strings.TrimSpace(document.Metadata["source_type"])
	if storedType == "" {
		storedType = sourceTypePDF
	}
	if requestedType != "" && requestedType != storedType {
		return domain.SourceIngestionResult{}, domain.ConflictError("source_type does not match existing source", map[string]any{"source_url": sourceURL, "source_type": requestedType, "existing_source_type": storedType})
	}
	switch storedType {
	case sourceTypePDF:
		return s.updatePDFSourceURL(ctx, input, sourceURL, document)
	case sourceTypeWeb:
		return s.updateWebSourceURL(ctx, input, sourceURL, document)
	default:
		return domain.SourceIngestionResult{}, domain.ConflictError("source URL belongs to an unsupported source type", map[string]any{"source_url": sourceURL, "source_type": storedType, "path": document.Path})
	}
}

func (s *Store) updatePDFSourceURL(ctx context.Context, input domain.SourceURLInput, sourceURL string, document domain.Document) (domain.SourceIngestionResult, error) {
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

	download, err := downloadSource(ctx, sourceURL, sourceTypePDF)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	sha := sha256.Sum256(download.Body)
	shaHex := hex.EncodeToString(sha[:])
	if strings.TrimSpace(document.Metadata["sha256"]) == shaHex {
		result, err := s.sourceIngestionResultFromDocument(ctx, document)
		if err != nil {
			return domain.SourceIngestionResult{}, err
		}
		return s.withSourceURLUpdateImpact(ctx, result, sourceURL, shaHex, shaHex, false)
	}

	tempAssetPath := assetAbsPath + ".openclerk-update-" + hashID("asset", sourceURL, shaHex)
	if err := osWriteBytes(tempAssetPath, download.Body); err != nil {
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
	body := buildPDFSourceNoteBody(sourceURL, sourcePath, assetPath, title, shaHex, int64(len(download.Body)), download.MIMEType, capturedAt, extracted)
	if err := s.replaceSourceAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, download.Body, body, title); err != nil {
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
	result := domain.SourceIngestionResult{
		DocID:       updated.DocID,
		SourcePath:  sourcePath,
		SourceURL:   sourceURL,
		SourceType:  sourceTypePDF,
		AssetPath:   assetPath,
		DerivedPath: sourcePath,
		Citations:   citations,
		SHA256:      shaHex,
		SizeBytes:   int64(len(download.Body)),
		MIMEType:    download.MIMEType,
		PageCount:   extracted.Pages,
		CapturedAt:  capturedAt,
		PDFMetadata: extracted.Metadata,
	}
	return s.withSourceURLUpdateImpact(ctx, result, sourceURL, strings.TrimSpace(document.Metadata["sha256"]), shaHex, true)
}

func (s *Store) updateWebSourceURL(ctx context.Context, input domain.SourceURLInput, sourceURL string, document domain.Document) (domain.SourceIngestionResult, error) {
	sourcePath := document.Path
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
		return domain.SourceIngestionResult{}, domain.ConflictError("web source does not have an asset path", map[string]any{"source_url": sourceURL, "asset_path_hint": input.AssetPathHint, "existing_path": sourcePath})
	}
	download, err := downloadSource(ctx, sourceURL, sourceTypeWeb)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	sha := sha256.Sum256(download.Body)
	shaHex := hex.EncodeToString(sha[:])
	if strings.TrimSpace(document.Metadata["sha256"]) == shaHex {
		result, err := s.sourceIngestionResultFromDocument(ctx, document)
		if err != nil {
			return domain.SourceIngestionResult{}, err
		}
		return s.withSourceURLUpdateImpact(ctx, result, sourceURL, shaHex, shaHex, false)
	}
	extracted, err := extractWebPage(download.Body)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	sourceAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(sourcePath))
	oldBody := document.Body
	capturedAt := s.now().UTC()
	title := resolvedSourceTitle(input.Title, extracted.Title, sourcePath)
	body := buildWebSourceNoteBody(sourceURL, sourcePath, title, shaHex, int64(len(download.Body)), download.MIMEType, capturedAt, extracted)
	if err := s.replaceSourceNote(ctx, sourcePath, sourceAbsPath, oldBody, body, title); err != nil {
		return domain.SourceIngestionResult{}, err
	}
	updated, err := s.GetDocument(ctx, document.DocID)
	if err != nil {
		return domain.SourceIngestionResult{}, s.restoreSourceNote(ctx, sourcePath, sourceAbsPath, oldBody, err)
	}
	citations, err := s.sourceDocumentCitations(ctx, updated)
	if err != nil {
		return domain.SourceIngestionResult{}, s.restoreSourceNote(ctx, sourcePath, sourceAbsPath, oldBody, err)
	}
	if len(citations) == 0 {
		return domain.SourceIngestionResult{}, s.restoreSourceNote(ctx, sourcePath, sourceAbsPath, oldBody, domain.InternalError("validate web source update citations", errors.New("updated source has no indexed citations")))
	}
	if err := validateIngestedSource(updated, sourceURL, sourcePath, "", ""); err != nil {
		return domain.SourceIngestionResult{}, s.restoreSourceNote(ctx, sourcePath, sourceAbsPath, oldBody, err)
	}
	result := domain.SourceIngestionResult{
		DocID:       updated.DocID,
		SourcePath:  sourcePath,
		SourceURL:   sourceURL,
		SourceType:  sourceTypeWeb,
		DerivedPath: sourcePath,
		Citations:   citations,
		SHA256:      shaHex,
		SizeBytes:   int64(len(download.Body)),
		MIMEType:    download.MIMEType,
		CapturedAt:  capturedAt,
	}
	return s.withSourceURLUpdateImpact(ctx, result, sourceURL, strings.TrimSpace(document.Metadata["sha256"]), shaHex, true)
}

func (s *Store) withSourceURLUpdateImpact(ctx context.Context, result domain.SourceIngestionResult, sourceURL string, previousSHA string, newSHA string, changed bool) (domain.SourceIngestionResult, error) {
	result.UpdateStatus = "no_op"
	if changed {
		result.UpdateStatus = "changed"
	}
	result.NormalizedSourceURL = sourceURL
	result.SourceDocID = result.DocID
	result.PreviousSHA256 = previousSHA
	result.NewSHA256 = newSHA
	result.Changed = changed
	result.DuplicateStatus = "existing_source_matched_no_duplicate_created"
	result.SynthesisRepaired = false

	if !changed {
		result.StaleDependents = []domain.SourceStaleDependent{}
		result.ProjectionRefs = []domain.SourceProjectionRef{}
		result.ProvenanceRefs = []domain.SourceProvenanceRef{}
		result.NoRepairWarning = "Source refresh detected no content change; no synthesis repair was performed."
		return result, nil
	}

	cursor := ""
	for {
		projections, err := s.ListProjectionStates(ctx, domain.ProjectionStateQuery{Projection: "synthesis", Limit: 100, Cursor: cursor})
		if err != nil {
			return domain.SourceIngestionResult{}, err
		}
		for _, projection := range projections.Projections {
			staleRefs := splitPathList(projection.Details["stale_source_refs"])
			if !strings.EqualFold(projection.Freshness, "stale") || !sourcePathListContains(staleRefs, result.SourcePath) {
				continue
			}
			dependentPath := strings.TrimSpace(projection.Details["synthesis_path"])
			result.StaleDependents = append(result.StaleDependents, domain.SourceStaleDependent{
				Path:            dependentPath,
				DocID:           projection.RefID,
				Projection:      projection.Projection,
				Freshness:       projection.Freshness,
				StaleSourceRefs: staleRefs,
			})
			result.ProjectionRefs = append(result.ProjectionRefs, domain.SourceProjectionRef{
				Projection: projection.Projection,
				RefKind:    projection.RefKind,
				RefID:      projection.RefID,
				Freshness:  projection.Freshness,
				SourceRef:  projection.SourceRef,
			})
		}
		if !projections.PageInfo.HasMore {
			break
		}
		cursor = projections.PageInfo.NextCursor
	}
	sort.Slice(result.StaleDependents, func(i, j int) bool {
		return result.StaleDependents[i].Path < result.StaleDependents[j].Path
	})
	sort.Slice(result.ProjectionRefs, func(i, j int) bool {
		if result.ProjectionRefs[i].Projection != result.ProjectionRefs[j].Projection {
			return result.ProjectionRefs[i].Projection < result.ProjectionRefs[j].Projection
		}
		return result.ProjectionRefs[i].RefID < result.ProjectionRefs[j].RefID
	})

	provenanceRefs, err := s.sourceURLUpdateProvenanceRefs(ctx, result.DocID, previousSHA, newSHA, result.ProjectionRefs)
	if err != nil {
		return domain.SourceIngestionResult{}, err
	}
	result.ProvenanceRefs = provenanceRefs
	if len(result.StaleDependents) == 0 {
		result.NoRepairWarning = "Source refresh did not repair dependent synthesis automatically."
		return result, nil
	}
	paths := make([]string, 0, len(result.StaleDependents))
	for _, dependent := range result.StaleDependents {
		if dependent.Path != "" {
			paths = append(paths, dependent.Path)
		}
	}
	result.NoRepairWarning = "Source refresh did not repair dependent synthesis: " + strings.Join(paths, ", ")
	return result, nil
}

func (s *Store) sourceURLUpdateProvenanceRefs(ctx context.Context, sourceDocID string, previousSHA string, newSHA string, projections []domain.SourceProjectionRef) ([]domain.SourceProvenanceRef, error) {
	refs := []domain.SourceProvenanceRef{}
	sourceEvents, err := s.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "source", RefID: sourceDocID, Limit: 20})
	if err != nil {
		return nil, err
	}
	for _, event := range sourceEvents.Events {
		if event.EventType != "source_updated" ||
			event.Details["previous_sha256"] != previousSHA ||
			event.Details["new_sha256"] != newSHA {
			continue
		}
		refs = append(refs, sourceProvenanceRef(event))
		break
	}

	if len(projections) == 0 {
		return refs, nil
	}
	for _, projection := range projections {
		projectionEvents, err := s.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{RefKind: "projection", RefID: projection.Projection + ":" + projection.RefID, Limit: 20})
		if err != nil {
			return nil, err
		}
		for _, event := range projectionEvents.Events {
			if event.EventType != "projection_invalidated" {
				continue
			}
			refs = append(refs, sourceProvenanceRef(event))
			break
		}
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].EventType != refs[j].EventType {
			return refs[i].EventType < refs[j].EventType
		}
		return refs[i].RefID < refs[j].RefID
	})
	return refs, nil
}

func sourcePathListContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sourceProvenanceRef(event domain.ProvenanceEvent) domain.SourceProvenanceRef {
	return domain.SourceProvenanceRef{
		EventID:   event.EventID,
		EventType: event.EventType,
		RefKind:   event.RefKind,
		RefID:     event.RefID,
		SourceRef: event.SourceRef,
	}
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
	sourceType := strings.TrimSpace(document.Metadata["source_type"])
	if sourceType == "" {
		sourceType = sourceTypePDF
	}
	assetPath := strings.TrimSpace(document.Metadata["asset_path"])
	derivedPath := strings.TrimSpace(document.Metadata["derived_path"])
	if derivedPath == "" {
		derivedPath = document.Path
	}
	if sourceURL == "" || (sourceType == sourceTypePDF && assetPath == "") {
		return domain.SourceIngestionResult{}, domain.InternalError("source document is missing ingestion metadata", fmt.Errorf("path %q", document.Path))
	}
	assetAbsPath := ""
	if assetPath != "" {
		assetAbsPath = filepath.Join(s.vaultRoot, filepath.FromSlash(assetPath))
	}
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
		SourceURL:   sourceURL,
		SourceType:  sourceType,
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
	return s.replaceIngestedAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, newAssetBytes, newBody, title, noteMutationLabels{
		WriteAsset:     "write source asset",
		WriteNote:      "write source note",
		RestoreAsset:   "restore source asset after failed update",
		RestoreNote:    "restore source note after failed update",
		RestoreIndexed: "restore indexed source after failed update",
	})
}

func (s *Store) restoreSourceAssetAndNote(ctx context.Context, sourcePath string, sourceAbsPath string, assetAbsPath string, oldBody string, oldAssetBytes []byte, cause error) error {
	return s.restoreIngestedAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, cause, noteMutationLabels{
		RestoreAsset:   "restore source asset after failed update",
		RestoreNote:    "restore source note after failed update",
		RestoreIndexed: "restore indexed source after failed update",
	})
}

func (s *Store) replaceSourceNote(ctx context.Context, sourcePath string, sourceAbsPath string, oldBody string, newBody string, title string) error {
	return s.replaceIngestedNote(ctx, sourcePath, sourceAbsPath, oldBody, newBody, title, noteMutationLabels{
		WriteNote:      "write source note",
		RestoreNote:    "restore source note after failed update",
		RestoreIndexed: "restore indexed source after failed update",
	})
}

func (s *Store) restoreSourceNote(ctx context.Context, sourcePath string, sourceAbsPath string, oldBody string, cause error) error {
	return s.restoreIngestedNote(ctx, sourcePath, sourceAbsPath, oldBody, cause, noteMutationLabels{
		RestoreNote:    "restore source note after failed update",
		RestoreIndexed: "restore indexed source after failed update",
	})
}

func (s *Store) replaceIngestedAssetAndNote(ctx context.Context, sourcePath string, sourceAbsPath string, assetAbsPath string, oldBody string, oldAssetBytes []byte, newAssetBytes []byte, newBody string, title string, labels noteMutationLabels) error {
	if err := osWriteBytes(assetAbsPath, newAssetBytes); err != nil {
		return domain.InternalError(labels.WriteAsset, err)
	}
	if err := osWriteFile(sourceAbsPath, newBody); err != nil {
		return s.restoreIngestedAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, domain.InternalError(labels.WriteNote, err), labels)
	}
	if err := s.syncDocumentFromDisk(ctx, sourcePath, title); err != nil {
		return s.restoreIngestedAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err, labels)
	}
	return nil
}

func (s *Store) restoreIngestedAssetAndNote(ctx context.Context, sourcePath string, sourceAbsPath string, assetAbsPath string, oldBody string, oldAssetBytes []byte, cause error, labels noteMutationLabels) error {
	if restoreErr := osWriteBytes(assetAbsPath, oldAssetBytes); restoreErr != nil {
		return domain.InternalError(labels.RestoreAsset, errors.Join(cause, restoreErr))
	}
	if restoreErr := osWriteFile(sourceAbsPath, oldBody); restoreErr != nil {
		return domain.InternalError(labels.RestoreNote, errors.Join(cause, restoreErr))
	}
	if restoreErr := s.syncDocumentFromDisk(ctx, sourcePath, ""); restoreErr != nil {
		return domain.InternalError(labels.RestoreIndexed, errors.Join(cause, restoreErr))
	}
	return cause
}

func (s *Store) replaceIngestedNote(ctx context.Context, sourcePath string, sourceAbsPath string, oldBody string, newBody string, title string, labels noteMutationLabels) error {
	if err := osWriteFile(sourceAbsPath, newBody); err != nil {
		return domain.InternalError(labels.WriteNote, err)
	}
	if err := s.syncDocumentFromDisk(ctx, sourcePath, title); err != nil {
		return s.restoreIngestedNote(ctx, sourcePath, sourceAbsPath, oldBody, err, labels)
	}
	return nil
}

func (s *Store) restoreIngestedNote(ctx context.Context, sourcePath string, sourceAbsPath string, oldBody string, cause error, labels noteMutationLabels) error {
	if restoreErr := osWriteFile(sourceAbsPath, oldBody); restoreErr != nil {
		return domain.InternalError(labels.RestoreNote, errors.Join(cause, restoreErr))
	}
	if restoreErr := s.syncDocumentFromDisk(ctx, sourcePath, ""); restoreErr != nil {
		return domain.InternalError(labels.RestoreIndexed, errors.Join(cause, restoreErr))
	}
	return cause
}

func downloadSource(ctx context.Context, sourceURL string, requestedType string) (sourceDownload, error) {
	if fixturePath, ok, err := resolveEvalSourceFixturePath(sourceURL); err != nil {
		return sourceDownload{}, err
	} else if ok {
		body, err := osReadFile(fixturePath)
		if err != nil {
			return sourceDownload{}, domain.InternalError("read eval source fixture", err)
		}
		if len(body) > maxSourceDownloadBytes {
			return sourceDownload{}, domain.ValidationError("source URL exceeds maximum supported size", map[string]any{"max_bytes": maxSourceDownloadBytes})
		}
		kind, mimeType, err := classifySourceBody(body, "application/pdf", requestedType)
		if err != nil {
			return sourceDownload{}, err
		}
		return sourceDownload{Body: body, MIMEType: mimeType, SourceType: kind}, nil
	}
	parsedURL, err := validateSourceFetchURL(sourceURL)
	if err != nil {
		return sourceDownload{}, err
	}
	if shouldValidatePublicSourceHost(parsedURL, requestedType) {
		if err := validatePublicSourceHost(ctx, parsedURL); err != nil {
			return sourceDownload{}, err
		}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return sourceDownload{}, domain.ValidationError("source url must be fetchable", map[string]any{"url": sourceURL})
	}
	var redirectValidationErr error
	client := sourceHTTPClient(func(request *http.Request, _ []*http.Request) error {
		if requestedType != sourceTypeWeb {
			return nil
		}
		if request == nil || request.URL == nil {
			redirectValidationErr = domain.ValidationError("source url must be fetchable", map[string]any{"url": sourceURL})
			return redirectValidationErr
		}
		redirectValidationErr = validateSourceFetchURLTarget(ctx, request.URL)
		return redirectValidationErr
	})
	if client == nil {
		client = defaultSourceHTTPClient(nil)
	}
	response, err := client.Do(request)
	if err != nil {
		if redirectValidationErr != nil {
			return sourceDownload{}, redirectValidationErr
		}
		return sourceDownload{}, domain.InternalError("download source URL", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusProxyAuthRequired {
			return sourceDownload{}, domain.ValidationError("source URL requires unsupported authenticated or restricted access", map[string]any{"status": response.StatusCode})
		}
		return sourceDownload{}, domain.ValidationError("source URL returned a non-success status", map[string]any{"status": response.StatusCode})
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxSourceDownloadBytes+1))
	if err != nil {
		return sourceDownload{}, domain.InternalError("read source URL response", err)
	}
	if len(body) > maxSourceDownloadBytes {
		return sourceDownload{}, domain.ValidationError("source URL exceeds maximum supported size", map[string]any{"max_bytes": maxSourceDownloadBytes})
	}
	mimeType := ""
	if parsed, _, err := mime.ParseMediaType(response.Header.Get("Content-Type")); err == nil {
		mimeType = parsed
	} else {
		mimeType = http.DetectContentType(body)
	}
	kind, mimeType, err := classifySourceBody(body, mimeType, requestedType)
	if err != nil {
		return sourceDownload{}, err
	}
	if kind == sourceTypeWeb {
		finalURL := response.Request.URL
		if finalURL == nil {
			finalURL = parsedURL
		}
		if err := validateSourceFetchURLTarget(ctx, finalURL); err != nil {
			return sourceDownload{}, err
		}
	}
	return sourceDownload{Body: body, MIMEType: mimeType, SourceType: kind}, nil
}

func validateSourceFetchURL(sourceURL string) (*url.URL, error) {
	parsed, err := url.Parse(sourceURL)
	if err != nil || parsed.Hostname() == "" {
		return nil, domain.ValidationError("source url must be fetchable", map[string]any{"url": sourceURL})
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, domain.ValidationError("source URL must use http or https", map[string]any{"scheme": parsed.Scheme})
	}
	return parsed, nil
}

func shouldValidatePublicSourceHost(parsed *url.URL, requestedType string) bool {
	if requestedType == sourceTypeWeb {
		return true
	}
	if requestedType == sourceTypePDF {
		return false
	}
	return !strings.EqualFold(path.Ext(parsed.EscapedPath()), ".pdf")
}

func validatePublicSourceHost(ctx context.Context, parsed *url.URL) error {
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return domain.ValidationError("source url must be fetchable", map[string]any{"url": parsed.String()})
	}
	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") || strings.HasSuffix(lowerHost, ".local") {
		return domain.ValidationError("source URL must be publicly fetchable", map[string]any{"host": host})
	}
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicSourceIP(ip) {
			return domain.ValidationError("source URL must be publicly fetchable", map[string]any{"host": host})
		}
		return nil
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return domain.ValidationError("source URL host must resolve publicly", map[string]any{"host": host})
	}
	for _, addr := range addrs {
		if !isPublicSourceIP(addr.IP) {
			return domain.ValidationError("source URL must be publicly fetchable", map[string]any{"host": host})
		}
	}
	return nil
}

func validateSourceFetchURLTarget(ctx context.Context, parsed *url.URL) error {
	if parsed == nil || parsed.Hostname() == "" {
		return domain.ValidationError("source url must be fetchable", map[string]any{"url": ""})
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return domain.ValidationError("source URL must use http or https", map[string]any{"scheme": parsed.Scheme})
	}
	return validatePublicSourceHost(ctx, parsed)
}

func isPublicSourceIP(ip net.IP) bool {
	return ip != nil &&
		!ip.IsUnspecified() &&
		!ip.IsLoopback() &&
		!ip.IsPrivate() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsMulticast()
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

func looksLikeHTML(body []byte) bool {
	prefixLen := min(len(body), 4096)
	prefix := strings.ToLower(string(body[:prefixLen]))
	return strings.Contains(prefix, "<!doctype html") || strings.Contains(prefix, "<html") || strings.Contains(prefix, "<head") || strings.Contains(prefix, "<body")
}

func classifySourceBody(body []byte, mimeType string, requestedType string) (string, string, error) {
	if looksLikePDF(body) {
		if requestedType == sourceTypeWeb {
			return "", "", domain.ValidationError("source URL returned a PDF, not a web page", nil)
		}
		if mimeType == "" || mimeType == "application/octet-stream" || mimeType == "text/plain" {
			mimeType = "application/pdf"
		}
		return sourceTypePDF, mimeType, nil
	}
	if mimeType == "text/html" || looksLikeHTML(body) {
		if requestedType == sourceTypePDF {
			return "", "", domain.ValidationError("source URL did not return a PDF", nil)
		}
		return sourceTypeWeb, "text/html", nil
	}
	if requestedType == sourceTypePDF {
		return "", "", domain.ValidationError("source URL did not return a PDF", nil)
	}
	if requestedType == sourceTypeWeb {
		return "", "", domain.ValidationError("source URL did not return an HTML web page", map[string]any{"mime_type": mimeType})
	}
	return "", "", domain.ValidationError("source URL returned an unsupported content type", map[string]any{"mime_type": mimeType})
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

func buildPDFSourceNoteBody(sourceURL string, sourcePath string, assetPath string, title string, sha string, sizeBytes int64, mimeType string, capturedAt time.Time, extracted pdfExtraction) string {
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

func buildWebSourceNoteBody(sourceURL string, sourcePath string, title string, sha string, sizeBytes int64, mimeType string, capturedAt time.Time, extracted webExtraction) string {
	var body strings.Builder
	body.WriteString("---\n")
	body.WriteString("type: source\n")
	body.WriteString("source_type: web\n")
	body.WriteString("modality: markdown\n")
	body.WriteString("source_url: " + frontmatterScalar(sourceURL) + "\n")
	body.WriteString("derived_path: " + frontmatterScalar(sourcePath) + "\n")
	body.WriteString("sha256: " + frontmatterScalar(sha) + "\n")
	_, _ = fmt.Fprintf(&body, "size_bytes: %d\n", sizeBytes)
	body.WriteString("mime_type: " + frontmatterScalar(mimeType) + "\n")
	body.WriteString("captured_at: " + frontmatterScalar(capturedAt.Format(time.RFC3339Nano)) + "\n")
	if extracted.Title != "" {
		body.WriteString("source_title: " + frontmatterScalar(extracted.Title) + "\n")
	}
	body.WriteString("---\n")
	body.WriteString("# " + markdownLine(title) + "\n\n")
	body.WriteString("## Summary\n")
	body.WriteString("Web source ingested from " + sourceURL + ".\n\n")
	body.WriteString("## Source Page\n")
	body.WriteString("- Source URL: " + sourceURL + "\n")
	body.WriteString("- SHA256: " + sha + "\n")
	_, _ = fmt.Fprintf(&body, "- Size bytes: %d\n", sizeBytes)
	if extracted.Title != "" {
		body.WriteString("- Page title: " + extracted.Title + "\n")
	}
	body.WriteString("\n## Extracted Text\n")
	if extracted.Text != "" {
		body.WriteString(extracted.Text + "\n")
	} else {
		body.WriteString("No visible page text was found.\n")
	}
	return body.String()
}

var (
	htmlScriptStylePattern = regexp.MustCompile(`(?is)<(script|style|noscript|svg|template)\b[^>]*>.*?</(script|style|noscript|svg|template)>`)
	htmlCommentPattern     = regexp.MustCompile(`(?is)<!--.*?-->`)
	htmlTitlePattern       = regexp.MustCompile(`(?is)<title\b[^>]*>(.*?)</title>`)
	htmlTagPattern         = regexp.MustCompile(`(?is)<[^>]+>`)
	htmlWhitespacePattern  = regexp.MustCompile(`[ \t\r\n]+`)
)

func extractWebPage(body []byte) (webExtraction, error) {
	raw := string(body)
	lower := strings.ToLower(raw)
	for _, marker := range []string{"captcha", "sign in to continue", "login required", "access denied", "enable cookies"} {
		if strings.Contains(lower, marker) {
			return webExtraction{}, domain.ValidationError("source URL appears to require unsupported interactive, authenticated, or restricted access", nil)
		}
	}
	withoutHidden := htmlScriptStylePattern.ReplaceAllString(raw, " ")
	withoutHidden = htmlCommentPattern.ReplaceAllString(withoutHidden, " ")
	title := ""
	if match := htmlTitlePattern.FindStringSubmatch(withoutHidden); len(match) > 1 {
		title = cleanExtractedHTMLText(match[1])
	}
	text := htmlTagPattern.ReplaceAllString(withoutHidden, " ")
	text = cleanExtractedHTMLText(text)
	if text == "" && title == "" {
		return webExtraction{}, domain.ValidationError("source URL did not expose visible HTML text", nil)
	}
	return webExtraction{Title: title, Text: text}, nil
}

func cleanExtractedHTMLText(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = htmlWhitespacePattern.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
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
	sourceType := sourceTypePDF
	if assetPath == "" {
		sourceType = sourceTypeWeb
	}
	requiredMetadata := map[string]string{
		"type":         "source",
		"source_type":  sourceType,
		"modality":     "markdown",
		"source_url":   sourceURL,
		"derived_path": sourcePath,
	}
	if sourceType == sourceTypePDF {
		requiredMetadata["asset_path"] = assetPath
	}
	for key, want := range requiredMetadata {
		if got := strings.TrimSpace(document.Metadata[key]); got != want {
			return domain.InternalError("validate ingested source metadata", fmt.Errorf("metadata %s = %q, want %q", key, got, want))
		}
	}
	if assetAbsPath != "" {
		if _, err := osStat(assetAbsPath); err != nil {
			return domain.InternalError("validate ingested source asset", err)
		}
	} else if strings.TrimSpace(document.Metadata["asset_path"]) != "" {
		return domain.InternalError("validate ingested source asset", errors.New("web source must not record asset_path metadata"))
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
		"source_title",
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
