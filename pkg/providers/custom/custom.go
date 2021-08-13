package custom

import (
	"bytes"

	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/retryablehttp-go"
)

type Provider struct {
	Custom []*Options `yaml:"custom,omitempty"`
}

type Options struct {
	ID               string            `yaml:"id,omitempty"`
	CustomWebhookURL string            `yaml:"custom_webook_url,omitempty"`
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

	for _, pr := range p.Custom {

		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.CustomFormat))
		body := bytes.NewBufferString(msg)

		r, err := retryablehttp.NewRequest(pr.CustomMethod, pr.CustomWebhookURL, body)
		if err != nil {
			return err
		}

		for k, v := range pr.CustomHeaders {
			r.Header.Set(k, v)
		}

		client := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)
		_, err = client.Do(r)
		if err != nil {
			return err
		}
	}
	return nil
}
