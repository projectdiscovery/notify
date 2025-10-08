package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

type Provider struct {
	Slack   []*Options `yaml:"slack,omitempty"`
	counter int
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
	SlackIconEmoji  string `yaml:"slack_icon_emoji,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Slack = append(provider.Slack, o)
		}
	}

	provider.counter = 0

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var SlackErr error
	p.counter++
	for _, pr := range p.Slack {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.SlackFormat), p.counter)

		// Handle threaded messages separately
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
			// Send via webhook with emoji and username
			if !strings.HasPrefix(pr.SlackWebHookURL, "https://hooks.slack.com/services/") {
				err := errors.Wrap(fmt.Errorf("invalid slack webhook URL"),
					fmt.Sprintf("failed to send slack notification for id: %s ", pr.ID))
				SlackErr = multierr.Append(SlackErr, err)
				continue
			}

			payload := map[string]interface{}{
				"text": msg,
			}
			if pr.SlackUsername != "" {
				payload["username"] = pr.SlackUsername
			}
			if pr.SlackIconEmoji != "" {
				payload["icon_emoji"] = pr.SlackIconEmoji
			}

			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("failed to marshal Slack payload for id: %s", pr.ID))
				SlackErr = multierr.Append(SlackErr, err)
				continue
			}

			resp, err := http.Post(pr.SlackWebHookURL, "application/json", bytes.NewBuffer(jsonPayload))
			if err != nil || resp.StatusCode >= 400 {
				if err == nil {
					err = fmt.Errorf("received non-success status: %s", resp.Status)
				}
				err = errors.Wrap(err, fmt.Sprintf("failed to send slack notification for id: %s", pr.ID))
				SlackErr = multierr.Append(SlackErr, err)
				continue
			}
		}
		gologger.Verbose().Msgf("Slack notification sent for id: %s", pr.ID)
	}
	return SlackErr
}
