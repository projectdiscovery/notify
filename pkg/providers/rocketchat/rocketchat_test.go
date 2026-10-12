package rocketchat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestNewFiltersByID(t *testing.T) {
	options := []*Options{
		{ID: "soc", RocketchatWebHookURL: "https://chat.example.com/hooks/soc"},
		{ID: "vulns", RocketchatWebHookURL: "https://chat.example.com/hooks/vulns"},
	}

	provider, err := New(options, []string{"soc"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(provider.Rocketchat) != 1 || provider.Rocketchat[0].ID != "soc" {
		t.Fatalf("expected only id 'soc', got %+v", provider.Rocketchat)
	}

	provider, err = New(options, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(provider.Rocketchat) != 2 {
		t.Fatalf("expected all options when ids is empty, got %d", len(provider.Rocketchat))
	}
}

func TestSendSuccess(t *testing.T) {
	var received WebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode payload: %s", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider, err := New([]*Options{
		{
			ID:                   "soc",
			RocketchatWebHookURL: server.URL,
			RocketchatChannel:    "#alertas-seguranca",
			RocketchatUsername:   "ProjectDiscovery",
			RocketchatAvatar:     "https://example.com/avatar.png",
			RocketchatEmoji:      ":ghost:",
		},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := provider.Send("host.example.com is vulnerable", ""); err != nil {
		t.Fatalf("unexpected send error: %s", err)
	}

	if received.Text != "host.example.com is vulnerable" {
		t.Errorf("unexpected text: %q", received.Text)
	}
	if received.Channel != "#alertas-seguranca" {
		t.Errorf("unexpected channel: %q", received.Channel)
	}
	if received.Alias != "ProjectDiscovery" {
		t.Errorf("unexpected alias: %q", received.Alias)
	}
	if received.Avatar != "https://example.com/avatar.png" {
		t.Errorf("unexpected avatar: %q", received.Avatar)
	}
	if received.Emoji != ":ghost:" {
		t.Errorf("unexpected emoji: %q", received.Emoji)
	}
}

func TestSendAsAttachment(t *testing.T) {
	var received WebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode payload: %s", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider, err := New([]*Options{
		{
			ID:                   "soc",
			RocketchatWebHookURL: server.URL,
			RocketchatAttachment: true,
		},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := provider.Send("finding details", ""); err != nil {
		t.Fatalf("unexpected send error: %s", err)
	}

	if received.Text != "" {
		t.Errorf("expected empty text when using attachments, got %q", received.Text)
	}
	if len(received.Attachments) != 1 || received.Attachments[0].Text != "finding details" {
		t.Fatalf("unexpected attachments: %+v", received.Attachments)
	}
}

func TestSendInvalidWebhookURL(t *testing.T) {
	provider, err := New([]*Options{
		{ID: "soc", RocketchatWebHookURL: "not-a-url"},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	err = provider.Send("message", "")
	if err == nil || !strings.Contains(err.Error(), "invalid rocketchat webhook URL") {
		t.Fatalf("expected invalid webhook URL error, got: %v", err)
	}
}

func TestSendNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider, err := New([]*Options{
		{ID: "soc", RocketchatWebHookURL: server.URL},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := provider.Send("message", ""); err == nil {
		t.Fatal("expected error for non-success status code")
	}
}

func TestSendWithClientHonorsContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := sendWithClient(ctx, &http.Client{}, server.URL, WebhookPayload{Text: "message"})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestSendWithClientReportsBodyCloseError(t *testing.T) {
	client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       closeErrorBody{},
		}, nil
	})}

	err := sendWithClient(context.Background(), client, "http://example.com", WebhookPayload{Text: "message"})
	if err == nil || !strings.Contains(err.Error(), "close") {
		t.Fatalf("expected response body close error, got %v", err)
	}
}

func TestSendSplitsLargeMessages(t *testing.T) {
	var payloads []WebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatalf("failed to decode payload: %s", err)
		}
		payloads = append(payloads, p)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider, err := New([]*Options{
		{ID: "soc", RocketchatWebHookURL: server.URL},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	line := strings.Repeat("a", 100)
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = line
	}
	bigMessage := strings.Join(lines, "\n")

	if err := provider.Send(bigMessage, ""); err != nil {
		t.Fatalf("unexpected send error: %s", err)
	}
	if len(payloads) <= 1 {
		t.Fatalf("expected message to be split into multiple requests, got %d", len(payloads))
	}

	var rebuilt strings.Builder
	for i, p := range payloads {
		if len(p.Text) > defaultMaxMessageLength {
			t.Errorf("chunk exceeds max length: %d", len(p.Text))
		}
		if i > 0 {
			rebuilt.WriteByte('\n')
		}
		rebuilt.WriteString(p.Text)
	}
	if rebuilt.String() != bigMessage {
		t.Fatal("rebuilt message from chunks does not match original")
	}
}

func TestSplitMessage(t *testing.T) {
	t.Run("under limit returns single chunk", func(t *testing.T) {
		chunks := splitMessage("short message", 100)
		if len(chunks) != 1 || chunks[0] != "short message" {
			t.Fatalf("unexpected chunks: %+v", chunks)
		}
	})

	t.Run("splits on newlines without breaking lines", func(t *testing.T) {
		msg := strings.Repeat("a", 40) + "\n" + strings.Repeat("b", 40) + "\n" + strings.Repeat("c", 40)
		chunks := splitMessage(msg, 50)
		if len(chunks) != 3 {
			t.Fatalf("expected 3 chunks, got %d: %+v", len(chunks), chunks)
		}
		for _, c := range chunks {
			if len(c) > 50 {
				t.Errorf("chunk exceeds limit: %q", c)
			}
		}
	})

	t.Run("hard splits a single line longer than limit", func(t *testing.T) {
		msg := strings.Repeat("x", 250)
		chunks := splitMessage(msg, 100)
		if len(chunks) != 3 {
			t.Fatalf("expected 3 chunks, got %d", len(chunks))
		}
		if strings.Join(chunks, "") != msg {
			t.Fatal("rejoined chunks do not match original message")
		}
	})

	t.Run("does not split UTF-8 characters", func(t *testing.T) {
		msg := strings.Repeat("é", 60)
		chunks := splitMessage(msg, 50)
		if len(chunks) != 2 {
			t.Fatalf("expected 2 chunks, got %d", len(chunks))
		}
		for _, chunk := range chunks {
			if !utf8.ValidString(chunk) {
				t.Fatalf("chunk is not valid UTF-8: %q", chunk)
			}
			if len([]rune(chunk)) > 50 {
				t.Errorf("chunk exceeds rune limit: %d", len([]rune(chunk)))
			}
		}
		if strings.Join(chunks, "") != msg {
			t.Fatal("rejoined chunks do not match original message")
		}
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type closeErrorBody struct{}

func (closeErrorBody) Read([]byte) (int, error) { return 0, io.EOF }

func (closeErrorBody) Close() error { return io.ErrUnexpectedEOF }
