package main

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"strings"
	"testing"
)

func TestBroadContradictionAuditAnswerRequiresProjectionFreshness(t *testing.T) {
	message := "Search found source paths and citations. Updated synthesis/audit-runner-routing.md from sources/audit-runner-current.md; provenance and projection freshness are fresh. sources/audit-conflict-alpha.md says seven days and sources/audit-conflict-bravo.md says thirty days. Both sources are current, the conflict is unresolved because there is no source authority, and I cannot choose a winner. Neither a capability gap nor an ergonomics gap is proven. Current primitives can express the workflow safely, the UX is acceptable enough, and the decision is keep broad contradiction/audit reference/deferred, do not promote a broad semantic contradiction engine."
	if !broadContradictionAuditAnswerPass(message, true) {
		t.Fatalf("complete broad audit answer did not pass")
	}
	withoutProjection := strings.ReplaceAll(message, "projection freshness", "freshness")
	if broadContradictionAuditAnswerPass(withoutProjection, true) {
		t.Fatalf("broad audit answer passed without projection freshness")
	}
}

func TestVerifySourceLinkedSynthesisRequiresSourcesFreshnessAndWorkflow(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	incomplete := `---
type: synthesis
status: active
freshness: fresh
---

# OpenClerk runner

## Sources

## Freshness
Checked search results.
`
	if err := createSeedDocument(ctx, cfg, "synthesis/openclerk-runner.md", "OpenClerk runner", incomplete); err != nil {
		t.Fatalf("create incomplete synthesis: %v", err)
	}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify incomplete synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis without source_refs passed: %+v", result)
	}
	yamlListPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, yamlListPaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed YAML-list source_refs scenario: %v", err)
	}
	yamlListCfg := runclient.Config{DatabasePath: yamlListPaths.DatabasePath}
	yamlListSourceRefs := `---
type: synthesis
status: active
freshness: fresh
source_refs:
  - sources/openclerk-runner.md
---

# OpenClerk runner

## Summary
The runner preserves source refs.

## Sources
- sources/openclerk-runner.md

## Freshness
Checked runner search results for sources/openclerk-runner.md.
`
	if err := createSeedDocument(ctx, yamlListCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", yamlListSourceRefs); err != nil {
		t.Fatalf("create YAML-list source_refs synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, yamlListPaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify YAML-list source_refs synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis with YAML-list source_refs passed: %+v", result)
	}
	missingFreshnessPaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, missingFreshnessPaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed missing freshness scenario: %v", err)
	}
	missingFreshnessCfg := runclient.Config{DatabasePath: missingFreshnessPaths.DatabasePath}
	missingFreshness := `---
type: synthesis
status: active
source_refs: sources/openclerk-runner.md
---

# OpenClerk runner

## Sources
- sources/openclerk-runner.md
`
	if err := createSeedDocument(ctx, missingFreshnessCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", missingFreshness); err != nil {
		t.Fatalf("create missing freshness synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, missingFreshnessPaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify missing freshness synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis without freshness metadata passed: %+v", result)
	}
	completePaths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, completePaths, scenario{ID: "search-synthesis"}); err != nil {
		t.Fatalf("seed complete scenario: %v", err)
	}
	completeCfg := runclient.Config{DatabasePath: completePaths.DatabasePath}
	complete := `---
type: synthesis
status: active
freshness: fresh
source_refs: sources/openclerk-runner.md
---

# OpenClerk runner

## Summary
The runner preserves source refs.

## Sources
- sources/openclerk-runner.md

## Freshness
Checked runner search results for sources/openclerk-runner.md.
`
	if err := createSeedDocument(ctx, completeCfg, "synthesis/openclerk-runner.md", "OpenClerk runner", complete); err != nil {
		t.Fatalf("create complete synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, completePaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis/openclerk-runner.md.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify complete synthesis: %v", err)
	}
	if !result.Passed {
		t.Fatalf("complete synthesis failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, completePaths, scenario{ID: "search-synthesis"}, 1, "Created synthesis.", metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	})
	if err != nil {
		t.Fatalf("verify final answer path: %v", err)
	}
	if result.Passed {
		t.Fatalf("synthesis final answer without path passed: %+v", result)
	}
}

func TestVerifyStaleSynthesisUpdateRequiresCurrentSourceAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "stale-synthesis-update"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md.", noTools)
	if err != nil {
		t.Fatalf("verify stale before update: %v", err)
	}
	if result.Passed {
		t.Fatalf("stale synthesis passed before update: %+v", result)
	}
	replacement := "Current guidance: routine agents must use openclerk JSON runner.\n\nCurrent source: sources/runner-current-runner.md\n\nSupersedes: sources/runner-old-workaround.md\n\nThis stale claim is superseded by current guidance."
	replaceSeedSection(t, ctx, paths, "synthesis/runner-routing.md", "Summary", replacement)
	replaceSeedSection(t, ctx, paths, "synthesis/runner-routing.md", "Freshness", "Checked current source: sources/runner-current-runner.md\n\nChecked previous source: sources/runner-old-workaround.md")
	workflowMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		GetDocumentUsed:   true,
		EventTypeCounts:   map[string]int{},
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md with current guidance.", workflowMetrics)
	if err != nil {
		t.Fatalf("verify stale after update: %v", err)
	}
	if !result.Passed {
		t.Fatalf("updated stale synthesis failed: %+v", result)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/runner-routing-current.md", "OpenClerk runner Routing Current", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "stale-synthesis-update"}, 1, "Updated synthesis/runner-routing.md with current guidance.", workflowMetrics)
	if err != nil {
		t.Fatalf("verify stale duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("duplicate synthesis passed: %+v", result)
	}
}

func TestVerifySourceSensitiveAuditRepairRequiresProvenanceFreshnessAndNoDuplicate(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}); err != nil {
		t.Fatalf("seed source audit repair scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, "Updated "+sourceAuditSynthesisPath+".", noTools)
	if err != nil {
		t.Fatalf("verify source audit before repair: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit repair passed before update: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, sourceAuditSynthesisPath, "Summary", "Current audit guidance: use the installed openclerk JSON runner.\n\nCurrent source: "+sourceAuditCurrentSourcePath+"\n\nSuperseded source: "+sourceAuditOldSourcePath)
	replaceSeedSection(t, ctx, paths, sourceAuditSynthesisPath, "Freshness", "Checked provenance events and synthesis projection freshness after the current source update.")
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    5,
		ToolCalls:            5,
	}
	finalAnswer := "Updated " + sourceAuditSynthesisPath + " from " + sourceAuditCurrentSourcePath + "; projection freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit repair: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source audit repair failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/audit-runner-routing-v2.md", "Audit Runner Routing V2", "# Duplicate\n"); err != nil {
		t.Fatalf("create duplicate audit synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditRepairScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit duplicate: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit repair passed with duplicate synthesis: %+v", result)
	}
}

func TestVerifyHighTouchCompileSynthesisNaturalRequiresProvenance(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: highTouchCompileSynthesisNaturalScenarioID}); err != nil {
		t.Fatalf("seed high-touch compile synthesis scenario: %v", err)
	}
	replaceSeedSection(t, ctx, paths, synthesisCompilePath, "Summary", "Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient.\n\nCurrent source: "+synthesisCompileCurrentSrc+"\n\nSuperseded source: "+synthesisCompileOldSrc)
	synthesisDocID, found, err := documentIDByPath(ctx, paths, synthesisCompilePath)
	if err != nil {
		t.Fatalf("lookup synthesis doc id: %v", err)
	}
	if !found {
		t.Fatalf("missing synthesis doc id for %s", synthesisCompilePath)
	}
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		ReplaceSectionUsed:   true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    5,
		ToolCalls:            5,
	}
	finalAnswer := "Updated " + synthesisCompilePath + " from " + synthesisCompileCurrentSrc + "; projection freshness is fresh."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: highTouchCompileSynthesisNaturalScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify high-touch compile synthesis without provenance: %v", err)
	}
	if result.Passed {
		t.Fatalf("high-touch compile synthesis passed without provenance events: %+v", result)
	}
	workflowMetrics.ProvenanceEventsUsed = true
	workflowMetrics.ProvenanceEventRefIDs = []string{"unrelated-ref"}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: highTouchCompileSynthesisNaturalScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify high-touch compile synthesis with wrong provenance ref: %v", err)
	}
	if result.Passed {
		t.Fatalf("high-touch compile synthesis passed without synthesis provenance ref: %+v", result)
	}
	workflowMetrics.ProvenanceEventRefIDs = []string{"synthesis:" + synthesisDocID}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: highTouchCompileSynthesisNaturalScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify high-touch compile synthesis with provenance: %v", err)
	}
	if !result.Passed {
		t.Fatalf("high-touch compile synthesis failed with provenance events: %+v", result)
	}
}

