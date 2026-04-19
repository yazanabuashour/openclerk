package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/runner"
)

func TestRunnerDocumentAndRetrievalJSONRoundTrip(t *testing.T) {
	t.Parallel()

	dataDir := filepath.Join(t.TempDir(), "data")
	createRequest := `{"action":"create_document","document":{"path":"notes/runner.md","title":"Runner","body":"# Runner\n\n## Summary\nOpenClerk runner note.\n"}}`
	var createResult runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--data-dir", dataDir}, createRequest, &createResult)
	if code != 0 {
		t.Fatalf("create exit = %d stderr=%s", code, stderr)
	}
	if createResult.Document == nil || createResult.Document.DocID == "" {
		t.Fatalf("create result = %+v", createResult)
	}

	searchRequest := `{"action":"search","search":{"text":"runner","limit":10}}`
	var searchResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--data-dir", dataDir}, searchRequest, &searchResult)
	if code != 0 {
		t.Fatalf("search exit = %d stderr=%s", code, stderr)
	}
	if searchResult.Search == nil || len(searchResult.Search.Hits) == 0 {
		t.Fatalf("search result = %+v", searchResult)
	}

	serviceRequest := `{"action":"create_document","document":{"path":"records/services/openclerk-runner.md","title":"OpenClerk runner","body":"---\nservice_id: openclerk-runner\nservice_name: OpenClerk runner\nservice_status: active\nservice_owner: runner\nservice_interface: JSON runner\n---\n# OpenClerk runner\n\n## Summary\nProduction service.\n"}}`
	var serviceCreate runner.DocumentTaskResult
	code, stderr = runJSON(t, []string{"document", "--data-dir", dataDir}, serviceRequest, &serviceCreate)
	if code != 0 {
		t.Fatalf("create service exit = %d stderr=%s", code, stderr)
	}

	servicesRequest := `{"action":"services_lookup","services":{"text":"OpenClerk runner","interface":"JSON runner","limit":10}}`
	var servicesResult runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--data-dir", dataDir}, servicesRequest, &servicesResult)
	if code != 0 {
		t.Fatalf("services exit = %d stderr=%s", code, stderr)
	}
	if servicesResult.Services == nil || len(servicesResult.Services.Services) != 1 {
		t.Fatalf("services result = %+v", servicesResult)
	}

	serviceDetailRequest := `{"action":"service_record","service_id":"openclerk-runner"}`
	var serviceDetail runner.RetrievalTaskResult
	code, stderr = runJSON(t, []string{"retrieval", "--data-dir", dataDir}, serviceDetailRequest, &serviceDetail)
	if code != 0 {
		t.Fatalf("service detail exit = %d stderr=%s", code, stderr)
	}
	if serviceDetail.Service == nil || serviceDetail.Service.Interface != "JSON runner" {
		t.Fatalf("service detail = %+v", serviceDetail)
	}
}

func TestRunnerValidationRejectionDoesNotCreateDatabase(t *testing.T) {
	t.Parallel()

	dataDir := filepath.Join(t.TempDir(), "data")
	request := `{"action":"create_document","document":{"title":"Missing path","body":"# Missing path\n"}}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--data-dir", dataDir}, request, &result)
	if code != 0 {
		t.Fatalf("exit = %d stderr=%s", code, stderr)
	}
	if !result.Rejected || result.RejectionReason == "" {
		t.Fatalf("result = %+v, want rejection", result)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
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

	dataDir := filepath.Join(t.TempDir(), "data")
	request := `{"action":"get_document","doc_id":"missing"}`
	var result runner.DocumentTaskResult
	code, stderr := runJSON(t, []string{"document", "--data-dir", dataDir}, request, &result)
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
