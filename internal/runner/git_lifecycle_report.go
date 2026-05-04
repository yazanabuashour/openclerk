package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	gitLifecycleModeStatus     = "status"
	gitLifecycleModeHistory    = "history"
	gitLifecycleModeCheckpoint = "checkpoint"

	gitLifecycleValidationBoundaries = "local git metadata/checkpoint report only; no remote push, branch switch, checkout, reset, restore, raw diff output, direct SQLite, source-built runner, HTTP/MCP bypass, unsupported transport, or broad private content inspection"
	gitLifecycleAuthorityLimits      = "Git is storage-level history only; canonical markdown, source refs, citations, provenance events, projection freshness, and OpenClerk write results remain the product evidence"
)

func runGitLifecycleReport(ctx context.Context, vaultRoot string, options GitLifecycleOptions, config runclient.Config) (GitLifecycleReport, error) {
	mode := options.Mode
	if mode == "" {
		mode = gitLifecycleModeStatus
	}
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	report := GitLifecycleReport{
		Mode:                 mode,
		Paths:                options.Paths,
		WriteStatus:          "no_write",
		ApprovalBoundary:     gitLifecycleApprovalBoundary(mode),
		ValidationBoundaries: gitLifecycleValidationBoundaries,
		AuthorityLimits:      gitLifecycleAuthorityLimits,
	}

	repo, err := inspectGitLifecycleRepo(ctx, vaultRoot)
	if err != nil {
		if errors.Is(err, errGitLifecycleUnavailable) {
			report.GitStatus = "unavailable"
			report.CheckpointStatus = checkpointStatusForUnavailable(mode)
			report.AgentHandoff = gitLifecycleHandoff(report)
			return report, nil
		}
		return GitLifecycleReport{}, err
	}
	report.GitStatus = "available"
	report.Branch = repo.branch
	report.Head = repo.head

	statuses, err := gitLifecycleStatus(ctx, vaultRoot, options.Paths)
	if err != nil {
		return GitLifecycleReport{}, err
	}
	report.DirtyPaths = statuses
	if len(statuses) > 0 {
		report.GitStatus = "dirty"
	}

	switch mode {
	case gitLifecycleModeStatus:
	case gitLifecycleModeHistory:
		history, err := gitLifecycleHistory(ctx, vaultRoot, options.Paths, limit)
		if err != nil {
			return GitLifecycleReport{}, err
		}
		report.History = history
	case gitLifecycleModeCheckpoint:
		updated, err := gitLifecycleCheckpoint(ctx, vaultRoot, options, config)
		if err != nil {
			return GitLifecycleReport{}, err
		}
		report.CheckpointStatus = updated.CheckpointStatus
		report.CommitID = updated.CommitID
		report.WriteStatus = updated.WriteStatus
		report.Head = firstNonEmpty(updated.CommitID, report.Head)
		afterStatus, err := gitLifecycleStatus(ctx, vaultRoot, options.Paths)
		if err != nil {
			return GitLifecycleReport{}, err
		}
		report.DirtyPaths = afterStatus
		report.GitStatus = "available"
		if len(afterStatus) > 0 {
			report.GitStatus = "dirty"
		}
	}

	report.AgentHandoff = gitLifecycleHandoff(report)
	return report, nil
}

type gitLifecycleRepo struct {
	branch string
	head   string
}

var errGitLifecycleUnavailable = errors.New("git lifecycle unavailable")

func inspectGitLifecycleRepo(ctx context.Context, vaultRoot string) (gitLifecycleRepo, error) {
	inside, err := runGitLifecycleCommand(ctx, vaultRoot, "rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(inside) != "true" {
		return gitLifecycleRepo{}, errGitLifecycleUnavailable
	}
	branch, err := runGitLifecycleCommand(ctx, vaultRoot, "branch", "--show-current")
	if err != nil {
		return gitLifecycleRepo{}, err
	}
	head, err := runGitLifecycleCommand(ctx, vaultRoot, "rev-parse", "--short", "HEAD")
	if err != nil {
		head = "none"
	}
	branch = strings.TrimSpace(branch)
	if branch == "" {
		branch = "detached"
	}
	return gitLifecycleRepo{branch: branch, head: strings.TrimSpace(head)}, nil
}

