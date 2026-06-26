package runner_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestConfigTaskProfileInspectConfigureAndClear(t *testing.T) {
	t.Setenv("OPENCLERK_GIT_CHECKPOINTS", "")

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}

	inspect, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionInspectConfig})
	if err != nil {
		t.Fatalf("inspect config: %v", err)
	}
	if inspect.Rejected ||
		inspect.Profile.ApprovalMode != runner.ApprovalModeApproveWrite ||
		inspect.Profile.DraftingMode != runner.DraftingModeSuggestFields ||
		inspect.Profile.WriteTargetMode != runner.WriteTargetModeCreateOrUpdate ||
		inspect.Profile.CitationMode != runner.CitationModeBalanced ||
		inspect.Profile.PrivacyMode != runner.PrivacyModeAllowPaths ||
		inspect.Profile.AudienceMode != runner.AudienceModeTechnical {
		t.Fatalf("default inspect = %+v", inspect)
	}
	if inspect.Storage == nil ||
		inspect.Storage.DatabasePath != config.DatabasePath ||
		inspect.Storage.VaultRoot != filepath.Join(filepath.Dir(config.DatabasePath), "vault") ||
		inspect.Storage.DatabaseSource != "flag" {
		t.Fatalf("storage summary = %+v", inspect.Storage)
	}
	for _, path := range []string{".git", ".stversions", ".openclerk", ".backups"} {
		if !containsString(inspect.Storage.VaultIgnorePaths, path) {
			t.Fatalf("storage ignore paths = %+v, missing %s", inspect.Storage.VaultIgnorePaths, path)
		}
	}
	if inspect.GitLifecycle == nil ||
		inspect.GitLifecycle.CheckpointPersistence != "unsupported" ||
		inspect.GitLifecycle.CheckpointEnabledForInvocation ||
		inspect.GitLifecycle.CheckpointEnablementSource != "none" {
		t.Fatalf("git lifecycle summary = %+v", inspect.GitLifecycle)
	}

	configured, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action: runner.ConfigTaskActionConfigureProfile,
		Autonomy: runner.AutonomyModes{
			ApprovalMode: runner.ApprovalModeApproveWrite,
		},
		Profile: runner.AutonomyModes{
			ApprovalMode:    runner.ApprovalModeAutonomousDisposable,
			DraftingMode:    runner.DraftingModeAutonomousFields,
			WriteTargetMode: runner.WriteTargetModeCreateAllowed,
			CitationMode:    runner.CitationModeStrict,
			PrivacyMode:     runner.PrivacyModeAllowSnippets,
			AudienceMode:    runner.AudienceModePlainLanguage,
		},
	})
	if err != nil {
		t.Fatalf("configure profile: %v", err)
	}
	if configured.Rejected ||
		configured.Profile.ApprovalMode != runner.ApprovalModeAutonomousDisposable ||
		configured.Profile.DraftingMode != runner.DraftingModeAutonomousFields ||
		configured.Profile.WriteTargetMode != runner.WriteTargetModeCreateAllowed ||
		configured.Profile.CitationMode != runner.CitationModeStrict ||
		configured.Profile.PrivacyMode != runner.PrivacyModeAllowSnippets ||
		configured.Profile.AudienceMode != runner.AudienceModePlainLanguage {
		t.Fatalf("configured profile = %+v", configured)
	}

	reloaded, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionInspectConfig})
	if err != nil {
		t.Fatalf("reload profile: %v", err)
	}
	if reloaded.Profile != configured.Profile {
		t.Fatalf("reloaded profile = %+v, want %+v", reloaded.Profile, configured.Profile)
	}

	invalid, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action: runner.ConfigTaskActionConfigureProfile,
		Autonomy: runner.AutonomyModes{
			ApprovalMode: runner.ApprovalModeApproveWrite,
		},
		Profile: runner.AutonomyModes{AudienceMode: "boardroom"},
	})
	if err != nil {
		t.Fatalf("invalid configure profile: %v", err)
	}
	if !invalid.Rejected || !strings.Contains(invalid.RejectionReason, "profile.audience_mode") {
		t.Fatalf("invalid profile result = %+v", invalid)
	}

	implicitApproval, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action: runner.ConfigTaskActionConfigureProfile,
		Autonomy: runner.AutonomyModes{
			PrivacyMode: runner.PrivacyModeAllowPaths,
		},
		Profile: runner.AutonomyModes{AudienceMode: runner.AudienceModeExecutiveSummary},
	})
	if err != nil {
		t.Fatalf("implicit approval configure profile: %v", err)
	}
	if !implicitApproval.Rejected || !strings.Contains(implicitApproval.RejectionReason, "explicit autonomy.approval_mode") {
		t.Fatalf("implicit approval profile result = %+v", implicitApproval)
	}

	unsupported, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{Action: "write_raw_runtime_config"})
	if err != nil {
		t.Fatalf("unsupported config action: %v", err)
	}
	if !unsupported.Rejected || !strings.Contains(unsupported.RejectionReason, "unsupported config action") {
		t.Fatalf("unsupported config result = %+v", unsupported)
	}

	cleared, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action: runner.ConfigTaskActionClearProfile,
		Autonomy: runner.AutonomyModes{
			ApprovalMode: runner.ApprovalModeApproveWrite,
		},
	})
	if err != nil {
		t.Fatalf("clear profile: %v", err)
	}
	if cleared.Rejected ||
		cleared.Profile.ApprovalMode != runner.ApprovalModeApproveWrite ||
		cleared.Profile.AudienceMode != runner.AudienceModeTechnical {
		t.Fatalf("cleared profile = %+v", cleared)
	}
}

