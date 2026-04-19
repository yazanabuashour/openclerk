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
// The local SDK calls the embedded runtime directly. It does not bind a port or
// require a generated HTTP client.
package local
