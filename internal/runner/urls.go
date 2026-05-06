package runner

import (
	"net"
	"net/url"
	"os"
	"strings"
)

const canonicalGeminiAPIBase = "https://generativelanguage.googleapis.com/v1beta"

func validateRequiredRunnerHTTPURL(raw string, field string) (*url.URL, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, field + " is required"
	}
	return validateOptionalRunnerHTTPURL(trimmed, field)
}

func validateRequiredRunnerHTTPURLSyntax(raw string, field string) (*url.URL, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, field + " is required"
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, field + " must be a valid http or https URL"
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, field + " must use http or https"
	}
	return parsed, ""
}

func validateOptionalRunnerHTTPURL(raw string, field string) (*url.URL, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, field + " must be a valid http or https URL"
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, field + " must use http or https"
	}
	if rejection := validateRunnerPublicURLHost(parsed, field); rejection != "" {
		return nil, rejection
	}
	return parsed, ""
}

func validateRunnerPublicURLHost(parsed *url.URL, field string) string {
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return field + " must be a valid http or https URL"
	}
	lowerHost := strings.ToLower(host)
	if lowerHost == "openclerk-eval.local" &&
		strings.TrimSpace(os.Getenv("OPENCLERK_ENABLE_EVAL_SOURCE_FIXTURES")) == "1" &&
		strings.TrimSpace(os.Getenv("OPENCLERK_EVAL_SOURCE_FIXTURE_ROOT")) != "" {
		return ""
	}
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") || strings.HasSuffix(lowerHost, ".local") {
		return field + " must be publicly fetchable"
	}
	if ip := net.ParseIP(host); ip != nil && !isRunnerPublicIP(ip) {
		return field + " must be publicly fetchable"
	}
	return ""
}

func isRunnerPublicIP(ip net.IP) bool {
	return ip != nil &&
		!ip.IsUnspecified() &&
		!ip.IsLoopback() &&
		!ip.IsPrivate() &&
		!ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() &&
		!ip.IsMulticast()
}

func validateOptionalRunnerLoopbackHTTPURL(raw string, field string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return field + " must be a loopback HTTP URL"
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return ""
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return ""
	}
	return field + " must be a loopback HTTP URL"
}

func validateOptionalRunnerCanonicalGeminiAPIBase(raw string, field string) string {
	parsed, rejection := validateOptionalRunnerHTTPURL(raw, field)
	if rejection != "" {
		return field + " must be " + canonicalGeminiAPIBase
	}
	if parsed == nil {
		return ""
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" ||
		parsed.Scheme != "https" || parsed.Host != "generativelanguage.googleapis.com" || parsed.Path != "/v1beta" {
		return field + " must be " + canonicalGeminiAPIBase
	}
	return ""
}
