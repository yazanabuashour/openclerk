package sqlite

import (
	"context"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"sort"
	"strings"
	"time"
)

type synthesisProjectionInput struct {
	Document    domain.Document
	Frontmatter map[string]string
	SourceRefs  []string
}

func (s *Store) rebuildSynthesis(ctx context.Context) error {
	documents, err := s.loadAllDocuments(ctx)
	if err != nil {
		return err
	}
	previousStates, err := s.loadProjectionStateSnapshots(ctx, "synthesis")
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.InternalError("begin synthesis rebuild", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.ExecContext(ctx, `DELETE FROM projection_states WHERE projection_name = 'synthesis'`); err != nil {
		return domain.InternalError("reset synthesis projection", err)
	}

	now := s.now().UTC()
	documentIndex := documentsByPath(documents)
	synthesisDocs := synthesisProjectionInputs(documents)
	for _, synthesis := range synthesisDocs {
		projection := buildSynthesisProjectionState(synthesis, documentIndex, now)
		if previous, ok := previousStates[projection.RefID]; ok && previous.ProjectionVersion == projection.ProjectionVersion {
			projection.UpdatedAt = previous.UpdatedAt
		}
		if err := upsertProjectionState(ctx, tx, projection); err != nil {
			return err
		}
		previous, hadPrevious := previousStates[projection.RefID]
		if hadPrevious && previous.ProjectionVersion == projection.ProjectionVersion {
			continue
		}
		eventType := "projection_refreshed"
		if projection.Freshness == "stale" {
			eventType = "projection_invalidated"
		}
		if err := insertProvenanceEvent(ctx, tx, domain.ProvenanceEvent{
			EventID:    hashID("event", eventType, "synthesis", projection.RefID, projection.ProjectionVersion, now.Format(time.RFC3339Nano)),
			EventType:  eventType,
			RefKind:    "projection",
			RefID:      "synthesis:" + projection.RefID,
			SourceRef:  projection.SourceRef,
			OccurredAt: now,
			Details: map[string]string{
				"projection":       "synthesis",
				"path":             synthesis.Document.Path,
				"freshness":        projection.Freshness,
				"freshness_reason": projection.Details["freshness_reason"],
				"version":          projection.ProjectionVersion,
			},
		}); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.InternalError("commit synthesis rebuild", err)
	}
	return nil
}

func synthesisProjectionInputs(documents []domain.Document) []synthesisProjectionInput {
	result := []synthesisProjectionInput{}
	for _, doc := range documents {
		if !isSynthesisDocument(doc.Path, doc.Metadata) {
			continue
		}
		result = append(result, synthesisProjectionInput{
			Document:    doc,
			Frontmatter: doc.Metadata,
			SourceRefs:  splitPathList(doc.Metadata["source_refs"]),
		})
	}
	return result
}

func isSynthesisDocument(docPath string, frontmatter map[string]string) bool {
	return isSynthesisPath(docPath) && strings.EqualFold(strings.TrimSpace(frontmatter["type"]), "synthesis")
}

func isSynthesisPath(docPath string) bool {
	return strings.HasPrefix(docPath, rootSynthesisPathPrefix)
}

func buildSynthesisProjectionState(input synthesisProjectionInput, documentsByPath map[string]domain.Document, now time.Time) domain.ProjectionState {
	sourceSet := stringSet(input.SourceRefs)
	resolvedDocRefs := []string{}
	currentRefs := []string{}
	supersededRefs := []string{}
	missingRefs := []string{}
	staleRefs := []string{}
	reasons := []string{}
	versionInputs := []string{
		"synthesis:" + input.Document.DocID,
		"path:" + input.Document.Path,
		"updated:" + input.Document.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}

	for _, ref := range input.SourceRefs {
		source, ok := documentsByPath[ref]
		if !ok {
			missingRefs = appendUnique(missingRefs, ref)
			reasons = appendUnique(reasons, "missing source refs")
			versionInputs = append(versionInputs, "missing:"+ref)
			continue
		}
		resolvedDocRefs = appendUnique(resolvedDocRefs, "doc:"+source.DocID)
		versionInputs = append(versionInputs,
			"source:"+ref,
			"source_updated:"+source.UpdatedAt.UTC().Format(time.RFC3339Nano),
		)
		if source.UpdatedAt.After(input.Document.UpdatedAt) {
			staleRefs = appendUnique(staleRefs, ref)
			reasons = appendUnique(reasons, "source newer than synthesis")
		}

		supersededBy := splitPathList(source.Metadata["superseded_by"])
		if strings.EqualFold(strings.TrimSpace(source.Metadata["status"]), "superseded") {
			supersededRefs = appendUnique(supersededRefs, ref)
			if len(supersededBy) == 0 {
				staleRefs = appendUnique(staleRefs, ref)
				reasons = appendUnique(reasons, "superseded source has no current replacement")
			}
			for _, current := range supersededBy {
				currentRefs = appendUnique(currentRefs, current)
				if _, ok := sourceSet[current]; !ok {
					staleRefs = appendUnique(staleRefs, ref)
					reasons = appendUnique(reasons, "current replacement missing from source refs")
				}
			}
		} else {
			currentRefs = appendUnique(currentRefs, ref)
		}

		for _, superseded := range splitPathList(source.Metadata["supersedes"]) {
			supersededRefs = appendUnique(supersededRefs, superseded)
		}
	}

	if len(input.SourceRefs) == 0 {
		reasons = appendUnique(reasons, "missing source refs")
	}

	sort.Strings(resolvedDocRefs)
	sort.Strings(currentRefs)
	sort.Strings(supersededRefs)
	sort.Strings(missingRefs)
	sort.Strings(staleRefs)
	sort.Strings(reasons)
	freshness := "fresh"
	if len(reasons) > 0 || len(missingRefs) > 0 || len(staleRefs) > 0 {
		freshness = "stale"
	}
	freshnessReason := "sources current"
	if len(reasons) > 0 {
		freshnessReason = strings.Join(reasons, ", ")
	}
	details := map[string]string{
		"synthesis_path":         input.Document.Path,
		"source_refs":            strings.Join(input.SourceRefs, ", "),
		"current_source_refs":    strings.Join(currentRefs, ", "),
		"superseded_source_refs": strings.Join(supersededRefs, ", "),
		"missing_source_refs":    strings.Join(missingRefs, ", "),
		"stale_source_refs":      strings.Join(staleRefs, ", "),
		"freshness_reason":       freshnessReason,
	}
	for key, value := range details {
		versionInputs = append(versionInputs, key+":"+value)
	}
	versionInputs = append(versionInputs, "freshness:"+freshness)
	sort.Strings(versionInputs)
	return domain.ProjectionState{
		Projection:        "synthesis",
		RefKind:           "document",
		RefID:             input.Document.DocID,
		SourceRef:         strings.Join(resolvedDocRefs, ", "),
		Freshness:         freshness,
		ProjectionVersion: hashID("synthesis", strings.Join(versionInputs, "|")),
		UpdatedAt:         now,
		Details:           details,
	}
}
