package notion

type Page struct {
	ParentId   string                   `json:"parent_id"`
	Properties map[string]interface{}   `json:"properties"`
	Children   []map[string]interface{} `json:"children"`
}
