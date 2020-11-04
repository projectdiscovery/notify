package notify

type Options struct {
	// Slack
	SlackWebHookUrl string
	SlackUsername   string
	SlackChannel    string
	Slack           bool

	// Discord
	DiscordWebHookUrl       string
	DiscordWebHookUsername  string
	DiscordWebHookAvatarUrl string
	Discord                 bool
}