func TestVerifyCompileSynthesisResponseCandidateRequiresContractAndWorkflow(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: compileSynthesisResponseCandidateScenarioID}); err != nil {
		t.Fatalf("seed compile synthesis candidate scenario: %v", err)
	}
	replaceSeedSection(t, ctx, paths, synthesisCompilePath, "Summary", "Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient.\n\nCurrent source: "+synthesisCompileCurrentSrc+"\n\nSuperseded source: "+synthesisCompileOldSrc)
	synthesisDocID, found, err := documentIDByPath(ctx, paths, synthesisCompilePath)
	if err != nil {
		t.Fatalf("lookup synthesis doc id: %v", err)
	}
	if !found {
		t.Fatalf("missing synthesis doc id for %s", synthesisCompilePath)
	}
	workflowMetrics := metrics{
		AssistantCalls:        1,
		SearchUsed:            true,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		ProjectionStatesUsed:  true,
		ProvenanceEventsUsed:  true,
		ProvenanceEventRefIDs: []string{"synthesis:" + synthesisDocID},
		ReplaceSectionUsed:    true,
		EventTypeCounts:       map[string]int{},
		CommandExecutions:     6,
		ToolCalls:             6,
	}
	candidateAnswer := "```json\n{\"selected_path\":\"" + synthesisCompilePath + "\",\"existing_candidate\":true,\"source_refs\":[\"" + synthesisCompileCurrentSrc + "\",\"" + synthesisCompileOldSrc + "\"],\"source_evidence\":\"Current source " + synthesisCompileCurrentSrc + "; superseded source " + synthesisCompileOldSrc + "\",\"candidate_status\":\"selected " + synthesisCompilePath + " instead of decoy " + synthesisCompileDecoyPath + "\",\"duplicate_status\":\"exactly one target; no duplicate synthesis page was created\",\"provenance_refs\":[\"synthesis:" + synthesisDocID + "\",\"projection\",\"runner-owned no-bypass\"],\"projection_freshness\":\"fresh synthesis projection for " + synthesisCompilePath + "\",\"write_status\":\"updated with replace_section\",\"validation_boundaries\":\"no direct SQLite, no direct vault inspection, no direct file edits, no broad repo search, no source-built runner, no unsupported actions\",\"authority_limits\":\"canonical source docs and promoted records outrank synthesis; this eval-only response does not implement compile_synthesis\"}\n```"
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: compileSynthesisResponseCandidateScenarioID}, 1, candidateAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify compile synthesis candidate: %v", err)
	}
	if !result.Passed {
		t.Fatalf("compile synthesis candidate failed: %+v", result)
	}
	proseWrappedAnswer := "Candidate:\n" + candidateAnswer
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: compileSynthesisResponseCandidateScenarioID}, 1, proseWrappedAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify prose-wrapped compile synthesis candidate: %v", err)
	}
	if result.Passed {
		t.Fatalf("compile synthesis candidate passed with prose outside JSON fence: %+v", result)
	}
	missingProjectionAnswer := strings.Replace(candidateAnswer, "fresh synthesis projection", "synthesis state", 1)
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: compileSynthesisResponseCandidateScenarioID}, 1, missingProjectionAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify compile synthesis candidate missing projection freshness: %v", err)
	}
	if result.Passed {
		t.Fatalf("compile synthesis candidate passed without projection freshness: %+v", result)
	}
	extraFieldAnswer := strings.Replace(candidateAnswer, "}\n```", ",\"unexpected\":\"field\"}\n```", 1)
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: compileSynthesisResponseCandidateScenarioID}, 1, extraFieldAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify compile synthesis candidate with extra field: %v", err)
	}
	if result.Passed {
		t.Fatalf("compile synthesis candidate passed with an extra response field: %+v", result)
	}
	missingSourceRoleAnswer := strings.Replace(candidateAnswer, "superseded source "+synthesisCompileOldSrc, "older source "+synthesisCompileOldSrc, 1)
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: compileSynthesisResponseCandidateScenarioID}, 1, missingSourceRoleAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify compile synthesis candidate missing source role: %v", err)
	}
	if result.Passed {
		t.Fatalf("compile synthesis candidate passed without both source roles: %+v", result)
	}
	weakProvenanceAnswer := strings.Replace(candidateAnswer, "\"provenance_refs\":[\"synthesis:"+synthesisDocID+"\",\"projection\",\"runner-owned no-bypass\"]", "\"provenance_refs\":[\"projection\"]", 1)
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: compileSynthesisResponseCandidateScenarioID}, 1, weakProvenanceAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify compile synthesis candidate with weak provenance refs: %v", err)
	}
	if result.Passed {
		t.Fatalf("compile synthesis candidate passed without concrete provenance refs: %+v", result)
	}
	missingProvenanceMetrics := workflowMetrics
	missingProvenanceMetrics.ProvenanceEventRefIDs = []string{"unrelated-ref"}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: compileSynthesisResponseCandidateScenarioID}, 1, candidateAnswer, missingProvenanceMetrics)
	if err != nil {
		t.Fatalf("verify compile synthesis candidate missing provenance metrics: %v", err)
	}
	if result.Passed {
		t.Fatalf("compile synthesis candidate passed without synthesis provenance metrics: %+v", result)
	}
}

