# Constantine Trading Bot

Constantine est un bot de trading multi-agent pour les marchés de cryptomonnaies, construit avec une architecture modulaire permettant d'intégrer facilement différents exchanges. Le cœur du projet s'articule désormais autour d'un agrégateur multi-exchange, d'un moteur de stratégie temps réel et d'agents dédiés à l'exécution, au risque et à la télémétrie.

## 🎯 Caractéristiques

- **Multi-Exchange** : Agrégateur capable d'orchestrer dYdX v4, Hyperliquid, Coinbase
- **Architecture Agent-Based** : Agents dédiés (stratégie, risque, exécution, TUI, télémétrie)
- **Stratégie de Scalping** : EMA/RSI/Bollinger Bands avec seuils configurables via variables d'environnement
- **Backtesting Framework** : Testez vos stratégies sur données historiques (CLI `cmd/backtest`)
- **Agent d'exécution** : Gestion automatique des entrées/sorties avec stop loss & take profit
- **TUI & Headless Mode** : Interface terminal (Bubble Tea) ou mode headless pour serveurs
- **Gestion du risque** : Limites de positions, drawdown, cooldown, exposition par symbole
- **Observabilité** : Export Prometheus (`/metrics`), endpoints de santé `/healthz` & `/readyz`

## 📊 État des Exchanges

| Exchange | Statut | Authentification | Documentation |
|----------|--------|------------------|---------------|
| **dYdX v4** | ✅ Temps réel (REST/WS) | ⚠️ Trading mock | Mnemonic (ou API key) | [Guide](docs/DYDX_INTEGRATION.md) |
| Hyperliquid | 🔧 Mock data | 🔧 Mock | À implémenter | - |
| Coinbase | 🔧 Mock data | 🔧 Mock | À implémenter | - |

**⚠️ AVERTISSEMENT IMPORTANT** :
- **dYdX** : Les fonctions de trading (PlaceOrder, CancelOrder) retournent encore des mocks, même si les données de marché temps réel et les comptes sont fonctionnels.
- **NE PAS UTILISER EN PRODUCTION** pour du trading automatique.

Voir [EXCHANGE_STATUS.md](docs/EXCHANGE_STATUS.md) pour plus de détails.

## 🚀 Démarrage Rapide

### Prérequis

- Go 1.21+
- Compte sur un exchange supporté (recommandé: dYdX)
- Accès à un mnemonic dYdX (ou clés API) si vous activez l'exchange en temps réel

### Installation

```bash
# Cloner le repository
git clone https://github.com/guyghost/constantine
cd constantine

# Installer les dépendances
go mod download

# Compiler le bot multi-agent
go build -o bin/constantine ./cmd/bot

# Compiler l'outil de backtesting
go build -o bin/backtest ./cmd/backtest
```

### Configuration

Créez un fichier `.env` à la racine :

```bash
# Activer les exchanges
ENABLE_DYDX=true
ENABLE_HYPERLIQUID=false
ENABLE_COINBASE=false

# Authentification dYdX (lecture seule)
DYDX_MNEMONIC="word1 word2 ... word12"
DYDX_SUBACCOUNT_NUMBER=0

# Configuration de trading
TRADING_SYMBOL=BTC-USD
INITIAL_BALANCE=10000

# Paramètres de stratégie (override des valeurs par défaut)
STRATEGY_SHORT_EMA=9
STRATEGY_LONG_EMA=21
STRATEGY_RSI_PERIOD=14
STRATEGY_TAKE_PROFIT=0.5
STRATEGY_STOP_LOSS=0.25

# Observabilité
TELEMETRY_ADDR=":9100"
LOG_LEVEL=debug
```

⚠️ **Important** : Ajoutez `.env` à votre `.gitignore` !

> 💡 Pour éviter de stocker les clés en clair, utilisez l'intégration 1Password
> décrite dans [docs/SECRETS.md](docs/SECRETS.md) avec le template
> `.env.op.template` et `op run`.

### Lancer le Bot

```bash
# Mode headless (recommandé pour serveurs)
./bin/constantine --headless

# Mode TUI (interface terminal Bubble Tea)
./bin/constantine
```

> ℹ️ Le bot démarre un serveur de télémétrie si `TELEMETRY_ADDR` est défini :
> - `/metrics` (Prometheus)
> - `/healthz` (liveness)
> - `/readyz` (readiness)

## 📖 Documentation

### Guides principaux

- [🚀 Guide de démarrage rapide](docs/QUICKSTART.md)
- [🏗️ Architecture multi-agents](AGENTS.md)
- [📊 État des exchanges](docs/EXCHANGE_STATUS.md)

### Exchanges

- [dYdX Integration](docs/DYDX_INTEGRATION.md)
- [dYdX Mnemonic Guide](docs/DYDX_MNEMONIC_GUIDE.md)

### Fonctionnalités

