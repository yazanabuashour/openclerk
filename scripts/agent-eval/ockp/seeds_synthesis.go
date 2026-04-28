package main

import (
	"context"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func seedSynthesisCandidatePressure(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: sources/compiler-current.md
---
# Compiler Old Source

## Summary
Older compiler guidance said routine synthesis repairs need a dedicated compile_synthesis action.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCandidateOldSrc, "Compiler Old Source", oldBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/compiler-current.md, sources/compiler-old.md
---
# Compiler Routing

## Summary
Stale compiler claim: routine synthesis repairs require a dedicated compile_synthesis runner action.

## Sources
- sources/compiler-current.md
- sources/compiler-old.md

## Freshness
Checked before the latest compiler pressure source was registered.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCandidatePath, "Compiler Routing", synthesisBody); err != nil {
		return err
	}
	decoyBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/compiler-old.md
---
# Compiler Routing Decoy

## Summary
This decoy page is not the compiler pressure decision target.

## Sources
- sources/compiler-old.md

## Freshness
Checked decoy source only.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCandidateDecoyPath, "Compiler Routing Decoy", decoyBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
supersedes: sources/compiler-old.md
---
# Compiler Current Source

## Summary
Current compiler pressure guidance says existing document and retrieval actions are sufficient when agents search sources, list synthesis candidates, inspect freshness, and update without duplicates.
`) + "\n"
	return createSeedDocument(ctx, cfg, synthesisCandidateCurrentSrc, "Compiler Current Source", currentBody)
}
func seedSynthesisSourceSetPressure(ctx context.Context, cfg runclient.Config) error {
	sourceBodies := map[string]string{
		sourceSetAlphaPath: strings.TrimSpace(`---
type: source
status: active
source_set: compiler-pressure
---
# Source Set Alpha

## Summary
Alpha source says synthesis compiler pressure requires source search before durable synthesis.
`) + "\n",
		sourceSetBetaPath: strings.TrimSpace(`---
type: source
status: active
source_set: compiler-pressure
---
# Source Set Beta

## Summary
Beta source says synthesis compiler pressure requires listing existing synthesis candidates.
`) + "\n",
		sourceSetGammaPath: strings.TrimSpace(`---
type: source
status: active
source_set: compiler-pressure
---
# Source Set Gamma

## Summary
Gamma source says synthesis compiler pressure requires preserving freshness and source refs.
`) + "\n",
	}
	for _, path := range []string{sourceSetAlphaPath, sourceSetBetaPath, sourceSetGammaPath} {
		if err := createSeedDocument(ctx, cfg, path, sourceTitleFromPath(path), sourceBodies[path]); err != nil {
			return err
		}
	}
	return nil
}
func seedSynthesisCompileRevisit(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
type: source
status: superseded
superseded_by: sources/compile-revisit-current.md
---
# Compile Revisit Old Source

## Summary
Earlier compile_synthesis revisit notes claimed routine synthesis updates required a dedicated compile_synthesis runner action.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCompileOldSrc, "Compile Revisit Old Source", oldBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/compile-revisit-current.md, sources/compile-revisit-old.md
---
# Compile Revisit Routing

## Summary
Stale compile_synthesis revisit claim: routine synthesis updates require a dedicated compile_synthesis runner action.

## Sources
- sources/compile-revisit-current.md
- sources/compile-revisit-old.md

## Freshness
Checked before the latest compile_synthesis revisit source was registered.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCompilePath, "Compile Revisit Routing", synthesisBody); err != nil {
		return err
	}
	decoyBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/compile-revisit-old.md
---
# Compile Revisit Routing Decoy

## Summary
This decoy page is not the compile_synthesis revisit decision target.

## Sources
- sources/compile-revisit-old.md

## Freshness
Checked decoy source only.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, synthesisCompileDecoyPath, "Compile Revisit Routing Decoy", decoyBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
type: source
status: active
supersedes: sources/compile-revisit-old.md
---
# Compile Revisit Current Source

## Summary
Current compile_synthesis revisit guidance says existing document and retrieval actions are technically sufficient when agents search source evidence, list synthesis candidates, inspect freshness and provenance, and update the existing synthesis without duplicates.
`) + "\n"
	return createSeedDocument(ctx, cfg, synthesisCompileCurrentSrc, "Compile Revisit Current Source", currentBody)
}
func seedAgentChosenSynthesisPathSelection(ctx context.Context, cfg runclient.Config) error {
	sourceBodies := map[string]string{
		agentChosenSynthesisAlphaPath: strings.TrimSpace(`---
type: source
status: active
path_pressure: agent-chosen
---
# Path Alpha

## Summary
Alpha source says agent-chosen path selection must preserve explicit-path compatibility.
`) + "\n",
		agentChosenSynthesisBetaPath: strings.TrimSpace(`---
type: source
status: active
path_pressure: agent-chosen
---
# Path Beta

## Summary
Beta source says metadata remains authoritative for document type and identity.
`) + "\n",
		agentChosenSynthesisGammaPath: strings.TrimSpace(`---
type: source
status: active
path_pressure: agent-chosen
---
# Path Gamma

## Summary
Gamma source says freshness, source refs, and citations must remain inspectable.
`) + "\n",
	}
	for _, path := range []string{agentChosenSynthesisAlphaPath, agentChosenSynthesisBetaPath, agentChosenSynthesisGammaPath} {
		if err := createSeedDocument(ctx, cfg, path, sourceTitleFromPath(path), sourceBodies[path]); err != nil {
			return err
		}
	}
	return nil
}
func seedPathTitleMultiSourceDuplicatePressure(ctx context.Context, cfg runclient.Config) error {
	sourceBodies := map[string]string{
		pathTitleSynthesisAlphaPath: strings.TrimSpace(`---
type: source
status: active
path_title_pressure: multi-source
---
# Path Title Alpha

## Summary
Alpha source says constrained autonomy must search sources before choosing a durable synthesis path.
`) + "\n",
		pathTitleSynthesisBetaPath: strings.TrimSpace(`---
type: source
status: active
path_title_pressure: multi-source
---
# Path Title Beta

## Summary
Beta source says constrained autonomy must update existing synthesis candidates instead of creating duplicates.
`) + "\n",
	}
	for _, path := range []string{pathTitleSynthesisAlphaPath, pathTitleSynthesisBetaPath} {
		if err := createSeedDocument(ctx, cfg, path, sourceTitleFromPath(path), sourceBodies[path]); err != nil {
			return err
		}
	}
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/path-title/alpha.md, sources/path-title/beta.md
---
# Path Title Autonomy

## Summary
Existing synthesis candidate for path/title autonomy pressure.

## Sources
- sources/path-title/alpha.md
- sources/path-title/beta.md

## Freshness
Fresh before autonomy pressure checks.
`) + "\n"
	return createSeedDocument(ctx, cfg, pathTitleSynthesisPath, pathTitleSynthesisTitle, body)
}
func seedPathTitleDuplicateRiskPressure(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: source
status: active
---
# Existing OpenAI Harness

## Summary
Existing source note for the OpenAI harness URL. Duplicate risk marker: existing path/title source should be reused, not copied.
`) + "\n"
	return createSeedDocument(ctx, cfg, pathTitleDuplicateExistingPath, "Existing OpenAI Harness", body)
}
func seedDocumentThisDuplicateCandidate(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: source
status: active
---
# Existing Document This Article

## Summary
Document-this duplicate marker: the article source already captures strict runner intake guidance.

## Sources
- https://example.test/articles/document-this-intake
`) + "\n"
	return createSeedDocument(ctx, cfg, documentThisDuplicateExistingPath, "Existing Document This Article", body)
}
func seedDocumentThisExistingUpdate(ctx context.Context, cfg runclient.Config) error {
	body := strings.TrimSpace(`---
type: note
status: active
---
# Existing Document This Update

## Summary
Existing update target for document-this intake pressure.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, documentThisUpdateTargetPath, "Existing Document This Update", body); err != nil {
		return err
	}
	decoy := strings.TrimSpace(`---
type: note
status: active
---
# Existing Document This Update Decoy

## Summary
Decoy note that must not receive the document-this update.
`) + "\n"
	return createSeedDocument(ctx, cfg, documentThisUpdateDecoyPath, "Existing Document This Update Decoy", decoy)
}
func seedDocumentThisSynthesisFreshness(ctx context.Context, cfg runclient.Config) error {
	sourceBodies := map[string]string{
		documentThisArticlePath: strings.TrimSpace(`---
type: source
status: active
source_kind: article
---
# Document This Article Source

## Summary
Article source says document-this intake should check duplicate candidates before creating durable notes.
`) + "\n",
		documentThisDocsPath: strings.TrimSpace(`---
type: source
status: active
source_kind: docs-page
---
# Document This Docs Page Source

## Summary
Docs page source says explicit path, title, and body are required before strict runner JSON can create a document.
`) + "\n",
		documentThisPaperPath: strings.TrimSpace(`---
type: source
status: active
source_kind: paper
---
# Document This Paper Source

## Summary
Paper source says provenance and projection freshness must remain inspectable for synthesis updates.
`) + "\n",
		documentThisTranscriptPath: strings.TrimSpace(`---
type: transcript
status: active
source_kind: transcript
---
# Document This Transcript

## Summary
Transcript source says mixed-source intake should update existing synthesis instead of creating duplicates.
`) + "\n",
	}
	for _, path := range []string{documentThisArticlePath, documentThisDocsPath, documentThisPaperPath, documentThisTranscriptPath} {
		if err := createSeedDocument(ctx, cfg, path, sourceTitleFromPath(path), sourceBodies[path]); err != nil {
			return err
		}
	}
	body := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/document-this/article.md, sources/document-this/docs-page.md, sources/document-this/paper.md, transcripts/document-this/standup.md
---
# Document This Intake

## Summary
Stale document-this intake summary that needs current mixed-source guidance.

## Sources
- sources/document-this/article.md
- sources/document-this/docs-page.md
- sources/document-this/paper.md
- transcripts/document-this/standup.md

## Freshness
Fresh before document-this intake pressure checks.
	`) + "\n"
	return createSeedDocument(ctx, cfg, documentThisSynthesisPath, "Document This Intake", body)
}
func seedMTSynthesisDriftPressure(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: sources/drift-current.md
---
# Drift Old Source

## Summary
Older drift guidance said synthesis compiler pressure should be promoted immediately.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, mtDriftOldSourcePath, "Drift Old Source", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
supersedes: sources/drift-old.md
---
# Drift Current Source

## Summary
Initial drift guidance is still under review.
`) + "\n"
	return createSeedDocument(ctx, cfg, mtDriftCurrentPath, "Drift Current Source", currentBody)
}
func seedSourceSensitiveAuditRepair(ctx context.Context, cfg runclient.Config) error {
	oldBody := strings.TrimSpace(`---
status: superseded
superseded_by: sources/audit-runner-current.md
---
# Audit Runner Old Source

## Summary
Older source-sensitive audit guidance said agents should prefer a legacy command-path workaround for runner audit repairs.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditOldSourcePath, "Audit Runner Old Source", oldBody); err != nil {
		return err
	}
	currentBody := strings.TrimSpace(`---
supersedes: sources/audit-runner-old.md
---
# Audit Runner Current Source

## Summary
Current source-sensitive audit guidance says agents must use the installed openclerk JSON runner, inspect provenance and projection freshness, and repair source-linked synthesis without duplicate pages.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditCurrentSourcePath, "Audit Runner Current Source", currentBody); err != nil {
		return err
	}
	synthesisBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/audit-runner-current.md, sources/audit-runner-old.md
---
# Audit Runner Routing

## Summary
Stale audit claim: agents should prefer a legacy command-path workaround for runner audit repairs.

## Sources
- sources/audit-runner-current.md
- sources/audit-runner-old.md

## Freshness
Checked before the current audit source was registered.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditSynthesisPath, "Audit Runner Routing", synthesisBody); err != nil {
		return err
	}
	decoyBody := strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: sources/audit-runner-old.md