func TestVerifySourceSensitiveConflictRequiresUnresolvedExplanationAndNoSynthesis(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}); err != nil {
		t.Fatalf("seed source audit conflict scenario: %v", err)
	}
	alphaID, alphaFound, err := documentIDByPath(ctx, paths, sourceAuditConflictAlphaPath)
	if err != nil {
		t.Fatalf("lookup alpha id: %v", err)
	}
	bravoID, bravoFound, err := documentIDByPath(ctx, paths, sourceAuditConflictBravoPath)
	if err != nil {
		t.Fatalf("lookup bravo id: %v", err)
	}
	if !alphaFound || !bravoFound {
		t.Fatalf("missing conflict source ids: alpha=%v bravo=%v", alphaFound, bravoFound)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	finalAnswer := sourceAuditConflictAlphaPath + " says seven days; " + sourceAuditConflictBravoPath + " says thirty days. Both are current sources. This conflict is unresolved because there is no supersession metadata, so I cannot choose a winner without source authority."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, noTools)
	if err != nil {
		t.Fatalf("verify source audit conflict no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without runner workflow: %+v", result)
	}
	workflowMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ProvenanceEventsUsed: true,
		EventTypeCounts:      map[string]int{},
		CommandExecutions:    3,
		ToolCalls:            3,
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict missing provenance refs: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without both provenance refs: %+v", result)
	}
	workflowMetrics.ProvenanceEventRefIDs = []string{alphaID, bravoID}
	answerWithoutCurrentSources := sourceAuditConflictAlphaPath + " says seven days; " + sourceAuditConflictBravoPath + " says thirty days. This conflict is unresolved because there is no supersession metadata, so I cannot choose a winner without source authority."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, answerWithoutCurrentSources, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict missing current-source wording: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed without current-source wording: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source audit conflict failed: %+v", result)
	}

	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	if err := createSeedDocument(ctx, cfg, "synthesis/audit-conflict.md", "Audit Conflict", "# Audit Conflict\n"); err != nil {
		t.Fatalf("create forbidden conflict synthesis: %v", err)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: sourceAuditConflictScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify source audit conflict with synthesis: %v", err)
	}
	if result.Passed {
		t.Fatalf("source audit conflict passed after creating synthesis: %+v", result)
	}
}

