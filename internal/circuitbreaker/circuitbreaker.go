package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/logger"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota // Normal operation
	StateOpen                // Failing, reject requests
	StateHalfOpen            // Testing if service recovered
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when too many requests in half-open state
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// Config holds circuit breaker configuration
type Config struct {
	// MaxFailures is the number of failures before opening the circuit
	MaxFailures uint32
	// Timeout is the duration circuit stays open before entering half-open
	Timeout time.Duration
	// MaxHalfOpenRequests is max concurrent requests allowed in half-open state
	MaxHalfOpenRequests uint32
	// OnStateChange is called when state changes
	OnStateChange func(from, to State)
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig() *Config {
	return &Config{
		MaxFailures:         5,
		Timeout:             60 * time.Second,
		MaxHalfOpenRequests: 1,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name   string
	config *Config

	mu               sync.Mutex
	state            State
	failures         uint32
	lastFailureTime  time.Time
	lastStateChange  time.Time
	halfOpenRequests uint32

	log *logger.Logger
}

// New creates a new circuit breaker
func New(name string, config *Config) *CircuitBreaker {
	if config == nil {
		config = DefaultConfig()
	}

	return &CircuitBreaker{
		name:            name,
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
		log:             logger.Component("circuit-breaker").WithField("breaker", name),
	}
}

// Execute runs the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can execute
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record result
	cb.afterRequest(err)

	return err
}

// beforeRequest checks if request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Normal operation
		return nil

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailureTime) > cb.config.Timeout {
			cb.setState(StateHalfOpen)
			cb.log.Info("Circuit breaker transitioning to half-open",
				"timeout_elapsed", cb.config.Timeout)
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Limit concurrent requests in half-open state
		if cb.halfOpenRequests >= cb.config.MaxHalfOpenRequests {
			return ErrTooManyRequests
		}
		cb.halfOpenRequests++
		return nil

	default:
		return fmt.Errorf("unknown circuit breaker state: %v", cb.state)
	}
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		// Success
		cb.onSuccess()
	} else {
		// Failure
		cb.onFailure()
	}
}

// onSuccess handles successful requests
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failures = 0

	case StateHalfOpen:
		cb.halfOpenRequests--
		// Successful request in half-open state closes the circuit
		cb.setState(StateClosed)
		cb.failures = 0
		cb.log.Info("Circuit breaker closed after successful test")
	}
}

// onFailure handles failed requests
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.MaxFailures {
			cb.setState(StateOpen)
			cb.log.Warn("Circuit breaker opened due to failures",
				"failures", cb.failures,
				"max_failures", cb.config.MaxFailures)
		}

	case StateHalfOpen:
		cb.halfOpenRequests--
		// Failure in half-open state reopens the circuit
		cb.setState(StateOpen)
		cb.log.Warn("Circuit breaker reopened after failed test")
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	if cb.config.OnStateChange != nil {
		cb.config.OnStateChange(oldState, newState)
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Failures returns the current failure count
func (cb *CircuitBreaker) Failures() uint32 {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failures
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed)
	cb.failures = 0
	cb.halfOpenRequests = 0
	cb.log.Info("Circuit breaker manually reset")
}

// Stats returns circuit breaker statistics
type Stats struct {
	Name            string
	State           State
	Failures        uint32
	LastFailure     time.Time
	LastStateChange time.Time
}

// Stats returns current statistics
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return Stats{
		Name:            cb.name,
		State:           cb.state,
		Failures:        cb.failures,
		LastFailure:     cb.lastFailureTime,
		LastStateChange: cb.lastStateChange,
	}
}
