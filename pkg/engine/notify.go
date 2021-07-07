package engine

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/containrrr/shoutrrr"
	"github.com/projectdiscovery/notify/pkg/types"
)

// Notify handles the notification engine
type Notify struct {
	options *types.Options
}

// New notify instance
func New() (*Notify, error) {
	return &Notify{}, nil
}

// NewWithOptions create a new instance of notify with options
func NewWithOptions(options *types.Options) (*Notify, error) {
	return &Notify{options: options}, nil
}

// SendNotification to registered webhooks
func (n *Notify) SendNotification(message string) error {
	// strip unsupported color control chars
	message = stripansi.Strip(message)
	if n.options.Slack {
		slackTokens := strings.TrimPrefix(n.options.SlackWebHookURL, "https://hooks.slack.com/services/")
		url := &url.URL{
			Scheme:   "slack",
			Path:     slackTokens,
			RawQuery: fmt.Sprintf("thread_ts=%s", n.options.SlackThreadTS),
		}

		err := shoutrrr.Send(url.String(), message)
		if err != nil {
			return err
		}
	}

	if n.options.Discord {
		discordTokens := strings.TrimPrefix(n.options.DiscordWebHookURL, "https://discord.com/api/webhooks/")
		tokens := strings.Split(discordTokens, "/")
		if len(tokens) != 2 {
			return errors.New("Wrong discord configuration")
		}
		webhookID, token := tokens[0], tokens[1]
		url := fmt.Sprintf("discord://%s@%s", token, webhookID)
		err := shoutrrr.Send(url, message)
		if err != nil {
			return err
		}
	}

	if n.options.Telegram {
		url := fmt.Sprintf("telegram://%s@telegram?channels=%s", n.options.TelegramAPIKey, n.options.TelegramChatID)
		err := shoutrrr.Send(url, message)
		if err != nil {
			return err
		}
	}

	if n.options.SMTP {
		for _, provider := range n.options.SMTPProviders {
			url := fmt.Sprintf("smtp://%s:%s@%s/?fromAddress=%s&toAddresses=%s", provider.Username, provider.Password, provider.Server, provider.FromAddress, strings.Join(n.options.SMTPCC, ","))
			err := shoutrrr.Send(url, message)
			if err != nil {
				return err
			}
		}
	}

	if n.options.Pushover {
		url := fmt.Sprintf("pushover://shoutrrr:%s@%s/?devices=%s", n.options.PushoverApiToken, n.options.UserKey, strings.Join(n.options.PushoverDevices, ","))
		err := shoutrrr.Send(url, message)
		if err != nil {
			return err
		}
	}

	if n.options.Teams {
		teamsTokens := strings.TrimPrefix(n.options.TeamsWebHookURL, "https://outlook.office.com/webhook/")
		teamsTokens = strings.ReplaceAll(teamsTokens, "IncomingWebhook/", "")
		url := fmt.Sprintf("teams://%s", teamsTokens)
		err := shoutrrr.Send(url, message)
		if err != nil {
			return err
		}
	}

	return nil
}
