package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestRunnerVersion(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{"--version"}, {"version"}} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := run(args, strings.NewReader(""), &stdout, &stderr)
		if code != 0 {
			t.Fatalf("run %v exit = %d stderr=%s", args, code, stderr.String())
		}
		if got := strings.TrimSpace(stdout.String()); !strings.HasPrefix(got, "openclerk ") {
			t.Fatalf("version output = %q, want openclerk prefix", got)
		}
	}
}

func TestResolvedVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		linkerVersion string
		info          *debug.BuildInfo
		ok            bool
		want          string
	}{
		{
			name:          "linker version wins",
			linkerVersion: "v0.1.0",
			info:          &debug.BuildInfo{Main: debug.Module{Version: "v0.0.9"}},
			ok:            true,
			want:          "v0.1.0",
		},
		{
			name: "module version",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}},
			ok:   true,
			want: "v0.1.0",
		},
		{
			name: "development fallback",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: "dev",
		},
		{
			name: "missing build info fallback",
			want: "dev",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resolvedVersion(tt.linkerVersion, tt.info, tt.ok); got != tt.want {
				t.Fatalf("resolvedVersion = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunnerDocumentAndRetrievalJSONRoundTrip(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	createRequest := `{"action":"create_document","document":{"path":"notes/runner.md","title":"Runner","body":"# Runner\n\n## Summary\nOpenClerk runner note.\n"}}`
	var createResult runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, createRequest, &createResult)
	if code != 0 {
		t.Fatalf("create exit = %d stderr=%s", code, stderr)
	}
	if createResult.Document == nil || createResult.Document.DocID == "" {
		t.Fatalf("create result = %+v", createResult)
	}

	searchRequest := `{"action":"search","search":{"text":"runner","limit":10}}`
	var searchResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, searchRequest, &searchResult)
	if code != 0 {
		t.Fatalf("search exit = %d stderr=%s", code, stderr)
	}
	if searchResult.Search == nil || len(searchResult.Search.Hits) == 0 {
		t.Fatalf("search result = %+v", searchResult)
	}

	serviceRequest := `{"action":"create_document","document":{"path":"records/services/openclerk-runner.md","title":"OpenClerk runner","body":"---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service.\n"}}`
	var serviceCreate runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, serviceRequest, &serviceCreate)
	if code != 0 {
		t.Fatalf("create service exit = %d stderr=%s", code, stderr)
	}

	servicesRequest := `{"action":"services_lookup","services":{"text":"OpenClerk runner","interface":"JSON runner","limit":10}}`
	var servicesResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, servicesRequest, &servicesResult)
	if code != 0 {
		t.Fatalf("services exit = %d stderr=%s", code, stderr)
	}
	if servicesResult.Services == nil || len(servicesResult.Services.Services) != 1 {
		t.Fatalf("services result = %+v", servicesResult)
	}

	serviceDetailRequest := `{"action":"service_record","service_id":"openclerk-runner"}`
	var serviceDetail runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, serviceDetailRequest, &serviceDetail)
	if code != 0 {
		t.Fatalf("service detail exit = %d stderr=%s", code, stderr)
	}
	if serviceDetail.Service == nil || serviceDetail.Service.Interface != "JSON runner" {
		t.Fatalf("service detail = %+v", serviceDetail)
	}

	decisionRequest := `{"action":"create_document","document":{"path":"docs/architecture/runner-decision.md","title":"Runner decision","body":"---\ndecision_id: adr-runner\ndecision_title: Use JSON runner\ndecision_status: accepted\ndecision_scope: agentops\ndecision_owner: platform\ndecision_date: 2026-04-22\n---\n# Runner decision\n\n## Summary\nUse the JSON runner.\n"}}`
	var decisionCreate runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, decisionRequest, &decisionCreate)
	if code != 0 {
		t.Fatalf("create decision exit = %d stderr=%s", code, stderr)
	}

	decisionsRequest := `{"action":"decisions_lookup","decisions":{"text":"JSON runner","status":"accepted","scope":"agentops","owner":"platform","limit":10}}`
	var decisionsResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, decisionsRequest, &decisionsResult)
	if code != 0 {
		t.Fatalf("decisions exit = %d stderr=%s", code, stderr)
	}
	if decisionsResult.Decisions == nil || len(decisionsResult.Decisions.Decisions) != 1 {
		t.Fatalf("decisions result = %+v", decisionsResult)
	}

	decisionDetailRequest := `{"action":"decision_record","decision_id":"adr-runner"}`
	var decisionDetail runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--db", dbPath}, decisionDetailRequest, &decisionDetail)
	if code != 0 {
		t.Fatalf("decision detail exit = %d stderr=%s", code, stderr)
	}
	if decisionDetail.Decision == nil || decisionDetail.Decision.Status != "accepted" {
		t.Fatalf("decision detail = %+v", decisionDetail)
	}

	layoutRequest := `{"action":"inspect_layout"}`
	var layoutResult runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, layoutRequest, &layoutResult)
	if code != 0 {
		t.Fatalf("inspect layout exit = %d stderr=%s", code, stderr)
	}
	if layoutResult.Layout == nil || !layoutResult.Layout.Valid || layoutResult.Layout.Mode != "convention_first" {
		t.Fatalf("layout result = %+v", layoutResult)
	}
}

