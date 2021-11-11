package utils

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// ThrottledTransport is Rate Limited HTTP Client
type ThrottledTransport struct {
	roundTripperWrap http.RoundTripper
	ratelimiter      *rate.Limiter
}

// RoundTrip implements the http.RoundTripper interface
func (c *ThrottledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// This is a blocking call. Honors the rate limit
	err := c.ratelimiter.Wait(r.Context())
	if err != nil {
		return nil, err
	}
	return c.roundTripperWrap.RoundTrip(r)
}

// NewThrottledTransport wraps transport with a rate limitter
func NewThrottledTransport(limitPeriod time.Duration, requestCount int, transport http.RoundTripper) http.RoundTripper {
	return &ThrottledTransport{
		roundTripperWrap: transport,
		ratelimiter:      rate.NewLimiter(rate.Every(limitPeriod), requestCount),
	}
}
