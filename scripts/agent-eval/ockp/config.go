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

func parseRunConfig(args []string, stderr io.Writer) (runConfig, error) {
	fs := flag.NewFlagSet("ockp run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := runConfig{CacheMode: cacheModeShared}
	fs.IntVar(&config.Parallel, "parallel", defaultParallel, "number of independent eval jobs to run concurrently")
	fs.StringVar(&config.Variant, "variant", "", "comma-separated variant ids")
	fs.StringVar(&config.Scenario, "scenario", "", "comma-separated scenario ids")
	fs.StringVar(&config.RunRoot, "run-root", "", "directory for isolated run artifacts")
	fs.StringVar(&config.ReportDir, "report-dir", filepath.Join("docs", "evals", "results"), "directory for reduced reports")
	fs.StringVar(&config.ReportName, "report-name", "ockp-latest", "base filename for reduced reports, without extension")
	fs.StringVar(&config.CodexBin, "codex-bin", "codex", "codex executable")
	fs.StringVar(&config.RepoRoot, "repo-root", ".", "repository root to copy for each job")
	fs.StringVar(&config.CacheMode, "cache-mode", config.CacheMode, "Go cache mode: shared or isolated")
	if err := fs.Parse(args); err != nil {
		return runConfig{}, err
	}
	if fs.NArg() != 0 {
		return runConfig{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if config.Parallel < 1 {
		return runConfig{}, errors.New("--parallel must be at least 1")
	}
	if config.CacheMode != cacheModeShared && config.CacheMode != cacheModeIsolated {
		return runConfig{}, fmt.Errorf("--cache-mode must be %q or %q", cacheModeShared, cacheModeIsolated)
	}
	if config.RunRoot == "" {
		config.RunRoot = filepath.Join(os.TempDir(), fmt.Sprintf("openclerk-ockp-%d", time.Now().UnixNano()))
	}
	if strings.TrimSpace(config.ReportName) == "" {
		return runConfig{}, errors.New("--report-name must not be empty")
	}
	return config, nil
}
func selectedVariants(config runConfig) []string {
	if strings.TrimSpace(config.Variant) != "" {
		return splitCSV(config.Variant)
	}
	return []string{productionVariant}
}
func selectedScenarios(config runConfig) []scenario {
	scenarios := allScenarios()
	if strings.TrimSpace(config.Scenario) == "" {
		filtered := make([]scenario, 0, len(scenarios))
		for _, scenario := range scenarios {
			if isReleaseBlockingScenario(scenario.ID) {
				filtered = append(filtered, scenario)
			}
		}
		return filtered
	}
	wanted := map[string]struct{}{}
	for _, id := range splitCSV(config.Scenario) {
		wanted[id] = struct{}{}
	}
	filtered := make([]scenario, 0, len(wanted))
	for _, scenario := range scenarios {
		if _, ok := wanted[scenario.ID]; ok {
			filtered = append(filtered, scenario)
		}
	}
	return filtered
}
func selectedScenarioIDs(config runConfig) []string {
	scenarios := selectedScenarios(config)
	ids := make([]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		ids = append(ids, scenario.ID)
	}
	return ids
}
func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
func containsArgPair(args []string, key string, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}
func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}
func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}
