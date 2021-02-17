package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/internal/runner"
)

func main() {
	options := runner.ParseConfigFileOrOptions()

	notifyRunner, err := runner.NewRunner(options)
	if err != nil {
		gologger.Fatal().Msgf("Could not create runner: %s\n", err)
	}

	// Setup close handler
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println("\r- Ctrl+C pressed in Terminal")
			notifyRunner.Close()
			os.Exit(0)
		}()
	}()

	err = notifyRunner.Run()
	if err != nil {
		gologger.Fatal().Msgf("Could not run notifier: %s\n", err)
	}
}
