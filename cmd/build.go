package cmd

import (
	"fmt"
	"github.com/johanfylling/odm/printer"
	"github.com/johanfylling/odm/proj"
	"github.com/johanfylling/odm/utils"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	defaultTargetDir  = "build"
	defaultTargetFile = "bundle.tar.gz"
)

func init() {
	var noUpdate bool

	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build OPA bundle",
		Run: func(cmd *cobra.Command, args []string) {
			projPath := "."

			if !noUpdate {
				if err := doUpdate(projPath); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			}

			if err := doBuild(projPath, args); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}

	addNoUpdateFlag(buildCmd, &noUpdate)
	RootCommand.AddCommand(buildCmd)
}

func doBuild(projPath string, args []string) error {
	printer.Trace("--- Eval start ---")
	defer printer.Trace("--- Eval end ---")

	project, err := proj.ReadAndLoadProject(projPath, true)
	if err != nil {
		return err
	}

	outputDir, outputFile := filepath.Split(project.Build.Output)
	if outputFile == "" {
		outputFile = defaultTargetFile
		if outputDir == "" {
			outputDir = defaultTargetDir
		}
	}

	if outputDir != "" {
		outputDir = filepath.Join(project.Dir(), outputDir)
		if err := utils.MakeDir(outputDir); err != nil {
			return fmt.Errorf("error creating build directory: %s", err)
		}
	} else {
		outputDir = project.Dir()
	}

	outputPath := filepath.Join(filepath.Clean(outputDir), outputFile)

	dataLocations, err := project.DataLocations()
	if err != nil {
		return fmt.Errorf("error getting data locations: %s", err)
	}

	opa := utils.NewOpa(dataLocations...).
		WithEntrypoints(project.Build.Entrypoints).
		WithTarget(project.Build.Target)
	if output, err := opa.Build(outputPath, args...); err != nil {
		return fmt.Errorf("error running opa eval:\n %s", err)
	} else {
		printer.Info(output)
	}

	return nil
}
