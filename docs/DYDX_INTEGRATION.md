# dYdX v4 Integration

Constantine supporte désormais l'exchange décentralisé dYdX v4 pour le trading de perpetuals.

## Vue d'ensemble

dYdX v4 est un exchange décentralisé de contrats perpetuals construit sur la blockchain Cosmos. Constantine s'interface avec l'Indexer API de dYdX pour récupérer les données de marché en temps réel.

## Configuration

### Variables d'environnement

Ajoutez ces variables à votre fichier `.env` :

```bash
# Exchange selection
EXCHANGE=dydx

# Authentication via Mnemonic (IMPORTANT!)
# Pour dYdX, EXCHANGE_API_SECRET doit contenir votre phrase mnémonique (12 ou 24 mots)
EXCHANGE_API_KEY=          # Optionnel pour les endpoints publics
EXCHANGE_API_SECRET="word1 word2 word3 ... word12"  # Votre mnemonic phrase

# Trading configuration
TRADING_SYMBOL=BTC-USD
```

⚠️ **IMPORTANT** :
- Pour dYdX, utilisez `EXCHANGE_API_SECRET` pour stocker votre phrase mnémonique
- La phrase doit être entre guillemets si elle contient des espaces
- Gardez votre mnemonic **ABSOLUMENT SECRET** - ne le partagez jamais
- Utilisez un fichier `.env` qui n'est PAS versionné dans git

### Lancer le bot avec dYdX

```bash
# Mode headless
EXCHANGE=dydx ./bin/constantine --headless

# Mode TUI (dans votre terminal)
EXCHANGE=dydx ./bin/constantine
```

## Fonctionnalités implémentées

### ✅ Données de marché (lecture seule)

- **Ticker**: Prix bid/ask, dernier prix, volume 24h
- **Order Book**: Profondeur du carnet d'ordres
- **Candles historiques**: Données OHLCV pour le backtesting
- **WebSocket**: Flux en temps réel pour ticker, orderbook et trades

### ⚠️ Trading (implémentation partielle)

Les méthodes de trading sont actuellement des stubs qui nécessitent :
- Adresse de subaccount dYdX
- Signature des transactions avec clé privée
- Configuration du validator node

## Architecture

```
internal/exchanges/dydx/
├── client.go          # Client principal implémentant l'interface Exchange
├── http.go            # Client HTTP pour les requêtes REST
├── websocket.go       # Client WebSocket pour les données en temps réel
└── types.go           # Structures de données dYdX v4
```

## Utilisation avec le Backtesting

Le framework de backtesting supporte dYdX pour tester vos stratégies :

```bash
# Télécharger des données historiques depuis dYdX
# (utiliser leur API ou des sources tierces)

# Exécuter un backtest
./bin/backtest --data=dydx_btc_data.csv --symbol=BTC-USD
```

## Marchés supportés

dYdX v4 supporte de nombreux marchés perpetuals :

- **Crypto majeurs**: BTC-USD, ETH-USD, SOL-USD, AVAX-USD
- **Altcoins**: Plus de 50 paires disponibles
- **Leverage**: Jusqu'à 20x selon le marché

### Vérifier les marchés disponibles

```go
client := dydx.NewClient("", "")
symbols := client.SupportedSymbols()
```

## Endpoints API

Le client utilise les endpoints publics de l'Indexer dYdX v4 :

- **Base URL**: `https://indexer.dydx.trade`
- **WebSocket**: `wss://indexer.dydx.trade/v4/ws`

### Exemples d'endpoints

```
GET /v4/perpetualMarkets              # Liste des marchés et tickers
GET /v4/orderbooks/perpetualMarket/:market  # Order book
GET /v4/candles/perpetualMarkets/:market    # Candles historiques
```

## Limitations actuelles

### Trading

❌ **Non implémenté** :
- Placement d'ordres (nécessite signature blockchain)
- Annulation d'ordres
- Gestion des positions en temps réel
- Requêtes d'historique de compte

### Données de compte

❌ **Nécessite configuration additionnelle** :
- Balance du compte
- Positions ouvertes
- Historique des trades

Ces fonctionnalités nécessitent :
1. Un wallet Cosmos/dYdX configuré
2. Une adresse de subaccount
3. La capacité de signer des transactions

## Prochaines étapes

Pour activer le trading complet sur dYdX :

### 1. Configuration du wallet

```bash
# Installer dYdX CLI
npm install -g @dydxprotocol/v4-client

# Créer un wallet
dydx-cli wallet create
```

### 2. Intégration de la signature

Ajouter la bibliothèque de signature dYdX :

```bash
go get github.com/dydxprotocol/v4-chain/...
```

### 3. Implémentation des ordres

Mettre à jour `PlaceOrder()` dans `client.go` pour :
- Créer le message d'ordre
- Signer avec la clé privée
- Broadcaster la transaction

## Ressources

- [Documentation dYdX v4](https://docs.dydx.exchange/)
- [API Reference](https://docs.dydx.exchange/api_integration-indexer/indexer_api)
- [Client TypeScript officiel](https://github.com/dydxprotocol/v4-clients)
- [Discord dYdX](https://discord.gg/dydx)

## Exemples de code

### Récupérer le ticker

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/guyghost/constantine/internal/exchanges/dydx"
)

func main() {
    client := dydx.NewClient("", "")
    ctx := context.Background()

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()

    ticker, err := client.GetTicker(ctx, "BTC-USD")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("BTC-USD: Bid=%s Ask=%s Last=%s\n",
        ticker.Bid.String(),
        ticker.Ask.String(),
        ticker.Last.String())
}
```

### Streamer les mises à jour du ticker

```go
client := dydx.NewClient("", "")
ctx := context.Background()

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}

client.SubscribeTicker(ctx, "BTC-USD", func(ticker *exchanges.Ticker) {
    fmt.Printf("BTC-USD Update: $%s\n", ticker.Last.String())
})

// Keep running
select {}
```

## Support

Pour toute question ou problème avec l'intégration dYdX :
1. Vérifiez la [documentation officielle](https://docs.dydx.exchange/)
2. Consultez les logs du bot
3. Ouvrez une issue sur GitHub

## Sécurité

⚠️ **Important** :
- Ne commitez JAMAIS vos clés API dans le code
- Utilisez des variables d'environnement ou un gestionnaire de secrets
- Pour le trading réel, utilisez un subaccount dédié avec capital limité
- Testez toujours en mode paper trading d'abord
