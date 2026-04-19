package app

import (
	"context"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

type Service struct {
	store domain.Store
}

func New(store domain.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Close() error {
	return s.store.Close()
}

func (s *Service) Capabilities(ctx context.Context) (domain.Capabilities, error) {
	return s.store.Capabilities(ctx)
}

func (s *Service) Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResult, error) {
	return s.store.Search(ctx, query)
}

func (s *Service) ListDocuments(ctx context.Context, query domain.DocumentListQuery) (domain.DocumentListResult, error) {
	return s.store.ListDocuments(ctx, query)
}

func (s *Service) CreateDocument(ctx context.Context, input domain.CreateDocumentInput) (domain.Document, error) {
	return s.store.CreateDocument(ctx, input)
}

func (s *Service) GetDocument(ctx context.Context, docID string) (domain.Document, error) {
	return s.store.GetDocument(ctx, docID)
}

func (s *Service) GetDocumentLinks(ctx context.Context, docID string) (domain.DocumentLinks, error) {
	return s.store.GetDocumentLinks(ctx, docID)
}

func (s *Service) AppendDocument(ctx context.Context, docID string, input domain.AppendDocumentInput) (domain.Document, error) {
	return s.store.AppendDocument(ctx, docID, input)
}

func (s *Service) ReplaceDocumentSection(ctx context.Context, docID string, input domain.ReplaceSectionInput) (domain.Document, error) {
	return s.store.ReplaceDocumentSection(ctx, docID, input)
}

func (s *Service) GetChunk(ctx context.Context, chunkID string) (domain.Chunk, error) {
	return s.store.GetChunk(ctx, chunkID)
}

func (s *Service) GraphNeighborhood(ctx context.Context, input domain.GraphNeighborhoodInput) (domain.GraphNeighborhood, error) {
	return s.store.GraphNeighborhood(ctx, input)
}

func (s *Service) RecordsLookup(ctx context.Context, input domain.RecordLookupInput) (domain.RecordLookupResult, error) {
	return s.store.RecordsLookup(ctx, input)
}

func (s *Service) GetRecordEntity(ctx context.Context, entityID string) (domain.RecordEntity, error) {
	return s.store.GetRecordEntity(ctx, entityID)
}

func (s *Service) ServicesLookup(ctx context.Context, input domain.ServiceLookupInput) (domain.ServiceLookupResult, error) {
	return s.store.ServicesLookup(ctx, input)
}

func (s *Service) GetServiceRecord(ctx context.Context, serviceID string) (domain.ServiceRecord, error) {
	return s.store.GetServiceRecord(ctx, serviceID)
}

func (s *Service) ListProvenanceEvents(ctx context.Context, query domain.ProvenanceEventQuery) (domain.ProvenanceEventResult, error) {
	return s.store.ListProvenanceEvents(ctx, query)
}

func (s *Service) ListProjectionStates(ctx context.Context, query domain.ProjectionStateQuery) (domain.ProjectionStateResult, error) {
	return s.store.ListProjectionStates(ctx, query)
}
