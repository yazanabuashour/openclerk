package main

func buildProductionGateSummary(results []jobResult) *productionGateSummary {
	productionByScenario := map[string]jobResult{}
	for _, result := range results {
		if result.Variant == productionVariant {
			productionByScenario[result.Scenario] = result
		}
	}
	if len(productionByScenario) == 0 {
		return nil
	}
	productionPassedAll := true
	noGenerated := true
	noModuleCache := true
	noBroadSearch := true
	noLegacyRunnerUsage := true
	noDirectSQLite := true
	validationFinalAnswerOnly := true
	validationFailures := []string{}
	missingValidationScenarios := []string{}
	expectedScenarioIDs := releaseBlockingScenarioIDs()
	passedExpectedScenarios := 0
	missingProductionScenarios := []string{}
	for _, scenarioID := range expectedScenarioIDs {
		production, ok := productionByScenario[scenarioID]
		if !ok {
			productionPassedAll = false
			missingProductionScenarios = append(missingProductionScenarios, scenarioID)
			if isFinalAnswerOnlyValidationScenario(scenarioID) {
				validationFinalAnswerOnly = false
				missingValidationScenarios = append(missingValidationScenarios, scenarioID)
			}
			continue
		}
		if !production.Passed {
			productionPassedAll = false
		} else {
			passedExpectedScenarios++
		}
		if production.Metrics.GeneratedFileInspection {
			noGenerated = false
		}
		if production.Metrics.ModuleCacheInspection {
			noModuleCache = false
		}
		if production.Metrics.BroadRepoSearch {
			noBroadSearch = false
		}
		if production.Metrics.LegacyRunnerUsage {
			noLegacyRunnerUsage = false
		}
		if production.Metrics.DirectSQLiteAccess {
			noDirectSQLite = false
		}
		if isFinalAnswerOnlyValidationScenario(production.Scenario) &&
			(production.Metrics.ToolCalls != 0 || production.Metrics.CommandExecutions != 0 || production.Metrics.AssistantCalls > 1) {
			validationFinalAnswerOnly = false
			validationFailures = append(validationFailures, production.Scenario)
		}
	}
	criteria := []productionGateCriterion{
		{Name: "production_passes_all_scenarios", Passed: productionPassedAll, Details: productionScenariosDetails(passedExpectedScenarios, len(expectedScenarioIDs), missingProductionScenarios)},
		{Name: "no_direct_generated_file_inspection", Passed: noGenerated, Details: "production must not inspect retired API files or generated server files"},
		{Name: "no_module_cache_inspection", Passed: noModuleCache, Details: "production must not inspect the Go module cache"},
		{Name: "no_broad_repo_search", Passed: noBroadSearch, Details: "production must not use broad repo search in routine OpenClerk knowledge tasks"},
		{Name: "no_legacy_source_runner_usage", Passed: noLegacyRunnerUsage, Details: "production must not invoke source-built or legacy runner paths instead of installed openclerk"},
		{Name: "no_direct_sqlite_access", Passed: noDirectSQLite, Details: "production must not query SQLite directly"},
		{Name: "validation_scenarios_are_final_answer_only", Passed: validationFinalAnswerOnly, Details: validationFinalAnswerDetails(validationFailures, missingValidationScenarios)},
	}
	passes := true
	for _, criterion := range criteria {
		if !criterion.Passed {
			passes = false
			break
		}
	}
	recommendation := "fix_production_agentops_before_release"
	if passes {
		recommendation = "use_agentops_runner_for_routine_openclerk_operations"
	}
	return &productionGateSummary{
		Variant:        productionVariant,
		PassesGate:     passes,
		Recommendation: recommendation,
		Criteria:       criteria,
	}
}

