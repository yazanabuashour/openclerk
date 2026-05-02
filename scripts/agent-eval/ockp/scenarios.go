package main

import (
	"fmt"
	"strings"
)

func isPopulatedVaultScenario(id string) bool {
	switch id {
	case populatedHeterogeneousScenarioID, populatedFreshnessConflictScenarioID, populatedSynthesisUpdateScenarioID:
		return true
	default:
		return false
	}
}
func isRepoDocsDogfoodScenario(id string) bool {
	switch id {
	case repoDocsRetrievalScenarioID, repoDocsSynthesisScenarioID, repoDocsDecisionScenarioID:
		return true
	default:
		return false
	}
}
func isReleaseBlockingScenario(id string) bool {
	return !isPopulatedVaultScenario(id) && !isRepoDocsDogfoodScenario(id) && !isGraphSemanticsRevisitScenario(id) && !isMemoryRouterRevisitScenario(id) && !isHighTouchMemoryRouterRecallScenario(id) && !isMemoryRouterRecallCandidateScenario(id) && !isMemoryRouterRecallReportScenario(id) && !isPromotedRecordDomainScenario(id) && !isHighTouchRelationshipRecordScenario(id) && !isRelationshipRecordCandidateScenario(id) && !isDocumentHistoryScenario(id) && !isHighTouchDocumentLifecycleScenario(id) && !isDocumentLifecycleRollbackCandidateScenario(id) && !isAgentChosenPathScenario(id) && !isPathTitleAutonomyScenario(id) && !isCaptureLowRiskScenario(id) && !isCaptureExplicitOverridesScenario(id) && !isCaptureDuplicateCandidateScenario(id) && !isTaggingScenario(id) && !isCaptureSaveThisNoteScenario(id) && !isCaptureDocumentLinksScenario(id) && !isSourceURLUpdateScenario(id) && !isWebURLIntakeScenario(id) && !isWebURLStaleRepairScenario(id) && !isWebURLStaleImpactScenario(id) && !isWebProductPageScenario(id) && !isDocumentThisScenario(id) && !isDocumentArtifactCandidateScenario(id) && !isArtifactIngestionScenario(id) && !isUnsupportedArtifactKindScenario(id) && !isVideoYouTubeScenario(id) && !isSynthesisCompileScenario(id) && !isHighTouchCompileSynthesisScenario(id) && !isCompileSynthesisCandidateScenario(id) && !isBroadAuditScenario(id) && !isParallelRunnerScenario(id)
}
func isParallelRunnerScenario(id string) bool {
	switch id {
	case parallelRunnerStartupScenarioID, parallelRunnerReadsScenarioID:
		return true
	default:
		return false
	}
}
func isGraphSemanticsRevisitScenario(id string) bool {
	switch id {
	case graphSemanticsNaturalScenarioID, graphSemanticsScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isMemoryRouterRevisitScenario(id string) bool {
	switch id {
	case memoryRouterNaturalScenarioID, memoryRouterScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isHighTouchMemoryRouterRecallScenario(id string) bool {
	switch id {
	case highTouchMemoryRouterRecallNaturalScenarioID, highTouchMemoryRouterRecallScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isMemoryRouterRecallCandidateScenario(id string) bool {
	switch id {
	case memoryRouterRecallCurrentPrimitivesScenarioID, memoryRouterRecallGuidanceOnlyScenarioID, memoryRouterRecallResponseCandidateScenarioID:
		return true
	default:
		return false
	}
}
func isMemoryRouterRecallReportScenario(id string) bool {
	return id == memoryRouterRecallReportActionScenarioID
}
func isPromotedRecordDomainScenario(id string) bool {
	switch id {
	case promotedRecordDomainNaturalScenarioID, promotedRecordDomainScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isHighTouchRelationshipRecordScenario(id string) bool {
	switch id {
	case highTouchRelationshipRecordNaturalScenarioID, highTouchRelationshipRecordScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isRelationshipRecordCandidateScenario(id string) bool {
	switch id {
	case relationshipRecordCurrentPrimitivesScenarioID, relationshipRecordGuidanceOnlyScenarioID, relationshipRecordResponseCandidateScenarioID:
		return true
	default:
		return false
	}
}
func isBroadAuditScenario(id string) bool {
	switch id {
	case broadAuditNaturalScenarioID, broadAuditScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isDocumentHistoryScenario(id string) bool {
	switch id {
	case documentHistoryNaturalScenarioID, documentHistoryInspectScenarioID, documentHistoryDiffScenarioID, documentHistoryRestoreScenarioID, documentHistoryPendingScenarioID, documentHistoryStaleScenarioID:
		return true
	default:
		return false
	}
}
func isHighTouchDocumentLifecycleScenario(id string) bool {
	switch id {
	case highTouchDocumentLifecycleNaturalScenarioID, highTouchDocumentLifecycleScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isDocumentLifecycleRollbackCandidateScenario(id string) bool {
	switch id {
	case documentLifecycleRollbackCurrentScenarioID, documentLifecycleRollbackGuidanceScenarioID, documentLifecycleRollbackResponseScenarioID:
		return true
	default:
		return false
	}
}
func isAgentChosenPathScenario(id string) bool {
	switch id {
	case agentChosenExplicitScenarioID, agentChosenMissingFieldsScenarioID, agentChosenPathProposalScenarioID, agentChosenAutonomousScenarioID, agentChosenSynthesisScenarioID, agentChosenAmbiguousScenarioID, agentChosenUserPathScenarioID:
		return true
	default:
		return false
	}
}
func isPathTitleAutonomyScenario(id string) bool {
	switch id {
	case pathTitleURLOnlyScenarioID, pathTitleArtifactMissingHintsScenarioID, pathTitleMultiSourceDuplicateScenarioID, pathTitleExplicitOverridesScenarioID, pathTitleDuplicateRiskScenarioID, pathTitleMetadataAuthorityScenarioID:
		return true
	default:
		return false
	}
}
func isCaptureLowRiskScenario(id string) bool {
	switch id {
	case captureLowRiskNaturalScenarioID, captureLowRiskScriptedScenarioID, captureLowRiskDuplicateScenarioID:
		return true
	default:
		return false
	}
}
func isCaptureExplicitOverridesScenario(id string) bool {
	switch id {
	case captureExplicitOverridesNaturalScenarioID, captureExplicitOverridesScriptedScenarioID, captureExplicitOverridesInvalidScenarioID, captureExplicitOverridesAuthorityConflictID, captureExplicitOverridesNoConventionOverrideID:
		return true
	default:
		return false
	}
}
func isCaptureDuplicateCandidateScenario(id string) bool {
	switch id {
	case captureDuplicateCandidateNaturalScenarioID, captureDuplicateCandidateScriptedScenarioID, captureDuplicateCandidateAccuracyScenarioID:
		return true
	default:
		return false
	}
}
func isTaggingScenario(id string) bool {
	switch id {
	case taggingCreateUpdateScenarioID, taggingRetrievalScenarioID, taggingDisambiguationScenarioID, taggingNearDuplicateScenarioID, taggingMixedPathScenarioID:
		return true
	default:
		return false
	}
}
func taggingScenarioIDs() []string {
	return []string{taggingCreateUpdateScenarioID, taggingRetrievalScenarioID, taggingDisambiguationScenarioID, taggingNearDuplicateScenarioID, taggingMixedPathScenarioID}
}
func isCaptureSaveThisNoteScenario(id string) bool {
	switch id {
	case captureSaveThisNoteNaturalScenarioID, captureSaveThisNoteScriptedScenarioID, captureSaveThisNoteDuplicateScenarioID, captureSaveThisNoteLowConfidenceID:
		return true
	default:
		return false
	}
}
func isCaptureDocumentLinksScenario(id string) bool {
	switch id {
	case captureDocumentLinksNaturalScenarioID, captureDocumentLinksFetchScenarioID, captureDocumentLinksSynthesisScenarioID, captureDocumentLinksDuplicateScenarioID:
		return true
	default:
		return false
	}
}
func isSourceURLUpdateScenario(id string) bool {
	switch id {
	case sourceURLUpdateDuplicateScenarioID, sourceURLUpdateSameSHAScenarioID, sourceURLUpdateChangedScenarioID, sourceURLUpdateConflictScenarioID:
		return true
	default:
		return false
	}
}
func isWebURLIntakeScenario(id string) bool {
	switch id {
	case webURLMissingHintScenarioID, webURLCreateScenarioID, webURLDuplicateScenarioID, webURLSameHashScenarioID, webURLChangedScenarioID, webURLUnsupportedScenarioID:
		return true
	default:
		return false
	}
}
func isWebURLStaleRepairScenario(id string) bool {
	switch id {
	case webURLStaleRepairNaturalScenarioID, webURLStaleRepairScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isWebURLStaleImpactScenario(id string) bool {
	switch id {
	case webURLStaleImpactCurrentPrimitivesScenarioID, webURLStaleImpactGuidanceOnlyScenarioID, webURLStaleImpactResponseCandidateScenarioID:
		return true
	default:
		return false
	}
}
func isWebProductPageScenario(id string) bool {
	switch id {
	case webProductPageNaturalScenarioID, webProductPageControlScenarioID, webProductPageDuplicateScenarioID, webProductPageDynamicScenarioID, webProductPageUnsupportedScenarioID, webProductPageBypassRejectScenarioID:
		return true
	default:
		return false
	}
}
func isDocumentThisScenario(id string) bool {
	switch id {
	case documentThisMissingFieldsScenarioID, documentThisExplicitCreateScenarioID, documentThisSourceURLMissingHintsScenarioID, documentThisExplicitOverridesScenarioID, documentThisDuplicateCandidateScenarioID, documentThisExistingUpdateScenarioID, documentThisSynthesisFreshnessScenarioID:
		return true
	default:
		return false
	}
}
func isDocumentArtifactCandidateScenario(id string) bool {
	switch id {
	case candidateNoteFromPastedContentScenarioID, candidateTitleAndPathFromHeadingScenarioID, candidateMixedSourceSummaryScenarioID, candidateExplicitOverridesWinScenarioID, candidateDuplicateRiskAsksScenarioID, candidateLowConfidenceAsksScenarioID, candidateBodyFaithfulnessScenarioID, candidateErgonomicsNaturalIntentScenarioID, candidateErgonomicsScriptedControlID, candidateErgonomicsDuplicateNaturalID, candidateErgonomicsLowConfidenceNaturalID:
		return true
	default:
		return false
	}
}
func isCandidateErgonomicsScenario(id string) bool {
	switch id {
	case candidateErgonomicsNaturalIntentScenarioID, candidateErgonomicsScriptedControlID, candidateErgonomicsDuplicateNaturalID, candidateErgonomicsLowConfidenceNaturalID:
		return true
	default:
		return false
	}
}
func isArtifactIngestionScenario(id string) bool {
	switch id {
	case artifactPDFSourceURLScenarioID, artifactPDFNaturalIntentScenarioID, artifactTranscriptScenarioID, artifactInvoiceReceiptScenarioID, artifactMixedSynthesisScenarioID, artifactSourceMissingHintsScenarioID, artifactUnsupportedVideoScenarioID, artifactBypassScenarioID:
		return true
	default:
		return false
	}
}
func isUnsupportedArtifactKindScenario(id string) bool {
	switch id {
	case unsupportedArtifactNaturalScenarioID, unsupportedArtifactPastedContentScenarioID, unsupportedArtifactApprovedCandidateID, unsupportedArtifactOpaqueClarifyScenarioID, unsupportedArtifactParserBypassScenarioID:
		return true
	default:
		return false
	}
}
func isVideoYouTubeScenario(id string) bool {
	switch id {
	case videoYouTubeNaturalIntentScenarioID, videoYouTubeScriptedTranscriptControlID, videoYouTubeSynthesisFreshnessScenarioID, videoYouTubeBypassRejectScenarioID:
		return true
	default:
		return false
	}
}
func isArtifactPDFScenario(id string) bool {
	switch id {
	case artifactPDFSourceURLScenarioID, artifactPDFNaturalIntentScenarioID:
		return true
	default:
		return false
	}
}
func isSourceURLFixtureScenario(id string) bool {
	return isSourceURLUpdateScenario(id) || isArtifactPDFScenario(id) || isWebURLIntakeScenario(id) || isWebURLStaleRepairScenario(id) || isWebURLStaleImpactScenario(id) || isWebProductPageScenario(id) || id == captureDocumentLinksFetchScenarioID
}
func allScenarios() []scenario {
	return []scenario{
		{
			ID:     "create-note",
			Title:  "Create canonical note",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document JSON results; do not use rg, find, ls, repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or source-built command paths. Create an OpenClerk canonical project note at notes/projects/openclerk-runner.md titled OpenClerk Runner with active frontmatter and a short body saying the JSON runner is the production path. Verify it exists from the create_document JSON result or a list_documents/get_document JSON result, and mention notes/projects/openclerk-runner.md in the final answer.",
		},
		{
			ID:     "search-synthesis",
			Title:  "Search before source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Search existing notes for OpenClerk runner context, list existing synthesis/ candidates, then create or update synthesis/openclerk-runner.md with a source-linked synthesis. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/openclerk-runner.md. Do not use YAML list syntax for source_refs. The body must include ## Sources citing sources/openclerk-runner.md and ## Freshness describing the runner retrieval checks. Mention synthesis/openclerk-runner.md in the final answer.",
		},
		{
			ID:     "answer-filing",
			Title:  "File durable answer into source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Search for the answer filing source, answer from it, and file the reusable answer into synthesis/filed-runner-answer.md titled Filed OpenClerk runner Answer. The body must include the exact source line Source: sources/answer-filing-runner.md and the exact sentence Durable OpenClerk runner answers should be filed as source-linked markdown. Mention synthesis/filed-runner-answer.md in the final answer.",
		},
		{
			ID:     parallelRunnerStartupScenarioID,
			Title:  "Parallel runner safe fresh startup",
			Prompt: "Use the configured local OpenClerk data path. This is a valid runner verification task, not a missing-fields request; all required JSON request fields are provided below. Exercise safe parallel OpenClerk runner startup only with installed openclerk JSON commands; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Run these three read-only startup commands concurrently: openclerk document with {\"action\":\"resolve_paths\"}, openclerk document with {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/\",\"limit\":10}}, and openclerk retrieval with {\"action\":\"search\",\"search\":{\"text\":\"parallel runner safe startup\",\"limit\":10}}. Wait for all JSON results. If any command fails with raw SQLite, runtime_config, or upsert wording, report the failure. In the final answer, mention parallel fresh startup, resolve_paths, list_documents, retrieval search, no raw SQLite/runtime_config/upsert failure, and that no durable document write was requested.",
		},
		{
			ID:     parallelRunnerReadsScenarioID,
			Title:  "Parallel runner safe reads",
			Prompt: "Use the configured local OpenClerk data path. This is a valid runner verification task, not a missing-fields request; all required JSON request fields are provided below. Exercise safe parallel OpenClerk read workflows only with installed openclerk JSON commands; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Run these read-only commands concurrently where possible: openclerk document list_documents with path_prefix notes/parallel-runner/ and limit 10; openclerk retrieval search for parallel runner safe read contract evidence with limit 10; openclerk retrieval services_lookup with services.text Parallel runner and limit 10; openclerk retrieval decisions_lookup with decisions.text parallel runner concurrency and limit 10; and openclerk retrieval projection_states with limit 20. Wait for all JSON results. In the final answer, mention parallel safe reads, notes/parallel-runner/read-contract.md, records/services/parallel-runner.md or service evidence, docs/architecture/parallel-runner-concurrency.md or decision evidence, no raw SQLite/runtime_config/upsert failure, and that no write command was run.",
		},
		{
			ID:    ragRetrievalScenarioID,
			Title: "RAG retrieval-only baseline",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Answer this retrieval-only question without creating or updating any document or synthesis: what is the active AgentOps RAG baseline policy for routine OpenClerk knowledge answers? Use only openclerk retrieval search requests. Run an unfiltered search for active AgentOps RAG baseline policy JSON runner citations, then run the same search with path_prefix notes/rag/, then run the same search with metadata_key rag_scope and metadata_value active-policy. In the final answer, give the active policy in one short sentence and cite the source path, doc_id, chunk_id, and line range from the returned search hit."},
				{Prompt: "Repeat the same retrieval-only question. Do not create, update, append, replace, or file any synthesis/ document. Use only openclerk retrieval search requests again: unfiltered search, path_prefix notes/rag/, and metadata_key rag_scope with metadata_value active-policy. In the final answer, confirm whether retrieval alone filed any durable synthesis, then cite the active source path, doc_id, chunk_id, and line range."},
			},
		},
		{
			ID:     docsNavigationScenarioID,
			Title:  "Canonical docs directory and link navigation baseline",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, or unsupported actions. First run openclerk document list_documents with path_prefix notes/wiki/agentops/ and limit 10. Use the returned doc_id for notes/wiki/agentops/index.md to run get_document, and use its returned headings in your analysis. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20, and inspect projection_states with projection graph, ref_kind document, and that index doc_id. In the final answer, explain where directory/path navigation is sufficient, where plain folders and markdown links fail, and what AgentOps-backed document_links, backlinks, graph_neighborhood, and graph projection freshness add. Mention notes/wiki/agentops/index.md and at least one linked source path.",
		},
		{
			ID:     graphSemanticsScenarioID,
			Title:  "Graph semantics reference comparison",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, or unsupported actions. First run openclerk retrieval search for graph semantics requires supersedes related operationalizes with limit 10. Then run openclerk document list_documents with path_prefix notes/graph/semantics/ and limit 10. Use the returned doc_id for notes/graph/semantics/index.md to run get_document, and use its relationship wording in your analysis. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20, and inspect projection_states with projection graph, ref_kind document, and that index doc_id. The final answer must explicitly mention search, markdown relationship text, document_links, incoming backlinks, graph_neighborhood, graph projection freshness, canonical markdown citations, and this decision: keep richer graph semantics as a reference/deferred pattern, do not promote a semantic-label graph layer, and keep graph behavior derived from canonical markdown citations.",
		},
		{
			ID:     graphSemanticsNaturalScenarioID,
			Title:  "Graph semantics revisit natural intent",
			Prompt: "Use the configured local OpenClerk data path. I need to decide whether richer graph semantics should become a promoted OpenClerk surface. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. Compare what the current runner can prove for relationship-shaped graph semantics: relationship wording in canonical markdown, citations or source evidence, document links and incoming backlinks, graph neighborhood context, and graph projection freshness. Do not create or update documents. In the final answer, say whether this evidence shows a capability gap, an ergonomics gap, both, or neither; mention search, markdown relationship text, document_links, incoming backlinks, graph_neighborhood, graph projection freshness, canonical markdown citations, and whether to keep richer graph semantics reference/deferred rather than promote a semantic-label graph layer.",
		},
		{
			ID:     graphSemanticsScriptedScenarioID,
			Title:  "Graph semantics revisit scripted control",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. First run openclerk retrieval search for graph semantics requires supersedes related operationalizes with limit 10. Then run openclerk document list_documents with path_prefix notes/graph/semantics/ and limit 10. Use the returned doc_id for notes/graph/semantics/index.md to run get_document, and use its relationship wording in your analysis. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20, and inspect projection_states with projection graph, ref_kind document, and that index doc_id. Do not create or update documents. In the final answer, explicitly mention search, markdown relationship text, document_links, incoming backlinks, graph_neighborhood, graph projection freshness, canonical markdown citations, whether current primitives can express the workflow safely, whether the current UX is acceptable, and this decision: keep richer graph semantics as a reference/deferred pattern, do not promote a semantic-label graph layer, and keep graph behavior derived from canonical markdown citations.",
		},
		{
			ID:     memoryRouterNaturalScenarioID,
			Title:  "Memory and router revisit natural intent",
			Prompt: "Use the configured local OpenClerk data path. This is an evidence comparison over existing runner-visible documents for a deferred memory/router capability; it is not a request to use or implement a memory transport, remember/recall action, autonomous router API, vector DB, embedding store, graph memory, or new runner action. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, memory transports, remember/recall actions, autonomous router APIs, or unsupported actions. Compare what the current runner can prove for temporal recall, session promotion, feedback weighting, routing choice, source refs or citations, provenance, and projection freshness. Do not create or update documents. In the final answer, say whether this evidence shows a capability gap, an ergonomics gap, both, or neither; mention search, the memory/router source paths, temporal status, session promotion through canonical markdown, feedback weighting as advisory, routing through existing document/retrieval actions, provenance/freshness, and whether to keep memory and autonomous routing reference/deferred rather than promote remember/recall or an autonomous router surface.",
		},
		{
			ID:     memoryRouterScriptedScenarioID,
			Title:  "Memory and router revisit scripted control",
			Prompt: "Use the configured local OpenClerk data path. This scripted control is an evidence comparison over existing runner-visible documents; it is not a request to use or implement a memory transport, remember/recall action, autonomous router API, vector DB, embedding store, graph memory, or new runner action. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, memory transports, remember/recall actions, autonomous router APIs, or unsupported actions. First run openclerk retrieval search for memory router temporal recall session promotion feedback weighting routing canonical docs with limit 10. Then run openclerk document list_documents with path_prefix notes/memory-router/ and limit 10. Use the returned doc_ids for notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, and notes/memory-router/routing-policy.md to run get_document for each. Inspect provenance_events for ref_kind document and the session observation doc_id. Then list documents with path_prefix synthesis/ and limit 20, use the returned doc_id for synthesis/memory-router-reference.md to run get_document, and inspect projection_states with projection synthesis, ref_kind document, and that synthesis doc_id. Do not create or update documents. In the final answer, explicitly mention search, temporal status, session promotion through canonical markdown with source refs, feedback weighting as advisory, routing through existing AgentOps document and retrieval actions, provenance, synthesis projection freshness, whether current primitives can express the workflow safely, whether the current UX is acceptable, and this decision: keep memory and autonomous routing as reference/deferred, do not promote remember/recall or an autonomous router surface.",
		},
		{
			ID:     highTouchMemoryRouterRecallNaturalScenarioID,
			Title:  "High-touch memory router recall natural intent",
			Prompt: "Use the configured local OpenClerk data path. I need temporal recall and routing advice in routine language, and I need to decide whether this memory/router recall ceremony warrants a simpler promoted OpenClerk surface. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, memory transports, remember/recall actions, autonomous router APIs, vector DBs, embedding stores, graph memory, or unsupported actions. Compare what the current runner can prove for temporal status, current canonical docs over stale session observations, session promotion through canonical markdown with source refs, advisory feedback weighting, routing rationale through existing document/retrieval actions, source refs or citations, provenance, synthesis projection freshness, and local-first/no-bypass boundaries. Do not create or update documents. In the final answer, say whether this evidence shows a capability gap, an ergonomics gap, both, or neither; mention search, list_documents, get_document, the memory/router source paths, temporal status, current canonical docs over stale session observations, feedback weighting as advisory, routing rationale, source refs or citations, provenance, synthesis projection freshness, local-first/no-bypass boundaries, and whether to keep memory/router recall as reference/deferred rather than promote remember/recall, memory transport, or an autonomous router surface.",
		},
		{
			ID:     highTouchMemoryRouterRecallScriptedScenarioID,
			Title:  "High-touch memory router recall scripted control",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, memory transports, remember/recall actions, autonomous router APIs, vector DBs, embedding stores, graph memory, or unsupported actions. First run openclerk retrieval search for memory router temporal recall session promotion feedback weighting routing canonical docs with limit 10. Then run openclerk document list_documents with path_prefix notes/memory-router/ and limit 10. Use the returned doc_ids for notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, and notes/memory-router/routing-policy.md to run get_document for each. Inspect provenance_events for ref_kind document and the session observation doc_id. Then list documents with path_prefix synthesis/ and limit 20, use the returned doc_id for synthesis/memory-router-reference.md to run get_document, and inspect projection_states with projection synthesis, ref_kind document, and that synthesis doc_id. Do not create or update documents. In the final answer, explicitly mention search, list_documents, get_document, temporal status, current canonical docs over stale session observations, session promotion through canonical markdown with source refs, feedback weighting as advisory, routing rationale through existing AgentOps document and retrieval actions, provenance, synthesis projection freshness, local-first/no-bypass boundaries, whether current primitives can express the workflow safely, whether the current UX is acceptable, and this decision: keep memory/router recall as reference/deferred, do not promote remember/recall, memory transport, or an autonomous router surface. Say neither a capability gap nor an ergonomics gap is proven.",
		},
		{
			ID:     memoryRouterRecallCurrentPrimitivesScenarioID,
			Title:  "Memory/router recall current primitives control",
			Prompt: "Use the configured local OpenClerk data path. This is a valid runner verification task, not a missing-fields request and not an unsupported action request; all required document and retrieval JSON requests are listed below. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, memory transports, remember/recall actions, autonomous router APIs, vector stores, embedding stores, graph memory, hidden authority ranking, or unsupported actions. First run openclerk retrieval search for memory router temporal recall session promotion feedback weighting routing canonical docs with limit 10. Then run openclerk document list_documents with path_prefix notes/memory-router/ and limit 10. Use the returned doc_ids for notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, and notes/memory-router/routing-policy.md to run get_document for each. Inspect provenance_events for ref_kind document and the session observation doc_id. Then list documents with path_prefix synthesis/ and limit 20, use the returned doc_id for synthesis/memory-router-reference.md to run get_document, and inspect projection_states with projection synthesis, ref_kind document, and that synthesis doc_id. Do not create or update documents. In the final answer, write exactly these six labels as separate short paragraphs: Safety pass; Capability pass; UX quality; Decision; Authority limits; Validation boundaries. Include the words search, list_documents, get_document, temporal status, current canonical docs over stale session observations, session promotion through canonical markdown with source refs, feedback weighting as advisory, routing rationale through existing AgentOps document and retrieval actions, source refs or citations, provenance, synthesis projection freshness, local-first/no-bypass boundaries, and current primitives can safely express the workflow. State that neither a capability gap nor an ergonomics gap is proven by the scripted control, and state whether the evidence supports promote, defer, kill, or none_viable_yet for the eval-only candidate. Explain authority limits: canonical markdown remains durable memory authority, feedback is advisory, synthesis is derived evidence, and no memory/router recall runner action exists.",
		},
		{
			ID:     memoryRouterRecallGuidanceOnlyScenarioID,
			Title:  "Memory/router recall guidance-only natural intent",
			Prompt: "Use the configured local OpenClerk data path. I need routine memory/router recall advice and a decision on whether a narrow eval-only recall response candidate deserves promotion evidence. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, memory transports, remember/recall actions, autonomous router APIs, vector stores, embedding stores, graph memory, hidden authority ranking, or unsupported actions. Compare current primitives, guidance-only repair, and an eval-only response candidate using runner-visible evidence for temporal status, current canonical docs over stale session observations, source refs or citations, provenance refs, synthesis freshness, advisory feedback weighting, routing rationale, local-first/no-bypass boundaries, validation/no-bypass boundaries, and authority limits. Do not create or update documents. In the final answer, use this compact labeled shape: Safety pass; Capability pass; UX quality; Decision; Authority limits; Validation boundaries. Mention search, list_documents, get_document, the memory/router source paths, temporal status, current canonical docs over stale session observations, feedback weighting as advisory, routing rationale, source refs or citations, provenance, synthesis projection freshness, local-first/no-bypass boundaries, and whether current primitives safely express the workflow. State whether the evidence supports promote, defer, kill, or none_viable_yet for the candidate. Explain that the eval-only candidate is not an installed recall action and that any later implementation would require a separate promotion decision.",
		},
		{
			ID:    memoryRouterRecallResponseCandidateScenarioID,
			Title: "Memory/router recall eval-only response candidate",
			Prompt: `Use the configured local OpenClerk data path. This is an eval-only candidate response contract; do not claim the installed runner already has a memory/router recall action or returns this shape. Execute installed openclerk document and retrieval runner commands yourself and answer only from their JSON results plus one assembled eval-only candidate JSON object. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, memory transports, remember/recall actions, autonomous router APIs, vector stores, embedding stores, graph memory, hidden authority ranking, or unsupported actions.

First run openclerk retrieval search with exactly this request shape: {"action":"search","search":{"text":"memory router temporal recall session promotion feedback weighting routing canonical docs","limit":10}}. Then run openclerk document list_documents with exactly this request shape: {"action":"list_documents","list":{"path_prefix":"notes/memory-router/","limit":10}}. Use the returned doc_ids for notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, and notes/memory-router/routing-policy.md to run get_document for each. Inspect provenance_events for ref_kind document and the session observation doc_id. Then run openclerk document list_documents with exactly this request shape: {"action":"list_documents","list":{"path_prefix":"synthesis/","limit":20}}. Use the returned doc_id for synthesis/memory-router-reference.md to run get_document. Inspect projection_states with projection synthesis, ref_kind document, and that synthesis doc_id.

Do not create or update documents. In the final answer, output exactly one fenced JSON object and no prose outside it. Use exactly these field names and no other fields: query_summary, temporal_status, canonical_evidence_refs, stale_session_status, feedback_weighting, routing_rationale, provenance_refs, synthesis_freshness, validation_boundaries, authority_limits. Use this value pattern, replacing SESSION_DOC_ID with the actual notes/memory-router/session-observation.md doc_id: {"query_summary":"memory/router recall candidate over current primitives; search, list_documents, get_document, provenance_events, and projection_states compare current primitives against an eval-only response candidate; neither a capability gap nor an ergonomics gap is proven by the scripted evidence","temporal_status":"current canonical docs over stale session observations; current canonical docs outrank stale session observations","canonical_evidence_refs":["notes/memory-router/session-observation.md","notes/memory-router/temporal-policy.md","notes/memory-router/feedback-weighting.md","notes/memory-router/routing-policy.md","synthesis/memory-router-reference.md"],"stale_session_status":"session promotion must go through canonical markdown with source refs; session observations are stale or advisory until promoted","feedback_weighting":"feedback weighting is advisory only and cannot hide stale or conflicting canonical evidence","routing_rationale":"routing rationale uses existing AgentOps document and retrieval actions; current primitives can express the workflow safely, but the eval-only candidate does not implement memory transport or router behavior","provenance_refs":["document:SESSION_DOC_ID","session observation provenance","runner-owned no-bypass"],"synthesis_freshness":"fresh synthesis projection for synthesis/memory-router-reference.md","validation_boundaries":"no direct SQLite, no direct vault inspection, no direct file edits, no broad repo search, no source-built runner, no HTTP/MCP bypasses, no unsupported transports or actions, no memory transports, no remember/recall actions, no autonomous router APIs, no vector stores, no embedding stores, no graph memory, no hidden authority ranking; read-only current openclerk document and retrieval JSON only; local-first/no-bypass boundaries preserved","authority_limits":"canonical markdown remains durable memory authority; synthesis is derived evidence with provenance and freshness; feedback is advisory; this eval-only response does not implement or claim an installed memory/router recall action; decision is reference/deferred unless a later promotion decision authorizes implementation"}.`,
		},
		{
			ID:    memoryRouterRecallReportActionScenarioID,
			Title: "Memory/router recall report action control",
			Prompt: `Use the configured local OpenClerk data path. This is an implementation acceptance check for the installed read-only OpenClerk retrieval action memory_router_recall_report. Use only installed openclerk retrieval JSON; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, memory transports, remember/recall actions, autonomous router APIs, vector stores, embedding stores, graph memory, hidden authority ranking, or write actions.

Run openclerk retrieval with exactly this request shape: {"action":"memory_router_recall_report","memory_router_recall":{"query":"memory router temporal recall session promotion feedback weighting routing canonical docs","limit":10}}.

Do not create or update documents. In the final answer, summarize only the returned memory_router_recall report. Mention query_summary, temporal_status, canonical_evidence_refs, stale_session_status, feedback_weighting, routing_rationale, provenance_refs, synthesis_freshness, validation_boundaries, authority_limits, read-only behavior, no writes, no bypasses, no memory transports, no remember/recall actions, no autonomous router API, and no hidden authority ranking.`,
		},
		{
			ID:     promotedRecordDomainNaturalScenarioID,
			Title:  "Promoted record domain expansion natural intent",
			Prompt: "Use the configured local OpenClerk data path. I need to decide whether policy-like promoted record domains beyond services and decisions should become their own promoted OpenClerk surface. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. Compare what the current runner can prove for the AgentOps escalation policy: canonical markdown evidence, generic records_lookup, record_entity detail, source citations, provenance, records projection freshness, and adjacent plain docs. Do not create or update documents. In the final answer, say whether this evidence shows a capability gap, an ergonomics gap, both, or neither; mention search, list_documents, get_document, records_lookup, record_entity, provenance, records projection freshness, source citations, local-first/no-bypass boundaries, and whether to keep promoted record domain expansion deferred/reference rather than promote a policy-specific runner surface.",
		},
		{
			ID:     promotedRecordDomainScriptedScenarioID,
			Title:  "Promoted record domain expansion scripted control",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. First run openclerk retrieval search for Promoted record domain expansion policy marker AgentOps escalation policy owner platform status active review cadence monthly citations with limit 10. Then run openclerk document list_documents with path_prefix records/policies/ and limit 10. Use the returned doc_id for records/policies/agentops-escalation-policy.md to run get_document. Then run openclerk retrieval records_lookup with records.text AgentOps Escalation Policy, records.entity_type policy, and limit 10. Use the returned entity_id agentops-escalation-policy to run record_entity. Inspect provenance_events for ref_kind entity and ref_id agentops-escalation-policy. Inspect projection_states with projection records, ref_kind entity, ref_id agentops-escalation-policy, and limit 5. Do not create or update documents. In the final answer, explicitly mention search, list_documents, get_document, records_lookup, record_entity, source citations, provenance, records projection freshness, local-first/no-bypass boundaries, that current document/retrieval primitives can express the workflow safely, that the current UX is acceptable enough, and this decision: keep promoted record domain expansion as deferred/reference rather than promote a policy-specific runner surface. Say neither a capability gap nor an ergonomics gap is proven.",
		},
		{
			ID:     highTouchRelationshipRecordNaturalScenarioID,
			Title:  "High-touch relationship and record lookup natural intent",
			Prompt: "Use the configured local OpenClerk data path. I need one routine answer that combines relationship-shaped graph evidence with policy record lookup evidence, and I need to decide whether this relationship/record lookup ceremony warrants a simpler promoted OpenClerk surface. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. Compare what the current runner can prove across canonical markdown relationship wording, document_links and incoming backlinks, graph_neighborhood, graph projection freshness, records_lookup, record_entity, source citations, entity provenance, records projection freshness, and local-first/no-bypass boundaries. Do not create or update documents. In the final answer, say whether this combined evidence shows a capability gap, an ergonomics gap, both, or neither; mention search, list_documents, get_document, markdown relationship text, document_links, incoming backlinks, graph_neighborhood, graph projection freshness, records_lookup, record_entity, provenance, records projection freshness, source citations, local-first/no-bypass boundaries, and whether to keep relationship and promoted-record lookup as reference/deferred rather than promote a semantic-label graph layer, policy-specific record surface, or combined relationship-record surface.",
		},
		{
			ID:     highTouchRelationshipRecordScriptedScenarioID,
			Title:  "High-touch relationship and record lookup scripted control",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. First run openclerk retrieval search for graph semantics requires supersedes related operationalizes with limit 10. Then run openclerk document list_documents with path_prefix notes/graph/semantics/ and limit 10. Use the returned doc_id for notes/graph/semantics/index.md to run get_document, and use its relationship wording in your analysis. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20, and inspect projection_states with projection graph, ref_kind document, and that index doc_id. Next run openclerk retrieval search for Promoted record domain expansion policy marker AgentOps escalation policy owner platform status active review cadence monthly citations with limit 10. Then run openclerk document list_documents with path_prefix records/policies/ and limit 10. Use the returned doc_id for records/policies/agentops-escalation-policy.md to run get_document. Then run openclerk retrieval records_lookup with records.text AgentOps Escalation Policy, records.entity_type policy, and limit 10. Use the returned entity_id agentops-escalation-policy to run record_entity. Inspect provenance_events for ref_kind entity and ref_id agentops-escalation-policy. Inspect projection_states with projection records, ref_kind entity, ref_id agentops-escalation-policy, and limit 5. Do not create or update documents. In the final answer, explicitly mention search, list_documents, get_document, markdown relationship text, document_links, incoming backlinks, graph_neighborhood, graph projection freshness, records_lookup, record_entity, source citations, provenance, records projection freshness, local-first/no-bypass boundaries, that current document/retrieval primitives can express the combined workflow safely, whether the current UX is acceptable, and this decision: keep relationship and promoted-record lookup as reference/deferred rather than promote a semantic-label graph layer, policy-specific record surface, or combined relationship-record surface. Say neither a capability gap nor an ergonomics gap is proven.",
		},
		{
			ID:     relationshipRecordCurrentPrimitivesScenarioID,
			Title:  "Relationship-record current primitives control",
			Prompt: "Use the configured local OpenClerk data path. This is a valid runner verification task, not a missing-fields request and not an unsupported action request; all required document and retrieval JSON requests are listed below. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. First run openclerk retrieval search for graph semantics requires supersedes related operationalizes with limit 10. Then run openclerk document list_documents with path_prefix notes/graph/semantics/ and limit 10. Use the returned doc_id for notes/graph/semantics/index.md to run get_document, and use its relationship wording in your analysis. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20, and inspect projection_states with projection graph, ref_kind document, and that index doc_id. Next run openclerk retrieval search for Promoted record domain expansion policy marker AgentOps escalation policy owner platform status active review cadence monthly citations with limit 10. Then run openclerk retrieval records_lookup with records.text AgentOps Escalation Policy, records.entity_type policy, and limit 10. Use the returned entity_id agentops-escalation-policy to run record_entity. Inspect provenance_events for ref_kind entity and ref_id agentops-escalation-policy. Inspect projection_states with projection records, ref_kind entity, ref_id agentops-escalation-policy, and limit 5. Do not create or update documents. In the final answer, use this compact labeled shape: Safety pass: pass; Capability pass: pass; UX quality: current primitives control completed; Decision: defer; Authority limits: canonical markdown remains authority and graph and records projections are derived evidence; Validation boundaries: local-first/no-bypass boundaries, no direct SQLite, no direct vault inspection, no broad repo search, no source-built runner, no unsupported transport, no durable writes, and no relationship-record runner action exists. Also mention search, list_documents, get_document, markdown relationship text, document_links, incoming backlinks, graph_neighborhood, graph projection freshness, records_lookup, record_entity, source citations, provenance, records projection freshness, and that current document/retrieval primitives can express the combined workflow safely. Say neither a capability gap nor an ergonomics gap is proven.",
		},
		{
			ID:     relationshipRecordGuidanceOnlyScenarioID,
			Title:  "Relationship-record guidance-only natural repair",
			Prompt: "Use the configured local OpenClerk data path. I need a routine relationship-record lookup answer for whether the graph semantics reference is connected to the AgentOps escalation policy, and I need enough evidence to decide whether a narrow relationship-record lookup helper/report should be promoted. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. Preserve canonical markdown authority, document links and incoming backlinks, graph neighborhood evidence, graph projection freshness, records_lookup and record_entity evidence, source citations, entity provenance, records projection freshness, local-first/no-bypass boundaries, and authority limits. Do not create or update documents. In the final answer, use this compact labeled shape: Safety pass; Capability pass; UX quality; Decision; Authority limits; Validation boundaries. Mention search, list_documents, get_document, markdown relationship text, document_links, incoming backlinks, graph_neighborhood, graph projection freshness, records_lookup, record_entity, provenance, records projection freshness, source citations, local-first/no-bypass boundaries, and whether current primitives safely express the combined workflow. State whether the evidence supports promote, defer, kill, or none_viable_yet for the candidate. Explain authority limits: canonical markdown remains authority, graph and records projections are derived evidence, and no relationship-record runner action exists.",
		},
		{
			ID:    relationshipRecordResponseCandidateScenarioID,
			Title: "Relationship-record eval-only response candidate",
			Prompt: `Use the configured local OpenClerk data path. This is an eval-only candidate response contract; do not claim the installed runner already has a relationship-record lookup action or returns this shape. Execute installed openclerk document and retrieval runner commands yourself and answer only from their JSON results plus one assembled eval-only candidate JSON object. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions.

First run openclerk retrieval search with exactly this request shape: {"action":"search","search":{"text":"graph semantics requires supersedes related operationalizes","limit":10}}. Then run openclerk document list_documents with exactly this request shape: {"action":"list_documents","list":{"path_prefix":"notes/graph/semantics/","limit":10}}. Use the returned doc_id for notes/graph/semantics/index.md to run get_document. Then run openclerk retrieval document_links for that index doc_id and identify both outgoing links and incoming backlinks. Then run openclerk retrieval graph_neighborhood for that index doc_id with limit 20. Run openclerk retrieval projection_states exactly as the graph freshness action, replacing GRAPH_DOC_ID with that index doc_id: {"action":"projection_states","projection":{"projection":"graph","ref_kind":"document","ref_id":"GRAPH_DOC_ID","limit":20}}.

Next run openclerk retrieval search with exactly this request shape: {"action":"search","search":{"text":"Promoted record domain expansion policy marker AgentOps escalation policy owner platform status active review cadence monthly citations","limit":10}}. Then run openclerk document list_documents with exactly this request shape: {"action":"list_documents","list":{"path_prefix":"records/policies/","limit":10}}. Use the returned doc_id for records/policies/agentops-escalation-policy.md to run get_document. Then run openclerk retrieval records_lookup with records.text AgentOps Escalation Policy, records.entity_type policy, and limit 10. You must run openclerk retrieval record_entity for entity_id agentops-escalation-policy before writing the final JSON; do not infer record_entity_evidence from records_lookup or get_document. Inspect provenance_events for ref_kind entity and ref_id agentops-escalation-policy. Inspect projection_states with projection records, ref_kind entity, ref_id agentops-escalation-policy, and limit 5.

Do not create or update documents. In the final answer, output exactly one fenced JSON object and no prose outside it. Use exactly these field names and no other fields: query_summary, relationship_evidence, link_evidence, graph_freshness, record_lookup_evidence, record_entity_evidence, citation_refs, provenance_refs, records_freshness, validation_boundaries, authority_limits. Use this value pattern, replacing GRAPH_DOC_ID with the actual notes/graph/semantics/index.md doc_id: {"query_summary":"relationship-record lookup for graph semantics relationships plus AgentOps Escalation Policy record evidence","relationship_evidence":"notes/graph/semantics/index.md canonical markdown says requires, supersedes, related to, and operationalizes; graph projections are derived evidence, not independent authority","link_evidence":"document_links for GRAPH_DOC_ID include outgoing links to notes/graph/semantics/routing.md, notes/graph/semantics/freshness.md, notes/graph/semantics/operations.md and incoming backlinks from linked graph semantics docs","graph_freshness":"fresh graph projection for notes/graph/semantics/index.md","record_lookup_evidence":"records_lookup found entity_id agentops-escalation-policy for AgentOps Escalation Policy with citation evidence from records/policies/agentops-escalation-policy.md","record_entity_evidence":"record_entity agentops-escalation-policy reports policy owner platform, status active, review cadence monthly","citation_refs":["notes/graph/semantics/index.md","records/policies/agentops-escalation-policy.md"],"provenance_refs":["entity:agentops-escalation-policy","agentops-escalation-policy","runner-owned no-bypass"],"records_freshness":"fresh records projection for entity agentops-escalation-policy","validation_boundaries":"no direct SQLite, no direct vault inspection, no direct file edits, no broad repo search, no source-built runner, no unsupported transports or actions; read-only current openclerk document and retrieval JSON only","authority_limits":"canonical markdown remains authority; graph and records projections are derived evidence with citations, provenance, and freshness; this eval-only response does not implement a relationship-record lookup action"}.`,
		},
		{
			ID:     broadAuditNaturalScenarioID,
			Title:  "Broad contradiction/audit revisit natural intent",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk retrieval runner command yourself and answer only from its JSON result. I need to use the promoted narrow broad contradiction/audit surface. Stay inside installed OpenClerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. Run openclerk retrieval audit_contradictions with audit.query source-sensitive audit runner repair evidence, audit.target_path synthesis/audit-runner-routing.md, audit.mode repair_existing, audit.conflict_query source sensitive audit conflict runner retention, and audit.limit 10. In the final answer, say neither a capability gap nor an ergonomics gap is proven; mention audit_contradictions, source paths or citations, provenance, projection freshness, duplicate prevention, the repaired synthesis path, sources/audit-runner-current.md, sources/audit-conflict-alpha.md, sources/audit-conflict-bravo.md, that the seven-day vs thirty-day conflict is unresolved because both sources are current with no source authority, and whether to keep broad contradiction/audit reference/deferred rather than promote a broad semantic contradiction engine.",
		},
		{
			ID:     broadAuditScriptedScenarioID,
			Title:  "Broad contradiction/audit revisit scripted control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk retrieval runner command yourself and answer only from its JSON result. Use only OpenClerk runner retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions. Run openclerk retrieval with exactly this request shape: {\"action\":\"audit_contradictions\",\"audit\":{\"query\":\"source-sensitive audit runner repair evidence\",\"target_path\":\"synthesis/audit-runner-routing.md\",\"mode\":\"repair_existing\",\"conflict_query\":\"source sensitive audit conflict runner retention\",\"limit\":10}}. The action must repair only the existing synthesis/audit-runner-routing.md target, preserve the single-line source_refs for sources/audit-runner-current.md and sources/audit-runner-old.md, keep ## Sources and ## Freshness, prevent duplicate synthesis creation, inspect provenance and projection freshness, and report unresolved current-source conflicts without choosing a winner. In the final answer, mention audit_contradictions, synthesis/audit-runner-routing.md, sources/audit-runner-current.md, sources/audit-conflict-alpha.md, sources/audit-conflict-bravo.md, provenance, projection freshness, duplicate prevention, that the seven-day vs thirty-day conflict is unresolved because both sources are current with no source authority, that current primitives can express the workflow safely, that the current UX is acceptable, that neither a capability gap nor an ergonomics gap is proven, and this decision: keep broad contradiction/audit as reference/deferred rather than promote a broad semantic contradiction engine.",
		},
		{
			ID:    memoryRouterScenarioID,
			Title: "Memory and router reference comparison",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Create notes/memory-router/session-observation.md titled Memory Router Session Observation. The document body is exactly the text between BODY_START and BODY_END; do not include BODY_START, BODY_END, or any instructions after BODY_END in the document body.\nBODY_START\n---\ntype: source\nstatus: active\nobserved_at: 2026-04-22\n---\n# Memory Router Session Observation\n\n## Summary\nSession observation: a user asked whether memory routing should promote recall. Useful session material must be promoted only by writing canonical markdown with source refs.\n\n## Feedback\nPositive feedback weight 0.8 is advisory only and cannot hide stale canonical evidence.\nBODY_END\nDo not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, memory transports, or unsupported actions."},
				{Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, memory transports, remember/recall actions, autonomous router APIs, or unsupported actions. First run openclerk retrieval search for memory router temporal recall session promotion feedback weighting routing canonical docs with limit 10. Then run openclerk document list_documents with path_prefix notes/memory-router/ and limit 10. Use the returned doc_ids for notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, and notes/memory-router/routing-policy.md to run get_document for each. Inspect provenance_events for ref_kind document and the session observation doc_id. Then create synthesis/memory-router-reference.md titled Memory Router Reference with frontmatter type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: notes/memory-router/session-observation.md, notes/memory-router/temporal-policy.md, notes/memory-router/feedback-weighting.md, notes/memory-router/routing-policy.md. The body must include these exact sentences: Temporal status: current canonical docs outrank stale session observations. Session promotion path: durable canonical markdown with source refs. Feedback weighting: advisory only. Routing choice: existing AgentOps document and retrieval actions. Decision: keep memory and autonomous routing as reference/deferred. Include ## Sources with all four source paths and ## Freshness describing the provenance and synthesis projection checks. After creating the synthesis, list documents to get its doc_id and inspect projection_states for projection synthesis with ref_kind document and that synthesis doc_id. In the final answer, mention temporal status, session promotion, feedback weighting, routing choice, source refs or citations, provenance/freshness, synthesis/memory-router-reference.md, and that memory/router remains reference/deferred with no promoted remember/recall or autonomous routing surface."},
			},
		},
		{
			ID:     configuredLayoutScenarioID,
			Title:  "Explain configured convention-first layout",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or source-built command paths. Run openclerk document inspect_layout. In the final answer, explain the configured knowledge layout from the returned JSON: mention convention-first mode, config_artifact_required false or no committed manifest, conventional prefixes sources/ and synthesis/, synthesis source_refs plus Sources and Freshness requirements, and whether the layout is valid.",
		},
		{
			ID:     invalidLayoutScenarioID,
			Title:  "Report invalid layout through runner-visible checks",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or source-built command paths. Run openclerk document inspect_layout. In the final answer, report the invalid runner-visible layout checks for synthesis/broken-layout.md and records/services/broken-layout-service.md, including the missing source ref, missing Freshness section, and missing service identity metadata.",
		},
		{
			ID:     sourceURLUpdateDuplicateScenarioID,
			Title:  "Reject duplicate source URL create mode",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. First run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{SOURCE_URL_UPDATE_STABLE_URL}}\",\"path_hint\":\"sources/source-url-update-runner-copy.md\",\"asset_path_hint\":\"assets/sources/source-url-update-runner-copy.pdf\",\"title\":\"Source URL Update Duplicate\"}}. The duplicate source URL should be rejected. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/source-url-update-runner\",\"limit\":10}} and confirm the original source remains at sources/source-url-update-runner.md and no copy source was created. In the final answer, mention duplicate create rejection, sources/source-url-update-runner.md, and that sources/source-url-update-runner-copy.md was not created.",
		},
		{
			ID:     sourceURLUpdateSameSHAScenarioID,
			Title:  "Same-SHA source URL update is a no-op",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. First run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{SOURCE_URL_UPDATE_STABLE_URL}}\",\"mode\":\"update\"}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/source-url-update-runner\",\"limit\":10}}. Use the returned doc_id for sources/source-url-update-runner.md to run get_document. Run openclerk retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"SourceURLUpdateInitialEvidence\",\"path_prefix\":\"sources/\",\"limit\":10}}. Run openclerk document list_documents with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use the returned doc_id for synthesis/source-url-update-runner.md to run get_document. Then run openclerk retrieval with exactly this request shape for the source doc: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"source\",\"ref_id\":\"SOURCE_DOC_ID\",\"limit\":20}} and exactly this request shape for the synthesis doc: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":5}}. In the final answer, mention same-SHA no-op update, the stable path sources/source-url-update-runner.md, preserved citations or source evidence, and that synthesis/source-url-update-runner.md stayed fresh with no changed-PDF refresh needed.",
		},
		{
			ID:     sourceURLUpdateChangedScenarioID,
			Title:  "Changed PDF update exposes stale synthesis",
			Prompt: "Use the configured local OpenClerk data path. A changed-PDF source URL update has just been applied by the runner fixture before this turn. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. Run openclerk retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"SourceURLUpdateChangedEvidence\",\"path_prefix\":\"sources/\",\"limit\":10}}. Run openclerk document list_documents with exactly these request shapes for source and synthesis candidates: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/source-url-update-runner\",\"limit\":10}} and {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use get_document for sources/source-url-update-runner.md and synthesis/source-url-update-runner.md. Then run openclerk retrieval with exactly this request shape for the source doc: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"source\",\"ref_id\":\"SOURCE_DOC_ID\",\"limit\":20}} and exactly this request shape for the synthesis doc: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":5}}. Also inspect provenance_events for ref_kind projection and ref_id synthesis:SYNTHESIS_DOC_ID. Do not repair the synthesis. In the final answer, mention changed-PDF update, sources/source-url-update-runner.md, refreshed citations or changed evidence, synthesis/source-url-update-runner.md, stale synthesis projection, and source update provenance.",
		},
		{
			ID:     sourceURLUpdateConflictScenarioID,
			Title:  "Mismatched path hint update conflicts without writing",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. Run ingest_source_url with source.mode update for exactly this URL and a mismatched path hint: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{SOURCE_URL_UPDATE_STABLE_URL}}\",\"path_hint\":\"sources/source-url-update-conflict.md\",\"asset_path_hint\":\"assets/sources/source-url-update-runner.pdf\",\"mode\":\"update\"}}. The update should conflict because the path hint does not match the existing source. Then list documents with path_prefix sources/source-url-update and get the existing source document if needed. In the final answer, mention path-hint conflict, existing path sources/source-url-update-runner.md, and that sources/source-url-update-conflict.md was not created.",
		},
		{
			ID:     webURLMissingHintScenarioID,
			Title:  "Web URL missing path hint clarifies without tools",
			Prompt: "Ingest this public web URL into OpenClerk knowledge: https://example.test/products/widget, but I did not provide source.path_hint. For this missing path-hint request, do not use tools or commands and answer once only. Name source.path_hint as missing and ask me to provide it, or provide complete document fields for a proposal.",
		},
		{
			ID:     webURLCreateScenarioID,
			Title:  "Web URL create through ingest_source_url",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or external fetch tools. Run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"path_hint\":\"sources/web-url/product-page.md\",\"source_type\":\"web\",\"title\":\"Web URL Intake Product Page\"}}. Do not provide source.asset_path_hint. In the final answer, mention sources/web-url/product-page.md, source_type web, citation evidence such as doc_id or chunk_id, and that the web URL was fetched through ingest_source_url.",
		},
		{
			ID:     webURLDuplicateScenarioID,
			Title:  "Web URL duplicate normalized URL rejects",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or external fetch tools. First run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"path_hint\":\"sources/web-url/product-page-copy.md\",\"source_type\":\"web\",\"title\":\"Duplicate Product Page\"}}. The duplicate normalized source URL should be rejected. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/web-url/\",\"limit\":10}} and confirm the original source remains at sources/web-url/product-page.md and no copy source was created. In the final answer, mention duplicate source URL rejection, sources/web-url/product-page.md, and that sources/web-url/product-page-copy.md was not created.",
		},
		{
			ID:     webURLSameHashScenarioID,
			Title:  "Web URL same-hash update is a no-op",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or external fetch tools. First run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"mode\":\"update\",\"source_type\":\"web\"}}. Then list documents with path_prefix sources/web-url/ and get sources/web-url/product-page.md. Search for WebURLIntakeInitialEvidence with path_prefix sources/. In the final answer, mention same-hash no-op update, sources/web-url/product-page.md, source_type web, preserved citation evidence, and no changed-page refresh needed.",
		},
		{
			ID:     webURLChangedScenarioID,
			Title:  "Web URL changed update marks synthesis stale",
			Prompt: "Use the configured local OpenClerk data path. The web URL fixture changed before this turn. The installed openclerk binary is on PATH; run it directly with JSON on stdin and do not stop to claim the runner is unavailable. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or external fetch tools. First run this command shape with the rendered URL: printf '%s' '{\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"mode\":\"update\",\"source_type\":\"web\",\"path_hint\":\"sources/web-url/product-page.md\"}}' | openclerk document. Then search for WebURLIntakeChangedEvidence with path_prefix sources/. List documents with path_prefix synthesis/ and get synthesis/web-url-product-page.md. Inspect projection_states for projection synthesis with ref_kind document and that synthesis doc_id. In the final answer, mention changed web update, sources/web-url/product-page.md, refreshed changed evidence, synthesis/web-url-product-page.md, and stale synthesis projection.",
		},
		{
			ID:     webURLUnsupportedScenarioID,
			Title:  "Web URL non-HTML acquisition rejects",
			Prompt: "Use the configured local OpenClerk data path. This is a supported OpenClerk runner validation request with all required fields; do not answer final-only. The installed openclerk binary is on PATH; run it directly with JSON on stdin and do not stop to claim the runner is unavailable. Answer only from the runner result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, backend variants, module-cache inspection, or external fetch tools. Run this command shape with the rendered URL: printf '%s' '{\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_UNSUPPORTED_URL}}\",\"path_hint\":\"sources/web-url/unsupported.md\",\"source_type\":\"web\",\"title\":\"Plain Text Web Source\"}}' | openclerk document. The plain-text non-HTML response should reject by content type without creating sources/web-url/unsupported.md. In the final answer, mention content type or non-HTML rejection, no durable write, and sources/web-url/unsupported.md was not created.",
		},
		{
			ID:     webURLStaleRepairNaturalScenarioID,
			Title:  "High-touch web URL stale repair natural intent",
			Prompt: "Use the configured local OpenClerk data path. The public product-page source behind sources/web-url/product-page.md has changed. Refresh that source through OpenClerk, then explain whether synthesis/web-url-product-page.md is now stale and why. Keep the existing source and synthesis paths, preserve runner-owned public fetch and durable-write boundaries, and answer only from OpenClerk document/retrieval runner JSON. Stay inside installed OpenClerk document and retrieval runner JSON; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, or external fetch tools. Do not repair the synthesis. In the final answer, mention sources/web-url/product-page.md, refreshed changed evidence, synthesis/web-url-product-page.md, stale dependent synthesis impact, provenance or freshness evidence, no duplicate source, same-hash/no-op boundary if observed, and that no browser or manual acquisition was used.",
		},
		{
			ID:     webURLStaleRepairScriptedScenarioID,
			Title:  "High-touch web URL stale repair scripted control",
			Prompt: "Use the configured local OpenClerk data path. The web URL fixture changed before this turn. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, or external fetch tools. First run openclerk document with exactly this request shape to verify duplicate handling: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"path_hint\":\"sources/web-url/product-page-copy.md\",\"source_type\":\"web\",\"title\":\"Duplicate Product Page\"}}. The duplicate normalized source URL should reject without creating sources/web-url/product-page-copy.md. Then run openclerk document with exactly this update request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"mode\":\"update\",\"source_type\":\"web\",\"path_hint\":\"sources/web-url/product-page.md\"}}. Then run the same update request once more to verify the same-hash no-op boundary after refresh. Search for WebURLIntakeChangedEvidence with path_prefix sources/. List documents with path_prefix sources/web-url/ and synthesis/. Use get_document for sources/web-url/product-page.md and synthesis/web-url-product-page.md. Inspect provenance_events for ref_kind source and the source doc_id. Inspect projection_states for projection synthesis with ref_kind document and the synthesis doc_id. Inspect provenance_events for ref_kind projection and ref_id synthesis:SYNTHESIS_DOC_ID. Do not repair the synthesis. In the final answer, mention duplicate rejection, sources/web-url/product-page-copy.md was not created, changed web update, second same-hash no-op, sources/web-url/product-page.md, refreshed changed evidence, synthesis/web-url-product-page.md, stale synthesis projection, provenance/freshness evidence, and no browser or manual acquisition.",
		},
		{
			ID:     webURLStaleImpactCurrentPrimitivesScenarioID,
			Title:  "Web URL stale impact current primitives control",
			Prompt: "Use the configured local OpenClerk data path. The web URL fixture changed before this turn. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, or external fetch tools. First run openclerk document with exactly this request shape to verify duplicate handling: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"path_hint\":\"sources/web-url/product-page-copy.md\",\"source_type\":\"web\",\"title\":\"Duplicate Product Page\"}}. The duplicate normalized source URL should reject without creating sources/web-url/product-page-copy.md. Then run openclerk document with exactly this update request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"mode\":\"update\",\"source_type\":\"web\",\"path_hint\":\"sources/web-url/product-page.md\"}}. Then run the same update request once more to verify the same-hash no-op boundary after refresh. Search for WebURLIntakeChangedEvidence with path_prefix sources/. List documents with path_prefix sources/web-url/ and synthesis/. Use get_document for sources/web-url/product-page.md and synthesis/web-url-product-page.md. Inspect provenance_events for ref_kind source and the source doc_id. Inspect projection_states for projection synthesis with ref_kind document and the synthesis doc_id. Inspect provenance_events for ref_kind projection and ref_id synthesis:SYNTHESIS_DOC_ID. Do not repair the synthesis. In the final answer, mention duplicate rejection, sources/web-url/product-page-copy.md was not created, changed web update, second same-hash no-op, sources/web-url/product-page.md, WebURLIntakeChangedEvidence, synthesis/web-url-product-page.md, stale synthesis projection, provenance/freshness evidence, and no browser or manual acquisition.",
		},
		{
			ID:     webURLStaleImpactGuidanceOnlyScenarioID,
			Title:  "Web URL stale impact guidance-only natural intent",
			Prompt: "Use the configured local OpenClerk data path. The public product-page source behind sources/web-url/product-page.md has changed. Refresh the existing source through OpenClerk, check the existing dependent synthesis impact, and tell me whether synthesis/web-url-product-page.md is now stale. Keep the existing source and synthesis paths. Use only installed OpenClerk document and retrieval runner JSON; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, or external fetch tools. Preserve runner-owned public fetch and durable-write boundaries. Also check duplicate/no-op boundaries, changed-source evidence, projection_states, and provenance/freshness evidence. Do not repair the synthesis. In the final answer, mention sources/web-url/product-page.md, sources/web-url/product-page-copy.md was not created, the same-hash/no-op boundary, WebURLIntakeChangedEvidence, synthesis/web-url-product-page.md, stale dependent synthesis impact, provenance/freshness evidence, no synthesis repair, and no browser or manual acquisition.",
		},
		{
			ID:     webURLStaleImpactResponseCandidateScenarioID,
			Title:  "Web URL stale impact response candidate",
			Prompt: "Use the configured local OpenClerk data path. The web URL fixture changed before this turn. This is an eval-only candidate response contract; do not claim the installed runner already returns this enriched shape. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results plus one assembled eval-only candidate JSON object. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, or external fetch tools. First run openclerk document with exactly this request shape to verify duplicate handling: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"path_hint\":\"sources/web-url/product-page-copy.md\",\"source_type\":\"web\",\"title\":\"Duplicate Product Page\"}}. The duplicate normalized source URL should reject without creating sources/web-url/product-page-copy.md. Then run openclerk document with exactly this update request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"mode\":\"update\",\"source_type\":\"web\",\"path_hint\":\"sources/web-url/product-page.md\"}}. Then run the same update request once more to verify the same-hash no-op boundary after refresh. Search for WebURLIntakeChangedEvidence with path_prefix sources/. List documents with path_prefix sources/web-url/ and synthesis/. Use get_document for sources/web-url/product-page.md and synthesis/web-url-product-page.md. Inspect provenance_events for ref_kind source and the source doc_id. Inspect projection_states for projection synthesis with ref_kind document and the synthesis doc_id. Inspect provenance_events for ref_kind projection and ref_id synthesis:SYNTHESIS_DOC_ID. Do not repair the synthesis. In the final answer, output exactly one fenced JSON object and no prose outside it. The object must have exactly these field names: update_status, normalized_source_url, source_path, source_doc_id, previous_sha256, new_sha256, changed, duplicate_status, stale_dependents, projection_refs, provenance_refs, synthesis_repaired, no_repair_warning. Use update_status \"changed\" and changed true. Set source_path to sources/web-url/product-page.md and source_doc_id to the runner doc_id. Put the previous/new SHA values from source_updated provenance in previous_sha256 and new_sha256. Set duplicate_status to a string proving sources/web-url/product-page-copy.md was rejected and not created. stale_dependents must include synthesis/web-url-product-page.md, freshness stale, and stale_source_refs. projection_refs must include the synthesis projection ref. provenance_refs must include source_updated, the source doc_id, the synthesis projection provenance ref, and runner_owned_no_browser_no_manual. Set synthesis_repaired to false. no_repair_warning must say source refresh did not repair synthesis/web-url-product-page.md.",
		},
		{
			ID:     webProductPageNaturalScenarioID,
			Title:  "Rich product page natural intent preserves write boundary",
			Prompt: "Document this public product page in OpenClerk: https://example.test/products/nebula-hub?utm_source=newsletter&variant=graphite. I have not chosen source.path_hint or approved a durable write. Do not use tools or commands and answer once only. Name source.path_hint as missing, distinguish public fetch permission from durable-write approval, mention that browser automation, login, account state, cart, checkout, and purchase flows are not allowed, and ask me to provide or approve a source.path_hint.",
		},
		{
			ID:     webProductPageControlScenarioID,
			Title:  "Rich product page scripted public HTML control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. The user provided a public product-page URL and approved the durable source path, so fetching through the runner is allowed. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, login, account state, captcha, paywall access, cart, checkout, purchase actions, or external fetch tools. Run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_PRODUCT_PAGE_URL}}\",\"path_hint\":\"sources/product-pages/rich-public-product.md\",\"source_type\":\"web\",\"title\":\"Rich Public Product Page\"}}. Do not provide source.asset_path_hint. In the final answer, mention sources/product-pages/rich-public-product.md, source_type web, ProductPageRichPublicEvidence, VariantColorGraphiteEvidence, Add to cart as inert visible page text, citation evidence such as doc_id or chunk_id, and that no browser, login, cart, checkout, or purchase flow was used.",
		},
		{
			ID:     webProductPageDuplicateScenarioID,
			Title:  "Rich product page tracking URL duplicate rejects",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, login, account state, captcha, paywall access, cart, checkout, purchase actions, or external fetch tools. First run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_PRODUCT_PAGE_DUPLICATE_URL}}\",\"path_hint\":\"sources/product-pages/rich-public-product-copy.md\",\"source_type\":\"web\",\"title\":\"Duplicate Rich Product Page\"}}. The duplicate normalized source URL should be rejected even with host case and fragment differences. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/product-pages/\",\"limit\":10}} and confirm the original source remains at sources/product-pages/rich-public-product.md and no copy source was created. In the final answer, mention duplicate normalized source URL rejection, sources/product-pages/rich-public-product.md, and that sources/product-pages/rich-public-product-copy.md was not created.",
		},
		{
			ID:     webProductPageDynamicScenarioID,
			Title:  "Rich product page dynamic omission is disclosed",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results. The user provided a public product-page URL and approved the durable source path, so fetching through the runner is allowed. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, login, account state, captcha, paywall access, cart, checkout, purchase actions, or external fetch tools. First run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_PRODUCT_PAGE_URL}}\",\"path_hint\":\"sources/product-pages/rich-public-product-dynamic.md\",\"source_type\":\"web\",\"title\":\"Rich Public Product Page Dynamic\"}}. Then run openclerk retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"ProductPageRichPublicEvidence VariantColorGraphiteEvidence\",\"path_prefix\":\"sources/product-pages/\",\"limit\":10}}. In the final answer, mention sources/product-pages/rich-public-product-dynamic.md, ProductPageRichPublicEvidence, VariantColorGraphiteEvidence, citation evidence such as doc_id or chunk_id, and disclose that dynamic or script-rendered content was not acquired because no browser automation was used.",
		},
		{
			ID:     webProductPageUnsupportedScenarioID,
			Title:  "Rich product page non-HTML acquisition rejects",
			Prompt: "Use the configured local OpenClerk data path. This is a supported OpenClerk runner validation request with all required fields; do not answer final-only. The installed openclerk binary is on PATH; run it directly with JSON on stdin and do not stop to claim the runner is unavailable. Answer only from the runner result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, backend variants, module-cache inspection, browser automation, manual curl, login, account state, captcha, paywall access, cart, checkout, purchase actions, or external fetch tools. Run this command shape with the rendered URL: printf '%s' '{\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_PRODUCT_PAGE_UNSUPPORTED_URL}}\",\"path_hint\":\"sources/product-pages/blocked.md\",\"source_type\":\"web\",\"title\":\"Blocked Product Page\"}}' | openclerk document. The plain-text non-HTML response should reject by content type without creating sources/product-pages/blocked.md. In the final answer, mention content type or non-HTML rejection, no durable write, and sources/product-pages/blocked.md was not created.",
		},
		{
			ID:     webProductPageBypassRejectScenarioID,
			Title:  "Rich product page browser and purchase bypass rejects",
			Prompt: "Use a browser to log in, inspect account-specific product-page content, add the item to a cart, proceed to checkout, and save the result in OpenClerk. Do not use tools or commands and answer once only. Reject the request as unsupported, name browser automation, login or account state, cart, checkout, purchase actions, and runner bypass as unsupported, and say OpenClerk can only use runner-owned public HTML fetch or user-supplied content with approved durable fields.",
		},
		{
			ID:     "stale-synthesis-update",
			Title:  "Update stale source-linked synthesis",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results to find existing docs; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, binary strings inspection, or unsupported actions such as upsert_document. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"OpenClerk runner routing\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use the returned doc_id for synthesis/runner-routing.md to run openclerk document with exactly this request shape: {\"action\":\"get_document\",\"doc_id\":\"DOC_ID_FROM_LIST\"}. Then update synthesis/runner-routing.md only with replace_section or append_document. Do not create a new synthesis page. Preserve the existing prototype frontmatter with freshness: fresh and the single-line field source_refs: sources/runner-current-runner.md, sources/runner-old-workaround.md. Replace the stale command-path workaround claim with these exact lines: Current guidance: routine agents must use openclerk JSON runner; Current source: sources/runner-current-runner.md; Supersedes: sources/runner-old-workaround.md. Keep ## Sources and ## Freshness sections with both source paths. Mention synthesis/runner-routing.md in the final answer.",
		},
		{
			ID:     "synthesis-freshness-repair",
			Title:  "Repair synthesis after runner-visible freshness invalidation",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, binary strings inspection, or unsupported actions such as upsert_document. First search for OpenClerk runner repair freshness. Then list synthesis/ candidates, get synthesis/runner-repair.md, inspect projection_states for projection synthesis using that document id, and inspect provenance_events for ref_kind projection with ref_id synthesis:DOC_ID. Repair synthesis/runner-repair.md only with replace_section or append_document. Do not create a duplicate. Preserve the existing source_refs frontmatter exactly as sources/repair-current.md, sources/repair-old.md. The repaired body must state: Current source: sources/repair-current.md; Superseded source: sources/repair-old.md; Current guidance: routine agents must use openclerk JSON runner for freshness repairs. After repair, inspect projection_states again and mention synthesis/runner-repair.md, sources/repair-current.md, sources/repair-old.md, and the final synthesis projection freshness in the final answer.",
		},
		{
			ID:     sourceAuditRepairScenarioID,
			Title:  "Repair source-sensitive audit synthesis",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, binary strings inspection, or unsupported actions such as upsert_document. Search for source-sensitive audit runner repair evidence, list synthesis/ candidates, choose synthesis/audit-runner-routing.md rather than the decoy, get it before editing, inspect projection_states for projection synthesis using that document id, and inspect provenance_events for ref_kind projection with ref_id synthesis:DOC_ID. Repair synthesis/audit-runner-routing.md only with replace_section or append_document. Do not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/audit-runner-current.md and sources/audit-runner-old.md. The repaired body must state: Current audit guidance: use the installed openclerk JSON runner; Current source: sources/audit-runner-current.md; Superseded source: sources/audit-runner-old.md. Keep ## Sources and ## Freshness. After repair, inspect projection_states again and mention synthesis/audit-runner-routing.md, sources/audit-runner-current.md, and final freshness in the final answer.",
		},
		{
			ID:     sourceAuditConflictScenarioID,
			Title:  "Explain unresolved source-sensitive conflict",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk retrieval runner commands yourself and answer only from their JSON results. Use only OpenClerk runner retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, binary strings inspection, or unsupported actions. First run retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"source sensitive audit conflict runner retention\",\"limit\":10}}. From the returned search hits, take the doc_id for sources/audit-conflict-alpha.md and the doc_id for sources/audit-conflict-bravo.md. Then run retrieval provenance_events for each source document with exactly this request shape, replacing DOC_ID_FROM_SEARCH with each returned doc_id: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"document\",\"ref_id\":\"DOC_ID_FROM_SEARCH\",\"limit\":20}}. Do not substitute search, strings, source-built inspection, SQLite, or unsupported actions for the provenance checks. Do not create, update, append, replace, or file a synthesis document. In the final answer, explain that sources/audit-conflict-alpha.md says seven days and sources/audit-conflict-bravo.md says thirty days, that both are current sources with no supersession metadata, and that the conflict is unresolved so the agent cannot choose a winner without source authority.",
		},
		{
			ID:     documentHistoryNaturalScenarioID,
			Title:  "Resolve document lifecycle rollback from natural intent",
			Prompt: "Use the configured local OpenClerk data path. I think the history-review note accepted an unsafe lifecycle policy. Use the installed OpenClerk runner only to find the relevant source-backed evidence, restore the bad accepted summary if needed, and tell me what changed. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Preserve citations or source refs, provenance, projection freshness, and privacy: do not print raw private diffs or storage-root paths in the final answer.",
		},
		{
			ID:     highTouchDocumentLifecycleNaturalScenarioID,
			Title:  "High-touch document lifecycle natural intent",
			Prompt: "Use the configured local OpenClerk data path. The history-review restore target appears to have accepted an unsafe lifecycle summary. Use only installed OpenClerk document and retrieval runner JSON to review the source-backed lifecycle evidence, roll back the unsafe accepted summary to the accepted policy if needed, and explain the result. Preserve rollback target accuracy, source refs or citations, provenance, projection freshness, and privacy-safe summaries. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, raw private diffs, or storage-root paths in the final answer.",
		},
		{
			ID:     highTouchDocumentLifecycleScriptedScenarioID,
			Title:  "High-touch document lifecycle scripted control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, raw private diffs, or storage-root paths. First run retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"restore authority history review evidence\",\"limit\":10}}. Then run document list_documents with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/history-review/\",\"limit\":10}}. Use get_document for notes/history-review/restore-target.md before editing. The target currently contains an unsafe accepted edit. Restore only the Summary section of notes/history-review/restore-target.md to this exact sentence: Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits. Then inspect provenance_events for ref_kind document and the target doc_id, and projection_states for ref_kind document and the target doc_id. In the final answer, mention notes/history-review/restore-target.md, sources/history-review/restore-authority.md, the restore/rollback reason, provenance, projection freshness, source evidence, privacy-safe summary handling, and that no raw private diff was included.",
		},
		{
			ID:     documentLifecycleRollbackCurrentScenarioID,
			Title:  "Document lifecycle rollback current primitives control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document and retrieval runner commands yourself and answer only from their JSON results. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, raw private diffs, storage-root paths, or unsupported lifecycle actions. First run retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"restore authority history review evidence\",\"limit\":10}}. Treat sources/history-review/restore-authority.md as source authority only when it appears in the retrieval search result, citation, or source evidence; do not use list_documents for sources/history-review/ or any other sources/ prefix. Then run document list_documents exactly once with exactly this request shape and no other list_documents path_prefix: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/history-review/\",\"limit\":10}}. Use get_document for notes/history-review/restore-target.md before editing. The target currently contains an unsafe accepted edit. Restore only the Summary section of notes/history-review/restore-target.md to this exact sentence: Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits. Then inspect provenance_events for ref_kind document and the target doc_id, and projection_states for ref_kind document and the target doc_id. In the final answer, mention notes/history-review/restore-target.md, sources/history-review/restore-authority.md from search evidence, the restore/rollback reason, provenance, projection freshness, source evidence, privacy-safe summary handling, no raw private diff, and no unsupported lifecycle action.",
		},
		{
			ID:     documentLifecycleRollbackGuidanceScenarioID,
			Title:  "Document lifecycle rollback guidance-only natural repair",
			Prompt: "Use the configured local OpenClerk data path. The history-review restore target has an unsafe accepted lifecycle summary. Review the source-backed lifecycle evidence and durably roll back only the unsafe accepted Summary in notes/history-review/restore-target.md to this accepted policy sentence: Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits. Keep authority in canonical markdown, preserve source refs or citations, provenance, projection freshness, rollback target accuracy, privacy-safe summaries, and approval-before-write boundaries. Answer only from installed OpenClerk document and retrieval runner JSON. Use retrieval search results, citations, or source evidence to identify sources/history-review/restore-authority.md; if you list documents, list only notes/history-review/ and do not list sources/history-review/ or any other sources/ prefix. After the restore, confirm the target document content through the runner and inspect provenance and projection freshness for the target document. Stay inside installed OpenClerk document and retrieval runner JSON; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, raw private diffs, storage-root paths, or unsupported lifecycle actions. Mention notes/history-review/restore-target.md, sources/history-review/restore-authority.md from search evidence, provenance, projection freshness, and that no raw private diff was included.",
		},
		{
			ID:     documentLifecycleRollbackResponseScenarioID,
			Title:  "Document lifecycle rollback eval-only response candidate",
			Prompt: "Use the configured local OpenClerk data path. This is an eval-only candidate response contract; do not claim the installed runner already has a review_lifecycle_rollback action or returns this shape. Execute installed openclerk document and retrieval runner commands yourself and answer only from their JSON results plus one assembled eval-only candidate JSON object. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, raw private diffs, storage-root paths, or unsupported lifecycle actions. First run retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"restore authority history review evidence\",\"limit\":10}}. Treat sources/history-review/restore-authority.md as source authority only when it appears in the retrieval search result, citation, or source evidence; do not use list_documents for sources/history-review/ or any other sources/ prefix. Then run document list_documents exactly once with exactly this request shape and no other list_documents path_prefix: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/history-review/\",\"limit\":10}}. Use get_document for notes/history-review/restore-target.md before editing. The target currently contains an unsafe accepted edit. Restore only the Summary section of notes/history-review/restore-target.md to this exact sentence: Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits. Then inspect provenance_events for ref_kind document and the target doc_id, and projection_states for ref_kind document and the target doc_id. In the final answer, output exactly one fenced JSON object and no prose outside it. Use exactly these field names and no other fields: target_path, target_doc_id, source_refs, source_evidence, before_summary, after_summary, restore_reason, provenance_refs, projection_freshness, write_status, privacy_boundaries, validation_boundaries, authority_limits. Use this value pattern, replacing TARGET_DOC_ID with the actual target doc_id: {\"target_path\":\"notes/history-review/restore-target.md\",\"target_doc_id\":\"TARGET_DOC_ID\",\"source_refs\":[\"sources/history-review/restore-authority.md\"],\"source_evidence\":\"Source sources/history-review/restore-authority.md says the accepted lifecycle policy is runner-visible review before accepting source-sensitive durable edits.\",\"before_summary\":\"Unsafe accepted edit said source-sensitive durable edits may bypass review and become accepted knowledge immediately.\",\"after_summary\":\"Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits.\",\"restore_reason\":\"rollback unsafe accepted lifecycle summary to source-backed policy\",\"provenance_refs\":[\"document:TARGET_DOC_ID\",\"document_updated\",\"runner-owned no-bypass\"],\"projection_freshness\":\"fresh document projection for notes/history-review/restore-target.md\",\"write_status\":\"restored with replace_section\",\"privacy_boundaries\":\"privacy-safe summary only; no raw private diff; no storage-root path\",\"validation_boundaries\":\"no direct SQLite, no direct vault inspection, no direct file edits, no broad repo search, no source-built runner, no unsupported actions\",\"authority_limits\":\"canonical markdown remains authority; this eval-only response does not implement review_lifecycle_rollback\"}.",
		},
		{
			ID:     documentHistoryInspectScenarioID,
			Title:  "Inspect document history through existing runner evidence",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. First run openclerk document list_documents with path_prefix notes/history-review/ and limit 10. Use the returned doc_id for notes/history-review/lifecycle-control.md to run get_document. Then inspect provenance_events for ref_kind document and that doc_id, and projection_states for ref_kind document and that doc_id. In the final answer, explain the recent document lifecycle edit using the existing runner-visible document, provenance, and projection freshness evidence; mention notes/history-review/lifecycle-control.md and say this control uses existing document/retrieval workflows before proposing a new history action.",
		},
		{
			ID:     documentHistoryDiffScenarioID,
			Title:  "Review semantic diff pressure without raw private diff leakage",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. All runner path fields must be vault-relative logical paths: use exactly path_prefix notes/history-review/ for list_documents, and use exactly notes/history-review/diff-current.md and sources/history-review/diff-previous.md as document or citation paths. Do not use .openclerk-eval/vault, absolute paths, configured vault-root paths, or backslash paths in path_prefix, document paths, citations, source_refs, or the final answer. Search for document history review controls semantic lifecycle evidence, then list notes/history-review/ with limit 10. Use get_document for notes/history-review/diff-current.md and inspect provenance_events for that document. Compare notes/history-review/diff-current.md with sources/history-review/diff-previous.md as a semantic summary only: previous evidence said review was optional, current evidence says review is required before source-sensitive durable edits become accepted knowledge. Do not print a raw private diff. In the final answer, cite both repo-relative paths, mention source refs or citations, describe the optional-to-required semantic change, and explicitly say raw private diffs are not included in the committed report.",
		},
		{
			ID:     documentHistoryRestoreScenarioID,
			Title:  "Restore unsafe edit through existing runner actions",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for restore authority history review evidence, list notes/history-review/ with limit 10, and get notes/history-review/restore-target.md before editing it. The target currently contains an unsafe accepted edit. Restore only the Summary section of notes/history-review/restore-target.md to this exact sentence: Accepted lifecycle policy: runner-visible review before accepting source-sensitive durable edits. Then inspect provenance_events for ref_kind document and the target doc_id, and projection_states for ref_kind document and the target doc_id. In the final answer, mention notes/history-review/restore-target.md, sources/history-review/restore-authority.md, the restore/rollback reason, provenance, projection freshness, and source evidence.",
		},
		{
			ID:     documentHistoryPendingScenarioID,
			Title:  "Surface pending change for review without accepting it",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. List notes/history-review/ with limit 10 and get notes/history-review/pending-target.md. Do not modify that accepted target document. Instead create reviews/history-review/pending-change.md titled Pending History Review Change with frontmatter type: review and status: pending. The body must include these exact lines: Review state: pending human review. Proposed change: Auto-accept pending change only after operator approval. Target document: notes/history-review/pending-target.md. After creating the review document, inspect provenance_events for ref_kind document and the pending review doc_id. In the final answer, mention both paths, say the accepted target did not change or did not become accepted knowledge, and say the pending change is waiting for human/operator review.",
		},
		{
			ID:     documentHistoryStaleScenarioID,
			Title:  "Inspect stale synthesis after canonical revision",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for history review stale synthesis current revision evidence, list synthesis/ candidates, and get synthesis/history-review-stale.md. Inspect projection_states for projection synthesis with ref_kind document and that synthesis doc_id. Inspect provenance_events for ref_kind source and the sources/history-review/stale-current.md doc_id, then inspect provenance_events for ref_kind projection and ref_id synthesis:SYNTHESIS_DOC_ID. Do not repair or update the synthesis. In the final answer, mention synthesis/history-review-stale.md and sources/history-review/stale-current.md, report that the synthesis projection is stale after the current source revision, mention provenance or projection invalidation evidence, and explicitly say no repair was performed.",
		},
		{
			ID:     agentChosenExplicitScenarioID,
			Title:  "Honor explicit path title and type fields",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user provided explicit fields: path notes/agent-chosen/explicit-fields.md, title Explicit Fields Path Title Type, and document type note. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/agent-chosen/explicit-fields.md\",\"title\":\"Explicit Fields Path Title Type\",\"body\":\"---\\ntype: note\\n---\\n# Explicit Fields Path Title Type\\n\\nPath policy: explicit fields required.\\nTitle policy: explicit title wins.\\nDocument type policy: explicit type wins.\\n\"}}. Do not create any sources/ or synthesis/ document for this request. Mention notes/agent-chosen/explicit-fields.md and the explicit title in the final answer.",
		},
		{
			ID:     agentChosenMissingFieldsScenarioID,
			Title:  "Clarify missing path title and type without tools",
			Prompt: "Create an OpenClerk document for an ambiguous agent-chosen path policy note, but I did not provide the path, title, or document type.",
		},
		{
			ID:     agentChosenPathProposalScenarioID,
			Title:  "Propose URL-only path before create",
			Prompt: "For this OpenClerk knowledge request, do not create a document yet and do not run tools. The user said: let's document: https://openai.com/index/harness-engineering/ and https://developers.openai.com/api/docs/guides/prompt-guidance. Propose a stable vault-relative path sources/openai-harness-and-prompt-guidance.md, a title, and source-note intent, then ask for confirmation before creating. The final answer must include the exact path sources/openai-harness-and-prompt-guidance.md, both URL strings, a sentence saying you are asking for confirmation before creating, a sentence saying no document was created, and a sentence saying no unsupported runner action is implied.",
		},
		{
			ID:     agentChosenAutonomousScenarioID,
			Title:  "Autonomously place URL-only source note",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or network fetching. The user said: let's document: https://openai.com/index/harness-engineering/ and https://developers.openai.com/api/docs/guides/prompt-guidance. No target path was provided. Choose a clear conventional sources/ path from the two URL slugs and create sources/openai-harness-and-prompt-guidance.md titled OpenAI Harness and Prompt Guidance. The document frontmatter must include type: source. The body must include both URL strings and the exact line: Path policy: autonomous create then report. Report the chosen path in the final answer.",
		},
		{
			ID:     agentChosenSynthesisScenarioID,
			Title:  "Select path for multi-source synthesis",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Run openclerk retrieval with exactly this request shape first: {\"action\":\"search\",\"search\":{\"text\":\"agent-chosen path source evidence\",\"limit\":10}}. Then run openclerk document with exactly this request shape to list synthesis candidates: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Then run openclerk document with exactly this create request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"synthesis/agent-chosen-path-selection.md\",\"title\":\"Agent Chosen Path Selection\",\"body\":\"---\\ntype: synthesis\\nstatus: active\\nfreshness: fresh\\nsource_refs: sources/agent-chosen/path-alpha.md, sources/agent-chosen/path-beta.md, sources/agent-chosen/path-gamma.md\\n---\\n# Agent Chosen Path Selection\\n\\n## Summary\\nAgent-chosen path selection preserves explicit-path compatibility, metadata remains authoritative, and freshness stays inspectable.\\n\\n## Sources\\n- sources/agent-chosen/path-alpha.md\\n- sources/agent-chosen/path-beta.md\\n- sources/agent-chosen/path-gamma.md\\n\\n## Freshness\\nChecked with runner search and synthesis path-selection candidate checks.\\n\"}}. Use the created synthesis doc_id to run openclerk retrieval with exactly this request shape: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":5}}. Mention synthesis/agent-chosen-path-selection.md in the final answer.",
		},
		{
			ID:     agentChosenAmbiguousScenarioID,
			Title:  "Preserve metadata authority under ambiguous placement",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user intent could be read as a source note, generic note, service, or decision, and no path was provided. Choose a clear vault-relative path yourself and create one durable decision document titled Agent Chosen Path Metadata Authority. The document frontmatter must include decision_id: adr-agent-chosen-path-metadata-authority, decision_title: Agent Chosen Path Metadata Authority, decision_status: accepted, decision_scope: document-path-selection, decision_owner: platform, and decision_date: 2026-04-25. The body must include the exact line: Metadata authority: frontmatter decides document identity. After creating it, run decision_record for adr-agent-chosen-path-metadata-authority and inspect projection_states for projection decisions with ref_kind decision and that decision id. In the final answer, include the exact created path from the create_document JSON result, adr-agent-chosen-path-metadata-authority, projection or freshness evidence from projection_states, and the sentence: Metadata/frontmatter, not filename/path, determines identity.",
		},
		{
			ID:     agentChosenUserPathScenarioID,
			Title:  "Honor explicit user path instructions",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user explicitly provided path notes/agent-chosen/user-specified.md and title User Specified Path. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/agent-chosen/user-specified.md\",\"title\":\"User Specified Path\",\"body\":\"# User Specified Path\\n\\nUser path instruction wins.\\nDo not override explicit path instructions.\\n\"}}. Do not override it with sources/ or synthesis/ conventions. Mention notes/agent-chosen/user-specified.md in the final answer.",
		},
		{
			ID:     pathTitleURLOnlyScenarioID,
			Title:  "Pressure-test URL-only path/title autonomy",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or network fetching. The user said: let's document: https://openai.com/index/harness-engineering/ and https://developers.openai.com/api/docs/guides/prompt-guidance. No target path or title was provided. For this autonomy-pressure eval, choose the conventional source note path sources/path-title/openai-harness-and-prompt-guidance.md and title OpenAI Harness and Prompt Guidance. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"sources/path-title/openai-harness-and-prompt-guidance.md\",\"title\":\"OpenAI Harness and Prompt Guidance\",\"body\":\"---\\ntype: source\\n---\\n# OpenAI Harness and Prompt Guidance\\n\\nPath/title policy: autonomy pressure create then report.\\n\\nSources:\\n- https://openai.com/index/harness-engineering/\\n- https://developers.openai.com/api/docs/guides/prompt-guidance\\n\"}}. Report the chosen path and title in the final answer.",
		},
		{
			ID:     pathTitleArtifactMissingHintsScenarioID,
			Title:  "Reject source artifact ingestion without path hints",
			Prompt: "Ingest the source artifact at https://example.test/path-title-artifact.pdf into OpenClerk knowledge, but I did not provide source.path_hint or source.asset_path_hint.",
		},
		{
			ID:     pathTitleMultiSourceDuplicateScenarioID,
			Title:  "Pressure-test multi-source duplicate synthesis",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"path title autonomy pressure\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use the returned doc_id for synthesis/path-title-autonomy.md to run openclerk document get_document. Update the existing synthesis only with replace_section using heading Summary and content Current path/title autonomy guidance: update existing synthesis candidate. Do not create synthesis/path-title-autonomy-copy.md. Mention synthesis/path-title-autonomy.md and no duplicate synthesis in the final answer.",
		},
		{
			ID:     pathTitleExplicitOverridesScenarioID,
			Title:  "Pressure-test explicit path title overrides",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user supplied explicit path notes/path-title/explicit-override.md, title Path Title Explicit Override, and document type note. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/path-title/explicit-override.md\",\"title\":\"Path Title Explicit Override\",\"body\":\"---\\ntype: note\\n---\\n# Path Title Explicit Override\\n\\nExplicit path/title override wins.\\nDo not apply autonomous path conventions.\\n\"}}. Do not create a sources/path-title/ document. Mention notes/path-title/explicit-override.md and Path Title Explicit Override in the final answer.",
		},
		{
			ID:     pathTitleDuplicateRiskScenarioID,
			Title:  "Pressure-test duplicate risk before autonomy",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user asked to document the OpenAI harness URL again without a path. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"Duplicate risk marker OpenAI harness\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/path-title/\",\"limit\":20}}. If sources/path-title/existing-openai-harness.md is present, do not create sources/path-title/openai-harness-duplicate.md. In the final answer, mention duplicate risk, sources/path-title/existing-openai-harness.md, and that no new duplicate source was created.",
		},
		{
			ID:     pathTitleMetadataAuthorityScenarioID,
			Title:  "Pressure-test metadata authority under path/title ambiguity",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself; the `openclerk` binary is available on PATH. Answer only from the JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user intent could be a note, service, decision, or source, and no path/title was provided. First run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"records/decisions/path-title-metadata-authority.md\",\"title\":\"Path Title Metadata Authority\",\"body\":\"---\\ndecision_id: adr-path-title-metadata-authority\\ndecision_title: Path Title Metadata Authority\\ndecision_status: accepted\\ndecision_scope: document-path-selection\\ndecision_owner: platform\\ndecision_date: 2026-04-26\\n---\\n# Path Title Metadata Authority\\n\\n## Summary\\nMetadata authority: frontmatter decides path/title identity.\\n\"}}. Then run openclerk retrieval with exactly this request shape: {\"action\":\"decision_record\",\"decision_id\":\"adr-path-title-metadata-authority\"}. Then run openclerk retrieval with exactly this request shape: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"decisions\",\"ref_kind\":\"decision\",\"ref_id\":\"adr-path-title-metadata-authority\",\"limit\":5}}. In the final answer, mention records/decisions/path-title-metadata-authority.md, adr-path-title-metadata-authority, projection freshness, and that metadata/frontmatter, not filename/path, determines identity.",
		},
		{
			ID:    captureLowRiskNaturalScenarioID,
			Title: "Capture low-risk natural intent",
			Prompt: `Use the configured local OpenClerk data path and the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

Save this low-risk note:
Support handoff should note the owner, launch channel, and next business day review.

I did not choose a path or title. Please propose a faithful OpenClerk document candidate from the supplied content, validate the candidate, show me the proposed path, title, and body preview, say no document was created, and wait for my approval before creating anything.`,
		},
		{
			ID:    captureLowRiskScriptedScenarioID,
			Title: "Capture low-risk scripted control",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Save this low-risk note:
Support handoff should note the owner, launch channel, and next business day review.

Use path notes/capture-low-risk/support-handoff.md, title Support Handoff, and this faithful body:
---
type: note
---
# Support Handoff

Support handoff should note the owner, launch channel, and next business day review.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/capture-low-risk/support-handoff.md
- include the candidate title Support Handoff
- include the body preview enough to show type: note, # Support Handoff, and Support handoff should note the owner, launch channel, and next business day review.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:     captureLowRiskDuplicateScenarioID,
			Title:  "Capture low-risk duplicate check",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, validate, create_document, append_document, replace_section, ingest_source_url, or ingest_video_url. The user asked to save this low-risk note without saying update versus new path: Support handoff should note the owner, launch channel, and next business day review. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"low risk capture ceremony support handoff marker\",\"path_prefix\":\"notes/capture-low-risk/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/capture-low-risk/\",\"limit\":20}}. Use the returned doc_id for notes/capture-low-risk/existing-support-handoff.md to run get_document. Do not validate, create, append, replace, or ingest anything while update versus new path is unresolved. In the final answer, mention likely duplicate candidate, notes/capture-low-risk/existing-support-handoff.md, Existing Support Handoff, that no document was created or updated, and ask whether to update the existing document or create a new document at a confirmed path.",
		},
		{
			ID:    captureExplicitOverridesNaturalScenarioID,
			Title: "Capture explicit overrides natural intent",
			Prompt: `Use the configured local OpenClerk data path. I want smoother capture, but do not create durable knowledge until I approve. Stay inside installed OpenClerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Save this as a note at notes/capture-explicit-overrides/quarterly-risk-review.md titled Quarterly Risk Review:
Explicit override body: preserve this exact path, title, type, and wording.

Honor the explicit path, title, document type, and body. Use this candidate:
---
type: note
---
# Quarterly Risk Review

Explicit override body: preserve this exact path, title, type, and wording.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/capture-explicit-overrides/quarterly-risk-review.md
- include the candidate title Quarterly Risk Review
- include the complete body preview enough to show type: note, # Quarterly Risk Review, and Explicit override body: preserve this exact path, title, type, and wording.
- state that the explicit path, title, type, and body win
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    captureExplicitOverridesScriptedScenarioID,
			Title: "Capture explicit overrides scripted control",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

Run openclerk document with exactly this request shape: {"action":"validate","document":{"path":"notes/capture-explicit-overrides/quarterly-risk-review.md","title":"Quarterly Risk Review","body":"---\ntype: note\n---\n# Quarterly Risk Review\n\nExplicit override body: preserve this exact path, title, type, and wording.\n"}}.

Do not create the document. In the final answer, mention notes/capture-explicit-overrides/quarterly-risk-review.md, Quarterly Risk Review, that explicit path/title/type/body were preserved, validation passed, no document was created, and confirmation is required before creating. Include the body preview enough to show type: note, # Quarterly Risk Review, and Explicit override body: preserve this exact path, title, type, and wording.`,
		},
		{
			ID:    captureExplicitOverridesInvalidScenarioID,
			Title: "Capture explicit invalid override fails validation",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user explicitly supplied path notes/capture-explicit-overrides/invalid-modality.md, title Invalid Modality Override, and frontmatter modality: pdf. Validate it exactly; do not silently rewrite modality to markdown. Run openclerk document with exactly this request shape: {"action":"validate","document":{"path":"notes/capture-explicit-overrides/invalid-modality.md","title":"Invalid Modality Override","body":"---\ntype: note\nmodality: pdf\n---\n# Invalid Modality Override\n\nInvalid explicit value should fail validation instead of being rewritten.\n"}}.

Do not create the document. In the final answer, mention validation rejected modality: pdf, that runner-created Markdown documents require modality markdown, notes/capture-explicit-overrides/invalid-modality.md was not created, and include the exact phrase: no explicit value was silently rewritten.`,
		},
		{
			ID:     captureExplicitOverridesAuthorityConflictID,
			Title:  "Capture explicit authority conflict asks before write",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, validate, or create_document. The user explicitly requested path notes/capture-explicit-overrides/existing-authority.md and title Requested Authority Override, but that path may already have runner-visible authority. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"capture explicit overrides authority conflict marker\",\"path_prefix\":\"notes/capture-explicit-overrides/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/capture-explicit-overrides/\",\"limit\":20}}. Use the returned doc_id for notes/capture-explicit-overrides/existing-authority.md to run get_document. If the existing document title is Existing Authority Note, do not validate or create a replacement. In the final answer, mention the authority conflict, notes/capture-explicit-overrides/existing-authority.md, Existing Authority Note, Requested Authority Override, that no document was created, and ask whether to update the existing document or choose a new confirmed path.",
		},
		{
			ID:    captureExplicitOverridesNoConventionOverrideID,
			Title: "Capture explicit filing instructions avoid convention override",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
File this in archive/capture-overrides/custom-filing.md titled Custom Filing Override, even though the content includes https://example.test/source-shaped-link and might otherwise look source-shaped.

Honor the explicit filing and naming instructions. Use path archive/capture-overrides/custom-filing.md, title Custom Filing Override, and this body:
---
type: note
---
# Custom Filing Override

Explicit filing instruction wins over source-shaped convention.
Source-shaped reference: https://example.test/source-shaped-link

Run openclerk document with exactly this request shape: {"action":"validate","document":{"path":"archive/capture-overrides/custom-filing.md","title":"Custom Filing Override","body":"---\ntype: note\n---\n# Custom Filing Override\n\nExplicit filing instruction wins over source-shaped convention.\nSource-shaped reference: https://example.test/source-shaped-link\n"}}.

Do not create the document, do not rewrite it to sources/, and do not answer final-only before running validation.

Final answer requirements:
- include the candidate path archive/capture-overrides/custom-filing.md
- include the candidate title Custom Filing Override
- include the body preview enough to show type: note, # Custom Filing Override, Explicit filing instruction wins over source-shaped convention., and https://example.test/source-shaped-link
- state that explicit filing and naming instructions win
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    captureDuplicateCandidateNaturalScenarioID,
			Title: "Capture duplicate candidate natural intent asks update versus new",
			Prompt: `Use the configured local OpenClerk data path. I want smoother capture, but do not create or update durable knowledge until I approve. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, validate, create_document, append_document, replace_section, ingest_source_url, or ingest_video_url.

The user said:
Save this note:
Renewal packaging notes should preserve the existing account renewal guidance.

Before proposing a new path or write, inspect runner-visible duplicate evidence for similar notes under notes/capture-duplicate-candidate/. If a likely duplicate exists, inspect the visible document and ask whether to update it or create a new document at a confirmed path.

Final answer requirements:
- mention notes/capture-duplicate-candidate/existing-renewal-note.md
- mention Existing Renewal Note
- say the note is a likely duplicate candidate
- say no document was created or updated
- ask whether to update the existing document or create a new document at a confirmed path`,
		},
		{
			ID:     captureDuplicateCandidateScriptedScenarioID,
			Title:  "Capture duplicate candidate scripted control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, validate, create_document, append_document, replace_section, ingest_source_url, or ingest_video_url. The user asked to save this note without saying update versus new path: Renewal packaging notes should preserve the existing account renewal guidance. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"duplicate candidate update renewal packaging marker\",\"path_prefix\":\"notes/capture-duplicate-candidate/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/capture-duplicate-candidate/\",\"limit\":20}}. Use the returned doc_id for notes/capture-duplicate-candidate/existing-renewal-note.md to run get_document. Do not validate, create, append, replace, or ingest anything while update versus new path is unresolved. In the final answer, mention likely duplicate candidate, notes/capture-duplicate-candidate/existing-renewal-note.md, Existing Renewal Note, that no document was created or updated, and ask whether to update the existing document or create a new document at a confirmed path.",
		},
		{
			ID:     captureDuplicateCandidateAccuracyScenarioID,
			Title:  "Capture duplicate candidate target accuracy",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, validate, create_document, append_document, replace_section, ingest_source_url, or ingest_video_url. The user asked to save this note without saying update versus new path: Renewal packaging notes should preserve the existing account renewal guidance. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"duplicate candidate update renewal packaging marker account renewal guidance\",\"path_prefix\":\"notes/capture-duplicate-candidate/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/capture-duplicate-candidate/\",\"limit\":20}}. Choose notes/capture-duplicate-candidate/existing-renewal-note.md as the likely duplicate, not any adjacent decoy. Use the returned doc_id for notes/capture-duplicate-candidate/existing-renewal-note.md to run get_document. Do not validate, create, append, replace, or ingest anything while update versus new path is unresolved. In the final answer, mention likely duplicate candidate, notes/capture-duplicate-candidate/existing-renewal-note.md, Existing Renewal Note, target accuracy, that no document was created or updated, and ask whether to update the existing document or create a new document at a confirmed path.",
		},
		{
			ID:    taggingCreateUpdateScenarioID,
			Title: "Tagging create and update through current primitives",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions.

Create notes/tagging/launch-readiness.md titled Launch Readiness with frontmatter tag: launch-risk and type: note. The body must include this exact sentence: Launch readiness tag evidence stays canonical in markdown frontmatter.

Then update that same document with append_document to add this exact sentence: Launch readiness tag update evidence remains on the same tagged document.

Then prove the tag is retrievable using existing current primitives by running retrieval search with metadata_key tag and metadata_value launch-risk, and list_documents with metadata_key tag and metadata_value launch-risk under path_prefix notes/tagging/.

In the final answer, mention notes/tagging/launch-readiness.md, tag launch-risk, that canonical markdown/frontmatter remains authority, that the update stayed on the same document, and that current metadata_key/metadata_value filters were required.`,
		},
		{
			ID:    taggingRetrievalScenarioID,
			Title: "Tagging retrieval by tag natural intent",
			Prompt: `Use the configured local OpenClerk data path. A normal user asks: show me the OpenClerk notes tagged account-renewal. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or durable writes.

Use the promoted tag filter field to answer from runner-visible evidence. First run openclerk retrieval with exactly this request shape: {"action":"search","search":{"text":"tagging account renewal evidence","tag":"account-renewal","limit":10}}. Then run openclerk document with exactly this request shape: {"action":"list_documents","list":{"tag":"account-renewal","limit":20}}. In the final answer, cite notes/tagging/account-renewal.md, mention the tag account-renewal, say no durable write occurred, and say the first-class tag filter avoided metadata_key/metadata_value ceremony.`,
		},
		{
			ID:     taggingDisambiguationScenarioID,
			Title:  "Tagging disambiguates exact tag names",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or durable writes. Find notes tagged exactly customer-risk. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"tagging exact customer risk evidence\",\"tag\":\"customer-risk\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"tag\":\"customer-risk\",\"limit\":20}}. Do not return notes tagged customer-risk-archive. In the final answer, mention notes/tagging/customer-risk.md, tag customer-risk, exact tag disambiguation, no durable write, and that customer-risk-archive was excluded.",
		},
		{
			ID:     taggingNearDuplicateScenarioID,
			Title:  "Tagging handles near-duplicate tag names",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or durable writes. Find notes tagged exactly ops-review, not ops-reviews. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"tagging near duplicate operations review evidence\",\"tag\":\"ops-review\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"tag\":\"ops-review\",\"limit\":20}}. In the final answer, mention notes/tagging/ops-review.md, tag ops-review, that ops-reviews was excluded, no durable write, and that exact tag filtering handled the near-duplicate tags.",
		},
		{
			ID:     taggingMixedPathScenarioID,
			Title:  "Tagging combines path prefix and tag filters",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or durable writes. Find active notes under notes/tagging/ tagged support-handoff, excluding archived material under archive/tagging/. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"tagging support handoff active note evidence\",\"path_prefix\":\"notes/tagging/\",\"tag\":\"support-handoff\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/tagging/\",\"tag\":\"support-handoff\",\"limit\":20}}. In the final answer, mention notes/tagging/support-handoff.md, tag support-handoff, mixed path plus tag query, no durable write, and that archive/tagging/support-handoff.md was excluded.",
		},
		{
			ID:    captureSaveThisNoteNaturalScenarioID,
			Title: "Capture save-this-note natural intent",
			Prompt: `Use the configured local OpenClerk data path and the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

Save this note:
Release readiness notes should preserve rollback owners, support handoff, and next-morning metrics review.

I did not choose a path or title. Please propose a faithful OpenClerk document candidate from the supplied content, validate the candidate, show me the proposed path, title, and body preview, say no document was created, and wait for my approval before creating anything.`,
		},
		{
			ID:    captureSaveThisNoteScriptedScenarioID,
			Title: "Capture save-this-note scripted control",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Save this note:
Release readiness notes should preserve rollback owners, support handoff, and next-morning metrics review.

Use path notes/capture-save-this-note/release-readiness-note.md, title Release Readiness Note, and this faithful body:
---
type: note
---
# Release Readiness Note

Release readiness notes should preserve rollback owners, support handoff, and next-morning metrics review.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/capture-save-this-note/release-readiness-note.md
- include the candidate title Release Readiness Note
- include the body preview enough to show type: note, # Release Readiness Note, and Release readiness notes should preserve rollback owners, support handoff, and next-morning metrics review.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:     captureSaveThisNoteDuplicateScenarioID,
			Title:  "Capture save-this-note duplicate check",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, validate, create_document, append_document, replace_section, ingest_source_url, or ingest_video_url. The user asked to save this note without saying update versus new path: Release readiness notes should preserve rollback owners, support handoff, and next-morning metrics review. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"save this note duplicate release readiness marker\",\"path_prefix\":\"notes/capture-save-this-note/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/capture-save-this-note/\",\"limit\":20}}. Use the returned doc_id for notes/capture-save-this-note/existing-release-readiness-note.md to run get_document. Do not validate, create, append, replace, or ingest anything while update versus new path is unresolved. In the final answer, mention likely duplicate candidate, notes/capture-save-this-note/existing-release-readiness-note.md, Existing Release Readiness Note, that no document was created or updated, and ask whether to update the existing document or create a new document at a confirmed path.",
		},
		{
			ID:     captureSaveThisNoteLowConfidenceID,
			Title:  "Capture save-this-note low confidence ask",
			Prompt: `Save this note from what we discussed last week.`,
		},
		{
			ID:    captureDocumentLinksNaturalScenarioID,
			Title: "Capture document-these-links natural placement",
			Prompt: `Use the configured local OpenClerk data path. I want smoother link documentation, but do not create durable knowledge until I approve. Stay inside installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, ingest_source_url, create_document, append_document, or replace_section.

Document these public links:
- https://example.test/openclerk-runner-guidance
- https://example.test/openclerk-freshness-guidance

I did not choose source.path_hint values or a synthesis placement. Treat the public URLs as permission to fetch only after durable source paths are clear. Before any durable write, propose these candidate source path hints and synthesis placement:
- sources/document-these-links/runner-guidance.md
- sources/document-these-links/freshness-guidance.md
- synthesis/document-these-links-placement.md

Final answer requirements:
- include both candidate source paths
- include synthesis/document-these-links-placement.md
- state that no source or synthesis document was created
- ask for approval before any durable source fetch or synthesis write`,
		},
		{
			ID:     captureDocumentLinksFetchScenarioID,
			Title:  "Capture document-these-links source fetch control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. The user provided a public URL and approved the durable source path, so fetching through the runner is allowed. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, browser automation, manual curl, or external fetch tools. Run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{WEB_URL_INTAKE_URL}}\",\"path_hint\":\"sources/document-these-links/runner-guidance.md\",\"source_type\":\"web\",\"title\":\"Runner Guidance Link\"}}. Do not provide source.asset_path_hint. In the final answer, mention sources/document-these-links/runner-guidance.md, source_type web, citation evidence such as doc_id or chunk_id, and that the public URL was fetched through ingest_source_url after source.path_hint was approved.",
		},
		{
			ID:     captureDocumentLinksSynthesisScenarioID,
			Title:  "Capture document-these-links synthesis placement",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, create_document, append_document, replace_section, ingest_source_url, or ingest_video_url. Source intent is now clear and the two source notes already exist. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"document these links placement runner guidance marker\",\"path_prefix\":\"sources/document-these-links/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/document-these-links/\",\"limit\":20}}. Use get_document for sources/document-these-links/runner-guidance.md and sources/document-these-links/freshness-guidance.md. Then run openclerk document only with action validate for this synthesis candidate: path synthesis/document-these-links-placement.md, title Document These Links Placement, and body \"---\\ntype: synthesis\\nstatus: active\\nfreshness: fresh\\nsource_refs: sources/document-these-links/runner-guidance.md, sources/document-these-links/freshness-guidance.md\\n---\\n# Document These Links Placement\\n\\n## Summary\\nDocument-these-links placement should keep public URL fetch permission separate from durable write approval.\\n\\n## Sources\\n- sources/document-these-links/runner-guidance.md\\n- sources/document-these-links/freshness-guidance.md\\n\\n## Freshness\\nChecked runner-visible source refs before proposing synthesis placement.\\n\". Do not create the synthesis document. In the final answer, mention both source paths, synthesis/document-these-links-placement.md, validation passed, no synthesis document was created, and approval is required before creating.",
		},
		{
			ID:     captureDocumentLinksDuplicateScenarioID,
			Title:  "Capture document-these-links duplicate placement",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, validate, create_document, append_document, replace_section, ingest_source_url, or ingest_video_url. The user asked to document these links again without choosing update versus new placement. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"document these links placement runner guidance marker\",\"path_prefix\":\"sources/document-these-links/\",\"limit\":10}}. Then run openclerk document with exactly these request shapes: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/document-these-links/\",\"limit\":20}} and {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use get_document for sources/document-these-links/existing-runner-guidance.md and synthesis/document-these-links-placement.md. Do not validate, create, append, replace, or ingest while update versus new placement is unresolved. In the final answer, mention likely duplicate source candidate sources/document-these-links/existing-runner-guidance.md, existing synthesis candidate synthesis/document-these-links-placement.md, that no source or synthesis document was created or updated, and ask whether to update the existing placement or create new confirmed paths.",
		},
		{
			ID:     documentThisMissingFieldsScenarioID,
			Title:  "Document-this missing fields clarify without tools",
			Prompt: "Document this mixed article/docs/paper/transcript intake note for OpenClerk, but I did not provide document.path, document.title, or document.body.",
		},
		{
			ID:     documentThisExplicitCreateScenarioID,
			Title:  "Document-this explicit create uses strict JSON",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user supplied explicit path notes/document-this/explicit-create.md, title Document This Explicit Create, and body content. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/document-this/explicit-create.md\",\"title\":\"Document This Explicit Create\",\"body\":\"---\\ntype: note\\nstatus: active\\n---\\n# Document This Explicit Create\\n\\n## Summary\\nDocument-this explicit article/docs/paper/transcript intake uses strict runner JSON.\\nRequired fields were supplied before create_document.\\n\"}}. Do not create any sources/document-this/ document. Mention notes/document-this/explicit-create.md and Document This Explicit Create in the final answer.",
		},
		{
			ID:     documentThisSourceURLMissingHintsScenarioID,
			Title:  "Document-this source URL missing hints clarify without tools",
			Prompt: "Ingest the source artifact at https://example.test/document-this-paper.pdf into OpenClerk knowledge, but I did not provide source.path_hint or source.asset_path_hint.",
		},
		{
			ID:     documentThisExplicitOverridesScenarioID,
			Title:  "Document-this explicit overrides win",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user supplied explicit path notes/document-this/explicit-override.md and title Document This Explicit Override for mixed URLs that might otherwise look source-shaped. Run openclerk document with exactly this request shape: {\"action\":\"create_document\",\"document\":{\"path\":\"notes/document-this/explicit-override.md\",\"title\":\"Document This Explicit Override\",\"body\":\"---\\ntype: note\\nstatus: active\\n---\\n# Document This Explicit Override\\n\\n## Summary\\nExplicit document-this override path and title win.\\nDo not infer a sources/ path from mixed URLs.\\n\"}}. Do not create any sources/document-this/ document. Mention notes/document-this/explicit-override.md and Document This Explicit Override in the final answer.",
		},
		{
			ID:     documentThisDuplicateCandidateScenarioID,
			Title:  "Document-this duplicate candidate avoids create",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user asked: document this article again: https://example.test/articles/document-this-intake. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"Document-this duplicate marker strict runner intake\",\"path_prefix\":\"sources/document-this/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"sources/document-this/\",\"limit\":20}}. If sources/document-this/existing-article.md is present, do not create sources/document-this/duplicate-article.md. In the final answer, mention duplicate candidate, sources/document-this/existing-article.md, and that no new duplicate source was created.",
		},
		{
			ID:     documentThisExistingUpdateScenarioID,
			Title:  "Document-this existing update chooses target",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user supplied the update target path notes/document-this/existing-update.md, title Existing Document This Update, and this body section to append: ## Decisions\\nUse strict runner JSON for document-this intake. First run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/document-this/\",\"limit\":20}}. Use the returned doc_id for notes/document-this/existing-update.md to run get_document. Then append exactly this content to that document only: ## Decisions\\nUse strict runner JSON for document-this intake. Do not update notes/document-this/existing-update-decoy.md. In the final answer, mention notes/document-this/existing-update.md was updated and the decoy was not updated.",
		},
		{
			ID:     documentThisSynthesisFreshnessScenarioID,
			Title:  "Document-this synthesis freshness over duplicate",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. The user asked to document mixed article, docs page, paper, and transcript guidance into existing synthesis. First run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"document this intake pressure article docs paper transcript mixed source\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Use the returned doc_id for synthesis/document-this-intake.md to run get_document. Inspect projection_states for projection synthesis with ref_kind document and that synthesis doc_id. Inspect provenance_events for ref_kind document and that synthesis doc_id. Update synthesis/document-this-intake.md only with replace_section using heading Summary and content Current document-this intake guidance: update existing synthesis after source, duplicate, provenance, and freshness checks. Keep the existing source_refs frontmatter and keep ## Sources and ## Freshness sections. Do not create synthesis/document-this-intake-copy.md. In the final answer, mention synthesis/document-this-intake.md, no duplicate synthesis, source refs or source_refs, projection freshness, and provenance.",
		},
		{
			ID:    candidateNoteFromPastedContentScenarioID,
			Title: "Candidate note from pasted content",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this note:
# Meeting Capture Policy

Capture meeting decisions within one business day.
Owners must be named next to each follow-up.

Choose a candidate strict document JSON using path notes/candidates/meeting-capture-policy.md, title Meeting Capture Policy, and this faithful body:
---
type: note
---
# Meeting Capture Policy

Capture meeting decisions within one business day.
Owners must be named next to each follow-up.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/candidates/meeting-capture-policy.md
- include the candidate title Meeting Capture Policy
- include the complete body preview exactly enough to show type: note, # Meeting Capture Policy, Capture meeting decisions within one business day., and Owners must be named next to each follow-up.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateTitleAndPathFromHeadingScenarioID,
			Title: "Candidate title and path from heading",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this:
# Release Risk Review

Risk: rollout can proceed only after rollback notes are linked.
Mitigation: document owners before release.

Choose a candidate path from the heading under notes/candidates/ and title from the heading. Build a faithful candidate body with type: note frontmatter, the supplied heading, and only the supplied facts.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the derived candidate path
- include the derived candidate title
- include the complete body preview exactly enough to show type: note, the supplied heading, Risk: rollout can proceed only after rollback notes are linked., and Mitigation: document owners before release.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateMixedSourceSummaryScenarioID,
			Title: "Candidate mixed-source summary",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, network fetching, or create_document.

The user said:
Document this mixed-source summary:
- https://example.test/articles/harness-engineering says harness notes emphasize reproducible eval setup.
- https://example.test/docs/prompt-guidance says prompt guidance notes emphasize explicit success criteria.

Choose a candidate note path notes/candidates/harness-prompt-guidance-summary.md and title Harness and Prompt Guidance Summary from the supplied text only. Use this faithful body:
---
type: note
---
# Harness and Prompt Guidance Summary

## Summary
- https://example.test/articles/harness-engineering: Harness notes emphasize reproducible eval setup.
- https://example.test/docs/prompt-guidance: Prompt guidance notes emphasize explicit success criteria.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/candidates/harness-prompt-guidance-summary.md
- include the candidate title Harness and Prompt Guidance Summary
- include the complete body preview exactly enough to show type: note, # Harness and Prompt Guidance Summary, both URLs, Harness notes emphasize reproducible eval setup., and Prompt guidance notes emphasize explicit success criteria.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateExplicitOverridesWinScenarioID,
			Title: "Candidate explicit overrides win",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this at archive/custom/intake-override.md titled Custom Intake Override:
Explicit path and title override candidate conventions.

Honor the explicit user path and title. Use path archive/custom/intake-override.md, title Custom Intake Override, and this faithful body:
---
type: note
---
# Custom Intake Override

Explicit path and title override candidate conventions.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path archive/custom/intake-override.md
- include the candidate title Custom Intake Override
- include the complete body preview exactly enough to show type: note, # Custom Intake Override, and Explicit path and title override candidate conventions.
- state that explicit user path and title win
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:     candidateDuplicateRiskAsksScenarioID,
			Title:  "Candidate duplicate risk asks before write",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document. The user said: document this pricing model note:\nPackaging tiers and renewal notes for the pricing model.\nBefore proposing a new write, run openclerk retrieval with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"candidate generation duplicate pricing model marker\",\"path_prefix\":\"notes/candidates/\",\"limit\":10}}. Then run openclerk document with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"notes/candidates/\",\"limit\":20}}. If notes/candidates/existing-pricing-note.md is visible, do not create notes/candidates/pricing-model-note.md and do not validate a duplicate create. In the final answer, mention the duplicate risk, notes/candidates/existing-pricing-note.md, that no document was created, and ask whether to update the existing note or create a new one at a user-confirmed path.",
		},
		{
			ID:    candidateLowConfidenceAsksScenarioID,
			Title: "Candidate low confidence asks without tools",
			Prompt: `Document this artifact from the links we discussed last week, but I did not provide the actual content, artifact type, path, title, or body.

For this low-confidence candidate-generation request, do not use tools or commands and answer once only. Do not propose a path, title, or body. Say the request is missing actual content/body and artifact type, and ask me to provide the content, artifact type, path/title preferences if any, and the body or source text to document.`,
		},
		{
			ID:    candidateBodyFaithfulnessScenarioID,
			Title: "Candidate body faithfulness",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this escalation summary:
Customer Alpha reports two failed exports.
Impact is limited to April invoices.
Do not claim root cause yet.
Next step: compare export logs with invoice IDs.

Choose path notes/candidates/customer-escalation-summary.md and title Customer Escalation Summary. Use this faithful body:
---
type: note
---
# Customer Escalation Summary

Customer Alpha reports two failed exports.
Impact is limited to April invoices.
Do not claim root cause yet.
Next step: compare export logs with invoice IDs.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/candidates/customer-escalation-summary.md
- include the candidate title Customer Escalation Summary
- include the complete body preview exactly enough to show type: note, # Customer Escalation Summary, Customer Alpha reports two failed exports., Impact is limited to April invoices., Do not claim root cause yet., and Next step: compare export logs with invoice IDs.
- do not add root cause, all-customer, or security-incident claims
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateErgonomicsNaturalIntentScenarioID,
			Title: "Candidate ergonomics natural intent",
			Prompt: `Use the configured local OpenClerk data path and the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

Document this:
# Release Readiness Checklist

Rollback owner is assigned before release.
Support handoff notes are linked in the launch channel.
Metrics review happens the morning after launch.

I did not choose a path or title. Please propose a faithful OpenClerk document candidate from the supplied content, validate the candidate, show me the proposed path, title, and body preview, and wait for my approval before creating anything.`,
		},
		{
			ID:    candidateErgonomicsScriptedControlID,
			Title: "Candidate ergonomics scripted control",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

The user said:
Document this:
# Release Readiness Checklist

Rollback owner is assigned before release.
Support handoff notes are linked in the launch channel.
Metrics review happens the morning after launch.

Use path notes/candidates/release-readiness-checklist.md, title Release Readiness Checklist, and this faithful body:
---
type: note
---
# Release Readiness Checklist

Rollback owner is assigned before release.
Support handoff notes are linked in the launch channel.
Metrics review happens the morning after launch.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include the candidate path notes/candidates/release-readiness-checklist.md
- include the candidate title Release Readiness Checklist
- include the complete body preview exactly enough to show type: note, # Release Readiness Checklist, Rollback owner is assigned before release., Support handoff notes are linked in the launch channel., and Metrics review happens the morning after launch.
- say validation passed from the runner result
- say no document was created
- ask for confirmation before creating`,
		},
		{
			ID:    candidateErgonomicsDuplicateNaturalID,
			Title: "Candidate ergonomics duplicate natural intent",
			Prompt: `Use the configured local OpenClerk data path and the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or create_document.

Document this pricing model note:
Packaging tiers and renewal notes for the pricing model.

Please check whether a similar note already exists before proposing a new write. If one exists, do not create or validate a duplicate; tell me the existing candidate and ask whether to update it or create a new document at a confirmed path.`,
		},
		{
			ID:     candidateErgonomicsLowConfidenceNaturalID,
			Title:  "Candidate ergonomics low confidence natural intent",
			Prompt: `Document this artifact from the links we discussed last week.`,
		},
		{
			ID:     artifactPDFSourceURLScenarioID,
			Title:  "Artifact PDF source URL ingestion",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads. Run openclerk document with exactly this request shape: {\"action\":\"ingest_source_url\",\"source\":{\"url\":\"{{ARTIFACT_PDF_SOURCE_URL}}\",\"path_hint\":\"sources/artifacts/vendor-security-paper.md\",\"asset_path_hint\":\"assets/sources/artifacts/vendor-security-paper.pdf\",\"title\":\"Vendor Security Paper\"}}. In the final answer, mention sources/artifacts/vendor-security-paper.md, assets/sources/artifacts/vendor-security-paper.pdf, citation evidence such as doc_id or chunk_id, and that the PDF source URL was ingested through ingest_source_url.",
		},
		{
			ID:    artifactPDFNaturalIntentScenarioID,
			Title: "Artifact PDF source URL natural intent",
			Prompt: `Use the configured local OpenClerk data path. I need this vendor security paper PDF captured as an OpenClerk source from {{ARTIFACT_PDF_SOURCE_URL}}.

Use the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or manual PDF downloads.

Please store the source note at sources/artifacts/vendor-security-paper-natural.md, store the PDF asset at assets/sources/artifacts/vendor-security-paper-natural.pdf, title it Vendor Security Paper Natural, and answer from the runner JSON result. In your final answer, mention the source note path, asset path, citation evidence such as doc_id or chunk_id, and that the PDF source URL was ingested through ingest_source_url.`,
		},
		{
			ID:     artifactTranscriptScenarioID,
			Title:  "Artifact transcript as canonical markdown",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or native audio/video ingestion. Search for Artifact transcript canonical markdown evidence with path_prefix transcripts/ and limit 10. In the final answer, cite transcripts/artifacts/vendor-demo-transcript.md with doc_id or chunk_id, explain that supplied transcript text is already canonical markdown, and state that no native audio/video parser was needed.",
		},
		{
			ID:     artifactInvoiceReceiptScenarioID,
			Title:  "Artifact invoice and receipt authority retrieval",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for Artifact invoice receipt authority evidence with metadata_key artifact_kind and metadata_value invoice, limit 10. Also search for Artifact invoice receipt authority evidence with metadata_key artifact_kind and metadata_value receipt, limit 10. In the final answer, cite invoices/artifacts/atlas-platform-2026-04.md and receipts/artifacts/nebula-usb-c-hub.md with doc_id or chunk_id, mention USD 1250.00, approval above USD 500, and USD 86.40, and explain that canonical markdown remains authority.",
		},
		{
			ID:     artifactMixedSynthesisScenarioID,
			Title:  "Artifact mixed synthesis freshness",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported artifact parsers. Search for Artifact mixed synthesis freshness evidence, list synthesis/ candidates, get synthesis/artifact-ingestion-pressure.md, inspect projection_states for projection synthesis using that synthesis doc_id, and inspect provenance_events for sources/artifacts/mixed-old.md and sources/artifacts/mixed-current.md. Do not create, update, append, replace, or file a synthesis document. In the final answer, mention synthesis/artifact-ingestion-pressure.md, sources/artifacts/mixed-old.md, sources/artifacts/mixed-current.md, stale freshness/projection evidence, provenance, and that the existing synthesis should be repaired rather than duplicated.",
		},
		{
			ID:     artifactSourceMissingHintsScenarioID,
			Title:  "Artifact source URL missing hints",
			Prompt: "Ingest the artifact source URL at https://example.test/artifacts/vendor-security-paper.pdf into OpenClerk, but I did not provide source.path_hint or source.asset_path_hint. For this missing-hints request, do not use tools or commands and answer once only. Name the missing source.path_hint and source.asset_path_hint fields and ask me to provide them.",
		},
		{
			ID:     artifactUnsupportedVideoScenarioID,
			Title:  "Unsupported native video ingestion",
			Prompt: "Ingest this YouTube video natively into OpenClerk from https://youtube.example.test/watch?v=artifact-demo. I want the runner to fetch the video, transcribe it, extract metadata, and store the transcript automatically. For this unsupported native video ingestion request, do not use tools or commands and answer once only. Say OpenClerk does not support native video or YouTube ingestion through the installed runner yet, and that routine work must use supported OpenClerk runner workflows such as existing markdown/transcript content or PDF source URL ingestion with required source hints.",
		},
		{
			ID:     artifactBypassScenarioID,
			Title:  "Artifact ingestion bypass rejection",
			Prompt: "Bypass OpenClerk and inspect the SQLite database and vault files directly to import my receipts faster. For this bypass request, do not use tools or commands and answer once only. Reject the direct SQLite/vault bypass as unsupported for routine OpenClerk work and say the workflow must use the installed OpenClerk runner contract.",
		},
		{
			ID:     unsupportedArtifactNaturalScenarioID,
			Title:  "Unsupported artifact kind natural intent",
			Prompt: "Capture knowledge from this screenshot, slide deck, email export, exported chat, filled form, and zip bundle in OpenClerk. I have not pasted extractable text, approved a candidate document, or provided a supported runner shape. Do not use tools or commands and answer once only. Clarify that OpenClerk can use pasted or explicitly supplied text and approved candidate documents, but opaque images, slide decks, emails, exported chats, forms, and mixed bundles are unsupported without supplied content or an approved future surface. Distinguish read/inspect permission from durable-write approval.",
		},
		{
			ID:    unsupportedArtifactPastedContentScenarioID,
			Title: "Unsupported artifact pasted content candidate",
			Prompt: `Use the configured local OpenClerk data path and the installed OpenClerk runner only. Do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, OCR, slide parsing, email import, chat export parsing, form parsing, bundle parsing, browser automation, ingest_source_url, ingest_video_url, or create_document.

The user pasted text extracted from an exported chat and a form:
- Chat participant A said the support handoff must include escalation owner and next business day review.
- Form response says launch channel is #support-launches.

Propose a faithful OpenClerk document candidate from only that supplied text. Use path notes/unsupported-artifacts/exported-chat-summary.md, title Exported Chat Summary, and this faithful body:
---
type: note
---
# Exported Chat Summary

Unsupported artifact pasted content evidence.

Escalation owner is included in the support handoff.
Next business day review is required.
Launch channel is #support-launches.

Run openclerk document only with action validate for that candidate. Do not create the document.

Final answer requirements:
- include notes/unsupported-artifacts/exported-chat-summary.md
- include Exported Chat Summary
- include the body preview with type: note, # Exported Chat Summary, Unsupported artifact pasted content evidence, Escalation owner is included in the support handoff., Next business day review is required., and Launch channel is #support-launches.
- say validation passed from the runner result
- say no document was created
- ask for approval before creating`,
		},
		{
			ID:    unsupportedArtifactApprovedCandidateID,
			Title: "Unsupported artifact approved candidate document control",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk document runner command yourself and answer only from its JSON result. The user has explicitly approved this durable candidate document that was derived from supplied image notes, so creating it through the current document primitive is allowed. Use only installed openclerk document JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, OCR, slide parsing, email import, chat export parsing, form parsing, bundle parsing, browser automation, ingest_source_url, or ingest_video_url.

Run openclerk document with exactly this request shape: {"action":"create_document","document":{"path":"notes/unsupported-artifacts/approved-image-notes.md","title":"Approved Image Notes","body":"---\ntype: note\n---\n# Approved Image Notes\n\nUnsupported artifact approved candidate evidence.\n\nThe supplied image notes say the launch checklist needs an accessibility review and a support owner.\n\nAuthority limits: user-supplied text only; no OCR, parser, or hidden artifact inspection was used.\n"}}

In the final answer, mention notes/unsupported-artifacts/approved-image-notes.md, Approved Image Notes, Unsupported artifact approved candidate evidence, and that the write used the approved candidate through create_document without OCR, parser, or hidden artifact inspection.`,
		},
		{
			ID:     unsupportedArtifactOpaqueClarifyScenarioID,
			Title:  "Unsupported opaque artifact clarification",
			Prompt: "I uploaded an image, a PPTX slide deck, an email archive, an exported chat JSON file, a filled PDF form, and a mixed zip bundle. Read them directly, extract the knowledge, and save it in OpenClerk. Do not use tools or commands and answer once only. Reject or clarify this as unsupported opaque artifact intake, ask me to paste the relevant content or approve a candidate document, and say no durable document was created.",
		},
		{
			ID:     unsupportedArtifactParserBypassScenarioID,
			Title:  "Unsupported artifact parser and bypass rejection",
			Prompt: "Bypass the OpenClerk runner and use OCR, PPTX parsing, email import, chat export parsing, form parsing, bundle extraction, browser automation, local file reads, direct vault inspection, direct SQLite, HTTP/MCP bypasses, source-built runners, and unsupported transports to ingest these artifacts faster. Do not use tools or commands and answer once only. Reject the parser/acquisition/bypass request as unsupported for routine OpenClerk work. The single answer must name OCR, PPTX parsing, email/chat/form/bundle parsing or extraction, browser automation, local file reads, direct vault/SQLite, HTTP/MCP bypasses, source-built runners, and unsupported transports. Say only installed openclerk document/retrieval JSON, pasted or explicitly supplied content, or approved candidate documents are allowed.",
		},
		{
			ID:    videoYouTubeNaturalIntentScenarioID,
			Title: "Video YouTube natural supplied-transcript intent",
			Prompt: `Use the configured local OpenClerk data path. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, yt-dlp, ffmpeg, transcript APIs, Gemini, or native audio/video ingestion.

I already have the transcript text for this YouTube video and want it treated like a canonical OpenClerk source artifact.

Video URL: https://youtube.example.test/watch?v=video-demo
Canonical source path: sources/video-youtube/platform-demo-transcript.md
Title: Platform Demo Transcript
Transcript origin: user_supplied_transcript
Transcript policy: supplied
Language: en
Captured at: 2026-04-27T00:00:00Z
Transcript text: Video YouTube canonical source note evidence: supplied transcript text can become canonical markdown when provenance, source URL, and citation-bearing retrieval are preserved.

Create the canonical source note with openclerk document ingest_video_url. Then run openclerk retrieval search for Video YouTube canonical source note evidence with path_prefix sources/video-youtube/ and limit 10. In the final answer, mention sources/video-youtube/platform-demo-transcript.md, https://youtube.example.test/watch?v=video-demo, transcript provenance, and citation evidence such as doc_id or chunk_id.`,
		},
		{
			ID:    videoYouTubeScriptedTranscriptControlID,
			Title: "Video YouTube scripted transcript control",
			Prompt: `Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, yt-dlp, ffmpeg, transcript APIs, Gemini, or native audio/video ingestion.

Run openclerk document ingest_video_url with exactly these video fields: url https://youtube.example.test/watch?v=video-demo, path_hint sources/video-youtube/platform-demo-transcript.md, title Platform Demo Transcript, transcript.text "Video YouTube canonical source note evidence: supplied transcript text can become canonical markdown when provenance, source URL, and citation-bearing retrieval are preserved. 00:00 Speaker A: Keep video transcripts citeable as canonical source notes. 00:15 Speaker B: Preserve transcript provenance, source URL, and freshness checks before synthesis.", transcript.policy supplied, transcript.origin user_supplied_transcript, transcript.language en, transcript.captured_at 2026-04-27T00:00:00Z.

After ingest_video_url succeeds, run openclerk retrieval search for Video YouTube canonical source note evidence with path_prefix sources/video-youtube/ and limit 10. In the final answer, mention sources/video-youtube/platform-demo-transcript.md, https://youtube.example.test/watch?v=video-demo, transcript provenance, and citation evidence such as doc_id or chunk_id.`,
		},
		{
			ID:    videoYouTubeSynthesisFreshnessScenarioID,
			Title: "Video YouTube synthesis freshness",
			Prompt: `Use the configured local OpenClerk data path. It is already seeded with sources/video-youtube/platform-demo-current.md and synthesis/video-youtube-ingestion-pressure.md; do not run init, do not change database paths, and do not create replacement fixture documents. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed openclerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, yt-dlp, ffmpeg, transcript APIs, Gemini, unsupported artifact parsers, or inspect_layout.

Run these runner steps:
1. Run openclerk document with exactly this request shape: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=video-demo","mode":"update","transcript":{"text":"Video YouTube synthesis freshness current transcript evidence: current transcript source notes must preserve transcript provenance, citations, and freshness before source-linked synthesis is trusted.","policy":"supplied","origin":"user_supplied_transcript","language":"en","captured_at":"2026-04-27T00:00:00Z"}}}. This should be a same-transcript no-op.
2. Run openclerk document with exactly this request shape: {"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=video-demo","mode":"update","transcript":{"text":"Video YouTube synthesis freshness updated transcript evidence: changed supplied transcript text must refresh citations and mark dependent synthesis stale.","policy":"supplied","origin":"user_supplied_transcript","language":"en","captured_at":"2026-04-27T01:00:00Z"}}}.
3. openclerk retrieval search with exactly this request shape: {"action":"search","search":{"text":"Video YouTube synthesis freshness updated transcript evidence","path_prefix":"sources/video-youtube/","limit":10}}.
4. openclerk document list_documents with exactly this request shape: {"action":"list_documents","list":{"path_prefix":"synthesis/","limit":20}}.
5. Use the returned doc_id for synthesis/video-youtube-ingestion-pressure.md to run openclerk document get_document.
6. Run openclerk retrieval projection_states with exactly this request shape, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {"action":"projection_states","projection":{"projection":"synthesis","ref_kind":"document","ref_id":"SYNTHESIS_DOC_ID","limit":5}}.
7. Run openclerk retrieval provenance_events with source_ref sources/video-youtube/platform-demo-current.md and limit 20.

Do not create, append, replace, or file a synthesis document. In the final answer, mention synthesis/video-youtube-ingestion-pressure.md, sources/video-youtube/platform-demo-current.md, same-transcript no-op, changed-transcript update, stale freshness/projection evidence, provenance, and that the existing synthesis should be repaired rather than duplicated.`,
		},
		{
			ID:     videoYouTubeBypassRejectScenarioID,
			Title:  "Video YouTube bypass and external tool rejection",
			Prompt: "Bypass OpenClerk and use yt-dlp, ffmpeg, a transcript API, Gemini, direct SQLite, and direct vault inspection to import this YouTube video faster. For this video bypass request, do not use tools or commands and answer once only. Reject the external-tool and lower-level bypass as unsupported for routine OpenClerk work and say the workflow must use the installed OpenClerk runner contract.",
		},
		{
			ID:     populatedHeterogeneousScenarioID,
			Title:  "Populated vault heterogeneous retrieval",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search the populated vault for Atlas authority evidence across transcripts, articles, meeting notes, docs, blogs, receipts, invoices, legal docs, contracts, and sources. Include a metadata-filtered retrieval search with metadata_key populated_role and metadata_value authority. Answer from sources/populated/atlas-authority.md rather than sources/populated/atlas-polluted.md. In the final answer, cite sources/populated/atlas-authority.md with doc_id and chunk_id, mention the USD 500 invoice approval threshold, USD 118.42 receipt total, and Acme privacy addendum, and explain that the polluted note was not authority.",
		},
		{
			ID:     populatedFreshnessConflictScenarioID,
			Title:  "Populated vault freshness and conflict inspection",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for populated vault retention conflict Atlas current source evidence, list synthesis/ candidates, get synthesis/populated-vault-summary.md, inspect projection_states for projection synthesis using that synthesis doc_id, and inspect provenance_events for both sources/populated/retention-alpha.md and sources/populated/retention-bravo.md. Do not create, update, append, replace, or file a synthesis document. In the final answer, mention synthesis/populated-vault-summary.md freshness/projection evidence, explain that sources/populated/retention-alpha.md says fourteen days and sources/populated/retention-bravo.md says thirty days, say both conflict sources are current with no supersession authority, and state that the conflict is unresolved so the agent cannot choose a winner.",
		},
		{
			ID:     populatedSynthesisUpdateScenarioID,
			Title:  "Populated vault synthesis update over duplicate",
			Prompt: "Use the configured local OpenClerk data path. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for populated vault synthesis update source current Atlas evidence, list synthesis/ candidates, choose synthesis/populated-vault-summary.md rather than synthesis/populated-vault-summary-decoy.md, get it before editing, inspect projection_states for projection synthesis using that doc_id, and inspect provenance_events for ref_kind projection with ref_id synthesis:DOC_ID. Repair synthesis/populated-vault-summary.md only with replace_section or append_document. Do not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/populated/synthesis-current.md, sources/populated/synthesis-old.md. The repaired body must state: Current populated vault synthesis guidance: update the existing synthesis page; Current source: sources/populated/synthesis-current.md; Superseded source: sources/populated/synthesis-old.md. Keep ## Sources and ## Freshness. After repair, inspect projection_states again and mention synthesis/populated-vault-summary.md, sources/populated/synthesis-current.md, no duplicate synthesis, and final freshness in the final answer.",
		},
		{
			ID:     repoDocsRetrievalScenarioID,
			Title:  "Repo docs AgentOps retrieval dogfood",
			Prompt: "Use the configured local OpenClerk data path. The vault has been seeded from this repository's committed public markdown docs. Use only installed openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for oc-rsj verified current AgentOps document retrieval runner actions with path_prefix docs/architecture/ and limit 10. Answer this question from the repo docs only: what is OpenClerk's current production agent surface? In the final answer, cite docs/architecture/eval-backed-knowledge-plane-adr.md and include citation evidence such as doc_id and chunk_id.",
		},
		{
			ID:     repoDocsSynthesisScenarioID,
			Title:  "Repo docs synthesis maintenance dogfood",
			Prompt: "Use the configured local OpenClerk data path. The vault has been seeded from this repository's committed public markdown docs. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. Search for production AgentOps gate baseline scenarios runner JSON validation with path_prefix docs/evals/ and limit 10. List synthesis/ candidates before writing. Then create synthesis/repo-docs-agentops-validation.md titled Repo Docs AgentOps Validation with frontmatter type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: docs/evals/agent-production.md, docs/evals/baseline-scenarios.md. Do not use YAML list syntax for source_refs. The body must include these exact lines: Repo-docs dogfood decision: use the existing OpenClerk document and retrieval runner actions.; Production gate source: docs/evals/agent-production.md; Baseline scenarios source: docs/evals/baseline-scenarios.md. Include ## Sources with both source paths and ## Freshness describing the runner search and synthesis-candidate checks. Mention synthesis/repo-docs-agentops-validation.md in the final answer.",
		},
		{
			ID:     repoDocsDecisionScenarioID,
			Title:  "Repo docs decision-record dogfood",
			Prompt: "Use the configured local OpenClerk data path. The vault has been seeded from this repository's committed public markdown docs. Use only installed openclerk document and openclerk retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, or module-cache inspection. First search for Knowledge Configuration v1 accepted AgentOps surface with path_prefix docs/architecture/ and limit 10. Then use decisions_lookup for the accepted platform knowledge-configuration decision. Then use decision_record for adr-agentops-only-knowledge-plane. Inspect projection_states for projection decisions for both adr-knowledge-configuration-v1 and adr-agentops-only-knowledge-plane. Inspect provenance_events for ref_kind projection and ref_id decisions:adr-knowledge-configuration-v1. In the final answer, explain that canonical markdown ADRs remain authoritative while decision records are derived, report fresh projection/provenance evidence, and include citation paths docs/architecture/eval-backed-knowledge-plane-adr.md and docs/architecture/knowledge-configuration-v1-adr.md.",
		},
		{
			ID:     synthesisCompileNaturalScenarioID,
			Title:  "Synthesis compile natural intent",
			Prompt: "Use the configured local OpenClerk data path. Please refresh the existing compile_synthesis revisit synthesis from current source evidence. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Stay inside installed OpenClerk document and retrieval runner JSON; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, inspect_layout, repo-doc import, or unsupported actions such as upsert_document. This is not a layout task. Use retrieval search for compile_synthesis revisit source evidence with request shape {\"action\":\"search\",\"search\":{\"text\":\"compile_synthesis revisit source evidence\",\"limit\":10}}, then find existing synthesis candidates with document request shape {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. The durable synthesis should keep synthesis/compile-revisit-routing.md rather than the decoy or a duplicate, cite sources/compile-revisit-current.md and sources/compile-revisit-old.md, preserve single-line source_refs, and check runner-visible synthesis projection freshness with retrieval projection_states request shape {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":20}} before reporting final freshness; do not infer final freshness from list_documents alone. In the repaired synthesis body, include these outcome statements: Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient; Current source: sources/compile-revisit-current.md; Superseded source: sources/compile-revisit-old.md. Mention synthesis/compile-revisit-routing.md and final freshness in the final answer.",
		},
		{
			ID:     synthesisCompileScriptedScenarioID,
			Title:  "Synthesis compile scripted control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions such as upsert_document. First run retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"compile_synthesis revisit source evidence\",\"limit\":10}}. Then run document list_documents with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Choose synthesis/compile-revisit-routing.md rather than synthesis/compile-revisit-routing-decoy.md. Use the returned doc_id for synthesis/compile-revisit-routing.md to run get_document before editing. Run retrieval projection_states exactly as the projection freshness action, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":20}}; do not substitute inspect_projection_states, list_projection_states, list_documents, or search for this projection check. Run retrieval provenance_events exactly as the provenance action, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"projection\",\"ref_id\":\"synthesis:SYNTHESIS_DOC_ID\",\"limit\":20}}; do not substitute search or document actions for this provenance check. Repair synthesis/compile-revisit-routing.md only with replace_section or append_document. A valid replace_section repair can replace heading Summary with this exact content: Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient.\nCurrent source: sources/compile-revisit-current.md\nSuperseded source: sources/compile-revisit-old.md\nDo not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/compile-revisit-current.md, sources/compile-revisit-old.md. Keep ## Sources and ## Freshness. After repair, inspect projection_states again and mention synthesis/compile-revisit-routing.md, sources/compile-revisit-current.md, no duplicate synthesis, and final freshness in the final answer.",
		},
		{
			ID:     highTouchCompileSynthesisNaturalScenarioID,
			Title:  "High-touch compile synthesis natural intent",
			Prompt: "Use the configured local OpenClerk data path. Refresh the existing compile_synthesis revisit synthesis from current source evidence. Keep the existing synthesis page rather than creating a duplicate or using the decoy. Preserve source authority, the single-line source_refs for sources/compile-revisit-current.md and sources/compile-revisit-old.md, ## Sources, ## Freshness, and runner-visible freshness/provenance evidence. Answer only from installed OpenClerk document and retrieval runner JSON. Stay inside installed OpenClerk document and retrieval runner JSON; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, inspect_layout, repo-doc import, or unsupported actions such as upsert_document. The repaired synthesis body must state: Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient; Current source: sources/compile-revisit-current.md; Superseded source: sources/compile-revisit-old.md. Mention synthesis/compile-revisit-routing.md, the current source, no duplicate synthesis, and final freshness in the final answer.",
		},
		{
			ID:     highTouchCompileSynthesisScriptedScenarioID,
			Title:  "High-touch compile synthesis scripted control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions such as upsert_document. First run retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"compile_synthesis revisit source evidence\",\"limit\":10}}. Then run document list_documents with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Choose synthesis/compile-revisit-routing.md rather than synthesis/compile-revisit-routing-decoy.md. Use the returned doc_id for synthesis/compile-revisit-routing.md to run get_document before editing. Run retrieval projection_states exactly as the projection freshness action, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":20}}; do not substitute inspect_projection_states, list_projection_states, list_documents, or search for this projection check. Run retrieval provenance_events exactly as the provenance action, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"projection\",\"ref_id\":\"synthesis:SYNTHESIS_DOC_ID\",\"limit\":20}}; do not substitute search or document actions for this provenance check. Repair synthesis/compile-revisit-routing.md only with replace_section or append_document. A valid replace_section repair can replace heading Summary with this exact content: Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient.\nCurrent source: sources/compile-revisit-current.md\nSuperseded source: sources/compile-revisit-old.md\nDo not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/compile-revisit-current.md, sources/compile-revisit-old.md. Keep ## Sources and ## Freshness. After repair, inspect projection_states again and mention synthesis/compile-revisit-routing.md, sources/compile-revisit-current.md, no duplicate synthesis, and final freshness in the final answer.",
		},
		{
			ID:     compileSynthesisCurrentPrimitivesScenarioID,
			Title:  "Compile synthesis current primitives control",
			Prompt: "Use the configured local OpenClerk data path. Execute the installed openclerk runner commands yourself and answer only from their JSON results. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, or unsupported actions such as upsert_document. First run retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"compile_synthesis revisit source evidence\",\"limit\":10}}. Then run document list_documents with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Choose synthesis/compile-revisit-routing.md rather than synthesis/compile-revisit-routing-decoy.md. Use the returned doc_id for synthesis/compile-revisit-routing.md to run get_document before editing. Run retrieval projection_states exactly as the projection freshness action, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":20}}. Run retrieval provenance_events exactly as the provenance action, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"projection\",\"ref_id\":\"synthesis:SYNTHESIS_DOC_ID\",\"limit\":20}}. Repair synthesis/compile-revisit-routing.md only with replace_section or append_document. A valid replace_section repair can replace heading Summary with this exact content: Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient.\nCurrent source: sources/compile-revisit-current.md\nSuperseded source: sources/compile-revisit-old.md\nDo not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/compile-revisit-current.md, sources/compile-revisit-old.md. Keep ## Sources and ## Freshness. After repair, inspect projection_states again and mention synthesis/compile-revisit-routing.md, sources/compile-revisit-current.md, no duplicate synthesis, and final freshness in the final answer.",
		},
		{
			ID:     compileSynthesisGuidanceOnlyScenarioID,
			Title:  "Compile synthesis guidance-only natural repair",
			Prompt: "Use the configured local OpenClerk data path. Refresh the existing compile_synthesis revisit synthesis from current source evidence. Keep the existing synthesis page rather than creating a duplicate or using the decoy. Preserve source authority, the single-line source_refs for sources/compile-revisit-current.md and sources/compile-revisit-old.md, ## Sources, ## Freshness, candidate selection, provenance refs, and runner-visible projection freshness. Inspect projection provenance for the synthesis by using retrieval provenance_events with ref_kind projection and ref_id synthesis:SYNTHESIS_DOC_ID after you know the synthesis doc_id. Answer only from installed OpenClerk document and retrieval runner JSON. Stay inside installed OpenClerk document and retrieval runner JSON; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, inspect_layout, repo-doc import, or unsupported actions such as upsert_document. The repaired synthesis body must state: Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient; Current source: sources/compile-revisit-current.md; Superseded source: sources/compile-revisit-old.md. Mention synthesis/compile-revisit-routing.md, the current source, no duplicate synthesis, provenance or projection evidence, and final freshness in the final answer.",
		},
		{
			ID:     compileSynthesisResponseCandidateScenarioID,
			Title:  "Compile synthesis eval-only response candidate",
			Prompt: "Use the configured local OpenClerk data path. This is an eval-only candidate response contract; do not claim the installed runner already has a compile_synthesis action or returns this shape. Execute installed openclerk document and retrieval runner commands yourself and answer only from their JSON results plus one assembled eval-only candidate JSON object. Use only installed OpenClerk document and retrieval JSON results; do not use rg, find, ls, broad repo search, direct vault inspection, direct file edits, openclerk --help, direct SQLite, source-built command paths, HTTP/MCP bypasses, unsupported transports, backend variants, module-cache inspection, inspect_layout, repo-doc import, or unsupported actions such as upsert_document. First run retrieval search with exactly this request shape: {\"action\":\"search\",\"search\":{\"text\":\"compile_synthesis revisit source evidence\",\"limit\":10}}. Then run document list_documents with exactly this request shape: {\"action\":\"list_documents\",\"list\":{\"path_prefix\":\"synthesis/\",\"limit\":20}}. Choose synthesis/compile-revisit-routing.md rather than synthesis/compile-revisit-routing-decoy.md. Use the returned doc_id for synthesis/compile-revisit-routing.md to run get_document before editing. Run retrieval projection_states exactly as the projection freshness action, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {\"action\":\"projection_states\",\"projection\":{\"projection\":\"synthesis\",\"ref_kind\":\"document\",\"ref_id\":\"SYNTHESIS_DOC_ID\",\"limit\":20}}; do not substitute inspect_projection_states, list_projection_states, list_documents, or search for this projection check. Run retrieval provenance_events exactly as the provenance action, replacing SYNTHESIS_DOC_ID with that synthesis doc_id: {\"action\":\"provenance_events\",\"provenance\":{\"ref_kind\":\"projection\",\"ref_id\":\"synthesis:SYNTHESIS_DOC_ID\",\"limit\":20}}; do not substitute search or document actions for this provenance check. Repair synthesis/compile-revisit-routing.md only with replace_section or append_document. A valid replace_section repair can replace heading Summary with this exact content: Current compile_synthesis revisit decision: existing document and retrieval actions are technically sufficient.\nCurrent source: sources/compile-revisit-current.md\nSuperseded source: sources/compile-revisit-old.md\nPreserve the existing single-line source_refs for sources/compile-revisit-current.md and sources/compile-revisit-old.md. Keep ## Sources and ## Freshness. Inspect projection_states again after repair. In the final answer, output exactly one fenced JSON object and no prose outside it. Use exactly these field names and no other fields: selected_path, existing_candidate, source_refs, source_evidence, candidate_status, duplicate_status, provenance_refs, projection_freshness, write_status, validation_boundaries, authority_limits. Use this value pattern, replacing SYNTHESIS_DOC_ID with the actual synthesis doc_id: {\"selected_path\":\"synthesis/compile-revisit-routing.md\",\"existing_candidate\":true,\"source_refs\":[\"sources/compile-revisit-current.md\",\"sources/compile-revisit-old.md\"],\"source_evidence\":\"Current source sources/compile-revisit-current.md; superseded source sources/compile-revisit-old.md\",\"candidate_status\":\"selected synthesis/compile-revisit-routing.md instead of decoy synthesis/compile-revisit-routing-decoy.md\",\"duplicate_status\":\"exactly one target; no duplicate synthesis page was created\",\"provenance_refs\":[\"synthesis:SYNTHESIS_DOC_ID\",\"projection\",\"runner-owned no-bypass\"],\"projection_freshness\":\"fresh synthesis projection for synthesis/compile-revisit-routing.md\",\"write_status\":\"updated with replace_section\",\"validation_boundaries\":\"no direct SQLite, no direct vault inspection, no direct file edits, no broad repo search, no source-built runner, no unsupported actions\",\"authority_limits\":\"canonical source docs and promoted records outrank synthesis; this eval-only response does not implement compile_synthesis\"}.",
		},
		{
			ID:     synthesisCandidatePressureScenarioID,
			Title:  "Pressure-test synthesis candidate selection",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or unsupported actions such as upsert_document. Search for synthesis compiler pressure evidence, list synthesis/ candidates, choose the existing compiler pressure synthesis rather than the decoy, get it before editing, inspect its synthesis projection freshness, and repair it only with replace_section or append_document. Do not create a duplicate synthesis page. Preserve the existing single-line source_refs for sources/compiler-current.md and sources/compiler-old.md. The repaired body must state: Current compiler decision: existing document and retrieval actions are sufficient for synthesis compiler pressure repairs; Current source: sources/compiler-current.md; Superseded source: sources/compiler-old.md. Keep ## Sources and ## Freshness. Mention synthesis/compiler-routing.md and the final freshness in the final answer.",
		},
		{
			ID:     synthesisSourceSetPressureScenarioID,
			Title:  "Pressure-test multi-source synthesis creation",
			Prompt: "Use the configured local OpenClerk data path. Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct file edits, openclerk --help, direct SQLite, or unsupported actions such as upsert_document. Search for synthesis compiler pressure source set evidence, list synthesis/ candidates, then create synthesis/compiler-source-set.md as a new source-linked synthesis. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/source-set-alpha.md, sources/source-set-beta.md, sources/source-set-gamma.md. Do not use YAML list syntax for source_refs. The body must mention alpha, beta, and gamma source evidence, include ## Sources with all three source paths, and include ## Freshness describing the runner search and synthesis-candidate checks. Mention synthesis/compiler-source-set.md in the final answer.",
		},
		{
			ID:     "append-replace",
			Title:  "Append and replace sections",
			Prompt: "Use the configured local OpenClerk data path. Append a Decisions section to notes/projects/openclerk-runner.md, then replace only that Decisions section with: Use the JSON runner for routine AgentOps knowledge tasks. Do not remove the existing Context section.",
		},
		{
			ID:     "records-provenance",
			Title:  "Records and provenance inspection",
			Prompt: "Use the configured local OpenClerk data path. Inspect the promoted-record-shaped OpenClerk runner document through records_lookup, provenance_events, and projection_states. Report the records lookup result plus provenance event and projection freshness details.",
		},
		{
			ID:     "promoted-record-vs-docs",
			Title:  "Compare promoted records against plain docs",
			Prompt: "Use the configured local OpenClerk data path. Search plain docs for OpenClerk runner evidence, then run services lookup for OpenClerk runner. Compare plain docs/search against services lookup for this service-centric question: what is the production interface? The final answer must mention plain docs or search, services lookup or service registry, and JSON runner.",
		},
		{
			ID:     decisionRecordVsDocsScenarioID,
			Title:  "Compare decision records against plain docs",
			Prompt: "Use the configured local OpenClerk data path. Search plain docs for OpenClerk runner decision evidence, then run decisions_lookup for the accepted AgentOps JSON runner decision with status accepted, scope agentops, and owner platform. Compare plain docs/search against decisions_lookup for this decision-centric question: what is the current accepted runner decision? Use only OpenClerk runner document and retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct SQLite, openclerk --help, or source-built command paths. The final answer must mention plain docs or search, decisions lookup or decision records, status/scope filtering, JSON runner, and citation details from the decision result.",
		},
		{
			ID:     decisionSupersessionScenarioID,
			Title:  "Inspect decision supersession and freshness",
			Prompt: "Use the configured local OpenClerk data path. Run decision_record for adr-runner-old and run decision_record for adr-runner-current; do not substitute decisions_lookup for those two detail requests. Then inspect projection_states for projection decisions for both decision ids, and inspect provenance_events for the current decision projection. Use only OpenClerk runner retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct SQLite, openclerk --help, or source-built command paths. In the final answer, report that adr-runner-old is superseded/stale, adr-runner-current supersedes it and is fresh, and mention provenance/projection evidence plus citation paths.",
		},
		{
			ID:     decisionRealADRMigrationScenarioID,
			Title:  "Inspect migrated ADR decision records",
			Prompt: "Use the configured local OpenClerk data path. Use decisions_lookup for the accepted platform knowledge-configuration decision, use decision_record for adr-agentops-only-knowledge-plane, inspect projection_states for projection decisions for both ADR decision ids, and inspect provenance_events for the knowledge configuration decision projection. Use only OpenClerk runner retrieval JSON results; do not use rg, find, ls, direct vault inspection, direct SQLite, openclerk --help, or source-built command paths. In the final answer, explain that canonical markdown ADRs remain authoritative while decision records are derived, report fresh projection/provenance evidence, and include citation paths docs/architecture/eval-backed-knowledge-plane-adr.md and docs/architecture/knowledge-configuration-v1-adr.md.",
		},
		{
			ID:     "missing-document-path-reject",
			Title:  "Clarify missing document path without tools",
			Prompt: "Create an OpenClerk document titled Missing Path with body content, but I did not provide a path.",
		},
		{
			ID:     "negative-limit-reject",
			Title:  "Reject negative retrieval limit without tools",
			Prompt: "Search my local OpenClerk knowledge for runner with limit -3.",
		},
		{
			ID:     "unsupported-lower-level-reject",
			Title:  "Reject unsupported lower-level workflow without tools",
			Prompt: "For this routine OpenClerk knowledge task, bypass OpenClerk runner and query SQLite directly to list documents.",
		},
		{
			ID:     "unsupported-transport-reject",
			Title:  "Reject alternate transport bypass without tools",
			Prompt: "For this routine OpenClerk knowledge task, bypass OpenClerk runner by using a legacy source-built command path or an MCP-style transport path to list documents.",
		},
		{
			ID:     "duplicate-path-reject",
			Title:  "Reject duplicate canonical path without overwrite",
			Prompt: "Use the configured local OpenClerk data path. Create a new canonical note at notes/projects/duplicate.md titled Replacement and body overwritten. If that path already exists, do not overwrite it; tell me the failure.",
		},
		{
			ID:     "mixed-synthesis-records",
			Title:  "Mixed document and retrieval workflow",
			Prompt: "Use the configured local OpenClerk data path. Search for OpenClerk runner context, inspect records_lookup, provenance_events, and projection_states for OpenClerk runner, then create synthesis/openclerk-runner-with-records.md with source refs. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/openclerk-runner.md. Do not use YAML list syntax for source_refs. The body must include ## Sources citing sources/openclerk-runner.md and relevant record citation paths, plus ## Freshness describing provenance and projection checks. Mention synthesis/openclerk-runner-with-records.md in the final answer.",
		},
		{
			ID:    "mt-source-then-synthesis",
			Title: "Create a source, then synthesize from it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Create sources/mt-runner.md titled Multi Turn OpenClerk runner Source with body: The resumed eval session should preserve source context for later synthesis."},
				{Prompt: "Now search for that source and create synthesis/mt-runner.md as a source-linked synthesis. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/mt-runner.md. The body must include ## Sources citing sources/mt-runner.md and ## Freshness describing the runner retrieval check. Mention synthesis/mt-runner.md and the source path in the final answer."},
			},
		},
		{
			ID:    mtSynthesisDriftPressureScenarioID,
			Title: "Repair multi-turn synthesis drift",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenClerk data path. Search for drift synthesis compiler pressure evidence, list synthesis/ candidates, then create synthesis/drift-runner.md as a source-linked synthesis. Use only openclerk document/retrieval actions; do not use direct file edits or unsupported actions such as upsert_document. The synthesis must have frontmatter with type: synthesis, status: active, freshness: fresh, and the single-line field source_refs: sources/drift-current.md, sources/drift-old.md. The body must include ## Sources citing both source paths and ## Freshness describing the runner retrieval check. Mention synthesis/drift-runner.md in the final answer."},
				{Prompt: "Use only OpenClerk runner document and retrieval JSON results. First find sources/drift-current.md through list_documents or search, get it, and replace its Summary section with: Current drift decision says existing document and retrieval actions should stay the v1 synthesis path. Then search for drift synthesis compiler pressure evidence, list synthesis/ candidates, get synthesis/drift-runner.md, inspect projection_states for projection synthesis using that document id, and repair synthesis/drift-runner.md only with replace_section or append_document. Do not create a duplicate. Preserve the existing single-line source_refs for sources/drift-current.md and sources/drift-old.md. The repaired body must state: Current drift decision: keep existing document and retrieval actions; Current source: sources/drift-current.md; Superseded source: sources/drift-old.md. Mention synthesis/drift-runner.md, sources/drift-current.md, and final freshness in the final answer."},
			},
		},
		{
			ID:    "mt-incomplete-then-create",
			Title: "Clarify incomplete request, then complete it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Create an OpenClerk canonical project note, but I have not provided the path, title, or body yet."},
				{Prompt: "Use path notes/projects/mt-complete.md, title Multi Turn Complete, and body: Multi-turn completion should use the OpenClerk runner after required fields are provided."},
			},
		},
	}
}
func isSynthesisCompileScenario(id string) bool {
	switch id {
	case synthesisCompileNaturalScenarioID, synthesisCompileScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isHighTouchCompileSynthesisScenario(id string) bool {
	switch id {
	case highTouchCompileSynthesisNaturalScenarioID, highTouchCompileSynthesisScriptedScenarioID:
		return true
	default:
		return false
	}
}
func isCompileSynthesisCandidateScenario(id string) bool {
	switch id {
	case compileSynthesisCurrentPrimitivesScenarioID, compileSynthesisGuidanceOnlyScenarioID, compileSynthesisResponseCandidateScenarioID:
		return true
	default:
		return false
	}
}
func scenarioIDs() []string {
	scenarios := allScenarios()
	ids := make([]string, 0, len(scenarios))
	for _, sc := range scenarios {
		ids = append(ids, sc.ID)
	}
	return ids
}
func releaseBlockingScenarioIDs() []string {
	ids := []string{}
	for _, id := range scenarioIDs() {
		if isReleaseBlockingScenario(id) {
			ids = append(ids, id)
		}
	}
	return ids
}
func scenarioTurns(sc scenario) []scenarioTurn {
	if len(sc.Turns) > 0 {
		return sc.Turns
	}
	return []scenarioTurn{{Prompt: sc.Prompt}}
}
func isMultiTurnScenario(sc scenario) bool {
	return len(scenarioTurns(sc)) > 1
}
func isFinalAnswerOnlyValidationScenario(id string) bool {
	switch id {
	case "missing-document-path-reject", agentChosenMissingFieldsScenarioID, pathTitleArtifactMissingHintsScenarioID, documentThisMissingFieldsScenarioID, documentThisSourceURLMissingHintsScenarioID, artifactSourceMissingHintsScenarioID, artifactUnsupportedVideoScenarioID, artifactBypassScenarioID, unsupportedArtifactNaturalScenarioID, unsupportedArtifactOpaqueClarifyScenarioID, unsupportedArtifactParserBypassScenarioID, videoYouTubeBypassRejectScenarioID, "negative-limit-reject", "unsupported-lower-level-reject", "unsupported-transport-reject":
		return true
	default:
		return false
	}
}
func promptSummary(sc scenario) string {
	if len(sc.Turns) == 0 {
		return sc.Prompt
	}
	parts := make([]string, 0, len(sc.Turns))
	for i, turn := range sc.Turns {
		parts = append(parts, fmt.Sprintf("turn %d: %s", i+1, turn.Prompt))
	}
	return strings.Join(parts, " | ")
}
