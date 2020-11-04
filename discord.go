package notify

// From https://dev.to/arunx2/simple-slack-notification-with-golang-55i2

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/projectdiscovery/retryablehttp-go"
)

// DefaultDiscordTimeout to conclude operations
const DefaultDiscordTimeout = 5 * time.Second

// DiscordClient handling webhooks
type DiscordClient struct {
	client     *retryablehttp.Client
	WebHookURL string
	UserName   string
	Avatar     string
	TimeOut    time.Duration
}

// DiscordMessage json structure
type DiscordMessage struct {
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Content   string `json:"content,omitempty"`
}

// SendInfo to discord
func (dc *DiscordClient) SendInfo(message string) (err error) {
	return dc.SendDiscordNotification(&DiscordMessage{
		Content:   message,
		Username:  dc.UserName,
		AvatarURL: dc.Avatar,
	})
}

// SendDiscordNotification with json structure
func (dc *DiscordClient) SendDiscordNotification(discordMessage *DiscordMessage) error {
	return dc.sendHTTPRequest(discordMessage)
}

func (dc *DiscordClient) sendHTTPRequest(discordMessage *DiscordMessage) error {
	discordBody, err := json.Marshal(discordMessage)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, dc.WebHookURL, bytes.NewBuffer(discordBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
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

	if string(buf) != ok {
		return err
	}

	return nil
}
