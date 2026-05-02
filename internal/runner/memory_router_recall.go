package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	defaultMemoryRouterRecallQuery = "memory router temporal recall session promotion feedback weighting routing canonical docs"
	memoryRouterRecallPrefix       = "notes/memory-router/"
	memoryRouterSessionPath        = "notes/memory-router/session-observation.md"
	memoryRouterTemporalPath       = "notes/memory-router/temporal-policy.md"
	memoryRouterFeedbackPath       = "notes/memory-router/feedback-weighting.md"
	memoryRouterRoutingPath        = "notes/memory-router/routing-policy.md"
	memoryRouterSynthesisPath      = "synthesis/memory-router-reference.md"
)

var memoryRouterCanonicalPaths = []string{
	memoryRouterSessionPath,
	memoryRouterTemporalPath,
	memoryRouterFeedbackPath,
	memoryRouterRoutingPath,
}

type memoryRouterRecallDoc struct {
	Path      string
	DocID     string
	Body      string
	Found     bool
	Inspected bool
}

func runMemoryRouterRecallReport(ctx context.Context, client *runclient.Client, options MemoryRouterRecallOptions) (MemoryRouterRecallReport, error) {
	query := strings.TrimSpace(options.Query)
	if query == "" {
		query = defaultMemoryRouterRecallQuery
	}
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}

	search, err := client.Search(ctx, domain.SearchQuery{
		Text:  query,
		Limit: limit,
	})
	if err != nil {
		return MemoryRouterRecallReport{}, err
	}

	docs, err := memoryRouterRecallDocs(ctx, client, limit)
	if err != nil {
		return MemoryRouterRecallReport{}, err
	}
	synthesis, err := memoryRouterRecallSynthesis(ctx, client, limit)
	if err != nil {
		return MemoryRouterRecallReport{}, err
	}

	session := docs[memoryRouterSessionPath]
	var provenanceRefs []string
	var provenanceEvents []domain.ProvenanceEvent
	if session.Found {
		events, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
			RefKind: "document",
			RefID:   session.DocID,
			Limit:   limit,
		})
		if err != nil {
			return MemoryRouterRecallReport{}, err
		}
		provenanceEvents = events.Events
		provenanceRefs = append(provenanceRefs, "document:"+session.DocID)
		for _, event := range events.Events {
			provenanceRefs = append(provenanceRefs, event.EventType+":"+event.EventID)
		}
	} else {
		provenanceRefs = append(provenanceRefs, "missing:"+memoryRouterSessionPath)
	}

	var projections []domain.ProjectionState
	if synthesis.Found {
		states, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
			Projection: "synthesis",
			RefKind:    "document",
			RefID:      synthesis.DocID,
			Limit:      limit,
		})
		if err != nil {
			return MemoryRouterRecallReport{}, err
		}
		projections = states.Projections
	}

	return assembleMemoryRouterRecallReport(query, len(search.Hits), docs, synthesis, provenanceRefs, provenanceEvents, projections), nil
}

func memoryRouterRecallDocs(ctx context.Context, client *runclient.Client, limit int) (map[string]memoryRouterRecallDoc, error) {
	docs := make(map[string]memoryRouterRecallDoc, len(memoryRouterCanonicalPaths))
	for _, path := range memoryRouterCanonicalPaths {
		docs[path] = memoryRouterRecallDoc{Path: path}
	}

	remaining := len(memoryRouterCanonicalPaths)
	cursor := ""
	pageLimit := max(limit, len(memoryRouterCanonicalPaths))
	for {
		list, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix: memoryRouterRecallPrefix,
			Limit:      pageLimit,
			Cursor:     cursor,
		})
		if err != nil {
			return nil, err
		}

		for _, summary := range list.Documents {
			current, ok := docs[summary.Path]
			if !ok || current.Found {
				continue
			}
			doc, err := client.GetDocument(ctx, summary.DocID)
			if err != nil {
				return nil, err
			}
			docs[summary.Path] = memoryRouterRecallDoc{
				Path:      summary.Path,
				DocID:     summary.DocID,
				Body:      doc.Body,
				Found:     true,
				Inspected: true,
			}
			remaining--
		}

		if remaining == 0 || !list.PageInfo.HasMore || list.PageInfo.NextCursor == "" {
			break
		}
		cursor = list.PageInfo.NextCursor
	}
	return docs, nil
}

func memoryRouterRecallSynthesis(ctx context.Context, client *runclient.Client, limit int) (memoryRouterRecallDoc, error) {
	cursor := ""
	pageLimit := max(limit, 20)
	for {
		list, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix: "synthesis/",
			Limit:      pageLimit,
			Cursor:     cursor,
		})
		if err != nil {
			return memoryRouterRecallDoc{}, err
		}

		for _, summary := range list.Documents {
			if summary.Path != memoryRouterSynthesisPath {
				continue
			}
			doc, err := client.GetDocument(ctx, summary.DocID)
			if err != nil {
				return memoryRouterRecallDoc{}, err
			}
			return memoryRouterRecallDoc{
				Path:      summary.Path,
				DocID:     summary.DocID,
				Body:      doc.Body,
				Found:     true,
				Inspected: true,
			}, nil
		}

		if !list.PageInfo.HasMore || list.PageInfo.NextCursor == "" {
			break
		}
		cursor = list.PageInfo.NextCursor
	}
	return memoryRouterRecallDoc{Path: memoryRouterSynthesisPath}, nil
}

