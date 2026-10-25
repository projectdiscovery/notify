package gotify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGotify_Gotify3Token(t *testing.T) {
	var receivedMessage string
	var receivedTitle string
	var receivedToken string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/message" {
			t.Fatalf("expected /message path, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content-type, got %s", r.Header.Get("Content-Type"))
		}
		receivedToken = r.Header.Get("X-Gotify-Key")

		var payload gotifyPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode json body: %v", err)
		}

		receivedMessage = payload.Message
		receivedTitle = payload.Title

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1,"appid":1,"message":"ok"}`))
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse test server url: %v", err)
	}

	opts := &Options{
		ID:               "gotify3-test",
		GotifyHost:       u.Hostname(),
		GotifyPort:       u.Port(),
		GotifyToken:      "gtfya.ABC123xyz789", // Gotify 3 token (rejected by shoutrrr)
		GotifyTitle:      "Test Gotify 3",
		GotifyDisableTLS: true,
	}

	p, err := New([]*Options{opts}, nil)
	if err != nil {
		t.Fatalf("failed to create gotify provider: %v", err)
	}

	if err := p.Send("Hello from Gotify 3", ""); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if receivedMessage != "Hello from Gotify 3" {
		t.Fatalf("expected 'Hello from Gotify 3', got %q", receivedMessage)
	}
	if receivedTitle != "Test Gotify 3" {
		t.Fatalf("expected 'Test Gotify 3', got %q", receivedTitle)
	}
	if receivedToken != "gtfya.ABC123xyz789" {
		t.Fatalf("expected 'gtfya.ABC123xyz789', got %q", receivedToken)
	}
}

func TestGotify_Gotify3Token_ErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"Unauthorized","errorDescription":"you need to provide a valid token"}`, http.StatusUnauthorized)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse test server url: %v", err)
	}

	opts := &Options{
		ID:               "gotify3-err",
		GotifyHost:       u.Hostname(),
		GotifyPort:       u.Port(),
		GotifyToken:      "gtfya.invalid",
		GotifyDisableTLS: true,
	}

	p, err := New([]*Options{opts}, nil)
	if err != nil {
		t.Fatalf("failed to create gotify provider: %v", err)
	}

	err = p.Send("Message", "")
	if err == nil {
		t.Fatal("expected error on 401 response, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected error containing 401, got %v", err)
	}
}

func TestIsLegacyGotifyToken(t *testing.T) {
	// Legacy tokens: length 15, starts with 'A'
	if !isLegacyGotifyToken("Aaa.bbb.ccc.ddd") {
		t.Fatal("expected Aaa.bbb.ccc.ddd to be legacy token")
	}
	if !isLegacyGotifyToken("Ahwbsdyhwwgarxd") {
		t.Fatal("expected Ahwbsdyhwwgarxd to be legacy token")
	}

	// Non-legacy / Gotify 3 tokens
	if isLegacyGotifyToken("gtfya.ABC123xyz789") {
		t.Fatal("expected gtfya.ABC123xyz789 to not be legacy token")
	}
	if isLegacyGotifyToken("gtfyu.12345") {
		t.Fatal("expected gtfyu.12345 to not be legacy token")
	}
	if isLegacyGotifyToken("Chwbsdyhwwgarxd") {
		t.Fatal("expected Chwbsdyhwwgarxd to not be legacy token")
	}
	if isLegacyGotifyToken("Ahwbsdy") {
		t.Fatal("expected Ahwbsdy to not be legacy token")
	}
}
