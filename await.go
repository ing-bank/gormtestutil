package gormtestutil

import (
	"sync"
	"testing"
	"time"
)

const (
	// defaultTimeout is the default value for EnsureCompletion's config
	defaultTimeout = 30 * time.Second
)

// EnsureOption allows various options to be supplied to EnsureCompletion
type EnsureOption func(*ensureConfig)

// WithTimeout is used to set a timeout for EnsureCompletion
func WithTimeout(timeout time.Duration) EnsureOption {
	return func(config *ensureConfig) {
		config.timeout = timeout
	}
}

type ensureConfig struct {
	timeout time.Duration
}

// EnsureCompletion ensures that the waitgroup completes within a specified duration or else fails
func EnsureCompletion(t *testing.T, wg *sync.WaitGroup, options ...EnsureOption) bool {
	t.Helper()

	if wg == nil {
		t.Error("WithExpectation is nil")

		return false
	}

	config := &ensureConfig{
		timeout: defaultTimeout,
	}

	for _, option := range options {
		option(config)
	}

	// Run waitgroup in goroutine
	channel := make(chan struct{})
	go func() {
		t.Helper()
		defer close(channel)
		wg.Wait()
	}()

	// Select first response (waitgroup completion or time.After)
	select {
	case <-channel:
		return true
	case <-time.After(config.timeout):
		t.Errorf("tasks did not complete within: %v", config.timeout)

		return false
	}
}
