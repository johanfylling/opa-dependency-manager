package main

import (
	"github.com/johanfylling/odm/cmd"
	"os"
)

func main() {
	if err := cmd.RootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
