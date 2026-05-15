package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type evalManifestFile struct {
	SchemaVersion       string            `json:"schema_version"`
	Authority           manifestAuthority `json:"authority"`
	RepresentativeLanes []manifestLane    `json:"representative_lanes"`
}

type manifestAuthority struct {
	ReleaseAuthority string `json:"release_authority"`
	PromptfooStatus  string `json:"promptfoo_status"`
}

type manifestLane struct {
	LaneID             string             `json:"lane_id"`
	Objective          string             `json:"objective"`
	ScenarioClass      string             `json:"scenario_class"`
	ReleaseBlocking    bool               `json:"release_blocking"`
	RelatedOCKPCommand string             `json:"related_ockp_command"`
	RelatedReport      string             `json:"related_report"`
	RelatedReports     []string           `json:"related_reports,omitempty"`
	SourceEvidence     []manifestEvidence `json:"source_evidence"`
	Cases              []manifestCase     `json:"cases"`
}

type manifestCase struct {
	ScenarioID          string              `json:"scenario_id"`
	Objective           string              `json:"objective"`
	ScenarioClass       string              `json:"scenario_class"`
	SourceEvidence      []manifestEvidence  `json:"source_evidence"`
	DeterministicChecks []string            `json:"deterministic_checks"`
	Dimensions          map[string][]string `json:"dimensions"`
	ReleaseBlocking     bool                `json:"release_blocking"`
	RelatedOCKPCommand  string              `json:"related_ockp_command"`
	RelatedReport       string              `json:"related_report"`
}

type manifestEvidence struct {
	Path    string `json:"path"`
	Section string `json:"section"`
}

type calibrationFile struct {
	SchemaVersion string               `json:"schema_version"`
	Authority     calibrationAuthority `json:"authority"`
	Examples      []calibrationExample `json:"examples"`
}

type calibrationAuthority struct {
	ReleaseAuthority  string  `json:"release_authority"`
	PromptfooStatus   string  `json:"promptfoo_status"`
	LLMGraderUsed     bool    `json:"llm_grader_used"`
	LLMGraderProvider *string `json:"llm_grader_provider"`
}

type calibrationExample struct {
	ID                 string             `json:"id"`
	ScenarioID         string             `json:"scenario_id"`
	SourceEvidence     []manifestEvidence `json:"source_evidence"`
	InputSummary       string             `json:"input_summary"`
	ObservedSummary    string             `json:"observed_behavior_summary"`
	HumanExpectedLabel calibrationLabel   `json:"human_expected_label"`
}

type calibrationLabel struct {
	OverallLabel          string `json:"overall_label"`
	SafetyPass            string `json:"safety_pass"`
	CapabilityPass        string `json:"capability_pass"`
	UXQuality             string `json:"ux_quality"`
	FailureClassification string `json:"failure_classification"`
}

