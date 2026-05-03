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
	if wantsSubcommandHelp(args) {
		documentUsage(stdout)
		return 0
	}
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
	if wantsSubcommandHelp(args) {
		retrievalUsage(stdout)
		return 0
	}
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
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, false
	}
	return runclient.Config{DatabasePath: *databasePath}, true
}

func wantsSubcommandHelp(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "help", "-h", "--help":
			return true
		}
	}
	return false
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
	_, _ = fmt.Fprintln(stderr, "usage: openclerk <version|init|document|retrieval> [--db path]")
	_, _ = fmt.Fprintln(stderr, "       openclerk init [--db path] [--vault-root path]")
	_, _ = fmt.Fprintln(stderr, "       openclerk document --help")
	_, _ = fmt.Fprintln(stderr, "       openclerk retrieval --help")
	_, _ = fmt.Fprintln(stderr, "document/retrieval read strict JSON from stdin and use configured paths by default; pass --db only for an explicit dataset.")
	_, _ = fmt.Fprintln(stderr, "promoted workflow actions: compile_synthesis, source_audit_report, evidence_bundle_report")
}

func documentUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk document [--db path] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Uses configured paths by default; pass --db only for an explicit dataset.")
	_, _ = fmt.Fprintln(w, "Promoted workflow action:")
	_, _ = fmt.Fprintln(w, `  compile_synthesis: {"action":"compile_synthesis","synthesis":{"path":"synthesis/example.md","title":"Example","source_refs":["sources/a.md"],"body":"...","body_facts":["..."],"freshness_note":"...","mode":"create_or_update"}}`)
	_, _ = fmt.Fprintln(w, "  Requires path, title, non-empty source_refs, and either body or body_facts. mode defaults to create_or_update.")
	_, _ = fmt.Fprintln(w, "  Returns compile_synthesis.agent_handoff with final-answer evidence; use primitives only after rejection or explicit drill-down.")
}

func retrievalUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk retrieval [--db path] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Uses configured paths by default; pass --db only for an explicit dataset.")
	_, _ = fmt.Fprintln(w, "Promoted workflow actions:")
	_, _ = fmt.Fprintln(w, `  source_audit_report: {"action":"source_audit_report","source_audit":{"query":"...","target_path":"synthesis/example.md","mode":"explain","conflict_query":"...","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Default mode is explain; repair_existing may update only an existing synthesis target.")
	_, _ = fmt.Fprintln(w, `  evidence_bundle_report: {"action":"evidence_bundle_report","evidence_bundle":{"query":"...","entity_id":"...","decision_id":"...","ref_kind":"document","ref_id":"...","projection":"records","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns evidence_bundle.agent_handoff with citations, provenance, projection freshness, validation boundaries, and authority limits.")
}
