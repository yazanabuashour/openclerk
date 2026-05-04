package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

type generatedCorpusSummary struct {
	GeneratedBytes           int64
	GeneratedDocuments       int
	GeneratedSources         int
	GeneratedSynthesis       int
	GeneratedDecisions       int
	GeneratedDuplicates      int
	GeneratedStale           int
	GeneratedTaggedDocuments int
}

type maturityReport struct {
	Metadata   maturityReportMetadata `json:"metadata"`
	Corpus     maturityCorpusSummary  `json:"corpus"`
	Timings    maturityTimingSummary  `json:"timings"`
	ReadProbes []maturityReadProbe    `json:"read_probes"`
	Checks     maturityChecks         `json:"checks"`
	Outcomes   []maturityOutcome      `json:"outcomes"`
}

type maturityReportMetadata struct {
	GeneratedAt                   time.Time `json:"generated_at"`
	Lane                          string    `json:"lane"`
	Mode                          string    `json:"mode"`
	Tier                          string    `json:"tier,omitempty"`
	Seed                          int64     `json:"seed,omitempty"`
	Harness                       string    `json:"harness"`
	RunRootArtifactReference      string    `json:"run_root_artifact_reference"`
	PrivateVaultArtifactReference string    `json:"private_vault_artifact_reference,omitempty"`
	RawLogsCommitted              bool      `json:"raw_logs_committed"`
	RawContentCommitted           bool      `json:"raw_content_committed"`
}

type maturityCorpusSummary struct {
	TargetBytes                int64  `json:"target_bytes,omitempty"`
	GeneratedBytes             int64  `json:"generated_bytes,omitempty"`
	SQLiteStorageBytes         int64  `json:"sqlite_storage_bytes"`
	Documents                  int    `json:"documents"`
	SourceDocuments            int    `json:"source_documents"`
	SynthesisDocuments         int    `json:"synthesis_documents"`
	DecisionDocuments          int    `json:"decision_documents"`
	DuplicateMarkedDocuments   int    `json:"duplicate_marked_documents"`
	StaleMarkedDocuments       int    `json:"stale_marked_documents"`
	TaggedDocuments            int    `json:"tagged_documents"`
	ProjectionStateSampleCount int    `json:"projection_state_sample_count"`
	ProvenanceEventSampleCount int    `json:"provenance_event_sample_count"`
	CountingPolicy             string `json:"counting_policy"`
}

type maturityTimingSummary struct {
	GenerateSeconds        float64 `json:"generate_seconds,omitempty"`
	ImportSyncSeconds      float64 `json:"import_sync_seconds"`
	ReopenRebuildSeconds   float64 `json:"reopen_rebuild_seconds,omitempty"`
	ListLatencySeconds     float64 `json:"list_latency_seconds,omitempty"`
	GetLatencySeconds      float64 `json:"get_latency_seconds,omitempty"`
	ProjectionCheckSeconds float64 `json:"projection_check_seconds,omitempty"`
	ProvenanceCheckSeconds float64 `json:"provenance_check_seconds,omitempty"`
	SearchTotalSeconds     float64 `json:"search_total_seconds,omitempty"`
}

type maturityReadProbe struct {
	Name            string  `json:"name"`
	QueryReference  string  `json:"query_reference,omitempty"`
	Seconds         float64 `json:"seconds"`
	ResultCount     int     `json:"result_count"`
	Status          string  `json:"status"`
	EvidencePosture string  `json:"evidence_posture"`
}

type maturityChecks struct {
	ReducedReportOnly                 bool   `json:"reduced_report_only"`
	RawLogsCommitted                  bool   `json:"raw_logs_committed"`
	RawContentCommitted               bool   `json:"raw_content_committed"`
	MachineAbsoluteArtifactRefs       bool   `json:"machine_absolute_artifact_refs"`
	RoutineAgentBypassEventsAvailable bool   `json:"routine_agent_bypass_events_available"`
	Boundary                          string `json:"boundary"`
}

type maturityOutcome struct {
	Name            string `json:"name"`
	Status          string `json:"status"`
	SafetyPass      string `json:"safety_pass"`
	CapabilityPass  string `json:"capability_pass"`
	UXQuality       string `json:"ux_quality"`
	Performance     string `json:"performance"`
	EvidencePosture string `json:"evidence_posture"`
	Details         string `json:"details"`
}

