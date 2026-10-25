package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/projectdiscovery/notify/pkg/types"
)

// On first run with no -pc flag, NewRunner should create the default provider
// config (directory + commented template) instead of failing fatally (#487).
func TestNewRunner_FirstRunCreatesDefaultProviderConfig(t *testing.T) {
	tmpHome := t.TempDir()
	// os.UserHomeDir uses $HOME on unix and %USERPROFILE% on Windows.
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	options := &types.Options{} // empty ProviderConfig => default path
	runner, err := NewRunner(options)
	if err != nil {
		t.Fatalf("NewRunner returned error on first run: %v", err)
	}
	if runner == nil {
		t.Fatal("NewRunner returned nil runner")
	}

	want := filepath.Join(tmpHome, types.DefaultProviderConfigLocation)
	if options.ProviderConfig != want {
		t.Fatalf("ProviderConfig = %q, want %q", options.ProviderConfig, want)
	}

	info, statErr := os.Stat(options.ProviderConfig)
	if statErr != nil {
		t.Fatalf("default provider config was not created: %v", statErr)
	}
	if runtime.GOOS != "windows" {
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("config file perms = %o, want 0600", perm)
		}
	}

	data, readErr := os.ReadFile(options.ProviderConfig)
	if readErr != nil {
		t.Fatalf("read created config: %v", readErr)
	}
	if !strings.Contains(string(data), "notify provider configuration") {
		t.Errorf("created config missing template banner, got:\n%s", data)
	}
}

// An explicitly supplied -pc path that does not exist must error, not silently
// create a template.
func TestNewRunner_ExplicitMissingConfigReturnsError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope", "provider-config.yaml")
	_, err := NewRunner(&types.Options{ProviderConfig: missing})
	if err == nil {
		t.Fatal("expected error for missing explicit config path, got nil")
	}
	if !strings.Contains(err.Error(), "provider config file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(missing); !os.IsNotExist(statErr) {
		t.Error("explicit missing path should not have been created")
	}
}

// An empty / comment-only config (e.g. the freshly created template) must not
// block startup.
func TestNewRunner_CommentOnlyConfigIsTolerated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "provider-config.yaml")
	if err := os.WriteFile(path, []byte("# only comments, no providers\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner, err := NewRunner(&types.Options{ProviderConfig: path})
	if err != nil {
		t.Fatalf("comment-only config should be tolerated, got: %v", err)
	}
	if runner == nil {
		t.Fatal("NewRunner returned nil runner")
	}
}
