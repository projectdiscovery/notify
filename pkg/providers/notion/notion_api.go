package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http"
	"strings"
)

const notionApiUrl string = "https://api.notion.com/v1/"

func (options *Options) CreatePage(page Page) error {
	url := strings.Join([]string{notionApiUrl, "pages/"}, "")
	fmt.Println("Options:", options)
	apiKey := options.NotionAPIKey
	fmt.Println("API Key:", apiKey)
	requestBody := map[string]interface{}{
		"parent": map[string]string{
			"database_id": page.ParentId,
		},
		"properties": page.Properties,
	}
	if page.Children != nil {
		requestBody["children"] = page.Children
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshaling request body:", err)
		return err
	}
	fmt.Println("JSON Body:", string(jsonBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Notion-Version", "2022-02-22")
	req.Header.Add("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}

	fmt.Println(string(body))
	if res.StatusCode != 200 {
		fmt.Println("Error creating page:", res.Status)
		return fmt.Errorf(res.Status)
	}
	return nil
}

// func (options *Options) CreateNormalText(text string) error {
// 	var title = options.NotionDatabaseProperty
// 	newPage := Page{
// 		ParentId: options.NotionDatabaseId,
// 		Properties: map[string]interface{}{
// 			"Name": map[string]interface{}{
// 				title: []map[string]interface{}{
// 					{
// 						"text": map[string]string{
// 							"content": text,
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	// return options.CreatePage(newPage)
// 	return CreatePage(newPage)
// }

// func (options *Options) CreateBulkText(tite string, text string) error {
// 	var title = GetNotionDatabasePropertyTitle()
// 	newPage := Page{
// 		ParentId: GetNotionDatabaseId(),
// 		Properties: map[string]interface{}{
// 			"Name": map[string]interface{}{
// 				title: []map[string]interface{}{
// 					{
// 						"text": map[string]string{
// 							"content": tite,
// 						},
// 					},
// 				},
// 			},
// 		},
// 		Children: []map[string]interface{}{
// 			{
// 				"type": "code",
// 				"code": map[string]interface{}{
// 					"rich_text": []map[string]interface{}{
// 						{
// 							"type": "text",
// 							"text": map[string]string{
// 								"content": text,
// 							},
// 						},
// 					},
// 					"language": "javascript",
// 				},
// 			},
// 		},
// 	}
// 	return CreatePage(newPage)
// }
