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
			if err := doEval(args); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	RootCommand.AddCommand(evalCommand)
}

func doEval(args []string) error {
	printer.Trace("--- Eval start ---")
	defer printer.Trace("--- Eval end ---")

	project, err := proj.ReadProjectFromFile(".", true)
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
