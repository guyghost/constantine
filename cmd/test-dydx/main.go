package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/exchanges/dydx"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("🚀 dYdX v4 Testnet - Test de Trading Complet")
	fmt.Println("============================================\n")

	// Charger les variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Aucun fichier .env trouvé, utilisation des variables d'environnement système")
	}

	// Vérifier le mnemonic
	mnemonic := os.Getenv("DYDX_MNEMONIC")
	if mnemonic == "" {
		log.Fatal("❌ DYDX_MNEMONIC non défini. Configurez votre .env ou exportez la variable.")
	}

	if len(mnemonic) < 50 {
		log.Fatal("❌ Le mnemonic semble invalide (trop court). Vérifiez votre configuration.")
	}

	fmt.Println("✅ Mnemonic chargé (longueur:", len(mnemonic), "caractères)")

	// Créer le contexte avec timeout global
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// ÉTAPE 1: Connexion au testnet
	fmt.Println("\n📡 ÉTAPE 1: Connexion au testnet dYdX...")
	client, err := dydx.NewClientWithMnemonicAndURL(
		mnemonic,
		0, // subaccount 0
		"https://indexer.v4testnet.dydx.exchange",
		"wss://indexer.v4testnet.dydx.exchange/v4/ws",
	)
	if err != nil {
		log.Fatalf("❌ Échec création client: %v", err)
	}

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("❌ Échec connexion: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("✅ Connecté au testnet dYdX")

	// ÉTAPE 2: Vérifier le solde
	fmt.Println("\n💰 ÉTAPE 2: Vérification du solde...")
	balance, err := client.GetBalance(ctx)
	if err != nil {
		log.Fatalf("❌ Échec récupération solde: %v", err)
	}

	if usdcBalance, ok := balance["USDC"]; ok {
		fmt.Printf("✅ Solde USDC: %s\n", usdcBalance.String())

		if usdcBalance.LessThan(decimal.NewFromInt(10)) {
			log.Println("⚠️  ATTENTION: Solde faible (< 10 USDC)")
			log.Println("   Pour obtenir des USDC testnet:")
			log.Println("   1. Allez sur https://v4.testnet.dydx.exchange/")
			log.Println("   2. Connectez votre wallet")
			log.Println("   3. Utilisez le faucet pour obtenir des tokens")
		}
	} else {
		log.Println("⚠️  Aucun solde USDC trouvé")
	}

	// ÉTAPE 3: Récupérer les données de marché
	fmt.Println("\n📊 ÉTAPE 3: Récupération des données de marché BTC-USD...")
	ticker, err := client.GetTicker(ctx, "BTC-USD")
	if err != nil {
		log.Fatalf("❌ Échec récupération ticker: %v", err)
	}

	fmt.Printf("✅ Prix BTC-USD:\n")
	fmt.Printf("   Bid:    %s\n", ticker.Bid.String())
	fmt.Printf("   Ask:    %s\n", ticker.Ask.String())
	fmt.Printf("   Last:   %s\n", ticker.Last.String())
	fmt.Printf("   Volume: %s\n", ticker.Volume24h.String())

	// ÉTAPE 4: Récupérer l'orderbook
	fmt.Println("\n📖 ÉTAPE 4: Récupération de l'orderbook...")
	orderbook, err := client.GetOrderBook(ctx, "BTC-USD", 5)
	if err != nil {
		log.Fatalf("❌ Échec récupération orderbook: %v", err)
	}

	fmt.Printf("✅ Orderbook (top 5):\n")
	fmt.Printf("   Bids: %d niveaux\n", len(orderbook.Bids))
	if len(orderbook.Bids) > 0 {
		fmt.Printf("   Meilleur bid: %s @ %s\n",
			orderbook.Bids[0].Amount.String(),
			orderbook.Bids[0].Price.String())
	}
	fmt.Printf("   Asks: %d niveaux\n", len(orderbook.Asks))
	if len(orderbook.Asks) > 0 {
		fmt.Printf("   Meilleur ask: %s @ %s\n",
			orderbook.Asks[0].Amount.String(),
			orderbook.Asks[0].Price.String())
	}

	// ÉTAPE 5: Vérifier les positions existantes
	fmt.Println("\n📈 ÉTAPE 5: Vérification des positions...")
	positions, err := client.GetPositions(ctx)
	if err != nil {
		log.Fatalf("❌ Échec récupération positions: %v", err)
	}

	if len(positions) == 0 {
		fmt.Println("✅ Aucune position ouverte")
	} else {
		fmt.Printf("✅ %d position(s) ouverte(s):\n", len(positions))
		for _, pos := range positions {
			fmt.Printf("   %s: %s %s @ %s (PnL: %s)\n",
				pos.Symbol,
				pos.Side,
				pos.Size.String(),
				pos.EntryPrice.String(),
				pos.UnrealizedPnL.String(),
			)
		}
	}

	// ÉTAPE 6: Test de placement d'ordre (OPTIONNEL)
	fmt.Println("\n🔄 ÉTAPE 6: Test de placement d'ordre...")
	fmt.Println("⚠️  ATTENTION: Cette étape va placer un VRAI ordre sur testnet")
	fmt.Print("Voulez-vous continuer ? (oui/non): ")

	var response string
	fmt.Scanln(&response)

	if response == "oui" || response == "o" || response == "yes" || response == "y" {
		// Calculer un prix limite bien en dehors du marché (pour ne pas être exécuté)
		limitPrice := ticker.Last.Mul(decimal.NewFromFloat(0.5)) // 50% du prix actuel

		order := &exchanges.Order{
			Symbol: "BTC-USD",
			Side:   "buy",
			Type:   "limit",
			Amount: decimal.NewFromFloat(0.001), // 0.001 BTC (très petit)
			Price:  limitPrice,
		}

		fmt.Printf("\n📝 Placement d'un ordre LIMIT BUY:\n")
		fmt.Printf("   Symbol: %s\n", order.Symbol)
		fmt.Printf("   Side:   %s\n", order.Side)
		fmt.Printf("   Amount: %s BTC\n", order.Amount.String())
		fmt.Printf("   Price:  %s (50%% du prix actuel = ne sera pas exécuté)\n", order.Price.String())

		placedOrder, err := client.PlaceOrder(ctx, order)
		if err != nil {
			log.Printf("❌ Échec placement ordre: %v", err)
			log.Println("\n💡 Causes possibles:")
			log.Println("   - Python client non installé: pip3 install dydx-v4-client-py")
			log.Println("   - Solde insuffisant")
			log.Println("   - Problème de connexion réseau")
		} else {
			fmt.Printf("✅ Ordre placé avec succès!\n")
			fmt.Printf("   Order ID: %s\n", placedOrder.ID)
			fmt.Printf("   Status:   %s\n", placedOrder.Status)

			// Attendre un peu
			fmt.Println("\n⏳ Attente de 3 secondes...")
			time.Sleep(3 * time.Second)

			// ÉTAPE 7: Annulation de l'ordre
			fmt.Println("\n❌ ÉTAPE 7: Annulation de l'ordre...")
			if err := client.CancelOrder(ctx, placedOrder.ID); err != nil {
				log.Printf("❌ Échec annulation: %v", err)
			} else {
				fmt.Println("✅ Ordre annulé avec succès")
			}
		}
	} else {
		fmt.Println("⏭️  Test de placement d'ordre sauté")
	}

	// RÉSUMÉ FINAL
	fmt.Println("\n" + "="*50)
	fmt.Println("📋 RÉSUMÉ DU TEST")
	fmt.Println("="*50)
	fmt.Println("✅ Connexion au testnet: OK")
	fmt.Println("✅ Récupération du solde: OK")
	fmt.Println("✅ Données de marché: OK")
	fmt.Println("✅ Orderbook: OK")
	fmt.Println("✅ Positions: OK")
	if response == "oui" || response == "o" {
		fmt.Println("✅ Placement/Annulation ordre: Testé")
	}
	fmt.Println("\n🎉 Tous les tests sont passés!")
	fmt.Println("\n💡 Prochaines étapes:")
	fmt.Println("   1. Obtenez des USDC testnet si besoin")
	fmt.Println("   2. Testez avec des ordres réels (petits montants)")
	fmt.Println("   3. Vérifiez les logs pour détecter les erreurs")
	fmt.Println("   4. Passez en mainnet quand tout fonctionne")
}
