package runner

import (
	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func toPaths(paths runclient.Paths) Paths {
	return Paths{
		DatabasePath: paths.DatabasePath,
		VaultRoot:    paths.VaultRoot,
	}
}

func toDocument(document domain.Document) Document {
	return Document{
		DocID:     document.DocID,
		Path:      document.Path,
		Title:     document.Title,
		Body:      document.Body,
		Headings:  append([]string(nil), document.Headings...),
		Metadata:  cloneStringMap(document.Metadata),
		CreatedAt: document.CreatedAt,
		UpdatedAt: document.UpdatedAt,
	}
}

func toDocumentSummaries(documents []domain.DocumentSummary) []DocumentSummary {
	result := make([]DocumentSummary, 0, len(documents))
	for _, document := range documents {
		result = append(result, DocumentSummary{
			DocID:     document.DocID,
			Path:      document.Path,
			Title:     document.Title,
			Metadata:  cloneStringMap(document.Metadata),
			UpdatedAt: document.UpdatedAt,
		})
	}
	return result
}

func toPageInfo(pageInfo domain.PageInfo) PageInfo {
	return PageInfo{
		NextCursor: pageInfo.NextCursor,
		HasMore:    pageInfo.HasMore,
	}
}

func toSearchResult(result domain.SearchResult) SearchResult {
	hits := make([]SearchHit, 0, len(result.Hits))
	for _, hit := range result.Hits {
		hits = append(hits, SearchHit{
			Rank:      hit.Rank,
			Score:     hit.Score,
			DocID:     hit.DocID,
			ChunkID:   hit.ChunkID,
			Title:     hit.Title,
			Snippet:   hit.Snippet,
			Citations: toCitations(hit.Citations),
		})
	}
	return SearchResult{
		Hits:     hits,
		PageInfo: toPageInfo(result.PageInfo),
	}
}

func toSourceIngestionResult(result domain.SourceIngestionResult) SourceIngestionResult {
	return SourceIngestionResult{
		DocID:       result.DocID,
		SourcePath:  result.SourcePath,
		SourceURL:   result.SourceURL,
		SourceType:  result.SourceType,
		AssetPath:   result.AssetPath,
		DerivedPath: result.DerivedPath,
		Citations:   toCitations(result.Citations),
		SHA256:      result.SHA256,
		SizeBytes:   result.SizeBytes,
		MIMEType:    result.MIMEType,
		PageCount:   result.PageCount,
		CapturedAt:  result.CapturedAt,
		PDFMetadata: SourcePDFMetadata(result.PDFMetadata),
	}
}

func toVideoIngestionResult(result domain.VideoIngestionResult) VideoIngestionResult {
	return VideoIngestionResult{
		DocID:                    result.DocID,
		SourcePath:               result.SourcePath,
		SourceURL:                result.SourceURL,
		AssetPath:                result.AssetPath,
		Citations:                toCitations(result.Citations),
		TranscriptSHA256:         result.TranscriptSHA256,
		PreviousTranscriptSHA256: result.PreviousTranscriptSHA256,
		NewTranscriptSHA256:      result.NewTranscriptSHA256,
		CapturedAt:               result.CapturedAt,
		TranscriptPolicy:         result.TranscriptPolicy,
		TranscriptOrigin:         result.TranscriptOrigin,
		Language:                 result.Language,
		Tool:                     result.Tool,
		Model:                    result.Model,
	}
}

func toCitations(citations []domain.Citation) []Citation {
	result := make([]Citation, 0, len(citations))
	for _, citation := range citations {
		result = append(result, Citation{
			DocID:     citation.DocID,
			ChunkID:   citation.ChunkID,
			Path:      citation.Path,
			Heading:   citation.Heading,
			LineStart: citation.LineStart,
			LineEnd:   citation.LineEnd,
		})
	}
	return result
}

func toDocumentLinksResult(links domain.DocumentLinks) DocumentLinks {
	return DocumentLinks{
		DocID:    links.DocID,
		Outgoing: toDocumentLinks(links.Outgoing),
		Incoming: toDocumentLinks(links.Incoming),
	}
}

func toDocumentLinks(links []domain.DocumentLink) []DocumentLink {
	result := make([]DocumentLink, 0, len(links))
	for _, link := range links {
		result = append(result, DocumentLink{
			DocID:     link.DocID,
			Path:      link.Path,
			Title:     link.Title,
			Citations: toCitations(link.Citations),
		})
	}
	return result
}

func toGraphNeighborhood(neighborhood domain.GraphNeighborhood) GraphNeighborhood {
	nodes := make([]GraphNode, 0, len(neighborhood.Nodes))
	for _, node := range neighborhood.Nodes {
		nodes = append(nodes, GraphNode{
			NodeID:    node.NodeID,
			Type:      node.Type,
			Label:     node.Label,
			Citations: toCitations(node.Citations),
		})
	}
	edges := make([]GraphEdge, 0, len(neighborhood.Edges))
	for _, edge := range neighborhood.Edges {
		edges = append(edges, GraphEdge{
			EdgeID:     edge.EdgeID,
			FromNodeID: edge.FromNodeID,
			ToNodeID:   edge.ToNodeID,
			Kind:       edge.Kind,
			Citations:  toCitations(edge.Citations),
		})
	}
	return GraphNeighborhood{
		Nodes: nodes,
		Edges: edges,
	}
}

func toRecordLookupResult(result domain.RecordLookupResult) RecordLookupResult {
	return RecordLookupResult{
		Entities: toRecordEntities(result.Entities),
		PageInfo: toPageInfo(result.PageInfo),
	}
}

func toRecordEntities(entities []domain.RecordEntity) []RecordEntity {
	result := make([]RecordEntity, 0, len(entities))
	for _, entity := range entities {
		result = append(result, toRecordEntity(entity))
	}
	return result
}

func toRecordEntity(entity domain.RecordEntity) RecordEntity {
	facts := make([]RecordFact, 0, len(entity.Facts))
	for _, fact := range entity.Facts {
		facts = append(facts, RecordFact{
			Key:        fact.Key,
			Value:      fact.Value,
			ObservedAt: fact.ObservedAt,
		})
	}
	return RecordEntity{
		EntityID:   entity.EntityID,
		EntityType: entity.EntityType,
		Name:       entity.Name,
		Summary:    entity.Summary,
		Facts:      facts,
		Citations:  toCitations(entity.Citations),
		UpdatedAt:  entity.UpdatedAt,
	}
}

func toServiceLookupResult(result domain.ServiceLookupResult) ServiceLookupResult {
	return ServiceLookupResult{
		Services: toServiceRecords(result.Services),
		PageInfo: toPageInfo(result.PageInfo),
	}
}

func toServiceRecords(services []domain.ServiceRecord) []ServiceRecord {
	result := make([]ServiceRecord, 0, len(services))
	for _, service := range services {
		result = append(result, toServiceRecord(service))
	}
	return result
}

func toServiceRecord(service domain.ServiceRecord) ServiceRecord {
	facts := make([]ServiceFact, 0, len(service.Facts))
	for _, fact := range service.Facts {
		facts = append(facts, ServiceFact{
			Key:        fact.Key,
			Value:      fact.Value,
			ObservedAt: fact.ObservedAt,
		})
	}
	return ServiceRecord{
		ServiceID: service.ServiceID,
		Name:      service.Name,
		Status:    service.Status,
		Owner:     service.Owner,
		Interface: service.Interface,
		Summary:   service.Summary,
		Facts:     facts,
		Citations: toCitations(service.Citations),
		UpdatedAt: service.UpdatedAt,
	}
}

func toDecisionLookupResult(result domain.DecisionLookupResult) DecisionLookupResult {
	return DecisionLookupResult{
		Decisions: toDecisionRecords(result.Decisions),
		PageInfo:  toPageInfo(result.PageInfo),
	}
}

func toDecisionRecords(decisions []domain.DecisionRecord) []DecisionRecord {
	result := make([]DecisionRecord, 0, len(decisions))
	for _, decision := range decisions {
		result = append(result, toDecisionRecord(decision))
	}
	return result
}

func toDecisionRecord(decision domain.DecisionRecord) DecisionRecord {
	return DecisionRecord{
		DecisionID:   decision.DecisionID,
		Title:        decision.Title,
		Status:       decision.Status,
		Scope:        decision.Scope,
		Owner:        decision.Owner,
		Date:         decision.Date,
		Summary:      decision.Summary,
		Supersedes:   append([]string(nil), decision.Supersedes...),
		SupersededBy: append([]string(nil), decision.SupersededBy...),
		SourceRefs:   append([]string(nil), decision.SourceRefs...),
		Citations:    toCitations(decision.Citations),
		UpdatedAt:    decision.UpdatedAt,
	}
}

func toProvenanceEventList(list domain.ProvenanceEventResult) ProvenanceEventList {
	return ProvenanceEventList{
		Events:   toProvenanceEvents(list.Events),
		PageInfo: toPageInfo(list.PageInfo),
	}
}

func toProvenanceEvents(events []domain.ProvenanceEvent) []ProvenanceEvent {
	result := make([]ProvenanceEvent, 0, len(events))
	for _, event := range events {
		result = append(result, ProvenanceEvent{
			EventID:    event.EventID,
			EventType:  event.EventType,
			RefKind:    event.RefKind,
			RefID:      event.RefID,
			SourceRef:  event.SourceRef,
			OccurredAt: event.OccurredAt,
			Details:    cloneStringMap(event.Details),
		})
	}
	return result
}

func toProjectionStateList(list domain.ProjectionStateResult) ProjectionStateList {
	return ProjectionStateList{
		Projections: toProjectionStates(list.Projections),
		PageInfo:    toPageInfo(list.PageInfo),
	}
}

func toProjectionStates(projections []domain.ProjectionState) []ProjectionState {
	result := make([]ProjectionState, 0, len(projections))
	for _, projection := range projections {
		result = append(result, ProjectionState{
			Projection:        projection.Projection,
			RefKind:           projection.RefKind,
			RefID:             projection.RefID,
			SourceRef:         projection.SourceRef,
			Freshness:         projection.Freshness,
			ProjectionVersion: projection.ProjectionVersion,
			UpdatedAt:         projection.UpdatedAt,
			Details:           cloneStringMap(projection.Details),
		})
	}
	return result
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
