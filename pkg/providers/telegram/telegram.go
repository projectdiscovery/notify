package telegram

import (
	"fmt"

	"github.com/containrrr/shoutrrr"
	"github.com/projectdiscovery/notify/pkg/utils"
)

type Provider struct {
	Telegram []*Options `yaml:"telegram,omitempty"`
}

type Options struct {
	Profile        string `yaml:"profile,omitempty"`
	TelegramAPIKey string `yaml:"telegram_api_key,omitempty"`
	TelegramChatID string `yaml:"telegram_chat_id,omitempty"`
}

func New(options []*Options, profiles []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(profiles) == 0 || utils.Contains(profiles, o.Profile) {
			provider.Telegram = append(provider.Telegram, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message string) error {

	for _, pr := range p.Telegram {
		url := fmt.Sprintf("telegram://%s@telegram?channels=%s", pr.TelegramAPIKey, pr.TelegramChatID)
		err := shoutrrr.Send(url, message)
		if err != nil {
			return err
		}
	}
	return nil
}
