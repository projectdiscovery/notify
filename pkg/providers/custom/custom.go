package custom

import (
	"bytes"
	"encoding/json"
	"strings"
	"fmt"
	"net/http"
	"text/template"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"github.com/Masterminds/sprig"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/notify/pkg/utils"
	"github.com/projectdiscovery/notify/pkg/utils/httpreq"
	"github.com/projectdiscovery/sliceutil"
)

type Provider struct {
	Custom []*Options `yaml:"custom,omitempty"`
}

type Options struct {
	ID               string            `yaml:"id,omitempty"`
	CustomWebhookURL string            `yaml:"custom_webhook_url,omitempty"`
	CustomMethod     string            `yaml:"custom_method,omitempty"`
	CustomHeaders    map[string]string `yaml:"custom_headers,omitempty"`
	CustomFormat     string            `yaml:"custom_format,omitempty"`
	CustomSprig      string            `yaml:"custom_sprig,omitempty"`
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Custom = append(provider.Custom, o)
		}
	}

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
    var CustomErr error

    for _, pr := range p.Custom {
        var msg string
        if pr.CustomSprig != "" {
            // Convert a string to JSON
            var data map[string]interface{}
            if err := json.Unmarshal([]byte(message), &data); err != nil {
                return errors.Wrap(err, fmt.Sprintf("failed to unmarshal message to JSON for id: %s ", pr.ID))
            }

            funcMap := sprig.TxtFuncMap()
            // Add custom functions if needed using funcMap["funcName"] = func
            tmpl, err := template.New("sprig").Funcs(funcMap).Parse(pr.CustomSprig)
            if err != nil {
                err = errors.Wrap(err, fmt.Sprintf("failed to parse custom sprig template for id: %s ", pr.ID))
                CustomErr = multierr.Append(CustomErr, err)
                continue
            }
            var buf bytes.Buffer
            err = tmpl.Execute(&buf, data)
            if err != nil {
                err = errors.Wrap(err, fmt.Sprintf("failed to execute custom sprig template for id: %s ", pr.ID))
                CustomErr = multierr.Append(CustomErr, err)
                continue
            }
            msg = buf.String()
        } else if strings.Contains(pr.CustomFormat, "{{dataJsonString}}") {
            // Escape the message to a JSON string
            b, err := json.Marshal(message)
            if err != nil {
                return errors.Wrap(err, fmt.Sprintf("failed to escape message to JSON for id: %s ", pr.ID))
            }
            dataJsonString := string(b)

            // Replace the "{{dataJsonString}}" substring in the custom format with the escaped JSON string
            msg = strings.ReplaceAll(pr.CustomFormat, "{{dataJsonString}}", dataJsonString)
        } else {
            // Otherwise, use the original message
            msg = utils.FormatMessage(message, utils.SelectFormat(CliFormat, pr.CustomFormat))
        }

        body := bytes.NewBufferString(msg)
        gologger.Verbose().Msgf("custom body sent: %s", msg)

        r, err := http.NewRequest(pr.CustomMethod, pr.CustomWebhookURL, body)
        if err != nil {
            err = errors.Wrap(err, fmt.Sprintf("failed to send custom notification for id: %s ", pr.ID))
            CustomErr = multierr.Append(CustomErr, err)
            continue
        }

        for k, v := range pr.CustomHeaders {
            r.Header.Set(k, v)
        }

        _, err = httpreq.NewClient().Do(r)
        if err != nil {
            err = errors.Wrap(err, fmt.Sprintf("failed to send custom notification for id: %s ", pr.ID))
            CustomErr = multierr.Append(CustomErr, err)
            continue
        }
        gologger.Verbose().Msgf("custom notification sent for id: %s", pr.ID)
    }
    return CustomErr
}


