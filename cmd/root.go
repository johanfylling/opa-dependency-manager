package cmd

import (
	"github.com/johanfylling/odm/printer"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var RootCommand = &cobra.Command{
	Use:   path.Base(os.Args[0]),
	Short: "OPA Dependency Manager (ODM)",
}

func init() {
	// Add verbose flag to all commands
	RootCommand.PersistentFlags().CountVarP(&printer.LogLevel, "verbose", "v", "verbose output")
}

func addNoUpdateFlag(cmd *cobra.Command, v *bool) {
	cmd.Flags().BoolVar(v, "no-update", false, "do not sync dependencies before executing this command")
}
