package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"strings"
)

func verifyMemoryRouterSessionObservation(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found || doc == nil {
		failures = append(failures, "missing "+memoryRouterSessionObservationPath)
	} else {
		if doc.Title != memoryRouterSessionObservationTitle {
			failures = append(failures, "expected title "+memoryRouterSessionObservationTitle)
		}
		if strings.TrimSpace(doc.Body) != strings.TrimSpace(memoryRouterSessionObservationBody()) {
			failures = append(failures, "session observation body does not match exact fixture")
		}
	}
	assistantPass := strings.TrimSpace(finalMessage) != ""
	if !assistantPass {
		failures = append(failures, "missing final answer")
	}
	databasePass := found && doc != nil &&
		doc.Title == memoryRouterSessionObservationTitle &&
		strings.TrimSpace(doc.Body) == strings.TrimSpace(memoryRouterSessionObservationBody())
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{memoryRouterSessionObservationPath},
	}, nil
}

func verifyMemoryRouterReference(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	sourceRefs := []string{
		memoryRouterSessionObservationPath,
		memoryRouterTemporalPath,
		memoryRouterFeedbackPath,
		memoryRouterRoutingPath,
	}
	body, found, err := documentBodyByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	sessionDocID, sessionFound, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		return verificationResult{}, err
	}
	temporalDocID, temporalFound, err := documentIDByPath(ctx, paths, memoryRouterTemporalPath)
	if err != nil {
		return verificationResult{}, err
	}
	feedbackDocID, feedbackFound, err := documentIDByPath(ctx, paths, memoryRouterFeedbackPath)
	if err != nil {
		return verificationResult{}, err
	}
	routingDocID, routingFound, err := documentIDByPath(ctx, paths, memoryRouterRoutingPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDocID, synthesisDocIDFound, err := documentIDByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   sessionDocID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, synthesisDocID)
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Temporal status: current canonical docs outrank stale session observations.",
		"Session promotion path: durable canonical markdown with source refs.",
		"Feedback weighting: advisory only.",
		"Routing choice: existing AgentOps document and retrieval actions.",
		"Decision: keep memory and autonomous routing as reference/deferred.",
		"## Sources",
		"## Freshness",
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+memoryRouterSynthesisPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", memoryRouterSynthesisPath, exactCount))
	}
	if !sessionFound {
		failures = append(failures, "missing "+memoryRouterSessionObservationPath)
	}
	if !temporalFound {
		failures = append(failures, "missing "+memoryRouterTemporalPath)
	}
	if !feedbackFound {
		failures = append(failures, "missing "+memoryRouterFeedbackPath)
	}
	if !routingFound {
		failures = append(failures, "missing "+memoryRouterRoutingPath)
	}
	if !synthesisDocIDFound {
		failures = append(failures, "missing document id for "+memoryRouterSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	hasProvenance := sessionFound && provenance.Provenance != nil && len(provenance.Provenance.Events) > 0
	if !hasProvenance {
		failures = append(failures, "session observation provenance missing")
	}
	hasProjection := projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterSessionObservationPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterTemporalPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterFeedbackPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterRoutingPath)
	if !hasProjection {
		failures = append(failures, "memory/router synthesis projection is not fresh with all source refs")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	listedMemoryRouterPrefix := containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{memoryRouterPrefix})
	if !turnMetrics.ListDocumentsUsed || !listedMemoryRouterPrefix {
		failures = append(failures, "agent did not list memory/router reference docs with path prefix")
	}
	requiredGetDocIDs := []string{sessionDocID, temporalDocID, feedbackDocID, routingDocID}
	gotMemoryRouterDocs := containsAllStrings(turnMetrics.GetDocumentDocIDs, requiredGetDocIDs)
	if !turnMetrics.GetDocumentUsed || !gotMemoryRouterDocs {
		failures = append(failures, "agent did not get every canonical memory/router doc")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection freshness")
	}
	if turnMetrics.BroadRepoSearch {
		failures = append(failures, "agent used broad repo search")
	}
	if turnMetrics.DirectSQLiteAccess {
		failures = append(failures, "agent used direct SQLite")
	}
	if turnMetrics.LegacyRunnerUsage {
		failures = append(failures, "agent used source-built or legacy runner path")
	}
	assistantPass := memoryRouterReferenceAnswerPass(finalMessage)
	if !assistantPass {
		failures = append(failures, "final answer did not explain temporal status, session promotion, feedback weighting, routing, source refs, freshness/provenance, and reference/defer decision")
	}

	databasePass := found &&
		exactCount == 1 &&
		sessionFound &&
		temporalFound &&
		feedbackFound &&
		routingFound &&
		synthesisDocIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		hasProvenance &&
		hasProjection
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		listedMemoryRouterPrefix &&
		turnMetrics.GetDocumentUsed &&
		gotMemoryRouterDocs &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.ProjectionStatesUsed &&
		!turnMetrics.BroadRepoSearch &&
		!turnMetrics.DirectSQLiteAccess &&
		!turnMetrics.LegacyRunnerUsage
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     append([]string{memoryRouterSynthesisPath}, sourceRefs...),
	}, nil
}

