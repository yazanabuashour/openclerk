package knowledgeplane_test

import "testing"

type trackDecision struct {
	id                 string
	name               string
	outcome            string
	publicSurface      string
	safetyPass         string
	capabilityPass     string
	uxQuality          string
	evidencePosture    string
	implementationGate string
	candidates         []candidate
}

type candidate struct {
	name       string
	safe       bool
	capable    bool
	acceptable bool
}

var roadmapTrackDecisions = []trackDecision{
	{
		id:                 "oc-uj2y.2",
		name:               "hybrid embedding and vector retrieval",
		outcome:            "defer-hybrid-keep-lexical-default",
		publicSurface:      "existing retrieval search",
		safetyPass:         "pass",
		capabilityPass:     "pass-current-primitives-hybrid-unproven",
		uxQuality:          "acceptable-current-surface",
		evidencePosture:    "reduced-eval-only",
		implementationGate: "no-product-change",
		candidates: []candidate{
			{name: "sqlite-fts-default", safe: true, capable: true, acceptable: true},
			{name: "eval-only-hybrid-fusion", safe: true, capable: true, acceptable: false},
			{name: "external-vector-store", safe: false, capable: true, acceptable: false},
		},
	},
	{
		id:                 "oc-uj2y.3",
		name:               "memory architecture and recall",
		outcome:            "defer-separate-memory-layer-keep-readonly-report",
		publicSurface:      "existing memory_router_recall_report",
		safetyPass:         "pass",
		capabilityPass:     "pass-current-primitives",
		uxQuality:          "taste-debt-watch",
		evidencePosture:    "reduced-eval-only",
		implementationGate: "no-product-change",
		candidates: []candidate{
			{name: "existing-readonly-report", safe: true, capable: true, acceptable: true},
			{name: "mem0-recall-layer", safe: true, capable: true, acceptable: false},
			{name: "autonomous-memory-writes", safe: false, capable: true, acceptable: false},
		},
	},
	{
		id:                 "oc-uj2y.4",
		name:               "structured data and non-document canonical stores",
		outcome:            "select-existing-schema-backed-record-domains",
		publicSurface:      "existing records services decisions retrieval",
		safetyPass:         "pass",
		capabilityPass:     "pass-selected-domains",
		uxQuality:          "acceptable-current-surface",
		evidencePosture:    "reduced-eval-only",
		implementationGate: "no-product-change",
		candidates: []candidate{
			{name: "existing-schema-backed-domains", safe: true, capable: true, acceptable: true},
			{name: "dynamic-schema-store", safe: false, capable: true, acceptable: false},
			{name: "docs-only-for-all-facts", safe: true, capable: false, acceptable: false},
		},
	},
	{
		id:                 "oc-uj2y.5",
		name:               "skill reduction into runner heuristics",
		outcome:            "defer-additional-shrink-keep-thin-skill",
		publicSurface:      "existing runner help and handoffs",
		safetyPass:         "pass",
		capabilityPass:     "pass-current-skill",
		uxQuality:          "watch-for-repeated-choreography",
		evidencePosture:    "reduced-eval-only",
		implementationGate: "no-product-change",
		candidates: []candidate{
			{name: "keep-current-thin-skill", safe: true, capable: true, acceptable: true},
			{name: "move-more-policy-to-help", safe: true, capable: true, acceptable: false},
			{name: "remove-no-tools-policy", safe: false, capable: true, acceptable: false},
		},
	},
	{
		id:                 "oc-uj2y.6",
		name:               "git-backed version control and lifecycle",
		outcome:            "defer-git-lifecycle-runner-surface",
		publicSurface:      "none beyond existing document and retrieval actions",
		safetyPass:         "pass",
		capabilityPass:     "pass-storage-history-reference-only",
		uxQuality:          "needs-later-targeted-evidence",
		evidencePosture:    "reduced-eval-only",
		implementationGate: "no-product-change",
		candidates: []candidate{
			{name: "storage-level-history-reference", safe: true, capable: true, acceptable: true},
			{name: "privacy-safe-status-report", safe: true, capable: true, acceptable: false},
			{name: "local-checkpoint-action", safe: true, capable: true, acceptable: false},
			{name: "restore-or-remote-push", safe: false, capable: true, acceptable: false},
		},
	},
	{
		id:                 "oc-uj2y.7",
		name:               "harness-owned web search and fetch",
		outcome:            "defer-search-planning-keep-runner-fetch",
		publicSurface:      "existing ingest_source_url planning and approved fetch",
		safetyPass:         "pass",
		capabilityPass:     "pass-current-fetch",
		uxQuality:          "search-provider-evidence-missing",
		evidencePosture:    "reduced-eval-only",
		implementationGate: "no-product-change",
		candidates: []candidate{
			{name: "current-public-url-placement", safe: true, capable: true, acceptable: true},
			{name: "read-only-search-planning", safe: true, capable: true, acceptable: false},
			{name: "browser-or-http-bypass", safe: false, capable: true, acceptable: false},
		},
	},
	{
		id:                 "oc-uj2y.8",
		name:               "artifact intake, auto-filing, tags, and fields",
		outcome:            "select-proposal-first-current-primitives",
		publicSurface:      "existing candidate validation placement and duplicate reports",
		safetyPass:         "pass",
		capabilityPass:     "pass-supported-inputs",
		uxQuality:          "parser-ocr-evidence-missing",
		evidencePosture:    "reduced-eval-only",
		implementationGate: "no-product-change",
		candidates: []candidate{
			{name: "proposal-first-current-primitives", safe: true, capable: true, acceptable: true},
			{name: "metadata-autofill-planner", safe: true, capable: true, acceptable: false},
			{name: "opaque-parser-or-ocr-claims", safe: false, capable: true, acceptable: false},
		},
	},
}

