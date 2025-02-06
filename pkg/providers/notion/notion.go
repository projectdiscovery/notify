package notion

import (
	"fmt"

	"go.uber.org/multierr"

	"github.com/projectdiscovery/gologger"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

// Provider는 Notion 알림 전송을 위한 옵션 목록을 갖습니다.
type Provider struct {
	Notion  []*Options `yaml:"notion,omitempty"`
	counter int
}

// Options는 각 Notion 설정 항목을 정의합니다.
type Options struct {
	ID                     string `yaml:"id,omitempty"`
	NotionAPIKey         string `yaml:"notion_api_key,omitempty"`
	NotionDatabaseId       string `yaml:"notion_database_id,omitempty"`
	NotionTextPropertise string `yaml:"notion_text_propertise,omitempty"`
	NotionBulkText         string `yaml:"notion_bulk_text,omitempty"`
}

// New는 제공된 옵션과 id 목록을 기반으로 Provider 인스턴스를 생성합니다.
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

// Send는 메시지를 Notion에 전송합니다.
// 이 예제에서는 전송 로직으로 CreatenormalText 함수를 사용합니다.
func (p *Provider) Send(message, CliFormat string) error {
	var errs []error
	// 현재 CliFormat은 별도 포맷팅 없이 그대로 사용합니다.
	for _, pr := range p.Notion {
		// CreateNormalText 함수를 호출하여 Notion에 일반 텍스트 메시지를 전송합니다.



	var title = pr.NotionTextPropertise
	newPage := Page{
		ParentId: pr.NotionDatabaseId,
		Properties: map[string]interface{}{
			"Name": map[string]interface{}{
				title: []map[string]interface{}{
					{
						"text": map[string]string{
							"content": message,
						},
					},
				},
			},
		},
	}
	// return options.CreatePage(newPage)

		if success := pr.CreatePage(newPage); success != nil {
			err := fmt.Errorf("failed to send notion notification for id: %s", pr.ID)
			errs = append(errs, err)
			continue
		}

		gologger.Verbose().Msgf("Notion notification sent for id: %s", pr.ID)
	}
	return multierr.Combine(errs...)
}
