# Liste des TODOs - Constantine Trading Bot

Ce document liste tous les TODOs dans le code avec leurs emplacements exacts et leur priorité d'implémentation.

**Dernière mise à jour:** 2025-10-25

## 📊 Vue d'ensemble

| Catégorie | Nombre | Priorité | Statut |
|-----------|--------|----------|--------|
| dYdX WebSocket | 4 | 🟡 Moyenne | Non démarré |
| Hyperliquid Trading | 8 | 🔴 Haute | Non démarré |
| Coinbase Trading | 2 | 🟡 Moyenne | Non démarré |
| **TOTAL** | **14** | - | - |

---

## 🔴 Priorité Haute - dYdX Trading (Bloquant)

Ces TODOs doivent être implémentés en priorité car dYdX est le seul exchange avec données réelles.

### Aucun TODO haute priorité dans cette catégorie
Les fonctions de trading dYdX sont déjà marquées dans `EXCHANGE_STATUS.md` mais ne sont pas marquées TODO dans le code.

---

## 🟠 Priorité Haute - Hyperliquid Core Trading

Ces fonctions bloquent l'utilisation de Hyperliquid pour du trading réel.

### 1. PlaceOrder - Hyperliquid
**Fichier:** `internal/exchanges/hyperliquid/client.go:511`
```go
// TODO: Implement authentication and real API call
```

**Description:** Implémentation complète de placement d'ordre
- Signature Ethereum des transactions
- Appel à l'API Hyperliquid `/exchange`
- Gestion des réponses et erreurs

**Ressources:**
- [Hyperliquid API Docs](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/exchange-endpoint)
- Authentification via signature Ethereum (secp256k1)

### 2. GetBalance - Hyperliquid
**Fichier:** `internal/exchanges/hyperliquid/client.go:535`
```go
// TODO: Implement authentication and real API call
```

**Description:** Récupération du solde réel du compte
- Endpoint: `/info` avec type: `clearinghouseState`
- Parsing de la structure de réponse Hyperliquid

### 3. GetPositions - Hyperliquid
**Fichier:** `internal/exchanges/hyperliquid/client.go:568`
```go
// TODO: Implement authentication and real API call
```

**Description:** Récupération des positions ouvertes
- Endpoint: `/info` avec type: `clearinghouseState`
- Extraction des positions actives

### 4. GetOpenOrders - Hyperliquid
**Fichier:** `internal/exchanges/hyperliquid/client.go:612`
```go
// TODO: Implement authentication and real API call
```

**Description:** Liste des ordres ouverts
- Endpoint: `/info` avec type: `openOrders`

---

## 🟡 Priorité Moyenne - Hyperliquid Market Data

Ces fonctions sont nécessaires pour avoir des données de marché réelles.

### 5. GetTicker - Hyperliquid
**Fichier:** `internal/exchanges/hyperliquid/client.go:523`
```go
// TODO: Implement REST API call
```

**Description:** Prix en temps réel
- Endpoint: `/info` avec type: `allMids`
- Parsing des données ticker

### 6. GetOrderBook - Hyperliquid
**Fichier:** `internal/exchanges/hyperliquid/client.go:529`
```go
// TODO: Implement REST API call
```

**Description:** Carnet d'ordres
- Endpoint: `/info` avec type: `l2Book`
- Parsing bids/asks

### 7. GetCandles - Hyperliquid
**Fichier:** `internal/exchanges/hyperliquid/client.go:554`
```go
// TODO: Implement REST API call
```

**Description:** Données OHLCV historiques
- Endpoint: `/info` avec type: `candleSnapshot`

### 8. GetOrder - Hyperliquid
**Fichier:** `internal/exchanges/hyperliquid/client.go:641`
```go
// TODO: Implement REST API call
```

**Description:** Statut d'un ordre spécifique
- Endpoint: `/info` avec type: `orderStatus`

---

## 🟡 Priorité Moyenne - dYdX WebSocket

Ces TODOs concernent le parsing des messages WebSocket dYdX. Le WebSocket est déjà connecté, mais les messages ne sont pas parsés correctement.

### 9. Message Routing - dYdX WS
**Fichier:** `internal/exchanges/dydx/websocket.go:130`
```go
// TODO: Implement proper message routing based on dYdX's protocol
```

**Description:** Router les messages WebSocket vers les bons handlers
- Identifier le type de message (ticker, orderbook, trades, etc.)
- Dispatcher vers le handler approprié

