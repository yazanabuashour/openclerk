package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/yazanabuashour/openclerk/internal/api"
	"github.com/yazanabuashour/openclerk/internal/app"
	"github.com/yazanabuashour/openclerk/internal/domain"
	"github.com/yazanabuashour/openclerk/internal/infra/sqlite"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Fprintln(os.Stderr, "usage: openclerkd serve --backend=<fts|hybrid|graph|records> --db <path> --vault-root <path> [--addr 127.0.0.1:8080]")
		os.Exit(2)
	}

	command := flag.NewFlagSet("serve", flag.ExitOnError)
	backendFlag := command.String("backend", "fts", "backend kind: fts, hybrid, graph, or records")
	dbPath := command.String("db", "", "path to the sqlite database file")
	vaultRoot := command.String("vault-root", "", "path to the canonical markdown vault root")
	addr := command.String("addr", "127.0.0.1:8080", "listen address")
	embeddingProvider := command.String("embedding-provider", "", "embedding provider name; use 'local' to enable local hashed embeddings")
	command.Parse(os.Args[2:])

	backend, err := parseBackend(*backendFlag)
	if err != nil {
		log.Fatalf("invalid backend: %v", err)
	}

	ctx := context.Background()
	store, err := sqlite.New(ctx, sqlite.Config{
		Backend:           backend,
		DatabasePath:      *dbPath,
		VaultRoot:         *vaultRoot,
		EmbeddingProvider: *embeddingProvider,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	service := app.New(store)
	server := &http.Server{
		Addr:    *addr,
		Handler: api.NewHandler(service),
	}

	go func() {
		<-shutdownSignal()
		_ = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func shutdownSignal() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}

func parseBackend(value string) (domain.BackendKind, error) {
	switch domain.BackendKind(value) {
	case domain.BackendFTS, domain.BackendHybrid, domain.BackendGraph, domain.BackendRecords:
		return domain.BackendKind(value), nil
	default:
		return "", fmt.Errorf("unsupported backend %q", value)
	}
}
