package teams

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/projectdiscovery/notify/pkg/utils"
)

type Provider struct {
	Teams []*Options `yaml:"teams,omitempty"`
}

type Options struct {
	Profile         string `yaml:"profile,omitempty"`
	TeamsWebHookURL string `yaml:"teams_webhook_url,omitempty"`
}

func New(options []*Options, profiles []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(profiles) == 0 || utils.Contains(profiles, o.Profile) {
			provider.Teams = append(provider.Teams, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message string) error {

	for _, pr := range p.Teams {
		teamsTokens := strings.TrimPrefix(pr.TeamsWebHookURL, "https://outlook.office.com/webhook/")
		teamsTokens = strings.ReplaceAll(teamsTokens, "IncomingWebhook/", "")
		url := fmt.Sprintf("teams://%s", teamsTokens)
		err := shoutrrr.Send(url, message)
		if err != nil {
			return err
		}
	}
	return nil
}