func TestRoadmapTrackDecisionsRemainEvidenceGated(t *testing.T) {
	t.Parallel()

	if len(roadmapTrackDecisions) != 7 {
		t.Fatalf("track count = %d, want 7", len(roadmapTrackDecisions))
	}
	for _, track := range roadmapTrackDecisions {
		track := track
		t.Run(track.id, func(t *testing.T) {
			t.Parallel()
			if track.safetyPass != "pass" {
				t.Fatalf("safety pass = %q, want pass", track.safetyPass)
			}
			if track.publicSurface == "" {
				t.Fatal("public surface must be explicit")
			}
			if track.implementationGate != "no-product-change" {
				t.Fatalf("implementation gate = %q, want no-product-change", track.implementationGate)
			}
			if len(track.candidates) < 2 {
				t.Fatalf("candidate count = %d, want at least 2", len(track.candidates))
			}
			if track.evidencePosture != "reduced-eval-only" {
				t.Fatalf("evidence posture = %q, want reduced-eval-only", track.evidencePosture)
			}
		})
	}
}

func TestRejectedCandidatesDoNotBypassSafetyBoundaries(t *testing.T) {
	t.Parallel()

	for _, track := range roadmapTrackDecisions {
		track := track
		t.Run(track.id, func(t *testing.T) {
			t.Parallel()
			var safeAcceptable int
			for _, candidate := range track.candidates {
				if candidate.acceptable && !candidate.safe {
					t.Fatalf("%s is acceptable without safety", candidate.name)
				}
				if candidate.acceptable && !candidate.capable {
					t.Fatalf("%s is acceptable without capability", candidate.name)
				}
				if candidate.safe && candidate.acceptable {
					safeAcceptable++
				}
			}
			if safeAcceptable == 0 {
				t.Fatalf("%s has no safe acceptable baseline candidate", track.id)
			}
		})
	}
}

func TestNoDeferredTrackPromotesDurableWriteOrStorageBehavior(t *testing.T) {
	t.Parallel()

	for _, track := range roadmapTrackDecisions {
		track := track
		t.Run(track.id, func(t *testing.T) {
			t.Parallel()
			for _, forbidden := range []string{"remote-push", "branch-switch", "destructive-restore", "autonomous-memory-writes", "opaque-parser-or-ocr-claims", "browser-or-http-bypass", "external-vector-store"} {
				if track.outcome == forbidden || track.publicSurface == forbidden {
					t.Fatalf("track promoted forbidden surface %q", forbidden)
				}
			}
		})
	}
}
