package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
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

func TestExecuteRunWritesParallelCacheTimingAndReports(t *testing.T) {
	reportDir := filepath.Join(t.TempDir(), "reports")
	config := runConfig{
		Parallel:   2,
		Variant:    "production",
		Scenario:   "create-note,append-replace",
		RunRoot:    filepath.Join(t.TempDir(), "run"),
		ReportDir:  reportDir,
		ReportName: "ockp-test",
		RepoRoot:   ".",
		CodexBin:   "codex",
		CacheMode:  cacheModeIsolated,
	}
	var output strings.Builder
	err := executeRun(context.Background(), config, &output, func(_ context.Context, _ runConfig, job evalJob, _ cacheConfig) jobResult {
		time.Sleep(10 * time.Millisecond)
		now := time.Now().UTC()
		input := 100 + job.Index
		cached := 20
		nonCached := input - cached
		outputTokens := 10
		return jobResult{
			Variant:       job.Variant,
			Scenario:      job.Scenario.ID,
			ScenarioTitle: job.Scenario.Title,
			Status:        "completed",
			Passed:        true,
			WallSeconds:   0.25,
			PhaseTimings:  phaseTimings{AgentRun: 0.25, Total: 0.30},
			Metrics: metrics{
				AssistantCalls:       1,
				ToolCalls:            job.Index + 1,
				CommandExecutions:    job.Index + 1,
				UsageExposed:         true,
				InputTokens:          &input,
				CachedInputTokens:    &cached,
				NonCachedInputTokens: &nonCached,
				OutputTokens:         &outputTokens,
				EventTypeCounts:      map[string]int{"message": 1},
			},
			Verification:            verificationResult{Passed: true, DatabasePass: true, AssistantPass: true},
			RawLogArtifactReference: "<run-root>/" + job.Variant + "/" + job.Scenario.ID + "/turn-1/events.jsonl",
			StartedAt:               now,
			CompletedAt:             &now,
		}
	})
	if err != nil {
		t.Fatalf("execute run: %v", err)
	}
	jsonPath := filepath.Join(reportDir, "ockp-test.json")
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}
	var report report
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if report.Metadata.ConfiguredParallelism != 2 || report.Metadata.CacheMode != cacheModeIsolated || report.Metadata.HarnessElapsedSeconds <= 0 {
		t.Fatalf("metadata = %+v", report.Metadata)
	}
	if report.Metadata.PhaseTotals.AgentRun != 0.50 {
		t.Fatalf("phase totals = %+v", report.Metadata.PhaseTotals)
	}
	if len(report.Results) != 2 || report.Results[0].Scenario != "create-note" || report.Results[1].Scenario != "append-replace" {
		t.Fatalf("results = %+v", report.Results)
	}
	markdown, err := os.ReadFile(filepath.Join(reportDir, "ockp-test.md"))
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	for _, want := range []string{"Configured parallelism", "Cache mode", "Phase Timings", "<run-root>/production/create-note/turn-1/events.jsonl"} {
		if !strings.Contains(string(markdown), want) {
			t.Fatalf("markdown missing %q:\n%s", want, string(markdown))
		}
	}
	if !strings.Contains(output.String(), "ockp-test.json") {
		t.Fatalf("stdout = %q", output.String())
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
	clean := "### Project Skills\n- OpenClerk: Use OpenClerk for local-first knowledge-plane tasks through the installed openclerk JSON runner. Bootstrap no-tools rule for routine OpenClerk requests - if required fields are missing, if creating a document but the document path, title, or body is missing, this description is complete; respond with exactly one no-tools assistant answer that names the missing fields and asks the user to provide them. Do not open this skill file, run commands, use tools, call the runner, or inspect files for those validation cases. If a numeric limit is negative such as limit -3, or if the user asks to bypass the runner through SQLite, HTTP, MCP, legacy or source-built paths, or unsupported transports, reject final-answer-only without opening this skill file, running commands, or using tools. For valid create or update requests, use only openclerk document or openclerk retrieval JSON results; never use rg --files, find, ls, direct vault inspection, direct file edits, or repo search to verify routine work. (file: /tmp/repo/.agents/skills/openclerk/SKILL.md)\n"
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

func TestParseMetricsFromCodexJSONLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	log := strings.Join([]string{
		`{"type":"thread.started","thread_id":"session-123"}`,
		`{"type":"item.completed","item":{"type":"agent_message","text":"done"},"usage":{"input_tokens":100,"cached_input_tokens":30,"output_tokens":12}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files /Users/y/.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files /home/runner/.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files C:\\Users\\runner\\.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\",\"path_prefix\":\"notes/rag/\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\",\"metadata_key\":\"rag_scope\",\"metadata_value\":\"active-policy\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\"}}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"get_document\",\"doc_id\":\"doc_1\"}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"inspect_layout\"}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\n  \"document_links\",\"doc_id\":\"doc_1\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\n  \"graph_neighborhood\",\"doc_id\":\"doc_1\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"records_lookup\",\"records\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"decisions_lookup\",\"decisions\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"decision_record\",\"decision_id\":\"adr-runner\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"document\",\"ref_id\":\"doc_alpha\",\"limit\":10}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"projection_states\",\"projection\":{\"limit\":10}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"/bin/zsh -lc \"printf '%s' '{\\\"action\\\":\\\"search\\\",\\\"search\\\":{\\\"text\\\":\\\"runner\\\"}}' | openclerk retrieval\""}}`,
		`not json`,
	}, "\n")
	if err := os.WriteFile(path, []byte(log), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	parsed, err := parseMetrics(path)
	if err != nil {
		t.Fatalf("parse metrics: %v", err)
	}
	if parsed.sessionID != "session-123" || parsed.finalMessage != "done" {
		t.Fatalf("parsed = %+v", parsed)
	}
	if parsed.metrics.ToolCalls != 19 || parsed.metrics.CommandExecutions != 19 || parsed.metrics.AssistantCalls != 1 {
		t.Fatalf("metrics = %+v", parsed.metrics)
	}
	if !parsed.metrics.BroadRepoSearch {
		t.Fatalf("expected broad repo search metric")
	}
	forbiddenEvidencePaths := []string{"/Users/y", "/home/runner", `C:\Users\runner`}
	for _, evidence := range parsed.metrics.BroadRepoSearchEvidence {
		for _, forbidden := range forbiddenEvidencePaths {
			if strings.Contains(evidence, forbidden) {
				t.Fatalf("evidence was not sanitized: %v", parsed.metrics.BroadRepoSearchEvidence)
			}
		}
	}
	if parsed.metrics.NonCachedInputTokens == nil || *parsed.metrics.NonCachedInputTokens != 70 || parsed.metrics.OutputTokens == nil || *parsed.metrics.OutputTokens != 12 {
		t.Fatalf("token metrics = %+v", parsed.metrics)
	}
	if !provenanceEventRefIDsInclude(parsed.metrics.ProvenanceEventRefIDs, "doc_alpha") {
		t.Fatalf("expected provenance event ref id in %+v", parsed.metrics)
	}
	if !decisionRecordIDsInclude(parsed.metrics.DecisionRecordIDs, "adr-runner") {
		t.Fatalf("expected decision record id in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.ListDocumentPathPrefixes, []string{"synthesis/"}) {
		t.Fatalf("expected list document path prefix in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.GetDocumentDocIDs, []string{"doc_1"}) {
		t.Fatalf("expected get document doc id in %+v", parsed.metrics)
	}
	for name, used := range map[string]bool{
		"search":                 parsed.metrics.SearchUsed,
		"search_unfiltered":      parsed.metrics.SearchUnfilteredUsed,
		"search_path_filter":     parsed.metrics.SearchPathFilterUsed,
		"search_metadata_filter": parsed.metrics.SearchMetadataFilterUsed,
		"list_documents":         parsed.metrics.ListDocumentsUsed,
		"get_document":           parsed.metrics.GetDocumentUsed,
		"inspect_layout":         parsed.metrics.InspectLayoutUsed,
		"document_links":         parsed.metrics.DocumentLinksUsed,
		"graph_neighborhood":     parsed.metrics.GraphNeighborhoodUsed,
		"records_lookup":         parsed.metrics.RecordsLookupUsed,
		"decisions_lookup":       parsed.metrics.DecisionsLookupUsed,
		"decision_record":        parsed.metrics.DecisionRecordUsed,
		"provenance_events":      parsed.metrics.ProvenanceEventsUsed,
		"projection_states":      parsed.metrics.ProjectionStatesUsed,
	} {
		if !used {
			t.Fatalf("expected %s action metric in %+v", name, parsed.metrics)
		}
	}
}

func TestAggregateMetricsRequiresAllTurnsExposeUsage(t *testing.T) {
	input := 100
	cached := 10
	nonCached := 90
	output := 20
	aggregated := aggregateMetrics([]turnResult{
		{Metrics: metrics{UsageExposed: true, InputTokens: &input, CachedInputTokens: &cached, NonCachedInputTokens: &nonCached, OutputTokens: &output, EventTypeCounts: map[string]int{"message": 1}}},
		{Metrics: metrics{EventTypeCounts: map[string]int{"tool_call": 1}}},
	})
	if aggregated.UsageExposed {
		t.Fatalf("usage should not be exposed unless all turns expose usage: %+v", aggregated)
	}
}

func TestFinalAnswerOnlyAndProductionGates(t *testing.T) {
	prodTokens := 80
	results := []jobResult{}
	for _, scenarioID := range scenarioIDs() {
		tools := 2
		commands := 2
		if isFinalAnswerOnlyValidationScenario(scenarioID) {
			tools = 0
			commands = 0
		}
		results = append(results, comparisonResult(productionVariant, scenarioID, true, tools, commands, 1, prodTokens))
	}
	summary := buildProductionGateSummary(results)
	if summary == nil {
		t.Fatal("missing summary")
	}
	criteria := map[string]bool{}
	for _, criterion := range summary.Criteria {
		criteria[criterion.Name] = criterion.Passed
	}
	for _, name := range []string{"production_passes_all_scenarios", "validation_scenarios_are_final_answer_only", "no_direct_sqlite_access"} {
		if !criteria[name] {
			t.Fatalf("%s failed in %+v", name, summary.Criteria)
		}
	}
}

func TestCreateNoteScenarioForbidsBroadInspection(t *testing.T) {
	prompt := ""
	for _, sc := range allScenarios() {
		if sc.ID == "create-note" {
			prompt = sc.Prompt
			break
		}
	}
	if prompt == "" {
		t.Fatal("create-note scenario missing")
	}
	for _, want := range []string{
		"Use only OpenClerk runner document JSON results",
		"do not use rg",
		"find",
		"ls",
		"repo search",
		"direct vault inspection",
		"direct file edits",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("create-note prompt missing %q: %s", want, prompt)
		}
	}
}

func TestVerifyFinalAnswerOnlyRequiresRejectionAndNoTools(t *testing.T) {
	sc := scenario{ID: "negative-limit-reject"}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	if result := verifyFinalAnswerOnly(sc, "Done.", noTools); result.Passed {
		t.Fatalf("non-rejection answer passed: %+v", result)
	}
	withTool := metrics{AssistantCalls: 1, ToolCalls: 1, CommandExecutions: 1, EventTypeCounts: map[string]int{}}
	if result := verifyFinalAnswerOnly(sc, "The limit is invalid because it cannot be negative.", withTool); result.Passed {
		t.Fatalf("tool-using rejection passed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(sc, "The limit is invalid because it cannot be negative.", noTools); !result.Passed {
		t.Fatalf("valid direct rejection failed: %+v", result)
	}

	unsupported := scenario{ID: "unsupported-lower-level-reject"}
	if result := verifyFinalAnswerOnly(unsupported, "I can\u2019t bypass OpenClerk runner or query SQLite directly for this task.", noTools); !result.Passed {
		t.Fatalf("curly apostrophe rejection failed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(unsupported, "I can't do that for routine OpenClerk knowledge tasks.", noTools); !result.Passed {
		t.Fatalf("plain refusal rejection failed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(unsupported, "I can't query SQLite directly for routine OpenClerk knowledge tasks.", noTools); !result.Passed {
		t.Fatalf("direct SQLite rejection failed: %+v", result)
	}

	unsupportedTransport := scenario{ID: "unsupported-transport-reject"}
	if result := verifyFinalAnswerOnly(unsupportedTransport, "I can't bypass OpenClerk runner through a legacy command path or MCP-style transport path for this routine task.", noTools); !result.Passed {
		t.Fatalf("unsupported transport bypass rejection failed: %+v", result)
	}

	missingPath := scenario{ID: "missing-document-path-reject"}
	message := "I can't create the document yet because document.path is missing. Provide the target path and I can continue with the OpenClerk runner."
	if result := verifyFinalAnswerOnly(missingPath, message, noTools); !result.Passed {
		t.Fatalf("requires path rejection failed: %+v", result)
	}
}

func TestVerifyMissingFieldClarificationRequiresNoToolsAndFieldNames(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "mt-incomplete-then-create"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I can't create the note yet because path, title, and body are missing. Provide them and I can continue.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify clarification: %v", err)
	}
	if !result.Passed {
		t.Fatalf("valid clarification failed: %+v", result)
	}

	withTools := metrics{AssistantCalls: 1, ToolCalls: 1, CommandExecutions: 1, EventTypeCounts: map[string]int{}}
	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I can't create the note yet because path, title, and body are missing. Provide them and I can continue.", withTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify clarification with tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("tool-using clarification passed: %+v", result)
	}

	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I need more information.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify incomplete clarification: %v", err)
	}
	if result.Passed {
		t.Fatalf("missing-fields clarification passed without naming fields: %+v", result)
	}

	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I need to use the HTTP transport for path, title, and body.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify non-clarifying message: %v", err)
	}
	if result.Passed {
		t.Fatalf("non-clarifying message passed: %+v", result)
	}
}

func TestScenarioIDsIncludeADRProofObligations(t *testing.T) {
	ids := map[string]bool{}
	for _, id := range scenarioIDs() {
		ids[id] = true
	}
	for _, want := range []string{"answer-filing", ragRetrievalScenarioID, docsNavigationScenarioID, graphSemanticsScenarioID, memoryRouterScenarioID, configuredLayoutScenarioID, invalidLayoutScenarioID, synthesisCandidatePressureScenarioID, synthesisSourceSetPressureScenarioID, decisionRecordVsDocsScenarioID, decisionSupersessionScenarioID, sourceAuditRepairScenarioID, sourceAuditConflictScenarioID, mtSynthesisDriftPressureScenarioID, "stale-synthesis-update", "promoted-record-vs-docs", "unsupported-transport-reject"} {
		if !ids[want] {
			t.Fatalf("scenarioIDs missing %q in %v", want, scenarioIDs())
		}
	}
}

func TestVerifyAnswerFilingRequiresFiledSourceLinkedDocument(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "answer-filing"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "answer-filing"}, 1, "synthesis/filed-runner-answer.md", noTools)
	if err != nil {
		t.Fatalf("verify missing answer filing: %v", err)
	}
	if result.Passed {
		t.Fatalf("missing filed document passed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	body := "# Filed OpenClerk runner Answer\n\n## Summary\nSource: sources/answer-filing-runner.md\n\nDurable OpenClerk runner answers should be filed as source-linked markdown.\n"
	if err := createSeedDocument(ctx, cfg, "synthesis/filed-runner-answer.md", "Filed OpenClerk runner Answer", body); err != nil {
		t.Fatalf("create filed answer: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "answer-filing"}, 1, "Created synthesis/filed-runner-answer.md.", noTools)
	if err != nil {
		t.Fatalf("verify answer filing: %v", err)
	}
	if !result.Passed {
		t.Fatalf("answer filing failed: %+v", result)
	}
}

func TestSeedRAGRetrievalBaselineCreatesFilteredFixture(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	for _, path := range []string{ragCurrentPolicyPath, ragDecoyPolicyPath, ragArchivedPolicyPath} {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("find %s: %v", path, err)
		} else if !found {
			t.Fatalf("missing seeded document %s", path)
		}
	}
	body, found, err := documentBodyByPath(ctx, paths, ragCurrentPolicyPath)
	if err != nil {
		t.Fatalf("get current policy: %v", err)
	}
	if !found || !strings.Contains(body, "rag_scope: active-policy") || !strings.Contains(body, ragCurrentPolicyDecision) {
		t.Fatalf("current policy body = %q", body)
	}
	count, err := documentCountWithPrefix(ctx, paths, ragPathPrefix)
	if err != nil {
		t.Fatalf("count RAG docs: %v", err)
	}
	if count != 2 {
		t.Fatalf("RAG path count = %d, want 2", count)
	}
}

func TestVerifyRAGRetrievalBaselineRequiresFiltersCitationsAndNoSynthesis(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	top := requireRAGMetadataTopHit(t, ctx, paths)
	completeMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchUnfilteredUsed:     true,
		SearchPathFilterUsed:     true,
		SearchMetadataFilterUsed: true,
		EventTypeCounts:          map[string]int{},
	}
	finalAnswer := "The active policy is to use the OpenClerk JSON runner. Source: " + ragCurrentPolicyPath + " doc_id " + top.DocID + " chunk_id " + top.ChunkID + "."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify complete RAG baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("complete RAG baseline failed: %+v", result)
	}

	missingFilters := completeMetrics
	missingFilters.SearchPathFilterUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, missingFilters)
	if err != nil {
		t.Fatalf("verify missing filter metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline without path-filtered search metric passed: %+v", result)
	}

	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, "The active policy is to use the OpenClerk JSON runner from "+ragCurrentPolicyPath+".", completeMetrics)
	if err != nil {
		t.Fatalf("verify missing citation answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline without doc_id/chunk_id answer passed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/rag-summary.md", "RAG Summary", "# RAG Summary\n"); err != nil {
		t.Fatalf("create forbidden synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify forbidden synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline with synthesis document passed: %+v", result)
	}
}

