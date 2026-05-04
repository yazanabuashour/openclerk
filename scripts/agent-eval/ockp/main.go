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
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, "usage: ockp run ... | ockp maturity scale-ladder|real-vault ... | ockp semantic-recall ...")
		return 2
	}
	switch args[0] {
	case "run":
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
	case "maturity":
		config, err := parseMaturityConfig(args[1:], stderr)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		if err := executeMaturity(context.Background(), config, stdout); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	case "semantic-recall":
		config, err := parseSemanticRecallConfig(args[1:], stderr)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		if err := executeSemanticRecall(context.Background(), config, stdout); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	default:
		_, _ = fmt.Fprintln(stderr, "usage: ockp run ... | ockp maturity scale-ladder|real-vault ... | ockp semantic-recall ...")
		return 2
	}
}
