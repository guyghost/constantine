# État des Exchanges

Ce document détaille l'état d'implémentation de chaque exchange supporté par Constantine.

## Vue d'ensemble

| Exchange | Données marché | WebSocket | Trading | Authentification | Statut |
|----------|----------------|-----------|---------|------------------|---------|
| dYdX v4 | ✅ Complet | ✅ Complet | ❌ **NON IMPLÉMENTÉ** | ✅ Mnemonic | ⚠️ **LECTURE SEULE** |
| Hyperliquid | 🔧 Mock | ✅ Partiel | 🔧 Mock | ❌ À implémenter | **Demo** |
| Coinbase | 🔧 Mock | 🔧 Stub | 🔧 Mock | ❌ À implémenter | **Demo** |

**Légende** :
- ✅ Complet : Implémentation fonctionnelle et testée
- 🔧 Partiel/Mock : Structure présente mais données mockées ou simulation
- ❌ À implémenter : Non implémenté

## ⚠️ AVERTISSEMENT CRITIQUE

**AUCUN EXCHANGE N'EST ACTUELLEMENT FONCTIONNEL POUR LE TRADING AUTOMATIQUE**

- **dYdX v4** : Mode lecture seule uniquement - Les fonctions de trading retournent des données simulées
- **Hyperliquid** : Données simulées uniquement
- **Coinbase** : Données simulées uniquement

**NE PAS UTILISER EN PRODUCTION POUR DU TRADING RÉEL**

---

## dYdX v4 ⚠️ LECTURE SEULE

### État : LECTURE SEULE (Trading NON implémenté)

**Fonctionnalités implémentées** :

✅ **Données de marché** :
- `GetTicker()` - Prix en temps réel depuis l'API
- `GetOrderBook()` - Carnet d'ordres avec profondeur
- `GetCandles()` - Données historiques OHLCV
- URL API: `https://indexer.dydx.trade`

✅ **WebSocket** :
- Ticker en temps réel
- Order book updates
- Trade feed
- URL WS: `wss://indexer.dydx.trade/v4/ws`

