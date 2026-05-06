package main

import (
	"context"
	"os"

	"github.com/Team-NEEEE/envio-cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(context.Background(), cli.Runtime{
		Args:    os.Args[1:],
		Environ: os.Environ(),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		CWD:     mustGetwd(),
	}))
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
