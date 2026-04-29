package sqlite

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"io/fs"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type normalizedVideoTranscript struct {
	Text       string
	Policy     string
	Origin     string
	Language   string
	CapturedAt time.Time
	Tool       string
	Model      string
	SHA256     string
}

func (s *Store) IngestVideoURL(ctx context.Context, input domain.VideoURLInput) (domain.VideoIngestionResult, error) {
	videoURL, err := normalizeSourceURL(input.URL)
	if err != nil {
		return domain.VideoIngestionResult{}, err
	}
	mode, err := normalizeVideoURLMode(input.Mode)
	if err != nil {
		return domain.VideoIngestionResult{}, err
	}
	normalized, err := normalizeVideoTranscript(input.Transcript, s.now().UTC())
	if err != nil {
		return domain.VideoIngestionResult{}, err
	}
	switch mode {
	case sourceURLModeCreate:
		return s.createVideoURL(ctx, input, videoURL, normalized)
	case sourceURLModeUpdate:
		return s.updateVideoURL(ctx, input, videoURL, normalized)
	default:
		return domain.VideoIngestionResult{}, domain.ValidationError("video mode must be create or update", map[string]any{"mode": input.Mode})
	}
}

func (s *Store) createVideoURL(ctx context.Context, input domain.VideoURLInput, videoURL string, transcript normalizedVideoTranscript) (domain.VideoIngestionResult, error) {
	sourcePath, err := normalizeSourceDocumentPath(input.PathHint)
	if err != nil {
		return domain.VideoIngestionResult{}, err
	}
	assetPath := ""
	if strings.TrimSpace(input.AssetPathHint) != "" {
		assetPath, err = normalizeVideoAssetPath(input.AssetPathHint)
		if err != nil {
			return domain.VideoIngestionResult{}, err
		}
	}
	if exists, err := s.sourceURLExists(ctx, videoURL); err != nil {
		return domain.VideoIngestionResult{}, err
	} else if exists {
		return domain.VideoIngestionResult{}, domain.AlreadyExistsError("source URL", videoURL)
	}
	sourceAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(sourcePath))
	if _, err := osStat(sourceAbsPath); err == nil {
		return domain.VideoIngestionResult{}, domain.AlreadyExistsError("document path", sourcePath)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return domain.VideoIngestionResult{}, domain.InternalError("stat video source document path", err)
	}
	assetWritten := false
	if assetPath != "" {
		assetAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(assetPath))
		if _, err := osStat(assetAbsPath); err == nil {
			return domain.VideoIngestionResult{}, domain.AlreadyExistsError("asset path", assetPath)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return domain.VideoIngestionResult{}, domain.InternalError("stat video metadata asset path", err)
		}
		assetBytes, err := buildVideoMetadataAsset(videoURL, sourcePath, assetPath, transcript)
		if err != nil {
			return domain.VideoIngestionResult{}, err
		}
		if err := ensureDir(filepath.Dir(assetAbsPath)); err != nil {
			return domain.VideoIngestionResult{}, domain.InternalError("create video metadata asset directory", err)
		}
		if err := osWriteBytes(assetAbsPath, assetBytes); err != nil {
			return domain.VideoIngestionResult{}, domain.InternalError("write video metadata asset", err)
		}
		assetWritten = true
		defer func() {
			if assetWritten {
				if _, err := osStat(sourceAbsPath); errors.Is(err, fs.ErrNotExist) {
					_ = osRemove(assetAbsPath)
				}
			}
		}()
	}
	title := resolvedVideoTitle(input.Title, sourcePath)
	body := buildVideoSourceNoteBody(videoURL, sourcePath, assetPath, title, transcript)
	document, err := s.CreateDocument(ctx, domain.CreateDocumentInput{
		Path:  sourcePath,
		Title: title,
		Body:  body,
	})
	if err != nil {
		return domain.VideoIngestionResult{}, err
	}
	assetWritten = false
	citations, err := s.sourceDocumentCitations(ctx, document)
	if err != nil {
		return domain.VideoIngestionResult{}, err
	}
	if len(citations) == 0 {
		return domain.VideoIngestionResult{}, domain.InternalError("validate video source ingestion citations", errors.New("created video source has no indexed citations"))
	}
	if err := validateIngestedVideoSource(document, videoURL, sourcePath, assetPath); err != nil {
		return domain.VideoIngestionResult{}, err
	}
	return videoIngestionResultFromDocument(document, citations), nil
}

