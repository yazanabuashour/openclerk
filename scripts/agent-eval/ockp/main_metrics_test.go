package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMetricsFromCodexJSONLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	log := strings.Join([]string{
		`{"type":"thread.started","thread_id":"session-123"}`,
		`{"type":"item.completed","item":{"type":"agent_message","text":"done"},"usage":{"input_tokens":100,"cached_input_tokens":30,"output_tokens":12}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files /Users/example/.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files /home/runner/.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"rg --files C:\\Users\\runner\\.codex"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\",\"path_prefix\":\"notes/rag/\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\",\"metadata_key\":\"rag_scope\",\"metadata_value\":\"active-policy\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"search\",\"search\":{\"text\":\"runner\",\"tag\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"ingest_source_url\",\"source\":{\"url\":\"https://example.test/product\",\"path_hint\":\"sources/web-url/product-page-copy.md\"}}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"ingest_source_url\",\"source\":{\"url\":\"https://example.test/product\",\"path_hint\":\"sources/web-url/product-page.md\",\"mode\":\"update\"}}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"ingest_source_url\",\"source\":{\"url\":\"https://example.test/product\",\"path_hint\":\"sources/web-url/product-page.md\",\"mode\":\"update\"}}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"metadata_key\":\"tag\",\"metadata_value\":\"runner\",\"tag\":\"runner\"}}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"get_document\",\"doc_id\":\"doc_1\"}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"replace_section\",\"doc_id\":\"doc_1\",\"heading\":\"Summary\",\"content\":\"updated\"}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"append_document\",\"doc_id\":\"doc_1\",\"content\":\"updated\"}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"inspect_layout\"}' | openclerk document"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\n  \"document_links\",\"doc_id\":\"doc_1\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\n  \"graph_neighborhood\",\"doc_id\":\"doc_1\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"records_lookup\",\"records\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"record_entity\",\"entity_id\":\"runner\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"decisions_lookup\",\"decisions\":{\"text\":\"runner\"}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"decision_record\",\"decision_id\":\"adr-runner\"}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"document\",\"ref_id\":\"doc_alpha\",\"limit\":10}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"projection_states\",\"projection\":{\"limit\":10}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"printf '%s\n' '{\"action\":\"memory_router_recall_report\",\"memory_router_recall\":{\"query\":\"memory router\",\"limit\":10}}' | openclerk retrieval"}}`,
		`{"type":"tool_call","item":{"type":"tool_call","command":"/bin/zsh -lc \"printf '%s' '{\\\"action\\\":\\\"search\\\",\\\"search\\\":{\\\"text\\\":\\\"runner\\\"}}' | openclerk retrieval\""}}`,
		`not json`,
	}, "\n")
	if err := os.WriteFile(path, []byte(log), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
	parsed, err := parseMetrics(path)
	if err != nil {
		t.Fatalf("parse metrics: %v", err)
	}
	if parsed.sessionID != "session-123" || parsed.finalMessage != "done" {
		t.Fatalf("parsed = %+v", parsed)
	}
	if parsed.metrics.ToolCalls != 27 || parsed.metrics.CommandExecutions != 27 || parsed.metrics.AssistantCalls != 1 {
		t.Fatalf("metrics = %+v", parsed.metrics)
	}
	if !parsed.metrics.BroadRepoSearch {
		t.Fatalf("expected broad repo search metric")
	}
	forbiddenEvidencePaths := []string{"/Users/example", "/home/runner", `C:\Users\runner`}
	for _, evidence := range parsed.metrics.BroadRepoSearchEvidence {
		for _, forbidden := range forbiddenEvidencePaths {
			if strings.Contains(evidence, forbidden) {
				t.Fatalf("evidence was not sanitized: %v", parsed.metrics.BroadRepoSearchEvidence)
			}
		}
	}
	if parsed.metrics.NonCachedInputTokens == nil || *parsed.metrics.NonCachedInputTokens != 70 || parsed.metrics.OutputTokens == nil || *parsed.metrics.OutputTokens != 12 {
		t.Fatalf("token metrics = %+v", parsed.metrics)
	}
	if !provenanceEventRefIDsInclude(parsed.metrics.ProvenanceEventRefIDs, "doc_alpha") {
		t.Fatalf("expected provenance event ref id in %+v", parsed.metrics)
	}
	if !decisionRecordIDsInclude(parsed.metrics.DecisionRecordIDs, "adr-runner") {
		t.Fatalf("expected decision record id in %+v", parsed.metrics)
	}
	if !recordEntityIDsInclude(parsed.metrics.RecordEntityIDs, "runner") {
		t.Fatalf("expected record entity id in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.ListDocumentPathPrefixes, []string{"synthesis/"}) {
		t.Fatalf("expected list document path prefix in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.ListMetadataFilters, []string{"tag=runner"}) {
		t.Fatalf("expected list document metadata filter in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.SearchPathPrefixes, []string{"notes/rag/"}) {
		t.Fatalf("expected search path prefix in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.SearchMetadataFilters, []string{"rag_scope=active-policy"}) {
		t.Fatalf("expected search metadata filter in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.SearchTagFilters, []string{"runner"}) {
		t.Fatalf("expected search tag filter in %+v", parsed.metrics)
	}
	if !parsed.metrics.IngestSourceURLCreateUsed || !parsed.metrics.IngestSourceURLUpdateUsed || parsed.metrics.IngestSourceURLUpdateCount != 2 {
		t.Fatalf("expected source URL create and two updates in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.IngestSourceURLPathHints, []string{"sources/web-url/product-page-copy.md", "sources/web-url/product-page.md"}) {
		t.Fatalf("expected source URL path hints in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.ListTagFilters, []string{"runner"}) {
		t.Fatalf("expected list tag filter in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.GetDocumentDocIDs, []string{"doc_1"}) {
		t.Fatalf("expected get document doc id in %+v", parsed.metrics)
	}
	if !containsAllStrings(parsed.metrics.DocumentActionEvents, []string{"get_document:doc_1", "replace_section:doc_1"}) {
		t.Fatalf("expected ordered document action events in %+v", parsed.metrics)
	}
	for name, used := range map[string]bool{
		"search":                 parsed.metrics.SearchUsed,
		"search_unfiltered":      parsed.metrics.SearchUnfilteredUsed,
		"search_path_filter":     parsed.metrics.SearchPathFilterUsed,
		"search_metadata_filter": parsed.metrics.SearchMetadataFilterUsed,
		"search_tag_filter":      parsed.metrics.SearchTagFilterUsed,
		"list_documents":         parsed.metrics.ListDocumentsUsed,
		"list_metadata_filter":   parsed.metrics.ListMetadataFilterUsed,
		"list_tag_filter":        parsed.metrics.ListTagFilterUsed,
		"get_document":           parsed.metrics.GetDocumentUsed,
		"replace_section":        parsed.metrics.ReplaceSectionUsed,
		"append_document":        parsed.metrics.AppendDocumentUsed,
		"inspect_layout":         parsed.metrics.InspectLayoutUsed,
		"document_links":         parsed.metrics.DocumentLinksUsed,
		"graph_neighborhood":     parsed.metrics.GraphNeighborhoodUsed,
		"records_lookup":         parsed.metrics.RecordsLookupUsed,
		"record_entity":          parsed.metrics.RecordEntityUsed,
		"decisions_lookup":       parsed.metrics.DecisionsLookupUsed,
		"decision_record":        parsed.metrics.DecisionRecordUsed,
		"provenance_events":      parsed.metrics.ProvenanceEventsUsed,
		"projection_states":      parsed.metrics.ProjectionStatesUsed,
		"memory_router_recall":   parsed.metrics.MemoryRouterRecallReportUsed,
	} {
		if !used {
			t.Fatalf("expected %s action metric in %+v", name, parsed.metrics)
		}
	}
}

func TestAggregateMetricsRequiresAllTurnsExposeUsage(t *testing.T) {
	input := 100
	cached := 10
	nonCached := 90
	output := 20
	aggregated := aggregateMetrics([]turnResult{
		{Metrics: metrics{UsageExposed: true, InputTokens: &input, CachedInputTokens: &cached, NonCachedInputTokens: &nonCached, OutputTokens: &output, EventTypeCounts: map[string]int{"message": 1}}},
		{Metrics: metrics{EventTypeCounts: map[string]int{"tool_call": 1}}},
	})
	if aggregated.UsageExposed {
		t.Fatalf("usage should not be exposed unless all turns expose usage: %+v", aggregated)
	}
}

func TestClassifyCommandFlagsGenericNativeMediaFetches(t *testing.T) {
	for name, command := range map[string]string{
		"python_urlopen": `python -c 'import urllib.request; urllib.request.urlopen("https://video.example.test/watch?v=native-demo").read()'`,
		"node_fetch":     `node -e 'fetch("https://video.example.test/watch?v=native-demo").then(r => r.text())'`,
		"go_run":         `go run /tmp/fetch-media.go https://video.example.test/watch?v=native-demo`,
	} {
		metrics := emptyMetrics()
		classifyCommand(command, &metrics)
		if !metrics.NativeMediaAcquisition {
			t.Fatalf("%s did not set native media acquisition metric: %+v", name, metrics)
		}
		if len(metrics.NativeMediaAcquisitionEvidence) == 0 {
			t.Fatalf("%s did not record native media acquisition evidence: %+v", name, metrics)
		}
	}

	metrics := emptyMetrics()
	classifyCommand(`printf '%s\n' '{"action":"ingest_video_url","video":{"url":"https://video.example.test/watch?v=native-demo","transcript":{"text":"supplied","policy":"supplied","origin":"user_supplied_transcript"}}}' | openclerk document`, &metrics)
	if metrics.NativeMediaAcquisition {
		t.Fatalf("runner ingest_video_url command should not be classified as native acquisition: %+v", metrics)
	}
	if !metrics.IngestVideoURLUsed {
		t.Fatalf("runner ingest_video_url command should still set ingest metric: %+v", metrics)
	}
}