func executeMaturity(ctx context.Context, config maturityConfig, stdout io.Writer) error {
	runRoot := filepath.Clean(config.RunRoot)
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		return fmt.Errorf("create maturity run root: %w", err)
	}
	dbPath := filepath.Join(runRoot, "openclerk.sqlite")
	if err := removeSQLiteFiles(dbPath); err != nil {
		return err
	}

	vaultRoot := config.PrivateVaultRoot
	generated := generatedCorpusSummary{}
	generateSeconds := 0.0
	if config.Mode == maturityModeScaleLadder {
		vaultRoot = filepath.Join(runRoot, "scale-vault")
		if err := os.RemoveAll(vaultRoot); err != nil {
			return fmt.Errorf("reset generated scale vault: %w", err)
		}
		start := time.Now()
		var err error
		generated, err = generateScaleCorpus(vaultRoot, config)
		generateSeconds = roundSeconds(time.Since(start).Seconds())
		if err != nil {
			return err
		}
	} else if info, err := os.Stat(vaultRoot); err != nil {
		return fmt.Errorf("inspect private vault root: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("private vault root must be a directory")
	}

	runnerConfig := runclient.Config{DatabasePath: dbPath}
	if _, err := runclient.InitializePaths(runnerConfig, vaultRoot); err != nil {
		return fmt.Errorf("initialize maturity runtime paths: %w", err)
	}

	importStart := time.Now()
	client, err := runclient.Open(runnerConfig)
	importSeconds := roundSeconds(time.Since(importStart).Seconds())
	if err != nil {
		return fmt.Errorf("sync maturity vault: %w", err)
	}

	report, err := buildMaturityReport(ctx, config, client, dbPath, generated)
	if err != nil {
		_ = client.Close()
		return err
	}
	report.Timings.GenerateSeconds = generateSeconds
	report.Timings.ImportSyncSeconds = importSeconds

	if !config.SkipReopen {
		if err := client.Close(); err != nil {
			return fmt.Errorf("close maturity runtime before reopen: %w", err)
		}
		client = nil
		reopenStart := time.Now()
		reopened, err := runclient.Open(runnerConfig)
		report.Timings.ReopenRebuildSeconds = roundSeconds(time.Since(reopenStart).Seconds())
		if err != nil {
			return fmt.Errorf("reopen maturity runtime: %w", err)
		}
		if err := reopened.Close(); err != nil {
			return fmt.Errorf("close reopened maturity runtime: %w", err)
		}
	}
	if client != nil {
		if err := client.Close(); err != nil {
			return fmt.Errorf("close maturity runtime: %w", err)
		}
	}

	if err := os.MkdirAll(config.ReportDir, 0o755); err != nil {
		return fmt.Errorf("create maturity report dir: %w", err)
	}
	jsonPath := filepath.Join(config.ReportDir, config.ReportName+".json")
	markdownPath := filepath.Join(config.ReportDir, config.ReportName+".md")
	if err := writeJSON(jsonPath, report); err != nil {
		return fmt.Errorf("write maturity JSON report: %w", err)
	}
	if err := writeMaturityMarkdownReport(markdownPath, report); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "wrote %s and %s\n", filepath.ToSlash(jsonPath), filepath.ToSlash(markdownPath)); err != nil {
		return err
	}
	return nil
}

func buildMaturityReport(ctx context.Context, config maturityConfig, client *runclient.Client, dbPath string, generated generatedCorpusSummary) (maturityReport, error) {
	corpus, sampleDocID, err := collectMaturityCorpus(ctx, client, dbPath, config, generated)
	if err != nil {
		return maturityReport{}, err
	}
	probes, timings, err := runMaturityReadProbes(ctx, config, client, sampleDocID)
	if err != nil {
		return maturityReport{}, err
	}
	corpus.ProjectionStateSampleCount = countProjectionProbe(probes)
	corpus.ProvenanceEventSampleCount = countNamedProbe(probes, "provenance-sample")
	timings.SearchTotalSeconds = searchProbeSeconds(probes)

	metadata := maturityReportMetadata{
		GeneratedAt:              time.Now().UTC(),
		Lane:                     maturityLaneName(config),
		Mode:                     config.Mode,
		Harness:                  "maintainer-only OpenClerk embedded runtime maturity harness; reduced reports only",
		RunRootArtifactReference: "<run-root>",
		RawLogsCommitted:         false,
		RawContentCommitted:      false,
	}
	if config.Mode == maturityModeScaleLadder {
		metadata.Tier = config.Tier
		metadata.Seed = config.Seed
	} else {
		metadata.PrivateVaultArtifactReference = "<private-vault>"
	}
	return maturityReport{
		Metadata:   metadata,
		Corpus:     corpus,
		Timings:    timings,
		ReadProbes: probes,
		Checks: maturityChecks{
			ReducedReportOnly:                 true,
			RawLogsCommitted:                  false,
			RawContentCommitted:               false,
			MachineAbsoluteArtifactRefs:       false,
			RoutineAgentBypassEventsAvailable: false,
			Boundary:                          "This harness validates local runtime behavior and reduced-report hygiene; routine-agent bypass checks require an agent eval row with event logs.",
		},
		Outcomes: maturityOutcomes(config, corpus),
	}, nil
}

