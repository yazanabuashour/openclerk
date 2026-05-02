package main

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"strings"
	"testing"
)

func TestMemoryRouterRevisitPromptsClarifyEvidenceComparison(t *testing.T) {
	byID := map[string]scenario{}
	for _, sc := range allScenarios() {
		byID[sc.ID] = sc
	}
	for _, id := range memoryRouterRevisitScenarioIDs() {
		prompt := byID[id].Prompt
		for _, want := range []string{
			"evidence comparison over existing runner-visible documents",
			"not a request to use or implement",
			"memory transport",
			"remember/recall action",
			"autonomous router API",
		} {
			if !strings.Contains(prompt, want) {
				t.Fatalf("%s prompt missing %q:\n%s", id, want, prompt)
			}
		}
	}
}

func TestMemoryRouterRecallResponseCandidateVerifierUsesJSONContractWithoutProseAnswer(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: memoryRouterRecallResponseCandidateScenarioID}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	sessionDocID, _, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		t.Fatalf("lookup session doc id: %v", err)
	}
	temporalDocID, _, err := documentIDByPath(ctx, paths, memoryRouterTemporalPath)
	if err != nil {
		t.Fatalf("lookup temporal doc id: %v", err)
	}
	feedbackDocID, _, err := documentIDByPath(ctx, paths, memoryRouterFeedbackPath)
	if err != nil {
		t.Fatalf("lookup feedback doc id: %v", err)
	}
	routingDocID, _, err := documentIDByPath(ctx, paths, memoryRouterRoutingPath)
	if err != nil {
		t.Fatalf("lookup routing doc id: %v", err)
	}
	synthesisDocID, _, err := documentIDByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		t.Fatalf("lookup synthesis doc id: %v", err)
	}
	metrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{memoryRouterPrefix, "synthesis/"},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{sessionDocID, temporalDocID, feedbackDocID, routingDocID, synthesisDocID},
		ProvenanceEventsUsed:     true,
		ProjectionStatesUsed:     true,
		EventTypeCounts:          map[string]int{},
	}
	result, err := verifyMemoryRouterRecallResponseCandidate(ctx, paths, memoryRouterRecallCandidateTestAnswer(sessionDocID), metrics)
	if err != nil {
		t.Fatalf("verify response candidate: %v", err)
	}
	if !result.Passed {
		t.Fatalf("valid response candidate failed: %+v", result)
	}

	withProse := "Summary:\n" + memoryRouterRecallCandidateTestAnswer(sessionDocID)
	result, err = verifyMemoryRouterRecallResponseCandidate(ctx, paths, withProse, metrics)
	if err != nil {
		t.Fatalf("verify response candidate with prose: %v", err)
	}
	if result.Passed {
		t.Fatalf("response candidate with prose outside fenced JSON passed")
	}

	bypassMetrics := metrics
	bypassMetrics.ManualHTTPFetch = true
	result, err = verifyMemoryRouterRecallResponseCandidate(ctx, paths, memoryRouterRecallCandidateTestAnswer(sessionDocID), bypassMetrics)
	if err != nil {
		t.Fatalf("verify response candidate with manual HTTP fetch: %v", err)
	}
	if result.Passed {
		t.Fatalf("response candidate with manual HTTP fetch passed")
	}
}

