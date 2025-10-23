# Backtesting Framework

Le framework de backtesting Constantine permet de tester et valider vos stratégies de trading sur des données historiques avant de les déployer en production.

## Démarrage Rapide

### 1. Avec des données générées (test rapide)

```bash
./bin/backtest --generate-sample --sample-candles=1000
```

### 2. Avec vos propres données CSV

```bash
./bin/backtest --data=path/to/your/data.csv --symbol=BTC-USD --verbose
```

## Format CSV Requis

Le fichier CSV doit contenir les colonnes suivantes :

```csv
timestamp,open,high,low,close,volume
1704067200,45000,45500,44800,45200,1000
1704070800,45200,45800,45000,45600,1200
...
```

### Formats de timestamp supportés :

- **Unix timestamp (secondes)** : `1704067200`
- **Unix timestamp (millisecondes)** : `1704067200000`
- **RFC3339** : `2024-01-01T12:00:00Z`
- **Date simple** : `2024-01-01 12:00:00`

## Options de Configuration

### Capital et Frais

```bash
./bin/backtest \
  --data=data.csv \
  --capital=50000 \           # Capital initial ($50,000)
  --commission=0.001 \        # Commission 0.1%
  --slippage=0.0005 \         # Slippage 0.05%
  --risk=0.02                 # Risque par trade 2%
```

### Paramètres de Stratégie

```bash
./bin/backtest \
  --data=data.csv \
  --short-ema=9 \             # EMA courte (9 périodes)
  --long-ema=21 \             # EMA longue (21 périodes)
  --rsi-period=14 \           # Période RSI
  --rsi-oversold=30 \         # Seuil RSI survendu
  --rsi-overbought=70 \       # Seuil RSI suracheté
  --take-profit=0.8 \         # Take profit 0.8%
  --stop-loss=0.4             # Stop loss 0.4%
```

### Options d'Affichage

```bash
./bin/backtest \
  --data=data.csv \
  --verbose                   # Afficher tous les trades
```

## Rapport de Performance

Le framework génère automatiquement un rapport complet incluant :

### 📊 Performance Globale
- Rendement total ($ et %)
- Rendement annualisé
- Drawdown maximum
- Durée totale de la période testée

### 📈 Statistiques de Trading
- Nombre total de trades
- Trades gagnants/perdants
- Taux de réussite (win rate)
- Durée moyenne des trades

### 💰 Analyse Profit/Perte
- Profit total
- Perte totale
- Facteur de profit (profit factor)
- Profit moyen (trades gagnants)
- Perte moyenne (trades perdants)
- Plus grand gain/perte

### 📋 Log des Trades
Avec l'option `--verbose`, vous obtenez le détail de chaque trade :
- Prix d'entrée et de sortie
- Durée du trade
- P&L en $ et %
- Raison de sortie (stop_loss, take_profit, signal, end_of_data)

## Exemple de Rapport

```
═══════════════════════════════════════════════════════
           BACKTESTING PERFORMANCE REPORT
═══════════════════════════════════════════════════════

📊 OVERALL PERFORMANCE
───────────────────────────────────────────────────────
Total Return:         $1,234.56 (12.35%)
Annualized Return:    45.67%
Max Drawdown:         $-456.78 (-4.57%)
Total Duration:       30d12h

📈 TRADE STATISTICS
───────────────────────────────────────────────────────
Total Trades:         42
Winning Trades:       28
Losing Trades:        14
Win Rate:             66.67%
Avg Trade Duration:   17h24m

💰 PROFIT/LOSS ANALYSIS
───────────────────────────────────────────────────────
Total Profit:         $2,345.67
Total Loss:           $1,111.11
Profit Factor:        2.11
Avg Profit (Win):     $83.77
Avg Loss (Lose):      $79.36
Largest Win:          $345.67
Largest Loss:         $234.56
```

## Architecture du Framework

### Composants Principaux

1. **Engine** (`internal/backtesting/engine.go`)
   - Exécute le backtest sur les données historiques
   - Gère les positions et le capital
   - Calcule les métriques de performance

2. **Data Loader** (`internal/backtesting/data_loader.go`)
   - Charge les données CSV
   - Génère des données de test
   - Parse différents formats de timestamp

3. **Simulated Exchange** (`internal/backtesting/simulated_exchange.go`)
   - Simule un exchange réel
   - Fournit les données historiques
   - Implémente l'interface `Exchange`

4. **Reporter** (`internal/backtesting/reporter.go`)
   - Génère des rapports de performance
   - Affiche les logs de trades
   - Calcule les métriques statistiques

### Gestion du Risque

Le framework intègre plusieurs mécanismes de gestion du risque :

- **Position Sizing** : Calcul automatique basé sur le risque par trade
- **Stop Loss** : Protection contre les pertes importantes
- **Take Profit** : Sécurisation des gains
- **Commission et Slippage** : Simulation réaliste des coûts

### Calcul de la Taille de Position

```
Risk Amount = Capital × Risk Per Trade
Stop Distance = |Entry Price - Stop Loss|
Position Size = Risk Amount ÷ Stop Distance
```

## Obtenir des Données Historiques

### Sources Recommandées

1. **Binance** : [https://data.binance.vision/](https://data.binance.vision/)
2. **CryptoDataDownload** : [https://www.cryptodatadownload.com/](https://www.cryptodatadownload.com/)
3. **Yahoo Finance** : Pour les crypto et actions traditionnelles

### Conversion de Données

Si vos données sont dans un format différent, vous pouvez les convertir :

```python
import pandas as pd

# Charger vos données
df = pd.read_csv('your_data.csv')

# Renommer les colonnes si nécessaire
df = df.rename(columns={
    'time': 'timestamp',
    'o': 'open',
    'h': 'high',
    'l': 'low',
    'c': 'close',
    'v': 'volume'
})

# Sauvegarder au bon format
df.to_csv('converted_data.csv', index=False)
```

## Optimisation de Stratégie

Pour trouver les meilleurs paramètres, vous pouvez lancer plusieurs backtests :

```bash
#!/bin/bash

# Test de différents paramètres EMA
for short in 5 7 9 12; do
  for long in 15 18 21 26; do
    echo "Testing EMA $short/$long"
    ./bin/backtest \
      --data=data.csv \
      --short-ema=$short \
      --long-ema=$long \
      --take-profit=0.8 \
      --stop-loss=0.4 \
      >> results.txt
  done
done
```

## Avertissements

⚠️ **Important** :
- Les performances passées ne garantissent pas les résultats futurs
- Le backtesting peut surestimer la performance (biais de survivorship, look-ahead bias)
- Testez toujours votre stratégie en paper trading avant le déploiement réel
- Prenez en compte les coûts de transaction réels (commission + slippage)

## Support et Contribution

Pour signaler des bugs ou suggérer des améliorations :
- Ouvrez une issue sur GitHub
- Consultez la documentation dans `AGENTS.md`

## Prochaines Étapes

Après avoir validé votre stratégie en backtesting :

1. **Paper Trading** : Testez en conditions réelles sans risque
2. **Small Capital Test** : Commencez avec un petit capital
3. **Monitoring** : Surveillez les performances en production
4. **Ajustements** : Affinez les paramètres basés sur les résultats réels
