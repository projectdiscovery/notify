package providers

import (
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/providers/custom"
	"github.com/projectdiscovery/notify/pkg/providers/discord"
	"github.com/projectdiscovery/notify/pkg/providers/googlechat"
	"github.com/projectdiscovery/notify/pkg/providers/gotify"
	"github.com/projectdiscovery/notify/pkg/providers/pushover"
	"github.com/projectdiscovery/notify/pkg/providers/rocketchat"
	"github.com/projectdiscovery/notify/pkg/providers/slack"
	"github.com/projectdiscovery/notify/pkg/providers/smtp"
	"github.com/projectdiscovery/notify/pkg/providers/teams"
	"github.com/projectdiscovery/notify/pkg/providers/telegram"
	"github.com/projectdiscovery/notify/pkg/types"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

// ProviderOptions is configuration for notify providers
type ProviderOptions struct {
	Slack      []*slack.Options      `yaml:"slack,omitempty"`
	Discord    []*discord.Options    `yaml:"discord,omitempty"`
	Pushover   []*pushover.Options   `yaml:"pushover,omitempty"`
	SMTP       []*smtp.Options       `yaml:"smtp,omitempty"`
	Teams      []*teams.Options      `yaml:"teams,omitempty"`
	Telegram   []*telegram.Options   `yaml:"telegram,omitempty"`
	GoogleChat []*googlechat.Options `yaml:"googlechat,omitempty"`
	Custom     []*custom.Options     `yaml:"custom,omitempty"`
	Gotify     []*gotify.Options     `yaml:"gotify,omitempty"`
	RocketChat []*rocketchat.Options `yaml:"rocketchat,omitempty"`
}

// supportedProviders lists the provider names that -provider/-p accepts.  Each
// one matches a top-level key of the provider config file.
var supportedProviders = []string{
	"slack",
	"discord",
	"pushover",
	"smtp",
	"teams",
	"telegram",
	"googlechat",
	"custom",
	"gotify",
	"rocketchat",
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

	// A name that does not belong to any provider group used to be ignored
	// silently, so "notify -p slackk" exited successfully without sending
	// anything at all.
	for _, name := range options.Providers {
		if !sliceutil.Contains(supportedProviders, name) {
			return nil, errors.Errorf("unknown provider %q (available: %s)", name, strings.Join(supportedProviders, ", "))
		}
	}

	if providerOptions.Slack != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "slack")) {

		provider, err := slack.New(providerOptions.Slack, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create slack provider client")
		}

		client.providers = append(client.providers, provider)
	}
	if providerOptions.Discord != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "discord")) {

		provider, err := discord.New(providerOptions.Discord, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create discord provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.Pushover != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "pushover")) {

		provider, err := pushover.New(providerOptions.Pushover, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create pushover provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.GoogleChat != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "googlechat")) {

		provider, err := googlechat.New(providerOptions.GoogleChat, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create googlechat provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.SMTP != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "smtp")) {

		provider, err := smtp.New(providerOptions.SMTP, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create smtp provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.Teams != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "teams")) {

		provider, err := teams.New(providerOptions.Teams, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create teams provider client")
		}
		client.providers = append(client.providers, provider)
	}
	if providerOptions.Telegram != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "telegram")) {

		provider, err := telegram.New(providerOptions.Telegram, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create telegram provider client")
		}
		client.providers = append(client.providers, provider)
	}

	if providerOptions.Custom != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "custom")) {

		provider, err := custom.New(providerOptions.Custom, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create custom provider client")
		}
		client.providers = append(client.providers, provider)
	}

	if providerOptions.Gotify != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "gotify")) {

		provider, err := gotify.New(providerOptions.Gotify, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create gotify provider client")
		}
		client.providers = append(client.providers, provider)
	}

	if providerOptions.RocketChat != nil && (len(options.Providers) == 0 || sliceutil.Contains(options.Providers, "rocketchat")) {

		provider, err := rocketchat.New(providerOptions.RocketChat, options.IDs)
		if err != nil {
			return nil, errors.Wrap(err, "could not create rocketchat provider client")
		}
		client.providers = append(client.providers, provider)
	}

	// The names are valid but none of them is present in the provider config
	// file, which would otherwise run without sending anything.
	if len(options.Providers) > 0 && len(client.providers) == 0 {
		return nil, errors.Errorf("none of the requested providers (%s) are configured in the provider config file", strings.Join(options.Providers, ", "))
	}

	return client, nil
}

func (p *Client) Send(message string) error {

	// strip unsupported color control chars
	message = stripansi.Strip(message)

	for _, v := range p.providers {
		if err := v.Send(message, p.options.MessageFormat); err != nil {
			for _, v := range multierr.Errors(err) {
				gologger.Error().Msgf("%s", v)
			}
		}
	}

	return nil
}
