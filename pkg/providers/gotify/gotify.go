package gotify

import (
	"fmt"
	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	sliceutil "github.com/projectdiscovery/utils/slice"
	"go.uber.org/multierr"
)

type Provider struct {
	Gotify  []*Options `yaml:"gotify,omitempty"`
	counter int
}

type Options struct {
	ID           string `yaml:"id,omitempty"`
	GotifyHost   string `yaml:"gotify_host,omitempty"`
	GotifyPort   string `yaml:"gotify_port,omitempty"`
	GotifyToken  string `yaml:"gotify_token,omitempty"`
	GotifyFormat string `yaml:"gotify_format,omitempty"`
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
		url := fmt.Sprintf("gotify://%s:%s/%s", pr.GotifyHost, pr.GotifyPort, pr.GotifyToken)
		err := shoutrrr.Send(url, msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send gotify notification for id: %s ", pr.ID))
			GotifyErr = multierr.Append(GotifyErr, err)
			continue
		}
		gologger.Verbose().Msgf("gotify notification sent for id: %s", pr.ID)
	}

	return GotifyErr
}