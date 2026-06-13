package runner

import (
	"net/url"
	"testing"
)

func TestRunnerHTTPURLValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		validate func(string, string) (*url.URL, string)
		raw      string
		want     string
		wantPath string
	}{
		{
			name:     "required URL rejects missing",
			validate: validateRequiredRunnerHTTPURL,
			want:     "source.url is required",
		},
		{
			name:     "required URL accepts public HTTP URL",
			validate: validateRequiredRunnerHTTPURL,
			raw:      "https://example.test/source.pdf",
			wantPath: "/source.pdf",
		},
		{
			name:     "public URL rejects file URL",
			validate: validateOptionalRunnerHTTPURL,
			raw:      "file:///tmp/source.pdf",
			want:     "source.url must be a valid http or https URL",
		},
		{
			name:     "public URL rejects mailto URL",
			validate: validateOptionalRunnerHTTPURL,
			raw:      "mailto:user@example.test",
			want:     "source.url must be a valid http or https URL",
		},
		{
			name:     "public URL rejects unsupported scheme",
			validate: validateOptionalRunnerHTTPURL,
			raw:      "ftp://example.test/source.pdf",
			want:     "source.url must use http or https",
		},
		{
			name:     "public URL rejects userinfo",
			validate: validateOptionalRunnerHTTPURL,
			raw:      "https://user:pass@example.test/source.pdf",
			want:     "source.url must not include userinfo",
		},
		{
			name:     "optional URL accepts missing",
			validate: validateOptionalRunnerHTTPURL,
		},
		{
			name:     "public URL rejects loopback host",
			validate: validateOptionalRunnerHTTPURL,
			raw:      "http://127.0.0.1/source.pdf",
			want:     "source.url must be publicly fetchable",
		},
		{
			name:     "syntax-only URL accepts loopback host",
			validate: validateRequiredRunnerHTTPURLSyntax,
			raw:      "http://127.0.0.1/source.pdf",
			wantPath: "/source.pdf",
		},
		{
			name:     "syntax-only URL rejects userinfo",
			validate: validateRequiredRunnerHTTPURLSyntax,
			raw:      "https://user:pass@example.test/source.pdf",
			want:     "source.url must not include userinfo",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, rejection := tt.validate(tt.raw, "source.url")
			if rejection != tt.want {
				t.Fatalf("rejection = %q, want %q", rejection, tt.want)
			}
			if tt.wantPath != "" && (parsed == nil || parsed.Path != tt.wantPath) {
				t.Fatalf("parsed = %#v, want path %q", parsed, tt.wantPath)
			}
		})
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
	if rejection := validateOptionalRunnerLoopbackHTTPURL("http://user:pass@localhost:11434", "semantic_search.ollama_url"); rejection != "semantic_search.ollama_url must be a loopback HTTP URL" {
		t.Fatalf("userinfo loopback rejection = %q", rejection)
	}
}

func TestRunnerGeminiAPIBaseValidation(t *testing.T) {
	t.Parallel()

	if rejection := validateOptionalRunnerCanonicalGeminiAPIBase("https://generativelanguage.googleapis.com/v1beta", "semantic_search.gemini_api_base"); rejection != "" {
		t.Fatalf("canonical Gemini base rejection = %q", rejection)
	}
	for _, raw := range []string{
		"http://127.0.0.1:9999",
		"https://generativelanguage.googleapis.com/v1beta?key=inline",
		"https://user:pass@generativelanguage.googleapis.com/v1beta",
		"https://generativelanguage.googleapis.com/v1",
	} {
		if rejection := validateOptionalRunnerCanonicalGeminiAPIBase(raw, "semantic_search.gemini_api_base"); rejection != "semantic_search.gemini_api_base must be https://generativelanguage.googleapis.com/v1beta" {
			t.Fatalf("Gemini base %q rejection = %q", raw, rejection)
		}
	}
}
