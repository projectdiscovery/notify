package gotify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

type Provider struct {
	Gotify  []*Options `yaml:"gotify,omitempty"`
	counter int
}

type Options struct {
	ID               string `yaml:"id,omitempty"`
	GotifyHost       string `yaml:"gotify_host,omitempty"`
	GotifyPort       string `yaml:"gotify_port,omitempty"`
	GotifyToken      string `yaml:"gotify_token,omitempty"`
	GotifyFormat     string `yaml:"gotify_format,omitempty"`
	GotifyDisableTLS bool   `yaml:"gotify_disabletls,omitempty"`
	GotifyTitle      string `yaml:"gotify_title,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Gotify = append(provider.Gotify, o)
		}
	}

	provider.counter = 0

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var GotifyErr error
	p.counter++

	for _, pr := range p.Gotify {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.GotifyFormat), p.counter)
		err := pr.send(msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send gotify notification for id: %s ", pr.ID))
			GotifyErr = multierr.Append(GotifyErr, err)
			continue
		}
		gologger.Verbose().Msgf("gotify notification sent for id: %s", pr.ID)
	}

	return GotifyErr
}

// defaultTitle is the title shoutrrr used, kept so existing setups see no change.
const defaultTitle = "Shoutrrr notification"

type messageRequest struct {
	Message  string `json:"message"`
	Title    string `json:"title"`
	Priority int    `json:"priority"`
}

type errorResponse struct {
	Error       string `json:"error"`
	Description string `json:"errorDescription"`
}

// send talks to the Gotify API directly because shoutrrr only accepts the
// legacy 15-char tokens and rejects the "gtfya." tokens issued by Gotify 3.
func (o *Options) send(msg string) error {
	title := o.GotifyTitle
	if title == "" {
		title = defaultTitle
	}
	body, err := json.Marshal(messageRequest{Message: msg, Title: title})
	if err != nil {
		return err
	}

	scheme := "https"
	if o.GotifyDisableTLS {
		scheme = "http"
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s://%s:%s/message", scheme, o.GotifyHost, o.GotifyPort), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// A header keeps the token out of URLs, which end up in proxy and server logs.
	req.Header.Set("X-Gotify-Key", o.GotifyToken)

	client := &http.Client{
		Timeout: 10 * time.Second,
		// Gotify's default config redirects HTTP to HTTPS, so with TLS disabled the
		// redirect target is usually unverifiable; shoutrrr skipped verification too.
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: o.GotifyDisableTLS}},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var e errorResponse
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("gotify responded with %s: %s %s", resp.Status, e.Error, e.Description)
	}
	return nil
}
