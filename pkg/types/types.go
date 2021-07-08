package types

type Options struct {
	BIID string `yaml:"burp_biid,omitempty"`

	// Slack
	SlackWebHookURL string `yaml:"slack_webhook_url,omitempty"`
	SlackUsername   string `yaml:"slack_username,omitempty"`
	SlackChannel    string `yaml:"slack_channel,omitempty"`
	SlackThreadTS   string `yaml:"slack_thread_ts,omitempty"`
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
	SMTPProviders []SMTPProvider `yaml:"smtp_providers,omitempty"`
	SMTPCC        []string       `yaml:"smtp_cc,omitempty"`
	SMTP          bool           `yaml:"smtp,omitempty"`

	// Pushover
	Pushover         bool     `yaml:"pushover,omitempty"`
	PushoverApiToken string   `yaml:"pushover_api_token,omitempty"`
	UserKey          string   `yaml:"pushover_user_key,omitempty"`
	PushoverDevices  []string `yaml:"pushover_devices,omitempty"`

	// Teams
	Teams           bool   `yaml:"teams,omitempty"`
	TeamsWebHookURL string `yaml:"teams_webhook_url,omitempty"`

	Verbose     bool
	NoColor     bool
	Silent      bool
	Version     bool
	Interval    int    `yaml:"interval,omitempty"`
	HTTPMessage string `yaml:"http_message,omitempty"`
	DNSMessage  string `yaml:"dns_message,omitempty"`
	CLIMessage  string `yaml:"cli_message,omitempty"`
	SMTPMessage string `yaml:"smtp_message,omitempty"`

	Stdin bool
	Data  string `yaml:"data,omitempty"`
}

type SMTPProvider struct {
	Server      string `yaml:"smtp_server,omitempty"`
	Username    string `yaml:"smtp_username,omitempty"`
	Password    string `yaml:"smtp_password,omitempty"`
	FromAddress string `yaml:"from_address,omitempty"`
}