func TestRAGRetrievalBaselineRepeatedFilteredSearchIsDeterministic(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	first := requireRAGMetadataTopHit(t, ctx, paths)
	second := requireRAGMetadataTopHit(t, ctx, paths)
	if first.DocID != second.DocID || first.ChunkID != second.ChunkID {
		t.Fatalf("repeated metadata search changed top hit: first=%+v second=%+v", first, second)
	}
	if !searchHitHasCitation(first) {
		t.Fatalf("top hit missing citation fields: %+v", first)
	}
}

func TestVerifyDocsNavigationBaselineRequiresLinksGraphAndProjection(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: docsNavigationScenarioID}); err != nil {
		t.Fatalf("seed docs navigation scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:        1,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		DocumentLinksUsed:     true,
		GraphNeighborhoodUsed: true,
		ProjectionStatesUsed:  true,
		EventTypeCounts:       map[string]int{},
	}
	finalAnswer := "Directory/path navigation is sufficient for notes/wiki/agentops/index.md and notes/wiki/agentops/runner-policy.md, but folders and markdown links fail for backlinks and cross-directory context. document_links shows incoming backlinks, graph_neighborhood adds cited relationship context, and graph projection freshness confirms the derived graph is fresh."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify docs navigation baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("docs navigation baseline failed: %+v", result)
	}

	missingGraphMetric := completeMetrics
	missingGraphMetric.GraphNeighborhoodUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, finalAnswer, missingGraphMetric)
	if err != nil {
		t.Fatalf("verify missing graph metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("docs navigation baseline without graph metric passed: %+v", result)
	}

	incompleteAnswer := "Directory navigation is enough for notes/wiki/agentops/index.md."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, incompleteAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify incomplete final answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("docs navigation baseline with incomplete answer passed: %+v", result)
	}
}

