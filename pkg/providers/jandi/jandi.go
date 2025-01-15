package jandi

import (
	"github.com/oriser/regroup"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

var reJandiWebhook = regroup.MustCompile(`(?P<scheme>https?):\/\/(?P<domain>(?:ptb\.|canary\.)?Jandi(?:app)?\.com)\/api(?:\/)?(?P<api_version>v\d{1,2})?\/webhooks\/(?P<webhook_identifier>\d{17,19})\/(?P<webhook_token>[\w\-]{68})`)

type Provider struct {
	Jandi []*Options `yaml:"jandi,omitempty"`
	counter int
}

type Options struct {
	ID              string `yaml:"id,omitempty"`
	JandiWebHookURL string `yaml:"jandi_webhook_url,omitempty"`
	JandiFormat     string `yaml:"jandi_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Jandi = append(provider.Jandi, o)
		}
	}

	provider.counter = 0

	return provider, nil
}
func (p *Provider) Send(message, CliFormat string) error {
	var errs []error
	for _, pr := range p.Jandi {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.JandiFormat), p.counter)

		if err := pr.SendMessage(msg); err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to send jandi notification for id: %s ", pr.ID))
			continue
		}

		gologger.Verbose().Msgf("jandi notification sent for id: %s", pr.ID)
	}
	return multierr.Combine(errs...)
}