func maturityLaneName(config maturityConfig) string {
	if config.Mode == maturityModeScaleLadder {
		return "scale-ladder-validation"
	}
	return "representative-real-vault-dogfood"
}

func maturityOutcomes(config maturityConfig, corpus maturityCorpusSummary) []maturityOutcome {
	performance := "recorded"
	if config.Mode == maturityModeScaleLadder && corpus.GeneratedBytes < config.TargetBytes {
		performance = "incomplete"
	}
	capability := "pass"
	if corpus.Documents == 0 {
		capability = "fail"
	}
	return []maturityOutcome{
		{
			Name:            "reduced-report-boundary",
			Status:          "completed",
			SafetyPass:      "pass",
			CapabilityPass:  "pass",
			UXQuality:       "not_agent_ux_evidence",
			Performance:     "not_applicable",
			EvidencePosture: "repo-relative or neutral artifact references only; raw content and raw logs are not committed",
			Details:         "Report intentionally excludes document paths, titles, snippets, private vault roots, and machine-absolute run roots.",
		},
		{
			Name:            "runtime-maturity-readiness",
			Status:          "completed",
			SafetyPass:      "pass",
			CapabilityPass:  capability,
			UXQuality:       "not_agent_ux_evidence",
			Performance:     performance,
			EvidencePosture: "SQLite FTS, list/get, projection, and provenance probes executed through embedded OpenClerk runtime APIs.",
			Details:         "Use these numbers as decision input only; promotion decisions still need safety, capability, UX, performance, and evidence posture recorded separately.",
		},
	}
}

func collectMaturityCorpus(ctx context.Context, client *runclient.Client, dbPath string, config maturityConfig, generated generatedCorpusSummary) (maturityCorpusSummary, string, error) {
	corpus := maturityCorpusSummary{
		TargetBytes:        config.TargetBytes,
		GeneratedBytes:     generated.GeneratedBytes,
		SQLiteStorageBytes: sqliteStorageBytes(dbPath),
		CountingPolicy:     "Counts are derived from runner-visible document summaries and metadata; reduced reports do not include document paths, titles, snippets, or private roots.",
	}
	cursor := ""
	sampleDocID := ""
	for {
		result, err := client.ListDocuments(ctx, domain.DocumentListQuery{Limit: 100, Cursor: cursor})
		if err != nil {
			return maturityCorpusSummary{}, "", fmt.Errorf("list maturity documents: %w", err)
		}
		for _, doc := range result.Documents {
			corpus.Documents++
			if sampleDocID == "" {
				sampleDocID = doc.DocID
			}
			if strings.HasPrefix(doc.Path, "sources/") {
				corpus.SourceDocuments++
			}
			if strings.HasPrefix(doc.Path, "synthesis/") || doc.Metadata["type"] == "synthesis" {
				corpus.SynthesisDocuments++
			}
			if doc.Metadata["decision_id"] != "" || strings.HasPrefix(doc.Path, "decisions/") {
				corpus.DecisionDocuments++
			}
			status := strings.ToLower(doc.Metadata["status"])
			if status == "duplicate" || doc.Metadata["duplicates"] != "" {
				corpus.DuplicateMarkedDocuments++
			}
			if status == "archived" || status == "superseded" || status == "stale" || doc.Metadata["freshness"] == "stale" {
				corpus.StaleMarkedDocuments++
			}
			if doc.Metadata["tag"] != "" || doc.Metadata["tags"] != "" {
				corpus.TaggedDocuments++
			}
		}
		if !result.PageInfo.HasMore {
			break
		}
		cursor = result.PageInfo.NextCursor
	}
	return corpus, sampleDocID, nil
}