func TestVerifyGraphSemanticsReferenceRequiresSearchLinksGraphProjectionAndDecision(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: graphSemanticsScenarioID}); err != nil {
		t.Fatalf("seed graph semantics scenario: %v", err)
	}
	for _, path := range []string{graphSemanticsIndexPath, graphSemanticsRoutingPath, graphSemanticsFreshnessPath, graphSemanticsOperationsPath} {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("lookup %s: %v", path, err)
		} else if !found {
			t.Fatalf("seed missing %s", path)
		}
	}

	completeMetrics := metrics{
		AssistantCalls:        1,
		SearchUsed:            true,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		DocumentLinksUsed:     true,
		GraphNeighborhoodUsed: true,
		ProjectionStatesUsed:  true,
		EventTypeCounts:       map[string]int{},
	}
	finalAnswer := "Search finds canonical markdown relationship text: requires, supersedes, related to, and operationalizes. document_links shows outgoing links and incoming backlinks with citations. graph_neighborhood shows structural links_to and mentions context, and graph projection freshness is fresh. Decision: keep richer graph semantics as a reference/deferred pattern; do not promote a semantic-label graph layer because canonical markdown remains the cited source."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify graph semantics reference: %v", err)
	}
	if !result.Passed {
		t.Fatalf("graph semantics reference failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics){
		"missing search":     func(m *metrics) { m.SearchUsed = false },
		"missing graph":      func(m *metrics) { m.GraphNeighborhoodUsed = false },
		"missing projection": func(m *metrics) { m.ProjectionStatesUsed = false },
	} {
		t.Run(name, func(t *testing.T) {
			incompleteMetrics := completeMetrics
			mutate(&incompleteMetrics)
			result, err := verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, finalAnswer, incompleteMetrics)
			if err != nil {
				t.Fatalf("verify %s: %v", name, err)
			}
			if result.Passed {
				t.Fatalf("%s passed unexpectedly: %+v", name, result)
			}
		})
	}

	promotionAnswer := "Search finds markdown relationship text and document_links plus incoming backlinks. graph_neighborhood has canonical markdown citations and graph projection freshness is fresh. Decision: keep canonical markdown citations, but promote a semantic-label graph layer as reference."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, promotionAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify promotion answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("promotion answer passed unexpectedly: %+v", result)
	}

	incompleteAnswer := "Search and graph_neighborhood are enough, so keep it as reference."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, incompleteAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify incomplete answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete answer passed unexpectedly: %+v", result)
	}
}

