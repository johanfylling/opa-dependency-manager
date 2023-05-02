package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
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
	opaArgs := make([]string, 0, len(args)+1)
	opaArgs = append(opaArgs, "eval", "-d", ".opa/dependencies")
	opaArgs = append(opaArgs, args...)

	command := exec.Command("opa", opaArgs...)
	var outb, errb bytes.Buffer
	command.Stdout = &outb
	command.Stderr = &errb

	if err := command.Run(); err != nil {
		return fmt.Errorf("error running opa eval: %s", errb.String())
	} else {
		fmt.Println(outb.String())
	}

	return nil
}
