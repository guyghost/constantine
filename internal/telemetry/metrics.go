package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	metricsMu        sync.RWMutex
	orderCounts      = make(map[string]map[string]uint64)
	stopLossCounts   = make(map[string]uint64)
	takeProfitCounts = make(map[string]uint64)
	callbackPanics   uint64
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
