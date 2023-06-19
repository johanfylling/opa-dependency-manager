package utils

import (
	"fmt"
	"github.com/johanfylling/odm/printer"
	"os"
)

type Opa struct {
	location      string
	dataLocations []string
	entrypoints   []string
	target        string
}

func NewOpa(dataLocations ...string) *Opa {
	location, ok := os.LookupEnv("OPA_PATH")
	if !ok {
		location = "opa"
	}

	printer.Debug("Creating OPA instance\nlocation: %s\ndata: %v", location, dataLocations)

	return &Opa{
		location:      location,
		dataLocations: dataLocations,
	}
}

func (o *Opa) WithEntrypoints(entrypoints []string) *Opa {
	cpy := *o
	cpy.entrypoints = entrypoints
	return &cpy
}

func (o *Opa) WithTarget(target string) *Opa {
	cpy := *o
	cpy.target = target
	return &cpy
}

func (o *Opa) Eval(passThroughArgs ...string) (string, error) {
	printer.Info("Running OPA eval")

	opaArgs := make([]string, 0, 1+2*len(o.dataLocations)+len(passThroughArgs))
	opaArgs = append(opaArgs, "eval")

	for _, location := range o.dataLocations {
		opaArgs = append(opaArgs, "-d", location)
	}
	opaArgs = append(opaArgs, passThroughArgs...)

	return RunCommand(o.location, opaArgs...)
}

func (o *Opa) Test(passThroughArgs ...string) (string, error) {
	printer.Info("Running OPA test")

	opaArgs := make([]string, 0, 1+len(o.dataLocations)+len(passThroughArgs))
	opaArgs = append(opaArgs, "test")

	for _, location := range o.dataLocations {
		opaArgs = append(opaArgs, location)
	}
	opaArgs = append(opaArgs, passThroughArgs...)

	return RunCommand(o.location, opaArgs...)
}

func (o *Opa) Build(outputPath string, passThroughFlags ...string) (string, error) {
	printer.Info("Running OPA build")
	printer.Debug("Output bundle path: %s", outputPath)

	opaArgs := prefixEntrypoints(o.entrypoints, passThroughFlags)
	opaArgs = prefixOutput(outputPath, opaArgs)
	opaArgs = prefixTarget(o.target, opaArgs)
	// locations must be first in the list of arguments, so prefixed last
	opaArgs = prefixDataLocations(o.dataLocations, opaArgs, false)

	return runOpaCommand(o.location, "build", opaArgs...)
}

func (o *Opa) Refactor(fromPackage, toPackage string) error {
	printer.Info("Running OPA refactor")
	printer.Debug("From package: %s", fromPackage)
	printer.Debug("To package: %s", toPackage)

	mapping := fmt.Sprintf("%s:%s", fromPackage, toPackage)

	opaArgs := make([]string, 0, 4+len(o.dataLocations))
	opaArgs = append(opaArgs, "move")
	opaArgs = append(opaArgs, o.dataLocations...)
	opaArgs = append(opaArgs, "-w", "-p", mapping)

	_, err := runOpaCommand(o.location, "refactor", opaArgs...)
	return err
}

func runOpaCommand(opaLocation string, command string, flags ...string) (string, error) {
	opaArgs := make([]string, 0, 1+len(flags))
	opaArgs = append(opaArgs, command)
	opaArgs = append(opaArgs, flags...)

	return RunCommand(opaLocation, opaArgs...)
}

func prefixDataLocations(dataLocations []string, flags []string, namedFlag bool) []string {
	multiplier := 1
	if namedFlag {
		multiplier = 2
	}

	newFlags := make([]string, 0, len(dataLocations)*multiplier+len(flags))
	for _, location := range dataLocations {
		if namedFlag {
			newFlags = append(newFlags, "-d", location)
		} else {
			newFlags = append(newFlags, location)
		}
	}

	return append(newFlags, flags...)
}

func prefixEntrypoints(entrypoints []string, flags []string) []string {
	newFlags := make([]string, 0, len(entrypoints)*2+len(flags))
	for _, entrypoint := range entrypoints {
		newFlags = append(newFlags, "-e", entrypoint)
	}

	return append(newFlags, flags...)
}

func prefixOutput(outputPath string, flags []string) []string {
	newFlags := make([]string, 0, 2+len(flags))
	if !Contains(flags, "-o") && !Contains(flags, "--output") {
		newFlags = append(newFlags, "-o", outputPath)
	} else if outputPath != "" {
		printer.Debug("Output path present on pass-through flags to OPA, ignoring configured output path")
	}

	return append(newFlags, flags...)
}

func prefixTarget(target string, flags []string) []string {
	if target == "" {
		return flags
	}
	newFlags := make([]string, 0, 2+len(flags))
	if !Contains(flags, "-t") && !Contains(flags, "--target") {
		newFlags = append(newFlags, "-t", target)
	} else if target != "" {
		printer.Debug("Target present on pass-through flags to OPA, ignoring configured target")
	}

	return append(newFlags, flags...)
}
