package providers

import (
	"github.com/acarl005/stripansi"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/notify/pkg/providers/discord"
	"github.com/projectdiscovery/notify/pkg/providers/pushover"
	"github.com/projectdiscovery/notify/pkg/providers/slack"
	"github.com/projectdiscovery/notify/pkg/providers/smtp"
	"github.com/projectdiscovery/notify/pkg/providers/teams"
	"github.com/projectdiscovery/notify/pkg/providers/telegram"
	"github.com/projectdiscovery/notify/pkg/utils"
)

// Options is a configuration file for nuclei reporting module
type Options struct {
	Slack    []*slack.Options    `yaml:"slack,omitempty"`
	Discord  []*discord.Options  `yaml:"discord,omitempty"`
	Pushover []*pushover.Options `yaml:"pushover,omitempty"`
	SMTP     []*smtp.Options     `yaml:"smtp,omitempty"`
	Teams    []*teams.Options    `yaml:"teams,omitempty"`
	Telegram []*telegram.Options `yaml:"telegram,omitempty"`
}

// Provider is an interface implemented by providers
type Provider interface {
	Send(string) error
}

type Client struct {
	providers []Provider
	options   *Options
}

func New(options *Options, providers, profiles []string) (*Client, error) {

	client := &Client{options: options}

	if options.Slack != nil && (len(providers) == 0 || utils.Contains(providers, "slack")) {

		provider, err := slack.New(options.Slack, profiles)
		if err != nil {
			return nil, errors.Wrap(err, "could not create slack provider client")
		}

		client.providers = append(client.providers, provider)
	}
	if options.Discord != nil && (len(providers) == 0 || utils.Contains(providers, "discord")) {

		provider, err := discord.New(options.Discord, profiles)
		if err != nil {
			return nil, errors.Wrap(err, "could not create discord provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if options.Pushover != nil && (len(providers) == 0 || utils.Contains(providers, "pushover")) {

		provider, err := pushover.New(options.Pushover, profiles)
		if err != nil {
			return nil, errors.Wrap(err, "could not create pushover provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if options.SMTP != nil && (len(providers) == 0 || utils.Contains(providers, "smtp")) {

		provider, err := smtp.New(options.SMTP, profiles)
		if err != nil {
			return nil, errors.Wrap(err, "could not create smtp provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if options.Teams != nil && (len(providers) == 0 || utils.Contains(providers, "teams")) {

		provider, err := teams.New(options.Teams, profiles)
		if err != nil {
			return nil, errors.Wrap(err, "could not create teams provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if options.Telegram != nil && (len(providers) == 0 || utils.Contains(providers, "telegram")) {

		provider, err := telegram.New(options.Telegram, profiles)
		if err != nil {
			return nil, errors.Wrap(err, "could not create telegram provider client")
		}
		client.providers = append(client.providers, provider)
	}

	return client, nil
}

func (p *Client) Send(message string) error {

	// strip unsupported color control chars
	message = stripansi.Strip(message)

	for _, v := range p.providers {
		//nolint:errcheck
		v.Send(message)
	}

	return nil
}
