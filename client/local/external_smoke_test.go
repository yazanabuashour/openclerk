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
)

func main() {
	client, err := local.OpenClient(local.Config{})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	create, err := client.CreateDocument(context.Background(), local.DocumentInput{
		Path:  "notes/ops/agent-knowledge-plane.md",
		Title: "Agent knowledge plane",
		Body:  "---\ntype: note\nstatus: active\n---\n# Agent knowledge plane\n\n## Summary\nCanonical agent-facing context.\n\n## Related\nSee [operations](operations.md).\n",
	})
	if err != nil {
		panic(err)
	}

	related, err := client.CreateDocument(context.Background(), local.DocumentInput{
		Path:  "notes/ops/operations.md",
		Title: "Operations",
		Body:  "---\ntype: runbook\nstatus: draft\n---\n# Operations\n\n## Summary\nOperational notes for the workspace.\n",
	})
	if err != nil {
		panic(err)
	}
	_ = related

	record, err := client.CreateDocument(context.Background(), local.DocumentInput{
		Path:  "records/assets/transmission-solenoid.md",
		Title: "Transmission solenoid",
		Body:  "---\nentity_type: part\nentity_name: Transmission solenoid\nentity_id: transmission-solenoid\ntype: record\nstatus: active\n---\n# Transmission solenoid\n\n## Summary\nCanonical promoted-domain baseline.\n\n## Facts\n- sku: SOL-1\n- vendor: OpenClerk Motors\n",
	})
	if err != nil {
		panic(err)
	}
	_ = record

	list, err := client.ListDocuments(context.Background(), local.DocumentListOptions{PathPrefix: "notes/"})
	if err != nil {
		panic(err)
	}
	if len(list.Documents) != 2 {
		panic(fmt.Sprintf("docs=%d", len(list.Documents)))
	}

	links, err := client.GetDocumentLinks(context.Background(), create.DocID)
	if err != nil {
		panic(err)
	}
	if len(links.Outgoing) != 1 {
		panic(fmt.Sprintf("links=%d", len(links.Outgoing)))
	}

	lookup, err := client.LookupRecords(context.Background(), local.RecordLookupOptions{Text: "solenoid"})
	if err != nil {
		panic(err)
	}
	if len(lookup.Entities) != 1 {
		panic(fmt.Sprintf("entities=%d", len(lookup.Entities)))
	}

	events, err := client.ListProvenanceEvents(context.Background(), local.ProvenanceEventOptions{RefKind: "document", RefID: create.DocID})
	if err != nil {
		panic(err)
	}
	if len(events.Events) == 0 {
		panic("missing events")
	}

	fmt.Printf("backend=%s dataDir=%s docs=%d links=%d entity=%s events=%d\n", "openclerk", client.Paths().DataDir, len(list.Documents), len(links.Outgoing), lookup.Entities[0].EntityID, len(events.Events))
}
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
