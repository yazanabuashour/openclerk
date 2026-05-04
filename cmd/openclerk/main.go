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
	gitCheckpoints := fs.Bool("git-checkpoints", false, "enable explicit local git checkpoint writes")
	if err := fs.Parse(args); err != nil {
		return runclient.Config{}, false
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected positional arguments: %v\n", fs.Args())
		return runclient.Config{}, false
	}
	return runclient.Config{DatabasePath: *databasePath, GitCheckpoints: *gitCheckpoints}, true
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
	_, _ = fmt.Fprintln(stderr, "promoted workflow actions: compile_synthesis, ingest_source_url plan, git_lifecycle_report, source_audit_report, evidence_bundle_report, duplicate_candidate_report, workflow_guide_report, memory_router_recall_report, structured_store_report, hybrid_retrieval_report")
}

func documentUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "usage: openclerk document [--db path] [--git-checkpoints] < request.json")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Reads one strict JSON object from stdin and writes one JSON result.")
	_, _ = fmt.Fprintln(w, "Uses configured paths by default; pass --db only for an explicit dataset.")
	_, _ = fmt.Fprintln(w, "Primitive request shapes:")
	_, _ = fmt.Fprintln(w, `  validate/create_document: {"action":"validate","document":{"path":"notes/example.md","title":"Example","body":"# Example\n\nBody."}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url PDF create: {"action":"ingest_source_url","source":{"url":"https://example.test/source.pdf","path_hint":"sources/example.md","asset_path_hint":"assets/sources/example.pdf","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url web create: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","path_hint":"sources/web/example.md","source_type":"web","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url update: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"update","source_type":"web"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_source_url placement plan: {"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"plan","source_type":"web","title":"Optional title"}}`)
	_, _ = fmt.Fprintln(w, `  ingest_video_url create: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","path_hint":"sources/video-youtube/demo.md","transcript":{"text":"Supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}`)
	_, _ = fmt.Fprintln(w, `  ingest_video_url update: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","mode":"update","transcript":{"text":"Updated supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}`)
	_, _ = fmt.Fprintln(w, `  list/get/edit: {"action":"list_documents","list":{"path_prefix":"notes/","limit":20}} | {"action":"get_document","doc_id":"doc_id_from_json"} | {"action":"replace_section","doc_id":"doc_id_from_json","heading":"Summary","content":"Updated summary."}`)
	_, _ = fmt.Fprintln(w, `  diagnostics: {"action":"resolve_paths"} | {"action":"inspect_layout"}`)
	_, _ = fmt.Fprintln(w, "Promoted workflow action:")
	_, _ = fmt.Fprintln(w, `  compile_synthesis: {"action":"compile_synthesis","synthesis":{"path":"synthesis/example.md","title":"Example","source_refs":["sources/a.md"],"body":"...","body_facts":["..."],"freshness_note":"...","mode":"create_or_update"}}`)
	_, _ = fmt.Fprintln(w, "  Requires path, title, non-empty source_refs, and either body or body_facts. mode defaults to create_or_update.")
	_, _ = fmt.Fprintln(w, "  Returns compile_synthesis.agent_handoff with final-answer evidence; use primitives only after rejection or explicit drill-down.")
	_, _ = fmt.Fprintln(w, `  git_lifecycle_report status/history: {"action":"git_lifecycle_report","git_lifecycle":{"mode":"status","paths":["synthesis/example.md"],"limit":10}}`)
	_, _ = fmt.Fprintln(w, `  git_lifecycle_report checkpoint: {"action":"git_lifecycle_report","git_lifecycle":{"mode":"checkpoint","paths":["synthesis/example.md"],"message":"openclerk: update synthesis example"}}`)
	_, _ = fmt.Fprintln(w, "  Status/history are read-only. Checkpoint requires --git-checkpoints or OPENCLERK_GIT_CHECKPOINTS=1, never pushes, switches branches, restores, or emits raw diffs.")
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
	_, _ = fmt.Fprintln(w, `  duplicate_candidate_report: {"action":"duplicate_candidate_report","duplicate_candidate":{"query":"renewal packaging notes","path_prefix":"notes/","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns duplicate_candidate.agent_handoff with likely target, evidence inspected, no-write status, and approval boundary.")
	_, _ = fmt.Fprintln(w, `  workflow_guide_report: {"action":"workflow_guide_report","workflow_guide":{"intent":"should I update an existing note or create a new one?"}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns workflow_guide.agent_handoff with runner-owned surface selection guidance; it does not inspect storage or replace the selected action result.")
	_, _ = fmt.Fprintln(w, `  memory_router_recall_report: {"action":"memory_router_recall_report","memory_router_recall":{"query":"memory router temporal recall session promotion feedback weighting routing canonical docs","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns memory_router_recall.agent_handoff with canonical refs, stale-session posture, feedback weighting, provenance, freshness, and authority limits.")
	_, _ = fmt.Fprintln(w, `  structured_store_report: {"action":"structured_store_report","structured_store":{"domain":"records","query":"structured canonical record evidence","entity_type":"tool","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns structured_store.agent_handoff with promoted record/service/decision projection evidence, candidate-store boundaries, freshness, and authority limits.")
	_, _ = fmt.Fprintln(w, `  hybrid_retrieval_report: {"action":"hybrid_retrieval_report","hybrid_retrieval":{"query":"semantic recall citation quality","path_prefix":"docs/","limit":10}}`)
	_, _ = fmt.Fprintln(w, "  Read-only. Returns hybrid_retrieval.agent_handoff with lexical baseline evidence and candidate-surface boundaries; it does not create vectors or change default ranking.")
}
