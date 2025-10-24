package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter defines the interface for rate limiting
type Limiter interface {
	// Wait blocks until the rate limiter permits an action or context is canceled
	Wait(ctx context.Context) error
	// Allow returns true if an action can be performed immediately
	Allow() bool
	// Reserve reserves a token and returns the time to wait
	Reserve() time.Duration
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	rate       float64       // tokens per second
	burst      int           // maximum burst size
	tokens     float64       // current tokens
	lastUpdate time.Time     // last time tokens were added
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter
// rate: requests per second
// burst: maximum burst size (set to rate for no burst)
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	if burst < 1 {
		burst = 1
	}
	return &TokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// refill adds tokens based on elapsed time
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()

	// Add tokens based on elapsed time
	tb.tokens += elapsed * tb.rate

	// Cap at burst size
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}

	tb.lastUpdate = now
}

// Wait blocks until a token is available or context is canceled
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		// Check if we can proceed immediately
		tb.mu.Lock()
		tb.refill()

		if tb.tokens >= 1.0 {
			tb.tokens -= 1.0
			tb.mu.Unlock()
			return nil
		}

		// Calculate wait time
		tokensNeeded := 1.0 - tb.tokens
		waitDuration := time.Duration(tokensNeeded / tb.rate * float64(time.Second))
		tb.mu.Unlock()

		// Wait with context support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
			// Continue to next iteration to try again
		}
	}
}

// Allow returns true if a token is immediately available
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// Reserve reserves a token and returns the duration to wait
func (tb *TokenBucket) Reserve() time.Duration {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return 0
	}

	// Calculate wait time
	tokensNeeded := 1.0 - tb.tokens
	waitDuration := time.Duration(tokensNeeded / tb.rate * float64(time.Second))

	// Reserve the token (go negative)
	tb.tokens -= 1.0

	return waitDuration
}

// MultiLimiter manages multiple rate limiters for different endpoints
type MultiLimiter struct {
	limiters map[string]Limiter
	mu       sync.RWMutex
}

// NewMultiLimiter creates a new multi-limiter
func NewMultiLimiter() *MultiLimiter {
	return &MultiLimiter{
		limiters: make(map[string]Limiter),
	}
}

// AddLimiter adds a rate limiter for a specific key (e.g., endpoint path)
func (ml *MultiLimiter) AddLimiter(key string, limiter Limiter) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.limiters[key] = limiter
}

// Wait waits for the rate limiter associated with the key
func (ml *MultiLimiter) Wait(ctx context.Context, key string) error {
	ml.mu.RLock()
	limiter, exists := ml.limiters[key]
	ml.mu.RUnlock()

	if !exists {
		// No rate limit for this key
		return nil
	}

	return limiter.Wait(ctx)
}

// Allow checks if an action is allowed for the key
func (ml *MultiLimiter) Allow(key string) bool {
	ml.mu.RLock()
	limiter, exists := ml.limiters[key]
	ml.mu.RUnlock()

	if !exists {
		// No rate limit for this key
		return true
	}

	return limiter.Allow()
}

// NoOpLimiter is a rate limiter that never blocks
type NoOpLimiter struct{}

// NewNoOpLimiter creates a limiter that allows all requests
func NewNoOpLimiter() *NoOpLimiter {
	return &NoOpLimiter{}
}

func (n *NoOpLimiter) Wait(ctx context.Context) error {
	return nil
}

func (n *NoOpLimiter) Allow() bool {
	return true
}

func (n *NoOpLimiter) Reserve() time.Duration {
	return 0
}
