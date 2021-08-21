package httpreq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: http.DefaultClient,
	}
}

func (c *Client) Get(url string, response interface{}) error {
	res, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	if err := jsoniter.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("error trying to unmarshal the response: %v", err)
	}
	return nil
}

func (c *Client) Post(url string, requestBody interface{}, headers http.Header, response interface{}) error {
	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error creating payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	for key, val := range headers {
		req.Header.Set(key, val[0])
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send payload: %v", err)
	}

	if err = jsoniter.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("error trying to unmarshal the response: %v", err)
	}
	return nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}
