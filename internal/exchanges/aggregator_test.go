package exchanges

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewMultiExchangeAggregator(t *testing.T) {
	exchanges := map[string]Exchange{
		"exchange1": NewMockExchange("exchange1"),
		"exchange2": NewMockExchange("exchange2"),
	}

	aggregator := NewMultiExchangeAggregator(exchanges)

	if aggregator == nil {
		t.Fatal("aggregator should not be nil")
	}
	if aggregator.exchanges == nil {
		t.Fatal("exchanges map should not be nil")
	}
	if aggregator.data == nil {
		t.Fatal("data should not be nil")
	}
	if len(aggregator.exchanges) != 2 {
		t.Errorf("expected 2 exchanges, got %d", len(aggregator.exchanges))
	}
}

func TestMultiExchangeAggregator_ConnectAll(t *testing.T) {
	exchanges := map[string]Exchange{
		"exchange1": NewMockExchange("exchange1"),
		"exchange2": NewMockExchange("exchange2"),
	}

	aggregator := NewMultiExchangeAggregator(exchanges)

	ctx := context.Background()
	err := aggregator.ConnectAll(ctx)

	if err != nil {
		t.Fatalf("ConnectAll failed: %v", err)
	}

	data := aggregator.GetAggregatedData()
	if len(data.Exchanges) != 2 {
		t.Errorf("expected 2 exchanges in data, got %d", len(data.Exchanges))
	}

	for name, exchangeData := range data.Exchanges {
		if exchangeData.Name != name {
			t.Errorf("expected exchange name %s, got %s", name, exchangeData.Name)
		}
		if !exchangeData.Connected {
			t.Errorf("exchange %s should be connected", name)
		}
		if exchangeData.Error != nil {
			t.Errorf("exchange %s should not have error: %v", name, exchangeData.Error)
		}
	}
}

func TestMultiExchangeAggregator_RefreshData(t *testing.T) {
	exchanges := map[string]Exchange{
		"exchange1": NewMockExchange("exchange1"),
		"exchange2": NewMockExchange("exchange2"),
	}

	aggregator := NewMultiExchangeAggregator(exchanges)

	ctx := context.Background()
	err := aggregator.ConnectAll(ctx)
	if err != nil {
		t.Fatalf("ConnectAll failed: %v", err)
	}

	err = aggregator.RefreshData(ctx)
	if err != nil {
		t.Fatalf("RefreshData failed: %v", err)
	}

	data := aggregator.GetAggregatedData()

	// Check total balance (2 exchanges * 1100 USD each = 2200)
	expectedBalance := decimal.NewFromFloat(2200)
	if !data.TotalBalance.Equal(expectedBalance) {
		t.Errorf("expected total balance %s, got %s", expectedBalance, data.TotalBalance)
	}

	// Check total PnL (2 exchanges * 100 PnL each = 200)
	expectedPnL := decimal.NewFromFloat(200)
	if !data.TotalPnL.Equal(expectedPnL) {
		t.Errorf("expected total PnL %s, got %s", expectedPnL, data.TotalPnL)
	}

	// Check that LastUpdate is set
	if data.LastUpdate <= 0 {
		t.Error("LastUpdate should be greater than 0")
	}

	// Check individual exchange data
	if len(data.Exchanges) != 2 {
		t.Errorf("expected 2 exchanges, got %d", len(data.Exchanges))
	}
	for _, exchangeData := range data.Exchanges {
		if !exchangeData.Connected {
			t.Errorf("exchange %s should be connected", exchangeData.Name)
		}
		if exchangeData.Error != nil {
			t.Errorf("exchange %s should not have error: %v", exchangeData.Name, exchangeData.Error)
		}
		if len(exchangeData.Balances) != 1 {
			t.Errorf("exchange %s should have 1 balance, got %d", exchangeData.Name, len(exchangeData.Balances))
		}
		if len(exchangeData.Positions) != 1 {
			t.Errorf("exchange %s should have 1 position, got %d", exchangeData.Name, len(exchangeData.Positions))
		}
	}
}

func TestMultiExchangeAggregator_GetExchange(t *testing.T) {
	exchanges := map[string]Exchange{
		"exchange1": NewMockExchange("exchange1"),
		"exchange2": NewMockExchange("exchange2"),
	}

	aggregator := NewMultiExchangeAggregator(exchanges)

	// Test existing exchange
	exchange, exists := aggregator.GetExchange("exchange1")
	if !exists {
		t.Error("exchange1 should exist")
	}
	if exchange == nil {
		t.Error("exchange1 should not be nil")
	}
	if exchange.Name() != "exchange1" {
		t.Errorf("expected exchange name 'exchange1', got '%s'", exchange.Name())
	}

	// Test non-existing exchange
	exchange, exists = aggregator.GetExchange("nonexistent")
	if exists {
		t.Error("nonexistent exchange should not exist")
	}
	if exchange != nil {
		t.Error("nonexistent exchange should be nil")
	}
}

func TestMultiExchangeAggregator_PlaceOrder(t *testing.T) {
	exchanges := map[string]Exchange{
		"exchange1": NewMockExchange("exchange1"),
	}

	aggregator := NewMultiExchangeAggregator(exchanges)

	order := &Order{
		Symbol: "BTC-USD",
		Side:   OrderSideBuy,
		Type:   OrderTypeLimit,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.1),
	}

	ctx := context.Background()
	placedOrder, err := aggregator.PlaceOrder(ctx, "exchange1", order)

	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}
	if placedOrder == nil {
		t.Fatal("placedOrder should not be nil")
	}
	if placedOrder.ID != "new_order_exchange1" {
		t.Errorf("expected order ID 'new_order_exchange1', got '%s'", placedOrder.ID)
	}
	if placedOrder.Status != OrderStatusOpen {
		t.Errorf("expected order status 'open', got '%s'", placedOrder.Status)
	}
}
