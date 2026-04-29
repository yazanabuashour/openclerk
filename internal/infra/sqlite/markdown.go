package sqlite

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type section struct {
	Heading   string
	Level     int
	Content   string
	LineStart int
	LineEnd   int
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func frontmatterScalar(value string) string {
	return strconv.Quote(strings.TrimSpace(value))
}

func markdownLine(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	return strings.TrimSpace(value)
}

func parseMarkdown(body string, relPath string) ([]string, []section, map[string]string) {
	lines := strings.Split(body, "\n")
	frontmatter, contentStart := parseFrontmatter(lines)
	headings := []string{}
	sections := []section{}
	type headingInfo struct {
		index int
		level int
		title string
	}
	infos := []headingInfo{}
	for idx := contentStart; idx < len(lines); idx++ {
		matches := headingPattern.FindStringSubmatch(lines[idx])
		if len(matches) == 0 {
			continue
		}
		title := strings.TrimSpace(matches[2])
		headings = append(headings, title)
		infos = append(infos, headingInfo{
			index: idx,
			level: len(matches[1]),
			title: title,
		})
	}
	if len(infos) == 0 {
		title := documentTitle(relPath, body, nil, frontmatter)
		return []string{}, []section{{
			Heading:   title,
			Level:     1,
			Content:   strings.TrimSpace(body),
			LineStart: contentStart + 1,
			LineEnd:   len(lines),
		}}, frontmatter
	}
	if infos[0].index > contentStart {
		preamble := strings.TrimSpace(strings.Join(lines[contentStart:infos[0].index], "\n"))
		if preamble != "" {
			sections = append(sections, section{
				Heading:   documentTitle(relPath, body, headings, frontmatter),
				Level:     1,
				Content:   preamble,
				LineStart: contentStart + 1,
				LineEnd:   infos[0].index,
			})
		}
	}
	for i, info := range infos {
		end := len(lines)
		if i+1 < len(infos) {
			end = infos[i+1].index
		}
		sections = append(sections, section{
			Heading:   info.title,
			Level:     info.level,
			Content:   strings.TrimSpace(strings.Join(lines[info.index:end], "\n")),
			LineStart: info.index + 1,
			LineEnd:   end,
		})
	}
	return headings, sections, frontmatter
}

func parseFrontmatter(lines []string) (map[string]string, int) {
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return map[string]string{}, 0
	}
	frontmatter := map[string]string{}
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			return frontmatter, idx + 1
		}
		key, value, ok := strings.Cut(lines[idx], ":")
		if !ok {
			continue
		}
		frontmatter[strings.TrimSpace(strings.ToLower(key))] = cleanFrontmatterValue(value)
	}
	return map[string]string{}, 0
}

func cleanFrontmatterValue(value string) string {
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

func documentTitle(relPath string, body string, headings []string, frontmatter map[string]string) string {
	return resolvedDocumentTitle(relPath, body, headings, frontmatter, "", "")
}

func resolvedDocumentTitle(relPath string, body string, headings []string, frontmatter map[string]string, preferredTitle string, existingTitle string) string {
	if title := strings.TrimSpace(preferredTitle); title != "" {
		return title
	}
	if title := strings.TrimSpace(frontmatter["title"]); title != "" {
		return title
	}
	if len(headings) > 0 {
		return headings[0]
	}
	if title := strings.TrimSpace(existingTitle); title != "" {
		return title
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "---" || strings.Contains(line, ":") {
			continue
		}
		return strings.TrimPrefix(line, "# ")
	}
	return strings.TrimSuffix(path.Base(relPath), path.Ext(relPath))
}

func replaceSection(body string, targetHeading string, content string) (string, error) {
	lines := strings.Split(body, "\n")
	_, contentStart := parseFrontmatter(lines)
	targetHeading = strings.TrimSpace(targetHeading)
	for idx := contentStart; idx < len(lines); idx++ {
		matches := headingPattern.FindStringSubmatch(lines[idx])
		if len(matches) == 0 {
			continue
		}
		if strings.TrimSpace(matches[2]) != targetHeading {
			continue
		}
		level := len(matches[1])
		end := len(lines)
		for next := idx + 1; next < len(lines); next++ {
			nextMatches := headingPattern.FindStringSubmatch(lines[next])
			if len(nextMatches) == 0 {
				continue
			}
			if len(nextMatches[1]) <= level {
				end = next
				break
			}
		}
		replacement := []string{lines[idx]}
		replacement = append(replacement, strings.Split(strings.TrimSpace(content), "\n")...)
		updated := append([]string{}, lines[:idx]...)
		updated = append(updated, replacement...)
		updated = append(updated, lines[end:]...)
		return strings.TrimRight(strings.Join(updated, "\n"), "\n") + "\n", nil
	}
	return "", domain.NotFoundError("heading", targetHeading)
}

func docIDForPath(relPath string) string {
	return hashID("doc", relPath)
}

func chunkIDForSection(docID string, sec section) string {
	return hashID("chunk", docID, sec.Heading, sec.Content, strconv.Itoa(sec.LineStart), strconv.Itoa(sec.LineEnd))
}

func hashID(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:8])
}

func encodeCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeCursor(cursor string) int {
	if cursor == "" {
		return 0
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0
	}
	value, err := strconv.Atoi(string(decoded))
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func ftsExpression(text string) string {
	parts := wordPattern.FindAllString(strings.ToLower(text), -1)
	if len(parts) == 0 {
		return `"` + strings.ReplaceAll(strings.TrimSpace(text), `"`, `""`) + `"`
	}
	return strings.Join(parts, " ")
}

func documentsByPath(documents []domain.Document) map[string]domain.Document {
	result := make(map[string]domain.Document, len(documents))
	for _, doc := range documents {
		result[doc.Path] = doc
	}
	return result
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func splitPathList(value string) []string {
	parts := strings.Split(value, ",")
	result := []string{}
	seen := map[string]struct{}{}
	for _, part := range parts {
		normalized := normalizeReferencePath(part)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func splitCSVList(value string) []string {
	parts := strings.Split(value, ",")
	result := []string{}
	seen := map[string]struct{}{}
	for _, part := range parts {
		normalized := strings.TrimSpace(part)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func normalizeReferencePath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "doc:") {
		return value
	}
	clean := path.Clean(filepath.ToSlash(value))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return ""
	}
	if path.Ext(clean) == "" {
		clean += ".md"
	}
	return clean
}

func appendUnique(values []string, value string) []string {
	if strings.TrimSpace(value) == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func firstNRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func ensureDir(dir string) error {
	return osMkdirAll(dir, 0o755)
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func mustParseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
