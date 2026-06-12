// Package chronicler implements the first read-only Chronicler orchestration
// layer over the OpenClerk Core runner.
package chronicler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

const (
	SchemaVersion = "openclerk-clerk.v1"
	ActionRun     = "clerk_run"
)

type RunRequest struct {
	InboxPaths []string
	Task       string
	Query      string
	PathPrefix string
	Limit      int
}

type RunEnvelope struct {
	SchemaVersion string    `json:"schema_version"`
	Action        string    `json:"action"`
	Result        RunResult `json:"result"`
}

type RunResult struct {
	Mode             string           `json:"mode"`
	PlannedNoWrite   bool             `json:"planned_no_write"`
	WritesPerformed  int              `json:"writes_performed"`
	InboxCandidates  []InboxCandidate `json:"inbox_candidates"`
	ContextPacks     []ContextPack    `json:"context_packs"`
	StaleSynthesis   []StaleSynthesis `json:"stale_synthesis"`
	DuplicateRisks   []DuplicateRisk  `json:"duplicate_risks"`
	PendingReview    []PendingReview  `json:"pending_review"`
	Blockers         []string         `json:"blockers"`
	AuthorityLimits  string           `json:"authority_limits"`
	ApprovalBoundary string           `json:"approval_boundary"`
	Deferred         []string         `json:"deferred"`
}

