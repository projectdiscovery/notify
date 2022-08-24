package testutils

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunNotifyAndGetResults returns a list of results for a template
func RunNotifyAndGetResults(debug bool, args ...string) ([]string, error) {
	cmd := exec.Command("bash", "-c")
	cmdLine := `echo "hello from notify integration test :)"` + ` | ./notify `
	cmdLine += strings.Join(args, " ")

	cmdLine += " --v"

	cmd.Args = append(cmd.Args, cmdLine)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	parts := []string{}
	items := strings.Split(string(data), "\n")
	for _, i := range items {
		if i != "" {
			if debug {
				fmt.Printf("%s\n", i)
			}
			if strings.Contains(i, "notification sent for id:") {
				parts = append(parts, i)
			}
		}
	}
	return parts, nil
}

// TestCase is a single integration test case
type TestCase interface {
	// Execute executes a test case and returns any errors if occurred
	Execute() error
}
