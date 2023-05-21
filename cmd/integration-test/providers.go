package main

import (
	"fmt"
	"strings"

	"github.com/projectdiscovery/notify/internal/testutils"
)

func run(provider string) error {
	args := []string{"--provider", provider}
	if *providerConfig != "" {
		args = append(args, "--provider-config", *providerConfig)
	}
	results, err := testutils.RunNotifyAndGetResults(debug, args...)
	if err != nil {
		return err
	}
	if len(results) < 1 {
		return errIncorrectResultsCount(results)
	}
	for _, r := range results {
		if !strings.Contains(strings.ToLower(r), strings.ToLower(provider)) {
			return fmt.Errorf("incorrect result %s", results[0])
		}
	}
	return nil
}

type discord struct{}

func (h *discord) Execute() error {
	return run("discord")
}

type custom struct{}

func (h *custom) Execute() error {
	return run("custom")
}

type slack struct{}

func (h *slack) Execute() error {
	return run("slack")
}

// type pushover struct{}
//
// func (h *pushover) Execute() error {
// 	return run("pushover")
// }
//
// type smtp struct{}
//
// func (h *smtp) Execute() error {
// 	return run("smtp")
// }
//
// type teams struct{}
//
// func (h *teams) Execute() error {
// 	return run("teams")
// }
//
// type telegram struct{}
//
// func (h *telegram) Execute() error {
// 	return run("telegram")
// }

func errIncorrectResultsCount(results []string) error {
	return fmt.Errorf("incorrect number of results %s", strings.Join(results, "\n\t"))
}
