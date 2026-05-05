package runner

import (
	"path"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

func normalizeVaultRelativePath(raw string) string {
	clean, issue := domain.NormalizeOptionalVaultRelativePath(raw)
	if issue != domain.VaultPathOK {
		return strings.TrimSpace(raw)
	}
	return clean
}

func normalizeVaultRelativePrefix(raw string) string {
	clean, issue := domain.NormalizeOptionalVaultRelativePrefix(raw)
	if issue != domain.VaultPathOK {
		return strings.TrimSpace(raw)
	}
	return clean
}

func validateSourcePathHint(pathHint string) string {
	clean, rejection := validateRequiredVaultPath(pathHint, "source.path_hint is required", "source.path_hint must be relative to the vault root", "source.path_hint must stay inside the vault root")
	if rejection != "" {
		return rejection
	}
	if !strings.HasPrefix(clean, "sources/") || path.Ext(clean) != ".md" {
		return "source.path_hint must be a vault-relative sources/*.md path"
	}
	return ""
}

func validateVideoPathHint(pathHint string) string {
	clean, rejection := validateRequiredVaultPath(pathHint, "video.path_hint is required", "video.path_hint must be relative to the vault root", "video.path_hint must stay inside the vault root")
	if rejection != "" {
		return rejection
	}
	if !strings.HasPrefix(clean, "sources/") || path.Ext(clean) != ".md" {
		return "video.path_hint must be a vault-relative sources/*.md path"
	}
	return ""
}

func validateAssetPathHint(pathHint string) string {
	clean, rejection := validateRequiredVaultPath(pathHint, "source.asset_path_hint is required", "source.asset_path_hint must be relative to the vault root", "source.asset_path_hint must stay inside the vault root")
	if rejection != "" {
		return rejection
	}
	if !strings.HasPrefix(clean, "assets/") || path.Ext(clean) != ".pdf" {
		return "source.asset_path_hint must be a vault-relative assets/**/*.pdf path"
	}
	return ""
}

func validateVideoAssetPathHint(pathHint string) string {
	clean, rejection := validateRequiredVaultPath(pathHint, "video.asset_path_hint is required", "video.asset_path_hint must be relative to the vault root", "video.asset_path_hint must stay inside the vault root")
	if rejection != "" {
		return rejection
	}
	if !strings.HasPrefix(clean, "assets/") || path.Ext(clean) != ".json" {
		return "video.asset_path_hint must be a vault-relative assets/**/*.json path"
	}
	return ""
}

func validateRequiredVaultPath(raw string, requiredMessage string, relativeMessage string, stayInsideMessage string) (string, string) {
	clean, issue := domain.NormalizeVaultRelativePath(raw)
	switch issue {
	case domain.VaultPathOK:
		return clean, ""
	case domain.VaultPathMissing:
		return "", requiredMessage
	case domain.VaultPathAbsolute:
		return "", relativeMessage
	default:
		return "", stayInsideMessage
	}
}
