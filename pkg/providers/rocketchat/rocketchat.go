package rocketchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/notify/pkg/utils/httpreq"
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

		providerSucceeded := true
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
				providerSucceeded = false
				err = errors.Wrap(err, fmt.Sprintf("failed to send rocketchat notification for id: %s", pr.ID))
				rocketchatErr = multierr.Append(rocketchatErr, err)
				continue
			}
		}
		if providerSucceeded {
			gologger.Verbose().Msgf("rocketchat notification sent for id: %s", pr.ID)
		}
	}
	return rocketchatErr
}

func send(webhookURL string, payload WebhookPayload) error {
	return sendWithClient(context.Background(), httpreq.NewClient(), webhookURL, payload)
}

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

func sendWithClient(ctx context.Context, client httpDoer, webhookURL string, payload WebhookPayload) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal rocketchat payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	closeErr := resp.Body.Close()
	if resp.StatusCode >= 400 {
		statusErr := fmt.Errorf("received non-success status: %s", resp.Status)
		if closeErr != nil {
			return fmt.Errorf("%w; failed to close response body: %v", statusErr, closeErr)
		}
		return statusErr
	}
	if closeErr != nil {
		return errors.Wrap(closeErr, "failed to close response body")
	}
	return nil
}

// splitMessage splits msg into chunks no longer than limit, breaking on newlines
// where possible so multi-line messages aren't cut mid-line.
func splitMessage(msg string, limit int) []string {
	if utf8.RuneCountInString(msg) <= limit {
		return []string{msg}
	}

	var chunks []string
	var current strings.Builder
	currentLen := 0

	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		runes := []rune(line)
		for len(runes) > limit {
			if currentLen > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
				currentLen = 0
			}
			chunks = append(chunks, string(runes[:limit]))
			runes = runes[limit:]
		}

		if currentLen > 0 && currentLen+1+len(runes) > limit {
			chunks = append(chunks, current.String())
			current.Reset()
			currentLen = 0
		}

		if currentLen > 0 {
			current.WriteByte('\n')
			currentLen++
		}
		current.WriteString(string(runes))
		currentLen += len(runes)
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks
}
