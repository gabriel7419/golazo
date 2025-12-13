package fotmob

import (
	"sync"
	"time"
)

// RateLimiter provides conservative rate limiting for API requests.
type RateLimiter struct {
	mu              sync.Mutex
	lastRequestTime time.Time
	minInterval     time.Duration
}

// NewRateLimiter creates a new rate limiter with conservative settings.
// minInterval: minimum time between requests (default: 2 seconds)
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	if minInterval < 2*time.Second {
		minInterval = 2 * time.Second // Conservative default
	}
	return &RateLimiter{
		minInterval: minInterval,
	}
}

// Wait ensures minimum time has passed since last request.
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRequestTime)

	if elapsed < rl.minInterval {
		waitTime := rl.minInterval - elapsed
		time.Sleep(waitTime)
	}

	rl.lastRequestTime = time.Now()
}
