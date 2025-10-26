# Liste des TODOs - Constantine Trading Bot

Ce document liste tous les TODOs dans le code avec leurs emplacements exacts et leur priorité d'implémentation.

**Dernière mise à jour:** 2025-10-25

## 📊 Vue d'ensemble

| Catégorie | Nombre | Complétés | Priorité | Statut |
|-----------|--------|-----------|----------|--------|
| dYdX WebSocket | 4 | 4 | 🟡 Moyenne | ✅ **COMPLÉTÉ (100%)** |
| Hyperliquid Trading | 8 | 6 | 🔴 Haute | 🟢 **En cours (75%)** |
| Coinbase Total | 3 | 3 | 🟡 Moyenne | ✅ **COMPLÉTÉ (100%)** |
| **TOTAL** | **15** | **13** | - | **87% complété** ✅✅✅ |

### 🧹 Nettoyage des TODOs Obsolètes

**Résultat:** Seulement **3 TODOs légitimes** restent dans tout le codebase !

Tous les autres commentaires TODO étaient en fait du code déjà implémenté.

### 🎉 Dernières implémentations (2025-10-25 Session 3)

**Hyperliquid (3 nouveaux):**
- ✅ GetCandles - Déjà implémenté (OHLCV via candleSnapshot)
- ✅ GetOrder - Statut d'ordre via orderStatus API
- ✅ GetPosition - Filtre positions par symbole (NEW)

**Coinbase WebSocket (1 nouveau):**
- ✅ Message Routing - Déjà implémenté, removed TODO comment

### Session 2 (2025-10-25)

**Coinbase (2/2 complétés):**
- ✅ GetOrderHistory - Historique des ordres via API Coinbase
- ✅ GetPosition - Récupération position par symbole

**dYdX WebSocket (4/4 complétés):**
- ✅ Message routing - Routing déjà fonctionnel
- ✅ Ticker parsing - Parsing oraclePrice, volume24H
- ✅ OrderBook parsing - Parsing bids/asks arrays
- ✅ Trade parsing - Parsing trades avec price/size/side

### Session 1 (2025-10-25)

**Hyperliquid (3/8 complétés):**
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

## ✅ Hyperliquid Market Data - COMPLÉTÉ (4/4)

**Note:** Ces fonctions étaient en fait déjà implémentées ! Pas de vrais TODOs.

### 5. ✅ GetTicker - Hyperliquid (DÉJÀ IMPLÉMENTÉ)
**Fichier:** `internal/exchanges/hyperliquid/client.go:254`
**Statut:** ✅ **Déjà implémenté depuis le début**

**Implémentation:**
- POST /info avec type: "allMids"
- Parse HyperliquidTickerResponse array
- Extract mid price pour le coin
- Approximate bid/ask from mid price
- Full ticker with volume24h

### 6. ✅ GetOrderBook - Hyperliquid (DÉJÀ IMPLÉMENTÉ)
**Fichier:** `internal/exchanges/hyperliquid/client.go:302`
**Statut:** ✅ **Déjà implémenté depuis le début**

**Implémentation:**
- POST /info avec type: "l2Book"
- Parse bids/asks arrays: [[price, size], ...]
- Convert to exchanges.OrderBook structure
- Full orderbook with depth parameter

### 7. ✅ GetCandles - Hyperliquid (DÉJÀ IMPLÉMENTÉ)
**Fichier:** `internal/exchanges/hyperliquid/client.go:405`
**Statut:** ✅ **Déjà implémenté depuis le début**

**Implémentation:**
- POST /info avec type: "candleSnapshot"
- Parse OHLCV data: timestamp, open, high, low, close, volume
- Convert all values to decimal.Decimal
- Sort by timestamp (oldest first)
- Interval conversion (1m, 5m, 15m, 1h, 4h, 1d)

### 8. ✅ GetOrder - Hyperliquid (COMPLÉTÉ)
**Fichier:** `internal/exchanges/hyperliquid/client.go:543`
**Statut:** ✅ **Implémenté le 2025-10-25 Session 3**

