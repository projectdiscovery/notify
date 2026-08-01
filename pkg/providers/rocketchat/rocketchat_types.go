package rocketchat

// WebhookPayload is the body sent to a Rocket.Chat Incoming Webhook.
// Reference: https://docs.rocket.chat/docs/integrations#incoming-webhook-script
type WebhookPayload struct {
	Text        string       `json:"text,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Alias       string       `json:"alias,omitempty"`
	Avatar      string       `json:"avatar,omitempty"`
	Emoji       string       `json:"emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Rocket.Chat message attachment.
type Attachment struct {
	Text  string `json:"text,omitempty"`
	Color string `json:"color,omitempty"`
}
