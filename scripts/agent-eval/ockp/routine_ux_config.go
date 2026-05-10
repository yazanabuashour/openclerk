package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const routineUXModeRealVault = "real-vault"

type routineUXConfig struct {
	Mode             string
	Parallel         int
	RunRoot          string
	ReportDir        string
	ReportName       string
	CodexBin         string
	RepoRoot         string
	CacheMode        string
	PrivateVaultRoot string
	TaskManifestPath string
}

func parseRoutineUXConfig(args []string, stderr io.Writer) (routineUXConfig, error) {
	if len(args) == 0 {
		return routineUXConfig{}, errors.New("usage: ockp routine-ux real-vault [flags]")
	}
	mode := strings.TrimSpace(args[0])
	if mode != routineUXModeRealVault {
		return routineUXConfig{}, fmt.Errorf("unsupported routine-ux mode %q", mode)
	}

	fs := flag.NewFlagSet("ockp routine-ux "+mode, flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := routineUXConfig{
		Mode:       mode,
		Parallel:   defaultParallel,
		ReportDir:  filepath.Join("docs", "evals", "results"),
		ReportName: "ockp-real-vault-routine-ux",
		CodexBin:   "codex",
		RepoRoot:   ".",
		CacheMode:  cacheModeShared,
	}
	fs.IntVar(&config.Parallel, "parallel", config.Parallel, "number of private routine UX tasks to run concurrently")
	fs.StringVar(&config.RunRoot, "run-root", "", "directory for local private routine UX artifacts")
	fs.StringVar(&config.ReportDir, "report-dir", config.ReportDir, "directory for sanitized Markdown reports")
	fs.StringVar(&config.ReportName, "report-name", config.ReportName, "base filename for the sanitized Markdown report")
	fs.StringVar(&config.CodexBin, "codex-bin", config.CodexBin, "codex executable")
	fs.StringVar(&config.RepoRoot, "repo-root", config.RepoRoot, "repository root to copy for each job")
	fs.StringVar(&config.CacheMode, "cache-mode", config.CacheMode, "Go cache mode: shared or isolated")
	fs.StringVar(&config.PrivateVaultRoot, "vault-root", "", "private/local vault root for routine UX telemetry")
	fs.StringVar(&config.TaskManifestPath, "task-manifest", "", "local-only private task manifest")
	if err := fs.Parse(args[1:]); err != nil {
		return routineUXConfig{}, err
	}
	if fs.NArg() != 0 {
		return routineUXConfig{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if config.Parallel < 1 {
		return routineUXConfig{}, errors.New("--parallel must be at least 1")
	}
	if config.CacheMode != cacheModeShared && config.CacheMode != cacheModeIsolated {
		return routineUXConfig{}, fmt.Errorf("--cache-mode must be %q or %q", cacheModeShared, cacheModeIsolated)
	}
	if strings.TrimSpace(config.PrivateVaultRoot) == "" {
		return routineUXConfig{}, errors.New("routine-ux real-vault requires --vault-root")
	}
	if strings.TrimSpace(config.TaskManifestPath) == "" {
		return routineUXConfig{}, errors.New("routine-ux real-vault requires --task-manifest")
	}
	if strings.TrimSpace(config.ReportName) == "" {
		return routineUXConfig{}, errors.New("--report-name must not be empty")
	}
	if config.RunRoot == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-routine-ux-%d", time.Now().UnixNano()))
	}
	return config, nil
}
