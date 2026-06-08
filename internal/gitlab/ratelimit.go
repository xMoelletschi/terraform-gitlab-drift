package gitlab

import "golang.org/x/time/rate"

// newRateLimiter builds a token-bucket limiter from a requests-per-second value,
// for go-gitlab's WithCustomLimiter. A non-positive value disables limiting
// (rate.Inf). The burst is ~one second of capacity so concurrent fetchers can
// proceed up to the configured rate without starving each other.
func newRateLimiter(reqPerSec float64) *rate.Limiter {
	if reqPerSec <= 0 {
		return rate.NewLimiter(rate.Inf, 1)
	}
	burst := int(reqPerSec)
	if burst < 1 {
		burst = 1
	}
	return rate.NewLimiter(rate.Limit(reqPerSec), burst)
}
