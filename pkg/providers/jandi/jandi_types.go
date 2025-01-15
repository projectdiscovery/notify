package jandi

type APIRequest struct {
	Body    string `json:"body"`
	ConnectInfo   []ConnectInfo `json:"connectInfo,omitempty"`
}

type ConnectInfo struct {
	Title string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}