func runMaturityReadProbes(ctx context.Context, config maturityConfig, client *runclient.Client, sampleDocID string) ([]maturityReadProbe, maturityTimingSummary, error) {
	probes := []maturityReadProbe{}
	timings := maturityTimingSummary{}

	listStart := time.Now()
	list, err := client.ListDocuments(ctx, domain.DocumentListQuery{Limit: 50})
	timings.ListLatencySeconds = roundSeconds(time.Since(listStart).Seconds())
	if err != nil {
		return nil, maturityTimingSummary{}, fmt.Errorf("maturity list probe: %w", err)
	}
	probes = append(probes, maturityReadProbe{
		Name:            "list-documents",
		Seconds:         timings.ListLatencySeconds,
		ResultCount:     len(list.Documents),
		Status:          "completed",
		EvidencePosture: "runner-visible summaries counted without emitting paths or titles",
	})

	if sampleDocID != "" {
		getStart := time.Now()
		_, err := client.GetDocument(ctx, sampleDocID)
		timings.GetLatencySeconds = roundSeconds(time.Since(getStart).Seconds())
		if err != nil {
			return nil, maturityTimingSummary{}, fmt.Errorf("maturity get probe: %w", err)
		}
		probes = append(probes, maturityReadProbe{
			Name:            "get-document",
			Seconds:         timings.GetLatencySeconds,
			ResultCount:     1,
			Status:          "completed",
			EvidencePosture: "document body was read for timing only and is excluded from reduced reports",
		})
	}

	for i, query := range maturityQueries(config) {
		searchStart := time.Now()
		search, err := client.Search(ctx, domain.SearchQuery{Text: query, Limit: 10})
		seconds := roundSeconds(time.Since(searchStart).Seconds())
		if err != nil {
			return nil, maturityTimingSummary{}, fmt.Errorf("maturity search probe %d: %w", i+1, err)
		}
		probes = append(probes, maturityReadProbe{
			Name:            "fts-search",
			QueryReference:  maturityQueryReference(config, i, query),
			Seconds:         seconds,
			ResultCount:     len(search.Hits),
			Status:          "completed",
			EvidencePosture: "hit counts only; reduced report excludes snippets, paths, titles, doc ids, and chunk ids",
		})
	}

	projectionStart := time.Now()
	projections, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{Projection: "synthesis", Limit: 100})
	timings.ProjectionCheckSeconds = roundSeconds(time.Since(projectionStart).Seconds())
	if err != nil {
		return nil, maturityTimingSummary{}, fmt.Errorf("maturity projection probe: %w", err)
	}
	probes = append(probes, maturityReadProbe{
		Name:            "projection-synthesis-sample",
		Seconds:         timings.ProjectionCheckSeconds,
		ResultCount:     len(projections.Projections),
		Status:          "completed",
		EvidencePosture: "projection freshness count only; reduced report excludes projection refs",
	})

	provenanceStart := time.Now()
	provenance, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{Limit: 100})
	timings.ProvenanceCheckSeconds = roundSeconds(time.Since(provenanceStart).Seconds())
	if err != nil {
		return nil, maturityTimingSummary{}, fmt.Errorf("maturity provenance probe: %w", err)
	}
	probes = append(probes, maturityReadProbe{
		Name:            "provenance-sample",
		Seconds:         timings.ProvenanceCheckSeconds,
		ResultCount:     len(provenance.Events),
		Status:          "completed",
		EvidencePosture: "provenance event count only; reduced report excludes source refs and event ids",
	})

	return probes, timings, nil
}

func maturityQueryReference(config maturityConfig, index int, query string) string {
	if config.Mode == maturityModeScaleLadder {
		return query
	}
	return fmt.Sprintf("private-query-%d", index+1)
}

func countProjectionProbe(probes []maturityReadProbe) int {
	return countNamedProbe(probes, "projection-synthesis-sample")
}

func countNamedProbe(probes []maturityReadProbe, name string) int {
	for _, probe := range probes {
		if probe.Name == name {
			return probe.ResultCount
		}
	}
	return 0
}

