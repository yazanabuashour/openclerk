package runner

import "strings"

const (
	workflowGuideValidationBoundaries = "read-only routing report; uses runner-owned heuristic guidance only; does not inspect documents, query storage, fetch URLs, create candidates, create or update documents, repair synthesis, or bypass installed openclerk document/retrieval JSON"
	workflowGuideAuthorityLimits      = "this report routes workflow surface choice only; final answers and source-sensitive claims must come from the selected runner action result, citations, provenance, projection freshness, validation boundaries, and authority limits"
)

func runWorkflowGuideReport(options WorkflowGuideOptions) WorkflowGuideReport {
	intent := strings.TrimSpace(options.Intent)
	selected := selectWorkflowGuideCandidate(intent)
	report := WorkflowGuideReport{
		Intent:               intent,
		RecommendedSurface:   selected.Surface,
		RunnerDomain:         selected.RunnerDomain,
		RequestShape:         selected.RequestShape,
		UseWhen:              selected.UseWhen,
		DoNotUseFor:          selected.DoNotUseFor,
		CandidateSurfaces:    workflowGuideCandidates(),
		ValidationBoundaries: workflowGuideValidationBoundaries,
		AuthorityLimits:      workflowGuideAuthorityLimits,
	}
	report.AgentHandoff = &AgentHandoff{
		AnswerSummary:               selected.HandoffSummary,
		Evidence:                    []string{"intent:" + intent, "recommended_surface:" + selected.Surface, "runner_domain:" + selected.RunnerDomain},
		ValidationBoundaries:        report.ValidationBoundaries,
		AuthorityLimits:             report.AuthorityLimits,
		FollowUpPrimitiveInspection: "Run the recommended openclerk document or retrieval request next; use lower-level primitives only for explicit advanced/manual work or after a runner rejection.",
	}
	return report
}

type workflowGuideSelection struct {
	Surface        string
	RunnerDomain   string
	RequestShape   string
	UseWhen        string
	DoNotUseFor    []string
	HandoffSummary string
}