func verifyMemoryRouterRevisit(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, scripted bool) (verificationResult, error) {
	base, err := verifyMemoryRouterReference(ctx, paths, finalMessage, turnMetrics)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if base.Details != "ok" {
		failures = append(failures, base.Details)
	}
	if turnMetrics.CreateDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.AppendDocumentUsed {
		failures = append(failures, "revisit scenario created or updated documents")
	}
	assistantPass := memoryRouterRevisitAnswerPass(finalMessage, scripted)
	if !assistantPass {
		if scripted {
			failures = append(failures, "final answer did not compare memory/router evidence, current-primitives safety, UX acceptability, capability/ergonomics posture, and reference/defer decision")
		} else {
			failures = append(failures, "final answer did not compare memory/router evidence, capability/ergonomics posture, and reference/defer decision")
		}
	}
	noWrites := !turnMetrics.CreateDocumentUsed && !turnMetrics.ReplaceSectionUsed && !turnMetrics.AppendDocumentUsed
	return verificationResult{
		Passed:        base.DatabasePass && base.AssistantPass && assistantPass && noWrites,
		DatabasePass:  base.DatabasePass,
		AssistantPass: base.AssistantPass && assistantPass && noWrites,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyHighTouchMemoryRouterRecall(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, scripted bool) (verificationResult, error) {
	base, err := verifyMemoryRouterRevisit(ctx, paths, finalMessage, turnMetrics, scripted)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if base.Details != "ok" {
		failures = append(failures, base.Details)
	}
	synthesisDocID, synthesisFound, err := documentIDByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	listedSynthesisPrefix := containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{"synthesis/"})
	if !turnMetrics.ListDocumentsUsed || !listedSynthesisPrefix {
		failures = append(failures, "agent did not list synthesis documents before recall answer")
	}
	gotSynthesis := synthesisFound && containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{synthesisDocID})
	if !turnMetrics.GetDocumentUsed || !gotSynthesis {
		failures = append(failures, "agent did not get memory/router synthesis document")
	}
	assistantPass := highTouchMemoryRouterRecallAnswerPass(finalMessage, scripted)
	if !assistantPass {
		failures = append(failures, "final answer did not cover canonical docs over stale session observations, routing rationale, list/get evidence, local-first/no-bypass boundaries, and capability/UX posture")
	}
	activityPass := listedSynthesisPrefix && gotSynthesis
	return verificationResult{
		Passed:        base.DatabasePass && base.AssistantPass && assistantPass && activityPass,
		DatabasePass:  base.DatabasePass && synthesisFound,
		AssistantPass: base.AssistantPass && assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyMemoryRouterRecallEvidenceOnly(ctx context.Context, paths evalPaths, turnMetrics metrics) (verificationResult, error) {
	base, err := verifyMemoryRouterReferenceEvidence(ctx, paths, turnMetrics)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if base.Details != "ok" {
		failures = append(failures, base.Details)
	}
	synthesisDocID, synthesisFound, err := documentIDByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	listedSynthesisPrefix := containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{"synthesis/"})
	if !turnMetrics.ListDocumentsUsed || !listedSynthesisPrefix {
		failures = append(failures, "agent did not list synthesis documents before recall answer")
	}
	gotSynthesis := synthesisFound && containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{synthesisDocID})
	if !turnMetrics.GetDocumentUsed || !gotSynthesis {
		failures = append(failures, "agent did not get memory/router synthesis document")
	}
	if turnMetrics.CreateDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.AppendDocumentUsed {
		failures = append(failures, "memory/router recall evidence scenario created or updated documents")
	}
	activityPass := base.AssistantPass &&
		listedSynthesisPrefix &&
		gotSynthesis &&
		!turnMetrics.CreateDocumentUsed &&
		!turnMetrics.ReplaceSectionUsed &&
		!turnMetrics.AppendDocumentUsed
	return verificationResult{
		Passed:        base.DatabasePass && activityPass,
		DatabasePass:  base.DatabasePass && synthesisFound,
		AssistantPass: activityPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyMemoryRouterRecallCandidateCurrentPrimitives(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, scripted bool) (verificationResult, error) {
	base, err := verifyMemoryRouterRecallEvidenceOnly(ctx, paths, turnMetrics)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if base.Details != "ok" {
		failures = append(failures, base.Details)
	}
	assistantFailures := memoryRouterRecallCandidateAnswerFailures(finalMessage, scripted)
	failures = append(failures, assistantFailures...)
	bypassFailures := memoryRouterRecallCandidateBypassFailures(turnMetrics)
	failures = append(failures, bypassFailures...)
	assistantPass := len(assistantFailures) == 0
	return verificationResult{
		Passed:        base.DatabasePass && base.AssistantPass && assistantPass && len(bypassFailures) == 0,
		DatabasePass:  base.DatabasePass,
		AssistantPass: base.AssistantPass && assistantPass && len(bypassFailures) == 0,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyMemoryRouterRecallResponseCandidate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifyMemoryRouterRecallEvidenceOnly(ctx, paths, turnMetrics)
	if err != nil {
		return verificationResult{}, err
	}
	sessionDocID, sessionDocIDFound, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if base.Details != "ok" {
		failures = append(failures, base.Details)
	}
	assistantFailures := memoryRouterRecallCandidateObjectFailures(finalMessage, docIDOrEmptyString(sessionDocIDFound, sessionDocID))
	failures = append(failures, assistantFailures...)
	assistantPass := len(assistantFailures) == 0
	return verificationResult{
		Passed:        base.DatabasePass && base.AssistantPass && assistantPass,
		DatabasePass:  base.DatabasePass,
		AssistantPass: base.AssistantPass && assistantPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}

func verifyMemoryRouterRecallReportAction(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionMemoryRouterRecall,
		MemoryRouterRecall: runner.MemoryRouterRecallOptions{
			Query: memoryRouterSearchText,
			Limit: 10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	report := result.MemoryRouterRecall
	failures := []string{}
	databaseFailures := []string{}
	if result.Rejected || report == nil {
		databaseFailures = append(databaseFailures, "memory_router_recall_report did not return a report")
	}
	if report != nil {
		requiredStrings := []string{
			report.QuerySummary,
			report.TemporalStatus,
			report.StaleSessionStatus,
			report.FeedbackWeighting,
			report.RoutingRationale,
			report.SynthesisFreshness,
			report.ValidationBoundaries,
			report.AuthorityLimits,
		}
		for _, value := range requiredStrings {
			if strings.TrimSpace(value) == "" {
				databaseFailures = append(databaseFailures, "memory_router_recall report contains an empty string field")
				break
			}
		}
		for _, ref := range []string{
			memoryRouterSessionObservationPath,
			memoryRouterTemporalPath,
			memoryRouterFeedbackPath,
			memoryRouterRoutingPath,
			memoryRouterSynthesisPath,
		} {
			if !containsAllStrings(report.CanonicalEvidenceRefs, []string{ref}) {
				databaseFailures = append(databaseFailures, "memory_router_recall report missing canonical evidence ref "+ref)
			}
		}
		if len(report.ProvenanceRefs) == 0 {
			databaseFailures = append(databaseFailures, "memory_router_recall report missing provenance refs")
		}
		if !strings.Contains(report.SynthesisFreshness, "fresh synthesis projection") {
			databaseFailures = append(databaseFailures, "memory_router_recall report missing fresh synthesis projection")
		}
		if strings.Contains(report.ValidationBoundaries, "missing evidence") {
			databaseFailures = append(databaseFailures, "memory_router_recall report unexpectedly marked evidence missing")
		}
		if !strings.Contains(report.ValidationBoundaries, "no writes") ||
			!strings.Contains(report.ValidationBoundaries, "no memory transports") ||
			!strings.Contains(report.ValidationBoundaries, "no remember/recall actions") ||
			!strings.Contains(report.ValidationBoundaries, "no autonomous router APIs") ||
			!strings.Contains(report.ValidationBoundaries, "no hidden authority ranking") {
			databaseFailures = append(databaseFailures, "memory_router_recall report missing validation boundaries")
		}
	}
	failures = append(failures, databaseFailures...)
	bypassFailures := memoryRouterRecallCandidateBypassFailures(turnMetrics)
	failures = append(failures, bypassFailures...)
	if turnMetrics.CreateDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.AppendDocumentUsed {
		failures = append(failures, "agent used a mutating document action")
	}
	if !turnMetrics.MemoryRouterRecallReportUsed {
		failures = append(failures, "agent did not use memory_router_recall_report")
	}
	answerFields := []string{
		"query_summary",
		"temporal_status",
		"canonical_evidence_refs",
		"stale_session_status",
		"feedback_weighting",
		"routing_rationale",
		"provenance_refs",
		"synthesis_freshness",
		"validation_boundaries",
		"authority_limits",
	}
	answerLower := strings.ToLower(finalMessage)
	for _, field := range answerFields {
		if !strings.Contains(answerLower, field) {
			failures = append(failures, "final answer missing "+field)
		}
	}
	for _, boundary := range []string{"read-only", "no writes", "no bypass", "no memory transport", "no remember/recall", "no autonomous router", "no hidden authority"} {
		if !strings.Contains(answerLower, boundary) {
			failures = append(failures, "final answer missing boundary "+boundary)
		}
	}
	assistantFailureDetails := missingDetails(failures)
	databasePass := report != nil && len(databaseFailures) == 0
	assistantPass := turnMetrics.MemoryRouterRecallReportUsed &&
		!turnMetrics.CreateDocumentUsed &&
		!turnMetrics.ReplaceSectionUsed &&
		!turnMetrics.AppendDocumentUsed &&
		len(bypassFailures) == 0 &&
		!strings.Contains(assistantFailureDetails, "final answer missing") &&
		!strings.Contains(assistantFailureDetails, "agent did not use")
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents: []string{
			memoryRouterSessionObservationPath,
			memoryRouterTemporalPath,
			memoryRouterFeedbackPath,
			memoryRouterRoutingPath,
			memoryRouterSynthesisPath,
		},
	}, nil
}

func verifyMemoryRouterReferenceEvidence(ctx context.Context, paths evalPaths, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	sourceRefs := []string{
		memoryRouterSessionObservationPath,
		memoryRouterTemporalPath,
		memoryRouterFeedbackPath,
		memoryRouterRoutingPath,
	}
	body, found, err := documentBodyByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	exactCount, err := exactDocumentCount(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	sessionDocID, sessionFound, err := documentIDByPath(ctx, paths, memoryRouterSessionObservationPath)
	if err != nil {
		return verificationResult{}, err
	}
	temporalDocID, temporalFound, err := documentIDByPath(ctx, paths, memoryRouterTemporalPath)
	if err != nil {
		return verificationResult{}, err
	}
	feedbackDocID, feedbackFound, err := documentIDByPath(ctx, paths, memoryRouterFeedbackPath)
	if err != nil {
		return verificationResult{}, err
	}
	routingDocID, routingFound, err := documentIDByPath(ctx, paths, memoryRouterRoutingPath)
	if err != nil {
		return verificationResult{}, err
	}
	synthesisDocID, synthesisDocIDFound, err := documentIDByPath(ctx, paths, memoryRouterSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "document",
			RefID:   sessionDocID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, synthesisDocID)
	if err != nil {
		return verificationResult{}, err
	}

	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Temporal status: current canonical docs outrank stale session observations.",
		"Session promotion path: durable canonical markdown with source refs.",
		"Feedback weighting: advisory only.",
		"Routing choice: existing AgentOps document and retrieval actions.",
		"Decision: keep memory and autonomous routing as reference/deferred.",
		"## Sources",
		"## Freshness",
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+memoryRouterSynthesisPath)
	}
	if exactCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one %s document, got %d", memoryRouterSynthesisPath, exactCount))
	}
	if !sessionFound {
		failures = append(failures, "missing "+memoryRouterSessionObservationPath)
	}
	if !temporalFound {
		failures = append(failures, "missing "+memoryRouterTemporalPath)
	}
	if !feedbackFound {
		failures = append(failures, "missing "+memoryRouterFeedbackPath)
	}
	if !routingFound {
		failures = append(failures, "missing "+memoryRouterRoutingPath)
	}
	if !synthesisDocIDFound {
		failures = append(failures, "missing document id for "+memoryRouterSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, sourceRefs)...)
	hasProvenance := sessionFound && provenance.Provenance != nil && len(provenance.Provenance.Events) > 0
	if !hasProvenance {
		failures = append(failures, "session observation provenance missing")
	}
	hasProjection := projection != nil &&
		projection.Freshness == "fresh" &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterSessionObservationPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterTemporalPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterFeedbackPath) &&
		projectionDetailContains(projection.Details, "current_source_refs", memoryRouterRoutingPath)
	if !hasProjection {
		failures = append(failures, "memory/router synthesis projection is not fresh with all source refs")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	listedMemoryRouterPrefix := containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{memoryRouterPrefix})
	if !turnMetrics.ListDocumentsUsed || !listedMemoryRouterPrefix {
		failures = append(failures, "agent did not list memory/router reference docs with path prefix")
	}
	requiredGetDocIDs := []string{sessionDocID, temporalDocID, feedbackDocID, routingDocID}
	gotMemoryRouterDocs := containsAllStrings(turnMetrics.GetDocumentDocIDs, requiredGetDocIDs)
	if !turnMetrics.GetDocumentUsed || !gotMemoryRouterDocs {
		failures = append(failures, "agent did not get every canonical memory/router doc")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection freshness")
	}
	bypassFailures := memoryRouterRecallCandidateBypassFailures(turnMetrics)
	failures = append(failures, bypassFailures...)

	databasePass := found &&
		exactCount == 1 &&
		sessionFound &&
		temporalFound &&
		feedbackFound &&
		routingFound &&
		synthesisDocIDFound &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, sourceRefs)) == 0 &&
		hasProvenance &&
		hasProjection
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		listedMemoryRouterPrefix &&
		turnMetrics.GetDocumentUsed &&
		gotMemoryRouterDocs &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.ProjectionStatesUsed &&
		len(bypassFailures) == 0
	return verificationResult{
		Passed:        databasePass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: activityPass,
		Details:       missingDetails(failures),
		Documents:     append([]string{memoryRouterSynthesisPath}, sourceRefs...),
	}, nil
}