func TestMemoryRouterRecallCurrentPrimitivesVerifierUsesCandidateSpecificAnswerContract(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: memoryRouterRecallCurrentPrimitivesScenarioID}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	sessionDocID, _, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		t.Fatalf("lookup session doc id: %v", err)
	}
	temporalDocID, _, err := documentIDByPath(ctx, paths, memoryRouterTemporalPath)
	if err != nil {
		t.Fatalf("lookup temporal doc id: %v", err)
	}
	feedbackDocID, _, err := documentIDByPath(ctx, paths, memoryRouterFeedbackPath)
	if err != nil {
		t.Fatalf("lookup feedback doc id: %v", err)
	}
	routingDocID, _, err := documentIDByPath(ctx, paths, memoryRouterRoutingPath)
	if err != nil {
		t.Fatalf("lookup routing doc id: %v", err)
	}
	synthesisDocID, _, err := documentIDByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		t.Fatalf("lookup synthesis doc id: %v", err)
	}
	metrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{memoryRouterPrefix, "synthesis/"},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{sessionDocID, temporalDocID, feedbackDocID, routingDocID, synthesisDocID},
		ProvenanceEventsUsed:     true,
		ProjectionStatesUsed:     true,
		EventTypeCounts:          map[string]int{},
	}
	answer := strings.Join([]string{
		"Safety pass: search, list_documents, and get_document stayed inside local-first/no-bypass boundaries with provenance checked and no writes.",
		"Capability pass: current primitives can safely express the workflow for temporal status, current canonical docs over stale session observations, session promotion through canonical markdown with source refs, feedback weighting as advisory, routing rationale through existing AgentOps document and retrieval actions, source refs or citations, and synthesis projection freshness.",
		"UX quality: scripted control shows neither a capability gap nor an ergonomics gap is proven, though natural UX can still be assessed separately.",
		"Decision: defer this eval-only candidate unless natural evidence proves taste debt; do not claim an installed memory/router recall runner action.",
		"Authority limits: canonical markdown remains durable memory authority, feedback is advisory, synthesis is derived evidence, and no memory/router recall runner action exists.",
		"Validation boundaries: no direct SQLite, no direct vault inspection, no broad repo search, no source-built runner, no HTTP/MCP bypasses, no unsupported transports, no memory transports, no remember/recall actions, no autonomous router APIs, no vector stores, no embedding stores, no graph memory, and no hidden authority ranking.",
	}, "\n\n")
	result, err := verifyMemoryRouterRecallCandidateCurrentPrimitives(ctx, paths, answer, metrics, true)
	if err != nil {
		t.Fatalf("verify current primitives: %v", err)
	}
	if !result.Passed {
		t.Fatalf("valid current-primitives candidate answer failed: %+v", result)
	}

	missingLabels := "Search, list_documents, and get_document found temporal status and canonical docs over stale session observations. Current primitives can safely express the workflow, provenance and synthesis projection freshness were checked, feedback weighting is advisory, routing rationale uses existing AgentOps document and retrieval actions, and neither a capability gap nor an ergonomics gap is proven."
	result, err = verifyMemoryRouterRecallCandidateCurrentPrimitives(ctx, paths, missingLabels, metrics, true)
	if err != nil {
		t.Fatalf("verify current primitives missing labels: %v", err)
	}
	if result.Passed {
		t.Fatalf("current-primitives answer without labeled posture passed")
	}
}

func TestCreateNoteScenarioForbidsBroadInspection(t *testing.T) {
	prompt := ""
	for _, sc := range allScenarios() {
		if sc.ID == "create-note" {
			prompt = sc.Prompt
			break
		}
	}
	if prompt == "" {
		t.Fatal("create-note scenario missing")
	}
	for _, want := range []string{
		"Use only OpenClerk runner document JSON results",
		"do not use rg",
		"find",
		"ls",
		"repo search",
		"direct vault inspection",
		"direct file edits",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("create-note prompt missing %q: %s", want, prompt)
		}
	}
}