func (s *Store) updateVideoURL(ctx context.Context, input domain.VideoURLInput, videoURL string, transcript normalizedVideoTranscript) (domain.VideoIngestionResult, error) {
	document, err := s.videoDocumentByURL(ctx, videoURL)
	if err != nil {
		return domain.VideoIngestionResult{}, err
	}
	sourcePath := document.Path
	assetPath := strings.TrimSpace(document.Metadata["asset_path"])
	if strings.TrimSpace(input.PathHint) != "" {
		requestPath, err := normalizeSourceDocumentPath(input.PathHint)
		if err != nil {
			return domain.VideoIngestionResult{}, err
		}
		if requestPath != sourcePath {
			return domain.VideoIngestionResult{}, domain.ConflictError("video path hint does not match existing source", map[string]any{"source_url": videoURL, "path_hint": requestPath, "existing_path": sourcePath})
		}
	}
	if strings.TrimSpace(input.AssetPathHint) != "" {
		requestAssetPath, err := normalizeVideoAssetPath(input.AssetPathHint)
		if err != nil {
			return domain.VideoIngestionResult{}, err
		}
		if requestAssetPath != assetPath {
			return domain.VideoIngestionResult{}, domain.ConflictError("video asset path hint does not match existing source", map[string]any{"source_url": videoURL, "asset_path_hint": requestAssetPath, "existing_asset_path": assetPath})
		}
	}
	if existing := strings.TrimSpace(document.Metadata["transcript_sha256"]); existing == transcript.SHA256 {
		citations, err := s.sourceDocumentCitations(ctx, document)
		if err != nil {
			return domain.VideoIngestionResult{}, err
		}
		return videoIngestionResultFromDocument(document, citations), nil
	}

	sourceAbsPath := filepath.Join(s.vaultRoot, filepath.FromSlash(sourcePath))
	oldBody := document.Body
	title := resolvedVideoTitle(firstNonEmpty(input.Title, document.Title), sourcePath)
	body := buildVideoSourceNoteBody(videoURL, sourcePath, assetPath, title, transcript)
	assetAbsPath := ""
	var oldAssetBytes []byte
	if assetPath != "" {
		assetAbsPath = filepath.Join(s.vaultRoot, filepath.FromSlash(assetPath))
		oldAssetBytes, err = osReadFile(assetAbsPath)
		if err != nil {
			return domain.VideoIngestionResult{}, domain.InternalError("read existing video metadata asset", err)
		}
		newAssetBytes, err := buildVideoMetadataAsset(videoURL, sourcePath, assetPath, transcript)
		if err != nil {
			return domain.VideoIngestionResult{}, err
		}
		if err := s.replaceVideoAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, newAssetBytes, body, title); err != nil {
			return domain.VideoIngestionResult{}, err
		}
	} else {
		if err := osWriteFile(sourceAbsPath, body); err != nil {
			return domain.VideoIngestionResult{}, domain.InternalError("write video source note", err)
		}
		if err := s.syncDocumentFromDisk(ctx, sourcePath, title); err != nil {
			if restoreErr := osWriteFile(sourceAbsPath, oldBody); restoreErr != nil {
				return domain.VideoIngestionResult{}, domain.InternalError("restore video source note after failed update", errors.Join(err, restoreErr))
			}
			if restoreErr := s.syncDocumentFromDisk(ctx, sourcePath, ""); restoreErr != nil {
				return domain.VideoIngestionResult{}, domain.InternalError("restore indexed video source after failed update", errors.Join(err, restoreErr))
			}
			return domain.VideoIngestionResult{}, err
		}
	}

	updated, err := s.GetDocument(ctx, document.DocID)
	if err != nil {
		if assetPath != "" {
			return domain.VideoIngestionResult{}, s.restoreVideoAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err)
		}
		return domain.VideoIngestionResult{}, err
	}
	citations, err := s.sourceDocumentCitations(ctx, updated)
	if err != nil {
		if assetPath != "" {
			return domain.VideoIngestionResult{}, s.restoreVideoAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err)
		}
		return domain.VideoIngestionResult{}, err
	}
	if len(citations) == 0 {
		if assetPath != "" {
			return domain.VideoIngestionResult{}, s.restoreVideoAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, domain.InternalError("validate video source update citations", errors.New("updated video source has no indexed citations")))
		}
		return domain.VideoIngestionResult{}, domain.InternalError("validate video source update citations", errors.New("updated video source has no indexed citations"))
	}
	if err := validateIngestedVideoSource(updated, videoURL, sourcePath, assetPath); err != nil {
		if assetPath != "" {
			return domain.VideoIngestionResult{}, s.restoreVideoAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err)
		}
		return domain.VideoIngestionResult{}, err
	}
	result := videoIngestionResultFromDocument(updated, citations)
	result.PreviousTranscriptSHA256 = strings.TrimSpace(document.Metadata["transcript_sha256"])
	result.NewTranscriptSHA256 = transcript.SHA256
	return result, nil
}