✅ **Authentification** :
- Support mnemonic (12/24 mots)
- Dérivation BIP44 (m/44'/118'/0'/0/0)
- Génération d'adresse Cosmos
- Support subaccounts

✅ **Compte** :
- `GetBalance()` - Balance USDC
- `GetPositions()` - Positions ouvertes
- Intégration avec subaccount

❌ **Trading** (NON IMPLÉMENTÉ) :
- ⚠️ `PlaceOrder()` retourne des données simulées (TODO ligne 258)
- ⚠️ `CancelOrder()` retourne succès sans action (TODO ligne 266)
- ⚠️ `GetOrder()` retourne des données simulées (TODO ligne 274)
- Infrastructure d'authentification présente mais **API de trading v4 non implémentée**
- **DANGER** : Le code peut sembler fonctionner mais n'exécute AUCUN ordre réel

### Configuration

```bash
# .env
EXCHANGE=dydx
EXCHANGE_API_SECRET="your twelve word mnemonic phrase here"
TRADING_SYMBOL=BTC-USD
```

### Documentation

- [Guide d'intégration](./DYDX_INTEGRATION.md)
- [Guide mnemonic](./DYDX_MNEMONIC_GUIDE.md)
- [Exemple](../examples/dydx_mnemonic_example.go)

## Hyperliquid 🔧 DEMO MODE

### État : Demo - Données mockées

**Fonctionnalités** :

🔧 **Données de marché** (MOCK) :
- `GetTicker()` - Retourne des données fixes
- `GetOrderBook()` - Retourne des données fixes
- `GetCandles()` - Retourne vide

✅ **WebSocket** (Partiel) :
- Structure présente
- Connexion possible
- Callbacks configurés

🔧 **Trading** (MOCK) :
- `PlaceOrder()` - Retourne succès sans API call
- `GetBalance()` - Retourne données fixes ($11,000)
- `GetPositions()` - Retourne vide

❌ **Authentification** :
- Pas d'implémentation de signature
- API key/secret non utilisés

### Pour implémenter

1. **API REST** :
   - Créer HTTP client
   - Implémenter endpoints Hyperliquid
   - Documentation: https://hyperliquid.gitbook.io/

2. **Authentification** :
   - Système de signature Ethereum
   - Headers spécifiques Hyperliquid

3. **WebSocket** :
   - Parser les messages Hyperliquid
   - Implémenter les subscriptions

### Utilisation actuelle

```bash
# Fonctionne mais avec données mockées
EXCHANGE=hyperliquid ./bin/constantine --headless
```

## Coinbase 🔧 DEMO MODE

### État : Demo - Structure minimale

**Fonctionnalités** :

🔧 **Toutes les méthodes** :
- Structure de base présente
- Toutes les méthodes retournent des mocks
- Interface `Exchange` implémentée

### Pour implémenter

1. **API REST Coinbase Advanced Trade** :
   - Documentation: https://docs.cloud.coinbase.com/advanced-trade-api/
   - Authentification: Clé privée ECDSA (PEM) + JWT ES256
   - Endpoints de trading
   - ✅ Balances réelles implémentées

2. **WebSocket** :
   - Coinbase Advanced Trade WebSocket
   - Subscriptions market data

### Utilisation actuelle

```bash
# Structure présente mais non fonctionnelle
# EXCHANGE=coinbase ./bin/constantine --headless
```

## Mode de fonctionnement actuel

### Sans credentials (Demo Mode)

```bash
# Aucune variable d'environnement
./bin/constantine --headless
```

**Comportement** :
- ⚠️ Avertissement au démarrage
- Données mockées pour Hyperliquid/Coinbase
- Erreur pour dYdX (nécessite mnemonic)

### Avec credentials dYdX (Production)

```bash
EXCHANGE=dydx
EXCHANGE_API_SECRET="word1 word2 ... word12"
./bin/constantine --headless
```

**Comportement** :
- ✅ Connexion réelle à dYdX
- ✅ Données de marché en temps réel
- ✅ Balance et positions réelles
- 🔧 Trading en lecture seule

## Recommandations

### Pour développement/tests

**Utiliser dYdX** :
- Données réelles de marché
- Pas de frais pour lecture
- Testnet disponible

### Pour backtesting

**Utiliser** :
- CSV de données historiques
- Exchange simulé du framework de backtesting

### Pour production

**Actuellement** :
- ✅ dYdX (lecture seule)
- ❌ Hyperliquid (nécessite implémentation)
- ❌ Coinbase (nécessite implémentation)

## Roadmap

### Court terme (prioritaire)

1. **dYdX Trading** :
   - [ ] Implémenter PlaceOrder avec signature
   - [ ] Implémenter CancelOrder
   - [ ] Tests end-to-end

2. **Hyperliquid** :
   - [ ] HTTP client REST
   - [ ] Authentification Ethereum
   - [ ] GetTicker réel
   - [ ] GetOrderBook réel

### Moyen terme

3. **WebSocket improvements** :
   - [ ] Reconnexion automatique
   - [ ] Gestion d'erreurs robuste
   - [ ] Heartbeat/ping-pong

4. **Coinbase** :
   - ✅ Balances réelles implémentées
   - [ ] API Advanced Trade complète
   - [ ] Authentification (clé privée ECDSA)
   - [ ] Endpoints de trading

### Long terme

5. **Nouveaux exchanges** :
   - [ ] Binance
   - [ ] Bybit
   - [ ] OKX

## Tests

### dYdX

```bash
# Test mnemonic
go run examples/dydx_mnemonic_example.go

# Test connexion
EXCHANGE=dydx EXCHANGE_API_SECRET="..." ./bin/constantine --headless

# Test backtesting
./bin/backtest --data=dydx_data.csv --symbol=BTC-USD
```

### Hyperliquid

```bash
# Demo mode uniquement
EXCHANGE=hyperliquid ./bin/constantine --headless
```

## Contribuer

Pour implémenter un exchange :

1. Créer le dossier `internal/exchanges/[exchange]/`
2. Implémenter l'interface `Exchange`
3. Ajouter HTTP client et WebSocket
4. Documenter dans ce fichier
5. Ajouter tests et exemples

Voir `internal/exchanges/dydx/` comme référence.

## Support

- **dYdX** : [Documentation officielle](https://docs.dydx.exchange/)
- **Hyperliquid** : [Gitbook](https://hyperliquid.gitbook.io/)
- **Coinbase** : [API Docs](https://docs.cloud.coinbase.com/)

## Avertissement

⚠️ **Important** :
- Toujours tester en mode démo avant production
- Commencer avec un capital limité
- Vérifier l'état de chaque exchange avant utilisation
- Les données mockées ne reflètent PAS le marché réel
