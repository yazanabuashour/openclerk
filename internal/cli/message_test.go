package cli

import "testing"

func TestMessage(t *testing.T) {
	t.Parallel()

	const want = "openclerk bootstrap is wired and ready for development"

	if got := Message(); got != want {
		t.Fatalf("Message() = %q, want %q", got, want)
	}
}
