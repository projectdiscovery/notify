package runner

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/fileutil"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/providers"
	"github.com/projectdiscovery/notify/pkg/types"
	"github.com/projectdiscovery/notify/pkg/utils"
	"gopkg.in/yaml.v2"
)

// Runner contains the internal logic of the program
type Runner struct {
	options   *types.Options
	providers *providers.Client
}

// NewRunner instance
func NewRunner(options *types.Options) (*Runner, error) {
	var providerOptions providers.ProviderOptions

	if options.ProviderConfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		options.ProviderConfig = path.Join(home, types.DefaultProviderConfigLocation)
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

	// Discard all internal logs
	shoutrrr.SetLogger(log.New(ioutil.Discard, "", 0))

	prClient, err := providers.New(&providerOptions, options)
	if err != nil {
		return nil, err
	}

	return &Runner{options: options, providers: prClient}, nil
}

// Run polling and notification
func (r *Runner) Run() error {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	if r.options.Proxy != "" {
		proxyurl, err := url.Parse(r.options.Proxy)
		if err != nil || proxyurl == nil {
			gologger.Warning().Msgf("supplied proxy '%s' is not valid", r.options.Proxy)
		} else {
			defaultTransport = &http.Transport{
				Proxy:             http.ProxyURL(proxyurl),
				ForceAttemptHTTP2: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
		}
	}

	if r.options.RateLimit > 0 {
		http.DefaultClient.Transport = utils.NewThrottledTransport(time.Second, r.options.RateLimit, defaultTransport)
	}

	var inFile *os.File
	var err error

	switch {
	case r.options.Data != "":
		inFile, err = os.Open(r.options.Data)
		if err != nil {
			gologger.Fatal().Msgf("%s\n", err)
		}
	case fileutil.HasStdin():
		inFile = os.Stdin
	default:
		return errors.New("notify works with stdin or file using -data flag")
	}

	br := bufio.NewScanner(inFile)
	if r.options.Bulk {
		br.Split(bulkSplitter(r.options.CharLimit))
	}
	for br.Scan() {
		msg := br.Text()
		//nolint:errcheck
		r.sendMessage(msg)
	}
	return nil
}

// Return a SplitFunc that tries to split on newlines while giving as many bytes that are <= charLimit each time
func bulkSplitter(charLimit int) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// We need to be prepared to collect tokens because ScanLines trims trailing CR's for us
		tokens := make([]byte, 0, charLimit)

		advance, token, err = bufio.ScanLines(data, atEOF)

		if err != nil || token == nil {
			// Didn't get a line
			return
		}

		if len(token) >= charLimit {
			// Got too much. Give charLimit bytes and finish the rest of the line next time.
			advance = charLimit
			token = token[:charLimit]
			return
		}

		tokens = append(tokens, token...)

		// Keep getting lines until we exceed charLimit
		for {
			newAdvance, token, err := bufio.ScanLines(data[advance:], atEOF)

			if err != nil || token == nil {
				// Failed to get a line
				break
			}

			if len(tokens)+len(token) > charLimit {
				// Too much. Give what we had.
				return advance, tokens, err
			}

			advance += newAdvance
			tokens = append(tokens, '\n')
			tokens = append(tokens, token...)
		}

		// Stopped getting lines but still hungry for bytes

		// Are we done?
		if atEOF {
			return advance, tokens, nil
		}

		// Need more data
		return 0, nil, nil
	}
}

func (r *Runner) sendMessage(msg string) error {
	if len(msg) > 0 {
		fmt.Println(msg)
		err := r.providers.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close the runner instance
func (r *Runner) Close() {}