func gitLifecycleStatus(ctx context.Context, vaultRoot string, paths []string) ([]GitLifecyclePathStatus, error) {
	args := append([]string{"status", "--porcelain=v1", "--"}, paths...)
	output, err := runGitLifecycleCommand(ctx, vaultRoot, args...)
	if err != nil {
		return nil, err
	}
	output = strings.TrimRight(output, "\r\n")
	if output == "" {
		return nil, nil
	}
	lines := strings.Split(output, "\n")
	statuses := make([]GitLifecyclePathStatus, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		statuses = append(statuses, parseGitLifecycleStatusLine(line))
	}
	return statuses, nil
}

func parseGitLifecycleStatusLine(line string) GitLifecyclePathStatus {
	statusCode := strings.TrimSpace(line)
	pathText := ""
	if len(line) >= 3 {
		statusCode = strings.TrimSpace(line[:2])
		pathText = strings.TrimSpace(line[3:])
	}
	if renamed := strings.LastIndex(pathText, " -> "); renamed >= 0 {
		pathText = strings.TrimSpace(pathText[renamed+4:])
	}
	return GitLifecyclePathStatus{
		Path:   strings.Trim(pathText, `"`),
		Status: gitLifecycleStatusLabel(statusCode),
	}
}

func gitLifecycleStatusLabel(code string) string {
	switch {
	case code == "??":
		return "untracked"
	case strings.Contains(code, "A"):
		return "added"
	case strings.Contains(code, "D"):
		return "deleted"
	case strings.Contains(code, "R"):
		return "renamed"
	case strings.Contains(code, "M"):
		return "modified"
	default:
		return "changed"
	}
}

