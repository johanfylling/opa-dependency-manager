package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"styra.com/styrainc/odm/proj"
)

func init() {
	var namespace string
	var version string

	var depCommand = &cobra.Command{
		Use:   "dep <location> [flags]",
		Short: "Add a dependency to the project",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("expected exactly one dependency")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			dep := args[0]

			projPath := "."

			if err := doAddDependency(dep, namespace, version, projPath); err != nil {
				_, _ = cmd.OutOrStderr().Write([]byte(err.Error()))
				os.Exit(1)
			}
		},
	}

	depCommand.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace of the dependency")
	depCommand.Flags().StringVarP(&version, "version", "v", "", "version of the dependency")

	RootCommand.AddCommand(depCommand)
}

func doAddDependency(location string, namespace string, version string, projectPath string) error {
	project, err := proj.ReadProjectFromFile(projectPath)
	if err != nil {
		return err
	}

	dependency := proj.DependencyInfo{
		Namespace: namespace,
		Version:   version,
	}
	project.SetDependency(location, dependency)
	//project.Dependencies = append(project.Dependencies, dependency)

	return project.WriteToFile(projectPath, true)
}
