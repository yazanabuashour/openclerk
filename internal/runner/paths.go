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
	return validateVaultPathHint(pathHint, vaultPathHintRule{
		RequiredMessage:   "source.path_hint is required",
		RelativeMessage:   "source.path_hint must be relative to the vault root",
		StayInsideMessage: "source.path_hint must stay inside the vault root",
		Prefix:            "sources/",
		Extension:         ".md",
		ShapeMessage:      "source.path_hint must be a vault-relative sources/*.md path",
	})
}

func validateVideoPathHint(pathHint string) string {
	return validateVaultPathHint(pathHint, vaultPathHintRule{
		RequiredMessage:   "video.path_hint is required",
		RelativeMessage:   "video.path_hint must be relative to the vault root",
		StayInsideMessage: "video.path_hint must stay inside the vault root",
		Prefix:            "sources/",
		Extension:         ".md",
		ShapeMessage:      "video.path_hint must be a vault-relative sources/*.md path",
	})
}

func validateAssetPathHint(pathHint string) string {
	return validateVaultPathHint(pathHint, vaultPathHintRule{
		RequiredMessage:   "source.asset_path_hint is required",
		RelativeMessage:   "source.asset_path_hint must be relative to the vault root",
		StayInsideMessage: "source.asset_path_hint must stay inside the vault root",
		Prefix:            "assets/",
		Extension:         ".pdf",
		ShapeMessage:      "source.asset_path_hint must be a vault-relative assets/**/*.pdf path",
	})
}

func validateVideoAssetPathHint(pathHint string) string {
	return validateVaultPathHint(pathHint, vaultPathHintRule{
		RequiredMessage:   "video.asset_path_hint is required",
		RelativeMessage:   "video.asset_path_hint must be relative to the vault root",
		StayInsideMessage: "video.asset_path_hint must stay inside the vault root",
		Prefix:            "assets/",
		Extension:         ".json",
		ShapeMessage:      "video.asset_path_hint must be a vault-relative assets/**/*.json path",
	})
}

type vaultPathHintRule struct {
	RequiredMessage   string
	RelativeMessage   string
	StayInsideMessage string
	Prefix            string
	Extension         string
	ShapeMessage      string
}

func validateVaultPathHint(raw string, rule vaultPathHintRule) string {
	clean, rejection := validateRequiredVaultPath(raw, rule.RequiredMessage, rule.RelativeMessage, rule.StayInsideMessage)
	if rejection != "" {
		return rejection
	}
	if !strings.HasPrefix(clean, rule.Prefix) || path.Ext(clean) != rule.Extension {
		return rule.ShapeMessage
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
