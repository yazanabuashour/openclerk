package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
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
func verifyAnswerFiling(ctx context.Context, paths evalPaths, finalMessage string) (verificationResult, error) {
	docPath := "synthesis/filed-runner-answer.md"
	body, found, err := documentBodyByPath(ctx, paths, docPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if !found {
		failures = append(failures, "missing "+docPath)
	}
	failures = append(failures, missingRequired(body, []string{
		"sources/answer-filing-runner.md",
		"Durable OpenClerk runner answers should be filed as source-linked markdown",
	})...)
	assistantPass := messageContainsAll(finalMessage, []string{docPath})
	if !assistantPass {
		failures = append(failures, "final answer did not mention "+docPath)
	}
	databasePass := found && len(missingRequired(body, []string{
		"sources/answer-filing-runner.md",
		"Durable OpenClerk runner answers should be filed as source-linked markdown",
	})) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     []string{docPath},
	}, nil
}
func verifyRAGRetrievalBaseline(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	unfiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:  ragSearchText,
			Limit: 5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	pathFiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       ragSearchText,
			PathPrefix: ragPathPrefix,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	metadataFiltered, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	repeatedMetadata, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:          ragSearchText,
			MetadataKey:   ragMetadataKey,
			MetadataValue: ragMetadataValue,
			Limit:         5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	synthesisCount, err := documentCountWithPrefix(ctx, paths, "synthesis/")
	if err != nil {
		return verificationResult{}, err
	}

	failures := []string{}
	unfilteredTop, unfilteredTopFound := topSearchHit(unfiltered)
	pathTop, pathTopFound := topSearchHit(pathFiltered)
	metadataTop, metadataTopFound := topSearchHit(metadataFiltered)
	repeatedTop, repeatedTopFound := topSearchHit(repeatedMetadata)
	if !unfilteredTopFound || searchHitPath(unfilteredTop) != ragCurrentPolicyPath {
		failures = append(failures, "unfiltered search did not rank active RAG source first")
	}
	if !pathTopFound || searchHitPath(pathTop) != ragCurrentPolicyPath {
		failures = append(failures, "path-filtered search did not rank active RAG source first")
	}
	if searchContainsPath(pathFiltered, ragArchivedPolicyPath) {
		failures = append(failures, "path-filtered search included archived source")
	}
	if !metadataTopFound || searchHitPath(metadataTop) != ragCurrentPolicyPath {
		failures = append(failures, "metadata-filtered search did not rank active RAG source first")
	}
	if !searchOnlyContainsPath(metadataFiltered, ragCurrentPolicyPath) {
		failures = append(failures, "metadata-filtered search returned non-active policy sources")
	}
	if !metadataTopFound || !repeatedTopFound || metadataTop.DocID != repeatedTop.DocID || metadataTop.ChunkID != repeatedTop.ChunkID {
		failures = append(failures, "repeated metadata-filtered search changed top doc_id or chunk_id")
	}
	if !metadataTopFound || !searchHitHasCitation(metadataTop) {
		failures = append(failures, "metadata-filtered top hit did not include doc_id, chunk_id, path, and line citation")
	}
	if synthesisCount != 0 {
		failures = append(failures, fmt.Sprintf("retrieval-only baseline created %d synthesis documents", synthesisCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !turnMetrics.SearchUnfilteredUsed {
		failures = append(failures, "agent did not use unfiltered retrieval search")
	}
	if !turnMetrics.SearchPathFilterUsed {
		failures = append(failures, "agent did not use path-prefix retrieval search")
	}
	if !turnMetrics.SearchMetadataFilterUsed {
		failures = append(failures, "agent did not use metadata-filtered retrieval search")
	}

	assistantPass := metadataTopFound &&
		messageContainsAll(finalMessage, []string{ragCurrentPolicyPath, metadataTop.DocID, metadataTop.ChunkID}) &&
		messageContainsAny(finalMessage, []string{"json runner", "openclerk json runner"})
	if !assistantPass {
		failures = append(failures, "final answer did not cite active path, doc_id, chunk_id, and JSON runner policy")
	}
	databasePass := unfilteredTopFound &&
		pathTopFound &&
		metadataTopFound &&
		searchHitPath(unfilteredTop) == ragCurrentPolicyPath &&
		searchHitPath(pathTop) == ragCurrentPolicyPath &&
		searchHitPath(metadataTop) == ragCurrentPolicyPath &&
		!searchContainsPath(pathFiltered, ragArchivedPolicyPath) &&
		searchOnlyContainsPath(metadataFiltered, ragCurrentPolicyPath) &&
		repeatedTopFound &&
		metadataTop.DocID == repeatedTop.DocID &&
		metadataTop.ChunkID == repeatedTop.ChunkID &&
		searchHitHasCitation(metadataTop) &&
		synthesisCount == 0
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.SearchUnfilteredUsed &&
		turnMetrics.SearchPathFilterUsed &&
		turnMetrics.SearchMetadataFilterUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{ragCurrentPolicyPath, ragDecoyPolicyPath, ragArchivedPolicyPath},
	}, nil
}
func verifyDocsNavigationBaseline(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docsNavigationPrefix, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	indexDocID, indexFound := "", false
	policyFound := false
	onlyPrefix := true
	for _, doc := range list.Documents {
		if !strings.HasPrefix(doc.Path, docsNavigationPrefix) {
			onlyPrefix = false
		}
		switch doc.Path {
		case docsNavigationIndexPath:
			indexDocID = doc.DocID
			indexFound = true
		case docsNavigationPolicyPath:
			policyFound = true
		}
	}

	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasHeadings := got.Document != nil && containsAllStrings(got.Document.Headings, []string{"AgentOps Wiki Index", "Summary", "Links", "Limits"})

	links, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasOutgoing := links.Links != nil &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationPolicyPath) &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationArchPath) &&
		documentLinksContainPath(links.Links.Outgoing, docsNavigationOpsPath) &&
		documentLinksHaveCitations(links.Links.Outgoing)
	hasIncoming := links.Links != nil &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationPolicyPath) &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationArchPath) &&
		documentLinksContainPath(links.Links.Incoming, docsNavigationOpsPath) &&
		documentLinksHaveCitations(links.Links.Incoming)

	graph, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraph,
		DocID:  indexDocID,
		Limit:  20,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasGraph := graph.Graph != nil &&
		graphContainsNodeLabels(graph.Graph.Nodes, []string{"AgentOps Wiki Index", "Runner Policy", "Knowledge Plane", "Runner Playbook"}) &&
		graphContainsLinkEdge(graph.Graph.Edges) &&
		graphEdgesHaveCitations(graph.Graph.Edges)

	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "graph",
			RefKind:    "document",
			RefID:      indexDocID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh" &&
		projections.Projections.Projections[0].Details["path"] == docsNavigationIndexPath

	failures := []string{}
	if !indexFound {
		failures = append(failures, "path-prefix listing did not find "+docsNavigationIndexPath)
	}
	if !policyFound {
		failures = append(failures, "path-prefix listing did not find "+docsNavigationPolicyPath)
	}
	if !onlyPrefix || len(list.Documents) != 2 {
		failures = append(failures, "path-prefix listing did not stay scoped to agentops directory")
	}
	if !hasHeadings {
		failures = append(failures, "get_document did not expose expected index headings")
	}
	if !hasOutgoing {
		failures = append(failures, "document_links missing cited outgoing links")
	}
	if !hasIncoming {
		failures = append(failures, "document_links missing cited incoming backlinks")
	}
	if !hasGraph {
		failures = append(failures, "graph_neighborhood missing cited nodes or edges")
	}
	if !hasProjection {
		failures = append(failures, "graph projection state missing or not fresh")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not use list_documents")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not use get_document")
	}
	if !turnMetrics.DocumentLinksUsed {
		failures = append(failures, "agent did not use document_links")
	}
	if !turnMetrics.GraphNeighborhoodUsed {
		failures = append(failures, "agent did not use graph_neighborhood")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect graph projection state")
	}

	assistantPass := messageContainsAny(finalMessage, []string{"directory", "folder", "path-prefix", "path prefix"}) &&
		messageContainsAny(finalMessage, []string{"link", "markdown"}) &&
		messageContainsAny(finalMessage, []string{"backlink", "incoming"}) &&
		messageContainsAny(finalMessage, []string{"graph neighborhood", "graph_neighborhood"}) &&
		messageContainsAny(finalMessage, []string{"sufficient", "enough"}) &&
		messageContainsAny(finalMessage, []string{"fails", "fail", "limits", "not enough"}) &&
		messageContainsAll(finalMessage, []string{docsNavigationIndexPath})
	if !assistantPass {
		failures = append(failures, "final answer did not compare directory, links/backlinks, graph neighborhood, limits, and source path")
	}

	databasePass := indexFound &&
		policyFound &&
		onlyPrefix &&
		len(list.Documents) == 2 &&
		hasHeadings &&
		hasOutgoing &&
		hasIncoming &&
		hasGraph &&
		hasProjection
	activityPass := turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.DocumentLinksUsed &&
		turnMetrics.GraphNeighborhoodUsed &&
		turnMetrics.ProjectionStatesUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{docsNavigationIndexPath, docsNavigationPolicyPath, docsNavigationArchPath, docsNavigationOpsPath},
	}, nil
}
func verifyGraphSemanticsReference(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	return verifyGraphSemanticsWorkflow(ctx, paths, finalMessage, turnMetrics, true, graphSemanticsReferenceAnswerPass(finalMessage), "final answer did not compare search, links/backlinks, graph neighborhood, markdown relationship text, and reference/defer decision")
}
func verifyGraphSemanticsRevisit(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, scripted bool) (verificationResult, error) {
	assistantFailure := "final answer did not compare graph evidence, capability/ergonomics posture, and reference/defer decision"
	if scripted {
		assistantFailure = "final answer did not compare graph evidence, current-primitives safety, UX acceptability, capability/ergonomics posture, and reference/defer decision"
	}
	return verifyGraphSemanticsWorkflow(ctx, paths, finalMessage, turnMetrics, scripted, graphSemanticsRevisitAnswerPass(finalMessage, scripted), assistantFailure)
}
func verifyGraphSemanticsWorkflow(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics, requireListDocuments bool, assistantPass bool, assistantFailure string) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{Text: graphSemanticsSearchText, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: graphSemanticsPrefix, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}

	wantedPaths := []string{graphSemanticsIndexPath, graphSemanticsRoutingPath, graphSemanticsFreshnessPath, graphSemanticsOperationsPath}
	foundPaths := map[string]bool{}
	indexDocID := ""
	onlyPrefix := true
	for _, doc := range list.Documents {
		if !strings.HasPrefix(doc.Path, graphSemanticsPrefix) {
			onlyPrefix = false
		}
		foundPaths[doc.Path] = true
		if doc.Path == graphSemanticsIndexPath {
			indexDocID = doc.DocID
		}
	}

	got, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionGet,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	if got.Document != nil {
		body = got.Document.Body
	}

	links, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDocumentLinks,
		DocID:  indexDocID,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasOutgoing := links.Links != nil &&
		documentLinksContainPath(links.Links.Outgoing, graphSemanticsRoutingPath) &&
		documentLinksContainPath(links.Links.Outgoing, graphSemanticsFreshnessPath) &&
		documentLinksContainPath(links.Links.Outgoing, graphSemanticsOperationsPath) &&
		documentLinksHaveCitations(links.Links.Outgoing)
	hasIncoming := links.Links != nil &&
		documentLinksContainPath(links.Links.Incoming, graphSemanticsRoutingPath) &&
		documentLinksContainPath(links.Links.Incoming, graphSemanticsFreshnessPath) &&
		documentLinksContainPath(links.Links.Incoming, graphSemanticsOperationsPath) &&
		documentLinksHaveCitations(links.Links.Incoming)

	graph, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionGraph,
		DocID:  indexDocID,
		Limit:  20,
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasGraph := graph.Graph != nil &&
		graphContainsNodeLabels(graph.Graph.Nodes, []string{"Graph Semantics Reference", "Routing", "Freshness", "Operations"}) &&
		graphContainsStructuralEdge(graph.Graph.Edges) &&
		graphEdgesHaveCitations(graph.Graph.Edges) &&
		graphEdgesOnlyStructural(graph.Graph.Edges)

	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "graph",
			RefKind:    "document",
			RefID:      indexDocID,
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) == 1 &&
		projections.Projections.Projections[0].Freshness == "fresh" &&
		projections.Projections.Projections[0].Details["path"] == graphSemanticsIndexPath

	failures := []string{}
	if !searchContainsPath(search, graphSemanticsIndexPath) || !searchResultHasCitations(search) {
		failures = append(failures, "search did not expose cited canonical relationship text")
	}
	for _, path := range wantedPaths {
		if !foundPaths[path] {
			failures = append(failures, "path-prefix listing did not find "+path)
		}
	}
	if !onlyPrefix || len(list.Documents) != len(wantedPaths) {
		failures = append(failures, "path-prefix listing did not stay scoped to graph semantics fixture")
	}
	if !messageContainsAll(body, []string{"requires", "supersedes", "related to", "operationalizes"}) {
		failures = append(failures, "get_document did not expose expected relationship words")
	}
	if !hasOutgoing {
		failures = append(failures, "document_links missing cited outgoing relationships")
	}
	if !hasIncoming {
		failures = append(failures, "document_links missing cited incoming backlinks")
	}
	if !hasGraph {
		failures = append(failures, "graph_neighborhood missing cited structural graph context")
	}
	if !hasProjection {
		failures = append(failures, "graph projection state missing or not fresh")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if requireListDocuments && !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not use list_documents")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not use get_document")
	}
	if !turnMetrics.DocumentLinksUsed {
		failures = append(failures, "agent did not use document_links")
	}
	if !turnMetrics.GraphNeighborhoodUsed {
		failures = append(failures, "agent did not use graph_neighborhood")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect graph projection state")
	}

	if !assistantPass {
		failures = append(failures, assistantFailure)
	}

	databasePass := searchContainsPath(search, graphSemanticsIndexPath) &&
		searchResultHasCitations(search) &&
		allPathsFound(foundPaths, wantedPaths) &&
		onlyPrefix &&
		len(list.Documents) == len(wantedPaths) &&
		messageContainsAll(body, []string{"requires", "supersedes", "related to", "operationalizes"}) &&
		hasOutgoing &&
		hasIncoming &&
		hasGraph &&
		hasProjection
	activityPass := turnMetrics.SearchUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.DocumentLinksUsed &&
		turnMetrics.GraphNeighborhoodUsed &&
		turnMetrics.ProjectionStatesUsed
	if requireListDocuments {
		activityPass = activityPass && turnMetrics.ListDocumentsUsed
	}
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     wantedPaths,
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
func verifyDocumentHistoryInspection(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	docID, found, err := documentIDByPath(ctx, paths, documentHistoryPolicyPath)
	if err != nil {
		return verificationResult{}, err
	}
	doc, _, err := documentByPath(ctx, paths, documentHistoryPolicyPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "document", RefID: docID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			RefKind: "document",
			RefID:   docID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasUpdatedBody := doc != nil && strings.Contains(doc.Body, "Current state: lifecycle inspection uses list_documents")
	hasProvenance := provenance.Provenance != nil &&
		eventTypesInclude(provenance.Provenance.Events, "document_created") &&
		eventTypesInclude(provenance.Provenance.Events, "document_updated")
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness != ""
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentHistoryPolicyPath)
	}
	if !hasUpdatedBody {
		failures = append(failures, "history inspection fixture did not expose updated lifecycle text")
	}
	if !hasProvenance {
		failures = append(failures, "document provenance missing created and updated events")
	}
	if !hasProjection {
		failures = append(failures, "document projection state missing or not fresh")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryPolicyPath}) &&
		messageContainsAny(finalMessage, []string{"provenance", "document_updated", "updated"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness", "fresh"}) &&
		messageContainsAny(finalMessage, []string{"existing", "current", "document and retrieval", "runner"})
	if !assistantPass {
		failures = append(failures, "final answer did not report history inspection, provenance, projection freshness, and existing runner workflow")
	}
	databasePass := found && hasUpdatedBody && hasProvenance && hasProjection
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryPolicyPath},
	}, nil
}
func verifyDocumentHistoryDiffReview(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	previous, previousFound, err := documentByPath(ctx, paths, documentHistoryDiffPreviousPath)
	if err != nil {
		return verificationResult{}, err
	}
	current, currentFound, err := documentByPath(ctx, paths, documentHistoryDiffCurrentPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !previousFound || previous == nil {
		failures = append(failures, "missing "+documentHistoryDiffPreviousPath)
	}
	if !currentFound || current == nil {
		failures = append(failures, "missing "+documentHistoryDiffCurrentPath)
	}
	if previous == nil || !strings.Contains(previous.Body, "optional review") {
		failures = append(failures, "previous evidence missing optional review text")
	}
	if current == nil || !strings.Contains(current.Body, "required review") {
		failures = append(failures, "current evidence missing required review text")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance")...)
	pathFailures := invalidRunnerPathFailures("list_documents path_prefix", turnMetrics.ListDocumentPathPrefixes)
	pathFailures = append(pathFailures, exactRunnerPathFailures("list_documents path_prefix", turnMetrics.ListDocumentPathPrefixes, documentHistoryDiffListPrefix)...)
	finalAnswerPathFailures := invalidRunnerPathTextFailures("final answer", finalMessage)
	failures = append(failures, pathFailures...)
	failures = append(failures, finalAnswerPathFailures...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryDiffPreviousPath, documentHistoryDiffCurrentPath}) &&
		messageContainsAny(finalMessage, []string{"optional"}) &&
		messageContainsAny(finalMessage, []string{"required"}) &&
		messageContainsAny(finalMessage, []string{"citation", "cited", "source ref", "source_refs", "source"}) &&
		messageContainsAny(finalMessage, []string{"semantic", "summary"}) &&
		messageContainsAny(finalMessage, []string{"raw diff", "private diff", "do not expose raw", "no raw"}) &&
		len(finalAnswerPathFailures) == 0
	if !assistantPass {
		failures = append(failures, "final answer did not preserve cited semantic diff summary and raw-diff privacy handling")
	}
	databasePass := previousFound && currentFound &&
		previous != nil && current != nil &&
		strings.Contains(previous.Body, "optional review") &&
		strings.Contains(current.Body, "required review")
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 &&
		len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance")) == 0 &&
		len(pathFailures) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryDiffPreviousPath, documentHistoryDiffCurrentPath},
	}, nil
}
func verifyDocumentHistoryRestore(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	target, targetFound, err := documentByPath(ctx, paths, documentHistoryRestoreTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	targetID, _, err := documentIDByPath(ctx, paths, documentHistoryRestoreTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "document", RefID: targetID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projections, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			RefKind: "document",
			RefID:   targetID,
			Limit:   5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	if target != nil {
		body = target.Body
	}
	restored := strings.Contains(body, "Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits.") &&
		!strings.Contains(body, "may bypass review")
	hasProvenance := provenance.Provenance != nil && eventTypesInclude(provenance.Provenance.Events, "document_updated")
	hasProjection := projections.Projections != nil &&
		len(projections.Projections.Projections) > 0 &&
		projections.Projections.Projections[0].Freshness != ""
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !targetFound {
		failures = append(failures, "missing "+documentHistoryRestoreTargetPath)
	}
	if !restored {
		failures = append(failures, "restore target was not restored to accepted lifecycle policy")
	}
	if !hasProvenance {
		failures = append(failures, "restore target provenance missing document update")
	}
	if !hasProjection {
		failures = append(failures, "restore target projection missing or not fresh")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryRestoreTargetPath, documentHistoryRestoreSourcePath}) &&
		messageContainsAny(finalMessage, []string{"restored", "restore", "rollback"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"source", "evidence", "citation"})
	if !assistantPass {
		failures = append(failures, "final answer did not report restore evidence, source, provenance, and projection freshness")
	}
	databasePass := targetFound && restored && hasProvenance && hasProjection
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryRestoreSourcePath, documentHistoryRestoreTargetPath},
	}, nil
}
func verifyDocumentHistoryPendingReview(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	targetBody, targetFound, err := documentBodyByPath(ctx, paths, documentHistoryPendingTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	proposalBody, proposalFound, err := documentBodyByPath(ctx, paths, documentHistoryPendingProposalPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !targetFound {
		failures = append(failures, "missing "+documentHistoryPendingTargetPath)
	}
	if !strings.Contains(targetBody, "Accepted lifecycle policy: source-sensitive durable edits require human review before acceptance.") ||
		strings.Contains(targetBody, "Auto-accept pending change") {
		failures = append(failures, "accepted target changed instead of remaining under review")
	}
	requiredProposal := []string{
		"type: review",
		"status: pending",
		"Review state: pending human review.",
		"Proposed change: Auto-accept pending change only after operator approval.",
		"Target document: notes/history-review/pending-target.md",
	}
	if !proposalFound {
		failures = append(failures, "missing "+documentHistoryPendingProposalPath)
	}
	failures = append(failures, missingRequired(proposalBody, requiredProposal)...)
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryPendingTargetPath, documentHistoryPendingProposalPath}) &&
		messageContainsAny(finalMessage, []string{"pending", "review"}) &&
		messageContainsAny(finalMessage, []string{"not accepted", "not become accepted", "did not change", "no accepted change"}) &&
		messageContainsAny(finalMessage, []string{"human", "operator"})
	if !assistantPass {
		failures = append(failures, "final answer did not report pending review state and unchanged accepted target")
	}
	databasePass := targetFound && proposalFound &&
		strings.Contains(targetBody, "Accepted lifecycle policy: source-sensitive durable edits require human review before acceptance.") &&
		!strings.Contains(targetBody, "Auto-accept pending change") &&
		len(missingRequired(proposalBody, requiredProposal)) == 0
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "list", "get", "provenance")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryPendingTargetPath, documentHistoryPendingProposalPath},
	}, nil
}
func verifyDocumentHistoryStaleSynthesis(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	synthesisID, synthesisFound, err := documentIDByPath(ctx, paths, documentHistoryStaleSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	currentID, currentFound, err := documentIDByPath(ctx, paths, documentHistoryStaleCurrentSourcePath)
	if err != nil {
		return verificationResult{}, err
	}
	projection, err := firstSynthesisProjection(ctx, paths, synthesisID)
	if err != nil {
		return verificationResult{}, err
	}
	sourceEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "source", RefID: currentID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	projectionEvents, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{RefKind: "projection", RefID: "synthesis:" + synthesisID, Limit: 10},
	})
	if err != nil {
		return verificationResult{}, err
	}
	hasProjection := projection != nil &&
		projection.Freshness == "stale" &&
		projectionDetailContains(projection.Details, "stale_source_refs", documentHistoryStaleCurrentSourcePath)
	hasSourceEvents := currentFound && sourceEvents.Provenance != nil &&
		eventTypesInclude(sourceEvents.Provenance.Events, "source_updated")
	hasInvalidation := projectionEvents.Provenance != nil &&
		eventTypesInclude(projectionEvents.Provenance.Events, "projection_invalidated")
	failures := documentHistoryInvariantFailures(turnMetrics)
	if !synthesisFound {
		failures = append(failures, "missing "+documentHistoryStaleSynthesisPath)
	}
	if !currentFound {
		failures = append(failures, "missing "+documentHistoryStaleCurrentSourcePath)
	}
	if !hasProjection {
		failures = append(failures, "synthesis projection is not stale with current source ref")
	}
	if !hasSourceEvents {
		failures = append(failures, "current source provenance missing source update")
	}
	if !hasInvalidation {
		failures = append(failures, "synthesis projection invalidation event missing")
	}
	failures = append(failures, missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")...)
	assistantPass := messageContainsAll(finalMessage, []string{documentHistoryStaleSynthesisPath, documentHistoryStaleCurrentSourcePath}) &&
		messageContainsAny(finalMessage, []string{"stale"}) &&
		messageContainsAny(finalMessage, []string{"projection", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "invalidated", "source_updated", "updated"}) &&
		messageContainsAny(finalMessage, []string{"no repair", "not repair", "did not repair", "without repair"})
	if !assistantPass {
		failures = append(failures, "final answer did not report stale synthesis, provenance/invalidation, and no repair")
	}
	databasePass := synthesisFound && currentFound && hasProjection && hasSourceEvents && hasInvalidation
	activityPass := len(documentHistoryInvariantFailures(turnMetrics)) == 0 && len(missingDocumentHistoryMetrics(turnMetrics, "search", "list", "get", "provenance", "projection")) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentHistoryStaleSynthesisPath, documentHistoryStaleCurrentSourcePath, documentHistoryStaleOldSourcePath},
	}, nil
}
func documentHistoryInvariantFailures(turnMetrics metrics) []string {
	failures := []string{}
	if turnMetrics.BroadRepoSearch {
		failures = append(failures, "agent used broad repo search")
	}
	if turnMetrics.DirectSQLiteAccess {
		failures = append(failures, "agent used direct SQLite")
	}
	if turnMetrics.LegacyRunnerUsage {
		failures = append(failures, "agent used source-built or legacy runner path")
	}
	if turnMetrics.GeneratedFileInspection {
		failures = append(failures, "agent inspected generated files")
	}
	if turnMetrics.ModuleCacheInspection {
		failures = append(failures, "agent inspected module cache")
	}
	return failures
}
func invalidRunnerPathFailures(label string, values []string) []string {
	failures := []string{}
	for _, value := range values {
		if isInvalidRunnerPath(value) {
			failures = append(failures, label+" used non-vault-relative path "+value)
		}
	}
	return failures
}
func exactRunnerPathFailures(label string, values []string, allowed ...string) []string {
	failures := []string{}
	allowedSet := map[string]struct{}{}
	seen := map[string]bool{}
	for _, value := range allowed {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		allowedSet[trimmed] = struct{}{}
		seen[trimmed] = false
	}
	if len(values) == 0 {
		for value := range allowedSet {
			failures = append(failures, label+" missing required path "+value)
		}
		return failures
	}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if _, ok := allowedSet[trimmed]; ok {
			seen[trimmed] = true
			continue
		}
		failures = append(failures, label+" used unexpected path "+value)
	}
	for value, found := range seen {
		if !found {
			failures = append(failures, label+" missing required path "+value)
		}
	}
	return failures
}
func invalidRunnerPathTextFailures(label string, value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	lower := strings.ToLower(normalized)
	if strings.Contains(lower, ".openclerk-eval") ||
		strings.Contains(lower, "/vault/") ||
		strings.Contains(lower, "vault/") ||
		unixAbsolutePathPattern.MatchString(normalized) ||
		windowsDrivePathPattern.MatchString(trimmed) ||
		strings.Contains(trimmed, "\\") {
		return []string{label + " included non-vault-relative path text"}
	}
	return nil
}
func isInvalidRunnerPath(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	lower := strings.ToLower(normalized)
	if strings.Contains(lower, ".openclerk-eval") || strings.Contains(lower, "/vault/") || strings.HasPrefix(lower, "vault/") {
		return true
	}
	if strings.HasPrefix(normalized, "/") || strings.HasPrefix(normalized, "~") {
		return true
	}
	if len(trimmed) >= 3 && ((trimmed[0] >= 'A' && trimmed[0] <= 'Z') || (trimmed[0] >= 'a' && trimmed[0] <= 'z')) && trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/') {
		return true
	}
	return strings.Contains(trimmed, "\\")
}
func missingDocumentHistoryMetrics(turnMetrics metrics, required ...string) []string {
	failures := []string{}
	for _, requirement := range required {
		switch requirement {
		case "search":
			if !turnMetrics.SearchUsed {
				failures = append(failures, "agent did not use retrieval search")
			}
		case "list":
			if !turnMetrics.ListDocumentsUsed {
				failures = append(failures, "agent did not use list_documents")
			}
		case "get":
			if !turnMetrics.GetDocumentUsed {
				failures = append(failures, "agent did not use get_document")
			}
		case "provenance":
			if !turnMetrics.ProvenanceEventsUsed {
				failures = append(failures, "agent did not inspect provenance events")
			}
		case "projection":
			if !turnMetrics.ProjectionStatesUsed {
				failures = append(failures, "agent did not inspect projection states")
			}
		}
	}
	return failures
}
func verifyConfiguredLayoutScenario(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	layoutResult, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionInspectLayout})
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if layoutResult.Layout == nil {
		failures = append(failures, "inspect_layout returned no layout")
	} else if !layoutResult.Layout.Valid {
		failures = append(failures, "seeded configured layout was not valid")
	}
	if !turnMetrics.InspectLayoutUsed {
		failures = append(failures, "agent did not use inspect_layout")
	}
	if !messageContainsAll(finalMessage, []string{"convention", "sources/", "synthesis/", "source_refs"}) ||
		!messageContainsAny(finalMessage, []string{"no committed manifest", "no manifest", "config artifact required: false", "config_artifact_required false"}) {
		failures = append(failures, "answer did not explain convention-first layout and no-manifest decision")
	}
	if !messageReportsLayoutValid(finalMessage) {
		failures = append(failures, "answer did not report the layout as valid")
	}
	return verificationFromFailures(failures, "configured layout inspection passed", []string{"sources/layout-runner.md", "synthesis/layout-runner.md", "records/services/layout-runner.md"})
}
func verifyInvalidLayoutScenario(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	layoutResult, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{Action: runner.DocumentTaskActionInspectLayout})
	if err != nil {
		return verificationResult{}, err
	}
	failures := []string{}
	if layoutResult.Layout == nil {
		failures = append(failures, "inspect_layout returned no layout")
	} else {
		if layoutResult.Layout.Valid {
			failures = append(failures, "seeded invalid layout was reported valid")
		}
		for _, id := range []string{"synthesis_source_refs_resolve", "synthesis_freshness_section", "service_identity_metadata"} {
			if !layoutChecksInclude(layoutResult.Layout.Checks, id, "fail") {
				failures = append(failures, "layout result missing failing check "+id)
			}
		}
	}
	if !turnMetrics.InspectLayoutUsed {
		failures = append(failures, "agent did not use inspect_layout")
	}
	if !messageContainsAll(finalMessage, []string{"synthesis/broken-layout.md", "records/services/broken-layout-service.md"}) ||
		!messageContainsAny(finalMessage, []string{"invalid", "valid: false", "valid false"}) ||
		!messageContainsAny(finalMessage, []string{"missing source", "missing_source_refs", "sources/missing-layout-source.md"}) ||
		!messageContainsAny(finalMessage, []string{"service_name", "service identity"}) ||
		!messageContainsAny(finalMessage, []string{"freshness", "## Freshness"}) {
		failures = append(failures, "answer did not report runner-visible invalid layout failures")
	}
	return verificationFromFailures(failures, "invalid layout inspection passed", []string{"synthesis/broken-layout.md", "records/services/broken-layout-service.md"})
}
func verifyRepoDocsAgentOpsRetrieval(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	search, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       repoDocsRetrievalSearchText,
			PathPrefix: "docs/architecture/",
			Limit:      10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	top, topFound := topSearchHit(search)
	agentOpsDocID, hasAgentOpsDoc, err := documentIDByPath(ctx, paths, repoDocsAgentOpsADRPath)
	if err != nil {
		return verificationResult{}, err
	}
	hasAgentOpsADR := searchContainsPath(search, repoDocsAgentOpsADRPath) ||
		(hasAgentOpsDoc && stringValuesInclude(turnMetrics.GetDocumentDocIDs, agentOpsDocID))
	_, hasKnowledgeConfig, err := documentIDByPath(ctx, paths, repoDocsKnowledgeConfigPath)
	if err != nil {
		return verificationResult{}, err
	}
	assistantPass := messageContainsAll(finalMessage, []string{repoDocsAgentOpsADRPath}) &&
		messageContainsAny(finalMessage, []string{"AgentOps", "agentops"}) &&
		messageContainsAny(finalMessage, []string{"installed", "openclerk", "runner"}) &&
		messageContainsAny(finalMessage, []string{"doc_id", "chunk_id", "citation", "cited"})
	searchedArchitecture := turnMetrics.SearchUsed && containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"docs/architecture/"})
	activityPass := len(repoDocsBypassFailures(turnMetrics)) == 0 &&
		searchedArchitecture &&
		hasAgentOpsADR
	failures := repoDocsBypassFailures(turnMetrics)
	if !topFound || !searchHitHasCitation(top) {
		failures = append(failures, "repo-docs retrieval search did not return cited hits")
	}
	if !hasAgentOpsDoc {
		failures = append(failures, "repo-docs seed did not import AgentOps ADR")
	}
	if hasAgentOpsDoc && !hasAgentOpsADR {
		failures = append(failures, "repo-docs retrieval workflow did not expose AgentOps ADR")
	}
	if !hasKnowledgeConfig {
		failures = append(failures, "repo-docs seed did not import knowledge configuration ADR")
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not use retrieval search")
	}
	if !searchedArchitecture {
		failures = append(failures, "agent did not use a docs/architecture/ path-prefix search")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not cite repo AgentOps docs with runner evidence")
	}
	databasePass := topFound && searchHitHasCitation(top) && hasAgentOpsDoc && hasKnowledgeConfig
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{repoDocsAgentOpsADRPath, repoDocsKnowledgeConfigPath},
	}, nil
}
func verifyRepoDocsSynthesisMaintenance(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	base, err := verifySourceLinkedSynthesis(ctx, paths, repoDocsSynthesisPath, finalMessage, sourceLinkedSynthesisExpectations{
		SourceRefs:      []string{repoDocsAgentProductionPath, repoDocsBaselineScenariosPath},
		RequireSearch:   true,
		RequireList:     true,
		Metrics:         turnMetrics,
		FinalAnswerPath: true,
		AdditionalDocs:  []string{repoDocsAgentProductionPath, repoDocsBaselineScenariosPath},
		AdditionalBodyRequirements: []string{
			"Repo-docs dogfood decision: use the existing OpenClerk document and retrieval runner actions.",
			"Production gate source: " + repoDocsAgentProductionPath,
			"Baseline scenarios source: " + repoDocsBaselineScenariosPath,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	count, err := exactDocumentCount(ctx, paths, repoDocsSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	failures := repoDocsBypassFailures(turnMetrics)
	if !base.Passed {
		failures = append(failures, base.Details)
	}
	if count != 1 {
		failures = append(failures, fmt.Sprintf("expected one repo-docs synthesis document, got %d", count))
	}
	databasePass := base.DatabasePass && count == 1
	assistantPass := base.AssistantPass && len(repoDocsBypassFailures(turnMetrics)) == 0
	return verificationResult{
		Passed:        databasePass && assistantPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass,
		Details:       missingDetails(failures),
		Documents:     base.Documents,
	}, nil
}
func verifyRepoDocsDecisionRecords(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	lookup, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionsLookup,
		Decisions: runner.DecisionLookupOptions{
			Text:   "knowledge configuration",
			Status: "accepted",
			Scope:  "knowledge-configuration",
			Owner:  "platform",
			Limit:  5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	agentOpsDecision, agentOpsDecisionErr := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action:     runner.RetrievalTaskActionDecisionRecord,
		DecisionID: "adr-agentops-only-knowledge-plane",
	})
	configProjection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-knowledge-configuration-v1",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	agentOpsProjection, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProjectionStates,
		Projection: runner.ProjectionStateOptions{
			Projection: "decisions",
			RefKind:    "decision",
			RefID:      "adr-agentops-only-knowledge-plane",
			Limit:      5,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}
	provenance, err := runner.RunRetrievalTask(ctx, cfg, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionProvenanceEvents,
		Provenance: runner.ProvenanceEventOptions{
			RefKind: "projection",
			RefID:   "decisions:adr-knowledge-configuration-v1",
			Limit:   10,
		},
	})
	if err != nil {
		return verificationResult{}, err
	}

	searchedArchitecture := turnMetrics.SearchUsed && containsAllStrings(turnMetrics.SearchPathPrefixes, []string{"docs/architecture/"})
	hasConfigDecision := false
	if lookup.Decisions != nil {
		for _, decision := range lookup.Decisions.Decisions {
			if decision.DecisionID == "adr-knowledge-configuration-v1" &&
				decision.Status == "accepted" &&
				decision.Scope == "knowledge-configuration" &&
				decision.Owner == "platform" &&
				len(decision.Citations) > 0 &&
				decision.Citations[0].Path == repoDocsKnowledgeConfigPath {
				hasConfigDecision = true
				break
			}
		}
	}
	hasAgentOpsDecisionRecord := agentOpsDecisionErr == nil &&
		agentOpsDecision.Decision != nil &&
		agentOpsDecision.Decision.DecisionID == "adr-agentops-only-knowledge-plane" &&
		agentOpsDecision.Decision.Status == "accepted" &&
		agentOpsDecision.Decision.Scope == "knowledge-plane" &&
		len(agentOpsDecision.Decision.Citations) > 0 &&
		agentOpsDecision.Decision.Citations[0].Path == repoDocsAgentOpsADRPath
	hasAgentOpsDecision := hasAgentOpsDecisionRecord
	hasConfigProjection := configProjection.Projections != nil &&
		len(configProjection.Projections.Projections) == 1 &&
		configProjection.Projections.Projections[0].Freshness == "fresh" &&
		configProjection.Projections.Projections[0].Details["path"] == repoDocsKnowledgeConfigPath
	hasAgentOpsProjection := agentOpsProjection.Projections != nil &&
		len(agentOpsProjection.Projections.Projections) == 1 &&
		agentOpsProjection.Projections.Projections[0].Freshness == "fresh" &&
		agentOpsProjection.Projections.Projections[0].Details["path"] == repoDocsAgentOpsADRPath
	hasProvenance := provenance.Provenance != nil && eventTypesInclude(provenance.Provenance.Events, "projection_refreshed")
	inspectedAgentOpsDecision := decisionRecordIDsInclude(turnMetrics.DecisionRecordIDs, "adr-agentops-only-knowledge-plane")
	assistantPass := messageContainsAll(finalMessage, []string{repoDocsAgentOpsADRPath, repoDocsKnowledgeConfigPath}) &&
		messageContainsAny(finalMessage, []string{"canonical markdown", "canonical adr", "authoritative"}) &&
		messageContainsAny(finalMessage, []string{"decisions_lookup", "decisions lookup", "decision lookup", "decision records"}) &&
		messageContainsAny(finalMessage, []string{"decision_record", "decision record", "adr record"}) &&
		messageContainsAny(finalMessage, []string{"fresh", "freshness"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "projection"})
	activityPass := len(repoDocsBypassFailures(turnMetrics)) == 0 &&
		searchedArchitecture &&
		turnMetrics.DecisionsLookupUsed &&
		inspectedAgentOpsDecision &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	failures := repoDocsBypassFailures(turnMetrics)
	if !searchedArchitecture {
		failures = append(failures, "agent did not use a docs/architecture/ path-prefix search")
	}
	if !hasConfigDecision {
		failures = append(failures, "repo-docs knowledge configuration decision lookup missing")
	}
	if !hasAgentOpsDecision {
		failures = append(failures, "repo-docs AgentOps decision detail missing")
	}
	if !hasConfigProjection {
		failures = append(failures, "repo-docs knowledge configuration decision projection is not fresh")
	}
	if !hasAgentOpsProjection {
		failures = append(failures, "repo-docs AgentOps decision projection is not fresh")
	}
	if !hasProvenance {
		failures = append(failures, "repo-docs decision projection provenance missing")
	}
	if !activityPass {
		failures = append(failures, "agent did not use required search/decision/projection/provenance workflow")
	}
	if !assistantPass {
		failures = append(failures, "final answer did not report repo-docs decision-record evidence")
	}
	databasePass := hasConfigDecision && hasAgentOpsDecision && hasConfigProjection && hasAgentOpsProjection && hasProvenance
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{repoDocsAgentOpsADRPath, repoDocsKnowledgeConfigPath},
	}, nil
}
func repoDocsBypassFailures(turnMetrics metrics) []string {
	failures := []string{}
	if turnMetrics.BroadRepoSearch {
		failures = append(failures, "agent used broad repo search")
	}
	if turnMetrics.DirectSQLiteAccess {
		failures = append(failures, "agent used direct SQLite")
	}
	if turnMetrics.LegacyRunnerUsage {
		failures = append(failures, "agent used source-built or legacy runner path")
	}
	if turnMetrics.GeneratedFileInspection {
		failures = append(failures, "agent inspected generated files")
	}
	if turnMetrics.ModuleCacheInspection {
		failures = append(failures, "agent inspected module cache")
	}
	return failures
}
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
func verifyDocumentThisExplicitCreate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, documentThisExplicitPath)
	if err != nil {
		return verificationResult{}, err
	}
	body := ""
	title := ""
	if doc != nil {
		body = doc.Body
		title = doc.Title
	}
	sourcesCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: note",
		"Document-this explicit article/docs/paper/transcript intake uses strict runner JSON.",
		"Required fields were supplied before create_document.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisExplicitPath)
	}
	if found && title != documentThisExplicitTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", documentThisExplicitTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if sourcesCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no source autofiling docs, got %d", sourcesCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit document through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisExplicitPath, documentThisExplicitTitle})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit document path and title")
	}
	databasePass := found && title == documentThisExplicitTitle && len(missingRequired(body, required)) == 0 && sourcesCount == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisExplicitPath},
	}, nil
}
func verifyDocumentThisExplicitOverrides(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	doc, found, err := documentByPath(ctx, paths, documentThisOverridePath)
	if err != nil {
		return verificationResult{}, err
	}
	autofiledCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
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
		"Explicit document-this override path and title win.",
		"Do not infer a sources/ path from mixed URLs.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisOverridePath)
	}
	if found && title != documentThisOverrideTitle {
		failures = append(failures, fmt.Sprintf("expected stored title %q, got %q", documentThisOverrideTitle, title))
	}
	failures = append(failures, missingRequired(body, required)...)
	if autofiledCount != 0 {
		failures = append(failures, fmt.Sprintf("expected no inferred source docs, got %d", autofiledCount))
	}
	if turnMetrics.ToolCalls == 0 || turnMetrics.CommandExecutions == 0 {
		failures = append(failures, "agent did not create explicit override through installed runner commands")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisOverridePath, documentThisOverrideTitle})
	if !assistantPass {
		failures = append(failures, "final answer did not report explicit override path and title")
	}
	databasePass := found && title == documentThisOverrideTitle && len(missingRequired(body, required)) == 0 && autofiledCount == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ToolCalls > 0 && turnMetrics.CommandExecutions > 0
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisOverridePath},
	}, nil
}
func verifyDocumentThisDuplicateCandidate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	existingCount, err := exactDocumentCount(ctx, paths, documentThisDuplicateExistingPath)
	if err != nil {
		return verificationResult{}, err
	}
	candidateCount, err := exactDocumentCount(ctx, paths, documentThisDuplicateCandidatePath)
	if err != nil {
		return verificationResult{}, err
	}
	sourceCount, err := documentCountWithPrefix(ctx, paths, "sources/document-this/")
	if err != nil {
		return verificationResult{}, err
	}
	failures := documentThisBypassFailures(turnMetrics)
	if existingCount != 1 {
		failures = append(failures, fmt.Sprintf("expected one existing source %s, got %d", documentThisDuplicateExistingPath, existingCount))
	}
	if candidateCount != 0 {
		failures = append(failures, "created duplicate candidate "+documentThisDuplicateCandidatePath)
	}
	if sourceCount != 1 {
		failures = append(failures, fmt.Sprintf("expected only the seeded document-this source document, got %d", sourceCount))
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search for duplicate candidate")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list source candidates")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisDuplicateExistingPath}) &&
		messageContainsAny(finalMessage, []string{"duplicate", "existing", "already"}) &&
		messageContainsAny(finalMessage, []string{"not create", "did not create", "no new"})
	if !assistantPass {
		failures = append(failures, "final answer did not report duplicate candidate and no-create outcome")
	}
	databasePass := existingCount == 1 && candidateCount == 0 && sourceCount == 1
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.SearchUsed && turnMetrics.ListDocumentsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisDuplicateExistingPath, documentThisDuplicateCandidatePath},
	}, nil
}
func verifyDocumentThisExistingUpdate(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, documentThisUpdateTargetPath)
	if err != nil {
		return verificationResult{}, err
	}
	decoyBody, decoyFound, err := documentBodyByPath(ctx, paths, documentThisUpdateDecoyPath)
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"## Decisions",
		"Use strict runner JSON for document-this intake.",
	}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisUpdateTargetPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	if decoyFound && strings.Contains(decoyBody, "Use strict runner JSON for document-this intake.") {
		failures = append(failures, "updated decoy "+documentThisUpdateDecoyPath)
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list update candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not inspect existing target before update")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisUpdateTargetPath}) &&
		messageContainsAny(finalMessage, []string{"updated", "appended", "replaced"}) &&
		messageContainsAny(finalMessage, []string{"decoy", "not update", "did not update", "target"})
	if !assistantPass {
		failures = append(failures, "final answer did not report target update and decoy avoidance")
	}
	databasePass := found && len(missingRequired(body, required)) == 0 && (!decoyFound || !strings.Contains(decoyBody, "Use strict runner JSON for document-this intake."))
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 && turnMetrics.ListDocumentsUsed && turnMetrics.GetDocumentUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents:     []string{documentThisUpdateTargetPath, documentThisUpdateDecoyPath},
	}, nil
}
func verifyDocumentThisSynthesisFreshness(ctx context.Context, paths evalPaths, finalMessage string, turnMetrics metrics) (verificationResult, error) {
	body, found, err := documentBodyByPath(ctx, paths, documentThisSynthesisPath)
	if err != nil {
		return verificationResult{}, err
	}
	duplicateCount, err := exactDocumentCount(ctx, paths, documentThisSynthesisDuplicatePath)
	if err != nil {
		return verificationResult{}, err
	}
	required := []string{
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"Current document-this intake guidance: update existing synthesis after source, duplicate, provenance, and freshness checks.",
		"## Sources",
		"## Freshness",
	}
	expectedRefs := []string{documentThisArticlePath, documentThisDocsPath, documentThisPaperPath, documentThisTranscriptPath}
	failures := documentThisBypassFailures(turnMetrics)
	if !found {
		failures = append(failures, "missing "+documentThisSynthesisPath)
	}
	failures = append(failures, missingRequired(body, required)...)
	failures = append(failures, sourceRefsFrontmatterFailures(body, expectedRefs)...)
	if duplicateCount != 0 {
		failures = append(failures, "created duplicate synthesis "+documentThisSynthesisDuplicatePath)
	}
	if !turnMetrics.SearchUsed {
		failures = append(failures, "agent did not search source evidence")
	}
	if !turnMetrics.ListDocumentsUsed {
		failures = append(failures, "agent did not list synthesis candidates")
	}
	if !turnMetrics.GetDocumentUsed {
		failures = append(failures, "agent did not inspect existing synthesis before update")
	}
	if !turnMetrics.ProjectionStatesUsed {
		failures = append(failures, "agent did not inspect projection_states")
	}
	if !turnMetrics.ProvenanceEventsUsed {
		failures = append(failures, "agent did not inspect provenance_events")
	}
	assistantPass := messageContainsAll(finalMessage, []string{documentThisSynthesisPath}) &&
		messageContainsAny(finalMessage, []string{"freshness", "projection", "fresh"}) &&
		messageContainsAny(finalMessage, []string{"provenance", "source refs", "source_refs"}) &&
		messageContainsAny(finalMessage, []string{"no duplicate", "did not create", "not create"})
	if !assistantPass {
		failures = append(failures, "final answer did not report synthesis update, freshness/provenance, and duplicate avoidance")
	}
	databasePass := found &&
		duplicateCount == 0 &&
		len(missingRequired(body, required)) == 0 &&
		len(sourceRefsFrontmatterFailures(body, expectedRefs)) == 0
	activityPass := len(documentThisBypassFailures(turnMetrics)) == 0 &&
		turnMetrics.SearchUsed &&
		turnMetrics.ListDocumentsUsed &&
		turnMetrics.GetDocumentUsed &&
		turnMetrics.ProjectionStatesUsed &&
		turnMetrics.ProvenanceEventsUsed
	return verificationResult{
		Passed:        databasePass && assistantPass && activityPass,
		DatabasePass:  databasePass,
		AssistantPass: assistantPass && activityPass,
		Details:       missingDetails(failures),
		Documents: append([]string{
			documentThisSynthesisPath,
			documentThisSynthesisDuplicatePath,
		}, expectedRefs...),
	}, nil
}
