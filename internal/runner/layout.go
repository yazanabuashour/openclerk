package runner

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	layoutModeConventionFirst = "convention_first"
	layoutCheckPass           = "pass"
	layoutCheckWarn           = "warn"
	layoutCheckFail           = "fail"

	sourcesPathPrefix         = "sources/"
	synthesisPathPrefix       = "synthesis/"
	recordsPathPrefix         = "records/"
	serviceRecordsPathPrefix  = "records/services/"
	decisionRecordsPathPrefix = "records/decisions/"
)

func inspectKnowledgeLayout(ctx context.Context, client *runclient.Client) (KnowledgeLayout, error) {
	documents, err := listAllDocuments(ctx, client)
	if err != nil {
		return KnowledgeLayout{}, err
	}
	paths := toPaths(client.Paths())
	layout := KnowledgeLayout{
		Valid:                  true,
		Mode:                   layoutModeConventionFirst,
		ConfigArtifactRequired: false,
		ConfigArtifact:         "none",
		Paths:                  paths,
		ConventionalPaths:      layoutPathConventions(),
		DocumentKinds:          layoutDocumentKinds(),
		Checks: []KnowledgeLayoutCheck{
			{
				ID:      "paths_resolved",
				Status:  layoutCheckPass,
				Message: "Resolved effective OpenClerk data, database, and vault paths through the runner.",
			},
			{
				ID:      "configuration_model",
				Status:  layoutCheckPass,
				Message: "Layout is convention-first; no committed manifest or config artifact is required for v1.",
			},
		},
	}

	if len(documents) == 0 {
		layout.Checks = append(layout.Checks, KnowledgeLayoutCheck{
			ID:      "vault_documents_present",
			Status:  layoutCheckWarn,
			Message: "No registered canonical markdown documents were found; an empty vault is allowed but not production-complete.",
		})
		return layout, nil
	}

	documentByPath := make(map[string]DocumentSummary, len(documents))
	prefixCounts := map[string]int{}
	for _, document := range documents {
		documentByPath[document.Path] = document
		for _, convention := range layout.ConventionalPaths {
			if strings.HasPrefix(document.Path, convention.PathPrefix) {
				prefixCounts[convention.PathPrefix]++
			}
		}
	}

	layout.Checks = append(layout.Checks, KnowledgeLayoutCheck{
		ID:      "vault_documents_present",
		Status:  layoutCheckPass,
		Message: fmt.Sprintf("Found %d registered canonical markdown document(s).", len(documents)),
	})

	missingOptional := []string{}
	for _, convention := range layout.ConventionalPaths {
		if convention.Required || prefixCounts[convention.PathPrefix] > 0 {
			continue
		}
		missingOptional = append(missingOptional, convention.PathPrefix)
	}
	if len(missingOptional) > 0 {
		layout.Checks = append(layout.Checks, KnowledgeLayoutCheck{
			ID:      "optional_conventional_prefixes",
			Status:  layoutCheckWarn,
			Message: "Some optional conventional prefixes have no registered documents.",
			Details: map[string]string{
				"path_prefixes": strings.Join(missingOptional, ", "),
			},
		})
	}

	for _, summary := range documents {
		detailed, err := client.GetDocument(ctx, summary.DocID)
		if err != nil {
			return KnowledgeLayout{}, err
		}
		document := toDocument(detailed)
		layout.Checks = append(layout.Checks, inspectDocumentLayout(document, documentByPath)...)
	}

	layout.Valid = layoutChecksValid(layout.Checks)
	sort.SliceStable(layout.Checks, func(i, j int) bool {
		if layout.Checks[i].Status != layout.Checks[j].Status {
			return layoutStatusRank(layout.Checks[i].Status) < layoutStatusRank(layout.Checks[j].Status)
		}
		if layout.Checks[i].Path != layout.Checks[j].Path {
			return layout.Checks[i].Path < layout.Checks[j].Path
		}
		return layout.Checks[i].ID < layout.Checks[j].ID
	})
	return layout, nil
}

func listAllDocuments(ctx context.Context, client *runclient.Client) ([]DocumentSummary, error) {
	documents := []DocumentSummary{}
	cursor := ""
	for {
		result, err := client.ListDocuments(ctx, runclient.DocumentListOptions{Limit: 100, Cursor: cursor})
		if err != nil {
			return nil, err
		}
		documents = append(documents, toDocumentSummaries(result.Documents)...)
		if !result.PageInfo.HasMore {
			return documents, nil
		}
		cursor = result.PageInfo.NextCursor
	}
}

func inspectDocumentLayout(document Document, documentByPath map[string]DocumentSummary) []KnowledgeLayoutCheck {
	checks := []KnowledgeLayoutCheck{}
	if isSynthesisDocument(document) {
		checks = append(checks, inspectSynthesisDocumentLayout(document, documentByPath)...)
	}
	if isServiceDocumentCandidate(document) {
		checks = append(checks, inspectServiceDocumentLayout(document)...)
	} else if isDecisionDocumentCandidate(document) {
		checks = append(checks, inspectDecisionDocumentLayout(document)...)
	} else if isRecordDocumentCandidate(document) {
		checks = append(checks, inspectRecordDocumentLayout(document)...)
	}
	return checks
}

