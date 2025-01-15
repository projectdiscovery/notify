package jandi

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/projectdiscovery/notify/pkg/utils/httpreq"
)

func (options *Options) SendMessage(message string) error {
	payload := APIRequest{
		Body: message,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	body := bytes.NewReader(encoded)
	req, err := http.NewRequest(http.MethodPost, options.JandiWebHookURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.tosslab.jandi-v2+json")
	req.Header.Set("Content-Type", "application/json")

	_, err = httpreq.NewClient().Do(req)
	if err != nil {
		return err
	}

	return nil
}
