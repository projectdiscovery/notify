package custom

import (
	"bytes"
	"strings"

	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/retryablehttp-go"
)

type Provider struct {
	Custom []*Options `yaml:"custom,omitempty"`
}

type Options struct {
	Profile          string            `yaml:"profile,omitempty"`
	CustomWebhookURL string            `yaml:"custom_webook_url,omitempty"`
	CustomMethod     string            `yaml:"custom_method,omitempty"`
	CustomHeaders    map[string]string `yaml:"custom_headers,omitempty"`
	CustomBody       string            `yaml:"custom_body,omitempty"`
}

func New(options []*Options, profiles []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(profiles) == 0 || utils.Contains(profiles, o.Profile) {
			provider.Custom = append(provider.Custom, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message string) error {

	for _, pr := range p.Custom {

		rr := strings.NewReplacer(
			"{{data}}", message,
		)
		final := rr.Replace(pr.CustomBody)

		r, err := retryablehttp.NewRequest(pr.CustomMethod, pr.CustomWebhookURL, bytes.NewReader([]byte(final)))
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
