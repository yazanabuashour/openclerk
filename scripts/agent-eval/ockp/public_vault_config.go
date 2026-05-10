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
	publicVaultModeGoDocs         = "go-docs"
	publicVaultModeMobyDick       = "moby-dick"
)

type publicVaultProfile struct {
	Mode             string
	Lane             string
	DisplayName      string
	ReportName       string
	RepoURL          string
	RepoRef          string
	Subdir           string
	SourcePrefix     string
	TaskManifestPath string
	FileExtensions   []string
	SynthesisSlug    string
	Promotion        string
}

var publicVaultProfiles = map[string]publicVaultProfile{
	publicVaultModeKubernetesDocs: {
		Mode:             publicVaultModeKubernetesDocs,
		Lane:             "public-vault-kubernetes-docs",
		DisplayName:      "Kubernetes Docs",
		ReportName:       "ockp-public-vault-kubernetes-docs",
		RepoURL:          "https://github.com/kubernetes/website.git",
		RepoRef:          "7e7144c3969feb5d57a3c757ac462bd271f4a691",
		Subdir:           "content/en/docs",
		SourcePrefix:     "sources/kubernetes/website/content/en/docs",
		TaskManifestPath: "docs/evals/public-vault-kubernetes-docs-tasks.json",
		FileExtensions:   []string{".md"},
		SynthesisSlug:    "kubernetes-docs",
		Promotion:        "public-vault Kubernetes docs lane is promoted for recurring large public-vault UX validation; this promotes the eval lane only and does not add a new runner API",
	},
	publicVaultModeGoDocs: {
		Mode:             publicVaultModeGoDocs,
		Lane:             "public-vault-go-docs",
		DisplayName:      "Go Docs",
		ReportName:       "ockp-public-vault-go-docs",
		RepoURL:          "https://github.com/golang/website.git",
		RepoRef:          "31fb202f84245709e774bf7c85d13430925d45e5",
		Subdir:           "_content",
		SourcePrefix:     "sources/golang/website/_content",
		TaskManifestPath: "docs/evals/public-vault-go-docs-tasks.json",
		FileExtensions:   []string{".md"},
		SynthesisSlug:    "go-docs",
		Promotion:        "public-vault Go docs lane is promoted for a second large technical corpus autonomy validation; this promotes the eval lane only and does not add a new runner API",
	},
	publicVaultModeMobyDick: {
		Mode:             publicVaultModeMobyDick,
		Lane:             "public-vault-moby-dick",
		DisplayName:      "Moby-Dick",
		ReportName:       "ockp-public-vault-moby-dick",
		RepoURL:          "https://github.com/GITenberg/Moby-Dick--Or-The-Whale_2701.git",
		RepoRef:          "bdf1948e6cd00963730971e5624e764a35f238c3",
		Subdir:           ".",
		SourcePrefix:     "sources/gitenberg/moby-dick",
		TaskManifestPath: "docs/evals/public-vault-moby-dick-tasks.json",
		FileExtensions:   []string{".md", ".txt", ".asciidoc", ".rst"},
		SynthesisSlug:    "moby-dick",
		Promotion:        "public-vault Moby-Dick lane is promoted for non-technical public-corpus autonomy validation; this promotes the eval lane only and does not add a new runner API",
	},
}

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
	SourcePrefix     string
	FileExtensions   []string
	SynthesisSlug    string
	DisplayName      string
	Lane             string
	Promotion        string
	Autonomy         runnerAutonomyModes
}

type runnerAutonomyModes struct {
	ApprovalMode    string `json:"approval_mode"`
	DraftingMode    string `json:"drafting_mode"`
	WriteTargetMode string `json:"write_target_mode"`
	CitationMode    string `json:"citation_mode"`
	PrivacyMode     string `json:"privacy_mode"`
	AudienceMode    string `json:"audience_mode"`
}

func normalizeRunnerAutonomyModes(input runnerAutonomyModes) (runnerAutonomyModes, string) {
	normalized := runnerAutonomyModes{
		ApprovalMode:    strings.TrimSpace(input.ApprovalMode),
		DraftingMode:    strings.TrimSpace(input.DraftingMode),
		WriteTargetMode: strings.TrimSpace(input.WriteTargetMode),
		CitationMode:    strings.TrimSpace(input.CitationMode),
		PrivacyMode:     strings.TrimSpace(input.PrivacyMode),
		AudienceMode:    strings.TrimSpace(input.AudienceMode),
	}
	checks := []struct {
		name    string
		value   string
		allowed []string
	}{
		{"approval_mode", normalized.ApprovalMode, []string{"propose_only", "approve_write", "autonomous_disposable", "autonomous_trusted"}},
		{"drafting_mode", normalized.DraftingMode, []string{"require_explicit_fields", "suggest_fields", "autonomous_fields"}},
		{"write_target_mode", normalized.WriteTargetMode, []string{"existing_only", "create_or_update", "create_allowed"}},
		{"citation_mode", normalized.CitationMode, []string{"strict", "balanced", "lightweight"}},
		{"privacy_mode", normalized.PrivacyMode, []string{"private_summary_only", "allow_paths", "allow_titles", "allow_snippets"}},
		{"audience_mode", normalized.AudienceMode, []string{"technical", "plain_language", "executive_summary"}},
	}
	for _, check := range checks {
		if !stringIn(check.value, check.allowed...) {
			return runnerAutonomyModes{}, fmt.Sprintf("autonomy.%s must be one of %s", check.name, strings.Join(check.allowed, ", "))
		}
	}
	return normalized, ""
}

