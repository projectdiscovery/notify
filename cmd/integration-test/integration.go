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
	testCases      = map[string]providerTest{
		"discord": {test: &discord{}, requiredEnv: "DISCORD_WEBHOOK_URL"},
		"slack":   {test: &slack{}, requiredEnv: "SLACK_WEBHOOK_URL"},
		"custom":  {test: &custom{}, requiredEnv: "CUSTOM_WEBHOOK_URL"},
		//		"telegram": {test: &telegram{}, requiredEnv: "..."},
		//		"teams":    {test: &teams{}, requiredEnv: "..."},
		//		"smtp":     {test: &smtp{}, requiredEnv: "..."},
		//		"pushover": {test: &pushover{}, requiredEnv: "..."},
		"gotify": {test: &gotify{}},
	}
)

// providerTest pairs a test case with the env var that must be set for it to
// run. Tests whose required secret is missing are skipped, which lets fork PRs
// (that don't receive repository secrets) stay green while full coverage runs
// on internal branches where the secrets are available.
type providerTest struct {
	test        testutils.TestCase
	requiredEnv string
}

func main() {
	flag.Parse()

	for name, tc := range testCases {
		// run only gotify test for dependabot
		if isDependabot && name != "gotify" {
			continue
		}
		// skip provider tests whose required webhook secret is unavailable
		// (e.g. PRs from forks, which GitHub does not expose secrets to).
		if tc.requiredEnv != "" && os.Getenv(tc.requiredEnv) == "" {
			fmt.Printf("Skipping test cases for \"%s\": %s not set\n", aurora.Blue(name), tc.requiredEnv)
			continue
		}
		fmt.Printf("Running test cases for \"%s\"\n", aurora.Blue(name))
		err := tc.test.Execute()
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
