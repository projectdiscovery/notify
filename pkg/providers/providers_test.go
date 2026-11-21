package providers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/projectdiscovery/notify/pkg/providers/custom"
	"github.com/projectdiscovery/notify/pkg/types"
)

// TestNewRejectsUnknownProvider covers a -provider/-p value that matches no
// provider group.  It used to be ignored, so the run exited successfully
// without sending anything.
func TestNewRejectsUnknownProvider(t *testing.T) {
	providerOptions := &ProviderOptions{
		Custom: []*custom.Options{{ID: "local", CustomWebhookURL: "http://127.0.0.1:1/hook"}},
	}

	_, err := New(providerOptions, &types.Options{Providers: []string{"slackk"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), `unknown provider "slackk"`)
	require.Contains(t, err.Error(), "available: slack, discord")

	_, err = New(providerOptions, &types.Options{Providers: []string{"custom", "nosuch"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), `unknown provider "nosuch"`)
}

// TestNewRejectsUnconfiguredProvider covers valid provider names that the
// provider config file does not contain.
func TestNewRejectsUnconfiguredProvider(t *testing.T) {
	providerOptions := &ProviderOptions{
		Custom: []*custom.Options{{ID: "local", CustomWebhookURL: "http://127.0.0.1:1/hook"}},
	}

	_, err := New(providerOptions, &types.Options{Providers: []string{"slack"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "none of the requested providers (slack) are configured")
}

// TestNewAcceptsConfiguredProviders keeps the working paths intact.
func TestNewAcceptsConfiguredProviders(t *testing.T) {
	providerOptions := &ProviderOptions{
		Custom: []*custom.Options{{ID: "local", CustomWebhookURL: "http://127.0.0.1:1/hook"}},
	}

	client, err := New(providerOptions, &types.Options{})
	require.NoError(t, err)
	require.Len(t, client.providers, 1)

	client, err = New(providerOptions, &types.Options{Providers: []string{"custom"}})
	require.NoError(t, err)
	require.Len(t, client.providers, 1)
}
