package main

import (
	"context"
	"os"

	"github.com/Team-NEEEE/envio-cli/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(cli.Run(context.Background(), cli.Runtime{
		Args:    os.Args[1:],
		Environ: os.Environ(),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		CWD:     mustGetwd(),
		Version: version,
		Commit:  commit,
		Date:    date,
	}))
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
