package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const smokeVersion = "v0.0.0-smoke"

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("validateliveinstallsmoke", flag.ContinueOnError)
	reportDir := fs.String("report-dir", filepath.Join("docs", "evals", "results"), "directory for reduced smoke reports")
	reportName := fs.String("report-name", "ockp-live-install-upgrade-module-smoke", "base report filename without extension")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	report, err := runSmoke(context.Background())
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*reportDir, 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}
	if err := writeJSON(filepath.Join(*reportDir, *reportName+".json"), report); err != nil {
		return err
	}
	if err := writeMarkdown(filepath.Join(*reportDir, *reportName+".md"), report); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "validated live install/upgrade/module smoke; wrote %s and %s\n", filepath.ToSlash(filepath.Join(*reportDir, *reportName+".json")), filepath.ToSlash(filepath.Join(*reportDir, *reportName+".md")))
	return err
}

type smokeReport struct {
	SchemaVersion        string            `json:"schema_version"`
	GeneratedAt          time.Time         `json:"generated_at"`
	Version              string            `json:"version"`
	TempEnvironment      tempEnvironment   `json:"temp_environment"`
	ReleaseFixture       releaseFixture    `json:"release_fixture"`
	Install              installCheck      `json:"install"`
	Upgrade              installCheck      `json:"upgrade"`
	Skill                skillCheck        `json:"skill"`
	Module               moduleSmoke       `json:"module"`
	ValidationBoundaries string            `json:"validation_boundaries"`
	ArtifactPlaceholders map[string]string `json:"artifact_placeholders"`
}

type tempEnvironment struct {
	Home       string `json:"home"`
	CodexHome  string `json:"codex_home"`
	Database   string `json:"database"`
	InstallDir string `json:"install_dir"`
}

type releaseFixture struct {
	Transport       string `json:"transport"`
	Archive         string `json:"archive"`
	Checksum        string `json:"checksum"`
	InstallScript   string `json:"install_script"`
	RealBinary      bool   `json:"real_binary"`
	RealChecksum    bool   `json:"real_checksum"`
	InstallerSource string `json:"installer_source"`
}

type installCheck struct {
	Passed              bool   `json:"passed"`
	InstallerInvocation string `json:"installer_invocation"`
	BinaryPath          string `json:"binary_path"`
	CommandPath         string `json:"command_path"`
	VersionOutput       string `json:"version_output"`
	HelpChecked         bool   `json:"help_checked"`
}

type skillCheck struct {
	Passed    bool   `json:"passed"`
	SkillPath string `json:"skill_path"`
	Source    string `json:"source"`
}

type moduleSmoke struct {
	Passed            bool              `json:"passed"`
	Provider          string            `json:"provider"`
	ManifestPath      string            `json:"manifest_path"`
	SkillPath         string            `json:"skill_path"`
	InstallPassed     bool              `json:"install_passed"`
	ConfigurePassed   bool              `json:"configure_passed"`
	UpgradePassed     bool              `json:"upgrade_passed"`
	UpgradePreserved  bool              `json:"upgrade_preserved_config"`
	ListPassed        bool              `json:"list_passed"`
	RemovePassed      bool              `json:"remove_passed"`
	FinalListEmpty    bool              `json:"final_list_empty"`
	ProviderConfig    map[string]string `json:"provider_config"`
	VerificationState string            `json:"verification_state"`
	RedactionState    string            `json:"redaction_state"`
}

type smokeContext struct {
	workDir    string
	distDir    string
	toolsDir   string
	homeDir    string
	codexHome  string
	installDir string
	dbPath     string
	repoRoot   string
	archive    string
	checksum   string
	env        []string
}