func TestVerifyBroadContradictionAuditRevisitRequiresRepairConflictAndDecision(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: broadAuditScriptedScenarioID}); err != nil {
		t.Fatalf("seed broad audit scenario: %v", err)
	}
	alphaID, alphaFound, err := documentIDByPath(ctx, paths, sourceAuditConflictAlphaPath)
	if err != nil {
		t.Fatalf("lookup alpha id: %v", err)
	}
	bravoID, bravoFound, err := documentIDByPath(ctx, paths, sourceAuditConflictBravoPath)
	if err != nil {
		t.Fatalf("lookup bravo id: %v", err)
	}
	if !alphaFound || !bravoFound {
		t.Fatalf("missing conflict source ids: alpha=%v bravo=%v", alphaFound, bravoFound)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: broadAuditScriptedScenarioID}, 1, "Updated "+sourceAuditSynthesisPath+".", noTools)
	if err != nil {
		t.Fatalf("verify broad audit before repair: %v", err)
	}
	if result.Passed {
		t.Fatalf("broad audit passed before repair: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, sourceAuditSynthesisPath, "Summary", "Current audit guidance: use the installed openclerk JSON runner.\n\nCurrent source: "+sourceAuditCurrentSourcePath+"\n\nSuperseded source: "+sourceAuditOldSourcePath)
	replaceSeedSection(t, ctx, paths, sourceAuditSynthesisPath, "Freshness", "Checked provenance events and synthesis projection freshness after the current source update.")
	workflowMetrics := metrics{
		AssistantCalls:        1,
		SearchUsed:            true,
		ListDocumentsUsed:     true,
		GetDocumentUsed:       true,
		ProjectionStatesUsed:  true,
		ProvenanceEventsUsed:  true,
		ReplaceSectionUsed:    true,
		ProvenanceEventRefIDs: []string{alphaID, bravoID},
		EventTypeCounts:       map[string]int{},
		CommandExecutions:     8,
		ToolCalls:             8,
	}
	finalAnswer := "Search found source paths and citations. Updated " + sourceAuditSynthesisPath + " from " + sourceAuditCurrentSourcePath + "; provenance and projection freshness are fresh. " + sourceAuditConflictAlphaPath + " says seven days; " + sourceAuditConflictBravoPath + " says thirty days. Both sources are current, the conflict is unresolved because there is no source authority, and I cannot choose a winner. Neither a capability gap nor an ergonomics gap is proven. Current primitives can express the workflow safely, the UX is acceptable enough, and the decision is keep broad contradiction/audit reference/deferred, do not promote a broad semantic contradiction engine."
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: broadAuditScriptedScenarioID}, 1, finalAnswer, workflowMetrics)
	if err != nil {
		t.Fatalf("verify broad audit after repair: %v", err)
	}
	if !result.Passed {
		t.Fatalf("broad audit failed after repair: %+v", result)
	}
}

