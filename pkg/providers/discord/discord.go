package discord

import (
	"fmt"

	"github.com/containrrr/shoutrrr"
	"github.com/oriser/regroup"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

var reDiscordWebhook = regroup.MustCompile(`(?P<scheme>https?):\/\/(?P<domain>(?:ptb\.|canary\.)?discord(?:app)?\.com)\/api(?:\/)?(?P<api_version>v\d{1,2})?\/webhooks\/(?P<webhook_identifier>\d{17,19})\/(?P<webhook_token>[\w\-]{68})`)

type Provider struct {
	Discord []*Options `yaml:"discord,omitempty"`
	counter int
}

type Options struct {
	ID                      string `yaml:"id,omitempty"`
	DiscordWebHookURL       string `yaml:"discord_webhook_url,omitempty"`
	DiscordWebHookUsername  string `yaml:"discord_username,omitempty"`
	DiscordWebHookAvatarURL string `yaml:"discord_avatar,omitempty"`
	DiscordThreads          bool   `yaml:"discord_threads,omitempty"`
	DiscordThreadID         string `yaml:"discord_thread_id,omitempty"`
	DiscordFormat           string `yaml:"discord_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Discord = append(provider.Discord, o)
		}
	}

	provider.counter = 0

	return provider, nil
}
func (p *Provider) Send(message, CliFormat string) error {
	var errs []error
	for _, pr := range p.Discord {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.DiscordFormat), p.counter)

		if pr.DiscordThreads {
			if pr.DiscordThreadID == "" {
				err := fmt.Errorf("thread_id value is required when discord_threads is set to true. check your configuration at id: %s", pr.ID)
				errs = append(errs, err)
				continue
			}
			if err := pr.SendThreaded(msg); err != nil {
				err = errors.Wrapf(err, "failed to send discord notification for id: %s ", pr.ID)
				errs = append(errs, err)
				continue
			}

		} else {
			matchedGroups, err := reDiscordWebhook.Groups(pr.DiscordWebHookURL)
			if err != nil {
				err := fmt.Errorf("incorrect discord configuration for id: %s ", pr.ID)
				errs = append(errs, err)
				continue
			}

			webhookID, webhookToken := matchedGroups["webhook_identifier"], matchedGroups["webhook_token"]

			//Reference: https://containrrr.dev/shoutrrr/v0.6/getting-started/
			url := fmt.Sprintf("discord://%s@%s?username=%s&avatarurl=%s&splitlines=no",
				webhookToken,
				webhookID,
				pr.DiscordWebHookUsername,
				pr.DiscordWebHookAvatarURL)
			if err := shoutrrr.Send(url, msg); err != nil {
				errs = append(errs, errors.Wrapf(err, "failed to send discord notification for id: %s ", pr.ID))
			}
		}

		gologger.Verbose().Msgf("discord notification sent for id: %s", pr.ID)
	}
	return multierr.Combine(errs...)
}