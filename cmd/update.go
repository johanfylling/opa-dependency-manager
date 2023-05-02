package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"styra.com/styrainc/odm/proj"
	"styra.com/styrainc/odm/utils"
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
	project, err := proj.ReadProjectFromFile(projectPath, false)
	if err != nil {
		return err
	}

	rootDir := ".opa"
	depRootDir := fmt.Sprintf("%s/dependencies", rootDir)

	if !utils.FileExists(rootDir) {
		if err := os.Mkdir(rootDir, 0755); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(depRootDir); err != nil {
		return err
	}

	if err := os.Mkdir(depRootDir, 0755); err != nil {
		return err
	}

	for _, dependency := range project.Dependencies {
		if err := dependency.Update(depRootDir); err != nil {
			return err
		}
	}

	return nil
}
