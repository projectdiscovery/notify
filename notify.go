package notify

import (
	"github.com/acarl005/stripansi"
	"github.com/projectdiscovery/retryablehttp-go"
)

// Notify handles the notification engine
type Notify struct {
	options        *Options
	client         *retryablehttp.Client
	slackClient    *SlackClient
	discordClient  *DiscordClient
	telegramClient *TelegramClient
}

// New notify instance
func New() (*Notify, error) {
	retryhttp := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)
	return &Notify{client: retryhttp}, nil
}

// NewWithOptions create a new instance of notify with options
func NewWithOptions(options *Options) (*Notify, error) {
	notifier, err := New()
	if err != nil {
		return nil, err
	}
	SlackClient := &SlackClient{
		client:     notifier.client,
		WebHookURL: options.SlackWebHookURL,
		UserName:   options.SlackUsername,
		Channel:    options.SlackUsername,
		TimeOut:    DefaultSlackTimeout,
	}
	discordClient := &DiscordClient{
		client:     notifier.client,
		WebHookURL: options.DiscordWebHookURL,
		UserName:   options.DiscordWebHookUsername,
		Avatar:     options.DiscordWebHookAvatarURL,
	}
	telegramClient := &TelegramClient{
		client: notifier.client,
		apiKEY: options.TelegramAPIKey,
		chatID: options.TelegramChatID,
	}
	return &Notify{options: options, slackClient: SlackClient, discordClient: discordClient, telegramClient: telegramClient}, nil
}

// SendNotification to registered webhooks
func (n *Notify) SendNotification(message string) error {
	// strip unsupported color control chars
	message = stripansi.Strip(message)
	if n.options.Slack {
		err := n.slackClient.SendInfo(message)
		if err != nil {
			return err
		}
	}

	if n.options.Discord {
		err := n.discordClient.SendInfo(message)
		if err != nil {
			return err
		}
	}

	if n.options.Telegram {
		err := n.telegramClient.SendInfo(message)
		if err != nil {
			return err
		}
	}

	return nil
}