func searchProbeSeconds(probes []maturityReadProbe) float64 {
	total := 0.0
	for _, probe := range probes {
		if probe.Name == "fts-search" {
			total += probe.Seconds
		}
	}
	return roundSeconds(total)
}

func sqliteStorageBytes(dbPath string) int64 {
	var total int64
	for _, path := range []string{dbPath, dbPath + "-wal", dbPath + "-shm"} {
		info, err := os.Stat(path)
		if err == nil {
			total += info.Size()
		}
	}
	return total
}

func removeSQLiteFiles(dbPath string) error {
	for _, path := range []string{dbPath, dbPath + "-wal", dbPath + "-shm", dbPath + ".runner-write.lock"} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove prior sqlite artifact: %w", err)
		}
	}
	return nil
}

func generateScaleCorpus(vaultRoot string, config maturityConfig) (generatedCorpusSummary, error) {
	if err := os.MkdirAll(vaultRoot, 0o755); err != nil {
		return generatedCorpusSummary{}, fmt.Errorf("create scale vault: %w", err)
	}
	docTarget := scaleDocumentTargetBytes(config.TargetBytes)
	summary := generatedCorpusSummary{}
	for index := 1; summary.GeneratedBytes < config.TargetBytes; index++ {
		doc := buildScaleDocument(index, config.Seed, docTarget)
		target := filepath.Join(vaultRoot, filepath.FromSlash(doc.path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return generatedCorpusSummary{}, fmt.Errorf("create scale doc dir: %w", err)
		}
		if err := os.WriteFile(target, []byte(doc.body), 0o644); err != nil {
			return generatedCorpusSummary{}, fmt.Errorf("write scale doc: %w", err)
		}
		summary.GeneratedDocuments++
		summary.GeneratedBytes += int64(len(doc.body))
		summary.GeneratedTaggedDocuments++
		switch doc.kind {
		case "source":
			summary.GeneratedSources++
		case "synthesis":
			summary.GeneratedSynthesis++
		case "decision":
			summary.GeneratedDecisions++
		case "duplicate":
			summary.GeneratedDuplicates++
		case "stale":
			summary.GeneratedStale++
		}
	}
	return summary, nil
}

func scaleDocumentTargetBytes(targetBytes int64) int {
	if targetBytes < 128*1024 {
		return 4096
	}
	return 128 * 1024
}

type scaleDocument struct {
	path string
	body string
	kind string
}

func buildScaleDocument(index int, seed int64, targetBytes int) scaleDocument {
	kind := scaleDocumentKind(index)
	path := scaleDocumentPath(index, kind)
	title := fmt.Sprintf("Scale Ladder %s %06d", scaleDocumentKindTitle(kind), index)
	var b strings.Builder
	writeScaleFrontmatter(&b, index, seed, kind, path, title)
	fmt.Fprintf(&b, "# %s\n\n", title)
	fmt.Fprintf(&b, "Scale ladder authority marker seed %d document %06d.\n\n", seed, index)
	if kind == "synthesis" {
		b.WriteString("## Sources\n- sources/scale/source-000001.md\n\n## Freshness\nScale ladder synthesis freshness marker checked through projection state.\n\n")
	}
	if kind == "duplicate" {
		b.WriteString("Scale ladder duplicate candidate marker: this document intentionally repeats source evidence for duplicate-pressure counts.\n\n")
	}
	words := []string{"agentops", "citation", "freshness", "provenance", "lexical", "sqlite", "synthesis", "authority", "duplicate", "stale"}
	paragraph := 0
	for b.Len() < targetBytes {
		fmt.Fprintf(
			&b,
			"Paragraph %04d records scale ladder corpus evidence for %s. The deterministic token set includes %s, %s, %s, and seed-%d-index-%d for repeatable FTS pressure without private content.\n\n",
			paragraph,
			kind,
			words[(index+paragraph)%len(words)],
			words[(index+paragraph+3)%len(words)],
			words[(index+paragraph+6)%len(words)],
			seed,
			index,
		)
		paragraph++
	}
	return scaleDocument{path: path, body: b.String(), kind: kind}
}

func scaleDocumentKindTitle(kind string) string {
	if kind == "" {
		return "Document"
	}
	return strings.ToUpper(kind[:1]) + kind[1:]
}

func scaleDocumentKind(index int) string {
	switch index % 10 {
	case 1, 6, 8:
		return "source"
	case 2, 7:
		return "synthesis"
	case 3:
		return "decision"
	case 4:
		return "duplicate"
	case 5:
		return "stale"
	default:
		return "note"
	}
}

func scaleDocumentPath(index int, kind string) string {
	switch kind {
	case "source":
		return fmt.Sprintf("sources/scale/source-%06d.md", index)
	case "synthesis":
		return fmt.Sprintf("synthesis/scale/summary-%06d.md", index)
	case "decision":
		return fmt.Sprintf("decisions/scale/decision-%06d.md", index)
	case "duplicate":
		return fmt.Sprintf("notes/scale/duplicate-%06d.md", index)
	case "stale":
		return fmt.Sprintf("archive/scale/stale-%06d.md", index)
	default:
		return fmt.Sprintf("notes/scale/note-%06d.md", index)
	}
}

func writeScaleFrontmatter(b *strings.Builder, index int, seed int64, kind string, path string, title string) {
	b.WriteString("---\n")
	fmt.Fprintf(b, "type: %s\n", scaleDocumentType(kind))
	fmt.Fprintf(b, "status: %s\n", scaleDocumentStatus(kind))
	b.WriteString("tag: scale-ladder\n")
	fmt.Fprintf(b, "scale_seed: \"%d\"\n", seed)
	if kind == "synthesis" {
		b.WriteString("freshness: fresh\n")
		b.WriteString("source_refs: sources/scale/source-000001.md\n")
	}
	if kind == "duplicate" {
		b.WriteString("duplicates: sources/scale/source-000001.md\n")
	}
	if kind == "decision" {
		fmt.Fprintf(b, "decision_id: adr-scale-ladder-%06d\n", index)
		fmt.Fprintf(b, "decision_title: %s\n", title)
		b.WriteString("decision_status: accepted\n")
		b.WriteString("decision_scope: scale-ladder\n")
		b.WriteString("decision_owner: platform\n")
	}
	fmt.Fprintf(b, "scale_path_hint: %s\n", path)
	b.WriteString("---\n")
}

func scaleDocumentType(kind string) string {
	switch kind {
	case "source":
		return "source"
	case "synthesis":
		return "synthesis"
	case "decision":
		return "decision"
	default:
		return "note"
	}
}

func scaleDocumentStatus(kind string) string {
	switch kind {
	case "duplicate":
		return "duplicate"
	case "stale":
		return "superseded"
	default:
		return "active"
	}
}

func writeMaturityMarkdownReport(path string, rep maturityReport) error {
	var b strings.Builder
	b.WriteString("# OpenClerk Maturity Report\n\n")
	fmt.Fprintf(&b, "- Lane: `%s`\n", rep.Metadata.Lane)
	fmt.Fprintf(&b, "- Mode: `%s`\n", rep.Metadata.Mode)
	if rep.Metadata.Tier != "" {
		fmt.Fprintf(&b, "- Tier: `%s`\n", rep.Metadata.Tier)
	}
	if rep.Metadata.Seed != 0 {
		fmt.Fprintf(&b, "- Seed: `%d`\n", rep.Metadata.Seed)
	}
	fmt.Fprintf(&b, "- Harness: %s\n", rep.Metadata.Harness)
	fmt.Fprintf(&b, "- Run root: `%s`\n", rep.Metadata.RunRootArtifactReference)
	if rep.Metadata.PrivateVaultArtifactReference != "" {
		fmt.Fprintf(&b, "- Private vault: `%s`\n", rep.Metadata.PrivateVaultArtifactReference)
	}
	fmt.Fprintf(&b, "- Raw logs committed: `%t`\n", rep.Metadata.RawLogsCommitted)
	fmt.Fprintf(&b, "- Raw content committed: `%t`\n\n", rep.Metadata.RawContentCommitted)

	b.WriteString("## Corpus\n\n")
	b.WriteString("| Metric | Value |\n| --- | ---: |\n")
	if rep.Corpus.TargetBytes > 0 {
		fmt.Fprintf(&b, "| target_bytes | %d |\n", rep.Corpus.TargetBytes)
	}
	if rep.Corpus.GeneratedBytes > 0 {
		fmt.Fprintf(&b, "| generated_bytes | %d |\n", rep.Corpus.GeneratedBytes)
	}
	fmt.Fprintf(&b, "| sqlite_storage_bytes | %d |\n", rep.Corpus.SQLiteStorageBytes)
	fmt.Fprintf(&b, "| documents | %d |\n", rep.Corpus.Documents)
	fmt.Fprintf(&b, "| source_documents | %d |\n", rep.Corpus.SourceDocuments)
	fmt.Fprintf(&b, "| synthesis_documents | %d |\n", rep.Corpus.SynthesisDocuments)
	fmt.Fprintf(&b, "| decision_documents | %d |\n", rep.Corpus.DecisionDocuments)
	fmt.Fprintf(&b, "| duplicate_marked_documents | %d |\n", rep.Corpus.DuplicateMarkedDocuments)
	fmt.Fprintf(&b, "| stale_marked_documents | %d |\n", rep.Corpus.StaleMarkedDocuments)
	fmt.Fprintf(&b, "| tagged_documents | %d |\n", rep.Corpus.TaggedDocuments)
	fmt.Fprintf(&b, "| projection_state_sample_count | %d |\n", rep.Corpus.ProjectionStateSampleCount)
	fmt.Fprintf(&b, "| provenance_event_sample_count | %d |\n\n", rep.Corpus.ProvenanceEventSampleCount)
	fmt.Fprintf(&b, "Counting policy: %s\n\n", rep.Corpus.CountingPolicy)

	b.WriteString("## Timings\n\n")
	b.WriteString("| Probe | Seconds |\n| --- | ---: |\n")
	fmt.Fprintf(&b, "| generate | %.2f |\n", rep.Timings.GenerateSeconds)
	fmt.Fprintf(&b, "| import_sync | %.2f |\n", rep.Timings.ImportSyncSeconds)
	fmt.Fprintf(&b, "| reopen_rebuild | %.2f |\n", rep.Timings.ReopenRebuildSeconds)
	fmt.Fprintf(&b, "| list_latency | %.2f |\n", rep.Timings.ListLatencySeconds)
	fmt.Fprintf(&b, "| get_latency | %.2f |\n", rep.Timings.GetLatencySeconds)
	fmt.Fprintf(&b, "| projection_check | %.2f |\n", rep.Timings.ProjectionCheckSeconds)
	fmt.Fprintf(&b, "| provenance_check | %.2f |\n", rep.Timings.ProvenanceCheckSeconds)
	fmt.Fprintf(&b, "| search_total | %.2f |\n\n", rep.Timings.SearchTotalSeconds)

	b.WriteString("## Read Probes\n\n")
	b.WriteString("| Probe | Query reference | Status | Results | Seconds | Evidence posture |\n| --- | --- | --- | ---: | ---: | --- |\n")
	for _, probe := range rep.ReadProbes {
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | %d | %.2f | %s |\n", probe.Name, probe.QueryReference, probe.Status, probe.ResultCount, probe.Seconds, markdownCell(probe.EvidencePosture))
	}

	b.WriteString("\n## Checks\n\n")
	b.WriteString("| Check | Value |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| reduced_report_only | `%t` |\n", rep.Checks.ReducedReportOnly)
	fmt.Fprintf(&b, "| raw_logs_committed | `%t` |\n", rep.Checks.RawLogsCommitted)
	fmt.Fprintf(&b, "| raw_content_committed | `%t` |\n", rep.Checks.RawContentCommitted)
	fmt.Fprintf(&b, "| machine_absolute_artifact_refs | `%t` |\n", rep.Checks.MachineAbsoluteArtifactRefs)
	fmt.Fprintf(&b, "| routine_agent_bypass_events_available | `%t` |\n", rep.Checks.RoutineAgentBypassEventsAvailable)
	fmt.Fprintf(&b, "| boundary | %s |\n\n", markdownCell(rep.Checks.Boundary))

	b.WriteString("## Outcomes\n\n")
	b.WriteString("| Name | Status | Safety pass | Capability pass | UX quality | Performance | Evidence posture | Details |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, outcome := range rep.Outcomes {
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` | %s | %s |\n",
			outcome.Name,
			outcome.Status,
			outcome.SafetyPass,
			outcome.CapabilityPass,
			outcome.UXQuality,
			outcome.Performance,
			markdownCell(outcome.EvidencePosture),
			markdownCell(outcome.Details),
		)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write maturity Markdown report: %w", err)
	}
	return nil
}
