package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	metricsMu        sync.RWMutex
	orderCounts      = make(map[string]map[string]uint64)
	stopLossCounts   = make(map[string]uint64)
	takeProfitCounts = make(map[string]uint64)
	callbackPanics   uint64

	// New metrics for enhanced monitoring
	balanceUpdates      = make(map[string]float64)
	positionUpdates     = make(map[string]map[string]float64) // symbol -> field -> value
	pnlUpdates          = make(map[string]float64)
	signalCounts        = make(map[string]uint64)                     // signal type counters
	errorCounts         = make(map[string]uint64)                     // error type counters
	websocketReconnects = make(map[string]uint64)                     // exchange -> reconnect count
	apiRequestCounts    = make(map[string]map[string]uint64)          // exchange -> endpoint -> count
	apiRequestLatency   = make(map[string]map[string][]time.Duration) // exchange -> endpoint -> latencies
)

// RecordOrderPlaced increments the order placed counter.
func RecordOrderPlaced(symbol, side string) {
	if symbol == "" {
		symbol = "unknown"
	}
	if side == "" {
		side = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()

	if _, exists := orderCounts[symbol]; !exists {
		orderCounts[symbol] = make(map[string]uint64)
	}
	orderCounts[symbol][side]++
}

// RecordStopLossPlaced increments the stop loss counter.
func RecordStopLossPlaced(symbol string) {
	if symbol == "" {
		symbol = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	stopLossCounts[symbol]++
}

// RecordTakeProfitPlaced increments the take profit counter.
func RecordTakeProfitPlaced(symbol string) {
	if symbol == "" {
		symbol = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	takeProfitCounts[symbol]++
}

// RecordCallbackPanic records a recovered panic in callbacks.
func RecordCallbackPanic() {
	atomic.AddUint64(&callbackPanics, 1)
}

// RecordBalanceUpdate records account balance updates.
func RecordBalanceUpdate(asset string, amount float64) {
	if asset == "" {
		asset = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	balanceUpdates[asset] = amount
}

// RecordPositionUpdate records position updates.
func RecordPositionUpdate(symbol, field string, value float64) {
	if symbol == "" {
		symbol = "unknown"
	}
	if field == "" {
		field = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()

	if _, exists := positionUpdates[symbol]; !exists {
		positionUpdates[symbol] = make(map[string]float64)
	}
	positionUpdates[symbol][field] = value
}

// RecordPnLUpdate records P&L updates.
func RecordPnLUpdate(symbol string, pnl float64) {
	if symbol == "" {
		symbol = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	pnlUpdates[symbol] = pnl
}

// RecordSignal records trading signals.
func RecordSignal(signalType string) {
	if signalType == "" {
		signalType = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	signalCounts[signalType]++
}

// RecordError records errors by type.
func RecordError(errorType string) {
	if errorType == "" {
		errorType = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	errorCounts[errorType]++
}

// RecordWebSocketReconnect records WebSocket reconnection events.
func RecordWebSocketReconnect(exchange string) {
	if exchange == "" {
		exchange = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	websocketReconnects[exchange]++
}

// RecordAPIRequest records API request metrics.
func RecordAPIRequest(exchange, endpoint string, latency time.Duration) {
	if exchange == "" {
		exchange = "unknown"
	}
	if endpoint == "" {
		endpoint = "unknown"
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()

	// Record count
	if _, exists := apiRequestCounts[exchange]; !exists {
		apiRequestCounts[exchange] = make(map[string]uint64)
	}
	apiRequestCounts[exchange][endpoint]++

	// Record latency (keep last 100 samples)
	if _, exists := apiRequestLatency[exchange]; !exists {
		apiRequestLatency[exchange] = make(map[string][]time.Duration)
	}
	latencies := apiRequestLatency[exchange][endpoint]
	if len(latencies) >= 100 {
		// Remove oldest sample
		latencies = latencies[1:]
	}
	latencies = append(latencies, latency)
	apiRequestLatency[exchange][endpoint] = latencies
}

// Server exposes metrics and health endpoints.
type Server struct {
	srv        *http.Server
	readyState atomic.Bool
}

// NewServer creates a new telemetry server.
func NewServer(addr string) *Server {
	if addr == "" {
		return nil
	}

	server := &Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", server.metricsHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if server.readyState.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
	})

	server.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return server
}

func (s *Server) metricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	builder := &strings.Builder{}
	builder.WriteString("# HELP constantine_orders_total Total number of orders placed\n")
	builder.WriteString("# TYPE constantine_orders_total counter\n")

	metricsMu.RLock()
	symbols := make([]string, 0, len(orderCounts))
	for symbol := range orderCounts {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		sides := orderCounts[symbol]
		sideKeys := make([]string, 0, len(sides))
		for side := range sides {
			sideKeys = append(sideKeys, side)
		}
		sort.Strings(sideKeys)
		for _, side := range sideKeys {
			fmt.Fprintf(builder, "constantine_orders_total{symbol=\"%s\",side=\"%s\"} %d\n", symbol, side, sides[side])
		}
	}

	builder.WriteString("# HELP constantine_stop_loss_total Total number of stop loss orders placed\n")
	builder.WriteString("# TYPE constantine_stop_loss_total counter\n")
	symbols = symbols[:0]
	for symbol := range stopLossCounts {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		fmt.Fprintf(builder, "constantine_stop_loss_total{symbol=\"%s\"} %d\n", symbol, stopLossCounts[symbol])
	}

	builder.WriteString("# HELP constantine_take_profit_total Total number of take profit orders placed\n")
	builder.WriteString("# TYPE constantine_take_profit_total counter\n")
	symbols = symbols[:0]
	for symbol := range takeProfitCounts {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		fmt.Fprintf(builder, "constantine_take_profit_total{symbol=\"%s\"} %d\n", symbol, takeProfitCounts[symbol])
	}
	metricsMu.RUnlock()

	builder.WriteString("# HELP constantine_callback_panics_total Number of recovered panics from callbacks\n")
	builder.WriteString("# TYPE constantine_callback_panics_total counter\n")
	fmt.Fprintf(builder, "constantine_callback_panics_total %d\n", atomic.LoadUint64(&callbackPanics))

	// Balance metrics
	builder.WriteString("# HELP constantine_balance Current account balance by asset\n")
	builder.WriteString("# TYPE constantine_balance gauge\n")
	assets := make([]string, 0, len(balanceUpdates))
	for asset := range balanceUpdates {
		assets = append(assets, asset)
	}
	sort.Strings(assets)
	for _, asset := range assets {
		fmt.Fprintf(builder, "constantine_balance{asset=\"%s\"} %f\n", asset, balanceUpdates[asset])
	}

	// Position metrics
	builder.WriteString("# HELP constantine_position Current position values by symbol and field\n")
	builder.WriteString("# TYPE constantine_position gauge\n")
	symbols = symbols[:0]
	for symbol := range positionUpdates {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		fields := make([]string, 0, len(positionUpdates[symbol]))
		for field := range positionUpdates[symbol] {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		for _, field := range fields {
			fmt.Fprintf(builder, "constantine_position{symbol=\"%s\",field=\"%s\"} %f\n", symbol, field, positionUpdates[symbol][field])
		}
	}

	// P&L metrics
	builder.WriteString("# HELP constantine_pnl Current P&L by symbol\n")
	builder.WriteString("# TYPE constantine_pnl gauge\n")
	symbols = symbols[:0]
	for symbol := range pnlUpdates {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		fmt.Fprintf(builder, "constantine_pnl{symbol=\"%s\"} %f\n", symbol, pnlUpdates[symbol])
	}

	// Signal metrics
	builder.WriteString("# HELP constantine_signals_total Total trading signals generated by type\n")
	builder.WriteString("# TYPE constantine_signals_total counter\n")
	signalTypes := make([]string, 0, len(signalCounts))
	for signalType := range signalCounts {
		signalTypes = append(signalTypes, signalType)
	}
	sort.Strings(signalTypes)
	for _, signalType := range signalTypes {
		fmt.Fprintf(builder, "constantine_signals_total{type=\"%s\"} %d\n", signalType, signalCounts[signalType])
	}

	// Error metrics
	builder.WriteString("# HELP constantine_errors_total Total errors by type\n")
	builder.WriteString("# TYPE constantine_errors_total counter\n")
	errorTypes := make([]string, 0, len(errorCounts))
	for errorType := range errorCounts {
		errorTypes = append(errorTypes, errorType)
	}
	sort.Strings(errorTypes)
	for _, errorType := range errorTypes {
		fmt.Fprintf(builder, "constantine_errors_total{type=\"%s\"} %d\n", errorType, errorCounts[errorType])
	}

	// WebSocket reconnect metrics
	builder.WriteString("# HELP constantine_websocket_reconnects_total Total WebSocket reconnections by exchange\n")
	builder.WriteString("# TYPE constantine_websocket_reconnects_total counter\n")
	exchanges := make([]string, 0, len(websocketReconnects))
	for exchange := range websocketReconnects {
		exchanges = append(exchanges, exchange)
	}
	sort.Strings(exchanges)
	for _, exchange := range exchanges {
		fmt.Fprintf(builder, "constantine_websocket_reconnects_total{exchange=\"%s\"} %d\n", exchange, websocketReconnects[exchange])
	}

	// API request metrics
	builder.WriteString("# HELP constantine_api_requests_total Total API requests by exchange and endpoint\n")
	builder.WriteString("# TYPE constantine_api_requests_total counter\n")
	exchanges = exchanges[:0]
	for exchange := range apiRequestCounts {
		exchanges = append(exchanges, exchange)
	}
	sort.Strings(exchanges)
	for _, exchange := range exchanges {
		endpoints := make([]string, 0, len(apiRequestCounts[exchange]))
		for endpoint := range apiRequestCounts[exchange] {
			endpoints = append(endpoints, endpoint)
		}
		sort.Strings(endpoints)
		for _, endpoint := range endpoints {
			fmt.Fprintf(builder, "constantine_api_requests_total{exchange=\"%s\",endpoint=\"%s\"} %d\n", exchange, endpoint, apiRequestCounts[exchange][endpoint])
		}
	}

	// API latency metrics (average)
	builder.WriteString("# HELP constantine_api_latency_seconds Average API request latency by exchange and endpoint\n")
	builder.WriteString("# TYPE constantine_api_latency_seconds gauge\n")
	exchanges = exchanges[:0]
	for exchange := range apiRequestLatency {
		exchanges = append(exchanges, exchange)
	}
	sort.Strings(exchanges)
	for _, exchange := range exchanges {
		endpoints := make([]string, 0, len(apiRequestLatency[exchange]))
		for endpoint := range apiRequestLatency[exchange] {
			endpoints = append(endpoints, endpoint)
		}
		sort.Strings(endpoints)
		for _, endpoint := range endpoints {
			latencies := apiRequestLatency[exchange][endpoint]
			if len(latencies) > 0 {
				var sum time.Duration
				for _, lat := range latencies {
					sum += lat
				}
				avg := sum / time.Duration(len(latencies))
				fmt.Fprintf(builder, "constantine_api_latency_seconds{exchange=\"%s\",endpoint=\"%s\"} %f\n", exchange, endpoint, avg.Seconds())
			}
		}
	}

	metricsMu.RUnlock()

	_, _ = w.Write([]byte(builder.String()))
}

// Start begins serving metrics and health endpoints in a separate goroutine.
func (s *Server) Start() error {
	if s == nil || s.srv == nil {
		return nil
	}
	go func() {
		_ = s.srv.ListenAndServe()
	}()
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

// SetReady updates the readiness state exposed on /readyz.
func (s *Server) SetReady(ready bool) {
	if s == nil {
		return
	}
	s.readyState.Store(ready)
}
