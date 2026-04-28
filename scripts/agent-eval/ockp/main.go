package main

import (
	"context"
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, codexJobRunner))
}
func run(args []string, stdout io.Writer, stderr io.Writer, runner jobRunner) int {
	if len(args) == 0 || args[0] != "run" {
		_, _ = fmt.Fprintln(stderr, "usage: ockp run [--parallel N] [--variant ids] [--scenario ids] [--run-root path] [--report-dir path] [--codex-bin path] [--cache-mode shared|isolated]")
		return 2
	}
	config, err := parseRunConfig(args[1:], stderr)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	if err := executeRun(context.Background(), config, stdout, runner); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
