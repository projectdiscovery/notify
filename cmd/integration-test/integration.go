package main

import (
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/notify/internal/testutils"
)

var (
	debug     = os.Getenv("DEBUG") == "true"
	errored   = false
	success   = aurora.Green("[✓]").String()
	failed    = aurora.Red("[✘]").String()
	testCases = map[string]testutils.TestCase{
		"smtp":     &smtp{},
		"discord":  &discord{},
		"telegram": &telegram{},
		"teams":    &teams{},
		"slack":    &slack{},
		"pushover": &pushover{},
		"custom":   &custom{},
	}
)

func main() {
	for name, test := range testCases {
		fmt.Printf("Running test cases for \"%s\"\n", aurora.Blue(name))
		err := test.Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Test \"%s\" failed: %s\n", failed, name, err)
			errored = true
		} else {
			fmt.Printf("%s Test \"%s\" passed!\n", success, name)
		}
	}
	if errored {
		os.Exit(1)
	}
}
