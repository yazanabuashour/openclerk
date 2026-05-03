package runner

import (
	"fmt"
	"strings"
)

func projectionFreshnessSummary(states []ProjectionState) string {
	if len(states) == 0 {
		return "projection freshness unavailable"
	}
	parts := make([]string, 0, len(states))
	for _, state := range states {
		ref := state.RefID
		if state.SourceRef != "" {
			ref = state.SourceRef
		}
		if ref == "" {
			ref = state.RefKind
		}
		parts = append(parts, fmt.Sprintf("%s:%s", ref, state.Freshness))
	}
	return strings.Join(parts, ", ")
}

func projectionListFreshnessSummary(list *ProjectionStateList) string {
	if list == nil {
		return "projection freshness unavailable"
	}
	return projectionFreshnessSummary(list.Projections)
}

func citationPathSummary(citations []Citation) string {
	if len(citations) == 0 {
		return "no citations"
	}
	paths := make([]string, 0, len(citations))
	seen := map[string]struct{}{}
	for _, citation := range citations {
		if citation.Path == "" {
			continue
		}
		if _, ok := seen[citation.Path]; ok {
			continue
		}
		seen[citation.Path] = struct{}{}
		paths = append(paths, citation.Path)
	}
	if len(paths) == 0 {
		return "citations without paths"
	}
	return strings.Join(paths, ", ")
}
