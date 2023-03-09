package discord

import (
	"fmt"
	"net/url"

	"github.com/containrrr/shoutrrr"
	"github.com/oriser/regroup"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

type Provider struct {
	Discord []*Options `yaml:"discord,omitempty"`
	counter int
}

type Options struct {
	ID                      string `yaml:"id,omitempty"`
	DiscordWebHookURL       string `yaml:"discord_webhook_url,omitempty"`
	DiscordWebHookUsername  string `yaml:"discord_username,omitempty"`
	DiscordWebHookAvatarURL string `yaml:"discord_avatar,omitempty"`
	DiscordFormat           string `yaml:"discord_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Discord = append(provider.Discord, o)
		}
	}

	provider.counter = 0

	return provider, nil
}
func (p *Provider) Send(message, CliFormat string) error {
	var DiscordErr error
	p.counter++
	for _, pr := range p.Discord {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.DiscordFormat), p.counter)

		discordWebhookRegex := regroup.MustCompile(`(?P<scheme>https?):\/\/(?P<domain>(?:ptb\.|canary\.)?discord(?:app)?\.com)\/api(?:\/)?(?P<api_version>v\d{1,2})?\/webhooks\/(?P<webhook_identifier>\d{17,19})\/(?P<webhook_token>[\w\-]{68})`)
		matchedGroups, err := discordWebhookRegex.Groups(pr.DiscordWebHookURL)

		if err != nil {
			err := fmt.Errorf("incorrect discord configuration for id: %s ", pr.ID)
			DiscordErr = multierr.Append(DiscordErr, err)
			continue
		}

		webhookID, webhookToken := matchedGroups["webhook_identifier"], matchedGroups["webhook_token"]
		url := fmt.Sprintf("discord://%s@%s?splitlines=no&username=%s", webhookToken, webhookID,
			url.QueryEscape(pr.DiscordWebHookUsername))

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