func inspectSynthesisDocumentLayout(document Document, documentByPath map[string]DocumentSummary) []KnowledgeLayoutCheck {
	checks := []KnowledgeLayoutCheck{}
	sourceRefs := splitLayoutCSV(document.Metadata["source_refs"])
	if len(sourceRefs) == 0 {
		checks = append(checks, failLayoutDocumentCheck("synthesis_source_refs", document, "Synthesis documents must declare single-line comma-separated source_refs frontmatter.", nil))
	} else {
		missingRefs := []string{}
		for _, sourceRef := range sourceRefs {
			if _, ok := documentByPath[sourceRef]; !ok {
				missingRefs = append(missingRefs, sourceRef)
			}
		}
		if len(missingRefs) > 0 {
			checks = append(checks, failLayoutDocumentCheck("synthesis_source_refs_resolve", document, "Synthesis source_refs must resolve to registered canonical source paths.", map[string]string{
				"missing_source_refs": strings.Join(missingRefs, ", "),
			}))
		} else {
			checks = append(checks, passLayoutDocumentCheck("synthesis_source_refs_resolve", document, "Synthesis source_refs resolve to registered canonical source paths.", map[string]string{
				"source_refs": strings.Join(sourceRefs, ", "),
			}))
		}
	}

	for _, heading := range []string{"Sources", "Freshness"} {
		id := "synthesis_" + strings.ToLower(heading) + "_section"
		if !layoutBodyContainsHeadingLevel(document.Body, heading, 2) {
			checks = append(checks, failLayoutDocumentCheck(id, document, "Synthesis documents must include a ## "+heading+" section.", nil))
			continue
		}
		checks = append(checks, passLayoutDocumentCheck(id, document, "Synthesis document includes a ## "+heading+" section.", nil))
	}
	return checks
}

func inspectRecordDocumentLayout(document Document) []KnowledgeLayoutCheck {
	required := []string{"entity_type", "entity_name"}
	missing := missingMetadataKeys(document.Metadata, required)
	if len(missing) > 0 {
		return []KnowledgeLayoutCheck{failLayoutDocumentCheck("record_identity_metadata", document, "Record-shaped documents must include entity_type and entity_name metadata.", map[string]string{
			"missing": strings.Join(missing, ", "),
		})}
	}
	return []KnowledgeLayoutCheck{passLayoutDocumentCheck("record_identity_metadata", document, "Record-shaped document identity metadata is complete.", nil)}
}

func inspectServiceDocumentLayout(document Document) []KnowledgeLayoutCheck {
	required := []string{"service_id", "service_name"}
	if strings.EqualFold(document.Metadata["entity_type"], "service") && document.Metadata["service_id"] == "" && document.Metadata["service_name"] == "" {
		required = []string{"entity_id", "entity_name"}
	}
	missing := missingMetadataKeys(document.Metadata, required)
	if len(missing) > 0 {
		return []KnowledgeLayoutCheck{failLayoutDocumentCheck("service_identity_metadata", document, "Service-shaped documents must include service identity metadata.", map[string]string{
			"missing": strings.Join(missing, ", "),
		})}
	}
	return []KnowledgeLayoutCheck{passLayoutDocumentCheck("service_identity_metadata", document, "Service-shaped document identity metadata is complete.", nil)}
}

func inspectDecisionDocumentLayout(document Document) []KnowledgeLayoutCheck {
	required := []string{"decision_id", "decision_title", "decision_status"}
	missing := missingMetadataKeys(document.Metadata, required)
	if len(missing) > 0 {
		return []KnowledgeLayoutCheck{failLayoutDocumentCheck("decision_identity_metadata", document, "Decision-shaped documents must include decision identity and status metadata.", map[string]string{
			"missing": strings.Join(missing, ", "),
		})}
	}
	return []KnowledgeLayoutCheck{passLayoutDocumentCheck("decision_identity_metadata", document, "Decision-shaped document identity and status metadata is complete.", nil)}
}

func layoutPathConventions() []LayoutPathConvention {
	return []LayoutPathConvention{
		{Name: "vault_root", PathPrefix: "", Description: "Canonical markdown root resolved by the runner.", Required: true},
		{Name: "canonical_sources", PathPrefix: sourcesPathPrefix, Description: "Home for canonical source documents under the configured vault root.", Required: false},
		{Name: "source_linked_synthesis", PathPrefix: synthesisPathPrefix, Description: "Home for durable source-linked synthesis documents under the configured vault root.", Required: false},
		{Name: "generic_records", PathPrefix: recordsPathPrefix, Description: "Conventional home for promoted record-shaped canonical documents.", Required: false},
		{Name: "service_records", PathPrefix: serviceRecordsPathPrefix, Description: "Conventional home for service registry records.", Required: false},
		{Name: "decision_records", PathPrefix: decisionRecordsPathPrefix, Description: "Conventional home for decision and architecture records.", Required: false},
	}
}

