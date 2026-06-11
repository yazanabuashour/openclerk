package runner_test

import (
	"context"
	"encoding/json"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentTaskCreateListGetAndUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	create, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "notes/projects/roadmap.md",
			Title: "Roadmap",
			Body:  "# Roadmap\n\n## Summary\nCanonical project note.\n",
		},
	})
	if err != nil {
		t.Fatalf("create document task: %v", err)
	}
	if create.Rejected || create.Document == nil || create.Document.DocID == "" {
		t.Fatalf("create result = %+v", create)
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "notes/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list document task: %v", err)
	}
	if len(list.Documents) != 1 || list.Documents[0].Path != "notes/projects/roadmap.md" {
		t.Fatalf("list result = %+v", list)
	}

	appendResult, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionAppend,
		DocID:   create.Document.DocID,
		Content: "## Decisions\nUse the OpenClerk runner.\n",
	})
	if err != nil {
		t.Fatalf("append document task: %v", err)
	}
	if appendResult.Document == nil || !strings.Contains(appendResult.Document.Body, "OpenClerk runner") {
		t.Fatalf("append result = %+v", appendResult)
	}

	replace, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   create.Document.DocID,
		Heading: "Decisions",
		Content: "Use `openclerk` for routine agent work.",
	})
	if err != nil {
		t.Fatalf("replace document task: %v", err)
	}
	if replace.Document == nil ||
		!strings.Contains(replace.Document.Body, "openclerk") ||
		strings.Contains(replace.Document.Body, "OpenClerk runner") {
		t.Fatalf("replace result body = %q", replace.Document.Body)
	}

	cleared, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   create.Document.DocID,
		Heading: "Decisions",
		Content: "",
	})
	if err != nil {
		t.Fatalf("clear section task: %v", err)
	}
	if cleared.Document == nil ||
		!strings.Contains(cleared.Document.Body, "## Decisions") ||
		strings.Contains(cleared.Document.Body, "openclerk") {
		t.Fatalf("cleared section body = %q", cleared.Document.Body)
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  create.Document.DocID,
	})
	if err != nil {
		t.Fatalf("get document task: %v", err)
	}
	if get.Document == nil || get.Document.Path != create.Document.Path {
		t.Fatalf("get result = %+v", get)
	}

	if _, err := json.Marshal(get); err != nil {
		t.Fatalf("marshal document task result: %v", err)
	}
}

func TestDocumentTaskAutonomyModesValidateAndGateWrites(t *testing.T) {
	t.Parallel()

	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	invalid, err := runner.RunDocumentTask(context.Background(), config, runner.DocumentTaskRequest{
		Action:   runner.DocumentTaskActionValidate,
		Autonomy: runner.AutonomyModes{ApprovalMode: "surprise_me"},
	})
	if err != nil {
		t.Fatalf("invalid autonomy mode errored: %v", err)
	}
	if !invalid.Rejected || !strings.Contains(invalid.RejectionReason, "autonomy.approval_mode") {
		t.Fatalf("invalid autonomy rejection = %+v", invalid)
	}

	proposeOnly, err := runner.RunDocumentTask(context.Background(), config, runner.DocumentTaskRequest{
		Action:   runner.DocumentTaskActionCreate,
		Autonomy: runner.AutonomyModes{ApprovalMode: runner.ApprovalModeProposeOnly},
		Document: runner.DocumentInput{
			Path:  "notes/autonomy/propose-only.md",
			Title: "Propose Only",
			Body:  "# Propose Only\n",
		},
	})
	if err != nil {
		t.Fatalf("propose-only create errored: %v", err)
	}
	if !proposeOnly.Rejected || !strings.Contains(proposeOnly.RejectionReason, "propose_only") {
		t.Fatalf("propose-only rejection = %+v", proposeOnly)
	}

	existingOnly, err := runner.RunDocumentTask(context.Background(), config, runner.DocumentTaskRequest{
		Action:   runner.DocumentTaskActionCreate,
		Autonomy: runner.AutonomyModes{WriteTargetMode: runner.WriteTargetModeExistingOnly},
		Document: runner.DocumentInput{
			Path:  "notes/autonomy/existing-only.md",
			Title: "Existing Only",
			Body:  "# Existing Only\n",
		},
	})
	if err != nil {
		t.Fatalf("existing-only create errored: %v", err)
	}
	if !existingOnly.Rejected || !strings.Contains(existingOnly.RejectionReason, "existing_only") {
		t.Fatalf("existing-only rejection = %+v", existingOnly)
	}
}

func TestDocumentTaskMoveDocumentPlansAndAppliesMigration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	source := createDocument(t, ctx, config, "technology/projects.md", "Projects", "# Projects\n\n## Summary\nProject idea source.\n")
	index := createDocument(t, ctx, config, "technology/_index.md", "Technology Index", "# Technology\n\n- [Projects](projects.md)\n")
	backlog := createDocument(t, ctx, config, "projects/idea-backlog.md", "Idea Backlog", "# Idea Backlog\n\nSee [Projects](../technology/projects.md).\n")

	plan, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPlanMoveDocument,
		Move: runner.MoveDocumentOptions{
			Path:          "technology/projects.md",
			TargetPath:    "technology/project-ideas.md",
			UpdateIndexes: true,
		},
	})
	if err != nil {
		t.Fatalf("plan move document: %v", err)
	}
	if plan.Rejected ||
		plan.MovePlan == nil ||
		plan.MovePlan.DocID != source.DocID ||
		plan.MovePlan.WriteStatus != "no_write" ||
		!runnerMoveLinkUpdateIncludes(plan.MovePlan.LinkUpdates, backlog.Path, "../technology/project-ideas.md") ||
		!runnerMoveLinkUpdateIncludes(plan.MovePlan.LinkUpdates, index.Path, "project-ideas.md") {
		t.Fatalf("move plan result = %+v", plan)
	}

	proposeOnly, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:   runner.DocumentTaskActionMoveDocument,
		Autonomy: runner.AutonomyModes{ApprovalMode: runner.ApprovalModeProposeOnly},
		Move: runner.MoveDocumentOptions{
			Path:       "technology/projects.md",
			TargetPath: "technology/project-ideas.md",
		},
	})
	if err != nil {
		t.Fatalf("propose-only move document: %v", err)
	}
	if !proposeOnly.Rejected || !strings.Contains(proposeOnly.RejectionReason, "propose_only") {
		t.Fatalf("propose-only move rejection = %+v", proposeOnly)
	}

	moved, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionMoveDocument,
		Move: runner.MoveDocumentOptions{
			Path:          "technology/projects.md",
			TargetPath:    "technology/project-ideas.md",
			UpdateIndexes: true,
		},
	})
	if err != nil {
		t.Fatalf("move document: %v", err)
	}
	if moved.MoveResult == nil ||
		moved.MoveResult.Document.DocID != source.DocID ||
		moved.MoveResult.Document.Path != "technology/project-ideas.md" ||
		!strings.Contains(moved.MoveResult.Document.Body, "id: \""+source.DocID+"\"") ||
		!runnerMoveLinkUpdateIncludes(moved.MoveResult.LinkUpdatesApplied, backlog.Path, "../technology/project-ideas.md") {
		t.Fatalf("move result = %+v", moved)
	}
	updatedBacklog, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  backlog.DocID,
	})
	if err != nil {
		t.Fatalf("get backlog after move: %v", err)
	}
	if updatedBacklog.Document == nil ||
		!strings.Contains(updatedBacklog.Document.Body, "../technology/project-ideas.md") ||
		strings.Contains(updatedBacklog.Document.Body, "../technology/projects.md") {
		t.Fatalf("updated backlog = %+v", updatedBacklog.Document)
	}
}

func TestDocumentTaskRenameAndPromoteCandidateWrappers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	rough := createDocument(t, ctx, config, "notes/projects/rough-name.md", "Rough Name", "# Rough Name\n")
	candidate := createDocument(t, ctx, config, "notes/candidates/project-idea.md", "Project Idea", "# Project Idea\n")
	nonCandidate := createDocument(t, ctx, config, "notes/projects/not-candidate.md", "Not Candidate", "# Not Candidate\n")

	badRename, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionRenameDocument,
		Move: runner.MoveDocumentOptions{
			DocID:      rough.DocID,
			TargetPath: "notes/archive/rough-name.md",
		},
	})
	if err != nil {
		t.Fatalf("bad rename document: %v", err)
	}
	if !badRename.Rejected || !strings.Contains(badRename.RejectionReason, "same directory") {
		t.Fatalf("bad rename result = %+v", badRename)
	}

	renamed, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionRenameDocument,
		Move: runner.MoveDocumentOptions{
			DocID:      rough.DocID,
			TargetPath: "notes/projects/precise-name.md",
		},
	})
	if err != nil {
		t.Fatalf("rename document: %v", err)
	}
	if renamed.MoveResult == nil ||
		renamed.MoveResult.Document.DocID != rough.DocID ||
		renamed.MoveResult.Document.Path != "notes/projects/precise-name.md" {
		t.Fatalf("renamed result = %+v", renamed)
	}

	badPromote, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPromoteCandidate,
		Move: runner.MoveDocumentOptions{
			DocID:      nonCandidate.DocID,
			TargetPath: "notes/projects/promoted.md",
		},
	})
	if err != nil {
		t.Fatalf("bad promote candidate: %v", err)
	}
	if !badPromote.Rejected || !strings.Contains(badPromote.RejectionReason, "notes/candidates") {
		t.Fatalf("bad promote result = %+v", badPromote)
	}

	promoted, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPromoteCandidate,
		Move: runner.MoveDocumentOptions{
			DocID:      candidate.DocID,
			TargetPath: "notes/projects/project-idea.md",
		},
	})
	if err != nil {
		t.Fatalf("promote candidate: %v", err)
	}
	if promoted.MoveResult == nil ||
		promoted.MoveResult.Document.DocID != candidate.DocID ||
		promoted.MoveResult.Document.Path != "notes/projects/project-idea.md" {
		t.Fatalf("promoted result = %+v", promoted)
	}
}

func TestDocumentTaskPathCleanupPlansRenameAndCandidatePromotion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	rough := createDocument(t, ctx, config, "notes/projects/projects.md", "Project Ideas", "# Project Ideas\n")
	candidate := createDocument(t, ctx, config, "notes/candidates/project.md", "Great Project Idea", "# Great Project Idea\n")

	renamePlan, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPlanPathCleanup,
		PathCleanup: runner.PathCleanupOptions{
			DocID: rough.DocID,
		},
	})
	if err != nil {
		t.Fatalf("plan rename path cleanup: %v", err)
	}
	if renamePlan.PathCleanup == nil ||
		renamePlan.PathCleanup.WriteStatus != "planned_no_write" ||
		len(renamePlan.PathCleanup.Candidates) != 1 ||
		renamePlan.PathCleanup.Candidates[0].RecommendedAction != runner.DocumentTaskActionRenameDocument ||
		renamePlan.PathCleanup.Candidates[0].ProposedTargetPath != "notes/projects/project-ideas.md" ||
		renamePlan.PathCleanup.Candidates[0].DuplicateRisk != "none" ||
		!strings.Contains(renamePlan.PathCleanup.Candidates[0].NextRequest, "rename_document") {
		t.Fatalf("rename cleanup plan = %+v", renamePlan.PathCleanup)
	}

	promotionPlan, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPlanPathCleanup,
		PathCleanup: runner.PathCleanupOptions{
			Path:        candidate.Path,
			CleanupKind: "candidate_promotion",
		},
	})
	if err != nil {
		t.Fatalf("plan candidate promotion cleanup: %v", err)
	}
	if promotionPlan.PathCleanup == nil ||
		len(promotionPlan.PathCleanup.Candidates) != 1 ||
		promotionPlan.PathCleanup.Candidates[0].RecommendedAction != runner.DocumentTaskActionPromoteCandidate ||
		promotionPlan.PathCleanup.Candidates[0].ProposedTargetPath != "notes/projects/great-project-idea.md" ||
		!strings.Contains(promotionPlan.PathCleanup.Candidates[0].NextRequest, "promote_candidate") {
		t.Fatalf("promotion cleanup plan = %+v", promotionPlan.PathCleanup)
	}
}

