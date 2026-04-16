// Package local opens OpenClerk as an embedded runtime inside the caller's Go process.
//
// Until the first release tag is published, install the current development line:
//
//	go get github.com/yazanabuashour/openclerk/client/local@main
//
// Most callers use OpenClient to obtain the code-first local SDK facade:
//
//	client, err := local.OpenClient(local.Config{})
//	defer client.Close()
//
// Generated request and response types live in package
// github.com/yazanabuashour/openclerk/client/openclerk, which is part of the
// same module and does not require a second go get step. Use Open or
// Client.Generated only when raw OpenAPI response handling is required.
//
// The normal user path is embedded and does not bind a port. Use cmd/openclerkd
// and a remote client only for intentional HTTP debugging or compatibility
// work.
package local
