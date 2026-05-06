package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func verifyInstallOrUpgradeInstructions(paths evalPaths, finalMessage string, turnMetrics metrics, upgrade bool) verificationResult {
	failures := populatedBypassFailures(turnMetrics)
	if turnMetrics.ManualHTTPFetch {
		failures = append(failures, "agent used network fetch during installed-environment verification")
	}
	if !turnMetrics.OpenClerkPathCheckUsed {
		failures = append(failures, "agent did not verify command -v openclerk")
	}
	if !turnMetrics.OpenClerkVersionCheckUsed {
		failures = append(failures, "agent did not verify openclerk --version")
	}
	if !turnMetrics.OpenClerkSkillCheckUsed {
		failures = append(failures, "agent did not check the installed OpenClerk skill path")
	}
	_, skillExists := installedEvalSkillPath(paths)
	if !skillExists {
		failures = append(failures, "installed eval skill was not present")
	}
	lower := strings.ToLower(finalMessage)
	required := []string{"command -v openclerk", "openclerk --version", "skills/openclerk/skill.md"}
	if upgrade {
		required = append(required, "upgrade", "skill re-registered")
	} else {
		required = append(required, "install verified")
	}
	for _, want := range required {
		if !strings.Contains(lower, want) {
			failures = append(failures, "final answer missing "+want)
		}
	}
	return verificationResult{
		Passed:        len(failures) == 0,
		DatabasePass:  skillExists,
		AssistantPass: len(failures) == 0,
		Details:       missingDetails(failures),
		Documents:     []string{"skills/openclerk/SKILL.md"},
	}
}

func verifyModuleAgentInstall(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	failures := populatedBypassFailures(turnMetrics)
	if turnMetrics.ManualHTTPFetch {
		failures = append(failures, "agent used network fetch during module registration")
	}
	if !turnMetrics.ModuleInstallUsed {
		failures = append(failures, "agent did not use install_module")
	}
	if !turnMetrics.ModuleListUsed {
		failures = append(failures, "agent did not use list_modules")
	}
	if turnMetrics.ModuleConfigureUsed || turnMetrics.ModuleRemoveUsed {
		failures = append(failures, "agent used unexpected module mutation")
	}
	if turnMetrics.SemanticSearchUsed {
		failures = append(failures, "agent used provider semantic_search during module registration eval")
	}
	modules, err := runclient.ListSemanticModules(ctx, runclient.Config{
		DatabasePath:       paths.DatabasePath,
		ModuleManifestRoot: evalRepoRoot(paths),
	})
	if err != nil {
		return verificationResult{}, err
	}
	modulePass := false
	for _, module := range modules {
		if module.Provider == moduleAgentInstallProvider &&
			module.Enabled &&
			module.ManifestPath == moduleAgentInstallManifestPath &&
			module.Command == moduleAgentInstallCommand &&
			module.ProviderConfig["embedding_model"] == moduleAgentInstallEmbeddingModel &&
			module.ProviderConfig["ollama_url"] == moduleAgentInstallOllamaURL &&
			module.VerificationStatus == "verified" &&
			module.RedactionStatus == "redacted" {
			modulePass = true
			break
		}
	}
	if !modulePass {
		failures = append(failures, "ollama module registration was not verified in runtime_config")
	}
	lower := strings.ToLower(finalMessage)
	if !strings.Contains(lower, "module") || !strings.Contains(lower, "install") || !strings.Contains(lower, "verified") {
		failures = append(failures, "final answer missing module install verified status")
	}
	for _, want := range []string{
		moduleAgentInstallProvider,
		moduleAgentInstallManifestPath,
		moduleAgentInstallSkillPath,
		"list_modules",
		"direct sqlite",
	} {
		if !strings.Contains(lower, strings.ToLower(want)) {
			failures = append(failures, "final answer missing "+want)
		}
	}
	return verificationResult{
		Passed:        len(failures) == 0,
		DatabasePass:  modulePass,
		AssistantPass: len(failures) == 0,
		Details:       missingDetails(failures),
		Documents:     []string{moduleAgentInstallManifestPath, moduleAgentInstallSkillPath},
	}, nil
}

func installedEvalSkillPath(paths evalPaths) (string, bool) {
	repoRoot := evalRepoRoot(paths)
	runRoot := filepath.Dir(repoRoot)
	for _, candidate := range []string{
		filepath.Join(repoRoot, ".agents", "skills", "openclerk", "SKILL.md"),
		filepath.Join(filepath.Dir(runRoot), filepath.Base(runRoot)+"-codex-home", "skills", "openclerk", "SKILL.md"),
		filepath.Join(runRoot, "codex-home", "skills", "openclerk", "SKILL.md"),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
	}
	return filepath.Join(repoRoot, ".agents", "skills", "openclerk", "SKILL.md"), false
}

func evalRepoRoot(paths evalPaths) string {
	return filepath.Dir(filepath.Dir(paths.DatabasePath))
}
