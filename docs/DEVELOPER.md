# Developer Guide

Complete guide for developers working on the scalping bot.

## Architecture Overview

### Layer Architecture

```
┌─────────────────────────────────────────┐
│         Terminal UI (TUI)               │
│     (Bubble Tea Framework)              │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Application Layer               │
│   • Strategy Engine                     │
│   • Order Manager                       │
│   • Risk Manager                        │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Exchange Layer                  │
│   • Hyperliquid • Coinbase • dYdX      │
│   • REST API    • WebSocket            │
└─────────────────────────────────────────┘
```

## Code Organization

### Package Structure

#### `cmd/bot/`
Entry point of the application. Contains:
- Application initialization
- Component wiring
- Callback setup
- Signal handling

#### `internal/exchanges/`
Exchange integration layer:
- `interface.go`: Common interface for all exchanges
- Exchange-specific implementations (hyperliquid, coinbase, dydx)
- WebSocket connection management
- REST API wrappers

#### `internal/strategy/`
Trading strategy implementation:
- `scalping.go`: Main strategy logic
- `indicators.go`: Technical indicators
- `signals.go`: Signal generation and evaluation

#### `internal/order/`
Order lifecycle management:
- `manager.go`: Order placement, tracking, cancellation
- `types.go`: Order and position data structures

#### `internal/risk/`
Risk management:
- `manager.go`: Risk calculations and limits enforcement

#### `internal/tui/`
Terminal user interface:
- `model.go`: TUI state management
- `update.go`: Event handling
- `view.go`: Rendering logic
- `components/`: Reusable UI components

#### `pkg/utils/`
Shared utilities:
- Mathematical functions
- Helper utilities

## Key Interfaces

### Exchange Interface

```go
type Exchange interface {
    // Connection
    Connect(ctx context.Context) error
    Disconnect() error
    IsConnected() bool

    // Market Data
    GetTicker(ctx context.Context, symbol string) (*Ticker, error)
    GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error)
    GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]Candle, error)

    // Subscriptions
    SubscribeTicker(ctx context.Context, symbol string, callback func(*Ticker)) error
    SubscribeOrderBook(ctx context.Context, symbol string, callback func(*OrderBook)) error
    SubscribeTrades(ctx context.Context, symbol string, callback func(*Trade)) error

    // Trading
    PlaceOrder(ctx context.Context, order *Order) (*Order, error)
    CancelOrder(ctx context.Context, orderID string) error
    GetOrder(ctx context.Context, orderID string) (*Order, error)
    GetOpenOrders(ctx context.Context, symbol string) ([]Order, error)

    // Account
    GetBalance(ctx context.Context) ([]Balance, error)
    GetPositions(ctx context.Context) ([]Position, error)
    GetPosition(ctx context.Context, symbol string) (*Position, error)

    // Metadata
    Name() string
    SupportedSymbols() []string
}
```

## Adding a New Exchange

### Step 1: Create Package Structure

```bash
mkdir -p internal/exchanges/newexchange
touch internal/exchanges/newexchange/client.go
touch internal/exchanges/newexchange/websocket.go
```

### Step 2: Implement Client

```go
package newexchange

import (
    "context"
    "github.com/guyghost/constantine/internal/exchanges"
)

type Client struct {
    apiKey    string
    apiSecret string
    baseURL   string
    wsURL     string
    // ... other fields
}

func NewClient(apiKey, apiSecret string) *Client {
    return &Client{
        apiKey:    apiKey,
        apiSecret: apiSecret,
        baseURL:   "https://api.newexchange.com",
        wsURL:     "wss://ws.newexchange.com",
    }
}

// Implement all Exchange interface methods
func (c *Client) Connect(ctx context.Context) error {
    // Implementation
}

// ... implement all other methods
```

### Step 3: Implement WebSocket

