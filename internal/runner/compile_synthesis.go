package runner

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func runCompileSynthesis(ctx context.Context, client *runclient.Client, input CompileSynthesisInput) (CompileSynthesisResult, error) {
	body := compileSynthesisBody(input)
	candidates, matches, err := compileSynthesisCandidates(ctx, client, input.Path)
	if err != nil {
		return CompileSynthesisResult{}, err
	}
	if len(matches) > 1 {
		return CompileSynthesisResult{
			SelectedPath:         input.Path,
			SourceRefs:           input.SourceRefs,
			CandidateStatus:      "blocked_duplicate_target",
			DuplicateStatus:      "duplicate_target_path_detected",
			WriteStatus:          "skipped",
			ValidationBoundaries: compileSynthesisValidationBoundaries(),
			AuthorityLimits:      compileSynthesisAuthorityLimits(),
		}, nil
	}

	var document domain.Document
	writeStatus := "created"
	existingCandidate := false
	if len(matches) == 0 {
		document, err = client.CreateDocument(ctx, domain.CreateDocumentInput{
			Path:  input.Path,
			Title: input.Title,
			Body:  body,
		})
	} else {
		existingCandidate = true
		writeStatus = "updated"
		document, err = client.ReplaceDocument(ctx, matches[0].DocID, domain.ReplaceDocumentInput{
			Title: input.Title,
			Body:  body,
		})
	}
	if err != nil {
		return CompileSynthesisResult{}, err
	}

	sourceEvidence, err := compileSynthesisSourceEvidence(ctx, client, input.SourceRefs)
	if err != nil {
		return CompileSynthesisResult{}, err
	}
	projectionFreshness, err := compileSynthesisProjectionFreshness(ctx, client, document.DocID)
	if err != nil {
		return CompileSynthesisResult{}, err
	}
	provenanceRefs, err := compileSynthesisProvenanceRefs(ctx, client, document.DocID, input.SourceRefs)
	if err != nil {
		return CompileSynthesisResult{}, err
	}

	duplicateStatus := "no_duplicate_created"
	if existingCandidate {
		duplicateStatus = "existing_target_selected_no_duplicate_created"
	}
	if !stringSliceContains(candidates, input.Path) {
		candidates = append(candidates, input.Path)
		sort.Strings(candidates)
	}

	return CompileSynthesisResult{
		SelectedPath:         document.Path,
		DocumentID:           document.DocID,
		ExistingCandidate:    existingCandidate,
		SourceRefs:           input.SourceRefs,
		SourceEvidence:       sourceEvidence,
		CandidateStatus:      fmt.Sprintf("%s; candidates inspected: %s", writeStatus, strings.Join(candidates, ", ")),
		DuplicateStatus:      duplicateStatus,
		ProvenanceRefs:       provenanceRefs,
		ProjectionFreshness:  projectionFreshness,
		WriteStatus:          writeStatus,
		ValidationBoundaries: compileSynthesisValidationBoundaries(),
		AuthorityLimits:      compileSynthesisAuthorityLimits(),
	}, nil
}

func compileSynthesisBody(input CompileSynthesisInput) string {
	content := stripFrontmatter(strings.TrimSpace(input.Body))
	frontmatter := strings.Join([]string{
		"---",
		"type: synthesis",
		"status: active",
		"freshness: fresh",
		"source_refs: " + strings.Join(input.SourceRefs, ", "),
		"---",
		"",
	}, "\n")
	return strings.TrimRight(frontmatter+content, "\n") + "\n"
}

func stripFrontmatter(body string) string {
	lines := strings.Split(body, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return body
	}
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			return strings.TrimSpace(strings.Join(lines[idx+1:], "\n"))
		}
	}
	return body
}

func compileSynthesisCandidates(ctx context.Context, client *runclient.Client, targetPath string) ([]string, []domain.DocumentSummary, error) {
	candidatePaths := []string{}
	targetMatches := []domain.DocumentSummary{}
	cursor := ""
	for {
		list, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix: "synthesis/",
			Limit:      100,
			Cursor:     cursor,
		})
		if err != nil {
			return nil, nil, err
		}
		for _, document := range list.Documents {
			candidatePaths = appendUniqueString(candidatePaths, document.Path)
			if document.Path == targetPath {
				targetMatches = append(targetMatches, document)
			}
		}
		if !list.PageInfo.HasMore {
			break
		}
		cursor = list.PageInfo.NextCursor
		if cursor == "" {
			return nil, nil, fmt.Errorf("list synthesis candidates did not return next cursor")
		}
	}
	sort.Strings(candidatePaths)
	return candidatePaths, targetMatches, nil
}

func compileSynthesisSourceEvidence(ctx context.Context, client *runclient.Client, sourceRefs []string) ([]Citation, error) {
	citations := []Citation{}
	for _, ref := range sourceRefs {
		document, ok, err := auditDocumentByPath(ctx, client, ref)
		if err != nil {
			return nil, err
		}
		if !ok {
			citations = append(citations, Citation{Path: ref})
			continue
		}
		citations = append(citations, Citation{
			DocID: document.DocID,
			Path:  document.Path,
		})
	}
	return citations, nil
}

func compileSynthesisProjectionFreshness(ctx context.Context, client *runclient.Client, docID string) ([]ProjectionState, error) {
	states, err := client.ListProjectionStates(ctx, domain.ProjectionStateQuery{
		Projection: "synthesis",
		RefKind:    "document",
		RefID:      docID,
		Limit:      10,
	})
	if err != nil {
		return nil, err
	}
	return toProjectionStates(states.Projections), nil
}

func compileSynthesisProvenanceRefs(ctx context.Context, client *runclient.Client, docID string, sourceRefs []string) ([]string, error) {
	refs := []string{"document:" + docID, "projection:synthesis:" + docID}
	events, err := client.ListProvenanceEvents(ctx, domain.ProvenanceEventQuery{
		RefKind: "projection",
		RefID:   "synthesis:" + docID,
		Limit:   10,
	})
	if err != nil {
		return nil, err
	}
	for _, event := range events.Events {
		refs = appendUniqueString(refs, event.EventType+":"+event.EventID)
	}
	for _, sourceRef := range sourceRefs {
		refs = appendUniqueString(refs, "source_ref:"+sourceRef)
	}
	sort.Strings(refs)
	return refs, nil
}

func compileSynthesisValidationBoundaries() string {
	return "runner-owned compile_synthesis workflow; no broad repo search, direct vault inspection, direct file edits, direct SQLite, source-built runners, HTTP/MCP bypasses, unsupported transports, duplicate synthesis creation, or hidden authority promotion"
}

func compileSynthesisAuthorityLimits() string {
	return "canonical source documents and promoted records remain authority; synthesis is derived evidence with source refs, provenance, and projection freshness"
}

func stringSliceContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