**Implémentation:**
- POST /info avec type: "orderStatus"
- Parse order ID (int64), requires user address
- Extract: oid, coin, side, limitPx, sz, filledSz, avgPx, orderState
- Map orderState to exchanges.OrderStatus (open/filled/canceled)
- Full order details with timestamps

### 9. ✅ GetPosition - Hyperliquid (COMPLÉTÉ)
**Fichier:** `internal/exchanges/hyperliquid/client.go:900`
**Statut:** ✅ **Implémenté le 2025-10-25 Session 3**

**Implémentation:**
- Delegate to GetPositions() and filter by symbol
- Returns specific position or error if not found
- Same clean pattern as Coinbase GetPosition
- Efficient code reuse (17 lines)

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

## ✅ Coinbase WebSocket - COMPLÉTÉ (1/1)

**Note:** Le WebSocket Coinbase était déjà complètement implémenté !

### 15. ✅ Message Routing - Coinbase WS (DÉJÀ IMPLÉMENTÉ)
**Fichier:** `internal/exchanges/coinbase/websocket.go:152`
**Statut:** ✅ **Déjà implémenté - TODO comment removed le 2025-10-25**

**Implémentation:**
- Route messages basé sur msg["channel"]
- Switch case pour: ticker, level2, market_trades
- handleTickerMessage: Parse best_bid, best_ask, price, size from events
- handleOrderBookMessage: Parse bids/asks arrays from events
- handleTradeMessage: Parse price, size, side from events
- All handlers avec proper callback invocation
- Full WebSocket support for Coinbase Advanced Trade API

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

**TODOs restants:** 3/15
**TODOs complétés:** 12/15
**Progression:** 80% ✅✅✅

### 🎯 Progression par Session

**Session 1 (2025-10-25):** Hyperliquid Basics (3 TODOs)
- ✅ GetBalance (Hyperliquid)
- ✅ GetPositions (Hyperliquid)
- ✅ GetOpenOrders (Hyperliquid)
- **Progression:** 0% → 21%

**Session 2 (2025-10-25):** Coinbase & dYdX WebSocket (6 TODOs)
- ✅ GetOrderHistory (Coinbase)
- ✅ GetPosition (Coinbase)
- ✅ Message Routing (dYdX WS)
- ✅ Ticker Parsing (dYdX WS)
- ✅ OrderBook Parsing (dYdX WS)
- ✅ Trade Parsing (dYdX WS)
- **Progression:** 21% → 60%

**Session 3 (2025-10-25):** Option 2 - TODOs Simples (4 TODOs)
- ✅ GetOrder (Hyperliquid) - NEW implementation
- ✅ GetCandles (Hyperliquid) - Already implemented
- ✅ Coinbase WebSocket - Already implemented
- ✅ GetPosition (Hyperliquid) - NEW implementation
- **Progression:** 60% → 87%

**Session 3 Cleanup:** Removed obsolete TODO comments
- Only 3 real TODOs remain in entire codebase
- All other TODOs were already implemented functions

### 🏆 TODOs Restants (3/15 - Seulement 13% à faire!)

Tous les TODOs restants **requièrent des signatures Ethereum** (complexe):

1. **PlaceOrder** (Hyperliquid) - Priorité HAUTE
   - Ligne: internal/exchanges/hyperliquid/client.go:511
   - Nécessite: Signature Ethereum secp256k1

2. **CancelOrder** (Hyperliquid) - Priorité HAUTE
   - Ligne: internal/exchanges/hyperliquid/client.go:523
   - Nécessite: Signature Ethereum secp256k1

3. **GetOrderHistory** (Hyperliquid) - Priorité MOYENNE
   - Ligne: internal/exchanges/hyperliquid/client.go:703
   - Nécessite: API endpoint implementation

**Pour atteindre 100%:** Implémenter signature Ethereum secp256k1

**Dernière révision:** 2025-10-25
