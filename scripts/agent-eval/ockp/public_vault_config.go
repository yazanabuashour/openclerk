package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	publicVaultModeKubernetesDocs = "kubernetes-docs"
	defaultPublicVaultRepoURL     = "https://github.com/kubernetes/website.git"
	defaultPublicVaultRepoRef     = "7e7144c3969feb5d57a3c757ac462bd271f4a691"
	defaultPublicVaultSubdir      = "content/en/docs"
	defaultPublicVaultManifest    = "docs/evals/public-vault-kubernetes-docs-tasks.json"
)

type publicVaultConfig struct {
	Mode             string
	Parallel         int
	RunRoot          string
	ReportDir        string
	ReportName       string
	CodexBin         string
	RepoRoot         string
	CacheMode        string
	PublicRepoURL    string
	PublicRepoRef    string
	PublicSubdir     string
	TaskManifestPath string
}

func parsePublicVaultConfig(args []string, stderr io.Writer) (publicVaultConfig, error) {
	if len(args) == 0 {
		return publicVaultConfig{}, errors.New("usage: ockp public-vault kubernetes-docs [flags]")
	}
	mode := strings.TrimSpace(args[0])
	if mode != publicVaultModeKubernetesDocs {
		return publicVaultConfig{}, fmt.Errorf("unsupported public-vault mode %q", mode)
	}

	fs := flag.NewFlagSet("ockp public-vault "+mode, flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := publicVaultConfig{
		Mode:             mode,
		Parallel:         defaultParallel,
		ReportDir:        filepath.Join("docs", "evals", "results"),
		ReportName:       "ockp-public-vault-kubernetes-docs",
		CodexBin:         "codex",
		RepoRoot:         ".",
		CacheMode:        cacheModeShared,
		PublicRepoURL:    defaultPublicVaultRepoURL,
		PublicRepoRef:    defaultPublicVaultRepoRef,
		PublicSubdir:     defaultPublicVaultSubdir,
		TaskManifestPath: defaultPublicVaultManifest,
	}
	fs.IntVar(&config.Parallel, "parallel", config.Parallel, "number of public vault tasks to run concurrently")
	fs.StringVar(&config.RunRoot, "run-root", "", "directory for local public vault artifacts")
	fs.StringVar(&config.ReportDir, "report-dir", config.ReportDir, "directory for public vault reports")
	fs.StringVar(&config.ReportName, "report-name", config.ReportName, "base filename for the public vault report")
	fs.StringVar(&config.CodexBin, "codex-bin", config.CodexBin, "codex executable")
	fs.StringVar(&config.RepoRoot, "repo-root", config.RepoRoot, "repository root to copy for each job")
	fs.StringVar(&config.CacheMode, "cache-mode", config.CacheMode, "Go cache mode: shared or isolated")
	fs.StringVar(&config.PublicRepoURL, "repo-url", config.PublicRepoURL, "public source repository URL or local directory")
	fs.StringVar(&config.PublicRepoRef, "repo-ref", config.PublicRepoRef, "public source repository commit, branch, or tag")
	fs.StringVar(&config.PublicSubdir, "subdir", config.PublicSubdir, "public repository subtree to materialize")
	fs.StringVar(&config.TaskManifestPath, "task-manifest", config.TaskManifestPath, "committed public task manifest")
	if err := fs.Parse(args[1:]); err != nil {
		return publicVaultConfig{}, err
	}
	if fs.NArg() != 0 {
		return publicVaultConfig{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if config.Parallel < 1 {
		return publicVaultConfig{}, errors.New("--parallel must be at least 1")
	}
	if config.CacheMode != cacheModeShared && config.CacheMode != cacheModeIsolated {
		return publicVaultConfig{}, fmt.Errorf("--cache-mode must be %q or %q", cacheModeShared, cacheModeIsolated)
	}
	if strings.TrimSpace(config.PublicRepoURL) == "" {
		return publicVaultConfig{}, errors.New("--repo-url must not be empty")
	}
	if strings.TrimSpace(config.PublicRepoRef) == "" {
		return publicVaultConfig{}, errors.New("--repo-ref must not be empty")
	}
	config.PublicSubdir = cleanPublicVaultSubdir(config.PublicSubdir)
	if config.PublicSubdir == "" {
		return publicVaultConfig{}, errors.New("--subdir must be a repository-relative path")
	}
	if strings.TrimSpace(config.TaskManifestPath) == "" {
		return publicVaultConfig{}, errors.New("--task-manifest must not be empty")
	}
	if strings.TrimSpace(config.ReportName) == "" {
		return publicVaultConfig{}, errors.New("--report-name must not be empty")
	}
	if config.RunRoot == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-public-vault-%d", time.Now().UnixNano()))
	}
	return config, nil
}

func cleanPublicVaultSubdir(value string) string {
	raw := strings.Trim(strings.TrimSpace(strings.ReplaceAll(value, "\\", "/")), "/")
	if raw == "" || path.IsAbs(raw) {
		return ""
	}
	cleaned := path.Clean(raw)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	for _, part := range strings.Split(cleaned, "/") {
		if part == ".." {
			return ""
		}
	}
	return cleaned
}
