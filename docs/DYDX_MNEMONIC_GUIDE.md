# Guide d'authentification dYdX avec Mnemonic

Ce guide explique comment configurer et utiliser Constantine avec dYdX en utilisant une phrase mnémonique pour l'authentification.

## Qu'est-ce qu'un Mnemonic ?

Un mnemonic (phrase mnémonique) est une série de 12 ou 24 mots qui représente votre clé privée de manière lisible. C'est la méthode standard pour sécuriser les wallets crypto.

**Exemple de mnemonic (12 mots)** :
```
income caught rib fork awesome amount topic quote slide error symbol vote
```

⚠️ **AVERTISSEMENT DE SÉCURITÉ** :
- **NE PARTAGEZ JAMAIS** votre mnemonic avec personne
- Quiconque possède votre mnemonic a un accès TOTAL à vos fonds
- Stockez-le dans un endroit sûr (password manager, coffre-fort)
- Ne le commitez JAMAIS dans Git
- Utilisez un wallet dédié au trading avec capital limité

## Configuration Rapide

### Étape 1 : Obtenir un Mnemonic

Vous avez deux options :

#### Option A : Générer un nouveau mnemonic

```bash
# Utiliser l'exemple fourni
go run examples/dydx_mnemonic_example.go
```

Cela va générer :
- Un nouveau mnemonic de 24 mots
- Une adresse dYdX
- Un numéro de subaccount

**Sauvegardez le mnemonic quelque part de sûr !**

#### Option B : Utiliser un mnemonic existant

Si vous avez déjà un wallet dYdX (créé avec MetaMask, Keplr, etc.), utilisez son mnemonic.

### Étape 2 : Configurer le fichier .env

Créez un fichier `.env` à la racine du projet :

```bash
# Exchange selection
EXCHANGE=dydx

# Mnemonic Authentication
EXCHANGE_API_KEY=            # Laissez vide pour l'instant
EXCHANGE_API_SECRET="your twelve or twenty four word mnemonic phrase here"

# Trading parameters
TRADING_SYMBOL=BTC-USD
INITIAL_BALANCE=10000
```

⚠️ **Important** :
- Mettez le mnemonic entre guillemets
- Gardez les espaces entre les mots
- Vérifiez qu'il y a 12 ou 24 mots

### Étape 3 : Ajouter .env au .gitignore

**Crucial !** Assurez-vous que `.env` est dans votre `.gitignore` :

```bash
echo ".env" >> .gitignore
```

### Étape 4 : Lancer le bot

```bash
# Mode headless
./bin/constantine --headless

# Mode TUI (terminal)
./bin/constantine
```

## Vérification de l'authentification

Quand le bot démarre, il affichera :

```
2025/10/23 23:10:00 Wallet initialized successfully
2025/10/23 23:10:00 Address: dydx1abc...xyz
2025/10/23 23:10:00 SubAccount: 0
```

Si vous voyez ces messages, votre authentification est réussie ! ✅

## Utilisation programmatique

### Créer un client avec mnemonic

```go
package main

import (
    "context"
    "log"

    "github.com/guyghost/constantine/internal/exchanges/dydx"
)

func main() {
    // Votre mnemonic (12 ou 24 mots)
    mnemonic := "word1 word2 word3 ... word12"

    // Créer le client avec subaccount 0
    client, err := dydx.NewClientWithMnemonic(mnemonic, 0)
    if err != nil {
        log.Fatal(err)
    }

    // Se connecter
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()

    // Vérifier l'authentification
    if client.IsAuthenticated() {
        log.Printf("✓ Authenticated as: %s", client.GetWalletAddress())
        log.Printf("✓ SubAccount: %s", client.GetSubAccountAddress())
    }

    // Utiliser le client
    balance, err := client.GetBalance(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Balance: %+v", balance)
}
```

### Générer un nouveau mnemonic

```go
import "github.com/guyghost/constantine/internal/exchanges/dydx"

// Générer un mnemonic de 24 mots
mnemonic, err := dydx.GenerateMnemonic()
if err != nil {
    log.Fatal(err)
}

fmt.Println("Votre nouveau mnemonic:", mnemonic)
fmt.Println("⚠️  Sauvegardez-le en lieu sûr !")
```

