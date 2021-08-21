package slack

type APIRequest struct {
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text,omitempty"`
	TS      string `json:"thread_ts,omitempty"`
}

type APIResponse struct {
	Ok    bool   `json:"ok,omitempty"`
	TS    string `json:"ts,omitempty"`
	Error string `json:"error,omitempty"`
}
