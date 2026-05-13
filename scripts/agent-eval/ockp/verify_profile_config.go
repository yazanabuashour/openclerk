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
	modulePass := false
	for _, module := range inspect.Modules {
		if module.Provider == moduleAgentInstallProvider &&
			module.Kind == runclient.ModuleKindEmbeddingProvider &&
			module.ModuleName == moduleAgentInstallProvider+"-embeddings" &&
			module.Enabled &&
			module.RedactionStatus == "redacted" {
			modulePass = true
			break
		}
	}
	if !modulePass {
		failures = append(failures, fmt.Sprintf("inspect_config did not summarize seeded %s module: %+v", moduleAgentInstallProvider, inspect.Modules))
	}
	storagePass := inspect.Storage != nil &&
		inspect.Storage.DatabasePath == paths.DatabasePath &&
		inspect.Storage.DatabaseSource != "" &&
		inspect.Storage.VaultRoot != ""
	if !storagePass {
		failures = append(failures, fmt.Sprintf("inspect_config did not summarize storage: %+v", inspect.Storage))
	}
	gitLifecyclePass := inspect.GitLifecycle != nil &&
		inspect.GitLifecycle.CheckpointPersistence == "unsupported" &&
		inspect.GitLifecycle.CheckpointEnablementSource != ""
	if !gitLifecyclePass {
		failures = append(failures, fmt.Sprintf("inspect_config did not summarize git lifecycle checkpoint posture: %+v", inspect.GitLifecycle))
	}
	requiredAnswer := []string{
		"openclerk config",
		"configure_profile",
		"inspect_config",
		"clear_profile",
		"storage",
		"modules",
		moduleAgentInstallProvider,
		"git_lifecycle",
		"checkpoint_persistence",
		"unsupported",
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
	databasePass := createdCount == 1 && blockedCount == 0 && repairCount == 0 && defaultsRestored && modulePass && storagePass && gitLifecyclePass
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{profileConfigPath},
	}, nil
}
