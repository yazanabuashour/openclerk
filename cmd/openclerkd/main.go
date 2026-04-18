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
		fmt.Fprintln(os.Stderr, "usage: openclerkd serve --db <path> --vault-root <path> [--addr 127.0.0.1:8080]")
		os.Exit(2)
	}

	command := flag.NewFlagSet("serve", flag.ContinueOnError)
	command.SetOutput(os.Stderr)
	dbPath := command.String("db", "", "path to the sqlite database file")
	vaultRoot := command.String("vault-root", "", "path to the canonical markdown vault root")
	addr := command.String("addr", "127.0.0.1:8080", "listen address")
	embeddingProvider := command.String("embedding-provider", "", "embedding provider name; use 'local' to enable local hashed embeddings")
	if err := command.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
	}

	ctx := context.Background()
	store, err := sqlite.New(ctx, sqlite.Config{
		Backend:           domain.BackendOpenClerk,
		DatabasePath:      *dbPath,
		VaultRoot:         *vaultRoot,
		EmbeddingProvider: *embeddingProvider,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = store.Close()
	}()

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
