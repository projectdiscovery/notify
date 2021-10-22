package discord

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
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
	DiscordFormat           string `yaml:"discord_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || utils.Contains(ids, o.ID) {
			provider.Discord = append(provider.Discord, o)
		}
	}

	return provider, nil
}
func (p *Provider) Send(message, CliFormat string) error {
	var DiscordErr error

	for _, pr := range p.Discord {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.DiscordFormat))
		discordRegex := regexp.MustCompile(`https://(discord.com|discordapp.com)/api/webhooks/`)

		discordTokens := strings.TrimPrefix(pr.DiscordWebHookURL, discordRegex.FindString(pr.DiscordWebHookURL))
		tokens := strings.Split(discordTokens, "/")
		if len(tokens) < 2 {
			err := fmt.Errorf("incorrect discord configuration for id: %s ", pr.ID)
			DiscordErr = multierr.Append(DiscordErr, err)
			continue
		}
		webhookID, token := tokens[0], tokens[1]
		url := fmt.Sprintf("discord://%s@%s?splitlines=no", token, webhookID)
		err := shoutrrr.Send(url, msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send discord notification for id: %s ", pr.ID))
			DiscordErr = multierr.Append(DiscordErr, err)
			continue
		}
		gologger.Verbose().Msgf("discord notification sent for id: %s", pr.ID)
	}
	return DiscordErr
}