func TestVerifySynthesisSourceSetPressureRequiresAllSources(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}); err != nil {
		t.Fatalf("seed source set pressure scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/source-set-alpha.md, sources/source-set-beta.md, sources/source-set-gamma.md
---
# Compiler Source Set

## Summary
Alpha, beta, and gamma source refs show the synthesis compiler pressure workflow can preserve freshness.

## Sources
- sources/source-set-alpha.md
- sources/source-set-beta.md
- sources/source-set-gamma.md

## Freshness
Checked runner search results and synthesis candidate listing for all source refs.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisSourceSetPath, "Compiler Source Set", body); err != nil {
		t.Fatalf("create source set synthesis: %v", err)
	}
	completeMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	finalAnswer := "Created " + synthesisSourceSetPath + " with alpha, beta, and gamma sources."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}, 1, finalAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify source set pressure: %v", err)
	}
	if !result.Passed {
		t.Fatalf("source set pressure failed: %+v", result)
	}

	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: synthesisSourceSetPressureScenarioID}, 1, finalAnswer, metrics{AssistantCalls: 1, SearchUsed: true, EventTypeCounts: map[string]int{}})
	if err != nil {
		t.Fatalf("verify missing list metric: %v", err)
	}
	if result.Passed {
		t.Fatalf("source set pressure passed without candidate listing metric: %+v", result)
	}
}

