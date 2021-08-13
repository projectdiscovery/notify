package slack

import (
	"errors"
	"net/url"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/projectdiscovery/notify/pkg/utils"
)

type Provider struct {
	Slack []*Options `yaml:"slack,omitempty"`
}

type Options struct {
	ID              string `yaml:"id,omitempty"`
	SlackWebHookURL string `yaml:"slack_webhook_url,omitempty"`
	SlackUsername   string `yaml:"slack_username,omitempty"`
	SlackChannel    string `yaml:"slack_channel,omitempty"`
	SlackThreads    bool   `yaml:"slack_threads,omitempty"`
	SlackThreadTS   string `yaml:"slack_thread_ts,omitempty"`
	SlackToken      string `yaml:"slack_token,omitempty"`
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
	for _, pr := range p.Slack {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.SlackFormat))

		if pr.SlackThreads {
			if pr.SlackToken == "" {
				return errors.New("can't start a slack thread without slack_token value in provider config")
			}
			if pr.SlackChannel == "" {
				return errors.New("can't start a slack thread without slack_channel value in provider config")
			}
			return pr.SendThreaded(msg)
		} else {
			slackTokens := strings.TrimPrefix(pr.SlackWebHookURL, "https://hooks.slack.com/services/")
			url := &url.URL{
				Scheme: "slack",
				Path:   slackTokens,
			}

			err := shoutrrr.Send(url.String(), msg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
