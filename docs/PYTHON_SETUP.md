# Installation du Client Python dYdX v4

## ⚠️ Problème de Compatibilité Python 3.14

Le client officiel dYdX v4 **n'est pas encore compatible avec Python 3.14**. Vous devez utiliser **Python 3.11 ou 3.12**.

## Solution 1 : Installer Python 3.12 (Recommandé)

```bash
# Installer Python 3.12 via Homebrew
brew install python@3.12

# Vérifier l'installation
python3.12 --version

# Créer un environnement virtuel avec Python 3.12
python3.12 -m venv venv

# Activer l'environnement
source venv/bin/activate

# Installer le client dYdX
pip install dydx-v4-client v4-proto

# Vérifier l'installation
python -c "from v4_client_py import NodeClient; print('✅ Installation réussie!')"
```

## Solution 2 : Utiliser pyenv (Plus Flexible)

```bash
# Installer pyenv
brew install pyenv

# Installer Python 3.12
pyenv install 3.12.0

# Créer un environnement virtuel
pyenv virtualenv 3.12.0 dydx-env

# Activer l'environnement
pyenv activate dydx-env

# Installer le client dYdX
pip install dydx-v4-client v4-proto
```

## Solution 3 : Modifier le Script Python pour Utiliser Python 3.12

Si vous avez Python 3.12 installé, modifiez la configuration du client :

```bash
# Dans votre .env ou configuration
export DYDX_PYTHON_PATH=/opt/homebrew/bin/python3.12
```

Ou modifiez directement dans le code Go :

```go
pythonClient, err := NewPythonClient(&PythonClientConfig{
    PythonPath: "/opt/homebrew/bin/python3.12", // Chemin vers Python 3.12
    Network:    c.network,
    Mnemonic:   mnemonic,
})
```

## Vérification

Après installation, vérifiez que tout fonctionne :

```bash
# Activer l'environnement
source venv/bin/activate

# Tester l'import
python << 'EOF'
from v4_client_py import NodeClient, OrderFlags
from v4_client_py.clients.constants import Network
print("✅ Tous les modules importés avec succès")
EOF
```

## Versions Python Supportées

| Version Python | Support dYdX v4 | Recommandé |
|----------------|-----------------|------------|
| 3.14.x         | ❌ Non          | Non        |
| 3.13.x         | ⚠️  Partiel     | Non        |
| 3.12.x         | ✅ Oui          | **Oui**    |
| 3.11.x         | ✅ Oui          | Oui        |
| 3.10.x         | ⚠️  Limité      | Non        |

## Alternative : Mode Sans Python (Lecture Seule)

Si vous ne pouvez pas installer Python 3.12, vous pouvez utiliser le bot en **mode lecture seule** :

```bash
# Le bot fonctionnera pour :
# ✅ Recevoir des données de marché
# ✅ Analyser les signaux
# ✅ Afficher les positions
# ✅ Consulter le solde

# Mais PAS pour :
# ❌ Placer des ordres
# ❌ Annuler des ordres
# ❌ Trading automatique
```

## Prochaines Étapes

Une fois Python 3.12 installé avec le client dYdX :

1. Relancez le script de test :
   ```bash
   source venv/bin/activate
   go run cmd/test-dydx/main.go
   ```

2. Ou lancez le bot complet :
   ```bash
   source venv/bin/activate
   go run cmd/bot/main.go --headless
   ```

## Liens Utiles

- [dYdX v4 Python Client (PyPI)](https://pypi.org/project/dydx-v4-client/)
- [Documentation dYdX v4](https://docs.dydx.xyz/)
- [Homebrew Python](https://formulae.brew.sh/formula/python@3.12)
- [pyenv Documentation](https://github.com/pyenv/pyenv)

## Besoin d'Aide ?

Si vous rencontrez des problèmes :
1. Vérifiez votre version de Python : `python --version`
2. Consultez les logs d'erreur détaillés
3. Assurez-vous d'avoir activé le venv : `source venv/bin/activate`
4. Réinstallez proprement : `pip install --upgrade --force-reinstall dydx-v4-client`
