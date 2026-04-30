package main

func reportLane(ids []string) (string, bool) {
	if len(ids) == 0 {
		return populatedDefaultLaneName, true
	}
	populated := 0
	repoDocs := 0
	graphSemanticsRevisit := 0
	memoryRouterRevisit := 0
	promotedRecordDomain := 0
	parallelRunner := 0
	documentHistory := 0
	agentChosenPath := 0
	pathTitleAutonomy := 0
	captureExplicitOverrides := 0
	captureDuplicateCandidate := 0
	sourceURLUpdate := 0
	webURLIntake := 0
	documentThis := 0
	documentArtifactCandidate := 0
	artifactIngestion := 0
	videoYouTube := 0
	synthesisCompile := 0
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
		if isPromotedRecordDomainScenario(id) {
			promotedRecordDomain++
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
		if isAgentChosenPathScenario(id) {
			agentChosenPath++
			continue
		}
		if isPathTitleAutonomyScenario(id) {
			pathTitleAutonomy++
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
		if isSourceURLUpdateScenario(id) {
			sourceURLUpdate++
			continue
		}
		if isWebURLIntakeScenario(id) {
			webURLIntake++
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
	if promotedRecordDomain > 0 && promotedRecordDomain+validation == len(ids) {
		return promotedRecordDomainLaneName, false
	}
	if parallelRunner > 0 && parallelRunner == len(ids) {
		return parallelRunnerLaneName, false
	}
	if documentHistory > 0 && documentHistory+validation == len(ids) {
		return documentHistoryLaneName, false
	}
	if agentChosenPath > 0 && agentChosenPath+validation == len(ids) {
		return agentChosenPathLaneName, false
	}
	if pathTitleAutonomy > 0 && pathTitleAutonomy == len(ids) {
		return pathTitleAutonomyLaneName, false
	}
	if captureExplicitOverrides > 0 && captureExplicitOverrides+validation == len(ids) {
		return captureExplicitOverridesLaneName, false
	}
	if captureDuplicateCandidate > 0 && captureDuplicateCandidate+validation == len(ids) {
		return captureDuplicateCandidateLaneName, false
	}
	if sourceURLUpdate > 0 && sourceURLUpdate+validation == len(ids) {
		return sourceURLUpdateLaneName, false
	}
	if webURLIntake > 0 && webURLIntake+validation == len(ids) {
		return webURLIntakeLaneName, false
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
	if promotedRecordDomain > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if parallelRunner > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if documentHistory > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if agentChosenPath > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if pathTitleAutonomy > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if captureExplicitOverrides > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if captureDuplicateCandidate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if sourceURLUpdate > 0 {
		return populatedMixedLaneName, releaseBlocking
	}
	if webURLIntake > 0 {
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
	if lane == promotedRecordDomainLaneName {
		return "promoted record domain expansion rows report natural record-domain intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == parallelRunnerLaneName {
		return "parallel runner rows report fresh startup and safe-read command UX, tool count, command count, assistant calls, wall time, guidance dependence, safety risks, and raw SQLite/runtime_config/upsert failure absence"
	}
	if lane == documentHistoryLaneName {
		return "document lifecycle rows report natural intent, scripted current-primitives controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, privacy handling, and capability/ergonomics classification"
	}
	if lane == documentArtifactCandidateLaneName {
		return "document artifact candidate rows report candidate quality plus ergonomics scorecard fields: tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and final classification"
	}
	if lane == captureExplicitOverridesLaneName {
		return "explicit-overrides capture rows report natural explicit override intent, scripted validation control, invalid explicit value rejection, authority conflict handling, no convention override, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == captureDuplicateCandidateLaneName {
		return "duplicate-candidate capture rows report runner-visible search/list/get evidence, update-versus-new-path clarification, target accuracy, no duplicate write, approval-before-write, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == artifactIngestionLaneName {
		return "artifact ingestion rows report tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, fixture preflight, and final classification"
	}
	if lane == webURLIntakeLaneName {
		return "web URL intake rows report missing path-hint handling, web create, duplicate URL rejection, no-op update, changed-source stale synthesis evidence, unsupported acquisition rejection, and final classification"
	}
	if lane == videoYouTubeLaneName {
		return "video/YouTube rows report natural supplied-transcript intent, scripted transcript control, synthesis freshness, bypass rejection, ergonomics scorecard fields, and final capability classification"
	}
	if lane == synthesisCompileLaneName {
		return "synthesis compile revisit rows report natural compile intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	if lane == broadAuditLaneName {
		return "broad contradiction/audit revisit rows report natural audit intent, scripted current-primitives control, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and capability/ergonomics classification"
	}
	return ""
}
