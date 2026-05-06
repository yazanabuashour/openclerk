package runner

import (
	"strings"
	"testing"
)

func TestLimitedOCRBufferRejectsOversizedOutput(t *testing.T) {
	t.Parallel()

	var buffer limitedOCRBuffer
	buffer.limit = 8
	if _, err := buffer.Write([]byte(strings.Repeat("x", 9))); err == nil || !strings.Contains(err.Error(), "maximum supported text size") {
		t.Fatalf("limited OCR buffer error = %v, want size rejection", err)
	}
}