- [Backtesting Framework](docs/BACKTESTING.md)
- [Gestion du risque](docs/EXCHANGE_STATUS.md#recommandations) *(résumé dans le manager de risque)*
- [Architecture multi-agents](AGENTS.md)

## 🧪 Backtesting

Testez vos stratégies avant de les déployer :

```bash
# Avec vos données CSV
./bin/backtest --data=historical_data.csv --symbol=BTC-USD --verbose

# Avec données générées (test)
./bin/backtest --generate-sample --sample-candles=1000

# Utiliser le script helper
./scripts/run_backtest.sh data.csv
```

Voir [BACKTESTING.md](docs/BACKTESTING.md) pour plus de détails.

## 🏗️ Architecture

Constantine utilise une architecture basée sur des agents autonomes :

```
┌─────────────────────────────────────────┐
│            TUI Agent                     │
│  (Interface utilisateur)                 │
└───────────────┬─────────────────────────┘
                │
┌───────────────┴─────────────────────────┐
│    Agrégateur Multi-Exchange            │
│  (dYdX, Hyperliquid, Coinbase)          │
└───────────────┬─────────────────────────┘
                │
┌───────────────┴─────────────────────────┐
│         Strategy Agent                   │
│  (EMA/RSI/Bollinger Bands)              │
└───────────────┬─────────────────────────┘
                │
┌───────────────┴─────────────────────────┐
│  Order / Risk / Execution Agents         │
│  (SL/TP automatiques, validation risque) │
└─────────────────────────────────────────┘
```

Voir [AGENTS.md](AGENTS.md) pour l'architecture complète.

## 🔐 Sécurité

### Bonnes pratiques

- ✅ Stockez vos credentials dans `.env` (jamais dans le code)
- ✅ Ajoutez `.env` au `.gitignore`
- ✅ Utilisez un wallet dédié avec capital limité
- ✅ Testez toujours en backtesting avant production
- ✅ Commencez avec de petites positions

### ⚠️ Avertissements Critiques

**SÉCURITÉ** :
- ⚠️ **NE JAMAIS** committer votre fichier `.env` contenant vos mnémoniques/clés privées
- ⚠️ Les mnémoniques donnent **accès complet** à vos fonds - protégez-les comme de l'argent liquide
- ⚠️ Utilisez un gestionnaire de secrets (Vault, AWS Secrets Manager) en production
- ⚠️ Commencez avec un wallet dédié et un capital limité que vous pouvez perdre

**TRADING** :
- ⚠️ **Le trading de cryptomonnaies comporte des risques importants**
- Vous pouvez perdre **tout votre capital**
- Les marchés sont extrêmement volatils
- **Testez toujours en backtesting et mode démo avant production**
- Ce bot est fourni "tel quel" **sans aucune garantie**
- Les auteurs ne sont **pas responsables** des pertes financières

## 📁 Structure du Projet

```
constantine/
├── cmd/
│   ├── bot/          # Application principale
│   └── backtest/     # Outil de backtesting
├── internal/
│   ├── exchanges/      # Adaptateurs exchanges + agrégateur multi-exchange
│   ├── strategy/       # Stratégies de trading (scalping)
│   ├── order/          # Gestion des ordres & positions
│   ├── risk/           # Gestion du risque et exposure
│   ├── execution/      # Agent d'exécution automatique
│   ├── circuitbreaker/ # Protection contre les défaillances
│   ├── ratelimit/      # Limiteurs de taux token bucket
│   ├── telemetry/      # Serveur métriques & santé
│   ├── tui/            # Interface terminal Bubble Tea
│   ├── backtesting/    # Framework de backtesting
│   ├── logger/         # Wrapper slog + configuration
│   └── testutils/      # Helpers pour tests
├── pkg/               # Packages réutilisables (utils, etc.)
├── docs/              # Documentation détaillée
├── scripts/           # Scripts utilitaires
└── testdata/          # Données de test
```

## 🛠️ Développement

### Ajouter un nouvel exchange

1. Créer le dossier `internal/exchanges/[exchange]/`
2. Implémenter l'interface `Exchange`
3. Ajouter HTTP client et WebSocket
4. Documenter l'intégration
5. Ajouter des tests

Voir `internal/exchanges/dydx/` comme référence.

### Tests

```bash
# Tests unitaires
go test ./...

# Tests ciblés
go test ./internal/backtesting/...
go test ./internal/exchanges/... -run Test

# Vérifier la télémétrie
curl -sf http://localhost:9100/metrics
```

## 🤝 Contribution

Les contributions sont les bienvenues ! Pour contribuer :

1. Fork le projet
2. Créez une branche feature (`git checkout -b feature/AmazingFeature`)
3. Committez vos changements (`git commit -m 'Add AmazingFeature'`)
4. Push vers la branche (`git push origin feature/AmazingFeature`)
5. Ouvrez une Pull Request

## 📝 Roadmap

### Court terme

- [x] Support dYdX v4
- [x] Framework de backtesting
- [x] Authentification mnemonic
- [ ] Trading automatique dYdX (REST v4)
- [ ] Implémentation complète Hyperliquid (REST & signatures)

### Moyen terme

- [ ] Support Binance
- [ ] Dashboard web
- [ ] Notifications (Telegram, Discord)
- [ ] Multiples stratégies simultanées
- [ ] Paper trading mode

### Long terme

- [ ] Machine Learning pour optimisation
- [ ] Support DeFi protocols
- [ ] Mobile app
- [ ] Cloud deployment templates

## 📜 License

MIT License - voir [LICENSE](LICENSE) pour plus de détails.

## 📞 Support

- **Documentation** : [docs/](docs/)
- **Issues** : [GitHub Issues](https://github.com/guyghost/constantine/issues)
- **Discussions** : [GitHub Discussions](https://github.com/guyghost/constantine/discussions)

## 🙏 Remerciements

- [dYdX](https://dydx.exchange/) pour leur excellente documentation
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) pour le TUI framework
- La communauté Go crypto

---

**⚠️ Disclaimer** : Ce projet est à des fins éducatives et de recherche. Utilisez-le à vos propres risques. Les auteurs ne sont pas responsables des pertes financières.
