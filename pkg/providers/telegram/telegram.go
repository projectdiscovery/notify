package telegram

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

type Provider struct {
	Telegram []*Options `yaml:"telegram,omitempty"`
	counter  int
}

type Options struct {
	ID                string `yaml:"id,omitempty"`
	TelegramAPIKey    string `yaml:"telegram_api_key,omitempty"`
	TelegramChatID    string `yaml:"telegram_chat_id,omitempty"`
	TelegramThreadID  string `yaml:"telegram_thread_id,omitempty"`
	TelegramFormat    string `yaml:"telegram_format,omitempty"`
	TelegramParseMode string `yaml:"telegram_parsemode,omitempty"`
	// TelegramPreview enables link previews. Previews stay off unless this is set.
	TelegramPreview bool `yaml:"telegram_preview,omitempty"`
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

func telegramURL(pr *Options) string {
	preview := "No"
	if pr.TelegramPreview {
		preview = "Yes"
	}
	channel := pr.TelegramChatID
	if pr.TelegramThreadID != "" && !strings.Contains(pr.TelegramChatID, ":") {
		channel = fmt.Sprintf("%s:%s", pr.TelegramChatID, pr.TelegramThreadID)
	}
	return fmt.Sprintf("telegram://%s@telegram?channels=%s&parsemode=%s&preview=%s", pr.TelegramAPIKey, channel, pr.TelegramParseMode, preview)
}

func (p *Provider) Send(message, CliFormat string) error {
	var TelegramErr error
	p.counter++
	for _, pr := range p.Telegram {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.TelegramFormat), p.counter)
		if pr.TelegramParseMode == "" {
			pr.TelegramParseMode = "None"
		}
		url := telegramURL(pr)
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
