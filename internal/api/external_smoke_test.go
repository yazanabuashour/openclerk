package api_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedOpenClerkClientExternalModuleSmoke(t *testing.T) {
	t.Parallel()

	serverURL := newTestServer(t, "")
	tmpDir := t.TempDir()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	mainFile := `package main

import (
	"context"
	"fmt"
	"os"

	openclerk "github.com/yazanabuashour/openclerk/client/openclerk"
)

func main() {
	client, err := openclerk.NewClientWithResponses(os.Getenv("OPENCLERK_SERVER"))
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
`
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
	install := exec.Command("go", "get", "github.com/yazanabuashour/openclerk/client/openclerk@v0.0.0")
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
	if strings.TrimSpace(string(output)) != "openclerk" {
		t.Fatalf("external smoke output = %q, want openclerk", strings.TrimSpace(string(output)))
	}
}
