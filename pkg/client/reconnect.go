package client

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// BackoffStrategy defines how to calculate retry delays.
type BackoffStrategy int

const (
	// ConstantBackoff uses a constant delay between retries.
	ConstantBackoff BackoffStrategy = iota
	// LinearBackoff increases delay linearly with each retry.
	LinearBackoff
	// ExponentialBackoff doubles delay with each retry.
	ExponentialBackoff
)

// RetryOptions configures reconnection behavior.
type RetryOptions struct {
	// MaxRetries is the maximum number of retry attempts (0 = infinite).
	MaxRetries int

	// InitialDelay is the initial delay before first retry.
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration

	// BackoffStrategy determines how delays increase.
	BackoffStrategy BackoffStrategy

	// Multiplier for exponential/linear backoff (default: 2.0).
	Multiplier float64

	// JitterFactor adds randomness to delays to prevent thundering herd.
	// A value of 0.2 means delays vary by ±20% (e.g., 1s becomes 0.8-1.2s).
	// Default: 0.2 (20% jitter).
	JitterFactor float64

	// OnReconnecting is called when reconnection starts.
	OnReconnecting func(attempt int, delay time.Duration)

	// OnReconnected is called when reconnection succeeds.
	OnReconnected func(attempt int)

	// OnReconnectFailed is called when reconnection fails.
	OnReconnectFailed func(attempt int, err error)
}

// DefaultRetryOptions returns sensible default retry options.
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxRetries:      5,
		InitialDelay:    1 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffStrategy: ExponentialBackoff,
		Multiplier:      2.0,
		JitterFactor:    0.2, // 20% jitter
	}
}

// calculateDelay calculates the delay for a given retry attempt.
func (r *RetryOptions) calculateDelay(attempt int) time.Duration {
	var delay time.Duration

	switch r.BackoffStrategy {
	case ConstantBackoff:
		delay = r.InitialDelay

	case LinearBackoff:
		multiplier := r.Multiplier
		if multiplier == 0 {
			multiplier = 1.0
		}
		delay = time.Duration(float64(r.InitialDelay) * multiplier * float64(attempt))

	case ExponentialBackoff:
		multiplier := r.Multiplier
		if multiplier == 0 {
			multiplier = 2.0
		}
		delay = time.Duration(float64(r.InitialDelay) * math.Pow(multiplier, float64(attempt-1)))
	}

	// Cap at max delay
	if delay > r.MaxDelay {
		delay = r.MaxDelay
	}

	// Apply jitter to prevent thundering herd
	if r.JitterFactor > 0 {
		// jitter is in range [-jitterFactor, +jitterFactor]
		jitter := (rand.Float64()*2 - 1) * r.JitterFactor
		delay = time.Duration(float64(delay) * (1 + jitter))
	}

	return delay
}

// ConnectWithRetry creates a client connection with automatic retry.
func ConnectWithRetry(address string, opts RetryOptions) (*Client, error) {
	return ConnectWithRetryAndOptions(address, opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// ConnectWithRetryAndOptions creates a client connection with retry and custom gRPC options.
func ConnectWithRetryAndOptions(address string, retryOpts RetryOptions, grpcOpts ...grpc.DialOption) (*Client, error) {
	var lastErr error
	
	// Set defaults
	if retryOpts.InitialDelay == 0 {
		retryOpts.InitialDelay = 1 * time.Second
	}
	if retryOpts.MaxDelay == 0 {
		retryOpts.MaxDelay = 30 * time.Second
	}
	if retryOpts.Multiplier == 0 {
		retryOpts.Multiplier = 2.0
	}
	
	for attempt := 1; retryOpts.MaxRetries == 0 || attempt <= retryOpts.MaxRetries; attempt++ {
		// Try to connect
		client, err := ConnectWithOptions(address, grpcOpts...)
		if err == nil {
			// Success!
			if retryOpts.OnReconnected != nil {
				retryOpts.OnReconnected(attempt)
			}
			
			// Enable auto-reconnect on this client
			client.enableAutoReconnect(address, retryOpts, grpcOpts)
			
			return client, nil
		}
		
		lastErr = err
		
		// Failed - check if we should retry
		if retryOpts.MaxRetries > 0 && attempt >= retryOpts.MaxRetries {
			break
		}
		
		// Calculate delay
		delay := retryOpts.calculateDelay(attempt)
		
		// Notify
		if retryOpts.OnReconnecting != nil {
			retryOpts.OnReconnecting(attempt, delay)
		}
		
		// Wait before retry
		time.Sleep(delay)
	}
	
	// All retries failed
	if retryOpts.OnReconnectFailed != nil {
		retryOpts.OnReconnectFailed(retryOpts.MaxRetries, lastErr)
	}
	
	return nil, fmt.Errorf("failed to connect after %d attempts: %w", retryOpts.MaxRetries, lastErr)
}

// enableAutoReconnect enables automatic reconnection on connection loss.
func (c *Client) enableAutoReconnect(address string, retryOpts RetryOptions, grpcOpts []grpc.DialOption) {
	c.mu.Lock()
	c.autoReconnect = true
	c.reconnectAddress = address
	c.reconnectOpts = retryOpts
	c.reconnectGrpcOpts = grpcOpts
	c.mu.Unlock()
}

