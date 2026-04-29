package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseRunConfigDefaultsParallelAndSharedCache(t *testing.T) {
	config, err := parseRunConfig(nil, &strings.Builder{})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if config.Parallel != defaultParallel {
		t.Fatalf("parallel = %d, want %d", config.Parallel, defaultParallel)
	}
	if config.CacheMode != cacheModeShared {
		t.Fatalf("cache mode = %q, want %q", config.CacheMode, cacheModeShared)
	}
}

func TestParseRunConfigRejectsInvalidParallelAndCacheMode(t *testing.T) {
	if _, err := parseRunConfig([]string{"--parallel", "0"}, &strings.Builder{}); err == nil {
		t.Fatal("expected invalid parallel error")
	}
	if _, err := parseRunConfig([]string{"--cache-mode", "bad"}, &strings.Builder{}); err == nil || !strings.Contains(err.Error(), "--cache-mode") {
		t.Fatalf("cache-mode error = %v, want validation error", err)
	}
}

func TestRunJobsPreservesDeterministicOrder(t *testing.T) {
	jobs := []evalJob{
		{Index: 0, Variant: "production", Scenario: scenario{ID: "first", Title: "First"}},
		{Index: 1, Variant: "production", Scenario: scenario{ID: "second", Title: "Second"}},
		{Index: 2, Variant: "variant-smoke", Scenario: scenario{ID: "third", Title: "Third"}},
	}
	results := runJobs(context.Background(), runConfig{Parallel: 3}, jobs, cacheConfig{Mode: cacheModeIsolated}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		if job.Index == 0 {
			time.Sleep(30 * time.Millisecond)
		}
		return jobResult{
			Variant:  job.Variant,
			Scenario: job.Scenario.ID,
			Status:   "completed",
		}
	})
	for i, result := range results {
		if result.Scenario != jobs[i].Scenario.ID {
			t.Fatalf("result %d scenario = %q, want %q", i, result.Scenario, jobs[i].Scenario.ID)
		}
	}
}

func TestRunJobsPreservesErrorIdentity(t *testing.T) {
	jobs := []evalJob{
		{Index: 0, Variant: "production", Scenario: scenario{ID: "ok", Title: "OK"}},
		{Index: 1, Variant: "variant-smoke", Scenario: scenario{ID: "bad", Title: "Bad"}},
	}
	results := runJobs(context.Background(), runConfig{Parallel: 2}, jobs, cacheConfig{Mode: cacheModeIsolated}, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		result := jobResult{Variant: job.Variant, Scenario: job.Scenario.ID}
		if job.Scenario.ID == "bad" {
			result.Status = "failed"
			result.Error = "boom"
			return result
		}
		result.Status = "completed"
		return result
	})
	if results[1].Variant != "variant-smoke" || results[1].Scenario != "bad" || results[1].Error != "boom" {
		t.Fatalf("error result = %+v", results[1])
	}
}