func TestConfigTaskVaultIgnorePathsPersistInRuntimeConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	paths := []string{"scratch/", `private\drafts`}
	configured, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action:           runner.ConfigTaskActionConfigureVaultIgnores,
		Autonomy:         runner.AutonomyModes{ApprovalMode: runner.ApprovalModeApproveWrite},
		VaultIgnorePaths: &paths,
	})
	if err != nil {
		t.Fatalf("configure vault ignores: %v", err)
	}
	if configured.Storage == nil {
		t.Fatalf("configured storage = nil")
	}
	for _, path := range []string{".git", ".stversions", ".openclerk", ".backups", "scratch", "private/drafts"} {
		if !containsString(configured.Storage.VaultIgnorePaths, path) {
			t.Fatalf("configured ignore paths = %+v, missing %s", configured.Storage.VaultIgnorePaths, path)
		}
	}
	if !containsString(configured.Storage.CustomVaultIgnorePaths, "scratch") ||
		!containsString(configured.Storage.CustomVaultIgnorePaths, "private/drafts") {
		t.Fatalf("custom ignore paths = %+v", configured.Storage.CustomVaultIgnorePaths)
	}

	reloaded, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionInspectConfig})
	if err != nil {
		t.Fatalf("reload vault ignores: %v", err)
	}
	if reloaded.Storage == nil ||
		!containsString(reloaded.Storage.CustomVaultIgnorePaths, "scratch") ||
		!containsString(reloaded.Storage.CustomVaultIgnorePaths, "private/drafts") {
		t.Fatalf("reloaded storage = %+v", reloaded.Storage)
	}

	emptyPaths := []string{}
	cleared, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action:           runner.ConfigTaskActionConfigureVaultIgnores,
		Autonomy:         runner.AutonomyModes{ApprovalMode: runner.ApprovalModeApproveWrite},
		VaultIgnorePaths: &emptyPaths,
	})
	if err != nil {
		t.Fatalf("clear vault ignores: %v", err)
	}
	if cleared.Storage == nil || len(cleared.Storage.CustomVaultIgnorePaths) != 0 {
		t.Fatalf("cleared storage = %+v", cleared.Storage)
	}
	if !containsString(cleared.Storage.VaultIgnorePaths, ".git") {
		t.Fatalf("cleared effective ignore paths = %+v, want built-in defaults", cleared.Storage.VaultIgnorePaths)
	}

	missing, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionConfigureVaultIgnores})
	if err != nil {
		t.Fatalf("missing vault ignores config: %v", err)
	}
	if !missing.Rejected || !strings.Contains(missing.RejectionReason, "vault_ignore_paths") {
		t.Fatalf("missing vault ignores result = %+v", missing)
	}
}

func TestProfileDefaultsGateDocumentAndRetrievalWithRequestOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	if _, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action: runner.ConfigTaskActionConfigureProfile,
		Autonomy: runner.AutonomyModes{
			ApprovalMode: runner.ApprovalModeApproveWrite,
		},
		Profile: runner.AutonomyModes{ApprovalMode: runner.ApprovalModeProposeOnly},
	}); err != nil {
		t.Fatalf("configure profile: %v", err)
	}

	blockedDocument, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "notes/profile/propose-only.md",
			Title: "Propose Only",
			Body:  "# Propose Only\n",
		},
	})
	if err != nil {
		t.Fatalf("profile-gated document create: %v", err)
	}
	if !blockedDocument.Rejected || !strings.Contains(blockedDocument.RejectionReason, "propose_only") {
		t.Fatalf("blocked document = %+v", blockedDocument)
	}

	created, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:   runner.DocumentTaskActionCreate,
		Autonomy: runner.AutonomyModes{ApprovalMode: runner.ApprovalModeApproveWrite},
		Document: runner.DocumentInput{
			Path:  "notes/profile/override.md",
			Title: "Override",
			Body:  "# Override\n",
		},
	})
	if err != nil {
		t.Fatalf("override document create: %v", err)
	}
	if created.Rejected || created.Document == nil || created.Document.Path != "notes/profile/override.md" {
		t.Fatalf("created document = %+v", created)
	}

	blockedRetrieval, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSourceAuditReport,
		SourceAudit: runner.SourceAuditReportOptions{
			Query:      "profile repair",
			TargetPath: "synthesis/profile.md",
			Mode:       "repair_existing",
		},
	})
	if err != nil {
		t.Fatalf("profile-gated retrieval repair: %v", err)
	}
	if !blockedRetrieval.Rejected || !strings.Contains(blockedRetrieval.RejectionReason, "propose_only") {
		t.Fatalf("blocked retrieval = %+v", blockedRetrieval)
	}
}

