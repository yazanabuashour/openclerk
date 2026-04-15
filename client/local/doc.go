// Package local opens OpenClerk as an embedded runtime inside the caller's Go process.
//
// Install the tagged module with:
//
//	go get github.com/yazanabuashour/openclerk/client/local@v0.1.0
//
// Most callers use Open to obtain an in-process client and runtime:
//
//	client, runtime, err := local.Open(local.Config{})
//
// Generated request and response types live in package
// github.com/yazanabuashour/openclerk/client/openclerk, which is part of the
// same module and does not require a second go get step.
//
// The normal user path is embedded and does not bind a port. Use cmd/openclerkd
// and a remote client only for intentional HTTP debugging or compatibility
// work.
package local
