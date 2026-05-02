package main

func reportLane(ids []string) (string, bool) {
	if len(ids) == 0 {
		return populatedDefaultLaneName, true
	}
	populated := 0
	repoDocs := 0
	graphSemanticsRevisit := 0
	memoryRouterRevisit := 0
	highTouchMemoryRouterRecall := 0
	memoryRouterRecallCandidate := 0
	promotedRecordDomain := 0
	highTouchRelationshipRecord := 0
	relationshipRecordCandidate := 0
	parallelRunner := 0
	documentHistory := 0
	highTouchDocumentLifecycle := 0
	documentLifecycleRollbackCandidate := 0
	agentChosenPath := 0
	pathTitleAutonomy := 0
	captureLowRisk := 0
	captureExplicitOverrides := 0
	captureDuplicateCandidate := 0
	tagging := 0
	captureSaveThisNote := 0
	captureDocumentLinks := 0
	sourceURLUpdate := 0
	webURLIntake := 0
	webURLStaleRepair := 0
	webURLStaleImpact := 0
	webProductPage := 0
	documentThis := 0
	documentArtifactCandidate := 0
	artifactIngestion := 0
	videoYouTube := 0
	synthesisCompile := 0
	highTouchCompileSynthesis := 0
	compileSynthesisCandidate := 0
	broadAudit := 0
	validation := 0
	releaseBlocking := false
	for _, id := range ids {
		if isPopulatedVaultScenario(id) {
			populated++
			continue
		}
		if isRepoDocsDogfoodScenario(id) {
			repoDocs++
			continue
		}
		if isGraphSemanticsRevisitScenario(id) {
			graphSemanticsRevisit++
			continue
		}
		if isMemoryRouterRevisitScenario(id) {
			memoryRouterRevisit++
			continue
		}
		if isHighTouchMemoryRouterRecallScenario(id) {
			highTouchMemoryRouterRecall++
			continue
		}
		if isMemoryRouterRecallCandidateScenario(id) {
			memoryRouterRecallCandidate++
			continue
		}
		if isPromotedRecordDomainScenario(id) {
			promotedRecordDomain++
			continue
		}
		if isHighTouchRelationshipRecordScenario(id) {
			highTouchRelationshipRecord++
			continue
		}
		if isRelationshipRecordCandidateScenario(id) {
			relationshipRecordCandidate++
			continue
		}
		if isParallelRunnerScenario(id) {
			parallelRunner++
			continue
		}
		if isDocumentHistoryScenario(id) {
			documentHistory++
			continue
		}
		if isHighTouchDocumentLifecycleScenario(id) {
			highTouchDocumentLifecycle++
			continue
		}
		if isDocumentLifecycleRollbackCandidateScenario(id) {
			documentLifecycleRollbackCandidate++
			continue
		}
		if isAgentChosenPathScenario(id) {
			agentChosenPath++
			continue
		}
		if isPathTitleAutonomyScenario(id) {
			pathTitleAutonomy++
			continue
		}
		if isCaptureLowRiskScenario(id) {
			captureLowRisk++
			continue
		}
		if isCaptureExplicitOverridesScenario(id) {
			captureExplicitOverrides++
			continue
		}
		if isCaptureDuplicateCandidateScenario(id) {
			captureDuplicateCandidate++
			continue
		}
		if isTaggingScenario(id) {
			tagging++
			continue
		}
		if isCaptureSaveThisNoteScenario(id) {
			captureSaveThisNote++
			continue
		}
		if isCaptureDocumentLinksScenario(id) {
			captureDocumentLinks++
			continue
		}
		if isSourceURLUpdateScenario(id) {
			sourceURLUpdate++
			continue
		}
		if isWebURLIntakeScenario(id) {
			webURLIntake++
			continue
		}
		if isWebURLStaleRepairScenario(id) {
			webURLStaleRepair++
			continue
		}
		if isWebURLStaleImpactScenario(id) {
			webURLStaleImpact++
			continue
		}
		if isWebProductPageScenario(id) {
			webProductPage++
			continue
		}
		if isDocumentThisScenario(id) {
			documentThis++
			continue
		}
		if isDocumentArtifactCandidateScenario(id) {
			documentArtifactCandidate++
			continue
		}
		if isArtifactIngestionScenario(id) {
			artifactIngestion++
			continue
		}
		if isVideoYouTubeScenario(id) {
			videoYouTube++
			continue
		}
		if isSynthesisCompileScenario(id) {
			synthesisCompile++
			continue
		}
		if isHighTouchCompileSynthesisScenario(id) {
			highTouchCompileSynthesis++
			continue
		}
		if isCompileSynthesisCandidateScenario(id) {
			compileSynthesisCandidate++
			continue
		}
		if isBroadAuditScenario(id) {
			broadAudit++
			continue
		}
		if isFinalAnswerOnlyValidationScenario(id) {
			validation++
			continue
		}
		releaseBlocking = true
	}
	if populated == len(ids) {
		return populatedLaneName, false
	}
	if repoDocs == len(ids) {
		return repoDocsLaneName, false
	}
	if graphSemanticsRevisit > 0 && graphSemanticsRevisit+validation == len(ids) {
		return graphSemanticsRevisitLaneName, false
	}
	if memoryRouterRevisit > 0 && memoryRouterRevisit+validation == len(ids) {
		return memoryRouterRevisitLaneName, false
	}
	if highTouchMemoryRouterRecall > 0 && highTouchMemoryRouterRecall+validation == len(ids) {
		return highTouchMemoryRouterRecallLaneName, false
	}
	if memoryRouterRecallCandidate > 0 && memoryRouterRecallCandidate+validation == len(ids) {
		return memoryRouterRecallCandidateLaneName, false
	}
	if promotedRecordDomain > 0 && promotedRecordDomain+validation == len(ids) {
		return promotedRecordDomainLaneName, false
	}
	if highTouchRelationshipRecord > 0 && highTouchRelationshipRecord+validation == len(ids) {
		return highTouchRelationshipRecordLaneName, false
	}
	if relationshipRecordCandidate > 0 && relationshipRecordCandidate+validation == len(ids) {
		return relationshipRecordCandidateLaneName, false
	}
	if parallelRunner > 0 && parallelRunner == len(ids) {
		return parallelRunnerLaneName, false
	}
	if documentHistory > 0 && documentHistory+validation == len(ids) {
		return documentHistoryLaneName, false
	}
	if highTouchDocumentLifecycle > 0 && highTouchDocumentLifecycle+validation == len(ids) {
		return highTouchDocumentLifecycleLaneName, false
	}
	if documentLifecycleRollbackCandidate > 0 && documentLifecycleRollbackCandidate+validation == len(ids) {
		return documentLifecycleRollbackCandidateLaneName, false
	}
	if agentChosenPath > 0 && agentChosenPath+validation == len(ids) {
		return agentChosenPathLaneName, false
	}
	if pathTitleAutonomy > 0 && pathTitleAutonomy == len(ids) {
		return pathTitleAutonomyLaneName, false
	}
	if captureLowRisk > 0 && captureLowRisk+validation == len(ids) {
		return captureLowRiskLaneName, false
	}
	if captureExplicitOverrides > 0 && captureExplicitOverrides+validation == len(ids) {
		return captureExplicitOverridesLaneName, false
	}
	if captureDuplicateCandidate > 0 && captureDuplicateCandidate+validation == len(ids) {
		return captureDuplicateCandidateLaneName, false
	}
	if tagging > 0 && tagging+validation == len(ids) {
		return taggingLaneName, false
	}
	if captureSaveThisNote > 0 && captureSaveThisNote+validation == len(ids) {
		return captureSaveThisNoteLaneName, false
	}
	if captureDocumentLinks > 0 && captureDocumentLinks+validation == len(ids) {
		return captureDocumentLinksLaneName, false
	}
	if sourceURLUpdate > 0 && sourceURLUpdate+validation == len(ids) {
		return sourceURLUpdateLaneName, false
	}
	if webURLIntake > 0 && webURLIntake+validation == len(ids) {
		return webURLIntakeLaneName, false
	}
	if webURLStaleRepair > 0 && webURLStaleRepair+validation == len(ids) {
		return webURLStaleRepairLaneName, false
	}
	if webURLStaleImpact > 0 && webURLStaleImpact+validation == len(ids) {
		return webURLStaleImpactLaneName, false
	}
	if webProductPage > 0 && webProductPage+validation == len(ids) {
		return webProductPageLaneName, false
	}
	if documentThis > 0 && documentThis == len(ids) {
		return documentThisLaneName, false
	}
	if documentArtifactCandidate > 0 && documentArtifactCandidate == len(ids) {
		return documentArtifactCandidateLaneName, false
	}
	if artifactIngestion > 0 && artifactIngestion == len(ids) {
		return artifactIngestionLaneName, false
	}
	if videoYouTube > 0 && videoYouTube == len(ids) {
		return videoYouTubeLaneName, false
	}
	if synthesisCompile > 0 && synthesisCompile+validation == len(ids) {
		return synthesisCompileLaneName, false
	}
	if highTouchCompileSynthesis > 0 && highTouchCompileSynthesis+validation == len(ids) {
		return highTouchCompileSynthesisLaneName, false
	}
	if compileSynthesisCandidate > 0 && compileSynthesisCandidate+validation == len(ids) {
		return compileSynthesisCandidateLaneName, false
	}
	if broadAudit > 0 && broadAudit+validation == len(ids) {
		return broadAuditLaneName, false
	}
	if populated > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if repoDocs > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if graphSemanticsRevisit > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if memoryRouterRevisit > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if highTouchMemoryRouterRecall > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if memoryRouterRecallCandidate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if promotedRecordDomain > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if highTouchRelationshipRecord > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if relationshipRecordCandidate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if parallelRunner > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if documentHistory > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if highTouchDocumentLifecycle > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if documentLifecycleRollbackCandidate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if agentChosenPath > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if pathTitleAutonomy > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if captureLowRisk > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if captureExplicitOverrides > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if captureDuplicateCandidate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if tagging > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if captureSaveThisNote > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if captureDocumentLinks > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if sourceURLUpdate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if webURLIntake > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if webURLStaleRepair > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if webURLStaleImpact > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if webProductPage > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if documentThis > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if documentArtifactCandidate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if artifactIngestion > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if videoYouTube > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if synthesisCompile > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if highTouchCompileSynthesis > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if compileSynthesisCandidate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if broadAudit > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	return populatedDefaultLaneName, true
}

