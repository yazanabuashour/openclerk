package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

const semanticRetrievalContractVersion = "openclerk_semantic_retrieval.v1"

type semanticModuleSearchRequest struct {
	Query                     string `json:"query"`
	PathPrefix                string `json:"path_prefix,omitempty"`
	Tag                       string `json:"tag,omitempty"`
	MetadataKey               string `json:"metadata_key,omitempty"`
	MetadataValue             string `json:"metadata_value,omitempty"`
	Limit                     int    `json:"limit,omitempty"`
	Provider                  string `json:"provider,omitempty"`
	OllamaURL                 string `json:"ollama_url,omitempty"`
	EmbeddingModel            string `json:"embedding_model,omitempty"`
	GeminiAPIBase             string `json:"gemini_api_base,omitempty"`
	GeminiConfigKey           string `json:"gemini_config_key,omitempty"`
	EmbeddingOutputDimensions int    `json:"embedding_output_dimensions,omitempty"`
	CacheDir                  string `json:"cache_dir,omitempty"`
}

type semanticModuleSearchResponse struct {
	SchemaVersion        string                 `json:"schema_version"`
	Query                string                 `json:"query"`
	PathPrefix           string                 `json:"path_prefix,omitempty"`
	Tag                  string                 `json:"tag,omitempty"`
	MetadataKey          string                 `json:"metadata_key,omitempty"`
	MetadataValue        string                 `json:"metadata_value,omitempty"`
	Provider             SemanticProviderStatus `json:"provider"`
	Cache                SemanticCacheStatus    `json:"cache"`
	Results              []SearchHit            `json:"results,omitempty"`
	Hits                 []SearchHit            `json:"hits,omitempty"`
	HitCount             int                    `json:"hit_count"`
	DuplicateChunks      int                    `json:"duplicate_chunks"`
	Ranking              string                 `json:"ranking"`
	SearchStatus         string                 `json:"search_status"`
	PrivacyDisclosure    string                 `json:"privacy_disclosure"`
	ValidationBoundaries string                 `json:"validation_boundaries"`
	AuthorityLimits      string                 `json:"authority_limits"`
	AgentHandoff         struct {
		Summary                       string   `json:"summary"`
		EvidenceInspected             []string `json:"evidence_inspected,omitempty"`
		FollowUpPrimitiveInspection   string   `json:"follow_up_primitive_inspection"`
		ApprovalOrConfigurationNeeded string   `json:"approval_or_configuration_needed,omitempty"`
	} `json:"agent_handoff"`
}

func runSemanticSearch(ctx context.Context, client *runclient.Client, options SemanticSearchOptions) (SemanticSearchResult, error) {
	options = normalizeSemanticModuleSearchOptions(options)
	paths := client.Paths()
	moduleConfig, err := runclient.ReadSemanticModuleConfig(ctx, runclient.Config{DatabasePath: paths.DatabasePath}, options.Provider)
	if err != nil {
		return semanticModuleBlockedResult(options, "module_unverified", err), nil
	}
	if strings.TrimSpace(moduleConfig.ModuleName) == "" {
		return semanticModuleBlockedResult(options, "module_not_installed", errors.New("semantic embedding module is not installed")), nil
	}
	if !moduleConfig.Enabled {
		return semanticModuleBlockedResult(options, "module_disabled", errors.New("semantic embedding module is disabled")), nil
	}
	request := semanticModuleRequestFromOptions(options, moduleConfig.ProviderConfig)
	if request.Provider == runclient.SemanticModuleProviderOllama {
		if rejection := validateSemanticSearchOllamaURL(request.OllamaURL); rejection != "" {
			return semanticModuleBlockedResult(options, "module_config_invalid", errors.New(rejection)), nil
		}
	}
	response, err := executeSemanticModule(ctx, moduleConfig, paths.DatabasePath, request)
	if err != nil {
		return semanticModuleBlockedResult(options, "module_blocked", err), nil
	}
	result, err := semanticModuleResponseToResult(response, moduleConfig.ModuleName)
	if err != nil {
		return semanticModuleBlockedResult(options, "module_contract_invalid", err), nil
	}
	return result, nil
}

