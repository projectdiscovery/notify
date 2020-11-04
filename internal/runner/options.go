package runner

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/projectdiscovery/gologger"
)

type Options struct {
	BIID                    string
	SlackWebHookUrl         string
	SlackUsername           string
	SlackChannel            string
	Slack                   bool
	DiscordWebHookUrl       string
	DiscordWebHookUsername  string
	DiscordWebHookAvatarUrl string
	Discord                 bool
	Verbose                 bool
	NoColor                 bool
	Silent                  bool
	Version                 bool
	Interval                int
	InterceptBIID           bool
	InterceptBIIDTimeout    int
	HTTPMessage             string
	DNSMessage              string
}

func ParseConfigFileOrOptions() *Options {
	options := &Options{}

	flag.StringVar(&options.BIID, "biid", "", "burp collaborator unique id")
	flag.StringVar(&options.SlackWebHookUrl, "slack-webhook-url", "", "Slack Webhook URL")
	flag.StringVar(&options.SlackUsername, "slack-username", "", "Slack Username")
	flag.StringVar(&options.SlackChannel, "slack-channel", "", "Slack Channel")
	flag.BoolVar(&options.Slack, "slack", false, "Enable Slack")
	flag.StringVar(&options.DiscordWebHookUrl, "discord-webhook-url", "", "Discord Webhook URL")
	flag.StringVar(&options.DiscordWebHookUsername, "discord-username", "", "Discord Username")
	flag.StringVar(&options.DiscordWebHookAvatarUrl, "discord-channel", "", "Discord Channel")
	flag.BoolVar(&options.Discord, "discord", false, "Enable Discord")
	flag.BoolVar(&options.Silent, "silent", false, "Don't print the banner")
	flag.BoolVar(&options.Version, "version", false, "Show version of notify")
	flag.BoolVar(&options.Verbose, "v", false, "Show Verbose output")
	flag.BoolVar(&options.NoColor, "no-color", false, "Don't Use colors in output")
	flag.IntVar(&options.Interval, "interval", 2, "Polling interval in seconds")
	flag.BoolVar(&options.InterceptBIID, "intercept-biid", false, "Automatic BIID intercept")
	flag.IntVar(&options.InterceptBIIDTimeout, "intercept-biid-timeout", 120, "Automatic BIID intercept Timeout")
	flag.StringVar(&options.HTTPMessage, "message-http", defaultHTTPMessage, "HTTP Message")
	flag.StringVar(&options.DNSMessage, "message-dns", defaultDNSMessage, "DNS Message")

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
	dummyConfig.SlackWebHookUrl = "https://a.b.c/slack"
	dummyConfig.SlackUsername = "test"
	dummyConfig.SlackChannel = "test"
	dummyConfig.Slack = true
	dummyConfig.DiscordWebHookUrl = "https://a.b.c/discord"
	dummyConfig.DiscordWebHookUsername = "test"
	dummyConfig.DiscordWebHookAvatarUrl = "test"
	dummyConfig.Discord = true
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
		tmpFile.WriteString("# " + sc.Text() + "\n")
	}

	origFile.Close()
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	os.Rename(tmpFileName, configFile)

	gologger.Infof("Configuration file saved to %s\n", configFile)
}

func (options *Options) MergeFromConfig(configFileName string, ignoreError bool) {
	configFile, err := UnmarshalRead(configFileName)
	if err != nil {
		if ignoreError {
			gologger.Warningf("Could not read configuration file %s: %s\n", configFileName, err)
			return
		}
		gologger.Fatalf("Could not read configuration file %s: %s\n", configFileName, err)
	}

	if configFile.BIID != "" && !options.InterceptBIID {
		options.BIID = configFile.BIID
	}
	if configFile.SlackWebHookUrl != "" {
		options.SlackWebHookUrl = configFile.SlackWebHookUrl
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
	if configFile.DiscordWebHookUrl != "" {
		options.DiscordWebHookUrl = configFile.DiscordWebHookUrl
	}
	if configFile.DiscordWebHookUsername != "" {
		options.DiscordWebHookUsername = configFile.DiscordWebHookUsername
	}
	if configFile.DiscordWebHookAvatarUrl != "" {
		options.DiscordWebHookAvatarUrl = configFile.DiscordWebHookAvatarUrl
	}
	if configFile.Discord {
		options.Discord = configFile.Discord
	}
	if configFile.HTTPMessage != "" {
		options.HTTPMessage = configFile.HTTPMessage
	}
	if configFile.DNSMessage != "" {
		options.DNSMessage = configFile.DNSMessage
	}
	if configFile.Interval > 0 {
		options.Interval = configFile.Interval
	}
}