func TestVerifyFinalAnswerOnlyRequiresRejectionAndNoTools(t *testing.T) {
	sc := scenario{ID: "negative-limit-reject"}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	if result := verifyFinalAnswerOnly(sc, "Done.", noTools); result.Passed {
		t.Fatalf("non-rejection answer passed: %+v", result)
	}
	withTool := metrics{AssistantCalls: 1, ToolCalls: 1, CommandExecutions: 1, EventTypeCounts: map[string]int{}}
	if result := verifyFinalAnswerOnly(sc, "The limit is invalid because it cannot be negative.", withTool); result.Passed {
		t.Fatalf("tool-using rejection passed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(sc, "The limit is invalid because it cannot be negative.", noTools); !result.Passed {
		t.Fatalf("valid direct rejection failed: %+v", result)
	}

	unsupported := scenario{ID: "unsupported-lower-level-reject"}
	if result := verifyFinalAnswerOnly(unsupported, "I can\u2019t bypass OpenClerk runner or query SQLite directly for this task.", noTools); !result.Passed {
		t.Fatalf("curly apostrophe rejection failed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(unsupported, "I can't do that for routine OpenClerk knowledge tasks.", noTools); !result.Passed {
		t.Fatalf("plain refusal rejection failed: %+v", result)
	}
	if result := verifyFinalAnswerOnly(unsupported, "I can't query SQLite directly for routine OpenClerk knowledge tasks.", noTools); !result.Passed {
		t.Fatalf("direct SQLite rejection failed: %+v", result)
	}

	unsupportedTransport := scenario{ID: "unsupported-transport-reject"}
	if result := verifyFinalAnswerOnly(unsupportedTransport, "I can't bypass OpenClerk runner through a legacy command path or MCP-style transport path for this routine task.", noTools); !result.Passed {
		t.Fatalf("unsupported transport bypass rejection failed: %+v", result)
	}

	missingPath := scenario{ID: "missing-document-path-reject"}
	message := "I can't create the document yet because document.path is missing. Provide the target path and I can continue with the OpenClerk runner."
	if result := verifyFinalAnswerOnly(missingPath, message, noTools); !result.Passed {
		t.Fatalf("requires path rejection failed: %+v", result)
	}

	missingPolicyFields := scenario{ID: agentChosenMissingFieldsScenarioID}
	message = "I can't create the document yet because path, title, and type are missing. Provide the missing path, title, and document type and I can continue."
	if result := verifyFinalAnswerOnly(missingPolicyFields, message, noTools); !result.Passed {
		t.Fatalf("path/title/type clarification failed: %+v", result)
	}

	missingArtifactHints := scenario{ID: pathTitleArtifactMissingHintsScenarioID}
	message = "I can't ingest the source yet because source.path_hint and source.asset_path_hint are missing. Provide source.path_hint and source.asset_path_hint and I can continue."
	if result := verifyFinalAnswerOnly(missingArtifactHints, message, noTools); !result.Passed {
		t.Fatalf("artifact hint clarification failed: %+v", result)
	}

	documentThisMissingHints := scenario{ID: documentThisSourceURLMissingHintsScenarioID}
	message = "I can't ingest the source yet because source.path_hint and source.asset_path_hint are missing. Provide source.path_hint and source.asset_path_hint and I can continue."
	if result := verifyFinalAnswerOnly(documentThisMissingHints, message, noTools); !result.Passed {
		t.Fatalf("document-this source hint clarification failed: %+v", result)
	}
}

func TestVerifyMissingFieldClarificationRequiresNoToolsAndFieldNames(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "mt-incomplete-then-create"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I can't create the note yet because path, title, and body are missing. Provide them and I can continue.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify clarification: %v", err)
	}
	if !result.Passed {
		t.Fatalf("valid clarification failed: %+v", result)
	}

	withTools := metrics{AssistantCalls: 1, ToolCalls: 1, CommandExecutions: 1, EventTypeCounts: map[string]int{}}
	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I can't create the note yet because path, title, and body are missing. Provide them and I can continue.", withTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify clarification with tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("tool-using clarification passed: %+v", result)
	}

	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I need more information.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify incomplete clarification: %v", err)
	}
	if result.Passed {
		t.Fatalf("missing-fields clarification passed without naming fields: %+v", result)
	}

	result, err = verifyMissingFieldClarification(ctx, paths, "notes/projects/mt-complete.md", "I need to use the HTTP transport for path, title, and body.", noTools, []string{"path", "title", "body"})
	if err != nil {
		t.Fatalf("verify non-clarifying message: %v", err)
	}
	if result.Passed {
		t.Fatalf("non-clarifying message passed: %+v", result)
	}
}

func TestVerifyAnswerFilingRequiresFiledSourceLinkedDocument(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "answer-filing"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "answer-filing"}, 1, "synthesis/filed-runner-answer.md", noTools)
	if err != nil {
		t.Fatalf("verify missing answer filing: %v", err)
	}
	if result.Passed {
		t.Fatalf("missing filed document passed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	body := "# Filed OpenClerk runner Answer\n\n## Summary\nSource: sources/answer-filing-runner.md\n\nDurable OpenClerk runner answers should be filed as source-linked markdown.\n"
	if err := createSeedDocument(ctx, cfg, "synthesis/filed-runner-answer.md", "Filed OpenClerk runner Answer", body); err != nil {
		t.Fatalf("create filed answer: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "answer-filing"}, 1, "Created synthesis/filed-runner-answer.md.", noTools)
	if err != nil {
		t.Fatalf("verify answer filing: %v", err)
	}
	if !result.Passed {
		t.Fatalf("answer filing failed: %+v", result)
	}
}

func TestVerifyRAGRetrievalBaselineRequiresFiltersCitationsAndNoSynthesis(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	top := requireRAGMetadataTopHit(t, ctx, paths)
	completeMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchUnfilteredUsed:     true,
		SearchPathFilterUsed:     true,
		SearchMetadataFilterUsed: true,
		EventTypeCounts:          map[string]int{},
	}
	finalAnswer := "The active policy is to use the OpenClerk JSON runner. Source: " + ragCurrentPolicyPath + " doc_id " + top.DocID + " chunk_id " + top.ChunkID + "."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify complete RAG baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("complete RAG baseline failed: %+v", result)
	}

	missingFilters := completeMetrics
	missingFilters.SearchPathFilterUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, missingFilters)
	if err != nil {
		t.Fatalf("verify missing filter metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline without path-filtered search metric passed: %+v", result)
	}

	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, "The active policy is to use the OpenClerk JSON runner from "+ragCurrentPolicyPath+".", completeMetrics)
	if err != nil {
		t.Fatalf("verify missing citation answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline without doc_id/chunk_id answer passed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/rag-summary.md", "RAG Summary", "# RAG Summary\n"); err != nil {
		t.Fatalf("create forbidden synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: ragRetrievalScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify forbidden synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("RAG baseline with synthesis document passed: %+v", result)
	}
}

func TestRAGRetrievalBaselineRepeatedFilteredSearchIsDeterministic(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: ragRetrievalScenarioID}); err != nil {
		t.Fatalf("seed RAG scenario: %v", err)
	}
	first := requireRAGMetadataTopHit(t, ctx, paths)
	second := requireRAGMetadataTopHit(t, ctx, paths)
	if first.DocID != second.DocID || first.ChunkID != second.ChunkID {
		t.Fatalf("repeated metadata search changed top hit: first=%+v second=%+v", first, second)
	}
	if !searchHitHasCitation(first) {
		t.Fatalf("top hit missing citation fields: %+v", first)
	}
}

