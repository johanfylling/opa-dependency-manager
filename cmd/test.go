package cmd

import (
	"fmt"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/proj"
	"github.com/johanfylling/odm/utils"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	var noUpdate bool
	var includeDeps bool

	var testCommand = &cobra.Command{
		Use:   "test [flags] -- [opa test flags]",
		Short: "Run OPA tests",
		Long:  `Run OPA tests`,
		Run: func(cmd *cobra.Command, args []string) {
			projPath := "."

			if !noUpdate {
				if err := doUpdate(projPath); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			}

			if err := doTest(projPath, includeDeps, args); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	testCommand.Flags().BoolVar(&includeDeps, "include-deps", false, "Include dependency tests")
	addNoUpdateFlag(testCommand, &noUpdate)
	RootCommand.AddCommand(testCommand)
}

func doTest(projPath string, includeDependencies bool, args []string) error {
	printer.Trace("--- Test start ---")
	defer printer.Trace("--- Test end ---")

	project, err := proj.ReadAndLoadProject(projPath, true)
	if err != nil {
		return err
	}

	dataLocations, err := project.DataLocations()
	if err != nil {
		return fmt.Errorf("error getting data locations: %s", err)
	}

	testLocations, err := project.TestLocations(includeDependencies)
	if err != nil {
		return fmt.Errorf("error getting test locations: %s", err)
	}

	dataLocations = append(dataLocations, testLocations...)

	opa := utils.NewOpa(dataLocations...)
	if output, err := opa.Test(args...); err != nil {
		return fmt.Errorf("error running opa test:\n %s", err)
	} else {
		printer.Output(output)
	}

	return nil
}
