# Liste des TODOs - Constantine Trading Bot

Ce document liste tous les TODOs dans le code avec leurs emplacements exacts et leur priorité d'implémentation.

**Dernière mise à jour:** 2025-10-25

## 📊 Vue d'ensemble

| Catégorie | Nombre | Complétés | Priorité | Statut |
|-----------|--------|-----------|----------|--------|
| dYdX WebSocket | 4 | 4 | 🟡 Moyenne | ✅ **COMPLÉTÉ (100%)** |
| Hyperliquid Trading | 8 | 3 | 🔴 Haute | 🟢 **En cours (38%)** |
| Coinbase Trading | 2 | 2 | 🟡 Moyenne | ✅ **COMPLÉTÉ (100%)** |
| **TOTAL** | **14** | **9** | - | **64% complété** ✅ |

### 🎉 Dernières implémentations (2025-10-25 Session 2)

**Coinbase (2/2 complétés):**
- ✅ GetOrderHistory - Historique des ordres via API Coinbase
- ✅ GetPosition - Récupération position par symbole

**dYdX WebSocket (4/4 complétés):**
- ✅ Message routing - Routing déjà fonctionnel
- ✅ Ticker parsing - Parsing oraclePrice, volume24H
- ✅ OrderBook parsing - Parsing bids/asks arrays
- ✅ Trade parsing - Parsing trades avec price/size/side

**Hyperliquid (3/8 complétés - Session 1):**
- ✅ GetBalance - Récupération balance réelle
- ✅ GetPositions - Récupération positions avec PnL
- ✅ GetOpenOrders - Liste des ordres ouverts

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

### 2. ✅ GetBalance - Hyperliquid (COMPLÉTÉ)
**Fichier:** `internal/exchanges/hyperliquid/client.go:567`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Endpoint: `/info` avec type: `clearinghouseState`
- Parsing de la structure marginSummary
- Calcul: Free = AccountValue - TotalMarginUsed
- Retourne balance USDC avec total/free/locked

### 3. ✅ GetPositions - Hyperliquid (COMPLÉTÉ)
**Fichier:** `internal/exchanges/hyperliquid/client.go:661`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Endpoint: `/info` avec type: `clearinghouseState`
- Parse assetPositions array
- Extraction: coin, size (szi), entryPrice, unrealizedPnL, leverage
- Filtre les positions à taille zéro
- Détermine side basé sur le signe de szi

### 4. ✅ GetOpenOrders - Hyperliquid (COMPLÉTÉ)
**Fichier:** `internal/exchanges/hyperliquid/client.go:544`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Endpoint: `/info` avec type: `openOrders`
- Parse array d'ordres: oid, coin, side, limitPx, sz, timestamp
- Filtre optionnel par symbol
- Convertit side ("B"/"A") vers OrderSideBuy/Sell
- Retourne liste d'ordres avec timestamps

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

## ✅ dYdX WebSocket - COMPLÉTÉ (4/4)

Tous les TODOs WebSocket dYdX sont maintenant implémentés avec parsing complet.

### 9. ✅ Message Routing - dYdX WS (COMPLÉTÉ)
**Fichier:** `internal/exchanges/dydx/websocket.go:130`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Routing déjà fonctionnel basé sur msg["type"] et msg["channel"]
- Switch case pour v4_markets, v4_orderbook, v4_trades
- Removed TODO comment, code is production-ready

### 10. ✅ Ticker Parsing - dYdX WS (COMPLÉTÉ)
**Fichier:** `internal/exchanges/dydx/websocket.go:158`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Parse dYdX v4 format avec oraclePrice comme last price
- Extract volume24H from trades24H field
- Approximate bid/ask from oracle price
- Proper decimal parsing avec error handling
- Invoke callbacks pour symbols enregistrés

### 11. ✅ OrderBook Parsing - dYdX WS (COMPLÉTÉ)
**Fichier:** `internal/exchanges/dydx/websocket.go:204`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Parse bids/asks arrays: [[price, size], ...]
- Convert string values to decimal.Decimal
- Build exchanges.OrderBook avec Bids et Asks levels
- Support pour incremental updates

### 12. ✅ Trade Parsing - dYdX WS (COMPLÉTÉ)
**Fichier:** `internal/exchanges/dydx/websocket.go:272`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Parse trades array from contents
- Extract price, size, side, createdAt timestamp
- Convert side strings (BUY/SELL) to exchanges.OrderSide
- Support multiple trades per message
- RFC3339 timestamp parsing

---

## ✅ Coinbase Trading - COMPLÉTÉ (2/2)

Tous les TODOs Coinbase sont maintenant implémentés.

**Note:** GetCandles et GetOrder étaient déjà implémentés. Les vrais TODOs étaient GetOrderHistory et GetPosition.

### 13. ✅ GetOrderHistory - Coinbase (COMPLÉTÉ)
**Fichier:** `internal/exchanges/coinbase/client.go:893`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Endpoint: GET `/brokerage/orders/historical/batch?limit={limit}&product_id={symbol}`
- Filtre automatique des ordres OPEN (gérés par GetOpenOrders)
- Parse order configuration (market/limit types)
- Extract filled amounts, average prices, timestamps
- Support symbol filtering et limit parameter

### 14. ✅ GetPosition - Coinbase (COMPLÉTÉ)
**Fichier:** `internal/exchanges/coinbase/client.go:1064`
**Statut:** ✅ **Implémenté le 2025-10-25**

**Implémentation:**
- Delegate à GetPositions() puis filtre par symbol
- Proper pour spot trading (pas de leverage)
- Returns error si aucune position trouvée pour le symbol
- Efficient: évite code duplication

---

## 🟢 Priorité Basse - Coinbase WebSocket (Non traité)

Le WebSocket Coinbase a une structure mais nécessite l'implémentation du parsing.

### 15. Message Routing - Coinbase WS (TODO)
**Fichier:** `internal/exchanges/coinbase/websocket.go:159`

**Description:** Router les messages Coinbase Advanced Trade WebSocket
- Types de messages: ticker, level2, heartbeats, etc.
- **Statut:** Non implémenté (basse priorité)

**Documentation:** [Coinbase WebSocket API](https://docs.cloud.coinbase.com/advanced-trade-api/docs/ws-overview)

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

**TODOs restants:** 6/15
**TODOs complétés:** 9/15
**Progression:** 60% ✅✅

**Session 1 (2025-10-25):** Hyperliquid Basics
- ✅ GetBalance (Hyperliquid)
- ✅ GetPositions (Hyperliquid)
- ✅ GetOpenOrders (Hyperliquid)

**Session 2 (2025-10-25):** Coinbase & dYdX WebSocket
- ✅ GetOrderHistory (Coinbase)
- ✅ GetPosition (Coinbase)
- ✅ Message Routing (dYdX WS)
- ✅ Ticker Parsing (dYdX WS)
- ✅ OrderBook Parsing (dYdX WS)
- ✅ Trade Parsing (dYdX WS)

**Prochaine étape:**
- Option 1: Implémenter signature Ethereum pour Hyperliquid PlaceOrder/CancelOrder (5 TODOs)
- Option 2: Implémenter Coinbase WebSocket parsing (1 TODO)

**Dernière révision:** 2025-10-25