func stringIn(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func parsePublicVaultConfig(args []string, stderr io.Writer) (publicVaultConfig, error) {
	if len(args) == 0 {
		return publicVaultConfig{}, errors.New("usage: ockp public-vault kubernetes-docs|go-docs|moby-dick [flags]")
	}
	mode := strings.TrimSpace(args[0])
	profile, ok := publicVaultProfiles[mode]
	if !ok {
		return publicVaultConfig{}, fmt.Errorf("unsupported public-vault mode %q", mode)
	}

	fs := flag.NewFlagSet("ockp public-vault "+mode, flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := publicVaultConfig{
		Mode:             mode,
		Parallel:         defaultParallel,
		ReportDir:        filepath.Join("docs", "evals", "results"),
		ReportName:       profile.ReportName,
		CodexBin:         "codex",
		RepoRoot:         ".",
		CacheMode:        cacheModeShared,
		PublicRepoURL:    profile.RepoURL,
		PublicRepoRef:    profile.RepoRef,
		PublicSubdir:     profile.Subdir,
		TaskManifestPath: profile.TaskManifestPath,
		SourcePrefix:     profile.SourcePrefix,
		FileExtensions:   append([]string{}, profile.FileExtensions...),
		SynthesisSlug:    profile.SynthesisSlug,
		DisplayName:      profile.DisplayName,
		Lane:             profile.Lane,
		Promotion:        profile.Promotion,
		Autonomy: runnerAutonomyModes{
			ApprovalMode:    "autonomous_disposable",
			DraftingMode:    "autonomous_fields",
			WriteTargetMode: "create_or_update",
			CitationMode:    "balanced",
			PrivacyMode:     "allow_paths",
			AudienceMode:    "plain_language",
		},
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
	fs.StringVar(&config.Autonomy.ApprovalMode, "approval-mode", config.Autonomy.ApprovalMode, "approval mode: propose_only, approve_write, autonomous_disposable, or autonomous_trusted")
	fs.StringVar(&config.Autonomy.DraftingMode, "drafting-mode", config.Autonomy.DraftingMode, "drafting mode: require_explicit_fields, suggest_fields, or autonomous_fields")
	fs.StringVar(&config.Autonomy.WriteTargetMode, "write-target-mode", config.Autonomy.WriteTargetMode, "write target mode: existing_only, create_or_update, or create_allowed")
	fs.StringVar(&config.Autonomy.CitationMode, "citation-mode", config.Autonomy.CitationMode, "citation mode: strict, balanced, or lightweight")
	fs.StringVar(&config.Autonomy.PrivacyMode, "privacy-mode", config.Autonomy.PrivacyMode, "privacy mode: private_summary_only, allow_paths, allow_titles, or allow_snippets")
	fs.StringVar(&config.Autonomy.AudienceMode, "audience-mode", config.Autonomy.AudienceMode, "audience mode: technical, plain_language, or executive_summary")
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
	config.SourcePrefix = publicVaultProfileSourcePrefix(profile, config.PublicSubdir)
	if strings.TrimSpace(config.TaskManifestPath) == "" {
		return publicVaultConfig{}, errors.New("--task-manifest must not be empty")
	}
	if strings.TrimSpace(config.ReportName) == "" {
		return publicVaultConfig{}, errors.New("--report-name must not be empty")
	}
	autonomy, rejection := normalizeRunnerAutonomyModes(config.Autonomy)
	if rejection != "" {
		return publicVaultConfig{}, errors.New(rejection)
	}
	config.Autonomy = autonomy
	if config.RunRoot == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-public-vault-%d", time.Now().UnixNano()))
	}
	return config, nil
}

func publicVaultProfileSourcePrefix(profile publicVaultProfile, subdir string) string {
	cleanSubdir := cleanPublicVaultSubdir(subdir)
	if cleanSubdir == "" {
		return profile.SourcePrefix
	}
	if cleanSubdir == profile.Subdir {
		return profile.SourcePrefix
	}
	if profile.Subdir == "." {
		if cleanSubdir == "." {
			return profile.SourcePrefix
		}
		return path.Join(profile.SourcePrefix, cleanSubdir)
	}
	base := strings.TrimSuffix(profile.SourcePrefix, "/"+strings.Trim(profile.Subdir, "/"))
	if cleanSubdir == "." {
		return base
	}
	return path.Join(base, cleanSubdir)
}

func cleanPublicVaultSubdir(value string) string {
	raw := strings.Trim(strings.TrimSpace(strings.ReplaceAll(value, "\\", "/")), "/")
	if raw == "" {
		return "."
	}
	if path.IsAbs(raw) {
		return ""
	}
	for _, part := range strings.Split(raw, "/") {
		if part == ".." {
			return ""
		}
	}
	cleaned := path.Clean(raw)
	if cleaned == "." {
		return "."
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}