func normalizeSemanticModuleSearchOptions(options SemanticSearchOptions) SemanticSearchOptions {
	options.Query = strings.TrimSpace(options.Query)
	options.PathPrefix = strings.TrimSpace(options.PathPrefix)
	options.MetadataKey = strings.TrimSpace(options.MetadataKey)
	options.MetadataValue = strings.TrimSpace(options.MetadataValue)
	options.Tag = strings.TrimSpace(options.Tag)
	options.Provider = strings.ToLower(strings.TrimSpace(options.Provider))
	if options.Provider == "" {
		options.Provider = runclient.SemanticModuleProviderOllama
	}
	options.OllamaURL = strings.TrimRight(strings.TrimSpace(options.OllamaURL), "/")
	options.EmbeddingModel = strings.TrimSpace(options.EmbeddingModel)
	options.GeminiAPIBase = strings.TrimRight(strings.TrimSpace(options.GeminiAPIBase), "/")
	options.CacheDir = strings.TrimSpace(options.CacheDir)
	options.Limit = defaultRunnerLimit(options.Limit, 10)
	return options
}

func semanticModuleRequestFromOptions(options SemanticSearchOptions, config map[string]string) semanticModuleSearchRequest {
	request := semanticModuleSearchRequest{
		Query:                     options.Query,
		PathPrefix:                options.PathPrefix,
		Tag:                       options.Tag,
		MetadataKey:               options.MetadataKey,
		MetadataValue:             options.MetadataValue,
		Limit:                     options.Limit,
		Provider:                  options.Provider,
		OllamaURL:                 firstNonEmpty(options.OllamaURL, config["ollama_url"]),
		EmbeddingModel:            firstNonEmpty(options.EmbeddingModel, config["embedding_model"]),
		GeminiAPIBase:             firstNonEmpty(options.GeminiAPIBase, config["gemini_api_base"]),
		GeminiConfigKey:           "GEMINI_API_KEY",
		EmbeddingOutputDimensions: options.EmbeddingOutputDimensions,
		CacheDir:                  options.CacheDir,
	}
	if request.EmbeddingOutputDimensions == 0 {
		if value := strings.TrimSpace(config["embedding_output_dimensions"]); value != "" {
			var parsed int
			if _, err := fmt.Sscanf(value, "%d", &parsed); err == nil && parsed > 0 {
				request.EmbeddingOutputDimensions = parsed
			}
		}
	}
	return request
}

func executeSemanticModule(ctx context.Context, moduleConfig runclient.SemanticModuleConfig, databasePath string, request semanticModuleSearchRequest) (semanticModuleSearchResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return semanticModuleSearchResponse{}, err
	}
	args := append([]string{}, moduleConfig.CommandArgs...)
	args = append(args, "search", "--db", databasePath)
	command := exec.CommandContext(ctx, moduleConfig.Command, args...)
	command.Stdin = bytes.NewReader(payload)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return semanticModuleSearchResponse{}, fmt.Errorf("run semantic module %s: %w: %s", moduleConfig.ModuleName, err, firstNString(strings.TrimSpace(stderr.String()), 240))
	}
	decoder := json.NewDecoder(io.LimitReader(&stdout, 4<<20))
	var response semanticModuleSearchResponse
	if err := decoder.Decode(&response); err != nil {
		return semanticModuleSearchResponse{}, fmt.Errorf("decode semantic module response: %w", err)
	}
	return response, nil
}

