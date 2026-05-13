package main

import (
	"context"
	"fmt"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func verifyProfileConfiguration(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	failures := documentHistoryInvariantFailures(turnMetrics)
	createdCount, err := exactDocumentCount(ctx, paths, profileConfigPath)
	if err != nil {
		return verificationResult{}, err
	}
	blockedCount, err := exactDocumentCount(ctx, paths, profileConfigBlockedPath)
	if err != nil {
		return verificationResult{}, err
	}
	repairCount, err := exactDocumentCount(ctx, paths, profileConfigRepairPath)
	if err != nil {
		return verificationResult{}, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	inspect, err := runner.RunConfigTask(ctx, cfg, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionInspectConfig})
	if err != nil {
		return verificationResult{}, err
	}
	if createdCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", profileConfigPath, createdCount))
	}
	if blockedCount != 0 {
		failures = append(failures, fmt.Sprintf("profile-gated blocked document was created at %s", profileConfigBlockedPath))
	}
	if repairCount != 0 {
		failures = append(failures, fmt.Sprintf("profile-gated repair document was created at %s", profileConfigRepairPath))
	}
	defaultsRestored := inspect.Profile.ApprovalMode == runner.ApprovalModeApproveWrite &&
		inspect.Profile.DraftingMode == runner.DraftingModeSuggestFields &&
		inspect.Profile.WriteTargetMode == runner.WriteTargetModeCreateOrUpdate &&
		inspect.Profile.CitationMode == runner.CitationModeBalanced &&
		inspect.Profile.PrivacyMode == runner.PrivacyModeAllowPaths &&
		inspect.Profile.AudienceMode == runner.AudienceModeTechnical
	if !defaultsRestored {
		failures = append(failures, fmt.Sprintf("profile defaults not restored after clear_profile: %+v", inspect.Profile))
	}
	requiredAnswer := []string{
		"openclerk config",
		"configure_profile",
		"inspect_config",
		"clear_profile",
		profileConfigPath,
		"approval_mode",
		"drafting_mode",
		"write_target_mode",
		"citation_mode",
		"privacy_mode",
		"audience_mode",
	}
	assistantPass := messageContainsAll(finalMessage, requiredAnswer) &&
		messageContainsAny(finalMessage, []string{"blocked", "rejected"}) &&
		messageContainsAny(finalMessage, []string{"override", "request-level autonomy", "request autonomy"})
	if !assistantPass {
		failures = append(failures, "final answer did not report profile configuration, gating, override, clear, and all six modes")
	}
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	databasePass := createdCount == 1 && blockedCount == 0 && repairCount == 0 && defaultsRestored
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{profileConfigPath},
	}, nil
}
