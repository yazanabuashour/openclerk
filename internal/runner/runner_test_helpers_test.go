package runner_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/runclient"
	"github.com/yazanabuashour/openclerk/internal/runner"
	"strings"
	"testing"
)

func createDocument(t *testing.T, ctx context.Context, config runclient.Config, path string, title string, body string) runner.Document {
	t.Helper()
	result, err := runner.RunDocumentTask(ctx, config, runner.DocumentTaskRequest{
		Action: runner.DocumentTaskActionCreate,
		Document: runner.DocumentInput{
			Path:  path,
			Title: title,
			Body:  body,
		},
	})
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	if result.Document == nil {
		t.Fatalf("create %s result = %+v", path, result)
	}
	return *result.Document
}

func isClearConcurrencyConflict(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "already exists") || strings.Contains(message, "conflict")
}

func runnerEventTypesInclude(events []runner.ProvenanceEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func graphContextContainsPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func documentLinksContainRunnerPath(links []runner.DocumentLink, path string) bool {
	for _, link := range links {
		if link.Path == path && len(link.Citations) > 0 {
			return true
		}
	}
	return false
}

func graphContextEdgesHaveCitations(edges []runner.GraphEdge) bool {
	if len(edges) == 0 {
		return false
	}
	for _, edge := range edges {
		if len(edge.Citations) == 0 {
			return false
		}
	}
	return true
}

func graphRelationshipPathContains(paths []runner.GraphRelationshipPath, direction string, path string) bool {
	for _, candidate := range paths {
		if candidate.Direction == direction && candidate.Path == path && len(candidate.Citations) > 0 {
			return true
		}
	}
	return false
}

func graphRelationshipEvidenceContains(evidence []runner.GraphRelationshipEvidence, relationshipType string) bool {
	for _, candidate := range evidence {
		if candidate.RelationshipType == relationshipType && len(candidate.Citations) > 0 {
			return true
		}
	}
	return false
}

func graphRelationshipTypedCandidateContains(candidates []runner.GraphRelationshipTypeCandidate, relationshipType string) bool {
	for _, candidate := range candidates {
		if candidate.RelationshipType == relationshipType && candidate.Citation.Path != "" {
			return true
		}
	}
	return false
}

func graphRelationshipAuditFinding(findings []runner.GraphRelationshipAuditFinding, kind string, status string) bool {
	for _, finding := range findings {
		if finding.Kind == kind && finding.Status == status {
			return true
		}
	}
	return false
}

func graphRelationshipCandidatesInclude(candidates []runner.GraphRelationshipCandidate, surface string) bool {
	for _, candidate := range candidates {
		if candidate.Surface == surface {
			return true
		}
	}
	return false
}

func graphRelationshipMaintenanceActionContains(actions []runner.GraphRelationshipMaintenanceAction, kind string, status string) bool {
	for _, action := range actions {
		if action.Kind == kind && action.Status == status {
			return true
		}
	}
	return false
}

func graphRelationshipMaintenanceCandidatesInclude(candidates []runner.GraphRelationshipMaintenanceCandidate, surface string) bool {
	for _, candidate := range candidates {
		if candidate.Surface == surface {
			return true
		}
	}
	return false
}

func auditInspectedPath(inspections []runner.AuditProvenanceInspection, path string) bool {
	for _, inspection := range inspections {
		if inspection.SourcePath == path && len(inspection.EventIDs) > 0 {
			return true
		}
	}
	return false
}

func minimalPDF(title string, author string, text string) []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, 6)
	writeObject := func(id int, body string) {
		offsets = append(offsets, buf.Len())
		_, _ = fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", id, body)
	}
	writeObject(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObject(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObject(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>")
	writeObject(4, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	stream := fmt.Sprintf("BT /F1 24 Tf 72 720 Td (%s) Tj ET", pdfEscape(text))
	writeObject(5, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream))
	writeObject(6, fmt.Sprintf("<< /Title (%s) /Author (%s) /CreationDate (D:20260426000000Z) >>", pdfEscape(title), pdfEscape(author)))
	xrefStart := buf.Len()
	buf.WriteString("xref\n0 7\n")
	buf.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		_, _ = fmt.Fprintf(&buf, "%010d 00000 n \n", offset)
	}
	_, _ = fmt.Fprintf(&buf, "trailer\n<< /Size 7 /Root 1 0 R /Info 6 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefStart)
	return buf.Bytes()
}

func pdfEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "(", `\(`)
	value = strings.ReplaceAll(value, ")", `\)`)
	return value
}
