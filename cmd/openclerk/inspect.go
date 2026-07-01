package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

const inspectSchemaVersion = "openclerk-inspect.v1"

type inspectEnvelope struct {
	SchemaVersion string        `json:"schema_version"`
	Action        string        `json:"action"`
	Result        inspectResult `json:"result"`
}

type inspectResult struct {
	ReadOnly               bool                `json:"read_only"`
	WritesPerformed        int                 `json:"writes_performed"`
	Runner                 inspectRunner       `json:"runner"`
	Storage                inspectStorage      `json:"storage"`
	Vault                  inspectVault        `json:"vault"`
	Knowledge              inspectKnowledge    `json:"knowledge"`
	Modules                inspectModules      `json:"modules"`
	Git                    inspectGit          `json:"git"`
	RecommendedNextActions []inspectNextAction `json:"recommended_next_actions"`
	Warnings               []inspectIssue      `json:"warnings"`
	Blockers               []inspectIssue      `json:"blockers"`
}

type inspectRunner struct {
	Version        string   `json:"version"`
	Command        string   `json:"command"`
	PublicSurfaces []string `json:"public_surfaces"`
}

type inspectStorage struct {
	Status           string         `json:"status"`
	DatabaseBound    bool           `json:"database_bound"`
	DatabasePathKind string         `json:"database_path_kind"`
	Blockers         []inspectIssue `json:"blockers"`
}

type inspectVault struct {
	Status        string                  `json:"status"`
	VaultRootKind string                  `json:"vault_root_kind"`
	Documents     inspectDocumentCounters `json:"documents"`
	Blockers      []inspectIssue          `json:"blockers"`
}

type inspectDocumentCounters struct {
	KnownCount int `json:"known_count"`
	ChunkCount int `json:"chunk_count"`
}

type inspectKnowledge struct {
	CanonicalMarkdownAuthority bool                 `json:"canonical_markdown_authority"`
	DerivedLayers              inspectDerivedLayers `json:"derived_layers"`
	Synthesis                  inspectSynthesis     `json:"synthesis"`
	DuplicateRisk              inspectDuplicateRisk `json:"duplicate_risk"`
}

type inspectDerivedLayers struct {
	SearchIndex string `json:"search_index"`
	Graph       string `json:"graph"`
	Records     string `json:"records"`
	Decisions   string `json:"decisions"`
	Synthesis   string `json:"synthesis"`
}

type inspectSynthesis struct {
	FreshCount               int `json:"fresh_count"`
	StaleCount               int `json:"stale_count"`
	MissingSourceRefCount    int `json:"missing_source_ref_count"`
	SupersededSourceRefCount int `json:"superseded_source_ref_count"`
}

type inspectDuplicateRisk struct {
	Status         string `json:"status"`
	CandidateCount int    `json:"candidate_count"`
}

type inspectModules struct {
	Status         string                   `json:"status"`
	Installed      []inspectInstalledModule `json:"installed"`
	SemanticSearch inspectSemanticSearch    `json:"semantic_search"`
	OCRReview      inspectOCRReview         `json:"ocr_review"`
}

type inspectInstalledModule struct {
	Kind               string `json:"kind"`
	Provider           string `json:"provider"`
	ModuleName         string `json:"module_name"`
	Enabled            bool   `json:"enabled"`
	VerificationStatus string `json:"verification_status"`
}

type inspectSemanticSearch struct {
	Available             bool `json:"available"`
	DefaultRankingChanged bool `json:"default_ranking_changed"`
}

type inspectOCRReview struct {
	Available bool `json:"available"`
}

type inspectGit struct {
	Status                string         `json:"status"`
	GitCheckpointsEnabled bool           `json:"git_checkpoints_enabled"`
	Blockers              []inspectIssue `json:"blockers"`
}

type inspectNextAction struct {
	Label   string         `json:"label"`
	Kind    string         `json:"kind"`
	Command string         `json:"command,omitempty"`
	Surface string         `json:"surface,omitempty"`
	Request map[string]any `json:"request,omitempty"`
	When    string         `json:"when,omitempty"`
}

type inspectIssue struct {
	Code    string `json:"code"`
	Area    string `json:"area"`
	Message string `json:"message"`
}