func runSmoke(ctx context.Context) (smokeReport, error) {
	repoRoot, err := os.Getwd()
	if err != nil {
		return smokeReport{}, err
	}
	workDir, err := os.MkdirTemp("", "openclerk-live-install-smoke-*")
	if err != nil {
		return smokeReport{}, err
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	smoke := smokeContext{
		workDir:    workDir,
		distDir:    filepath.Join(workDir, "dist"),
		toolsDir:   filepath.Join(workDir, "tools"),
		homeDir:    filepath.Join(workDir, "codex-home"),
		codexHome:  filepath.Join(workDir, "codex-home", ".codex"),
		installDir: filepath.Join(workDir, "codex-home", ".local", "bin"),
		dbPath:     filepath.Join(workDir, "openclerk.sqlite"),
		repoRoot:   repoRoot,
	}
	if err := prepareReleaseFixture(ctx, &smoke); err != nil {
		return smokeReport{}, err
	}
	if err := prepareSmokeEnvironment(&smoke); err != nil {
		return smokeReport{}, err
	}

	install, err := runInstaller(ctx, smoke, false)
	if err != nil {
		return smokeReport{}, err
	}
	skill, err := installSkill(smoke)
	if err != nil {
		return smokeReport{}, err
	}
	if err := replaceWithOldBinary(smoke.installDir); err != nil {
		return smokeReport{}, err
	}
	upgrade, err := runInstaller(ctx, smoke, true)
	if err != nil {
		return smokeReport{}, err
	}
	module, err := runModuleSmoke(ctx, smoke)
	if err != nil {
		return smokeReport{}, err
	}

	return smokeReport{
		SchemaVersion: "openclerk-live-install-smoke.v1",
		GeneratedAt:   time.Now().UTC(),
		Version:       smokeVersion,
		TempEnvironment: tempEnvironment{
			Home:       "$HOME",
			CodexHome:  "$CODEX_HOME",
			Database:   "<temp-openclerk-db>",
			InstallDir: "$HOME/.local/bin",
		},
		ReleaseFixture: releaseFixture{
			Transport:       "local curl shim for GitHub release URLs",
			Archive:         filepath.ToSlash(smoke.archive),
			Checksum:        filepath.ToSlash(smoke.checksum),
			InstallScript:   "dist/install.sh",
			RealBinary:      true,
			RealChecksum:    true,
			InstallerSource: "scripts/install.sh",
		},
		Install:              install,
		Upgrade:              upgrade,
		Skill:                skill,
		Module:               module,
		ValidationBoundaries: "local temp HOME/CODEX_HOME only; release transport is a deterministic local fixture for the GitHub release URLs; no durable host install, no network release fetch, no external provider call, no direct SQLite edit, and no source-built runner invocation after install",
		ArtifactPlaceholders: map[string]string{
			"<temp-openclerk-db>": "isolated smoke database",
			"$CODEX_HOME":         "isolated temp Codex home",
			"$HOME":               "isolated temp home",
		},
	}, nil
}

func prepareReleaseFixture(ctx context.Context, smoke *smokeContext) error {
	for _, dir := range []string{smoke.distDir, smoke.toolsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	osName, arch, err := installerPlatform()
	if err != nil {
		return err
	}
	assetVersion := strings.TrimPrefix(smokeVersion, "v")
	name := fmt.Sprintf("openclerk_%s_%s_%s", assetVersion, osName, arch)
	binDir := filepath.Join(smoke.distDir, name)
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "go", "build", "-trimpath", "-ldflags=-s -w -X main.version="+smokeVersion, "-o", filepath.Join(binDir, "openclerk"), "./cmd/openclerk")
	cmd.Dir = smoke.repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build smoke openclerk binary: %w: %s", err, strings.TrimSpace(string(output)))
	}
	archive := name + ".tar.gz"
	tarCmd := exec.CommandContext(ctx, "tar", "-C", smoke.distDir, "-czf", filepath.Join(smoke.distDir, archive), name)
	output, err = tarCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("archive smoke openclerk binary: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if err := os.RemoveAll(binDir); err != nil {
		return err
	}
	checksum := fmt.Sprintf("openclerk_%s_checksums.txt", assetVersion)
	digest, err := fileSHA256(filepath.Join(smoke.distDir, archive))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(smoke.distDir, checksum), []byte(fmt.Sprintf("%s  %s\n", digest, archive)), 0o644); err != nil {
		return err
	}
	sourceInstall, err := os.ReadFile(filepath.Join(smoke.repoRoot, "scripts", "install.sh"))
	if err != nil {
		return err
	}
	renderedInstall := strings.ReplaceAll(string(sourceInstall), "__OPENCLERK_VERSION__", smokeVersion)
	if err := os.WriteFile(filepath.Join(smoke.distDir, "install.sh"), []byte(renderedInstall), 0o755); err != nil {
		return err
	}
	smoke.archive = archive
	smoke.checksum = checksum
	return writeCurlShim(smoke)
}