func memoryRouterRecallCandidateBypassFailures(turnMetrics metrics) []string {
	failures := populatedBypassFailures(turnMetrics)
	if turnMetrics.ManualHTTPFetch {
		failures = append(failures, "agent used manual HTTP fetch")
	}
	if turnMetrics.BrowserAutomation {
		failures = append(failures, "agent used browser automation")
	}
	return failures
}

func memoryRouterRecallCandidateObjectFailures(finalMessage string, expectedSessionDocID string) []string {
	object, ok := exactFencedJSONObject(finalMessage)
	if !ok {
		return []string{"final answer must be exactly one fenced JSON object and no prose outside it"}
	}
	var fields map[string]any
	if err := json.Unmarshal([]byte(object), &fields); err != nil {
		return []string{"final JSON object did not parse: " + err.Error()}
	}
	expected := []string{
		"query_summary",
		"temporal_status",
		"canonical_evidence_refs",
		"stale_session_status",
		"feedback_weighting",
		"routing_rationale",
		"provenance_refs",
		"synthesis_freshness",
		"validation_boundaries",
		"authority_limits",
	}
	failures := []string{}
	for _, field := range expected {
		if _, found := fields[field]; !found {
			failures = append(failures, "missing candidate field "+field)
		}
	}
	for field := range fields {
		if !containsAllStrings(expected, []string{field}) {
			failures = append(failures, "unexpected candidate field "+field)
		}
	}
	if strings.Contains(strings.ToLower(object), "session_doc_id") {
		failures = append(failures, "candidate JSON did not replace SESSION_DOC_ID with the actual session document id")
	}
	if expectedSessionDocID != "" && !strings.Contains(strings.ToLower(object), "document:"+strings.ToLower(expectedSessionDocID)) {
		failures = append(failures, "candidate JSON missing actual session document provenance ref document:"+expectedSessionDocID)
	}
	normalized := normalizeValidationMessage(object)
	required := []string{
		"current canonical docs over stale session observations",
		"canonical docs outrank stale session observations",
		"session promotion",
		"canonical markdown",
		"source refs",
		"feedback weighting",
		"advisory",
		"routing rationale",
		"existing agentops document and retrieval",
		"provenance",
		"fresh synthesis projection",
		"local-first/no-bypass",
		"no memory transports",
		"no remember/recall actions",
		"no autonomous router apis",
		"no vector stores",
		"no embedding stores",
		"no graph memory",
		"no hidden authority ranking",
		"does not implement or claim an installed memory/router recall action",
	}
	for _, phrase := range required {
		if !strings.Contains(normalized, phrase) {
			failures = append(failures, "candidate JSON missing "+phrase)
		}
	}
	if memoryRouterRecallCandidateClaimsInstalledAction(normalized) {
		failures = append(failures, "candidate JSON claimed an installed memory/router recall action exists")
	}
	return failures
}

