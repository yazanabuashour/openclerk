package main

import (
	"context"
	"strings"

	"github.com/yazanabuashour/openclerk/internal/runclient"
)

func seedTaggingWorkflows(ctx context.Context, cfg runclient.Config) error {
	docs := []struct {
		path  string
		title string
		tag   string
		body  string
	}{
		{
			path:  taggingRetrievalPath,
			title: taggingRetrievalTitle,
			tag:   taggingAccountRenewalTag,
			body:  "Account renewal tagging evidence says renewal packaging depends on customer notice timing.",
		},
		{
			path:  taggingRetrievalDecoyPath,
			title: taggingRetrievalDecoyTitle,
			tag:   taggingRenewalTag,
			body:  "Security renewal tagging decoy evidence must not satisfy account-renewal lookup.",
		},
		{
			path:  taggingDisambiguationTargetPath,
			title: "Customer Risk",
			tag:   taggingCustomerRiskTag,
			body:  "Tagging exact customer risk evidence belongs to the active customer-risk tag.",
		},
		{
			path:  taggingDisambiguationDecoyPath,
			title: "Customer Risk Archive",
			tag:   taggingCustomerRiskArchiveTag,
			body:  "Tagging exact customer risk archive evidence is a separate tag and must be excluded.",
		},
		{
			path:  taggingNearDuplicateTargetPath,
			title: "Ops Review",
			tag:   taggingOpsReviewTag,
			body:  "Tagging near duplicate operations review evidence belongs to the singular ops-review tag.",
		},
		{
			path:  taggingNearDuplicateDecoyPath,
			title: "Ops Reviews",
			tag:   taggingOpsReviewsTag,
			body:  "Tagging near duplicate operations reviews evidence belongs to the plural ops-reviews tag.",
		},
		{
			path:  taggingMixedPathTargetPath,
			title: "Support Handoff",
			tag:   taggingSupportHandoffTag,
			body:  "Tagging support handoff active note evidence belongs under active notes.",
		},
		{
			path:  taggingMixedPathArchivePath,
			title: "Archived Support Handoff",
			tag:   taggingSupportHandoffTag,
			body:  "Tagging support handoff archived note evidence must be excluded by path prefix.",
		},
	}
	for _, doc := range docs {
		if err := createSeedDocument(ctx, cfg, doc.path, doc.title, taggedSeedBody(doc.title, doc.tag, doc.body)); err != nil {
			return err
		}
	}
	return nil
}

func taggedSeedBody(title string, tag string, body string) string {
	return strings.TrimSpace(`---
type: note
tag: `+tag+`
---
# `+title+`

## Summary
`+body+`
`) + "\n"
}
