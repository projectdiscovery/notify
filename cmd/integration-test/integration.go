package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"

	"github.com/projectdiscovery/notify/internal/testutils"
)

var (
	providerConfig = flag.String("provider-config", "", "provider config to use for testing")
	debug          = os.Getenv("DEBUG") == "true"
	isDependabot   = os.Getenv("DEPENDABOT") == "true"
	errored        = false
	success        = aurora.Green("[✓]").String()
	failed         = aurora.Red("[✘]").String()
	testCases      = map[string]testutils.TestCase{
		"discord": &discord{},
		"slack":   &slack{},
		"custom":  &custom{},
		//		"telegram": &telegram{},
		//		"teams":    &teams{},
		//		"smtp":     &smtp{},
		//		"pushover": &pushover{},
		"gotify": &gotify{},
	}
	// Fork pull requests do not receive repository secrets. Skip those
	// providers instead of failing the job. Gotify uses the local server.
	providerEnv = map[string]string{
		"discord": "DISCORD_WEBHOOK_URL",
		"slack":   "SLACK_WEBHOOK_URL",
		"custom":  "CUSTOM_WEBHOOK_URL",
	}
)

func main() {
	flag.Parse()

	for name, test := range testCases {
		// run only gotify test for dependabot
		if isDependabot && name != "gotify" {
			continue
		}
		if key, ok := providerEnv[name]; ok && os.Getenv(key) == "" {
			fmt.Printf("Skipping %s because %s is unset\n", name, key)
			continue
		}
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