func TestVerifyMTSynthesisDriftPressureRequiresSourceUpdateAndRepair(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	sc := requireScenarioByID(t, mtSynthesisDriftPressureScenarioID)
	if err := seedScenario(ctx, paths, sc); err != nil {
		t.Fatalf("seed drift pressure scenario: %v", err)
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	initialBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/drift-current.md, sources/drift-old.md
---
# Drift Runner

## Summary
Initial drift synthesis says the decision is still under review.

## Sources
- sources/drift-current.md
- sources/drift-old.md

## Freshness
Checked initial source refs.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, mtDriftSynthesisPath, "Drift Runner", initialBody); err != nil {
		t.Fatalf("create initial drift synthesis: %v", err)
	}
	turnOneMetrics := metrics{
		AssistantCalls:    1,
		SearchUsed:        true,
		ListDocumentsUsed: true,
		EventTypeCounts:   map[string]int{},
	}
	result, err := verifyScenarioTurn(ctx, paths, sc, 1, "Created "+mtDriftSynthesisPath+".", turnOneMetrics)
	if err != nil {
		t.Fatalf("verify drift turn one: %v", err)
	}
	if !result.Passed {
		t.Fatalf("drift turn one failed: %+v", result)
	}

	replaceSeedSection(t, ctx, paths, mtDriftCurrentPath, "Summary", "Current drift decision says existing document and retrieval actions should stay the v1 synthesis path.")
	replaceSeedSection(t, ctx, paths, mtDriftSynthesisPath, "Summary", "Current drift decision: keep existing document and retrieval actions.\n\nCurrent source: "+mtDriftCurrentPath+"\n\nSuperseded source: "+mtDriftOldSourcePath)
	replaceSeedSection(t, ctx, paths, mtDriftSynthesisPath, "Freshness", "Checked synthesis projection freshness after the current source update.")
	turnTwoMetrics := metrics{
		AssistantCalls:       1,
		SearchUsed:           true,
		ListDocumentsUsed:    true,
		GetDocumentUsed:      true,
		ProjectionStatesUsed: true,
		EventTypeCounts:      map[string]int{},
	}
	finalAnswer := "Updated " + mtDriftSynthesisPath + " from " + mtDriftCurrentPath + "; final freshness is fresh."
	result, err = verifyScenarioTurn(ctx, paths, sc, 2, finalAnswer, turnTwoMetrics)
	if err != nil {
		t.Fatalf("verify drift turn two: %v", err)
	}
	if !result.Passed {
		t.Fatalf("drift turn two failed: %+v", result)
	}
}

