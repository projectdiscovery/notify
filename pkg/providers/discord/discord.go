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
	Profile                 string `yaml:"profile,omitempty"`
	DiscordWebHookURL       string `yaml:"discord_webhook_url,omitempty"`
	DiscordWebHookUsername  string `yaml:"discord_username,omitempty"`
	DiscordWebHookAvatarURL string `yaml:"discord_avatar,omitempty"`
}

func New(options []*Options, profiles []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(profiles) == 0 || utils.Contains(profiles, o.Profile) {
			provider.Discord = append(provider.Discord, o)
		}
	}

	return provider, nil
}
func (p *Provider) Send(message string) error {

	for _, pr := range p.Discord {
		discordTokens := strings.TrimPrefix(pr.DiscordWebHookURL, "https://discord.com/api/webhooks/")
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
	return nil
}