func layoutDocumentKinds() []LayoutDocumentKind {
	return []LayoutDocumentKind{
		{Kind: "canonical_doc", Description: "Any registered markdown document under the runner-resolved vault root.", Selectors: []string{"*.md"}},
		{Kind: "canonical_source_doc", Description: "Canonical source authority document, conventionally under sources/.", Selectors: []string{"path_prefix:sources/"}},
		{Kind: "synthesis_doc", Description: "Durable source-linked synthesis that remains subordinate to canonical sources.", Selectors: []string{"path_prefix:synthesis/", "metadata:type=synthesis"}, Required: []string{"source_refs", "## Sources", "## Freshness"}},
		{Kind: "record_doc", Description: "Canonical markdown document that feeds the generic promoted records projection.", Selectors: []string{"path_prefix:records/", "metadata:entity_type", "metadata:entity_name"}, Required: []string{"entity_type", "entity_name"}},
		{Kind: "service_doc", Description: "Canonical markdown document that feeds the service registry projection.", Selectors: []string{"path_prefix:records/services/", "metadata:service_id", "metadata:service_name"}, Required: []string{"service_id", "service_name"}},
		{Kind: "decision_doc", Description: "Canonical markdown document that feeds the decision records projection.", Selectors: []string{"path_prefix:records/decisions/", "metadata:decision_id", "metadata:decision_title", "metadata:decision_status"}, Required: []string{"decision_id", "decision_title", "decision_status"}},
	}
}

func layoutChecksValid(checks []KnowledgeLayoutCheck) bool {
	for _, check := range checks {
		if check.Status == layoutCheckFail {
			return false
		}
	}
	return true
}

func layoutStatusRank(status string) int {
	switch status {
	case layoutCheckFail:
		return 0
	case layoutCheckWarn:
		return 1
	default:
		return 2
	}
}

func isSynthesisDocument(document Document) bool {
	return (isSynthesisPath(document.Path) && !isIndexDocumentPath(document.Path)) ||
		strings.EqualFold(strings.TrimSpace(document.Metadata["type"]), "synthesis")
}

func isSynthesisPath(docPath string) bool {
	return strings.HasPrefix(docPath, synthesisPathPrefix)
}

func isIndexDocumentPath(docPath string) bool {
	return path.Base(docPath) == "_index.md"
}

func isRecordDocumentCandidate(document Document) bool {
	return strings.HasPrefix(document.Path, recordsPathPrefix) ||
		document.Metadata["entity_id"] != "" ||
		document.Metadata["entity_type"] != "" ||
		document.Metadata["entity_name"] != ""
}

func isServiceDocumentCandidate(document Document) bool {
	return strings.HasPrefix(document.Path, serviceRecordsPathPrefix) ||
		document.Metadata["service_id"] != "" ||
		document.Metadata["service_name"] != "" ||
		document.Metadata["service_status"] != "" ||
		document.Metadata["service_owner"] != "" ||
		document.Metadata["service_interface"] != "" ||
		strings.EqualFold(strings.TrimSpace(document.Metadata["entity_type"]), "service")
}

func isDecisionDocumentCandidate(document Document) bool {
	return strings.HasPrefix(document.Path, decisionRecordsPathPrefix) ||
		document.Metadata["decision_id"] != "" ||
		document.Metadata["decision_title"] != "" ||
		document.Metadata["decision_status"] != ""
}

func layoutBodyContainsHeadingLevel(body string, want string, level int) bool {
	prefix := strings.Repeat("#", level)
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, prefix+" ") {
			continue
		}
		if len(trimmed) > level && trimmed[level] == '#' {
			continue
		}
		heading := strings.TrimSpace(trimmed[level:])
		heading = strings.TrimSpace(strings.TrimRight(heading, "#"))
		if strings.EqualFold(heading, want) {
			return true
		}
	}
	return false
}

func splitLayoutCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := []string{}
	seen := map[string]struct{}{}
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean == "" {
			continue
		}
		clean = path.Clean(strings.ReplaceAll(clean, "\\", "/"))
		if clean == "." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
			out = append(out, clean)
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}

func missingMetadataKeys(metadata map[string]string, required []string) []string {
	missing := []string{}
	for _, key := range required {
		if strings.TrimSpace(metadata[key]) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func passLayoutDocumentCheck(id string, document Document, message string, details map[string]string) KnowledgeLayoutCheck {
	return layoutDocumentCheck(id, layoutCheckPass, document, message, details)
}

func failLayoutDocumentCheck(id string, document Document, message string, details map[string]string) KnowledgeLayoutCheck {
	return layoutDocumentCheck(id, layoutCheckFail, document, message, details)
}

func layoutDocumentCheck(id string, status string, document Document, message string, details map[string]string) KnowledgeLayoutCheck {
	return KnowledgeLayoutCheck{
		ID:      id,
		Status:  status,
		Message: message,
		Path:    document.Path,
		DocID:   document.DocID,
		Details: details,
	}
}
