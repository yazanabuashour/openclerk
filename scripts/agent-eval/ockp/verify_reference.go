package main

import (
	"context"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"strings"
)

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
