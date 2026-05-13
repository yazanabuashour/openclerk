package runner_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestConfigTaskProfileInspectConfigureAndClear(t *testing.T) {
	t.Parallel()

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

	configured, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action: runner.ConfigTaskActionConfigureProfile,
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
		Action:  runner.ConfigTaskActionConfigureProfile,
		Profile: runner.AutonomyModes{AudienceMode: "boardroom"},
	})
	if err != nil {
		t.Fatalf("invalid configure profile: %v", err)
	}
	if !invalid.Rejected || !strings.Contains(invalid.RejectionReason, "profile.audience_mode") {
		t.Fatalf("invalid profile result = %+v", invalid)
	}

	cleared, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{Action: runner.ConfigTaskActionClearProfile})
	if err != nil {
		t.Fatalf("clear profile: %v", err)
	}
	if cleared.Rejected ||
		cleared.Profile.ApprovalMode != runner.ApprovalModeApproveWrite ||
		cleared.Profile.AudienceMode != runner.AudienceModeTechnical {
		t.Fatalf("cleared profile = %+v", cleared)
	}
}

func TestProfileDefaultsGateDocumentAndRetrievalWithRequestOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	if _, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action:  runner.ConfigTaskActionConfigureProfile,
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
