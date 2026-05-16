package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	searchDiagnosticsValidationBoundaries = "read-only diagnostics report; runs current lexical search and inspects module configuration only; it does not execute semantic modules, create embeddings, change default ranking, write documents, read raw vault files, query SQLite directly, call HTTP/MCP, or bypass the installed runner"
	searchDiagnosticsAuthorityLimits      = "diagnostics guide mode selection only; source-sensitive answers must come from search or explicit semantic_search results with citations, provenance, projection freshness, and canonical markdown authority"
)

func runSearchDiagnosticsReport(ctx context.Context, client *runclient.Client, config runclient.Config, options SearchDiagnosticsOptions) (SearchDiagnosticsReport, error) {
	limit := defaultRunnerLimit(options.Limit, 10)
	search, err := client.Search(ctx, domain.SearchQuery{
		Text:          options.Query,
		PathPrefix:    options.PathPrefix,
		MetadataKey:   options.MetadataKey,
		MetadataValue: options.MetadataValue,
		Tag:           options.Tag,
		Limit:         limit,
	})
	if err != nil {
		return SearchDiagnosticsReport{}, err
	}
	convertedSearch := toSearchResult(search)
	modules := searchModulePostures(ctx, config)
	recommendedAction, reason := searchDiagnosticsRecommendation(options, convertedSearch, modules)
	report := SearchDiagnosticsReport{
		Query:                  options.Query,
		Intent:                 options.Intent,
		PathPrefix:             options.PathPrefix,
		Tag:                    options.Tag,
		MetadataKey:            options.MetadataKey,
		MetadataValue:          options.MetadataValue,
		LexicalSearch:          &convertedSearch,
		RecommendedAction:      recommendedAction,
		RecommendationReason:   reason,
		ModePostures:           searchModePostures(),
		ModulePostures:         modules,
		TuningVisibility:       searchTuningVisibility(options, limit),
		NoDefaultRankingChange: true,
		CostPosture:            "lexical search has local SQLite cost only; semantic_search can add local model latency for Ollama or explicit remote API cost/egress for Gemini",
		LatencyPosture:         "diagnostics performs one bounded lexical search; semantic_search latency depends on module process startup, embedding provider, cache state, and result limit",
		ValidationBoundaries:   searchDiagnosticsValidationBoundaries,
		AuthorityLimits:        searchDiagnosticsAuthorityLimits,
	}
	report.AgentHandoff = &AgentHandoff{
		AnswerSummary:               fmt.Sprintf("use %s for this request; %s", report.RecommendedAction, report.RecommendationReason),
		Evidence:                    []string{"query=" + report.Query, fmt.Sprintf("lexical_hits=%d", len(convertedSearch.Hits)), "recommended_action=" + report.RecommendedAction, "no_default_ranking_change=true"},
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "run the recommended search or semantic_search request next; inspect provenance_events and projection_states before durable writes or authority claims",
	}
	return report, nil
}

func searchDiagnosticsRecommendation(options SearchDiagnosticsOptions, lexical SearchResult, modules []SearchModulePosture) (string, string) {
	provider := options.Provider
	if provider == "" {
		provider = runclient.SemanticModuleProviderOllama
	}
	semanticIntent := searchDiagnosticsSemanticIntent(options.Intent + " " + options.Query)
	if !semanticIntent {
		return RetrievalTaskActionSearch, "lexical search remains the default for citation-bearing exact/source-sensitive retrieval"
	}
	posture := searchModulePostureForProvider(modules, provider)
	if posture != nil && posture.Readiness == "ready" {
		return RetrievalTaskActionSemanticSearch, "semantic intent is present and the requested provider module is ready; keep it explicit and citation-bearing"
	}
	if len(lexical.Hits) == 0 {
		return RetrievalTaskActionSearch, "semantic intent is present but no verified enabled semantic module is ready; configure a module before explicit semantic_search"
	}
	return RetrievalTaskActionSearch, "semantic intent is present, but module readiness is incomplete; use lexical search until semantic_search is explicitly configured"
}

func searchDiagnosticsSemanticIntent(value string) bool {
	normalized := strings.ToLower(value)
	return containsAny(normalized, "semantic", "similar", "meaning", "concept", "vector", "embedding", "recall", "related idea", "near match")
}

