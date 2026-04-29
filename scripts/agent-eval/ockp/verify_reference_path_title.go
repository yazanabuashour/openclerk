package main

import (
	"context"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func verifyAgentChosenPathProposal(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	count, err := exactDocumentCount(ctx, paths, agentChosenProposalPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if count != 0 {
		failures = append(failures, "path proposal scenario created "+agentChosenProposalPath)
	}
	if turnMetrics.ToolCalls != 0 || turnMetrics.CommandExecutions != 0 {
		failures = append(failures, fmt.Sprintf("path proposal should ask before writing without tools, got tools=%d commands=%d", turnMetrics.ToolCalls, turnMetrics.CommandExecutions))
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenProposalPath, agentChosenURLHarness, agentChosenURLPromptGuidance}) &&
		messageContainsAny(finalMessage, []string{"confirm", "confirmation", "approve", "proceed", "before creating", "before i create"}) &&
		messageContainsAny(finalMessage, []string{"no document was created", "did not create", "not create", "nothing was created"}) &&
		messageContainsAny(finalMessage, []string{"no unsupported runner action", "unsupported runner action is implied", "no unsupported action"})
	if !assistantPass {
		failures = append(failures, "final answer did not propose the stable path and ask for confirmation before creating")
	}
	databasePass := count == 0
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls == 0 && turnMetrics.CommandExecutions == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{agentChosenProposalPath},
	}, nil
}

func verifyAgentChosenExplicitFields(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, agentChosenExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, agentChosenExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/")
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: note",
		"Path policy: explicit fields required.",
		"Title policy: explicit title wins.",
		"Document type policy: explicit type wins.",
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+agentChosenExplicitPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", agentChosenExplicitPath, exactCount))
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous source documents for explicit-fields scenario, got %d", sourcesCount))
	}
	if synthesisCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous synthesis documents for explicit-fields scenario, got %d", synthesisCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit-fields document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenExplicitPath}) &&
		messageContainsAny(finalMessage, []string{"Explicit Fields Path Title Type", "explicit title", "title"}) &&
		messageContainsAny(finalMessage, []string{"explicit", "provided", "user-specified"})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit path/title/type handling")
	}
	databasePass := found &&
		exactCount == 1 &&
		len(missingRequired(body, required)) == 0 &&
		sourcesCount == 0 &&
		synthesisCount == 0
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{agentChosenExplicitPath},
	}, nil
}

func verifyAgentChosenAutonomousPlacement(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, agentChosenAutonomousPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, agentChosenAutonomousPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := documentCountWithPrefix(ctx, paths, "sources/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: source",
		agentChosenURLHarness,
		agentChosenURLPromptGuidance,
		"Path policy: autonomous create then report",
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+agentChosenAutonomousPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", agentChosenAutonomousPath, exactCount))
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one autonomous source document, got %d", sourceCount))
	}
	failures = append(failures, missingRequired(body, required)...)
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenAutonomousPath}) &&
		messageContainsAny(finalMessage, []string{"created", "wrote", "filed"})
	if !assistantPass {
		failures = append(failures, "final answer did not report the chosen autonomous path")
	}
	databasePass := found && exactCount == 1 && sourceCount == 1 && len(missingRequired(body, required)) == 0
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{agentChosenAutonomousPath},
	}, nil
}

