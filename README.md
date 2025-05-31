# Projet Goit

Goit est une implémentation simplifiée de Git écrite en Go.  
Elle permet d'initialiser un dépôt local, d'ajouter des fichiers et de créer des commits, à la manière de Git.

---

## 🚀 Fonctionnalités

- Initialiser un dépôt (`init`)
- Ajouter des fichiers au staging (`add <file>`)
- Commiter les fichiers (`commit -m "<message>"`)

---

## 🛠️ Installation

### Prérequis
- Go 1.20+ installé ([Installer Go](https://golang.org/doc/install))

### Cloner le projet
```bash
git clone https://github.com/PedroDSS/projet-go-git/tree/main
cd projet-go-git
```

### Installer Goit (Linux/MacOS)
## Compiler le programme
```bash
# Il faut se situer sur le chemin du dossier avant.
go build -o goit ./cmd/goit
```

## (Optionnel) Déplacer l'executable dans le $PATH (Si on veut pouvoir l'utiliser partout)
```bash
sudo mv goit /usr/local/bin
```

## Vérifier le fonctionnement / Initialiser le repository
```bash
# Si le programme a été installé (executable partout)
goit init
# Si le package a été build
./goit init
# Sinon
go run ./cmd/goit init
```

## Voir les commandes disponibles
```bash
# Si le programme a été installé (executable partout)
goit help
# Si le package a été build
./goit help
# Sinon
go run ./cmd/goit help
```

### Désintaller
## Supprimer l'executable (Si il a été déplacer pouvoir l'utiliser partout)
```bash
sudo rm /usr/local/bin/goit
```

### Tester le workflow

```bash
./goit init
echo "Hello Golang World" > test.txt
./goit add test.txt
./goit commit -m "Mon premier commit en Go"
./goit status
./goit log
./goit branch feature
./goit checkout feature
./goit branch
```

### Exemple d'utilisation

```bash
# Initialisation du repository
goit init

# Ajout des fichiers en staging
goit add file1.txt
goit add file2.txt

# Création du commit avec les fichiers
goit commit -m "Ajout des fichiers"

# Vérifier le status
goit status

# Créer et changer de branches
goit branch feature-1
goit checkout feature-1

# Lister de toutes les branches
goit branch

# Voir l'historique de commit
goit log

# Vérifier les différences
goit diff file1.txt
```