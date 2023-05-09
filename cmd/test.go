package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"styra.com/styrainc/odm/proj"
	"styra.com/styrainc/odm/utils"
)

func init() {
	var testCommand = &cobra.Command{
		Use:   "test [flags] -- [opa test flags]",
		Short: "Run OPA tests",
		Long:  `Run OPA tests`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := doTest(args); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	RootCommand.AddCommand(testCommand)
}

func doTest(args []string) error {
	opaArgs := make([]string, 0, len(args)+2)
	opaArgs = append(opaArgs, "test", ".opa/dependencies")

	project, err := proj.ReadProjectFromFile(".", true)
	if err == nil && project.SourceDir != "" {
		src, err := utils.NormalizeFilePath(project.SourceDir)
		if err != nil {
			return err
		}
		opaArgs = append(opaArgs, src)
	}

	opaArgs = append(opaArgs, args...)
	//opaArgs = append(opaArgs, "--ignore", ".opa/dependencies/*")

	if output, err := utils.RunCommand("opa", opaArgs...); err != nil {
		return fmt.Errorf("error running opa test:\n %s", err)
	} else {
		fmt.Println(output)
	}

	return nil
}
