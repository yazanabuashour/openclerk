package sqlite

import (
	"bytes"
	"context"
	"fmt"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"strings"
	"testing"
	"time"
)

func openTestStore(t *testing.T, backend domain.BackendKind, dbPath string, vaultRoot string) *Store {
	t.Helper()

	store, err := New(context.Background(), Config{
		Backend:      backend,
		DatabasePath: dbPath,
		VaultRoot:    vaultRoot,
	})
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	return store
}

func testClock() time.Time {
	return time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
}

func synthesisBody(sourceRefs string, summary string) string {
	return strings.TrimSpace(`---
type: synthesis
status: active
freshness: fresh
source_refs: `+sourceRefs+`
---
# Synthesis

## Summary
`+summary+`

## Sources
- `+sourceRefs+`

## Freshness
Checked source refs.
`) + "\n"
}

func hasEventType(events []domain.ProvenanceEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func countEventType(events []domain.ProvenanceEvent, eventType string) int {
	count := 0
	for _, event := range events {
		if event.EventType == eventType {
			count++
		}
	}
	return count
}

func minimalStorePDF(title string, author string, text string) []byte {
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