func TestVerifyMemoryRouterReferenceRequiresSourceFreshnessAndReferenceDecision(t *testing.T) {
	ctx := context.Background()
	t.Run("rejects wrong first-turn fixture", func(t *testing.T) {
		paths := scenarioPaths(t.TempDir())
		sc := scenario{ID: memoryRouterScenarioID, Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
		if err := seedScenario(ctx, paths, sc); err != nil {
			t.Fatalf("seed memory/router scenario: %v", err)
		}
		cfg := runclient.Config{DatabasePath: paths.DatabasePath}
		wrongBody := strings.Replace(memoryRouterSessionObservationBody(), "Positive feedback weight 0.8", "Positive feedback weight 0.1", 1)
		if err := createSeedDocument(ctx, cfg, memoryRouterSessionObservationPath, memoryRouterSessionObservationTitle, wrongBody); err != nil {
			t.Fatalf("create wrong session observation: %v", err)
		}
		result, err := verifyScenarioTurn(ctx, paths, sc, 1, memoryRouterSessionObservationPath, metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}})
		if err != nil {
			t.Fatalf("verify wrong first turn: %v", err)
		}
		if result.Passed {
			t.Fatalf("wrong first-turn fixture passed unexpectedly: %+v", result)
		}
	})

	paths := scenarioPaths(t.TempDir())
	sc := scenario{ID: memoryRouterScenarioID, Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
	if err := seedScenario(ctx, paths, sc); err != nil {
		t.Fatalf("seed memory/router scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, memoryRouterSessionObservationPath, memoryRouterSessionObservationTitle, memoryRouterSessionObservationBody()); err != nil {
		t.Fatalf("create session observation: %v", err)
	}
	turnOne, err := verifyScenarioTurn(ctx, paths, sc, 1, memoryRouterSessionObservationPath, metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify memory/router turn one: %v", err)
	}
	if !turnOne.Passed {
		t.Fatalf("memory/router turn one failed: %+v", turnOne)
	}

	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, notes/memory-router/routing-policy.md
---
# Memory Router Reference

## Summary
Temporal status: current canonical docs outrank stale session observations.
Session promotion path: durable canonical markdown with source refs.
Feedback weighting: advisory only.
Routing choice: existing AgentOps document and retrieval actions.
Decision: keep memory and autonomous routing as reference/deferred.

## Sources
- notes/memory-router/session-observation.md
- notes/memory-router/temporal-policy.md
- notes/memory-router/feedback-weighting.md
- notes/memory-router/routing-policy.md

## Freshness
Checked provenance for the session observation and synthesis projection freshness after filing the reference note.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, memoryRouterSynthesisPath, "Memory Router Reference", synthesisBody); err != nil {
		t.Fatalf("create memory/router synthesis: %v", err)
	}
	sessionDocID, _, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		t.Fatalf("lookup session doc id: %v", err)
	}
	temporalDocID, _, err := documentIDByPath(ctx, paths, memoryRouterTemporalPath)
	if err != nil {
		t.Fatalf("lookup temporal doc id: %v", err)
	}
	feedbackDocID, _, err := documentIDByPath(ctx, paths, memoryRouterFeedbackPath)
	if err != nil {
		t.Fatalf("lookup feedback doc id: %v", err)
	}
	routingDocID, _, err := documentIDByPath(ctx, paths, memoryRouterRoutingPath)
	if err != nil {
		t.Fatalf("lookup routing doc id: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{memoryRouterPrefix},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{sessionDocID, temporalDocID, feedbackDocID, routingDocID},
		ProvenanceEventsUsed:     true,
		ProjectionStatesUsed:     true,
		EventTypeCounts:          map[string]int{},
	}
	completeAnswer := "Temporal status is current for canonical docs and stale for unpromoted session observations. Session promotion happened through canonical markdown in synthesis/memory-router-reference.md with source refs. Feedback weighting is advisory, routing stays on existing AgentOps document and retrieval actions, and provenance plus projection freshness were checked. Decision: keep memory/router reference/deferred and do not promote remember/recall or autonomous routing."
	result, err := verifyScenarioTurn(ctx, paths, sc, 2, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify memory/router reference: %v", err)
	}
	if !result.Passed {
		t.Fatalf("memory/router reference failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics, *string){
		"missing list prefix": func(m *metrics, _ *string) { m.ListDocumentPathPrefixes = nil },
		"missing temporal get": func(m *metrics, _ *string) {
			m.GetDocumentDocIDs = []string{sessionDocID, feedbackDocID, routingDocID}
		},
		"missing provenance": func(m *metrics, _ *string) { m.ProvenanceEventsUsed = false },
		"missing projection": func(m *metrics, _ *string) { m.ProjectionStatesUsed = false },
		"missing session promotion": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Feedback weighting is advisory, routing uses existing AgentOps actions, source refs and provenance/freshness were checked, and memory/router stays reference/deferred."
		},
		"missing feedback": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, routing uses existing AgentOps actions, source refs and provenance/freshness were checked, and memory/router stays reference/deferred."
		},
		"missing routing": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, feedback weighting is advisory, source refs and provenance/freshness were checked, and this memory capability stays reference/deferred."
		},
		"promoted memory/router": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, feedback weighting is advisory, routing is clear, source refs and provenance/freshness were checked. Decision: promote memory/router."
		},
	} {
		t.Run(name, func(t *testing.T) {
			incompleteMetrics := completeMetrics
			answer := completeAnswer
			mutate(&incompleteMetrics, &answer)
			result, err := verifyScenarioTurn(ctx, paths, sc, 2, answer, incompleteMetrics)
			if err != nil {
				t.Fatalf("verify %s: %v", name, err)
			}
			if result.Passed {
				t.Fatalf("%s passed unexpectedly: %+v", name, result)
			}
		})
	}
}

