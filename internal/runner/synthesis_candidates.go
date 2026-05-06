package runner

import (
	"context"
	"fmt"
	"sort"

	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/runclient"
)

type synthesisCandidateInspection struct {
	Paths         []string
	TargetMatches []domain.DocumentSummary
}

func inspectSynthesisCandidates(ctx context.Context, client *runclient.Client, targetPath string) (synthesisCandidateInspection, error) {
	inspection := synthesisCandidateInspection{}
	cursor := ""
	for {
		list, err := client.ListDocuments(ctx, domain.DocumentListQuery{
			PathPrefix: "synthesis/",
			Limit:      100,
			Cursor:     cursor,
		})
		if err != nil {
			return synthesisCandidateInspection{}, err
		}
		for _, document := range list.Documents {
			inspection.Paths = appendUniqueString(inspection.Paths, document.Path)
			if document.Path == targetPath {
				inspection.TargetMatches = append(inspection.TargetMatches, document)
			}
		}
		if !list.PageInfo.HasMore {
			break
		}
		cursor = list.PageInfo.NextCursor
		if cursor == "" {
			return synthesisCandidateInspection{}, fmt.Errorf("list synthesis candidates did not return next cursor")
		}
	}
	sort.Strings(inspection.Paths)
	return inspection, nil
}
