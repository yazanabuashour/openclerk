package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	machineUnixPathPattern    = regexp.MustCompile(`/(Users|home)/[^/\s"'\\)]+`)
	machineWindowsPathPattern = regexp.MustCompile(`(?i)\b[A-Z]:\\Users\\[^\\\s"']+`)
	privateNotesPathPattern   = regexp.MustCompile(`(^|[\s"'(])~?/notes(/|\b)`)
	realVaultDocPathPattern   = regexp.MustCompile(`\b(sources|synthesis|notes|transcripts|articles|meetings|blogs|receipts|invoices|legal|contracts)/[A-Za-z0-9._/-]+\.md\b`)
	realVaultDocIDPattern     = regexp.MustCompile(`\bdoc_[A-Za-z0-9][A-Za-z0-9_-]*\b`)
	realVaultChunkIDPattern   = regexp.MustCompile(`\bchunk_[A-Za-z0-9][A-Za-z0-9_-]*\b`)
)

var privateResearchNames = []string{
	"agent-first-knowledge-plane-architecture-2026-04-07",
	"full-operator-stack-architecture-2026-03-09",
	"agentic-vault-retrieval-architecture-2026-03-09",
	"mem0-openclaw-self-hosting-vault-comparison-2026-04-06",
	"open-source-canonical-notes-documents-for-agents-2026-04-06",
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 0 {
		return errors.New("usage: scripts/validate-committed-artifacts.sh")
	}
	if err := validateCommittedArtifacts("."); err != nil {
		return err
	}
	_, err := fmt.Fprintln(stdout, "validated committed artifacts")
	return err
}

func validateCommittedArtifacts(root string) error {
	files, err := trackedFiles(root)
	if err != nil {
		return err
	}
	if err := validateArtifactFiles(root, files); err != nil {
		return err
	}
	if err := validateRealVaultReport(root); err != nil {
		return err
	}
	if err := validateNoRealVaultJSONReports(files); err != nil {
		return err
	}
	return nil
}

func trackedFiles(root string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", "-z")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list tracked files: %w", err)
	}
	parts := bytes.Split(output, []byte{0})
	files := []string{}
	for _, part := range parts {
		name := strings.TrimSpace(string(part))
		if name != "" {
			files = append(files, filepath.ToSlash(name))
		}
	}
	return files, nil
}

func validateArtifactFiles(root string, files []string) error {
	for _, rel := range files {
		if !isPublicArtifactPath(rel) {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		if err := validatePublicArtifactText(rel, strings.ReplaceAll(string(content), "\r\n", "\n")); err != nil {
			return err
		}
	}
	return nil
}

func isPublicArtifactPath(rel string) bool {
	rel = filepath.ToSlash(rel)
	ext := filepath.Ext(rel)
	switch ext {
	case ".md", ".json", ".yml", ".yaml", ".toml", ".txt":
	default:
		return false
	}
	if strings.HasPrefix(rel, "docs/") ||
		strings.HasPrefix(rel, "skills/") ||
		strings.HasPrefix(rel, ".github/") {
		return true
	}
	switch rel {
	case "AGENTS.md", "README.md", "CHANGELOG.md", "CONTRIBUTING.md", "SECURITY.md", "CODE_OF_CONDUCT.md":
		return true
	default:
		return false
	}
}

func validatePublicArtifactText(rel string, text string) error {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lineNumber := i + 1
		if match := machineUnixPathPattern.FindString(line); match != "" {
			return fmt.Errorf("%s:%d contains machine-absolute path %q", rel, lineNumber, match)
		}
		if match := machineWindowsPathPattern.FindString(line); match != "" {
			return fmt.Errorf("%s:%d contains machine-absolute path %q", rel, lineNumber, match)
		}
		if match := privateNotesPathPattern.FindString(line); match != "" {
			return fmt.Errorf("%s:%d contains private notes path %q", rel, lineNumber, strings.TrimSpace(match))
		}
		for _, name := range privateResearchNames {
			if strings.Contains(line, name) {
				return fmt.Errorf("%s:%d contains private research note reference %q", rel, lineNumber, name)
			}
		}
		if strings.Contains(line, "events.jsonl") && !containsRunRootPlaceholder(line) {
			return fmt.Errorf("%s:%d references raw eval logs without <run-root> placeholder", rel, lineNumber)
		}
		if strings.Contains(line, `"raw_logs_committed": true`) {
			return fmt.Errorf("%s:%d marks raw eval logs as committed", rel, lineNumber)
		}
	}
	return nil
}

func containsRunRootPlaceholder(line string) bool {
	return strings.Contains(line, "<run-root>/") ||
		strings.Contains(line, `\u003crun-root\u003e/`)
}

func validateRealVaultReport(root string) error {
	rel := filepath.ToSlash(filepath.Join("docs", "evals", "results", "ockp-real-vault-agentops-trial.md"))
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", rel, err)
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	required := []string{
		"sanitized real-vault trial report",
		"omits private paths, titles, snippets, citations, and document identifiers",
		"omits validation document paths and raw JSON",
		"Raw logs",
	}
	normalized := strings.Join(strings.Fields(text), " ")
	for _, want := range required {
		if !strings.Contains(normalized, want) {
			return fmt.Errorf("%s must document sanitized real-vault evidence policy: missing %q", rel, want)
		}
	}
	forbiddenPatterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{name: "vault-relative markdown path", pattern: realVaultDocPathPattern},
		{name: "document id", pattern: realVaultDocIDPattern},
		{name: "chunk id", pattern: realVaultChunkIDPattern},
		{name: "machine-absolute Unix path", pattern: machineUnixPathPattern},
		{name: "machine-absolute Windows path", pattern: machineWindowsPathPattern},
	}
	for _, forbidden := range forbiddenPatterns {
		if match := forbidden.pattern.FindString(text); match != "" {
			return fmt.Errorf("%s contains private real-vault %s %q", rel, forbidden.name, match)
		}
	}
	if strings.Contains(text, "events.jsonl") {
		return fmt.Errorf("%s must not reference private real-vault raw log files", rel)
	}
	return nil
}

func validateNoRealVaultJSONReports(files []string) error {
	for _, rel := range files {
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "docs/evals/results/ockp-real-vault") && strings.HasSuffix(rel, ".json") {
			return fmt.Errorf("%s must stay local-only; commit only the sanitized markdown real-vault report", rel)
		}
	}
	return nil
}