func TestOCKPEvalManifestCoversRepresentativeCases(t *testing.T) {
	var manifest evalManifestFile
	readRepoJSON(t, filepath.Join("docs", "evals", "ockp-eval-manifest.json"), &manifest)

	if manifest.SchemaVersion != "openclerk.ockp.eval_manifest.v1" {
		t.Fatalf("schema_version = %q", manifest.SchemaVersion)
	}
	if manifest.Authority.ReleaseAuthority != "scripts/agent-eval/ockp" {
		t.Fatalf("release authority = %q", manifest.Authority.ReleaseAuthority)
	}
	if manifest.Authority.PromptfooStatus != "optional_non_authoritative_smoke" {
		t.Fatalf("promptfoo status = %q", manifest.Authority.PromptfooStatus)
	}

	scenarios := map[string]bool{}
	for _, id := range scenarioIDs() {
		scenarios[id] = true
	}

	caseCount := 0
	seenCases := map[string]bool{}
	for _, lane := range manifest.RepresentativeLanes {
		if strings.TrimSpace(lane.LaneID) == "" || strings.TrimSpace(lane.Objective) == "" || strings.TrimSpace(lane.ScenarioClass) == "" {
			t.Fatalf("lane has missing identity fields: %+v", lane)
		}
		requireOCKPCommand(t, lane.RelatedOCKPCommand)
		requireReportReference(t, lane.RelatedReport)
		for _, report := range lane.RelatedReports {
			requireReportReference(t, report)
		}
		requireEvidencePaths(t, lane.SourceEvidence)
		if len(lane.Cases) == 0 {
			t.Fatalf("lane %s has no cases", lane.LaneID)
		}
		for _, c := range lane.Cases {
			caseCount++
			if seenCases[c.ScenarioID] {
				t.Fatalf("duplicate manifest case %q", c.ScenarioID)
			}
			seenCases[c.ScenarioID] = true
			if !scenarios[c.ScenarioID] {
				t.Fatalf("manifest case %q is not an OCKP scenario id", c.ScenarioID)
			}
			if strings.TrimSpace(c.Objective) == "" || strings.TrimSpace(c.ScenarioClass) == "" {
				t.Fatalf("manifest case %q missing objective or class", c.ScenarioID)
			}
			if len(c.DeterministicChecks) == 0 {
				t.Fatalf("manifest case %q has no deterministic checks", c.ScenarioID)
			}
			for _, key := range []string{"safety", "capability", "ux"} {
				if len(c.Dimensions[key]) == 0 {
					t.Fatalf("manifest case %q missing %s dimension", c.ScenarioID, key)
				}
			}
			if c.ReleaseBlocking != isReleaseBlockingScenario(c.ScenarioID) {
				t.Fatalf("manifest case %q release_blocking = %t, want %t", c.ScenarioID, c.ReleaseBlocking, isReleaseBlockingScenario(c.ScenarioID))
			}
			requireOCKPCommand(t, c.RelatedOCKPCommand)
			requireReportReference(t, c.RelatedReport)
			requireEvidencePaths(t, c.SourceEvidence)
		}
	}
	if caseCount < 5 || caseCount > 10 {
		t.Fatalf("manifest has %d cases, want a representative 5-10 case layer", caseCount)
	}
}

func TestOCKPGraderCalibrationLabelsCoverRequiredOutcomes(t *testing.T) {
	var calibration calibrationFile
	readRepoJSON(t, filepath.Join("docs", "evals", "calibration", "ockp-grader-calibration.json"), &calibration)

	if calibration.SchemaVersion != "openclerk.ockp.grader_calibration.v1" {
		t.Fatalf("schema_version = %q", calibration.SchemaVersion)
	}
	if calibration.Authority.ReleaseAuthority != "scripts/agent-eval/ockp" {
		t.Fatalf("release authority = %q", calibration.Authority.ReleaseAuthority)
	}
	if calibration.Authority.PromptfooStatus != "optional_non_authoritative_smoke" {
		t.Fatalf("promptfoo status = %q", calibration.Authority.PromptfooStatus)
	}
	if calibration.Authority.LLMGraderUsed {
		if calibration.Authority.LLMGraderProvider == nil || strings.TrimSpace(*calibration.Authority.LLMGraderProvider) == "" {
			t.Fatal("LLM grader is marked used without a pinned provider")
		}
	}

	scenarios := map[string]bool{}
	for _, id := range scenarioIDs() {
		scenarios[id] = true
	}
	requiredLabels := map[string]bool{
		"pass":                         false,
		"safety_fail":                  false,
		"capability_pass_ux_debt":      false,
		"unsupported_bypass_rejection": false,
	}
	for _, example := range calibration.Examples {
		if strings.TrimSpace(example.ID) == "" {
			t.Fatal("calibration example missing id")
		}
		if !scenarios[example.ScenarioID] {
			t.Fatalf("calibration example %q references unknown scenario %q", example.ID, example.ScenarioID)
		}
		if strings.TrimSpace(example.InputSummary) == "" || strings.TrimSpace(example.ObservedSummary) == "" {
			t.Fatalf("calibration example %q missing summaries", example.ID)
		}
		requireEvidencePaths(t, example.SourceEvidence)
		if _, ok := requiredLabels[example.HumanExpectedLabel.OverallLabel]; ok {
			requiredLabels[example.HumanExpectedLabel.OverallLabel] = true
		}
		if strings.TrimSpace(example.HumanExpectedLabel.SafetyPass) == "" ||
			strings.TrimSpace(example.HumanExpectedLabel.CapabilityPass) == "" ||
			strings.TrimSpace(example.HumanExpectedLabel.UXQuality) == "" ||
			strings.TrimSpace(example.HumanExpectedLabel.FailureClassification) == "" {
			t.Fatalf("calibration example %q has incomplete label: %+v", example.ID, example.HumanExpectedLabel)
		}
	}
	for label, covered := range requiredLabels {
		if !covered {
			t.Fatalf("calibration labels missing required outcome %q", label)
		}
	}
}

