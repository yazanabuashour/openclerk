package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type demoKnowledgePackTemplate struct {
	Summary             string
	VaultFiles          map[string]string
	StaleAfterSyncPaths []string
}

var demoKnowledgePackTemplates = map[string]demoKnowledgePackTemplate{
	"codebase-decisions": {
		Summary: "seeded codebase decisions knowledge pack demo",
		VaultFiles: map[string]string{
			"sources/auth-callback.md": `---
type: source
status: active
---
# Auth Callback Source

## Summary
The auth callback validates the state value, exchanges the short-lived code,
and returns users to the requested repo-relative route.

## Notes
- Preserve the state value check before any redirect.
- Keep callback errors visible to the caller.
`,
			"sources/session-handling.md": `---
type: source
status: active
---
# Session Handling Source

## Summary
Session renewal happens after callback validation and before the user-visible
redirect.

## Notes
- Renewal failures should return a retryable error.
- Logging should avoid sensitive values.
`,
			"decisions/adr-auth-callback.md": `---
type: decision
decision_id: adr-auth-callback
decision_title: Keep callback validation before redirects
decision_status: accepted
decision_scope: auth
decision_owner: platform
decision_date: 2026-06-15
source_refs: sources/auth-callback.md, sources/session-handling.md
---
# Keep Callback Validation Before Redirects

## Decision
Validate callback state and renew the session before redirecting to a caller
route.

## Rationale
This keeps redirect behavior predictable and preserves a single validation
boundary for auth callback changes.
`,
		},
	},
	"stale-runbook": {
		Summary: "seeded stale runbook knowledge pack demo",
		VaultFiles: map[string]string{
			"sources/deploy-source.md": `---
type: source
status: active
---
# Deploy Source

## Summary
Deploys now use a two-person approval window and a post-deploy smoke check.

## Current Procedure
1. Confirm approval from the reviewer.
2. Run the deploy.
3. Run the smoke check.
`,
			"synthesis/deploy-runbook.md": `---
type: synthesis
status: active
freshness: fresh
source_refs: sources/deploy-source.md
---
# Deploy Runbook

## Summary
Deploys use one approval and a manual follow-up note.

## Sources
- sources/deploy-source.md

## Freshness
This runbook intentionally carries older guidance for stale-report demos.
`,
		},
		StaleAfterSyncPaths: []string{"sources/deploy-source.md"},
	},
	"research-vault": {
		Summary: "seeded research vault knowledge pack demo",
		VaultFiles: map[string]string{
			"sources/citation-quality.md": `---
type: source
status: active
---
# Citation Quality Notes

## Summary
Retrieval is useful when answers carry citations that agents can inspect before
acting.
`,
			"sources/freshness-checks.md": `---
type: source
status: active
---
# Freshness Checks Notes

## Summary
Derived summaries should expose whether source refs are fresh, stale, missing,
or superseded.
`,
			"synthesis/retrieval-notes.md": `---
type: synthesis
status: active
freshness: fresh
source_refs: sources/citation-quality.md, sources/freshness-checks.md
---
# Retrieval Notes

## Summary
Citation-bearing retrieval helps agents inspect evidence, and freshness checks
keep derived summaries honest.

## Sources
- sources/citation-quality.md
- sources/freshness-checks.md

## Freshness
Checked source refs for this example.
`,
		},
	},
	"incident-handoff": {
		Summary: "seeded incident handoff knowledge pack demo",
		VaultFiles: map[string]string{
			"incidents/cache-delay.md": `---
type: source
status: active
---
# Cache Delay Timeline

## Summary
A synthetic cache delay caused stale dashboard data for a small test window.

## Timeline
- 10:00: dashboard delay noticed.
- 10:10: cache key mismatch identified.
- 10:20: cache key normalized and dashboard refreshed.
`,
		},
	},
	"handwritten-inbox": {
		Summary: "seeded handwritten inbox knowledge pack demo",
		VaultFiles: map[string]string{
			"sources/review-policy.md": `---
type: source
status: active
---
# Review Policy Source

## Summary
Unclear notes and extracted text require human review before durable writes.
`,
		},
	},
	"agent-session-to-docs": {
		Summary: "seeded agent session to docs knowledge pack demo",
		VaultFiles: map[string]string{
			"sources/auth-callback-current.md": `---
type: source
status: active
---
# Auth Callback Current Notes

## Summary
The callback validates state before session renewal and returns retryable
errors for renewal failures.
`,
		},
	},
}

func availableDemoTemplates() []string {
	templates := make([]string, 0, len(demoKnowledgePackTemplates))
	for name := range demoKnowledgePackTemplates {
		templates = append(templates, name)
	}
	sort.Strings(templates)
	return templates
}

func writeDemoKnowledgePackTemplate(root string, template demoKnowledgePackTemplate) error {
	paths := make([]string, 0, len(template.VaultFiles))
	for path := range template.VaultFiles {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		target := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create template directory: %w", err)
		}
		if err := os.WriteFile(target, []byte(normalizeDemoTemplateFile(template.VaultFiles[path])), 0o644); err != nil {
			return fmt.Errorf("write template file %s: %w", path, err)
		}
	}
	return nil
}

func normalizeDemoTemplateFile(content string) string {
	return strings.TrimSpace(content) + "\n"
}
