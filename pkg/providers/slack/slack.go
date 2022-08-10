package slack

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/sliceutil"
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
	SlackThreads    bool   `yaml:"slack_threads,omitempty"`
	SlackThreadTS   string `yaml:"slack_thread_ts,omitempty"`
	SlackToken      string `yaml:"slack_token,omitempty"`
	SlackFormat     string `yaml:"slack_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Slack = append(provider.Slack, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var SlackErr error
	for _, pr := range p.Slack {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.SlackFormat))

		if pr.SlackThreads {
			if pr.SlackToken == "" {
				err := errors.Wrap(fmt.Errorf("slack_token value is required to start a thread"),
					fmt.Sprintf("failed to send slack notification for id: %s ", pr.ID))
				SlackErr = multierr.Append(SlackErr, err)
				continue
			}
			if pr.SlackChannel == "" {
				err := errors.Wrap(fmt.Errorf("slack_channel value is required to start a thread"),
					fmt.Sprintf("failed to send slack notification for id: %s ", pr.ID))
				SlackErr = multierr.Append(SlackErr, err)
				continue
			}
			if err := pr.SendThreaded(msg); err != nil {
				err = errors.Wrap(err,
					fmt.Sprintf("failed to send slack notification for id: %s ", pr.ID))
				SlackErr = multierr.Append(SlackErr, err)
				continue
			}
		} else {
			slackTokens := strings.TrimPrefix(pr.SlackWebHookURL, "https://hooks.slack.com/services/")
			url := &url.URL{
				Scheme: "slack",
				Path:   slackTokens,
			}

			err := shoutrrr.Send(url.String(), msg)
			if err != nil {
				err = errors.Wrap(err,
					fmt.Sprintf("failed to send slack notification for id: %s ", pr.ID))
				SlackErr = multierr.Append(SlackErr, err)
				continue
			}
		}
		gologger.Verbose().Msgf("Slack notification sent successfully for id: %s", pr.ID)

	}
	return SlackErr
}
