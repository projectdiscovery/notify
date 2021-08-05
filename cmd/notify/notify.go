package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/internal/runner"
	"github.com/projectdiscovery/notify/pkg/types"
)

var (
	cfgFile string
	options = &types.Options{}
)

func main() {
	readConfig()

	runner.ParseOptions(options)

	notifyRunner, err := runner.NewRunner(options)
	if err != nil {
		gologger.Fatal().Msgf("Could not create runner: %s\n", err)
	}

	// Setup close handler
	go func() {
		c := make(chan os.Signal, 1)
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

func readConfig() {
	set := goflags.NewFlagSet()
	set.Marshal = true
	set.SetDescription(`Notify is a general notification tool`)
	set.StringVar(&cfgFile, "config", "", "Notify configuration file")
	set.StringVar(&options.BIID, "biid", "", "burp collaborator unique id")
	set.BoolVar(&options.Silent, "silent", false, "Don't print the banner")
	set.BoolVar(&options.Version, "version", false, "Show version of notify")
	set.BoolVar(&options.Verbose, "v", false, "Show Verbose output")
	set.BoolVar(&options.NoColor, "no-color", false, "Don't Use colors in output")
	set.IntVar(&options.Interval, "interval", 2, "Polling interval in seconds")
	set.StringVar(&options.HTTPMessage, "message-http", types.DefaultHTTPMessage, "HTTP Message")
	set.StringVar(&options.DNSMessage, "message-dns", types.DefaultDNSMessage, "DNS Message")
	set.StringVar(&options.SMTPMessage, "message-smtp", types.DefaultSMTPMessage, "SMTP Message")
	set.StringVar(&options.CLIMessage, "message-cli", types.DefaultCLIMessage, "CLI Message")
	set.StringVar(&options.Data, "data", "", "File path to read data from")
	set.BoolVar(&options.Bulk, "bulk", false, "Read the input and send it in bulk, character limit can be set using char-limit flag")
	set.IntVar(&options.CharLimit, "char-limit", 4000, "Character limit for message")
	set.StringVar(&options.ProviderConfig, "provider-config", "", "provider config path (default: $HOME/.config/notify/provider-config.yaml)")
	set.StringSliceVar(&options.Providers, "provider", []string{}, "provider to send the notification to (optional)")
	set.StringSliceVar(&options.IDs, "id", []string{}, "id to send the notification to (optional)")
	set.StringVar(&options.Proxy, "proxy", "", "Set http proxy to be used by notify")

	_ = set.Parse()

	if cfgFile != "" {
		if err := set.MergeConfigFile(cfgFile); err != nil {
			gologger.Fatal().Msgf("Could not read config: %s\n", err)
		}
	}
}
