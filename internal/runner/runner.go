package runner

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/projectdiscovery/collaborator"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/engine"
	"github.com/projectdiscovery/notify/pkg/providers"
	"github.com/projectdiscovery/notify/pkg/types"
	"gopkg.in/yaml.v2"
)

// Runner contains the internal logic of the program
type Runner struct {
	options    *types.Options
	burpcollab *collaborator.BurpCollaborator
	notifier   *engine.Notify
	providers  *providers.Client
}

// NewRunner instance
func NewRunner(options *types.Options) (*Runner, error) {
	burpcollab := collaborator.NewBurpCollaborator()

	notifier, err := engine.NewWithOptions(options)
	if err != nil {
		return nil, err
	}
	var providerOptions providers.Options

	if options.ProviderConfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		options.ProviderConfig = path.Join(home, "/.config/notify/provider-config.yaml")
		gologger.Print().Msgf("Using default provider config: %s\n", options.ProviderConfig)
	}

	file, err := os.Open(options.ProviderConfig)
	if err != nil {
		return nil, errors.Wrap(err, "could not open provider config file")
	}

	if parseErr := yaml.NewDecoder(file).Decode(&providerOptions); parseErr != nil {
		file.Close()
		return nil, errors.Wrap(parseErr, "could not parse provider config file")
	}

	file.Close()

	prClient, err := providers.New(&providerOptions, options.Providers, options.Profiles)
	if err != nil {
		return nil, err
	}

	return &Runner{options: options, burpcollab: burpcollab, notifier: notifier, providers: prClient}, nil
}

// Run polling and notification
func (r *Runner) Run() error {
	// If stdin is present pass everything to webhooks and exit
	if hasStdin() || r.options.Data != "" {
		var br *bufio.Scanner

		switch {
		case hasStdin():
			br = bufio.NewScanner(os.Stdin)

		case r.options.Data != "":
			inFile, err := os.Open(r.options.Data)
			if err != nil {
				gologger.Fatal().Msgf("%s\n", err)
			}
			br = bufio.NewScanner(inFile)

		}

		for br.Scan() {
			msg := br.Text()
			if msg == "" {
				continue
			}
			rr := strings.NewReplacer(
				"{{data}}", msg,
			)
			msg = rr.Replace(r.options.CLIMessage)
			gologger.Print().Msgf(msg)
			//nolint:errcheck // silent fail
			r.providers.Send(msg)
		}
		os.Exit(0)
	}

	// otherwise works as long term collaborator poll and notify via webhook
	// If BIID passed via cli
	if r.options.BIID != "" {
		gologger.Print().Msgf("Using BIID: %s", r.options.BIID)
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
					gologger.Print().Msgf(msg)

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
					gologger.Print().Msgf(msg)

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
					gologger.Print().Msgf(msg)

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
