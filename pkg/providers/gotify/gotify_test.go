package gotify

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, status int, got *http.Request, body *messageRequest) *Options {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = *r.Clone(r.Context())
		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			t.Errorf("decode body: %s", err)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"error":"Unauthorized","errorCode":401,"errorDescription":"you need to provide a valid access token"}`))
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatal(err)
	}
	return &Options{ID: "gotify", GotifyHost: host, GotifyPort: port, GotifyDisableTLS: true, GotifyFormat: "{{data}}"}
}

func TestSend(t *testing.T) {
	for name, token := range map[string]string{
		"gotify 3 token": "gtfya." + strings.Repeat("x", 43),
		"legacy token":   "AbCdEfGhIjKlMnO",
	} {
		t.Run(name, func(t *testing.T) {
			var req http.Request
			var body messageRequest
			opts := newTestServer(t, http.StatusOK, &req, &body)
			opts.GotifyToken = token
			opts.GotifyTitle = "scan"

			p, err := New([]*Options{opts}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := p.Send("hello", ""); err != nil {
				t.Fatalf("send: %s", err)
			}

			if req.Method != http.MethodPost || req.URL.Path != "/message" {
				t.Errorf("got %s %s, want POST /message", req.Method, req.URL.Path)
			}
			if got := req.Header.Get("X-Gotify-Key"); got != token {
				t.Errorf("X-Gotify-Key = %q, want %q", got, token)
			}
			if req.URL.Query().Get("token") != "" {
				t.Error("token must not be sent in the URL")
			}
			if body.Message != "hello" || body.Title != "scan" {
				t.Errorf("body = %+v, want message hello and title scan", body)
			}
		})
	}
}

func TestSendDefaultTitle(t *testing.T) {
	var req http.Request
	var body messageRequest
	opts := newTestServer(t, http.StatusOK, &req, &body)
	opts.GotifyToken = "AbCdEfGhIjKlMnO"

	p, err := New([]*Options{opts}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Send("hello", ""); err != nil {
		t.Fatalf("send: %s", err)
	}
	if body.Title != defaultTitle {
		t.Errorf("title = %q, want %q", body.Title, defaultTitle)
	}
}

func TestSendErrorStatus(t *testing.T) {
	var req http.Request
	var body messageRequest
	opts := newTestServer(t, http.StatusUnauthorized, &req, &body)
	opts.GotifyToken = "AbCdEfGhIjKlMnO"

	p, err := New([]*Options{opts}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Send("hello", ""); err == nil {
		t.Fatal("expected an error for a 401 response")
	}
}
