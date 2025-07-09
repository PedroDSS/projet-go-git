# Projet Goit - Une Implémentation de Git en Go

Goit est une implémentation simplifiée de Git écrite en Go, conçue pour comprendre les mécanismes internes d'un système de contrôle de version. Ce projet explore les concepts fondamentaux de Git tout en appliquant des principes de programmation fonctionnelle en Go.

## 📚 Table des Matières

- [Vue d'ensemble](#-vue-densemble)
- [Fonctionnalités Complètes](#-fonctionnalités-complètes)
- [Architecture et Choix Techniques](#-architecture-et-choix-techniques)
- [Installation et Utilisation](#-installation-et-utilisation)
- [Décisions de Conception](#-décisions-de-conception)
- [Limitations et Perspectives](#-limitations-et-perspectives)

## 🎯 Vue d'Ensemble

Goit implémente les fonctionnalités essentielles de Git en utilisant une approche event sourcing, où seuls les ajouts sont enregistrés dans l'historique. Le projet suit une architecture modulaire avec des packages internes dédiés à chaque fonctionnalité.

### Philosophie du Projet

- **Event Sourcing** : Sauvegarde uniquement les logs d'ajouts, pas les suppressions
- **Programmation Fonctionnelle** : Éviter les boucles `for` au profit d'approches fonctionnelles
- **Simplicité** : Code lisible et architecture claire pour faciliter la compréhension
- **Aucune Dépendance Externe** : Utilisation exclusive de la bibliothèque standard Go

## 🚀 Fonctionnalités Complètes

### 1. Gestion de Dépôt

#### `goit init`
- Initialise un nouveau dépôt dans le répertoire courant
- Crée la structure `.goit/` avec :
  - `objects/` : Stockage des objets (commits, trees)
  - `refs/heads/` : Références des branches
  - `HEAD` : Pointeur vers la branche courante
  - `index` : Zone de staging

#### `goit add <fichier>`
- Stage les fichiers pour le prochain commit
- Supporte `goit add .` pour tous les fichiers modifiés
- **Détection intelligente** : Ne stage que les fichiers réellement modifiés
- Compare les hash SHA-1 pour éviter les doublons

#### `goit commit -m "<message>"`
- Crée un commit avec les fichiers stagés
- Génère un objet tree et un objet commit
- Maintient la chaîne de parenté des commits
- Horodatage UTC pour la cohérence

### 2. Inspection et État

#### `goit status`
- Affiche l'état du dépôt avec code couleur :
  - **Vert** : Fichiers stagés
  - **Rouge** : Fichiers modifiés ou non suivis
- Détecte automatiquement les nouveaux fichiers
- Compare working directory, index et dernier commit

#### `goit diff [fichier]`
- Montre les différences entre les versions
- Affiche la taille des fichiers modifiés
- Support pour un fichier spécifique ou tous les fichiers

#### `goit log [--compact]`
- **Mode détaillé** : Hash complet, date, auteur, message, branches
- **Mode compact** : Hash court + message
- Suit la chaîne de parenté des commits
- Affichage coloré des références (HEAD, branches)

### 3. Gestion des Branches

#### `goit branch [nom]`
- Sans argument : Liste toutes les branches (courante marquée avec `*`)
- Avec argument : Crée une nouvelle branche au commit actuel
- Validation des noms (pas d'espaces, pas de slashes)

#### `goit checkout <branche>`
- Change de branche active
- Met à jour la référence HEAD
- Vérifie l'existence de la branche cible

### 4. Commandes Utilitaires

#### `goit help`
- Affiche la liste des commandes disponibles
- Guide d'utilisation rapide

## 🏗️ Architecture et Choix Techniques

### Structure du Projet

```
projet-go-git/
├── cmd/goit/main.go         # Point d'entrée CLI
├── internal/                # Logique métier (packages internes)
│   ├── branch/              # Gestion des branches
│   ├── checkout/            # Changement de branches
│   ├── index/               # Zone de staging
│   ├── log/                 # Affichage de l'historique
│   ├── objects/             # Stockage des objets Git
│   ├── repository/          # Opérations du dépôt
│   └── status/              # État et différences
└── .goit/                   # Répertoire Git local
    ├── HEAD                 # Référence de branche courante
    ├── index                # Fichiers stagés
    ├── objects/             # Objets (SHA-1 comme nom)
    └── refs/heads/          # Références des branches
```

### Modèle de Stockage

#### 1. **Stockage Adressable par Contenu**
- Utilisation de SHA-1 pour identifier les objets
- Chaque modification crée un nouvel objet
- Pas de déduplication (simplicité privilégiée)

#### 2. **Types d'Objets**
- **Tree** : Représente l'état d'un répertoire
- **Commit** : Métadonnées + référence au tree + parent
- Format texte simple pour faciliter le débogage

#### 3. **Format de l'Index**
```
<sha1-hash> <nom-fichier>
```
Simple et efficace pour les opérations de base

### Décisions Techniques Clés

#### 1. **Event Sourcing**
- **Choix** : Ne sauvegarder que les ajouts
- **Justification** : Simplicité et immutabilité des données
- **Avantage** : Historique complet sans corruption possible

#### 2. **Pas de Compression**
- **Choix** : Objets stockés en texte brut
- **Justification** : Facilite le débogage et la compréhension
- **Compromis** : Espace disque vs lisibilité

#### 3. **Architecture Modulaire**
- **Choix** : Un package par fonctionnalité
- **Justification** : Séparation des responsabilités
- **Avantage** : Code maintenable et extensible

#### 4. **Gestion des Couleurs**
- **Choix** : Codes ANSI pour le terminal
- **Justification** : Meilleure expérience utilisateur
- **Implementation** : Jaune (commits), Bleu (HEAD), Rose (branches)

#### 5. **Validation Stricte**
- **Choix** : Vérifications avant chaque opération
- **Justification** : Éviter la corruption du dépôt
- **Exemple** : Détection des changements avant staging

## 🛠️ Installation et Utilisation

### Prérequis
- Go 1.21+ installé ([Installer Go](https://golang.org/doc/install))

### Installation

```bash
# Cloner le projet
git clone https://github.com/PedroDSS/projet-go-git
cd projet-go-git

# Compiler
go build -o goit ./cmd/goit

# (Optionnel) Installation globale
sudo mv goit /usr/local/bin
```

### Workflow Exemple

```bash
# Initialiser un nouveau dépôt
goit init

# Créer et modifier des fichiers
echo "Hello World" > hello.txt
echo "Goit Project" > README.txt

# Ajouter au staging
goit add hello.txt
goit add README.txt

# Créer un commit
goit commit -m "Initial commit"

# Vérifier le statut
goit status

# Créer une nouvelle branche
goit branch feature-1

# Changer de branche
goit checkout feature-1

# Modifier des fichiers
echo "Modified content" >> hello.txt

# Voir les différences
goit diff hello.txt

# Commiter les changements
goit add hello.txt
goit commit -m "Update hello.txt"

# Voir l'historique
goit log
goit log --compact

# Retour sur main
goit checkout main
```

## 🔍 Décisions de Conception

### 1. **Pourquoi Go ?**
- Performance native
- Gestion simple de la concurrence
- Compilation en binaire unique
- Excellent support des opérations système

### 2. **Pourquoi l'Event Sourcing ?**
- Immutabilité garantie
- Historique complet et auditable
- Récupération facile en cas d'erreur
- Alignement avec la philosophie Git

### 3. **Pourquoi Pas de Boucles For ?**
- Encourager la pensée fonctionnelle
- Code plus déclaratif
- Meilleure composition des fonctions
- Challenge technique intéressant

### 4. **Pourquoi des Packages Internes ?**
- Encapsulation forte
- API claire et définie
- Éviter les dépendances circulaires
- Faciliter les tests futurs

## 📊 Limitations et Perspectives

### Limitations Actuelles

1. **Pas de Merge** : Fusion des branches non implémentée
2. **Diff Basique** : Pas de diff ligne par ligne
3. **Pas de Remote** : Aucune opération réseau
4. **Checkout Incomplet** : Ne restaure pas les fichiers
5. **Pas de Tags** : Seules les branches sont supportées
6. **Pas de .gitignore** : Patterns d'exclusion codés en dur

### Améliorations Futures Possibles

1. **Implémentation du Merge**
   - Stratégies de fusion (fast-forward, 3-way)
   - Résolution de conflits

2. **Amélioration du Diff**
   - Algorithme de Myers pour diff ligne par ligne
   - Coloration syntaxique

3. **Support Remote**
   - Protocoles Git (HTTP, SSH)
   - Push/Pull/Clone

4. **Optimisations**
   - Pack files pour la compression
   - Déduplication des blobs
   - Cache de performances

5. **Fonctionnalités Avancées**
   - Tags et releases
   - Stash pour sauvegarder temporairement
   - Rebase interactif

### Fichiers Ignorés Automatiquement

- `.goit/`, `.git/` (répertoires)
- `.gitignore`, `goit` (exécutable)
- `.DS_Store`, `Thumbs.db`, `desktop.ini` (système)

## 🎓 Apprentissages du Projet

Ce projet nous a permis d'explorer :

1. **Les Internals de Git** : Compréhension profonde du modèle d'objets
2. **L'Event Sourcing** : Application pratique du pattern
3. **La Programmation Fonctionnelle en Go** : Défis et solutions
4. **L'Architecture Logicielle** : Conception modulaire et maintenable
5. **Les Systèmes de Fichiers** : Gestion efficace du stockage

## 📝 Conclusion

Goit représente une exploration approfondie des concepts de contrôle de version, alliant théorie et pratique. Bien que simplifié par rapport à Git, il capture l'essence des mécanismes fondamentaux tout en restant accessible et pédagogique.

Le projet démontre qu'il est possible de créer un système de contrôle de version fonctionnel avec relativement peu de code, tout en maintenant une architecture claire et extensible.

---

**Auteurs** :
- DA SILVA SOUSA Pedro
- GODARD Lucie
- JOUVET Erwann
**Licence** : MIT