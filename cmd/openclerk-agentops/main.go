package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yazanabuashour/openclerk/agentops"
	"github.com/yazanabuashour/openclerk/client/local"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}

	switch args[0] {
	case "document":
		return runDocument(args[1:], stdin, stdout, stderr)
	case "retrieval":
		return runRetrieval(args[1:], stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown openclerk-agentops command %q\n", args[0])
		usage(stderr)
		return 2
	}
}

func runDocument(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	config, ok := parseConfig("document", args, stderr)
	if !ok {
		return 2
	}
	var request agentops.DocumentTaskRequest
	if err := json.NewDecoder(stdin).Decode(&request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode document request: %v\n", err)
		return 1
	}
	result, err := agentops.RunDocumentTask(context.Background(), config, request)
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
	var request agentops.RetrievalTaskRequest
	if err := json.NewDecoder(stdin).Decode(&request); err != nil {
		_, _ = fmt.Fprintf(stderr, "decode retrieval request: %v\n", err)
		return 1
	}
	result, err := agentops.RunRetrievalTask(context.Background(), config, request)
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

func parseConfig(name string, args []string, stderr io.Writer) (local.Config, bool) {
	fs := flag.NewFlagSet("openclerk-agentops "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	dataDir := fs.String("data-dir", "", "OpenClerk data directory")
	databasePath := fs.String("db", "", "OpenClerk SQLite database path")
	vaultRoot := fs.String("vault-root", "", "OpenClerk vault root")
	embeddingProvider := fs.String("embedding-provider", "", "embedding provider name")
	if err := fs.Parse(args); err != nil {
		return local.Config{}, false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return local.Config{}, false
	}
	return local.Config{
		DataDir:           *dataDir,
		DatabasePath:      *databasePath,
		VaultRoot:         *vaultRoot,
		EmbeddingProvider: *embeddingProvider,
	}, true
}

func usage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, "usage: openclerk-agentops <document|retrieval> [--data-dir path] [--db path] [--vault-root path] [--embedding-provider name]")
}
