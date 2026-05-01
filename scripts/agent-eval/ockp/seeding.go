package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
)

func seedScenario(ctx context.Context, paths evalPaths, sc scenario) error {
	return seedScenarioWithFixtures(ctx, paths, sc, nil)
}
func seedScenarioWithFixtures(ctx context.Context, paths evalPaths, sc scenario, fixtures *sourceURLUpdateFixtures) error {
	cfg := runclient.Config{
		DatabasePath: paths.DatabasePath,
	}
	switch sc.ID {
	case "search-synthesis", "mt-source-then-synthesis":
		if err := createSeedDocument(ctx, cfg, "sources/openclerk-runner.md", "OpenClerk Runner Source", "The OpenClerk runner uses JSON requests for OpenClerk knowledge tasks.\n\nIt preserves source refs for synthesis pages."); err != nil {
			return err
		}
	case "answer-filing":
		if err := createSeedDocument(ctx, cfg, "sources/answer-filing-runner.md", "OpenClerk runner Answer Filing Source", "The OpenClerk runner JSON runner is the production path for reusable OpenClerk knowledge tasks.\n\nDurable OpenClerk runner answers should be filed as source-linked markdown."); err != nil {
			return err
		}
	case ragRetrievalScenarioID:
		if err := seedRAGRetrievalBaseline(ctx, cfg); err != nil {
			return err
		}
	case parallelRunnerReadsScenarioID:
		if err := seedParallelRunnerReads(ctx, cfg); err != nil {
			return err
		}
	case docsNavigationScenarioID:
		if err := seedDocsNavigationBaseline(ctx, cfg); err != nil {
			return err
		}
	case graphSemanticsScenarioID, graphSemanticsNaturalScenarioID, graphSemanticsScriptedScenarioID:
		if err := seedGraphSemanticsReference(ctx, cfg); err != nil {
			return err
		}
	case memoryRouterScenarioID:
		if err := seedMemoryRouterReference(ctx, cfg); err != nil {
			return err
		}
	case memoryRouterNaturalScenarioID, memoryRouterScriptedScenarioID:
		if err := seedMemoryRouterRevisit(ctx, cfg); err != nil {
			return err
		}
	case promotedRecordDomainNaturalScenarioID, promotedRecordDomainScriptedScenarioID:
		if err := seedPromotedRecordDomainExpansion(ctx, cfg); err != nil {
			return err
		}
	case configuredLayoutScenarioID:
		if err := seedConfiguredLayoutScenario(ctx, cfg); err != nil {
			return err
		}
	case invalidLayoutScenarioID:
		if err := seedInvalidLayoutScenario(ctx, cfg); err != nil {
			return err
		}
	case sourceURLUpdateDuplicateScenarioID, sourceURLUpdateConflictScenarioID:
		if fixtures == nil {
			return errors.New("source URL update fixture server is required")
		}
		if err := seedSourceURLUpdateSource(ctx, cfg, fixtures.stableURL()); err != nil {
			return err
		}
	case sourceURLUpdateSameSHAScenarioID:
		if fixtures == nil {
			return errors.New("source URL update fixture server is required")
		}
		if err := seedSourceURLUpdateSource(ctx, cfg, fixtures.stableURL()); err != nil {
			return err
		}
		if err := seedSourceURLUpdateSynthesis(ctx, cfg); err != nil {
			return err
		}
	case sourceURLUpdateChangedScenarioID:
		if fixtures == nil {
			return errors.New("source URL update fixture server is required")
		}
		if err := seedSourceURLUpdateSource(ctx, cfg, fixtures.changedURL()); err != nil {
			return err
		}
		if err := seedSourceURLUpdateSynthesis(ctx, cfg); err != nil {
			return err
		}
	case webURLDuplicateScenarioID, webURLSameHashScenarioID, webURLChangedScenarioID, webURLStaleRepairNaturalScenarioID, webURLStaleRepairScriptedScenarioID, webURLStaleImpactCurrentPrimitivesScenarioID, webURLStaleImpactGuidanceOnlyScenarioID, webURLStaleImpactResponseCandidateScenarioID:
		if fixtures == nil {
			return errors.New("web URL intake fixture server is required")
		}
		if err := seedWebURLIntakeSource(ctx, cfg, fixtures.stableURL(), fixtures.initialHTML); err != nil {
			return err
		}
		if sc.ID == webURLSameHashScenarioID || sc.ID == webURLChangedScenarioID || isWebURLStaleRepairScenario(sc.ID) || isWebURLStaleImpactScenario(sc.ID) {
			if err := seedWebURLIntakeSynthesis(ctx, cfg); err != nil {
				return err
			}
		}
	case webProductPageDuplicateScenarioID:
		if fixtures == nil {
			return errors.New("web product page fixture server is required")
		}
		if err := seedWebProductPageSource(ctx, cfg, webProductPageEvalSourceURL, fixtures.productPageHTML, webProductPageSourcePath, webProductPageTitle); err != nil {
			return err
		}
	case synthesisCandidatePressureScenarioID:
		if err := seedSynthesisCandidatePressure(ctx, cfg); err != nil {
			return err
		}
	case synthesisSourceSetPressureScenarioID:
		if err := seedSynthesisSourceSetPressure(ctx, cfg); err != nil {
			return err
		}
	case synthesisCompileNaturalScenarioID, synthesisCompileScriptedScenarioID, highTouchCompileSynthesisNaturalScenarioID, highTouchCompileSynthesisScriptedScenarioID:
		if err := seedSynthesisCompileRevisit(ctx, cfg); err != nil {
			return err
		}
	case broadAuditNaturalScenarioID, broadAuditScriptedScenarioID:
		if err := seedBroadContradictionAuditRevisit(ctx, cfg); err != nil {
			return err
		}
	case decisionRecordVsDocsScenarioID:
		if err := seedDecisionRecordVsDocs(ctx, cfg); err != nil {
			return err
		}
	case decisionSupersessionScenarioID:
		if err := seedDecisionSupersession(ctx, cfg); err != nil {
			return err
		}
	case decisionRealADRMigrationScenarioID:
		if err := seedDecisionRealADRMigration(ctx, cfg); err != nil {
			return err
		}
	case sourceAuditRepairScenarioID:
		if err := seedSourceSensitiveAuditRepair(ctx, cfg); err != nil {
			return err
		}
	case sourceAuditConflictScenarioID:
		if err := seedSourceSensitiveConflict(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryNaturalScenarioID:
		if err := seedDocumentHistoryRestore(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryInspectScenarioID:
		if err := seedDocumentHistoryInspection(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryDiffScenarioID:
		if err := seedDocumentHistoryDiffReview(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryRestoreScenarioID:
		if err := seedDocumentHistoryRestore(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryPendingScenarioID:
		if err := seedDocumentHistoryPendingReview(ctx, cfg); err != nil {
			return err
		}
	case documentHistoryStaleScenarioID:
		if err := seedDocumentHistoryStaleSynthesis(ctx, cfg); err != nil {
			return err
		}
	case mtSynthesisDriftPressureScenarioID:
		if err := seedMTSynthesisDriftPressure(ctx, cfg); err != nil {
			return err
		}
	case populatedHeterogeneousScenarioID, populatedFreshnessConflictScenarioID, populatedSynthesisUpdateScenarioID:
		if err := seedPopulatedVaultFixture(ctx, cfg); err != nil {
			return err
		}
	case repoDocsRetrievalScenarioID, repoDocsSynthesisScenarioID, repoDocsDecisionScenarioID:
		if err := seedRepoDocsDogfood(ctx, cfg); err != nil {
			return err
		}
	case agentChosenSynthesisScenarioID:
		if err := seedAgentChosenSynthesisPathSelection(ctx, cfg); err != nil {
			return err
		}
	case pathTitleMultiSourceDuplicateScenarioID:
		if err := seedPathTitleMultiSourceDuplicatePressure(ctx, cfg); err != nil {
			return err
		}
	case pathTitleDuplicateRiskScenarioID:
		if err := seedPathTitleDuplicateRiskPressure(ctx, cfg); err != nil {
			return err
		}
	case documentThisDuplicateCandidateScenarioID:
		if err := seedDocumentThisDuplicateCandidate(ctx, cfg); err != nil {
			return err
		}
	case documentThisExistingUpdateScenarioID:
		if err := seedDocumentThisExistingUpdate(ctx, cfg); err != nil {
			return err
		}
	case documentThisSynthesisFreshnessScenarioID:
		if err := seedDocumentThisSynthesisFreshness(ctx, cfg); err != nil {
			return err
		}
	case candidateDuplicateRiskAsksScenarioID, candidateErgonomicsDuplicateNaturalID:
		if err := seedDocumentArtifactCandidateDuplicate(ctx, cfg); err != nil {
			return err
		}
	case captureExplicitOverridesAuthorityConflictID:
		if err := seedCaptureExplicitOverridesAuthorityConflict(ctx, cfg); err != nil {
			return err
		}
	case captureLowRiskDuplicateScenarioID:
		if err := seedCaptureLowRiskDuplicate(ctx, cfg); err != nil {
			return err
		}
	case captureDuplicateCandidateNaturalScenarioID, captureDuplicateCandidateScriptedScenarioID, captureDuplicateCandidateAccuracyScenarioID:
		if err := seedCaptureDuplicateCandidate(ctx, cfg); err != nil {
			return err
		}
	case taggingRetrievalScenarioID, taggingDisambiguationScenarioID, taggingNearDuplicateScenarioID, taggingMixedPathScenarioID:
		if err := seedTaggingWorkflows(ctx, cfg); err != nil {
			return err
		}
	case captureSaveThisNoteDuplicateScenarioID:
		if err := seedCaptureSaveThisNoteDuplicate(ctx, cfg); err != nil {
			return err
		}
	case captureDocumentLinksSynthesisScenarioID:
		if err := seedCaptureDocumentLinksSources(ctx, cfg); err != nil {
			return err
		}
	case captureDocumentLinksDuplicateScenarioID:
		if err := seedCaptureDocumentLinksDuplicate(ctx, cfg); err != nil {
			return err
		}
	case artifactTranscriptScenarioID:
		if err := seedArtifactTranscript(ctx, cfg); err != nil {
			return err
		}
	case artifactInvoiceReceiptScenarioID:
		if err := seedArtifactInvoiceReceipt(ctx, cfg); err != nil {
			return err
		}
	case artifactMixedSynthesisScenarioID:
		if err := seedArtifactMixedSynthesis(ctx, cfg); err != nil {
			return err
		}
	case videoYouTubeSynthesisFreshnessScenarioID:
		if err := seedVideoYouTubeSynthesisFreshness(ctx, cfg); err != nil {
			return err
		}
	case "stale-synthesis-update":
		if err := createSeedDocument(ctx, cfg, "sources/runner-old-workaround.md", "Old OpenClerk runner Routing Source", "Older guidance said routine agents may bypass OpenClerk runner through a temporary command-path workaround."); err != nil {
			return err
		}
		if err := createSeedDocument(ctx, cfg, "sources/runner-current-runner.md", "Current OpenClerk runner Routing Source", "Current guidance says routine agents must use openclerk JSON runner for OpenClerk knowledge tasks."); err != nil {
			return err
		}
		body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/runner-current-runner.md, sources/runner-old-workaround.md
---

# OpenClerk runner Routing

## Summary
Stale claim: routine agents may bypass OpenClerk runner through a temporary command-path workaround.

## Sources
- sources/runner-current-runner.md
- sources/runner-old-workaround.md

## Freshness
Checked source: sources/runner-old-workaround.md
`)
		if err := createSeedDocument(ctx, cfg, "synthesis/runner-routing.md", "OpenClerk runner Routing", body); err != nil {
			return err
		}
	case "synthesis-freshness-repair":
		oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: sources/repair-current.md
---
# Old OpenClerk runner Repair Source

## Summary
Older repair guidance mentioned a temporary command-path workaround.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "sources/repair-old.md", "Old OpenClerk runner Repair Source", oldBody); err != nil {
			return err
		}
		currentBody := strings.TrimSpace(`---
supersedes: sources/repair-old.md
---
# Current OpenClerk runner Repair Source

## Summary
Current guidance says routine agents must use openclerk JSON runner for freshness repairs.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "sources/repair-current.md", "Current OpenClerk runner Repair Source", currentBody); err != nil {
			return err
		}
		synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/repair-current.md, sources/repair-old.md
---
# OpenClerk runner Freshness Repair

## Summary
Stale repair claim: routine agents may use a temporary command-path workaround.

## Sources
- sources/repair-current.md
- sources/repair-old.md

## Freshness
Checked before the latest source update.
`) + "\n"
		if err := createSeedDocument(ctx, cfg, "synthesis/runner-repair.md", "OpenClerk runner Freshness Repair", synthesisBody); err != nil {
			return err
		}
		if err := replaceScenarioSeedSection(ctx, cfg, "sources/repair-current.md", "Summary", "Current guidance says routine agents must use openclerk JSON runner for freshness repairs, and sources/repair-old.md is superseded."); err != nil {
			return err
		}
	case "append-replace":
		if err := createSeedDocument(ctx, cfg, "notes/projects/openclerk-runner.md", "OpenClerk Runner", "## Context\nExisting context stays intact."); err != nil {
			return err
		}
	case "records-provenance":
		if err := createSeedDocument(ctx, cfg, "records/services/openclerk-runner.md", "OpenClerk runner", recordBody("openclerk-runner", "service", "OpenClerk runner")); err != nil {
			return err
		}
	case "mixed-synthesis-records":
		if err := createSeedDocument(ctx, cfg, "sources/openclerk-runner.md", "OpenClerk Runner Source", "The OpenClerk runner uses JSON requests for OpenClerk knowledge tasks.\n\nIt preserves source refs for synthesis pages."); err != nil {
			return err
		}
		if err := createSeedDocument(ctx, cfg, "records/services/openclerk-runner.md", "OpenClerk runner", recordBody("openclerk-runner", "service", "OpenClerk runner")); err != nil {
			return err
		}
	case "promoted-record-vs-docs":
		if err := createSeedDocument(ctx, cfg, "notes/reference/runner-service.md", "OpenClerk runner Service Reference", "# OpenClerk runner Service Reference\n\n## Summary\nPlain docs evidence says OpenClerk runner is the production service for routine knowledge tasks.\n\n## Details\nPlain docs evidence is narrative and searchable.\n"); err != nil {
			return err
		}
		body := strings.TrimSpace(`---
service_id: openclerk-runner
service_name: OpenClerk runner
service_status: active
service_owner: runner
service_interface: JSON runner
---

# OpenClerk runner

## Facts
- production_path: true
`)
		if err := createSeedDocument(ctx, cfg, "records/services/openclerk-runner.md", "OpenClerk runner", body); err != nil {
			return err
		}
	case "duplicate-path-reject":
		if err := createSeedDocument(ctx, cfg, "notes/projects/duplicate.md", "Duplicate Source", "This canonical path already exists."); err != nil {
			return err
		}
	}
	return nil
}
func sourceTitleFromPath(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), ".md")
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func seedCaptureDocumentLinksSources(ctx context.Context, cfg runclient.Config) error {
	firstBody := strings.TrimSpace(`---
type: source
status: active
source_url: https://example.test/openclerk-runner-guidance
---
# Runner Guidance Link

## Summary
Document-these-links public evidence says source path hints must be confirmed before durable source writes.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, captureDocumentLinksSourcePath, captureDocumentLinksSourceTitle, firstBody); err != nil {
		return err
	}
	secondBody := strings.TrimSpace(`---
type: source
status: active
source_url: https://example.test/openclerk-freshness-guidance
---
# Freshness Guidance Link

## Summary
Document-these-links public evidence says synthesis placement should be proposed only after source intent is clear.
`) + "\n"
	return createSeedDocument(ctx, cfg, captureDocumentLinksSecondSourcePath, captureDocumentLinksSecondSourceTitle, secondBody)
}

func seedCaptureDocumentLinksDuplicate(ctx context.Context, cfg runclient.Config) error {
	sourceBody := strings.TrimSpace(`---
type: source
status: active
source_url: https://example.test/openclerk-runner-guidance
---
# Existing Runner Guidance Link

## Summary
document these links placement runner guidance marker: existing public source evidence already covers runner guidance.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, captureDocumentLinksDuplicateSourcePath, captureDocumentLinksDuplicateSourceTitle, sourceBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/document-these-links/existing-runner-guidance.md
---
# Document These Links Placement

## Summary
Existing synthesis candidate for document-these-links placement.

## Sources
- sources/document-these-links/existing-runner-guidance.md

## Freshness
Checked existing source placement before duplicate capture.
`) + "\n"
	return createSeedDocument(ctx, cfg, captureDocumentLinksSynthesisPath, captureDocumentLinksSynthesisTitle, synthesisBody)
}

func seedParallelRunnerReads(ctx context.Context, cfg runclient.Config) error {
	if err := createSeedDocument(ctx, cfg, parallelRunnerDocPath, "Parallel Runner Read Contract", "# Parallel Runner Read Contract\n\n## Summary\nParallel runner safe read contract evidence says resolve_paths, list_documents, retrieval search, service lookup, decision lookup, provenance, and projection reads may run concurrently without raw SQLite runtime_config or upsert failures.\n"); err != nil {
		return err
	}
	serviceBody := strings.TrimSpace(`---
service_id: parallel-runner
service_name: Parallel runner
service_status: active
service_owner: runner
service_interface: JSON runner
---
# Parallel runner

## Summary
Parallel runner service evidence confirms safe read workflows stay on the installed JSON runner.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, parallelRunnerServicePath, "Parallel runner", serviceBody); err != nil {
		return err
	}
	decisionBody := strings.TrimSpace(`---
decision_id: adr-parallel-runner-concurrency
decision_title: Parallel runner concurrency
decision_status: accepted
decision_scope: runner
decision_owner: platform
decision_date: 2026-04-29
---
# Parallel runner concurrency

## Summary
The accepted parallel runner concurrency decision permits safe read workflows while writes remain serialized.
`) + "\n"
	return createSeedDocument(ctx, cfg, parallelRunnerDecisionPath, "Parallel runner concurrency", decisionBody)
}
func createSeedDocument(ctx context.Context, cfg runclient.Config, path, title, body string) error {
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  path,
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	if result.Rejected {
		return errors.New(result.RejectionReason)
	}
	return nil
}
func seedSourceURLUpdateSource(ctx context.Context, cfg runclient.Config, sourceURL string) error {
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           sourceURL,
			PathHint:      sourceURLUpdateSourcePath,
			AssetPathHint: sourceURLUpdateAssetPath,
			Title:         "Source URL Update Runner",
		},
	})
	if err != nil {
		return err
	}
	if result.Ingestion == nil || result.Ingestion.SourcePath != sourceURLUpdateSourcePath {
		return fmt.Errorf("source URL update seed ingestion = %+v", result.Ingestion)
	}
	return nil
}
func seedSourceURLUpdateSynthesis(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/source-url-update-runner.md
---

# Source URL Update Runner

## Summary
Initial synthesis depends on SourceURLUpdateInitialEvidence.

## Sources
- sources/source-url-update-runner.md

## Freshness
Checked source URL update source before PDF refresh.
	`) + "\n"
	return createSeedDocument(ctx, cfg, sourceURLUpdateSynthesisPath, "Source URL Update Runner", body)
}

func seedWebURLIntakeSource(ctx context.Context, cfg runclient.Config, sourceURL string, htmlBody []byte) error {
	sha := sha256.Sum256(htmlBody)
	shaHex := hex.EncodeToString(sha[:])
	body := strings.TrimSpace(fmt.Sprintf(`---
type: source
source_type: web
modality: markdown
source_url: "%s"
derived_path: "%s"
sha256: "%s"
size_bytes: %d
mime_type: "text/html"
captured_at: "2026-04-29T00:00:00Z"
source_title: "%s"
---
# %s

## Summary
Web source ingested from %s.

## Source Page
- Source URL: %s
- SHA256: %s
- Size bytes: %d
- Page title: %s

## Extracted Text
%s %s visible public product-page evidence. Add to cart
`, sourceURL, webURLSourcePath, shaHex, len(htmlBody), webURLTitle, webURLTitle, sourceURL, sourceURL, shaHex, len(htmlBody), webURLTitle, webURLTitle, webURLInitialText)) + "\n"
	return createSeedDocument(ctx, cfg, webURLSourcePath, webURLTitle, body)
}

func seedWebProductPageSource(ctx context.Context, cfg runclient.Config, sourceURL string, htmlBody []byte, sourcePath string, title string) error {
	sha := sha256.Sum256(htmlBody)
	shaHex := hex.EncodeToString(sha[:])
	body := strings.TrimSpace(fmt.Sprintf(`---
type: source
source_type: web
modality: markdown
source_url: "%s"
derived_path: "%s"
sha256: "%s"
size_bytes: %d
mime_type: "text/html"
captured_at: "2026-04-29T00:00:00Z"
source_title: "%s"
---
# %s

## Summary
Product page source ingested from %s.

## Source Page
- Source URL: %s
- SHA256: %s
- Size bytes: %d
- Page title: %s

## Extracted Text
%s visible public product-page evidence.
%s selected variant copy.
Add to cart
`, sourceURL, sourcePath, shaHex, len(htmlBody), title, title, sourceURL, sourceURL, shaHex, len(htmlBody), title, webProductPageText, webProductPageVariantText)) + "\n"
	return createSeedDocument(ctx, cfg, sourcePath, title, body)
}

func seedWebURLIntakeSynthesis(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/web-url/product-page.md
---

# Web URL Product Page

## Summary
Initial synthesis depends on WebURLIntakeInitialEvidence.

## Sources
- sources/web-url/product-page.md

## Freshness
Checked web URL intake source before web refresh.
`) + "\n"
	return createSeedDocument(ctx, cfg, webURLSynthesisPath, "Web URL Product Page", body)
}

func prepareSourceURLUpdateAgentState(ctx context.Context, paths evalPaths, sc scenario, fixtures *sourceURLUpdateFixtures) error {
	if sc.ID != sourceURLUpdateChangedScenarioID {
		return nil
	}
	cfg := runclient.Config{DatabasePath: paths.DatabasePath}
	result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionIngestSourceURL,
		Source: runner.SourceURLInput{
			URL:           fixtures.changedURL(),
			PathHint:      sourceURLUpdateSourcePath,
			AssetPathHint: sourceURLUpdateAssetPath,
			Mode:          "update",
		},
	})
	if err != nil {
		return err
	}
	if result.Ingestion == nil || result.Ingestion.SourcePath != sourceURLUpdateSourcePath {
		return fmt.Errorf("source URL update preparation = %+v", result.Ingestion)
	}
	return nil
}
func replaceScenarioSeedSection(ctx context.Context, cfg runclient.Config, docPath, heading, content string) error {
	list, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionList,
		List:   runner.DocumentListOptions{PathPrefix: docPath, Limit: 5},
	})
	if err != nil {
		return err
	}
	for _, doc := range list.Documents {
		if doc.Path != docPath {
			continue
		}
		result, err := runner.RunDocumentTask(ctx, cfg, runner.DocumentTaskRequest{
			Action:  runner.DocumentTaskActionReplaceSection,
			DocID:   doc.DocID,
			Heading: heading,
			Content: content,
		})
		if err != nil {
			return err
		}
		if result.Rejected {
			return errors.New(result.RejectionReason)
		}
		return nil
	}
	return fmt.Errorf("seed document %s not found", docPath)
}
func recordBody(entityID, entityType, name string) string {
	return strings.TrimSpace(fmt.Sprintf(`---
entity_id: %s
entity_type: %s
entity_name: %s
---

# %s

## Facts
- status: active
- owner: runner
`, entityID, entityType, name, name))
}
