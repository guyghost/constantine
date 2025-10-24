package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	// Create a limiter with 10 requests per second, burst of 5
	limiter := NewTokenBucket(10, 5)

	// Should allow burst size requests immediately
	for i := 0; i < 5; i++ {
		if !limiter.Allow() {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	// Next request should be denied (bucket empty)
	if limiter.Allow() {
		t.Error("Request should be denied when bucket is empty")
	}

	// Wait for refill (100ms = 1 token at 10 req/s)
	time.Sleep(110 * time.Millisecond)

	// Should allow one more request
	if !limiter.Allow() {
		t.Error("Request should be allowed after refill")
	}
}

func TestTokenBucket_Wait(t *testing.T) {
	// Create a limiter with 5 requests per second, burst of 2
	limiter := NewTokenBucket(5, 2)

	ctx := context.Background()

	// First two requests should be immediate
	start := time.Now()
	for i := 0; i < 2; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}

	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		t.Errorf("First burst should be immediate, took %v", elapsed)
	}

	// Third request should wait ~200ms (1/5 second)
	start = time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
	elapsed = time.Since(start)

	if elapsed < 150*time.Millisecond || elapsed > 250*time.Millisecond {
		t.Errorf("Expected ~200ms wait, got %v", elapsed)
	}
}

func TestTokenBucket_ContextCancellation(t *testing.T) {
	limiter := NewTokenBucket(1, 1)

	// Exhaust the bucket
	limiter.Allow()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := limiter.Wait(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Wait should return error when context is canceled")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}

	if elapsed > 100*time.Millisecond {
		t.Errorf("Wait should return quickly on context cancellation, took %v", elapsed)
	}
}

func TestTokenBucket_Reserve(t *testing.T) {
	limiter := NewTokenBucket(10, 3)

	// First three reserves should return 0 (immediate)
	for i := 0; i < 3; i++ {
		wait := limiter.Reserve()
		if wait != 0 {
			t.Errorf("Reserve %d should be immediate, got wait %v", i, wait)
		}
	}

	// Fourth reserve should return a wait duration
	wait := limiter.Reserve()
	if wait <= 0 {
		t.Error("Reserve should return positive wait duration when bucket is empty")
	}

	// Wait duration should be approximately 100ms (1/10 second)
	expectedWait := 100 * time.Millisecond
	if wait < 50*time.Millisecond || wait > 150*time.Millisecond {
		t.Errorf("Expected ~%v wait, got %v", expectedWait, wait)
	}
}

func TestMultiLimiter(t *testing.T) {
	ml := NewMultiLimiter()

	// Add limiters for different endpoints
	ml.AddLimiter("/orders", NewTokenBucket(5, 2))
	ml.AddLimiter("/ticker", NewTokenBucket(10, 5))

	ctx := context.Background()

	// Test /orders endpoint
	for i := 0; i < 2; i++ {
		if !ml.Allow("/orders") {
			t.Errorf("Order request %d should be allowed", i)
		}
	}

	// Third order request should block or fail
	if ml.Allow("/orders") {
		t.Error("Third order request should be denied")
	}

	// Ticker endpoint should still work (different limit)
	if !ml.Allow("/ticker") {
		t.Error("Ticker request should be allowed")
	}

	// Unknown endpoint should not be rate limited
	if err := ml.Wait(ctx, "/unknown"); err != nil {
		t.Error("Unknown endpoint should not be rate limited")
	}
}

func TestNoOpLimiter(t *testing.T) {
	limiter := NewNoOpLimiter()
	ctx := context.Background()

	// Should always allow
	for i := 0; i < 1000; i++ {
		if !limiter.Allow() {
			t.Error("NoOpLimiter should always allow")
		}

		if err := limiter.Wait(ctx); err != nil {
			t.Errorf("NoOpLimiter Wait should never error: %v", err)
		}

		if wait := limiter.Reserve(); wait != 0 {
			t.Errorf("NoOpLimiter Reserve should return 0, got %v", wait)
		}
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	limiter := NewTokenBucket(10, 5)

	// Exhaust the bucket
	for i := 0; i < 5; i++ {
		limiter.Allow()
	}

	// Bucket should be empty
	if limiter.Allow() {
		t.Error("Bucket should be empty")
	}

	// Wait for partial refill (300ms = 3 tokens at 10 req/s)
	time.Sleep(300 * time.Millisecond)

	// Should allow approximately 3 requests
	allowed := 0
	for i := 0; i < 5; i++ {
		if limiter.Allow() {
			allowed++
		}
	}

	// Allow some tolerance for timing
	if allowed < 2 || allowed > 4 {
		t.Errorf("Expected ~3 requests after partial refill, got %d", allowed)
	}
}

func BenchmarkTokenBucket_Allow(b *testing.B) {
	limiter := NewTokenBucket(1000, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow()
	}
}

func BenchmarkTokenBucket_Wait(b *testing.B) {
	limiter := NewTokenBucket(10000, 1000) // High rate to minimize blocking
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Wait(ctx)
	}
}

func BenchmarkMultiLimiter_Allow(b *testing.B) {
	ml := NewMultiLimiter()
	ml.AddLimiter("test", NewTokenBucket(10000, 1000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ml.Allow("test")
	}
}