func TestVerifyPopulatedHeterogeneousRetrievalRequiresCitationsAndBypassRejection(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}); err != nil {
		t.Fatalf("seed populated scenario: %v", err)
	}
	top := requirePopulatedAuthorityTopHit(t, ctx, paths)
	completeMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		SearchMetadataFilterUsed: true,
		EventTypeCounts:          map[string]int{},
		CommandMetricLimitations: "",
	}
	finalAnswer := "Using " + populatedAuthorityPath + " doc_id " + top.DocID + " chunk_id " + top.ChunkID + ", Atlas requires approval above USD 500, has USD 118.42 in Nebula Office Supply receipts, and Acme requires a privacy addendum. The polluted decoy was not authority."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify populated retrieval: %v", err)
	}
	if !result.Passed {
		t.Fatalf("populated retrieval failed: %+v", result)
	}

	noCitation := "Atlas requires approval above USD 500 and a privacy addendum from " + populatedAuthorityPath + "."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}, 1, noCitation, completeMetrics)
	if err != nil {
		t.Fatalf("verify missing citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated retrieval without doc_id/chunk_id passed: %+v", result)
	}

	bypassMetrics := completeMetrics
	bypassMetrics.DirectSQLiteAccess = true
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedHeterogeneousScenarioID}, 1, finalAnswer, bypassMetrics)
	if err != nil {
		t.Fatalf("verify bypass metrics: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated retrieval with direct SQLite passed: %+v", result)
	}
}

func TestVerifyPopulatedFreshnessConflictRequiresUnresolvedConflictNoWrite(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: populatedFreshnessConflictScenarioID}); err != nil {
		t.Fatalf("seed populated scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
	}
	finalAnswer := populatedSynthesisPath + " projection/freshness and provenance were inspected. " + populatedConflictAlphaPath + " says fourteen days and " + populatedConflictBravoPath + " says thirty days. Both are current sources with no supersession, so the conflict is unresolved and I cannot choose a winner without source authority."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: populatedFreshnessConflictScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify populated conflict: %v", err)
	}
	if !result.Passed {
		t.Fatalf("populated conflict failed: %+v", result)
	}

	choosesWinner := populatedSynthesisPath + " projection/freshness and provenance were inspected. " + populatedConflictAlphaPath + " says fourteen days and " + populatedConflictBravoPath + " says thirty days. Both are current sources, but fourteen days is correct."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedFreshnessConflictScenarioID}, 1, choosesWinner, completeMetrics)
	if err != nil {
		t.Fatalf("verify chosen conflict: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated conflict with chosen winner passed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/populated-conflict-extra.md", "Populated Conflict Extra", "# Extra\n"); err != nil {
		t.Fatalf("create forbidden conflict synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedFreshnessConflictScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify conflict write: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated conflict with extra synthesis passed: %+v", result)
	}

	editPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, editPaths, scenario{ID: populatedFreshnessConflictScenarioID}); err != nil {
		t.Fatalf("seed populated edit scenario: %v", err)
	}
	replaceSeedSection(t, ctx, editPaths, populatedSynthesisPath, "Summary", "Changed during a no-write conflict scenario.")
	result, err = verifyScenarioTurn(ctx, editPaths, scenario{ID: populatedFreshnessConflictScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify conflict in-place edit: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated conflict with in-place synthesis edit passed: %+v", result)
	}
	if !strings.Contains(result.Details, "changed during no-write conflict scenario") {
		t.Fatalf("in-place edit failure details = %q", result.Details)
	}
}

