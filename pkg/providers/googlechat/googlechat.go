package googlechat

import (
	"fmt"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/sliceutil"
	"go.uber.org/multierr"
)

type Provider struct {
	GoogleChat []*Options `yaml:"googleChat,omitempty"`
}

type Options struct {
	ID               string `yaml:"id,omitempty"`
	Space            string `yaml:"space,omitempty"`
	Key              string `yaml:"key,omitempty"`
	Token            string `yaml:"token,omitempty"`
	GoogleChatFormat string `yaml:"google_chat_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.GoogleChat = append(provider.GoogleChat, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var GoogleChatErr error
	for _, pr := range p.GoogleChat {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.GoogleChatFormat))
		url := fmt.Sprintf("googlechat://chat.googleapis.com/v1/spaces/%s/messages?key=%s&token=%s", pr.Space, pr.Key, pr.Token)
		err := shoutrrr.Send(url, msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send googleChat notification for id: %s ", pr.ID))
			GoogleChatErr = multierr.Append(GoogleChatErr, err)
			continue
		}
		gologger.Verbose().Msgf("googleChat notification sent for id: %s", pr.ID)
	}
	return GoogleChatErr
}
