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
	set := goflags.New()
	set.Marshal = true
	set.SetDescription(`Notify is a general notification tool`)
	set.StringVar(&cfgFile, "config", "", "Notify configuration file")
	set.StringVar(&options.BIID, "biid", "", "burp collaborator unique id")
	set.StringVar(&options.SlackWebHookURL, "slack-webhook-url", "", "Slack Webhook URL")
	set.StringVar(&options.SlackUsername, "slack-username", "", "Slack Username")
	set.StringVar(&options.SlackChannel, "slack-channel", "", "Slack Channel")
	set.BoolVar(&options.Slack, "slack", false, "Enable Slack")
	set.StringVar(&options.DiscordWebHookURL, "discord-webhook-url", "", "Discord Webhook URL")
	set.StringVar(&options.DiscordWebHookUsername, "discord-username", "", "Discord Username")
	set.StringVar(&options.DiscordWebHookAvatarURL, "discord-channel", "", "Discord Channel")
	set.BoolVar(&options.Discord, "discord", false, "Enable Discord")
	set.StringVar(&options.TelegramAPIKey, "telegram-api-key", "", "Telegram API Key")
	set.StringVar(&options.TelegramChatID, "telegram-chat-id", "", "Telegram Chat ID")
	set.BoolVar(&options.Telegram, "telegram", false, "Enable Telegram")
	set.BoolVar(&options.Silent, "silent", false, "Don't print the banner")
	set.BoolVar(&options.Version, "version", false, "Show version of notify")
	set.BoolVar(&options.Verbose, "v", false, "Show Verbose output")
	set.BoolVar(&options.NoColor, "no-color", false, "Don't Use colors in output")
	set.IntVar(&options.Interval, "interval", 2, "Polling interval in seconds")
	set.StringVar(&options.HTTPMessage, "message-http", types.DefaultHTTPMessage, "HTTP Message")
	set.StringVar(&options.DNSMessage, "message-dns", types.DefaultDNSMessage, "DNS Message")
	set.StringVar(&options.SMTPMessage, "message-smtp", types.DefaultSMTPMessage, "SMTP Message")
	set.StringVar(&options.CLIMessage, "message-cli", types.DefaultCLIMessage, "CLI Message")
	set.StringVar(&options.Data, "data", "", "file path to read data from")

	_ = set.Parse()

	if cfgFile != "" {
		if err := set.MergeConfigFile(cfgFile); err != nil {
			gologger.Fatal().Msgf("Could not read config: %s\n", err)
		}
	}
}
