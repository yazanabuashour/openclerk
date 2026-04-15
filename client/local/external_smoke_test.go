package local_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedClientsExternalModuleSmoke(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	mainFile := `package main

import (
	"context"
	"fmt"

	local "github.com/yazanabuashour/openclerk/client/local"
	openclerk "github.com/yazanabuashour/openclerk/client/openclerk"
)

func main() {
	client, runtime, err := local.Open(local.Config{})
	if err != nil {
		panic(err)
	}
	defer runtime.Close()

	create, err := client.CreateDocumentWithResponse(context.Background(), openclerk.CreateDocumentRequest{
		Path:  "notes/ops/agent-knowledge-plane.md",
		Title: "Agent knowledge plane",
		Body:  "---\ntype: note\nstatus: active\n---\n# Agent knowledge plane\n\n## Summary\nCanonical agent-facing context.\n\n## Related\nSee [operations](operations.md).\n",
	})
	if err != nil {
		panic(err)
	}
	if create.JSON201 == nil {
		panic(string(create.Body))
	}

	related, err := client.CreateDocumentWithResponse(context.Background(), openclerk.CreateDocumentRequest{
		Path:  "notes/ops/operations.md",
		Title: "Operations",
		Body:  "---\ntype: runbook\nstatus: draft\n---\n# Operations\n\n## Summary\nOperational notes for the workspace.\n",
	})
	if err != nil {
		panic(err)
	}
	if related.JSON201 == nil {
		panic(string(related.Body))
	}

	record, err := client.CreateDocumentWithResponse(context.Background(), openclerk.CreateDocumentRequest{
		Path:  "records/assets/transmission-solenoid.md",
		Title: "Transmission solenoid",
		Body:  "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\ntype: record\nstatus: active\n---\n# Transmission solenoid\n\n## Summary\nCanonical promoted-domain baseline.\n\n## Facts\n- sku: SOL-1\n- vendor: OpenClerk Motors\n",
	})
	if err != nil {
		panic(err)
	}
	if record.JSON201 == nil {
		panic(string(record.Body))
	}

	list, err := client.ListDocumentsWithResponse(context.Background(), &openclerk.ListDocumentsParams{PathPrefix: ptr("notes/")})
	if err != nil {
		panic(err)
	}
	if list.JSON200 == nil || len(list.JSON200.Documents) != 2 {
		panic(string(list.Body))
	}

	links, err := client.GetDocumentLinksWithResponse(context.Background(), create.JSON201.DocId)
	if err != nil {
		panic(err)
	}
	if links.JSON200 == nil || len(links.JSON200.Outgoing) != 1 {
		panic(string(links.Body))
	}

	lookup, err := client.RecordsLookupWithResponse(context.Background(), openclerk.RecordsLookupRequest{Text: "solenoid"})
	if err != nil {
		panic(err)
	}
	if lookup.JSON200 == nil || len(lookup.JSON200.Entities) != 1 {
		panic(string(lookup.Body))
	}

	events, err := client.ListProvenanceEventsWithResponse(context.Background(), &openclerk.ListProvenanceEventsParams{RefKind: ptr("document"), RefId: &create.JSON201.DocId})
	if err != nil {
		panic(err)
	}
	if events.JSON200 == nil || len(events.JSON200.Events) == 0 {
		panic(string(events.Body))
	}

	fmt.Printf("backend=%s dataDir=%s docs=%d links=%d entity=%s events=%d\n", openclerk.CapabilitiesBackendOpenclerk, runtime.Paths().DataDir, len(list.JSON200.Documents), len(links.JSON200.Outgoing), lookup.JSON200.Entities[0].EntityId, len(events.JSON200.Events))
}

func ptr(value string) *string { return &value }
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainFile), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	goMod := fmt.Sprintf(`module example.com/openclerk-embedded-smoke

go 1.26.0

replace github.com/yazanabuashour/openclerk => %s
`, repoRoot)
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	install := exec.Command("go", "get", "github.com/yazanabuashour/openclerk/client/local@v0.1.0")
	install.Dir = tmpDir
	install.Env = os.Environ()
	installOutput, err := install.CombinedOutput()
	if err != nil {
		t.Fatalf("go get embedded smoke: %v\n%s", err, string(installOutput))
	}

	xdgDataHome := filepath.Join(tmpDir, "xdg")
	run := exec.Command("go", "run", ".")
	run.Dir = tmpDir
	run.Env = append(os.Environ(), "XDG_DATA_HOME="+xdgDataHome)
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("go run embedded smoke: %v\n%s", err, string(output))
	}

	got := strings.TrimSpace(string(output))
	expectedPrefix := fmt.Sprintf("backend=openclerk dataDir=%s docs=2 links=1 entity=transmission-solenoid events=", filepath.Join(xdgDataHome, "openclerk"))
	if !strings.HasPrefix(got, expectedPrefix) {
		t.Fatalf("embedded smoke output = %q, want prefix %q", got, expectedPrefix)
	}

	if _, err := os.Stat(filepath.Join(xdgDataHome, "openclerk", "openclerk.sqlite")); err != nil {
		t.Fatalf("stat sqlite database: %v", err)
	}
	if _, err := os.Stat(filepath.Join(xdgDataHome, "openclerk", "vault", "notes", "ops", "agent-knowledge-plane.md")); err != nil {
		t.Fatalf("stat canonical vault document: %v", err)
	}
}
