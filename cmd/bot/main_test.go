package main

import (
	"context"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/guyghost/constantine/internal/testutils"
)

func TestStartBotComponentsIntegration(t *testing.T) {
	exchange := testutils.NewTestExchange("integration")
	manager := order.NewManager(exchange)
	strat := strategy.NewScalpingStrategy(strategy.DefaultConfig(), exchange)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- startBotComponents(ctx, strat, manager)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("startBotComponents returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("startBotComponents did not exit after cancellation")
	}
}
