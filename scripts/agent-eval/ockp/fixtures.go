package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yazanabuashour/openclerk/internal/runner"
)

type sourceURLUpdateFixtures struct {
	server          *httptest.Server
	mu              sync.Mutex
	initialPDF      []byte
	changedPDF      []byte
	serveChangedPDF bool
	artifactPDF     bool
}

func startSourceURLUpdateFixtures(scenarioID string) *sourceURLUpdateFixtures {
	if !isSourceURLUpdateScenario(scenarioID) && !isArtifactPDFScenario(scenarioID) {
		return nil
	}
	fixtures := &sourceURLUpdateFixtures{
		initialPDF: minimalEvalPDF("Source URL Update Stable", "OpenClerk Eval", sourceURLUpdateInitialText),
		changedPDF: minimalEvalPDF("Source URL Update Changed", "OpenClerk Eval", sourceURLUpdateChangedText),
	}
	if isArtifactPDFScenario(scenarioID) {
		fixtures.initialPDF = minimalEvalPDF("Artifact PDF Source", "OpenClerk Eval", artifactPDFEvidenceText)
		fixtures.artifactPDF = true
		return fixtures
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/stable.pdf", func(w http.ResponseWriter, _ *http.Request) {
		fixtures.mu.Lock()
		changed := fixtures.serveChangedPDF
		fixtures.mu.Unlock()
		body := fixtures.initialPDF
		if changed {
			body = fixtures.changedPDF
		}
		servePDF(w, body)
	})
	fixtures.server = httptest.NewServer(mux)
	return fixtures
}
func (f *sourceURLUpdateFixtures) Close() {
	if f != nil && f.server != nil {
		f.server.Close()
	}
}
func (f *sourceURLUpdateFixtures) stableURL() string {
	if f.artifactPDF {
		return artifactPDFEvalSourceURL
	}
	return f.server.URL + "/stable.pdf"
}
func (f *sourceURLUpdateFixtures) changedURL() string {
	return f.stableURL()
}
func (f *sourceURLUpdateFixtures) prepareForAgent(scenarioID string) {
	if scenarioID != sourceURLUpdateChangedScenarioID {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.serveChangedPDF = true
}
func (f *sourceURLUpdateFixtures) renderPrompt(prompt string) string {
	if f == nil {
		return prompt
	}
	prompt = strings.ReplaceAll(prompt, sourceURLUpdateStableURLToken, f.stableURL())
	prompt = strings.ReplaceAll(prompt, sourceURLUpdateChangedURLToken, f.changedURL())
	return strings.ReplaceAll(prompt, artifactPDFSourceURLToken, f.stableURL())
}
func (f *sourceURLUpdateFixtures) prepareFiles(runDir string) error {
	if f == nil || !f.artifactPDF {
		return nil
	}
	target := filepath.Join(evalSourceFixtureRoot(runDir), "artifacts", "vendor-security-paper.pdf")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, f.initialPDF, 0o644)
}
func evalSourceFixtureRoot(runDir string) string {
	return filepath.Join(runDir, "source-fixtures")
}
func runArtifactPDFFixturePreflight(ctx context.Context, runDir string, paths evalPaths, cache cacheConfig, fixtures *sourceURLUpdateFixtures) fixturePreflight {
	const sourcePath = "sources/artifacts/preflight-vendor-security-paper.md"
	const assetPath = "assets/sources/artifacts/preflight-vendor-security-paper.pdf"
	result := fixturePreflight{
		Name:       "artifact_pdf_source_url_fixture",
		Documents:  []string{sourcePath},
		SourcePath: sourcePath,
		AssetPath:  assetPath,
	}
	if fixtures == nil {
		result.Details = "missing PDF fixture server"
		return result
	}
	preflightRunDir := filepath.Join(runDir, "fixture-preflight")
	preflightPaths := paths
	preflightPaths.DatabasePath = filepath.Join(preflightRunDir, "openclerk-preflight.db")
	if err := os.MkdirAll(filepath.Join(preflightRunDir, "tmp"), 0o755); err != nil {
		result.Details = err.Error()
		return result
	}
	request := runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           fixtures.stableURL(),
			PathHint:      sourcePath,
			AssetPathHint: assetPath,
			Title:         "Vendor Security Paper Preflight",
		},
	}
	body, err := json.Marshal(request)
	if err != nil {
		result.Details = err.Error()
		return result
	}
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, filepath.Join(runDir, "bin", "openclerk"), "document")
	cmd.Dir = runDir
	cmd.Env = evalEnv(runDir, preflightPaths, cache)
	cmd.Stdin = bytes.NewReader(body)
	output, err := cmd.CombinedOutput()
	if cmdCtx.Err() == context.DeadlineExceeded {
		result.Details = "preflight timed out"
		return result
	}
	if err != nil {
		result.Details = fmt.Sprintf("%v: %s", err, strings.TrimSpace(string(output)))
		return result
	}
	var decoded runner.DocumentTaskResult
	if err := json.Unmarshal(output, &decoded); err != nil {
		result.Details = fmt.Sprintf("decode preflight result: %v", err)
		return result
	}
	if decoded.Rejected {
		result.Details = "preflight rejected: " + decoded.RejectionReason
		return result
	}
	if decoded.Ingestion == nil {
		result.Details = "preflight returned no ingestion result"
		return result
	}
	if decoded.Ingestion.SourcePath != sourcePath || decoded.Ingestion.AssetPath != assetPath || len(decoded.Ingestion.Citations) == 0 {
		result.Details = fmt.Sprintf("unexpected preflight ingestion source=%q asset=%q citations=%d", decoded.Ingestion.SourcePath, decoded.Ingestion.AssetPath, len(decoded.Ingestion.Citations))
		return result
	}
	result.Passed = true
	result.Details = "generated HTTP PDF ingested through built openclerk binary"
	return result
}
func servePDF(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/pdf")
	_, _ = w.Write(body)
}
func minimalEvalPDF(title string, author string, text string) []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, 6)
	writeObject := func(id int, body string) {
		offsets = append(offsets, buf.Len())
		_, _ = fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", id, body)
	}
	writeObject(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObject(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObject(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>")
	writeObject(4, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	stream := fmt.Sprintf("BT /F1 24 Tf 72 720 Td (%s) Tj ET", pdfEscape(text))
	writeObject(5, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream))
	writeObject(6, fmt.Sprintf("<< /Title (%s) /Author (%s) /CreationDate (D:20260426000000Z) >>", pdfEscape(title), pdfEscape(author)))
	xrefStart := buf.Len()
	buf.WriteString("xref\n0 7\n")
	buf.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		_, _ = fmt.Fprintf(&buf, "%010d 00000 n \n", offset)
	}
	_, _ = fmt.Fprintf(&buf, "trailer\n<< /Size 7 /Root 1 0 R /Info 6 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefStart)
	return buf.Bytes()
}
func pdfEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "(", `\(`)
	value = strings.ReplaceAll(value, ")", `\)`)
	return value
}
