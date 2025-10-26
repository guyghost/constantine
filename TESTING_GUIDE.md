# Guide de Test - dYdX v4 Testnet

Ce guide vous accompagne pas à pas pour tester le trading bot avec dYdX v4 sur testnet.

## 📋 Prérequis

### 1. Client Python dYdX (OBLIGATOIRE)

⚠️ **IMPORTANT** : Le client dYdX v4 requiert **Python 3.11 ou 3.12**. Python 3.14 n'est **pas encore supporté**.

```bash
# Vérifier votre version Python
python3 --version

# Si vous avez Python 3.14, installez Python 3.12
brew install python@3.12

# Créer un environnement virtuel avec Python 3.12
python3.12 -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# Installer le client officiel dYdX v4 (notez le nom correct du package)
pip install dydx-v4-client v4-proto

# Vérifier l'installation
python -c "from v4_client_py import NodeClient; print('✅ Client dYdX installé')"
```

Si vous rencontrez des problèmes, consultez **`PYTHON_SETUP.md`** pour des instructions détaillées.

### 2. Mnemonic dYdX Testnet

Vous avez besoin d'un wallet dYdX avec des USDC testnet.

**Option A : Créer un nouveau wallet testnet**

1. Allez sur https://v4.testnet.dydx.exchange/
2. Connectez-vous avec MetaMask
3. Créez un nouveau wallet ou importez-en un
4. Sauvegardez votre mnemonic (12 ou 24 mots)
5. Obtenez des USDC testnet via le faucet

**Option B : Utiliser un wallet existant**

Si vous avez déjà un mnemonic, assurez-vous qu'il a des fonds testnet.

### 3. Configuration des Variables d'Environnement

Éditez votre fichier `.env` :

```bash
# dYdX Configuration
ENABLE_DYDX=true
DYDX_MNEMONIC="your twelve or twenty four word mnemonic phrase here"
DYDX_SUBACCOUNT_NUMBER=0
```

⚠️ **IMPORTANT** : Utilisez un mnemonic TESTNET uniquement !

## 🚀 Tests Étape par Étape

### Test 1 : Vérification de l'Installation Python

```bash
# Test simple
python3 << 'EOF'
import sys
try:
    from v4_client_py import NodeClient, OrderFlags
    from v4_client_py.clients.constants import Network
    print("✅ Tous les modules importés avec succès")
    sys.exit(0)
except ImportError as e:
    print(f"❌ Erreur d'import: {e}")
    print("\n💡 Solution:")
    print("   pip3 install dydx-v4-client-py v4-proto")
    sys.exit(1)
EOF
```

### Test 2 : Vérification du Mnemonic

```bash
# Vérifier que le mnemonic est défini
if [ -z "$DYDX_MNEMONIC" ]; then
    echo "❌ DYDX_MNEMONIC non défini"
    echo "💡 Ajoutez-le à votre .env"
else
    echo "✅ DYDX_MNEMONIC défini (${#DYDX_MNEMONIC} caractères)"
fi
```

### Test 3 : Script de Test Complet

Nous avons créé un script de test dédié :

```bash
# Compiler et lancer le test
go run cmd/test-dydx/main.go
```

Ce script va :
- ✅ Se connecter au testnet dYdX
- ✅ Vérifier votre solde USDC
- ✅ Récupérer les données de marché BTC-USD
- ✅ Afficher l'orderbook
- ✅ Lister vos positions
- ✅ (Optionnel) Placer et annuler un ordre de test

### Test 4 : Test avec le Bot Complet

```bash
# Lancer le bot en mode headless
go run cmd/bot/main.go --headless
```

Observez les logs :
- `"exchange enabled"` avec `"auth":"mnemonic"` ✅
- `"WebSocket connected"` ✅
- `"received orderbook"` ✅

Arrêtez avec `Ctrl+C`.

## 🔍 Diagnostic des Erreurs Courantes

### Erreur : "Python client not initialized"

**Cause** : Le client n'a pas été créé avec un mnemonic.

**Solution** :
```bash
# Vérifiez que DYDX_MNEMONIC est défini
echo $DYDX_MNEMONIC

# Rechargez les variables
source .env
```

### Erreur : "dydx-v4-client-py not installed"

**Cause** : Le package Python n'est pas installé.

**Solution** :
```bash
pip3 install dydx-v4-client-py v4-proto

# Vérifier l'installation
pip3 list | grep dydx
```

### Erreur : "script not found at ..."

**Cause** : Le script Python n'est pas trouvé (déjà corrigé).

**Solution** : Lancez depuis la racine du projet :
```bash
cd /path/to/constantine
go run cmd/test-dydx/main.go
```

### Erreur : "invalid mnemonic"

**Cause** : Le mnemonic est invalide ou mal formaté.

**Solution** :
```bash
# Vérifiez le format (12 ou 24 mots séparés par des espaces)
echo $DYDX_MNEMONIC | wc -w

# Devrait afficher 12 ou 24
```

