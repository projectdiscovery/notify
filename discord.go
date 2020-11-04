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

const DefaultDiscordTimeout = 5 * time.Second

type DiscordClient struct {
	client     *retryablehttp.Client
	WebHookUrl string
	UserName   string
	Avatar     string
	TimeOut    time.Duration
}

type DiscordMessage struct {
	Username  string `json:"username,omitempty"`
	AvatarUrl string `json:"avatar_url,omitempty"`
	Content   string `json:"content,omitempty"`
}

func (dc *DiscordClient) SendInfo(message string) (err error) {
	return dc.SendDiscordNotification(&DiscordMessage{
		Content:   message,
		Username:  dc.UserName,
		AvatarUrl: dc.Avatar,
	})
}

func (dc *DiscordClient) SendDiscordNotification(discordMessage *DiscordMessage) error {
	return dc.sendHttpRequest(discordMessage)
}

func (dc *DiscordClient) sendHttpRequest(discordMessage *DiscordMessage) error {
	discordBody, err := json.Marshal(discordMessage)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, dc.WebHookUrl, bytes.NewBuffer(discordBody))
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
	defer resp.Body.Close()

	if string(buf) != "ok" {
		return err
	}

	return nil
}
