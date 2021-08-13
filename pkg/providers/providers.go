package providers

import (
	"github.com/acarl005/stripansi"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/providers/custom"
	"github.com/projectdiscovery/notify/pkg/providers/discord"
	"github.com/projectdiscovery/notify/pkg/providers/pushover"
	"github.com/projectdiscovery/notify/pkg/providers/slack"
	"github.com/projectdiscovery/notify/pkg/providers/smtp"
	"github.com/projectdiscovery/notify/pkg/providers/teams"
	"github.com/projectdiscovery/notify/pkg/providers/telegram"
	"github.com/projectdiscovery/notify/pkg/types"
	"github.com/projectdiscovery/notify/pkg/utils"
)

// ProviderOptions is configuration for notify providers
type ProviderOptions struct {
	Slack    []*slack.Options    `yaml:"slack,omitempty"`
	Discord  []*discord.Options  `yaml:"discord,omitempty"`
	Pushover []*pushover.Options `yaml:"pushover,omitempty"`
	SMTP     []*smtp.Options     `yaml:"smtp,omitempty"`
	Teams    []*teams.Options    `yaml:"teams,omitempty"`
	Telegram []*telegram.Options `yaml:"telegram,omitempty"`
	Custom   []*custom.Options   `yaml:"custom,omitempty"`
}

// Provider is an interface implemented by providers
type Provider interface {
	Send(message, CliFormat string) error
}

type Client struct {
	providers       []Provider
	providerOptions *ProviderOptions
	options         *types.Options
}

func New(providerOptions *ProviderOptions, options *types.Options) (*Client, error) {

	client := &Client{providerOptions: providerOptions, options: options}

	if providerOptions.Slack != nil && (len(options.Providers) == 0 || utils.Contains(options.Providers, "slack")) {

		provider, err := slack.New(providerOptions.Slack, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create slack provider client")
		}

		client.providers = append(client.providers, provider)
	}
	if providerOptions.Discord != nil && (len(options.Providers) == 0 || utils.Contains(options.Providers, "discord")) {

		provider, err := discord.New(providerOptions.Discord, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create discord provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.Pushover != nil && (len(options.Providers) == 0 || utils.Contains(options.Providers, "pushover")) {

		provider, err := pushover.New(providerOptions.Pushover, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create pushover provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.SMTP != nil && (len(options.Providers) == 0 || utils.Contains(options.Providers, "smtp")) {

		provider, err := smtp.New(providerOptions.SMTP, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create smtp provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.Teams != nil && (len(options.Providers) == 0 || utils.Contains(options.Providers, "teams")) {

		provider, err := teams.New(providerOptions.Teams, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create teams provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.Telegram != nil && (len(options.Providers) == 0 || utils.Contains(options.Providers, "telegram")) {

		provider, err := telegram.New(providerOptions.Telegram, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create telegram provider client")
		}
		client.providers = append(client.providers, provider)
	}

	if options.Custom != nil && (len(providers) == 0 || utils.Contains(providers, "custom")) {

		provider, err := custom.New(options.Custom, ids)
		if err != nil {
			return nil, errors.Wrap(err, "could not create custom provider client")
		}
		client.providers = append(client.providers, provider)
	}

	return client, nil
}

func (p *Client) Send(message string) error {

	// strip unsupported color control chars
	message = stripansi.Strip(message)

	for _, v := range p.providers {
		if err := v.Send(message, p.options.MessageFormat); err != nil {
			gologger.Error().Msgf("error while sending message: %s", err)
		}
	}

	return nil
}
