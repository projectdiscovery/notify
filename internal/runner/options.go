package runner

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/projectdiscovery/gologger"
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
	Verbose                 bool
	NoColor                 bool
	Silent                  bool
	Version                 bool
	Interval                int
	HTTPMessage             string
	DNSMessage              string
	CLIMessage              string
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
	flag.StringVar(&options.CLIMessage, "message-cli", defaultCLIMessage, "CLI Message")

	flag.Parse()

	// Read the inputs and configure the logging
	options.configureOutput()

	// write default conf file template if it doesn't exist
	options.writeDefaultConfig()

	if options.Version {
		gologger.Infof("Current Version: %s\n", Version)
		os.Exit(0)
	}

	// If a config file is provided, merge the options
	defaultConfigPath, err := getDefaultConfigFile()
	if err != nil {
		gologger.Errorf("Program exiting: %s\n", err)
	}
	options.MergeFromConfig(defaultConfigPath, true)

	// Show the user the banner
	showBanner()

	return options
}

func (options *Options) configureOutput() {
	if options.Verbose {
		gologger.MaxLevel = gologger.Verbose
	}
	if options.NoColor {
		gologger.UseColors = false
	}
	if options.Silent {
		gologger.MaxLevel = gologger.Silent
	}
}

func (options *Options) writeDefaultConfig() {
	configFile, err := getDefaultConfigFile()
	if err != nil {
		gologger.Warningf("Could not get default configuration file: %s\n", err)
	}

	if fileExists(configFile) {
		return
	}

	// Skip config file creation if run as root to avoid permission issues
	if os.Getuid() == 0 {
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
	dummyConfig.CLIMessage = "{{data}}"

	err = dummyConfig.MarshalWrite(configFile)
	if err != nil {
		gologger.Warningf("Could not write configuration file to %s: %s\n", configFile, err)
		return
	}

	// turn all lines into comments
	origFile, err := os.Open(configFile)
	if err != nil {
		gologger.Warningf("Could not process temporary file: %s\n", err)
		return
	}
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		log.Println(err)
		gologger.Warningf("Could not process temporary file: %s\n", err)
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

	gologger.Infof("Configuration file saved to %s\n", configFile)
}

// MergeFromConfig with existing options
func (options *Options) MergeFromConfig(configFileName string, ignoreError bool) {
	configFile, err := UnmarshalRead(configFileName)
	if err != nil {
		if ignoreError {
			gologger.Warningf("Could not read configuration file %s: %s\n", configFileName, err)
			return
		}
		gologger.Fatalf("Could not read configuration file %s: %s\n", configFileName, err)
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
	if configFile.HTTPMessage != "" {
		options.HTTPMessage = configFile.HTTPMessage
	}
	if configFile.DNSMessage != "" {
		options.DNSMessage = configFile.DNSMessage
	}
	if configFile.CLIMessage != "" {
		options.CLIMessage = configFile.CLIMessage
	}
	if configFile.Interval > 0 {
		options.Interval = configFile.Interval
	}
}
