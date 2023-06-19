package cmd

import (
	"fmt"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/proj"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func init() {
	var noUpdate bool
	var includeTestDirs bool
	var includeDepTests bool

	var listCommand = &cobra.Command{
		Use:   "list",
		Short: "List project resources",
	}

	addNoUpdateFlag(listCommand, &noUpdate)
	RootCommand.AddCommand(listCommand)

	var listSourceCommand = &cobra.Command{
		Use:   "source",
		Short: "List OPA project source folders",
		Run: func(cmd *cobra.Command, args []string) {
			projPath := "."

			if !noUpdate {
				if err := doUpdate(projPath); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			}

			if err := doListSource(projPath, includeTestDirs, includeDepTests); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	listSourceCommand.Flags().BoolVarP(&includeTestDirs, "include-test-dirs", "t", false, "Include test directories in the list")
	listSourceCommand.Flags().BoolVar(&includeDepTests, "include-dep-tests", false, "Include dependency tests")
	listCommand.AddCommand(listSourceCommand)
}

func doListSource(projPath string, includeTestDirs, includeDepTests bool) error {
	printer.Trace("--- List sources start ---")
	defer printer.Trace("--- List sources end ---")

	project, err := proj.ReadAndLoadProject(projPath, true)
	if err != nil {
		return err
	}

	dataLocations, err := project.DataLocations()
	if err != nil {
		return fmt.Errorf("error getting data locations: %s", err)
	}

	if includeTestDirs {
		testDataLocations, err := project.TestLocations(includeDepTests)
		if err != nil {
			return fmt.Errorf("error getting test data locations: %s", err)
		}

		dataLocations = append(dataLocations, testDataLocations...)
	}

	printer.Output(strings.Join(dataLocations, "\n"))

	return nil
}
