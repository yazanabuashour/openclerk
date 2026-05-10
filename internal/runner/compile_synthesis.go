package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func runCompileSynthesis(ctx context.Context, client *runclient.Client, input CompileSynthesisInput) (CompileSynthesisResult, error) {
	sourceEvidence, sourceRoleFacts, err := compileSynthesisSourceEvidence(ctx, client, input.SourceRefs)
	if err != nil {
		return CompileSynthesisResult{}, err
	}
	body := compileSynthesisBody(compileSynthesisInputWithSourceRoleFacts(input, sourceRoleFacts))
	candidateInspection, err := inspectSynthesisCandidates(ctx, client, input.Path)
	if err != nil {
		return CompileSynthesisResult{}, err
	}
	candidates := candidateInspection.Paths
	matches := candidateInspection.TargetMatches
	if len(matches) > 1 {
		validationBoundaries := compileSynthesisValidationBoundaries()
		authorityLimits := compileSynthesisAuthorityLimits()
		return CompileSynthesisResult{
			SelectedPath:         input.Path,
			SourceRefs:           input.SourceRefs,
			CandidateStatus:      "blocked_duplicate_target",
			DuplicateStatus:      "duplicate_target_path_detected",
			WriteStatus:          "skipped",
			ValidationBoundaries: validationBoundaries,
			AuthorityLimits:      authorityLimits,
			AgentHandoff: compileSynthesisHandoff(
				input.Path,
				input.SourceRefs,
				"blocked duplicate synthesis target; no write applied",
				[]string{"duplicate_status=duplicate_target_path_detected"},
				validationBoundaries,
				authorityLimits,
				"required: resolve duplicate synthesis target before retrying compile_synthesis",
			),
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

	validationBoundaries := compileSynthesisValidationBoundaries()
	authorityLimits := compileSynthesisAuthorityLimits()
	projectionSummary := projectionFreshnessSummary(projectionFreshness)
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
		ValidationBoundaries: validationBoundaries,
		AuthorityLimits:      authorityLimits,
		AgentHandoff: compileSynthesisHandoff(
			document.Path,
			input.SourceRefs,
			fmt.Sprintf("compile_synthesis %s %s with %s; %s", writeStatus, document.Path, strings.Join(input.SourceRefs, ", "), projectionSummary),
			[]string{
				"selected_path=" + document.Path,
				"source_refs=" + strings.Join(input.SourceRefs, ", "),
				"duplicate_status=" + duplicateStatus,
				"provenance_refs=" + strings.Join(provenanceRefs, ", "),
				"projection_freshness=" + projectionSummary,
				"write_status=" + writeStatus,
			},
			validationBoundaries,
			authorityLimits,
			"not required for routine answer; use primitives only for explicit follow-up inspection or runner rejection repair",
		),
	}, nil
}

func runValidationSynthesis(ctx context.Context, client *runclient.Client, input ValidationSynthesisInput) (CompileSynthesisResult, error) {
	if err := validateDisposableValidationRuntime(ctx, client); err != nil {
		return CompileSynthesisResult{}, err
	}
	if input.DocID != "" {
		document, err := client.GetDocument(ctx, input.DocID)
		if err != nil {
			return CompileSynthesisResult{}, err
		}
		if input.Path == "" {
			input.Path = document.Path
		}
		if input.Title == "" {
			input.Title = document.Title
		}
	}
	synthesisInput := CompileSynthesisInput{
		Path:          firstNonEmpty(input.Path, "synthesis/routine-ux-validation.md"),
		Title:         firstNonEmpty(input.Title, "Routine UX Validation Synthesis"),
		SourceRefs:    input.SourceRefs,
		Body:          input.Body,
		BodyFacts:     input.BodyFacts,
		FreshnessNote: firstNonEmpty(input.FreshnessNote, "Checked against disposable validation source evidence through validation_synthesis_report."),
		Mode:          "create_or_update",
	}
	if len(synthesisInput.SourceRefs) == 0 {
		synthesisInput.SourceRefs = []string{"sources/routine-ux-validation/source.md"}
	}
	if strings.TrimSpace(synthesisInput.Body) == "" && len(synthesisInput.BodyFacts) == 0 {
		synthesisInput.BodyFacts = []string{"Validation synthesis refreshed through the runner-owned disposable validation workflow."}
	}
	result, err := runCompileSynthesis(ctx, client, trimCompileSynthesisInput(synthesisInput))
	if err != nil {
		return CompileSynthesisResult{}, err
	}
	result.ValidationBoundaries = "runner-owned validation_synthesis_report workflow for disposable validation copies; no live private vault mutation, broad repo search, direct vault inspection, direct file edits, direct SQLite, source-built runners, HTTP/MCP bypasses, unsupported transports, duplicate synthesis creation, or hidden authority promotion"
	result.AgentHandoff = compileSynthesisHandoff(
		result.SelectedPath,
		result.SourceRefs,
		fmt.Sprintf("validation_synthesis_report %s %s with disposable validation source refs; %s", result.WriteStatus, result.SelectedPath, projectionFreshnessSummary(result.ProjectionFreshness)),
		[]string{
			"selected_path=" + result.SelectedPath,
			"source_refs=" + strings.Join(result.SourceRefs, ", "),
			"duplicate_status=" + result.DuplicateStatus,
			"provenance_refs=" + strings.Join(result.ProvenanceRefs, ", "),
			"projection_freshness=" + projectionFreshnessSummary(result.ProjectionFreshness),
			"write_status=" + result.WriteStatus,
			"live_private_vault_mutated=false",
		},
		result.ValidationBoundaries,
		result.AuthorityLimits,
		"not required for routine validation answer; use compile_synthesis or primitives only for explicit drill-down or runner rejection repair",
	)
	return result, nil
}

func validateDisposableValidationRuntime(ctx context.Context, client *runclient.Client) error {
	vaultRoot := filepath.Clean(client.Paths().VaultRoot)
	if filepath.Base(vaultRoot) != "private-vault-copy" {
		return domain.ValidationError("validation_synthesis_report requires the routine UX disposable vault copy", nil)
	}
	source, ok, err := auditDocumentByPath(ctx, client, "sources/routine-ux-validation/source.md")
	if err != nil {
		return err
	}
	if !ok || !strings.Contains(source.Body, "disposable source exists only inside the routine UX telemetry vault copy") {
		return domain.ValidationError("validation_synthesis_report requires disposable validation source marker", nil)
	}
	return nil
}

func compileSynthesisBody(input CompileSynthesisInput) string {
	content := compileSynthesisBodyContent(input)
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

func compileSynthesisBodyContent(input CompileSynthesisInput) string {
	if body := strings.TrimSpace(input.Body); body != "" {
		stripped := stripFrontmatter(body)
		if compileSynthesisBodyHasRequiredSections(stripped) {
			return stripped
		}
		return compileSynthesisAssembledContent(input.Title, input.SourceRefs, []string{stripped}, input.FreshnessNote, false)
	}
	return compileSynthesisAssembledContent(input.Title, input.SourceRefs, input.BodyFacts, input.FreshnessNote, true)
}

func compileSynthesisAssembledContent(title string, sourceRefs []string, summaryItems []string, freshnessNote string, bulletSummary bool) string {
	lines := []string{
		"# " + title,
		"",
		"## Summary",
	}
	for _, item := range summaryItems {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if bulletSummary {
			lines = append(lines, "- "+trimmed)
		} else {
			lines = append(lines, trimmed)
		}
	}
	lines = append(lines, "", "## Sources")
	for _, sourceRef := range sourceRefs {
		lines = append(lines, "- "+sourceRef)
	}
	if freshnessNote == "" {
		freshnessNote = "Checked current source evidence through compile_synthesis."
	}
	lines = append(lines, "", "## Freshness", freshnessNote)
	return strings.Join(lines, "\n")
}

func compileSynthesisInputWithSourceRoleFacts(input CompileSynthesisInput, sourceRoleFacts []string) CompileSynthesisInput {
	if len(sourceRoleFacts) == 0 {
		return input
	}
	if body := strings.TrimSpace(input.Body); body != "" {
		for _, fact := range sourceRoleFacts {
			if !strings.Contains(body, fact) {
				body += "\n" + fact
			}
		}
		input.Body = body
		return input
	}
	for _, fact := range sourceRoleFacts {
		if !stringSliceContains(input.BodyFacts, fact) {
			input.BodyFacts = append(input.BodyFacts, fact)
		}
	}
	return input
}

func compileSynthesisBodyHasRequiredSections(body string) bool {
	hasSources := strings.Contains(body, "\n## Sources") || strings.HasPrefix(body, "## Sources")
	hasFreshness := strings.Contains(body, "\n## Freshness") || strings.HasPrefix(body, "## Freshness")
	return hasSources && hasFreshness
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

func compileSynthesisSourceEvidence(ctx context.Context, client *runclient.Client, sourceRefs []string) ([]Citation, []string, error) {
	citations := []Citation{}
	roleFacts := []string{}
	for _, ref := range sourceRefs {
		document, ok, err := auditDocumentByPath(ctx, client, ref)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			citations = append(citations, Citation{Path: ref})
			continue
		}
		citations = append(citations, Citation{
			DocID: document.DocID,
			Path:  document.Path,
		})
		switch {
		case strings.EqualFold(strings.TrimSpace(document.Metadata["status"]), "superseded") ||
			strings.TrimSpace(document.Metadata["superseded_by"]) != "":
			roleFacts = appendUniqueString(roleFacts, "Superseded source: "+document.Path)
		case strings.TrimSpace(document.Metadata["supersedes"]) != "":
			roleFacts = appendUniqueString(roleFacts, "Current source: "+document.Path)
			if summaryFact := compileSynthesisCurrentSourceSummaryFact(document.Body); summaryFact != "" {
				roleFacts = appendUniqueString(roleFacts, summaryFact)
			}
		}
	}
	return citations, roleFacts, nil
}

func compileSynthesisCurrentSourceSummaryFact(body string) string {
	summary := compileSynthesisSummarySection(body)
	const revisitPrefix = "Current compile_synthesis revisit guidance says "
	if strings.HasPrefix(strings.ToLower(summary), strings.ToLower(revisitPrefix)) {
		return "Current compile_synthesis revisit decision: " + strings.TrimSpace(summary[len(revisitPrefix):])
	}
	return ""
}

func compileSynthesisSummarySection(body string) string {
	lines := strings.Split(stripFrontmatter(body), "\n")
	inSummary := false
	summaryLines := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			if inSummary {
				break
			}
			inSummary = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")), "Summary")
			continue
		}
		if inSummary && trimmed != "" {
			summaryLines = append(summaryLines, strings.TrimPrefix(trimmed, "- "))
		}
	}
	return strings.TrimSpace(strings.Join(summaryLines, " "))
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

func compileSynthesisHandoff(path string, sourceRefs []string, answerSummary string, evidence []string, validationBoundaries string, authorityLimits string, followUp string) *AgentHandoff {
	if answerSummary == "" {
		answerSummary = fmt.Sprintf("compile_synthesis selected %s from %s", path, strings.Join(sourceRefs, ", "))
	}
	return &AgentHandoff{
		AnswerSummary:               answerSummary,
		Evidence:                    evidence,
		ValidationBoundaries:        validationBoundaries,
		AuthorityLimits:             authorityLimits,
		FollowUpPrimitiveInspection: followUp,
	}
}

func stringSliceContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
