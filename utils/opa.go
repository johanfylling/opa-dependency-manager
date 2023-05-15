package utils

import (
	"fmt"
	"os"
)

type Opa struct {
	location      string
	dataLocations []string
}

func NewOpa(dataLocations []string) *Opa {
	location, ok := os.LookupEnv("OPA_PATH")
	if !ok {
		location = "opa"
	}

	return &Opa{
		location:      location,
		dataLocations: dataLocations,
	}
}

func (o *Opa) Eval(passThroughArgs ...string) (string, error) {
	opaArgs := make([]string, 0, 1+2*len(o.dataLocations)+len(passThroughArgs))
	opaArgs = append(opaArgs, "eval")

	for _, location := range o.dataLocations {
		opaArgs = append(opaArgs, "-d", location)
	}
	opaArgs = append(opaArgs, passThroughArgs...)

	fmt.Printf("Executing: %s %s\n", o.location, opaArgs)
	return RunCommand(o.location, opaArgs...)
}

func (o *Opa) Test(passThroughArgs ...string) (string, error) {
	opaArgs := make([]string, 0, 1+len(o.dataLocations)+len(passThroughArgs))
	opaArgs = append(opaArgs, "test")

	for _, location := range o.dataLocations {
		opaArgs = append(opaArgs, location)
	}
	opaArgs = append(opaArgs, passThroughArgs...)

	fmt.Printf("Executing: %s %s\n", o.location, opaArgs)
	return RunCommand(o.location, opaArgs...)
}
