package runner

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/notify"
)

// Options of the internal runner
//nolint:maligned // used once
type Options struct {
	BIID                    string
	SlackWebHookURL         string
	SlackUsername           string
	SlackChannel            string
	Slack                   bool
	DiscordWebHookURL       string
	DiscordWebHookUsername  string
	DiscordWebHookAvatarURL string
	Discord                 bool
	TelegramAPIKey          string
	TelegramChatID          string
	Telegram                bool
	SMTP                    bool
	SMTPProviders           []notify.SMTPProvider
	SMTPCC                  []string
	Verbose                 bool
	NoColor                 bool
	Silent                  bool
	Version                 bool
	Interval                int
	HTTPMessage             string
	DNSMessage              string
	CLIMessage              string
	SMTPMessage             string
}

// ParseConfigFileOrOptions combining all settings
func ParseConfigFileOrOptions() *Options {
	options := &Options{}

	flag.StringVar(&options.BIID, "biid", "", "burp collaborator unique id")
	flag.StringVar(&options.SlackWebHookURL, "slack-webhook-url", "", "Slack Webhook URL")
	flag.StringVar(&options.SlackUsername, "slack-username", "", "Slack Username")
	flag.StringVar(&options.SlackChannel, "slack-channel", "", "Slack Channel")
	flag.BoolVar(&options.Slack, "slack", false, "Enable Slack")
	flag.StringVar(&options.DiscordWebHookURL, "discord-webhook-url", "", "Discord Webhook URL")
	flag.StringVar(&options.DiscordWebHookUsername, "discord-username", "", "Discord Username")
	flag.StringVar(&options.DiscordWebHookAvatarURL, "discord-channel", "", "Discord Channel")
	flag.BoolVar(&options.Discord, "discord", false, "Enable Discord")
	flag.StringVar(&options.TelegramAPIKey, "telegram-api-key", "", "Telegram API Key")
	flag.StringVar(&options.TelegramChatID, "telegram-chat-id", "", "Telegram Chat ID")
	flag.BoolVar(&options.Telegram, "telegram", false, "Enable Telegram")
	flag.BoolVar(&options.Silent, "silent", false, "Don't print the banner")
	flag.BoolVar(&options.Version, "version", false, "Show version of notify")
	flag.BoolVar(&options.Verbose, "v", false, "Show Verbose output")
	flag.BoolVar(&options.NoColor, "no-color", false, "Don't Use colors in output")
	flag.IntVar(&options.Interval, "interval", 2, "Polling interval in seconds")
	flag.StringVar(&options.HTTPMessage, "message-http", defaultHTTPMessage, "HTTP Message")
	flag.StringVar(&options.DNSMessage, "message-dns", defaultDNSMessage, "DNS Message")
	flag.StringVar(&options.SMTPMessage, "message-smtp", defaultSMTPMessage, "SMTP Message")
	flag.StringVar(&options.CLIMessage, "message-cli", defaultCLIMessage, "CLI Message")

	flag.Parse()

	// Read the inputs and configure the logging
	options.configureOutput()

	// write default conf file template if it doesn't exist
	options.writeDefaultConfig()

	if options.Version {
		gologger.Info().Msgf("Current Version: %s\n", Version)
		os.Exit(0)
	}

	// If a config file is provided, merge the options
	defaultConfigPath, err := getDefaultConfigFile()
	if err != nil {
		gologger.Error().Msgf("Program exiting: %s\n", err)
	}
	options.MergeFromConfig(defaultConfigPath, true)

	// Show the user the banner
	showBanner()

	return options
}

func (options *Options) configureOutput() {
	if options.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	}
	// Not used
	// if options.NoColor {
	// 	gologger.UseColors = false
	// }
	if options.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	}
}