```go
package newexchange

import (
    "context"
    "github.com/gorilla/websocket"
)

type WebSocketClient struct {
    conn *websocket.Conn
    // ... callbacks
}

func (ws *WebSocketClient) Connect(ctx context.Context) error {
    // Implementation
}

func (ws *WebSocketClient) handleMessages() {
    // Message processing loop
}
```

### Step 4: Register in Main

```go
// In cmd/bot/main.go
switch exchangeType {
case "newexchange":
    exchange = newexchange.NewClient(apiKey, apiSecret)
// ...
}
```

## Adding a New Indicator

### Step 1: Implement Function

```go
// In internal/strategy/indicators.go

// MyNewIndicator calculates a custom indicator
func MyNewIndicator(prices []decimal.Decimal, period int) []decimal.Decimal {
    if len(prices) < period {
        return []decimal.Decimal{}
    }

    result := make([]decimal.Decimal, len(prices)-period+1)

    for i := 0; i <= len(prices)-period; i++ {
        // Your calculation logic
        value := calculateValue(prices[i:i+period])
        result[i] = value
    }

    return result
}
```

### Step 2: Use in Strategy

```go
// In internal/strategy/signals.go

func (sg *SignalGenerator) GenerateSignal(...) *Signal {
    // Calculate your indicator
    myIndicator := MyNewIndicator(prices, period)

    // Use in signal generation
    if /* condition based on myIndicator */ {
        return &Signal{...}
    }
}
```

## Adding a New Strategy

### Step 1: Create Strategy File

```go
// In internal/strategy/mystrategy.go

package strategy

type MyStrategy struct {
    config   *MyStrategyConfig
    exchange exchanges.Exchange
    // ... fields
}

func NewMyStrategy(config *MyStrategyConfig, exchange exchanges.Exchange) *MyStrategy {
    return &MyStrategy{
        config:   config,
        exchange: exchange,
    }
}

func (s *MyStrategy) Start(ctx context.Context) error {
    // Initialize and start strategy
}

func (s *MyStrategy) Stop() error {
    // Cleanup
}
```

### Step 2: Implement Signal Generation

```go
func (s *MyStrategy) generateSignal(data MarketData) *Signal {
    // Your strategy logic
}
```

### Step 3: Wire in Main

```go
// In cmd/bot/main.go

// Choose strategy based on config
strategy := mystrategy.NewMyStrategy(config, exchange)
```

## Testing Guide

### Unit Tests

```go
// Example: internal/strategy/indicators_test.go

package strategy

import (
    "testing"
    "github.com/shopspring/decimal"
)

func TestEMA(t *testing.T) {
    prices := []decimal.Decimal{
        decimal.NewFromFloat(100),
        decimal.NewFromFloat(101),
        decimal.NewFromFloat(102),
        // ... more prices
    }

    result := EMA(prices, 5)

    if len(result) == 0 {
        t.Error("Expected non-empty result")
    }

    // Add more assertions
}
```

### Integration Tests

```go
// Example: internal/exchanges/hyperliquid/client_test.go

package hyperliquid

import (
    "context"
    "testing"
)

func TestConnect(t *testing.T) {
    client := NewClient("test_key", "test_secret")

    ctx := context.Background()
    err := client.Connect(ctx)

    if err != nil {
        t.Errorf("Connect failed: %v", err)
    }

    defer client.Disconnect()
}
```

### Mock Exchange

```go
// For testing without real exchange

type MockExchange struct {
    // Mock implementation
}

func (m *MockExchange) Connect(ctx context.Context) error {
    return nil // Simulate success
}

// Implement other methods as mocks
```

## Debugging Tips

### Enable Verbose Logging

```go
// Add logging throughout the code
import "log"

log.Printf("Signal generated: %+v", signal)
log.Printf("Order placed: %s", orderID)
log.Printf("Position updated: %+v", position)
```

### TUI Debugging

The TUI can make debugging difficult. For debugging:

```go
// Option 1: Write to file
f, _ := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
log.SetOutput(f)

// Option 2: Run without TUI
// Comment out TUI initialization in main.go
```