func buildInspectEnvelope(ctx context.Context, config runclient.Config, version string) inspectEnvelope {
	result := inspectResult{
		ReadOnly:        true,
		WritesPerformed: 0,
		Runner: inspectRunner{
			Version:        version,
			Command:        "openclerk",
			PublicSurfaces: []string{"inspect", "config", "module", "document", "retrieval", "clerk", "capabilities", "demo"},
		},
		Storage: inspectStorage{
			Status:           "unknown",
			DatabaseBound:    false,
			DatabasePathKind: "unknown",
			Blockers:         []inspectIssue{},
		},
		Vault: inspectVault{
			Status:        "unknown",
			VaultRootKind: "unknown",
			Documents:     inspectDocumentCounters{},
			Blockers:      []inspectIssue{},
		},
		Knowledge: inspectKnowledge{
			CanonicalMarkdownAuthority: true,
			DerivedLayers: inspectDerivedLayers{
				SearchIndex: "unknown",
				Graph:       "unknown",
				Records:     "unknown",
				Decisions:   "unknown",
				Synthesis:   "unknown",
			},
			DuplicateRisk: inspectDuplicateRisk{Status: "not_checked"},
		},
		Modules: inspectModules{
			Status:         "unknown",
			Installed:      []inspectInstalledModule{},
			SemanticSearch: inspectSemanticSearch{Available: false, DefaultRankingChanged: false},
			OCRReview:      inspectOCRReview{Available: false},
		},
		Git: inspectGit{
			Status:                "unknown",
			GitCheckpointsEnabled: runner.GitLifecycleCheckpointsEnabled(config),
			Blockers:              []inspectIssue{},
		},
		Warnings: []inspectIssue{},
		Blockers: []inspectIssue{},
	}

	inspection, err := runclient.InspectExistingRuntime(ctx, config)
	if err != nil {
		result.Storage.Status = "error"
		addInspectBlocker(&result, &result.Storage.Blockers, "storage_error", "storage", err.Error())
		result.RecommendedNextActions = inspectRecommendedNextActions(result)
		return inspectEnvelope{SchemaVersion: inspectSchemaVersion, Action: "inspect", Result: result}
	}

	result.Storage.DatabasePathKind = inspectDatabasePathKind(inspection.DatabaseSource)
	result.Storage.DatabaseBound = inspection.DatabaseExists
	result.Vault.Documents = inspectDocumentCounters{KnownCount: inspection.DocumentCount, ChunkCount: inspection.ChunkCount}

	switch {
	case !inspection.DatabaseExists:
		result.Storage.Status = "missing"
		addInspectBlocker(&result, &result.Storage.Blockers, "storage_missing", "storage", "OpenClerk storage does not exist; inspect will not initialize it")
		result.Vault.Status = "unbound"
		result.Vault.VaultRootKind = "unconfigured"
	case !inspection.DatabaseInitialized:
		result.Storage.Status = "uninitialized"
		addInspectBlocker(&result, &result.Storage.Blockers, "storage_uninitialized", "storage", "OpenClerk storage exists without a configured vault root; inspect will not bind it")
		result.Vault.Status = "unbound"
		result.Vault.VaultRootKind = "unconfigured"
	default:
		result.Storage.Status = "ready"
		result.Vault.VaultRootKind = "configured"
		result.Vault.Status = inspectVaultStatus(inspection.VaultRoot, &result)
	}

	result.Knowledge = inspectKnowledgePosture(inspection, result.Storage.Status == "ready" && result.Vault.Status == "ready")
	result.Modules = inspectModulePosture(inspection, result.Storage.Status == "ready")
	result.Git = inspectGitPosture(ctx, inspection.VaultRoot, config, result.Vault.Status)
	result.Warnings = inspectWarnings(result)
	result.RecommendedNextActions = inspectRecommendedNextActions(result)

	return inspectEnvelope{SchemaVersion: inspectSchemaVersion, Action: "inspect", Result: result}
}

func inspectDatabasePathKind(source string) string {
	switch source {
	case "default":
		return "xdg_default"
	case "env":
		return "env_override"
	case "flag":
		return "flag_override"
	default:
		return "unknown"
	}
}

