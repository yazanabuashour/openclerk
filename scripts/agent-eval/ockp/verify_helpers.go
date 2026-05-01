package main

import (
	"context"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func verifyDocumentContains(ctx context.Context, paths evalPaths, docPath string, required []string, forbidden []string) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	if !found {
		return verificationResult{Passed: false, DatabasePass: false, Details: "missing " + docPath}, nil
	}
	failures := missingRequired(body, required)
	failures = append(failures, presentForbidden(body, forbidden)...)
	return verificationResult{
		Passed:        len(failures) == 0,
		DatabasePass:  len(failures) == 0,
		AssistantPass: true,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}
func documentByPath(ctx context.Context, paths evalPaths, docPath string) (*runner.Document, bool, error) {
	docID, found, err := documentIDByPath(ctx, paths, docPath)
	if err != nil || !found {
		return nil, found, err
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionGet, DocID: docID})
	if err != nil {
		return nil, false, err
	}
	if got.Document != nil {
		return got.Document, true, nil
	}
	return nil, false, nil
}
func documentBodyByPath(ctx context.Context, paths evalPaths, docPath string) (string, bool, error) {
	doc, found, err := documentByPath(ctx, paths, docPath)
	if err != nil || !found {
		return "", found, err
	}
	if doc != nil {
		return doc.Body, true, nil
	}
	return "", false, nil
}
func documentContaining(ctx context.Context, paths evalPaths, needle string) (string, string, bool, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{Limit: 100},
	})
	if err != nil {
		return "", "", false, err
	}
	for _, entry := range list.Documents {
		doc, found, err := documentByPath(ctx, paths, entry.Path)
		if err != nil {
			return "", "", false, err
		}
		if !found || doc == nil {
			continue
		}
		if strings.Contains(doc.Body, needle) {
			return entry.Path, doc.Body, true, nil
		}
	}
	return "", "", false, nil
}
func documentIDByPath(ctx context.Context, paths evalPaths, docPath string) (string, bool, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 100},
	})
	if err != nil {
		return "", false, err
	}
	for _, doc := range list.Documents {
		if doc.Path == docPath {
			return doc.DocID, true, nil
		}
	}
	return "", false, nil
}
func exactDocumentCount(ctx context.Context, paths evalPaths, docPath string) (int, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 100},
	})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, doc := range list.Documents {
		if doc.Path == docPath {
			count++
		}
	}
	return count, nil
}
func documentCountWithPrefix(ctx context.Context, paths evalPaths, pathPrefix string) (int, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: pathPrefix, Limit: 100},
	})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, doc := range list.Documents {
		if strings.HasPrefix(doc.Path, pathPrefix) {
			count++
		}
	}
	return count, nil
}
func disallowedDocumentPathsWithPrefix(ctx context.Context, paths evalPaths, pathPrefix string, allowed map[string]bool) ([]string, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: pathPrefix, Limit: 100},
	})
	if err != nil {
		return nil, err
	}
	disallowed := []string{}
	for _, doc := range list.Documents {
		if strings.HasPrefix(doc.Path, pathPrefix) && !allowed[doc.Path] {
			disallowed = append(disallowed, doc.Path)
		}
	}
	sort.Strings(disallowed)
	return disallowed, nil
}
func firstSynthesisProjection(ctx context.Context, paths evalPaths, docID string) (*runner.ProjectionState, error) {
	if strings.TrimSpace(docID) == "" {
		return nil, nil
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      docID,
			Limit:      5,
		},
	})
	if err != nil {
		return nil, err
	}
	if projections.Projections == nil || len(projections.Projections.Projections) == 0 {
		return nil, nil
	}
	projection := projections.Projections.Projections[0]
	return &projection, nil
}
func projectionDetailContains(details map[string]string, key string, value string) bool {
	return strings.Contains(details[key], value)
}
func topSearchHit(result runner.RetrievalTaskResult) (runner.SearchHit, bool) {
	if result.Search == nil || len(result.Search.Hits) == 0 {
		return runner.SearchHit{}, false
	}
	return result.Search.Hits[0], true
}
func searchContainsPath(result runner.RetrievalTaskResult, path string) bool {
	if result.Search == nil {
		return false
	}
	for _, hit := range result.Search.Hits {
		if searchHitPath(hit) == path {
			return true
		}
	}
	return false
}
func searchResultHasCitations(result runner.RetrievalTaskResult) bool {
	if result.Search == nil || len(result.Search.Hits) == 0 {
		return false
	}
	for _, hit := range result.Search.Hits {
		if searchHitHasCitation(hit) {
			return true
		}
	}
	return false
}
func searchOnlyContainsPath(result runner.RetrievalTaskResult, path string) bool {
	if result.Search == nil || len(result.Search.Hits) == 0 {
		return false
	}
	for _, hit := range result.Search.Hits {
		if searchHitPath(hit) != path {
			return false
		}
	}
	return true
}
func searchHitPath(hit runner.SearchHit) string {
	if len(hit.Citations) > 0 {
		return hit.Citations[0].Path
	}
	return ""
}
func searchHitHasCitation(hit runner.SearchHit) bool {
	if hit.DocID == "" || hit.ChunkID == "" {
		return false
	}
	for _, citation := range hit.Citations {
		if citation.DocID != "" &&
			citation.ChunkID != "" &&
			citation.Path != "" &&
			citation.LineStart > 0 &&
			citation.LineEnd >= citation.LineStart {
			return true
		}
	}
	return false
}
func allPathsFound(found map[string]bool, expected []string) bool {
	for _, path := range expected {
		if !found[path] {
			return false
		}
	}
	return true
}
func missingRequired(body string, required []string) []string {
	failures := []string{}
	for _, value := range required {
		if !strings.Contains(body, value) {
			failures = append(failures, "missing "+value)
		}
	}
	return failures
}
func missingRequiredFold(body string, required []string) []string {
	failures := []string{}
	lowerBody := strings.ToLower(body)
	for _, value := range required {
		if !strings.Contains(lowerBody, strings.ToLower(value)) {
			failures = append(failures, "missing "+value)
		}
	}
	return failures
}
func sourceRefsFrontmatterFailures(body string, expected []string) []string {
	value, found, singleLine := sourceRefsFrontmatterValue(body)
	if !found {
		return []string{"missing source_refs frontmatter"}
	}
	if !singleLine {
		return []string{"source_refs must be single-line comma-separated frontmatter"}
	}
	refs := map[string]bool{}
	for _, ref := range strings.Split(value, ",") {
		normalized := strings.Trim(strings.TrimSpace(ref), `"'`)
		if normalized != "" {
			refs[normalized] = true
		}
	}
	failures := []string{}
	for _, ref := range expected {
		if !refs[ref] {
			failures = append(failures, "missing source ref "+ref)
		}
	}
	return failures
}
func sourceRefsFrontmatterValue(body string) (string, bool, bool) {
	lines := strings.Split(body, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", false, false
	}
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "source_refs") {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || strings.HasPrefix(value, "[") || strings.HasSuffix(value, "]") {
			return value, true, false
		}
		return value, true, true
	}
	return "", false, false
}
func presentForbidden(body string, forbidden []string) []string {
	failures := []string{}
	for _, value := range forbidden {
		if strings.Contains(body, value) {
			failures = append(failures, "unexpected "+value)
		}
	}
	return failures
}
func messageContainsAll(message string, values []string) bool {
	lower := normalizeValidationMessage(message)
	for _, value := range values {
		if !strings.Contains(lower, strings.ToLower(value)) {
			return false
		}
	}
	return true
}
func messageContainsAny(message string, values []string) bool {
	return containsAny(normalizeValidationMessage(message), lowerStrings(values))
}
func graphSemanticsReferenceAnswerPass(message string) bool {
	normalized := normalizeValidationMessage(message)
	if messagePromotesGraphSemantics(normalized) {
		return false
	}
	return containsAny(normalized, []string{"search"}) &&
		containsAny(normalized, []string{"document_links", "links", "link"}) &&
		containsAny(normalized, []string{"backlink", "incoming"}) &&
		containsAny(normalized, []string{"graph_neighborhood", "graph neighborhood"}) &&
		containsAny(normalized, []string{"markdown", "relationship text", "relationship wording"}) &&
		containsAny(normalized, []string{"citation", "cited", "source", "canonical", "derived"}) &&
		containsAny(normalized, []string{"projection", "fresh", "freshness"}) &&
		containsAny(normalized, []string{"reference", "defer", "deferred", "not promote", "do not promote", "not promoted", "keep"})
}
func graphSemanticsRevisitAnswerPass(message string, scripted bool) bool {
	normalized := normalizeValidationMessage(message)
	if !graphSemanticsReferenceAnswerPass(message) {
		return false
	}
	safeCurrentPrimitives := containsAny(normalized, []string{"current primitives can express", "current document/retrieval primitives can express", "existing primitives can express", "current primitives were sufficient", "sufficient to repair", "workflow safely", "express the workflow safely"})
	uxPosture := containsAny(normalized, []string{"ux", "user experience"}) &&
		containsAny(normalized, []string{"acceptable", "not acceptable", "unacceptable"})
	hasGapClassification := containsAny(normalized, []string{"capability gap", "ergonomics gap", "neither"})
	if !hasGapClassification && (!scripted || !safeCurrentPrimitives || !uxPosture) {
		return false
	}
	if !scripted {
		return true
	}
	return safeCurrentPrimitives && uxPosture
}
func broadContradictionAuditAnswerPass(message string, scripted bool) bool {
	normalized := normalizeValidationMessage(message)
	requiredEvidence := containsAny(normalized, []string{"search", "audit_contradictions"}) &&
		containsAny(normalized, []string{"citation", "cited", "current source", "current_source_paths", "superseded source", "superseded_source_paths", "source path", "source paths", "source docs", "source refs", "source_refs"}) &&
		containsAny(normalized, []string{"provenance"}) &&
		containsAny(normalized, []string{"projection"}) &&
		containsAny(normalized, []string{"fresh", "freshness"}) &&
		containsAny(normalized, []string{"synthesis/audit-runner-routing.md"}) &&
		containsAny(normalized, []string{"sources/audit-conflict-alpha.md"}) &&
		containsAny(normalized, []string{"sources/audit-conflict-bravo.md"}) &&
		containsAny(normalized, []string{"unresolved", "no supersession", "no source authority", "cannot choose", "do not choose"}) &&
		containsAny(normalized, []string{"reference", "defer", "deferred", "not promote", "do not promote", "not promoted", "keep"})
	if !requiredEvidence {
		return false
	}
	capabilityPosture := containsAny(normalized, []string{"capability gap", "capability_gap", "neither"}) ||
		containsAny(normalized, []string{"current primitives can express", "current document/retrieval primitives can express", "existing primitives can express", "workflow safely", "express the workflow safely"})
	ergonomicsPosture := containsAny(normalized, []string{"ergonomics gap", "ergonomics_gap", "neither"}) ||
		containsAny(normalized, []string{"ux", "user experience"})
	if !capabilityPosture || !ergonomicsPosture {
		return false
	}
	if !scripted {
		return true
	}
	safeCurrentPrimitives := containsAny(normalized, []string{"current primitives can express", "current document/retrieval primitives can express", "existing primitives can express", "workflow safely", "express the workflow safely"})
	uxPosture := containsAny(normalized, []string{"ux", "user experience"}) &&
		containsAny(normalized, []string{"acceptable", "not acceptable", "unacceptable"})
	return safeCurrentPrimitives && uxPosture
}
func messagePromotesGraphSemantics(normalized string) bool {
	promotionPhrases := []string{
		"decision: promote",
		"promote graph semantics",
		"promote richer graph",
		"promote semantic graph",
		"add semantic graph",
		"new graph authority",
		"independent semantic",
		"promote a semantic-label graph layer",
		"promote semantic-label graph layer",
		"semantic-label graph layer should be promoted",
	}
	for _, phrase := range promotionPhrases {
		if strings.Contains(normalized, phrase) &&
			!strings.Contains(normalized, "do not "+phrase) &&
			!strings.Contains(normalized, "not "+phrase) &&
			!strings.Contains(normalized, "not an "+phrase) &&
			!strings.Contains(normalized, "not a "+phrase) &&
			!strings.Contains(normalized, "rather than "+phrase) &&
			!strings.Contains(normalized, "instead of "+phrase) {
			return true
		}
	}
	return false
}
func memoryRouterReferenceAnswerPass(message string) bool {
	normalized := normalizeValidationMessage(message)
	if messagePromotesMemoryRouter(normalized) {
		return false
	}
	return containsAny(normalized, []string{"temporal", "current", "stale", "effective"}) &&
		containsAny(normalized, []string{"session promotion", "session-derived", "session observation", "promotion path", "durable markdown", "canonical markdown", "canonicalization"}) &&
		containsAny(normalized, []string{"feedback", "weight", "weighted", "advisory"}) &&
		containsAny(normalized, []string{"routing", "route", "router"}) &&
		containsAny(normalized, []string{"source_refs", "source ref", "source refs", "source evidence", "citation", "citations", "cited", "source path", "source paths"}) &&
		containsAny(normalized, []string{"freshness", "fresh", "provenance", "projection"}) &&
		containsAny(normalized, []string{"reference", "defer", "deferred", "not promote", "do not promote", "not promoted", "keep"})
}
func memoryRouterRevisitAnswerPass(message string, scripted bool) bool {
	normalized := normalizeValidationMessage(message)
	if !memoryRouterReferenceAnswerPass(message) {
		return false
	}
	capabilityPosture := containsAny(normalized, []string{"capability gap", "capability_gap"}) ||
		containsAny(normalized, []string{"express the workflow safely", "can express this workflow safely", "can express safely", "current primitives can express", "current primitives express"})
	ergonomicsPosture := containsAny(normalized, []string{"ergonomics gap", "ergonomics_gap", "ux", "user experience"}) ||
		containsAny(normalized, []string{"current ux is acceptable", "ux is acceptable", "acceptable enough", "current workflow is acceptable"})
	required := capabilityPosture &&
		ergonomicsPosture &&
		containsAny(normalized, []string{"search"}) &&
		containsAny(normalized, []string{"current primitives", "current workflow", "document and retrieval", "document/retrieval", "existing agentops document and retrieval", "existing document/retrieval", "existing runner actions"})
	if scripted {
		required = required &&
			containsAny(normalized, []string{"express", "safely express", "can express", "workflow safely"}) &&
			containsAny(normalized, []string{"acceptable", "ux acceptable", "current ux"})
	}
	return required
}
func promotedRecordDomainAnswerPass(message string, scripted bool) bool {
	normalized := normalizeValidationMessage(message)
	if messagePromotesRecordDomain(normalized) {
		return false
	}
	requiredEvidence := containsAny(normalized, []string{"search"}) &&
		containsAny(normalized, []string{"list_documents", "list documents"}) &&
		containsAny(normalized, []string{"get_document", "get document"}) &&
		containsAny(normalized, []string{"records_lookup", "records lookup", "generic records"}) &&
		containsAny(normalized, []string{"record_entity", "record entity"}) &&
		containsAny(normalized, []string{"provenance"}) &&
		containsAny(normalized, []string{"projection", "freshness", "fresh"}) &&
		containsAny(normalized, []string{"citation", "citations", "cited", "source"}) &&
		containsAny(normalized, []string{"local-first", "no-bypass", "bypass boundaries", "no bypass"}) &&
		containsAny(normalized, []string{"reference", "defer", "deferred", "not promote", "do not promote", "not promoted", "keep"})
	if !requiredEvidence {
		return false
	}
	capabilityPosture := containsAny(normalized, []string{"capability gap", "capability_gap", "neither"}) ||
		containsAny(normalized, []string{"current primitives can express", "existing primitives can express", "can express the workflow safely", "express the workflow safely"})
	ergonomicsPosture := containsAny(normalized, []string{"ergonomics gap", "ergonomics_gap", "neither", "ux", "user experience"}) ||
		containsAny(normalized, []string{"current ux is acceptable", "ux is acceptable", "acceptable enough", "current workflow is acceptable"})
	if !capabilityPosture || !ergonomicsPosture {
		return false
	}
	if !scripted {
		return true
	}
	return containsAny(normalized, []string{"current primitives", "existing primitives", "document and retrieval", "document/retrieval", "existing runner actions"}) &&
		containsAny(normalized, []string{"express", "safely express", "can express", "workflow safely"}) &&
		containsAny(normalized, []string{"acceptable", "ux acceptable", "current ux"})
}
func relationshipRecordCeremonyAnswerPass(message string, scripted bool) bool {
	normalized := normalizeValidationMessage(message)
	if !graphSemanticsRevisitAnswerPass(message, scripted) || !promotedRecordDomainAnswerPass(message, scripted) {
		return false
	}
	requiredEvidence := containsAny(normalized, []string{"relationship", "relationship-shaped", "markdown relationship"}) &&
		containsAny(normalized, []string{"record", "promoted-record", "records_lookup"}) &&
		containsAny(normalized, []string{"graph_neighborhood", "graph neighborhood"}) &&
		containsAny(normalized, []string{"record_entity", "record entity"}) &&
		containsAny(normalized, []string{"graph projection", "graph freshness"}) &&
		containsAny(normalized, []string{"records projection", "records freshness"}) &&
		containsAny(normalized, []string{"combined", "relationship/record", "relationship and record"}) &&
		containsAny(normalized, []string{"reference", "defer", "deferred", "not promote", "do not promote", "not promoted", "keep"})
	if !requiredEvidence {
		return false
	}
	if !scripted {
		return true
	}
	return containsAny(normalized, []string{"current primitives can express", "current document/retrieval primitives can express", "existing primitives can express", "combined workflow safely", "express the combined workflow safely"}) &&
		containsAny(normalized, []string{"ux", "user experience"}) &&
		containsAny(normalized, []string{"acceptable", "not acceptable", "unacceptable"})
}
func relationshipRecordCandidateAnswerPass(message string, scripted bool) bool {
	normalized := normalizeValidationMessage(message)
	if messagePromotesGraphSemantics(normalized) || messagePromotesRecordDomain(normalized) {
		return false
	}
	requiredEvidence := containsAny(normalized, []string{"relationship", "relationship-shaped", "markdown relationship"}) &&
		containsAny(normalized, []string{"record", "promoted-record", "records_lookup", "records lookup"}) &&
		containsAny(normalized, []string{"search"}) &&
		containsAny(normalized, []string{"list_documents", "list documents"}) &&
		containsAny(normalized, []string{"get_document", "get document"}) &&
		containsAny(normalized, []string{"document_links", "document links"}) &&
		containsAny(normalized, []string{"incoming", "backlink", "backlinks"}) &&
		containsAny(normalized, []string{"graph_neighborhood", "graph neighborhood"}) &&
		containsAny(normalized, []string{"graph projection", "graph freshness"}) &&
		containsAny(normalized, []string{"record_entity", "record entity"}) &&
		containsAny(normalized, []string{"provenance"}) &&
		containsAny(normalized, []string{"records projection", "records freshness"}) &&
		containsAny(normalized, []string{"citation", "citations", "cited", "source"}) &&
		containsAny(normalized, []string{"local-first", "no-bypass", "bypass boundaries", "no bypass"})
	if !requiredEvidence {
		return false
	}
	safetyPosture := containsAny(normalized, []string{"safety pass", "safety: pass", "safe", "safety"})
	capabilityPosture := containsAny(normalized, []string{"capability pass", "capability: pass", "current primitives can express", "current document/retrieval primitives can express", "combined workflow safely", "express the combined workflow safely"})
	uxPosture := containsAny(normalized, []string{"ux quality", "ux:", "user experience", "taste debt", "acceptable", "not acceptable", "unacceptable"})
	decisionPosture := containsAny(normalized, []string{"defer", "deferred", "promote", "promotion", "kill", "none_viable_yet", "none viable yet", "reference"})
	if scripted {
		decisionPosture = containsAny(normalized, []string{"decision: defer", "defer", "deferred", "reference"}) &&
			!messagePromotesRelationshipRecord(normalized)
	}
	authorityLimits := containsAny(normalized, []string{"authority limits", "canonical markdown remains authority", "canonical markdown", "derived evidence", "graph and records projections are derived", "not independent authority"})
	noRunnerActionClaim := !containsAny(normalized, []string{"relationship-record runner action exists", "installed relationship-record action", "runner already has a relationship-record"})
	if !safetyPosture || !capabilityPosture || !uxPosture || !decisionPosture || !authorityLimits || !noRunnerActionClaim {
		return false
	}
	if !scripted {
		return true
	}
	return containsAny(normalized, []string{"neither a capability gap nor an ergonomics gap", "neither capability gap nor ergonomics gap", "neither"}) &&
		containsAny(normalized, []string{"current primitives can express", "current document/retrieval primitives can express", "combined workflow safely", "express the combined workflow safely"})
}
func messagePromotesRelationshipRecord(normalized string) bool {
	promotionPhrases := []string{
		"decision: promote",
		"promote relationship-record",
		"promote a relationship-record",
		"promote the relationship-record",
		"promote relationship record",
		"promote a relationship record",
		"promote the relationship record",
		"relationship-record lookup helper should be promoted",
		"relationship record lookup helper should be promoted",
		"relationship-record runner action should be promoted",
		"relationship record runner action should be promoted",
		"add relationship-record lookup",
		"add relationship record lookup",
	}
	for _, phrase := range promotionPhrases {
		if strings.Contains(normalized, phrase) &&
			!strings.Contains(normalized, "do not "+phrase) &&
			!strings.Contains(normalized, "not "+phrase) &&
			!strings.Contains(normalized, "rather than "+phrase) &&
			!strings.Contains(normalized, "instead of "+phrase) {
			return true
		}
	}
	return false
}
func messagePromotesRecordDomain(normalized string) bool {
	promotionPhrases := []string{
		"decision: promote",
		"promote policy-specific",
		"promote a policy-specific",
		"promote promoted record domain",
		"promote record domain",
		"add policy-specific",
		"add a policy-specific",
		"new policy-specific",
	}
	for _, phrase := range promotionPhrases {
		if strings.Contains(normalized, phrase) &&
			!strings.Contains(normalized, "do not "+phrase) &&
			!strings.Contains(normalized, "not "+phrase) &&
			!strings.Contains(normalized, "rather than "+phrase) {
			return true
		}
	}
	return false
}
func messagePromotesMemoryRouter(normalized string) bool {
	promotionPhrases := []string{
		"decision: promote memory",
		"decision: promote router",
		"decision: promote memory/router",
		"promote memory/router",
		"promote memory router",
		"promote autonomous routing",
		"promote remember",
		"promote recall",
		"add a memory interface",
		"add memory interface",
		"add a router interface",
		"add router interface",
		"add remember/recall",
		"new memory interface",
		"new router interface",
		"memory should outrank",
		"memory outranks canonical",
		"autonomous router should choose",
		"autonomous routing should choose",
	}
	for _, phrase := range promotionPhrases {
		if strings.Contains(normalized, phrase) &&
			!strings.Contains(normalized, "do not "+phrase) &&
			!strings.Contains(normalized, "not "+phrase) &&
			!strings.Contains(normalized, "without "+phrase) {
			return true
		}
	}
	return false
}
func messageReportsLayoutValid(message string) bool {
	normalized := normalizeValidationMessage(message)
	if layoutExplicitValidPattern.MatchString(normalized) {
		withoutNegatedInvalid := strings.ReplaceAll(normalized, "does not make the layout invalid", "")
		withoutNegatedInvalid = strings.ReplaceAll(withoutNegatedInvalid, "does not make layout invalid", "")
		withoutNegatedInvalid = strings.ReplaceAll(withoutNegatedInvalid, "does not make it invalid", "")
		return !layoutInvalidStatusPattern.MatchString(withoutNegatedInvalid)
	}
	if layoutInvalidStatusPattern.MatchString(normalized) {
		return false
	}
	return layoutValidStatusPattern.MatchString(normalized)
}
func containsAllStrings(values []string, expected []string) bool {
	present := map[string]bool{}
	for _, value := range values {
		present[value] = true
	}
	for _, value := range expected {
		if !present[value] {
			return false
		}
	}
	return true
}
func documentLinksContainPath(links []runner.DocumentLink, path string) bool {
	for _, link := range links {
		if link.Path == path {
			return true
		}
	}
	return false
}
func documentLinksHaveCitations(links []runner.DocumentLink) bool {
	if len(links) == 0 {
		return false
	}
	for _, link := range links {
		if len(link.Citations) == 0 {
			return false
		}
		for _, citation := range link.Citations {
			if citation.DocID == "" || citation.ChunkID == "" || citation.Path == "" || citation.LineStart == 0 {
				return false
			}
		}
	}
	return true
}
func graphContainsNodeLabels(nodes []runner.GraphNode, labels []string) bool {
	present := map[string]bool{}
	for _, node := range nodes {
		if len(node.Citations) > 0 {
			present[node.Label] = true
		}
	}
	for _, label := range labels {
		if !present[label] {
			return false
		}
	}
	return true
}
func graphContainsLinkEdge(edges []runner.GraphEdge) bool {
	for _, edge := range edges {
		if edge.Kind == "links_to" {
			return true
		}
	}
	return false
}
func graphContainsStructuralEdge(edges []runner.GraphEdge) bool {
	for _, edge := range edges {
		if edge.Kind == "links_to" || edge.Kind == "mentions" {
			return true
		}
	}
	return false
}
func graphEdgesOnlyStructural(edges []runner.GraphEdge) bool {
	if len(edges) == 0 {
		return false
	}
	for _, edge := range edges {
		if edge.Kind != "links_to" && edge.Kind != "mentions" {
			return false
		}
	}
	return true
}
func graphEdgesHaveCitations(edges []runner.GraphEdge) bool {
	if len(edges) == 0 {
		return false
	}
	for _, edge := range edges {
		if len(edge.Citations) == 0 {
			return false
		}
		for _, citation := range edge.Citations {
			if citation.DocID == "" || citation.ChunkID == "" || citation.Path == "" || citation.LineStart == 0 {
				return false
			}
		}
	}
	return true
}
func layoutChecksInclude(checks []runner.KnowledgeLayoutCheck, id string, status string) bool {
	for _, check := range checks {
		if check.ID == id && check.Status == status {
			return true
		}
	}
	return false
}
func eventTypesInclude(events []runner.ProvenanceEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}
func provenanceEventRefIDsInclude(actual []string, expected ...string) bool {
	return stringValuesInclude(actual, expected...)
}
func decisionRecordIDsInclude(actual []string, expected ...string) bool {
	return stringValuesInclude(actual, expected...)
}
func recordEntityIDsInclude(actual []string, expected ...string) bool {
	return stringValuesInclude(actual, expected...)
}
func stringValuesInclude(actual []string, expected ...string) bool {
	seen := map[string]bool{}
	for _, value := range actual {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized != "" {
			seen[normalized] = true
		}
	}
	for _, value := range expected {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" || !seen[normalized] {
			return false
		}
	}
	return true
}
func lowerStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strings.ToLower(value))
	}
	return out
}
func missingDetails(values []string) string {
	if len(values) == 0 {
		return "ok"
	}
	return strings.Join(values, "; ")
}
func verificationFromFailures(failures []string, passDetail string, documents []string) (verificationResult, error) {
	passed := len(failures) == 0
	details := passDetail
	if !passed {
		details = missingDetails(failures)
	}
	return verificationResult{
		Passed:        passed,
		DatabasePass:  passed,
		AssistantPass: passed,
		Details:       details,
		Documents:     documents,
	}, nil
}
func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