func (s *Store) videoDocumentByURL(ctx context.Context, sourceURL string) (domain.Document, error) {
	document, err := s.sourceDocumentByURL(ctx, sourceURL)
	if err != nil {
		return domain.Document{}, err
	}
	if strings.TrimSpace(document.Metadata["source_type"]) != "video_transcript" {
		return domain.Document{}, domain.ConflictError("source URL belongs to a non-video source", map[string]any{"source_url": sourceURL, "source_type": document.Metadata["source_type"], "path": document.Path})
	}
	return document, nil
}

func videoIngestionResultFromDocument(document domain.Document, citations []domain.Citation) domain.VideoIngestionResult {
	return domain.VideoIngestionResult{
		DocID:            document.DocID,
		SourcePath:       document.Path,
		SourceURL:        strings.TrimSpace(document.Metadata["source_url"]),
		AssetPath:        strings.TrimSpace(document.Metadata["asset_path"]),
		Citations:        citations,
		TranscriptSHA256: strings.TrimSpace(document.Metadata["transcript_sha256"]),
		CapturedAt:       parseTimeMetadata(document.Metadata["captured_at"]),
		TranscriptPolicy: strings.TrimSpace(document.Metadata["transcript_policy"]),
		TranscriptOrigin: strings.TrimSpace(document.Metadata["transcript_origin"]),
		Language:         strings.TrimSpace(document.Metadata["language"]),
		Tool:             strings.TrimSpace(document.Metadata["tool"]),
		Model:            strings.TrimSpace(document.Metadata["model"]),
	}
}

func (s *Store) replaceVideoAssetAndNote(ctx context.Context, sourcePath string, sourceAbsPath string, assetAbsPath string, oldBody string, oldAssetBytes []byte, newAssetBytes []byte, newBody string, title string) error {
	if err := osWriteBytes(assetAbsPath, newAssetBytes); err != nil {
		return domain.InternalError("write video metadata asset", err)
	}
	if err := osWriteFile(sourceAbsPath, newBody); err != nil {
		return s.restoreVideoAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, domain.InternalError("write video source note", err))
	}
	if err := s.syncDocumentFromDisk(ctx, sourcePath, title); err != nil {
		return s.restoreVideoAssetAndNote(ctx, sourcePath, sourceAbsPath, assetAbsPath, oldBody, oldAssetBytes, err)
	}
	return nil
}

