package smtp

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/projectdiscovery/notify/pkg/utils"
)

type Provider struct {
	SMTP []*Options `yaml:"smtp,omitempty"`
}

type Options struct {
	ID          string   `yaml:"id,omitempty"`
	Server      string   `yaml:"smtp_server,omitempty"`
	Username    string   `yaml:"smtp_username,omitempty"`
	Password    string   `yaml:"smtp_password,omitempty"`
	FromAddress string   `yaml:"from_address,omitempty"`
	SMTPCC      []string `yaml:"smtp_cc,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || utils.Contains(ids, o.ID) {
			provider.SMTP = append(provider.SMTP, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message string) error {

	for _, pr := range p.SMTP {
		url := fmt.Sprintf("smtp://%s:%s@%s/?fromAddress=%s&toAddresses=%s", pr.Username, pr.Password, pr.Server, pr.FromAddress, strings.Join(pr.SMTPCC, ","))
		err := shoutrrr.Send(url, message)
		if err != nil {
			return err
		}
	}
	return nil
}
