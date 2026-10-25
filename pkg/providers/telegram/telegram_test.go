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

func TestTelegramURLIncludesThreadIDWhenSet(t *testing.T) {
	got := telegramURL(&Options{
		TelegramAPIKey:    "token",
		TelegramChatID:    "-1001234567890",
		TelegramThreadID:  "42",
		TelegramParseMode: "None",
	})
	want := "telegram://token@telegram?channels=-1001234567890:42&parsemode=None&preview=No"
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestTelegramURLPreservesExistingColonInChatID(t *testing.T) {
	got := telegramURL(&Options{
		TelegramAPIKey:    "token",
		TelegramChatID:    "-1001234567890:42",
		TelegramThreadID:  "99", // should not duplicate colon
		TelegramParseMode: "None",
	})
	want := "telegram://token@telegram?channels=-1001234567890:42&parsemode=None&preview=No"
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