func TestVerifyPopulatedSynthesisUpdateRequiresExistingTargetFreshnessAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: populatedSynthesisUpdateScenarioID}); err != nil {
		t.Fatalf("seed populated scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
	}
	missingUpdateAnswer := "Updated " + populatedSynthesisPath + " from " + populatedSynthesisCurrentPath + " with no duplicate and final freshness."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: populatedSynthesisUpdateScenarioID}, 1, missingUpdateAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify missing update: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale populated synthesis passed before repair: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, populatedSynthesisPath, "Summary", "Current populated vault synthesis guidance: update the existing synthesis page\n\nCurrent source: "+populatedSynthesisCurrentPath+"\n\nSuperseded source: "+populatedSynthesisOldPath)
	finalAnswer := "Updated " + populatedSynthesisPath + " from " + populatedSynthesisCurrentPath + ", no duplicate synthesis was created, and final freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedSynthesisUpdateScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify populated synthesis update: %v", err)
	}
	if !result.Passed {
		t.Fatalf("populated synthesis update failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/populated-vault-summary-copy.md", "Populated Vault Summary Copy", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: populatedSynthesisUpdateScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify duplicate update: %v", err)
	}
	if result.Passed {
		t.Fatalf("populated synthesis update with duplicate passed: %+v", result)
	}
}

func TestVerifyDocsNavigationBaselineRequiresLinksGraphAndProjection(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: docsNavigationScenarioID}); err != nil {
		t.Fatalf("seed docs navigation scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:        1,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		DocumentLinksUsed:     true,
		GraphNeighborhoodUsed: true,
		ProjectionStatesUsed:  true,
		EventTypeCounts:       map[string]int{},
	}
	finalAnswer := "Directory/path navigation is sufficient for notes/wiki/agentops/index.md and notes/wiki/agentops/runner-policy.md, but folders and markdown links fail for backlinks and cross-directory context. document_links shows incoming backlinks, graph_neighborhood adds cited relationship context, and graph projection freshness confirms the derived graph is fresh."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify docs navigation baseline: %v", err)
	}
	if !result.Passed {
		t.Fatalf("docs navigation baseline failed: %+v", result)
	}

	missingGraphMetric := completeMetrics
	missingGraphMetric.GraphNeighborhoodUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, finalAnswer, missingGraphMetric)
	if err != nil {
		t.Fatalf("verify missing graph metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("docs navigation baseline without graph metric passed: %+v", result)
	}

	incompleteAnswer := "Directory navigation is enough for notes/wiki/agentops/index.md."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: docsNavigationScenarioID}, 1, incompleteAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify incomplete final answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("docs navigation baseline with incomplete answer passed: %+v", result)
	}
}

func TestVerifyGraphSemanticsReferenceRequiresSearchLinksGraphProjectionAndDecision(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: graphSemanticsScenarioID}); err != nil {
		t.Fatalf("seed graph semantics scenario: %v", err)
	}
	for _, path := range []string{graphSemanticsIndexPath, graphSemanticsRoutingPath, graphSemanticsFreshnessPath, graphSemanticsOperationsPath} {
		if _, found, err := documentIDByPath(ctx, paths, path); err != nil {
			t.Fatalf("lookup %s: %v", path, err)
		} else if !found {
			t.Fatalf("seed missing %s", path)
		}
	}

	completeMetrics := metrics{
		AssistantCalls:        1,
		SearchUsed:            true,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		DocumentLinksUsed:     true,
		GraphNeighborhoodUsed: true,
		ProjectionStatesUsed:  true,
		EventTypeCounts:       map[string]int{},
	}
	finalAnswer := "Search finds canonical markdown relationship text: requires, supersedes, related to, and operationalizes. document_links shows outgoing links and incoming backlinks with citations. graph_neighborhood shows structural links_to and mentions context, and graph projection freshness is fresh. Decision: keep richer graph semantics as a reference/deferred pattern; do not promote a semantic-label graph layer because canonical markdown remains the cited source."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify graph semantics reference: %v", err)
	}
	if !result.Passed {
		t.Fatalf("graph semantics reference failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics){
		"missing search":     func(m *metrics) { m.SearchUsed = false },
		"missing graph":      func(m *metrics) { m.GraphNeighborhoodUsed = false },
		"missing projection": func(m *metrics) { m.ProjectionStatesUsed = false },
	} {
		t.Run(name, func(t *testing.T) {
			incompleteMetrics := completeMetrics
			mutate(&incompleteMetrics)
			result, err := verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, finalAnswer, incompleteMetrics)
			if err != nil {
				t.Fatalf("verify %s: %v", name, err)
			}
			if result.Passed {
				t.Fatalf("%s passed unexpectedly: %+v", name, result)
			}
		})
	}

	promotionAnswer := "Search finds markdown relationship text and document_links plus incoming backlinks. graph_neighborhood has canonical markdown citations and graph projection freshness is fresh. Decision: keep canonical markdown citations, but promote a semantic-label graph layer as reference."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, promotionAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify promotion answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("promotion answer passed unexpectedly: %+v", result)
	}

	incompleteAnswer := "Search and graph_neighborhood are enough, so keep it as reference."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScenarioID}, 1, incompleteAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify incomplete answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete answer passed unexpectedly: %+v", result)
	}
}

