package runner

import (
	"os"
	"path"

	"github.com/Shopify/yaml"
)

const ConfigDefaultFilename = "notify.conf"

type ConfigFile struct {
	BIID string `yaml:"burp_biid,omitempty"`
	// Slack
	SlackWebHookUrl string `yaml:"slack_webhook_url,omitempty"`
	SlackUsername   string `yaml:"slack_username,omitempty"`
	SlackChannel    string `yaml:"slack_channel,omitempty"`
	Slack           bool   `yaml:"slack,omitempty"`

	// Discord
	DiscordWebHookUrl       string `yaml:"discord_webhook_url,omitempty"`
	DiscordWebHookUsername  string `yaml:"discord_username,omitempty"`
	DiscordWebHookAvatarUrl string `yaml:"discord_avatar,omitempty"`
	Discord                 bool   `yaml:"discord,omitempty"`
	Interval                int    `yaml:"interval,omitempty"`
	HTTPMessage             string `yaml:"http_message,omitempty"`
	DNSMessage              string `yaml:"dns_message,omitempty"`
}

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

func CheckConfigExists(configPath string) bool {
	if _, err := os.Stat(configPath); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	return false
}

func (c *ConfigFile) MarshalWrite(file string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	// Indent the spaces too
	enc := yaml.NewEncoder(f)
	err = enc.Encode(&c)
	f.Close()
	return err
}

func UnmarshalRead(file string) (ConfigFile, error) {
	config := ConfigFile{}

	f, err := os.Open(file)
	if err != nil {
		return config, err
	}
	err = yaml.NewDecoder(f).Decode(&config)
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
