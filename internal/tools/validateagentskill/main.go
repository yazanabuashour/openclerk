package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	skillNamePattern       = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	markdownLinkPattern    = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	retiredRunnerPattern   = regexp.MustCompile(`\b(openclerkd|openclerk-agentops)\b`)
	retiredTransportRegexp = regexp.MustCompile(`\b(openapi|OpenAPI|HTTP server|MCP transport|SQLite fallback|generated client)\b`)
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: scripts/validate-agent-skill.sh <skill-directory>")
	}
	skillDir := strings.TrimRight(args[0], string(os.PathSeparator))
	if skillDir == "" {
		skillDir = "."
	}
	if err := validateSkillDir(skillDir); err != nil {
		return err
	}
	_, err := fmt.Fprintf(stdout, "validated %s\n", skillDir)
	return err
}

func validateSkillDir(skillDir string) error {
	info, err := os.Stat(skillDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill directory not found: %s", skillDir)
		}
		return fmt.Errorf("stat skill directory %s: %w", skillDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill directory not found: %s", skillDir)
	}

	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return fmt.Errorf("read skill directory %s: %w", skillDir, err)
	}
	if len(entries) != 1 || entries[0].Name() != "SKILL.md" || entries[0].IsDir() {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		return fmt.Errorf("%s must contain only SKILL.md; found %s", skillDir, strings.Join(names, ", "))
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", skillFile, err)
	}
	metadata, err := extractScalarFrontmatter(skillFile, string(content))
	if err != nil {
		return err
	}
	if err := validateMetadata(skillDir, skillFile, metadata); err != nil {
		return err
	}
	if err := validateMarkdownLinks(skillDir, skillFile, string(content)); err != nil {
		return err
	}
	if err := validateRetiredGuidance(skillFile, string(content)); err != nil {
		return err
	}
	return nil
}

func extractScalarFrontmatter(skillFile string, content string) (map[string]string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSuffix(lines[0], "\r") != "---" {
		return nil, fmt.Errorf("%s must start with YAML frontmatter delimited by ---", skillFile)
	}

	metadata := map[string]string{}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSuffix(lines[i], "\r")
		if line == "---" {
			if i == 1 {
				return nil, fmt.Errorf("%s frontmatter must contain at least the required fields", skillFile)
			}
			return metadata, nil
		}
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("%s frontmatter field %q must be a scalar key-value pair", skillFile, line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("%s frontmatter keys must be non-empty strings", skillFile)
		}
		if _, exists := metadata[key]; exists {
			return nil, fmt.Errorf("%s frontmatter field %q must not be duplicated", skillFile, key)
		}
		metadata[key] = strings.Trim(value, `"'`)
	}
	return nil, fmt.Errorf("%s must include a closing --- line for YAML frontmatter", skillFile)
}

func validateMetadata(skillDir string, skillFile string, metadata map[string]string) error {
	name := metadata["name"]
	if name == "" {
		return fmt.Errorf("%s frontmatter must define a non-empty name", skillFile)
	}
	parentDir := filepath.Base(skillDir)
	if name != parentDir {
		return fmt.Errorf("%s name must match the parent directory (%q)", skillFile, parentDir)
	}
	if len([]rune(name)) > 64 {
		return fmt.Errorf("%s name must be 64 characters or fewer", skillFile)
	}
	if !skillNamePattern.MatchString(name) {
		return fmt.Errorf("%s name must use lowercase letters, numbers, and single hyphens only", skillFile)
	}

	description := metadata["description"]
	if description == "" {
		return fmt.Errorf("%s frontmatter must define a non-empty description", skillFile)
	}
	if len([]rune(description)) > 1024 {
		return fmt.Errorf("%s description must be 1024 characters or fewer", skillFile)
	}

	if compatibility, ok := metadata["compatibility"]; ok {
		if compatibility == "" {
			return fmt.Errorf("%s compatibility must be non-empty when provided", skillFile)
		}
		if len([]rune(compatibility)) > 500 {
			return fmt.Errorf("%s compatibility must be 500 characters or fewer", skillFile)
		}
	}
	return nil
}

func validateMarkdownLinks(skillDir string, skillFile string, content string) error {
	for _, match := range markdownLinkPattern.FindAllStringSubmatch(content, -1) {
		target := match[1]
		if shouldSkipLinkTarget(target) {
			continue
		}
		targetPath, err := containedSkillLinkTarget(skillDir, target)
		if err != nil {
			return fmt.Errorf("%s link target %q escapes skill directory", skillFile, target)
		}
		if _, err := os.Stat(targetPath); err != nil {
			return fmt.Errorf("%s link target %q is not installed with the skill: %w", skillFile, target, err)
		}
	}
	return nil
}

func containedSkillLinkTarget(skillDir string, target string) (string, error) {
	base, err := filepath.Abs(skillDir)
	if err != nil {
		return "", err
	}
	targetPath, err := filepath.Abs(filepath.Join(skillDir, target))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(base, targetPath)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escapes skill directory")
	}
	return targetPath, nil
}

func shouldSkipLinkTarget(target string) bool {
	if target == "" || strings.HasPrefix(target, "#") || filepath.IsAbs(target) {
		return true
	}
	if parsed, err := url.Parse(target); err == nil && parsed.Scheme != "" {
		return true
	}
	return false
}

func validateRetiredGuidance(skillFile string, content string) error {
	forbiddenSubstrings := []string{
		"go run ./cmd/openclerk",
		"CLI fallback",
		"Generated Client Fallback",
		"temporary Go module",
		"generated files",
		"Open" + "Client",
		"client/" + "openclerk",
		"cmd/openclerkd",
		".agents/skills",
		".claude/skills",
		".openclaw/skills",
	}
	for _, forbidden := range forbiddenSubstrings {
		if strings.Contains(content, forbidden) {
			return fmt.Errorf("%s contains retired product guidance %q", skillFile, forbidden)
		}
	}
	if retiredRunnerPattern.MatchString(content) {
		return fmt.Errorf("%s contains retired product binary name", skillFile)
	}
	if retiredTransportRegexp.MatchString(content) {
		return fmt.Errorf("%s contains retired transport guidance", skillFile)
	}
	return nil
}