func TestDocumentTaskPathCleanupReportsDuplicateRiskAndAppliesAutonomously(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	duplicateSource := createDocument(t, ctx, config, "notes/projects/projects.md", "Project Ideas", "# Project Ideas\n")
	if duplicateSource.DocID == "" {
		t.Fatalf("duplicate source missing doc id")
	}
	createDocument(t, ctx, config, "notes/projects/project-ideas.md", "Existing Project Ideas", "# Existing Project Ideas\n")
	applySource := createDocument(t, ctx, config, "notes/projects/rough.md", "Precise Project Name", "# Precise Project Name\n")

	duplicatePlan, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPlanPathCleanup,
		PathCleanup: runner.PathCleanupOptions{
			Path: "notes/projects/projects.md",
		},
	})
	if err != nil {
		t.Fatalf("plan duplicate cleanup: %v", err)
	}
	if duplicatePlan.PathCleanup == nil ||
		len(duplicatePlan.PathCleanup.Candidates) != 1 ||
		duplicatePlan.PathCleanup.Candidates[0].DuplicateRisk != "target_document_exists" ||
		duplicatePlan.PathCleanup.Candidates[0].NextRequest != "" {
		t.Fatalf("duplicate cleanup plan = %+v", duplicatePlan.PathCleanup)
	}

	blockedApply, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPlanPathCleanup,
		PathCleanup: runner.PathCleanupOptions{
			DocID: applySource.DocID,
			Mode:  "apply",
		},
	})
	if err != nil {
		t.Fatalf("blocked apply path cleanup: %v", err)
	}
	if !blockedApply.Rejected || !strings.Contains(blockedApply.RejectionReason, "autonomous_trusted") {
		t.Fatalf("blocked apply result = %+v", blockedApply)
	}

	applied, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPlanPathCleanup,
		Autonomy: runner.AutonomyModes{
			ApprovalMode: runner.ApprovalModeAutonomousTrusted,
		},
		PathCleanup: runner.PathCleanupOptions{
			DocID: applySource.DocID,
			Mode:  "apply",
		},
	})
	if err != nil {
		t.Fatalf("autonomous apply path cleanup: %v", err)
	}
	if applied.PathCleanup == nil ||
		applied.PathCleanup.AppliedCount != 1 ||
		applied.PathCleanup.Candidates[0].WriteStatus != "applied" ||
		applied.PathCleanup.Candidates[0].MoveResult == nil ||
		applied.PathCleanup.Candidates[0].MoveResult.Document.Path != "notes/projects/precise-project-name.md" {
		t.Fatalf("applied path cleanup = %+v", applied.PathCleanup)
	}

	persistedSource := createDocument(t, ctx, config, "notes/projects/temp-name.md", "Persisted Profile Name", "# Persisted Profile Name\n")
	if _, err := runner.RunConfigTask(ctx, config, runner.ConfigTaskRequest{
		Action:  runner.ConfigTaskActionConfigureProfile,
		Profile: runner.AutonomyModes{ApprovalMode: runner.ApprovalModeAutonomousTrusted},
	}); err != nil {
		t.Fatalf("configure autonomous profile: %v", err)
	}
	persistedApplied, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPlanPathCleanup,
		PathCleanup: runner.PathCleanupOptions{
			DocID: persistedSource.DocID,
			Mode:  "apply",
		},
	})
	if err != nil {
		t.Fatalf("persisted autonomous apply path cleanup: %v", err)
	}
	if persistedApplied.PathCleanup == nil ||
		persistedApplied.PathCleanup.AppliedCount != 1 ||
		persistedApplied.PathCleanup.Candidates[0].MoveResult == nil ||
		persistedApplied.PathCleanup.Candidates[0].MoveResult.Document.Path != "notes/projects/persisted-profile-name.md" {
		t.Fatalf("persisted applied path cleanup = %+v", persistedApplied.PathCleanup)
	}
}

func TestDocumentTaskPathCleanupSkipsTargetCollisionsBeforeApply(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	first := createDocument(t, ctx, config, "notes/projects/alpha.md", "Shared Cleanup Name", "# Shared Cleanup Name\n")
	second := createDocument(t, ctx, config, "notes/projects/beta.md", "Shared Cleanup Name", "# Shared Cleanup Name\n")

	applied, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionPlanPathCleanup,
		Autonomy: runner.AutonomyModes{
			ApprovalMode: runner.ApprovalModeAutonomousTrusted,
		},
		PathCleanup: runner.PathCleanupOptions{
			PathPrefix: "notes/projects/",
			Mode:       "apply",
		},
	})
	if err != nil {
		t.Fatalf("apply colliding path cleanup: %v", err)
	}
	if applied.PathCleanup == nil ||
		applied.PathCleanup.AppliedCount != 0 ||
		applied.PathCleanup.WriteStatus != "no_low_risk_candidates_applied" ||
		len(applied.PathCleanup.Candidates) != 2 {
		t.Fatalf("colliding path cleanup = %+v", applied.PathCleanup)
	}
	for _, candidate := range applied.PathCleanup.Candidates {
		if candidate.DuplicateRisk != "target_collision_in_plan" ||
			candidate.NextRequest != "" ||
			candidate.WriteStatus != "skipped_not_low_risk" ||
			!strings.Contains(candidate.Reason, "target_collision_in_plan") {
			t.Fatalf("colliding candidate = %+v", candidate)
		}
	}

	firstRead, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  first.DocID,
	})
	if err != nil {
		t.Fatalf("read first colliding source: %v", err)
	}
	secondRead, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  second.DocID,
	})
	if err != nil {
		t.Fatalf("read second colliding source: %v", err)
	}
	if firstRead.Document == nil || firstRead.Document.Path != first.Path ||
		secondRead.Document == nil || secondRead.Document.Path != second.Path {
		t.Fatalf("colliding sources moved unexpectedly: first=%+v second=%+v", firstRead.Document, secondRead.Document)
	}
}

func TestDocumentTaskGitLifecycleStatusAndHistory(t *testing.T) {
	t.Parallel()

	vaultRoot := initGitLifecycleTestRepo(t)
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	if _, err := runclient.InitializePaths(runclient.Config{DatabasePath: dbPath}, vaultRoot); err != nil {
		t.Fatalf("initialize paths: %v", err)
	}
	config := runclient.Config{DatabasePath: dbPath}
	ctx := context.Background()

	created := createDocument(t, ctx, config, "notes/git-lifecycle.md", "Git Lifecycle", "# Git Lifecycle\n\n## Summary\nLocal checkpoint evidence.\n")
	status, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGitLifecycle,
		GitLifecycle: runner.GitLifecycleOptions{
			Mode:  "status",
			Paths: []string{"notes/git-lifecycle.md"},
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("git lifecycle status: %v", err)
	}
	if status.GitLifecycle == nil ||
		status.GitLifecycle.WriteStatus != "no_write" ||
		status.GitLifecycle.AgentHandoff == nil ||
		!gitLifecycleDirtyPath(status.GitLifecycle.DirtyPaths, "notes/git-lifecycle.md") ||
		!strings.Contains(status.GitLifecycle.AuthorityLimits, "storage-level history only") {
		t.Fatalf("git lifecycle status result = %+v", status.GitLifecycle)
	}

	checkpoint, err := runner.RunDocumentTask(ctx, runclient.Config{DatabasePath: dbPath, GitCheckpoints: true}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGitLifecycle,
		GitLifecycle: runner.GitLifecycleOptions{
			Mode:    "checkpoint",
			Paths:   []string{"notes/git-lifecycle.md"},
			Message: "openclerk: checkpoint git lifecycle note",
		},
	})
	if err != nil {
		t.Fatalf("git lifecycle checkpoint: %v", err)
	}
	if checkpoint.GitLifecycle == nil ||
		checkpoint.GitLifecycle.CheckpointStatus != "created" ||
		checkpoint.GitLifecycle.WriteStatus != "checkpoint_created" ||
		checkpoint.GitLifecycle.CommitID == "" {
		t.Fatalf("git lifecycle checkpoint result = %+v", checkpoint.GitLifecycle)
	}

	history, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGitLifecycle,
		GitLifecycle: runner.GitLifecycleOptions{
			Mode:  "history",
			Paths: []string{"notes/git-lifecycle.md"},
			Limit: 5,
		},
	})
	if err != nil {
		t.Fatalf("git lifecycle history: %v", err)
	}
	if history.GitLifecycle == nil ||
		len(history.GitLifecycle.History) == 0 ||
		!strings.Contains(history.GitLifecycle.History[0].Summary, "checkpoint git lifecycle note") {
		t.Fatalf("git lifecycle history result = %+v", history.GitLifecycle)
	}

	if _, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionAppend,
		DocID:   created.DocID,
		Content: "## Update\nTracked worktree change.\n",
	}); err != nil {
		t.Fatalf("append tracked git lifecycle document: %v", err)
	}
	trackedStatus, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGitLifecycle,
		GitLifecycle: runner.GitLifecycleOptions{
			Mode:  "status",
			Paths: []string{"notes/git-lifecycle.md"},
		},
	})
	if err != nil {
		t.Fatalf("git lifecycle tracked status: %v", err)
	}
	if trackedStatus.GitLifecycle == nil ||
		!gitLifecycleDirtyPath(trackedStatus.GitLifecycle.DirtyPaths, "notes/git-lifecycle.md") {
		t.Fatalf("tracked git lifecycle status result = %+v", trackedStatus.GitLifecycle)
	}
}

func TestDocumentTaskGitLifecycleCheckpointRequiresConfig(t *testing.T) {
	t.Parallel()

	vaultRoot := initGitLifecycleTestRepo(t)
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	if _, err := runclient.InitializePaths(runclient.Config{DatabasePath: dbPath}, vaultRoot); err != nil {
		t.Fatalf("initialize paths: %v", err)
	}
	config := runclient.Config{DatabasePath: dbPath}
	ctx := context.Background()
	createDocument(t, ctx, config, "notes/no-checkpoint.md", "No Checkpoint", "# No Checkpoint\n\n## Summary\nUncheckpointed evidence.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGitLifecycle,
		GitLifecycle: runner.GitLifecycleOptions{
			Mode:    "checkpoint",
			Paths:   []string{"notes/no-checkpoint.md"},
			Message: "openclerk: should not checkpoint",
		},
	})
	if err != nil {
		t.Fatalf("git lifecycle checkpoint disabled: %v", err)
	}
	if result.GitLifecycle == nil ||
		result.GitLifecycle.CheckpointStatus != "disabled" ||
		result.GitLifecycle.WriteStatus != "rejected" {
		t.Fatalf("git lifecycle disabled result = %+v", result.GitLifecycle)
	}
}

func TestDocumentTaskGitLifecycleUnavailableForNestedVault(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for git lifecycle tests")
	}
	parent := t.TempDir()
	runGitLifecycleTestCommand(t, parent, "init")
	runGitLifecycleTestCommand(t, parent, "config", "user.name", "OpenClerk Test")
	runGitLifecycleTestCommand(t, parent, "config", "user.email", "openclerk@example.test")
	if err := os.WriteFile(filepath.Join(parent, "outside.md"), []byte("outside\n"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	runGitLifecycleTestCommand(t, parent, "add", "outside.md")
	runGitLifecycleTestCommand(t, parent, "commit", "-m", "outside")

	vaultRoot := filepath.Join(parent, "vault")
	if err := os.MkdirAll(filepath.Join(vaultRoot, "notes"), 0o755); err != nil {
		t.Fatalf("mkdir vault: %v", err)
	}
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	if _, err := runclient.InitializePaths(runclient.Config{DatabasePath: dbPath}, vaultRoot); err != nil {
		t.Fatalf("initialize paths: %v", err)
	}
	config := runclient.Config{DatabasePath: dbPath}
	ctx := context.Background()
	createDocument(t, ctx, config, "notes/nested.md", "Nested", "# Nested\n")

	status, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:       runner.DocumentTaskActionGitLifecycle,
		GitLifecycle: runner.GitLifecycleOptions{Mode: "status"},
	})
	if err != nil {
		t.Fatalf("git lifecycle nested status: %v", err)
	}
	if status.GitLifecycle == nil ||
		status.GitLifecycle.GitStatus != "unavailable" ||
		len(status.GitLifecycle.DirtyPaths) != 0 {
		t.Fatalf("nested status result = %+v", status.GitLifecycle)
	}

	history, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:       runner.DocumentTaskActionGitLifecycle,
		GitLifecycle: runner.GitLifecycleOptions{Mode: "history"},
	})
	if err != nil {
		t.Fatalf("git lifecycle nested history: %v", err)
	}
	if history.GitLifecycle == nil ||
		history.GitLifecycle.GitStatus != "unavailable" ||
		len(history.GitLifecycle.DirtyPaths) != 0 ||
		len(history.GitLifecycle.History) != 0 {
		t.Fatalf("nested history result = %+v", history.GitLifecycle)
	}
}

