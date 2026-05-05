package runner

import (
	"net"
	"net/url"
	"strings"
)

func validateRequiredRunnerHTTPURL(raw string, field string) (*url.URL, string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, field + " is required"
	}
	return validateOptionalRunnerHTTPURL(trimmed, field)
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
	return parsed, ""
}

func validateOptionalRunnerLoopbackHTTPURL(raw string, field string) string {
	parsed, rejection := validateOptionalRunnerHTTPURL(raw, field)
	if rejection != "" {
		return field + " must be a loopback HTTP URL"
	}
	if parsed == nil {
		return ""
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
