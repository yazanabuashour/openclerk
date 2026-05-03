package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

func parseMetrics(eventsPath string) (parsedTurn, error) {
	file, err := os.Open(eventsPath)
	if err != nil {
		return parsedTurn{metrics: emptyMetrics()}, err
	}
	defer func() { _ = file.Close() }()
	out := parsedTurn{metrics: emptyMetrics()}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	inputTotal := 0
	cachedTotal := 0
	outputTotal := 0
	usageExposed := false
	workflowActionObserved := false
	assistantMessagesAfterWorkflowAction := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.Type != "" {
			out.metrics.EventTypeCounts[event.Type]++
		}
		if event.ThreadID != "" {
			out.sessionID = event.ThreadID
		}
		itemText := string(event.Item)
		if strings.Contains(itemText, "run document task: download source PDF") {
			out.metrics.SourcePDFDownloadFailure = true
		}
		if event.Usage != nil {
			usageExposed = true
			input, cached, output := usageNumbers(*event.Usage)
			inputTotal += input
			cachedTotal += cached
			outputTotal += output
		}
		if event.Type == "message" || strings.Contains(itemText, `"type":"message"`) || strings.Contains(itemText, `"type":"agent_message"`) {
			if strings.Contains(itemText, `"role":"assistant"`) || strings.Contains(itemText, `"type":"message"`) || strings.Contains(itemText, `"type":"agent_message"`) {
				out.metrics.AssistantCalls++
				if workflowActionObserved {
					assistantMessagesAfterWorkflowAction++
				}
				if msg := extractAssistantText(event.Item); msg != "" {
					out.finalMessage = msg
				}
			}
		}
		commands := []string{}
		if event.Type != "item.started" {
			commands = commandTexts(event.Item)
		}
		if len(commands) > 0 {
			out.metrics.ToolCalls += len(commands)
		} else if event.Type == "tool_call" || strings.Contains(itemText, `"type":"tool_call"`) || strings.Contains(itemText, `"call_id"`) {
			out.metrics.ToolCalls++
		}
		for _, command := range commands {
			out.metrics.CommandExecutions++
			classifyCommand(command, &out.metrics)
			actionText := strings.ReplaceAll(strings.ToLower(command), `\"`, `"`)
			if commandContainsWorkflowAction(actionText) {
				workflowActionObserved = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	if usageExposed {
		nonCached := inputTotal - cachedTotal
		if nonCached < 0 {
			nonCached = 0
		}
		out.metrics.UsageExposed = true
		out.metrics.InputTokens = &inputTotal
		out.metrics.CachedInputTokens = &cachedTotal
		out.metrics.NonCachedInputTokens = &nonCached
		out.metrics.OutputTokens = &outputTotal
	}
	if out.metrics.WorkflowActionCallCount > 0 && assistantMessagesAfterWorkflowAction > 1 {
		out.metrics.FinalAnswerRepairTurns = assistantMessagesAfterWorkflowAction - 1
	}
	return out, nil
}
func emptyMetrics() metrics {
	return metrics{
		EventTypeCounts:          map[string]int{},
		CommandMetricLimitations: "Command/file inspection metrics are inferred from codex exec JSON command events, not OS-level tracing.",
	}
}
func usageNumbers(value usage) (input int, cached int, output int) {
	input = value.InputTokens
	if input == 0 {
		input = value.PromptTokens
	}
	output = value.OutputTokens
	if output == 0 {
		output = value.CompletionTokens
	}
	cached = value.CachedInputTokens
	if value.InputTokensDetails != nil {
		cached += value.InputTokensDetails.CachedTokens
	}
	if value.PromptDetails != nil {
		cached += value.PromptDetails.CachedTokens
	}
	return input, cached, output
}
func extractAssistantText(raw json.RawMessage) string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	texts := []string{}
	collectTextValues(value, &texts)
	if len(texts) == 0 {
		return ""
	}
	return strings.Join(texts, "\n")
}
func collectTextValues(value any, texts *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		if role, _ := typed["role"].(string); role == "assistant" {
			if content, ok := typed["content"].(string); ok && strings.TrimSpace(content) != "" {
				*texts = append(*texts, content)
			}
		}
		if typ, _ := typed["type"].(string); typ == "agent_message" {
			if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
				*texts = append(*texts, text)
			}
		}
		if typ, _ := typed["type"].(string); typ == "output_text" || typ == "text" {
			if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
				*texts = append(*texts, text)
			}
		}
		for _, nested := range typed {
			collectTextValues(nested, texts)
		}
	case []any:
		for _, nested := range typed {
			collectTextValues(nested, texts)
		}
	}
}
func commandTexts(raw json.RawMessage) []string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	out := []string{}
	collectCommandTexts(value, &out)
	return out
}
func collectCommandTexts(value any, out *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"cmd", "command"} {
			switch command := typed[key].(type) {
			case string:
				if command != "" {
					*out = append(*out, command)
				}
			case []any:
				parts := []string{}
				for _, part := range command {
					if s, ok := part.(string); ok {
						parts = append(parts, s)
					}
				}
				if len(parts) > 0 {
					*out = append(*out, strings.Join(parts, " "))
				}
			}
		}
		for _, nested := range typed {
			collectCommandTexts(nested, out)
		}
	case []any:
		for _, nested := range typed {
			collectCommandTexts(nested, out)
		}
	}
}
func classifyCommand(command string, m *metrics) {
	lower := strings.ToLower(command)
	actionText := strings.ReplaceAll(lower, `\"`, `"`)
	workflowActionCommand := commandContainsWorkflowAction(actionText)
	primitiveCommand := commandContainsWorkflowPrimitive(actionText)
	if workflowActionCommand {
		if m.WorkflowActionFirstCommandIndex == 0 {
			m.WorkflowActionFirstCommandIndex = m.CommandExecutions
		}
		m.WorkflowActionCallCount++
	} else if primitiveCommand {
		if m.WorkflowActionFirstCommandIndex == 0 {
			m.PreActionPrimitiveCommandCount++
		} else {
			m.PostActionPrimitiveCommandCount++
		}
	}
	evidence := sanitizeMetricEvidence(command)
	addEvidence := func(target *[]string) {
		if len(*target) < 6 {
			*target = append(*target, evidence)
		}
	}
	if strings.Contains(command, "client.gen.go") || strings.Contains(command, "openapi.gen.go") || strings.Contains(command, "internal/api/openapi.gen.go") {
		m.GeneratedFileInspection = true
		addEvidence(&m.GeneratedFileEvidence)
	}
	if strings.Contains(command, "GOMODCACHE") || strings.Contains(command, "/pkg/mod") || strings.Contains(command, "go env GOMODCACHE") {
		m.ModuleCacheInspection = true
		addEvidence(&m.ModuleCacheEvidence)
	}
	if strings.Contains(command, "rg --files") || isBroadFindCommand(command) {
		m.BroadRepoSearch = true
		addEvidence(&m.BroadRepoSearchEvidence)
	}
	if strings.Contains(lower, "sqlite3") || strings.Contains(lower, "select ") || strings.Contains(lower, "pragma ") {
		m.DirectSQLiteAccess = true
		addEvidence(&m.DirectSQLiteEvidence)
	}
	if isFileInspectionCommand(lower) {
		m.FileInspectionCommands++
	}
	if strings.Contains(command, "go run ./cmd/openclerk ") || strings.Contains(command, "go run ./cmd/openclerk\n") || strings.Contains(command, " ./cmd/openclerk ") {
		m.LegacyRunnerUsage = true
		addEvidence(&m.LegacyRunnerEvidence)
	}
	if strings.Contains(command, "internal/runner") || strings.Contains(command, "cmd/openclerk") || strings.Contains(command, "scripts/agent-eval/ockp") {
		m.LegacyRunnerUsage = true
		addEvidence(&m.LegacyRunnerEvidence)
	}
	if strings.Contains(lower, "curl ") || strings.Contains(lower, "wget ") || strings.Contains(lower, "http ") || strings.Contains(lower, "httpie ") {
		m.ManualHTTPFetch = true
		addEvidence(&m.ManualHTTPFetchEvidence)
	}
	if strings.Contains(lower, "playwright") || strings.Contains(lower, "puppeteer") || strings.Contains(lower, "selenium") || strings.Contains(lower, "browser-use") {
		m.BrowserAutomation = true
		addEvidence(&m.BrowserAutomationEvidence)
	}
	if isNativeMediaAcquisitionCommand(lower) {
		m.NativeMediaAcquisition = true
		addEvidence(&m.NativeMediaAcquisitionEvidence)
	}
	m.DocumentActionEvents = append(m.DocumentActionEvents, orderedRunnerActionEvents(actionText)...)
	classifySearchCommand(actionText, m)
	if commandContainsAction(actionText, "ingest_source_url") {
		m.IngestSourceURLUsed = true
		m.IngestSourceURLPathHints = append(m.IngestSourceURLPathHints, actionFieldValues(actionText, "ingest_source_url", "path_hint")...)
		if actionHasFieldValue(actionText, "ingest_source_url", "mode", "update") {
			m.IngestSourceURLUpdateUsed = true
			m.IngestSourceURLUpdateCount++
		} else {
			m.IngestSourceURLCreateUsed = true
		}
	}
	if commandContainsAction(actionText, "ingest_video_url") {
		m.IngestVideoURLUsed = true
		if actionHasFieldValue(actionText, "ingest_video_url", "mode", "update") {
			m.IngestVideoURLUpdateUsed = true
		}
	}
	if commandContainsAction(actionText, "validate") {
		m.ValidateUsed = true
	}
	if commandContainsAction(actionText, "create_document") {
		m.CreateDocumentUsed = true
	}
	if commandContainsAction(actionText, "replace_section") {
		m.ReplaceSectionUsed = true
	}
	if commandContainsAction(actionText, "append_document") {
		m.AppendDocumentUsed = true
	}
	if commandContainsAction(actionText, "list_documents") {
		m.ListDocumentsUsed = true
		m.ListDocumentPathPrefixes = append(m.ListDocumentPathPrefixes, actionFieldValues(actionText, "list_documents", "path_prefix")...)
		classifyListDocumentsCommand(actionText, m)
	}
	if commandContainsAction(actionText, "get_document") {
		m.GetDocumentUsed = true
		m.GetDocumentDocIDs = append(m.GetDocumentDocIDs, actionFieldValues(actionText, "get_document", "doc_id")...)
	}
	if commandContainsAction(actionText, "inspect_layout") {
		m.InspectLayoutUsed = true
	}
	if commandContainsAction(actionText, "document_links") {
		m.DocumentLinksUsed = true
	}
	if commandContainsAction(actionText, "graph_neighborhood") {
		m.GraphNeighborhoodUsed = true
	}
	if commandContainsAction(actionText, "records_lookup") {
		m.RecordsLookupUsed = true
	}
	if commandContainsAction(actionText, "record_entity") {
		m.RecordEntityUsed = true
		m.RecordEntityIDs = append(m.RecordEntityIDs, actionFieldValues(actionText, "record_entity", "entity_id")...)
	}
	if commandContainsAction(actionText, "decisions_lookup") {
		m.DecisionsLookupUsed = true
	}
	if commandContainsAction(actionText, "decision_record") {
		m.DecisionRecordUsed = true
		m.DecisionRecordIDs = append(m.DecisionRecordIDs, actionFieldValues(actionText, "decision_record", "decision_id")...)
	}
	if commandContainsAction(actionText, "provenance_events") {
		m.ProvenanceEventsUsed = true
		m.ProvenanceEventRefIDs = append(m.ProvenanceEventRefIDs, actionRefIDs(actionText, "provenance_events")...)
	}
	if commandContainsAction(actionText, "projection_states") {
		m.ProjectionStatesUsed = true
	}
	if commandContainsAction(actionText, "audit_contradictions") {
		m.AuditContradictionsUsed = true
		m.AuditContradictionsModes = append(m.AuditContradictionsModes, actionFieldValues(actionText, "audit_contradictions", "mode")...)
	}
	if commandContainsAction(actionText, "memory_router_recall_report") {
		m.MemoryRouterRecallReportUsed = true
	}
	if commandContainsAction(actionText, "compile_synthesis") {
		m.CompileSynthesisUsed = true
	}
	if commandContainsAction(actionText, "source_audit_report") {
		m.SourceAuditReportUsed = true
		m.SourceAuditReportModes = append(m.SourceAuditReportModes, actionFieldValues(actionText, "source_audit_report", "mode")...)
	}
	if commandContainsAction(actionText, "evidence_bundle_report") {
		m.EvidenceBundleReportUsed = true
	}
}
func commandContainsAction(actionText string, action string) bool {
	compacted := strings.Join(strings.Fields(actionText), "")
	return strings.Contains(compacted, `"action":"`+action+`"`)
}
func commandContainsWorkflowAction(actionText string) bool {
	return commandContainsAction(actionText, "compile_synthesis") ||
		commandContainsAction(actionText, "source_audit_report") ||
		commandContainsAction(actionText, "evidence_bundle_report")
}
func commandContainsWorkflowPrimitive(actionText string) bool {
	for _, action := range []string{
		"validate",
		"ingest_source_url",
		"ingest_video_url",
		"search",
		"list_documents",
		"get_document",
		"create_document",
		"replace_section",
		"append_document",
		"inspect_layout",
		"document_links",
		"graph_neighborhood",
		"records_lookup",
		"record_entity",
		"decisions_lookup",
		"decision_record",
		"provenance_events",
		"projection_states",
		"audit_contradictions",
		"memory_router_recall_report",
	} {
		if commandContainsAction(actionText, action) {
			return true
		}
	}
	return false
}
func actionRefIDs(actionText string, action string) []string {
	return actionFieldValues(actionText, action, "ref_id")
}
func actionFieldValues(actionText string, action string, field string) []string {
	compacted := strings.Join(strings.Fields(actionText), "")
	marker := `"action":"` + action + `"`
	values := []string{}
	for _, part := range strings.Split(compacted, marker)[1:] {
		if next := strings.Index(part, `"action":"`); next >= 0 {
			part = part[:next]
		}
		fieldMarker := `"` + field + `":"`
		valueStart := strings.Index(part, fieldMarker)
		if valueStart < 0 {
			continue
		}
		valueStart += len(fieldMarker)
		valueEnd := strings.Index(part[valueStart:], `"`)
		if valueEnd < 0 {
			continue
		}
		value := strings.TrimSpace(part[valueStart : valueStart+valueEnd])
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
func actionHasFieldValue(actionText string, action string, field string, value string) bool {
	for _, got := range actionFieldValues(actionText, action, field) {
		if got == value {
			return true
		}
	}
	return false
}
func orderedRunnerActionEvents(actionText string) []string {
	compacted := strings.Join(strings.Fields(actionText), "")
	const marker = `"action":"`
	events := []string{}
	for offset := 0; offset < len(compacted); {
		index := strings.Index(compacted[offset:], marker)
		if index < 0 {
			break
		}
		actionStart := offset + index + len(marker)
		actionEndRelative := strings.Index(compacted[actionStart:], `"`)
		if actionEndRelative < 0 {
			break
		}
		actionEnd := actionStart + actionEndRelative
		action := compacted[actionStart:actionEnd]
		segmentEnd := len(compacted)
		if next := strings.Index(compacted[actionEnd:], marker); next >= 0 {
			segmentEnd = actionEnd + next
		}
		if event := runnerActionEvent(action, compacted[actionEnd:segmentEnd]); event != "" {
			events = append(events, event)
		}
		offset = segmentEnd
	}
	return events
}
func runnerActionEvent(action string, segment string) string {
	switch action {
	case "search", "inspect_layout":
		return action
	case "list_documents":
		return actionWithOptionalValue(action, fieldValueFromCompactedAction(segment, "path_prefix"))
	case "get_document", "replace_section", "append_document", "create_document":
		return actionWithOptionalValue(action, fieldValueFromCompactedAction(segment, "doc_id"))
	case "provenance_events", "projection_states":
		return actionWithOptionalValue(action, fieldValueFromCompactedAction(segment, "ref_id"))
	default:
		return ""
	}
}
func actionWithOptionalValue(action string, value string) string {
	if value == "" {
		return action
	}
	return action + ":" + value
}
func classifySearchCommand(actionText string, m *metrics) {
	compacted := strings.Join(strings.Fields(actionText), "")
	const marker = `"action":"search"`
	if !strings.Contains(compacted, marker) {
		return
	}
	m.SearchUsed = true
	parts := strings.Split(compacted, marker)
	for _, part := range parts[1:] {
		if next := strings.Index(part, `"action":"`); next >= 0 {
			part = part[:next]
		}
		hasPathFilter := strings.Contains(part, `"path_prefix":`)
		hasMetadataFilter := strings.Contains(part, `"metadata_key":`) || strings.Contains(part, `"metadata_value":`)
		hasTagFilter := strings.Contains(part, `"tag":`)
		if hasPathFilter {
			m.SearchPathFilterUsed = true
			m.SearchPathPrefixes = append(m.SearchPathPrefixes, fieldValueFromCompactedAction(part, "path_prefix"))
		}
		if hasMetadataFilter {
			m.SearchMetadataFilterUsed = true
			key := fieldValueFromCompactedAction(part, "metadata_key")
			value := fieldValueFromCompactedAction(part, "metadata_value")
			if key != "" || value != "" {
				m.SearchMetadataFilters = append(m.SearchMetadataFilters, key+"="+value)
			}
		}
		if hasTagFilter {
			m.SearchTagFilterUsed = true
			if value := fieldValueFromCompactedAction(part, "tag"); value != "" {
				m.SearchTagFilters = append(m.SearchTagFilters, value)
			}
		}
		if !hasPathFilter && !hasMetadataFilter && !hasTagFilter {
			m.SearchUnfilteredUsed = true
		}
	}
}
func classifyListDocumentsCommand(actionText string, m *metrics) {
	compacted := strings.Join(strings.Fields(actionText), "")
	const marker = `"action":"list_documents"`
	if !strings.Contains(compacted, marker) {
		return
	}
	parts := strings.Split(compacted, marker)
	for _, part := range parts[1:] {
		if next := strings.Index(part, `"action":"`); next >= 0 {
			part = part[:next]
		}
		hasMetadataFilter := strings.Contains(part, `"metadata_key":`) || strings.Contains(part, `"metadata_value":`)
		hasTagFilter := strings.Contains(part, `"tag":`)
		if hasMetadataFilter {
			m.ListMetadataFilterUsed = true
			key := fieldValueFromCompactedAction(part, "metadata_key")
			value := fieldValueFromCompactedAction(part, "metadata_value")
			if key != "" || value != "" {
				m.ListMetadataFilters = append(m.ListMetadataFilters, key+"="+value)
			}
		}
		if hasTagFilter {
			m.ListTagFilterUsed = true
			if value := fieldValueFromCompactedAction(part, "tag"); value != "" {
				m.ListTagFilters = append(m.ListTagFilters, value)
			}
		}
	}
}
func fieldValueFromCompactedAction(part string, field string) string {
	fieldMarker := `"` + field + `":"`
	valueStart := strings.Index(part, fieldMarker)
	if valueStart < 0 {
		return ""
	}
	valueStart += len(fieldMarker)
	valueEnd := strings.Index(part[valueStart:], `"`)
	if valueEnd < 0 {
		return ""
	}
	return strings.TrimSpace(part[valueStart : valueStart+valueEnd])
}
func sanitizeMetricEvidence(value string) string {
	replacements := []string{}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		replacements = append(replacements, home, "<home>")
	}
	if tmp := strings.TrimSpace(os.TempDir()); tmp != "" {
		replacements = append(replacements, tmp, "<tmp>")
	}
	if len(replacements) == 0 {
		return sanitizeKnownHomePrefixes(value)
	}
	return sanitizeKnownHomePrefixes(strings.NewReplacer(replacements...).Replace(value))
}
func sanitizeKnownHomePrefixes(value string) string {
	value = unixHomePathPattern.ReplaceAllString(value, "<home>")
	return windowsHomePathPattern.ReplaceAllString(value, "<home>")
}
func isFileInspectionCommand(command string) bool {
	for _, prefix := range []string{"cat ", "sed ", "nl ", "head ", "tail ", "less ", "grep ", "rg "} {
		if strings.HasPrefix(strings.TrimSpace(command), prefix) {
			return true
		}
	}
	return false
}
func isNativeMediaAcquisitionCommand(command string) bool {
	if containsAny(command, []string{"yt-dlp", "youtube-dl", "ffmpeg", "whisper", "transcript api", "transcript-api", "youtube-transcript", "gemini"}) {
		return true
	}
	if !containsLikelyNativeMediaTarget(command) {
		return false
	}
	return containsAny(command, []string{
		"python -c", "python3 -c", "node -e", "deno eval", "go run ", "ruby -e", "perl -e", "php -r",
		"urllib.request", "urlopen", "requests.get", "requests.post", "httpx.", "fetch(", "axios.", "http.get", "https.get", "net/http",
	})
}
func containsLikelyNativeMediaTarget(command string) bool {
	return containsAny(command, []string{
		"video.example.test",
		"youtube.example.test",
		"youtube.com/watch",
		"youtu.be/",
		".mp4",
		".mp3",
		".m4a",
		".wav",
		".mov",
		".webm",
		".mkv",
		".aac",
		".flac",
	})
}
func isBroadFindCommand(command string) bool {
	trimmed := strings.TrimSpace(command)
	if !strings.Contains(trimmed, "find .") && !strings.Contains(trimmed, "find ..") {
		return false
	}
	if strings.Contains(trimmed, "-type d") && !strings.Contains(trimmed, "-type f") {
		return false
	}
	return true
}
func aggregateMetrics(turns []turnResult) metrics {
	out := emptyMetrics()
	allUsageExposed := len(turns) > 0
	inputTotal := 0
	cachedTotal := 0
	nonCachedTotal := 0
	outputTotal := 0
	for _, turn := range turns {
		current := turn.Metrics
		commandOffset := out.CommandExecutions
		out.AssistantCalls += current.AssistantCalls
		out.ToolCalls += current.ToolCalls
		out.CommandExecutions += current.CommandExecutions
		out.FileInspectionCommands += current.FileInspectionCommands
		out.GeneratedFileInspection = out.GeneratedFileInspection || current.GeneratedFileInspection
		out.ModuleCacheInspection = out.ModuleCacheInspection || current.ModuleCacheInspection
		out.BroadRepoSearch = out.BroadRepoSearch || current.BroadRepoSearch
		out.DirectSQLiteAccess = out.DirectSQLiteAccess || current.DirectSQLiteAccess
		out.LegacyRunnerUsage = out.LegacyRunnerUsage || current.LegacyRunnerUsage
		out.ManualHTTPFetch = out.ManualHTTPFetch || current.ManualHTTPFetch
		out.BrowserAutomation = out.BrowserAutomation || current.BrowserAutomation
		out.NativeMediaAcquisition = out.NativeMediaAcquisition || current.NativeMediaAcquisition
		out.SearchUsed = out.SearchUsed || current.SearchUsed
		out.SearchUnfilteredUsed = out.SearchUnfilteredUsed || current.SearchUnfilteredUsed
		out.SearchPathFilterUsed = out.SearchPathFilterUsed || current.SearchPathFilterUsed
		out.SearchPathPrefixes = append(out.SearchPathPrefixes, current.SearchPathPrefixes...)
		out.SearchMetadataFilterUsed = out.SearchMetadataFilterUsed || current.SearchMetadataFilterUsed
		out.SearchMetadataFilters = append(out.SearchMetadataFilters, current.SearchMetadataFilters...)
		out.SearchTagFilterUsed = out.SearchTagFilterUsed || current.SearchTagFilterUsed
		out.SearchTagFilters = append(out.SearchTagFilters, current.SearchTagFilters...)
		out.IngestSourceURLUsed = out.IngestSourceURLUsed || current.IngestSourceURLUsed
		out.IngestSourceURLCreateUsed = out.IngestSourceURLCreateUsed || current.IngestSourceURLCreateUsed
		out.IngestSourceURLUpdateUsed = out.IngestSourceURLUpdateUsed || current.IngestSourceURLUpdateUsed
		out.IngestSourceURLUpdateCount += current.IngestSourceURLUpdateCount
		out.IngestSourceURLPathHints = append(out.IngestSourceURLPathHints, current.IngestSourceURLPathHints...)
		out.IngestVideoURLUsed = out.IngestVideoURLUsed || current.IngestVideoURLUsed
		out.IngestVideoURLUpdateUsed = out.IngestVideoURLUpdateUsed || current.IngestVideoURLUpdateUsed
		out.SourcePDFDownloadFailure = out.SourcePDFDownloadFailure || current.SourcePDFDownloadFailure
		out.ValidateUsed = out.ValidateUsed || current.ValidateUsed
		out.CreateDocumentUsed = out.CreateDocumentUsed || current.CreateDocumentUsed
		out.ReplaceSectionUsed = out.ReplaceSectionUsed || current.ReplaceSectionUsed
		out.AppendDocumentUsed = out.AppendDocumentUsed || current.AppendDocumentUsed
		out.ListDocumentsUsed = out.ListDocumentsUsed || current.ListDocumentsUsed
		out.ListDocumentPathPrefixes = append(out.ListDocumentPathPrefixes, current.ListDocumentPathPrefixes...)
		out.ListMetadataFilterUsed = out.ListMetadataFilterUsed || current.ListMetadataFilterUsed
		out.ListMetadataFilters = append(out.ListMetadataFilters, current.ListMetadataFilters...)
		out.ListTagFilterUsed = out.ListTagFilterUsed || current.ListTagFilterUsed
		out.ListTagFilters = append(out.ListTagFilters, current.ListTagFilters...)
		out.GetDocumentUsed = out.GetDocumentUsed || current.GetDocumentUsed
		out.GetDocumentDocIDs = append(out.GetDocumentDocIDs, current.GetDocumentDocIDs...)
		out.DocumentActionEvents = append(out.DocumentActionEvents, current.DocumentActionEvents...)
		out.InspectLayoutUsed = out.InspectLayoutUsed || current.InspectLayoutUsed
		out.DocumentLinksUsed = out.DocumentLinksUsed || current.DocumentLinksUsed
		out.GraphNeighborhoodUsed = out.GraphNeighborhoodUsed || current.GraphNeighborhoodUsed
		out.RecordsLookupUsed = out.RecordsLookupUsed || current.RecordsLookupUsed
		out.RecordEntityUsed = out.RecordEntityUsed || current.RecordEntityUsed
		out.RecordEntityIDs = append(out.RecordEntityIDs, current.RecordEntityIDs...)
		out.DecisionsLookupUsed = out.DecisionsLookupUsed || current.DecisionsLookupUsed
		out.DecisionRecordUsed = out.DecisionRecordUsed || current.DecisionRecordUsed
		out.DecisionRecordIDs = append(out.DecisionRecordIDs, current.DecisionRecordIDs...)
		out.ProvenanceEventsUsed = out.ProvenanceEventsUsed || current.ProvenanceEventsUsed
		out.ProvenanceEventRefIDs = append(out.ProvenanceEventRefIDs, current.ProvenanceEventRefIDs...)
		out.ProjectionStatesUsed = out.ProjectionStatesUsed || current.ProjectionStatesUsed
		out.AuditContradictionsUsed = out.AuditContradictionsUsed || current.AuditContradictionsUsed
		out.AuditContradictionsModes = append(out.AuditContradictionsModes, current.AuditContradictionsModes...)
		out.MemoryRouterRecallReportUsed = out.MemoryRouterRecallReportUsed || current.MemoryRouterRecallReportUsed
		out.CompileSynthesisUsed = out.CompileSynthesisUsed || current.CompileSynthesisUsed
		out.SourceAuditReportUsed = out.SourceAuditReportUsed || current.SourceAuditReportUsed
		out.SourceAuditReportModes = append(out.SourceAuditReportModes, current.SourceAuditReportModes...)
		out.EvidenceBundleReportUsed = out.EvidenceBundleReportUsed || current.EvidenceBundleReportUsed
		if current.WorkflowActionFirstCommandIndex != 0 {
			first := commandOffset + current.WorkflowActionFirstCommandIndex
			if out.WorkflowActionFirstCommandIndex == 0 || first < out.WorkflowActionFirstCommandIndex {
				out.WorkflowActionFirstCommandIndex = first
			}
		}
		out.WorkflowActionCallCount += current.WorkflowActionCallCount
		out.PreActionPrimitiveCommandCount += current.PreActionPrimitiveCommandCount
		out.PostActionPrimitiveCommandCount += current.PostActionPrimitiveCommandCount
		out.FinalAnswerRepairTurns += current.FinalAnswerRepairTurns
		out.GeneratedFileEvidence = append(out.GeneratedFileEvidence, current.GeneratedFileEvidence...)
		out.ModuleCacheEvidence = append(out.ModuleCacheEvidence, current.ModuleCacheEvidence...)
		out.BroadRepoSearchEvidence = append(out.BroadRepoSearchEvidence, current.BroadRepoSearchEvidence...)
		out.DirectSQLiteEvidence = append(out.DirectSQLiteEvidence, current.DirectSQLiteEvidence...)
		out.LegacyRunnerEvidence = append(out.LegacyRunnerEvidence, current.LegacyRunnerEvidence...)
		out.ManualHTTPFetchEvidence = append(out.ManualHTTPFetchEvidence, current.ManualHTTPFetchEvidence...)
		out.BrowserAutomationEvidence = append(out.BrowserAutomationEvidence, current.BrowserAutomationEvidence...)
		out.NativeMediaAcquisitionEvidence = append(out.NativeMediaAcquisitionEvidence, current.NativeMediaAcquisitionEvidence...)
		for eventType, count := range current.EventTypeCounts {
			out.EventTypeCounts[eventType] += count
		}
		if !current.UsageExposed || current.InputTokens == nil || current.CachedInputTokens == nil || current.NonCachedInputTokens == nil || current.OutputTokens == nil {
			allUsageExposed = false
			continue
		}
		inputTotal += *current.InputTokens
		cachedTotal += *current.CachedInputTokens
		nonCachedTotal += *current.NonCachedInputTokens
		outputTotal += *current.OutputTokens
	}
	if allUsageExposed {
		out.UsageExposed = true
		out.InputTokens = &inputTotal
		out.CachedInputTokens = &cachedTotal
		out.NonCachedInputTokens = &nonCachedTotal
		out.OutputTokens = &outputTotal
	}
	return out
}
