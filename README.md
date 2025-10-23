# Constantine Trading Bot

Constantine est un bot de trading multi-agent pour les marchés de cryptomonnaies, construit avec une architecture modulaire permettant d'intégrer facilement différents exchanges.

## 🎯 Caractéristiques

- **Multi-Exchange** : Support de dYdX v4, Hyperliquid, Coinbase
- **Architecture Agent-Based** : Composants modulaires et découplés
- **Stratégie de Scalping** : EMA, RSI, et Bollinger Bands
- **Backtesting Framework** : Testez vos stratégies sur données historiques
- **TUI & Headless Mode** : Interface terminal ou mode sans interface
- **Gestion du risque** : Position sizing, stop loss, take profit
- **WebSocket Real-time** : Données de marché en temps réel

## 📊 État des Exchanges

| Exchange | Statut | Authentification | Documentation |
|----------|--------|------------------|---------------|
| **dYdX v4** | ✅ **Production Ready** | Mnemonic | [Guide](docs/DYDX_INTEGRATION.md) |
| Hyperliquid | 🔧 Demo Mode | À implémenter | - |
| Coinbase | 🔧 Demo Mode | À implémenter | - |

Voir [EXCHANGE_STATUS.md](docs/EXCHANGE_STATUS.md) pour plus de détails.

## 🚀 Démarrage Rapide

### Prérequis

- Go 1.21+
- Compte sur un exchange supporté (recommandé: dYdX)

### Installation

```bash
# Cloner le repository
git clone https://github.com/guyghost/constantine
cd constantine

# Installer les dépendances
go mod download

# Compiler le bot
go build -o bin/constantine ./cmd/bot

# Compiler l'outil de backtesting
go build -o bin/backtest ./cmd/backtest
```

### Configuration

Créez un fichier `.env` à la racine :

```bash
# Exchange selection (dydx recommandé)
EXCHANGE=dydx

# Pour dYdX: utilisez votre mnemonic (12 ou 24 mots)
EXCHANGE_API_SECRET="word1 word2 word3 ... word12"

# Configuration de trading
TRADING_SYMBOL=BTC-USD
INITIAL_BALANCE=10000

# Paramètres de stratégie
SHORT_EMA_PERIOD=9
LONG_EMA_PERIOD=21
RSI_PERIOD=14
TAKE_PROFIT_PERCENT=0.5
STOP_LOSS_PERCENT=0.25
```

⚠️ **Important** : Ajoutez `.env` à votre `.gitignore` !

### Lancer le Bot

```bash
# Mode headless (recommandé pour serveurs)
./bin/constantine --headless

# Mode TUI (interface terminal)
./bin/constantine
```

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
- [Gestion du risque](docs/RISK_MANAGEMENT.md) *(à venir)*
- [Stratégies de trading](docs/STRATEGIES.md) *(à venir)*

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
│         Exchange Agents                  │
│  (dYdX, Hyperliquid, Coinbase)          │
└───────────────┬─────────────────────────┘
                │
┌───────────────┴─────────────────────────┐
│         Strategy Agent                   │
│  (EMA/RSI/Bollinger Bands)              │
└───────────────┬─────────────────────────┘
                │
┌───────────────┴─────────────────────────┐
│      Order & Risk Management             │
│  (Position sizing, SL/TP)               │
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

### Avertissement

⚠️ **Le trading de cryptomonnaies comporte des risques** :
- Vous pouvez perdre tout votre capital
- Les marchés sont volatils
- Testez toujours en démo avant production
- Ce bot est fourni "tel quel" sans garantie

## 📁 Structure du Projet

```
constantine/
├── cmd/
│   ├── bot/          # Application principale
│   └── backtest/     # Outil de backtesting
├── internal/
│   ├── exchanges/    # Adaptateurs exchanges
│   │   ├── dydx/     # ✅ dYdX v4 (production ready)
│   │   ├── hyperliquid/
│   │   └── coinbase/
│   ├── strategy/     # Stratégies de trading
│   ├── order/        # Gestion des ordres
│   ├── risk/         # Gestion du risque
│   ├── tui/          # Interface terminal
│   └── backtesting/  # Framework de backtesting
├── docs/             # Documentation
├── examples/         # Exemples de code
├── scripts/          # Scripts utilitaires
└── testdata/         # Données de test
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

# Test spécifique
go test ./internal/backtesting/...

# Test d'intégration dYdX
go run examples/dydx_mnemonic_example.go
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
- [ ] Trading automatique dYdX
- [ ] Implémentation complète Hyperliquid

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