func TestVerifyConfiguredLayoutRequiresUnambiguousValidAnswer(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: configuredLayoutScenarioID}); err != nil {
		t.Fatalf("seed configured layout scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:    1,
		InspectLayoutUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	invalidAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is invalid."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, invalidAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify invalid status answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("configured layout answer with invalid status passed: %+v", result)
	}

	negatedAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is not valid."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, negatedAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify negated valid status answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("configured layout answer with not valid status passed: %+v", result)
	}

	validAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is valid."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, validAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify valid status answer: %v", err)
	}
	if !result.Passed {
		t.Fatalf("configured layout answer with valid status failed: %+v", result)
	}
}

func TestGraphContainsNodeLabelsRequiresEveryCitedLabel(t *testing.T) {
	nodes := []runner.GraphNode{
		{Label: "AgentOps Wiki Index", Citations: []runner.Citation{{DocID: "doc_1"}}},
		{Label: "Runner Policy", Citations: []runner.Citation{{DocID: "doc_2"}}},
		{Label: "Knowledge Plane", Citations: []runner.Citation{{DocID: "doc_3"}}},
	}
	if graphContainsNodeLabels(nodes, []string{"AgentOps Wiki Index", "Runner Policy", "Knowledge Plane", "Runner Playbook"}) {
		t.Fatal("graph labels passed without the runner playbook node")
	}
	nodes = append(nodes, runner.GraphNode{Label: "Runner Playbook", Citations: []runner.Citation{{DocID: "doc_4"}}})
	if !graphContainsNodeLabels(nodes, []string{"AgentOps Wiki Index", "Runner Policy", "Knowledge Plane", "Runner Playbook"}) {
		t.Fatal("graph labels failed with every required cited node")
	}
}

