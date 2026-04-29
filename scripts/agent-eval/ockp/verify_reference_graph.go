package main

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"strings"
)

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
