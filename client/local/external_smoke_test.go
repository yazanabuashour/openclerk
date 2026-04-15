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
	records "github.com/yazanabuashour/openclerk/client/records"
)

func main() {
	client, runtime, err := local.OpenRecords(local.Config{})
	if err != nil {
		panic(err)
	}
	defer runtime.Close()

	create, err := client.CreateDocumentWithResponse(context.Background(), records.CreateDocumentRequest{
		Path:  "health/glucose.md",
		Title: "Fasting glucose",
		Body:  "---\nentity_type: lab_result\nentity_name: Fasting glucose\nentity_id: fasting-glucose\n---\n# Fasting glucose\n\n## Summary\nRoutine health observation.\n\n## Facts\n- value: 92 mg/dL\n- status: normal\n",
	})
	if err != nil {
		panic(err)
	}
	if create.JSON201 == nil {
		panic(string(create.Body))
	}

	lookup, err := client.RecordsLookupWithResponse(context.Background(), records.RecordsLookupRequest{Text: "glucose"})
	if err != nil {
		panic(err)
	}
	if lookup.JSON200 == nil || len(lookup.JSON200.Entities) != 1 {
		panic(string(lookup.Body))
	}

	entity, err := client.GetRecordEntityWithResponse(context.Background(), lookup.JSON200.Entities[0].EntityId)
	if err != nil {
		panic(err)
	}
	if entity.JSON200 == nil {
		panic(string(entity.Body))
	}

	fmt.Printf("backend=%s dataDir=%s entity=%s facts=%d doc=%s\n", records.CapabilitiesBackendRecords, runtime.Paths().DataDir, entity.JSON200.EntityId, len(entity.JSON200.Facts), create.JSON201.DocId)
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

	install := exec.Command("go", "get", "github.com/yazanabuashour/openclerk/client/local@v0.0.0")
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
	expectedPrefix := fmt.Sprintf("backend=records dataDir=%s entity=fasting-glucose facts=2 ", filepath.Join(xdgDataHome, "openclerk"))
	if !strings.HasPrefix(got, expectedPrefix) {
		t.Fatalf("embedded smoke output = %q, want prefix %q", got, expectedPrefix)
	}

	if _, err := os.Stat(filepath.Join(xdgDataHome, "openclerk", "openclerk.sqlite")); err != nil {
		t.Fatalf("stat sqlite database: %v", err)
	}
	if _, err := os.Stat(filepath.Join(xdgDataHome, "openclerk", "vault", "health", "glucose.md")); err != nil {
		t.Fatalf("stat canonical vault document: %v", err)
	}
}