---
# Audit Runner Decoy

## Summary
This decoy page is not the source-sensitive audit repair target.

## Sources
- sources/audit-runner-old.md

## Freshness
Checked decoy source only.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditDecoyPath, "Audit Runner Decoy", decoyBody); err != nil {
		return err
	}
	return replaceScenarioSeedSection(ctx, cfg, sourceAuditCurrentSourcePath, "Summary", "Current source-sensitive audit guidance says agents must use the installed openclerk JSON runner, inspect provenance and projection freshness, and repair source-linked synthesis without duplicate pages. "+sourceAuditOldSourcePath+" is superseded.")
}
func seedSourceSensitiveConflict(ctx context.Context, cfg runclient.Config) error {
	alphaBody := strings.TrimSpace(`---
type: source
audit_case: runner-retention
---
# Audit Conflict Alpha

## Summary
Alpha current source says source sensitive audit conflict runner retention should be seven days.
`) + "\n"
	if err := createSeedDocument(ctx, cfg, sourceAuditConflictAlphaPath, "Audit Conflict Alpha", alphaBody); err != nil {
		return err
	}
	bravoBody := strings.TrimSpace(`---
type: source
audit_case: runner-retention
---
# Audit Conflict Bravo

## Summary
Bravo current source says source sensitive audit conflict runner retention should be thirty days.
`) + "\n"
	return createSeedDocument(ctx, cfg, sourceAuditConflictBravoPath, "Audit Conflict Bravo", bravoBody)
}
func seedBroadContradictionAuditRevisit(ctx context.Context, cfg runclient.Config) error {
	if err := seedSourceSensitiveAuditRepair(ctx, cfg); err != nil {
		return err
	}
	return seedSourceSensitiveConflict(ctx, cfg)
}
