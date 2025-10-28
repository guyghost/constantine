package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/exchanges/dydx"
	"github.com/guyghost/constantine/internal/execution"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/risk"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

// MockOrderManager implements the OrderManager interface for testing
type MockOrderManager struct {
	positions []*order.ManagedPosition
	trades    []*exchanges.Order
}

func NewMockOrderManager() *MockOrderManager {
	return &MockOrderManager{
		positions: []*order.ManagedPosition{},
		trades:    []*exchanges.Order{},
	}
}

func (m *MockOrderManager) GetPositions() []*order.ManagedPosition {
	return m.positions
}

func (m *MockOrderManager) PlaceOrder(ctx context.Context, req *order.OrderRequest) (*exchanges.Order, error) {
	ord := &exchanges.Order{
		Symbol: req.Symbol,
		Side:   req.Side,
		Type:   req.Type,
		Amount: req.Amount,
		Price:  req.Price,
		ID:     fmt.Sprintf("MOCK-%d", time.Now().UnixNano()),
		Status: "pending",
	}
	m.trades = append(m.trades, ord)
	fmt.Printf("✅ Ordre placé en mock:\n")
	fmt.Printf("   ID:     %s\n", ord.ID)
	fmt.Printf("   Symbol: %s\n", ord.Symbol)
	fmt.Printf("   Side:   %s\n", ord.Side)
	fmt.Printf("   Amount: %s\n", ord.Amount.String())
	fmt.Printf("   Price:  %s\n", ord.Price.String())
	fmt.Printf("   StopLoss:   %s\n", req.StopLoss.String())
	fmt.Printf("   TakeProfit: %s\n", req.TakeProfit.String())
	return ord, nil
}

func (m *MockOrderManager) ClosePosition(ctx context.Context, symbol string) error {
	fmt.Printf("❌ Position fermée pour %s\n", symbol)
	return nil
}

