package gitlab

import (
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
	"golang.org/x/time/rate"
)

// newRateLimiter must produce a value usable as go-gitlab's RateLimiter.
var _ gl.RateLimiter = newRateLimiter(1)

func TestNewRateLimiter(t *testing.T) {
	t.Run("non-positive disables limiting", func(t *testing.T) {
		if got := newRateLimiter(0).Limit(); got != rate.Inf {
			t.Errorf("limit = %v, want Inf", got)
		}
	})

	t.Run("positive sets rate and burst", func(t *testing.T) {
		l := newRateLimiter(10)
		if l.Limit() != rate.Limit(10) {
			t.Errorf("limit = %v, want 10", l.Limit())
		}
		if l.Burst() != 10 {
			t.Errorf("burst = %d, want 10", l.Burst())
		}
	})

	t.Run("fractional rate keeps burst at least 1", func(t *testing.T) {
		if got := newRateLimiter(0.5).Burst(); got < 1 {
			t.Errorf("burst = %d, want >= 1", got)
		}
	})
}
