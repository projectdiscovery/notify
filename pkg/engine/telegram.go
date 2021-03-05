package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/projectdiscovery/retryablehttp-go"
)

// DefaultTelegraTimeout to conclude operations
const (
	DefaultTelegraTimeout = 5 * time.Second
	Endpoint              = "https://api.telegram.org/bot{{apikey}}/sendMessage?chat_id={{chatid}}&text={{message}}"
)

// TelegramClient handling webhooks
type TelegramClient struct {
	client  *retryablehttp.Client
	apiKEY  string
	chatID  string
	TimeOut time.Duration
}

// SendInfo to telegram
func (dc *TelegramClient) SendInfo(message string) (err error) {
	return dc.sendHTTPRequest(message)
}

func (dc *TelegramClient) sendHTTPRequest(message string) error {
	r := strings.NewReplacer(
		"{{apikey}}", dc.apiKEY,
		"{{chatid}}", dc.chatID,
		"{{message}}", message,
	)
	URL := r.Replace(Endpoint)
	req, err := retryablehttp.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return err
	}
	resp, err := dc.client.Do(req)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//nolint:errcheck // silent fail
	defer resp.Body.Close()

	var tgresponse TelegramResponse

	err = json.Unmarshal(buf, &tgresponse)
	if err != nil {
		return err
	}

	if !tgresponse.Ok {
		return fmt.Errorf("%s", tgresponse.Description)
	}

	return nil
}

// TelegramResponse structure
type TelegramResponse struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}
