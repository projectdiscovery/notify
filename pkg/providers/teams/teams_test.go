package teams

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTeamsWebhookKind(t *testing.T) {
	kind, err := teamsWebhookKind("https://outlook.office.com/webhookb2/aaa/IncomingWebhook/bbb")
	if err != nil || kind != teamsWebhookConnector {
		t.Fatalf("connector: kind %q err %v", kind, err)
	}
	kind, err = teamsWebhookKind("https://prod.westus.logic.azure.com/workflows/abc/triggers/manual/paths/invoke")
	if err != nil || kind != teamsWebhookWorkflow {
		t.Fatalf("workflow: kind %q err %v", kind, err)
	}
	if _, err = teamsWebhookKind("https://example.com"); err == nil {
		t.Fatal("short url should fail")
	}
	if _, err = teamsWebhookKind("not a url"); err == nil {
		t.Fatal("garbage url should fail")
	}
}

func TestWorkflowPayloadEscapesQuotesAndNewlines(t *testing.T) {
	payload, err := workflowPayload("say \"hi\"\nnext <tag>")
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if body.Text != "say &#34;hi&#34;<br>next &lt;tag&gt;" {
		t.Fatalf("payload text %q", body.Text)
	}
}

func TestSendWorkflowPostsJSONAndRequiresAccepted(t *testing.T) {
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, _ := io.ReadAll(r.Body)
		got = string(buf)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	provider := &Provider{Teams: []*Options{{
		ID:              "one",
		TeamsWebHookURL: server.URL + "/workflows/abc/triggers/manual/paths/invoke",
	}}}
	if err := provider.Send("hello \"world\"", ""); err != nil {
		t.Fatal(err)
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(got), &body); err != nil {
		t.Fatal(err)
	}
	if body.Text != "hello &#34;world&#34;" {
		t.Fatalf("body %s", got)
	}

	rejected := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer rejected.Close()
	provider.Teams[0].TeamsWebHookURL = rejected.URL + "/workflows/abc/triggers/manual/paths/invoke"
	if err := provider.Send("hello", ""); err == nil {
		t.Fatal("expected status error")
	}
}

func TestSendRejectsUnsupportedURLWithoutPanic(t *testing.T) {
	provider := &Provider{Teams: []*Options{{
		ID:              "one",
		TeamsWebHookURL: "https://example.com/not-teams",
	}}}
	err := provider.Send("hello", "")
	if err == nil || !strings.Contains(err.Error(), "unsupported webhook url") {
		t.Fatalf("err %v", err)
	}
}
