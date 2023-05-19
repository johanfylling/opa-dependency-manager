package cmd

import (
	"fmt"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/proj"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	var namespace string
	var namespaced bool

	var depCommand = &cobra.Command{
		Use:   "depend <name> <location> [flags]",
		Short: "Add a dependency to the project",
		Long: `Add a dependency to the project

Supported location types:
- Git repository: git+http://..., git+https://..., git+ssh://...
- Local file/directory: file://path/to/dir, file:/../path/to/dir

Example:`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("expected exactly one dependency name and one location")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			location := args[1]

			projPath := "."

			if namespaced && namespace == "" {
				namespace = name
			}

			if err := doAddDependency(name, location, namespace, projPath); err != nil {
				_, _ = cmd.OutOrStderr().Write([]byte(err.Error()))
				os.Exit(1)
			}
		},
	}

	depCommand.Flags().BoolVarP(&namespaced, "namespaced", "n", false, "use name of dependency as namespace. Ignored if --namespace is set")
	depCommand.Flags().StringVarP(&namespace, "namespace", "N", "", "namespace of the dependency")

	RootCommand.AddCommand(depCommand)
}

func doAddDependency(name string, location string, namespace string, projectPath string) error {
	printer.Trace("--- Dep start ---")
	defer printer.Trace("--- Dep end ---")

	printer.Info("Setting dependency '%s' @ '%s'", name, location)

	project, err := proj.ReadProjectFromFile(projectPath, false)
	if err != nil {
		return err
	}

	dependency := proj.DependencyInfo{
		Namespace: namespace,
		Location:  location,
	}

	project.SetDependency(name, dependency)

	return project.WriteToFile(projectPath, true)
}
