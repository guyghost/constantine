# Constantine Documentation Hub

Cette page fournit une vue d'ensemble des composants techniques du bot multi-agent Constantine ainsi que les pointeurs vers la documentation détaillée.

## 🧱 Architecture Principale

Constantine repose sur un ensemble d'agents coopératifs :

- **Agrégateur multi-exchange** (`internal/exchanges/aggregator.go`) : orchestre plusieurs clients d'exchange et consolide balances, positions et ordres.
- **Clients d'exchange** (`internal/exchanges/{dydx,hyperliquid,coinbase}/`) : implémentations spécifiques, avec dYdX v4 en lecture réelle (trading encore mock) et Hyperliquid/Coinbase en mode démo.
- **Stratégie Scalping** (`internal/strategy/`) : calcule les signaux via EMA/RSI/Bollinger Bands, configurable via variables d'environnement.
- **Gestionnaire d'ordres** (`internal/order/`) : ouvre, suit et clôture les positions tout en exposant des callbacks pour la stratégie, l'exécution et la TUI.
- **Gestionnaire de risque** (`internal/risk/`) : applique limites de drawdown, taille de position, cooldown et exposition par symbole.
- **Agent d'exécution** (`internal/execution/`) : automatise l'entrée/sortie selon la force du signal et injecte stop loss / take profit.
- **Interface Terminal (TUI)** (`internal/tui/`) : tableau de bord Bubble Tea affichant signaux, positions agrégées et statut des exchanges.
- **Télémétrie & Observabilité** (`internal/telemetry/metrics.go`) : serveur HTTP optionnel exposant `/metrics`, `/healthz`, `/readyz`.
- **Résilience** : modules `internal/circuitbreaker/` et `internal/ratelimit/` fournissent respectivement coupe-circuits et limiteurs de débit réutilisables.
- **Backtesting** (`internal/backtesting/` & `cmd/backtest/`) : moteur historique avec exchange simulé et reporting.

## 📁 Cartographie des dossiers

```
cmd/
  ├── bot/           # Point d'entrée du bot temps réel
  └── backtest/      # CLI de backtesting
internal/
  ├── exchanges/     # Interface commune + clients & agrégateur
  ├── strategy/      # Génération de signaux
  ├── order/         # Gestion des ordres/positions
  ├── risk/          # Limites de risque et suivi PnL
  ├── execution/     # Agent d'exécution automatisée
  ├── circuitbreaker/# Coupe-circuit générique
  ├── ratelimit/     # Limiteur token bucket & multi-limiter
  ├── telemetry/     # Serveur métriques Prometheus
  ├── tui/           # Interface Bubble Tea
  ├── backtesting/   # Moteur de simulation
  ├── logger/        # Configuration slog centralisée
  └── testutils/     # Helpers de tests
pkg/
  └── ...            # Packages utilitaires partagés
scripts/             # Scripts shell (ex: run_backtest)
docs/                # Documentation détaillée
```

## ⚙️ Configuration par variables d'environnement

Le bot charge automatiquement un fichier `.env` si présent. Les variables clés :

| Catégorie | Variables | Description |
|-----------|-----------|-------------|
| Activation exchanges | `ENABLE_DYDX`, `ENABLE_HYPERLIQUID`, `ENABLE_COINBASE` | Active ou désactive chaque client. Au moins un exchange doit être actif. |
| Auth dYdX | `DYDX_MNEMONIC`, `DYDX_SUBACCOUNT_NUMBER`, `DYDX_API_KEY`, `DYDX_API_SECRET` | Mnemonic recommandé (lecture réelle des comptes & marché). Les appels trading restent mock. |
| Trading | `TRADING_SYMBOL`, `INITIAL_BALANCE` | Symbole suivi par la stratégie et balance initiale du gestionnaire de risque. |
| Stratégie | `STRATEGY_SHORT_EMA`, `STRATEGY_LONG_EMA`, `STRATEGY_RSI_PERIOD`, `STRATEGY_TAKE_PROFIT`, `STRATEGY_STOP_LOSS`, `STRATEGY_MAX_POSITION_SIZE`, etc. | Surcharges des valeurs de `strategy.DefaultConfig()`. |
| Exécution | Config programmée (`execution.DefaultConfig()`) | AutoExecute activé, stop loss 0.5 %, take profit 1 %, seuil signal 0.5. |
| Logs | `LOG_LEVEL`, `LOG_FORMAT`, `LOG_ADD_SOURCE`, `LOG_OUTPUT_PATH` | Configurent `internal/logger`. |
| Observabilité | `TELEMETRY_ADDR` | Démarre le serveur métriques/healthcheck sur l'adresse fournie. |

Consultez `cmd/bot/main.go` pour la liste exhaustive et l'initialisation des agents.

## ▶️ Lancement rapide

```bash
# Installer les dépendances
go mod download

# Compiler les binaires
go build -o bin/constantine ./cmd/bot
go build -o bin/backtest ./cmd/backtest

# Démarrer en mode TUI
ENABLE_DYDX=true DYDX_MNEMONIC="word1 ... word12" ./bin/constantine

# Mode headless
./bin/constantine --headless

# Télémétrie (si TELEMETRY_ADDR défini)
curl -sf http://localhost:9100/metrics
```

Pour plus de détails :

- [README principal](../README.md) — onboarding & avertissements.
- [EXCHANGE_STATUS.md](./EXCHANGE_STATUS.md) — état fonctionnel de chaque intégration.
- [DYDX_INTEGRATION.md](./DYDX_INTEGRATION.md) — configuration dYdX.
- [BACKTESTING.md](./BACKTESTING.md) — moteur historique & CLI.
- [DEVELOPER.md](./DEVELOPER.md) — procédures de contribution.

## ✅ Tests recommandés

```bash
# Suite complète
go test ./...

# Couverture backtesting
go test ./internal/backtesting/...

# Vérification des clients d'exchange
go test ./internal/exchanges/... -run Test
```

## 🔎 Points de vigilance

- Le trading réel n'est **pas** encore implémenté pour dYdX/Hyperliquid/Coinbase (`internal/exchanges/*` retournent des mocks pour `PlaceOrder`).
- Le gestionnaire de risque applique des limites strictes : surveillez les logs lorsque `CanTrade()` renvoie `false`.
- Les callbacks de la stratégie et de l'agent d'exécution peuvent panic — les compteurs sont traqués via `internal/telemetry/metrics.go`.

Cette page sera mise à jour à mesure que les agents évoluent.
