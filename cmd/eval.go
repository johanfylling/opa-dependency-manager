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

	var evalCommand = &cobra.Command{
		Use:   "eval [flags] -- [opa eval flags]",
		Short: "Evaluate a Rego query using OPA",
		Long: `Evaluate a Rego query using OPA

Convenience command for running 'opa eval' with project dependencies.

Example:
'odm eval -- -d policy.rego "data.main.allow"' is equivalent to running:
'opa eval -d ./opa/dependencies -d policy.rego "data.main.allow"'
`,
		Run: func(cmd *cobra.Command, args []string) {
			projPath := "."

			if !noUpdate {
				if err := doUpdate(projPath); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			}

			if err := doEval(projPath, args); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	addNoUpdateFlag(evalCommand, &noUpdate)
	RootCommand.AddCommand(evalCommand)
}

func doEval(projPath string, args []string) error {
	printer.Trace("--- Eval start ---")
	defer printer.Trace("--- Eval end ---")

	if len(args) == 0 {
		// We're still calling OPA, so it can print its usage message
		printer.Info("no OPA flags provided")
	}

	project, err := proj.ReadAndLoadProject(projPath, true)
	if err != nil {
		return err
	}

	dataLocations, err := project.DataLocations()
	if err != nil {
		return fmt.Errorf("error getting data locations: %s", err)
	}

	opa := utils.NewOpa(dataLocations)
	if output, err := opa.Eval(args...); err != nil {
		return fmt.Errorf("error running opa eval:\n %s", err)
	} else {
		printer.Output(output)
	}

	return nil
}