func (options *Options) writeDefaultConfig() {
	configFile, err := getDefaultConfigFile()
	if err != nil {
		gologger.Print().Msgf("Could not get default configuration file: %s\n", err)
	}

	if fileExists(configFile) {
		gologger.Print().Msgf("Found existing config file: %s\n", configFile)
		return
	}

	// Skip config file creation if run as root to avoid permission issues
	if os.Getuid() == 0 {
		gologger.Print().Msgf("Running as root, skipping config file write to avoid permissions issues: %s\n", configFile)
		return
	}

	var dummyConfig ConfigFile
	dummyConfig.BIID = "123456798"
	dummyConfig.SlackWebHookURL = "https://a.b.c/slack"
	//nolint:goconst // test data
	dummyConfig.SlackUsername = "test"
	//nolint:goconst // test data
	dummyConfig.SlackChannel = "test"
	dummyConfig.Slack = true
	dummyConfig.DiscordWebHookURL = "https://a.b.c/discord"
	//nolint:goconst // test data
	dummyConfig.DiscordWebHookUsername = "test"
	//nolint:goconst // test data
	dummyConfig.DiscordWebHookAvatarURL = "test"
	dummyConfig.Discord = true
	dummyConfig.TelegramAPIKey = "123456879"
	dummyConfig.TelegramChatID = "123"
	dummyConfig.Telegram = true
	dummyConfig.SMTPProviders = append(dummyConfig.SMTPProviders, notify.SMTPProvider{
		AuthenticationType: "basic",
		Server:             "smtp.server.something:25",
		Username:           "myusername@oremail.address",
		Password:           "mysecretpassword",
	})
	dummyConfig.SMTPCC = append(dummyConfig.SMTPCC, "receiver@email.address")
	dummyConfig.SMTP = true
	dummyConfig.Interval = 2
	dummyConfig.HTTPMessage = "The collaborator server received an {{protocol}} request from {{from}} at {{time}}:\n" +
		"```\n" +
		"{{request}}\n" +
		"{{response}}\n" +
		"```"
	dummyConfig.DNSMessage = "The collaborator server received a DNS lookup of type {{type}} for the domain name {{domain}} from {{from}} at {{time}}:\n" +
		"```\n" +
		"{{request}}\n" +
		"```"
	dummyConfig.SMTPMessage = "The collaborator server received an SMTP connection from IP address {{from}} at {{time}}.\n" +
		"The email details were:\n\n" +
		"From:\n{{sender}}\n\n" +
		"To:\n{{recipients}}\n\n" +
		"Message:\n{{message}}\n\n" +
		"SMTP Conversation:\n{{conversation}}"
	dummyConfig.CLIMessage = "{{data}}"

	err = dummyConfig.MarshalWrite(configFile)
	if err != nil {
		gologger.Print().Msgf("Could not write configuration file to %s: %s\n", configFile, err)
		return
	}

	// turn all lines into comments
	origFile, err := os.Open(configFile)
	if err != nil {
		gologger.Print().Msgf("Could not process temporary file: %s\n", err)
		return
	}
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		log.Println(err)
		gologger.Print().Msgf("Could not process temporary file: %s\n", err)
		return
	}
	sc := bufio.NewScanner(origFile)
	for sc.Scan() {
		//nolint:errcheck // silent fail
		tmpFile.WriteString("# " + sc.Text() + "\n")
	}
	//nolint:errcheck // silent fail
	origFile.Close()
	tmpFileName := tmpFile.Name()
	//nolint:errcheck // silent fail
	tmpFile.Close()
	//nolint:errcheck // silent fail
	os.Rename(tmpFileName, configFile)

	gologger.Print().Msgf("Configuration file saved to %s\n", configFile)
}

// MergeFromConfig with existing options
func (options *Options) MergeFromConfig(configFileName string, ignoreError bool) {
	configFile, err := UnmarshalRead(configFileName)
	if err != nil {
		if ignoreError {
			gologger.Print().Msgf("Could not read configuration file %s - ignoring error: %s\n", configFileName, err)
			return
		}
		gologger.Print().Msgf("Could not read configuration file %s: %s\n", configFileName, err)
	}

	if configFile.BIID != "" {
		options.BIID = configFile.BIID
	}
	if configFile.SlackWebHookURL != "" {
		options.SlackWebHookURL = configFile.SlackWebHookURL
	}
	if configFile.SlackUsername != "" {
		options.SlackUsername = configFile.SlackUsername
	}
	if configFile.SlackChannel != "" {
		options.SlackChannel = configFile.SlackChannel
	}
	if configFile.Slack {
		options.Slack = configFile.Slack
	}
	if configFile.DiscordWebHookURL != "" {
		options.DiscordWebHookURL = configFile.DiscordWebHookURL
	}
	if configFile.DiscordWebHookUsername != "" {
		options.DiscordWebHookUsername = configFile.DiscordWebHookUsername
	}
	if configFile.DiscordWebHookAvatarURL != "" {
		options.DiscordWebHookAvatarURL = configFile.DiscordWebHookAvatarURL
	}
	if configFile.Discord {
		options.Discord = configFile.Discord
	}
	if configFile.TelegramAPIKey != "" {
		options.TelegramAPIKey = configFile.TelegramAPIKey
	}
	if configFile.TelegramChatID != "" {
		options.TelegramChatID = configFile.TelegramChatID
	}
	if configFile.Telegram {
		options.Telegram = configFile.Telegram
	}
	if len(configFile.SMTPProviders) > 0 {
		options.SMTPProviders = configFile.SMTPProviders
	}
	if len(configFile.SMTPCC) > 0 {
		options.SMTPCC = configFile.SMTPCC
	}
	if configFile.SMTP {
		options.SMTP = configFile.SMTP
	}
	if configFile.HTTPMessage != "" {
		options.HTTPMessage = configFile.HTTPMessage
	}
	if configFile.DNSMessage != "" {
		options.DNSMessage = configFile.DNSMessage
	}
	if configFile.SMTPMessage != "" {
		options.SMTPMessage = configFile.SMTPMessage
	}
	if configFile.CLIMessage != "" {
		options.CLIMessage = configFile.CLIMessage
	}
	if configFile.Interval > 0 {
		options.Interval = configFile.Interval
	}
}