func assembleMemoryRouterRecallReport(query string, searchHits int, docs map[string]memoryRouterRecallDoc, synthesis memoryRouterRecallDoc, provenanceRefs []string, provenanceEvents []domain.ProvenanceEvent, projections []domain.ProjectionState) MemoryRouterRecallReport {
	refs := make([]string, 0, len(memoryRouterCanonicalPaths)+1)
	missing := []string{}
	foundCanonical := 0
	for _, path := range memoryRouterCanonicalPaths {
		doc := docs[path]
		if doc.Found && doc.Inspected {
			refs = append(refs, path)
			foundCanonical++
			continue
		}
		refs = append(refs, "missing:"+path)
		missing = append(missing, path)
	}
	if synthesis.Found && synthesis.Inspected {
		refs = append(refs, memoryRouterSynthesisPath)
	} else {
		refs = append(refs, "missing:"+memoryRouterSynthesisPath)
		missing = append(missing, memoryRouterSynthesisPath)
	}

	provenanceStatus := "session observation provenance was inspected"
	if len(provenanceEvents) == 0 {
		provenanceStatus = "missing session observation provenance"
		missing = append(missing, "provenance:"+memoryRouterSessionPath)
	}

	freshness := synthesisFreshnessSummary(projections)
	if !strings.Contains(freshness, "fresh synthesis projection") {
		missing = append(missing, "projection:"+memoryRouterSynthesisPath)
	}

	validation := "read-only openclerk retrieval report; no writes, no memory transports, no remember/recall actions, no autonomous router APIs, no vector stores, no embedding stores, no graph memory, no direct SQLite, no direct vault inspection, no HTTP/MCP bypasses, no unsupported transports, no source-built runners, and no hidden authority ranking"
	if len(missing) > 0 {
		validation += "; missing evidence: " + strings.Join(missing, ", ")
	}

	return MemoryRouterRecallReport{
		QuerySummary:          fmt.Sprintf("memory/router recall report for %q; search returned %d hits; canonical evidence %d/%d present", query, searchHits, foundCanonical, len(memoryRouterCanonicalPaths)),
		TemporalStatus:        temporalStatusSummary(docs, synthesis),
		CanonicalEvidenceRefs: refs,
		StaleSessionStatus:    fmt.Sprintf("session observations are stale or advisory until promoted through canonical markdown with source refs; %s", provenanceStatus),
		FeedbackWeighting:     feedbackWeightingSummary(docs),
		RoutingRationale:      routingRationaleSummary(docs),
		ProvenanceRefs:        provenanceRefs,
		SynthesisFreshness:    freshness,
		ValidationBoundaries:  validation,
		AuthorityLimits:       "canonical markdown remains durable memory authority; synthesis is derived evidence with provenance and freshness; feedback is advisory; this report is read-only and does not create hidden memory authority or autonomous routing decisions",
	}
}

func temporalStatusSummary(docs map[string]memoryRouterRecallDoc, synthesis memoryRouterRecallDoc) string {
	if docs[memoryRouterSessionPath].Found &&
		docs[memoryRouterTemporalPath].Found &&
		synthesis.Found &&
		strings.Contains(synthesis.Body, "current canonical docs outrank stale session observations") {
		return "current canonical docs over stale session observations; current canonical docs outrank stale session observations"
	}
	return "incomplete temporal status evidence; current canonical docs should outrank stale session observations when canonical docs and synthesis freshness are present"
}

func feedbackWeightingSummary(docs map[string]memoryRouterRecallDoc) string {
	if docs[memoryRouterFeedbackPath].Found {
		return "feedback weighting is advisory only and cannot hide stale or conflicting canonical evidence"
	}
	return "missing feedback weighting evidence; feedback must remain advisory until canonical evidence is present"
}

func routingRationaleSummary(docs map[string]memoryRouterRecallDoc) string {
	if docs[memoryRouterRoutingPath].Found {
		return "routing rationale uses existing AgentOps document and retrieval actions; no autonomous router API or hidden authority ranking is introduced"
	}
	return "missing routing policy evidence; no autonomous router API or hidden authority ranking is introduced"
}

func synthesisFreshnessSummary(projections []domain.ProjectionState) string {
	if len(projections) == 0 {
		return "missing synthesis projection for " + memoryRouterSynthesisPath
	}
	for _, projection := range projections {
		if projection.Freshness == "fresh" {
			return "fresh synthesis projection for " + memoryRouterSynthesisPath
		}
	}
	return "synthesis projection for " + memoryRouterSynthesisPath + " is not fresh"
}