### Erreur : "Python script timeout"

**Cause** : Le script Python ne répond pas dans les 30 secondes.

**Solutions** :
1. Vérifiez votre connexion internet
2. Testez manuellement le script Python
3. Vérifiez que le réseau testnet est opérationnel : https://status.dydx.exchange/

### Erreur : "order placement failed: insufficient balance"

**Cause** : Solde USDC insuffisant sur testnet.

**Solution** :
1. Allez sur https://v4.testnet.dydx.exchange/
2. Connectez votre wallet
3. Utilisez le faucet pour obtenir des USDC testnet

## 📊 Vérification des Résultats

### Connexion Réussie

Vous devriez voir dans les logs :

```json
{"level":"INFO","msg":"exchange enabled","exchange":"dydx","auth":"mnemonic"}
{"level":"DEBUG","msg":"WebSocket connected","exchange":"dydx"}
{"level":"INFO","msg":"headless mode initialized"}
```

### Données de Marché

```json
{"level":"DEBUG","msg":"received orderbook","symbol":"BTC-USD","bids_count":10,"asks_count":10}
{"level":"DEBUG","msg":"received ticker","symbol":"BTC-USD","last":"67234.50"}
```

### Placement d'Ordre

```json
{"level":"INFO","msg":"order placed","order_id":"...","symbol":"BTC-USD","side":"buy"}
```

## 🎯 Scénario de Test Complet

### Scénario : Placement et Annulation d'Ordre

1. **Démarrer le test** :
   ```bash
   go run cmd/test-dydx/main.go
   ```

2. **Vérifier le solde** :
   - Assurez-vous d'avoir au moins 10 USDC testnet

3. **Accepter le test d'ordre** :
   - Tapez `oui` quand demandé
   - Un ordre LIMIT BUY sera placé à 50% du prix (ne sera pas exécuté)

4. **Observer l'annulation** :
   - L'ordre est automatiquement annulé après 3 secondes

5. **Vérifier le résultat** :
   - Tous les tests doivent être ✅

### Scénario : Trading avec le Bot

1. **Configurer le bot** (`.env`) :
   ```bash
   ENABLE_DYDX=true
   DYDX_MNEMONIC="votre mnemonic testnet"

   # Paramètres conservateurs pour testnet
   MAX_POSITION_SIZE=100
   RISK_PER_TRADE=1
   ```

2. **Lancer en mode headless** :
   ```bash
   go run cmd/bot/main.go --headless
   ```

3. **Observer les logs** :
   - Le bot reçoit les données de marché
   - Il calcule les indicateurs (EMA, RSI)
   - Il génère des signaux

4. **Arrêter proprement** :
   - `Ctrl+C` pour un arrêt propre

## 📈 Métriques de Succès

Votre test est réussi si :

- ✅ Connexion établie au testnet
- ✅ Solde USDC récupéré
- ✅ Données de marché reçues en temps réel
- ✅ Orderbook mis à jour régulièrement
- ✅ Ordre placé et annulé avec succès
- ✅ Aucune erreur critique dans les logs

## 🚨 Checklist de Sécurité

Avant de passer en production :

- [ ] Tests réussis sur testnet pendant au moins 24h
- [ ] Pas d'erreurs critiques dans les logs
- [ ] Solde testnet géré correctement
- [ ] Ordres placés et annulés sans problème
- [ ] WebSocket reconnecte automatiquement
- [ ] Positions affichées correctement
- [ ] Stop loss et take profit fonctionnent

## 🎓 Conseils Testnet

1. **Commencez petit** : Testez avec 0.001 BTC ou moins
2. **Observez d'abord** : Laissez tourner en mode observation pendant quelques heures
3. **Vérifiez les logs** : Cherchez les erreurs ou comportements étranges
4. **Testez tous les cas** : Placement, annulation, exécution partielle
5. **Simulez des pannes** : Coupez la connexion, redémarrez le bot

## 📚 Ressources

- [Documentation dYdX v4](https://docs.dydx.xyz/)
- [Testnet dYdX](https://v4.testnet.dydx.exchange/)
- [Status dYdX](https://status.dydx.exchange/)
- [Discord dYdX](https://discord.gg/dydx) - Support communautaire

## ❓ Problèmes ?

Si vous rencontrez des problèmes :

1. Vérifiez les logs : `tail -f scalping-bot.log`
2. Testez le script Python directement
3. Vérifiez le status du testnet : https://status.dydx.exchange/
4. Consultez `SECURITY_FIXES.md` pour les corrections de sécurité
5. Relisez `DYDX_QUICKSTART.md` pour la configuration

---

**⚠️ RAPPEL IMPORTANT** : Ne testez JAMAIS avec un mnemonic mainnet ou des fonds réels sur testnet !

Bon test ! 🚀
