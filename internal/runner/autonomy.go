package runner

import (
	"fmt"
	"strings"
)

const (
	ApprovalModeProposeOnly          = "propose_only"
	ApprovalModeApproveWrite         = "approve_write"
	ApprovalModeAutonomousDisposable = "autonomous_disposable"
	ApprovalModeAutonomousTrusted    = "autonomous_trusted"

	DraftingModeRequireExplicitFields = "require_explicit_fields"
	DraftingModeSuggestFields         = "suggest_fields"
	DraftingModeAutonomousFields      = "autonomous_fields"

	WriteTargetModeExistingOnly   = "existing_only"
	WriteTargetModeCreateOrUpdate = "create_or_update"
	WriteTargetModeCreateAllowed  = "create_allowed"

	CitationModeStrict      = "strict"
	CitationModeBalanced    = "balanced"
	CitationModeLightweight = "lightweight"

	PrivacyModePrivateSummaryOnly = "private_summary_only"
	PrivacyModeAllowPaths         = "allow_paths"
	PrivacyModeAllowTitles        = "allow_titles"
	PrivacyModeAllowSnippets      = "allow_snippets"

	AudienceModeTechnical        = "technical"
	AudienceModePlainLanguage    = "plain_language"
	AudienceModeExecutiveSummary = "executive_summary"
)

func normalizeAutonomyModes(input AutonomyModes) (AutonomyModes, string) {
	normalized := AutonomyModes{
		ApprovalMode:    strings.TrimSpace(input.ApprovalMode),
		DraftingMode:    strings.TrimSpace(input.DraftingMode),
		WriteTargetMode: strings.TrimSpace(input.WriteTargetMode),
		CitationMode:    strings.TrimSpace(input.CitationMode),
		PrivacyMode:     strings.TrimSpace(input.PrivacyMode),
		AudienceMode:    strings.TrimSpace(input.AudienceMode),
	}
	if normalized.ApprovalMode == "" {
		normalized.ApprovalMode = ApprovalModeApproveWrite
	}
	if normalized.DraftingMode == "" {
		normalized.DraftingMode = DraftingModeSuggestFields
	}
	if normalized.WriteTargetMode == "" {
		normalized.WriteTargetMode = WriteTargetModeCreateOrUpdate
	}
	if normalized.CitationMode == "" {
		normalized.CitationMode = CitationModeBalanced
	}
	if normalized.PrivacyMode == "" {
		normalized.PrivacyMode = PrivacyModeAllowPaths
	}
	if normalized.AudienceMode == "" {
		normalized.AudienceMode = AudienceModeTechnical
	}
	checks := []struct {
		name    string
		value   string
		allowed []string
	}{
		{"approval_mode", normalized.ApprovalMode, []string{ApprovalModeProposeOnly, ApprovalModeApproveWrite, ApprovalModeAutonomousDisposable, ApprovalModeAutonomousTrusted}},
		{"drafting_mode", normalized.DraftingMode, []string{DraftingModeRequireExplicitFields, DraftingModeSuggestFields, DraftingModeAutonomousFields}},
		{"write_target_mode", normalized.WriteTargetMode, []string{WriteTargetModeExistingOnly, WriteTargetModeCreateOrUpdate, WriteTargetModeCreateAllowed}},
		{"citation_mode", normalized.CitationMode, []string{CitationModeStrict, CitationModeBalanced, CitationModeLightweight}},
		{"privacy_mode", normalized.PrivacyMode, []string{PrivacyModePrivateSummaryOnly, PrivacyModeAllowPaths, PrivacyModeAllowTitles, PrivacyModeAllowSnippets}},
		{"audience_mode", normalized.AudienceMode, []string{AudienceModeTechnical, AudienceModePlainLanguage, AudienceModeExecutiveSummary}},
	}
	for _, check := range checks {
		if !oneOf(check.value, check.allowed...) {
			return AutonomyModes{}, fmt.Sprintf("autonomy.%s must be one of %s", check.name, strings.Join(check.allowed, ", "))
		}
	}
	return normalized, ""
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