func TestRunnerValidationRejectionDoesNotCreateDatabase(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	request := `{"action":"create_document","document":{"title":"Missing path","body":"# Missing path\n"}}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code != 0 {
		t.Fatalf("exit = %d stderr=%s", code, stderr)
	}
	if !result.Rejected || result.RejectionReason == "" {
		t.Fatalf("result = %+v, want rejection", result)
	}
	if _, err := os.Stat(filepath.Dir(dbPath)); !os.IsNotExist(err) {
		t.Fatalf("data dir exists after rejected request: %v", err)
	}
}

func TestRunnerErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		args   []string
		input  string
		want   int
		stderr string
	}{
		{name: "unknown command", args: []string{"unknown"}, input: `{}`, want: 2, stderr: "unknown openclerk command"},
		{name: "bad json", args: []string{"document"}, input: `{`, want: 1, stderr: "decode document request"},
		{name: "multiple json", args: []string{"document"}, input: `{} {}`, want: 1, stderr: "multiple JSON values"},
		{name: "unknown json field", args: []string{"document"}, input: `{"action":"validate","extra":true}`, want: 1, stderr: "unknown field"},
		{name: "unexpected arg", args: []string{"retrieval", "extra"}, input: `{}`, want: 2, stderr: "unexpected positional arguments"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := run(tt.args, strings.NewReader(tt.input), &stdout, &stderr)
			if code != tt.want {
				t.Fatalf("exit = %d, want %d; stderr=%s", code, tt.want, stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.stderr) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.stderr)
			}
		})
	}
}

func TestRunnerRuntimeErrorExitsNonZero(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "data", "openclerk.sqlite")
	request := `{"action":"get_document","doc_id":"missing"}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero")
	}
	if !strings.Contains(stderr, "run document task") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestRunnerDBFlag(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "custom", "openclerk.sqlite")
	request := `{"action":"resolve_paths"}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code != 0 {
		t.Fatalf("exit = %d stderr=%s", code, stderr)
	}
	if result.Paths == nil || result.Paths.DatabasePath != dbPath {
		t.Fatalf("paths = %+v, want db %q", result.Paths, dbPath)
	}
	if result.Paths.VaultRoot != filepath.Join(filepath.Dir(dbPath), "vault") {
		t.Fatalf("paths = %+v, want default sibling vault", result.Paths)
	}
}

func TestRunnerInitBindsVaultRoot(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "custom", "openclerk.sqlite")
	vaultRoot := filepath.Join(t.TempDir(), "wiki")
	var initResult struct {
		Paths runner.Paths `json:"paths"`
	}
	code, stderr := runJSON(t, []string{"init", "--db", dbPath, "--vault-root", vaultRoot}, "", &initResult)
	if code != 0 {
		t.Fatalf("init exit = %d stderr=%s", code, stderr)
	}
	if initResult.Paths.DatabasePath != dbPath || initResult.Paths.VaultRoot != vaultRoot {
		t.Fatalf("init paths = %+v", initResult.Paths)
	}

	request := `{"action":"resolve_paths"}`
	var result runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--db", dbPath}, request, &result)
	if code != 0 {
		t.Fatalf("resolve exit = %d stderr=%s", code, stderr)
	}
	if result.Paths == nil || result.Paths.VaultRoot != vaultRoot {
		t.Fatalf("paths = %+v, want vault %q", result.Paths, vaultRoot)
	}
}

func runJSON(t *testing.T, args []string, input string, output any) (int, string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(args, strings.NewReader(input), &stdout, &stderr)
	if output != nil && stdout.Len() > 0 {
		if err := json.Unmarshal(stdout.Bytes(), output); err != nil {
			t.Fatalf("decode stdout %q: %v", stdout.String(), err)
		}
	}
	return code, stderr.String()
}
