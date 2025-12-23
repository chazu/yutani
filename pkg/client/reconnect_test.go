package client

import (
	"testing"
	"time"
)

func TestRetryOptions_calculateDelay(t *testing.T) {
	tests := []struct {
		name     string
		opts     RetryOptions
		attempt  int
		expected time.Duration
	}{
		{
			name: "constant backoff",
			opts: RetryOptions{
				InitialDelay:    1 * time.Second,
				MaxDelay:        10 * time.Second,
				BackoffStrategy: ConstantBackoff,
			},
			attempt:  3,
			expected: 1 * time.Second,
		},
		{
			name: "linear backoff",
			opts: RetryOptions{
				InitialDelay:    1 * time.Second,
				MaxDelay:        10 * time.Second,
				BackoffStrategy: LinearBackoff,
				Multiplier:      1.0,
			},
			attempt:  3,
			expected: 3 * time.Second,
		},
		{
			name: "exponential backoff",
			opts: RetryOptions{
				InitialDelay:    1 * time.Second,
				MaxDelay:        10 * time.Second,
				BackoffStrategy: ExponentialBackoff,
				Multiplier:      2.0,
			},
			attempt:  3,
			expected: 4 * time.Second, // 1 * 2^(3-1) = 4
		},
		{
			name: "exponential backoff capped at max",
			opts: RetryOptions{
				InitialDelay:    1 * time.Second,
				MaxDelay:        5 * time.Second,
				BackoffStrategy: ExponentialBackoff,
				Multiplier:      2.0,
			},
			attempt:  10,
			expected: 5 * time.Second, // Capped at MaxDelay
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := tt.opts.calculateDelay(tt.attempt)
			if delay != tt.expected {
				t.Errorf("calculateDelay() = %v, want %v", delay, tt.expected)
			}
		})
	}
}

func TestDefaultRetryOptions(t *testing.T) {
	opts := DefaultRetryOptions()

	if opts.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries = 5, got %d", opts.MaxRetries)
	}

	if opts.InitialDelay != 1*time.Second {
		t.Errorf("Expected InitialDelay = 1s, got %v", opts.InitialDelay)
	}

	if opts.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay = 30s, got %v", opts.MaxDelay)
	}

	if opts.BackoffStrategy != ExponentialBackoff {
		t.Errorf("Expected BackoffStrategy = ExponentialBackoff, got %v", opts.BackoffStrategy)
	}

	if opts.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier = 2.0, got %f", opts.Multiplier)
	}

	if opts.JitterFactor != 0.2 {
		t.Errorf("Expected JitterFactor = 0.2, got %f", opts.JitterFactor)
	}
}

func TestRetryOptions_calculateDelay_withJitter(t *testing.T) {
	opts := RetryOptions{
		InitialDelay:    1 * time.Second,
		MaxDelay:        10 * time.Second,
		BackoffStrategy: ConstantBackoff,
		JitterFactor:    0.2, // 20% jitter means delay in [0.8s, 1.2s]
	}

	// Run multiple times to verify jitter produces variation
	minDelay := 800 * time.Millisecond  // 1s * 0.8
	maxDelay := 1200 * time.Millisecond // 1s * 1.2

	var sawVariation bool
	var lastDelay time.Duration

	for i := 0; i < 100; i++ {
		delay := opts.calculateDelay(1)

		// Verify delay is within expected bounds
		if delay < minDelay || delay > maxDelay {
			t.Errorf("Delay %v outside expected range [%v, %v]", delay, minDelay, maxDelay)
		}

		// Check for variation
		if lastDelay != 0 && delay != lastDelay {
			sawVariation = true
		}
		lastDelay = delay
	}

	if !sawVariation {
		t.Error("Expected jitter to produce variation in delays")
	}
}

func TestConnectionState_String(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateConnected, "Connected"},
		{StateDisconnected, "Disconnected"},
		{StateReconnecting, "Reconnecting"},
		{ConnectionState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClient_GetConnectionState(t *testing.T) {
	// This would require a mock client or test server
	// For now, we'll test the state transitions
	c := &Client{
		connectionState: StateConnected,
	}

	if state := c.GetConnectionState(); state != StateConnected {
		t.Errorf("Expected StateConnected, got %v", state)
	}

	if !c.IsConnected() {
		t.Error("Expected IsConnected() to be true")
	}

	c.connectionState = StateDisconnected

	if c.IsConnected() {
		t.Error("Expected IsConnected() to be false")
	}
}

func TestClient_OnConnectionStateChange(t *testing.T) {
	c := &Client{
		connectionState: StateConnected,
		stateCallbacks:  make([]ConnectionStateCallback, 0),
	}

	callbackCalled := false
	var oldState, newState ConnectionState

	c.OnConnectionStateChange(func(old, new ConnectionState) {
		callbackCalled = true
		oldState = old
		newState = new
	})

	// Trigger state change
	c.setConnectionState(StateDisconnected)

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}

	if oldState != StateConnected {
		t.Errorf("Expected oldState = StateConnected, got %v", oldState)
	}

	if newState != StateDisconnected {
		t.Errorf("Expected newState = StateDisconnected, got %v", newState)
	}
}

