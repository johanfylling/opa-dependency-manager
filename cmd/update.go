package cmd

import (
	"fmt"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/proj"
	"github.com/johanfylling/odm/utils"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func init() {
	var updateCommand = &cobra.Command{
		Use:   "update",
		Short: "Update OPA project dependencies",
		Run: func(cmd *cobra.Command, args []string) {
			projPath := "."

			if err := doUpdate(projPath); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	RootCommand.AddCommand(updateCommand)
}

func doUpdate(projectPath string) error {
	printer.Trace("--- Project update start ---")
	defer printer.Trace("--- Project update end ---")

	project, err := proj.ReadProjectFromFile(projectPath, false)
	if err != nil {
		return err
	}

	printer.Info("Updating project '%s'", project.Name)

	dotOpaDir := filepath.Join(project.Dir(), ".opa")
	depRootDir := fmt.Sprintf("%s/dependencies", dotOpaDir)

	if !utils.FileExists(dotOpaDir) {
		if err := os.Mkdir(dotOpaDir, 0755); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(depRootDir); err != nil {
		return err
	}

	if err := os.Mkdir(depRootDir, 0755); err != nil {
		return err
	}

	if err := project.Update(); err != nil {
		return err
	}

	return nil
}
