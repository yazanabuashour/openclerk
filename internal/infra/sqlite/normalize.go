package sqlite

import (
	"github.com/yazanabuashour/openclerk/internal/domain"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

func normalizePath(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", domain.ValidationError("path is required", nil)
	}
	if filepath.IsAbs(raw) {
		return "", domain.ValidationError("path must be repo-relative to the vault root", map[string]any{"path": raw})
	}
	clean := path.Clean(filepath.ToSlash(raw))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", domain.ValidationError("path must stay inside the vault root", map[string]any{"path": raw})
	}
	if ext := path.Ext(clean); ext == "" {
		clean += ".md"
	} else if ext != ".md" {
		return "", domain.ValidationError("path must end with .md", map[string]any{"path": raw})
	}
	return clean, nil
}

func normalizeSourceURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", domain.ValidationError("source url is required", nil)
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", domain.ValidationError("source url must be a valid http or https URL", nil)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", domain.ValidationError("source url must use http or https", map[string]any{"scheme": parsed.Scheme})
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	return parsed.String(), nil
}

func normalizeSourceURLMode(raw string) (string, error) {
	mode := strings.TrimSpace(raw)
	if mode == "" {
		return sourceURLModeCreate, nil
	}
	if mode != sourceURLModeCreate && mode != sourceURLModeUpdate {
		return "", domain.ValidationError("source mode must be create or update", map[string]any{"mode": raw})
	}
	return mode, nil
}

func normalizeSourceType(raw string) (string, error) {
	sourceType := strings.TrimSpace(raw)
	if sourceType == "" {
		return "", nil
	}
	if sourceType != sourceTypePDF && sourceType != sourceTypeWeb {
		return "", domain.ValidationError("source.source_type must be pdf or web", map[string]any{"source_type": raw})
	}
	return sourceType, nil
}

func normalizeVideoURLMode(raw string) (string, error) {
	mode := strings.TrimSpace(raw)
	if mode == "" {
		return sourceURLModeCreate, nil
	}
	if mode != sourceURLModeCreate && mode != sourceURLModeUpdate {
		return "", domain.ValidationError("video mode must be create or update", map[string]any{"mode": raw})
	}
	return mode, nil
}

func normalizeSourceDocumentPath(raw string) (string, error) {
	clean, err := normalizeStrictVaultPath(raw)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(clean, "sources/") || path.Ext(clean) != ".md" {
		return "", domain.ValidationError("source path hint must be a vault-relative sources/*.md path", map[string]any{"path": raw})
	}
	return clean, nil
}

func normalizeSourceAssetPath(raw string) (string, error) {
	clean, err := normalizeStrictVaultPath(raw)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(clean, "assets/") || path.Ext(clean) != ".pdf" {
		return "", domain.ValidationError("source asset path hint must be a vault-relative assets/**/*.pdf path", map[string]any{"path": raw})
	}
	return clean, nil
}

func normalizeVideoAssetPath(raw string) (string, error) {
	clean, err := normalizeStrictVaultPath(raw)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(clean, "assets/") || path.Ext(clean) != ".json" {
		return "", domain.ValidationError("video asset path hint must be a vault-relative assets/**/*.json path", map[string]any{"path": raw})
	}
	return clean, nil
}

func normalizeStrictVaultPath(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", domain.ValidationError("path hint is required", nil)
	}
	if filepath.IsAbs(raw) || strings.HasPrefix(strings.TrimSpace(raw), "/") {
		return "", domain.ValidationError("path hint must be relative to the vault root", map[string]any{"path": raw})
	}
	clean := path.Clean(filepath.ToSlash(strings.TrimSpace(raw)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", domain.ValidationError("path hint must stay inside the vault root", map[string]any{"path": raw})
	}
	return clean, nil
}