func TestVerifySourceLinkedSynthesisRequiresSourcesFreshnessAndWorkflow(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	incomplete := `---
type: synthesis
status: active
freshness: fresh
---

# OpenClerk runner

## Sources

## Freshness
Checked search results.
`
	if err := createSeedDocument(ctx, cfg, "synthesis/openclerk-runner.md", "OpenClerk runner", incomplete); err != nil {
		t.Fatalf("create incomplete synthesis: %v", err)
	}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify incomplete synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis without source_refs passed: %+v", result)
	}
	yamlListPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, yamlListPaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed YAML-list source_refs scenario: %v", err)
	}
	yamlListCfg := runclient.Config{DatabasePath: yamlListPaths.DatabasePath}
	yamlListSourceRefs := `---
type: synthesis
status: active
freshness: fresh
source_refs:
  - sources/openclerk-runner.md
---

# OpenClerk runner

## Summary
The runner preserves source refs.

## Sources
- sources/openclerk-runner.md

## Freshness
Checked runner search results for sources/openclerk-runner.md.
`
	if err := createSeedDocument(ctx, yamlListCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", yamlListSourceRefs); err != nil {
		t.Fatalf("create YAML-list source_refs synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, yamlListPaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify YAML-list source_refs synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis with YAML-list source_refs passed: %+v", result)
	}
	missingFreshnessPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, missingFreshnessPaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed missing freshness scenario: %v", err)
	}
	missingFreshnessCfg := runclient.Config{DatabasePath: missingFreshnessPaths.DatabasePath}
	missingFreshness := `---
type: synthesis
status: active
source_refs: sources/openclerk-runner.md
---

# OpenClerk runner

## Sources
- sources/openclerk-runner.md
`
	if err := createSeedDocument(ctx, missingFreshnessCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", missingFreshness); err != nil {
		t.Fatalf("create missing freshness synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, missingFreshnessPaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify missing freshness synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis without freshness metadata passed: %+v", result)
	}
	completePaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, completePaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed complete scenario: %v", err)
	}
	completeCfg := runclient.Config{DatabasePath: completePaths.DatabasePath}
	complete := `---
type: synthesis
status: active
freshness: fresh
source_refs: sources/openclerk-runner.md
---

# OpenClerk runner

## Summary
The runner preserves source refs.

## Sources
- sources/openclerk-runner.md

## Freshness
Checked runner search results for sources/openclerk-runner.md.
`
	if err := createSeedDocument(ctx, completeCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", complete); err != nil {
		t.Fatalf("create complete synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, completePaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify complete synthesis: %v", err)
	}
	if !result.Passed {
		t.Fatalf("complete synthesis failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, completePaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify final answer path: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis final answer without path passed: %+v", result)
	}
}

func TestVerifyStaleSynthesisUpdateRequiresCurrentSourceAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "stale-synthesis-update"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md.", noTools)
	if err != nil {
		t.Fatalf("verify stale before update: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale synthesis passed before update: %+v", result)
	}
	replacement := "Current guidance: routine agents must use openclerk JSON runner.\n\nCurrent source: sources/runner-current-runner.md\n\nSupersedes: sources/runner-old-workaround.md\n\nThis stale claim is superseded by current guidance."
	replaceSeedSection(t, ctx, paths, "synthesis/runner-routing.md", "Summary", replacement)
	replaceSeedSection(t, ctx, paths, "synthesis/runner-routing.md", "Freshness", "Checked current source: sources/runner-current-runner.md\n\nChecked previous source: sources/runner-old-workaround.md")
	workflowMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		GetDocumentUsed:   true,
		EventTypeCounts:   map[string]int{},
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md with current guidance.", workflowMetrics)
	if err != nil {
		t.Fatalf("verify stale after update: %v", err)
	}
	if !result.Passed {
		t.Fatalf("updated stale synthesis failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/runner-routing-current.md", "OpenClerk runner Routing Current", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md with current guidance.", workflowMetrics)
	if err != nil {
		t.Fatalf("verify stale duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("duplicate synthesis passed: %+v", result)
	}
}

func TestVerifySourceSensitiveAuditRepairRequiresProvenanceFreshnessAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}); err != nil {
		t.Fatalf("seed source audit repair scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, "Updated "+sourceAuditSynthesisPath+".", noTools)
	if err != nil {
		t.Fatalf("verify source audit before repair: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit repair passed before update: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, sourceAuditSynthesisPath, "Summary", "Current audit guidance: use the installed openclerk JSON runner.\n\nCurrent source: "+sourceAuditCurrentSourcePath+"\n\nSuperseded source: "+sourceAuditOldSourcePath)
	replaceSeedSection(t, ctx, paths, sourceAuditSynthesisPath, "Freshness", "Checked provenance events and synthesis projection freshness after the current source update.")
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    5,
		ToolCalls:            5,
	}
	finalAnswer := "Updated " + sourceAuditSynthesisPath + " from " + sourceAuditCurrentSourcePath + "; projection freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit repair: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source audit repair failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/audit-runner-routing-v2.md", "Audit Runner Routing V2", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate audit synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit repair passed with duplicate synthesis: %+v", result)
	}
}

func TestVerifySourceSensitiveConflictRequiresUnresolvedExplanationAndNoSynthesis(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}); err != nil {
		t.Fatalf("seed source audit conflict scenario: %v", err)
	}
	alphaID, alphaFound, err := documentIDByPath(ctx, paths, sourceAuditConflictAlphaPath)
	if err != nil {
		t.Fatalf("lookup alpha id: %v", err)
	}
	bravoID, bravoFound, err := documentIDByPath(ctx, paths, sourceAuditConflictBravoPath)
	if err != nil {
		t.Fatalf("lookup bravo id: %v", err)
	}
	if !alphaFound || !bravoFound {
		t.Fatalf("missing conflict source ids: alpha=%v bravo=%v", alphaFound, bravoFound)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	finalAnswer := sourceAuditConflictAlphaPath + " says seven days; " + sourceAuditConflictBravoPath + " says thirty days. Both are current sources. This conflict is unresolved because there is no supersession metadata, so I cannot choose a winner without source authority."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, noTools)
	if err != nil {
		t.Fatalf("verify source audit conflict no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without runner workflow: %+v", result)
	}
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    3,
		ToolCalls:            3,
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict missing provenance refs: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without both provenance refs: %+v", result)
	}
	workflowMetrics.ProvenanceEventRefIDs = []string{alphaID, bravoID}
	answerWithoutCurrentSources := sourceAuditConflictAlphaPath + " says seven days; " + sourceAuditConflictBravoPath + " says thirty days. This conflict is unresolved because there is no supersession metadata, so I cannot choose a winner without source authority."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, answerWithoutCurrentSources, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict missing current-source wording: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without current-source wording: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source audit conflict failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/audit-conflict.md", "Audit Conflict", "# Audit Conflict\n"); err != nil {
		t.Fatalf("create forbidden conflict synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict with synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed after creating synthesis: %+v", result)
	}
}