func (s *Store) restoreVideoAssetAndNote(ctx context.Context, sourcePath string, sourceAbsPath string, assetAbsPath string, oldBody string, oldAssetBytes []byte, cause error) error {
	if restoreErr := osWriteBytes(assetAbsPath, oldAssetBytes); restoreErr != nil {
		return domain.InternalError("restore video metadata asset after failed update", errors.Join(cause, restoreErr))
	}
	if restoreErr := osWriteFile(sourceAbsPath, oldBody); restoreErr != nil {
		return domain.InternalError("restore video source note after failed update", errors.Join(cause, restoreErr))
	}
	if restoreErr := s.syncDocumentFromDisk(ctx, sourcePath, ""); restoreErr != nil {
		return domain.InternalError("restore indexed video source after failed update", errors.Join(cause, restoreErr))
	}
	return cause
}

func normalizeVideoTranscript(input domain.VideoTranscriptInput, defaultCapturedAt time.Time) (normalizedVideoTranscript, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return normalizedVideoTranscript{}, domain.ValidationError("video transcript text is required; ingest_video_url v1 does not download videos, read platform captions, run local transcription, call transcript APIs, or use Gemini extraction", nil)
	}
	policy := strings.TrimSpace(input.Policy)
	if policy == "" {
		policy = "supplied"
	}
	if policy != "supplied" && policy != "local_first" {
		return normalizedVideoTranscript{}, domain.ValidationError("video transcript policy is not supported by ingest_video_url v1", map[string]any{"policy": input.Policy, "supported": "supplied, local_first"})
	}
	capturedAt := defaultCapturedAt
	if raw := strings.TrimSpace(input.CapturedAt); raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return normalizedVideoTranscript{}, domain.ValidationError("video transcript captured_at must be RFC3339", map[string]any{"captured_at": input.CapturedAt})
		}
		capturedAt = parsed.UTC()
	}
	sum := sha256.Sum256([]byte(text))
	shaHex := hex.EncodeToString(sum[:])
	if supplied := strings.ToLower(strings.TrimSpace(input.SHA256)); supplied != "" && supplied != shaHex {
		return normalizedVideoTranscript{}, domain.ValidationError("video transcript sha256 does not match transcript.text", map[string]any{"sha256": input.SHA256})
	}
	origin := strings.TrimSpace(input.Origin)
	if origin == "" {
		origin = "user_supplied_transcript"
	}
	return normalizedVideoTranscript{
		Text:       text,
		Policy:     policy,
		Origin:     origin,
		Language:   strings.TrimSpace(input.Language),
		CapturedAt: capturedAt,
		Tool:       strings.TrimSpace(input.Tool),
		Model:      strings.TrimSpace(input.Model),
		SHA256:     shaHex,
	}, nil
}

