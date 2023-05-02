package main

import (
	"os"
	"styra.com/styrainc/odm/cmd"
)

func main() {
	if err := cmd.RootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