func TestDocumentTaskGitLifecycleRejectsPathspecMagic(t *testing.T) {
	t.Parallel()

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGitLifecycle,
		GitLifecycle: runner.GitLifecycleOptions{
			Mode:  "status",
			Paths: []string{":(top,glob)**"},
		},
	})
	if err != nil {
		t.Fatalf("git lifecycle pathspec validation: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "git_lifecycle.paths entries must be literal vault-relative paths" {
		t.Fatalf("pathspec validation result = %+v", result)
	}
}

func TestDocumentTaskWebSearchPlanReturnsPlacementHints(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/web/existing-example.md", "Existing Example", strings.TrimSpace(`---
source_url: https://example.test/existing
source_type: web
---
# Existing Example

## Summary
Existing public web source evidence.
`)+"\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionWebSearchPlan,
		WebSearch: runner.WebSearchPlanOptions{
			Query: "public source planning evidence",
			Results: []runner.WebSearchResultInput{
				{
					URL:     "https://example.test/new-report.pdf",
					Title:   "New Report",
					Snippet: "Public report search snippet.",
				},
				{
					URL:     "https://example.test/existing",
					Title:   "Existing Example",
					Snippet: "Duplicate source search snippet.",
				},
				{
					URL:          "https://example.test/private",
					Title:        "Private Example",
					Snippet:      "Login required snippet.",
					AccessStatus: "authenticated",
				},
			},
			Limit: 10,
		},
	})
	if err != nil {
		t.Fatalf("web search plan: %v", err)
	}
	plan := result.WebSearchPlan
	if plan == nil ||
		len(plan.Candidates) != 3 ||
		plan.FetchStatus != "planned_no_fetch" ||
		plan.WriteStatus != "planned_no_write" ||
		plan.AgentHandoff == nil ||
		!strings.Contains(plan.AuthorityLimits, "discovery hints only") {
		t.Fatalf("web search plan result = %+v", plan)
	}
	first := plan.Candidates[0]
	if first.Rank != 1 ||
		first.SourceType != "pdf" ||
		!containsString(first.CandidateSourcePaths, "sources/new-report.md") ||
		!containsString(first.CandidateAssetPaths, "assets/sources/new-report.pdf") ||
		!strings.Contains(first.NextIngestSourceRequest, "ingest_source_url") {
		t.Fatalf("first web search candidate = %+v", first)
	}
	second := plan.Candidates[1]
	if second.DuplicateStatus != "existing_source_url_found_no_fetch_no_write" ||
		second.ExistingSource == nil ||
		second.ExistingSource.Path != "sources/web/existing-example.md" ||
		second.CandidateSynthesisPath != "" ||
		!strings.Contains(second.NextIngestSourceRequest, `"mode":"update"`) {
		t.Fatalf("duplicate web search candidate = %+v", second)
	}
	third := plan.Candidates[2]
	if third.CandidateStatus != "unsupported_private_or_authenticated_no_fetch" ||
		third.AccessStatus != "authenticated" ||
		third.NextIngestSourceRequest != "" {
		t.Fatalf("authenticated web search candidate = %+v", third)
	}

	privateOnly, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionWebSearchPlan,
		WebSearch: runner.WebSearchPlanOptions{
			Query: "private only",
			Results: []runner.WebSearchResultInput{
				{URL: "https://example.test/private-only", Title: "Private Only", AccessStatus: "private"},
			},
		},
	})
	if err != nil {
		t.Fatalf("private-only web search plan: %v", err)
	}
	if privateOnly.WebSearchPlan == nil ||
		privateOnly.WebSearchPlan.AgentHandoff == nil ||
		!strings.Contains(privateOnly.WebSearchPlan.AgentHandoff.FollowUpPrimitiveInspection, "do not call ingest_source_url") {
		t.Fatalf("private-only web search plan = %+v", privateOnly.WebSearchPlan)
	}
}

func TestDocumentTaskWebSearchPlanRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionWebSearchPlan,
		WebSearch: runner.WebSearchPlanOptions{
			Query: "invalid source",
			Results: []runner.WebSearchResultInput{
				{URL: "file:///tmp/source.md"},
			},
		},
	})
	if err != nil {
		t.Fatalf("web search invalid input: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "web_search.results.url must be a valid http or https URL" {
		t.Fatalf("invalid URL result = %+v", result)
	}

	privateURL, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionWebSearchPlan,
		WebSearch: runner.WebSearchPlanOptions{
			Query: "private source",
			Results: []runner.WebSearchResultInput{
				{URL: "http://127.0.0.1/internal.pdf", SourceType: "pdf"},
			},
		},
	})
	if err != nil {
		t.Fatalf("web search private URL input: %v", err)
	}
	if !privateURL.Rejected || privateURL.RejectionReason != "web_search.results.url must be publicly fetchable" {
		t.Fatalf("private URL result = %+v", privateURL)
	}

	negative, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionWebSearchPlan,
		WebSearch: runner.WebSearchPlanOptions{
			Query: "negative limit",
			Results: []runner.WebSearchResultInput{
				{URL: "https://example.test/page.html"},
			},
			Limit: -1,
		},
	})
	if err != nil {
		t.Fatalf("web search negative limit: %v", err)
	}
	if !negative.Rejected || negative.RejectionReason != "limit must be greater than or equal to 0" {
		t.Fatalf("negative limit result = %+v", negative)
	}
}

func TestDocumentTaskArtifactCandidatePlanReturnsCandidate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			Content:      "# April Invoice\n\nVendor: Acme Services\nAmount due: 42 USD\n",
			ArtifactKind: "invoice",
			Fields: map[string]string{
				"vendor_hint": "Acme Services",
			},
		},
	})
	if err != nil {
		t.Fatalf("artifact candidate plan: %v", err)
	}
	plan := result.ArtifactPlan
	if plan == nil ||
		plan.WriteStatus != "planned_no_write" ||
		plan.FetchStatus != "planned_no_fetch" ||
		plan.ArtifactKind != "invoice" ||
		plan.SourceType != "explicit_content" ||
		plan.CandidatePath != "artifacts/invoices/april-invoice.md" ||
		plan.CandidateTitle != "April Invoice" ||
		plan.Confidence != "medium" ||
		plan.MetadataFields["vendor_hint"] != "Acme Services" ||
		!containsString(plan.Tags, "invoice") ||
		!strings.Contains(plan.BodyPreview, "# April Invoice") ||
		!strings.Contains(plan.NextCreateRequest, "create_document") ||
		plan.NextIngestSourceRequest != "" ||
		plan.AgentHandoff == nil ||
		!strings.Contains(plan.ValidationBoundaries, "no OCR") {
		t.Fatalf("artifact candidate plan result = %+v", plan)
	}
}

func TestDocumentTaskArtifactCandidatePlanPreservesOverridesAndDuplicateBoundary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "artifacts/receipts/existing-coffee-receipt.md", "Existing Coffee Receipt", strings.TrimSpace(`---
type: artifact
artifact_kind: receipt
tag: receipt
---
# Existing Coffee Receipt

Coffee receipt total paid by the support team.
`)+"\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			Content:        "Coffee receipt total paid by the support team.",
			ArtifactKind:   "receipt",
			Path:           "notes/manual/override.md",
			Title:          "Manual Override",
			Tags:           []string{"finance"},
			Fields:         map[string]string{"type": "note", "owner": "ap"},
			DuplicateQuery: "coffee receipt total paid",
			PathPrefix:     "artifacts/receipts/",
		},
	})
	if err != nil {
		t.Fatalf("artifact override duplicate plan: %v", err)
	}
	plan := result.ArtifactPlan
	if plan == nil ||
		plan.CandidatePath != "notes/manual/override.md" ||
		plan.CandidateTitle != "Manual Override" ||
		plan.MetadataFields["type"] != "note" ||
		plan.MetadataFields["owner"] != "ap" ||
		!containsString(plan.Tags, "finance") ||
		plan.LikelyDuplicate == nil ||
		plan.DuplicateStatus != "likely_duplicate_candidate_no_write" ||
		plan.NextCreateRequest != "" ||
		!strings.Contains(plan.AgentHandoff.FollowUpPrimitiveInspection, "update-versus-new") {
		t.Fatalf("artifact override duplicate plan result = %+v", plan)
	}
}

func TestDocumentTaskArtifactCandidatePlanPreservesPathPrefixDirectoryBoundary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "artifacts/receipts-old/decoy-coffee-receipt.md", "Decoy Coffee Receipt", strings.TrimSpace(`---
type: artifact
artifact_kind: receipt
tag: receipt
---
# Decoy Coffee Receipt

Coffee receipt total paid by the support team.
`)+"\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			Content:        "# Coffee Receipt\n\nCoffee receipt total paid by the support team.",
			ArtifactKind:   "receipt",
			DuplicateQuery: "coffee receipt total paid",
			PathPrefix:     "artifacts/receipts/",
		},
	})
	if err != nil {
		t.Fatalf("artifact path prefix boundary plan: %v", err)
	}
	plan := result.ArtifactPlan
	if plan == nil ||
		plan.LikelyDuplicate != nil ||
		plan.DuplicateStatus != "no_duplicate_found" ||
		plan.NextCreateRequest == "" {
		t.Fatalf("artifact path prefix boundary result = %+v", plan)
	}
}

func TestDocumentTaskArtifactCandidatePlanSourceURLHandoffDoesNotFetch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/web/existing-artifact-source.md", "Existing Artifact Source", strings.TrimSpace(`---
type: source
source_url: https://Example.test/artifact#section
source_type: web
---
# Existing Artifact Source

Existing source evidence.
`)+"\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			SourceURL:    "https://Example.test/artifact#section",
			ArtifactKind: "source_summary",
		},
	})
	if err != nil {
		t.Fatalf("artifact source URL plan: %v", err)
	}
	plan := result.ArtifactPlan
	if plan == nil ||
		plan.SourceURL != "https://example.test/artifact" ||
		plan.BodyPreview != "" ||
		plan.Confidence != "low" ||
		plan.ExistingSource == nil ||
		plan.ExistingSource.Path != "sources/web/existing-artifact-source.md" ||
		plan.DuplicateStatus != "existing_source_url_found_no_write" ||
		!strings.Contains(plan.NextIngestSourceRequest, `"mode":"update"`) ||
		plan.NextCreateRequest != "" ||
		!strings.Contains(plan.ApprovalBoundary, "public fetch") {
		t.Fatalf("artifact source URL plan result = %+v", plan)
	}
}

func TestDocumentTaskArtifactCandidatePlanRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			Content: "Opaque file placeholder",
			Path:    "../escape.md",
		},
	})
	if err != nil {
		t.Fatalf("artifact invalid path: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "artifact.path must stay inside the vault root" {
		t.Fatalf("invalid artifact path result = %+v", result)
	}

	missing, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action:   runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{ArtifactKind: "invoice"},
	})
	if err != nil {
		t.Fatalf("artifact missing content: %v", err)
	}
	if !missing.Rejected || missing.RejectionReason != "artifact.content, artifact.body, artifact.source_url, or artifact.local_path is required" {
		t.Fatalf("missing artifact content result = %+v", missing)
	}

	privateURL, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			SourceURL:  "http://169.254.169.254/latest/meta-data/report.pdf",
			SourceType: "pdf",
		},
	})
	if err != nil {
		t.Fatalf("artifact private source URL reject: %v", err)
	}
	if !privateURL.Rejected || privateURL.RejectionReason != "artifact.source_url must be publicly fetchable" {
		t.Fatalf("private artifact URL result = %+v", privateURL)
	}

	ocrProviderOnly, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			Content:     "Receipt text",
			OCRProvider: "tesseract",
		},
	})
	if err != nil {
		t.Fatalf("artifact OCR provider-only reject: %v", err)
	}
	if !ocrProviderOnly.Rejected || ocrProviderOnly.RejectionReason != "artifact.ocr_provider requires artifact.text_extraction ocr_review" {
		t.Fatalf("OCR provider-only result = %+v", ocrProviderOnly)
	}

	ocrReviewWithoutLocalPath, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			Content:        "Receipt text",
			TextExtraction: "ocr_review",
		},
	})
	if err != nil {
		t.Fatalf("artifact OCR review without local path reject: %v", err)
	}
	if !ocrReviewWithoutLocalPath.Rejected || ocrReviewWithoutLocalPath.RejectionReason != "artifact.local_path is required for artifact.text_extraction ocr_review" {
		t.Fatalf("OCR review without local path result = %+v", ocrReviewWithoutLocalPath)
	}
}

func TestDocumentTaskArtifactCandidatePlanReadsExplicitLocalTextArtifact(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	artifactPath := filepath.Join(t.TempDir(), "receipt.txt")
	if err := os.WriteFile(artifactPath, []byte("Coffee receipt\nTotal paid: 42 USD\n"), 0o644); err != nil {
		t.Fatalf("write local artifact: %v", err)
	}

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			LocalPath:    artifactPath,
			ArtifactKind: "receipt",
		},
	})
	if err != nil {
		t.Fatalf("artifact local plan: %v", err)
	}
	plan := result.ArtifactPlan
	if plan == nil ||
		plan.LocalArtifact == nil ||
		plan.LocalArtifact.SourceRef != "user_supplied_local_artifact" ||
		plan.LocalArtifact.TextStatus != "extracted" ||
		plan.WriteStatus != "planned_no_write" ||
		!strings.Contains(plan.BodyPreview, "Total paid: 42 USD") ||
		!strings.Contains(plan.ValidationBoundaries, "artifact.local_path") ||
		!strings.Contains(plan.NextCreateRequest, "create_document") {
		t.Fatalf("local artifact plan = %+v", plan)
	}
}

func TestDocumentTaskArtifactCandidatePlanReadsExplicitLocalPDFArtifact(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	artifactPath := filepath.Join(t.TempDir(), "artifact.pdf")
	if err := os.WriteFile(artifactPath, minimalPDF("Artifact PDF", "OpenClerk Test", "Parser promotion evidence text"), 0o644); err != nil {
		t.Fatalf("write local PDF: %v", err)
	}

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			LocalPath:    artifactPath,
			ArtifactKind: "source_summary",
		},
	})
	if err != nil {
		t.Fatalf("artifact local PDF plan: %v", err)
	}
	if result.ArtifactPlan == nil ||
		result.ArtifactPlan.LocalArtifact == nil ||
		result.ArtifactPlan.LocalArtifact.MIMEType != "application/pdf" ||
		result.ArtifactPlan.LocalArtifact.PageCount != 1 ||
		!strings.Contains(strings.ReplaceAll(result.ArtifactPlan.BodyPreview, " ", ""), "Parserpromotionevidencetext") {
		t.Fatalf("local PDF artifact plan = %+v", result.ArtifactPlan)
	}
}

func TestDocumentTaskArtifactCandidatePlanRejectsUnsupportedLocalOCRArtifact(t *testing.T) {
	t.Parallel()

	artifactPath := filepath.Join(t.TempDir(), "receipt.png")
	if err := os.WriteFile(artifactPath, []byte{0x89, 0x50, 0x4e, 0x47}, 0o644); err != nil {
		t.Fatalf("write local image: %v", err)
	}
	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			LocalPath: artifactPath,
		},
	})
	if err != nil {
		t.Fatalf("artifact unsupported local plan: %v", err)
	}
	if !result.Rejected || !strings.Contains(result.RejectionReason, "OCR/image parsing is unsupported") {
		t.Fatalf("unsupported local artifact result = %+v", result)
	}
}

func TestDocumentTaskArtifactCandidatePlanRejectsOCRReviewWithoutModule(t *testing.T) {
	t.Parallel()

	artifactPath := filepath.Join(t.TempDir(), "receipt.png")
	if err := os.WriteFile(artifactPath, []byte{0x89, 0x50, 0x4e, 0x47}, 0o644); err != nil {
		t.Fatalf("write local image: %v", err)
	}
	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			LocalPath:      artifactPath,
			TextExtraction: "ocr_review",
			OCRProvider:    "tesseract",
		},
	})
	if err != nil {
		t.Fatalf("artifact OCR review without module: %v", err)
	}
	if !result.Rejected || !strings.Contains(result.RejectionReason, "OCR module is not installed") {
		t.Fatalf("OCR review without module result = %+v", result)
	}
}

func TestDocumentTaskArtifactCandidatePlanOCRReviewLocalImageAndScannedPDF(t *testing.T) {
	t.Parallel()
	requireOCRFixtureTools(t)

	ctx := context.Background()
	manifestRoot := t.TempDir()
	manifestDir := filepath.Join(manifestRoot, "modules", "tesseract-ocr")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("create OCR manifest dir: %v", err)
	}
	config := runclient.Config{
		DatabasePath:       filepath.Join(t.TempDir(), "data", "openclerk.sqlite"),
		ModuleManifestRoot: manifestRoot,
	}
	manifestPath := writeRunnerOCRModuleManifest(t, manifestDir)
	manifestRelPath, err := filepath.Rel(manifestRoot, manifestPath)
	if err != nil {
		t.Fatalf("rel OCR manifest path: %v", err)
	}
	if _, err := runclient.InstallOCRModule(ctx, config, runclient.SemanticModuleInstallInput{
		Kind:         runclient.ModuleKindOCRProvider,
		Provider:     runclient.OCRModuleProviderTesseract,
		ManifestPath: filepath.ToSlash(manifestRelPath),
		Command:      "tesseract",
		ProviderConfig: map[string]string{
			"ocrmypdf_command": "ocrmypdf",
			"language":         "eng",
		},
	}); err != nil {
		t.Fatalf("install OCR module: %v", err)
	}
	fixtureDir := t.TempDir()
	imagePath, pdfPath := writeOCRFixtureArtifacts(t, fixtureDir)

	imageResult, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			LocalPath:      imagePath,
			ArtifactKind:   "receipt",
			TextExtraction: "ocr_review",
			OCRProvider:    "tesseract",
		},
	})
	if err != nil {
		t.Fatalf("image OCR plan: %v", err)
	}
	assertOCRPlan(t, imageResult, "OC-7781")

	pdfResult, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			LocalPath:      pdfPath,
			ArtifactKind:   "receipt",
			TextExtraction: "ocr_review",
			OCRProvider:    "tesseract",
		},
	})
	if err != nil {
		t.Fatalf("PDF OCR plan: %v", err)
	}
	assertOCRPlan(t, pdfResult, "Total paid 42 USD")
}

func TestDocumentTaskArtifactCandidatePlanRejectsNonRegularLocalArtifact(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionArtifactPlan,
		Artifact: runner.ArtifactPlanOptions{
			LocalPath:    os.DevNull,
			ArtifactKind: "receipt",
			Limit:        5,
		},
	})
	if err != nil {
		t.Fatalf("artifact plan non-regular local file: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "artifact.local_path must be a regular file" {
		t.Fatalf("non-regular local artifact result = %+v", result)
	}
}

func TestDocumentTaskCompileSynthesisCreatesAndUpdatesOneTarget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/synthesis-a.md", "Synthesis source A", "# Synthesis source A\n\n## Summary\nCurrent synthesis workflow evidence A.\n")
	createDocument(t, ctx, config, "sources/synthesis-b.md", "Synthesis source B", "# Synthesis source B\n\n## Summary\nCurrent synthesis workflow evidence B.\n")

	body := strings.TrimSpace(`# Workflow Synthesis

## Summary
Initial workflow synthesis.

## Sources
- sources/synthesis-a.md
- sources/synthesis-b.md

## Freshness
Checked current source evidence.
`) + "\n"
	created, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:       "synthesis/workflow.md",
			Title:      "Workflow Synthesis",
			SourceRefs: []string{"sources/synthesis-a.md", "sources/synthesis-b.md"},
			Body:       body,
			Mode:       "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis create: %v", err)
	}
	if created.Rejected || created.CompileSynthesis == nil {
		t.Fatalf("compile synthesis create result = %+v", created)
	}
	if created.CompileSynthesis.WriteStatus != "created" ||
		created.CompileSynthesis.DuplicateStatus != "no_duplicate_created" ||
		len(created.CompileSynthesis.SourceEvidence) != 2 ||
		len(created.CompileSynthesis.ProjectionFreshness) == 0 ||
		created.CompileSynthesis.AgentHandoff == nil ||
		!strings.Contains(created.CompileSynthesis.AgentHandoff.AnswerSummary, "compile_synthesis created synthesis/workflow.md") ||
		!strings.Contains(created.CompileSynthesis.AgentHandoff.FollowUpPrimitiveInspection, "not required") ||
		!strings.Contains(created.CompileSynthesis.ValidationBoundaries, "no broad repo search") {
		t.Fatalf("compile synthesis create report = %+v", created.CompileSynthesis)
	}

	updateBody := strings.Replace(body, "Initial workflow synthesis.", "Updated workflow synthesis.", 1)
	updated, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:       "synthesis/workflow.md",
			Title:      "Workflow Synthesis",
			SourceRefs: []string{"sources/synthesis-a.md", "sources/synthesis-b.md"},
			Body:       updateBody,
			Mode:       "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis update: %v", err)
	}
	if updated.CompileSynthesis == nil ||
		!updated.CompileSynthesis.ExistingCandidate ||
		updated.CompileSynthesis.WriteStatus != "updated" ||
		updated.CompileSynthesis.DocumentID != created.CompileSynthesis.DocumentID {
		t.Fatalf("compile synthesis update report = %+v", updated.CompileSynthesis)
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "synthesis/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list synthesis docs: %v", err)
	}
	if len(list.Documents) != 1 {
		t.Fatalf("synthesis docs = %+v, want one target", list.Documents)
	}
	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  created.CompileSynthesis.DocumentID,
	})
	if err != nil {
		t.Fatalf("get synthesis: %v", err)
	}
	if get.Document == nil ||
		!strings.Contains(get.Document.Body, "source_refs: sources/synthesis-a.md, sources/synthesis-b.md") ||
		!strings.Contains(get.Document.Body, "Updated workflow synthesis.") {
		t.Fatalf("compiled synthesis body = %q", get.Document.Body)
	}
}

func TestDocumentTaskCompileSynthesisBuildsBodyFromFacts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/fact-current.md", "Fact source current", "# Fact source current\n\n## Summary\nCurrent fact source.\n")
	createDocument(t, ctx, config, "sources/fact-old.md", "Fact source old", "# Fact source old\n\n## Summary\nSuperseded fact source.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:          "synthesis/fact-built.md",
			Title:         "Fact Built",
			SourceRefs:    []string{"sources/fact-current.md", "sources/fact-old.md"},
			BodyFacts:     []string{"Current source: sources/fact-current.md", "Superseded source: sources/fact-old.md"},
			FreshnessNote: "Checked through runner-owned body assembly.",
			Mode:          "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis body facts: %v", err)
	}
	if result.Rejected || result.CompileSynthesis == nil || result.CompileSynthesis.AgentHandoff == nil {
		t.Fatalf("compile synthesis body facts result = %+v", result)
	}
	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  result.CompileSynthesis.DocumentID,
	})
	if err != nil {
		t.Fatalf("get fact-built synthesis: %v", err)
	}
	for _, want := range []string{
		"source_refs: sources/fact-current.md, sources/fact-old.md",
		"## Summary",
		"- Current source: sources/fact-current.md",
		"## Sources",
		"- sources/fact-current.md",
		"## Freshness",
		"Checked through runner-owned body assembly.",
	} {
		if get.Document == nil || !strings.Contains(get.Document.Body, want) {
			t.Fatalf("fact-built body missing %q:\n%s", want, get.Document.Body)
		}
	}
}

func TestDocumentTaskCompileSynthesisAddsSourceRoleFacts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/role-current.md", "Role current", "---\nsupersedes: sources/role-old.md\n---\n# Role current\n\n## Summary\nCurrent compile_synthesis revisit guidance says promoted compile_synthesis handles routine synthesis refresh.\n")
	createDocument(t, ctx, config, "sources/role-old.md", "Role old", "---\nstatus: superseded\nsuperseded_by: sources/role-current.md\n---\n# Role old\n\n## Summary\nOld source.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:       "synthesis/role-built.md",
			Title:      "Role Built",
			SourceRefs: []string{"sources/role-current.md", "sources/role-old.md"},
			BodyFacts:  []string{"Role-aware compile_synthesis body."},
			Mode:       "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis role facts: %v", err)
	}
	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  result.CompileSynthesis.DocumentID,
	})
	if err != nil {
		t.Fatalf("get role-built synthesis: %v", err)
	}
	for _, want := range []string{
		"- Current source: sources/role-current.md",
		"- Current compile_synthesis revisit decision: promoted compile_synthesis handles routine synthesis refresh.",
		"- Superseded source: sources/role-old.md",
	} {
		if get.Document == nil || !strings.Contains(get.Document.Body, want) {
			t.Fatalf("role-built body missing %q:\n%s", want, get.Document.Body)
		}
	}
}

func TestDocumentTaskCompileSynthesisRejectsMissingFields(t *testing.T) {
	t.Parallel()

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:  "synthesis/missing.md",
			Title: "Missing",
			Body:  "# Missing\n\n## Summary\nMissing source refs.\n",
			Mode:  "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "synthesis.source_refs is required" {
		t.Fatalf("compile synthesis rejection = %+v", result)
	}
}

func TestDocumentTaskIngestSourceURLPlanModeIsReadOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        "https://Example.Test/product/page.html#section",
			Mode:       "plan",
			SourceType: "web",
			Title:      "Runner Product Page",
		},
	})
	if err != nil {
		t.Fatalf("source placement plan: %v", err)
	}
	if result.Rejected || result.SourcePlacement == nil {
		t.Fatalf("placement plan result = %+v", result)
	}
	plan := result.SourcePlacement
	if plan.SourceURL != "https://example.test/product/page.html" ||
		plan.SourceType != "web" ||
		plan.DuplicateStatus != "no_existing_source_url_found" ||
		plan.FetchStatus != "planned_no_fetch" ||
		plan.WriteStatus != "planned_no_write" ||
		plan.AgentHandoff == nil ||
		!strings.Contains(plan.ApprovalBoundary, "durable-write approval") ||
		!containsString(plan.CandidateSourcePaths, "sources/web/runner-product-page.md") ||
		plan.CandidateSynthesisPath != "synthesis/runner-product-page.md" ||
		!strings.Contains(plan.ValidationBoundaries, "no fetch") {
		t.Fatalf("placement plan = %+v", plan)
	}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: "sources/", Limit: 10},
	})
	if err != nil {
		t.Fatalf("list sources after plan: %v", err)
	}
	if len(list.Documents) != 0 {
		t.Fatalf("plan mode wrote source documents: %+v", list.Documents)
	}
}

func TestDocumentTaskIngestSourceURLPlanRejectsPrivateURL(t *testing.T) {
	t.Parallel()

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        "http://127.0.0.1/internal.pdf",
			Mode:       "plan",
			SourceType: "pdf",
		},
	})
	if err != nil {
		t.Fatalf("private source placement plan reject: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "source.url must be publicly fetchable" {
		t.Fatalf("private source placement result = %+v", result)
	}
}

func TestDocumentTaskIngestSourceURLNormalizesVaultPathSeparators(t *testing.T) {
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "runner-product.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Runner Web Title</title></head><body><h1>Runner Web Title</h1><p>Runner evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        "http://openclerk-eval.local/web/runner-product.html",
			PathHint:   `sources\web\runner-product.md`,
			SourceType: "web",
		},
	})
	if err != nil {
		t.Fatalf("ingest source URL with backslash path hints: %v", err)
	}
	if result.Rejected || result.Ingestion == nil || result.Ingestion.SourcePath != "sources/web/runner-product.md" {
		t.Fatalf("ingestion result = %+v", result)
	}
}

func TestDocumentTaskIngestSourceURLPlanModeReportsExistingSource(t *testing.T) {
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "existing.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Existing Source</title></head><body><h1>Existing Source</h1><p>Existing source evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	sourceURL := "http://openclerk-eval.local/web/existing.html"
	created, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			PathHint:   "sources/web/existing.md",
			SourceType: "web",
		},
	})
	if err != nil {
		t.Fatalf("create web source: %v", err)
	}
	plan, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			Mode:       "plan",
			SourceType: "web",
		},
	})
	if err != nil {
		t.Fatalf("source placement duplicate plan: %v", err)
	}
	if plan.SourcePlacement == nil ||
		plan.SourcePlacement.ExistingSource == nil ||
		plan.SourcePlacement.ExistingSource.DocID != created.Ingestion.DocID ||
		plan.SourcePlacement.DuplicateStatus != "existing_source_url_found_no_fetch_no_write" ||
		plan.SourcePlacement.CandidateSynthesisPath != "" ||
		!strings.Contains(plan.SourcePlacement.AgentHandoff.AnswerSummary, "no fetch or write occurred") {
		t.Fatalf("duplicate placement plan = %+v", plan.SourcePlacement)
	}
}

func TestDocumentTaskCompileSynthesisAssemblesPlainBodyAndAliases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/plain-current.md", "Plain source", "# Plain source\n\n## Summary\nPlain body source.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action:        runner.DocumentTaskActionCompileSynthesis,
		Document:      runner.DocumentInput{Path: "synthesis/plain-body", Title: "Plain Body"},
		Body:          "Plain synthesis summary from the user.",
		SourceRefs:    []string{"sources/plain-current.md"},
		FreshnessNote: "Plain body wrapped by compile_synthesis.",
	})
	if err != nil {
		t.Fatalf("compile synthesis plain body aliases: %v", err)
	}
	if result.Rejected || result.CompileSynthesis == nil {
		t.Fatalf("compile synthesis plain body result = %+v", result)
	}
	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  result.CompileSynthesis.DocumentID,
	})
	if err != nil {
		t.Fatalf("get plain-body synthesis: %v", err)
	}
	for _, want := range []string{
		"source_refs: sources/plain-current.md",
		"# Plain Body",
		"## Summary",
		"Plain synthesis summary from the user.",
		"## Sources",
		"- sources/plain-current.md",
		"## Freshness",
		"Plain body wrapped by compile_synthesis.",
	} {
		if get.Document == nil || !strings.Contains(get.Document.Body, want) {
			t.Fatalf("plain-body result missing %q:\n%s", want, get.Document.Body)
		}
	}
}

func TestDocumentTaskCompileSynthesisRejectsMissingBodyAndFactsWithoutWrite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCompileSynthesis,
		Synthesis: runner.CompileSynthesisInput{
			Path:       "synthesis/missing-body.md",
			Title:      "Missing Body",
			SourceRefs: []string{"sources/a.md"},
			Mode:       "create_or_update",
		},
	})
	if err != nil {
		t.Fatalf("compile synthesis missing body: %v", err)
	}
	if !result.Rejected || result.RejectionReason != "synthesis.body or synthesis.body_facts is required" {
		t.Fatalf("compile synthesis missing body rejection = %+v", result)
	}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after missing body reject: %v", err)
	}
	if len(list.Documents) != 0 {
		t.Fatalf("missing body rejection wrote documents: %+v", list.Documents)
	}
}

func TestDocumentTaskValidationSynthesisUsesDisposableDefaults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()
	config := runclient.Config{DatabasePath: filepath.Join(tempDir, "data", "openclerk.sqlite")}
	if _, err := runclient.InitializePaths(config, filepath.Join(tempDir, "task-01-synthesis_create_update", "private-vault-copy")); err != nil {
		t.Fatalf("initialize disposable paths: %v", err)
	}
	createDocument(t, ctx, config, "sources/routine-ux-validation/source.md", "Routine UX Validation Source", "# Routine UX Validation Source\n\n## Summary\nThis disposable source exists only inside the routine UX telemetry vault copy.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionValidationSynthesis,
		ValidationSynthesis: runner.ValidationSynthesisInput{
			DisposableValidation: true,
			BodyFacts:            []string{"Disposable validation synthesis evidence."},
		},
	})
	if err != nil {
		t.Fatalf("validation synthesis: %v", err)
	}
	if result.Rejected ||
		result.ValidationSynthesis == nil ||
		result.ValidationSynthesis.SelectedPath != "synthesis/routine-ux-validation.md" ||
		result.ValidationSynthesis.WriteStatus != "created" ||
		result.ValidationSynthesis.AgentHandoff == nil ||
		!strings.Contains(result.ValidationSynthesis.AgentHandoff.AnswerSummary, "validation_synthesis_report created") ||
		!strings.Contains(result.ValidationSynthesis.AgentHandoff.ValidationBoundaries, "no live private vault mutation") {
		t.Fatalf("validation synthesis result = %+v", result)
	}
}

func TestDocumentTaskValidationSynthesisRejectsLiveVault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "sources/routine-ux-validation/source.md", "Routine UX Validation Source", "# Routine UX Validation Source\n\n## Summary\nThis disposable source exists only inside the routine UX telemetry vault copy.\n")

	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionValidationSynthesis,
		ValidationSynthesis: runner.ValidationSynthesisInput{
			DisposableValidation: true,
			BodyFacts:            []string{"Should not write live vault."},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "routine UX disposable vault copy") {
		t.Fatalf("live vault validation synthesis result = %+v, err = %v", result, err)
	}
}

func TestDocumentTaskCompileSynthesisRejectsCleanedPathsOutsideNamespaces(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	body := strings.TrimSpace(`# Invalid

## Summary
Invalid synthesis.

## Sources
- sources/source.md

## Freshness
Checked.
`) + "\n"

	for _, tt := range []struct {
		name      string
		path      string
		sourceRef string
		want      string
	}{
		{
			name:      "target traversal",
			path:      "synthesis/../notes/escaped.md",
			sourceRef: "sources/source.md",
			want:      "synthesis.path must be under synthesis/",
		},
		{
			name:      "source ref traversal",
			path:      "synthesis/valid.md",
			sourceRef: "sources/../notes/source.md",
			want:      "synthesis.source_refs entries must be under sources/",
		},
		{
			name:      "source ref separator",
			path:      "synthesis/valid.md",
			sourceRef: "sources/source.md, sources/injected.md",
			want:      "synthesis.source_refs entries must be single vault-relative paths without separators",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionCompileSynthesis,
				Synthesis: runner.CompileSynthesisInput{
					Path:       tt.path,
					Title:      "Invalid",
					SourceRefs: []string{tt.sourceRef},
					Body:       body,
					Mode:       "create_or_update",
				},
			})
			if err != nil {
				t.Fatalf("compile synthesis reject: %v", err)
			}
			if !result.Rejected || result.RejectionReason != tt.want {
				t.Fatalf("compile synthesis rejection = %+v, want %q", result, tt.want)
			}
		})
	}

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 10},
	})
	if err != nil {
		t.Fatalf("list after rejected compile synthesis: %v", err)
	}
	if len(list.Documents) != 0 {
		t.Fatalf("rejected compile synthesis wrote documents: %+v", list.Documents)
	}
}

func TestDocumentTaskRejectsInvalidCreateFrontmatterBeforeRuntimeFiles(t *testing.T) {
	t.Parallel()

	dataDir := filepath.Join(t.TempDir(), "data")
	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(dataDir, "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "sources/uploaded-pdf.md",
			Title: "Uploaded PDF",
			Body: strings.TrimSpace(`---
type: source
modality: pdf
---
# Uploaded PDF

## Summary
Extracted note.
`) + "\n",
		},
	})
	if err != nil {
		t.Fatalf("document task: %v", err)
	}
	if !result.Rejected || !strings.Contains(result.RejectionReason, "modality") || !strings.Contains(result.RejectionReason, "markdown") {
		t.Fatalf("result = %+v, want modality rejection", result)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Fatalf("data dir exists after validation rejection: %v", err)
	}
}

func TestDocumentTaskAllowsMarkdownSourceWithPDFSourceType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "sources/uploaded-pdf.md",
			Title: "Uploaded PDF",
			Body: strings.TrimSpace(`---
type: source
source_type: pdf
modality: markdown
---
# Uploaded PDF

## Summary
Markdown notes extracted from a PDF source.
`) + "\n",
		},
	})
	if err != nil {
		t.Fatalf("document task: %v", err)
	}
	if result.Rejected || result.Document == nil {
		t.Fatalf("result = %+v, want created source document", result)
	}
	if result.Document.Metadata["source_type"] != "pdf" || result.Document.Metadata["modality"] != "markdown" {
		t.Fatalf("metadata = %+v", result.Document.Metadata)
	}
}

func TestDocumentTaskRejectsInvalidSourceURLIngestBeforeRuntimeFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  runner.SourceURLInput
		wantErr string
	}{
		{
			name: "missing url",
			source: runner.SourceURLInput{
				PathHint:      "sources/uploaded-pdf.md",
				AssetPathHint: "assets/sources/uploaded-pdf.pdf",
			},
			wantErr: "source.url is required",
		},
		{
			name: "invalid scheme",
			source: runner.SourceURLInput{
				URL:           "file:///tmp/uploaded-pdf.pdf",
				PathHint:      "sources/uploaded-pdf.md",
				AssetPathHint: "assets/sources/uploaded-pdf.pdf",
			},
			wantErr: "http or https",
		},
		{
			name: "unsafe source path",
			source: runner.SourceURLInput{
				URL:           "https://example.test/uploaded-pdf.pdf",
				PathHint:      "../uploaded-pdf.md",
				AssetPathHint: "assets/sources/uploaded-pdf.pdf",
			},
			wantErr: "source.path_hint",
		},
		{
			name: "unsafe asset path",
			source: runner.SourceURLInput{
				URL:           "https://example.test/uploaded-pdf.pdf",
				PathHint:      "sources/uploaded-pdf.md",
				AssetPathHint: "../uploaded-pdf.pdf",
			},
			wantErr: "source.asset_path_hint",
		},
		{
			name: "invalid mode",
			source: runner.SourceURLInput{
				URL:           "https://example.test/uploaded-pdf.pdf",
				PathHint:      "sources/uploaded-pdf.md",
				AssetPathHint: "assets/sources/uploaded-pdf.pdf",
				Mode:          "replace",
			},
			wantErr: "source.mode",
		},
		{
			name: "invalid source type",
			source: runner.SourceURLInput{
				URL:        "https://example.test/uploaded.html",
				PathHint:   "sources/uploaded-web.md",
				SourceType: "html",
			},
			wantErr: "source.source_type",
		},
		{
			name: "web source asset path",
			source: runner.SourceURLInput{
				URL:           "https://example.test/uploaded.html",
				PathHint:      "sources/uploaded-web.md",
				AssetPathHint: "assets/sources/uploaded-web.pdf",
				SourceType:    "web",
			},
			wantErr: "source.asset_path_hint",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dataDir := filepath.Join(t.TempDir(), "data")
			result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(dataDir, "openclerk.sqlite")}, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionIngestSourceURL,
				Source: tt.source,
			})
			if err != nil {
				t.Fatalf("document task: %v", err)
			}
			if !result.Rejected || !strings.Contains(result.RejectionReason, tt.wantErr) {
				t.Fatalf("result = %+v, want rejection containing %q", result, tt.wantErr)
			}
			if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
				t.Fatalf("data dir exists after validation rejection: %v", err)
			}
		})
	}
}

func TestDocumentTaskRejectsInvalidVideoURLIngestBeforeRuntimeFiles(t *testing.T) {
	t.Parallel()

	validTranscript := runner.VideoTranscriptInput{Text: "Supplied transcript evidence.", Policy: "supplied"}
	tests := []struct {
		name    string
		video   runner.VideoURLInput
		wantErr string
	}{
		{
			name: "missing url",
			video: runner.VideoURLInput{
				PathHint:   "sources/video-youtube/uploaded.md",
				Transcript: validTranscript,
			},
			wantErr: "video.url is required",
		},
		{
			name: "unsafe source path",
			video: runner.VideoURLInput{
				URL:        "https://www.youtube.com/watch?v=openclerk",
				PathHint:   "../uploaded.md",
				Transcript: validTranscript,
			},
			wantErr: "video.path_hint",
		},
		{
			name: "non source markdown path",
			video: runner.VideoURLInput{
				URL:        "https://www.youtube.com/watch?v=openclerk",
				PathHint:   "notes/uploaded.md",
				Transcript: validTranscript,
			},
			wantErr: "sources/*.md",
		},
		{
			name: "invalid mode",
			video: runner.VideoURLInput{
				URL:        "https://www.youtube.com/watch?v=openclerk",
				PathHint:   "sources/video-youtube/uploaded.md",
				Mode:       "replace",
				Transcript: validTranscript,
			},
			wantErr: "video.mode",
		},
		{
			name: "missing transcript",
			video: runner.VideoURLInput{
				URL:      "https://www.youtube.com/watch?v=openclerk",
				PathHint: "sources/video-youtube/uploaded.md",
			},
			wantErr: "video.transcript.text",
		},
		{
			name: "unsupported policy",
			video: runner.VideoURLInput{
				URL:      "https://www.youtube.com/watch?v=openclerk",
				PathHint: "sources/video-youtube/uploaded.md",
				Transcript: runner.VideoTranscriptInput{
					Text:   "Supplied transcript evidence.",
					Policy: "platform_caption",
				},
			},
			wantErr: "video.transcript.policy",
		},
		{
			name: "unsafe asset path",
			video: runner.VideoURLInput{
				URL:           "https://www.youtube.com/watch?v=openclerk",
				PathHint:      "sources/video-youtube/uploaded.md",
				AssetPathHint: "../uploaded.json",
				Transcript:    validTranscript,
			},
			wantErr: "video.asset_path_hint",
		},
		{
			name: "non json asset path",
			video: runner.VideoURLInput{
				URL:           "https://www.youtube.com/watch?v=openclerk",
				PathHint:      "sources/video-youtube/uploaded.md",
				AssetPathHint: "assets/video-youtube/uploaded.txt",
				Transcript:    validTranscript,
			},
			wantErr: "assets/**/*.json",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dataDir := filepath.Join(t.TempDir(), "data")
			result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(dataDir, "openclerk.sqlite")}, runner.DocumentTaskRequest{
				Action: runner.DocumentTaskActionIngestVideoURL,
				Video:  tt.video,
			})
			if err != nil {
				t.Fatalf("document task: %v", err)
			}
			if !result.Rejected || !strings.Contains(result.RejectionReason, tt.wantErr) {
				t.Fatalf("result = %+v, want rejection containing %q", result, tt.wantErr)
			}
			if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
				t.Fatalf("data dir exists after validation rejection: %v", err)
			}
		})
	}
}

func TestDocumentTaskIngestVideoURLAllowsPrivateSuppliedTranscriptURL(t *testing.T) {
	t.Parallel()

	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestVideoURL,
		Video: runner.VideoURLInput{
			URL:      "http://127.0.0.1/private-meeting",
			PathHint: "sources/video-youtube/private-meeting.md",
			Transcript: runner.VideoTranscriptInput{
				Text:   "Private meeting supplied transcript evidence.",
				Policy: "supplied",
			},
		},
	})
	if err != nil {
		t.Fatalf("ingest private supplied transcript URL: %v", err)
	}
	if result.Rejected || result.VideoIngestion == nil || result.VideoIngestion.SourceURL != "http://127.0.0.1/private-meeting" {
		t.Fatalf("result = %+v, want private supplied transcript URL accepted", result)
	}
}

func TestDocumentTaskIngestSourceURLUpdateStaleImpactResponse(t *testing.T) {
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "runner-product.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Runner Web Title</title></head><body><h1>Runner Web Title</h1><p>Initial runner evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	sourceURL := "http://openclerk-eval.local/web/runner-product.html"
	created, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:      sourceURL,
			PathHint: "sources/web/runner-product.md",
		},
	})
	if err != nil {
		t.Fatalf("create web source: %v", err)
	}
	if created.Ingestion == nil || created.Ingestion.UpdateStatus != "" {
		t.Fatalf("create ingestion = %+v", created.Ingestion)
	}
	createJSON, err := json.Marshal(created.Ingestion)
	if err != nil {
		t.Fatalf("marshal create ingestion: %v", err)
	}
	if strings.Contains(string(createJSON), "update_status") {
		t.Fatalf("create ingestion leaked update fields: %s", createJSON)
	}

	if _, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  "synthesis/web-runner.md",
			Title: "Web Runner Synthesis",
			Body:  "---\ntype: synthesis\nsource_refs: sources/web/runner-product.md\n---\n# Web Runner Synthesis\n\n## Summary\nInitial runner evidence.\n",
		},
	}); err != nil {
		t.Fatalf("create synthesis: %v", err)
	}

	same, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:  sourceURL,
			Mode: "update",
		},
	})
	if err != nil {
		t.Fatalf("same web update: %v", err)
	}
	if same.Ingestion == nil ||
		same.Ingestion.UpdateStatus != "no_op" ||
		same.Ingestion.NormalizedSourceURL != sourceURL ||
		same.Ingestion.SourceDocID != created.Ingestion.DocID ||
		same.Ingestion.PreviousSHA256 != created.Ingestion.SHA256 ||
		same.Ingestion.NewSHA256 != created.Ingestion.SHA256 ||
		same.Ingestion.Changed == nil || *same.Ingestion.Changed ||
		same.Ingestion.SynthesisRepaired == nil || *same.Ingestion.SynthesisRepaired ||
		same.Ingestion.StaleDependents == nil || len(*same.Ingestion.StaleDependents) != 0 ||
		same.Ingestion.ProjectionRefs == nil || len(*same.Ingestion.ProjectionRefs) != 0 ||
		same.Ingestion.ProvenanceRefs == nil || len(*same.Ingestion.ProvenanceRefs) != 0 {
		t.Fatalf("same update ingestion = %+v", same.Ingestion)
	}

	if err := os.WriteFile(fixturePath, []byte(`<!doctype html><html><head><title>Runner Web Title Updated</title></head><body><h1>Runner Web Title Updated</h1><p>Updated runner evidence.</p></body></html>`), 0o644); err != nil {
		t.Fatalf("write updated web fixture: %v", err)
	}
	changed, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			PathHint:   "sources/web/runner-product.md",
			SourceType: "web",
			Mode:       "update",
		},
	})
	if err != nil {
		t.Fatalf("changed web update: %v", err)
	}
	if changed.Ingestion == nil ||
		changed.Ingestion.UpdateStatus != "changed" ||
		changed.Ingestion.NormalizedSourceURL != sourceURL ||
		changed.Ingestion.SourceDocID != created.Ingestion.DocID ||
		changed.Ingestion.PreviousSHA256 != created.Ingestion.SHA256 ||
		changed.Ingestion.NewSHA256 == created.Ingestion.SHA256 ||
		changed.Ingestion.Changed == nil || !*changed.Ingestion.Changed ||
		changed.Ingestion.SynthesisRepaired == nil || *changed.Ingestion.SynthesisRepaired ||
		changed.Ingestion.StaleDependents == nil || len(*changed.Ingestion.StaleDependents) != 1 ||
		changed.Ingestion.ProjectionRefs == nil || len(*changed.Ingestion.ProjectionRefs) == 0 ||
		changed.Ingestion.ProvenanceRefs == nil || !runnerSourceProvenanceRefsInclude(*changed.Ingestion.ProvenanceRefs, "source_updated") ||
		!strings.Contains(changed.Ingestion.NoRepairWarning, "synthesis/web-runner.md") {
		t.Fatalf("changed update ingestion = %+v", changed.Ingestion)
	}
	updateJSON, err := json.Marshal(changed.Ingestion)
	if err != nil {
		t.Fatalf("marshal changed ingestion: %v", err)
	}
	for _, want := range []string{`"update_status":"changed"`, `"normalized_source_url":"` + sourceURL + `"`, `"source_doc_id":"`, `"previous_sha256":"`, `"new_sha256":"`, `"changed":true`, `"stale_dependents":[`, `"projection_refs":[`, `"provenance_refs":[`, `"synthesis_repaired":false`, `"no_repair_warning":"`} {
		if !strings.Contains(string(updateJSON), want) {
			t.Fatalf("changed ingestion JSON missing %s: %s", want, updateJSON)
		}
	}
}

func runnerSourceProvenanceRefsInclude(refs []runner.SourceProvenanceRef, eventType string) bool {
	for _, ref := range refs {
		if ref.EventType == eventType {
			return true
		}
	}
	return false
}

func TestDocumentTaskIngestSourceURLPDF(t *testing.T) {
	pdfBytes := minimalPDF("Runner Intake PDF Title", "OpenClerk Test", "Runner intake unique text")
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "pdf", "runner.pdf")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir PDF fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, pdfBytes, 0o644); err != nil {
		t.Fatalf("write PDF fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)
	sourceURL := "http://openclerk-eval.local/pdf/runner.pdf"

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           sourceURL,
			PathHint:      "sources/runner-ingest.md",
			AssetPathHint: "assets/sources/runner-ingest.pdf",
			Title:         "Runner Ingest Override",
		},
	})
	if err != nil {
		t.Fatalf("ingest source URL: %v", err)
	}
	if result.Rejected || result.Ingestion == nil {
		t.Fatalf("ingest result = %+v", result)
	}
	ingestion := result.Ingestion
	if ingestion.DocID == "" ||
		ingestion.SourcePath != "sources/runner-ingest.md" ||
		ingestion.AssetPath != "assets/sources/runner-ingest.pdf" ||
		ingestion.DerivedPath != "sources/runner-ingest.md" ||
		ingestion.PageCount != 1 ||
		ingestion.SizeBytes != int64(len(pdfBytes)) ||
		ingestion.MIMEType != "application/pdf" ||
		len(ingestion.Citations) == 0 ||
		len(ingestion.SHA256) != 64 {
		t.Fatalf("ingestion = %+v", ingestion)
	}
	if ingestion.PDFMetadata.Title != "Runner Intake PDF Title" || ingestion.PDFMetadata.Author != "OpenClerk Test" {
		t.Fatalf("pdf metadata = %+v", ingestion.PDFMetadata)
	}
	assetPath := filepath.Join(filepath.Dir(dbPath), "vault", "assets", "sources", "runner-ingest.pdf")
	if _, err := os.Stat(assetPath); err != nil {
		t.Fatalf("asset stat: %v", err)
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  ingestion.DocID,
	})
	if err != nil {
		t.Fatalf("get ingested document: %v", err)
	}
	if get.Document == nil ||
		get.Document.Metadata["source_url"] != sourceURL ||
		get.Document.Metadata["asset_path"] != "assets/sources/runner-ingest.pdf" ||
		get.Document.Metadata["source_type"] != "pdf" ||
		get.Document.Metadata["mime_type"] != "application/pdf" ||
		!strings.Contains(get.Document.Body, "Runner Ingest Override") ||
		!strings.Contains(get.Document.Body, "Runner Intake PDF Title") ||
		!strings.Contains(get.Document.Body, "Runnerintakeuniquetext") {
		t.Fatalf("ingested document = %+v", get.Document)
	}

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "Runner Intake PDF Title",
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("search ingested text: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || search.Search.Hits[0].Citations[0].Path != "sources/runner-ingest.md" {
		t.Fatalf("search = %+v", search.Search)
	}

	provenance, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "source",
			RefID:   ingestion.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("source provenance: %v", err)
	}
	if provenance.Provenance == nil || len(provenance.Provenance.Events) == 0 {
		t.Fatalf("provenance = %+v", provenance.Provenance)
	}
	if got := provenance.Provenance.Events[0].Details["source_url"]; got != sourceURL {
		t.Fatalf("source provenance details = %+v", provenance.Provenance.Events[0].Details)
	}

	sameUpdate, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:  sourceURL,
			Mode: "update",
		},
	})
	if err != nil {
		t.Fatalf("same source URL update: %v", err)
	}
	if sameUpdate.Rejected || sameUpdate.Ingestion == nil ||
		sameUpdate.Ingestion.DocID != ingestion.DocID ||
		sameUpdate.Ingestion.SourcePath != ingestion.SourcePath ||
		sameUpdate.Ingestion.AssetPath != ingestion.AssetPath ||
		sameUpdate.Ingestion.SHA256 != ingestion.SHA256 {
		t.Fatalf("same source URL update = %+v, want existing ingestion", sameUpdate)
	}

	duplicate, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           sourceURL,
			PathHint:      "sources/runner-duplicate.md",
			AssetPathHint: "assets/sources/runner-duplicate.pdf",
		},
	})
	if err == nil {
		t.Fatalf("duplicate result = %+v, want error", duplicate)
	}
	if !strings.Contains(err.Error(), "source URL") {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestDocumentTaskIngestSourceURLWeb(t *testing.T) {
	htmlBody := `<!doctype html>
<html>
<head><title>Runner Web Product</title><style>.hidden{display:none}</style></head>
<body>
<h1>Runner Web Product</h1>
<p>Visible web source evidence for OpenClerk.</p>
<script>doNotIndex()</script>
</body>
</html>`
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "runner-product.html")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir web fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(htmlBody), 0o644); err != nil {
		t.Fatalf("write web fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)
	sourceURL := "http://openclerk-eval.local/web/runner-product.html?ref=tracker"

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			PathHint:   "sources/web/runner-product.md",
			SourceType: "web",
		},
	})
	if err != nil {
		t.Fatalf("ingest web source URL: %v", err)
	}
	if result.Rejected || result.Ingestion == nil {
		t.Fatalf("ingest result = %+v", result)
	}
	ingestion := result.Ingestion
	if ingestion.DocID == "" ||
		ingestion.SourcePath != "sources/web/runner-product.md" ||
		ingestion.SourceURL != sourceURL ||
		ingestion.SourceType != "web" ||
		ingestion.AssetPath != "" ||
		ingestion.DerivedPath != "sources/web/runner-product.md" ||
		ingestion.PageCount != 0 ||
		ingestion.MIMEType != "text/html" ||
		len(ingestion.Citations) == 0 ||
		len(ingestion.SHA256) != 64 {
		t.Fatalf("ingestion = %+v", ingestion)
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  ingestion.DocID,
	})
	if err != nil {
		t.Fatalf("get ingested web document: %v", err)
	}
	if get.Document == nil ||
		get.Document.Metadata["source_url"] != sourceURL ||
		get.Document.Metadata["source_type"] != "web" ||
		get.Document.Metadata["asset_path"] != "" ||
		get.Document.Metadata["mime_type"] != "text/html" ||
		!strings.Contains(get.Document.Body, "Runner Web Product") ||
		!strings.Contains(get.Document.Body, "Visible web source evidence for OpenClerk.") ||
		strings.Contains(get.Document.Body, "doNotIndex") {
		t.Fatalf("ingested web document = %+v", get.Document)
	}

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "Visible web source evidence",
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("search ingested web text: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || search.Search.Hits[0].Citations[0].Path != "sources/web/runner-product.md" {
		t.Fatalf("search = %+v", search.Search)
	}

	provenance, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "source",
			RefID:   ingestion.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("source provenance: %v", err)
	}
	if provenance.Provenance == nil || len(provenance.Provenance.Events) == 0 {
		t.Fatalf("provenance = %+v", provenance.Provenance)
	}
	details := provenance.Provenance.Events[0].Details
	if details["source_url"] != sourceURL || details["source_type"] != "web" || details["asset_path"] != "" {
		t.Fatalf("source provenance details = %+v", details)
	}

	duplicate, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:      sourceURL,
			PathHint: "sources/web/runner-product-copy.md",
		},
	})
	if err == nil {
		t.Fatalf("duplicate result = %+v, want error", duplicate)
	}
	if !strings.Contains(err.Error(), "source URL") {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestDocumentTaskIngestSourceURLWebMarkdown(t *testing.T) {
	markdownBody := "# Runner Markdown README\n\nRunner Markdown source evidence for OpenClerk.\n\n## Install\n\nStore public README docs through ingest_source_url.\n"
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "README.md")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir markdown fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(markdownBody), 0o644); err != nil {
		t.Fatalf("write markdown fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)
	sourceURL := "http://openclerk-eval.local/web/README.md"

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			PathHint:   "sources/web/runner-markdown-readme.md",
			SourceType: "web",
		},
	})
	if err != nil {
		t.Fatalf("ingest markdown source URL: %v", err)
	}
	if result.Rejected || result.Ingestion == nil {
		t.Fatalf("ingest result = %+v", result)
	}
	if result.Ingestion.SourceType != "web" ||
		result.Ingestion.SourceURL != sourceURL ||
		result.Ingestion.MIMEType != "text/markdown" ||
		result.Ingestion.AssetPath != "" ||
		len(result.Ingestion.Citations) == 0 {
		t.Fatalf("markdown ingestion = %+v", result.Ingestion)
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  result.Ingestion.DocID,
	})
	if err != nil {
		t.Fatalf("get markdown source document: %v", err)
	}
	if get.Document == nil ||
		get.Document.Metadata["source_title"] != "Runner Markdown README" ||
		get.Document.Metadata["mime_type"] != "text/markdown" ||
		!strings.Contains(get.Document.Body, "Runner Markdown source evidence for OpenClerk.") ||
		!strings.Contains(get.Document.Body, "Store public README docs through ingest_source_url.") {
		t.Fatalf("markdown document = %+v", get.Document)
	}
}

func TestDocumentTaskIngestSourceURLInspectMarkdownLinks(t *testing.T) {
	markdownBody := strings.Join([]string{
		"# last30days-skill",
		"",
		"Runner Markdown source evidence for OpenClerk.",
		"",
		"See [Skill](skills/last30days/SKILL.md), [Hermes Setup](HERMES_SETUP.md), [Login](/login), [External](https://example.com/nope.md), and [Badge](badge.svg).",
	}, "\n")
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "web", "README.md")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir markdown fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte(markdownBody), 0o644); err != nil {
		t.Fatalf("write markdown fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)
	sourceURL := "http://openclerk-eval.local/web/README.md"

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:        sourceURL,
			Mode:       "inspect",
			SourceType: "web",
			Title:      "last30days-skill",
			Limit:      5,
		},
	})
	if err != nil {
		t.Fatalf("inspect markdown source URL: %v", err)
	}
	plan := result.SourceIntakePlan
	if result.Rejected ||
		plan == nil ||
		plan.SourceURL != sourceURL ||
		plan.SourceType != "web" ||
		plan.MIMEType != "text/markdown" ||
		plan.FetchStatus != "inspected_public_source" ||
		plan.WriteStatus != "planned_no_write" ||
		plan.PrimaryCandidate == nil ||
		plan.AgentHandoff == nil ||
		!strings.Contains(plan.TextPreview, "Runner Markdown source evidence") {
		t.Fatalf("inspect plan = %+v", plan)
	}
	if plan.PrimaryCandidate.Relation != "primary" ||
		plan.PrimaryCandidate.CandidateStatus != "public_candidate_requires_ingest_source_url_approval" ||
		!containsString(plan.PrimaryCandidate.CandidateSourcePaths, "sources/web/last30days-skill.md") ||
		!strings.Contains(plan.PrimaryCandidate.NextIngestSourceRequest, `"action":"ingest_source_url"`) ||
		!strings.Contains(plan.PrimaryCandidate.NextIngestSourceRequest, `"path_hint":"sources/web/last30days-skill.md"`) {
		t.Fatalf("primary inspect candidate = %+v", plan.PrimaryCandidate)
	}
	var sawSkill, sawHermes bool
	for _, candidate := range plan.RelatedCandidates {
		if candidate.URL == "http://openclerk-eval.local/web/skills/last30days/SKILL.md" &&
			candidate.LinkText == "Skill" &&
			containsString(candidate.CandidateSourcePaths, "sources/web/skill.md") &&
			strings.Contains(candidate.NextIngestSourceRequest, `"url":"http://openclerk-eval.local/web/skills/last30days/SKILL.md"`) {
			sawSkill = true
		}
		if candidate.URL == "http://openclerk-eval.local/web/HERMES_SETUP.md" &&
			candidate.LinkText == "Hermes Setup" &&
			containsString(candidate.CandidateSourcePaths, "sources/web/hermes-setup.md") {
			sawHermes = true
		}
		if strings.Contains(candidate.URL, "example.com") || strings.HasSuffix(candidate.URL, ".svg") || strings.HasSuffix(candidate.URL, "/login") {
			t.Fatalf("inspect related candidates included unsupported link: %+v", candidate)
		}
	}
	if !sawSkill || !sawHermes {
		t.Fatalf("related inspect candidates = %+v", plan.RelatedCandidates)
	}
	if !strings.Contains(plan.ApprovalBoundary, "not durable-write approval") ||
		!strings.Contains(plan.ValidationBoundaries, "no recursive crawl") ||
		!strings.Contains(plan.AuthorityLimits, "candidate evidence only") {
		t.Fatalf("inspect boundaries = approval:%q validation:%q authority:%q", plan.ApprovalBoundary, plan.ValidationBoundaries, plan.AuthorityLimits)
	}
	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("list after inspect: %v", err)
	}
	if len(list.Documents) != 0 {
		t.Fatalf("inspect wrote source documents: %+v", list.Documents)
	}
}

func TestDocumentTaskIngestVideoURLSuppliedTranscript(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	config := runclient.Config{DatabasePath: dbPath}
	videoURL := "https://www.youtube.com/watch?v=openclerk-demo"
	transcript := "OpenClerk video canonical transcript unique evidence."
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestVideoURL,
		Video: runner.VideoURLInput{
			URL:           videoURL,
			PathHint:      "sources/video-youtube/runner-demo.md",
			AssetPathHint: "assets/video-youtube/runner-demo.json",
			Title:         "Runner Video Demo",
			Transcript: runner.VideoTranscriptInput{
				Text:       transcript,
				Policy:     "supplied",
				Origin:     "user_fixture",
				Language:   "en",
				CapturedAt: "2026-04-27T10:00:00Z",
				Tool:       "manual",
				Model:      "none",
			},
		},
	})
	if err != nil {
		t.Fatalf("ingest video URL: %v", err)
	}
	if result.Rejected || result.VideoIngestion == nil {
		t.Fatalf("ingest result = %+v", result)
	}
	ingestion := result.VideoIngestion
	if ingestion.DocID == "" ||
		ingestion.SourcePath != "sources/video-youtube/runner-demo.md" ||
		ingestion.SourceURL != videoURL ||
		ingestion.AssetPath != "assets/video-youtube/runner-demo.json" ||
		ingestion.TranscriptPolicy != "supplied" ||
		ingestion.TranscriptOrigin != "user_fixture" ||
		ingestion.Language != "en" ||
		ingestion.Tool != "manual" ||
		ingestion.Model != "none" ||
		len(ingestion.TranscriptSHA256) != 64 ||
		len(ingestion.Citations) == 0 {
		t.Fatalf("ingestion = %+v", ingestion)
	}
	assetPath := filepath.Join(filepath.Dir(dbPath), "vault", "assets", "video-youtube", "runner-demo.json")
	assetBytes, err := os.ReadFile(assetPath)
	if err != nil {
		t.Fatalf("asset read: %v", err)
	}
	if strings.Contains(string(assetBytes), transcript) || !strings.Contains(string(assetBytes), ingestion.TranscriptSHA256) {
		t.Fatalf("metadata asset = %s", string(assetBytes))
	}

	get, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  ingestion.DocID,
	})
	if err != nil {
		t.Fatalf("get ingested video document: %v", err)
	}
	if get.Document == nil ||
		get.Document.Metadata["source_type"] != "video_transcript" ||
		get.Document.Metadata["source_url"] != videoURL ||
		get.Document.Metadata["transcript_sha256"] != ingestion.TranscriptSHA256 ||
		get.Document.Metadata["asset_path"] != "assets/video-youtube/runner-demo.json" ||
		!strings.Contains(get.Document.Body, "Runner Video Demo") ||
		!strings.Contains(get.Document.Body, transcript) {
		t.Fatalf("ingested document = %+v", get.Document)
	}

	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       "canonical transcript unique evidence",
			PathPrefix: "sources/",
			Limit:      10,
		},
	})
	if err != nil {
		t.Fatalf("search ingested transcript: %v", err)
	}
	if search.Search == nil || len(search.Search.Hits) == 0 || search.Search.Hits[0].Citations[0].Path != "sources/video-youtube/runner-demo.md" {
		t.Fatalf("search = %+v", search.Search)
	}

	provenance, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "source",
			RefID:   ingestion.DocID,
			Limit:   10,
		},
	})
	if err != nil {
		t.Fatalf("source provenance: %v", err)
	}
	if provenance.Provenance == nil || len(provenance.Provenance.Events) == 0 {
		t.Fatalf("provenance = %+v", provenance.Provenance)
	}
	details := provenance.Provenance.Events[0].Details
	if details["source_url"] != videoURL || details["transcript_sha256"] != ingestion.TranscriptSHA256 {
		t.Fatalf("source provenance details = %+v", details)
	}
}

func TestDocumentTaskRejectsNonPDFSourceURL(t *testing.T) {
	fixtureRoot := t.TempDir()
	fixturePath := filepath.Join(fixtureRoot, "pdf", "not-pdf.txt")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
		t.Fatalf("mkdir non-PDF fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, []byte("not a PDF"), 0o644); err != nil {
		t.Fatalf("write non-PDF fixture: %v", err)
	}
	t.Setenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES", "1")
	t.Setenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT", fixtureRoot)

	_, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           "http://openclerk-eval.local/pdf/not-pdf.txt",
			PathHint:      "sources/not-pdf.md",
			AssetPathHint: "assets/sources/not-pdf.pdf",
			SourceType:    "pdf",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "PDF") {
		t.Fatalf("err = %v, want non-PDF rejection", err)
	}
}

func TestValidationRejectionDoesNotCreateRuntimeFiles(t *testing.T) {
	t.Parallel()

	dataDir := filepath.Join(t.TempDir(), "data")
	result, err := runner.RunDocumentTask(context.Background(), runclient.Config{DatabasePath: filepath.Join(dataDir, "openclerk.sqlite")}, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Title: "Missing path",
			Body:  "# Missing path\n",
		},
	})
	if err != nil {
		t.Fatalf("document task: %v", err)
	}
	if !result.Rejected || result.RejectionReason == "" {
		t.Fatalf("result = %+v, want rejected", result)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Fatalf("data dir exists after validation rejection: %v", err)
	}
}

func TestDocumentTaskListTagFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := runclient.Config{DatabasePath: filepath.Join(t.TempDir(), "data", "openclerk.sqlite")}
	createDocument(t, ctx, config, "notes/tagging/support-handoff.md", "Support Handoff", strings.TrimSpace(`---
tag: support-handoff
---
# Support Handoff

## Summary
Support handoff tag list evidence belongs under active notes.
`)+"\n")
	createDocument(t, ctx, config, "archive/tagging/support-handoff.md", "Archived Support Handoff", strings.TrimSpace(`---
tag: support-handoff
---
# Archived Support Handoff

## Summary
Archived support handoff tag list evidence must be excluded by path prefix.
`)+"\n")
	createDocument(t, ctx, config, "notes/tagging/support-handoffs.md", "Support Handoffs", strings.TrimSpace(`---
tag: support-handoffs
---
# Support Handoffs

## Summary
Plural support handoffs tag list evidence must not match singular support-handoff.
`)+"\n")

	list, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			PathPrefix: "notes/tagging/",
			Tag:        " support-handoff ",
			Limit:      20,
		},
	})
	if err != nil {
		t.Fatalf("list tag: %v", err)
	}
	if len(list.Documents) != 1 || list.Documents[0].Path != "notes/tagging/support-handoff.md" {
		t.Fatalf("list tag result = %+v", list.Documents)
	}

	backCompat, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			MetadataKey:   "tag",
			MetadataValue: "support-handoff",
			Limit:         20,
		},
	})
	if err != nil {
		t.Fatalf("list metadata tag: %v", err)
	}
	if len(backCompat.Documents) != 2 {
		t.Fatalf("backward-compatible metadata tag result = %+v", backCompat.Documents)
	}

	mixed, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			Tag:           "support-handoff",
			MetadataKey:   "tag",
			MetadataValue: "support-handoff",
		},
	})
	if err != nil {
		t.Fatalf("mixed tag list validation: %v", err)
	}
	if !mixed.Rejected || mixed.RejectionReason != "list.tag cannot be combined with metadata_key or metadata_value" {
		t.Fatalf("mixed list result = %+v", mixed)
	}

	empty, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List: runner.DocumentListOptions{
			Tag: " ",
		},
	})
	if err != nil {
		t.Fatalf("empty tag list validation: %v", err)
	}
	if !empty.Rejected || empty.RejectionReason != "list.tag must be non-empty" {
		t.Fatalf("empty list result = %+v", empty)
	}
}

func initGitLifecycleTestRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for git lifecycle tests")
	}
	vaultRoot := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vaultRoot, 0o755); err != nil {
		t.Fatalf("create vault root: %v", err)
	}
	runGitLifecycleTestCommand(t, vaultRoot, "init", "-b", "main")
	runGitLifecycleTestCommand(t, vaultRoot, "config", "user.email", "openclerk@example.test")
	runGitLifecycleTestCommand(t, vaultRoot, "config", "user.name", "OpenClerk Test")
	if err := os.WriteFile(filepath.Join(vaultRoot, ".gitkeep"), []byte("openclerk test repo\n"), 0o644); err != nil {
		t.Fatalf("write gitkeep: %v", err)
	}
	runGitLifecycleTestCommand(t, vaultRoot, "add", ".gitkeep")
	runGitLifecycleTestCommand(t, vaultRoot, "commit", "-m", "initial")
	return vaultRoot
}

func requireOCRFixtureTools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"tesseract", "ocrmypdf", "img2pdf", "python3"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not available", tool)
		}
	}
	cmd := exec.Command("python3", "-c", "import PIL")
	if err := cmd.Run(); err != nil {
		t.Skip("python3 PIL package not available")
	}
}

func writeOCRFixtureArtifacts(t *testing.T, dir string) (string, string) {
	t.Helper()
	script := `import sys
from PIL import Image, ImageDraw, ImageFont
root=sys.argv[1]
text="OPENCLERK OCR FIXTURE\nReceipt ID OC-7781\nTotal paid 42 USD\nScanned image evidence"
img=Image.new("RGB",(1200,700),"white")
d=ImageDraw.Draw(img)
try:
    font=ImageFont.truetype("Arial.ttf",54)
except Exception:
    font=ImageFont.load_default(size=54)
d.multiline_text((80,90),text,fill="black",font=font,spacing=24)
img.save(root+"/receipt.png")
`
	if output, err := exec.Command("python3", "-c", script, dir).CombinedOutput(); err != nil {
		t.Fatalf("write OCR image fixture: %v\n%s", err, string(output))
	}
	imagePath := filepath.Join(dir, "receipt.png")
	pdfPath := filepath.Join(dir, "scanned.pdf")
	if output, err := exec.Command("img2pdf", imagePath, "-o", pdfPath).CombinedOutput(); err != nil {
		t.Fatalf("write OCR PDF fixture: %v\n%s", err, string(output))
	}
	return imagePath, pdfPath
}

func writeRunnerOCRModuleManifest(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "module.json")
	manifest := map[string]any{
		"schema_version": "openclerk-module.v1",
		"module": map[string]any{
			"name":    "tesseract-ocr",
			"version": "0.1.0",
			"kind":    "ocr_provider",
		},
		"provides": []map[string]any{{
			"type": "command",
			"name": "tesseract ocr",
		}},
		"authority": map[string]any{
			"default":        "read_only",
			"durable_writes": "forbidden",
			"forbidden":      []string{"write_documents", "hidden_cloud_egress"},
		},
		"release": map[string]any{
			"status": "supported_optional_module",
		},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal OCR manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write OCR manifest: %v", err)
	}
	return path
}

func assertOCRPlan(t *testing.T, result runner.DocumentTaskResult, wantText string) {
	t.Helper()
	plan := result.ArtifactPlan
	if result.Rejected ||
		plan == nil ||
		plan.OCRExtraction == nil ||
		plan.OCRExtraction.Provider != "tesseract" ||
		plan.OCRExtraction.PrivacyPosture != "local_process_no_network" ||
		plan.LocalArtifact == nil ||
		plan.LocalArtifact.TextStatus != "ocr_review_extracted" ||
		!strings.Contains(plan.BodyPreview, wantText) ||
		plan.WriteStatus != "planned_no_write" ||
		!strings.Contains(plan.NextCreateRequest, "create_document") {
		t.Fatalf("OCR artifact plan = %+v", result)
	}
}

func runGitLifecycleTestCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(output))
	}
}

func gitLifecycleDirtyPath(statuses []runner.GitLifecyclePathStatus, path string) bool {
	for _, status := range statuses {
		if status.Path == path {
			return true
		}
	}
	return false
}

func runnerMoveLinkUpdateIncludes(updates []runner.DocumentLinkUpdate, path string, newTarget string) bool {
	for _, update := range updates {
		if update.Path == path && update.NewTarget == newTarget && update.Occurrences > 0 {
			return true
		}
	}
	return false
}