func searchModulePostures(ctx context.Context, config runclient.Config) []SearchModulePosture {
	providers := []string{runclient.SemanticModuleProviderOllama, runclient.SemanticModuleProviderGemini}
	postures := make([]SearchModulePosture, 0, len(providers))
	for _, provider := range providers {
		module, err := runclient.ReadSemanticModuleConfig(ctx, config, provider)
		posture := SearchModulePosture{
			Provider:       provider,
			CostPosture:    searchModuleCostPosture(provider),
			LatencyPosture: searchModuleLatencyPosture(provider),
		}
		if strings.TrimSpace(module.ModuleName) == "" {
			posture.Readiness = "not_installed"
			postures = append(postures, posture)
			continue
		}
		posture.ModuleName = module.ModuleName
		posture.Enabled = module.Enabled
		posture.VerificationStatus = module.VerificationStatus
		if err != nil {
			posture.Readiness = "verification_failed"
			posture.ErrorSummary = firstNString(err.Error(), 240)
		} else if !module.Enabled {
			posture.Readiness = "disabled"
		} else if module.VerificationStatus == "verified" {
			posture.Readiness = "ready"
		} else {
			posture.Readiness = firstNonEmpty(module.VerificationStatus, "unknown")
		}
		postures = append(postures, posture)
	}
	return postures
}

func searchModulePostureForProvider(postures []SearchModulePosture, provider string) *SearchModulePosture {
	for i := range postures {
		if postures[i].Provider == provider {
			return &postures[i]
		}
	}
	return nil
}

func searchModuleCostPosture(provider string) string {
	switch provider {
	case runclient.SemanticModuleProviderGemini:
		return "explicit remote provider opt-in; API billing and egress may apply"
	default:
		return "local-first Ollama provider; no external API billing from OpenClerk"
	}
}

func searchModuleLatencyPosture(provider string) string {
	switch provider {
	case runclient.SemanticModuleProviderGemini:
		return "remote embedding request plus module startup/cache overhead"
	default:
		return "local model request plus module startup/cache overhead"
	}
}

func searchModePostures() []SearchModePosture {
	return []SearchModePosture{
		{
			Mode:           "search",
			Status:         "default",
			Ranking:        "lexical_fts",
			UseWhen:        "use for citation-bearing exact, source-sensitive, or authority-oriented retrieval",
			CostPosture:    "local SQLite FTS only",
			LatencyPosture: "bounded local query; no module startup",
			Boundary:       "default ranking remains lexical",
		},
		{
			Mode:           "semantic_search",
			Status:         "explicit_module_gated",
			Ranking:        "provider_semantic_or_hybrid",
			UseWhen:        "use only when the user explicitly needs semantic/vector recall and a verified module is ready",
			CostPosture:    "provider-dependent; Ollama local, Gemini explicit remote opt-in",
			LatencyPosture: "module process, embedding provider, and cache state affect latency",
			Boundary:       "never used as hidden fallback or default ranking",
		},
	}
}

func searchTuningVisibility(options SearchDiagnosticsOptions, effectiveLimit int) []SearchTuningKnob {
	knobs := []SearchTuningKnob{
		{Knob: "limit", Value: fmt.Sprintf("%d", effectiveLimit), Visibility: "request_visible", Boundary: "bounded result count only; does not alter ranking semantics"},
		{Knob: "path_prefix", Value: firstNonEmpty(options.PathPrefix, "none"), Visibility: "request_visible", Boundary: "filters candidate documents before ranking"},
		{Knob: "metadata_filter", Value: searchDiagnosticsMetadataValue(options), Visibility: "request_visible", Boundary: "metadata/tag filters are exact filters, not authority ranking"},
		{Knob: "semantic_provider", Value: firstNonEmpty(options.Provider, runclient.SemanticModuleProviderOllama), Visibility: "request_visible_for_semantic_search", Boundary: "used only by explicit semantic_search"},
	}
	return knobs
}

func searchDiagnosticsMetadataValue(options SearchDiagnosticsOptions) string {
	if options.MetadataKey == "" {
		return "none"
	}
	return options.MetadataKey + "=" + options.MetadataValue
}