func isUnsafeGitLifecyclePath(path string) bool {
	return strings.HasPrefix(path, ":") ||
		strings.ContainsAny(path, "*?[") ||
		strings.Contains(path, `\`)
}

func gitLifecycleHistory(ctx context.Context, vaultRoot string, paths []string, limit int) ([]GitLifecycleCommit, error) {
	format := "%H%x1f%ct%x1f%s"
	args := append([]string{"log", "--max-count=" + strconv.Itoa(limit), "--format=" + format, "--"}, paths...)
	output, err := runGitLifecycleCommand(ctx, vaultRoot, args...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	commits := make([]GitLifecycleCommit, 0, len(lines))
	pathScope := "all"
	if len(paths) > 0 {
		pathScope = strings.Join(paths, ",")
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\x1f", 3)
		if len(parts) != 3 {
			continue
		}
		authored := ""
		if unixSeconds, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			authored = time.Unix(unixSeconds, 0).UTC().Format(time.RFC3339)
		}
		shortID := parts[0]
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		commits = append(commits, GitLifecycleCommit{
			CommitID:  parts[0],
			ShortID:   shortID,
			Authored:  authored,
			Summary:   sanitizeGitLifecycleMessage(parts[2]),
			PathScope: pathScope,
		})
	}
	return commits, nil
}

func gitLifecycleCheckpoint(ctx context.Context, vaultRoot string, options GitLifecycleOptions, config runclient.Config) (GitLifecycleReport, error) {
	if !gitLifecycleCheckpointsEnabled(config) {
		return GitLifecycleReport{
			CheckpointStatus: "disabled",
			WriteStatus:      "rejected",
		}, nil
	}
	if err := runGitLifecycleCommandNoOutput(ctx, vaultRoot, append([]string{"add", "--"}, options.Paths...)...); err != nil {
		return GitLifecycleReport{}, err
	}
	hasChanges, err := gitLifecycleStagedChanges(ctx, vaultRoot, options.Paths)
	if err != nil {
		return GitLifecycleReport{}, err
	}
	if !hasChanges {
		return GitLifecycleReport{
			CheckpointStatus: "unchanged",
			WriteStatus:      "unchanged",
		}, nil
	}
	if err := runGitLifecycleCommandNoOutput(ctx, vaultRoot, append([]string{"-c", "commit.gpgsign=false", "commit", "-m", options.Message, "--"}, options.Paths...)...); err != nil {
		return GitLifecycleReport{}, err
	}
	head, err := runGitLifecycleCommand(ctx, vaultRoot, "rev-parse", "--short", "HEAD")
	if err != nil {
		return GitLifecycleReport{}, err
	}
	return GitLifecycleReport{
		CheckpointStatus: "created",
		CommitID:         strings.TrimSpace(head),
		WriteStatus:      "checkpoint_created",
	}, nil
}

func gitLifecycleStagedChanges(ctx context.Context, vaultRoot string, paths []string) (bool, error) {
	args := append([]string{"diff", "--cached", "--quiet", "--"}, paths...)
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", vaultRoot}, args...)...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return false, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("run local git lifecycle diff check: %w: %s", err, strings.TrimSpace(string(output)))
}

func runGitLifecycleCommand(ctx context.Context, vaultRoot string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", vaultRoot}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("run local git lifecycle command: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func runGitLifecycleCommandNoOutput(ctx context.Context, vaultRoot string, args ...string) error {
	_, err := runGitLifecycleCommand(ctx, vaultRoot, args...)
	return err
}

func gitLifecycleCheckpointsEnabled(config runclient.Config) bool {
	if config.GitCheckpoints {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("OPENCLERK_GIT_CHECKPOINTS"))) {
	case "1", "true", "enabled", "on", "yes":
		return true
	default:
		return false
	}
}

func sanitizeGitLifecycleMessage(message string) string {
	return strings.Join(strings.Fields(message), " ")
}

func checkpointStatusForUnavailable(mode string) string {
	if mode == gitLifecycleModeCheckpoint {
		return "unavailable"
	}
	return ""
}

func gitLifecycleApprovalBoundary(mode string) string {
	if mode == gitLifecycleModeCheckpoint {
		return "status and history are read-only; checkpoint writes require explicit git_lifecycle mode checkpoint plus --git-checkpoints or OPENCLERK_GIT_CHECKPOINTS=1 and only commit caller-specified vault-relative paths"
	}
	return "status and history are read-only; checkpoint writes require explicit mode checkpoint plus runner config"
}

func gitLifecycleSummary(report GitLifecycleReport) string {
	switch report.Mode {
	case gitLifecycleModeHistory:
		return fmt.Sprintf("returned %d git lifecycle commits", len(report.History))
	case gitLifecycleModeCheckpoint:
		if report.CheckpointStatus == "created" {
			return fmt.Sprintf("created local git checkpoint %s", report.CommitID)
		}
		return "returned git lifecycle checkpoint report"
	default:
		return fmt.Sprintf("returned git lifecycle status with %d changed paths", len(report.DirtyPaths))
	}
}

func gitLifecycleHandoff(report GitLifecycleReport) *AgentHandoff {
	evidence := []string{
		"mode:" + report.Mode,
		"git_status:" + report.GitStatus,
		"write_status:" + report.WriteStatus,
	}
	if report.Branch != "" {
		evidence = append(evidence, "branch:"+report.Branch)
	}
	if report.Head != "" {
		evidence = append(evidence, "head:"+report.Head)
	}
	if report.CommitID != "" {
		evidence = append(evidence, "checkpoint:"+report.CommitID)
	}
	for _, path := range report.Paths {
		evidence = append(evidence, "path:"+filepath.ToSlash(path))
	}
	return &AgentHandoff{
		AnswerSummary:               gitLifecycleAnswerSummary(report),
		Evidence:                    evidence,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "Use provenance_events and projection_states for OpenClerk semantic evidence; use git_lifecycle_report only for local storage checkpoint/status/history metadata.",
	}
}

func gitLifecycleAnswerSummary(report GitLifecycleReport) string {
	switch report.Mode {
	case gitLifecycleModeHistory:
		return fmt.Sprintf("git_lifecycle_report returned %d local storage-history entries without raw diffs.", len(report.History))
	case gitLifecycleModeCheckpoint:
		if report.CheckpointStatus == "created" {
			return fmt.Sprintf("git_lifecycle_report created local checkpoint %s for caller-specified vault-relative paths.", report.CommitID)
		}
		return "git_lifecycle_report did not create a checkpoint; inspect checkpoint_status and write_status for the reason."
	default:
		return fmt.Sprintf("git_lifecycle_report found %d changed local storage paths without raw diffs.", len(report.DirtyPaths))
	}
}
