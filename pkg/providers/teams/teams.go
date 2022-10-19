package teams

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/sliceutil"
)

type Provider struct {
	Teams []*Options `yaml:"teams,omitempty"`
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

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var TeamsErr error
	for _, pr := range p.Teams {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.TeamsFormat))
		webhookParts := strings.Split(pr.TeamsWebHookURL, "/webhookb2/")
		if len(webhookParts) != 2 {
			err := fmt.Errorf("teams: invalid webhook url for id: %s ", pr.ID)
			TeamsErr = multierr.Append(TeamsErr, err)
		}
		teamsHost := strings.TrimPrefix(webhookParts[0], "https://")
		teamsTokens := strings.ReplaceAll(webhookParts[1], "IncomingWebhook/", "")
		url := fmt.Sprintf("teams://%s?host=%s", teamsTokens, teamsHost)
		err := shoutrrr.Send(url, msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send teams notification for id: %s ", pr.ID))
			TeamsErr = multierr.Append(TeamsErr, err)
			continue
		}
		gologger.Verbose().Msgf("teams notification sent for id: %s", pr.ID)
	}
	return TeamsErr
}
