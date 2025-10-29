# Constantine Trading Bot

[![CI](https://github.com/guyghost/constantine/workflows/CI/badge.svg)](https://github.com/guyghost/constantine/actions/workflows/ci.yml)
[![Security](https://github.com/guyghost/constantine/workflows/Security%20Scanning/badge.svg)](https://github.com/guyghost/constantine/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/guyghost/constantine)](https://goreportcard.com/report/github.com/guyghost/constantine)
[![codecov](https://codecov.io/gh/guyghost/constantine/branch/main/graph/badge.svg)](https://codecov.io/gh/guyghost/constantine)
[![Go Version](https://img.shields.io/github/go-mod/go-version/guyghost/constantine)](go.mod)
[![License](https://img.shields.io/github/license/guyghost/constantine)](LICENSE)

> **⚠️ AVERTISSEMENT CRITIQUE - LECTURE OBLIGATOIRE**
>
> **CE BOT N'EST PAS PRÊT POUR LE TRADING EN PRODUCTION**
>
> - **14 fonctionnalités critiques** ne sont pas implémentées (voir [docs/TODO.md](docs/TODO.md))
> - **Aucun exchange ne peut trader automatiquement** actuellement
> - **dYdX**: Données réelles mais trading en mode lecture seule uniquement
> - **Hyperliquid/Coinbase**: Données simulées uniquement
>
> **Utilisez ce bot uniquement pour:**
> - ✅ Backtesting de stratégies
> - ✅ Observation des marchés (dYdX)
> - ✅ Développement et tests
>
> **NE PAS utiliser pour du trading automatique réel!**
> Voir [EXCHANGE_STATUS.md](docs/EXCHANGE_STATUS.md) pour les détails complets.

---

Constantine est un bot de trading multi-agent pour les marchés de cryptomonnaies, construit avec une architecture modulaire permettant d'intégrer facilement différents exchanges. Le cœur du projet s'articule désormais autour d'un agrégateur multi-exchange, d'un moteur de stratégie temps réel et d'agents dédiés à l'exécution, au risque et à la télémétrie.

## 🎯 Caractéristiques

- **Multi-Exchange** : Agrégateur capable d'orchestrer dYdX v4, Hyperliquid, Coinbase
- **Architecture Agent-Based** : Agents dédiés (stratégie, risque, exécution, TUI, télémétrie)
- **Stratégie de Scalping** : EMA/RSI/Bollinger Bands avec seuils configurables via variables d'environnement
- **Backtesting Framework** : Testez vos stratégies sur données historiques avec 100% de taux de réussite validé (CLI `cmd/backtest`)
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

# Paramètres de gestion du risque
RISK_MAX_POSITION_SIZE=1000
RISK_MAX_POSITIONS=3
RISK_MAX_LEVERAGE=5
RISK_MAX_DAILY_LOSS=100
RISK_MAX_DRAWDOWN=10
RISK_PER_TRADE=1
RISK_MIN_ACCOUNT_BALANCE=20
RISK_DAILY_TRADING_LIMIT=50
RISK_COOLDOWN_PERIOD_MINUTES=15
RISK_CONSECUTIVE_LOSS_LIMIT=3
RISK_MAX_EXPOSURE_PER_SYMBOL=30
RISK_MAX_SAME_SYMBOL_POSITIONS=2

# Observabilité
TELEMETRY_ADDR=":9100"
LOG_LEVEL=debug
LOG_SENSITIVE_DATA=false
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
- [🔐 Gestion des secrets & CI](docs/SECRETS.md) • [CI Integration](docs/CI.md) • [CI/CD Pipeline](docs/CI_PIPELINE.md)
- [📊 État des exchanges](docs/EXCHANGE_STATUS.md)
- [🔒 Security Policy](SECURITY.md)

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
- ✅ Activez `LOG_SENSITIVE_DATA=false` pour protéger les données financières dans les logs

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

### Configuration du développement

Constantine utilise les meilleures pratiques Go pour le développement et l'intégration continue.

```bash
# Installer les dépendances
make install-deps

# Formatter le code
make fmt

# Lancer tous les checks locaux (comme CI)
make ci

# Exécuter les tests avec race detector
make test-race

# Lancer le linting
make lint

# Vérifier les vulnérabilités de sécurité
make vulncheck
```

### Intégration Continue (CI)

Le projet utilise GitHub Actions avec 5 workflows principaux pour assurer la qualité du code:

1. **CI Principal** - Validation, tests, linting, build multi-plateforme, et sécurité
2. **Security** - Scans de sécurité complets (govulncheck, gosec, Trivy, Nancy, SBOM, licenses)
3. **Benchmarks** - Suivi des performances et détection de régressions
4. **Code Quality** - Analyse de documentation, complexité, duplication, et code mort
5. **Release** - Builds automatisés multi-plateformes avec changelogs

**Fonctionnalités avancées:**
- ✅ Coverage automatique uploadé vers [Codecov](https://codecov.io/gh/guyghost/constantine)
- ✅ Seuil de couverture minimum de 40% (configurable)
- ✅ Génération de SBOM (Software Bill of Materials) pour la sécurité de la chaîne d'approvisionnement
- ✅ Vérification de conformité des licences
- ✅ Détection de régressions de performance (alerte à +150%)
- ✅ Dependabot pour mises à jour automatiques des dépendances
- ✅ Pre-commit hooks pour validation locale
- ✅ Scans de sécurité multi-couches quotidiens

Toutes les vérifications CI peuvent être exécutées localement:

```bash
# Exécuter TOUS les jobs CI localement
make ci

# Simuler les jobs individuels
make ci-validate    # Formatage, vet, mod
make ci-test        # Tests avec coverage
make ci-lint        # golangci-lint
make ci-build       # Build multi-plateforme
make ci-security    # Scans de vulnérabilités

# Nouvelles vérifications de qualité
make quality        # Code mort, duplication, complexité
make audit          # Audit de sécurité complet
make sbom           # Générer SBOM
make deadcode       # Détecter code inutilisé
make duplication    # Détecter code dupliqué
make complexity     # Analyser complexité

# Pre-commit hooks (nécessite Python)
pip install pre-commit
make pre-commit     # Setup une fois
pre-commit run --all-files
```

**Documentation complète:** Voir [docs/CI_PIPELINE.md](docs/CI_PIPELINE.md) pour les détails complets de la pipeline CI/CD.

La CI s'exécute automatiquement sur les branches `main` et toutes les pull requests. Les tests sont exécutés sur Go 1.23 et 1.24 pour assurer la compatibilité.

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

# Tests avec race detector (recommandé)
make test-race

# Tests avec coverage
make test-coverage

# Tests ciblés
go test ./internal/backtesting/...
go test ./internal/exchanges/... -run Test

# Vérifier la télémétrie
curl -sf http://localhost:9100/metrics
```

### Qualité du code

Le projet maintient des standards de qualité élevés:

- **Coverage**: Les rapports de coverage sont générés automatiquement et disponibles sur [Codecov](https://codecov.io/gh/guyghost/constantine). Les packages critiques (execution, risk, strategy, order, backtesting) visent un coverage minimum de 60%.
- **Linting**: golangci-lint avec 19 linters (govet, staticcheck, gosec, gocyclo, etc.)
- **Race detection**: Tous les tests CI utilisent `-race` pour détecter les problèmes de concurrence
- **Security**: Scan automatique des vulnérabilités avec govulncheck à chaque commit
- **Documentation**: godoc pour tous les packages publics


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
