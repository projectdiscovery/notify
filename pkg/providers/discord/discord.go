package discord

import (
	"fmt"

	"github.com/containrrr/shoutrrr"
	"github.com/oriser/regroup"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/sliceutil"
	"go.uber.org/multierr"
)

type Provider struct {
	Discord []*Options `yaml:"discord,omitempty"`
}

type Options struct {
	ID                      string `yaml:"id,omitempty"`
	DiscordWebHookURL       string `yaml:"discord_webhook_url,omitempty"`
	DiscordWebHookUsername  string `yaml:"discord_username,omitempty"`
	DiscordWebHookAvatarURL string `yaml:"discord_avatar,omitempty"`
	DiscordThreads          bool   `yaml:"discord_threads,omitempty"`
	DiscordThreadID         string `yaml:"discord_thread_id,omitempty"`
	DiscordFormat           string `yaml:"discord_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Discord = append(provider.Discord, o)
		}
	}

	return provider, nil
}
func (p *Provider) Send(message, CliFormat string) error {
	var DiscordErr error

	for _, pr := range p.Discord {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.DiscordFormat))

		discordWebhookRegex := regroup.MustCompile(`(?P<scheme>https?):\/\/(?P<domain>(?:ptb\.|canary\.)?discord(?:app)?\.com)\/api(?:\/)?(?P<api_version>v\d{1,2})?\/webhooks\/(?P<webhook_identifier>\d{17,19})\/(?P<webhook_token>[\w\-]{68})`)
		matchedGroups, err := discordWebhookRegex.Groups(pr.DiscordWebHookURL)

		if err != nil {
			err := fmt.Errorf("incorrect discord configuration for id: %s ", pr.ID)
			DiscordErr = multierr.Append(DiscordErr, err)
			continue
		}

		webhookID, webhookToken := matchedGroups["webhook_identifier"], matchedGroups["webhook_token"]
		url := fmt.Sprintf("discord://%s@%s?username=%s&avatarurl=%s&splitlines=no",
			webhookToken,
			webhookID,
			pr.DiscordWebHookUsername,
			pr.DiscordWebHookAvatarURL)
		sendErr := shoutrrr.Send(url, msg)
		if sendErr != nil {
			sendErr = errors.Wrap(sendErr, fmt.Sprintf("failed to send discord notification for id: %s ", pr.ID))
			DiscordErr = multierr.Append(DiscordErr, sendErr)
			continue
		}
		gologger.Verbose().Msgf("discord notification sent for id: %s", pr.ID)
	}
	return DiscordErr
}
