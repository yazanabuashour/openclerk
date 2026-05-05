package runner

import "testing"

func TestRunnerHTTPURLValidation(t *testing.T) {
	t.Parallel()

	if _, rejection := validateRequiredRunnerHTTPURL("", "source.url"); rejection != "source.url is required" {
		t.Fatalf("missing required URL rejection = %q", rejection)
	}
	parsed, rejection := validateRequiredRunnerHTTPURL("https://example.test/source.pdf", "source.url")
	if rejection != "" || parsed == nil || parsed.Path != "/source.pdf" {
		t.Fatalf("valid required URL = %#v, %q; want parsed URL, empty rejection", parsed, rejection)
	}
	if _, rejection := validateOptionalRunnerHTTPURL("file:///tmp/source.pdf", "source.url"); rejection != "source.url must be a valid http or https URL" {
		t.Fatalf("absolute file URL rejection = %q", rejection)
	}
	if _, rejection := validateOptionalRunnerHTTPURL("mailto:user@example.test", "source.url"); rejection != "source.url must be a valid http or https URL" {
		t.Fatalf("mailto URL rejection = %q", rejection)
	}
	if _, rejection := validateOptionalRunnerHTTPURL("ftp://example.test/source.pdf", "source.url"); rejection != "source.url must use http or https" {
		t.Fatalf("unsupported scheme rejection = %q", rejection)
	}
	if _, rejection := validateOptionalRunnerHTTPURL("", "source.url"); rejection != "" {
		t.Fatalf("empty optional URL rejection = %q, want empty", rejection)
	}
}

func TestRunnerLoopbackHTTPURLValidation(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"http://localhost:11434", "http://127.0.0.1:11434", "http://[::1]:11434"} {
		if rejection := validateOptionalRunnerLoopbackHTTPURL(raw, "semantic_search.ollama_url"); rejection != "" {
			t.Fatalf("loopback URL %q rejection = %q, want empty", raw, rejection)
		}
	}
	if rejection := validateOptionalRunnerLoopbackHTTPURL("https://embeddings.example.test", "semantic_search.ollama_url"); rejection != "semantic_search.ollama_url must be a loopback HTTP URL" {
		t.Fatalf("remote loopback rejection = %q", rejection)
	}
	if rejection := validateOptionalRunnerLoopbackHTTPURL("file:///tmp/ollama.sock", "semantic_search.ollama_url"); rejection != "semantic_search.ollama_url must be a loopback HTTP URL" {
		t.Fatalf("invalid loopback URL rejection = %q", rejection)
	}
}
