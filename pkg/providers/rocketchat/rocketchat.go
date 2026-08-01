package rocketchat

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

// defaultMaxMessageLength matches Rocket.Chat's default Message_MaxAllowedSize setting.
const defaultMaxMessageLength = 5000

type Provider struct {
	Rocketchat []*Options `yaml:"rocketchat,omitempty"`
	counter    int
}

type Options struct {
	ID                   string `yaml:"id,omitempty"`
	RocketchatWebHookURL string `yaml:"rocketchat_webhook_url,omitempty"`
	RocketchatChannel    string `yaml:"rocketchat_channel,omitempty"`
	RocketchatUsername   string `yaml:"rocketchat_username,omitempty"`
	RocketchatAvatar     string `yaml:"rocketchat_avatar,omitempty"`
	RocketchatEmoji      string `yaml:"rocketchat_emoji,omitempty"`
	RocketchatFormat     string `yaml:"rocketchat_format,omitempty"`
	// RocketchatAttachment sends the message as an attachment instead of plain text.
	RocketchatAttachment bool `yaml:"rocketchat_attachment,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Rocketchat = append(provider.Rocketchat, o)
		}
	}

	provider.counter = 0

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var rocketchatErr error
	p.counter++

	for _, pr := range p.Rocketchat {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.RocketchatFormat), p.counter)

		if !strings.HasPrefix(pr.RocketchatWebHookURL, "http://") && !strings.HasPrefix(pr.RocketchatWebHookURL, "https://") {
			err := errors.Wrap(fmt.Errorf("invalid rocketchat webhook URL"),
				fmt.Sprintf("failed to send rocketchat notification for id: %s ", pr.ID))
			rocketchatErr = multierr.Append(rocketchatErr, err)
			continue
		}

		for _, chunk := range splitMessage(msg, defaultMaxMessageLength) {
			payload := WebhookPayload{
				Channel: pr.RocketchatChannel,
				Alias:   pr.RocketchatUsername,
				Avatar:  pr.RocketchatAvatar,
				Emoji:   pr.RocketchatEmoji,
			}
			if pr.RocketchatAttachment {
				payload.Attachments = []Attachment{{Text: chunk}}
			} else {
				payload.Text = chunk
			}

			if err := send(pr.RocketchatWebHookURL, payload); err != nil {
				err = errors.Wrap(err, fmt.Sprintf("failed to send rocketchat notification for id: %s", pr.ID))
				rocketchatErr = multierr.Append(rocketchatErr, err)
				continue
			}
		}
		gologger.Verbose().Msgf("rocketchat notification sent for id: %s", pr.ID)
	}
	return rocketchatErr
}

func send(webhookURL string, payload WebhookPayload) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal rocketchat payload")
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("received non-success status: %s", resp.Status)
	}
	return nil
}

// splitMessage splits msg into chunks no longer than limit, breaking on newlines
// where possible so multi-line messages aren't cut mid-line.
func splitMessage(msg string, limit int) []string {
	if len(msg) <= limit {
		return []string{msg}
	}

	var chunks []string
	var current strings.Builder

	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		for len(line) > limit {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
			}
			chunks = append(chunks, line[:limit])
			line = line[limit:]
		}

		if current.Len() > 0 && current.Len()+1+len(line) > limit {
			chunks = append(chunks, current.String())
			current.Reset()
		}

		if current.Len() > 0 {
			current.WriteByte('\n')
		}
		current.WriteString(line)
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks
}