func verifyAgentChosenSynthesisPathSelection(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, agentChosenSynthesisPath, finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:              []string{agentChosenSynthesisAlphaPath, agentChosenSynthesisBetaPath, agentChosenSynthesisGammaPath},
		RequireSearch:           true,
		RequireList:             true,
		RequireProjectionStates: true,
		Metrics:                 turnMetrics,
		FinalAnswerPath:         true,
		AdditionalDocs:          []string{agentChosenSynthesisAlphaPath, agentChosenSynthesisBetaPath, agentChosenSynthesisGammaPath},
		AdditionalBodyRequirements: []string{
			"explicit-path compatibility",
			"metadata",
			"freshness",
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if synthesisCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one chosen synthesis document, got %d", synthesisCount))
	}
	databasePass := base.DatabasePass && synthesisCount == 1
	assistantPass := base.AssistantPass && len(agentChosenBypassFailures(turnMetrics)) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyAgentChosenAmbiguousDocumentType(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	docPath, body, found, err := documentContaining(ctx, paths, "decision_id: "+agentChosenAmbiguousDecisionID)
	if err != nil {
		return verificationResult{}, err
	}
	decision, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: agentChosenAmbiguousDecisionID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      agentChosenAmbiguousDecisionID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"decision_id: " + agentChosenAmbiguousDecisionID,
		"decision_status: accepted",
		"decision_scope: document-path-selection",
		"Metadata authority: frontmatter decides document identity.",
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing ambiguous decision document")
	}
	failures = append(failures, missingRequired(body, required)...)
	hasDecision := decision.Decision != nil &&
		decision.Decision.DecisionID == agentChosenAmbiguousDecisionID &&
		decision.Decision.Status == "accepted" &&
		decision.Decision.Scope == "document-path-selection" &&
		len(decision.Decision.Citations) > 0
	if !hasDecision {
		failures = append(failures, "decision_record did not expose metadata-derived decision identity")
	}
	hasProjection := projection.Projections != nil &&
		len(projection.Projections.Projections) == 1 &&
		projection.Projections.Projections[0].Freshness == "fresh"
	if !hasProjection {
		failures = append(failures, "decision projection is not fresh")
	}
	inspectedDecision := decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, agentChosenAmbiguousDecisionID)
	if !inspectedDecision {
		failures = append(failures, "agent did not inspect decision_record for metadata-derived identity")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect decision projection freshness")
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenAmbiguousDecisionID}) &&
		messageContainsAny(finalMessage, []string{"metadata", "frontmatter"}) &&
		messageContainsAny(finalMessage, []string{"not filename", "not the filename", "not path", "not the path", "not filename/path"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness", "fresh"}) &&
		docPath != "" && messageContainsAll(finalMessage, []string{docPath})
	if !assistantPass {
		failures = append(failures, "final answer did not report chosen path and metadata authority")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && hasDecision && hasProjection
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && inspectedDecision && turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}

func verifyAgentChosenUserPathInstructions(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, agentChosenUserSpecifiedPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/")
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"User path instruction wins.",
		"Do not override explicit path instructions.",
	}
	failures := agentChosenBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+agentChosenUserSpecifiedPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous sources when user path wins, got %d", sourcesCount))
	}
	if synthesisCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous synthesis when user path wins, got %d", synthesisCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit-path document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{agentChosenUserSpecifiedPath})
	if !assistantPass {
		failures = append(failures, "final answer did not mention explicit user path")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && sourcesCount == 0 && synthesisCount == 0
	activityPass := len(agentChosenBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{agentChosenUserSpecifiedPath},
	}, nil
}

func verifyPathTitleURLOnlyAutonomy(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, pathTitleURLOnlyPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, pathTitleURLOnlyPath)
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	required := []string{
		"type: source",
		agentChosenURLHarness,
		agentChosenURLPromptGuidance,
		"Path/title policy: autonomy pressure create then report.",
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+pathTitleURLOnlyPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", pathTitleURLOnlyPath, exactCount))
	}
	if found && title != pathTitleURLOnlyTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", pathTitleURLOnlyTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create URL-only source through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{pathTitleURLOnlyPath}) &&
		messageContainsAny(finalMessage, []string{pathTitleURLOnlyTitle, "harness", "prompt guidance"})
	if !assistantPass {
		failures = append(failures, "final answer did not report chosen path/title")
	}
	databasePass := found && exactCount == 1 && title == pathTitleURLOnlyTitle && len(missingRequired(body, required)) == 0
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{pathTitleURLOnlyPath},
	}, nil
}

func verifyPathTitleMultiSourceDuplicate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, pathTitleSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, pathTitleSynthesisDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current path/title autonomy guidance: update existing synthesis candidate.",
		"## Sources",
		"## Freshness",
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+pathTitleSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, []string{pathTitleSynthesisAlphaPath, pathTitleSynthesisBetaPath})...)
	if duplicateCount != 0 {
		failures = append(failures, "created duplicate synthesis "+pathTitleSynthesisDuplicatePath)
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not inspect existing synthesis before update")
	}
	assistantPass := messageContainsAll(finalMessage, []string{pathTitleSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "repaired", "existing"}) &&
		messageContainsAny(finalMessage, []string{"no duplicate", "avoided duplicate", "not create a duplicate"})
	if !assistantPass {
		failures = append(failures, "final answer did not report existing synthesis update and duplicate avoidance")
	}
	databasePass := found &&
		duplicateCount == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, []string{pathTitleSynthesisAlphaPath, pathTitleSynthesisBetaPath})) == 0
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 && turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{pathTitleSynthesisPath, pathTitleSynthesisDuplicatePath, pathTitleSynthesisAlphaPath, pathTitleSynthesisBetaPath},
	}, nil
}

