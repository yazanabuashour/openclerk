package api_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

func TestGeneratedClientsExternalModuleSmoke(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		backend    domain.BackendKind
		importPath string
		newClient  string
	}{
		{
			name:       "fts",
			backend:    domain.BackendFTS,
			importPath: "github.com/yazanabuashour/openclerk/client/fts",
			newClient:  "fts.NewClientWithResponses",
		},
		{
			name:       "hybrid",
			backend:    domain.BackendHybrid,
			importPath: "github.com/yazanabuashour/openclerk/client/hybrid",
			newClient:  "hybrid.NewClientWithResponses",
		},
		{
			name:       "graph",
			backend:    domain.BackendGraph,
			importPath: "github.com/yazanabuashour/openclerk/client/graph",
			newClient:  "graph.NewClientWithResponses",
		},
		{
			name:       "records",
			backend:    domain.BackendRecords,
			importPath: "github.com/yazanabuashour/openclerk/client/records",
			newClient:  "records.NewClientWithResponses",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			serverURL := newTestServer(t, tc.backend, "")
			tmpDir := t.TempDir()
			repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
			if err != nil {
				t.Fatalf("repo root: %v", err)
			}
			mainFile := externalSmokeProgram(tc.importPath, tc.newClient)
			if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainFile), 0o644); err != nil {
				t.Fatalf("write main.go: %v", err)
			}
			goMod := fmt.Sprintf(`module example.com/openclerk-smoke

go 1.26.0

replace github.com/yazanabuashour/openclerk => %s
`, repoRoot)
			if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644); err != nil {
				t.Fatalf("write go.mod: %v", err)
			}
			install := exec.Command("go", "get", tc.importPath+"@v0.0.0")
			install.Dir = tmpDir
			install.Env = os.Environ()
			installOutput, err := install.CombinedOutput()
			if err != nil {
				t.Fatalf("go get external smoke: %v\n%s", err, string(installOutput))
			}
			cmd := exec.Command("go", "run", ".")
			cmd.Dir = tmpDir
			cmd.Env = append(os.Environ(), "OPENCLERK_SERVER="+serverURL)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go run external smoke: %v\n%s", err, string(output))
			}
			if strings.TrimSpace(string(output)) != string(tc.backend) {
				t.Fatalf("external smoke output = %q, want %q", strings.TrimSpace(string(output)), tc.backend)
			}
		})
	}
}

func externalSmokeProgram(importPath string, constructor string) string {
	packageName := pathBase(importPath)
	return fmt.Sprintf(`package main

import (
	"context"
	"fmt"
	"os"

	%s "%s"
)

func main() {
	client, err := %s(os.Getenv("OPENCLERK_SERVER"))
	if err != nil {
		panic(err)
	}
	response, err := client.GetCapabilitiesWithResponse(context.Background())
	if err != nil {
		panic(err)
	}
	if response.JSON200 == nil {
		panic(string(response.Body))
	}
	fmt.Println(response.JSON200.Backend)
}
`, packageName, importPath, constructor)
}

func pathBase(value string) string {
	parts := strings.Split(value, "/")
	return parts[len(parts)-1]
}