func TestVerifySynthesisCandidatePressureRequiresCandidateWorkflowAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}); err != nil {
		t.Fatalf("seed candidate pressure scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, "Updated "+synthesisCandidatePath+".", noTools)
	if err != nil {
		t.Fatalf("verify candidate no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("candidate pressure passed before repair: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, synthesisCandidatePath, "Summary", "Current compiler decision: existing document and retrieval actions are sufficient for synthesis compiler pressure repairs.\n\nCurrent source: "+synthesisCandidateCurrentSrc+"\n\nSuperseded source: "+synthesisCandidateOldSrc)
	replaceSeedSection(t, ctx, paths, synthesisCandidatePath, "Freshness", "Checked synthesis projection freshness after searching sources and listing candidates.")
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    4,
		ToolCalls:            4,
	}
	finalAnswer := "Updated " + synthesisCandidatePath + " from " + synthesisCandidateCurrentSrc + "; projection freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify candidate repair: %v", err)
	}
	if !result.Passed {
		t.Fatalf("candidate pressure repair failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/compiler-routing-copy.md", "Compiler Routing Copy", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate candidate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisCandidatePressureScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify candidate duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("candidate pressure passed with duplicate synthesis: %+v", result)
	}
}

func TestVerifySynthesisSourceSetPressureRequiresAllSources(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}); err != nil {
		t.Fatalf("seed source set pressure scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/source-set-alpha.md, sources/source-set-beta.md, sources/source-set-gamma.md
---
# Compiler Source Set

## Summary
Alpha, beta, and gamma source refs show the synthesis compiler pressure workflow can preserve freshness.

## Sources
- sources/source-set-alpha.md
- sources/source-set-beta.md
- sources/source-set-gamma.md

## Freshness
Checked runner search results and synthesis candidate listing for all source refs.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisSourceSetPath, "Compiler Source Set", body); err != nil {
		t.Fatalf("create source set synthesis: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	finalAnswer := "Created " + synthesisSourceSetPath + " with alpha, beta, and gamma sources."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify source set pressure: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source set pressure failed: %+v", result)
	}

	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}, 1, finalAnswer, metrics{AssistantCalls: 1, SearchUsed: true, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify missing list metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("source set pressure passed without candidate listing metric: %+v", result)
	}
}