func targetedAcceptanceNote(lane string) string {
	if lane == repoDocsLaneName {
		return "repo-docs dogfood rows import committed public markdown into an isolated eval vault and report retrieval, synthesis, and decision-record behavior without private vault evidence"
	}
	if lane == graphSemanticsRevisitLaneName {
		return "graph semantics revisit rows report natural relationship intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == memoryRouterRevisitLaneName {
		return "memory and autonomous router revisit rows report natural memory/router intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == highTouchMemoryRouterRecallLaneName {
		return "high-touch memory/router recall ceremony rows report natural temporal recall and routing intent, scripted current-primitives control, canonical markdown memory authority, current canonical docs over stale session observations, advisory feedback weighting, routing rationale, source refs, provenance, synthesis freshness, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, and separate safety/capability/UX classification"
	}
	if lane == memoryRouterRecallCandidateLaneName {
		return "memory/router recall candidate rows compare current primitives, guidance-only repair, and an eval-only candidate response contract, while reporting query summary, temporal status, canonical evidence refs, stale session status, feedback weighting, routing rationale, provenance refs, synthesis freshness, validation/no-bypass boundaries, authority limits, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, safety pass, capability pass, and UX quality"
	}
	if lane == promotedRecordDomainLaneName {
		return "promoted record domain expansion rows report natural record-domain intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == highTouchRelationshipRecordLaneName {
		return "high-touch relationship-record ceremony rows report natural combined relationship and record lookup intent, scripted current-primitives control, canonical markdown relationship authority, links/backlinks, graph freshness, record citations, provenance, records freshness, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, and separate safety/capability/UX classification"
	}
	if lane == relationshipRecordCandidateLaneName {
		return "relationship-record lookup candidate rows compare current primitives, guidance-only repair, and an eval-only candidate response contract, while reporting query summary, relationship evidence, link/backlink evidence, graph freshness, record lookup/entity evidence, citation refs, provenance refs, records freshness, validation/no-bypass boundaries, authority limits, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, safety pass, capability pass, and UX quality"
	}
	if lane == parallelRunnerLaneName {
		return "parallel runner rows report fresh startup and safe-read command UX, tool count, command count, assistant calls, wall time, guidance dependence, safety risks, and raw SQLite/runtime_config/upsert failure absence"
	}
	if lane == documentHistoryLaneName {
		return "document lifecycle rows report natural intent, scripted current-primitives controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, privacy handling, and capability/ergonomics classification"
	}
	if lane == highTouchDocumentLifecycleLaneName {
		return "high-touch document lifecycle ceremony rows report natural lifecycle review and rollback intent, scripted history/provenance/freshness control, rollback target accuracy, privacy-safe summaries, no raw private diffs in committed artifacts, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, and separate safety/capability/UX classification"
	}
	if lane == documentLifecycleRollbackCandidateLaneName {
		return "document lifecycle rollback candidate rows compare current primitives, guidance-only repair, and an eval-only candidate response contract, while reporting target identity, source evidence, before/after summaries, restore reason, provenance refs, projection freshness, write status, privacy/no-diff boundaries, validation/no-bypass boundaries, authority limits, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, safety pass, capability pass, and UX quality"
	}
	if lane == documentArtifactCandidateLaneName {
		return "document artifact candidate rows report candidate quality plus ergonomics scorecard fields: tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and final classification"
	}
	if lane == captureLowRiskLaneName {
		return "low-risk capture rows report natural low-risk save intent, scripted candidate validation control, duplicate checks, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == captureExplicitOverridesLaneName {
		return "explicit-overrides capture rows report natural explicit override intent, scripted validation control, invalid explicit value rejection, authority conflict handling, no convention override, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == captureDuplicateCandidateLaneName {
		return "duplicate-candidate capture rows report runner-visible search/list/get evidence, update-versus-new-path clarification, target accuracy, no duplicate write, approval-before-write, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == taggingLaneName {
		return "tagging rows report tagged create/update, retrieval by tag, exact tag disambiguation, near-duplicate tag exclusion, mixed path-plus-tag queries, metadata_key/metadata_value ceremony, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and separate safety/capability/UX classification"
	}
	if lane == captureSaveThisNoteLaneName {
		return "save-this-note capture rows report natural save intent, scripted candidate validation control, duplicate checks, low-confidence clarification, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == captureDocumentLinksLaneName {
		return "document-these-links placement rows report natural public-link placement intent, approved source fetch control, synthesis placement proposal, duplicate source/synthesis handling, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == artifactIngestionLaneName {
		return "artifact ingestion rows report tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, fixture preflight, and final classification"
	}
	if lane == webURLIntakeLaneName {
		return "web URL intake rows report missing path-hint handling, web create, duplicate URL rejection, no-op update, changed-source stale synthesis evidence, unsupported acquisition rejection, and final classification"
	}
	if lane == webURLStaleRepairLaneName {
		return "high-touch web URL stale repair rows report natural refresh intent, scripted update-mode control, duplicate/no-op behavior, changed-source freshness evidence, dependent synthesis stale visibility, provenance/freshness, no browser/manual acquisition, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and separate safety/capability/UX classification"
	}
	if lane == webURLStaleImpactLaneName {
		return "web URL stale-impact update response rows compare current primitives, guidance-only repair, and an eval-only candidate response contract, while reporting duplicate/no-op behavior, changed hash evidence, stale dependent synthesis refs, projection/provenance refs, no-repair warnings, no browser/manual acquisition, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, safety pass, capability pass, and UX quality"
	}
	if lane == webProductPageLaneName {
		return "rich public product-page rows report natural product-page intent, approved public HTML fetch control, tracking/variant duplicate normalization, visible text fidelity, dynamic omission disclosure, blocked or non-HTML rejection, no-browser/no-login/no-cart/no-checkout/no-purchase controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == videoYouTubeLaneName {
		return "video/YouTube rows report natural supplied-transcript intent, scripted transcript control, synthesis freshness, bypass rejection, ergonomics scorecard fields, and final capability classification"
	}
	if lane == synthesisCompileLaneName {
		return "synthesis compile revisit rows report natural compile intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == highTouchCompileSynthesisLaneName {
		return "high-touch compile synthesis ceremony rows report natural source-backed synthesis maintenance, scripted current-primitives control, source refs, Sources and Freshness sections, duplicate prevention, freshness/provenance visibility, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, and separate safety/capability/UX classification"
	}
	if lane == compileSynthesisCandidateLaneName {
		return "compile synthesis candidate rows compare current primitives, guidance-only repair, and an eval-only candidate response contract, while reporting selected path, source refs, source evidence, candidate/duplicate status, provenance refs, projection freshness, write status, validation/no-bypass boundaries, authority limits, tool count, command count, assistant calls, wall time, prompt specificity, retries, latency, brittleness, guidance dependence, safety risks, safety pass, capability pass, and UX quality"
	}
	if lane == broadAuditLaneName {
		return "broad contradiction/audit revisit rows report natural audit intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	return ""
}
