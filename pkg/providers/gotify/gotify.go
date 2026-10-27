package gotify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
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

type gotifyPayload struct {
	Message  string `json:"message"`
	Title    string `json:"title,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// gotifyCheckRedirect follows redirects on the configured host, including
// HTTP to HTTPS, and stops before the Gotify token is sent to another host.
func gotifyCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) == 0 || req.URL == nil || via[0].URL == nil {
		return nil
	}
	if !strings.EqualFold(req.URL.Hostname(), via[0].URL.Hostname()) {
		return http.ErrUseLastResponse
	}
	return nil
}

var gotifyHTTPClient = &http.Client{
	Timeout:       10 * time.Second,
	CheckRedirect: gotifyCheckRedirect,
}

func isLegacyGotifyToken(token string) bool {
	if len(token) != 15 || token[0] != 'A' {
		return false
	}
	const tokenChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_"
	for _, c := range token {
		if !strings.ContainsRune(tokenChars, c) {
			return false
		}
	}
	return true
}

func sendGotifyDirect(pr *Options, msg string) error {
	scheme := "https"
	if pr.GotifyDisableTLS {
		scheme = "http"
	}
	hostPort := pr.GotifyHost
	if pr.GotifyPort != "" {
		hostPort = net.JoinHostPort(pr.GotifyHost, pr.GotifyPort)
	}
	endpoint := fmt.Sprintf("%s://%s/message", scheme, hostPort)

	payload := gotifyPayload{
		Message: msg,
		Title:   pr.GotifyTitle,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal gotify payload")
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "failed to create gotify request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", pr.GotifyToken)

	client := gotifyHTTPClient
	if pr.GotifyDisableTLS {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout:       10 * time.Second,
			CheckRedirect: gotifyCheckRedirect,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send gotify request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gotify server returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	// Drain the success body so the connection can be reused.
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var GotifyErr error
	p.counter++

	for _, pr := range p.Gotify {
		params := url.Values{}
		if pr.GotifyTitle != "" {
			params.Add("title", pr.GotifyTitle)
		}
		if pr.GotifyDisableTLS {
			params.Add("disabletls", "true")
		}
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.GotifyFormat), p.counter)

		var err error
		if isLegacyGotifyToken(pr.GotifyToken) {
			url := fmt.Sprintf("gotify://%s:%s/%s", pr.GotifyHost, pr.GotifyPort, pr.GotifyToken)
			if len(params) > 0 {
				url += "?" + params.Encode()
			}
			err = shoutrrr.Send(url, msg)
			if err != nil && strings.Contains(err.Error(), "invalid gotify token") {
				err = sendGotifyDirect(pr, msg)
			}
		} else {
			// Gotify 3 tokens (e.g. gtfya.*) are rejected by shoutrrr v0.8.0 which expects legacy 15-char tokens starting with 'A'.
			// Direct REST API POST sends directly to the Gotify server.
			err = sendGotifyDirect(pr, msg)
		}

		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send gotify notification for id: %s ", pr.ID))
			GotifyErr = multierr.Append(GotifyErr, err)
			continue
		}
		gologger.Verbose().Msgf("gotify notification sent for id: %s", pr.ID)
	}

	return GotifyErr
}
