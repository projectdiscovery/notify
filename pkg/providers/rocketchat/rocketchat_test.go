package rocketchat

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestRocketChatSendWebhook(t *testing.T) {
	var receivedPayloads []webhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var p webhookPayload
		err = json.Unmarshal(body, &p)
		require.NoError(t, err)

		receivedPayloads = append(receivedPayloads, p)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	opts := []*Options{
		{
			ID:                   "test-rc",
			RocketChatWebhookURL: server.URL,
			Channel:              "#security-alerts",
			Username:             "NotifyBot",
		},
	}

	provider, err := New(opts, []string{"test-rc"})
	require.NoError(t, err)
	require.NotNil(t, provider)

	err = provider.Send("Security alert: port 22 open", "")
	require.NoError(t, err)
	require.Len(t, receivedPayloads, 1)
	require.Equal(t, "Security alert: port 22 open", receivedPayloads[0].Text)
	require.Equal(t, "#security-alerts", receivedPayloads[0].Channel)
	require.Equal(t, "NotifyBot", receivedPayloads[0].Alias)
}

func TestRocketChatShoutrrrURLIncludesChannelAndUsername(t *testing.T) {
	got, err := rocketchatShoutrrrURL(&Options{
		RocketChatWebhookURL: "rocketchat://chat.example.com/tokenA/tokenB",
		Channel:              "#security-alerts",
		Username:             "NotifyBot",
	})
	require.NoError(t, err)
	require.Equal(t, "rocketchat://NotifyBot@chat.example.com/tokenA/tokenB/security-alerts", got)
}

func TestRocketChatSplitMessage(t *testing.T) {
	part1 := strings.Repeat("A", 3000)
	part2 := strings.Repeat("B", 3000)
	fullMsg := part1 + "\n" + part2

	chunks := splitMessage(fullMsg, 4000)
	require.Len(t, chunks, 2)
	require.Equal(t, part1, chunks[0])
	require.Equal(t, part2, chunks[1])

	accented := strings.Repeat("A", 3999) + "é" + strings.Repeat("B", 10)
	chunks = splitMessage(accented, 4000)
	require.NotEmpty(t, chunks)
	for _, chunk := range chunks {
		require.True(t, utf8.ValidString(chunk))
	}
	require.Equal(t, accented, strings.Join(chunks, ""))
}