func prepareSmokeEnvironment(smoke *smokeContext) error {
	for _, dir := range []string{filepath.Join(smoke.homeDir, ".local", "bin"), smoke.codexHome, filepath.Join(smoke.workDir, "tmp")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	pathValue := strings.Join([]string{smoke.toolsDir, smoke.installDir, systemPath()}, string(os.PathListSeparator))
	smoke.env = append(filteredEnv(os.Environ(), "HOME", "CODEX_HOME", "PATH", "TMPDIR", "OPENCLERK_DATABASE_PATH", "OPENCLERK_INSTALL_DIR", "OPENCLERK_VERSION"),
		"HOME="+smoke.homeDir,
		"CODEX_HOME="+smoke.codexHome,
		"PATH="+pathValue,
		"TMPDIR="+filepath.Join(smoke.workDir, "tmp"),
		"OPENCLERK_DATABASE_PATH="+smoke.dbPath,
	)
	return nil
}

func runInstaller(ctx context.Context, smoke smokeContext, upgrade bool) (installCheck, error) {
	invocation := `sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"`
	cmd := exec.CommandContext(ctx, "sh", "-c", `sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"`)
	cmd.Dir = smoke.repoRoot
	cmd.Env = smoke.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		action := "install"
		if upgrade {
			action = "upgrade"
		}
		return installCheck{}, fmt.Errorf("%s installer failed: %w: %s", action, err, strings.TrimSpace(string(output)))
	}
	commandPath, err := runTrimmed(ctx, smoke, "sh", "-c", "command -v openclerk")
	if err != nil {
		return installCheck{}, err
	}
	wantPath := filepath.Join(smoke.installDir, "openclerk")
	if commandPath != wantPath {
		return installCheck{}, fmt.Errorf("command -v openclerk = %q, want %q", commandPath, wantPath)
	}
	versionOutput, err := runTrimmed(ctx, smoke, filepath.Join(smoke.installDir, "openclerk"), "--version")
	if err != nil {
		return installCheck{}, err
	}
	if versionOutput != "openclerk "+smokeVersion {
		return installCheck{}, fmt.Errorf("openclerk --version = %q, want %q", versionOutput, "openclerk "+smokeVersion)
	}
	if _, err := runTrimmed(ctx, smoke, filepath.Join(smoke.installDir, "openclerk"), "--help"); err != nil {
		return installCheck{}, err
	}
	return installCheck{
		Passed:              true,
		InstallerInvocation: invocation,
		BinaryPath:          "$HOME/.local/bin/openclerk",
		CommandPath:         "$HOME/.local/bin/openclerk",
		VersionOutput:       versionOutput,
		HelpChecked:         true,
	}, nil
}

func installSkill(smoke smokeContext) (skillCheck, error) {
	src := filepath.Join(smoke.repoRoot, "skills", "openclerk")
	dst := filepath.Join(smoke.codexHome, "skills", "openclerk")
	if err := os.RemoveAll(dst); err != nil {
		return skillCheck{}, err
	}
	if err := copyDir(src, dst); err != nil {
		return skillCheck{}, err
	}
	srcBytes, err := os.ReadFile(filepath.Join(src, "SKILL.md"))
	if err != nil {
		return skillCheck{}, err
	}
	dstBytes, err := os.ReadFile(filepath.Join(dst, "SKILL.md"))
	if err != nil {
		return skillCheck{}, err
	}
	if !bytes.Equal(srcBytes, dstBytes) {
		return skillCheck{}, errors.New("installed skill does not match source skills/openclerk/SKILL.md")
	}
	return skillCheck{
		Passed:    true,
		SkillPath: "$CODEX_HOME/skills/openclerk/SKILL.md",
		Source:    "skills/openclerk/SKILL.md",
	}, nil
}