type InboxCandidate struct {
	SourceRef         string            `json:"source_ref"`
	SourceFile        string            `json:"source_file"`
	SourceType        string            `json:"source_type,omitempty"`
	ProposedTitle     string            `json:"proposed_title,omitempty"`
	ProposedPath      string            `json:"proposed_path,omitempty"`
	Type              string            `json:"type,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	Summary           string            `json:"summary,omitempty"`
	SourceRefs        []string          `json:"source_refs,omitempty"`
	DuplicateRisk     string            `json:"duplicate_risk"`
	RecommendedAction string            `json:"recommended_action"`
	WriteStatus       string            `json:"write_status"`
	ApprovalBoundary  string            `json:"approval_boundary"`
	MetadataFields    map[string]string `json:"metadata_fields,omitempty"`
}

type ContextPack struct {
	Task                  string            `json:"task"`
	Query                 string            `json:"query"`
	PathPrefix            string            `json:"path_prefix,omitempty"`
	Summary               string            `json:"summary"`
	MustRead              []ContextDocument `json:"must_read"`
	RelevantDecisions     []ContextDecision `json:"relevant_decisions"`
	StaleOrMissingContext []string          `json:"stale_or_missing_context"`
	OpenQuestions         []string          `json:"open_questions"`
	Citations             []runner.Citation `json:"citations"`
	WriteStatus           string            `json:"write_status"`
	ValidationBoundaries  string            `json:"validation_boundaries"`
	AuthorityLimits       string            `json:"authority_limits"`
}

type ContextDocument struct {
	Rank      int               `json:"rank"`
	DocID     string            `json:"doc_id"`
	ChunkID   string            `json:"chunk_id,omitempty"`
	Path      string            `json:"path"`
	Title     string            `json:"title"`
	Snippet   string            `json:"snippet,omitempty"`
	Citations []runner.Citation `json:"citations,omitempty"`
}

type ContextDecision struct {
	DecisionID string            `json:"decision_id"`
	Title      string            `json:"title"`
	Status     string            `json:"status,omitempty"`
	Scope      string            `json:"scope,omitempty"`
	Summary    string            `json:"summary,omitempty"`
	Citations  []runner.Citation `json:"citations,omitempty"`
}

type StaleSynthesis struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type DuplicateRisk struct {
	SourceRef         string `json:"source_ref"`
	DuplicateStatus   string `json:"duplicate_status"`
	LikelyTargetPath  string `json:"likely_target_path,omitempty"`
	RecommendedAction string `json:"recommended_action"`
}

type PendingReview struct {
	SourceRef         string `json:"source_ref"`
	ProposedPath      string `json:"proposed_path,omitempty"`
	RecommendedAction string `json:"recommended_action"`
}

func RunOnce(ctx context.Context, config runclient.Config, request RunRequest) (RunEnvelope, error) {
	result := RunResult{
		Mode:             "once",
		PlannedNoWrite:   true,
		WritesPerformed:  0,
		InboxCandidates:  []InboxCandidate{},
		ContextPacks:     []ContextPack{},
		StaleSynthesis:   []StaleSynthesis{},
		DuplicateRisks:   []DuplicateRisk{},
		PendingReview:    []PendingReview{},
		Blockers:         []string{},
		AuthorityLimits:  "Chronicler is read-only orchestration; OpenClerk Core canonical markdown, citations, provenance, and projection freshness remain authority.",
		ApprovalBoundary: "Chronicler MVP planning is not durable-write approval; future writes must go through approved OpenClerk document lifecycle APIs.",
		Deferred: []string{
			"daemon/watch mode",
			"review approval queue",
			"auto-filing",
			"autonomous routing",
			"broad memory",
		},
	}
	if request.Limit < 0 {
		result.Blockers = append(result.Blockers, "limit must be greater than or equal to 0")
		return runEnvelope(result), nil
	}
	normalizedPathPrefix, pathPrefixBlocker := normalizeContextPathPrefix(request.PathPrefix)
	if pathPrefixBlocker != "" {
		result.Blockers = append(result.Blockers, pathPrefixBlocker)
	} else {
		request.PathPrefix = normalizedPathPrefix
	}

	inboxFiles, blockers := explicitInboxFiles(request.InboxPaths)
	result.Blockers = append(result.Blockers, blockers...)
	for _, inboxFile := range inboxFiles {
		taskResult, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
			Action: runner.DocumentTaskActionArtifactPlan,
			Artifact: runner.ArtifactPlanOptions{
				LocalPath:    inboxFile,
				ArtifactKind: "note",
				Limit:        planLimit(request.Limit),
			},
		})
		if err != nil {
			return RunEnvelope{}, err
		}
		sourceRef := inboxSourceRef(inboxFile)
		if taskResult.Rejected {
			result.Blockers = append(result.Blockers, fmt.Sprintf("%s: %s", sourceRef, taskResult.RejectionReason))
			continue
		}
		if taskResult.ArtifactPlan == nil {
			result.Blockers = append(result.Blockers, fmt.Sprintf("%s: artifact_candidate_plan returned no plan", sourceRef))
			continue
		}
		candidate := inboxCandidateFromPlan(sourceRef, *taskResult.ArtifactPlan)
		result.InboxCandidates = append(result.InboxCandidates, candidate)
		result.PendingReview = append(result.PendingReview, PendingReview{
			SourceRef:         candidate.SourceRef,
			ProposedPath:      candidate.ProposedPath,
			RecommendedAction: candidate.RecommendedAction,
		})
		if isDuplicateRiskStatus(candidate.DuplicateRisk) {
			risk := DuplicateRisk{
				SourceRef:         candidate.SourceRef,
				DuplicateStatus:   candidate.DuplicateRisk,
				RecommendedAction: "resolve_duplicate_before_write",
			}
			if taskResult.ArtifactPlan.LikelyDuplicate != nil && len(taskResult.ArtifactPlan.LikelyDuplicate.Citations) > 0 {
				risk.LikelyTargetPath = taskResult.ArtifactPlan.LikelyDuplicate.Citations[0].Path
			}
			result.DuplicateRisks = append(result.DuplicateRisks, risk)
		}
	}

	query := contextQuery(request)
	if query != "" && pathPrefixBlocker == "" {
		contextPack, err := buildContextPack(ctx, config, request, query)
		if err != nil {
			return RunEnvelope{}, err
		}
		result.ContextPacks = append(result.ContextPacks, contextPack)
	}

	return runEnvelope(result), nil
}

func normalizeContextPathPrefix(pathPrefix string) (string, string) {
	normalized, issue := domain.NormalizeOptionalVaultRelativePrefix(pathPrefix)
	if issue == domain.VaultPathOK {
		return normalized, ""
	}
	return "", "path_prefix must be vault-relative and stay inside the vault root"
}

func runEnvelope(result RunResult) RunEnvelope {
	return RunEnvelope{
		SchemaVersion: SchemaVersion,
		Action:        ActionRun,
		Result:        result,
	}
}

func explicitInboxFiles(inboxPaths []string) ([]string, []string) {
	files := []string{}
	blockers := []string{}
	seen := map[string]bool{}
	for _, rawPath := range inboxPaths {
		inboxPath := strings.TrimSpace(rawPath)
		if inboxPath == "" {
			blockers = append(blockers, "inbox_path is required when supplied")
			continue
		}
		info, err := os.Stat(inboxPath)
		if err != nil {
			blockers = append(blockers, fmt.Sprintf("%s is not readable", inboxSourceRef(inboxPath)))
			continue
		}
		if info.IsDir() {
			entries, err := os.ReadDir(inboxPath)
			if err != nil {
				blockers = append(blockers, fmt.Sprintf("%s is not readable", inboxSourceRef(inboxPath)))
				continue
			}
			sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
			for _, entry := range entries {
				if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
					continue
				}
				candidatePath := filepath.Join(inboxPath, entry.Name())
				if !isPreferredInboxTextPath(candidatePath) {
					continue
				}
				if !seen[candidatePath] {
					seen[candidatePath] = true
					files = append(files, candidatePath)
				}
			}
			continue
		}
		if !info.Mode().IsRegular() {
			blockers = append(blockers, fmt.Sprintf("%s must be a regular file or directory", inboxSourceRef(inboxPath)))
			continue
		}
		if !isPreferredInboxTextPath(inboxPath) {
			blockers = append(blockers, fmt.Sprintf("%s must be markdown or text for the Chronicler MVP", inboxSourceRef(inboxPath)))
			continue
		}
		if !seen[inboxPath] {
			seen[inboxPath] = true
			files = append(files, inboxPath)
		}
	}
	sort.Strings(files)
	return files, blockers
}

func isPreferredInboxTextPath(value string) bool {
	switch strings.ToLower(filepath.Ext(value)) {
	case ".md", ".markdown", ".txt":
		return true
	default:
		return false
	}
}

func inboxSourceRef(value string) string {
	name := strings.TrimSpace(filepath.Base(value))
	if name == "" || name == "." || name == string(filepath.Separator) {
		name = "inbox"
	}
	return "local_inbox:" + name
}

func inboxCandidateFromPlan(sourceRef string, plan runner.ArtifactCandidatePlan) InboxCandidate {
	sourceFile := strings.TrimPrefix(sourceRef, "local_inbox:")
	sourceRefs := []string{sourceRef}
	if plan.LocalArtifact != nil {
		sourceFile = plan.LocalArtifact.FileName
		sourceRefs = append(sourceRefs, plan.LocalArtifact.SourceRef, "sha256:"+plan.LocalArtifact.SHA256)
	}
	recommendedAction := "review_then_approve_create_document"
	if plan.LikelyDuplicate != nil || plan.ExistingSource != nil {
		recommendedAction = "resolve_duplicate_before_write"
	}
	return InboxCandidate{
		SourceRef:         sourceRef,
		SourceFile:        sourceFile,
		SourceType:        plan.SourceType,
		ProposedTitle:     plan.CandidateTitle,
		ProposedPath:      plan.CandidatePath,
		Type:              plan.ArtifactKind,
		Tags:              append([]string(nil), plan.Tags...),
		Summary:           compactSummary(plan.BodyPreview),
		SourceRefs:        sourceRefs,
		DuplicateRisk:     plan.DuplicateStatus,
		RecommendedAction: recommendedAction,
		WriteStatus:       plan.WriteStatus,
		ApprovalBoundary:  plan.ApprovalBoundary,
		MetadataFields:    cloneStringMap(plan.MetadataFields),
	}
}

func isDuplicateRiskStatus(status string) bool {
	switch status {
	case "likely_duplicate_candidate_no_write", "existing_source_url_found_no_write", "likely_duplicate_found":
		return true
	default:
		return false
	}
}

func contextQuery(request RunRequest) string {
	if strings.TrimSpace(request.Query) != "" {
		return strings.TrimSpace(request.Query)
	}
	return strings.TrimSpace(request.Task)
}

func buildContextPack(ctx context.Context, config runclient.Config, request RunRequest, query string) (ContextPack, error) {
	limit := planLimit(request.Limit)
	search, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionSearch,
		Search: runner.SearchOptions{
			Text:       query,
			PathPrefix: strings.TrimSpace(request.PathPrefix),
			Limit:      limit,
		},
	})
	if err != nil {
		return ContextPack{}, err
	}
	if search.Rejected {
		return ContextPack{}, fmt.Errorf("context search rejected: %s", search.RejectionReason)
	}
	decision, err := runner.RunRetrievalTask(ctx, config, runner.RetrievalTaskRequest{
		Action: runner.RetrievalTaskActionDecisionLookup,
		DecisionLookup: runner.DecisionLookupReportOptions{
			Query: query,
			Limit: limit,
		},
	})
	if err != nil {
		return ContextPack{}, err
	}
	if decision.Rejected {
		return ContextPack{}, fmt.Errorf("decision lookup rejected: %s", decision.RejectionReason)
	}

	pack := ContextPack{
		Task:                  strings.TrimSpace(request.Task),
		Query:                 query,
		PathPrefix:            strings.TrimSpace(request.PathPrefix),
		MustRead:              []ContextDocument{},
		RelevantDecisions:     []ContextDecision{},
		StaleOrMissingContext: []string{},
		OpenQuestions:         []string{},
		Citations:             []runner.Citation{},
		WriteStatus:           "read_only_no_write",
		ValidationBoundaries:  "context pack uses existing OpenClerk retrieval reports only; no writes, no autonomous browsing, no direct SQLite, no direct canonical markdown mutation, and no hidden memory",
		AuthorityLimits:       "context pack citations are supporting evidence; canonical markdown, promoted records, provenance, and projection freshness remain authority",
	}
	if search.Search != nil {
		pack.MustRead = contextDocumentsFromSearch(*search.Search)
		pack.Citations = append(pack.Citations, citationsFromSearchHits(search.Search.Hits)...)
	}
	if decision.DecisionLookup != nil {
		pack.RelevantDecisions = contextDecisionsFromReport(*decision.DecisionLookup)
		if pack.PathPrefix != "" {
			pack.RelevantDecisions = filterContextDecisionsByPathPrefix(pack.RelevantDecisions, pack.PathPrefix)
			pack.Citations = append(pack.Citations, citationsFromContextDecisions(pack.RelevantDecisions)...)
		} else {
			pack.Citations = append(pack.Citations, decision.DecisionLookup.Citations...)
		}
	}
	pack.Citations = dedupeCitations(pack.Citations)
	if len(pack.MustRead) == 0 {
		pack.StaleOrMissingContext = append(pack.StaleOrMissingContext, "no search hits returned for context query")
	}
	if len(pack.RelevantDecisions) == 0 {
		pack.OpenQuestions = append(pack.OpenQuestions, "No formal decision evidence was found for this task; verify whether a decision record exists before relying on policy assumptions.")
	}
	pack.Summary = fmt.Sprintf("Read-only context pack for %q with %d must-read documents, %d relevant decisions, and %d citations.", query, len(pack.MustRead), len(pack.RelevantDecisions), len(pack.Citations))
	return pack, nil
}

func contextDocumentsFromSearch(search runner.SearchResult) []ContextDocument {
	result := make([]ContextDocument, 0, len(search.Hits))
	for _, hit := range search.Hits {
		path := ""
		if len(hit.Citations) > 0 {
			path = hit.Citations[0].Path
		}
		result = append(result, ContextDocument{
			Rank:      hit.Rank,
			DocID:     hit.DocID,
			ChunkID:   hit.ChunkID,
			Path:      path,
			Title:     hit.Title,
			Snippet:   hit.Snippet,
			Citations: append([]runner.Citation(nil), hit.Citations...),
		})
	}
	return result
}

func citationsFromSearchHits(hits []runner.SearchHit) []runner.Citation {
	citations := []runner.Citation{}
	for _, hit := range hits {
		citations = append(citations, hit.Citations...)
	}
	return citations
}

func contextDecisionsFromReport(report runner.DecisionLookupReport) []ContextDecision {
	if report.Decisions == nil {
		return nil
	}
	result := make([]ContextDecision, 0, len(report.Decisions.Decisions))
	for _, decision := range report.Decisions.Decisions {
		result = append(result, ContextDecision{
			DecisionID: decision.DecisionID,
			Title:      decision.Title,
			Status:     decision.Status,
			Scope:      decision.Scope,
			Summary:    decision.Summary,
			Citations:  append([]runner.Citation(nil), decision.Citations...),
		})
	}
	return result
}

func filterContextDecisionsByPathPrefix(decisions []ContextDecision, pathPrefix string) []ContextDecision {
	if pathPrefix == "" {
		return decisions
	}
	filtered := make([]ContextDecision, 0, len(decisions))
	for _, decision := range decisions {
		if decisionHasCitationWithPathPrefix(decision, pathPrefix) {
			filtered = append(filtered, decision)
		}
	}
	return filtered
}

func decisionHasCitationWithPathPrefix(decision ContextDecision, pathPrefix string) bool {
	for _, citation := range decision.Citations {
		if strings.HasPrefix(citation.Path, pathPrefix) {
			return true
		}
	}
	return false
}

func citationsFromContextDecisions(decisions []ContextDecision) []runner.Citation {
	citations := []runner.Citation{}
	for _, decision := range decisions {
		citations = append(citations, decision.Citations...)
	}
	return citations
}

func dedupeCitations(citations []runner.Citation) []runner.Citation {
	result := []runner.Citation{}
	seen := map[string]bool{}
	for _, citation := range citations {
		key := fmt.Sprintf("%s\x00%s\x00%d\x00%d", citation.DocID, citation.ChunkID, citation.LineStart, citation.LineEnd)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, citation)
	}
	return result
}

func planLimit(limit int) int {
	if limit == 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func compactSummary(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	const maxRunes = 240
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes]) + "..."
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