func main() {
	fmt.Println("🚀 Test Signal d'Achat Artificiel - dYdX BTC-USD")
	fmt.Println("================================================")

	// Charger les variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Aucun fichier .env trouvé")
	}

	// Créer le contexte avec timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ÉTAPE 1: Connexion au dYdX (essayer en lecture seule d'abord)
	fmt.Println("\n📡 ÉTAPE 1: Connexion à dYdX...")
	mnemonic := os.Getenv("DYDX_MNEMONIC")
	if mnemonic == "" {
		log.Fatal("❌ DYDX_MNEMONIC non défini")
	}

	// Créer le client avec accès en lecture seule
	apiKey := os.Getenv("DYDX_API_KEY")
	if apiKey == "" {
		apiKey = "" // Peut être vide pour lecture seule
	}

	client, err := dydx.NewClient(apiKey, mnemonic)
	if err != nil {
		log.Fatalf("❌ Échec création client: %v", err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("❌ Échec connexion: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("✅ Connecté à dYdX")

	// ÉTAPE 2: Récupérer les prix actuels
	fmt.Println("\n💰 ÉTAPE 2: Récupération des prix BTC-USD...")
	ticker, err := client.GetTicker(ctx, "BTC-USD")
	if err != nil {
		log.Printf("⚠️  Impossible de récupérer le ticker réel: %v. Utilisation de données mockées.\n", err)
		// Utiliser des données mockées
		ticker = &exchanges.Ticker{
			Symbol:    "BTC-USD",
			Bid:       decimal.NewFromFloat(114400.0),
			Ask:       decimal.NewFromFloat(114450.0),
			Last:      decimal.NewFromFloat(114450.0),
			Volume24h: decimal.NewFromFloat(1000000.0),
		}
	}

	currentPrice := ticker.Last
	fmt.Printf("✅ Prix BTC-USD: %s\n", currentPrice.String())
	fmt.Printf("   Bid: %s, Ask: %s\n", ticker.Bid.String(), ticker.Ask.String())

	// ÉTAPE 3: Récupérer le solde (ou utiliser des données mockées)
	fmt.Println("\n💵 ÉTAPE 3: Configuration du solde...")
	accountBalance := decimal.NewFromFloat(5000.0) // 5000 USDC par défaut pour le test
	usedRealBalance := false

	// Essayer de récupérer le solde réel
	balances, err := client.GetBalance(ctx)
	if err == nil && len(balances) > 0 {
		for _, bal := range balances {
			if bal.Asset == "USDC" {
				// Utiliser le solde réel seulement s'il est suffisant (> 100 USDC)
				if bal.Free.GreaterThan(decimal.NewFromFloat(100.0)) {
					accountBalance = bal.Free
					usedRealBalance = true
					fmt.Printf("✅ Solde USDC réel: %s\n", accountBalance.String())
				} else {
					fmt.Printf("⚠️  Solde réel insuffisant (%s USDC), utilisation de 5000 USDC pour le test\n", bal.Free.String())
				}
				break
			}
		}
	} else {
		fmt.Printf("⚠️  Impossible de récupérer le solde réel, utilisation de 5000 USDC pour le test\n")
	}

	if !usedRealBalance {
		fmt.Printf("   Solde utilisé: %s USDC (mocké)\n", accountBalance.String())
	}

	// ÉTAPE 4: Créer un signal d'achat artificiel
	fmt.Println("\n📊 ÉTAPE 4: Création du signal d'achat artificiel...")
	artificialSignal := &strategy.Signal{
		Type:      strategy.SignalTypeEntry,
		Side:      exchanges.OrderSideBuy,
		Symbol:    "BTC-USD",
		Price:     currentPrice,
		Strength:  0.75, // 75% de confiance
		Reason:    "Signal artificiel de test",
		Timestamp: time.Now().Unix(),
	}

	fmt.Printf("✅ Signal créé:\n")
	fmt.Printf("   Type:      %s\n", artificialSignal.Type)
	fmt.Printf("   Side:      %s\n", artificialSignal.Side)
	fmt.Printf("   Symbol:    %s\n", artificialSignal.Symbol)
	fmt.Printf("   Price:     %s\n", artificialSignal.Price.String())
	fmt.Printf("   Strength:  %.2f\n", artificialSignal.Strength)
	fmt.Printf("   Reason:    %s\n", artificialSignal.Reason)

	// ÉTAPE 5: Créer les managers pour l'exécution
	fmt.Println("\n⚙️  ÉTAPE 5: Initialisation des managers...")

	// Créer le mock order manager
	mockOrderManager := NewMockOrderManager()

	// Créer le risk manager
	riskConfig := &risk.Config{
		MaxPositionSize:        decimal.NewFromFloat(10100), // Max 10100 USD per position (leave headroom)
		MaxPositions:           3,                           // Max 3 positions
		MaxLeverage:            decimal.NewFromInt(1),
		MaxDailyLoss:           decimal.NewFromFloat(200), // Max 200 USD loss per day
		MaxDrawdown:            decimal.NewFromFloat(20),  // 20% max drawdown
		RiskPerTrade:           decimal.NewFromFloat(1.0), // 1% risk per trade
		MinAccountBalance:      decimal.NewFromFloat(10),  // Min 10 USD
		DailyTradingLimit:      50,
		CooldownPeriod:         15 * time.Minute,
		ConsecutiveLossLimit:   3,
		MaxExposurePerSymbol:   decimal.NewFromFloat(100), // 100% max exposure per symbol (permettre au moins le test)
		MaxSameSymbolPositions: 2,                         // Allow 2 positions per symbol
	}

	riskManager := risk.NewManager(riskConfig, accountBalance)

	fmt.Println("✅ Managers initialisés")
	fmt.Printf("   Risk Config: MaxPositionSize=%s, MaxDrawdown=%.1f%%\n",
		riskConfig.MaxPositionSize.String(),
		riskConfig.MaxDrawdown.InexactFloat64())

	// ÉTAPE 6: Créer et exécuter le signal via ExecutionAgent
	fmt.Println("\n🎯 ÉTAPE 6: Exécution du signal via ExecutionAgent...")

	executionConfig := execution.DefaultConfig()
	executionConfig.AutoExecute = true
	executionConfig.MinSignalStrength = 0.3                        // Accepter les signaux > 30%
	executionConfig.StopLossPercent = decimal.NewFromFloat(0.01)   // 1% stop loss
	executionConfig.TakeProfitPercent = decimal.NewFromFloat(0.02) // 2% take profit

	executionAgent := execution.NewExecutionAgent(
		mockOrderManager,
		riskManager,
		executionConfig,
	)

	err = executionAgent.HandleSignal(ctx, artificialSignal)
	if err != nil {
		fmt.Printf("❌ Erreur durant l'exécution: %v\n", err)
	} else {
		fmt.Println("✅ Signal exécuté avec succès!")
	}

	// ÉTAPE 7: Résumé
	fmt.Println("\n" + repeat("=", 60))
	fmt.Println("📋 RÉSUMÉ DU TEST")
	fmt.Println(repeat("=", 60))
	fmt.Printf("Ticker BTC-USD:        %s (Bid: %s, Ask: %s)\n",
		currentPrice.String(),
		ticker.Bid.String(),
		ticker.Ask.String())
	fmt.Printf("Solde USDC:            %s\n", accountBalance.String())
	fmt.Printf("Signal généré:         BUY avec confiance %.0f%%\n", artificialSignal.Strength*100)
	fmt.Printf("Ordres placés:         %d\n", len(mockOrderManager.trades))

	if len(mockOrderManager.trades) > 0 {
		fmt.Println("\n✅ Détails des ordres générés:")
		for i, ord := range mockOrderManager.trades {
			positionValue := ord.Amount.Mul(ord.Price)
			fmt.Printf("\n   Ordre #%d:\n", i+1)
			fmt.Printf("      ID:     %s\n", ord.ID)
			fmt.Printf("      Side:   %s\n", ord.Side)
			fmt.Printf("      Amount: %s\n", ord.Amount.String())
			fmt.Printf("      Price:  %s\n", ord.Price.String())
			fmt.Printf("      Value:  %s USD\n", positionValue.String())
		}
	} else {
		fmt.Println("\n⚠️  Aucun ordre n'a été généré (peut être dû aux vérifications de risque)")
	}

	fmt.Println("\n" + repeat("=", 60))
	fmt.Println("🎉 Test signal artificiel complété!")
	fmt.Println(repeat("=", 60))
}

// Helper function for string repeat
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
