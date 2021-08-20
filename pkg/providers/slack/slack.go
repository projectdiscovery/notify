package slack

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/notify/pkg/utils"
	"go.uber.org/multierr"
)

type Provider struct {
	Slack []*Options `yaml:"slack,omitempty"`
}

type Options struct {
	ID              string `yaml:"id,omitempty"`
	SlackWebHookURL string `yaml:"slack_webhook_url,omitempty"`
	SlackUsername   string `yaml:"slack_username,omitempty"`
	SlackChannel    string `yaml:"slack_channel,omitempty"`
	SlackFormat     string `yaml:"slack_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || utils.Contains(ids, o.ID) {
			provider.Slack = append(provider.Slack, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var SlackErr error
	for _, pr := range p.Slack {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.SlackFormat))

		slackTokens := strings.TrimPrefix(pr.SlackWebHookURL, "https://hooks.slack.com/services/")
		url := fmt.Sprintf("slack://%s", slackTokens)
		err := shoutrrr.Send(url, msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send slack notification for id: %s ", pr.ID))
			SlackErr = multierr.Append(SlackErr, err)
		}
	}
	return SlackErr
}
