# Guide de Démarrage Rapide - dYdX v4

## 🎉 Cycle Complet de Trading Implémenté !

Le bot Constantine supporte maintenant un **cycle complet de trading** avec dYdX v4 :

```
✅ Connexion → ✅ Données de marché → ✅ Analyse → ✅ Passage d'ordres → ✅ Gestion de positions → ✅ Fermeture de positions
```

## Installation

### 1. Client Python dYdX (Requis pour le trading)

```bash
# Option 1 : Installation globale
pip3 install dydx-v4-client-py v4-proto

# Option 2 : Environnement virtuel (recommandé)
python3 -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install dydx-v4-client-py v4-proto
```

### 2. Vérification

```bash
python3 -c "from v4_client_py import NodeClient; print('✅ Client dYdX installé !')"
```

## Configuration

### 1. Créer/Modifier votre fichier `.env`

```bash
# Copiez le template
cp .env.op.template .env

# Ajoutez vos credentials dYdX
```

### 2. Variables d'environnement dYdX

```env
# dYdX Configuration
EXCHANGE=dydx
DYDX_MNEMONIC="your twelve or twenty-four word mnemonic phrase here"

# Network (testnet pour débuter)
DYDX_NETWORK=testnet

# OU pour production
# DYDX_NETWORK=mainnet
```

### 3. Obtenir une Mnemonic

Si vous n'avez pas encore de wallet dYdX :

1. Allez sur https://trade.dydx.exchange/ (mainnet) ou https://v4.testnet.dydx.exchange/ (testnet)
2. Connectez-vous avec MetaMask ou créez un nouveau wallet
3. Sauvegardez votre phrase de récupération (mnemonic) de 12 ou 24 mots
4. **⚠️ IMPORTANT** : Ne partagez JAMAIS votre mnemonic !

## Utilisation

### Exemple de Trading Complet

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/guyghost/constantine/internal/exchanges/dydx"
    "github.com/shopspring/decimal"
)

