package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var releaseTagPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: scripts/validate-release-docs.sh <tag>")
	}
	tag := strings.TrimSpace(args[0])
	if err := validateReleaseDocs(".", tag); err != nil {
		return err
	}
	_, err := fmt.Fprintf(stdout, "validated release docs for %s\n", tag)
	return err
}

func validateReleaseDocs(root string, tag string) error {
	if !releaseTagPattern.MatchString(tag) {
		return fmt.Errorf("tag must match vMAJOR.MINOR.PATCH: %q", tag)
	}

	notesPath := filepath.Join("docs", "release-notes", tag+".md")
	notes, err := readRepoText(root, notesPath)
	if err != nil {
		return err
	}
	if err := validateReleaseNotes(notesPath, notes, tag); err != nil {
		return err
	}

	changelog, err := readRepoText(root, "CHANGELOG.md")
	if err != nil {
		return err
	}
	releaseURL := "https://github.com/yazanabuashour/openclerk/releases/tag/" + tag
	if !strings.Contains(changelog, releaseURL) {
		return fmt.Errorf("CHANGELOG.md must link to %s", releaseURL)
	}
	return nil
}

func readRepoText(root string, relPath string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%s not found", relPath)
		}
		return "", fmt.Errorf("read %s: %w", relPath, err)
	}
	return strings.ReplaceAll(string(content), "\r\n", "\n"), nil
}

func validateReleaseNotes(notesPath string, content string, tag string) error {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "# OpenClerk "+tag {
		return fmt.Errorf("%s must start with %q", notesPath, "# OpenClerk "+tag)
	}
	if !hasMarkdownHeading(lines, "## Changed") {
		return fmt.Errorf("%s must include ## Changed", notesPath)
	}
	if !hasMarkdownHeading(lines, "## Verification") {
		return fmt.Errorf("%s must include ## Verification", notesPath)
	}
	if err := validateNoHardWrappedProse(notesPath, lines); err != nil {
		return err
	}
	if isV02ReleaseTag(tag) {
		if err := validateV02ReleaseNotes(notesPath, content); err != nil {
			return err
		}
	}
	return nil
}

func isV02ReleaseTag(tag string) bool {
	return strings.HasPrefix(tag, "v0.2.")
}

func validateV02ReleaseNotes(notesPath string, content string) error {
	normalized := strings.Join(strings.Fields(content), " ")
	required := []struct {
		name  string
		terms []string
	}{
		{
			name:  "source URL update mode defaults",
			terms: []string{"`ingest_source_url` update mode", "missing `source.mode` defaults to `create`", "`source.mode: \"update\"` re-ingests an existing `source.url`"},
		},
		{
			name:  "duplicate and path-hint conflicts",
			terms: []string{"duplicate creates reject", "`path_hint`", "`asset_path_hint`", "conflict without writing extra docs or assets"},
		},
		{
			name:  "same-SHA no-op behavior",
			terms: []string{"same-SHA updates are no-ops", "`source_updated` provenance", "projection invalidation churn"},
		},
		{
			name:  "changed-PDF stale synthesis visibility",
			terms: []string{"changed-PDF updates preserve source and asset paths", "refresh searchable extracted text and citations", "previous/new SHA provenance", "`projection_states`"},
		},
		{
			name:  "targeted source URL update evidence",
			terms: []string{"docs/evals/results/ockp-source-url-update-mode.md", "path-hint conflict no-write behavior"},
		},
	}
	for _, requirement := range required {
		for _, term := range requirement.terms {
			if !strings.Contains(normalized, term) {
				return fmt.Errorf("%s must document v0.2.0 %s: missing %q", notesPath, requirement.name, term)
			}
		}
	}
	return nil
}

func hasMarkdownHeading(lines []string, heading string) bool {
	for _, line := range lines {
		if strings.TrimSpace(line) == heading {
			return true
		}
	}
	return false
}

func validateNoHardWrappedProse(notesPath string, lines []string) error {
	inFence := false
	previousPlainLine := 0
	previousListLine := 0
	for i, line := range lines {
		lineNumber := i + 1
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			previousPlainLine = 0
			previousListLine = 0
			continue
		}
		if inFence || trimmed == "" {
			previousPlainLine = 0
			previousListLine = 0
			continue
		}
		if previousListLine != 0 && !isListItemLine(line) && !isMarkdownSectionBoundary(trimmed) {
			return fmt.Errorf("%s line %d appears to hard-wrap list item from line %d; keep release-note list items on one source line", notesPath, lineNumber, previousListLine)
		}
		if isListItemLine(line) {
			previousPlainLine = 0
			previousListLine = lineNumber
			continue
		}
		previousListLine = 0
		if !isPlainProseLine(line) {
			previousPlainLine = 0
			continue
		}
		if previousPlainLine != 0 {
			return fmt.Errorf("%s line %d appears to hard-wrap prose from line %d; keep release-note prose paragraphs on one source line", notesPath, lineNumber, previousPlainLine)
		}
		previousPlainLine = lineNumber
	}
	return nil
}

func isPlainProseLine(line string) bool {
	if line == "" || strings.TrimSpace(line) == "" {
		return false
	}
	if line != strings.TrimLeft(line, " \t") {
		return false
	}
	trimmed := strings.TrimSpace(line)
	nonProsePrefixes := []string{
		"#",
		"- ",
		"* ",
		"+ ",
		">",
		"|",
		"```",
		"~~~",
		"[",
		"<!--",
	}
	for _, prefix := range nonProsePrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return false
		}
	}
	if isOrderedListItem(trimmed) || isMarkdownRule(trimmed) {
		return false
	}
	return true
}

func isListItemLine(line string) bool {
	if line != strings.TrimLeft(line, " \t") {
		return false
	}
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "- ") ||
		strings.HasPrefix(trimmed, "* ") ||
		strings.HasPrefix(trimmed, "+ ") ||
		isOrderedListItem(trimmed)
}

func isMarkdownSectionBoundary(trimmed string) bool {
	return strings.HasPrefix(trimmed, "#") || isMarkdownRule(trimmed) || strings.HasPrefix(trimmed, "<!--")
}

func isOrderedListItem(line string) bool {
	dot := strings.IndexByte(line, '.')
	if dot <= 0 || dot == len(line)-1 || line[dot+1] != ' ' {
		return false
	}
	for _, r := range line[:dot] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isMarkdownRule(line string) bool {
	if len(line) < 3 {
		return false
	}
	for _, marker := range []rune{'-', '*', '_'} {
		allMarker := true
		for _, r := range line {
			if r != marker {
				allMarker = false
				break
			}
		}
		if allMarker {
			return true
		}
	}
	return false
}
