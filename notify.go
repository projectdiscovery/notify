package notify

import (
	"github.com/acarl005/stripansi"
	"github.com/projectdiscovery/retryablehttp-go"
)

type Notify struct {
	options       *Options
	client        *retryablehttp.Client
	slackClient   *SlackClient
	discordClient *DiscordClient
}

func New() (*Notify, error) {
	retryablehttp := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)
	return &Notify{client: retryablehttp}, nil
}

func NewWithOptions(options *Options) (*Notify, error) {
	notifier, err := New()
	if err != nil {
		return nil, err
	}
	SlackClient := &SlackClient{
		client:     notifier.client,
		WebHookUrl: options.SlackWebHookUrl,
		UserName:   options.SlackUsername,
		Channel:    options.SlackUsername,
		TimeOut:    DefaultSlackTimeout,
	}
	discordClient := &DiscordClient{
		client:     notifier.client,
		WebHookUrl: options.DiscordWebHookUrl,
		UserName:   options.DiscordWebHookUsername,
		Avatar:     options.DiscordWebHookAvatarUrl,
	}
	return &Notify{options: options, slackClient: SlackClient, discordClient: discordClient}, nil
}

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

	return nil
}
