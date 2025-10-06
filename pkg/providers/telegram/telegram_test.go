package telegram

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

// capturedURL will store the URL passed to the mocked shoutrrrSendFunc
var capturedURL string

// TestTelegramSendURLWithThreadID tests the Send method of the Telegram provider,
// focusing on how the URL is constructed with and without a thread ID.
func TestTelegramSendURLWithThreadID(t *testing.T) {
	// Store the original shoutrrrSendFunc and defer its restoration
	originalSendFunc := shoutrrrSendFunc
	defer func() {
		shoutrrrSendFunc = originalSendFunc
	}()

	// Mock shoutrrrSendFunc to capture the URL and avoid actual sending
	shoutrrrSendFunc = func(serviceURL string, message string) error {
		capturedURL = serviceURL
		return nil // Simulate success
	}

	tests := []struct {
		name               string
		options            Options
		expectedChatIDInURL string
	}{
		{
			name: "with thread_id",
			options: Options{
				ID:               "test-with-thread",
				TelegramAPIKey:   "testAPIKey",
				TelegramChatID:   "testChatID",
				TelegramThreadID: "testThreadID",
			},
			expectedChatIDInURL: "testChatID:testThreadID",
		},
		{
			name: "without thread_id",
			options: Options{
				ID:             "test-without-thread",
				TelegramAPIKey: "testAPIKey2",
				TelegramChatID: "testChatID2",
			},
			expectedChatIDInURL: "testChatID2",
		},
		{
			name: "with thread_id but empty",
			options: Options{
				ID:               "test-with-empty-thread",
				TelegramAPIKey:   "testAPIKey3",
				TelegramChatID:   "testChatID3",
				TelegramThreadID: "", // Explicitly empty
			},
			expectedChatIDInURL: "testChatID3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedURL = "" // Reset captured URL for each test run
			provider, err := New([]*Options{&tt.options}, nil)
			require.NoError(t, err, "New() should not return an error")
			require.NotNil(t, provider, "New() should return a provider")
			require.Len(t, provider.Telegram, 1, "Provider should have one Telegram option")

			err = provider.Send("test message", "")
			require.NoError(t, err, "Send() should not return an error")

			parsedURL, err := url.Parse(capturedURL)
			require.NoError(t, err, "Captured URL should be parseable")

			channels := parsedURL.Query().Get("channels")
			require.Equal(t, tt.expectedChatIDInURL, channels, "Chat ID in URL does not match expected")

			// Verify other parts of the URL
			expectedScheme := "telegram"
			require.Equal(t, expectedScheme, parsedURL.Scheme, "URL scheme does not match")

			// Check API Key (Username part of Userinfo)
			expectedAPIKey := tt.options.TelegramAPIKey
			require.NotNil(t, parsedURL.User, "URL Userinfo should not be nil")
			actualAPIKey := parsedURL.User.Username()
			require.Equal(t, expectedAPIKey, actualAPIKey, "URL API key (username) does not match")

			// Check Host part
			expectedHost := "telegram"
			require.Equal(t, expectedHost, parsedURL.Host, "URL host does not match")

			parseMode := parsedURL.Query().Get("parsemode")
			// Default ParseMode is "None" if not specified in options
			expectedParseMode := "None"
			if tt.options.TelegramParseMode != "" {
				expectedParseMode = tt.options.TelegramParseMode
			}
			require.Equal(t, expectedParseMode, parseMode, "URL parsemode does not match")
		})
	}
}
