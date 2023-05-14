package cmd

import (
	"fmt"
	"github.com/johanfylling/odm/proj"
	"github.com/johanfylling/odm/utils"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	var sourceDir string
	var noSource bool

	var initCommand = &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new OPA project",
		Run: func(cmd *cobra.Command, args []string) {
			path := "."
			var name string
			if len(args) == 1 {
				name = args[0]
				path = fmt.Sprintf("./%s", name)
			}
			if noSource {
				sourceDir = ""
			}
			if err := doInit(path, name, sourceDir); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	initCommand.Flags().StringVarP(&sourceDir, "source", "s", "src", "source directory for the project. Mutually exclusive with --no-source")
	initCommand.Flags().BoolVarP(&noSource, "no-source", "", false, "don't assign a source directory for the project. Mutually exclusive with --source")

	RootCommand.AddCommand(initCommand)
}

func doInit(path string, name string, sourceDir string) error {
	fmt.Println("initializing OPA project")

	project := proj.Project{
		Name: name,
	}

	if sourceDir != "" {
		project.SourceDir = sourceDir
	}

	if !utils.FileExists(path) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("error creating project directory: %s", err)
		}
	} else {
		fmt.Printf("directory %s already exists, not creating new\n", path)
	}

	err := project.WriteToFile(path, false) //createProjectFile(path, project)
	if err != nil {
		return err
	}

	err = createDotOpaDirectory(path)
	if err != nil {
		return err
	}

	return nil
}

// create a new .opa directory in working directory
func createDotOpaDirectory(path string) error {
	path = path + "/.opa"

	// check if .opa directory already exists
	if utils.FileExists(path) {
		fmt.Printf("directory %s already exists, not creating new\n", path)
		return nil
	}

	// create directory at path
	fmt.Printf("creating directory %s\n", path)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		//fmt.Printf("error creating .opa directory: %s\n", err)
		return fmt.Errorf("error creating .opa directory: %s", err)
	}
	return nil
}