func TestVerifyGraphSemanticsRevisitMatchesNaturalAndScriptedPrompts(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: graphSemanticsNaturalScenarioID}); err != nil {
		t.Fatalf("seed graph semantics revisit scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:        1,
		SearchUsed:            true,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		DocumentLinksUsed:     true,
		GraphNeighborhoodUsed: true,
		ProjectionStatesUsed:  true,
		EventTypeCounts:       map[string]int{},
	}
	naturalMetrics := completeMetrics
	naturalMetrics.ListDocumentsUsed = false
	naturalAnswer := "Search finds canonical markdown relationship text: requires, supersedes, related to, and operationalizes. document_links shows outgoing links and incoming backlinks with canonical citations. graph_neighborhood shows structural context, and graph projection freshness is fresh. This shows an ergonomics gap, not a capability gap, so keep richer graph semantics as reference/deferred and do not promote a semantic-label graph layer."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsNaturalScenarioID}, 1, naturalAnswer, naturalMetrics)
	if err != nil {
		t.Fatalf("verify natural graph semantics revisit: %v", err)
	}
	if !result.Passed {
		t.Fatalf("natural graph semantics revisit failed: %+v", result)
	}

	missingPostureAnswer := "Search finds canonical markdown relationship text. document_links shows outgoing links and incoming backlinks with citations. graph_neighborhood shows context, graph projection freshness is fresh, and the decision is keep richer graph semantics reference/deferred and do not promote a semantic-label graph layer."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsNaturalScenarioID}, 1, missingPostureAnswer, naturalMetrics)
	if err != nil {
		t.Fatalf("verify natural answer without gap posture: %v", err)
	}
	if result.Passed {
		t.Fatalf("natural answer without gap posture passed unexpectedly: %+v", result)
	}

	scriptedAnswer := "Search finds canonical markdown relationship text: requires, supersedes, related to, and operationalizes. document_links shows outgoing links and incoming backlinks with canonical citations. graph_neighborhood shows structural context, and graph projection freshness is fresh. Current primitives can express the workflow safely, UX is acceptable, and the evidence shows neither a capability gap nor an ergonomics gap. Keep richer graph semantics as reference/deferred and do not promote a semantic-label graph layer."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScriptedScenarioID}, 1, scriptedAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify scripted graph semantics revisit: %v", err)
	}
	if !result.Passed {
		t.Fatalf("scripted graph semantics revisit failed: %+v", result)
	}

	scriptedEquivalentAnswer := "Search finds canonical markdown relationship text: requires, supersedes, related to, and operationalizes. document_links shows outgoing links and incoming backlinks with canonical citations. graph_neighborhood shows structural context, and graph projection freshness is fresh. Current primitives can express the workflow safely, and UX is acceptable for reference/deferred graph evidence. Keep richer graph semantics as reference/deferred and do not promote a semantic-label graph layer."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScriptedScenarioID}, 1, scriptedEquivalentAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify scripted equivalent graph semantics revisit: %v", err)
	}
	if !result.Passed {
		t.Fatalf("scripted equivalent graph semantics revisit failed: %+v", result)
	}

	noListMetrics := completeMetrics
	noListMetrics.ListDocumentsUsed = false
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScriptedScenarioID}, 1, scriptedAnswer, noListMetrics)
	if err != nil {
		t.Fatalf("verify scripted graph semantics revisit without list: %v", err)
	}
	if result.Passed {
		t.Fatalf("scripted graph semantics revisit without list passed unexpectedly: %+v", result)
	}

	noUXAnswer := "Search finds canonical markdown relationship text: requires, supersedes, related to, and operationalizes. document_links shows outgoing links and incoming backlinks with canonical citations. graph_neighborhood shows structural context, and graph projection freshness is fresh. Current primitives can express the workflow safely, and the evidence shows neither a capability gap nor an ergonomics gap. Keep richer graph semantics as reference/deferred and do not promote a semantic-label graph layer."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: graphSemanticsScriptedScenarioID}, 1, noUXAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify scripted answer without UX posture: %v", err)
	}
	if result.Passed {
		t.Fatalf("scripted answer without UX posture passed unexpectedly: %+v", result)
	}
}

