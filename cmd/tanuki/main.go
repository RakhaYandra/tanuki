package main

import (
	"fmt"
	"os"

	"github.com/RakhaYandra/tanuki/internal/cli"
	tanukierr "github.com/RakhaYandra/tanuki/internal/errors"
)

var version = "dev"

func main() {
	cli.Version = version
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(tanukierr.CodeOf(err))
	}
}
