package discord

import (
	"errors"
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/projectdiscovery/notify/pkg/utils"
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

	for _, pr := range p.Discord {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.DiscordFormat))

		discordTokens := strings.TrimPrefix(pr.DiscordWebHookURL, "https://discord.com/api/webhooks/")
		tokens := strings.Split(discordTokens, "/")
		if len(tokens) != 2 {
			return errors.New("Wrong discord configuration")
		}
		webhookID, token := tokens[0], tokens[1]
		url := fmt.Sprintf("discord://%s@%s", token, webhookID)
		err := shoutrrr.Send(url, msg)
		if err != nil {
			return err
		}
	}
	return nil
}
