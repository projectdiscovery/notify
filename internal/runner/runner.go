package runner

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/providers"
	"github.com/projectdiscovery/notify/pkg/types"
	"github.com/projectdiscovery/notify/pkg/utils"
	fileutil "github.com/projectdiscovery/utils/file"
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
		options.ProviderConfig = filepath.Join(home, types.DefaultProviderConfigLocation)
		gologger.Print().Msgf("Using default provider config: %s\n", options.ProviderConfig)
	}

	reader, err := readProviderConfig(options.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if parseErr := yaml.NewDecoder(reader).Decode(&providerOptions); parseErr != nil {
		return nil, errors.Wrap(parseErr, "could not parse provider config file")
	}

	shoutrrr.SetLogger(log.New(io.Discard, "", 0))

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
	var splitter bufio.SplitFunc

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

	if r.options.CharLimit > bufio.MaxScanTokenSize {
		// Satisfy the condition of our splitters, which is that charLimit is <= the size of the bufio.Scanner buffer
		buffer := make([]byte, 0, r.options.CharLimit)
		br.Buffer(buffer, r.options.CharLimit)
	}

	if r.options.Bulk {
		splitter, err = bulkSplitter(r.options.CharLimit)
	} else {
		splitter, err = lineLengthSplitter(r.options.CharLimit)
	}

	if err != nil {
		return err
	}

	br.Split(splitter)

	for br.Scan() {
		msg := br.Text()
		//nolint:errcheck
		r.sendMessage(msg)
	}
	return br.Err()
}

func (r *Runner) sendMessage(msg string) error {
	if len(msg) > 0 {
		if r.options.Delay > 0 {
			time.Sleep(time.Duration(r.options.Delay) * time.Second)
		}
		gologger.Silent().Msgf("%s\n", msg)
		err := r.providers.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close the runner instance
func (r *Runner) Close() {}

func readProviderConfig(filepath string) (io.Reader, error) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Create a string builder to accumulate the modified data
	var sb strings.Builder

	// Iterate over each line and do variable substitution
	for scanner.Scan() {
		line := scanner.Text()
		newLine := substituteEnvVars(line)
		sb.WriteString(newLine)
		sb.WriteString("\n")
	}
	// Check for errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return strings.NewReader(sb.String()), nil
}

func substituteEnvVars(line string) string {
	for _, word := range strings.Fields(line) {
		word = strings.Trim(word, `"`)
		if strings.HasPrefix(word, "$") {
			key := strings.TrimPrefix(word, "$")
			substituteEnv := os.Getenv(key)
			if substituteEnv != "" {
				line = strings.Replace(line, word, substituteEnv, 1)
			}
		}
	}
	return line
}
