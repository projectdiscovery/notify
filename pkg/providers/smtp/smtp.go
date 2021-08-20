package smtp

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"go.uber.org/multierr"
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
	SMTPFormat  string   `yaml:"smtp_format,omitempty"`
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

func (p *Provider) Send(message, CliFormat string) error {
	var SmtpErr error
	for _, pr := range p.SMTP {
		msg := utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.SMTPFormat))

		url := fmt.Sprintf("smtp://%s:%s@%s/?fromAddress=%s&toAddresses=%s", pr.Username, pr.Password, pr.Server, pr.FromAddress, strings.Join(pr.SMTPCC, ","))
		err := shoutrrr.Send(url, msg)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to send smtp notification for id: %s ", pr.ID))
			SmtpErr = multierr.Append(SmtpErr, err)
		}
		gologger.Verbose().Msgf("smtp notification sent for id: %s", pr.ID)
	}
	return SmtpErr
}