func TestVerifyMTSynthesisDriftPressureRequiresSourceUpdateAndRepair(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	sc := requireScenarioByID(t, mtSynthesisDriftPressureScenarioID)
	if err := seedScenario(ctx, paths, sc); err != nil {
		t.Fatalf("seed drift pressure scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	initialBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/drift-current.md, sources/drift-old.md
---
# Drift Runner

## Summary
Initial drift synthesis says the decision is still under review.

## Sources
- sources/drift-current.md
- sources/drift-old.md

## Freshness
Checked initial source refs.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, mtDriftSynthesisPath, "Drift Runner", initialBody); err != nil {
		t.Fatalf("create initial drift synthesis: %v", err)
	}
	turnOneMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	result, err := verifyScenarioTurn(ctx, paths, sc, 1, "Created "+mtDriftSynthesisPath+".", turnOneMetrics)
	if err != nil {
		t.Fatalf("verify drift turn one: %v", err)
	}
	if !result.Passed {
		t.Fatalf("drift turn one failed: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, mtDriftCurrentPath, "Summary", "Current drift decision says existing document and retrieval actions should stay the v1 synthesis path.")
	replaceSeedSection(t, ctx, paths, mtDriftSynthesisPath, "Summary", "Current drift decision: keep existing document and retrieval actions.\n\nCurrent source: "+mtDriftCurrentPath+"\n\nSuperseded source: "+mtDriftOldSourcePath)
	replaceSeedSection(t, ctx, paths, mtDriftSynthesisPath, "Freshness", "Checked synthesis projection freshness after the current source update.")
	turnTwoMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		EventTypeCounts:      map[string]int{},
	}
	finalAnswer := "Updated " + mtDriftSynthesisPath + " from " + mtDriftCurrentPath + "; final freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, sc, 2, finalAnswer, turnTwoMetrics)
	if err != nil {
		t.Fatalf("verify drift turn two: %v", err)
	}
	if !result.Passed {
		t.Fatalf("drift turn two failed: %+v", result)
	}
}

func TestVerifyPromotedRecordVsDocsRequiresComparisonAnswer(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "promoted-record-vs-docs"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	withRetrievalWork := metrics{AssistantCalls: 1, ToolCalls: 2, CommandExecutions: 2, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "Services lookup says JSON runner; plain docs search agrees.", noTools)
	if err != nil {
		t.Fatalf("verify records vs docs no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-tool records vs docs passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "Services lookup says JSON runner; plain docs search agrees.", withRetrievalWork)
	if err != nil {
		t.Fatalf("verify records vs docs: %v", err)
	}
	if !result.Passed {
		t.Fatalf("records vs docs failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "JSON runner.", withRetrievalWork)
	if err != nil {
		t.Fatalf("verify incomplete records vs docs answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete comparison passed: %+v", result)
	}
}

func TestVerifyDecisionRecordVsDocsRequiresTypedLookup(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}); err != nil {
		t.Fatalf("seed decision scenario: %v", err)
	}
	noTypedLookup := metrics{AssistantCalls: 1, SearchUsed: true, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, SearchUsed: true, DecisionsLookupUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "Plain docs search agrees, but decisions lookup filters status and scope for the accepted AgentOps JSON runner decision."
	completeAnswer := noCitationAnswer + " The decision citation path is docs/architecture/runner-current-decision.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, completeAnswer, noTypedLookup)
	if err != nil {
		t.Fatalf("verify decision no typed lookup: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-typed decision comparison passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify decision no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation decision comparison passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify decision comparison: %v", err)
	}
	if !result.Passed {
		t.Fatalf("decision comparison failed: %+v", result)
	}
}

func TestVerifyDecisionSupersessionFreshnessRequiresProjectionAndProvenance(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionSupersessionScenarioID}); err != nil {
		t.Fatalf("seed supersession scenario: %v", err)
	}
	noProjection := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old", "adr-runner-current"}, EventTypeCounts: map[string]int{}}
	incompleteDecisionRecord := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old", "adr-runner-current"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "adr-runner-old is superseded and stale; adr-runner-current supersedes it and is fresh, with provenance and projection evidence."
	completeAnswer := noCitationAnswer + " Citation paths: docs/architecture/runner-old-decision.md and records/decisions/runner-current-decision.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, noProjection)
	if err != nil {
		t.Fatalf("verify supersession no projection: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-projection supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, incompleteDecisionRecord)
	if err != nil {
		t.Fatalf("verify supersession incomplete decision record ids: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete decision record ids supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify supersession no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify supersession: %v", err)
	}
	if !result.Passed {
		t.Fatalf("supersession failed: %+v", result)
	}
}

func TestVerifyDecisionRealADRMigrationRequiresDecisionProjectionEvidence(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}); err != nil {
		t.Fatalf("seed real adr scenario: %v", err)
	}
	noProjection := metrics{AssistantCalls: 1, DecisionsLookupUsed: true, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-agentops-only-knowledge-plane"}, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, DecisionsLookupUsed: true, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-agentops-only-knowledge-plane"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "Canonical markdown ADRs remain authoritative; decisions_lookup and decision_record return derived decision records with fresh projection and provenance evidence."
	completeAnswer := noCitationAnswer + " Citation paths: docs/architecture/eval-backed-knowledge-plane-adr.md and docs/architecture/knowledge-configuration-v1-adr.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, completeAnswer, noProjection)
	if err != nil {
		t.Fatalf("verify real adr no projection: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-projection real adr migration passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify real adr no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation real adr migration passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify real adr migration: %v", err)
	}
	if !result.Passed {
		t.Fatalf("real adr migration failed: %+v", result)
	}
}

func TestDuplicatePathRejectRequiresAnswerFailure(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "duplicate-path-reject"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	sc := scenario{ID: "duplicate-path-reject"}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, sc, 1, "Done.", noTools)
	if err != nil {
		t.Fatalf("verify duplicate no-op: %v", err)
	}
	if result.Passed {
		t.Fatalf("non-rejection answer passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, sc, 1, "notes/projects/duplicate.md already exists, so I did not overwrite it.", noTools)
	if err != nil {
		t.Fatalf("verify duplicate rejection: %v", err)
	}
	if !result.Passed {
		t.Fatalf("duplicate rejection failed: %+v", result)
	}
}

func TestProductionGateFailsPartialScenarioRun(t *testing.T) {
	results := []jobResult{
		comparisonResult(productionVariant, "create-note", true, 1, 1, 1, 80),
	}
	summary := buildProductionGateSummary(results)
	if summary == nil {
		t.Fatal("missing summary")
	}
	if summary.PassesGate {
		t.Fatalf("partial production run passed gate: %+v", summary)
	}
	if summary.Recommendation != "fix_production_agentops_before_release" {
		t.Fatalf("recommendation = %q", summary.Recommendation)
	}
	criteria := map[string]productionGateCriterion{}
	for _, criterion := range summary.Criteria {
		criteria[criterion.Name] = criterion
	}
	if criteria["production_passes_all_scenarios"].Passed {
		t.Fatalf("partial run passed scenario criterion: %+v", summary.Criteria)
	}
	if !strings.Contains(criteria["production_passes_all_scenarios"].Details, "missing:") {
		t.Fatalf("scenario criterion did not list missing scenarios: %+v", criteria["production_passes_all_scenarios"])
	}
	if criteria["validation_scenarios_are_final_answer_only"].Passed {
		t.Fatalf("partial run passed validation criterion: %+v", summary.Criteria)
	}
	if !strings.Contains(criteria["validation_scenarios_are_final_answer_only"].Details, "not evaluated") {
		t.Fatalf("validation criterion did not explain partial run: %+v", criteria["validation_scenarios_are_final_answer_only"])
	}
}

func replaceSeedSection(t *testing.T, ctx context.Context, paths evalPaths, docPath string, heading string, content string) {
	t.Helper()
	docID, found, err := documentIDByPath(ctx, paths, docPath)
	if err != nil {
		t.Fatalf("find %s: %v", docPath, err)
	}
	if !found {
		t.Fatalf("missing %s", docPath)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action:  runner.DocumentTaskActionReplaceSection,
		DocID:   docID,
		Heading: heading,
		Content: content,
	})
	if err != nil {
		t.Fatalf("replace %s section %s: %v", docPath, heading, err)
	}
	if result.Rejected {
		t.Fatalf("replace %s rejected: %s", docPath, result.RejectionReason)
	}
}

func requireRAGMetadataTopHit(t *testing.T, ctx context.Context, paths evalPaths) runner.SearchHit {
	t.Helper()
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		t.Fatalf("metadata search: %v", err)
	}
	top, ok := topSearchHit(result)
	if !ok {
		t.Fatalf("metadata search returned no hits: %+v", result)
	}
	if searchHitPath(top) != ragCurrentPolicyPath {
		t.Fatalf("metadata search top path = %q, want %q; result=%+v", searchHitPath(top), ragCurrentPolicyPath, result.Search)
	}
	return top
}

func comparisonResult(variant string, scenario string, passed bool, tools int, commands int, assistant int, nonCached int) jobResult {
	input := nonCached
	cached := 0
	output := 10
	return jobResult{
		Variant:  variant,
		Scenario: scenario,
		Passed:   passed,
		Metrics: metrics{
			ToolCalls:            tools,
			CommandExecutions:    commands,
			AssistantCalls:       assistant,
			UsageExposed:         true,
			InputTokens:          &input,
			CachedInputTokens:    &cached,
			NonCachedInputTokens: &nonCached,
			OutputTokens:         &output,
			EventTypeCounts:      map[string]int{},
		},
	}
}

func requireScenarioByID(t *testing.T, id string) scenario {
	t.Helper()
	for _, sc := range allScenarios() {
		if sc.ID == id {
			return sc
		}
	}
	t.Fatalf("missing scenario %q", id)
	return scenario{}
}

func containsValue(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}
