package runner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/projectdiscovery/collaborator"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/engine"
	"github.com/projectdiscovery/notify/pkg/types"
)

// Runner contains the internal logic of the program
type Runner struct {
	options    *types.Options
	burpcollab *collaborator.BurpCollaborator
	notifier   *engine.Notify
}

// NewRunner instance
func NewRunner(options *types.Options) (*Runner, error) {
	burpcollab := collaborator.NewBurpCollaborator()

	notifier, err := engine.NewWithOptions(options)
	if err != nil {
		return nil, err
	}

	return &Runner{options: options, burpcollab: burpcollab, notifier: notifier}, nil
}

// Run polling and notification
func (r *Runner) Run() error {

	// If stdin/file input is present pass everything to webhooks and exit
	if hasStdin() || r.options.Data != "" {
		var inFile *os.File
		var err error

		switch {
		case hasStdin():
			inFile = os.Stdin

		case r.options.Data != "":
			inFile, err = os.Open(r.options.Data)
			if err != nil {
				gologger.Fatal().Msgf("%s\n", err)
			}
		}

		if r.options.StdinAll {
			fi, err := inFile.Stat()
			if err != nil {
				gologger.Fatal().Msgf("%s\n", err)
			}

			// LimitReader can be used to read large file content in smaller chunks (size of which is supported by platforms)
			// Although this might cause one issue:  messages won't necessarily get delivered in the same order that they are sent
			reader := io.LimitReader(inFile, fi.Size())
			msgB := make([]byte, fi.Size())
			for {
				n, err := reader.Read(msgB)
				if err != nil {
					if err == io.EOF {
						break
					}
					gologger.Fatal().Msgf("%s\n", err)
				}

				if n == 0 {
					break
				}

				msg := string(msgB)
				if err := r.sendMessage(msg); err != nil {
					gologger.Fatal().Msgf("%s\n", err)
				}
			}

			os.Exit(0)
		}

		br := bufio.NewScanner(inFile)
		for br.Scan() {
			msg := br.Text()
			//nolint:errcheck
			r.sendMessage(msg)
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

func (r *Runner) sendMessage(msg string) error {
	rr := strings.NewReplacer(
		"{{data}}", msg,
	)
	msg = rr.Replace(r.options.CLIMessage)
	if len(msg) > 0 {
		gologger.Print().Msgf(msg)
		err := r.notifier.SendNotification(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close the runner instance
func (r *Runner) Close() {
	r.burpcollab.Empty()
}
