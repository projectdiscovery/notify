package discord

type APIRequest struct {
	Content   string `json:"content,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Username  string `json:"username,omitempty"`
}
