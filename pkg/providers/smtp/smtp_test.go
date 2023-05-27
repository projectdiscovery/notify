package smtp

import (
	"net/url"
	"testing"
)

// TestBuildUrl checks the output of buildUrl is valid and parsable by url.Parse
func TestBuildUrl(t *testing.T) {
	options := &Options{
		Server:      "mail.example.com",
		Username:    "test@example.com",
		Password:    "password",
		FromAddress: "from@email.com",
		SMTPCC:      []string{"to@email.com"},
		Subject:     "Email subject",
	}
	t.Run("with provider config example", func(t *testing.T) {
		u := buildUrl(options)
		_, err := url.Parse(u)
		if err != nil {
			t.Errorf("Failed to parse url: %s", err)
		}
	})
	t.Run("with octohorpe in password", func(t *testing.T) {
		options.Password = "passwordwith#inside"
		u := buildUrl(options)
		_, err := url.Parse(u)
		if err != nil {
			t.Errorf("Failed to parse url: %s", err)
		}
	})
}
