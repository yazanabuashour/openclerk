package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func RunConfigTask(ctx context.Context, config runclient.Config, request ConfigTaskRequest) (ConfigTaskResult, error) {
	action := strings.TrimSpace(request.Action)
	if action == "" {
		action = ConfigTaskActionInspectConfig
	}
	paths, err := runclient.ResolvePaths(config)
	if err != nil {
		return ConfigTaskResult{}, err
	}
	convertedPaths := toPaths(paths)

	switch action {
	case ConfigTaskActionInspectConfig:
		profile, err := configuredProfileAutonomyModes(ctx, config)
		if err != nil {
			return ConfigTaskResult{}, err
		}
		return ConfigTaskResult{
			Paths:   &convertedPaths,
			Profile: profile,
			Summary: "inspected OpenClerk config",
		}, nil
	case ConfigTaskActionConfigureProfile:
		current, err := configuredProfileAutonomyModes(ctx, config)
		if err != nil {
			return ConfigTaskResult{}, err
		}
		profile, rejection := normalizeProfileAutonomyModes(mergeAutonomyModes(current, request.Profile))
		if rejection != "" {
			return rejectedConfig(convertedPaths, rejection), nil
		}
		if err := runclient.WriteDefaultProfileConfig(ctx, config, profileAutonomyValues(profile)); err != nil {
			return ConfigTaskResult{}, err
		}
		return ConfigTaskResult{
			Paths:   &convertedPaths,
			Profile: profile,
			Summary: "configured default profile",
		}, nil
	case ConfigTaskActionClearProfile:
		if err := runclient.ClearDefaultProfileConfig(ctx, config); err != nil {
			return ConfigTaskResult{}, err
		}
		profile, rejection := normalizeProfileAutonomyModes(AutonomyModes{})
		if rejection != "" {
			return ConfigTaskResult{}, fmt.Errorf("normalize default profile: %s", rejection)
		}
		return ConfigTaskResult{
			Paths:   &convertedPaths,
			Profile: profile,
			Summary: "cleared default profile",
		}, nil
	default:
		return rejectedConfig(convertedPaths, fmt.Sprintf("unsupported config action %q", action)), nil
	}
}

func applyProfileDefaults(ctx context.Context, config runclient.Config, input AutonomyModes) (AutonomyModes, error) {
	profile, err := configuredProfileAutonomyModes(ctx, config)
	if err != nil {
		return AutonomyModes{}, err
	}
	return mergeAutonomyModes(profile, input), nil
}

func configuredProfileAutonomyModes(ctx context.Context, config runclient.Config) (AutonomyModes, error) {
	values, err := runclient.ReadDefaultProfileConfig(ctx, config)
	if err != nil {
		return AutonomyModes{}, err
	}
	profile := AutonomyModes{
		ApprovalMode:    values["approval_mode"],
		DraftingMode:    values["drafting_mode"],
		WriteTargetMode: values["write_target_mode"],
		CitationMode:    values["citation_mode"],
		PrivacyMode:     values["privacy_mode"],
		AudienceMode:    values["audience_mode"],
	}
	normalized, rejection := normalizeProfileAutonomyModes(profile)
	if rejection != "" {
		return AutonomyModes{}, fmt.Errorf("%s", rejection)
	}
	return normalized, nil
}

func normalizeProfileAutonomyModes(input AutonomyModes) (AutonomyModes, string) {
	normalized, rejection := normalizeAutonomyModes(input)
	if rejection == "" {
		return normalized, ""
	}
	return AutonomyModes{}, strings.Replace(rejection, "autonomy.", "profile.", 1)
}

func mergeAutonomyModes(base AutonomyModes, override AutonomyModes) AutonomyModes {
	merged := base
	if strings.TrimSpace(override.ApprovalMode) != "" {
		merged.ApprovalMode = override.ApprovalMode
	}
	if strings.TrimSpace(override.DraftingMode) != "" {
		merged.DraftingMode = override.DraftingMode
	}
	if strings.TrimSpace(override.WriteTargetMode) != "" {
		merged.WriteTargetMode = override.WriteTargetMode
	}
	if strings.TrimSpace(override.CitationMode) != "" {
		merged.CitationMode = override.CitationMode
	}
	if strings.TrimSpace(override.PrivacyMode) != "" {
		merged.PrivacyMode = override.PrivacyMode
	}
	if strings.TrimSpace(override.AudienceMode) != "" {
		merged.AudienceMode = override.AudienceMode
	}
	return merged
}

func profileAutonomyValues(profile AutonomyModes) map[string]string {
	return map[string]string{
		"approval_mode":     profile.ApprovalMode,
		"drafting_mode":     profile.DraftingMode,
		"write_target_mode": profile.WriteTargetMode,
		"citation_mode":     profile.CitationMode,
		"privacy_mode":      profile.PrivacyMode,
		"audience_mode":     profile.AudienceMode,
	}
}

func rejectedConfig(paths Paths, reason string) ConfigTaskResult {
	return ConfigTaskResult{
		Rejected:        true,
		RejectionReason: reason,
		Paths:           &paths,
		Summary:         reason,
	}
}
