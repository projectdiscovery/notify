package slack

import (
	"fmt"
	"net/http"

	"github.com/projectdiscovery/notify/pkg/utils/httpreq"
)

const SlackPostMessageAPI = "https://slack.com/api/chat.postMessage"

func (options *Options) SendThreaded(message string) error {

	payload := APIRequest{
		Channel: options.SlackChannel,
		Text:    message,
		TS:      options.SlackThreadTS,
	}

	headers := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", options.SlackToken)},
	}

	var response *APIResponse

	err := httpreq.NewClient().Post(SlackPostMessageAPI, &payload, headers, &response)
	if err != nil {
		return err
	}
	if !response.Ok {
		return fmt.Errorf("error while sending slack message: %s ", response.Error)
	}

	if options.SlackThreadTS == "" {
		options.SlackThreadTS = response.TS
	}
	return nil
}
