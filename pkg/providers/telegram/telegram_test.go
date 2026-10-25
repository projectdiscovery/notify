package telegram

import "testing"

func TestTelegramURLDisablesPreviewByDefault(t *testing.T) {
	got := telegramURL(&Options{
		TelegramAPIKey:    "token",
		TelegramChatID:    "123",
		TelegramParseMode: "None",
	})
	want := "telegram://token@telegram?channels=123&parsemode=None&preview=No"
	if got != want {
		t.Fatalf("got %s", got)
	}
}

func TestTelegramURLEnablesPreviewWhenSet(t *testing.T) {
	got := telegramURL(&Options{
		TelegramAPIKey:    "token",
		TelegramChatID:    "123",
		TelegramParseMode: "None",
		TelegramPreview:   true,
	})
	want := "telegram://token@telegram?channels=123&parsemode=None&preview=Yes"
	if got != want {
		t.Fatalf("got %s", got)
	}
}