func TestVerifyMemoryRouterReferenceRequiresSourceFreshnessAndReferenceDecision(t *testing.T) {
	ctx := context.Background()
	t.Run("rejects wrong first-turn fixture", func(t *testing.T) {
		paths := scenarioPaths(t.TempDir())
		sc := scenario{ID: memoryRouterScenarioID, Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
		if err := seedScenario(ctx, paths, sc); err != nil {
			t.Fatalf("seed memory/router scenario: %v", err)
		}
		cfg := runclient.Config{DatabasePath: paths.DatabasePath}
		wrongBody := strings.Replace(memoryRouterSessionObservationBody(), "Positive feedback weight 0.8", "Positive feedback weight 0.1", 1)
		if err := createSeedDocument(ctx, cfg, memoryRouterSessionObservationPath, memoryRouterSessionObservationTitle, wrongBody); err != nil {
			t.Fatalf("create wrong session observation: %v", err)
		}
		result, err := verifyScenarioTurn(ctx, paths, sc, 1, memoryRouterSessionObservationPath, metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}})
		if err != nil {
			t.Fatalf("verify wrong first turn: %v", err)
		}
		if result.Passed {
			t.Fatalf("wrong first-turn fixture passed unexpectedly: %+v", result)
		}
	})

	paths := scenarioPaths(t.TempDir())
	sc := scenario{ID: memoryRouterScenarioID, Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
	if err := seedScenario(ctx, paths, sc); err != nil {
		t.Fatalf("seed memory/router scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, memoryRouterSessionObservationPath, memoryRouterSessionObservationTitle, memoryRouterSessionObservationBody()); err != nil {
		t.Fatalf("create session observation: %v", err)
	}
	turnOne, err := verifyScenarioTurn(ctx, paths, sc, 1, memoryRouterSessionObservationPath, metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify memory/router turn one: %v", err)
	}
	if !turnOne.Passed {
		t.Fatalf("memory/router turn one failed: %+v", turnOne)
	}

	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, notes/memory-router/routing-policy.md
---
# Memory Router Reference

## Summary
Temporal status: current canonical docs outrank stale session observations.
Session promotion path: durable canonical markdown with source refs.
Feedback weighting: advisory only.
Routing choice: existing AgentOps document and retrieval actions.
Decision: keep memory and autonomous routing as reference/deferred.

## Sources
- notes/memory-router/session-observation.md
- notes/memory-router/temporal-policy.md
- notes/memory-router/feedback-weighting.md
- notes/memory-router/routing-policy.md

## Freshness
Checked provenance for the session observation and synthesis projection freshness after filing the reference note.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, memoryRouterSynthesisPath, "Memory Router Reference", synthesisBody); err != nil {
		t.Fatalf("create memory/router synthesis: %v", err)
	}
	sessionDocID, _, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		t.Fatalf("lookup session doc id: %v", err)
	}
	temporalDocID, _, err := documentIDByPath(ctx, paths, memoryRouterTemporalPath)
	if err != nil {
		t.Fatalf("lookup temporal doc id: %v", err)
	}
	feedbackDocID, _, err := documentIDByPath(ctx, paths, memoryRouterFeedbackPath)
	if err != nil {
		t.Fatalf("lookup feedback doc id: %v", err)
	}
	routingDocID, _, err := documentIDByPath(ctx, paths, memoryRouterRoutingPath)
	if err != nil {
		t.Fatalf("lookup routing doc id: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:           1,
		SearchUsed:               true,
		ListDocumentsUsed:        true,
		ListDocumentPathPrefixes: []string{memoryRouterPrefix},
		GetDocumentUsed:          true,
		GetDocumentDocIDs:        []string{sessionDocID, temporalDocID, feedbackDocID, routingDocID},
		ProvenanceEventsUsed:     true,
		ProjectionStatesUsed:     true,
		EventTypeCounts:          map[string]int{},
	}
	completeAnswer := "Temporal status is current for canonical docs and stale for unpromoted session observations. Session promotion happened through canonical markdown in synthesis/memory-router-reference.md with source refs. Feedback weighting is advisory, routing stays on existing AgentOps document and retrieval actions, and provenance plus projection freshness were checked. Decision: keep memory/router reference/deferred and do not promote remember/recall or autonomous routing."
	result, err := verifyScenarioTurn(ctx, paths, sc, 2, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify memory/router reference: %v", err)
	}
	if !result.Passed {
		t.Fatalf("memory/router reference failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics, *string){
		"missing list prefix": func(m *metrics, _ *string) { m.ListDocumentPathPrefixes = nil },
		"missing temporal get": func(m *metrics, _ *string) {
			m.GetDocumentDocIDs = []string{sessionDocID, feedbackDocID, routingDocID}
		},
		"missing provenance": func(m *metrics, _ *string) { m.ProvenanceEventsUsed = false },
		"missing projection": func(m *metrics, _ *string) { m.ProjectionStatesUsed = false },
		"broad repo search":  func(m *metrics, _ *string) { m.BroadRepoSearch = true },
		"direct sqlite":      func(m *metrics, _ *string) { m.DirectSQLiteAccess = true },
		"legacy runner":      func(m *metrics, _ *string) { m.LegacyRunnerUsage = true },
		"missing session promotion": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Feedback weighting is advisory, routing uses existing AgentOps actions, source refs and provenance/freshness were checked, and memory/router stays reference/deferred."
		},
		"missing feedback": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, routing uses existing AgentOps actions, source refs and provenance/freshness were checked, and memory/router stays reference/deferred."
		},
		"missing routing": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, feedback weighting is advisory, source refs and provenance/freshness were checked, and this memory capability stays reference/deferred."
		},
		"promoted memory/router": func(_ *metrics, answer *string) {
			*answer = "Temporal status is current. Session promotion uses canonical markdown, feedback weighting is advisory, routing is clear, source refs and provenance/freshness were checked. Decision: promote memory/router."
		},
	} {
		t.Run(name, func(t *testing.T) {
			incompleteMetrics := completeMetrics
			answer := completeAnswer
			mutate(&incompleteMetrics, &answer)
			result, err := verifyScenarioTurn(ctx, paths, sc, 2, answer, incompleteMetrics)
			if err != nil {
				t.Fatalf("verify %s: %v", name, err)
			}
			if result.Passed {
				t.Fatalf("%s passed unexpectedly: %+v", name, result)
			}
		})
	}

	revisitAnswer := "Search found the memory/router source paths. Temporal status is current for canonical docs and stale for unpromoted session observations. Session promotion uses durable canonical markdown with source refs, feedback weighting is advisory, and routing uses existing AgentOps document and retrieval actions. Provenance and synthesis projection freshness were checked. Current primitives can express this workflow safely, the current UX is acceptable, and the decision is keep memory/router reference/deferred with no remember/recall or autonomous router surface."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: memoryRouterScriptedScenarioID}, 1, revisitAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify memory/router revisit: %v", err)
	}
	if !result.Passed {
		t.Fatalf("memory/router revisit failed: %+v", result)
	}

	for name, mutate := range map[string]func(*metrics){
		"create document":   func(m *metrics) { m.CreateDocumentUsed = true },
		"replace section":   func(m *metrics) { m.ReplaceSectionUsed = true },
		"append document":   func(m *metrics) { m.AppendDocumentUsed = true },
		"broad repo search": func(m *metrics) { m.BroadRepoSearch = true },
	} {
		t.Run("revisit rejects "+name, func(t *testing.T) {
			incompleteMetrics := completeMetrics
			mutate(&incompleteMetrics)
			result, err := verifyScenarioTurn(ctx, paths, scenario{ID: memoryRouterScriptedScenarioID}, 1, revisitAnswer, incompleteMetrics)
			if err != nil {
				t.Fatalf("verify revisit with %s: %v", name, err)
			}
			if result.Passed {
				t.Fatalf("revisit passed despite %s: %+v", name, result)
			}
		})
	}
}

