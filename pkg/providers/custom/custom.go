package custom

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/notify/pkg/utils/httpreq"
)

type Provider struct {
	Custom []*Options `yaml:"custom,omitempty"`
}

type Options struct {
	ID               string            `yaml:"id,omitempty"`
	CustomWebhookURL string            `yaml:"custom_webhook_url,omitempty"`
	CustomMethod     string            `yaml:"custom_method,omitempty"`
	CustomHeaders    map[string]string `yaml:"custom_headers,omitempty"`
	CustomFormat     string            `yaml:"custom_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || utils.Contains(ids, o.ID) {
			provider.Custom = append(provider.Custom, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var CustomErr error

	for _, pr := range p.Custom {

		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.CustomFormat))
		body := bytes.NewBufferString(msg)

		r, err := http.NewRequest(pr.CustomMethod, pr.CustomWebhookURL, body)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send custom notification for id: %s ", pr.ID))
			CustomErr = multierr.Append(CustomErr, err)
			continue
		}

		for k, v := range pr.CustomHeaders {
			r.Header.Set(k, v)
		}

		_, err = httpreq.NewClient().Do(r)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send custom notification for id: %s ", pr.ID))
			CustomErr = multierr.Append(CustomErr, err)
			continue
		}
		gologger.Verbose().Msgf("custom notification sent for id: %s", pr.ID)
	}
	return CustomErr
}
