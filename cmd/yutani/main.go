// Package main provides the Yutani CLI tool.
package main

import (
	"fmt"
	"os"

	"github.com/chazu/yutani/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