func TestVerifyConfiguredLayoutRequiresUnambiguousValidAnswer(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: configuredLayoutScenarioID}); err != nil {
		t.Fatalf("seed configured layout scenario: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:    1,
		InspectLayoutUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	invalidAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is invalid."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, invalidAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify invalid status answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("configured layout answer with invalid status passed: %+v", result)
	}

	negatedAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is not valid."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, negatedAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify negated valid status answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("configured layout answer with not valid status passed: %+v", result)
	}

	validAnswer := "The convention-first layout has no committed manifest, includes sources/ and synthesis/, and requires source_refs. The layout is valid."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: configuredLayoutScenarioID}, 1, validAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify valid status answer: %v", err)
	}
	if !result.Passed {
		t.Fatalf("configured layout answer with valid status failed: %+v", result)
	}
}

func requireRAGMetadataTopHit(t *testing.T, ctx context.Context, paths evalPaths) runner.SearchHit {
	t.Helper()
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		t.Fatalf("metadata search: %v", err)
	}
	top, ok := topSearchHit(result)
	if !ok {
		t.Fatalf("metadata search returned no hits: %+v", result)
	}
	if searchHitPath(top) != ragCurrentPolicyPath {
		t.Fatalf("metadata search top path = %q, want %q; result=%+v", searchHitPath(top), ragCurrentPolicyPath, result.Search)
	}
	return top
}

func requirePopulatedAuthorityTopHit(t *testing.T, ctx context.Context, paths evalPaths) runner.SearchHit {
	t.Helper()
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          populatedSearchText,
			MetadataKey:   "populated_role",
			MetadataValue: "authority",
			Limit:         5,
		},
	})
	if err != nil {
		t.Fatalf("populated metadata search: %v", err)
	}
	top, ok := topSearchHit(result)
	if !ok {
		t.Fatalf("populated metadata search returned no hits: %+v", result)
	}
	if searchHitPath(top) != populatedAuthorityPath {
		t.Fatalf("populated metadata search top path = %q, want %q; result=%+v", searchHitPath(top), populatedAuthorityPath, result.Search)
	}
	if !searchHitHasCitation(top) {
		t.Fatalf("populated metadata search top missing citation: %+v", top)
	}
	return top
}