func main() {
    ctx := context.Background()

    // 1. Connexion
    mnemonic := "your mnemonic here"
    client, err := dydx.NewClientWithMnemonic(mnemonic, 0)
    if err != nil {
        log.Fatal(err)
    }

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()

    fmt.Println("✅ Connecté à dYdX")

    // 2. Obtenir des données de marché
    ticker, err := client.GetTicker(ctx, "BTC-USD")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("📈 BTC-USD Prix: %s\n", ticker.Last)

    // 3. Vérifier le solde
    balance, err := client.GetBalance(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("💰 Solde: %s USDC\n", balance["USDC"])

    // 4. Passer un ordre LIMIT
    order := &exchanges.Order{
        Symbol: "BTC-USD",
        Side:   "buy",
        Type:   "limit",
        Amount: decimal.NewFromFloat(0.01),  // 0.01 BTC
        Price:  decimal.NewFromFloat(50000), // Prix limite
    }

    placedOrder, err := client.PlaceOrder(ctx, order)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ Ordre placé: %s\n", placedOrder.ID)

    // 5. Attendre un peu
    time.Sleep(5 * time.Second)

    // 6. Annuler l'ordre
    if err := client.CancelOrder(ctx, placedOrder.ID); err != nil {
        log.Fatal(err)
    }
    fmt.Println("✅ Ordre annulé")

    // 7. Vérifier les positions
    positions, err := client.GetPositions(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("📊 Positions ouvertes: %d\n", len(positions))
    for _, pos := range positions {
        fmt.Printf("  - %s: %s %s @ %s\n",
            pos.Symbol,
            pos.Side,
            pos.Size,
            pos.EntryPrice,
        )
    }
}
```

### Ordre Market (Exécution Immédiate)

```go
marketOrder := &exchanges.Order{
    Symbol: "ETH-USD",
    Side:   "buy",
    Type:   "market",
    Amount: decimal.NewFromFloat(0.1),  // 0.1 ETH
}

placedOrder, err := client.PlaceOrder(ctx, marketOrder)
```

### WebSocket en Temps Réel

```go
// S'abonner aux mises à jour de prix
callback := func(ticker *exchanges.Ticker) {
    fmt.Printf("📊 %s: Bid=%s Ask=%s Last=%s\n",
        ticker.Symbol,
        ticker.Bid,
        ticker.Ask,
        ticker.Last,
    )
}

client.SubscribeTicker(ctx, "BTC-USD", callback)

// S'abonner à l'orderbook
obCallback := func(ob *exchanges.OrderBook) {
    fmt.Printf("📖 Orderbook: %d bids, %d asks\n",
        len(ob.Bids),
        len(ob.Asks),
    )
}

client.SubscribeOrderBook(ctx, "BTC-USD", obCallback)
```

## Marchés Disponibles

Marchés populaires sur dYdX v4 :
- `BTC-USD` - Bitcoin
- `ETH-USD` - Ethereum
- `SOL-USD` - Solana
- `AVAX-USD` - Avalanche
- Et 50+ autres paires de trading

## Types d'Ordres

### Ordres Court-Terme (20 blocs ~2-3 secondes)
- **Market** : Exécution immédiate au meilleur prix
- **IOC (Immediate-or-Cancel)** : Exécuté immédiatement ou annulé

### Ordres Long-Terme (Stateful)
- **Limit** : Prix limite avec expiration
- **Post-Only** : Ajouté au carnet sans exécution immédiate

### Ordres Conditionnels
- **Stop Loss** : Déclenché quand le prix baisse
- **Take Profit** : Déclenché quand le prix monte

## Testnet vs Mainnet

### Testnet (Recommandé pour débuter)
```env
DYDX_NETWORK=testnet
```
- **Indexer**: https://indexer.v4testnet.dydx.exchange
- **WebSocket**: wss://indexer.v4testnet.dydx.exchange/v4/ws
- **Chain ID**: dydx-testnet-4
- Utilisez des fonds de test

### Mainnet (Production)
```env
DYDX_NETWORK=mainnet
```
- **Indexer**: https://indexer.dydx.trade
- **WebSocket**: wss://indexer.dydx.trade/v4/ws
- **Chain ID**: dydx-mainnet-1
- ⚠️ **FONDS RÉELS** - Testez d'abord sur testnet !

## Résolution de Problèmes

### Erreur : "Python client not initialized"

**Solution** : Utilisez `NewClientWithMnemonic` au lieu de `NewClient`

```go
// ❌ Ne fonctionnera pas pour le trading
client, _ := dydx.NewClient("", "")

// ✅ Correct
client, _ := dydx.NewClientWithMnemonic(mnemonic, 0)
```

### Erreur : "dydx-v4-client-py not installed"

**Solution** :
```bash
pip3 install dydx-v4-client-py v4-proto
```

### Erreur : "invalid mnemonic"

**Solution** : Vérifiez que votre mnemonic est une phrase BIP39 valide (12 ou 24 mots)

### Erreur : "Python script error: no such file or directory"

**Solution** : Le script doit être à `internal/exchanges/dydx/scripts/dydx_client.py`

Vérifiez que vous exécutez le bot depuis la racine du projet :
```bash
cd /path/to/constantine
./cmd/bot/bot
```

## Architecture Technique

L'implémentation utilise une approche hybride :

1. **Données de marché** : Implémentation native Go (HTTP + WebSocket)
2. **Trading** : Wrapper Python du client officiel dYdX v4
3. **Authentification** : Wallet BIP39 avec dérivation Cosmos

```
┌─────────────┐
│   Bot Go    │
└──────┬──────┘
       │
       ├─── HTTP/WebSocket ──→ dYdX Indexer (données)
       │
       └─── Python Wrapper ──→ dYdX Chain (ordres)
```

## Fonctionnalités

### ✅ Implémenté
- Connexion et authentification
- Données de marché en temps réel (ticker, orderbook, trades)
- Historique OHLCV
- Passage d'ordres (market, limit, stop, take profit)
- Annulation d'ordres
- Consultation des positions
- Consultation du solde

### ⚠️ Limitations
- Les queries d'ordres sont limitées par l'API indexer
- Pas de modification d'ordre (non supporté par dYdX v4)

### 🚀 Roadmap Future
- Implémentation native Go (sans Python)
- Gestion du levier
- Monitoring de liquidation
- Averaging de positions

## Ressources

- [Documentation dYdX v4](https://docs.dydx.xyz/)
- [Client Python dYdX](https://github.com/dydxprotocol/v4-clients/tree/main/v4-client-py-v2)
- [Documentation Détaillée](./internal/exchanges/dydx/README.md)

## ⚠️ Avertissement

**Ceci est un logiciel alpha. Testez TOUJOURS sur testnet avant d'utiliser des fonds réels.**

- Commencez avec de petites positions
- Utilisez des stop loss
- Ne tradez que ce que vous pouvez vous permettre de perdre
- La crypto est volatile - tradez de manière responsable

## Support

Pour plus d'informations :
1. Consultez `internal/exchanges/dydx/README.md`
2. Lisez la documentation dYdX officielle
3. Testez sur testnet en premier

Bon trading ! 🚀
