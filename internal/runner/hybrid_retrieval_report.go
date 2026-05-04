package runner

import (
	"context"
	"fmt"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const (
	hybridRetrievalValidationBoundaries = "read-only report; uses installed OpenClerk retrieval JSON and current lexical search only; does not create embeddings, vectors, vector stores, external API calls, generated corpora, direct SQLite reads, direct vault inspection, HTTP/MCP bypasses, source-built runners, or default ranking changes"
	hybridRetrievalAuthorityLimits      = "canonical markdown and promoted records remain authoritative; lexical hits are citation-bearing evidence; this report is decision support for retrieval infrastructure and not a source of truth"
)

func runHybridRetrievalReport(ctx context.Context, client *runclient.Client, options HybridRetrievalOptions) (HybridRetrievalReport, error) {
	limit := options.Limit
	if limit == 0 {
		limit = 10
	}
	if limit < 1 || limit > 100 {
		return HybridRetrievalReport{}, domain.ValidationError("hybrid_retrieval.limit must be between 1 and 100", map[string]any{"limit": limit})
	}

	search, err := client.Search(ctx, domain.SearchQuery{
		Text:       options.Query,
		PathPrefix: options.PathPrefix,
		Limit:      limit,
	})
	if err != nil {
		return HybridRetrievalReport{}, err
	}
	convertedSearch := toSearchResult(search)
	evidenceInspected := []string{fmt.Sprintf("search:%s", options.Query)}
	if options.PathPrefix != "" {
		evidenceInspected = append(evidenceInspected, "path_prefix:"+options.PathPrefix)
	}
	evidenceInspected = append(evidenceInspected, fmt.Sprintf("lexical_hits:%d", len(convertedSearch.Hits)))

	report := HybridRetrievalReport{
		Query:                options.Query,
		PathPrefix:           options.PathPrefix,
		LexicalSearch:        &convertedSearch,
		CandidateSurfaces:    hybridRetrievalCandidates(),
		Recommendation:       "keep lexical search as the default retrieval path; use this read-only report to package citation-bearing baseline evidence before any future embedding/vector POC, and promote durable hybrid ranking only after scale and citation-quality evidence justify it",
		SafetyPass:           "passes: report is read-only, local-first, runner-only, and preserves citations, provenance/freshness boundaries, and approval-before-write",
		CapabilityPass:       "partial pass for decision support: current FTS evidence is packaged with candidate-surface guidance; no vector recall or embedding store is claimed",
		UXQuality:            "improves deferred-capability review ergonomics by reducing repeated baseline-search plus policy-summary choreography to one retrieval action",
		PerformancePosture:   "bounded by one current lexical search with the requested limit; no corpus-wide vector scan, import job, remote embedding call, or index rebuild is performed",
		EvidencePosture:      "baseline evidence comes from citation-bearing OpenClerk search results; durable hybrid promotion still requires targeted POC/eval evidence on recall, citation correctness, freshness, and 100 MB/1 GB scale cost",
		ValidationBoundaries: hybridRetrievalValidationBoundaries,
		AuthorityLimits:      hybridRetrievalAuthorityLimits,
		EvidenceInspected:    evidenceInspected,
	}
	report.AgentHandoff = &AgentHandoff{
		AnswerSummary:               report.Recommendation,
		Evidence:                    evidenceInspected,
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "Use retrieval search, provenance_events, and projection_states directly only for drill-down after this report; do not infer vector evidence from this report.",
	}
	return report, nil
}

func hybridRetrievalCandidates() []HybridRetrievalCandidate {
	return []HybridRetrievalCandidate{
		{
			Surface:    "current_lexical_default",
			Status:     "keep",
			Safety:     "passes; current FTS preserves local-first operation and citation-bearing chunks",
			Capability: "passes for exact/source-sensitive lookup and existing scale evidence, but does not prove semantic recall gains",
			UXQuality:  "acceptable for routine source-grounded retrieval; weak as repeated deferred-capability evidence because agents must restate policy boundaries",
			Implementation: []string{
				"no default ranking change",
				"no schema change",
			},
		},
		{
			Surface:    "durable_embedding_vector_index",
			Status:     "not_promoted",
			Safety:     "unproven until embedding provenance, refresh, stale-index, and local/offline failure behavior are specified",
			Capability: "candidate for future recall gains; requires deterministic evals before product behavior changes",
			UXQuality:  "could simplify semantic retrieval after evidence, but would add operational ceremony if exposed too early",
			Implementation: []string{
				"requires storage/index design",
				"requires scale and freshness gates",
				"requires citation regression checks",
			},
		},
		{
			Surface:    "external_or_hosted_vector_store",
			Status:     "not_promoted",
			Safety:     "does not fit routine local-first OpenClerk boundaries without a stronger authority and privacy model",
			Capability: "may be useful as comparison evidence, not as the default OpenClerk product surface",
			UXQuality:  "adds provider, sync, and approval complexity for normal local users",
			Implementation: []string{
				"treat as benchmark/reference only",
				"do not bypass installed runner",
			},
		},
		{
			Surface:    "hybrid_retrieval_report",
			Status:     "promoted_read_only",
			Safety:     "passes because it packages existing runner evidence and declares what it does not prove",
			Capability: "passes for baseline evidence packaging and candidate comparison; intentionally does not claim vector-ranked retrieval",
			UXQuality:  "improves the decision workflow with one natural retrieval action and agent_handoff",
			Implementation: []string{
				"runner JSON action under openclerk retrieval",
				"help and skill action index",
				"unit tests and reduced eval report",
			},
		},
	}
}
