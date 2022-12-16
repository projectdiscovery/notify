package telegram

import (
	"fmt"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/sliceutil"
)

type Provider struct {
	Telegram []*Options `yaml:"telegram,omitempty"`
	counter  int
}

type Options struct {
	ID                string `yaml:"id,omitempty"`
	TelegramAPIKey    string `yaml:"telegram_api_key,omitempty"`
	TelegramChatID    string `yaml:"telegram_chat_id,omitempty"`
	TelegramFormat    string `yaml:"telegram_format,omitempty"`
	TelegramParseMode string `yaml:"telegram_parsemode,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Telegram = append(provider.Telegram, o)
		}
	}

	provider.counter = 0

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var TelegramErr error
	p.counter++
	for _, pr := range p.Telegram {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.TelegramFormat), p.counter)
		if pr.TelegramParseMode == "" {
			pr.TelegramParseMode = "None"
		}
		url := fmt.Sprintf("telegram://%s@telegram?channels=%s&parsemode=%s", pr.TelegramAPIKey, pr.TelegramChatID, pr.TelegramParseMode)
		err := shoutrrr.Send(url, msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send telegram notification for id: %s ", pr.ID))
			TelegramErr = multierr.Append(TelegramErr, err)
			continue
		}
		gologger.Verbose().Msgf("telegram notification sent for id: %s", pr.ID)
	}
	return TelegramErr
}
