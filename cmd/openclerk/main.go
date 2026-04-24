package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

var version string

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		usage(stdout)
		return 0
	case "version", "--version":
		writeVersion(stdout)
		return 0
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "document":
		return runDocument(args[1:], stdin, stdout, stderr)
	case "retrieval":
		return runRetrieval(args[1:], stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown openclerk command %q\n", args[0])
		usage(stderr)
		return 2
	}
}

func runInit(args []string, stdout io.Writer, stderr io.Writer) int {
	config, vaultRoot, ok := parseInitConfig(args, stderr)
	if !ok {
		return 2
	}
	paths, err := runclient.InitializePaths(config, vaultRoot)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "initialize OpenClerk paths: %v\n", err)
		return 1
	}
	result := struct {
		Paths   runner.Paths `json:"paths"`
		Summary string       `json:"summary"`
	}{
		Paths: runner.Paths{
			DataDir:      paths.DataDir,
			DatabasePath: paths.DatabasePath,
			VaultRoot:    paths.VaultRoot,
		},
		Summary: "initialized OpenClerk paths",
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode init result: %v\n", err)
		return 1
	}
	return 0
}

func runDocument(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	config, ok := parseConfig("document", args, stderr)
	if !ok {
		return 2
	}
	var request runner.DocumentTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode document request: %v\n", err)
		return 1
	}
	result, err := runner.RunDocumentTask(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run document task: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode document result: %v\n", err)
		return 1
	}
	return 0
}

func runRetrieval(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	config, ok := parseConfig("retrieval", args, stderr)
	if !ok {
		return 2
	}
	var request runner.RetrievalTaskRequest
	if err := decodeRequest(stdin, &request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode retrieval request: %v\n", err)
		return 1
	}
	result, err := runner.RunRetrievalTask(context.Background(), config, request)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "run retrieval task: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "encode retrieval result: %v\n", err)
		return 1
	}
	return 0
}

func parseConfig(name string, args []string, stderr io.Writer) (runclient.Config, bool) {
	fs := flag.NewFlagSet("openclerk "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	databasePath := fs.String("db", "", "OpenClerk SQLite database path")
	embeddingProvider := fs.String("embedding-provider", "", "embedding provider name")
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, false
	}
	return runclient.Config{
		DatabasePath:      *databasePath,
		EmbeddingProvider: *embeddingProvider,
	}, true
}

func parseInitConfig(args []string, stderr io.Writer) (runclient.Config, string, bool) {
	fs := flag.NewFlagSet("openclerk init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	databasePath := fs.String("db", "", "OpenClerk SQLite database path")
	vaultRoot := fs.String("vault-root", "", "OpenClerk vault root")
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, "", false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, "", false
	}
	return runclient.Config{DatabasePath: *databasePath}, *vaultRoot, true
}

func decodeRequest[T any](stdin io.Reader, request *T) error {
	decoder := json.NewDecoder(stdin)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		return err
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values are not supported")
		}
		return err
	}
	return nil
}

func writeVersion(w io.Writer) {
	info, ok := readBuildInfo()
	_, _ = fmt.Fprintf(w, "openclerk %s\n", resolvedVersion(version, info, ok))
}

func readBuildInfo() (*debug.BuildInfo, bool) {
	return debug.ReadBuildInfo()
}

func resolvedVersion(linkerVersion string, info *debug.BuildInfo, ok bool) string {
	if linkerVersion != "" {
		return linkerVersion
	}
	if ok && info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func usage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, "usage: openclerk <version|init|document|retrieval> [--db path] [--embedding-provider name]")
	_, _ = fmt.Fprintln(stderr, "       openclerk init [--db path] [--vault-root path]")
}