func inspectVaultStatus(vaultRoot string, result *inspectResult) string {
	info, err := os.Stat(vaultRoot)
	if err == nil {
		if !info.IsDir() {
			addInspectBlocker(result, &result.Vault.Blockers, "vault_not_directory", "vault", "configured vault root is not a directory")
			return "error"
		}
		if _, err := os.ReadDir(vaultRoot); err != nil {
			addInspectBlocker(result, &result.Vault.Blockers, "vault_inaccessible", "vault", "configured vault root is not readable")
			return "error"
		}
		return "ready"
	}
	if os.IsNotExist(err) {
		addInspectBlocker(result, &result.Vault.Blockers, "vault_missing", "vault", "configured vault root is missing; inspect will not create it")
		return "missing"
	}
	addInspectBlocker(result, &result.Vault.Blockers, "vault_error", "vault", "configured vault root could not be inspected")
	return "error"
}

func addInspectBlocker(result *inspectResult, scoped *[]inspectIssue, code string, area string, message string) {
	issue := inspectIssue{Code: code, Area: area, Message: message}
	*scoped = append(*scoped, issue)
	result.Blockers = append(result.Blockers, issue)
}

func inspectKnowledgePosture(inspection runclient.RuntimeInspection, ready bool) inspectKnowledge {
	posture := inspectKnowledge{
		CanonicalMarkdownAuthority: true,
		DerivedLayers: inspectDerivedLayers{
			SearchIndex: "unknown",
			Graph:       "unknown",
			Records:     "unknown",
			Decisions:   "unknown",
			Synthesis:   "unknown",
		},
		Synthesis: inspectSynthesis{
			FreshCount:               inspection.Synthesis.FreshCount,
			StaleCount:               inspection.Synthesis.StaleCount,
			MissingSourceRefCount:    inspection.Synthesis.MissingSourceRefCount,
			SupersededSourceRefCount: inspection.Synthesis.SupersededSourceRefCount,
		},
		DuplicateRisk: inspectDuplicateRisk{Status: "not_checked", CandidateCount: 0},
	}
	if !ready {
		posture.DerivedLayers = inspectDerivedLayers{
			SearchIndex: "unavailable",
			Graph:       "unavailable",
			Records:     "unavailable",
			Decisions:   "unavailable",
			Synthesis:   "unavailable",
		}
		return posture
	}
	if inspection.Tables["chunks"] {
		posture.DerivedLayers.SearchIndex = "ready"
		if inspection.DocumentCount > 0 && inspection.ChunkCount == 0 {
			posture.DerivedLayers.SearchIndex = "stale"
		}
	} else if inspection.DocumentCount == 0 {
		posture.DerivedLayers.SearchIndex = "ready"
	} else {
		posture.DerivedLayers.SearchIndex = "unknown"
	}
	posture.DerivedLayers.Graph = inspectProjectionLayerStatus(inspection, "graph")
	posture.DerivedLayers.Records = inspectProjectionLayerStatus(inspection, "records")
	posture.DerivedLayers.Decisions = inspectProjectionLayerStatus(inspection, "decisions")
	posture.DerivedLayers.Synthesis = inspectProjectionLayerStatus(inspection, "synthesis")
	if inspection.Tables["projection_states"] {
		posture.DuplicateRisk.Status = "checked"
		posture.DuplicateRisk.CandidateCount = inspection.DuplicateProjectionCount
	}
	return posture
}

func inspectProjectionLayerStatus(inspection runclient.RuntimeInspection, projection string) string {
	if !inspection.Tables["projection_states"] {
		if inspection.DocumentCount == 0 {
			return "ready"
		}
		return "unknown"
	}
	state := inspection.Projections[projection]
	if state.Stale > 0 || state.Unknown > 0 {
		return "stale"
	}
	return "ready"
}

func inspectModulePosture(inspection runclient.RuntimeInspection, storageReady bool) inspectModules {
	modules := inspectModules{
		Status:         "unknown",
		Installed:      []inspectInstalledModule{},
		SemanticSearch: inspectSemanticSearch{Available: false, DefaultRankingChanged: false},
		OCRReview:      inspectOCRReview{Available: false},
	}
	if !storageReady {
		return modules
	}
	if len(inspection.Modules) == 0 {
		modules.Status = "none_installed"
		return modules
	}
	modules.Status = "ready"
	for _, module := range inspection.Modules {
		modules.Installed = append(modules.Installed, inspectInstalledModule{
			Kind:               module.Kind,
			Provider:           module.Provider,
			ModuleName:         module.ModuleName,
			Enabled:            module.Enabled,
			VerificationStatus: module.VerificationStatus,
		})
		if module.Enabled && module.VerificationStatus != "verified" {
			modules.Status = "partial"
		}
		if module.Kind == runclient.ModuleKindEmbeddingProvider && module.Enabled && module.VerificationStatus == "verified" {
			modules.SemanticSearch.Available = true
		}
		if module.Kind == runclient.ModuleKindOCRProvider && module.Enabled && module.VerificationStatus == "verified" {
			modules.OCRReview.Available = true
		}
	}
	return modules
}