func verifyPathTitleExplicitOverrides(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, pathTitleExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/path-title/")
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	required := []string{
		"type: note",
		"Explicit path/title override wins.",
		"Do not apply autonomous path conventions.",
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+pathTitleExplicitPath)
	}
	if found && title != pathTitleExplicitTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", pathTitleExplicitTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no autonomous path-title source docs, got %d", sourcesCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit override document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{pathTitleExplicitPath, pathTitleExplicitTitle})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit override path and title")
	}
	databasePass := found && title == pathTitleExplicitTitle && len(missingRequired(body, required)) == 0 && sourcesCount == 0
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{pathTitleExplicitPath},
	}, nil
}

func verifyPathTitleDuplicateRisk(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingCount, err := exactDocumentCount(ctx, paths, pathTitleDuplicateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	pathTitleSourceCount, err := documentCountWithPrefix(ctx, paths, "sources/path-title/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing source %s, got %d", pathTitleDuplicateExistingPath, existingCount))
	}
	if pathTitleSourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected only the seeded path-title source document, got %d", pathTitleSourceCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search for duplicate risk")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source candidates")
	}
	assistantPass := messageContainsAll(finalMessage, []string{pathTitleDuplicateExistingPath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "reuse"}) &&
		messageContainsAny(finalMessage, []string{"not create", "did not create", "no new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate risk and no-create outcome")
	}
	databasePass := existingCount == 1 && pathTitleSourceCount == 1
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 && turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{pathTitleDuplicateExistingPath, pathTitleDuplicateCandidatePath},
	}, nil
}

func verifyPathTitleMetadataAuthority(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	docPath, body, found, err := documentContaining(ctx, paths, "decision_id: "+pathTitleMetadataDecisionID)
	if err != nil {
		return verificationResult{}, err
	}
	decision, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: pathTitleMetadataDecisionID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      pathTitleMetadataDecisionID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"decision_id: " + pathTitleMetadataDecisionID,
		"decision_title: " + pathTitleMetadataTitle,
		"decision_status: accepted",
		"Metadata authority: frontmatter decides path/title identity.",
	}
	failures := pathTitleBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing path/title metadata authority decision")
	}
	failures = append(failures, missingRequired(body, required)...)
	hasDecision := decision.Decision != nil &&
		decision.Decision.DecisionID == pathTitleMetadataDecisionID &&
		decision.Decision.Status == "accepted" &&
		len(decision.Decision.Citations) > 0
	if !hasDecision {
		failures = append(failures, "decision_record did not expose metadata authority decision")
	}
	hasProjection := projection.Projections != nil &&
		len(projection.Projections.Projections) == 1 &&
		projection.Projections.Projections[0].Freshness == "fresh"
	if !hasProjection {
		failures = append(failures, "decision projection is not fresh")
	}
	if !decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, pathTitleMetadataDecisionID) {
		failures = append(failures, "agent did not inspect decision_record")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection_states")
	}
	assistantPass := docPath != "" &&
		messageContainsAll(finalMessage, []string{docPath, pathTitleMetadataDecisionID}) &&
		messageContainsAny(finalMessage, []string{"metadata", "frontmatter"}) &&
		messageContainsAny(finalMessage, []string{"not filename", "not path", "not filename/path"}) &&
		messageContainsAny(finalMessage, []string{"fresh", "projection"})
	if !assistantPass {
		failures = append(failures, "final answer did not report metadata authority and projection evidence")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && hasDecision && hasProjection
	activityPass := len(pathTitleBypassFailures(turnMetrics)) == 0 &&
		decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, pathTitleMetadataDecisionID) &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}