**Documentation:** [dYdX WebSocket Protocol](https://docs.dydx.exchange/developers/indexer/indexer_websocket)

### 10. Ticker Parsing - dYdX WS
**Fichier:** `internal/exchanges/dydx/websocket.go:159`
```go
// TODO: Parse ticker data according to dYdX format
```

**Description:** Parser les données ticker du WebSocket
- Format: `{"type": "channel_data", "channel": "v4_markets",...}`
- Extraction: price, volume, high, low, etc.

### 11. OrderBook Parsing - dYdX WS
**Fichier:** `internal/exchanges/dydx/websocket.go:184`
```go
// TODO: Parse order book data according to dYdX format
```

**Description:** Parser les mises à jour du carnet d'ordres
- Format: `{"type": "channel_data", "channel": "v4_orderbook",...}`
- Gestion des updates incrémentaux

### 12. Trade Parsing - dYdX WS
**Fichier:** `internal/exchanges/dydx/websocket.go:206`
```go
// TODO: Parse trade data according to dYdX format
```

**Description:** Parser les trades exécutés
- Format: `{"type": "channel_data", "channel": "v4_trades",...}`
- Extraction: price, size, side, timestamp

---

## 🟢 Priorité Basse - Coinbase WebSocket

Le WebSocket Coinbase a une structure mais nécessite l'implémentation du parsing.

### 13. Message Routing - Coinbase WS
**Fichier:** `internal/exchanges/coinbase/websocket.go:159`
```go
// TODO: Implement proper message routing based on Coinbase's protocol
```

**Description:** Router les messages Coinbase Advanced Trade WebSocket
- Types de messages: ticker, level2, heartbeats, etc.

**Documentation:** [Coinbase WebSocket API](https://docs.cloud.coinbase.com/advanced-trade-api/docs/ws-overview)

---

## 🟢 Priorité Basse - Coinbase Trading

### 14. GetCandles - Coinbase
**Fichier:** `internal/exchanges/coinbase/client.go:894`
```go
// TODO: Implement REST API call
```

**Description:** Données OHLCV Coinbase
- Endpoint: `/api/v3/brokerage/products/{product_id}/candles`

### 15. GetOrder - Coinbase
**Fichier:** `internal/exchanges/coinbase/client.go:999`
```go
// TODO: Implement REST API call
```

**Description:** Statut d'un ordre
- Endpoint: `/api/v3/brokerage/orders/historical/{order_id}`

---

## 🎯 Plan d'Implémentation Recommandé

### Phase 1: Hyperliquid Core (2-3 semaines)
1. ✅ Créer HTTP client avec rate limiting
2. ⬜ Implémenter authentification Ethereum
3. ⬜ GetTicker (données marché)
4. ⬜ GetOrderBook (carnet d'ordres)
5. ⬜ GetBalance (soldes)
6. ⬜ PlaceOrder (trading)
7. ⬜ Tests d'intégration avec testnet

### Phase 2: dYdX WebSocket (1 semaine)
1. ⬜ Message routing
2. ⬜ Ticker parsing
3. ⬜ OrderBook parsing
4. ⬜ Trade parsing
5. ⬜ Tests avec données réelles

### Phase 3: Coinbase (2 semaines)
1. ⬜ GetCandles
2. ⬜ GetOrder
3. ⬜ WebSocket message routing
4. ⬜ Tests d'intégration

---

## 📝 Template pour Implémentation

Lorsque vous implémentez un TODO:

```go
// AVANT
func (c *Client) PlaceOrder(...) (*exchanges.Order, error) {
    // TODO: Implement authentication and real API call
    return nil, fmt.Errorf("not implemented")
}

// APRÈS
func (c *Client) PlaceOrder(ctx context.Context, req *exchanges.OrderRequest) (*exchanges.Order, error) {
    // Validate request
    if err := validateOrderRequest(req); err != nil {
        return nil, fmt.Errorf("invalid order request: %w", err)
    }

    // Apply rate limiting
    if err := c.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }

    // Create request payload
    payload := createOrderPayload(req)

    // Sign request (exchange-specific)
    signature, err := c.signRequest(payload)
    if err != nil {
        return nil, fmt.Errorf("signature failed: %w", err)
    }

    // Execute API call
    var response OrderResponse
    if err := c.post(ctx, "/v1/order", payload, &response); err != nil {
        telemetry.RecordOrderError("place_order_failed")
        return nil, fmt.Errorf("API error: %w", err)
    }

    // Convert to common format
    order := convertToOrder(response)
    telemetry.RecordOrder(order.Symbol, order.Side)

    return order, nil
}
```

---

## ✅ Checklist pour chaque TODO

Avant de marquer un TODO comme complété:

- [ ] Fonction implémentée avec gestion d'erreurs
- [ ] Rate limiting appliqué
- [ ] Métriques Prometheus ajoutées
- [ ] Tests unitaires écrits
- [ ] Test d'intégration avec exchange réel (ou testnet)
- [ ] Documentation mise à jour
- [ ] Exemple d'utilisation fourni
- [ ] Revue de code effectuée

---

## 🔗 Ressources

### Documentation Exchanges
- [dYdX v4 API](https://docs.dydx.exchange/)
- [Hyperliquid Docs](https://hyperliquid.gitbook.io/)
- [Coinbase Advanced Trade](https://docs.cloud.coinbase.com/advanced-trade-api/)

### Fichiers Référence
- `internal/exchanges/interface.go` - Interface Exchange
- `internal/exchanges/dydx/client.go` - Implémentation de référence
- `docs/EXCHANGE_STATUS.md` - État des exchanges

### Outils Utiles
- Testnet dYdX: https://testnet.dydx.exchange/
- Hyperliquid Testnet: https://app.hyperliquid-testnet.xyz/
- Coinbase Sandbox: https://docs.cloud.coinbase.com/advanced-trade-api/docs/sandbox

---

## 📊 Suivi

**TODOs restants:** 14/14
**TODOs complétés:** 0/14
**Progression:** 0%

**Dernière révision:** 2025-10-25