func verifyPromotedRecordDomainExpansion(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, scripted bool) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: promotedRecordDomainSearchText, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: promotedRecordDomainPrefix, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	foundPaths := map[string]bool{}
	primaryDocID := ""
	onlyPrefix := true
	for _, doc := range list.Documents {
		if !strings.HasPrefix(doc.Path, promotedRecordDomainPrefix) {
			onlyPrefix = false
		}
		foundPaths[doc.Path] = true
		if doc.Path == promotedRecordDomainPrimaryPath {
			primaryDocID = doc.DocID
		}
	}
	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  primaryDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	if got.Document != nil {
		body = got.Document.Body
	}
	records, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionRecordsLookup,
		Records: runner.RecordLookupOptions{
			Text:       promotedRecordDomainEntityName,
			EntityType: promotedRecordDomainEntityType,
			Limit:      10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	entity, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:   runner.RetrievalTaskActionRecordEntity,
		EntityID: promotedRecordDomainEntityID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "entity",
			RefID:   promotedRecordDomainEntityID,
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "records",
			RefKind:    "entity",
			RefID:      promotedRecordDomainEntityID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	wantedPaths := []string{promotedRecordDomainPrimaryPath, promotedRecordDomainAdjacentPath}
	hasRecord := records.Records != nil &&
		len(records.Records.Entities) == 1 &&
		records.Records.Entities[0].EntityID == promotedRecordDomainEntityID &&
		records.Records.Entities[0].EntityType == promotedRecordDomainEntityType &&
		len(records.Records.Entities[0].Citations) > 0
	hasEntity := entity.Entity != nil &&
		entity.Entity.EntityID == promotedRecordDomainEntityID &&
		entity.Entity.EntityType == promotedRecordDomainEntityType &&
		entity.Entity.Name == promotedRecordDomainEntityName &&
		recordFactContains(entity.Entity, "status", "active") &&
		recordFactContains(entity.Entity, "owner", "platform") &&
		recordFactContains(entity.Entity, "review_cadence", "monthly") &&
		len(entity.Entity.Citations) > 0
	hasProvenance := provenance.Provenance != nil && len(provenance.Provenance.Events) > 0
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh" &&
		projections.Projections.Projections[0].Details["path"] == promotedRecordDomainPrimaryPath

	failures := populatedBypassFailures(turnMetrics)
	if !searchContainsPath(search, promotedRecordDomainPrimaryPath) || !searchResultHasCitations(search) {
		failures = append(failures, "search did not expose cited canonical promoted-record policy evidence")
	}
	for _, path := range wantedPaths {
		if !foundPaths[path] {
			failures = append(failures, "path-prefix listing did not find "+path)
		}
	}
	if !onlyPrefix || len(list.Documents) != len(wantedPaths) {
		failures = append(failures, "path-prefix listing did not stay scoped to promoted record policy fixture")
	}
	if !messageContainsAll(body, []string{"owner is platform", "status active", "review cadence monthly", "citations must stay with canonical markdown"}) {
		failures = append(failures, "get_document did not expose required canonical policy evidence")
	}
	if !hasRecord {
		failures = append(failures, "records_lookup did not expose exactly the promoted policy record with citations")
	}
	if !hasEntity {
		failures = append(failures, "record_entity did not expose policy identity, facts, and citations")
	}
	if !hasProvenance {
		failures = append(failures, "entity provenance missing")
	}
	if !hasProjection {
		failures = append(failures, "records projection state missing or not fresh")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.ListDocumentsUsed || !containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{promotedRecordDomainPrefix}) {
		failures = append(failures, "agent did not list promoted record domain docs with path prefix")
	}
	gotPrimaryDocument := containsAllStrings(turnMetrics.GetDocumentDocIDs, []string{primaryDocID})
	if !turnMetrics.GetDocumentUsed || !gotPrimaryDocument {
		failures = append(failures, "agent did not get canonical promoted record document")
	}
	if !turnMetrics.RecordsLookupUsed {
		failures = append(failures, "agent did not use records_lookup")
	}
	inspectedPromotedEntity := recordEntityIDsInclude(turnMetrics.RecordEntityIDs, promotedRecordDomainEntityID)
	if !turnMetrics.RecordEntityUsed || !inspectedPromotedEntity {
		failures = append(failures, "agent did not use record_entity")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance events")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect records projection freshness")
	}
	if turnMetrics.CreateDocumentUsed || turnMetrics.ReplaceSectionUsed || turnMetrics.AppendDocumentUsed {
		failures = append(failures, "promoted record domain revisit scenario created or updated documents")
	}
	assistantPass := promotedRecordDomainAnswerPass(finalMessage, scripted)
	if !assistantPass {
		failures = append(failures, "final answer did not compare promoted-record evidence, capability/ergonomics posture, and reference/defer decision")
	}

	databasePass := searchContainsPath(search, promotedRecordDomainPrimaryPath) &&
		searchResultHasCitations(search) &&
		allPathsFound(foundPaths, wantedPaths) &&
		onlyPrefix &&
		len(list.Documents) == len(wantedPaths) &&
		messageContainsAll(body, []string{"owner is platform", "status active", "review cadence monthly", "citations must stay with canonical markdown"}) &&
		hasRecord &&
		hasEntity &&
		hasProvenance &&
		hasProjection
	activityPass := len(populatedBypassFailures(turnMetrics)) == 0 &&
		!turnMetrics.CreateDocumentUsed &&
		!turnMetrics.ReplaceSectionUsed &&
		!turnMetrics.AppendDocumentUsed &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		containsAllStrings(turnMetrics.ListDocumentPathPrefixes, []string{promotedRecordDomainPrefix}) &&
		turnMetrics.GetDocumentUsed &&
		gotPrimaryDocument &&
		turnMetrics.RecordsLookupUsed &&
		turnMetrics.RecordEntityUsed &&
		inspectedPromotedEntity &&
		turnMetrics.ProvenanceEventsUsed &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     append([]string{promotedRecordDomainNarrativePath}, wantedPaths...),
	}, nil
}

func recordFactContains(entity *runner.RecordEntity, key string, value string) bool {
	if entity == nil {
		return false
	}
	for _, fact := range entity.Facts {
		if fact.Key == key && fact.Value == value {
			return true
		}
	}
	return false
}
