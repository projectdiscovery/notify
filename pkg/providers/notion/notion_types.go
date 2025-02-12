package notion


type Page struct {
	ParentId   string                   `json:"parent_id"`
	Properties map[string]interface{}   `json:"properties"`
	Children   []map[string]interface{} `json:"children"`
}

type APIRequest struct {
	Parent 		map[string]interface{} `json:"parent,omitempty"`
	Properties	map[string]interface{} `json:"properties,omitempty"`
	Children	[]map[string]interface{} `json:"children,omitempty"`
}

type APIResponse struct {
	Object			string `json:"object,omitempty"`
	RequestID    	string `json:"request_id,omitempty"`
	Message 		string `json:"message,omitempty"`
}