func TestVerifyPromotedRecordVsDocsRequiresComparisonAnswer(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "promoted-record-vs-docs"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	withRetrievalWork := metrics{AssistantCalls: 1, ToolCalls: 2, CommandExecutions: 2, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "Services lookup says JSON runner; plain docs search agrees.", noTools)
	if err != nil {
		t.Fatalf("verify records vs docs no tools: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-tool records vs docs passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "Services lookup says JSON runner; plain docs search agrees.", withRetrievalWork)
	if err != nil {
		t.Fatalf("verify records vs docs: %v", err)
	}
	if !result.Passed {
		t.Fatalf("records vs docs failed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: "promoted-record-vs-docs"}, 1, "JSON runner.", withRetrievalWork)
	if err != nil {
		t.Fatalf("verify incomplete records vs docs answer: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete comparison passed: %+v", result)
	}
}

func TestVerifyDecisionRecordVsDocsRequiresTypedLookup(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}); err != nil {
		t.Fatalf("seed decision scenario: %v", err)
	}
	noTypedLookup := metrics{AssistantCalls: 1, SearchUsed: true, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, SearchUsed: true, DecisionsLookupUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "Plain docs search agrees, but decisions lookup filters status and scope for the accepted AgentOps JSON runner decision."
	completeAnswer := noCitationAnswer + " The decision citation path is docs/architecture/runner-current-decision.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, completeAnswer, noTypedLookup)
	if err != nil {
		t.Fatalf("verify decision no typed lookup: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-typed decision comparison passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify decision no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation decision comparison passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRecordVsDocsScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify decision comparison: %v", err)
	}
	if !result.Passed {
		t.Fatalf("decision comparison failed: %+v", result)
	}
}

func TestVerifyDecisionSupersessionFreshnessRequiresProjectionAndProvenance(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionSupersessionScenarioID}); err != nil {
		t.Fatalf("seed supersession scenario: %v", err)
	}
	noProjection := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old", "adr-runner-current"}, EventTypeCounts: map[string]int{}}
	incompleteDecisionRecord := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-runner-old", "adr-runner-current"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "adr-runner-old is superseded and stale; adr-runner-current supersedes it and is fresh, with provenance and projection evidence."
	completeAnswer := noCitationAnswer + " Citation paths: docs/architecture/runner-old-decision.md and records/decisions/runner-current-decision.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, noProjection)
	if err != nil {
		t.Fatalf("verify supersession no projection: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-projection supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, incompleteDecisionRecord)
	if err != nil {
		t.Fatalf("verify supersession incomplete decision record ids: %v", err)
	}
	if result.Passed {
		t.Fatalf("incomplete decision record ids supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify supersession no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation supersession passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionSupersessionScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify supersession: %v", err)
	}
	if !result.Passed {
		t.Fatalf("supersession failed: %+v", result)
	}
}

func TestVerifyDecisionRealADRMigrationRequiresDecisionProjectionEvidence(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}); err != nil {
		t.Fatalf("seed real adr scenario: %v", err)
	}
	noProjection := metrics{AssistantCalls: 1, DecisionsLookupUsed: true, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-agentops-only-knowledge-plane"}, EventTypeCounts: map[string]int{}}
	completeMetrics := metrics{AssistantCalls: 1, DecisionsLookupUsed: true, DecisionRecordUsed: true, DecisionRecordIDs: []string{"adr-agentops-only-knowledge-plane"}, ProjectionStatesUsed: true, ProvenanceEventsUsed: true, EventTypeCounts: map[string]int{}}
	noCitationAnswer := "Canonical markdown ADRs remain authoritative; decisions_lookup and decision_record return derived decision records with fresh projection and provenance evidence."
	completeAnswer := noCitationAnswer + " Citation paths: docs/architecture/eval-backed-knowledge-plane-adr.md and docs/architecture/knowledge-configuration-v1-adr.md."
	result, err := verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, completeAnswer, noProjection)
	if err != nil {
		t.Fatalf("verify real adr no projection: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-projection real adr migration passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, noCitationAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify real adr no citation: %v", err)
	}
	if result.Passed {
		t.Fatalf("no-citation real adr migration passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, scenario{ID: decisionRealADRMigrationScenarioID}, 1, completeAnswer, completeMetrics)
	if err != nil {
		t.Fatalf("verify real adr migration: %v", err)
	}
	if !result.Passed {
		t.Fatalf("real adr migration failed: %+v", result)
	}
}

func TestDuplicatePathRejectRequiresAnswerFailure(t *testing.T) {
	ctx := context.Background()
	paths := scenarioPaths(t.TempDir())
	if err := seedScenario(ctx, paths, scenario{ID: "duplicate-path-reject"}); err != nil {
		t.Fatalf("seed scenario: %v", err)
	}
	sc := scenario{ID: "duplicate-path-reject"}
	noTools := metrics{AssistantCalls: 1, EventTypeCounts: map[string]int{}}
	result, err := verifyScenarioTurn(ctx, paths, sc, 1, "Done.", noTools)
	if err != nil {
		t.Fatalf("verify duplicate no-op: %v", err)
	}
	if result.Passed {
		t.Fatalf("non-rejection answer passed: %+v", result)
	}
	result, err = verifyScenarioTurn(ctx, paths, sc, 1, "notes/projects/duplicate.md already exists, so I did not overwrite it.", noTools)
	if err != nil {
		t.Fatalf("verify duplicate rejection: %v", err)
	}
	if !result.Passed {
		t.Fatalf("duplicate rejection failed: %+v", result)
	}
}