func replaceWithOldBinary(installDir string) error {
	oldBinary := filepath.Join(installDir, "openclerk")
	return os.WriteFile(oldBinary, []byte("#!/bin/sh\nprintf '%s\\n' 'openclerk v0.0.0-old'\n"), 0o755)
}

func runModuleSmoke(ctx context.Context, smoke smokeContext) (moduleSmoke, error) {
	installResult, err := runModuleCommand(ctx, smoke, `{"action":"install_module","module":{"provider":"ollama","manifest_path":"modules/ollama-embeddings/module.json","command":"semantic-retrieval-adapter","provider_config":{"embedding_model":"embeddinggemma","ollama_url":"http://localhost:11434"}}}`)
	if err != nil {
		return moduleSmoke{}, err
	}
	if installResult.Rejected || installResult.Module == nil {
		return moduleSmoke{}, fmt.Errorf("module install rejected: %s", installResult.RejectionReason)
	}
	if installResult.Module.Provider != "ollama" ||
		installResult.Module.ManifestPath != "modules/ollama-embeddings/module.json" ||
		installResult.Module.Command != "semantic-retrieval-adapter" ||
		!installResult.Module.Enabled ||
		installResult.Module.ProviderConfig["embedding_model"] != "embeddinggemma" ||
		installResult.Module.ProviderConfig["ollama_url"] != "http://localhost:11434" ||
		installResult.Module.VerificationStatus != "verified" ||
		installResult.Module.RedactionStatus != "redacted" {
		return moduleSmoke{}, fmt.Errorf("unexpected module install result: %+v", installResult.Module)
	}

	configureResult, err := runModuleCommand(ctx, smoke, `{"action":"configure_module","config":{"provider":"ollama","enabled":false,"provider_config":{"embedding_model":"nomic-embed-text"}}}`)
	if err != nil {
		return moduleSmoke{}, err
	}
	if configureResult.Rejected || configureResult.Module == nil {
		return moduleSmoke{}, fmt.Errorf("module configure rejected: %s", configureResult.RejectionReason)
	}
	if configureResult.Module.Enabled || configureResult.Module.ProviderConfig["embedding_model"] != "nomic-embed-text" {
		return moduleSmoke{}, fmt.Errorf("unexpected module configure result: %+v", configureResult.Module)
	}

	listResult, err := runModuleCommandInDir(ctx, smoke, filepath.Join(smoke.workDir, "non-repo-cwd"), `{"action":"list_modules"}`)
	if err != nil {
		return moduleSmoke{}, err
	}
	if len(listResult.Modules) != 1 ||
		listResult.Modules[0].Provider != "ollama" ||
		listResult.Modules[0].Enabled ||
		listResult.Modules[0].ProviderConfig["embedding_model"] != "nomic-embed-text" {
		return moduleSmoke{}, fmt.Errorf("unexpected module list result: %+v", listResult.Modules)
	}
	preservedConfig := listResult.Modules[0].ProviderConfig
	upgradeResult, err := runModuleCommand(ctx, smoke, fmt.Sprintf(`{"action":"install_module","module":{"provider":"ollama","manifest_path":"modules/ollama-embeddings/module.json","command":"semantic-retrieval-adapter","enabled":false,"provider_config":{"embedding_model":%q,"ollama_url":%q}}}`, preservedConfig["embedding_model"], preservedConfig["ollama_url"]))
	if err != nil {
		return moduleSmoke{}, err
	}
	if upgradeResult.Rejected || upgradeResult.Module == nil {
		return moduleSmoke{}, fmt.Errorf("module upgrade refresh rejected: %s", upgradeResult.RejectionReason)
	}
	if upgradeResult.Module.Provider != "ollama" ||
		upgradeResult.Module.Enabled ||
		upgradeResult.Module.ProviderConfig["embedding_model"] != "nomic-embed-text" ||
		upgradeResult.Module.ProviderConfig["ollama_url"] != "http://localhost:11434" ||
		upgradeResult.Module.VerificationStatus != "verified" ||
		upgradeResult.Module.RedactionStatus != "redacted" {
		return moduleSmoke{}, fmt.Errorf("unexpected module upgrade refresh result: %+v", upgradeResult.Module)
	}
	upgradeList, err := runModuleCommand(ctx, smoke, `{"action":"list_modules"}`)
	if err != nil {
		return moduleSmoke{}, err
	}
	if len(upgradeList.Modules) != 1 ||
		upgradeList.Modules[0].Provider != "ollama" ||
		upgradeList.Modules[0].ProviderConfig["embedding_model"] != "nomic-embed-text" ||
		upgradeList.Modules[0].ProviderConfig["ollama_url"] != "http://localhost:11434" {
		return moduleSmoke{}, fmt.Errorf("unexpected module list after upgrade refresh: %+v", upgradeList.Modules)
	}

	removeResult, err := runModuleCommand(ctx, smoke, `{"action":"remove_module","provider":"ollama"}`)
	if err != nil {
		return moduleSmoke{}, err
	}
	if removeResult.Rejected || removeResult.Module == nil || removeResult.Module.Enabled {
		return moduleSmoke{}, fmt.Errorf("unexpected module remove result: %+v", removeResult)
	}
	finalList, err := runModuleCommand(ctx, smoke, `{"action":"list_modules"}`)
	if err != nil {
		return moduleSmoke{}, err
	}
	if len(finalList.Modules) != 0 {
		return moduleSmoke{}, fmt.Errorf("module list after remove is not empty: %+v", finalList.Modules)
	}
	return moduleSmoke{
		Passed:            true,
		Provider:          "ollama",
		ManifestPath:      "modules/ollama-embeddings/module.json",
		SkillPath:         "modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md",
		InstallPassed:     true,
		ConfigurePassed:   true,
		UpgradePassed:     true,
		UpgradePreserved:  true,
		ListPassed:        true,
		RemovePassed:      true,
		FinalListEmpty:    true,
		ProviderConfig:    map[string]string{"embedding_model": "nomic-embed-text", "ollama_url": "http://localhost:11434"},
		VerificationState: "verified",
		RedactionState:    "redacted",
	}, nil
}

