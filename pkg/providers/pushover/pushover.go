package pushover

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/projectdiscovery/notify/pkg/utils"
)

type Provider struct {
	Pushover []*Options `yaml:"pushover,omitempty"`
}

type Options struct {
	Profile          string   `yaml:"profile,omitempty"`
	PushoverApiToken string   `yaml:"pushover_api_token,omitempty"`
	UserKey          string   `yaml:"pushover_user_key,omitempty"`
	PushoverDevices  []string `yaml:"pushover_devices,omitempty"`
}

func New(options []*Options, profiles []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(profiles) == 0 || utils.Contains(profiles, o.Profile) {
			provider.Pushover = append(provider.Pushover, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message string) error {

	for _, pr := range p.Pushover {
		url := fmt.Sprintf("pushover://shoutrrr:%s@%s/?devices=%s", pr.PushoverApiToken, pr.UserKey, strings.Join(pr.PushoverDevices, ","))
		err := shoutrrr.Send(url, message)
		if err != nil {
			return err
		}
	}
	return nil
}
