package main

import (
	_ "embed"
	"fmt"
	"github.com/bmorton/flipull/cmd"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:     "flipull",
		Usage:    "A tool for automating the creation of pull requests",
		Commands: []*cli.Command{cmd.ReplaceCommand},
	}

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
