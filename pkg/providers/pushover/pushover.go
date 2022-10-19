package pushover

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/sliceutil"
)

type Provider struct {
	Pushover []*Options `yaml:"pushover,omitempty"`
}

type Options struct {
	ID               string   `yaml:"id,omitempty"`
	PushoverApiToken string   `yaml:"pushover_api_token,omitempty"`
	UserKey          string   `yaml:"pushover_user_key,omitempty"`
	PushoverDevices  []string `yaml:"pushover_devices,omitempty"`
	PushoverFormat   string   `yaml:"pushover_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Pushover = append(provider.Pushover, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var PushoverErr error
	for _, pr := range p.Pushover {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.PushoverFormat))

		url := fmt.Sprintf("pushover://shoutrrr:%s@%s/?devices=%s", pr.PushoverApiToken, pr.UserKey, strings.Join(pr.PushoverDevices, ","))
		err := shoutrrr.Send(url, msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send pushover notification for id: %s ", pr.ID))
			PushoverErr = multierr.Append(PushoverErr, err)
			continue
		}
		gologger.Verbose().Msgf("pushover notification sent for id: %s", pr.ID)
	}
	return PushoverErr
}
