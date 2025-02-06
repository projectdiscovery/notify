package notion

import (
	"fmt"
	"time"

	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

type Provider struct {
	Notion  []*Options `yaml:"notion,omitempty"`
	counter int
}

type Options struct {
	ID                     string `yaml:"id,omitempty"`
	NotionAPIKey         string `yaml:"notion_api_key,omitempty"`
	NotionDatabaseId       string `yaml:"notion_database_id,omitempty"`
	NotionInPageTitle string `yaml:"notion_in_page_title,omitempty"` // optional
}

func New(options []*Options, ids []string) (*Provider, error) {
	provider := &Provider{}

	for _, o := range options {
		if len(ids) == 0 || sliceutil.Contains(ids, o.ID) {
			provider.Notion = append(provider.Notion, o)
		}
	}

	provider.counter = 0

	return provider, nil
}

func (p *Provider) Send(message, CliFormat string) error {
	var errs []error
	for _, pr := range p.Notion {

		newPage := func() Page {
			if pr.NotionInPageTitle != "" {
				titleWithTimestamp := fmt.Sprintf("%s %s", pr.NotionInPageTitle, time.Now().Format("2006-01-02 15:04:05 MST"))
				return pr.CreateInpageText(titleWithTimestamp, message)
			}
			return pr.CreateNormalText(message)
			
		}()

		if success := pr.CreatePage(newPage); success != nil {
			err := fmt.Errorf("failed to send notion notification for id: %s", pr.ID)
			errs = append(errs, err)
			continue
		}

		gologger.Verbose().Msgf("Notion notification sent for id: %s", pr.ID)
	}
	return multierr.Combine(errs...)
}
