package custom

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func newRecorder(t *testing.T) (*httptest.Server, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		mu.Unlock()
	}))
	t.Cleanup(srv.Close)
	return srv, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string{}, bodies...)
	}
}

func send(t *testing.T, opts *Options, message string) string {
	t.Helper()
	srv, bodies := newRecorder(t)
	opts.ID = "test"
	opts.CustomWebhookURL = srv.URL
	opts.CustomMethod = http.MethodPost
	p, err := New([]*Options{opts}, nil)
	require.NoError(t, err)
	require.NoError(t, p.Send(message, ""))
	got := bodies()
	require.Len(t, got, 1)
	return got[0]
}

var jsonHeaders = map[string]string{"Content-Type": "application/json"}

func TestCustomFormatDataInJSONStringIsEscaped(t *testing.T) {
	line := `https://target.example [x","channel":"#other-chan","text":"All clear, no findings]`
	body := send(t, &Options{CustomFormat: `{"channel":"#scan-results","text":"{{data}}"}`, CustomHeaders: jsonHeaders}, line)

	var parsed map[string]string
	require.NoError(t, json.Unmarshal([]byte(body), &parsed), body)
	require.Equal(t, "#scan-results", parsed["channel"])
	require.Equal(t, line, parsed["text"])
}

func TestCustomFormatRawDataIsForwarded(t *testing.T) {
	line := `{"url":"https://target.example/","status":200}`
	body := send(t, &Options{CustomFormat: `{{data}}`, CustomHeaders: jsonHeaders}, line)
	require.Equal(t, line, body)

	body = send(t, &Options{CustomFormat: `{"event":{{data}},"note":"{{data}}"}`, CustomHeaders: jsonHeaders}, line)
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(body), &parsed), body)
	require.Equal(t, line, parsed["note"])
	require.Equal(t, map[string]interface{}{"url": "https://target.example/", "status": float64(200)}, parsed["event"])
}

func TestCustomFormatWithoutJSONIsUnchanged(t *testing.T) {
	line := `a "quoted" line`
	body := send(t, &Options{CustomFormat: `text: {{data}}`}, line)
	require.Equal(t, `text: a "quoted" line`, body)
}

func TestCustomSprigValuesInJSONStringAreEscaped(t *testing.T) {
	line := `{"url":"https://target.example/x\",\"text\":\"extra: text","host":"h\"1","n":3}`
	body := send(t, &Options{CustomSprig: `{"text":"{{ .url }}"}`, CustomHeaders: jsonHeaders}, line)
	var parsed map[string]string
	require.NoError(t, json.Unmarshal([]byte(body), &parsed), body)
	require.Equal(t, `https://target.example/x","text":"extra: text`, parsed["text"])

	// branches, numbers and values outside strings
	body = send(t, &Options{
		CustomSprig:   `{"a":"{{ if .host }}{{ .host }}{{ else }}none{{ end }}","n":{{ .n }},"u":{{ .url | toJson }}}`,
		CustomHeaders: jsonHeaders,
	}, line)
	var parsed2 map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(body), &parsed2), body)
	require.Equal(t, `h"1`, parsed2["a"])
	require.Equal(t, float64(3), parsed2["n"])
	require.Equal(t, `https://target.example/x","text":"extra: text`, parsed2["u"])
}

func TestInJSONString(t *testing.T) {
	require.True(t, inJSONString(`{"a":"`, false))
	require.False(t, inJSONString(`{"a":"x\"y"`, false))
	require.True(t, inJSONString(`{"a":"x\\`, false))
	require.False(t, inJSONString(`"`, true))
}
