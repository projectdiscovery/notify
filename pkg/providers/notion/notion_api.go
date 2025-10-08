package notion

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/projectdiscovery/notify/pkg/utils/httpreq"
)

const notionApiUrl string = "https://api.notion.com/v1/"

func (options *Options) CreatePage(page Page) error {
	url := strings.Join([]string{notionApiUrl, "pages/"}, "")
	apiKey := options.NotionAPIKey
	body := APIRequest{
		Parent:     map[string]interface{}{"database_id": page.ParentId},
		Properties: page.Properties,
		Children:   page.Children,
	}

	headers := http.Header{
		"Content-Type":  {"application/json"},
		"Notion-Version": {"2022-02-22"},
		"Authorization": {fmt.Sprintf("Bearer %s", apiKey)},
	}

	var response *APIResponse

	err := httpreq.NewClient().Post(url, &body, headers, &response)
	if err != nil {
		return err
	}
	if response.Object != "page" {
		return fmt.Errorf("error while sending notion message: %s ", response.Message)
	}

	return nil
}

func (options *Options) CreateNormalText(text string) Page {
	newPage := Page{
		ParentId: options.NotionDatabaseId,
		Properties: map[string]interface{}{
			"Name": map[string]interface{}{
				"title": []map[string]interface{}{
					{
						"text": map[string]string{
							"content": text,
						},
					},
				},
			},
		},
	}

	return newPage
}

func (options *Options) CreateInpageText(tite string, text string) Page {
	newPage := Page{
		ParentId: options.NotionDatabaseId,
		Properties: map[string]interface{}{
			"Name": map[string]interface{}{
				"title": []map[string]interface{}{
					{
						"text": map[string]string{
							"content": tite,
						},
					},
				},
			},
		},
		Children: []map[string]interface{}{
			{
				"type": "code",
				"code": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": text,
							},
						},
					},
					"language": "bash",
				},
			},
		},
	}
	return newPage
}
