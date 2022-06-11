package utils

import (
	"net/http"
	"time"

	"go.uber.org/ratelimit"
)

// ThrottledTransport is Rate Limited HTTP Client
type ThrottledTransport struct {
	roundTripperWrap http.RoundTripper
	ratelimiter      ratelimit.Limiter
}

// RoundTrip implements the http.RoundTripper interface
func (c *ThrottledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// This is a blocking call. Honors the rate limit
	c.ratelimiter.Take()
	return c.roundTripperWrap.RoundTrip(r)
}

// NewThrottledTransport wraps transport with a rate limitter
func NewThrottledTransport(limitPeriod time.Duration, requestCount int, transport http.RoundTripper) http.RoundTripper {
	return &ThrottledTransport{
		roundTripperWrap: transport,
		ratelimiter:      ratelimit.New(requestCount, ratelimit.Per(limitPeriod)),
	}
}
