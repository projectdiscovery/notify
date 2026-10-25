package teams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

type Provider struct {
	Teams   []*Options `yaml:"teams,omitempty"`
	counter int
}

type Options struct {
	ID              string `yaml:"id,omitempty"`
	TeamsWebHookURL string `yaml:"teams_webhook_url,omitempty"`
	TeamsFormat     string `yaml:"teams_format,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Teams = append(provider.Teams, o)
		}
	}

	provider.counter = 0

	return provider, nil
}

// teamsHTTPClient posts Power Automate workflow messages. Office 365 connector
// webhooks still go through shoutrrr.
var teamsHTTPClient = &http.Client{Timeout: 15 * time.Second}

func (p *Provider) Send(message, CliFormat string) error {
	var teamsErr error
	p.counter++
	for _, pr := range p.Teams {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.TeamsFormat), p.counter)
		if err := sendTeams(pr, msg); err != nil {
			teamsErr = multierr.Append(teamsErr, err)
			continue
		}
		gologger.Verbose().Msgf("teams notification sent for id: %s", pr.ID)
	}
	return teamsErr
}

func sendTeams(pr *Options, msg string) error {
	kind, err := teamsWebhookKind(pr.TeamsWebHookURL)
	if err != nil {
		return fmt.Errorf("teams: %w for id: %s", err, pr.ID)
	}
	switch kind {
	case teamsWebhookConnector:
		return sendConnector(pr, msg)
	case teamsWebhookWorkflow:
		return sendWorkflow(pr, msg)
	default:
		return fmt.Errorf("teams: unsupported webhook url for id: %s", pr.ID)
	}
}

const (
	teamsWebhookConnector = "webhookb2"
	teamsWebhookWorkflow  = "workflows"
)

func teamsWebhookKind(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.Path == "" {
		return "", fmt.Errorf("invalid webhook url")
	}
	switch {
	case strings.Contains(parsed.Path, "/webhookb2/"):
		return teamsWebhookConnector, nil
	case strings.Contains(parsed.Path, "/workflows/"):
		return teamsWebhookWorkflow, nil
	default:
		return "", fmt.Errorf("unsupported webhook url")
	}
}

func sendConnector(pr *Options, msg string) error {
	webhookParts := strings.Split(pr.TeamsWebHookURL, "/webhookb2/")
	if len(webhookParts) != 2 || webhookParts[1] == "" {
		return fmt.Errorf("teams: invalid webhook url for id: %s", pr.ID)
	}
	teamsHost := strings.TrimPrefix(webhookParts[0], "https://")
	teamsTokens := strings.ReplaceAll(webhookParts[1], "IncomingWebhook/", "")
	webhookURL := fmt.Sprintf("teams://%s?host=%s", teamsTokens, teamsHost)
	if err := shoutrrr.Send(webhookURL, msg); err != nil {
		return errors.Wrapf(err, "failed to send webhook teams notification for id: %s", pr.ID)
	}
	return nil
}

func sendWorkflow(pr *Options, msg string) error {
	payload, err := workflowPayload(msg)
	if err != nil {
		return errors.Wrapf(err, "failed to encode workflow teams notification for id: %s", pr.ID)
	}
	resp, err := teamsHTTPClient.Post(pr.TeamsWebHookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return errors.Wrapf(err, "failed to send workflow teams notification for id: %s", pr.ID)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusAccepted {
		return errors.Errorf("failed to send workflow teams notification for id: %s, got status code: %d", pr.ID, resp.StatusCode)
	}
	return nil
}

func workflowPayload(msg string) ([]byte, error) {
	text := strings.ReplaceAll(html.EscapeString(msg), "\n", "<br>")
	return json.Marshal(struct {
		Text string `json:"text"`
	}{Text: text})
}
