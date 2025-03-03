package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RequestTracker tracks request counts in a sliding window
type RequestTracker struct {
	mu           sync.Mutex
	windowStart  time.Time
	currentCount int
}

// NewRequestTracker creates a new request tracker
func NewRequestTracker() *RequestTracker {
	return &RequestTracker{
		windowStart:  time.Now(),
		currentCount: 0,
	}
}

// TrackRequest records a request and returns the current rate
func (rt *RequestTracker) TrackRequest() int {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	now := time.Now()

	// If more than a second has passed, reset the window
	if now.Sub(rt.windowStart) >= time.Second {
		rt.windowStart = now
		rt.currentCount = 0
	}

	rt.currentCount++
	return rt.currentCount
}

// GetCurrentRPS returns the current request rate
func (rt *RequestTracker) GetCurrentRPS() int {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	// If the current window is still valid, return current count
	if time.Since(rt.windowStart) < time.Second {
		return rt.currentCount
	}
	return 0
}

// RateLimiter for per-client rate limiting
type RateLimiter struct {
	clients     map[string]*rate.Limiter
	requests    map[string]*RequestTracker
	mu          sync.Mutex
	rate        rate.Limit
	rps         int
	cleanupTick *time.Ticker
}

// NewRateLimiter initializes a rate limiter
func NewRateLimiter(rps int) *RateLimiter {
	rl := &RateLimiter{
		clients:     make(map[string]*rate.Limiter),
		requests:    make(map[string]*RequestTracker),
		rate:        rate.Every(time.Second / time.Duration(rps)),
		rps:         rps,
		cleanupTick: time.NewTicker(time.Minute), // Cleanup old entries every minute
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup periodically removes old entries from the maps
func (rl *RateLimiter) cleanup() {
	for range rl.cleanupTick.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, tracker := range rl.requests {
			tracker.mu.Lock()
			if now.Sub(tracker.windowStart) > time.Minute {
				delete(rl.requests, ip)
				delete(rl.clients, ip)
			}
			tracker.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// GetLimiter returns a rate limiter and request tracker for the given IP
func (rl *RateLimiter) GetLimiter(ip string) (*rate.Limiter, *RequestTracker) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.clients[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, 1)
		rl.clients[ip] = limiter
	}

	tracker, exists := rl.requests[ip]
	if !exists {
		tracker = NewRequestTracker()
		rl.requests[ip] = tracker
	}

	return limiter, tracker
}

// Close stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Close() {
	rl.cleanupTick.Stop()
}
