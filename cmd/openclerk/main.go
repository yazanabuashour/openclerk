package main

import (
	"fmt"
	"os"

	"github.com/yazanabuashour/openclerk/internal/cli"
)

func main() {
	if _, err := fmt.Fprintln(os.Stdout, cli.Message()); err != nil {
		os.Exit(1)
	}
}