### WebSocket Debugging

```go
// Log all WebSocket messages
func (ws *WebSocketClient) processMessage(message []byte) {
    log.Printf("WS Message: %s", string(message))
    // ... rest of processing
}
```

## Performance Optimization

### Reduce Memory Allocations

```go
// Bad: Creates new slice every time
prices = append(prices, newPrice)

// Good: Pre-allocate with capacity
prices := make([]decimal.Decimal, 0, 1000)
```

### Use Buffered Channels

```go
// Bad: Unbuffered channel can block
ch := make(chan Signal)

// Good: Buffered channel
ch := make(chan Signal, 100)
```

### Pool Frequently Used Objects

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

buf := bufferPool.Get().(*bytes.Buffer)
defer bufferPool.Put(buf)
```

## Common Patterns

### Callback Pattern

```go
type Component struct {
    onEvent func(Event)
}

func (c *Component) SetEventCallback(cb func(Event)) {
    c.onEvent = cb
}

func (c *Component) emitEvent(event Event) {
    if c.onEvent != nil {
        c.onEvent(event)
    }
}
```

### Context-Based Cancellation

```go
func (c *Component) Start(ctx context.Context) error {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                // Do work
            }
        }
    }()
}
```

### Safe Concurrent Access

```go
type SafeData struct {
    mu   sync.RWMutex
    data map[string]Value
}

func (s *SafeData) Get(key string) Value {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.data[key]
}

func (s *SafeData) Set(key string, value Value) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
}
```

## Code Style Guidelines

### Naming Conventions

- Packages: lowercase, single word
- Interfaces: noun or noun phrase
- Functions: verb or verb phrase
- Constants: CamelCase or UPPER_CASE
- Private: start with lowercase
- Public: start with uppercase

### Documentation

```go
// Component does X and Y.
// It is used for Z purpose.
//
// Example:
//
//     comp := NewComponent(config)
//     comp.Start(ctx)
//
type Component struct {
    // ...
}

// NewComponent creates a new Component with the given config.
// Returns an error if config is invalid.
func NewComponent(config *Config) (*Component, error) {
    // ...
}
```

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}

// Check specific errors
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

// Type assertions for custom errors
if apiErr, ok := err.(*APIError); ok {
    log.Printf("API error code: %d", apiErr.Code)
}
```

## Git Workflow

### Branch Naming

- `feature/new-exchange-binance`
- `bugfix/websocket-reconnect`
- `hotfix/critical-order-bug`
- `refactor/order-manager`

### Commit Messages

```
feat: add Binance exchange support

- Implement REST API client
- Add WebSocket connection
- Update exchange factory

Closes #123
```

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added
- [ ] Integration tests added
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
```

## Deployment

### Building for Production

```bash
# Build with optimizations
go build -ldflags="-s -w" -o docs cmd/bot/main.go

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o docs-linux cmd/bot/main.go
```

### Configuration Management

```bash
# Use environment-specific configs
cp .env.production .env

# Or use command-line flags
./docs -config=production.yaml
```

### Monitoring

```go
// Add metrics collection
import "github.com/prometheus/client_golang/prometheus"

var (
    ordersPlaced = prometheus.NewCounter(...)
    positionsOpen = prometheus.NewGauge(...)
)

// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

## Resources

### Documentation
- Go documentation: https://golang.org/doc/
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- WebSocket: https://github.com/gorilla/websocket
- Decimal: https://github.com/shopspring/decimal

### Tools
- golangci-lint: Code linting
- delve: Debugging
- pprof: Profiling

### Learning Resources
- Effective Go
- Go Code Review Comments
- The Go Blog

## Getting Help

- Check existing issues
- Read code comments
- Review test cases
- Ask in discussions

## Contributing

See CONTRIBUTING.md (to be created) for:
- Code style guide
- Pull request process
- Issue reporting
- Feature requests

---

Happy coding! 🚀
