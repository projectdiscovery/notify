package runner

import (
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/projectdiscovery/notify/pkg/types"
)

// newConnectProxy starts a local proxy that tunnels CONNECT requests.
func newConnectProxy(t *testing.T, connected *atomic.Bool) *httptest.Server {
	t.Helper()
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			http.Error(w, "only CONNECT", http.StatusMethodNotAllowed)
			return
		}
		connected.Store(true)
		upstream, err := net.Dial("tcp", r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			_ = upstream.Close()
			return
		}
		w.WriteHeader(http.StatusOK)
		client, _, err := hijacker.Hijack()
		if err != nil {
			_ = upstream.Close()
			return
		}
		go func() {
			_, _ = io.Copy(upstream, client)
			_ = upstream.Close()
		}()
		_, _ = io.Copy(client, upstream)
		_ = client.Close()
	}))
	t.Cleanup(proxy.Close)
	return proxy
}

func TestRunProxyKeepsCertificateVerification(t *testing.T) {
	origTransport := http.DefaultClient.Transport
	t.Cleanup(func() { http.DefaultClient.Transport = origTransport })

	var received atomic.Bool
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		received.Store(true)
	}))
	t.Cleanup(provider.Close)
	var connected atomic.Bool
	proxy := newConnectProxy(t, &connected)

	data := filepath.Join(t.TempDir(), "empty.txt")
	require.NoError(t, os.WriteFile(data, nil, 0o600))
	r := &Runner{options: &types.Options{Proxy: proxy.URL, Data: data, CharLimit: 4000}}
	require.NoError(t, r.Run())

	req, err := http.NewRequest(http.MethodPost, provider.URL+"/hook", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.Error(t, err)
	var unknownAuthority x509.UnknownAuthorityError
	require.ErrorAs(t, err, &unknownAuthority)
	require.True(t, connected.Load(), "request must go through the proxy")
	require.False(t, received.Load(), "request must not reach a server with an untrusted certificate")
}
