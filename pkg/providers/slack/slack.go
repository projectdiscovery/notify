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
	Profile         string `yaml:"profile,omitempty"`
	SlackWebHookURL string `yaml:"slack_webhook_url,omitempty"`
	SlackUsername   string `yaml:"slack_username,omitempty"`
	SlackChannel    string `yaml:"slack_channel,omitempty"`
	SlackThreadTS   string `yaml:"slack_thread_ts,omitempty"`
	SlackThreads    bool   `yaml:"slack_threads,omitempty"`
	SlackToken      string `yaml:"slack_token,omitempty"`
}

func New(options []*Options, profiles []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(profiles) == 0 || utils.Contains(profiles, o.Profile) {
			provider.Slack = append(provider.Slack, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message string) error {

	for _, pr := range p.Slack {

		if pr.SlackThreads {
			if pr.SlackToken == "" {
				return errors.New("can't start a slack thread without slack_token value in provider config")
			}
			if pr.SlackChannel == "" {
				return errors.New("can't start a slack thread without slack_channel value in provider config")
			}
			return pr.SendThreaded(message)
		} else {
			slackTokens := strings.TrimPrefix(pr.SlackWebHookURL, "https://hooks.slack.com/services/")
			url := &url.URL{
				Scheme: "slack",
				Path:   slackTokens,
			}

			err := shoutrrr.Send(url.String(), message)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
