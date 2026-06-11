package sqlite

import (
	"net/url"
	"path"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

func normalizePath(raw string) (string, error) {
	clean, issue := domain.NormalizeVaultRelativePath(raw)
	switch issue {
	case domain.VaultPathMissing:
		return "", domain.ValidationError("path is required", nil)
	case domain.VaultPathAbsolute:
		return "", domain.ValidationError("path must be repo-relative to the vault root", map[string]any{"path": raw})
	case domain.VaultPathEscapesRoot:
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

func normalizeDocumentSourceURL(raw string) (string, error) {
	sourceURL, _, err := normalizeDocumentSourceURLWithAliases(raw)
	return sourceURL, err
}

func normalizeDocumentSourceURLWithAliases(raw string) (string, []string, error) {
	originalURL, err := normalizeSourceURL(raw)
	if err != nil {
		return "", nil, err
	}
	sourceURL := normalizeGitHubMarkdownSourceURL(originalURL)
	lookupURLs := []string{}
	lookupURLs = appendUnique(lookupURLs, sourceURL)
	lookupURLs = appendUnique(lookupURLs, originalURL)
	for _, alias := range githubMarkdownSourceURLAliases(sourceURL) {
		lookupURLs = appendUnique(lookupURLs, alias)
	}
	for _, alias := range githubMarkdownSourceURLAliases(originalURL) {
		lookupURLs = appendUnique(lookupURLs, alias)
	}
	return sourceURL, lookupURLs, nil
}

func normalizeGitHubMarkdownSourceURL(sourceURL string) string {
	parsed, err := url.Parse(sourceURL)
	if err != nil {
		return sourceURL
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" {
		return sourceURL
	}
	segments, ok := sourceURLPathSegments(parsed)
	if !ok || len(segments) < 2 {
		return sourceURL
	}
	owner := segments[0]
	repo := strings.TrimSuffix(segments[1], ".git")
	if owner == "" || repo == "" {
		return sourceURL
	}
	if len(segments) == 2 {
		return githubRawContentURL(owner, repo, "HEAD", []string{"README.md"})
	}
	if len(segments) >= 5 && (segments[2] == "blob" || segments[2] == "raw") && isMarkdownPathSegments(segments[4:]) {
		return githubRawContentURL(owner, repo, segments[3], segments[4:])
	}
	if len(segments) == 4 && segments[2] == "tree" {
		return githubRawContentURL(owner, repo, segments[3], []string{"README.md"})
	}
	return sourceURL
}

func sourceURLPathSegments(parsed *url.URL) ([]string, bool) {
	trimmed := strings.Trim(parsed.EscapedPath(), "/")
	if trimmed == "" {
		return nil, false
	}
	rawSegments := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, rawSegment := range rawSegments {
		segment, err := url.PathUnescape(rawSegment)
		if err != nil || segment == "" {
			return nil, false
		}
		segments = append(segments, segment)
	}
	return segments, true
}

func githubRawContentURL(owner string, repo string, ref string, filePath []string) string {
	segments := []string{owner, repo, ref}
	segments = append(segments, filePath...)
	return "https://raw.githubusercontent.com/" + escapedURLPath(segments)
}

func githubMarkdownSourceURLAliases(sourceURL string) []string {
	parsed, err := url.Parse(sourceURL)
	if err != nil {
		return nil
	}
	if strings.ToLower(parsed.Hostname()) != "raw.githubusercontent.com" {
		return nil
	}
	segments, ok := sourceURLPathSegments(parsed)
	if !ok || len(segments) < 4 {
		return nil
	}
	filePath := segments[3:]
	if !isMarkdownPathSegments(filePath) {
		return nil
	}
	aliases := []string{}
	owner := segments[0]
	repo := segments[1]
	ref := segments[2]
	aliases = appendUnique(aliases, githubWebContentURL(owner, repo, "blob", ref, filePath))
	if ref == "HEAD" && len(filePath) == 1 && strings.EqualFold(filePath[0], "README.md") {
		repoURL := "https://github.com/" + escapedURLPath([]string{owner, repo})
		aliases = appendUnique(aliases, repoURL)
		aliases = appendUnique(aliases, repoURL+"/")
	}
	return aliases
}

func githubWebContentURL(owner string, repo string, route string, ref string, filePath []string) string {
	segments := []string{owner, repo, route, ref}
	segments = append(segments, filePath...)
	return "https://github.com/" + escapedURLPath(segments)
}

func escapedURLPath(segments []string) string {
	escaped := make([]string, 0, len(segments))
	for _, segment := range segments {
		escaped = append(escaped, url.PathEscape(segment))
	}
	return strings.Join(escaped, "/")
}

func isMarkdownPathSegments(segments []string) bool {
	if len(segments) == 0 {
		return false
	}
	ext := strings.ToLower(path.Ext(segments[len(segments)-1]))
	return isMarkdownPathExtension(ext)
}

func uniqueNonEmptyStrings(values []string) []string {
	unique := []string{}
	for _, value := range values {
		unique = appendUnique(unique, strings.TrimSpace(value))
	}
	return unique
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
	clean, issue := domain.NormalizeVaultRelativePath(raw)
	switch issue {
	case domain.VaultPathMissing:
		return "", domain.ValidationError("path hint is required", nil)
	case domain.VaultPathAbsolute:
		return "", domain.ValidationError("path hint must be relative to the vault root", map[string]any{"path": raw})
	case domain.VaultPathEscapesRoot:
		return "", domain.ValidationError("path hint must stay inside the vault root", map[string]any{"path": raw})
	}
	return clean, nil
}