func readRepoJSON(t *testing.T, repoRelativePath string, target any) {
	t.Helper()
	requireRepoReference(t, repoRelativePath)
	content, err := os.ReadFile(repoPath(repoRelativePath))
	if err != nil {
		t.Fatalf("read %s: %v", repoRelativePath, err)
	}
	if strings.Contains(string(content), "/Users/") {
		t.Fatalf("%s contains a machine-absolute path", repoRelativePath)
	}
	if err := json.Unmarshal(content, target); err != nil {
		t.Fatalf("decode %s: %v", repoRelativePath, err)
	}
}

func requireOCKPCommand(t *testing.T, command string) {
	t.Helper()
	if !strings.HasPrefix(command, "mise exec -- go run ./scripts/agent-eval/ockp ") {
		t.Fatalf("OCKP command %q does not use repo-pinned go through mise", command)
	}
	if strings.Contains(command, "/Users/") {
		t.Fatalf("OCKP command %q contains a machine-absolute path", command)
	}
}

func requireEvidencePaths(t *testing.T, evidence []manifestEvidence) {
	t.Helper()
	if len(evidence) == 0 {
		t.Fatal("missing source evidence")
	}
	for _, item := range evidence {
		requireRepoReference(t, item.Path)
		if strings.TrimSpace(item.Section) == "" {
			t.Fatalf("source evidence %q has no section", item.Path)
		}
		requireMarkdownSection(t, item.Path, item.Section)
	}
}

func requireReportReference(t *testing.T, path string) {
	t.Helper()
	requireRepoReference(t, path)
	info, err := os.Stat(repoPath(path))
	if err != nil {
		t.Fatalf("report reference %q does not exist: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("report reference %q points at a directory", path)
	}
}

func requireRepoReference(t *testing.T, path string) {
	t.Helper()
	if strings.TrimSpace(path) == "" {
		t.Fatal("empty repo reference")
	}
	if filepath.IsAbs(path) || strings.Contains(path, "/Users/") {
		t.Fatalf("path %q is not repo-relative", path)
	}
	if strings.HasPrefix(path, "..") {
		t.Fatalf("path %q escapes repo", path)
	}
	if strings.HasPrefix(path, "docs/") || strings.HasPrefix(path, "skills/") {
		if _, err := os.Stat(repoPath(path)); err != nil {
			t.Fatalf("repo reference %q does not exist: %v", path, err)
		}
	}
}

func requireMarkdownSection(t *testing.T, path string, section string) {
	t.Helper()
	if !strings.HasSuffix(path, ".md") {
		return
	}
	content, err := os.ReadFile(repoPath(path))
	if err != nil {
		t.Fatalf("read source evidence %q: %v", path, err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
		if heading == section {
			return
		}
	}
	t.Fatalf("source evidence %q references missing markdown section %q", path, section)
}

func repoPath(repoRelativePath string) string {
	return filepath.Join("..", "..", "..", filepath.FromSlash(repoRelativePath))
}
