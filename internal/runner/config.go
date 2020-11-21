package runner

import (
	"os"
	"path"

	"github.com/Shopify/yaml"
	"github.com/projectdiscovery/notify"
)

// ConfigDefaultFilename containing configuration
const ConfigDefaultFilename = "notify.conf"

// ConfigFile structure
//nolint:maligned // used once
type ConfigFile struct {
	BIID string `yaml:"burp_biid,omitempty"`
	// Slack
	SlackWebHookURL string `yaml:"slack_webhook_url,omitempty"`
	SlackUsername   string `yaml:"slack_username,omitempty"`
	SlackChannel    string `yaml:"slack_channel,omitempty"`
	Slack           bool   `yaml:"slack,omitempty"`

	// Discord
	DiscordWebHookURL       string `yaml:"discord_webhook_url,omitempty"`
	DiscordWebHookUsername  string `yaml:"discord_username,omitempty"`
	DiscordWebHookAvatarURL string `yaml:"discord_avatar,omitempty"`
	Discord                 bool   `yaml:"discord,omitempty"`

	// Telegram
	TelegramAPIKey string `yaml:"telegram_apikey,omitempty"`
	TelegramChatID string `yaml:"telegram_chat_id,omitempty"`
	Telegram       bool   `yaml:"telegram,omitempty"`

	// SMTP
	SMTPProviders []notify.SMTPProvider `yaml:"smtp_providers,omitempty"`
	SMTPCC        []string              `yaml:"smtp_cc,omitempty"`
	SMTP          bool                  `yaml:"smtp,omitempty"`

	Interval    int    `yaml:"interval,omitempty"`
	HTTPMessage string `yaml:"http_message,omitempty"`
	DNSMessage  string `yaml:"dns_message,omitempty"`
	CLIMessage  string `yaml:"cli_message,omitempty"`
	SMTPMessage string `yaml:"smtp_message,omitempty"`
}

// GetConfigDirectory from the system
func GetConfigDirectory() (string, error) {
	var config string

	directory, err := os.UserHomeDir()
	if err != nil {
		return config, err
	}
	config = directory + "/.config/notify"

	// Create All directory for notify even if they exist
	err = os.MkdirAll(config, os.ModePerm)
	if err != nil {
		return config, err
	}

	return config, nil
}

// CheckConfigExists in the specified path
func CheckConfigExists(configPath string) bool {
	if _, err := os.Stat(configPath); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	return false
}

// MarshalWrite to location
func (c *ConfigFile) MarshalWrite(file string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	// Indent the spaces too
	enc := yaml.NewEncoder(f)
	err = enc.Encode(&c)

	//nolint:errcheck // silent fail
	f.Close()
	return err
}

// UnmarshalRead the config file from location
func UnmarshalRead(file string) (ConfigFile, error) {
	config := ConfigFile{}

	f, err := os.Open(file)
	if err != nil {
		return config, err
	}
	err = yaml.NewDecoder(f).Decode(&config)
	//nolint:errcheck // silent fail
	f.Close()
	return config, err
}

func getDefaultConfigFile() (string, error) {
	directory, err := GetConfigDirectory()
	if err != nil {
		return "", err
	}
	return path.Join(directory, ConfigDefaultFilename), nil
}
