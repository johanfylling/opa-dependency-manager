package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"path"
)

var RootCommand = &cobra.Command{
	Use:   path.Base(os.Args[0]),
	Short: "OPA Dependency Manager (ODM)",
}
