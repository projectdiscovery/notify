package runner

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/projectdiscovery/collaborator"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify"
)

const (
	defaultHTTPMessage = "The collaborator server received an {{protocol}} request from {{from}} at {{time}}:\n```\n{{request}}\n{{response}}```"
	defaultDNSMessage  = "The collaborator server received a DNS lookup of type {{type}} for the domain name {{domain}} from {{from}} at {{time}}:\n```{{request}}```"
	defaultSMTPMessage = "The collaborator server received an SMTP connection from IP address {{from}} at {{time}}\n\nThe email details were:\n\nFrom:\n{{sender}}\n\nTo:\n{{recipients}}\n\nMessage:\n{{message}}\n\nSMTP Conversation:\n{{conversation}}"
	defaultCLIMessage  = "{{data}}"
)

// Runner contains the internal logic of the program
type Runner struct {
	options    *Options
	burpcollab *collaborator.BurpCollaborator
	notifier   *notify.Notify
}

// NewRunner instance
func NewRunner(options *Options) (*Runner, error) {
	burpcollab := collaborator.NewBurpCollaborator()

	notifier, err := notify.NewWithOptions(&notify.Options{
		SlackWebHookURL:         options.SlackWebHookURL,
		SlackUsername:           options.SlackUsername,
		SlackChannel:            options.SlackChannel,
		Slack:                   options.Slack,
		DiscordWebHookURL:       options.DiscordWebHookURL,
		DiscordWebHookUsername:  options.DiscordWebHookUsername,
		DiscordWebHookAvatarURL: options.DiscordWebHookAvatarURL,
		Discord:                 options.Discord,
		TelegramAPIKey:          options.TelegramAPIKey,
		TelegramChatID:          options.TelegramChatID,
		Telegram:                options.Telegram,
		SMTP:                    options.SMTP,
		SMTPProviders:           options.SMTPProviders,
		SMTPCC:                  options.SMTPCC,
	})
	if err != nil {
		return nil, err
	}

	return &Runner{options: options, burpcollab: burpcollab, notifier: notifier}, nil
}

// Run polling and notification
func (r *Runner) Run() error {
	// If stdin is present pass everything to webhooks and exit
	if hasStdin() {
		br := bufio.NewScanner(os.Stdin)
		for br.Scan() {
			msg := br.Text()
			rr := strings.NewReplacer(
				"{{data}}", msg,
			)
			msg = rr.Replace(r.options.CLIMessage)
			gologger.Printf(msg)
			//nolint:errcheck // silent fail
			r.notifier.SendNotification(msg)
		}
		os.Exit(0)
	}

	// otherwise works as long term collaborator poll and notify via webhook
	// If BIID passed via cli
	if r.options.BIID != "" {
		gologger.Printf("Using BIID: %s", r.options.BIID)
		r.burpcollab.AddBIID(r.options.BIID)
	}

	if r.options.BIID == "" {
		return fmt.Errorf("BIID not specified or not found")
	}

	err := r.burpcollab.Poll()
	if err != nil {
		return err
	}

	pollTime := time.Duration(r.options.Interval) * time.Second
	for {
		time.Sleep(pollTime)
		//nolint:errcheck // silent fail
		r.burpcollab.Poll()

		for _, httpresp := range r.burpcollab.RespBuffer {
			for i := range httpresp.Responses {
				resp := httpresp.Responses[i]
				var at int64
				at, _ = strconv.ParseInt(resp.Time, 10, 64)
				atTime := time.Unix(0, at*int64(time.Millisecond))
				switch resp.Protocol {
				case "http", "https":
					rr := strings.NewReplacer(
						"{{protocol}}", strings.ToUpper(resp.Protocol),
						"{{from}}", resp.Client,
						"{{time}}", atTime.String(),
						"{{request}}", resp.Data.RequestDecoded,
						"{{response}}", resp.Data.ResponseDecoded,
					)

					msg := rr.Replace(r.options.HTTPMessage)
					gologger.Printf(msg)

					//nolint:errcheck // silent fail
					r.notifier.SendNotification(msg)
				case "dns":
					rr := strings.NewReplacer(
						"{{type}}", resp.Data.RequestType,
						"{{domain}} ", resp.Data.SubDomain,
						"{{from}}", resp.Client,
						"{{time}}", atTime.String(),
						"{{request}}", resp.Data.RawRequestDecoded,
					)
					msg := rr.Replace(r.options.DNSMessage)
					gologger.Printf(msg)

					//nolint:errcheck // silent fail
					r.notifier.SendNotification(msg)
				case "smtp":
					rr := strings.NewReplacer(
						"{{from}}", resp.Client,
						"{{time}}", atTime.String(),
						"{{sender}}", resp.Data.SenderDecoded,
						"{{recipients}}", strings.Join(resp.Data.RecipientsDecoded, ","),
						"{{message}}", resp.Data.MessageDecoded,
						"{{conversation}}", resp.Data.ConversationDecoded,
					)
					msg := rr.Replace(r.options.SMTPMessage)
					gologger.Printf(msg)

					//nolint:errcheck // silent fail
					r.notifier.SendNotification(msg)
				}
			}
		}

		r.burpcollab.Empty()
	}
}

// Close the runner instance
func (r *Runner) Close() {
	r.burpcollab.Empty()
}
