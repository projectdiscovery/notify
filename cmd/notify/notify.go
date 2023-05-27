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
	set.StringVar(&cfgFile, "config", "", "notify configuration file")
	set.StringVarP(&options.ProviderConfig, "provider-config", "pc", "", "provider config path (default: $HOME/.config/notify/provider-config.yaml)")
	set.StringVarP(&options.Data, "data", "i", "", "input file to send for notify")
	set.StringSliceVarP(&options.Providers, "provider", "p", []string{}, "provider to send the notification to (optional)", goflags.NormalizedStringSliceOptions)
	set.StringSliceVar(&options.IDs, "id", []string{}, "id to send the notification to (optional)", goflags.NormalizedStringSliceOptions)
	set.IntVarP(&options.RateLimit, "rate-limit", "rl", 1, "maximum number of HTTP requests to send per second")
	set.IntVarP(&options.Delay, "delay", "d", 0, "delay in seconds between each notification")
	set.BoolVar(&options.Bulk, "bulk", false, "enable bulk processing")
	set.IntVarP(&options.CharLimit, "char-limit", "cl", 4000, "max character limit per message")
	set.StringVarP(&options.MessageFormat, "msg-format", "mf", "", "add custom formatting to message")
	set.BoolVar(&options.Silent, "silent", false, "enable silent mode")
	set.BoolVarP(&options.Verbose, "verbose", "v", false, "enable verbose mode")
	set.BoolVar(&options.Version, "version", false, "display version")
	set.BoolVarP(&options.NoColor, "no-color", "nc", false, "disable colors in output")
	set.StringVar(&options.Proxy, "proxy", "", "HTTP Proxy to use with notify")
	set.CallbackVarP(runner.GetUpdateCallback(), "update", "up", "update notify to latest version")
	set.BoolVarP(&options.DisableUpdateCheck, "disable-update-check", "duc", false, "disable automatic notify update check")

	_ = set.Parse()

	if cfgFile != "" {
		if err := set.MergeConfigFile(cfgFile); err != nil {
			gologger.Fatal().Msgf("Could not read config: %s\n", err)
		}
	}
}
