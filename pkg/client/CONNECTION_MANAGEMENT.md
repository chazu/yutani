# Connection Management

Comprehensive connection management features for production-ready Yutani applications.

## Overview

The connection management package provides:

- **Auto-Reconnection** - Automatic reconnection with configurable backoff strategies
- **Connection Pooling** - Manage multiple connections efficiently
- **State Preservation** - Save and restore client state across reconnections
- **Health Checks** - Monitor connection health and trigger reconnections
- **Connection Events** - React to connection state changes

## Auto-Reconnection

### Basic Usage

```go
// Connect with automatic retry
opts := client.DefaultRetryOptions()
c, err := client.ConnectWithRetry("localhost:7755", opts)
if err != nil {
    log.Fatal(err)
}
defer c.Close()
```

### Retry Options

```go
opts := client.RetryOptions{
    MaxRetries:      5,                           // Max retry attempts (0 = infinite)
    InitialDelay:    1 * time.Second,             // Initial delay
    MaxDelay:        30 * time.Second,            // Maximum delay
    BackoffStrategy: client.ExponentialBackoff,   // Backoff strategy
    Multiplier:      2.0,                         // Backoff multiplier
}

c, err := client.ConnectWithRetry("localhost:7755", opts)
```

### Backoff Strategies

#### Constant Backoff
Always uses the same delay between retries.

```go
opts := client.RetryOptions{
    InitialDelay:    2 * time.Second,
    BackoffStrategy: client.ConstantBackoff,
}
```

#### Linear Backoff
Increases delay linearly: `delay = InitialDelay * Multiplier * attempt`

```go
opts := client.RetryOptions{
    InitialDelay:    1 * time.Second,
    BackoffStrategy: client.LinearBackoff,
    Multiplier:      1.5,
}
// Delays: 1.5s, 3s, 4.5s, 6s, ...
```

#### Exponential Backoff (Recommended)
Doubles delay with each retry: `delay = InitialDelay * Multiplier^(attempt-1)`

```go
opts := client.RetryOptions{
    InitialDelay:    1 * time.Second,
    BackoffStrategy: client.ExponentialBackoff,
    Multiplier:      2.0,
}
// Delays: 1s, 2s, 4s, 8s, 16s, ...
```

### Retry Callbacks

```go
opts := client.RetryOptions{
    MaxRetries:   5,
    InitialDelay: 1 * time.Second,
    
    OnReconnecting: func(attempt int, delay time.Duration) {
        log.Printf("Reconnecting (attempt %d) in %v...", attempt, delay)
    },
    
    OnReconnected: func(attempt int) {
        log.Printf("Reconnected successfully after %d attempts", attempt)
    },
    
    OnReconnectFailed: func(attempt int, err error) {
        log.Printf("Reconnection failed after %d attempts: %v", attempt, err)
    },
}
```

### Manual Reconnection

```go
// Manually trigger reconnection
if err := c.Reconnect(); err != nil {
    log.Printf("Reconnection failed: %v", err)
}
```

## Connection State

### Checking Connection State

```go
// Get current state
state := c.GetConnectionState()
fmt.Println(state) // "Connected", "Disconnected", or "Reconnecting"

// Check if connected
if c.IsConnected() {
    // Connection is active
}

// Check connection health
if c.IsHealthy() {
    // Connection is healthy
}
```

### Connection State Callbacks

```go
c.OnConnectionStateChange(func(oldState, newState client.ConnectionState) {
    log.Printf("Connection state changed: %s -> %s", oldState, newState)
    
    switch newState {
    case client.StateConnected:
        log.Println("Connected to server")
    case client.StateDisconnected:
        log.Println("Disconnected from server")
    case client.StateReconnecting:
        log.Println("Attempting to reconnect...")
    }
})
```

## Health Checks

### Periodic Health Checks

```go
// Start health checks every 30 seconds
c.StartHealthCheck(30 * time.Second)

// Health checks will automatically trigger reconnection if needed
```

### Manual Health Check

```go
if !c.IsHealthy() {
    log.Println("Connection unhealthy, reconnecting...")
    c.Reconnect()
}
```

### Stop Health Checks

```go
c.StopHealthCheck()
```

## Connection Pooling

### Creating a Pool

```go
opts := client.DefaultPoolOptions()
pool, err := client.NewConnectionPool("localhost:7755", opts)
if err != nil {
    log.Fatal(err)
}
defer pool.Close()
```

### Pool Options