func buildTargetedLaneSummary(lane string, releaseBlocking bool, results []jobResult) *targetedLaneSummary {
	if releaseBlocking {
		return nil
	}
	if lane != populatedLaneName && lane != repoDocsLaneName && lane != graphSemanticsRevisitLaneName && lane != memoryRouterRevisitLaneName && lane != highTouchMemoryRouterRecallLaneName && lane != memoryRouterRecallCandidateLaneName && lane != memoryRouterRecallReportLaneName && lane != promotedRecordDomainLaneName && lane != highTouchRelationshipRecordLaneName && lane != relationshipRecordCandidateLaneName && lane != parallelRunnerLaneName && lane != documentHistoryLaneName && lane != highTouchDocumentLifecycleLaneName && lane != documentLifecycleRollbackCandidateLaneName && lane != agentChosenPathLaneName && lane != pathTitleAutonomyLaneName && lane != captureLowRiskLaneName && lane != captureExplicitOverridesLaneName && lane != captureDuplicateCandidateLaneName && lane != taggingLaneName && lane != captureSaveThisNoteLaneName && lane != captureDocumentLinksLaneName && lane != sourceURLUpdateLaneName && lane != webURLIntakeLaneName && lane != webURLStaleRepairLaneName && lane != webURLStaleImpactLaneName && lane != webProductPageLaneName && lane != documentThisLaneName && lane != documentArtifactCandidateLaneName && lane != artifactIngestionLaneName && lane != unsupportedArtifactKindLaneName && lane != videoYouTubeLaneName && lane != synthesisCompileLaneName && lane != highTouchCompileSynthesisLaneName && lane != compileSynthesisCandidateLaneName && lane != broadAuditLaneName {
		return nil
	}
	summary := targetedLaneSummary{
		Lane:            lane,
		PublicSurface:   []string{"openclerk document", "openclerk retrieval"},
		ReleaseBlocking: releaseBlocking,
	}
	if lane == documentArtifactCandidateLaneName {
		summary.PublicSurface = []string{"skills/openclerk/SKILL.md", "openclerk document", "openclerk retrieval"}
	}
	for _, result := range results {
		include := false
		classification, posture := "", ""
		switch lane {
		case populatedLaneName:
			include = isPopulatedVaultScenario(result.Scenario)
			classification, posture = classifyTargetedPopulatedResult(result)
		case repoDocsLaneName:
			include = isRepoDocsDogfoodScenario(result.Scenario)
			classification, posture = classifyTargetedRepoDocsResult(result)
		case graphSemanticsRevisitLaneName:
			include = isGraphSemanticsRevisitScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedGraphSemanticsRevisitResult(result)
		case memoryRouterRevisitLaneName:
			include = isMemoryRouterRevisitScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedMemoryRouterRevisitResult(result)
		case highTouchMemoryRouterRecallLaneName:
			include = isHighTouchMemoryRouterRecallScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedHighTouchMemoryRouterRecallResult(result)
		case memoryRouterRecallCandidateLaneName:
			include = isMemoryRouterRecallCandidateScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedMemoryRouterRecallCandidateResult(result)
		case memoryRouterRecallReportLaneName:
			include = isMemoryRouterRecallReportScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedMemoryRouterRecallReportResult(result)
		case promotedRecordDomainLaneName:
			include = isPromotedRecordDomainScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedPromotedRecordDomainResult(result)
		case highTouchRelationshipRecordLaneName:
			include = isHighTouchRelationshipRecordScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedHighTouchRelationshipRecordResult(result)
		case relationshipRecordCandidateLaneName:
			include = isRelationshipRecordCandidateScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedRelationshipRecordCandidateResult(result)
		case parallelRunnerLaneName:
			include = isParallelRunnerScenario(result.Scenario)
			classification, posture = classifyTargetedParallelRunnerResult(result)
		case documentHistoryLaneName:
			include = isDocumentHistoryScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedDocumentHistoryResult(result)
		case highTouchDocumentLifecycleLaneName:
			include = isHighTouchDocumentLifecycleScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedHighTouchDocumentLifecycleResult(result)
		case documentLifecycleRollbackCandidateLaneName:
			include = isDocumentLifecycleRollbackCandidateScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedDocumentLifecycleRollbackCandidateResult(result)
		case agentChosenPathLaneName:
			include = isAgentChosenPathScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedAgentChosenPathResult(result)
		case pathTitleAutonomyLaneName:
			include = isPathTitleAutonomyScenario(result.Scenario)
			classification, posture = classifyTargetedPathTitleAutonomyResult(result)
		case captureLowRiskLaneName:
			include = isCaptureLowRiskScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedCaptureLowRiskResult(result)
		case captureExplicitOverridesLaneName:
			include = isCaptureExplicitOverridesScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedCaptureExplicitOverridesResult(result)
		case captureDuplicateCandidateLaneName:
			include = isCaptureDuplicateCandidateScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedCaptureDuplicateCandidateResult(result)
		case taggingLaneName:
			include = isTaggingScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedTaggingResult(result)
		case captureSaveThisNoteLaneName:
			include = isCaptureSaveThisNoteScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedCaptureSaveThisNoteResult(result)
		case captureDocumentLinksLaneName:
			include = isCaptureDocumentLinksScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedCaptureDocumentLinksResult(result)
		case sourceURLUpdateLaneName:
			include = isSourceURLUpdateScenario(result.Scenario)
			classification, posture = classifyTargetedSourceURLUpdateResult(result)
		case webURLIntakeLaneName:
			include = isWebURLIntakeScenario(result.Scenario)
			classification, posture = classifyTargetedWebURLIntakeResult(result)
		case webURLStaleRepairLaneName:
			include = isWebURLStaleRepairScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedWebURLStaleRepairResult(result)
		case webURLStaleImpactLaneName:
			include = isWebURLStaleImpactScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedWebURLStaleImpactResult(result)
		case webProductPageLaneName:
			include = isWebProductPageScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedWebProductPageResult(result)
		case documentThisLaneName:
			include = isDocumentThisScenario(result.Scenario)
			classification, posture = classifyTargetedDocumentThisResult(result)
		case documentArtifactCandidateLaneName:
			include = isDocumentArtifactCandidateScenario(result.Scenario)
			classification, posture = classifyTargetedDocumentArtifactCandidateResult(result)
		case artifactIngestionLaneName:
			include = isArtifactIngestionScenario(result.Scenario)
			classification, posture = classifyTargetedArtifactIngestionResult(result)
		case unsupportedArtifactKindLaneName:
			include = isUnsupportedArtifactKindScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedUnsupportedArtifactKindResult(result)
		case videoYouTubeLaneName:
			include = isVideoYouTubeScenario(result.Scenario)
			classification, posture = classifyTargetedVideoYouTubeResult(result)
		case synthesisCompileLaneName:
			include = isSynthesisCompileScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedSynthesisCompileResult(result)
		case highTouchCompileSynthesisLaneName:
			include = isHighTouchCompileSynthesisScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedSynthesisCompileResult(result)
		case compileSynthesisCandidateLaneName:
			include = isCompileSynthesisCandidateScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedCompileSynthesisCandidateResult(result)
		case broadAuditLaneName:
			include = isBroadAuditScenario(result.Scenario) || isFinalAnswerOnlyValidationScenario(result.Scenario)
			classification, posture = classifyTargetedBroadAuditResult(result)
		}
		if !include {
			continue
		}
		summary.ScenarioClassifications = append(summary.ScenarioClassifications, targetedScenarioClassification{
			Variant:               result.Variant,
			Scenario:              result.Scenario,
			Status:                result.Status,
			FailureClassification: classification,
			EvidencePosture:       posture,
			ToolCalls:             result.Metrics.ToolCalls,
			CommandExecutions:     result.Metrics.CommandExecutions,
			AssistantCalls:        result.Metrics.AssistantCalls,
			WallSeconds:           result.WallSeconds,
			PromptSpecificity:     promptSpecificity(result.Scenario),
			UX:                    scenarioUX(result),
			Brittleness:           scenarioBrittleness(result),
			Retries:               scenarioRetries(result),
			StepCount:             scenarioStepCount(result),
			Latency:               scenarioLatency(result),
			GuidanceDependence:    scenarioGuidanceDependence(result),
			SafetyPass:            scenarioSafetyPass(result, classification),
			CapabilityPass:        scenarioCapabilityPass(result, classification),
			UXQuality:             scenarioUXQuality(result, classification),
			SafetyRisks:           scenarioSafetyRisks(result),
			FixturePreflight:      fixturePreflightStatus(result.FixturePreflight),
		})
	}
	if len(summary.ScenarioClassifications) == 0 {
		return nil
	}
	switch lane {
	case populatedLaneName:
		summary.Decision = "keep_as_reference"
		summary.Promotion = "no promoted runner action, schema, migration, storage API, product behavior, or public OpenClerk interface"
	case repoDocsLaneName:
		summary.Decision = "keep_as_public_dogfood_lane"
		summary.Promotion = "targeted repo-docs dogfood evidence only; no promoted runner action, schema, migration, storage API, product behavior, or public OpenClerk interface"
	case graphSemanticsRevisitLaneName:
		summary.Decision = graphSemanticsRevisitDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted graph semantics revisit evidence only; no semantic-label graph layer, runner action, schema, migration, storage behavior, or public API change from this eval"
	case memoryRouterRevisitLaneName:
		summary.Decision = memoryRouterRevisitDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted memory and autonomous router revisit evidence only; no remember/recall action, memory transport, autonomous router API, schema, migration, storage behavior, or public API change from this eval"
	case highTouchMemoryRouterRecallLaneName:
		summary.Decision = highTouchMemoryRouterRecallDecision(summary.ScenarioClassifications)
		summary.Promotion = highTouchMemoryRouterRecallPromotion(summary.Decision)
	case memoryRouterRecallCandidateLaneName:
		summary.Decision = memoryRouterRecallCandidateDecision(summary.ScenarioClassifications)
		summary.Promotion = memoryRouterRecallCandidatePromotion(summary.Decision)
	case memoryRouterRecallReportLaneName:
		summary.Decision = memoryRouterRecallReportImplementationDecision(summary.ScenarioClassifications)
		summary.Promotion = "implemented read-only memory_router_recall_report retrieval action; no schema, storage, document write behavior, memory transport, remember/recall action, autonomous router API, vector/embedding/graph memory, bypass, or hidden authority ranking"
	case promotedRecordDomainLaneName:
		summary.Decision = promotedRecordDomainDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted promoted record domain expansion evidence only; no policy-specific record action, typed domain runner surface, schema, migration, storage behavior, or public API change from this eval"
	case highTouchRelationshipRecordLaneName:
		summary.Decision = highTouchRelationshipRecordDecision(summary.ScenarioClassifications)
		summary.Promotion = highTouchRelationshipRecordPromotion(summary.Decision)
	case relationshipRecordCandidateLaneName:
		summary.Decision = relationshipRecordCandidateDecision(summary.ScenarioClassifications)
		summary.Promotion = relationshipRecordCandidatePromotion(summary.Decision)
	case parallelRunnerLaneName:
		summary.Decision = "relax_skill_guidance_for_safe_parallel_reads"
		summary.Promotion = "targeted parallel runner UX evidence for documented safe read/startup workflows; no public JSON schema, storage schema, or write-concurrency expansion"
	case documentHistoryLaneName:
		summary.Decision = documentHistoryDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted document lifecycle evidence only; no promoted history, diff, review, restore, rollback, schema, migration, storage behavior, or public API change from this eval"
	case highTouchDocumentLifecycleLaneName:
		summary.Decision = highTouchDocumentLifecycleDecision(summary.ScenarioClassifications)
		summary.Promotion = highTouchDocumentLifecyclePromotion(summary.Decision)
	case documentLifecycleRollbackCandidateLaneName:
		summary.Decision = documentLifecycleRollbackCandidateDecision(summary.ScenarioClassifications)
		summary.Promotion = documentLifecycleRollbackCandidatePromotion(summary.Decision)
	case agentChosenPathLaneName:
		summary.Decision = agentChosenPathDecision(summary.ScenarioClassifications)
		summary.Promotion = "no promoted runner action, schema, migration, storage API, product behavior, public OpenClerk interface, or change to missing-path clarification"
	case pathTitleAutonomyLaneName:
		summary.Decision = "evaluate_for_oc_iat"
		summary.Promotion = "no promoted runner action, schema, migration, skill behavior, storage API, product behavior, or public OpenClerk interface from this eval"
	case captureLowRiskLaneName:
		summary.Decision = captureLowRiskDecision(summary.ScenarioClassifications)
		summary.Promotion = captureLowRiskPromotion(summary.Decision)
	case captureExplicitOverridesLaneName:
		summary.Decision = captureExplicitOverridesDecision(summary.ScenarioClassifications)
		summary.Promotion = captureExplicitOverridesPromotion(summary.Decision)
	case captureDuplicateCandidateLaneName:
		summary.Decision = captureDuplicateCandidateDecision(summary.ScenarioClassifications)
		summary.Promotion = captureDuplicateCandidatePromotion(summary.Decision)
	case taggingLaneName:
		summary.Decision = taggingDecision(summary.ScenarioClassifications)
		summary.Promotion = taggingPromotion(summary.Decision)
	case captureSaveThisNoteLaneName:
		summary.Decision = captureSaveThisNoteDecision(summary.ScenarioClassifications)
		summary.Promotion = captureSaveThisNotePromotion(summary.Decision)
	case captureDocumentLinksLaneName:
		summary.Decision = captureDocumentLinksDecision(summary.ScenarioClassifications)
		summary.Promotion = captureDocumentLinksPromotion(summary.Decision)
	case sourceURLUpdateLaneName:
		summary.Decision = "keep_existing_update_mode"
		summary.Promotion = "targeted AgentOps evidence for existing ingest_source_url source.mode update behavior; no new runner action, schema, storage API, or transport"
	case webURLIntakeLaneName:
		summary.Decision = webURLIntakeDecision(summary.ScenarioClassifications)
		summary.Promotion = "promote ingest_source_url web source handling; same runner action, source.source_type extension, no external acquisition tools"
	case webURLStaleRepairLaneName:
		summary.Decision = webURLStaleRepairDecision(summary.ScenarioClassifications)
		summary.Promotion = webURLStaleRepairPromotion(summary.Decision)
	case webURLStaleImpactLaneName:
		summary.Decision = webURLStaleImpactDecision(summary.ScenarioClassifications)
		summary.Promotion = webURLStaleImpactPromotion(summary.Decision)
	case webProductPageLaneName:
		summary.Decision = webProductPageDecision(summary.ScenarioClassifications)
		summary.Promotion = webProductPagePromotion(summary.Decision)
	case documentThisLaneName:
		summary.Decision = "evaluate_for_oc_99z"
		summary.Promotion = "no promoted runner action, schema, migration, skill behavior, storage API, product behavior, or public OpenClerk interface from this eval"
	case documentArtifactCandidateLaneName:
		summary.Decision = documentArtifactCandidateDecision(summary.ScenarioClassifications)
		switch summary.Decision {
		case "promote_propose_before_create_skill_policy":
			summary.Promotion = "skill policy supports propose-before-create candidate path/title/body generation only; no runner action, schema, storage, migration, direct create, or public API change"
		case "defer_for_candidate_ergonomics_repair":
			summary.Promotion = "ergonomics promotion deferred; existing shipped propose-before-create skill policy needs natural-intent repair before oc-99z can promote it; no runner action, schema, storage, migration, direct create, or public API change"
		default:
			summary.Promotion = "no promoted skill policy yet; repair candidate quality gaps before any propose-before-create skill behavior change"
		}
	case artifactIngestionLaneName:
		summary.Decision = artifactIngestionDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted evidence only; no promoted runner action, parser, schema, storage migration, direct create behavior, or public API change"
	case unsupportedArtifactKindLaneName:
		summary.Decision = unsupportedArtifactKindDecision(summary.ScenarioClassifications)
		summary.Promotion = unsupportedArtifactKindPromotion(summary.Decision)
	case videoYouTubeLaneName:
		summary.Decision = videoYouTubeDecision(summary.ScenarioClassifications)
		summary.Promotion = "keep supplied-transcript ingest_video_url as the promoted surface; native acquisition dependencies remain deferred"
	case synthesisCompileLaneName:
		summary.Decision = synthesisCompileDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted evidence only; no compile_synthesis runner action, schema, migration, storage behavior, direct vault behavior, or public API change from this eval"
	case highTouchCompileSynthesisLaneName:
		summary.Decision = highTouchCompileSynthesisDecision(summary.ScenarioClassifications)
		summary.Promotion = highTouchCompileSynthesisPromotion(summary.Decision)
	case compileSynthesisCandidateLaneName:
		summary.Decision = compileSynthesisCandidateDecision(summary.ScenarioClassifications)
		summary.Promotion = compileSynthesisCandidatePromotion(summary.Decision)
	case broadAuditLaneName:
		summary.Decision = broadAuditDecision(summary.ScenarioClassifications)
		summary.Promotion = "targeted broad contradiction/audit revisit evidence only; no broad semantic contradiction engine, audit runner action, schema, migration, storage behavior, or public API change from this eval"
	}
	return &summary
}
