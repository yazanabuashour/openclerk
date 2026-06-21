package domain

import (
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

type VaultPathIssue string

const (
	VaultPathOK          VaultPathIssue = ""
	VaultPathMissing     VaultPathIssue = "missing"
	VaultPathAbsolute    VaultPathIssue = "absolute"
	VaultPathEscapesRoot VaultPathIssue = "escapes_root"
)

var defaultVaultIgnorePaths = []string{
	".stversions",
	".git",
	".openclerk",
	".backups",
}

type VaultIgnoreMatcher struct {
	rules []string
}

func NormalizeVaultRelativePath(raw string) (string, VaultPathIssue) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", VaultPathMissing
	}
	slashed := slashVaultPath(trimmed)
	if filepath.IsAbs(trimmed) || strings.HasPrefix(slashed, "/") || isWindowsAbsolutePath(slashed) {
		return "", VaultPathAbsolute
	}
	clean := path.Clean(slashed)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", VaultPathEscapesRoot
	}
	return clean, VaultPathOK
}

func DefaultVaultIgnorePaths() []string {
	return append([]string(nil), defaultVaultIgnorePaths...)
}

func EffectiveVaultIgnorePaths(paths []string) ([]string, error) {
	combined := append(DefaultVaultIgnorePaths(), paths...)
	return NormalizeVaultIgnorePaths(combined)
}

func NormalizeVaultIgnorePaths(paths []string) ([]string, error) {
	normalized := make([]string, 0, len(paths))
	seen := map[string]struct{}{}
	for _, raw := range paths {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		clean, issue := NormalizeVaultRelativePath(raw)
		if issue != VaultPathOK {
			return nil, ValidationError("vault ignore path must be vault-relative", map[string]any{
				"path":  raw,
				"issue": string(issue),
			})
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		normalized = append(normalized, clean)
	}
	return normalized, nil
}

func NewVaultIgnoreMatcher(paths []string) (VaultIgnoreMatcher, error) {
	rules, err := NormalizeVaultIgnorePaths(paths)
	if err != nil {
		return VaultIgnoreMatcher{}, err
	}
	return VaultIgnoreMatcher{rules: rules}, nil
}

func (m VaultIgnoreMatcher) Rules() []string {
	return append([]string(nil), m.rules...)
}

func (m VaultIgnoreMatcher) Matches(raw string) bool {
	clean, issue := NormalizeVaultRelativePath(raw)
	if issue != VaultPathOK {
		return false
	}
	for _, rule := range m.rules {
		if clean == rule || strings.HasPrefix(clean, rule+"/") {
			return true
		}
	}
	return false
}

func NormalizeOptionalVaultRelativePath(raw string) (string, VaultPathIssue) {
	if strings.TrimSpace(raw) == "" {
		return "", VaultPathOK
	}
	return NormalizeVaultRelativePath(raw)
}

func NormalizeOptionalVaultRelativePrefix(raw string) (string, VaultPathIssue) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", VaultPathOK
	}
	trailingSlash := strings.HasSuffix(slashVaultPath(trimmed), "/")
	clean, issue := NormalizeVaultRelativePath(trimmed)
	if issue != VaultPathOK {
		return "", issue
	}
	if trailingSlash && !strings.HasSuffix(clean, "/") {
		clean += "/"
	}
	return clean, VaultPathOK
}

func slashVaultPath(raw string) string {
	return strings.ReplaceAll(filepath.ToSlash(raw), `\`, "/")
}

func isWindowsAbsolutePath(path string) bool {
	return len(path) >= 3 &&
		unicode.IsLetter(rune(path[0])) &&
		path[1] == ':' &&
		path[2] == '/'
}