func selectWorkflowGuideCandidate(intent string) workflowGuideSelection {
	normalized := strings.ToLower(intent)
	switch {
	case containsAny(normalized, "duplicate", "already exists", "update existing", "update versus new", "same note"):
		return workflowGuideSelection{
			Surface:        "duplicate_candidate_report",
			RunnerDomain:   "retrieval",
			RequestShape:   `{"action":"duplicate_candidate_report","duplicate_candidate":{"query":"...","path_prefix":"notes/","limit":10}}`,
			UseWhen:        "use before choosing update versus new for plausible duplicate document requests",
			DoNotUseFor:    []string{"durable document writes", "unrelated source-sensitive audits"},
			HandoffSummary: "Use retrieval duplicate_candidate_report, then answer from duplicate_candidate.agent_handoff before asking for durable-write approval.",
		}
	case containsAny(normalized, "public url", "public link", "source url", "source placement", "fetch url", "ingest url", "web page", "pdf"):
		return workflowGuideSelection{
			Surface:        "ingest_source_url plan",
			RunnerDomain:   "document",
			RequestShape:   `{"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"plan","source_type":"web","title":"Optional title"}}`,
			UseWhen:        "use for public URL source placement before durable fetch or write approval",
			DoNotUseFor:    []string{"login-gated pages", "purchases", "captcha/paywall bypasses", "non-runner HTTP/browser fetch"},
			HandoffSummary: "Use document ingest_source_url with mode plan for public-link placement; public read/fetch permission is separate from durable-write approval.",
		}
	case containsAny(normalized, "synthesis", "compile", "summarize sources", "source-linked answer", "file answer"):
		return workflowGuideSelection{
			Surface:        "compile_synthesis",
			RunnerDomain:   "document",
			RequestShape:   `{"action":"compile_synthesis","synthesis":{"path":"synthesis/example.md","title":"Example","source_refs":["sources/a.md"],"body_facts":["..."],"mode":"create_or_update"}}`,
			UseWhen:        "use for approved source-linked synthesis create/update",
			DoNotUseFor:    []string{"uncited claims", "duplicate synthesis creation", "unapproved durable writes"},
			HandoffSummary: "Use document compile_synthesis for approved source-linked synthesis, then answer from compile_synthesis.agent_handoff.",
		}
	case containsAny(normalized, "source audit", "conflict", "contradiction", "stale source", "repair synthesis"):
		return workflowGuideSelection{
			Surface:        "source_audit_report",
			RunnerDomain:   "retrieval",
			RequestShape:   `{"action":"source_audit_report","source_audit":{"query":"...","target_path":"synthesis/example.md","mode":"explain","limit":10}}`,
			UseWhen:        "use for source-sensitive audit explanation or approved existing-target repair",
			DoNotUseFor:    []string{"broad contradiction engine", "new synthesis creation"},
			HandoffSummary: "Use retrieval source_audit_report; explain mode is read-only, repair_existing may update only an existing synthesis target.",
		}
	case containsAny(normalized, "memory", "router", "recall"):
		return workflowGuideSelection{
			Surface:        "memory_router_recall_report",
			RunnerDomain:   "retrieval",
			RequestShape:   `{"action":"memory_router_recall_report","memory_router_recall":{"query":"...","limit":10}}`,
			UseWhen:        "use for routine read-only memory/router recall evidence",
			DoNotUseFor:    []string{"remember/recall transports", "autonomous router APIs", "memory writes"},
			HandoffSummary: "Use retrieval memory_router_recall_report and answer from returned memory_router_recall evidence.",
		}
	case containsAny(normalized, "structured", "canonical store", "time-series", "time series", "metrics", "measurements", "inventory", "finance", "health"):
		return workflowGuideSelection{
			Surface:        "structured_store_report",
			RunnerDomain:   "retrieval",
			RequestShape:   `{"action":"structured_store_report","structured_store":{"domain":"records","query":"...","limit":10}}`,
			UseWhen:        "use for structured-data and non-document canonical-store decision support",
			DoNotUseFor:    []string{"independent canonical tables", "external store connectors", "durable structured writes"},
			HandoffSummary: "Use retrieval structured_store_report and answer from structured_store.agent_handoff.",
		}
	case containsAny(normalized, "hybrid", "vector", "embedding", "semantic retrieval"):
		return workflowGuideSelection{
			Surface:        "hybrid_retrieval_report",
			RunnerDomain:   "retrieval",
			RequestShape:   `{"action":"hybrid_retrieval_report","hybrid_retrieval":{"query":"...","path_prefix":"docs/","limit":10}}`,
			UseWhen:        "use for read-only hybrid/vector retrieval decision support",
			DoNotUseFor:    []string{"vector-ranked answers", "embedding-store evidence", "default ranking changes"},
			HandoffSummary: "Use retrieval hybrid_retrieval_report and do not claim vector-ranked retrieval from it.",
		}
	case containsAny(normalized, "record", "records", "decision", "provenance", "projection", "evidence bundle", "freshness"):
		return workflowGuideSelection{
			Surface:        "evidence_bundle_report",
			RunnerDomain:   "retrieval",
			RequestShape:   `{"action":"evidence_bundle_report","evidence_bundle":{"query":"...","projection":"records","limit":10}}`,
			UseWhen:        "use for records, decisions, provenance, projection freshness, and citation evidence bundles",
			DoNotUseFor:    []string{"durable writes", "memory transport", "hidden ranking"},
			HandoffSummary: "Use retrieval evidence_bundle_report, then answer from evidence_bundle.agent_handoff.",
		}
	default:
		return workflowGuideSelection{
			Surface:        "current_primitives_or_runner_help",
			RunnerDomain:   "document/retrieval",
			RequestShape:   `{"action":"search","search":{"text":"...","limit":10}}`,
			UseWhen:        "use openclerk document --help or openclerk retrieval --help, then choose primitives for explicit manual or unsupported promoted workflow requests",
			DoNotUseFor:    []string{"bypass workflows", "durable writes without approval"},
			HandoffSummary: "Use compact runner help and current primitives; do not expand SKILL.md with a long recipe for this intent.",
		}
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func workflowGuideCandidates() []WorkflowGuideCandidate {
	return []WorkflowGuideCandidate{
		{
			Surface:        "current_primitives_or_runner_help",
			Status:         "keep_for_manual_or_advanced",
			SelectionRule:  "use when no promoted action matches, after runner rejection, or when the user explicitly asks for primitives",
			Boundary:       "do not repair routine UX by adding long SKILL.md recipes",
			RequestExample: `{"action":"search","search":{"text":"authority evidence","limit":10}}`,
		},
		{
			Surface:        "existing_natural_runner_action",
			Status:         "prefer_when_input_belongs_to_existing_action",
			SelectionRule:  "use when an adjacent mode on an existing action preserves the natural workflow surface",
			Boundary:       "public read/fetch/inspect permission is separate from durable-write approval",
			RequestExample: `{"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"plan","source_type":"web"}}`,
		},
		{
			Surface:        "promoted_workflow_action_with_agent_handoff",
			Status:         "prefer_for_repeated_routine_workflows",
			SelectionRule:  "use when current primitives are safe but too ceremonial, scripted, slow, or guidance-dependent",
			Boundary:       "must preserve citations, provenance, freshness, duplicate handling, local-first runner-only access, and approval-before-write",
			RequestExample: `{"action":"duplicate_candidate_report","duplicate_candidate":{"query":"...","limit":10}}`,
		},
		{
			Surface:       "skill_recipe_expansion",
			Status:        "avoid",
			SelectionRule: "do not choose unless it is a compact safety bridge or action index entry for an already-promoted surface",
			Boundary:      "long request examples, field catalogs, and exact command choreography belong in runner help, reports, docs, or evals",
		},
	}
}