func TestConfigInspectSummarizesModulesWithoutProviderSecrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	manifestPath := writeRunnerSemanticModuleManifest(t, t.TempDir(), "gemini")
	if _, err := runclient.InstallSemanticModule(ctx, config, runclient.SemanticModuleInstallInput{
		Provider:     "gemini",
		ManifestPath: manifestPath,
		Command:      "semantic-retrieval-adapter",
		ProviderConfig: map[string]string{
			"embedding_model": "gemini-embedding-001",
			"gemini_api_base": "https://generativelanguage.googleapis.com/v1beta",
			"api_key":         "do-not-leak",
		},
	}); err != nil {
		t.Fatalf("install module: %v", err)
	}

	inspect, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionInspectConfig})
	if err != nil {
		t.Fatalf("inspect config: %v", err)
	}
	if len(inspect.Modules) != 1 ||
		inspect.Modules[0].Provider != "gemini" ||
		inspect.Modules[0].Kind != runclient.ModuleKindEmbeddingProvider ||
		!inspect.Modules[0].Enabled ||
		!filepath.IsAbs(inspect.Modules[0].Command) ||
		filepath.Base(inspect.Modules[0].Command) != "semantic-retrieval-adapter" ||
		inspect.Modules[0].RedactionStatus != "redacted" {
		t.Fatalf("module summaries = %+v", inspect.Modules)
	}
	data, err := json.Marshal(inspect)
	if err != nil {
		t.Fatalf("marshal inspect: %v", err)
	}
	if strings.Contains(string(data), "do-not-leak") || strings.Contains(string(data), "provider_config") {
		t.Fatalf("inspect leaked provider config: %s", data)
	}
}

func TestConfigInspectReportsGitCheckpointInvocationGate(t *testing.T) {
	ctx := context.Background()

	flagConfig := runclient.Config{
		DatabasePath:   filepath.Join(t.TempDir(), "flag", "openclerk.sqlite"),
		GitCheckpoints: true,
	}
	flagInspect, err := runner.RunConfigTask(ctx, flagConfig, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionInspectConfig})
	if err != nil {
		t.Fatalf("inspect flag config: %v", err)
	}
	if flagInspect.GitLifecycle == nil ||
		!flagInspect.GitLifecycle.CheckpointEnabledForInvocation ||
		flagInspect.GitLifecycle.CheckpointEnablementSource != "flag" ||
		flagInspect.GitLifecycle.CheckpointPersistence != "unsupported" {
		t.Fatalf("flag git lifecycle = %+v", flagInspect.GitLifecycle)
	}

	t.Setenv("OPENCLERK_GIT_CHECKPOINTS", "yes")
	envConfig := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "env", "openclerk.sqlite")}
	envInspect, err := runner.RunConfigTask(ctx, envConfig, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionInspectConfig})
	if err != nil {
		t.Fatalf("inspect env config: %v", err)
	}
	if envInspect.GitLifecycle == nil ||
		!envInspect.GitLifecycle.CheckpointEnabledForInvocation ||
		envInspect.GitLifecycle.CheckpointEnablementSource != "env" {
		t.Fatalf("env git lifecycle = %+v", envInspect.GitLifecycle)
	}
}

func writeRunnerSemanticModuleManifest(t *testing.T, dir string, provider string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create manifest dir: %v", err)
	}
	path := filepath.Join(dir, "module.json")
	manifest := map[string]any{
		"schema_version": "openclerk-module.v1",
		"module": map[string]any{
			"name":    provider + "-embeddings",
			"version": "0.1.0",
			"kind":    "embedding_provider",
		},
		"provides": []map[string]any{{
			"type": "command",
			"name": "semantic-retrieval-adapter search",
		}},
		"authority": map[string]any{
			"default":        "read_only",
			"durable_writes": "forbidden",
			"forbidden":      []string{"write_documents", "change_openclerk_search_default"},
		},
		"release": map[string]any{
			"status": "supported_optional_module",
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}