type moduleCommandResult struct {
	Rejected        bool                   `json:"rejected"`
	RejectionReason string                 `json:"rejection_reason"`
	Module          *semanticModuleConfig  `json:"module"`
	Modules         []semanticModuleConfig `json:"modules"`
	Summary         string                 `json:"summary"`
}

type semanticModuleConfig struct {
	Provider           string            `json:"provider"`
	ModuleName         string            `json:"module_name"`
	ManifestPath       string            `json:"manifest_path"`
	Command            string            `json:"command"`
	Enabled            bool              `json:"enabled"`
	ProviderConfig     map[string]string `json:"provider_config"`
	VerificationStatus string            `json:"verification_status"`
	RedactionStatus    string            `json:"redaction_status"`
}

func runModuleCommand(ctx context.Context, smoke smokeContext, request string) (moduleCommandResult, error) {
	return runModuleCommandInDir(ctx, smoke, smoke.repoRoot, request)
}

func runModuleCommandInDir(ctx context.Context, smoke smokeContext, dir string, request string) (moduleCommandResult, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return moduleCommandResult{}, err
	}
	cmd := exec.CommandContext(ctx, filepath.Join(smoke.installDir, "openclerk"), "module", "--db", smoke.dbPath)
	cmd.Dir = dir
	cmd.Env = smoke.env
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return moduleCommandResult{}, fmt.Errorf("openclerk module command failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	var result moduleCommandResult
	if err := json.Unmarshal(output, &result); err != nil {
		return moduleCommandResult{}, fmt.Errorf("decode module command output: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return result, nil
}

func runTrimmed(ctx context.Context, smoke smokeContext, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = smoke.repoRoot
	cmd.Env = smoke.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func installerPlatform() (string, string, error) {
	osName := runtime.GOOS
	if osName != "darwin" && osName != "linux" {
		return "", "", fmt.Errorf("unsupported installer OS %q", osName)
	}
	arch := runtime.GOARCH
	if arch != "amd64" && arch != "arm64" {
		return "", "", fmt.Errorf("unsupported installer architecture %q", arch)
	}
	return osName, arch, nil
}

func systemPath() string {
	if runtime.GOOS == "darwin" {
		return "/usr/bin:/bin:/usr/sbin:/sbin"
	}
	return "/usr/bin:/bin"
}

func filteredEnv(env []string, keys ...string) []string {
	blocked := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		blocked[key] = struct{}{}
	}
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, found := strings.Cut(entry, "=")
		if found {
			if _, ok := blocked[key]; ok {
				continue
			}
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func fileSHA256(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func writeCurlShim(smoke *smokeContext) error {
	script := fmt.Sprintf(`#!/bin/sh
set -eu
dist=%q
output=""
url=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      shift
      output="$1"
      ;;
    http://*|https://*)
      url="$1"
      ;;
  esac
  shift || true
done
[ -n "$url" ] || { echo "curl shim missing URL" >&2; exit 2; }
case "$url" in
  *"/releases/latest")
    printf '%%s\n' '{"tag_name":"%s"}'
    ;;
  *"/install.sh")
    if [ -n "$output" ]; then
      cp "$dist/install.sh" "$output"
    else
      cat "$dist/install.sh"
    fi
    ;;
  *)
    file="${url##*/}"
    [ -f "$dist/$file" ] || { echo "curl shim missing fixture asset: $file" >&2; exit 2; }
    cp "$dist/$file" "$output"
    ;;
esac
`, smoke.distDir, smokeVersion)
	return os.WriteFile(filepath.Join(smoke.toolsDir, "curl"), []byte(script), 0o755)
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode().Perm())
	})
}

func writeJSON(path string, report smokeReport) error {
	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func writeMarkdown(path string, report smokeReport) error {
	var b strings.Builder
	b.WriteString("# OpenClerk Live Install/Upgrade/Module Smoke\n\n")
	b.WriteString("This report records a local temp HOME/CODEX_HOME smoke validation for install, upgrade, skill registration, and module agent install commands.\n\n")
	b.WriteString("## Result\n\n")
	fmt.Fprintf(&b, "- Version: `%s`\n", report.Version)
	fmt.Fprintf(&b, "- Install: `%t`\n", report.Install.Passed)
	fmt.Fprintf(&b, "- Upgrade: `%t`\n", report.Upgrade.Passed)
	fmt.Fprintf(&b, "- Skill: `%t`\n", report.Skill.Passed)
	fmt.Fprintf(&b, "- Module install/config/upgrade/list/remove: `%t`\n\n", report.Module.Passed)
	b.WriteString("## Evidence\n\n")
	fmt.Fprintf(&b, "- Installer invocation: `%s`\n", report.Install.InstallerInvocation)
	fmt.Fprintf(&b, "- Binary path: `%s`\n", report.Install.BinaryPath)
	fmt.Fprintf(&b, "- Command path: `%s`\n", report.Install.CommandPath)
	fmt.Fprintf(&b, "- Version output: `%s`\n", report.Install.VersionOutput)
	fmt.Fprintf(&b, "- Skill path: `%s`\n", report.Skill.SkillPath)
	fmt.Fprintf(&b, "- Module manifest: `%s`\n", report.Module.ManifestPath)
	fmt.Fprintf(&b, "- Module skill: `%s`\n", report.Module.SkillPath)
	fmt.Fprintf(&b, "- Module provider config: `embedding_model=%s`, `ollama_url=%s`\n", report.Module.ProviderConfig["embedding_model"], report.Module.ProviderConfig["ollama_url"])
	fmt.Fprintf(&b, "- Module upgrade preserved config: `%t`\n", report.Module.UpgradePreserved)
	fmt.Fprintf(&b, "- Module verification/redaction: `%s` / `%s`\n\n", report.Module.VerificationState, report.Module.RedactionState)
	b.WriteString("## Boundaries\n\n")
	fmt.Fprintf(&b, "%s.\n", report.ValidationBoundaries)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