func semanticModuleResponseToResult(response semanticModuleSearchResponse, moduleName string) (SemanticSearchResult, error) {
	if response.SchemaVersion != semanticRetrievalContractVersion && response.SchemaVersion != "semantic_retrieval_adapter.v1" {
		return SemanticSearchResult{}, fmt.Errorf("unsupported semantic module schema_version %q", response.SchemaVersion)
	}
	hits := response.Hits
	if len(hits) == 0 {
		hits = response.Results
	}
	for _, hit := range hits {
		if len(hit.Citations) == 0 {
			return SemanticSearchResult{}, fmt.Errorf("semantic module %s returned hit %s without citations", moduleName, hit.DocID)
		}
		for _, citation := range hit.Citations {
			if strings.TrimSpace(citation.DocID) == "" || strings.TrimSpace(citation.Path) == "" {
				return SemanticSearchResult{}, fmt.Errorf("semantic module %s returned incomplete citation", moduleName)
			}
		}
	}
	result := SemanticSearchResult{
		Query:                response.Query,
		PathPrefix:           response.PathPrefix,
		Tag:                  response.Tag,
		MetadataKey:          response.MetadataKey,
		MetadataValue:        response.MetadataValue,
		Provider:             response.Provider,
		Cache:                response.Cache,
		Ranking:              response.Ranking,
		SearchStatus:         response.SearchStatus,
		DuplicateChunks:      response.DuplicateChunks,
		Hits:                 hits,
		PrivacyDisclosure:    response.PrivacyDisclosure,
		ValidationBoundaries: response.ValidationBoundaries,
		AuthorityLimits:      response.AuthorityLimits,
		AgentHandoff: &AgentHandoff{
			AnswerSummary:               firstNonEmpty(strings.TrimSpace(response.AgentHandoff.Summary), "semantic_search returned citation-bearing module results"),
			Evidence:                    response.AgentHandoff.EvidenceInspected,
			ValidationBoundaries:        firstNonEmpty(strings.TrimSpace(response.ValidationBoundaries), "explicit semantic_search through verified optional module; default search remains lexical"),
			AuthorityLimits:             firstNonEmpty(strings.TrimSpace(response.AuthorityLimits), "semantic similarity is retrieval evidence only; canonical citations remain authority"),
			FollowUpPrimitiveInspection: firstNonEmpty(strings.TrimSpace(response.AgentHandoff.FollowUpPrimitiveInspection), "use get_document, provenance_events, and projection_states for authority drill-down before durable writes"),
		},
	}
	if result.Query == "" {
		result.Query = response.Query
	}
	if result.SearchStatus == "" {
		result.SearchStatus = "completed"
	}
	if result.Ranking == "" {
		result.Ranking = "module_hybrid_vector_lexical"
	}
	return result, nil
}

func semanticModuleBlockedResult(options SemanticSearchOptions, status string, err error) SemanticSearchResult {
	provider := options.Provider
	if provider == "" {
		provider = runclient.SemanticModuleProviderOllama
	}
	return SemanticSearchResult{
		Query:         options.Query,
		PathPrefix:    options.PathPrefix,
		Tag:           options.Tag,
		MetadataKey:   options.MetadataKey,
		MetadataValue: options.MetadataValue,
		Provider: SemanticProviderStatus{
			Provider:     provider,
			Model:        options.EmbeddingModel,
			Status:       "provider_blocked",
			ErrorSummary: semanticModuleErrorSummary(err),
		},
		Cache:                SemanticCacheStatus{Status: "not_used"},
		Ranking:              "module_verified_only",
		SearchStatus:         "provider_blocked",
		PrivacyDisclosure:    "semantic_search uses verified optional embedding modules only; no hidden provider fallback is used",
		ValidationBoundaries: "explicit semantic_search mode; default search remains lexical; install and enable an Ollama or Gemini semantic module before semantic ranking",
		AuthorityLimits:      "semantic similarity is retrieval evidence only; canonical markdown citations and approved OpenClerk runner writes remain authority",
		AgentHandoff: &AgentHandoff{
			AnswerSummary:               fmt.Sprintf("semantic_search blocked before ranking: %s", status),
			ValidationBoundaries:        "verified optional module required; no provider fallback or default semantic ranking promotion",
			AuthorityLimits:             "use lexical search and canonical citations as authority until a module returns citation-bearing evidence",
			FollowUpPrimitiveInspection: "install/configure the provider module, then rerun semantic_search; use lexical search for authoritative retrieval meanwhile",
		},
	}
}

func semanticModuleErrorSummary(err error) string {
	if err == nil {
		return ""
	}
	return firstNString(err.Error(), 240)
}

func firstNString(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