```go
opts := client.PoolOptions{
    MinConnections:      2,                    // Minimum connections
    MaxConnections:      10,                   // Maximum connections
    IdleTimeout:         5 * time.Minute,      // Idle connection timeout
    MaxLifetime:         30 * time.Minute,     // Max connection lifetime
    HealthCheckInterval: 30 * time.Second,     // Health check interval
    RetryOptions:        client.DefaultRetryOptions(),
}
```

### Using the Pool

```go
// Get a connection from the pool
c, err := pool.Get()
if err != nil {
    log.Fatal(err)
}

// Use the connection
list, _ := c.NewList().Title("My List").Build()

// Return connection to pool when done
pool.Put(c)
```

### Pool Statistics

```go
stats := pool.Stats()
fmt.Printf("Total: %d, Available: %d, In Use: %d, Idle: %d\n",
    stats.TotalConnections,
    stats.AvailableConnections,
    stats.InUseConnections,
    stats.IdleConnections)
```

## State Preservation

### Saving State

```go
// Save current client state
state, err := c.SaveState()
if err != nil {
    log.Fatal(err)
}

// Serialize to JSON
data, err := c.SerializeState()
if err != nil {
    log.Fatal(err)
}

// Save to file
os.WriteFile("client-state.json", data, 0644)
```

### Restoring State

```go
// Load from file
data, err := os.ReadFile("client-state.json")
if err != nil {
    log.Fatal(err)
}

// Deserialize and restore
if err := c.DeserializeState(data); err != nil {
    log.Fatal(err)
}
```

**Note**: State preservation is currently a framework. Full widget state restoration requires widget-specific serialization methods.

## Complete Examples

### Example 1: Resilient Client

```go
func main() {
    // Configure retry options
    retryOpts := client.RetryOptions{
        MaxRetries:      0, // Infinite retries
        InitialDelay:    1 * time.Second,
        MaxDelay:        30 * time.Second,
        BackoffStrategy: client.ExponentialBackoff,
        Multiplier:      2.0,
        
        OnReconnecting: func(attempt int, delay time.Duration) {
            log.Printf("Reconnecting in %v (attempt %d)...", delay, attempt)
        },
        
        OnReconnected: func(attempt int) {
            log.Printf("Reconnected after %d attempts", attempt)
        },
    }
    
    // Connect with retry
    c, err := client.ConnectWithRetry("localhost:7755", retryOpts)
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()
    
    // Start health checks
    c.StartHealthCheck(30 * time.Second)
    
    // Monitor connection state
    c.OnConnectionStateChange(func(old, new client.ConnectionState) {
        log.Printf("State: %s -> %s", old, new)
    })
    
    // Use client normally
    list, _ := c.NewList().Title("My List").Build()
    c.SetRoot(list)
    c.StartEventStream()
    
    select {}
}
```

### Example 2: Connection Pool

```go
func main() {
    // Create connection pool
    poolOpts := client.PoolOptions{
        MinConnections:      3,
        MaxConnections:      10,
        IdleTimeout:         5 * time.Minute,
        HealthCheckInterval: 30 * time.Second,
    }
    
    pool, err := client.NewConnectionPool("localhost:7755", poolOpts)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()
    
    // Use pool in goroutines
    for i := 0; i < 5; i++ {
        go func(id int) {
            c, err := pool.Get()
            if err != nil {
                log.Printf("Worker %d: failed to get connection: %v", id, err)
                return
            }
            defer pool.Put(c)
            
            // Use connection
            list, _ := c.NewList().Title(fmt.Sprintf("List %d", id)).Build()
            list.AddItem("Item 1", "", nil)
        }(i)
    }
    
    // Monitor pool stats
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        stats := pool.Stats()
        log.Printf("Pool: %d total, %d available, %d in use",
            stats.TotalConnections,
            stats.AvailableConnections,
            stats.InUseConnections)
    }
}
```

## Best Practices

1. **Use Exponential Backoff** - Prevents overwhelming the server
2. **Set Max Delay** - Prevents excessively long waits
3. **Enable Health Checks** - Detect connection issues early
4. **Monitor State Changes** - React to connection events
5. **Use Connection Pooling** - For high-concurrency applications
6. **Handle Reconnection** - Design for temporary disconnections
7. **Save State** - Preserve important data across reconnections

## See Also

- [Client Library Documentation](README.md)
- [Testing Utilities](testing/README.md)
- [Examples](../../examples/README.md)