func inspectGitPosture(ctx context.Context, vaultRoot string, config runclient.Config, vaultStatus string) inspectGit {
	posture := inspectGit{
		Status:                "unknown",
		GitCheckpointsEnabled: runner.GitLifecycleCheckpointsEnabled(config),
		Blockers:              []inspectIssue{},
	}
	if vaultStatus != "ready" {
		return posture
	}
	report, err := runner.InspectGitLifecycleStatus(ctx, vaultRoot, config)
	if err != nil {
		posture.Status = "unknown"
		posture.Blockers = append(posture.Blockers, inspectIssue{Code: "git_error", Area: "git", Message: err.Error()})
		return posture
	}
	switch report.GitStatus {
	case "available":
		posture.Status = "clean"
	case "dirty":
		posture.Status = "dirty"
	case "unavailable":
		posture.Status = "not_git"
	default:
		posture.Status = "unknown"
	}
	return posture
}

func inspectWarnings(result inspectResult) []inspectIssue {
	warnings := []inspectIssue{}
	if result.Storage.Status == "ready" && result.Vault.Status == "ready" {
		for layer, status := range map[string]string{
			"search_index": result.Knowledge.DerivedLayers.SearchIndex,
			"graph":        result.Knowledge.DerivedLayers.Graph,
			"records":      result.Knowledge.DerivedLayers.Records,
			"decisions":    result.Knowledge.DerivedLayers.Decisions,
			"synthesis":    result.Knowledge.DerivedLayers.Synthesis,
		} {
			if status == "unknown" {
				warnings = append(warnings, inspectIssue{Code: "knowledge_unknown", Area: "knowledge", Message: layer + " posture could not be computed"})
			}
		}
		if result.Git.Status == "unknown" {
			warnings = append(warnings, inspectIssue{Code: "git_unknown", Area: "git", Message: "git posture could not be computed without mutating git state"})
		}
	}
	return warnings
}

func inspectRecommendedNextActions(result inspectResult) []inspectNextAction {
	actions := []inspectNextAction{}
	if result.Storage.Status != "ready" || result.Vault.Status == "unbound" || result.Vault.Status == "missing" || result.Vault.Status == "error" {
		actions = append(actions, inspectNextAction{
			Label:   "Bind a vault",
			Kind:    "shell",
			Command: "openclerk init --vault-root path/to/vault",
			When:    "storage.status is not ready or vault.status is not ready",
		})
	}
	if result.Storage.Status == "ready" && result.Vault.Status == "ready" {
		actions = append(actions, inspectNextAction{
			Label:   "Get task context",
			Kind:    "runner_request",
			Surface: "clerk",
			Command: "openclerk clerk context_pack",
			Request: map[string]any{"task": "describe the task here", "limit": 5},
		})
	}
	if result.Knowledge.Synthesis.StaleCount > 0 || result.Knowledge.Synthesis.MissingSourceRefCount > 0 || result.Knowledge.Synthesis.SupersededSourceRefCount > 0 {
		actions = append(actions, inspectNextAction{
			Label:   "Review stale synthesis",
			Kind:    "runner_request",
			Surface: "retrieval",
			Command: "openclerk retrieval",
			Request: map[string]any{"action": "maintenance_report"},
			When:    "knowledge.synthesis has stale, missing, or superseded source refs",
		})
	}
	return actions
}

func inspectUsage(w interface{ Write([]byte) (int, error) }) {
	_, _ = fmt.Fprintln(w, "usage: openclerk inspect [--db path]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Writes one openclerk-inspect.v1 JSON posture report for agents before work starts.")
	_, _ = fmt.Fprintln(w, "Read-only: no init, sync, repair, refresh, source fetch, ingest, document write, SQLite mutation, module config mutation, or git mutation.")
	_, _ = fmt.Fprintln(w, "Reports storage, vault, derived knowledge layers, optional modules, git posture, blockers, and next safe runner requests.")
	_, _ = fmt.Fprintln(w, "If storage or vault binding is missing, inspect returns blockers and an init recommendation instead of creating storage.")
}
