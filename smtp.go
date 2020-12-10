package notify

import (
	"net/smtp"
	"time"

	"github.com/projectdiscovery/gologger"
)

// DefaultTelegraTimeout to conclude operations
const (
	DefaultSMTPTimeout = 10 * time.Second
)

type AuthenticationType int

const (
	PlainAuth = iota + 1
	CRAMMD5Auth
)

type SMTPProvider struct {
	Server             string `yaml:"smtp_server,omitempty"`
	Username           string `yaml:"smtp_username,omitempty"`
	Password           string `yaml:"smtp_password,omitempty"`
	AuthenticationType string `yaml:"smtp_authentication_type,omitempty"`
}

// TelegramClient handling webhooks
type SMTPClient struct {
	Providers []SMTPProvider
	CC        []string
	TimeOut   time.Duration
}

// SendInfo to telegram
func (sm *SMTPClient) SendInfo(message string) (err error) {
	// Create connection if not done already
	for _, provider := range sm.Providers {
		var auth smtp.Auth
		if provider.AuthenticationType == "basic" {
			auth = smtp.PlainAuth("", provider.Username, provider.Password, provider.Server)
		} else if provider.AuthenticationType == "crammd5" {
			auth = smtp.CRAMMD5Auth(provider.Username, provider.Password)
		} else if provider.AuthenticationType == "none" {
			auth = nil
		} else {
			continue
		}

		err := smtp.SendMail(provider.Server, auth, provider.Username, sm.CC, []byte(message))
		if err != nil {
			gologger.Errorf("%s\n", err)
			return err
		}
	}

	return nil
}
