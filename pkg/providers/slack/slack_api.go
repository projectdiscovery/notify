package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/projectdiscovery/retryablehttp-go"
)

const SlackPostMessageAPI = "https://slack.com/api/chat.postMessage"

func (options *Options) SendThreaded(message string) error {

	payload := APIRequest{
		Channel: options.SlackChannel,
		Text:    message,
		TS:      options.SlackThreadTS,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error while sending slack message: %s ", err)
	}

	r, err := retryablehttp.NewRequest(http.MethodPost, SlackPostMessageAPI, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	r.Header.Set("Authorization", "Bearer "+options.SlackToken)
	r.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient(retryablehttp.DefaultOptionsSingle)
	res, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("error while sending slack message: %s ", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error while sending slack message: %s ", fmt.Errorf("request failed with status code %d ", res.StatusCode))
	}

	var response APIResponse
	if err = jsoniter.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("error trying to unmarshal the response: %v", err)
	}

	if !response.Ok {
		return fmt.Errorf("error while sending slack message: %s ", response.Error)
	}

	if options.SlackThreadTS == "" {
		options.SlackThreadTS = response.TS
	}
	return nil
}
