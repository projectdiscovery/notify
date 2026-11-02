package rocketchat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

const maxMessageLength = 4000

type Provider struct {
	RocketChat []*Options `yaml:"rocketchat,omitempty"`
	counter    int
}

type Options struct {
	ID                   string `yaml:"id,omitempty"`
	RocketChatWebhookURL string `yaml:"rocketchat_webhook_url,omitempty"`
	Channel              string `yaml:"rocketchat_channel,omitempty"`
	Username             string `yaml:"rocketchat_username,omitempty"`
	RocketChatFormat     string `yaml:"rocketchat_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.RocketChat = append(provider.RocketChat, o)
		}
	}

	provider.counter = 0

	return provider, nil
}

var rocketchatHTTPClient = &http.Client{Timeout: 15 * time.Second}

type webhookPayload struct {
	Text    string `json:"text"`
	Channel string `json:"channel,omitempty"`
	Alias   string `json:"alias,omitempty"`
}

func (p *Provider) Send(message, CliFormat string) error {
	var rocketChatErr error
	p.counter++
	for _, pr := range p.RocketChat {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.RocketChatFormat), p.counter)
		if err := sendRocketChat(pr, msg); err != nil {
			rocketChatErr = multierr.Append(rocketChatErr, err)
			continue
		}
		gologger.Verbose().Msgf("rocketchat notification sent for id: %s", pr.ID)
	}
	return rocketChatErr
}

func sendRocketChat(pr *Options, msg string) error {
	if pr.RocketChatWebhookURL == "" {
		return fmt.Errorf("rocketchat: webhook url is empty for id: %s", pr.ID)
	}

	chunks := splitMessage(msg, maxMessageLength)
	for _, chunk := range chunks {
		if strings.HasPrefix(pr.RocketChatWebhookURL, "rocketchat://") {
			if err := shoutrrr.Send(pr.RocketChatWebhookURL, chunk); err != nil {
				return errors.Wrapf(err, "failed to send shoutrrr rocketchat notification for id: %s", pr.ID)
			}
			continue
		}

		payload := webhookPayload{
			Text:    chunk,
			Channel: pr.Channel,
			Alias:   pr.Username,
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal rocketchat payload for id: %s", pr.ID)
		}

		resp, err := rocketchatHTTPClient.Post(pr.RocketChatWebhookURL, "application/json", bytes.NewReader(data))
		if err != nil {
			return errors.Wrapf(err, "failed to send rocketchat notification for id: %s", pr.ID)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("rocketchat: unexpected status code %d for id: %s", resp.StatusCode, pr.ID)
		}
	}

	return nil
}

func splitMessage(msg string, limit int) []string {
	if len(msg) <= limit {
		return []string{msg}
	}

	var chunks []string
	for len(msg) > limit {
		splitIndex := strings.LastIndex(msg[:limit], "\n")
		if splitIndex <= 0 {
			splitIndex = strings.LastIndex(msg[:limit], " ")
		}
		if splitIndex <= 0 {
			splitIndex = limit
		}

		chunks = append(chunks, msg[:splitIndex])
		msg = strings.TrimPrefix(msg[splitIndex:], "\n")
	}

	if len(msg) > 0 {
		chunks = append(chunks, msg)
	}

	return chunks
}
