package notify

// Options of internal webhooks
//nolint:maligned // used once
type Options struct {
	// Slack
	SlackWebHookURL string
	SlackUsername   string
	SlackChannel    string
	Slack           bool

	// Discord
	DiscordWebHookURL       string
	DiscordWebHookUsername  string
	DiscordWebHookAvatarURL string
	Discord                 bool

	// Telegram
	TelegramAPIKey string
	TelegramChatID string
	Telegram       bool
}
