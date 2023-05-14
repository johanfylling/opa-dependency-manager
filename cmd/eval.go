package cmd

import (
	"fmt"
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
	opaArgs := make([]string, 0, len(args)+3)
	opaArgs = append(opaArgs, "eval", "-d", ".opa/dependencies")
	opaArgs = append(opaArgs, args...)

	project, err := proj.ReadProjectFromFile(".", true)
	if err == nil && project.SourceDir != "" {
		src, err := utils.NormalizeFilePath(project.SourceDir)
		if err != nil {
			return err
		}
		opaArgs = append(opaArgs, "-d", src)
	}

	if output, err := utils.RunCommand("opa", opaArgs...); err != nil {
		return fmt.Errorf("error running opa eval:\n %s", err)
	} else {
		fmt.Println(output)
	}

	return nil
}