func TestCodexArgsForSingleAndResumedTurns(t *testing.T) {
	cache := cacheConfig{Mode: cacheModeShared, RunRoot: "run-root"}
	single := scenario{ID: "single", Prompt: "single prompt"}
	singleArgs := codexArgsForTurn("codex", "run-root/production/single/repo", "run-root/production/single", single, scenarioTurn{Prompt: "single prompt"}, 1, "", cache)
	if !containsArgPair(singleArgs, "--add-dir", filepath.Join("run-root", "shared-cache")) {
		t.Fatalf("single args missing shared cache writable root: %v", singleArgs)
	}
	if !containsValue(singleArgs, "--ephemeral") {
		t.Fatalf("single args must use --ephemeral: %v", singleArgs)
	}
	if !containsValue(singleArgs, "--ignore-user-config") {
		t.Fatalf("single args missing --ignore-user-config: %v", singleArgs)
	}

	multi := scenario{ID: "multi", Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
	firstArgs := codexArgsForTurn("codex", "run-root/production/multi/repo", "run-root/production/multi", multi, scenarioTurn{Prompt: "first"}, 1, "", cache)
	if !containsValue(firstArgs, "--ignore-user-config") {
		t.Fatalf("first multi-turn args missing --ignore-user-config: %v", firstArgs)
	}
	if containsValue(firstArgs, "--ephemeral") {
		t.Fatalf("first multi-turn args must persist the session: %v", firstArgs)
	}
	resumeArgs := codexArgsForTurn("codex", "run-root/production/multi/repo", "run-root/production/multi", multi, scenarioTurn{Prompt: "second"}, 2, "session-123", cache)
	if containsValue(resumeArgs, "--ephemeral") {
		t.Fatalf("resume args must not use --ephemeral: %v", resumeArgs)
	}
	if !containsValue(resumeArgs, "--ignore-user-config") {
		t.Fatalf("resume args missing --ignore-user-config: %v", resumeArgs)
	}
	if !containsValue(resumeArgs, "resume") || !containsValue(resumeArgs, "session-123") {
		t.Fatalf("resume args must persist the multi-turn session: %v", resumeArgs)
	}
}

func TestEvalEnvSharedAndIsolatedCache(t *testing.T) {
	paths := scenarioPaths(filepath.Join("run-root", "production", "create", "repo"))
	shared := strings.Join(evalEnv(filepath.Join("run-root", "production", "create"), paths, cacheConfig{Mode: cacheModeShared, RunRoot: "run-root"}), "\n")
	for _, want := range []string{
		"OPENCLERK_DATABASE_PATH=" + filepath.Join("run-root", "production", "create", "repo", ".openclerk-eval", "openclerk.db"),
		"GOCACHE=" + filepath.Join("run-root", "shared-cache", "gocache"),
		"GOMODCACHE=" + filepath.Join("run-root", "shared-cache", "gomodcache"),
	} {
		if !strings.Contains(shared, want) {
			t.Fatalf("shared env missing %q in %s", want, shared)
		}
	}
	isolated := strings.Join(evalEnv(filepath.Join("run-root", "production", "create"), paths, cacheConfig{Mode: cacheModeIsolated, RunRoot: "run-root"}), "\n")
	if !strings.Contains(isolated, "GOCACHE="+filepath.Join("run-root", "production", "create", "gocache")) {
		t.Fatalf("isolated env = %s", isolated)
	}
}

func TestCopyRepoSkipsEvalContextContamination(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "copy")
	files := map[string]string{
		"README.md":                                      "kept",
		"AGENTS.md":                                      "root instructions",
		".agents/skills/openclerk/SKILL.md":              "stale skill",
		".dolt/config":                                   "dolt",
		"docs/evals/results/previous.md":                 "report",
		"scripts/agent-eval/ockp/main.go":                "harness",
		"scripts/agent-eval/ockp/nested/fixture.txt":     "harness fixture",
		"scripts/agent-eval/other-harness/kept-file.txt": "other harness",
	}
	for path, content := range files {
		target := filepath.Join(src, path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	if err := copyRepo(src, dst); err != nil {
		t.Fatalf("copy repo: %v", err)
	}
	for _, want := range []string{"README.md", "scripts/agent-eval/other-harness/kept-file.txt"} {
		if _, err := os.Stat(filepath.Join(dst, want)); err != nil {
			t.Fatalf("expected copied %s: %v", want, err)
		}
	}
	for _, skipped := range []string{
		"AGENTS.md",
		".agents/skills/openclerk/SKILL.md",
		".dolt/config",
		"docs/evals/results/previous.md",
		"scripts/agent-eval/ockp/main.go",
		"scripts/agent-eval/ockp/nested/fixture.txt",
	} {
		if _, err := os.Stat(filepath.Join(dst, skipped)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be skipped, stat err=%v", skipped, err)
		}
	}
}

func TestInstallVariantInstallsExactSkillAndNoAgentsFile(t *testing.T) {
	repoRoot := t.TempDir()
	repoDir := t.TempDir()
	sourceSkill := []byte("# OpenClerk\n\nUse the runner.\n")
	skillDir := filepath.Join(repoRoot, "skills", "openclerk")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), sourceSkill, 0o644); err != nil {
		t.Fatalf("write source skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill_payload_test.go"), []byte("package ignored\n"), 0o644); err != nil {
		t.Fatalf("write source test helper: %v", err)
	}

	if err := installVariant(repoRoot, repoDir, productionVariant); err != nil {
		t.Fatalf("install production variant: %v", err)
	}
	installedSkill, err := os.ReadFile(filepath.Join(repoDir, ".agents", "skills", "openclerk", "SKILL.md"))
	if err != nil {
		t.Fatalf("read installed skill: %v", err)
	}
	if !bytes.Equal(installedSkill, sourceSkill) {
		t.Fatalf("installed skill bytes = %q, want %q", installedSkill, sourceSkill)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("installVariant must not create AGENTS.md, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".agents", "skills", "openclerk", "skill_payload_test.go")); !os.IsNotExist(err) {
		t.Fatalf("installVariant must not install Go test payload, stat err=%v", err)
	}
}

func TestVariantSelectionProductionOnly(t *testing.T) {
	variants := selectedVariants(runConfig{})
	if len(variants) != 1 || variants[0] != productionVariant {
		t.Fatalf("default variants = %v, want [%s]", variants, productionVariant)
	}

	repoRoot := t.TempDir()
	repoDir := t.TempDir()
	if err := installVariant(repoRoot, repoDir, "unknown"); err == nil {
		t.Fatal("expected unknown variant error")
	}
}

func TestEvalEnvOverridesHostOpenClerkPaths(t *testing.T) {
	t.Setenv("CODEX_HOME", "/host/.codex")
	t.Setenv("OPENCLERK_DATA_DIR", "/host/data")
	t.Setenv("OPENCLERK_DATABASE_PATH", "/host/openclerk.db")
	t.Setenv("OPENCLERK_VAULT_ROOT", "/host/vault")

	runDir := filepath.Join(t.TempDir(), "run")
	paths := evalPaths{
		DatabasePath: filepath.Join(runDir, "openclerk.db"),
	}
	env := evalEnv(runDir, paths, cacheConfig{})
	got := map[string]string{}
	for _, entry := range env {
		key, value, found := strings.Cut(entry, "=")
		if found {
			got[key] = value
		}
	}
	if got["CODEX_HOME"] != filepath.Join(runDir, "codex-home") {
		t.Fatalf("CODEX_HOME = %q, want %q", got["CODEX_HOME"], filepath.Join(runDir, "codex-home"))
	}
	if got["ZDOTDIR"] != filepath.Join(runDir, "zdotdir") {
		t.Fatalf("ZDOTDIR = %q, want %q", got["ZDOTDIR"], filepath.Join(runDir, "zdotdir"))
	}
	if _, ok := got["OPENCLERK_DATA_DIR"]; ok {
		t.Fatalf("OPENCLERK_DATA_DIR should not be exported: %v", got)
	}
	if got["OPENCLERK_DATABASE_PATH"] != paths.DatabasePath {
		t.Fatalf("OPENCLERK_DATABASE_PATH = %q, want %q", got["OPENCLERK_DATABASE_PATH"], paths.DatabasePath)
	}
	if _, ok := got["OPENCLERK_VAULT_ROOT"]; ok {
		t.Fatalf("OPENCLERK_VAULT_ROOT should not be exported: %v", got)
	}
}

func TestPromptInputPreflightFlagsOpenClerkAgentsInstructions(t *testing.T) {
	clean := "### Project Skills\n- OpenClerk: Use OpenClerk for local-first knowledge-plane tasks through the installed openclerk JSON runner. Bootstrap no-tools rule - if required retrieval, source, or video fields are missing; if document path, title, or body is missing and no faithful propose-before-create candidate can be formed from explicit user content; if a numeric limit is negative; or if the user asks to bypass the runner with SQLite, raw vault/file/repo inspection, HTTP/MCP, legacy/source-built paths, unsupported transports, backend variants, module-cache inspection, rg, find, or ls, this description is complete; do not open this skill file, run commands, use tools, or call the runner; respond with exactly one no-tools assistant answer that names the missing/invalid fields or rejects the unsupported workflow. For valid work, use only openclerk document or openclerk retrieval JSON. (file: /tmp/repo/.agents/skills/openclerk/SKILL.md)\n"
	if !containsOpenClerkSkillDiscovery(clean) {
		t.Fatalf("clean skill discovery is missing the OpenClerk skill marker: %s", clean)
	}
	if containsOpenClerkAgentsInstructions(clean) {
		t.Fatalf("clean skill discovery was flagged: %s", clean)
	}
	if !containsOpenClerkBootstrapRejectionGuidance(clean) {
		t.Fatalf("clean skill discovery is missing bootstrap rejection guidance: %s", clean)
	}

	missingBootstrap := "### Project Skills\n- openclerk: Use OpenClerk. (file: /tmp/repo/.agents/skills/openclerk/SKILL.md)\n"
	if containsOpenClerkBootstrapRejectionGuidance(missingBootstrap) {
		t.Fatalf("incomplete skill discovery passed bootstrap guidance check: %s", missingBootstrap)
	}

	contaminated := "# AGENTS.md instructions for /tmp/repo\n\nUse `openclerk document` with create_document JSON action names.\n"
	if !containsOpenClerkAgentsInstructions(contaminated) {
		t.Fatalf("contaminated AGENTS block was not flagged: %s", contaminated)
	}
}

func TestPrepareRunDirCreatesAuthOnlyCodexHome(t *testing.T) {
	hostCodexHome := filepath.Join(t.TempDir(), "host-codex-home")
	if err := os.MkdirAll(filepath.Join(hostCodexHome, "sessions"), 0o755); err != nil {
		t.Fatalf("mkdir host codex home: %v", err)
	}
	for name, content := range map[string]string{
		"auth.json":          `{"token":"test"}`,
		"config.toml":        "model = \"gpt-5.4-mini\"\n",
		"installation_id":    "test-installation\n",
		"sessions/old.jsonl": "old session\n",
	} {
		if err := os.WriteFile(filepath.Join(hostCodexHome, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write host codex home %s: %v", name, err)
		}
	}
	t.Setenv("CODEX_HOME", hostCodexHome)

	runDir := filepath.Join(t.TempDir(), "run")
	if err := prepareRunDir(runDir, cacheConfig{}); err != nil {
		t.Fatalf("prepare run dir: %v", err)
	}
	for _, want := range []string{
		filepath.Join(runDir, "codex-home"),
		filepath.Join(runDir, "zdotdir"),
		filepath.Join(runDir, "tmp"),
	} {
		if info, err := os.Stat(want); err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s, stat err=%v", want, err)
		}
	}
	got, err := os.ReadFile(filepath.Join(runDir, "codex-home", "auth.json"))
	if err != nil {
		t.Fatalf("read seeded auth: %v", err)
	}
	if string(got) != `{"token":"test"}` {
		t.Fatalf("seeded auth = %q, want copied auth", got)
	}
	homeInfo, err := os.Stat(filepath.Join(runDir, "codex-home"))
	if err != nil {
		t.Fatalf("stat eval codex home: %v", err)
	}
	if homeInfo.Mode().Perm()&0o077 != 0 {
		t.Fatalf("eval codex home permissions = %v, want no group/other access", homeInfo.Mode().Perm())
	}
	for _, unwanted := range []string{"config.toml", "installation_id", filepath.Join("sessions", "old.jsonl")} {
		if _, err := os.Stat(filepath.Join(runDir, "codex-home", unwanted)); !os.IsNotExist(err) {
			t.Fatalf("unexpected copied %s: stat error = %v", unwanted, err)
		}
	}
}

func TestSetupEvalCodexHomeRequiresAuth(t *testing.T) {
	err := setupEvalCodexHomeFromSource(filepath.Join(t.TempDir(), "codex-home"), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "run codex login") {
		t.Fatalf("setupEvalCodexHomeFromSource() error = %v, want login guidance", err)
	}
}

func TestGoListDoesNotExposePublicLocalClientPackage(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..", "..")
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list ./...: %v\n%s", err, string(output))
	}
	forbiddenPackage := "github.com/yazanabuashour/openclerk/" + "client/" + "local"
	if strings.Contains(string(output), forbiddenPackage) {
		t.Fatalf("go list exposes removed package %q:\n%s", forbiddenPackage, string(output))
	}
}

func TestRepositoryDoesNotDocumentRemovedPublicClientInterface(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..", "..")
	forbidden := []string{
		"client/" + "local",
		"local." + "Open" + "Client",
		"Open" + "Client",
		"sdk-" + "baseline",
		"Local Go " + "S" + "DK",
		"direct-local Go " + "package",
		"examples/openclerk-" + "client",
		"examples/openclerk-" + "query",
	}
	if err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".beads", ".openclerk-eval":
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".md", ".json", ".yaml", ".yml", ".toml", ".txt":
		default:
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(content)
		displayPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			displayPath = path
		}
		for _, stale := range forbidden {
			if strings.Contains(text, stale) {
				t.Fatalf("%s contains removed public client interface text %q", displayPath, stale)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("scan repo: %v", err)
	}
}