### Valider un mnemonic

```go
import "github.com/guyghost/constantine/internal/exchanges/dydx"

mnemonic := "word1 word2 ... word12"

if err := dydx.ValidateMnemonic(mnemonic); err != nil {
    log.Fatal("Mnemonic invalide:", err)
}

fmt.Println("✓ Mnemonic valide")
```

## Fonctionnalités disponibles avec authentification

Une fois authentifié avec votre mnemonic, vous pouvez :

✅ **Lecture de compte** :
- `GetBalance()` - Voir votre balance USDC
- `GetPositions()` - Voir vos positions ouvertes
- `GetPosition(symbol)` - Voir une position spécifique

🔜 **Trading** (à venir) :
- `PlaceOrder()` - Placer des ordres
- `CancelOrder()` - Annuler des ordres
- Gestion automatique des positions

## Dérivation d'adresse

Constantine utilise le standard BIP44 pour dériver votre adresse dYdX :

```
Chemin de dérivation: m/44'/118'/0'/0/0
                        ↑    ↑    ↑  ↑  ↑
                        │    │    │  │  └─ Index d'adresse (0)
                        │    │    │  └──── Change (0)
                        │    │    └─────── Compte (0)
                        │    └──────────── Coin type (118 = Cosmos)
                        └───────────────── Purpose (44 = BIP44)
```

- **Coin type 118** : Standard Cosmos/dYdX
- **Subaccount** : Par défaut 0 (peut être changé)

## Subaccounts

dYdX supporte plusieurs subaccounts par adresse :

```go
// Créer avec subaccount 0 (par défaut)
client, _ := dydx.NewClientWithMnemonic(mnemonic, 0)

// Créer avec subaccount 1
client, _ := dydx.NewClientWithMnemonic(mnemonic, 1)

// Créer avec subaccount 2
client, _ := dydx.NewClientWithMnemonic(mnemonic, 2)
```

Chaque subaccount a :
- Sa propre balance
- Ses propres positions
- Ses propres ordres

Cela vous permet de séparer différentes stratégies sur le même wallet.

## Troubleshooting

### Erreur: "invalid mnemonic phrase"

- Vérifiez que vous avez 12 ou 24 mots
- Vérifiez qu'il n'y a pas de fautes de frappe
- Les mots doivent être en minuscules
- Utilisez des espaces (pas de virgules ou points)

### Erreur: "wallet not initialized"

- Vérifiez que `EXCHANGE_API_SECRET` est défini dans `.env`
- Le mnemonic doit être entre guillemets
- Relancez le bot après modification

### Balance à zéro

- Votre wallet dYdX est peut-être vide
- Déposez des fonds via l'interface web dYdX
- Assurez-vous d'utiliser le bon subaccount (défaut: 0)

### Adresse différente de l'interface web

- Constantine dérive l'adresse selon BIP44 standard
- Si vous avez créé votre wallet différemment, l'adresse peut différer
- Utilisez le mnemonic original de création du wallet

## Sécurité

### ✅ Bonnes pratiques

- Utilisez un wallet dédié au trading
- Limitez le capital sur ce wallet
- Activez l'authentification à deux facteurs (2FA) sur votre compte principal
- Sauvegardez le mnemonic dans un gestionnaire de mots de passe
- Ne stockez JAMAIS le mnemonic en clair dans le code

### ❌ À éviter

- Partager votre mnemonic
- Commiter le fichier `.env` dans Git
- Utiliser le même wallet que pour vos économies
- Tester en production avec capital réel sans backtesting
- Donner accès à votre serveur sans sécurité

## Ressources

- [dYdX Documentation](https://docs.dydx.exchange/)
- [BIP39 Specification](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki)
- [BIP44 Specification](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)
- [Cosmos Address Format](https://docs.cosmos.network/main/build/spec/addresses/bech32)

## Support

Pour toute question :
1. Consultez la documentation dYdX
2. Vérifiez les logs du bot
3. Ouvrez une issue sur GitHub
