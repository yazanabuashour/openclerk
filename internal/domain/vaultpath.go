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
