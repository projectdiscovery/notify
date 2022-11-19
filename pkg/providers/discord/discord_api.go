package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/projectdiscovery/notify/pkg/utils/httpreq"
)

func (options *Options) SendThreaded(message string) error {

	payload := APIRequest{
		Content:   message,
		Username:  options.DiscordWebHookUsername,
		AvatarURL: options.DiscordWebHookAvatarURL,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	body := bytes.NewReader(encoded)

	webHookURL := fmt.Sprintf("%s?thread_id=%s", options.DiscordWebHookURL, options.DiscordThreadID)

	req, err := http.NewRequest(http.MethodPost, webHookURL, body)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	_, err = httpreq.NewClient().Do(req)
	if err != nil {
		return err
	}

	return nil
}