func parseInt64Metadata(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseIntMetadata(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}

func parseTimeMetadata(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func resolvedVideoTitle(requestTitle string, sourcePath string) string {
	if title := strings.TrimSpace(requestTitle); title != "" {
		return title
	}
	return strings.TrimSuffix(path.Base(sourcePath), path.Ext(sourcePath))
}

func buildVideoSourceNoteBody(sourceURL string, sourcePath string, assetPath string, title string, transcript normalizedVideoTranscript) string {
	var body strings.Builder
	body.WriteString("---\n")
	body.WriteString("type: source\n")
	body.WriteString("source_type: video_transcript\n")
	body.WriteString("modality: markdown\n")
	body.WriteString("source_url: " + frontmatterScalar(sourceURL) + "\n")
	if assetPath != "" {
		body.WriteString("asset_path: " + frontmatterScalar(assetPath) + "\n")
	}
	body.WriteString("derived_path: " + frontmatterScalar(sourcePath) + "\n")
	body.WriteString("transcript_origin: " + frontmatterScalar(transcript.Origin) + "\n")
	body.WriteString("transcript_policy: " + frontmatterScalar(transcript.Policy) + "\n")
	if transcript.Language != "" {
		body.WriteString("language: " + frontmatterScalar(transcript.Language) + "\n")
	}
	body.WriteString("captured_at: " + frontmatterScalar(transcript.CapturedAt.Format(time.RFC3339Nano)) + "\n")
	body.WriteString("transcript_sha256: " + frontmatterScalar(transcript.SHA256) + "\n")
	if transcript.Tool != "" {
		body.WriteString("tool: " + frontmatterScalar(transcript.Tool) + "\n")
	}
	if transcript.Model != "" {
		body.WriteString("model: " + frontmatterScalar(transcript.Model) + "\n")
	}
	body.WriteString("---\n")
	body.WriteString("# " + markdownLine(title) + "\n\n")
	body.WriteString("## Summary\n")
	body.WriteString("Video transcript source ingested from " + sourceURL + ".\n\n")
	body.WriteString("## Source Video\n")
	body.WriteString("- Source URL: " + sourceURL + "\n")
	if assetPath != "" {
		body.WriteString("- Metadata asset path: " + assetPath + "\n")
	}
	body.WriteString("- Transcript SHA256: " + transcript.SHA256 + "\n\n")
	body.WriteString("## Transcript Provenance\n")
	body.WriteString("- Origin: " + transcript.Origin + "\n")
	body.WriteString("- Policy: " + transcript.Policy + "\n")
	if transcript.Language != "" {
		body.WriteString("- Language: " + transcript.Language + "\n")
	}
	body.WriteString("- Captured at: " + transcript.CapturedAt.Format(time.RFC3339Nano) + "\n")
	if transcript.Tool != "" {
		body.WriteString("- Tool: " + transcript.Tool + "\n")
	}
	if transcript.Model != "" {
		body.WriteString("- Model: " + transcript.Model + "\n")
	}
	body.WriteString("\n## Transcript\n")
	body.WriteString(transcript.Text + "\n")
	return body.String()
}

func buildVideoMetadataAsset(sourceURL string, sourcePath string, assetPath string, transcript normalizedVideoTranscript) ([]byte, error) {
	payload := map[string]string{
		"source_url":        sourceURL,
		"source_path":       sourcePath,
		"asset_path":        assetPath,
		"source_type":       "video_transcript",
		"transcript_origin": transcript.Origin,
		"transcript_policy": transcript.Policy,
		"captured_at":       transcript.CapturedAt.Format(time.RFC3339Nano),
		"transcript_sha256": transcript.SHA256,
	}
	if transcript.Language != "" {
		payload["language"] = transcript.Language
	}
	if transcript.Tool != "" {
		payload["tool"] = transcript.Tool
	}
	if transcript.Model != "" {
		payload["model"] = transcript.Model
	}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, domain.InternalError("build video metadata asset", err)
	}
	return append(body, '\n'), nil
}

func validateIngestedVideoSource(document domain.Document, sourceURL string, sourcePath string, assetPath string) error {
	if document.DocID == "" || document.Path != sourcePath {
		return domain.InternalError("validate ingested video source document", fmt.Errorf("created document path %q does not match %q", document.Path, sourcePath))
	}
	requiredMetadata := map[string]string{
		"type":              "source",
		"source_type":       "video_transcript",
		"modality":          "markdown",
		"source_url":        sourceURL,
		"derived_path":      sourcePath,
		"transcript_policy": strings.TrimSpace(document.Metadata["transcript_policy"]),
		"transcript_origin": strings.TrimSpace(document.Metadata["transcript_origin"]),
		"transcript_sha256": strings.TrimSpace(document.Metadata["transcript_sha256"]),
		"captured_at":       strings.TrimSpace(document.Metadata["captured_at"]),
	}
	if assetPath != "" {
		requiredMetadata["asset_path"] = assetPath
	}
	for key, want := range requiredMetadata {
		if want == "" {
			return domain.InternalError("validate ingested video source metadata", fmt.Errorf("metadata %s is required", key))
		}
		if got := strings.TrimSpace(document.Metadata[key]); got != want {
			return domain.InternalError("validate ingested video source metadata", fmt.Errorf("metadata %s = %q, want %q", key, got, want))
		}
	}
	return nil
}
